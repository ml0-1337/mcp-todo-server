package core

import (
	"strings"
	"testing"
	"time"
)

func TestTreeFormatterBasic(t *testing.T) {
	// Create test hierarchy
	todos := []*Todo{
		{
			ID:       "project",
			Task:     "Main Project",
			Status:   "in_progress",
			Priority: "high",
			Type:     "multi-phase",
			Started:  time.Now(),
		},
		{
			ID:       "phase1",
			Task:     "Phase 1",
			Status:   "completed",
			Priority: "medium",
			Type:     "phase",
			ParentID: "project",
			Started:  time.Now(),
		},
		{
			ID:       "phase2",
			Task:     "Phase 2",
			Status:   "in_progress",
			Priority: "high",
			Type:     "phase",
			ParentID: "project",
			Started:  time.Now(),
		},
		{
			ID:       "task2-1",
			Task:     "Task 2.1",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "subtask",
			ParentID: "phase2",
			Started:  time.Now(),
		},
	}

	roots, orphans := BuildTodoHierarchy(todos)
	formatter := NewTreeFormatter()

	output := formatter.FormatHierarchy(roots, orphans)

	// Debug: print the actual output
	t.Logf("Actual output:\n%s", output)

	// Verify the output contains expected elements
	if !strings.Contains(output, "[→] project: Main Project [HIGH] [multi-phase]") {
		t.Error("Output should contain project with correct formatting")
	}

	// phase2 comes first due to status/priority sorting
	if !strings.Contains(output, "├── [→] phase2: Phase 2 [HIGH] [phase]") {
		t.Error("Output should contain phase2 with tree branch")
	}

	if !strings.Contains(output, "└── [✓] phase1: Phase 1 [phase]") {
		t.Error("Output should contain phase1 as last child")
	}

	if !strings.Contains(output, "│   └── [→] task2-1: Task 2.1 [subtask]") {
		t.Error("Output should contain nested task2-1")
	}
}

func TestTreeFormatterWithOrphans(t *testing.T) {
	todos := []*Todo{
		{
			ID:       "root",
			Task:     "Root Task",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "feature",
			Started:  time.Now(),
		},
		{
			ID:       "orphan1",
			Task:     "Orphan Phase 1",
			Status:   "blocked",
			Priority: "high",
			Type:     "phase",
			ParentID: "missing",
			Started:  time.Now(),
		},
		{
			ID:       "orphan2",
			Task:     "Orphan Subtask",
			Status:   "in_progress",
			Priority: "low",
			Type:     "subtask",
			ParentID: "also-missing",
			Started:  time.Now(),
		},
	}

	roots, orphans := BuildTodoHierarchy(todos)
	formatter := NewTreeFormatter()

	output := formatter.FormatHierarchy(roots, orphans)

	// Check for orphans section
	if !strings.Contains(output, "ORPHANED PHASES/SUBTASKS (need parent assignment):") {
		t.Error("Output should contain orphans section header")
	}

	if !strings.Contains(output, "[✗] orphan1: Orphan Phase 1 [HIGH] [phase]") {
		t.Error("Output should contain orphan1")
	}

	if !strings.Contains(output, "[parent: missing not found]") {
		t.Error("Output should indicate missing parent")
	}
}

func TestTreeFormatterCustomSettings(t *testing.T) {
	todos := []*Todo{
		{
			ID:       "task1",
			Task:     "Task 1",
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
			Started:  time.Now(),
		},
	}

	roots, _ := BuildTodoHierarchy(todos)

	// Test with custom settings
	formatter := &TreeFormatter{
		ShowStatus:   false,
		ShowPriority: false,
		ShowType:     false,
		IndentSize:   2,
		Branch:       "+ ",
		LastBranch:   "` ",
		Vertical:     "| ",
		Space:        "  ",
	}

	output := formatter.FormatHierarchy(roots, nil)

	// Should not contain status, priority, or type indicators
	if strings.Contains(output, "[→]") {
		t.Error("Output should not contain status indicator when ShowStatus=false")
	}

	if strings.Contains(output, "[HIGH]") {
		t.Error("Output should not contain priority indicator when ShowPriority=false")
	}

	if strings.Contains(output, "[feature]") {
		t.Error("Output should not contain type indicator when ShowType=false")
	}
}

