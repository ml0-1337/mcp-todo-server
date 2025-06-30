package core

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// TestNewTodoLinker tests the TodoLinker constructor
func TestNewTodoLinker(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "todo-linker-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	// Create linker
	linker := NewTodoLinker(manager)

	// Verify linker is created
	if linker == nil {
		t.Fatal("Expected non-nil linker")
	}

	// Verify manager is set
	if linker.manager != manager {
		t.Error("Expected linker to have the provided manager")
	}
}

// TestLinkTodos tests linking todos with parent-child relationship
func TestLinkTodos(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "link-todos-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager and linker
	manager := NewTodoManager(tempDir)
	linker := NewTodoLinker(manager)

	tests := []struct {
		name        string
		setup       func() (parentID, childID string)
		linkType    string
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful parent-child link",
			setup: func() (string, string) {
				parent, _ := manager.CreateTodo("Parent task", "high", "multi-phase")
				child, _ := manager.CreateTodo("Child task", "medium", "phase")
				return parent.ID, child.ID
			},
			linkType:    "parent-child",
			expectError: false,
		},
		{
			name: "parent todo not found",
			setup: func() (string, string) {
				child, _ := manager.CreateTodo("Child task", "medium", "phase")
				return "nonexistent-parent", child.ID
			},
			linkType:    "parent-child",
			expectError: true,
			errorMsg:    "parent todo not found",
		},
		{
			name: "child todo not found",
			setup: func() (string, string) {
				parent, _ := manager.CreateTodo("Parent task", "high", "multi-phase")
				return parent.ID, "nonexistent-child"
			},
			linkType:    "parent-child",
			expectError: true,
			errorMsg:    "child todo not found",
		},
		{
			name: "unsupported link type",
			setup: func() (string, string) {
				parent, _ := manager.CreateTodo("Task 1", "high", "feature")
				child, _ := manager.CreateTodo("Task 2", "medium", "feature")
				return parent.ID, child.ID
			},
			linkType:    "blocks",
			expectError: true,
			errorMsg:    "unsupported link type: blocks",
		},
		{
			name: "empty parent ID",
			setup: func() (string, string) {
				child, _ := manager.CreateTodo("Child task", "medium", "phase")
				return "", child.ID
			},
			linkType:    "parent-child",
			expectError: true,
			errorMsg:    "parent todo not found",
		},
		{
			name: "empty child ID",
			setup: func() (string, string) {
				parent, _ := manager.CreateTodo("Parent task", "high", "multi-phase")
				return parent.ID, ""
			},
			linkType:    "parent-child",
			expectError: true,
			errorMsg:    "child todo not found",
		},
		{
			name: "link to self",
			setup: func() (string, string) {
				todo, _ := manager.CreateTodo("Self task", "high", "feature")
				return todo.ID, todo.ID
			},
			linkType:    "parent-child",
			expectError: false, // The current implementation doesn't prevent self-linking
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			parentID, childID := tt.setup()

			// Link todos
			err := linker.LinkTodos(parentID, childID, tt.linkType)

			// Check error
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				// Verify link was created (for parent-child)
				if tt.linkType == "parent-child" && childID != "" {
					// Read child todo to verify parent_id was set
					child, err := manager.ReadTodo(childID)
					if err != nil {
						t.Fatalf("Failed to read child todo: %v", err)
					}
					if child.ParentID != parentID {
						t.Errorf("Expected child parent_id to be %s, got %s", parentID, child.ParentID)
					}
				}
			}
		})
	}
}

// TestLinkTodosUpdateBehavior tests the update behavior when linking todos
func TestLinkTodosUpdateBehavior(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "link-update-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager and linker
	manager := NewTodoManager(tempDir)
	linker := NewTodoLinker(manager)

	// Create parent and child todos
	parent, err := manager.CreateTodo("Parent task", "high", "multi-phase")
	if err != nil {
		t.Fatalf("Failed to create parent: %v", err)
	}

	child, err := manager.CreateTodo("Child task", "medium", "phase")
	if err != nil {
		t.Fatalf("Failed to create child: %v", err)
	}

	// Add some content to child
	err = manager.UpdateTodo(child.ID, "findings", "append", "Initial findings", nil)
	if err != nil {
		t.Fatalf("Failed to update child: %v", err)
	}

	// Link todos
	err = linker.LinkTodos(parent.ID, child.ID, "parent-child")
	if err != nil {
		t.Fatalf("Failed to link todos: %v", err)
	}

	// Read child to verify
	updatedChild, content, err := manager.ReadTodoWithContent(child.ID)
	if err != nil {
		t.Fatalf("Failed to read child with content: %v", err)
	}

	// Verify parent_id is set
	if updatedChild.ParentID != parent.ID {
		t.Errorf("Expected parent_id %s, got %s", parent.ID, updatedChild.ParentID)
	}

	// Verify existing content is preserved
	if !strings.Contains(content, "Initial findings") {
		t.Error("Expected existing content to be preserved after linking")
	}
}

