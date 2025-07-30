package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	
	"github.com/gofrs/flock"
)

func TestServerLock_NewServerLock(t *testing.T) {
	port := "8080"
	lock, err := NewServerLock(port)
	
	if err != nil {
		t.Fatalf("NewServerLock failed: %v", err)
	}
	
	if lock == nil {
		t.Fatal("Expected lock to be non-nil")
	}
	
	if lock.IsLocked() {
		t.Error("New lock should not be locked initially")
	}
	
	// Verify lock path contains port
	lockPath := lock.GetLockPath()
	if lockPath == "" {
		t.Error("Lock path should not be empty")
	}
}

func TestServerLock_TryLock(t *testing.T) {
	port := "8081"
	lock, err := NewServerLock(port)
	if err != nil {
		t.Fatalf("NewServerLock failed: %v", err)
	}
	defer lock.Unlock()
	
	// First lock should succeed
	err = lock.TryLock()
	if err != nil {
		t.Fatalf("First TryLock should succeed: %v", err)
	}
	
	if !lock.IsLocked() {
		t.Error("Lock should be locked after TryLock")
	}
	
	// Second lock on same instance should be idempotent
	err = lock.TryLock()
	if err != nil {
		t.Errorf("Second TryLock on same instance should succeed: %v", err)
	}
}

