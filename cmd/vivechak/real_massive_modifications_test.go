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

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMassiveLiveMCP_AllNewTools_And_ModifiedProtocols runs an end-to-end live stdio
// MCP integration test verifying all recent enhancements:
// 1. Full 13-tool roster discovery and annotations (challenge, replan, visualize).
// 2. Modern comparison namespace (C-01) with auto-ID extraction in save_plan.
// 3. In-flight DAG mutations via vivechak_replan (add, update prompt, update deps, cycle prevention, remove).
// 4. DAG visualization via vivechak_visualize (Mermaid graph with colored styling and status table).
// 5. Adversarial challenge generation via vivechak_challenge (red_team, evidence_audit, cross_session).
// 6. ADR supersession protocol and stale reference detection (W-STALE-DECISION-REFERENCE).
// 7. Phase 0 exit gate precision guardrails and CLI doctor workspace health verification.
func TestMassiveLiveMCP_AllNewTools_And_ModifiedProtocols(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	// 1. Build live binary
	binDir := t.TempDir()
	binName := "vck_massive_mod_trial"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	// 2. Launch child process and connect over stdio MCP
	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "massive-mod-client", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	callAndParse := func(toolName string, args map[string]any) testEnvelope {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		})
		if err != nil {
			t.Fatalf("call to %s failed: %v", toolName, err)
		}
		return parseTestEnvelope(t, res)
	}

	// =========================================================================
	// PHASE 1: TOOL DISCOVERY & ANNOTATIONS (ALL 13 TOOLS)
	// =========================================================================
	t.Run("ToolDiscovery_And_Annotations", func(t *testing.T) {
		var toolNames []string
		toolMap := make(map[string]*mcp.Tool)
		for tool, err := range cs.Tools(ctx, nil) {
			if err != nil {
				t.Fatalf("error listing tools: %v", err)
			}
			toolNames = append(toolNames, tool.Name)
			toolMap[tool.Name] = tool
		}

		if len(toolNames) != 13 {
			t.Fatalf("expected exactly 13 tools registered, got %d: %v", len(toolNames), toolNames)
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
			"vivechak_challenge",
			"vivechak_replan",
			"vivechak_visualize",
		}

		for _, name := range expectedTools {
			tool, ok := toolMap[name]
			if !ok {
				t.Fatalf("expected tool %s not found in registered tools", name)
			}
			if tool.Description == "" {
				t.Errorf("tool %s has empty description", name)
			}
		}

		// Verify tool annotations for new tools
		challengeTool := toolMap["vivechak_challenge"]
		if challengeTool.Annotations == nil || !challengeTool.Annotations.ReadOnlyHint {
			t.Errorf("vivechak_challenge expected ReadOnlyHint == true")
		}

		visualizeTool := toolMap["vivechak_visualize"]
		if visualizeTool.Annotations == nil || !visualizeTool.Annotations.ReadOnlyHint {
			t.Errorf("vivechak_visualize expected ReadOnlyHint == true")
		}

		replanTool := toolMap["vivechak_replan"]
		if replanTool.Annotations == nil || replanTool.Annotations.ReadOnlyHint {
			t.Errorf("vivechak_replan expected ReadOnlyHint == false")
		}
	})

	// =========================================================================
	// PHASE 2: MODERN COMPARISON NAMESPACE (C-01) & AUTO-ID EXTRACTION
	// =========================================================================
	t.Run("ModernComparisonNamespace_C01_AutoIDExtraction", func(t *testing.T) {
		cmpDir := t.TempDir()

		// 1. Init comparison workspace
		envInit := callAndParse("vivechak_init", map[string]any{
			"project_root": cmpDir,
			"scope":        "comparison",
		})
		if !envInit.Success {
			t.Fatalf("init comparison failed: %s", envInit.Message)
		}

		// 2. Save plan without decision_id argument (exercises auto-ID extraction)
		cmpPlan := `# Comparison Plan: C-01

| Field | Value |
|---|---|
| **ID** | C-01 |
| **Output File** | sessions/C-01.md |

` + "```prompt\n# Brief\nCompare RocksDB vs BadgerDB for embedded key-value storage.\n```\n"

		envPlan := callAndParse("vivechak_save_plan", map[string]any{
			"project_root": cmpDir,
			"scope":        "comparison",
			"content":      cmpPlan,
		})
		if !envPlan.Success {
			t.Fatalf("save comparison plan failed: %s", envPlan.Message)
		}

		// Verify plan saved on disk as C-01-comparison.md
		planPath := filepath.Join(cmpDir, "research", "C-01-comparison.md")
		if _, err := os.Stat(planPath); err != nil {
			t.Fatalf("expected C-01-comparison.md to exist on disk: %v", err)
		}

		// 3. Next session returns C-01
		envNext := callAndParse("vivechak_next_session", map[string]any{
			"project_root": cmpDir,
		})
		if envNext.Data["session_id"] != "C-01" {
			t.Fatalf("expected next session to be C-01, got: %v", envNext.Data["session_id"])
		}

		// 4. Save session with WEP matrix and Grade A evidence
		wepContent := `---
session_id: C-01
title: Embedded Key-Value Storage Comparison
date: 2026-10-05
status: complete
---
# Embedded Key-Value Storage Comparison

## Recommendation
Adopt RocksDB due to superior write throughput and production maturity. Grade A (https://rocksdb.org)

## Weighted Evaluation Matrix (WEP)
| Criterion | Weight | RocksDB | BadgerDB |
|---|---|---|---|
| Throughput | 40% | 9 (Grade A) | 7 (Grade B) |
| Memory Footprint | 30% | 8 (Grade A) | 8 (Grade B) |
| Operational Simplicity | 30% | 7 (Grade B) | 9 (Grade A) |
| **Weighted Score** | 100% | **8.1** | **7.9** |

## Alternatives Considered
- BadgerDB: Pure Go implementation, rejected due to higher SSD write amplification under high ingestion. Grade A (https://dgraph.io/badger)
`
		envSave := callAndParse("vivechak_save_session", map[string]any{
			"project_root": cmpDir,
			"session_id":   "C-01",
			"content":      wepContent,
		})
		if !envSave.Success {
			t.Fatalf("save C-01 session failed: %s", envSave.Message)
		}

		// 5. Run gate on comparison workspace
		envGate := callAndParse("vivechak_run_gate", map[string]any{
			"project_root": cmpDir,
		})
		if !envGate.Success {
			t.Fatalf("run_gate on comparison workspace failed: %s", envGate.Message)
		}
	})

	// =========================================================================
	// PHASE 3: LIVE DAG MUTATION WITH VIVECHAK_REPLAN
	// =========================================================================
	t.Run("Replan_InFlightDAGMutations_And_CycleGuards", func(t *testing.T) {
		projDir := t.TempDir()

		// 1. Init project workspace
		callAndParse("vivechak_init", map[string]any{
			"project_root": projDir,
			"scope":        "project",
		})

		// 2. Save initial 3-session pipeline (T0-01 -> T1-01 -> SYN-01)
		initialPipeline := `# Research Pipeline: Cloud Platform

| ID | Topic | Dependencies | Output File |
|---|---|---|---|
| T0-01 | Architecture Landscape | none | sessions/T0-01.md |
| T1-01 | Database Layer | T0-01 | sessions/T1-01.md |
| SYN-01 | Architecture Synthesis | T1-01 | sessions/SYN-01.md |

---

### Session T0-01: Architecture Landscape
| **Field** | **Value** |
|---|---|
| **ID** | T0-01 |
| **Layer** | 0 |
| **Dependencies** | none |
| **Output** | sessions/T0-01.md |

` + "```prompt\n# Brief\nResearch Landscape.\n```\n\n" + `### Session T1-01: Database Layer
| **Field** | **Value** |
|---|---|
| **ID** | T1-01 |
| **Layer** | 1 |
| **Dependencies** | T0-01 |
| **Output** | sessions/T1-01.md |

` + "```prompt\n# Brief\nResearch Database.\n```\n\n" + `### Session SYN-01: Architecture Synthesis
| **Field** | **Value** |
|---|---|
| **ID** | SYN-01 |
| **Layer** | 2 |
| **Dependencies** | T1-01 |
| **Output** | sessions/SYN-01.md |

` + "```prompt\n# Brief\nSynthesize FAD.\n```\n"

		callAndParse("vivechak_save_plan", map[string]any{
			"project_root": projDir,
			"content":      initialPipeline,
		})

		// 3. Mutate DAG: add_session T1-02 (Cache Layer) depending on T0-01
		envAdd := callAndParse("vivechak_replan", map[string]any{
			"project_root": projDir,
			"operation":    "add_session",
			"session_id":   "T1-02",
			"topic":        "Distributed Cache Layer",
			"layer":        "1",
			"dependencies": "T0-01",
			"door_type":    "two-way",
			"prompt":       "# Brief\nEvaluate Redis vs Dragonfly caching.",
		})
		if !envAdd.Success {
			t.Fatalf("replan add_session failed: %s", envAdd.Message)
		}

		// Verify RESEARCH-PIPELINE.md on disk contains T1-02
		pipeBytes, _ := os.ReadFile(filepath.Join(projDir, "research", "RESEARCH-PIPELINE.md"))
		if !strings.Contains(string(pipeBytes), "T1-02") {
			t.Fatalf("RESEARCH-PIPELINE.md does not contain added session T1-02")
		}

		// 4. Mutate DAG: update_prompt of T1-02
		envUpPrompt := callAndParse("vivechak_replan", map[string]any{
			"project_root": projDir,
			"operation":    "update_prompt",
			"session_id":   "T1-02",
			"new_prompt":   "# Brief\nDeep dive Dragonfly memory concurrency vs Redis Cluster.",
		})
		if !envUpPrompt.Success {
			t.Fatalf("replan update_prompt failed: %s", envUpPrompt.Message)
		}

		// 5. Mutate DAG: update_deps of SYN-01 to depend on both T1-01 and T1-02
		envUpDeps := callAndParse("vivechak_replan", map[string]any{
			"project_root":     projDir,
			"operation":        "update_deps",
			"session_id":       "SYN-01",
			"new_dependencies": "T1-01, T1-02",
		})
		if !envUpDeps.Success {
			t.Fatalf("replan update_deps failed: %s", envUpDeps.Message)
		}

		// 6. Test cycle prevention: attempt to add a session that depends on SYN-01, but T0-01 depends on it
		envCycle := callAndParse("vivechak_replan", map[string]any{
			"project_root":     projDir,
			"operation":        "update_deps",
			"session_id":       "T0-01",
			"new_dependencies": "SYN-01",
		})
		if envCycle.Success {
			t.Fatalf("expected cycle to be rejected, but update_deps succeeded")
		}

		// 7. Add throwaway session T1-03 and remove it
		callAndParse("vivechak_replan", map[string]any{
			"project_root": projDir,
			"operation":    "add_session",
			"session_id":   "T1-03",
			"topic":        "Message Queue",
			"dependencies": "T0-01",
		})
		envRem := callAndParse("vivechak_replan", map[string]any{
			"project_root": projDir,
			"operation":    "remove_session",
			"session_id":   "T1-03",
		})
		if !envRem.Success {
			t.Fatalf("replan remove_session failed: %s", envRem.Message)
		}
		pipeBytesAfterRem, _ := os.ReadFile(filepath.Join(projDir, "research", "RESEARCH-PIPELINE.md"))
		if strings.Contains(string(pipeBytesAfterRem), "T1-03") {
			t.Fatalf("T1-03 was not removed from RESEARCH-PIPELINE.md")
		}
	})

	// =========================================================================
	// PHASE 4: LIVE MERMAID & TABLE VISUALIZATION WITH VIVECHAK_VISUALIZE
	// =========================================================================
	t.Run("Visualize_MermaidAndTableRepresentations", func(t *testing.T) {
		projDir := t.TempDir()

		callAndParse("vivechak_init", map[string]any{
			"project_root": projDir,
			"scope":        "project",
		})

		pipelineContent := `# Research Pipeline

| ID | Topic | Dependencies | Output File |
|---|---|---|---|
| T0-01 | Ingestion | none | sessions/T0-01.md |
| T1-01 | Storage | T0-01 | sessions/T1-01.md |
| SYN-01 | Synthesis | T1-01 | sessions/SYN-01.md |

---

### Session T0-01: Ingestion
| **ID** | T0-01 |
| **Output** | sessions/T0-01.md |
` + "```prompt\nIngestion prompt\n```\n\n" + `### Session T1-01: Storage
| **ID** | T1-01 |
| **Output** | sessions/T1-01.md |
` + "```prompt\nStorage prompt\n```\n\n" + `### Session SYN-01: Synthesis
| **ID** | SYN-01 |
| **Output** | sessions/SYN-01.md |
` + "```prompt\nSynthesis prompt\n```\n"

		callAndParse("vivechak_save_plan", map[string]any{
			"project_root": projDir,
			"content":      pipelineContent,
		})

		// 1. Mermaid visualization before completion
		envMermaid := callAndParse("vivechak_visualize", map[string]any{
			"project_root": projDir,
			"format":       "mermaid",
		})
		if !envMermaid.Success {
			t.Fatalf("visualize mermaid failed: %s", envMermaid.Message)
		}
		mermaidText, ok := envMermaid.Data["mermaid"].(string)
		if !ok || !strings.Contains(mermaidText, "graph TD") {
			t.Fatalf("expected valid Mermaid graph TD definition, got: %v", envMermaid.Data["mermaid"])
		}
		// Unblocked node T0-01 should be styled with amber (#f59e0b)
		if !strings.Contains(mermaidText, "#f59e0b") {
			t.Errorf("expected ready node T0-01 to have amber styling (#f59e0b)")
		}

		// 2. Table visualization
		envTable := callAndParse("vivechak_visualize", map[string]any{
			"project_root": projDir,
			"format":       "table",
		})
		if !envTable.Success {
			t.Fatalf("visualize table failed: %s", envTable.Message)
		}
		tableText, ok := envTable.Data["table"].(string)
		if !ok || !strings.Contains(tableText, "T0-01") {
			t.Fatalf("expected status table to include T0-01, got: %v", envTable.Data["table"])
		}
	})

	// =========================================================================
	// PHASE 5: RESEARCH & ADVERSARIAL STRESS-TESTING (VIVECHAK_CHALLENGE)
	// =========================================================================
	t.Run("AdversarialChallenge_RedTeam_EvidenceAudit_CrossSession", func(t *testing.T) {
		projDir := t.TempDir()

		callAndParse("vivechak_init", map[string]any{
			"project_root": projDir,
			"scope":        "project",
		})

		// Save session T0-01 with strong recommendation and documented rejected alternatives
		s1 := `---
session_id: T0-01
title: Cloud Datastore Evaluation
date: 2026-10-05
status: complete
door_type: one-way
---
# Cloud Datastore Evaluation

## Recommendation
Deploy CockroachDB for multi-region active-active persistence. Grade A (https://cockroachlabs.com/docs)

## Key Findings
- CockroachDB supports serializable multi-region transactions with 15ms latency overhead. Grade A (https://cockroachlabs.com/docs)
- TiDB requires separate placement driver and TiKV storage daemons. Grade B (benchmarks)

## Alternatives Considered
- Google Cloud Spanner: Proprietary cloud lock-in, rejected due to sovereign hosting requirement. Grade A (https://cloud.google.com/spanner)
- TiDB: Complex multi-tier operational architecture. Grade B (benchmarks)
`
		callAndParse("vivechak_save_session", map[string]any{
			"project_root": projDir,
			"session_id":   "T0-01",
			"content":      s1,
		})

		// 1. Red Team Mode
		envRed := callAndParse("vivechak_challenge", map[string]any{
			"project_root": projDir,
			"session_id":   "T0-01",
			"mode":         "red_team",
		})
		if !envRed.Success {
			t.Fatalf("challenge red_team failed: %s", envRed.Message)
		}
		prompts, ok := envRed.Data["challenge_prompts"].([]any)
		if !ok || len(prompts) < 3 {
			t.Fatalf("expected at least 3 challenge prompts in red_team mode, got: %v", envRed.Data["challenge_prompts"])
		}
		recSnippet, _ := envRed.Data["context"].(map[string]any)["recommendation"].(string)
		if !strings.Contains(recSnippet, "CockroachDB") {
			t.Errorf("expected recommendation context to mention CockroachDB, got: %s", recSnippet)
		}

		// 2. Evidence Audit Mode
		envAudit := callAndParse("vivechak_challenge", map[string]any{
			"project_root": projDir,
			"session_id":   "T0-01",
			"mode":         "evidence_audit",
		})
		if !envAudit.Success {
			t.Fatalf("challenge evidence_audit failed: %s", envAudit.Message)
		}
		auditPrompts, ok := envAudit.Data["challenge_prompts"].([]any)
		if !ok || len(auditPrompts) == 0 {
			t.Fatalf("expected evidence audit prompts, got none")
		}

		// 3. Save additional sessions and run Cross-Session Mode
		s2 := `---
session_id: T1-01
title: Cache Layer Evaluation
date: 2026-10-05
status: complete
---
# Cache Layer Evaluation
## Recommendation
Adopt Dragonfly for in-memory caching. Grade A (https://dragonflydb.io)
`
		s3 := `---
session_id: T1-02
title: Auth Layer Evaluation
date: 2026-10-05
status: complete
---
# Auth Layer Evaluation
## Recommendation
Adopt Ory Kratos for identity management. Grade A (https://ory.sh)
`
		callAndParse("vivechak_save_session", map[string]any{
			"project_root": projDir,
			"session_id":   "T1-01",
			"content":      s2,
		})
		callAndParse("vivechak_save_session", map[string]any{
			"project_root": projDir,
			"session_id":   "T1-02",
			"content":      s3,
		})

		envCross := callAndParse("vivechak_challenge", map[string]any{
			"project_root": projDir,
			"session_id":   "T0-01",
			"mode":         "cross_session",
		})
		if !envCross.Success {
			t.Fatalf("challenge cross_session failed: %s", envCross.Message)
		}
	})

	// =========================================================================
	// PHASE 6: ADR SUPERSESSION & STALE REFERENCE DETECTION
	// =========================================================================
	t.Run("ADRSupersession_And_StaleReferenceDetection", func(t *testing.T) {
		projDir := t.TempDir()

		callAndParse("vivechak_init", map[string]any{
			"project_root": projDir,
			"scope":        "project",
		})

		// 1. Record D-001 (Accepted One-Way Door)
		d1Content := `---
decision_id: D-001
title: Monolithic Architecture
status: accepted
door_type: one-way
review_trigger: Re-evaluate when active developers exceed 25
---
# D-001: Monolithic Architecture

## Context
Initial phase with 3 engineers.

## Decision
Build as modular monolith.
`
		callAndParse("vivechak_record_decision", map[string]any{
			"project_root": projDir,
			"decision_id":  "D-001",
			"content":      d1Content,
		})

		// 2. Save a session that cites D-001
		sSession := `---
session_id: T1-01
title: Scalability Spike
date: 2026-10-05
status: complete
---
# Scalability Spike
This investigation directly refers to decision D-001. We found that the monolith under [D-001] hits database locking limits.
`
		callAndParse("vivechak_save_session", map[string]any{
			"project_root": projDir,
			"session_id":   "T1-01",
			"content":      sSession,
		})

		// 3. Record D-002 superseding D-001
		d2Content := `---
decision_id: D-002
title: Microservices Architecture
status: accepted
door_type: one-way
review_trigger: Re-evaluate when service count exceeds 15
---
# D-002: Microservices Architecture

## Context
Engineering team grew to 40 developers.

## Decision
Split modular monolith into domain microservices.
`
		envSupersede := callAndParse("vivechak_record_decision", map[string]any{
			"project_root": projDir,
			"decision_id":  "D-002",
			"content":      d2Content,
			"supersedes":   "D-001",
		})

		if !envSupersede.Success {
			t.Fatalf("record_decision with supersedes failed: %s", envSupersede.Message)
		}

		// Verify warning emitted for stale reference in T1-01
		hasStaleWarn := false
		for _, w := range envSupersede.Warnings {
			if strings.Contains(w, "W-STALE-DECISION-REFERENCE") && strings.Contains(w, "T1-01") {
				hasStaleWarn = true
				break
			}
		}
		if !hasStaleWarn {
			t.Errorf("expected W-STALE-DECISION-REFERENCE warning mentioning T1-01, got warnings: %v", envSupersede.Warnings)
		}

		// Verify D-001 frontmatter updated to status: superseded, superseded_by: D-002
		d1Bytes, _ := os.ReadFile(filepath.Join(projDir, "research", "D-001-decision.md"))
		d1Str := string(d1Bytes)
		if !strings.Contains(d1Str, "status: superseded") || !strings.Contains(d1Str, "superseded_by: D-002") {
			t.Errorf("D-001 was not properly updated with status: superseded and superseded_by: D-002. Content:\n%s", d1Str)
		}

		// Verify D-002 frontmatter contains supersedes: D-001
		d2Bytes, _ := os.ReadFile(filepath.Join(projDir, "research", "D-002-decision.md"))
		d2Str := string(d2Bytes)
		if !strings.Contains(d2Str, "supersedes: D-001") {
			t.Errorf("D-002 does not contain supersedes: D-001. Content:\n%s", d2Str)
		}

		// Verify consolidated DECISIONS.md registry updated
		regBytes, _ := os.ReadFile(filepath.Join(projDir, "research", "DECISIONS.md"))
		regStr := string(regBytes)
		if !strings.Contains(regStr, "D-001") || !strings.Contains(regStr, "superseded") || !strings.Contains(regStr, "D-002") {
			t.Errorf("DECISIONS.md does not accurately reflect both decisions. Content:\n%s", regStr)
		}
	})

	// =========================================================================
	// PHASE 7: PHASE 0 GATE PRECISION GUARDRAILS & CLI DOCTOR
	// =========================================================================
	t.Run("GateGuardrails_PremortemPlaceholderRejection_And_Doctor", func(t *testing.T) {
		projDir := t.TempDir()

		callAndParse("vivechak_init", map[string]any{
			"project_root": projDir,
			"scope":        "decision",
		})

		// 1. Record decision with unedited template placeholder for review_trigger
		badTriggerADR := `---
decision_id: D-010
title: Event Sourcing Engine
status: accepted
door_type: one-way
review_trigger: "[Condition or date for mandatory re-evaluation]"
---
# D-010: Event Sourcing Engine

## Context
High compliance audit requirements.

## Decision
Implement event sourcing with EventStoreDB.
`
		callAndParse("vivechak_record_decision", map[string]any{
			"project_root": projDir,
			"decision_id":  "D-010",
			"content":      badTriggerADR,
		})

		// Save decision session
		callAndParse("vivechak_save_session", map[string]any{
			"project_root": projDir,
			"session_id":   "D-010-session",
			"content": `---
session_id: D-010-session
title: Event Sourcing Session
date: 2026-10-05
status: complete
---
# Event Sourcing
## Recommendation
Adopt EventStoreDB. Grade A (https://eventstore.com)
`,
		})

		// Run gate: must fail because review_trigger has template placeholder
		envGateFail := callAndParse("vivechak_run_gate", map[string]any{
			"project_root": projDir,
		})
		if envGateFail.Success && strings.Contains(envGateFail.Message, "PASS") {
			t.Fatalf("expected gate to fail due to placeholder review_trigger, but passed: %s", envGateFail.Message)
		}

		// 2. Fix review trigger to specific measurable condition
		goodTriggerADR := `---
decision_id: D-010
title: Event Sourcing Engine
status: accepted
door_type: one-way
review_trigger: Re-evaluate if projection lag exceeds 500ms or storage cost exceeds $10k/month
---
# D-010: Event Sourcing Engine

## Context
High compliance audit requirements.

## Decision
Implement event sourcing with EventStoreDB.
`
		callAndParse("vivechak_record_decision", map[string]any{
			"project_root": projDir,
			"decision_id":  "D-010",
			"content":      goodTriggerADR,
		})

		// Run gate again: should now pass
		envGatePass := callAndParse("vivechak_run_gate", map[string]any{
			"project_root": projDir,
		})
		if !envGatePass.Success || !strings.Contains(envGatePass.Message, "PASS") {
			t.Fatalf("expected gate to pass after fixing review_trigger, got: %s", envGatePass.Message)
		}

		// 3. Test CLI vck doctor command on this workspace
		docCmd := exec.CommandContext(ctx, binPath, "doctor", projDir)
		docOut, docErr := docCmd.CombinedOutput()
		if docErr != nil {
			t.Fatalf("vck doctor returned error: %v\noutput: %s", docErr, string(docOut))
		}
		if !strings.Contains(string(docOut), "Workspace is healthy") {
			t.Errorf("expected doctor to output 'Workspace is healthy', got:\n%s", string(docOut))
		}
	})
}
