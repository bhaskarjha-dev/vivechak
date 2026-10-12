package store

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
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

func TestWriteFileAtomic_Concurrent(t *testing.T) {
	tmpDir := t.TempDir()
	w, err := OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	root := w.Root()
	target := "concurrent.txt"
	const numWriters = 20
	var wg sync.WaitGroup

	errCh := make(chan error, numWriters)

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			payload := []byte(fmt.Sprintf("version_%03d_data_payload_content", id))
			if err := WriteFileAtomic(root, target, payload, 0644); err != nil {
				errCh <- fmt.Errorf("writer %d: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent write error: %v", err)
	}

	// Verify target file exists and content is a valid complete version
	data, err := w.ReadFile(target)
	if err != nil {
		t.Fatalf("reading target file: %v", err)
	}
	dataStr := string(data)
	if !strings.HasPrefix(dataStr, "version_") || !strings.HasSuffix(dataStr, "_data_payload_content") {
		t.Errorf("corrupted content after concurrent writes: %q", dataStr)
	}

	// Verify no leftover temp files
	entries, err := w.ListDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp_") {
			t.Errorf("temp file leaked after concurrent writes: %s", e.Name())
		}
	}
}

func TestWriteFileAtomic_CleanupOnFailure(t *testing.T) {
	tmpDir := t.TempDir()
	w, err := OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	root := w.Root()
	// Target path in a non-existent subdirectory will fail at rename or open
	err = WriteFileAtomic(root, "nonexistent_dir/target.txt", []byte("data"), 0644)
	if err == nil {
		t.Error("expected error for nonexistent directory target, got nil")
	}

	// Verify no temp file left in root directory
	entries, err := w.ListDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp_") {
			t.Errorf("temp file leaked after failed write: %s", e.Name())
		}
	}
}

func TestWriteFileAtomic_ContentIntegrity(t *testing.T) {
	tmpDir := t.TempDir()
	w, err := OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	root := w.Root()
	target := "large.txt"
	largePayload := bytes.Repeat([]byte("Vivechak evidentiary grounding and architectural decision testing.\n"), 1000)

	if err := WriteFileAtomic(root, target, largePayload, 0644); err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}

	readBack, err := w.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if !bytes.Equal(readBack, largePayload) {
		t.Errorf("payload mismatch: expected %d bytes, got %d bytes", len(largePayload), len(readBack))
	}
}
