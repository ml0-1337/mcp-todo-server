package handlers

import (
	"context"
	"fmt"
	"github.com/user/mcp-todo-server/core"
	"testing"
)

func TestHandleTodoUpdateChecklistValidation(t *testing.T) {
	// Test 16: Validate checklist syntax in checklist schema
	// Input: Various checklist content formats
	// Expected: Only valid checkbox syntax accepted

	ctx := context.Background()

	// Create mock dependencies
	mockManager := NewMockTodoManager()
	mockSearch := NewMockSearchEngine()
	mockStats := NewMockStatsEngine()
	mockTemplates := NewMockTemplateManager()

	handlers := NewTodoHandlersWithDependencies(
		mockManager,
		mockSearch,
		mockStats,
		mockTemplates,
	)

	// Setup test todo with checklist section
	testTodo := &core.Todo{
		ID:   "test-todo",
		Task: "Test Todo",
		Sections: map[string]*core.SectionDefinition{
			"checklist": {
				Title:    "## Checklist",
				Order:    1,
				Schema:   "checklist",
				Required: false,
			},
		},
	}

	// Configure mock to return test todo
	mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
		if id == "test-todo" {
			return testTodo, nil
		}
		return nil, fmt.Errorf("todo not found")
	}

	// Track update calls
	var updateCalled bool
	mockManager.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
		updateCalled = true
		return nil
	}

	// Test: Valid checklist syntax accepted
	t.Run("Valid checklist syntax", func(t *testing.T) {
		updateCalled = false

		request := MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id":        "test-todo",
				"section":   "checklist",
				"operation": "replace",
				"content": `- [ ] Task 1
- [x] Task 2 completed
- [ ] Task 3 pending
- [x] Task 4 done`,
			},
		}

		result, err := handlers.HandleTodoUpdate(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.IsError {
			t.Fatal("Expected success but got error")
		}

		if !updateCalled {
			t.Error("Expected update to be called")
		}
	})

	// Test: Invalid checkbox syntax rejected
	t.Run("Invalid checkbox syntax", func(t *testing.T) {
		updateCalled = false

		testCases := []struct {
			name    string
			content string
		}{
			{
				name: "Missing space after dash",
				content: `-[ ] No space after dash
-[x] Also invalid`,
			},
			{
				name: "Missing brackets",
				content: `- Task without brackets
- [ ] Valid task
- Another invalid task`,
			},
			{
				name: "Invalid checkbox format",
				content: `- [] Missing space in brackets
- [ ] Valid task  
- [X] Capital X invalid`,
			},
			{
				name: "Mixed content",
				content: `- [ ] Valid checklist item
Some random text that isn't a checklist
- [x] Another valid item`,
			},
			{
				name: "Partial checkbox syntax",
				content: `- [ Incomplete checkbox
- [x] Valid item
- [] No space inside`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				updateCalled = false

				request := MockCallToolRequest{
					Arguments: map[string]interface{}{
						"id":        "test-todo",
						"section":   "checklist",
						"operation": "replace",
						"content":   tc.content,
					},
				}

				result, err := handlers.HandleTodoUpdate(ctx, request.ToCallToolRequest())
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}

				if !result.IsError {
					t.Errorf("Expected validation error for: %s", tc.name)
				}

				if updateCalled {
					t.Errorf("Update should not be called for invalid content: %s", tc.name)
				}
			})
		}
	})

	// Test: Empty lines allowed
	t.Run("Empty lines allowed", func(t *testing.T) {
		updateCalled = false

		request := MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id":        "test-todo",
				"section":   "checklist",
				"operation": "replace",
				"content": `- [ ] Task 1

- [x] Task 2 with empty line above

- [ ] Task 3 with multiple empty lines above`,
			},
		}

		result, err := handlers.HandleTodoUpdate(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.IsError {
			t.Fatal("Expected success but got error - empty lines should be allowed")
		}

		if !updateCalled {
			t.Error("Expected update to be called")
		}
	})

	// Test: Whitespace-only content allowed
	t.Run("Whitespace only content", func(t *testing.T) {
		updateCalled = false

		request := MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id":        "test-todo",
				"section":   "checklist",
				"operation": "replace",
				"content":   "   \n\t\n   ",
			},
		}

		result, err := handlers.HandleTodoUpdate(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.IsError {
			t.Fatal("Expected success - whitespace-only content should be allowed")
		}

		if !updateCalled {
			t.Error("Expected update to be called")
		}
	})

	// Test: Non-checklist section not validated with checklist rules
	t.Run("Non-checklist section uses different validation", func(t *testing.T) {
		// Add a research section to the todo
		testTodo.Sections["findings"] = &core.SectionDefinition{
			Title:  "## Findings & Research",
			Order:  2,
			Schema: "research",
		}

		updateCalled = false

		request := MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id":        "test-todo",
				"section":   "findings",
				"operation": "replace",
				"content": `This is free-form research content.
No checkbox syntax required here.
- This dash doesn't need to be a checkbox`,
			},
		}

		result, err := handlers.HandleTodoUpdate(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.IsError {
			t.Fatal("Expected success - research sections don't require checkbox syntax")
		}

		if !updateCalled {
			t.Error("Expected update to be called")
		}
	})
}
