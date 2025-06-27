package server

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// TodoServer represents the MCP server for todo management
type TodoServer struct {
	mcpServer *server.MCPServer
}

// NewTodoServer creates a new MCP todo server with all tools registered
func NewTodoServer() *TodoServer {
	// Create MCP server instance
	s := server.NewMCPServer(
		"MCP Todo Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	
	// Create todo server wrapper
	ts := &TodoServer{
		mcpServer: s,
	}
	
	// Register all tools
	ts.registerTools()
	
	return ts
}

// ListTools returns all registered tools
func (ts *TodoServer) ListTools() []mcp.Tool {
	// The mark3labs/mcp-go library doesn't expose ListTools directly,
	// so we'll maintain our own list for testing
	tools := []mcp.Tool{
		mcp.NewTool("todo_create", mcp.WithDescription("Create a new todo")),
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
			mcp.WithDescription("Create a new todo with full metadata"),
		),
		ts.handleTodoCreate,
	)
	
	// Register todo_read
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_read",
			mcp.WithDescription("Read todo(s)"),
		),
		ts.handleTodoRead,
	)
	
	// Register todo_update
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_update",
			mcp.WithDescription("Update a todo"),
		),
		ts.handleTodoUpdate,
	)
	
	// Register todo_search
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_search",
			mcp.WithDescription("Search todos"),
		),
		ts.handleTodoSearch,
	)
	
	// Register todo_archive
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_archive",
			mcp.WithDescription("Archive a todo"),
		),
		ts.handleTodoArchive,
	)
	
	// Register todo_template
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_template",
			mcp.WithDescription("Create from template"),
		),
		ts.handleTodoTemplate,
	)
	
	// Register todo_link
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_link",
			mcp.WithDescription("Link related todos"),
		),
		ts.handleTodoLink,
	)
	
	// Register todo_stats
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_stats",
			mcp.WithDescription("Get todo statistics"),
		),
		ts.handleTodoStats,
	)
	
	// Register todo_clean
	ts.mcpServer.AddTool(
		mcp.NewTool("todo_clean",
			mcp.WithDescription("Clean up todos"),
		),
		ts.handleTodoClean,
	)
}

// Tool handler stubs - minimal implementation for now
func (ts *TodoServer) handleTodoCreate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Validate required parameters
	task, err := request.RequireString("task")
	if err != nil {
		return mcp.NewToolResultError("Missing required parameter: task"), nil
	}
	
	// Get optional parameters with defaults
	priority := request.GetString("priority", "high")
	
	// Validate priority value
	if priority != "high" && priority != "medium" && priority != "low" {
		priority = "high" // Use default for invalid values
	}
	
	// For now, just return success with the validated parameters
	return mcp.NewToolResultText("Todo created: " + task + " (priority: " + priority + ")"), nil
}

func (ts *TodoServer) handleTodoRead(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("todo_read not implemented"), nil
}

func (ts *TodoServer) handleTodoUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("todo_update not implemented"), nil
}

func (ts *TodoServer) handleTodoSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("todo_search not implemented"), nil
}

func (ts *TodoServer) handleTodoArchive(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("todo_archive not implemented"), nil
}

func (ts *TodoServer) handleTodoTemplate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("todo_template not implemented"), nil
}

func (ts *TodoServer) handleTodoLink(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("todo_link not implemented"), nil
}

func (ts *TodoServer) handleTodoStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("todo_stats not implemented"), nil
}

func (ts *TodoServer) handleTodoClean(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("todo_clean not implemented"), nil
}

// Start starts the MCP server
func (ts *TodoServer) Start() error {
	return server.ServeStdio(ts.mcpServer)
}