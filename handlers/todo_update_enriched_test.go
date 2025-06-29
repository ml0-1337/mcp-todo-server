package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

func TestHandleTodoUpdateEnrichedResponse(t *testing.T) {
	// Create mock manager with test todo
	testTodo := &core.Todo{
		ID:       "test-todo",
		Task:     "Test todo with checklist",
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
		Sections: map[string]*core.SectionDefinition{
			"findings": {
				Title:  "## Findings & Research",
				Schema: core.SchemaResearch,
			},
			"checklist": {
				Title:  "## Checklist",
				Schema: core.SchemaChecklist,
			},
			"scratchpad": {
				Title:  "## Working Scratchpad",
				Schema: core.SchemaFreeform,
			},
		},
	}

	testContent := `---
todo_id: test-todo
status: in_progress
priority: high
type: feature
---

# Task: Test todo with checklist

## Findings & Research

Some research content here.

## Checklist

- [x] First completed item
- [ ] Second pending item
- [>] Third in progress item

## Working Scratchpad

Some notes here.`

	mockManager := NewMockTodoManager()
	mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
		if id == "test-todo" {
			return testTodo, nil
		}
		return nil, fmt.Errorf("todo not found")
	}
	mockManager.ReadTodoContentFunc = func(id string) (string, error) {
		if id == "test-todo" {
			return testContent, nil
		}
		return "", fmt.Errorf("content not found")
	}
	mockManager.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
		// Update the test content based on operation
		if section == "checklist" && operation == "append" {
			testContent = testContent + "\n" + content
		} else if section == "checklist" && operation == "replace" {
			// Simple replacement logic for test
			parts := strings.Split(testContent, "## Checklist")
			if len(parts) > 1 {
				nextSection := strings.Index(parts[1], "\n##")
				if nextSection > 0 {
					testContent = parts[0] + "## Checklist\n\n" + content + parts[1][nextSection:]
				} else {
					testContent = parts[0] + "## Checklist\n\n" + content
				}
			}
		}
		return nil
	}

	mockSearch := NewMockSearchEngine()
	handlers := &TodoHandlers{
		manager: mockManager,
		search:  mockSearch,
	}

	tests := []struct {
		name           string
		request        *MockCallToolRequest
		validateResult func(t *testing.T, result *mcp.CallToolResult)
		expectError    bool
	}{
		{
			name: "update checklist returns enriched response",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-todo",
					"section":   "checklist",
					"operation": "append",
					"content":   "- [ ] Fourth new pending item",
				},
			},
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				// Check that result is text content
				if len(result.Content) != 1 {
					t.Fatalf("Expected 1 content item, got %d", len(result.Content))
				}

				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Fatalf("Expected TextContent, got %T", result.Content[0])
				}

				// Parse JSON response
				var response TodoUpdateResponse
				err := json.Unmarshal([]byte(textContent.Text), &response)
				if err != nil {
					t.Logf("Response text: %s", textContent.Text)
					t.Fatalf("Failed to parse response JSON: %v", err)
				}

				// Validate response structure
				if response.Message == "" {
					t.Error("Expected message in response")
				}

				if response.Todo == nil {
					t.Fatal("Expected todo in response")
				}

				// Check todo fields
				if response.Todo.ID != "test-todo" {
					t.Errorf("Expected todo ID 'test-todo', got %s", response.Todo.ID)
				}

				// Check checklist parsing (original 3 items since mock doesn't update content)
				if len(response.Todo.Checklist) != 3 {
					t.Errorf("Expected 3 checklist items, got %d", len(response.Todo.Checklist))
				} else {
					// Verify checklist states
					expectedStates := []string{"completed", "pending", "in_progress"}
					for i, item := range response.Todo.Checklist {
						if item.Status != expectedStates[i] {
							t.Errorf("Checklist item %d: expected status %s, got %s", i, expectedStates[i], item.Status)
						}
					}
				}

				// Check progress
				if response.Progress == nil {
					t.Error("Expected progress in response")
				} else {
					// Check checklist breakdown - it's parsed from JSON so it's map[string]interface{}
					if breakdown, ok := response.Progress["checklist_breakdown"].(map[string]interface{}); ok {
						if completed, ok := breakdown["completed"].(float64); ok {
							if int(completed) != 1 {
								t.Errorf("Expected 1 completed, got %v", completed)
							}
						}
						if inProgress, ok := breakdown["in_progress"].(float64); ok {
							if int(inProgress) != 1 {
								t.Errorf("Expected 1 in_progress, got %v", inProgress)
							}
						}
						if pending, ok := breakdown["pending"].(float64); ok {
							if int(pending) != 1 {
								t.Errorf("Expected 1 pending, got %v", pending)
							}
						}
					} else {
						t.Errorf("Expected checklist_breakdown in progress, got type %T", response.Progress["checklist_breakdown"])
					}
				}
			},
		},
		{
			name: "update with three-state checklist items",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-todo",
					"section":   "checklist",
					"operation": "replace",
					"content": `- [x] Completed task
- [ ] Pending task
- [>] Arrow in-progress
- [-] Dash in-progress
- [~] Tilde in-progress`,
				},
			},
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				textContent := result.Content[0].(mcp.TextContent)
				var response TodoUpdateResponse
				json.Unmarshal([]byte(textContent.Text), &response)

				if len(response.Todo.Checklist) != 5 {
					t.Errorf("Expected 5 checklist items, got %d", len(response.Todo.Checklist))
				}

				// Count states
				stateCount := map[string]int{}
				for _, item := range response.Todo.Checklist {
					stateCount[item.Status]++
				}

				if stateCount["completed"] != 1 {
					t.Errorf("Expected 1 completed, got %d", stateCount["completed"])
				}
				if stateCount["pending"] != 1 {
					t.Errorf("Expected 1 pending, got %d", stateCount["pending"])
				}
				if stateCount["in_progress"] != 3 {
					t.Errorf("Expected 3 in_progress, got %d", stateCount["in_progress"])
				}
			},
		},
		{
			name: "update non-checklist section still returns enriched response",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-todo",
					"section":   "findings",
					"operation": "append",
					"content":   "\nAdditional research findings.",
				},
			},
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				textContent := result.Content[0].(mcp.TextContent)
				var response TodoUpdateResponse
				json.Unmarshal([]byte(textContent.Text), &response)

				// Should still parse and return checklist
				if len(response.Todo.Checklist) == 0 {
					t.Error("Expected checklist to be parsed even when updating different section")
				}

				// Check sections
				if response.Todo.Sections == nil {
					t.Error("Expected sections in response")
				} else {
					if !response.Todo.Sections["findings"].HasContent {
						t.Error("Expected findings section to have content")
					}
				}
			},
		},
		{
			name: "handles empty checklist",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-todo",
					"section":   "checklist",
					"operation": "replace",
					"content":   "",
				},
			},
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				textContent := result.Content[0].(mcp.TextContent)
				var response TodoUpdateResponse
				json.Unmarshal([]byte(textContent.Text), &response)

				if len(response.Todo.Checklist) != 0 {
					t.Errorf("Expected 0 checklist items for empty content, got %d", len(response.Todo.Checklist))
				}

				// Progress should handle empty checklist gracefully
				if response.Progress["checklist"] != nil {
					t.Error("Expected no checklist progress for empty checklist")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handlers.HandleTodoUpdate(context.Background(), tt.request.ToCallToolRequest())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			tt.validateResult(t, result)
		})
	}
}

