package core

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	
	interrors "github.com/user/mcp-todo-server/internal/errors"
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
	return tm.CreateTodoWithTemplate(task, priority, todoType, "")
}

// CreateTodoWithTemplate creates a new todo with optional template content
func (tm *TodoManager) CreateTodoWithTemplate(task, priority, todoType, templateContent string) (*Todo, error) {
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

	// Write todo to file (with template content if provided)
	if templateContent != "" {
		err := tm.writeTodoWithContent(todo, templateContent)
		if err != nil {
			return nil, interrors.Wrap(err, "failed to write todo with template")
		}
	} else {
		err := tm.writeTodo(todo)
		if err != nil {
			return nil, interrors.Wrap(err, "failed to write todo")
		}
	}

	return todo, nil
}

// UpdateTodo updates a todo's content or metadata
func (tm *TodoManager) UpdateTodo(id, section, operation, content string, metadata map[string]string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Read existing todo
	filename, err := ResolveTodoPath(tm.basePath, id)
	if err != nil {
		if os.IsNotExist(err) {
			return interrors.NewNotFoundError("todo", id)
		}
		return interrors.Wrap(err, "failed to resolve todo path")
	}
	
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return interrors.NewNotFoundError("todo", id)
		}
		return interrors.Wrap(err, "failed to read todo")
	}

	// Handle metadata updates (like status changes)
	if section == "" && metadata != nil && len(metadata) > 0 {
		// Parse the file to get the todo
		todo, err := tm.parseTodoFile(string(fileContent))
		if err != nil {
			return interrors.Wrap(err, "failed to parse todo")
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
			// Try multiple time formats
			formats := []string{
				time.RFC3339,
				"2006-01-02 15:04:05",
				"2006-01-02T15:04:05Z",
			}
			for _, format := range formats {
				if parsedTime, err := time.Parse(format, started); err == nil {
					todo.Started = parsedTime
					break
				}
			}
		}
		if completed, ok := metadata["completed"]; ok {
			// Try multiple time formats
			formats := []string{
				time.RFC3339,
				"2006-01-02 15:04:05",
				"2006-01-02T15:04:05Z",
			}
			for _, format := range formats {
				if parsedTime, err := time.Parse(format, completed); err == nil {
					todo.Completed = parsedTime
					break
				}
			}
		}

		// Update the frontmatter in the content while preserving the rest
		updatedContent, err := updateFrontmatter(string(fileContent), todo)
		if err != nil {
			return interrors.Wrap(err, "failed to update frontmatter")
		}
		
		// Write back the updated content
		if err := ioutil.WriteFile(filename, []byte(updatedContent), 0644); err != nil {
			return interrors.NewOperationError("write", "todo file", "failed to save changes", err)
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
	} else if operation == "prepend" {
		// Prepend operation
		updatedContent = prependToSection(updatedContent, section, content)
	}

	// Resolve the path and write back to file
	filename, err := ResolveTodoPath(tm.basePath, id)
	if err != nil {
		return interrors.Wrap(err, "failed to resolve todo path for update")
	}
	
	if err := ioutil.WriteFile(filename, []byte(updatedContent), 0644); err != nil {
		return interrors.NewOperationError("write", "todo section", "failed to save section update", err)
	}

	return nil
}

// updateFrontmatter updates the YAML frontmatter while preserving the rest of the content
func updateFrontmatter(content string, todo *Todo) (string, error) {
	// Split the content into frontmatter and body
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return "", interrors.NewValidationError("content", content, "invalid markdown format: missing frontmatter delimiters")
	}
	
	// Marshal the updated todo to YAML
	yamlData, err := yaml.Marshal(todo)
	if err != nil {
		return "", interrors.Wrap(err, "failed to marshal todo")
	}
	
	// Reconstruct the content with updated frontmatter
	return "---\n" + string(yamlData) + "---" + parts[2], nil
}

