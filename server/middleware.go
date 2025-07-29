package server

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
	
	ctxkeys "github.com/user/mcp-todo-server/internal/context"
	"github.com/user/mcp-todo-server/internal/logging"
)

// SessionInfo stores session-specific information
type SessionInfo struct {
	ID               string
	WorkingDirectory string
	LastActivity     time.Time
}

// SessionManager manages session information
type SessionManager struct {
	sessions map[string]*SessionInfo
	mu       sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*SessionInfo),
	}
}

// GetOrCreateSession retrieves or creates a session
func (sm *SessionManager) GetOrCreateSession(sessionID string, workingDir string) *SessionInfo {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	now := time.Now()
	
	// Check if session ID already exists
	if session, exists := sm.sessions[sessionID]; exists {
		// Update last activity
		session.LastActivity = now
		
		// Update working directory if provided and different
		if workingDir != "" && session.WorkingDirectory != workingDir {
			logging.Infof("Updating working directory for session %s: %s -> %s", 
				sessionID, session.WorkingDirectory, workingDir)
			session.WorkingDirectory = workingDir
		}
		return session
	}
	
	// Create new session
	session := &SessionInfo{
		ID:               sessionID,
		WorkingDirectory: workingDir,
		LastActivity:     now,
	}
	sm.sessions[sessionID] = session
	
	logging.Infof("Created new session %s with working directory: %s", sessionID, workingDir)
	return session
}

// RemoveSession removes a session
func (sm *SessionManager) RemoveSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if _, exists := sm.sessions[sessionID]; exists {
		delete(sm.sessions, sessionID)
		logging.Infof("Removed session %s", sessionID)
	}
}

// CleanupStaleSessions removes sessions that haven't been active for the given duration
func (sm *SessionManager) CleanupStaleSessions(inactivityTimeout time.Duration) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	now := time.Now()
	removedCount := 0
	
	for id, session := range sm.sessions {
		if now.Sub(session.LastActivity) > inactivityTimeout {
			delete(sm.sessions, id)
			removedCount++
			logging.Infof("Removed stale session %s (inactive for %v)", id, now.Sub(session.LastActivity))
		}
	}
	
	if removedCount > 0 {
		logging.Infof("Cleaned up %d stale sessions", removedCount)
	}
	
	return removedCount
}

// GetActiveSessions returns the count of active sessions
func (sm *SessionManager) GetActiveSessions() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// GetSessionsForDirectory returns all sessions for a given working directory
func (sm *SessionManager) GetSessionsForDirectory(workingDir string) []*SessionInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	var sessions []*SessionInfo
	for _, session := range sm.sessions {
		if session.WorkingDirectory == workingDir {
			sessions = append(sessions, session)
		}
	}
	return sessions
}


// GetSessionStats returns detailed session statistics
func (sm *SessionManager) GetSessionStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	stats := make(map[string]interface{})
	stats["total_sessions"] = len(sm.sessions)
	
	// Count sessions per directory
	dirCount := make(map[string]int)
	for _, session := range sm.sessions {
		if session.WorkingDirectory != "" {
			dirCount[session.WorkingDirectory]++
		}
	}
	stats["sessions_per_directory"] = dirCount
	
	return stats
}

