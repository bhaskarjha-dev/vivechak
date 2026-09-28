package mcputil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

func TestScanCompletedSessions(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, core.SessionsDir)
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sessions dir: %v", err)
	}

	// 1. Session with frontmatter session_id
	s1Content := `---
session_id: T1-01
title: DB Selection
---
# Content`
	if err := os.WriteFile(filepath.Join(sessionsDir, "T1-01-database.md"), []byte(s1Content), 0o644); err != nil {
		t.Fatalf("failed to write s1: %v", err)
	}

	// 2. Session with frontmatter id alias
	s2Content := `---
id: T1-02
title: Auth
---
# Content`
	if err := os.WriteFile(filepath.Join(sessionsDir, "auth.md"), []byte(s2Content), 0o644); err != nil {
		t.Fatalf("failed to write s2: %v", err)
	}

	// 3. Session without frontmatter -> filename stem
	s3Content := `# No frontmatter`
	if err := os.WriteFile(filepath.Join(sessionsDir, "T2-01.md"), []byte(s3Content), 0o644); err != nil {
		t.Fatalf("failed to write s3: %v", err)
	}

	// 4. FAD at research/FAD.md (H-04 fix)
	fadContent := `---
session_id: FAD
title: Founding Architecture Document
---
# Final Synthesis`
	if err := os.WriteFile(filepath.Join(tmpDir, core.FADFile), []byte(fadContent), 0o644); err != nil {
		t.Fatalf("failed to write FAD: %v", err)
	}

	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("open workspace failed: %v", err)
	}
	defer ws.Close()

	completed, err := scanCompletedSessions(ws)
	if err != nil {
		t.Fatalf("scanCompletedSessions failed: %v", err)
	}

	expectedIDs := []string{"T1-01", "T1-02", "T2-01", "FAD"}
	for _, id := range expectedIDs {
		if !completed[id] {
			t.Errorf("expected session %q to be marked completed, got %v", id, completed)
		}
	}

	// Verify T1-01 was NOT registered as "T1-01-database"
	if completed["T1-01-database"] {
		t.Errorf("T1-01-database should not be registered when frontmatter specified T1-01")
	}
}
