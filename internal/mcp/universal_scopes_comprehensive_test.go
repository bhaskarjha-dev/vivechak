package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

// TestUniversal_PrepareGenerator_AllScopes verifies prepare_generator for technical, algorithmic, and business contexts.
func TestUniversal_PrepareGenerator_AllScopes(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Comparison Scope — AI Models (Explicit Scope)
	resComp, envComp, err := handlePrepareGenerator(ctx, nil, PrepareGeneratorInput{
		ProjectRoot: tmpDir,
		Scope:       "comparison",
		Context:     "Claude 3.5 Sonnet vs GPT-4o for complex multi-turn coding agent workflows",
	})
	if err != nil || resComp.IsError {
		t.Fatalf("prepare_generator comparison failed: %v, env: %+v", err, envComp)
	}
	dataComp := envComp.Data.(map[string]any)
	if dataComp["scope"] != core.ScopeComparison {
		t.Errorf("expected scope comparison, got %v", dataComp["scope"])
	}
	promptComp := dataComp["prompt"].(string)
	if !strings.Contains(promptComp, "Claude 3.5 Sonnet vs GPT-4o") {
		t.Errorf("expected options injected into prompt, got: %s", promptComp[:200])
	}

	// 2. Decision Scope — Strategic Business Decision (Auto-inferred Scope)
	resDec, envDec, err := handlePrepareGenerator(ctx, nil, PrepareGeneratorInput{
		ProjectRoot: tmpDir,
		Context:     "strategic decision: build in-house sales team vs partner agency for US enterprise market entry",
	})
	if err != nil || resDec.IsError {
		t.Fatalf("prepare_generator decision failed: %v, env: %+v", err, envDec)
	}
	dataDec := envDec.Data.(map[string]any)
	if dataDec["scope"] != core.ScopeDecision {
		t.Errorf("expected auto-inferred scope decision, got %v", dataDec["scope"])
	}

	// 3. Project Scope — Business Venture Pipeline (Auto-inferred Scope)
	resProj, envProj, err := handlePrepareGenerator(ctx, nil, PrepareGeneratorInput{
		ProjectRoot: tmpDir,
		Context:     "we are planning a new business venture for global freight brokerage and logistics marketplace",
	})
	if err != nil || resProj.IsError {
		t.Fatalf("prepare_generator project failed: %v, env: %+v", err, envProj)
	}
	dataProj := envProj.Data.(map[string]any)
	if dataProj["scope"] != core.ScopeProject {
		t.Errorf("expected auto-inferred scope project, got %v", dataProj["scope"])
	}
}

// TestUniversal_SavePlan_DecisionAutoRecovery verifies decision_id recovery and auto-init.
func TestUniversal_SavePlan_DecisionAutoRecovery(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	planContent := `# Decision Plan: Distributed Cache Architecture
## Routing
| Field | Value |
|---|---|
| Decision | D-05 |
| Door | one-way |

## Session DAG
### Session D-05-S1: Cache Shootout
| Field | Value |
|---|---|
| ID | D-05-S1 |
| Layer | 1 |
| Door Type | one-way |
| Dependencies | none |
| Output File | research/sessions/D-05-S1.md |

` + "```prompt\nEvaluate Redis vs KeyDB.\n```\n"

	// Intentionally omit decision_id; verify it auto-recovers to D-05
	res, env, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot: tmpDir,
		Scope:       "decision",
		Content:     planContent,
	})
	if err != nil || res.IsError {
		t.Fatalf("save_plan failed: %v, env: %+v", err, env)
	}

	savedFile := filepath.Join(tmpDir, core.ResearchDir, "D-05-plan.md")
	if _, err := os.Stat(savedFile); os.IsNotExist(err) {
		t.Errorf("expected %s to be created via auto-recovered decision_id", savedFile)
	}
}

