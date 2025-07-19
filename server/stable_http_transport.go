package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	
	"github.com/mark3labs/mcp-go/server"
	"github.com/user/mcp-todo-server/internal/logging"
)

// ConnectionState represents the state of a connection
type ConnectionState int32

const (
	StateConnecting ConnectionState = iota
	StateActive
	StateClosing
	StateClosed
)

// StableHTTPConnection represents a managed HTTP connection
type StableHTTPConnection struct {
	ID               string
	SessionID        string
	WorkingDirectory string
	State            atomic.Int32
	LastActivity     time.Time
	Created          time.Time
	
	// Request handling
	requestQueue     chan *httpRequest
	responseChannels map[string]chan *httpResponse
	responseMutex    sync.RWMutex
	
	// Health monitoring
	lastHeartbeat    time.Time
	heartbeatMissed  int32
	healthCheckTimer *time.Timer
	
	// Resource management
	ctx              context.Context
	cancel           context.CancelFunc
	closeOnce        sync.Once
	
	// Metrics
	requestCount     atomic.Int64
	errorCount       atomic.Int64
	droppedMessages  atomic.Int64
}

type httpRequest struct {
	id       string
	request  *http.Request
	writer   http.ResponseWriter
	body     []byte
	response chan *httpResponse
}

type httpResponse struct {
	statusCode int
	headers    http.Header
	body       []byte
	err        error
}

// StableHTTPTransport wraps the MCP HTTP server with stability improvements
type StableHTTPTransport struct {
	baseServer       *server.StreamableHTTPServer
	connections      sync.Map // map[string]*StableHTTPConnection
	
	// Configuration
	maxRequestsPerConnection int
	requestTimeout          time.Duration
	heartbeatInterval       time.Duration
	connectionTimeout       time.Duration
	
	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	
	// Metrics
	totalConnections atomic.Int64
	activeConnections atomic.Int64
	totalRequests    atomic.Int64
	totalErrors      atomic.Int64
}

// NewStableHTTPTransport creates a new stable HTTP transport wrapper
func NewStableHTTPTransport(baseServer *server.StreamableHTTPServer, opts ...StableHTTPOption) *StableHTTPTransport {
	ctx, cancel := context.WithCancel(context.Background())
	
	transport := &StableHTTPTransport{
		baseServer:               baseServer,
		maxRequestsPerConnection: 100,
		requestTimeout:          30 * time.Second,
		heartbeatInterval:       15 * time.Second,
		connectionTimeout:       5 * time.Minute,
		ctx:                     ctx,
		cancel:                  cancel,
	}
	
	// Apply options
	for _, opt := range opts {
		opt(transport)
	}
	
	// Start monitoring goroutine
	transport.wg.Add(1)
	go transport.monitorConnections()
	
	return transport
}

// StableHTTPOption configures the stable transport
type StableHTTPOption func(*StableHTTPTransport)

// WithMaxRequestsPerConnection sets the request queue size
func WithMaxRequestsPerConnection(max int) StableHTTPOption {
	return func(t *StableHTTPTransport) {
		t.maxRequestsPerConnection = max
	}
}

// WithRequestTimeout sets the request timeout
func WithRequestTimeout(timeout time.Duration) StableHTTPOption {
	return func(t *StableHTTPTransport) {
		t.requestTimeout = timeout
	}
}

// WithConnectionTimeout sets the connection timeout
func WithConnectionTimeout(timeout time.Duration) StableHTTPOption {
	return func(t *StableHTTPTransport) {
		t.connectionTimeout = timeout
	}
}

