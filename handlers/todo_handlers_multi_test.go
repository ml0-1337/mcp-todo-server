package handlers

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	"strings"
	"testing"
)

func TestHandleTodoCreateMulti(t *testing.T) {
	// Create mocks
	mockManager := NewMockTodoManager()
	mockSearch := NewMockSearchEngine()
	mockStats := NewMockStatsEngine()
	mockTemplates := NewMockTemplateManager()

	// Create handlers
	handlers := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

	// Track created todos
	createdTodos := []*core.Todo{}
	todoCounter := 0

	// Mock CreateTodo to return incrementing IDs
	mockManager.CreateTodoFunc = func(task, priority, todoType string) (*core.Todo, error) {
		todoCounter++
		todo := &core.Todo{
			ID:       generatedID(task),
			Task:     task,
			Priority: priority,
			Type:     todoType,
			Status:   "in_progress",
		}
		createdTodos = append(createdTodos, todo)
		return todo, nil
	}

	// Mock UpdateTodo for setting parent_id
	updatedMetadata := make(map[string]map[string]string)
	mockManager.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
		if metadata != nil && metadata["parent_id"] != "" {
			updatedMetadata[id] = metadata
		}
		return nil
	}

	// Create request
	req := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"parent": map[string]interface{}{
				"task":     "API Development Project",
				"priority": "high",
				"type":     "multi-phase",
			},
			"children": []interface{}{
				map[string]interface{}{
					"task":     "Phase 1: Design API endpoints",
					"priority": "high",
					"type":     "phase",
				},
				map[string]interface{}{
					"task":     "Phase 2: Implement core endpoints",
					"priority": "high",
					"type":     "phase",
				},
				map[string]interface{}{
					"task":     "Phase 3: Write documentation",
					"priority": "medium",
					"type":     "phase",
				},
			},
		},
	}

	// Call handler
	result, err := handlers.HandleTodoCreateMulti(context.Background(), req.ToCallToolRequest())

	// Check no error
	if err != nil {
		t.Fatalf("HandleTodoCreateMulti() error = %v", err)
	}

	// Check result
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Check result content
	content := result.Content[0].(mcp.TextContent).Text

	// Verify output contains success message
	if !strings.Contains(content, "Created multi-phase project") {
		t.Error("Result should contain success message")
	}

	// Verify tree structure is shown
	if !strings.Contains(content, "Project Structure:") {
		t.Error("Result should show project structure")
	}

	// Check correct number of todos created
	if len(createdTodos) != 4 { // 1 parent + 3 children
		t.Errorf("Expected 4 todos created, got %d", len(createdTodos))
	}

	// Verify parent todo
	if createdTodos[0].Task != "API Development Project" {
		t.Errorf("First todo should be parent, got %s", createdTodos[0].Task)
	}
	if createdTodos[0].Type != "multi-phase" {
		t.Errorf("Parent type should be multi-phase, got %s", createdTodos[0].Type)
	}

	// Verify children have parent_id set
	parentID := createdTodos[0].ID
	for i := 1; i < 4; i++ {
		childID := createdTodos[i].ID
		if metadata, exists := updatedMetadata[childID]; exists {
			if metadata["parent_id"] != parentID {
				t.Errorf("Child %d should have parent_id=%s, got %s", i, parentID, metadata["parent_id"])
			}
		} else {
			t.Errorf("Child %d parent_id was not set", i)
		}
	}

	// Verify visualization includes all todos
	for _, todo := range createdTodos {
		if !strings.Contains(content, todo.ID) {
			t.Errorf("Result should include todo ID %s", todo.ID)
		}
		if !strings.Contains(content, todo.Task) {
			t.Errorf("Result should include todo task %s", todo.Task)
		}
	}

	// Check summary
	if !strings.Contains(content, "Successfully created 4 todos (1 parent, 3 children)") {
		t.Error("Result should show correct count summary")
	}
}

func TestHandleTodoCreateMulti_CreateParentError(t *testing.T) {
	// Create mocks
	mockManager := NewMockTodoManager()
	mockSearch := NewMockSearchEngine()
	mockStats := NewMockStatsEngine()
	mockTemplates := NewMockTemplateManager()

	// Create handlers
	handlers := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

	// Mock CreateTodo to fail
	mockManager.CreateTodoFunc = func(task, priority, todoType string) (*core.Todo, error) {
		return nil, fmt.Errorf("Database error")
	}

	// Create request
	req := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"parent": map[string]interface{}{
				"task": "Test Project",
			},
			"children": []interface{}{
				map[string]interface{}{
					"task": "Child 1",
				},
			},
		},
	}

	// Call handler
	result, err := handlers.HandleTodoCreateMulti(context.Background(), req.ToCallToolRequest())

	// Should not return error (errors are handled via result)
	if err != nil {
		t.Fatalf("HandleTodoCreateMulti() error = %v", err)
	}

	// Check error in result
	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	content := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(content, "failed to create parent todo") {
		t.Errorf("Result should contain parent creation error, got: %s", content)
	}
}

func TestHandleTodoCreateMulti_CreateChildError(t *testing.T) {
	// Create mocks
	mockManager := NewMockTodoManager()
	mockSearch := NewMockSearchEngine()
	mockStats := NewMockStatsEngine()
	mockTemplates := NewMockTemplateManager()

	// Create handlers
	handlers := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

	callCount := 0
	// Mock CreateTodo to succeed for parent, fail for child
	mockManager.CreateTodoFunc = func(task, priority, todoType string) (*core.Todo, error) {
		callCount++
		if callCount == 1 {
			// Parent creation succeeds
			return &core.Todo{
				ID:       "parent-123",
				Task:     task,
				Priority: priority,
				Type:     todoType,
				Status:   "in_progress",
			}, nil
		}
		// Child creation fails
		return nil, fmt.Errorf("Failed to create child")
	}

	// Create request
	req := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"parent": map[string]interface{}{
				"task": "Test Project",
			},
			"children": []interface{}{
				map[string]interface{}{
					"task": "Child 1",
				},
			},
		},
	}

	// Call handler
	result, err := handlers.HandleTodoCreateMulti(context.Background(), req.ToCallToolRequest())

	// Should not return error
	if err != nil {
		t.Fatalf("HandleTodoCreateMulti() error = %v", err)
	}

	// Check error in result
	content := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(content, "failed to create child 0") {
		t.Errorf("Result should contain child creation error, got: %s", content)
	}
}

// Helper to generate deterministic IDs from task names
func generatedID(task string) string {
	// Simple ID generation for testing
	id := strings.ToLower(task)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, ":", "")
	return id
}
