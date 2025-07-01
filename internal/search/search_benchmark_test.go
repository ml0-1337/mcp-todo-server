package search

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
	
	"github.com/user/mcp-todo-server/internal/domain"
)

// BenchmarkSearchEngine tests search performance with various dataset sizes
func BenchmarkSearchEngine(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			// Setup
			tempDir := b.TempDir()
			
			// Create search engine
			indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
			searchEngine, err := NewEngine(indexPath, filepath.Join(tempDir, ".claude", "todos"))
			if err != nil {
				b.Fatalf("Failed to create search engine: %v", err)
			}
			defer searchEngine.Close()
			
			// Create todo manager
			manager := NewTestTodoManager(tempDir)
			
			// Populate with test data
			b.Logf("Creating %d todos for benchmark...", size)
			start := time.Now()
			
			for i := 0; i < size; i++ {
				task := fmt.Sprintf("Task %d: %s feature implementation with priority handling", i, generateSearchTerms(i))
				priority := []string{"high", "medium", "low"}[i%3]
				todoType := []string{"feature", "bug", "refactor", "test", "docs"}[i%5]
				
				todo, err := manager.CreateTodo(task, priority, todoType)
				if err != nil {
					b.Fatalf("Failed to create todo %d: %v", i, err)
				}
				
				// Index with rich content
				content := fmt.Sprintf(`
# Task: %s

## Description
This is a detailed description for task %d. It includes various keywords like:
- Implementation details: %s
- Technical requirements: %s
- Business logic: %s

## Test Cases
1. Test case for %s functionality
2. Edge case handling for %s
3. Performance testing for %s

## Notes
Created for benchmark testing with index %d
Tags: %s
`, task, i, generateTechnicalTerm(i), generateBusinessTerm(i), generateRandomTerm(i),
					generateFeatureName(i), generateModuleName(i), generateComponentName(i),
					i, generateTags(i))
				
				if err := searchEngine.Index(todo, content); err != nil {
					b.Fatalf("Failed to index todo %d: %v", i, err)
				}
			}
			
			indexTime := time.Since(start)
			b.Logf("Indexed %d todos in %v (%.2f todos/sec)", size, indexTime, float64(size)/indexTime.Seconds())
			
			// Benchmark search operations
			queries := []string{
				"implementation",
				"priority high",
				"feature AND test",
				"bug OR refactor",
				"performance testing",
				"technical requirements",
				"task 500",
				"functionality",
				"edge case",
				"business logic",
			}
			
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				query := queries[i%len(queries)]
				
				results, err := searchEngine.Search(query, nil, 20)
				if err != nil {
					b.Fatalf("Search failed for query '%s': %v", query, err)
				}
				
				// Verify we got results
				if len(results) == 0 && size > 100 {
					b.Errorf("No results for query '%s' with %d todos", query, size)
				}
			}
			
			b.StopTimer()
			
			// Report additional metrics
			b.ReportMetric(float64(size), "todos")
			b.ReportMetric(indexTime.Seconds(), "index_time_sec")
		})
	}
}

// BenchmarkSearchComplexQueries tests performance of complex search queries
func BenchmarkSearchComplexQueries(b *testing.B) {
	// Setup
	tempDir, err := os.MkdirTemp("", "search-complex-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create search engine and manager
	indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
	searchEngine, err := NewEngine(indexPath, filepath.Join(tempDir, ".claude", "todos"))
	if err != nil {
		b.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()
	
	manager := NewTestTodoManager(tempDir)
	
	// Create diverse dataset
	todoCount := 5000
	for i := 0; i < todoCount; i++ {
		task := fmt.Sprintf("Task %d: %s", i, generateComplexDescription(i))
		priority := []string{"high", "medium", "low"}[i%3]
		todoType := []string{"feature", "bug", "refactor", "test", "docs"}[i%5]
		
		todo, err := manager.CreateTodo(task, priority, todoType)
		if err != nil {
			b.Fatalf("Failed to create todo: %v", err)
		}
		
		content := generateComplexContent(i)
		if err := searchEngine.Index(todo, content); err != nil {
			b.Fatalf("Failed to index todo: %v", err)
		}
	}
	
	// Define complex queries
	complexQueries := []struct {
		name  string
		query string
	}{
		{"PhraseMatch", `"implement authentication"`},
		{"WildcardSearch", "implement*"},
		{"FuzzySearch", "implementaion~"},
		{"RangeQuery", "priority:[high TO medium]"},
		{"BooleanComplex", "(feature OR bug) AND priority:high AND NOT archived"},
		{"NestedBoolean", "((authentication OR authorization) AND security) OR (login AND user)"},
		{"MixedQuery", `type:feature AND "user interface" AND (responsive OR mobile)`},
		{"DateRange", "started:[2025-01-01 TO 2025-12-31]"},
	}
	
	// Run benchmarks for each query type
	for _, tc := range complexQueries {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				results, err := searchEngine.Search(tc.query, nil, 50)
				if err != nil {
					b.Fatalf("Search failed for query '%s': %v", tc.query, err)
				}
				
				// Use results to prevent optimization
				_ = len(results)
			}
		})
	}
}

