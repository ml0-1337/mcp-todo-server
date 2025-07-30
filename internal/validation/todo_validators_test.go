package validation

import (
	"testing"
)

func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority string
		want     bool
	}{
		{
			name:     "valid high priority",
			priority: "high",
			want:     true,
		},
		{
			name:     "valid medium priority",
			priority: "medium",
			want:     true,
		},
		{
			name:     "valid low priority",
			priority: "low",
			want:     true,
		},
		{
			name:     "invalid priority uppercase",
			priority: "HIGH",
			want:     false,
		},
		{
			name:     "invalid priority number",
			priority: "1",
			want:     false,
		},
		{
			name:     "invalid priority empty",
			priority: "",
			want:     false,
		},
		{
			name:     "invalid priority random",
			priority: "critical",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPriority(tt.priority)
			if got != tt.want {
				t.Errorf("IsValidPriority(%q) = %v, want %v", tt.priority, got, tt.want)
			}
		})
	}
}

func TestIsValidTodoType(t *testing.T) {
	tests := []struct {
		name     string
		todoType string
		want     bool
	}{
		{
			name:     "valid feature type",
			todoType: "feature",
			want:     true,
		},
		{
			name:     "valid bug type",
			todoType: "bug",
			want:     true,
		},
		{
			name:     "valid refactor type",
			todoType: "refactor",
			want:     true,
		},
		{
			name:     "valid research type",
			todoType: "research",
			want:     true,
		},
		{
			name:     "valid prd type",
			todoType: "prd",
			want:     true,
		},
		{
			name:     "valid multi-phase type",
			todoType: "multi-phase",
			want:     true,
		},
		{
			name:     "valid phase type",
			todoType: "phase",
			want:     true,
		},
		{
			name:     "valid subtask type",
			todoType: "subtask",
			want:     true,
		},
		{
			name:     "invalid type empty",
			todoType: "",
			want:     false,
		},
		{
			name:     "invalid type uppercase",
			todoType: "FEATURE",
			want:     false,
		},
		{
			name:     "invalid type random",
			todoType: "task",
			want:     false,
		},
		{
			name:     "invalid type with spaces",
			todoType: "multi phase",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTodoType(tt.todoType)
			if got != tt.want {
				t.Errorf("IsValidTodoType(%q) = %v, want %v", tt.todoType, got, tt.want)
			}
		})
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   bool
	}{
		{
			name:   "valid full format",
			format: "full",
			want:   true,
		},
		{
			name:   "valid summary format",
			format: "summary",
			want:   true,
		},
		{
			name:   "valid list format",
			format: "list",
			want:   true,
		},
		{
			name:   "invalid format empty",
			format: "",
			want:   false,
		},
		{
			name:   "invalid format uppercase",
			format: "FULL",
			want:   false,
		},
		{
			name:   "invalid format random",
			format: "detailed",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidFormat(tt.format)
			if got != tt.want {
				t.Errorf("IsValidFormat(%q) = %v, want %v", tt.format, got, tt.want)
			}
		})
	}
}

func TestIsValidOperation(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		want      bool
	}{
		{
			name:      "valid append operation",
			operation: "append",
			want:      true,
		},
		{
			name:      "valid replace operation",
			operation: "replace",
			want:      true,
		},
		{
			name:      "valid prepend operation",
			operation: "prepend",
			want:      true,
		},
		{
			name:      "valid toggle operation",
			operation: "toggle",
			want:      true,
		},
		{
			name:      "invalid operation empty",
			operation: "",
			want:      false,
		},
		{
			name:      "invalid operation uppercase",
			operation: "APPEND",
			want:      false,
		},
		{
			name:      "invalid operation random",
			operation: "insert",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidOperation(tt.operation)
			if got != tt.want {
				t.Errorf("IsValidOperation(%q) = %v, want %v", tt.operation, got, tt.want)
			}
		})
	}
}

func TestGetValidPriorities(t *testing.T) {
	expected := []string{"high", "medium", "low"}
	got := GetValidPriorities()

	if len(got) != len(expected) {
		t.Errorf("GetValidPriorities() returned %d items, want %d", len(got), len(expected))
	}

	for i, v := range expected {
		if i >= len(got) || got[i] != v {
			t.Errorf("GetValidPriorities()[%d] = %q, want %q", i, got[i], v)
		}
	}
}

func TestGetValidTodoTypes(t *testing.T) {
	expected := []string{"feature", "bug", "refactor", "research", "prd", "multi-phase", "phase", "subtask"}
	got := GetValidTodoTypes()

	if len(got) != len(expected) {
		t.Errorf("GetValidTodoTypes() returned %d items, want %d", len(got), len(expected))
	}

	for i, v := range expected {
		if i >= len(got) || got[i] != v {
			t.Errorf("GetValidTodoTypes()[%d] = %q, want %q", i, got[i], v)
		}
	}
}

func TestGetValidFormats(t *testing.T) {
	expected := []string{"full", "summary", "list"}
	got := GetValidFormats()

	if len(got) != len(expected) {
		t.Errorf("GetValidFormats() returned %d items, want %d", len(got), len(expected))
	}

	for i, v := range expected {
		if i >= len(got) || got[i] != v {
			t.Errorf("GetValidFormats()[%d] = %q, want %q", i, got[i], v)
		}
	}
}

func TestGetValidOperations(t *testing.T) {
	expected := []string{"append", "replace", "prepend", "toggle"}
	got := GetValidOperations()

	if len(got) != len(expected) {
		t.Errorf("GetValidOperations() returned %d items, want %d", len(got), len(expected))
	}

	for i, v := range expected {
		if i >= len(got) || got[i] != v {
			t.Errorf("GetValidOperations()[%d] = %q, want %q", i, got[i], v)
		}
	}
}
