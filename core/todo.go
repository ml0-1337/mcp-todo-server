package core

import (
	"fmt"
	"strings"
	"time"
	"sync"
	"os"
	"path/filepath"
	"io/ioutil"
	"gopkg.in/yaml.v3"
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
	Sections  map[string]*SectionDefinition `yaml:"sections,omitempty"`
}

// TodoManager handles todo operations
type TodoManager struct {
	basePath string
	mu       sync.Mutex
	idCounts map[string]int // Track ID usage for uniqueness
}

// NewTodoManager creates a new todo manager
func NewTodoManager(basePath string) *TodoManager {
	return &TodoManager{
		basePath: basePath,
		idCounts: make(map[string]int),
	}
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
	}
	
	// Write todo to file
	err := tm.writeTodo(todo)
	if err != nil {
		return nil, fmt.Errorf("failed to write todo: %w", err)
	}
	
	return todo, nil
}

// generateBaseID creates a kebab-case ID from the task description
func generateBaseID(task string) string {
	// Remove null bytes and other invalid characters first
	cleaned := strings.ReplaceAll(task, "\x00", "")
	
	// Convert to lowercase
	lower := strings.ToLower(cleaned)
	
	// Replace spaces and special characters with hyphens
	// Keep numbers and dots for version numbers
	replacer := strings.NewReplacer(
		" ", "-",
		"_", "-",
		":", "",
		";", "",
		",", "",
		"!", "",
		"?", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
		"/", "-",
		"\\", "-",
		"\"", "",
		"'", "",
		"@", "",
		"#", "",
		"$", "",
		"%", "",
		"^", "",
		"&", "",
		"*", "",
		"+", "",
		"=", "",
		"|", "",
		"<", "",
		">", "",
		"`", "",
		"~", "",
	)
	
	kebab := replacer.Replace(lower)
	
	// Replace dots between numbers with nothing (v2.3.4 -> v234)
	// but keep other dots as separators
	kebab = strings.ReplaceAll(kebab, ".", "")
	
	// Remove multiple consecutive hyphens
	for strings.Contains(kebab, "--") {
		kebab = strings.ReplaceAll(kebab, "--", "-")
	}
	
	// Trim hyphens from start and end
	kebab = strings.Trim(kebab, "-")
	
	// If empty after processing, use default
	if kebab == "" {
		return "todo"
	}
	
	// Limit length to make IDs manageable
	if len(kebab) > 50 {
		kebab = kebab[:50]
		// Ensure we don't cut off in the middle of a word
		lastHyphen := strings.LastIndex(kebab, "-")
		if lastHyphen > 30 {
			kebab = kebab[:lastHyphen]
		}
	}
	
	// Final trim in case truncation left a trailing hyphen
	kebab = strings.Trim(kebab, "-")
	
	// Final check for empty string
	if kebab == "" {
		return "todo"
	}
	
	return kebab
}

// writeTodo writes a todo to disk in markdown format with YAML frontmatter
func (tm *TodoManager) writeTodo(todo *Todo) error {
	// Ensure directory exists
	err := os.MkdirAll(tm.basePath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Generate file path
	filePath := filepath.Join(tm.basePath, todo.ID + ".md")
	
	// Format timestamp
	timestamp := todo.Started.Format("2006-01-02 15:04:05")
	
	// Create YAML frontmatter
	frontmatter := map[string]interface{}{
		"todo_id":   todo.ID,
		"started":   timestamp,
		"completed": "",
		"status":    todo.Status,
		"priority":  todo.Priority,
		"type":      todo.Type,
	}
	
	// Add sections if defined
	if todo.Sections != nil && len(todo.Sections) > 0 {
		frontmatter["sections"] = todo.Sections
	}
	
	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	
	// Build markdown content
	var contentBuilder strings.Builder
	contentBuilder.WriteString("---\n")
	contentBuilder.Write(yamlData)
	contentBuilder.WriteString("---\n\n")
	contentBuilder.WriteString(fmt.Sprintf("# Task: %s\n\n", todo.Task))
	
	// Generate sections based on metadata or use defaults
	if todo.Sections != nil && len(todo.Sections) > 0 {
		// Use defined sections in order
		ordered := GetOrderedSections(todo.Sections)
		for _, section := range ordered {
			contentBuilder.WriteString(section.Definition.Title + "\n\n")
		}
	} else {
		// Use default sections for new todos
		contentBuilder.WriteString(`## Findings & Research

## Test Strategy

## Test List

## Test Cases

## Maintainability Analysis

## Test Results Log

## Checklist

## Working Scratchpad
`)
	}
	
	content := contentBuilder.String()
	
	// Write to temp file first (atomic write)
	tempFile := filePath + ".tmp"
	err = ioutil.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Rename temp file to final location
	err = os.Rename(tempFile, filePath)
	if err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename file: %w", err)
	}
	
	return nil
}

