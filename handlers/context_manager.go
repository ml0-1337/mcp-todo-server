package handlers

import (
	"context"
	"log"
	"sync"
	
	"github.com/user/mcp-todo-server/core"
	ctxkeys "github.com/user/mcp-todo-server/internal/context"
	"github.com/user/mcp-todo-server/utils"
)

// ContextualTodoManager manages todo managers for different working directories
type ContextualTodoManager struct {
	defaultPath string
	managers    map[string]*core.TodoManager
	mu          sync.RWMutex
}

// NewContextualTodoManager creates a new contextual todo manager
func NewContextualTodoManager(defaultPath string) *ContextualTodoManager {
	return &ContextualTodoManager{
		defaultPath: defaultPath,
		managers:    make(map[string]*core.TodoManager),
	}
}

// GetManagerForContext returns the appropriate todo manager for the context
func (c *ContextualTodoManager) GetManagerForContext(ctx context.Context) *core.TodoManager {
	var todoPath string
	var err error
	
	// Try to get working directory from context (HTTP header)
	if workingDir, ok := ctx.Value(ctxkeys.WorkingDirectoryKey).(string); ok && workingDir != "" {
		todoPath, err = utils.ResolveTodoPathFromWorkingDir(workingDir)
		if err != nil {
			log.Printf("Failed to resolve todo path from working directory %s: %v", workingDir, err)
			todoPath = c.defaultPath
		}
	} else {
		// Use default path
		todoPath = c.defaultPath
	}
	
	// Check if we already have a manager for this path
	c.mu.RLock()
	if manager, exists := c.managers[todoPath]; exists {
		c.mu.RUnlock()
		return manager
	}
	c.mu.RUnlock()
	
	// Create new manager for this path
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Double-check after acquiring write lock
	if manager, exists := c.managers[todoPath]; exists {
		return manager
	}
	
	// Create and cache new manager
	manager := core.NewTodoManager(todoPath)
	c.managers[todoPath] = manager
	log.Printf("Created new todo manager for path: %s", todoPath)
	
	return manager
}

// CreateTodoWithContext creates a todo using the appropriate manager for the context
func (h *TodoHandlers) CreateTodoWithContext(ctx context.Context, task, priority, todoType string) (*core.Todo, error) {
	// Check if we have a contextual manager
	if ctxManager, ok := h.manager.(*ContextualTodoManagerWrapper); ok {
		manager := ctxManager.GetManagerForContext(ctx)
		return manager.CreateTodo(task, priority, todoType)
	}
	
	// Fall back to default manager
	return h.manager.CreateTodo(task, priority, todoType)
}

// ContextualTodoManagerWrapper wraps ContextualTodoManager to implement TodoManagerInterface
type ContextualTodoManagerWrapper struct {
	*ContextualTodoManager
	defaultManager *core.TodoManager
}

// NewContextualTodoManagerWrapper creates a new wrapper
func NewContextualTodoManagerWrapper(defaultPath string) *ContextualTodoManagerWrapper {
	ctxManager := NewContextualTodoManager(defaultPath)
	defaultManager := core.NewTodoManager(defaultPath)
	
	return &ContextualTodoManagerWrapper{
		ContextualTodoManager: ctxManager,
		defaultManager:        defaultManager,
	}
}

// Implement TodoManagerInterface methods (delegate to default manager for non-context methods)

func (w *ContextualTodoManagerWrapper) CreateTodo(task, priority, todoType string) (*core.Todo, error) {
	return w.defaultManager.CreateTodo(task, priority, todoType)
}

func (w *ContextualTodoManagerWrapper) ReadTodo(id string) (*core.Todo, error) {
	return w.defaultManager.ReadTodo(id)
}

func (w *ContextualTodoManagerWrapper) ReadTodoWithContent(id string) (*core.Todo, string, error) {
	return w.defaultManager.ReadTodoWithContent(id)
}

func (w *ContextualTodoManagerWrapper) UpdateTodo(id, section, operation, content string, metadata map[string]string) error {
	return w.defaultManager.UpdateTodo(id, section, operation, content, metadata)
}

func (w *ContextualTodoManagerWrapper) SaveTodo(todo *core.Todo) error {
	return w.defaultManager.SaveTodo(todo)
}

func (w *ContextualTodoManagerWrapper) ListTodos(status, priority string, days int) ([]*core.Todo, error) {
	return w.defaultManager.ListTodos(status, priority, days)
}

func (w *ContextualTodoManagerWrapper) ReadTodoContent(id string) (string, error) {
	return w.defaultManager.ReadTodoContent(id)
}

func (w *ContextualTodoManagerWrapper) ArchiveTodo(id, quarter string) error {
	return w.defaultManager.ArchiveTodo(id, quarter)
}

func (w *ContextualTodoManagerWrapper) ArchiveOldTodos(days int) (int, error) {
	return w.defaultManager.ArchiveOldTodos(days)
}

func (w *ContextualTodoManagerWrapper) FindDuplicateTodos() ([][]string, error) {
	return w.defaultManager.FindDuplicateTodos()
}

func (w *ContextualTodoManagerWrapper) GetBasePath() string {
	return w.defaultManager.GetBasePath()
}