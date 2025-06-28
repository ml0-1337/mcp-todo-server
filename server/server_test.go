package server

import (
	"testing"
)

// Test 1: MCP server should start and register tools successfully
func TestServerInitialization(t *testing.T) {
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
	defer server.Close()
	
	// Get the list of registered tools
	tools := server.ListTools()
	
	// Expected tools
	expectedTools := []string{
		"todo_create",
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