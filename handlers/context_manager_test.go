package handlers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/user/mcp-todo-server/core"
	ctxkeys "github.com/user/mcp-todo-server/internal/context"
)

// TestNewContextualTodoManager tests the creation of a new contextual todo manager
func TestNewContextualTodoManager(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "contextual-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewContextualTodoManager(tempDir)

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.defaultPath != tempDir {
		t.Errorf("Expected defaultPath %s, got %s", tempDir, manager.defaultPath)
	}

	if manager.managers == nil {
		t.Fatal("Expected managers map to be initialized")
	}

	if len(manager.managers) != 0 {
		t.Errorf("Expected empty managers map, got %d entries", len(manager.managers))
	}
}

// TestGetManagerForContext tests getting the appropriate manager for a context
func TestGetManagerForContext(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "contextual-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create working directories
	workDir1 := filepath.Join(tempDir, "project1")
	workDir2 := filepath.Join(tempDir, "project2")
	os.MkdirAll(workDir1, 0755)
	os.MkdirAll(workDir2, 0755)

	manager := NewContextualTodoManager(tempDir)

	tests := []struct {
		name        string
		ctx         context.Context
		expectedPath string
	}{
		{
			name:        "Default context",
			ctx:         context.Background(),
			expectedPath: tempDir,
		},
		{
			name:        "Context with working directory 1",
			ctx:         context.WithValue(context.Background(), ctxkeys.WorkingDirectoryKey, workDir1),
			expectedPath: workDir1,
		},
		{
			name:        "Context with working directory 2",
			ctx:         context.WithValue(context.Background(), ctxkeys.WorkingDirectoryKey, workDir2),
			expectedPath: workDir2,
		},
		{
			name:        "Context with empty working directory",
			ctx:         context.WithValue(context.Background(), ctxkeys.WorkingDirectoryKey, ""),
			expectedPath: tempDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoManager := manager.GetManagerForContext(tt.ctx)
			if todoManager == nil {
				t.Fatal("Expected non-nil todo manager")
			}

			// Verify the manager is cached
			todoManager2 := manager.GetManagerForContext(tt.ctx)
			if todoManager != todoManager2 {
				t.Error("Expected same manager instance from cache")
			}
		})
	}

	// Verify we have different managers for different contexts
	manager1 := manager.GetManagerForContext(context.WithValue(context.Background(), ctxkeys.WorkingDirectoryKey, workDir1))
	manager2 := manager.GetManagerForContext(context.WithValue(context.Background(), ctxkeys.WorkingDirectoryKey, workDir2))
	if manager1 == manager2 {
		t.Error("Expected different managers for different working directories")
	}
}

