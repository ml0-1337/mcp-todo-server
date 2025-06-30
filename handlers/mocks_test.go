package handlers

import (
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/user/mcp-todo-server/core"
	"sync"
	"time"
)

// MockCall represents a method call with arguments
type MockCall struct {
	Method string
	Args   []interface{}
}

// MockTodoManager is a mock implementation of TodoManagerInterface
type MockTodoManager struct {
	mu    sync.Mutex
	calls []MockCall

	// Configurable responses
	CreateTodoFunc          func(task, priority, todoType string) (*core.Todo, error)
	ReadTodoFunc            func(id string) (*core.Todo, error)
	ReadTodoWithContentFunc func(id string) (*core.Todo, string, error)
	UpdateTodoFunc          func(id, section, operation, content string, metadata map[string]string) error
	SaveTodoFunc            func(todo *core.Todo) error
	ListTodosFunc           func(status, priority string, days int) ([]*core.Todo, error)
	ReadTodoContentFunc     func(id string) (string, error)
	ArchiveTodoFunc         func(id, quarter string) error
	ArchiveOldTodosFunc     func(days int) (int, error)
	FindDuplicateTodosFunc  func() ([][]string, error)
	GetBasePathFunc         func() string
}

func NewMockTodoManager() *MockTodoManager {
	return &MockTodoManager{
		calls: make([]MockCall, 0),
	}
}

func (m *MockTodoManager) recordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, MockCall{Method: method, Args: args})
}

func (m *MockTodoManager) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MockCall{}, m.calls...)
}

func (m *MockTodoManager) CreateTodo(task, priority, todoType string) (*core.Todo, error) {
	m.recordCall("CreateTodo", task, priority, todoType)
	if m.CreateTodoFunc != nil {
		return m.CreateTodoFunc(task, priority, todoType)
	}
	// Default implementation
	return &core.Todo{
		ID:       fmt.Sprintf("test-%d", time.Now().Unix()),
		Task:     task,
		Started:  time.Now(),
		Status:   "in_progress",
		Priority: priority,
		Type:     todoType,
	}, nil
}

func (m *MockTodoManager) ReadTodo(id string) (*core.Todo, error) {
	m.recordCall("ReadTodo", id)
	if m.ReadTodoFunc != nil {
		return m.ReadTodoFunc(id)
	}
	return &core.Todo{ID: id, Task: "Test Todo", Status: "in_progress"}, nil
}

func (m *MockTodoManager) ReadTodoWithContent(id string) (*core.Todo, string, error) {
	m.recordCall("ReadTodoWithContent", id)
	if m.ReadTodoWithContentFunc != nil {
		return m.ReadTodoWithContentFunc(id)
	}
	// Return default todo and content
	todo := &core.Todo{ID: id, Task: "Test Todo", Status: "in_progress"}
	content := "# Test Todo\n\n## Checklist\n- [ ] Item 1"
	return todo, content, nil
}

func (m *MockTodoManager) UpdateTodo(id, section, operation, content string, metadata map[string]string) error {
	m.recordCall("UpdateTodo", id, section, operation, content, metadata)
	if m.UpdateTodoFunc != nil {
		return m.UpdateTodoFunc(id, section, operation, content, metadata)
	}
	return nil
}

func (m *MockTodoManager) SaveTodo(todo *core.Todo) error {
	m.recordCall("SaveTodo", todo)
	if m.SaveTodoFunc != nil {
		return m.SaveTodoFunc(todo)
	}
	return nil
}

func (m *MockTodoManager) ListTodos(status, priority string, days int) ([]*core.Todo, error) {
	m.recordCall("ListTodos", status, priority, days)
	if m.ListTodosFunc != nil {
		return m.ListTodosFunc(status, priority, days)
	}
	return []*core.Todo{}, nil
}

func (m *MockTodoManager) ReadTodoContent(id string) (string, error) {
	m.recordCall("ReadTodoContent", id)
	if m.ReadTodoContentFunc != nil {
		return m.ReadTodoContentFunc(id)
	}
	return "# Test Todo Content", nil
}

func (m *MockTodoManager) ArchiveTodo(id, quarter string) error {
	m.recordCall("ArchiveTodo", id, quarter)
	if m.ArchiveTodoFunc != nil {
		return m.ArchiveTodoFunc(id, quarter)
	}
	return nil
}

func (m *MockTodoManager) ArchiveOldTodos(days int) (int, error) {
	m.recordCall("ArchiveOldTodos", days)
	if m.ArchiveOldTodosFunc != nil {
		return m.ArchiveOldTodosFunc(days)
	}
	return 5, nil
}

