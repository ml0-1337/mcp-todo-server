package repository

import (
	"context"
	"github.com/user/mcp-todo-server/internal/domain"
)

// TodoRepository defines the interface for todo persistence
type TodoRepository interface {
	// Save creates or updates a todo
	Save(ctx context.Context, todo *domain.Todo) error
	
	// FindByID retrieves a todo by its ID
	FindByID(ctx context.Context, id string) (*domain.Todo, error)
	
	// FindByIDWithContent retrieves a todo and its full content
	FindByIDWithContent(ctx context.Context, id string) (*domain.Todo, string, error)
	
	// List retrieves todos based on filters
	List(ctx context.Context, filters ListFilters) ([]*domain.Todo, error)
	
	// Delete removes a todo
	Delete(ctx context.Context, id string) error
	
	// Archive moves a todo to archive
	Archive(ctx context.Context, id string, archivePath string) error
	
	// UpdateContent updates a specific section of a todo
	UpdateContent(ctx context.Context, id string, section string, content string) error
	
	// GetContent retrieves the full content of a todo
	GetContent(ctx context.Context, id string) (string, error)
}

// ListFilters contains filtering options for listing todos
type ListFilters struct {
	Status   string
	Priority string
	Days     int
	ParentID string
}

// TodoRepositoryFactory creates new instances of TodoRepository
type TodoRepositoryFactory interface {
	Create(basePath string) TodoRepository
}