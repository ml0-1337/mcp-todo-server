package core

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"sort"
	"strings"
)

// SectionSchema represents the validation schema for a section
type SectionSchema string

const (
	SchemaResearch  SectionSchema = "research"   // Free text with citations
	SchemaStrategy  SectionSchema = "strategy"   // Structured plan
	SchemaChecklist SectionSchema = "checklist"  // Checkbox items
	SchemaTestCases SectionSchema = "test_cases" // Code blocks
	SchemaResults   SectionSchema = "results"    // Timestamped logs
	SchemaFreeform  SectionSchema = "freeform"   // No validation
)

// SectionDefinition represents metadata for a todo section
type SectionDefinition struct {
	Title    string                 `yaml:"title"`
	Order    int                    `yaml:"order"`
	Schema   SectionSchema          `yaml:"schema"`
	Required bool                   `yaml:"required"`
	Custom   bool                   `yaml:"custom,omitempty"`
	Metadata map[string]interface{} `yaml:"metadata,omitempty"`
}

// SectionValidator validates section content based on schema
type SectionValidator interface {
	Validate(content string) error
	GetMetrics(content string) map[string]interface{}
}

// TodoWithSections represents the YAML structure with sections
type todoFrontmatterWithSections struct {
	TodoID   string                        `yaml:"todo_id"`
	Started  string                        `yaml:"started"`
	Status   string                        `yaml:"status"`
	Sections map[string]*SectionDefinition `yaml:"sections"`
}

// ParseSectionDefinitions parses section definitions from YAML frontmatter
func ParseSectionDefinitions(yamlData []byte) (map[string]*SectionDefinition, error) {
	var frontmatter todoFrontmatterWithSections
	err := yaml.Unmarshal(yamlData, &frontmatter)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// If no sections defined, return nil (backwards compatibility)
	if frontmatter.Sections == nil {
		return nil, nil
	}

	// Validate schema types
	for key, section := range frontmatter.Sections {
		if !isValidSchema(section.Schema) {
			return nil, fmt.Errorf("invalid schema type '%s' for section '%s'", section.Schema, key)
		}
	}

	return frontmatter.Sections, nil
}

// isValidSchema checks if a schema type is valid
func isValidSchema(schema SectionSchema) bool {
	switch schema {
	case SchemaResearch, SchemaStrategy, SchemaChecklist, SchemaTestCases, SchemaResults, SchemaFreeform:
		return true
	default:
		return false
	}
}

// GetValidator returns the appropriate validator for a schema type
func GetValidator(schema SectionSchema) SectionValidator {
	switch schema {
	case SchemaResearch:
		return &ResearchValidator{}
	case SchemaStrategy:
		return &StrategyValidator{}
	case SchemaChecklist:
		return &ChecklistValidator{}
	case SchemaTestCases:
		return &TestCasesValidator{}
	case SchemaResults:
		return &ResultsValidator{}
	case SchemaFreeform:
		return &FreeformValidator{}
	default:
		return nil
	}
}

// ResearchValidator validates research content (accepts any text)
type ResearchValidator struct{}

func (v *ResearchValidator) Validate(content string) error {
	// Research sections accept any content
	return nil
}

func (v *ResearchValidator) GetMetrics(content string) map[string]interface{} {
	return map[string]interface{}{
		"word_count": len(strings.Fields(content)),
	}
}

// ChecklistValidator validates checklist content
type ChecklistValidator struct{}

func (v *ChecklistValidator) Validate(content string) error {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Must start with valid checkbox syntax
		validPrefixes := []string{
			"- [ ]",          // pending
			"- [x]", "- [X]", // completed
			"- [>]", "- [-]", "- [~]", // in_progress
		}

		hasValidPrefix := false
		for _, prefix := range validPrefixes {
			if strings.HasPrefix(trimmed, prefix) {
				hasValidPrefix = true
				break
			}
		}

		if !hasValidPrefix {
			if strings.HasPrefix(trimmed, "- []") || strings.HasPrefix(trimmed, "- [") {
				return fmt.Errorf("invalid checkbox syntax")
			}
			return fmt.Errorf("non-checklist content found")
		}
	}
	return nil
}

func (v *ChecklistValidator) GetMetrics(content string) map[string]interface{} {
	completed := 0
	inProgress := 0
	pending := 0

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
			completed++
		} else if strings.HasPrefix(trimmed, "- [>]") || strings.HasPrefix(trimmed, "- [-]") || strings.HasPrefix(trimmed, "- [~]") {
			inProgress++
		} else if strings.HasPrefix(trimmed, "- [ ]") {
			pending++
		}
	}

	total := completed + inProgress + pending
	return map[string]interface{}{
		"completed":   completed,
		"in_progress": inProgress,
		"pending":     pending,
		"total":       total,
	}
}

// TestCasesValidator validates test case content
type TestCasesValidator struct{}

func (v *TestCasesValidator) Validate(content string) error {
	// Check for code blocks
	if !strings.Contains(content, "```") {
		return fmt.Errorf("no code blocks found")
	}
	return nil
}

func (v *TestCasesValidator) GetMetrics(content string) map[string]interface{} {
	codeBlocks := strings.Count(content, "```") / 2
	return map[string]interface{}{
		"code_blocks": codeBlocks,
	}
}

// ResultsValidator validates timestamped log entries
type ResultsValidator struct{}

func (v *ResultsValidator) Validate(content string) error {
	// Now accepts both timestamped and non-timestamped entries
	// The automatic timestamping happens at the UpdateTodo level
	// So the validator should be more permissive
	return nil
}

