# Go MCP Server Research

## Overview

Model Context Protocol (MCP) is a standard protocol that enables Large Language Models (LLMs) to interact with external tools and data sources. This document contains comprehensive research on implementing MCP servers in Go, with a focus on creating a web search tool.

## Official Go SDK Status

The official `modelcontextprotocol/go-sdk` is:
- **Status**: Unreleased and currently unstable (as of January 2025)
- **Planned Stable Release**: August 2025
- **Repository**: https://github.com/modelcontextprotocol/go-sdk
- **License**: MIT (copyright "Go SDK Authors")
- **Warning**: Should not be used in production projects yet

### Official SDK Features

The official SDK provides:
1. **Primary API Package (`mcp`)**: Core APIs for MCP clients and servers
2. **JSON Schema Package (`jsonschema`)**: JSON Schema implementation for tool input/output
3. **Consistent API Design**: Following Go conventions like net/http and net/rpc
4. **Type-Safe Tool Creation**: Generic functions for strong typing
5. **Multiple Transport Options**: Stdio, in-memory, logging wrapper

### Official SDK Architecture

#### Server Structure
```go
type Server struct {
    name string
    version string
    opts ServerOptions
    mu sync.Mutex
    prompts *featureSet[*ServerPrompt]
    tools *featureSet[*ServerTool]
    resources *featureSet[*ServerResource]
    resourceTemplates *featureSet[*ServerResourceTemplate]
    sessions []*ServerSession
    sendingMethodHandler_ MethodHandler[*ServerSession]
    receivingMethodHandler_ MethodHandler[*ServerSession]
}
```

#### Tool Registration Pattern
```go
type ServerTool struct {
    Tool *Tool
    Handler ToolHandler
    rawHandler rawToolHandler
    inputResolved, outputResolved *jsonschema.Resolved
}

type ToolHandler func(context.Context, *ServerSession, *CallToolParamsFor[map[string]any]) (*CallToolResult, error)
type ToolHandlerFor[In, Out any] func(context.Context, *ServerSession, *CallToolParamsFor[In]) (*CallToolResultFor[Out], error)
```

### Complete Example with Official SDK (Unstable)
```go
package main

import (
    "context"
    "log"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Define typed parameters for the tool
type HiParams struct {
    Name string `json:"name"`
}

// Tool handler with typed parameters
func SayHi(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[HiParams]) (*mcp.CallToolResultFor[any], error) {
    return &mcp.CallToolResultFor[any]{
        Content: []mcp.Content{&mcp.TextContent{Text: "Hi " + params.Arguments.Name}},
    }, nil
}

func main() {
    // Create a server with name and version
    server := mcp.NewServer("greeter", "v1.0.0", nil)
    
    // Add tools to the server
    server.AddTools(
        mcp.NewServerTool("greet", "say hi", SayHi, 
            mcp.Input(
                mcp.Property("name", 
                    mcp.Description("the name of the person to greet"),
                    mcp.Required(true),
                ),
            ),
        ),
    )
    
    // Create stdio transport
    transport := mcp.NewStdioTransport()
    
    // Run the server
    if err := server.Run(context.Background(), transport); err != nil {
        log.Fatal(err)
    }
}
```

### Key API Methods

#### Server Creation and Configuration
```go
// Create new server
func NewServer(name, version string, opts *ServerOptions) *Server

// Add features
func (s *Server) AddTools(tools ...*ServerTool)
func (s *Server) AddPrompts(prompts ...*ServerPrompt)
func (s *Server) AddResources(resources ...*ServerResource)
func (s *Server) AddResourceTemplates(templates ...*ServerResourceTemplate)

// Connection management
func (s *Server) Connect(ctx context.Context, t Transport) (*ServerSession, error)
func (s *Server) Run(ctx context.Context, t Transport) error
```

#### Transport Options
```go
// Transport interface
type Transport interface {
    Connect(ctx context.Context) (Connection, error)
}

// Available transports
func NewStdioTransport() *StdioTransport                          // stdin/stdout
func NewInMemoryTransports() (*InMemoryTransport, *InMemoryTransport) // for testing
func NewLoggingTransport(delegate Transport, w io.Writer) *LoggingTransport // for debugging
```

### Schema Definition Patterns

The official SDK provides a fluent API for defining tool input schemas:

