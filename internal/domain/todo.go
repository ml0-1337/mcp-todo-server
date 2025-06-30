package domain

import (
	"errors"
	"time"
)

// ErrTodoNotFound is returned when a todo is not found
var ErrTodoNotFound = errors.New("todo not found")

// ErrInvalidTodo is returned when a todo fails validation
var ErrInvalidTodo = errors.New("invalid todo")

// Todo represents a todo item in the domain layer
type Todo struct {
	ID        string
	Task      string
	Started   time.Time
	Completed time.Time
	Status    string
	Priority  string
	Type      string
	ParentID  string
	Tags      []string
	Sections  map[string]*SectionDefinition
}

// SectionDefinition represents a section within a todo
type SectionDefinition struct {
	Title      string                 `yaml:"title,omitempty"`
	Content    string                 `yaml:"content,omitempty"`
	Metadata   map[string]interface{} `yaml:"metadata,omitempty"`
	Order      int                    `yaml:"order,omitempty"`
	Visible    bool                   `yaml:"visible,omitempty"`
	Persistent bool                   `yaml:"persistent,omitempty"`
}

// NewTodo creates a new todo with validation
func NewTodo(task, priority, todoType string) (*Todo, error) {
	if task == "" {
		return nil, errors.New("task cannot be empty")
	}
	
	if priority == "" {
		priority = "medium"
	}
	
	if todoType == "" {
		todoType = "task"
	}
	
	return &Todo{
		Task:     task,
		Started:  time.Now(),
		Status:   "in_progress",
		Priority: priority,
		Type:     todoType,
		Sections: make(map[string]*SectionDefinition),
	}, nil
}

// Validate ensures the todo is in a valid state
func (t *Todo) Validate() error {
	if t.ID == "" {
		return errors.New("todo ID cannot be empty")
	}
	if t.Task == "" {
		return errors.New("todo task cannot be empty")
	}
	if t.Status == "" {
		return errors.New("todo status cannot be empty")
	}
	return nil
}

// IsCompleted returns true if the todo is completed
func (t *Todo) IsCompleted() bool {
	return t.Status == "completed"
}

// Complete marks the todo as completed
func (t *Todo) Complete() {
	t.Status = "completed"
	t.Completed = time.Now()
}