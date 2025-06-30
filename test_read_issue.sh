#!/bin/bash

# Test to verify todo_read issue with context-aware directories

set -e

echo "Testing todo_read context awareness issue..."

# Test directory
TEST_DIR="/tmp/mcp-read-test"
rm -rf $TEST_DIR
mkdir -p $TEST_DIR

echo "Test directory: $TEST_DIR"

# Initialize session
echo -e "\n1. Initializing session..."
INIT_RESPONSE=$(curl -s -i -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}')

# Extract session ID
SESSION_ID=$(echo "$INIT_RESPONSE" | grep -i "Mcp-Session-Id:" | cut -d' ' -f2 | tr -d '\r')
echo "Session ID: $SESSION_ID"

# Create a todo in test directory
echo -e "\n2. Creating todo in test directory..."
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "todo_create",
      "arguments": {
        "task": "Test todo for read verification",
        "priority": "high",
        "type": "feature"
      }
    }
  }')

echo "Create response:"
echo "$CREATE_RESPONSE" | jq -r '.result.content[0].text' | jq .

# Extract the created todo ID
TODO_ID=$(echo "$CREATE_RESPONSE" | jq -r '.result.content[0].text' | jq -r '.id')
echo "Created todo ID: $TODO_ID"

# List files in test directory
echo -e "\n3. Files in test directory:"
ls -la $TEST_DIR/.claude/todos/ 2>/dev/null || echo "No todos directory found"

# Now try to read todos with todo_read
echo -e "\n4. Reading todos with todo_read (list all)..."
READ_RESPONSE=$(curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-Working-Directory: $TEST_DIR" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "todo_read",
      "arguments": {
        "format": "list"
      }
    }
  }')

echo "Read response:"
echo "$READ_RESPONSE" | jq .

# Parse the response
if echo "$READ_RESPONSE" | jq -e '.result.content[0].text' > /dev/null 2>&1; then
    TODOS_TEXT=$(echo "$READ_RESPONSE" | jq -r '.result.content[0].text')
    echo -e "\nTodos found:"
    echo "$TODOS_TEXT"
    
    # Check if our test todo is in the list
    if echo "$TODOS_TEXT" | grep -q "$TODO_ID"; then
        echo -e "\n✓ SUCCESS: Found our test todo in the list!"
    else
        echo -e "\n✗ ISSUE CONFIRMED: Our test todo is NOT in the list!"
        echo "This means todo_read is reading from the server directory, not the test directory."
        
        # List server directory todos
        echo -e "\nServer directory todos:"
        ls -la .claude/todos/*.md 2>/dev/null | tail -5 || echo "No todos found"
    fi
else
    echo -e "\n✗ ERROR: Could not parse todo_read response"
fi

echo -e "\nTest complete!"