// BenchmarkSearchConcurrent tests concurrent search performance
func BenchmarkSearchConcurrent(b *testing.B) {
	// Setup
	tempDir, err := os.MkdirTemp("", "search-concurrent-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create search engine
	indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
	searchEngine, err := NewEngine(indexPath, filepath.Join(tempDir, ".claude", "todos"))
	if err != nil {
		b.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()
	
	manager := NewTestTodoManager(tempDir)
	
	// Create dataset
	todoCount := 5000
	for i := 0; i < todoCount; i++ {
		task := fmt.Sprintf("Task %d: Concurrent test %s", i, generateSearchTerms(i))
		todo, err := manager.CreateTodo(task, "medium", "test")
		if err != nil {
			b.Fatalf("Failed to create todo: %v", err)
		}
		
		if err := searchEngine.Index(todo, task); err != nil {
			b.Fatalf("Failed to index todo: %v", err)
		}
	}
	
	queries := []string{
		"concurrent",
		"test",
		"task",
		"implementation",
		"feature",
	}
	
	// Benchmark with different concurrency levels
	concurrencyLevels := []int{1, 5, 10, 20}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrent-%d", concurrency), func(b *testing.B) {
			b.SetParallelism(concurrency)
			b.ResetTimer()
			
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					query := queries[i%len(queries)]
					results, err := searchEngine.Search(query, nil, 20)
					if err != nil {
						b.Fatalf("Search failed: %v", err)
					}
					_ = len(results)
					i++
				}
			})
		})
	}
}

// BenchmarkIndexingPerformance tests todo indexing speed
func BenchmarkIndexingPerformance(b *testing.B) {
	// Setup
	tempDir, err := os.MkdirTemp("", "index-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create search engine
	indexPath := filepath.Join(tempDir, ".claude", "index", "todos.bleve")
	searchEngine, err := NewEngine(indexPath, filepath.Join(tempDir, ".claude", "todos"))
	if err != nil {
		b.Fatalf("Failed to create search engine: %v", err)
	}
	defer searchEngine.Close()
	
	manager := NewTestTodoManager(tempDir)
	
	// Pre-create todos
	todos := make([]*domain.Todo, b.N)
	for i := 0; i < b.N; i++ {
		todo, err := manager.CreateTodo(
			fmt.Sprintf("Benchmark todo %d", i),
			"medium",
			"test",
		)
		if err != nil {
			b.Fatalf("Failed to create todo: %v", err)
		}
		todos[i] = todo
	}
	
	// Benchmark indexing
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		content := fmt.Sprintf("Content for todo %d with searchable text", i)
		if err := searchEngine.Index(todos[i], content); err != nil {
			b.Fatalf("Failed to index todo: %v", err)
		}
	}
	
	b.StopTimer()
	
	// Report indexing rate
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "todos/sec")
}

// Helper functions for generating test data
func generateSearchTerms(i int) string {
	terms := []string{
		"authentication", "authorization", "database", "api", "frontend",
		"backend", "middleware", "security", "performance", "optimization",
		"refactoring", "testing", "deployment", "configuration", "monitoring",
	}
	return terms[i%len(terms)]
}

func generateTechnicalTerm(i int) string {
	terms := []string{
		"microservices", "kubernetes", "docker", "redis", "postgresql",
		"elasticsearch", "kafka", "rabbitmq", "nginx", "prometheus",
	}
	return terms[i%len(terms)]
}

func generateBusinessTerm(i int) string {
	terms := []string{
		"customer", "revenue", "analytics", "reporting", "dashboard",
		"metrics", "kpi", "roi", "conversion", "engagement",
	}
	return terms[i%len(terms)]
}

