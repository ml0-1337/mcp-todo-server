package handlers

import (
	"strings"
	"testing"
)

func TestGetUpdatePrompts(t *testing.T) {
	tests := []struct {
		name      string
		section   string
		operation string
		todoType  string
		wantWords []string // Key words/phrases that should appear in the prompt
	}{
		{
			name:      "findings append",
			section:   "findings",
			operation: "append",
			todoType:  "feature",
			wantWords: []string{"Content added to findings", "why", "implications", "follow-up questions"},
		},
		{
			name:      "findings replace",
			section:   "findings",
			operation: "replace",
			todoType:  "feature",
			wantWords: []string{"Findings updated", "essential insights", "implications", "unexpected discoveries"},
		},
		{
			name:      "tests append for bug",
			section:   "tests",
			operation: "append",
			todoType:  "bug",
			wantWords: []string{"Test added", "prevent regression", "reproduce the bug", "edge cases"},
		},
		{
			name:      "tests replace for bug",
			section:   "tests",
			operation: "replace",
			todoType:  "bug",
			wantWords: []string{"Test updated", "fail without the fix", "catch similar bugs", "exact failure condition"},
		},
		{
			name:      "tests append for feature",
			section:   "tests",
			operation: "append",
			todoType:  "feature",
			wantWords: []string{"Test added", "comprehensive coverage", "expected behavior", "edge cases"},
		},
		{
			name:      "checklist toggle",
			section:   "checklist",
			operation: "toggle",
			todoType:  "feature",
			wantWords: []string{"Checklist item toggled", "reveal new tasks", "priorities", "next highest priority"},
		},
		{
			name:      "checklist append",
			section:   "checklist",
			operation: "append",
			todoType:  "feature",
			wantWords: []string{"Checklist updated", "maintain momentum", "specific and actionable", "next unchecked item"},
		},
		{
			name:      "scratchpad update",
			section:   "scratchpad",
			operation: "append",
			todoType:  "feature",
			wantWords: []string{"Scratchpad updated", "move to findings", "action items", "rapid capture"},
		},
		{
			name:      "custom section append",
			section:   "custom_section",
			operation: "append",
			todoType:  "feature",
			wantWords: []string{"Content added", "maintain quality", "enhance", "well-structured"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getUpdatePrompts(tt.section, tt.operation, tt.todoType)
			
			// Check that prompt is not empty
			if got == "" {
				t.Errorf("getUpdatePrompts() returned empty string")
			}
			
			// Check that all expected words/phrases appear
			for _, word := range tt.wantWords {
				if !strings.Contains(got, word) {
					t.Errorf("getUpdatePrompts() missing expected phrase %q\nGot: %s", word, got)
				}
			}
		})
	}
}

func TestUpdateHandlerWithPrompts(t *testing.T) {
	// This test would require mocking the manager and verifying the full response
	// For now, we're testing the prompt generation separately
	t.Skip("Integration test - implement with mock manager")
}