package search

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	manager := NewTestTodoManager(tempDir)
	
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
	todosPath := filepath.Join(tempDir, ".claude", "todos")
	searchEngine, err := NewSearchEngineWithBleveDateRange(indexPath, todosPath)
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
	
	results, err := searchEngine.Search("task", filters, 10)
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
	// Read the file directly and update the started date
	filePath := filepath.Join(basePath, ".claude", "todos", todoID+".md")
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read todo file: %w", err)
	}
	
	// Simple regex replace for the started date
	contentStr := string(content)
	newValue := fmt.Sprintf(`started: "%s"`, newDate.Format(time.RFC3339))
	
	// Find and replace the started date
	lines := strings.Split(contentStr, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "started:") {
			lines[i] = newValue
			break
		}
	}
	
	// Write back
	updatedContent := strings.Join(lines, "\n")
	return ioutil.WriteFile(filePath, []byte(updatedContent), 0644)
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
	manager := NewTestTodoManager(tempDir)
	
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
	searchEngine, err := NewSearchEngineWithBleveDateRange(indexPath, filepath.Join(tempDir, ".claude", "todos"))
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()

	// Test: Search with end date (7 days ago) - should return only todo1
	sevenDaysAgo := time.Now().UTC().AddDate(0, 0, -7)
	filters := map[string]string{
		"date_to": sevenDaysAgo.Format("2006-01-02"),
	}
	
	results, err := searchEngine.Search("task", filters, 10)
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

// Test 3: Bleve date range query with both dates returns todos in range
func TestBleveDateRangeQuery_BothDates(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-bleve-bothdate-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todos with different dates
	manager := NewTestTodoManager(tempDir)
	
	// Create todo from 15 days ago (should NOT be in range)
	todo1, err := manager.CreateTodo("Very old task", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo 1: %v", err)
	}
	fifteenDaysAgo := time.Now().UTC().AddDate(0, 0, -15)
	err = updateTodoStartDate(tempDir, todo1.ID, fifteenDaysAgo)
	if err != nil {
		t.Fatalf("Failed to update todo 1 date: %v", err)
	}

	// Create todo from 8 days ago (should be in range)
	todo2, err := manager.CreateTodo("Week old task", "medium", "bug")
	if err != nil {
		t.Fatalf("Failed to create todo 2: %v", err)
	}
	eightDaysAgo := time.Now().UTC().AddDate(0, 0, -8)
	err = updateTodoStartDate(tempDir, todo2.ID, eightDaysAgo)
	if err != nil {
		t.Fatalf("Failed to update todo 2 date: %v", err)
	}

	// Create todo from 5 days ago (should be in range)
	todo3, err := manager.CreateTodo("Recent task", "low", "research")
	if err != nil {
		t.Fatalf("Failed to create todo 3: %v", err)
	}
	fiveDaysAgo := time.Now().UTC().AddDate(0, 0, -5)
	err = updateTodoStartDate(tempDir, todo3.ID, fiveDaysAgo)
	if err != nil {
		t.Fatalf("Failed to update todo 3 date: %v", err)
	}

	// Create todo from 2 days ago (should NOT be in range)
	todo4, err := manager.CreateTodo("Very recent task", "high", "refactor")
	if err != nil {
		t.Fatalf("Failed to create todo 4: %v", err)
	}
	twoDaysAgo := time.Now().UTC().AddDate(0, 0, -2)
	err = updateTodoStartDate(tempDir, todo4.ID, twoDaysAgo)
	if err != nil {
		t.Fatalf("Failed to update todo 4 date: %v", err)
	}

	// Create search engine AFTER all todos are created and dates are updated
	indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
	todosPath := filepath.Join(tempDir, ".claude", "todos")
	searchEngine, err := NewSearchEngineWithBleveDateRange(indexPath, todosPath)
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()

	// Test: Search with date range from 10 days ago to 4 days ago
	// Should return todo2 (8 days ago) and todo3 (5 days ago)
	tenDaysAgo := time.Now().UTC().AddDate(0, 0, -10)
	fourDaysAgo := time.Now().UTC().AddDate(0, 0, -4)
	
	filters := map[string]string{
		"date_from": tenDaysAgo.Format("2006-01-02"),
		"date_to":   fourDaysAgo.Format("2006-01-02"),
	}
	
	// Debug: Log the date range being used
	t.Logf("Searching with date range: %s to %s", filters["date_from"], filters["date_to"])
	t.Logf("Expected to find todos between: %s and %s", tenDaysAgo.String(), fourDaysAgo.String())
	
	results, err := searchEngine.Search("task", filters, 10)
	if err != nil {
		t.Fatalf("Failed to search with date range filter: %v", err)
	}

	// Should return 2 results (todo2 and todo3)
	if len(results) != 2 {
		t.Errorf("Expected 2 results for date range filter, got %d", len(results))
		for i, r := range results {
			t.Logf("Result %d: ID=%s, Task=%s", i, r.ID, r.Task)
		}
		
		// Debug: Log all todo dates
		todo1Read, _ := manager.ReadTodo(todo1.ID)
		todo2Read, _ := manager.ReadTodo(todo2.ID)
		todo3Read, _ := manager.ReadTodo(todo3.ID)
		todo4Read, _ := manager.ReadTodo(todo4.ID)
		t.Logf("Todo1 (15 days ago) started: %s", todo1Read.Started.String())
		t.Logf("Todo2 (8 days ago) started: %s", todo2Read.Started.String())
		t.Logf("Todo3 (5 days ago) started: %s", todo3Read.Started.String())
		t.Logf("Todo4 (2 days ago) started: %s", todo4Read.Started.String())
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
		// Should NOT find todo1 or todo4
		if result.ID == todo1.ID {
			t.Errorf("Should not have found todo1 (15 days ago) in range")
		}
		if result.ID == todo4.ID {
			t.Errorf("Should not have found todo4 (2 days ago) in range")
		}
	}

	if !foundTodo2 {
		t.Error("Expected todo2 (8 days ago) to be in results")
	}
	if !foundTodo3 {
		t.Error("Expected todo3 (5 days ago) to be in results")
	}
}

// NewSearchEngineWithBleveDateRange creates a search engine that uses bleve date range queries
// This is a test helper that will enable the commented out date range code
func NewSearchEngineWithBleveDateRange(indexPath, todosPath string) (*Engine, error) {
	// For now, just create a regular search engine
	// We'll modify this after uncommenting the date range code
	return NewEngine(indexPath, todosPath)
}