// appendToSection appends content to a section
func appendToSection(fileContent, section, content string) string {
	lines := strings.Split(fileContent, "\n")
	
	// Map section names to their proper titles
	sectionTitles := map[string]string{
		"findings":      "Findings & Research",
		"web_searches":  "Web Searches",
		"test_strategy": "Test Strategy",
		"test_list":     "Test List",
		"tests":         "Test Cases",
		"test_results":  "Test Results Log",
		"checklist":     "Checklist",
		"scratchpad":    "Working Scratchpad",
	}
	
	// Get the proper section title
	sectionTitle := sectionTitles[section]
	if sectionTitle == "" {
		// Default to titlecase if not in map
		sectionTitle = strings.Title(strings.Replace(section, "_", " ", -1))
	}
	
	// Only add timestamp for test_results section
	contentToAppend := content
	if section == "test_results" {
		contentToAppend = formatWithTimestamp(content)
	}
	
	sectionHeader := "## " + sectionTitle
	sectionIndex := -1
	nextSectionIndex := len(lines)
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == sectionHeader {
			sectionIndex = i
		} else if sectionIndex > -1 && strings.HasPrefix(trimmed, "## ") {
			nextSectionIndex = i
			break
		}
	}
	
	if sectionIndex == -1 {
		// Section doesn't exist, append it at the end
		return fileContent + "\n\n" + sectionHeader + "\n\n" + contentToAppend
	}
	
	// Find the last non-empty line in the section
	lastContentIndex := sectionIndex
	for i := sectionIndex + 1; i < nextSectionIndex; i++ {
		if strings.TrimSpace(lines[i]) != "" {
			lastContentIndex = i
		}
	}
	
	// Insert after the last content
	insertIndex := lastContentIndex + 1
	
	// Handle spacing
	if lastContentIndex > sectionIndex {
		// There's existing content, add newline before appending
		contentToAppend = "\n" + contentToAppend
	} else {
		// Section is empty
		// Check if there's already an empty line after the header
		if insertIndex < len(lines) && strings.TrimSpace(lines[insertIndex]) == "" {
			// There's already an empty line, insert after it to preserve spacing
			insertIndex++
		}
		// No need to add extra newline, the structure is already correct
	}
	
	// Insert the content
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIndex]...)
	newLines = append(newLines, contentToAppend)
	newLines = append(newLines, lines[insertIndex:]...)
	
	return strings.Join(newLines, "\n")
}

// replaceSection replaces a section's content
func replaceSection(fileContent, section, content string) string {
	lines := strings.Split(fileContent, "\n")
	
	// Map section names to their proper titles
	sectionTitles := map[string]string{
		"findings":      "Findings & Research",
		"web_searches":  "Web Searches",
		"test_strategy": "Test Strategy",
		"test_list":     "Test List",
		"tests":         "Test Cases",
		"test_results":  "Test Results Log",
		"checklist":     "Checklist",
		"scratchpad":    "Working Scratchpad",
	}
	
	// Get the proper section title
	sectionTitle := sectionTitles[section]
	if sectionTitle == "" {
		// Default to titlecase if not in map
		sectionTitle = strings.Title(strings.Replace(section, "_", " ", -1))
	}
	
	sectionHeader := "## " + sectionTitle
	sectionIndex := -1
	nextSectionIndex := len(lines)
	
	// Find the section
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == sectionHeader {
			sectionIndex = i
		} else if sectionIndex > -1 && strings.HasPrefix(trimmed, "## ") {
			nextSectionIndex = i
			break
		}
	}
	
	if sectionIndex == -1 {
		// Section doesn't exist, append it at the end
		return fileContent + "\n\n" + sectionHeader + "\n\n" + content
	}
	
	// Build the new content
	newLines := make([]string, 0, len(lines))
	
	// Add all lines before the section
	newLines = append(newLines, lines[:sectionIndex+1]...)
	
	// Add empty line after header if content is not empty
	if content != "" {
		newLines = append(newLines, "")
		// Add the new content
		contentLines := strings.Split(strings.TrimSpace(content), "\n")
		newLines = append(newLines, contentLines...)
	}
	
	// Add lines after the section (skip old section content)
	if nextSectionIndex < len(lines) {
		// Add empty line before next section if we added content
		if content != "" {
			newLines = append(newLines, "")
		}
		// Add all remaining lines starting from the next section
		newLines = append(newLines, lines[nextSectionIndex:]...)
	}
	
	return strings.Join(newLines, "\n")
}

