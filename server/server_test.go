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
		// Note: todo_archive is no longer in default list due to auto-archive feature
		"todo_template",
		"todo_link",
		"todo_stats",
		"todo_clean",
	}

	// Verify all expected tools are registered (minus 1 for todo_archive)
	if len(tools) != len(expectedTools)-1 {
		t.Errorf("Expected %d tools, got %d", len(expectedTools)-1, len(tools))
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

// Test auto-archive tool registration behavior
func TestTodoArchive_ToolRegistration(t *testing.T) {
	// Setup test environment
	setupTestEnv := func(t *testing.T) {
		tempDir := t.TempDir()
		
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
	}

	t.Run("todo_archive tool should not appear in tools list by default", func(t *testing.T) {
		setupTestEnv(t)
		
		// Create server with default options (auto-archive enabled, noAutoArchive=false)
		server, err := NewTodoServer()
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}
		defer server.Close()

		// Get list of tools
		tools := server.ListTools()

		// Check that todo_archive is NOT in the list
		for _, tool := range tools {
			if tool.Name == "todo_archive" {
				t.Error("todo_archive tool should not be registered when auto-archive is enabled (default)")
			}
		}

		// Verify other tools are present
		expectedTools := []string{"todo_create", "todo_read", "todo_update", "todo_search"}
		for _, expected := range expectedTools {
			found := false
			for _, tool := range tools {
				if tool.Name == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected tool %s not found in tools list", expected)
			}
		}
	})

	t.Run("todo_archive tool should appear when auto-archive is disabled", func(t *testing.T) {
		setupTestEnv(t)
		
		// Create server with auto-archive disabled
		server, err := NewTodoServer(WithNoAutoArchive(true))
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}
		defer server.Close()

		// Get list of tools
		tools := server.ListTools()

		// Check that todo_archive IS in the list
		found := false
		for _, tool := range tools {
			if tool.Name == "todo_archive" {
				found = true
				break
			}
		}

		if !found {
			t.Error("todo_archive tool should be registered when auto-archive is disabled")
		}
	})
}
