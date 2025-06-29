package core

import (
	"testing"
	"time"
)

func TestBuildTodoHierarchy(t *testing.T) {
	// Create test todos
	todos := []*Todo{
		{
			ID:       "parent-1",
			Task:     "Parent Task 1",
			Status:   "in_progress",
			Priority: "high",
			Type:     "multi-phase",
			Started:  time.Now(),
		},
		{
			ID:       "child-1-1",
			Task:     "Child Task 1.1",
			Status:   "completed",
			Priority: "medium",
			Type:     "phase",
			ParentID: "parent-1",
			Started:  time.Now(),
		},
		{
			ID:       "child-1-2",
			Task:     "Child Task 1.2",
			Status:   "in_progress",
			Priority: "high",
			Type:     "phase",
			ParentID: "parent-1",
			Started:  time.Now(),
		},
		{
			ID:       "grandchild-1-2-1",
			Task:     "Grandchild Task 1.2.1",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "subtask",
			ParentID: "child-1-2",
			Started:  time.Now(),
		},
		{
			ID:       "standalone",
			Task:     "Standalone Task",
			Status:   "in_progress",
			Priority: "low",
			Type:     "feature",
			Started:  time.Now(),
		},
		{
			ID:       "orphan-phase",
			Task:     "Orphan Phase",
			Status:   "blocked",
			Priority: "high",
			Type:     "phase",
			ParentID: "non-existent",
			Started:  time.Now(),
		},
	}
	
	// Build hierarchy
	roots, orphans := BuildTodoHierarchy(todos)
	
	// Test roots
	if len(roots) != 2 {
		t.Errorf("Expected 2 root nodes, got %d", len(roots))
	}
	
	// Find parent-1 in roots
	var parent1Node *TodoNode
	for _, root := range roots {
		if root.Todo.ID == "parent-1" {
			parent1Node = root
			break
		}
	}
	
	if parent1Node == nil {
		t.Fatal("parent-1 should be a root node")
	}
	
	// Test children of parent-1
	if len(parent1Node.Children) != 2 {
		t.Errorf("parent-1 should have 2 children, got %d", len(parent1Node.Children))
	}
	
	// Find child-1-2 and check its children
	var child12Node *TodoNode
	for _, child := range parent1Node.Children {
		if child.Todo.ID == "child-1-2" {
			child12Node = child
			break
		}
	}
	
	if child12Node == nil {
		t.Fatal("child-1-2 should be a child of parent-1")
	}
	
	if len(child12Node.Children) != 1 {
		t.Errorf("child-1-2 should have 1 child, got %d", len(child12Node.Children))
	}
	
	if child12Node.Children[0].Todo.ID != "grandchild-1-2-1" {
		t.Error("grandchild-1-2-1 should be a child of child-1-2")
	}
	
	// Test orphans
	if len(orphans) != 1 {
		t.Errorf("Expected 1 orphan, got %d", len(orphans))
	}
	
	if orphans[0].ID != "orphan-phase" {
		t.Error("orphan-phase should be in orphans list")
	}
}

func TestCircularReferenceDetection(t *testing.T) {
	// Create todos with circular reference
	todos := []*Todo{
		{
			ID:       "todo-a",
			Task:     "Task A",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "feature",
			ParentID: "todo-c", // Creates cycle: A -> C -> B -> A
			Started:  time.Now(),
		},
		{
			ID:       "todo-b",
			Task:     "Task B",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "feature",
			ParentID: "todo-a",
			Started:  time.Now(),
		},
		{
			ID:       "todo-c",
			Task:     "Task C",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "feature",
			ParentID: "todo-b",
			Started:  time.Now(),
		},
	}
	
	// Build hierarchy
	roots, orphans := BuildTodoHierarchy(todos)
	
	// All todos should be either roots or orphans due to circular reference
	totalProcessed := len(roots) + len(orphans)
	if totalProcessed != len(todos) {
		t.Errorf("Not all todos were processed. Expected %d, got %d", len(todos), totalProcessed)
	}
	
	// Verify no infinite loops occurred (test would hang if there was an issue)
}

func TestSortNodes(t *testing.T) {
	// Create test todos with different statuses and priorities
	todos := []*Todo{
		{
			ID:       "completed-low",
			Task:     "Completed Low Priority",
			Status:   "completed",
			Priority: "low",
			Type:     "feature",
			Started:  time.Now(),
		},
		{
			ID:       "progress-high",
			Task:     "In Progress High Priority",
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
			Started:  time.Now(),
		},
		{
			ID:       "blocked-medium",
			Task:     "Blocked Medium Priority",
			Status:   "blocked",
			Priority: "medium",
			Type:     "feature",
			Started:  time.Now(),
		},
		{
			ID:       "progress-low",
			Task:     "In Progress Low Priority",
			Status:   "in_progress",
			Priority: "low",
			Type:     "feature",
			Started:  time.Now(),
		},
	}
	
	// Build hierarchy (all should be roots)
	roots, _ := BuildTodoHierarchy(todos)
	
	// Verify sort order
	expectedOrder := []string{"progress-high", "progress-low", "blocked-medium", "completed-low"}
	
	if len(roots) != len(expectedOrder) {
		t.Fatalf("Expected %d roots, got %d", len(expectedOrder), len(roots))
	}
	
	for i, expectedID := range expectedOrder {
		if roots[i].Todo.ID != expectedID {
			t.Errorf("Position %d: expected %s, got %s", i, expectedID, roots[i].Todo.ID)
		}
	}
}

