#!/usr/bin/env bash
# Install the Claude restart-on-wake service

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLIST_FILE="$SCRIPT_DIR/com.claude.restart-on-wake.plist"
LAUNCH_AGENTS_DIR="$HOME/Library/LaunchAgents"
INSTALLED_PLIST="$LAUNCH_AGENTS_DIR/com.claude.restart-on-wake.plist"

echo -e "${YELLOW}Installing Claude restart-on-wake service...${NC}"

# Check if plist exists
if [ ! -f "$PLIST_FILE" ]; then
    echo -e "${RED}Error: plist file not found at $PLIST_FILE${NC}"
    exit 1
fi

# Create LaunchAgents directory if it doesn't exist
mkdir -p "$LAUNCH_AGENTS_DIR"

# Unload existing service if present
if [ -f "$INSTALLED_PLIST" ]; then
    echo "Unloading existing service..."
    launchctl unload "$INSTALLED_PLIST" 2>/dev/null || true
fi

# Copy plist file
echo "Installing plist file..."
cp "$PLIST_FILE" "$INSTALLED_PLIST"

# Update the path in the plist to use absolute path
sed -i '' "s|/Users/macbook/Programming/go_projects/mcp-todo-server|$SCRIPT_DIR/../..|g" "$INSTALLED_PLIST"

# Load the service
echo "Loading service..."
launchctl load "$INSTALLED_PLIST"

# Verify it's loaded
if launchctl list | grep -q "com.claude.restart-on-wake"; then
    echo -e "${GREEN}✓ Service installed and loaded successfully${NC}"
    echo
    echo "The service will automatically restart Claude Code when your Mac wakes from sleep."
    echo
    echo "To test the service manually:"
    echo "  $SCRIPT_DIR/claude-restart-on-wake.sh test"
    echo
    echo "To uninstall the service:"
    echo "  launchctl unload $INSTALLED_PLIST"
    echo "  rm $INSTALLED_PLIST"
    echo
    echo "Logs are stored at:"
    echo "  ~/.claude/logs/restart-on-wake.log"
else
    echo -e "${RED}✗ Failed to load service${NC}"
    exit 1
fi