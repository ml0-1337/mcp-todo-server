package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Test 7: Integration test - todo_read with format=full returns complete content
func TestHandleTodoReadFullFormat(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-read-full-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test todo with content
	todoContent := `---
todo_id: test-integration-full
started: "2025-06-29T10:00:00Z"
completed:
status: in_progress
priority: high
type: feature
parent_id: parent-123
tags: [integration, test]
---

# Task: Integration test for full format

## Findings & Research
This is the findings section with research content.
Multiple lines of important information.

## Test Strategy
- Unit tests coverage
- Integration test suite
- Performance benchmarks

## Test List
- [ ] Test item 1
- [ ] Test item 2

## Checklist
- [ ] Setup environment
- [x] Write tests
- [>] Run tests

## Working Scratchpad
Temporary notes and ideas.
`

	// Create .claude/todos directory
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	err = os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create todos directory: %v", err)
	}

	// Write test todo file
	todoPath := filepath.Join(todosDir, "test-integration-full.md")
	err = ioutil.WriteFile(todoPath, []byte(todoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test todo: %v", err)
	}

	// Create handlers with test dependencies
	handlers, err := NewTodoHandlers(tempDir, "")
	if err != nil {
		t.Fatalf("Failed to create handlers: %v", err)
	}
	defer handlers.Close()

	// Create request for single todo with full format
	request := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"id":     "test-integration-full",
			"format": "full",
		},
	}

	// Call HandleTodoRead with proper context
	ctx := context.Background()
	result, err := handlers.HandleTodoRead(ctx, request.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoRead failed: %v", err)
	}

	// Parse result
	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}

	textContent := result.Content[0].(mcp.TextContent)
	t.Logf("Result content: %s", textContent.Text) // Debug output
	var output map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &output)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify metadata
	if output["id"] != "test-integration-full" {
		t.Errorf("Expected id 'test-integration-full', got %v", output["id"])
	}
	if output["task"] != "Integration test for full format" {
		t.Errorf("Expected task 'Integration test for full format', got %v", output["task"])
	}
	if output["parent_id"] != "parent-123" {
		t.Errorf("Expected parent_id 'parent-123', got %v", output["parent_id"])
	}

	// Verify sections exist
	sections, ok := output["sections"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'sections' field in output")
	}

	// Verify findings section content
	findings, ok := sections["findings"].(string)
	if !ok {
		t.Fatal("Expected 'findings' section to be a string")
	}
	if !strings.Contains(findings, "research content") {
		t.Errorf("Findings section missing expected content")
	}
	if !strings.Contains(findings, "Multiple lines") {
		t.Errorf("Findings section missing multi-line content")
	}

	// Verify test_strategy section
	testStrategy, ok := sections["test_strategy"].(string)
	if !ok {
		t.Fatal("Expected 'test_strategy' section to be a string")
	}
	if !strings.Contains(testStrategy, "Integration test suite") {
		t.Errorf("Test strategy section missing expected content")
	}

	// Verify checklist is parsed as array
	checklist, ok := sections["checklist"].([]interface{})
	if !ok {
		t.Fatal("Expected 'checklist' section to be an array")
	}
	if len(checklist) != 3 {
		t.Errorf("Expected 3 checklist items, got %d", len(checklist))
	}

	// Verify checklist item states
	items := []struct {
		text   string
		status string
	}{
		{"Setup environment", "pending"},
		{"Write tests", "completed"},
		{"Run tests", "in_progress"},
	}

	for i, expected := range items {
		if item, ok := checklist[i].(map[string]interface{}); ok {
			if item["text"] != expected.text {
				t.Errorf("Checklist item %d: expected text '%s', got %v", i, expected.text, item["text"])
			}
			if item["status"] != expected.status {
				t.Errorf("Checklist item %d: expected status '%s', got %v", i, expected.status, item["status"])
			}
		}
	}

	// Verify scratchpad section
	scratchpad, ok := sections["scratchpad"].(string)
	if !ok {
		t.Fatal("Expected 'scratchpad' section to be a string")
	}
	if !strings.Contains(scratchpad, "Temporary notes") {
		t.Errorf("Scratchpad section missing expected content")
	}
}

// Test multiple todos with full format
func TestHandleTodoReadMultipleFullFormat(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-read-multi-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create multiple test todos
	todos := []struct {
		id      string
		content string
	}{
		{
			id: "todo-1",
			content: `---
todo_id: todo-1
started: "2025-06-29T10:00:00Z"
status: in_progress
priority: high
type: feature
---

# Task: First todo

## Checklist
- [x] Completed item
`},
		{
			id: "todo-2",
			content: `---
todo_id: todo-2
started: "2025-06-29T11:00:00Z"
status: in_progress
priority: medium
type: bug
---

# Task: Second todo

## Findings & Research
Bug investigation results.
`},
	}

	// Create .claude/todos directory
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	err = os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create todos directory: %v", err)
	}

	// Write test todos
	for _, todo := range todos {
		todoPath := filepath.Join(todosDir, todo.id+".md")
		err = ioutil.WriteFile(todoPath, []byte(todo.content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test todo %s: %v", todo.id, err)
		}
	}

	// Create handlers
	handlers, err := NewTodoHandlers(tempDir, "")
	if err != nil {
		t.Fatalf("Failed to create handlers: %v", err)
	}
	defer handlers.Close()

	// Create request for all todos with full format
	request := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"format": "full",
		},
	}

	// Call HandleTodoRead with proper context
	ctx := context.Background()
	result, err := handlers.HandleTodoRead(ctx, request.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoRead failed: %v", err)
	}

	// Parse result
	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}

	textContent := result.Content[0].(mcp.TextContent)
	var output []map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &output)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, textContent.Text)
	}

	// Debug output
	t.Logf("Full format output: %s", textContent.Text)

	// Should have 2 todos
	if len(output) != 2 {
		t.Fatalf("Expected 2 todos, got %d", len(output))
	}

	// Verify both todos have sections
	for i, todo := range output {
		if _, ok := todo["sections"]; !ok {
			t.Errorf("Todo %d missing 'sections' field", i)
		}
	}
}
