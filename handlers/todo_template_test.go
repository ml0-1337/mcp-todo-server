package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

func TestHandleTodoTemplate(t *testing.T) {
	tests := []struct {
		name           string
		request        *MockCallToolRequest
		setupMocks     func(*MockTodoManager, *MockSearchEngine, *MockTemplateManager)
		context        context.Context
		expectError    bool
		expectedResult func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name: "list templates successfully",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					// No template parameter - should list templates
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tmpl.ListTemplatesFunc = func() ([]string, error) {
					return []string{"bug-fix", "feature", "research", "tdd-cycle"}, nil
				}
			},
			context:     context.Background(),
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success but got error")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "Available templates:") {
					t.Errorf("Expected 'Available templates:' in response, got: %s", content)
				}
				if !strings.Contains(content, "- bug-fix") {
					t.Errorf("Expected '- bug-fix' in response, got: %s", content)
				}
				if !strings.Contains(content, "- feature") {
					t.Errorf("Expected '- feature' in response, got: %s", content)
				}
				if !strings.Contains(content, "- research") {
					t.Errorf("Expected '- research' in response, got: %s", content)
				}
				if !strings.Contains(content, "- tdd-cycle") {
					t.Errorf("Expected '- tdd-cycle' in response, got: %s", content)
				}
			},
		},
		{
			name: "list templates with error",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					// No template parameter
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tmpl.ListTemplatesFunc = func() ([]string, error) {
					return nil, errors.New("failed to read templates directory")
				}
			},
			context:     context.Background(),
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error result")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "failed to read templates directory") {
					t.Errorf("Expected error message in response, got: %s", content)
				}
			},
		},
		{
			name: "list templates empty",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					// No template parameter
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tmpl.ListTemplatesFunc = func() ([]string, error) {
					return []string{}, nil
				}
			},
			context:     context.Background(),
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success but got error")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "Available templates:") {
					t.Errorf("Expected 'Available templates:' in response, got: %s", content)
				}
			},
		},
		{
			name: "create from template successfully",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"template": "bug-fix",
					"task":     "Fix authentication bug",
					"priority": "high",
					"type":     "bug",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tmpl.CreateFromTemplateFunc = func(templateName, task, priority, todoType string) (*core.Todo, error) {
					if templateName != "bug-fix" {
						t.Errorf("Expected template name 'bug-fix', got %s", templateName)
					}
					if task != "Fix authentication bug" {
						t.Errorf("Expected task 'Fix authentication bug', got %s", task)
					}
					if priority != "high" {
						t.Errorf("Expected priority 'high', got %s", priority)
					}
					if todoType != "bug" {
						t.Errorf("Expected type 'bug', got %s", todoType)
					}
					return &core.Todo{
						ID:       "bug-123",
						Task:     task,
						Priority: priority,
						Type:     todoType,
						Status:   "in_progress",
					}, nil
				}
				se.IndexTodoFunc = func(todo *core.Todo, content string) error {
					return nil
				}
			},
			context:     context.Background(),
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success but got error")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "\"id\": \"bug-123\"") {
					t.Errorf("Expected todo ID in response, got: %s", content)
				}
				if !strings.Contains(content, "\"template\": \"Applied template sections and structure\"") {
					t.Errorf("Expected template applied message, got: %s", content)
				}
				if !strings.Contains(content, "Todo created from template successfully: bug-123") {
					t.Errorf("Expected success message, got: %s", content)
				}
			},
		},
		{
			name: "create from template with missing task parameter",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"template": "bug-fix",
					// Missing task parameter
					"priority": "high",
					"type":     "bug",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				// The handler will call CreateFromTemplate with empty task
				tmpl.CreateFromTemplateFunc = func(templateName, task, priority, todoType string) (*core.Todo, error) {
					if task == "" {
						return nil, errors.New("task cannot be empty")
					}
					return &core.Todo{
						ID:       "bug-123",
						Task:     task,
						Priority: priority,
						Type:     todoType,
						Status:   "in_progress",
					}, nil
				}
			},
			context:     context.Background(),
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				// Since task, _ := request.RequireString("task") ignores error,
				// the validation should happen in CreateFromTemplate
				if !result.IsError {
					t.Errorf("Expected error result for empty task")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "task cannot be empty") {
					t.Errorf("Expected error about empty task, got: %s", content)
				}
			},
		},
		{
			name: "create from template with default priority and type",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"template": "feature",
					"task":     "Add new feature",
					// Using defaults for priority and type
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tmpl.CreateFromTemplateFunc = func(templateName, task, priority, todoType string) (*core.Todo, error) {
					if priority != "high" {
						t.Errorf("Expected default priority 'high', got %s", priority)
					}
					if todoType != "feature" {
						t.Errorf("Expected default type 'feature', got %s", todoType)
					}
					return &core.Todo{
						ID:       "feature-123",
						Task:     task,
						Priority: priority,
						Type:     todoType,
						Status:   "in_progress",
					}, nil
				}
				se.IndexTodoFunc = func(todo *core.Todo, content string) error {
					return nil
				}
			},
			context:     context.Background(),
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success but got error")
				}
			},
		},
		{
			name: "create from template error",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"template": "non-existent",
					"task":     "Test task",
					"priority": "high",
					"type":     "feature",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tmpl.CreateFromTemplateFunc = func(templateName, task, priority, todoType string) (*core.Todo, error) {
					return nil, errors.New("template not found: non-existent")
				}
			},
			context:     context.Background(),
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error result")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "template not found") && !strings.Contains(content, "Todo not found") {
					t.Errorf("Expected template not found error, got: %s", content)
				}
			},
		},
		{
			name: "create from template with indexing failure",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"template": "bug-fix",
					"task":     "Fix bug",
					"priority": "high",
					"type":     "bug",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tmpl.CreateFromTemplateFunc = func(templateName, task, priority, todoType string) (*core.Todo, error) {
					return &core.Todo{
						ID:       "bug-456",
						Task:     task,
						Priority: priority,
						Type:     todoType,
						Status:   "in_progress",
					}, nil
				}
				se.IndexTodoFunc = func(todo *core.Todo, content string) error {
					return errors.New("index not available")
				}
			},
			context:     context.Background(),
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				// Should succeed even if indexing fails
				if result.IsError {
					t.Errorf("Expected success despite indexing failure")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "\"id\": \"bug-456\"") {
					t.Errorf("Expected todo ID in response, got: %s", content)
				}
			},
		},
		{
			name: "create from template with custom working directory context",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"template": "feature",
					"task":     "Context-aware feature",
					"priority": "medium",
					"type":     "feature",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tmpl.CreateFromTemplateFunc = func(templateName, task, priority, todoType string) (*core.Todo, error) {
					return &core.Todo{
						ID:       "ctx-123",
						Task:     task,
						Priority: priority,
						Type:     todoType,
						Status:   "in_progress",
					}, nil
				}
				se.IndexTodoFunc = func(todo *core.Todo, content string) error {
					return nil
				}
			},
			context:     context.Background(),
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success but got error")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				// The path should be affected by the context
				if !strings.Contains(content, "\"path\":") {
					t.Errorf("Expected path in response, got: %s", content)
				}
				if !strings.Contains(content, "ctx-123") {
					t.Errorf("Expected todo ID in path, got: %s", content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockManager := NewMockTodoManager()
			mockSearch := NewMockSearchEngine()
			mockTemplates := NewMockTemplateManager()

			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(mockManager, mockSearch, mockTemplates)
			}

			// Create handlers
			handlers := &TodoHandlers{
				manager:   mockManager,
				search:    mockSearch,
				templates: mockTemplates,
			}

			// Call handler
			result, err := handlers.HandleTodoTemplate(tt.context, tt.request.ToCallToolRequest())

			// Check error
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check result
			if tt.expectedResult != nil && result != nil {
				tt.expectedResult(t, result)
			}
		})
	}
}

// TestHandleTodoTemplateGetBasePathForContext has been removed
// The system no longer supports context-aware path resolution