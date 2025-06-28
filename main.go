package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/user/mcp-todo-server/server"
)

func main() {
	// Parse command line flags
	var (
		transport = flag.String("transport", "http", "Transport type: stdio, http (default: http)")
		port      = flag.String("port", "8080", "Port for HTTP transport (default: 8080)")
		host      = flag.String("host", "localhost", "Host for HTTP transport (default: localhost)")
	)
	flag.Parse()
	
	// Create server with transport type
	todoServer, err := server.NewTodoServer(server.WithTransport(*transport))
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
			log.Println("Starting MCP Todo Server v1.0.0 (STDIO mode)...")
			if err := todoServer.StartStdio(); err != nil {
				errChan <- err
			}
		case "http":
			addr := fmt.Sprintf("%s:%s", *host, *port)
			log.Printf("Starting MCP Todo Server v1.0.0 (HTTP mode) on %s...", addr)
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