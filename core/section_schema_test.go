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

// Test 2: Validate section schema types
func TestValidateSectionSchemaTypes(t *testing.T) {
	tests := []struct {
		name    string
		schema  SectionSchema
		content string
		wantErr bool
		errMsg  string
	}{
		// Research schema tests
		{
			name:    "research schema accepts any text",
			schema:  SchemaResearch,
			content: "Some research findings with references [1] and citations.",
			wantErr: false,
		},
		{
			name:    "research schema accepts empty content",
			schema:  SchemaResearch,
			content: "",
			wantErr: false,
		},
		// Checklist schema tests
		{
			name:    "checklist schema accepts valid checkboxes",
			schema:  SchemaChecklist,
			content: "- [ ] Task 1\n- [x] Task 2\n- [ ] Task 3",
			wantErr: false,
		},
		{
			name:    "checklist schema rejects invalid checkbox syntax",
			schema:  SchemaChecklist,
			content: "- [] Missing space\n- [x] Valid\n- [y] Invalid marker",
			wantErr: true,
			errMsg:  "invalid checkbox syntax",
		},
		{
			name:    "checklist schema rejects non-list items",
			schema:  SchemaChecklist,
			content: "This is not a checklist\n- [x] Mixed content",
			wantErr: true,
			errMsg:  "non-checklist content found",
		},
		// Test cases schema tests
		{
			name:    "test_cases schema accepts code blocks",
			schema:  SchemaTestCases,
			content: "```go\nfunc TestExample(t *testing.T) {\n\t// test code\n}\n```",
			wantErr: false,
		},
		{
			name:    "test_cases schema accepts multiple languages",
			schema:  SchemaTestCases,
			content: "```python\ndef test_example():\n    pass\n```\n\n```javascript\ntest('example', () => {});\n```",
			wantErr: false,
		},
		{
			name:    "test_cases schema warns on missing code blocks",
			schema:  SchemaTestCases,
			content: "Just some text without code blocks",
			wantErr: true,
			errMsg:  "no code blocks found",
		},
		// Results schema tests
		{
			name:    "results schema accepts timestamped entries",
			schema:  SchemaResults,
			content: "[2025-01-01 10:00:00] Test started\n[2025-01-01 10:01:00] Test passed",
			wantErr: false,
		},
		{
			name:    "results schema rejects entries without timestamps",
			schema:  SchemaResults,
			content: "Test started\n[2025-01-01 10:01:00] Test passed",
			wantErr: true,
			errMsg:  "entries must start with timestamp",
		},
		// Strategy schema tests
		{
			name:    "strategy schema accepts structured content",
			schema:  SchemaStrategy,
			content: "### Approach\n- Step 1\n- Step 2\n\n### Risks\n- Risk 1",
			wantErr: false,
		},
		// Freeform schema tests
		{
			name:    "freeform schema accepts anything",
			schema:  SchemaFreeform,
			content: "Any content\n```\ncode\n```\n- [ ] checkbox\n[timestamp] log",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := GetValidator(tt.schema)
			if validator == nil {
				t.Fatalf("GetValidator(%v) returned nil", tt.schema)
			}
			
			err := validator.Validate(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, want error containing %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// Test 3: Preserve section order from metadata
func TestPreserveSectionOrder(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		expectedOrder []string // Expected section keys in order
	}{
		{
			name: "sections ordered by order field",
			yaml: `sections:
  test_list:
    title: "## Test List"
    order: 3
    schema: "checklist"
    required: true
  findings:
    title: "## Findings"
    order: 1
    schema: "research"
    required: true
  strategy:
    title: "## Strategy"
    order: 2
    schema: "strategy"
    required: true`,
			expectedOrder: []string{"findings", "strategy", "test_list"},
		},
		{
			name: "handle duplicate order values (sort by key as tiebreaker)",
			yaml: `sections:
  zebra:
    title: "## Zebra"
    order: 1
    schema: "freeform"
  alpha:
    title: "## Alpha"
    order: 1
    schema: "freeform"
  beta:
    title: "## Beta"
    order: 2
    schema: "freeform"`,
			expectedOrder: []string{"alpha", "zebra", "beta"},
		},
		{
			name: "handle missing order (defaults to 0)",
			yaml: `sections:
  no_order:
    title: "## No Order"
    schema: "freeform"
  first:
    title: "## First"
    order: 1
    schema: "freeform"
  negative:
    title: "## Negative"
    order: -1
    schema: "freeform"`,
			expectedOrder: []string{"negative", "no_order", "first"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse sections
			sections, err := ParseSectionDefinitions([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("ParseSectionDefinitions() error = %v", err)
			}
			
			// Get ordered sections
			ordered := GetOrderedSections(sections)
			
			// Check order
			if len(ordered) != len(tt.expectedOrder) {
				t.Errorf("GetOrderedSections() returned %d sections, want %d", len(ordered), len(tt.expectedOrder))
				return
			}
			
			for i, expectedKey := range tt.expectedOrder {
				if ordered[i].Key != expectedKey {
					t.Errorf("GetOrderedSections() position %d = %v, want %v", i, ordered[i].Key, expectedKey)
				}
			}
		})
	}
}

// Test 4: Handle missing section metadata (backwards compatibility)
func TestBackwardsCompatibilityForLegacyTodos(t *testing.T) {
	tests := []struct {
		name             string
		markdownContent  string
		expectedSections map[string]*SectionDefinition
	}{
		{
			name: "infer sections from standard todo format",
			markdownContent: `---
todo_id: "legacy-todo"
started: "2025-06-29 15:00:00"
status: "in_progress"
priority: "high"
type: "feature"
---

# Task: Legacy todo without section metadata

## Findings & Research

Some research content here.

## Test Strategy

Testing approach documented.

## Test List

- [ ] Test 1
- [x] Test 2

## Test Cases

` + "```go" + `
func TestExample(t *testing.T) {}
` + "```" + `

## Checklist

- [x] Item 1
- [ ] Item 2

## Working Scratchpad

Notes and scratch work.`,
			expectedSections: map[string]*SectionDefinition{
				"findings": {
					Title:    "## Findings & Research",
					Order:    1,
					Schema:   SchemaResearch,
					Required: false,
				},
				"test_strategy": {
					Title:    "## Test Strategy",
					Order:    2,
					Schema:   SchemaStrategy,
					Required: false,
				},
				"test_list": {
					Title:    "## Test List",
					Order:    3,
					Schema:   SchemaChecklist,
					Required: false,
				},
				"tests": {
					Title:    "## Test Cases",
					Order:    4,
					Schema:   SchemaTestCases,
					Required: false,
				},
				"checklist": {
					Title:    "## Checklist",
					Order:    5,
					Schema:   SchemaChecklist,
					Required: false,
				},
				"scratchpad": {
					Title:    "## Working Scratchpad",
					Order:    6,
					Schema:   SchemaFreeform,
					Required: false,
				},
			},
		},
		{
			name: "handle non-standard sections",
			markdownContent: `---
todo_id: "custom-todo"
---

# Task: Todo with custom sections

## Custom Analysis

Some custom content.

## Performance Metrics

Performance data here.

## Findings & Research

Standard section mixed with custom.`,
			expectedSections: map[string]*SectionDefinition{
				"custom_analysis": {
					Title:    "## Custom Analysis",
					Order:    1,
					Schema:   SchemaFreeform,
					Required: false,
					Custom:   true,
				},
				"performance_metrics": {
					Title:    "## Performance Metrics",
					Order:    2,
					Schema:   SchemaFreeform,
					Required: false,
					Custom:   true,
				},
				"findings": {
					Title:    "## Findings & Research",
					Order:    3,
					Schema:   SchemaResearch,
					Required: false,
				},
			},
		},
		{
			name: "empty todo with no sections",
			markdownContent: `---
todo_id: "empty-todo"
---

# Task: Empty todo`,
			expectedSections: map[string]*SectionDefinition{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Infer sections from markdown content
			sections := InferSectionsFromMarkdown(tt.markdownContent)
			
			// Check section count
			if len(sections) != len(tt.expectedSections) {
				t.Errorf("InferSectionsFromMarkdown() returned %d sections, want %d", len(sections), len(tt.expectedSections))
				return
			}
			
			// Check each section
			for key, expectedDef := range tt.expectedSections {
				gotDef, exists := sections[key]
				if !exists {
					t.Errorf("InferSectionsFromMarkdown() missing section %s", key)
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
			}
		})
	}
}

// Test 6: Calculate section metrics (word count, checkbox completion)
func TestCalculateSectionMetrics(t *testing.T) {
	tests := []struct {
		name     string
		schema   SectionSchema
		content  string
		expected map[string]interface{}
	}{
		{
			name:    "research section word count",
			schema:  SchemaResearch,
			content: "This is a research section with some findings and analysis.",
			expected: map[string]interface{}{
				"word_count": 10,
			},
		},
		{
			name:    "empty research section",
			schema:  SchemaResearch,
			content: "",
			expected: map[string]interface{}{
				"word_count": 0,
			},
		},
		{
			name:    "checklist completion metrics",
			schema:  SchemaChecklist,
			content: "- [x] Completed task\n- [ ] Pending task\n- [x] Another done\n- [ ] Not done",
			expected: map[string]interface{}{
				"completed": 2,
				"total":     4,
			},
		},
		{
			name:    "empty checklist",
			schema:  SchemaChecklist,
			content: "",
			expected: map[string]interface{}{
				"completed": 0,
				"total":     0,
			},
		},
		{
			name:    "test cases code block count",
			schema:  SchemaTestCases,
			content: "```go\nfunc Test1() {}\n```\n\nSome text\n\n```python\ndef test2():\n    pass\n```",
			expected: map[string]interface{}{
				"code_blocks": 2,
			},
		},
		{
			name:    "results log entry count",
			schema:  SchemaResults,
			content: "[2025-01-01 10:00:00] Test started\n[2025-01-01 10:01:00] Test passed\n\nSome notes\n\n[2025-01-01 10:02:00] All done",
			expected: map[string]interface{}{
				"entries": 3,
			},
		},
		{
			name:    "strategy section count",
			schema:  SchemaStrategy,
			content: "### Approach\nDetails\n\n### Implementation\nMore details\n\n### Testing Strategy\nTest plan",
			expected: map[string]interface{}{
				"sections": 3,
			},
		},
		{
			name:    "freeform content length",
			schema:  SchemaFreeform,
			content: "This is some freeform content.",
			expected: map[string]interface{}{
				"length": 30,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := GetValidator(tt.schema)
			if validator == nil {
				t.Fatalf("GetValidator(%v) returned nil", tt.schema)
			}
			
			metrics := validator.GetMetrics(tt.content)
			
			// Check each expected metric
			for key, expectedValue := range tt.expected {
				gotValue, exists := metrics[key]
				if !exists {
					t.Errorf("GetMetrics() missing metric %s", key)
					continue
				}
				
				if gotValue != expectedValue {
					t.Errorf("GetMetrics() metric %s = %v, want %v", key, gotValue, expectedValue)
				}
			}
			
			// Check for unexpected metrics
			for key := range metrics {
				if _, expected := tt.expected[key]; !expected {
					t.Errorf("GetMetrics() has unexpected metric %s", key)
				}
			}
		})
	}
}

// Test 7: Validate required sections are present
func TestValidateRequiredSections(t *testing.T) {
	tests := []struct {
		name            string
		sections        map[string]*SectionDefinition
		markdownContent string
		wantErr         bool
		errMsg          string
	}{
		{
			name: "all required sections present",
			sections: map[string]*SectionDefinition{
				"findings": {
					Title:    "## Findings & Research",
					Required: true,
					Schema:   SchemaResearch,
				},
				"test_list": {
					Title:    "## Test List",
					Required: true,
					Schema:   SchemaChecklist,
				},
				"optional": {
					Title:    "## Optional Section",
					Required: false,
					Schema:   SchemaFreeform,
				},
			},
			markdownContent: `# Task: Test Todo

## Findings & Research

Research content here.

## Test List

- [ ] Test 1
- [ ] Test 2

## Some Other Section

This section is not tracked.`,
			wantErr: false,
		},
		{
			name: "missing required section",
			sections: map[string]*SectionDefinition{
				"findings": {
					Title:    "## Findings & Research",
					Required: true,
					Schema:   SchemaResearch,
				},
				"test_list": {
					Title:    "## Test List",
					Required: true,
					Schema:   SchemaChecklist,
				},
			},
			markdownContent: `# Task: Test Todo

## Findings & Research

Research content here.

## Some Other Section

Missing the required Test List section.`,
			wantErr: true,
			errMsg:  "missing required section: Test List",
		},
		{
			name: "empty content with required sections",
			sections: map[string]*SectionDefinition{
				"findings": {
					Title:    "## Findings & Research",
					Required: true,
					Schema:   SchemaResearch,
				},
			},
			markdownContent: `# Task: Empty Todo`,
			wantErr:         true,
			errMsg:          "missing required section: Findings & Research",
		},
		{
			name: "no required sections",
			sections: map[string]*SectionDefinition{
				"optional1": {
					Title:    "## Optional 1",
					Required: false,
					Schema:   SchemaFreeform,
				},
				"optional2": {
					Title:    "## Optional 2",
					Required: false,
					Schema:   SchemaFreeform,
				},
			},
			markdownContent: `# Task: Test Todo

Some content without any sections.`,
			wantErr: false,
		},
		{
			name: "required section with different case",
			sections: map[string]*SectionDefinition{
				"findings": {
					Title:    "## Findings & Research",
					Required: true,
					Schema:   SchemaResearch,
				},
			},
			markdownContent: `# Task: Test Todo

## findings & research

Should not match due to case sensitivity.`,
			wantErr: true,
			errMsg:  "missing required section: Findings & Research",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredSections(tt.sections, tt.markdownContent)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequiredSections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("ValidateRequiredSections() error = %v, want error %v", err.Error(), tt.errMsg)
			}
		})
	}
}