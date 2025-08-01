package core

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

// Test period filtering in StatsEngine
func TestStatsEngine_GenerateTodoStatsForPeriod(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "stats-period-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)
	statsEngine := NewStatsEngine(manager)

	// Create todos with different dates
	now := time.Now()

	// Create an old todo (2 months ago)
	oldTodo, err := manager.CreateTodo("Old task from 2 months ago", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create old todo: %v", err)
	}
	oldTodo.Started = now.AddDate(0, -2, 0)
	oldTodo.Completed = now.AddDate(0, -2, 5)
	oldTodo.Status = "completed"
	err = manager.SaveTodo(oldTodo)
	if err != nil {
		t.Fatalf("Failed to save old todo: %v", err)
	}

	// Create a recent todo (3 days ago)
	recentTodo, err := manager.CreateTodo("Recent task from 3 days ago", "medium", "bug")
	if err != nil {
		t.Fatalf("Failed to create recent todo: %v", err)
	}
	recentTodo.Started = now.AddDate(0, 0, -3)
	recentTodo.Status = "in_progress"
	err = manager.SaveTodo(recentTodo)
	if err != nil {
		t.Fatalf("Failed to save recent todo: %v", err)
	}

	// Create a very recent todo (today)
	todayTodo, err := manager.CreateTodo("Task started today", "low", "refactor")
	if err != nil {
		t.Fatalf("Failed to create today todo: %v", err)
	}
	todayTodo.Started = now
	todayTodo.Status = "in_progress"
	err = manager.SaveTodo(todayTodo)
	if err != nil {
		t.Fatalf("Failed to save today todo: %v", err)
	}

	t.Run("Period all returns all todos", func(t *testing.T) {
		stats, err := statsEngine.GenerateTodoStatsForPeriod("all")
		if err != nil {
			t.Fatalf("Failed to generate stats: %v", err)
		}

		if stats.TotalTodos != 3 {
			t.Errorf("Expected 3 total todos for 'all' period, got %d", stats.TotalTodos)
		}
		if stats.CompletedTodos != 1 {
			t.Errorf("Expected 1 completed todo, got %d", stats.CompletedTodos)
		}
		if stats.InProgressTodos != 2 {
			t.Errorf("Expected 2 in-progress todos, got %d", stats.InProgressTodos)
		}
	})

	t.Run("Period week returns only recent todos", func(t *testing.T) {
		stats, err := statsEngine.GenerateTodoStatsForPeriod("week")
		if err != nil {
			t.Fatalf("Failed to generate stats: %v", err)
		}

		if stats.TotalTodos != 2 {
			t.Errorf("Expected 2 total todos for 'week' period (3 days + today), got %d", stats.TotalTodos)
		}
		if stats.CompletedTodos != 0 {
			t.Errorf("Expected 0 completed todos (old one excluded), got %d", stats.CompletedTodos)
		}
		if stats.InProgressTodos != 2 {
			t.Errorf("Expected 2 in-progress todos, got %d", stats.InProgressTodos)
		}
	})

	t.Run("Period month excludes old todos", func(t *testing.T) {
		stats, err := statsEngine.GenerateTodoStatsForPeriod("month")
		if err != nil {
			t.Fatalf("Failed to generate stats: %v", err)
		}

		if stats.TotalTodos != 2 {
			t.Errorf("Expected 2 total todos for 'month' period (excludes 2-month old), got %d", stats.TotalTodos)
		}
		if stats.CompletedTodos != 0 {
			t.Errorf("Expected 0 completed todos (old one excluded), got %d", stats.CompletedTodos)
		}
	})

	t.Run("Invalid period defaults to all", func(t *testing.T) {
		stats, err := statsEngine.GenerateTodoStatsForPeriod("invalid-period")
		if err != nil {
			t.Fatalf("Failed to generate stats: %v", err)
		}

		if stats.TotalTodos != 3 {
			t.Errorf("Expected 3 total todos for invalid period (defaults to all), got %d", stats.TotalTodos)
		}
	})

	t.Run("Empty period defaults to all", func(t *testing.T) {
		stats, err := statsEngine.GenerateTodoStatsForPeriod("")
		if err != nil {
			t.Fatalf("Failed to generate stats: %v", err)
		}

		if stats.TotalTodos != 3 {
			t.Errorf("Expected 3 total todos for empty period (defaults to all), got %d", stats.TotalTodos)
		}
	})

	t.Run("Todo at exact boundary is included", func(t *testing.T) {
		// Create a todo at exactly 7 days ago
		boundaryTodo, err := manager.CreateTodo("Task at 7-day boundary", "high", "bug")
		if err != nil {
			t.Fatalf("Failed to create boundary todo: %v", err)
		}
		boundaryTodo.Started = now.AddDate(0, 0, -7).Add(time.Second) // Just after boundary
		boundaryTodo.Status = "in_progress"
		err = manager.SaveTodo(boundaryTodo)
		if err != nil {
			t.Fatalf("Failed to save boundary todo: %v", err)
		}

		stats, err := statsEngine.GenerateTodoStatsForPeriod("week")
		if err != nil {
			t.Fatalf("Failed to generate stats: %v", err)
		}

		// Should include: recent (3 days), today, and boundary (just after 7 days) = 3
		if stats.TotalTodos != 3 {
			t.Errorf("Expected 3 todos including boundary todo, got %d", stats.TotalTodos)
		}
	})

	t.Run("Empty todo list returns zero counts", func(t *testing.T) {
		// Create new empty manager
		emptyDir, err := ioutil.TempDir("", "empty-stats-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(emptyDir)

		emptyManager := NewTodoManager(emptyDir)
		emptyStats := NewStatsEngine(emptyManager)

		// Test with various periods - all should return zero
		periods := []string{"all", "week", "month", "quarter", "year"}
		for _, period := range periods {
			stats, err := emptyStats.GenerateTodoStatsForPeriod(period)
			if err != nil {
				t.Fatalf("Failed to generate stats for period %s: %v", period, err)
			}

			if stats.TotalTodos != 0 {
				t.Errorf("Expected 0 todos for empty list with period %s, got %d", period, stats.TotalTodos)
			}
			if stats.CompletedTodos != 0 {
				t.Errorf("Expected 0 completed todos for period %s, got %d", period, stats.CompletedTodos)
			}
			if stats.InProgressTodos != 0 {
				t.Errorf("Expected 0 in-progress todos for period %s, got %d", period, stats.InProgressTodos)
			}
		}
	})
}
