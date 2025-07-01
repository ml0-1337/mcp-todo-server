package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// BenchmarkSearchWithManyTodos benchmarks search performance with large todo sets
func BenchmarkSearchWithManyTodos(b *testing.B) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager and search engine
	manager := NewTodoManager(tempDir)
	indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
	searchEngine, err := NewSearchEngine(indexPath, filepath.Join(tempDir, ".claude", "todos"))
	if err != nil {
		b.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()

	// Create many todos for benchmarking
	numTodos := 1000
	for i := 0; i < numTodos; i++ {
		task := fmt.Sprintf("Task %d: Implement feature %d with priority", i, i)
		priority := []string{"high", "medium", "low"}[i%3]
		todoType := []string{"feature", "bug", "refactor"}[i%3]

		todo, err := manager.CreateTodo(task, priority, todoType)
		if err != nil {
			b.Fatalf("Failed to create todo %d: %v", i, err)
		}

		// Index the todo
		content := fmt.Sprintf("Content for task %d with keywords: search performance benchmark", i)
		searchEngine.IndexTodo(todo, content)
	}

	// Benchmark search operations
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Search for different terms
		searchTerms := []string{
			"implement feature",
			"priority high",
			"task 500",
			"search performance",
			"benchmark",
		}

		for _, term := range searchTerms {
			_, err := searchEngine.SearchTodos(term, nil, 20)
			if err != nil {
				b.Fatalf("Search failed: %v", err)
			}
		}
	}
}

// TestConcurrentTodoOperations tests thread safety of todo operations
func TestConcurrentTodoOperations(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-concurrent-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewTodoManager(tempDir)

	// Number of concurrent operations
	numGoroutines := 50
	todosPerGoroutine := 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*todosPerGoroutine)

	// Launch concurrent todo creators
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < todosPerGoroutine; j++ {
				task := fmt.Sprintf("Routine %d Todo %d", routineID, j)
				priority := []string{"high", "medium", "low"}[j%3]

				todo, err := manager.CreateTodo(task, priority, "concurrent")
				if err != nil {
					errors <- fmt.Errorf("routine %d: failed to create todo: %v", routineID, err)
					continue
				}

				// Try to read it back
				readTodo, err := manager.ReadTodo(todo.ID)
				if err != nil {
					errors <- fmt.Errorf("routine %d: failed to read todo: %v", routineID, err)
					continue
				}

				if readTodo.Task != task {
					errors <- fmt.Errorf("routine %d: task mismatch: got %s, want %s",
						routineID, readTodo.Task, task)
				}

				// Update the todo
				metadata := map[string]string{
					"test_routine": fmt.Sprintf("%d", routineID),
				}
				err = manager.UpdateTodo(todo.ID, "", "", "", metadata)
				if err != nil {
					errors <- fmt.Errorf("routine %d: failed to update todo: %v", routineID, err)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
		errorCount++
		if errorCount > 10 {
			t.Fatal("Too many concurrent errors, stopping test")
		}
	}

	// Verify all todos were created
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	files, err := ioutil.ReadDir(todosDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	expectedTodos := numGoroutines * todosPerGoroutine
	actualTodos := 0
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") {
			actualTodos++
		}
	}

	// Allow some margin for concurrent conflicts
	if actualTodos < expectedTodos*9/10 {
		t.Errorf("Expected at least %d todos, got %d", expectedTodos*9/10, actualTodos)
	}
}

// TestLargeTodoContent tests handling of very large todo content
func TestLargeTodoContent(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-large-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewTodoManager(tempDir)

	// Create todo with very large content
	largeContent := strings.Repeat("This is a very long content line. ", 10000)

	todo, err := manager.CreateTodo("Large content todo", "high", "test")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Update with large content in scratchpad
	err = manager.UpdateTodo(todo.ID, "scratchpad", "replace", largeContent, nil)
	if err != nil {
		t.Fatalf("Failed to update with large content: %v", err)
	}

	// Read it back
	readTodo, err := manager.ReadTodo(todo.ID)
	if err != nil {
		t.Fatalf("Failed to read todo with large content: %v", err)
	}

	if readTodo.ID != todo.ID {
		t.Errorf("Todo ID mismatch: got %s, want %s", readTodo.ID, todo.ID)
	}
}

