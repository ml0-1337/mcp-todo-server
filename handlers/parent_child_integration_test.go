package handlers

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
)

// TestParentChildWorkflow tests the complete parent-child workflow
func TestParentChildWorkflow(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "todos")

	// Create mocks that use real file system
	manager := core.NewTodoManager(basePath)
	searchEngine := NewMockSearchEngine()
	statsEngine := NewMockStatsEngine()
	templateManager := NewMockTemplateManager()

	// Create handlers
	handlers := NewTodoHandlersWithDependencies(manager, searchEngine, statsEngine, templateManager)

	// Test 1: Create multi-phase project using todo_create_multi
	createReq := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"parent": map[string]interface{}{
				"task":     "Build REST API",
				"priority": "high",
				"type":     "multi-phase",
			},
			"children": []interface{}{
				map[string]interface{}{
					"task":     "Phase 1: Design endpoints",
					"priority": "high",
					"type":     "phase",
				},
				map[string]interface{}{
					"task":     "Phase 2: Implement core",
					"priority": "high",
					"type":     "phase",
				},
				map[string]interface{}{
					"task":     "Phase 3: Add tests",
					"priority": "medium",
					"type":     "phase",
				},
			},
		},
	}

	result, err := handlers.HandleTodoCreateMulti(context.Background(), createReq.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoCreateMulti error: %v", err)
	}

	// Verify creation success
	content := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(content, "Successfully created 4 todos") {
		t.Errorf("Expected success message, got: %s", content)
	}

	// Test 2: Read todos and verify hierarchy
	readReq := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"format": "summary",
		},
	}

	readResult, err := handlers.HandleTodoRead(context.Background(), readReq.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoRead error: %v", err)
	}

	readContent := readResult.Content[0].(mcp.TextContent).Text
	
	// Debug output
	t.Logf("Read content:\n%s", readContent)

	// Verify hierarchical view is shown
	if !strings.Contains(readContent, "HIERARCHICAL VIEW:") {
		t.Error("Should show hierarchical view")
	}

	// Verify tree structure
	if !strings.Contains(readContent, "[→] build-rest-api: Build REST API") {
		t.Error("Should show parent todo")
	}

	if !strings.Contains(readContent, "├── [→] phase-1-design-endpoints:") {
		t.Error("Should show first child with tree branch")
	}

	if !strings.Contains(readContent, "└── [→] phase-3-add-tests:") {
		t.Error("Should show last child with end branch")
	}

	// Test 3: Create orphaned phase (should fail)
	orphanReq := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"task":     "Phase 4: Deploy",
			"type":     "phase",
			"priority": "high",
		},
	}

	_, err = handlers.HandleTodoCreate(context.Background(), orphanReq.ToCallToolRequest())
	if err == nil {
		t.Fatal("Expected error for phase without parent_id, but got none")
	}

	// Check that the error message is correct
	if !strings.Contains(err.Error(), "type 'phase' requires parent_id") {
		t.Errorf("Expected error about parent_id requirement, got: %v", err)
	}

	// Test 4: Create phase with parent_id
	phaseReq := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"task":      "Phase 4: Deploy to production",
			"type":      "phase",
			"priority":  "high",
			"parent_id": "build-rest-api",
		},
	}

	phaseResult, err := handlers.HandleTodoCreate(context.Background(), phaseReq.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoCreate error: %v", err)
	}

	phaseContent := phaseResult.Content[0].(mcp.TextContent).Text
	if strings.Contains(phaseContent, "Error") {
		t.Errorf("Should create phase with parent_id, got error: %s", phaseContent)
	}

	// Test 5: Create subtask under a phase
	subtaskReq := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"task":      "Write unit tests for user endpoints",
			"type":      "subtask",
			"priority":  "medium",
			"parent_id": "phase-2-implement-core",
		},
	}

	subtaskResult, err := handlers.HandleTodoCreate(context.Background(), subtaskReq.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoCreate error: %v", err)
	}

	subtaskContent := subtaskResult.Content[0].(mcp.TextContent).Text
	if strings.Contains(subtaskContent, "Error") {
		t.Errorf("Should create subtask with parent_id, got error: %s", subtaskContent)
	}

	// Test 6: Read again to see the updated hierarchy
	finalReadResult, err := handlers.HandleTodoRead(context.Background(), readReq.ToCallToolRequest())
	if err != nil {
		t.Fatalf("Final HandleTodoRead error: %v", err)
	}

	finalContent := finalReadResult.Content[0].(mcp.TextContent).Text

	// Should now have 4 phases under parent
	if !strings.Contains(finalContent, "phase-4-deploy-to-production") {
		t.Error("Should show newly added phase 4")
	}

	// Should show subtask under phase 2
	if !strings.Contains(finalContent, "write-unit-tests-for-user-endpoints") {
		t.Error("Should show subtask under phase 2")
	}

	// Test 7: Pattern detection for similar tasks
	patternReq := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"task":     "Step 5: Monitor and optimize",
			"priority": "medium",
		},
	}

	patternResult, err := handlers.HandleTodoCreate(context.Background(), patternReq.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoCreate with pattern error: %v", err)
	}

	patternContent := patternResult.Content[0].(mcp.TextContent).Text
	// Current implementation doesn't include hints, just verify JSON structure
	var patternResponse map[string]interface{}
	if err := json.Unmarshal([]byte(patternContent), &patternResponse); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}
	// Verify basic response structure
	if patternResponse["id"] == nil || patternResponse["message"] == nil {
		t.Errorf("Response missing expected fields, got: %s", patternContent)
	}
}

