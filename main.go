package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/mcp-todo-server/server"
)

const Version = "2.1.0"

func main() {
	// Check if running in STDIO mode before ANY output
	args := os.Args[1:]
	isStdio := false
	for i, arg := range args {
		if arg == "-transport" && i+1 < len(args) && args[i+1] == "stdio" {
			isStdio = true
			break
		}
		if arg == "-transport=stdio" {
			isStdio = true
			break
		}
	}
	
	// Redirect ALL logging to stderr immediately if STDIO mode
	if isStdio {
		log.SetOutput(os.Stderr)
	}
	
	// Now safe to log
	fmt.Fprintf(os.Stderr, "MCP Todo Server starting...\n")
	
	// Parse command line flags
	var (
		transport        = flag.String("transport", "http", "Transport type: stdio, http (default: http)")
		port             = flag.String("port", "8080", "Port for HTTP transport (default: 8080)")
		host             = flag.String("host", "localhost", "Host for HTTP transport (default: localhost)")
		version          = flag.Bool("version", false, "Print version and exit")
		sessionTimeout   = flag.Duration("session-timeout", 7*24*time.Hour, "Session timeout duration (default: 7d, 0 to disable)")
		managerTimeout   = flag.Duration("manager-timeout", 24*time.Hour, "Manager set timeout duration (default: 24h, 0 to disable)")
		heartbeatInterval = flag.Duration("heartbeat-interval", 30*time.Second, "HTTP heartbeat interval (default: 30s, 0 to disable)")
	)
	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Printf("MCP Todo Server v%s\n", Version)
		os.Exit(0)
	}

	// Create server with transport type and timeout options
	todoServer, err := server.NewTodoServer(
		server.WithTransport(*transport),
		server.WithSessionTimeout(*sessionTimeout),
		server.WithManagerTimeout(*managerTimeout),
		server.WithHeartbeatInterval(*heartbeatInterval),
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		switch *transport {
		case "stdio":
			// Log to stderr in STDIO mode
			fmt.Fprintf(os.Stderr, "Starting MCP Todo Server v%s (STDIO mode)...\n", Version)
			fmt.Fprintf(os.Stderr, "Session timeout: %v, Manager timeout: %v, Heartbeat: %v\n", *sessionTimeout, *managerTimeout, *heartbeatInterval)
			if err := todoServer.StartStdio(); err != nil {
				errChan <- err
			}
		case "http":
			addr := fmt.Sprintf("%s:%s", *host, *port)
			log.Printf("Starting MCP Todo Server v%s (HTTP mode) on %s...", Version, addr)
			log.Printf("Session timeout: %v, Manager timeout: %v, Heartbeat: %v", *sessionTimeout, *managerTimeout, *heartbeatInterval)
			if err := todoServer.StartHTTP(addr); err != nil {
				errChan <- err
			}
		default:
			errChan <- fmt.Errorf("unsupported transport: %s", *transport)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-errChan:
		log.Fatalf("Server error: %v", err)
	case sig := <-sigChan:
		fmt.Fprintf(os.Stderr, "\nReceived signal %v, shutting down...\n", sig)
		if err := todoServer.Close(); err != nil {
			log.Printf("Error closing server: %v", err)
		}
		os.Exit(0)
	}
}
