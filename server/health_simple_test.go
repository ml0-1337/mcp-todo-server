package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

func TestHealthCheckHandlerSimple(t *testing.T) {
	// Create a minimal test server
	mcpServer := server.NewMCPServer("Test", "1.0.0")
	ts := &TodoServer{
		mcpServer: mcpServer,
		transport: "http",
		startTime: time.Now(),
	}

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	ts.handleHealthCheck(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
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

	if sessions, ok := body["sessions"].(float64); !ok || sessions != 0 {
		t.Errorf("Expected sessions 0, got %v", body["sessions"])
	}
}