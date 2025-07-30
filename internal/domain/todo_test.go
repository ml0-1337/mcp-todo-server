package domain

import (
	"testing"
	"time"
)

func TestNewTodo(t *testing.T) {
	tests := []struct {
		name     string
		task     string
		priority string
		todoType string
		wantErr  bool
		validate func(t *testing.T, todo *Todo)
	}{
		{
			name:     "valid todo with all fields",
			task:     "Test task",
			priority: "high",
			todoType: "feature",
			wantErr:  false,
			validate: func(t *testing.T, todo *Todo) {
				if todo.Task != "Test task" {
					t.Errorf("Task = %q, want %q", todo.Task, "Test task")
				}
				if todo.Priority != "high" {
					t.Errorf("Priority = %q, want %q", todo.Priority, "high")
				}
				if todo.Type != "feature" {
					t.Errorf("Type = %q, want %q", todo.Type, "feature")
				}
				if todo.Status != "in_progress" {
					t.Errorf("Status = %q, want %q", todo.Status, "in_progress")
				}
				if todo.Sections == nil {
					t.Error("Sections should be initialized")
				}
			},
		},
		{
			name:     "empty task should error",
			task:     "",
			priority: "high",
			todoType: "bug",
			wantErr:  true,
		},
		{
			name:     "default priority when empty",
			task:     "Test task",
			priority: "",
			todoType: "feature",
			wantErr:  false,
			validate: func(t *testing.T, todo *Todo) {
				if todo.Priority != "medium" {
					t.Errorf("Priority = %q, want %q", todo.Priority, "medium")
				}
			},
		},
		{
			name:     "default type when empty",
			task:     "Test task",
			priority: "low",
			todoType: "",
			wantErr:  false,
			validate: func(t *testing.T, todo *Todo) {
				if todo.Type != "task" {
					t.Errorf("Type = %q, want %q", todo.Type, "task")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo, err := NewTodo(tt.task, tt.priority, tt.todoType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTodo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, todo)
			}
		})
	}
}

