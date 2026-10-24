package mcputil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMassive_ComparisonLifecycle_RealE2E executes the full comparison scope workflow
// on a real filesystem directory using live MCP calls:
// Auto-Inference -> Auto-Init Save Plan -> Next Session -> Save Session -> Run Gate -> Inspect Workspace.
func TestMassive_ComparisonLifecycle_RealE2E(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Prepare generator with NO scope parameter — should auto-infer "comparison"
	prepRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_prepare_generator",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"context":      "Winston vs Pino for high throughput JSON logging in Express API",
		},
	})
	if err != nil {
		t.Fatalf("prepare_generator failed: %v", err)
	}
	prepEnv := parseEnvelope(t, prepRes)
	if !prepEnv.Success {
		t.Fatalf("prepare_generator returned error: %s", prepEnv.Message)
	}
	prepData := prepEnv.Data.(map[string]any)
	if prepData["scope"] != "comparison" {
		t.Errorf("expected scope 'comparison', got: %v", prepData["scope"])
	}

	// 2. Save plan directly without running vivechak_init first — should auto-initialize workspace
	compPlanContent := `# RESEARCH BRIEF: Winston vs Pino Logging Comparison

## DECISION
Which logging library to use for high-throughput Node.js API; informs D-001. Door: two-way.

## BRIEF
Compare Winston vs Pino for high throughput JSON logging.
`
	savePlanRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "comparison",
			"decision_id":  "C-01",
			"slug":         "logging-shootout",
			"content":      compPlanContent,
		},
	})
	if err != nil {
		t.Fatalf("save_plan failed: %v", err)
	}
	savePlanEnv := parseEnvelope(t, savePlanRes)
	if !savePlanEnv.Success {
		t.Fatalf("save_plan returned error: %s", savePlanEnv.Message)
	}

	// Verify workspace structure was auto-created and templates copied
	if _, err := os.Stat(filepath.Join(tmpDir, "research", "templates", "COMPARISON-SESSION.template.md")); err != nil {
		t.Fatalf("auto-init failed to create comparison template: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "research", "C-01-cmp-logging-shootout.md")); err != nil {
		t.Fatalf("expected comparison plan file not found: %v", err)
	}

	// 3. Next session — should retrieve C-01 prompt
	nextRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("next_session failed: %v", err)
	}
	nextEnv := parseEnvelope(t, nextRes)
	if !nextEnv.Success {
		t.Fatalf("next_session returned error: %s", nextEnv.Message)
	}
	nextData := nextEnv.Data.(map[string]any)
	if nextData["session_id"] != "C-01" {
		t.Errorf("expected session_id 'C-01', got: %v", nextData["session_id"])
	}

	// 4. Save session C-01 with complete WEP evaluation matrix and evidence grades
	sessionOutput := `---
session_id: C-01
title: Winston vs Pino Shootout
date: 2026-10-15
status: draft
scope_level: comparison
---
# Winston vs Pino Shootout

## Prior
- I believe Pino is faster than Winston because of worker-thread serialization.
- I believe Winston has wider community ecosystem.

## Research Question
Which logging library provides superior throughput and reliability for high-traffic JSON logging?

## Key Findings
- Pino throughput benchmarked at 50,000 req/sec [Grade A, official docs, https://getpino.io, fetched].
- Winston throughput benchmarked at 15,000 req/sec with default json transport [Grade B, empirical benchmark, fetched].
- Pino uses asynchronous sonic-boom logging transport avoiding event-loop lag [Grade A, source code, fetched].
- Winston provides seamless multi-transport routing to CloudWatch and Datadog [Grade A, official docs, fetched].

## Recommendation
Pino is recommended for performance-critical microservices. Winston is rejected due to 3x higher CPU overhead.

## Alternatives Considered
- Bunyan (stale maintenance)
- Log4js (higher memory consumption)

## Detailed Findings & Weighted Evaluation Matrix

| Criterion | Weight | Winston Score (1-5) | Pino Score (1-5) | Weighted Winston | Weighted Pino |
|---|---|---|---|---|---|
| Throughput | 0.40 | 3 (15k ops) | 5 (50k ops) | 1.20 | 2.00 |
| Memory Overhead | 0.30 | 3 (45MB) | 5 (18MB) | 0.90 | 1.50 |
| Ecosystem Transports | 0.30 | 5 (50+ transports) | 4 (pino-transports) | 1.50 | 1.20 |
| **Total** | **1.00** | | | **3.60** | **4.70** |

Sensitivity testing: At throughput weight -20%, Pino still maintains a decisive lead (4.30 vs 3.75).

## Concrete Failure Modes
- **Pino:** Buffered log loss on sudden process SIGKILL before sonic-boom flush.
- **Winston:** Event loop lag spikes on heavy synchronous JSON stringification.

## Discovered Concerns
Worker thread transport offloading requires Node.js v16+; older microservices require synchronous fallback mode.

## Open Questions & Risks
Reversal trigger: If enterprise compliance mandates proprietary on-prem transport with no Pino plugin, revert to Winston.

## Delta
| Prior Belief | Status | Evidence | Impact |
|---|---|---|---|
| Pino is faster | Confirmed | Official benchmarks show 3x throughput | High |
| Winston ecosystem | Confirmed | 50+ community transports verified | Medium |

## Sources & Evidence Ledger
1. Pino Official Benchmarks (https://getpino.io/#/docs/benchmarks) [Grade A, official docs, fetched]
2. Winston Transport Docs (https://github.com/winstonjs/winston) [Grade A, official docs, fetched]
`
	saveSessRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "C-01",
			"content":      sessionOutput,
		},
	})
	if err != nil {
		t.Fatalf("save_session failed: %v", err)
	}
	saveSessEnv := parseEnvelope(t, saveSessRes)
	if !saveSessEnv.Success {
		t.Fatalf("save_session returned error: %s", saveSessEnv.Message)
	}

	// 5. Run Phase 0 Exit Gate — should evaluate comparison criteria and pass
	gateRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("run_gate failed: %v", err)
	}
	gateEnv := parseEnvelope(t, gateRes)
	gateData := gateEnv.Data.(map[string]any)
	if gateData["gate_passed"] != true {
		t.Errorf("expected comparison gate to pass, got gate_passed=false, issues: %v", gateData["blocking_issues"])
	}

	// 6. Inspect workspace directly — should reflect ScopeComparison and 1 session completed
	info := core.InspectWorkspace(tmpDir)
	if info.Scope != core.ScopeComparison {
		t.Errorf("InspectWorkspace scope = %q, want %q", info.Scope, core.ScopeComparison)
	}
	if info.SessionCount != 1 {
		t.Errorf("InspectWorkspace session_count = %d, want 1", info.SessionCount)
	}
	if !info.HasPipeline {
		t.Error("InspectWorkspace has_pipeline = false, want true")
	}
}

