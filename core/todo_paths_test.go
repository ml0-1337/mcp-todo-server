package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test 1: GetDailyPath returns correct YYYY/MM/DD format
func TestGetDailyPath(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Regular date",
			input:    time.Date(2025, 1, 20, 12, 0, 0, 0, time.UTC),
			expected: "2025/01/20",
		},
		{
			name:     "Single digit month and day",
			input:    time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC),
			expected: "2025/03/05",
		},
		{
			name:     "End of year",
			input:    time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2024/12/31",
		},
		{
			name:     "Beginning of year",
			input:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2025/01/01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDailyPath(tt.input)
			if result != tt.expected {
				t.Errorf("GetDailyPath(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Test 2: GetDateBasedTodoPath generates full path with date structure
func TestGetDateBasedTodoPath(t *testing.T) {
	basePath := "/test/base"
	todoID := "implement-feature"
	started := time.Date(2025, 1, 20, 10, 30, 0, 0, time.UTC)

	expected := filepath.Join("/test/base", ".claude", "todos", "2025", "01", "20", "implement-feature.md")
	result := GetDateBasedTodoPath(basePath, todoID, started)

	if result != expected {
		t.Errorf("GetDateBasedTodoPath() = %v, want %v", result, expected)
	}
}

// Test 3: ResolveTodoPath finds todo in flat structure (backwards compatibility)
func TestResolveTodoPath_FlatStructure(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	if err := os.MkdirAll(todosDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a todo in flat structure
	todoID := "test-todo"
	todoPath := filepath.Join(todosDir, todoID+".md")
	if err := os.WriteFile(todoPath, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Clear cache to ensure fresh search
	globalPathCache.Clear()

	// Test resolution
	resolvedPath, err := ResolveTodoPath(tempDir, todoID)
	if err != nil {
		t.Fatalf("ResolveTodoPath failed: %v", err)
	}

	if resolvedPath != todoPath {
		t.Errorf("ResolveTodoPath() = %v, want %v", resolvedPath, todoPath)
	}

	// Verify it was cached
	if cachedPath, found := globalPathCache.Get(todoID); !found || cachedPath != todoPath {
		t.Error("Path was not cached correctly")
	}
}

// Test 4: ResolveTodoPath finds todo in date-based structure
func TestResolveTodoPath_DateBasedStructure(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	dateDir := filepath.Join(tempDir, ".claude", "todos", "2025", "01", "20")
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a todo in date-based structure
	todoID := "date-based-todo"
	todoPath := filepath.Join(dateDir, todoID+".md")
	if err := os.WriteFile(todoPath, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Clear cache to ensure fresh search
	globalPathCache.Clear()

	// Test resolution
	resolvedPath, err := ResolveTodoPath(tempDir, todoID)
	if err != nil {
		t.Fatalf("ResolveTodoPath failed: %v", err)
	}

	if resolvedPath != todoPath {
		t.Errorf("ResolveTodoPath() = %v, want %v", resolvedPath, todoPath)
	}
}

// Test 5: ResolveTodoPath returns error for non-existent todo
func TestResolveTodoPath_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	if err := os.MkdirAll(todosDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Clear cache
	globalPathCache.Clear()

	// Test resolution for non-existent todo
	_, err := ResolveTodoPath(tempDir, "non-existent-todo")
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

// Test 6: PathCache stores and retrieves paths correctly
func TestPathCache_StoreAndRetrieve(t *testing.T) {
	cache := NewPathCache(10)

	// Test set and get
	cache.Set("todo1", "/path/to/todo1.md")
	cache.Set("todo2", "/path/to/todo2.md")

	// Test retrieval
	if path, found := cache.Get("todo1"); !found || path != "/path/to/todo1.md" {
		t.Errorf("Failed to retrieve todo1: found=%v, path=%v", found, path)
	}

	if path, found := cache.Get("todo2"); !found || path != "/path/to/todo2.md" {
		t.Errorf("Failed to retrieve todo2: found=%v, path=%v", found, path)
	}

	// Test non-existent
	if _, found := cache.Get("todo3"); found {
		t.Error("Found non-existent todo3")
	}
}

// Test 7: PathCache evicts entries when full (LRU)
func TestPathCache_LRUEviction(t *testing.T) {
	cache := NewPathCache(3)

	// Fill cache
	cache.Set("todo1", "/path/1")
	cache.Set("todo2", "/path/2")
	cache.Set("todo3", "/path/3")

	// All should be present
	if _, found := cache.Get("todo1"); !found {
		t.Error("todo1 should be in cache")
	}

	// Add one more, should evict todo1 (oldest)
	cache.Set("todo4", "/path/4")

	// todo1 should be evicted
	if _, found := cache.Get("todo1"); found {
		t.Error("todo1 should have been evicted")
	}

	// Others should still be present
	for _, id := range []string{"todo2", "todo3", "todo4"} {
		if _, found := cache.Get(id); !found {
			t.Errorf("%s should be in cache", id)
		}
	}
}

// Test for ScanDateRange
func TestScanDateRange(t *testing.T) {
	tempDir := t.TempDir()

	// Create todos in different date directories
	dates := []string{"2025/01/18", "2025/01/19", "2025/01/20", "2025/01/21"}
	for _, date := range dates {
		dir := filepath.Join(tempDir, ".claude", "todos", date)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create a todo file
		todoFile := filepath.Join(dir, "todo-"+filepath.Base(date)+".md")
		if err := os.WriteFile(todoFile, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test scanning a range
	startDate := time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)

	paths, err := ScanDateRange(tempDir, startDate, endDate)
	if err != nil {
		t.Fatalf("ScanDateRange failed: %v", err)
	}

	// Should find 2 files (19th and 20th)
	if len(paths) != 2 {
		t.Errorf("Expected 2 files, got %d", len(paths))
	}
}

// Test for MigrateToDateStructure
func TestMigrateToDateStructure(t *testing.T) {
	tempDir := t.TempDir()
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	if err := os.MkdirAll(todosDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a todo in flat structure
	todoID := "migrate-test"
	sourcePath := filepath.Join(todosDir, todoID+".md")
	content := []byte("todo content")
	if err := os.WriteFile(sourcePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Clear cache
	globalPathCache.Clear()

	// Migrate it
	started := time.Date(2025, 1, 20, 10, 0, 0, 0, time.UTC)
	err := MigrateToDateStructure(tempDir, todoID, started)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify source no longer exists
	if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
		t.Error("Source file should have been moved")
	}

	// Verify destination exists
	destPath := GetDateBasedTodoPath(tempDir, todoID, started)
	if _, err := os.Stat(destPath); err != nil {
		t.Errorf("Destination file not found: %v", err)
	}

	// Verify content is preserved
	migratedContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(migratedContent) != string(content) {
		t.Error("Content was not preserved during migration")
	}

	// Verify cache was updated
	if cachedPath, found := globalPathCache.Get(todoID); !found || cachedPath != destPath {
		t.Error("Cache was not updated after migration")
	}
}
