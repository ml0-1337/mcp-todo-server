package search

import (
	"fmt"
	"os"
	"runtime"
)

// SafeExecute wraps a function with panic recovery
func SafeExecute(operation string, fn func() error) error {
	defer func() {
		if r := recover(); r != nil {
			// Log the panic with stack trace
			stackBuf := make([]byte, 4096)
			stackSize := runtime.Stack(stackBuf, false)
			
			fmt.Fprintf(os.Stderr, "PANIC in %s: %v\n", operation, r)
			fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", stackBuf[:stackSize])
		}
	}()
	
	return fn()
}

// SafeExecuteWithResult wraps a function with panic recovery and returns both result and error
func SafeExecuteWithResult[T any](operation string, fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Log the panic with stack trace
			stackBuf := make([]byte, 4096)
			stackSize := runtime.Stack(stackBuf, false)
			
			fmt.Fprintf(os.Stderr, "PANIC in %s: %v\n", operation, r)
			fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", stackBuf[:stackSize])
			
			// Return zero value and error
			var zero T
			result = zero
			err = fmt.Errorf("operation %s panicked: %v", operation, r)
		}
	}()
	
	result, err = fn()
	return result, err
}