package search

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/user/mcp-todo-server/internal/domain"
	domainSearch "github.com/user/mcp-todo-server/internal/domain/search"
)

// Engine manages the bleve search index
type Engine struct {
	index    bleve.Index
	basePath string
}

// NewEngine creates or opens a search index
func NewEngine(indexPath, todosPath string) (*Engine, error) {
	// Check if index exists
	index, err := bleve.Open(indexPath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		// Create new index
		mapping := buildIndexMapping()
		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
	} else if err != nil {
		// Try to handle corruption by recreating
		os.RemoveAll(indexPath)
		mapping := buildIndexMapping()
		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to recreate corrupted index: %w", err)
		}
	}

	engine := &Engine{
		index:    index,
		basePath: todosPath,
	}

	// Index existing todos
	err = engine.indexExistingTodos()
	if err != nil {
		return nil, fmt.Errorf("failed to index existing todos: %w", err)
	}

	return engine, nil
}

// indexExistingTodos indexes all existing todo files
func (e *Engine) indexExistingTodos() error {
	// Read all .md files in basePath
	files, err := ioutil.ReadDir(e.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No todos directory yet, that's OK
			return nil
		}
		return fmt.Errorf("failed to read todos directory: %w", err)
	}

	// Create a batch for efficient indexing
	batch := e.index.NewBatch()

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		// Read and parse todo file
		todoID := strings.TrimSuffix(file.Name(), ".md")
		filePath := filepath.Join(e.basePath, file.Name())

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			continue // Skip files we can't read
		}

		// Parse todo to get structured data
		todo, err := parseTodoFile(todoID, string(content))
		if err != nil {
			continue // Skip malformed files
		}

		// Create search document
		doc := Document{
			ID:        todoID,
			Task:      todo.Task,
			Status:    todo.Status,
			Priority:  todo.Priority,
			Type:      todo.Type,
			Started:   todo.Started,
			Completed: todo.Completed,
			Content:   string(content),
		}

		// Extract sections for better search
		doc.Findings = extractSection(string(content), "## Findings & Research")
		doc.Tests = extractSection(string(content), "## Test Cases")

		// Add to batch
		batch.Index(todoID, doc)
	}

	// Execute batch
	err = e.index.Batch(batch)
	if err != nil {
		return fmt.Errorf("failed to index batch: %w", err)
	}

	return nil
}

// Close closes the search index
func (e *Engine) Close() error {
	return e.index.Close()
}

// GetIndexedCount returns the number of indexed documents
func (e *Engine) GetIndexedCount() (uint64, error) {
	count, err := e.index.DocCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get document count: %w", err)
	}
	return count, nil
}

