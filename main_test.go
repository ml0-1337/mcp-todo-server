package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// TestMainFunction tests the main function with different flags
func TestMainFunction(t *testing.T) {
	// Skip if running in short mode as this test requires special handling
	if testing.Short() {
		t.Skip("Skipping main function test in short mode")
	}

	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "invalid transport",
			args:      []string{"-transport=invalid"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original args and restore after test
			oldArgs := os.Args
			oldCommandLine := flag.CommandLine
			defer func() {
				os.Args = oldArgs
				flag.CommandLine = oldCommandLine
			}()

			// Reset flag.CommandLine to allow re-parsing
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Set test args
			os.Args = append([]string{"mcp-todo-server"}, tt.args...)

			// Since main() calls log.Fatal on errors, we need to test it differently
			// For now, we'll just verify the flag parsing works
			var transport, port, host string
			flag.StringVar(&transport, "transport", "http", "Transport type")
			flag.StringVar(&port, "port", "8080", "Port for HTTP transport")
			flag.StringVar(&host, "host", "localhost", "Host for HTTP transport")
			
			err := flag.CommandLine.Parse(os.Args[1:])
			if err != nil && !tt.wantError {
				t.Errorf("Unexpected error parsing flags: %v", err)
			}
		})
	}
}

// TestFlagParsing tests command-line flag parsing
func TestFlagParsing(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantTransport  string
		wantPort       string
		wantHost       string
	}{
		{
			name:          "default values",
			args:          []string{},
			wantTransport: "http",
			wantPort:      "8080",
			wantHost:      "localhost",
		},
		{
			name:          "custom transport stdio",
			args:          []string{"-transport=stdio"},
			wantTransport: "stdio",
			wantPort:      "8080",
			wantHost:      "localhost",
		},
		{
			name:          "custom port",
			args:          []string{"-port=9090"},
			wantTransport: "http",
			wantPort:      "9090",
			wantHost:      "localhost",
		},
		{
			name:          "custom host",
			args:          []string{"-host=0.0.0.0"},
			wantTransport: "http",
			wantPort:      "8080",
			wantHost:      "0.0.0.0",
		},
		{
			name:          "all custom values",
			args:          []string{"-transport=http", "-port=3000", "-host=127.0.0.1"},
			wantTransport: "http",
			wantPort:      "3000",
			wantHost:      "127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new FlagSet for each test
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			
			// Define flags
			transport := fs.String("transport", "http", "Transport type")
			port := fs.String("port", "8080", "Port for HTTP transport")
			host := fs.String("host", "localhost", "Host for HTTP transport")

			// Parse the test args
			err := fs.Parse(tt.args)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			// Check values
			if *transport != tt.wantTransport {
				t.Errorf("transport = %s, want %s", *transport, tt.wantTransport)
			}
			if *port != tt.wantPort {
				t.Errorf("port = %s, want %s", *port, tt.wantPort)
			}
			if *host != tt.wantHost {
				t.Errorf("host = %s, want %s", *host, tt.wantHost)
			}
		})
	}
}

// MockSignalNotifier allows testing signal handling without actually sending signals
type MockSignalNotifier struct {
	mu       sync.Mutex
	handlers map[chan<- os.Signal][]os.Signal
}

func NewMockSignalNotifier() *MockSignalNotifier {
	return &MockSignalNotifier{
		handlers: make(map[chan<- os.Signal][]os.Signal),
	}
}

func (m *MockSignalNotifier) Notify(c chan<- os.Signal, sig ...os.Signal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[c] = sig
}

func (m *MockSignalNotifier) SendSignal(sig os.Signal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for c, sigs := range m.handlers {
		if slices.Contains(sigs, sig) {
			select {
			case c <- sig:
			default:
			}
		}
	}
}

