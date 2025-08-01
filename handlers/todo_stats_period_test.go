package handlers

import (
	"context"
	"testing"

	"github.com/user/mcp-todo-server/core"
)

// Test 1: Stats with period "all" should return stats for all todos
func TestHandleTodoStats_PeriodAll(t *testing.T) {
	// Setup mocks
	mockManager := NewMockTodoManager()
	mockSearch := NewMockSearchEngine()
	mockStats := NewMockStatsEngine()
	mockTemplates := NewMockTemplateManager()

	// Create stats with todos from various dates
	mockStats.GenerateTodoStatsForPeriodFunc = func(period string) (*core.TodoStats, error) {
		return &core.TodoStats{
			TotalTodos:      10,
			InProgressTodos: 4,
			CompletedTodos:  5,
			BlockedTodos:    1,
		}, nil
	}

	handler := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

	// Test with period "all"
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

	// Verify the mock was called with "all" period
	calls := mockStats.GetCalls()
	if len(calls) != 1 || calls[0].Method != "GenerateTodoStatsForPeriod" {
		t.Errorf("Expected GenerateTodoStatsForPeriod to be called once")
	}
	if len(calls[0].Args) != 1 || calls[0].Args[0] != "all" {
		t.Errorf("Expected period 'all' to be passed, got %v", calls[0].Args)
	}
}

// Test 2: Stats with period "week" should only include todos from last 7 days
func TestHandleTodoStats_PeriodWeek_ShouldFilterByDate(t *testing.T) {
	// Setup mocks
	mockManager := NewMockTodoManager()
	mockSearch := NewMockSearchEngine()
	mockStats := NewMockStatsEngine()
	mockTemplates := NewMockTemplateManager()

	// Mock should return filtered stats for week period
	mockStats.GenerateTodoStatsForPeriodFunc = func(period string) (*core.TodoStats, error) {
		if period == "week" {
			// Should only include todos from last 7 days
			return &core.TodoStats{
				TotalTodos:      1, // Only recent todo
				InProgressTodos: 1,
				CompletedTodos:  0,
				BlockedTodos:    0,
			}, nil
		}
		return &core.TodoStats{
			TotalTodos:      2, // All todos
			InProgressTodos: 1,
			CompletedTodos:  1,
			BlockedTodos:    0,
		}, nil
	}

	handler := NewTodoHandlersWithDependencies(mockManager, mockSearch, mockStats, mockTemplates)

	// Test with period "week"
	request := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"period": "week",
		},
	}

	result, err := handler.HandleTodoStats(context.Background(), request.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoStats returned error: %v", err)
	}

	if result == nil || len(result.Content) == 0 {
		t.Errorf("Expected result with content")
	}

	// Verify the correct method was called with correct parameter
	calls := mockStats.GetCalls()
	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "GenerateTodoStatsForPeriod" {
		t.Errorf("Expected GenerateTodoStatsForPeriod to be called, got %s", calls[0].Method)
	}
	if len(calls[0].Args) != 1 || calls[0].Args[0] != "week" {
		t.Errorf("Expected period 'week' to be passed, got %v", calls[0].Args)
	}
}

// Test 3: Stats with period "month" should only include todos from last 30 days
func TestHandleTodoStats_PeriodMonth(t *testing.T) {
	mockStats := NewMockStatsEngine()
	mockStats.GenerateTodoStatsForPeriodFunc = func(period string) (*core.TodoStats, error) {
		if period == "month" {
			return &core.TodoStats{
				TotalTodos:      5,
				InProgressTodos: 3,
				CompletedTodos:  2,
				BlockedTodos:    0,
			}, nil
		}
		return nil, nil
	}

	handler := NewTodoHandlersWithDependencies(NewMockTodoManager(), NewMockSearchEngine(), mockStats, NewMockTemplateManager())
	request := &MockCallToolRequest{Arguments: map[string]interface{}{"period": "month"}}

	_, err := handler.HandleTodoStats(context.Background(), request.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoStats returned error: %v", err)
	}

	calls := mockStats.GetCalls()
	if calls[0].Args[0] != "month" {
		t.Errorf("Expected period 'month', got %v", calls[0].Args)
	}
}

// Test 4: Stats with period "quarter" should only include todos from last 90 days
func TestHandleTodoStats_PeriodQuarter(t *testing.T) {
	mockStats := NewMockStatsEngine()
	mockStats.GenerateTodoStatsForPeriodFunc = func(period string) (*core.TodoStats, error) {
		if period == "quarter" {
			return &core.TodoStats{
				TotalTodos:      15,
				InProgressTodos: 8,
				CompletedTodos:  6,
				BlockedTodos:    1,
			}, nil
		}
		return nil, nil
	}

	handler := NewTodoHandlersWithDependencies(NewMockTodoManager(), NewMockSearchEngine(), mockStats, NewMockTemplateManager())
	request := &MockCallToolRequest{Arguments: map[string]interface{}{"period": "quarter"}}

	_, err := handler.HandleTodoStats(context.Background(), request.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoStats returned error: %v", err)
	}

	calls := mockStats.GetCalls()
	if calls[0].Args[0] != "quarter" {
		t.Errorf("Expected period 'quarter', got %v", calls[0].Args)
	}
}

// Test 5: Stats with period "year" should only include todos from last 365 days
func TestHandleTodoStats_PeriodYear(t *testing.T) {
	mockStats := NewMockStatsEngine()
	mockStats.GenerateTodoStatsForPeriodFunc = func(period string) (*core.TodoStats, error) {
		if period == "year" {
			return &core.TodoStats{
				TotalTodos:      50,
				InProgressTodos: 10,
				CompletedTodos:  35,
				BlockedTodos:    5,
			}, nil
		}
		return nil, nil
	}

	handler := NewTodoHandlersWithDependencies(NewMockTodoManager(), NewMockSearchEngine(), mockStats, NewMockTemplateManager())
	request := &MockCallToolRequest{Arguments: map[string]interface{}{"period": "year"}}

	_, err := handler.HandleTodoStats(context.Background(), request.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoStats returned error: %v", err)
	}

	calls := mockStats.GetCalls()
	if calls[0].Args[0] != "year" {
		t.Errorf("Expected period 'year', got %v", calls[0].Args)
	}
}

// Test 6: Stats with invalid period should default to "all"
func TestHandleTodoStats_InvalidPeriod(t *testing.T) {
	mockStats := NewMockStatsEngine()
	mockStats.GenerateTodoStatsForPeriodFunc = func(period string) (*core.TodoStats, error) {
		// The stats engine should handle invalid periods by returning all todos
		return &core.TodoStats{
			TotalTodos:      10,
			InProgressTodos: 4,
			CompletedTodos:  5,
			BlockedTodos:    1,
		}, nil
	}

	handler := NewTodoHandlersWithDependencies(NewMockTodoManager(), NewMockSearchEngine(), mockStats, NewMockTemplateManager())
	request := &MockCallToolRequest{Arguments: map[string]interface{}{"period": "invalid-period"}}

	result, err := handler.HandleTodoStats(context.Background(), request.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoStats returned error: %v", err)
	}

	if result == nil || len(result.Content) == 0 {
		t.Errorf("Expected result with content for invalid period")
	}

	// Handler should pass the invalid period to stats engine, which handles it
	calls := mockStats.GetCalls()
	if calls[0].Args[0] != "invalid-period" {
		t.Errorf("Expected invalid period to be passed to stats engine")
	}
}
