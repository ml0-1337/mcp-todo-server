package handlers

import (
	"encoding/json"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	"strings"
	"testing"
	"time"
)

func TestFormatTodoCreateResponseWithHints(t *testing.T) {
	tests := []struct {
		name           string
		todo           *core.Todo
		filePath       string
		wantHint       bool
		wantPattern    string
		wantSuggestion string
	}{
		{
			name: "Phase pattern generates hint",
			todo: &core.Todo{
				ID:       "test-phase-2",
				Task:     "Phase 2: Implementation",
				Started:  time.Now(),
				Status:   "in_progress",
				Priority: "high",
				Type:     "feature",
			},
			filePath:       "/path/to/test-phase-2.md",
			wantHint:       true,
			wantPattern:    "phase",
			wantSuggestion: "type 'phase' with a parent_id",
		},
		{
			name: "Regular todo no hint",
			todo: &core.Todo{
				ID:       "regular-todo",
				Task:     "Implement user authentication",
				Started:  time.Now(),
				Status:   "in_progress",
				Priority: "high",
				Type:     "feature",
			},
			filePath: "/path/to/regular-todo.md",
			wantHint: false,
		},
		{
			name: "Step pattern generates hint",
			todo: &core.Todo{
				ID:       "step-1-setup",
				Task:     "Step 1: Setup database",
				Started:  time.Now(),
				Status:   "in_progress",
				Priority: "medium",
				Type:     "feature",
			},
			filePath:       "/path/to/step-1-setup.md",
			wantHint:       true,
			wantPattern:    "step",
			wantSuggestion: "type 'subtask' with a parent_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTodoCreateResponseWithHints(tt.todo, tt.filePath, nil)

			// Parse the JSON response
			// The result is a CallToolResult with Content field
			var response map[string]interface{}
			if len(result.Content) == 0 {
				t.Fatal("No content in response")
			}
			textContent := result.Content[0].(mcp.TextContent)
			err := json.Unmarshal([]byte(textContent.Text), &response)
			if err != nil {
				t.Fatalf("Failed to parse response JSON: %v", err)
			}

			// Check basic fields
			if response["id"] != tt.todo.ID {
				t.Errorf("Expected id %q, got %v", tt.todo.ID, response["id"])
			}

			if response["path"] != tt.filePath {
				t.Errorf("Expected path %q, got %v", tt.filePath, response["path"])
			}

			// Check hint presence
			hint, hasHint := response["hint"]
			if tt.wantHint && !hasHint {
				t.Errorf("Expected hint in response, but none found")
			}

			if !tt.wantHint && hasHint {
				t.Errorf("Expected no hint, but found: %v", hint)
			}

			// Check hint content if expected
			if tt.wantHint && hasHint {
				hintMap, ok := hint.(map[string]interface{})
				if !ok {
					t.Fatalf("Hint is not a map: %T", hint)
				}

				if pattern, ok := hintMap["pattern"].(string); ok {
					if pattern != tt.wantPattern {
						t.Errorf("Expected pattern %q, got %q", tt.wantPattern, pattern)
					}
				}

				if message, ok := hintMap["message"].(string); ok {
					if !strings.Contains(message, tt.wantSuggestion) {
						t.Errorf("Expected message to contain %q, got %q", tt.wantSuggestion, message)
					}
				}
			}
		})
	}
}

func TestFormatTodoCreateResponseWithSimilarTodos(t *testing.T) {
	todo := &core.Todo{
		ID:       "phase-3-deployment",
		Task:     "Phase 3: Deployment",
		Started:  time.Now(),
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
	}

	existingTodos := []*core.Todo{
		{ID: "phase-1-planning", Task: "Phase 1: Planning"},
		{ID: "phase-2-implementation", Task: "Phase 2: Implementation"},
		{ID: "unrelated-todo", Task: "Fix bug in login"},
	}

	result := FormatTodoCreateResponseWithHints(todo, "/path/to/todo.md", existingTodos)

	// Parse response
	var response map[string]interface{}
	if len(result.Content) == 0 {
		t.Fatal("No content in response")
	}
	textContent := result.Content[0].(mcp.TextContent)
	err := json.Unmarshal([]byte(textContent.Text), &response)
	if err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Check for similar todos
	similar, hasSimilar := response["similar_todos"]
	if !hasSimilar {
		t.Error("Expected similar_todos in response")
	}

	if similarList, ok := similar.([]interface{}); ok {
		if len(similarList) != 2 {
			t.Errorf("Expected 2 similar todos, got %d", len(similarList))
		}
	}
}
