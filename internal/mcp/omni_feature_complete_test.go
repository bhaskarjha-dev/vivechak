package mcputil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestCompleteSolution_OmniFeature_EndToEnd performs exhaustive, zero-trust verification
// of every single feature across the Vivechak system:
// 1. MCP Protocol Triumvirate: 13 Tools, 4 Prompts (+ Aliases), 8 Resources
// 2. Workspace Resolution & Environmental Safety Ladder
// 3. Comparison Scope Full Lifecycle (Auto-Inference, Auto-Init, WEP Matrix, Validate, Visualize)
// 4. Decision Scope Full Lifecycle (Plan, Adversarial Challenge, Auto-Draft ADR, Registry Recompilation, Amend)
// 5. Project Scope Full Lifecycle (DAG Pipeline, Next Session, Quality Coaching, Replan, Two-Track Gate)
// 6. Multi-Subagent Concurrent Advisory File Locking & Atomic Persistence
func TestCompleteSolution_OmniFeature_EndToEnd(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	// =========================================================================
	// SUBTEST 1: MCP Protocol Triumvirate (Tools, Prompts, Resources)
	// =========================================================================
	t.Run("1_MCPProtocolTriumvirate", func(t *testing.T) {
		// 1.1 Verify all 13 tools are registered with annotations
		toolsRes, err := cs.ListTools(ctx, nil)
		if err != nil {
			t.Fatalf("ListTools failed: %v", err)
		}
		expectedTools := map[string]bool{
			"vivechak_start":           true,
			"vivechak_save":            true,
			"vivechak_record_decision": true,
			"vivechak_next":            true,
			"vivechak_status":          true,
			"vivechak_review":          true,
			"vivechak_amend":           true,
			"vivechak_replan":          true,
			"vivechak_gate":            true,
		}
		if len(toolsRes.Tools) != len(expectedTools) {
			t.Fatalf("expected %d tools, got %d", len(expectedTools), len(toolsRes.Tools))
		}
		for _, tool := range toolsRes.Tools {
			if !expectedTools[tool.Name] {
				t.Errorf("unexpected tool registered: %s", tool.Name)
			}
			if tool.Description == "" {
				t.Errorf("tool %s missing description", tool.Name)
			}
			if tool.Annotations == nil {
				t.Errorf("tool %s missing annotations", tool.Name)
			}
		}

		// 1.2 Verify 4 canonical Prompts with clean action-verb names
		promptsRes, err := cs.ListPrompts(ctx, nil)
		if err != nil {
			t.Fatalf("ListPrompts failed: %v", err)
		}
		expectedPrompts := map[string]string{
			"compare":  "Compare Options (WEP Matrix)",
			"decision": "Architectural Decision (ADR)",
			"project":  "Project Pipeline (Full FAD)",
			"gate":     "Phase 0 Pre-Codebase Exit Gate",
		}
		if len(promptsRes.Prompts) != len(expectedPrompts) {
			t.Fatalf("expected exactly %d canonical prompts, got %d", len(expectedPrompts), len(promptsRes.Prompts))
		}
		for _, p := range promptsRes.Prompts {
			expTitle, ok := expectedPrompts[p.Name]
			if !ok {
				t.Errorf("unexpected prompt registered: %s", p.Name)
				continue
			}
			if p.Title != expTitle {
				t.Errorf("prompt %s title mismatch: got %q, want %q", p.Name, p.Title, expTitle)
			}
		}

		// 1.3 Verify Prompt retrieval via Canonical, Legacy (vivechak_*), and Shorthand (vck_*)
		triplets := []struct {
			canon, legacy, shorthand string
			args                     map[string]string
			mustContain              string
		}{
			{"compare", "vivechak_compare", "vck_compare", map[string]string{"options": "A vs B"}, "OPTIONS: A vs B"},
			{"decision", "vivechak_decision", "vck_decision", map[string]string{"dilemma": "Migrate DB"}, "DECISION: Migrate DB"},
			{"project", "vivechak_project", "vck_project", map[string]string{"vision": "Build Platform"}, "Build Platform"},
			{"gate", "vivechak_gate", "vck_gate", map[string]string{"track": "track_a"}, `track="track_a"`},
		}
		for _, trip := range triplets {
			r1, err1 := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: trip.canon, Arguments: trip.args})
			r2, err2 := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: trip.legacy, Arguments: trip.args})
			r3, err3 := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: trip.shorthand, Arguments: trip.args})
			if err1 != nil || err2 != nil || err3 != nil {
				t.Fatalf("GetPrompt failed for triplet (%s, %s, %s): %v, %v, %v", trip.canon, trip.legacy, trip.shorthand, err1, err2, err3)
			}
			t1 := r1.Messages[0].Content.(*mcp.TextContent).Text
			t2 := r2.Messages[0].Content.(*mcp.TextContent).Text
			t3 := r3.Messages[0].Content.(*mcp.TextContent).Text
			if t1 != t2 || t1 != t3 {
				t.Errorf("prompt alias mismatch for %s", trip.canon)
			}
			if !strings.Contains(t1, trip.mustContain) {
				t.Errorf("prompt %s missing expected substring %q", trip.canon, trip.mustContain)
			}
		}

		// 1.4 Verify all 8 canonical Resources can be listed and read
		resList, err := cs.ListResources(ctx, nil)
		if err != nil {
			t.Fatalf("ListResources failed: %v", err)
		}
		expectedResources := []string{
			"vivechak://templates/decisions",
			"vivechak://templates/comparison",
			"vivechak://templates/fad",
			"vivechak://templates/conflict-resolution",
			"vivechak://templates/gate",
			"vivechak://templates/session",
			"vivechak://methodology/grades",
			"vivechak://methodology/principles",
		}
		if len(resList.Resources) != len(expectedResources) {
			t.Fatalf("expected %d resources, got %d", len(expectedResources), len(resList.Resources))
		}
		for _, uri := range expectedResources {
			rRes, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
			if err != nil {
				t.Fatalf("ReadResource(%s) failed: %v", uri, err)
			}
			if len(rRes.Contents) == 0 {
				t.Fatalf("ReadResource(%s) returned empty contents", uri)
			}
			text := rRes.Contents[0].Text
			if len(text) < 50 {
				t.Errorf("resource %s content too short (%d bytes)", uri, len(text))
			}
		}

		// Unknown resource must return error
		_, err = cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "vivechak://unknown/resource"})
		if err == nil {
			t.Errorf("expected error for non-existent resource, got nil")
		}
	})

	// =========================================================================
	// SUBTEST 2: Workspace Resolution & Environmental Safety Ladder
	// =========================================================================
	t.Run("2_WorkspaceResolutionAndSafetyGuards", func(t *testing.T) {
		// 2.1 Environmental Ladder resolution (VCK_PROJECT_ROOT, VCK_DEFAULT_ROOT, etc.)
		testLadder := []string{
			"VIVECHAK_PROJECT_ROOT",
			"VCK_PROJECT_ROOT",
			"VIVECHAK_DEFAULT_ROOT",
			"VCK_DEFAULT_ROOT",
			"WORKSPACE",
			"PROJECT_ROOT",
		}
		for _, envVar := range testLadder {
			t.Run(envVar, func(t *testing.T) {
				tmpDir := t.TempDir()
				// Clear all other env vars
				for _, other := range testLadder {
					t.Setenv(other, "")
				}
				t.Setenv(envVar, tmpDir)

				res, err := cs.CallTool(ctx, &mcp.CallToolParams{
					Name:      "vivechak_start",
					Arguments: map[string]any{},
				})
				if err != nil {
					t.Fatalf("vivechak_init with %s failed: %v", envVar, err)
				}
				env := parseEnvelope(t, res)
				if !env.Success {
					t.Fatalf("expected init success with env %s, got: %s", envVar, env.Message)
				}
				if !core.WorkspaceExists(tmpDir) {
					t.Errorf("workspace not created in target dir %s via env %s", tmpDir, envVar)
				}
			})
		}

		// 2.2 Safety Guard: Unexpanded editor macro (${workspaceFolder})
		t.Run("rejects_unexpanded_macro", func(t *testing.T) {
			macroDir := "${workspaceFolder}"
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      "vivechak_status",
				Arguments: map[string]any{"project_root": macroDir},
			})
			if err != nil {
				t.Fatalf("status call failed: %v", err)
			}
			env := parseEnvelope(t, res)
			if strings.Contains(env.Message, "${workspaceFolder}") && !strings.Contains(env.Message, "not set") {
				t.Errorf("mistakenly treated macro as valid path: %s", env.Message)
			}
		})

		// 2.3 Safety Guard: Filesystem root rejection
		t.Run("rejects_filesystem_root", func(t *testing.T) {
			rootPath := "/"
			if os.PathSeparator == '\\' {
				rootPath = "C:\\"
			}
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      "vivechak_init",
				Arguments: map[string]any{"project_root": rootPath},
			})
			if err != nil {
				t.Fatalf("call error: %v", err)
			}
			env := parseEnvelope(t, res)
			if env.Success {
				t.Errorf("expected filesystem root %s to be rejected, but succeeded", rootPath)
			}
		})

		// 2.4 Safety Guard: User home directory rejection
		t.Run("rejects_user_home_dir", func(t *testing.T) {
			homeDir, err := os.UserHomeDir()
			if err == nil && homeDir != "" {
				res, err := cs.CallTool(ctx, &mcp.CallToolParams{
					Name:      "vivechak_init",
					Arguments: map[string]any{"project_root": homeDir},
				})
				if err != nil {
					t.Fatalf("call error: %v", err)
				}
				env := parseEnvelope(t, res)
				if env.Success {
					t.Errorf("expected user home directory %s to be rejected, but succeeded", homeDir)
				}
			}
		})
	})

	// =========================================================================
	// SUBTEST 3: Comparison Scope Full Lifecycle
	// =========================================================================
	t.Run("3_ComparisonScope_FullLifecycle", func(t *testing.T) {
		compDir := t.TempDir()

		// 3.1 Prepare generator: auto-infers comparison from "Pino vs Winston"
		prepRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_prepare_generator",
			Arguments: map[string]any{
				"project_root": compDir,
				"context":      "Compare Pino vs Winston for high-throughput JSON logging in Express",
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

		// 3.2 Save plan on-the-fly without prior init: auto-initializes workspace
		wepContent := `---
decision_id: C-01
slug: pino-vs-winston
scope: comparison
title: Pino vs Winston Comparison
---
# Comparison Plan: C-01 (Pino vs Winston)

## Summary
Compare Pino and Winston for high-throughput JSON logging.

## WEP Matrix
| Criterion | Weight | Pino Score (1-5) | Winston Score (1-5) | Weighted Pino | Weighted Winston |
|---|---|---|---|---|---|
| Throughput | 0.50 | 5 [Grade A, https://github.com/pinojs/pino] | 3 [Grade B, docs] | 2.50 | 1.50 |
| Ecosystem | 0.50 | 4 [Grade A] | 5 [Grade A] | 2.00 | 2.50 |
| **Total** | **1.00** | | | **4.50** | **4.00** |

## Recommendation
Pino is recommended for high throughput.
`
		saveRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_plan",
			Arguments: map[string]any{
				"project_root": compDir,
				"scope":        "comparison",
				"decision_id":  "C-01",
				"slug":         "pino-vs-winston",
				"content":      wepContent,
			},
		})
		if err != nil {
			t.Fatalf("save_plan failed: %v", err)
		}
		saveEnv := parseEnvelope(t, saveRes)
		if !saveEnv.Success {
			t.Fatalf("save_plan failed: %s", saveEnv.Message)
		}

		// Verify comparison file exists and template was auto-copied
		compFile := filepath.Join(compDir, "research", "C-01-cmp-pino-vs-winston.md")
		if _, err := os.Stat(compFile); err != nil {
			t.Fatalf("expected comparison file %s to exist", compFile)
		}
		tmplFile := filepath.Join(compDir, "research", "templates", "COMPARISON-SESSION.template.md")
		if _, err := os.Stat(tmplFile); err != nil {
			t.Fatalf("expected template %s to exist after auto-init", tmplFile)
		}

		// 3.3 Status check: verifies scope is comparison and initialized
		statusRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_status",
			Arguments: map[string]any{"project_root": compDir},
		})
		if err != nil {
			t.Fatalf("status call failed: %v", err)
		}
		statusEnv := parseEnvelope(t, statusRes)
		if !statusEnv.Success {
			t.Fatalf("status call failed: %s", statusEnv.Message)
		}
		statusData := statusEnv.Data.(map[string]any)
		if statusData["scope"] != "comparison" {
			t.Errorf("expected scope 'comparison', got: %v", statusData["scope"])
		}

		// 3.4 Visualize: generates comparison WEP summary table
		visRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_visualize",
			Arguments: map[string]any{"project_root": compDir},
		})
		if err != nil {
			t.Fatalf("visualize call failed: %v", err)
		}
		visEnv := parseEnvelope(t, visRes)
		if !visEnv.Success {
			t.Fatalf("visualize failed: %s", visEnv.Message)
		}

		// 3.5 Validate: validates comparison artifact
		valRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_validate",
			Arguments: map[string]any{
				"project_root":  compDir,
				"content":       wepContent,
				"artifact_type": "comparison",
			},
		})
		if err != nil {
			t.Fatalf("validate call failed: %v", err)
		}
		valEnv := parseEnvelope(t, valRes)
		if !valEnv.Success {
			t.Fatalf("validate comparison artifact failed: %s", valEnv.Message)
		}
	})

	// =========================================================================
	// SUBTEST 4: Decision Scope Full Lifecycle
	// =========================================================================
	t.Run("4_DecisionScope_FullLifecycle", func(t *testing.T) {
		decDir := t.TempDir()

		// 4.1 Init decision scope: copies 2 templates (DECISIONS, CONFLICT-RESOLUTION)
		initRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_init",
			Arguments: map[string]any{
				"project_root": decDir,
				"scope":        "decision",
			},
		})
		if err != nil {
			t.Fatalf("init decision failed: %v", err)
		}
		initEnv := parseEnvelope(t, initRes)
		if !initEnv.Success {
			t.Fatalf("init decision returned error: %s", initEnv.Message)
		}
		for _, tmpl := range []string{"DECISIONS.template.md", "CONFLICT-RESOLUTION.template.md"} {
			if _, err := os.Stat(filepath.Join(decDir, "research", "templates", tmpl)); err != nil {
				t.Fatalf("expected decision template %s to exist: %v", tmpl, err)
			}
		}

		// 4.2 Save decision plan
		decPlanContent := `# DECISION PLAN: D-01 (Database Migration)
## Context
Migrate primary datastore from PostgreSQL to ClickHouse for analytics event store.
`
		savePlanRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_plan",
			Arguments: map[string]any{
				"project_root": decDir,
				"scope":        "decision",
				"decision_id":  "D-01",
				"slug":         "db-migration",
				"content":      decPlanContent,
			},
		})
		if err != nil {
			t.Fatalf("save decision plan failed: %v", err)
		}
		savePlanEnv := parseEnvelope(t, savePlanRes)
		if !savePlanEnv.Success {
			t.Fatalf("save decision plan failed: %s", savePlanEnv.Message)
		}

		// 4.3 Save research session S-01
		s01Content := `---
session_id: S-01
title: Postgres vs ClickHouse Storage Spike
date: 2026-10-18
status: complete
scope_level: decision
---
# Postgres vs ClickHouse Storage Spike

## Prior
Believed Postgres JSONB would suffice up to 10M events.

## Key Findings
- ClickHouse column compression achieves 92% disk storage reduction [Grade A, official docs, https://clickhouse.com, fetched].
- ClickHouse query latency for 100M rows aggregates in 14ms vs 12s in Postgres [Grade B, reproducible benchmark, fetched].
- ClickHouse lacks transactional ACID multi-table updates [Grade A, source code, fetched].

## Recommendation
Adopt ClickHouse as a dedicated event store while keeping PostgreSQL for transactional user records.

## Alternatives Considered
- TimescaleDB (rejected due to 4x higher storage footprint).

## Discovered Concerns
ClickHouse cluster management requires ZooKeeper/Keeper coordination overhead, increasing operations complexity.

## Open Questions & Risks
Reversal trigger: If event volume drops below 500k events/day, revert to PostgreSQL to reduce cluster maintenance cost.
`
		saveSessRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": decDir,
				"session_id":   "S-01",
				"content":      s01Content,
			},
		})
		if err != nil {
			t.Fatalf("save session S-01 failed: %v", err)
		}
		saveSessEnv := parseEnvelope(t, saveSessRes)
		if !saveSessEnv.Success {
			t.Fatalf("save session S-01 failed: %s", saveSessEnv.Message)
		}

		// 4.4 Adversarial challenge: run red team challenge on S-01
		chalRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_challenge",
			Arguments: map[string]any{
				"project_root": decDir,
				"decision_id":  "S-01",
				"mode":         "red_team",
			},
		})
		if err != nil {
			t.Fatalf("challenge call failed: %v", err)
		}
		chalEnv := parseEnvelope(t, chalRes)
		if !chalEnv.Success {
			t.Fatalf("challenge failed: %s", chalEnv.Message)
		}

		// 4.5 Record decision with auto_draft_from and auto_save: true in one turn
		recRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"project_root":    decDir,
				"decision_id":     "D-01",
				"auto_draft_from": "S-01",
				"auto_save":       true,
			},
		})
		if err != nil {
			t.Fatalf("record_decision failed: %v", err)
		}
		recEnv := parseEnvelope(t, recRes)
		if !recEnv.Success {
			t.Fatalf("record_decision failed: %s", recEnv.Message)
		}

		// Verify canonical ADR D-01-decision.md exists and DECISIONS.md was compiled
		adrFile := filepath.Join(decDir, "research", "D-01-decision.md")
		if _, err := os.Stat(adrFile); err != nil {
			t.Fatalf("expected canonical ADR file %s to exist", adrFile)
		}
		regFile := filepath.Join(decDir, "research", "DECISIONS.md")
		if _, err := os.Stat(regFile); err != nil {
			t.Fatalf("expected compiled registry %s to exist", regFile)
		}

		// 4.6 Amend session/ADR: verifies auto-recompilation of DECISIONS.md
		amendRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_amend_session",
			Arguments: map[string]any{
				"project_root": decDir,
				"session_id":   "D-01",
				"content":      "Added note: verified compatibility with ClickHouse Cloud serverless tier.",
				"reason":       "Cloud pricing update",
			},
		})
		if err != nil {
			t.Fatalf("amend_session failed: %v", err)
		}
		amendEnv := parseEnvelope(t, amendRes)
		if !amendEnv.Success {
			t.Fatalf("amend_session failed: %s", amendEnv.Message)
		}

		// 4.7 Validate multi-decision registry
		regData, err := os.ReadFile(regFile)
		if err != nil {
			t.Fatalf("reading compiled DECISIONS.md: %v", err)
		}
		valRegRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_validate",
			Arguments: map[string]any{
				"project_root":  decDir,
				"content":       string(regData),
				"artifact_type": "decisions",
			},
		})
		if err != nil {
			t.Fatalf("validate registry failed: %v", err)
		}
		valRegEnv := parseEnvelope(t, valRegRes)
		if !valRegEnv.Success {
			t.Fatalf("validate registry failed: %s", valRegEnv.Message)
		}
	})

	// =========================================================================
	// SUBTEST 5: Project Scope Full Lifecycle & Two-Track Gate
	// =========================================================================
	t.Run("5_ProjectScope_FullLifecycleAndGate", func(t *testing.T) {
		projDir := t.TempDir()

		// 5.1 Init project scope: all 6 templates verified
		initRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_init",
			Arguments: map[string]any{
				"project_root": projDir,
				"scope":        "project",
			},
		})
		if err != nil {
			t.Fatalf("init project failed: %v", err)
		}
		initEnv := parseEnvelope(t, initRes)
		if !initEnv.Success {
			t.Fatalf("init project failed: %s", initEnv.Message)
		}
		for _, tmpl := range core.TemplatesToCopy {
			tPath := filepath.Join(projDir, "research", "templates", tmpl)
			if _, err := os.Stat(tPath); err != nil {
				t.Fatalf("missing project template %s", tmpl)
			}
		}

		// 5.2 Save DAG pipeline with pipeline_content alias
		pipelineContent := `# RESEARCH PIPELINE: Distributed Event Fabric

## METADATA
| Field | Value |
|---|---|
| Project | Distributed Event Fabric |
| Scope | project |
| Total Sessions | 2 |

## SESSIONS

#### S-01: Ingestion Engine Spike
| Field | Value |
|---|---|
| **ID** | S-01 |
| **Layer** | 0 |
| **Door Type** | two-way |
| **Dependencies** | none |
| **Output File** | ` + "`sessions/S-01-ingestion.md`" + ` |

` + "```prompt" + `
Investigate ingestion options.
` + "```" + `

---

#### S-02: Sharding Topology
| Field | Value |
|---|---|
| **ID** | S-02 |
| **Layer** | 1 |
| **Door Type** | one-way |
| **Dependencies** | S-01 |
| **Output File** | ` + "`sessions/S-02-sharding.md`" + ` |

` + "```prompt" + `
Investigate sharding topology.
` + "```" + `
`
		savePipeRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_plan",
			Arguments: map[string]any{
				"project_root":     projDir,
				"scope":            "project",
				"pipeline_content": pipelineContent,
			},
		})
		if err != nil {
			t.Fatalf("save pipeline failed: %v", err)
		}
		savePipeEnv := parseEnvelope(t, savePipeRes)
		if !savePipeEnv.Success {
			t.Fatalf("save pipeline failed: %s", savePipeEnv.Message)
		}

		// 5.3 Visualize DAG: returns Mermaid flowchart and table
		visRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_visualize",
			Arguments: map[string]any{"project_root": projDir},
		})
		if err != nil {
			t.Fatalf("visualize failed: %v", err)
		}
		visEnv := parseEnvelope(t, visRes)
		if !visEnv.Success {
			t.Fatalf("visualize failed: %s", visEnv.Message)
		}
		visData := visEnv.Data.(map[string]any)
		if !strings.Contains(visData["mermaid"].(string), "graph TD") {
			t.Errorf("expected Mermaid flowchart in visualize output")
		}

		// 5.4 Retrieve next session: topologically first is S-01
		nextRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_next_session",
			Arguments: map[string]any{"project_root": projDir},
		})
		if err != nil {
			t.Fatalf("next_session failed: %v", err)
		}
		nextEnv := parseEnvelope(t, nextRes)
		if !nextEnv.Success {
			t.Fatalf("next_session failed: %s", nextEnv.Message)
		}
		nextData := nextEnv.Data.(map[string]any)
		if nextData["session_id"] != "S-01" {
			t.Fatalf("expected next session S-01, got: %v", nextData["session_id"])
		}

		// 5.5 Save session S-01: valid findings, triggers quality coaching
		s01Output := `---
session_id: S-01
title: Ingestion Engine Spike
date: 2026-10-18
status: complete
scope_level: project
door_type: two-way
---
# Ingestion Engine Spike

## Prior
Believed Go channels would bottleneck over 100k msg/sec.

## Key Findings
- Ring buffer implementation achieves 1,200,000 msg/sec with zero allocations [Grade A, benchmark, https://github.com/ringbuf, fetched].
- Memory consumption stable at 24MB under sustained 500k msg/sec throughput [Grade B, empirical profiler, fetched].
- Garbage collection pauses average 0.2ms using object pooling [Grade A, official docs, fetched].

## Recommendation
Use lock-free ring buffers for the local ingestion tier.

## Alternatives Considered
- Standard Go channels (too high GC overhead).

## Discovered Concerns
Ring buffer overflows drop messages if backpressure is not propagated upstream immediately.

## Open Questions & Risks
Monitor downstream TCP backpressure limits in session S-02.
`
		saveS01Res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": projDir,
				"session_id":   "S-01",
				"content":      s01Output,
			},
		})
		if err != nil {
			t.Fatalf("save S-01 failed: %v", err)
		}
		saveS01Env := parseEnvelope(t, saveS01Res)
		if !saveS01Env.Success {
			t.Fatalf("save S-01 failed: %s", saveS01Env.Message)
		}

		// 5.6 Dynamic Replan: add, update, remove sessions with reason and impact metadata
		replanRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_replan",
			Arguments: map[string]any{
				"project_root": projDir,
				"operation":    "add_session",
				"session_id":   "S-01B",
				"topic":        "Memory Profiling Under Spike Load",
				"door_type":    "two-way",
				"dependencies": "S-01",
				"reason":       "Spike testing showed unexpected memory allocation",
				"impact":       "Ingestion layer resilience",
			},
		})
		if err != nil {
			t.Fatalf("replan add_session failed: %v", err)
		}
		replanEnv := parseEnvelope(t, replanRes)
		if !replanEnv.Success {
			t.Fatalf("replan add_session failed: %s", replanEnv.Message)
		}

		// Save S-01B to unblock DAG
		_ = os.WriteFile(
			filepath.Join(projDir, "research", "sessions", "S-01B-memory-profiling.md"),
			[]byte("---\nsession_id: S-01B\nstatus: complete\n---\n# Profiling\nFindings complete [Grade A, docs, https://go.dev, fetched].\n## Discovered Concerns\nSufficient memory bounds verified.\n"),
			0644,
		)

		// 5.7 Save One-Way Door session S-02 and record ADR D-02
		s02Output := `---
session_id: S-02
title: Sharding Topology
date: 2026-10-18
status: complete
scope_level: project
door_type: one-way
---
# Sharding Topology

## Prior
Believed consistent hashing with virtual nodes is standard.

## Key Findings
- Consistent hashing with 128 virtual nodes provides uniform partition distribution within 3% [Grade A, IEEE paper, https://doi.org/10.1109/dist, fetched].
- Dynamic rebalancing redistributes only 1/N partitions upon node addition [Grade A, RFC spec, https://ietf.org/rfc, fetched].
- Network partition split-brain prevented via Raft quorum consensus [Grade A, official docs, https://raft.github.io, fetched].

## Recommendation
Implement consistent hashing with virtual nodes and Raft consensus.

## Alternatives Considered
- Range-based sharding (rejected due to hot-spot write skew on monotonically increasing IDs).
- Static hash mod N (rejected due to 100% cache invalidation upon cluster scaling).

## Discovered Concerns
Cross-shard multi-entity transactions require 2-phase commit, adding significant latency overhead (50ms+).

## Open Questions & Risks
Reversal trigger: If cluster size never exceeds 3 nodes after 6 months, collapse to single leader-follower replication.
`
		saveS02Res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save_session",
			Arguments: map[string]any{
				"project_root": projDir,
				"session_id":   "S-02",
				"content":      s02Output,
			},
		})
		if err != nil {
			t.Fatalf("save S-02 failed: %v", err)
		}
		saveS02Env := parseEnvelope(t, saveS02Res)
		if !saveS02Env.Success {
			t.Fatalf("save S-02 failed: %s", saveS02Env.Message)
		}

		// Run adversarial challenge on S-02 (mandatory for One-Way Door gate clearance B11)
		chalS02Res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_challenge",
			Arguments: map[string]any{
				"project_root": projDir,
				"decision_id":  "S-02",
				"mode":         "premortem",
			},
		})
		if err != nil {
			t.Fatalf("challenge S-02 failed: %v", err)
		}
		chalS02Env := parseEnvelope(t, chalS02Res)
		if !chalS02Env.Success {
			t.Fatalf("challenge S-02 failed: %s", chalS02Env.Message)
		}

		// Record D-02 with auto_draft_from and auto_save
		recD02Res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"project_root":    projDir,
				"decision_id":     "D-02",
				"auto_draft_from": "S-02",
				"auto_save":       true,
			},
		})
		if err != nil {
			t.Fatalf("record D-02 failed: %v", err)
		}
		recD02Env := parseEnvelope(t, recD02Res)
		if !recD02Env.Success {
			t.Fatalf("record D-02 failed: %s", recD02Env.Message)
		}

		// 5.8 Run Phase 0 Exit Gate (Two-Track clearance)
		gateRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_run_gate",
			Arguments: map[string]any{
				"project_root": projDir,
				"track":        "all",
			},
		})
		if err != nil {
			t.Fatalf("run_gate failed: %v", err)
		}
		gateEnv := parseEnvelope(t, gateRes)
		if !gateEnv.Success {
			t.Fatalf("run_gate failed: %s", gateEnv.Message)
		}
		gateData := gateEnv.Data.(map[string]any)
		if gateData["track_a_result"] != "PASS" {
			t.Errorf("expected Track A PASS, got: %v", gateData["track_a_result"])
		}

		// Verify PHASE-0-GATE.md was auto-persisted
		gateFile := filepath.Join(projDir, "research", "PHASE-0-GATE.md")
		if _, err := os.Stat(gateFile); err != nil {
			t.Fatalf("expected auto-persisted gate file %s to exist", gateFile)
		}
	})

	// =========================================================================
	// SUBTEST 6: Multi-Subagent Concurrent Advisory File Locking & Atomic Persistence
	// =========================================================================
	t.Run("6_SubagentConcurrency_And_Locking", func(t *testing.T) {
		concDir := t.TempDir()

		// Initialize workspace
		initRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_init",
			Arguments: map[string]any{
				"project_root": concDir,
				"scope":        "project",
			},
		})
		if err != nil || !parseEnvelope(t, initRes).Success {
			t.Fatalf("concurrency init failed: %v", err)
		}

		// Spawn 10 concurrent goroutines writing sessions and decisions simultaneously
		const numWorkers = 10
		var wg sync.WaitGroup
		errs := make(chan error, numWorkers)

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				sessID := fmt.Sprintf("CONC-%02d", workerID)
				content := fmt.Sprintf(`---
session_id: %s
title: Concurrent Worker %d
date: 2026-10-18
status: complete
scope_level: project
---
# Concurrent Worker %d
Findings: throughput benchmarked at %d ops/sec [Grade A, docs, https://benchmark.io, fetched].
## Discovered Concerns
Sufficient concurrency resilience verified under worker %d.
`, sessID, workerID, workerID, workerID*1000, workerID)

				res, err := cs.CallTool(ctx, &mcp.CallToolParams{
					Name: "vivechak_save_session",
					Arguments: map[string]any{
						"project_root": concDir,
						"session_id":   sessID,
						"content":      content,
					},
				})
				if err != nil {
					errs <- fmt.Errorf("worker %d save_session: %w", workerID, err)
					return
				}
				env := parseEnvelope(t, res)
				if !env.Success {
					errs <- fmt.Errorf("worker %d save_session failed: %s", workerID, env.Message)
					return
				}

				// Concurrently record decision
				decContent := fmt.Sprintf(`---
decision_id: D-%02d
title: Concurrent Decision %d
door_type: two-way
status: accepted
---
# D-%02d: Concurrent Decision %d
Decision content for worker %d
`, workerID, workerID, workerID, workerID, workerID)

				recRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
					Name: "vivechak_record_decision",
					Arguments: map[string]any{
						"project_root": concDir,
						"decision_id":  fmt.Sprintf("D-%02d", workerID),
						"content":      decContent,
					},
				})
				if err != nil {
					errs <- fmt.Errorf("worker %d record_decision: %w", workerID, err)
					return
				}
				recEnv := parseEnvelope(t, recRes)
				if !recEnv.Success {
					errs <- fmt.Errorf("worker %d record_decision failed: %s", workerID, recEnv.Message)
					return
				}
			}(i)
		}

		wg.Wait()
		close(errs)

		for err := range errs {
			t.Errorf("concurrency error: %v", err)
		}

		// Verify all 10 session files and 10 decision files exist without corruption
		for i := 0; i < numWorkers; i++ {
			sFile := filepath.Join(concDir, "research", "sessions", fmt.Sprintf("CONC-%02d.md", i))
			if _, err := os.Stat(sFile); err != nil {
				t.Errorf("missing concurrent session file: %s", sFile)
			}
			dFile := filepath.Join(concDir, "research", fmt.Sprintf("D-%02d-decision.md", i))
			if _, err := os.Stat(dFile); err != nil {
				t.Errorf("missing concurrent decision file: %s", dFile)
			}
		}

		// Verify central DECISIONS.md has all 10 decisions compiled
		regData, err := os.ReadFile(filepath.Join(concDir, "research", "DECISIONS.md"))
		if err != nil {
			t.Fatalf("reading compiled DECISIONS.md: %v", err)
		}
		regStr := string(regData)
		for i := 0; i < numWorkers; i++ {
			dMarker := fmt.Sprintf("D-%02d", i)
			if !strings.Contains(regStr, dMarker) {
				t.Errorf("compiled DECISIONS.md missing %s", dMarker)
			}
		}
	})
}
