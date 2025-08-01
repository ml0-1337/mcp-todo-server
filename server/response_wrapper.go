package server

import (
	"bytes"
	"net/http"
	"sync"
)

// responseWrapper captures HTTP response data safely
type responseWrapper struct {
	statusCode    int
	headers       http.Header
	body          bytes.Buffer
	headerWritten bool
	mu            sync.Mutex
}

// newResponseWrapper creates a new response wrapper
func newResponseWrapper() *responseWrapper {
	return &responseWrapper{
		headers:    make(http.Header),
		statusCode: http.StatusOK,
	}
}

// Header returns the response headers
func (w *responseWrapper) Header() http.Header {
	return w.headers
}

// WriteHeader writes the status code
func (w *responseWrapper) WriteHeader(code int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.headerWritten {
		w.statusCode = code
		w.headerWritten = true
	}
}

// Write writes data to the response body
func (w *responseWrapper) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.headerWritten {
		w.headerWritten = true
	}

	return w.body.Write(data)
}

// Flush implements http.Flusher
func (w *responseWrapper) Flush() {
	// No-op for captured responses
}

// CloseNotify implements http.CloseNotifier (deprecated but may be used)
func (w *responseWrapper) CloseNotify() <-chan bool {
	ch := make(chan bool)
	return ch
}