// ReadTodo reads and parses a todo file by ID
func (tm *TodoManager) ReadTodo(id string) (*Todo, error) {
	// Construct file path
	filePath := filepath.Join(tm.basePath, id + ".md")
	
	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("todo not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read todo file: %w", err)
	}
	
	// Parse the file
	todo, err := tm.parseTodoFile(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse todo file: %w", err)
	}
	
	return todo, nil
}

// parseTodoFile parses markdown content with YAML frontmatter into a Todo
func (tm *TodoManager) parseTodoFile(content string) (*Todo, error) {
	// Split content by frontmatter delimiters
	parts := strings.Split(content, "---\n")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown format: missing frontmatter delimiters")
	}
	
	// Parse YAML frontmatter
	yamlContent := parts[1]
	
	// Define a struct for parsing with string timestamps
	var frontmatter struct {
		TodoID    string   `yaml:"todo_id"`
		Started   string   `yaml:"started"`
		Completed string   `yaml:"completed"`
		Status    string   `yaml:"status"`
		Priority  string   `yaml:"priority"`
		Type      string   `yaml:"type"`
		ParentID  string   `yaml:"parent_id"`
		Tags      []string `yaml:"tags"`
		Sections  map[string]*SectionDefinition `yaml:"sections"`
	}
	
	err := yaml.Unmarshal([]byte(yamlContent), &frontmatter)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}
	
	// Create todo struct
	todo := &Todo{
		ID:       frontmatter.TodoID,
		Status:   frontmatter.Status,
		Priority: frontmatter.Priority,
		Type:     frontmatter.Type,
		ParentID: frontmatter.ParentID,
		Tags:     frontmatter.Tags,
		Sections: frontmatter.Sections,
	}
	
	// Parse timestamps with multiple format support
	if frontmatter.Started != "" {
		startTime, err := parseTimestamp(frontmatter.Started)
		if err != nil {
			return nil, fmt.Errorf("failed to parse started timestamp: %w", err)
		}
		todo.Started = startTime
	}
	
	if frontmatter.Completed != "" {
		completedTime, err := parseTimestamp(frontmatter.Completed)
		if err != nil {
			return nil, fmt.Errorf("failed to parse completed timestamp: %w", err)
		}
		todo.Completed = completedTime
	}
	
	// Extract task from markdown content
	markdownContent := parts[2]
	todo.Task = extractTask(markdownContent)
	
	// Handle section metadata - if not defined, infer from markdown
	if todo.Sections == nil {
		todo.Sections = InferSectionsFromMarkdown(markdownContent)
	}
	
	return todo, nil
}

// parseTimestamp attempts to parse a timestamp string using multiple formats
func parseTimestamp(timestamp string) (time.Time, error) {
	// Try multiple formats in order of preference
	formats := []string{
		"2006-01-02 15:04:05",   // Standard format used by the system
		time.RFC3339,            // RFC3339 format (2006-01-02T15:04:05Z07:00)
		time.RFC3339Nano,        // RFC3339 with nanoseconds
	}
	
	var lastErr error
	for _, format := range formats {
		if t, err := time.Parse(format, timestamp); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	
	// Return the last error if all formats fail
	return time.Time{}, lastErr
}

// extractTask extracts the task description from the markdown content
func extractTask(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# Task: ") {
			return strings.TrimPrefix(line, "# Task: ")
		}
	}
	// Return empty string if no task heading found
	return ""
}

