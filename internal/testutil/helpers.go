package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	
	"github.com/user/mcp-todo-server/core"
)

// SetupTestDir creates a temporary directory for testing and returns a cleanup function
func SetupTestDir(t testing.TB) (string, func()) {
	t.Helper()
	
	// Use t.TempDir() for automatic cleanup
	tempDir := t.TempDir()
	
	// Ensure the todos subdirectory exists
	todosDir := filepath.Join(tempDir, ".claude/todos")
	if err := os.MkdirAll(todosDir, 0755); err != nil {
		t.Fatalf("Failed to create todos directory: %v", err)
	}
	
	// Return a no-op cleanup since t.TempDir() handles it
	return tempDir, func() {}
}

// SetupTestTodoManager creates a TodoManager with a temporary directory
func SetupTestTodoManager(t testing.TB) (*core.TodoManager, string) {
	t.Helper()
	
	tempDir := t.TempDir()
	
	// Ensure the todos subdirectory exists
	todosDir := filepath.Join(tempDir, ".claude/todos")
	if err := os.MkdirAll(todosDir, 0755); err != nil {
		t.Fatalf("Failed to create todos directory: %v", err)
	}
	
	manager := core.NewTodoManager(tempDir)
	return manager, tempDir
}

// CreateTestTodo creates a todo for testing
func CreateTestTodo(t testing.TB, manager *core.TodoManager, task, priority, todoType string) *core.Todo {
	t.Helper()
	
	todo, err := manager.CreateTodo(task, priority, todoType)
	if err != nil {
		t.Fatalf("Failed to create test todo: %v", err)
	}
	return todo
}

// CreateTestTodoWithDate creates a todo with a specific started date
func CreateTestTodoWithDate(t testing.TB, manager *core.TodoManager, task string, startedDate time.Time) *core.Todo {
	t.Helper()
	
	todo := CreateTestTodo(t, manager, task, "high", "feature")
	
	// Read the file and update the started date directly
	path := filepath.Join(manager.GetBasePath(), ".claude/todos", todo.ID+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read todo file: %v", err)
	}
	
	// Update the started date in the content
	contentStr := string(content)
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

// CreateTestTodoWithContent creates a todo with specific content
func CreateTestTodoWithContent(t testing.TB, manager *core.TodoManager, task, content string) *core.Todo {
	t.Helper()
	
	todo := CreateTestTodo(t, manager, task, "high", "feature")
	
	// Write the content directly to the file
	path := filepath.Join(manager.GetBasePath(), ".claude/todos", todo.ID+".md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	
	return todo
}

// AssertFileExists checks if a file exists at the given path
func AssertFileExists(t testing.TB, path string) {
	t.Helper()
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist at %s", path)
	}
}

// AssertFileNotExists checks if a file does not exist at the given path
func AssertFileNotExists(t testing.TB, path string) {
	t.Helper()
	
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Expected file to not exist at %s", path)
	}
}

// AssertNoError fails the test if an error occurred
func AssertNoError(t testing.TB, err error, message string) {
	t.Helper()
	
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError fails the test if no error occurred
func AssertError(t testing.TB, err error, message string) {
	t.Helper()
	
	if err == nil {
		t.Fatalf("%s: expected error but got nil", message)
	}
}

// AssertEqual fails the test if expected and actual are not equal
func AssertEqual(t testing.TB, expected, actual interface{}, message string) {
	t.Helper()
	
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertContains fails the test if the string does not contain the substring
func AssertContains(t testing.TB, str, substr, message string) {
	t.Helper()
	
	if !strings.Contains(str, substr) {
		t.Errorf("%s: expected '%s' to contain '%s'", message, str, substr)
	}
}

// WaitForFile waits for a file to exist with a timeout
func WaitForFile(t testing.TB, path string, timeout time.Duration) {
	t.Helper()
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	t.Fatalf("File %s did not appear within %v", path, timeout)
}

// CleanupOnFailure runs cleanup function only if the test failed
func CleanupOnFailure(t testing.TB, cleanup func()) {
	t.Helper()
	
	t.Cleanup(func() {
		if t.Failed() {
			cleanup()
		}
	})
}

// RequireFileContains checks that a file contains a substring and fails if not
func RequireFileContains(t testing.TB, path, substring string) {
	t.Helper()
	
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	
	if !strings.Contains(string(content), substring) {
		t.Fatalf("File %s does not contain substring: %s", path, substring)
	}
}

// AssertPathExists checks that a path exists
func AssertPathExists(t testing.TB, path string) {
	t.Helper()
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected path to exist: %s", path)
	}
}

// AssertFileContains checks that a file contains a substring
func AssertFileContains(t testing.TB, path, substring string) {
	t.Helper()
	
	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Failed to read file %s: %v", path, err)
		return
	}
	
	if !strings.Contains(string(content), substring) {
		t.Errorf("File %s does not contain substring: %s", path, substring)
	}
}