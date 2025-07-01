package core

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test 4: todo_create should generate unique IDs for new todos
func TestUniqueIDGeneration(t *testing.T) {
	// Create a todo manager
	manager := NewTodoManager("/tmp/test-todos")

	// Create multiple todos
	todos := []string{
		"Implement authentication",
		"Add user management",
		"Create API documentation",
		"Write unit tests",
		"Deploy to production",
	}

	// Track generated IDs
	generatedIDs := make(map[string]bool)

	for _, task := range todos {
		todo, err := manager.CreateTodo(task, "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Check ID is not empty
		if todo.ID == "" {
			t.Error("Generated todo ID is empty")
		}

		// Check ID format (should be kebab-case)
		if strings.Contains(todo.ID, " ") || strings.Contains(todo.ID, "_") {
			t.Errorf("ID should be kebab-case, got: %s", todo.ID)
		}

		// Check ID is unique
		if generatedIDs[todo.ID] {
			t.Errorf("Duplicate ID generated: %s", todo.ID)
		}
		generatedIDs[todo.ID] = true

		// Check ID is derived from task
		if !strings.Contains(todo.ID, strings.ToLower(strings.Split(task, " ")[0])) {
			t.Errorf("ID should be derived from task. Task: %s, ID: %s", task, todo.ID)
		}
	}

	// Test duplicate task names generate unique IDs
	t.Run("Duplicate tasks get unique IDs", func(t *testing.T) {
		task := "Duplicate task"

		todo1, err := manager.CreateTodo(task, "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create first todo: %v", err)
		}

		todo2, err := manager.CreateTodo(task, "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create second todo: %v", err)
		}

		if todo1.ID == todo2.ID {
			t.Errorf("Duplicate tasks should have unique IDs. Got: %s and %s", todo1.ID, todo2.ID)
		}
	})
}

// Test 5: todo_create should create valid markdown files with frontmatter
func TestMarkdownFileCreation(t *testing.T) {
	// Create temp directory for test
	tempDir, err := ioutil.TempDir("", "todo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after test

	// Create todo manager with temp directory
	manager := NewTodoManager(tempDir)

	// Test data
	task := "Implement authentication system"
	priority := "high"
	todoType := "feature"

	// Create todo
	todo, err := manager.CreateTodo(task, priority, todoType)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Check file exists at expected path
	expectedPath := filepath.Join(tempDir, ".claude", "todos", todo.ID+".md")
	fileInfo, err := os.Stat(expectedPath)
	if err != nil {
		t.Errorf("Todo file not created at expected path %s: %v", expectedPath, err)
	}

	// Check file permissions (should be readable/writable)
	if fileInfo != nil {
		mode := fileInfo.Mode()
		if mode.Perm() != 0644 {
			t.Errorf("File permissions incorrect. Expected 0644, got %o", mode.Perm())
		}
	}

	// Read file content
	content, err := ioutil.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read todo file: %v", err)
	}

	// Verify file starts with YAML frontmatter
	contentStr := string(content)
	if !strings.HasPrefix(contentStr, "---\n") {
		t.Error("File should start with YAML frontmatter delimiter '---'")
	}

	// Extract frontmatter (between first two ---)
	parts := strings.SplitN(contentStr, "---\n", 3)
	if len(parts) < 3 {
		t.Fatal("Invalid markdown format: missing frontmatter delimiters")
	}

	// Parse YAML frontmatter
	var frontmatter struct {
		TodoID    string `yaml:"todo_id"`
		Started   string `yaml:"started"`
		Completed string `yaml:"completed"`
		Status    string `yaml:"status"`
		Priority  string `yaml:"priority"`
		Type      string `yaml:"type"`
	}

	err = yaml.Unmarshal([]byte(parts[1]), &frontmatter)
	if err != nil {
		t.Fatalf("Failed to parse YAML frontmatter: %v", err)
	}

	// Verify frontmatter values
	if frontmatter.TodoID != todo.ID {
		t.Errorf("Frontmatter todo_id mismatch. Expected %s, got %s", todo.ID, frontmatter.TodoID)
	}

	if frontmatter.Status != "in_progress" {
		t.Errorf("Frontmatter status incorrect. Expected 'in_progress', got %s", frontmatter.Status)
	}

	if frontmatter.Priority != priority {
		t.Errorf("Frontmatter priority incorrect. Expected %s, got %s", priority, frontmatter.Priority)
	}

	if frontmatter.Type != todoType {
		t.Errorf("Frontmatter type incorrect. Expected %s, got %s", todoType, frontmatter.Type)
	}

	// Check timestamp format (should be YYYY-MM-DD HH:MM:SS)
	if !strings.Contains(frontmatter.Started, "-") || !strings.Contains(frontmatter.Started, ":") {
		t.Errorf("Started timestamp format incorrect: %s", frontmatter.Started)
	}

	// Verify markdown content contains all required sections
	markdownContent := parts[2] // Content after frontmatter

	requiredSections := []string{
		"# Task: " + task,
		"## Findings & Research",
		"## Test Strategy",
		"## Test List",
		"## Test Cases",
		"## Maintainability Analysis",
		"## Test Results Log",
		"## Checklist",
		"## Working Scratchpad",
	}

	for _, section := range requiredSections {
		if !strings.Contains(markdownContent, section) {
			t.Errorf("Missing required section: %s", section)
		}
	}
}

