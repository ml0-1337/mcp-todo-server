package search

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestCircuitBreaker_GetState(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond, 500*time.Millisecond)
	
	// Initial state should be closed
	if state := cb.GetState(); state != CircuitClosed {
		t.Errorf("Expected initial state to be CircuitClosed, got %v", state)
	}
	
	// Cause failures to open the circuit
	ctx := context.Background()
	testErr := errors.New("test error")
	for i := 0; i < 3; i++ {
		err := cb.Execute(ctx, func() error {
			return testErr
		})
		if err != testErr {
			t.Errorf("Expected test error, got %v", err)
		}
	}
	
	// Circuit should now be open
	if state := cb.GetState(); state != CircuitOpen {
		t.Errorf("Expected state to be CircuitOpen after failures, got %v", state)
	}
	
	// Wait for resetTimeout to transition to half-open
	time.Sleep(550 * time.Millisecond)
	
	// Make a request to trigger state check
	cb.Execute(ctx, func() error {
		return nil
	})
	
	// Should be half-open or closed depending on execution
	state := cb.GetState()
	if state != CircuitHalfOpen && state != CircuitClosed {
		t.Errorf("Expected state to be CircuitHalfOpen or CircuitClosed, got %v", state)
	}
}

func TestCircuitBreaker_GetFailureCount(t *testing.T) {
	cb := NewCircuitBreaker(5, 100*time.Millisecond, 500*time.Millisecond)
	ctx := context.Background()
	
	// Initial failure count should be 0
	if count := cb.GetFailureCount(); count != 0 {
		t.Errorf("Expected initial failure count to be 0, got %d", count)
	}
	
	// Cause some failures
	testErr := errors.New("test error")
	for i := 0; i < 3; i++ {
		cb.Execute(ctx, func() error {
			return testErr
		})
	}
	
	// Check failure count
	if count := cb.GetFailureCount(); count != 3 {
		t.Errorf("Expected failure count to be 3, got %d", count)
	}
	
	// Successful execution should reset count
	cb.Execute(ctx, func() error {
		return nil
	})
	
	if count := cb.GetFailureCount(); count != 0 {
		t.Errorf("Expected failure count to be reset to 0, got %d", count)
	}
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cb := NewCircuitBreaker(2, 100*time.Millisecond, 500*time.Millisecond)
	ctx := context.Background()
	
	// Verify initial state
	if state := cb.GetState(); state != CircuitClosed {
		t.Fatalf("Expected initial state CircuitClosed, got %v", state)
	}
	
	// Test Closed -> Open transition
	testErr := errors.New("failure")
	for i := 0; i < 2; i++ {
		err := cb.Execute(ctx, func() error {
			return testErr
		})
		if err != testErr {
			t.Errorf("Expected error %v, got %v", testErr, err)
		}
	}
	
	if state := cb.GetState(); state != CircuitOpen {
		t.Errorf("Expected state CircuitOpen after failures, got %v", state)
	}
	
	// Test that requests fail fast when open
	start := time.Now()
	err := cb.Execute(ctx, func() error {
		return nil
	})
	elapsed := time.Since(start)
	
	if err == nil || !contains(err.Error(), "circuit breaker is open") {
		t.Errorf("Expected 'circuit breaker is open' error, got %v", err)
	}
	if elapsed > 10*time.Millisecond {
		t.Errorf("Expected fast fail, but took %v", elapsed)
	}
	
	// Wait for resetTimeout to allow half-open
	time.Sleep(550 * time.Millisecond)
	
	// Test Open -> Half-Open -> Closed transition (successful request)
	err = cb.Execute(ctx, func() error {
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected successful execution, got error %v", err)
	}
	if state := cb.GetState(); state != CircuitClosed {
		t.Errorf("Expected state CircuitClosed after successful half-open, got %v", state)
	}
	
	// Reset and test Half-Open -> Open transition (failed request)
	// First, open the circuit again
	for i := 0; i < 2; i++ {
		cb.Execute(ctx, func() error {
			return testErr
		})
	}
	
	if state := cb.GetState(); state != CircuitOpen {
		t.Fatalf("Failed to reopen circuit for half-open test")
	}
	
	// Wait for resetTimeout
	time.Sleep(550 * time.Millisecond)
	
	// Fail in half-open state
	err = cb.Execute(ctx, func() error {
		return errors.New("half-open failure")
	})
	
	if state := cb.GetState(); state != CircuitOpen {
		t.Errorf("Expected state CircuitOpen after half-open failure, got %v", state)
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker(10, 100*time.Millisecond, 500*time.Millisecond)
	ctx := context.Background()
	
	// Run concurrent operations
	done := make(chan bool, 20)
	
	// 10 goroutines causing failures
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			err := cb.Execute(ctx, func() error {
				return fmt.Errorf("error from goroutine %d", id)
			})
			
			if err == nil {
				t.Errorf("Goroutine %d: expected error", id)
			}
		}(i)
	}
	
	// 10 goroutines with successful operations
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			err := cb.Execute(ctx, func() error {
				return nil
			})
			
			// May succeed or fail depending on circuit state
			_ = err
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
	
	// Circuit should be open due to failures
	state := cb.GetState()
	if state != CircuitOpen {
		t.Logf("Warning: Expected CircuitOpen state, got %v (may be timing dependent)", state)
	}
	
	// Failure count should be reasonable
	count := cb.GetFailureCount()
	if count > 10 {
		t.Errorf("Failure count %d exceeds maximum possible failures", count)
	}
}

func TestCircuitBreaker_TimeoutBehavior(t *testing.T) {
	// Short timeouts for faster test
	cb := NewCircuitBreaker(1, 50*time.Millisecond, 100*time.Millisecond)
	ctx := context.Background()
	
	// Open the circuit
	cb.Execute(ctx, func() error {
		return errors.New("open circuit")
	})
	
	if state := cb.GetState(); state != CircuitOpen {
		t.Fatalf("Failed to open circuit")
	}
	
	// Verify it stays open during timeout
	time.Sleep(50 * time.Millisecond)
	if state := cb.GetState(); state != CircuitOpen {
		t.Errorf("Circuit should remain open during timeout period")
	}
	
	// Wait past resetTimeout (100ms total)
	time.Sleep(60 * time.Millisecond)
	
	// Next execution should attempt (half-open)
	executed := false
	cb.Execute(ctx, func() error {
		executed = true
		return nil
	})
	
	if !executed {
		t.Error("Expected execution in half-open state")
	}
	
	// Should be closed after successful execution
	if state := cb.GetState(); state != CircuitClosed {
		t.Errorf("Expected CircuitClosed after successful half-open execution, got %v", state)
	}
}

func TestCircuitBreaker_ContextCancellation(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second, 5*time.Second)
	
	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	err := cb.Execute(ctx, func() error {
		t.Error("Function should not execute with cancelled context")
		return nil
	})
	
	if err == nil {
		t.Error("Expected error due to cancelled context")
	}
	
	// Circuit should remain closed (no failure recorded for context cancellation)
	if state := cb.GetState(); state != CircuitClosed {
		t.Errorf("Expected circuit to remain closed, got %v", state)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}