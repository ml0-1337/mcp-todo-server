package core

import (
	"testing"
)

func TestParseChecklistAdvancedEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []ChecklistItem
	}{
		{
			name:     "empty content",
			content:  "",
			expected: []ChecklistItem{},
		},
		{
			name:     "only whitespace",
			content:  "   \n\t\n   ",
			expected: []ChecklistItem{},
		},
		{
			name:     "no checklist items",
			content:  "This is just regular text\nNo checkboxes here",
			expected: []ChecklistItem{},
		},
		{
			name: "mixed content with checklist",
			content: `Some intro text
- [ ] Valid item
Regular bullet point
- [x] Another valid item
* Not a checklist item`,
			expected: []ChecklistItem{
				{Text: "Valid item", Status: "pending"},
				{Text: "Another valid item", Status: "completed"},
			},
		},
		{
			name: "malformed checkbox syntax",
			content: `- [] Missing space
- [ Extra space at start
-[ ] No space after dash
- [y] Invalid marker
- [ ] Valid item`,
			expected: []ChecklistItem{
				{Text: "Valid item", Status: "pending"},
			},
		},
		{
			name: "empty checkbox items",
			content: `- [ ]
- [ ] 
- [ ] Valid item
- [x]`,
			expected: []ChecklistItem{
				{Text: "Valid item", Status: "pending"},
			},
		},
		{
			name: "special characters in text",
			content: `- [ ] Item with [brackets]
- [x] Item with "quotes"
- [>] Item with <angles>
- [ ] Item with & ampersand`,
			expected: []ChecklistItem{
				{Text: "Item with [brackets]", Status: "pending"},
				{Text: `Item with "quotes"`, Status: "completed"},
				{Text: "Item with <angles>", Status: "in_progress"},
				{Text: "Item with & ampersand", Status: "pending"},
			},
		},
		{
			name: "unicode in text",
			content: `- [ ] 日本語のタスク
- [x] Émojis 🎉🎊
- [-] Ñoño español`,
			expected: []ChecklistItem{
				{Text: "日本語のタスク", Status: "pending"},
				{Text: "Émojis 🎉🎊", Status: "completed"},
				{Text: "Ñoño español", Status: "in_progress"},
			},
		},
		{
			name: "very long item text",
			content: `- [ ] This is a very long checklist item that contains a lot of text and might wrap to multiple lines in some editors but should still be parsed correctly as a single item
- [x] Short item`,
			expected: []ChecklistItem{
				{Text: "This is a very long checklist item that contains a lot of text and might wrap to multiple lines in some editors but should still be parsed correctly as a single item", Status: "pending"},
				{Text: "Short item", Status: "completed"},
			},
		},
		{
			name: "nested lists parsed as separate items",
			content: `- [ ] Top level item
  - [ ] Nested item (indented)
- [x] Another top level`,
			expected: []ChecklistItem{
				{Text: "Top level item", Status: "pending"},
				{Text: "Nested item (indented)", Status: "pending"},
				{Text: "Another top level", Status: "completed"},
			},
		},
		{
			name: "extra whitespace handling",
			content: `- [ ]    Extra spaces before text   
- [x]  	Mixed tabs and spaces	
- [>]Immediate text after bracket`,
			expected: []ChecklistItem{
				{Text: "Extra spaces before text", Status: "pending"},
				{Text: "Mixed tabs and spaces", Status: "completed"},
				{Text: "Immediate text after bracket", Status: "in_progress"},
			},
		},
		{
			name: "all supported markers",
			content: `- [ ] Pending
- [x] Completed lowercase
- [X] Completed uppercase
- [>] In progress arrow
- [-] In progress dash
- [~] In progress tilde`,
			expected: []ChecklistItem{
				{Text: "Pending", Status: "pending"},
				{Text: "Completed lowercase", Status: "completed"},
				{Text: "Completed uppercase", Status: "completed"},
				{Text: "In progress arrow", Status: "in_progress"},
				{Text: "In progress dash", Status: "in_progress"},
				{Text: "In progress tilde", Status: "in_progress"},
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
			
			for i, item := range result {
				if item.Text != tt.expected[i].Text {
					t.Errorf("Item %d: got text %q, want %q", i, item.Text, tt.expected[i].Text)
				}
				if item.Status != tt.expected[i].Status {
					t.Errorf("Item %d: got status %q, want %q", i, item.Status, tt.expected[i].Status)
				}
			}
		})
	}
}

func TestToggleChecklistItemEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		itemText string
		expected string
	}{
		{
			name:     "item not found",
			content:  "- [ ] Item one\n- [ ] Item two",
			itemText: "Item three",
			expected: "- [ ] Item one\n- [ ] Item two", // No change
		},
		{
			name:     "empty content",
			content:  "",
			itemText: "Any item",
			expected: "",
		},
		{
			name:     "empty item text",
			content:  "- [ ] Item one",
			itemText: "",
			expected: "- [ ] Item one", // No change
		},
		{
			name: "case sensitive matching",
			content: `- [ ] Item One
- [ ] item one`,
			itemText: "item one",
			expected: `- [ ] Item One
- [>] item one`, // Only exact match toggled
		},
		{
			name: "partial match not toggled",
			content: `- [ ] Complete task
- [ ] Complete task with details`,
			itemText: "Complete task",
			expected: `- [>] Complete task
- [ ] Complete task with details`,
		},
		{
			name: "whitespace in item text",
			content: `- [ ] Item with  extra   spaces`,
			itemText: "Item with  extra   spaces",
			expected: `- [>] Item with  extra   spaces`,
		},
		{
			name: "special regex characters",
			content: `- [ ] Item with $pecial (chars) [test]`,
			itemText: "Item with $pecial (chars) [test]",
			expected: `- [>] Item with $pecial (chars) [test]`,
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