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
	replacer := strings.NewReplacer(
		" ", "-",
		"_", "-",
		":", "",
		";", "",
		".", "",
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
	)
	
	kebab := replacer.Replace(lower)
	
	// Remove multiple consecutive hyphens
	for strings.Contains(kebab, "--") {
		kebab = strings.ReplaceAll(kebab, "--", "-")
	}
	
	// Trim hyphens from start and end
	kebab = strings.Trim(kebab, "-")
	
	// Limit length to make IDs manageable
	if len(kebab) > 50 {
		kebab = kebab[:50]
		// Ensure we don't cut off in the middle of a word
		lastHyphen := strings.LastIndex(kebab, "-")
		if lastHyphen > 30 {
			kebab = kebab[:lastHyphen]
		}
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
	
	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	
	// Build markdown content
	content := fmt.Sprintf(`---
%s---

# Task: %s

## Findings & Research

## Test Strategy

## Test List

## Test Cases

## Maintainability Analysis

## Test Results Log

## Checklist

## Working Scratchpad
`, string(yamlData), todo.Task)
	
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
	}
	
	// Parse timestamps
	if frontmatter.Started != "" {
		startTime, err := time.Parse("2006-01-02 15:04:05", frontmatter.Started)
		if err != nil {
			return nil, fmt.Errorf("failed to parse started timestamp: %w", err)
		}
		todo.Started = startTime
	}
	
	if frontmatter.Completed != "" {
		completedTime, err := time.Parse("2006-01-02 15:04:05", frontmatter.Completed)
		if err != nil {
			return nil, fmt.Errorf("failed to parse completed timestamp: %w", err)
		}
		todo.Completed = completedTime
	}
	
	// Extract task from markdown content
	markdownContent := parts[2]
	todo.Task = extractTask(markdownContent)
	
	return todo, nil
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