package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
	
	"github.com/user/mcp-todo-server/internal/domain"
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