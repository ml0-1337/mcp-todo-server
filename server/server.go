package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/user/mcp-todo-server/handlers"
	ctxkeys "github.com/user/mcp-todo-server/internal/context"
	"github.com/user/mcp-todo-server/internal/logging"
	"github.com/user/mcp-todo-server/utils"
)

// TodoServer represents the MCP server for todo management
type TodoServer struct {
	mcpServer         *server.MCPServer
	handlers          *handlers.TodoHandlers
	transport         string
	
	// HTTP transport layers (each serves a specific purpose):
	httpServer        *server.StreamableHTTPServer    // Base MCP HTTP server from mark3labs/mcp-go
	stableTransport   *StableHTTPTransport            // Stability wrapper - fixes connection issues, adds queuing & heartbeats
	httpWrapper       *StreamableHTTPServerWrapper    // Middleware layer - adds session management & header extraction
	
	startTime         time.Time
	sessionTimeout    time.Duration
	managerTimeout    time.Duration
	heartbeatInterval time.Duration
	noAutoArchive     bool
	
	// HTTP timeout configurations
	requestTimeout    time.Duration
	httpReadTimeout   time.Duration
	httpWriteTimeout  time.Duration
	httpIdleTimeout   time.Duration
	
	closeMu           sync.Mutex
	closed            bool
}

// ServerOption is a function that configures a TodoServer
type ServerOption func(*TodoServer)

// WithTransport sets the transport type
func WithTransport(transport string) ServerOption {
	return func(s *TodoServer) {
		s.transport = transport
	}
}

// WithSessionTimeout sets the session timeout duration
func WithSessionTimeout(timeout time.Duration) ServerOption {
	return func(s *TodoServer) {
		s.sessionTimeout = timeout
	}
}

// WithManagerTimeout sets the manager timeout duration
func WithManagerTimeout(timeout time.Duration) ServerOption {
	return func(s *TodoServer) {
		s.managerTimeout = timeout
	}
}

// WithHeartbeatInterval sets the heartbeat interval duration
func WithHeartbeatInterval(interval time.Duration) ServerOption {
	return func(s *TodoServer) {
		s.heartbeatInterval = interval
	}
}

// WithNoAutoArchive sets whether to disable auto-archiving of completed todos
func WithNoAutoArchive(noAutoArchive bool) ServerOption {
	return func(s *TodoServer) {
		s.noAutoArchive = noAutoArchive
	}
}

// WithHTTPRequestTimeout sets the HTTP request timeout
func WithHTTPRequestTimeout(timeout time.Duration) ServerOption {
	return func(s *TodoServer) {
		s.requestTimeout = timeout
	}
}

// WithHTTPReadTimeout sets the HTTP server read timeout
func WithHTTPReadTimeout(timeout time.Duration) ServerOption {
	return func(s *TodoServer) {
		s.httpReadTimeout = timeout
	}
}

// WithHTTPWriteTimeout sets the HTTP server write timeout
func WithHTTPWriteTimeout(timeout time.Duration) ServerOption {
	return func(s *TodoServer) {
		s.httpWriteTimeout = timeout
	}
}

// WithHTTPIdleTimeout sets the HTTP server idle timeout
func WithHTTPIdleTimeout(timeout time.Duration) ServerOption {
	return func(s *TodoServer) {
		s.httpIdleTimeout = timeout
	}
}

