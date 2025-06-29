package handlers

import (
	"testing"
)

func TestExtractTodoCreateMultiParams(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr string
		check   func(t *testing.T, params *TodoCreateMultiParams)
	}{
		{
			name: "valid parent and children",
			args: map[string]interface{}{
				"parent": map[string]interface{}{
					"task":     "Main Project",
					"priority": "high",
					"type":     "multi-phase",
				},
				"children": []interface{}{
					map[string]interface{}{
						"task":     "Phase 1",
						"priority": "high",
						"type":     "phase",
					},
					map[string]interface{}{
						"task":     "Phase 2",
						"priority": "medium",
						"type":     "phase",
					},
				},
			},
			check: func(t *testing.T, params *TodoCreateMultiParams) {
				if params.Parent.Task != "Main Project" {
					t.Errorf("parent task = %v, want Main Project", params.Parent.Task)
				}
				if params.Parent.Priority != "high" {
					t.Errorf("parent priority = %v, want high", params.Parent.Priority)
				}
				if params.Parent.Type != "multi-phase" {
					t.Errorf("parent type = %v, want multi-phase", params.Parent.Type)
				}
				if len(params.Children) != 2 {
					t.Errorf("children count = %v, want 2", len(params.Children))
				}
			},
		},
		{
			name: "defaults for optional fields",
			args: map[string]interface{}{
				"parent": map[string]interface{}{
					"task": "Main Project",
				},
				"children": []interface{}{
					map[string]interface{}{
						"task": "Phase 1",
					},
				},
			},
			check: func(t *testing.T, params *TodoCreateMultiParams) {
				if params.Parent.Priority != "high" {
					t.Errorf("parent priority = %v, want high (default)", params.Parent.Priority)
				}
				if params.Parent.Type != "multi-phase" {
					t.Errorf("parent type = %v, want multi-phase (default)", params.Parent.Type)
				}
				if params.Children[0].Priority != "medium" {
					t.Errorf("child priority = %v, want medium (default)", params.Children[0].Priority)
				}
				if params.Children[0].Type != "phase" {
					t.Errorf("child type = %v, want phase (default)", params.Children[0].Type)
				}
			},
		},
		{
			name: "missing parent",
			args: map[string]interface{}{
				"children": []interface{}{
					map[string]interface{}{
						"task": "Phase 1",
					},
				},
			},
			wantErr: "parent is required",
		},
		{
			name: "missing parent task",
			args: map[string]interface{}{
				"parent": map[string]interface{}{
					"priority": "high",
				},
				"children": []interface{}{
					map[string]interface{}{
						"task": "Phase 1",
					},
				},
			},
			wantErr: "parent.task is required",
		},
		{
			name: "missing children",
			args: map[string]interface{}{
				"parent": map[string]interface{}{
					"task": "Main Project",
				},
			},
			wantErr: "children array is required",
		},
		{
			name: "empty children array",
			args: map[string]interface{}{
				"parent": map[string]interface{}{
					"task": "Main Project",
				},
				"children": []interface{}{},
			},
			wantErr: "at least one child is required",
		},
		{
			name: "missing child task",
			args: map[string]interface{}{
				"parent": map[string]interface{}{
					"task": "Main Project",
				},
				"children": []interface{}{
					map[string]interface{}{
						"priority": "high",
					},
				},
			},
			wantErr: "children[0].task is required",
		},
		{
			name: "invalid parent priority",
			args: map[string]interface{}{
				"parent": map[string]interface{}{
					"task":     "Main Project",
					"priority": "invalid",
				},
				"children": []interface{}{
					map[string]interface{}{
						"task": "Phase 1",
					},
				},
			},
			wantErr: "invalid parent priority 'invalid'",
		},
		{
			name: "invalid child priority",
			args: map[string]interface{}{
				"parent": map[string]interface{}{
					"task": "Main Project",
				},
				"children": []interface{}{
					map[string]interface{}{
						"task":     "Phase 1",
						"priority": "invalid",
					},
				},
			},
			wantErr: "invalid priority 'invalid' for child 0",
		},
		{
			name: "child cannot be multi-phase",
			args: map[string]interface{}{
				"parent": map[string]interface{}{
					"task": "Main Project",
				},
				"children": []interface{}{
					map[string]interface{}{
						"task": "Sub Project",
						"type": "multi-phase",
					},
				},
			},
			wantErr: "child 0 cannot be of type 'multi-phase'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &MockCallToolRequest{
				Arguments: tt.args,
			}

			params, err := ExtractTodoCreateMultiParams(req.ToCallToolRequest())

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("ExtractTodoCreateMultiParams() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if err.Error() != tt.wantErr {
					t.Errorf("ExtractTodoCreateMultiParams() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractTodoCreateMultiParams() unexpected error = %v", err)
				return
			}

			if tt.check != nil {
				tt.check(t, params)
			}
		})
	}
}
