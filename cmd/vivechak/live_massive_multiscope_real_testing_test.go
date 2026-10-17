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

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestLiveMCP_MassiveMultiScope_RealTesting executes an end-to-end, zero-trust live integration suite
// against a genuinely compiled vck binary over stdio JSON-RPC. It thoroughly tests:
// 1. Comparison Scope (pino vs winston, auto-inference, WEP matrix on disk, doctor verification)
// 2. Decision Scope (PostgreSQL to ClickHouse, auto-inference, ADR on disk, DECISIONS.md registry, doctor verification)
// 3. Project Scope (Full pipeline, 6 templates, DAG traversal, upstream injection, FAD synthesis, Phase 0 Gate)
// 4. Developer CLI tooling (vck doctor, vck setup, vck mcp-config, vck version)
func TestLiveMCP_MassiveMultiScope_RealTesting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping massive live multi-scope testing in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	// Step 1: Build the real native binary
	binDir := t.TempDir()
	binName := "vck_real_multiscope_test"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to compile native vivechak binary: %v\noutput: %s", err, string(out))
	}

	// Step 2: Spawn live MCP server as a real subprocess over stdio JSON-RPC
	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "massive-multiscope-tester", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server over stdio: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	callTool := func(toolName string, args map[string]any) testEnvelope {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		})
		if err != nil {
			t.Fatalf("cs.CallTool(%q) protocol failure: %v", toolName, err)
		}
		return parseTestEnvelope(t, res)
	}

	// =========================================================================
	// TEST 1: COMPARISON SCOPE (Zero DAG, Zero Gate, 1 Session WEP Matrix)
	// =========================================================================
	t.Run("ComparisonScope_PinoVsWinston_FullFlow", func(t *testing.T) {
		compRoot := t.TempDir()

		// 1.1: Call prepare_generator with natural language comparison and omitted scope
		env1 := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": compRoot,
			"context":      "Which one should I choose between Pino and Winston for high throughput logging?",
		})
		if !env1.Success {
			t.Fatalf("prepare_generator comparison failed: %s", env1.Message)
		}

		data1 := env1.Data
		if data1 == nil {
			t.Fatalf("expected data map in prepare_generator response, got nil")
		}
		if data1["scope"] != "comparison" {
			t.Errorf("expected auto-inferred scope 'comparison', got: %v", data1["scope"])
		}

		// Verify scope-aware next_step
		if !strings.Contains(env1.NextStep, "WEP comparison matrix") {
			t.Errorf("expected comparison next_step to mention 'WEP comparison matrix', got: %q", env1.NextStep)
		}

		// 1.2: Save WEP matrix directly via vivechak_save_plan
		wepContent := `# COMPARISON: Pino vs Winston for High-Throughput Node Ingestion

## Summary Recommendation
Pino is recommended for high-throughput stream processing based on V8 serialization optimizations.

## Evaluated Options
- Option 1: Pino (Grade A - official docs https://getpino.io)
- Option 2: Winston (Grade A - official docs https://github.com/winstonjs/winston)

## Weighted Evaluation Protocol (WEP) Matrix
| Criterion | Weight | Pino (1-5) | Winston (1-5) | Evidence |
|---|---|---|---|---|
| Raw Throughput | 0.40 | 5 | 3 | Grade B (benchmarks) |
| Memory Overhead | 0.30 | 5 | 3 | Grade B (heap snapshots) |
| Plugin Ecosystem | 0.30 | 4 | 5 | Grade A (npm downloads) |
| **Weighted Total** | **1.00** | **4.70** | **3.60** | |

## Sensitivity Analysis
Shifting Raw Throughput weight by ±20% maintains Pino's lead (4.60 vs 3.70).
`
		env2 := callTool("vivechak_save_plan", map[string]any{
			"project_root": compRoot,
			"scope":        "comparison",
			"decision_id":  "C-001",
			"slug":         "pino-vs-winston",
			"content":      wepContent,
		})
		if !env2.Success {
			t.Fatalf("save_plan for comparison failed: %s", env2.Message)
		}

		// Verify on-disk file creation
		targetFile := filepath.Join(compRoot, "research", "C-001-cmp-pino-vs-winston.md")
		if _, err := os.Stat(targetFile); err != nil {
			t.Fatalf("comparison matrix file not created at %s: %v", targetFile, err)
		}

		// Verify only 1 template copied
		tplDir := filepath.Join(compRoot, "research", "templates")
		entries, err := os.ReadDir(tplDir)
		if err != nil {
			t.Fatalf("failed to read templates dir: %v", err)
		}
		if len(entries) != 1 || entries[0].Name() != "COMPARISON-SESSION.template.md" {
			t.Errorf("expected exactly 1 template (COMPARISON-SESSION.template.md), found: %v", entries)
		}

		// Verify CLI doctor reports healthy on this comparison workspace
		docCmd := exec.CommandContext(ctx, binPath, "doctor", compRoot)
		docOut, docErr := docCmd.CombinedOutput()
		if docErr != nil {
			t.Fatalf("vck doctor failed on comparison workspace: %v\noutput: %s", docErr, string(docOut))
		}
		if !strings.Contains(string(docOut), "Workspace is healthy.") {
			t.Errorf("expected 'Workspace is healthy.', got: %s", string(docOut))
		}
	})

	// =========================================================================
	// TEST 2: DECISION SCOPE (ADR D-NNN, 1-3 Sessions, Reversal Triggers)
	// =========================================================================
	t.Run("DecisionScope_PostgresToClickhouse_FullFlow", func(t *testing.T) {
		decRoot := t.TempDir()

		// 2.1: Call prepare_generator with single architectural dilemma
		env1 := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": decRoot,
			"context":      "Should we migrate our analytics datastore from PostgreSQL to ClickHouse?",
		})
		if !env1.Success {
			t.Fatalf("prepare_generator decision failed: %s", env1.Message)
		}

		data1 := env1.Data
		if data1 == nil {
			t.Fatalf("expected data map in prepare_generator response, got nil")
		}
		if data1["scope"] != "decision" {
			t.Errorf("expected auto-inferred scope 'decision', got: %v", data1["scope"])
		}

		// Verify scope-aware next_step
		if !strings.Contains(env1.NextStep, "decision plan and ADR skeleton") {
			t.Errorf("expected decision next_step to mention 'decision plan and ADR skeleton', got: %q", env1.NextStep)
		}

		// 2.2: Save decision plan
		planContent := `# Decision Plan: D-020 Datastore Migration
Profile: R=2, N=1, B=2, X=1 -> One-Way Door.
### D-020-S1: Datastore Landscape
`
		env2 := callTool("vivechak_save_plan", map[string]any{
			"project_root": decRoot,
			"scope":        "decision",
			"decision_id":  "D-020",
			"slug":         "clickhouse-migration",
			"content":      planContent,
		})
		if !env2.Success {
			t.Fatalf("save_plan for decision failed: %s", env2.Message)
		}

		// Verify plan on disk
		planFile := filepath.Join(decRoot, "research", "D-020-plan.md")
		if _, err := os.Stat(planFile); err != nil {
			t.Fatalf("decision plan file not found at %s: %v", planFile, err)
		}

		// 2.3: Stress test with adversarial red-team challenge
		envChal := callTool("vivechak_challenge", map[string]any{
			"project_root": decRoot,
			"decision_id":  "D-020",
			"mode":         "red_team",
		})
		if !envChal.Success {
			t.Fatalf("vivechak_challenge failed: %s", envChal.Message)
		}

		// 2.4: Lock canonical ADR with One-Way Door and reversal triggers
		adrContent := `---
id: D-020
title: Datastore Migration to ClickHouse
status: accepted
door_type: one-way
date: 2026-10-18
confidence: high
evidence_refs: [E-001, E-002]
review_trigger: Re-evaluate if operational overhead exceeds 10 DevOps hours/month
review_date: 2027-04-18
prediction: ClickHouse will sustain 20M events/day with <200ms P95 query latency
tags: [datastore, analytics]
authored_by: platform-lead
human_reviewed: true
schema_version: "0.1.0"
---

## Context & Problem Statement
PostgreSQL write throughput is saturating at 10M events/day.

## Evaluated Options
- ClickHouse (Grade A - https://clickhouse.com/docs)
- PostgreSQL + TimescaleDB (Grade A - https://docs.timescale.com)

## Decision Outcome & Rationale
Migrate event ingestion to ClickHouse. Primary rationale is columnar compression (4x reduction) and vectorized execution.

## Rejected Alternatives & Tradeoffs
TimescaleDB rejected due to high I/O saturation on aggregation queries.

## Failure Modes & Reversal Triggers
Reversal trigger: Re-evaluate if cluster maintenance requires dedicated SRE hire.
`
		env3 := callTool("vivechak_record_decision", map[string]any{
			"project_root": decRoot,
			"decision_id":  "D-020",
			"slug":         "clickhouse-migration",
			"content":      adrContent,
		})
		if !env3.Success {
			t.Fatalf("record_decision failed: %s", env3.Message)
		}

		// Verify on-disk ADR and DECISIONS.md registry
		adrFile := filepath.Join(decRoot, "research", "D-020-clickhouse-migration.md")
		if _, err := os.Stat(adrFile); err != nil {
			t.Fatalf("ADR file not found at %s: %v", adrFile, err)
		}

		decRegFile := filepath.Join(decRoot, "research", "DECISIONS.md")
		decBytes, err := os.ReadFile(decRegFile)
		if err != nil {
			t.Fatalf("failed to read DECISIONS.md: %v", err)
		}
		if !strings.Contains(string(decBytes), "D-020") {
			t.Errorf("DECISIONS.md missing registered D-020 entry")
		}

		// Verify exactly 2 templates copied for decision scope
		tplDir := filepath.Join(decRoot, "research", "templates")
		entries, err := os.ReadDir(tplDir)
		if err != nil {
			t.Fatalf("failed to read templates dir: %v", err)
		}
		if len(entries) != 2 {
			t.Errorf("expected exactly 2 templates for decision scope, found: %d", len(entries))
		}

		// Verify CLI doctor passes
		docCmd := exec.CommandContext(ctx, binPath, "doctor", decRoot)
		docOut, docErr := docCmd.CombinedOutput()
		if docErr != nil {
			t.Fatalf("vck doctor failed on decision workspace: %v\noutput: %s", docErr, string(docOut))
		}
		if !strings.Contains(string(docOut), "Workspace is healthy.") {
			t.Errorf("expected 'Workspace is healthy.', got: %s", string(docOut))
		}
	})

	// =========================================================================
	// TEST 3: PROJECT SCOPE (Full Pipeline -> DAG -> FAD -> Phase 0 Gate)
	// =========================================================================
	t.Run("ProjectScope_FullPipeline_FAD_And_Gate", func(t *testing.T) {
		projRoot := t.TempDir()

		// 3.1: Explicit project init
		envInit := callTool("vivechak_init", map[string]any{
			"project_root": projRoot,
			"scope":        "project",
		})
		if !envInit.Success {
			t.Fatalf("vivechak_init failed: %s", envInit.Message)
		}

		// Verify all 6 templates copied
		tplDir := filepath.Join(projRoot, "research", "templates")
		entries, err := os.ReadDir(tplDir)
		if err != nil {
			t.Fatalf("failed to read templates dir: %v", err)
		}
		if len(entries) != 6 {
			t.Errorf("expected 6 templates for project scope, got: %d", len(entries))
		}

		// 3.2: Prepare generator for full project
		envPrep := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": projRoot,
			"scope":        "project",
			"context":      "Architect a real-time event streaming pipeline in Go handling 50k events/sec with Kafka and Postgres",
		})
		if !envPrep.Success {
			t.Fatalf("prepare_generator project failed: %s", envPrep.Message)
		}
		if !strings.Contains(envPrep.NextStep, "generate your research pipeline") {
			t.Errorf("expected project next_step to mention 'generate your research pipeline', got: %q", envPrep.NextStep)
		}

		// 3.3: Save DAG pipeline with 2 research sessions and 1 synthesis session
		dagPipeline := `# RESEARCH PIPELINE

## SESSIONS

#### T1-01: Message Queue Architecture
| **ID** | T1-01 |
| **Layer** | 0 |
| **Dependencies** | None |
| **Door Type** | One-Way |
| **Decision Ref** | D-001 |
| **Output File** | sessions/T1-01-queue.md |

` + "```prompt\n# RESEARCH BRIEF: T1-01 Queue\n```" + `

#### T2-01: State Persistence
| **ID** | T2-01 |
| **Layer** | 1 |
| **Dependencies** | T1-01 |
| **Door Type** | One-Way |
| **Decision Ref** | D-002 |
| **Output File** | sessions/T2-01-persistence.md |

` + "```prompt\n# RESEARCH BRIEF: T2-01 Persistence\n```" + `

#### SYN-01: Founding Architecture Synthesis
| **ID** | SYN-01 |
| **Layer** | 2 |
| **Dependencies** | T1-01, T2-01 |
| **Door Type** | Two-Way |
| **Decision Ref** | None |
| **Output File** | FAD.md |

` + "```prompt\n# RESEARCH BRIEF: SYN-01 Grand Synthesis\n```"
		envPlan := callTool("vivechak_save_plan", map[string]any{
			"project_root": projRoot,
			"scope":        "project",
			"content":      dagPipeline,
		})
		if !envPlan.Success {
			t.Fatalf("save_plan project failed: %s", envPlan.Message)
		}

		// 3.4: Traverse DAG - NextSession should give T1-01
		envNext1 := callTool("vivechak_next_session", map[string]any{
			"project_root": projRoot,
		})
		if !envNext1.Success {
			t.Fatalf("next_session failed: %s", envNext1.Message)
		}
		dataNext1 := envNext1.Data
		if dataNext1["session_id"] != "T1-01" {
			t.Fatalf("expected next session T1-01, got: %v", dataNext1["session_id"])
		}

		// 3.5: Save session T1-01
		s1Content := `---
session_id: T1-01
title: Message Queue Architecture
status: complete
date: 2026-10-18
door_type: one-way
informs_decisions: [D-001]
confidence: high
---

## Prior
I believe Kafka is required because of throughput.

## Research Question
Which message queue architecture satisfies 50k events/sec?

## Key Findings
- Apache Kafka supports 100k+ msgs/sec per partition (Grade A - official docs https://kafka.apache.org).
- Redpanda provides Kafka API compatibility with lower memory footprint (Grade B - benchmark tests).
- RabbitMQ saturates CPU at 25k msgs/sec under persistent delivery (Grade B - independent testing).

## Recommendation
Use Apache Kafka for primary message ingestion.

## Alternatives Considered
RabbitMQ rejected due to throughput ceiling.

## Open Questions & Risks
Partition rebalancing latency during consumer pod restarts.

## Discovered Concerns
Disk I/O saturation during consumer lag recovery requires dedicated NVMe storage volumes.

## Delta
| Prior Belief | Status | Evidence | Impact |
|---|---|---|---|
| Kafka is needed | Confirmed | Grade A docs | Locks pipeline architecture |

## Sources & Evidence Ledger
| Claim | Grade | Modifiers | Verification | Source |
|---|---|---|---|---|
| Kafka throughput | Grade A | corroborated · fresh · direct | fetched | https://kafka.apache.org |
`
		envSaveS1 := callTool("vivechak_save_session", map[string]any{
			"project_root": projRoot,
			"session_id":   "T1-01",
			"content":      s1Content,
		})
		if !envSaveS1.Success {
			t.Fatalf("save_session T1-01 failed: %s", envSaveS1.Message)
		}

		// 3.6: Record decision D-001 for Kafka
		d1Content := `---
id: D-001
title: Primary Ingestion Queue Selection
status: accepted
door_type: one-way
date: 2026-10-18
confidence: high
evidence_refs: [E-001]
review_trigger: Re-evaluate if throughput falls below 10k ops/sec
review_date: 2027-04-18
prediction: Kafka will handle 50k events/sec with zero dropped messages
tags: [infrastructure, queue]
authored_by: architect
human_reviewed: true
schema_version: "0.1.0"
---

## Context & Problem Statement
High-throughput ingestion layer requires partitioned queue.

## Evaluated Options
- Apache Kafka (Grade A - https://kafka.apache.org)
- RabbitMQ (Grade B)

## Decision Outcome & Rationale
Deploy Apache Kafka.

## Rejected Alternatives & Tradeoffs
RabbitMQ rejected due to throughput bottleneck.

## Failure Modes & Reversal Triggers
Broker disk exhaustion.
`
		envRecD1 := callTool("vivechak_record_decision", map[string]any{
			"project_root": projRoot,
			"decision_id":  "D-001",
			"slug":         "ingestion-queue",
			"content":      d1Content,
		})
		if !envRecD1.Success {
			t.Fatalf("record_decision D-001 failed: %s", envRecD1.Message)
		}

		// 3.7: NextSession should now yield T2-01 with upstream context from T1-01 injected!
		envNext2 := callTool("vivechak_next_session", map[string]any{
			"project_root": projRoot,
		})
		if !envNext2.Success {
			t.Fatalf("next_session failed for T2-01: %s", envNext2.Message)
		}
		dataNext2 := envNext2.Data
		if dataNext2["session_id"] != "T2-01" {
			t.Fatalf("expected session T2-01, got: %v", dataNext2["session_id"])
		}
		promptT2, _ := dataNext2["prompt"].(string)
		if !strings.Contains(promptT2, "UPSTREAM RESEARCH CONTEXT") {
			t.Errorf("expected upstream research context injected into T2-01 prompt")
		}

		// 3.8: Save T2-01 session
		s2Content := `---
session_id: T2-01
title: State Persistence
status: complete
date: 2026-10-18
door_type: one-way
informs_decisions: [D-002]
confidence: high
---

## Prior
I believe PostgreSQL 16 is sufficient.

## Research Question
Can PostgreSQL sustain 50k events/sec?

## Key Findings
- PostgreSQL 16 with bulk copy sustains 60k inserts/sec (Grade A - https://postgresql.org/docs).
- Single-row inserts saturate connection pool at 8k inserts/sec (Grade B - benchmark).

## Recommendation
Use PostgreSQL 16 with batched micro-commit buffer.

## Alternatives Considered
DynamoDB rejected due to egress pricing at scale.

## Open Questions & Risks
WAL replication lag during traffic spikes.

## Discovered Concerns
Vacuum maintenance overhead during heavy write cycles requires custom autovacuum tuning parameters.

## Delta
| Prior Belief | Status | Evidence | Impact |
|---|---|---|---|
| Postgres works | Confirmed | Grade A docs | Confirms persistence layer |

## Sources & Evidence Ledger
| Claim | Grade | Modifiers | Verification | Source |
|---|---|---|---|---|
| Postgres bulk copy | Grade A | corroborated · fresh · direct | fetched | https://postgresql.org/docs |
`
		envSaveS2 := callTool("vivechak_save_session", map[string]any{
			"project_root": projRoot,
			"session_id":   "T2-01",
			"content":      s2Content,
		})
		if !envSaveS2.Success {
			t.Fatalf("save_session T2-01 failed: %s", envSaveS2.Message)
		}

		// Record D-002
		d2Content := `---
id: D-002
title: State Persistence Engine
status: accepted
door_type: one-way
date: 2026-10-18
confidence: high
evidence_refs: [E-002]
review_trigger: Re-evaluate if storage exceeds 5TB
review_date: 2027-04-18
prediction: PostgreSQL 16 handles projected scale
tags: [database]
authored_by: architect
human_reviewed: true
schema_version: "0.1.0"
---

## Context & Problem Statement
Relational state storage.

## Evaluated Options
- PostgreSQL 16 (Grade A - https://postgresql.org/docs)

## Decision Outcome & Rationale
Adopt PostgreSQL 16.

## Rejected Alternatives & Tradeoffs
MongoDB was evaluated and rejected due to absence of ACID transaction guarantees across normalized relational schemas.

## Failure Modes & Reversal Triggers
Connection pool exhaustion under unexpected concurrent tenant queries.
`
		callTool("vivechak_record_decision", map[string]any{
			"project_root": projRoot,
			"decision_id":  "D-002",
			"slug":         "persistence-engine",
			"content":      d2Content,
		})

		// 3.9: NextSession should now yield SYN-01 (Grand Synthesis) with all findings
		envNextSyn := callTool("vivechak_next_session", map[string]any{
			"project_root": projRoot,
		})
		if !envNextSyn.Success {
			t.Fatalf("next_session failed for SYN-01: %s", envNextSyn.Message)
		}
		dataNextSyn := envNextSyn.Data
		if dataNextSyn["session_id"] != "SYN-01" {
			t.Fatalf("expected session SYN-01, got: %v", dataNextSyn["session_id"])
		}

		// 3.10: Save FAD synthesis with valid frontmatter and evidence grades
		fadContent := `---
id: FAD-001
title: EventStream Founding Architecture Document
synthesis_date: 2026-10-18
status: complete
---

# EventStream — Founding Architecture Document

## Executive Summary
This document establishes the founding technical architecture for EventStream.

## System Architecture
- Ingestion: Apache Kafka (per D-001) [Grade A: corroborated]
- Persistence: PostgreSQL 16 (per D-002) [Grade A: fresh]

## Invariant Non-Negotiables
1. All streaming data must flow through partitioned Kafka topics.
2. Direct client-to-database writes are prohibited.
`
		envSaveFAD := callTool("vivechak_save_session", map[string]any{
			"project_root": projRoot,
			"session_id":   "SYN-01",
			"content":      fadContent,
		})
		if !envSaveFAD.Success {
			t.Fatalf("save_session SYN-01 failed: %s", envSaveFAD.Message)
		}

		// Verify research/FAD.md and root mirror exist
		if _, err := os.Stat(filepath.Join(projRoot, "research", "FAD.md")); err != nil {
			t.Errorf("research/FAD.md missing: %v", err)
		}
		if _, err := os.Stat(filepath.Join(projRoot, "FOUNDING-ARCHITECTURE.md")); err != nil {
			t.Errorf("root FOUNDING-ARCHITECTURE.md missing: %v", err)
		}

		// 3.11: Run Phase 0 Gate
		envGate := callTool("vivechak_run_gate", map[string]any{
			"project_root": projRoot,
		})
		if !envGate.Success {
			t.Fatalf("run_gate failed: %s", envGate.Message)
		}
		dataGate := envGate.Data
		if dataGate["gate_passed"] != true {
			t.Errorf("expected gate_passed=true, got: %v (issues: %v)", dataGate["gate_passed"], dataGate)
		}

		// Verify research/PHASE-0-GATE.md was persisted
		if _, err := os.Stat(filepath.Join(projRoot, "research", "PHASE-0-GATE.md")); err != nil {
			t.Errorf("research/PHASE-0-GATE.md missing: %v", err)
		}

		// 3.12: Run CLI doctor on this project workspace
		docCmd := exec.CommandContext(ctx, binPath, "doctor", projRoot)
		docOut, docErr := docCmd.CombinedOutput()
		if docErr != nil {
			t.Fatalf("vck doctor failed on project workspace: %v\noutput: %s", docErr, string(docOut))
		}
		if !strings.Contains(string(docOut), "Workspace is healthy.") {
			t.Errorf("expected 'Workspace is healthy.', got: %s", string(docOut))
		}
	})

	// =========================================================================
	// TEST 4: CLI DEVELOPER ERGONOMICS (Version, Doctor, Setup, MCP-Config)
	// =========================================================================
	t.Run("CLIDeveloperErgonomics", func(t *testing.T) {
		// 4.1: vck version
		vCmd := exec.CommandContext(ctx, binPath, "version")
		vOut, err := vCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("vck version failed: %v", err)
		}
		if !strings.Contains(string(vOut), "vivechak 0.1.0") {
			t.Errorf("expected 'vivechak 0.1.0', got: %s", string(vOut))
		}

		// 4.2: vck mcp-config
		cfgCmd := exec.CommandContext(ctx, binPath, "mcp-config")
		cfgOut, err := cfgCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("vck mcp-config failed: %v", err)
		}
		var parsedJSON map[string]any
		if err := json.Unmarshal(cfgOut, &parsedJSON); err != nil {
			t.Fatalf("vck mcp-config did not output valid JSON: %v\nraw: %s", err, string(cfgOut))
		}
		if _, ok := parsedJSON["mcpServers"]; !ok {
			t.Errorf("expected 'mcpServers' key in mcp-config output")
		}

		// 4.3: vck setup with preset
		presetCmd := exec.CommandContext(ctx, binPath, "setup", "cursor", "--dry-run")
		presetOut, err := presetCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("vck setup cursor --dry-run failed: %v\noutput: %s", err, string(presetOut))
		}
		if !strings.Contains(string(presetOut), "cursor") {
			t.Errorf("expected setup output to reference cursor, got: %s", string(presetOut))
		}
	})
}
