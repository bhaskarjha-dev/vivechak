package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMassive_LiveIDEInheritance_RejectionAndGuidance compiles the real Vivechak binary,
// spawns it with CWD set to a simulated IDE program folder (mimicking Antigravity / Cursor / VS Code),
// and verifies that:
// 1. Un-targeted vivechak_init calls are safely blocked and return guided worker errors.
// 2. No artifacts or templates are written to the IDE installation directory.
// 3. Explicitly targeted vivechak_init calls successfully create and orchestrate workspaces.
func TestMassive_LiveIDEInheritance_RejectionAndGuidance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)

	// 1. Build live binary
	binDir := t.TempDir()
	binName := "vck_massive_test"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	// 2. Create simulated IDE program folder
	mockIDEDir := filepath.Join(t.TempDir(), "AppData", "Local", "Programs", "Antigravity IDE")
	if err := os.MkdirAll(mockIDEDir, 0o755); err != nil {
		t.Fatalf("creating mock IDE folder: %v", err)
	}

	// 3. Spawn live MCP server with CWD = mockIDEDir (simulating host CWD inheritance)
	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	serverCmd.Dir = mockIDEDir
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "massive-test-client", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connecting to live MCP server: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	// 4. Case A: vivechak_init with empty arguments MUST fail
	resEmpty, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("callTool vivechak_init (empty) error: %v", err)
	}
	envEmpty := parseTestEnvelope(t, resEmpty)
	if envEmpty.Success {
		t.Fatal("expected vivechak_init with empty args in IDE CWD to fail, got success")
	}
	if !strings.Contains(envEmpty.Message, "application or system directory") {
		t.Errorf("expected error to mention application or system directory, got: %s", envEmpty.Message)
	}
	if !strings.Contains(envEmpty.NextStep, "project_root") {
		t.Errorf("expected NextStep to instruct passing project_root, got: %s", envEmpty.NextStep)
	}

	// Verify mockIDEDir is completely untouched
	if _, err := os.Stat(filepath.Join(mockIDEDir, "research")); !os.IsNotExist(err) {
		t.Fatal("research/ directory was mistakenly created inside mock IDE program folder!")
	}

	// 5. Case B: vivechak_init with project_root="." MUST also fail
	resDot, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": "."},
	})
	if err != nil {
		t.Fatalf("callTool vivechak_init (.) error: %v", err)
	}
	envDot := parseTestEnvelope(t, resDot)
	if envDot.Success {
		t.Fatal("expected vivechak_init with '.' in IDE CWD to fail, got success")
	}
	if _, err := os.Stat(filepath.Join(mockIDEDir, "research")); !os.IsNotExist(err) {
		t.Fatal("research/ directory was mistakenly created inside mock IDE program folder on '.'!")
	}

	// 6. Case C: vivechak_init with EXPLICIT project workspace MUST succeed
	realProjectDir := filepath.Join(t.TempDir(), "lab", "personal-finance-dashboard")
	resExplicit, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": realProjectDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("callTool vivechak_init (explicit) error: %v", err)
	}
	envExplicit := parseTestEnvelope(t, resExplicit)
	if !envExplicit.Success {
		t.Fatalf("expected explicit vivechak_init to succeed, got: %s", envExplicit.Message)
	}

	// Verify workspace structure in realProjectDir
	if !core.WorkspaceExists(realProjectDir) {
		t.Fatalf("expected workspace to exist at %s", realProjectDir)
	}
	templatesDir := filepath.Join(realProjectDir, core.TemplatesDir)
	for _, tmpl := range core.TemplatesToCopy {
		tPath := filepath.Join(templatesDir, tmpl)
		if fi, err := os.Stat(tPath); err != nil || fi.Size() == 0 {
			t.Errorf("template %s not properly created: %v", tmpl, err)
		}
	}

	// Mock IDE folder MUST still have NO research/ directory
	if _, err := os.Stat(filepath.Join(mockIDEDir, "research")); !os.IsNotExist(err) {
		t.Fatal("research/ directory leaked into mock IDE directory during explicit init!")
	}

	// 7. Execute full MCP pipeline against realProjectDir
	// 7.1 status
	resStatus, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_status",
		Arguments: map[string]any{"project_root": realProjectDir},
	})
	if err != nil {
		t.Fatalf("status call: %v", err)
	}
	envStatus := parseTestEnvelope(t, resStatus)
	if !envStatus.Success {
		t.Fatalf("status failed: %s", envStatus.Message)
	}

	// 7.2 prepare_generator
	resPrep, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_prepare_generator",
		Arguments: map[string]any{
			"project_root": realProjectDir,
			"scope":        "project",
			"context":      "Personal Finance Dashboard with encrypted local database",
		},
	})
	if err != nil {
		t.Fatalf("prepare_generator call: %v", err)
	}
	envPrep := parseTestEnvelope(t, resPrep)
	if !envPrep.Success {
		t.Fatalf("prepare_generator failed: %s", envPrep.Message)
	}

	// 7.3 save_plan with 3 sessions (T0-01, T1-01, SYN-01)
	pipelineDAG := `# Research Pipeline: Personal Finance Dashboard
## Sessions
### Session T0-01: Aggregation Landscape
| Field | Value |
|---|---|
| **ID** | T0-01 |
| **Layer** | 0 |
| **Type** | Two-Way |
| **Dependencies** | None |
| **Output File** | research/sessions/T0-01.md |

#### Research Question
Landscape of financial aggregators.

### Session T1-01: Embedded Database
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Layer** | 1 |
| **Type** | One-Way |
| **Informs Decision** | D-001 |
| **Dependencies** | T0-01 |
| **Output File** | research/sessions/T1-01.md |

#### Research Question
Which embedded database satisfies privacy and encryption requirements?

### Session SYN-01: Terminal Architecture Synthesis
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Layer** | 2 |
| **Type** | One-Way |
| **Dependencies** | T1-01 |
| **Output File** | research/FAD.md |

#### Research Question
Synthesize all findings into Founding Architecture Document.
`
	resPlan, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": realProjectDir,
			"content":      pipelineDAG,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("save_plan call: %v", err)
	}
	envPlan := parseTestEnvelope(t, resPlan)
	if !envPlan.Success {
		t.Fatalf("save_plan failed: %s", envPlan.Message)
	}

	// 7.4 next_session (returns T0-01)
	resNext0, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": realProjectDir},
	})
	if err != nil {
		t.Fatalf("next_session call: %v", err)
	}
	envNext0 := parseTestEnvelope(t, resNext0)
	if !envNext0.Success || envNext0.Data["session_id"] != "T0-01" {
		t.Fatalf("expected next session T0-01, got: %v", envNext0.Data)
	}

	// 7.4b save_session (T0-01)
	sess0Content := `---
session_id: T0-01
title: "Aggregation Landscape"
status: complete
date: "2026-10-06"
---
# T0-01: Aggregation Landscape

## Prior
We assumed Plaid was required.

## Research Question
What is the aggregation landscape?

## Key Findings
- SimpleFIN Bridge provides BYOK financial sync. Grade A (https://simplefin.org)

## Recommendation
Select SimpleFIN.

## Alternatives Considered
- Plaid.

## Open Questions & Risks
- Bank auth latency.

## Discovered Concerns
- None.

## Delta
| Belief | Shift |
|---|---|
| Plaid default | Shifted to SimpleFIN |

## Evidence Ledger
| Finding | Grade | Source |
|---|---|---|
| BYOK sync | A | https://simplefin.org |
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": realProjectDir,
			"session_id":   "T0-01",
			"content":      sess0Content,
		},
	})

	// 7.4c next_session (should now return T1-01)
	resNext, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": realProjectDir},
	})
	if err != nil {
		t.Fatalf("next_session call: %v", err)
	}
	envNext := parseTestEnvelope(t, resNext)
	if !envNext.Success || envNext.Data["session_id"] != "T1-01" {
		t.Fatalf("expected next session T1-01, got: %v", envNext.Data)
	}

	// 7.5 save_session (T1-01)
	sessionContent := `---