func (v *ResultsValidator) GetMetrics(content string) map[string]interface{} {
	entries := 0
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && strings.HasPrefix(trimmed, "[") {
			entries++
		}
	}
	return map[string]interface{}{
		"entries": entries,
	}
}

// StrategyValidator validates strategy content
type StrategyValidator struct{}

func (v *StrategyValidator) Validate(content string) error {
	// Strategy sections accept structured content
	// For now, just ensure it's not empty
	return nil
}

func (v *StrategyValidator) GetMetrics(content string) map[string]interface{} {
	sections := strings.Count(content, "###")
	return map[string]interface{}{
		"sections": sections,
	}
}

// FreeformValidator validates freeform content (accepts anything)
type FreeformValidator struct{}

func (v *FreeformValidator) Validate(content string) error {
	// Freeform accepts any content
	return nil
}

func (v *FreeformValidator) GetMetrics(content string) map[string]interface{} {
	return map[string]interface{}{
		"length": len(content),
	}
}

// OrderedSection represents a section with its key for ordering
type OrderedSection struct {
	Key        string
	Definition *SectionDefinition
}

// GetOrderedSections returns sections sorted by their order field
func GetOrderedSections(sections map[string]*SectionDefinition) []OrderedSection {
	if sections == nil {
		return nil
	}

	// Convert map to slice
	ordered := make([]OrderedSection, 0, len(sections))
	for key, def := range sections {
		ordered = append(ordered, OrderedSection{
			Key:        key,
			Definition: def,
		})
	}

	// Sort by order field, then by key as tiebreaker
	sort.Slice(ordered, func(i, j int) bool {
		// Compare by order first
		if ordered[i].Definition.Order != ordered[j].Definition.Order {
			return ordered[i].Definition.Order < ordered[j].Definition.Order
		}
		// Use key as tiebreaker
		return ordered[i].Key < ordered[j].Key
	})

	return ordered
}

// Standard section mappings for backwards compatibility
var standardSectionMappings = map[string]struct {
	Key    string
	Schema SectionSchema
}{
	"## Findings & Research":      {Key: "findings", Schema: SchemaResearch},
	"## Web Searches":             {Key: "web_searches", Schema: SchemaResearch},
	"## Test Strategy":            {Key: "test_strategy", Schema: SchemaStrategy},
	"## Test List":                {Key: "test_list", Schema: SchemaChecklist},
	"## Test Cases":               {Key: "tests", Schema: SchemaTestCases},
	"## Maintainability Analysis": {Key: "maintainability", Schema: SchemaFreeform},
	"## Test Results Log":         {Key: "test_results", Schema: SchemaResults},
	"## Checklist":                {Key: "checklist", Schema: SchemaChecklist},
	"## Working Scratchpad":       {Key: "scratchpad", Schema: SchemaFreeform},
}

// InferSectionsFromMarkdown analyzes markdown content to infer section definitions
func InferSectionsFromMarkdown(content string) map[string]*SectionDefinition {
	sections := make(map[string]*SectionDefinition)

	// Split content into lines
	lines := strings.Split(content, "\n")
	order := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for section headings (## Something)
		if strings.HasPrefix(trimmed, "## ") {
			order++

			// Check if it's a standard section
			if mapping, exists := standardSectionMappings[trimmed]; exists {
				sections[mapping.Key] = &SectionDefinition{
					Title:    trimmed,
					Order:    order,
					Schema:   mapping.Schema,
					Required: false, // Legacy sections not required
				}
			} else {
				// Custom section - generate key from title
				key := generateSectionKey(trimmed)
				sections[key] = &SectionDefinition{
					Title:    trimmed,
					Order:    order,
					Schema:   SchemaFreeform,
					Required: false,
					Custom:   true,
				}
			}
		}
	}

	return sections
}

// generateSectionKey converts a section title to a key
func generateSectionKey(title string) string {
	// Remove "## " prefix
	clean := strings.TrimPrefix(title, "## ")

	// Convert to lowercase and replace spaces with underscores
	key := strings.ToLower(clean)
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "&", "and")

	// Remove special characters
	replacer := strings.NewReplacer(
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
		":", "",
		";", "",
		",", "",
		".", "",
		"!", "",
		"?", "",
		"'", "",
		"\"", "",
		"/", "_",
		"\\", "_",
		"-", "_",
	)
	key = replacer.Replace(key)

	// Remove duplicate underscores
	for strings.Contains(key, "__") {
		key = strings.ReplaceAll(key, "__", "_")
	}

	// Trim underscores
	key = strings.Trim(key, "_")

	return key
}

// ValidateRequiredSections checks if all required sections are present in the markdown content
func ValidateRequiredSections(sections map[string]*SectionDefinition, markdownContent string) error {
	// First, check if there are any required sections
	hasRequired := false
	for _, def := range sections {
		if def.Required {
			hasRequired = true
			break
		}
	}

	// If no required sections, validation passes
	if !hasRequired {
		return nil
	}

	// Check each required section
	for _, def := range sections {
		if !def.Required {
			continue
		}

		// Look for the section title in the markdown
		if !strings.Contains(markdownContent, def.Title) {
			// Extract just the section name from the title (remove "## " prefix)
			sectionName := strings.TrimPrefix(def.Title, "## ")
			return fmt.Errorf("missing required section: %s", sectionName)
		}
	}

	return nil
}
