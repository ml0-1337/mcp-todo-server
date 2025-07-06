package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/user/mcp-todo-server/handlers"
	"github.com/user/mcp-todo-server/utils"
)

// TodoServer represents the MCP server for todo management
type TodoServer struct {
	mcpServer     *server.MCPServer
	handlers      *handlers.TodoHandlers
	transport     string
	httpServer    *server.StreamableHTTPServer
	httpWrapper   *StreamableHTTPServerWrapper
	startTime     time.Time
	
	closeMu       sync.Mutex
	closed        bool
}

// ServerOption is a function that configures a TodoServer
type ServerOption func(*TodoServer)

// WithTransport sets the transport type
func WithTransport(transport string) ServerOption {
	return func(s *TodoServer) {
		s.transport = transport
	}
}

// NewTodoServer creates a new MCP todo server with all tools registered
func NewTodoServer(opts ...ServerOption) (*TodoServer, error) {
	log.Printf("Creating new TodoServer...")
	
	// Resolve paths dynamically
	todoPath, err := utils.ResolveTodoPath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve todo path: %w", err)
	}

	templatePath, err := utils.ResolveTemplatePath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve template path: %w", err)
	}

	// Create handlers with resolved paths
	todoHandlers, err := handlers.NewTodoHandlers(todoPath, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create handlers: %w", err)
	}

	// Create MCP server instance
	s := server.NewMCPServer(
		"MCP Todo Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Create todo server wrapper with default transport
	ts := &TodoServer{
		mcpServer: s,
		handlers:  todoHandlers,
		transport: "stdio",
		startTime: time.Now(),
	}

	// Apply options
	for _, opt := range opts {
		opt(ts)
	}

	// Register all tools
	ts.registerTools()

	// Create HTTP server if needed
	if ts.transport == "http" {
		ts.httpServer = server.NewStreamableHTTPServer(s)
		// Wrap with middleware for header extraction
		ts.httpWrapper = NewStreamableHTTPServerWrapper(ts.httpServer)
	}

	return ts, nil
}

// ListTools returns all registered tools
func (ts *TodoServer) ListTools() []mcp.Tool {
	// The mark3labs/mcp-go library doesn't expose ListTools directly,
	// so we'll maintain our own list for testing
	tools := []mcp.Tool{
		mcp.NewTool("todo_create", mcp.WithDescription("Create a new todo")),
		mcp.NewTool("todo_create_multi", mcp.WithDescription("Create multiple todos with parent-child relationships")),
		mcp.NewTool("todo_read", mcp.WithDescription("Read todo(s)")),
		mcp.NewTool("todo_update", mcp.WithDescription("Update a todo")),
		mcp.NewTool("todo_search", mcp.WithDescription("Search todos")),
		mcp.NewTool("todo_archive", mcp.WithDescription("Archive a todo")),
		mcp.NewTool("todo_template", mcp.WithDescription("Create from template")),
		mcp.NewTool("todo_link", mcp.WithDescription("Link related todos")),
		mcp.NewTool("todo_stats", mcp.WithDescription("Get todo statistics")),
		mcp.NewTool("todo_clean", mcp.WithDescription("Clean up todos")),
	}
	return tools
}

