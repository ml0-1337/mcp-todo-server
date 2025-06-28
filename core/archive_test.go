package core

import (
	"testing"
	"time"
	"os"
	"path/filepath"
	"io/ioutil"
	"fmt"
	"sync"
)

// Test 16: todo_archive should move files to correct quarterly folder
func TestArchiveTodoToQuarterlyFolder(t *testing.T) {
	// Create temp directory for test
	tempDir, err := ioutil.TempDir("", "archive-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a test todo
	todo, err := manager.CreateTodo("Test archive functionality", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Get current quarter for verification
	now := time.Now()
	expectedQuarter := fmt.Sprintf("%d-Q%d", now.Year(), (int(now.Month())+2)/3)
	expectedArchivePath := filepath.Join(tempDir, "..", "archive", expectedQuarter, todo.ID+".md")

	// Archive the todo
	err = manager.ArchiveTodo(todo.ID, "")
	if err != nil {
		t.Fatalf("Failed to archive todo: %v", err)
	}

	// Verify original file no longer exists
	originalPath := filepath.Join(tempDir, todo.ID+".md")
	if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
		t.Error("Original todo file should have been moved")
	}

	// Verify file exists in archive
	if _, err := os.Stat(expectedArchivePath); os.IsNotExist(err) {
		t.Errorf("Todo file should exist in archive at %s", expectedArchivePath)
	}

	// Test with quarter override
	t.Run("Archive with quarter override", func(t *testing.T) {
		// Create another test todo
		todo2, err := manager.CreateTodo("Test quarter override", "medium", "bug")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Archive with specific quarter
		overrideQuarter := "2024-Q4"
		err = manager.ArchiveTodo(todo2.ID, overrideQuarter)
		if err != nil {
			t.Fatalf("Failed to archive todo with override: %v", err)
		}

		// Verify file in override quarter
		overridePath := filepath.Join(tempDir, "..", "archive", overrideQuarter, todo2.ID+".md")
		if _, err := os.Stat(overridePath); os.IsNotExist(err) {
			t.Errorf("Todo file should exist in override quarter at %s", overridePath)
		}
	})

	// Test archiving non-existent todo
	t.Run("Archive non-existent todo", func(t *testing.T) {
		err := manager.ArchiveTodo("non-existent-id", "")
		if err == nil {
			t.Error("Archiving non-existent todo should return error")
		}
		if !os.IsNotExist(err) && err.Error() != "todo not found: non-existent-id" {
			t.Errorf("Expected 'todo not found' error, got: %v", err)
		}
	})
}

// Utility function to calculate quarter
func TestGetQuarter(t *testing.T) {
	tests := []struct {
		date     time.Time
		expected string
	}{
		{time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "2025-Q1"},
		{time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC), "2025-Q1"},
		{time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC), "2025-Q1"},
		{time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), "2025-Q2"},
		{time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC), "2025-Q2"},
		{time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), "2025-Q3"},
		{time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC), "2025-Q3"},
		{time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), "2025-Q4"},
		{time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "2025-Q4"},
	}

	for _, test := range tests {
		result := GetQuarter(test.date)
		if result != test.expected {
			t.Errorf("GetQuarter(%v) = %s, expected %s", test.date, result, test.expected)
		}
	}
}

