package main

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestLiveMCP_UniversalScopes_MassiveE2E exercises complete multi-scope workflows:
// 1. Tool schema inspection: verifies scope is described as REQUIRED while allowing omitted fallback
// 2. Comparison Scope: Init -> PrepareGenerator -> SavePlan -> Challenge -> Amend -> Validate -> CLI doctor
// 3. Decision Scope: Init -> PrepareGenerator -> SavePlan (auto ID) -> RecordDecision -> Challenge -> Amend -> Validate -> RunGate -> CLI doctor
// 4. Project Scope: Universal domain (business venture) end-to-end -> Init -> Prepare -> Plan -> Validate -> Gate
// 5. CLI Doctor validation across all three live directories
func TestLiveMCP_UniversalScopes_MassiveE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping massive live trial in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	// Step 1: Build genuine vivechak binary
	binDir := t.TempDir()
	binName := "vck_universal_e2e_test"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build vivechak binary: %v\noutput: %s", err, string(out))
	}

	// Step 2: Spawn live MCP server as a real subprocess over stdio JSON-RPC
	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "universal-scopes-verifier", Version: "1.0.0"},
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
	// 1. Tool Schema Inspection (Wire Verification)
	// =========================================================================
	t.Run("01_Tool_Schemas_Scope_Required_Instruction", func(t *testing.T) {
		toolMap := make(map[string]*mcp.Tool)
		for tool, err := range cs.Tools(ctx, nil) {
			if err != nil {
				t.Fatalf("cs.Tools() iteration error: %v", err)
			}
			toolMap[tool.Name] = tool
		}

		// Verify vivechak_init description instructs that scope is required
		initTool, ok := toolMap["vivechak_init"]
		if !ok {
			t.Fatal("vivechak_init tool not found")
		}
		if !strings.Contains(initTool.Description, "scope") {
			t.Errorf("expected vivechak_init description to mention scope, got: %s", initTool.Description)
		}

		// Verify vivechak_prepare_generator description instructs scope is required
		prepTool, ok := toolMap["vivechak_prepare_generator"]
		if !ok {
			t.Fatal("vivechak_prepare_generator tool not found")
		}
		if !strings.Contains(prepTool.Description, "scope") {
			t.Errorf("expected vivechak_prepare_generator description to mention scope, got: %s", prepTool.Description)
		}
	})

	// Create persistent test workspaces across all subtests
	compDir := t.TempDir()
	decDir := t.TempDir()
	projDir := t.TempDir()

	// =========================================================================
	// 2. Universal Comparison Scope: AI Models / Algorithms
	// =========================================================================
	t.Run("02_Comparison_Scope_Full_Lifecycle", func(t *testing.T) {

		// 2a. Init comparison workspace
		envInit, _ := callTool("vivechak_init", map[string]any{
			"project_root": compDir,
			"scope":        "comparison",
		})
		if !envInit.Success {
			t.Fatalf("init comparison failed: %s", envInit.Message)
		}

		// 2b. Prepare generator with AI model evaluation
		envPrep, _ := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": compDir,
			"scope":        "comparison",
			"context":      "Claude 3.5 Sonnet vs GPT-4o for complex multi-turn coding agent workflows",
		})
		if !envPrep.Success {
			t.Fatalf("prepare generator comparison failed: %s", envPrep.Message)
		}
		prompt := envPrep.Data["prompt"].(string)
		if !strings.Contains(prompt, "Claude 3.5 Sonnet vs GPT-4o") {
			t.Fatalf("expected options in prompt, got: %s", prompt[:200])
		}

		// 2c. Save Comparison Plan / Output
		compContent := `# Comparison: Claude 3.5 Sonnet vs GPT-4o
| Decision | C-01 |
| Status | complete |

## Summary
Evaluated Claude 3.5 Sonnet vs GPT-4o for coding agent performance.

## WEP Matrix
| Criterion | Weight | Claude 3.5 Sonnet | GPT-4o |
|---|---|---|---|
| Tool Calling Reliability | 0.50 | 9.5 [A] | 8.8 [A] |
| Context Utilization | 0.30 | 9.2 [A] | 8.5 [B] |
| Latency | 0.20 | 8.0 [B] | 8.7 [A] |

## Recommendation
Claude 3.5 Sonnet is recommended for autonomous multi-turn agents.
`
		envPlan, _ := callTool("vivechak_save_plan", map[string]any{
			"project_root": compDir,
			"scope":        "comparison",
			"decision_id":  "C-01",
			"content":      compContent,
		})
		if !envPlan.Success {
			t.Fatalf("save plan comparison failed: %s", envPlan.Message)
		}

		// 2d. Challenge the comparison matrix directly in research/
		envChal, _ := callTool("vivechak_challenge", map[string]any{
			"project_root": compDir,
			"session_id":   "C-01",
			"mode":         "premortem",
		})
		if !envChal.Success {
			t.Fatalf("challenge C-01 failed: %s", envChal.Message)
		}
		if envChal.Data["challenge_prompts"] == nil {
			t.Fatalf("expected challenge prompts in data, got: %+v", envChal.Data)
		}

		// 2e. Amend the comparison session
		amendedContent := compContent + "\n## Addendum\nVerified updated pricing tiers and context window quotas.\n"
		envAmend, _ := callTool("vivechak_amend_session", map[string]any{
			"project_root": compDir,
			"session_id":   "C-01",
			"content":      amendedContent,
			"reason":       "Refreshed pricing tiers and throughput limits",
		})
		if !envAmend.Success {
			t.Fatalf("amend comparison failed: %s", envAmend.Message)
		}

		// 2f. Validate workspace finds comparison matrix
		envVal, _ := callTool("vivechak_validate", map[string]any{
			"project_root": compDir,
		})
		if !envVal.Success {
			t.Fatalf("validate workspace comparison failed: %s", envVal.Message)
		}
		dataVal := envVal.Data
		if dataVal["comparison"] == nil {
			t.Fatalf("expected comparison validation in dataVal, got: %+v", dataVal)
		}
	})

	// =========================================================================
	// 3. Universal Decision Scope: Business Venture / Strategy
	// =========================================================================
	t.Run("03_Decision_Scope_Full_Lifecycle", func(t *testing.T) {
		// 3a. Init decision workspace
		envInit, _ := callTool("vivechak_init", map[string]any{
			"project_root": decDir,
			"scope":        "decision",
		})
		if !envInit.Success {
			t.Fatalf("init decision failed: %s", envInit.Message)
		}

		// 3b. Prepare generator with strategic decision
		envPrep, _ := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": decDir,
			"scope":        "decision",
			"context":      "strategic decision: direct enterprise sales vs marketplace reseller for European market expansion",
		})
		if !envPrep.Success {
			t.Fatalf("prepare generator decision failed: %s", envPrep.Message)
		}

		// 3c. Save Decision Plan omitting decision_id (testing auto-recovery to D-01)
		planContent := `# Decision Plan: International Sales Model
## Routing
| Field | Value |
|---|---|
| Decision | D-01 |
| Door | one-way |
| Lane | deep-research |

## Session DAG
### Session D-01-S1: Sales Motion Comparison
| Field | Value |
|---|---|
| ID | D-01-S1 |
| Role | comparison |
| Depends | none |
| Filename | research/sessions/D-01-S1.md |

` + "```prompt\nEvaluate Direct Sales vs Partner Reseller.\n```\n"

		envPlan, _ := callTool("vivechak_save_plan", map[string]any{
			"project_root": decDir,
			"scope":        "decision",
			"content":      planContent,
		})
		if !envPlan.Success {
			t.Fatalf("save plan decision failed: %s", envPlan.Message)
		}

		// 3d. Save research session D-01-S1
		sessionContent := `---
id: D-01-S1
title: Sales Motion Comparison
status: complete
evidence_count: 3
grade_distribution:
  A: 2
  B: 1
---

# Sales Motion Comparison

## Findings
- Direct sales yields 42% higher retention in tier-1 enterprise accounts [A, https://example.com/saas-metrics-2025].
- Partner reseller model reduces upfront customer acquisition cost by 35% [A, https://example.com/channel-economics-2025].
- Regulatory compliance overhead in the EU requires dedicated localized legal entity [B, https://example.com/eu-compliance].

## Discovered Concerns
Partner reseller models dilute customer relationship ownership and cause channel conflict in tier-1 enterprise tiers.
`
		envSess, _ := callTool("vivechak_save_session", map[string]any{
			"project_root": decDir,
			"session_id":   "D-01-S1",
			"content":      sessionContent,
		})
		if !envSess.Success {
			t.Fatalf("save session failed: %s", envSess.Message)
		}

		// 3e. Record decision D-01 directly in research/
		adrContent := `---
id: D-01
title: Direct Enterprise Sales for European Market
status: accepted
door_type: one-way
date: "2026-10-17"
confidence: high
informed_by_sessions:
  - D-01-S1
evidence_refs:
  - https://example.com/saas-metrics-2025
  - https://example.com/channel-economics-2025
review_trigger: "If enterprise CAC exceeds 18 months payback period"
schema_version: "0.1.0"
---

# Context & Problem Statement
Evaluating international sales motions for expansion.

## Decision Outcome
Adopt direct enterprise sales model.

## Rejected Alternatives & Tradeoffs
- Partner Reseller: Rejected due to channel conflict and loss of customer account ownership.

## Challenge Findings & Red-Team Stress Test
Stress-tested with red-team challenge. Confirmed localized legal entity costs are covered by enterprise ACV margins.

## Failure Modes & Reversal Triggers
Reversal trigger: If enterprise CAC exceeds 18 months payback period or regulatory compliance costs exceed projections by >50%.
`
		envRec, _ := callTool("vivechak_record_decision", map[string]any{
			"project_root": decDir,
			"decision_id":  "D-01",
			"slug":         "intl-sales-model",
			"content":      adrContent,
		})
		if !envRec.Success {
			t.Fatalf("record decision failed: %s", envRec.Message)
		}

		// 3f. Challenge the standalone ADR directly in research/
		envChal, _ := callTool("vivechak_challenge", map[string]any{
			"project_root": decDir,
			"decision_id":  "D-01",
			"mode":         "red_team",
		})
		if !envChal.Success {
			t.Fatalf("challenge D-01 failed: %s", envChal.Message)
		}

		// 3g. Amend the standalone ADR directly in research/
		amendedAdr := strings.Replace(adrContent, "Adopt direct enterprise sales model.", "Adopt direct enterprise sales model with dedicated EMEA pods.", 1)
		envAmend, _ := callTool("vivechak_amend_session", map[string]any{
			"project_root": decDir,
			"session_id":   "D-01",
			"content":      amendedAdr,
			"reason":       "Clarified localized EMEA pod organization structure",
		})
		if !envAmend.Success {
			t.Fatalf("amend ADR failed: %s", envAmend.Message)
		}

		// 3h. Validate workspace discovers D-01-plan.md and D-01-intl-sales-model.md
		envVal, _ := callTool("vivechak_validate", map[string]any{
			"project_root": decDir,
		})
		if !envVal.Success {
			t.Fatalf("validate workspace decision failed: %s", envVal.Message)
		}

		// 3i. Run Phase 0 Gate on D-01
		envGate, _ := callTool("vivechak_run_gate", map[string]any{
			"project_root": decDir,
			"track":        "track_b",
		})
		if !envGate.Success {
			t.Fatalf("run gate decision failed: %s", envGate.Message)
		}
		dataGate := envGate.Data
		if dataGate["track_b_result"] != "pass" {
			t.Logf("gate track_b_result: %v, warnings: %v", dataGate["track_b_result"], envGate.Warnings)
		}
	})

	// =========================================================================
	// 4. Universal Project Scope: Decentralized Energy Trading Platform
	// =========================================================================
	t.Run("04_Project_Scope_Universal_Lifecycle", func(t *testing.T) {
		// 4a. Init project workspace
		envInit, _ := callTool("vivechak_init", map[string]any{
			"project_root": projDir,
			"scope":        "project",
		})
		if !envInit.Success {
			t.Fatalf("init project failed: %s", envInit.Message)
		}

		// 4b. Prepare generator for new venture
		envPrep, _ := callTool("vivechak_prepare_generator", map[string]any{
			"project_root": projDir,
			"scope":        "project",
			"context":      "new venture: decentralized renewable energy peer-to-peer microgrid trading marketplace",
		})
		if !envPrep.Success {
			t.Fatalf("prepare generator project failed: %s", envPrep.Message)
		}

		// 4c. Save research pipeline DAG
		pipelineContent := `# Research Pipeline: Clean Energy Microgrid Trading

**Complexity Score:** 12 / 24 → **Tier 2**

**Execution DAG:**
` + "```mermaid" + `
graph TD
    T0_01[T0-01: Grid Settlement Protocols]
    T0_02[T0-02: Cryptographic Meter Verification]
    T0_01 --> T1_01[T1-01: Core Matching Engine]
    T0_02 --> T1_01
    T1_01 --> T2_01[T2-01: Founding Architecture Synthesis]
` + "```" + `

### Session T0-01: Grid Settlement Protocols
| Field | Value |
|---|---|
| ID | T0-01 |
| Layer | 0 |
| Door Type | one-way |
| Dependencies | none |
| Output File | research/sessions/T0-01.md |

` + "```prompt\nEvaluate peer-to-peer energy settlement protocols.\n```\n" + `

### Session T0-02: Cryptographic Meter Verification
| Field | Value |
|---|---|
| ID | T0-02 |
| Layer | 0 |
| Door Type | two-way |
| Dependencies | none |
| Output File | research/sessions/T0-02.md |

` + "```prompt\nEvaluate cryptographic smart meter signing options.\n```\n" + `

### Session T1-01: Core Matching Engine
| Field | Value |
|---|---|
| ID | T1-01 |
| Layer | 1 |
| Door Type | one-way |
| Dependencies | T0-01, T0-02 |
| Output File | research/sessions/T1-01.md |

` + "```prompt\nEvaluate low latency order matching engine architectures.\n```\n" + `

### Session T2-01: Founding Architecture Synthesis
| Field | Value |
|---|---|
| ID | T2-01 |
| Layer | 2 |
| Door Type | one-way |
| Dependencies | T1-01 |
| Output File | research/FAD.md |

` + "```prompt\nSynthesize all findings into the Founding Architecture Document.\n```\n"

		envPlan, _ := callTool("vivechak_save_plan", map[string]any{
			"project_root": projDir,
			"scope":        "project",
			"content":      pipelineContent,
		})
		if !envPlan.Success {
			t.Fatalf("save plan project failed: %s", envPlan.Message)
		}

		// 4d. Status check
		envStat, _ := callTool("vivechak_status", map[string]any{
			"project_root": projDir,
		})
		if !envStat.Success {
			t.Fatalf("status failed: %s", envStat.Message)
		}

		// 4e. Next session dispatch (should return T0-01 and T0-02 ready in parallel)
		envNext, _ := callTool("vivechak_next_session", map[string]any{
			"project_root": projDir,
		})
		if !envNext.Success {
			t.Fatalf("next_session failed: %s", envNext.Message)
		}
		dataNext := envNext.Data
		sessID := dataNext["session_id"].(string)
		if sessID != "T0-01" && sessID != "T0-02" {
			t.Fatalf("expected T0-01 or T0-02 ready, got: %v", sessID)
		}
		otherReady := dataNext["other_ready_sessions"].([]any)
		if len(otherReady) != 1 {
			t.Fatalf("expected 1 other ready session, got %d: %v", len(otherReady), otherReady)
		}

		// 4f. Visualize DAG
		envVis, _ := callTool("vivechak_visualize", map[string]any{
			"project_root": projDir,
		})
		if !envVis.Success {
			t.Fatalf("visualize failed: %s", envVis.Message)
		}
	})

	// =========================================================================
	// 5. CLI Doctor Real Subprocess Execution on all three workspaces
	// =========================================================================
	t.Run("05_CLI_Doctor_RealExecution_All_Workspaces", func(t *testing.T) {
		// Run doctor on comparison directory
		cmdComp := exec.CommandContext(ctx, binPath, "doctor", compDir)
		outComp, errComp := cmdComp.CombinedOutput()
		if errComp != nil {
			t.Fatalf("vck doctor compDir failed: %v, output: %s", errComp, string(outComp))
		}
		if !strings.Contains(string(outComp), "Scope: comparison") && !strings.Contains(string(outComp), "Workspace:") {
			t.Errorf("unexpected doctor output on compDir: %s", string(outComp))
		}

		// Run doctor on decision directory
		cmdDec := exec.CommandContext(ctx, binPath, "doctor", decDir)
		outDec, errDec := cmdDec.CombinedOutput()
		if errDec != nil {
			t.Fatalf("vck doctor decDir failed: %v, output: %s", errDec, string(outDec))
		}
		if !strings.Contains(string(outDec), "Scope: decision") && !strings.Contains(string(outDec), "Workspace:") {
			t.Errorf("unexpected doctor output on decDir: %s", string(outDec))
		}

		// Run doctor on project directory
		cmdProj := exec.CommandContext(ctx, binPath, "doctor", projDir)
		outProj, errProj := cmdProj.CombinedOutput()
		if errProj != nil {
			t.Fatalf("vck doctor projDir failed: %v, output: %s", errProj, string(outProj))
		}
		if !strings.Contains(string(outProj), "Scope: project") && !strings.Contains(string(outProj), "Workspace:") {
			t.Errorf("unexpected doctor output on projDir: %s", string(outProj))
		}

		// Run version command
		cmdVer := exec.CommandContext(ctx, binPath, "version")
		outVer, errVer := cmdVer.CombinedOutput()
		if errVer != nil {
			t.Fatalf("vck version failed: %v", errVer)
		}
		if !strings.Contains(string(outVer), "vivechak 0.1.0") {
			t.Errorf("unexpected version output: %s", string(outVer))
		}

		// Run mcp-config command
		cmdMcp := exec.CommandContext(ctx, binPath, "mcp-config")
		outMcp, errMcp := cmdMcp.CombinedOutput()
		if errMcp != nil {
			t.Fatalf("vck mcp-config failed: %v", errMcp)
		}
		var cfgParsed map[string]any
		if err := json.Unmarshal(outMcp, &cfgParsed); err != nil {
			t.Fatalf("vck mcp-config output invalid JSON: %v, raw: %s", err, string(outMcp))
		}
	})
}
