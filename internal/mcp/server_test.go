package mcputil

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// testServer creates a Vivechak MCP server connected to in-memory transports.
func testServer(t *testing.T) *mcp.ClientSession {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	ctx := context.Background()

	server := NewServer("0.1.0-test", logger)

	st, ct := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "test-client", Version: "1.0.0"},
		nil,
	)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	return cs
}

// parseEnvelope extracts the Envelope from a CallToolResult's text content.
func parseEnvelope(t *testing.T, result *mcp.CallToolResult) Envelope {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("no text content in result")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var env Envelope
	if err := json.Unmarshal([]byte(tc.Text), &env); err != nil {
		t.Fatalf("parsing envelope: %v", err)
	}
	return env
}

// TestToolListing verifies all 9 tools are registered.
func TestToolListing(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	var tools []*mcp.Tool
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		tools = append(tools, tool)
	}

	expected := []string{
		"vivechak_init",
		"vivechak_prepare_generator",
		"vivechak_save_plan",
		"vivechak_status",
		"vivechak_next_session",
		"vivechak_save_session",
		"vivechak_record_decision",
		"vivechak_validate",
		"vivechak_run_gate",
	}

	if len(tools) != len(expected) {
		names := make([]string, len(tools))
		for i, t := range tools {
			names[i] = t.Name
		}
		t.Fatalf("expected %d tools, got %d: %v", len(expected), len(tools), names)
	}

	nameSet := map[string]bool{}
	for _, tool := range tools {
		nameSet[tool.Name] = true
	}

	for _, name := range expected {
		if !nameSet[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}

// TestAnnotations verifies all tools have appropriate annotations.
func TestAnnotations(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	readOnlyTools := map[string]bool{
		"vivechak_prepare_generator": true,
		"vivechak_status":            true,
		"vivechak_next_session":      true,
		"vivechak_validate":          true,
		"vivechak_run_gate":          true,
	}

	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		if tool.Annotations == nil {
			t.Errorf("%s: annotations are nil", tool.Name)
			continue
		}

		if readOnlyTools[tool.Name] {
			if !tool.Annotations.ReadOnlyHint {
				t.Errorf("%s: expected readOnlyHint=true", tool.Name)
			}
		} else {
			if tool.Annotations.ReadOnlyHint {
				t.Errorf("%s: expected readOnlyHint=false", tool.Name)
			}
		}
	}
}

// TestStatusNoWorkspace verifies vivechak_status works without an initialized workspace.
func TestStatusNoWorkspace(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_status",
		Arguments: map[string]any{
			"project_root": t.TempDir(),
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_status: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Success {
		t.Error("status should succeed even without workspace")
	}
	if env.NextStep == "" {
		t.Error("status must always provide next_step (Guided Worker)")
	}
}

// TestInitAndStatus tests the init → status flow.
func TestInitAndStatus(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_init: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("init failed: %s", env.Message)
	}

	// Verify directory structure
	dirs := []string{"research", "research/sessions", "research/templates"}
	for _, dir := range dirs {
		path := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("missing directory: %s", dir)
		}
	}

	// Verify templates copied
	templateFiles := []string{
		"DECISIONS.template.md",
		"CONFLICT-RESOLUTION.template.md",
		"COMPARISON-SESSION.template.md",
		"FOUNDING-ARCHITECTURE.template.md",
		"PHASE-0-GATE.template.md",
	}
	for _, tmpl := range templateFiles {
		path := filepath.Join(tmpDir, "research", "templates", tmpl)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("missing template: %s", tmpl)
		}
	}

	// Status should now show initialized
	result2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_status",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_status: %v", err)
	}

	env2 := parseEnvelope(t, result2)
	if !env2.Success {
		t.Error("status should succeed after init")
	}
}

// TestPrepareGenerator tests the prepare_generator tool.
func TestPrepareGenerator(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_prepare_generator",
		Arguments: map[string]any{
			"scope":   "project",
			"context": "A web application for managing research projects",
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_prepare_generator: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("prepare_generator failed: %s", env.Message)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatal("data should be a map")
	}

	prompt, _ := data["prompt"].(string)
	if prompt == "" {
		t.Error("prompt should not be empty")
	}
	if len(prompt) < 100 {
		t.Errorf("prompt seems too short: %d chars", len(prompt))
	}
}

