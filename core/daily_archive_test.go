package core

import (
	"testing"
	"time"
	"os"
	"path/filepath"
	"io/ioutil"
	"strings"
)

// Test daily archive structure implementation
func TestDailyArchiveStructure(t *testing.T) {
	t.Run("Archive creates YYYY/MM/DD directory structure", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := ioutil.TempDir("", "daily-archive-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a test todo with a specific date
		todo, err := manager.CreateTodo("Test daily archive", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Archive the todo
		err = manager.ArchiveTodo(todo.ID, "")
		if err != nil {
			t.Fatalf("Failed to archive todo: %v", err)
		}

		// Verify original file no longer exists
		originalPath := filepath.Join(tempDir, todo.ID+".md")
		if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
			t.Error("Original todo file should have been moved")
		}

		// Verify file exists in daily archive structure
		now := time.Now()
		year := now.Format("2006")
		month := now.Format("01")
		day := now.Format("02")
		expectedArchivePath := filepath.Join(tempDir, "..", "archive", year, month, day, todo.ID+".md")
		
		if _, err := os.Stat(expectedArchivePath); os.IsNotExist(err) {
			t.Errorf("Todo file should exist in archive at %s", expectedArchivePath)
		}

		// Verify completed timestamp was set
		content, err := ioutil.ReadFile(expectedArchivePath)
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
		tempDir, err := ioutil.TempDir("", "cross-month-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create todos on different dates
		// Todo 1: End of January
		todo1, err := manager.CreateTodo("End of month todo", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo1: %v", err)
		}
		// We need to update the file with the new started date
		// For now, let's use a workaround by updating metadata
		metadata1 := map[string]string{
			"started": "2025-01-31 12:00:00",
		}
		err = manager.UpdateTodo(todo1.ID, "", "", "", metadata1)
		if err != nil {
			t.Fatalf("Failed to update todo1 date: %v", err)
		}

		// Todo 2: Beginning of February
		todo2, err := manager.CreateTodo("Start of month todo", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo2: %v", err)
		}
		// Update the started date
		metadata2 := map[string]string{
			"started": "2025-02-01 12:00:00",
		}
		err = manager.UpdateTodo(todo2.ID, "", "", "", metadata2)
		if err != nil {
			t.Fatalf("Failed to update todo2 date: %v", err)
		}

		// For now, just archive normally (will need to implement date override)
		err = manager.ArchiveTodo(todo1.ID, "")
		if err != nil {
			t.Fatalf("Failed to archive todo1: %v", err)
		}

		err = manager.ArchiveTodo(todo2.ID, "")
		if err != nil {
			t.Fatalf("Failed to archive todo2: %v", err)
		}

		// Verify files are in correct directories
		archive1Path := filepath.Join(tempDir, "..", "archive", "2025", "01", "31", todo1.ID+".md")
		archive2Path := filepath.Join(tempDir, "..", "archive", "2025", "02", "01", todo2.ID+".md")

		if _, err := os.Stat(archive1Path); os.IsNotExist(err) {
			t.Errorf("Todo1 should be archived at %s", archive1Path)
		}

		if _, err := os.Stat(archive2Path); os.IsNotExist(err) {
			t.Errorf("Todo2 should be archived at %s", archive2Path)
		}
	})

	t.Run("Archive handles cross-year boundaries correctly", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := ioutil.TempDir("", "cross-year-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create todos on different years
		// Todo 1: End of 2024
		todo1, err := manager.CreateTodo("End of year todo", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo1: %v", err)
		}
		metadata1 := map[string]string{
			"started": "2024-12-31 12:00:00",
		}
		err = manager.UpdateTodo(todo1.ID, "", "", "", metadata1)
		if err != nil {
			t.Fatalf("Failed to update todo1 date: %v", err)
		}

		// Todo 2: Beginning of 2025
		todo2, err := manager.CreateTodo("Start of year todo", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo2: %v", err)
		}
		metadata2 := map[string]string{
			"started": "2025-01-01 12:00:00",
		}
		err = manager.UpdateTodo(todo2.ID, "", "", "", metadata2)
		if err != nil {
			t.Fatalf("Failed to update todo2 date: %v", err)
		}

		// For now, just archive normally (will need to implement date override)
		err = manager.ArchiveTodo(todo1.ID, "")
		if err != nil {
			t.Fatalf("Failed to archive todo1: %v", err)
		}

		err = manager.ArchiveTodo(todo2.ID, "")
		if err != nil {
			t.Fatalf("Failed to archive todo2: %v", err)
		}

		// Verify files are in correct directories
		archive1Path := filepath.Join(tempDir, "..", "archive", "2024", "12", "31", todo1.ID+".md")
		archive2Path := filepath.Join(tempDir, "..", "archive", "2025", "01", "01", todo2.ID+".md")

		if _, err := os.Stat(archive1Path); os.IsNotExist(err) {
			t.Errorf("Todo1 should be archived at %s", archive1Path)
		}

		if _, err := os.Stat(archive2Path); os.IsNotExist(err) {
			t.Errorf("Todo2 should be archived at %s", archive2Path)
		}
	})

	t.Run("Archive creates parent directories if missing", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := ioutil.TempDir("", "parent-dirs-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Ensure archive directory doesn't exist
		archiveBase := filepath.Join(tempDir, "..", "archive")
		os.RemoveAll(archiveBase)

		// Create and archive a todo
		todo, err := manager.CreateTodo("Test parent directory creation", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		err = manager.ArchiveTodo(todo.ID, "")
		if err != nil {
			t.Fatalf("Failed to archive todo: %v", err)
		}

		// Verify all parent directories were created
		now := time.Now()
		year := now.Format("2006")
		month := now.Format("01")
		day := now.Format("02")
		
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
		tempDir, err := ioutil.TempDir("", "preserve-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

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

		// Read original content
		originalPath := filepath.Join(tempDir, todo.ID+".md")
		_, err = ioutil.ReadFile(originalPath)
		if err != nil {
			t.Fatalf("Failed to read original file: %v", err)
		}

		// Archive the todo
		err = manager.ArchiveTodo(todo.ID, "")
		if err != nil {
			t.Fatalf("Failed to archive todo: %v", err)
		}

		// Read archived content
		now := time.Now()
		archivePath := filepath.Join(tempDir, "..", "archive", now.Format("2006"), now.Format("01"), now.Format("02"), todo.ID+".md")
		archivedContent, err := ioutil.ReadFile(archivePath)
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