func TestFormatEnrichedTodoUpdateResponse(t *testing.T) {
	todo := &core.Todo{
		ID:       "test-todo",
		Task:     "Test Task",
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
	}

	tests := []struct {
		name           string
		content        string
		section        string
		operation      string
		validateResult func(t *testing.T, result *mcp.CallToolResult)
	}{
		{
			name:      "formats checklist with progress",
			content:   "## Checklist\n- [x] Done\n- [ ] Not done\n- [>] In progress",
			section:   "checklist",
			operation: "update",
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				text := result.Content[0].(mcp.TextContent).Text

				// Check for expected elements
				if !strings.Contains(text, `"completed": 1`) {
					t.Error("Expected completed count in response")
				}
				if !strings.Contains(text, `"in_progress": 1`) {
					t.Error("Expected in_progress count in response")
				}
				if !strings.Contains(text, `"pending": 1`) {
					t.Error("Expected pending count in response")
				}
				if !strings.Contains(text, "1/3 completed (33%)") {
					t.Error("Expected progress percentage in response")
				}
			},
		},
		{
			name:      "handles sections without special parsing",
			content:   "## Findings & Research\nSome findings\n\n## Checklist\n- [x] Item",
			section:   "findings",
			operation: "append",
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				text := result.Content[0].(mcp.TextContent).Text

				// Should still parse checklist from other sections
				if !strings.Contains(text, `"status": "completed"`) {
					t.Error("Expected checklist to be parsed")
				}

				// Should show section has content
				if !strings.Contains(text, `"hasContent": true`) {
					t.Error("Expected findings section to show content")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatEnrichedTodoUpdateResponse(todo, tt.content, tt.section, tt.operation)

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			tt.validateResult(t, result)
		})
	}
}
