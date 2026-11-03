package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func TestHandleAmend_Validation(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Missing target ID
	res, env, err := handleAmend(ctx, nil, AmendInput{
		ProjectRoot: tmpDir,
		SessionID:   "",
		Content:     "Some amendment",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !res.IsError || env.Ok {
		t.Fatalf("expected error for missing target ID, got env=%+v", env)
	}

	// 2. Invalid target ID format
	res, env, err = handleAmend(ctx, nil, AmendInput{
		ProjectRoot: tmpDir,
		SessionID:   "invalid..id!",
		Content:     "Some amendment",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !res.IsError || env.Ok {
		t.Fatalf("expected error for invalid target ID, got env=%+v", env)
	}

	// 3. Missing amendment content
	res, env, err = handleAmend(ctx, nil, AmendInput{
		ProjectRoot: tmpDir,
		SessionID:   "T1-01",
		Content:     "",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !res.IsError || env.Ok {
		t.Fatalf("expected error for missing content, got env=%+v", env)
	}
}

func TestHandleAmend_Success(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize workspace with start
	_, envStart, err := handleStart(ctx, nil, StartInput{
		ProjectRoot: tmpDir,
		Scope:       "decision",
		Question:    "Database choice for high-throughput messaging",
	})
	if err != nil || !envStart.Ok {
		t.Fatalf("handleStart failed: %v, env=%+v", err, envStart)
	}

	// Create a dummy session file in research/sessions/T1-01-database.md
	sessDir := filepath.Join(tmpDir, core.SessionsDir)
	if err := os.MkdirAll(sessDir, 0o755); err != nil {
		t.Fatalf("failed to create sessions dir: %v", err)
	}
	sessFile := filepath.Join(sessDir, "T1-01-database.md")
	initialContent := "---\nid: T1-01\ntitle: Database Evaluation\n---\n# T1-01: Database Evaluation\n\nInitial findings suggest PostgreSQL.\n"
	if err := os.WriteFile(sessFile, []byte(initialContent), 0o644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	// Amend the session
	amendmentNote := "PostgreSQL connection limits under spike loads require PgBouncer pooler."
	res, env, err := handleAmend(ctx, nil, AmendInput{
		ProjectRoot: tmpDir,
		SessionID:   "T1-01",
		Content:     amendmentNote,
		Reason:      "Spike load connection exhaustion discovery",
	})
	if err != nil {
		t.Fatalf("handleAmend failed: %v", err)
	}
	if res.IsError || !env.Ok || !env.Success {
		t.Fatalf("expected success, got res.IsError=%v, env=%+v", res.IsError, env)
	}

	// Verify the file content contains the amendment
	updatedBytes, err := os.ReadFile(sessFile)
	if err != nil {
		t.Fatalf("failed to read amended file: %v", err)
	}
	updatedStr := string(updatedBytes)
	if !strings.Contains(updatedStr, amendmentNote) {
		t.Errorf("amended file does not contain amendment note; content:\n%s", updatedStr)
	}
	if !strings.Contains(updatedStr, "Spike load connection exhaustion discovery") {
		t.Errorf("amended file does not contain reason; content:\n%s", updatedStr)
	}
}
