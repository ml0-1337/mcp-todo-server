package application

import (
	"context"
	"testing"
	
	"github.com/user/mcp-todo-server/internal/infrastructure/persistence/filesystem"
)

func TestTodoService_ValidateBeforeSaving(t *testing.T) {
	// Test 6: Service should validate todo before saving via repository
	// Input: Various invalid todo configurations
	// Expected: Service rejects invalid todos before persisting
	
	// Arrange
	tmpDir := t.TempDir()
	
	repo := filesystem.NewTodoRepository(tmpDir)
	service := NewTodoService(repo)
	ctx := context.Background()
	
	// Test 1: Empty task should fail
	_, err := service.CreateTodo(ctx, "", "high", "feature")
	if err == nil {
		t.Error("Expected error for empty task, but got nil")
	}
	
	// Test 2: Valid todo should succeed
	todo, err := service.CreateTodo(ctx, "Valid task description", "high", "feature")
	if err != nil {
		t.Errorf("Expected valid todo to succeed, but got error: %v", err)
	}
	
	if todo == nil {
		t.Fatal("Expected todo to be created, but got nil")
	}
	
	// Verify todo has required fields set
	if todo.ID == "" {
		t.Error("Todo ID should not be empty")
	}
	
	if todo.Task != "Valid task description" {
		t.Errorf("Task mismatch: expected 'Valid task description', got '%s'", todo.Task)
	}
	
	if todo.Status != "in_progress" {
		t.Errorf("Status should default to 'in_progress', got '%s'", todo.Status)
	}
	
	if todo.Priority != "high" {
		t.Errorf("Priority mismatch: expected 'high', got '%s'", todo.Priority)
	}
	
	if todo.Type != "feature" {
		t.Errorf("Type mismatch: expected 'feature', got '%s'", todo.Type)
	}
	
	// Test 3: Empty priority should get default
	todo2, err := service.CreateTodo(ctx, "Task with default priority", "", "bug")
	if err != nil {
		t.Errorf("Expected todo with empty priority to succeed with default, but got error: %v", err)
	}
	
	if todo2.Priority != "medium" {
		t.Errorf("Empty priority should default to 'medium', got '%s'", todo2.Priority)
	}
	
	// Test 4: Empty type should get default
	todo3, err := service.CreateTodo(ctx, "Task with default type", "low", "")
	if err != nil {
		t.Errorf("Expected todo with empty type to succeed with default, but got error: %v", err)
	}
	
	if todo3.Type != "task" {
		t.Errorf("Empty type should default to 'task', got '%s'", todo3.Type)
	}
	
	// Test 5: Verify service generates unique IDs
	todo4, err := service.CreateTodo(ctx, "Valid task description", "high", "feature")
	if err != nil {
		t.Errorf("Failed to create second todo with same description: %v", err)
	}
	
	if todo4.ID == todo.ID {
		t.Errorf("Service should generate unique IDs, but got duplicate: %s", todo.ID)
	}
	
	// Verify the ID is based on the task but made unique
	if todo4.ID != "valid-task-description-2" {
		t.Errorf("Expected ID to be 'valid-task-description-2', got '%s'", todo4.ID)
	}
}

func TestTodoService_HandleRepositoryErrors(t *testing.T) {
	// Test 7: Service should handle repository errors gracefully
	// Input: Various error conditions from repository
	// Expected: Service propagates errors with proper context
	
	// Arrange
	tmpDir := t.TempDir()
	
	repo := filesystem.NewTodoRepository(tmpDir)
	service := NewTodoService(repo)
	ctx := context.Background()
	
	// Test 1: GetTodo with non-existent ID should return error
	_, err := service.GetTodo(ctx, "non-existent-todo")
	if err == nil {
		t.Error("Expected error when getting non-existent todo, but got nil")
	}
	
	// Test 2: GetTodoWithContent with non-existent ID should return error
	_, _, err = service.GetTodoWithContent(ctx, "another-non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent todo with content, but got nil")
	}
	
	// Test 3: UpdateTodoStatus on non-existent todo should fail
	err = service.UpdateTodoStatus(ctx, "update-non-existent", "completed")
	if err == nil {
		t.Error("Expected error when updating non-existent todo status, but got nil")
	}
	
	// Test 4: ArchiveTodo on non-existent todo should fail
	err = service.ArchiveTodo(ctx, "archive-non-existent")
	if err == nil {
		t.Error("Expected error when archiving non-existent todo, but got nil")
	}
	
	// Test 5: Create a todo and try to archive it while in_progress
	todo, err := service.CreateTodo(ctx, "Cannot archive in progress", "high", "bug")
	if err != nil {
		t.Fatalf("Failed to create test todo: %v", err)
	}
	
	err = service.ArchiveTodo(ctx, todo.ID)
	if err == nil {
		t.Error("Expected error when archiving incomplete todo, but got nil")
	}
	
	expectedError := "cannot archive incomplete todo"
	if err != nil && err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
	
	// Test 6: Complete the todo and archive should succeed
	err = service.UpdateTodoStatus(ctx, todo.ID, "completed")
	if err != nil {
		t.Errorf("Failed to update todo status: %v", err)
	}
	
	err = service.ArchiveTodo(ctx, todo.ID)
	if err != nil {
		t.Errorf("Failed to archive completed todo: %v", err)
	}
	
	// Test 7: Verify ListTodos handles empty results gracefully
	todos, err := service.ListTodos(ctx, "non-existent-status", "", 0)
	if err != nil {
		t.Errorf("ListTodos should not error on empty results, but got: %v", err)
	}
	
	if len(todos) != 0 {
		t.Errorf("Expected 0 todos with non-existent status, got %d", len(todos))
	}
}