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

func TestBuildNextStep_Synthesis_ExitGate(t *testing.T) {
	// The synthesis session is the plan's kind: synthesis session (or a kind="fad" save), whatever its id.
	nextStep := buildNextStep("SYN-01", true, 0, nil, nil)
	if !strings.Contains(nextStep, "vivechak_gate") {
		t.Errorf("expected synthesis nextStep to mandate vivechak_gate, got: %s", nextStep)
	}

	nextStepFAD := buildNextStep("FINAL-01", true, 0, nil, nil)
	if !strings.Contains(nextStepFAD, "vivechak_gate") {
		t.Errorf("expected FAD nextStep to mandate vivechak_gate, got: %s", nextStepFAD)
	}

	if ordinary := buildNextStep("SYN-01", false, 0, nil, nil); strings.Contains(ordinary, "vivechak_gate") {
		t.Errorf("a session named SYN-01 that is not the synthesis session must not route to the gate, got: %s", ordinary)
	}
}

func TestBuildNextStep_ADRAutoDraftGuidance(t *testing.T) {
	dag := &core.DAG{
		Sessions: []core.Session{
			{ID: "T1-01", DoorType: "one-way"},
			{ID: "T1-02", DoorType: "two-way"},
		},
	}
	owdStep := buildNextStep("T1-01", false, 0, nil, dag)
	if !strings.Contains(owdStep, `auto_draft_from="T1-01"`) {
		t.Errorf("expected OWD nextStep to include auto_draft_from guidance, got: %s", owdStep)
	}

	twdStep := buildNextStep("T1-02", false, 0, nil, dag)
	if !strings.Contains(twdStep, `auto_draft_from="T1-02"`) {
		t.Errorf("expected TWD nextStep to include auto_draft_from guidance, got: %s", twdStep)
	}
}

func TestHandleSaveSession_FrontmatterStatusSync(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// Initialize workspace
	initRes, initEnv, err := handleStart(ctx, nil, StartInput{Question: "test", ProjectRoot: tmpDir, Scope: "project"})
	if err != nil || initRes.IsError {
		t.Fatalf("handleStart failed: %v, env: %+v", err, initEnv)
	}

	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("OpenWorkspace failed: %v", err)
	}
	defer ws.Close()

	// Save pipeline with an irreversible One-Way Door session
	pipelineMD := testPlan("Pipeline", "PROJECT", "", []string{
		"{id: T1-01, title: Primary Datastore, layer: 1, door: one-way, informs: [D-001], depends: []}",
		"{id: T1-02, title: Cache Layer, layer: 1, door: two-way, depends: [T1-01]}",
		"{id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}",
	}, testSection("T1-01", "Primary Datastore", "Prompt text"), testSection("T1-02", "Cache Layer", "Prompt text"),
		testSection("SYN-01", "Synthesis", "Synthesize."))
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
	res1, env1, err := handleSave(ctx, nil, SaveInput{
		Kind:        "session",
		ProjectRoot: tmpDir,
		ID:          "T1-01",
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
	res2, env2, err := handleSave(ctx, nil, SaveInput{
		Kind:        "session",
		ProjectRoot: tmpDir,
		ID:          "T1-01",
		Content:     validOWDContent,
		Revise:      true,
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

func TestHandleSaveSession_NoRootFADCopy(t *testing.T) {
	ctx := t.Context()
	tmpDir := t.TempDir()

	// Initialize workspace
	initRes, initEnv, err := handleStart(ctx, nil, StartInput{Question: "test", ProjectRoot: tmpDir, Scope: "project"})
	if err != nil || initRes.IsError {
		t.Fatalf("handleStart failed: %v, env: %+v", err, initEnv)
	}

	fadContent := `---
session_id: SYN-01
title: Founding Architecture Document
status: complete
date: 2026-10-14
---
# Founding Architecture Document
Content here with Grade A (canonical docs: https://example.com) architecture.
`
	// Save the synthesis (kind="fad"; with no plan, an id alone never makes a session the synthesis session)
	res, env, err := handleSave(ctx, nil, SaveInput{
		ProjectRoot: tmpDir,
		Kind:        "fad",
		ID:          "SYN-01",
		Content:     fadContent,
	})
	if err != nil || res.IsError {
		t.Fatalf("handleSave failed: %v, env: %+v", err, env)
	}

	// 1. research/FAD.md must exist
	fadPath := filepath.Join(tmpDir, "research", "FAD.md")
	if _, err := os.Stat(fadPath); err != nil {
		t.Fatalf("expected research/FAD.md to exist: %v", err)
	}

	// 2. FOUNDING-ARCHITECTURE.md in project root must NOT exist
	rootFAD := filepath.Join(tmpDir, "FOUNDING-ARCHITECTURE.md")
	if _, err := os.Stat(rootFAD); !os.IsNotExist(err) {
		t.Errorf("expected FOUNDING-ARCHITECTURE.md in root to NOT exist, but stat returned: %v", err)
	}
}

func TestHandleAmendSession_NoRootFADCopy(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	// Initialize workspace
	_, _, err := handleStart(ctx, nil, StartInput{Question: "test", ProjectRoot: tmpDir, Scope: "project"})
	if err != nil {
		t.Fatalf("handleStart failed: %v", err)
	}

	fadContent := `---
session_id: SYN-01
title: Founding Architecture Document
status: complete
date: 2026-10-14
---
# Founding Architecture Document
Base architecture content with Grade A (canonical docs: https://example.com).
`
	// Save initial SYN-01 as the synthesis (kind="fad"; with no plan, an id alone never makes it the synthesis)
	res, env, err := handleSave(ctx, nil, SaveInput{
		ProjectRoot: tmpDir,
		Kind:        "fad",
		ID:          "SYN-01",
		Content:     fadContent,
	})
	if err != nil || res.IsError {
		t.Fatalf("handleSave failed: %v, env: %+v", err, env)
	}

	// Amend synthesis session SYN-01
	resAmend, envAmend, err := handleAmend(ctx, nil, AmendInput{
		ProjectRoot: tmpDir,
		ID:          "SYN-01",
		Note:        "### Post-Hoc Synthesis Note\nValidated edge caching layer.\n",
	})
	if err != nil || resAmend.IsError {
		t.Fatalf("handleAmend failed: %v, env: %+v", err, envAmend)
	}

	// 1. research/FAD.md must contain amendment
	fadBytes, err := os.ReadFile(filepath.Join(tmpDir, "research", "FAD.md"))
	if err != nil {
		t.Fatalf("reading research/FAD.md: %v", err)
	}
	if !strings.Contains(string(fadBytes), "Validated edge caching layer.") {
		t.Errorf("research/FAD.md does not contain amendment")
	}

	// 2. Root FOUNDING-ARCHITECTURE.md must NOT exist
	rootFAD := filepath.Join(tmpDir, "FOUNDING-ARCHITECTURE.md")
	if _, err := os.Stat(rootFAD); !os.IsNotExist(err) {
		t.Errorf("expected FOUNDING-ARCHITECTURE.md in root to NOT exist, but stat returned: %v", err)
	}
}
