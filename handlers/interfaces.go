package handlers

import (
	"github.com/user/mcp-todo-server/core"
	"time"
)

// TodoManager defines the interface for todo management operations
type TodoManager interface {
	CreateTodo(task, priority, todoType string) (*core.Todo, error)
	ReadTodo(id string) (*core.Todo, error)
	ReadTodoWithContent(id string) (*core.Todo, string, error)
	UpdateTodo(id, section, operation, content string, metadata map[string]string) error
	SaveTodo(todo *core.Todo) error
	ListTodos(status, priority string, days int) ([]*core.Todo, error)
	ReadTodoContent(id string) (string, error)
	ArchiveTodo(id string) error
	ArchiveOldTodos(days int) (int, error)
	FindDuplicateTodos() ([][]string, error)
	GetBasePath() string
}

// SearchEngine defines the interface for search operations
type SearchEngine interface {
	IndexTodo(todo *core.Todo, content string) error
	DeleteTodo(id string) error
	SearchTodos(queryStr string, filters map[string]string, limit int) ([]core.SearchResult, error)
	Close() error
	GetIndexedCount() (uint64, error)
}

// StatsEngine defines the interface for statistics operations
type StatsEngine interface {
	GenerateTodoStats() (*core.TodoStats, error)
	GenerateTodoStatsForPeriod(period string) (*core.TodoStats, error)
	CalculateCompletionRatesByType() (map[string]float64, error)
	CalculateCompletionRatesByPriority() (map[string]float64, error)
	CalculateAverageCompletionTime() (time.Duration, error)
	CalculateTestCoverage(todoID string) (float64, error)
}

// TemplateManager defines the interface for template operations
type TemplateManager interface {
	LoadTemplate(name string) (*core.Template, error)
	ListTemplates() ([]string, error)
	CreateFromTemplate(templateName, task, priority, todoType string) (*core.Todo, error)
	ExecuteTemplate(tmpl *core.Template, vars map[string]interface{}) (string, error)
}

// TodoLinker defines the interface for linking operations
type TodoLinker interface {
	LinkTodos(parentID, childID, linkType string) error
}
