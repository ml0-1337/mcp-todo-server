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