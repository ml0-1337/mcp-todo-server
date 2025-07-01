package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthCheckEndpoint(t *testing.T) {
	// Create a test server
	ts, err := NewTodoServer(WithTransport("http"))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer ts.Close()

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
				
				if version, ok := body["version"].(string); !ok || version != "2.0.0" {
					t.Errorf("Expected version '2.0.0', got %v", body["version"])
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

func TestHealthCheckUptime(t *testing.T) {
	// Create a test server
	ts, err := NewTodoServer(WithTransport("http"))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer ts.Close()

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
	tolerance := float64(1000) // 1 second tolerance

	if uptimeMs < expectedMs-tolerance || uptimeMs > expectedMs+tolerance {
		t.Errorf("Expected uptime around %fms, got %fms", expectedMs, uptimeMs)
	}
}

func TestHealthCheckWithSessions(t *testing.T) {
	// Create a test server
	ts, err := NewTodoServer(WithTransport("http"))
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer ts.Close()

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