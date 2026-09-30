package store

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
			_ = root.Remove(tmpPath)
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

	// Atomic rename over the target path.
	err = root.Rename(tmpPath, relPath)
	if err != nil && runtime.GOOS == "windows" {
		backoff := 50 * time.Millisecond
		for attempt := 0; attempt < 3; attempt++ {
			time.Sleep(backoff)
			backoff *= 2
			if err = root.Rename(tmpPath, relPath); err == nil {
				break
			}
		}
	}
	if err != nil {
		return err
	}

	success = true
	return nil
}
