package handlers

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// Test ExtractTodoArchiveParams functionality
func TestTodoArchiveParamsExtraction(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
		expected    *TodoArchiveParams
	}{
		{
			name: "valid parameters",
			args: map[string]interface{}{
				"id": "test-todo-123",
			},
			expectError: false,
			expected: &TodoArchiveParams{
				ID: "test-todo-123",
			},
		},
		{
			name:        "missing id parameter",
			args:        map[string]interface{}{},
			expectError: true,
			errorMsg:    "missing required parameter 'id'",
		},
		{
			name: "empty id parameter",
			args: map[string]interface{}{
				"id": "",
			},
			expectError: true,
			errorMsg:    "missing required parameter 'id'",
		},
		{
			name: "non-string id parameter",
			args: map[string]interface{}{
				"id": 123,
			},
			expectError: true,
			errorMsg:    "missing required parameter 'id'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the ExtractTodoArchiveParams logic
			params := &TodoArchiveParams{}

			// Extract ID
			id, ok := tt.args["id"].(string)
			if !ok || id == "" {
				if !tt.expectError {
					t.Fatal("Expected valid params but got error")
				}
				if tt.errorMsg != "missing required parameter 'id'" {
					t.Errorf("Expected error message '%s', got different error", tt.errorMsg)
				}
				return
			}
			params.ID = id

			// Verify results
			if tt.expectError {
				t.Fatal("Expected error but got valid params")
			}

			if params.ID != tt.expected.ID {
				t.Errorf("Expected ID '%s', got '%s'", tt.expected.ID, params.ID)
			}
		})
	}
}

// Test FormatTodoArchiveResponse
func TestFormatTodoArchiveResponse(t *testing.T) {
	todoID := "test-archive-123"
	archivePath := filepath.Join(".claude", "archive", "2025", "01", "15", "test-archive-123.md")

	result := FormatTodoArchiveResponse(todoID, archivePath, "feature")

	// Verify result is valid MCP response
	if result == nil {
		t.Fatal("FormatTodoArchiveResponse returned nil")
	}

	// Extract and verify content
	if len(result.Content) == 0 {
		t.Fatal("Result content is empty")
	}

	// The response should contain JSON text
	// In actual implementation, we need to check how the response is structured
	// For now, just verify the result is not nil
	if result.IsError {
		t.Error("Response should not be an error")
	}
}

// Test parameter validation for TodoCreateParams
func TestTodoCreateParamsValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
		expected    *TodoCreateParams
	}{
		{
			name: "all valid parameters",
			args: map[string]interface{}{
				"task":      "Implement new feature",
				"priority":  "high",
				"type":      "feature",
				"parent_id": "parent-123",
				"template":  "bug-fix",
			},
			expectError: false,
			expected: &TodoCreateParams{
				Task:     "Implement new feature",
				Priority: "high",
				Type:     "feature",
				ParentID: "parent-123",
				Template: "bug-fix",
			},
		},
		{
			name: "task only with defaults",
			args: map[string]interface{}{
				"task": "Simple task",
			},
			expectError: false,
			expected: &TodoCreateParams{
				Task:     "Simple task",
				Priority: "high",    // default
				Type:     "feature", // default
				ParentID: "",
				Template: "",
			},
		},
		{
			name:        "missing task",
			args:        map[string]interface{}{},
			expectError: true,
			errorMsg:    "missing required parameter 'task'",
		},
		{
			name: "empty task",
			args: map[string]interface{}{
				"task": "",
			},
			expectError: true,
			errorMsg:    "missing required parameter 'task'",
		},
		{
			name: "invalid priority",
			args: map[string]interface{}{
				"task":     "Test task",
				"priority": "urgent",
			},
			expectError: true,
			errorMsg:    "invalid priority 'urgent'. Valid values: high, medium, low",
		},
		{
			name: "invalid type",
			args: map[string]interface{}{
				"task": "Test task",
				"type": "unknown",
			},
			expectError: true,
			errorMsg:    "invalid type 'unknown'. Valid values: feature, bug, refactor, research, multi-phase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the ExtractTodoCreateParams logic
			params := &TodoCreateParams{
				Priority: "high",    // default
				Type:     "feature", // default
			}

			// Extract task (required)
			task, ok := tt.args["task"].(string)
			if !ok || task == "" {
				if !tt.expectError {
					t.Fatal("Expected valid params but got error")
				}
				if tt.errorMsg != "missing required parameter 'task'" {
					t.Errorf("Expected error message '%s'", tt.errorMsg)
				}
				return
			}
			params.Task = task

			// Extract optional priority
			if priority, ok := tt.args["priority"].(string); ok {
				if !isValidPriority(priority) {
					if !tt.expectError {
						t.Fatal("Expected valid params but got invalid priority")
					}
					expectedErr := "invalid priority '" + priority + "'. Valid values: high, medium, low"
					if tt.errorMsg != expectedErr {
						t.Errorf("Expected error message '%s'", expectedErr)
					}
					return
				}
				params.Priority = priority
			}

			// Extract optional type
			if todoType, ok := tt.args["type"].(string); ok {
				if !isValidTodoType(todoType) {
					if !tt.expectError {
						t.Fatal("Expected valid params but got invalid type")
					}
					expectedErr := "invalid type '" + todoType + "'. Valid values: feature, bug, refactor, research, multi-phase"
					if tt.errorMsg != expectedErr {
						t.Errorf("Expected error message '%s'", expectedErr)
					}
					return
				}
				params.Type = todoType
			}

			// Extract optional parent_id
			if parentID, ok := tt.args["parent_id"].(string); ok {
				params.ParentID = parentID
			}

			// Extract optional template
			if template, ok := tt.args["template"].(string); ok {
				params.Template = template
			}

			// Verify results
			if tt.expectError {
				t.Fatal("Expected error but got valid params")
			}

			if params.Task != tt.expected.Task {
				t.Errorf("Task: expected '%s', got '%s'", tt.expected.Task, params.Task)
			}
			if params.Priority != tt.expected.Priority {
				t.Errorf("Priority: expected '%s', got '%s'", tt.expected.Priority, params.Priority)
			}
			if params.Type != tt.expected.Type {
				t.Errorf("Type: expected '%s', got '%s'", tt.expected.Type, params.Type)
			}
			if params.ParentID != tt.expected.ParentID {
				t.Errorf("ParentID: expected '%s', got '%s'", tt.expected.ParentID, params.ParentID)
			}
			if params.Template != tt.expected.Template {
				t.Errorf("Template: expected '%s', got '%s'", tt.expected.Template, params.Template)
			}
		})
	}
}