// HTTPMiddleware wraps an http.Handler to extract headers and manage sessions
func HTTPMiddleware(sessionManager *SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log incoming request details
			logging.Connectionf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			
			// Extract working directory from header
			workingDir := r.Header.Get("X-Working-Directory")
			if workingDir != "" {
				logging.Headerf("X-Working-Directory: %s", workingDir)
			}
			
			// Extract session ID from header
			sessionID := r.Header.Get("Mcp-Session-Id")
			if sessionID != "" {
				logging.Headerf("Mcp-Session-Id: %s", sessionID)
			}
			
			// Log other relevant headers
			userAgent := r.Header.Get("User-Agent")
			if userAgent != "" {
				logging.Headerf("User-Agent: %s", userAgent)
			}
			
			// If we have both, manage the session
			if sessionID != "" && workingDir != "" {
				session := sessionManager.GetOrCreateSession(sessionID, workingDir)
				// Add session info to context
				ctx := context.WithValue(r.Context(), ctxkeys.SessionIDKey, session.ID)
				ctx = context.WithValue(ctx, ctxkeys.WorkingDirectoryKey, session.WorkingDirectory)
				r = r.WithContext(ctx)
			} else if workingDir != "" {
				// Just add working directory to context without session
				ctx := context.WithValue(r.Context(), ctxkeys.WorkingDirectoryKey, workingDir)
				r = r.WithContext(ctx)
			} else if sessionID != "" {
				// Try to get existing session and update activity
				sessionManager.mu.Lock()
				if session, exists := sessionManager.sessions[sessionID]; exists {
					session.LastActivity = time.Now()
					sessionManager.mu.Unlock()
					ctx := context.WithValue(r.Context(), ctxkeys.SessionIDKey, session.ID)
					ctx = context.WithValue(ctx, ctxkeys.WorkingDirectoryKey, session.WorkingDirectory)
					r = r.WithContext(ctx)
				} else {
					sessionManager.mu.Unlock()
				}
			}
			
			// Handle DELETE requests for session cleanup
			if r.Method == http.MethodDelete && sessionID != "" {
				sessionManager.RemoveSession(sessionID)
			}
			
			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// GetWorkingDirectoryFromContext extracts working directory from context
func GetWorkingDirectoryFromContext(ctx context.Context) (string, bool) {
	workingDir, ok := ctx.Value(ctxkeys.WorkingDirectoryKey).(string)
	return workingDir, ok
}

// GetSessionIDFromContext extracts session ID from context
func GetSessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(ctxkeys.SessionIDKey).(string)
	return sessionID, ok
}

// StreamableHTTPServerWrapper wraps StreamableHTTPServer with middleware
type StreamableHTTPServerWrapper struct {
	server         http.Handler
	sessionManager *SessionManager
	sessionTimeout time.Duration
	cleanupStop    chan struct{}
	cleanupDone    chan struct{}
	stopOnce       sync.Once
}

// NewStreamableHTTPServerWrapper creates a new wrapper with middleware
func NewStreamableHTTPServerWrapper(streamableServer http.Handler, sessionTimeout time.Duration) *StreamableHTTPServerWrapper {
	sessionManager := NewSessionManager()
	wrapper := &StreamableHTTPServerWrapper{
		server:         streamableServer,
		sessionManager: sessionManager,
		sessionTimeout: sessionTimeout,
		cleanupStop:    make(chan struct{}),
		cleanupDone:    make(chan struct{}),
	}
	
	// Start cleanup routine
	go wrapper.cleanupRoutine()
	
	return wrapper
}

// ServeHTTP implements http.Handler with middleware
func (w *StreamableHTTPServerWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// Apply middleware
	handler := HTTPMiddleware(w.sessionManager)(w.server)
	handler.ServeHTTP(rw, r)
}

// cleanupRoutine periodically cleans up stale sessions
func (w *StreamableHTTPServerWrapper) cleanupRoutine() {
	defer close(w.cleanupDone)
	
	// Run cleanup every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	// Use configured session timeout (0 means no cleanup)
	if w.sessionTimeout == 0 {
		logging.Infof("Session cleanup disabled (timeout=0)")
		return
	}
	
	for {
		select {
		case <-ticker.C:
			removed := w.sessionManager.CleanupStaleSessions(w.sessionTimeout)
			if removed > 0 {
				logging.Infof("Session cleanup: removed %d stale sessions", removed)
			}
		case <-w.cleanupStop:
			logging.Infof("Stopping session cleanup routine")
			return
		}
	}
}

// Stop stops the cleanup routine
func (w *StreamableHTTPServerWrapper) Stop() {
	w.stopOnce.Do(func() {
		close(w.cleanupStop)
		<-w.cleanupDone
	})
}

// LoggingMiddleware logs incoming requests (optional, for debugging)
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.Debugf("HTTP %s %s", r.Method, r.URL.Path)
		
		// Log interesting headers
		for name, values := range r.Header {
			if strings.HasPrefix(name, "X-") || strings.HasPrefix(name, "Mcp-") {
				logging.Debugf("  Header %s: %s", name, strings.Join(values, ", "))
			}
		}
		
		next.ServeHTTP(w, r)
	})
}