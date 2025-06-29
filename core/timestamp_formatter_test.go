package core

import (
	"strings"
	"testing"
	"time"
)

// Test 1: formatWithTimestamp() adds timestamp to single line
func TestFormatWithTimestamp_SingleLine(t *testing.T) {
	// Test input
	content := "Test entry without timestamp"
	
	// Format with timestamp
	result := formatWithTimestamp(content)
	
	// Check that result starts with timestamp
	if !strings.HasPrefix(result, "[") {
		t.Error("Result should start with timestamp bracket")
	}
	
	// Check that timestamp is in correct format
	if !strings.Contains(result, "] ") {
		t.Error("Result should contain closing bracket followed by space")
	}
	
	// Check that original content is preserved
	if !strings.Contains(result, "Test entry without timestamp") {
		t.Error("Original content should be preserved")
	}
	
	// Verify timestamp format is valid
	parts := strings.SplitN(result, "] ", 2)
	if len(parts) != 2 {
		t.Fatal("Could not split timestamp from content")
	}
	
	timestampPart := strings.TrimPrefix(parts[0], "[")
	_, err := time.Parse("2006-01-02 15:04:05", timestampPart)
	if err != nil {
		t.Errorf("Timestamp format invalid: %v", err)
	}
}