func TestGetHierarchyDepth(t *testing.T) {
	// Create todos with known depth
	todos := []*Todo{
		{
			ID:       "root",
			Task:     "Root",
			Status:   "in_progress",
			Priority: "high",
			Type:     "multi-phase",
			Started:  time.Now(),
		},
		{
			ID:       "level1",
			Task:     "Level 1",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "phase",
			ParentID: "root",
			Started:  time.Now(),
		},
		{
			ID:       "level2",
			Task:     "Level 2",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "subtask",
			ParentID: "level1",
			Started:  time.Now(),
		},
		{
			ID:       "level3",
			Task:     "Level 3",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "subtask",
			ParentID: "level2",
			Started:  time.Now(),
		},
	}
	
	roots, _ := BuildTodoHierarchy(todos)
	depth := GetHierarchyDepth(roots)
	
	if depth != 4 {
		t.Errorf("Expected depth of 4, got %d", depth)
	}
}

func TestGetOrphanedPhases(t *testing.T) {
	todos := []*Todo{
		{
			ID:       "valid-parent",
			Task:     "Valid Parent",
			Status:   "in_progress",
			Priority: "high",
			Type:     "multi-phase",
			Started:  time.Now(),
		},
		{
			ID:       "valid-phase",
			Task:     "Valid Phase",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "phase",
			ParentID: "valid-parent",
			Started:  time.Now(),
		},
		{
			ID:       "orphan-phase-1",
			Task:     "Orphan Phase 1",
			Status:   "blocked",
			Priority: "high",
			Type:     "phase",
			ParentID: "missing-parent",
			Started:  time.Now(),
		},
		{
			ID:       "orphan-subtask",
			Task:     "Orphan Subtask",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "subtask",
			ParentID: "another-missing-parent",
			Started:  time.Now(),
		},
		{
			ID:       "regular-with-missing-parent",
			Task:     "Regular Todo with Missing Parent",
			Status:   "in_progress",
			Priority: "low",
			Type:     "feature",
			ParentID: "yet-another-missing",
			Started:  time.Now(),
		},
	}
	
	orphans := GetOrphanedPhases(todos)
	
	// Should only include phase/subtask types with missing parents
	if len(orphans) != 2 {
		t.Errorf("Expected 2 orphaned phases/subtasks, got %d", len(orphans))
	}
	
	// Verify the correct todos are identified as orphans
	orphanIDs := make(map[string]bool)
	for _, orphan := range orphans {
		orphanIDs[orphan.ID] = true
	}
	
	if !orphanIDs["orphan-phase-1"] {
		t.Error("orphan-phase-1 should be identified as orphan")
	}
	
	if !orphanIDs["orphan-subtask"] {
		t.Error("orphan-subtask should be identified as orphan")
	}
	
	if orphanIDs["regular-with-missing-parent"] {
		t.Error("regular todos should not be in orphaned phases list")
	}
}

func TestValidateHierarchy(t *testing.T) {
	todos := []*Todo{
		// Valid parent-child
		{
			ID:       "valid-parent",
			Task:     "Valid Parent",
			Status:   "in_progress",
			Priority: "high",
			Type:     "multi-phase",
			Started:  time.Now(),
		},
		{
			ID:       "valid-phase",
			Task:     "Valid Phase",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "phase",
			ParentID: "valid-parent",
			Started:  time.Now(),
		},
		// Phase without parent_id
		{
			ID:       "phase-without-parent",
			Task:     "Phase Without Parent",
			Status:   "blocked",
			Priority: "high",
			Type:     "phase",
			Started:  time.Now(),
		},
		// Orphaned phase
		{
			ID:       "orphan-phase",
			Task:     "Orphan Phase",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "phase",
			ParentID: "non-existent",
			Started:  time.Now(),
		},
		// Circular reference
		{
			ID:       "circular-a",
			Task:     "Circular A",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "feature",
			ParentID: "circular-b",
			Started:  time.Now(),
		},
		{
			ID:       "circular-b",
			Task:     "Circular B",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "feature",
			ParentID: "circular-a",
			Started:  time.Now(),
		},
	}
	
	issues := ValidateHierarchy(todos)
	
	// Should have at least 3 issues
	if len(issues) < 3 {
		t.Errorf("Expected at least 3 validation issues, got %d", len(issues))
		for _, issue := range issues {
			t.Logf("Issue: %s", issue)
		}
	}
	
	// Check for specific issues
	hasPhaseWithoutParent := false
	hasOrphanedPhase := false
	hasCircularRef := false
	
	for _, issue := range issues {
		if containsString(issue, "phase-without-parent") && containsString(issue, "should have a parent_id") {
			hasPhaseWithoutParent = true
		}
		if containsString(issue, "orphan-phase") && containsString(issue, "non-existent") {
			hasOrphanedPhase = true
		}
		if containsString(issue, "Circular reference") {
			hasCircularRef = true
		}
	}
	
	if !hasPhaseWithoutParent {
		t.Error("Should detect phase without parent_id")
	}
	if !hasOrphanedPhase {
		t.Error("Should detect orphaned phase")
	}
	if !hasCircularRef {
		t.Error("Should detect circular reference")
	}
}

// Helper function
func containsString(str, substr string) bool {
	return len(substr) > 0 && len(str) >= len(substr) && (str == substr || (len(str) > len(substr) && containsSubstring(str, substr)))
}

func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}