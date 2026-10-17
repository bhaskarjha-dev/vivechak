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

// TestLiveComprehensiveHardeningE2E conducts a massive, zero-trust end-to-end verification
// of all modifications (E-28 through E-36) through real binary builds, subprocess CLI commands,
// live MCP JSON-RPC protocol transport, multi-scope workspace files, and Phase 0 Gate execution.
func TestLiveComprehensiveHardeningE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping comprehensive live hardening test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	// Step 1: Build genuine vivechak binary
	binDir := t.TempDir()
	binName := "vck_hardening_test"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build vivechak binary: %v\noutput: %s", err, string(out))
	}

	// Step 2: Subprocess CLI Verification
	t.Run("CLI_Commands", func(t *testing.T) {
		// Version command
		vCmd := exec.CommandContext(ctx, binPath, "version")
		vOut, err := vCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("vck version failed: %v\noutput: %s", err, string(vOut))
		}
		if !strings.Contains(string(vOut), "vivechak") && !strings.Contains(string(vOut), "vck") {
			t.Errorf("expected version output to mention vivechak, got: %s", string(vOut))
		}

		// MCP config command
		cfgCmd := exec.CommandContext(ctx, binPath, "mcp-config")
		cfgOut, err := cfgCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("vck mcp-config failed: %v\noutput: %s", err, string(cfgOut))
		}
		var parsedConfig map[string]any
		if err := json.Unmarshal(cfgOut, &parsedConfig); err != nil {
			t.Fatalf("vck mcp-config produced invalid JSON: %v\noutput: %s", err, string(cfgOut))
		}
		if _, ok := parsedConfig["mcpServers"]; !ok {
			t.Errorf("expected mcpServers key in mcp-config output, got: %s", string(cfgOut))
		}
	})

	// Step 3: Spawn live MCP server as a real subprocess over stdio JSON-RPC
	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "hardening-e2e-verifier", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	callTool := func(toolName string, args map[string]any) (testEnvelope, *mcp.CallToolResult) {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		})
		if err != nil {
			t.Fatalf("cs.CallTool(%q) protocol error: %v", toolName, err)
		}
		env := parseTestEnvelope(t, res)
		return env, res
	}

	// =========================================================================
	// Scenario A: Comparison Scope & Code Fence Tracker Verification
	// =========================================================================
	t.Run("Scenario_Comparison_Fences", func(t *testing.T) {
		cmpDir := t.TempDir()

		// 1. Init Comparison Workspace
		env, _ := callTool("vivechak_init", map[string]any{
			"project_root": cmpDir,
			"scope":        "comparison",
		})
		if !env.Success {
			t.Fatalf("vivechak_init failed: %s", env.Message)
		}

		// Verify only comparison template was copied
		tplEntries, _ := os.ReadDir(filepath.Join(cmpDir, core.TemplatesDir))
		if len(tplEntries) != 1 || tplEntries[0].Name() != "COMPARISON-SESSION.template.md" {
			t.Errorf("expected only COMPARISON-SESSION.template.md, got: %v", tplEntries)
		}

		// 2. Save Comparison Plan
		planContent := `# Bounded Comparison: BadgerDB vs Pebble
| Option | Focus | Status |
|---|---|---|
| BadgerDB | Pure Go LSM-tree | candidate |
| Pebble | CockroachDB RocksDB replacement | candidate |
`
		env, _ = callTool("vivechak_save_plan", map[string]any{
			"project_root": cmpDir,
			"scope":        "comparison",
			"decision_id":  "C-001",
			"slug":         "badger-pebble",
			"content":      planContent,
		})
		if !env.Success {
			t.Fatalf("vivechak_save_plan failed: %s", env.Message)
		}

		// 3. Save Comparison Session with arbitrary code fences (``` and ~~~)
		sessionMD := `---
session_id: C-001
title: BadgerDB vs Pebble
date: 2026-10-17
status: complete
---
## Prior
BadgerDB is pure Go, Pebble is optimized for CockroachDB.

## Key Findings
- BadgerDB uses WiscKey separation for value pointers. Grade A (https://github.com/dgraph-io/badger)
- Pebble provides efficient iterator bounds. Grade A (https://github.com/cockroachdb/pebble)

Here is sample markdown inside tilde fences that must NOT prematurely end the section:
~~~~markdown
## Sources
| Grade | Status |
| Grade A | fetched |
~~~~

And sample code inside 4-backtick code blocks:
` + "````go" + `
// ## Evaluated Options
func main() {}
` + "````" + `

## Discovered Concerns
BadgerDB value log garbage collection creates noticeable I/O spikes under sustained heavy write loads.

## Delta
| Prior Belief | Status | Evidence | Impact |
| BadgerDB is fast | Updated | Grade A (benchmarks) | High |

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Status | Supports |
| 1 | https://github.com/dgraph-io/badger | Grade A | fresh | fetched | BadgerDB |
| 2 | https://github.com/cockroachdb/pebble | Grade A | fresh | fetched | Pebble |
`
		env, _ = callTool("vivechak_save_session", map[string]any{
			"project_root": cmpDir,
			"session_id":   "C-001",
			"content":      sessionMD,
		})
		if !env.Success {
			t.Fatalf("vivechak_save_session failed: %s", env.Message)
		}

		// 4. Validate workspace
		env, _ = callTool("vivechak_validate", map[string]any{
			"project_root": cmpDir,
		})
		if !env.Success {
			t.Fatalf("vivechak_validate failed: %s", env.Message)
		}
	})

	// =========================================================================
	// Scenario B: Decision Scope & Unified Case-Insensitive Gate Verification
	// =========================================================================
	t.Run("Scenario_Decision_Gate_Rigor", func(t *testing.T) {
		decDir := t.TempDir()

		// 1. Init Decision Workspace
		env, _ := callTool("vivechak_init", map[string]any{
			"project_root": decDir,
			"scope":        "decision",
		})
		if !env.Success {
			t.Fatalf("vivechak_init failed: %s", env.Message)
		}

		// 2. Save Decision Plan
		planMD := `# Decision Plan: D-001 Consensus Engine
| ID | Topic | Dependencies | Output File |
|---|---|---|---|
| D-001-S01 | Raft Protocol Evaluation | none | sessions/D-001-S01.md |
`
		env, _ = callTool("vivechak_save_plan", map[string]any{
			"project_root": decDir,
			"scope":        "decision",
			"decision_id":  "D-001",
			"slug":         "consensus",
			"content":      planMD,
		})
		if !env.Success {
			t.Fatalf("vivechak_save_plan failed: %s", env.Message)
		}

		// 3. Save Research Session D-001-S01 (One-Way Door with red-team challenge findings)
		sessMD := `---
session_id: D-001-S01
title: Raft Protocol Evaluation
date: 2026-10-17
door_type: one-way
status: complete
informs_decisions:
  - D-001
---
## Prior
Raft is simpler than Multi-Paxos and sufficient for leader-elected clusters.

## Key Findings
- Raft guarantees deterministic log consistency under majority quorum. Grade A (https://raft.github.io)
- Etcd raft library is widely used in production. Grade A (https://etcd.io/docs/v3.5/learning/design-learner/)

## Discovered Concerns
Pre-vote extension is strictly mandatory in lossy networks to avoid disruptive elections caused by partitioned nodes.

## Delta
| Prior | Status | Evidence | Impact |
| Raft is easy | Updated | Grade A (https://raft.github.io) | Medium |

## Red-Team Stress Test
Adversarial split-brain network partition analysis was executed. Split votes and split-brain scenarios are prevented by Raft joint consensus membership changes and lease checks.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Status | Supports |
| 1 | https://raft.github.io | Grade A | fresh | fetched | Raft safety |
| 2 | https://etcd.io/docs/v3.5/learning/design-learner/ | Grade A | fresh | fetched | Etcd raft |
`
		env, _ = callTool("vivechak_save_session", map[string]any{
			"project_root": decDir,
			"session_id":   "D-001-S01",
			"content":      sessMD,
		})
		if !env.Success {
			t.Fatalf("vivechak_save_session failed: %s", env.Message)
		}

		// 4. Record Decision D-001 (One-Way Door with review trigger and rejected alternatives)
		adrMD := `---
decision_id: D-001
title: Raft Consensus Engine
date: 2026-10-17
status: accepted
door_type: one-way
review_trigger: Cluster size exceeds 9 voting members or cross-region latency p99 > 80ms
informed_by_sessions:
  - D-001-S01
---
# D-001: Raft Consensus Engine

## Context
Distributed state machine replication requires strong consistency guarantees.

## Decision
We select hashicorp/raft or etcd/raft for leader-based cluster consensus. Grade A (https://raft.github.io)

## Alternatives Considered
- Multi-Paxos: significantly higher implementation complexity and subtle edge-case failure modes under concurrent leaders.
- Primary-Backup: insufficient fault tolerance; single point of failure during split-brain network partitioning.

## Consequences
- Positive: battle-tested leader leases and deterministic log compaction.
- Negative: write throughput bounded by single active leader network bandwidth.

## Red-Team Stress Test
Validated split-brain partition tolerance under asymmetric network partitioning.
`
		env, _ = callTool("vivechak_record_decision", map[string]any{
			"project_root": decDir,
			"decision_id":  "D-001",
			"content":      adrMD,
		})
		if !env.Success {
			t.Fatalf("vivechak_record_decision failed: %s", env.Message)
		}

		// 5. Test Case-Insensitive Deduplication in Validate
		// Both standalone D-001.md and DECISIONS.md exist in decDir
		env, _ = callTool("vivechak_validate", map[string]any{
			"project_root": decDir,
		})
		if !env.Success {
			t.Fatalf("vivechak_validate failed: %s", env.Message)
		}
		decStats := env.Data["decisions"].(map[string]any)
		validCount := int(decStats["valid"].(float64))
		if validCount != 1 {
			t.Errorf("expected exactly 1 deduplicated valid decision, got: %d", validCount)
		}

		// 6. Test Phase 0 Gate Execution
		env, _ = callTool("vivechak_run_gate", map[string]any{
			"project_root": decDir,
		})
		if !env.Success {
			t.Fatalf("vivechak_run_gate failed: %s", env.Message)
		}
		if env.Data["gate_passed"] != true {
			t.Fatalf("expected gate to pass, got: %+v\nwarnings: %v", env.Data, env.Warnings)
		}
		if env.Data["gate_status"] != "PASS" {
			t.Errorf("expected gate status 'PASS', got: %v", env.Data["gate_status"])
		}

		// Verify PHASE-0-GATE.md was auto-persisted
		gatePath := filepath.Join(decDir, core.ResearchDir, "PHASE-0-GATE.md")
		if _, err := os.Stat(gatePath); err != nil {
			t.Errorf("expected PHASE-0-GATE.md to be created, err: %v", err)
		}

		// 7. Verify CLI Doctor on Decision Workspace
		docCmd := exec.CommandContext(ctx, binPath, "doctor", decDir)
		docOut, err := docCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("vck doctor failed on valid workspace: %v\noutput: %s", err, string(docOut))
		}
		if !strings.Contains(string(docOut), "Workspace is healthy") {
			t.Errorf("expected 'Workspace is healthy', got: %s", string(docOut))
		}
	})

	// =========================================================================
	// Scenario C: Discovered Concerns Length Gate Enforcement
	// =========================================================================
	t.Run("Scenario_Concerns_Threshold_Gate", func(t *testing.T) {
		shortDir := t.TempDir()
		_, _ = callTool("vivechak_init", map[string]any{
			"project_root": shortDir,
			"scope":        "project",
		})

		// 49 characters under ## Discovered Concerns (must fail for one-way door)
		shortSess := `---
session_id: T1-01
title: Auth Strategy
date: 2026-10-17
door_type: one-way
status: complete
---
## Key Findings
- Finding Grade A (https://example.com)

## Discovered Concerns
1234567890123456789012345678901234567890123456789
`
		env, _ := callTool("vivechak_save_session", map[string]any{
			"project_root": shortDir,
			"session_id":   "T1-01",
			"content":      shortSess,
		})
		// Tool returns success=true with status: draft enforcement and defect observations
		if env.Data["status"] != "draft" {
			t.Errorf("expected session with 49 chars of concerns to remain draft, got: %v", env.Data["status"])
		}

		// 50 characters under ## Discovered Concerns (meets threshold)
		validSess := strings.Replace(shortSess,
			"1234567890123456789012345678901234567890123456789",
			"12345678901234567890123456789012345678901234567890", 1)
		env, _ = callTool("vivechak_save_session", map[string]any{
			"project_root": shortDir,
			"session_id":   "T1-01",
			"content":      validSess,
		})
		if env.Data["status"] != "valid" && env.Data["status"] != "valid-with-warnings" {
			t.Errorf("expected session with 50 chars of concerns to be valid, got: %v", env.Data["status"])
		}
		if env.Data["validation_passed"] != true {
			t.Errorf("expected validation_passed to be true, got: %v", env.Data["validation_passed"])
		}
		savedData, _ := os.ReadFile(filepath.Join(shortDir, core.SessionsDir, "T1-01.md"))
		if !strings.Contains(string(savedData), "status: complete") {
			t.Errorf("expected saved file frontmatter to have status: complete, got:\n%s", string(savedData))
		}
	})

	// =========================================================================
	// Scenario D: CLI Scope Inference Verification
	// =========================================================================
	t.Run("Scenario_CLI_Scope_Inference", func(t *testing.T) {
		cases := []struct {
			input string
			want  core.Scope
		}{
			{"vck plan: Distributed Database Engine", core.ScopeProject},
			{"vck decide: BadgerDB vs Pebble", core.ScopeDecision},
			{"vck compare: RocksDB versus LevelDB", core.ScopeComparison},
			{"vck vs: Redis vs Memcached", core.ScopeComparison},
			{"should we choose PostgreSQL or MySQL?", core.ScopeDecision},
			{"compare grpc and rest", core.ScopeComparison},
			{"hey im confused bw winston and pino for my express app which one shud i use?? need fast json logging pls help", core.ScopeComparison},
			{"bro shuld i use react 19 server actions or tanstack react query for fetching data in my nextjs dashboard? im totally confused", core.ScopeDecision},
			{"full architecture for cloud native stream processing", core.ScopeProject},
		}

		for _, tc := range cases {
			got := core.InferScope(tc.input)
			if got != tc.want {
				t.Errorf("core.InferScope(%q) = %q, want %q", tc.input, got, tc.want)
			}
		}
	})
}
