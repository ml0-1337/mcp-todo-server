package logging

import (
	"fmt"
	"os"
	"time"
)

// Logf logs a formatted message with timestamp to stderr
func Logf(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02T15:04:05-07:00")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[%s] %s\n", timestamp, message)
}

// CategoryLogf logs a formatted message with timestamp and category to stderr
func CategoryLogf(category, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02T15:04:05-07:00")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[%s] [%s] %s\n", timestamp, category, message)
}

// Debugf logs a debug message with timestamp to stderr
func Debugf(format string, args ...interface{}) {
	CategoryLogf("DEBUG", format, args...)
}

// Infof logs an info message with timestamp to stderr
func Infof(format string, args ...interface{}) {
	CategoryLogf("INFO", format, args...)
}

// Warnf logs a warning message with timestamp to stderr
func Warnf(format string, args ...interface{}) {
	CategoryLogf("WARNING", format, args...)
}

// Errorf logs an error message with timestamp to stderr
func Errorf(format string, args ...interface{}) {
	CategoryLogf("ERROR", format, args...)
}

// Timingf logs a timing message with timestamp to stderr
func Timingf(format string, args ...interface{}) {
	CategoryLogf("TIMING", format, args...)
}

// Connectionf logs a connection message with timestamp to stderr
func Connectionf(format string, args ...interface{}) {
	CategoryLogf("Connection", format, args...)
}

// Headerf logs a header message with timestamp to stderr
func Headerf(format string, args ...interface{}) {
	CategoryLogf("Header", format, args...)
}

// StableHTTPf logs a StableHTTP message with timestamp to stderr
func StableHTTPf(format string, args ...interface{}) {
	CategoryLogf("StableHTTP", format, args...)
}

// Performancef logs a performance message with timestamp to stderr
func Performancef(format string, args ...interface{}) {
	CategoryLogf("PERFORMANCE", format, args...)
}

// Progressf logs a progress message with timestamp to stderr
func Progressf(format string, args ...interface{}) {
	CategoryLogf("PROGRESS", format, args...)
}