#!/bin/bash
# Start MCP server with debug logging

echo "Starting MCP Todo Server with debug logging..."
echo "Watch this terminal to see what X-Working-Directory header Claude Code sends"
echo "="

./build/mcp-todo-server-darwin-arm64 -transport http -port 8080 2>&1 | while read line; do
    echo "$line"
    if [[ "$line" == *"[Header] X-Working-Directory:"* ]]; then
        echo ">>> FOUND HEADER: $line"
    fi
    if [[ "$line" == *"Creating new manager"* ]]; then
        echo ">>> MANAGER CREATION: $line"
    fi
    if [[ "$line" == *"failed to create todo directory"* ]]; then
        echo ">>> ERROR: $line"
    fi
done