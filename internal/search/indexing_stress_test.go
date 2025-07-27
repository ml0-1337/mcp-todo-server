package search

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestInitialIndexingWithRealisticFiles tests indexing performance with larger, more realistic todo files
func TestInitialIndexingWithRealisticFiles(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
	
	err := os.MkdirAll(todosDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create todos directory: %v", err)
	}
	
	// Create 65 realistic todo files with substantial content
	t.Logf("Creating 65 realistic todo files...")
	createStart := time.Now()
	
	totalSize := int64(0)
	for i := 0; i < 65; i++ {
		content := generateRealisticTodoContent(i)
		todoPath := filepath.Join(todosDir, fmt.Sprintf("todo-%05d.md", i))
		err := ioutil.WriteFile(todoPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write todo file: %v", err)
		}
		totalSize += int64(len(content))
	}
	
	createTime := time.Since(createStart)
	avgSize := totalSize / 65
	t.Logf("Created 65 todo files in %v (avg size: %d bytes, total: %d MB)", 
		createTime, avgSize, totalSize/1024/1024)
	
	// Time the initial indexing
	indexStart := time.Now()
	
	engine, err := NewEngine(indexPath, todosDir)
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}
	defer engine.Close()
	
	indexTime := time.Since(indexStart)
	t.Logf("Initial indexing of 65 todos took %v (%.2f todos/sec)", 
		indexTime, float64(65)/indexTime.Seconds())
	
	// Verify all todos were indexed
	count, err := engine.GetIndexedCount()
	if err != nil {
		t.Fatalf("Failed to get indexed count: %v", err)
	}
	
	if count != 65 {
		t.Errorf("Expected 65 indexed documents, got %d", count)
	}
	
	// Log performance metrics
	if indexTime > 5*time.Second {
		t.Logf("WARNING: Indexing took longer than 5 seconds: %v", indexTime)
	}
	if indexTime > 30*time.Second {
		t.Errorf("FAILED: Indexing exceeded 30 second timeout: %v", indexTime)
	}
}

// generateRealisticTodoContent creates a large, realistic todo file similar to actual usage
func generateRealisticTodoContent(index int) string {
	// Generate substantial content to simulate real todos (5-10KB each)
	findings := generateLargeSection("Findings", index, 3)
	webSearches := generateLargeSection("Web Searches", index, 2)
	testCases := generateLargeSection("Test Cases", index, 2)
	
	// Build the complete todo content
	return fmt.Sprintf(`---
todo_id: realistic-todo-%05d
started: "%s"
completed:
status: in_progress
priority: %s
type: %s
parent_id:
current_test: "Test %d: %s validation"
---

# Task: %s - Implement %s with %s support

## Findings & Research
%s

## Web Searches
%s

## Test Strategy

- **Test Framework**: Jest with React Testing Library
- **Test Types**: Unit tests, Integration tests, E2E tests
- **Coverage Target**: 85%% minimum, 95%% for critical paths
- **Edge Cases**: 
  - Null/undefined inputs
  - Empty arrays and objects
  - Boundary conditions
  - Concurrent operations
  - Network failures
  - Permission errors

## Test List

- [x] Test 1: Basic %s functionality works correctly
- [x] Test 2: %s handles edge cases properly
- [>] Test 3: Performance meets requirements under load
- [ ] Test 4: Security vulnerabilities are addressed
- [ ] Test 5: Integration with %s works seamlessly
- [ ] Test 6: Error messages are user-friendly
- [ ] Test 7: Accessibility standards are met
- [ ] Test 8: Mobile responsiveness is verified

**Current Test**: Working on Test 3: Performance optimization

## Test Cases
%s

## Test Results Log

# Initial test run (Red phase)
[2025-07-18 14:30:00] npm test -- --coverage

FAIL src/components/Component.test.js
  Component
    × should render correctly when props are valid

    expect(received).toEqual(expected)

    Expected: {"status": "success", "data": [...]}
    Received: undefined

      42 |     const result = await component.process(input);
    > 43 |     expect(result).toEqual(expected);
         |                    ^

# After implementation (Green phase)
[2025-07-18 14:45:00] npm test -- --coverage

PASS src/components/Component.test.js
  Component
    ✓ should render correctly when props are valid (45ms)
    ✓ should handle errors gracefully (12ms)

Test Suites: 1 passed, 1 total
Tests:       2 passed, 2 total
Coverage:    87.5%% (14/16 lines)

## Checklist

- [x] Research existing patterns and best practices
- [x] Design component architecture
- [x] Write comprehensive tests
- [x] Implement core functionality
- [>] Optimize performance
- [ ] Add error handling
- [ ] Write documentation
- [ ] Code review
- [ ] Deploy to staging
- [ ] Performance testing
- [ ] Security audit
- [ ] Deploy to production

## Working Scratchpad

### Current Focus
Working on performance optimization. The current implementation processes %d items/second but we need to reach %d items/second for production requirements.

### Ideas to explore:
1. Implement caching layer with Redis
2. Use worker threads for parallel processing
3. Optimize database queries with proper indexing
4. Consider using streaming for large datasets

### Blockers:
- Waiting for DBA team to review index proposals
- Need approval for Redis infrastructure

### Next Steps:
1. Profile the current implementation to identify bottlenecks
2. Implement caching for frequently accessed data
3. Run load tests with production-like data
`,
		index,
		time.Now().Add(time.Duration(-index)*time.Hour*24).Format(time.RFC3339),
		[]string{"high", "medium", "low"}[index%3],
		[]string{"feature", "bug", "refactor", "research", "test"}[index%5],
		index+1,
		generateFeatureName(index),
		generateComplexDescription(index),
		generateTechnicalTerm(index),
		generateBusinessTerm(index),
		findings,
		webSearches,
		generateFeatureName(index),
		generateTechnicalTerm(index),
		generateBusinessTerm(index),
		testCases,
		(index+1)*100,
		(index+1)*500,
	)
}

