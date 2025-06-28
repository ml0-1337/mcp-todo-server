#!/bin/bash

# Quick setup script for MCP Todo Server

echo "🚀 MCP Todo Server Setup"
echo "========================"
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Build the server
echo "📦 Building server..."
if go build -o mcp-todo-server; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi

# Ask user for transport preference
echo
echo "Select transport mode:"
echo "1) HTTP (recommended - supports multiple instances)"
echo "2) STDIO (legacy - single instance only)"
echo
read -p "Enter choice (1 or 2): " choice

case $choice in
    1)
        echo
        echo "🌐 Setting up HTTP transport..."
        
        # Check if port 8080 is available
        if lsof -i :8080 > /dev/null 2>&1; then
            echo "⚠️  Port 8080 is already in use"
            read -p "Enter alternative port: " port
        else
            port=8080
        fi
        
        echo
        echo "To complete setup:"
        echo
        echo "1. Start the server:"
        echo "   ./mcp-todo-server -transport http -port $port"
        echo
        echo "2. Add to Claude Code:"
        echo "   claude mcp add --transport http todo http://localhost:$port/mcp"
        echo
        echo "3. Restart Claude Code"
        ;;
        
    2)
        echo
        echo "📡 Setting up STDIO transport..."
        
        # Get absolute path
        SERVER_PATH=$(pwd)/mcp-todo-server
        
        echo
        echo "To complete setup:"
        echo
        echo "1. Add to Claude Code:"
        echo "   claude mcp add todo $SERVER_PATH --args \"-transport\" \"stdio\""
        echo
        echo "2. Restart Claude Code"
        echo
        echo "Note: The server will start automatically when Claude Code connects"
        ;;
        
    *)
        echo "❌ Invalid choice"
        exit 1
        ;;
esac

echo
echo "📚 For more information, see:"
echo "   - README.md"
echo "   - TRANSPORT_GUIDE.md"
echo
echo "✅ Setup complete!"