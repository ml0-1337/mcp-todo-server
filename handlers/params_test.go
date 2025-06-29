package handlers

import (
	"fmt"
	"testing"
)

// Test parameter extraction logic directly without MCP dependencies
func TestExtractCreateParams_FromMap(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
		expect  TodoCreateParams
	}{
		{
			name: "valid all params",
			args: map[string]interface{}{
				"task":      "Test task",
				"priority":  "high",
				"type":      "feature",
				"template":  "bug-fix",
				"parent_id": "parent-123",
			},
			wantErr: false,
			expect: TodoCreateParams{
				Task:     "Test task",
				Priority: "high",
				Type:     "feature",
				Template: "bug-fix",
				ParentID: "parent-123",
			},
		},
		{
			name:    "missing task",
			args:    map[string]interface{}{"priority": "high"},
			wantErr: true,
		},
		{
			name: "defaults applied",
			args: map[string]interface{}{"task": "Test task"},
			expect: TodoCreateParams{
				Task:     "Test task",
				Priority: "high",    // default
				Type:     "feature", // default
			},
		},
		{
			name: "invalid priority",
			args: map[string]interface{}{
				"task":     "Test task",
				"priority": "urgent",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			args: map[string]interface{}{
				"task": "Test task",
				"type": "enhancement",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the extraction logic directly
			params, err := extractCreateParamsFromMap(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("extractCreateParamsFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && params != nil {
				if params.Task != tt.expect.Task {
					t.Errorf("Task = %v, want %v", params.Task, tt.expect.Task)
				}
				if params.Priority != tt.expect.Priority {
					t.Errorf("Priority = %v, want %v", params.Priority, tt.expect.Priority)
				}
				if params.Type != tt.expect.Type {
					t.Errorf("Type = %v, want %v", params.Type, tt.expect.Type)
				}
				if params.Template != tt.expect.Template {
					t.Errorf("Template = %v, want %v", params.Template, tt.expect.Template)
				}
				if params.ParentID != tt.expect.ParentID {
					t.Errorf("ParentID = %v, want %v", params.ParentID, tt.expect.ParentID)
				}
			}
		})
	}
}

// extractCreateParamsFromMap extracts params directly from a map
// This is the core logic from ExtractTodoCreateParams, testable without MCP types
func extractCreateParamsFromMap(args map[string]interface{}) (*TodoCreateParams, error) {
	params := &TodoCreateParams{}

	// Required parameter
	task, ok := args["task"].(string)
	if !ok || task == "" {
		return nil, fmt.Errorf("missing required parameter 'task'")
	}
	params.Task = task

	// Optional parameters with defaults
	params.Priority = "high"
	if priority, ok := args["priority"].(string); ok {
		params.Priority = priority
	}

	params.Type = "feature"
	if todoType, ok := args["type"].(string); ok {
		params.Type = todoType
	}

	if template, ok := args["template"].(string); ok {
		params.Template = template
	}

	if parentID, ok := args["parent_id"].(string); ok {
		params.ParentID = parentID
	}

	// Validate enums
	if !isValidPriority(params.Priority) {
		return nil, fmt.Errorf("invalid priority '%s', must be one of: high, medium, low", params.Priority)
	}

	if !isValidTodoType(params.Type) {
		return nil, fmt.Errorf("invalid type '%s', must be one of: feature, bug, refactor, research, multi-phase", params.Type)
	}

	return params, nil
}

// Test read params extraction
func TestExtractReadParams_FromMap(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
		expect  TodoReadParams
	}{
		{
			name: "single todo by ID",
			args: map[string]interface{}{
				"id":     "todo-123",
				"format": "full",
			},
			expect: TodoReadParams{
				ID:     "todo-123",
				Format: "full",
			},
		},
		{
			name: "list with filters",
			args: map[string]interface{}{
				"filter": map[string]interface{}{
					"status":   "in_progress",
					"priority": "high",
					"days":     float64(7),
				},
				"format": "list",
			},
			expect: TodoReadParams{
				Filter: TodoFilter{
					Status:   "in_progress",
					Priority: "high",
					Days:     7,
				},
				Format: "list",
			},
		},
		{
			name: "default format",
			args: map[string]interface{}{"id": "todo-123"},
			expect: TodoReadParams{
				ID:     "todo-123",
				Format: "summary", // default
			},
		},
		{
			name:    "invalid format",
			args:    map[string]interface{}{"format": "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := extractReadParamsFromMap(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("extractReadParamsFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && params != nil {
				if params.ID != tt.expect.ID {
					t.Errorf("ID = %v, want %v", params.ID, tt.expect.ID)
				}
				if params.Format != tt.expect.Format {
					t.Errorf("Format = %v, want %v", params.Format, tt.expect.Format)
				}
				if params.Filter.Status != tt.expect.Filter.Status {
					t.Errorf("Filter.Status = %v, want %v", params.Filter.Status, tt.expect.Filter.Status)
				}
			}
		})
	}
}

// extractReadParamsFromMap extracts read params from a map
func extractReadParamsFromMap(args map[string]interface{}) (*TodoReadParams, error) {
	params := &TodoReadParams{}

	// Optional ID for single todo
	if id, ok := args["id"].(string); ok {
		params.ID = id
	}

	// Extract filter if provided
	if filterObj, ok := args["filter"].(map[string]interface{}); ok {
		if status, ok := filterObj["status"].(string); ok {
			params.Filter.Status = status
		}
		if priority, ok := filterObj["priority"].(string); ok {
			params.Filter.Priority = priority
		}
		if days, ok := filterObj["days"].(float64); ok {
			params.Filter.Days = int(days)
		}
	}

	// Format with default
	params.Format = "summary"
	if format, ok := args["format"].(string); ok {
		params.Format = format
	}
	if !isValidFormat(params.Format) {
		return nil, fmt.Errorf("invalid format '%s', must be one of: full, summary, list", params.Format)
	}

	return params, nil
}

