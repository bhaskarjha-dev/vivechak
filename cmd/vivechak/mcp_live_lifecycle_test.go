package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMCPLive_Lifecycle executes the canonical multi-scope end-to-end lifecycle
// across Project, Decision, and Comparison scopes, verifying all 13 MCP tools over genuine
// subprocess stdio JSON-RPC protocol against the compiled Vivechak binary.
func TestMCPLive_Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live MCP subprocess lifecycle test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	binPath := resolveOrBuildTestBinary(t)

	// =========================================================================
	// 1. PROJECT SCOPE: FULL PIPELINE DAG, FAD SYNTHESIS, ROOT MIRROR & GATE
	// =========================================================================
	t.Run("01_ProjectScope_CompletePipeline", func(t *testing.T) {
		projectDir := t.TempDir()

		cs, cleanup := startLiveMCPServer(t, ctx, binPath, projectDir)
		t.Cleanup(cleanup)

		callTool := func(toolName string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      toolName,
				Arguments: args,
			})
			if err != nil {
				t.Fatalf("call to %s failed: %v", toolName, err)
			}
			return parseTestEnvelope(t, res)
		}

		// 1.1 vivechak_init (Project scope)
		initEnv := callTool("vivechak_init", map[string]any{
			"project_root": projectDir,
			"scope":        "project",
		})
		if !initEnv.Success {
			t.Fatalf("vivechak_init failed: %s", initEnv.Message)
		}

		// Verify research/.gitignore created with *.lock
		giData, err := os.ReadFile(filepath.Join(projectDir, core.ResearchDir, ".gitignore"))
		if err != nil || !strings.Contains(string(giData), "*.lock") {
			t.Errorf("expected research/.gitignore to contain *.lock, got err=%v content=%s", err, string(giData))
		}

		// Verify 6 templates copied, including SESSION.template.md
		tmpls, ok := initEnv.Data["templates_copied"].([]any)
		if !ok || len(tmpls) != 6 {
			t.Fatalf("expected 6 templates copied, got: %v", initEnv.Data["templates_copied"])
		}
		if _, err := os.Stat(filepath.Join(projectDir, core.TemplatesDir, "SESSION.template.md")); err != nil {
			t.Errorf("expected SESSION.template.md in templates/: %v", err)
		}

		// 1.2 vivechak_prepare_generator
		prepEnv := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": projectDir,
			"scope":        "project",
			"context":      "Architecting high-frequency financial distributed ledger with audit guarantees.",
		})
		if !prepEnv.Success {
			t.Fatalf("prepare_generator failed: %s", prepEnv.Message)
		}
		prompt := prepEnv.Data["prompt"].(string)
		if !strings.Contains(prompt, "financial distributed ledger") {
			t.Errorf("expected context injected into generator prompt")
		}

		// 1.3 vivechak_save_plan: 3-layer DAG with large guidance
		largeGuidance := strings.Repeat("Ensure rigorous verification and benchmark latency under load. ", 700)
		planContent := `# Research Pipeline

## Session DAG

### Session S-01 — Storage Engine
| Field | Value |
|---|---|
| **ID** | S-01 |
| **Layer** | 0 |
| **Door Type** | one-way |
| **Dependencies** | None |
| **Output File** | ` + "`S-01-storage.md`" + ` |

` + "```prompt\n" + largeGuidance + "\nEvaluate RocksDB vs LMDB.\n```" + `

### Session S-02 — Consensus Protocol
| Field | Value |
|---|---|
| **ID** | S-02 |
| **Layer** | 1 |
| **Door Type** | one-way |
| **Dependencies** | S-01 |
| **Output File** | ` + "`S-02-consensus.md`" + ` |

` + "```prompt\n" + `
Investigate consensus based on storage findings:
[UPSTREAM_FINDINGS]
` + "```" + `

### Session SYN-01 — Founding Architecture
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Layer** | 2 |
| **Door Type** | one-way |
| **Dependencies** | S-01, S-02 |
| **Output File** | ` + "`FAD.md`" + ` |

` + "```prompt\n" + largeGuidance + `
Synthesize founding architecture:
[ALL_SESSION_FINDINGS]
` + "```" + `
`
		planEnv := callTool("vivechak_save_plan", map[string]any{
			"project_root": projectDir,
			"content":      planContent,
		})
		if !planEnv.Success {
			t.Fatalf("save_plan failed: %s", planEnv.Message)
		}

		// 1.4 vivechak_status
		statEnv := callTool("vivechak_status", map[string]any{
			"project_root": projectDir,
		})
		if !statEnv.Success {
			t.Fatalf("status failed: %s", statEnv.Message)
		}
		if statEnv.Data["session_count"].(float64) != 0 {
			t.Errorf("expected 0 completed sessions, got %v", statEnv.Data["session_count"])
		}

		// 1.5 vivechak_next_session for S-01 (prompt truncation check when verbose=false)
		next1Env := callTool("vivechak_next_session", map[string]any{
			"project_root": projectDir,
			"verbose":      false,
		})
		if !next1Env.Success || next1Env.Data["session_id"] != "S-01" {
			t.Fatalf("expected next session S-01, got: %+v", next1Env)
		}
		p1 := next1Env.Data["prompt"].(string)
		if !strings.Contains(p1, "[truncated") {
			t.Errorf("expected large prompt to be truncated when verbose=false")
		}

		// 1.6 vivechak_save_session for S-01
		s1Content := `---
session_id: S-01
title: Storage Engine
date: 2026-10-01
status: complete
door_type: one-way
informs_decisions:
  - D-001
---
# Storage Engine Evaluation

## Key Findings
- RocksDB provides high write throughput with LSM tree design. Grade A (https://rocksdb.org)
- Value log separation allows scaling write bandwidth. Grade B (internal benchmark v8.1)

## Discovered Concerns
Write stall latency spikes occur during L0 compaction under heavy random write workloads exceeding 50,000 IOPS.

## Delta
| Prior | Status | Evidence | Impact |
| RocksDB is uniformly fast | Updated | Grade A (https://rocksdb.org) | High |

## Red-Team Stress Test
Compaction stall boundary analysis confirmed memory stalls under concurrent writer threads.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Status | Supports |
| 1 | https://rocksdb.org | Grade A | fresh | fetched | RocksDB throughput |
| 2 | internal benchmark v8.1 | Grade B | fresh | fetched | Compaction stalls |
`
		saveS1Env := callTool("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "S-01",
			"content":      s1Content,
		})
		if !saveS1Env.Success {
			t.Fatalf("save_session S-01 failed: %s", saveS1Env.Message)
		}

		// 1.7 vivechak_record_decision for D-001 (ADR auto-drafting from S-01)
		adrContent := `---
decision_id: D-001
title: RocksDB as Primary Engine
date: 2026-10-01
status: accepted
door_type: one-way
review_trigger: Dataset exceeds 100 TB or p99 write latency > 50ms
informed_by_sessions:
  - S-01
---
# D-001: RocksDB Primary Storage Engine

## Context
High-frequency distributed ledger requires persistent KV store with high random write throughput.

## Decision
We select RocksDB for primary state storage. Grade A (https://rocksdb.org)

## Alternatives Considered
- LMDB: rejected due to write concurrency bottleneck. Grade A (https://symas.com/lmdb)
- BadgerDB: rejected due to GC compaction pause anomalies under sustained 100K ops/sec write load.

## Consequences
- Positive: battle-tested LSM engine with flexible compaction filters.
- Negative: memory overhead for block cache and memtables requires active tuning.

## Red-Team Stress Test
Adversarial disk saturation probe verified write stall recovery within acceptable SLA bounds.
`
		recEnv := callTool("vivechak_record_decision", map[string]any{
			"project_root": projectDir,
			"decision_id":  "D-001",
			"content":      adrContent,
		})
		if !recEnv.Success {
			t.Fatalf("record_decision D-001 failed: %s", recEnv.Message)
		}

		// 1.8 Save S-02
		s2Content := `---
session_id: S-02
title: Consensus Protocol
date: 2026-10-01
status: complete
door_type: two-way
---
## Key Findings
- Raft protocol guarantees safety across network partitions. Grade A (https://raft.github.io)
- Election timeouts between 150-300ms avoid split votes. Grade A (https://raft.github.io)

## Discovered Concerns
Pre-vote extension must be enabled to prevent disrupted elections during cluster churn.

## Delta
| Prior | Status | Evidence | Impact |
| Raft is complex | Updated | Grade A (https://raft.github.io) | Low |

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Status | Supports |
| 1 | https://raft.github.io | Grade A | fresh | fetched | Raft consensus |
`
		_ = callTool("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "S-02",
			"content":      s2Content,
		})

		// 1.9 vivechak_amend_session: amend S-01 now that S-02 is completed (triggers W-STALE-DOWNSTREAM)
		amendEnv := callTool("vivechak_amend_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "S-01",
			"content":      s1Content + "\n\n// Amendment: tune write_buffer_size to 128MB.\n",
			"reason":       "Tuning buffer parameters for write stall mitigation.",
		})
		if !amendEnv.Success {
			t.Fatalf("amend_session S-01 failed: %s", amendEnv.Message)
		}
		hasDownstreamWarn := false
		for _, w := range amendEnv.Warnings {
			if strings.Contains(w, "W-STALE-DOWNSTREAM") && strings.Contains(w, "S-02") {
				hasDownstreamWarn = true
				break
			}
		}
		if !hasDownstreamWarn {
			t.Errorf("expected W-STALE-DOWNSTREAM warning for S-02, got: %v", amendEnv.Warnings)
		}

		// 1.10 vivechak_visualize
		vizEnv := callTool("vivechak_visualize", map[string]any{
			"project_root": projectDir,
		})
		if !vizEnv.Success {
			t.Fatalf("visualize failed: %s", vizEnv.Message)
		}
		mermaid, _ := vizEnv.Data["mermaid"].(string)
		if !strings.Contains(mermaid, "graph TD") || !strings.Contains(mermaid, "S-01") {
			t.Errorf("visualize missing expected Mermaid DAG content: %s", mermaid)
		}

		// 1.11 vivechak_replan: add session S-03 (dependencies as comma-separated string)
		replanEnv := callTool("vivechak_replan", map[string]any{
			"project_root": projectDir,
			"operation":    "add_session",
			"session_id":   "S-03",
			"topic":        "Disaster Recovery and Snapshots",
			"dependencies": "S-01",
			"door_type":    "two-way",
			"reason":       "Add backup architecture exploration before synthesis.",
		})
		if !replanEnv.Success {
			t.Fatalf("replan add_session failed: %s", replanEnv.Message)
		}

		s3Content := `---
session_id: S-03
title: Disaster Recovery
date: 2026-10-01
status: complete
door_type: two-way
---
## Key Findings
- S3 object storage enables zero-loss continuous point-in-time recovery. Grade A (https://aws.amazon.com/s3)
- Snapshot intervals of 15 minutes bound WAL playback time. Grade B (internal DR spike)

## Discovered Concerns
S3 multipart upload aborts require lifecycle policies to prevent orphaned part storage costs.

## Delta
| Prior | Status | Evidence | Impact |
| DR is slow | Confirmed | Grade A (https://aws.amazon.com/s3) | Low |

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Status | Supports |
| 1 | https://aws.amazon.com/s3 | Grade A | fresh | fetched | S3 storage |
`
		_ = callTool("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "S-03",
			"content":      s3Content,
		})

		// 1.12 Save SYN-01 as FAD.md and verify Root Mirror Copy
		fadContent := `---
session_id: SYN-01
title: Founding Architecture Document
date: 2026-10-01
status: complete
---
# Founding Architecture Document: High-Frequency Ledger

## Executive Summary
This architecture specifies the storage and consensus foundation for high-frequency financial transactions.

## Storage Layer
RocksDB state store with write buffer size tuned to 128 MB. Grade A (https://rocksdb.org)

## Consensus Layer
Raft majority quorum with pre-vote enabled. Grade A (https://raft.github.io)

## Disaster Recovery
Continuous WAL archiving to S3 object storage. Grade A (https://aws.amazon.com/s3)
`
		saveFADEnv := callTool("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "SYN-01",
			"content":      fadContent,
		})
		if !saveFADEnv.Success {
			t.Fatalf("save_session SYN-01 (FAD) failed: %s", saveFADEnv.Message)
		}

		// Verify research/FAD.md was created
		fadPath := filepath.Join(projectDir, core.FADFile)
		if _, err := os.Stat(fadPath); err != nil {
			t.Errorf("expected research/FAD.md to exist: %v", err)
		}

		// Sign off One-Way Door D-001 so Track B Gate can pass
		ws, err := store.OpenWorkspace(projectDir)
		if err != nil {
			t.Fatalf("open workspace: %v", err)
		}
		if _, err := core.SignoffDecision(ws, "D-001", "Lead Architect"); err != nil {
			ws.Close()
			t.Fatalf("signoff D-001: %v", err)
		}
		ws.Close()

		// 1.13 vivechak_run_gate
		gateEnv := callTool("vivechak_run_gate", map[string]any{
			"project_root": projectDir,
		})
		if !gateEnv.Success {
			t.Fatalf("run_gate failed: %s", gateEnv.Message)
		}
		if gateEnv.Data["gate_passed"] != true {
			t.Fatalf("expected gate to pass, got: %+v\nwarnings: %v", gateEnv.Data, gateEnv.Warnings)
		}
		if gateEnv.Data["gate_status"] != "PASS" {
			t.Errorf("expected gate status 'PASS', got: %v", gateEnv.Data["gate_status"])
		}

		// Verify PHASE-0-GATE.md was auto-persisted
		gateArtifact := filepath.Join(projectDir, core.ResearchDir, "PHASE-0-GATE.md")
		if _, err := os.Stat(gateArtifact); err != nil {
			t.Errorf("expected PHASE-0-GATE.md on disk: %v", err)
		}

		// 1.14 Run CLI doctor on valid workspace
		docCmd := exec.CommandContext(ctx, binPath, "doctor", projectDir)
		docOut, err := docCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("vck doctor failed: %v\noutput: %s", err, string(docOut))
		}
		if !strings.Contains(string(docOut), "Workspace is healthy") {
			t.Errorf("expected 'Workspace is healthy', got: %s", string(docOut))
		}
	})

	// =========================================================================
	// 2. DECISION SCOPE: STANDALONE ADR, CHALLENGE, DEDUPLICATION & GATE
	// =========================================================================
	t.Run("02_DecisionScope_CompleteADR", func(t *testing.T) {
		decDir := t.TempDir()

		cs, cleanup := startLiveMCPServer(t, ctx, binPath, decDir)
		t.Cleanup(cleanup)

		callTool := func(toolName string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      toolName,
				Arguments: args,
			})
			if err != nil {
				t.Fatalf("call to %s failed: %v", toolName, err)
			}
			return parseTestEnvelope(t, res)
		}

		// 2.1 Init Decision Workspace
		initEnv := callTool("vivechak_init", map[string]any{
			"project_root": decDir,
			"scope":        "decision",
		})
		if !initEnv.Success {
			t.Fatalf("init decision failed: %s", initEnv.Message)
		}

		// Verify exactly 2 templates copied
		tplEntries, _ := os.ReadDir(filepath.Join(decDir, core.TemplatesDir))
		if len(tplEntries) != 2 {
			t.Errorf("expected exactly 2 templates for decision scope, got %d: %v", len(tplEntries), tplEntries)
		}

		// 2.2 Save Decision Plan
		planMD := `# Decision Plan: D-001 Consensus Protocol
| ID | Topic | Dependencies | Output File |
|---|---|---|---|
| D-001-S01 | Raft Protocol Evaluation | none | sessions/D-001-S01.md |
`
		planEnv := callTool("vivechak_save_plan", map[string]any{
			"project_root": decDir,
			"scope":        "decision",
			"decision_id":  "D-001",
			"slug":         "consensus",
			"content":      planMD,
		})
		if !planEnv.Success {
			t.Fatalf("save_plan decision failed: %s", planEnv.Message)
		}

		// 2.3 Save Session D-001-S01
		sessMD := `---
session_id: D-001-S01
title: Raft Consensus Evaluation
date: 2026-10-18
door_type: one-way
status: complete
informs_decisions:
  - D-001
---
## Key Findings
- Raft guarantees deterministic log consistency under majority quorum. Grade A (https://raft.github.io)
- Etcd raft implementation is heavily tested in production. Grade A (https://etcd.io)

## Discovered Concerns
Pre-vote extension is strictly mandatory in lossy networks to avoid disruptive elections caused by partitioned nodes.

## Delta
| Prior | Status | Evidence | Impact |
| Raft is easy | Updated | Grade A (https://raft.github.io) | Medium |

## Red-Team Stress Test
Split-brain partition analysis verified safety leases and quorum requirements.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Status | Supports |
| 1 | https://raft.github.io | Grade A | fresh | fetched | Raft safety |
| 2 | https://etcd.io | Grade A | fresh | fetched | Etcd raft |
`
		saveSessEnv := callTool("vivechak_save_session", map[string]any{
			"project_root": decDir,
			"session_id":   "D-001-S01",
			"content":      sessMD,
		})
		if !saveSessEnv.Success {
			t.Fatalf("save_session D-001-S01 failed: %s", saveSessEnv.Message)
		}

		// 2.4 Adversarial Challenge Tool Call
		challEnv := callTool("vivechak_challenge", map[string]any{
			"project_root": decDir,
			"session_id":   "D-001-S01",
			"mode":         "red_team",
		})
		if !challEnv.Success {
			t.Fatalf("vivechak_challenge failed: %s", challEnv.Message)
		}

		// 2.5 Record Decision D-001
		adrMD := `---
decision_id: D-001
title: Raft Consensus Protocol
date: 2026-10-18
status: accepted
door_type: one-way
review_trigger: Cluster exceeds 9 voting members or cross-region latency p99 > 80ms
informed_by_sessions:
  - D-001-S01
---
# D-001: Raft Consensus Protocol

## Context
Distributed cluster consensus requires strong leader election and log replication guarantees.

## Decision
We adopt hashicorp/raft for leader-based cluster consensus. Grade A (https://raft.github.io)

## Alternatives Considered
- Multi-Paxos: rejected due to high operational complexity and subtle concurrency failure modes.
- Primary-Backup: rejected due to lack of automatic failover and risk of split-brain data loss.

## Consequences
- Positive: deterministic leader election and battle-tested log replication.
- Negative: writes must pass through the active leader bottleneck.

## Red-Team Stress Test
Partition tolerance tested across asymmetric network partitions.
`
		recEnv := callTool("vivechak_record_decision", map[string]any{
			"project_root": decDir,
			"decision_id":  "D-001",
			"content":      adrMD,
		})
		if !recEnv.Success {
			t.Fatalf("record_decision failed: %s", recEnv.Message)
		}

		// 2.6 Validate Workspace (deduplication check between D-001.md and DECISIONS.md)
		valEnv := callTool("vivechak_validate", map[string]any{
			"project_root": decDir,
		})
		if !valEnv.Success {
			t.Fatalf("validate failed: %s", valEnv.Message)
		}
		decStats := valEnv.Data["decisions"].(map[string]any)
		if int(decStats["valid"].(float64)) != 1 {
			t.Errorf("expected exactly 1 deduplicated valid decision, got: %v", decStats["valid"])
		}

		// Sign off One-Way Door D-001 so Track B Gate can pass
		wsDec, err := store.OpenWorkspace(decDir)
		if err != nil {
			t.Fatalf("open workspace: %v", err)
		}
		if _, err := core.SignoffDecision(wsDec, "D-001", "Lead Architect"); err != nil {
			wsDec.Close()
			t.Fatalf("signoff D-001: %v", err)
		}
		wsDec.Close()

		// 2.7 Run Gate (Track B)
		gateEnv := callTool("vivechak_run_gate", map[string]any{
			"project_root": decDir,
			"track":        "track_b",
		})
		if !gateEnv.Success || gateEnv.Data["gate_passed"] != true {
			t.Fatalf("expected decision gate to pass, got: %+v", gateEnv.Data)
		}
	})

	// =========================================================================
	// 3. COMPARISON SCOPE: WEP MATRIX, OPTIONS, CRITERIA & AUTO-RECOVERY
	// =========================================================================
	t.Run("03_ComparisonScope_CompleteWEP", func(t *testing.T) {
		cmpDir := t.TempDir()

		cs, cleanup := startLiveMCPServer(t, ctx, binPath, cmpDir)
		t.Cleanup(cleanup)

		callTool := func(toolName string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      toolName,
				Arguments: args,
			})
			if err != nil {
				t.Fatalf("call to %s failed: %v", toolName, err)
			}
			return parseTestEnvelope(t, res)
		}

		// 3.1 Init Comparison Workspace
		initEnv := callTool("vivechak_init", map[string]any{
			"project_root": cmpDir,
			"scope":        "comparison",
		})
		if !initEnv.Success {
			t.Fatalf("init comparison failed: %s", initEnv.Message)
		}

		// Verify only COMPARISON-SESSION.template.md was copied
		tplEntries, _ := os.ReadDir(filepath.Join(cmpDir, core.TemplatesDir))
		if len(tplEntries) != 1 || tplEntries[0].Name() != "COMPARISON-SESSION.template.md" {
			t.Errorf("expected only COMPARISON-SESSION.template.md, got: %v", tplEntries)
		}

		// 3.2 Prepare Generator with comparison prompt
		prepEnv := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": cmpDir,
			"scope":        "comparison",
			"context":      "Compare Pino vs Winston for high-throughput Node.js microservices.",
		})
		if !prepEnv.Success {
			t.Fatalf("prepare_generator comparison failed: %s", prepEnv.Message)
		}

		// 3.3 Save Comparison Plan (Auto ID extraction C-001)
		planMD := `# Bounded Comparison: Pino vs Winston
| Option | Focus | Status |
|---|---|---|
| Pino | Extreme performance JSON logger | candidate |
| Winston | Flexible multi-transport logger | candidate |
`
		planEnv := callTool("vivechak_save_plan", map[string]any{
			"project_root": cmpDir,
			"scope":        "comparison",
			"decision_id":  "C-001",
			"slug":         "pino-winston",
			"content":      planMD,
		})
		if !planEnv.Success {
			t.Fatalf("save_plan comparison failed: %s", planEnv.Message)
		}

		// 3.4 Save Comparison Session
		sessMD := `---
session_id: C-001
title: Pino vs Winston Comparison
date: 2026-10-18
status: complete
---
## Prior
Pino is faster because it delays formatting to worker threads. Winston is more pluggable.

## Key Findings
- Pino achieves 5x higher throughput in synthetic JSON benchmarks. Grade A (https://getpino.io)
- Winston supports arbitrary custom transports with built-in daily rotation. Grade A (https://github.com/winstonjs/winston)

## Discovered Concerns
Pino asynchronous logging may lose trailing log lines on unhandled process crash unless pino.final is bound.

## Delta
| Prior | Status | Evidence | Impact |
| Pino is faster | Confirmed | Grade A (https://getpino.io) | High |

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Status | Supports |
| 1 | https://getpino.io | Grade A | fresh | fetched | Pino throughput |
| 2 | https://github.com/winstonjs/winston | Grade A | fresh | fetched | Winston transports |
`
		saveSessEnv := callTool("vivechak_save_session", map[string]any{
			"project_root": cmpDir,
			"session_id":   "C-001",
			"content":      sessMD,
		})
		if !saveSessEnv.Success {
			t.Fatalf("save_session comparison failed: %s", saveSessEnv.Message)
		}

		// 3.5 Validate Comparison Workspace
		valEnv := callTool("vivechak_validate", map[string]any{
			"project_root": cmpDir,
		})
		if !valEnv.Success {
			t.Fatalf("validate comparison failed: %s", valEnv.Message)
		}
	})

	// =========================================================================
	// 4. PROTOCOL DISCOVERY, ERROR ENVELOPES & CLI COMMANDS
	// =========================================================================
	t.Run("04_ProtocolDiscovery_And_CLI_Commands", func(t *testing.T) {
		tempDir := t.TempDir()
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, tempDir)
		t.Cleanup(cleanup)

		// 4.1 Verify exactly 13 tools registered
		var toolNames []string
		for tool, err := range cs.Tools(ctx, nil) {
			if err != nil {
				t.Fatalf("error listing tools: %v", err)
			}
			toolNames = append(toolNames, tool.Name)
		}
		if len(toolNames) != 9 {
			t.Fatalf("expected 9 tools from live server, got %d: %v", len(toolNames), toolNames)
		}

		// 4.2 Verify unknown tool returns protocol error
		_, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_nonexistent_tool",
			Arguments: map[string]any{},
		})
		if err == nil {
			t.Error("expected unknown tool call to return error, got nil")
		}

		// 4.3 Subprocess CLI commands verification
		// vck version
		vCmd := exec.CommandContext(ctx, binPath, "version")
		vOut, err := vCmd.CombinedOutput()
		if err != nil || !strings.Contains(string(vOut), "vivechak") {
			t.Errorf("vck version failed: %v\noutput: %s", err, string(vOut))
		}

		// vck mcp-config
		cfgCmd := exec.CommandContext(ctx, binPath, "mcp-config")
		cfgOut, err := cfgCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("vck mcp-config failed: %v\noutput: %s", err, string(cfgOut))
		}
		var parsedConfig map[string]any
		if err := json.Unmarshal(cfgOut, &parsedConfig); err != nil {
			t.Fatalf("invalid JSON from mcp-config: %v", err)
		}
		if _, ok := parsedConfig["mcpServers"]; !ok {
			t.Errorf("missing mcpServers key in mcp-config output: %s", string(cfgOut))
		}
	})

	// =========================================================================
	// 5. LIVE SUBPROCESS: MULTI-DECISION REGISTRY, DETAILS FRONTMATTER & DOCTOR
	// =========================================================================
	t.Run("05_ProjectScope_MultiDecisionRegistry_LiveSubprocess", func(t *testing.T) {
		projectDir := t.TempDir()

		cs, cleanup := startLiveMCPServer(t, ctx, binPath, projectDir)
		t.Cleanup(cleanup)

		callTool := func(toolName string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      toolName,
				Arguments: args,
			})
			if err != nil {
				t.Fatalf("call to %s failed: %v", toolName, err)
			}
			return parseTestEnvelope(t, res)
		}

		// 5.1 Initialize workspace
		envInit := callTool("vivechak_init", map[string]any{
			"project_root": projectDir,
			"scope":        "project",
		})
		if !envInit.Success {
			t.Fatalf("live init failed: %s", envInit.Message)
		}

		// 5.2 Prepare generator and verify Tier 3 scratch staging guidance
		envPrep := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": projectDir,
			"scope":        "project",
			"context":      "Distributed graph analytics engine",
		})
		if !envPrep.Success {
			t.Fatalf("live prepare_generator failed: %s", envPrep.Message)
		}
		if !strings.Contains(envPrep.NextStep, "assemble sections in internal scratch") {
			t.Errorf("expected scratch guidance in next_step: %s", envPrep.NextStep)
		}

		// 5.3 Save plan with multi-decision registry
		decisionsRegistry := `# Architectural Decision Records
Overview of system architectural choices.

| Decision | Title | Door Type | Status | Date |
|---|---|---|---|---|
| D-01 | Primary Graph Engine | one-way | proposed | 2026-10-21 |
| D-02 | Ingest Buffer | two-way | proposed | 2026-10-21 |

<!-- DECISION: D-01 -->
<details>
<summary><b>D-01: Primary Graph Engine</b></summary>

---
id: D-01
title: Primary Graph Engine
door_type: one-way
status: proposed
date: 2026-10-21
---

### Context
Evaluating Cypher vs Gremlin based engines.
</details>
<!-- /DECISION: D-01 -->

---

<!-- DECISION: D-02 -->
<details>
<summary><b>D-02: Ingest Buffer</b></summary>

` + "```yaml" + `
id: D-02
title: Ingest Buffer
door_type: two-way
status: proposed
date: 2026-10-21
` + "```" + `

### Context
Evaluating NATS JetStream vs Apache Kafka.
</details>
<!-- /DECISION: D-02 -->
`
		pipelineContent := `# Research Pipeline: Graph Analytics
**Complexity Score:** 18 / 24 → **Tier 3 (17–30 Sessions)**

## Sessions
#### T0-01: Protocol Benchmark
| Field | Value |
|---|---|
| ID | T0-01 |
| Layer | 0 |
| Door Type | Two-Way |
| Dependencies | None |

` + "```prompt" + `
Investigate gRPC vs Arrow Flight for graph query results.
` + "```" + `
`
		envSave := callTool("vivechak_save_plan", map[string]any{
			"project_root":      projectDir,
			"scope":             "project",
			"content":           pipelineContent,
			"decisions_content": decisionsRegistry,
		})
		if !envSave.Success {
			t.Fatalf("live save_plan failed: %s", envSave.Message)
		}
		if len(envSave.Warnings) != 0 {
			t.Errorf("expected 0 warnings on clean plan save, got: %v", envSave.Warnings)
		}

		// 5.4 Validate registry directly via artifact_type="decisions"
		envValDec := callTool("vivechak_validate", map[string]any{
			"project_root":  projectDir,
			"artifact_type": "decisions",
			"content":       decisionsRegistry,
		})
		if !envValDec.Success {
			t.Fatalf("live validate decisions failed: %s", envValDec.Message)
		}
		valData := envValDec.Data
		if valData["status"] != "valid" {
			t.Errorf("expected status 'valid', got: %v (warnings: %v)", valData["status"], envValDec.Warnings)
		}

		// 5.5 Validate entire workspace via workspace discovery mode
		envValWs := callTool("vivechak_validate", map[string]any{
			"project_root": projectDir,
		})
		if !envValWs.Success {
			t.Fatalf("live workspace validate failed: %s", envValWs.Message)
		}
		wsData := envValWs.Data
		decMap, ok := wsData["decisions"].(map[string]any)
		if !ok {
			t.Fatalf("expected decisions map in workspace validate data: %v", wsData)
		}
		if decMap["valid"] != float64(2) && decMap["valid"] != 2 {
			t.Errorf("expected 2 valid decisions discovered in workspace DECISIONS.md, got: %v (%T)", decMap["valid"], decMap["valid"])
		}
		if decMap["errors"] != float64(0) && decMap["errors"] != 0 {
			t.Errorf("expected 0 errors in workspace DECISIONS.md, got: %v (%T)", decMap["errors"], decMap["errors"])
		}

		// 5.6 Execute vck doctor child process and verify healthy workspace
		docCmd := exec.CommandContext(ctx, binPath, "doctor", projectDir)
		docOut, err := docCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("vck doctor failed: %v\noutput: %s", err, string(docOut))
		}
		if !strings.Contains(string(docOut), "Workspace is healthy") {
			t.Errorf("expected 'Workspace is healthy' in doctor output, got:\n%s", string(docOut))
		}
	})
}
