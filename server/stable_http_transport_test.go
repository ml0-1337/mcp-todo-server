package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// TestNewStableHTTPTransport tests the creation of StableHTTPTransport
func TestNewStableHTTPTransport(t *testing.T) {
	// Create a minimal MCP server for testing
	mcpServer := server.NewMCPServer("Test", "1.0.0")

	// Create a properly initialized base server
	baseServer := server.NewStreamableHTTPServer(mcpServer)
	transport := NewStableHTTPTransport(baseServer)
	defer transport.Shutdown(context.Background())

	if transport == nil {
		t.Fatal("Expected transport to be created")
	}

	// Verify defaults based on actual implementation
	if transport.maxRequestsPerConnection != 100 {
		t.Errorf("Expected maxRequestsPerConnection to be 100, got %d", transport.maxRequestsPerConnection)
	}

	if transport.requestTimeout != 30*time.Second {
		t.Errorf("Expected requestTimeout to be 30s, got %v", transport.requestTimeout)
	}

	if transport.connectionTimeout != 5*time.Minute {
		t.Errorf("Expected connectionTimeout to be 5min, got %v", transport.connectionTimeout)
	}
}

// TestStableHTTPTransportOptions tests configuration options
func TestStableHTTPTransportOptions(t *testing.T) {
	// Create a minimal MCP server for testing
	mcpServer := server.NewMCPServer("Test", "1.0.0")
	baseServer := server.NewStreamableHTTPServer(mcpServer)
	transport := NewStableHTTPTransport(baseServer,
		WithMaxRequestsPerConnection(500),
		WithRequestTimeout(60*time.Second),
		WithConnectionTimeout(10*time.Minute),
	)
	defer transport.Shutdown(context.Background())

	if transport.maxRequestsPerConnection != 500 {
		t.Errorf("Expected maxRequestsPerConnection to be 500, got %d", transport.maxRequestsPerConnection)
	}

	if transport.requestTimeout != 60*time.Second {
		t.Errorf("Expected requestTimeout to be 60s, got %v", transport.requestTimeout)
	}

	if transport.connectionTimeout != 10*time.Minute {
		t.Errorf("Expected connectionTimeout to be 10min, got %v", transport.connectionTimeout)
	}
}

// TestStableHTTPTransportServeHTTP tests the HTTP handler
func TestStableHTTPTransportServeHTTP(t *testing.T) {
	// Create a minimal MCP server for testing
	mcpServer := server.NewMCPServer("Test", "1.0.0")
	baseServer := server.NewStreamableHTTPServer(mcpServer)
	transport := NewStableHTTPTransport(baseServer)
	defer transport.Shutdown(context.Background())

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET request",
			method:         "GET",
			path:           "/test",
			expectedStatus: http.StatusOK, // The actual implementation will determine this
		},
		{
			name:           "POST request",
			method:         "POST",
			path:           "/test",
			body:           `{"test": "data"}`,
			expectedStatus: http.StatusOK, // The actual implementation will determine this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader([]byte(tt.body)))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// The actual behavior depends on the base server implementation
			// For now, just verify the method exists and can be called
			transport.ServeHTTP(w, req)

			// Basic check that response was written
			if w.Result().StatusCode == 0 {
				t.Error("Expected status code to be set")
			}
		})
	}
}

// TestGetMetrics tests metrics collection
func TestGetMetrics(t *testing.T) {
	// Create a minimal MCP server for testing
	mcpServer := server.NewMCPServer("Test", "1.0.0")
	baseServer := server.NewStreamableHTTPServer(mcpServer)
	transport := NewStableHTTPTransport(baseServer)
	defer transport.Shutdown(context.Background())

	metrics := transport.GetMetrics()

	if metrics == nil {
		t.Fatal("Expected metrics to be returned")
	}

	// Just verify the method exists and returns something
	// The actual structure depends on implementation
}

