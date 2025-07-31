package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

// Test 16: todo_archive should move files to correct quarterly folder
func TestArchiveTodoToQuarterlyFolder(t *testing.T) {
	// Setup test environment
	manager, tempDir, cleanup := SetupTestTodoManager(t)
	defer cleanup()

	// Create a test todo
	todo := CreateTestTodo(t, manager, "Test archive functionality", "high", "feature")

	// Get expected archive path
	expectedArchivePath := GetArchivePath(tempDir, todo, "")

	// Archive the todo
	err := manager.ArchiveTodo(todo.ID)
	if err != nil {
		t.Fatalf("Failed to archive todo: %v", err)
	}

	// Verify original file no longer exists
	VerifyTodoNotExists(t, tempDir, todo.ID)

	// Verify file exists in archive
	if _, err := os.Stat(expectedArchivePath); os.IsNotExist(err) {
		t.Errorf("Todo file should exist in archive at %s", expectedArchivePath)
	}

	// Test with quarter override (deprecated, should use daily path)
	t.Run("Archive with quarter override", func(t *testing.T) {
		// Create another test todo
		todo2, err := manager.CreateTodo("Test quarter override", "medium", "bug")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Archive todo
		err = manager.ArchiveTodo(todo2.ID)
		if err != nil {
			t.Fatalf("Failed to archive todo: %v", err)
		}

		// Verify file is in daily structure based on started date
		expectedArchivePath := GetArchivePath(tempDir, todo2, "")
		if _, err := os.Stat(expectedArchivePath); os.IsNotExist(err) {
			t.Errorf("Todo file should exist in daily archive at %s", expectedArchivePath)
		}
	})

	// Test archiving non-existent todo
	t.Run("Archive non-existent todo", func(t *testing.T) {
		err := manager.ArchiveTodo("non-existent-id")
		if err == nil {
			t.Error("Archiving non-existent todo should return error")
		}
		if !os.IsNotExist(err) && err.Error() != "todo 'non-existent-id' not found" {
			t.Errorf("Expected 'todo not found' error, got: %v", err)
		}
	})
}

