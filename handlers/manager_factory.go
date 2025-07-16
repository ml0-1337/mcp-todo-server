package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/user/mcp-todo-server/core"
	ctxkeys "github.com/user/mcp-todo-server/internal/context"
	interrors "github.com/user/mcp-todo-server/internal/errors"
)

// ManagerFactory creates and caches managers for different working directories
type ManagerFactory struct {
	mu            sync.RWMutex
	managers      map[string]*managerSet
	baseManager   TodoManager // Fallback when no context
	baseSearch    SearchEngine
	baseStats     StatsEngine
	baseTemplates TemplateManager
}

// managerSet contains all managers for a specific directory
type managerSet struct {
	manager      TodoManager
	search       SearchEngine
	stats        StatsEngine
	templates    TemplateManager
	lastAccessed time.Time
}

// NewManagerFactory creates a new manager factory with a base manager for fallback
func NewManagerFactory(baseManager TodoManager, baseSearch SearchEngine, baseStats StatsEngine, baseTemplates TemplateManager) *ManagerFactory {
	return &ManagerFactory{
		managers:      make(map[string]*managerSet),
		baseManager:   baseManager,
		baseSearch:    baseSearch,
		baseStats:     baseStats,
		baseTemplates: baseTemplates,
	}
}

// GetManagers returns the appropriate managers for the given context
func (f *ManagerFactory) GetManagers(ctx context.Context) (TodoManager, SearchEngine, StatsEngine, TemplateManager, error) {
	// Add timeout for manager creation to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// Try to get working directory from context
	workingDir, ok := ctx.Value(ctxkeys.WorkingDirectoryKey).(string)
	if !ok || workingDir == "" {
		// No context, use base managers
		fmt.Fprintf(os.Stderr, "No working directory in context, using base managers\n")
		return f.baseManager, f.baseSearch, f.baseStats, f.baseTemplates, nil
	}
	
	fmt.Fprintf(os.Stderr, "GetManagers called with working directory: %s\n", workingDir)

	// Check cache first
	f.mu.RLock()
	if set, exists := f.managers[workingDir]; exists {
		set.lastAccessed = time.Now()
		f.mu.RUnlock()
		return set.manager, set.search, set.stats, set.templates, nil
	}
	f.mu.RUnlock()

	// Need to create new managers
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock
	if set, exists := f.managers[workingDir]; exists {
		set.lastAccessed = time.Now()
		return set.manager, set.search, set.stats, set.templates, nil
	}

	// Create new manager set for this directory
	fmt.Fprintf(os.Stderr, "Creating new manager set for working directory: %s\n", workingDir)
	
	// Resolve paths relative to working directory
	todoPath := filepath.Join(workingDir, ".claude", "todos")
	templatePath := filepath.Join(workingDir, ".claude", "templates")
	indexPath := filepath.Join(workingDir, ".claude", "index", "todos.bleve")

	// Ensure directories exist
	if err := os.MkdirAll(todoPath, 0755); err != nil {
		return nil, nil, nil, nil, interrors.Wrap(err, "failed to create todo directory")
	}
	if err := os.MkdirAll(templatePath, 0755); err != nil {
		return nil, nil, nil, nil, interrors.Wrap(err, "failed to create template directory")
	}
	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		return nil, nil, nil, nil, interrors.Wrap(err, "failed to create index directory")
	}

	// Create managers - TodoManager expects the working directory, not the todos path
	manager := core.NewTodoManager(workingDir)
	
	// Create search engine with timeout and graceful fallback
	search, err := f.createSearchEngineWithTimeout(ctx, indexPath, todoPath)
	if err != nil {
		// Log error but continue without search functionality
		fmt.Fprintf(os.Stderr, "Warning: Failed to create search engine for %s: %v. Continuing without search functionality.\n", workingDir, err)
		search = nil // Will be handled gracefully by handlers
	}

	stats := core.NewStatsEngine(manager)
	templates := core.NewTemplateManager(templatePath)

	// Cache the managers
	f.managers[workingDir] = &managerSet{
		manager:      manager,
		search:       search,
		stats:        stats,
		templates:    templates,
		lastAccessed: time.Now(),
	}

	fmt.Fprintf(os.Stderr, "Created manager set for %s (total cached: %d)\n", workingDir, len(f.managers))
	return manager, search, stats, templates, nil
}

// CleanupStale removes managers that haven't been accessed recently
func (f *ManagerFactory) CleanupStale(maxAge time.Duration) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	removed := 0

	for dir, set := range f.managers {
		if now.Sub(set.lastAccessed) > maxAge {
			// Close search engine if it has a Close method
			if closer, ok := set.search.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "Error closing search engine for %s: %v\n", dir, err)
				}
			}
			delete(f.managers, dir)
			removed++
			fmt.Fprintf(os.Stderr, "Removed stale manager set for %s\n", dir)
		}
	}

	if removed > 0 {
		fmt.Fprintf(os.Stderr, "Cleaned up %d stale manager sets\n", removed)
	}

	return removed
}

// GetActiveCount returns the number of cached manager sets
func (f *ManagerFactory) GetActiveCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.managers)
}

// createSearchEngineWithTimeout creates a search engine with timeout protection
func (f *ManagerFactory) createSearchEngineWithTimeout(ctx context.Context, indexPath, todoPath string) (SearchEngine, error) {
	// Channel to receive result from goroutine
	type result struct {
		engine SearchEngine
		err    error
	}
	resultCh := make(chan result, 1)
	
	// Start search engine creation in goroutine
	go func() {
		engine, err := core.NewSearchEngine(indexPath, todoPath)
		resultCh <- result{engine: engine, err: err}
	}()
	
	// Wait for either completion or timeout
	select {
	case res := <-resultCh:
		return res.engine, res.err
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("search engine creation timed out after 30 seconds")
		}
		return nil, ctx.Err()
	}
}