// ServeHTTP implements http.Handler with stability improvements
func (t *StableHTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Wrap with panic recovery
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Fprintf(os.Stderr, "[StableHTTP] Panic recovered in ServeHTTP: %v\n", recovered)
			t.totalErrors.Add(1)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()
	
	// Extract connection info
	sessionID := r.Header.Get("Mcp-Session-Id")
	workingDir := r.Header.Get("X-Working-Directory")
	
	// Get or create connection
	conn := t.getOrCreateConnection(sessionID, workingDir)
	
	// Check connection state
	state := ConnectionState(conn.State.Load())
	if state == StateClosing || state == StateClosed {
		http.Error(w, "Connection is closing", http.StatusServiceUnavailable)
		return
	}
	
	// Update activity
	conn.LastActivity = time.Now()
	
	// Handle special endpoints
	if r.URL.Path == "/mcp/heartbeat" {
		t.handleHeartbeat(conn, w, r)
		return
	}
	
	// Read request body with size limit
	body, err := t.readRequestBody(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request: %v", err), http.StatusBadRequest)
		return
	}
	
	// Create request with timeout
	ctx, cancel := context.WithTimeout(r.Context(), t.requestTimeout)
	defer cancel()
	
	requestID := fmt.Sprintf("%s-%d", sessionID, time.Now().UnixNano())
	request := &httpRequest{
		id:       requestID,
		request:  r.WithContext(ctx),
		writer:   w,
		body:     body,
		response: make(chan *httpResponse, 1),
	}
	
	// Try to queue request
	select {
	case conn.requestQueue <- request:
		conn.requestCount.Add(1)
		t.totalRequests.Add(1)
	case <-time.After(100 * time.Millisecond):
		// Queue is full, apply backpressure
		conn.droppedMessages.Add(1)
		http.Error(w, "Server is overloaded, please retry", http.StatusServiceUnavailable)
		return
	}
	
	// Wait for response
	select {
	case resp := <-request.response:
		if resp.err != nil {
			conn.errorCount.Add(1)
			t.totalErrors.Add(1)
			http.Error(w, resp.err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Write response headers
		for key, values := range resp.headers {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		
		// Write status code
		if resp.statusCode != 0 {
			w.WriteHeader(resp.statusCode)
		}
		
		// Write body
		if len(resp.body) > 0 {
			if _, err := w.Write(resp.body); err != nil {
				logging.StableHTTPf("Failed to write response: %v", err)
			}
		}
		
	case <-ctx.Done():
		conn.errorCount.Add(1)
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
		
	case <-conn.ctx.Done():
		http.Error(w, "Connection closed", http.StatusServiceUnavailable)
	}
}

// getOrCreateConnection manages connection lifecycle
func (t *StableHTTPTransport) getOrCreateConnection(sessionID, workingDir string) *StableHTTPConnection {
	// Try to get existing connection
	if existing, ok := t.connections.Load(sessionID); ok {
		conn := existing.(*StableHTTPConnection)
		// Update working directory if changed
		if workingDir != "" && conn.WorkingDirectory != workingDir {
			conn.WorkingDirectory = workingDir
		}
		return conn
	}
	
	// Create new connection
	ctx, cancel := context.WithCancel(t.ctx)
	conn := &StableHTTPConnection{
		ID:               fmt.Sprintf("conn-%s-%d", sessionID, time.Now().UnixNano()),
		SessionID:        sessionID,
		WorkingDirectory: workingDir,
		LastActivity:     time.Now(),
		Created:          time.Now(),
		requestQueue:     make(chan *httpRequest, t.maxRequestsPerConnection),
		responseChannels: make(map[string]chan *httpResponse),
		ctx:              ctx,
		cancel:           cancel,
		lastHeartbeat:    time.Now(),
	}
	
	// Set initial state
	conn.State.Store(int32(StateConnecting))
	
	// Store connection
	actual, loaded := t.connections.LoadOrStore(sessionID, conn)
	if loaded {
		// Another goroutine created it first
		cancel()
		return actual.(*StableHTTPConnection)
	}
	
	// Start connection handler
	t.wg.Add(1)
	go t.handleConnection(conn)
	
	// Update metrics
	t.totalConnections.Add(1)
	t.activeConnections.Add(1)
	
	// Set to active state
	conn.State.Store(int32(StateActive))
	
	logging.StableHTTPf("Created new connection %s for session %s", conn.ID, sessionID)
	
	return conn
}

// handleConnection processes requests for a connection
func (t *StableHTTPTransport) handleConnection(conn *StableHTTPConnection) {
	defer t.wg.Done()
	defer func() {
		if recovered := recover(); recovered != nil {
			logging.StableHTTPf("Panic in connection handler %s: %v", conn.ID, recovered)
		}
		t.closeConnection(conn)
	}()
	
	logging.StableHTTPf("Starting connection handler for %s", conn.ID)
	
	for {
		select {
		case req := <-conn.requestQueue:
			t.processRequest(conn, req)
			
		case <-conn.ctx.Done():
			logging.StableHTTPf("Connection %s context cancelled", conn.ID)
			return
			
		case <-t.ctx.Done():
			logging.StableHTTPf("Transport shutting down, closing connection %s", conn.ID)
			return
		}
	}
}

// processRequest handles a single request with proper error recovery
func (t *StableHTTPTransport) processRequest(conn *StableHTTPConnection, req *httpRequest) {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Fprintf(os.Stderr, "[StableHTTP] Panic processing request %s: %v\n", req.id, recovered)
			req.response <- &httpResponse{
				statusCode: http.StatusInternalServerError,
				err:        fmt.Errorf("internal server error"),
			}
		}
	}()
	
	// Create a wrapped response writer that captures the response
	wrapper := newResponseWrapper()
	
	// Restore request body
	if len(req.body) > 0 {
		req.request.Body = io.NopCloser(bytes.NewReader(req.body))
		req.request.ContentLength = int64(len(req.body))
	}
	
	// Forward to base server with timeout
	done := make(chan struct{})
	go func() {
		defer close(done)
		t.baseServer.ServeHTTP(wrapper, req.request)
	}()
	
	select {
	case <-done:
		// Success
		req.response <- &httpResponse{
			statusCode: wrapper.statusCode,
			headers:    wrapper.Header(),
			body:       wrapper.body.Bytes(),
		}
		
	case <-req.request.Context().Done():
		// Request timeout
		req.response <- &httpResponse{
			statusCode: http.StatusRequestTimeout,
			err:        fmt.Errorf("request timeout"),
		}
		
	case <-conn.ctx.Done():
		// Connection closed
		req.response <- &httpResponse{
			statusCode: http.StatusServiceUnavailable,
			err:        fmt.Errorf("connection closed"),
		}
	}
}

// handleHeartbeat processes heartbeat requests
func (t *StableHTTPTransport) handleHeartbeat(conn *StableHTTPConnection, w http.ResponseWriter, r *http.Request) {
	conn.lastHeartbeat = time.Now()
	conn.heartbeatMissed = 0
	
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"sessionId": conn.SessionID,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// monitorConnections periodically checks connection health
func (t *StableHTTPTransport) monitorConnections() {
	defer t.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			t.checkConnections()
			
		case <-t.ctx.Done():
			return
		}
	}
}