// TestMassive_DecisionLifecycle_RealE2E executes the full single-decision workflow:
// Auto-Inference -> Plan Save -> Session Research -> Red-Team Challenge -> ADR Record -> Gate Verification.
func TestMassive_DecisionLifecycle_RealE2E(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Prepare generator with decision framing
	prepRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_prepare_generator",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"context":      "decision: PostgreSQL vs MongoDB for SaaS clinic medical billing records",
		},
	})
	if err != nil {
		t.Fatalf("prepare_generator failed: %v", err)
	}
	prepEnv := parseEnvelope(t, prepRes)
	prepData := prepEnv.Data.(map[string]any)
	if prepData["scope"] != "decision" {
		t.Errorf("expected scope 'decision', got: %v", prepData["scope"])
	}

	// 2. Save decision plan and proposed ADR
	decPlan := `# Decision Plan: D-001

| Parameter | Value |
|---|---|
| **ID** | D-001 |
| **Door** | one-way |
| **Reversal Cost** | 3 (data migration & compliance) |

### D-001-S1: Datastore Selection
- **Role**: Comparison
- **Filename**: research/sessions/D-001-S1.md
`
	adrDraft := `---
id: D-001
title: Primary Datastore for Clinic Medical Billing
status: proposed
door_type: one-way
date: 2026-10-15
---
# Primary Datastore Selection
`
	savePlanRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root":      tmpDir,
			"scope":             "decision",
			"decision_id":       "D-001",
			"slug":              "clinic-datastore",
			"content":           decPlan,
			"decisions_content": adrDraft,
		},
	})
	if err != nil {
		t.Fatalf("save_plan failed: %v", err)
	}
	savePlanEnv := parseEnvelope(t, savePlanRes)
	if !savePlanEnv.Success {
		t.Fatalf("save_plan error: %s", savePlanEnv.Message)
	}

	// 3. Save research session D-001-S1
	sessionOutput := `---
session_id: D-001-S1
title: Datastore Comparison
date: 2026-10-15
status: complete
door_type: one-way
informs_decisions: ["D-001"]
---
# Datastore Comparison

## Research Question
Postgres vs MongoDB for medical billing records.

## Key Findings
- PostgreSQL provides strict ACID compliance with row-level locks [Grade A, official docs, fetched].
- MongoDB ACID transactions introduce high latency on sharded clusters [Grade B, empirical benchmark, fetched].
- HIPAA audit logging in PostgreSQL is mature with pgaudit extension [Grade A, official docs, fetched].

## Recommendation
PostgreSQL is recommended for compliance and transactional integrity.

## Discovered Concerns
Substantive concern: HIPAA compliance requires encrypted tablespaces with pg_crypto and dedicated audit trail partitioning.

## Sources
1. PostgreSQL Documentation [Grade A, official docs, fetched]
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "D-001-S1",
			"content":      sessionOutput,
		},
	})
	if err != nil {
		t.Fatalf("save_session: %v", err)
	}

	// 4. Run adversarial red-team challenge on D-001
	chalRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_challenge",
		Arguments: map[string]any{
			"project_root":   tmpDir,
			"decision_id":    "D-001",
			"challenge_type": "premortem",
		},
	})
	if err != nil {
		t.Fatalf("challenge failed: %v", err)
	}
	chalEnv := parseEnvelope(t, chalRes)
	if !chalEnv.Success {
		t.Fatalf("challenge returned error: %s", chalEnv.Message)
	}

	// 5. Record locked decision ADR with challenge findings and reversal triggers
	finalADR := `---
