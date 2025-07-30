package server

import (
	"testing"
	"time"
)

// TestConnectionStateString tests the String method of ConnectionState
func TestConnectionStateString(t *testing.T) {
	tests := []struct {
		state    ConnectionState
		expected string
	}{
		{StateConnecting, "connecting"},
		{StateActive, "active"},
		{StateClosing, "closing"},
		{StateClosed, "closed"},
		{ConnectionState(99), "unknown"},
		{ConnectionState(-1), "unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("ConnectionState.String() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// TestStableHTTPConnectionGetters tests getter methods
func TestStableHTTPConnectionGetters(t *testing.T) {
	conn := &StableHTTPConnection{
		ID:               "test-conn-123",
		SessionID:        "session-456",
		WorkingDirectory: "/test/dir",
		Created:          time.Now(),
		LastActivity:     time.Now(),
	}
	
	// Test initial state
	conn.State.Store(int32(StateActive))
	
	if state := ConnectionState(conn.State.Load()); state != StateActive {
		t.Errorf("Expected state %v, got %v", StateActive, state)
	}
	
	// Test state transitions
	conn.State.Store(int32(StateClosing))
	if state := ConnectionState(conn.State.Load()); state != StateClosing {
		t.Errorf("Expected state %v after transition, got %v", StateClosing, state)
	}
}

// TestHTTPRequest tests httpRequest struct
func TestHTTPRequest(t *testing.T) {
	req := &httpRequest{
		id:       "req-123",
		body:     []byte(`{"test": "data"}`),
		response: make(chan *httpResponse, 1),
	}
	
	// Test request fields
	if req.id != "req-123" {
		t.Errorf("Expected request id 'req-123', got %s", req.id)
	}
	
	if string(req.body) != `{"test": "data"}` {
		t.Errorf("Expected request body to match")
	}
	
	// Test response channel
	resp := &httpResponse{
		body:       []byte(`{"result": "ok"}`),
		statusCode: 200,
	}
	
	req.response <- resp
	
	select {
	case received := <-req.response:
		if received.statusCode != 200 {
			t.Errorf("Expected status 200, got %d", received.statusCode)
		}
		if string(received.body) != `{"result": "ok"}` {
			t.Errorf("Expected response body to match")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for response")
	}
}

// TestHTTPResponse tests httpResponse struct  
func TestHTTPResponse(t *testing.T) {
	tests := []struct {
		name   string
		resp   httpResponse
		expect struct {
			statusCode int
			body       string
			err        error
		}
	}{
		{
			name: "Success response",
			resp: httpResponse{
				body:       []byte(`{"success": true}`),
				statusCode: 200,
				err:        nil,
			},
			expect: struct {
				statusCode int
				body       string
				err        error
			}{
				statusCode: 200,
				body:       `{"success": true}`,
				err:        nil,
			},
		},
		{
			name: "Error response",
			resp: httpResponse{
				body:       []byte(`{"error": "bad request"}`),
				statusCode: 400,
				err:        nil,
			},
			expect: struct {
				statusCode int
				body       string
				err        error
			}{
				statusCode: 400,
				body:       `{"error": "bad request"}`,
				err:        nil,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resp.statusCode != tt.expect.statusCode {
				t.Errorf("Expected status %d, got %d", tt.expect.statusCode, tt.resp.statusCode)
			}
			if string(tt.resp.body) != tt.expect.body {
				t.Errorf("Expected body %s, got %s", tt.expect.body, string(tt.resp.body))
			}
			if tt.resp.err != tt.expect.err {
				t.Errorf("Expected error %v, got %v", tt.expect.err, tt.resp.err)
			}
		})
	}
}