// checkConnections removes stale connections
func (t *StableHTTPTransport) checkConnections() {
	now := time.Now()
	var toClose []*StableHTTPConnection
	
	t.connections.Range(func(key, value interface{}) bool {
		conn := value.(*StableHTTPConnection)
		
		// Check if connection is stale
		if now.Sub(conn.LastActivity) > t.connectionTimeout {
			toClose = append(toClose, conn)
			fmt.Fprintf(os.Stderr, "[StableHTTP] Connection %s is stale (inactive for %v)\n", 
				conn.ID, now.Sub(conn.LastActivity))
		}
		
		// Check heartbeat
		if now.Sub(conn.lastHeartbeat) > t.heartbeatInterval*3 {
			atomic.AddInt32(&conn.heartbeatMissed, 1)
			if conn.heartbeatMissed > 3 {
				toClose = append(toClose, conn)
				fmt.Fprintf(os.Stderr, "[StableHTTP] Connection %s missed too many heartbeats\n", conn.ID)
			}
		}
		
		return true
	})
	
	// Close stale connections
	for _, conn := range toClose {
		t.closeConnection(conn)
	}
	
	// Log metrics
	logging.StableHTTPf("Monitor: %d active connections, %d total requests, %d errors",
		t.activeConnections.Load(), t.totalRequests.Load(), t.totalErrors.Load())
}

