package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSignoff_RequiresInteractiveTTY(t *testing.T) {
	var stdout bytes.Buffer
	err := runSignoffWithStreams([]string{"D-001"}, nil, &stdout, func() bool { return false })
	if err == nil {
		t.Fatalf("expected error from runSignoff in non-TTY test environment, got nil")
	}
	if !strings.Contains(err.Error(), "E-TTY-REQUIRED") {
		t.Errorf("expected error to contain E-TTY-REQUIRED, got: %v", err)
	}
}

func TestSignoff_RequiresInteractiveTTY_EvenWithForce(t *testing.T) {
	var stdout bytes.Buffer
	err := runSignoffWithStreams([]string{"D-001", "--force"}, nil, &stdout, func() bool { return false })
	if err == nil {
		t.Fatalf("expected error from runSignoff in non-TTY test environment even with --force, got nil")
	}
	if !strings.Contains(err.Error(), "E-TTY-REQUIRED") {
		t.Errorf("expected error to contain E-TTY-REQUIRED, got: %v", err)
	}
}

func TestSignoff_Help(t *testing.T) {
	var stdout bytes.Buffer
	if err := runSignoffWithStreams([]string{"--help"}, nil, &stdout, func() bool { return false }); err != nil {
		t.Errorf("expected nil error for --help, got: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage: vivechak signoff") {
		t.Errorf("expected help output, got: %s", stdout.String())
	}
}

func TestSignoff_InteractiveConfirmation_Success(t *testing.T) {
	tmpDir := t.TempDir()
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

	var stdout bytes.Buffer
	stdin := strings.NewReader("CONFIRM\n")
	err := runSignoffWithStreams([]string{"D-001", "--project-root", tmpDir, "--reviewer", "Alice"}, stdin, &stdout, func() bool { return true })
	if err != nil {
		t.Fatalf("expected successful signoff, got: %v", err)
	}

	updated, err := os.ReadFile(filepath.Join(decisionsDir, "D-001.md"))
	if err != nil {
		t.Fatalf("failed to read updated ADR: %v", err)
	}
	updatedStr := string(updated)
	if !strings.Contains(updatedStr, "human_reviewed: true") {
		t.Errorf("expected ADR to contain human_reviewed: true, got:\n%s", updatedStr)
	}
	if !strings.Contains(updatedStr, "reviewed_by: Alice") {
		t.Errorf("expected ADR to contain reviewed_by: Alice, got:\n%s", updatedStr)
	}
}

func TestSignoff_InteractiveConfirmation_Cancelled(t *testing.T) {
	tmpDir := t.TempDir()
	decisionsDir := filepath.Join(tmpDir, "research", "decisions")
	_ = os.MkdirAll(decisionsDir, 0755)

	adrContent := `---
id: D-001
title: Core Storage Engine
door_type: one-way
status: proposed
---
# Decision
`
	_ = os.WriteFile(filepath.Join(decisionsDir, "D-001.md"), []byte(adrContent), 0644)

	var stdout bytes.Buffer
	stdin := strings.NewReader("NO\n")
	err := runSignoffWithStreams([]string{"D-001", "--project-root", tmpDir}, stdin, &stdout, func() bool { return true })
	if err == nil || !strings.Contains(err.Error(), "signoff cancelled: expected 'CONFIRM', got \"NO\"") {
		t.Fatalf("expected cancellation error, got: %v", err)
	}
}

func TestSignoff_ForceFlagWithTerminal(t *testing.T) {
	tmpDir := t.TempDir()
	decisionsDir := filepath.Join(tmpDir, "research", "decisions")
	_ = os.MkdirAll(decisionsDir, 0755)

	adrContent := `---
id: D-002
title: Security Boundary
door_type: one-way
status: proposed
review_trigger: Annual audit
---
# Decision
`
	_ = os.WriteFile(filepath.Join(decisionsDir, "D-002.md"), []byte(adrContent), 0644)

	var stdout bytes.Buffer
	err := runSignoffWithStreams([]string{"D-002", "--project-root", tmpDir, "--force", "--reviewer", "ci-admin@acme.corp"}, nil, &stdout, func() bool { return true })
	if err != nil {
		t.Fatalf("expected successful signoff with force and terminal, got: %v", err)
	}

	updated, err := os.ReadFile(filepath.Join(decisionsDir, "D-002.md"))
	if err != nil {
		t.Fatalf("failed to read updated ADR: %v", err)
	}
	updatedStr := string(updated)
	if !strings.Contains(updatedStr, "human_reviewed: true") {
		t.Errorf("expected ADR to have human_reviewed: true, got:\n%s", updatedStr)
	}
	if !strings.Contains(updatedStr, "reviewed_by: ci-admin@acme.corp") {
		t.Errorf("expected ADR to have reviewed_by: ci-admin@acme.corp, got:\n%s", updatedStr)
	}
}

func TestSignoff_IgnoresPlanFilesAndSignsADR(t *testing.T) {
	tmpDir := t.TempDir()
	rDir := filepath.Join(tmpDir, "research")
	_ = os.MkdirAll(rDir, 0755)

	// Plan file starts with D-001-
	planContent := `# D-001 Research Plan
1 session spike.
`
	_ = os.WriteFile(filepath.Join(rDir, "D-001-clinic-datastore-plan.md"), []byte(planContent), 0644)

	// Real ADR file
	adrContent := `---
id: D-001
title: Datastore Architecture
status: accepted
door_type: one-way
human_reviewed: false
---
# Context
Storage decision.
`
	_ = os.WriteFile(filepath.Join(rDir, "D-001-clinic-datastore.md"), []byte(adrContent), 0644)

	var stdout bytes.Buffer
	if err := runSignoffWithStreams([]string{"D-001", "--project-root", tmpDir, "--force", "--reviewer", "Lead Architect"}, nil, &stdout, func() bool { return true }); err != nil {
		t.Fatalf("runSignoff failed: %v", err)
	}

	// Verify plan file was NOT touched
	planAfter, _ := os.ReadFile(filepath.Join(rDir, "D-001-clinic-datastore-plan.md"))
	if strings.Contains(string(planAfter), "human_reviewed") {
		t.Errorf("plan file was mistakenly signed off!")
	}

	// Verify real ADR was signed off
	adrAfter, _ := os.ReadFile(filepath.Join(rDir, "D-001-clinic-datastore.md"))
	if !strings.Contains(string(adrAfter), "human_reviewed: true") {
		t.Errorf("expected real ADR to have human_reviewed: true, got:\n%s", string(adrAfter))
	}
}
