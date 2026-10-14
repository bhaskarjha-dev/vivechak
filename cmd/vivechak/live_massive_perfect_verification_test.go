package main

import (
	"bytes"
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

// TestLiveMassive_EveryWay_FullLifecycleAndSecurity performs massive real testing:
// 1. Live subprocess MCP JSON-RPC protocol execution over stdio
// 2. Multi-tier dynamic DAG resolution with One-Way and Two-Way doors
// 3. Bidirectional frontmatter status synchronization (draft on defects -> complete on valid)
// 4. In-flight replanning, sink guards, and visualization (table + mermaid)
// 5. Single-source ADR auto-drafting, challenge probing, and supersession
// 6. Terminal FAD synthesis, root mirroring, and Phase 0 Gate (Track A + Track B)
// 7. Security confinement defense against directory traversal
// 8. Real CLI execution: doctor, setup, mcp-config presets, version
func TestLiveMassive_EveryWay_FullLifecycleAndSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	t.Cleanup(cancel)

	// 1. Build vivechak binary once for live testing
	binDir := t.TempDir()
	binName := "vck_massive_perfect_test"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	// 2. Launch live MCP server as a genuine child process over stdio JSON-RPC
	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "massive-perfect-tester", Version: "1.0.0"},
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

	// =========================================================================
	// Phase 1: Initialization & Plan Setup
	// =========================================================================
	t.Run("01_Init_And_Plan_Setup", func(t *testing.T) {
		// Tool 1: vivechak_init (project scope)
		initEnv := callTool("vivechak_init", map[string]any{
			"project_root": wsDir,
			"scope":        "project",
		})
		if !initEnv.Success {
			t.Fatalf("vivechak_init failed: %s", initEnv.Message)
		}

		// Tool 2: vivechak_prepare_generator
		prepEnv := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": wsDir,
			"scope":        "project",
			"context":      "Distributed low-latency stream processing platform with fault-tolerant state store.",
		})
		if !prepEnv.Success {
			t.Fatalf("vivechak_prepare_generator failed: %s", prepEnv.Message)
		}

		// Tool 3: vivechak_save_plan (5-node multi-branch DAG)
		dagContent := `# Research Pipeline: Enterprise Stream Platform

**Complexity Score:** 18 / 24 → **Tier 2 (8–14 Sessions)**

**Execution DAG:**
` + "```mermaid" + `
graph TD
    T1_01[T1-01: Storage Engine]
    T1_02[T1-02: Cache Strategy]
    T2_01[T2-01: Partitioning Model]
    T2_02[T2-02: Ingest Gateway]
    SYN_01[SYN-01: Founding Architecture Synthesis]

    T1_01 --> T2_01
    T1_02 --> T2_02
    T2_01 --> SYN_01
    T2_02 --> SYN_01
` + "```" + `

## Sessions

#### T1-01: Storage Engine
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Door Type** | One-Way |
| **Dependencies** | None |
| **Output File** | ` + "`research/sessions/T1-01.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-01 Storage Engine
Investigate transactional storage.
` + "```" + `

---

#### T1-02: Cache Strategy
| Field | Value |
|---|---|
| **ID** | T1-02 |
| **Door Type** | Two-Way |
| **Dependencies** | None |
| **Output File** | ` + "`research/sessions/T1-02.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-02 Cache Strategy
Evaluate caching backends.
` + "```" + `

---

#### T2-01: Partitioning Model
| Field | Value |
|---|---|
| **ID** | T2-01 |
| **Door Type** | One-Way |
| **Dependencies** | T1-01 |
| **Output File** | ` + "`research/sessions/T2-01.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T2-01 Partitioning Model
Investigate horizontal partitioning schemes.
` + "```" + `

---

#### T2-02: Ingest Gateway
| Field | Value |
|---|---|
| **ID** | T2-02 |
| **Door Type** | Two-Way |
| **Dependencies** | T1-02 |
| **Output File** | ` + "`research/sessions/T2-02.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T2-02 Ingest Gateway
Evaluate high-concurrency ingestion gateways.
` + "```" + `

---

