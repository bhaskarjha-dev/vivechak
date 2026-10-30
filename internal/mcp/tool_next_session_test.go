package mcputil

import (
	"context"
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
	_, envT1, err := buildSessionResponse("vivechak_next", dag.Sessions[0], "Base prompt for T1-01", map[string]bool{}, dag, false, nil, nil)
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
	if !strings.Contains(promptT1, "Upstream") {
		t.Errorf("expected upstream != exhaustive rule in T1-01 prompt")
	}

	// 2. Synthesis session SYN-01 is exempted
	_, envSyn, err := buildSessionResponse("vivechak_next", dag.Sessions[1], "Base synthesis prompt", map[string]bool{}, dag, false, nil, nil)
	if err != nil {
		t.Fatalf("buildSessionResponse for SYN-01 failed: %v", err)
	}
	dataSyn := envSyn.Data.(map[string]any)
	promptSyn, _ := dataSyn["prompt"].(string)
	if strings.Contains(promptSyn, "### RESEARCH CALIBRATION") {
		t.Errorf("synthesis session SYN-01 should NOT have calibration block injected, got: %s", promptSyn)
	}

	// 3. Synthesis session FAD is exempted
	_, envFAD, err := buildSessionResponse("vivechak_next", dag.Sessions[2], "Base FAD prompt", map[string]bool{}, dag, false, nil, nil)
	if err != nil {
		t.Fatalf("buildSessionResponse for FAD failed: %v", err)
	}
	dataFAD := envFAD.Data.(map[string]any)
	promptFAD, _ := dataFAD["prompt"].(string)
	if strings.Contains(promptFAD, "### RESEARCH CALIBRATION") {
		t.Errorf("synthesis session FAD should NOT have calibration block injected, got: %s", promptFAD)
	}

	// 4. Empty prompt does not inject calibration block
	_, envEmpty, err := buildSessionResponse("vivechak_next", dag.Sessions[0], "", map[string]bool{}, dag, false, nil, nil)
	if err != nil {
		t.Fatalf("buildSessionResponse for empty prompt failed: %v", err)
	}
	dataEmpty := envEmpty.Data.(map[string]any)
	if promptEmpty, exists := dataEmpty["prompt"]; exists && promptEmpty != "" {
		t.Errorf("expected no prompt for empty input, got: %v", promptEmpty)
	}
}

func TestHandleNextSession_DraftBlocksDownstream(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	_, envInit, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir})
	if err != nil || !envInit.Success {
		t.Fatalf("init failed: %v", err)
	}

	pipelineContent := `# Research Pipeline

### T0-01: Baseline
| **ID** | T0-01 |
| **Door Type** | Two-Way |
| **Dependencies** | None |
` + "```prompt" + `
baseline prompt
` + "```" + `

### T1-01: Downstream
| **ID** | T1-01 |
| **Door Type** | One-Way |
| **Dependencies** | T0-01 |
` + "```prompt" + `
downstream prompt
` + "```" + `

### SYN-01: Synthesis
| **ID** | SYN-01 |
| **Door Type** | Synthesis |
| **Dependencies** | T1-01 |
` + "```prompt" + `
synthesis prompt
` + "```" + `
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(pipelineContent), 0o644)

	sessionsDir := filepath.Join(tmpDir, core.SessionsDir)

	// Save T0-01 with status: draft
	t0Draft := `---
session_id: T0-01
status: draft
---
# Draft Findings`
	_ = os.WriteFile(filepath.Join(sessionsDir, "T0-01.md"), []byte(t0Draft), 0o644)

	_, env, err := handleNextSession(ctx, nil, NextSessionInput{ProjectRoot: tmpDir})
	if err != nil {
		t.Fatalf("handleNextSession failed: %v", err)
	}

	// T0-01 should be the ready session (since its draft status means it is incomplete)
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %T", env.Data)
	}
	if data["session_id"] != "T0-01" {
		t.Errorf("expected ready session to be T0-01, got %v", data["session_id"])
	}

	// T1-01 should be reported in blocked_sessions with draft explanation
	blocked, ok := data["blocked_sessions"].([]map[string]any)
	if !ok || len(blocked) == 0 {
		t.Fatalf("expected blocked_sessions in data, got %v", data["blocked_sessions"])
	}
	foundT1 := false
	for _, b := range blocked {
		if b["session_id"] == "T1-01" {
			foundT1 = true
			blockedBy, _ := b["blocked_by"].([]string)
			hasDraftNotice := false
			for _, dep := range blockedBy {
				if strings.Contains(dep, "T0-01 (saved as draft — resolve blocking issues and re-save)") {
					hasDraftNotice = true
				}
			}
			if !hasDraftNotice {
				t.Errorf("expected blocked_by to mention draft notice, got: %v", blockedBy)
			}
		}
	}
	if !foundT1 {
		t.Errorf("expected T1-01 in blocked_sessions")
	}
}

func TestHandleNextSession_PreSynthesisAdvisory(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	_, envInit, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir})
	if err != nil || !envInit.Success {
		t.Fatalf("init failed: %v", err)
	}

	pipelineContent := `# Research Pipeline

### T1-01: Architecture Selection
| **ID** | T1-01 |
| **Door Type** | One-Way |
| **Dependencies** | None |
` + "```prompt" + `
architecture prompt
` + "```" + `

### SYN-01: Synthesis
| **ID** | SYN-01 |
| **Door Type** | Synthesis |
| **Dependencies** | T1-01 |
` + "```prompt" + `
synthesis prompt
` + "```" + `
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(pipelineContent), 0o644)

	sessionsDir := filepath.Join(tmpDir, core.SessionsDir)

	// Save T1-01 as completed, but with Grade D recalled claims in recommendation and lacking adversarial challenge
	t1Content := `---
session_id: T1-01
status: completed
---
# T1-01 Findings

## Recommendation
We recommend Option A based on recalled knowledge [Grade D | recalled].

## Key Findings
- Some findings
`
	_ = os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(t1Content), 0o644)

	_, env, err := handleNextSession(ctx, nil, NextSessionInput{ProjectRoot: tmpDir})
	if err != nil {
		t.Fatalf("handleNextSession failed: %v", err)
	}

	data, _ := env.Data.(map[string]any)
	if data["session_id"] != "SYN-01" {
		t.Errorf("expected next session to be SYN-01, got %v", data["session_id"])
	}

	// Verify advisory warning is present without blocking SYN-01
	hasAdvisory := false
	for _, w := range env.Warnings {
		if strings.Contains(w, "PRE-SYNTHESIS ADVISORY") && strings.Contains(w, "Grade D recalled") {
			hasAdvisory = true
		}
	}
	if !hasAdvisory {
		t.Errorf("expected PRE-SYNTHESIS ADVISORY warning in env.Warnings, got: %v", env.Warnings)
	}

	if !strings.Contains(env.NextStep, "PRE-SYNTHESIS ADVISORY") {
		t.Errorf("expected PRE-SYNTHESIS ADVISORY in nextStep, got: %s", env.NextStep)
	}
	if env.Next != nil && env.Next.Args != nil {
		if env.Next.Args["kind"] != "fad" {
			t.Errorf("expected Next.Args['kind'] to be 'fad', got %v", env.Next.Args["kind"])
		}
	}
	if !strings.Contains(env.NextStep, "kind='fad'") {
		t.Errorf("expected nextStep to suggest kind='fad', got: %s", env.NextStep)
	}
}
