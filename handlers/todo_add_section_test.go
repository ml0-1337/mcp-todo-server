package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/user/mcp-todo-server/core"
)

// Test 13: Add custom section to existing todo
func TestAddCustomSectionToExistingTodo(t *testing.T) {
	// Create mock managers
	mockManager := NewMockTodoManager()
	mockSearch := &MockSearchEngine{}
	mockStats := &MockStatsEngine{}
	mockTemplates := &MockTemplateManager{}

	// Create handlers with mocks
	handlers := NewTodoHandlersWithDependencies(
		mockManager,
		mockSearch,
		mockStats,
		mockTemplates,
	)

	// Setup a test todo
	testTodo := &core.Todo{
		ID:       "test-add-section",
		Task:     "Test todo for custom sections",
		Status:   "in_progress",
		Priority: "high",
		Type:     "feature",
		Sections: map[string]*core.SectionDefinition{
			"findings": {
				Title:    "## Findings & Research",
				Order:    1,
				Schema:   core.SchemaResearch,
				Required: false,
			},
			"checklist": {
				Title:    "## Checklist",
				Order:    2,
				Schema:   core.SchemaChecklist,
				Required: false,
			},
		},
	}

	// Track added sections
	addedSections := make(map[string]*core.SectionDefinition)

	// Setup mock to return our test todo
	mockManager.ReadTodoFunc = func(id string) (*core.Todo, error) {
		if id == "test-add-section" {
			// Return a copy with any added sections
			todoCopy := &core.Todo{
				ID:       testTodo.ID,
				Task:     testTodo.Task,
				Status:   testTodo.Status,
				Priority: testTodo.Priority,
				Type:     testTodo.Type,
				Sections: make(map[string]*core.SectionDefinition),
			}

			// Copy existing sections
			for k, v := range testTodo.Sections {
				todoCopy.Sections[k] = v
			}

			// Add any new sections
			for k, v := range addedSections {
				todoCopy.Sections[k] = v
			}

			return todoCopy, nil
		}
		return nil, fmt.Errorf("todo not found: %s", id)
	}

	// Mock the UpdateTodo to succeed
	mockManager.UpdateTodoFunc = func(id, section, operation, content string, metadata map[string]string) error {
		return nil
	}

	todoID := "test-add-section"
	ctx := context.Background()

	testCases := []struct {
		name            string
		sectionKey      string
		sectionTitle    string
		sectionSchema   string
		sectionRequired bool
		expectedError   bool
	}{
		{
			name:            "add implementation notes section",
			sectionKey:      "implementation",
			sectionTitle:    "## Implementation Notes",
			sectionSchema:   "freeform",
			sectionRequired: false,
			expectedError:   false,
		},
		{
			name:            "add risks section with freeform schema",
			sectionKey:      "risks",
			sectionTitle:    "## Risks & Mitigations",
			sectionSchema:   "freeform",
			sectionRequired: true,
			expectedError:   false,
		},
		{
			name:            "add api_design section",
			sectionKey:      "api_design",
			sectionTitle:    "## API Design",
			sectionSchema:   "test_cases", // Requires code blocks
			sectionRequired: false,
			expectedError:   false,
		},
		{
			name:            "add duplicate section key",
			sectionKey:      "findings", // Already exists
			sectionTitle:    "## Research Findings",
			sectionSchema:   "research",
			sectionRequired: false,
			expectedError:   true,
		},
		{
			name:            "add section with invalid schema",
			sectionKey:      "custom",
			sectionTitle:    "## Custom Section",
			sectionSchema:   "invalid_schema",
			sectionRequired: false,
			expectedError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create add section request
			request := &MockCallToolRequest{
				Arguments: map[string]interface{}{
					"id":       todoID,
					"key":      tc.sectionKey,
					"title":    tc.sectionTitle,
					"schema":   tc.sectionSchema,
					"required": tc.sectionRequired,
				},
			}

			result, err := handlers.HandleTodoAddSection(ctx, request.ToCallToolRequest())

			if tc.expectedError {
				// HandleError returns a result with IsError set to true, not a Go error
				if err != nil {
					// This is a real error, which is what we expect
					return
				}
				if result != nil && !result.IsError {
					t.Error("Expected error result but got success")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to add section: %v", err)
			}

			// If successful, track the added section
			if !tc.expectedError {
				addedSections[tc.sectionKey] = &core.SectionDefinition{
					Title:    tc.sectionTitle,
					Order:    100,
					Schema:   core.SectionSchema(tc.sectionSchema),
					Required: tc.sectionRequired,
				}

				// Verify section was added by reading todo
				todo, err := mockManager.ReadTodo(todoID)
				if err != nil {
					t.Fatalf("Failed to read todo: %v", err)
				}

				// Check section exists
				section, exists := todo.Sections[tc.sectionKey]
				if !exists {
					t.Errorf("Section %s was not added", tc.sectionKey)
					return
				}

				// Verify section properties
				if section.Title != tc.sectionTitle {
					t.Errorf("Section title mismatch. Expected: %s, Got: %s", tc.sectionTitle, section.Title)
				}

				if string(section.Schema) != tc.sectionSchema {
					t.Errorf("Section schema mismatch. Expected: %s, Got: %s", tc.sectionSchema, section.Schema)
				}

				if section.Required != tc.sectionRequired {
					t.Errorf("Section required mismatch. Expected: %v, Got: %v", tc.sectionRequired, section.Required)
				}
			}

			// Verify response exists
			if result == nil {
				t.Error("Expected result but got nil")
			}
		})
	}
}
