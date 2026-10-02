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

// TestRealMassive_E2E_AllFeatures performs an end-to-end operational test against the live
// compiled vivechak binary using real filesystem I/O and real stdio MCP communication.
func TestRealMassive_E2E_AllFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)

	// 1. Build the live binary
	binDir := t.TempDir()
	binName := "vivechak_real_test"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	// 2. Connect to live MCP server over stdio
	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "real-test-client", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server over stdio: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	// 3. Verify tools/list has exactly 10 tools with expected schemas
	var toolNames []string
	toolMap := make(map[string]*mcp.Tool)
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("error listing tools: %v", err)
		}
		toolNames = append(toolNames, tool.Name)
		toolMap[tool.Name] = tool
	}

	if len(toolNames) != 10 {
		t.Fatalf("expected 10 tools from live server, got %d: %v", len(toolNames), toolNames)
	}

	expectedTools := []string{
		"vivechak_init",
		"vivechak_prepare_generator",
		"vivechak_save_plan",
		"vivechak_status",
		"vivechak_next_session",
		"vivechak_save_session",
		"vivechak_record_decision",
		"vivechak_amend_session",
		"vivechak_validate",
		"vivechak_run_gate",
	}
	for _, expected := range expectedTools {
		if _, ok := toolMap[expected]; !ok {
			t.Errorf("missing expected tool: %s", expected)
		}
	}

	// 4. Create real project workspace
	projectDir := t.TempDir()

	// 4.1 vivechak_init
	resInit, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": projectDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("vivechak_init failed: %v", err)
	}
	envInit := parseTestEnvelope(t, resInit)
	if !envInit.Success {
		t.Fatalf("vivechak_init message: %s", envInit.Message)
	}

	// Verify research/.gitignore exists and has *.lock
	giPath := filepath.Join(projectDir, core.ResearchDir, ".gitignore")
	giData, err := os.ReadFile(giPath)
	if err != nil {
		t.Fatalf("reading .gitignore: %v", err)
	}
	if !strings.Contains(string(giData), "*.lock") {
		t.Errorf("expected .gitignore to contain *.lock, got:\n%s", string(giData))
	}

	// 4.2 vivechak_prepare_generator
	resPrep, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_prepare_generator",
		Arguments: map[string]any{
			"project_root": projectDir,
			"scope":        "project",
			"context":      "Building high-frequency financial distributed ledger with audit guarantees.",
		},
	})
	if err != nil {
		t.Fatalf("prepare_generator failed: %v", err)
	}
	envPrep := parseTestEnvelope(t, resPrep)
	if !envPrep.Success {
		t.Fatalf("prepare_generator error: %s", envPrep.Message)
	}
	prompt := envPrep.Data["prompt"].(string)
	if !strings.Contains(prompt, "financial distributed ledger") {
		t.Errorf("expected context injected into generator prompt")
	}

	// 4.3 vivechak_save_plan: 3 layers: Layer 0 (Storage), Layer 1 (Consensus), Layer 2 (Synthesis)
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
	resPlan, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": projectDir,
			"content":      planContent,
		},
	})
	if err != nil {
		t.Fatalf("save_plan failed: %v", err)
	}
	envPlan := parseTestEnvelope(t, resPlan)
	if !envPlan.Success {
		t.Fatalf("save_plan failed: %s", envPlan.Message)
	}

	// 4.4 vivechak_status
	resStat, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_status",
		Arguments: map[string]any{"project_root": projectDir},
	})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	envStat := parseTestEnvelope(t, resStat)
	if envStat.Data["session_count"].(float64) != 0 {
		t.Errorf("expected 0 completed sessions, got %v", envStat.Data["session_count"])
	}

	// 4.5 vivechak_next_session for S-01 (prompt > 10K tokens should be truncated since verbose=false)
	resNext1, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": projectDir, "verbose": false},
	})
	if err != nil {
		t.Fatalf("next_session 1 failed: %v", err)
	}
	envNext1 := parseTestEnvelope(t, resNext1)
	if envNext1.Data["session_id"] != "S-01" {
		t.Fatalf("expected next session S-01, got %v", envNext1.Data["session_id"])
	}
	p1 := envNext1.Data["prompt"].(string)
	if !strings.Contains(p1, "[truncated") {
		t.Errorf("expected regular session prompt > 10K tokens to be truncated when verbose=false")
	}

	// 4.6 Save Session S-01 with Recommendations, Alternatives, Discovered Concerns, and Delta
	s1Content := `---
session_id: S-01
title: Storage Engine
date: 2026-10-01
status: complete
---
# Storage Engine Evaluation

## Evaluated Options & Alternatives
1. RocksDB: LSM-tree KV store. Grade A (direct benchmark v8.1)
2. LMDB: Memory-mapped B-tree. Grade B (docs analysis)

## Recommendations
We recommend RocksDB for the primary persistence layer. Grade A (empirical benchmark)

## Discovered Concerns & Failure Modes
- Compaction stalls during high concurrent ingestion.
- WAL write amplification under random writes.

## Delta
| Prior | Status | Impact |
|---|---|---|
| LMDB would be faster | Contradicted | High |
`
	resSave1, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": projectDir,
			"session_id":   "S-01",
			"content":      s1Content,
		},
	})
	if err != nil {
		t.Fatalf("save S-01 failed: %v", err)
	}
	envSave1 := parseTestEnvelope(t, resSave1)
	if !envSave1.Success {
		t.Fatalf("save S-01 unsuccessful: %s", envSave1.Message)
	}

	// 4.7 Post-Hoc Amendment: Amend S-01
	resAmend, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_amend_session",
		Arguments: map[string]any{
			"project_root":        projectDir,
			"session_id":          "S-01",
			"amending_session_id": "PRE-02",
			"amendment":           "Preliminary review noted: compaction stalls can be mitigated via dynamic level base size.",
		},
	})
	if err != nil {
		t.Fatalf("amend S-01 failed: %v", err)
	}
	envAmend := parseTestEnvelope(t, resAmend)
	if !envAmend.Success {
		t.Fatalf("amend S-01 unsuccessful: %s", envAmend.Message)
	}

	// Verify on-disk amendment
	s1Path := filepath.Join(projectDir, envAmend.Data["file_path"].(string))
	s1OnDisk, err := os.ReadFile(s1Path)
	if err != nil {
		t.Fatalf("reading amended S-01 from %s: %v", s1Path, err)
	}
	if !strings.Contains(string(s1OnDisk), "## Post-Hoc Amendment (appended by PRE-02)") {
		t.Errorf("expected amendment header on disk: %s", string(s1OnDisk))
	}
	if !strings.Contains(string(s1OnDisk), "compaction stalls can be mitigated") {
		t.Errorf("expected amendment body on disk")
	}

	// 4.8 Auto-Draft Decision D-001 from S-01
	resDraft, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":    projectDir,
			"decision_id":     "D-001",
			"auto_draft_from": "S-01",
		},
	})
	if err != nil {
		t.Fatalf("auto-draft D-001 failed: %v", err)
	}
	envDraft := parseTestEnvelope(t, resDraft)
	if !envDraft.Success || envDraft.Data["draft_content"] == nil {
		t.Fatalf("expected auto-draft content for D-001, got: %v", envDraft.Data)
	}
	draftContent := envDraft.Data["draft_content"].(string)
	if !strings.Contains(draftContent, "RocksDB") {
		t.Errorf("expected RocksDB in draft: %s", draftContent)
	}
	if !strings.Contains(draftContent, "Compaction stalls") {
		t.Errorf("expected failure modes in draft: %s", draftContent)
	}

	// 4.9 Record Decision D-001 with one-way door and human_reviewed: true
	d1Content := `---
id: D-001
title: Adopt RocksDB as Primary Storage
status: accepted
door_type: one-way
human_reviewed: true
review_trigger: "Write stalls exceed 2%"
review_date: 2026-12-01
---
# Context
We need a robust embedded KV store for financial ledger records.

# Decision
We adopt RocksDB. Grade A (empirical benchmark v8.1)

# Consequences
High write throughput achieved. Dynamic level base size used for compaction.
`
	resRec1, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": projectDir,
			"decision_id":  "D-001",
			"content":      d1Content,
		},
	})
	if err != nil {
		t.Fatalf("record D-001 failed: %v", err)
	}
	envRec1 := parseTestEnvelope(t, resRec1)
	if !envRec1.Success {
		t.Fatalf("record D-001 unsuccessful: %s", envRec1.Message)
	}

	// Verify DECISIONS.md contains collapsible <details> block
	decPath := filepath.Join(projectDir, core.DecisionsFile)
	decDisk, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("reading DECISIONS.md: %v", err)
	}
	if !strings.Contains(string(decDisk), "<details>") || !strings.Contains(string(decDisk), "<strong>D-001</strong>") {
		t.Errorf("expected collapsible details block in DECISIONS.md, got:\n%s", string(decDisk))
	}

	// 4.10 Layer Transition: Layer 0 complete -> next_session S-02 (Layer 1)
	resNext2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": projectDir},
	})
	if err != nil {
		t.Fatalf("next_session 2 failed: %v", err)
	}
	envNext2 := parseTestEnvelope(t, resNext2)
	if envNext2.Data["session_id"] != "S-02" {
		t.Fatalf("expected next session S-02, got %v", envNext2.Data["session_id"])
	}

	// Verify layer replan advisory warning was triggered
	hasLayerTransition := false
	for _, w := range envNext2.Warnings {
		if strings.Contains(w, "LAYER-TRANSITION") && strings.Contains(w, "Layer 0 complete") {
			hasLayerTransition = true
			break
		}
	}
	if !hasLayerTransition {
		t.Errorf("expected LAYER-TRANSITION advisory in warnings, got: %v", envNext2.Warnings)
	}

	// Verify Upstream Concerns are injected into S-02 prompt
	p2 := envNext2.Data["prompt"].(string)
	if !strings.Contains(p2, "Upstream Concerns (stress-test these)") {
		t.Errorf("expected Upstream Concerns in S-02 prompt, got: %s", p2)
	}
	if !strings.Contains(p2, "Compaction stalls") {
		t.Errorf("expected compaction stalls concern in S-02 prompt")
	}

	// 4.11 Save Session S-02 with findings and Delta
	s2Content := `---
session_id: S-02
title: Consensus Protocol
date: 2026-10-01
status: complete
---
# Consensus Findings

## Recommendations
We recommend Raft with pipelined quorum over TLS. Grade A (formal TLA+ spec)

## Delta
| Prior | Status | Impact |
|---|---|---|
| Paxos was required | Contradicted | Medium |
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": projectDir,
			"session_id":   "S-02",
			"content":      s2Content,
		},
	})

	// 4.12 Synthesis Session SYN-01: Exemption from truncation & Delta Aggregation
	resSyn, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": projectDir, "verbose": false},
	})
	if err != nil {
		t.Fatalf("next_session SYN-01 failed: %v", err)
	}
	envSyn := parseTestEnvelope(t, resSyn)
	if envSyn.Data["session_id"] != "SYN-01" {
		t.Fatalf("expected SYN-01, got %v", envSyn.Data["session_id"])
	}

	synPrompt := envSyn.Data["prompt"].(string)
	// Must NOT be truncated despite verbose=false
	if strings.Contains(synPrompt, "[truncated") {
		t.Errorf("synthesis prompt should NOT be truncated when verbose=false!")
	}
	// Must contain aggregated belief evolution
	if !strings.Contains(synPrompt, "BELIEF EVOLUTION ACROSS SESSIONS") {
		t.Errorf("expected BELIEF EVOLUTION ACROSS SESSIONS in synthesis prompt")
	}
	if !strings.Contains(synPrompt, "LMDB would be faster") || !strings.Contains(synPrompt, "Paxos was required") {
		t.Errorf("expected Delta tables aggregated from both sessions in synthesis prompt")
	}
	// Must contain NextStep synthesis quality guidance
	if !strings.Contains(envSyn.NextStep, "Belief Evolution section") || !strings.Contains(envSyn.NextStep, "strongest and weakest signals") {
		t.Errorf("expected NextStep synthesis quality guidance, got: %s", envSyn.NextStep)
	}

	// 4.13 Save Synthesis FAD
	fadContent := `---
