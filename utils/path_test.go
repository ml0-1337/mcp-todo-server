package utils

import (
	"os"
	"testing"
)

// Test 1: GetEnv returns environment variable when set
func TestGetEnv_WithEnvironmentVariable(t *testing.T) {
	// Set up test environment variable
	testKey := "TEST_ENV_VAR"
	testValue := "test_value"
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)
	
	// Call GetEnv
	result := GetEnv(testKey, "fallback")
	
	// Assert result equals test value
	if result != testValue {
		t.Errorf("GetEnv(%s, fallback) = %s; want %s", testKey, result, testValue)
	}
}