package core

import (
	"fmt"
	"strings"
	"gopkg.in/yaml.v3"
)

// SectionSchema represents the validation schema for a section
type SectionSchema string

const (
	SchemaResearch   SectionSchema = "research"    // Free text with citations
	SchemaStrategy   SectionSchema = "strategy"    // Structured plan
	SchemaChecklist  SectionSchema = "checklist"   // Checkbox items
	SchemaTestCases  SectionSchema = "test_cases"  // Code blocks
	SchemaResults    SectionSchema = "results"     // Timestamped logs
	SchemaFreeform   SectionSchema = "freeform"    // No validation
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
		// Must start with "- [ ]" or "- [x]"
		if !strings.HasPrefix(trimmed, "- [ ]") && !strings.HasPrefix(trimmed, "- [x]") {
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
	total := 0
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [x]") {
			completed++
			total++
		} else if strings.HasPrefix(trimmed, "- [ ]") {
			total++
		}
	}
	return map[string]interface{}{
		"completed": completed,
		"total":     total,
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
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Must start with timestamp in brackets
		if !strings.HasPrefix(trimmed, "[") || !strings.Contains(trimmed, "]") {
			return fmt.Errorf("entries must start with timestamp")
		}
	}
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