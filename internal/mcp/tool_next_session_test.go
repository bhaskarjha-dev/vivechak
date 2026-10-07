package mcputil

import (
	"os"
	"path/filepath"
	"strings"
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

	// 5. Session with slugged filename and NO frontmatter (Fix HIGH-02)
	s4Content := `# Storage Layer Findings`
	if err := os.WriteFile(filepath.Join(sessionsDir, "T1-03-storage-layer.md"), []byte(s4Content), 0o644); err != nil {
		t.Fatalf("failed to write s4: %v", err)
	}

	// 6. Custom session with arbitrary DAG ID and slug
	s5Content := `# Custom Step Findings`
	if err := os.WriteFile(filepath.Join(sessionsDir, "CUSTOM-STEP-analysis.md"), []byte(s5Content), 0o644); err != nil {
		t.Fatalf("failed to write s5: %v", err)
	}

	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("open workspace failed: %v", err)
	}
	defer ws.Close()

	completed, err := scanCompletedSessions(ws, "CUSTOM-STEP")
	if err != nil {
		t.Fatalf("scanCompletedSessions failed: %v", err)
	}

	expectedIDs := []string{"T1-01", "T1-02", "T2-01", "FAD", "T1-03", "CUSTOM-STEP"}
	for _, id := range expectedIDs {
		if !completed[id] {
			t.Errorf("expected session %q to be marked completed, got %v", id, completed)
		}
	}

	// Verify T1-01 was NOT registered as "T1-01-database" because frontmatter was authoritative
	if completed["T1-01-database"] {
		t.Errorf("T1-01-database should not be registered when frontmatter specified T1-01")
	}
}

func TestNextSession_CalibrationBlockInjection(t *testing.T) {
	dag := &core.DAG{
		Sessions: []core.Session{
			{ID: "T1-01", Title: "DB Research"},
			{ID: "SYN-01", Title: "Synthesis"},
			{ID: "FAD", Title: "Founding Architecture Document"},
		},
	}

	// 1. Regular research session gets calibration rules injected
	_, envT1, err := buildSessionResponse("vivechak_next_session", dag.Sessions[0], "Base prompt for T1-01", map[string]bool{}, dag, false, nil)
	if err != nil {
		t.Fatalf("buildSessionResponse for T1-01 failed: %v", err)
	}
	dataT1, ok := envT1.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %T", envT1.Data)
	}
	promptT1, ok := dataT1["prompt"].(string)
	if !ok || !strings.Contains(promptT1, "### RESEARCH CALIBRATION (Active Rules)") {
		t.Errorf("expected calibration block in T1-01 prompt, got: %s", promptT1)
	}
	if !strings.Contains(promptT1, "Qualified Provenance") {
		t.Errorf("expected qualified provenance rule in T1-01 prompt")
	}

	// 2. Synthesis session SYN-01 is exempted
	_, envSyn, err := buildSessionResponse("vivechak_next_session", dag.Sessions[1], "Base synthesis prompt", map[string]bool{}, dag, false, nil)
	if err != nil {
		t.Fatalf("buildSessionResponse for SYN-01 failed: %v", err)
	}
	dataSyn := envSyn.Data.(map[string]any)
	promptSyn, _ := dataSyn["prompt"].(string)
	if strings.Contains(promptSyn, "### RESEARCH CALIBRATION") {
		t.Errorf("synthesis session SYN-01 should NOT have calibration block injected, got: %s", promptSyn)
	}

	// 3. Synthesis session FAD is exempted
	_, envFAD, err := buildSessionResponse("vivechak_next_session", dag.Sessions[2], "Base FAD prompt", map[string]bool{}, dag, false, nil)
	if err != nil {
		t.Fatalf("buildSessionResponse for FAD failed: %v", err)
	}
	dataFAD := envFAD.Data.(map[string]any)
	promptFAD, _ := dataFAD["prompt"].(string)
	if strings.Contains(promptFAD, "### RESEARCH CALIBRATION") {
		t.Errorf("synthesis session FAD should NOT have calibration block injected, got: %s", promptFAD)
	}

	// 4. Empty prompt does not inject calibration block
	_, envEmpty, err := buildSessionResponse("vivechak_next_session", dag.Sessions[0], "", map[string]bool{}, dag, false, nil)
	if err != nil {
		t.Fatalf("buildSessionResponse for empty prompt failed: %v", err)
	}
	dataEmpty := envEmpty.Data.(map[string]any)
	if promptEmpty, exists := dataEmpty["prompt"]; exists && promptEmpty != "" {
		t.Errorf("expected no prompt for empty input, got: %v", promptEmpty)
	}
}