```go
// Basic property definition
mcp.Property("fieldName", 
    mcp.Description("Field description"),
    mcp.Required(true),
    mcp.Type("string"),
)

// Enum values
mcp.Property("status",
    mcp.Enum("pending", "in_progress", "completed"),
    mcp.Description("Task status"),
)

// Nested objects
mcp.Property("metadata",
    mcp.Type("object"),
    mcp.Properties(
        mcp.Property("priority", mcp.Type("string")),
        mcp.Property("tags", mcp.Type("array")),
    ),
)
```

### Migration Considerations

#### Key Differences from Community SDKs

1. **Type Safety**: Official SDK uses generics for stronger type checking
2. **API Design**: More idiomatic Go patterns following stdlib conventions
3. **Middleware Support**: Built-in support for request/response middleware
4. **Error Handling**: Standardized error types and handling patterns
5. **Feature Management**: Generic `featureSet` for managing tools/resources/prompts

#### Migration Strategy

Since the official SDK is unstable until August 2025, the recommended approach is:

1. **Current Implementation**: Use mark3labs/mcp-go for production
2. **Design for Migration**: Structure code to minimize future changes
3. **Abstraction Layer**: Consider creating an interface layer to ease migration

Example abstraction:
```go
// Define interface matching both SDKs
type MCPServer interface {
    AddTool(name, description string, handler ToolHandler) error
    Run(ctx context.Context) error
}

// Wrapper for current SDK
type Mark3LabsServer struct {
    server *server.MCPServer
}

// Wrapper for future official SDK
type OfficialServer struct {
    server *mcp.Server
}
```

## Community Go MCP Libraries

Since the official SDK is not yet stable, several community libraries are available:

### 1. mark3labs/mcp-go (Recommended for Production)

The most mature and widely used community implementation.

**Key Features:**
- Fast and simple with minimal boilerplate
- Complete implementation of core MCP specification
- Easy tool registration and handling
- Supports stdio, SSE, and HTTP transports
- Active maintenance and community support

**Installation:**
```bash
go get github.com/mark3labs/mcp-go
```

**Example Implementation:**
```go
package main

import (
    "context"
    "fmt"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

func main() {
    s := server.NewMCPServer(
        "Demo Server",
        "1.0.0",
        server.WithToolCapabilities(false),
    )

    tool := mcp.NewTool("hello_world",
        mcp.WithDescription("Say hello to someone"),
        mcp.WithString("name",
            mcp.Required(),
            mcp.Description("Name of the person to greet"),
        ),
    )

    s.AddTool(tool, helloHandler)

    if err := server.ServeStdio(s); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    name, err := request.RequireString("name")
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }
    return mcp.NewToolResultText(fmt.Sprintf("Hello, %s!", name)), nil
}
```

### 2. metoro-io/mcp-golang

An unofficial implementation with type safety and low boilerplate. Excellent documentation at https://mcpgolang.com.

**Key Features:**
- Type-safe tool arguments using native Go structs
- Automatic schema generation from struct tags
- Built-in transports (stdio, HTTP)
- Bi-directional support (server and client)

**Example Implementation:**
```go
package main

import (
    "fmt"
    mcp_golang "github.com/metoro-io/mcp-golang"
    "github.com/metoro-io/mcp-golang/transport/stdio"
)

type SearchArgs struct {
    Query string `json:"query" jsonschema:"required,description=The search query"`
    Limit int    `json:"limit,omitempty" jsonschema:"description=Maximum results to return"`
}

func main() {
    server := mcp_golang.NewServer(stdio.NewStdioServerTransport())
    
    err := server.RegisterTool("google_search", "Search the web", 
        func(args SearchArgs) (*mcp_golang.ToolResponse, error) {
            // Implement search logic here
            result := fmt.Sprintf("Searching for: %s", args.Query)
            return mcp_golang.NewToolResponse(
                mcp_golang.NewTextContent(result),
            ), nil
        })
    
    if err != nil {
        panic(err)
    }
    
    err = server.Serve()
    if err != nil {
        panic(err)
    }
}
```

### 3. cbrgm/go-mcp-server

A learning-focused implementation built from scratch. Not recommended for production but useful for understanding the protocol.

## MCP Protocol Details

### Core Architecture

MCP follows a client-server architecture with four main components:
1. **Host**: Coordinates the system and manages LLM interactions
2. **Clients**: Connect hosts to servers (1:1 relationships)
3. **Servers**: Provide specialized capabilities through tools, resources, and prompts
4. **Base Protocol**: Defines communication using JSON-RPC 2.0