// registerTools registers all todo management tools
func (ts *TodoServer) registerTools() {
	// Register todo_create
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_create",
			mcp.WithDescription("Create a new todo with full metadata. TIP: Use parent_id for phases and subtasks. Types 'phase' and 'subtask' require parent_id."),
			mcp.WithString("task",
				mcp.Required(),
				mcp.Description("Task description")),
			mcp.WithString("priority",
				mcp.Description("Task priority (high, medium, low)"),
				mcp.DefaultString("high")),
			mcp.WithString("type",
				mcp.Description("Todo type (feature, bug, refactor, research, multi-phase, phase, subtask)"),
				mcp.DefaultString("feature")),
			mcp.WithString("template",
				mcp.Description("Optional template name")),
			mcp.WithString("parent_id",
				mcp.Description("Parent todo ID (required for phase/subtask types)")),
		),
		ts.handlers.HandleTodoCreate,
	)

	// Register todo_create_multi
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_create_multi",
			mcp.WithDescription("Create multiple todos with parent-child relationships in one operation. Perfect for multi-phase projects."),
			mcp.WithObject("parent",
				mcp.Required(),
				mcp.Description("Parent todo information"),
				mcp.Properties(map[string]any{
					"task": map[string]any{
						"type":        "string",
						"description": "Parent task description",
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Priority (high, medium, low)",
						"default":     "high",
					},
					"type": map[string]any{
						"type":        "string",
						"description": "Todo type (defaults to multi-phase)",
						"default":     "multi-phase",
					},
				})),
			mcp.WithArray("children",
				mcp.Required(),
				mcp.Description("Array of child todos"),
				mcp.Items(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"task": map[string]any{
							"type":        "string",
							"description": "Child task description",
						},
						"priority": map[string]any{
							"type":        "string",
							"description": "Priority (high, medium, low)",
							"default":     "medium",
						},
						"type": map[string]any{
							"type":        "string",
							"description": "Todo type (defaults to phase)",
							"default":     "phase",
						},
					},
					"required": []string{"task"},
				})),
		),
		ts.handlers.HandleTodoCreateMulti,
	)

	// Register todo_read
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_read",
			mcp.WithDescription("Read single todo or list all todos"),
			mcp.WithString("id",
				mcp.Description("Specific todo ID")),
			mcp.WithObject("filter",
				mcp.Description("Filter options"),
				mcp.Properties(map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "Status filter (in_progress, completed, blocked, all)",
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Priority filter (high, medium, low, all)",
					},
					"days": map[string]any{
						"type":        "number",
						"description": "Todos from last N days",
					},
				})),
			mcp.WithString("format",
				mcp.Description("Output format (full, summary, list)"),
				mcp.DefaultString("summary")),
		),
		ts.handlers.HandleTodoRead,
	)

	// Register todo_update
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_update",
			mcp.WithDescription("Update todo content or metadata"),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Todo ID to update")),
			mcp.WithString("section",
				mcp.Description("Section to update (status, findings, tests, checklist, scratchpad)")),
			mcp.WithString("operation",
				mcp.Description("Update operation (append, replace, prepend, toggle)"),
				mcp.DefaultString("append")),
			mcp.WithString("content",
				mcp.Description("Content to add/update")),
			mcp.WithObject("metadata",
				mcp.Description("Metadata to update"),
				mcp.Properties(map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "Todo status",
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Todo priority",
					},
					"current_test": map[string]any{
						"type":        "string",
						"description": "Current test being worked on",
					},
				})),
		),
		ts.handlers.HandleTodoUpdate,
	)

	// Register todo_search
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_search",
			mcp.WithDescription("Full-text search across all todos"),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Search terms")),
			mcp.WithArray("scope",
				mcp.Description("Search scope (task, findings, tests, all)"),
				mcp.Items(map[string]any{"type": "string"})),
			mcp.WithObject("filters",
				mcp.Description("Search filters"),
				mcp.Properties(map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "Status filter",
					},
					"date_from": map[string]any{
						"type":        "string",
						"description": "Date in YYYY-MM-DD format",
					},
					"date_to": map[string]any{
						"type":        "string",
						"description": "Date in YYYY-MM-DD format",
					},
				})),
			mcp.WithNumber("limit",
				mcp.Description("Maximum results"),
				mcp.DefaultNumber(20),
				mcp.Max(100)),
		),
		ts.handlers.HandleTodoSearch,
	)

	// Register todo_archive
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_archive",
			mcp.WithDescription("Archive completed todo to daily folder"),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Todo ID to archive")),
		),
		ts.handlers.HandleTodoArchive,
	)

	// Register todo_template
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_template",
			mcp.WithDescription("Create todo from template or list templates"),
			mcp.WithString("template",
				mcp.Description("Template name (leave empty to list)")),
			mcp.WithString("task",
				mcp.Description("Task description")),
			mcp.WithString("priority",
				mcp.Description("Task priority"),
				mcp.DefaultString("high")),
			mcp.WithString("type",
				mcp.Description("Todo type"),
				mcp.DefaultString("feature")),
		),
		ts.handlers.HandleTodoTemplate,
	)

	// Register todo_link
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_link",
			mcp.WithDescription("Link related todos"),
			mcp.WithString("parent_id",
				mcp.Required(),
				mcp.Description("Parent todo ID")),
			mcp.WithString("child_id",
				mcp.Required(),
				mcp.Description("Child todo ID")),
			mcp.WithString("link_type",
				mcp.Description("Type of link (parent-child, blocks, relates-to)"),
				mcp.DefaultString("parent-child")),
		),
		ts.handlers.HandleTodoLink,
	)

	// Register todo_stats
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_stats",
			mcp.WithDescription("Get todo statistics and analytics"),
			mcp.WithString("period",
				mcp.Description("Time period for stats (all, week, month, quarter)"),
				mcp.DefaultString("all")),
		),
		ts.handlers.HandleTodoStats,
	)

	// Register todo_clean
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_clean",
			mcp.WithDescription("Clean up and manage todos"),
			mcp.WithString("operation",
				mcp.Description("Cleanup operation (archive_old, find_duplicates)"),
				mcp.DefaultString("archive_old")),
			mcp.WithNumber("days",
				mcp.Description("Days threshold for archive_old"),
				mcp.DefaultNumber(90)),
		),
		ts.handlers.HandleTodoClean,
	)
}

