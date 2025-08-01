package handlers

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/user/mcp-todo-server/core"
)

func TestFormatTodosFull(t *testing.T) {
	tests := []struct {
		name        string
		todos       []*core.Todo
		wantContain []string
	}{
		{
			name: "format multiple todos",
			todos: []*core.Todo{
				{
					ID:       "todo1",
					Task:     "First task",
					Status:   "in_progress",
					Priority: "high",
					Type:     "feature",
					Started:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Tags:     []string{"tag1", "tag2"},
				},
				{
					ID:        "todo2",
					Task:      "Second task",
					Status:    "completed",
					Priority:  "medium",
					Type:      "bug",
					Started:   time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
					Completed: time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
					Tags:      []string{"tag3"},
				},
			},
			wantContain: []string{
				`"id": "todo1"`,
				`"task": "First task"`,
				`"status": "in_progress"`,
				`"priority": "high"`,
				`"type": "feature"`,
				`"started": "2024-01-01T10:00:00Z"`,
				`"tags": [`,
				`"tag1"`,
				`"tag2"`,
				`"id": "todo2"`,
				`"task": "Second task"`,
				`"status": "completed"`,
				`"priority": "medium"`,
				`"type": "bug"`,
				`"started": "2024-01-02T10:00:00Z"`,
				`"completed": "2024-01-03T10:00:00Z"`,
				`"tag3"`,
			},
		},
		{
			name:  "empty todo list",
			todos: []*core.Todo{},
			wantContain: []string{
				"null", // json.MarshalIndent returns "null" for nil slice
			},
		},
		{
			name: "todo without completed time",
			todos: []*core.Todo{
				{
					ID:       "todo1",
					Task:     "Ongoing task",
					Status:   "in_progress",
					Priority: "low",
					Type:     "research",
					Started:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				},
			},
			wantContain: []string{
				`"id": "todo1"`,
				`"task": "Ongoing task"`,
				`"status": "in_progress"`,
			},
		},
		{
			name: "todo with nil tags",
			todos: []*core.Todo{
				{
					ID:       "todo1",
					Task:     "Task without tags",
					Status:   "blocked",
					Priority: "high",
					Type:     "feature",
					Started:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Tags:     nil,
				},
			},
			wantContain: []string{
				`"id": "todo1"`,
				`"tags": null`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTodosFull(tt.todos)

			// Verify result is not nil
			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			// Get the text content
			content := getResultContent(result)

			// Verify all expected strings are present
			for _, want := range tt.wantContain {
				if !strings.Contains(content, want) {
					t.Errorf("Result should contain %q, but got:\n%s", want, content)
				}
			}

			// Verify JSON is valid
			var parsed interface{}
			if err := json.Unmarshal([]byte(content), &parsed); err != nil {
				t.Errorf("Result should be valid JSON, but got error: %v\nContent: %s", err, content)
			}

			// For non-empty todos, verify completed field presence/absence
			if len(tt.todos) > 0 {
				for _, todo := range tt.todos {
					if todo.Completed.IsZero() {
						if strings.Contains(content, `"completed"`) && strings.Contains(content, todo.ID) {
							// More precise check needed
							var results []map[string]interface{}
							json.Unmarshal([]byte(content), &results)
							for _, r := range results {
								if r["id"] == todo.ID {
									if _, hasCompleted := r["completed"]; hasCompleted {
										t.Errorf("Todo %s should not have completed field", todo.ID)
									}
								}
							}
						}
					}
				}
			}
		})
	}
}

