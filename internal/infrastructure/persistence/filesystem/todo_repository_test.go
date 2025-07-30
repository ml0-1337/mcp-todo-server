package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	tmpDir := t.TempDir()

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
	err := repo.Save(ctx, todo)

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
	tmpDir := t.TempDir()

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
			ID:        "todo-completed",
			Task:      "Completed task",
			Started:   time.Now().Add(-24 * time.Hour),
			Completed: time.Now(),
			Status:    "completed",
			Priority:  "high",
			Type:      "feature",
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
	tmpDir := t.TempDir()

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
	err := repo.Save(ctx, originalTodo)
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
	tmpDir := t.TempDir()

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

func TestRepository_NotFoundError(t *testing.T) {
	// Test 5: Repository should return error when todo not found
	// Input: Request for non-existent todo ID
	// Expected: Returns domain.ErrTodoNotFound

	// Arrange
	tmpDir := t.TempDir()

	repo := NewTodoRepository(tmpDir)
	ctx := context.Background()

	// Act - Try to find non-existent todo
	_, err := repo.FindByID(ctx, "non-existent-todo-id")

	// Assert - Should return ErrTodoNotFound
	if err == nil {
		t.Error("Expected error for non-existent todo, but got nil")
	}

	if err != domain.ErrTodoNotFound {
		t.Errorf("Expected domain.ErrTodoNotFound, got: %v", err)
	}

	// Act - Try to find with content
	_, _, err = repo.FindByIDWithContent(ctx, "another-non-existent-id")

	// Assert - Should also return ErrTodoNotFound
	if err == nil {
		t.Error("Expected error for non-existent todo with content, but got nil")
	}

	if err != domain.ErrTodoNotFound {
		t.Errorf("Expected domain.ErrTodoNotFound for content retrieval, got: %v", err)
	}

	// Act - Try to delete non-existent todo
	err = repo.Delete(ctx, "delete-non-existent")

	// Assert - Delete should also return ErrTodoNotFound
	if err == nil {
		t.Error("Expected error for deleting non-existent todo, but got nil")
	}

	if err != domain.ErrTodoNotFound {
		t.Errorf("Expected domain.ErrTodoNotFound for delete, got: %v", err)
	}

	// Act - Try to get content of non-existent todo
	_, err = repo.GetContent(ctx, "content-non-existent")

	// Assert - GetContent should return ErrTodoNotFound
	if err == nil {
		t.Error("Expected error for getting content of non-existent todo, but got nil")
	}

	if err != domain.ErrTodoNotFound {
		t.Errorf("Expected domain.ErrTodoNotFound for GetContent, got: %v", err)
	}
}

func TestRepository_Archive(t *testing.T) {
	// Test: Repository should archive todos correctly
	tmpDir := t.TempDir()
	repo := NewTodoRepository(tmpDir)
	ctx := context.Background()

	// Create a todo to archive
	todo := &domain.Todo{
		ID:       "test-archive-1",
		Task:     "Task to archive",
		Started:  time.Now(),
		Status:   "completed",
		Priority: "medium",
		Type:     "task",
	}

	err := repo.Save(ctx, todo)
	if err != nil {
		t.Fatalf("Failed to save todo: %v", err)
	}

	// Archive the todo
	archivePath := "2025/07/29"
	err = repo.Archive(ctx, todo.ID, archivePath)
	if err != nil {
		t.Fatalf("Failed to archive todo: %v", err)
	}

	// Verify original file is gone
	originalPath := filepath.Join(tmpDir, "test-archive-1.md")
	if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
		t.Error("Original todo file should not exist after archiving")
	}

	// Verify archive file exists
	archiveFilePath := filepath.Join(tmpDir, "archive", archivePath, "test-archive-1.md")
	if _, err := os.Stat(archiveFilePath); os.IsNotExist(err) {
		t.Error("Archive file should exist")
	}
}

