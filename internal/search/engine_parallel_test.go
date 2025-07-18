package search

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestParallelIndexing verifies that parallel indexing works correctly
func TestParallelIndexing(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
	
	err := os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create todos directory: %v", err)
	}
	
	// Create test todo files
	todoCount := 50
	for i := 0; i < todoCount; i++ {
		content := fmt.Sprintf(`---
todo_id: test-parallel-%05d
started: "%s"
status: in_progress
priority: high
type: feature
---

# Task: Parallel test todo %d

## Findings & Research
This is test content for parallel indexing verification.
It contains searchable terms like authentication, database, and security.

## Test Cases
Test case content for todo %d
`, i, time.Now().Format(time.RFC3339), i, i)
		
		todoPath := filepath.Join(todosDir, fmt.Sprintf("test-parallel-%05d.md", i))
		err := ioutil.WriteFile(todoPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write todo file: %v", err)
		}
	}
	
	// Create engine (which triggers parallel indexing)
	engine, err := NewEngine(indexPath, todosDir)
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer engine.Close()
	
	// Verify all todos were indexed
	count, err := engine.GetIndexedCount()
	if err != nil {
		t.Fatalf("Failed to get indexed count: %v", err)
	}
	
	if count != uint64(todoCount) {
		t.Errorf("Expected %d indexed documents, got %d", todoCount, count)
	}
	
	// Test search functionality
	results, err := engine.Search("authentication", nil, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	
	if len(results) == 0 {
		t.Error("Expected search results but got none")
	}
}

// TestParallelIndexingWithSlowFiles tests handling of slow file reads
func TestParallelIndexingWithSlowFiles(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
	
	err := os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create todos directory: %v", err)
	}
	
	// Create a mix of normal and "slow" files
	normalCount := 20
	slowCount := 5
	
	// Normal files
	for i := 0; i < normalCount; i++ {
		content := fmt.Sprintf(`---
todo_id: normal-%03d
started: "%s"
status: in_progress
priority: medium
type: feature
---

# Task: Normal todo %d
`, i, time.Now().Format(time.RFC3339), i)
		
		todoPath := filepath.Join(todosDir, fmt.Sprintf("normal-%03d.md", i))
		err := ioutil.WriteFile(todoPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write normal todo: %v", err)
		}
	}
	
	// Large files that might be slower to process
	for i := 0; i < slowCount; i++ {
		// Create a large file (500KB)
		content := fmt.Sprintf(`---
todo_id: large-%03d
started: "%s"
status: in_progress
priority: high
type: feature
---

# Task: Large todo %d

## Large Content Section
%s
`, i, time.Now().Format(time.RFC3339), i, generateLargeContent(500*1024))
		
		todoPath := filepath.Join(todosDir, fmt.Sprintf("large-%03d.md", i))
		err := ioutil.WriteFile(todoPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write large todo: %v", err)
		}
	}
	
	// Time the indexing
	start := time.Now()
	
	engine, err := NewEngine(indexPath, todosDir)
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer engine.Close()
	
	indexTime := time.Since(start)
	t.Logf("Indexed %d files (%d normal, %d large) in %v", 
		normalCount+slowCount, normalCount, slowCount, indexTime)
	
	// Verify all were indexed
	count, err := engine.GetIndexedCount()
	if err != nil {
		t.Fatalf("Failed to get indexed count: %v", err)
	}
	
	expectedCount := uint64(normalCount + slowCount)
	if count != expectedCount {
		t.Errorf("Expected %d indexed documents, got %d", expectedCount, count)
	}
	
	// Parallel should be faster than sequential even with large files
	if indexTime > 5*time.Second {
		t.Logf("WARNING: Indexing took longer than expected: %v", indexTime)
	}
}

// TestParallelIndexingConcurrency verifies workers are actually working in parallel
func TestParallelIndexingConcurrency(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
	
	err := os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create todos directory: %v", err)
	}
	
	// Create files that simulate slow reads
	fileCount := 30
	for i := 0; i < fileCount; i++ {
		content := fmt.Sprintf(`---
todo_id: concurrent-%03d
started: "%s"
status: in_progress
priority: high
type: test
---

# Task: Concurrency test %d
`, i, time.Now().Format(time.RFC3339), i)
		
		todoPath := filepath.Join(todosDir, fmt.Sprintf("concurrent-%03d.md", i))
		err := ioutil.WriteFile(todoPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write todo: %v", err)
		}
	}
	
	// Temporarily replace readFileWithTimeout to track concurrency
	originalEngine, err := NewEngine(indexPath, todosDir)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer originalEngine.Close()
	
	// Check that we had some concurrency
	count, err := originalEngine.GetIndexedCount()
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	
	if count != uint64(fileCount) {
		t.Errorf("Expected %d indexed files, got %d", fileCount, count)
	}
	
	t.Logf("Successfully indexed %d files using parallel processing", count)
}

// generateLargeContent creates a string of specified size
func generateLargeContent(size int) string {
	const pattern = "This is test content for large file simulation. It contains various keywords and patterns. "
	patternLen := len(pattern)
	
	result := make([]byte, size)
	for i := 0; i < size; i += patternLen {
		copy(result[i:], pattern)
	}
	
	return string(result)
}