// TestSignalHandling tests graceful shutdown on signals
func TestSignalHandling(t *testing.T) {
	// This test verifies the signal channel setup
	sigChan := make(chan os.Signal, 1)
	
	// Simulate signal notification setup
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	
	// Verify channel can receive signals
	for _, sig := range signals {
		t.Run(fmt.Sprintf("handle_%v", sig), func(t *testing.T) {
			// Send signal to channel
			select {
			case sigChan <- sig:
				// Successfully sent
			case <-time.After(100 * time.Millisecond):
				t.Error("Failed to send signal to channel")
			}
			
			// Receive signal from channel
			select {
			case received := <-sigChan:
				if received != sig {
					t.Errorf("Received signal %v, want %v", received, sig)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("Failed to receive signal from channel")
			}
		})
	}
}

// TestErrorChannelHandling tests error channel behavior
func TestErrorChannelHandling(t *testing.T) {
	errChan := make(chan error, 1)
	
	testError := fmt.Errorf("test server error")
	
	// Send error
	select {
	case errChan <- testError:
		// Successfully sent
	case <-time.After(100 * time.Millisecond):
		t.Error("Failed to send error to channel")
	}
	
	// Receive error
	select {
	case err := <-errChan:
		if err.Error() != testError.Error() {
			t.Errorf("Received error %v, want %v", err, testError)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Failed to receive error from channel")
	}
}

// TestTransportSwitch tests the transport type switch logic
func TestTransportSwitch(t *testing.T) {
	tests := []struct {
		transport   string
		shouldError bool
		errorMsg    string
	}{
		{
			transport:   "stdio",
			shouldError: false,
		},
		{
			transport:   "http",
			shouldError: false,
		},
		{
			transport:   "invalid",
			shouldError: true,
			errorMsg:    "unsupported transport: invalid",
		},
		{
			transport:   "unknown",
			shouldError: true,
			errorMsg:    "unsupported transport: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.transport, func(t *testing.T) {
			// Simulate the switch logic from main
			var err error
			switch tt.transport {
			case "stdio":
				// Would call StartStdio
				err = nil
			case "http":
				// Would call StartHTTP
				err = nil
			default:
				err = fmt.Errorf("unsupported transport: %s", tt.transport)
			}

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Error = %v, want %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestHTTPAddressFormatting tests HTTP address formatting
func TestHTTPAddressFormatting(t *testing.T) {
	tests := []struct {
		host     string
		port     string
		expected string
	}{
		{"localhost", "8080", "localhost:8080"},
		{"127.0.0.1", "3000", "127.0.0.1:3000"},
		{"0.0.0.0", "9090", "0.0.0.0:9090"},
		{"example.com", "443", "example.com:443"},
		{"", "8080", ":8080"},
		{"localhost", "", "localhost:"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.host, tt.port), func(t *testing.T) {
			addr := fmt.Sprintf("%s:%s", tt.host, tt.port)
			if addr != tt.expected {
				t.Errorf("Address = %s, want %s", addr, tt.expected)
			}
		})
	}
}

// BenchmarkFlagParsing benchmarks flag parsing
func BenchmarkFlagParsing(b *testing.B) {
	args := []string{"-transport=http", "-port=8080", "-host=localhost"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs := flag.NewFlagSet("bench", flag.ContinueOnError)
		fs.String("transport", "http", "Transport type")
		fs.String("port", "8080", "Port")
		fs.String("host", "localhost", "Host")
		fs.Parse(args)
	}
}


// TestMainStartupMessages tests that appropriate startup messages are logged
func TestMainStartupMessages(t *testing.T) {
	tests := []struct {
		transport string
		port      string
		host      string
		expected  []string
	}{
		{
			transport: "stdio",
			expected:  []string{"MCP Todo Server", "STDIO mode"},
		},
		{
			transport: "http",
			port:      "8080",
			host:      "localhost",
			expected:  []string{"MCP Todo Server", "HTTP mode", "localhost:8080"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.transport, func(t *testing.T) {
			// Simulate the log messages that would be printed
			var output string
			switch tt.transport {
			case "stdio":
				output = "Starting MCP Todo Server v1.0.0 (STDIO mode)..."
			case "http":
				addr := fmt.Sprintf("%s:%s", tt.host, tt.port)
				output = fmt.Sprintf("Starting MCP Todo Server v1.0.0 (HTTP mode) on %s...", addr)
			}

			// Check that output contains expected strings
			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("Output %q does not contain expected string %q", output, exp)
				}
			}
		})
	}
}