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

	// Create directory structure
	todosDir := filepath.Join(tmpDir, ".claude", "todos")
	err = os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Write test todo file
	filePath := filepath.Join(todosDir, todoID+".md")
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

// Test 5: UpdateTodo auto-timestamps prepend operations for results sections
func TestUpdateTodo_AutoTimestampsPrependResultsSection(t *testing.T) {
	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "todo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create TodoManager
	tm := NewTodoManager(tmpDir)

	// Create a test todo with section metadata
	todoID := "test-prepend-results"
	todoContent := `---
todo_id: test-prepend-results
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

# Task: Test prepend for results

## Test Results Log

[2025-01-01 10:00:00] Existing test entry
`

	// Create directory structure
	todosDir := filepath.Join(tmpDir, ".claude", "todos")
	err = os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Write test todo file
	filePath := filepath.Join(todosDir, todoID+".md")
	err = ioutil.WriteFile(filePath, []byte(todoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Update with prepend operation (without timestamp)
	err = tm.UpdateTodo(todoID, "test_results", "prepend", "Prepended test result", nil)
	if err != nil {
		t.Fatalf("UpdateTodo failed: %v", err)
	}

	// Read the updated content
	updatedContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the prepended content has a timestamp and comes first
	contentStr := string(updatedContent)
	lines := strings.Split(contentStr, "\n")

	// Find the test results section
	inSection := false
	firstEntry := ""
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Test Results Log" {
			inSection = true
			// Skip empty lines after section header
			for j := i + 1; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) != "" {
					firstEntry = lines[j]
					break
				}
			}
			break
		}
	}

	if !inSection {
		t.Fatal("Could not find Test Results Log section")
	}

	// Verify the first entry is our prepended content with timestamp
	if !strings.Contains(firstEntry, "Prepended test result") {
		t.Errorf("First entry should be prepended content, got: %s", firstEntry)
	}

	if !strings.HasPrefix(strings.TrimSpace(firstEntry), "[") {
		t.Error("Prepended entry should have been timestamped")
	}
}

// Test 6: UpdateTodo doesn't add timestamps for replace operations
func TestUpdateTodo_NoTimestampsForReplaceOperation(t *testing.T) {
	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "todo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create TodoManager
	tm := NewTodoManager(tmpDir)

	// Create a test todo with section metadata
	todoID := "test-replace-results"
	todoContent := `---
todo_id: test-replace-results
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

# Task: Test replace for results

## Test Results Log

[2025-01-01 10:00:00] Initial test entry
`

	// Create directory structure
	todosDir := filepath.Join(tmpDir, ".claude", "todos")
	err = os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Write test todo file
	filePath := filepath.Join(todosDir, todoID+".md")
	err = ioutil.WriteFile(filePath, []byte(todoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Update with replace operation (with and without timestamps)
	replaceContent := `Manual content without timestamp
[2025-01-01 11:00:00] Manually timestamped entry`

	err = tm.UpdateTodo(todoID, "test_results", "replace", replaceContent, nil)
	if err != nil {
		t.Fatalf("UpdateTodo failed: %v", err)
	}

	// Read the updated content
	updatedContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	// Check that replace operation didn't add timestamps
	contentStr := string(updatedContent)

	// Verify the manual content is preserved as-is
	if !strings.Contains(contentStr, "Manual content without timestamp") {
		t.Error("Replace operation should preserve content without timestamps")
	}

	// Verify manually timestamped entries are preserved
	if !strings.Contains(contentStr, "[2025-01-01 11:00:00] Manually timestamped entry") {
		t.Error("Replace operation should preserve manually timestamped entries")
	}
}

// Test 7: UpdateTodo doesn't add timestamps for non-results sections
func TestUpdateTodo_NoTimestampsForNonResultsSections(t *testing.T) {
	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "todo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create TodoManager
	tm := NewTodoManager(tmpDir)

	// Create a test todo with different section types
	todoID := "test-non-results"
	todoContent := `---
todo_id: test-non-results
started: "2025-01-01 10:00:00"
status: in_progress
priority: high
type: feature
sections:
  findings:
    title: "## Findings & Research"
    order: 1
    schema: research
    required: false
  checklist:
    title: "## Checklist"
    order: 2
    schema: checklist
    required: false
---

# Task: Test non-results sections

## Findings & Research

Initial research content

## Checklist

- [ ] Initial checklist item
`

	// Create directory structure
	todosDir := filepath.Join(tmpDir, ".claude", "todos")
	err = os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Write test todo file
	filePath := filepath.Join(todosDir, todoID+".md")
	err = ioutil.WriteFile(filePath, []byte(todoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Update findings section (should NOT add timestamp)
	err = tm.UpdateTodo(todoID, "findings", "append", "New research findings without timestamp", nil)
	if err != nil {
		t.Fatalf("UpdateTodo failed: %v", err)
	}

	// Update checklist section (should NOT add timestamp)
	err = tm.UpdateTodo(todoID, "checklist", "append", "- [ ] New checklist item", nil)
	if err != nil {
		t.Fatalf("UpdateTodo failed: %v", err)
	}

	// Read the updated content
	updatedContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(updatedContent)

	// Verify findings section has no timestamps
	if !strings.Contains(contentStr, "New research findings without timestamp") {
		t.Error("Findings content should be added without modification")
	}

	// Check that no timestamps were added to findings
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "New research findings") && strings.HasPrefix(strings.TrimSpace(line), "[") {
			t.Error("Non-results sections should not have automatic timestamps")
		}
	}

	// Verify checklist content is preserved as-is
	if !strings.Contains(contentStr, "- [ ] New checklist item") {
		t.Error("Checklist content should be added without modification")
	}
}
