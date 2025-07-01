package core

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// getDefaultSections returns the default sections for a new todo
func getDefaultSections() map[string]*SectionDefinition {
	sections := make(map[string]*SectionDefinition)
	order := 0

	// Add all standard sections in order
	standardSections := []struct {
		Key   string
		Title string
		Schema SectionSchema
	}{
		{"findings", "Findings & Research", SchemaResearch},
		{"web_searches", "Web Searches", SchemaResearch},
		{"test_strategy", "Test Strategy", SchemaStrategy},
		{"test_list", "Test List", SchemaChecklist},
		{"tests", "Test Cases", SchemaTestCases},
		{"maintainability", "Maintainability Analysis", SchemaFreeform},
		{"test_results", "Test Results Log", SchemaResults},
		{"checklist", "Checklist", SchemaChecklist},
		{"scratchpad", "Working Scratchpad", SchemaFreeform},
	}

	for _, std := range standardSections {
		order++
		sections[std.Key] = &SectionDefinition{
			Title:    std.Title,
			Order:    order,
			Schema:   std.Schema,
			Required: false,
		}
	}

	return sections
}

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
		Sections: getDefaultSections(),
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
		if parentID, ok := metadata["parent_id"]; ok {
			todo.ParentID = parentID
		}
		if started, ok := metadata["started"]; ok {
			if parsedTime, err := time.Parse(time.RFC3339, started); err == nil {
				todo.Started = parsedTime
			}
		}

		// Update the frontmatter in the content while preserving the rest
		updatedContent, err := updateFrontmatter(string(fileContent), todo)
		if err != nil {
			return fmt.Errorf("failed to update frontmatter: %w", err)
		}
		
		// Write back the updated content
		filename := filepath.Join(tm.basePath, ".claude", "todos", fmt.Sprintf("%s.md", id))
		if err := ioutil.WriteFile(filename, []byte(updatedContent), 0644); err != nil {
			return fmt.Errorf("failed to write updated todo: %w", err)
		}
		
		return nil
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

// updateFrontmatter updates the YAML frontmatter while preserving the rest of the content
func updateFrontmatter(content string, todo *Todo) (string, error) {
	// Split the content into frontmatter and body
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid markdown format: missing frontmatter delimiters")
	}
	
	// Marshal the updated todo to YAML
	yamlData, err := yaml.Marshal(todo)
	if err != nil {
		return "", fmt.Errorf("failed to marshal todo: %w", err)
	}
	
	// Reconstruct the content with updated frontmatter
	return "---\n" + string(yamlData) + "---" + parts[2], nil
}

// appendToSection appends content to a section
func appendToSection(fileContent, section, content string) string {
	lines := strings.Split(fileContent, "\n")
	timestampedContent := formatWithTimestamp(content)
	
	// Find the section
	sectionHeader := "## " + strings.Title(section)
	sectionIndex := -1
	nextSectionIndex := len(lines)
	
	for i, line := range lines {
		if strings.TrimSpace(line) == sectionHeader {
			sectionIndex = i
		} else if sectionIndex > -1 && strings.HasPrefix(strings.TrimSpace(line), "## ") {
			nextSectionIndex = i
			break
		}
	}
	
	if sectionIndex == -1 {
		// Section doesn't exist, append it at the end
		return fileContent + "\n\n" + sectionHeader + "\n\n" + timestampedContent
	}
	
	// Find where to insert the content (before the next section)
	insertIndex := nextSectionIndex
	
	// Insert the content
	newLines := append(lines[:insertIndex], append([]string{timestampedContent}, lines[insertIndex:]...)...)
	return strings.Join(newLines, "\n")
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