// NewTodoServer creates a new MCP todo server with all tools registered
func NewTodoServer(opts ...ServerOption) (*TodoServer, error) {
	logging.Infof("Creating new TodoServer...")
	
	// Resolve paths dynamically
	todoPath, err := utils.ResolveTodoPath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve todo path: %w", err)
	}

	templatePath, err := utils.ResolveTemplatePath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve template path: %w", err)
	}

	// Don't create handlers yet - we need to apply options first to get timeout values
	// Create MCP server instance
	s := server.NewMCPServer(
		"MCP Todo Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Create todo server wrapper with default transport
	ts := &TodoServer{
		mcpServer:         s,
		transport:         "stdio",
		startTime:         time.Now(),
		sessionTimeout:    7 * 24 * time.Hour, // Default: 7 days
		managerTimeout:    24 * time.Hour,     // Default: 24 hours
		heartbeatInterval: 30 * time.Second,    // Default: 30 seconds
		requestTimeout:    30 * time.Second,    // Default: 30 seconds
		httpReadTimeout:   60 * time.Second,    // Default: 60 seconds
		httpWriteTimeout:  60 * time.Second,    // Default: 60 seconds
		httpIdleTimeout:   120 * time.Second,   // Default: 120 seconds (unchanged)
	}

	// Apply options
	for _, opt := range opts {
		opt(ts)
	}

	// Now create handlers with the configured manager timeout
	logging.Infof("Creating handlers with todoPath=%s, templatePath=%s, managerTimeout=%v, noAutoArchive=%v", todoPath, templatePath, ts.managerTimeout, ts.noAutoArchive)
	todoHandlers, err := handlers.NewTodoHandlers(todoPath, templatePath, ts.managerTimeout, ts.noAutoArchive)
	if err != nil {
		return nil, fmt.Errorf("failed to create handlers: %w", err)
	}
	ts.handlers = todoHandlers
	logging.Infof("Handlers created successfully")

	// Register all tools
	logging.Infof("Registering MCP tools...")
	ts.registerTools()
	logging.Infof("Tools registered successfully")

	// Create HTTP server if needed
	if ts.transport == "http" {
		// HTTP mode uses a 3-layer architecture for stability:
		// 1. Base StreamableHTTPServer (MCP protocol implementation)
		// 2. StableHTTPTransport (fixes connection stability issues)
		// 3. StreamableHTTPServerWrapper (adds session management)
		
		// Create HTTP server with context function to pass through request context
		options := []server.StreamableHTTPOption{
			server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
				// Extract working directory from header
				workingDir := r.Header.Get("X-Working-Directory")
				if workingDir != "" {
					ctx = context.WithValue(ctx, ctxkeys.WorkingDirectoryKey, workingDir)
				}
				
				// Extract session ID from header
				sessionID := r.Header.Get("Mcp-Session-Id")
				if sessionID != "" {
					ctx = context.WithValue(ctx, ctxkeys.SessionIDKey, sessionID)
				}
				
				return ctx
			}),
		}
		
		// Add heartbeat interval if configured
		if ts.heartbeatInterval > 0 {
			logging.Infof("Configuring heartbeat interval: %v", ts.heartbeatInterval)
			options = append(options, server.WithHeartbeatInterval(ts.heartbeatInterval))
		} else {
			logging.Infof("Heartbeat disabled (interval=0)")
		}
		
		ts.httpServer = server.NewStreamableHTTPServer(s, options...)
		
		// Create stable transport wrapper
		ts.stableTransport = NewStableHTTPTransport(
			ts.httpServer,
			WithRequestTimeout(ts.requestTimeout),
			WithConnectionTimeout(ts.sessionTimeout),
			WithMaxRequestsPerConnection(1000),
		)
		
		// Wrap with middleware for header extraction
		ts.httpWrapper = NewStreamableHTTPServerWrapper(ts.stableTransport, ts.sessionTimeout)
	}

	return ts, nil
}

// ListTools returns all registered tools
func (ts *TodoServer) ListTools() []mcp.Tool {
	// The mark3labs/mcp-go library doesn't expose ListTools directly,
	// so we'll maintain our own list for testing
	tools := []mcp.Tool{
		mcp.NewTool("todo_create", mcp.WithDescription("Start tracking a new task, feature, or bug. Creates a markdown file to capture your work progress, findings, and test results.")),
		mcp.NewTool("todo_create_multi", mcp.WithDescription("Plan a multi-phase project by creating a parent task and all its phases at once. Automatically links phases together for easy progress tracking.")),
		mcp.NewTool("todo_read", mcp.WithDescription("View your tasks and their current status. Check a specific todo's details or see all active work with filtering options.")),
		mcp.NewTool("todo_update", mcp.WithDescription("Add progress notes, test results, or findings to a todo. Update status when blocked or completed (auto-archives on completion).")),
		mcp.NewTool("todo_search", mcp.WithDescription("Find past solutions, code snippets, or similar work across all your todos. Searches through task descriptions, findings, and test results.")),
	}
	
	// Only include todo_archive if auto-archive is disabled
	if ts.noAutoArchive {
		tools = append(tools, mcp.NewTool("todo_archive", mcp.WithDescription("Move a completed todo to the archive folder organized by date. Usually happens automatically when you mark a todo as completed.")))
	}
	
	tools = append(tools, []mcp.Tool{
		mcp.NewTool("todo_template", mcp.WithDescription("Start with a pre-structured todo for common tasks. Templates include sections and checklists tailored to specific workflows.")),
		mcp.NewTool("todo_link", mcp.WithDescription("Connect related tasks together. Useful for dependencies, blocking relationships, or grouping related work.")),
		mcp.NewTool("todo_stats", mcp.WithDescription("View your productivity metrics: completed tasks, time spent, task distribution, and work patterns.")),
		mcp.NewTool("todo_clean", mcp.WithDescription("Maintain your todo system by archiving old incomplete tasks or finding potential duplicates.")),
	}...)
	
	return tools
}

