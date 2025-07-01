package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/user/mcp-todo-server/core"
	"github.com/user/mcp-todo-server/internal/infrastructure/factory"
)

// TodoHandlers contains handlers for all todo operations
type TodoHandlers struct {
	manager   TodoManagerInterface
	search    SearchEngineInterface
	stats     StatsEngineInterface
	templates TemplateManagerInterface
	// Factory function for creating TodoLinker (to avoid type assertion)
	createLinker func(TodoManagerInterface) TodoLinkerInterface
}

// NewTodoHandlers creates new todo handlers with dependencies
func NewTodoHandlers(todoPath, templatePath string) (*TodoHandlers, error) {
	// Create todo manager using the new architecture via factory
	// This will be replaced with direct repository/service creation after transition
	manager := NewContextualTodoManagerWrapper(todoPath)

	// Create search engine
	indexPath := filepath.Join(todoPath, "..", "index", "todos.bleve")
	searchEngine, err := core.NewSearchEngine(indexPath, filepath.Join(todoPath, ".claude", "todos"))
	if err != nil {
		return nil, fmt.Errorf("failed to create search engine: %w", err)
	}

	// Create stats engine with the default manager
	// (stats needs concrete TodoManager, not the wrapper)
	stats := core.NewStatsEngine(manager.defaultManager)

	// Create template manager
	templates := core.NewTemplateManager(templatePath)

	return &TodoHandlers{
		manager:   manager,
		search:    searchEngine,
		stats:     stats,
		templates: templates,
		createLinker: func(m TodoManagerInterface) TodoLinkerInterface {
			// Handle ContextualTodoManagerWrapper
			if wrapper, ok := m.(*ContextualTodoManagerWrapper); ok {
				return core.NewTodoLinker(wrapper.defaultManager)
			}
			// Type assert to concrete type for core.NewTodoLinker
			if tm, ok := m.(*core.TodoManager); ok {
				return core.NewTodoLinker(tm)
			}
			return nil
		},
	}, nil
}

// NewTodoHandlersWithRepository creates handlers using the new repository pattern
func NewTodoHandlersWithRepository(todoPath, templatePath string) (*TodoHandlers, error) {
	// Create a simple context manager that uses the repository pattern
	contextManager := &SimpleContextManagerWrapper{
		defaultPath: todoPath,
		managers:    make(map[string]TodoManagerInterface),
	}

	// Create search engine
	indexPath := filepath.Join(todoPath, "..", "index", "todos.bleve")
	searchEngine, err := core.NewSearchEngine(indexPath, filepath.Join(todoPath, ".claude", "todos"))
	if err != nil {
		return nil, fmt.Errorf("failed to create search engine: %w", err)
	}

	// Create a default manager for stats (temporary until stats is refactored)
	defaultManager := core.NewTodoManager(todoPath)
	stats := core.NewStatsEngine(defaultManager)

	// Create template manager
	templates := core.NewTemplateManager(templatePath)

	return &TodoHandlers{
		manager:   contextManager,
		search:    searchEngine,
		stats:     stats,
		templates: templates,
		createLinker: func(m TodoManagerInterface) TodoLinkerInterface {
			// For now, use the default manager
			return core.NewTodoLinker(defaultManager)
		},
	}, nil
}

// SimpleContextManagerWrapper implements TodoManagerInterface with context awareness
type SimpleContextManagerWrapper struct {
	defaultPath string
	managers    map[string]TodoManagerInterface
	mu          sync.RWMutex
}

// Implement all TodoManagerInterface methods
func (w *SimpleContextManagerWrapper) CreateTodo(task, priority, todoType string) (*core.Todo, error) {
	adapter := factory.CreateTodoManager(w.defaultPath)
	return adapter.CreateTodo(task, priority, todoType)
}

func (w *SimpleContextManagerWrapper) ReadTodo(id string) (*core.Todo, error) {
	adapter := factory.CreateTodoManager(w.defaultPath)
	return adapter.ReadTodo(id)
}

func (w *SimpleContextManagerWrapper) ReadTodoWithContent(id string) (*core.Todo, string, error) {
	adapter := factory.CreateTodoManager(w.defaultPath)
	return adapter.ReadTodoWithContent(id)
}

func (w *SimpleContextManagerWrapper) UpdateTodo(id, section, operation, content string, metadata map[string]string) error {
	adapter := factory.CreateTodoManager(w.defaultPath)
	return adapter.UpdateTodo(id, section, operation, content, metadata)
}

func (w *SimpleContextManagerWrapper) SaveTodo(todo *core.Todo) error {
	adapter := factory.CreateTodoManager(w.defaultPath)
	return adapter.SaveTodo(todo)
}

func (w *SimpleContextManagerWrapper) ListTodos(status, priority string, days int) ([]*core.Todo, error) {
	adapter := factory.CreateTodoManager(w.defaultPath)
	return adapter.ListTodos(status, priority, days)
}

func (w *SimpleContextManagerWrapper) ReadTodoContent(id string) (string, error) {
	adapter := factory.CreateTodoManager(w.defaultPath)
	return adapter.ReadTodoContent(id)
}

func (w *SimpleContextManagerWrapper) ArchiveTodo(id, quarter string) error {
	adapter := factory.CreateTodoManager(w.defaultPath)
	return adapter.ArchiveTodo(id, quarter)
}

func (w *SimpleContextManagerWrapper) ArchiveOldTodos(days int) (int, error) {
	adapter := factory.CreateTodoManager(w.defaultPath)
	return adapter.ArchiveOldTodos(days)
}

func (w *SimpleContextManagerWrapper) FindDuplicateTodos() ([][]string, error) {
	adapter := factory.CreateTodoManager(w.defaultPath)
	return adapter.FindDuplicateTodos()
}

func (w *SimpleContextManagerWrapper) GetBasePath() string {
	return w.defaultPath
}

// NewTodoHandlersWithDependencies creates new todo handlers with explicit dependencies (for testing)
func NewTodoHandlersWithDependencies(
	manager TodoManagerInterface,
	search SearchEngineInterface,
	stats StatsEngineInterface,
	templates TemplateManagerInterface,
) *TodoHandlers {
	return &TodoHandlers{
		manager:   manager,
		search:    search,
		stats:     stats,
		templates: templates,
		createLinker: func(m TodoManagerInterface) TodoLinkerInterface {
			// For testing, we'll need a mock linker
			return nil
		},
	}
}

// Close cleans up resources
func (h *TodoHandlers) Close() error {
	if h.search != nil {
		return h.search.Close()
	}
	return nil
}

// GetBasePathForContext returns the appropriate base path for the context
func (h *TodoHandlers) GetBasePathForContext(ctx context.Context) string {
	// Check if we have a contextual manager wrapper
	if ctxWrapper, ok := h.manager.(*ContextualTodoManagerWrapper); ok {
		manager := ctxWrapper.GetManagerForContext(ctx)
		return manager.GetBasePath()
	}
	
	// Fall back to default manager
	return h.manager.GetBasePath()
}

// getManagerForContext returns the appropriate todo manager for the context
func (h *TodoHandlers) getManagerForContext(ctx context.Context) TodoManagerInterface {
	// Check if we have a contextual manager wrapper
	if ctxWrapper, ok := h.manager.(*ContextualTodoManagerWrapper); ok {
		return ctxWrapper.GetManagerForContext(ctx)
	}
	
	// Otherwise return the default manager
	return h.manager
}

