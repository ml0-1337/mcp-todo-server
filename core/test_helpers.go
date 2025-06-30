package core

import (
	"os"
	"path/filepath"
	"testing"
)

// Path constants for consistent directory structure
const (
	TodoSubdir    = ".claude/todos"
	ArchiveSubdir = ".claude/archive"
)

// CreateTestTodo creates a todo for testing and ensures it's properly saved
func CreateTestTodo(t *testing.T, manager *TodoManager, task string, priority string, todoType string) *Todo {
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
	// Archives are stored one directory up from basePath
	archiveBase := filepath.Join(filepath.Dir(basePath), "archive")
	
	if quarter != "" {
		return filepath.Join(archiveBase, quarter, todo.ID+".md")
	}
	// Use daily path based on started date
	dayPath := GetDailyPath(todo.Started)
	return filepath.Join(archiveBase, dayPath, todo.ID+".md")
}

// VerifyTodoExists checks if a todo file exists at the expected location
func VerifyTodoExists(t *testing.T, basePath, todoID string) {
	path := GetTodoPath(basePath, todoID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected todo file to exist at %s", path)
	}
}

// VerifyTodoNotExists checks if a todo file does not exist at the expected location
func VerifyTodoNotExists(t *testing.T, basePath, todoID string) {
	path := GetTodoPath(basePath, todoID)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Expected todo file to not exist at %s", path)
	}
}

// VerifyArchiveExists checks if a todo exists in the archive
func VerifyArchiveExists(t *testing.T, basePath string, todo *Todo, quarter string) {
	path := GetArchivePath(basePath, todo, quarter)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected archive file to exist at %s", path)
	}
}

// CreateTestTodoWithContent creates a todo and writes specific content to it
func CreateTestTodoWithContent(t *testing.T, manager *TodoManager, task, content string) *Todo {
	todo := CreateTestTodo(t, manager, task, "high", "feature")
	
	// Write the content directly to the file
	path := GetTodoPath(manager.basePath, todo.ID)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	
	return todo
}

// SetupTestTodoManager creates a TodoManager with a temporary directory
func SetupTestTodoManager(t *testing.T) (*TodoManager, string, func()) {
	tempDir, err := os.MkdirTemp("", "todo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Ensure the todos subdirectory exists
	todosDir := filepath.Join(tempDir, TodoSubdir)
	if err := os.MkdirAll(todosDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create todos directory: %v", err)
	}

	manager := NewTodoManager(tempDir)
	
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	
	return manager, tempDir, cleanup
}