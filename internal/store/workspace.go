package store

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	if relPath == "" {
		relPath = "."
	}
	return w.root.Stat(relPath)
}

// FileExists reports whether relPath exists and can be stated within the workspace root.
func (w *Workspace) FileExists(relPath string) bool {
	_, err := w.root.Stat(relPath)
	return err == nil
}

// MkdirAll creates directories within the workspace using component-wise
// os.Root confinement. Each directory component is created via w.root.Mkdir
// to ensure strict containment without TOCTOU symlink escape risks.
func (w *Workspace) MkdirAll(relPath string, perm os.FileMode) error {
	if !filepath.IsLocal(relPath) {
		return fmt.Errorf("path escapes workspace: %s", relPath)
	}

	clean := filepath.Clean(relPath)
	if clean == "." || clean == "" {
		return nil
	}

	parts := strings.Split(filepath.ToSlash(clean), "/")
	var current string
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if current == "" {
			current = part
		} else {
			current = filepath.Join(current, part)
		}

		err := w.root.Mkdir(current, perm)
		if err != nil {
			if os.IsExist(err) {
				fi, statErr := w.root.Stat(current)
				if statErr != nil {
					return statErr
				}
				if !fi.IsDir() {
					return fmt.Errorf("path component %s exists and is not a directory", current)
				}
				continue
			}
			return err
		}
	}
	return nil
}

func (w *Workspace) ListDir(relPath string) ([]os.DirEntry, error) {
	if relPath == "" {
		relPath = "."
	}
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
