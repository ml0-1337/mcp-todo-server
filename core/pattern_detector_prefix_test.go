package core

import "testing"

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		title      string
		wantPrefix string
	}{
		{"Phase 1: Planning", "Phase"},
		{"Phase 2: Implementation", "Phase"},
		{"Step 1: Setup", "Step"},
		{"Part 1 of 3: Introduction", "Part"},
		{"Regular todo without pattern", ""},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := extractPrefix(tt.title)
			if got != tt.wantPrefix {
				t.Errorf("extractPrefix(%q) = %q, want %q", tt.title, got, tt.wantPrefix)
			}
		})
	}
}