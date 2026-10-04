package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func TestHandleChallenge_RedTeam(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, core.SessionsDir)
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("failed to create sessions dir: %v", err)
	}

	sessionContent := `---
id: T1-01
title: Database Selection
---
# T1-01: Database Selection

## Recommendation
SQLite in WAL mode is recommended for embedded storage.

## Alternatives Considered
| Option | Verdict | Tradeoff |
|---|---|---|
| SQLite WAL | Recommended | Zero-dependency |
| PostgreSQL | Rejected | Heavy daemon overhead |
`
	if err := os.WriteFile(filepath.Join(sessionsDir, "T1-01-database.md"), []byte(sessionContent), 0o644); err != nil {
		t.Fatalf("failed to write session: %v", err)
	}

	ctx := context.Background()
	res, env, err := handleChallenge(ctx, nil, ChallengeInput{
		ProjectRoot: tmpDir,
		SessionID:   "T1-01",
		Mode:        "red_team",
	})
	if err != nil {
		t.Fatalf("handleChallenge failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got error: %s", env.Message)
	}

	data, ok := env.Data.(*core.ChallengeResult)
	if !ok {
		t.Fatalf("expected *core.ChallengeResult, got %T", env.Data)
	}
	if len(data.Prompts) != 4 {
		t.Errorf("expected 4 prompts, got %d", len(data.Prompts))
	}
	if env.NextStep == "" {
		t.Errorf("expected non-empty next_step")
	}
}

func TestHandleChallenge_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755); err != nil {
		t.Fatalf("failed to create sessions dir: %v", err)
	}

	ctx := context.Background()
	res, env, err := handleChallenge(ctx, nil, ChallengeInput{
		ProjectRoot: tmpDir,
		SessionID:   "T9-99",
	})
	if err != nil {
		t.Fatalf("unexpected handler err: %v", err)
	}
	if !res.IsError {
		t.Errorf("expected error result for missing session")
	}
	if env.Message == "" {
		t.Errorf("expected error message in envelope")
	}
}
