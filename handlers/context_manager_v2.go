package handlers

import (
	"context"
	"log"
	"sync"
	
	"github.com/user/mcp-todo-server/internal/infrastructure/factory"
	ctxkeys "github.com/user/mcp-todo-server/internal/context"
	"github.com/user/mcp-todo-server/utils"
)

// SimpleContextManager manages todo managers for different working directories using new architecture
type SimpleContextManager struct {
	defaultPath string
	managers    map[string]TodoManagerInterface
	mu          sync.RWMutex
}

// NewSimpleContextManager creates a new context manager using repository pattern
func NewSimpleContextManager(defaultPath string) *SimpleContextManager {
	return &SimpleContextManager{
		defaultPath: defaultPath,
		managers:    make(map[string]TodoManagerInterface),
	}
}

// GetManagerForContext returns the appropriate todo manager for the context
func (s *SimpleContextManager) GetManagerForContext(ctx context.Context) TodoManagerInterface {
	var todoPath string
	var err error
	
	// Try to get working directory from context (HTTP header)
	if workingDir, ok := ctx.Value(ctxkeys.WorkingDirectoryKey).(string); ok && workingDir != "" {
		todoPath, err = utils.ResolveTodoPathFromWorkingDir(workingDir)
		if err != nil {
			log.Printf("Failed to resolve todo path from working directory %s: %v", workingDir, err)
			todoPath = s.defaultPath
		}
	} else {
		// Use default path
		todoPath = s.defaultPath
	}
	
	// Check if we already have a manager for this path
	s.mu.RLock()
	if manager, exists := s.managers[todoPath]; exists {
		s.mu.RUnlock()
		return manager
	}
	s.mu.RUnlock()
	
	// Create new manager for this path using the factory
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Double-check after acquiring write lock
	if manager, exists := s.managers[todoPath]; exists {
		return manager
	}
	
	// Create and cache new manager using the new architecture
	manager := factory.CreateTodoManager(todoPath)
	s.managers[todoPath] = manager
	log.Printf("Created new todo manager for path: %s", todoPath)
	
	return manager
}