id: D-001
title: Adopt PostgreSQL for Medical Billing
status: accepted
door_type: one-way
date: 2026-10-15
confidence: high
informed_by_sessions: ["D-001-S1"]
human_reviewed: false
review_trigger: "If write volume exceeds 100k tx/sec requiring horizontal write sharding"
---
# D-001: Adopt PostgreSQL for Medical Billing

## Context & Problem Statement
Medical billing requires strict ACID compliance and HIPAA auditability.

## Decision Outcome
PostgreSQL with pgaudit is chosen as the primary datastore.

## Rejected Alternatives
- MongoDB: Rejected due to transaction latency overhead on multi-document operations.

## Challenge Findings & Red-Team Stress Test
Premortem probe executed: Failure mode of schema migrations managed via pgroll zero-downtime tool.

## Reversal Triggers
If write volume exceeds 100k tx/sec requiring horizontal write sharding, re-evaluate CockroachDB.
`
	recRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"decision_id":  "D-001",
			"slug":         "clinic-datastore",
			"content":      finalADR,
		},
	})
	if err != nil {
		t.Fatalf("record_decision failed: %v", err)
	}
	recEnv := parseEnvelope(t, recRes)
	if !recEnv.Success {
		t.Fatalf("record_decision error: %s", recEnv.Message)
	}

	// Sign off decision via SignoffDecision
	wsDec, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("open workspace: %v", err)
	}
	if _, err := core.SignoffDecision(wsDec, "D-001", "Dr. Lead"); err != nil {
		wsDec.Close()
		t.Fatalf("signoff D-001: %v", err)
	}
	wsDec.Close()

	// 6. Run Phase 0 Exit Gate — should pass Track B checks
	gateRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("run_gate failed: %v", err)
	}
	gateEnv := parseEnvelope(t, gateRes)
	gateData := gateEnv.Data.(map[string]any)
	if gateData["gate_passed"] != true {
		t.Errorf("expected decision gate to pass, got gate_passed=false: %v", gateData["blocking_issues"])
	}
}

func resolveOrBuildVck(t *testing.T) string {
	t.Helper()
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}

	binName := "vck"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(repoRoot, binName)
	if _, err := os.Stat(binPath); err == nil {
		return binPath
	}

	tempBin := filepath.Join(t.TempDir(), binName)
	buildCmd := exec.Command("go", "build", "-o", tempBin, "./cmd/vivechak")
	buildCmd.Dir = repoRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("building vck binary for test: %v\noutput: %s", err, string(out))
	}
	return tempBin
}

// TestMassive_CompiledBinary_RealCLI tests the compiled vck.exe (or native binary)
// via real OS process execution to verify CLI commands work in real terminal environments.
func TestMassive_CompiledBinary_RealCLI(t *testing.T) {
	binPath := resolveOrBuildVck(t)

	// 1. Test version command
	cmdVer := exec.Command(binPath, "version")
	outVer, err := cmdVer.CombinedOutput()
	if err != nil {
		t.Fatalf("vck version failed: %v, output: %s", err, string(outVer))
	}
	if !strings.Contains(string(outVer), "0.1.0") {
		t.Errorf("expected version 0.1.0 in output, got: %s", string(outVer))
	}

	// 2. Test mcp-config command
	cmdConfig := exec.Command(binPath, "mcp-config")
	outConfig, err := cmdConfig.CombinedOutput()
	if err != nil {
		t.Fatalf("vck mcp-config failed: %v, output: %s", err, string(outConfig))
	}
	if !strings.Contains(string(outConfig), "mcpServers") || !strings.Contains(string(outConfig), "vivechak") {
		t.Errorf("mcp-config output invalid JSON block: %s", string(outConfig))
	}

	// 3. Test setup dry-run
	cmdSetup := exec.Command(binPath, "setup", "cursor", "--dry-run")
	outSetup, err := cmdSetup.CombinedOutput()
	if err != nil {
		t.Fatalf("vck setup dry-run failed: %v, output: %s", err, string(outSetup))
	}
	if !strings.Contains(string(outSetup), "Would write Vivechak configuration") {
		t.Errorf("expected dry-run preview message, got: %s", string(outSetup))
	}

	// 4. Test gate --help
	cmdGateHelp := exec.Command(binPath, "gate", "--help")
	outGateHelp, err := cmdGateHelp.CombinedOutput()
	if err != nil {
		t.Fatalf("vck gate --help failed: %v, output: %s", err, string(outGateHelp))
	}
	if !strings.Contains(string(outGateHelp), "Usage: vivechak gate") {
		t.Errorf("expected gate usage output, got: %s", string(outGateHelp))
	}
}

// TestMassive_ProjectLifecycle_RealE2E executes the full multi-tier project lifecycle:
// Init -> Save Plan with DAG -> Next Session (DAG resolution) -> Save Sessions -> Record Decisions -> Synthesis FAD -> Run Gate.
func TestMassive_ProjectLifecycle_RealE2E(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Initialize project workspace
	initRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("vivechak_init failed: %v", err)
	}
	initEnv := parseEnvelope(t, initRes)
	if !initEnv.Success {
		t.Fatalf("vivechak_init error: %s", initEnv.Message)
	}

	// 2. Save project pipeline with DAG dependencies:
	// T1-01 (Audio Ingestion) -> no deps
	// T1-02 (Transcription Dispatcher) -> no deps
	// T2-01 (Vector Storage) -> depends on T1-01
	// FAD (Synthesis) -> depends on T1-01, T1-02, T2-01
	projectPipeline := `# RESEARCH PIPELINE: Katha AI Audio Platform

