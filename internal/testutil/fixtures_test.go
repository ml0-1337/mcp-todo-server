package testutil_test

import (
	"fmt"
	"strings"
	"testing"
	"time"
	
	"github.com/user/mcp-todo-server/internal/testutil"
)

func TestTodoBuilder(t *testing.T) {
	t.Helper()
	
	// Test basic builder
	task := "Test task"
	todo := testutil.NewTodoBuilder(task).Build()
	
	if todo.Task != task {
		t.Errorf("Expected task %q, got %q", task, todo.Task)
	}
	
	// Test default values
	if todo.Priority != "high" {
		t.Errorf("Expected default priority 'high', got %q", todo.Priority)
	}
	
	if todo.Type != "feature" {
		t.Errorf("Expected default type 'feature', got %q", todo.Type)
	}
	
	if todo.Status != "in_progress" {
		t.Errorf("Expected default status 'in_progress', got %q", todo.Status)
	}
}

func TestTodoBuilderWithChaining(t *testing.T) {
	t.Helper()
	
	task := "Complex task"
	priority := "low"
	todoType := "refactor"
	status := "completed"
	parentID := "parent-123"
	started := time.Now().Add(-24 * time.Hour)
	completed := time.Now()
	
	todo := testutil.NewTodoBuilder(task).
		WithPriority(priority).
		WithType(todoType).
		WithStatus(status).
		WithParentID(parentID).
		WithStarted(started).
		WithCompleted(completed).
		Build()
	
	if todo.Task != task {
		t.Errorf("Expected task %q, got %q", task, todo.Task)
	}
	
	if todo.Priority != priority {
		t.Errorf("Expected priority %q, got %q", priority, todo.Priority)
	}
	
	if todo.Type != todoType {
		t.Errorf("Expected type %q, got %q", todoType, todo.Type)
	}
	
	if todo.Status != status {
		t.Errorf("Expected status %q, got %q", status, todo.Status)
	}
	
	if todo.ParentID != parentID {
		t.Errorf("Expected parent ID %q, got %q", parentID, todo.ParentID)
	}
}

func TestSampleTodos(t *testing.T) {
	t.Helper()
	
	// Test FeatureHigh
	featureTodo := testutil.SampleTodos.FeatureHigh()
	if featureTodo.Task != "Implement user authentication" {
		t.Errorf("Unexpected task for FeatureHigh: %s", featureTodo.Task)
	}
	if featureTodo.Priority != "high" {
		t.Errorf("Expected priority 'high', got %q", featureTodo.Priority)
	}
	if featureTodo.Type != "feature" {
		t.Errorf("Expected type 'feature', got %q", featureTodo.Type)
	}
	
	// Test BugMedium
	bugTodo := testutil.SampleTodos.BugMedium()
	if bugTodo.Task != "Fix login timeout issue" {
		t.Errorf("Unexpected task for BugMedium: %s", bugTodo.Task)
	}
	if bugTodo.Priority != "medium" {
		t.Errorf("Expected priority 'medium', got %q", bugTodo.Priority)
	}
	if bugTodo.Type != "bug" {
		t.Errorf("Expected type 'bug', got %q", bugTodo.Type)
	}
	
	// Test RefactorLow
	refactorTodo := testutil.SampleTodos.RefactorLow()
	if refactorTodo.Priority != "low" {
		t.Errorf("Expected priority 'low', got %q", refactorTodo.Priority)
	}
	if refactorTodo.Type != "refactor" {
		t.Errorf("Expected type 'refactor', got %q", refactorTodo.Type)
	}
	
	// Test MultiPhaseParent
	parentTodo := testutil.SampleTodos.MultiPhaseParent()
	if parentTodo.Type != "multi-phase" {
		t.Errorf("Expected type 'multi-phase', got %q", parentTodo.Type)
	}
	
	// Test PhaseChild
	childTodo := testutil.SampleTodos.PhaseChild(parentTodo.ID)
	if childTodo.ParentID != parentTodo.ID {
		t.Errorf("Expected parent ID %q, got %q", parentTodo.ID, childTodo.ParentID)
	}
	if childTodo.Type != "phase" {
		t.Errorf("Expected type 'phase', got %q", childTodo.Type)
	}
}

