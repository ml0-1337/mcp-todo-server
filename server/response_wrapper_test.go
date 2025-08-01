package server

import (
	"net/http"
	"testing"
)

func TestResponseWrapper(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		wrapper := newResponseWrapper()

		// Test Header() method
		header := wrapper.Header()
		header.Set("X-Test-Header", "test-value")

		// Test WriteHeader() method
		wrapper.WriteHeader(http.StatusCreated)

		// Verify status code was captured
		if wrapper.statusCode != http.StatusCreated {
			t.Errorf("Expected status code 201, got %d", wrapper.statusCode)
		}

		// Test Write() method
		testData := []byte("test response body")
		n, err := wrapper.Write(testData)
		if err != nil {
			t.Errorf("Unexpected error writing: %v", err)
		}
		if n != len(testData) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
		}

		// Verify the body was written
		if wrapper.body.String() != string(testData) {
			t.Errorf("Expected body '%s', got '%s'", string(testData), wrapper.body.String())
		}

		// Verify header was set
		if wrapper.headers.Get("X-Test-Header") != "test-value" {
			t.Errorf("Expected header 'test-value', got '%s'", wrapper.headers.Get("X-Test-Header"))
		}
	})

	t.Run("default status code", func(t *testing.T) {
		wrapper := newResponseWrapper()

		// Write without calling WriteHeader
		wrapper.Write([]byte("test"))

		// Should default to 200 OK
		if wrapper.statusCode != http.StatusOK {
			t.Errorf("Expected default status code 200, got %d", wrapper.statusCode)
		}

		// Should mark header as written
		if !wrapper.headerWritten {
			t.Error("Expected headerWritten to be true after Write")
		}
	})

	t.Run("multiple writes", func(t *testing.T) {
		wrapper := newResponseWrapper()

		// Multiple writes
		data1 := []byte("first ")
		data2 := []byte("second ")
		data3 := []byte("third")

		n1, _ := wrapper.Write(data1)
		n2, _ := wrapper.Write(data2)
		n3, _ := wrapper.Write(data3)

		totalBytes := n1 + n2 + n3
		expectedTotal := len(data1) + len(data2) + len(data3)

		if totalBytes != expectedTotal {
			t.Errorf("Expected total %d bytes written, got %d", expectedTotal, totalBytes)
		}

		// Verify all data was concatenated
		expectedBody := "first second third"
		if wrapper.body.String() != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, wrapper.body.String())
		}
	})

	t.Run("write header multiple times", func(t *testing.T) {
		wrapper := newResponseWrapper()

		// First WriteHeader should set status
		wrapper.WriteHeader(http.StatusCreated)
		if wrapper.statusCode != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", wrapper.statusCode)
		}

		// Second WriteHeader should be ignored
		wrapper.WriteHeader(http.StatusBadRequest)
		if wrapper.statusCode != http.StatusCreated {
			t.Errorf("Expected status to remain 201, got %d", wrapper.statusCode)
		}
	})

	t.Run("concurrent writes", func(t *testing.T) {
		wrapper := newResponseWrapper()

		// Test concurrent access (mutex should protect)
		done := make(chan bool, 2)

		go func() {
			for i := 0; i < 100; i++ {
				wrapper.Write([]byte("A"))
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 100; i++ {
				wrapper.Write([]byte("B"))
			}
			done <- true
		}()

		<-done
		<-done

		// Should have 200 characters total
		if len(wrapper.body.String()) != 200 {
			t.Errorf("Expected 200 characters, got %d", len(wrapper.body.String()))
		}
	})

	t.Run("flush method", func(t *testing.T) {
		wrapper := newResponseWrapper()

		// Test Flush() method - should not panic
		wrapper.Flush()

		// Write some data
		wrapper.Write([]byte("test"))

		// Flush again
		wrapper.Flush()

		// Data should still be there
		if wrapper.body.String() != "test" {
			t.Errorf("Expected body 'test', got '%s'", wrapper.body.String())
		}
	})

	t.Run("close notify", func(t *testing.T) {
		wrapper := newResponseWrapper()

		// Test CloseNotify() method
		ch := wrapper.CloseNotify()

		// Should return a non-nil channel
		if ch == nil {
			t.Error("Expected non-nil channel from CloseNotify")
		}

		// Channel should not be closed initially
		select {
		case <-ch:
			t.Error("Channel should not be closed initially")
		default:
			// Good, channel is not closed
		}
	})

	t.Run("write after WriteHeader", func(t *testing.T) {
		wrapper := newResponseWrapper()

		// Set status first
		wrapper.WriteHeader(http.StatusAccepted)

		// Then write data
		data := []byte("after header")
		n, err := wrapper.Write(data)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
		}

		// Verify status remained the same
		if wrapper.statusCode != http.StatusAccepted {
			t.Errorf("Expected status to remain 202, got %d", wrapper.statusCode)
		}

		// Verify data was written
		if wrapper.body.String() != string(data) {
			t.Errorf("Expected body '%s', got '%s'", string(data), wrapper.body.String())
		}
	})

	t.Run("header modifications", func(t *testing.T) {
		wrapper := newResponseWrapper()

		// Add multiple headers
		wrapper.Header().Set("Content-Type", "application/json")
		wrapper.Header().Add("X-Custom-Header", "value1")
		wrapper.Header().Add("X-Custom-Header", "value2")

		// Write to trigger header written flag
		wrapper.Write([]byte("{}"))

		// Verify headers
		if ct := wrapper.headers.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", ct)
		}

		customHeaders := wrapper.headers["X-Custom-Header"]
		if len(customHeaders) != 2 {
			t.Errorf("Expected 2 custom headers, got %d", len(customHeaders))
		}
		if customHeaders[0] != "value1" || customHeaders[1] != "value2" {
			t.Errorf("Expected custom headers ['value1', 'value2'], got %v", customHeaders)
		}
	})
}