// registerTools registers all todo management tools
func (ts *TodoServer) registerTools() {
	// Register todo_create
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_create",
			mcp.WithDescription("Start tracking a new task, feature, or bug. Creates a markdown file to capture your work progress, findings, and test results. Use for single tasks or as phases within larger projects."),
			mcp.WithString("task",
				mcp.Required(),
				mcp.Description("What you want to accomplish (e.g., 'Add user authentication', 'Fix login timeout bug')")),
			mcp.WithString("priority",
				mcp.Description("How urgent is this? (high=today/tomorrow, medium=this week, low=when possible)"),
				mcp.DefaultString("high")),
			mcp.WithString("type",
				mcp.Description("What kind of work? (feature=new capability, bug=fix issue, refactor=improve code, research=investigate)"),
				mcp.DefaultString("feature")),
			mcp.WithString("template",
				mcp.Description("Use a pre-built structure (bug-fix, feature, research, refactor, tdd-cycle)")),
			mcp.WithString("parent_id",
				mcp.Description("Link to parent task when breaking down large projects into phases")),
		),
		ts.handlers.HandleTodoCreate,
	)

	// Register todo_create_multi
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_create_multi",
			mcp.WithDescription("Plan a multi-phase project by creating a parent task and all its phases at once. Automatically links phases together for easy progress tracking."),
			mcp.WithObject("parent",
				mcp.Required(),
				mcp.Description("The main project or feature"),
				mcp.Properties(map[string]any{
					"task": map[string]any{
						"type":        "string",
						"description": "Overall project goal (e.g., 'Implement user authentication system')",
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Project urgency (high=this sprint, medium=next sprint, low=backlog)",
						"default":     "high",
					},
					"type": map[string]any{
						"type":        "string",
						"description": "Usually 'multi-phase' for projects with multiple steps",
						"default":     "multi-phase",
					},
				})),
			mcp.WithArray("children",
				mcp.Required(),
				mcp.Description("List of phases or subtasks to complete the project"),
				mcp.Items(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"task": map[string]any{
							"type":        "string",
							"description": "Phase description (e.g., 'Design database schema', 'Create login UI')",
						},
						"priority": map[string]any{
							"type":        "string",
							"description": "Phase urgency (typically matches parent priority)",
							"default":     "medium",
						},
						"type": map[string]any{
							"type":        "string",
							"description": "Usually 'phase' for project phases",
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
			mcp.WithDescription("View your tasks and their current status. Check a specific todo's details or see all active work with filtering options."),
			mcp.WithString("id",
				mcp.Description("View specific todo (e.g., 'implement-auth-feature'). Leave empty to list all")),
			mcp.WithObject("filter",
				mcp.Description("Options to filter the todo list"),
				mcp.Properties(map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "Show only todos in this state (in_progress=active work, completed=done, blocked=waiting)",
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Show only this urgency level (high, medium, low, all)",
					},
					"days": map[string]any{
						"type":        "number",
						"description": "Show todos from the last N days (e.g., 7 for past week)",
					},
				})),
			mcp.WithString("format",
				mcp.Description("How much detail? (full=everything, summary=overview, list=just titles)"),
				mcp.DefaultString("summary")),
		),
		ts.handlers.HandleTodoRead,
	)

	// Register todo_update
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_update",
			mcp.WithDescription("Add progress notes, test results, or findings to a todo. Update status when blocked or completed (auto-archives on completion)."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Which todo to update (e.g., 'fix-login-bug')")),
			mcp.WithString("section",
				mcp.Description("Where to add content (findings=research notes, tests=test results, checklist=task items, scratchpad=rough notes)")),
			mcp.WithString("operation",
				mcp.Description("How to add content (append=add to end, replace=overwrite, prepend=add to beginning, toggle=check/uncheck)"),
				mcp.DefaultString("append")),
			mcp.WithString("content",
				mcp.Description("Your notes, code, test results, or findings to add")),
			mcp.WithObject("metadata",
				mcp.Description("Update todo properties (status, priority, current test)"),
				mcp.Properties(map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "Change work state (in_progress=working, completed=done & archive, blocked=waiting)",
					},
					"priority": map[string]any{
						"type":        "string",
						"description": "Change urgency (high, medium, low)",
					},
					"current_test": map[string]any{
						"type":        "string",
						"description": "Track which test you're working on (e.g., 'Test 3: user validation')",
					},
				})),
		),
		ts.handlers.HandleTodoUpdate,
	)

	// Register todo_search
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_search",
			mcp.WithDescription("Find past solutions, code snippets, or similar work across all your todos. Searches through task descriptions, findings, and test results."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("What to search for (e.g., 'authentication', 'timeout bug', 'JWT implementation')")),
			mcp.WithArray("scope",
				mcp.Description("Where to search (task=titles only, findings=research notes, tests=test code, all=everywhere)"),
				mcp.Items(map[string]any{"type": "string"})),
			mcp.WithObject("filters",
				mcp.Description("Additional search filters to narrow results"),
				mcp.Properties(map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "Search only in todos with this state",
					},
					"date_from": map[string]any{
						"type":        "string",
						"description": "Start date for search range (YYYY-MM-DD)",
					},
					"date_to": map[string]any{
						"type":        "string",
						"description": "End date for search range (YYYY-MM-DD)",
					},
				})),
			mcp.WithNumber("limit",
				mcp.Description("Maximum results to return (default 20, max 100)"),
				mcp.DefaultNumber(20),
				mcp.Max(100)),
		),
		ts.handlers.HandleTodoSearch,
	)

	// Register todo_archive only if auto-archive is disabled
	if ts.noAutoArchive {
		ts.mcpServer.AddTool(
			mcp.NewTool("todo_archive",
				mcp.WithDescription("Move a completed todo to the archive folder organized by date. Usually happens automatically when you mark a todo as completed."),
				mcp.WithString("id",
					mcp.Required(),
					mcp.Description("Todo to archive (e.g., 'implement-feature')")),
			),
			ts.handlers.HandleTodoArchive,
		)
	}

	// Register todo_template
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_template",
			mcp.WithDescription("Start with a pre-structured todo for common tasks. Templates include sections and checklists tailored to specific workflows."),
			mcp.WithString("template",
				mcp.Description("Template name (bug-fix, feature, research, refactor, tdd-cycle). Leave empty to see all available")),
			mcp.WithString("task",
				mcp.Description("Your specific task description to fill the template")),
			mcp.WithString("priority",
				mcp.Description("How urgent? (high, medium, low)"),
				mcp.DefaultString("high")),
			mcp.WithString("type",
				mcp.Description("Work type (usually matches template type)"),
				mcp.DefaultString("feature")),
		),
		ts.handlers.HandleTodoTemplate,
	)

	// Register todo_link
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_link",
			mcp.WithDescription("Connect related tasks together. Useful for dependencies, blocking relationships, or grouping related work."),
			mcp.WithString("parent_id",
				mcp.Required(),
				mcp.Description("The main/blocking todo (e.g., 'design-api')")),
			mcp.WithString("child_id",
				mcp.Required(),
				mcp.Description("The dependent/related todo (e.g., 'implement-endpoints')")),
			mcp.WithString("link_type",
				mcp.Description("Relationship type (parent-child=breakdown, blocks=dependency, relates-to=connection)"),
				mcp.DefaultString("parent-child")),
		),
		ts.handlers.HandleTodoLink,
	)

	// Register todo_stats
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_stats",
			mcp.WithDescription("View your productivity metrics: completed tasks, time spent, task distribution, and work patterns."),
			mcp.WithString("period",
				mcp.Description("Time range to analyze (all=everything, week=last 7 days, month=last 30 days, quarter=last 90 days)"),
				mcp.DefaultString("all")),
		),
		ts.handlers.HandleTodoStats,
	)

	// Register todo_clean
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_clean",
			mcp.WithDescription("Maintain your todo system by archiving old incomplete tasks or finding potential duplicates."),
			mcp.WithString("operation",
				mcp.Description("What to clean (archive_old=move stale todos, find_duplicates=identify similar tasks)"),
				mcp.DefaultString("archive_old")),
			mcp.WithNumber("days",
				mcp.Description("For archive_old: how many days before considering a todo stale (default 90)"),
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
	
	// Shutdown stable transport if present
	if ts.stableTransport != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := ts.stableTransport.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error shutting down stable transport: %v\n", err)
		}
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