// Test 17: todo_archive should update completed timestamp
func TestArchiveUpdatesCompletedTimestamp(t *testing.T) {
	// Setup test environment
	manager, tempDir, cleanup := SetupTestTodoManager(t)
	defer cleanup()

	// Create a test todo
	todo := CreateTestTodo(t, manager, "Test timestamp update", "high", "feature")

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
	err = manager.ArchiveTodo(todo.ID)
	if err != nil {
		t.Fatalf("Failed to archive todo: %v", err)
	}

	// Read archived todo from daily structure
	archivePath := GetArchivePath(tempDir, todo, "")

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
				errors[index] = manager.ArchiveTodo(todo.ID)
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
		archivePath := GetArchivePath(tempDir, todo, "")
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			t.Error("Todo should be archived")
		}

		// Original should not exist
		originalPath, _ := ResolveTodoPath(tempDir, todo.ID)
		if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
			t.Error("Original todo should not exist after successful archive")
		}
	})

	// Test scenario 3: Verify atomicity with write failure
	t.Run("Verify atomicity on write failure", func(t *testing.T) {
		// Skip on Windows - permission handling is different
		if runtime.GOOS == "windows" {
			t.Skip("Skipping write failure test on Windows")
		}
		// Create a test todo
		todo, err := manager.CreateTodo("Test write failure", "low", "refactor")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Create archive directory but make it read-only
		now := time.Now()
		archiveDir := filepath.Join(tempDir, ".claude", "archive", now.Format("2006"), now.Format("01"), now.Format("02"))
		os.MkdirAll(archiveDir, 0755)

		// Make archive directory read-only to prevent writes
		os.Chmod(archiveDir, 0555)
		defer os.Chmod(archiveDir, 0755) // Restore for cleanup

		// Try to archive - should fail on write
		err = manager.ArchiveTodo(todo.ID)
		if err == nil {
			t.Error("Archive should fail when write fails")
		} else {
			t.Logf("Archive failed as expected with error: %v", err)
		}

		// Original should still exist and be unchanged
		// Use ResolveTodoPath to find the actual location
		originalPath, resolveErr := ResolveTodoPath(tempDir, todo.ID)
		if resolveErr != nil {
			t.Errorf("Failed to resolve todo path: %v", resolveErr)
		} else {
			if _, err := os.Stat(originalPath); os.IsNotExist(err) {
				t.Error("Original todo should still exist after failed archive")

				// Check if file was moved to archive
				archivePath := filepath.Join(archiveDir, todo.ID+".md")
				if _, err := os.Stat(archivePath); err == nil {
					t.Error("Todo was archived despite write failure!")
				}
			}
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
	// Use ResolveTodoPath to find actual location
	todo3Path, _ := ResolveTodoPath(tempDir, todo3.ID)
	err = os.Remove(todo3Path)
	if err != nil {
		t.Logf("Warning: Failed to delete todo3 file: %v", err)
	} else {
		t.Logf("Deleted todo3 file at: %s", todo3Path)
	}

	// Verify it's really gone
	if _, err := os.Stat(todo3Path); !os.IsNotExist(err) {
		t.Error("Todo3 file still exists after deletion")
	}

	// Create list of IDs including a non-existent one
	todoIDs := []string{
		todo1.ID,
		todo2.ID,
		todo3.ID,          // Missing file
		"non-existent-id", // Never existed
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
				t.Logf("Todo %s succeeded as expected", result.ID)
			}
		} else {
			if result.Success {
				t.Errorf("Todo %s should have failed, but succeeded", result.ID)
			} else {
				failureCount++
				if result.Error == nil {
					t.Errorf("Failed result for %s should have error", result.ID)
				} else {
					t.Logf("Todo %s failed as expected with error: %v", result.ID, result.Error)
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
	now := time.Now()
	archiveDir := filepath.Join(tempDir, ".claude", "archive", now.Format("2006"), now.Format("01"), now.Format("02"))

	for id, shouldSucceed := range expectedSuccess {
		if shouldSucceed && id != "non-existent-id" {
			archivePath := filepath.Join(archiveDir, id+".md")
			if _, err := os.Stat(archivePath); os.IsNotExist(err) {
				t.Errorf("Successfully archived todo %s not found in archive", id)
			}
		}
	}

	// Verify failed todos remain in original location (if they existed)
	originalPath := GetTodoPath(tempDir, todo3.ID)
	if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
		t.Error("Failed todo3 should not exist in original location (was already deleted)")
	}
}

// Test: Archive creates directory within .claude structure (not parent dir)
func TestArchiveCreatesDirectoryWithinClaudeStructure(t *testing.T) {
	// Setup test environment
	manager, tempDir, cleanup := SetupTestTodoManager(t)
	defer cleanup()

	// Create a test todo
	todo := CreateTestTodo(t, manager, "Test archive directory structure", "high", "feature")

	// Archive the todo
	err := manager.ArchiveTodo(todo.ID)
	if err != nil {
		t.Fatalf("Failed to archive todo: %v", err)
	}

	// Check where archive was created - it should be within .claude structure
	now := time.Now()
	dailyPath := GetDailyPath(now)

	// Current behavior: archive is created one level up from basePath
	currentArchivePath := filepath.Join(filepath.Dir(tempDir), "archive", dailyPath, todo.ID+".md")

	// Desired behavior: archive should be within .claude directory
	desiredArchivePath := filepath.Join(tempDir, ".claude", "archive", dailyPath, todo.ID+".md")

	// Check that archive should be in .claude structure (this will fail initially)
	if _, err := os.Stat(desiredArchivePath); os.IsNotExist(err) {
		t.Errorf("Archive should be within .claude structure at: %s", desiredArchivePath)
	}

	// Check that archive should NOT be in parent directory
	if _, err := os.Stat(currentArchivePath); !os.IsNotExist(err) {
		t.Errorf("Archive should NOT be in parent directory: %s", currentArchivePath)
	}

	// Log paths for clarity
	t.Logf("Base path: %s", tempDir)
	t.Logf("Current archive path: %s", currentArchivePath)
	t.Logf("Desired archive path: %s", desiredArchivePath)
}