// Close cleans up server resources
func (ts *TodoServer) Close() error {
	ts.closeMu.Lock()
	defer ts.closeMu.Unlock()
	
	// Check if already closed
	if ts.closed {
		return nil
	}
	
	// Mark as closed
	ts.closed = true
	
	// Stop cleanup routine if present
	if ts.httpWrapper != nil {
		ts.httpWrapper.Stop()
	}
	
	// Shutdown HTTP server if present
	if ts.httpServer != nil {
		ctx := context.Background()
		if err := ts.httpServer.Shutdown(ctx); err != nil {
			return err
		}
	}
	
	// Close handlers
	if ts.handlers != nil {
		return ts.handlers.Close()
	}
	
	return nil
}

// Start starts the MCP server (deprecated - use StartStdio or StartHTTP)
func (ts *TodoServer) Start() error {
	return ts.StartStdio()
}

// StartStdio starts the MCP server in STDIO mode
func (ts *TodoServer) StartStdio() error {
	log.Printf("Starting STDIO server...")
	err := server.ServeStdio(ts.mcpServer)
	if err != nil {
		log.Printf("STDIO server error: %v", err)
	}
	return err
}

// StartHTTP starts the MCP server in HTTP mode
func (ts *TodoServer) StartHTTP(addr string) error {
	if ts.httpWrapper == nil {
		return fmt.Errorf("HTTP server not initialized")
	}
	
	// Use custom HTTP server with middleware
	http.Handle("/mcp", ts.httpWrapper)
	
	// Add health check endpoint
	http.HandleFunc("/health", ts.handleHealthCheck)
	
	// Configure server with proper timeouts for connection resilience
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	log.Printf("Starting HTTP server with middleware on %s", addr)
	return server.ListenAndServe()
}

// handleHealthCheck handles the /health endpoint
func (ts *TodoServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Calculate uptime
	uptime := time.Since(ts.startTime)
	
	// Get session count if using HTTP
	sessionCount := 0
	if ts.httpWrapper != nil && ts.httpWrapper.sessionManager != nil {
		ts.httpWrapper.sessionManager.mu.RLock()
		sessionCount = len(ts.httpWrapper.sessionManager.sessions)
		ts.httpWrapper.sessionManager.mu.RUnlock()
	}
	
	// Build health response
	health := map[string]interface{}{
		"status":       "healthy",
		"uptime":       uptime.String(),
		"uptimeMs":     uptime.Milliseconds(),
		"serverTime":   time.Now().Format(time.RFC3339),
		"transport":    ts.transport,
		"version":      "2.0.0",
		"sessions":     sessionCount,
	}
	
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// Write response
	if err := json.NewEncoder(w).Encode(health); err != nil {
		log.Printf("Error encoding health response: %v", err)
	}
}
