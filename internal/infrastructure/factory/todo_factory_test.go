package factory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/mcp-todo-server/internal/infrastructure/adapters"
)

func TestCreateTodoManager(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		setup    func(t *testing.T, basePath string)
		validate func(t *testing.T, adapter *adapters.TodoManagerAdapter)
	}{
		{
			name:     "creates todo manager with valid path",
			basePath: t.TempDir(),
			setup:    func(t *testing.T, basePath string) {},
			validate: func(t *testing.T, adapter *adapters.TodoManagerAdapter) {
				if adapter == nil {
					t.Fatal("Expected non-nil adapter")
				}
			},
		},
		{
			name:     "creates todo manager with non-existent path",
			basePath: filepath.Join(t.TempDir(), "non-existent"),
			setup:    func(t *testing.T, basePath string) {},
			validate: func(t *testing.T, adapter *adapters.TodoManagerAdapter) {
				if adapter == nil {
					t.Fatal("Expected non-nil adapter")
				}
			},
		},
		{
			name:     "creates todo manager with existing directory structure",
			basePath: t.TempDir(),
			setup: func(t *testing.T, basePath string) {
				// Pre-create the claude/todos structure
				todosDir := filepath.Join(basePath, ".claude", "todos")
				if err := os.MkdirAll(todosDir, 0755); err != nil {
					t.Fatalf("Failed to create todos directory: %v", err)
				}
			},
			validate: func(t *testing.T, adapter *adapters.TodoManagerAdapter) {
				if adapter == nil {
					t.Fatal("Expected non-nil adapter")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t, tt.basePath)
			
			adapter := CreateTodoManager(tt.basePath)
			
			tt.validate(t, adapter)
			
			// Additional validation: test that the adapter is functional
			t.Run("adapter is functional", func(t *testing.T) {
				// Create a todo to verify the adapter works
				todo, err := adapter.CreateTodo("Test todo", "high", "feature")
				if err != nil {
					t.Fatalf("Failed to create todo: %v", err)
				}
				if todo == nil {
					t.Fatal("Expected non-nil todo")
				}
				if todo.Task != "Test todo" {
					t.Errorf("Expected task 'Test todo', got '%s'", todo.Task)
				}
				if todo.Priority != "high" {
					t.Errorf("Expected priority 'high', got '%s'", todo.Priority)
				}
				if todo.Type != "feature" {
					t.Errorf("Expected type 'feature', got '%s'", todo.Type)
				}
				
				// Read the todo back
				readTodo, err := adapter.ReadTodo(todo.ID)
				if err != nil {
					t.Fatalf("Failed to read todo: %v", err)
				}
				if readTodo.ID != todo.ID {
					t.Errorf("Expected todo ID '%s', got '%s'", todo.ID, readTodo.ID)
				}
			})
		})
	}
}

func TestFactoryIntegration(t *testing.T) {
	// Test that multiple managers can work with the same base path
	basePath := t.TempDir()
	
	// Create first manager and add a todo
	manager1 := CreateTodoManager(basePath)
	todo1, err := manager1.CreateTodo("First todo", "medium", "bug")
	if err != nil {
		t.Fatalf("Failed to create todo with manager1: %v", err)
	}
	
	// Create second manager and verify it can see the todo
	manager2 := CreateTodoManager(basePath)
	readTodo, err := manager2.ReadTodo(todo1.ID)
	if err != nil {
		t.Fatalf("Failed to read todo with manager2: %v", err)
	}
	if readTodo.Task != "First todo" {
		t.Errorf("Expected task 'First todo', got '%s'", readTodo.Task)
	}
	
	// Create todo with second manager
	todo2, err := manager2.CreateTodo("Second todo", "low", "refactor")
	if err != nil {
		t.Fatalf("Failed to create todo with manager2: %v", err)
	}
	
	// Verify first manager can see both todos
	readTodo1, err := manager1.ReadTodo(todo1.ID)
	if err != nil {
		t.Fatalf("Failed to read todo1 with manager1: %v", err)
	}
	if readTodo1.Task != "First todo" {
		t.Errorf("Expected task 'First todo', got '%s'", readTodo1.Task)
	}
	
	readTodo2, err := manager1.ReadTodo(todo2.ID)
	if err != nil {
		t.Fatalf("Failed to read todo2 with manager1: %v", err)
	}
	if readTodo2.Task != "Second todo" {
		t.Errorf("Expected task 'Second todo', got '%s'", readTodo2.Task)
	}
}