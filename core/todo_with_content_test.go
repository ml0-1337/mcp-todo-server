package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test 1: ReadTodoWithContent returns todo metadata and full content
func TestReadTodoWithContent(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-with-content-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create a test todo with content
	todoContent := `---
todo_id: test-todo-full
started: "2025-06-29T10:00:00Z"
status: in_progress
priority: high
type: feature
---

# Task: Test todo for full content

## Findings & Research
This is the findings section with some content.

## Test Strategy
- Unit tests
- Integration tests

## Checklist
- [ ] Item 1
- [x] Item 2

## Working Scratchpad
Some notes here.
`

	// Create directory structure
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	err = os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	
	// Write test todo file
	todoPath := filepath.Join(todosDir, "test-todo-full.md")
	err = ioutil.WriteFile(todoPath, []byte(todoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test todo: %v", err)
	}

	// Test ReadTodoWithContent
	todo, content, err := manager.ReadTodoWithContent("test-todo-full")
	if err != nil {
		t.Fatalf("ReadTodoWithContent failed: %v", err)
	}

	// Verify todo metadata
	if todo.ID != "test-todo-full" {
		t.Errorf("Expected ID 'test-todo-full', got '%s'", todo.ID)
	}
	if todo.Status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got '%s'", todo.Status)
	}
	if todo.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", todo.Priority)
	}
	if todo.Type != "feature" {
		t.Errorf("Expected type 'feature', got '%s'", todo.Type)
	}
	if todo.Task != "Test todo for full content" {
		t.Errorf("Expected task 'Test todo for full content', got '%s'", todo.Task)
	}

	// Verify content is returned
	if content != todoContent {
		t.Errorf("Content mismatch\nExpected:\n%s\nGot:\n%s", todoContent, content)
	}

	// Verify content contains all sections
	requiredSections := []string{
		"## Findings & Research",
		"## Test Strategy",
		"## Checklist",
		"## Working Scratchpad",
	}
	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			t.Errorf("Content missing section: %s", section)
		}
	}
}

// Test error handling when todo doesn't exist
func TestReadTodoWithContentNotFound(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-with-content-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Try to read non-existent todo
	_, _, err = manager.ReadTodoWithContent("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent todo, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}