// Test 17: todo_archive should update completed timestamp
func TestArchiveUpdatesCompletedTimestamp(t *testing.T) {
	// Create temp directory for test
	tempDir, err := ioutil.TempDir("", "archive-timestamp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a test todo
	todo, err := manager.CreateTodo("Test timestamp update", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Verify initial state has no completed timestamp
	originalTodo, err := manager.ReadTodo(todo.ID)
	if err != nil {
		t.Fatalf("Failed to read todo: %v", err)
	}
	if !originalTodo.Completed.IsZero() {
		t.Error("New todo should not have completed timestamp")
	}
	if originalTodo.Status != "in_progress" {
		t.Errorf("New todo should have status 'in_progress', got: %s", originalTodo.Status)
	}

	// Archive the todo
	err = manager.ArchiveTodo(todo.ID, "")
	if err != nil {
		t.Fatalf("Failed to archive todo: %v", err)
	}

	// Read archived todo
	quarter := GetQuarter(time.Now())
	archivePath := filepath.Join(filepath.Dir(tempDir), "archive", quarter, todo.ID+".md")
	
	content, err := ioutil.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("Failed to read archived todo: %v", err)
	}

	// Parse archived todo
	archivedTodo, err := manager.parseTodoFile(string(content))
	if err != nil {
		t.Fatalf("Failed to parse archived todo: %v", err)
	}

	// Verify completed timestamp was set
	if archivedTodo.Completed.IsZero() {
		t.Error("Archived todo should have completed timestamp")
	}

	// Verify timestamp format is correct (should be parseable)
	// and year/month/day match today
	now := time.Now()
	if archivedTodo.Completed.Year() != now.Year() {
		t.Errorf("Completed year %d doesn't match current year %d", 
			archivedTodo.Completed.Year(), now.Year())
	}
	if archivedTodo.Completed.Month() != now.Month() {
		t.Errorf("Completed month %v doesn't match current month %v", 
			archivedTodo.Completed.Month(), now.Month())
	}
	if archivedTodo.Completed.Day() != now.Day() {
		t.Errorf("Completed day %d doesn't match current day %d", 
			archivedTodo.Completed.Day(), now.Day())
	}

	// Verify status was updated
	if archivedTodo.Status != "completed" {
		t.Errorf("Archived todo should have status 'completed', got: %s", archivedTodo.Status)
	}

	// Verify other fields remain unchanged
	if archivedTodo.Task != originalTodo.Task {
		t.Error("Task should remain unchanged after archive")
	}
	if archivedTodo.Priority != originalTodo.Priority {
		t.Error("Priority should remain unchanged after archive")
	}
	if archivedTodo.Type != originalTodo.Type {
		t.Error("Type should remain unchanged after archive")
	}
}

// Test 18: Archive operation should be atomic (no data loss)
func TestArchiveOperationIsAtomic(t *testing.T) {
	// Create temp directory for test
	tempDir, err := ioutil.TempDir("", "archive-atomic-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Test scenario 2: Concurrent archive attempts
	t.Run("Concurrent archive attempts", func(t *testing.T) {
		// Create a test todo
		todo, err := manager.CreateTodo("Test concurrent archive", "medium", "bug")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Try to archive concurrently
		var wg sync.WaitGroup
		errors := make([]error, 5)
		successCount := 0
		var successMutex sync.Mutex

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				errors[index] = manager.ArchiveTodo(todo.ID, "")
				if errors[index] == nil {
					successMutex.Lock()
					successCount++
					successMutex.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// Exactly one should succeed
		if successCount != 1 {
			t.Errorf("Expected exactly 1 successful archive, got %d", successCount)
		}

		// Verify todo was archived exactly once
		quarter := GetQuarter(time.Now())
		archivePath := filepath.Join(filepath.Dir(tempDir), "archive", quarter, todo.ID+".md")
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			t.Error("Todo should be archived")
		}

		// Original should not exist
		originalPath := filepath.Join(tempDir, todo.ID+".md")
		if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
			t.Error("Original todo should not exist after successful archive")
		}
	})

	// Test scenario 3: Verify atomicity with write failure
	t.Run("Verify atomicity on write failure", func(t *testing.T) {
		// Create a test todo
		todo, err := manager.CreateTodo("Test write failure", "low", "refactor")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Create archive directory but make it read-only
		quarter := GetQuarter(time.Now())
		archiveDir := filepath.Join(filepath.Dir(tempDir), "archive", quarter)
		os.MkdirAll(archiveDir, 0755)
		
		// Make archive directory read-only to prevent writes
		os.Chmod(archiveDir, 0555)
		defer os.Chmod(archiveDir, 0755) // Restore for cleanup

		// Try to archive - should fail on write
		err = manager.ArchiveTodo(todo.ID, "")
		if err == nil {
			t.Error("Archive should fail when write fails")
		}

		// Original should still exist and be unchanged
		originalPath := filepath.Join(tempDir, todo.ID+".md")
		if _, err := os.Stat(originalPath); os.IsNotExist(err) {
			t.Error("Original todo should still exist after failed archive")
		}

		// Verify todo can still be read and is unchanged
		readTodo, err := manager.ReadTodo(todo.ID)
		if err != nil {
			t.Errorf("Should be able to read todo after failed archive: %v", err)
		}
		if readTodo != nil && readTodo.Status != "in_progress" {
			t.Error("Todo status should remain unchanged after failed archive")
		}
	})
}

// Test 19: Bulk operations should handle errors per-item
func TestBulkArchiveHandlesErrorsPerItem(t *testing.T) {
	// Create temp directory for test
	tempDir, err := ioutil.TempDir("", "bulk-archive-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create test todos
	todo1, _ := manager.CreateTodo("Todo 1 - Valid", "high", "feature")
	todo2, _ := manager.CreateTodo("Todo 2 - Valid", "medium", "bug")
	todo3, _ := manager.CreateTodo("Todo 3 - Will be deleted", "low", "refactor")
	todo4, _ := manager.CreateTodo("Todo 4 - Valid", "high", "feature")

	// Delete todo3's file to simulate missing todo
	os.Remove(filepath.Join(tempDir, todo3.ID+".md"))

	// Create list of IDs including a non-existent one
	todoIDs := []string{
		todo1.ID,
		todo2.ID,
		todo3.ID,           // Missing file
		"non-existent-id",  // Never existed
		todo4.ID,
	}

	// Perform bulk archive
	results := manager.BulkArchiveTodos(todoIDs)

	// Verify results
	if len(results) != len(todoIDs) {
		t.Errorf("Expected %d results, got %d", len(todoIDs), len(results))
	}

	// Check individual results
	expectedSuccess := map[string]bool{
		todo1.ID:          true,
		todo2.ID:          true,
		todo3.ID:          false,
		"non-existent-id": false,
		todo4.ID:          true,
	}

	successCount := 0
	failureCount := 0

	for i, result := range results {
		if result.ID != todoIDs[i] {
			t.Errorf("Result %d has wrong ID: expected %s, got %s", i, todoIDs[i], result.ID)
		}

		if expectedSuccess[result.ID] {
			if !result.Success {
				t.Errorf("Todo %s should have succeeded, but got error: %v", result.ID, result.Error)
			} else {
				successCount++
			}
		} else {
			if result.Success {
				t.Errorf("Todo %s should have failed, but succeeded", result.ID)
			} else {
				failureCount++
				if result.Error == nil {
					t.Errorf("Failed result for %s should have error", result.ID)
				}
			}
		}
	}

	if successCount != 3 {
		t.Errorf("Expected 3 successful archives, got %d", successCount)
	}
	if failureCount != 2 {
		t.Errorf("Expected 2 failed archives, got %d", failureCount)
	}

	// Verify successful todos are actually archived
	quarter := GetQuarter(time.Now())
	archiveDir := filepath.Join(filepath.Dir(tempDir), "archive", quarter)

	for id, shouldSucceed := range expectedSuccess {
		if shouldSucceed && id != "non-existent-id" {
			archivePath := filepath.Join(archiveDir, id+".md")
			if _, err := os.Stat(archivePath); os.IsNotExist(err) {
				t.Errorf("Successfully archived todo %s not found in archive", id)
			}
		}
	}

	// Verify failed todos remain in original location (if they existed)
	originalPath := filepath.Join(tempDir, todo3.ID+".md")
	if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
		t.Error("Failed todo3 should not exist in original location (was already deleted)")
	}
}