package core

import (
	"fmt"
	"path/filepath"
	"os"
	"io/ioutil"
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
	ID       string
	Task     string
	Score    float64
	Snippet  string
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
			// For non-phrase queries, use field-specific queries with boosting
			// Create separate queries for each field with different boost values
			taskQuery := bleve.NewMatchQuery(queryStr)
			taskQuery.SetField("task")
			taskQuery.SetBoost(3.0) // Highest priority for task/title matches
			
			findingsQuery := bleve.NewMatchQuery(queryStr)
			findingsQuery.SetField("findings")
			findingsQuery.SetBoost(1.5) // Medium priority for findings
			
			testsQuery := bleve.NewMatchQuery(queryStr)
			testsQuery.SetField("tests")
			testsQuery.SetBoost(1.0) // Lower priority for test cases
			
			contentQuery := bleve.NewMatchQuery(queryStr)
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
		
		// Date range filter - commented out for now, will do post-filtering
		// TODO: Fix bleve date range query
		/*
		if dateFrom, ok := filters["date_from"]; ok && dateFrom != "" {
			// Parse the dates in UTC
			fromTime, err := time.ParseInLocation("2006-01-02", dateFrom, time.UTC)
			if err != nil {
				return nil, fmt.Errorf("invalid date_from format: %w", err)
			}
			
			var toTime *time.Time
			if dateTo, ok := filters["date_to"]; ok && dateTo != "" {
				t, err := time.ParseInLocation("2006-01-02", dateTo, time.UTC)
				if err != nil {
					return nil, fmt.Errorf("invalid date_to format: %w", err)
				}
				// Set to end of day
				t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
				toTime = &t
			}
			
			// Create date range query for started field
			var dateRangeQuery *query.DateRangeQuery
			if toTime != nil {
				dateRangeQuery = bleve.NewDateRangeQuery(fromTime, *toTime)
			} else {
				// No end date means from date to now
				now := time.Now().UTC()
				dateRangeQuery = bleve.NewDateRangeQuery(fromTime, now)
				}
			dateRangeQuery.SetField("started")
			queries = append(queries, dateRangeQuery)
		}
		*/
		
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
	
	// Convert results and apply date filtering
	var results []SearchResult
	
	// Parse date filters for post-filtering
	var dateFrom, dateTo *time.Time
	if df, ok := filters["date_from"]; ok && df != "" {
		t, err := time.ParseInLocation("2006-01-02", df, time.UTC)
		if err == nil {
			dateFrom = &t
		}
	}
	if dt, ok := filters["date_to"]; ok && dt != "" {
		t, err := time.ParseInLocation("2006-01-02", dt, time.UTC)
		if err == nil {
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			dateTo = &t
		}
	}
	
	for _, hit := range searchResults.Hits {
		// If we have date filters, check the started date
		if dateFrom != nil || dateTo != nil {
			// Read the todo to get the actual started date
			todoPath := filepath.Join(se.basePath, hit.ID+".md")
			content, err := ioutil.ReadFile(todoPath)
			if err != nil {
				continue
			}
			
			// Parse the todo
			manager := &TodoManager{basePath: se.basePath}
			todo, err := manager.parseTodoFile(string(content))
			if err != nil {
				continue
			}
			
			// Check date range
			if dateFrom != nil && todo.Started.Before(*dateFrom) {
				continue
			}
			if dateTo != nil && todo.Started.After(*dateTo) {
				continue
			}
		}
		
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