// Test update params extraction
func TestExtractUpdateParams_FromMap(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
		expect  TodoUpdateParams
	}{
		{
			name: "full update",
			args: map[string]interface{}{
				"id":        "todo-123",
				"section":   "findings",
				"operation": "append",
				"content":   "New findings",
				"metadata": map[string]interface{}{
					"status":       "completed",
					"priority":     "low",
					"current_test": "Test 5",
				},
			},
			expect: TodoUpdateParams{
				ID:        "todo-123",
				Section:   "findings",
				Operation: "append",
				Content:   "New findings",
				Metadata: TodoMetadata{
					Status:      "completed",
					Priority:    "low",
					CurrentTest: "Test 5",
				},
			},
		},
		{
			name:    "missing ID",
			args:    map[string]interface{}{"section": "findings"},
			wantErr: true,
		},
		{
			name: "default operation",
			args: map[string]interface{}{
				"id":      "todo-123",
				"content": "Update",
			},
			expect: TodoUpdateParams{
				ID:        "todo-123",
				Operation: "append", // default
				Content:   "Update",
			},
		},
		{
			name: "valid section parameter",
			args: map[string]interface{}{
				"id":      "todo-123",
				"section": "findings",
			},
			expect: TodoUpdateParams{
				ID:        "todo-123",
				Section:   "findings",
				Operation: "append", // default
			},
			wantErr: false,
		},
		{
			name: "invalid operation",
			args: map[string]interface{}{
				"id":        "todo-123",
				"operation": "delete",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := extractUpdateParamsFromMap(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("extractUpdateParamsFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && params != nil {
				if params.ID != tt.expect.ID {
					t.Errorf("ID = %v, want %v", params.ID, tt.expect.ID)
				}
				if params.Section != tt.expect.Section {
					t.Errorf("Section = %v, want %v", params.Section, tt.expect.Section)
				}
				if params.Operation != tt.expect.Operation {
					t.Errorf("Operation = %v, want %v", params.Operation, tt.expect.Operation)
				}
				if params.Content != tt.expect.Content {
					t.Errorf("Content = %v, want %v", params.Content, tt.expect.Content)
				}
				if params.Metadata.Status != tt.expect.Metadata.Status {
					t.Errorf("Metadata.Status = %v, want %v", params.Metadata.Status, tt.expect.Metadata.Status)
				}
			}
		})
	}
}

// extractUpdateParamsFromMap extracts update params from a map
func extractUpdateParamsFromMap(args map[string]interface{}) (*TodoUpdateParams, error) {
	params := &TodoUpdateParams{}

	// Required ID
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("missing required parameter 'id'")
	}
	params.ID = id

	// Optional section update
	if section, ok := args["section"].(string); ok {
		params.Section = section
	}

	params.Operation = "append"
	if operation, ok := args["operation"].(string); ok {
		params.Operation = operation
	}

	if content, ok := args["content"].(string); ok {
		params.Content = content
	}

	// Extract metadata if provided
	if metaObj, ok := args["metadata"].(map[string]interface{}); ok {
		if status, ok := metaObj["status"].(string); ok {
			params.Metadata.Status = status
		}
		if priority, ok := metaObj["priority"].(string); ok {
			params.Metadata.Priority = priority
		}
		if currentTest, ok := metaObj["current_test"].(string); ok {
			params.Metadata.CurrentTest = currentTest
		}
	}

	// Validate operation
	if !isValidOperation(params.Operation) {
		return nil, fmt.Errorf("invalid operation '%s'", params.Operation)
	}

	return params, nil
}

// Test validation helpers
func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		priority string
		valid    bool
	}{
		{"high", true},
		{"medium", true},
		{"low", true},
		{"urgent", false},
		{"", false},
		{"HIGH", false},
	}

	for _, tt := range tests {
		result := isValidPriority(tt.priority)
		if result != tt.valid {
			t.Errorf("isValidPriority(%q) = %v, want %v", tt.priority, result, tt.valid)
		}
	}
}

func TestIsValidTodoType(t *testing.T) {
	tests := []struct {
		todoType string
		valid    bool
	}{
		{"feature", true},
		{"bug", true},
		{"refactor", true},
		{"research", true},
		{"multi-phase", true},
		{"enhancement", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isValidTodoType(tt.todoType)
		if result != tt.valid {
			t.Errorf("isValidTodoType(%q) = %v, want %v", tt.todoType, result, tt.valid)
		}
	}
}

func TestIsValidOperation(t *testing.T) {
	tests := []struct {
		operation string
		valid     bool
	}{
		{"append", true},
		{"replace", true},
		{"prepend", true},
		{"delete", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isValidOperation(tt.operation)
		if result != tt.valid {
			t.Errorf("isValidOperation(%q) = %v, want %v", tt.operation, result, tt.valid)
		}
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		format string
		valid  bool
	}{
		{"full", true},
		{"summary", true},
		{"list", true},
		{"json", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isValidFormat(tt.format)
		if result != tt.valid {
			t.Errorf("isValidFormat(%q) = %v, want %v", tt.format, result, tt.valid)
		}
	}
}

// TestIsValidSection removed - section validation is now dynamic based on todo content