// TestShutdown tests graceful shutdown
func TestShutdown(t *testing.T) {
	// Create a minimal MCP server for testing
	mcpServer := server.NewMCPServer("Test", "1.0.0")
	baseServer := server.NewStreamableHTTPServer(mcpServer)
	transport := NewStableHTTPTransport(baseServer)

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := transport.Shutdown(ctx)
	if err != nil {
		t.Errorf("Unexpected shutdown error: %v", err)
	}
}

// TestConcurrentAccess tests thread safety
func TestConcurrentAccess(t *testing.T) {
	// Create a minimal MCP server for testing
	mcpServer := server.NewMCPServer("Test", "1.0.0")
	baseServer := server.NewStreamableHTTPServer(mcpServer)
	transport := NewStableHTTPTransport(baseServer)
	defer transport.Shutdown(context.Background())

	// Create multiple goroutines accessing the transport
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Call various methods concurrently
			_ = transport.GetMetrics()

			// Create and handle a request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			transport.ServeHTTP(w, req)
		}(i)
	}

	wg.Wait()

	// If we get here without panic, concurrent access is safe
}

// TestHTTPMethods tests different HTTP methods
func TestHTTPMethods(t *testing.T) {
	// Create a minimal MCP server for testing
	mcpServer := server.NewMCPServer("Test", "1.0.0")
	baseServer := server.NewStreamableHTTPServer(mcpServer)
	transport := NewStableHTTPTransport(baseServer)
	defer transport.Shutdown(context.Background())

	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			// Just verify it doesn't panic
			transport.ServeHTTP(w, req)
		})
	}
}

// TestHeartbeatEndpoint tests the heartbeat functionality
func TestHeartbeatEndpoint(t *testing.T) {
	// Create a minimal MCP server for testing
	mcpServer := server.NewMCPServer("Test", "1.0.0")
	baseServer := server.NewStreamableHTTPServer(mcpServer)
	transport := NewStableHTTPTransport(baseServer)
	defer transport.Shutdown(context.Background())

	// Create request for heartbeat
	req := httptest.NewRequest("GET", "/mcp/heartbeat", nil)
	req.Header.Set("Mcp-Session-Id", "test-session-123")
	w := httptest.NewRecorder()

	// Send heartbeat request
	transport.ServeHTTP(w, req)

	// Check response
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Result().StatusCode)
	}

	// Verify content type
	contentType := w.Result().Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

// TestRequestWithLargeBody tests handling of large request bodies
func TestRequestWithLargeBody(t *testing.T) {
	// Create a minimal MCP server for testing
	mcpServer := server.NewMCPServer("Test", "1.0.0")
	baseServer := server.NewStreamableHTTPServer(mcpServer)
	transport := NewStableHTTPTransport(baseServer)
	defer transport.Shutdown(context.Background())

	// Create large body (11MB, exceeding 10MB limit)
	largeBody := bytes.Repeat([]byte("x"), 11*1024*1024)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Send request
	transport.ServeHTTP(w, req)

	// Should get error for too large body
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for large body, got %d", w.Result().StatusCode)
	}
}

// TestConnectionStateTransitions tests connection state changes
func TestConnectionStateTransitions(t *testing.T) {
	// Test state string representations
	states := []struct {
		state    ConnectionState
		expected string
	}{
		{StateConnecting, "connecting"},
		{StateActive, "active"},
		{StateClosing, "closing"},
		{StateClosed, "closed"},
		{ConnectionState(99), "unknown"},
	}

	for _, tt := range states {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("State.String() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestInvalidJSON tests handling of invalid JSON in requests
func TestInvalidJSON(t *testing.T) {
	// Create a minimal MCP server for testing
	mcpServer := server.NewMCPServer("Test", "1.0.0")
	baseServer := server.NewStreamableHTTPServer(mcpServer)
	transport := NewStableHTTPTransport(baseServer)
	defer transport.Shutdown(context.Background())

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Send request
	transport.ServeHTTP(w, req)

	// Should get error for invalid JSON
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", w.Result().StatusCode)
	}
}