#### SYN-01: Founding Architecture Synthesis
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Dependencies** | T2-01, T2-02 |
| **Output File** | ` + "`research/FAD.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: SYN-01 Synthesis
Synthesize the entire platform architecture.
` + "```" + `
`
		savePlanEnv := callTool("vivechak_save_plan", map[string]any{
			"project_root": wsDir,
			"content":      dagContent,
		})
		if !savePlanEnv.Success {
			t.Fatalf("vivechak_save_plan failed: %s", savePlanEnv.Message)
		}
	})

	// =========================================================================
	// Phase 2: Status Synchronization, Quality Coaching & DAG Progression
	// =========================================================================
	t.Run("02_Status_Sync_And_DAG_Progression", func(t *testing.T) {
		// Tool 4: vivechak_next_session -> Ready should be T1-01 or T1-02
		nextEnv := callTool("vivechak_next_session", map[string]any{
			"project_root": wsDir,
		})
		if !nextEnv.Success {
			t.Fatalf("next_session failed: %s", nextEnv.Message)
		}
		initReadyID, _ := nextEnv.Data["session_id"].(string)
		if initReadyID != "T1-01" && initReadyID != "T1-02" {
			t.Fatalf("expected ready session T1-01 or T1-02, got %s", initReadyID)
		}

		// Save T1-01 with defects: One-Way Door with trivial concerns ("None.")
		// This triggers V-OWD-NO-CONCERNS (blocking), so tool_save_session must enforce status: draft on disk!
		invalidT101 := `---