// Test 6: todo_create should handle file system errors gracefully
func TestFileSystemErrorHandling(t *testing.T) {
	// Test 1: Read-only directory
	t.Run("Read-only directory error", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-readonly-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Make directory read-only
		err = os.Chmod(tempDir, 0555) // r-xr-xr-x (no write permission)
		if err != nil {
			t.Fatalf("Failed to set directory permissions: %v", err)
		}

		// Create todo manager with read-only directory
		manager := NewTodoManager(tempDir)

		// Try to create todo - should fail
		todo, err := manager.CreateTodo("Test task", "high", "feature")

		// Verify error occurred
		if err == nil {
			t.Error("Expected error when writing to read-only directory, but got none")
		}

		// Verify todo is nil on error
		if todo != nil {
			t.Error("Expected nil todo on error, but got a todo object")
		}

		// Verify error message is helpful
		if err != nil && !strings.Contains(err.Error(), "permission") && !strings.Contains(err.Error(), "read-only") {
			t.Errorf("Error message should mention permission issue, got: %v", err)
		}

		// Restore write permission for cleanup
		os.Chmod(tempDir, 0755)
	})

	// Test 2: File already exists and is read-only
	t.Run("Read-only file conflict", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-file-conflict-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a todo first
		todo1, err := manager.CreateTodo("Existing task", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create first todo: %v", err)
		}

		// Make the file read-only
		filePath := filepath.Join(tempDir, ".claude", "todos", todo1.ID+".md")
		err = os.Chmod(filePath, 0444) // r--r--r-- (read-only)
		if err != nil {
			t.Fatalf("Failed to set file permissions: %v", err)
		}

		// Try to create another todo with same name (should generate same ID)
		// This tests if we handle conflicts when file exists
		todo2, err := manager.CreateTodo("Existing task", "low", "bug")

		// Should succeed with different ID (existing-task-2)
		if err != nil {
			t.Errorf("Should handle existing file by generating new ID, got error: %v", err)
		}

		if todo2 != nil && todo2.ID == todo1.ID {
			t.Error("Should generate different ID when file exists")
		}
	})

	// Test 3: Invalid characters in task name
	t.Run("Invalid filesystem characters", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-invalid-chars-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Test various invalid characters that might cause issues
		invalidTasks := []string{
			"Task with \x00 null byte",
			"Task with / forward slash",
			"Task with \\ backslash",
		}

		for _, task := range invalidTasks {
			todo, err := manager.CreateTodo(task, "high", "feature")

			// Should succeed - our ID generation should sanitize these
			if err != nil {
				t.Errorf("Failed to handle task with special chars %q: %v", task, err)
			}

			// Verify file was created
			if todo != nil {
				filePath := filepath.Join(tempDir, ".claude", "todos", todo.ID+".md")
				if _, err := os.Stat(filePath); err != nil {
					t.Errorf("File not created for task %q: %v", task, err)
				}
			}
		}
	})

	// Test 4: Temp file cleanup on error
	t.Run("Temp file cleanup on rename error", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-cleanup-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a todo
		todo, err := manager.CreateTodo("Test cleanup", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Check that no .tmp files remain
		files, err := ioutil.ReadDir(tempDir)
		if err != nil {
			t.Fatalf("Failed to read directory: %v", err)
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".tmp") {
				t.Errorf("Temp file not cleaned up: %s", file.Name())
			}
		}

		// Verify the actual file exists
		expectedFile := todo.ID + ".md"
		found := false
		for _, file := range files {
			if file.Name() == expectedFile {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected file %s not found", expectedFile)
		}
	})
}

