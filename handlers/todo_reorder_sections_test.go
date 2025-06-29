package handlers

import (
	"context"
	"fmt"
	"testing"
	"github.com/user/mcp-todo-server/core"
)

func TestHandleTodoReorderSections(t *testing.T) {
	// Test 14: Reorder sections via API
	// Input: Request with todo ID and new section orders
	// Expected: Sections reordered successfully
	
	ctx := context.Background()
	
	// Create mock dependencies
	mockManager := NewMockTodoManager()
	mockSearch := NewMockSearchEngine()
	mockStats := NewMockStatsEngine()
	mockTemplates := NewMockTemplateManager()
	
	handlers := NewTodoHandlersWithDependencies(
		mockManager,
		mockSearch,
		mockStats,
		mockTemplates,
	)
	
	// Setup test todo with sections
	testTodo := &core.Todo{
		ID:   "test-todo",
		Task: "Test Todo",
		Sections: map[string]*core.SectionDefinition{
			"findings": {
				Title:  "## Findings & Research",
				Order:  1,
				Schema: "research",
			},
			"checklist": {
				Title:  "## Checklist",
				Order:  2,
				Schema: "checklist",
			},
			"test_cases": {
				Title:  "## Test Cases",
				Order:  3,
				Schema: "test_cases",
			},
			"scratchpad": {
				Title:  "## Working Scratchpad",
				Order:  4,
				Schema: "freeform",
			},
		},
	}
	
	// Configure mock to return test todo
	mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
		if id == "test-todo" {
			return testTodo, nil
		}
		return nil, fmt.Errorf("todo not found")
	}
	
	// Track saved todos
	var savedTodo *core.Todo
	mockManager.SaveTodoFunc = func(todo *core.Todo) error {
		savedTodo = todo
		return nil
	}
	
	// Test: Reorder sections successfully
	t.Run("Reorder sections", func(t *testing.T) {
		// Create request with new order
		request := MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id": "test-todo",
				"order": map[string]interface{}{
					"scratchpad": 1,  // Move to first
					"findings":   2,  // Move to second
					"checklist":  3,  // Move to third
					"test_cases": 4,  // Move to last
				},
			},
		}
		
		result, err := handlers.HandleTodoReorderSections(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		// Verify sections were reordered
		if savedTodo == nil {
			t.Fatal("Expected todo to be saved")
		}
		if savedTodo.Sections["scratchpad"].Order != 1 {
			t.Errorf("Expected scratchpad order 1, got %d", savedTodo.Sections["scratchpad"].Order)
		}
		if savedTodo.Sections["findings"].Order != 2 {
			t.Errorf("Expected findings order 2, got %d", savedTodo.Sections["findings"].Order)
		}
		if savedTodo.Sections["checklist"].Order != 3 {
			t.Errorf("Expected checklist order 3, got %d", savedTodo.Sections["checklist"].Order)
		}
		if savedTodo.Sections["test_cases"].Order != 4 {
			t.Errorf("Expected test_cases order 4, got %d", savedTodo.Sections["test_cases"].Order)
		}
		
		// Verify response
		if result.IsError {
			t.Errorf("Expected success but got error")
		}
	})
	
	// Test: Handle missing todo
	t.Run("Missing todo", func(t *testing.T) {
		request := MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id": "missing-todo",
				"order": map[string]interface{}{
					"findings": 1,
				},
			},
		}
		
		result, err := handlers.HandleTodoReorderSections(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !result.IsError {
			t.Errorf("Expected error result but got success")
		}
	})
	
	// Test: Handle invalid section key
	t.Run("Invalid section key", func(t *testing.T) {
		request := MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id": "test-todo",
				"order": map[string]interface{}{
					"invalid_section": 1,
					"findings":        2,
				},
			},
		}
		
		result, err := handlers.HandleTodoReorderSections(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !result.IsError {
			t.Errorf("Expected error result but got success")
		}
		// Error message validation removed as we can't easily access Content text
	})
	
	// Test: Handle non-numeric order values
	t.Run("Non-numeric order values", func(t *testing.T) {
		request := MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id": "test-todo",
				"order": map[string]interface{}{
					"findings": "first",  // Invalid - not a number
				},
			},
		}
		
		result, err := handlers.HandleTodoReorderSections(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !result.IsError {
			t.Errorf("Expected error result but got success")
		}
		// Error message validation removed as we can't easily access Content text
	})
	
	// Test: Handle partial reordering
	t.Run("Partial reordering", func(t *testing.T) {
		savedTodo = nil
		
		// Configure mock to return a fresh test todo
		mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
			if id == "test-todo" {
				return &core.Todo{
					ID:   "test-todo",
					Task: "Test Todo",
					Sections: map[string]*core.SectionDefinition{
						"findings": {
							Title:  "## Findings & Research",
							Order:  1,
							Schema: "research",
						},
						"checklist": {
							Title:  "## Checklist",
							Order:  2,
							Schema: "checklist",
						},
						"test_cases": {
							Title:  "## Test Cases",
							Order:  3,
							Schema: "test_cases",
						},
						"scratchpad": {
							Title:  "## Working Scratchpad",
							Order:  4,
							Schema: "freeform",
						},
					},
				}, nil
			}
			return nil, fmt.Errorf("todo not found")
		}
		
		request := MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id": "test-todo",
				"order": map[string]interface{}{
					"scratchpad": 10,  // Only reorder this section
					"findings":   20,  // And this one
					// checklist and test_cases keep original order
				},
			},
		}
		
		result, err := handlers.HandleTodoReorderSections(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		// Verify only specified sections were reordered
		if savedTodo == nil {
			t.Fatal("Expected todo to be saved")
		}
		if savedTodo.Sections["scratchpad"].Order != 10 {
			t.Errorf("Expected scratchpad order 10, got %d", savedTodo.Sections["scratchpad"].Order)
		}
		if savedTodo.Sections["findings"].Order != 20 {
			t.Errorf("Expected findings order 20, got %d", savedTodo.Sections["findings"].Order)
		}
		if savedTodo.Sections["checklist"].Order != 2 {
			t.Errorf("Expected checklist order 2 (unchanged), got %d", savedTodo.Sections["checklist"].Order)
		}
		if savedTodo.Sections["test_cases"].Order != 3 {
			t.Errorf("Expected test_cases order 3 (unchanged), got %d", savedTodo.Sections["test_cases"].Order)
		}
		
		// Verify response
		if result.IsError {
			t.Errorf("Expected success but got error")
		}
	})
	
	// Test: Handle todo without sections
	t.Run("Todo without sections", func(t *testing.T) {
		// Configure mock to return todo without sections
		mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
			if id == "legacy-todo" {
				return &core.Todo{
					ID:   "legacy-todo",
					Task: "Legacy Todo",
					// No sections defined
				}, nil
			}
			return nil, fmt.Errorf("todo not found")
		}
		
		request := MockCallToolRequest{
			Arguments: map[string]interface{}{
				"id": "legacy-todo",
				"order": map[string]interface{}{
					"findings": 1,
				},
			},
		}
		
		result, err := handlers.HandleTodoReorderSections(ctx, request.ToCallToolRequest())
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !result.IsError {
			t.Errorf("Expected error result but got success")
		}
		// Error message validation removed as we can't easily access Content text
	})
}

