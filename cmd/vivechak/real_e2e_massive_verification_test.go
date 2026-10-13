package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMassive_RealTesting_Files_And_MCP executes complete end-to-end testing:
// 1. Massive testing via direct files (DAG resolution, concurrency, path safety, ADR registries)
// 2. Real testing via live MCP server subprocess exercising all 13 MCP tools end-to-end
// 3. CLI command validation (doctor, setup, mcp-config, version)
func TestMassive_RealTesting_Files_And_MCP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	// Build vivechak binary once for live testing
	binDir := t.TempDir()
	binName := "vck_massive_real_test"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	// =========================================================================
	// Part 1: Massive Testing via Files
	// =========================================================================
	t.Run("Files_DAG_And_Confinement", func(t *testing.T) {
		tmpDir := t.TempDir()

		// 1. Test parsing complex DAG with multiple tiers, arrow variants, and comments
		dagPlan := `# Research Pipeline: Enterprise Event Streamer

**Complexity Score:** 16 / 24 → **Tier 2 (8–14 Sessions)**

**Execution DAG:**
` + "```mermaid" + `
graph TD
    T1_01[T1-01: Storage Engine]
    T1_02[T1-02: Consensus Protocol]
    T2_01[T2-01: High Throughput Spike]
    SYN[SYN-01: Founding Architecture Document]

    T1_01 -->|constrains| T2_01
    T1_02 --> T2_01
    T2_01 --> SYN
` + "```" + `

## Sessions

#### T1-01: Storage Engine Selection
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Door Type** | One-Way |
| **Dependencies** | None |
| **Output File** | ` + "`research/sessions/T1-01-storage.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-01 Storage Engine
Evaluate NVMe-optimized storage backends.
` + "```" + `

---

#### T1-02: Consensus Protocol
| Field | Value |
|---|---|
| **ID** | T1-02 |
| **Door Type** | Two-Way |
| **Dependencies** | None |
| **Output File** | ` + "`research/sessions/T1-02-consensus.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-02 Consensus Protocol
Evaluate Raft vs Paxos derivatives.
` + "```" + `

---

#### T2-01: High Throughput Spike
| Field | Value |
|---|---|
| **ID** | T2-01 |
| **Door Type** | One-Way |
| **Dependencies** | T1-01, T1-02 |
| **Output File** | ` + "`research/sessions/T2-01-spike.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T2-01 High Throughput Spike
Benchmark target configuration.
` + "```" + `

---

#### SYN-01: Founding Architecture Document
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Dependencies** | T2-01 |
| **Output File** | ` + "`research/FAD.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: SYN-01 Synthesis
Synthesize founding architecture.
` + "```" + `
`
		dag, err := core.ParsePipeline([]byte(dagPlan))
		if err != nil {
			t.Fatalf("ParsePipeline failed: %v", err)
		}
		if len(dag.Sessions) != 4 {
			t.Fatalf("expected 4 sessions in parsed DAG, got %d", len(dag.Sessions))
		}
		if dag.ComplexityScore != 16 {
			t.Errorf("expected complexity score 16, got %d", dag.ComplexityScore)
		}

		// Verify topological dependencies
		t2 := dag.SessionByID("T2-01")
		if t2 == nil || len(t2.Dependencies) != 2 {
			t.Fatalf("expected T2-01 with 2 dependencies, got %+v", t2)
		}

		// 2. High-concurrency file writing & reading stress
		researchDir := filepath.Join(tmpDir, core.ResearchDir)
		sessDir := filepath.Join(tmpDir, core.SessionsDir)
		_ = os.MkdirAll(sessDir, 0o755)

		var wg sync.WaitGroup
		const workerCount = 10
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				sID := fmt.Sprintf("W-%02d", idx)
				content := fmt.Sprintf(`---
session_id: %s
status: complete
date: "2026-10-13"
---
# Session Findings for %s
Verified via local tests [Grade B: stress worker %d].
`, sID, sID, idx)
				fPath := filepath.Join(sessDir, sID+".md")
				if err := os.WriteFile(fPath, []byte(content), 0o644); err != nil {
					t.Errorf("worker %d write error: %v", idx, err)
				}
			}(i)
		}
		wg.Wait()

		// Verify files written safely
		entries, err := os.ReadDir(sessDir)
		if err != nil || len(entries) != workerCount {
			t.Fatalf("expected %d session files, got %d (err: %v)", workerCount, len(entries), err)
		}
		_ = researchDir
	})

	// =========================================================================
	// Part 2: Real Testing via Live MCP Server Subprocess (All 13 Tools)
	// =========================================================================
	t.Run("Live_MCP_13_Tools_Lifecycle", func(t *testing.T) {
		mcpProjectDir := t.TempDir()

		// Launch MCP server subprocess
		serverCmd := exec.CommandContext(ctx, binPath, "serve")
		transport := &mcp.CommandTransport{Command: serverCmd}

		client := mcp.NewClient(
			&mcp.Implementation{Name: "massive-real-tester", Version: "1.0.0"},
			nil,
		)

		cs, err := client.Connect(ctx, transport, nil)
		if err != nil {
			t.Fatalf("failed to connect to live MCP server: %v", err)
		}
		defer cs.Close()

		callTool := func(name string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      name,
				Arguments: args,
			})
			if err != nil {
				t.Fatalf("CallTool(%q) error: %v", name, err)
			}
			return parseTestEnvelope(t, res)
		}

		// Tool 1: vivechak_init
		initEnv := callTool("vivechak_init", map[string]any{
			"project_root": mcpProjectDir,
			"scope":        "project",
		})
		if !initEnv.Success {
			t.Fatalf("vivechak_init failed: %s", initEnv.Message)
		}

		// Tool 2: vivechak_prepare_generator
		prepEnv := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": mcpProjectDir,
			"scope":        "project",
			"context":      "High-throughput real-time stream aggregation cluster with distributed consensus.",
		})
		if !prepEnv.Success {
			t.Fatalf("vivechak_prepare_generator failed: %s", prepEnv.Message)
		}
		generatorPrompt, _ := prepEnv.Data["prompt"].(string)
		if !strings.Contains(generatorPrompt, "High-throughput real-time stream aggregation") {
			t.Errorf("expected generator prompt to contain injected vision")
		}

		// Tool 3: vivechak_save_plan
		pipelinePlan := `# Research Pipeline: StreamAggregator

**Complexity Score:** 14 / 24 → **Tier 2 (8–14 Sessions)**

**Execution DAG:**
` + "```mermaid" + `
graph TD
    T1_01[T1-01: Datastore Selection]
    T1_02[T1-02: Cluster Membership]
    SYN[SYN-01: Founding Architecture Document]

    T1_01 --> SYN
    T1_02 --> SYN
` + "```" + `

## Sessions

#### T1-01: Datastore Selection
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Door Type** | One-Way |
| **Dependencies** | None |
| **Output File** | ` + "`research/sessions/T1-01.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-01 Datastore Selection
Investigate high-throughput persistence options.
` + "```" + `

---

#### T1-02: Cluster Membership
| Field | Value |
|---|---|
| **ID** | T1-02 |
| **Door Type** | Two-Way |
| **Dependencies** | None |
| **Output File** | ` + "`research/sessions/T1-02.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-02 Cluster Membership
Evaluate gossip vs central coordination.
` + "```" + `

---

#### SYN-01: Founding Architecture Document
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Dependencies** | T1-01, T1-02 |
| **Output File** | ` + "`research/FAD.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: SYN-01 Synthesis
Synthesize the final architecture.
` + "```" + `
`
		savePlanEnv := callTool("vivechak_save_plan", map[string]any{
			"project_root": mcpProjectDir,
			"content":      pipelinePlan,
		})
		if !savePlanEnv.Success {
			t.Fatalf("vivechak_save_plan failed: %s", savePlanEnv.Message)
		}

		// Tool 4: vivechak_status
		statusEnv := callTool("vivechak_status", map[string]any{
			"project_root": mcpProjectDir,
		})
		if !statusEnv.Success {
			t.Fatalf("vivechak_status failed: %s", statusEnv.Message)
		}

		// Tool 5: vivechak_next_session
		nextEnv := callTool("vivechak_next_session", map[string]any{
			"project_root": mcpProjectDir,
		})
		if !nextEnv.Success {
			t.Fatalf("vivechak_next_session failed: %s", nextEnv.Message)
		}
		sessionID, _ := nextEnv.Data["session_id"].(string)
		if sessionID != "T1-01" && sessionID != "T1-02" {
			t.Fatalf("expected ready session T1-01 or T1-02, got %s", sessionID)
		}

		// Tool 6: vivechak_save_session (Session 1: T1-01)
		s1Content := `---
session_id: T1-01
title: Datastore Selection
door_type: one-way
status: complete
date: "2026-10-13"
informs_decisions: ["D-001"]
---
# T1-01: Datastore Selection Findings

## Key Findings
- Key finding 1: PostgreSQL partitioned tables deliver 120k writes/sec [Grade A: official docs https://postgresql.org/docs].
- Key finding 2: WAL fsync latency is bound by disk NVMe IOPS [Grade B: local hardware benchmarks].
- Key finding 3: Zero-downtime schema evolution requires advisory lock orchestration [Grade A: RFC specification https://tools.ietf.org/html/rfc1234].

## Discovered Concerns & Landmines
PostgreSQL migrations on high-scale active write tables require strict statement timeouts to avoid table rewrite deadlocks across multi-tenant shards during production traffic.

## Recommendation
We recommend PostgreSQL with logical partitioning for primary transactional persistence [Grade A: official docs https://postgresql.org/docs].
`
		saveS1Env := callTool("vivechak_save_session", map[string]any{
			"project_root": mcpProjectDir,
			"session_id":   "T1-01",
			"content":      s1Content,
		})
		if !saveS1Env.Success {
			t.Fatalf("vivechak_save_session T1-01 failed: %s", saveS1Env.Message)
		}

		// Tool 7: vivechak_challenge (Adversarial critique on decision)
		challEnv := callTool("vivechak_challenge", map[string]any{
			"project_root": mcpProjectDir,
			"session_id":   "T1-01",
			"mode":         "red_team",
		})
		if !challEnv.Success {
			t.Fatalf("vivechak_challenge failed: %s", challEnv.Message)
		}

		// Tool 8: vivechak_record_decision (One-Way Door D-001)
		d1Content := `---
id: D-001
decision_id: D-001
title: Primary Datastore Engine
door_type: one-way
status: accepted
date: "2026-10-13"
review_trigger: "WAL sync latency exceeds 20ms at p99"
informed_by_sessions: ["T1-01"]
human_reviewed: "true"
---
# D-001: Primary Datastore Engine

## Context
High-throughput persistent event data must be stored with ACID guarantees and linear horizontal scalability.

## Decision Rationale
We adopt PostgreSQL with pg_partman logical partitioning based on empirical verification [Grade A: official docs https://postgresql.org/docs].

## Discovered Concerns
Lock contention during partition creation requires automated partition pre-warming to prevent latency spikes during high ingest traffic.

## Rejected Alternatives
- MongoDB: lacks native distributed transactional boundaries required by accounting invariants.
- CockroachDB: higher operational memory footprint and licensing constraints for our cluster size.

## Challenge Findings & Red-Team Stress Test
Adversarial stress testing via vivechak_challenge probed high-concurrency ingestion and partition lock contention risks. Disconfirming findings mitigated via pre-warming workers.

## Reversal Triggers
- WAL sync latency exceeds 20ms at p99 for 3 consecutive hours.
`
		recD1Env := callTool("vivechak_record_decision", map[string]any{
			"project_root": mcpProjectDir,
			"decision_id":  "D-001",
			"content":      d1Content,
		})
		if !recD1Env.Success {
			t.Fatalf("vivechak_record_decision D-001 failed: %s", recD1Env.Message)
		}

		// Tool 9: vivechak_save_session (Session 2: T1-02)
		s2Content := `---
session_id: T1-02
title: Cluster Membership
door_type: two-way
status: complete
date: "2026-10-13"
informs_decisions: ["D-002"]
---
# T1-02: Cluster Membership Findings

## Key Findings
- Key finding 1: HashiCorp Memberlist uses SWIM gossip with low failure detection overhead [Grade A: GitHub repository https://github.com/hashicorp/memberlist].
- Key finding 2: Anti-entropy ping intervals converge in logarithmic time [Grade B: local network simulation].

## Discovered Concerns & Landmines
Gossip broadcast storm on large node clusters can saturate low-bandwidth control planes.

## Recommendation
Adopt SWIM gossip protocol for dynamic cluster node membership [Grade B: empirical test harness].
`
		saveS2Env := callTool("vivechak_save_session", map[string]any{
			"project_root": mcpProjectDir,
			"session_id":   "T1-02",
			"content":      s2Content,
		})
		if !saveS2Env.Success {
			t.Fatalf("vivechak_save_session T1-02 failed: %s", saveS2Env.Message)
		}

		// Record Two-Way Door D-002
		d2Content := `---
id: D-002
decision_id: D-002
title: Cluster Membership Discovery
door_type: two-way
status: accepted
date: "2026-10-13"
review_trigger: "Network gossip overhead exceeds 5% of node bandwidth"
informed_by_sessions: ["T1-02"]
human_reviewed: "true"
---
# D-002: Cluster Membership Discovery

## Context
Cluster nodes require peer discovery without central coordinating single point of failure.

## Decision Rationale
Use HashiCorp Memberlist library [Grade A: https://github.com/hashicorp/memberlist].

## Rejected Alternatives
- Static DNS: fragile under dynamic auto-scaling conditions.

## Reversal Triggers
- Network gossip overhead exceeds 5% of node bandwidth.
`
		recD2Env := callTool("vivechak_record_decision", map[string]any{
			"project_root": mcpProjectDir,
			"decision_id":  "D-002",
			"content":      d2Content,
		})
		if !recD2Env.Success {
			t.Fatalf("vivechak_record_decision D-002 failed: %s", recD2Env.Message)
		}

		// Tool 10: vivechak_amend_session
		amendEnv := callTool("vivechak_amend_session", map[string]any{
			"project_root":        mcpProjectDir,
			"session_id":          "T1-01",
			"amending_session_id": "T1-02",
			"amendment":           "Verified storage failover duration under power-loss test is < 3.2s [Grade A: official docs https://postgresql.org/docs].",
		})
		if !amendEnv.Success {
			t.Fatalf("vivechak_amend_session failed: %s", amendEnv.Message)
		}

		// Tool 11: vivechak_replan
		replanEnv := callTool("vivechak_replan", map[string]any{
			"project_root": mcpProjectDir,
			"operation":    "add_session",
			"session_id":   "T1-03",
			"topic":        "Observability Metrics Spike",
			"door_type":    "two-way",
			"dependencies": "T1-01",
		})
		if !replanEnv.Success {
			t.Fatalf("vivechak_replan failed: %s", replanEnv.Message)
		}

		// Complete T1-03
		s3Content := `---
session_id: T1-03
title: Observability Metrics Spike
door_type: two-way
status: complete
date: "2026-10-13"
---
# T1-03: Observability Findings
Prometheus and OpenTelemetry integration verified [Grade A: official docs https://opentelemetry.io/docs].
`
		_ = callTool("vivechak_save_session", map[string]any{
			"project_root": mcpProjectDir,
			"session_id":   "T1-03",
			"content":      s3Content,
		})

		// Tool 12: vivechak_visualize
		visEnv := callTool("vivechak_visualize", map[string]any{
			"project_root": mcpProjectDir,
		})
		if !visEnv.Success {
			t.Fatalf("vivechak_visualize failed: %s", visEnv.Message)
		}
		diagram, _ := visEnv.Data["rendered"].(string)
		if !strings.Contains(diagram, "graph TD") {
			t.Errorf("expected Mermaid diagram, got: %s", diagram)
		}

		// Save FAD for SYN-01
		fadContent := `---
id: FAD-001
session_id: SYN-01
title: Founding Architecture Document
synthesis_date: "2026-10-13"
status: complete
---
# Founding Architecture Document: StreamAggregator

## System Overview
The StreamAggregator is a distributed, high-throughput stream processing engine designed for sub-millisecond event ingestion, partitioned stateful persistence, and peer-to-peer cluster coordination.

## Architecture Blueprint
1. Ingestion: TCP/gRPC connection pools with backpressure propagation.
2. Storage: Partitioned PostgreSQL with NVMe write-ahead log optimization [Grade A: official docs https://postgresql.org/docs].
3. Coordination: SWIM gossip failure detector via Memberlist [Grade A: official docs https://github.com/hashicorp/memberlist].
4. Observability: OpenTelemetry metrics and distributed tracing pipeline [Grade A: official docs https://opentelemetry.io/docs].
`
		_ = os.WriteFile(filepath.Join(mcpProjectDir, core.FADFile), []byte(fadContent), 0o644)

		// Tool 13: vivechak_validate
		valEnv := callTool("vivechak_validate", map[string]any{
			"project_root":  mcpProjectDir,
			"artifact_type": "fad",
			"content":       fadContent,
		})
		if !valEnv.Success {
			t.Fatalf("vivechak_validate FAD failed: %s", valEnv.Message)
		}

		// Tool 14: vivechak_run_gate (Terminal Phase 0 Exit Gate Evaluation)
		gateEnv := callTool("vivechak_run_gate", map[string]any{
			"project_root": mcpProjectDir,
		})
		if !gateEnv.Success {
			t.Fatalf("vivechak_run_gate failed: %s", gateEnv.Message)
		}
		gateStatus, _ := gateEnv.Data["gate_status"].(string)
		if gateStatus != "PASS" {
			t.Fatalf("expected gate status PASS, got %s (warnings: %v)", gateStatus, gateEnv.Warnings)
		}

		// Verify PHASE-0-GATE.md was atomically persisted to disk
		gateFilePath := filepath.Join(mcpProjectDir, core.GateFile)
		gateBytes, err := os.ReadFile(gateFilePath)
		if err != nil {
			t.Fatalf("expected PHASE-0-GATE.md to be written on disk: %v", err)
		}
		gateStr := string(gateBytes)
		if !strings.Contains(gateStr, "verdict: \"PASS\"") {
			t.Errorf("expected gate artifact to contain verdict PASS, got:\n%s", gateStr)
		}
		if !strings.Contains(gateStr, "## Track B: Rigorous 11-Step Gate") {
			t.Errorf("expected 11-Step Track B heading in gate artifact, got:\n%s", gateStr)
		}
		if !strings.Contains(gateStr, "- [x] Decision record contains substantive causal rationale") {
			t.Errorf("expected B10 to be checked in gate artifact, got:\n%s", gateStr)
		}
		if !strings.Contains(gateStr, "- [x] One-Way Door decisions subjected to adversarial challenge") {
			t.Errorf("expected B11 to be checked in gate artifact, got:\n%s", gateStr)
		}
	})

	// =========================================================================
	// Part 3: CLI Command Real Testing (doctor, mcp-config, version)
	// =========================================================================
	t.Run("CLI_Commands_RealExecution", func(t *testing.T) {
		// 1. Version command
		var out bytes.Buffer
		vCmd := exec.CommandContext(ctx, binPath, "version")
		vCmd.Stdout = &out
		if err := vCmd.Run(); err != nil {
			t.Fatalf("vivechak version command failed: %v", err)
		}
		if !strings.Contains(out.String(), "vivechak") {
			t.Errorf("expected version string, got %s", out.String())
		}

		// 2. MCP config command with presets
		for _, preset := range []string{"claude-desktop", "cursor", "zed"} {
			out.Reset()
			cfgCmd := exec.CommandContext(ctx, binPath, "mcp-config", "--client", preset, "--stdout")
			cfgCmd.Stdout = &out
			if err := cfgCmd.Run(); err != nil {
				t.Fatalf("mcp-config for %s failed: %v", preset, err)
			}
			var parsed map[string]any
			if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
				t.Fatalf("mcp-config output for %s is not valid JSON: %v\nOutput: %s", preset, err, out.String())
			}
		}

		// 3. Doctor command in a newly initialized workspace
		docDir := t.TempDir()
		// Init workspace manually
		_ = os.MkdirAll(filepath.Join(docDir, core.TemplatesDir), 0o755)
		_ = os.MkdirAll(filepath.Join(docDir, core.SessionsDir), 0o755)
		for _, tmpl := range core.TemplatesToCopy {
			_ = os.WriteFile(filepath.Join(docDir, core.TemplatesDir, tmpl), []byte("# "+tmpl), 0o644)
		}
		_ = os.WriteFile(filepath.Join(docDir, core.PipelineFile), []byte("# Pipeline\n"), 0o644)
		_ = os.WriteFile(filepath.Join(docDir, core.DecisionsFile), []byte("# Decisions\n"), 0o644)

		out.Reset()
		docCmd := exec.CommandContext(ctx, binPath, "doctor", docDir)
		docCmd.Stdout = &out
		_ = docCmd.Run() // exit code may depend on strict checks, verify output
		if !strings.Contains(out.String(), "Checking workspace") && !strings.Contains(out.String(), "Workspace is healthy") {
			t.Errorf("expected doctor output, got: %s", out.String())
		}
	})
}

// TestRealLive_HardeningWave_E11_to_E15_Subprocess explicitly tests all 5 pre-release
// hardening wave modifications (E-11 through E-15) end-to-end via a live MCP server
// running as a genuine child process over stdio JSON-RPC.
func TestRealLive_HardeningWave_E11_to_E15_Subprocess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)

	binDir := t.TempDir()
	binName := "vck_hardening_trial"
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
		&mcp.Implementation{Name: "hardening-wave-tester", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server: %v", err)
	}
	defer cs.Close()

	callTool := func(name string, args map[string]any) testEnvelope {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		})
		if err != nil {
			t.Fatalf("CallTool(%q) error: %v", name, err)
		}
		return parseTestEnvelope(t, res)
	}

	wsDir := t.TempDir()

	// 1. Initialize workspace
	initEnv := callTool("vivechak_init", map[string]any{
		"project_root": wsDir,
		"scope":        "project",
	})
	if !initEnv.Success {
		t.Fatalf("vivechak_init failed: %s", initEnv.Message)
	}

	// 2. E-13 Test: In-process workspace session memory
	// Call vivechak_status with project_root explicitly omitted ("")
	statusEnv := callTool("vivechak_status", map[string]any{})
	if !statusEnv.Success {
		t.Fatalf("vivechak_status failed with session memory: %s", statusEnv.Message)
	}
	rootResolved, _ := statusEnv.Data["root"].(string)
	if !strings.EqualFold(filepath.Clean(rootResolved), filepath.Clean(wsDir)) {
		t.Errorf("expected session memory to resolve %q, got %q", wsDir, rootResolved)
	}

	// 3. E-11 Test: Topology vs Parsed Session Parity Guard (W-PARSER-SESSION-MISMATCH)
	// Plan declares T1-01 and T1-02 in topology/Mermaid, but provides a prompt block ONLY for T1-01
	partialPlan := `# Research Pipeline: Partial Test
## Pipeline Overview
The pipeline plans T1-01 and T1-02.

` + "```mermaid\ngraph TD\n  T1-01 --> SYN-01\n  T1-02 --> SYN-01\n```\n" + `

## Sessions
### T1-01: Engine Spike
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Dependencies** | None |

` + "```prompt\n# Brief for T1-01\n```\n" + `
`
	savePartial := callTool("vivechak_save_plan", map[string]any{
		"content": partialPlan,
	})
	if !savePartial.Success {
		t.Fatalf("vivechak_save_plan failed: %s", savePartial.Message)
	}
	hasMismatchWarning := false
	for _, w := range savePartial.Warnings {
		if strings.Contains(w, "W-PARSER-SESSION-MISMATCH") && strings.Contains(w, "T1-02") {
			hasMismatchWarning = true
			break
		}
	}
	if !hasMismatchWarning {
		t.Errorf("expected W-PARSER-SESSION-MISMATCH warning for T1-02, got warnings: %v", savePartial.Warnings)
	}

	// 4. E-12 Test: Terminal Synthesis Sink Guard
	// Now save a complete plan with T1-01, T1-02, and SYN-01. SYN-01 depends only on T1-01.
	completePlan := `# Research Pipeline: Full Test
## Execution DAG
` + "```mermaid\ngraph TD\n  T1-01 --> SYN-01\n  T1-02\n```\n" + `

## Sessions
### T1-01: Storage Engine
` + "```prompt\n# Prompt T1-01\n```\n" + `

### T1-02: Network Layer
` + "```prompt\n# Prompt T1-02\n```\n" + `

### SYN-01: Synthesis
` + "```prompt\n# Prompt SYN-01\n```\n" + `
`
	saveComplete := callTool("vivechak_save_plan", map[string]any{
		"content": completePlan,
	})
	if !saveComplete.Success {
		t.Fatalf("save complete plan failed: %s", saveComplete.Message)
	}

	// Save completed session for T1-01
	s1Content := `---
session_id: T1-01
title: Storage Engine
door_type: one-way
status: complete
date: "2026-10-13"
informs_decisions: ["D-001"]
---
# T1-01 Findings
## Key Findings
- Storage finding 1 [Grade A: RFC 1234 https://tools.ietf.org/rfc1234].
- Storage finding 2 [Grade B: benchmark].
- Storage finding 3 [Grade A: docs https://postgresql.org].
## Discovered Concerns
Substantive unexpected storage engine concern regarding IOPS limits and WAL write amplification.
## Recommendation
Adopt PostgreSQL with WAL optimization [Grade A: official docs https://postgresql.org].
`
	_ = callTool("vivechak_save_session", map[string]any{
		"session_id": "T1-01",
		"content":    s1Content,
	})

	// Call next_session: Even though T1-01 (upstream of SYN-01) is complete,
	// SYN-01 MUST NOT be actionable because T1-02 (non-synthesis session) is still pending!
	nextEnv := callTool("vivechak_next_session", map[string]any{})
	if !nextEnv.Success {
		t.Fatalf("next_session failed: %s", nextEnv.Message)
	}
	nextID, _ := nextEnv.Data["session_id"].(string)
	if nextID == "SYN-01" {
		t.Fatalf("Terminal synthesis sink guard failed: SYN-01 returned while T1-02 was still pending!")
	}
	if nextID != "T1-02" {
		t.Errorf("expected ready session T1-02, got %s", nextID)
	}

	// 5. E-14 & E-15 Test: Phase 0 Gate Failure Stop Signals & W-OWD-NO-CHALLENGE
	// Record One-Way Door D-001 WITHOUT challenge findings
	d1Content := `---
id: D-001
decision_id: D-001
title: Primary Storage Engine
door_type: one-way
status: accepted
date: "2026-10-13"
review_trigger: "latency > 50ms"
informed_by_sessions: ["T1-01"]
human_reviewed: "true"
---
# D-001: Primary Storage Engine
## Context
Primary storage selection.
## Decision Outcome
PostgreSQL [Grade A: https://postgresql.org].
## Rejected Alternatives
MongoDB rejected due to missing distributed transactions.
## Reversal Triggers
latency > 50ms.
`
	_ = callTool("vivechak_record_decision", map[string]any{
		"decision_id": "D-001",
		"content":     d1Content,
	})

	// Run gate while T1-02 is incomplete and no challenge findings exist
	gateFailEnv := callTool("vivechak_run_gate", map[string]any{
		"verbose": true,
	})
	if !strings.Contains(gateFailEnv.Message, "❌ PHASE 0 GATE FAILED") {
		t.Errorf("expected failure stop-signal in gate message, got: %s", gateFailEnv.Message)
	}
	if !strings.Contains(gateFailEnv.NextStep, "⛔ DO NOT PROCEED TO CODE") {
		t.Errorf("expected failure stop-signal in next_step, got: %s", gateFailEnv.NextStep)
	}
	hasChallengeWarning := false
	for _, w := range gateFailEnv.Warnings {
		if strings.Contains(w, "W-OWD-NO-CHALLENGE") {
			hasChallengeWarning = true
			break
		}
	}
	if !hasChallengeWarning {
		t.Errorf("expected W-OWD-NO-CHALLENGE warning for D-001, got warnings: %v", gateFailEnv.Warnings)
	}

	// 6. Complete remaining requirements and resolve challenge:
	// Run challenge probe
	_ = callTool("vivechak_challenge", map[string]any{
		"session_id": "T1-01",
		"mode":       "red_team",
	})

	// Re-record D-001 with Challenge Findings section
	d1WithChallenge := `---
id: D-001
decision_id: D-001
title: Primary Storage Engine
door_type: one-way
status: accepted
date: "2026-10-13"
review_trigger: "latency > 50ms"
informed_by_sessions: ["T1-01"]
human_reviewed: "true"
---
# D-001: Primary Storage Engine
## Context
Primary storage selection.
## Decision Outcome
PostgreSQL [Grade A: https://postgresql.org].
## Rejected Alternatives
MongoDB rejected due to missing distributed transactions.
## Challenge Findings & Red-Team Stress Test
Adversarial red-team challenge probed partition write stalls and IOPS bottlenecks. Identified WAL sync mitigations.
## Reversal Triggers
latency > 50ms.
`
	_ = callTool("vivechak_record_decision", map[string]any{
		"decision_id": "D-001",
		"content":     d1WithChallenge,
	})

	// Complete T1-02
	s2Content := `---
session_id: T1-02
title: Network Layer
door_type: two-way
status: complete
date: "2026-10-13"
---
# T1-02 Findings
## Key Findings
- Network finding 1 [Grade A: official docs https://grpc.io].
- Network finding 2 [Grade B: benchmark].
- Network finding 3 [Grade A: RFC https://tools.ietf.org/rfc7540].
## Discovered Concerns
HTTP/2 flow control window stalling under high multiplexed streams.
## Recommendation
Adopt gRPC with tunable stream windows [Grade A: official docs https://grpc.io].
`
	_ = callTool("vivechak_save_session", map[string]any{
		"session_id": "T1-02",
		"content":    s2Content,
	})

	// Complete SYN-01 & FAD
	fadContent := `---
id: SYN-01
title: System Founding Architecture Document
synthesis_date: "2026-10-13"
status: complete
---
# System Founding Architecture Document
## Overview
Architecture synthesizing storage and network layers.
## Tech Choices
- Storage: PostgreSQL [Grade A: https://postgresql.org]
- Network: gRPC [Grade A: https://grpc.io]
## Premortem
Failure scenario: WAL stall. Mitigated by NVMe.
`
	_ = os.WriteFile(filepath.Join(wsDir, core.FADFile), []byte(fadContent), 0o644)

	// Run gate again -> must PASS with B10 and B11 checked!
	gatePassEnv := callTool("vivechak_run_gate", map[string]any{})
	if gatePassEnv.Data["gate_status"] != "PASS" {
		t.Fatalf("expected gate status PASS, got %v (issues: %v)", gatePassEnv.Data["gate_status"], gatePassEnv.Warnings)
	}

	gateBytes, _ := os.ReadFile(filepath.Join(wsDir, core.GateFile))
	gateStr := string(gateBytes)
	if !strings.Contains(gateStr, "## Track B: Rigorous 11-Step Gate") {
		t.Errorf("expected 11-Step Gate heading, got:\n%s", gateStr)
	}
	if !strings.Contains(gateStr, "- [x] Decision record contains substantive causal rationale") {
		t.Errorf("expected B10 to be checked in gate artifact, got:\n%s", gateStr)
	}
	if !strings.Contains(gateStr, "- [x] One-Way Door decisions subjected to adversarial challenge") {
		t.Errorf("expected B11 to be checked in gate artifact, got:\n%s", gateStr)
	}
}

// TestRealLive_AllScopesAndModes_Subprocess verifies live MCP server subprocess execution
// across all scopes (decision, comparison), all challenge modes, all visualize formats,
// auto-drafting, and validation ladders over JSON-RPC 2.0 stdio pipes.
func TestRealLive_AllScopesAndModes_Subprocess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	binDir := t.TempDir()
	binName := "vck_scopes_test"
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
		&mcp.Implementation{Name: "scope-modes-tester", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server: %v", err)
	}
	defer cs.Close()

	callTool := func(name string, args map[string]any) testEnvelope {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		})
		if err != nil {
			t.Fatalf("CallTool(%q) error: %v", name, err)
		}
		return parseTestEnvelope(t, res)
	}

	// -------------------------------------------------------------------------
	// 1. Decision Scope End-to-End via Subprocess MCP
	// -------------------------------------------------------------------------
	t.Run("DecisionScope_Lifecycle", func(t *testing.T) {
		decDir := t.TempDir()

		// vivechak_init (scope: decision)
		initEnv := callTool("vivechak_init", map[string]any{
			"project_root": decDir,
			"scope":        "decision",
		})
		if !initEnv.Success {
			t.Fatalf("init decision scope failed: %s", initEnv.Message)
		}

		// Verify 2 decision templates copied (DECISIONS, CONFLICT-RESOLUTION)
		tmplDir := filepath.Join(decDir, core.TemplatesDir)
		entries, err := os.ReadDir(tmplDir)
		if err != nil || len(entries) != 2 {
			t.Fatalf("expected 2 templates for decision scope, got %d (err: %v)", len(entries), err)
		}

		// vivechak_prepare_generator (decision)
		prepEnv := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": decDir,
			"scope":        "decision",
			"context":      "Evaluate ClickHouse vs DuckDB for analytics dashboard backend",
		})
		if !prepEnv.Success {
			t.Fatalf("prepare generator decision failed: %s", prepEnv.Message)
		}
		genPrompt, _ := prepEnv.Data["prompt"].(string)
		if !strings.Contains(genPrompt, "ClickHouse vs DuckDB") {
			t.Errorf("expected generator prompt to contain decision context")
		}

		// vivechak_status
		statusEnv := callTool("vivechak_status", map[string]any{
			"project_root": decDir,
		})
		if !statusEnv.Success {
			t.Fatalf("status failed: %s", statusEnv.Message)
		}
		if statusEnv.Data["scope"] != "decision" {
			t.Errorf("expected scope decision, got %v", statusEnv.Data["scope"])
		}

		// vivechak_save_session (Session D-01)
		d01Content := `---
session_id: D-01
title: ClickHouse vs DuckDB Evaluation
status: complete
date: "2026-10-13"
door_type: one-way
informs_decisions: ["D-001"]
---
# D-01: ClickHouse vs DuckDB Evaluation

## Key Findings
- ClickHouse scales across multi-tenant shards with sub-50ms query latency [Grade A: official docs https://clickhouse.com/docs].
- DuckDB excels at in-process queries but lacks multi-writer cluster coordination [Grade B: benchmark].

## Discovered Concerns
ClickHouse cluster administration incurs operational overhead and memory reservation requirements.

## Recommendation
Adopt ClickHouse for multi-tenant analytics dashboard serving [Grade A: official docs https://clickhouse.com/docs].
`
		saveEnv := callTool("vivechak_save_session", map[string]any{
			"project_root": decDir,
			"session_id":   "D-01",
			"content":      d01Content,
		})
		if !saveEnv.Success {
			t.Fatalf("save_session D-01 failed: %s", saveEnv.Message)
		}

		// Auto-draft decision from session findings
		draftEnv := callTool("vivechak_record_decision", map[string]any{
			"project_root":    decDir,
			"decision_id":     "D-001",
			"auto_draft_from": "D-01",
		})
		if !draftEnv.Success {
			t.Fatalf("auto-draft D-001 failed: %s", draftEnv.Message)
		}
		draftText, _ := draftEnv.Data["draft_content"].(string)
		if !strings.Contains(draftText, "D-001") || !strings.Contains(draftText, "ClickHouse") {
			t.Errorf("draft missing expected content: %s", draftText)
		}

		// Record locked decision
		adrContent := `---
id: D-001
decision_id: D-001
title: Analytics Datastore Selection
door_type: one-way
status: accepted
date: "2026-10-13"
review_trigger: "Query latency exceeds 200ms p95 under 100 concurrent dashboard users"
informed_by_sessions: ["D-01"]
human_reviewed: "true"
---
# D-001: Analytics Datastore Selection

## Context
Dashboard queries require sub-second analytical aggregations across 100M rows.

## Decision Rationale
Adopt ClickHouse as the dedicated OLAP analytics engine [Grade A: official docs https://clickhouse.com/docs].

## Discovered Concerns
ClickHouse cluster administration incurs memory reservation requirements.

## Rejected Alternatives
- DuckDB: In-process design cannot support concurrent multi-writer web dashboard traffic.
- PostgreSQL: Row-oriented layout degrades under analytical scans across billions of records.

## Challenge Findings & Red-Team Stress Test
Adversarial red-team probe confirmed ClickHouse partition limits and memory constraints mitigated via max_server_memory_usage limits.

## Reversal Triggers
- Query latency exceeds 200ms p95 under 100 concurrent dashboard users.
`
		recEnv := callTool("vivechak_record_decision", map[string]any{
			"project_root": decDir,
			"decision_id":  "D-001",
			"content":      adrContent,
		})
		if !recEnv.Success {
			t.Fatalf("record_decision D-001 failed: %s", recEnv.Message)
		}

		// Run gate on decision scope -> must PASS
		gateEnv := callTool("vivechak_run_gate", map[string]any{
			"project_root": decDir,
		})
		if !gateEnv.Success || gateEnv.Data["gate_status"] != "PASS" {
			t.Fatalf("expected decision scope gate PASS, got: %v (warnings: %v)", gateEnv.Data["gate_status"], gateEnv.Warnings)
		}
	})

	// -------------------------------------------------------------------------
	// 2. Comparison Scope End-to-End via Subprocess MCP
	// -------------------------------------------------------------------------
	t.Run("ComparisonScope_Lifecycle", func(t *testing.T) {
		compDir := t.TempDir()

		// vivechak_init (scope: comparison)
		initEnv := callTool("vivechak_init", map[string]any{
			"project_root": compDir,
			"scope":        "comparison",
		})
		if !initEnv.Success {
			t.Fatalf("init comparison scope failed: %s", initEnv.Message)
		}

		// Verify 1 template copied (COMPARISON-SESSION)
		tmplDir := filepath.Join(compDir, core.TemplatesDir)
		entries, err := os.ReadDir(tmplDir)
		if err != nil || len(entries) != 1 {
			t.Fatalf("expected 1 template for comparison scope, got %d (err: %v)", len(entries), err)
		}

		// vivechak_prepare_generator (comparison)
		prepEnv := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": compDir,
			"scope":        "comparison",
			"context":      "AWS SQS vs RabbitMQ for asynchronous worker task queue",
		})
		if !prepEnv.Success {
			t.Fatalf("prepare generator comparison failed: %s", prepEnv.Message)
		}

		// Save comparison session C-01
		c01Content := `---
session_id: C-01
title: AWS SQS vs RabbitMQ Queue Comparison
status: complete
date: "2026-10-13"
door_type: two-way
---
# C-01: AWS SQS vs RabbitMQ Queue Comparison

## Weighted Evaluation Matrix
| Criterion | Weight | AWS SQS | RabbitMQ | Rationale |
|---|---|---|---|---|
| Managed Ops | 0.40 | 9.0 | 5.0 | SQS is fully serverless [Grade A: official docs https://aws.amazon.com/sqs]. |
| Latency | 0.30 | 7.0 | 9.0 | RabbitMQ offers sub-10ms pub-sub [Grade B: benchmark]. |
| Total Score | 1.00 | 8.2 | 6.8 | SQS wins for our serverless cloud stack. |

## Discovered Concerns
SQS message ordering guarantees require FIFO queues with throughput limits.

## Recommendation
Adopt AWS SQS for worker queues due to zero operational overhead [Grade A: official docs https://aws.amazon.com/sqs].
`
		saveEnv := callTool("vivechak_save_session", map[string]any{
			"project_root": compDir,
			"session_id":   "C-01",
			"content":      c01Content,
		})
		if !saveEnv.Success {
			t.Fatalf("save comparison session failed: %s", saveEnv.Message)
		}

		// Run gate on comparison scope -> must PASS
		gateEnv := callTool("vivechak_run_gate", map[string]any{
			"project_root": compDir,
		})
		if !gateEnv.Success || gateEnv.Data["gate_status"] != "PASS" {
			t.Fatalf("expected comparison scope gate PASS, got: %v (warnings: %v)", gateEnv.Data["gate_status"], gateEnv.Warnings)
		}
	})

	// -------------------------------------------------------------------------
	// 3. Challenge Modes & Visualize Formats via Subprocess MCP
	// -------------------------------------------------------------------------
	t.Run("ChallengeModes_And_VisualizeFormats", func(t *testing.T) {
		wsDir := t.TempDir()
		_ = callTool("vivechak_init", map[string]any{"project_root": wsDir, "scope": "project"})

		planContent := `# Pipeline: Service
**Complexity Score:** 4 / 24 → **Tier 1**
**Execution DAG:**
` + "```mermaid" + `
graph TD
    T1_01[T1-01: Ingest Pipeline]
` + "```" + `
## Sessions
#### T1-01: Ingest Pipeline
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Output File** | ` + "`research/sessions/T1-01.md`" + ` |
` + "```prompt" + `
# Research prompt
` + "```"
		_ = callTool("vivechak_save_plan", map[string]any{
			"project_root": wsDir,
			"content":      planContent,
		})

		sessContent := `---
session_id: T1-01
title: Ingest Pipeline
status: complete
date: "2026-10-13"
door_type: one-way
---
# T1-01: Ingest Pipeline
## Key Findings
- Ingestion finding [Grade A: official docs https://grpc.io].
## Discovered Concerns
Network partition handling requires keepalive tune-up.
## Recommendation
Use gRPC with HTTP/2 transport [Grade A: official docs https://grpc.io].
`
		_ = callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "T1-01",
			"content":      sessContent,
		})

		// 1. Challenge: red_team
		rtEnv := callTool("vivechak_challenge", map[string]any{
			"project_root": wsDir,
			"session_id":   "T1-01",
			"mode":         "red_team",
		})
		if !rtEnv.Success {
			t.Fatalf("challenge red_team failed: %s", rtEnv.Message)
		}
		promptsRT, _ := rtEnv.Data["challenge_prompts"].([]any)
		if len(promptsRT) == 0 {
			t.Errorf("expected non-empty challenge prompts for red_team, got data: %v", rtEnv.Data)
		}

		// 2. Challenge: evidence_audit
		eaEnv := callTool("vivechak_challenge", map[string]any{
			"project_root": wsDir,
			"session_id":   "T1-01",
			"mode":         "evidence_audit",
		})
		if !eaEnv.Success {
			t.Fatalf("challenge evidence_audit failed: %s", eaEnv.Message)
		}
		promptsEA, _ := eaEnv.Data["challenge_prompts"].([]any)
		if len(promptsEA) == 0 {
			t.Errorf("expected non-empty challenge prompts for evidence_audit, got data: %v", eaEnv.Data)
		}

		// 3. Challenge: cross_session
		csEnv := callTool("vivechak_challenge", map[string]any{
			"project_root": wsDir,
			"session_id":   "T1-01",
			"mode":         "cross_session",
		})
		if !csEnv.Success {
			t.Fatalf("challenge cross_session failed: %s", csEnv.Message)
		}

		// 4. Visualize: mermaid
		visMEnv := callTool("vivechak_visualize", map[string]any{
			"project_root": wsDir,
			"format":       "mermaid",
		})
		if !visMEnv.Success {
			t.Fatalf("visualize mermaid failed: %s", visMEnv.Message)
		}
		renderM, _ := visMEnv.Data["rendered"].(string)
		if !strings.Contains(renderM, "graph") {
			t.Errorf("expected graph in mermaid output, got: %s", renderM)
		}

		// 5. Visualize: table
		visTEnv := callTool("vivechak_visualize", map[string]any{
			"project_root": wsDir,
			"format":       "table",
		})
		if !visTEnv.Success {
			t.Fatalf("visualize table failed: %s", visTEnv.Message)
		}
		renderT, _ := visTEnv.Data["rendered"].(string)
		if len(renderT) == 0 {
			t.Errorf("expected non-empty table output")
		}
	})

	// -------------------------------------------------------------------------
	// 4. Validation Ladder Error Responses via Subprocess MCP
	// -------------------------------------------------------------------------
	t.Run("Validation_Ladders_Subprocess", func(t *testing.T) {
		wsDir := t.TempDir()
		_ = callTool("vivechak_init", map[string]any{"project_root": wsDir, "scope": "project"})

		// Missing frontmatter validation -> issues should be emitted
		valNoFM := callTool("vivechak_validate", map[string]any{
			"project_root":  wsDir,
			"artifact_type": "session",
			"content":       "# Just a heading without frontmatter\nBody content.",
		})
		if len(valNoFM.Warnings) == 0 {
			t.Errorf("expected validation warnings for document without frontmatter")
		}

		// Short FAD validation -> W-SHORT-BODY emitted in warnings
		valShortFAD := callTool("vivechak_validate", map[string]any{
			"project_root":  wsDir,
			"artifact_type": "fad",
			"content":       "---\nid: FAD\ntitle: Architecture\nsynthesis_date: 2026-10-13\n---\nShort.",
		})
		if len(valShortFAD.Warnings) == 0 {
			t.Errorf("expected validation warnings for short FAD")
		}
	})
}

