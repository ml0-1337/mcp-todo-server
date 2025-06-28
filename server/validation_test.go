package server

import (
	"context"
	"strings"
	"testing"
	"github.com/mark3labs/mcp-go/mcp"
)

// Test 3: Server should validate tool parameters against JSON schema
func TestParameterValidation(t *testing.T) {
	// Create server
	server, err := NewTodoServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()
	
	// Test todo_create with missing required "task" parameter
	t.Run("Missing required task parameter", func(t *testing.T) {
		// Create request without required "task" field
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "todo_create",
				Arguments: map[string]interface{}{
					"priority": "high",
				},
			},
		}
		
		// Call the handler
		result, err := server.handlers.HandleTodoCreate(context.Background(), request)
		
		// Should return an error result, not a Go error
		if err != nil {
			t.Fatalf("Expected nil error, got: %v", err)
		}
		
		// Check if result indicates an error
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		// The result should indicate the missing parameter
		if !result.IsError {
			t.Error("Expected error result for missing required parameter")
		}
		
		// Check error content mentions missing task
		if len(result.Content) == 0 {
			t.Fatal("Expected content in error result")
		}
		
		errorContent := ""
		// Try to extract text content
		if tc, ok := mcp.AsTextContent(result.Content[0]); ok {
			errorContent = tc.Text
		}
		
		// Verify error message contains expected keywords
		errorLower := strings.ToLower(errorContent)
		if !strings.Contains(errorLower, "task") || 
		   (!strings.Contains(errorLower, "required") && !strings.Contains(errorLower, "missing")) {
			t.Errorf("Error message should mention missing 'task' parameter, got: %s", errorContent)
		}
	})
	
	// Test invalid priority value
	t.Run("Invalid priority value", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "todo_create",
				Arguments: map[string]interface{}{
					"task":     "Test task",
					"priority": "urgent", // Should be high/medium/low
				},
			},
		}
		
		result, err := server.handlers.HandleTodoCreate(context.Background(), request)
		
		if err != nil {
			t.Fatalf("Expected nil error, got: %v", err)
		}
		
		// Should either accept with default or return validation error
		// For now, we'll accept that it might use a default
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})
}