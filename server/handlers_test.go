package server

import (
	"testing"
	"os"
	"path/filepath"
)

// TestNewTodoServer tests creating a new todo server
func TestNewTodoServer(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-todo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create required directories
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	err = os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create todos dir: %v", err)
	}

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
	ts, err := NewTodoServer()
	if err != nil {
		t.Fatalf("Failed to create todo server: %v", err)
	}
	defer ts.Close()

	tools := ts.ListTools()
	
	// Check we have the expected number of tools
	expectedTools := 9
	if len(tools) != expectedTools {
		t.Errorf("Expected %d tools, got %d", expectedTools, len(tools))
	}

	// Check specific tools exist
	toolNames := map[string]bool{
		"todo_create":   false,
		"todo_read":     false,
		"todo_update":   false,
		"todo_search":   false,
		"todo_archive":  false,
		"todo_template": false,
		"todo_link":     false,
		"todo_stats":    false,
		"todo_clean":    false,
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