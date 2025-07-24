package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test 13: Migration moves todos to correct date directories
func TestMigration_MovesToCorrectDateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewTodoManager(tempDir)

	// Create flat structure todos directory
	flatDir := filepath.Join(tempDir, ".claude", "todos")
	if err := os.MkdirAll(flatDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create several todos in flat structure with different dates
	todos := []struct {
		id      string
		content string
		date    string
	}{
		{
			id:   "migrate-1",
			date: "2025-01-18T10:00:00Z",
			content: `---
todo_id: migrate-1
started: 2025-01-18T10:00:00Z
status: in_progress
priority: high
type: feature
---

# Task: First migration test`,
		},
		{
			id:   "migrate-2",
			date: "2025-01-20T14:30:00Z",
			content: `---
todo_id: migrate-2
started: 2025-01-20T14:30:00Z
status: completed
priority: medium
type: bug
---

# Task: Second migration test`,
		},
		{
			id:   "migrate-3",
			date: "2025-02-01T09:00:00Z",
			content: `---
todo_id: migrate-3
started: 2025-02-01T09:00:00Z
status: in_progress
priority: low
type: feature
---

# Task: Third migration test`,
		},
	}

	// Write todos to flat structure
	for _, td := range todos {
		path := filepath.Join(flatDir, td.id+".md")
		if err := os.WriteFile(path, []byte(td.content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Clear cache before migration
	globalPathCache.Clear()

	// Run migration
	stats, err := manager.MigrateToDateStructure()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify stats
	if stats.Migrated != 3 {
		t.Errorf("Expected 3 migrated, got %d", stats.Migrated)
	}
	if stats.Failed != 0 {
		t.Errorf("Expected 0 failed, got %d", stats.Failed)
		for _, err := range stats.Errors {
			t.Logf("Error: %v", err)
		}
	}

	// Verify files are in correct locations
	expectedPaths := map[string]string{
		"migrate-1": filepath.Join(tempDir, ".claude", "todos", "2025", "01", "18", "migrate-1.md"),
		"migrate-2": filepath.Join(tempDir, ".claude", "todos", "2025", "01", "20", "migrate-2.md"),
		"migrate-3": filepath.Join(tempDir, ".claude", "todos", "2025", "02", "01", "migrate-3.md"),
	}

	for id, expectedPath := range expectedPaths {
		if _, err := os.Stat(expectedPath); err != nil {
			t.Errorf("Todo %s not found at expected path %s: %v", id, expectedPath, err)
		}

		// Verify content is preserved
		content, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Errorf("Failed to read migrated file %s: %v", id, err)
		}
		if !strings.Contains(string(content), "# Task:") {
			t.Errorf("Content not preserved for %s", id)
		}
	}

	// Verify flat structure is empty
	files, err := os.ReadDir(flatDir)
	if err != nil {
		t.Fatal(err)
	}
	
	mdCount := 0
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
			mdCount++
		}
	}
	if mdCount > 0 {
		t.Errorf("Found %d .md files in flat structure after migration", mdCount)
	}
}

