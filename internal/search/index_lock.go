package search

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

// IndexLock provides file-based locking for Bleve indexes to prevent concurrent access
type IndexLock struct {
	flock    *flock.Flock
	lockPath string
}

// NewIndexLock creates a new index lock for the given index path
func NewIndexLock(indexPath string) *IndexLock {
	lockPath := indexPath + ".lock"
	
	// Ensure the directory for the lock file exists
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create lock directory: %v\n", err)
	}
	
	return &IndexLock{
		flock:    flock.New(lockPath),
		lockPath: lockPath,
	}
}

// TryLock attempts to acquire the lock with a timeout
func (il *IndexLock) TryLock(timeout time.Duration) error {
	// Try to acquire lock with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	locked, err := il.flock.TryLockContext(ctx, time.Millisecond*100)
	if err != nil {
		return fmt.Errorf("failed to acquire index lock: %w", err)
	}
	
	if !locked {
		return fmt.Errorf("index is locked by another process (timeout after %v)", timeout)
	}
	
	return nil
}

// Unlock releases the lock
func (il *IndexLock) Unlock() error {
	if il.flock.Locked() {
		return il.flock.Unlock()
	}
	return nil
}

// IsLocked returns whether the lock is currently held
func (il *IndexLock) IsLocked() bool {
	return il.flock.Locked()
}

// GetLockPath returns the path to the lock file
func (il *IndexLock) GetLockPath() string {
	return il.lockPath
}