// closeConnection cleanly shuts down a connection
func (t *StableHTTPTransport) closeConnection(conn *StableHTTPConnection) {
	conn.closeOnce.Do(func() {
		// Update state
		conn.State.Store(int32(StateClosing))
		
		// Cancel context
		conn.cancel()
		
		// Remove from map
		t.connections.Delete(conn.SessionID)
		
		// Update metrics
		t.activeConnections.Add(-1)
		
		// Close request queue
		close(conn.requestQueue)
		
		// Log closure
		logging.StableHTTPf("Closed connection %s (requests: %d, errors: %d, dropped: %d)",
			conn.ID, conn.requestCount.Load(), conn.errorCount.Load(), conn.droppedMessages.Load())
		
		// Update state
		conn.State.Store(int32(StateClosed))
	})
}

// Shutdown gracefully shuts down the transport
func (t *StableHTTPTransport) Shutdown(ctx context.Context) error {
	logging.StableHTTPf("Shutting down transport")
	
	// Cancel transport context
	t.cancel()
	
	// Close all connections
	t.connections.Range(func(key, value interface{}) bool {
		conn := value.(*StableHTTPConnection)
		t.closeConnection(conn)
		return true
	})
	
	// Wait for all goroutines with timeout
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		logging.StableHTTPf("Transport shutdown complete")
		return nil
	case <-ctx.Done():
		logging.StableHTTPf("Transport shutdown timeout")
		return ctx.Err()
	}
}

// readRequestBody reads the request body with size limit
func (t *StableHTTPTransport) readRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	
	// Limit body size to 10MB
	maxSize := int64(10 * 1024 * 1024)
	limitedReader := io.LimitReader(r.Body, maxSize+1)
	body, err := io.ReadAll(limitedReader)
	r.Body.Close()
	
	if err != nil {
		return nil, err
	}
	
	// Check if body exceeds size limit
	if int64(len(body)) > maxSize {
		return nil, fmt.Errorf("request body too large: exceeds %d bytes", maxSize)
	}
	
	// Validate JSON if content type is application/json
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") && len(body) > 0 {
		var js json.RawMessage
		if err := json.Unmarshal(body, &js); err != nil {
			return nil, fmt.Errorf("invalid JSON: %v", err)
		}
	}
	
	return body, nil
}

// GetMetrics returns current transport metrics
func (t *StableHTTPTransport) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"total_connections":  t.totalConnections.Load(),
		"active_connections": t.activeConnections.Load(),
		"total_requests":    t.totalRequests.Load(),
		"total_errors":      t.totalErrors.Load(),
	}
	
	// Add per-connection metrics
	connections := make([]map[string]interface{}, 0)
	t.connections.Range(func(key, value interface{}) bool {
		conn := value.(*StableHTTPConnection)
		connections = append(connections, map[string]interface{}{
			"id":               conn.ID,
			"session_id":       conn.SessionID,
			"state":           ConnectionState(conn.State.Load()).String(),
			"requests":        conn.requestCount.Load(),
			"errors":          conn.errorCount.Load(),
			"dropped":         conn.droppedMessages.Load(),
			"last_activity":   conn.LastActivity.Format(time.RFC3339),
			"uptime":          time.Since(conn.Created).String(),
		})
		return true
	})
	metrics["connections"] = connections
	
	return metrics
}

// String returns the string representation of a connection state
func (s ConnectionState) String() string {
	switch s {
	case StateConnecting:
		return "connecting"
	case StateActive:
		return "active"
	case StateClosing:
		return "closing"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}