package core

import (
	"testing"
)

// Test 8: ResultsValidator accepts entries without timestamps
func TestResultsValidator_AcceptsEntriesWithoutTimestamps(t *testing.T) {
	validator := &ResultsValidator{}
	
	// Test content without timestamps
	content := `Test entry without timestamp
Another line without timestamp
[2025-01-01 10:00:00] Mixed with timestamped entry
Final line without timestamp`
	
	// Validate should pass (after we update the validator)
	err := validator.Validate(content)
	if err != nil {
		t.Errorf("ResultsValidator should accept entries without timestamps, got error: %v", err)
	}
}

// Test 9: ResultsValidator still accepts entries with timestamps
func TestResultsValidator_StillAcceptsTimestampedEntries(t *testing.T) {
	validator := &ResultsValidator{}
	
	// Test content with all timestamps
	content := `[2025-01-01 10:00:00] First entry
[2025-01-01 10:01:00] Second entry
[2025-01-01 10:02:00] Third entry`
	
	// Validate should pass
	err := validator.Validate(content)
	if err != nil {
		t.Errorf("ResultsValidator should still accept timestamped entries, got error: %v", err)
	}
}