package core

import (
	"testing"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
	"fmt"
	"strings"
)

// Test 11: Search index should be created on server start
func TestSearchIndexCreation(t *testing.T) {
	// Test 1: Create index on first run
	t.Run("Create index on first run", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-search-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create some test todos first
		manager := NewTodoManager(tempDir)
		todo1, err := manager.CreateTodo("Implement authentication", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo 1: %v", err)
		}
		
		_, err = manager.CreateTodo("Fix database connection", "medium", "bug")
		if err != nil {
			t.Fatalf("Failed to create todo 2: %v", err)
		}
		
		_, err = manager.CreateTodo("Write documentation", "low", "research")
		if err != nil {
			t.Fatalf("Failed to create todo 3: %v", err)
		}
		
		// Create search engine
		indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
		searchEngine, err := NewSearchEngine(indexPath, tempDir)
		if err != nil {
			t.Fatalf("Failed to create search engine: %v", err)
		}
		defer searchEngine.Close()
		
		// Verify index directory was created
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			t.Error("Index directory was not created")
		}
		
		// Verify todos were indexed
		count, err := searchEngine.GetIndexedCount()
		if err != nil {
			t.Fatalf("Failed to get indexed count: %v", err)
		}
		
		if count != 3 {
			t.Errorf("Expected 3 todos indexed, got %d", count)
		}
		
		// Verify each todo is searchable
		results, err := searchEngine.SearchTodos("authentication", nil, 10)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		
		if len(results) != 1 {
			t.Errorf("Expected 1 result for 'authentication', got %d", len(results))
		}
		
		if len(results) > 0 && results[0].ID != todo1.ID {
			t.Errorf("Expected result ID %s, got %s", todo1.ID, results[0].ID)
		}
	})
	
	// Test 2: Open existing index on subsequent runs
	t.Run("Open existing index on subsequent runs", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-search-reopen-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create some todos
		manager := NewTodoManager(tempDir)
		_, err = manager.CreateTodo("First todo", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}
		
		// Create first search engine instance
		indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
		searchEngine1, err := NewSearchEngine(indexPath, tempDir)
		if err != nil {
			t.Fatalf("Failed to create first search engine: %v", err)
		}
		
		// Close first instance
		searchEngine1.Close()
		
		// Create another todo while index is closed
		_, err = manager.CreateTodo("Second todo", "medium", "bug")
		if err != nil {
			t.Fatalf("Failed to create second todo: %v", err)
		}
		
		// Create second search engine instance - should open existing index
		searchEngine2, err := NewSearchEngine(indexPath, tempDir)
		if err != nil {
			t.Fatalf("Failed to create second search engine: %v", err)
		}
		defer searchEngine2.Close()
		
		// Should have indexed the new todo on startup
		count, err := searchEngine2.GetIndexedCount()
		if err != nil {
			t.Fatalf("Failed to get indexed count: %v", err)
		}
		
		if count != 2 {
			t.Errorf("Expected 2 todos indexed after reopening, got %d", count)
		}
	})
	
	// Test 3: Index all existing todos on creation
	t.Run("Index all existing todos on creation", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-batch-index-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create many todos before search engine
		manager := NewTodoManager(tempDir)
		todoCount := 50
		
		for i := 0; i < todoCount; i++ {
			task := fmt.Sprintf("Task number %d", i+1)
			priority := []string{"high", "medium", "low"}[i%3]
			todoType := []string{"feature", "bug", "refactor"}[i%3]
			
			_, err := manager.CreateTodo(task, priority, todoType)
			if err != nil {
				t.Fatalf("Failed to create todo %d: %v", i+1, err)
			}
		}
		
		// Measure indexing time
		startTime := time.Now()
		
		// Create search engine - should batch index all todos
		indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
		searchEngine, err := NewSearchEngine(indexPath, tempDir)
		if err != nil {
			t.Fatalf("Failed to create search engine: %v", err)
		}
		defer searchEngine.Close()
		
		indexingTime := time.Since(startTime)
		
		// Verify all todos were indexed
		count, err := searchEngine.GetIndexedCount()
		if err != nil {
			t.Fatalf("Failed to get indexed count: %v", err)
		}
		
		if count != uint64(todoCount) {
			t.Errorf("Expected %d todos indexed, got %d", todoCount, count)
		}
		
		// Performance check - should be fast
		if indexingTime > 5*time.Second {
			t.Errorf("Indexing %d todos took too long: %v", todoCount, indexingTime)
		}
		
		// Verify search works
		results, err := searchEngine.SearchTodos("Task number 25", nil, 10)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		
		if len(results) != 1 {
			t.Errorf("Expected 1 result for 'Task number 25', got %d", len(results))
		}
	})
	
	// Test 4: Handle corrupted index gracefully
	t.Run("Handle corrupted index gracefully", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-corrupt-index-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create index directory with corrupted data
		indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
		err = os.MkdirAll(indexPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create index directory: %v", err)
		}
		
		// Write garbage to index directory
		garbageFile := filepath.Join(indexPath, "store")
		err = ioutil.WriteFile(garbageFile, []byte("corrupted data"), 0644)
		if err != nil {
			t.Fatalf("Failed to write garbage file: %v", err)
		}
		
		// Create some todos
		manager := NewTodoManager(tempDir)
		_, err = manager.CreateTodo("Test todo", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}
		
		// Try to create search engine - should handle corruption
		searchEngine, err := NewSearchEngine(indexPath, tempDir)
		if err != nil {
			// It's OK if it fails, but should give meaningful error
			if !strings.Contains(err.Error(), "index") && !strings.Contains(err.Error(), "corrupt") {
				t.Errorf("Error should mention index corruption, got: %v", err)
			}
			return
		}
		defer searchEngine.Close()
		
		// If it succeeded by recreating index, verify it works
		count, err := searchEngine.GetIndexedCount()
		if err != nil {
			t.Fatalf("Failed to get indexed count: %v", err)
		}
		
		if count < 1 {
			t.Error("Should have reindexed existing todos after corruption")
		}
	})
}