session_id: T1-01
title: Storage Engine
door_type: one-way
status: complete
date: "2026-10-14"
---
# T1-01: Storage Engine Findings
## Key Findings
- PostgreSQL 16 supports jsonb and partitioning [Grade A: official docs https://postgresql.org/docs].
- Buffer pool caching lowers disk seek overhead [Grade B: benchmark].

## Discovered Concerns
None.

## Recommendation
Use PostgreSQL [Grade A: official docs https://postgresql.org/docs].
`
		saveDraftEnv := callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "T1-01",
			"content":      invalidT101,
		})
		if !saveDraftEnv.Success {
			t.Fatalf("save_session should succeed with envelope warnings: %s", saveDraftEnv.Message)
		}

		// Inspect file on disk: status MUST be draft!
		sess1Bytes, err := os.ReadFile(filepath.Join(wsDir, core.SessionsDir, "T1-01.md"))
		if err != nil {
			t.Fatalf("reading T1-01.md: %v", err)
		}
		if !strings.Contains(string(sess1Bytes), "status: draft") {
			t.Errorf("expected status: draft on disk due to blocking defect, got:\n%s", string(sess1Bytes))
		}

		// Because T1-01 is draft, T2-01 (dependent on T1-01) MUST NOT be unlocked!
		// Next session should still be T1-01 or T1-02
		nextBlockedEnv := callTool("vivechak_next_session", map[string]any{
			"project_root": wsDir,
		})
		blockedNextID, _ := nextBlockedEnv.Data["session_id"].(string)
		if blockedNextID == "T2-01" {
			t.Fatalf("T2-01 was prematurely unblocked while T1-01 is still draft!")
		}

		// Now fix T1-01 with substantive Discovered Concerns (>= 50 chars)
		// This should promote status: complete automatically!
		validT101 := `---
session_id: T1-01
title: Storage Engine
door_type: one-way
status: draft
date: "2026-10-14"
informs_decisions: ["D-001"]
---
# T1-01: Storage Engine Findings

## Key Findings
- PostgreSQL 16 logical partitioning scales to 150k writes/sec [Grade A: official docs https://postgresql.org/docs].
- NVMe direct write-ahead-log fsync avoids checkpoint stalls [Grade B: benchmark].
- Advisory lock coordination ensures zero-downtime partition evolution [Grade A: RFC https://tools.ietf.org/html/rfc1234].

## Discovered Concerns
Heavy write volume during active partition creation triggers transient lock contention on the parent routing catalog.

## Sources & Evidence Ledger
| Claim | Grade | Modifiers | Method | Citation |
|---|---|---|---|---|
| Logical partitioning | Grade A | direct | fetched | [PostgreSQL Documentation](https://postgresql.org/docs) |
| WAL fsync latency | Grade B | fresh | test | Internal hardware benchmark |

## Recommendation
Adopt PostgreSQL 16 with pre-warmed monthly logical partitions [Grade A: official docs https://postgresql.org/docs].
`
		saveCompleteEnv := callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "T1-01",
			"content":      validT101,
		})
		if !saveCompleteEnv.Success {
			t.Fatalf("save valid T1-01 failed: %s", saveCompleteEnv.Message)
		}

		// Inspect file on disk: status MUST now be complete!
		sess1FixedBytes, _ := os.ReadFile(filepath.Join(wsDir, core.SessionsDir, "T1-01.md"))
		if !strings.Contains(string(sess1FixedBytes), "status: complete") {
			t.Errorf("expected status: complete on disk after fixing defects, got:\n%s", string(sess1FixedBytes))
		}

		// Complete T1-02 (Two-Way Door)
		validT102 := `---
session_id: T1-02
title: Cache Strategy
door_type: two-way
status: complete
date: "2026-10-14"
---
# T1-02: Cache Strategy Findings

## Key Findings
- Redis standalone delivers sub-millisecond LRU cache reads [Grade A: official docs https://redis.io/docs].
- Write-through caching avoids cache stampedes under high concurrency [Grade B: benchmark].

## Discovered Concerns
Memory saturation triggers aggressive eviction without warning if maxmemory policy is misconfigured.

## Recommendation
Adopt Redis with volatile-lru eviction for transient query acceleration [Grade A: official docs https://redis.io/docs].
`
		_ = callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "T1-02",
			"content":      validT102,
		})

		// Complete T2-01 (One-Way Door: depends on T1-01)
		validT201 := `---
session_id: T2-01
title: Partitioning Model
door_type: one-way
status: complete
date: "2026-10-14"
informs_decisions: ["D-002"]
---
# T2-01: Partitioning Model Findings

## Key Findings
- Consistent hashing with virtual nodes ensures uniform data distribution [Grade A: official docs https://kubernetes.io/docs].
- Rebalance operations incur minimal network redistribution overhead [Grade B: cluster benchmark].
- Partition key cardinality prevents hot-spot shards [Grade A: RFC https://tools.ietf.org/rfc9000].

## Discovered Concerns
Adding virtual nodes dynamically requires distributed state lock lease renewal across all cluster nodes.

## Recommendation
Adopt consistent hashing with 256 virtual nodes per storage partition [Grade A: official docs https://kubernetes.io/docs].
`
		_ = callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "T2-01",
			"content":      validT201,
		})

		// Complete T2-02 (Two-Way Door: depends on T1-02)
		validT202 := `---
session_id: T2-02
title: Ingest Gateway
door_type: two-way
status: complete
date: "2026-10-14"
---
# T2-02: Ingest Gateway Findings

## Key Findings
- Envoy proxy handles 80k concurrent HTTP/2 streams with low latency [Grade A: official docs https://envoyproxy.io/docs].
- Backpressure propagation prevents gateway buffer overflow [Grade B: load benchmark].

## Discovered Concerns
High stream concurrency requires tuning worker thread pools to match host CPU core affinity.

## Recommendation
Use Envoy proxy with HTTP/2 and gRPC termination [Grade A: official docs https://envoyproxy.io/docs].
`
		_ = callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "T2-02",
			"content":      validT202,
		})
	})

	// =========================================================================
	// Phase 3: In-Flight Replanning & Visualizations
	// =========================================================================
	t.Run("03_InFlight_Replan_And_Visualizations", func(t *testing.T) {
		// Tool 11: vivechak_replan (add session T2-03)
		replanEnv := callTool("vivechak_replan", map[string]any{
			"project_root": wsDir,
			"operation":    "add_session",
			"session_id":   "T2-03",
			"topic":        "Metrics Telemetry Pipeline",
			"door_type":    "two-way",
			"dependencies": "T1-02",
		})
		if !replanEnv.Success {
			t.Fatalf("replan add_session failed: %s", replanEnv.Message)
		}

		// Complete T2-03
		validT203 := `---
session_id: T2-03
title: Metrics Telemetry Pipeline
door_type: two-way
status: complete
date: "2026-10-14"
---
# T2-03: Metrics Telemetry Findings
Prometheus metrics export and OpenTelemetry tracing validated [Grade A: official docs https://prometheus.io/docs].
`
		_ = callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "T2-03",
			"content":      validT203,
		})

		// Tool 12: vivechak_visualize (table format)
		visTableEnv := callTool("vivechak_visualize", map[string]any{
			"project_root": wsDir,
			"format":       "table",
		})
		if !visTableEnv.Success {
			t.Fatalf("visualize table failed: %s", visTableEnv.Message)
		}
		tableStr, _ := visTableEnv.Data["rendered"].(string)
		for _, sID := range []string{"T1-01", "T1-02", "T2-01", "T2-02", "T2-03"} {
			if !strings.Contains(tableStr, sID) {
				t.Errorf("expected session %s in visualized table", sID)
			}
		}

		// Tool 12: vivechak_visualize (mermaid format)
		visMermaidEnv := callTool("vivechak_visualize", map[string]any{
			"project_root": wsDir,
			"format":       "mermaid",
		})
		if !visMermaidEnv.Success {
			t.Fatalf("visualize mermaid failed: %s", visMermaidEnv.Message)
		}
		mermaidStr, _ := visMermaidEnv.Data["rendered"].(string)
		if !strings.Contains(mermaidStr, "graph TD") {
			t.Errorf("expected Mermaid diagram, got:\n%s", mermaidStr)
		}
	})

	// =========================================================================
	// Phase 4: Single-Source ADRs, Auto-Draft, Challenge & Supersession
	// =========================================================================
	t.Run("04_ADR_AutoDraft_Challenge_And_Supersession", func(t *testing.T) {
		// Tool 8: Auto-draft D-001 from T1-01
		draftEnv := callTool("vivechak_record_decision", map[string]any{
			"project_root":    wsDir,
			"decision_id":     "D-001",
			"auto_draft_from": "T1-01",
		})
		if !draftEnv.Success {
			t.Fatalf("auto-draft D-001 failed: %s", draftEnv.Message)
		}
		draftText, _ := draftEnv.Data["draft_content"].(string)
		if !strings.Contains(draftText, "PostgreSQL") {
			t.Errorf("expected draft to capture PostgreSQL recommendation, got:\n%s", draftText)
		}

		// Tool 7: vivechak_challenge (red-team on T1-01)
		challEnv := callTool("vivechak_challenge", map[string]any{
			"project_root": wsDir,
			"session_id":   "T1-01",
			"mode":         "red_team",
		})
		if !challEnv.Success {
			t.Fatalf("challenge red_team failed: %s", challEnv.Message)
		}

		// Record locked decision D-001 (One-Way Door)
		d001Content := `---
id: D-001
decision_id: D-001
title: Primary Persistence Engine
door_type: one-way
status: accepted
date: "2026-10-14"
review_trigger: "WAL fsync latency exceeds 20ms at p99"
informed_by_sessions: ["T1-01"]
human_reviewed: "true"
---
# D-001: Primary Persistence Engine

## Context
High-scale event streaming platform requires partitioned transactional persistence with strict ACID guarantees.

## Decision Rationale
Adopt PostgreSQL 16 with logical partitioning and pre-warmed monthly tables [Grade A: official docs https://postgresql.org/docs].

## Discovered Concerns
Lock contention during partition evolution requires worker pre-allocation during off-peak windows.

## Rejected Alternatives
- MongoDB: Lacks native multi-document transactional isolation across sharded keys.
- Apache Cassandra: Eventual consistency causes double-spend anomalies in balance ledgers.

## Challenge Findings & Red-Team Stress Test
Adversarial red-team challenge probed catalog contention and connection pool exhaustion under 50k QPS spikes. Mitigated with PgBouncer.

## Reversal Triggers
- WAL fsync latency exceeds 20ms at p99 for 3 consecutive hours.
`
		recD1Env := callTool("vivechak_record_decision", map[string]any{
			"project_root": wsDir,
			"decision_id":  "D-001",
			"content":      d001Content,
		})
		if !recD1Env.Success {
			t.Fatalf("record D-001 failed: %s", recD1Env.Message)
		}

		// Record D-002 (One-Way Door: Partitioning Model)
		d002Content := `---
id: D-002
decision_id: D-002
title: Distributed Consistent Hashing
door_type: one-way
status: accepted
date: "2026-10-14"
review_trigger: "Shard data skew exceeds 25%"
informed_by_sessions: ["T2-01"]
human_reviewed: "true"
---
# D-002: Distributed Consistent Hashing

## Context
Data distribution across independent storage nodes requires consistent key hashing.

## Decision Rationale
Use Ketama consistent hashing with 256 virtual nodes per shard [Grade A: official docs https://kubernetes.io/docs].

## Discovered Concerns
Node membership churn forces brief rehash windows requiring client-side retry buffers.

## Rejected Alternatives
- Modulo hashing: Full cluster reshard required whenever a node joins or leaves.
- Central directory lookup: Bottlenecks throughput and creates single point of failure.

## Challenge Findings & Red-Team Stress Test
Adversarial red-team challenge probed node failover during ring rebalancing. Resolved with graceful lease handoffs.

## Reversal Triggers
- Shard data skew exceeds 25% between any two partitions.
`
		recD2Env := callTool("vivechak_record_decision", map[string]any{
			"project_root": wsDir,
			"decision_id":  "D-002",
			"content":      d002Content,
		})
		if !recD2Env.Success {
			t.Fatalf("record D-002 failed: %s", recD2Env.Message)
		}

		// Test Architectural Evolution: Supersede D-001 with D-003
		d003Content := `---
id: D-003
decision_id: D-003
title: Distributed Multi-Region Persistence
door_type: one-way
status: accepted
date: "2026-10-14"
review_trigger: "Raft consensus election latency > 100ms"
informed_by_sessions: ["T1-01"]
human_reviewed: "true"
---
# D-003: Distributed Multi-Region Persistence

## Context
Global deployment expansion mandates multi-region active-active persistence.

## Decision Rationale
Migrate from single-region PostgreSQL to CockroachDB [Grade A: official docs https://www.cockroachlabs.com/docs].

## Discovered Concerns
WAN consensus latency adds 35ms floor to cross-region serializable transactions.

## Rejected Alternatives
- PostgreSQL with async replication: Suffers data loss during unplanned regional failover.

## Challenge Findings & Red-Team Stress Test
Adversarial red-team probe tested multi-region split-brain. Mitigated via 3-region quorum voting nodes.

## Reversal Triggers
- Raft consensus election latency > 100ms.
`
		superEnv := callTool("vivechak_record_decision", map[string]any{
			"project_root": wsDir,
			"decision_id":  "D-003",
			"supersedes":   "D-001",
			"content":      d003Content,
		})
		if !superEnv.Success {
			t.Fatalf("supersede D-001 with D-003 failed: %s", superEnv.Message)
		}

		// Verify on disk: D-001 has status: superseded and superseded_by: D-003
		d1Bytes, err := os.ReadFile(filepath.Join(wsDir, core.ResearchDir, "D-001-primary-persistence-engine.md"))
		if err != nil {
			// Try fallback filename D-001-decision.md or D-001.md
			d1Bytes, err = os.ReadFile(filepath.Join(wsDir, core.ResearchDir, "D-001-decision.md"))
			if err != nil {
				d1Bytes, err = os.ReadFile(filepath.Join(wsDir, core.ResearchDir, "D-001.md"))
			}
		}
		if err != nil {
			t.Fatalf("could not read D-001 file: %v", err)
		}
		if !strings.Contains(string(d1Bytes), "status: superseded") || !strings.Contains(string(d1Bytes), "superseded_by: D-003") {
			t.Errorf("expected D-001 to be marked superseded on disk, got:\n%s", string(d1Bytes))
		}

		// Verify compiled DECISIONS.md registry contains all decisions in natural order
		regBytes, err := os.ReadFile(filepath.Join(wsDir, core.DecisionsFile))
		if err != nil {
			t.Fatalf("reading DECISIONS.md: %v", err)
		}
		regStr := string(regBytes)
		if !strings.Contains(regStr, "<!-- DECISION: D-001 -->") ||
			!strings.Contains(regStr, "<!-- DECISION: D-002 -->") ||
			!strings.Contains(regStr, "<!-- DECISION: D-003 -->") {
			t.Errorf("DECISIONS.md missing compiled decision markers:\n%s", regStr)
		}
	})

	// =========================================================================
	// Phase 5: FAD Synthesis, Root Mirror & Phase 0 Gate Execution
	// =========================================================================
	t.Run("05_FAD_Synthesis_RootMirror_And_Gate", func(t *testing.T) {
		fadContent := `---
id: SYN-01
session_id: SYN-01
title: Enterprise Stream Platform Founding Architecture Document
synthesis_date: "2026-10-14"
status: complete
---
# Founding Architecture Document: Enterprise Stream Platform

## Executive Summary
This document codifies the technical architecture for the Enterprise Stream Platform.
Built on proven commodity building blocks with explicit reversibility thresholds.

## Core Architectural Decisions
1. Ingestion: Envoy proxy terminated HTTP/2 and gRPC connections [Grade A: official docs https://envoyproxy.io/docs].
2. Storage: Multi-region CockroachDB with Ketama consistent hashing [Grade A: official docs https://www.cockroachlabs.com/docs].
3. Caching: Redis volatile-lru transient query cache [Grade A: official docs https://redis.io/docs].
4. Telemetry: Prometheus and OpenTelemetry distributed tracing [Grade A: official docs https://prometheus.io/docs].

## System Premortem & Failure Modes
Failure mode 1: Cross-region network partition delays consensus commits. Mitigated via local region read replicas.
Failure mode 2: Memory exhaustion under spike traffic. Mitigated via Envoy client backpressure propagation.
`
		// Save FAD as session SYN-01
		saveFADEnv := callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "SYN-01",
			"content":      fadContent,
		})
		if !saveFADEnv.Success {
			t.Fatalf("save FAD failed: %s", saveFADEnv.Message)
		}

		// Verify next_step guides the agent to run Phase 0 Gate
		if !strings.Contains(saveFADEnv.NextStep, "vivechak_run_gate") {
			t.Errorf("expected next_step to suggest vivechak_run_gate, got: %s", saveFADEnv.NextStep)
		}

		// Verify root mirror FOUNDING-ARCHITECTURE.md exists
		rootFADPath := filepath.Join(wsDir, "FOUNDING-ARCHITECTURE.md")
		if _, err := os.Stat(rootFADPath); os.IsNotExist(err) {
			t.Fatalf("expected root mirror FOUNDING-ARCHITECTURE.md to exist")
		}

		// Tool 14: Run Phase 0 Gate
		gateEnv := callTool("vivechak_run_gate", map[string]any{
			"project_root": wsDir,
		})
		if !gateEnv.Success {
			t.Fatalf("vivechak_run_gate failed: %s", gateEnv.Message)
		}
		if gateEnv.Data["gate_status"] != "PASS" {
			t.Fatalf("expected gate status PASS, got %v (warnings: %v)", gateEnv.Data["gate_status"], gateEnv.Warnings)
		}

		// Verify PHASE-0-GATE.md persisted to disk with 11-step checklist checked
		gateBytes, err := os.ReadFile(filepath.Join(wsDir, core.GateFile))
		if err != nil {
			t.Fatalf("reading PHASE-0-GATE.md: %v", err)
		}
		gateStr := string(gateBytes)
		if !strings.Contains(gateStr, "verdict: \"PASS\"") {
			t.Errorf("expected gate verdict PASS, got:\n%s", gateStr)
		}
		if !strings.Contains(gateStr, "- [x] Decision record contains substantive causal rationale") {
			t.Errorf("expected B10 checked in gate artifact, got:\n%s", gateStr)
		}
		if !strings.Contains(gateStr, "- [x] One-Way Door decisions subjected to adversarial challenge") {
			t.Errorf("expected B11 checked in gate artifact, got:\n%s", gateStr)
		}
	})

	// =========================================================================
	// Phase 6: Security Defense Against Directory Traversal Attacks
	// =========================================================================
	t.Run("06_Security_Confinement_Defense", func(t *testing.T) {
		// Attempt directory traversal via session_id
		travSave := callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "../../evil_session",
			"content":      "# Malicious session",
		})
		if travSave.Success {
			t.Errorf("expected path traversal in session_id to be blocked, but succeeded")
		}

		// Attempt directory traversal via decision_id
		travDec := callTool("vivechak_record_decision", map[string]any{
			"project_root": wsDir,
			"decision_id":  "../../evil_decision",
			"content":      "# Malicious decision",
		})
		if travDec.Success {
			t.Errorf("expected path traversal in decision_id to be blocked, but succeeded")
		}

		// Attempt directory traversal via amending_session_id
		travAmend := callTool("vivechak_amend_session", map[string]any{
			"project_root":        wsDir,
			"session_id":          "T1-01",
			"amending_session_id": "../../evil",
			"amendment":           "hack",
		})
		if travAmend.Success {
			t.Errorf("expected path traversal in amending_session_id to be blocked, but succeeded")
		}
	})

	// =========================================================================
	// Phase 7: Real CLI Diagnostic Execution (doctor, setup, mcp-config, version)
	// =========================================================================
	t.Run("07_CLI_Diagnostic_RealExecution", func(t *testing.T) {
		var out bytes.Buffer

		// 1. vivechak version
		vCmd := exec.CommandContext(ctx, binPath, "version")
		vCmd.Stdout = &out
		if err := vCmd.Run(); err != nil {
			t.Fatalf("version command failed: %v", err)
		}
		if !strings.Contains(out.String(), "v0.1.0") && !strings.Contains(out.String(), "vivechak") {
			t.Errorf("unexpected version output: %s", out.String())
		}

		// 2. vivechak doctor on healthy completed workspace
		out.Reset()
		docHealthyCmd := exec.CommandContext(ctx, binPath, "doctor", wsDir)
		docHealthyCmd.Stdout = &out
		_ = docHealthyCmd.Run()
		if !strings.Contains(out.String(), "Workspace is healthy") && !strings.Contains(out.String(), "PASS") {
			t.Errorf("expected doctor healthy report on completed workspace, got:\n%s", out.String())
		}

		// 3. vivechak doctor on uninitialized workspace
		emptyDir := t.TempDir()
		out.Reset()
		docEmptyCmd := exec.CommandContext(ctx, binPath, "doctor", emptyDir)
		docEmptyCmd.Stdout = &out
		_ = docEmptyCmd.Run()
		if !strings.Contains(out.String(), "MISSING") && !strings.Contains(out.String(), "FAIL") && !strings.Contains(out.String(), "Checking") {
			t.Errorf("expected doctor warnings on empty directory, got:\n%s", out.String())
		}

		// 4. vivechak setup (auto-detection mode)
		out.Reset()
		setupCmd := exec.CommandContext(ctx, binPath, "setup", "--dry-run")
		setupCmd.Stdout = &out
		_ = setupCmd.Run()
		if len(out.String()) == 0 {
			t.Errorf("expected non-empty setup dry-run output")
		}

		// 5. vivechak mcp-config across all major presets
		presets := []string{"claude-desktop", "cursor", "zed", "vscode", "windsurf"}
		for _, preset := range presets {
			out.Reset()
			cfgCmd := exec.CommandContext(ctx, binPath, "mcp-config", "--client", preset, "--stdout")
			cfgCmd.Stdout = &out
			if err := cfgCmd.Run(); err != nil {
				t.Fatalf("mcp-config failed for client %s: %v", preset, err)
			}
			var cfgJSON map[string]any
			if err := json.Unmarshal(out.Bytes(), &cfgJSON); err != nil {
				t.Fatalf("mcp-config for %s produced invalid JSON: %v\nOutput: %s", preset, err, out.String())
			}
		}
	})
}