func (m *MockTodoManager) FindDuplicateTodos() ([][]string, error) {
	m.recordCall("FindDuplicateTodos")
	if m.FindDuplicateTodosFunc != nil {
		return m.FindDuplicateTodosFunc()
	}
	return [][]string{}, nil
}

func (m *MockTodoManager) GetBasePath() string {
	m.recordCall("GetBasePath")
	if m.GetBasePathFunc != nil {
		return m.GetBasePathFunc()
	}
	return "/test/path"
}

// MockSearchEngine is a mock implementation of SearchEngineInterface
type MockSearchEngine struct {
	mu    sync.Mutex
	calls []MockCall

	IndexTodoFunc       func(todo *core.Todo, content string) error
	DeleteTodoFunc      func(id string) error
	SearchTodosFunc     func(queryStr string, filters map[string]string, limit int) ([]core.SearchResult, error)
	CloseFunc           func() error
	GetIndexedCountFunc func() (uint64, error)
}

func NewMockSearchEngine() *MockSearchEngine {
	return &MockSearchEngine{
		calls: make([]MockCall, 0),
	}
}

func (m *MockSearchEngine) recordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, MockCall{Method: method, Args: args})
}

func (m *MockSearchEngine) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MockCall{}, m.calls...)
}

func (m *MockSearchEngine) IndexTodo(todo *core.Todo, content string) error {
	m.recordCall("IndexTodo", todo, content)
	if m.IndexTodoFunc != nil {
		return m.IndexTodoFunc(todo, content)
	}
	return nil
}

func (m *MockSearchEngine) DeleteTodo(id string) error {
	m.recordCall("DeleteTodo", id)
	if m.DeleteTodoFunc != nil {
		return m.DeleteTodoFunc(id)
	}
	return nil
}

func (m *MockSearchEngine) SearchTodos(queryStr string, filters map[string]string, limit int) ([]core.SearchResult, error) {
	m.recordCall("SearchTodos", queryStr, filters, limit)
	if m.SearchTodosFunc != nil {
		return m.SearchTodosFunc(queryStr, filters, limit)
	}
	return []core.SearchResult{}, nil
}

func (m *MockSearchEngine) Close() error {
	m.recordCall("Close")
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockSearchEngine) GetIndexedCount() (uint64, error) {
	m.recordCall("GetIndexedCount")
	if m.GetIndexedCountFunc != nil {
		return m.GetIndexedCountFunc()
	}
	return 0, nil
}

// MockStatsEngine is a mock implementation of StatsEngineInterface
type MockStatsEngine struct {
	mu    sync.Mutex
	calls []MockCall

	GenerateTodoStatsFunc                  func() (*core.TodoStats, error)
	CalculateCompletionRatesByTypeFunc     func() (map[string]float64, error)
	CalculateCompletionRatesByPriorityFunc func() (map[string]float64, error)
	CalculateAverageCompletionTimeFunc     func() (time.Duration, error)
	CalculateTestCoverageFunc              func(todoID string) (float64, error)
}

func NewMockStatsEngine() *MockStatsEngine {
	return &MockStatsEngine{
		calls: make([]MockCall, 0),
	}
}

func (m *MockStatsEngine) recordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, MockCall{Method: method, Args: args})
}

func (m *MockStatsEngine) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MockCall{}, m.calls...)
}

func (m *MockStatsEngine) GenerateTodoStats() (*core.TodoStats, error) {
	m.recordCall("GenerateTodoStats")
	if m.GenerateTodoStatsFunc != nil {
		return m.GenerateTodoStatsFunc()
	}
	return &core.TodoStats{
		TotalTodos:      10,
		CompletedTodos:  5,
		InProgressTodos: 3,
		BlockedTodos:    2,
	}, nil
}

func (m *MockStatsEngine) CalculateCompletionRatesByType() (map[string]float64, error) {
	m.recordCall("CalculateCompletionRatesByType")
	if m.CalculateCompletionRatesByTypeFunc != nil {
		return m.CalculateCompletionRatesByTypeFunc()
	}
	return map[string]float64{"feature": 0.75}, nil
}

func (m *MockStatsEngine) CalculateCompletionRatesByPriority() (map[string]float64, error) {
	m.recordCall("CalculateCompletionRatesByPriority")
	if m.CalculateCompletionRatesByPriorityFunc != nil {
		return m.CalculateCompletionRatesByPriorityFunc()
	}
	return map[string]float64{"high": 0.80}, nil
}

