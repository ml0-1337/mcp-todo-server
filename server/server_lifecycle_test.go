package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestServerStartupShutdown tests server creation and cleanup
func TestServerStartupShutdown(t *testing.T) {
	tests := []struct {
		name      string
		transport string
	}{
		{
			name:      "http server creation and cleanup",
			transport: "http",
		},
		{
			name:      "stdio server creation and cleanup",
			transport: "stdio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create unique temp directory for this test to avoid conflicts
			tempDir, err := os.MkdirTemp("", "server-lifecycle-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Set environment variables to use isolated directories
			oldTodoPath := os.Getenv("CLAUDE_TODO_PATH")
			oldTemplatePath := os.Getenv("CLAUDE_TEMPLATE_PATH")
			oldBlevePath := os.Getenv("CLAUDE_BLEVE_PATH")
			
			todosDir := filepath.Join(tempDir, "todos")
			templatesDir := filepath.Join(tempDir, "templates")
			bleveDir := filepath.Join(tempDir, "search_index")
			
			os.Setenv("CLAUDE_TODO_PATH", todosDir)
			os.Setenv("CLAUDE_TEMPLATE_PATH", templatesDir)
			os.Setenv("CLAUDE_BLEVE_PATH", bleveDir)
			
			defer func() {
				os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
				os.Setenv("CLAUDE_TEMPLATE_PATH", oldTemplatePath)
				os.Setenv("CLAUDE_BLEVE_PATH", oldBlevePath)
			}()

			// Create server
			server, err := NewTodoServer(WithTransport(tt.transport))
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			// Verify server is properly initialized
			if server.mcpServer == nil {
				t.Error("mcpServer should not be nil")
			}
			if server.handlers == nil {
				t.Error("handlers should not be nil")
			}
			if server.transport != tt.transport {
				t.Errorf("Expected transport %s, got %s", tt.transport, server.transport)
			}

			// For HTTP transport, verify HTTP components are initialized
			if tt.transport == "http" {
				if server.httpServer == nil {
					t.Error("httpServer should not be nil for HTTP transport")
				}
				if server.httpWrapper == nil {
					t.Error("httpWrapper should not be nil for HTTP transport")
				}
			}

			// Test cleanup
			err = server.Close()
			if err != nil {
				t.Errorf("Failed to close server: %v", err)
			}
		})
	}
}

