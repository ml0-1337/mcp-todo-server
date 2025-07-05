package core

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// Test 23: Linked todos should maintain referential integrity
func TestLinkedTodosReferentialIntegrity(t *testing.T) {
	// Create temp directory for test
	tempDir, err := ioutil.TempDir("", "linked-todos-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create todo manager
	manager := NewTodoManager(tempDir)

	t.Run("Create todo with valid parent_id", func(t *testing.T) {
		// Create parent todo first
		parent, err := manager.CreateTodo("Multi-phase project", "high", "multi-phase")
		if err != nil {
			t.Fatalf("Failed to create parent todo: %v", err)
		}

		// Create child todo with parent_id
		child, err := manager.CreateTodoWithParent("Phase 1: Planning", "high", "feature", parent.ID)
		if err != nil {
			t.Fatalf("Failed to create child todo: %v", err)
		}

		// Verify child has parent_id set
		if child.ParentID != parent.ID {
			t.Errorf("Expected child parent_id to be %s, got %s", parent.ID, child.ParentID)
		}

		// Read child to verify parent_id persisted
		readChild, err := manager.ReadTodo(child.ID)
		if err != nil {
			t.Fatalf("Failed to read child todo: %v", err)
		}
		if readChild.ParentID != parent.ID {
			t.Errorf("Parent ID not persisted correctly")
		}
	})

	t.Run("Error when creating todo with non-existent parent_id", func(t *testing.T) {
		// Try to create todo with invalid parent_id
		_, err := manager.CreateTodoWithParent("Orphan task", "medium", "feature", "non-existent-parent")
		if err == nil {
			t.Error("Should return error when parent_id doesn't exist")
		}
		if err != nil && err.Error() != "parent todo 'non-existent-parent' not found" {
			t.Errorf("Expected 'parent todo not found' error, got: %v", err)
		}
	})

	t.Run("List children of parent todo", func(t *testing.T) {
		// Create parent and multiple children
		parent, _ := manager.CreateTodo("Epic: New feature", "high", "multi-phase")
		child1, _ := manager.CreateTodoWithParent("Task 1", "high", "feature", parent.ID)
		child2, _ := manager.CreateTodoWithParent("Task 2", "medium", "feature", parent.ID)
		child3, _ := manager.CreateTodoWithParent("Task 3", "low", "bug", parent.ID)

		// Get children
		children, err := manager.GetChildren(parent.ID)
		if err != nil {
			t.Fatalf("Failed to get children: %v", err)
		}

		// Verify all children returned
		if len(children) != 3 {
			t.Errorf("Expected 3 children, got %d", len(children))
		}

		// Verify children IDs
		childIDs := make(map[string]bool)
		for _, child := range children {
			childIDs[child.ID] = true
		}

		if !childIDs[child1.ID] || !childIDs[child2.ID] || !childIDs[child3.ID] {
			t.Error("Not all children returned")
		}
	})

	t.Run("Prevent archiving parent with active children", func(t *testing.T) {
		// Create parent with active children
		parent, _ := manager.CreateTodo("Project with children", "high", "multi-phase")
		manager.CreateTodoWithParent("Active child 1", "high", "feature", parent.ID)
		manager.CreateTodoWithParent("Active child 2", "medium", "feature", parent.ID)

		// Try to archive parent
		err := manager.ArchiveTodo(parent.ID)
		if err == nil {
			t.Error("Should not allow archiving parent with active children")
		}
		if err != nil && !strings.Contains(err.Error(), "has active children") {
			t.Errorf("Error should mention active children, got: %v", err)
		}
	})

	t.Run("Allow archiving parent with all completed children", func(t *testing.T) {
		// Create parent with children
		parent, _ := manager.CreateTodo("Completed project", "high", "multi-phase")
		child1, _ := manager.CreateTodoWithParent("Completed task 1", "high", "feature", parent.ID)
		child2, _ := manager.CreateTodoWithParent("Completed task 2", "medium", "feature", parent.ID)

		// Mark children as completed
		manager.UpdateTodo(child1.ID, "", "", "", map[string]string{"status": "completed"})
		manager.UpdateTodo(child2.ID, "", "", "", map[string]string{"status": "completed"})

		// Archive children first
		err := manager.ArchiveTodo(child1.ID)
		if err != nil {
			t.Fatalf("Failed to archive child1: %v", err)
		}
		err = manager.ArchiveTodo(child2.ID)
		if err != nil {
			t.Fatalf("Failed to archive child2: %v", err)
		}

		// Now archive parent should succeed
		err = manager.ArchiveTodo(parent.ID)
		if err != nil {
			t.Errorf("Should allow archiving parent when all children are archived, got: %v", err)
		}
	})

	t.Run("Cascade archive option", func(t *testing.T) {
		// Create parent with completed children
		parent, _ := manager.CreateTodo("Project for cascade", "high", "multi-phase")
		child1, _ := manager.CreateTodoWithParent("Child for cascade 1", "high", "feature", parent.ID)
		child2, _ := manager.CreateTodoWithParent("Child for cascade 2", "medium", "feature", parent.ID)

		// Mark all as completed
		manager.UpdateTodo(parent.ID, "", "", "", map[string]string{"status": "completed"})
		manager.UpdateTodo(child1.ID, "", "", "", map[string]string{"status": "completed"})
		manager.UpdateTodo(child2.ID, "", "", "", map[string]string{"status": "completed"})

		// Archive with cascade option
		err := manager.ArchiveTodoWithCascade(parent.ID, true)
		if err != nil {
			t.Fatalf("Failed to cascade archive: %v", err)
		}

		// Verify parent and children are archived
		// Check parent archived
		parentPath := GetArchivePath(tempDir, parent, "")
		if _, err := os.Stat(parentPath); os.IsNotExist(err) {
			t.Error("Parent should be archived")
		}

		// Check children archived
		child1Path := GetArchivePath(tempDir, child1, "")
		if _, err := os.Stat(child1Path); os.IsNotExist(err) {
			t.Error("Child 1 should be archived with cascade")
		}

		child2Path := GetArchivePath(tempDir, child2, "")
		if _, err := os.Stat(child2Path); os.IsNotExist(err) {
			t.Error("Child 2 should be archived with cascade")
		}
	})
}
