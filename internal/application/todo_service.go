// Package application provides the application service layer for todo operations.
package application

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/user/mcp-todo-server/internal/domain"
	"github.com/user/mcp-todo-server/internal/domain/repository"
)

// TodoService handles todo business logic
type TodoService struct {
	repo     repository.TodoRepository
	mu       sync.Mutex
	idCounts map[string]int
}

// NewTodoService creates a new todo service
func NewTodoService(repo repository.TodoRepository) *TodoService {
	return &TodoService{
		repo:     repo,
		idCounts: make(map[string]int),
	}
}

// CreateTodo creates a new todo with a unique ID
func (s *TodoService) CreateTodo(ctx context.Context, task, priority, todoType string) (*domain.Todo, error) {
	// Create domain todo
	todo, err := domain.NewTodo(task, priority, todoType)
	if err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	// Generate unique ID
	s.mu.Lock()
	todo.ID = s.generateUniqueID(task)
	s.mu.Unlock()

	// Save via repository
	if err := s.repo.Save(ctx, todo); err != nil {
		return nil, fmt.Errorf("failed to save todo: %w", err)
	}

	return todo, nil
}

// GetTodo retrieves a todo by ID
func (s *TodoService) GetTodo(ctx context.Context, id string) (*domain.Todo, error) {
	return s.repo.FindByID(ctx, id)
}

// GetTodoWithContent retrieves a todo with its full content
func (s *TodoService) GetTodoWithContent(ctx context.Context, id string) (*domain.Todo, string, error) {
	return s.repo.FindByIDWithContent(ctx, id)
}

// ListTodos lists todos based on filters
func (s *TodoService) ListTodos(ctx context.Context, status, priority string, days int) ([]*domain.Todo, error) {
	filters := repository.ListFilters{
		Status:   status,
		Priority: priority,
		Days:     days,
	}

	return s.repo.List(ctx, filters)
}

// UpdateTodoStatus updates the status of a todo
func (s *TodoService) UpdateTodoStatus(ctx context.Context, id string, status string) error {
	todo, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	todo.Status = status
	if status == "completed" {
		todo.Complete()
	}

	return s.repo.Save(ctx, todo)
}

// ArchiveTodo archives a completed todo
func (s *TodoService) ArchiveTodo(ctx context.Context, id string) error {
	todo, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if !todo.IsCompleted() {
		return fmt.Errorf("cannot archive incomplete todo")
	}

	// Determine archive path based on started date
	now := todo.Started
	archivePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())

	return s.repo.Archive(ctx, id, archivePath)
}

// generateUniqueID creates a unique ID from the task description
func (s *TodoService) generateUniqueID(task string) string {
	// Clean and convert to kebab-case
	baseID := generateBaseID(task)

	// Ensure uniqueness
	finalID := baseID
	if count, exists := s.idCounts[baseID]; exists {
		finalID = fmt.Sprintf("%s-%d", baseID, count+1)
		s.idCounts[baseID] = count + 1
	} else {
		s.idCounts[baseID] = 1
	}

	return finalID
}

// generateBaseID creates a kebab-case ID from the task description
func generateBaseID(task string) string {
	// Remove null bytes and other invalid characters
	cleaned := strings.ReplaceAll(task, "\x00", "")

	// Convert to lowercase
	lower := strings.ToLower(cleaned)

	// Replace spaces and special characters with hyphens
	replacer := strings.NewReplacer(
		" ", "-",
		"_", "-",
		":", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
		"/", "-",
		"\\", "-",
		"\"", "",
		"'", "",
		"`", "",
		"~", "",
		"!", "",
		"@", "",
		"#", "",
		"$", "",
		"%", "",
		"^", "",
		"&", "",
		"*", "",
		"+", "",
		"=", "",
		"|", "",
		";", "",
		",", "",
		"<", "",
		">", "",
		"?", "",
		"\n", "-",
		"\r", "-",
		"\t", "-",
	)

	id := replacer.Replace(lower)

	// Replace multiple hyphens with single hyphen
	for strings.Contains(id, "--") {
		id = strings.ReplaceAll(id, "--", "-")
	}

	// Trim hyphens from start and end
	id = strings.Trim(id, "-")

	// Limit length
	if len(id) > 100 {
		id = id[:100]
	}

	// Ensure we have something
	if id == "" {
		id = "todo"
	}

	return id
}
