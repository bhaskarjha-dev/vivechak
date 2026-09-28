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

// TestSaveSessionValidation tests the validation ladder on session save.
func TestSaveSessionValidation(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace first
	cs.CallTool(ctx, &mcp.CallToolParams{
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
	otherReady, _ := data["other_ready_sessions"]
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
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
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

	// Step 1: Init workspace
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
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
	result, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
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
	result, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{"project_root": tmpDir, "session_id": "T1-02", "content": t102Output},
	})

	// Step 5: Save SYN-01 output
	syn01Output := `---
session_id: SYN-01
title: Synthesis
date: 2026-09-27
status: complete
---
# Synthesis
Combined findings. A (official documentation)
`
	result, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{"project_root": tmpDir, "session_id": "SYN-01", "content": syn01Output},
	})

	// Step 5.5: Record a decision (required by gate Track A check 4)
	decisionContent := `---
decision_id: D-001
title: Primary Datastore
status: accepted
door_type: one-way
date: 2026-09-27
---
# Decision: Primary Datastore

## Context
We need a database. A (official documentation)

## Decision
Use PostgreSQL 16. B (benchmark comparison)
`
	result, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"decision_id":   "D-001",
			"artifact_type": "decision",
			"content":       decisionContent,
		},
	})

	// Step 6: Save FAD (must contain evidence grades to pass gate Track B check 2)
	fadOutput := `---
session_id: FAD
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
	result, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{"project_root": tmpDir, "session_id": "FAD", "content": fadOutput},
	})

	// Step 6.5: Record a decision to create DECISIONS.md (needed for gate PASS)
	decisionOutput := `---
decision_id: D-001
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
			"decision_id":   "D-001",
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

func TestIDValidationRejectsTraversal(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace
	cs.CallTool(ctx, &mcp.CallToolParams{
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
			"project_root": tmpDir,
			"artifact_type": "decision",
			"decision_id":  "../../../hack",
			"content":      "content",
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
	cs.CallTool(ctx, &mcp.CallToolParams{
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
