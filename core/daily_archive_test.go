package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test daily archive structure implementation
func TestDailyArchiveStructure(t *testing.T) {
	t.Run("Archive creates YYYY/MM/DD directory structure", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := os.MkdirTemp("", "daily-archive-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo directory structure
		todoDir := filepath.Join(tempDir, ".claude", "todos")
		err = os.MkdirAll(todoDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create todo directory: %v", err)
		}

		// Create todo manager with the base path
		manager := NewTodoManager(tempDir)

		// Create a test todo with a specific date
		todo, err := manager.CreateTodo("Test daily archive", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Archive the todo
		err = manager.ArchiveTodo(todo.ID)
		if err != nil {
			t.Fatalf("Failed to archive todo: %v", err)
		}

		// Verify original file no longer exists
		originalPath := filepath.Join(todoDir, todo.ID+".md")
		if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
			t.Error("Original todo file should have been moved")
		}

		// Verify file exists in daily archive structure
		// Archive path uses the todo's started date
		year := todo.Started.Format("2006")
		month := todo.Started.Format("01")
		day := todo.Started.Format("02")
		expectedArchivePath := filepath.Join(tempDir, ".claude", "archive", year, month, day, todo.ID+".md")

		if _, err := os.Stat(expectedArchivePath); os.IsNotExist(err) {
			t.Errorf("Todo file should exist in archive at %s", expectedArchivePath)
		}

		// Verify completed timestamp was set
		content, err := os.ReadFile(expectedArchivePath)
		if err != nil {
			t.Fatalf("Failed to read archived file: %v", err)
		}

		if !strings.Contains(string(content), "completed:") ||
			!strings.Contains(string(content), "status: completed") {
			t.Error("Archived todo should have completed timestamp and status")
		}
	})

	t.Run("Archive handles cross-month boundaries correctly", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := os.MkdirTemp("", "cross-month-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo directory structure
		todoDir := filepath.Join(tempDir, ".claude", "todos")
		err = os.MkdirAll(todoDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create todo directory: %v", err)
		}

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create todos on different dates
		// Todo 1: End of January
		date1 := time.Date(2025, 1, 31, 12, 0, 0, 0, time.UTC)
		todo1 := CreateTestTodoWithDate(t, manager, "End of month todo", date1)

		// Todo 2: Beginning of February
		date2 := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)
		todo2 := CreateTestTodoWithDate(t, manager, "Start of month todo", date2)

		// For now, just archive normally (will need to implement date override)
		err = manager.ArchiveTodo(todo1.ID)
		if err != nil {
			t.Fatalf("Failed to archive todo1: %v", err)
		}

		err = manager.ArchiveTodo(todo2.ID)
		if err != nil {
			t.Fatalf("Failed to archive todo2: %v", err)
		}

		// Verify files are in correct directories
		archive1Path := filepath.Join(tempDir, ".claude", "archive", "2025", "01", "31", todo1.ID+".md")
		archive2Path := filepath.Join(tempDir, ".claude", "archive", "2025", "02", "01", todo2.ID+".md")

		if _, err := os.Stat(archive1Path); os.IsNotExist(err) {
			t.Errorf("Todo1 should be archived at %s", archive1Path)
		}

		if _, err := os.Stat(archive2Path); os.IsNotExist(err) {
			t.Errorf("Todo2 should be archived at %s", archive2Path)
		}
	})

	t.Run("Archive handles cross-year boundaries correctly", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := os.MkdirTemp("", "cross-year-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo directory structure
		todoDir := filepath.Join(tempDir, ".claude", "todos")
		err = os.MkdirAll(todoDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create todo directory: %v", err)
		}

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create todos on different years
		// Todo 1: End of 2024
		date1 := time.Date(2024, 12, 31, 12, 0, 0, 0, time.UTC)
		todo1 := CreateTestTodoWithDate(t, manager, "End of year todo", date1)

		// Todo 2: Beginning of 2025
		date2 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		todo2 := CreateTestTodoWithDate(t, manager, "Start of year todo", date2)

		// For now, just archive normally (will need to implement date override)
		err = manager.ArchiveTodo(todo1.ID)
		if err != nil {
			t.Fatalf("Failed to archive todo1: %v", err)
		}

		err = manager.ArchiveTodo(todo2.ID)
		if err != nil {
			t.Fatalf("Failed to archive todo2: %v", err)
		}

		// Verify files are in correct directories
		archive1Path := filepath.Join(tempDir, ".claude", "archive", "2024", "12", "31", todo1.ID+".md")
		archive2Path := filepath.Join(tempDir, ".claude", "archive", "2025", "01", "01", todo2.ID+".md")

		if _, err := os.Stat(archive1Path); os.IsNotExist(err) {
			t.Errorf("Todo1 should be archived at %s", archive1Path)
		}

		if _, err := os.Stat(archive2Path); os.IsNotExist(err) {
			t.Errorf("Todo2 should be archived at %s", archive2Path)
		}
	})

	t.Run("Archive creates parent directories if missing", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := os.MkdirTemp("", "parent-dirs-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo directory structure
		todoDir := filepath.Join(tempDir, ".claude", "todos")
		err = os.MkdirAll(todoDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create todo directory: %v", err)
		}

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Ensure archive directory doesn't exist
		archiveBase := filepath.Join(tempDir, ".claude", "archive")
		os.RemoveAll(archiveBase)

		// Create and archive a todo
		todo, err := manager.CreateTodo("Test parent directory creation", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		err = manager.ArchiveTodo(todo.ID)
		if err != nil {
			t.Fatalf("Failed to archive todo: %v", err)
		}

		// Verify all parent directories were created
		// Archive uses the todo's started date
		year := todo.Started.Format("2006")
		month := todo.Started.Format("01")
		day := todo.Started.Format("02")

		yearDir := filepath.Join(archiveBase, year)
		monthDir := filepath.Join(yearDir, month)
		dayDir := filepath.Join(monthDir, day)

		for _, dir := range []string{archiveBase, yearDir, monthDir, dayDir} {
			if info, err := os.Stat(dir); os.IsNotExist(err) || !info.IsDir() {
				t.Errorf("Directory should exist: %s", dir)
			}
		}
	})

	t.Run("Archive preserves file content and metadata", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := os.MkdirTemp("", "preserve-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo directory structure
		todoDir := filepath.Join(tempDir, ".claude", "todos")
		err = os.MkdirAll(todoDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create todo directory: %v", err)
		}

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a todo and update it with content
		todo, err := manager.CreateTodo("Test content preservation", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Update with some content
		testContent := "\n\nThis is test content in findings section."
		err = manager.UpdateTodo(todo.ID, "findings", "append", testContent, nil)
		if err != nil {
			t.Fatalf("Failed to update todo: %v", err)
		}

		// Read original content using ResolveTodoPath to handle date-based structure
		originalPath, err := ResolveTodoPath(tempDir, todo.ID)
		if err != nil {
			t.Fatalf("Failed to resolve todo path: %v", err)
		}
		_, err = os.ReadFile(originalPath)
		if err != nil {
			t.Fatalf("Failed to read original file: %v", err)
		}

		// Archive the todo
		err = manager.ArchiveTodo(todo.ID)
		if err != nil {
			t.Fatalf("Failed to archive todo: %v", err)
		}

		// Read archived content
		// Archive uses the todo's started date
		archivePath := filepath.Join(tempDir, ".claude", "archive", todo.Started.Format("2006"), todo.Started.Format("01"), todo.Started.Format("02"), todo.ID+".md")
		archivedContent, err := os.ReadFile(archivePath)
		if err != nil {
			t.Fatalf("Failed to read archived file: %v", err)
		}

		// Verify content is preserved (except for completed timestamp and status)
		if !strings.Contains(string(archivedContent), testContent) {
			t.Error("Archived file should preserve custom content")
		}

		if !strings.Contains(string(archivedContent), "# Task: Test content preservation") {
			t.Error("Archived file should preserve task title")
		}
	})
}
