package mcputil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

func TestBuildNextStep_Synthesis_ExitGate(t *testing.T) {
	nextStep := buildNextStep("SYN-01", 0, nil, nil)
	if !strings.Contains(nextStep, "vivechak_run_gate") {
		t.Errorf("expected synthesis nextStep to mandate vivechak_run_gate, got: %s", nextStep)
	}

	nextStepFAD := buildNextStep("FAD", 0, nil, nil)
	if !strings.Contains(nextStepFAD, "vivechak_run_gate") {
		t.Errorf("expected FAD nextStep to mandate vivechak_run_gate, got: %s", nextStepFAD)
	}
}

func TestBuildNextStep_ADRAutoDraftGuidance(t *testing.T) {
	dag := &core.DAG{
		Sessions: []core.Session{
			{ID: "T1-01", DoorType: "one-way"},
			{ID: "T1-02", DoorType: "two-way"},
		},
	}
	owdStep := buildNextStep("T1-01", 0, nil, dag)
	if !strings.Contains(owdStep, `auto_draft_from="T1-01"`) {
		t.Errorf("expected OWD nextStep to include auto_draft_from guidance, got: %s", owdStep)
	}

	twdStep := buildNextStep("T1-02", 0, nil, dag)
	if !strings.Contains(twdStep, `auto_draft_from="T1-02"`) {
		t.Errorf("expected TWD nextStep to include auto_draft_from guidance, got: %s", twdStep)
	}
}

func TestHandleSaveSession_FrontmatterStatusSync(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// Initialize workspace
	initRes, initEnv, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "project"})
	if err != nil || initRes.IsError {
		t.Fatalf("handleInit failed: %v, env: %+v", err, initEnv)
	}

	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("OpenWorkspace failed: %v", err)
	}
	defer ws.Close()

	// Save pipeline with an irreversible One-Way Door session
	pipelineMD := `# Pipeline

### Session T1-01: Primary Datastore

| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Layer** | 1 |
| **Door Type** | One-Way |
| **Decision** | D-001 |
| **Dependencies** | none |
| **Output File** | research/sessions/T1-01.md |

` + "```prompt" + `
Prompt text
` + "```" + `

### Session T1-02: Cache Layer

| Field | Value |
|---|---|
| **ID** | T1-02 |
| **Layer** | 1 |
| **Door Type** | Two-Way |
| **Dependencies** | T1-01 |
| **Output File** | research/sessions/T1-02.md |

` + "```prompt" + `
Prompt text
` + "```\n"
	if err := os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(pipelineMD), 0o644); err != nil {
		t.Fatalf("writing pipeline failed: %v", err)
	}

	// 1. Session with blocking L2 error (One-Way Door missing Discovered Concerns)
	// Even if user provides status: complete, save_session MUST save it as status: draft on disk
	invalidOWDContent := `---
session_id: T1-01
title: Primary Datastore
status: complete
date: 2026-10-14
---

# T1-01: Primary Datastore

## Recommendation
Postgres is selected (Grade A · canonical | fetched https://www.postgresql.org).
`
	res1, env1, err := handleSaveSession(ctx, nil, SaveSessionInput{
		ProjectRoot: tmpDir,
		SessionID:   "T1-01",
		Content:     invalidOWDContent,
	})
	if err != nil || res1.IsError {
		t.Fatalf("save invalid session failed: %v, env: %+v", err, env1)
	}

	// Check file on disk has status: draft
	diskData1, err := ws.ReadFile("research/sessions/T1-01.md")
	if err != nil {
		t.Fatalf("reading saved session failed: %v", err)
	}
	fm1, _, err := core.ParseFrontmatter(diskData1)
	if err != nil || fm1 == nil {
		t.Fatalf("parsing saved frontmatter failed: %v", err)
	}
	if fm1.GetString("status") != "draft" {
		t.Errorf("expected status on disk to be 'draft' due to blocking L2 error, got: %q", fm1.GetString("status"))
	}

	// Verify scanCompletedSessions does NOT mark T1-01 as completed
	completed1, err := scanCompletedSessions(ws, "T1-01", "T1-02")
	if err != nil {
		t.Fatalf("scanCompletedSessions failed: %v", err)
	}
	if completed1["T1-01"] {
		t.Errorf("session with blocking errors must not be marked completed")
	}

	// 2. Fix the session with substantive Discovered Concerns (0 blocking errors)
	// Input starts with status: draft (like SESSION.template.md)
	validOWDContent := `---
session_id: T1-01
title: Primary Datastore
status: draft
date: 2026-10-14
---

# T1-01: Primary Datastore

## Recommendation
Postgres is selected (Grade A · canonical | fetched https://www.postgresql.org).

## Discovered Concerns
Found major connection limits and serialization anomalies under high concurrent connection spikes requiring pgbouncer.
`
	res2, env2, err := handleSaveSession(ctx, nil, SaveSessionInput{
		ProjectRoot: tmpDir,
		SessionID:   "T1-01",
		Content:     validOWDContent,
	})
	if err != nil || res2.IsError {
		t.Fatalf("save valid session failed: %v, env: %+v", err, env2)
	}

	// Check file on disk was promoted to status: complete
	diskData2, err := ws.ReadFile("research/sessions/T1-01.md")
	if err != nil {
		t.Fatalf("reading saved session failed: %v", err)
	}
	fm2, _, err := core.ParseFrontmatter(diskData2)
	if err != nil || fm2 == nil {
		t.Fatalf("parsing saved frontmatter failed: %v", err)
	}
	if fm2.GetString("status") != "complete" {
		t.Errorf("expected status on disk to be promoted to 'complete', got: %q", fm2.GetString("status"))
	}

	// Verify scanCompletedSessions now marks T1-01 as completed
	completed2, err := scanCompletedSessions(ws, "T1-01", "T1-02")
	if err != nil {
		t.Fatalf("scanCompletedSessions failed: %v", err)
	}
	if !completed2["T1-01"] {
		t.Errorf("valid session must be marked completed")
	}
}

func TestHandleSaveSession_SynthesisLockTimeoutWarning(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// Initialize workspace
	initRes, initEnv, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "project"})
	if err != nil || initRes.IsError {
		t.Fatalf("handleInit failed: %v, env: %+v", err, initEnv)
	}

	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("OpenWorkspace failed: %v", err)
	}
	defer ws.Close()

	// Hold an exclusive lock on FOUNDING-ARCHITECTURE.md.lock
	rootFADPath := filepath.Join(tmpDir, "FOUNDING-ARCHITECTURE.md")
	unlockRoot, err := store.LockFile(ctx, rootFADPath, 1*time.Second)
	if err != nil {
		t.Fatalf("failed to hold test lock on root FAD: %v", err)
	}
	defer func() { _ = unlockRoot() }()

	fadContent := `---
session_id: SYN-01
title: Founding Architecture Document
status: complete
date: 2026-10-14
---
# Founding Architecture Document
Content here.
`
	// Attempt to save synthesis session SYN-01
	res, env, err := handleSaveSession(ctx, nil, SaveSessionInput{
		ProjectRoot: tmpDir,
		SessionID:   "SYN-01",
		Content:     fadContent,
	})
	if err != nil || res.IsError {
		t.Fatalf("handleSaveSession failed: %v, env: %+v", err, env)
	}

	hasLockWarning := false
	for _, w := range env.Warnings {
		if strings.Contains(w, "W-LOCK-TIMEOUT") {
			hasLockWarning = true
			break
		}
	}
	if !hasLockWarning {
		t.Errorf("expected W-LOCK-TIMEOUT warning when root FAD lock is held, got warnings: %v", env.Warnings)
	}
}

