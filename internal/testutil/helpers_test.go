package testutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/mcp-todo-server/core"
	"github.com/user/mcp-todo-server/internal/testutil"
)

func TestSetupTestTodoManager(t *testing.T) {
	t.Helper()

	// Test basic setup
	manager, tempDir := testutil.SetupTestTodoManager(t)

	// Verify manager is not nil
	if manager == nil {
		t.Fatal("Expected manager to be created, got nil")
	}

	// Verify temp directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Expected temp directory to exist: %s", tempDir)
	}

	// Verify todos directory exists
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	if _, err := os.Stat(todosDir); os.IsNotExist(err) {
		t.Errorf("Expected todos directory to exist: %s", todosDir)
	}

	// Test that we can create a todo
	task := "Test todo creation"
	todo, err := manager.CreateTodo(task, "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create test todo: %v", err)
	}

	if todo.Task != task {
		t.Errorf("Expected task %q, got %q", task, todo.Task)
	}

	// Verify todo file exists (using ResolveTodoPath to handle date-based structure)
	todoPath, err := core.ResolveTodoPath(tempDir, todo.ID)
	if err != nil {
		t.Errorf("Failed to resolve todo path: %v", err)
	} else if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		t.Errorf("Expected todo file to exist: %s", todoPath)
	}
}

func TestAssertPathExists(t *testing.T) {
	t.Helper()

	// Create a temp directory
	tempDir := t.TempDir()

	// Test with existing path
	testutil.AssertPathExists(t, tempDir)

	// Test with non-existing path - this should fail
	nonExistentPath := filepath.Join(tempDir, "non-existent")

	mock := newMockTB(t)
	testutil.AssertPathExists(mock, nonExistentPath)

	if !mock.errorCalled {
		t.Error("Expected AssertPathExists to fail for non-existent path")
	}
}

func TestAssertFileContains(t *testing.T) {
	t.Helper()

	// Create a temp file with content
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := "This is a test file\nwith multiple lines\nand some content"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with substring that exists
	testutil.AssertFileContains(t, testFile, "test file")
	testutil.AssertFileContains(t, testFile, "multiple lines")

	// Test with substring that doesn't exist
	mock := newMockTB(t)
	testutil.AssertFileContains(mock, testFile, "does not exist")

	if !mock.errorCalled {
		t.Error("Expected AssertFileContains to fail for non-existent substring")
	}
}

func TestCreateTestTodo(t *testing.T) {
	t.Helper()

	manager, _ := testutil.SetupTestTodoManager(t)

	// Test creating a todo
	task := "Test task"
	priority := "medium"
	todoType := "bug"

	todo := testutil.CreateTestTodo(t, manager, task, priority, todoType)

	if todo == nil {
		t.Fatal("Expected todo to be created, got nil")
	}

	if todo.Task != task {
		t.Errorf("Expected task %q, got %q", task, todo.Task)
	}

	if todo.Priority != priority {
		t.Errorf("Expected priority %q, got %q", priority, todo.Priority)
	}

	if todo.Type != todoType {
		t.Errorf("Expected type %q, got %q", todoType, todo.Type)
	}
}
