package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
)

// TodoDocument represents a todo in the search index
type TodoDocument struct {
	ID        string    `json:"id"`
	Task      string    `json:"task"`
	Status    string    `json:"status"`
	Priority  string    `json:"priority"`
	Type      string    `json:"type"`
	Started   time.Time `json:"started"`
	Completed time.Time `json:"completed,omitempty"`
	Content   string    `json:"content"`
	Findings  string    `json:"findings"`
	Tests     string    `json:"tests"`
}

// SearchResult represents a search result
type SearchResult struct {
	ID      string
	Task    string
	Score   float64
	Snippet string
}

// SearchEngine manages the bleve search index
type SearchEngine struct {
	index    bleve.Index
	basePath string
}

// NewSearchEngine creates or opens a search index
func NewSearchEngine(indexPath, todosPath string) (*SearchEngine, error) {
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

	engine := &SearchEngine{
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

// buildIndexMapping creates the index mapping for todos
func buildIndexMapping() mapping.IndexMapping {
	// Create a standard document mapping
	todoMapping := bleve.NewDocumentMapping()

	// Task field - boosted for relevance
	taskFieldMapping := bleve.NewTextFieldMapping()
	taskFieldMapping.Analyzer = standard.Name
	taskFieldMapping.Store = true
	taskFieldMapping.IncludeInAll = true
	todoMapping.AddFieldMappingsAt("task", taskFieldMapping)

	// Content fields
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = standard.Name
	textFieldMapping.Store = true
	textFieldMapping.IncludeInAll = true

	todoMapping.AddFieldMappingsAt("content", textFieldMapping)
	todoMapping.AddFieldMappingsAt("findings", textFieldMapping)
	todoMapping.AddFieldMappingsAt("tests", textFieldMapping)

	// Metadata fields
	keywordFieldMapping := bleve.NewKeywordFieldMapping()
	keywordFieldMapping.Store = true

	todoMapping.AddFieldMappingsAt("id", keywordFieldMapping)
	todoMapping.AddFieldMappingsAt("status", keywordFieldMapping)
	todoMapping.AddFieldMappingsAt("priority", keywordFieldMapping)
	todoMapping.AddFieldMappingsAt("type", keywordFieldMapping)

	// Date fields
	dateFieldMapping := bleve.NewDateTimeFieldMapping()
	dateFieldMapping.Store = true
	dateFieldMapping.Index = true
	// Specify that we're using RFC3339 format for dates
	dateFieldMapping.DateFormat = "dateTimeOptional"

	todoMapping.AddFieldMappingsAt("started", dateFieldMapping)
	todoMapping.AddFieldMappingsAt("completed", dateFieldMapping)

	// Create index mapping
	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("todo", todoMapping)
	indexMapping.DefaultMapping = todoMapping

	return indexMapping
}

// indexExistingTodos indexes all existing todo files
func (se *SearchEngine) indexExistingTodos() error {
	// Read all .md files in basePath
	files, err := ioutil.ReadDir(se.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No todos directory yet, that's OK
			return nil
		}
		return fmt.Errorf("failed to read todos directory: %w", err)
	}

	// Create a batch for efficient indexing
	batch := se.index.NewBatch()

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		// Read and parse todo file
		todoID := strings.TrimSuffix(file.Name(), ".md")
		filePath := filepath.Join(se.basePath, file.Name())

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			continue // Skip files we can't read
		}

		// Parse todo to get structured data
		manager := &TodoManager{basePath: se.basePath}
		todo, err := manager.parseTodoFile(string(content))
		if err != nil {
			continue // Skip malformed files
		}

		// Create search document
		doc := TodoDocument{
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
	err = se.index.Batch(batch)
	if err != nil {
		return fmt.Errorf("failed to index batch: %w", err)
	}

	return nil
}

// extractSection extracts content under a specific heading
func extractSection(content, heading string) string {
	lines := strings.Split(content, "\n")
	inSection := false
	var sectionLines []string

	for _, line := range lines {
		if line == heading {
			inSection = true
			continue
		}

		if inSection && strings.HasPrefix(line, "## ") {
			break
		}

		if inSection {
			sectionLines = append(sectionLines, line)
		}
	}

	return strings.Join(sectionLines, "\n")
}

// Close closes the search index
func (se *SearchEngine) Close() error {
	return se.index.Close()
}

// GetIndexedCount returns the number of indexed documents
func (se *SearchEngine) GetIndexedCount() (uint64, error) {
	count, err := se.index.DocCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get document count: %w", err)
	}
	return count, nil
}

// SearchTodos searches for todos matching the query
func (se *SearchEngine) SearchTodos(queryStr string, filters map[string]string, limit int) ([]SearchResult, error) {
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
			sanitized := sanitizeSearchQuery(queryStr)

			// If sanitization removed all content, return empty results
			if sanitized == "" {
				return []SearchResult{}, nil
			}

			// Debug: log sanitized query
			// fmt.Printf("DEBUG SearchTodos - Original: %q, Sanitized: %q\n", queryStr, sanitized)

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
	searchResults, err := se.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results - date filtering now handled by bleve
	var results []SearchResult

	// Post-filtering disabled - using bleve date range queries
	for _, hit := range searchResults.Hits {

		result := SearchResult{
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

// IndexTodo indexes a single todo (for updates)
func (se *SearchEngine) IndexTodo(todo *Todo, content string) error {
	doc := TodoDocument{
		ID:        todo.ID,
		Task:      todo.Task,
		Status:    todo.Status,
		Priority:  todo.Priority,
		Type:      todo.Type,
		Started:   todo.Started,
		Completed: todo.Completed,
		Content:   content,
	}

	return se.index.Index(todo.ID, doc)
}

// DeleteTodo removes a todo from the index
func (se *SearchEngine) DeleteTodo(id string) error {
	return se.index.Delete(id)
}

// sanitizeSearchQuery cleans up the search query to handle special characters safely
// This prevents regex injection attacks and ensures queries don't break the search engine
func sanitizeSearchQuery(query string) string {
	// First, handle potential regex patterns - remove forward slashes that would indicate regex
	if strings.HasPrefix(query, "/") && strings.HasSuffix(query, "/") {
		// Don't execute as regex - treat as literal by removing slashes
		query = query[1 : len(query)-1]
	}

	// Remove null bytes and other control characters
	query = strings.ReplaceAll(query, "\x00", " ") // Replace null with space instead of removing
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\r", " ")
	query = strings.ReplaceAll(query, "\t", " ")

	// Replace special characters that break queries with spaces
	// Keep alphanumeric, spaces, and some safe punctuation
	var result []rune
	for _, r := range query {
		switch {
		case r >= 'a' && r <= 'z':
			result = append(result, r)
		case r >= 'A' && r <= 'Z':
			result = append(result, r)
		case r >= '0' && r <= '9':
			result = append(result, r)
		case r == ' ':
			result = append(result, r)
		case r == '-' || r == '_':
			// Keep hyphens and underscores as they're common in identifiers
			result = append(result, r)
		default:
			// Replace other special characters with space
			result = append(result, ' ')
		}
	}

	// Convert back to string and clean up extra spaces
	cleaned := string(result)

	// Replace multiple consecutive spaces with single space
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}

	// Trim leading and trailing spaces
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}
