// +build ignore

// This file contains the patterns for fixing handler tests
// Run with: go run fix_handler_tests.go

package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func main() {
	// Read the test file
	content, err := ioutil.ReadFile("e2e_test.go")
	if err != nil {
		panic(err)
	}

	text := string(content)

	// Pattern 1: Fix simple handler calls with MockCallToolRequest
	// Match: handlers.HandleXXX(context.Background(), &MockCallToolRequest{
	re1 := regexp.MustCompile(`(\s+)result := handlers\.Handle(\w+)\(context\.Background\(\), &MockCallToolRequest\{`)
	text = re1.ReplaceAllString(text, `${1}mockReq := &MockCallToolRequest{`)

	// Pattern 2: Add the actual handler call after MockCallToolRequest definition
	// This is complex and needs manual handling for now

	// Pattern 3: Fix result = handlers... (assignment)
	re2 := regexp.MustCompile(`(\s+)result = handlers\.Handle(\w+)\(context\.Background\(\), &MockCallToolRequest\{`)
	text = re2.ReplaceAllString(text, `${1}mockReq := &MockCallToolRequest{`)

	// Pattern 4: Fix IsError checks
	text = strings.ReplaceAll(text, "if result.IsError {", "if err != nil {")
	text = strings.ReplaceAll(text, "result.Content", "err")

	// Pattern 5: Fix result content access
	text = strings.ReplaceAll(text, "results, ok := result.Content.([]interface{})", "// Check result content")

	fmt.Println("Fixed patterns:")
	fmt.Println("- Handler calls to use mockReq")
	fmt.Println("- IsError checks to use err != nil")
	fmt.Println("- result.Content to err")
	fmt.Println("\nManual fixes still needed:")
	fmt.Println("- Add handler calls after mockReq definitions")
	fmt.Println("- Fix result content parsing")
	fmt.Println("- Update variable names where needed")

	// Write back
	err = ioutil.WriteFile("e2e_test_fixed.go", []byte(text), 0644)
	if err != nil {
		panic(err)
	}
}