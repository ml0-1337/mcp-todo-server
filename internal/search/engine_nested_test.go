package search

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/mcp-todo-server/internal/domain"
)

// Test 12: Search indexing handles nested directories
func TestSearchIndexing_NestedDirectories(t *testing.T) {
	tempDir := t.TempDir()
	todosPath := filepath.Join(tempDir, ".claude", "todos")
	indexPath := filepath.Join(tempDir, ".claude", "search_index")

	// Create nested directory structure with todos
	dates := []struct {
		year  string
		month string
		day   string
		todos []struct {
			id      string
			content string
		}
	}{
		{
			year:  "2025",
			month: "01",
			day:   "18",
			todos: []struct {
				id      string
				content string
			}{
				{
					id: "search-test-1",
					content: `---
todo_id: search-test-1
task: Implement search functionality
started: 2025-01-18T10:00:00Z
status: in_progress
priority: high
type: feature
---

# Task: Implement search functionality

## Findings & Research

Found that Bleve is a good search engine for Go applications.

## Test Cases

` + "```go\n" + `func TestSearch(t *testing.T) {
    // Test search implementation
}
` + "```",
				},
			},
		},
		{
			year:  "2025",
			month: "01",
			day:   "20",
			todos: []struct {
				id      string
				content string
			}{
				{
					id: "search-test-2",
					content: `---
todo_id: search-test-2
task: Fix search performance
started: 2025-01-20T14:00:00Z
status: completed
priority: medium
type: bug
---

# Task: Fix search performance

## Findings & Research

Search was slow due to missing index.
`,
				},
				{
					id: "search-test-3",
					content: `---
todo_id: search-test-3
task: Add search filters
started: 2025-01-20T16:00:00Z
status: in_progress
priority: low
type: feature
---

# Task: Add search filters

## Test Cases

` + "```javascript\n" + `test('filters work', () => {
    expect(search('test', {status: 'completed'})).toHaveLength(1);
});
` + "```",
				},
			},
		},
	}

	// Create all the todos in nested directories
	for _, date := range dates {
		dir := filepath.Join(todosPath, date.year, date.month, date.day)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}

		for _, todo := range date.todos {
			path := filepath.Join(dir, todo.id+".md")
			if err := os.WriteFile(path, []byte(todo.content), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Create search engine
	engine, err := NewEngine(indexPath, todosPath)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Verify all todos were indexed
	count, err := engine.GetIndexedCount()
	if err != nil {
		t.Fatalf("Failed to get indexed count: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 indexed documents, got %d", count)
	}

	// Test search functionality
	testCases := []struct {
		name          string
		query         string
		filters       map[string]string
		expectedCount int
		expectedIDs   []string
	}{
		{
			name:          "Search all",
			query:         "search",
			expectedCount: 3,
			expectedIDs:   []string{"search-test-1", "search-test-2", "search-test-3"},
		},
		{
			name:          "Search with status filter",
			query:         "search",
			filters:       map[string]string{"status": "completed"},
			expectedCount: 1,
			expectedIDs:   []string{"search-test-2"},
		},
		{
			name:          "Search in findings",
			query:         "Bleve",
			expectedCount: 1,
			expectedIDs:   []string{"search-test-1"},
		},
		{
			name:          "Search with date range",
			query:         "",
			filters: map[string]string{
				"date_from": "2025-01-19",
				"date_to":   "2025-01-20",
			},
			expectedCount: 2,
			expectedIDs:   []string{"search-test-2", "search-test-3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := engine.Search(tc.query, tc.filters, 10)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if len(results) != tc.expectedCount {
				t.Errorf("Expected %d results, got %d", tc.expectedCount, len(results))
			}

			// Verify expected IDs
			foundIDs := make(map[string]bool)
			for _, result := range results {
				foundIDs[result.ID] = true
			}

			for _, expectedID := range tc.expectedIDs {
				if !foundIDs[expectedID] {
					t.Errorf("Expected to find ID %s in results", expectedID)
				}
			}
		})
	}
}

// Test incremental indexing for new todos in nested structure
func TestSearchIndexing_IncrementalUpdate(t *testing.T) {
	tempDir := t.TempDir()
	todosPath := filepath.Join(tempDir, ".claude", "todos")
	indexPath := filepath.Join(tempDir, ".claude", "search_index")

	// Create initial todo
	initialDir := filepath.Join(todosPath, "2025", "01", "20")
	if err := os.MkdirAll(initialDir, 0755); err != nil {
		t.Fatal(err)
	}

	initialContent := `---
todo_id: initial-todo
task: Initial todo
started: 2025-01-20T10:00:00Z
status: in_progress
priority: high
type: feature
---

# Task: Initial todo`

	if err := os.WriteFile(filepath.Join(initialDir, "initial-todo.md"), []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create engine
	engine, err := NewEngine(indexPath, todosPath)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Verify initial count
	count, err := engine.GetIndexedCount()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("Expected 1 indexed document, got %d", count)
	}

	// Add a new todo in a different date directory
	newDir := filepath.Join(todosPath, "2025", "01", "21")
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatal(err)
	}

	newTodo := &domain.Todo{
		ID:       "new-todo",
		Task:     "New incremental todo",
		Started:  time.Date(2025, 1, 21, 10, 0, 0, 0, time.UTC),
		Status:   "in_progress",
		Priority: "medium",
		Type:     "feature",
	}

	newContent := `---
todo_id: new-todo
task: New incremental todo
started: 2025-01-21T10:00:00Z
status: in_progress
priority: medium
type: feature
---

# Task: New incremental todo

## Findings & Research

This is a new todo added after initial indexing.`

	// Index the new todo
	err = engine.Index(newTodo, newContent)
	if err != nil {
		t.Fatalf("Failed to index new todo: %v", err)
	}

	// Search for the new todo
	results, err := engine.Search("incremental", nil, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].ID != "new-todo" {
		t.Errorf("Expected to find new-todo, got %s", results[0].ID)
	}
}