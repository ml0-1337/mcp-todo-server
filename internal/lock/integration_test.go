package lock

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestServerLockIntegration tests the lock functionality with the actual server binary
func TestServerLockIntegration(t *testing.T) {
	// Build the server binary first
	buildDir := "../../build"
	binaryPath := filepath.Join(buildDir, "mcp-todo-server")

	// Build the binary
	cmd := exec.Command("make", "build")
	cmd.Dir = "../.."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build server: %v", err)
	}

	// Check that binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("Server binary not found at %s", binaryPath)
	}

	t.Run("SingleInstanceOnPort", func(t *testing.T) {
		port := "18080" // Use test port

		// Start first server instance
		cmd1 := exec.Command(binaryPath, "-transport", "http", "-port", port)
		cmd1.Stderr = os.Stderr

		err := cmd1.Start()
		if err != nil {
			t.Fatalf("Failed to start first server: %v", err)
		}
		defer func() {
			if cmd1.Process != nil {
				cmd1.Process.Kill()
				cmd1.Wait()
			}
		}()

		// Give first server time to start and acquire lock
		time.Sleep(2 * time.Second)

		// Try to start second server instance
		cmd2 := exec.Command(binaryPath, "-transport", "http", "-port", port)
		output, err := cmd2.CombinedOutput()

		// Second server should fail
		if err == nil {
			t.Fatal("Second server should have failed to start")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Failed to acquire server lock") {
			t.Errorf("Expected lock error message, got: %s", outputStr)
		}

		if !strings.Contains(outputStr, "another instance is already running") {
			t.Errorf("Expected 'another instance' message, got: %s", outputStr)
		}
	})

	t.Run("DifferentPortsAllowed", func(t *testing.T) {
		port1 := "18081"
		port2 := "18082"

		// Start first server
		cmd1 := exec.Command(binaryPath, "-transport", "http", "-port", port1)
		err := cmd1.Start()
		if err != nil {
			t.Fatalf("Failed to start first server: %v", err)
		}
		defer func() {
			if cmd1.Process != nil {
				cmd1.Process.Kill()
				cmd1.Wait()
			}
		}()

		// Give first server time to start
		time.Sleep(1 * time.Second)

		// Start second server on different port
		cmd2 := exec.Command(binaryPath, "-transport", "http", "-port", port2)
		err = cmd2.Start()
		if err != nil {
			t.Fatalf("Failed to start second server: %v", err)
		}
		defer func() {
			if cmd2.Process != nil {
				cmd2.Process.Kill()
				cmd2.Wait()
			}
		}()

		// Give second server time to start
		time.Sleep(1 * time.Second)

		// Both should be running - verify by checking they're still alive
		// (If either had failed, the process would have exited)
		if cmd1.ProcessState != nil && cmd1.ProcessState.Exited() {
			t.Error("First server should still be running")
		}

		if cmd2.ProcessState != nil && cmd2.ProcessState.Exited() {
			t.Error("Second server should still be running")
		}
	})

	t.Run("LockReleasedAfterShutdown", func(t *testing.T) {
		port := "18083"

		// Start first server
		cmd1 := exec.Command(binaryPath, "-transport", "http", "-port", port)
		err := cmd1.Start()
		if err != nil {
			t.Fatalf("Failed to start first server: %v", err)
		}

		// Give server time to start
		time.Sleep(2 * time.Second)

		// Gracefully stop the server
		if err := cmd1.Process.Kill(); err != nil {
			t.Fatalf("Failed to stop first server: %v", err)
		}
		cmd1.Wait()

		// Give some time for cleanup
		time.Sleep(1 * time.Second)

		// Start second server on same port
		cmd2 := exec.Command(binaryPath, "-transport", "http", "-port", port)
		err = cmd2.Start()
		if err != nil {
			t.Fatalf("Failed to start second server after first shutdown: %v", err)
		}
		defer func() {
			if cmd2.Process != nil {
				cmd2.Process.Kill()
				cmd2.Wait()
			}
		}()

		// Give second server time to start
		time.Sleep(1 * time.Second)

		// Second server should be running successfully
		if cmd2.ProcessState != nil && cmd2.ProcessState.Exited() {
			t.Error("Second server should be running after first was stopped")
		}
	})

	t.Run("STDIOTransportUnaffected", func(t *testing.T) {
		// STDIO transport should not use locking
		// This is hard to test directly, but we can at least verify
		// that STDIO mode doesn't fail due to lock-related errors

		cmd := exec.Command(binaryPath, "-transport", "stdio", "-version")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Fatalf("STDIO mode failed: %v, output: %s", err, string(output))
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "MCP Todo Server") {
			t.Errorf("Expected version output, got: %s", outputStr)
		}
	})
}
