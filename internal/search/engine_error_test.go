package search

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/mcp-todo-server/internal/domain"
)

// TestDirectoryWalkError tests error handling during directory walk
func TestDirectoryWalkError(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "test.bleve")
	todosPath := filepath.Join(tempDir, "todos")

	// Create todos directory
	err := os.MkdirAll(todosPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create todos directory: %v", err)
	}

	// Create a valid todo file
	todoContent := `---
todo_id: test-todo
status: in_progress
---
# Task: Test todo
`
	err = os.WriteFile(filepath.Join(todosPath, "test.md"), []byte(todoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test todo: %v", err)
	}

	// Create engine
	engine, err := NewEngine(indexPath, todosPath)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Create a subdirectory with no read permissions
	badDir := filepath.Join(todosPath, "noaccess")
	err = os.Mkdir(badDir, 0000)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	defer os.Chmod(badDir, 0755) // Restore permissions for cleanup

	// Try to index - should handle the permission error gracefully
	todo := &domain.Todo{
		ID:       "test-todo-2",
		Task:     "Another test todo",
		Status:   "in_progress",
		Started:  time.Now(),
		Type:     "feature",
		Priority: "medium",
	}
	err = engine.Index(todo, "Another test todo content")

	// Indexing individual todo should succeed
	if err != nil {
		t.Errorf("Expected indexing to succeed, got: %v", err)
	}

	// Verify the accessible todo was indexed
	results, err := engine.Search("test todo", nil, 10)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	if len(results) < 1 {
		t.Errorf("Expected at least 1 result, got %d", len(results))
	}
}

// TestInvalidDateFormat tests date parsing error handling
func TestInvalidDateFormat(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "test.bleve")
	todosPath := filepath.Join(tempDir, "todos")

	// Create todos directory
	err := os.MkdirAll(todosPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create todos directory: %v", err)
	}

	// Create test todos
	todos := []struct {
		id      string
		content string
	}{
		{
			id: "todo1",
			content: `---
todo_id: todo1
started: "2025-01-01T10:00:00Z"
status: in_progress
---
# Task: Valid date todo
`,
		},
		{
			id: "todo2",
			content: `---
todo_id: todo2
started: "2025-01-15T10:00:00Z"
status: in_progress
---
# Task: Another valid date todo
`,
		},
	}

	for _, todo := range todos {
		err = os.WriteFile(filepath.Join(todosPath, todo.id+".md"), []byte(todo.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test todo: %v", err)
		}
	}

	// Create engine
	engine, err := NewEngine(indexPath, todosPath)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Test invalid date formats in search
	testCases := []struct {
		name        string
		dateFrom    string
		dateTo      string
		expectError bool
	}{
		{
			name:        "Invalid dateFrom format",
			dateFrom:    "not-a-date",
			dateTo:      "",
			expectError: true,
		},
		{
			name:        "Invalid dateTo format",
			dateFrom:    "",
			dateTo:      "invalid-date",
			expectError: true,
		},
		{
			name:        "Both dates invalid",
			dateFrom:    "2025-13-45", // Invalid month and day
			dateTo:      "2025/01/01", // Wrong format
			expectError: true,
		},
		{
			name:        "Valid date range",
			dateFrom:    "2025-01-01",
			dateTo:      "2025-01-31",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := make(map[string]string)
			if tc.dateFrom != "" {
				filters["date_from"] = tc.dateFrom
			}
			if tc.dateTo != "" {
				filters["date_to"] = tc.dateTo
			}

			results, err := engine.Search("task", filters, 10)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error for invalid date format")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// For valid case, we should get results
				if len(results) == 0 {
					t.Error("Expected results for valid date range")
				}
			}
		})
	}
}
