package search

import (
	"github.com/user/mcp-todo-server/internal/domain"
)

// Result represents a search result
type Result struct {
	ID      string
	Task    string
	Score   float64
	Snippet string
}

// Repository defines the interface for search operations
type Repository interface {
	// Index adds or updates a todo in the search index
	Index(todo *domain.Todo, content string) error
	
	// Delete removes a todo from the search index
	Delete(id string) error
	
	// Search performs a search with the given query and filters
	Search(query string, filters map[string]string, limit int) ([]Result, error)
	
	// GetIndexedCount returns the number of indexed documents
	GetIndexedCount() (uint64, error)
	
	// Close closes the search index
	Close() error
}