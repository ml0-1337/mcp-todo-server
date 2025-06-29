package core

import (
	"testing"
)

func TestDetectPattern_PhasePattern(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		wantHint bool
		wantType string
	}{
		{
			name:     "Phase with number",
			title:    "Phase 1: Core Implementation",
			wantHint: true,
			wantType: "phase",
		},
		{
			name:     "Phase lowercase",
			title:    "phase 2: API Endpoints",
			wantHint: true,
			wantType: "phase",
		},
		{
			name:     "Phase without colon",
			title:    "Phase 3 Migration Tools",
			wantHint: true,
			wantType: "phase",
		},
		{
			name:     "Phase with decimal",
			title:    "Phase 2.5: Intermediate Step",
			wantHint: true,
			wantType: "phase",
		},
		{
			name:     "Not a phase",
			title:    "Implement user authentication",
			wantHint: false,
			wantType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := DetectPattern(tt.title)
			
			if tt.wantHint && hint == nil {
				t.Errorf("Expected pattern hint for %q, got nil", tt.title)
			}
			
			if !tt.wantHint && hint != nil {
				t.Errorf("Expected no pattern hint for %q, got %+v", tt.title, hint)
			}
			
			if tt.wantHint && hint != nil {
				if hint.SuggestedType != tt.wantType {
					t.Errorf("Expected type %q, got %q", tt.wantType, hint.SuggestedType)
				}
				
				if hint.Pattern != "phase" {
					t.Errorf("Expected pattern 'phase', got %q", hint.Pattern)
				}
			}
		})
	}
}

func TestDetectPattern_PartPattern(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		wantHint bool
		wantType string
	}{
		{
			name:     "Part N of M",
			title:    "Part 1 of 3: Setup Infrastructure",
			wantHint: true,
			wantType: "phase",
		},
		{
			name:     "Part without of",
			title:    "Part 2: Implementation",
			wantHint: true,
			wantType: "phase",
		},
		{
			name:     "Not a part",
			title:    "Partial implementation of feature",
			wantHint: false,
			wantType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := DetectPattern(tt.title)
			
			if tt.wantHint && hint == nil {
				t.Errorf("Expected pattern hint for %q, got nil", tt.title)
			}
			
			if !tt.wantHint && hint != nil {
				t.Errorf("Expected no pattern hint for %q, got %+v", tt.title, hint)
			}
			
			if tt.wantHint && hint != nil && hint.SuggestedType != tt.wantType {
				t.Errorf("Expected type %q, got %q", tt.wantType, hint.SuggestedType)
			}
		})
	}
}

func TestDetectPattern_StepPattern(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		wantHint bool
		wantType string
	}{
		{
			name:     "Step with number",
			title:    "Step 1: Initialize Database",
			wantHint: true,
			wantType: "subtask",
		},
		{
			name:     "Step without colon",
			title:    "Step 2 Configure Settings",
			wantHint: true,
			wantType: "subtask",
		},
		{
			name:     "Not a step",
			title:    "Next steps for project",
			wantHint: false,
			wantType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := DetectPattern(tt.title)
			
			if tt.wantHint && hint == nil {
				t.Errorf("Expected pattern hint for %q, got nil", tt.title)
			}
			
			if !tt.wantHint && hint != nil {
				t.Errorf("Expected no pattern hint for %q, got %+v", tt.title, hint)
			}
			
			if tt.wantHint && hint != nil && hint.SuggestedType != tt.wantType {
				t.Errorf("Expected type %q, got %q", tt.wantType, hint.SuggestedType)
			}
		})
	}
}

func TestDetectPattern_NumberedPattern(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		wantHint bool
		wantType string
	}{
		{
			name:     "Bracketed number",
			title:    "[1] Setup development environment",
			wantHint: true,
			wantType: "subtask",
		},
		{
			name:     "Number with dot",
			title:    "1. Configure database",
			wantHint: true,
			wantType: "subtask",
		},
		{
			name:     "Number with parenthesis",
			title:    "1) Install dependencies",
			wantHint: true,
			wantType: "subtask",
		},
		{
			name:     "Not numbered",
			title:    "Update configuration files",
			wantHint: false,
			wantType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := DetectPattern(tt.title)
			
			if tt.wantHint && hint == nil {
				t.Errorf("Expected pattern hint for %q, got nil", tt.title)
			}
			
			if !tt.wantHint && hint != nil {
				t.Errorf("Expected no pattern hint for %q, got %+v", tt.title, hint)
			}
			
			if tt.wantHint && hint != nil && hint.SuggestedType != tt.wantType {
				t.Errorf("Expected type %q, got %q", tt.wantType, hint.SuggestedType)
			}
		})
	}
}

func TestDetectPattern_HintMessage(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		wantMessage string
	}{
		{
			name:        "Phase pattern hint",
			title:       "Phase 2: Implementation",
			wantMessage: "This looks like a phase. Consider using type 'phase' with a parent_id.",
		},
		{
			name:        "Step pattern hint",
			title:       "Step 3: Deploy to production",
			wantMessage: "This looks like a step. Consider using type 'subtask' with a parent_id.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := DetectPattern(tt.title)
			
			if hint == nil {
				t.Fatal("Expected pattern hint, got nil")
			}
			
			if hint.Message != tt.wantMessage {
				t.Errorf("Expected message %q, got %q", tt.wantMessage, hint.Message)
			}
		})
	}
}

func TestFindSimilarTodos(t *testing.T) {
	existingTodos := []*Todo{
		{ID: "phase-1-planning", Task: "Phase 1: Planning"},
		{ID: "phase-2-implementation", Task: "Phase 2: Implementation"},
		{ID: "unrelated-todo", Task: "Fix bug in login"},
		{ID: "another-phase", Task: "Phase 3: Testing"},
		{ID: "step-1", Task: "Step 1: Setup"},
	}
	
	tests := []struct {
		name     string
		title    string
		wantIDs  []string
	}{
		{
			name:    "Find similar phase todos",
			title:   "Phase 4: Deployment",
			wantIDs: []string{"phase-1-planning", "phase-2-implementation", "another-phase"},
		},
		{
			name:    "Find similar step todos",
			title:   "Step 2: Configuration",
			wantIDs: []string{"step-1"},
		},
		{
			name:    "No similar todos",
			title:   "Implement new feature",
			wantIDs: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similar := FindSimilarTodos(existingTodos, tt.title)
			
			if len(similar) != len(tt.wantIDs) {
				t.Errorf("Expected %d similar todos, got %d: %v", len(tt.wantIDs), len(similar), similar)
			}
			
			// Check that all expected IDs are present
			for _, wantID := range tt.wantIDs {
				found := false
				for _, id := range similar {
					if id == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find %q in similar todos", wantID)
				}
			}
		})
	}
}