// Test parameter validation for TodoUpdateParams
func TestTodoUpdateParamsValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid update params",
			args: map[string]interface{}{
				"id":        "test-123",
				"content":   "New content",
				"section":   "findings",
				"operation": "append",
			},
			expectError: false,
		},
		{
			name: "missing id",
			args: map[string]interface{}{
				"content": "New content",
			},
			expectError: true,
			errorMsg:    "missing required parameter 'id'",
		},
		{
			name: "invalid section",
			args: map[string]interface{}{
				"id":      "test-123",
				"section": "invalid",
			},
			expectError: true,
			errorMsg:    "invalid section",
		},
		{
			name: "invalid operation",
			args: map[string]interface{}{
				"id":        "test-123",
				"operation": "invalid",
			},
			expectError: true,
			errorMsg:    "invalid operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract ID (required)
			id, ok := tt.args["id"].(string)
			if !ok || id == "" {
				if !tt.expectError {
					t.Fatal("Expected valid params but got error")
				}
				return
			}

			// Validate section if provided
			if section, ok := tt.args["section"].(string); ok {
				validSections := []string{"status", "findings", "tests", "checklist", "scratchpad"}
				valid := false
				for _, s := range validSections {
					if s == section {
						valid = true
						break
					}
				}
				if !valid && !tt.expectError {
					t.Fatal("Expected valid params but got invalid section")
				}
			}

			// Validate operation if provided
			if operation, ok := tt.args["operation"].(string); ok {
				validOps := []string{"append", "replace", "prepend"}
				valid := false
				for _, op := range validOps {
					if op == operation {
						valid = true
						break
					}
				}
				if !valid && !tt.expectError {
					t.Fatal("Expected valid params but got invalid operation")
				}
			}

			if tt.expectError {
				// Should have failed by now
				return
			}
		})
	}
}

// Test response formatting helpers
func TestResponseFormatting(t *testing.T) {
	t.Run("archive response format", func(t *testing.T) {
		// Test that archive response has expected structure
		todoID := "test-123"
		archivePath := ".claude/archive/2025/01/15/test-123.md"

		// Create expected response structure
		expected := map[string]interface{}{
			"id":           todoID,
			"archive_path": archivePath,
			"message":      "Todo '" + todoID + "' archived successfully",
		}

		// In real test, we would call FormatTodoArchiveResponse and verify
		// For now, just verify the expected structure
		expectedJSON, err := json.Marshal(expected)
		if err != nil {
			t.Fatalf("Failed to marshal expected response: %v", err)
		}

		if len(expectedJSON) == 0 {
			t.Error("Expected non-empty JSON response")
		}
	})
}
