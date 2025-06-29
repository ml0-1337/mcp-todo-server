package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	
	ctxkeys "github.com/user/mcp-todo-server/internal/context"
)

func TestHTTPMiddleware_ExtractsWorkingDirectory(t *testing.T) {
	// Create session manager
	sessionManager := NewSessionManager()
	
	// Create test handler that checks context
	var capturedWorkingDir string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wd, ok := r.Context().Value(ctxkeys.WorkingDirectoryKey).(string); ok {
			capturedWorkingDir = wd
		}
		w.WriteHeader(http.StatusOK)
	})
	
	// Apply middleware
	handler := HTTPMiddleware(sessionManager)(testHandler)
	
	// Create test request with X-Working-Directory header
	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("X-Working-Directory", "/test/project")
	
	// Execute request
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	// Verify working directory was extracted
	if capturedWorkingDir != "/test/project" {
		t.Errorf("Expected working directory '/test/project', got '%s'", capturedWorkingDir)
	}
}

func TestHTTPMiddleware_SessionManagement(t *testing.T) {
	// Create session manager
	sessionManager := NewSessionManager()
	
	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	// Apply middleware
	handler := HTTPMiddleware(sessionManager)(testHandler)
	
	// First request: Create session
	req1 := httptest.NewRequest("POST", "/mcp", nil)
	req1.Header.Set("X-Working-Directory", "/project1")
	req1.Header.Set("Mcp-Session-Id", "session-123")
	
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	
	// Verify session was created
	sessionManager.mu.RLock()
	session, exists := sessionManager.sessions["session-123"]
	sessionManager.mu.RUnlock()
	
	if !exists {
		t.Fatal("Session was not created")
	}
	
	if session.WorkingDirectory != "/project1" {
		t.Errorf("Expected working directory '/project1', got '%s'", session.WorkingDirectory)
	}
	
	// Second request: Use existing session
	var capturedWorkingDir string
	testHandler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wd, ok := r.Context().Value(ctxkeys.WorkingDirectoryKey).(string); ok {
			capturedWorkingDir = wd
		}
		w.WriteHeader(http.StatusOK)
	})
	
	handler2 := HTTPMiddleware(sessionManager)(testHandler2)
	
	req2 := httptest.NewRequest("POST", "/mcp", nil)
	req2.Header.Set("Mcp-Session-Id", "session-123")
	// Note: No X-Working-Directory header this time
	
	rr2 := httptest.NewRecorder()
	handler2.ServeHTTP(rr2, req2)
	
	// Verify working directory was retrieved from session
	if capturedWorkingDir != "/project1" {
		t.Errorf("Expected working directory '/project1' from session, got '%s'", capturedWorkingDir)
	}
}

func TestHTTPMiddleware_SessionCleanup(t *testing.T) {
	// Create session manager
	sessionManager := NewSessionManager()
	
	// Create session
	sessionManager.GetOrCreateSession("session-456", "/project2")
	
	// Verify session exists
	sessionManager.mu.RLock()
	_, exists := sessionManager.sessions["session-456"]
	sessionManager.mu.RUnlock()
	
	if !exists {
		t.Fatal("Session was not created")
	}
	
	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	// Apply middleware
	handler := HTTPMiddleware(sessionManager)(testHandler)
	
	// Send DELETE request
	req := httptest.NewRequest("DELETE", "/mcp", nil)
	req.Header.Set("Mcp-Session-Id", "session-456")
	
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	
	// Verify session was removed
	sessionManager.mu.RLock()
	_, exists = sessionManager.sessions["session-456"]
	sessionManager.mu.RUnlock()
	
	if exists {
		t.Error("Session was not removed after DELETE request")
	}
}