// Test 14: Migration rollback on failure
func TestMigration_RollbackOnFailure(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewTodoManager(tempDir)

	// Create flat structure with todos
	flatDir := filepath.Join(tempDir, ".claude", "todos")
	if err := os.MkdirAll(flatDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid todo
	validContent := `---
todo_id: valid-todo
started: 2025-01-20T10:00:00Z
status: in_progress
priority: high
type: feature
---

# Task: Valid todo`

	validPath := filepath.Join(flatDir, "valid-todo.md")
	if err := os.WriteFile(validPath, []byte(validContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an invalid todo (will cause parse error)
	invalidContent := `This is not valid YAML frontmatter
# Task: Invalid todo`

	invalidPath := filepath.Join(flatDir, "invalid-todo.md")
	if err := os.WriteFile(invalidPath, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Clear cache
	globalPathCache.Clear()

	// Run migration (should fail and rollback)
	stats, err := manager.MigrateToDateStructure()
	if err == nil {
		t.Error("Expected migration to fail")
	}

	// Verify stats show failure
	if stats.Failed == 0 {
		t.Error("Expected at least one failure in stats")
	}

	// Verify original files are still in flat structure
	if _, err := os.Stat(validPath); err != nil {
		t.Error("Valid todo should still exist after rollback")
	}
	if _, err := os.Stat(invalidPath); err != nil {
		t.Error("Invalid todo should still exist after rollback")
	}

	// Verify no date-based directories were created
	dateDir := filepath.Join(tempDir, ".claude", "todos", "2025")
	if _, err := os.Stat(dateDir); !os.IsNotExist(err) {
		t.Error("Date directories should not exist after rollback")
	}
}

// Test 15: Migration handles todos with missing dates
func TestMigration_HandlesMissingDates(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewTodoManager(tempDir)

	// Create flat structure
	flatDir := filepath.Join(tempDir, ".claude", "todos")
	if err := os.MkdirAll(flatDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create todo without started date
	noDateContent := `---
todo_id: no-date-todo
status: in_progress
priority: high
type: feature
---

# Task: Todo without date`

	noDatePath := filepath.Join(flatDir, "no-date-todo.md")
	if err := os.WriteFile(noDatePath, []byte(noDateContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Set a specific mod time for testing
	modTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	if err := os.Chtimes(noDatePath, modTime, modTime); err != nil {
		t.Fatal(err)
	}

	// Clear cache
	globalPathCache.Clear()

	// Run migration
	stats, err := manager.MigrateToDateStructure()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	if stats.Failed != 0 {
		t.Errorf("Expected no failures, got %d", stats.Failed)
		for _, err := range stats.Errors {
			t.Logf("Error: %v", err)
		}
	}

	// The todo should be migrated using file mod time
	expectedPath := filepath.Join(tempDir, ".claude", "todos", "2025", "01", "15", "no-date-todo.md")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("Todo not found at expected path: %v", err)
	}

	// Read the migrated file and verify started date was added
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatal(err)
	}

	// Parse to verify started date
	todo, err := manager.parseTodoFile(string(content))
	if err != nil {
		t.Fatalf("Failed to parse migrated todo: %v", err)
	}

	if todo.Started.IsZero() {
		t.Error("Started date should have been set during migration")
	}

	// The started date should match the mod time (approximately)
	if todo.Started.Sub(modTime).Abs() > time.Minute {
		t.Errorf("Started date %v doesn't match expected mod time %v", todo.Started, modTime)
	}
}

// Test that needsMigration correctly identifies migration status
func TestNeedsMigration(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewTodoManager(tempDir)

	todosDir := filepath.Join(tempDir, ".claude", "todos")

	// Case 1: No todos directory
	needed, err := manager.needsMigration(todosDir)
	if err != nil {
		t.Fatal(err)
	}
	if needed {
		t.Error("Should not need migration when directory doesn't exist")
	}

	// Case 2: Empty todos directory
	if err := os.MkdirAll(todosDir, 0755); err != nil {
		t.Fatal(err)
	}
	needed, err = manager.needsMigration(todosDir)
	if err != nil {
		t.Fatal(err)
	}
	if needed {
		t.Error("Should not need migration when directory is empty")
	}

	// Case 3: Only date-based structure (no flat files)
	dateDir := filepath.Join(todosDir, "2025", "01", "20")
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dateDir, "test.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	needed, err = manager.needsMigration(todosDir)
	if err != nil {
		t.Fatal(err)
	}
	if needed {
		t.Error("Should not need migration when only date-based structure exists")
	}

	// Case 4: Has flat structure files
	if err := os.WriteFile(filepath.Join(todosDir, "flat-todo.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	needed, err = manager.needsMigration(todosDir)
	if err != nil {
		t.Fatal(err)
	}
	if !needed {
		t.Error("Should need migration when flat structure files exist")
	}
}

// Test rollback functionality
func TestRollbackMigration(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewTodoManager(tempDir)

	// Create a backup directory simulating failed migration
	backupDir := filepath.Join(tempDir, ".claude", "todos_migrating")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Add some files to backup
	content := []byte("backup content")
	if err := os.WriteFile(filepath.Join(backupDir, "todo1.md"), content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, "todo2.md"), content, 0644); err != nil {
		t.Fatal(err)
	}

	// Create current todos directory with different content
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	if err := os.MkdirAll(todosDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(todosDir, "new.md"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run rollback
	err := manager.RollbackMigration()
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify backup was restored
	if _, err := os.Stat(filepath.Join(todosDir, "todo1.md")); err != nil {
		t.Error("todo1.md should exist after rollback")
	}
	if _, err := os.Stat(filepath.Join(todosDir, "todo2.md")); err != nil {
		t.Error("todo2.md should exist after rollback")
	}

	// Verify new file is gone
	if _, err := os.Stat(filepath.Join(todosDir, "new.md")); !os.IsNotExist(err) {
		t.Error("new.md should not exist after rollback")
	}

	// Verify backup directory is gone
	if _, err := os.Stat(backupDir); !os.IsNotExist(err) {
		t.Error("Backup directory should be removed after rollback")
	}
}