func TestFormatSimpleTree(t *testing.T) {
	todos := []*Todo{
		{
			ID:       "parent",
			Task:     "Parent",
			Status:   "in_progress",
			Priority: "high",
			Type:     "multi-phase",
			Started:  time.Now(),
		},
		{
			ID:       "child1",
			Task:     "Child 1",
			Status:   "completed",
			Priority: "medium",
			Type:     "phase",
			ParentID: "parent",
			Started:  time.Now(),
		},
		{
			ID:       "child2",
			Task:     "Child 2",
			Status:   "in_progress",
			Priority: "low",
			Type:     "phase",
			ParentID: "parent",
			Started:  time.Now(),
		},
	}

	roots, _ := BuildTodoHierarchy(todos)
	formatter := NewTreeFormatter()

	output := formatter.FormatSimpleTree(roots)

	// Check indentation
	lines := strings.Split(output, "\n")
	if len(lines) < 3 {
		t.Error("Should have at least 3 lines")
	}

	// Parent should not be indented
	if strings.HasPrefix(lines[0], " ") {
		t.Error("Root should not be indented")
	}

	// Children should be indented
	if !strings.HasPrefix(lines[1], "    ") {
		t.Error("Children should be indented")
	}
}

func TestFormatCompactTree(t *testing.T) {
	todos := []*Todo{
		{
			ID:       "root1",
			Task:     "Root 1",
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
			Started:  time.Now(),
		},
		{
			ID:       "root2",
			Task:     "Root 2",
			Status:   "completed",
			Priority: "medium",
			Type:     "feature",
			Started:  time.Now(),
		},
		{
			ID:       "child",
			Task:     "Child of Root 2",
			Status:   "blocked",
			Priority: "low",
			Type:     "subtask",
			ParentID: "root2",
			Started:  time.Now(),
		},
	}

	roots, _ := BuildTodoHierarchy(todos)
	formatter := NewTreeFormatter()

	output := formatter.FormatCompactTree(roots)

	// Check compact format
	if !strings.Contains(output, "→ root1: Root 1") {
		t.Error("Should use compact status indicators")
	}

	if !strings.Contains(output, "✓ root2: Root 2") {
		t.Error("Should show completed status")
	}

	if !strings.Contains(output, "└─ ✗ child: Child of Root 2") {
		t.Logf("Compact output:\n%s", output)
		t.Error("Should show child with compact branch")
	}
}

func TestFormatHierarchyWithStats(t *testing.T) {
	todos := []*Todo{
		{
			ID:       "parent",
			Task:     "Parent",
			Status:   "in_progress",
			Priority: "high",
			Type:     "multi-phase",
			Started:  time.Now(),
		},
		{
			ID:       "child1",
			Task:     "Child 1",
			Status:   "completed",
			Priority: "medium",
			Type:     "phase",
			ParentID: "parent",
			Started:  time.Now(),
		},
		{
			ID:       "child2",
			Task:     "Child 2",
			Status:   "in_progress",
			Priority: "low",
			Type:     "phase",
			ParentID: "parent",
			Started:  time.Now(),
		},
		{
			ID:       "orphan",
			Task:     "Orphan",
			Status:   "blocked",
			Priority: "high",
			Type:     "phase",
			ParentID: "missing",
			Started:  time.Now(),
		},
	}

	roots, orphans := BuildTodoHierarchy(todos)
	formatter := NewTreeFormatter()

	output := formatter.FormatHierarchyWithStats(roots, orphans, todos)

	// Check for stats header
	if !strings.Contains(output, "Todo Hierarchy (4 total, 1 roots, depth: 2)") {
		t.Error("Should contain hierarchy statistics")
	}

	// Check for summary footer - parent count should be 3 (child1, child2, and orphan all have parent_id)
	if !strings.Contains(output, "Summary: 3 with parents, 1 orphaned") {
		t.Logf("Stats output:\n%s", output)
		t.Error("Should contain summary statistics")
	}
}

