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
	if !env.Ok || !env.Success {
		t.Errorf("expected Ok=true, Success=true, got Ok=%v, Success=%v", env.Ok, env.Success)
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
	if !env.Ok || !env.Success {
		t.Errorf("expected Ok=true and Success=true, got Ok=%v, Success=%v", env.Ok, env.Success)
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
		ProjectRoot:   tmpDir,
		Mode:          "challenge",
		ChallengeType: "red_team",
		Target:        "T1-01",
	})
	if err != nil {
		t.Fatalf("handleReview challenge failed with go error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected challenge mode success, got error: %s", env.Message)
	}
	if !env.Ok || !env.Success {
		t.Errorf("expected Ok=true and Success=true, got Ok=%v, Success=%v", env.Ok, env.Success)
	}
}
