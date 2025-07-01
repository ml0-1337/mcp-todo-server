package handlers

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	"strings"
	"testing"
)

// Test 9: HandleTodoSections returns all sections with metadata
func TestHandleTodoSectionsReturnsAllSectionsWithMetadata(t *testing.T) {
	// Create mock managers
	mockManager := NewMockTodoManager()
	mockManager.ReadTodoWithContentFunc = func(id string) (*core.Todo, string, error) {
		if id == "test-todo" {
			todo := &core.Todo{
				ID:       "test-todo",
				Task:     "Test todo with sections",
				Status:   "in_progress",
				Priority: "high",
				Type:     "feature",
				Sections: map[string]*core.SectionDefinition{
					"findings": {
						Title:    "Findings & Research",
						Order:    1,
						Schema:   core.SchemaResearch,
						Required: true,
					},
					"test_list": {
						Title:    "Test List",
						Order:    2,
						Schema:   core.SchemaChecklist,
						Required: true,
						Metadata: map[string]interface{}{
							"completed": 4,
							"total":     8,
						},
					},
					"custom_security": {
						Title:    "Security Analysis",
						Order:    9,
						Schema:   core.SchemaFreeform,
						Required: false,
						Custom:   true,
					},
				},
			}
			content := `# Task: Test todo with sections

## Findings & Research

Some research content here.

## Test List

- [x] Test 1
- [x] Test 2
- [x] Test 3
- [x] Test 4
- [ ] Test 5
- [ ] Test 6
- [ ] Test 7
- [ ] Test 8

## Security Analysis

Custom security analysis content.`
			return todo, content, nil
		}
		return nil, "", fmt.Errorf("todo not found: %s", id)
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
		validate    func(t *testing.T, result *mcp.CallToolResult)
	}{
		{
			name: "get sections for existing todo",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id": "test-todo",
				},
			},
			expectError: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				// Should return structured data about sections
				if result.IsError {
					t.Errorf("Expected success, got error: %v", result.Content)
					return
				}

				// Parse response content
				content, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}

				text := content.Text
				t.Logf("Response text:\n%s", text)

				// Should include section information
				expectedContent := []string{
					`"findings"`,
					`"title": "Findings \u0026 Research"`,
					`"schema": "research"`,
					`"required": true`,
					`"test_list"`,
					`"title": "Test List"`,
					`"schema": "checklist"`,
					`"custom_security"`,
					`"title": "Security Analysis"`,
					`"schema": "freeform"`,
				}

				for _, expected := range expectedContent {
					if !contains(text, expected) {
						t.Errorf("Response missing expected content: %s", expected)
					}
				}
			},
		},
		{
			name: "todo not found",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id": "non-existent",
				},
			},
			expectError: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error for non-existent todo")
					return
				}

				content, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}

				if !contains(strings.ToLower(content.Text), "todo not found") {
					t.Errorf("Expected 'todo not found' error, got: %s", content.Text)
				}
			},
		},
		{
			name: "missing id parameter",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{},
			},
			expectError: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error for missing id")
					return
				}

				content, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}

				if !contains(content.Text, "'id'") || !contains(content.Text, "required parameter") {
					t.Errorf("Expected missing parameter error with 'id', got: %s", content.Text)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call handler
			result, err := handlers.HandleTodoSections(context.Background(), tt.request.ToCallToolRequest())

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("HandleTodoSections() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Validate result
			if tt.validate != nil && result != nil {
				tt.validate(t, result)
			}
		})
	}
}

// Test 10: HandleTodoSections works with legacy todos (no metadata)
func TestHandleTodoSectionsWorksWithLegacyTodos(t *testing.T) {
	// Create mock managers
	mockManager := NewMockTodoManager()
	mockManager.ReadTodoWithContentFunc = func(id string) (*core.Todo, string, error) {
		if id == "legacy-todo" {
			todo := &core.Todo{
				ID:       "legacy-todo",
				Task:     "Legacy todo without section metadata",
				Status:   "in_progress",
				Priority: "high",
				Type:     "feature",
				// No Sections field - nil
			}
			// Content without any section headers
			content := "# Task: Legacy todo without section metadata\n\nJust some plain content without sections."
			return todo, content, nil
		}
		return nil, "", fmt.Errorf("todo not found: %s", id)
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

	// Test case
	request := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"id": "legacy-todo",
		},
	}

	// Call handler
	result, err := handlers.HandleTodoSections(context.Background(), request.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoSections() error = %v", err)
	}

	// Should not be an error
	if result.IsError {
		t.Errorf("Expected success, got error: %v", result.Content)
		return
	}

	// Parse response content
	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Errorf("Expected TextContent, got %T", result.Content[0])
		return
	}

	text := content.Text

	// Should indicate no sections found (legacy todo with no identifiable sections)
	if !contains(text, "No sections found") {
		t.Errorf("Expected 'No sections found' message, got: %s", text)
	}

	// Should still show todo ID
	if !contains(text, "legacy-todo") {
		t.Errorf("Response missing todo ID")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
