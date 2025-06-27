package core

import (
	"fmt"
	"strings"
	"time"
	"sync"
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