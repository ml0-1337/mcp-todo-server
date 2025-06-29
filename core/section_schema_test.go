package core

import (
	"testing"
)

// Test 1: Parse section definitions from YAML frontmatter
func TestParseSectionDefinitionsFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected map[string]*SectionDefinition
		wantErr  bool
	}{
		{
			name: "parse basic section definitions",
			yaml: `todo_id: "test-todo"
started: "2025-06-29 15:00:00"
status: "in_progress"
sections:
  findings:
    title: "## Findings & Research"
    order: 1
    schema: "research"
    required: true
  test_list:
    title: "## Test List"
    order: 2
    schema: "checklist"
    required: true`,
			expected: map[string]*SectionDefinition{
				"findings": {
					Title:    "## Findings & Research",
					Order:    1,
					Schema:   SchemaResearch,
					Required: true,
				},
				"test_list": {
					Title:    "## Test List",
					Order:    2,
					Schema:   SchemaChecklist,
					Required: true,
				},
			},
			wantErr: false,
		},
		{
			name: "parse section with metadata",
			yaml: `sections:
  test_list:
    title: "## Test List"
    order: 1
    schema: "checklist"
    required: true
    metadata:
      completed: 4
      total: 8`,
			expected: map[string]*SectionDefinition{
				"test_list": {
					Title:    "## Test List",
					Order:    1,
					Schema:   SchemaChecklist,
					Required: true,
					Metadata: map[string]interface{}{
						"completed": 4,
						"total":     8,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "parse custom section",
			yaml: `sections:
  custom_security:
    title: "## Security Analysis"
    order: 9
    schema: "freeform"
    required: false
    custom: true`,
			expected: map[string]*SectionDefinition{
				"custom_security": {
					Title:    "## Security Analysis",
					Order:    9,
					Schema:   SchemaFreeform,
					Required: false,
					Custom:   true,
				},
			},
			wantErr: false,
		},
		{
			name: "no sections defined (backwards compatibility)",
			yaml: `todo_id: "legacy-todo"
started: "2025-06-29 15:00:00"
status: "in_progress"`,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "invalid schema type",
			yaml: `sections:
  bad_section:
    title: "## Bad Section"
    order: 1
    schema: "invalid_schema"
    required: true`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse YAML to get sections
			sections, err := ParseSectionDefinitions([]byte(tt.yaml))
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSectionDefinitions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr {
				return
			}
			
			// Check if sections match expected
			if len(sections) != len(tt.expected) {
				t.Errorf("ParseSectionDefinitions() got %d sections, want %d", len(sections), len(tt.expected))
				return
			}
			
			for key, expectedDef := range tt.expected {
				gotDef, exists := sections[key]
				if !exists {
					t.Errorf("ParseSectionDefinitions() missing section %s", key)
					continue
				}
				
				// Compare fields
				if gotDef.Title != expectedDef.Title {
					t.Errorf("Section %s: Title = %v, want %v", key, gotDef.Title, expectedDef.Title)
				}
				if gotDef.Order != expectedDef.Order {
					t.Errorf("Section %s: Order = %v, want %v", key, gotDef.Order, expectedDef.Order)
				}
				if gotDef.Schema != expectedDef.Schema {
					t.Errorf("Section %s: Schema = %v, want %v", key, gotDef.Schema, expectedDef.Schema)
				}
				if gotDef.Required != expectedDef.Required {
					t.Errorf("Section %s: Required = %v, want %v", key, gotDef.Required, expectedDef.Required)
				}
				if gotDef.Custom != expectedDef.Custom {
					t.Errorf("Section %s: Custom = %v, want %v", key, gotDef.Custom, expectedDef.Custom)
				}
				
				// Compare metadata if present
				if expectedDef.Metadata != nil {
					if gotDef.Metadata == nil {
						t.Errorf("Section %s: Metadata is nil, want %v", key, expectedDef.Metadata)
					} else {
						for mk, mv := range expectedDef.Metadata {
							if gotDef.Metadata[mk] != mv {
								t.Errorf("Section %s: Metadata[%s] = %v, want %v", key, mk, gotDef.Metadata[mk], mv)
							}
						}
					}
				}
			}
		})
	}
}