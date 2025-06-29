package handlers

import (
	"strings"
	"testing"
	"time"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

func TestFormatTodosSummaryWithHierarchy(t *testing.T) {
	// Test with hierarchical todos
	todos := []*core.Todo{
		{
			ID:       "project-main",
			Task:     "Main Project",
			Status:   "in_progress",
			Priority: "high",
			Type:     "multi-phase",
			Started:  time.Now(),
		},
		{
			ID:       "phase-1",
			Task:     "Phase 1: Design",
			Status:   "completed",
			Priority: "high",
			Type:     "phase",
			ParentID: "project-main",
			Started:  time.Now(),
		},
		{
			ID:       "phase-2",
			Task:     "Phase 2: Implementation",
			Status:   "in_progress",
			Priority: "high",
			Type:     "phase",
			ParentID: "project-main",
			Started:  time.Now(),
		},
		{
			ID:       "subtask-2-1",
			Task:     "Build core features",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "subtask",
			ParentID: "phase-2",
			Started:  time.Now(),
		},
		{
			ID:       "subtask-2-2",
			Task:     "Write tests",
			Status:   "blocked",
			Priority: "high",
			Type:     "subtask",
			ParentID: "phase-2",
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
			Task:     "Orphaned Phase",
			Status:   "blocked",
			Priority: "high",
			Type:     "phase",
			ParentID: "non-existent",
			Started:  time.Now(),
		},
	}
	
	result := formatTodosSummary(todos)
	output := result.Content[0].(mcp.TextContent).Text
	
	// Debug output
	t.Logf("Full output:\n%s", output)
	
	// Check for hierarchical view section
	if !strings.Contains(output, "HIERARCHICAL VIEW:") {
		t.Error("Output should contain HIERARCHICAL VIEW section")
	}
	
	// Check for tree structure
	if !strings.Contains(output, "[→] project-main: Main Project [HIGH] [multi-phase]") {
		t.Error("Should show project-main as root")
	}
	
	if !strings.Contains(output, "├── [→] phase-2: Phase 2: Implementation [HIGH] [phase]") {
		t.Error("Should show phase-2 with tree branch")
	}
	
	if !strings.Contains(output, "│   ├── [→] subtask-2-1: Build core features [subtask]") {
		t.Error("Should show subtask-2-1 nested under phase-2")
	}
	
	// Check for orphans
	if !strings.Contains(output, "ORPHANED PHASES/SUBTASKS") {
		t.Error("Should show orphaned phases section")
	}
	
	if !strings.Contains(output, "[✗] orphan-phase: Orphaned Phase [HIGH] [phase]") {
		t.Error("Should show orphan-phase in orphans section")
	}
	
	// Check for grouped view section
	if !strings.Contains(output, "GROUPED BY STATUS:") {
		t.Error("Should also show grouped view")
	}
}

func TestFormatTodosSummaryWithoutHierarchy(t *testing.T) {
	// Test with non-hierarchical todos
	todos := []*core.Todo{
		{
			ID:       "task-1",
			Task:     "Task 1",
			Status:   "in_progress",
			Priority: "high",
			Type:     "feature",
			Started:  time.Now(),
		},
		{
			ID:       "task-2",
			Task:     "Task 2",
			Status:   "completed",
			Priority: "medium",
			Type:     "bug",
			Started:  time.Now(),
		},
		{
			ID:       "task-3",
			Task:     "Task 3",
			Status:   "blocked",
			Priority: "low",
			Type:     "feature",
			Started:  time.Now(),
		},
	}
	
	result := formatTodosSummary(todos)
	output := result.Content[0].(mcp.TextContent).Text
	
	// Should not show hierarchical view
	if strings.Contains(output, "HIERARCHICAL VIEW:") {
		t.Error("Should not show hierarchical view for flat todos")
	}
	
	// Should show grouped view
	if !strings.Contains(output, "GROUPED BY STATUS:") {
		t.Error("Should show grouped view")
	}
	
	// Check status groups
	if !strings.Contains(output, "IN_PROGRESS (1):") {
		t.Error("Should show in_progress group")
	}
	
	if !strings.Contains(output, "[→] task-1: Task 1 [HIGH]") {
		t.Error("Should show task-1 in correct format")
	}
}

func TestFormatTodosSummaryEmpty(t *testing.T) {
	// Test with empty todo list
	todos := []*core.Todo{}
	
	result := formatTodosSummary(todos)
	output := result.Content[0].(mcp.TextContent).Text
	
	if output != "No todos found" {
		t.Errorf("Expected 'No todos found', got: %s", output)
	}
}

func TestFormatTodosSummaryMixedHierarchy(t *testing.T) {
	// Test with some hierarchical and some flat todos
	todos := []*core.Todo{
		{
			ID:       "parent",
			Task:     "Parent Task",
			Status:   "in_progress",
			Priority: "high",
			Type:     "multi-phase",
			Started:  time.Now(),
		},
		{
			ID:       "child",
			Task:     "Child Task",
			Status:   "in_progress",
			Priority: "medium",
			Type:     "phase",
			ParentID: "parent",
			Started:  time.Now(),
		},
		{
			ID:       "standalone-1",
			Task:     "Standalone 1",
			Status:   "completed",
			Priority: "low",
			Type:     "feature",
			Started:  time.Now(),
		},
		{
			ID:       "standalone-2",
			Task:     "Standalone 2",
			Status:   "blocked",
			Priority: "high",
			Type:     "bug",
			Started:  time.Now(),
		},
	}
	
	result := formatTodosSummary(todos)
	output := result.Content[0].(mcp.TextContent).Text
	
	// Should show hierarchical view because we have parent-child relationships
	if !strings.Contains(output, "HIERARCHICAL VIEW:") {
		t.Error("Should show hierarchical view when any relationships exist")
	}
	
	// Should show both hierarchical and standalone todos
	if !strings.Contains(output, "[→] parent: Parent Task [HIGH] [multi-phase]") {
		t.Error("Should show parent in hierarchy")
	}
	
	if !strings.Contains(output, "[✓] standalone-1: Standalone 1 [LOW]") {
		t.Error("Should show standalone todos as roots")
	}
}