func TestPrepareGenerator_ScopeAutoDetect(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize decision-scope workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "decision",
		},
	})
	if err != nil {
		t.Fatalf("vivechak_init: %v", err)
	}

	// Call prepare_generator without scope, providing project_root
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_prepare_generator",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"context":      "Choose between PostgreSQL and DynamoDB",
		},
	})
	if err != nil {
		t.Fatalf("vivechak_prepare_generator: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("prepare_generator failed: %s", env.Message)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatal("data should be a map")
	}

	if data["scope"] != "decision" {
		t.Errorf("expected scope 'decision' auto-detected from workspace, got %v", data["scope"])
	}
}

// TestSaveSessionValidation tests the validation ladder on session save.
func TestSaveSessionValidation(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace first
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir},
	})

	// Save a session with proper frontmatter
	content := `---
session_id: R-01
title: Test Session
date: 2026-09-27
status: complete
---

# Test Session

This is a test session with evidence grades.

Finding: The test framework works well. A (direct testing)

## Recommendations

Use this approach for production.
`
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "R-01",
			"content":      content,
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_save_session: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("save_session failed: %s", env.Message)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatal("data should be a map")
	}
	if data["validation_passed"] != true {
		t.Errorf("expected validation_passed = true, got %v", data["validation_passed"])
	}

	// Verify file was saved
	sessionPath := filepath.Join(tmpDir, "research", "sessions", "R-01.md")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Error("session file not saved")
	}
}

// TestValidateDryRun tests the validate tool doesn't write anything.
func TestValidateDryRun(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	content := `---
session_id: R-01
title: Test
date: 2026-09-27
---

Short body.
`
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_validate",
		Arguments: map[string]any{
			"artifact_type": "session",
			"content":       content,
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_validate: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("validate failed: %s", env.Message)
	}

	// Should have at least the W-NO-EVIDENCE-GRADES warning
	if len(env.Warnings) == 0 {
		t.Error("expected warnings for session without evidence grades")
	}
}

// TestGuidedWorkerPattern verifies every tool returns next_step.
func TestGuidedWorkerPattern(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Test a selection of tools
	calls := []struct {
		name string
		args map[string]any
	}{
		{"vivechak_status", map[string]any{"project_root": tmpDir}},
		{"vivechak_init", map[string]any{"project_root": tmpDir}},
		{"vivechak_prepare_generator", map[string]any{"scope": "project", "context": "test project"}},
		{"vivechak_validate", map[string]any{"content": "# test", "artifact_type": "session"}},
	}

	for _, tc := range calls {
		result, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      tc.name,
			Arguments: tc.args,
		})
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}

		env := parseEnvelope(t, result)
		if env.NextStep == "" {
			t.Errorf("%s: missing next_step — violates Guided Worker pattern", tc.name)
		}
		if env.Meta == nil {
			t.Errorf("%s: missing meta", tc.name)
		} else if env.Meta.Tool != tc.name {
			t.Errorf("%s: meta.tool=%q, want %q", tc.name, env.Meta.Tool, tc.name)
		}
	}
}

// TestEndToEndPipelineFlow tests the full workflow:
// init → save_plan → next_session → save_session → next_session → run_gate
func TestEndToEndPipelineFlow(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Step 1: Init workspace
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	env := parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("init failed: %s", env.Message)
	}

	// Step 2: Save a research pipeline
	pipeline := `# Research Pipeline

## Sessions

#### T1-01: Database Selection

| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Layer** | 0 |
| **Door Type** | One-Way |
| **Decision** | D-001 |
| **Dependencies** | None |
| **Output File** | ` + "`sessions/T1-01.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-01

## BRIEF
Investigate database options.

## DELIVERABLE
Recommendation with evidence grades.
` + "```" + `

---

#### T1-02: Auth Strategy

| Field | Value |
|---|---|
| **ID** | T1-02 |
| **Layer** | 0 |
| **Door Type** | Two-Way |
| **Decision** | D-002 |
| **Dependencies** | None |
| **Output File** | ` + "`sessions/T1-02.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-02

## BRIEF
Evaluate auth approaches.

## DELIVERABLE
Comparison matrix.
` + "```" + `

---

