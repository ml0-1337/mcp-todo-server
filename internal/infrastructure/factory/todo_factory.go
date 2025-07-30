package factory

import (
	"github.com/user/mcp-todo-server/internal/application"
	"github.com/user/mcp-todo-server/internal/infrastructure/adapters"
	"github.com/user/mcp-todo-server/internal/infrastructure/persistence/filesystem"
)

// CreateTodoManager creates a todo manager using the new architecture
// Returns the adapter which implements the old interface for compatibility
func CreateTodoManager(basePath string) *adapters.TodoManagerAdapter {
	// Create repository
	repo := filesystem.NewTodoRepository(basePath)

	// Create service
	service := application.NewTodoService(repo)

	// Create adapter
	adapter := adapters.NewTodoManagerAdapter(service, repo, basePath)

	return adapter
}