// prependToSection prepends content to a section
func prependToSection(fileContent, section, content string) string {
	lines := strings.Split(fileContent, "\n")
	
	// Map section names to their proper titles
	sectionTitles := map[string]string{
		"findings":      "Findings & Research",
		"web_searches":  "Web Searches",
		"test_strategy": "Test Strategy",
		"test_list":     "Test List",
		"tests":         "Test Cases",
		"test_results":  "Test Results Log",
		"checklist":     "Checklist",
		"scratchpad":    "Working Scratchpad",
	}
	
	// Get the proper section title
	sectionTitle := sectionTitles[section]
	if sectionTitle == "" {
		// Default to titlecase if not in map
		sectionTitle = strings.Title(strings.Replace(section, "_", " ", -1))
	}
	
	// Only add timestamp for test_results section
	contentToPrepend := content
	if section == "test_results" {
		contentToPrepend = formatWithTimestamp(content)
	}
	
	sectionHeader := "## " + sectionTitle
	sectionIndex := -1
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == sectionHeader {
			sectionIndex = i
			break
		}
	}
	
	if sectionIndex == -1 {
		// Section doesn't exist, append it at the end
		return fileContent + "\n\n" + sectionHeader + "\n\n" + contentToPrepend
	}
	
	// Find where to insert the content (right after the section header)
	insertIndex := sectionIndex + 1
	
	// Skip the first empty line after header if it exists
	if insertIndex < len(lines) && strings.TrimSpace(lines[insertIndex]) == "" {
		insertIndex++
	}
	
	// Insert the content at the beginning of the section
	newLines := make([]string, 0, len(lines)+2)
	newLines = append(newLines, lines[:insertIndex]...)
	newLines = append(newLines, contentToPrepend)
	
	// Add empty line if there's existing content
	if insertIndex < len(lines) && strings.TrimSpace(lines[insertIndex]) != "" {
		newLines = append(newLines, "")
	}
	
	newLines = append(newLines, lines[insertIndex:]...)
	
	return strings.Join(newLines, "\n")
}

// ListTodos lists todos based on filters
func (tm *TodoManager) ListTodos(status, priority string, days int) ([]*Todo, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var todos []*Todo
	todosRoot := filepath.Join(tm.basePath, ".claude", "todos")
	cutoffTime := time.Now().AddDate(0, 0, -days)

	// Check if todos directory exists
	if _, err := os.Stat(todosRoot); os.IsNotExist(err) {
		return []*Todo{}, nil
	}

	// Optimization: If days filter is active and reasonable, only scan relevant directories
	if days > 0 && days < 365 {
		// Use ScanDateRange for optimized scanning
		startDate := cutoffTime
		endDate := time.Now()
		
		paths, err := ScanDateRange(tm.basePath, startDate, endDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ListTodos: Error in ScanDateRange: %v\n", err)
			// Fall back to full scan on error
		} else {
			// Process found files
			for _, path := range paths {
				content, err := ioutil.ReadFile(path)
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

				todos = append(todos, todo)
			}
			return todos, nil
		}
	}

	// Full recursive scan for no date filter or fallback
	err := filepath.Walk(todosRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on error
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		todo, err := tm.parseTodoFile(string(content))
		if err != nil {
			return nil // Skip malformed files
		}

		// Apply filters (treat "all" as wildcard)
		if status != "" && status != "all" && !strings.EqualFold(todo.Status, status) {
			return nil
		}

		if priority != "" && priority != "all" && !strings.EqualFold(todo.Priority, priority) {
			return nil
		}

		if days > 0 && todo.Started.Before(cutoffTime) {
			return nil
		}

		todos = append(todos, todo)
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to walk todos directory: %w", err)
	}

	// Sort by started date (newest first)
	sort.Slice(todos, func(i, j int) bool {
		return todos[i].Started.After(todos[j].Started)
	})

	return todos, nil
}