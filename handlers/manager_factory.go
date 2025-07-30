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
	"github.com/user/mcp-todo-server/internal/logging"
)

// ManagerFactory creates and caches managers for different working directories
type ManagerFactory struct {
	mu            sync.RWMutex
	managers      map[string]*managerSet
	baseManager   TodoManager // Fallback when no context
	baseSearch    SearchEngine
	baseStats     StatsEngine
	baseTemplates TemplateManager
	
	// Circuit breaker for manager creation
	creationAttempts  map[string]int
	lastFailureTime   map[string]time.Time
	maxCreationAttempts int
	backoffDuration   time.Duration
	
	// Test support - allow injecting custom linker for tests
	testLinker    TodoLinker
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
		managers:            make(map[string]*managerSet),
		baseManager:         baseManager,
		baseSearch:          baseSearch,
		baseStats:           baseStats,
		baseTemplates:       baseTemplates,
		creationAttempts:    make(map[string]int),
		lastFailureTime:     make(map[string]time.Time),
		maxCreationAttempts: 3,
		backoffDuration:     30 * time.Second,
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

	// Circuit breaker: Check if we should allow manager creation
	if f.shouldBlockManagerCreation(workingDir) {
		fmt.Fprintf(os.Stderr, "Circuit breaker: Blocking manager creation for %s (too many recent failures)\n", workingDir)
		return f.baseManager, f.baseSearch, f.baseStats, f.baseTemplates, nil
	}

	// Create new manager set for this directory
	fmt.Fprintf(os.Stderr, "Creating new manager set for working directory: %s\n", workingDir)
	
	// Resolve paths relative to working directory
	todoPath := filepath.Join(workingDir, ".claude", "todos")
	templatePath := filepath.Join(workingDir, ".claude", "templates")
	indexPath := filepath.Join(workingDir, ".claude", "index", "todos.bleve")

	// Ensure directories exist
	if err := os.MkdirAll(todoPath, 0755); err != nil {
		f.recordManagerCreationFailure(workingDir)
		return nil, nil, nil, nil, interrors.Wrap(err, "failed to create todo directory")
	}
	if err := os.MkdirAll(templatePath, 0755); err != nil {
		f.recordManagerCreationFailure(workingDir)
		return nil, nil, nil, nil, interrors.Wrap(err, "failed to create template directory")
	}
	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		f.recordManagerCreationFailure(workingDir)
		return nil, nil, nil, nil, interrors.Wrap(err, "failed to create index directory")
	}

	// Create managers - TodoManager expects the working directory, not the todos path
	manager := core.NewTodoManager(workingDir)
	
	// Create search engine with timeout and graceful fallback
	search, err := f.createSearchEngineWithTimeout(ctx, indexPath, todoPath)
	if err != nil {
		// Log error but continue without search functionality
		logging.Warnf("Failed to create search engine for %s: %v. Continuing without search functionality.", workingDir, err)
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

	// Record successful creation to reset circuit breaker
	f.recordManagerCreationSuccess(workingDir)

	logging.Infof("Created manager set for %s (total cached: %d)", workingDir, len(f.managers))
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
					logging.Errorf("Error closing search engine for %s: %v", dir, err)
				}
			}
			delete(f.managers, dir)
			removed++
			logging.Infof("Removed stale manager set for %s", dir)
		}
	}

	if removed > 0 {
		logging.Infof("Cleaned up %d stale manager sets", removed)
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

// shouldBlockManagerCreation checks if manager creation should be blocked due to circuit breaker
func (f *ManagerFactory) shouldBlockManagerCreation(workingDir string) bool {
	attempts := f.creationAttempts[workingDir]
	lastFailure := f.lastFailureTime[workingDir]
	
	// If we haven't exceeded max attempts, allow creation
	if attempts < f.maxCreationAttempts {
		return false
	}
	
	// If we've exceeded max attempts, check if enough time has passed
	if time.Since(lastFailure) > f.backoffDuration {
		// Reset the circuit breaker
		f.creationAttempts[workingDir] = 0
		delete(f.lastFailureTime, workingDir)
		return false
	}
	
	// Block creation - still in backoff period
	return true
}

// recordManagerCreationFailure records a failure for circuit breaker purposes
func (f *ManagerFactory) recordManagerCreationFailure(workingDir string) {
	f.creationAttempts[workingDir]++
	f.lastFailureTime[workingDir] = time.Now()
	fmt.Fprintf(os.Stderr, "Recorded manager creation failure for %s (attempt %d/%d)\n", 
		workingDir, f.creationAttempts[workingDir], f.maxCreationAttempts)
}

// recordManagerCreationSuccess resets the circuit breaker on success
func (f *ManagerFactory) recordManagerCreationSuccess(workingDir string) {
	delete(f.creationAttempts, workingDir)
	delete(f.lastFailureTime, workingDir)
}

// CreateLinker creates a TodoLinker for the given manager
func (f *ManagerFactory) CreateLinker(manager TodoManager) TodoLinker {
	// If we have a test linker injected, use it
	if f.testLinker != nil {
		return f.testLinker
	}
	
	// Try to cast to concrete TodoManager for linker creation
	if concreteManager, ok := manager.(*core.TodoManager); ok {
		return core.NewTodoLinker(concreteManager)
	}
	// Return nil if not a concrete TodoManager
	return nil
}

// SetTestLinker sets a test linker for testing purposes
func (f *ManagerFactory) SetTestLinker(linker TodoLinker) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.testLinker = linker
}