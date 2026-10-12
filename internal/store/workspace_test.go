package store

import (
	"os"
	"path/filepath"
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
	}
}

func TestWorkspace_MkdirAllHardening(t *testing.T) {
	tmpDir := t.TempDir()
	w, err := OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Deeply nested
	if err := w.MkdirAll("level1/level2/level3", 0755); err != nil {
		t.Fatalf("MkdirAll failed for nested path: %v", err)
	}

	// Idempotent call
	if err := w.MkdirAll("level1/level2/level3", 0755); err != nil {
		t.Fatalf("MkdirAll failed on second call: %v", err)
	}

	// Dot and empty paths
	if err := w.MkdirAll(".", 0755); err != nil {
		t.Fatalf("MkdirAll failed on dot path: %v", err)
	}

	// Traversal escapes rejected
	if err := w.MkdirAll("../escape", 0755); err == nil {
		t.Error("expected error for escaping path, got nil")
	}

	// File collision: create a file, then try to MkdirAll through it
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := w.MkdirAll("file.txt/child", 0755); err == nil {
		t.Error("expected error when path component is a file, got nil")
	}
}
