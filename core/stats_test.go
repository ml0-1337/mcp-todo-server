package core

import (
	"testing"
	"time"
	"io/ioutil"
	"os"
)

// Test 22: todo_stats should calculate metrics accurately
func TestTodoStatsCalculation(t *testing.T) {
	// Create temp directory for test
	tempDir, err := ioutil.TempDir("", "stats-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create various todos for testing
	createTestTodos(t, manager)

	// Create stats engine
	stats := NewStatsEngine(manager)

	t.Run("Calculate completion rates by type", func(t *testing.T) {
		rates, err := stats.CalculateCompletionRatesByType()
		if err != nil {
			t.Fatalf("Failed to calculate completion rates by type: %v", err)
		}

		// Verify rates
		if rates["feature"] != 50.0 { // 2 completed out of 4
			t.Errorf("Expected feature completion rate 50%%, got %.1f%%", rates["feature"])
		}
		if rates["bug"] != 66.7 { // 2 completed out of 3, rounded
			t.Errorf("Expected bug completion rate 66.7%%, got %.1f%%", rates["bug"])
		}
		if rates["refactor"] != 100.0 { // 1 completed out of 1
			t.Errorf("Expected refactor completion rate 100%%, got %.1f%%", rates["refactor"])
		}
	})

	t.Run("Calculate completion rates by priority", func(t *testing.T) {
		rates, err := stats.CalculateCompletionRatesByPriority()
		if err != nil {
			t.Fatalf("Failed to calculate completion rates by priority: %v", err)
		}

		// Verify rates
		if rates["high"] != 75.0 { // 3 completed out of 4
			t.Errorf("Expected high priority completion rate 75%%, got %.1f%%", rates["high"])
		}
		if rates["medium"] != 50.0 { // 1 completed out of 2
			t.Errorf("Expected medium priority completion rate 50%%, got %.1f%%", rates["medium"])
		}
		if rates["low"] != 50.0 { // 1 completed out of 2
			t.Errorf("Expected low priority completion rate 50%%, got %.1f%%", rates["low"])
		}
	})

	t.Run("Calculate average time to complete", func(t *testing.T) {
		avgTime, err := stats.CalculateAverageCompletionTime()
		if err != nil {
			t.Fatalf("Failed to calculate average completion time: %v", err)
		}

		// Should be around 24 hours (based on test data)
		expectedHours := 24.0
		tolerance := 1.0 // Allow 1 hour tolerance
		
		if avgTime.Hours() < expectedHours-tolerance || avgTime.Hours() > expectedHours+tolerance {
			t.Errorf("Expected average completion time around %v hours, got %v", expectedHours, avgTime.Hours())
		}
	})

	t.Run("Calculate test coverage from Test List", func(t *testing.T) {
		// Create a todo with test list
		todoWithTests, err := manager.CreateTodo("Feature with tests", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo with tests: %v", err)
		}

		// Update with test list content in the Test Cases section
		testContent := `
- [x] Test 1: Basic functionality works
- [x] Test 2: Edge case handled
- [ ] Test 3: Error handling
- [x] Test 4: Performance test
- [ ] Test 5: Integration test
`
		err = manager.UpdateTodo(todoWithTests.ID, "tests", "replace", testContent, nil)
		if err != nil {
			t.Fatalf("Failed to update todo with test list: %v", err)
		}

		coverage, err := stats.CalculateTestCoverage(todoWithTests.ID)
		if err != nil {
			t.Fatalf("Failed to calculate test coverage: %v", err)
		}

		// Should be 60% (3 out of 5 tests checked)
		expectedCoverage := 60.0
		if coverage != expectedCoverage {
			t.Errorf("Expected test coverage %.1f%%, got %.1f%%", expectedCoverage, coverage)
		}
	})

	t.Run("Generate overall statistics", func(t *testing.T) {
		todoStats, err := stats.GenerateTodoStats()
		if err != nil {
			t.Fatalf("Failed to generate overall stats: %v", err)
		}

		// Verify basic counts (including the todoWithTests from previous test)
		if todoStats.TotalTodos != 9 { // 8 from createTestTodos + 1 from test coverage test
			t.Errorf("Expected 9 total todos, got %d", todoStats.TotalTodos)
		}
		if todoStats.CompletedTodos != 5 { // 5 marked as completed
			t.Errorf("Expected 5 completed todos, got %d", todoStats.CompletedTodos)
		}
		if todoStats.InProgressTodos != 4 { // 3 + 1 from test coverage test
			t.Errorf("Expected 4 in-progress todos, got %d", todoStats.InProgressTodos)
		}

		// Verify type distribution
		if todoStats.TodosByType["feature"] != 5 { // 4 + 1 from test coverage test
			t.Errorf("Expected 5 feature todos, got %d", todoStats.TodosByType["feature"])
		}
		if todoStats.TodosByType["bug"] != 3 {
			t.Errorf("Expected 3 bug todos, got %d", todoStats.TodosByType["bug"])
		}

		// Verify priority distribution
		if todoStats.TodosByPriority["high"] != 5 { // 4 + 1 from test coverage test
			t.Errorf("Expected 5 high priority todos, got %d", todoStats.TodosByPriority["high"])
		}
	})

	t.Run("Handle empty todo collection", func(t *testing.T) {
		// Create empty directory
		emptyDir, err := ioutil.TempDir("", "empty-stats-test-*")
		if err != nil {
			t.Fatalf("Failed to create empty directory: %v", err)
		}
		defer os.RemoveAll(emptyDir)

		emptyManager := NewTodoManager(emptyDir)
		emptyStats := NewStatsEngine(emptyManager)

		// Should return zero values, not errors
		rates, err := emptyStats.CalculateCompletionRatesByType()
		if err != nil {
			t.Fatalf("Should handle empty collection gracefully: %v", err)
		}
		if len(rates) != 0 {
			t.Error("Empty collection should return empty rates map")
		}

		avgTime, err := emptyStats.CalculateAverageCompletionTime()
		if err != nil {
			t.Fatalf("Should handle empty collection gracefully: %v", err)
		}
		if avgTime != 0 {
			t.Error("Empty collection should return zero average time")
		}

		todoStats, err := emptyStats.GenerateTodoStats()
		if err != nil {
			t.Fatalf("Should handle empty collection gracefully: %v", err)
		}
		if todoStats.TotalTodos != 0 {
			t.Error("Empty collection should have zero todos")
		}
	})
}

// Helper function to create test todos with various states
func createTestTodos(t *testing.T, manager *TodoManager) {
	// Create completed todos with different types and priorities
	todo1, _ := manager.CreateTodo("Implement login feature", "high", "feature")
	todo2, _ := manager.CreateTodo("Fix memory leak", "high", "bug")
	todo3, _ := manager.CreateTodo("Add user profile", "medium", "feature")
	todo4, _ := manager.CreateTodo("Fix CSS alignment", "low", "bug")
	todo5, _ := manager.CreateTodo("Refactor database layer", "high", "refactor")

	// Mark some as completed (with timestamps 24 hours ago)
	completedTime := time.Now().Add(-24 * time.Hour)
	markAsCompleted(t, manager, todo1.ID, completedTime)
	markAsCompleted(t, manager, todo2.ID, completedTime)
	markAsCompleted(t, manager, todo3.ID, completedTime.Add(-48*time.Hour)) // Older
	markAsCompleted(t, manager, todo4.ID, completedTime)
	markAsCompleted(t, manager, todo5.ID, completedTime)

	// Create in-progress todos
	manager.CreateTodo("Add dark mode", "high", "feature")
	manager.CreateTodo("Fix validation bug", "medium", "bug")
	manager.CreateTodo("Implement search", "low", "feature")
}

// Helper to mark todo as completed with specific timestamp
func markAsCompleted(t *testing.T, manager *TodoManager, id string, completedTime time.Time) {
	// Calculate started time (24 hours before completed)
	startedTime := completedTime.Add(-24 * time.Hour)
	
	// Update using the UpdateTodo method with metadata
	metadata := map[string]string{
		"status": "completed",
		"completed": completedTime.Format("2006-01-02 15:04:05"),
		"started": startedTime.Format("2006-01-02 15:04:05"),
	}
	
	err := manager.UpdateTodo(id, "", "", "", metadata)
	if err != nil {
		t.Fatalf("Failed to update todo: %v", err)
	}
}

