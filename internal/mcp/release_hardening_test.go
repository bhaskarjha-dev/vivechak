package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSavePlanComparisonTerminalNextStep(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "comparison",
			"decision_id":  "C-01",
			"content":      "# Weighted Evaluation Protocol\n\n| Option | Score |\n|---|---|\n| Pino | 8.5 [A] (pino.dev) |\n| Winston | 7.0 [B] (github.com) |\n",
		},
	})
	if err != nil {
		t.Fatalf("save_plan comparison: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("expected success, got error: %s", env.Message)
	}
	if !strings.Contains(env.NextStep, "Comparison scope is complete") {
		t.Errorf("expected terminal next_step for comparison scope, got: %s", env.NextStep)
	}
	if strings.Contains(env.NextStep, "vivechak_next_session") {
		t.Errorf("comparison scope next_step must not mandate vivechak_next_session")
	}
}

func TestSavePlanAutoInitGitignore(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir() // Fresh directory without vivechak_init

	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "decision",
			"decision_id":  "D-01",
			"content":      "# Decision Plan\n\n## S1: Landscape\n## S2: Comparison\n",
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}
	gitIgnorePath := filepath.Join(tmpDir, "research", ".gitignore")
	data, err := os.ReadFile(gitIgnorePath)
	if err != nil {
		t.Fatalf("expected research/.gitignore to exist on auto-init: %v", err)
	}
	if !strings.Contains(string(data), "*.lock") {
		t.Errorf("expected .gitignore to ignore *.lock, got: %s", string(data))
	}
}

func TestRunGateTrackFiltering(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize project workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// 1. Test invalid track parameter returns error
	resErr, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"track":        "invalid_track",
		},
	})
	if err != nil {
		t.Fatalf("run_gate call: %v", err)
	}
	envErr := parseEnvelope(t, resErr)
	if envErr.Success {
		t.Errorf("expected failure envelope on invalid track parameter, got success")
	}

	// 2. Test track_a filter on empty project
	resA, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"track":        "track_a",
		},
	})
	if err != nil {
		t.Fatalf("run_gate track_a: %v", err)
	}
	envA := parseEnvelope(t, resA)
	dataA := envA.Data.(map[string]any)
	if dataA["track_filter"] != "track_a" {
		t.Errorf("expected track_filter == 'track_a', got: %v", dataA["track_filter"])
	}
	if !strings.Contains(envA.Message, "Track A requirements not met") {
		t.Errorf("expected Track A failure message, got: %s", envA.Message)
	}

	// 3. Test track_b filter on empty project
	resB, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"track":        "track_b",
		},
	})
	if err != nil {
		t.Fatalf("run_gate track_b: %v", err)
	}
	envB := parseEnvelope(t, resB)
	dataB := envB.Data.(map[string]any)
	if dataB["track_filter"] != "track_b" {
		t.Errorf("expected track_filter == 'track_b', got: %v", dataB["track_filter"])
	}
	if !strings.Contains(envB.Message, "Track B requirements not met") {
		t.Errorf("expected Track B failure message, got: %s", envB.Message)
	}
}

func TestAmendADRRecompilesRegistry(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "decision",
		},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Record D-01 ADR
	adrContent := "---\ndecision_id: D-01\ntitle: Storage Engine\ndoor_type: one-way\nstatus: accepted\nreview_trigger: latency > 50ms\n---\n\n## Context\nInitial rationale for Postgres storage.\n"
	resRec, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"decision_id":  "D-01",
			"content":      adrContent,
		},
	})
	if err != nil {
		t.Fatalf("record_decision: %v", err)
	}
	envRec := parseEnvelope(t, resRec)
	if !envRec.Success {
		t.Fatalf("record_decision failed: %s", envRec.Message)
	}

	// Verify DECISIONS.md contains initial rationale
	decPath := filepath.Join(tmpDir, "research", "DECISIONS.md")
	decInitial, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("reading DECISIONS.md: %v", err)
	}
	if !strings.Contains(string(decInitial), "Initial rationale for Postgres storage") {
		t.Fatalf("expected initial rationale in DECISIONS.md")
	}

	// Amend D-01 via vivechak_amend_session
	amendment := "Benchmark update: ClickHouse evaluated for analytical workloads."
	resAmend, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_amend_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "D-01",
			"content":      amendment,
		},
	})
	if err != nil {
		t.Fatalf("amend_session: %v", err)
	}
	envAmend := parseEnvelope(t, resAmend)
	if !envAmend.Success {
		t.Fatalf("amend_session failed: %s", envAmend.Message)
	}
	amendData := envAmend.Data.(map[string]any)
	if amendData["registry_recompiled"] != core.DecisionsFile {
		t.Errorf("expected registry_recompiled in data, got: %v", amendData["registry_recompiled"])
	}

	// Verify DECISIONS.md was recompiled and contains amendment
	decUpdated, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("reading updated DECISIONS.md: %v", err)
	}
	if !strings.Contains(string(decUpdated), amendment) {
		t.Errorf("expected amended text in recompiled DECISIONS.md, got:\n%s", string(decUpdated))
	}
}

