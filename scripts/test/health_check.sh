#!/usr/bin/env bash
# Test the health check endpoint

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default server URL
SERVER_URL="${1:-http://localhost:8080}"

echo -e "${YELLOW}Testing health check endpoint at ${SERVER_URL}/health${NC}"
echo

# Make the request
response=$(curl -s -w "\n%{http_code}" "${SERVER_URL}/health" 2>/dev/null || echo "CURL_ERROR")

# Check if curl failed
if [[ "$response" == "CURL_ERROR" ]]; then
    echo -e "${RED}✗ Failed to connect to server${NC}"
    echo "  Make sure the server is running with: ./mcp-todo-server -transport http"
    exit 1
fi

# Extract body and status code
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

# Check status code
if [[ "$http_code" == "200" ]]; then
    echo -e "${GREEN}✓ Health check returned 200 OK${NC}"
else
    echo -e "${RED}✗ Health check returned $http_code${NC}"
    exit 1
fi

# Parse JSON and display
echo -e "${YELLOW}Response:${NC}"
echo "$body" | jq . 2>/dev/null || echo "$body"

# Check if server is healthy
status=$(echo "$body" | jq -r '.status' 2>/dev/null || echo "")
if [[ "$status" == "healthy" ]]; then
    echo
    echo -e "${GREEN}✓ Server is healthy${NC}"
    
    # Display key metrics
    uptime=$(echo "$body" | jq -r '.uptime' 2>/dev/null || echo "unknown")
    version=$(echo "$body" | jq -r '.version' 2>/dev/null || echo "unknown")
    sessions=$(echo "$body" | jq -r '.sessions' 2>/dev/null || echo "0")
    
    echo "  Version: $version"
    echo "  Uptime: $uptime"
    echo "  Active sessions: $sessions"
else
    echo
    echo -e "${RED}✗ Server status is not healthy${NC}"
    exit 1
fi

echo
echo -e "${GREEN}Health check passed!${NC}"