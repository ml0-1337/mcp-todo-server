package search

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker protects against cascading failures
type CircuitBreaker struct {
	mu           sync.RWMutex
	state        CircuitState
	failureCount int
	lastFailTime time.Time
	
	// Configuration
	failureThreshold int
	timeout          time.Duration
	resetTimeout     time.Duration
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, timeout, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitClosed,
		failureThreshold: failureThreshold,
		timeout:          timeout,
		resetTimeout:     resetTimeout,
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	cb.mu.RLock()
	state := cb.state
	failureCount := cb.failureCount
	lastFailTime := cb.lastFailTime
	cb.mu.RUnlock()

	// Check if circuit is open
	if state == CircuitOpen {
		if time.Since(lastFailTime) < cb.resetTimeout {
			return fmt.Errorf("circuit breaker is open (failures: %d)", failureCount)
		}
		// Try to transition to half-open
		cb.mu.Lock()
		if cb.state == CircuitOpen {
			cb.state = CircuitHalfOpen
		}
		cb.mu.Unlock()
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(ctx, cb.timeout)
	defer cancel()

	errCh := make(chan error, 1)
	done := make(chan struct{})
	defer close(done)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errCh <- fmt.Errorf("function panicked: %v", r)
			}
		}()

		select {
		case <-done:
			return
		default:
			err := fn()
			select {
			case errCh <- err:
			case <-done:
				return
			}
		}
	}()

	select {
	case err := <-errCh:
		if err != nil {
			cb.recordFailure()
			return err
		}
		cb.recordSuccess()
		return nil
	case <-ctx.Done():
		cb.recordFailure()
		return fmt.Errorf("operation timed out after %v", cb.timeout)
	}
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.failureCount >= cb.failureThreshold {
		cb.state = CircuitOpen
	}
}

// recordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	cb.state = CircuitClosed
}

// GetState returns the current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailureCount returns the current failure count
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failureCount
}