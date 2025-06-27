package core

import (
	"testing"
	"strings"
)

// Test 4: todo_create should generate unique IDs for new todos
func TestUniqueIDGeneration(t *testing.T) {
	// Create a todo manager
	manager := NewTodoManager("/tmp/test-todos")
	
	// Create multiple todos
	todos := []string{
		"Implement authentication",
		"Add user management", 
		"Create API documentation",
		"Write unit tests",
		"Deploy to production",
	}
	
	// Track generated IDs
	generatedIDs := make(map[string]bool)
	
	for _, task := range todos {
		todo, err := manager.CreateTodo(task, "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}
		
		// Check ID is not empty
		if todo.ID == "" {
			t.Error("Generated todo ID is empty")
		}
		
		// Check ID format (should be kebab-case)
		if strings.Contains(todo.ID, " ") || strings.Contains(todo.ID, "_") {
			t.Errorf("ID should be kebab-case, got: %s", todo.ID)
		}
		
		// Check ID is unique
		if generatedIDs[todo.ID] {
			t.Errorf("Duplicate ID generated: %s", todo.ID)
		}
		generatedIDs[todo.ID] = true
		
		// Check ID is derived from task
		if !strings.Contains(todo.ID, strings.ToLower(strings.Split(task, " ")[0])) {
			t.Errorf("ID should be derived from task. Task: %s, ID: %s", task, todo.ID)
		}
	}
	
	// Test duplicate task names generate unique IDs
	t.Run("Duplicate tasks get unique IDs", func(t *testing.T) {
		task := "Duplicate task"
		
		todo1, err := manager.CreateTodo(task, "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create first todo: %v", err)
		}
		
		todo2, err := manager.CreateTodo(task, "high", "feature") 
		if err != nil {
			t.Fatalf("Failed to create second todo: %v", err)
		}
		
		if todo1.ID == todo2.ID {
			t.Errorf("Duplicate tasks should have unique IDs. Got: %s and %s", todo1.ID, todo2.ID)
		}
	})
}