package main

import (
	"context"
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

// TestMassiveLive_DepthAudit_All13Tools tests the complete 13-tool MCP roster,
// CLI commands, and all latest engine hardening features against a live compiled binary.
func TestMassiveLive_DepthAudit_All13Tools(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	// 1. Build live binary
	binDir := t.TempDir()
	binName := "vck_depth_audit"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	// 2. Start MCP server over stdio
	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "depth-audit-client", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	callAndParse := func(toolName string, args map[string]any) testEnvelope {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		})
		if err != nil {
			t.Fatalf("call to %s failed: %v", toolName, err)
		}
		return parseTestEnvelope(t, res)
	}

	// 3. Verify exactly 13 tools registered
	var toolNames []string
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("error listing tools: %v", err)
		}
		toolNames = append(toolNames, tool.Name)
	}
	if len(toolNames) != 13 {
		t.Fatalf("expected 13 registered tools, got %d: %v", len(toolNames), toolNames)
	}

	projectDir := t.TempDir()

	// =========================================================================
	// TOOL 1: vivechak_init
	// =========================================================================
	t.Run("Tool1_Init_And_Guards", func(t *testing.T) {
		// Initialize workspace
		env := callAndParse("vivechak_init", map[string]any{
			"project_root": projectDir,
			"scope":        "project",
		})
		if !env.Success {
			t.Fatalf("vivechak_init failed: %s", env.Message)
		}

		// Verify templates copied and .gitignore created
		if _, err := os.Stat(filepath.Join(projectDir, core.TemplatesDir, "SESSION.template.md")); err != nil {
			t.Errorf("SESSION.template.md missing: %v", err)
		}
		gi, err := os.ReadFile(filepath.Join(projectDir, core.ResearchDir, ".gitignore"))
		if err != nil || !strings.Contains(string(gi), "*.lock") {
			t.Errorf("expected *.lock in .gitignore, got err=%v content=%s", err, string(gi))
		}

		// Guard test: reject filesystem root
		badRoot := "/"
		if runtime.GOOS == "windows" {
			badRoot = "C:\\"
		}
		envBad := callAndParse("vivechak_init", map[string]any{
			"project_root": badRoot,
			"scope":        "project",
		})
		if envBad.Success {
			t.Errorf("expected failure when initializing root directory")
		}
	})

	// =========================================================================
	// TOOL 2: vivechak_prepare_generator
	// =========================================================================
	t.Run("Tool2_PrepareGenerator", func(t *testing.T) {
		env := callAndParse("vivechak_prepare_generator", map[string]any{
			"project_root": projectDir,
			"context":      "Distributed low-latency state machine replication engine",
		})
		if !env.Success {
			t.Fatalf("prepare_generator failed: %s", env.Message)
		}
		prompt, ok := env.Data["prompt"].(string)
		if !ok || !strings.Contains(prompt, "Distributed low-latency state machine") {
			t.Errorf("expected context in generator prompt, got: %v", env.Data["prompt"])
		}
	})

	// =========================================================================
	// TOOL 3: vivechak_save_plan (with Preamble, Epilogue, and multi-tick fences)
	// =========================================================================
	t.Run("Tool3_SavePlan_PreambleEpilogueAndFences", func(t *testing.T) {
		rawDAG := `# Distributed Replication Engine Research Pipeline

> **Tier:** Tier 1 (4-8 Sessions) · **Complexity Score:** 12
> Custom Preamble line that must be preserved by DAG serialization.

## Pipeline Topology

| Session | Title | Door | Deps | Output |
|---|---|---|---|---|
| T1-01 | Consensus Algorithm | one-way | None | sessions/T1-01.md |
| T1-02 | Network Transport | two-way | None | sessions/T1-02.md |
| T1-03 | Write-Ahead Log | one-way | T1-01, T1-02 | sessions/T1-03.md |
| SYN-01 | Architecture Synthesis | synthesis | T1-03 | research/FAD.md |

## Session Prompts

### Session T1-01: Consensus Algorithm
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Door Type** | one-way |
| **Dependencies** | None |
| **Output File** | sessions/T1-01.md |

` + "````prompt\n" + `# T1-01: Consensus Algorithm

Evaluate Raft vs Paxos vs Viewstamped Replication.
Include benchmark JSON configuration:
` + "```json\n" + `{
  "heartbeat_interval_ms": 50,
  "election_timeout_ms": 300
}
` + "```\n" + `Ensure multi-tick fences do not truncate.
` + "````\n\n" + `---

### Session T1-02: Network Transport
| Field | Value |
|---|---|
| **ID** | T1-02 |
| **Door Type** | two-way |
| **Dependencies** | None |
| **Output File** | sessions/T1-02.md |

` + "````prompt\n" + `# T1-02: Network Transport
Evaluate gRPC vs QUIC vs raw TCP.
` + "````\n\n" + `---

### Session T1-03: Write-Ahead Log
| Field | Value |
|---|---|
| **ID** | T1-03 |
| **Door Type** | one-way |
| **Dependencies** | T1-01, T1-02 |
| **Output File** | sessions/T1-03.md |

` + "````prompt\n" + `# T1-03: Write-Ahead Log
Evaluate io_uring vs synchronous mmap for WAL appending.
` + "````\n\n" + `---

### Session SYN-01: Architecture Synthesis
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Door Type** | synthesis |
| **Dependencies** | T1-03 |
| **Output File** | research/FAD.md |

` + "````prompt\n" + `# SYN-01: Architecture Synthesis
Synthesize full architecture into FAD.
` + "````\n\n" + `---

## Phase 0 Exit Gate Criteria (Epilogue)
Custom exit gate epilogue criteria to verify preservation across mutations.
`
		env := callAndParse("vivechak_save_plan", map[string]any{
			"project_root": projectDir,
			"scope":        "project",
			"content":      rawDAG,
		})
		if !env.Success {
			t.Fatalf("save_plan failed: %s", env.Message)
		}

		// Verify pipeline file on disk contains custom preamble and epilogue
		savedBytes, err := os.ReadFile(filepath.Join(projectDir, core.PipelineFile))
		if err != nil {
			t.Fatalf("reading pipeline file: %v", err)
		}
		savedStr := string(savedBytes)
		if !strings.Contains(savedStr, "Custom Preamble line that must be preserved") {
			t.Errorf("saved pipeline missing custom preamble")
		}
		if !strings.Contains(savedStr, "Phase 0 Exit Gate Criteria (Epilogue)") {
			t.Errorf("saved pipeline missing epilogue")
		}
	})

	// =========================================================================
	// TOOL 4: vivechak_status
	// =========================================================================
	t.Run("Tool4_Status", func(t *testing.T) {
		env := callAndParse("vivechak_status", map[string]any{
			"project_root": projectDir,
		})
		if !env.Success {
			t.Fatalf("status failed: %s", env.Message)
		}
		if initialized, ok := env.Data["initialized"].(bool); !ok || !initialized {
			t.Errorf("expected initialized=true, got %v", env.Data["initialized"])
		}
		if hasPipeline, ok := env.Data["has_pipeline"].(bool); !ok || !hasPipeline {
			t.Errorf("expected has_pipeline=true, got %v", env.Data["has_pipeline"])
		}
		sessCount := int(env.Data["session_count"].(float64))
		tmplCount := int(env.Data["template_count"].(float64))
		if sessCount != 0 || tmplCount != 6 {
			t.Errorf("unexpected counts: sessCount=%d, tmplCount=%d", sessCount, tmplCount)
		}
	})

	// =========================================================================
	// TOOL 5: vivechak_next_session
	// =========================================================================
	t.Run("Tool5_NextSession_Parallelism_And_CalibrationInjection", func(t *testing.T) {
		env := callAndParse("vivechak_next_session", map[string]any{
			"project_root": projectDir,
		})
		if !env.Success {
			t.Fatalf("next_session failed: %s", env.Message)
		}
		// Parallelism hint should be present because both T1-01 and T1-02 are ready
		if env.Data["parallelism_hint"] == nil {
			t.Errorf("expected parallelism_hint when multiple sessions are ready")
		}
		prompt := env.Data["prompt"].(string)
		// Active research calibration block must be injected
		if !strings.Contains(prompt, "### RESEARCH CALIBRATION (Active Rules)") {
			t.Errorf("expected research calibration block injected into non-synthesis session prompt")
		}
	})

	// =========================================================================
	// TOOL 6: vivechak_save_session (One-Way vs Two-Way Hardening & Quality Coaching)
	// =========================================================================
	t.Run("Tool6_SaveSession_OWDEnforcement_And_QualityObservations", func(t *testing.T) {
		// 1. One-way door T1-01 fails if Discovered Concerns is trivial
		badOWD := `---
session_id: T1-01
title: Consensus Algorithm
date: 2026-09-29
door_type: one-way
status: complete
---
## Key Findings
- Raft algorithm offers leader-based consensus. Grade A (https://raft.github.io)

## Discovered Concerns
None.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | https://raft.github.io | Grade A | fresh | fetched | Raft specification |
`
		envBad := callAndParse("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "T1-01",
			"content":      badOWD,
		})
		if passed, ok := envBad.Data["validation_passed"].(bool); !ok || passed {
			t.Errorf("expected validation_passed=false for one-way door with trivial concerns, got %v", envBad.Data["validation_passed"])
		}

		// 2. Fix T1-01 with substantive concerns and technical in-memory citation (not penalized)
		goodOWD := `---
session_id: T1-01
title: Consensus Algorithm
date: 2026-09-29
door_type: one-way
status: complete
---
## Key Findings
- Raft algorithm offers deterministic state machine replication. Grade A (https://raft.github.io)
- In-memory state caching maintains sub-millisecond read latency. Grade A (official docs: in-memory caching)

## Discovered Concerns
- Network partitioning during cross-DC cluster rebalancing triggers election storms unless pre-vote extension is enabled.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | https://raft.github.io | Grade A | fresh | fetched | Raft specification |
| 2 | https://etcd.io/docs/v3.5/learning/raft/ | Grade A | fresh | fetched | in-memory state caching |
`
		envGood := callAndParse("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "T1-01",
			"content":      goodOWD,
		})
		if passed, ok := envGood.Data["validation_passed"].(bool); !ok || !passed {
			t.Fatalf("expected validation_passed=true on good OWD session, got: %v (warnings: %v)", envGood.Data["validation_passed"], envGood.Warnings)
		}
		// next_step must route one-way door to red_team challenge
		if !strings.Contains(envGood.NextStep, "vivechak_challenge") || !strings.Contains(envGood.NextStep, "red_team") {
			t.Errorf("expected one-way door to route to red_team challenge in next_step, got: %s", envGood.NextStep)
		}

		// 3. Save two-way door T1-02 (vague provenance produces warning, not block)
		twoWayContent := `---
session_id: T1-02
title: Network Transport
date: 2026-09-29
door_type: two-way
status: complete
---
## Key Findings
- gRPC over HTTP/2 handles multiplexed RPC streaming. Grade B (benchmarks)

## Discovered Concerns
- High connection churn degrades HTTP/2 multiplexing benefits.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | online benchmarks | Grade B | fresh | fetched | throughput |
`
		envTwoWay := callAndParse("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "T1-02",
			"content":      twoWayContent,
		})
		if passed, ok := envTwoWay.Data["validation_passed"].(bool); !ok || !passed {
			t.Fatalf("expected validation_passed=true on two-way door, got: %v (warnings: %v)", envTwoWay.Data["validation_passed"], envTwoWay.Warnings)
		}
	})

	// =========================================================================
	// TOOL 7: vivechak_challenge
	// =========================================================================
	t.Run("Tool7_Challenge_RedTeam_And_Audit", func(t *testing.T) {
		env := callAndParse("vivechak_challenge", map[string]any{
			"project_root": projectDir,
			"session_id":   "T1-01",
			"mode":         "red_team",
		})
		if !env.Success {
			t.Fatalf("challenge failed: %s", env.Message)
		}
		promptsList, ok := env.Data["challenge_prompts"].([]any)
		if !ok || len(promptsList) == 0 {
			t.Fatalf("expected challenge_prompts in data, got: %v", env.Data)
		}
		firstPrompt := promptsList[0].(map[string]any)
		pText, ok := firstPrompt["prompt"].(string)
		if !ok || pText == "" {
			t.Errorf("expected non-empty prompt in first challenge prompt")
		}
	})

	// =========================================================================
	// TOOL 8: vivechak_record_decision
	// =========================================================================
	t.Run("Tool8_RecordDecision_OneWay_And_TwoWay", func(t *testing.T) {
		// Record one-way ADR D-001
		adr1 := `---
id: D-001
title: Consensus Architecture
status: accepted
door_type: one-way
review_trigger: Cluster size exceeds 15 nodes
informing_sessions: [T1-01]
---
# D-001: Consensus Architecture

## Context & Rationale
We adopt Raft consensus algorithm with pre-vote extension for deterministic state machine replication across the cluster.
This forms the foundational coordination mechanism and cannot be altered without cluster migration. Grade A (https://raft.github.io)

## Rejected Alternatives
1. Multi-Paxos: excessive protocol implementation complexity and lack of reference formal verification.
2. Viewstamped Replication: smaller ecosystem tooling and library adoption in target programming language.
`
		env1 := callAndParse("vivechak_record_decision", map[string]any{
			"project_root":  projectDir,
			"decision_id":   "D-001",
			"artifact_type": "decision",
			"content":       adr1,
		})
		if !env1.Success {
			t.Fatalf("record D-001 failed: %s", env1.Message)
		}

		// Record two-way ADR D-002
		adr2 := `---
id: D-002
title: Transport Protocol
status: accepted
door_type: two-way
informing_sessions: [T1-02]
---
# D-002: Transport Protocol
## Context & Rationale
We adopt gRPC with protobuf serialization. Easy to switch transport layer via abstraction interface. Grade B (benchmarks)
`
		env2 := callAndParse("vivechak_record_decision", map[string]any{
			"project_root":  projectDir,
			"decision_id":   "D-002",
			"artifact_type": "decision",
			"content":       adr2,
		})
		if !env2.Success {
			t.Fatalf("record D-002 failed: %s", env2.Message)
		}

		// Verify DECISIONS.md registry contains both with collapsible details
		decData, err := os.ReadFile(filepath.Join(projectDir, core.DecisionsFile))
		if err != nil {
			t.Fatalf("reading DECISIONS.md: %v", err)
		}
		decStr := string(decData)
		if !strings.Contains(decStr, "<details>") || !strings.Contains(decStr, "D-001") || !strings.Contains(decStr, "D-002") {
			t.Errorf("DECISIONS.md missing collapsible entries: %s", decStr)
		}
	})

	// =========================================================================
	// TOOL 9: vivechak_amend_session (with Root Mirroring on FAD)
	// =========================================================================
	t.Run("Tool9_AmendSession_RootMirroring", func(t *testing.T) {
		// Save T1-03 first
		t3Content := `---
session_id: T1-03
title: Write-Ahead Log
date: 2026-09-29
door_type: one-way
status: complete
---
## Key Findings
- Direct I/O WAL appending avoids kernel page cache pollution. Grade A (https://kernel.org)

## Discovered Concerns
- Fallocate pre-allocation alignment required on NVMe storage controllers.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | https://kernel.org | Grade A | fresh | fetched | Direct I/O |
`
		callAndParse("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "T1-03",
			"content":      t3Content,
		})

		// Save initial FAD via SYN-01
		fadContent := `---
id: SYN-01
title: Founding Architecture Document
synthesis_date: 2026-09-29
status: complete
---
# Founding Architecture Document

## Executive Summary
Replication engine with Raft consensus and gRPC transport. Grade A (https://raft.github.io)

## Technology Commitments
1. Consensus: Raft protocol. Grade A (https://raft.github.io)
2. Network: gRPC over HTTP/2. Grade B (benchmarks)
3. WAL: Direct I/O append. Grade A (https://kernel.org)
`
		envFAD := callAndParse("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "SYN-01",
			"content":      fadContent,
		})
		if !envFAD.Success {
			t.Fatalf("saving SYN-01 FAD failed: %s", envFAD.Message)
		}

		// Amend SYN-01 FAD: must mirror to root FOUNDING-ARCHITECTURE.md and return root_copy in envelope
		envAmend := callAndParse("vivechak_amend_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "SYN-01",
			"amendment":    "## Production Operations\nAutomated leader stepdown on missed heartbeat thresholds.",
		})
		if !envAmend.Success {
			t.Fatalf("amend SYN-01 failed: %s", envAmend.Message)
		}
		if envAmend.Data["root_copy"] != "FOUNDING-ARCHITECTURE.md" {
			t.Errorf("expected root_copy='FOUNDING-ARCHITECTURE.md', got: %v", envAmend.Data["root_copy"])
		}
		rootFAD, err := os.ReadFile(filepath.Join(projectDir, "FOUNDING-ARCHITECTURE.md"))
		if err != nil || !strings.Contains(string(rootFAD), "Automated leader stepdown") {
			t.Errorf("root FOUNDING-ARCHITECTURE.md not mirrored correctly: err=%v", err)
		}
	})

	// =========================================================================
	// TOOL 10: vivechak_validate (Workspace Batch Mode)
	// =========================================================================
	t.Run("Tool10_Validate_WorkspaceBatch", func(t *testing.T) {
		env := callAndParse("vivechak_validate", map[string]any{
			"project_root": projectDir,
		})
		if !env.Success {
			t.Fatalf("workspace validate failed: %s", env.Message)
		}
		sessData := env.Data["sessions"].(map[string]any)
		decData := env.Data["decisions"].(map[string]any)
		if int(sessData["errors"].(float64)) != 0 {
			t.Errorf("expected 0 session errors, got: %v", sessData["errors"])
		}
		if int(decData["valid"].(float64)) < 2 {
			t.Errorf("expected at least 2 valid decisions, got: %v", decData["valid"])
		}
	})

	// =========================================================================
	// TOOL 11: vivechak_replan (Preserving Preamble & Epilogue)
	// =========================================================================
	t.Run("Tool11_Replan_PreservesPreambleAndEpilogue", func(t *testing.T) {
		env := callAndParse("vivechak_replan", map[string]any{
			"project_root": projectDir,
			"operation":    "add_session",
			"session_id":   "T2-01",
			"layer":        "2",
			"topic":        "Telemetry & Profiling",
			"door_type":    "two-way",
			"dependencies": "T1-02",
			"prompt":       "Investigate OpenTelemetry metrics export.",
		})
		if !env.Success {
			t.Fatalf("replan add_session failed: %s", env.Message)
		}

		// Verify pipeline file preserved preamble and epilogue
		pipeBytes, err := os.ReadFile(filepath.Join(projectDir, core.PipelineFile))
		if err != nil {
			t.Fatalf("reading replanned pipeline: %v", err)
		}
		pipeStr := string(pipeBytes)
		if !strings.Contains(pipeStr, "Custom Preamble line that must be preserved") {
			t.Errorf("preamble lost after replan: %s", pipeStr)
		}
		if !strings.Contains(pipeStr, "Phase 0 Exit Gate Criteria (Epilogue)") {
			t.Errorf("epilogue lost after replan: %s", pipeStr)
		}
	})

	// =========================================================================
	// TOOL 12: vivechak_visualize
	// =========================================================================
	t.Run("Tool12_Visualize_MermaidAndTable", func(t *testing.T) {
		// Test mermaid output
		envM := callAndParse("vivechak_visualize", map[string]any{
			"project_root": projectDir,
			"format":       "mermaid",
		})
		if !envM.Success {
			t.Fatalf("visualize mermaid failed: %s", envM.Message)
		}
		mermaidStr := envM.Data["rendered"].(string)
		if !strings.Contains(mermaidStr, "graph TD") && !strings.Contains(mermaidStr, "flowchart") {
			t.Errorf("expected mermaid graph syntax, got: %s", mermaidStr)
		}

		// Test table output
		envT := callAndParse("vivechak_visualize", map[string]any{
			"project_root": projectDir,
			"format":       "table",
		})
		if !envT.Success {
			t.Fatalf("visualize table failed: %s", envT.Message)
		}
		tableStr := envT.Data["rendered"].(string)
		if !strings.Contains(tableStr, "| Session ID |") || !strings.Contains(tableStr, "T1-01") {
			t.Errorf("expected markdown table with | Session ID |, got: %s", tableStr)
		}
	})

	// =========================================================================
	// TOOL 13: vivechak_run_gate (Decoupled Track A & Track B)
	// =========================================================================
	t.Run("Tool13_RunGate_DecoupledTracks_And_AutoPersistence", func(t *testing.T) {
		// Complete the replanned T2-01 session so all DAG sessions are complete
		t201Content := `---
session_id: T2-01
title: Telemetry & Profiling
date: 2026-09-29
door_type: two-way
status: complete
---
## Key Findings
- OpenTelemetry metrics export over gRPC maintains low CPU overhead. Grade B (benchmarks)

## Discovered Concerns
- High cardinality tag dimensions can bloat memory usage.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | https://opentelemetry.io | Grade B | fresh | fetched | metrics |
`
		callAndParse("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "T2-01",
			"content":      t201Content,
		})

		env := callAndParse("vivechak_run_gate", map[string]any{
			"project_root": projectDir,
			"verbose":      true,
		})
		if !env.Success {
			t.Fatalf("run_gate failed: %s", env.Message)
		}
		// Verify PHASE-0-GATE.md auto-persisted
		gatePath := filepath.Join(projectDir, core.ResearchDir, "PHASE-0-GATE.md")
		gateBytes, err := os.ReadFile(gatePath)
		if err != nil {
			t.Fatalf("reading auto-persisted PHASE-0-GATE.md: %v", err)
		}
		gateStr := string(gateBytes)
		if !strings.Contains(gateStr, "Track A Result:** PASS") {
			t.Errorf("expected Track A PASS in gate artifact:\n%s", gateStr)
		}
		if !strings.Contains(gateStr, "Track B Result:** PASS") {
			t.Errorf("expected Track B PASS in gate artifact:\n%s", gateStr)
		}
	})

	// =========================================================================
	// SUBPROCESS CLI COMMANDS: doctor, setup, version, mcp-config
	// =========================================================================
	t.Run("CLI_Commands_Integrity", func(t *testing.T) {
		// 1. vivechak doctor on active workspace
		docCmd := exec.CommandContext(ctx, binPath, "doctor", projectDir)
		docOut, err := docCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("vivechak doctor failed: %v\noutput: %s", err, string(docOut))
		}
		if !strings.Contains(strings.ToLower(string(docOut)), "healthy") && !strings.Contains(string(docOut), "pipeline:") {
			t.Errorf("unexpected doctor output: %s", string(docOut))
		}

		// 2. vivechak version
		verCmd := exec.CommandContext(ctx, binPath, "version")
		verOut, err := verCmd.CombinedOutput()
		if err != nil || !strings.Contains(string(verOut), "vivechak") {
			t.Errorf("unexpected version output: err=%v, out=%s", err, string(verOut))
		}

		// 3. vivechak mcp-config
		cfgCmd := exec.CommandContext(ctx, binPath, "mcp-config", "--client", "cursor")
		cfgOut, err := cfgCmd.CombinedOutput()
		if err != nil || !strings.Contains(string(cfgOut), "mcpServers") {
			t.Errorf("unexpected mcp-config output: err=%v, out=%s", err, string(cfgOut))
		}

		// 4. vivechak setup in directory target
		freshDir := t.TempDir()
		setupCmd := exec.CommandContext(ctx, binPath, "setup", freshDir)
		setupOut, err := setupCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("vivechak setup failed: %v\noutput: %s", err, string(setupOut))
		}
		if _, err := os.Stat(filepath.Join(freshDir, "mcp.json")); err != nil {
			t.Errorf("mcp.json missing after setup: %v (output: %s)", err, string(setupOut))
		}
	})
}
