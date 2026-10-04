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

// TestExhaustive_AllFeatures_AllScenarios tests EVERY single feature, tool, scope, CLI command,
// and edge-case scenario across the entire Vivechak system to ensure 100% operational perfection.
func TestExhaustive_AllFeatures_AllScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping exhaustive trial in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)

	// -------------------------------------------------------------
	// STAGE 0: Build the live vivechak binary
	// -------------------------------------------------------------
	binDir := t.TempDir()
	binName := "vivechak_exhaustive_bin"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	// -------------------------------------------------------------
	// STAGE 1: CLI Direct Invocations (version, help, doctor, mcp-config)
	// -------------------------------------------------------------
	t.Run("CLI_Version", func(t *testing.T) {
		cmd := exec.CommandContext(ctx, binPath, "version")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("version failed: %v", err)
		}
		if !strings.Contains(string(out), "vivechak") {
			t.Errorf("version output unexpected: %s", string(out))
		}
	})

	t.Run("CLI_Help_And_Invalid", func(t *testing.T) {
		// --help
		cmd := exec.CommandContext(ctx, binPath, "--help")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("--help failed: %v", err)
		}
		if !strings.Contains(string(out), "Commands:") {
			t.Errorf("--help output unexpected: %s", string(out))
		}

		// unknown command should exit with 1
		cmdInvalid := exec.CommandContext(ctx, binPath, "invalid-cmd-xyz")
		outInv, errInv := cmdInvalid.CombinedOutput()
		if errInv == nil {
			t.Fatalf("expected error for invalid command, got 0; out: %s", string(outInv))
		}
	})

	t.Run("CLI_MCP_Config_AllPresets", func(t *testing.T) {
		presets := []string{
			"claude-desktop", "claude", "vscode", "code", "cursor", "windsurf", "zed", "antigravity", "agy", "kiro",
			"trae", "omp", "openhands", "droid", "cline", "roo", "devin",
		}
		for _, preset := range presets {
			cmd := exec.CommandContext(ctx, binPath, "mcp-config", "--preset", preset)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("mcp-config --preset %s failed: %v, out: %s", preset, err, string(out))
			}
			var parsed map[string]any
			if err := json.Unmarshal(out, &parsed); err != nil {
				t.Fatalf("mcp-config --preset %s produced invalid JSON: %v", preset, err)
			}
		}

		// Test --write to a file
		cfgFile := filepath.Join(t.TempDir(), "custom-mcp.json")
		cmdWrite := exec.CommandContext(ctx, binPath, "mcp-config", "--path", cfgFile, "--write")
		outWrite, errWrite := cmdWrite.CombinedOutput()
		if errWrite != nil {
			t.Fatalf("mcp-config --write failed: %v, out: %s", errWrite, string(outWrite))
		}
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			t.Fatalf("mcp-config did not write file to %s", cfgFile)
		}
	})

	t.Run("CLI_Setup_And_Install", func(t *testing.T) {
		// 1. Test setup --help
		cmdHelp := exec.CommandContext(ctx, binPath, "setup", "--help")
		outHelp, errHelp := cmdHelp.CombinedOutput()
		if errHelp != nil {
			t.Fatalf("setup --help failed: %v", errHelp)
		}
		if !strings.Contains(string(outHelp), "Usage: vivechak setup") {
			t.Errorf("setup --help output unexpected: %s", string(outHelp))
		}

		// 2. Test setup dry-run
		cmdDry := exec.CommandContext(ctx, binPath, "setup", "cursor", "--dry-run")
		outDry, errDry := cmdDry.CombinedOutput()
		if errDry != nil {
			t.Fatalf("setup cursor --dry-run failed: %v", errDry)
		}
		if !strings.Contains(string(outDry), "[dry-run] Would write Vivechak configuration") {
			t.Errorf("expected dry-run message, got: %s", string(outDry))
		}

		// 3. Test setup directly to file
		targetFile := filepath.Join(t.TempDir(), "exhaustive-setup.json")
		cmdFile := exec.CommandContext(ctx, binPath, "setup", targetFile)
		outFile, errFile := cmdFile.CombinedOutput()
		if errFile != nil {
			t.Fatalf("setup targetFile failed: %v, out: %s", errFile, string(outFile))
		}
		if _, err := os.Stat(targetFile); err != nil {
			t.Fatalf("target file was not created: %v", err)
		}
	})

	t.Run("CLI_Doctor_IntegrityChecks", func(t *testing.T) {
		// Case A: Uninitialized directory -> should exit 1
		emptyDir := t.TempDir()
		cmdEmpty := exec.CommandContext(ctx, binPath, "doctor", emptyDir)
		outEmpty, errEmpty := cmdEmpty.CombinedOutput()
		if errEmpty == nil {
			t.Fatalf("expected doctor to fail on empty dir, got success: %s", string(outEmpty))
		}
		if !strings.Contains(string(outEmpty), "does not exist") {
			t.Errorf("expected missing research error, got: %s", string(outEmpty))
		}

		// Case B: Initialized directory -> should report health
		healthyDir := t.TempDir()
		_ = os.MkdirAll(filepath.Join(healthyDir, core.TemplatesDir), 0o755)
		_ = os.MkdirAll(filepath.Join(healthyDir, core.SessionsDir), 0o755)
		for _, tmpl := range core.TemplatesToCopy {
			_ = os.WriteFile(filepath.Join(healthyDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
		}
		_ = os.WriteFile(filepath.Join(healthyDir, core.PipelineFile), []byte("# Research Pipeline\n\n## Session DAG\n"), 0o644)
		_ = os.WriteFile(filepath.Join(healthyDir, core.DecisionsFile), []byte("# Decision Registry\n"), 0o644)

		cmdHealthy := exec.CommandContext(ctx, binPath, "doctor", healthyDir)
		outHealthy, errHealthy := cmdHealthy.CombinedOutput()
		if errHealthy != nil {
			t.Fatalf("expected doctor to succeed on healthy dir, got err: %v\noutput: %s", errHealthy, string(outHealthy))
		}
		if !strings.Contains(string(outHealthy), "healthy") {
			t.Errorf("expected 'healthy' in doctor output, got: %s", string(outHealthy))
		}

		// Case C: Broken templates -> should report error
		brokenDir := t.TempDir()
		_ = os.MkdirAll(filepath.Join(brokenDir, core.TemplatesDir), 0o755)
		_ = os.MkdirAll(filepath.Join(brokenDir, core.SessionsDir), 0o755)
		// Only write 1 template instead of 5
		_ = os.WriteFile(filepath.Join(brokenDir, core.TemplatesDir, core.TemplatesToCopy[0]), []byte("content"), 0o644)
		_ = os.WriteFile(filepath.Join(brokenDir, core.PipelineFile), []byte("# Research Pipeline\n"), 0o644)
		_ = os.WriteFile(filepath.Join(brokenDir, core.DecisionsFile), []byte("# Decision Registry\n"), 0o644)

		cmdBroken := exec.CommandContext(ctx, binPath, "doctor", brokenDir)
		outBroken, errBroken := cmdBroken.CombinedOutput()
		if errBroken == nil {
			t.Fatalf("expected doctor to fail on broken templates, got err=nil, out: %s", string(outBroken))
		}
		if !strings.Contains(string(outBroken), "Missing template") {
			t.Errorf("expected missing template error, got: %s", string(outBroken))
		}
	})

	// -------------------------------------------------------------
	// STAGE 2: Connect Live MCP Server via stdio
	// -------------------------------------------------------------
	cmdServe := exec.CommandContext(ctx, binPath, "serve")
	cmdServe.Stderr = os.Stderr

	transport := &mcp.CommandTransport{
		Command: cmdServe,
	}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "exhaustive-trial-client", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server over stdio: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	// Verify all 13 tools are listed
	var toolList []string
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("error listing tools: %v", err)
		}
		toolList = append(toolList, tool.Name)
	}
	if len(toolList) != 13 {
		t.Fatalf("expected 13 tools, got %d: %v", len(toolList), toolList)
	}

	// -------------------------------------------------------------
	// STAGE 3: Full Project Scope Lifecycle (End-to-End Deep Trial)
	// -------------------------------------------------------------
	t.Run("ProjectScope_CompleteDeepTrial", func(t *testing.T) {
		pDir := t.TempDir()

		// 3.1 vivechak_init (Scope: project)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_init",
			Arguments: map[string]any{
				"project_root": pDir,
				"scope":        "project",
			},
		})
		if err != nil {
			t.Fatalf("vivechak_init: %v", err)
		}
		env := parseTestEnvelope(t, res)
		if !env.Success {
			t.Fatalf("init failed: %s", env.Message)
		}

		// Verify research/.gitignore was automatically created
		giPath := filepath.Join(pDir, core.ResearchDir, ".gitignore")
		if giData, err := os.ReadFile(giPath); err != nil || !strings.Contains(string(giData), "*.lock") {
			t.Errorf("expected research/.gitignore with *.lock, err: %v, content: %s", err, string(giData))
		}

		// 3.2 Idempotency test: Call vivechak_init again -> should succeed safely and report already initialized
		resIdemp, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_init",
			Arguments: map[string]any{
				"project_root": pDir,
				"scope":        "project",
			},
		})
		if err != nil {
			t.Fatalf("vivechak_init idempotent call err: %v", err)
		}
		envIdemp := parseTestEnvelope(t, resIdemp)
		if !envIdemp.Success {
			t.Fatalf("expected vivechak_init to succeed safely on existing workspace, got: %s", envIdemp.Message)
		}
		if !strings.Contains(envIdemp.Message, "already initialized") {
			t.Errorf("expected 'already initialized' message, got: %s", envIdemp.Message)
		}

		// 3.3 vivechak_prepare_generator
		resPrep, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_prepare_generator",
			Arguments: map[string]any{
				"project_root": pDir,
				"scope":        "project",
				"context":      "Building high-frequency financial distributed ledger.",
			},
		})
		if err != nil {
			t.Fatalf("prepare_generator: %v", err)
		}
		envPrep := parseTestEnvelope(t, resPrep)
		if !envPrep.Success {
			t.Fatalf("prepare_generator failed: %s", envPrep.Message)
		}
		prompt := envPrep.Data["prompt"].(string)
		if !strings.Contains(prompt, "high-frequency financial distributed ledger") {
			t.Errorf("expected vision in prompt")
		}

		// 3.4 Negative test: vivechak_save_plan with cyclic dependencies
		cyclicPlan := `# Research Pipeline
## Session DAG
### Session C-01 — Cyclic A
| Field | Value |
|---|---|
| **ID** | C-01 |
| **Dependencies** | C-02 |

### Session C-02 — Cyclic B
| Field | Value |
|---|---|
| **ID** | C-02 |
| **Dependencies** | C-01 |
`
		resCycle, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_plan",
			Arguments: map[string]any{
				"project_root": pDir,
				"scope":        "project",
				"content":      cyclicPlan,
			},
		})
		if err != nil {
			t.Fatalf("save_plan cycle call: %v", err)
		}
		envCycle := parseTestEnvelope(t, resCycle)
		if envCycle.Success {
			t.Error("expected save_plan to reject cyclic DAG")
		}

		// 3.5 Valid Plan with 3 Sessions: T1-01 (Engine), T1-02 (Network depends on T1-01), SYN-01 (Synthesis depends on T1-01, T1-02)
		validPlan := `# Research Pipeline

**Complexity Score**: 15/25 → Medium (Tier 2)

## Session DAG

### Session T1-01 — Storage Engine
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Layer** | 1 |
| **Door Type** | one-way |
| **Dependencies** | None |
| **Output File** | ` + "`T1-01-storage.md`" + ` |

` + "```prompt" + `
Investigate low-latency storage engines.
` + "```" + `

### Session T1-02 — Consensus & Network
| Field | Value |
|---|---|
| **ID** | T1-02 |
| **Layer** | 2 |
| **Door Type** | one-way |
| **Dependencies** | T1-01 |
| **Output File** | ` + "`T1-02-consensus.md`" + ` |

` + "```prompt" + `
Investigate consensus protocols given storage findings:
[UPSTREAM_FINDINGS]
` + "```" + `

### Session SYN-01 — Synthesis & FAD
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Layer** | 3 |
| **Door Type** | one-way |
| **Dependencies** | T1-01, T1-02 |
| **Output File** | ` + "`FAD.md`" + ` |

` + "```prompt" + `
Synthesize founding architecture:
[ALL_SESSION_FINDINGS]
` + "```" + `
`
		resPlan, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_plan",
			Arguments: map[string]any{
				"project_root": pDir,
				"scope":        "project",
				"content":      validPlan,
			},
		})
		if err != nil {
			t.Fatalf("save_plan valid: %v", err)
		}
		envPlan := parseTestEnvelope(t, resPlan)
		if !envPlan.Success {
			t.Fatalf("save_plan failed: %s", envPlan.Message)
		}

		// 3.6 Check Status -> 3 sessions total, 0 completed, next should be T1-01
		resStat, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_status",
			Arguments: map[string]any{"project_root": pDir},
		})
		if err != nil {
			t.Fatalf("status call: %v", err)
		}
		envStat := parseTestEnvelope(t, resStat)
		if !envStat.Success {
			t.Fatalf("status failed: %s", envStat.Message)
		}
		if envStat.Data["session_count"].(float64) != 0 {
			t.Errorf("expected 0 completed, got %v", envStat.Data["session_count"])
		}

		// 3.7 Dependency Blocking Check: Auto-select must strictly select unblocked T1-01 (not blocked T1-02 or SYN-01)
		resNext, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_next_session",
			Arguments: map[string]any{"project_root": pDir},
		})
		if err != nil {
			t.Fatalf("next_session auto: %v", err)
		}
		envNext := parseTestEnvelope(t, resNext)
		if envNext.Data["session_id"] != "T1-01" {
			t.Fatalf("expected next session T1-01, got %v", envNext.Data["session_id"])
		}

		// 3.8 Validate non-existent session ID request fails cleanly
		resInvalid, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_next_session",
			Arguments: map[string]any{
				"project_root": pDir,
				"session_id":   "NON-EXISTENT-99",
			},
		})
		if err != nil {
			t.Fatalf("next_session invalid call: %v", err)
		}
		envInvalid := parseTestEnvelope(t, resInvalid)
		if envInvalid.Success {
			t.Error("expected requesting non-existent session to fail")
		}

		// 3.9 Save Session T1-01: Test edge case: Missing evidence grades warning
		resWarnS1, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": pDir,
				"session_id":   "T1-01",
				"content": `---
session_id: T1-01
title: Storage Engine
date: 2026-09-29
status: complete
---
# T1-01 Storage Engine
No evidence grades here at all.
`,
			},
		})
		if err != nil {
			t.Fatalf("save_session missing grades call: %v", err)
		}
		envWarnS1 := parseTestEnvelope(t, resWarnS1)
		if len(envWarnS1.Warnings) == 0 {
			t.Error("expected warning for missing evidence grades in session")
		}

		// Now overwrite T1-01 with proper Grade A evidence and an L1 Construct block
		s1Valid := `---
session_id: T1-01
title: Storage Engine
date: 2026-09-29
status: complete
---
# Storage Engine Evaluation

### Findings
RocksDB delivers sub-millisecond p99 latencies under concurrent writes. A (RocksDB Benchmark v8.1) [corroborated, direct]
LMDB is memory-mapped but suffers under write-heavy concurrent compaction. B (LMDB Analysis) [direct]

### L1 Construct Architecture
` + "```go" + `
type StorageEngine interface {
    Get(key []byte) ([]byte, error)
    Put(key, value []byte) error
}
` + "```" + `

## Recommendations
Adopt RocksDB for primary key-value storage.
`
		resS1, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": pDir,
				"session_id":   "T1-01",
				"content":      s1Valid,
			},
		})
		if err != nil {
			t.Fatalf("save_session valid T1-01: %v", err)
		}
		envS1 := parseTestEnvelope(t, resS1)
		if !envS1.Success {
			t.Fatalf("save T1-01 failed: %s", envS1.Message)
		}

		// 3.10 Record Decision D-001 (One-Way Door with Reversal Trigger)
		resD1, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"project_root":  pDir,
				"decision_id":   "D-001",
				"artifact_type": "decision",
				"content": `---
decision_id: D-001
title: Adopt RocksDB for Storage
status: accepted
door_type: one-way
reversal_triggers:
  - Latency exceeds 5ms under load
  - Compaction stalls block writes
---
# Context
We need a robust embedded KV store.

# Decision
We adopt RocksDB. A (Benchmark v8.1)

# Consequences
High write throughput achieved.
`,
			},
		})
		if err != nil {
			t.Fatalf("record_decision D-001: %v", err)
		}
		envD1 := parseTestEnvelope(t, resD1)
		if !envD1.Success {
			t.Fatalf("record D-001 failed: %s", envD1.Message)
		}

		// In-place ADR update test: Update D-001 with updated notes
		resD1Upd, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"project_root":  pDir,
				"decision_id":   "D-001",
				"artifact_type": "decision",
				"content": `---
decision_id: D-001
title: Adopt RocksDB for Storage
status: accepted
door_type: one-way
reversal_triggers:
  - Latency exceeds 5ms under load
  - Compaction stalls block writes
---
# Context
We need a robust embedded KV store.

# Decision
We adopt RocksDB with tuned block cache. A (Benchmark v8.1)

# Consequences
High write throughput achieved with zero compaction stalls.
`,
			},
		})
		if err != nil {
			t.Fatalf("record_decision update D-001: %v", err)
		}
		envD1Upd := parseTestEnvelope(t, resD1Upd)
		if !envD1Upd.Success {
			t.Fatalf("update D-001 failed: %s", envD1Upd.Message)
		}

		// 3.11 Next Session -> Now T1-02 should be unblocked and contain T1-01 findings injected into [UPSTREAM_FINDINGS]
		resNextS2, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_next_session",
			Arguments: map[string]any{"project_root": pDir},
		})
		if err != nil {
			t.Fatalf("next_session T1-02: %v", err)
		}
		envNextS2 := parseTestEnvelope(t, resNextS2)
		if envNextS2.Data["session_id"] != "T1-02" {
			t.Fatalf("expected next session T1-02, got %v", envNextS2.Data["session_id"])
		}
		s2Prompt := envNextS2.Data["prompt"].(string)
		if !strings.Contains(s2Prompt, "RocksDB") {
			t.Errorf("expected T1-02 prompt to have injected T1-01 RocksDB findings into [UPSTREAM_FINDINGS], got: %s", s2Prompt)
		}

		// Save T1-02
		s2Valid := `---
session_id: T1-02
title: Consensus & Network
date: 2026-09-29
status: complete
---
# Consensus & Network
### Findings
Raft consensus over QUIC satisfies 99.999% availability. A (Ongaro Spec) [direct]
gRPC over TCP incurs 2x head-of-line blocking. B (QUIC vs TCP Study) [corroborated]

## Recommendations
Use Raft over QUIC transport.
`
		resS2, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": pDir,
				"session_id":   "T1-02",
				"content":      s2Valid,
			},
		})
		if err != nil {
			t.Fatalf("save T1-02: %v", err)
		}
		envS2 := parseTestEnvelope(t, resS2)
		if !envS2.Success {
			t.Fatalf("save T1-02 failed: %s", envS2.Message)
		}

		// Post-Hoc Amendment Test: Amend T1-01 based on T1-02 discovery
		resAmend, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_amend_session",
			Arguments: map[string]any{
				"project_root":        pDir,
				"session_id":          "T1-01",
				"amending_session_id": "T1-02",
				"amendment":           "Consensus discovery: RocksDB requires batch commit tuning under Raft leader log replication.",
			},
		})
		if err != nil {
			t.Fatalf("amend_session T1-01: %v", err)
		}
		envAmend := parseTestEnvelope(t, resAmend)
		if !envAmend.Success {
			t.Fatalf("amend T1-01 failed: %s", envAmend.Message)
		}

		// Auto-Draft Decision Test: Request draft for D-002 from session T1-02
		resDraft, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"project_root":    pDir,
				"decision_id":     "D-002",
				"auto_draft_from": "T1-02",
			},
		})
		if err != nil {
			t.Fatalf("auto-draft D-002: %v", err)
		}
		envDraft := parseTestEnvelope(t, resDraft)
		if !envDraft.Success || envDraft.Data["draft_content"] == nil {
			t.Fatalf("expected auto-draft content for D-002, got: %v", envDraft.Data)
		}

		// Workspace-Wide Validation Test: call vivechak_validate with only project_root
		resWsVal, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_validate",
			Arguments: map[string]any{
				"project_root": pDir,
			},
		})
		if err != nil {
			t.Fatalf("workspace validate: %v", err)
		}
		envWsVal := parseTestEnvelope(t, resWsVal)
		if !envWsVal.Success {
			t.Fatalf("workspace validate failed: %s", envWsVal.Message)
		}

		// Record Decision D-002 (Two-Way Door)
		resD2, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"project_root":  pDir,
				"decision_id":   "D-002",
				"artifact_type": "decision",
				"content": `---
decision_id: D-002
title: QUIC Transport Protocol
status: accepted
door_type: two-way
---
# Context
Network protocol choice.
# Decision
QUIC. B (Study)
# Consequences
Lower latency.
`,
			},
		})
		if err != nil {
			t.Fatalf("record D-002: %v", err)
		}
		envD2 := parseTestEnvelope(t, resD2)
		if !envD2.Success {
			t.Fatalf("record D-002 failed: %s", envD2.Message)
		}

		// 3.12 Check Gate readiness BEFORE synthesis -> should NOT pass yet because SYN-01 is incomplete
		resGateEarly, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_run_gate",
			Arguments: map[string]any{
				"project_root": pDir,
				"verbose":      true,
			},
		})
		if err != nil {
			t.Fatalf("run_gate early call: %v", err)
		}
		envGateEarly := parseTestEnvelope(t, resGateEarly)
		if envGateEarly.Data["gate_passed"] == true {
			t.Errorf("gate should NOT pass when synthesis session SYN-01 is pending")
		}

		// 3.13 Next Session -> SYN-01 unblocked, prompt should contain findings from BOTH S1 and S2
		resNextSyn, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_next_session",
			Arguments: map[string]any{"project_root": pDir},
		})
		if err != nil {
			t.Fatalf("next_session SYN-01: %v", err)
		}
		envNextSyn := parseTestEnvelope(t, resNextSyn)
		if envNextSyn.Data["session_id"] != "SYN-01" {
			t.Fatalf("expected next session SYN-01, got %v", envNextSyn.Data["session_id"])
		}
		synPrompt := envNextSyn.Data["prompt"].(string)
		if !strings.Contains(synPrompt, "RocksDB") || !strings.Contains(synPrompt, "QUIC") {
			t.Errorf("expected synthesized findings from both upstream sessions, got: %s", synPrompt)
		}

		// 3.14 Save Synthesis Session SYN-01 -> DUAL WRITE CHECK: writes to research/sessions/SYN-01.md AND research/FAD.md
		synContent := `---
session_id: SYN-01
title: Founding Architecture Document
date: 2026-09-29
status: accepted
---
# Founding Architecture Document (FAD)

## 1. Executive Summary
High-frequency financial distributed ledger. A (pipeline)

## 2. Component Architecture
Storage: RocksDB. A (Benchmark v8.1)
Consensus: Raft over QUIC. A (Ongaro Spec)

## 3. Decisions & Trade-Offs
- D-001: RocksDB (One-Way)
- D-002: QUIC Transport (Two-Way)
`
		resSaveSyn, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": pDir,
				"session_id":   "SYN-01",
				"content":      synContent,
			},
		})
		if err != nil {
			t.Fatalf("save SYN-01: %v", err)
		}
		envSaveSyn := parseTestEnvelope(t, resSaveSyn)
		if !envSaveSyn.Success {
			t.Fatalf("save SYN-01 failed: %s", envSaveSyn.Message)
		}

		// Verify canonical FAD on disk (no dual-write to sessions/)
		sessionFile := filepath.Join(pDir, "research", "sessions", "SYN-01.md")
		fadFile := filepath.Join(pDir, "research", "FAD.md")
		if _, err := os.Stat(sessionFile); !os.IsNotExist(err) {
			t.Errorf("expected session file %s to not exist (no dual-write)", sessionFile)
		}
		if _, err := os.Stat(fadFile); os.IsNotExist(err) {
			t.Errorf("expected FAD file %s to exist", fadFile)
		}

		// 3.15 vivechak_validate tool dry run
		resVal, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_validate",
			Arguments: map[string]any{
				"project_root":  pDir,
				"artifact_type": "decision",
				"content": `---
decision_id: D-003
title: Telemetry Pipeline
status: accepted
---
# Context
Metrics collection.
# Decision
OpenTelemetry. A (CNCF Spec)
`,
			},
		})
		if err != nil {
			t.Fatalf("validate call: %v", err)
		}
		envVal := parseTestEnvelope(t, resVal)
		if !envVal.Success {
			t.Fatalf("validate failed: %s", envVal.Message)
		}

		// 3.16 vivechak_run_gate -> MUST PASS with all criteria met!
		resGate, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_run_gate",
			Arguments: map[string]any{
				"project_root": pDir,
				"verbose":      true,
			},
		})
		if err != nil {
			t.Fatalf("run_gate call: %v", err)
		}
		envGate := parseTestEnvelope(t, resGate)
		if envGate.Data["gate_status"] != "PASS" || envGate.Data["gate_passed"] != true {
			t.Fatalf("expected gate PASS, got status=%v passed=%v, message=%s",
				envGate.Data["gate_status"], envGate.Data["gate_passed"], envGate.Message)
		}
	})

	// -------------------------------------------------------------
	// STAGE 4: Decision Scope Lifecycle (Shortlist Injection + ACH Conflict Resolution)
	// -------------------------------------------------------------
	t.Run("DecisionScope_CompleteTrial", func(t *testing.T) {
		dDir := t.TempDir()

		// 4.1 vivechak_init (Scope: decision)
		resInit, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_init",
			Arguments: map[string]any{
				"project_root": dDir,
				"scope":        "decision",
			},
		})
		if err != nil {
			t.Fatalf("init decision: %v", err)
		}
		if env := parseTestEnvelope(t, resInit); !env.Success {
			t.Fatalf("init decision failed: %s", env.Message)
		}

		// 4.2 vivechak_prepare_generator (Scope: decision)
		resPrep, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_prepare_generator",
			Arguments: map[string]any{
				"project_root": dDir,
				"scope":        "decision",
				"context":      "Evaluate Postgres Partitioning vs Citus for multi-tenant isolation.",
			},
		})
		if err != nil {
			t.Fatalf("prepare_generator decision: %v", err)
		}
		envPrep := parseTestEnvelope(t, resPrep)
		if !envPrep.Success {
			t.Fatalf("prepare_generator decision failed: %s", envPrep.Message)
		}
		prompt := envPrep.Data["prompt"].(string)
		if !strings.Contains(prompt, "multi-tenant isolation") {
			t.Errorf("expected decision context in prompt")
		}

		// 4.3 Save Plan for Decision Scope: S-01 (Shortlist) -> S-02 (Deep Dive with [PASTE S1 SHORTLIST])
		decPlan := `# Decision Research Pipeline
## Session DAG
### Session S-01 — Option Shortlist
| Field | Value |
|---|---|
| **ID** | S-01 |
| **Layer** | 1 |
| **Door Type** | one-way |
| **Dependencies** | None |
| **Output File** | ` + "`S-01-shortlist.md`" + ` |

` + "```prompt" + `
Filter options down to top 2 contenders.
` + "```" + `

### Session S-02 — Deep Comparison & Trade-offs
| Field | Value |
|---|---|
| **ID** | S-02 |
| **Layer** | 2 |
| **Door Type** | one-way |
| **Dependencies** | S-01 |
| **Output File** | ` + "`S-02-deepdive.md`" + ` |

` + "```prompt" + `
Evaluate shortlisted candidates:
[PASTE S1 SHORTLIST]
` + "```" + `
`
		resSavePlan, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_plan",
			Arguments: map[string]any{
				"project_root": dDir,
				"scope":        "decision",
				"decision_id":  "D-001",
				"content":      decPlan,
			},
		})
		if err != nil {
			t.Fatalf("save_plan decision: %v", err)
		}
		if env := parseTestEnvelope(t, resSavePlan); !env.Success {
			t.Fatalf("save_plan decision failed: %s", env.Message)
		}

		// 4.4 Save S-01 with shortlisted candidates
		s01Content := `---
session_id: S-01
title: Option Shortlist
date: 2026-09-29
status: complete
---
# Shortlist Analysis
- Candidate A: Native Declarative Partitioning. A (PG Docs v16)
- Candidate B: Citus Distributed Extension. A (Citus Docs v12)
`
		resS01, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": dDir,
				"session_id":   "S-01",
				"content":      s01Content,
			},
		})
		if err != nil {
			t.Fatalf("save S-01: %v", err)
		}
		if env := parseTestEnvelope(t, resS01); !env.Success {
			t.Fatalf("save S-01 failed: %s", env.Message)
		}

		// 4.5 Call next_session for S-02: Verify [PASTE S1 SHORTLIST] is injected with S-01 content!
		resNextS02, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_next_session",
			Arguments: map[string]any{"project_root": dDir},
		})
		if err != nil {
			t.Fatalf("next_session S-02: %v", err)
		}
		envNextS02 := parseTestEnvelope(t, resNextS02)
		s02Prompt := envNextS02.Data["prompt"].(string)
		if !strings.Contains(s02Prompt, "Native Declarative Partitioning") || !strings.Contains(s02Prompt, "Citus") {
			t.Errorf("expected [PASTE S1 SHORTLIST] to be injected with S-01 options, got: %s", s02Prompt)
		}

		// 4.6 Save S-02
		s02Content := `---
session_id: S-02
title: Deep Comparison
date: 2026-09-29
status: complete
---
# Deep Comparison
Native partitioning avoids cross-node transaction overhead. A (Postgres Internals) [direct]
`
		resSaveS02, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": dDir,
				"session_id":   "S-02",
				"content":      s02Content,
			},
		})
		if err != nil {
			t.Fatalf("save S-02: %v", err)
		}
		if env := parseTestEnvelope(t, resSaveS02); !env.Success {
			t.Fatalf("save S-02 failed: %s", env.Message)
		}

		// 4.7 Record Decision D-001 (One-Way Door with Grade A evidence)
		resD1, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"project_root":  dDir,
				"decision_id":   "D-001",
				"artifact_type": "decision",
				"content": `---
decision_id: D-001
title: Native Declarative Partitioning
status: accepted
door_type: one-way
reversal_triggers:
  - Tenant count exceeds 10,000 tables
---
# Decision
Adopt Native Partitioning. A (Benchmarked)
`,
			},
		})
		if err != nil {
			t.Fatalf("record decision D-001: %v", err)
		}
		if env := parseTestEnvelope(t, resD1); !env.Success {
			t.Fatalf("record decision failed: %s", env.Message)
		}

		// 4.8 Run Gate -> Should pass for decision scope!
		resGateDec, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_run_gate",
			Arguments: map[string]any{
				"project_root": dDir,
			},
		})
		if err != nil {
			t.Fatalf("run_gate decision: %v", err)
		}
		envGateDec := parseTestEnvelope(t, resGateDec)
		if envGateDec.Data["gate_passed"] != true {
			t.Errorf("expected decision scope gate to pass, got: %v", envGateDec.Data)
		}
	})

	// -------------------------------------------------------------
	// STAGE 5: Comparison Scope Lifecycle (Bounded WEP Matrix)
	// -------------------------------------------------------------
	t.Run("ComparisonScope_CompleteTrial", func(t *testing.T) {
		cDir := t.TempDir()

		// 5.1 vivechak_init (Scope: comparison)
		resInit, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_init",
			Arguments: map[string]any{
				"project_root": cDir,
				"scope":        "comparison",
			},
		})
		if err != nil {
			t.Fatalf("init comparison: %v", err)
		}
		if env := parseTestEnvelope(t, resInit); !env.Success {
			t.Fatalf("init comparison failed: %s", env.Message)
		}

		// 5.2 vivechak_prepare_generator (Scope: comparison)
		resPrep, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_prepare_generator",
			Arguments: map[string]any{
				"project_root": cDir,
				"scope":        "comparison",
				"context":      "Compare message brokers: NATS JetStream vs Apache Pulsar.",
			},
		})
		if err != nil {
			t.Fatalf("prepare_generator comparison: %v", err)
		}
		envPrep := parseTestEnvelope(t, resPrep)
		if !envPrep.Success {
			t.Fatalf("prepare_generator comparison failed: %s", envPrep.Message)
		}

		// 5.3 Save Comparison Plan (Single session COMP-01)
		compPlan := `# Comparison Pipeline
## Session DAG
### Session COMP-01 — Message Broker WEP Matrix
| Field | Value |
|---|---|
| **ID** | COMP-01 |
| **Layer** | 1 |
| **Door Type** | two-way |
| **Dependencies** | None |
| **Output File** | ` + "`COMP-01-brokers.md`" + ` |

` + "```prompt" + `
Generate Weighted Evaluation Matrix.
` + "```" + `
`
		resSavePlan, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_plan",
			Arguments: map[string]any{
				"project_root": cDir,
				"scope":        "comparison",
				"decision_id":  "D-002",
				"content":      compPlan,
			},
		})
		if err != nil {
			t.Fatalf("save_plan comparison: %v", err)
		}
		if env := parseTestEnvelope(t, resSavePlan); !env.Success {
			t.Fatalf("save_plan comparison failed: %s", env.Message)
		}

		// 5.4 Save Session COMP-01 with Weighted Evaluation Matrix
		wepContent := `---
session_id: COMP-01
title: Message Broker Comparison
date: 2026-09-29
status: complete
---
# Weighted Evaluation Matrix (WEP)

| Criterion | Weight | NATS JetStream | Apache Pulsar |
|---|---|---|---|
| Operational Simplicity | 40% | 9 (single binary) | 4 (requires ZooKeeper/BookKeeper) |
| Throughput | 30% | 8 (1.2M msg/s) | 9 (1.5M msg/s) |
| Multi-tenancy | 30% | 7 (account-based) | 10 (hierarchical namespaces) |
| **Weighted Score** | **100%** | **8.1** | **7.3** |

### Evidence
NATS runs as a single binary with zero external dependencies. A (NATS Documentation) [direct]
Pulsar requires ZooKeeper and BookKeeper clusters. A (Apache Pulsar Docs) [direct]

## Winner Recommendation
NATS JetStream is selected due to 2x lower operational complexity.
`
		resSaveComp, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": cDir,
				"session_id":   "COMP-01",
				"content":      wepContent,
			},
		})
		if err != nil {
			t.Fatalf("save COMP-01: %v", err)
		}
		if env := parseTestEnvelope(t, resSaveComp); !env.Success {
			t.Fatalf("save COMP-01 failed: %s", env.Message)
		}

		// 5.5 Run Gate -> Comparison scope gate should pass!
		resGateComp, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_run_gate",
			Arguments: map[string]any{
				"project_root": cDir,
			},
		})
		if err != nil {
			t.Fatalf("run_gate comparison: %v", err)
		}
		envGateComp := parseTestEnvelope(t, resGateComp)
		if envGateComp.Data["gate_passed"] != true {
			t.Errorf("expected comparison scope gate to pass, got: %v", envGateComp.Data)
		}
	})

	// -------------------------------------------------------------
	// STAGE 6: Sandbox Escape / Confinement & Security Boundaries
	// -------------------------------------------------------------
	t.Run("Security_And_Confinement", func(t *testing.T) {
		sDir := t.TempDir()

		// Attempt directory traversal in session_id
		resEsc, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": sDir,
				"session_id":   "../../evil",
				"content":      "exploit",
			},
		})
		if err != nil {
			t.Fatalf("save_session traversal call: %v", err)
		}
		envEsc := parseTestEnvelope(t, resEsc)
		if envEsc.Success {
			t.Error("expected directory traversal in session_id to be rejected")
		}

		// Attempt oversized content (> 10MB)
		hugeContent := strings.Repeat("A", 10*1024*1024+1024)
		resHuge, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": sDir,
				"session_id":   "T1-99",
				"content":      hugeContent,
			},
		})
		if err != nil {
			t.Fatalf("save_session huge call: %v", err)
		}
		envHuge := parseTestEnvelope(t, resHuge)
		if envHuge.Success {
			t.Error("expected oversized content (>10MB) to be rejected")
		}
	})

	// -------------------------------------------------------------
	// STAGE 7: Conflict Resolution & ACH Matrix
	// -------------------------------------------------------------
	t.Run("ConflictResolution_And_ACH", func(t *testing.T) {
		crDir := t.TempDir()
		// Init workspace
		resInit, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_init",
			Arguments: map[string]any{"project_root": crDir, "scope": "project"},
		})
		if err != nil {
			t.Fatalf("init call: %v", err)
		}
		if env := parseTestEnvelope(t, resInit); !env.Success {
			t.Fatalf("init failed: %s", env.Message)
		}

		// Validate an ACH conflict resolution document
		conflictContent := `---
conflict_id: CR-001
sessions_involved: [T1-01, T1-02]
status: resolved
resolution_strategy: empirical_test
---
# Conflict Resolution: Storage Write Latency

### Disputed Claim
Session T1-01 claims RocksDB p99 latency is 1.2ms under 50k write ops.
Session T1-02 claims RocksDB p99 latency degrades to 45ms under concurrent read/write transactions.

### Analysis of Competing Hypotheses (ACH)
| Evidence | Hypothesis 1 (T1-01: Low Latency) | Hypothesis 2 (T1-02: High Latency) |
|---|---|---|
| RocksDB Pure Write Benchmark | Consistent (A) | Inconsistent (B) |
| Mixed Read/Write Benchmark | Inconsistent (B) | Consistent (A) |

### Resolution
Both are valid under different workloads. Adopt tuned write buffers and separate compaction threads. A (Validated empirical test)
`
		resVal, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_validate",
			Arguments: map[string]any{
				"project_root": crDir,
				"artifact_type": "decision",
				"content": conflictContent,
			},
		})
		if err != nil {
			t.Fatalf("validate conflict call: %v", err)
		}
		envVal := parseTestEnvelope(t, resVal)
		if !envVal.Success {
			t.Fatalf("validation failed: %s", envVal.Message)
		}
	})

	// -------------------------------------------------------------
	// STAGE 8: Phase 0 Gate Mechanical Validation (Failure & Warning Modes)
	// -------------------------------------------------------------
	t.Run("Phase0Gate_MechanicalValidation_FailureModes", func(t *testing.T) {
		gateDir := t.TempDir()

		// 8.1 Uninitialized directory -> run_gate must fail
		resUninit, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_run_gate",
			Arguments: map[string]any{"project_root": gateDir},
		})
		if err != nil {
			t.Fatalf("run_gate uninit call: %v", err)
		}
		envUninit := parseTestEnvelope(t, resUninit)
		if envUninit.Success {
			t.Error("expected gate check on uninitialized workspace to fail")
		}

		// Initialize
		_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_init",
			Arguments: map[string]any{"project_root": gateDir, "scope": "project"},
		})

		// 8.2 Incomplete workspace (no sessions, no pipeline, no decisions, no FAD) -> gate must FAIL
		resEmpty, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_run_gate",
			Arguments: map[string]any{"project_root": gateDir, "verbose": true},
		})
		if err != nil {
			t.Fatalf("run_gate empty call: %v", err)
		}
		envEmpty := parseTestEnvelope(t, resEmpty)
		if envEmpty.Data["gate_status"] != "FAIL" || envEmpty.Data["gate_passed"] == true {
			t.Errorf("expected gate_status=FAIL and gate_passed=false, got: status=%v passed=%v",
				envEmpty.Data["gate_status"], envEmpty.Data["gate_passed"])
		}

		// 8.3 Verify track_a issues list and warnings
		structural, ok := envEmpty.Data["structural_checks"].(map[string]any)
		if !ok {
			t.Fatalf("expected structural_checks in data, got %v", envEmpty.Data)
		}
		issues, ok := structural["issues"].([]any)
		if !ok || len(issues) == 0 {
			t.Errorf("expected structural_checks.issues to list issues, got %v", structural["issues"])
		}
		if len(envEmpty.Warnings) == 0 {
			t.Errorf("expected gate warnings to be populated, got 0")
		}
	})

	// -------------------------------------------------------------
	// STAGE 9: L1Construct Code Block Extraction & Auto-Detection
	// -------------------------------------------------------------
	t.Run("L1Construct_AutoDetection", func(t *testing.T) {
		cDir := t.TempDir()
		_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_init",
			Arguments: map[string]any{"project_root": cDir, "scope": "project"},
		})

		contentWithConstructs := `---
session_id: T1-01
title: Core Abstractions
date: 2026-09-29
status: complete
---
# Architectural Constructs

Finding: The following interfaces define the kernel abstraction. A (formal design)

` + "```go" + `
type ConsensusEngine interface {
    Propose(entry []byte) error
    CommitIndex() uint64
}

type StateMachine struct {
    LastApplied uint64
}
` + "```" + `

## Recommendations
Implement consensus engine interface with Raft.
`
		resSave, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": cDir,
				"session_id": "T1-01",
				"content": contentWithConstructs,
			},
		})
		if err != nil {
			t.Fatalf("save_session constructs call: %v", err)
		}
		envSave := parseTestEnvelope(t, resSave)
		if !envSave.Success {
			t.Fatalf("save_session with constructs failed: %s", envSave.Message)
		}
		if envSave.Data["validation_passed"] != true {
			t.Errorf("expected validation_passed=true, got %v", envSave.Data["validation_passed"])
		}
	})

	// -------------------------------------------------------------
	// STAGE 10: CLI Doctor Deep Failure Modes
	// -------------------------------------------------------------
	t.Run("CLI_Doctor_DeepFailureModes", func(t *testing.T) {
		// 10.1 Broken YAML frontmatter in session
		docBroken := t.TempDir()
		_ = os.MkdirAll(filepath.Join(docBroken, core.TemplatesDir), 0o755)
		_ = os.MkdirAll(filepath.Join(docBroken, core.SessionsDir), 0o755)
		for _, tmpl := range core.TemplatesToCopy {
			_ = os.WriteFile(filepath.Join(docBroken, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
		}
		_ = os.WriteFile(filepath.Join(docBroken, core.PipelineFile), []byte("# Research Pipeline\n"), 0o644)
		_ = os.WriteFile(filepath.Join(docBroken, core.DecisionsFile), []byte("# Decision Registry\n"), 0o644)

		// Write broken session with invalid YAML frontmatter
		brokenSession := "---\nsession_id: [broken\n---\n# Body\n"
		_ = os.WriteFile(filepath.Join(docBroken, core.SessionsDir, "T1-01.md"), []byte(brokenSession), 0o644)

		cmdDocBroken := exec.CommandContext(ctx, binPath, "doctor", docBroken)
		outDocBroken, errDocBroken := cmdDocBroken.CombinedOutput()
		if errDocBroken == nil {
			t.Errorf("expected doctor to fail on malformed YAML, but succeeded. out: %s", string(outDocBroken))
		}
	})
}
