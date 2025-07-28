package adapters

import (
	"testing"
	
	"github.com/user/mcp-todo-server/internal/application"
	"github.com/user/mcp-todo-server/internal/infrastructure/persistence/filesystem"
)

func TestTodoManagerAdapter_CreateAndRead(t *testing.T) {
	// Test that the adapter correctly bridges between old and new interfaces
	
	// Arrange
	tmpDir := t.TempDir()
	
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Act - Create todo
	todo, err := adapter.CreateTodo("Test adapter task", "high", "feature")
	
	// Assert - Create succeeded
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}
	
	if todo.Task != "Test adapter task" {
		t.Errorf("Task mismatch: expected 'Test adapter task', got '%s'", todo.Task)
	}
	
	// Act - Read todo
	retrieved, err := adapter.ReadTodo(todo.ID)
	
	// Assert - Read succeeded
	if err != nil {
		t.Fatalf("Failed to read todo: %v", err)
	}
	
	if retrieved.ID != todo.ID {
		t.Errorf("ID mismatch: expected '%s', got '%s'", todo.ID, retrieved.ID)
	}
	
	if retrieved.Task != todo.Task {
		t.Errorf("Task mismatch: expected '%s', got '%s'", todo.Task, retrieved.Task)
	}
}

func TestTodoManagerAdapter_ReadTodoWithContent(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Create a todo
	todo, err := adapter.CreateTodo("Test task with content", "medium", "bug")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}
	
	// Act - Read todo with content
	retrieved, content, err := adapter.ReadTodoWithContent(todo.ID)
	
	// Assert
	if err != nil {
		t.Fatalf("Failed to read todo with content: %v", err)
	}
	
	if retrieved.ID != todo.ID {
		t.Errorf("ID mismatch: expected '%s', got '%s'", todo.ID, retrieved.ID)
	}
	
	if content == "" {
		t.Error("Expected non-empty content")
	}
	
	// Test with non-existent todo
	_, _, err = adapter.ReadTodoWithContent("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent todo")
	}
}

func TestTodoManagerAdapter_UpdateTodo(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Create a todo
	todo, err := adapter.CreateTodo("Test update", "low", "refactor")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}
	
	// Test status update via metadata
	err = adapter.UpdateTodo(todo.ID, "", "", "", map[string]string{"status": "completed"})
	if err != nil {
		t.Errorf("Failed to update todo status: %v", err)
	}
	
	// Test section update (should fail as not implemented)
	err = adapter.UpdateTodo(todo.ID, "findings", "append", "test content", nil)
	if err == nil {
		t.Error("Expected error for section update")
	}
	if err.Error() != "section updates not yet implemented in adapter" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestTodoManagerAdapter_SaveTodo(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Create a todo to modify
	todo, err := adapter.CreateTodo("Test save", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}
	
	// Modify the todo
	todo.Task = "Modified task"
	todo.Priority = "medium"
	
	// Act - Save modified todo
	err = adapter.SaveTodo(todo)
	
	// Assert
	if err != nil {
		t.Errorf("Failed to save todo: %v", err)
	}
	
	// Verify changes were saved
	retrieved, err := adapter.ReadTodo(todo.ID)
	if err != nil {
		t.Fatalf("Failed to read saved todo: %v", err)
	}
	
	if retrieved.Task != "Modified task" {
		t.Errorf("Task not updated: expected 'Modified task', got '%s'", retrieved.Task)
	}
	
	if retrieved.Priority != "medium" {
		t.Errorf("Priority not updated: expected 'medium', got '%s'", retrieved.Priority)
	}
}

func TestTodoManagerAdapter_ListTodos(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Create multiple todos
	_, _ = adapter.CreateTodo("High priority task", "high", "feature")
	todo2, _ := adapter.CreateTodo("Medium priority task", "medium", "bug")
	_, _ = adapter.CreateTodo("Low priority task", "low", "refactor")
	
	// Update one to completed
	adapter.UpdateTodo(todo2.ID, "", "", "", map[string]string{"status": "completed"})
	
	tests := []struct {
		name     string
		status   string
		priority string
		days     int
		wantMin  int
	}{
		{
			name:     "all todos",
			status:   "",
			priority: "",
			days:     0,
			wantMin:  3,
		},
		{
			name:     "high priority only",
			status:   "",
			priority: "high",
			days:     0,
			wantMin:  1,
		},
		{
			name:     "completed only",
			status:   "completed",
			priority: "",
			days:     0,
			wantMin:  1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todos, err := adapter.ListTodos(tt.status, tt.priority, tt.days)
			if err != nil {
				t.Errorf("Failed to list todos: %v", err)
			}
			
			if len(todos) < tt.wantMin {
				t.Errorf("Expected at least %d todos, got %d", tt.wantMin, len(todos))
			}
		})
	}
}

func TestTodoManagerAdapter_ReadTodoContent(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Create a todo
	todo, err := adapter.CreateTodo("Test content reading", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}
	
	// Act
	content, err := adapter.ReadTodoContent(todo.ID)
	
	// Assert
	if err != nil {
		t.Errorf("Failed to read todo content: %v", err)
	}
	
	if content == "" {
		t.Error("Expected non-empty content")
	}
	
	// Test with non-existent todo
	_, err = adapter.ReadTodoContent("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent todo")
	}
}

