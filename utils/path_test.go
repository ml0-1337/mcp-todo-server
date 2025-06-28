package utils

import (
	"io/ioutil"
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

// Test 2: GetEnv returns fallback when variable not set
func TestGetEnv_WithoutEnvironmentVariable(t *testing.T) {
	// Ensure test key doesn't exist
	testKey := "NON_EXISTENT_ENV_VAR"
	os.Unsetenv(testKey)
	
	// Call GetEnv with fallback
	fallback := "default_value"
	result := GetEnv(testKey, fallback)
	
	// Assert result equals fallback
	if result != fallback {
		t.Errorf("GetEnv(%s, %s) = %s; want %s", testKey, fallback, result, fallback)
	}
}

// Test 3: FileExists returns true for existing file
func TestFileExists_WithExistingFile(t *testing.T) {
	// Create a temporary file
	tempFile, err := ioutil.TempFile("", "test-file-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	// Test FileExists
	exists := FileExists(tempFile.Name())
	
	// Assert file exists
	if !exists {
		t.Errorf("FileExists(%s) = false; want true", tempFile.Name())
	}
}