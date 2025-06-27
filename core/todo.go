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
	// Convert to lowercase
	lower := strings.ToLower(task)
	
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