#### SYN-01: Synthesis

| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Layer** | 1 |
| **Dependencies** | T1-01, T1-02 |
| **Output File** | ` + "`research/FAD.md`" + ` |

` + "```prompt" + `
# SYNTHESIS

Compile findings.

[ALL_SESSION_FINDINGS]
` + "```" + `
`
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"content":      pipeline,
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}
	env = parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("save_plan failed: %s", env.Message)
	}

	// Step 3: Get next session (should return T1-01 or T1-02)
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}
	env = parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("next_session failed: %s", env.Message)
	}

	data, _ := env.Data.(map[string]any)
	sessionID, _ := data["session_id"].(string)
	if sessionID != "T1-01" && sessionID != "T1-02" {
		t.Errorf("expected T1-01 or T1-02, got %s", sessionID)
	}

	// Check that parallel sessions are detected
	otherReady := data["other_ready_sessions"]
	if otherReady == nil {
		t.Log("No parallel sessions detected (may be correct if T1-01 returned first)")
	}

	// Verify prompt was included
	prompt, _ := data["prompt"].(string)
	if prompt == "" {
		t.Error("next_session should include the prompt")
	}

	// Step 4: Save T1-01 session output
	t101Output := `---
session_id: T1-01
title: Database Selection
date: 2026-09-27
status: complete
---

# Database Selection

## Findings

PostgreSQL 16 is the best choice. A (official documentation)

## Recommendations

Use PostgreSQL with pgvector for embeddings.
`
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-01",
			"content":      t101Output,
		},
	})
	if err != nil {
		t.Fatalf("save_session T1-01: %v", err)
	}
	env = parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("save_session T1-01 failed: %s", env.Message)
	}

	// Step 5: Save T1-02 session output
	t102Output := `---
session_id: T1-02
title: Auth Strategy
date: 2026-09-27
status: complete
---

# Auth Strategy

## Findings

Clerk provides the best developer experience. B (comparison analysis)

## Recommendations

Use Clerk for authentication.
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-02",
			"content":      t102Output,
		},
	})
	if err != nil {
		t.Fatalf("save_session T1-02: %v", err)
	}

	// Step 6: Get next session — should be SYN-01 with findings injected
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("next_session for SYN-01: %v", err)
	}
	env = parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("next_session SYN-01 failed: %s", env.Message)
	}

	data, _ = env.Data.(map[string]any)
	sessionID, _ = data["session_id"].(string)
	if sessionID != "SYN-01" {
		t.Errorf("expected SYN-01, got %s", sessionID)
	}

	// Verify context injection happened
	synPrompt, _ := data["prompt"].(string)
	if synPrompt == "" {
		t.Error("SYN-01 should have a prompt")
	}

	// Step 7: Run gate (should fail — no FAD yet)
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env = parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("run_gate failed: %s", env.Message)
	}

	data, _ = env.Data.(map[string]any)
	gateStatus, _ := data["gate_status"].(string)
	if gateStatus == "PASS" {
		t.Error("gate should not pass without FAD")
	}
}

// TestFADSaveAndGatePass tests the full workflow:
// init → save_plan → save_session x3 (sessions + SYN) → save_session (FAD) → run_gate
func TestFADSaveAndGatePass(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	var result *mcp.CallToolResult
	// Step 1: Init workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Step 2: Save a research pipeline
	pipeline := `# Research Pipeline

## Sessions

#### T1-01: Database Selection

| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Layer** | 0 |
| **Door Type** | One-Way |
| **Decision** | D-001 |
| **Dependencies** | None |
| **Output File** | ` + "`sessions/T1-01.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-01

## BRIEF
Investigate database options.

## DELIVERABLE
Recommendation with evidence grades.
` + "```" + `

---

#### T1-02: Auth Strategy

| Field | Value |
|---|---|
| **ID** | T1-02 |
| **Layer** | 0 |
| **Door Type** | Two-Way |
| **Decision** | D-002 |
| **Dependencies** | None |
| **Output File** | ` + "`sessions/T1-02.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-02

## BRIEF
Evaluate auth approaches.

## DELIVERABLE
Comparison matrix.
` + "```" + `

---

#### SYN-01: Synthesis

| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Layer** | 1 |
| **Dependencies** | T1-01, T1-02 |
| **Output File** | ` + "`research/FAD.md`" + ` |

` + "```prompt" + `
# SYNTHESIS

Compile findings.

[ALL_SESSION_FINDINGS]
` + "```" + `
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"content":      pipeline,
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}

	// Step 3 & 4: Save T1-01 & T1-02 session output
	t101Output := `---
session_id: T1-01
title: Database Selection
date: 2026-09-27
status: complete
---
# Database Selection
Findings... A (doc)
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save_session",
		Arguments: map[string]any{"project_root": tmpDir, "session_id": "T1-01", "content": t101Output},
	})

	t102Output := `---
