package server

import (
	"testing"
	"time"
)

func TestServerOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   ServerOption
		validate func(*testing.T, *TodoServer)
	}{
		{
			name: "WithTransport",
			option: WithTransport("http"),
			validate: func(t *testing.T, s *TodoServer) {
				if s.transport != "http" {
					t.Errorf("Expected transport 'http', got '%s'", s.transport)
				}
			},
		},
		{
			name: "WithSessionTimeout",
			option: WithSessionTimeout(5 * time.Minute),
			validate: func(t *testing.T, s *TodoServer) {
				if s.sessionTimeout != 5*time.Minute {
					t.Errorf("Expected session timeout 5m, got %v", s.sessionTimeout)
				}
			},
		},
		{
			name: "WithManagerTimeout",
			option: WithManagerTimeout(10 * time.Minute),
			validate: func(t *testing.T, s *TodoServer) {
				if s.managerTimeout != 10*time.Minute {
					t.Errorf("Expected manager timeout 10m, got %v", s.managerTimeout)
				}
			},
		},
		{
			name: "WithHeartbeatInterval",
			option: WithHeartbeatInterval(30 * time.Second),
			validate: func(t *testing.T, s *TodoServer) {
				if s.heartbeatInterval != 30*time.Second {
					t.Errorf("Expected heartbeat interval 30s, got %v", s.heartbeatInterval)
				}
			},
		},
		{
			name: "WithNoAutoArchive",
			option: WithNoAutoArchive(true),
			validate: func(t *testing.T, s *TodoServer) {
				if !s.noAutoArchive {
					t.Error("Expected noAutoArchive to be true")
				}
			},
		},
		{
			name: "WithHTTPRequestTimeout",
			option: WithHTTPRequestTimeout(60 * time.Second),
			validate: func(t *testing.T, s *TodoServer) {
				if s.requestTimeout != 60*time.Second {
					t.Errorf("Expected request timeout 60s, got %v", s.requestTimeout)
				}
			},
		},
		{
			name: "WithHTTPReadTimeout",
			option: WithHTTPReadTimeout(15 * time.Second),
			validate: func(t *testing.T, s *TodoServer) {
				if s.httpReadTimeout != 15*time.Second {
					t.Errorf("Expected HTTP read timeout 15s, got %v", s.httpReadTimeout)
				}
			},
		},
		{
			name: "WithHTTPWriteTimeout",
			option: WithHTTPWriteTimeout(15 * time.Second),
			validate: func(t *testing.T, s *TodoServer) {
				if s.httpWriteTimeout != 15*time.Second {
					t.Errorf("Expected HTTP write timeout 15s, got %v", s.httpWriteTimeout)
				}
			},
		},
		{
			name: "WithHTTPIdleTimeout",
			option: WithHTTPIdleTimeout(120 * time.Second),
			validate: func(t *testing.T, s *TodoServer) {
				if s.httpIdleTimeout != 120*time.Second {
					t.Errorf("Expected HTTP idle timeout 120s, got %v", s.httpIdleTimeout)
				}
			},
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a minimal server with the option
			server := &TodoServer{}
			tc.option(server)
			tc.validate(t, server)
		})
	}
}

func TestMultipleServerOptions(t *testing.T) {
	// Test that multiple options can be applied
	server := &TodoServer{}
	
	options := []ServerOption{
		WithTransport("stdio"),
		WithSessionTimeout(10 * time.Minute),
		WithManagerTimeout(5 * time.Minute),
		WithHeartbeatInterval(15 * time.Second),
		WithNoAutoArchive(true),
		WithHTTPRequestTimeout(30 * time.Second),
		WithHTTPReadTimeout(10 * time.Second),
		WithHTTPWriteTimeout(10 * time.Second),
		WithHTTPIdleTimeout(60 * time.Second),
	}
	
	for _, opt := range options {
		opt(server)
	}
	
	// Validate all options were applied
	if server.transport != "stdio" {
		t.Errorf("Expected transport 'stdio', got '%s'", server.transport)
	}
	if server.sessionTimeout != 10*time.Minute {
		t.Errorf("Expected session timeout 10m, got %v", server.sessionTimeout)
	}
	if server.managerTimeout != 5*time.Minute {
		t.Errorf("Expected manager timeout 5m, got %v", server.managerTimeout)
	}
	if server.heartbeatInterval != 15*time.Second {
		t.Errorf("Expected heartbeat interval 15s, got %v", server.heartbeatInterval)
	}
	if !server.noAutoArchive {
		t.Error("Expected noAutoArchive to be true")
	}
	if server.requestTimeout != 30*time.Second {
		t.Errorf("Expected request timeout 30s, got %v", server.requestTimeout)
	}
	if server.httpReadTimeout != 10*time.Second {
		t.Errorf("Expected HTTP read timeout 10s, got %v", server.httpReadTimeout)
	}
	if server.httpWriteTimeout != 10*time.Second {
		t.Errorf("Expected HTTP write timeout 10s, got %v", server.httpWriteTimeout)
	}
	if server.httpIdleTimeout != 60*time.Second {
		t.Errorf("Expected HTTP idle timeout 60s, got %v", server.httpIdleTimeout)
	}
}