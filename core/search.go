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
	// Build query - use match phrase for exact matching
	q := bleve.NewMatchPhraseQuery(queryStr)
	
	// Create search request
	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.Size = limit
	searchRequest.Fields = []string{"task", "id"}
	
	// Execute search
	searchResults, err := se.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	
	// Convert results
	var results []SearchResult
	for _, hit := range searchResults.Hits {
		result := SearchResult{
			ID:    hit.ID,
			Score: hit.Score,
		}
		
		// Get task from stored fields
		if task, ok := hit.Fields["task"].(string); ok {
			result.Task = task
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