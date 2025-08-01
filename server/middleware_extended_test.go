package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ctxkeys "github.com/user/mcp-todo-server/internal/context"
)

func TestSessionManager_CleanupStaleSessions(t *testing.T) {
	sm := NewSessionManager()

	// Create sessions with different last activity times
	now := time.Now()

	// Active session
	activeSession := sm.GetOrCreateSession("active", "/active")
	activeSession.LastActivity = now

	// Stale session (older than timeout)
	staleSession := sm.GetOrCreateSession("stale", "/stale")
	staleSession.LastActivity = now.Add(-2 * time.Hour)

	// Another stale session
	staleSession2 := sm.GetOrCreateSession("stale2", "/stale2")
	staleSession2.LastActivity = now.Add(-3 * time.Hour)

	// Verify all sessions exist
	if sm.GetActiveSessions() != 3 {
		t.Errorf("Expected 3 sessions, got %d", sm.GetActiveSessions())
	}

	// Clean up stale sessions (with 1 hour timeout)
	removed := sm.CleanupStaleSessions(1 * time.Hour)

	// Should have removed 2 stale sessions
	if removed != 2 {
		t.Errorf("Expected 2 sessions removed, got %d", removed)
	}

	// Verify only active session remains
	if sm.GetActiveSessions() != 1 {
		t.Errorf("Expected 1 session remaining, got %d", sm.GetActiveSessions())
	}

	// Verify the correct session remains
	sm.mu.RLock()
	_, hasActive := sm.sessions["active"]
	_, hasStale := sm.sessions["stale"]
	_, hasStale2 := sm.sessions["stale2"]
	sm.mu.RUnlock()

	if !hasActive {
		t.Error("Active session was incorrectly removed")
	}
	if hasStale || hasStale2 {
		t.Error("Stale sessions were not removed")
	}
}

func TestSessionManager_GetSessionStats(t *testing.T) {
	sm := NewSessionManager()

	// Create sessions in different directories
	sm.GetOrCreateSession("session1", "/project/a")
	sm.GetOrCreateSession("session2", "/project/a")
	sm.GetOrCreateSession("session3", "/project/b")
	sm.GetOrCreateSession("session4", "/project/c")
	sm.GetOrCreateSession("session5", "") // No directory

	stats := sm.GetSessionStats()

	// Check total sessions
	totalSessions, ok := stats["total_sessions"].(int)
	if !ok {
		t.Fatal("total_sessions not found in stats")
	}
	if totalSessions != 5 {
		t.Errorf("Expected 5 total sessions, got %d", totalSessions)
	}

	// Check sessions per directory
	sessionsPerDir, ok := stats["sessions_per_directory"].(map[string]int)
	if !ok {
		t.Fatal("sessions_per_directory not found in stats")
	}

	if sessionsPerDir["/project/a"] != 2 {
		t.Errorf("Expected 2 sessions in /project/a, got %d", sessionsPerDir["/project/a"])
	}
	if sessionsPerDir["/project/b"] != 1 {
		t.Errorf("Expected 1 session in /project/b, got %d", sessionsPerDir["/project/b"])
	}
	if sessionsPerDir["/project/c"] != 1 {
		t.Errorf("Expected 1 session in /project/c, got %d", sessionsPerDir["/project/c"])
	}
}

func TestGetWorkingDirectoryFromContext(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func() context.Context
		expected   string
		shouldFind bool
	}{
		{
			name: "with working directory",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.WorkingDirectoryKey, "/test/dir")
			},
			expected:   "/test/dir",
			shouldFind: true,
		},
		{
			name: "without working directory",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expected:   "",
			shouldFind: false,
		},
		{
			name: "with wrong type value",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.WorkingDirectoryKey, 123)
			},
			expected:   "",
			shouldFind: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupCtx()
			result, _ := GetWorkingDirectoryFromContext(ctx)

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestGetSessionIDFromContext(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func() context.Context
		expected   string
		shouldFind bool
	}{
		{
			name: "with session ID",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.SessionIDKey, "test-session-123")
			},
			expected:   "test-session-123",
			shouldFind: true,
		},
		{
			name: "without session ID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expected:   "",
			shouldFind: false,
		},
		{
			name: "with wrong type value",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), ctxkeys.SessionIDKey, []byte("not-a-string"))
			},
			expected:   "",
			shouldFind: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupCtx()
			result, _ := GetSessionIDFromContext(ctx)

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestStreamableHTTPServerWrapper(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create wrapper with 30 minute session timeout
	wrapper := NewStreamableHTTPServerWrapper(testHandler, 30*time.Minute)

	// Test basic request
	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("X-Working-Directory", "/test/project")
	req.Header.Set("Mcp-Session-Id", "wrapper-test")

	w := httptest.NewRecorder()
	wrapper.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify wrapper was created successfully
	if wrapper == nil {
		t.Error("Expected non-nil wrapper")
	}
}