func TestStatusIndicators(t *testing.T) {
	formatter := NewTreeFormatter()

	tests := []struct {
		status   string
		expected string
	}{
		{"completed", "[✓]"},
		{"in_progress", "[→]"},
		{"blocked", "[✗]"},
		{"unknown", "[ ]"},
	}

	for _, test := range tests {
		result := formatter.getStatusIndicator(test.status)
		if result != test.expected {
			t.Errorf("Status %s: expected %s, got %s", test.status, test.expected, result)
		}
	}
}

func TestPriorityIndicators(t *testing.T) {
	formatter := NewTreeFormatter()

	tests := []struct {
		priority string
		expected string
	}{
		{"high", "[HIGH]"},
		{"low", "[LOW]"},
		{"medium", ""},
	}

	for _, test := range tests {
		result := formatter.getPriorityIndicator(test.priority)
		if result != test.expected {
			t.Errorf("Priority %s: expected %s, got %s", test.priority, test.expected, result)
		}
	}
}

func TestFormatFlatWithIndication(t *testing.T) {
	todos := []*Todo{
		{
			ID:       "parent1",
			Task:     "Parent Task 1",
			Status:   "in_progress",
			Priority: "high",
			Type:     "multi-phase",
			Started:  time.Now(),
		},
		{
			ID:       "parent2",
			Task:     "Parent Task 2",
			Status:   "completed",
			Priority: "medium",
			Type:     "feature",
			Started:  time.Now(),
		},
		{
			ID:       "child1",
			Task:     "Child of Parent 1",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "phase",
			ParentID: "parent1",
			Started:  time.Now(),
		},
		{
			ID:       "child2",
			Task:     "Another Child of Parent 1",
			Status:   "completed",
			Priority: "low",
			Type:     "phase",
			ParentID: "parent1",
			Started:  time.Now(),
		},
		{
			ID:       "standalone",
			Task:     "Standalone Task",
			Status:   "blocked",
			Priority: "high",
			Type:     "bug",
			Started:  time.Now(),
		},
		{
			ID:       "grandchild",
			Task:     "Grandchild",
			Status:   "pending",
			Priority: "low",
			Type:     "subtask",
			ParentID: "child1",
			Started:  time.Now(),
		},
	}

	formatter := NewTreeFormatter()
	output := formatter.FormatFlatWithIndication(todos)

	// Test that output has status sections
	if !strings.Contains(output, "IN_PROGRESS") {
		t.Error("Output should contain IN_PROGRESS section")
	}
	if !strings.Contains(output, "BLOCKED") {
		t.Error("Output should contain BLOCKED section")
	}
	if !strings.Contains(output, "COMPLETED") {
		t.Error("Output should contain COMPLETED section")
	}

	// Test parent indications
	if !strings.Contains(output, "[parent: parent1]") {
		t.Error("child1 should show parent indication")
	}

	// Test children counts - note: countChildren only counts within same status group
	lines := strings.Split(output, "\n")
	parent1Found := false
	child1Found := false
	
	for _, line := range lines {
		if strings.Contains(line, "parent1: Parent Task 1") {
			parent1Found = true
			// parent1 has 1 child in in_progress status (child1)
			if !strings.Contains(line, "[1 children]") {
				t.Errorf("parent1 line should show [1 children] (within in_progress), but got: %s", line)
			}
		}
		if strings.Contains(line, "child1: Child of Parent 1") {
			child1Found = true
			// child1 has no children in in_progress status
			if strings.Contains(line, "children]") {
				t.Errorf("child1 line should not show children count, but got: %s", line)
			}
		}
	}
	
	if !parent1Found {
		t.Error("Could not find parent1 in output")
	}
	if !child1Found {
		t.Error("Could not find child1 in output")
	}

	// Verify parent2 has no children
	for _, line := range lines {
		if strings.Contains(line, "parent2: Parent Task 2") && strings.Contains(line, "children]") {
			t.Error("parent2 should not show children count")
		}
	}

	// Verify standalone has no parent indication
	for _, line := range lines {
		if strings.Contains(line, "standalone: Standalone Task") && strings.Contains(line, "[parent:") {
			t.Error("standalone should not have parent indication")
		}
	}
}

