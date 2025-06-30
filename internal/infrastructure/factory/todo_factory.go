package factory

import (
	"github.com/user/mcp-todo-server/handlers"
	"github.com/user/mcp-todo-server/internal/application"
	"github.com/user/mcp-todo-server/internal/infrastructure/adapters"
	"github.com/user/mcp-todo-server/internal/infrastructure/persistence/filesystem"
)

// CreateTodoManager creates a TodoManagerInterface using the new architecture
func CreateTodoManager(basePath string) handlers.TodoManagerInterface {
	// Create repository
	repo := filesystem.NewTodoRepository(basePath)
	
	// Create service
	service := application.NewTodoService(repo)
	
	// Create adapter
	adapter := adapters.NewTodoManagerAdapter(service, repo, basePath)
	
	return adapter
}