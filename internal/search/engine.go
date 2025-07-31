// Package search provides full-text search capabilities for todos using the Bleve search engine.
// It includes parallel indexing, query parsing, and result highlighting.
package search

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/user/mcp-todo-server/internal/domain"
	domainSearch "github.com/user/mcp-todo-server/internal/domain/search"
	"github.com/user/mcp-todo-server/internal/logging"
)

// Engine manages the bleve search index
type Engine struct {
	index          bleve.Index
	basePath       string
	lock           *IndexLock
	circuitBreaker *CircuitBreaker
}

// NewEngine creates or opens a search index
func NewEngine(indexPath, todosPath string) (*Engine, error) {
	startTime := time.Now()

	// Create index lock to prevent concurrent access
	lockStart := time.Now()
	indexLock := NewIndexLock(indexPath)

	// Try to acquire lock with timeout
	err := indexLock.TryLock(10 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire index lock (index may be in use by another process): %w", err)
	}
	lockTime := time.Since(lockStart)
	logging.Timingf("Index lock acquisition took %v", lockTime)

	// Check if index exists
	openStart := time.Now()
	index, err := bleve.Open(indexPath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		// Create new index
		logging.Timingf("Index does not exist, creating new index at %s", indexPath)
		mapping := buildIndexMapping()
		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
	} else if err != nil {
		// Try to handle corruption by recreating
		logging.Timingf("Index corrupted, recreating at %s", indexPath)
		os.RemoveAll(indexPath)
		mapping := buildIndexMapping()
		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to recreate corrupted index: %w", err)
		}
	}
	openTime := time.Since(openStart)
	logging.Timingf("Index open/create took %v", openTime)

	engine := &Engine{
		index:          index,
		basePath:       todosPath,
		lock:           indexLock,
		circuitBreaker: NewCircuitBreaker(3, 15*time.Second, 30*time.Second),
	}

	// Index existing todos with timeout
	indexingStart := time.Now()
	logging.Infof("Indexing existing todos from %s...", todosPath)
	indexCtx, indexCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer indexCancel()

	err = engine.indexExistingTodosWithTimeout(indexCtx)
	indexingTime := time.Since(indexingStart)

	if err != nil {
		logging.Warnf("Failed to index existing todos after %v: %v. Search may have incomplete results.", indexingTime, err)
		// Don't fail engine creation - continue with empty index
	} else {
		logging.Infof("Finished indexing existing todos in %v", indexingTime)
	}

	totalTime := time.Since(startTime)
	logging.Timingf("Total NewEngine time: %v (lock: %v, open: %v, indexing: %v)",
		totalTime, lockTime, openTime, indexingTime)

	return engine, nil
}

// indexExistingTodosWithTimeout indexes all existing todo files with timeout protection
func (e *Engine) indexExistingTodosWithTimeout(ctx context.Context) error {
	// Channel to receive result from goroutine
	type result struct {
		err error
	}
	resultCh := make(chan result, 1)

	// Run indexing in goroutine to enable timeout
	go func() {
		// Use parallel indexing for better performance
		err := e.indexExistingTodosParallel()
		resultCh <- result{err: err}
	}()

	// Wait for either completion or timeout
	select {
	case res := <-resultCh:
		return res.err
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("indexing existing todos timed out after 60 seconds")
		}
		return ctx.Err()
	}
}

