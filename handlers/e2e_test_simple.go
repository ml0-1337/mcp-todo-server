package handlers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/mcp-todo-server/core"
	"github.com/user/mcp-todo-server/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestE2EBasicWorkflow tests a basic todo creation workflow
func TestE2EBasicWorkflow(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "e2e-basic-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize components
	todoManager := core.NewTodoManager(tempDir)
	searchEngine, err := core.NewSearchEngine(
		filepath.Join(tempDir, ".claude", "index", "todos.bleve"),
		tempDir,
	)
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()

	statsEngine := core.NewStatsEngine(todoManager)
	templateManager := core.NewTemplateManager(utils.GetEnv("CLAUDE_TODO_TEMPLATES_PATH", "templates"))

	handlers := NewTodoHandlersWithDependencies(
		todoManager,
		searchEngine,
		statsEngine,
		templateManager,
	)

	// Test 1: Create a simple todo
	t.Run("CreateSimpleTodo", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"task":     "Test todo creation",
					"priority": "high",
					"type":     "feature",
				},
			},
		}

		result, err := handlers.HandleTodoCreate(context.Background(), req)
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}

		// Check that result has content
		if len(result.Content) == 0 {
			t.Fatal("Expected non-empty result content")
		}

		// The result should contain text with the created todo info
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			t.Logf("Created todo: %s", textContent.Text)
		} else {
			t.Error("Expected text content in result")
		}
	})

	// Test 2: Search for the created todo
	t.Run("SearchTodo", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"query": "Test todo",
					"limit": 10,
				},
			},
		}

		result, err := handlers.HandleTodoSearch(context.Background(), req)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(result.Content) == 0 {
			t.Error("Expected search results")
		}
	})
}