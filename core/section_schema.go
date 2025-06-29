package core

import (
	"fmt"
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