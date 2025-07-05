package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Path constants for consistent directory structure
const (
	TodoSubdir    = ".claude/todos"
	ArchiveSubdir = ".claude/archive"
)

// CreateTestTodo creates a todo for testing and ensures it's properly saved
func CreateTestTodo(t *testing.T, manager *TodoManager, task string, priority string, todoType string) *Todo {
	t.Helper()
	todo, err := manager.CreateTodo(task, priority, todoType)
	if err != nil {
		t.Fatalf("Failed to create test todo: %v", err)
	}
	return todo
}

// GetTodoPath returns the expected path for a todo file
func GetTodoPath(basePath, todoID string) string {
	return filepath.Join(basePath, TodoSubdir, todoID+".md")
}

// GetArchivePath returns the expected archive path for a todo
func GetArchivePath(basePath string, todo *Todo, quarter string) string {
	// Archives are stored within .claude directory structure
	archiveBase := filepath.Join(basePath, ".claude", "archive")
	
	if quarter != "" {
		return filepath.Join(archiveBase, quarter, todo.ID+".md")
	}
	// Use daily path based on started date
	dayPath := GetDailyPath(todo.Started)
	return filepath.Join(archiveBase, dayPath, todo.ID+".md")
}

// VerifyTodoExists checks if a todo file exists at the expected location
func VerifyTodoExists(t *testing.T, basePath, todoID string) {
	t.Helper()
	path := GetTodoPath(basePath, todoID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected todo file to exist at %s", path)
	}
}

// VerifyTodoNotExists checks if a todo file does not exist at the expected location
func VerifyTodoNotExists(t *testing.T, basePath, todoID string) {
	t.Helper()
	path := GetTodoPath(basePath, todoID)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Expected todo file to not exist at %s", path)
	}
}

// VerifyArchiveExists checks if a todo exists in the archive
func VerifyArchiveExists(t *testing.T, basePath string, todo *Todo, quarter string) {
	t.Helper()
	path := GetArchivePath(basePath, todo, quarter)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected archive file to exist at %s", path)
	}
}

// CreateTestTodoWithContent creates a todo and writes specific content to it
func CreateTestTodoWithContent(t *testing.T, manager *TodoManager, task, content string) *Todo {
	t.Helper()
	todo := CreateTestTodo(t, manager, task, "high", "feature")
	
	// Write the content directly to the file
	path := GetTodoPath(manager.basePath, todo.ID)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	
	return todo
}

// CreateTestTodoWithDate creates a todo with a specific started date
func CreateTestTodoWithDate(t *testing.T, manager *TodoManager, task string, startedDate time.Time) *Todo {
	t.Helper()
	todo := CreateTestTodo(t, manager, task, "high", "feature")
	
	// Read the file content
	path := GetTodoPath(manager.basePath, todo.ID)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read todo file: %v", err)
	}
	
	// Update the started date in the frontmatter
	contentStr := string(content)
	// Replace the started date line
	lines := strings.Split(contentStr, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "started:") {
			lines[i] = "started: " + startedDate.Format(time.RFC3339)
			break
		}
	}
	
	// Write back the updated content
	updatedContent := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("Failed to write updated content: %v", err)
	}
	
	// Re-read the todo to get the updated object
	updatedTodo, err := manager.ReadTodo(todo.ID)
	if err != nil {
		t.Fatalf("Failed to re-read todo: %v", err)
	}
	
	return updatedTodo
}

// SetupTestTodoManager creates a TodoManager with a temporary directory
func SetupTestTodoManager(t *testing.T) (*TodoManager, string, func()) {
	t.Helper()
	
	// Use t.TempDir() for automatic cleanup
	tempDir := t.TempDir()

	// Ensure the todos subdirectory exists
	todosDir := filepath.Join(tempDir, TodoSubdir)
	if err := os.MkdirAll(todosDir, 0755); err != nil {
		t.Fatalf("Failed to create todos directory: %v", err)
	}

	manager := NewTodoManager(tempDir)
	
	// Return a no-op cleanup function since t.TempDir() handles it
	cleanup := func() {}
	
	return manager, tempDir, cleanup
}