### Tool Implementation

MCP servers expose tools through two main endpoints:
- `tools/list`: Returns all available tools with their schemas
- `tools/call`: Executes a specific tool with provided parameters

### Transport Options

1. **Stdio**: Communication via stdin/stdout (most common)
2. **SSE (Server-Sent Events)**: For web-based implementations
3. **HTTP Streaming**: For HTTP-based communication

## Security Best Practices (2025)

### 1. Environment Variables and Secrets

**Never hardcode credentials:**
```go
// Bad
const API_KEY = "sk_1234567890abcdef"

// Good
apiKey := os.Getenv("GEMINI_API_KEY")
```

### 2. OAuth 2.1 Migration

As of 2025, OAuth 2.1 is becoming mandatory for MCP servers. Consider implementing OAuth flows instead of API keys for better security.

### 3. Configuration Security

When configuring MCP servers in settings.json:
```json
{
  "mcpServers": {
    "gemini-grounding": {
      "command": "go",
      "args": ["run", "./mcp-gemini-grounding"],
      "env": {
        "GEMINI_API_KEY": "$GEMINI_API_KEY",
        "GOOGLE_SEARCH_ENGINE_ID": "$GOOGLE_CSE_ID"
      }
    }
  }
}
```

### 4. Known Issues

- Environment variables from the `env` section in configuration files may not be passed correctly in some implementations
- Store credentials using OS-specific secure storage (Windows Credentials API, macOS Keychain)

## Web Search Implementation Options

### Option 1: Google Custom Search API

**Pros:**
- Direct API access with clear pricing ($5 per 1000 queries)
- 100 free queries per day
- Full control over search parameters
- Well-documented API

**Cons:**
- Requires Google Cloud account and billing setup
- API key management
- Rate limiting considerations

### Option 2: Gemini API Grounding

**Pros:**
- Built into Gemini API (no separate search API needed)
- Returns grounding metadata with sources
- Automatic citation insertion
- More accurate results with reduced hallucinations

**Cons:**
- $35 per 1000 grounded queries (more expensive)
- Requires Gemini API access
- Limited to Gemini models
- Display requirements for search suggestions

**Implementation Example:**
```go
// Using Gemini API for grounding
type GroundingRequest struct {
    Contents []Content `json:"contents"`
    Tools    []Tool    `json:"tools"`
}

type Tool struct {
    GoogleSearch *GoogleSearch `json:"googleSearch,omitempty"`
}

type GoogleSearch struct{}

// Make request with grounding
request := GroundingRequest{
    Contents: []Content{{
        Role: "user",
        Parts: []Part{{Text: query}},
    }},
    Tools: []Tool{{
        GoogleSearch: &GoogleSearch{},
    }},
}
```

### Option 3: Web Scraping

**Pros:**
- No API keys required
- Free to use

**Cons:**
- Against Google's Terms of Service
- Unreliable and prone to breaking
- Rate limiting and blocking risks
- Not recommended for production

## Existing MCP Web Search Implementations

### 1. Google-Search-MCP-Server (TypeScript)
- Uses Google Custom Search API
- Includes webpage content analysis
- Caching mechanisms for performance
- Reference: github.com/mixelpixx/Google-Search-MCP-Server

### 2. WebSearch-MCP
- Alternative implementations using Perplexity, Tavily APIs
- Various approaches to web search integration

## Best Practices for MCP Tool Development

1. **Tool Naming**: Keep names under 63 characters, use only alphanumeric, underscore, dot, hyphen
2. **Schema Validation**: Always validate input parameters
3. **Error Handling**: Return errors in tool results, not as protocol errors
4. **Descriptive Metadata**: Include clear descriptions for tools and parameters
5. **Timeout Management**: Configure appropriate timeouts (default 10 minutes)
6. **Connection Persistence**: Maintain connections for registered tools

## Authentication Integration with Gemini CLI

### Using Gemini CLI Credentials

The MCP server can leverage existing Gemini CLI authentication:

1. **API Key Support**: Read `GEMINI_API_KEY` environment variable
2. **OAuth Credential Sharing**: Read from `~/.gemini/oauth_creds.json`
3. **Credential Format**:
```json
{
  "type": "authorized_user",
  "client_id": "...",
  "client_secret": "...",
  "refresh_token": "...",
  "access_token": "...",
  "expiry": "2025-01-01T00:00:00Z"
}
```

