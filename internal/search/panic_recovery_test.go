package search

import (
	"errors"
	"testing"
)

func TestSafeExecute(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		fn        func() error
		wantErr   bool
		errCheck  func(error) bool
	}{
		{
			name:      "successful operation",
			operation: "test-success",
			fn: func() error {
				return nil
			},
			wantErr: false,
		},
		{
			name:      "operation returns error",
			operation: "test-error",
			fn: func() error {
				return errors.New("test error")
			},
			wantErr:  true,
			errCheck: func(err error) bool {
				return err.Error() == "test error"
			},
		},
		{
			name:      "operation panics",
			operation: "test-panic",
			fn: func() error {
				panic("test panic")
			},
			wantErr: false, // SafeExecute recovers from panic without setting error
		},
		{
			name:      "operation panics with error",
			operation: "test-panic-error",
			fn: func() error {
				panic(errors.New("panic with error"))
			},
			wantErr: false, // SafeExecute recovers without setting error
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SafeExecute(tt.operation, tt.fn)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeExecute() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if tt.errCheck != nil && err != nil && !tt.errCheck(err) {
				t.Errorf("SafeExecute() error = %v, expected different error", err)
			}
		})
	}
}

func TestSafeExecuteWithResult(t *testing.T) {
	t.Run("successful operation", func(t *testing.T) {
		result, err := SafeExecuteWithResult("test-success", func() (string, error) {
			return "success", nil
		})
		
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result != "success" {
			t.Errorf("Expected result 'success', got '%s'", result)
		}
	})
	
	t.Run("operation returns error", func(t *testing.T) {
		result, err := SafeExecuteWithResult("test-error", func() (int, error) {
			return 0, errors.New("test error")
		})
		
		if err == nil || err.Error() != "test error" {
			t.Errorf("Expected 'test error', got %v", err)
		}
		if result != 0 {
			t.Errorf("Expected zero value result, got %d", result)
		}
	})
	
	t.Run("operation panics with string", func(t *testing.T) {
		result, err := SafeExecuteWithResult("test-panic", func() (string, error) {
			panic("test panic")
		})
		
		if err == nil {
			t.Errorf("Expected error from panic, got nil")
		} else if err.Error() != "operation test-panic panicked: test panic" {
			t.Errorf("Expected error 'operation test-panic panicked: test panic', got %v", err)
		}
		if result != "" {
			t.Errorf("Expected empty string result, got '%s'", result)
		}
	})
	
	t.Run("operation panics with error", func(t *testing.T) {
		result, err := SafeExecuteWithResult("test-panic-error", func() (int, error) {
			panic(errors.New("panic with error"))
		})
		
		if err == nil {
			t.Errorf("Expected error from panic, got nil")
		} else if err.Error() != "operation test-panic-error panicked: panic with error" {
			t.Errorf("Expected error 'operation test-panic-error panicked: panic with error', got %v", err)
		}
		if result != 0 {
			t.Errorf("Expected zero value result, got %d", result)
		}
	})
	
	t.Run("operation panics with nil", func(t *testing.T) {
		result, err := SafeExecuteWithResult("test-panic-nil", func() (bool, error) {
			panic(nil)
		})
		
		if err == nil {
			t.Errorf("Expected error from panic, got nil")
		} else if err.Error() != "operation test-panic-nil panicked: panic called with nil argument" {
			t.Errorf("Expected error 'operation test-panic-nil panicked: panic called with nil argument', got %v", err)
		}
		if result != false {
			t.Errorf("Expected false result, got %v", result)
		}
	})
	
	t.Run("complex type zero value", func(t *testing.T) {
		type ComplexType struct {
			Name  string
			Value int
			Items []string
		}
		
		result, err := SafeExecuteWithResult("test-complex-panic", func() (ComplexType, error) {
			panic("complex panic")
		})
		
		if err == nil {
			t.Error("Expected error from panic")
		} else if err.Error() != "operation test-complex-panic panicked: complex panic" {
			t.Errorf("Expected error 'operation test-complex-panic panicked: complex panic', got %v", err)
		}
		
		// Verify zero value is returned
		if result.Name != "" || result.Value != 0 || len(result.Items) != 0 {
			t.Errorf("Expected zero value for ComplexType, got %+v", result)
		}
	})
}

func TestSafeExecute_MultipleOperations(t *testing.T) {
	// Test that multiple SafeExecute calls don't interfere with each other
	errorsChan := make(chan error, 3)
	
	// Run multiple operations concurrently
	go func() {
		err := SafeExecute("op1", func() error {
			return nil
		})
		errorsChan <- err
	}()
	
	go func() {
		err := SafeExecute("op2", func() error {
			panic("panic in op2")
		})
		errorsChan <- err
	}()
	
	go func() {
		err := SafeExecute("op3", func() error {
			return errors.New("error in op3")
		})
		errorsChan <- err
	}()
	
	// Collect results
	var nilCount, errorCount int
	for i := 0; i < 3; i++ {
		err := <-errorsChan
		if err == nil {
			nilCount++
		} else {
			errorCount++
		}
	}
	
	// We expect 2 nil results (successful op and recovered panic) and 1 error
	if nilCount != 2 {
		t.Errorf("Expected 2 nil results, got %d", nilCount)
	}
	if errorCount != 1 {
		t.Errorf("Expected 1 error result, got %d", errorCount)
	}
}