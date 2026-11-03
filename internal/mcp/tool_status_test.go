package mcputil

import (
	"context"
	"testing"
)

func TestHandleStatus_Uninitialized(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	res, env, err := handleStatus(ctx, nil, StatusInput{ProjectRoot: tmpDir})
	if err != nil {
		t.Fatalf("handleStatus failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected isError=false for uninitialized workspace, got: %v", res.IsError)
	}
	if !env.Ok || !env.Success {
		t.Errorf("expected Ok and Success true, got Ok=%v, Success=%v", env.Ok, env.Success)
	}
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any in Data, got %T", env.Data)
	}
	if data["initialized"] != false {
		t.Errorf("expected initialized=false, got %v", data["initialized"])
	}
}

func TestHandleStatus_Initialized(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize workspace
	_, envStart, err := handleStart(ctx, nil, StartInput{ProjectRoot: tmpDir, Scope: "comparison", Question: "pino vs winston"})
	if err != nil || !envStart.Ok {
		t.Fatalf("handleStart failed: %v, env=%+v", err, envStart)
	}

	res, env, err := handleStatus(ctx, nil, StatusInput{ProjectRoot: tmpDir})
	if err != nil {
		t.Fatalf("handleStatus failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected isError=false for initialized workspace, got: %v", res.IsError)
	}
	if !env.Ok || !env.Success {
		t.Errorf("expected Ok and Success true, got Ok=%v, Success=%v", env.Ok, env.Success)
	}
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any in Data, got %T", env.Data)
	}
	if data["initialized"] != true {
		t.Errorf("expected initialized=true, got %v", data["initialized"])
	}
	if data["scope"] != "comparison" {
		t.Errorf("expected scope=comparison, got %v", data["scope"])
	}
}