func TestStreamableHTTPServerWrapper_Stop(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test with zero timeout (no cleanup)
	t.Run("with zero timeout", func(t *testing.T) {
		wrapper := NewStreamableHTTPServerWrapper(testHandler, 0)

		// Give cleanup routine time to exit
		time.Sleep(10 * time.Millisecond)

		// Stop should complete quickly
		done := make(chan bool, 1)
		go func() {
			wrapper.Stop()
			done <- true
		}()

		select {
		case <-done:
			// Good
		case <-time.After(100 * time.Millisecond):
			t.Error("Stop() did not complete quickly")
		}
	})

	// Test with timeout
	t.Run("with timeout", func(t *testing.T) {
		wrapper := NewStreamableHTTPServerWrapper(testHandler, 30*time.Minute)

		// Give cleanup routine time to start
		time.Sleep(10 * time.Millisecond)

		// Stop the wrapper
		wrapper.Stop()

		// Verify wrapper can handle request after stop
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Should not panic
		wrapper.ServeHTTP(w, req)
	})
}

func TestLoggingMiddleware(t *testing.T) {
	// Count how many times the handler is called
	handlerCalled := 0
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled++
		w.WriteHeader(http.StatusOK)
	})

	// Apply logging middleware
	handler := LoggingMiddleware(testHandler)

	// Create test request
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Verify handler was called
	if handlerCalled != 1 {
		t.Errorf("Expected handler to be called once, was called %d times", handlerCalled)
	}

	// Verify response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPMiddleware_ComplexScenarios(t *testing.T) {
	sessionManager := NewSessionManager()

	t.Run("session with changing working directory", func(t *testing.T) {
		var capturedWorkingDir string
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if wd, ok := r.Context().Value(ctxkeys.WorkingDirectoryKey).(string); ok {
				capturedWorkingDir = wd
			}
			w.WriteHeader(http.StatusOK)
		})

		handler := HTTPMiddleware(sessionManager)(testHandler)

		// First request with directory A
		req1 := httptest.NewRequest("POST", "/mcp", nil)
		req1.Header.Set("X-Working-Directory", "/project/a")
		req1.Header.Set("Mcp-Session-Id", "changing-session")

		w1 := httptest.NewRecorder()
		handler.ServeHTTP(w1, req1)

		if capturedWorkingDir != "/project/a" {
			t.Errorf("Expected working directory '/project/a', got '%s'", capturedWorkingDir)
		}

		// Second request with directory B (same session)
		req2 := httptest.NewRequest("POST", "/mcp", nil)
		req2.Header.Set("X-Working-Directory", "/project/b")
		req2.Header.Set("Mcp-Session-Id", "changing-session")

		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, req2)

		if capturedWorkingDir != "/project/b" {
			t.Errorf("Expected working directory updated to '/project/b', got '%s'", capturedWorkingDir)
		}

		// Verify session was updated
		sessions := sessionManager.GetSessionsForDirectory("/project/b")
		if len(sessions) != 1 {
			t.Errorf("Expected 1 session in /project/b, got %d", len(sessions))
		}
	})

	t.Run("request with all headers", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := HTTPMiddleware(sessionManager)(testHandler)

		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("X-Working-Directory", "/full/test")
		req.Header.Set("Mcp-Session-Id", "full-session")
		req.Header.Set("User-Agent", "TestClient/1.0")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Verify session was created
		if sessionManager.GetActiveSessions() < 1 {
			t.Error("Expected at least one active session")
		}
	})
}
