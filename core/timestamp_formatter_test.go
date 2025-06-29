package core

import (
	"fmt"
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

// Test 2: formatWithTimestamp() handles multi-line content
func TestFormatWithTimestamp_MultiLine(t *testing.T) {
	// Test input with multiple lines
	content := "Line 1: Test started\nLine 2: Test in progress\nLine 3: Test completed"
	
	// Format with timestamp
	result := formatWithTimestamp(content)
	
	// Split result into lines
	lines := strings.Split(result, "\n")
	
	// Each line should have a timestamp
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}
	
	for i, line := range lines {
		// Check that each line starts with timestamp
		if !strings.HasPrefix(line, "[") {
			t.Errorf("Line %d should start with timestamp bracket", i+1)
		}
		
		// Check that timestamp is followed by content
		if !strings.Contains(line, "] ") {
			t.Errorf("Line %d should contain closing bracket followed by space", i+1)
		}
		
		// Verify original content is preserved
		expectedContent := fmt.Sprintf("Line %d:", i+1)
		if !strings.Contains(line, expectedContent) {
			t.Errorf("Line %d content not preserved", i+1)
		}
	}
}

// Test 3: formatWithTimestamp() preserves existing timestamps
func TestFormatWithTimestamp_PreservesExisting(t *testing.T) {
	// Test input with existing timestamps
	content := "[2025-01-01 10:00:00] Already timestamped entry\nNew entry without timestamp\n[2025-01-01 11:00:00] Another timestamped entry"
	
	// Format with timestamp
	result := formatWithTimestamp(content)
	
	// Split result into lines
	lines := strings.Split(result, "\n")
	
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}
	
	// First line should keep original timestamp
	if !strings.HasPrefix(lines[0], "[2025-01-01 10:00:00]") {
		t.Error("First line should preserve original timestamp")
	}
	
	// Second line should get new timestamp
	if !strings.HasPrefix(lines[1], "[") || strings.HasPrefix(lines[1], "[2025-01-01") {
		t.Error("Second line should have a new timestamp")
	}
	
	// Third line should keep original timestamp
	if !strings.HasPrefix(lines[2], "[2025-01-01 11:00:00]") {
		t.Error("Third line should preserve original timestamp")
	}
}