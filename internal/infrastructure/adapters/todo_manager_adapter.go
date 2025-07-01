package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"
	
	"github.com/user/mcp-todo-server/core"
	"github.com/user/mcp-todo-server/internal/application"
	"github.com/user/mcp-todo-server/internal/domain"
	"github.com/user/mcp-todo-server/internal/domain/repository"
)

// TodoManagerAdapter adapts the new TodoService to the old TodoManager interface
type TodoManagerAdapter struct {
	service  *application.TodoService
	repo     repository.TodoRepository
	basePath string
}

// NewTodoManagerAdapter creates a new adapter
func NewTodoManagerAdapter(service *application.TodoService, repo repository.TodoRepository, basePath string) *TodoManagerAdapter {
	return &TodoManagerAdapter{
		service:  service,
		repo:     repo,
		basePath: basePath,
	}
}

// CreateTodo creates a new todo
func (a *TodoManagerAdapter) CreateTodo(task, priority, todoType string) (*core.Todo, error) {
	ctx := context.Background()
	domainTodo, err := a.service.CreateTodo(ctx, task, priority, todoType)
	if err != nil {
		return nil, err
	}
	
	return a.domainToCoreTodo(domainTodo), nil
}

// ReadTodo reads a todo by ID
func (a *TodoManagerAdapter) ReadTodo(id string) (*core.Todo, error) {
	ctx := context.Background()
	domainTodo, err := a.service.GetTodo(ctx, id)
	if err != nil {
		if err == domain.ErrTodoNotFound {
			return nil, fmt.Errorf("todo not found: %s", id)
		}
		return nil, err
	}
	
	return a.domainToCoreTodo(domainTodo), nil
}

// ReadTodoWithContent reads a todo with its content
func (a *TodoManagerAdapter) ReadTodoWithContent(id string) (*core.Todo, string, error) {
	ctx := context.Background()
	domainTodo, content, err := a.service.GetTodoWithContent(ctx, id)
	if err != nil {
		if err == domain.ErrTodoNotFound {
			return nil, "", fmt.Errorf("todo not found: %s", id)
		}
		return nil, "", err
	}
	
	return a.domainToCoreTodo(domainTodo), content, nil
}

// UpdateTodo updates a todo section
func (a *TodoManagerAdapter) UpdateTodo(id, section, operation, content string, metadata map[string]string) error {
	ctx := context.Background()
	
	// For now, we'll implement a simplified version
	// In a full implementation, this would handle all section operations
	if metadata != nil && metadata["status"] != "" {
		return a.service.UpdateTodoStatus(ctx, id, metadata["status"])
	}
	
	// For other updates, we need to implement section handling in the service
	return fmt.Errorf("section updates not yet implemented in adapter")
}

// SaveTodo saves a todo
func (a *TodoManagerAdapter) SaveTodo(todo *core.Todo) error {
	ctx := context.Background()
	domainTodo := a.coreToDomainTodo(todo)
	return a.repo.Save(ctx, domainTodo)
}

// ListTodos lists todos with filters
func (a *TodoManagerAdapter) ListTodos(status, priority string, days int) ([]*core.Todo, error) {
	ctx := context.Background()
	domainTodos, err := a.service.ListTodos(ctx, status, priority, days)
	if err != nil {
		return nil, err
	}
	
	coreTodos := make([]*core.Todo, len(domainTodos))
	for i, dt := range domainTodos {
		coreTodos[i] = a.domainToCoreTodo(dt)
	}
	
	return coreTodos, nil
}

// ReadTodoContent reads the full content of a todo
func (a *TodoManagerAdapter) ReadTodoContent(id string) (string, error) {
	ctx := context.Background()
	return a.repo.GetContent(ctx, id)
}

// ArchiveTodo archives a todo
func (a *TodoManagerAdapter) ArchiveTodo(id string) error {
	ctx := context.Background()
	return a.service.ArchiveTodo(ctx, id)
}

// ArchiveOldTodos archives todos older than specified days
func (a *TodoManagerAdapter) ArchiveOldTodos(days int) (int, error) {
	ctx := context.Background()
	
	// Get todos older than specified days
	todos, err := a.service.ListTodos(ctx, "", "", 0)
	if err != nil {
		return 0, err
	}
	
	count := 0
	cutoff := time.Now().AddDate(0, 0, -days)
	
	for _, todo := range todos {
		if todo.IsCompleted() && todo.Completed.Before(cutoff) {
			if err := a.service.ArchiveTodo(ctx, todo.ID); err == nil {
				count++
			}
		}
	}
	
	return count, nil
}

// FindDuplicateTodos finds duplicate todos
func (a *TodoManagerAdapter) FindDuplicateTodos() ([][]string, error) {
	ctx := context.Background()
	todos, err := a.service.ListTodos(ctx, "", "", 0)
	if err != nil {
		return nil, err
	}
	
	// Group by normalized task name
	groups := make(map[string][]string)
	for _, todo := range todos {
		normalized := strings.ToLower(strings.TrimSpace(todo.Task))
		groups[normalized] = append(groups[normalized], todo.ID)
	}
	
	// Find duplicates
	var duplicates [][]string
	for _, ids := range groups {
		if len(ids) > 1 {
			duplicates = append(duplicates, ids)
		}
	}
	
	return duplicates, nil
}

// GetBasePath returns the base path
func (a *TodoManagerAdapter) GetBasePath() string {
	return a.basePath
}

// domainToCoreTodo converts domain.Todo to core.Todo
func (a *TodoManagerAdapter) domainToCoreTodo(dt *domain.Todo) *core.Todo {
	if dt == nil {
		return nil
	}
	
	// Convert sections
	sections := make(map[string]*core.SectionDefinition)
	for k, v := range dt.Sections {
		sections[k] = &core.SectionDefinition{
			Title:    v.Title,
			Order:    v.Order,
			Schema:   core.SchemaFreeform, // Default to freeform
			Required: false,
			Custom:   false,
			Metadata: v.Metadata,
		}
	}
	
	return &core.Todo{
		ID:        dt.ID,
		Task:      dt.Task,
		Started:   dt.Started,
		Completed: dt.Completed,
		Status:    dt.Status,
		Priority:  dt.Priority,
		Type:      dt.Type,
		ParentID:  dt.ParentID,
		Tags:      dt.Tags,
		Sections:  sections,
	}
}

// coreToDomainTodo converts core.Todo to domain.Todo
func (a *TodoManagerAdapter) coreToDomainTodo(ct *core.Todo) *domain.Todo {
	if ct == nil {
		return nil
	}
	
	// Convert sections
	sections := make(map[string]*domain.SectionDefinition)
	for k, v := range ct.Sections {
		sections[k] = &domain.SectionDefinition{
			Title:    v.Title,
			Order:    v.Order,
			Metadata: v.Metadata,
			// Note: Content, Visible, and Persistent are not in core.SectionDefinition
			// These would need to be stored in metadata or handled differently
		}
	}
	
	return &domain.Todo{
		ID:        ct.ID,
		Task:      ct.Task,
		Started:   ct.Started,
		Completed: ct.Completed,
		Status:    ct.Status,
		Priority:  ct.Priority,
		Type:      ct.Type,
		ParentID:  ct.ParentID,
		Tags:      ct.Tags,
		Sections:  sections,
	}
}