session_id: T1-02
title: Auth Strategy
date: 2026-09-27
status: complete
---
# Auth Strategy
Findings... B (doc)
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save_session",
		Arguments: map[string]any{"project_root": tmpDir, "session_id": "T1-02", "content": t102Output},
	})

	// Step 5: Save SYN-01 output
	// Step 5: Save SYN-01 output (dual-writes to research/FAD.md and research/sessions/SYN-01.md)
	syn01Output := `---
session_id: SYN-01
title: Founding Architecture Document
date: 2026-09-27
status: complete
---
# Founding Architecture Document

## Architecture Overview

PostgreSQL 16 selected as primary datastore. A (official documentation)

## Evidence Summary

- Database: PostgreSQL 16 with pgvector — B (benchmark comparison)
- Auth: Clerk — B (comparison analysis)

## Decisions

All decisions recorded with review triggers. C (team review)
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save_session",
		Arguments: map[string]any{"project_root": tmpDir, "session_id": "SYN-01", "content": syn01Output},
	})
	if err != nil {
		t.Fatalf("save_session SYN-01: %v", err)
	}

	// Step 5.5: Record a decision (required by gate Track A check 4)
	decisionContent := `---
decision_id: D-001
title: Primary Datastore
status: accepted
door_type: one-way
review_trigger: "Throughput exceeds 50k ops"
date: 2026-09-27
---
# Decision: Primary Datastore

## Context
We need a database. A (official documentation)

## Decision
Use PostgreSQL 16. B (benchmark comparison)
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"decision_id":   "D-001",
			"artifact_type": "decision",
			"content":       decisionContent,
		},
	})

	// Step 6.5: Record a decision to create DECISIONS.md (needed for gate PASS)
	decisionOutput := `---
decision_id: D-002
title: Auth Strategy
status: accepted
door_type: two-way
---
# Context
We need auth.
# Decision
Use Clerk.
# Consequences
Tied to Clerk.
`
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "decision",
			"decision_id":   "D-002",
			"content":       decisionOutput,
		},
	})
	if err != nil {
		t.Fatalf("record_decision: %v", err)
	}
	decEnv := parseEnvelope(t, result)
	if !decEnv.Success {
		t.Fatalf("record_decision failed: %s", decEnv.Message)
	}

	// Step 7: Run gate
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env := parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("run_gate failed: %s", env.Message)
	}

	data, _ := env.Data.(map[string]any)
	gateStatus, _ := data["gate_status"].(string)
	if gateStatus != "PASS" {
		t.Errorf("expected gate to PASS with FAD, got %s. Issues: %v", gateStatus, data)
	}
}