// TestGetManagerForContextConcurrency tests concurrent access to GetManagerForContext
func TestGetManagerForContextConcurrency(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "contextual-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewContextualTodoManager(tempDir)

	// Create multiple working directories
	numDirs := 10
	workDirs := make([]string, numDirs)
	for i := 0; i < numDirs; i++ {
		workDirs[i] = filepath.Join(tempDir, fmt.Sprintf("project%d", i))
		os.MkdirAll(workDirs[i], 0755)
	}

	// Run concurrent access test
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			
			// Each goroutine accesses a random working directory
			workDir := workDirs[idx%numDirs]
			ctx := context.WithValue(context.Background(), ctxkeys.WorkingDirectoryKey, workDir)
			
			// Get manager multiple times
			for j := 0; j < 10; j++ {
				todoManager := manager.GetManagerForContext(ctx)
				if todoManager == nil {
					t.Errorf("Goroutine %d: Expected non-nil manager", idx)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify we have the correct number of cached managers
	manager.mu.RLock()
	numCached := len(manager.managers)
	manager.mu.RUnlock()

	// We should have at most numDirs + 1 managers (including default)
	if numCached > numDirs+1 {
		t.Errorf("Expected at most %d cached managers, got %d", numDirs+1, numCached)
	}
}

// TestNewContextualTodoManagerWrapper tests the wrapper creation
func TestNewContextualTodoManagerWrapper(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "wrapper-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	wrapper := NewContextualTodoManagerWrapper(tempDir)

	if wrapper == nil {
		t.Fatal("Expected non-nil wrapper")
	}

	if wrapper.ContextualTodoManager == nil {
		t.Fatal("Expected non-nil ContextualTodoManager")
	}

	if wrapper.defaultManager == nil {
		t.Fatal("Expected non-nil defaultManager")
	}
}

// TestContextualTodoManagerWrapperMethods tests all wrapper methods
func TestContextualTodoManagerWrapperMethods(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "wrapper-methods-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	wrapper := NewContextualTodoManagerWrapper(tempDir)

	// Test CreateTodo
	t.Run("CreateTodo", func(t *testing.T) {
		todo, err := wrapper.CreateTodo("Test task", "high", "feature")
		if err != nil {
			t.Fatalf("Failed to create todo: %v", err)
		}
		if todo == nil {
			t.Fatal("Expected non-nil todo")
		}
		if todo.Task != "Test task" {
			t.Errorf("Expected task 'Test task', got %s", todo.Task)
		}
	})

	// Test ReadTodo
	t.Run("ReadTodo", func(t *testing.T) {
		// Create a todo first
		todo, _ := wrapper.CreateTodo("Read test task", "medium", "bug")
		
		// Read it back
		readTodo, err := wrapper.ReadTodo(todo.ID)
		if err != nil {
			t.Fatalf("Failed to read todo: %v", err)
		}
		if readTodo.ID != todo.ID {
			t.Errorf("Expected todo ID %s, got %s", todo.ID, readTodo.ID)
		}
	})

	// Test ReadTodoWithContent
	t.Run("ReadTodoWithContent", func(t *testing.T) {
		// Create a todo first
		todo, _ := wrapper.CreateTodo("Content test task", "low", "research")
		
		// Read it back with content
		readTodo, content, err := wrapper.ReadTodoWithContent(todo.ID)
		if err != nil {
			t.Fatalf("Failed to read todo with content: %v", err)
		}
		if readTodo.ID != todo.ID {
			t.Errorf("Expected todo ID %s, got %s", todo.ID, readTodo.ID)
		}
		if content == "" {
			t.Error("Expected non-empty content")
		}
	})

	// Test UpdateTodo
	t.Run("UpdateTodo", func(t *testing.T) {
		// Create a todo first
		todo, _ := wrapper.CreateTodo("Update test task", "high", "feature")
		
		// Update it
		err := wrapper.UpdateTodo(todo.ID, "findings", "append", "Test findings", nil)
		if err != nil {
			t.Fatalf("Failed to update todo: %v", err)
		}
		
		// Verify update
		_, content, _ := wrapper.ReadTodoWithContent(todo.ID)
		if !strings.Contains(content, "Test findings") {
			t.Error("Expected content to contain 'Test findings'")
		}
	})

	// Test SaveTodo
	t.Run("SaveTodo", func(t *testing.T) {
		todo := &core.Todo{
			ID:       "test-save-todo",
			Task:     "Save test task",
			Started:  time.Now(),
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
		}
		
		err := wrapper.SaveTodo(todo)
		if err != nil {
			t.Fatalf("Failed to save todo: %v", err)
		}
		
		// Verify it was saved
		readTodo, err := wrapper.ReadTodo(todo.ID)
		if err != nil {
			t.Fatalf("Failed to read saved todo: %v", err)
		}
		if readTodo.Task != todo.Task {
			t.Errorf("Expected task %s, got %s", todo.Task, readTodo.Task)
		}
	})

	// Test ListTodos
	t.Run("ListTodos", func(t *testing.T) {
		// Create multiple todos
		wrapper.CreateTodo("List test 1", "high", "feature")
		wrapper.CreateTodo("List test 2", "medium", "bug")
		wrapper.CreateTodo("List test 3", "low", "research")
		
		// List all todos
		todos, err := wrapper.ListTodos("", "", 0)
		if err != nil {
			t.Fatalf("Failed to list todos: %v", err)
		}
		if len(todos) < 3 {
			t.Errorf("Expected at least 3 todos, got %d", len(todos))
		}
	})

	// Test ReadTodoContent
	t.Run("ReadTodoContent", func(t *testing.T) {
		// Create a todo
		todo, _ := wrapper.CreateTodo("Content read test", "high", "feature")
		
		// Read its content
		content, err := wrapper.ReadTodoContent(todo.ID)
		if err != nil {
			t.Fatalf("Failed to read todo content: %v", err)
		}
		if content == "" {
			t.Error("Expected non-empty content")
		}
	})

	// Test ArchiveTodo
	t.Run("ArchiveTodo", func(t *testing.T) {
		// Create a todo
		todo, _ := wrapper.CreateTodo("Archive test", "high", "feature")
		
		// Mark it as completed
		wrapper.UpdateTodo(todo.ID, "", "", "", map[string]string{
			"status": "completed",
			"completed": time.Now().Format("2006-01-02 15:04:05"),
		})
		
		// Archive it
		err := wrapper.ArchiveTodo(todo.ID, "")
		if err != nil {
			t.Fatalf("Failed to archive todo: %v", err)
		}
		
		// Verify it's archived (should not be in active list)
		todos, _ := wrapper.ListTodos("", "", 0)
		for _, activeTodo := range todos {
			if activeTodo.ID == todo.ID {
				t.Error("Archived todo should not appear in active list")
			}
		}
	})

	// Test ArchiveOldTodos
	t.Run("ArchiveOldTodos", func(t *testing.T) {
		// This is a more complex test that would require setting up old todos
		// For now, just verify the method doesn't error
		count, err := wrapper.ArchiveOldTodos(30)
		if err != nil {
			t.Fatalf("Failed to archive old todos: %v", err)
		}
		if count < 0 {
			t.Errorf("Expected non-negative count, got %d", count)
		}
	})

	// Test FindDuplicateTodos
	t.Run("FindDuplicateTodos", func(t *testing.T) {
		// Create duplicate todos
		wrapper.CreateTodo("Duplicate task", "high", "feature")
		wrapper.CreateTodo("Duplicate task", "high", "feature")
		
		// Find duplicates
		duplicates, err := wrapper.FindDuplicateTodos()
		if err != nil {
			t.Fatalf("Failed to find duplicate todos: %v", err)
		}
		// Should have at least one set of duplicates
		if len(duplicates) < 1 {
			t.Error("Expected to find duplicates")
		}
	})

	// Test GetBasePath
	t.Run("GetBasePath", func(t *testing.T) {
		basePath := wrapper.GetBasePath()
		if basePath != tempDir {
			t.Errorf("Expected base path %s, got %s", tempDir, basePath)
		}
	})
}

// TestCreateTodoWithContext tests the context-aware todo creation
func TestCreateTodoWithContext(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "create-context-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a working directory
	workDir := filepath.Join(tempDir, "project")
	os.MkdirAll(workDir, 0755)

	// Create handlers with contextual manager
	wrapper := NewContextualTodoManagerWrapper(tempDir)
	handlers := &TodoHandlers{
		manager: wrapper,
	}

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "With default context",
			ctx:  context.Background(),
		},
		{
			name: "With working directory context",
			ctx:  context.WithValue(context.Background(), ctxkeys.WorkingDirectoryKey, workDir),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo, err := handlers.CreateTodoWithContext(tt.ctx, "Context test task", "high", "feature")
			if err != nil {
				t.Fatalf("Failed to create todo with context: %v", err)
			}
			if todo == nil {
				t.Fatal("Expected non-nil todo")
			}
			if todo.Task != "Context test task" {
				t.Errorf("Expected task 'Context test task', got %s", todo.Task)
			}
		})
	}
}

// TestCreateTodoWithContextFallback tests fallback behavior
func TestCreateTodoWithContextFallback(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "fallback-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create handlers with regular manager (not contextual)
	regularManager := core.NewTodoManager(tempDir)
	handlers := &TodoHandlers{
		manager: regularManager,
	}

	// Should fall back to regular manager
	todo, err := handlers.CreateTodoWithContext(context.Background(), "Fallback test", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}
	if todo == nil {
		t.Fatal("Expected non-nil todo")
	}
	if todo.Task != "Fallback test" {
		t.Errorf("Expected task 'Fallback test', got %s", todo.Task)
	}
}

