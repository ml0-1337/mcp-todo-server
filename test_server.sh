#!/bin/bash

# Test script for MCP Todo Server

echo "Testing MCP Todo Server..."

# Test that the server can be started and responds to initialization
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0.0","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | ./mcp-todo-server | head -20

echo -e "\n\nServer test complete!"