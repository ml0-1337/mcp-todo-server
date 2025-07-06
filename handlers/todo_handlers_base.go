package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/mcp-todo-server/core"
	interrors "github.com/user/mcp-todo-server/internal/errors"
)

// TodoHandlers contains handlers for all todo operations
type TodoHandlers struct {
	factory *ManagerFactory
	// Direct reference to TodoManager for linker creation
	baseManager *core.TodoManager
	// Cleanup routine control
	cleanupStop chan struct{}
	cleanupDone chan struct{}
}

// NewTodoHandlers creates new todo handlers with dependencies
func NewTodoHandlers(todoPath, templatePath string) (*TodoHandlers, error) {
	// Create a simple todo manager without context wrapper
	baseManager := core.NewTodoManager(todoPath)

	// Create search engine
	indexPath := filepath.Join(todoPath, "..", "index", "todos.bleve")
	fmt.Fprintf(os.Stderr, "Creating search engine with indexPath=%s, todoPath=%s\n", indexPath, todoPath)
	searchEngine, err := core.NewSearchEngine(indexPath, todoPath)
	if err != nil {
		return nil, interrors.Wrap(err, "failed to create search engine")
	}

	// Create stats engine with the manager
	stats := core.NewStatsEngine(baseManager)

	// Create template manager
	templates := core.NewTemplateManager(templatePath)

	// Create factory with base managers
	factory := NewManagerFactory(baseManager, searchEngine, stats, templates)

	h := &TodoHandlers{
		factory:     factory,
		baseManager: baseManager,
		cleanupStop: make(chan struct{}),
		cleanupDone: make(chan struct{}),
	}

	// Start cleanup routine
	go h.cleanupRoutine()

	return h, nil
}


// NewTodoHandlersWithDependencies creates new todo handlers with explicit dependencies (for testing)
func NewTodoHandlersWithDependencies(
	manager TodoManager,
	search SearchEngine,
	stats StatsEngine,
	templates TemplateManager,
) *TodoHandlers {
	// Create factory with provided managers
	factory := NewManagerFactory(manager, search, stats, templates)
	
	return &TodoHandlers{
		factory: factory,
		// baseManager is nil for test scenarios - tests should set it if needed
		baseManager: nil,
	}
}

// Close cleans up resources
func (h *TodoHandlers) Close() error {
	// Stop cleanup routine
	if h.cleanupStop != nil {
		close(h.cleanupStop)
		<-h.cleanupDone
	}
	
	// Factory manages cleanup of all search engines
	// No need to close individual search engines here
	return nil
}

// cleanupRoutine periodically cleans up stale manager sets
func (h *TodoHandlers) cleanupRoutine() {
	defer close(h.cleanupDone)
	
	// Run cleanup every 10 minutes
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	// Manager timeout is 15 minutes
	managerTimeout := 15 * time.Minute
	
	for {
		select {
		case <-ticker.C:
			removed := h.factory.CleanupStale(managerTimeout)
			if removed > 0 {
				fmt.Fprintf(os.Stderr, "Manager cleanup: removed %d stale manager sets\n", removed)
			}
		case <-h.cleanupStop:
			fmt.Fprintln(os.Stderr, "Stopping manager cleanup routine")
			return
		}
	}
}