### Implementation Example:
```go
// Check for OAuth credentials first
func loadCredentials() (*oauth2.Token, error) {
    homeDir, _ := os.UserHomeDir()
    credPath := filepath.Join(homeDir, ".gemini", "oauth_creds.json")
    
    data, err := os.ReadFile(credPath)
    if err != nil {
        // Fall back to API key
        return nil, err
    }
    
    var creds OAuthCredentials
    if err := json.Unmarshal(data, &creds); err != nil {
        return nil, err
    }
    
    return &oauth2.Token{
        AccessToken:  creds.AccessToken,
        RefreshToken: creds.RefreshToken,
        Expiry:       creds.Expiry,
    }, nil
}
```

## Resources

- MCP Specification: https://modelcontextprotocol.io/specification/2025-06-18
- Official Go SDK (unstable): https://github.com/modelcontextprotocol/go-sdk
- mark3labs/mcp-go: https://github.com/mark3labs/mcp-go
- mcp-golang docs: https://mcpgolang.com
- MCP Security Guide: https://cloudsecurityalliance.org/blog/2025/06/23/a-primer-on-model-context-protocol-mcp-secure-implementation

## 3. Complete Example: Web Search MCP Server

Now, create a `main.go` file. The following code sets up a simple server with one tool called `hello_world`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// helloHandler is the function that executes when the "hello_world" tool is called.
func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract the "name" argument from the tool request.
	name, err := request.RequireString("name")
	if err != nil {
		// Return an error result if the required argument is missing.
		return mcp.NewToolResultError(err.Error()), nil
	}

	// If successful, return a text result.
	return mcp.NewToolResultText(fmt.Sprintf("Hello, %s!", name)), nil
}

func main() {
	// 1. Create a new MCP server instance.
	s := server.NewMCPServer(
		"HelloWorldServer",
		"1.0.0",
		// You can configure server capabilities, for example, disabling resource capabilities.
		server.WithToolCapabilities(false),
	)

	// 2. Define a tool.
	tool := mcp.NewTool(
		"hello_world",
		mcp.WithDescription("Say hello to someone."),
		// Define the arguments the tool accepts. This one takes a required string named "name".
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("The name of the person to greet."),
		),
	)

	// 3. Add the tool to the server and associate it with its handler function.
	s.AddTool(tool, helloHandler)

	fmt.Println("Starting MCP server over stdio...")

	// 4. Start the server using the stdio transport.
	// This is a common transport for local MCP servers.
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
```

## 4. How It Works

1.  **Server Initialization**: `server.NewMCPServer` creates the core server, where you define its name and version.
2.  **Tool Definition**: `mcp.NewTool` defines a function that an LLM can call. You provide a name, a description (which helps the LLM decide when to use it), and definitions for its arguments. The arguments are converted into a JSON schema that the client uses.
3.  **Handler Function**: The `helloHandler` is the actual implementation of your tool. It receives the tool call request, extracts arguments, performs its logic, and returns a result.
4.  **Serving**: `server.ServeStdio(s)` starts the server, making it listen for requests over standard input and send responses over standard output. The `mcp-go` library handles the complex details of the MCP protocol, letting you focus on your tool's logic.

## 5. Build and Run the Server

To run your server, you first need to build it:

```bash
go build
```

This command will create an executable file (e.g., `mcp-server-example`). You can then run it from your terminal:

```bash
./mcp-server-example
```

The server is now running and waiting for an MCP client (like a compatible IDE or desktop application) to connect to it via its standard I/O streams. When a client connects and asks to use the "hello_world" tool, your server will execute the `helloHandler` function and return the greeting.

## Implementation Recommendations for MCP Todo Server

Given the current state of the official SDK and our project requirements, here are the recommended approaches:

### 1. Use mark3labs/mcp-go for Production (Recommended)

Since the official SDK is unstable until August 2025, we should use the mature mark3labs/mcp-go library:

**Advantages:**
- Production-ready and well-tested
- Active community support
- Simpler API with less boilerplate
- Already proven in production environments

**Implementation Pattern for Todo Tools:**
```go
// Example todo_create handler
func todoCreateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Extract parameters
    task, err := request.RequireString("task")
    if err != nil {
        return mcp.NewToolResultError("task parameter is required"), nil
    }
    
    priority := request.GetStringOrDefault("priority", "high")
    todoType := request.GetStringOrDefault("type", "feature")
    
    // Use existing TodoManager
    todo, err := todoManager.CreateTodo(task, priority, todoType)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Failed to create todo: %v", err)), nil
    }
    
    // Return result
    result := map[string]interface{}{
        "id": todo.ID,
        "path": todo.FilePath,
        "message": fmt.Sprintf("Created todo '%s'", todo.ID),
    }
    
    return mcp.NewToolResultContent(mcp.NewTextContent(
        fmt.Sprintf("Todo created successfully: %s", todo.ID),
    )), nil
}
```

### 2. Design for Future Migration

Structure the code to minimize changes when migrating to the official SDK:

```go
// handlers/mcp_adapter.go
package handlers