## Pipeline Topology

| Session | Topic | Dependencies | Output File |
|---|---|---|---|
| T1-01 | Audio Ingestion Engine | none | research/sessions/T1-01.md |
| T1-02 | Transcription Dispatcher | none | research/sessions/T1-02.md |
| T2-01 | Vector Storage & Retrieval | T1-01 | research/sessions/T2-01.md |
| FAD | Founding Architecture Document | T1-01, T1-02, T2-01 | research/FAD.md |

## Session Prompts

### T1-01: Audio Ingestion Engine
| **ID** | T1-01 |
| **Dependencies** | none |
| **Door** | one-way |

Research high-throughput PCM audio stream processing.

### T1-02: Transcription Dispatcher
| **ID** | T1-02 |
| **Dependencies** | none |
| **Door** | two-way |

Research worker queue architectures for Whisper inference.

### T2-01: Vector Storage & Retrieval
| **ID** | T2-01 |
| **Dependencies** | T1-01 |
| **Door** | one-way |

Research Milvus vs Qdrant vs pgvector.

### FAD: Founding Architecture Document
| **ID** | FAD |
| **Dependencies** | T1-01, T1-02, T2-01 |
| **Door** | one-way |

Synthesize all sessions into FAD.
`
	decisionsRegistry := `# DECISION REGISTRY

| ID | Title | Status | Door | Reversal Cost |
|---|---|---|---|---|
| D-001 | Streaming Audio Pipeline | proposed | one-way | 4 |
| D-002 | Worker Queue Engine | proposed | two-way | 2 |
| D-003 | Vector Database | proposed | one-way | 3 |
`

	savePlanRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root":      tmpDir,
			"scope":             "project",
			"content":           projectPipeline,
			"decisions_content": decisionsRegistry,
		},
	})
	if err != nil {
		t.Fatalf("save_plan failed: %v", err)
	}
	savePlanEnv := parseEnvelope(t, savePlanRes)
	if !savePlanEnv.Success {
		t.Fatalf("save_plan error: %s", savePlanEnv.Message)
	}

	// 3. Query next session — T1-01 and T1-02 should be ready, but NOT T2-01 or FAD
	nextRes1, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("next_session failed: %v", err)
	}
	nextEnv1 := parseEnvelope(t, nextRes1)
	nextData1 := nextEnv1.Data.(map[string]any)
	primarySession1, _ := nextData1["session_id"].(string)
	otherReady1, _ := nextData1["other_ready_sessions"].([]any)
	allReady1 := []string{primarySession1}
	for _, r := range otherReady1 {
		allReady1 = append(allReady1, r.(string))
	}
	for _, id := range allReady1 {
		if id == "T2-01" || id == "FAD" {
			t.Errorf("session %s should NOT be ready before dependencies are met (all ready: %v)", id, allReady1)
		}
	}

	// 4. Save session T1-01
	s1Content := `---
session_id: T1-01
title: Audio Ingestion Engine
date: 2026-10-15
status: complete
door_type: one-way
informs_decisions: ["D-001"]
---
# Audio Ingestion Engine

## Research Question
Optimal streaming architecture for low-latency audio capture.

## Key Findings
- WebRTC with Opus codec achieves sub-100ms transmission latency [Grade A, RFC 7845, fetched].
- Memory zero-copy buffers reduce CPU utilization by 40% [Grade B, Linux kernel docs, fetched].

## Recommendation
Adopt WebRTC data channels with custom Go Opus decoders.

## Discovered Concerns
Audio retention compliance requires strict transient in-memory ring buffers without disk caching.

## Sources
1. RFC 7845 Opus Codec [Grade A, canonical, fetched]
`
	saveS1Res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-01",
			"content":      s1Content,
		},
	})
	if err != nil {
		t.Fatalf("save_session T1-01 failed: %v", err)
	}
	saveS1Env := parseEnvelope(t, saveS1Res)
	if !saveS1Env.Success {
		t.Fatalf("save_session T1-01 error: %s", saveS1Env.Message)
	}

	// 5. Query next session again — T2-01 should NOW be ready (its dependency T1-01 is complete)
	nextRes2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("next_session failed: %v", err)
	}
	nextEnv2 := parseEnvelope(t, nextRes2)
	nextData2 := nextEnv2.Data.(map[string]any)
	primarySession2, _ := nextData2["session_id"].(string)
	otherReady2, _ := nextData2["other_ready_sessions"].([]any)
	allReady2 := []string{primarySession2}
	for _, r := range otherReady2 {
		allReady2 = append(allReady2, r.(string))
	}
	hasT201 := false
	for _, id := range allReady2 {
		if id == "T2-01" {
			hasT201 = true
			break
		}
	}
	if !hasT201 {
		t.Errorf("expected T2-01 to be ready after T1-01 completed, got ready: %v", allReady2)
	}

	// 6. Save session T1-02
	s2Content := `---
