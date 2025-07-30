package server

import (
	"testing"
	"time"
)

func TestServerTimeoutConfiguration(t *testing.T) {
	tests := []struct {
		name              string
		opts              []ServerOption
		expectedRequest   time.Duration
		expectedReadWrite time.Duration
		expectedIdle      time.Duration
	}{
		{
			name:              "Default timeouts",
			opts:              []ServerOption{},
			expectedRequest:   60 * time.Second,
			expectedReadWrite: 120 * time.Second,
			expectedIdle:      120 * time.Second,
		},
		{
			name: "Custom timeouts for devcontainer",
			opts: []ServerOption{
				WithHTTPRequestTimeout(180 * time.Second),
				WithHTTPReadTimeout(300 * time.Second),
				WithHTTPWriteTimeout(300 * time.Second),
				WithHTTPIdleTimeout(600 * time.Second),
			},
			expectedRequest:   180 * time.Second,
			expectedReadWrite: 300 * time.Second,
			expectedIdle:      600 * time.Second,
		},
		{
			name: "Zero timeouts (disabled)",
			opts: []ServerOption{
				WithHTTPRequestTimeout(0),
				WithHTTPReadTimeout(0),
				WithHTTPWriteTimeout(0),
				WithHTTPIdleTimeout(0),
			},
			expectedRequest:   0,
			expectedReadWrite: 0,
			expectedIdle:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal server just to test configuration
			ts := &TodoServer{
				transport:         "http",
				startTime:         time.Now(),
				sessionTimeout:    7 * 24 * time.Hour,
				managerTimeout:    24 * time.Hour,
				heartbeatInterval: 30 * time.Second,
				requestTimeout:    60 * time.Second,
				httpReadTimeout:   120 * time.Second,
				httpWriteTimeout:  120 * time.Second,
				httpIdleTimeout:   120 * time.Second,
			}

			// Apply options
			for _, opt := range tt.opts {
				opt(ts)
			}

			// Verify timeouts
			if ts.requestTimeout != tt.expectedRequest {
				t.Errorf("Expected request timeout %v, got %v", tt.expectedRequest, ts.requestTimeout)
			}

			if ts.httpReadTimeout != tt.expectedReadWrite {
				t.Errorf("Expected HTTP read timeout %v, got %v", tt.expectedReadWrite, ts.httpReadTimeout)
			}

			if ts.httpWriteTimeout != tt.expectedReadWrite {
				t.Errorf("Expected HTTP write timeout %v, got %v", tt.expectedReadWrite, ts.httpWriteTimeout)
			}

			if ts.httpIdleTimeout != tt.expectedIdle {
				t.Errorf("Expected HTTP idle timeout %v, got %v", tt.expectedIdle, ts.httpIdleTimeout)
			}
		})
	}
}
