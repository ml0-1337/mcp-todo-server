package server

import (
	"os"
	"path/filepath"
	"testing"
)

// Test 1: MCP server should start and register tools successfully
func TestServerInitialization(t *testing.T) {
	t.Helper()
	
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Set environment variables to use temp directory
	oldTodoPath := os.Getenv("CLAUDE_TODO_PATH")
	oldTemplatePath := os.Getenv("CLAUDE_TEMPLATE_PATH")
	t.Cleanup(func() {
		os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
		os.Setenv("CLAUDE_TEMPLATE_PATH", oldTemplatePath)
	})

	todosDir := filepath.Join(tempDir, ".claude", "todos")
	templatesDir := filepath.Join(tempDir, ".claude", "templates")
	os.MkdirAll(todosDir, 0755)
	os.MkdirAll(templatesDir, 0755)

	os.Setenv("CLAUDE_TODO_PATH", todosDir)
	os.Setenv("CLAUDE_TEMPLATE_PATH", templatesDir)

	// Create a new MCP todo server
	server, err := NewTodoServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Verify server is not nil
	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	// Clean up
	t.Cleanup(func() {
		server.Close()
	})

	// Get the list of registered tools
	tools := server.ListTools()

	// Expected tools
	expectedTools := []string{
		"todo_create",
		"todo_create_multi",
		"todo_read",
		"todo_update",
		"todo_search",
		"todo_archive",
		"todo_template",
		"todo_link",
		"todo_stats",
		"todo_clean",
	}

	// Verify all expected tools are registered
	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	// Check each tool is present
	toolMap := make(map[string]bool)
	for _, tool := range tools {
		toolMap[tool.Name] = true
	}

	for _, expectedTool := range expectedTools {
		if !toolMap[expectedTool] {
			t.Errorf("Expected tool '%s' not found", expectedTool)
		}
	}
}
