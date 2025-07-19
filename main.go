package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/mcp-todo-server/internal/lock"
	"github.com/user/mcp-todo-server/internal/logging"
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
	logging.Logf("MCP Todo Server starting...")
	
	// Parse command line flags
	var (
		transport        = flag.String("transport", "http", "Transport type: stdio, http (default: http)")
		port             = flag.String("port", "8080", "Port for HTTP transport (default: 8080)")
		host             = flag.String("host", "localhost", "Host for HTTP transport (default: localhost)")
		version          = flag.Bool("version", false, "Print version and exit")
		sessionTimeout   = flag.Duration("session-timeout", 7*24*time.Hour, "Session timeout duration (default: 7d, 0 to disable)")
		managerTimeout   = flag.Duration("manager-timeout", 24*time.Hour, "Manager set timeout duration (default: 24h, 0 to disable)")
		heartbeatInterval = flag.Duration("heartbeat-interval", 30*time.Second, "HTTP heartbeat interval (default: 30s, 0 to disable)")
		noAutoArchive    = flag.Bool("no-auto-archive", false, "Disable automatic archiving when todo status is set to completed")
		requestTimeout   = flag.Duration("request-timeout", 60*time.Second, "HTTP request timeout (default: 60s, 0 to disable)")
		httpReadTimeout  = flag.Duration("http-read-timeout", 120*time.Second, "HTTP server read timeout (default: 120s)")
		httpWriteTimeout = flag.Duration("http-write-timeout", 120*time.Second, "HTTP server write timeout (default: 120s)")
		httpIdleTimeout  = flag.Duration("http-idle-timeout", 120*time.Second, "HTTP server idle timeout (default: 120s)")
	)
	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Printf("MCP Todo Server v%s\n", Version)
		os.Exit(0)
	}

	// Check environment variable for auto-archive override
	if envNoAutoArchive := os.Getenv("CLAUDE_TODO_NO_AUTO_ARCHIVE"); envNoAutoArchive != "" {
		if envNoAutoArchive == "true" || envNoAutoArchive == "1" {
			*noAutoArchive = true
		}
	}

	// For HTTP transport, acquire exclusive lock to prevent multiple instances
	var serverLock *lock.ServerLock
	var err error
	if *transport == "http" {
		serverLock, err = lock.NewServerLock(*port)
		if err != nil {
			log.Fatalf("Failed to create server lock: %v", err)
		}
		
		err = serverLock.TryLock()
		if err != nil {
			log.Fatalf("Failed to acquire server lock: %v", err)
		}
		
		logging.Logf("Acquired exclusive lock for port %s", *port)
	}

	// Create server with transport type and timeout options
	todoServer, err := server.NewTodoServer(
		server.WithTransport(*transport),
		server.WithSessionTimeout(*sessionTimeout),
		server.WithManagerTimeout(*managerTimeout),
		server.WithHeartbeatInterval(*heartbeatInterval),
		server.WithNoAutoArchive(*noAutoArchive),
		server.WithHTTPRequestTimeout(*requestTimeout),
		server.WithHTTPReadTimeout(*httpReadTimeout),
		server.WithHTTPWriteTimeout(*httpWriteTimeout),
		server.WithHTTPIdleTimeout(*httpIdleTimeout),
	)
	if err != nil {
		if serverLock != nil {
			serverLock.Unlock()
		}
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
			logging.Logf("Starting MCP Todo Server v%s (STDIO mode)...", Version)
			logging.Logf("Session timeout: %v, Manager timeout: %v, Heartbeat: %v", *sessionTimeout, *managerTimeout, *heartbeatInterval)
			if err := todoServer.StartStdio(); err != nil {
				errChan <- err
			}
		case "http":
			addr := fmt.Sprintf("%s:%s", *host, *port)
			logging.Logf("Starting MCP Todo Server v%s (HTTP mode) on %s...", Version, addr)
			logging.Logf("Session timeout: %v, Manager timeout: %v, Heartbeat: %v", *sessionTimeout, *managerTimeout, *heartbeatInterval)
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
		if serverLock != nil {
			if unlockErr := serverLock.Unlock(); unlockErr != nil {
				logging.Errorf("Error releasing server lock: %v", unlockErr)
			}
		}
		log.Fatalf("Server error: %v", err)
	case sig := <-sigChan:
		logging.Logf("Received signal %v, shutting down...", sig)
		if err := todoServer.Close(); err != nil {
			logging.Errorf("Error closing server: %v", err)
		}
		if serverLock != nil {
			if err := serverLock.Unlock(); err != nil {
				logging.Errorf("Error releasing server lock: %v", err)
			} else {
				logging.Logf("Released server lock")
			}
		}
		os.Exit(0)
	}
}
