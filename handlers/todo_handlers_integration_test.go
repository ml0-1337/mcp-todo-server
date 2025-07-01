package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// Test HandleTodoCreate
func TestHandleTodoCreate(t *testing.T) {
	tests := []struct {
		name           string
		request        *MockCallToolRequest
		setupMocks     func(*MockTodoManager, *MockSearchEngine, *MockTemplateManager)
		expectError    bool
		expectedResult func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name: "successful create without template",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"task":     "Test todo task",
					"priority": "high",
					"type":     "feature",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tm.CreateTodoFunc = func(task, priority, todoType string) (*core.Todo, error) {
					return &core.Todo{
						ID:       "test-123",
						Task:     task,
						Priority: priority,
						Type:     todoType,
						Status:   "in_progress",
						Started:  time.Now(),
					}, nil
				}
				se.IndexTodoFunc = func(todo *core.Todo, content string) error {
					return nil
				}
			},
			expectError: false,
			expectedResult: func(t *testing.T, result *mcp.CallToolResult) {
				// Result should contain content
				if result == nil || len(result.Content) == 0 {
					t.Errorf("Expected result with content")
				}
			},
		},
		{
			name: "successful create with template",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"task":     "Fix bug",
					"priority": "high",
					"type":     "bug",
					"template": "bug-fix",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tmpl.CreateFromTemplateFunc = func(templateName, task, priority, todoType string) (*core.Todo, error) {
					return &core.Todo{
						ID:       "template-123",
						Task:     task,
						Priority: priority,
						Type:     todoType,
						Status:   "in_progress",
						Started:  time.Now(),
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "create with parent ID",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"task":      "Subtask",
					"priority":  "medium",
					"type":      "feature",
					"parent_id": "parent-123",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tm.CreateTodoFunc = func(task, priority, todoType string) (*core.Todo, error) {
					return &core.Todo{
						ID:       "child-123",
						Task:     task,
						Priority: priority,
						Type:     todoType,
						Status:   "in_progress",
						Started:  time.Now(),
					}, nil
				}
				tm.SaveTodoFunc = func(todo *core.Todo) error {
					// Verify parent_id is set
					if todo.ParentID != "parent-123" {
						t.Errorf("Expected parent_id=parent-123, got %v", todo.ParentID)
					}
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "missing task parameter",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"priority": "high",
					"type":     "feature",
				},
			},
			expectError: true,
		},
		{
			name: "invalid priority",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"task":     "Test task",
					"priority": "urgent",
					"type":     "feature",
				},
			},
			expectError: true,
		},
		{
			name: "create todo error",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"task":     "Test task",
					"priority": "high",
					"type":     "feature",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tm.CreateTodoFunc = func(task, priority, todoType string) (*core.Todo, error) {
					return nil, errors.New("database error")
				}
			},
			expectError: true,
		},
		{
			name: "index todo failure (non-fatal)",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"task":     "Test task",
					"priority": "high",
					"type":     "feature",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine, tmpl *MockTemplateManager) {
				tm.CreateTodoFunc = func(task, priority, todoType string) (*core.Todo, error) {
					return &core.Todo{
						ID:   "test-123",
						Task: task,
					}, nil
				}
				se.IndexTodoFunc = func(todo *core.Todo, content string) error {
					return errors.New("index error")
				}
			},
			expectError: false, // Index error is non-fatal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockManager := NewMockTodoManager()
			mockSearch := NewMockSearchEngine()
			mockStats := NewMockStatsEngine()
			mockTemplates := NewMockTemplateManager()

			if tt.setupMocks != nil {
				tt.setupMocks(mockManager, mockSearch, mockTemplates)
			}

			// Create handler
			handler := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

			// Execute
			result, err := handler.HandleTodoCreate(context.Background(), tt.request.ToCallToolRequest())

			// Verify
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
			}

			if tt.expectedResult != nil {
				tt.expectedResult(t, result)
			}
		})
	}
}

