package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
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

func TestRepository_UpdatePreservingMetadata(t *testing.T) {
	// Test 3: Repository should update todo content preserving metadata
	// Input: Existing todo with metadata, then update only status
	// Expected: Status changes but other fields remain unchanged
	
	// Arrange
	tmpDir, err := os.MkdirTemp("", "todo-repo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	repo := NewTodoRepository(tmpDir)
	ctx := context.Background()
	
	// Create initial todo with all fields
	originalTodo := &domain.Todo{
		ID:       "test-update-1",
		Task:     "Original task description",
		Started:  time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
		Tags:     []string{"important", "backend"},
		Sections: map[string]*domain.SectionDefinition{
			"notes": {
				Title:    "Implementation Notes",
				Content:  "Original notes content",
				Order:    1,
				Metadata: map[string]interface{}{"author": "test"},
			},
		},
	}
	
	// Save original
	err = repo.Save(ctx, originalTodo)
	if err != nil {
		t.Fatalf("Failed to save original todo: %v", err)
	}
	
	// Act - Update only the status
	updatedTodo, err := repo.FindByID(ctx, "test-update-1")
	if err != nil {
		t.Fatalf("Failed to retrieve todo for update: %v", err)
	}
	
	updatedTodo.Status = "completed"
	updatedTodo.Completed = time.Now()
	
	err = repo.Save(ctx, updatedTodo)
	if err != nil {
		t.Fatalf("Failed to save updated todo: %v", err)
	}
	
	// Assert - Retrieve and verify
	final, err := repo.FindByID(ctx, "test-update-1")
	if err != nil {
		t.Fatalf("Failed to retrieve updated todo: %v", err)
	}
	
	// Check that status was updated
	if final.Status != "completed" {
		t.Errorf("Status not updated: expected 'completed', got '%s'", final.Status)
	}
	
	// Check that other fields were preserved
	if final.Task != originalTodo.Task {
		t.Errorf("Task changed unexpectedly: expected '%s', got '%s'", originalTodo.Task, final.Task)
	}
	
	if !final.Started.Equal(originalTodo.Started) {
		t.Errorf("Started time changed unexpectedly: expected %v, got %v", originalTodo.Started, final.Started)
	}
	
	if final.Priority != originalTodo.Priority {
		t.Errorf("Priority changed unexpectedly: expected '%s', got '%s'", originalTodo.Priority, final.Priority)
	}
	
	if final.Type != originalTodo.Type {
		t.Errorf("Type changed unexpectedly: expected '%s', got '%s'", originalTodo.Type, final.Type)
	}
	
	if len(final.Tags) != len(originalTodo.Tags) {
		t.Errorf("Tags changed unexpectedly: expected %d tags, got %d", len(originalTodo.Tags), len(final.Tags))
	}
	
	// Check sections were preserved
	if final.Sections["notes"] == nil {
		t.Error("Notes section was lost during update")
	} else {
		if final.Sections["notes"].Title != originalTodo.Sections["notes"].Title {
			t.Errorf("Section title changed: expected '%s', got '%s'", 
				originalTodo.Sections["notes"].Title, 
				final.Sections["notes"].Title)
		}
		if final.Sections["notes"].Metadata["author"] != "test" {
			t.Error("Section metadata was lost during update")
		}
	}
}

func TestRepository_ConcurrentAccess(t *testing.T) {
	// Test 4: Repository should handle concurrent access safely
	// Input: Multiple goroutines reading and writing simultaneously
	// Expected: All operations complete without race conditions or data corruption
	
	// Arrange
	tmpDir, err := os.MkdirTemp("", "todo-repo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	repo := NewTodoRepository(tmpDir)
	ctx := context.Background()
	
	// Create initial todos
	for i := 0; i < 5; i++ {
		todo := &domain.Todo{
			ID:       fmt.Sprintf("concurrent-todo-%d", i),
			Task:     fmt.Sprintf("Concurrent task %d", i),
			Started:  time.Now(),
			Status:   "in_progress",
			Priority: "medium",
			Type:     "task",
		}
		if err := repo.Save(ctx, todo); err != nil {
			t.Fatalf("Failed to create initial todo %d: %v", i, err)
		}
	}
	
	// Act - Concurrent operations
	var wg sync.WaitGroup
	errors := make(chan error, 100)
	
	// Multiple readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Read random todos multiple times
			for j := 0; j < 5; j++ {
				todoID := fmt.Sprintf("concurrent-todo-%d", id%5)
				_, err := repo.FindByID(ctx, todoID)
				if err != nil {
					errors <- fmt.Errorf("reader %d failed: %w", id, err)
				}
				
				// Also do list operations
				_, err = repo.List(ctx, repository.ListFilters{Status: "in_progress"})
				if err != nil {
					errors <- fmt.Errorf("reader %d list failed: %w", id, err)
				}
			}
		}(i)
	}
	
	// Multiple writers updating different todos
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			todoID := fmt.Sprintf("concurrent-todo-%d", id)
			
			// Update the todo multiple times
			for j := 0; j < 3; j++ {
				todo, err := repo.FindByID(ctx, todoID)
				if err != nil {
					errors <- fmt.Errorf("writer %d read failed: %w", id, err)
					return
				}
				
				// Modify and save
				todo.Priority = "high"
				todo.Tags = append(todo.Tags, fmt.Sprintf("update-%d", j))
				
				if err := repo.Save(ctx, todo); err != nil {
					errors <- fmt.Errorf("writer %d save failed: %w", id, err)
				}
				
				// Small delay to increase chance of contention
				time.Sleep(time.Millisecond)
			}
		}(i)
	}
	
	// Wait for all operations to complete
	wg.Wait()
	close(errors)
	
	// Assert - Check for errors
	var errorCount int
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
		errorCount++
	}
	
	if errorCount > 0 {
		t.Fatalf("Had %d errors during concurrent operations", errorCount)
	}
	
	// Verify data integrity - all todos should still exist
	for i := 0; i < 5; i++ {
		todoID := fmt.Sprintf("concurrent-todo-%d", i)
		todo, err := repo.FindByID(ctx, todoID)
		if err != nil {
			t.Errorf("Todo %s missing after concurrent operations: %v", todoID, err)
			continue
		}
		
		// Should have been updated
		if todo.Priority != "high" {
			t.Errorf("Todo %s not properly updated, priority is %s", todoID, todo.Priority)
		}
		
		// Should have update tags
		if len(todo.Tags) != 3 {
			t.Errorf("Todo %s has %d tags, expected 3", todoID, len(todo.Tags))
		}
	}
}