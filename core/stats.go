package core

import (
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	interrors "github.com/user/mcp-todo-server/internal/errors"
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

// calculateCompletionRates is a generic helper for calculating completion rates by a given field
func (se *StatsEngine) calculateCompletionRates(getField func(*Todo) string, defaultValue string) (map[string]float64, error) {
	todos, err := se.getAllTodos()
	if err != nil {
		return nil, err
	}

	// Count by field and status
	totalByField := make(map[string]int)
	completedByField := make(map[string]int)

	for _, todo := range todos {
		fieldValue := getField(todo)
		if fieldValue == "" {
			fieldValue = defaultValue
		}

		totalByField[fieldValue]++
		if todo.Status == "completed" {
			completedByField[fieldValue]++
		}
	}

	// Calculate rates
	rates := make(map[string]float64)
	for fieldValue, total := range totalByField {
		if total > 0 {
			completed := completedByField[fieldValue]
			rate := float64(completed) / float64(total) * 100.0
			// Round to 1 decimal place
			rates[fieldValue] = math.Round(rate*10) / 10
		}
	}

	return rates, nil
}

// CalculateCompletionRatesByType calculates completion percentage by todo type
func (se *StatsEngine) CalculateCompletionRatesByType() (map[string]float64, error) {
	return se.calculateCompletionRates(func(t *Todo) string { return t.Type }, "unknown")
}

// CalculateCompletionRatesByPriority calculates completion percentage by priority
func (se *StatsEngine) CalculateCompletionRatesByPriority() (map[string]float64, error) {
	return se.calculateCompletionRates(func(t *Todo) string { return t.Priority }, "medium")
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
	todoPath, err := ResolveTodoPath(se.manager.basePath, todoID)
	if err != nil {
		return 0, err
	}
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

	return se.generateStatsFromTodos(todos)
}

// GenerateTodoStatsForPeriod generates statistics filtered by time period
func (se *StatsEngine) GenerateTodoStatsForPeriod(period string) (*TodoStats, error) {
	todos, err := se.getAllTodos()
	if err != nil {
		return nil, err
	}

	// Filter todos by period
	filteredTodos := se.filterTodosByPeriod(todos, period)

	return se.generateStatsFromTodos(filteredTodos)
}

// filterTodosByPeriod filters todos based on the specified period
func (se *StatsEngine) filterTodosByPeriod(todos []*Todo, period string) []*Todo {
	if period == "all" || period == "" {
		return todos
	}

	now := time.Now()
	var cutoffTime time.Time

	switch period {
	case "week":
		cutoffTime = now.AddDate(0, 0, -7)
	case "month":
		cutoffTime = now.AddDate(0, 0, -30)
	case "quarter":
		cutoffTime = now.AddDate(0, 0, -90)
	case "year":
		cutoffTime = now.AddDate(0, 0, -365)
	default:
		// Invalid period, return all todos
		return todos
	}

	var filtered []*Todo
	for _, todo := range todos {
		// Include todo if it was started after the cutoff time
		if !todo.Started.IsZero() && todo.Started.After(cutoffTime) {
			filtered = append(filtered, todo)
		}
	}

	return filtered
}

// generateStatsFromTodos generates statistics from a list of todos
func (se *StatsEngine) generateStatsFromTodos(todos []*Todo) (*TodoStats, error) {
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

	// Calculate completion rates only from filtered todos
	completedByType := make(map[string]int)
	totalByType := make(map[string]int)

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
	for todoType, total := range totalByType {
		if total > 0 {
			completed := completedByType[todoType]
			rate := float64(completed) / float64(total) * 100.0
			stats.CompletionRates[todoType] = math.Round(rate*10) / 10
		}
	}

	// Calculate average completion time only from filtered todos
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

	if completedCount > 0 {
		stats.AverageCompletionTime = totalDuration / time.Duration(completedCount)
	}

	return stats, nil
}

// Helper to get all todos
func (se *StatsEngine) getAllTodos() ([]*Todo, error) {
	// Read all .md files in the todos directory tree
	todosDir := filepath.Join(se.manager.basePath, ".claude", "todos")

	var todos []*Todo

	// Walk through all subdirectories
	err := filepath.Walk(todosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// If the todos directory doesn't exist, return empty slice
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Process only .md files
		if strings.HasSuffix(info.Name(), ".md") {
			// Extract ID from filename
			todoID := strings.TrimSuffix(info.Name(), ".md")

			// Read the todo
			todo, err := se.manager.ReadTodo(todoID)
			if err != nil {
				// Skip files that can't be read as todos
				return nil
			}

			todos = append(todos, todo)
		}

		return nil
	})

	if err != nil {
		return nil, interrors.Wrap(err, "failed to walk todos directory")
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
