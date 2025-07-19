package handlers

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

func TestTodoUpdate_AutoArchive(t *testing.T) {
	t.Run("setting status to completed should auto-archive by default", func(t *testing.T) {
		// Setup mocks
		mockManager := NewMockTodoManager()
		mockSearch := NewMockSearchEngine()
		mockStats := NewMockStatsEngine()
		mockTemplates := NewMockTemplateManager()

		// Create handlers
		handlers := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

		// Test todo
		testTodo := &core.Todo{
			ID:       "test-auto-archive",
			Task:     "Test auto-archive feature",
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
			Started:  time.Now(),
		}

		// Mock ReadTodo to return our test todo
		mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
			if id == "test-auto-archive" {
				return testTodo, nil
			}
			return nil, fmt.Errorf("todo not found")
		}

		// Track if ArchiveTodo was called
		archiveCalled := false
		mockManager.ArchiveTodoFunc = func(id string) error {
			if id == "test-auto-archive" {
				archiveCalled = true
				return nil
			}
			return fmt.Errorf("unexpected todo ID")
		}

		// Mock UpdateTodo to update the status
		mockManager.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
			if id == "test-auto-archive" && metadata != nil && metadata["status"] == "completed" {
				testTodo.Status = "completed"
				return nil
			}
			return fmt.Errorf("unexpected update")
		}

		// Mock search engine delete
		searchDeleteCalled := false
		mockSearch.DeleteTodoFunc = func(id string) error {
			if id == "test-auto-archive" {
				searchDeleteCalled = true
				return nil
			}
			return fmt.Errorf("unexpected todo ID")
		}

		// Create update request
		ctx := context.Background()
		request := &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id": "test-auto-archive",
				"metadata": map[string]interface{}{
					"status": "completed",
				},
			},
		}

		// Execute update
		result, err := handlers.HandleTodoUpdate(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("HandleTodoUpdate failed: %v", err)
		}

		// Check that response mentions archiving
		if result == nil || len(result.Content) == 0 {
			t.Fatal("Expected result with content")
		}

		// Type assert to TextContent
		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent in result")
		}

		responseText := textContent.Text
		if !strings.Contains(responseText, "archived") {
			t.Errorf("Response should mention 'archived', got: %s", responseText)
		}
		if !strings.Contains(responseText, ".claude/archive/") {
			t.Errorf("Response should contain archive path, got: %s", responseText)
		}

		// Verify ArchiveTodo was called
		if !archiveCalled {
			t.Error("ArchiveTodo should have been called when status set to completed")
		}

		// Verify search index was updated
		if !searchDeleteCalled {
			t.Error("Search index DeleteTodo should have been called")
		}
	})

	t.Run("auto-archive should return archive path in response message", func(t *testing.T) {
		// Setup mocks
		mockManager := NewMockTodoManager()
		mockSearch := NewMockSearchEngine()
		mockStats := NewMockStatsEngine()
		mockTemplates := NewMockTemplateManager()

		// Create handlers with auto-archive enabled (noAutoArchive = false)
		handlers := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

		// Test todo
		testTime := time.Date(2025, 1, 19, 10, 30, 0, 0, time.UTC)
		testTodo := &core.Todo{
			ID:       "test-archive-path",
			Task:     "Test archive path in response",
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
			Started:  testTime,
		}

		// Mock ReadTodo
		mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
			if id == "test-archive-path" {
				return testTodo, nil
			}
			return nil, fmt.Errorf("todo not found")
		}

		// Mock ArchiveTodo
		mockManager.ArchiveTodoFunc = func(id string) error {
			return nil
		}

		// Mock UpdateTodo
		mockManager.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
			if metadata != nil && metadata["status"] == "completed" {
				testTodo.Status = "completed"
				return nil
			}
			return fmt.Errorf("unexpected update")
		}

		// Mock search delete
		mockSearch.DeleteTodoFunc = func(id string) error {
			return nil
		}

		// Create update request
		ctx := context.Background()
		request := &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id": "test-archive-path",
				"metadata": map[string]interface{}{
					"status": "completed",
				},
			},
		}

		// Execute update
		result, err := handlers.HandleTodoUpdate(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("HandleTodoUpdate failed: %v", err)
		}

		// Check response
		if result == nil || len(result.Content) == 0 {
			t.Fatal("Expected result with content")
		}

		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent in result")
		}

		// Verify the response contains the correct archive path
		expectedPath := ".claude/archive/2025/01/19/test-archive-path.md"
		if !strings.Contains(textContent.Text, expectedPath) {
			t.Errorf("Response should contain archive path %s, got: %s", expectedPath, textContent.Text)
		}

		// Verify the message format
		expectedPrefix := "Todo 'test-archive-path' status updated to completed and archived to"
		if !strings.HasPrefix(textContent.Text, expectedPrefix) {
			t.Errorf("Response should start with '%s', got: %s", expectedPrefix, textContent.Text)
		}
	})
}