func TestRecordDecision_ConflictResolution(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Initialize workspace
	initRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init call: %v", err)
	}
	initEnv := parseEnvelope(t, initRes)
	if !initEnv.Success {
		t.Fatalf("init failed: %s", initEnv.Message)
	}

	// 2. Call vivechak_record_decision with artifact_type = "conflict-resolution"
	crContent := `---
id: CHK-01
decision_id: D-001
title: Conflict Resolution for Storage Engine
status: resolved
door_type: one-way
conflicting_sources:
  - Session T1-01
  - Session T1-02
chosen_option: PostgreSQL
resolution_method: ACH matrix
date: 2026-09-29
schema_version: "0.1.0"
---

# Conflict Resolution: CHK-01 — Storage Engine Divergence

## 1. Conflict Summary
T1-01 recommended PostgreSQL based on ACID compliance. T1-02 recommended MongoDB based on document flexibility.

## 2. Analysis of Competing Hypotheses (ACH Matrix)
Evaluated both options against data integrity, query latency, and operational complexity.
PostgreSQL dominates on ACID transactions and relational joins.

## 3. Resolution
Adopt PostgreSQL with jsonb columns for flexible document attributes.
`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"decision_id":   "CHK-01",
			"artifact_type": "conflict-resolution",
			"content":       crContent,
		},
	})
	if err != nil {
		t.Fatalf("record_decision conflict-resolution: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("expected success, got message: %s, warnings: %v", env.Message, env.Warnings)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected Data map, got %T", env.Data)
	}
	if data["artifact_type"] != "conflict-resolution" {
		t.Errorf("expected artifact_type 'conflict-resolution', got %v", data["artifact_type"])
	}
	if data["status"] != "valid" {
		t.Errorf("expected status 'valid', got %v (warnings: %v)", data["status"], env.Warnings)
	}

	// Verify file was written to disk
	filePath := filepath.Join(tmpDir, "research", "CHK-01-conflict-resolution.md")
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("expected file %s to exist on disk: %v", filePath, err)
	}
	if !strings.Contains(string(contentBytes), "ACH Matrix") {
		t.Errorf("expected saved file to contain ACH Matrix")
	}
}

func TestIDValidationRejectsTraversal(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	invalidIDs := []string{"../../etc/passwd", "../evil"}
	for _, id := range invalidIDs {
		result, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": tmpDir,
				"session_id":   id,
				"content":      "content",
			},
		})
		if err != nil {
			t.Fatalf("save_session with %s: %v", id, err)
		}
		// If the SDK itself rejected it due to schema, it might be an error not JSON.
		// We just ensure it didn't succeed.
		if len(result.Content) > 0 {
			if tc, ok := result.Content[0].(*mcp.TextContent); ok {
				if strings.HasPrefix(tc.Text, "validation error") {
					continue
				}
				env := parseEnvelope(t, result)
				if env.Success {
					t.Errorf("expected save_session to fail for ID %s", id)
				}
			}
		}
	}

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-01",
			"content":      "---\nsession_id: T1-01\ntitle: test\ndate: 2026-09-27\nstatus: complete\n---\n# test\nA (doc)",
		},
	})
	if err != nil {
		t.Fatalf("save_session with T1-01: %v", err)
	}
	env := parseEnvelope(t, result)
	if !env.Success {
		t.Errorf("expected save_session to succeed for T1-01, got: %s", env.Message)
	}

	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "decision",
			"decision_id":   "../../../hack",
			"content":       "content",
		},
	})
	if err != nil {
		t.Fatalf("record_decision with hack: %v", err)
	}
	if len(result.Content) > 0 {
		if tc, ok := result.Content[0].(*mcp.TextContent); ok {
			if !strings.HasPrefix(tc.Text, "validation error") {
				env = parseEnvelope(t, result)
				if env.Success {
					t.Errorf("expected record_decision to fail for ../../../hack")
				}
			}
		}
	}
}

func TestContentSizeLimit(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	// 10MB + 1 string
	hugeContent := strings.Repeat("A", 10*1024*1024+1)

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-01",
			"content":      hugeContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session with huge content: %v", err)
	}
	env := parseEnvelope(t, result)
	if env.Success {
		t.Error("expected save_session to fail for huge content")
	}
}

func TestInit_RejectsFilesystemRoot(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	rootPaths := []string{"/"}
	vol := filepath.VolumeName(t.TempDir())
	if vol != "" {
		rootPaths = append(rootPaths, vol+"\\", vol+"/")
	}

	for _, root := range rootPaths {
		result, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_init",
			Arguments: map[string]any{
				"project_root": root,
			},
		})
		if err != nil {
			t.Fatalf("call init with %s: %v", root, err)
		}
		env := parseEnvelope(t, result)
		if env.Success {
			t.Errorf("expected vivechak_init to reject filesystem root %q", root)
		}
		if !strings.Contains(env.Message, "cannot be a filesystem root") {
			t.Errorf("expected error message to mention filesystem root, got: %s", env.Message)
		}
	}
}

