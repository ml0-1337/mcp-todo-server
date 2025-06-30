package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
	
	"github.com/user/mcp-todo-server/internal/domain"
	"github.com/user/mcp-todo-server/internal/domain/repository"
)

func TestRepository_SaveAndRetrieveByID(t *testing.T) {
	// Test 1: Repository should save a todo and retrieve it by ID
	// Input: A valid todo with all fields populated
	// Expected: Todo is saved and can be retrieved with same data
	
	// Arrange
	tmpDir, err := os.MkdirTemp("", "todo-repo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	repo := NewTodoRepository(tmpDir)
	ctx := context.Background()
	
	todo := &domain.Todo{
		ID:       "test-todo-1",
		Task:     "Test saving and retrieving",
		Started:  time.Now(),
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
		Tags:     []string{"test", "repository"},
		Sections: map[string]*domain.SectionDefinition{
			"test": {
				Title:   "Test Section",
				Content: "Test content",
			},
		},
	}
	
	// Act - Save
	err = repo.Save(ctx, todo)
	
	// Assert - Save should succeed
	if err != nil {
		t.Fatalf("Failed to save todo: %v", err)
	}
	
	// Verify file was created
	expectedPath := filepath.Join(tmpDir, "test-todo-1.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Todo file was not created at expected path: %s", expectedPath)
	}
	
	// Act - Retrieve
	retrieved, err := repo.FindByID(ctx, "test-todo-1")
	
	// Assert - Retrieve should succeed
	if err != nil {
		t.Fatalf("Failed to retrieve todo: %v", err)
	}
	
	// Assert - Data should match
	if retrieved.ID != todo.ID {
		t.Errorf("ID mismatch: expected %s, got %s", todo.ID, retrieved.ID)
	}
	if retrieved.Task != todo.Task {
		t.Errorf("Task mismatch: expected %s, got %s", todo.Task, retrieved.Task)
	}
	if retrieved.Status != todo.Status {
		t.Errorf("Status mismatch: expected %s, got %s", todo.Status, retrieved.Status)
	}
	if retrieved.Priority != todo.Priority {
		t.Errorf("Priority mismatch: expected %s, got %s", todo.Priority, retrieved.Priority)
	}
	if retrieved.Type != todo.Type {
		t.Errorf("Type mismatch: expected %s, got %s", todo.Type, retrieved.Type)
	}
	if len(retrieved.Tags) != len(todo.Tags) {
		t.Errorf("Tags length mismatch: expected %d, got %d", len(todo.Tags), len(retrieved.Tags))
	}
	if retrieved.Sections["test"] == nil {
		t.Error("Expected test section not found")
	}
}

func TestRepository_ListFilteredByStatus(t *testing.T) {
	// Test 2: Repository should list todos filtered by status
	// Input: Multiple todos with different statuses
	// Expected: Only todos matching the filter status are returned
	
	// Arrange
	tmpDir, err := os.MkdirTemp("", "todo-repo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	repo := NewTodoRepository(tmpDir)
	ctx := context.Background()
	
	// Create todos with different statuses
	todos := []*domain.Todo{
		{
			ID:       "todo-in-progress-1",
			Task:     "In progress task 1",
			Started:  time.Now(),
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
		},
		{
			ID:       "todo-in-progress-2",
			Task:     "In progress task 2",
			Started:  time.Now(),
			Status:   "in_progress",
			Priority: "medium",
			Type:     "bug",
		},
		{
			ID:       "todo-completed",
			Task:     "Completed task",
			Started:  time.Now().Add(-24 * time.Hour),
			Completed: time.Now(),
			Status:   "completed",
			Priority: "high",
			Type:     "feature",
		},
		{
			ID:       "todo-blocked",
			Task:     "Blocked task",
			Started:  time.Now().Add(-48 * time.Hour),
			Status:   "blocked",
			Priority: "low",
			Type:     "research",
		},
	}
	
	// Save all todos
	for _, todo := range todos {
		if err := repo.Save(ctx, todo); err != nil {
			t.Fatalf("Failed to save todo %s: %v", todo.ID, err)
		}
	}
	
	// Act - List only in_progress todos
	filters := repository.ListFilters{
		Status: "in_progress",
	}
	result, err := repo.List(ctx, filters)
	
	// Assert
	if err != nil {
		t.Fatalf("Failed to list todos: %v", err)
	}
	
	if len(result) != 2 {
		t.Errorf("Expected 2 in_progress todos, got %d", len(result))
	}
	
	// Verify all returned todos have in_progress status
	for _, todo := range result {
		if todo.Status != "in_progress" {
			t.Errorf("Expected status 'in_progress', got '%s' for todo %s", todo.Status, todo.ID)
		}
	}
	
	// Act - List completed todos
	filters.Status = "completed"
	result, err = repo.List(ctx, filters)
	
	if err != nil {
		t.Fatalf("Failed to list completed todos: %v", err)
	}
	
	if len(result) != 1 {
		t.Errorf("Expected 1 completed todo, got %d", len(result))
	}
	
	if len(result) > 0 && result[0].Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", result[0].Status)
	}
}