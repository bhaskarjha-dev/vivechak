package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func TestHandleReview_DryRunValidate(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	validSessionContent := `---
id: T1-01
title: Cache Architecture
status: complete
---
# T1-01: Cache Architecture

## Executive Summary
Redis Sentinel is selected for high availability caching.

## Discovered Concerns
Memory footprint during burst key invalidation requires eviction policy.

## Evidence Matrix
| Finding | Grade | Provenance |
|---|---|---|
| Sentinel failover sub-second | E-A | [redis.io/docs](https://redis.io) |
`

	res, env, err := handleReview(ctx, nil, ReviewInput{
		ProjectRoot:  tmpDir,
		Mode:         "validate",
		ArtifactType: "session",
		Content:      validSessionContent,
	})
	if err != nil {
		t.Fatalf("handleReview failed with go error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected validate not to error, got message: %s", env.Message)
	}
	if !env.Ok {
		t.Errorf("expected Ok=true, Success=true, got Ok=%v, Success=%v", env.Ok, env.Ok)
	}
}

func TestHandleReview_WorkspaceValidate(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Uninitialized workspace check
	res, env, err := handleReview(ctx, nil, ReviewInput{
		ProjectRoot: tmpDir,
		Mode:        "validate",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !res.IsError || env.Ok {
		t.Fatalf("expected error on uninitialized workspace, got res.IsError=%v, env=%+v", res.IsError, env)
	}

	// 2. Initialize workspace
	_, envStart, err := handleStart(ctx, nil, StartInput{
		ProjectRoot: tmpDir,
		Scope:       "comparison",
		Question:    "Valkey vs Redis",
	})
	if err != nil || !envStart.Ok {
		t.Fatalf("handleStart failed: %v, env=%+v", err, envStart)
	}

	// 3. Workspace validate on initialized workspace
	res, env, err = handleReview(ctx, nil, ReviewInput{
		ProjectRoot: tmpDir,
		Mode:        "validate",
	})
	if err != nil {
		t.Fatalf("handleReview failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success on initialized workspace review, got: %s", env.Message)
	}
	if !env.Ok {
		t.Errorf("expected Ok=true and Success=true, got Ok=%v, Success=%v", env.Ok, env.Ok)
	}
}

func TestHandleReview_Challenge(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize workspace
	_, envStart, err := handleStart(ctx, nil, StartInput{
		ProjectRoot: tmpDir,
		Scope:       "decision",
		Question:    "Database storage engine choice",
	})
	if err != nil || !envStart.Ok {
		t.Fatalf("handleStart failed: %v, env=%+v", err, envStart)
	}

	// Create session file
	sessDir := filepath.Join(tmpDir, core.SessionsDir)
	_ = os.MkdirAll(sessDir, 0o755)
	sessContent := `---
id: T1-01
title: Storage Engine
---
# T1-01: Storage Engine

## Recommendation
SQLite in WAL mode.

## Alternatives Considered
| Option | Verdict |
|---|---|
| PostgreSQL | Rejected |
`
	_ = os.WriteFile(filepath.Join(sessDir, "T1-01-storage.md"), []byte(sessContent), 0o644)

	// Run review in challenge mode
	res, env, err := handleReview(ctx, nil, ReviewInput{
		ProjectRoot: tmpDir,
		Mode:        "red_team",
		ID:          "T1-01",
	})
	if err != nil {
		t.Fatalf("handleReview challenge failed with go error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected challenge mode success, got error: %s", env.Message)
	}
	if !env.Ok {
		t.Errorf("expected Ok=true and Success=true, got Ok=%v, Success=%v", env.Ok, env.Ok)
	}
}

func TestHandleReview_DecisionPlanValidateAndUnrecordedDecisionRedTeam(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, envStart, err := handleStart(ctx, nil, StartInput{
		ProjectRoot: tmpDir,
		Scope:       "decision",
		Question:    "PostgreSQL vs ClickHouse for event analytics",
	})
	if err != nil || !envStart.Ok {
		t.Fatalf("handleStart failed: %v, env=%+v", err, envStart)
	}

	planContent := "# D-001-PLAN: Event Storage Engine\n\n```yaml vivechak-plan\nplan: D-001\ndecisions:\n  - id: D-001\n    title: Event Storage Engine\n    door: one-way\nsessions:\n  - id: D-001-S1\n    title: Landscape\n    layer: 1\n    role: L\n    door: one-way\n    informs: [D-001]\n    depends: []\n```\n\n## Sessions\n\n### D-001-S1: Landscape\n```prompt\nCompare PostgreSQL and ClickHouse.\n```\n"
	if _, envSave, err := handleSave(ctx, nil, SaveInput{
		ProjectRoot: tmpDir,
		Kind:        "plan",
		ID:          "D-001",
		Content:     planContent,
	}); err != nil || !envSave.Ok {
		t.Fatalf("save plan failed: %v, env=%+v", err, envSave)
	}

	// 1. Validating D-001-plan by ID must locate research/D-001-plan.md as a plan artifact.
	resValPlan, envValPlan, err := handleReview(ctx, nil, ReviewInput{
		ProjectRoot: tmpDir,
		Mode:        "validate",
		ID:          "D-001-plan",
	})
	if err != nil || resValPlan.IsError || !envValPlan.Ok {
		t.Fatalf("validate D-001-plan failed: err=%v env=%+v", err, envValPlan)
	}
	valData, _ := envValPlan.Data.(map[string]any)
	if valData["artifact_type"] != "plan" || valData["valid"] != true {
		t.Fatalf("expected valid plan validation result, got: %+v", valData)
	}

	// 2. Saving an informing session must NOT allow red_team(id="D-001") before the ADR exists.
	sessContent := "---\nid: D-001-S1\ntitle: Landscape\ninforms_decisions: [D-001]\ndoor_type: one-way\nstatus: complete\n---\n# D-001-S1: Landscape\n\n## Findings\n- ClickHouse MergeTree [Grade A: https://clickhouse.com/docs | fetched].\n\n## Discovered Concerns\nClickHouse mutations rewrite parts asynchronously and do not suit OLTP row updates.\n\n## Recommendation\nUse ClickHouse for append-only event analytics.\n"
	if _, envSess, err := handleSave(ctx, nil, SaveInput{
		ProjectRoot: tmpDir,
		Kind:        "session",
		ID:          "D-001-S1",
		Content:     sessContent,
	}); err != nil || !envSess.Ok {
		t.Fatalf("save session failed: %v, env=%+v", err, envSess)
	}

	resPremature, envPremature, err := handleReview(ctx, nil, ReviewInput{
		ProjectRoot: tmpDir,
		Mode:        "red_team",
		ID:          "D-001",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !resPremature.IsError || envPremature.Code != "E-NOT-FOUND" {
		t.Fatalf("expected E-NOT-FOUND on red_team(D-001) before ADR exists, got isError=%v code=%q msg=%q", resPremature.IsError, envPremature.Code, envPremature.Message)
	}
	if _, statErr := os.Stat(filepath.Join(tmpDir, core.ResearchDir, "D-001-R1.md")); !os.IsNotExist(statErr) {
		t.Fatalf("expected no D-001-R1.md written when ADR is absent, got stat err=%v", statErr)
	}
}