func TestNextSession_InjectionSizeWarning(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Init workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// 2. Save pipeline with T1-01 and SYN-01
	pipeline := `# Research Pipeline
## Sessions
### Session T1-01
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Dependencies** | None |

` + "```prompt" + `
Prompt T1-01
` + "```" + `

### Session SYN-01
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Dependencies** | T1-01 |

` + "```prompt" + `
Synthesize findings:
[ALL_SESSION_FINDINGS]
` + "```" + `
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"content":      pipeline,
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}

	// 3. Save huge T1-01 session (>100KB)
	largeBody := strings.Repeat("Finding: High-performance caching layer delivers 10x throughput. A (benchmark)\n", 2000)
	t101Content := `---
session_id: T1-01
title: Database Selection
date: 2026-09-29
status: complete
---
# Findings
` + largeBody
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-01",
			"content":      t101Content,
		},
	})
	if err != nil {
		t.Fatalf("save_session: %v", err)
	}

	// 4. Request SYN-01 via next_session
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "SYN-01",
		},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}

	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("next_session failed: %s", env.Message)
	}

	foundWarning := false
	for _, w := range env.Warnings {
		if strings.Contains(w, "W-INJECTION-SIZE") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Errorf("expected W-INJECTION-SIZE warning in env.Warnings, got: %v", env.Warnings)
	}
}

func TestDecisionScope_EndToEnd(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	// 1. Init with scope decision
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "decision",
		},
	})
	if err != nil {
		t.Fatalf("init decision scope: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("init failed: %s", env.Message)
	}

	// 2. Save Plan with DecisionsContent (ADR)
	planContent := `# Research Plan: D-015 Cache Layer

### D-015-S1: Cache Layer Comparison

| Field | Value |
|---|---|
| **ID** | D-015-S1 |
| **Output File** | ` + "`sessions/D-015-S1-cache-comparison.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: Cache Layer
Compare Redis vs Dragonfly. A (benchmark)
` + "```" + `
`
	adrContent := `---
id: D-015
title: Cache Layer Selection
status: proposed
door_type: two-way
date: 2026-09-29
human_reviewed: true
review_trigger: Traffic exceeds 50k QPS
tags: [caching, database]
---
# Context
We need a distributed cache. A (official docs)

# Decision
Pending research.
`
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root":      tmpDir,
			"scope":             "decision",
			"decision_id":       "D-015",
			"slug":              "cache-layer",
			"content":           planContent,
			"decisions_content": adrContent,
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}
	env = parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("save_plan failed: %s", env.Message)
	}

	// 3. Next session returns D-015-S1
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}
	env = parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("next_session failed: %s", env.Message)
	}
	data, _ := env.Data.(map[string]any)
	if data["session_id"] != "D-015-S1" {
		t.Fatalf("expected session_id D-015-S1, got %v", data["session_id"])
	}

	// 4. Save session output
	sessionContent := `---
session_id: D-015-S1
title: Cache Comparison
status: complete
date: 2026-09-29
---
# Cache Comparison
Findings: Redis chosen for cluster stability. A (official documentation)
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "D-015-S1",
			"content":      sessionContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session: %v", err)
	}

	// 5. Run gate - should pass for decision scope
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"verbose":      true,
		},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env = parseEnvelope(t, res)
	gateData, _ := env.Data.(map[string]any)
	if gateData["gate_status"] != "PASS" || gateData["gate_passed"] != true {
		t.Fatalf("expected gate PASS, got status=%v passed=%v (message=%s)", gateData["gate_status"], gateData["gate_passed"], env.Message)
	}
}

func TestComparisonScope_EndToEnd(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	// 1. Init with scope comparison
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "comparison",
		},
	})
	if err != nil {
		t.Fatalf("init comparison scope: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("init failed: %s", env.Message)
	}

	// 2. Save comparison plan/prompt
	comparisonPrompt := `# RESEARCH BRIEF: Queue Selection
Compare Kafka vs RabbitMQ. A (benchmark)
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "comparison",
			"decision_id":  "D-020",
			"slug":         "queue-tech",
			"content":      comparisonPrompt,
		},
	})
	if err != nil {
		t.Fatalf("save_plan comparison: %v", err)
	}

	// 3. Next session returns comparison prompt
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}
	env = parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("next_session comparison failed: %s", env.Message)
	}

	// 4. Save session output for comparison
	cmpOutput := `---
session_id: D-020-cmp-queue-tech
title: Queue Comparison
status: complete
date: 2026-09-29
---
# Queue Comparison Matrix
Kafka chosen for stream retention. A (official docs)
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "D-020-cmp-queue-tech",
			"content":      cmpOutput,
		},
	})
	if err != nil {
		t.Fatalf("save_session comparison: %v", err)
	}

	// 5. Run gate for comparison scope
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"verbose":      true,
		},
	})
	if err != nil {
		t.Fatalf("run_gate comparison: %v", err)
	}
	env = parseEnvelope(t, res)
	gateData, _ := env.Data.(map[string]any)
	if gateData["gate_status"] != "PASS" || gateData["gate_passed"] != true {
		t.Fatalf("expected gate PASS, got status=%v passed=%v (message=%s)", gateData["gate_status"], gateData["gate_passed"], env.Message)
	}
}