session_id: T1-02
title: Transcription Dispatcher
date: 2026-10-15
status: complete
door_type: two-way
informs_decisions: ["D-002"]
---
# Transcription Dispatcher

## Research Question
Queue architecture for Whisper workers.

## Key Findings
- Redis Streams provides bounded memory queues with consumer groups [Grade B, Redis documentation, fetched].

## Recommendation
Use Redis Streams for task distribution.

## Discovered Concerns
Worker backpressure during peak traffic requires dynamic scaling.

## Sources
1. Redis Streams Documentation [Grade B, official docs, fetched]
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-02",
			"content":      s2Content,
		},
	})
	if err != nil {
		t.Fatalf("save_session T1-02 failed: %v", err)
	}

	// 7. Save session T2-01
	s3Content := `---
session_id: T2-01
title: Vector Storage & Retrieval
date: 2026-10-15
status: complete
door_type: one-way
informs_decisions: ["D-003"]
---
# Vector Storage & Retrieval

## Research Question
Dedicated vector DB vs pgvector for audio transcript embeddings.

## Key Findings
- Qdrant provides HNSW indexing with fast filtered search and payload storage [Grade A, official benchmarks, fetched].
- pgvector IVFFlat requires frequent index rebuilding as embeddings scale beyond 10M [Grade B, PostgreSQL docs, fetched].

## Recommendation
Adopt Qdrant for dedicated vector indexing.

## Discovered Concerns
Cold starts on serverless nodes can increase p99 latency to 250ms.

## Sources
1. Qdrant Documentation [Grade A, official docs, fetched]
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T2-01",
			"content":      s3Content,
		},
	})
	if err != nil {
		t.Fatalf("save_session T2-01 failed: %v", err)
	}

	// 8. Next session — FAD should now be ready
	nextRes3, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("next_session failed: %v", err)
	}
	nextEnv3 := parseEnvelope(t, nextRes3)
	nextData3 := nextEnv3.Data.(map[string]any)
	primary3, _ := nextData3["session_id"].(string)
	otherReady3, _ := nextData3["other_ready_sessions"].([]any)
	allReady3 := []string{primary3}
	for _, r := range otherReady3 {
		allReady3 = append(allReady3, r.(string))
	}
	hasFAD := false
	for _, id := range allReady3 {
		if id == "FAD" {
			hasFAD = true
			break
		}
	}
	if !hasFAD {
		t.Errorf("expected FAD to be ready after all dependencies completed, got: %v", allReady3)
	}

	// 9. Record decisions D-001, D-002, D-003
	d1Content := `---
id: D-001
title: WebRTC Audio Ingestion
status: accepted
door_type: one-way
date: 2026-10-15
confidence: high
informed_by_sessions: ["T1-01"]
human_reviewed: false
review_trigger: "If audio streaming bandwidth exceeds 500 MB/s"
---
# WebRTC Audio Ingestion
## Context
Real-time audio streaming.
## Decision
Adopt WebRTC data channels.
## Rejected Alternatives
- WebSockets: Higher jitter under network loss.
## Reversal Triggers
If audio streaming bandwidth exceeds 500 MB/s.
`
	d2Content := `---
id: D-002
title: Redis Streams Worker Dispatcher
status: accepted
door_type: two-way
date: 2026-10-15
informed_by_sessions: ["T1-02"]
---
# Redis Streams Worker Dispatcher
## Context
Task queuing for transcription.
## Decision
Use Redis Streams.
`
	d3Content := `---
id: D-003
title: Qdrant Vector Store
status: accepted
door_type: one-way
date: 2026-10-15
confidence: high
informed_by_sessions: ["T2-01"]
human_reviewed: false
review_trigger: "If vector collection exceeds 100M embeddings"
---
# Qdrant Vector Store
## Context
Vector similarity search.
## Decision
Adopt Qdrant.
## Rejected Alternatives
- Milvus: Higher operational footprint for early-stage workload.
## Reversal Triggers
If vector collection exceeds 100M embeddings.
`
	for id, content := range map[string]string{
		"D-001": d1Content,
		"D-002": d2Content,
		"D-003": d3Content,
	} {
		recRes, rErr := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"project_root": tmpDir,
				"decision_id":  id,
				"content":      content,
			},
		})
		if rErr != nil {
			t.Fatalf("record_decision %s: %v", id, rErr)
		}
		recEnv := parseEnvelope(t, recRes)
		if !recEnv.Success {
			t.Fatalf("record_decision %s failed: %s", id, recEnv.Message)
		}
	}

	// Sign off one-way door decisions
	wsProj, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("open workspace: %v", err)
	}
	if _, err := core.SignoffDecision(wsProj, "D-001", "Lead Architect"); err != nil {
		wsProj.Close()
		t.Fatalf("signoff D-001: %v", err)
	}
	if _, err := core.SignoffDecision(wsProj, "D-003", "Lead Architect"); err != nil {
		wsProj.Close()
		t.Fatalf("signoff D-003: %v", err)
	}
	wsProj.Close()

	// 10. Save FAD synthesis session
	fadContent := `---
