# Architecture Overview

This document describes the architecture and package structure of the MCP Todo Server after the Phase 4 refactoring.

## Package Structure

```
mcp-todo-server/
├── main.go                 # Entry point
├── server/                 # MCP server implementation
│   └── server.go          # Tool registration and setup
├── handlers/              # MCP tool handlers
│   ├── interfaces.go      # Clean interface definitions
│   ├── todo_handlers_*.go # Handler implementations
│   ├── params_*.go        # Parameter extraction
│   ├── responses_*.go     # Response formatting
│   └── errors.go          # Error handling
├── core/                  # Business logic
│   ├── todo_manager.go    # Todo lifecycle management
│   ├── search_engine.go   # Bleve-based search
│   ├── stats_engine.go    # Analytics
│   ├── template_manager.go # Template system
│   └── todo_linker.go     # Multi-phase relationships
├── internal/              # Internal packages
│   ├── errors/            # Structured error handling
│   │   ├── errors.go      # Sentinel errors and utilities
│   │   └── types.go       # Custom error types
│   ├── testutil/          # Test utilities
│   │   ├── helpers.go     # Test setup helpers
│   │   ├── fixtures.go    # Test data builders
│   │   └── assertions.go  # Custom assertions
│   └── validation/        # Shared validation
│       └── todo_validators.go # Validation functions
└── storage/               # File system operations
    └── paths.go           # Path utilities
```

## Architectural Layers

### 1. Transport Layer (`main.go`, `server/`)

Handles MCP protocol communication:
- STDIO and HTTP transport modes
- Request routing to handlers
- Protocol-level error handling

### 2. Handler Layer (`handlers/`)

Processes MCP tool requests:
- Parameter extraction and validation
- Business logic orchestration
- Response formatting
- Error conversion to MCP responses

**Key Interfaces (Phase 4 improvement):**
```go
// TodoManager - Core todo operations
type TodoManager interface {
    CreateTodo(task, priority, todoType string) (*core.Todo, error)
    ReadTodo(id string) (*core.Todo, error)
    UpdateTodo(id, section, operation, content string, metadata map[string]string) error
    // ... other methods
}

// SearchEngine - Full-text search operations
type SearchEngine interface {
    IndexTodo(todo *core.Todo, content string) error
    SearchTodos(query string, filters map[string]string, limit int) ([]core.SearchResult, error)
    // ... other methods
}

// StatsEngine - Analytics operations
type StatsEngine interface {
    GenerateTodoStats() (*core.TodoStats, error)
    // ... other methods
}

// TemplateManager - Template operations
type TemplateManager interface {
    CreateFromTemplate(template, task, priority, todoType string) (*core.Todo, error)
    // ... other methods
}
```

### 3. Core Layer (`core/`)

Implements business logic:
- Todo lifecycle management
- Search indexing with Bleve
- Statistics calculation
- Template processing
- Multi-phase project linking

### 4. Internal Packages (`internal/`)

Shared utilities and cross-cutting concerns:

**Error Handling (`internal/errors/`):**
- Structured error types (ValidationError, NotFoundError, etc.)
- Error wrapping with context
- Category-based error handling

**Validation (`internal/validation/`):**
- Centralized validation constants
- Reusable validation functions
- Consistent error messages

**Test Utilities (`internal/testutil/`):**
- Test setup helpers with t.Helper()
- Fixture builders for test data
- Custom assertions for domain objects

### 5. Storage Layer (`storage/`)

File system operations:
- Path resolution
- Archive management
- Directory structure maintenance

## Key Design Decisions

### 1. Interface Naming (Phase 4)

Following Go idioms, interfaces no longer have "Interface" suffix:
- `TodoManagerInterface` → `TodoManager`
- `SearchEngineInterface` → `SearchEngine`
- `StatsEngineInterface` → `StatsEngine`
- `TemplateManagerInterface` → `TemplateManager`

### 2. Error Handling

Structured errors with context:
```go
// In core layer
if os.IsNotExist(err) {
    return interrors.NewNotFoundError("todo", id)
}
return interrors.Wrap(err, "failed to read todo")

// In handler layer
if interrors.IsNotFound(err) {
    return mcp.NewToolResultError("Todo not found")
}
```

### 3. Validation Centralization

All validation logic centralized in `internal/validation`:
```go
// Before (duplicated across handlers)
validPriorities := []string{"high", "medium", "low"}
// validation logic...

// After (centralized)
if !validation.IsValidPriority(priority) {
    return fmt.Errorf("invalid priority '%s'", priority)
}
```

### 4. Dependency Injection

Handlers use interfaces for testability:
```go
type TodoHandlers struct {
    manager   TodoManager    // Interface, not concrete type
    search    SearchEngine   // Interface, not concrete type
    stats     StatsEngine    // Interface, not concrete type
    templates TemplateManager // Interface, not concrete type
}
```

## Data Flow

1. **Request Flow:**
   ```
   MCP Client → Transport → Handler → Core → Storage
   ```

2. **Response Flow:**
   ```
   Storage → Core → Handler → Response Formatter → Transport → MCP Client
   ```

3. **Error Flow:**
   ```
   Error → Wrap with Context → Convert to MCP Error → Client
   ```

## Testing Strategy

### Unit Tests
- Mock interfaces for isolated testing
- Table-driven tests for comprehensive coverage
- Error path testing

### Integration Tests
- Real file system operations
- End-to-end tool testing
- Multi-component interactions

### Test Organization
- Unit tests alongside implementation files
- Integration tests with `_integration_test.go` suffix
- Shared test utilities in `internal/testutil`

## Security Considerations

1. **Path Traversal Protection:** All file operations validate paths
2. **Input Validation:** All inputs validated before processing
3. **Error Messages:** Technical details logged, generic messages to users
4. **No Sensitive Data:** No credentials or secrets in todos

## Performance Optimizations

1. **Search Index:** Bleve provides fast full-text search
2. **Daily Archives:** Prevents directory listing performance issues
3. **Lazy Loading:** Content loaded only when needed
4. **Concurrent Operations:** HTTP mode supports multiple connections

## Future Extensibility

The architecture supports future enhancements:
1. **New Todo Types:** Add to validation constants
2. **New Operations:** Add to interfaces and implement
3. **New Storage Backends:** Implement storage interfaces
4. **New Search Features:** Extend SearchEngine interface

## Maintenance Guidelines

1. **Adding New Tools:** Create handler, add to server registration
2. **Adding Validation:** Update `internal/validation` package
3. **Adding Error Types:** Extend `internal/errors` package
4. **Adding Tests:** Use `internal/testutil` helpers

This architecture provides a clean, testable, and maintainable structure for the MCP Todo Server.