func TestServerLock_PreventDuplicates(t *testing.T) {
	port := "8082"
	
	// Clean up any stale lock files first
	lockPath := filepath.Join(os.TempDir(), fmt.Sprintf("mcp-todo-server-%s.lock", port))
	os.Remove(lockPath) // Ignore error, file might not exist
	
	// Create first lock
	lock1, err := NewServerLock(port)
	if err != nil {
		t.Fatalf("NewServerLock failed: %v", err)
	}
	defer lock1.Unlock()
	
	// Create second lock for same port
	lock2, err := NewServerLock(port)
	if err != nil {
		t.Fatalf("NewServerLock failed: %v", err)
	}
	defer lock2.Unlock()
	
	// Debug: check if paths are the same
	if lock1.GetLockPath() != lock2.GetLockPath() {
		t.Fatalf("Lock paths should be the same: %s vs %s", lock1.GetLockPath(), lock2.GetLockPath())
	}
	
	// Acquire first lock
	err = lock1.TryLock()
	if err != nil {
		t.Fatalf("First lock should succeed: %v", err)
	}
	
	// Second lock should fail
	err = lock2.TryLock()
	if err == nil {
		t.Fatalf("Second lock should fail when first is held. Lock1 locked: %v, Lock2 locked: %v", lock1.IsLocked(), lock2.IsLocked())
	}
	t.Logf("Expected error from second lock: %v", err)
	
	// Error should mention another instance
	if err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestServerLock_Unlock(t *testing.T) {
	port := "8083"
	lock, err := NewServerLock(port)
	if err != nil {
		t.Fatalf("NewServerLock failed: %v", err)
	}
	
	// Lock first
	err = lock.TryLock()
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	
	// Unlock
	err = lock.Unlock()
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
	
	if lock.IsLocked() {
		t.Error("Lock should not be locked after Unlock")
	}
	
	// Should be able to unlock multiple times safely
	err = lock.Unlock()
	if err != nil {
		t.Errorf("Multiple Unlock should be safe: %v", err)
	}
}

func TestServerLock_ReleaseAfterUnlock(t *testing.T) {
	port := "8084"
	
	// Create and acquire first lock
	lock1, err := NewServerLock(port)
	if err != nil {
		t.Fatalf("NewServerLock failed: %v", err)
	}
	
	err = lock1.TryLock()
	if err != nil {
		t.Fatalf("First lock should succeed: %v", err)
	}
	
	// Release first lock
	err = lock1.Unlock()
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
	
	// Second lock should now succeed
	lock2, err := NewServerLock(port)
	if err != nil {
		t.Fatalf("NewServerLock failed: %v", err)
	}
	defer lock2.Unlock()
	
	err = lock2.TryLock()
	if err != nil {
		t.Fatalf("Second lock should succeed after first is released: %v", err)
	}
}

func TestServerLock_DeadProcessCleanup(t *testing.T) {
	// This test simulates what happens when a process dies without cleanup
	port := "8085"
	
	// Create a lock and manually write a fake PID to simulate dead process
	lock, err := NewServerLock(port)
	if err != nil {
		t.Fatalf("NewServerLock failed: %v", err)
	}
	
	lockPath := lock.GetLockPath()
	
	// Write a fake PID that doesn't exist (very high number)
	fakeContent := "999999\n"
	err = os.WriteFile(lockPath, []byte(fakeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write fake lock file: %v", err)
	}
	
	// TryLock should succeed by cleaning up the dead process lock
	err = lock.TryLock()
	if err != nil {
		t.Fatalf("TryLock should succeed when cleaning up dead process: %v", err)
	}
	
	defer lock.Unlock()
	
	if !lock.IsLocked() {
		t.Error("Lock should be acquired after cleanup")
	}
}

func TestServerLock_ConcurrentAccess(t *testing.T) {
	port := "8086"
	numGoroutines := 10
	successCount := 0
	errorCount := 0
	
	results := make(chan bool, numGoroutines)
	startChan := make(chan struct{})
	
	// Launch multiple goroutines trying to acquire the same lock
	for i := 0; i < numGoroutines; i++ {
		go func() {
			// Wait for all goroutines to be ready
			<-startChan
			
			lock, err := NewServerLock(port)
			if err != nil {
				results <- false
				return
			}
			
			// Try to acquire lock and hold it briefly
			err = lock.TryLock()
			success := err == nil
			
			if success {
				// Hold the lock for a moment to ensure overlap
				// time.Sleep(10 * time.Millisecond) // Not using time, keep it simple
				for i := 0; i < 1000000; i++ {
					// Small busy wait to ensure some overlap
				}
				lock.Unlock()
			}
			
			results <- success
		}()
	}
	
	// Start all goroutines at once
	close(startChan)
	
	// Collect results
	for i := 0; i < numGoroutines; i++ {
		success := <-results
		if success {
			successCount++
		} else {
			errorCount++
		}
	}
	
	// Exactly one should succeed
	if successCount != 1 {
		t.Errorf("Expected exactly 1 success, got %d", successCount)
	}
	
	if errorCount != numGoroutines-1 {
		t.Errorf("Expected %d errors, got %d", numGoroutines-1, errorCount)
	}
}

func TestServerLock_DifferentPorts(t *testing.T) {
	// Different ports should not interfere with each other
	lock1, err := NewServerLock("8087")
	if err != nil {
		t.Fatalf("NewServerLock for port 8087 failed: %v", err)
	}
	defer lock1.Unlock()
	
	lock2, err := NewServerLock("8088")
	if err != nil {
		t.Fatalf("NewServerLock for port 8088 failed: %v", err)
	}
	defer lock2.Unlock()
	
	// Both should be able to acquire locks
	err = lock1.TryLock()
	if err != nil {
		t.Fatalf("Lock for port 8087 should succeed: %v", err)
	}
	
	err = lock2.TryLock()
	if err != nil {
		t.Fatalf("Lock for port 8088 should succeed: %v", err)
	}
	
	// Both should be locked
	if !lock1.IsLocked() {
		t.Error("Lock1 should be locked")
	}
	
	if !lock2.IsLocked() {
		t.Error("Lock2 should be locked")
	}
}

func TestFlockDirectly(t *testing.T) {
	// Test the flock library directly to understand its behavior  
	tempDir := os.TempDir()
	lockPath := tempDir + "/test-flock-direct.lock"
	defer os.Remove(lockPath)
	
	lock1 := flock.New(lockPath)
	lock2 := flock.New(lockPath)
	
	// First lock should succeed
	locked, err := lock1.TryLock()
	if err != nil {
		t.Fatalf("First TryLock should succeed: %v", err)
	}
	if !locked {
		t.Fatal("First TryLock should return true")
	}
	defer lock1.Unlock()
	
	// Check if lock file exists
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("Lock file should exist after TryLock")
	}
	
	// Second lock should fail
	locked2, err := lock2.TryLock()
	if err != nil {
		t.Fatalf("Second TryLock should not error, but should return false: %v", err)
	}
	if locked2 {
		defer lock2.Unlock()
		t.Fatal("Second TryLock should return false when first is held")
	}
	
	t.Logf("Successfully tested flock behavior - second lock correctly failed")
}