// Test HandleTodoRead
func TestHandleTodoRead(t *testing.T) {
	tests := []struct {
		name        string
		request     *MockCallToolRequest
		setupMocks  func(*MockTodoManager)
		expectError bool
	}{
		{
			name: "read single todo",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":     "test-123",
					"format": "full",
				},
			},
			setupMocks: func(tm *MockTodoManager) {
				tm.ReadTodoFunc = func(id string) (*core.Todo, error) {
					return &core.Todo{
						ID:       id,
						Task:     "Test Todo",
						Priority: "high",
						Status:   "in_progress",
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "list todos with filters",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"filter": map[string]interface{}{
						"status":   "in_progress",
						"priority": "high",
						"days":     float64(7),
					},
					"format": "list",
				},
			},
			setupMocks: func(tm *MockTodoManager) {
				tm.ListTodosFunc = func(status, priority string, days int) ([]*core.Todo, error) {
					return []*core.Todo{
						{ID: "todo-1", Task: "Task 1", Status: status, Priority: priority},
						{ID: "todo-2", Task: "Task 2", Status: status, Priority: priority},
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "todo not found",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id": "non-existent",
				},
			},
			setupMocks: func(tm *MockTodoManager) {
				tm.ReadTodoWithContentFunc = func(id string) (*core.Todo, string, error) {
					return nil, "", fmt.Errorf("todo not found: %s", id)
				}
			},
			expectError: false, // HandleError returns result, not error
		},
		{
			name: "invalid format",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":     "test-123",
					"format": "invalid",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockManager := NewMockTodoManager()
			mockSearch := NewMockSearchEngine()
			mockStats := NewMockStatsEngine()
			mockTemplates := NewMockTemplateManager()

			if tt.setupMocks != nil {
				tt.setupMocks(mockManager)
			}

			// Create handler
			handler := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

			// Execute
			result, err := handler.HandleTodoRead(context.Background(), tt.request.ToCallToolRequest())

			// Verify
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				// For "todo not found" case, check if result contains error
				if tt.name == "todo not found" && result != nil {
					content := result.Content[0].(mcp.TextContent).Text
					if !strings.Contains(content, "Todo not found") {
						t.Errorf("Expected 'Todo not found' error in result, got: %s", content)
					}
				}
			}
		})
	}
}

// Test HandleTodoUpdate
func TestHandleTodoUpdate(t *testing.T) {
	tests := []struct {
		name        string
		request     *MockCallToolRequest
		setupMocks  func(*MockTodoManager, *MockSearchEngine)
		expectError bool
		verifyCalls func(*testing.T, *MockTodoManager)
	}{
		{
			name: "update section content",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":        "test-123",
					"section":   "findings",
					"operation": "append",
					"content":   "New findings",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine) {
				tm.ReadTodoFunc = func(id string) (*core.Todo, error) {
					return &core.Todo{
						ID: id,
						Sections: map[string]*core.SectionDefinition{
							"findings": {
								Title:  "## Findings & Research",
								Schema: core.SchemaResearch,
							},
						},
					}, nil
				}
				tm.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
					return nil
				}
				tm.ReadTodoContentFunc = func(id string) (string, error) {
					return "Todo content", nil
				}
			},
			expectError: false,
			verifyCalls: func(t *testing.T, tm *MockTodoManager) {
				calls := tm.GetCalls()
				if len(calls) < 1 {
					t.Error("Expected UpdateTodo to be called")
				}
			},
		},
		{
			name: "update metadata",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id": "test-123",
					"metadata": map[string]interface{}{
						"status":       "completed",
						"priority":     "low",
						"current_test": "Test 5",
					},
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine) {
				tm.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
					// Verify metadata
					if metadata["status"] != "completed" {
						t.Errorf("Expected status=completed, got %v", metadata["status"])
					}
					if metadata["priority"] != "low" {
						t.Errorf("Expected priority=low, got %v", metadata["priority"])
					}
					if metadata["current_test"] != "Test 5" {
						t.Errorf("Expected current_test=Test 5, got %v", metadata["current_test"])
					}
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "missing ID",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"section": "findings",
				},
			},
			expectError: true,
		},
		{
			name: "invalid section",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":      "test-123",
					"section": "invalid",
					"content": "Some content",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine) {
				// Mock update to return error for invalid section
				tm.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
					return fmt.Errorf("invalid section: %s", section)
				}
			},
			expectError: true, // Returns Go error, not HandleError
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockManager := NewMockTodoManager()
			mockSearch := NewMockSearchEngine()
			mockStats := NewMockStatsEngine()
			mockTemplates := NewMockTemplateManager()

			if tt.setupMocks != nil {
				tt.setupMocks(mockManager, mockSearch)
			}

			// Create handler
			handler := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

			// Execute
			_, err := handler.HandleTodoUpdate(context.Background(), tt.request.ToCallToolRequest())

			// Verify
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
			}

			if tt.verifyCalls != nil {
				tt.verifyCalls(t, mockManager)
			}
		})
	}
}