func TestCountChildrenIndirect(t *testing.T) {
	// Test countChildren indirectly through FormatFlatWithIndication
	// Note: countChildren only counts within the same status group
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
			ID:       "child1",
			Task:     "Child 1",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "phase",
			ParentID: "root",
			Started:  time.Now(),
		},
		{
			ID:       "child2",
			Task:     "Child 2",
			Status:   "in_progress",
			Priority: "low",
			Type:     "phase",
			ParentID: "root",
			Started:  time.Now(),
		},
		{
			ID:       "child3",
			Task:     "Child 3",
			Status:   "completed",
			Priority: "medium",
			Type:     "phase",
			ParentID: "root",
			Started:  time.Now(),
		},
		{
			ID:       "grandchild1",
			Task:     "Grandchild 1",
			Status:   "in_progress",
			Priority: "low",
			Type:     "subtask",
			ParentID: "child1",
			Started:  time.Now(),
		},
		{
			ID:       "grandchild2",
			Task:     "Grandchild 2",
			Status:   "completed",
			Priority: "low",
			Type:     "subtask",
			ParentID: "child1",
			Started:  time.Now(),
		},
	}

	formatter := NewTreeFormatter()
	output := formatter.FormatFlatWithIndication(todos)

	// Check that root shows 2 children in in_progress (child1 and child2)
	lines := strings.Split(output, "\n")
	rootFound := false
	child1Found := false
	
	for _, line := range lines {
		if strings.Contains(line, "root: Root") {
			rootFound = true
			// root has 2 children in in_progress status
			if !strings.Contains(line, "[2 children]") {
				t.Errorf("Root line should show [2 children] (within in_progress), but got: %s", line)
			}
		}
		if strings.Contains(line, "child1: Child 1") {
			child1Found = true
			// child1 has 1 child in in_progress status
			if !strings.Contains(line, "[1 children]") {
				t.Errorf("Child1 line should show [1 children] (within in_progress), but got: %s", line)
			}
		}
	}
	
	if !rootFound {
		t.Error("Could not find root in output")
	}
	if !child1Found {
		t.Error("Could not find child1 in output")
	}

	// Check that child2 doesn't show children count
	for _, line := range lines {
		if strings.Contains(line, "child2: Child 2") && strings.Contains(line, "children]") {
			t.Error("Child2 should not show children count")
		}
	}

	// Check completed section
	completedSectionFound := false
	for _, line := range lines {
		if strings.Contains(line, "child3: Child 3") {
			completedSectionFound = true
			// child3 should not have children in completed status
			if strings.Contains(line, "children]") {
				t.Error("Child3 should not show children count")
			}
		}
	}
	if !completedSectionFound {
		t.Error("Could not find child3 in completed section")
	}

	// Test with single child - all same status
	singleChildTodos := []*Todo{
		{
			ID:       "parent",
			Task:     "Parent",
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
			Started:  time.Now(),
		},
		{
			ID:       "onlychild",
			Task:     "Only Child",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "subtask",
			ParentID: "parent",
			Started:  time.Now(),
		},
	}

	output2 := formatter.FormatFlatWithIndication(singleChildTodos)
	if !strings.Contains(output2, "[1 children]") {
		t.Error("Parent with single child should show [1 children]")
	}
}