// generateLargeSection creates a large section of content by repeating patterns
func generateLargeSection(sectionName string, index int, repeatCount int) string {
	baseContent := fmt.Sprintf(`
### %s Section %d

After investigating the codebase, I found several key insights:

1. **Pattern Analysis**: The current implementation uses %s pattern which has implications for %s.
   - Performance impact: %s operations per second
   - Memory usage: Approximately %d MB
   - Scalability concerns: %s

2. **Code Structure**: The %s module is organized into %d layers:
   - Transport layer: Handles %s
   - Business logic: Implements %s
   - Data access: Manages %s

3. **Dependencies**: This component depends on:
   - %s library (version %d.%d.%d)
   - %s framework for %s
   - Internal modules: %s, %s, %s

4. **Test Coverage**: Current coverage is at %d%% with the following breakdown:
   - Unit tests: %d%%
   - Integration tests: %d%%
   - E2E tests: %d%%

5. **Performance Metrics**:
   - Average response time: %dms
   - Throughput: %d requests/second
   - Memory footprint: %dMB
   - CPU usage: %d%%

`,
		sectionName,
		index,
		generateTechnicalTerm(index),
		generateBusinessTerm(index),
		generateRandomTerm(index),
		(index%10+1)*10,
		generateFeatureName(index),
		generateModuleName(index),
		index%5+3,
		generateComponentName(index),
		generateTechnicalTerm((index+1)%10),
		generateBusinessTerm((index+1)%10),
		generateTechnicalTerm((index+2)%10),
		index%10, index%5, index%3,
		generateTechnicalTerm((index+3)%10),
		generateBusinessTerm((index+2)%10),
		generateModuleName((index+1)%10),
		generateModuleName((index+2)%10),
		generateModuleName((index+3)%10),
		75 + index%20,
		80 + index%15,
		70 + index%25,
		60 + index%30,
		(index%9+1)*10,
		(index%9+1)*100,
		(index%9+1)*50,
		(index%9+1)*10,
	)
	
	return strings.Repeat(baseContent, repeatCount)
}