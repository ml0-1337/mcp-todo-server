package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/user/mcp-todo-server/server"
)

func main() {
	// Create and start server
	todoServer, err := server.NewTodoServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	
	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Println("Starting MCP Todo Server v1.0.0...")
		if err := todoServer.Start(); err != nil {
			errChan <- err
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