func (m *MockStatsEngine) CalculateAverageCompletionTime() (time.Duration, error) {
	m.recordCall("CalculateAverageCompletionTime")
	if m.CalculateAverageCompletionTimeFunc != nil {
		return m.CalculateAverageCompletionTimeFunc()
	}
	return 24 * time.Hour, nil
}

func (m *MockStatsEngine) CalculateTestCoverage(todoID string) (float64, error) {
	m.recordCall("CalculateTestCoverage", todoID)
	if m.CalculateTestCoverageFunc != nil {
		return m.CalculateTestCoverageFunc(todoID)
	}
	return 0.85, nil
}

// MockTemplateManager is a mock implementation of TemplateManagerInterface
type MockTemplateManager struct {
	mu    sync.Mutex
	calls []MockCall

	LoadTemplateFunc       func(name string) (*core.Template, error)
	ListTemplatesFunc      func() ([]string, error)
	CreateFromTemplateFunc func(templateName, task, priority, todoType string) (*core.Todo, error)
	ExecuteTemplateFunc    func(tmpl *core.Template, vars map[string]interface{}) (string, error)
}

func NewMockTemplateManager() *MockTemplateManager {
	return &MockTemplateManager{
		calls: make([]MockCall, 0),
	}
}

func (m *MockTemplateManager) recordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, MockCall{Method: method, Args: args})
}

func (m *MockTemplateManager) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MockCall{}, m.calls...)
}

func (m *MockTemplateManager) LoadTemplate(name string) (*core.Template, error) {
	m.recordCall("LoadTemplate", name)
	if m.LoadTemplateFunc != nil {
		return m.LoadTemplateFunc(name)
	}
	return &core.Template{Name: name, Description: "Test template"}, nil
}

func (m *MockTemplateManager) ListTemplates() ([]string, error) {
	m.recordCall("ListTemplates")
	if m.ListTemplatesFunc != nil {
		return m.ListTemplatesFunc()
	}
	return []string{"bug-fix", "feature", "research"}, nil
}

func (m *MockTemplateManager) CreateFromTemplate(templateName, task, priority, todoType string) (*core.Todo, error) {
	m.recordCall("CreateFromTemplate", templateName, task, priority, todoType)
	if m.CreateFromTemplateFunc != nil {
		return m.CreateFromTemplateFunc(templateName, task, priority, todoType)
	}
	return &core.Todo{
		ID:       fmt.Sprintf("template-%d", time.Now().Unix()),
		Task:     task,
		Started:  time.Now(),
		Status:   "in_progress",
		Priority: priority,
		Type:     todoType,
	}, nil
}

func (m *MockTemplateManager) ExecuteTemplate(tmpl *core.Template, vars map[string]interface{}) (string, error) {
	m.recordCall("ExecuteTemplate", tmpl, vars)
	if m.ExecuteTemplateFunc != nil {
		return m.ExecuteTemplateFunc(tmpl, vars)
	}
	return "Executed template content", nil
}

// MockTodoLinker is a mock implementation of TodoLinkerInterface
type MockTodoLinker struct {
	mu    sync.Mutex
	calls []MockCall

	LinkTodosFunc func(parentID, childID, linkType string) error
}

func NewMockTodoLinker() *MockTodoLinker {
	return &MockTodoLinker{
		calls: make([]MockCall, 0),
	}
}

func (m *MockTodoLinker) recordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, MockCall{Method: method, Args: args})
}

func (m *MockTodoLinker) GetCalls() []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MockCall{}, m.calls...)
}

func (m *MockTodoLinker) LinkTodos(parentID, childID, linkType string) error {
	m.recordCall("LinkTodos", parentID, childID, linkType)
	if m.LinkTodosFunc != nil {
		return m.LinkTodosFunc(parentID, childID, linkType)
	}
	return nil
}

// MockCallToolRequest wraps arguments for testing
type MockCallToolRequest struct {
	Arguments map[string]interface{}
}

// GetArguments implements mcp.CallToolRequest interface
func (m *MockCallToolRequest) GetArguments() map[string]interface{} {
	return m.Arguments
}

// ToCallToolRequest converts MockCallToolRequest to mcp.CallToolRequest
func (m *MockCallToolRequest) ToCallToolRequest() mcp.CallToolRequest {
	// Create a minimal CallToolRequest with our arguments
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: m.Arguments,
		},
	}
}