// StartStdio starts the MCP server in STDIO mode
func (ts *TodoServer) StartStdio() error {
	logging.Infof("StartStdio called, starting MCP STDIO server...")
	err := server.ServeStdio(ts.mcpServer)
	if err != nil {
		logging.Errorf("STDIO server error: %v", err)
	}
	logging.Infof("StartStdio returning")
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
	
	// Add heartbeat endpoint for stable transport
	http.HandleFunc("/mcp/heartbeat", ts.handleHeartbeat)
	
	// Add debug endpoints
	http.HandleFunc("/debug/connections", ts.handleDebugConnections)
	http.HandleFunc("/debug/sessions", ts.handleDebugSessions)
	http.HandleFunc("/debug/transport", ts.handleDebugTransport)
	
	// Configure server with proper timeouts for connection resilience
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  ts.httpReadTimeout,
		WriteTimeout: ts.httpWriteTimeout,
		IdleTimeout:  ts.httpIdleTimeout,
	}
	
	logging.Infof("Starting HTTP server with middleware on %s", addr)
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
		fmt.Fprintf(os.Stderr, "Error encoding health response: %v\n", err)
	}
}

// handleDebugConnections shows active connection information
func (ts *TodoServer) handleDebugConnections(w http.ResponseWriter, r *http.Request) {
	if ts.httpWrapper == nil || ts.httpWrapper.sessionManager == nil {
		http.Error(w, "Debug endpoint only available in HTTP mode", http.StatusNotImplemented)
		return
	}
	
	// Get session stats
	stats := ts.httpWrapper.sessionManager.GetSessionStats()
	
	// Add server info
	debug := map[string]interface{}{
		"server": map[string]interface{}{
			"uptime":     time.Since(ts.startTime).String(),
			"startTime":  ts.startTime.Format(time.RFC3339),
			"transport":  ts.transport,
		},
		"sessions": stats,
		"request": map[string]interface{}{
			"remoteAddr": r.RemoteAddr,
			"userAgent":  r.Header.Get("User-Agent"),
			"headers":    r.Header,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(debug)
}

// handleDebugSessions shows detailed session information
func (ts *TodoServer) handleDebugSessions(w http.ResponseWriter, r *http.Request) {
	if ts.httpWrapper == nil || ts.httpWrapper.sessionManager == nil {
		http.Error(w, "Debug endpoint only available in HTTP mode", http.StatusNotImplemented)
		return
	}
	
	sm := ts.httpWrapper.sessionManager
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	sessions := make([]map[string]interface{}, 0)
	for id, session := range sm.sessions {
		sessions = append(sessions, map[string]interface{}{
			"id":               id,
			"workingDirectory": session.WorkingDirectory,
			"lastActivity":     session.LastActivity.Format(time.RFC3339),
			"inactiveDuration": time.Since(session.LastActivity).String(),
		})
	}
	
	response := map[string]interface{}{
		"totalSessions": len(sessions),
		"sessions":      sessions,
		"serverTime":    time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDebugTransport shows stable transport metrics
func (ts *TodoServer) handleDebugTransport(w http.ResponseWriter, r *http.Request) {
	if ts.stableTransport == nil {
		http.Error(w, "Stable transport not available", http.StatusNotImplemented)
		return
	}
	
	metrics := ts.stableTransport.GetMetrics()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleHeartbeat handles heartbeat requests for the stable transport
func (ts *TodoServer) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	// This is handled by the stable transport's ServeHTTP method
	// Just forward it through the transport
	if ts.stableTransport != nil {
		ts.stableTransport.ServeHTTP(w, r)
	} else {
		http.Error(w, "Heartbeat not available", http.StatusNotImplemented)
	}
}