session_id: FAD
title: Founding Architecture Document
date: 2026-10-15
status: complete
door_type: one-way
informs_decisions: ["D-001", "D-002", "D-003"]
---
# Founding Architecture Document: Katha AI Audio Platform

## System Topology & Architecture Overview
Katha ingests streaming audio via WebRTC [Grade A, RFC 7845], dispatches transcription jobs through Redis Streams [Grade B, Redis docs], and stores transcript embeddings in Qdrant [Grade A, official docs].

## Core Subsystems
1. Audio Ingestion: Go WebRTC cluster.
2. Worker Dispatch: Redis Streams consumer groups.
3. Vector Retrieval: Qdrant vector database.

## Discovered Concerns & Mitigations
- In-memory audio retention prevents disk leaks.
- Rate limiting prevents queue starvation during traffic bursts.

## Sources
1. RFC 7845 Opus Codec [Grade A, canonical, fetched]
2. Redis Streams Documentation [Grade B, official docs, fetched]
3. Qdrant Documentation [Grade A, official docs, fetched]
`
	saveFadRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "FAD",
			"content":      fadContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session FAD failed: %v", err)
	}
	saveFadEnv := parseEnvelope(t, saveFadRes)
	if !saveFadEnv.Success {
		t.Fatalf("save_session FAD error: %s", saveFadEnv.Message)
	}
	if !strings.Contains(saveFadEnv.NextStep, "vivechak_gate") && !strings.Contains(saveFadEnv.NextStep, "vivechak_run_gate") {
		t.Errorf("expected next_step to suggest vivechak_gate, got: %s", saveFadEnv.NextStep)
	}

	// 11. Run Phase 0 Gate
	gateRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("run_gate failed: %v", err)
	}
	gateEnv := parseEnvelope(t, gateRes)
	gateData := gateEnv.Data.(map[string]any)
	if gateData["gate_passed"] != true {
		t.Errorf("expected gate to pass, got gate_passed=false: %v", gateData["blocking_issues"])
	}

	// 12. Verify filesystem artifacts
	gateFile := filepath.Join(tmpDir, "research", "PHASE-0-GATE.md")
	gateBytes, err := os.ReadFile(gateFile)
	if err != nil {
		t.Fatalf("reading PHASE-0-GATE.md: %v", err)
	}
	if !strings.Contains(string(gateBytes), `verdict: "PASS"`) {
		t.Errorf("expected 'verdict: \"PASS\"' in gate file, got:\n%s", string(gateBytes))
	}

	// 13. Inspect workspace metadata
	wsInfo := core.InspectWorkspace(tmpDir)
	if wsInfo.Scope != core.ScopeProject {
		t.Errorf("expected scope 'project', got %s", wsInfo.Scope)
	}
	if !wsInfo.HasPipeline {
		t.Error("expected HasPipeline=true")
	}
	if !wsInfo.HasDecisions {
		t.Error("expected HasDecisions=true")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "research", "FAD.md")); os.IsNotExist(err) {
		if _, sErr := os.Stat(filepath.Join(tmpDir, "research", "sessions", "FAD.md")); os.IsNotExist(sErr) {
			t.Error("expected FAD.md in research/ or research/sessions/")
		}
	}
	if wsInfo.SessionCount < 4 {
		t.Errorf("expected at least 4 sessions, got %d", wsInfo.SessionCount)
	}
}

// TestMassive_Doctor_RealCLI runs the compiled binary's doctor command on various workspace states.
func TestMassive_Doctor_RealCLI(t *testing.T) {
	binPath := resolveOrBuildVck(t)

	cs := testServer(t)
	ctx := context.Background()

	// 1. Doctor on empty non-workspace directory
	emptyDir := t.TempDir()
	cmdEmpty := exec.Command(binPath, "doctor", emptyDir)
	outEmpty, _ := cmdEmpty.CombinedOutput()
	outEmptyStr := string(outEmpty)
	if !strings.Contains(outEmptyStr, "does not exist") && !strings.Contains(outEmptyStr, "FAIL") {
		t.Errorf("expected failure/uninitialized report for empty dir, got: %s", outEmptyStr)
	}

	// 2. Doctor on initialized comparison workspace via vivechak_init
	compDir := t.TempDir()
	initCompRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": compDir,
			"scope":        "comparison",
		},
	})
	if err != nil {
		t.Fatalf("vivechak_init comp: %v", err)
	}
	initCompEnv := parseEnvelope(t, initCompRes)
	if !initCompEnv.Success {
		t.Fatalf("init comp failed: %s", initCompEnv.Message)
	}

	resDir := filepath.Join(compDir, "research")
	sessDir := filepath.Join(resDir, "sessions")
	if err := os.WriteFile(filepath.Join(resDir, "C-01-comparison.md"), []byte("# Comparison Plan\n"), 0o644); err != nil {
		t.Fatalf("write comparison plan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessDir, "C-01-winston-vs-pino.md"), []byte(`---
