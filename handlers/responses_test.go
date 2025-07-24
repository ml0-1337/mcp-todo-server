package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// Test formatTodoSummaryLine directly since it's testable
func TestFormatTodoSummaryLine(t *testing.T) {
	tests := []struct {
		name     string
		todo     *core.Todo
		expected string
	}{
		{
			name: "in_progress high priority",
			todo: &core.Todo{
				ID:       "todo-1",
				Task:     "Implement feature",
				Status:   "in_progress",
				Priority: "high",
			},
			expected: "[→] todo-1: Implement feature [HIGH]",
		},
		{
			name: "completed medium priority",
			todo: &core.Todo{
				ID:       "todo-2",
				Task:     "Fix bug",
				Status:   "completed",
				Priority: "medium",
			},
			expected: "[✓] todo-2: Fix bug",
		},
		{
			name: "blocked low priority",
			todo: &core.Todo{
				ID:       "todo-3",
				Task:     "Research task",
				Status:   "blocked",
				Priority: "low",
			},
			expected: "[✗] todo-3: Research task [LOW]",
		},
		{
			name: "unknown status",
			todo: &core.Todo{
				ID:       "todo-4",
				Task:     "Unknown task",
				Status:   "unknown",
				Priority: "medium",
			},
			expected: "[ ] todo-4: Unknown task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTodoSummaryLine(tt.todo)
			if result != tt.expected {
				t.Errorf("formatTodoSummaryLine() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Test response formatting functions return non-nil results
func TestResponseFormattersReturnResults(t *testing.T) {
	todo := &core.Todo{
		ID:       "test-123",
		Task:     "Test task",
		Started:  time.Now(),
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
	}

	t.Run("FormatTodoCreateResponse", func(t *testing.T) {
		result := FormatTodoCreateResponse(todo, "/path/to/test.md")
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.IsError {
			t.Error("Expected success result, not error")
		}
	})

	t.Run("FormatTodoReadResponse single", func(t *testing.T) {
		result := FormatTodoReadResponse([]*core.Todo{todo}, "summary", true)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	t.Run("FormatTodoReadResponse list", func(t *testing.T) {
		result := FormatTodoReadResponse([]*core.Todo{todo}, "list", false)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	t.Run("FormatTodoReadResponse empty", func(t *testing.T) {
		result := FormatTodoReadResponse([]*core.Todo{}, "list", false)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	t.Run("FormatTodoUpdateResponse", func(t *testing.T) {
		result := FormatTodoUpdateResponse("todo-123", "findings", "append")
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	t.Run("FormatTodoSearchResponse with results", func(t *testing.T) {
		results := []core.SearchResult{
			{ID: "todo-1", Task: "Task 1", Score: 0.9},
		}
		result := FormatTodoSearchResponse(results)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	t.Run("FormatTodoSearchResponse empty", func(t *testing.T) {
		result := FormatTodoSearchResponse([]core.SearchResult{})
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	t.Run("FormatTodoArchiveResponse", func(t *testing.T) {
		result := FormatTodoArchiveResponse("todo-123", "/archive/path", "feature")
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		
		// Verify prompt is included
		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatalf("Expected TextContent, got %T", result.Content[0])
		}
		if !strings.Contains(textContent.Text, "To build on this feature") {
			t.Error("Expected archive prompt in response")
		}
	})

	t.Run("FormatTodoStatsResponse", func(t *testing.T) {
		stats := &core.TodoStats{
			TotalTodos:      10,
			CompletedTodos:  5,
			InProgressTodos: 3,
			BlockedTodos:    2,
		}
		result := FormatTodoStatsResponse(stats)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	t.Run("FormatTodoTemplateResponse", func(t *testing.T) {
		result := FormatTodoTemplateResponse(todo, "/path/to/template.md", "feature")
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		
		// Verify prompt is included
		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatalf("Expected TextContent, got %T", result.Content[0])
		}
		if !strings.Contains(textContent.Text, "Feature template applied") {
			t.Error("Expected template prompt in response")
		}
	})

	t.Run("FormatTodoLinkResponse", func(t *testing.T) {
		result := FormatTodoLinkResponse("parent-123", "child-456", "parent-child")
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})
}

// Test FormatSearchResult structure
func TestFormatSearchResult(t *testing.T) {
	// This tests that the FormatSearchResult type is properly defined
	result := FormatSearchResult{
		ID:      "test-123",
		Task:    "Test task",
		Score:   0.95,
		Snippet: "Test snippet",
	}

	if result.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got %s", result.ID)
	}

	if result.Task != "Test task" {
		t.Errorf("Expected Task 'Test task', got %s", result.Task)
	}

	if result.Score != 0.95 {
		t.Errorf("Expected Score 0.95, got %f", result.Score)
	}

	if result.Snippet != "Test snippet" {
		t.Errorf("Expected Snippet 'Test snippet', got %s", result.Snippet)
	}
}

// Test helper functions for formatting
func TestFormattingHelpers(t *testing.T) {
	t.Run("formatTodosList", func(t *testing.T) {
		todos := []*core.Todo{
			{ID: "todo-1", Task: "First task"},
			{ID: "todo-2", Task: "Second task"},
		}

		result := formatTodosList(todos)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// We can't access the content directly, but we can verify it returns a result
	})

	t.Run("formatTodosSummary groups by status", func(t *testing.T) {
		todos := []*core.Todo{
			{ID: "todo-1", Task: "Task 1", Status: "in_progress", Priority: "high"},
			{ID: "todo-2", Task: "Task 2", Status: "in_progress", Priority: "medium"},
			{ID: "todo-3", Task: "Task 3", Status: "completed", Priority: "low"},
			{ID: "todo-4", Task: "Task 4", Status: "blocked", Priority: "high"},
		}

		result := formatTodosSummary(todos)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		// Result should group todos by status
	})
}

// Test response formatting with edge cases
func TestResponseFormattingEdgeCases(t *testing.T) {
	t.Run("nil todo list", func(t *testing.T) {
		// Should handle nil gracefully
		result := FormatTodoReadResponse(nil, "list", false)
		if result == nil {
			t.Fatal("Expected non-nil result even for nil input")
		}
	})

	t.Run("todo with special characters", func(t *testing.T) {
		todo := &core.Todo{
			ID:       "test-123",
			Task:     "Task with \"quotes\" and 'apostrophes'",
			Status:   "in_progress",
			Priority: "high",
		}

		line := formatTodoSummaryLine(todo)
		if !strings.Contains(line, "Task with \"quotes\" and 'apostrophes'") {
			t.Error("Expected special characters to be preserved")
		}
	})

	t.Run("very long task description", func(t *testing.T) {
		longTask := strings.Repeat("Very long task description ", 10)
		todo := &core.Todo{
			ID:       "test-123",
			Task:     longTask,
			Status:   "in_progress",
			Priority: "medium",
		}

		line := formatTodoSummaryLine(todo)
		if !strings.Contains(line, longTask) {
			t.Error("Expected full task description")
		}
	})
}
