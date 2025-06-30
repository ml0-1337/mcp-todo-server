#!/usr/bin/env bash
#
# test_http_headers.sh - Test HTTP header-based working directory isolation
#
# This script verifies that the X-Working-Directory HTTP header correctly
# isolates todo files between different projects/directories.
#
# Usage: ./test_http_headers.sh

set -euo pipefail  # Exit on error, undefined vars, pipe failures

echo "Testing HTTP header-based working directory..."

# Create test directories
TEST_DIR="/tmp/mcp-todo-http-test"
PROJECT1_DIR="$TEST_DIR/project1"
PROJECT2_DIR="$TEST_DIR/project2"

# Clean up previous test
rm -rf $TEST_DIR
mkdir -p $PROJECT1_DIR
mkdir -p $PROJECT2_DIR

echo "Created test directories:"
echo "  - $PROJECT1_DIR"
echo "  - $PROJECT2_DIR"

# Build the server
echo "Building server..."
# Build in test directory
go build -o $TEST_DIR/mcp-todo-server

# Start server in background
echo "Starting MCP Todo Server in HTTP mode..."
$TEST_DIR/mcp-todo-server -transport http -port 8080 &
SERVER_PID=$!

# Give server time to start
sleep 2

# Test 1: Create todo in PROJECT1 using X-Working-Directory header
echo -e "\n=== Test 1: Create todo in PROJECT1 using header ==="
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $PROJECT1_DIR" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "0.1.0",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }'

echo -e "\n\nCreating todo..."
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $PROJECT1_DIR" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "todo_create",
      "arguments": {
        "task": "Test todo in PROJECT1 via HTTP header",
        "priority": "high",
        "type": "feature"
      }
    }
  }'

# Test 2: Create todo in PROJECT2 using different header
echo -e "\n\n=== Test 2: Create todo in PROJECT2 using header ==="
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $PROJECT2_DIR" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "todo_create",
      "arguments": {
        "task": "Test todo in PROJECT2 via HTTP header",
        "priority": "medium",
        "type": "bug"
      }
    }
  }'

# Give time for files to be written
sleep 1

echo -e "\n\n=== Checking results ==="
echo "PROJECT1 todos:"
if ls $PROJECT1_DIR/.claude/todos/*.md 2>/dev/null; then
    echo "✓ Todo created in PROJECT1"
    echo "Contents:"
    head -n 10 $PROJECT1_DIR/.claude/todos/*.md
else
    echo "✗ No todos found in PROJECT1"
fi

echo -e "\nPROJECT2 todos:"
if ls $PROJECT2_DIR/.claude/todos/*.md 2>/dev/null; then
    echo "✓ Todo created in PROJECT2"
    echo "Contents:"
    head -n 10 $PROJECT2_DIR/.claude/todos/*.md
else
    echo "✗ No todos found in PROJECT2"
fi

# Clean up
echo -e "\n\nCleaning up..."
kill $SERVER_PID 2>/dev/null || true

echo -e "\n✓ Test complete!"