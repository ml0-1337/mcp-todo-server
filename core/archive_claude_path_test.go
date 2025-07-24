package core

import (
	"os"
	"path/filepath"
	"testing"
)

// Test 2: Archive works correctly with CLAUDE_TODO_PATH set
func TestArchiveWithClaudeTodoPath(t *testing.T) {
	// Create a temporary directory to simulate a project
	projectDir := t.TempDir()
	
	// basePath simulates what would be set from CLAUDE_TODO_PATH
	basePath := projectDir
	
	// Create todo manager with this base path
	manager := NewTodoManager(basePath)
	
	// Create a test todo
	todo, err := manager.CreateTodo("Test with CLAUDE_TODO_PATH", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}
	
	// Verify todo exists using ResolveTodoPath
	todoPath, err := ResolveTodoPath(basePath, todo.ID)
	if err != nil {
		t.Errorf("Failed to resolve todo path: %v", err)
	}
	if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		t.Errorf("Todo should exist at: %s", todoPath)
	}
	
	// Archive the todo
	err = manager.ArchiveTodo(todo.ID)
	if err != nil {
		t.Fatalf("Failed to archive todo: %v", err)
	}
	
	// Verify todo was removed from original location
	_, err = ResolveTodoPath(basePath, todo.ID)
	if err == nil {
		t.Error("Original todo file should have been removed")
	}
	
	// Verify archive was created within the project directory
	// Use the todo's started date for the archive path
	archivePath := filepath.Join(basePath, ".claude", "archive", 
		todo.Started.Format("2006"), todo.Started.Format("01"), todo.Started.Format("02"), 
		todo.ID+".md")
	
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Errorf("Archive should exist within project at: %s", archivePath)
	}
	
	// Important: Verify archive is NOT in parent directory
	wrongArchivePath := filepath.Join(filepath.Dir(basePath), "archive",
		todo.Started.Format("2006"), todo.Started.Format("01"), todo.Started.Format("02"), 
		todo.ID+".md")
	
	if _, err := os.Stat(wrongArchivePath); !os.IsNotExist(err) {
		t.Errorf("Archive should NOT exist in parent directory: %s", wrongArchivePath)
	}
	
	t.Logf("Base path (CLAUDE_TODO_PATH): %s", basePath)
	t.Logf("Archive path (correct): %s", archivePath)
	t.Logf("Archive path (wrong): %s", wrongArchivePath)
}