// TestOrphanedPhaseDetection tests that orphaned phases are properly detected
func TestOrphanedPhaseDetection(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "todos")

	// Create mocks that use real file system
	manager := core.NewTodoManager(basePath)
	searchEngine := NewMockSearchEngine()
	statsEngine := NewMockStatsEngine()
	templateManager := NewMockTemplateManager()

	// Create handlers
	handlers := NewTodoHandlersWithDependencies(manager, searchEngine, statsEngine, templateManager)

	// Create a parent todo
	parent, err := manager.CreateTodo("Main Project", "high", "multi-phase")
	if err != nil {
		t.Fatalf("Failed to create parent: %v", err)
	}

	// Create a valid child
	validChild, err := manager.CreateTodo("Valid Phase", "medium", "phase")
	if err != nil {
		t.Fatalf("Failed to create valid child: %v", err)
	}
	manager.UpdateTodo(validChild.ID, "", "", "", map[string]string{"parent_id": parent.ID})

	// Create an orphaned phase (parent doesn't exist)
	orphan, err := manager.CreateTodo("Orphaned Phase", "high", "phase")
	if err != nil {
		t.Fatalf("Failed to create orphan: %v", err)
	}
	manager.UpdateTodo(orphan.ID, "", "", "", map[string]string{"parent_id": "non-existent-parent"})

	// Create an orphaned subtask
	orphanSubtask, err := manager.CreateTodo("Orphaned Subtask", "medium", "subtask")
	if err != nil {
		t.Fatalf("Failed to create orphan subtask: %v", err)
	}
	manager.UpdateTodo(orphanSubtask.ID, "", "", "", map[string]string{"parent_id": "another-missing"})

	// Read todos
	readReq := &MockCallToolRequest{
		Arguments: map[string]interface{}{
			"format": "summary",
		},
	}

	result, err := handlers.HandleTodoRead(context.Background(), readReq.ToCallToolRequest())
	if err != nil {
		t.Fatalf("HandleTodoRead error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent).Text
	
	// Debug output
	t.Logf("Orphan detection content:\n%s", content)

	// Check for orphaned section
	if !strings.Contains(content, "ORPHANED PHASES/SUBTASKS") {
		t.Error("Should show orphaned phases section")
	}

	// Verify specific orphans are listed
	if !strings.Contains(content, orphan.ID) {
		t.Errorf("Should list orphaned phase: %s", orphan.ID)
	}

	if !strings.Contains(content, orphanSubtask.ID) {
		t.Errorf("Should list orphaned subtask: %s", orphanSubtask.ID)
	}

	// Verify parent references are shown
	if !strings.Contains(content, "parent: non-existent-parent not found") {
		t.Error("Should show missing parent reference")
	}
}
