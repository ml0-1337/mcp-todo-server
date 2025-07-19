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
		expectedPrefix := "Todo 'test-archive-path' has been completed and archived to"
		if !strings.HasPrefix(textContent.Text, expectedPrefix) {
			t.Errorf("Response should start with '%s', got: %s", expectedPrefix, textContent.Text)
		}
		
		// Verify it contains the reflection prompt
		if !strings.Contains(textContent.Text, "Task completed successfully") {
			t.Error("Response should contain task completion prompt")
		}
	})

	t.Run("no-auto-archive flag should disable auto-archiving", func(t *testing.T) {
		// Setup mocks
		mockManager := NewMockTodoManager()
		mockSearch := NewMockSearchEngine()
		mockStats := NewMockStatsEngine()
		mockTemplates := NewMockTemplateManager()

		// Create handlers with auto-archive disabled
		handlers := &TodoHandlers{
			factory:       NewManagerFactory(mockManager, mockSearch, mockStats, mockTemplates),
			noAutoArchive: true, // Simulate --no-auto-archive flag
		}

		// Test todo
		testTodo := &core.Todo{
			ID:       "test-no-auto-archive",
			Task:     "Test with auto-archive disabled",
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
			Started:  time.Now(),
		}

		// Mock ReadTodo
		mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
			if id == "test-no-auto-archive" {
				return testTodo, nil
			}
			return nil, fmt.Errorf("todo not found")
		}

		// Track if ArchiveTodo was called
		archiveCalled := false
		mockManager.ArchiveTodoFunc = func(id string) error {
			archiveCalled = true
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

		// Mock search engine - should re-index since not archiving
		indexCalled := false
		mockSearch.IndexTodoFunc = func(todo *core.Todo, content string) error {
			if todo.ID == "test-no-auto-archive" {
				indexCalled = true
				return nil
			}
			return fmt.Errorf("unexpected todo")
		}

		// Mock ReadTodoContent for re-indexing
		mockManager.ReadTodoContentFunc = func(id string) (string, error) {
			return "# Test Todo Content", nil
		}

		// Create update request
		ctx := context.Background()
		request := &MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id": "test-no-auto-archive",
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

		// Verify ArchiveTodo was NOT called
		if archiveCalled {
			t.Error("ArchiveTodo should not have been called when noAutoArchive is true")
		}

		// Verify search index was updated (re-indexed, not deleted)
		if !indexCalled {
			t.Error("Search index should have been updated when auto-archive is disabled")
		}

		// Verify response does NOT mention archiving
		if strings.Contains(textContent.Text, "archived") {
			t.Errorf("Response should not mention archiving when auto-archive is disabled, got: %s", textContent.Text)
		}

		// Verify standard update message with completion prompt
		expectedPrefix := "Todo 'test-no-auto-archive' metadata updated: status: completed"
		if !strings.HasPrefix(textContent.Text, expectedPrefix) {
			t.Errorf("Response should start with '%s', got: %s", expectedPrefix, textContent.Text)
		}
		
		// Verify it contains the reflection prompt
		if !strings.Contains(textContent.Text, "Task marked as completed") {
			t.Error("Response should contain task completion prompt")
		}
	})
}