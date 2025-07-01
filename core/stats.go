package core

import (
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// TodoStats represents aggregated statistics about todos
type TodoStats struct {
	TotalTodos            int
	CompletedTodos        int
	InProgressTodos       int
	BlockedTodos          int
	TodosByType           map[string]int
	TodosByPriority       map[string]int
	CompletionRates       map[string]float64
	AverageCompletionTime time.Duration
}

// StatsEngine calculates statistics from todo data
type StatsEngine struct {
	manager *TodoManager
}

// NewStatsEngine creates a new stats engine
func NewStatsEngine(manager *TodoManager) *StatsEngine {
	return &StatsEngine{
		manager: manager,
	}
}

// CalculateCompletionRatesByType calculates completion percentage by todo type
func (se *StatsEngine) CalculateCompletionRatesByType() (map[string]float64, error) {
	todos, err := se.getAllTodos()
	if err != nil {
		return nil, err
	}

	// Count by type and status
	totalByType := make(map[string]int)
	completedByType := make(map[string]int)

	for _, todo := range todos {
		todoType := todo.Type
		if todoType == "" {
			todoType = "unknown"
		}

		totalByType[todoType]++
		if todo.Status == "completed" {
			completedByType[todoType]++
		}
	}

	// Calculate rates
	rates := make(map[string]float64)
	for todoType, total := range totalByType {
		if total > 0 {
			completed := completedByType[todoType]
			rate := float64(completed) / float64(total) * 100.0
			// Round to 1 decimal place
			rates[todoType] = math.Round(rate*10) / 10
		}
	}

	return rates, nil
}

// CalculateCompletionRatesByPriority calculates completion percentage by priority
func (se *StatsEngine) CalculateCompletionRatesByPriority() (map[string]float64, error) {
	todos, err := se.getAllTodos()
	if err != nil {
		return nil, err
	}

	// Count by priority and status
	totalByPriority := make(map[string]int)
	completedByPriority := make(map[string]int)

	for _, todo := range todos {
		priority := todo.Priority
		if priority == "" {
			priority = "medium"
		}

		totalByPriority[priority]++
		if todo.Status == "completed" {
			completedByPriority[priority]++
		}
	}

	// Calculate rates
	rates := make(map[string]float64)
	for priority, total := range totalByPriority {
		if total > 0 {
			completed := completedByPriority[priority]
			rate := float64(completed) / float64(total) * 100.0
			// Round to 1 decimal place
			rates[priority] = math.Round(rate*10) / 10
		}
	}

	return rates, nil
}

// CalculateAverageCompletionTime calculates average time from start to completion
func (se *StatsEngine) CalculateAverageCompletionTime() (time.Duration, error) {
	todos, err := se.getAllTodos()
	if err != nil {
		return 0, err
	}

	var totalDuration time.Duration
	completedCount := 0

	for _, todo := range todos {
		if todo.Status == "completed" && !todo.Started.IsZero() && !todo.Completed.IsZero() {
			duration := todo.Completed.Sub(todo.Started)
			if duration > 0 {
				totalDuration += duration
				completedCount++
			}
		}
	}

	if completedCount == 0 {
		return 0, nil
	}

	return totalDuration / time.Duration(completedCount), nil
}

// CalculateTestCoverage calculates test coverage from Test List checkmarks
func (se *StatsEngine) CalculateTestCoverage(todoID string) (float64, error) {
	_, err := se.manager.ReadTodo(todoID)
	if err != nil {
		return 0, err
	}

	// Read the full content to get test section
	todoPath := filepath.Join(se.manager.basePath, todoID+".md")
	content, err := ioutil.ReadFile(todoPath)
	if err != nil {
		return 0, err
	}

	// Extract test list section
	contentStr := string(content)
	testListSection := se.extractSection(contentStr, "## Test List", "##")

	// Also try extracting from test section if not found
	if testListSection == "" {
		testListSection = se.extractSection(contentStr, "## Test Cases", "##")
	}

	if testListSection == "" {
		// No test list found
		return 0, nil
	}

	// Count checked and total test items
	checkedPattern := regexp.MustCompile(`- \[x\]`)
	uncheckedPattern := regexp.MustCompile(`- \[ \]`)

	checkedCount := len(checkedPattern.FindAllString(testListSection, -1))
	uncheckedCount := len(uncheckedPattern.FindAllString(testListSection, -1))
	totalCount := checkedCount + uncheckedCount

	if totalCount == 0 {
		return 0, nil
	}

	coverage := float64(checkedCount) / float64(totalCount) * 100.0
	return math.Round(coverage*10) / 10, nil
}

// GenerateTodoStats generates comprehensive statistics
func (se *StatsEngine) GenerateTodoStats() (*TodoStats, error) {
	todos, err := se.getAllTodos()
	if err != nil {
		return nil, err
	}

	stats := &TodoStats{
		TodosByType:     make(map[string]int),
		TodosByPriority: make(map[string]int),
		CompletionRates: make(map[string]float64),
	}

	// Count todos by status, type, and priority
	for _, todo := range todos {
		stats.TotalTodos++

		// Count by status
		switch todo.Status {
		case "completed":
			stats.CompletedTodos++
		case "in_progress":
			stats.InProgressTodos++
		case "blocked":
			stats.BlockedTodos++
		}

		// Count by type
		todoType := todo.Type
		if todoType == "" {
			todoType = "unknown"
		}
		stats.TodosByType[todoType]++

		// Count by priority
		priority := todo.Priority
		if priority == "" {
			priority = "medium"
		}
		stats.TodosByPriority[priority]++
	}

	// Calculate completion rates
	typeRates, err := se.CalculateCompletionRatesByType()
	if err == nil {
		for k, v := range typeRates {
			stats.CompletionRates[k] = v
		}
	}

	// Calculate average completion time
	stats.AverageCompletionTime, _ = se.CalculateAverageCompletionTime()

	return stats, nil
}

// Helper to get all todos
func (se *StatsEngine) getAllTodos() ([]*Todo, error) {
	// Read all .md files in the todos directory
	todosDir := filepath.Join(se.manager.basePath, ".claude", "todos")
	files, err := ioutil.ReadDir(todosDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read todos directory: %w", err)
	}

	var todos []*Todo
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			// Extract ID from filename
			todoID := strings.TrimSuffix(file.Name(), ".md")

			// Read the todo
			todo, err := se.manager.ReadTodo(todoID)
			if err != nil {
				// Skip files that can't be read as todos
				continue
			}

			todos = append(todos, todo)
		}
	}

	return todos, nil
}

// Helper to extract a section from markdown content
func (se *StatsEngine) extractSection(content, startMarker, endMarker string) string {
	lines := strings.Split(content, "\n")
	inSection := false
	var sectionLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, startMarker) {
			inSection = true
			continue
		}

		if inSection && endMarker != "" && strings.HasPrefix(line, endMarker) && line != startMarker {
			break
		}

		if inSection {
			sectionLines = append(sectionLines, line)
		}
	}

	return strings.Join(sectionLines, "\n")
}
