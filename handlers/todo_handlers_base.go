package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/mcp-todo-server/core"
	interrors "github.com/user/mcp-todo-server/internal/errors"
)

// TodoHandlers contains handlers for all todo operations
type TodoHandlers struct {
	manager   TodoManager
	search    SearchEngine
	stats     StatsEngine
	templates TemplateManager
	// Direct reference to TodoManager for linker creation
	baseManager *core.TodoManager
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

	return &TodoHandlers{
		manager:     baseManager,
		search:      searchEngine,
		stats:       stats,
		templates:   templates,
		baseManager: baseManager,
	}, nil
}


// NewTodoHandlersWithDependencies creates new todo handlers with explicit dependencies (for testing)
func NewTodoHandlersWithDependencies(
	manager TodoManager,
	search SearchEngine,
	stats StatsEngine,
	templates TemplateManager,
) *TodoHandlers {
	return &TodoHandlers{
		manager:   manager,
		search:    search,
		stats:     stats,
		templates: templates,
		// baseManager is nil for test scenarios - tests should set it if needed
		baseManager: nil,
	}
}

// Close cleans up resources
func (h *TodoHandlers) Close() error {
	if h.search != nil {
		return h.search.Close()
	}
	return nil
}


