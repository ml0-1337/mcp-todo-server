package search

import (
	"github.com/blevesearch/bleve/v2"
)

// HealthCheck returns comprehensive health status of the search engine
func (e *Engine) HealthCheck() map[string]interface{} {
	health := make(map[string]interface{})

	// Basic engine info
	health["status"] = "healthy"
	health["index_path"] = e.lock.GetLockPath() // Derived from lock path
	health["todo_path"] = e.basePath

	// Document count
	if count, err := e.index.DocCount(); err == nil {
		health["document_count"] = count
	} else {
		health["document_count"] = uint64(0)
		health["document_count_error"] = err.Error()
		health["status"] = "degraded"
	}

	// Check circuit breaker state
	cbState := e.circuitBreaker.GetState()
	health["circuit_breaker_state"] = getCircuitBreakerStateName(cbState)
	health["circuit_breaker_failures"] = e.circuitBreaker.GetFailureCount()

	// Check index lock status
	if e.lock != nil {
		health["index_locked"] = e.lock.IsLocked()
		health["lock_path"] = e.lock.GetLockPath()
	} else {
		health["index_locked"] = false
	}

	// Test basic index functionality
	testQuery := bleve.NewMatchAllQuery()
	testRequest := bleve.NewSearchRequest(testQuery)
	testRequest.Size = 1

	_, err := e.index.Search(testRequest)
	if err != nil {
		health["status"] = "unhealthy"
		health["index_error"] = err.Error()
		health["index_healthy"] = false
	} else {
		health["index_healthy"] = true
	}

	return health
}

// Helper function to get circuit breaker state name
func getCircuitBreakerStateName(state CircuitState) string {
	switch state {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
