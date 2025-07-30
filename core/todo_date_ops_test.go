package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test 8: writeTodo creates correct directory structure
func TestWriteTodo_CreatesDateStructure(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewTodoManager(tempDir)

	// Create a todo
	todo := &Todo{
		ID:       "test-date-write",
		Task:     "Test date-based write",
		Started:  time.Date(2025, 1, 20, 10, 0, 0, 0, time.UTC),
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
		Sections: getDefaultSections(),
	}

	// Clear cache to ensure fresh write
	globalPathCache.Clear()

	// Write the todo
	err := manager.writeTodo(todo)
	if err != nil {
		t.Fatalf("Failed to write todo: %v", err)
	}

	// Verify the file was created in the correct location
	expectedPath := filepath.Join(tempDir, ".claude", "todos", "2025", "01", "20", "test-date-write.md")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("Todo file not created at expected path: %v", err)
	}

	// Verify the path was cached
	if cachedPath, found := globalPathCache.Get(todo.ID); !found || cachedPath != expectedPath {
		t.Errorf("Path not cached correctly: found=%v, path=%v", found, cachedPath)
	}
}

// Test 9: writeTodo handles concurrent writes to same date
func TestWriteTodo_ConcurrentWrites(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewTodoManager(tempDir)

	// Create multiple todos with same date
	date := time.Date(2025, 1, 20, 10, 0, 0, 0, time.UTC)
	todos := []*Todo{
		{ID: "concurrent-1", Task: "Task 1", Started: date, Status: "in_progress", Priority: "high", Type: "feature", Sections: getDefaultSections()},
		{ID: "concurrent-2", Task: "Task 2", Started: date, Status: "in_progress", Priority: "high", Type: "feature", Sections: getDefaultSections()},
		{ID: "concurrent-3", Task: "Task 3", Started: date, Status: "in_progress", Priority: "high", Type: "feature", Sections: getDefaultSections()},
	}

	// Write todos concurrently
	errChan := make(chan error, len(todos))
	for _, todo := range todos {
		go func(t *Todo) {
			errChan <- manager.writeTodo(t)
		}(todo)
	}

	// Check for errors
	for i := 0; i < len(todos); i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent write failed: %v", err)
		}
	}

	// Verify all files were created
	dateDir := filepath.Join(tempDir, ".claude", "todos", "2025", "01", "20")
	files, err := os.ReadDir(dateDir)
	if err != nil {
		t.Fatalf("Failed to read date directory: %v", err)
	}

	if len(files) != len(todos) {
		t.Errorf("Expected %d files, got %d", len(todos), len(files))
	}
}

// Test 10: ListTodos finds todos in nested directories
func TestListTodos_NestedDirectories(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewTodoManager(tempDir)

	// Create todos in different date directories
	todos := []struct {
		id   string
		task string
		date time.Time
	}{
		{"nested-1", "Task 1", time.Date(2025, 1, 18, 0, 0, 0, 0, time.UTC)},
		{"nested-2", "Task 2", time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)},
		{"nested-3", "Task 3", time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)},
	}

	// Create todos
	for _, td := range todos {
		todo := &Todo{
			ID:       td.id,
			Task:     td.task,
			Started:  td.date,
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
			Sections: getDefaultSections(),
		}
		if err := manager.writeTodo(todo); err != nil {
			t.Fatalf("Failed to write todo %s: %v", td.id, err)
		}
	}

	// List all todos
	found, err := manager.ListTodos("", "", 0)
	if err != nil {
		t.Fatalf("ListTodos failed: %v", err)
	}

	if len(found) != len(todos) {
		t.Errorf("Expected %d todos, found %d", len(todos), len(found))
	}

	// Verify they're sorted by started date (newest first)
	for i := 1; i < len(found); i++ {
		if found[i-1].Started.Before(found[i].Started) {
			t.Error("Todos not sorted by started date (newest first)")
		}
	}
}

// Test 11: ListTodos with date filter optimizes directory traversal
func TestListTodos_DateFilterOptimization(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewTodoManager(tempDir)

	// Create todos across many dates, centered around current time
	now := time.Now()
	for i := -30; i < 30; i++ {
		date := now.AddDate(0, 0, i)
		todo := &Todo{
			ID:       fmt.Sprintf("date-opt-%d", i+30),
			Task:     fmt.Sprintf("Task %d", i+30),
			Started:  date,
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
			Sections: getDefaultSections(),
		}
		if err := manager.writeTodo(todo); err != nil {
			t.Fatalf("Failed to write todo: %v", err)
		}
	}

	// List todos from last 7 days (should use optimized path)
	found, err := manager.ListTodos("", "", 7)
	if err != nil {
		t.Fatalf("ListTodos with date filter failed: %v", err)
	}

	// Should find approximately 7-8 todos (including today)
	if len(found) < 7 || len(found) > 8 {
		t.Errorf("Expected 7-8 todos, found %d", len(found))
	}

	// All found todos should be within the last 7 days
	cutoff := now.AddDate(0, 0, -7)
	for _, todo := range found {
		if todo.Started.Before(cutoff) {
			t.Errorf("Found todo older than cutoff: %s started at %v", todo.ID, todo.Started)
		}
	}
}

// Test reading and updating with date structure
func TestReadUpdateWithDateStructure(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewTodoManager(tempDir)

	// Create a todo
	todo, err := manager.CreateTodo("Test read/update", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Read it back
	readTodo, err := manager.ReadTodo(todo.ID)
	if err != nil {
		t.Fatalf("Failed to read todo: %v", err)
	}

	if readTodo.Task != todo.Task {
		t.Errorf("Read todo has wrong task: got %s, want %s", readTodo.Task, todo.Task)
	}

	// Update it
	err = manager.UpdateTodo(todo.ID, "findings", "append", "Test finding", nil)
	if err != nil {
		t.Fatalf("Failed to update todo: %v", err)
	}

	// Read the content
	content, err := manager.ReadTodoContent(todo.ID)
	if err != nil {
		t.Fatalf("Failed to read todo content: %v", err)
	}

	if !strings.Contains(content, "Test finding") {
		t.Error("Updated content not found")
	}
}