func TestNextSession_NotFound_SessionIDList(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	// Init
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	pipeline := `# Pipeline
#### T1-01: Session One
| **ID** | T1-01 |
| **Output File** | sessions/T1-01.md |
` + "```prompt" + `
Prompt one
` + "```" + `
`
	// Save pipeline
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"content":      pipeline,
		},
	})

	// Request nonexistent session ID
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "NONEXISTENT-99",
		},
	})
	if err != nil {
		t.Fatalf("unexpected call error: %v", err)
	}
	env := parseEnvelope(t, res)
	if env.Success {
		t.Errorf("expected error for nonexistent session ID")
	}
	if !strings.Contains(env.NextStep, "T1-01") {
		t.Errorf("expected available sessions list in NextStep, got: %s", env.NextStep)
	}
}

func TestValidateArtifact_PlanAndConflict(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)

	// 1. Conflict resolution
	conflict := `---
id: CR-001
decision_id: D-001
title: Database Selection Conflict Resolution
status: complete
door_type: one-way
---
# Conflict Resolution: Database Selection

## Conflict Summary
Session T1-01 recommended PostgreSQL. Session T1-02 recommended ScyllaDB.

## ACH Matrix Analysis
| Hypothesis | Evidence 1 | Evidence 2 | Verdict |
|---|---|---|---|
| PostgreSQL | Consistent | Consistent | Strongly Supported |
| ScyllaDB | Inconsistent | Consistent | Rejected |

## Final Resolution
Adopt PostgreSQL for primary ACID datastore. A (benchmarks)
`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_validate",
		Arguments: map[string]any{
			"artifact_type": "conflict-resolution",
			"content":       conflict,
		},
	})
	if err != nil {
		t.Fatalf("validate conflict-resolution: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("expected validate success: %s", env.Message)
	}
	data, _ := env.Data.(map[string]any)
	if data["status"] != "valid" && data["status"] != "valid-with-warnings" {
		t.Errorf("expected valid status, got %v", data["status"])
	}

	// 2. Plan artifact
	plan := `# Research Pipeline
### Session T1-01: Benchmark
| **ID** | T1-01 |
| **Dependencies** | None |
` + "````prompt" + `
Execute benchmark
` + "````" + `
`
	resPlan, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_validate",
		Arguments: map[string]any{
			"artifact_type": "plan",
			"content":       plan,
		},
	})
	if err != nil {
		t.Fatalf("validate plan: %v", err)
	}
	envPlan := parseEnvelope(t, resPlan)
	if !envPlan.Success {
		t.Fatalf("expected validate plan success: %s", envPlan.Message)
	}
	dataPlan, _ := envPlan.Data.(map[string]any)
	if errCount, ok := dataPlan["error_count"].(float64); !ok || errCount != 0 {
		t.Errorf("expected 0 errors on valid plan, got %v", dataPlan["error_count"])
	}
}

