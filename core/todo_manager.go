package core

import (
	"sync"
)

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

// GetBasePath returns the base path for todo storage
func (tm *TodoManager) GetBasePath() string {
	return tm.basePath
}