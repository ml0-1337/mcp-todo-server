package handlers

import (
	"strings"
	"testing"
)

func TestExtractTodoCreateParams_PhaseRequiresParentID(t *testing.T) {
	// Test that phase type requires parent_id
	req := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"task":     "Phase 1: Core Implementation",
			"type":     "phase",
			"priority": "high",
		},
	}

	params, err := ExtractTodoCreateParams(req.ToCallToolRequest())

	// Should fail because phase type requires parent_id
	if err == nil {
		t.Errorf("Expected error for phase type without parent_id, got nil")
	}
	if params != nil {
		t.Errorf("Expected nil params, got %v", params)
	}
	if err != nil && !strings.Contains(err.Error(), "parent_id") {
		t.Errorf("Error should mention parent_id, got: %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "require") {
		t.Errorf("Error should mention require, got: %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "phase") {
		t.Errorf("Error should mention phase, got: %v", err)
	}
}

func TestExtractTodoCreateParams_SubtaskRequiresParentID(t *testing.T) {
	// Test that subtask type requires parent_id
	req := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"task":     "Update documentation",
			"type":     "subtask",
			"priority": "medium",
		},
	}

	params, err := ExtractTodoCreateParams(req.ToCallToolRequest())

	// Should fail because subtask type requires parent_id
	if err == nil {
		t.Errorf("Expected error for subtask type without parent_id, got nil")
	}
	if params != nil {
		t.Errorf("Expected nil params, got %v", params)
	}
	if err != nil && !strings.Contains(err.Error(), "parent_id") {
		t.Errorf("Error should mention parent_id, got: %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "require") {
		t.Errorf("Error should mention require, got: %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "subtask") {
		t.Errorf("Error should mention subtask, got: %v", err)
	}
}

func TestExtractTodoCreateParams_PhaseWithParentIDSucceeds(t *testing.T) {
	// Test that phase type with parent_id succeeds
	req := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"task":      "Phase 1: Core Implementation",
			"type":      "phase",
			"priority":  "high",
			"parent_id": "parent-todo-123",
		},
	}

	params, err := ExtractTodoCreateParams(req.ToCallToolRequest())

	// Should succeed
	if err != nil {
		t.Errorf("Expected no error for phase with parent_id, got: %v", err)
	}
	if params == nil {
		t.Fatal("Expected params, got nil")
	}
	if params.Task != "Phase 1: Core Implementation" {
		t.Errorf("Expected task %q, got %q", "Phase 1: Core Implementation", params.Task)
	}
	if params.Type != "phase" {
		t.Errorf("Expected type %q, got %q", "phase", params.Type)
	}
	if params.ParentID != "parent-todo-123" {
		t.Errorf("Expected parent_id %q, got %q", "parent-todo-123", params.ParentID)
	}
}

func TestExtractTodoCreateParams_RegularTypesDoNotRequireParentID(t *testing.T) {
	// Test that regular types don't require parent_id
	regularTypes := []string{"feature", "bug", "refactor", "research", "multi-phase"}

	for _, todoType := range regularTypes {
		t.Run(todoType, func(t *testing.T) {
			req := &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"task": "Some task",
					"type": todoType,
				},
			}

			params, err := ExtractTodoCreateParams(req.ToCallToolRequest())

			// Should succeed without parent_id
			if err != nil {
				t.Errorf("Expected no error for %s type without parent_id, got: %v", todoType, err)
			}
			if params == nil {
				t.Fatal("Expected params, got nil")
			}
			if params.Type != todoType {
				t.Errorf("Expected type %q, got %q", todoType, params.Type)
			}
			if params.ParentID != "" {
				t.Errorf("Expected empty parent_id, got %q", params.ParentID)
			}
		})
	}
}