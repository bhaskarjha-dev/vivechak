package store

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func WriteFileAtomic(root *os.Root, relPath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(relPath)
	if dir == "." {
		dir = ""
	}
	tmpName := fmt.Sprintf(".tmp_%d_%d", time.Now().UnixNano(), os.Getpid())
	tmpPath := filepath.Join(dir, tmpName)

	f, err := root.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	success := false
	defer func() {
		if !success {
			root.Remove(tmpPath)
		}
	}()

	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}

	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	// Rename over the target
	// Go 1.25+ os.Root doesn't necessarily have Rename in all early drafts, let's try calling it.
	// We will compile and fix if it fails.
	err = root.Rename(tmpPath, relPath)
	if err != nil {
		return err
	}

	success = true
	return nil
}