// TestServerMultipleClose tests calling Close multiple times
func TestServerMultipleClose(t *testing.T) {
	// Create unique temp directory
	tempDir, err := os.MkdirTemp("", "server-multiple-close-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set isolated paths
	oldTodoPath := os.Getenv("CLAUDE_TODO_PATH")
	oldTemplatePath := os.Getenv("CLAUDE_TEMPLATE_PATH")
	oldBlevePath := os.Getenv("CLAUDE_BLEVE_PATH")
	
	os.Setenv("CLAUDE_TODO_PATH", filepath.Join(tempDir, "todos"))
	os.Setenv("CLAUDE_TEMPLATE_PATH", filepath.Join(tempDir, "templates"))
	os.Setenv("CLAUDE_BLEVE_PATH", filepath.Join(tempDir, "search_index"))
	
	defer func() {
		os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
		os.Setenv("CLAUDE_TEMPLATE_PATH", oldTemplatePath)
		os.Setenv("CLAUDE_BLEVE_PATH", oldBlevePath)
	}()

	server, err := NewTodoServer(WithTransport("http"))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// First close should succeed
	err = server.Close()
	if err != nil {
		t.Errorf("First Close() failed: %v", err)
	}

	// Second close should also succeed (idempotent)
	err = server.Close()
	if err != nil {
		t.Errorf("Second Close() failed: %v", err)
	}
}

// TestHTTPServerInitialization tests HTTP server initialization
func TestHTTPServerInitialization(t *testing.T) {
	// Create unique temp directory
	tempDir, err := os.MkdirTemp("", "server-http-init-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set isolated paths
	oldTodoPath := os.Getenv("CLAUDE_TODO_PATH")
	oldTemplatePath := os.Getenv("CLAUDE_TEMPLATE_PATH")
	oldBlevePath := os.Getenv("CLAUDE_BLEVE_PATH")
	
	os.Setenv("CLAUDE_TODO_PATH", filepath.Join(tempDir, "todos"))
	os.Setenv("CLAUDE_TEMPLATE_PATH", filepath.Join(tempDir, "templates"))
	os.Setenv("CLAUDE_BLEVE_PATH", filepath.Join(tempDir, "search_index"))
	
	defer func() {
		os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
		os.Setenv("CLAUDE_TEMPLATE_PATH", oldTemplatePath)
		os.Setenv("CLAUDE_BLEVE_PATH", oldBlevePath)
	}()

	server, err := NewTodoServer(WithTransport("http"))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()

	// Verify HTTP components are initialized
	if server.httpServer == nil {
		t.Fatal("httpServer should not be nil")
	}
	if server.httpWrapper == nil {
		t.Fatal("httpWrapper should not be nil")
	}

	// Verify that we can access transport type
	if server.transport != "http" {
		t.Error("Transport should be 'http'")
	}
}

// TestServerWithInvalidTransport tests server behavior with invalid transport
func TestServerWithInvalidTransport(t *testing.T) {
	// Create unique temp directory
	tempDir, err := os.MkdirTemp("", "server-invalid-transport-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set isolated paths
	oldTodoPath := os.Getenv("CLAUDE_TODO_PATH")
	oldTemplatePath := os.Getenv("CLAUDE_TEMPLATE_PATH")
	oldBlevePath := os.Getenv("CLAUDE_BLEVE_PATH")
	
	os.Setenv("CLAUDE_TODO_PATH", filepath.Join(tempDir, "todos"))
	os.Setenv("CLAUDE_TEMPLATE_PATH", filepath.Join(tempDir, "templates"))
	os.Setenv("CLAUDE_BLEVE_PATH", filepath.Join(tempDir, "search_index"))
	
	defer func() {
		os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
		os.Setenv("CLAUDE_TEMPLATE_PATH", oldTemplatePath)
		os.Setenv("CLAUDE_BLEVE_PATH", oldBlevePath)
	}()

	// Create server with invalid transport
	server, err := NewTodoServer(WithTransport("invalid"))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()

	// The server should be created but fail when starting
	// This tests error handling in the main function
	if server.transport != "invalid" {
		t.Errorf("Expected transport 'invalid', got %s", server.transport)
	}
}

// TestConcurrentStartStop tests concurrent start/stop operations
func TestConcurrentStartStop(t *testing.T) {
	// Create unique temp directory
	tempDir, err := os.MkdirTemp("", "server-concurrent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set isolated paths
	oldTodoPath := os.Getenv("CLAUDE_TODO_PATH")
	oldTemplatePath := os.Getenv("CLAUDE_TEMPLATE_PATH")
	oldBlevePath := os.Getenv("CLAUDE_BLEVE_PATH")
	
	os.Setenv("CLAUDE_TODO_PATH", filepath.Join(tempDir, "todos"))
	os.Setenv("CLAUDE_TEMPLATE_PATH", filepath.Join(tempDir, "templates"))
	os.Setenv("CLAUDE_BLEVE_PATH", filepath.Join(tempDir, "search_index"))
	
	defer func() {
		os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
		os.Setenv("CLAUDE_TEMPLATE_PATH", oldTemplatePath)
		os.Setenv("CLAUDE_BLEVE_PATH", oldBlevePath)
	}()

	// This test simulates multiple goroutines trying to start/stop the server
	server, err := NewTodoServer(WithTransport("http"))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Use different ports for each goroutine
	const numGoroutines = 5
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Try to close the server
			// Only one should succeed, others should handle gracefully
			_ = server.Close()
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// TestServerWithContext tests server creation with context
func TestServerWithContext(t *testing.T) {
	// Create unique temp directory
	tempDir, err := os.MkdirTemp("", "server-context-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set isolated paths
	oldTodoPath := os.Getenv("CLAUDE_TODO_PATH")
	oldTemplatePath := os.Getenv("CLAUDE_TEMPLATE_PATH")
	oldBlevePath := os.Getenv("CLAUDE_BLEVE_PATH")
	
	os.Setenv("CLAUDE_TODO_PATH", filepath.Join(tempDir, "todos"))
	os.Setenv("CLAUDE_TEMPLATE_PATH", filepath.Join(tempDir, "templates"))
	os.Setenv("CLAUDE_BLEVE_PATH", filepath.Join(tempDir, "search_index"))
	
	defer func() {
		os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
		os.Setenv("CLAUDE_TEMPLATE_PATH", oldTemplatePath)
		os.Setenv("CLAUDE_BLEVE_PATH", oldBlevePath)
	}()

	server, err := NewTodoServer(WithTransport("http"))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create a context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Wait for context to expire
	<-ctx.Done()

	// Server should still be closeable after context expiration
	err = server.Close()
	if err != nil {
		t.Errorf("Failed to close server after context expiration: %v", err)
	}
}

// BenchmarkServerStartupShutdown benchmarks server startup and shutdown
func BenchmarkServerStartupShutdown(b *testing.B) {
	// Create a single temp directory for all benchmark iterations
	tempDir, err := os.MkdirTemp("", "server-benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set isolated paths
	oldTodoPath := os.Getenv("CLAUDE_TODO_PATH")
	oldTemplatePath := os.Getenv("CLAUDE_TEMPLATE_PATH")
	oldBlevePath := os.Getenv("CLAUDE_BLEVE_PATH")
	
	defer func() {
		os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
		os.Setenv("CLAUDE_TEMPLATE_PATH", oldTemplatePath)
		os.Setenv("CLAUDE_BLEVE_PATH", oldBlevePath)
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use unique subdirectories for each iteration
		iterDir := filepath.Join(tempDir, fmt.Sprintf("iter-%d", i))
		os.Setenv("CLAUDE_TODO_PATH", filepath.Join(iterDir, "todos"))
		os.Setenv("CLAUDE_TEMPLATE_PATH", filepath.Join(iterDir, "templates"))
		os.Setenv("CLAUDE_BLEVE_PATH", filepath.Join(iterDir, "search_index"))

		server, err := NewTodoServer(WithTransport("http"))
		if err != nil {
			b.Fatalf("Failed to create server: %v", err)
		}

		// Just test creation and close, not actual startup
		// as that would be too slow for benchmarking
		server.Close()
	}
}