func TestTodoManagerAdapter_ArchiveTodo(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Create and complete a todo
	todo, err := adapter.CreateTodo("Test archive", "medium", "bug")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}
	
	// Update to completed
	err = adapter.UpdateTodo(todo.ID, "", "", "", map[string]string{"status": "completed"})
	if err != nil {
		t.Fatalf("Failed to update todo status: %v", err)
	}
	
	// Act - Archive the todo
	err = adapter.ArchiveTodo(todo.ID)
	
	// Assert
	if err != nil {
		t.Errorf("Failed to archive todo: %v", err)
	}
	
	// Try to read archived todo - should fail
	_, err = adapter.ReadTodo(todo.ID)
	if err == nil {
		t.Error("Expected error reading archived todo")
	}
}

func TestTodoManagerAdapter_ArchiveOldTodos(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Create todos
	todo1, _ := adapter.CreateTodo("Old completed task", "high", "feature")
	_, _ = adapter.CreateTodo("Recent task", "medium", "bug")
	_, _ = adapter.CreateTodo("In progress task", "low", "refactor")
	
	// Mark todo1 as completed
	adapter.UpdateTodo(todo1.ID, "", "", "", map[string]string{"status": "completed"})
	
	// Act - Archive todos older than 0 days (should archive all completed)
	count, err := adapter.ArchiveOldTodos(0)
	
	// Assert
	if err != nil {
		t.Errorf("Failed to archive old todos: %v", err)
	}
	
	if count < 1 {
		t.Errorf("Expected at least 1 archived todo, got %d", count)
	}
}

func TestTodoManagerAdapter_FindDuplicateTodos(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Create todos with duplicate tasks
	todo1, _ := adapter.CreateTodo("Duplicate Task", "high", "feature")
	todo2, _ := adapter.CreateTodo("duplicate task", "medium", "bug") // Different case
	todo3, _ := adapter.CreateTodo("  Duplicate Task  ", "low", "refactor") // Extra spaces
	_, _ = adapter.CreateTodo("Unique Task", "high", "feature")
	
	// Act
	duplicates, err := adapter.FindDuplicateTodos()
	
	// Assert
	if err != nil {
		t.Errorf("Failed to find duplicates: %v", err)
	}
	
	if len(duplicates) < 1 {
		t.Error("Expected at least one duplicate group")
	}
	
	// Check that duplicate group contains expected todos
	foundDuplicate := false
	for _, group := range duplicates {
		if len(group) >= 3 {
			foundDuplicate = true
			// Verify the group contains our duplicate IDs
			ids := map[string]bool{
				todo1.ID: false,
				todo2.ID: false,
				todo3.ID: false,
			}
			for _, id := range group {
				if _, ok := ids[id]; ok {
					ids[id] = true
				}
			}
		}
	}
	
	if !foundDuplicate {
		t.Error("Expected to find duplicate group with 3 todos")
	}
}

func TestTodoManagerAdapter_GetBasePath(t *testing.T) {
	// Arrange
	expectedPath := "/test/path"
	repo := filesystem.NewTodoRepository(expectedPath)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, expectedPath)
	
	// Act
	basePath := adapter.GetBasePath()
	
	// Assert
	if basePath != expectedPath {
		t.Errorf("GetBasePath() = %q, want %q", basePath, expectedPath)
	}
}

func TestTodoManagerAdapter_ErrorHandling(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Test ReadTodo with non-existent ID
	_, err := adapter.ReadTodo("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent todo")
	}
	
	// Test UpdateTodo with non-existent ID
	err = adapter.UpdateTodo("non-existent-id", "", "", "", map[string]string{"status": "completed"})
	if err == nil {
		t.Error("Expected error updating non-existent todo")
	}
}

func TestTodoManagerAdapter_Conversions(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	repo := filesystem.NewTodoRepository(tmpDir)
	service := application.NewTodoService(repo)
	adapter := NewTodoManagerAdapter(service, repo, tmpDir)
	
	// Test domainToCoreTodo with nil
	result := adapter.domainToCoreTodo(nil)
	if result != nil {
		t.Error("Expected nil for nil input")
	}
	
	// Test coreToDomainTodo with nil
	domainResult := adapter.coreToDomainTodo(nil)
	if domainResult != nil {
		t.Error("Expected nil for nil input")
	}
	
	// Test with actual todo to ensure conversions work
	todo, err := adapter.CreateTodo("Conversion test", "high", "feature")
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}
	
	// Verify fields are properly converted
	if todo.Task != "Conversion test" {
		t.Errorf("Task conversion failed: got %q", todo.Task)
	}
	if todo.Priority != "high" {
		t.Errorf("Priority conversion failed: got %q", todo.Priority)
	}
	if todo.Type != "feature" {
		t.Errorf("Type conversion failed: got %q", todo.Type)
	}
}