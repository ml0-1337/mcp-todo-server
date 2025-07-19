package search

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	
	"github.com/user/mcp-todo-server/internal/logging"
)

// indexExistingTodosParallel indexes all existing todo files using parallel processing
func (e *Engine) indexExistingTodosParallel() error {
	totalStart := time.Now()
	
	// Read all .md files in basePath
	readStart := time.Now()
	logging.Infof("Reading todos directory: %s", e.basePath)
	files, err := ioutil.ReadDir(e.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No todos directory yet, that's OK
			logging.Infof("Todos directory doesn't exist yet, skipping indexing")
			return nil
		}
		return fmt.Errorf("failed to read todos directory: %w", err)
	}
	readTime := time.Since(readStart)
	logging.Infof("Found %d files in todos directory (readdir took %v)", len(files), readTime)

	// Filter markdown files
	var mdFiles []os.FileInfo
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") {
			mdFiles = append(mdFiles, file)
		}
	}
	logging.Timingf("Found %d markdown files out of %d total files", len(mdFiles), len(files))
	
	if len(mdFiles) == 0 {
		return nil
	}

	// Create channels for worker pool
	const numWorkers = 10
	workCh := make(chan os.FileInfo, len(mdFiles))
	resultCh := make(chan *indexResult, len(mdFiles))
	
	// Start workers
	var wg sync.WaitGroup
	processStart := time.Now()
	
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go e.indexWorker(i, workCh, resultCh, &wg)
	}
	
	// Send work to workers
	for _, file := range mdFiles {
		workCh <- file
	}
	close(workCh)
	
	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()
	
	// Collect results and create batch
	batch := e.index.NewBatch()
	processedCount := 0
	skippedCount := 0
	totalFileSize := int64(0)
	progressTicker := time.NewTicker(time.Second)
	defer progressTicker.Stop()
	
	done := false
	for !done {
		select {
		case result, ok := <-resultCh:
			if !ok {
				done = true
				break
			}
			
			if result.err != nil {
				skippedCount++
				if result.fileSize > 0 {
					fmt.Fprintf(os.Stderr, "[WARNING] Failed to process %s: %v\n", result.fileName, result.err)
				}
				continue
			}
			
			if result.doc != nil {
				batch.Index(result.doc.ID, result.doc)
				processedCount++
				totalFileSize += result.fileSize
			}
			
		case <-progressTicker.C:
			if processedCount > 0 {
				elapsed := time.Since(processStart)
				rate := float64(processedCount) / elapsed.Seconds()
				logging.Progressf("Indexed %d/%d files (%.1f files/sec)", 
					processedCount, len(mdFiles), rate)
			}
		}
	}
	
	processTime := time.Since(processStart)
	avgFileSize := int64(0)
	if processedCount > 0 {
		avgFileSize = totalFileSize / int64(processedCount)
	}
	logging.Timingf("Processed %d files in %v using %d workers (skipped %d, avg size: %d bytes)", 
		processedCount, processTime, numWorkers, skippedCount, avgFileSize)

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
	stopCh := make(chan struct{})
	defer close(stopCh)
	
	go func() {
		defer func() {
			// Recover from any panics in batch operation
			if r := recover(); r != nil {
				logging.Errorf("PANIC in batch indexing: %v", r)
				batchCh <- batchResult{err: fmt.Errorf("batch operation panicked: %v", r)}
			}
		}()
		
		select {
		case <-stopCh:
			// Context cancelled, exit gracefully
			return
		default:
			// Proceed with batch operation
			err := e.index.Batch(batch)
			
			// Only send result if context is still active
			select {
			case batchCh <- batchResult{err: err}:
			case <-stopCh:
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
	logging.Timingf("Total parallel indexing time: %v (readdir: %v, process: %v, batch: %v)",
		totalTime, readTime, processTime, time.Since(batchStart))
	logging.Performancef("Indexed %d todos at %.1f todos/sec", 
		processedCount, float64(processedCount)/totalTime.Seconds())

	return nil
}

// indexResult holds the result of indexing a single file
type indexResult struct {
	doc      *Document
	err      error
	fileName string
	fileSize int64
}

// indexWorker processes files from the work channel
func (e *Engine) indexWorker(id int, workCh <-chan os.FileInfo, resultCh chan<- *indexResult, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for file := range workCh {
		result := e.processFile(file)
		resultCh <- result
	}
}

// processFile processes a single todo file
func (e *Engine) processFile(file os.FileInfo) *indexResult {
	fileName := file.Name()
	todoID := strings.TrimSuffix(fileName, ".md")
	filePath := filepath.Join(e.basePath, fileName)
	
	// Read file with timeout
	fileCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	content, fileSize, err := e.readFileWithTimeout(fileCtx, filePath)
	if err != nil {
		return &indexResult{
			err:      fmt.Errorf("failed to read file: %w", err),
			fileName: fileName,
			fileSize: fileSize,
		}
	}
	
	// Parse todo to get structured data
	todo, err := parseTodoFile(todoID, string(content))
	if err != nil {
		return &indexResult{
			err:      fmt.Errorf("failed to parse todo: %w", err),
			fileName: fileName,
			fileSize: fileSize,
		}
	}
	
	// Create search document
	doc := &Document{
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
	
	return &indexResult{
		doc:      doc,
		fileName: fileName,
		fileSize: fileSize,
	}
}

// readFileWithTimeout reads a file with timeout protection
func (e *Engine) readFileWithTimeout(ctx context.Context, filePath string) ([]byte, int64, error) {
	type result struct {
		content []byte
		size    int64
		err     error
	}
	
	resultCh := make(chan result, 1)
	
	go func() {
		content, err := ioutil.ReadFile(filePath)
		resultCh <- result{
			content: content,
			size:    int64(len(content)),
			err:     err,
		}
	}()
	
	select {
	case res := <-resultCh:
		return res.content, res.size, res.err
	case <-ctx.Done():
		// Try to get file size even on timeout
		if info, err := os.Stat(filePath); err == nil {
			return nil, info.Size(), fmt.Errorf("file read timed out after 5 seconds")
		}
		return nil, 0, fmt.Errorf("file read timed out after 5 seconds")
	}
}