func TestRepository_ListWithFilters(t *testing.T) {
	// Test: Repository should filter todos by priority and days
	tmpDir := t.TempDir()
	repo := NewTodoRepository(tmpDir)
	ctx := context.Background()

	// Create todos with different attributes
	now := time.Now()
	todos := []*domain.Todo{
		{
			ID:       "todo-high-recent",
			Task:     "High priority recent",
			Started:  now.Add(-24 * time.Hour),
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
		},
		{
			ID:       "todo-high-old",
			Task:     "High priority old",
			Started:  now.Add(-10 * 24 * time.Hour),
			Status:   "in_progress",
			Priority: "high",
			Type:     "bug",
		},
		{
			ID:       "todo-low-recent",
			Task:     "Low priority recent",
			Started:  now.Add(-2 * 24 * time.Hour),
			Status:   "in_progress",
			Priority: "low",
			Type:     "task",
		},
		{
			ID:       "todo-with-parent",
			Task:     "Child todo",
			Started:  now,
			Status:   "in_progress",
			Priority: "medium",
			Type:     "task",
			ParentID: "parent-123",
		},
	}

	for _, todo := range todos {
		if err := repo.Save(ctx, todo); err != nil {
			t.Fatalf("Failed to save todo %s: %v", todo.ID, err)
		}
	}

	// Test priority filter
	result, err := repo.List(ctx, repository.ListFilters{Priority: "high"})
	if err != nil {
		t.Fatalf("Failed to list with priority filter: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 high priority todos, got %d", len(result))
	}

	// Test days filter (last 7 days)
	result, err = repo.List(ctx, repository.ListFilters{Days: 7})
	if err != nil {
		t.Fatalf("Failed to list with days filter: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("Expected 3 recent todos (within 7 days), got %d", len(result))
	}

	// Test parent ID filter
	result, err = repo.List(ctx, repository.ListFilters{ParentID: "parent-123"})
	if err != nil {
		t.Fatalf("Failed to list with parent ID filter: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 todo with parent, got %d", len(result))
	}
	if len(result) > 0 && result[0].ID != "todo-with-parent" {
		t.Errorf("Expected todo-with-parent, got %s", result[0].ID)
	}

	// Test combined filters
	result, err = repo.List(ctx, repository.ListFilters{
		Priority: "high",
		Days:     5,
	})
	if err != nil {
		t.Fatalf("Failed to list with combined filters: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 high priority recent todo, got %d", len(result))
	}
}

func TestRepository_SaveValidationError(t *testing.T) {
	// Test: Repository should return validation error for invalid todo
	tmpDir := t.TempDir()
	repo := NewTodoRepository(tmpDir)
	ctx := context.Background()

	// Create invalid todo (empty ID)
	invalidTodo := &domain.Todo{
		ID:       "",
		Task:     "Invalid todo",
		Started:  time.Now(),
		Status:   "in_progress",
		Priority: "high",
		Type:     "task",
	}

	err := repo.Save(ctx, invalidTodo)
	if err == nil {
		t.Error("Expected validation error for todo with empty ID")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestRepository_FindByIDWithContent(t *testing.T) {
	// Test: Repository should return todo with full content
	tmpDir := t.TempDir()
	repo := NewTodoRepository(tmpDir)
	ctx := context.Background()

	// Create todo with sections
	todo := &domain.Todo{
		ID:       "test-with-content",
		Task:     "Todo with content",
		Started:  time.Now(),
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
		Sections: map[string]*domain.SectionDefinition{
			"description": {
				Title:   "Description",
				Content: "This is a detailed description",
				Order:   1,
			},
			"checklist": {
				Title:   "Checklist",
				Content: "- [ ] Item 1\n- [x] Item 2",
				Order:   2,
			},
		},
	}

	err := repo.Save(ctx, todo)
	if err != nil {
		t.Fatalf("Failed to save todo: %v", err)
	}

	// Retrieve with content
	retrievedTodo, content, err := repo.FindByIDWithContent(ctx, "test-with-content")
	if err != nil {
		t.Fatalf("Failed to find todo with content: %v", err)
	}

	// Verify todo is returned
	if retrievedTodo.ID != todo.ID {
		t.Errorf("Expected todo ID %s, got %s", todo.ID, retrievedTodo.ID)
	}

	// Verify content contains expected sections
	if !strings.Contains(content, "# Todo with content") {
		t.Error("Content should contain task as heading")
	}
	if !strings.Contains(content, "## Description") {
		t.Error("Content should contain Description section")
	}
	if !strings.Contains(content, "This is a detailed description") {
		t.Error("Content should contain description text")
	}
	if !strings.Contains(content, "## Checklist") {
		t.Error("Content should contain Checklist section")
	}
	if !strings.Contains(content, "- [ ] Item 1") {
		t.Error("Content should contain checklist items")
	}
}

func TestRepository_ErrorHandling(t *testing.T) {
	// Test: Repository should handle file system errors gracefully
	tmpDir := t.TempDir()
	ctx := context.Background()

	// Test save to read-only directory
	t.Run("save to read-only directory", func(t *testing.T) {
		// Create a read-only directory
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		os.Mkdir(readOnlyDir, 0755)
		os.Chmod(readOnlyDir, 0555)
		defer os.Chmod(readOnlyDir, 0755) // Restore for cleanup

		readOnlyRepo := NewTodoRepository(readOnlyDir)
		todo := &domain.Todo{
			ID:       "test-readonly",
			Task:     "Test task",
			Started:  time.Now(),
			Status:   "in_progress",
			Priority: "medium",
			Type:     "task",
		}

		err := readOnlyRepo.Save(ctx, todo)
		if err == nil {
			t.Error("Expected error when saving to read-only directory")
		}
	})

	// Test list with invalid directory
	t.Run("list non-existent directory", func(t *testing.T) {
		nonExistentRepo := NewTodoRepository("/path/that/does/not/exist")
		result, err := nonExistentRepo.List(ctx, repository.ListFilters{})

		// Should return empty list, not error
		if err != nil {
			t.Errorf("Expected no error for non-existent directory, got: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected empty list, got %d items", len(result))
		}
	})
}

func TestRepository_UpdateContent(t *testing.T) {
	// Test: UpdateContent should return not implemented error
	tmpDir := t.TempDir()
	repo := NewTodoRepository(tmpDir)
	ctx := context.Background()

	err := repo.UpdateContent(ctx, "any-id", "section", "content")
	if err == nil {
		t.Error("Expected error for UpdateContent")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("Expected 'not implemented' error, got: %v", err)
	}
}
