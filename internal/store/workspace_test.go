package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create some files
	err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	w, err := OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if w.AbsPath() == "" {
		t.Error("expected non-empty abs path")
	}

	// Test ReadFile
	data, err := w.ReadFile("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("expected hello, got %s", string(data))
	}

	// Test MkdirAll
	err = w.MkdirAll("sub/dir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Test Stat
	info, err := w.Stat("sub/dir")
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}

	// Test ListDir
	entries, err := w.ListDir(".")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range entries {
		if e.Name() == "test.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("test.txt not found in ListDir")
	}

	// Test Path Traversal Rejection
	// os.Root should reject paths starting with ../
	_, err = w.ReadFile("../test.txt")
	if err == nil {
		t.Error("expected error for path traversal, got nil")
	} else if !strings.Contains(err.Error(), "invalid path") && !os.IsNotExist(err) && !os.IsPermission(err) {
		// Just ensuring it failed.
	}
}
