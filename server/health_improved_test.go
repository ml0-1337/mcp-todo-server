package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	
	"github.com/mark3labs/mcp-go/server"
	"github.com/user/mcp-todo-server/handlers"
)

// createTestServer creates a server with temporary directories for testing
func createTestServer(t *testing.T, transport string) (*TodoServer, func()) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "mcp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	// Create todo and template directories
	todoPath := filepath.Join(tempDir, ".claude", "todos")
	templatePath := filepath.Join(tempDir, ".claude", "templates")
	
	if err := os.MkdirAll(todoPath, 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create todo dir: %v", err)
	}
	
	if err := os.MkdirAll(templatePath, 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create template dir: %v", err)
	}
	
	// Set environment variable to use our temp directory
	oldTodoPath := os.Getenv("CLAUDE_TODO_PATH")
	os.Setenv("CLAUDE_TODO_PATH", todoPath)
	
	// Create handlers directly with our paths to bypass path resolution
	todoHandlers, err := handlers.NewTodoHandlers(todoPath, templatePath, 24*time.Hour, false)
	if err != nil {
		os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create handlers: %v", err)
	}
	
	// Create MCP server
	s := server.NewMCPServer(
		"MCP Todo Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	
	// Create todo server
	ts := &TodoServer{
		mcpServer:         s,
		handlers:          todoHandlers,
		transport:         transport,
		startTime:         time.Now(),
		sessionTimeout:    7 * 24 * time.Hour,
		managerTimeout:    24 * time.Hour,
		heartbeatInterval: 30 * time.Second,
		requestTimeout:    30 * time.Second,
		httpReadTimeout:   60 * time.Second,
		httpWriteTimeout:  60 * time.Second,
		httpIdleTimeout:   120 * time.Second,
	}
	
	// Register tools
	ts.registerTools()
	
	// Create HTTP components if needed
	if transport == "http" {
		ts.httpServer = server.NewStreamableHTTPServer(s)
		ts.stableTransport = NewStableHTTPTransport(ts.httpServer)
		ts.httpWrapper = NewStreamableHTTPServerWrapper(ts.stableTransport, ts.sessionTimeout)
	}
	
	// Cleanup function
	cleanup := func() {
		// Close server first
		ts.Close()
		
		// Restore environment
		os.Setenv("CLAUDE_TODO_PATH", oldTodoPath)
		
		// Remove temp directory
		os.RemoveAll(tempDir)
	}
	
	return ts, cleanup
}

func TestHealthCheckEndpointImproved(t *testing.T) {
	// Create a test server with cleanup
	ts, cleanup := createTestServer(t, "http")
	defer cleanup()
	
	// Test cases
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "GET health check returns 200",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				// Check required fields
				if status, ok := body["status"].(string); !ok || status != "healthy" {
					t.Errorf("Expected status 'healthy', got %v", body["status"])
				}
				
				if _, ok := body["uptime"].(string); !ok {
					t.Error("Missing uptime field")
				}
				
				if _, ok := body["uptimeMs"].(float64); !ok {
					t.Error("Missing uptimeMs field")
				}
				
				if _, ok := body["serverTime"].(string); !ok {
					t.Error("Missing serverTime field")
				}
				
				if transport, ok := body["transport"].(string); !ok || transport != "http" {
					t.Errorf("Expected transport 'http', got %v", body["transport"])
				}
				
				if version, ok := body["version"].(string); !ok || version == "" {
					t.Errorf("Expected non-empty version, got %v", body["version"])
				}
				
				if sessions, ok := body["sessions"].(float64); !ok || sessions < 0 {
					t.Errorf("Expected sessions >= 0, got %v", body["sessions"])
				}
			},
		},
		{
			name:           "POST health check also works",
			method:         "POST",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if status, ok := body["status"].(string); !ok || status != "healthy" {
					t.Errorf("Expected status 'healthy', got %v", body["status"])
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()
			
			// Call handler
			ts.handleHealthCheck(w, req)
			
			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			// Check content type
			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("Expected Content-Type to contain 'application/json', got %s", contentType)
			}
			
			// Parse response
			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			
			// Run custom checks
			tt.checkResponse(t, body)
		})
	}
}

func TestHealthCheckUptimeImproved(t *testing.T) {
	// Create a test server with cleanup
	ts, cleanup := createTestServer(t, "http")
	defer cleanup()
	
	// Set start time to a known value
	ts.startTime = time.Now().Add(-5 * time.Minute)
	
	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	ts.handleHealthCheck(w, req)
	
	// Parse response
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	// Check uptime is approximately 5 minutes
	uptimeMs, ok := body["uptimeMs"].(float64)
	if !ok {
		t.Fatal("Missing uptimeMs field")
	}
	
	expectedMs := float64(5 * 60 * 1000) // 5 minutes in milliseconds
	tolerance := float64(1000)            // 1 second tolerance
	
	if uptimeMs < expectedMs-tolerance || uptimeMs > expectedMs+tolerance {
		t.Errorf("Expected uptime around %fms, got %fms", expectedMs, uptimeMs)
	}
}

func TestHealthCheckWithSessionsImproved(t *testing.T) {
	// Create a test server with cleanup
	ts, cleanup := createTestServer(t, "http")
	defer cleanup()
	
	// Simulate some sessions
	if ts.httpWrapper != nil && ts.httpWrapper.sessionManager != nil {
		ts.httpWrapper.sessionManager.GetOrCreateSession("session1", "/path1")
		ts.httpWrapper.sessionManager.GetOrCreateSession("session2", "/path2")
		ts.httpWrapper.sessionManager.GetOrCreateSession("session3", "/path3")
	}
	
	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	ts.handleHealthCheck(w, req)
	
	// Parse response
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	// Check session count
	sessions, ok := body["sessions"].(float64)
	if !ok {
		t.Fatal("Missing sessions field")
	}
	
	if sessions != 3 {
		t.Errorf("Expected 3 sessions, got %v", sessions)
	}
}

// TestHealthCheckConcurrent tests concurrent health check requests
func TestHealthCheckConcurrent(t *testing.T) {
	// Create a test server with cleanup
	ts, cleanup := createTestServer(t, "http")
	defer cleanup()
	
	// Number of concurrent requests
	numRequests := 10
	done := make(chan bool, numRequests)
	
	// Send concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			
			ts.handleHealthCheck(w, req)
			
			if w.Code != http.StatusOK {
				t.Errorf("Request %d: Expected status 200, got %d", id, w.Code)
			}
			
			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Errorf("Request %d: Failed to parse response: %v", id, err)
			}
			
			if status, ok := body["status"].(string); !ok || status != "healthy" {
				t.Errorf("Request %d: Expected status 'healthy', got %v", id, body["status"])
			}
		}(i)
	}
	
	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		select {
		case <-done:
			// Good
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent requests timed out")
		}
	}
}

// TestServerCloseNoHang ensures server closes without hanging
func TestServerCloseNoHang(t *testing.T) {
	// Create a test server with cleanup
	ts, cleanup := createTestServer(t, "http")
	defer cleanup()
	
	// Close server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	done := make(chan bool, 1)
	go func() {
		ts.Close()
		done <- true
	}()
	
	select {
	case <-done:
		// Good, server closed
	case <-ctx.Done():
		t.Fatal("Server close timed out")
	}
}