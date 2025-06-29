package core

import (
	"testing"
)

// Test 1: web_searches section is included in standardSectionMappings
func TestWebSearchesSectionInStandardMappings(t *testing.T) {
	// Check if "## Web Searches" is in the standard mappings
	mapping, exists := standardSectionMappings["## Web Searches"]
	
	if !exists {
		t.Fatal("web_searches section not found in standardSectionMappings")
	}
	
	// Verify the key is correct
	if mapping.Key != "web_searches" {
		t.Errorf("Expected key 'web_searches', got '%s'", mapping.Key)
	}
	
	// Verify it uses research schema
	if mapping.Schema != SchemaResearch {
		t.Errorf("Expected schema SchemaResearch, got '%s'", mapping.Schema)
	}
}