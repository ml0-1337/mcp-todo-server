package adapters

import (
	"testing"
	
	"github.com/user/mcp-todo-server/internal/application"
	"github.com/user/mcp-todo-server/internal/infrastructure/persistence/filesystem"
)

func TestTodoManagerAdapter_CreateAndRead(t *testing.T) {
	// Test that the adapter correctly bridges between old and new interfaces
	
	// Arrange
	tmpDir := t.TempDir()
	
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Act - Create todo
	todo, err := adapter.CreateTodo("Test adapter task", "high", "feature")
	
	// Assert - Create succeeded
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}
	
	if todo.Task != "Test adapter task" {
		t.Errorf("Task mismatch: expected 'Test adapter task', got '%s'", todo.Task)
	}
	
	// Act - Read todo
	retrieved, err := adapter.ReadTodo(todo.ID)
	
	// Assert - Read succeeded
	if err != nil {
		t.Fatalf("Failed to read todo: %v", err)
	}
	
	if retrieved.ID != todo.ID {
		t.Errorf("ID mismatch: expected '%s', got '%s'", todo.ID, retrieved.ID)
	}
	
	if retrieved.Task != todo.Task {
		t.Errorf("Task mismatch: expected '%s', got '%s'", todo.Task, retrieved.Task)
	}
}