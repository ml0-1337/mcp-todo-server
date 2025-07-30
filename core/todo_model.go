package core

import (
	"time"
)

// ChecklistItem represents a single checklist item with status
type ChecklistItem struct {
	Text   string `json:"text"`
	Status string `json:"status"` // "pending", "in_progress", "completed"
}

// Todo represents a todo item
type Todo struct {
	ID        string    `yaml:"todo_id"`
	Task      string    `yaml:"-"` // Task is in the heading, not frontmatter
	Started   time.Time `yaml:"started"`
	Completed time.Time `yaml:"completed,omitempty"`
	Status    string    `yaml:"status"`
	Priority  string    `yaml:"priority"`
	Type      string    `yaml:"type"`
	ParentID  string    `yaml:"parent_id,omitempty"`
	Tags      []string  `yaml:"tags,omitempty"`

	// Section metadata (new)
	Sections map[string]*SectionDefinition `yaml:"sections,omitempty"`
}

// IsCompleted returns true if the todo is completed
func (t *Todo) IsCompleted() bool {
	return t.Status == "completed"
}

// IsBlocked returns true if the todo is blocked
func (t *Todo) IsBlocked() bool {
	return t.Status == "blocked"
}

// IsInProgress returns true if the todo is in progress
func (t *Todo) IsInProgress() bool {
	return t.Status == "in_progress"
}

// HasParent returns true if the todo has a parent
func (t *Todo) HasParent() bool {
	return t.ParentID != ""
}

// GetAge returns the age of the todo since it was started
func (t *Todo) GetAge() time.Duration {
	return time.Since(t.Started)
}

// GetCompletionTime returns the time it took to complete the todo
func (t *Todo) GetCompletionTime() time.Duration {
	if t.IsCompleted() && !t.Completed.IsZero() {
		return t.Completed.Sub(t.Started)
	}
	return 0
}
