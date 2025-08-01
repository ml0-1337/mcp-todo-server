package search

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/v2/mapping"
	"strings"
)

// buildIndexMapping creates the index mapping for todos
func buildIndexMapping() mapping.IndexMapping {
	// Create a standard document mapping
	todoMapping := bleve.NewDocumentMapping()

	// Task field - boosted for relevance
	taskFieldMapping := bleve.NewTextFieldMapping()
	taskFieldMapping.Analyzer = standard.Name
	taskFieldMapping.Store = true
	taskFieldMapping.IncludeInAll = true
	todoMapping.AddFieldMappingsAt("task", taskFieldMapping)

	// Content fields
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = standard.Name
	textFieldMapping.Store = true
	textFieldMapping.IncludeInAll = true

	todoMapping.AddFieldMappingsAt("content", textFieldMapping)
	todoMapping.AddFieldMappingsAt("findings", textFieldMapping)
	todoMapping.AddFieldMappingsAt("tests", textFieldMapping)

	// Metadata fields
	keywordFieldMapping := bleve.NewKeywordFieldMapping()
	keywordFieldMapping.Store = true

	todoMapping.AddFieldMappingsAt("id", keywordFieldMapping)
	todoMapping.AddFieldMappingsAt("status", keywordFieldMapping)
	todoMapping.AddFieldMappingsAt("priority", keywordFieldMapping)
	todoMapping.AddFieldMappingsAt("type", keywordFieldMapping)

	// Date fields
	dateFieldMapping := bleve.NewDateTimeFieldMapping()
	dateFieldMapping.Store = true
	dateFieldMapping.Index = true
	// Specify that we're using RFC3339 format for dates
	dateFieldMapping.DateFormat = "dateTimeOptional"

	todoMapping.AddFieldMappingsAt("started", dateFieldMapping)
	todoMapping.AddFieldMappingsAt("completed", dateFieldMapping)

	// Create index mapping
	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("todo", todoMapping)
	indexMapping.DefaultMapping = todoMapping

	return indexMapping
}

// extractSection extracts content under a specific heading
func extractSection(content, heading string) string {
	lines := strings.Split(content, "\n")
	inSection := false
	var sectionLines []string

	for _, line := range lines {
		if line == heading {
			inSection = true
			continue
		}

		if inSection && strings.HasPrefix(line, "## ") {
			break
		}

		if inSection {
			sectionLines = append(sectionLines, line)
		}
	}

	return strings.Join(sectionLines, "\n")
}
