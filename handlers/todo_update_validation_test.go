package handlers

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	"testing"
)

// Test 11: Update section with schema validation enabled
func TestUpdateSectionWithSchemaValidation(t *testing.T) {
	t.Skip("Skipping test - schema validation not implemented in UpdateTodo")
	// Create mock managers
	mockManager := NewMockTodoManager()

	// Setup todo with sections that have schemas
	testTodo := &core.Todo{
		ID:       "test-todo-validation",
		Task:     "Test todo with validation",
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
		Sections: map[string]*core.SectionDefinition{
			"checklist": {
				Title:    "## Checklist",
				Order:    1,
				Schema:   core.SchemaChecklist,
				Required: true,
			},
			"test_cases": {
				Title:    "## Test Cases",
				Order:    2,
				Schema:   core.SchemaTestCases,
				Required: true,
			},
			"results": {
				Title:    "## Test Results Log",
				Order:    3,
				Schema:   core.SchemaResults,
				Required: false,
			},
		},
	}

	mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
		if id == "test-todo-validation" {
			return testTodo, nil
		}
		return nil, nil
	}

	mockManager.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
		// The real UpdateTodo would validate against the hardcoded section map
		// For now, just accept the section
		return nil
	}

	mockSearch := &MockSearchEngine{}
	mockStats := &MockStatsEngine{}
	mockTemplates := &MockTemplateManager{}

	// Create handlers with mocks
	handlers := NewTodoHandlersWithDependencies(
		mockManager,
		mockSearch,
		mockStats,
		mockTemplates,
	)

	// Test cases
	tests := []struct {
		name        string
		request     *MockCallToolRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid checklist content",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-todo-validation",
					"section":   "checklist",
					"operation": "append",
					"content":   "\n- [ ] New task\n- [x] Completed task",
				},
			},
			expectError: false,
		},
		{
			name: "invalid checklist content",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-todo-validation",
					"section":   "checklist",
					"operation": "append",
					"content":   "\nThis is not a checklist item",
				},
			},
			expectError: true,
			errorMsg:    "non-checklist content found",
		},
		{
			name: "invalid checkbox syntax",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-todo-validation",
					"section":   "checklist",
					"operation": "append",
					"content":   "\n- [] Missing space in checkbox",
				},
			},
			expectError: true,
			errorMsg:    "invalid checkbox syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call handler
			result, err := handlers.HandleTodoUpdate(context.Background(), tt.request.ToCallToolRequest())

			// Check error
			if err != nil {
				t.Errorf("HandleTodoUpdate() unexpected error = %v", err)
				return
			}

			// Check result
			if tt.expectError {
				if !result.IsError {
					t.Errorf("Expected error but got success")
					return
				}

				content, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}

				if !contains(content.Text, tt.errorMsg) {
					t.Errorf("Expected error message containing '%s', got: %s", tt.errorMsg, content.Text)
				}
			} else {
				if result.IsError {
					content, _ := result.Content[0].(mcp.TextContent)
					t.Errorf("Expected success but got error: %s", content.Text)
				}
			}
		})
	}
}

// Test 12: Reject update that violates schema
func TestRejectUpdateThatViolatesSchema(t *testing.T) {
	t.Skip("Skipping test - schema validation not implemented in UpdateTodo")
	// Create mock managers
	mockManager := NewMockTodoManager()

	// Setup todo with various section schemas
	testTodo := &core.Todo{
		ID:       "test-todo-strict",
		Task:     "Test todo with strict validation",
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
		Sections: map[string]*core.SectionDefinition{
			"tests": {
				Title:    "## Test Cases",
				Order:    1,
				Schema:   core.SchemaTestCases,
				Required: true,
			},
			"findings": {
				Title:    "## Findings & Research",
				Order:    2,
				Schema:   core.SchemaResearch,
				Required: false,
			},
			"checklist": {
				Title:    "## Checklist",
				Order:    3,
				Schema:   core.SchemaChecklist,
				Required: true,
			},
		},
	}

	mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
		if id == "test-todo-strict" {
			return testTodo, nil
		}
		return nil, fmt.Errorf("todo not found: %s", id)
	}

	mockManager.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
		// Mock accepts the update since validation happens in handler
		return nil
	}

	mockSearch := &MockSearchEngine{}
	mockStats := &MockStatsEngine{}
	mockTemplates := &MockTemplateManager{}

	// Create handlers with mocks
	handlers := NewTodoHandlersWithDependencies(
		mockManager,
		mockSearch,
		mockStats,
		mockTemplates,
	)

	// Test cases for different schema violations
	tests := []struct {
		name     string
		request  *MockCallToolRequest
		errorMsg string
	}{
		{
			name: "tests section without code blocks",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-todo-strict",
					"section":   "tests",
					"operation": "replace",
					"content":   "Just some plain text without any code blocks",
				},
			},
			errorMsg: "no code blocks found",
		},
		{
			name: "findings section validation (research allows any text)",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-todo-strict",
					"section":   "findings",
					"operation": "append",
					"content":   "\nAny text is valid for research sections",
				},
			},
			errorMsg: "", // This should succeed, not fail
		},
		{
			name: "checklist with mixed content",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-todo-strict",
					"section":   "checklist",
					"operation": "append",
					"content":   "\n- [x] Valid checkbox\nSome random text\n- [ ] Another checkbox",
				},
			},
			errorMsg: "non-checklist content found",
		},
		{
			name: "checklist with malformed checkbox",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-todo-strict",
					"section":   "checklist",
					"operation": "append",
					"content":   "\n- [] Missing space in checkbox",
				},
			},
			errorMsg: "invalid checkbox syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call handler
			result, err := handlers.HandleTodoUpdate(context.Background(), tt.request.ToCallToolRequest())

			// Should not return a Go error
			if err != nil {
				t.Errorf("HandleTodoUpdate() unexpected error = %v", err)
				return
			}

			// Check if we expect an error
			if tt.errorMsg != "" {
				// Result should be an error
				if !result.IsError {
					t.Errorf("Expected validation error but got success")
					return
				}

				// Check error message
				content, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}

				if !contains(content.Text, tt.errorMsg) {
					t.Errorf("Expected error message containing '%s', got: %s", tt.errorMsg, content.Text)
				}
			} else {
				// Should succeed
				if result.IsError {
					content, _ := result.Content[0].(mcp.TextContent)
					t.Errorf("Expected success but got error: %s", content.Text)
				}
			}
		})
	}
}