// MCPHandler abstracts the handler interface
type MCPHandler interface {
    Register(server interface{}) error
}

// TodoHandlers contains all todo-related handlers
type TodoHandlers struct {
    manager *core.TodoManager
    search  *core.SearchEngine
}

// RegisterMark3Labs registers handlers with mark3labs SDK
func (h *TodoHandlers) RegisterMark3Labs(s *server.MCPServer) error {
    // Register todo_create
    s.AddTool(
        mcp.NewTool("todo_create",
            mcp.WithDescription("Create a new todo with full metadata"),
            mcp.WithString("task", mcp.Required(), mcp.Description("Task description")),
            mcp.WithString("priority", mcp.Description("Task priority")),
            mcp.WithString("type", mcp.Description("Todo type")),
        ),
        h.todoCreateHandler,
    )
    
    // Register other tools...
    return nil
}

// RegisterOfficial registers handlers with official SDK (future)
func (h *TodoHandlers) RegisterOfficial(s *mcp.Server) error {
    // Future implementation for official SDK
    return nil
}
```

### 3. Tool Implementation Guidelines

For the MCP Todo Server, follow these patterns:

#### Tool Naming Convention
- Use snake_case for tool names: `todo_create`, `todo_read`, `todo_update`
- Keep names descriptive but under 63 characters
- Group related functionality: `todo_*`, `template_*`

#### Parameter Validation
```go
// Consistent parameter extraction pattern
func extractTodoParams(request mcp.CallToolRequest) (*TodoParams, error) {
    params := &TodoParams{}
    
    // Required parameters
    task, err := request.RequireString("task")
    if err != nil {
        return nil, fmt.Errorf("missing required parameter 'task'")
    }
    params.Task = task
    
    // Optional parameters with defaults
    params.Priority = request.GetStringOrDefault("priority", "high")
    params.Type = request.GetStringOrDefault("type", "feature")
    
    // Validate enums
    if !isValidPriority(params.Priority) {
        return nil, fmt.Errorf("invalid priority: %s", params.Priority)
    }
    
    return params, nil
}
```

#### Error Handling Pattern
```go
// Consistent error responses
func handleTodoError(err error) *mcp.CallToolResult {
    switch {
    case errors.Is(err, core.ErrTodoNotFound):
        return mcp.NewToolResultError("Todo not found")
    case errors.Is(err, core.ErrInvalidID):
        return mcp.NewToolResultError("Invalid todo ID format")
    default:
        return mcp.NewToolResultError(fmt.Sprintf("Operation failed: %v", err))
    }
}
```

### 4. Configuration for Claude Code

Add to Claude Code's settings.json:
```json
{
  "mcpServers": {
    "mcp-todo-server": {
      "command": "/path/to/mcp-todo-server",
      "args": [],
      "env": {
        "CLAUDE_BASE_PATH": "${CLAUDE_BASE_PATH}"
      }
    }
  }
}
```

### 5. Testing Strategy

Create integration tests that work with both SDKs:
```go
// Test with in-memory transport
func TestTodoTools(t *testing.T) {
    // Create test server
    s := server.NewMCPServer("test-server", "1.0.0")
    
    // Register handlers
    handlers := NewTodoHandlers(testManager, testSearch)
    handlers.RegisterMark3Labs(s)
    
    // Create in-memory transport for testing
    client, server := mcp.NewInMemoryTransports()
    
    // Test tool calls
    // ...
}
```

### Summary

1. **Immediate Action**: Implement using mark3labs/mcp-go
2. **Future-Proof**: Create abstraction layer for easy migration
3. **Maintain Compatibility**: Follow MCP specification strictly
4. **Test Thoroughly**: Ensure all tools work with Claude Code
5. **Document Migration Path**: Keep notes on differences for future SDK switch
