package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test 1: Bleve date range query with only start date returns todos after that date
func TestBleveDateRangeQuery_StartDateOnly(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-bleve-daterange-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todos with different dates
	manager := NewTodoManager(tempDir)
	
	// Create todo from 10 days ago
	todo1, err := manager.CreateTodo("Old task", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo 1: %v", err)
	}
	tenDaysAgo := time.Now().UTC().AddDate(0, 0, -10)
	err = updateTodoStartDate(tempDir, todo1.ID, tenDaysAgo)
	if err != nil {
		t.Fatalf("Failed to update todo 1 date: %v", err)
	}

	// Create todo from 5 days ago
	todo2, err := manager.CreateTodo("Recent task", "medium", "bug")
	if err != nil {
		t.Fatalf("Failed to create todo 2: %v", err)
	}
	fiveDaysAgo := time.Now().UTC().AddDate(0, 0, -5)
	err = updateTodoStartDate(tempDir, todo2.ID, fiveDaysAgo)
	if err != nil {
		t.Fatalf("Failed to update todo 2 date: %v", err)
	}

	// Create todo from today
	todo3, err := manager.CreateTodo("Today task", "low", "research")
	if err != nil {
		t.Fatalf("Failed to create todo 3: %v", err)
	}
	// Update todo3 to have a consistent time (beginning of today)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	err = updateTodoStartDate(tempDir, todo3.ID, today)
	if err != nil {
		t.Fatalf("Failed to update todo 3 date: %v", err)
	}

	// Create search engine AFTER all todos are created and dates are updated
	// This ensures proper indexing of the updated dates
	indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
	searchEngine, err := NewSearchEngineWithBleveDateRange(indexPath, tempDir)
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()

	// Test: Search with start date (7 days ago) - should return todo2 and todo3
	sevenDaysAgo := time.Now().UTC().AddDate(0, 0, -7)
	filters := map[string]string{
		"date_from": sevenDaysAgo.Format("2006-01-02"),
	}
	
	// Debug: Log the date filter being used
	t.Logf("Searching with date_from: %s", filters["date_from"])
	t.Logf("Expected to find todos after: %s", sevenDaysAgo.String())
	
	// Also log the actual todo dates
	todo1Read, _ := manager.ReadTodo(todo1.ID)
	todo2Read, _ := manager.ReadTodo(todo2.ID)
	todo3Read, _ := manager.ReadTodo(todo3.ID)
	t.Logf("Todo1 started: %s", todo1Read.Started.String())
	t.Logf("Todo2 started: %s", todo2Read.Started.String())
	t.Logf("Todo3 started: %s", todo3Read.Started.String())
	
	results, err := searchEngine.SearchTodos("task", filters, 10)
	if err != nil {
		t.Fatalf("Failed to search with date filter: %v", err)
	}

	// Should return 2 results (todo2 and todo3)
	if len(results) != 2 {
		t.Errorf("Expected 2 results for date_from filter, got %d", len(results))
		for i, r := range results {
			t.Logf("Result %d: ID=%s, Task=%s", i, r.ID, r.Task)
		}
	}

	// Verify the correct todos were returned
	foundTodo2 := false
	foundTodo3 := false
	for _, result := range results {
		if result.ID == todo2.ID {
			foundTodo2 = true
		}
		if result.ID == todo3.ID {
			foundTodo3 = true
		}
	}

	if !foundTodo2 {
		t.Error("Expected todo2 (5 days ago) to be in results")
	}
	if !foundTodo3 {
		t.Error("Expected todo3 (today) to be in results")
	}
}

// Helper function to update todo start date
func updateTodoStartDate(basePath, todoID string, newDate time.Time) error {
	manager := &TodoManager{basePath: basePath}
	
	// Use UpdateTodo with metadata to update the started date
	metadata := map[string]string{
		"started": newDate.Format("2006-01-02T15:04:05Z"),
	}
	
	err := manager.UpdateTodo(todoID, "", "", "", metadata)
	if err != nil {
		return fmt.Errorf("failed to update todo date: %w", err)
	}

	return nil
}

// Test 2: Bleve date range query with only end date returns todos before that date
func TestBleveDateRangeQuery_EndDateOnly(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-bleve-enddate-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todos with different dates
	manager := NewTodoManager(tempDir)
	
	// Create todo from 10 days ago
	todo1, err := manager.CreateTodo("Old task", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo 1: %v", err)
	}
	tenDaysAgo := time.Now().UTC().AddDate(0, 0, -10)
	err = updateTodoStartDate(tempDir, todo1.ID, tenDaysAgo)
	if err != nil {
		t.Fatalf("Failed to update todo 1 date: %v", err)
	}

	// Create todo from 5 days ago
	todo2, err := manager.CreateTodo("Recent task", "medium", "bug")
	if err != nil {
		t.Fatalf("Failed to create todo 2: %v", err)
	}
	fiveDaysAgo := time.Now().UTC().AddDate(0, 0, -5)
	err = updateTodoStartDate(tempDir, todo2.ID, fiveDaysAgo)
	if err != nil {
		t.Fatalf("Failed to update todo 2 date: %v", err)
	}

	// Create todo from today
	todo3, err := manager.CreateTodo("Today task", "low", "research")
	if err != nil {
		t.Fatalf("Failed to create todo 3: %v", err)
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	err = updateTodoStartDate(tempDir, todo3.ID, today)
	if err != nil {
		t.Fatalf("Failed to update todo 3 date: %v", err)
	}

	// Create search engine AFTER all todos are created and dates are updated
	indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
	searchEngine, err := NewSearchEngineWithBleveDateRange(indexPath, tempDir)
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()

	// Test: Search with end date (7 days ago) - should return only todo1
	sevenDaysAgo := time.Now().UTC().AddDate(0, 0, -7)
	filters := map[string]string{
		"date_to": sevenDaysAgo.Format("2006-01-02"),
	}
	
	results, err := searchEngine.SearchTodos("task", filters, 10)
	if err != nil {
		t.Fatalf("Failed to search with date filter: %v", err)
	}

	// Should return 1 result (only todo1)
	if len(results) != 1 {
		t.Errorf("Expected 1 result for date_to filter, got %d", len(results))
		for i, r := range results {
			t.Logf("Result %d: ID=%s, Task=%s", i, r.ID, r.Task)
		}
	}

	// Verify it's todo1
	if len(results) > 0 && results[0].ID != todo1.ID {
		t.Errorf("Expected todo1 (10 days ago) to be the only result, got %s", results[0].ID)
	}
}

// NewSearchEngineWithBleveDateRange creates a search engine that uses bleve date range queries
// This is a test helper that will enable the commented out date range code
func NewSearchEngineWithBleveDateRange(indexPath, todosPath string) (*SearchEngine, error) {
	// For now, just create a regular search engine
	// We'll modify this after uncommenting the date range code
	return NewSearchEngine(indexPath, todosPath)
}