// UpdateTodo updates a specific section or metadata of a todo
func (tm *TodoManager) UpdateTodo(id, section, operation, content string, metadata map[string]string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	
	// Read existing todo file
	filePath := filepath.Join(tm.basePath, id + ".md")
	existingContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("todo not found: %s", id)
		}
		return fmt.Errorf("failed to read todo file: %w", err)
	}
	
	// Parse existing content
	contentStr := string(existingContent)
	parts := strings.Split(contentStr, "---\n")
	if len(parts) < 3 {
		return fmt.Errorf("invalid markdown format: missing frontmatter delimiters")
	}
	
	// Handle metadata updates
	if metadata != nil && len(metadata) > 0 {
		// Parse existing frontmatter
		var frontmatter map[string]interface{}
		err = yaml.Unmarshal([]byte(parts[1]), &frontmatter)
		if err != nil {
			return fmt.Errorf("failed to parse YAML frontmatter: %w", err)
		}
		
		// Update metadata fields
		for key, value := range metadata {
			frontmatter[key] = value
		}
		
		// Marshal back to YAML
		yamlData, err := yaml.Marshal(frontmatter)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		
		parts[1] = string(yamlData)
	}
	
	// Handle section updates
	if section != "" && operation != "" && content != "" {
		markdownContent := parts[2]
		
		// First, read the todo to get section metadata
		todo, err := tm.ReadTodo(id)
		if err != nil {
			return fmt.Errorf("failed to read todo: %w", err)
		}
		
		// Get section heading from metadata
		var sectionHeading string
		
		if todo.Sections != nil && len(todo.Sections) > 0 {
			// Use section metadata if available
			sectionDef, ok := todo.Sections[section]
			if !ok {
				return fmt.Errorf("invalid section: %s", section)
			}
			sectionHeading = sectionDef.Title
		} else {
			// Fall back to hardcoded map for backwards compatibility
			sectionMap := map[string]string{
				"findings":    "## Findings & Research",
				"tests":       "## Test Cases",
				"checklist":   "## Checklist",
				"scratchpad":  "## Working Scratchpad",
			}
			
			var ok bool
			sectionHeading, ok = sectionMap[section]
			if !ok {
				return fmt.Errorf("invalid section: %s", section)
			}
		}
		
		// Find section boundaries
		lines := strings.Split(markdownContent, "\n")
		sectionStart := -1
		sectionEnd := len(lines)
		
		for i, line := range lines {
			if line == sectionHeading {
				sectionStart = i
			} else if sectionStart > -1 && strings.HasPrefix(line, "## ") && i > sectionStart {
				sectionEnd = i
				break
			}
		}
		
		if sectionStart == -1 {
			return fmt.Errorf("section not found: %s", section)
		}
		
		// Extract section content
		sectionLines := lines[sectionStart+1:sectionEnd]
		sectionContent := strings.Join(sectionLines, "\n")
		
		// Apply operation
		switch operation {
		case "append":
			sectionContent = strings.TrimRight(sectionContent, "\n") + "\n" + content
		case "prepend":
			sectionContent = content + "\n" + strings.TrimLeft(sectionContent, "\n")
		case "replace":
			sectionContent = content
		case "toggle":
			// Toggle checklist item status
			if section == "checklist" || section == "test_list" {
				sectionContent = toggleChecklistItem(sectionContent, content)
			} else {
				return fmt.Errorf("toggle operation only supported for checklist sections")
			}
		default:
			return fmt.Errorf("invalid operation: %s", operation)
		}
		
		// Rebuild markdown content
		var newLines []string
		newLines = append(newLines, lines[:sectionStart+1]...)
		newLines = append(newLines, strings.Split(sectionContent, "\n")...)
		newLines = append(newLines, lines[sectionEnd:]...)
		markdownContent = strings.Join(newLines, "\n")
		
		parts[2] = markdownContent
	}
	
	// Reconstruct file content
	newContent := "---\n" + parts[1] + "---\n" + parts[2]
	
	// Write to temp file first (atomic write)
	tempFile := filePath + ".tmp"
	err = ioutil.WriteFile(tempFile, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Rename temp file to final location
	err = os.Rename(tempFile, filePath)
	if err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename file: %w", err)
	}
	
	return nil
}

// GetBasePath returns the base path for todos
func (tm *TodoManager) GetBasePath() string {
	return tm.basePath
}