func TestStatusMultiScopeNextStep(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	// 1. Test Comparison Scope status after WEP save
	cmpDir := t.TempDir()
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": cmpDir,
			"scope":        "comparison",
			"decision_id":  "C-01",
			"content":      "# Weighted Evaluation Protocol\n\n| Option | Score |\n|---|---|\n| Pino | 8.5 [A] |\n",
		},
	})
	resCmpStatus, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_status",
		Arguments: map[string]any{
			"project_root": cmpDir,
		},
	})
	if err != nil {
		t.Fatalf("status cmp: %v", err)
	}
	envCmpStatus := parseEnvelope(t, resCmpStatus)
	if !strings.Contains(envCmpStatus.NextStep, "Comparison scope is finished") {
		t.Errorf("expected finished nextStep for comparison scope status, got: %s", envCmpStatus.NextStep)
	}
	if strings.Contains(envCmpStatus.NextStep, "vivechak_run_gate") {
		t.Errorf("comparison scope status must never mandate vivechak_run_gate")
	}

	// 2. Test Decision Scope status after ADR record
	decDir := t.TempDir()
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": decDir,
			"scope":        "decision",
		},
	})
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": decDir,
			"decision_id":  "D-01",
			"content":      "---\ndecision_id: D-01\ntitle: Caching\ndoor_type: two-way\nstatus: accepted\n---\n\n## Context\nRedis choice.\n",
		},
	})
	resDecStatus, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_status",
		Arguments: map[string]any{
			"project_root": decDir,
		},
	})
	if err != nil {
		t.Fatalf("status dec: %v", err)
	}
	envDecStatus := parseEnvelope(t, resDecStatus)
	if !strings.Contains(envDecStatus.NextStep, "Decision scope is complete") {
		t.Errorf("expected complete nextStep for decision scope with recorded ADR, got: %s", envDecStatus.NextStep)
	}
}

func TestComparisonScopeSessionCountingAndGate(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	cmpDir := t.TempDir()

	// Save comparison WEP matrix directly in research/C-01-comparison.md
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": cmpDir,
			"scope":        "comparison",
			"decision_id":  "C-01",
			"content":      "# Weighted Evaluation Protocol\n\n| Option | Score |\n|---|---|\n| Pino | 8.5 [Grade: A] (https://github.com/pinojs/pino) |\n| Winston | 7.0 [Grade: B] (https://github.com/winstonjs/winston) |\n",
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}

	// Inspect workspace
	info := core.InspectWorkspace(cmpDir)
	if info.Scope != core.ScopeComparison {
		t.Errorf("expected ScopeComparison, got: %s", info.Scope)
	}
	if info.SessionCount != 1 {
		t.Errorf("expected SessionCount == 1 for comparison file in research/, got: %d", info.SessionCount)
	}

	// If gate is run on comparison scope with evidence grades, it should evaluate Check B2 cleanly
	resGate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": cmpDir,
			"verbose":      true,
		},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	envGate := parseEnvelope(t, resGate)
	dataGate := envGate.Data.(map[string]any)
	if dataGate["scope"] != "comparison" {
		t.Errorf("expected scope 'comparison', got: %v", dataGate["scope"])
	}
	qb := dataGate["quality_checks"].(map[string]any)
	if qb["passed"] != true {
		t.Errorf("expected quality checks to pass on comparison matrix with grades, got: %v (issues: %v)", qb["passed"], qb["issues"])
	}
}

