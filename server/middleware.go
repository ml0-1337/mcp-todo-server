package server

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	
	ctxkeys "github.com/user/mcp-todo-server/internal/context"
)

// SessionInfo stores session-specific information
type SessionInfo struct {
	ID               string
	WorkingDirectory string
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
	
	if session, exists := sm.sessions[sessionID]; exists {
		// Update working directory if provided
		if workingDir != "" && session.WorkingDirectory != workingDir {
			log.Printf("Updating working directory for session %s: %s -> %s", 
				sessionID, session.WorkingDirectory, workingDir)
			session.WorkingDirectory = workingDir
		}
		return session
	}
	
	// Create new session
	session := &SessionInfo{
		ID:               sessionID,
		WorkingDirectory: workingDir,
	}
	sm.sessions[sessionID] = session
	log.Printf("Created new session %s with working directory: %s", sessionID, workingDir)
	return session
}

// RemoveSession removes a session
func (sm *SessionManager) RemoveSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if _, exists := sm.sessions[sessionID]; exists {
		delete(sm.sessions, sessionID)
		log.Printf("Removed session %s", sessionID)
	}
}

// HTTPMiddleware wraps an http.Handler to extract headers and manage sessions
func HTTPMiddleware(sessionManager *SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract working directory from header
			workingDir := r.Header.Get("X-Working-Directory")
			if workingDir != "" {
				log.Printf("Received X-Working-Directory header: %s", workingDir)
			}
			
			// Extract session ID from header
			sessionID := r.Header.Get("Mcp-Session-Id")
			
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
				// Try to get existing session
				sessionManager.mu.RLock()
				if session, exists := sessionManager.sessions[sessionID]; exists {
					sessionManager.mu.RUnlock()
					ctx := context.WithValue(r.Context(), ctxkeys.SessionIDKey, session.ID)
					ctx = context.WithValue(ctx, ctxkeys.WorkingDirectoryKey, session.WorkingDirectory)
					r = r.WithContext(ctx)
				} else {
					sessionManager.mu.RUnlock()
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
}

// NewStreamableHTTPServerWrapper creates a new wrapper with middleware
func NewStreamableHTTPServerWrapper(streamableServer http.Handler) *StreamableHTTPServerWrapper {
	sessionManager := NewSessionManager()
	return &StreamableHTTPServerWrapper{
		server:         streamableServer,
		sessionManager: sessionManager,
	}
}

// ServeHTTP implements http.Handler with middleware
func (w *StreamableHTTPServerWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// Apply middleware
	handler := HTTPMiddleware(w.sessionManager)(w.server)
	handler.ServeHTTP(rw, r)
}

// LoggingMiddleware logs incoming requests (optional, for debugging)
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("HTTP %s %s", r.Method, r.URL.Path)
		
		// Log interesting headers
		for name, values := range r.Header {
			if strings.HasPrefix(name, "X-") || strings.HasPrefix(name, "Mcp-") {
				log.Printf("  Header %s: %s", name, strings.Join(values, ", "))
			}
		}
		
		next.ServeHTTP(w, r)
	})
}