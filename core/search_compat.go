package core

import (
	"github.com/user/mcp-todo-server/internal/domain"
	"github.com/user/mcp-todo-server/internal/search"
)

// SearchAdapter adapts the search Engine to implement the existing interface
type SearchAdapter struct {
	engine *search.Engine
}

// NewSearchEngine creates a new search engine (compatibility wrapper)
// This maintains backward compatibility while we migrate to the new structure
func NewSearchEngine(indexPath, todosPath string) (*SearchAdapter, error) {
	engine, err := search.NewEngine(indexPath, todosPath)
	if err != nil {
		return nil, err
	}
	return &SearchAdapter{engine: engine}, nil
}

// IndexTodo indexes a single todo (for updates)
func (a *SearchAdapter) IndexTodo(todo *Todo, content string) error {
	// Convert core.Todo to domain.Todo
	domainTodo := &domain.Todo{
		ID:        todo.ID,
		Task:      todo.Task,
		Started:   todo.Started,
		Completed: todo.Completed,
		Status:    todo.Status,
		Priority:  todo.Priority,
		Type:      todo.Type,
		ParentID:  todo.ParentID,
		Tags:      todo.Tags,
	}
	return a.engine.Index(domainTodo, content)
}

// DeleteTodo removes a todo from the index
func (a *SearchAdapter) DeleteTodo(id string) error {
	return a.engine.Delete(id)
}

// SearchTodos searches for todos matching the query
func (a *SearchAdapter) SearchTodos(queryStr string, filters map[string]string, limit int) ([]SearchResult, error) {
	results, err := a.engine.Search(queryStr, filters, limit)
	if err != nil {
		return nil, err
	}

	// Convert domain search results to core search results
	coreResults := make([]SearchResult, len(results))
	for i, r := range results {
		coreResults[i] = SearchResult{
			ID:      r.ID,
			Task:    r.Task,
			Score:   r.Score,
			Snippet: r.Snippet,
		}
	}

	return coreResults, nil
}

// Close closes the search index
func (a *SearchAdapter) Close() error {
	return a.engine.Close()
}

// GetIndexedCount returns the number of indexed documents
func (a *SearchAdapter) GetIndexedCount() (uint64, error) {
	return a.engine.GetIndexedCount()
}

// SearchEngine is a type alias for backward compatibility
type SearchEngine = SearchAdapter