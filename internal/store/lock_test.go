package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestLockFile(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "mylock")

	unlock1, err := LockFile(lockPath, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// Second lock attempt should fail
	_, err = LockFile(lockPath, 50*time.Millisecond)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}

	// Unlock first lock
	err = unlock1()
	if err != nil {
		t.Fatal(err)
	}

	// Now second lock attempt should succeed
	unlock2, err := LockFile(lockPath, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock2()
}