session_id: C-01
status: complete
door_type: two-way
---
# Comparison
## Key Findings
- Pino is fast [Grade A, benchmark, fetched].
## Sources
1. Docs [Grade A, official, fetched]
`), 0o644); err != nil {
		t.Fatalf("write session: %v", err)
	}

	cmdComp := exec.Command(binPath, "doctor", compDir)
	outComp, err := cmdComp.CombinedOutput()
	outCompStr := string(outComp)
	if err != nil && !strings.Contains(outCompStr, "PASS") && !strings.Contains(outCompStr, "WARN") {
		t.Fatalf("doctor failed on comparison workspace: %v, output: %s", err, outCompStr)
	}
	if !strings.Contains(outCompStr, "Scope: comparison") {
		t.Errorf("expected doctor to output 'Scope: comparison', got: %s", outCompStr)
	}
}

// TestMassive_QualityCoaching_RealWorkflow verifies that non-blocking quality coaching observations
// (Q-GRADE-INFLATION, Q-NO-DISCOVERED-CONCERNS, etc.) surface in warnings and next_step without failing save.
func TestMassive_QualityCoaching_RealWorkflow(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Auto-init workspace via direct plan save
	planContent := `# RESEARCH BRIEF: Quality Coaching Test
## DECISION
Test quality coaching observations on one-way door session.
`
	savePlanRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "decision",
			"decision_id":  "D-001",
			"slug":         "quality-coaching",
			"content":      planContent,
		},
	})
	if err != nil {
		t.Fatalf("save_plan failed: %v", err)
	}
	savePlanEnv := parseEnvelope(t, savePlanRes)
	if !savePlanEnv.Success {
		t.Fatalf("save_plan error: %s", savePlanEnv.Message)
	}

	// 2. Save a session with unanchored Grade A claims and low findings count
	inflatedSession := `---
session_id: D-001-S1
title: Architectural Engine
date: 2026-10-15
status: complete
door_type: one-way
informs_decisions: ["D-001"]
---
# Architectural Engine

## Key Findings
- Engine provides 100k req/sec [Grade A, official docs, fetched].
- Memory consumption remains under 200MB [Grade A, official docs, fetched].

## Recommendation
Adopt Engine.

## Discovered Concerns
Substantive architectural concern: Garbage collection pauses during high heap allocation require tuning GOGC and memory ballast settings.

## Sources
1. Official Engine Benchmark Report [Grade A, official docs, fetched]
`
	saveRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "D-001-S1",
			"content":      inflatedSession,
		},
	})
	if err != nil {
		t.Fatalf("save_session failed: %v", err)
	}
	saveEnv := parseEnvelope(t, saveRes)
	if !saveEnv.Success {
		t.Fatalf("save_session failed unexpectedly: %s", saveEnv.Message)
	}

	// Coaching warnings should be present in warnings and next_step
	hasCoachingWarning := false
	for _, w := range saveEnv.Warnings {
		if strings.HasPrefix(w, "Q-") {
			hasCoachingWarning = true
			break
		}
	}
	if !hasCoachingWarning {
		t.Errorf("expected quality coaching observation (Q-*) in warnings, got: %v", saveEnv.Warnings)
	}
	if !strings.Contains(saveEnv.NextStep, "QUALITY ADVISORIES") && !strings.Contains(saveEnv.NextStep, "[Q-") {
		t.Errorf("expected coaching guidance in NextStep, got: %s", saveEnv.NextStep)
	}
}

// TestMassive_ConcurrentSubagent_Locking verifies advisory locking and atomic I/O under
// heavy multi-goroutine contention simulating concurrent subagents writing to the same workspace.
func TestMassive_ConcurrentSubagent_Locking(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize workspace
	initRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	initEnv := parseEnvelope(t, initRes)
	if !initEnv.Success {
		t.Fatalf("init failed: %s", initEnv.Message)
	}

	// Launch 10 concurrent subagents saving distinct sessions simultaneously
	const numSubagents = 10
	var wg sync.WaitGroup
	errCh := make(chan error, numSubagents)

	for i := 1; i <= numSubagents; i++ {
		wg.Add(1)
		go func(agentID int) {
			defer wg.Done()
			sid := fmt.Sprintf("T1-%02d", agentID)
			content := fmt.Sprintf(`---
session_id: %s
title: Subagent Session %d
date: 2026-10-15
status: complete
door_type: two-way
---
# Subagent Session %d

## Key Findings
- Finding 1 from agent %d [Grade B, docs, fetched].
- Finding 2 from agent %d [Grade B, docs, fetched].

## Recommendation
Recommendation from agent %d.

## Discovered Concerns
Concern from agent %d.

