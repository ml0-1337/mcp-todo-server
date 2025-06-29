package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test 4: UpdateTodo auto-timestamps append operations for results sections
func TestUpdateTodo_AutoTimestampsAppendForResults(t *testing.T) {
	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "todo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create TodoManager
	tm := NewTodoManager(tmpDir)
	
	// Create a test todo with section metadata
	todoID := "test-results-todo"
	todoContent := `---
todo_id: test-results-todo
started: "2025-01-01 10:00:00"
status: in_progress
priority: high
type: feature
sections:
  test_results:
    title: "## Test Results Log"
    order: 1
    schema: results
    required: false
---

# Task: Test todo for results

## Test Results Log

[2025-01-01 10:00:00] Initial test entry
`
	
	// Write test todo file
	filePath := filepath.Join(tmpDir, todoID+".md")
	err = ioutil.WriteFile(filePath, []byte(todoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	// Update with new content (without timestamp)
	err = tm.UpdateTodo(todoID, "test_results", "append", "New test result without timestamp", nil)
	if err != nil {
		t.Fatalf("UpdateTodo failed: %v", err)
	}
	
	// Read the updated content
	updatedContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	
	// Check that the new content has a timestamp
	contentStr := string(updatedContent)
	lines := strings.Split(contentStr, "\n")
	
	// Find the test results section
	foundNewEntry := false
	for _, line := range lines {
		if strings.Contains(line, "New test result without timestamp") {
			foundNewEntry = true
			// Verify it has a timestamp
			if !strings.HasPrefix(strings.TrimSpace(line), "[") {
				t.Error("New entry should have been timestamped")
			}
			break
		}
	}
	
	if !foundNewEntry {
		t.Error("Could not find the new entry in the file")
	}
}