// TestLinkTodosWithMultipleChildren tests linking multiple children to same parent
func TestLinkTodosWithMultipleChildren(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "multi-link-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager and linker
	manager := NewTodoManager(tempDir)
	linker := NewTodoLinker(manager)

	// Create parent
	parent, err := manager.CreateTodo("Multi-phase project", "high", "multi-phase")
	if err != nil {
		t.Fatalf("Failed to create parent: %v", err)
	}

	// Create and link multiple children
	childIDs := []string{}
	for i := 1; i <= 3; i++ {
		child, err := manager.CreateTodo(
			"Phase "+string(rune('0'+i)), 
			"medium", 
			"phase",
		)
		if err != nil {
			t.Fatalf("Failed to create child %d: %v", i, err)
		}

		err = linker.LinkTodos(parent.ID, child.ID, "parent-child")
		if err != nil {
			t.Fatalf("Failed to link child %d: %v", i, err)
		}

		childIDs = append(childIDs, child.ID)
	}

	// Verify all children have correct parent_id
	for i, childID := range childIDs {
		child, err := manager.ReadTodo(childID)
		if err != nil {
			t.Fatalf("Failed to read child %d: %v", i+1, err)
		}
		if child.ParentID != parent.ID {
			t.Errorf("Child %d: expected parent_id %s, got %s", i+1, parent.ID, child.ParentID)
		}
	}

	// Use GetChildren to verify relationships
	children, err := manager.GetChildren(parent.ID)
	if err != nil {
		t.Fatalf("Failed to get children: %v", err)
	}
	if len(children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(children))
	}
}

// TestLinkTodosErrorRecovery tests error recovery scenarios
func TestLinkTodosErrorRecovery(t *testing.T) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "link-error-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager and linker
	manager := NewTodoManager(tempDir)
	linker := NewTodoLinker(manager)

	t.Run("linking after parent is deleted", func(t *testing.T) {
		// Create parent and child
		parent, _ := manager.CreateTodo("Parent", "high", "multi-phase")
		child, _ := manager.CreateTodo("Child", "medium", "phase")

		// Delete parent file
		os.Remove(tempDir + "/" + parent.ID + ".md")

		// Try to link
		err := linker.LinkTodos(parent.ID, child.ID, "parent-child")
		if err == nil {
			t.Error("Expected error when parent is deleted")
		}
		if !strings.Contains(err.Error(), "parent todo not found") {
			t.Errorf("Expected 'parent todo not found' error, got: %v", err)
		}
	})

	t.Run("linking after child is deleted", func(t *testing.T) {
		// Create parent and child
		parent, _ := manager.CreateTodo("Parent", "high", "multi-phase")
		child, _ := manager.CreateTodo("Child", "medium", "phase")

		// Delete child file
		os.Remove(tempDir + "/" + child.ID + ".md")

		// Try to link
		err := linker.LinkTodos(parent.ID, child.ID, "parent-child")
		if err == nil {
			t.Error("Expected error when child is deleted")
		}
		if !strings.Contains(err.Error(), "child todo not found") {
			t.Errorf("Expected 'child todo not found' error, got: %v", err)
		}
	})
}

// BenchmarkLinkTodos benchmarks the LinkTodos operation
func BenchmarkLinkTodos(b *testing.B) {
	// Create temp directory
	tempDir, err := ioutil.TempDir("", "link-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager and linker
	manager := NewTodoManager(tempDir)
	linker := NewTodoLinker(manager)

	// Pre-create todos
	parent, _ := manager.CreateTodo("Parent", "high", "multi-phase")
	children := make([]*Todo, b.N)
	for i := 0; i < b.N; i++ {
		child, _ := manager.CreateTodo("Child", "medium", "phase")
		children[i] = child
	}

	b.ResetTimer()

	// Benchmark linking
	for i := 0; i < b.N; i++ {
		linker.LinkTodos(parent.ID, children[i].ID, "parent-child")
	}
}