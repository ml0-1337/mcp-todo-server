package handlers

import (
	"context"
	"github.com/user/mcp-todo-server/core"
	"testing"
)

// Test 10: Integration test - update via handlers works without timestamps
func TestHandleTodoUpdate_AcceptsContentWithoutTimestamps(t *testing.T) {
	// Create mock managers
	mockManager := NewMockTodoManager()

	// Setup todo with results section
	testTodo := &core.Todo{
		ID:       "test-results-todo",
		Task:     "Test todo with results section",
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
		Sections: map[string]*core.SectionDefinition{
			"test_results": {
				Title:    "## Test Results Log",
				Order:    1,
				Schema:   core.SchemaResults,
				Required: false,
			},
		},
	}

	// Mock ReadTodo to return our test todo
	mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
		if id == "test-results-todo" {
			return testTodo, nil
		}
		return nil, nil
	}

	// Track if UpdateTodo was called
	updateCalled := false
	mockManager.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
		updateCalled = true
		// The content is passed through to core.UpdateTodo which will add timestamps
		return nil
	}

	mockSearch := &MockSearchEngine{}
	mockStats := &MockStatsEngine{}
	mockTemplates := &MockTemplateManager{}

	// Create handlers
	handlers := NewTodoHandlersWithDependencies(
		mockManager,
		mockSearch,
		mockStats,
		mockTemplates,
	)

	// Create update request without timestamp
	request := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"id":        "test-results-todo",
			"section":   "test_results",
			"operation": "append",
			"content":   "Test completed successfully",
		},
	}

	// Call handler
	result, err := handlers.HandleTodoUpdate(context.Background(), request.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoUpdate() error = %v", err)
	}

	// Verify success
	if result.IsError {
		t.Error("Expected success but got error")
	}

	// Verify UpdateTodo was called
	if !updateCalled {
		t.Error("Expected UpdateTodo to be called")
	}

	// The handlers layer passes content through to core.UpdateTodo
	// which handles the automatic timestamping
}