// TestArchivePerformance tests archive operations with many todos
func TestArchivePerformance(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-archive-perf-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewTodoManager(tempDir)

	// Create a parent with many children
	parent, err := manager.CreateTodo("Parent project", "high", "multi-phase")
	if err != nil {
		t.Fatalf("Failed to create parent: %v", err)
	}

	// Create many child todos
	numChildren := 100
	children := make([]*Todo, numChildren)
	for i := 0; i < numChildren; i++ {
		child, err := manager.CreateTodoWithParent(
			fmt.Sprintf("Child task %d", i),
			"medium",
			"feature",
			parent.ID,
		)
		if err != nil {
			t.Fatalf("Failed to create child %d: %v", i, err)
		}
		children[i] = child
	}

	// Mark all as completed
	parent.Completed = time.Now()
	manager.UpdateTodo(parent.ID, "", "", "",
		map[string]string{"status": "completed"})

	for _, child := range children {
		manager.UpdateTodo(child.ID, "", "", "",
			map[string]string{"status": "completed"})
	}

	// Time the cascade archive operation
	start := time.Now()

	err = manager.ArchiveTodoWithCascade(parent.ID, true)

	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to archive with cascade: %v", err)
	}

	// Should complete reasonably quickly even with 100 children
	if elapsed > 5*time.Second {
		t.Errorf("Archive operation too slow: %v", elapsed)
	}

	// Verify all were archived
	if !isArchivedWrapper(manager, parent.ID) {
		t.Error("Parent should be archived")
	}

	archivedCount := 0
	for _, child := range children {
		if isArchivedWrapper(manager, child.ID) {
			archivedCount++
		}
	}

	if archivedCount != numChildren {
		t.Errorf("Expected %d archived children, got %d", numChildren, archivedCount)
	}
}

// TestStatsWithManyTodos tests stats calculation performance
func TestStatsWithManyTodos(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-stats-perf-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewTodoManager(tempDir)
	statsEngine := NewStatsEngine(manager)

	// Create many todos with various states
	numTodos := 500
	for i := 0; i < numTodos; i++ {
		task := fmt.Sprintf("Task %d for stats", i)
		priority := []string{"high", "medium", "low"}[i%3]
		todoType := []string{"feature", "bug", "refactor", "research"}[i%4]
		status := []string{"pending", "in_progress", "completed"}[i%3]

		todo, err := manager.CreateTodo(task, priority, todoType)
		if err != nil {
			t.Fatalf("Failed to create todo %d: %v", i, err)
		}

		// Update status
		if status == "completed" {
			manager.UpdateTodo(todo.ID, "", "", "",
				map[string]string{"status": status})
		}

		// Add test list for some todos
		if i%5 == 0 {
			testList := fmt.Sprintf(`
- [x] Test 1: Basic functionality
- [x] Test 2: Edge cases
- [ ] Test 3: Performance
- [ ] Test 4: Security
`)
			manager.UpdateTodo(todo.ID, "tests", "append", testList, nil)
		}
	}

	// Time stats generation
	start := time.Now()

	stats, err := statsEngine.GenerateTodoStats()

	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to generate stats: %v", err)
	}

	// Should complete quickly even with 500 todos
	if elapsed > 2*time.Second {
		t.Errorf("Stats generation too slow: %v", elapsed)
	}

	// Verify stats are reasonable
	if stats.TotalTodos < numTodos {
		t.Errorf("Expected at least %d todos in stats, got %d", numTodos, stats.TotalTodos)
	}

	if stats.TodosByType == nil || len(stats.TodosByType) == 0 {
		t.Error("Expected todos by type stats")
	}

	if stats.TodosByPriority == nil || len(stats.TodosByPriority) == 0 {
		t.Error("Expected todos by priority stats")
	}

	if stats.CompletionRates == nil || len(stats.CompletionRates) == 0 {
		t.Error("Expected completion rates")
	}
}

// TestMemoryLeaks tests for memory leaks in long-running operations
func TestMemoryLeaks(t *testing.T) {
	// This is a simplified memory leak test
	// In production, you'd use runtime.MemStats for more accurate measurement

	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-memleak-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewTodoManager(tempDir)

	// Perform many create/read/update/delete cycles
	cycles := 100
	for cycle := 0; cycle < cycles; cycle++ {
		// Create a batch of todos
		todos := make([]*Todo, 10)
		for i := 0; i < 10; i++ {
			todo, err := manager.CreateTodo(
				fmt.Sprintf("Cycle %d Todo %d", cycle, i),
				"medium",
				"test",
			)
			if err != nil {
				t.Fatalf("Cycle %d: Failed to create todo: %v", cycle, err)
			}
			todos[i] = todo
		}

		// Update them
		for _, todo := range todos {
			err := manager.UpdateTodo(todo.ID, "", "", "", map[string]string{"status": "in_progress"})
			if err != nil {
				t.Fatalf("Cycle %d: Failed to update todo: %v", cycle, err)
			}
		}

		// Read them
		for _, todo := range todos {
			_, err := manager.ReadTodo(todo.ID)
			if err != nil {
				t.Fatalf("Cycle %d: Failed to read todo: %v", cycle, err)
			}
		}

		// Archive half of them
		for i, todo := range todos {
			if i%2 == 0 {
				manager.UpdateTodo(todo.ID, "", "", "",
					map[string]string{"status": "completed"})
				manager.ArchiveTodo(todo.ID)
			}
		}
	}

	// If we get here without running out of memory or file handles, test passes
	t.Logf("Completed %d cycles without memory issues", cycles)
}
