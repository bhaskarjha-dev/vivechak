package store

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNormalizeLineEndings(t *testing.T) {
	crlf := []byte("line1\r\nline2\r\nline3\r\n")
	lf := NormalizeLineEndings(crlf)
	if bytes.Contains(lf, []byte("\r")) {
		t.Errorf("expected no CR in normalized output: %q", lf)
	}

	bomCRLF := append([]byte("\xef\xbb\xbf"), crlf...)
	cleaned := NormalizeLineEndings(bomCRLF)
	if bytes.HasPrefix(cleaned, []byte("\xef\xbb\xbf")) {
		t.Error("expected UTF-8 BOM to be stripped")
	}
	if string(cleaned) != "line1\nline2\nline3\n" {
		t.Errorf("unexpected normalized output: %q", string(cleaned))
	}
}

func TestExtractFrontmatterID(t *testing.T) {
	doc1 := []byte("---\nid: D-001\ntitle: Test\n---\nBody")
	if id := ExtractFrontmatterID(doc1); id != "D-001" {
		t.Errorf("got %q, want D-001", id)
	}

	doc2 := []byte("---\nsession_id: \"T1-01\"\n---\nBody")
	if id := ExtractFrontmatterID(doc2); id != "T1-01" {
		t.Errorf("got %q, want T1-01", id)
	}

	doc3 := []byte("---\ndecision_id: 'C-002'\n---\nBody")
	if id := ExtractFrontmatterID(doc3); id != "C-002" {
		t.Errorf("got %q, want C-002", id)
	}

	doc4 := []byte("# Just markdown\nNo frontmatter")
	if id := ExtractFrontmatterID(doc4); id != "" {
		t.Errorf("expected empty id, got %q", id)
	}
}

func TestWriteArtifact_ConfinementAndConflict(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	// 1. Path confinement check
	err = ws.WriteArtifact("../escaped.md", []byte("data"), 0o644, false)
	if err == nil || !strings.Contains(err.Error(), "E-CONFINEMENT") {
		t.Errorf("expected E-CONFINEMENT error on escaping path, got: %v", err)
	}

	err = ws.WriteArtifact("outside/file.md", []byte("data"), 0o644, false)
	if err == nil || !strings.Contains(err.Error(), "E-CONFINEMENT") {
		t.Errorf("expected E-CONFINEMENT error on non-research path, got: %v", err)
	}

	// 2. Initial write under research/
	target := "research/decisions/D-001.md"
	content1 := []byte("---\nid: D-001\ntitle: Initial Decision\n---\n# Decision 1")
	if err := ws.WriteArtifact(target, content1, 0o644, false); err != nil {
		t.Fatalf("initial write failed: %v", err)
	}

	// 3. Conflict detection on differing IDs
	contentDiffering := []byte("---\nid: D-002\ntitle: Different Decision\n---\n# Decision 2")
	err = ws.WriteArtifact(target, contentDiffering, 0o644, false)
	if err == nil || !strings.Contains(err.Error(), "E-ID-CONFLICT") {
		t.Errorf("expected E-ID-CONFLICT error, got: %v", err)
	}

	// 4. Revision archiving on revise=true
	contentRevised := []byte("---\nid: D-001\ntitle: Revised Decision\n---\n# Decision 1 Revised")
	if err := ws.WriteArtifact(target, contentRevised, 0o644, true); err != nil {
		t.Fatalf("revised write failed: %v", err)
	}

	// Verify current content
	curData, err := ws.ReadFile(target)
	if err != nil || !strings.Contains(string(curData), "Revised Decision") {
		t.Errorf("expected target to have revised content, got %s, err: %v", string(curData), err)
	}

	// Verify archive exists in research/.history/
	historyDir := "research/.history/decisions"
	entries, err := ws.ListDir(historyDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("expected archived revision in %s, got err: %v, entries: %d", historyDir, err, len(entries))
	}
	if !strings.HasPrefix(entries[0].Name(), "D-001.md@") {
		t.Errorf("expected archive file starting with D-001.md@, got: %s", entries[0].Name())
	}
}

func TestArchiveRevision_Cap10(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	target := "research/decisions/D-010.md"
	content := []byte("---\nid: D-010\n---\nBase")
	_ = ws.WriteArtifact(target, content, 0o644, false)

	// Create 15 revisions with small time steps
	for i := 1; i <= 15; i++ {
		revised := []byte(fmt.Sprintf("---\nid: D-010\n---\nRevision %d", i))
		_ = ws.WriteArtifact(target, revised, 0o644, true)
		time.Sleep(10 * time.Millisecond)
	}

	entries, err := ws.ListDir("research/.history/decisions")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) > 10 {
		t.Errorf("expected maximum 10 archived revisions, got %d", len(entries))
	}
}

func TestAllocateID(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	// 1. Allocate initial IDs
	id1, err := AllocateID(ws, "D")
	if err != nil || id1 != "D-001" {
		t.Errorf("expected D-001, got %q, err: %v", id1, err)
	}

	id2, err := AllocateID(ws, "D")
	if err != nil || id2 != "D-002" {
		t.Errorf("expected D-002, got %q, err: %v", id2, err)
	}

	idC1, err := AllocateID(ws, "C")
	if err != nil || idC1 != "C-001" {
		t.Errorf("expected C-001, got %q, err: %v", idC1, err)
	}

	// 2. Pre-existing files without manifest counter
	tmpDir2 := t.TempDir()
	ws2, err := OpenWorkspace(tmpDir2)
	if err != nil {
		t.Fatal(err)
	}
	defer ws2.Close()

	_ = ws2.MkdirAll("research/decisions", 0o755)
	_ = os.WriteFile(filepath.Join(tmpDir2, "research", "decisions", "D-004-existing.md"), []byte("data"), 0o644)

	idNext, err := AllocateID(ws2, "D")
	if err != nil || idNext != "D-005" {
		t.Errorf("expected D-005 scanning pre-existing files, got %q, err: %v", idNext, err)
	}
}

func TestProperty_1000RandomizedWrites_ZeroDataLoss(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	r := rand.New(rand.NewSource(42))
	docCount := 20
	latestContent := make(map[string]string)

	for seq := 0; seq < 1000; seq++ {
		docIndex := r.Intn(docCount)
		id := fmt.Sprintf("D-%03d", docIndex+1)
		relPath := fmt.Sprintf("research/decisions/%s.md", id)

		payload := fmt.Sprintf("---\nid: %s\niteration: %d\n---\n# Data for iteration %d\nRandom nonce: %d",
			id, seq, seq, r.Int63())

		err := ws.WriteArtifact(relPath, []byte(payload), 0o644, true)
		if err != nil {
			t.Fatalf("sequence %d write failed: %v", seq, err)
		}
		latestContent[relPath] = payload
	}

	// Verify all written artifacts match their latest iteration exactly
	for relPath, expected := range latestContent {
		data, err := ws.ReadFile(relPath)
		if err != nil {
			t.Errorf("reading back %s failed: %v", relPath, err)
			continue
		}
		if string(data) != expected {
			t.Errorf("data mismatch in %s: expected %s, got %s", relPath, expected, string(data))
		}
	}
}
