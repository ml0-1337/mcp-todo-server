package core

import (
	"github.com/user/mcp-todo-server/internal/search"
)

// TodoDocument is a type alias for backward compatibility
type TodoDocument = search.Document

// SearchResult represents a search result (kept in core for compatibility)
type SearchResult struct {
	ID      string
	Task    string
	Score   float64
	Snippet string
}