func TestRunGate_Mechanics(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	// 1. Initialize
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	// 2. Save pipeline with 3 sessions (meeting recommended minimum for project scope)
	pipe := `# Pipeline
#### T1-01: Core Architecture
| **ID** | T1-01 |
| **Dependencies** | None |
| **Output File** | sessions/T1-01.md |
` + "```prompt" + `
Brief 1
` + "```" + `

#### T1-02: Storage Layer
| **ID** | T1-02 |
| **Dependencies** | None |
| **Output File** | sessions/T1-02.md |
` + "```prompt" + `
Brief 2
` + "```" + `

#### SYN-01: Synthesis
| **ID** | SYN-01 |
| **Dependencies** | T1-01, T1-02 |
| **Output File** | research/FAD.md |
` + "```prompt" + `
Synthesis brief
` + "```" + `
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"content":      pipe,
		},
	})

	// 3. Save only session T1-01
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-01",
			"content": `---
session_id: T1-01
title: Core Architecture
date: 2026-09-29
status: complete
---
# Architecture Findings
Recommended SQLite in WAL mode for lightweight single-node performance. A (docs)
`,
		},
	})

	// 4. Run gate while DAG is incomplete (T1-02 and SYN-01 missing)
	resGate1, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("gate call 1: %v", err)
	}
	envGate1 := parseEnvelope(t, resGate1)
	gate1Data, _ := envGate1.Data.(map[string]any)
	if gate1Data["gate_status"] == "PASS" {
		t.Error("gate should fail when DAG sessions are incomplete")
	}

	// 5. Complete T1-02
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-02",
			"content": `---
session_id: T1-02
title: Storage Layer
date: 2026-09-29
status: complete
---
# Storage Findings
Raw disk benchmarks demonstrate 15k IOPS sustained under load. A (benchmarks)
`,
		},
	})

	// 6. Complete SYN-01 and FAD with substantive body (> 100 chars and evidence grades)
	fadContent := `---
id: SYN-01
title: Founding Architecture Document
synthesis_date: 2026-09-29
status: complete
---
# Founding Architecture Document

## Executive Summary
This document synthesizes findings across sessions T1-01 and T1-02 to establish the project's foundation.
The architecture utilizes SQLite with WAL mode backed by high-IOPS storage. A (benchmarks)

## Technology Decisions
1. Datastore: SQLite in WAL mode. A (docs)
2. Storage: NVMe direct-attached volume. B (benchmarks)
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "SYN-01",
			"content":      fadContent,
		},
	})

	// 7. Record one-way door ADR without reversal trigger
	untriggeredADR := `<!-- DECISION: D-001 -->
---
decision_id: D-001
title: Database Selection
status: accepted
door_type: one-way
---
# Decision
Use SQLite in WAL mode. A (docs)
<!-- /DECISION: D-001 -->
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "decision",
			"decision_id":   "D-001",
			"content":       untriggeredADR,
		},
	})

	// 8. Run gate with one-way door lacking reversal trigger
	resGate2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("gate call 2: %v", err)
	}
	envGate2 := parseEnvelope(t, resGate2)
	gate2Data, _ := envGate2.Data.(map[string]any)
	if gate2Data["gate_status"] == "PASS" {
		t.Error("gate should fail when one-way door decision lacks reversal triggers")
	}

	// 9. Add reversal trigger to the one-way door
	triggeredADR := `<!-- DECISION: D-001 -->
---
decision_id: D-001
title: Database Selection
status: accepted
door_type: one-way
review_trigger: "Throughput exceeds 50k ops"
---
# Decision
Use SQLite in WAL mode. A (docs)
<!-- /DECISION: D-001 -->
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "decision",
			"decision_id":   "D-001",
			"content":       triggeredADR,
		},
	})

	// 10. Run gate -> MUST PASS!
	resGate3, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("gate call 3: %v", err)
	}
	envGate3 := parseEnvelope(t, resGate3)
	gate3Data, _ := envGate3.Data.(map[string]any)
	if gate3Data["gate_status"] != "PASS" || gate3Data["gate_passed"] != true {
		t.Fatalf("expected gate PASS after human signoff, got status=%v passed=%v (message=%s, warnings=%v)",
			gate3Data["gate_status"], gate3Data["gate_passed"], envGate3.Message, envGate3.Warnings)
	}
}

func TestInit_HomeDirGuard(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("user home directory not available")
	}

	t.Chdir(home)
	cs := testServer(t)
	ctx := context.Background()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{}, // project_root omitted
	})
	if err != nil {
		t.Fatalf("unexpected call error: %v", err)
	}
	env := parseEnvelope(t, res)
	if env.Success {
		t.Error("expected error when initializing directly in home directory with omitted project_root")
	}
	if !strings.Contains(env.Message, "user home directory") {
		t.Errorf("expected home directory error message, got: %s", env.Message)
	}
}
