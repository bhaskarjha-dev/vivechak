package store

import (
	"context"
	"time"

	"github.com/gofrs/flock"
)

// LockFile acquires an exclusive advisory lock on the given path.
// Returns an unlock function. The lock file is created at path + ".lock".
// Timeout is how long to wait for the lock before giving up.
func LockFile(ctx context.Context, path string, timeout time.Duration) (func() error, error) {
	lockPath := path + ".lock"
	f := flock.New(lockPath)

	lockCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Retry every 10ms
	locked, err := f.TryLockContext(lockCtx, 10*time.Millisecond)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, context.DeadlineExceeded
	}

	unlock := func() error {
		// Lock file is intentionally NOT removed on unlock.
		// Removing it creates a TOCTOU race: another process can acquire the
		// lock between Unlock() and Remove(), then we delete their lock file.
		return f.Unlock()
	}

	return unlock, nil
}
