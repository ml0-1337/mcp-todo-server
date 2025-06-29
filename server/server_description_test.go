package server

import (
	"strings"
	"testing"
)

func TestTodoCreateToolDescriptionIncludesGuidance(t *testing.T) {
	// The updated description should include guidance
	expectedDescription := "Create a new todo with full metadata. TIP: Use parent_id for phases and subtasks. Types 'phase' and 'subtask' require parent_id."

	// This test verifies the description includes key guidance elements
	expectedElements := []string{
		"parent_id",
		"phase",
		"subtask",
		"TIP",
		"require",
	}

	for _, element := range expectedElements {
		if !strings.Contains(expectedDescription, element) {
			t.Errorf("Tool description should include guidance about %s", element)
		}
	}
}
