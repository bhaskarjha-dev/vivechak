package store

import (
	"testing"
)

func TestWriteFileAtomic(t *testing.T) {
	tmpDir := t.TempDir()

	w, err := OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	root := w.Root()

	err = WriteFileAtomic(root, "atomic.txt", []byte("atomic data"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	data, err := w.ReadFile("atomic.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "atomic data" {
		t.Errorf("expected atomic data, got %s", string(data))
	}

	// Verify no temp files left behind
	entries, err := w.ListDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "atomic.txt" {
			t.Errorf("unexpected file found: %s", e.Name())
		}
	}
}
