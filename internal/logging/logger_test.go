package logging

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// captureStderr captures stderr output during function execution
func captureStderr(f func()) string {
	// Save current stderr
	oldStderr := os.Stderr

	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Create a channel to signal when capture is complete
	done := make(chan string)

	// Start goroutine to read from pipe
	go func() {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r)
		done <- buf.String()
	}()

	// Execute the function
	f()

	// Close writer and restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Wait for capture to complete
	return <-done
}

func TestLogf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "simple message",
			format: "Test message",
			args:   []interface{}{},
			want:   "Test message",
		},
		{
			name:   "formatted message with string",
			format: "Hello %s",
			args:   []interface{}{"World"},
			want:   "Hello World",
		},
		{
			name:   "formatted message with number",
			format: "The answer is %d",
			args:   []interface{}{42},
			want:   "The answer is 42",
		},
		{
			name:   "multiple format specifiers",
			format: "%s: %d errors found",
			args:   []interface{}{"Scanner", 3},
			want:   "Scanner: 3 errors found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStderr(func() {
				Logf(tt.format, tt.args...)
			})

			// Check that output contains the expected message
			if !strings.Contains(output, tt.want) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.want, output)
			}

			// Check that output has timestamp format
			if !strings.Contains(output, "[20") { // Basic check for timestamp
				t.Error("Expected output to contain timestamp")
			}
		})
	}
}

func TestCategoryLogf(t *testing.T) {
	tests := []struct {
		name     string
		category string
		format   string
		args     []interface{}
		want     []string
	}{
		{
			name:     "simple category message",
			category: "TEST",
			format:   "Test message",
			args:     []interface{}{},
			want:     []string{"[TEST]", "Test message"},
		},
		{
			name:     "formatted category message",
			category: "DATABASE",
			format:   "Connected to %s:%d",
			args:     []interface{}{"localhost", 5432},
			want:     []string{"[DATABASE]", "Connected to localhost:5432"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStderr(func() {
				CategoryLogf(tt.category, tt.format, tt.args...)
			})

			// Check that output contains all expected parts
			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain '%s', got '%s'", want, output)
				}
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		category string
		message  string
	}{
		{
			name:     "debug log",
			logFunc:  Debugf,
			category: "[DEBUG]",
			message:  "Debug message",
		},
		{
			name:     "info log",
			logFunc:  Infof,
			category: "[INFO]",
			message:  "Info message",
		},
		{
			name:     "warning log",
			logFunc:  Warnf,
			category: "[WARNING]",
			message:  "Warning message",
		},
		{
			name:     "error log",
			logFunc:  Errorf,
			category: "[ERROR]",
			message:  "Error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStderr(func() {
				tt.logFunc("%s", tt.message)
			})

			if !strings.Contains(output, tt.category) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.category, output)
			}

			if !strings.Contains(output, tt.message) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.message, output)
			}
		})
	}
}

func TestSpecializedLoggers(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		category string
		message  string
		args     []interface{}
	}{
		{
			name:     "timing log",
			logFunc:  Timingf,
			category: "[TIMING]",
			message:  "Operation took %dms",
			args:     []interface{}{125},
		},
		{
			name:     "connection log",
			logFunc:  Connectionf,
			category: "[Connection]",
			message:  "Client %s connected",
			args:     []interface{}{"192.168.1.100"},
		},
		{
			name:     "header log",
			logFunc:  Headerf,
			category: "[Header]",
			message:  "Content-Type: %s",
			args:     []interface{}{"application/json"},
		},
		{
			name:     "stable http log",
			logFunc:  StableHTTPf,
			category: "[StableHTTP]",
			message:  "Request %s %s",
			args:     []interface{}{"GET", "/api/todos"},
		},
		{
			name:     "performance log",
			logFunc:  Performancef,
			category: "[PERFORMANCE]",
			message:  "CPU usage: %.2f%%",
			args:     []interface{}{45.67},
		},
		{
			name:     "progress log",
			logFunc:  Progressf,
			category: "[PROGRESS]",
			message:  "Processed %d/%d items",
			args:     []interface{}{50, 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStderr(func() {
				tt.logFunc(tt.message, tt.args...)
			})

			if !strings.Contains(output, tt.category) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.category, output)
			}

			expectedMessage := fmt.Sprintf(tt.message, tt.args...)
			if !strings.Contains(output, expectedMessage) {
				t.Errorf("Expected output to contain '%s', got '%s'", expectedMessage, output)
			}
		})
	}
}

func TestTimestampFormat(t *testing.T) {
	output := captureStderr(func() {
		Logf("Timestamp test")
	})

	// Extract timestamp from output
	// Format: [2006-01-02T15:04:05-07:00] ...
	start := strings.Index(output, "[")
	end := strings.Index(output, "]")

	if start == -1 || end == -1 || start >= end {
		t.Fatalf("Could not find timestamp in output: %s", output)
	}

	timestamp := output[start+1 : end]

	// Try to parse the timestamp
	_, err := time.Parse("2006-01-02T15:04:05-07:00", timestamp)
	if err != nil {
		t.Errorf("Timestamp '%s' does not match expected format: %v", timestamp, err)
	}
}

func TestConcurrentLogging(t *testing.T) {
	// Test that concurrent logging doesn't cause issues
	done := make(chan bool, 10)

	output := captureStderr(func() {
		for i := 0; i < 10; i++ {
			go func(n int) {
				Infof("Concurrent log %d", n)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	// Check that we have 10 log entries
	logCount := strings.Count(output, "[INFO]")
	if logCount != 10 {
		t.Errorf("Expected 10 log entries, got %d", logCount)
	}
}
