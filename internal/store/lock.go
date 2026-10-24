package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

// resolveLockPath determines the lock file path in a dedicated .locks directory.
// For paths within a research directory, it isolates locks under research/.locks/.
// For other paths, it isolates locks under <dir>/.locks/.
func resolveLockPath(targetPath string) (string, error) {
	abs, err := filepath.Abs(targetPath)
	if err != nil {
		abs = targetPath
	}
	clean := filepath.Clean(abs)

	// Check if path contains "research" directory
	var locksDir string
	parts := strings.Split(clean, string(filepath.Separator))
	researchIdx := -1
	for i, part := range parts {
		if strings.EqualFold(part, "research") {
			researchIdx = i
			break
		}
	}

	if researchIdx >= 0 {
		researchRoot := strings.Join(parts[:researchIdx+1], string(filepath.Separator))
		locksDir = filepath.Join(researchRoot, ".locks")
	} else {
		locksDir = filepath.Join(filepath.Dir(clean), ".locks")
	}

	if err := os.MkdirAll(locksDir, 0o755); err != nil {
		return "", fmt.Errorf("creating locks directory %q: %w", locksDir, err)
	}

	base := filepath.Base(clean)
	h := sha256.Sum256([]byte(clean))
	hashSuffix := hex.EncodeToString(h[:4])
	sanitized := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '_'
	}, base)

	lockFileName := fmt.Sprintf("%s_%s.lock", sanitized, hashSuffix)
	return filepath.Join(locksDir, lockFileName), nil
}

// LockFile acquires an exclusive advisory lock on the given path.
// Returns an unlock function. The lock file is placed in a .locks directory.
// Timeout is how long to wait for the lock before giving up.
func LockFile(ctx context.Context, path string, timeout time.Duration) (func() error, error) {
	lockPath, err := resolveLockPath(path)
	if err != nil {
		return nil, err
	}
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