func generateRandomTerm(i int) string {
	terms := []string{
		"alpha", "beta", "gamma", "delta", "epsilon",
		"zeta", "eta", "theta", "iota", "kappa",
	}
	return terms[i%len(terms)]
}

func generateFeatureName(i int) string {
	features := []string{
		"user-management", "payment-processing", "notification-system",
		"search-functionality", "reporting-module", "analytics-dashboard",
		"api-gateway", "data-pipeline", "cache-layer", "queue-system",
	}
	return features[i%len(features)]
}

func generateModuleName(i int) string {
	modules := []string{
		"core", "handlers", "utils", "storage", "transport",
		"auth", "api", "web", "cli", "sdk",
	}
	return modules[i%len(modules)]
}

func generateComponentName(i int) string {
	components := []string{
		"controller", "service", "repository", "model", "view",
		"middleware", "helper", "validator", "transformer", "adapter",
	}
	return components[i%len(components)]
}

func generateTags(i int) string {
	tags := []string{
		"urgent", "blocked", "in-review", "needs-testing", "ready",
		"wip", "done", "archived", "deprecated", "experimental",
	}
	tag1 := tags[i%len(tags)]
	tag2 := tags[(i+3)%len(tags)]
	return fmt.Sprintf("%s, %s", tag1, tag2)
}

func generateComplexDescription(i int) string {
	descriptions := []string{
		"Implement authentication system with JWT tokens and refresh mechanism",
		"Optimize database queries for better performance in user dashboard",
		"Refactor legacy code to use modern design patterns and clean architecture",
		"Add comprehensive unit tests for payment processing module",
		"Fix critical security vulnerability in file upload functionality",
		"Develop responsive user interface for mobile devices",
		"Integrate third-party API for real-time data synchronization",
		"Implement caching layer to reduce database load",
		"Create automated deployment pipeline with CI/CD",
		"Document REST API endpoints with OpenAPI specification",
	}
	return descriptions[i%len(descriptions)]
}

func generateComplexContent(i int) string {
	return fmt.Sprintf(`
# Task %d: Complex Content

## Overview
This task involves implementing %s with consideration for %s and %s.

## Technical Requirements
- Framework: %s
- Database: %s
- Cache: %s
- Message Queue: %s

## Implementation Details
The implementation should follow %s pattern and ensure %s.
Key considerations include:
1. %s
2. %s
3. %s

## Testing Strategy
- Unit tests for %s
- Integration tests for %s
- Performance tests for %s
- End-to-end tests for %s

## Security Considerations
- Authentication: %s
- Authorization: %s
- Data encryption: %s
- Input validation: %s

## Performance Targets
- Response time: < %dms
- Throughput: > %d requests/second
- Memory usage: < %dMB
- CPU usage: < %d%%

## Dependencies
- %s version %d.%d
- %s version %d.%d
- %s version %d.%d

Tags: %s
Priority: %s
Assigned: Team %d
Sprint: %d
`, 
		i,
		generateFeatureName(i),
		generateTechnicalTerm(i),
		generateBusinessTerm(i),
		generateTechnicalTerm((i+1)%10),
		generateTechnicalTerm((i+2)%10),
		generateTechnicalTerm((i+3)%10),
		generateTechnicalTerm((i+4)%10),
		generateComponentName(i),
		generateBusinessTerm((i+1)%10),
		generateFeatureName((i+1)%10),
		generateFeatureName((i+2)%10),
		generateFeatureName((i+3)%10),
		generateModuleName(i),
		generateModuleName((i+1)%10),
		generateModuleName((i+2)%10),
		generateModuleName((i+3)%10),
		generateRandomTerm(i),
		generateRandomTerm((i+1)%10),
		generateRandomTerm((i+2)%10),
		generateRandomTerm((i+3)%10),
		(i%9+1)*10,
		(i%9+1)*100,
		(i%9+1)*50,
		(i%9+1)*10,
		generateTechnicalTerm((i+5)%10), i%10, i%5,
		generateTechnicalTerm((i+6)%10), (i+1)%10, (i+1)%5,
		generateTechnicalTerm((i+7)%10), (i+2)%10, (i+2)%5,
		generateTags(i),
		[]string{"low", "medium", "high", "critical"}[i%4],
		i%5+1,
		i%4+1,
	)
}