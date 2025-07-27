package testutil

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/user/mcp-todo-server/core"
)

// TodoBuilder helps build test todos with custom properties
type TodoBuilder struct {
	task       string
	priority   string
	todoType   string
	status     string
	started    time.Time
	completed  time.Time
	parentID   string
	sections   map[string]string
}

// NewTodoBuilder creates a new todo builder with defaults
func NewTodoBuilder(task string) *TodoBuilder {
	return &TodoBuilder{
		task:     task,
		priority: "high",
		todoType: "feature",
		status:   "in_progress",
		started:  time.Now(),
		sections: make(map[string]string),
	}
}

// WithPriority sets the todo priority
func (b *TodoBuilder) WithPriority(priority string) *TodoBuilder {
	b.priority = priority
	return b
}

// WithType sets the todo type
func (b *TodoBuilder) WithType(todoType string) *TodoBuilder {
	b.todoType = todoType
	return b
}

// WithStatus sets the todo status
func (b *TodoBuilder) WithStatus(status string) *TodoBuilder {
	b.status = status
	return b
}

// WithStarted sets the started date
func (b *TodoBuilder) WithStarted(started time.Time) *TodoBuilder {
	b.started = started
	return b
}

// WithCompleted sets the completed date
func (b *TodoBuilder) WithCompleted(completed time.Time) *TodoBuilder {
	b.completed = completed
	return b
}

// WithParentID sets the parent todo ID
func (b *TodoBuilder) WithParentID(parentID string) *TodoBuilder {
	b.parentID = parentID
	return b
}

// WithSection adds a section with content
func (b *TodoBuilder) WithSection(key, content string) *TodoBuilder {
	b.sections[key] = content
	return b
}

// Build creates the todo object
func (b *TodoBuilder) Build() *core.Todo {
	// Generate ID similar to how core does it
	id := generateTodoID(b.task)
	
	return &core.Todo{
		ID:        id,
		Task:      b.task,
		Priority:  b.priority,
		Type:      b.todoType,
		Status:    b.status,
		Started:   b.started,
		Completed: b.completed,
		ParentID:  b.parentID,
	}
}

// generateTodoID creates a todo ID from the task name
func generateTodoID(task string) string {
	// Simple ID generation - lowercase and replace spaces with hyphens
	id := strings.ToLower(task)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")
	// Remove special characters
	id = strings.ReplaceAll(id, ":", "")
	id = strings.ReplaceAll(id, ",", "")
	id = strings.ReplaceAll(id, ".", "")
	id = strings.ReplaceAll(id, "'", "")
	id = strings.ReplaceAll(id, "\"", "")
	return id
}

// SampleTodos provides a set of sample todos for testing
var SampleTodos = struct {
	FeatureHigh      func() *core.Todo
	BugMedium        func() *core.Todo
	RefactorLow      func() *core.Todo
	MultiPhaseParent func() *core.Todo
	PhaseChild       func(parentID string) *core.Todo
}{
	FeatureHigh: func() *core.Todo {
		return NewTodoBuilder("Implement user authentication").
			WithPriority("high").
			WithType("feature").
			Build()
	},
	
	BugMedium: func() *core.Todo {
		return NewTodoBuilder("Fix login timeout issue").
			WithPriority("medium").
			WithType("bug").
			Build()
	},
	
	RefactorLow: func() *core.Todo {
		return NewTodoBuilder("Refactor database connection pool").
			WithPriority("low").
			WithType("refactor").
			Build()
	},
	
	MultiPhaseParent: func() *core.Todo {
		return NewTodoBuilder("Implement search functionality").
			WithPriority("high").
			WithType("multi-phase").
			Build()
	},
	
	PhaseChild: func(parentID string) *core.Todo {
		return NewTodoBuilder("Phase 1: Design search API").
			WithPriority("medium").
			WithType("phase").
			WithParentID(parentID).
			Build()
	},
}

// GenerateTestContent creates test markdown content for a todo
func GenerateTestContent(todo *core.Todo, sections map[string]string) string {
	completedStr := ""
	if !todo.Completed.IsZero() {
		completedStr = todo.Completed.Format(time.RFC3339)
	}
	
	content := fmt.Sprintf(`---
todo_id: %s
started: %s
completed: %s
status: %s
priority: %s
type: %s
parent_id: %s
---

# Task: %s

`, todo.ID, todo.Started.Format(time.RFC3339), completedStr, 
   todo.Status, todo.Priority, todo.Type, todo.ParentID, todo.Task)
	
	// Add default sections if not provided
	defaultSections := []string{
		"Findings & Research",
		"Web Searches", 
		"Test Strategy",
		"Test List",
		"Test Cases",
		"Test Results Log",
		"Checklist",
		"Working Scratchpad",
	}
	
	for _, section := range defaultSections {
		content += fmt.Sprintf("## %s\n\n", section)
		
		// Add custom content if provided
		key := generateSectionKey(section)
		if customContent, exists := sections[key]; exists {
			content += customContent + "\n\n"
		}
	}
	
	return content
}

// generateSectionKey converts a section title to a key
func generateSectionKey(title string) string {
	// Remove "## " prefix if present
	clean := strings.TrimPrefix(title, "## ")
	
	// Convert to lowercase and replace spaces with underscores
	key := strings.ToLower(clean)
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "&", "and")
	
	// Remove special characters
	key = strings.ReplaceAll(key, "(", "")
	key = strings.ReplaceAll(key, ")", "")
	key = strings.ReplaceAll(key, ":", "")
	key = strings.ReplaceAll(key, ",", "")
	key = strings.ReplaceAll(key, ".", "")
	
	return key
}

// CreateChecklistItems generates test checklist items
func CreateChecklistItems(items ...string) string {
	content := ""
	for i, item := range items {
		if i%3 == 0 {
			content += fmt.Sprintf("- [x] %s\n", item)
		} else if i%3 == 1 {
			content += fmt.Sprintf("- [ ] %s\n", item)
		} else {
			content += fmt.Sprintf("- [>] %s\n", item)
		}
	}
	return content
}

// CreateTestCases generates test case content
func CreateTestCases(cases ...string) string {
	content := ""
	for i, testCase := range cases {
		content += fmt.Sprintf("```go\n// Test %d: %s\nfunc Test%d(t *testing.T) {\n\t// Test implementation\n}\n```\n\n", 
			i+1, testCase, i+1)
	}
	return content
}

// CreateWebSearches generates web search content
func CreateWebSearches(queries ...string) string {
	content := ""
	baseTime := time.Now().Add(-1 * time.Hour)
	
	for i, query := range queries {
		timestamp := baseTime.Add(time.Duration(i*5) * time.Minute)
		content += fmt.Sprintf("[%s] Query: \"%s\"\n", 
			timestamp.Format("2006-01-02 15:04:05"), query)
		content += fmt.Sprintf("- Found relevant information about %s\n\n", query)
	}
	
	return content
}