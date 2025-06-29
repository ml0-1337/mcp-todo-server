package handlers

import (
	"context"
	"testing"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// Test 11: Update section with schema validation enabled
func TestUpdateSectionWithSchemaValidation(t *testing.T) {
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