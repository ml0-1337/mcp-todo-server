package core

import (
	"reflect"
	"testing"
)

func TestParseChecklist(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []ChecklistItem
	}{
		{
			name: "parse all three states",
			content: `- [ ] Pending task
- [>] In progress task
- [x] Completed task`,
			expected: []ChecklistItem{
				{Text: "Pending task", Status: "pending"},
				{Text: "In progress task", Status: "in_progress"},
				{Text: "Completed task", Status: "completed"},
			},
		},
		{
			name: "parse multiple in-progress markers",
			content: `- [>] Arrow marker
- [-] Dash marker
- [~] Tilde marker`,
			expected: []ChecklistItem{
				{Text: "Arrow marker", Status: "in_progress"},
				{Text: "Dash marker", Status: "in_progress"},
				{Text: "Tilde marker", Status: "in_progress"},
			},
		},
		{
			name: "handle uppercase X",
			content: `- [X] Uppercase completed
- [x] Lowercase completed`,
			expected: []ChecklistItem{
				{Text: "Uppercase completed", Status: "completed"},
				{Text: "Lowercase completed", Status: "completed"},
			},
		},
		{
			name: "ignore non-checklist lines",
			content: `Some regular text
- [ ] Valid checklist item
Another line
- [x] Another valid item
- Invalid item without checkbox`,
			expected: []ChecklistItem{
				{Text: "Valid checklist item", Status: "pending"},
				{Text: "Another valid item", Status: "completed"},
			},
		},
		{
			name: "handle empty checklist items",
			content: `- [ ] 
- [x] Valid item
- [ ]`,
			expected: []ChecklistItem{
				{Text: "Valid item", Status: "completed"},
			},
		},
		{
			name: "handle special characters in text",
			content: `- [ ] Task with [brackets] and (parens)
- [x] Task with "quotes" and 'apostrophes'
- [>] Task with émojis 🚀 and unicode`,
			expected: []ChecklistItem{
				{Text: "Task with [brackets] and (parens)", Status: "pending"},
				{Text: "Task with \"quotes\" and 'apostrophes'", Status: "completed"},
				{Text: "Task with émojis 🚀 and unicode", Status: "in_progress"},
			},
		},
		{
			name: "mixed content with indentation",
			content: `## Checklist Section
- [ ] Main task
  - [ ] This is also parsed (indented)
- [x] Another main task
  Some description text
- [>] Task in progress`,
			expected: []ChecklistItem{
				{Text: "Main task", Status: "pending"},
				{Text: "This is also parsed (indented)", Status: "pending"},
				{Text: "Another main task", Status: "completed"},
				{Text: "Task in progress", Status: "in_progress"},
			},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []ChecklistItem{},
		},
		{
			name:     "whitespace only",
			content:  "   \n\t\n   ",
			expected: []ChecklistItem{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseChecklist(tt.content)

			if len(result) != len(tt.expected) {
				t.Errorf("ParseChecklist() returned %d items, want %d", len(result), len(tt.expected))
				return
			}

			for i, item := range result {
				if !reflect.DeepEqual(item, tt.expected[i]) {
					t.Errorf("ParseChecklist() item[%d] = %+v, want %+v", i, item, tt.expected[i])
				}
			}
		})
	}
}

func TestParseChecklistEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []ChecklistItem
	}{
		{
			name: "malformed checkboxes ignored",
			content: `- [] Missing space
- [ Missing closing bracket
- [abc] Invalid marker
- [ ] Valid item
-[x] Missing space before bracket`,
			expected: []ChecklistItem{
				{Text: "Valid item", Status: "pending"},
			},
		},
		{
			name:    "very long text",
			content: "- [ ] " + string(make([]byte, 1000, 1000)) + "very long text",
			expected: []ChecklistItem{
				{Text: string(make([]byte, 1000, 1000)) + "very long text", Status: "pending"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseChecklist(tt.content)

			if len(result) != len(tt.expected) {
				t.Errorf("ParseChecklist() returned %d items, want %d", len(result), len(tt.expected))
				return
			}

			// For very long text test, just check status and length
			if tt.name == "very long text" && len(result) > 0 {
				if result[0].Status != tt.expected[0].Status {
					t.Errorf("ParseChecklist() status = %s, want %s", result[0].Status, tt.expected[0].Status)
				}
				if len(result[0].Text) != len(tt.expected[0].Text) {
					t.Errorf("ParseChecklist() text length = %d, want %d", len(result[0].Text), len(tt.expected[0].Text))
				}
			} else {
				for i, item := range result {
					if !reflect.DeepEqual(item, tt.expected[i]) {
						t.Errorf("ParseChecklist() item[%d] = %+v, want %+v", i, item, tt.expected[i])
					}
				}
			}
		})
	}
}