func TestFormatTodoSectionsResponse(t *testing.T) {
	tests := []struct {
		name           string
		todo           *core.Todo
		wantContain    []string
		wantNotContain []string
	}{
		{
			name: "todo with sections",
			todo: &core.Todo{
				ID: "test-todo",
				Sections: map[string]*core.SectionDefinition{
					"findings": {
						Title:    "Findings",
						Order:    1,
						Schema:   "research",
						Required: true,
					},
					"tests": {
						Title:  "Test Cases",
						Order:  2,
						Schema: "checklist",
						Metadata: map[string]interface{}{
							"test_framework": "jest",
						},
					},
					"custom_section": {
						Title:  "Custom Section",
						Order:  3,
						Schema: "freeform",
						Custom: true,
					},
				},
			},
			wantContain: []string{
				"Todo: test-todo",
				"Sections:",
				"findings:",
				"  title: Findings",
				"  order: 1",
				"  schema: research",
				"  required: true",
				"tests:",
				"  title: Test Cases",
				"  order: 2",
				"  schema: checklist",
				"  metadata:",
				"    test_framework: jest",
				"custom_section:",
				"  title: Custom Section",
				"  order: 3",
				"  custom: true",
			},
			wantNotContain: []string{
				"No section metadata defined",
			},
		},
		{
			name: "todo without sections",
			todo: &core.Todo{
				ID:       "legacy-todo",
				Sections: nil,
			},
			wantContain: []string{
				"Todo: legacy-todo",
				"Sections:",
				"No section metadata defined (legacy todo)",
			},
			wantNotContain: []string{
				"findings:",
				"tests:",
			},
		},
		{
			name: "todo with empty sections map",
			todo: &core.Todo{
				ID:       "empty-sections",
				Sections: map[string]*core.SectionDefinition{},
			},
			wantContain: []string{
				"Todo: empty-sections",
				"No section metadata defined (legacy todo)",
			},
		},
		{
			name: "todo with section without metadata",
			todo: &core.Todo{
				ID: "no-metadata",
				Sections: map[string]*core.SectionDefinition{
					"simple": {
						Title:    "Simple Section",
						Order:    1,
						Schema:   "research",
						Required: false,
						Metadata: nil,
					},
				},
			},
			wantContain: []string{
				"simple:",
				"  title: Simple Section",
				"  required: false",
			},
			wantNotContain: []string{
				"  metadata:",
				"  custom: true",
			},
		},
		{
			name: "todo with complex metadata",
			todo: &core.Todo{
				ID: "complex-meta",
				Sections: map[string]*core.SectionDefinition{
					"advanced": {
						Title:    "Advanced",
						Order:    1,
						Schema:   "custom",
						Required: true,
						Metadata: map[string]interface{}{
							"version": 2,
							"enabled": true,
							"config":  map[string]interface{}{"key": "value"},
							"tags":    []string{"a", "b"},
						},
					},
				},
			},
			wantContain: []string{
				"  metadata:",
				"    version: 2",
				"    enabled: true",
				"    config: map[key:value]",
				"    tags: [a b]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTodoSectionsResponse(tt.todo)

			// Verify result is not nil
			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			// Get the text content
			content := getResultContent(result)

			// Verify all expected strings are present
			for _, want := range tt.wantContain {
				if !strings.Contains(content, want) {
					t.Errorf("Result should contain %q, but got:\n%s", want, content)
				}
			}

			// Verify strings that should not be present
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(content, notWant) {
					t.Errorf("Result should not contain %q, but got:\n%s", notWant, content)
				}
			}
		})
	}
}

func TestFormatTodoSectionsResponse_Ordering(t *testing.T) {
	// Test that sections are ordered correctly
	todo := &core.Todo{
		ID: "ordered-todo",
		Sections: map[string]*core.SectionDefinition{
			"third": {
				Title:  "Third Section",
				Order:  3,
				Schema: "research",
			},
			"first": {
				Title:  "First Section",
				Order:  1,
				Schema: "research",
			},
			"second": {
				Title:  "Second Section",
				Order:  2,
				Schema: "research",
			},
		},
	}

	result := FormatTodoSectionsResponse(todo)
	content := getResultContent(result)

	// Find positions of each section
	firstPos := strings.Index(content, "first:")
	secondPos := strings.Index(content, "second:")
	thirdPos := strings.Index(content, "third:")

	// Verify ordering
	if firstPos == -1 || secondPos == -1 || thirdPos == -1 {
		t.Fatalf("Not all sections found in output: %s", content)
	}

	if firstPos > secondPos || secondPos > thirdPos {
		t.Errorf("Sections not in correct order. Positions: first=%d, second=%d, third=%d\nContent:\n%s",
			firstPos, secondPos, thirdPos, content)
	}
}

// Helper function to extract content from MCP result
func getResultContent(result interface{}) string {
	// Use reflection to access the private content field
	// In a real scenario, we'd use the proper interface methods
	if textResult, ok := result.(*struct{ content []interface{} }); ok {
		if len(textResult.content) > 0 {
			if text, ok := textResult.content[0].(map[string]interface{}); ok {
				if content, ok := text["text"].(string); ok {
					return content
				}
			}
		}
	}

	// Alternative approach: convert to JSON and extract
	jsonBytes, _ := json.Marshal(result)
	var data map[string]interface{}
	json.Unmarshal(jsonBytes, &data)

	if contents, ok := data["content"].([]interface{}); ok && len(contents) > 0 {
		if firstContent, ok := contents[0].(map[string]interface{}); ok {
			if text, ok := firstContent["text"].(string); ok {
				return text
			}
		}
	}

	return string(jsonBytes)
}
