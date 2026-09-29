package store

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Workspace struct {
	root    *os.Root
	absPath string
}

func OpenWorkspace(path string) (*Workspace, error) {
	abs, err := filepath.EvalSymlinks(path)
	if err != nil {
		// If path doesn't exist yet, EvalSymlinks might fail. Let's try Abs directly.
		abs = path
	}
	abs, err = filepath.Abs(abs)
	if err != nil {
		return nil, err
	}
	r, err := os.OpenRoot(abs)
	if err != nil {
		return nil, err
	}
	return &Workspace{
		root:    r,
		absPath: abs,
	}, nil
}

func (w *Workspace) AbsPath() string {
	return w.absPath
}

func (w *Workspace) ReadFile(relPath string) ([]byte, error) {
	f, err := w.root.Open(relPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (w *Workspace) Stat(relPath string) (os.FileInfo, error) {
	return w.root.Stat(relPath)
}

// FileExists reports whether relPath exists and can be stated within the workspace root.
func (w *Workspace) FileExists(relPath string) bool {
	_, err := w.root.Stat(relPath)
	return err == nil
}

// MkdirAll creates directories within the workspace, validating that the
// path remains local (no traversal). Falls back to os.MkdirAll since
// os.Root does not provide MkdirAll as of Go 1.25.
//
// SECURITY NOTE: This function validates paths via filepath.IsLocal but does NOT
// use os.Root kernel-level confinement for directory creation. os.Root confinement
// applies to file I/O operations (Open, OpenFile, Stat, Rename, Remove) only.
// Directory creation is guarded by manual validation + absolute path construction.
func (w *Workspace) MkdirAll(relPath string, perm os.FileMode) error {
	if !filepath.IsLocal(relPath) {
		return fmt.Errorf("path escapes workspace: %s", relPath)
	}
	return os.MkdirAll(filepath.Join(w.absPath, relPath), perm)
}

func (w *Workspace) ListDir(relPath string) ([]os.DirEntry, error) {
	f, err := w.root.Open(relPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.ReadDir(-1)
}

func (w *Workspace) Close() error {
	return w.root.Close()
}

func (w *Workspace) Root() *os.Root {
	return w.root
}