session_id: T1-01
title: "Embedded Database Selection"
status: complete
date: "2026-10-06"
informs_decisions: ["D-001"]
---

# T1-01: Embedded Database Selection

## Prior
We assumed SQLite would be standard.

## Research Question
Which embedded database satisfies privacy and native encryption?

## Key Findings
- libsql supports AEAD AES-GCM native database-level encryption A (https://github.com/tursodatabase/libsql).
- Go driver requires explicit build tags B (benchmarks spike).

## Recommendation
Select libsql with native encryption.

## Alternatives Considered
- SQLite vanilla: lacks native encryption.
- SQLCipher: requires complex CGO compilation.

## Open Questions & Risks
- Driver build complexity on Windows ARM64.

## Discovered Concerns
- Turso licensing stability over multi-year horizon.

## Delta
| Belief | Shift | Rationale |
|---|---|---|
| SQLite is default | Shifted to libsql | Native AEAD encryption |

## Evidence Ledger
| Finding | Grade | Source |
|---|---|---|
| Native AEAD encryption | A | https://github.com/tursodatabase/libsql |
`
	resSaveSess, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": realProjectDir,
			"session_id":   "T1-01",
			"content":      sessionContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session call: %v", err)
	}
	envSaveSess := parseTestEnvelope(t, resSaveSess)
	if !envSaveSess.Success {
		t.Fatalf("save_session failed: %s", envSaveSess.Message)
	}

	// 7.6 record_decision (D-001)
	decisionContent := `---
id: D-001
title: "Select libsql as Primary Encrypted Datastore"
door_type: "one-way"
status: "accepted"
date: "2026-10-06"
confidence: "high"
human_reviewed: true
review_trigger: "Write latency exceeds 50ms"
review_date: "2026-12-01"
informing_sessions: ["T1-01"]
---

# ADR D-001: Select libsql as Primary Encrypted Datastore

## Context
Personal finance dashboard requires local AES-GCM encryption at rest.

## Decision
Adopt libsql with native AEAD encryption. Corroborated by Grade A docs. A (https://github.com/tursodatabase/libsql).

## Consequences
- Requires CGO build flags.
- High write throughput.

## Rejected Alternatives
- SQLite vanilla: lacks encryption.
- SQLCipher: compilation friction.

## Reversal Trigger
If Turso changes licensing or Go driver deprecates native encryption.
`
	resDec, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": realProjectDir,
			"decision_id":  "D-001",
			"content":      decisionContent,
		},
	})
	if err != nil {
		t.Fatalf("record_decision call: %v", err)
	}
	envDec := parseTestEnvelope(t, resDec)
	if !envDec.Success {
		t.Fatalf("record_decision failed: %s", envDec.Message)
	}

	// 7.7 save synthesis FAD session
	fadContent := `---
session_id: SYN-01
title: "Founding Architecture Document"
status: complete
date: "2026-10-06"
informs_decisions: ["D-001"]
---

# Founding Architecture Document: Personal Finance Dashboard

## Executive Summary
Privacy-first finance dashboard built on Wails v2 and encrypted libsql.

## Architecture
See ADR D-001 for datastore selection. Corroborated with Grade A evidence. A (https://github.com/tursodatabase/libsql).

## Prior
None.

## Research Question
Terminal architecture synthesis.

## Key Findings
- Single binary desktop app viable. A (specs)

## Recommendation
Proceed to implementation.

## Alternatives Considered
- Electron shell.

## Open Questions & Risks
- Pre-coding spike needed.

## Discovered Concerns
- None.

## Delta
| Item | Shift |
|---|---|
| Shell | Selected Wails |

## Evidence Ledger
| Finding | Grade | Source |
|---|---|---|
| Viability | A | specs |
`
	resFAD, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": realProjectDir,
			"session_id":   "SYN-01",
			"content":      fadContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session (FAD) call: %v", err)
	}
	envFAD := parseTestEnvelope(t, resFAD)
	if !envFAD.Success {
		t.Fatalf("save_session FAD failed: %s", envFAD.Message)
	}

	// 7.8 run Phase 0 exit gate
	resGate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": realProjectDir,
		},
	})
	if err != nil {
		t.Fatalf("run_gate call: %v", err)
	}
	envGate := parseTestEnvelope(t, resGate)
	if envGate.Data["gate_status"] != "PASS" || envGate.Data["gate_passed"] != true {
		t.Fatalf("expected gate PASS, got status=%v passed=%v warnings=%v message=%s",
			envGate.Data["gate_status"], envGate.Data["gate_passed"], envGate.Warnings, envGate.Message)
	}

	// 8. Physical disk assertion on realProjectDir
	assertFilesExist(t, realProjectDir, []string{
		"research/RESEARCH-PIPELINE.md",
		"research/DECISIONS.md",
		"research/D-001-decision.md",
		"research/sessions/T1-01.md",
		"research/FAD.md",
		"FOUNDING-ARCHITECTURE.md",
		"research/PHASE-0-GATE.md",
		"research/templates/SESSION.template.md",
	})
}

// TestMassive_SandboxesAndEnterprisePaths_LiveMCPMatrix tests that Vivechak functions
// flawlessly across all 6 major sandbox and enterprise deployment path archetypes.
func TestMassive_SandboxesAndEnterprisePaths_LiveMCPMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)

	binDir := t.TempDir()
	binName := "vck_sandbox_matrix"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "sandbox-matrix-client", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connecting to live MCP server: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	testCases := []struct {
		name     string
		relPath  string
		scope    string
	}{
		{"SWE-bench Docker sandbox", "testbed/my-repo", "project"},
		{"OpenHands container workspace", "workspace/agent-task", "decision"},
		{"GitHub Codespaces repo folder", "workspaces/org-cloud-repo", "comparison"},
		{"Enterprise Linux /opt deployment", "opt/company/analytics-service", "project"},
		{"Web server /var/www document root", "var/www/html/customer-portal", "decision"},
		{"MicroVM ephemeral tmp sandbox", "tmp/microvm-run-9481/sandbox", "comparison"},
	}

	baseDir := t.TempDir()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := filepath.Join(baseDir, tc.relPath)

			// 1. vivechak_init
			resInit, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_init",
				Arguments: map[string]any{
					"project_root": targetDir,
					"scope":        tc.scope,
				},
			})
			if err != nil {
				t.Fatalf("[%s] vivechak_init call error: %v", tc.name, err)
			}
			envInit := parseTestEnvelope(t, resInit)
			if !envInit.Success {
				t.Fatalf("[%s] vivechak_init failed: %s", tc.name, envInit.Message)
			}

			// Verify templates exist
			info := core.InspectWorkspace(targetDir)
			if !info.Initialized {
				t.Fatalf("[%s] workspace not initialized at %s", tc.name, targetDir)
			}
			expectedTemplates := len(core.TemplatesForScope(core.Scope(tc.scope)))
			if info.TemplateCount != expectedTemplates {
				t.Errorf("[%s] expected %d templates, got %d", tc.name, expectedTemplates, info.TemplateCount)
			}

			// 2. vivechak_status
			resStatus, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      "vivechak_status",
				Arguments: map[string]any{"project_root": targetDir},
			})
			if err != nil {
				t.Fatalf("[%s] status call error: %v", tc.name, err)
			}
			envStatus := parseTestEnvelope(t, resStatus)
			if !envStatus.Success {
				t.Fatalf("[%s] status failed: %s", tc.name, envStatus.Message)
			}

			// 3. Physical check: metadata file exists and matches scope
			metaData, err := os.ReadFile(filepath.Join(targetDir, core.MetadataFile))
			if err != nil {
				t.Fatalf("[%s] reading metadata file: %v", tc.name, err)
			}
			var meta core.WorkspaceMeta
			if err := json.Unmarshal(metaData, &meta); err != nil || string(meta.Scope) != tc.scope {
				t.Errorf("[%s] expected metadata scope %s, got %v", tc.name, tc.scope, meta)
			}
		})
	}
}

// TestMassive_ContainerEnvironmentVariables_LiveMCP tests that standard container environment
// variables (WORKSPACE, PROJECT_ROOT, VIVECHAK_PROJECT_ROOT, VIVECHAK_DEFAULT_ROOT) properly
// route un-targeted tool calls to the desired sandbox directory.
func TestMassive_ContainerEnvironmentVariables_LiveMCP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)

	binDir := t.TempDir()
	binName := "vck_env_test"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	envCases := []struct {
		envKey string
		name   string
	}{
		{"VIVECHAK_PROJECT_ROOT", "Canonical Vivechak Project Root"},
		{"VIVECHAK_DEFAULT_ROOT", "Default Vivechak Root Fallback"},
		{"WORKSPACE", "Standard OpenHands / GitHub Actions WORKSPACE"},
		{"PROJECT_ROOT", "Container Standard PROJECT_ROOT"},
	}

	for _, tc := range envCases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := filepath.Join(t.TempDir(), "sandboxed", tc.envKey)

			serverCmd := exec.CommandContext(ctx, binPath, "serve")
			serverCmd.Env = append(os.Environ(), tc.envKey+"="+targetDir)
			transport := &mcp.CommandTransport{Command: serverCmd}

			client := mcp.NewClient(
				&mcp.Implementation{Name: "env-test-client", Version: "1.0.0"},
				nil,
			)

			cs, err := client.Connect(ctx, transport, nil)
			if err != nil {
				t.Fatalf("connecting to server with %s: %v", tc.envKey, err)
			}
			defer cs.Close()

			// Call vivechak_init with EMPTY arguments
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      "vivechak_init",
				Arguments: map[string]any{},
			})
			if err != nil {
				t.Fatalf("callTool with %s failed: %v", tc.envKey, err)
			}
			env := parseTestEnvelope(t, res)
			if !env.Success {
				t.Fatalf("expected success with env %s, got: %s", tc.envKey, env.Message)
			}

			// Verify files landed in targetDir
			if !core.WorkspaceExists(targetDir) {
				t.Errorf("expected workspace to be created at %s via %s", targetDir, tc.envKey)
			}
		})
	}

	t.Run("safely ignores unexpanded macro ${workspaceFolder}", func(t *testing.T) {
		serverCmd := exec.CommandContext(ctx, binPath, "serve")
		serverCmd.Env = append(os.Environ(), "VIVECHAK_PROJECT_ROOT=${workspaceFolder}")
		transport := &mcp.CommandTransport{Command: serverCmd}

		client := mcp.NewClient(
			&mcp.Implementation{Name: "macro-test-client", Version: "1.0.0"},
			nil,
		)

		cs, err := client.Connect(ctx, transport, nil)
		if err != nil {
			t.Fatalf("connecting to server with macro: %v", err)
		}
		defer cs.Close()

		// Call vivechak_status with empty arguments
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_status",
			Arguments: map[string]any{},
		})
		if err != nil {
			t.Fatalf("status call with macro: %v", err)
		}
		env := parseTestEnvelope(t, res)
		// Should safely report uninitialized or resolution error, NEVER crash or resolve to literal "${workspaceFolder}"
		if strings.Contains(env.Message, "${workspaceFolder}") && !strings.Contains(env.Message, "not set") {
			t.Errorf("server mistakenly resolved literal macro: %s", env.Message)
		}
	})
}

// TestMassive_CLICommands_RealExec tests the CLI entry points (vck doctor, vck mcp-config, vck setup)
// as real executed standalone OS processes.
func TestMassive_CLICommands_RealExec(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)

	binDir := t.TempDir()
	binName := "vck_cli_massive"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	// 1. Test vck doctor on application program folder -> must exit 1 with descriptive error
	mockAppDir := filepath.Join(t.TempDir(), "AppData", "Local", "Programs", "Antigravity IDE")
	_ = os.MkdirAll(mockAppDir, 0o755)

	docFailCmd := exec.CommandContext(ctx, binPath, "doctor", mockAppDir)
	docFailOut, docFailErr := docFailCmd.CombinedOutput()
	if docFailErr == nil {
		t.Errorf("expected vck doctor on AppData program directory to fail with exit 1, got 0")
	}
	if !strings.Contains(string(docFailOut), "application or system directory") {
		t.Errorf("expected doctor output to mention application or system directory, got:\n%s", string(docFailOut))
	}

	// 2. Test vck mcp-config outputs valid JSON with vivechak entry
	cfgCmd := exec.CommandContext(ctx, binPath, "mcp-config")
	cfgOut, cfgErr := cfgCmd.CombinedOutput()
	if cfgErr != nil {
		t.Fatalf("vck mcp-config failed: %v\nOutput: %s", cfgErr, string(cfgOut))
	}
	var parsedCfg map[string]any
	if err := json.Unmarshal(cfgOut, &parsedCfg); err != nil {
		t.Fatalf("vck mcp-config output is not valid JSON: %v\nOutput: %s", err, string(cfgOut))
	}
	servers, ok := parsedCfg["mcpServers"].(map[string]any)
	if !ok || servers["vivechak"] == nil {
		t.Errorf("mcp-config missing mcpServers.vivechak: %v", parsedCfg)
	}

	// 3. Test vck setup --dry-run
	setupCmd := exec.CommandContext(ctx, binPath, "setup", "--dry-run", "cursor")
	setupOut, setupErr := setupCmd.CombinedOutput()
	if setupErr != nil {
		t.Fatalf("vck setup --dry-run failed: %v\nOutput: %s", setupErr, string(setupOut))
	}
	if !strings.Contains(string(setupOut), "[dry-run]") {
		t.Errorf("expected dry-run preview, got:\n%s", string(setupOut))
	}

	// 4. Test vck version
	verCmd := exec.CommandContext(ctx, binPath, "version")
	verOut, verErr := verCmd.CombinedOutput()
	if verErr != nil {
		t.Fatalf("vck version failed: %v\nOutput: %s", verErr, string(verOut))
	}
	if !strings.Contains(string(verOut), "vck version") && !strings.Contains(string(verOut), "vivechak") {
		t.Errorf("unexpected version output: %s", string(verOut))
	}
}

// assertFilesExist asserts that all relPaths exist within root.
func assertFilesExist(t *testing.T, root string, relPaths []string) {
	t.Helper()
	for _, p := range relPaths {
		full := filepath.Join(root, p)
		if fi, err := os.Stat(full); err != nil || fi.Size() == 0 {
			t.Errorf("expected file %s to exist and be non-empty: %v", p, err)
		}
	}
}
