package search

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
	
	"github.com/user/mcp-todo-server/internal/domain"
	"gopkg.in/yaml.v3"
)

// TestTodoManager provides todo management for tests without importing core
type TestTodoManager struct {
	basePath string
}

// NewTestTodoManager creates a test todo manager
func NewTestTodoManager(basePath string) *TestTodoManager {
	return &TestTodoManager{basePath: basePath}
}

// CreateTodo creates a test todo
func (tm *TestTodoManager) CreateTodo(task, priority, todoType string) (*domain.Todo, error) {
	// Generate ID
	id := generateTestID(task)
	
	// Create todo
	todo := &domain.Todo{
		ID:       id,
		Task:     task,
		Started:  time.Now(),
		Status:   "in_progress",
		Priority: priority,
		Type:     todoType,
	}
	
	// Write to file
	if err := tm.writeTodo(todo); err != nil {
		return nil, err
	}
	
	return todo, nil
}

// writeTodo writes a todo to disk
func (tm *TestTodoManager) writeTodo(todo *domain.Todo) error {
	// Ensure directory exists
	dir := filepath.Join(tm.basePath, ".claude", "todos")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the todo file with frontmatter
	filename := filepath.Join(dir, fmt.Sprintf("%s.md", todo.ID))
	
	// Create frontmatter
	frontmatter := map[string]interface{}{
		"todo_id":  todo.ID,
		"started":  todo.Started.Format(time.RFC3339),
		"status":   todo.Status,
		"priority": todo.Priority,
		"type":     todo.Type,
	}
	
	if !todo.Completed.IsZero() {
		frontmatter["completed"] = todo.Completed.Format(time.RFC3339)
	}
	
	// Marshal the frontmatter
	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return fmt.Errorf("failed to marshal todo: %w", err)
	}

	// Build content
	content := fmt.Sprintf("---\n%s---\n\n# Task: %s\n\n## Findings & Research\n\n## Web Searches\n\n## Test Strategy\n\n## Test List\n\n## Test Cases\n\n## Maintainability Analysis\n\n## Test Results Log\n\n## Checklist\n\n## Working Scratchpad\n",
		string(yamlData), todo.Task)

	// Write file
	if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write todo file: %w", err)
	}

	return nil
}

// ReadTodo reads a todo from disk
func (tm *TestTodoManager) ReadTodo(id string) (*domain.Todo, error) {
	filePath := filepath.Join(tm.basePath, ".claude", "todos", id+".md")
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("todo not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read todo file: %w", err)
	}
	
	return parseTodoFile(id, string(content))
}

// UpdateTodo updates a todo (simplified for tests)
func (tm *TestTodoManager) UpdateTodo(id, section, operation, content string, metadata map[string]string) error {
	// For tests, we'll just update the file content
	filePath := filepath.Join(tm.basePath, ".claude", "todos", id+".md")
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read todo file: %w", err)
	}
	
	// Simple append operation for tests
	if operation == "append" && content != "" {
		fileContent = append(fileContent, []byte("\n"+content)...)
		return ioutil.WriteFile(filePath, fileContent, 0644)
	}
	
	return nil
}

// generateTestID creates a kebab-case ID from the task description
func generateTestID(task string) string {
	// Simple implementation for tests
	result := ""
	for _, r := range task {
		switch {
		case r >= 'a' && r <= 'z':
			result += string(r)
		case r >= 'A' && r <= 'Z':
			result += string(r + 32) // lowercase
		case r >= '0' && r <= '9':
			result += string(r)
		case r == ' ':
			result += "-"
		}
	}
	return result
}