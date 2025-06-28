# MCP Todo Server

An enhanced Model Context Protocol (MCP) server for todo management, built in Go using the mark3labs/mcp-go library.

## Features

- **9 Specialized Todo Tools**: Create, read, update, search, archive, template, link, stats, and clean
- **Full-Text Search**: Powered by Bleve v2 for fast, indexed searching
- **Daily Archive Structure**: Organized YYYY/MM/DD archive folders
- **Template System**: Pre-configured templates for common todo types
- **Todo Linking**: Parent-child relationships for multi-phase projects
- **Statistics Engine**: Comprehensive analytics and reporting
- **Atomic Operations**: Safe concurrent access with proper locking

## Building

```bash
go build -o mcp-todo-server .
```

## Running

The server runs over stdio and implements the MCP protocol:

```bash
./mcp-todo-server
```

## Configuration

The server expects the following directory structure:

```
.claude/
├── todos/        # Active todos
├── archive/      # Archived todos (YYYY/MM/DD)
├── templates/    # Todo templates
└── index/        # Search index
```

## Available Tools

1. **todo_create** - Create a new todo with metadata
2. **todo_read** - Read single todo or list with filters
3. **todo_update** - Update todo content or metadata
4. **todo_search** - Full-text search across todos
5. **todo_archive** - Archive completed todos
6. **todo_template** - Create from templates
7. **todo_link** - Link related todos
8. **todo_stats** - Generate statistics
9. **todo_clean** - Cleanup operations

## Integration with Claude

To use with Claude Code, add to your MCP configuration:

```json
{
  "mcpServers": {
    "todo": {
      "command": "/path/to/mcp-todo-server"
    }
  }
}
```

## Testing

Run all tests:

```bash
go test ./... -v
```

## Implementation Status

✅ Core functionality (all 23 tests passing)
✅ MCP handlers implemented with mark3labs/mcp-go
✅ Server builds and runs
✅ All 9 tools registered with proper schemas

## Future Migration

When the official modelcontextprotocol/go-sdk becomes stable (August 2025), migration will involve:
1. Updating imports from mark3labs to official SDK
2. Adjusting tool registration API calls
3. Updating parameter extraction methods

The core business logic will remain unchanged.