package mcputil

import (
	"context"
	"os"
	"path/filepath"
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
	if !env.Ok || !env.Success {
		t.Errorf("expected Ok=true and Success=true, got Ok=%v, Success=%v", env.Ok, env.Success)
	}
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any data, got %T", env.Data)
	}
	if data["gate_passed"] != true {
		t.Errorf("expected comparison scope gate_passed=true, got %v", data["gate_passed"])
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
	if !env.Ok || !env.Success {
		t.Fatalf("expected envelope Ok and Success true, got Ok=%v, Success=%v", env.Ok, env.Success)
	}

	// Verify that PHASE-0-GATE.md was rendered and written
	gatePath := filepath.Join(tmpDir, core.GateFile)
	if _, err := os.Stat(gatePath); os.IsNotExist(err) {
		t.Errorf("expected %s to be created, but was not found", gatePath)
	}
}
