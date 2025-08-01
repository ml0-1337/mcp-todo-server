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

// Test 2: formatWithTimestamp() handles multi-line content
func TestFormatWithTimestamp_MultiLine(t *testing.T) {
	// Test input with multiple lines
	content := "Line 1: Test started\nLine 2: Test in progress\nLine 3: Test completed"

	// Format with timestamp
	result := formatWithTimestamp(content)

	// The implementation adds a single timestamp to the entire content
	// Check that result starts with timestamp
	if !strings.HasPrefix(result, "[") {
		t.Error("Result should start with timestamp bracket")
	}

	// Check that timestamp is properly closed
	if !strings.Contains(result, "] ") {
		t.Error("Result should contain closing bracket followed by space")
	}

	// Extract the timestamp part
	bracketEnd := strings.Index(result, "] ")
	if bracketEnd == -1 {
		t.Fatal("Could not find end of timestamp")
	}

	// Verify the timestamp format (YYYY-MM-DD HH:MM:SS)
	timestamp := result[1:bracketEnd]
	if len(timestamp) != 19 {
		t.Errorf("Timestamp has unexpected length: %d", len(timestamp))
	}

	// Verify all original content is preserved after the timestamp
	contentAfterTimestamp := result[bracketEnd+2:]
	if contentAfterTimestamp != content {
		t.Errorf("Original content not preserved. Expected: %s, Got: %s", content, contentAfterTimestamp)
	}
}

// Test 3: formatWithTimestamp() adds timestamp to entire content
func TestFormatWithTimestamp_WithExistingTimestamps(t *testing.T) {
	// Test input with existing timestamps in the content
	content := "[2025-01-01 10:00:00] Already timestamped entry\nNew entry without timestamp\n[2025-01-01 11:00:00] Another timestamped entry"

	// Format with timestamp
	result := formatWithTimestamp(content)

	// The implementation adds a single timestamp to the entire content block
	// It doesn't parse or modify existing timestamps within the content

	// Check that result starts with a new timestamp
	if !strings.HasPrefix(result, "[") {
		t.Error("Result should start with timestamp bracket")
	}

	// Extract the new timestamp
	bracketEnd := strings.Index(result, "] ")
	if bracketEnd == -1 {
		t.Fatal("Could not find end of timestamp")
	}

	// The new timestamp should be current (not 2025-01-01)
	timestamp := result[1:bracketEnd]
	if strings.HasPrefix(timestamp, "2025-01-01") {
		t.Error("Should have added a new current timestamp, not an old one")
	}

	// Verify all original content (including the existing timestamps) is preserved
	contentAfterTimestamp := result[bracketEnd+2:]
	if contentAfterTimestamp != content {
		t.Errorf("Original content not preserved. Expected: %s, Got: %s", content, contentAfterTimestamp)
	}
}