// indexExistingTodos indexes all existing todo files
func (e *Engine) indexExistingTodos() error {
	totalStart := time.Now()

	// Check if todos directory exists
	logging.Infof("Starting recursive index of todos directory: %s", e.basePath)
	if _, err := os.Stat(e.basePath); os.IsNotExist(err) {
		logging.Infof("Todos directory doesn't exist yet, skipping indexing")
		return nil
	}

	// Create a batch for efficient indexing
	batch := e.index.NewBatch()

	// Process files recursively
	processStart := time.Now()
	processedCount := 0
	skippedCount := 0
	totalFileSize := int64(0)

	// Walk the directory tree
	err := filepath.Walk(e.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logging.Warnf("Error accessing path %s: %v", path, err)
			return nil // Continue walking
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Log progress every 10 files
		if processedCount > 0 && processedCount%10 == 0 {
			elapsed := time.Since(processStart)
			rate := float64(processedCount) / elapsed.Seconds()
			logging.Progressf("Indexed %d files (%.1f files/sec)",
				processedCount, rate)
		}

		// Extract todo ID from filename
		todoID := strings.TrimSuffix(info.Name(), ".md")

		fileStart := time.Now()
		content, err := os.ReadFile(path)
		if err != nil {
			skippedCount++
			logging.Warnf("Failed to read file %s: %v", path, err)
			return nil // Skip files we can't read
		}
		totalFileSize += int64(len(content))

		// Parse todo to get structured data
		todo, err := parseTodoFile(todoID, string(content))
		if err != nil {
			skippedCount++
			logging.Warnf("Failed to parse todo %s: %v", todoID, err)
			return nil // Skip malformed files
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
		processedCount++

		// Log slow files
		fileTime := time.Since(fileStart)
		if fileTime > 100*time.Millisecond {
			logging.Timingf("Slow file %s: %v (size: %d bytes)",
				info.Name(), fileTime, len(content))
		}

		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error walking directory: %w", err)
	}

	processTime := time.Since(processStart)
	var avgFileSize int64
	if processedCount > 0 {
		avgFileSize = totalFileSize / int64(processedCount)
	}
	logging.Timingf("Processed %d files in %v (skipped %d, avg size: %d bytes)",
		processedCount, processTime, skippedCount, avgFileSize)

	// Execute batch with timeout protection
	batchStart := time.Now()
	batchCtx, batchCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer batchCancel()

	// Channel to receive batch result
	type batchResult struct {
		err error
	}
	batchCh := make(chan batchResult, 1)

	// Channel to signal goroutine to stop
	done := make(chan struct{})
	defer close(done)

	go func() {
		defer func() {
			// Recover from any panics in batch operation
			if r := recover(); r != nil {
				logging.Errorf("PANIC in batch indexing: %v", r)
				batchCh <- batchResult{err: fmt.Errorf("batch operation panicked: %v", r)}
			}
		}()

		select {
		case <-done:
			// Context cancelled, exit gracefully
			return
		default:
			// Proceed with batch operation
			err := e.index.Batch(batch)

			// Only send result if context is still active
			select {
			case batchCh <- batchResult{err: err}:
			case <-done:
				// Context cancelled while trying to send result
				return
			}
		}
	}()

	select {
	case res := <-batchCh:
		batchTime := time.Since(batchStart)
		if res.err != nil {
			return fmt.Errorf("failed to index batch after %v: %w", batchTime, res.err)
		}
		logging.Timingf("Batch commit took %v", batchTime)
	case <-batchCtx.Done():
		return fmt.Errorf("batch indexing timed out after 10 seconds")
	}

	totalTime := time.Since(totalStart)
	logging.Timingf("Total indexExistingTodos time: %v (process: %v, batch: %v)",
		totalTime, processTime, time.Since(batchStart))

	return nil
}

// Close closes the search index
func (e *Engine) Close() error {
	defer func() {
		if e.lock != nil {
			if err := e.lock.Unlock(); err != nil {
				logging.Warnf("Failed to unlock index: %v", err)
			}
		}
	}()
	return e.index.Close()
}

// HealthCheck returns the health status of the search engine
// HealthCheck is now implemented in engine_health.go

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

	// Execute search with circuit breaker protection
	var searchResults *bleve.SearchResult
	err := e.circuitBreaker.Execute(context.Background(), func() error {
		var searchErr error
		searchResults, searchErr = e.index.Search(searchRequest)
		return searchErr
	})
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

	return e.circuitBreaker.Execute(context.Background(), func() error {
		return e.index.Index(todo.ID, doc)
	})
}

// Delete removes a todo from the index
func (e *Engine) Delete(id string) error {
	return e.circuitBreaker.Execute(context.Background(), func() error {
		return e.index.Delete(id)
	})
}
