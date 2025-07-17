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

func TestSessionManager_MultipleSessionsSameDirectory(t *testing.T) {
	sm := NewSessionManager()
	workingDir := "/test/project"
	
	// Create first session
	session1 := sm.GetOrCreateSession("session-1", workingDir)
	if session1.ID != "session-1" {
		t.Errorf("Expected session ID 'session-1', got '%s'", session1.ID)
	}
	if session1.WorkingDirectory != workingDir {
		t.Errorf("Expected working directory '%s', got '%s'", workingDir, session1.WorkingDirectory)
	}
	
	// Create second session with same working directory
	session2 := sm.GetOrCreateSession("session-2", workingDir)
	if session2.ID != "session-2" {
		t.Errorf("Expected session ID 'session-2', got '%s'", session2.ID)
	}
	if session2.WorkingDirectory != workingDir {
		t.Errorf("Expected working directory '%s', got '%s'", workingDir, session2.WorkingDirectory)
	}
	
	// Verify they are different sessions
	if session1 == session2 {
		t.Error("Expected different session objects for different session IDs")
	}
	
	// Verify both sessions exist
	if sm.GetActiveSessions() != 2 {
		t.Errorf("Expected 2 active sessions, got %d", sm.GetActiveSessions())
	}
	
	// Verify sessions for directory returns both
	sessions := sm.GetSessionsForDirectory(workingDir)
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions for directory, got %d", len(sessions))
	}
}

func TestSessionManager_IndependentSessions(t *testing.T) {
	sm := NewSessionManager()
	
	// Create sessions for different directories
	session1 := sm.GetOrCreateSession("session-1", "/project/a")
	session2 := sm.GetOrCreateSession("session-2", "/project/b")
	
	if session1.WorkingDirectory == session2.WorkingDirectory {
		t.Error("Sessions should have different working directories")
	}
	
	// Remove one session
	sm.RemoveSession("session-1")
	
	// Verify only one session remains
	if sm.GetActiveSessions() != 1 {
		t.Errorf("Expected 1 active session after removal, got %d", sm.GetActiveSessions())
	}
	
	// Verify session-2 still exists
	session2Again := sm.GetOrCreateSession("session-2", "/project/b")
	if session2Again != session2 {
		t.Error("Expected to get existing session-2")
	}
}