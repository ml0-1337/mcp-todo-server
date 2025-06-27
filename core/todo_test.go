package core

import (
	"testing"
	"strings"
	"os"
	"path/filepath"
	"io/ioutil"
	"gopkg.in/yaml.v3"
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