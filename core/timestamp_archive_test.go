package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test 5: Archive todo with standard timestamp - Should archive to correct daily path
func TestArchiveTodoWithStandardTimestamp(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "archive-timestamp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager
	manager := NewTodoManager(tempDir)

	// Create a todo with standard format
	todo, err := manager.CreateTodo("Test standard timestamp archive", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Archive the todo
	err = manager.ArchiveTodo(todo.ID, "")
	if err != nil {
		t.Fatalf("Failed to archive todo: %v", err)
	}

	// Verify file was moved to archive
	originalPath := filepath.Join(tempDir, todo.ID+".md")
	if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
		t.Error("Original todo file should have been moved")
	}

	// Verify archive structure
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	expectedArchivePath := filepath.Join(filepath.Dir(tempDir), "archive", year, month, day, todo.ID+".md")

	if _, err := os.Stat(expectedArchivePath); os.IsNotExist(err) {
		t.Errorf("Archive file not found at expected path: %s", expectedArchivePath)
	}
}

// Test 6: Archive todo with RFC3339 timestamp - Should archive to correct daily path
func TestArchiveTodoWithRFC3339Timestamp(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "archive-rfc3339-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test todo content with RFC3339 format
	content := `---
todo_id: test-rfc3339-archive
started: 2025-01-29T02:55:00Z
completed:
status: in_progress
priority: low
type: test
---

# Task: Test RFC3339 timestamp archive

## Description
This todo uses RFC3339 timestamp format to test archiving.
`

	// Write test file
	filePath := filepath.Join(tempDir, "test-rfc3339-archive.md")
	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create manager and archive todo
	manager := NewTodoManager(tempDir)
	err = manager.ArchiveTodo("test-rfc3339-archive", "")
	if err != nil {
		t.Fatalf("Failed to archive todo with RFC3339 timestamp: %v", err)
	}

	// Verify file was moved to archive
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Original todo file should have been moved")
	}

	// Verify archive structure based on started date
	expectedArchivePath := filepath.Join(filepath.Dir(tempDir), "archive", "2025", "01", "29", "test-rfc3339-archive.md")
	if _, err := os.Stat(expectedArchivePath); os.IsNotExist(err) {
		t.Errorf("Archive file not found at expected path: %s", expectedArchivePath)
	}

	// Read archived file to verify completion timestamp was added
	archivedContent, err := ioutil.ReadFile(expectedArchivePath)
	if err != nil {
		t.Fatalf("Failed to read archived file: %v", err)
	}

	// Verify that completed timestamp was added
	if !contains(string(archivedContent), "completed:") || contains(string(archivedContent), "completed:\n") {
		t.Error("Archived file should have a non-empty completed timestamp")
	}
}