session_id: FAD
title: Founding Architecture Document
date: 2026-10-01
status: complete
---
# Founding Architecture Document

## 1. System Vision
Distributed high-frequency financial ledger.

## 2. Core Architecture
RocksDB storage layer with Raft consensus. Grade A (direct benchmark)

## 3. Storage & Consensus Selection
RocksDB selected over LMDB based on sub-millisecond p99 benchmarks. Grade A (empirical data)
Raft selected over Paxos for operational simplicity and pipelined quorums. Grade A (spec analysis)

## 4. Consequences & Operational Model
Level-based compaction tuned to avoid stalls. Monitored continuously. Grade B (operational playbook)

## 5. Rejected Alternatives
- LMDB rejected due to write concurrency bottleneck. Grade A (benchmark)
- Paxos rejected due to implementation complexity. Grade B (RFC analysis)
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": projectDir,
			"session_id":   "SYN-01",
			"content":      fadContent,
		},
	})

	// Copy FAD to research/FAD.md for standard gate verification
	fadTarget := filepath.Join(projectDir, core.FADFile)
	_ = os.WriteFile(fadTarget, []byte(fadContent), 0o644)

	// 4.14 Workspace-Level Validation
	resWsVal, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_validate",
		Arguments: map[string]any{
			"project_root": projectDir,
		},
	})
	if err != nil {
		t.Fatalf("workspace validate failed: %v", err)
	}
	envWsVal := parseTestEnvelope(t, resWsVal)
	if !envWsVal.Success {
		t.Fatalf("workspace validate unsuccessful: %s", envWsVal.Message)
	}
	dataWsVal := envWsVal.Data
	sessionsVal := dataWsVal["sessions"].(map[string]any)
	if sessionsVal["valid"].(float64) < 2 {
		t.Errorf("expected at least 2 valid sessions, got: %v", sessionsVal)
	}
	decisionsVal := dataWsVal["decisions"].(map[string]any)
	if decisionsVal["errors"].(float64) != 0 {
		t.Errorf("expected 0 decision errors, got: %v", decisionsVal)
	}

	// 4.15 Phase 0 Gate Execution
	resGate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": projectDir,
			"verbose":      true,
		},
	})
	if err != nil {
		t.Fatalf("run_gate failed: %v", err)
	}
	envGate := parseTestEnvelope(t, resGate)
	if envGate.Data["gate_status"] != "PASS" || envGate.Data["gate_passed"] != true {
		t.Fatalf("expected gate PASS, got status=%v passed=%v warnings=%v message=%s",
			envGate.Data["gate_status"], envGate.Data["gate_passed"], envGate.Warnings, envGate.Message)
	}

	// 5. Test CLI Doctor on the completed workspace
	doctorCmd := exec.CommandContext(ctx, binPath, "doctor", projectDir)
	docOut, docErr := doctorCmd.CombinedOutput()
	if docErr != nil {
		t.Fatalf("doctor failed on completed workspace: %v\nOutput: %s", docErr, string(docOut))
	}
	docStr := string(docOut)
	if !strings.Contains(docStr, "✓ Workspace exists") {
		t.Errorf("expected doctor to pass workspace check, got:\n%s", docStr)
	}
}
