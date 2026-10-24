package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSignoff_RequiresInteractiveTTY(t *testing.T) {
	// In standard go test execution, Stdin is not an interactive character device terminal.
	// runSignoff MUST strictly reject non-TTY execution with E-TTY-REQUIRED.
	err := runSignoff([]string{"D-001"})
	if err == nil {
		t.Fatalf("expected error from runSignoff in non-TTY test environment, got nil")
	}
	if !strings.Contains(err.Error(), "E-TTY-REQUIRED") {
		t.Errorf("expected error to contain E-TTY-REQUIRED, got: %v", err)
	}
}

func TestSignoff_RequiresInteractiveTTY_EvenWithForce(t *testing.T) {
	// Autonomous agent attempting to pass --force must still fail with E-TTY-REQUIRED without a real TTY.
	err := runSignoff([]string{"D-001", "--force"})
	if err == nil {
		t.Fatalf("expected error from runSignoff in non-TTY test environment even with --force, got nil")
	}
	if !strings.Contains(err.Error(), "E-TTY-REQUIRED") {
		t.Errorf("expected error to contain E-TTY-REQUIRED, got: %v", err)
	}
}

func TestSignoff_Help(t *testing.T) {
	if err := runSignoff([]string{"--help"}); err != nil {
		t.Errorf("expected nil error for --help, got: %v", err)
	}
}

func TestSignoff_ExecutionWithTestBypass(t *testing.T) {
	t.Setenv("VCK_TEST_ALLOW_NON_TTY_SIGNOFF", "1")
	tmpDir := t.TempDir()

	// Setup minimal decision workspace
	decisionsDir := filepath.Join(tmpDir, "research", "decisions")
	if err := os.MkdirAll(decisionsDir, 0755); err != nil {
		t.Fatalf("failed to create decisions dir: %v", err)
	}

	adrContent := `---
id: D-001
title: Core Storage Engine
door_type: one-way
status: proposed
review_trigger: Migration to distributed database
---
# Decision
Adopt PostgreSQL.

## Rejected Alternatives
MySQL was evaluated and rejected due to JSON querying limitations.
`
	if err := os.WriteFile(filepath.Join(decisionsDir, "D-001.md"), []byte(adrContent), 0644); err != nil {
		t.Fatalf("failed to write ADR: %v", err)
	}

	// Test signoff with --project-root and --force
	if err := runSignoff([]string{"D-001", "--force", "--project-root", tmpDir}); err != nil {
		t.Fatalf("runSignoff failed with test bypass: %v", err)
	}

	// Verify ADR was signed off
	updated, err := os.ReadFile(filepath.Join(decisionsDir, "D-001.md"))
	if err != nil {
		t.Fatalf("failed to read updated ADR: %v", err)
	}
	updatedStr := string(updated)
	if !strings.Contains(updatedStr, "human_reviewed: true") {
		t.Errorf("expected ADR to contain human_reviewed: true, got:\n%s", updatedStr)
	}
	if !strings.Contains(updatedStr, "reviewed_by:") {
		t.Errorf("expected ADR to contain reviewed_by:, got:\n%s", updatedStr)
	}
}

