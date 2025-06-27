package core

import (
	"testing"
	"strings"
	"os"
	"path/filepath"
	"io/ioutil"
	"gopkg.in/yaml.v3"
	"time"
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
	expectedPath := filepath.Join(tempDir, todo.ID + ".md")
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
		filePath := filepath.Join(tempDir, todo1.ID + ".md")
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
				filePath := filepath.Join(tempDir, todo.ID + ".md")
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
		
		// Check timestamp is parsed correctly (should be close to creation time)
		timeDiff := readTodo.Started.Sub(createdTodo.Started).Abs()
		if timeDiff > time.Second {
			t.Errorf("Started time mismatch. Difference: %v", timeDiff)
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
		
		// Write test file
		filePath := filepath.Join(tempDir, "test-optional-fields.md")
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
		
		// Write test file
		filePath := filepath.Join(tempDir, "invalid-yaml.md")
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
		
		// Write test file
		filePath := filepath.Join(tempDir, "no-task-heading.md")
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