func TestGenerateTestContent(t *testing.T) {
	t.Helper()
	
	todo := testutil.NewTodoBuilder("Test todo").Build()
	sections := map[string]string{
		"findings_and_research": "Custom research content",
		"checklist": "- [ ] Item 1\n- [x] Item 2",
	}
	
	content := testutil.GenerateTestContent(todo, sections)
	
	// Check that content includes frontmatter
	if !strings.Contains(content, "---") {
		t.Error("Expected content to contain frontmatter delimiters")
	}
	
	// Check that content includes task
	if !strings.Contains(content, "# Task: Test todo") {
		t.Error("Expected content to contain task heading")
	}
	
	// Check that custom sections are included
	if !strings.Contains(content, "Custom research content") {
		t.Error("Expected custom research content to be included")
	}
	
	if !strings.Contains(content, "- [ ] Item 1") {
		t.Error("Expected checklist content to be included")
	}
	
	// Check that all default sections are present
	expectedSections := []string{
		"## Findings & Research",
		"## Web Searches",
		"## Test Strategy",
		"## Test List",
		"## Test Cases",
		"## Maintainability Analysis",
		"## Test Results Log",
		"## Checklist",
		"## Working Scratchpad",
	}
	
	for _, section := range expectedSections {
		if !strings.Contains(content, section) {
			t.Errorf("Expected section %q to be present", section)
		}
	}
}

func TestCreateChecklistItems(t *testing.T) {
	t.Helper()
	
	items := []string{"Item 1", "Item 2", "Item 3", "Item 4", "Item 5"}
	content := testutil.CreateChecklistItems(items...)
	
	lines := strings.Split(content, "\n")
	
	// Check pattern: completed, pending, in-progress, completed, pending
	expectedPatterns := []string{
		"- [x] Item 1",
		"- [ ] Item 2",
		"- [>] Item 3",
		"- [x] Item 4",
		"- [ ] Item 5",
	}
	
	for i, expected := range expectedPatterns {
		if !strings.Contains(lines[i], expected) {
			t.Errorf("Line %d: expected %q, got %q", i, expected, lines[i])
		}
	}
}

func TestCreateTestCases(t *testing.T) {
	t.Helper()
	
	cases := []string{"Basic functionality", "Edge case handling", "Error scenarios"}
	content := testutil.CreateTestCases(cases...)
	
	// Check that content includes test case blocks
	if !strings.Contains(content, "```go") {
		t.Error("Expected content to contain Go code blocks")
	}
	
	// Check that each test case is included
	for i, testCase := range cases {
		expectedComment := fmt.Sprintf("// Test %d: %s", i+1, testCase)
		if !strings.Contains(content, expectedComment) {
			t.Errorf("Expected content to contain %q", expectedComment)
		}
		
		expectedFunc := fmt.Sprintf("func Test%d(t *testing.T)", i+1)
		if !strings.Contains(content, expectedFunc) {
			t.Errorf("Expected content to contain %q", expectedFunc)
		}
	}
}

func TestCreateWebSearches(t *testing.T) {
	t.Helper()
	
	queries := []string{"golang best practices", "error handling patterns", "test driven development"}
	content := testutil.CreateWebSearches(queries...)
	
	// Check that each query is included
	for _, query := range queries {
		expectedQuery := fmt.Sprintf(`Query: "%s"`, query)
		if !strings.Contains(content, expectedQuery) {
			t.Errorf("Expected content to contain %q", expectedQuery)
		}
		
		expectedResult := fmt.Sprintf("Found relevant information about %s", query)
		if !strings.Contains(content, expectedResult) {
			t.Errorf("Expected content to contain result for %q", query)
		}
	}
	
	// Check that timestamps are included
	if !strings.Contains(content, "[202") { // Partial year match
		t.Error("Expected content to contain timestamps")
	}
}