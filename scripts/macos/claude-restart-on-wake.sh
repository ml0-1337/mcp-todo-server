#!/usr/bin/env bash
# Script to restart Claude Code when Mac wakes from sleep
# This works around the connection issue after sleep/resume

set -euo pipefail

# Configuration
CLAUDE_APP="Claude Code"
LOG_FILE="$HOME/.claude/logs/restart-on-wake.log"

# Create log directory if it doesn't exist
mkdir -p "$(dirname "$LOG_FILE")"

# Function to log with timestamp
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

# Function to check if Claude Code is running
is_claude_running() {
    pgrep -f "$CLAUDE_APP" > /dev/null 2>&1
}

# Function to restart Claude Code
restart_claude() {
    log "Restarting Claude Code..."
    
    # Kill existing instance
    if is_claude_running; then
        log "Killing existing Claude Code instance"
        pkill -f "$CLAUDE_APP" || true
        sleep 2
    fi
    
    # Start Claude Code
    # Note: Adjust the path to your Claude Code application
    if [ -d "/Applications/Claude Code.app" ]; then
        log "Starting Claude Code"
        open -a "Claude Code"
    elif [ -d "$HOME/Applications/Claude Code.app" ]; then
        log "Starting Claude Code from user Applications"
        open -a "$HOME/Applications/Claude Code.app"
    else
        log "ERROR: Claude Code.app not found"
        exit 1
    fi
    
    log "Claude Code restarted successfully"
}

# Main logic
case "${1:-}" in
    "wake")
        log "Mac woke from sleep - checking Claude Code"
        if is_claude_running; then
            log "Claude Code is running - restarting to fix connection"
            restart_claude
        else
            log "Claude Code is not running - no action needed"
        fi
        ;;
    "test")
        log "Test mode - forcing restart"
        restart_claude
        ;;
    *)
        echo "Usage: $0 {wake|test}"
        echo "  wake - Called when Mac wakes from sleep"
        echo "  test - Test the restart functionality"
        exit 1
        ;;
esac