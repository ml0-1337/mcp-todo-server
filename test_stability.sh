#!/bin/bash

echo "=== MCP Todo Server Stability Test ==="
echo "Testing server stability improvements..."
echo

# Test 1: Concurrent Requests
echo "Test 1: Handling 50 concurrent requests..."
for i in {1..50}; do
    (
        curl -s -X POST http://localhost:8080/mcp \
            -H "Content-Type: application/json" \
            -H "Mcp-Session-Id: stress-test-$i" \
            -H "X-Working-Directory: /Users/macbook/stress-test-$i" \
            -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' > /dev/null
        
        if [ $? -eq 0 ]; then
            echo -n "."
        else
            echo -n "F"
        fi
    ) &
done
wait
echo " Done!"

# Check transport metrics
echo -e "\nTransport Metrics after concurrent test:"
curl -s http://localhost:8080/debug/transport | jq '.total_connections, .total_requests, .total_errors'

# Test 2: Long-running connections with heartbeats
echo -e "\nTest 2: Long-running connections with heartbeats..."
SESSION_ID="long-running-$(uuidgen | tr '[:upper:]' '[:lower:]')"

# Initialize session
curl -s -X POST http://localhost:8080/mcp \
    -H "Content-Type: application/json" \
    -H "Mcp-Session-Id: $SESSION_ID" \
    -H "X-Working-Directory: /Users/macbook/long-running-test" \
    -d '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}}},"id":1}' > /dev/null

echo "Sending heartbeats every 5 seconds for 30 seconds..."
for i in {1..6}; do
    sleep 5
    curl -s -X POST http://localhost:8080/mcp/heartbeat \
        -H "Mcp-Session-Id: $SESSION_ID" \
        -H "X-Working-Directory: /Users/macbook/long-running-test" \
        -d '{}' > /dev/null
    echo -n "♥"
done
echo " Done!"

# Test 3: Connection cleanup
echo -e "\nTest 3: Testing stale connection cleanup..."
echo "Current active connections:"
curl -s http://localhost:8080/debug/transport | jq '.active_connections'

echo "Waiting for monitor to clean up stale connections (this may take up to 30 seconds)..."
sleep 35

echo "Active connections after cleanup:"
curl -s http://localhost:8080/debug/transport | jq '.active_connections'

# Test 4: Error recovery
echo -e "\nTest 4: Testing error recovery..."
# Send invalid JSON
response=$(curl -s -X POST http://localhost:8080/mcp \
    -H "Content-Type: application/json" \
    -H "Mcp-Session-Id: error-test-json" \
    -H "X-Working-Directory: /Users/macbook/error-test" \
    -d 'invalid json' -w "\n%{http_code}" 2>/dev/null)
echo "$response" | grep -q "400" && echo "✓ Invalid JSON handled correctly (HTTP 400)" || echo "✗ Invalid JSON not handled"

# Send oversized request (>10MB)
echo "Testing oversized request handling..."
response=$(dd if=/dev/zero bs=1M count=11 2>/dev/null | curl -s -X POST http://localhost:8080/mcp \
    -H "Content-Type: application/json" \
    -H "Mcp-Session-Id: error-test-size" \
    -H "X-Working-Directory: /Users/macbook/error-test" \
    --data-binary @- -w "\n%{http_code}" 2>/dev/null | tail -1)
[[ "$response" == "400" || "$response" == "413" ]] && echo "✓ Oversized request handled correctly (HTTP $response)" || echo "✗ Oversized request not handled"

# Final metrics
echo -e "\n=== Final Transport Metrics ==="
curl -s http://localhost:8080/debug/transport | jq '{
    total_connections: .total_connections,
    active_connections: .active_connections,
    total_requests: .total_requests,
    total_errors: .total_errors,
    error_rate: ((.total_errors / .total_requests) * 100)
}'

echo -e "\n=== Server Health Check ==="
curl -s http://localhost:8080/health | jq .

echo -e "\nStability test complete!"