// Search performs a search with the given query and filters
func (e *Engine) Search(queryStr string, filters map[string]string, limit int) ([]domainSearch.Result, error) {
	// Build composite query
	var searchQuery query.Query

	// Handle text query
	if queryStr != "" {
		// Check if it's a phrase query (quoted)
		if strings.HasPrefix(queryStr, "\"") && strings.HasSuffix(queryStr, "\"") {
			// Keep phrase queries as-is for exact matching
			searchQuery = bleve.NewQueryStringQuery(queryStr)
		} else {
			// Sanitize query to handle special characters safely
			sanitized := sanitizeQuery(queryStr)

			// If sanitization removed all content, return empty results
			if sanitized == "" {
				return []domainSearch.Result{}, nil
			}

			// For non-phrase queries, use field-specific queries with boosting
			// Create separate queries for each field with different boost values
			taskQuery := bleve.NewMatchQuery(sanitized)
			taskQuery.SetField("task")
			taskQuery.SetBoost(3.0) // Highest priority for task/title matches

			findingsQuery := bleve.NewMatchQuery(sanitized)
			findingsQuery.SetField("findings")
			findingsQuery.SetBoost(1.5) // Medium priority for findings

			testsQuery := bleve.NewMatchQuery(sanitized)
			testsQuery.SetField("tests")
			testsQuery.SetBoost(1.0) // Lower priority for test cases

			contentQuery := bleve.NewMatchQuery(sanitized)
			contentQuery.SetField("content")
			contentQuery.SetBoost(0.5) // Lowest priority for full content

			// Use DisjunctionQuery for "best match" behavior
			searchQuery = bleve.NewDisjunctionQuery(taskQuery, findingsQuery, testsQuery, contentQuery)
		}
	} else {
		// If no query string, match all documents
		searchQuery = bleve.NewMatchAllQuery()
	}

	// Apply filters if provided
	if len(filters) > 0 {
		var queries []query.Query
		queries = append(queries, searchQuery)

		// Status filter
		if status, ok := filters["status"]; ok && status != "" {
			statusQuery := bleve.NewTermQuery(status)
			statusQuery.SetField("status")
			queries = append(queries, statusQuery)
		}

		// Date range filter using bleve native support
		var fromTime, toTime *time.Time
		
		// Parse date_from if provided
		if dateFrom, ok := filters["date_from"]; ok && dateFrom != "" {
			// Parse as start of day in UTC
			t, err := time.ParseInLocation("2006-01-02", dateFrom, time.UTC)
			if err != nil {
				return nil, fmt.Errorf("invalid date_from format: %w", err)
			}
			fromTime = &t
		}
		
		// Parse date_to if provided
		if dateTo, ok := filters["date_to"]; ok && dateTo != "" {
			t, err := time.ParseInLocation("2006-01-02", dateTo, time.UTC)
			if err != nil {
				return nil, fmt.Errorf("invalid date_to format: %w", err)
			}
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			toTime = &t
		}
		
		// Create date range query if we have at least one date
		if fromTime != nil || toTime != nil {
			var dateRangeQuery query.Query
			trueVal := true
			
			if fromTime != nil && toTime != nil {
				// Both dates provided
				dateRangeQuery = bleve.NewDateRangeInclusiveQuery(*fromTime, *toTime, &trueVal, &trueVal)
			} else if fromTime != nil {
				// Only start date - search from date to now
				now := time.Now().UTC()
				dateRangeQuery = bleve.NewDateRangeInclusiveQuery(*fromTime, now, &trueVal, &trueVal)
			} else {
				// Only end date - search from beginning of time to date
				// Use a very old date as the start
				veryOldDate := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
				dateRangeQuery = bleve.NewDateRangeInclusiveQuery(veryOldDate, *toTime, &trueVal, &trueVal)
			}
			
			dateRangeQuery.(*query.DateRangeQuery).SetField("started")
			queries = append(queries, dateRangeQuery)
		}

		// Combine all queries with AND
		searchQuery = bleve.NewConjunctionQuery(queries...)
	}

	// Create search request
	searchRequest := bleve.NewSearchRequest(searchQuery)
	searchRequest.Size = limit
	searchRequest.Fields = []string{"task", "id", "started", "status"}
	searchRequest.Highlight = bleve.NewHighlight() // Enable snippets

	// Execute search
	searchResults, err := e.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results
	var results []domainSearch.Result

	for _, hit := range searchResults.Hits {
		result := domainSearch.Result{
			ID:    hit.ID,
			Score: hit.Score,
		}

		// Get task from stored fields
		if task, ok := hit.Fields["task"].(string); ok {
			result.Task = task
		}

		// Get snippet from highlights
		if len(hit.Fragments) > 0 {
			for _, fragments := range hit.Fragments {
				if len(fragments) > 0 {
					result.Snippet = fragments[0]
					break
				}
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// Index adds or updates a todo in the search index
func (e *Engine) Index(todo *domain.Todo, content string) error {
	doc := Document{
		ID:        todo.ID,
		Task:      todo.Task,
		Status:    todo.Status,
		Priority:  todo.Priority,
		Type:      todo.Type,
		Started:   todo.Started,
		Completed: todo.Completed,
		Content:   content,
	}

	// Extract sections for better search
	doc.Findings = extractSection(content, "## Findings & Research")
	doc.Tests = extractSection(content, "## Test Cases")

	return e.index.Index(todo.ID, doc)
}

// Delete removes a todo from the index
func (e *Engine) Delete(id string) error {
	return e.index.Delete(id)
}