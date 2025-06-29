package core

import (
	"testing"
)

func TestToggleChecklistItem(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		itemText string
		expected string
	}{
		{
			name: "toggle pending to in_progress",
			content: `- [ ] Task one
- [ ] Task two
- [x] Task three`,
			itemText: "Task two",
			expected: `- [ ] Task one
- [>] Task two
- [x] Task three`,
		},
		{
			name: "toggle in_progress to completed",
			content: `- [ ] Task one
- [>] Task two
- [x] Task three`,
			itemText: "Task two",
			expected: `- [ ] Task one
- [x] Task two
- [x] Task three`,
		},
		{
			name: "toggle completed to pending",
			content: `- [ ] Task one
- [x] Task two
- [x] Task three`,
			itemText: "Task two",
			expected: `- [ ] Task one
- [ ] Task two
- [x] Task three`,
		},
		{
			name: "toggle with different in_progress markers",
			content: `- [-] Dash marker
- [~] Tilde marker`,
			itemText: "Dash marker",
			expected: `- [x] Dash marker
- [~] Tilde marker`,
		},
		{
			name: "toggle uppercase X",
			content: `- [X] Uppercase task`,
			itemText: "Uppercase task",
			expected: `- [ ] Uppercase task`,
		},
		{
			name: "handle item not found",
			content: `- [ ] Task one
- [ ] Task two`,
			itemText: "Task three",
			expected: `- [ ] Task one
- [ ] Task two`,
		},
		{
			name: "exact match required",
			content: `- [ ] Task one
- [ ] Task one extended`,
			itemText: "Task one",
			expected: `- [>] Task one
- [ ] Task one extended`,
		},
		{
			name: "preserve indentation",
			content: `  - [ ] Indented task
- [ ] Normal task`,
			itemText: "Indented task",
			expected: `  - [>] Indented task
- [ ] Normal task`,
		},
		{
			name: "handle special characters",
			content: `- [ ] Task with [brackets]
- [ ] Task with "quotes"`,
			itemText: "Task with [brackets]",
			expected: `- [>] Task with [brackets]
- [ ] Task with "quotes"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toggleChecklistItem(tt.content, tt.itemText)
			if result != tt.expected {
				t.Errorf("toggleChecklistItem() = %q, want %q", result, tt.expected)
			}
		})
	}
}