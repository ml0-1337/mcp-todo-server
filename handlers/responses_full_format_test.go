package handlers

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// Test 2: formatTodosFull includes all section contents for single todo
func TestFormatTodosFullWithContent(t *testing.T) {
	// Create test todo
	todo := &core.Todo{
		ID:       "test-full-format",
		Task:     "Test full format output",
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
		Started:  time.Date(2025, 6, 29, 10, 0, 0, 0, time.UTC),
		ParentID: "parent-123",
		Tags:     []string{"test", "format"},
	}

	// Test content with all sections
	content := `---
todo_id: test-full-format
started: "2025-06-29 10:00:00"
completed: ""
status: in_progress
priority: high
type: feature
parent_id: parent-123
tags: [test, format]
---

# Task: Test full format output

## Findings & Research
This is the findings section with important research.

## Test Strategy
- Unit tests
- Integration tests
- Edge case tests

## Checklist
- [ ] Item 1
- [x] Item 2
- [>] Item 3

## Working Scratchpad
Working notes and temporary content.
`

	// Call formatTodosFullWithContent (to be implemented)
	result := formatTodosFullWithContent([]*core.Todo{todo}, map[string]string{todo.ID: content})

	// Parse result
	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}

	textContent := result.Content[0].(mcp.TextContent)
	var output []map[string]interface{}
	err := json.Unmarshal([]byte(textContent.Text), &output)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify we have one todo
	if len(output) != 1 {
		t.Fatalf("Expected 1 todo, got %d", len(output))
	}

	todoData := output[0]

	// Verify metadata fields
	if todoData["id"] != "test-full-format" {
		t.Errorf("Expected id 'test-full-format', got %v", todoData["id"])
	}
	if todoData["task"] != "Test full format output" {
		t.Errorf("Expected task 'Test full format output', got %v", todoData["task"])
	}
	if todoData["status"] != "in_progress" {
		t.Errorf("Expected status 'in_progress', got %v", todoData["status"])
	}
	if todoData["parent_id"] != "parent-123" {
		t.Errorf("Expected parent_id 'parent-123', got %v", todoData["parent_id"])
	}

	// Verify sections exist
	sections, ok := todoData["sections"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'sections' field in output")
	}

	// Verify findings section
	findings, ok := sections["findings"].(string)
	if !ok {
		t.Fatal("Expected 'findings' section to be a string")
	}
	if !strings.Contains(findings, "important research") {
		t.Errorf("Findings section missing expected content")
	}

	// Verify test_strategy section
	testStrategy, ok := sections["test_strategy"].(string)
	if !ok {
		t.Fatal("Expected 'test_strategy' section to be a string")
	}
	if !strings.Contains(testStrategy, "Integration tests") {
		t.Errorf("Test strategy section missing expected content")
	}

	// Verify checklist is parsed
	checklist, ok := sections["checklist"].([]interface{})
	if !ok {
		t.Fatal("Expected 'checklist' section to be an array")
	}
	if len(checklist) != 3 {
		t.Errorf("Expected 3 checklist items, got %d", len(checklist))
	}

	// Verify first checklist item
	if item, ok := checklist[0].(map[string]interface{}); ok {
		if item["text"] != "Item 1" {
			t.Errorf("Expected checklist item 'Item 1', got %v", item["text"])
		}
		if item["status"] != "pending" {
			t.Errorf("Expected status 'pending', got %v", item["status"])
		}
	}

	// Verify scratchpad section
	scratchpad, ok := sections["scratchpad"].(string)
	if !ok {
		t.Fatal("Expected 'scratchpad' section to be a string")
	}
	if !strings.Contains(scratchpad, "Working notes") {
		t.Errorf("Scratchpad section missing expected content")
	}
}

// Test that single todo uses formatSingleTodoWithContent
func TestFormatSingleTodoFullWithContent(t *testing.T) {
	todo := &core.Todo{
		ID:       "single-todo",
		Task:     "Single todo test",
		Status:   "in_progress",
		Priority: "medium",
		Type:     "bug",
		Started:  time.Date(2025, 6, 29, 10, 0, 0, 0, time.UTC),
	}

	content := `---
todo_id: single-todo
started: "2025-06-29 10:00:00"
status: in_progress
priority: medium
type: bug
---

# Task: Single todo test

## Findings & Research
Single todo findings.

## Checklist
- [ ] Fix the bug
`

	// Call formatSingleTodoWithContent (to be implemented)
	result := formatSingleTodoWithContent(todo, content, "full")

	// Parse result
	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}

	textContent := result.Content[0].(mcp.TextContent)
	var output map[string]interface{}
	err := json.Unmarshal([]byte(textContent.Text), &output)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify it's a single todo object, not an array
	if output["id"] != "single-todo" {
		t.Errorf("Expected id 'single-todo', got %v", output["id"])
	}

	// Verify sections exist
	if _, ok := output["sections"]; !ok {
		t.Error("Expected 'sections' field in single todo output")
	}
}
