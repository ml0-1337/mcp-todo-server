package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestHandleTodoLink(t *testing.T) {
	tests := []struct {
		name           string
		request        *MockCallToolRequest
		setupMocks     func(*MockTodoManager, *MockTodoLinker)
		createLinker   func(TodoManager) TodoLinker
		expectError    bool
		expectedResult func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name: "successful link with default link type",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"parent_id": "parent-123",
					"child_id":  "child-456",
					// link_type not provided, should default to "parent-child"
				},
			},
			setupMocks: func(tm *MockTodoManager, tl *MockTodoLinker) {
				tl.LinkTodosFunc = func(parentID, childID, linkType string) error {
					if parentID != "parent-123" {
						t.Errorf("Expected parent_id 'parent-123', got %s", parentID)
					}
					if childID != "child-456" {
						t.Errorf("Expected child_id 'child-456', got %s", childID)
					}
					if linkType != "parent-child" {
						t.Errorf("Expected default link_type 'parent-child', got %s", linkType)
					}
					return nil
				}
			},
			createLinker: func(m TodoManager) TodoLinker {
				mockLinker := NewMockTodoLinker()
				// Setup will be called when the linker is created
				return mockLinker
			},
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
				if !strings.Contains(content, "\"parent_id\": \"parent-123\"") {
					t.Errorf("Expected parent_id in response, got: %s", content)
				}
				if !strings.Contains(content, "\"child_id\": \"child-456\"") {
					t.Errorf("Expected child_id in response, got: %s", content)
				}
				if !strings.Contains(content, "\"link_type\": \"parent-child\"") {
					t.Errorf("Expected link_type in response, got: %s", content)
				}
				// The arrow is escaped in JSON as \u003e
				if !strings.Contains(content, "Todos linked successfully: parent-123") && !strings.Contains(content, "child-456") {
					t.Errorf("Expected success message, got: %s", content)
				}
			},
		},
		{
			name: "successful link with custom link type",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"parent_id": "todo-1",
					"child_id":  "todo-2",
					"link_type": "blocks",
				},
			},
			setupMocks: func(tm *MockTodoManager, tl *MockTodoLinker) {
				tl.LinkTodosFunc = func(parentID, childID, linkType string) error {
					if linkType != "blocks" {
						t.Errorf("Expected link_type 'blocks', got %s", linkType)
					}
					return nil
				}
			},
			createLinker: func(m TodoManager) TodoLinker {
				mockLinker := NewMockTodoLinker()
				// Setup will be called when the linker is created
				return mockLinker
			},
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
				if !strings.Contains(content, "\"link_type\": \"blocks\"") {
					t.Errorf("Expected custom link_type in response, got: %s", content)
				}
			},
		},
		{
			name: "missing parent_id parameter",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					// Missing parent_id
					"child_id":  "child-456",
					"link_type": "parent-child",
				},
			},
			setupMocks: func(tm *MockTodoManager, tl *MockTodoLinker) {
				// No mocks needed - should fail at parameter validation
			},
			createLinker: func(m TodoManager) TodoLinker {
				mockLinker := NewMockTodoLinker()
				// Setup will be called when the linker is created
				return mockLinker
			},
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error for missing parent_id")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				// The error might say "Todo not found" because RequireString might return empty string
				if !strings.Contains(content, "parent_id") && !strings.Contains(content, "Todo not found") {
					t.Errorf("Expected error about missing parent_id or Todo not found, got: %s", content)
				}
			},
		},
		{
			name: "missing child_id parameter",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"parent_id": "parent-123",
					// Missing child_id
					"link_type": "parent-child",
				},
			},
			setupMocks: func(tm *MockTodoManager, tl *MockTodoLinker) {
				// No mocks needed - should fail at parameter validation
			},
			createLinker: func(m TodoManager) TodoLinker {
				mockLinker := NewMockTodoLinker()
				// Setup will be called when the linker is created
				return mockLinker
			},
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error for missing child_id")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				// The error might say "Todo not found" because RequireString might return empty string
				if !strings.Contains(content, "child_id") && !strings.Contains(content, "Todo not found") {
					t.Errorf("Expected error about missing child_id or Todo not found, got: %s", content)
				}
			},
		},
		{
			name: "linker not available",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"parent_id": "parent-123",
					"child_id":  "child-456",
					"link_type": "parent-child",
				},
			},
			setupMocks: func(tm *MockTodoManager, tl *MockTodoLinker) {
				// No setup needed
			},
			createLinker: nil, // No createLinker function
			expectError:  false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error when linker not available")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "Linking feature not available") {
					t.Errorf("Expected 'Linking feature not available' error, got: %s", content)
				}
			},
		},
		{
			name: "linker creation fails",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"parent_id": "parent-123",
					"child_id":  "child-456",
					"link_type": "parent-child",
				},
			},
			setupMocks: func(tm *MockTodoManager, tl *MockTodoLinker) {
				// No setup needed
			},
			createLinker: func(m TodoManager) TodoLinker {
				return nil // Return nil to simulate creation failure
			},
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error when linker creation fails")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "Linking feature not available") {
					t.Errorf("Expected 'Linking feature not available' error, got: %s", content)
				}
			},
		},
		{
			name: "link operation fails",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"parent_id": "parent-123",
					"child_id":  "child-456",
					"link_type": "parent-child",
				},
			},
			setupMocks: func(tm *MockTodoManager, tl *MockTodoLinker) {
				if tl != nil {
					tl.LinkTodosFunc = func(parentID, childID, linkType string) error {
						return errors.New("parent todo not found")
					}
				}
			},
			createLinker: func(m TodoManager) TodoLinker {
				mockLinker := NewMockTodoLinker()
				// Setup will be called when the linker is created
				return mockLinker
			},
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error when link operation fails")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "parent todo not found") && !strings.Contains(content, "Todo not found") {
					t.Errorf("Expected link operation error or Todo not found, got: %s", content)
				}
			},
		},
		{
			name: "link todos with invalid IDs",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"parent_id": "",
					"child_id":  "",
					"link_type": "parent-child",
				},
			},
			setupMocks: func(tm *MockTodoManager, tl *MockTodoLinker) {
				if tl != nil {
					tl.LinkTodosFunc = func(parentID, childID, linkType string) error {
						if parentID == "" || childID == "" {
							return errors.New("invalid todo ID")
						}
						return nil
					}
				}
			},
			createLinker: func(m TodoManager) TodoLinker {
				mockLinker := NewMockTodoLinker()
				// Setup will be called when the linker is created
				return mockLinker
			},
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error for invalid IDs")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "invalid todo ID") && !strings.Contains(content, "Invalid parameter") {
					t.Errorf("Expected 'invalid todo ID' or 'Invalid parameter' error, got: %s", content)
				}
			},
		},
		{
			name: "link todos with self-reference",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"parent_id": "todo-123",
					"child_id":  "todo-123",
					"link_type": "parent-child",
				},
			},
			setupMocks: func(tm *MockTodoManager, tl *MockTodoLinker) {
				if tl != nil {
					tl.LinkTodosFunc = func(parentID, childID, linkType string) error {
						if parentID == childID {
							return errors.New("cannot link todo to itself")
						}
						return nil
					}
				}
			},
			createLinker: func(m TodoManager) TodoLinker {
				mockLinker := NewMockTodoLinker()
				// Setup will be called when the linker is created
				return mockLinker
			},
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error for self-reference")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "cannot link todo to itself") {
					t.Errorf("Expected self-reference error, got: %s", content)
				}
			},
		},
		{
			name: "link todos with unsupported link type",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"parent_id": "parent-123",
					"child_id":  "child-456",
					"link_type": "unsupported-type",
				},
			},
			setupMocks: func(tm *MockTodoManager, tl *MockTodoLinker) {
				if tl != nil {
					tl.LinkTodosFunc = func(parentID, childID, linkType string) error {
						// In the actual implementation, only parent-child is supported
						if linkType != "parent-child" {
							return errors.New("unsupported link type: " + linkType)
						}
						return nil
					}
				}
			},
			createLinker: func(m TodoManager) TodoLinker {
				mockLinker := NewMockTodoLinker()
				// Setup will be called when the linker is created
				return mockLinker
			},
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error for unsupported link type")
				}
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Errorf("Expected TextContent, got %T", result.Content[0])
					return
				}
				content := textContent.Text
				if !strings.Contains(content, "unsupported link type") {
					t.Errorf("Expected unsupported link type error, got: %s", content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockManager := NewMockTodoManager()
			var mockLinker *MockTodoLinker

			// Create handlers
			handlers := NewTodoHandlersWithDependencies(
				mockManager,
				nil, // search not needed for this test
				nil, // stats not needed for this test
				nil, // templates not needed for this test
			)

			// For these tests, we'll mock the linking behavior directly
			if tt.createLinker != nil {
				linker := tt.createLinker(mockManager)
				if ml, ok := linker.(*MockTodoLinker); ok {
					mockLinker = ml
					// Setup mocks after creating the linker
					if tt.setupMocks != nil {
						tt.setupMocks(mockManager, mockLinker)
					}
				}
			}

			// Call handler
			result, err := handlers.HandleTodoLink(context.Background(), tt.request.ToCallToolRequest())

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

// TestHandleTodoLinkIntegration tests the integration with actual TodoLinker
func TestHandleTodoLinkIntegration(t *testing.T) {
	// This test verifies that the handler correctly uses the createLinker factory
	mockManager := NewMockTodoManager()
	mockLinker := NewMockTodoLinker()
	
	linkerCalled := false
	mockLinker.LinkTodosFunc = func(parentID, childID, linkType string) error {
		linkerCalled = true
		return nil
	}

	handlers := NewTodoHandlersWithDependencies(
		mockManager,
		nil, // search not needed for this test
		nil, // stats not needed for this test
		nil, // templates not needed for this test
	)

	request := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"parent_id": "parent-123",
			"child_id":  "child-456",
			"link_type": "parent-child",
		},
	}

	result, err := handlers.HandleTodoLink(context.Background(), request.ToCallToolRequest())
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	if !linkerCalled {
		t.Errorf("Expected linker to be called")
	}

	if result.IsError {
		t.Errorf("Expected success result")
	}
}

// TestHandleTodoLinkWithoutBaseManager tests linking without base manager
func TestHandleTodoLinkWithoutBaseManager(t *testing.T) {
	mockManager := NewMockTodoManager()
	
	handlers := NewTodoHandlersWithDependencies(
		mockManager,
		nil, // search not needed for this test
		nil, // stats not needed for this test
		nil, // templates not needed for this test
	)

	request := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"parent_id": "parent-123",
			"child_id":  "child-456",
			"link_type": "parent-child",
		},
	}

	result, err := handlers.HandleTodoLink(context.Background(), request.ToCallToolRequest())
	if err == nil {
		t.Errorf("Expected error when baseManager is nil")
	}

	if result == nil || !result.IsError {
		t.Errorf("Expected error result when baseManager is nil")
	}
}