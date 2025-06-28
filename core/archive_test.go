package core

import (
	"testing"
	"time"
	"os"
	"path/filepath"
	"io/ioutil"
	"fmt"
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