// toggleChecklistItem toggles the status of a checklist item by its text
func toggleChecklistItem(content string, itemText string) string {
	lines := strings.Split(content, "\n")
	itemText = strings.TrimSpace(itemText)
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Check if this line is a checklist item
		if strings.HasPrefix(trimmed, "- [") && len(trimmed) > 5 {
			// Extract the text part (everything after the checkbox)
			text := strings.TrimSpace(trimmed[5:])
			
			// Check if this is the item we're looking for
			if text == itemText {
				// Determine current state and toggle
				if strings.HasPrefix(trimmed, "- [ ]") {
					// Pending -> In Progress
					lines[i] = strings.Replace(line, "- [ ]", "- [>]", 1)
				} else if strings.HasPrefix(trimmed, "- [>]") || strings.HasPrefix(trimmed, "- [-]") || strings.HasPrefix(trimmed, "- [~]") {
					// In Progress -> Completed
					lines[i] = strings.Replace(line, trimmed[:5], "- [x]", 1)
				} else if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
					// Completed -> Pending
					lines[i] = strings.Replace(line, trimmed[:5], "- [ ]", 1)
				}
				break
			}
		}
	}
	
	return strings.Join(lines, "\n")
}

// ParseChecklist extracts checklist items from markdown content with status
func ParseChecklist(content string) []ChecklistItem {
	var items []ChecklistItem
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Check for different checkbox states
		var status string
		var text string
		
		// Completed: - [x] or - [X]
		if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
			status = "completed"
			text = strings.TrimSpace(trimmed[5:]) // Skip "- [x]"
		} else if strings.HasPrefix(trimmed, "- [>]") || strings.HasPrefix(trimmed, "- [-]") || strings.HasPrefix(trimmed, "- [~]") {
			// In progress: - [>], - [-], or - [~]
			status = "in_progress"
			text = strings.TrimSpace(trimmed[5:]) // Skip "- [>]"
		} else if strings.HasPrefix(trimmed, "- [ ]") {
			// Pending: - [ ]
			status = "pending"
			text = strings.TrimSpace(trimmed[5:]) // Skip "- [ ]"
		} else {
			// Not a checklist item
			continue
		}
		
		if text != "" {
			items = append(items, ChecklistItem{
				Text:   text,
				Status: status,
			})
		}
	}
	
	return items
}

// SaveTodo writes the entire todo including sections to disk
func (tm *TodoManager) SaveTodo(todo *Todo) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	return tm.writeTodo(todo)
}

// ListTodos returns todos filtered by status, priority, and/or days
func (tm *TodoManager) ListTodos(status, priority string, days int) ([]*Todo, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	// Read all todo files
	files, err := ioutil.ReadDir(tm.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read todos directory: %w", err)
	}
	
	var todos []*Todo
	cutoffTime := time.Time{}
	if days > 0 {
		cutoffTime = time.Now().AddDate(0, 0, -days)
	}
	
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".md" {
			continue
		}
		
		// Read todo
		id := strings.TrimSuffix(file.Name(), ".md")
		todo, err := tm.ReadTodo(id)
		if err != nil {
			// Skip files that can't be parsed
			continue
		}
		
		// Apply filters
		if status != "" && status != "all" && todo.Status != status {
			continue
		}
		
		if priority != "" && priority != "all" && todo.Priority != priority {
			continue
		}
		
		if days > 0 && todo.Started.Before(cutoffTime) {
			continue
		}
		
		todos = append(todos, todo)
	}
	
	return todos, nil
}

// ReadTodoContent reads the full content of a todo file
func (tm *TodoManager) ReadTodoContent(id string) (string, error) {
	filePath := filepath.Join(tm.basePath, id+".md")
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read todo file: %w", err)
	}
	return string(content), nil
}

// ArchiveOldTodos archives todos older than specified days
func (tm *TodoManager) ArchiveOldTodos(days int) (int, error) {
	todos, err := tm.ListTodos("completed", "", days)
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, todo := range todos {
		if todo.Status == "completed" {
			err = tm.ArchiveTodo(todo.ID, "")
			if err == nil {
				count++
			}
		}
	}
	
	return count, nil
}

// FindDuplicateTodos finds todos with similar tasks
func (tm *TodoManager) FindDuplicateTodos() ([][]string, error) {
	todos, err := tm.ListTodos("", "", 0)
	if err != nil {
		return nil, err
	}
	
	// Group by normalized task
	groups := make(map[string][]string)
	for _, todo := range todos {
		// Normalize task for comparison
		normalized := strings.ToLower(strings.TrimSpace(todo.Task))
		groups[normalized] = append(groups[normalized], todo.ID)
	}
	
	// Find groups with duplicates
	var duplicates [][]string
	for _, group := range groups {
		if len(group) > 1 {
			duplicates = append(duplicates, group)
		}
	}
	
	return duplicates, nil
}