// TestUniversal_Challenge_ADRInResearchDir verifies that challenge finds standalone ADRs in research/.
func TestUniversal_Challenge_ADRInResearchDir(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize workspace
	_, _, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "decision"})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	adrContent := `---
id: D-01
title: Strategic Sales Sourcing
door_type: one-way
status: proposed
date: 2026-10-17
---
# D-01: Strategic Sales Sourcing

## Recommendation
Build in-house enterprise sales team. A (SaaS Metrics Survey) [direct]

## Alternatives Considered
- Partner agency: Lower upfront cost, but low alignment.
`
	adrPath := filepath.Join(tmpDir, core.ResearchDir, "D-01-sales.md")
	if err := os.WriteFile(adrPath, []byte(adrContent), 0o644); err != nil {
		t.Fatalf("writing ADR: %v", err)
	}

	// Challenge by decision_id; should resolve directly from research/
	res, env, err := handleChallenge(ctx, nil, ChallengeInput{
		ProjectRoot: tmpDir,
		SessionID:   "D-01",
		Mode:        "red_team",
	})
	if err != nil || res.IsError {
		t.Fatalf("challenge on ADR in research/ failed: %v, env: %+v", err, env)
	}
	cr, ok := env.Data.(*core.ChallengeResult)
	if !ok || len(cr.Prompts) == 0 {
		t.Errorf("expected challenge prompts generated from ADR, got: %+v", env.Data)
	}
}

// TestUniversal_AmendSession_ResearchDirFiles verifies amending files directly in research/ (comparisons and ADRs).
func TestUniversal_AmendSession_ResearchDirFiles(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "comparison"})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	compContent := `---
id: C-01
title: Model Shootout
date: 2026-10-17
status: complete
---
# Model Shootout
Initial comparison notes.
`
	compPath := filepath.Join(tmpDir, core.ResearchDir, "C-01-comparison.md")
	if err := os.WriteFile(compPath, []byte(compContent), 0o644); err != nil {
		t.Fatalf("writing comparison: %v", err)
	}

	// Amend comparison in research/
	res, env, err := handleAmendSession(ctx, nil, AmendSessionInput{
		ProjectRoot: tmpDir,
		SessionID:   "C-01",
		Amendment:   "### Update 2026-10-17\nVerified sub-millisecond inference latency on H100s.\n",
	})
	if err != nil || res.IsError {
		t.Fatalf("amend comparison failed: %v, env: %+v", err, env)
	}

	updated, err := os.ReadFile(compPath)
	if err != nil {
		t.Fatalf("reading amended file: %v", err)
	}
	if !strings.Contains(string(updated), "sub-millisecond inference latency") {
		t.Errorf("expected amendment in comparison file, got: %s", string(updated))
	}
}

// TestUniversal_WorkspaceValidate_AllScopes verifies workspace validation across all 3 scope layouts.
func TestUniversal_WorkspaceValidate_AllScopes(t *testing.T) {
	ctx := context.Background()

	// 1. Decision Scope Workspace Validation
	tmpDec := t.TempDir()
	_, _, _ = handleInit(ctx, nil, InitInput{ProjectRoot: tmpDec, Scope: "decision"})
	decPlan := `# Decision Plan
### Session D-01-S1: Shortlist
| Field | Value |
|---|---|
| ID | D-01-S1 |
| Layer | 1 |
| Door Type | two-way |
| Dependencies | none |
| Output File | research/sessions/D-01-S1.md |

` + "```prompt\nShortlist candidates.\n```\n"
	_ = os.WriteFile(filepath.Join(tmpDec, core.ResearchDir, "D-01-plan.md"), []byte(decPlan), 0o644)

	resDec, envDec, err := handleValidate(ctx, nil, ValidateInput{ProjectRoot: tmpDec})
	if err != nil || resDec.IsError {
		t.Fatalf("validate decision workspace failed: %v, env: %+v", err, envDec)
	}
	dataDec := envDec.Data.(map[string]any)
	if _, ok := dataDec["pipeline"]; !ok {
		t.Errorf("expected plan validation in decision workspace results, got: %+v", dataDec)
	}

	// 2. Comparison Scope Workspace Validation
	tmpComp := t.TempDir()
	_, _, _ = handleInit(ctx, nil, InitInput{ProjectRoot: tmpComp, Scope: "comparison"})
	compMD := `---
id: C-01
title: Model Shootout
date: 2026-10-17
status: complete
---
# Model Shootout
Candidate A outperforms Candidate B. A (Benchmark Report) [direct]
`
	_ = os.WriteFile(filepath.Join(tmpComp, core.ResearchDir, "C-01-comparison.md"), []byte(compMD), 0o644)

	resComp, envComp, err := handleValidate(ctx, nil, ValidateInput{ProjectRoot: tmpComp})
	if err != nil || resComp.IsError {
		t.Fatalf("validate comparison workspace failed: %v, env: %+v", err, envComp)
	}
	dataComp := envComp.Data.(map[string]any)
	if _, ok := dataComp["comparison"]; !ok {
		t.Errorf("expected comparison validation in comparison workspace results, got: %+v", dataComp)
	}
}
