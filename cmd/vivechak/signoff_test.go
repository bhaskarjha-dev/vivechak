package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSignoff_RequiresInteractiveTTY(t *testing.T) {
	t.Setenv("VCK_TEST_FORCE_NON_TTY_SIGNOFF", "1")
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
	t.Setenv("VCK_TEST_FORCE_NON_TTY_SIGNOFF", "1")
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

func TestSignoff_CISecretAuthorization(t *testing.T) {
	// Force non-TTY simulation
	t.Setenv("VCK_TEST_FORCE_NON_TTY_SIGNOFF", "1")
	t.Setenv("VCK_TEST_ALLOW_NON_TTY_SIGNOFF", "0")
	t.Setenv("VCK_APPROVER_SECRET", "")

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

	// Attempt signoff without secret
	err := runSignoff([]string{"D-002", "--project-root", tmpDir})
	if err == nil || !strings.Contains(err.Error(), "E-TTY-REQUIRED") {
		t.Fatalf("expected E-TTY-REQUIRED error without secret, got: %v", err)
	}

	// Now provision valid 16+ char secret and explicit reviewer
	t.Setenv("VCK_APPROVER_SECRET", "super-secret-approval-token-123456")
	if err := runSignoff([]string{"D-002", "--project-root", tmpDir, "--reviewer", "ci-bot@acme.corp"}); err != nil {
		t.Fatalf("expected successful signoff with valid CI secret, got: %v", err)
	}

	updated, err := os.ReadFile(filepath.Join(decisionsDir, "D-002.md"))
	if err != nil {
		t.Fatalf("failed to read updated ADR: %v", err)
	}
	updatedStr := string(updated)
	if !strings.Contains(updatedStr, "human_reviewed: true") {
		t.Errorf("expected ADR to have human_reviewed: true, got:\n%s", updatedStr)
	}
	if !strings.Contains(updatedStr, "reviewed_by: ci-bot@acme.corp") {
		t.Errorf("expected ADR to have reviewed_by: ci-bot@acme.corp, got:\n%s", updatedStr)
	}
}

func TestSignoff_IgnoresPlanFilesAndSignsADR(t *testing.T) {
	t.Setenv("VCK_TEST_ALLOW_NON_TTY_SIGNOFF", "1")

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

	if err := runSignoff([]string{"D-001", "--project-root", tmpDir, "--force", "--reviewer", "Lead Architect"}); err != nil {
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


