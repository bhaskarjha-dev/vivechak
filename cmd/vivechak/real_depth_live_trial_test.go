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

// TestDepthLiveMCPServer_EveryFeature exercises every single feature of the Vivechak
// MCP server and CLI up to depth, connecting over live stdio CommandTransport.
func TestDepthLiveMCPServer_EveryFeature(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping depth live MCP trial in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)

	binDir := t.TempDir()
	binName := "vck_depth_test"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build vivechak binary: %v\noutput: %s", err, string(out))
	}

	workDir := t.TempDir()

	// 1. Launch MCP server over stdio
	cmd := exec.CommandContext(ctx, binPath, "serve")
	cmd.Stderr = os.Stderr

	transport := &mcp.CommandTransport{
		Command: cmd,
	}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "depth-trial-client", Version: "0.1.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	// Helper to call tool and parse envelope
	callTool := func(name string, args map[string]any) testEnvelope {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		})
		if err != nil {
			t.Fatalf("CallTool %s failed: %v", name, err)
		}
		if len(res.Content) == 0 {
			t.Fatalf("CallTool %s returned empty content", name)
		}
		tc, ok := res.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatalf("CallTool %s content is not TextContent: %T", name, res.Content[0])
		}
		var env testEnvelope
		if err := json.Unmarshal([]byte(tc.Text), &env); err != nil {
			t.Fatalf("unmarshal envelope for %s: %v (raw: %s)", name, err, tc.Text)
		}
		return env
	}

	// 2. Introspect: Verify all 13 MCP tools registered
	var toolNames []string
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		toolNames = append(toolNames, tool.Name)
	}
	if len(toolNames) != 13 {
		t.Fatalf("expected 13 tools, got %d: %v", len(toolNames), toolNames)
	}

	// Feature 1: vivechak_init
	envInit := callTool("vivechak_init", map[string]any{
		"project_root": workDir,
		"scope":        "project",
	})
	if !envInit.Success {
		t.Fatalf("vivechak_init failed: %s", envInit.Message)
	}
	if _, err := os.Stat(filepath.Join(workDir, "research", ".vivechak.json")); err != nil {
		t.Fatalf("missing .vivechak.json on disk")
	}

	// Feature 2: vivechak_prepare_generator
	envGen := callTool("vivechak_prepare_generator", map[string]any{
		"project_root": workDir,
		"scope":        "project",
		"context":      "Building a distributed zero-knowledge verification pipeline.",
	})
	if !envGen.Success {
		t.Fatalf("vivechak_prepare_generator failed: %s", envGen.Message)
	}
	genPrompt, _ := envGen.Data["prompt"].(string)
	if !strings.Contains(genPrompt, "zero-knowledge verification pipeline") {
		t.Errorf("generator prompt missing injected context")
	}

	// Feature 3: vivechak_save_plan
	pipelinePlan := `# Research Pipeline: ZK Pipeline

**Complexity Score:** 18/25 => Tier 2 (8 sessions)

### T0-01: Proving System Architecture
| **ID** | T0-01 |
| **Door Type** | one-way |
| **Dependencies** | None |

` + "```prompt" + `
Investigate Halo2 vs Plonky3 for fast proof generation.
` + "```" + `

### T1-01: Verifier On-Chain Gas Optimization
| **ID** | T1-01 |
| **Dependencies** | T0-01 |

` + "```prompt" + `
Investigate EVM verifier contract gas consumption.
` + "```" + `

### SYN-01: Founding Architecture Synthesis
| **ID** | SYN-01 |
| **Dependencies** | T1-01 |

` + "```prompt" + `
Compile comprehensive Founding Architecture Document.
` + "```" + `
`
	decisionsRegistry := `# Architectural Decisions

<!-- DECISION: D-001 -->
`
	envPlan := callTool("vivechak_save_plan", map[string]any{
		"project_root":      workDir,
		"content":           pipelinePlan,
		"decisions_content": decisionsRegistry,
	})
	if !envPlan.Success {
		t.Fatalf("vivechak_save_plan failed: %s", envPlan.Message)
	}

	// Feature 4: vivechak_status
	envStatus := callTool("vivechak_status", map[string]any{
		"project_root": workDir,
	})
	if !envStatus.Success {
		t.Fatalf("vivechak_status failed: %s", envStatus.Message)
	}
	if !strings.Contains(envStatus.NextStep, "vivechak_next_session") {
		t.Errorf("status next_step should point to vivechak_next_session, got: %s", envStatus.NextStep)
	}

	// Feature 5: vivechak_visualize
	envViz := callTool("vivechak_visualize", map[string]any{
		"project_root": workDir,
	})
	if !envViz.Success {
		t.Fatalf("vivechak_visualize failed: %s", envViz.Message)
	}
	mermaidCode, _ := envViz.Data["mermaid"].(string)
	if !strings.Contains(mermaidCode, "graph TD") || !strings.Contains(mermaidCode, "T0-01") {
		t.Errorf("visualize missing mermaid graph: %s", mermaidCode)
	}

	// Feature 6: vivechak_next_session
	envNext := callTool("vivechak_next_session", map[string]any{
		"project_root": workDir,
	})
	if !envNext.Success {
		t.Fatalf("vivechak_next_session failed: %s", envNext.Message)
	}
	if envNext.Data["session_id"] != "T0-01" {
		t.Fatalf("expected next session T0-01, got: %v", envNext.Data["session_id"])
	}
	nextPrompt, _ := envNext.Data["prompt"].(string)
	if !strings.Contains(nextPrompt, "Halo2 vs Plonky3") {
		t.Errorf("prompt missing brief content: %s", nextPrompt)
	}
	if !strings.Contains(nextPrompt, "RESEARCH CALIBRATION") {
		t.Errorf("prompt missing research calibration rules injection: %s", nextPrompt)
	}

	// Feature 7: vivechak_validate (dry-run check)
	sessT0Content := `---
session_id: T0-01
title: Proving System Architecture
status: complete
date: 2026-10-07
door_type: one-way
---

## Prior
We initially assumed Halo2 would be fastest.

# Key Findings
- Finding 1 [Grade A] https://github.com/Plonky3/Plonky3 proves 10x faster recursion
- Finding 2 [Grade A] https://eprint.iacr.org/2022/1216 formal security proof
- Finding 3 [Grade B] https://benchmark.zk.org microbenchmarks
- Finding 4 [Grade B] community audits

## Discovered Concerns
Recursion verification overhead on constrained proving hardware.

## Delta
Contradicted initial assumption that Halo2 was the fastest proving system.

## Evaluated Options
- Plonky3: selected.
- Halo2: rejected due to latency.

## Recommendation
We select Plonky3 for recursive proving over Halo2. Grade A (https://github.com/Plonky3/Plonky3)
`
	envVal := callTool("vivechak_validate", map[string]any{
		"project_root":  workDir,
		"artifact_type": "session",
		"content":       sessT0Content,
	})
	if !envVal.Success {
		t.Fatalf("vivechak_validate failed: %s", envVal.Message)
	}

	// Feature 8: vivechak_save_session
	envSaveT0 := callTool("vivechak_save_session", map[string]any{
		"project_root": workDir,
		"session_id":   "T0-01",
		"content":      sessT0Content,
	})
	if !envSaveT0.Success {
		t.Fatalf("vivechak_save_session T0-01 failed: %s", envSaveT0.Message)
	}
	if !strings.Contains(envSaveT0.NextStep, "vivechak_challenge") {
		t.Errorf("one-way door save_session should route to vivechak_challenge, got: %s", envSaveT0.NextStep)
	}

	// Feature 9: vivechak_replan (mid-flight mutation)
	envReplan := callTool("vivechak_replan", map[string]any{
		"project_root": workDir,
		"operation":    "add_session",
		"session_id":   "T1-02",
		"topic":        "Memory Benchmark Spike",
		"layer":        "1",
		"dependencies": "T0-01",
		"prompt":       "Benchmark memory usage during high-degree polynomial FFT.",
	})
	if !envReplan.Success {
		t.Fatalf("vivechak_replan failed: %s", envReplan.Message)
	}

	// Feature 10: vivechak_challenge (red-team probe)
	envChal := callTool("vivechak_challenge", map[string]any{
		"project_root": workDir,
		"session_id":   "T0-01",
		"mode":         "red_team",
	})
	if !envChal.Success {
		t.Fatalf("vivechak_challenge failed: %s", envChal.Message)
	}
	if !strings.Contains(envChal.Message, "challenge prompts") {
		t.Errorf("expected challenge prompts generated, got: %s", envChal.Message)
	}

	// Feature 11: vivechak_record_decision
	d1Content := `# Architectural Decision: D-001 Proving System
---
id: D-001
title: Proving System Selection
status: accepted
door_type: one-way
date: 2026-10-07
human_reviewed: true
informed_by_sessions:
  - T0-01
review_trigger: "Plonky3 audit discovers soundess defect"
review_date: 2026-12-01
---

# Context
High-throughput ZK verification architecture requires sub-second recursive proofs.

# Decision
Adopt Plonky3 for proof generation. Grade A (https://github.com/Plonky3/Plonky3)

## Rejected Alternatives
- Halo2: Rejected due to 5x higher proving latency on GPU.
- Groth16: Rejected due to trusted setup requirements.

# Consequences
Requires customized FRI polynomial commitment integration.
`
	envRecD1 := callTool("vivechak_record_decision", map[string]any{
		"project_root": workDir,
		"decision_id":  "D-001",
		"content":      d1Content,
	})
	if !envRecD1.Success {
		t.Fatalf("vivechak_record_decision D-001 failed: %s", envRecD1.Message)
	}

	// Feature 12: Supersede decision D-001 with D-002, verifying NO duplicate frontmatter
	d2Content := `---
id: D-002
title: Plonky3 with Custom FRI Parameters
status: accepted
door_type: one-way
date: 2026-10-07
human_reviewed: true
informed_by_sessions:
  - T0-01
review_trigger: "FRI parameters exceed 200KB proof size"
review_date: 2026-12-01
---

# Context
Upgraded proving system with custom blowup factor.

# Decision
Adopt Plonky3 with tuned FRI parameters. Grade A (https://github.com/Plonky3/Plonky3)

## Rejected Alternatives
- Standard Plonky3: Rejected due to oversized proofs.
`
	envRecD2 := callTool("vivechak_record_decision", map[string]any{
		"project_root": workDir,
		"decision_id":  "D-002",
		"supersedes":   "D-001",
		"content":      d2Content,
	})
	if !envRecD2.Success {
		t.Fatalf("supersede D-001 with D-002 failed: %s", envRecD2.Message)
	}

	// Verify old D-001 has no duplicate frontmatter fences
	d1File := filepath.Join(workDir, "research", "D-001-decision.md")
	d1Bytes, err := os.ReadFile(d1File)
	if err != nil {
		t.Fatalf("read %s: %v", d1File, err)
	}
	if strings.Count(string(d1Bytes), "---") != 2 {
		t.Fatalf("duplicate frontmatter fence bug detected in D-001.md! Content:\n%s", string(d1Bytes))
	}
	if !strings.Contains(string(d1Bytes), "status: superseded") {
		t.Errorf("D-001 status not updated to superseded")
	}

	// Feature 13: vivechak_amend_session
	envAmend := callTool("vivechak_amend_session", map[string]any{
		"project_root": workDir,
		"session_id":   "T0-01",
		"amendment":    "### Amendment 2026-10-07\nVerified Plonky3 recursion overhead with AVX-512 acceleration.",
	})
	if !envAmend.Success {
		t.Fatalf("vivechak_amend_session failed: %s", envAmend.Message)
	}

	// Save remaining sessions: T1-01, T1-02, and SYN-01
	sessT1_01 := `---
session_id: T1-01
title: Verifier On-Chain Gas Optimization
status: complete
date: 2026-10-07
---

# Key Findings
- Finding 1 [Grade A] https://eips.ethereum.org/EIPS/eip-2537 BLS precompile
- Finding 2 [Grade A] https://eips.ethereum.org/EIPS/eip-4844 blob storage
- Finding 3 [Grade B] gas estimation benchmarks

## Discovered Concerns
EVM memory expansion costs during batch proof verification.

## Delta
Contradicted assumption that calldata was cheaper than blobs.

## Recommendation
Use EIP-4844 blobs for proof submission. Grade A (https://eips.ethereum.org/EIPS/eip-4844)
`
	callTool("vivechak_save_session", map[string]any{
		"project_root": workDir,
		"session_id":   "T1-01",
		"content":      sessT1_01,
	})

	sessT1_02 := `---
session_id: T1-02
title: Memory Benchmark Spike
status: complete
date: 2026-10-07
---

# Key Findings
- Finding 1 [Grade A] https://github.com/Plonky3/Plonky3 memory profiler
- Finding 2 [Grade B] heap profiling telemetry
- Finding 3 [Grade B] swap usage under 64GB RAM

## Discovered Concerns
OOM errors on 32GB instances when batch size exceeds 128.

## Delta
Contradicted assumption that 32GB RAM was sufficient.

## Recommendation
Mandate 64GB RAM minimum for worker nodes. Grade A (https://github.com/Plonky3/Plonky3)
`
	callTool("vivechak_save_session", map[string]any{
		"project_root": workDir,
		"session_id":   "T1-02",
		"content":      sessT1_02,
	})

	// SYN-01 (Founding Architecture Synthesis)
	sessSYN := `---
session_id: SYN-01
title: Founding Architecture Document
status: complete
date: 2026-10-07
---

# Founding Architecture Document
Comprehensive ZK verification pipeline specification.
All components verified with Grade A evidence (https://github.com/Plonky3/Plonky3).
Verifier contracts integrate EIP-4844 (https://eips.ethereum.org/EIPS/eip-4844).
`
	envSaveSYN := callTool("vivechak_save_session", map[string]any{
		"project_root": workDir,
		"session_id":   "SYN-01",
		"content":      sessSYN,
	})
	if !envSaveSYN.Success {
		t.Fatalf("save SYN-01 failed: %s", envSaveSYN.Message)
	}

	// Verify root FOUNDING-ARCHITECTURE.md mirror was created
	rootFADFile := filepath.Join(workDir, "FOUNDING-ARCHITECTURE.md")
	if _, err := os.Stat(rootFADFile); err != nil {
		t.Fatalf("expected mirrored root FOUNDING-ARCHITECTURE.md, err: %v", err)
	}

	// Feature 14: vivechak_run_gate
	envGate := callTool("vivechak_run_gate", map[string]any{
		"project_root": workDir,
	})
	if !envGate.Success {
		t.Fatalf("vivechak_run_gate failed: %s", envGate.Message)
	}
	gateData := envGate.Data
	gatePassed, _ := gateData["gate_passed"].(bool)
	if !gatePassed || gateData["gate_status"] != "PASS" {
		t.Fatalf("expected Gate PASS, got status=%v passed=%v (warnings: %v)", gateData["gate_status"], gatePassed, envGate.Warnings)
	}

	// Verify PHASE-0-GATE.md was auto-persisted to disk
	gateArtifact := filepath.Join(workDir, "research", "PHASE-0-GATE.md")
	gateBytes, err := os.ReadFile(gateArtifact)
	if err != nil {
		t.Fatalf("reading PHASE-0-GATE.md: %v", err)
	}
	if !strings.Contains(string(gateBytes), "Phase 0 Exit Gate") {
		t.Errorf("gate artifact missing title")
	}

	// 3. Test CLI subcommands on the completed workspace
	// CLI doctor
	docCmd := exec.CommandContext(ctx, binPath, "doctor", workDir)
	docOut, err := docCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("vck doctor failed on valid workspace: %v\noutput: %s", err, string(docOut))
	}
	if !strings.Contains(string(docOut), "All checks passed") && !strings.Contains(string(docOut), "healthy") && !strings.Contains(string(docOut), "PASS") {
		t.Logf("vck doctor output: %s", string(docOut))
	}

	// CLI version
	verCmd := exec.CommandContext(ctx, binPath, "version")
	verOut, err := verCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("vck version failed: %v", err)
	}
	if !strings.Contains(string(verOut), "0.1.0") {
		t.Errorf("expected 0.1.0 in version output, got: %s", string(verOut))
	}

	// CLI mcp-config
	cfgCmd := exec.CommandContext(ctx, binPath, "mcp-config")
	cfgOut, err := cfgCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("vck mcp-config failed: %v", err)
	}
	if !strings.Contains(string(cfgOut), "mcpServers") || !strings.Contains(string(cfgOut), "vivechak") {
		t.Errorf("expected JSON mcpServers config, got: %s", string(cfgOut))
	}
}
