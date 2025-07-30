package lock

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

// ServerLock manages file-based locking to prevent multiple server instances
type ServerLock struct {
	flock  *flock.Flock
	locked bool
}

// NewServerLock creates a new server lock for the given port
func NewServerLock(port string) (*ServerLock, error) {
	// Create lock file in system temp directory with port-specific name
	lockPath := filepath.Join(os.TempDir(), fmt.Sprintf("mcp-todo-server-%s.lock", port))

	fl := flock.New(lockPath)

	return &ServerLock{
		flock:  fl,
		locked: false,
	}, nil
}

// TryLock attempts to acquire the lock
func (sl *ServerLock) TryLock() error {
	if sl.locked {
		return nil // Already locked
	}

	locked, err := sl.flock.TryLock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		return fmt.Errorf("another instance is already running on this port")
	}

	sl.locked = true
	return nil
}

// Unlock releases the lock
func (sl *ServerLock) Unlock() error {
	if !sl.locked {
		return nil // Not locked
	}

	err := sl.flock.Unlock()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	sl.locked = false
	return nil
}

// IsLocked returns whether the lock is currently held
func (sl *ServerLock) IsLocked() bool {
	return sl.locked
}

// GetLockPath returns the path to the lock file
func (sl *ServerLock) GetLockPath() string {
	return sl.flock.Path()
}
