package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func TestHandleGate_Uninitialized(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	res, env, err := handleGate(ctx, nil, GateInput{
		ProjectRoot: tmpDir,
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !res.IsError || env.Ok {
		t.Fatalf("expected error result on uninitialized workspace, got res.IsError=%v, env=%+v", res.IsError, env)
	}
}

func TestHandleGate_ComparisonScope(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize comparison workspace
	_, envStart, err := handleStart(ctx, nil, StartInput{
		ProjectRoot: tmpDir,
		Scope:       "comparison",
		Question:    "Pino vs Winston",
	})
	if err != nil || !envStart.Ok {
		t.Fatalf("handleStart failed: %v, env=%+v", err, envStart)
	}

	// Run gate evaluation on comparison scope (gate is not applicable / 0 requirements)
	res, env, err := handleGate(ctx, nil, GateInput{
		ProjectRoot: tmpDir,
		Verbose:     true,
	})
	if err != nil {
		t.Fatalf("handleGate returned unexpected go error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected gate not to fail execution, got error message: %s", env.Message)
	}
	if !env.Ok {
		t.Errorf("expected Ok=true and Success=true, got Ok=%v, Success=%v", env.Ok, env.Ok)
	}
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any data, got %T", env.Data)
	}
	if data["verdict"] != "not_applicable" {
		t.Errorf("expected comparison scope verdict not_applicable, got %v", data["verdict"])
	}
}

func TestHandleGate_ProjectScope(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize project workspace
	_, envStart, err := handleStart(ctx, nil, StartInput{
		ProjectRoot: tmpDir,
		Scope:       "project",
		Question:    "Architect scalable event ingestion platform",
	})
	if err != nil || !envStart.Ok {
		t.Fatalf("handleStart failed: %v, env=%+v", err, envStart)
	}

	// Run gate evaluation on newly initialized project scope (expected to report status/issues without crash)
	res, env, err := handleGate(ctx, nil, GateInput{
		ProjectRoot: tmpDir,
		Verbose:     true,
	})
	if err != nil {
		t.Fatalf("handleGate returned unexpected go error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error response: %s", env.Message)
	}
	if !env.Ok {
		t.Fatalf("expected envelope Ok and Success true, got Ok=%v, Success=%v", env.Ok, env.Ok)
	}

	// Verify that PHASE-0-GATE.md was rendered and written
	gatePath := filepath.Join(tmpDir, core.GateFile)
	if _, err := os.Stat(gatePath); os.IsNotExist(err) {
		t.Errorf("expected %s to be created, but was not found", gatePath)
	}
}

// A comparison followed by an accepted one-way ADR FAILs the gate (CRIT-12); the response must then not
// carry the comparison scope's "Proceed to implementation" advisory, which belongs to NOT_APPLICABLE only.
func TestHandleGate_ComparisonThenOneWayADR_NoProceedAdvisory(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	research := filepath.Join(tmpDir, core.ResearchDir)
	if err := os.MkdirAll(research, 0o755); err != nil {
		t.Fatal(err)
	}
	cmp := "---\nid: C-001\ntype: comparison\ntitle: Pino vs Winston\nstatus: complete\ndate: 2026-11-12\nvivechak: 1\n---\n# Pino vs Winston\n"
	adr := "---\nid: D-001\ntype: decision\ntitle: Event store\nstatus: accepted\ndoor_type: one-way\ndate: 2026-11-12\nreview_trigger: \"p95 write latency above 50 ms\"\nvivechak: 1\n---\n# Context\nStorage choice.\n# Decision\nUse PostgreSQL.\n"
	if err := os.WriteFile(filepath.Join(research, "C-001-comparison.md"), []byte(cmp), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(research, "D-001-decision.md"), []byte(adr), 0o644); err != nil {
		t.Fatal(err)
	}

	_, env, err := handleGate(ctx, nil, GateInput{ProjectRoot: tmpDir})
	if err != nil {
		t.Fatalf("handleGate returned unexpected go error: %v", err)
	}
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any data, got %T", env.Data)
	}
	if data["verdict"] != "FAIL" {
		t.Fatalf("an accepted, unsigned one-way ADR after a comparison must FAIL the gate, got %v (%s)", data["verdict"], env.Message)
	}
	for _, w := range envWarningStrings(env) {
		if strings.Contains(w, "Proceed to implementation") {
			t.Errorf("a FAIL must not advise proceeding to implementation, got warning %q", w)
		}
	}
}