// Test 7: todo_read should parse existing markdown files correctly
func TestReadTodo(t *testing.T) {
	// Test 1: Basic file reading
	t.Run("Read basic todo file", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-read-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a todo first to have something to read
		task := "Implement authentication"
		priority := "high"
		todoType := "feature"

		createdTodo, err := manager.CreateTodo(task, priority, todoType)
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Now read it back
		readTodo, err := manager.ReadTodo(createdTodo.ID)
		if err != nil {
			t.Fatalf("Failed to read todo: %v", err)
		}

		// Verify all fields match
		if readTodo.ID != createdTodo.ID {
			t.Errorf("ID mismatch. Expected %s, got %s", createdTodo.ID, readTodo.ID)
		}

		if readTodo.Task != task {
			t.Errorf("Task mismatch. Expected %s, got %s", task, readTodo.Task)
		}

		if readTodo.Priority != priority {
			t.Errorf("Priority mismatch. Expected %s, got %s", priority, readTodo.Priority)
		}

		if readTodo.Type != todoType {
			t.Errorf("Type mismatch. Expected %s, got %s", todoType, readTodo.Type)
		}

		if readTodo.Status != "in_progress" {
			t.Errorf("Status mismatch. Expected 'in_progress', got %s", readTodo.Status)
		}

		// Check timestamp is parsed correctly
		// We just verify it's not zero since exact matching can have timezone issues
		if readTodo.Started.IsZero() {
			t.Error("Started timestamp should be parsed")
		}
	})

	// Test 2: Handle optional fields
	t.Run("Read todo with optional fields", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-optional-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a file with optional fields
		todoContent := `---
todo_id: test-optional-fields
started: 2025-01-27 14:30:00
completed: 2025-01-27 15:45:00
status: completed
priority: medium
type: bug
parent_id: parent-task-123
tags: [bug-fix, urgent, backend]
---

# Task: Fix database connection leak

## Findings & Research

Database connections not being closed properly.

## Test Strategy

Unit tests for connection pool.
`

		// Create directory structure
		todosDir := filepath.Join(tempDir, ".claude", "todos")
		err = os.MkdirAll(todosDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Write test file
		filePath := filepath.Join(todosDir, "test-optional-fields.md")
		err = ioutil.WriteFile(filePath, []byte(todoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Create manager and read
		manager := NewTodoManager(tempDir)
		todo, err := manager.ReadTodo("test-optional-fields")
		if err != nil {
			t.Fatalf("Failed to read todo: %v", err)
		}

		// Verify optional fields
		if todo.ParentID != "parent-task-123" {
			t.Errorf("ParentID not parsed. Expected 'parent-task-123', got %s", todo.ParentID)
		}

		if len(todo.Tags) != 3 {
			t.Errorf("Tags not parsed correctly. Expected 3 tags, got %d", len(todo.Tags))
		}

		if todo.Status != "completed" {
			t.Errorf("Status incorrect. Expected 'completed', got %s", todo.Status)
		}

		// Check completed timestamp
		if todo.Completed.IsZero() {
			t.Error("Completed timestamp should be parsed")
		}

		// Verify task extraction from markdown
		if todo.Task != "Fix database connection leak" {
			t.Errorf("Task not extracted correctly. Got: %s", todo.Task)
		}
	})

	// Test 3: Handle missing file
	t.Run("Error on non-existent file", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-missing-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create manager
		manager := NewTodoManager(tempDir)

		// Try to read non-existent todo
		todo, err := manager.ReadTodo("does-not-exist")

		// Should return error
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}

		// Todo should be nil
		if todo != nil {
			t.Error("Expected nil todo for non-existent file")
		}

		// Error should mention file not found
		if err != nil && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "no such file") {
			t.Errorf("Error should mention file not found, got: %v", err)
		}
	})

	// Test 4: Handle invalid YAML
	t.Run("Error on invalid YAML", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-invalid-yaml-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a file with invalid YAML
		todoContent := `---
todo_id: invalid-yaml
started: not-a-valid-timestamp
status: [this should not be an array]
priority high  # missing colon
---

# Task: Test invalid YAML
`

		// Create directory structure
		todosDir := filepath.Join(tempDir, ".claude", "todos")
		err = os.MkdirAll(todosDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Write test file
		filePath := filepath.Join(todosDir, "invalid-yaml.md")
		err = ioutil.WriteFile(filePath, []byte(todoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Create manager and try to read
		manager := NewTodoManager(tempDir)
		todo, err := manager.ReadTodo("invalid-yaml")

		// Should return error
		if err == nil {
			t.Error("Expected error for invalid YAML, got nil")
		}

		// Todo should be nil
		if todo != nil {
			t.Error("Expected nil todo for invalid YAML")
		}
	})

	// Test 5: Handle missing task heading
	t.Run("Handle missing task heading", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-no-task-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a file without # Task: heading
		todoContent := `---
todo_id: no-task-heading
started: 2025-01-27 14:30:00
completed:
status: in_progress
priority: low
type: research
---

## Findings & Research

This file has no task heading.
`

		// Create directory structure
		todosDir := filepath.Join(tempDir, ".claude", "todos")
		err = os.MkdirAll(todosDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Write test file
		filePath := filepath.Join(todosDir, "no-task-heading.md")
		err = ioutil.WriteFile(filePath, []byte(todoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Create manager and read
		manager := NewTodoManager(tempDir)
		todo, err := manager.ReadTodo("no-task-heading")

		// Should succeed but task should be empty or have default
		if err != nil {
			t.Fatalf("Should handle missing task heading gracefully: %v", err)
		}

		// Task should be empty or indicate missing
		if todo.Task != "" && !strings.Contains(todo.Task, "Untitled") {
			t.Errorf("Task should be empty or indicate missing, got: %s", todo.Task)
		}
	})
}

// Test 9: todo_update should modify specific sections atomically
func TestUpdateTodo(t *testing.T) {
	// Test 1: Update findings section with append
	t.Run("Update findings section with append", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-update-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a todo first
		task := "Test update functionality"
		todo, err := manager.CreateTodo(task, "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Update findings section with append
		newFindings := "\n\n### Additional Research\n\nFound new information about the topic."
		err = manager.UpdateTodo(todo.ID, "findings", "append", newFindings, nil)
		if err != nil {
			t.Fatalf("Failed to update todo: %v", err)
		}

		// Read back and verify
		updated, err := manager.ReadTodo(todo.ID)
		if err != nil {
			t.Fatalf("Failed to read updated todo: %v", err)
		}

		// Read file content to check findings section
		filePath := filepath.Join(tempDir, ".claude", "todos", todo.ID+".md")
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		// Check that new findings were appended
		if !strings.Contains(string(content), "### Additional Research") {
			t.Error("New findings not found in file")
		}

		// Verify task is unchanged
		if updated.Task != task {
			t.Errorf("Task changed unexpectedly. Expected %s, got %s", task, updated.Task)
		}
	})

	// Test 2: Replace test cases section
	t.Run("Update test cases with replace", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-replace-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a todo
		todo, err := manager.CreateTodo("Test replace operation", "medium", "bug")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Replace test cases section
		newTests := `
` + "```go" + `
// Test 1: New test case
func TestNewFeature(t *testing.T) {
    // Test implementation
}
` + "```" + `
`
		err = manager.UpdateTodo(todo.ID, "tests", "replace", newTests, nil)
		if err != nil {
			t.Fatalf("Failed to update todo: %v", err)
		}

		// Read file and verify
		filePath := filepath.Join(tempDir, ".claude", "todos", todo.ID+".md")
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		// Check that test cases were replaced
		if !strings.Contains(string(content), "TestNewFeature") {
			t.Error("New test cases not found in file")
		}

		// Verify other sections still exist
		if !strings.Contains(string(content), "## Findings & Research") {
			t.Error("Other sections were removed")
		}
	})

	// Test 3: Prepend to checklist
	t.Run("Update checklist with prepend", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-prepend-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a todo
		todo, err := manager.CreateTodo("Test prepend operation", "low", "refactor")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Prepend to checklist
		newItems := `
- [ ] Urgent: Review security implications
- [ ] Urgent: Update documentation
`
		err = manager.UpdateTodo(todo.ID, "checklist", "prepend", newItems, nil)
		if err != nil {
			t.Fatalf("Failed to update todo: %v", err)
		}

		// Read file and verify order
		filePath := filepath.Join(tempDir, ".claude", "todos", todo.ID+".md")
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		contentStr := string(content)
		urgentPos := strings.Index(contentStr, "Urgent: Review security")
		checklistPos := strings.Index(contentStr, "## Checklist")

		if urgentPos == -1 {
			t.Fatal("Prepended items not found")
		}

		if urgentPos < checklistPos {
			t.Error("Prepended items appear before checklist header")
		}
	})

	// Test 4: Update metadata status
	t.Run("Update metadata status", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-status-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a todo
		todo, err := manager.CreateTodo("Test status update", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Update status metadata
		metadata := map[string]string{
			"status": "completed",
		}
		err = manager.UpdateTodo(todo.ID, "", "", "", metadata)
		if err != nil {
			t.Fatalf("Failed to update metadata: %v", err)
		}

		// Read back and verify
		updated, err := manager.ReadTodo(todo.ID)
		if err != nil {
			t.Fatalf("Failed to read updated todo: %v", err)
		}

		if updated.Status != "completed" {
			t.Errorf("Status not updated. Expected 'completed', got %s", updated.Status)
		}

		// Verify other metadata unchanged
		if updated.Priority != "high" {
			t.Errorf("Priority changed unexpectedly. Expected 'high', got %s", updated.Priority)
		}
	})

	// Test 5: Update priority metadata
	t.Run("Update priority metadata", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-priority-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a todo
		todo, err := manager.CreateTodo("Test priority update", "low", "research")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Update priority and current_test metadata
		metadata := map[string]string{
			"priority":     "high",
			"current_test": "Test 3: Integration testing",
		}
		err = manager.UpdateTodo(todo.ID, "", "", "", metadata)
		if err != nil {
			t.Fatalf("Failed to update metadata: %v", err)
		}

		// Read back and verify
		updated, err := manager.ReadTodo(todo.ID)
		if err != nil {
			t.Fatalf("Failed to read updated todo: %v", err)
		}

		if updated.Priority != "high" {
			t.Errorf("Priority not updated. Expected 'high', got %s", updated.Priority)
		}

		// Read file to check current_test was added
		filePath := filepath.Join(tempDir, ".claude", "todos", todo.ID+".md")
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if !strings.Contains(string(content), "current_test:") {
			t.Error("current_test metadata not found in file")
		}
		if !strings.Contains(string(content), "Test 3: Integration testing") {
			t.Error("current_test value not found in file")
		}
	})

	// Test 6: Preserve other sections
	t.Run("Preserve other sections", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-preserve-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a todo
		todo, err := manager.CreateTodo("Test preservation", "medium", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// First add some content to scratchpad
		scratchContent := "\n\nSome notes here\nMore notes"
		err = manager.UpdateTodo(todo.ID, "scratchpad", "append", scratchContent, nil)
		if err != nil {
			t.Fatalf("Failed to update scratchpad: %v", err)
		}

		// Now update findings
		findingsContent := "\n\nNew findings here"
		err = manager.UpdateTodo(todo.ID, "findings", "append", findingsContent, nil)
		if err != nil {
			t.Fatalf("Failed to update findings: %v", err)
		}

		// Read file and verify both updates exist
		filePath := filepath.Join(tempDir, ".claude", "todos", todo.ID+".md")
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "Some notes here") {
			t.Error("Scratchpad content was lost")
		}

		if !strings.Contains(contentStr, "New findings here") {
			t.Error("Findings content not found")
		}

		// Verify all sections still exist
		requiredSections := []string{
			"## Findings & Research",
			"## Test Strategy",
			"## Test List",
			"## Test Cases",
			"## Maintainability Analysis",
			"## Test Results Log",
			"## Checklist",
			"## Working Scratchpad",
		}

		for _, section := range requiredSections {
			if !strings.Contains(contentStr, section) {
				t.Errorf("Section %s was removed", section)
			}
		}
	})

	// Test 7: Handle non-existent todo
	t.Run("Handle non-existent todo", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-notfound-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Try to update non-existent todo
		err = manager.UpdateTodo("does-not-exist", "findings", "append", "content", nil)

		// Should return error
		if err == nil {
			t.Error("Expected error for non-existent todo, got nil")
		}

		// Error should mention not found
		if err != nil && !strings.Contains(err.Error(), "not found") {
			t.Errorf("Error should mention 'not found', got: %v", err)
		}
	})

	// Test 8: Atomic updates (concurrent safety)
	t.Run("Handle concurrent updates", func(t *testing.T) {
		// Create temp directory
		tempDir, err := ioutil.TempDir("", "todo-atomic-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create todo manager
		manager := NewTodoManager(tempDir)

		// Create a todo
		todo, err := manager.CreateTodo("Test atomic updates", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Simulate concurrent updates
		done := make(chan bool, 2)
		errors := make(chan error, 2)

		// Goroutine 1: Update findings
		go func() {
			err := manager.UpdateTodo(todo.ID, "findings", "append", "\n\nConcurrent update 1", nil)
			errors <- err
			done <- true
		}()

		// Goroutine 2: Update checklist
		go func() {
			err := manager.UpdateTodo(todo.ID, "checklist", "append", "\n- [ ] Concurrent task", nil)
			errors <- err
			done <- true
		}()

		// Wait for both to complete
		<-done
		<-done

		// Check for errors
		err1 := <-errors
		err2 := <-errors

		if err1 != nil {
			t.Errorf("Goroutine 1 error: %v", err1)
		}
		if err2 != nil {
			t.Errorf("Goroutine 2 error: %v", err2)
		}

		// Verify file is not corrupted
		filePath := filepath.Join(tempDir, ".claude", "todos", todo.ID+".md")
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		// Check both updates made it
		contentStr := string(content)
		if !strings.Contains(contentStr, "Concurrent update 1") || !strings.Contains(contentStr, "Concurrent task") {
			t.Error("One or both concurrent updates were lost")
		}

		// Verify file structure is intact
		if !strings.HasPrefix(contentStr, "---\n") {
			t.Error("File structure corrupted - missing YAML frontmatter")
		}
	})
}