func TestRunGateTrackPassingIsolation(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize project workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Structural completeness: pipeline, session, decisions, FAD
	pipelineContent := "# RESEARCH PIPELINE\n\n## SESSIONS\n\n#### T1-01: Ingestion\n#### SYN-01: Synthesis\n"
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "RESEARCH-PIPELINE.md"), []byte(pipelineContent), 0o644)

	s1 := "---\nsession_id: T1-01\ntitle: Ingestion\nstatus: complete\n---\nCompleted findings.\n"
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "sessions", "T1-01.md"), []byte(s1), 0o644)

	dec := "# Architectural Decisions\n\n## D-01: Decision\n"
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "DECISIONS.md"), []byte(dec), 0o644)

	// FAD without evidence grades and only 1 session => Track B quality checks fail (< 3 sessions, no grades in FAD)
	fad := "# Founding Architecture Document\n\nArchitecture description without grades.\n"
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "FAD.md"), []byte(fad), 0o644)

	// 1. Evaluate with track_a
	resA, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"track":        "track_a",
		},
	})
	if err != nil {
		t.Fatalf("run_gate track_a: %v", err)
	}
	envA := parseEnvelope(t, resA)
	dataA := envA.Data.(map[string]any)
	if dataA["gate_passed"] != true {
		t.Errorf("expected track_a to pass structurally, got: %v (msg: %s)", dataA["gate_passed"], envA.Message)
	}
	if dataA["gate_status"] != "PASS" {
		t.Errorf("expected gate_status PASS for track_a, got: %v", dataA["gate_status"])
	}

	// 2. Evaluate with track_b
	resB, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"track":        "track_b",
		},
	})
	if err != nil {
		t.Fatalf("run_gate track_b: %v", err)
	}
	envB := parseEnvelope(t, resB)
	dataB := envB.Data.(map[string]any)
	if dataB["gate_passed"] != false {
		t.Errorf("expected track_b to fail quality checks, got: %v", dataB["gate_passed"])
	}
	if dataB["gate_status"] != "FAIL" {
		t.Errorf("expected gate_status FAIL for track_b, got: %v", dataB["gate_status"])
	}
	if !strings.Contains(envB.Message, "Track B requirements not met") {
		t.Errorf("expected 'Track B requirements not met', got: %s", envB.Message)
	}
}

func TestSavePlanGitignorePreservesExisting(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	researchDir := filepath.Join(tmpDir, "research")
	_ = os.MkdirAll(researchDir, 0o755)
	existingContent := "# Custom Team Ignore\n*.secret\n*.env\n"
	gitIgnorePath := filepath.Join(researchDir, ".gitignore")
	if err := os.WriteFile(gitIgnorePath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("writing .gitignore: %v", err)
	}

	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "comparison",
			"decision_id":  "C-01",
			"content":      "# WEP Matrix\n\n| Option | Score |\n|---|---|\n| Pino | 8.5 [A] |\n",
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}

	readBack, err := os.ReadFile(gitIgnorePath)
	if err != nil {
		t.Fatalf("reading .gitignore: %v", err)
	}
	if string(readBack) != existingContent {
		t.Errorf("expected existing .gitignore to be preserved exactly, got: %s", string(readBack))
	}
}

func TestAmendNonADRDoesNotRecompileRegistry(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
		},
	})

	// Save session T1-01
	sessContent := "---\nsession_id: T1-01\ntitle: Test Session\nstatus: complete\n---\nOriginal text.\n"
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-01",
			"content":      sessContent,
		},
	})

	// Amend session T1-01
	resAmend, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_amend_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-01",
			"content":      "Amended observation.",
		},
	})
	if err != nil {
		t.Fatalf("amend_session: %v", err)
	}
	envAmend := parseEnvelope(t, resAmend)
	dataMap := envAmend.Data.(map[string]any)
	if dataMap["registry_recompiled"] != nil {
		t.Errorf("non-ADR session amendment must not trigger registry recompilation, got: %v", dataMap["registry_recompiled"])
	}
}

func TestInspectWorkspaceMultipleComparisonFiles(t *testing.T) {
	tmpDir := t.TempDir()
	resDir := filepath.Join(tmpDir, "research")
	_ = os.MkdirAll(resDir, 0o755)

	c1 := "# Comparison 1\n| Pino | 8.5 [Grade: A] |\n"
	c2 := "# Comparison 2\n| Zod | 9.0 [Grade: B] |\n"
	_ = os.WriteFile(filepath.Join(resDir, "C-01-comparison.md"), []byte(c1), 0o644)
	_ = os.WriteFile(filepath.Join(resDir, "C-02-cmp-validation.md"), []byte(c2), 0o644)

	info := core.InspectWorkspace(tmpDir)
	if info.Scope != core.ScopeComparison {
		t.Errorf("expected ScopeComparison, got: %s", info.Scope)
	}
	if info.SessionCount != 2 {
		t.Errorf("expected SessionCount == 2 for two comparison files, got: %d", info.SessionCount)
	}
}