func TestTodo_Validate(t *testing.T) {
	tests := []struct {
		name    string
		todo    *Todo
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid todo",
			todo: &Todo{
				ID:     "test-id",
				Task:   "Test task",
				Status: "in_progress",
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			todo: &Todo{
				ID:     "",
				Task:   "Test task",
				Status: "in_progress",
			},
			wantErr: true,
			errMsg:  "todo ID cannot be empty",
		},
		{
			name: "empty task",
			todo: &Todo{
				ID:     "test-id",
				Task:   "",
				Status: "in_progress",
			},
			wantErr: true,
			errMsg:  "todo task cannot be empty",
		},
		{
			name: "empty status",
			todo: &Todo{
				ID:     "test-id",
				Task:   "Test task",
				Status: "",
			},
			wantErr: true,
			errMsg:  "todo status cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.todo.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestTodo_IsCompleted(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "completed status",
			status: "completed",
			want:   true,
		},
		{
			name:   "in_progress status",
			status: "in_progress",
			want:   false,
		},
		{
			name:   "blocked status",
			status: "blocked",
			want:   false,
		},
		{
			name:   "empty status",
			status: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo := &Todo{Status: tt.status}
			if got := todo.IsCompleted(); got != tt.want {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTodo_Complete(t *testing.T) {
	todo := &Todo{
		ID:     "test-id",
		Task:   "Test task",
		Status: "in_progress",
	}

	// Record time before completion
	beforeComplete := time.Now()

	// Complete the todo
	todo.Complete()

	// Record time after completion
	afterComplete := time.Now()

	// Verify status changed
	if todo.Status != "completed" {
		t.Errorf("Status = %q, want %q", todo.Status, "completed")
	}

	// Verify completed time was set
	if todo.Completed.IsZero() {
		t.Error("Completed time should be set")
	}

	// Verify completed time is reasonable
	if todo.Completed.Before(beforeComplete) || todo.Completed.After(afterComplete) {
		t.Error("Completed time is outside expected range")
	}
}

func TestSectionDefinition(t *testing.T) {
	// Test section creation
	section := &SectionDefinition{
		Title:   "Test Section",
		Content: "Test content",
		Metadata: map[string]interface{}{
			"key": "value",
		},
		Order:      1,
		Visible:    true,
		Persistent: false,
	}

	// Verify fields
	if section.Title != "Test Section" {
		t.Errorf("Title = %q, want %q", section.Title, "Test Section")
	}
	if section.Content != "Test content" {
		t.Errorf("Content = %q, want %q", section.Content, "Test content")
	}
	if section.Order != 1 {
		t.Errorf("Order = %d, want %d", section.Order, 1)
	}
	if !section.Visible {
		t.Error("Visible should be true")
	}
	if section.Persistent {
		t.Error("Persistent should be false")
	}
	if section.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %v, want %v", section.Metadata["key"], "value")
	}
}

func TestTodo_WithSections(t *testing.T) {
	todo, err := NewTodo("Task with sections", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Add sections
	todo.Sections["findings"] = &SectionDefinition{
		Title:   "Findings",
		Content: "Research findings here",
		Order:   1,
		Visible: true,
	}

	todo.Sections["tests"] = &SectionDefinition{
		Title:   "Tests",
		Content: "Test results here",
		Order:   2,
		Visible: true,
	}

	// Verify sections
	if len(todo.Sections) != 2 {
		t.Errorf("Expected 2 sections, got %d", len(todo.Sections))
	}

	findings, ok := todo.Sections["findings"]
	if !ok {
		t.Error("Findings section not found")
	} else {
		if findings.Title != "Findings" {
			t.Errorf("Findings title = %q, want %q", findings.Title, "Findings")
		}
		if findings.Order != 1 {
			t.Errorf("Findings order = %d, want %d", findings.Order, 1)
		}
	}

	tests, ok := todo.Sections["tests"]
	if !ok {
		t.Error("Tests section not found")
	} else {
		if tests.Title != "Tests" {
			t.Errorf("Tests title = %q, want %q", tests.Title, "Tests")
		}
		if tests.Order != 2 {
			t.Errorf("Tests order = %d, want %d", tests.Order, 2)
		}
	}
}

func TestTodo_StateTransitions(t *testing.T) {
	todo, err := NewTodo("State transition test", "medium", "bug")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Initial state
	if todo.Status != "in_progress" {
		t.Errorf("Initial status = %q, want %q", todo.Status, "in_progress")
	}
	if !todo.Completed.IsZero() {
		t.Error("Completed time should be zero initially")
	}

	// Complete the todo
	todo.Complete()

	// Verify completed state
	if todo.Status != "completed" {
		t.Errorf("Status after complete = %q, want %q", todo.Status, "completed")
	}
	if todo.Completed.IsZero() {
		t.Error("Completed time should be set after completion")
	}
	if !todo.IsCompleted() {
		t.Error("IsCompleted() should return true after completion")
	}
}

func TestTodo_WithTags(t *testing.T) {
	todo, err := NewTodo("Task with tags", "low", "refactor")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Add tags
	todo.Tags = []string{"backend", "api", "urgent"}

	// Verify tags
	if len(todo.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(todo.Tags))
	}

	expectedTags := []string{"backend", "api", "urgent"}
	for i, tag := range expectedTags {
		if i >= len(todo.Tags) || todo.Tags[i] != tag {
			t.Errorf("Tag[%d] = %q, want %q", i, todo.Tags[i], tag)
		}
	}
}

func TestTodo_WithParent(t *testing.T) {
	parent, err := NewTodo("Parent task", "high", "multi-phase")
	if err != nil {
		t.Fatalf("Failed to create parent todo: %v", err)
	}
	parent.ID = "parent-id"

	child, err := NewTodo("Child task", "high", "phase")
	if err != nil {
		t.Fatalf("Failed to create child todo: %v", err)
	}
	child.ID = "child-id"
	child.ParentID = parent.ID

	// Verify parent-child relationship
	if child.ParentID != parent.ID {
		t.Errorf("Child ParentID = %q, want %q", child.ParentID, parent.ID)
	}
}
