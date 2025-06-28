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
		
		// Verify search works - use quotes for exact phrase
		results, err := searchEngine.SearchTodos(`"Task number 25"`, nil, 10)
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

// Test 12: todo_search should find todos by content keywords
func TestSearchTodosByKeywords(t *testing.T) {
	// Test 1: Basic keyword search
	t.Run("Find todos by single keyword", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-keyword-search-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create test todos with different content
		manager := NewTodoManager(tempDir)
		
		todo1, err := manager.CreateTodo("Implement authentication system", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo 1: %v", err)
		}
		
		// Update todo1 with findings about JWT
		err = manager.UpdateTodo(todo1.ID, "findings", "append", "\n\nNeed to implement JWT token generation and validation.", nil)
		if err != nil {
			t.Fatalf("Failed to update todo 1: %v", err)
		}
		
		_, err = manager.CreateTodo("Fix database connection pooling", "medium", "bug")
		if err != nil {
			t.Fatalf("Failed to create todo 2: %v", err)
		}
		
		_, err = manager.CreateTodo("Write API documentation", "low", "documentation")
		if err != nil {
			t.Fatalf("Failed to create todo 3: %v", err)
		}
		
		// Create search engine and index
		indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
		searchEngine, err := NewSearchEngine(indexPath, tempDir)
		if err != nil {
			t.Fatalf("Failed to create search engine: %v", err)
		}
		defer searchEngine.Close()
		
		// Search for "authentication"
		results, err := searchEngine.SearchTodos("authentication", nil, 10)
		if err != nil {
			t.Fatalf("Failed to search for 'authentication': %v", err)
		}
		
		if len(results) != 1 {
			t.Errorf("Expected 1 result for 'authentication', got %d", len(results))
		}
		
		if len(results) > 0 && results[0].ID != todo1.ID {
			t.Errorf("Expected todo ID %s, got %s", todo1.ID, results[0].ID)
		}
		
		// Verify result has score
		if len(results) > 0 && results[0].Score == 0 {
			t.Error("Search result should have non-zero relevance score")
		}
	})
	
	// Test 2: Search in findings section
	t.Run("Find todos by content in findings", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-findings-search-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create todos
		manager := NewTodoManager(tempDir)
		
		todo1, err := manager.CreateTodo("Research task", "high", "research")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}
		
		// Add unique content to findings
		err = manager.UpdateTodo(todo1.ID, "findings", "append", "\n\nDiscovered that WebSocket connections provide better real-time performance.", nil)
		if err != nil {
			t.Fatalf("Failed to update findings: %v", err)
		}
		
		// Create and index
		indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
		searchEngine, err := NewSearchEngine(indexPath, tempDir)
		if err != nil {
			t.Fatalf("Failed to create search engine: %v", err)
		}
		defer searchEngine.Close()
		
		// Search for content in findings
		results, err := searchEngine.SearchTodos("WebSocket", nil, 10)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		
		if len(results) != 1 {
			t.Errorf("Expected 1 result for 'WebSocket', got %d", len(results))
		}
		
		if len(results) > 0 && results[0].ID != todo1.ID {
			t.Errorf("Expected todo ID %s, got %s", todo1.ID, results[0].ID)
		}
	})
	
	// Test 3: Multi-word search
	t.Run("Find todos by multi-word query", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-multiword-search-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create todos
		manager := NewTodoManager(tempDir)
		
		todo1, err := manager.CreateTodo("Implement user authentication with OAuth", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo 1: %v", err)
		}
		
		_, err = manager.CreateTodo("Add social media login", "medium", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo 2: %v", err)
		}
		
		_, err = manager.CreateTodo("Fix login page styling", "low", "bug")
		if err != nil {
			t.Fatalf("Failed to create todo 3: %v", err)
		}
		
		// Create and index
		indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
		searchEngine, err := NewSearchEngine(indexPath, tempDir)
		if err != nil {
			t.Fatalf("Failed to create search engine: %v", err)
		}
		defer searchEngine.Close()
		
		// Search for multi-word query
		results, err := searchEngine.SearchTodos("user authentication", nil, 10)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		
		// Should find at least the first todo
		if len(results) < 1 {
			t.Error("Expected at least 1 result for 'user authentication'")
		}
		
		// First result should be todo1 (exact match)
		if len(results) > 0 && results[0].ID != todo1.ID {
			t.Errorf("Expected first result to be %s, got %s", todo1.ID, results[0].ID)
		}
		
		// Results should be ranked by relevance
		if len(results) > 1 && results[0].Score <= results[1].Score {
			t.Error("Results should be ordered by relevance score (descending)")
		}
	})
	
	// Test 4: No results for non-existent terms
	t.Run("Return empty results for non-existent terms", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-no-results-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create some todos
		manager := NewTodoManager(tempDir)
		
		_, err = manager.CreateTodo("Normal task", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}
		
		// Create and index
		indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
		searchEngine, err := NewSearchEngine(indexPath, tempDir)
		if err != nil {
			t.Fatalf("Failed to create search engine: %v", err)
		}
		defer searchEngine.Close()
		
		// Search for non-existent term
		results, err := searchEngine.SearchTodos("xyznonexistentterm", nil, 10)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		
		if len(results) != 0 {
			t.Errorf("Expected 0 results for non-existent term, got %d", len(results))
		}
	})
	
	// Test 5: Case-insensitive search
	t.Run("Search should be case-insensitive", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-case-search-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create todos with mixed case
		manager := NewTodoManager(tempDir)
		
		todo1, err := manager.CreateTodo("Implement GraphQL API", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}
		
		// Create and index
		indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
		searchEngine, err := NewSearchEngine(indexPath, tempDir)
		if err != nil {
			t.Fatalf("Failed to create search engine: %v", err)
		}
		defer searchEngine.Close()
		
		// Search with different cases
		testCases := []string{"graphql", "GRAPHQL", "GraphQL", "gRaPhQl"}
		
		for _, query := range testCases {
			results, err := searchEngine.SearchTodos(query, nil, 10)
			if err != nil {
				t.Fatalf("Failed to search for '%s': %v", query, err)
			}
			
			if len(results) != 1 {
				t.Errorf("Expected 1 result for '%s', got %d", query, len(results))
			}
			
			if len(results) > 0 && results[0].ID != todo1.ID {
				t.Errorf("Expected todo ID %s for query '%s', got %s", todo1.ID, query, results[0].ID)
			}
		}
	})
	
	// Test 6: Limit results
	t.Run("Respect result limit", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-limit-search-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)
		
		// Create many todos with common term
		manager := NewTodoManager(tempDir)
		
		for i := 0; i < 20; i++ {
			task := fmt.Sprintf("Task %d: Implement feature", i+1)
			_, err := manager.CreateTodo(task, "medium", "feature")
			if err != nil {
				t.Fatalf("Failed to create todo %d: %v", i+1, err)
			}
		}
		
		// Create and index
		indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
		searchEngine, err := NewSearchEngine(indexPath, tempDir)
		if err != nil {
			t.Fatalf("Failed to create search engine: %v", err)
		}
		defer searchEngine.Close()
		
		// Search with limit
		limit := 5
		results, err := searchEngine.SearchTodos("feature", nil, limit)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}
		
		if len(results) != limit {
			t.Errorf("Expected %d results (limit), got %d", limit, len(results))
		}
	})
}