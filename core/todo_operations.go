package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CreateTodo creates a new todo with a unique ID
func (tm *TodoManager) CreateTodo(task, priority, todoType string) (*Todo, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Generate unique ID from task
	baseID := generateBaseID(task)

	// Ensure uniqueness
	finalID := baseID
	if count, exists := tm.idCounts[baseID]; exists {
		finalID = fmt.Sprintf("%s-%d", baseID, count+1)
		tm.idCounts[baseID] = count + 1
	} else {
		tm.idCounts[baseID] = 1
	}

	// Create todo
	todo := &Todo{
		ID:       finalID,
		Task:     task,
		Started:  time.Now(),
		Status:   "in_progress",
		Priority: priority,
		Type:     todoType,
	}

	// Write todo to file
	err := tm.writeTodo(todo)
	if err != nil {
		return nil, fmt.Errorf("failed to write todo: %w", err)
	}

	return todo, nil
}

// UpdateTodo updates a todo's content or metadata
func (tm *TodoManager) UpdateTodo(id, section, operation, content string, metadata map[string]string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Read existing todo
	filename := filepath.Join(tm.basePath, ".claude", "todos", fmt.Sprintf("%s.md", id))
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read todo: %w", err)
	}

	// Handle metadata updates (like status changes)
	if section == "" && metadata != nil && len(metadata) > 0 {
		// Parse the file to get the todo
		todo, err := tm.parseTodoFile(string(fileContent))
		if err != nil {
			return fmt.Errorf("failed to parse todo: %w", err)
		}

		// Update metadata fields
		if status, ok := metadata["status"]; ok {
			todo.Status = status
			if status == "completed" {
				todo.Completed = time.Now()
			}
		}
		if priority, ok := metadata["priority"]; ok {
			todo.Priority = priority
		}
		if todoType, ok := metadata["type"]; ok {
			todo.Type = todoType
		}

		// Write back
		return tm.writeTodo(todo)
	}

	// For section updates, use sophisticated section-aware update
	return tm.updateTodoSection(id, string(fileContent), section, operation, content)
}

// updateTodoSection handles section-specific updates
func (tm *TodoManager) updateTodoSection(id, fileContent, section, operation, content string) error {
	// For now, implement a simple section update
	// This will be replaced with the sophisticated section updater later
	
	updatedContent := fileContent
	
	// Handle special cases
	if section == "checklist" && operation == "toggle" {
		updatedContent = toggleChecklistItem(updatedContent, content)
	} else if operation == "append" {
		// Simple append operation
		updatedContent = appendToSection(updatedContent, section, content)
	} else if operation == "replace" {
		// Simple replace operation
		updatedContent = replaceSection(updatedContent, section, content)
	}

	// Write back to file
	filename := filepath.Join(tm.basePath, ".claude", "todos", fmt.Sprintf("%s.md", id))
	if err := ioutil.WriteFile(filename, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated todo: %w", err)
	}

	return nil
}

// appendToSection appends content to a section
func appendToSection(fileContent, section, content string) string {
	// This is a placeholder - the real implementation would be more sophisticated
	// For now, just append the content with timestamp
	timestampedContent := formatWithTimestamp(content)
	return fileContent + "\n" + timestampedContent
}

// replaceSection replaces a section's content
func replaceSection(fileContent, section, content string) string {
	// This is a placeholder - the real implementation would be more sophisticated
	return fileContent
}

// ListTodos lists todos based on filters
func (tm *TodoManager) ListTodos(status, priority string, days int) ([]*Todo, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	todosDir := filepath.Join(tm.basePath, ".claude", "todos")
	files, err := ioutil.ReadDir(todosDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Todo{}, nil
		}
		return nil, fmt.Errorf("failed to read todos directory: %w", err)
	}

	var todos []*Todo
	cutoffTime := time.Now().AddDate(0, 0, -days)

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".md" {
			continue
		}

		content, err := ioutil.ReadFile(filepath.Join(todosDir, file.Name()))
		if err != nil {
			continue
		}

		todo, err := tm.parseTodoFile(string(content))
		if err != nil {
			continue
		}

		// Apply filters
		if status != "" && !strings.EqualFold(todo.Status, status) {
			continue
		}

		if priority != "" && !strings.EqualFold(todo.Priority, priority) {
			continue
		}

		if days > 0 && todo.Started.Before(cutoffTime) {
			continue
		}

		todos = append(todos, todo)
	}

	return todos, nil
}