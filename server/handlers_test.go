package server

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewTodoServer tests creating a new todo server
func TestNewTodoServer(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Set environment variables to use temp directory
	oldTodoPath := os.Getenv("CLAUDE_TODO_PATH")
	oldTemplatePath := os.Getenv("CLAUDE_TEMPLATE_PATH")
	defer func() {
		os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
		os.Setenv("CLAUDE_TEMPLATE_PATH", oldTemplatePath)
	}()

	// Create required directories
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	templatesDir := filepath.Join(tempDir, ".claude", "templates")
	err = os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create todos dir: %v", err)
	}
	err = os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Set environment to use temp directories
	os.Setenv("CLAUDE_TODO_PATH", todosDir)
	os.Setenv("CLAUDE_TEMPLATE_PATH", templatesDir)

	// Create server
	ts, err := NewTodoServer()
	if err != nil {
		t.Fatalf("Failed to create todo server: %v", err)
	}

	// Basic validation
	if ts == nil {
		t.Fatal("Expected non-nil server")
	}
	if ts.mcpServer == nil {
		t.Fatal("Expected non-nil MCP server")
	}
	if ts.handlers == nil {
		t.Fatal("Expected non-nil handlers")
	}

	// Close the server
	err = ts.Close()
	if err != nil {
		t.Errorf("Failed to close server: %v", err)
	}
}

// TestListTools tests listing available tools
func TestListTools(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Set environment variables to use temp directory
	oldTodoPath := os.Getenv("CLAUDE_TODO_PATH")
	oldTemplatePath := os.Getenv("CLAUDE_TEMPLATE_PATH")
	defer func() {
		os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
		os.Setenv("CLAUDE_TEMPLATE_PATH", oldTemplatePath)
	}()

	todosDir := filepath.Join(tempDir, ".claude", "todos")
	templatesDir := filepath.Join(tempDir, ".claude", "templates")
	os.MkdirAll(todosDir, 0755)
	os.MkdirAll(templatesDir, 0755)

	os.Setenv("CLAUDE_TODO_PATH", todosDir)
	os.Setenv("CLAUDE_TEMPLATE_PATH", templatesDir)

	ts, err := NewTodoServer()
	if err != nil {
		t.Fatalf("Failed to create todo server: %v", err)
	}
	defer ts.Close()

	tools := ts.ListTools()

	// Check we have the expected number of tools
	expectedTools := 10 // Including todo_create_multi
	if len(tools) != expectedTools {
		t.Errorf("Expected %d tools, got %d", expectedTools, len(tools))
	}

	// Check specific tools exist
	toolNames := map[string]bool{
		"todo_create":       false,
		"todo_create_multi": false,
		"todo_read":         false,
		"todo_update":       false,
		"todo_search":       false,
		"todo_archive":      false,
		"todo_template":     false,
		"todo_link":         false,
		"todo_stats":        false,
		"todo_clean":        false,
	}

	for _, tool := range tools {
		if _, exists := toolNames[tool.Name]; exists {
			toolNames[tool.Name] = true
		}
	}

	for name, found := range toolNames {
		if !found {
			t.Errorf("Expected tool %s not found", name)
		}
	}
}
