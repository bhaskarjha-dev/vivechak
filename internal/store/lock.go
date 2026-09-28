package store

import (
	"context"
	"time"

	"github.com/gofrs/flock"
)

// LockFile acquires an exclusive advisory lock on the given path.
// Returns an unlock function. The lock file is created at path + ".lock".
// Timeout is how long to wait for the lock before giving up.
func LockFile(path string, timeout time.Duration) (func() error, error) {
	lockPath := path + ".lock"
	f := flock.New(lockPath)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Retry every 10ms
	locked, err := f.TryLockContext(ctx, 10*time.Millisecond)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, context.DeadlineExceeded
	}

	unlock := func() error {
		return f.Unlock()
	}

	return unlock, nil
}