// Test HandleTodoArchive
func TestHandleTodoArchive(t *testing.T) {
	tests := []struct {
		name        string
		request     *MockCallToolRequest
		setupMocks  func(*MockTodoManager, *MockSearchEngine)
		expectError bool
	}{
		{
			name: "archive with quarter",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":      "test-123",
					"quarter": "2025-Q1",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine) {
				tm.ReadTodoFunc = func(id string) (*core.Todo, error) {
					return &core.Todo{
						ID:      id,
						Task:    "Test Todo",
						Started: time.Now(),
					}, nil
				}
				tm.ArchiveTodoFunc = func(id, quarter string) error {
					return nil
				}
				se.DeleteTodoFunc = func(id string) error {
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "archive without quarter",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id": "test-123",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine) {
				tm.ReadTodoFunc = func(id string) (*core.Todo, error) {
					return &core.Todo{
						ID:      id,
						Task:    "Test Todo",
						Started: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
					}, nil
				}
				tm.ArchiveTodoFunc = func(id, quarter string) error {
					return nil
				}
			},
			expectError: false,
		},
		{
			name: "archive with read error (non-fatal)",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id": "test-123",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine) {
				tm.ReadTodoFunc = func(id string) (*core.Todo, error) {
					return nil, errors.New("read error")
				}
				tm.ArchiveTodoFunc = func(id, quarter string) error {
					return nil // Archive still succeeds
				}
			},
			expectError: false,
		},
		{
			name: "missing ID",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{},
			},
			expectError: true,
		},
		{
			name: "archive error",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id": "test-123",
				},
			},
			setupMocks: func(tm *MockTodoManager, se *MockSearchEngine) {
				tm.ArchiveTodoFunc = func(id, quarter string) error {
					return errors.New("archive failed")
				}
			},
			expectError: false, // HandleError returns result, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockManager := NewMockTodoManager()
			mockSearch := NewMockSearchEngine()
			mockStats := NewMockStatsEngine()
			mockTemplates := NewMockTemplateManager()

			if tt.setupMocks != nil {
				tt.setupMocks(mockManager, mockSearch)
			}

			// Create handler
			handler := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

			// Execute
			result, err := handler.HandleTodoArchive(context.Background(), tt.request.ToCallToolRequest())

			// Verify
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				// For "archive error" case, check if result contains error
				if tt.name == "archive error" && result != nil {
					content := result.Content[0].(mcp.TextContent).Text
					if !strings.Contains(content, "Archive operation failed") {
						t.Errorf("Expected archive operation failed error in result, got: %s", content)
					}
				}
			}
		})
	}
}