## Sources
1. Documentation %d [Grade B, docs, fetched]
`, sid, agentID, agentID, agentID, agentID, agentID, agentID, agentID)

			res, cErr := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_save_session",
				Arguments: map[string]any{
					"project_root": tmpDir,
					"session_id":   sid,
					"content":      content,
				},
			})
			if cErr != nil {
				errCh <- fmt.Errorf("agent %d CallTool error: %w", agentID, cErr)
				return
			}
			env := parseEnvelope(t, res)
			if !env.Success {
				errCh <- fmt.Errorf("agent %d returned failure: %s", agentID, env.Message)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for e := range errCh {
		t.Errorf("concurrent subagent error: %v", e)
	}

	// Verify all 10 session markdown files exist, are non-empty, and contain valid frontmatter
	sessionsDir := filepath.Join(tmpDir, "research", "sessions")
	entries, rErr := os.ReadDir(sessionsDir)
	if rErr != nil {
		t.Fatalf("reading sessions dir: %v", rErr)
	}
	var validSessionCount int
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".lock") || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		validSessionCount++
		data, fErr := os.ReadFile(filepath.Join(sessionsDir, e.Name()))
		if fErr != nil {
			t.Errorf("reading file %s: %v", e.Name(), fErr)
			continue
		}
		if len(data) == 0 {
			t.Errorf("file %s is empty!", e.Name())
		}
		fm, _, pErr := core.ParseFrontmatter(data)
		if pErr != nil || fm == nil {
			t.Errorf("corrupt frontmatter in %s: %v", e.Name(), pErr)
		}
	}
	if validSessionCount != numSubagents {
		t.Errorf("expected %d valid session markdown files, got %d", numSubagents, validSessionCount)
	}
}

// TestMassive_HybridManual_AutoSync tests native hybrid execution where a human or agent
// creates files directly on disk without using MCP tools, and Vivechak immediately auto-synchronizes.
func TestMassive_HybridManual_AutoSync(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize workspace
	initRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	initEnv := parseEnvelope(t, initRes)
	if !initEnv.Success {
		t.Fatalf("init error: %s", initEnv.Message)
	}

	// 1. Manually write pipeline directly to disk
	manualPipeline := `# RESEARCH PIPELINE

## Pipeline Topology

| Session | Topic | Dependencies | Output File |
|---|---|---|---|
| T1-01 | Manual Session 1 | none | research/sessions/T1-01.md |
| T2-01 | Manual Session 2 | T1-01 | research/sessions/T2-01.md |

## Session Prompts

### T1-01: Manual Session 1
| **ID** | T1-01 |
| **Dependencies** | none |
| **Door** | two-way |

Prompt 1.

### T2-01: Manual Session 2
| **ID** | T2-01 |
| **Dependencies** | T1-01 |
| **Door** | one-way |

Prompt 2.
`
	if err := os.WriteFile(filepath.Join(tmpDir, "research", "RESEARCH-PIPELINE.md"), []byte(manualPipeline), 0o644); err != nil {
		t.Fatalf("write pipeline: %v", err)
	}

	// 2. Query status via MCP — should see initialized=true and has_pipeline=true
	statusRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_status",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	statusEnv := parseEnvelope(t, statusRes)
	statusData := statusEnv.Data.(map[string]any)
	if statusData["initialized"] != true {
		t.Errorf("expected initialized=true, got: %v", statusData["initialized"])
	}
	if statusData["has_pipeline"] != true {
		t.Errorf("expected has_pipeline=true, got: %v", statusData["has_pipeline"])
	}

	// 3. Manually write completed session file T1-01 directly to disk
	manualS1 := `---
session_id: T1-01
title: Manual Session 1
date: 2026-10-15
status: complete
door_type: two-way
---
# Manual Session 1
## Key Findings
- Direct disk write works [Grade A, official docs, fetched].
## Sources
1. Docs [Grade A, official docs, fetched]
`
	if err := os.WriteFile(filepath.Join(tmpDir, "research", "sessions", "T1-01.md"), []byte(manualS1), 0o644); err != nil {
		t.Fatalf("write session: %v", err)
	}

	// 4. Query next session via MCP — T1-01 should be detected as completed, and T2-01 returned!
	nextRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("next_session failed: %v", err)
	}
	nextEnv := parseEnvelope(t, nextRes)
	nextData := nextEnv.Data.(map[string]any)
	if nextData["session_id"] != "T2-01" {
		t.Errorf("expected next session to be T2-01 after manual T1-01 completion, got: %v", nextData["session_id"])
	}

	// 5. Manually write ADR D-001 directly to disk
	manualADR := `---
id: D-001
title: Direct Disk ADR
door_type: two-way
status: accepted
date: 2026-10-15
---
# Direct Disk ADR
`
	if err := os.WriteFile(filepath.Join(tmpDir, "research", "D-001.md"), []byte(manualADR), 0o644); err != nil {
		t.Fatalf("write ADR: %v", err)
	}

	// 6. Run gate via MCP — should pick up manual ADR in decisions list
	gateRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("run_gate failed: %v", err)
	}
	gateEnv := parseEnvelope(t, gateRes)
	gateData := gateEnv.Data.(map[string]any)
	decisionsRaw, ok := gateData["decisions"].([]any)
	if !ok || len(decisionsRaw) == 0 {
		t.Fatalf("expected manual ADR D-001 in gate decisions, got: %v", gateData["decisions"])
	}
	foundD001 := false
	for _, d := range decisionsRaw {
		dm := d.(map[string]any)
		if dm["id"] == "D-001" {
			foundD001 = true
			if dm["door_type"] != "two-way" {
				t.Errorf("expected door_type two-way, got: %v", dm["door_type"])
			}
			break
		}
	}
	if !foundD001 {
		t.Errorf("expected manual ADR D-001 to be detected by gate, got decisions: %v", decisionsRaw)
	}
}

