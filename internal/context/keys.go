package context

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// WorkingDirectoryKey is the context key for working directory
	WorkingDirectoryKey ContextKey = "working-directory"
	// SessionIDKey is the context key for session ID
	SessionIDKey ContextKey = "session-id"
)