// Test HandleTodoSearch
func TestHandleTodoSearch(t *testing.T) {
	tests := []struct {
		name        string
		request     *MockCallToolRequest
		setupMocks  func(*MockSearchEngine)
		expectError bool
	}{
		{
			name: "search with filters",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"query": "test search",
					"filters": map[string]interface{}{
						"status":    "in_progress",
						"date_from": "2025-01-01",
						"date_to":   "2025-01-31",
					},
					"limit": float64(10),
				},
			},
			setupMocks: func(se *MockSearchEngine) {
				se.SearchTodosFunc = func(queryStr string, filters map[string]string, limit int) ([]core.SearchResult, error) {
					return []core.SearchResult{
						{
							ID:      "todo-1",
							Score:   0.95,
							Snippet: "Test <mark>search</mark> result",
						},
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "missing query",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"limit": float64(10),
				},
			},
			expectError: true,
		},
		{
			name: "search error",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"query": "test",
				},
			},
			setupMocks: func(se *MockSearchEngine) {
				se.SearchTodosFunc = func(queryStr string, filters map[string]string, limit int) ([]core.SearchResult, error) {
					return nil, errors.New("search engine error")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockManager := NewMockTodoManager()
			mockSearch := NewMockSearchEngine()
			mockStats := NewMockStatsEngine()
			mockTemplates := NewMockTemplateManager()

			if tt.setupMocks != nil {
				tt.setupMocks(mockSearch)
			}

			// Create handler
			handler := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

			// Execute
			_, err := handler.HandleTodoSearch(context.Background(), tt.request.ToCallToolRequest())

			// Verify
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
			}
		})
	}
}

// Test HandleTodoStats
func TestHandleTodoStats(t *testing.T) {
	mockManager := NewMockTodoManager()
	mockSearch := NewMockSearchEngine()
	mockStats := NewMockStatsEngine()
	mockTemplates := NewMockTemplateManager()

	mockStats.GenerateTodoStatsFunc = func() (*core.TodoStats, error) {
		return &core.TodoStats{
			TotalTodos:      10,
			InProgressTodos: 4,
			CompletedTodos:  5,
			BlockedTodos:    1,
		}, nil
	}

	handler := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

	request := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"period": "all",
		},
	}

	result, err := handler.HandleTodoStats(context.Background(), request.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoStats returned error: %v", err)
	}

	if result == nil || len(result.Content) == 0 {
		t.Errorf("Expected result with content")
	}
}

// Test HandleTodoClean
func TestHandleTodoClean(t *testing.T) {
	tests := []struct {
		name        string
		request     *MockCallToolRequest
		setupMocks  func(*MockTodoManager)
		expectError bool
	}{
		{
			name: "archive old todos",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"operation": "archive_old",
					"days":      float64(90),
				},
			},
			setupMocks: func(tm *MockTodoManager) {
				tm.ArchiveOldTodosFunc = func(days int) (int, error) {
					return 5, nil
				}
			},
			expectError: false,
		},
		{
			name: "find duplicates",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"operation": "find_duplicates",
				},
			},
			setupMocks: func(tm *MockTodoManager) {
				tm.FindDuplicateTodosFunc = func() ([][]string, error) {
					return [][]string{
						{"todo-1", "todo-2"},
						{"todo-3", "todo-4"},
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "unknown operation",
			request: &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"operation": "unknown",
				},
			},
			expectError: false, // HandleError returns result, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockManager := NewMockTodoManager()
			mockSearch := NewMockSearchEngine()
			mockStats := NewMockStatsEngine()
			mockTemplates := NewMockTemplateManager()

			if tt.setupMocks != nil {
				tt.setupMocks(mockManager)
			}

			// Create handler
			handler := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

			// Execute
			result, err := handler.HandleTodoClean(context.Background(), tt.request.ToCallToolRequest())

			// Verify
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				// For "unknown operation" case, check if result contains error
				if tt.name == "unknown operation" && result != nil {
					content := result.Content[0].(mcp.TextContent).Text
					if !strings.Contains(content, "unknown operation") {
						t.Errorf("Expected 'unknown operation' error in result, got: %s", content)
					}
				}
			}
		})
	}
}

// Test error handling patterns
func TestHandlerErrorPatterns(t *testing.T) {
	t.Run("parameter extraction errors are handled", func(t *testing.T) {
		mockManager := NewMockTodoManager()
		mockSearch := NewMockSearchEngine()
		mockStats := NewMockStatsEngine()
		mockTemplates := NewMockTemplateManager()

		handler := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

		// Test with invalid request
		request := &MockCallToolRequest{
			Arguments: map[string]interface{}{
				// Missing required 'task' parameter
				"priority": "invalid",
			},
		}

		_, err := handler.HandleTodoCreate(context.Background(), request.ToCallToolRequest())

		// Should return error, not panic
		if err == nil {
			t.Error("Expected error for invalid parameters")
		}
	})
}
