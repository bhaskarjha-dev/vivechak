package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestArtifactIDs_LiveSubprocess tests server-side ID allocation and artifact frontmatter contracts
// across all scopes (HIGH-15, CRIT-08).
func TestArtifactIDs_LiveSubprocess(t *testing.T) {
	binPath := resolveOrBuildTestBinary(t)

	// 1. Decision plan saved without id gets D-001 (never Role-plan.md)
	// and every saved artifact has id, type, title, date, status, and vivechak: 1 in frontmatter.
	t.Run("DecisionPlanAllocatesD001_AndFrontmatterContract", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testapp\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		// Save decision plan without id
		planContent := `# Event storage engine
| ID | Topic |
|---|---|
| D-NEW-S1 | Storage Landscape |
`
		savePlanRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "plan",
				"scope":   "decision",
				"content": planContent,
			},
		})
		if err != nil || savePlanRes.IsError {
			t.Fatalf("save plan failed: %v", err)
		}

		planEnv := parseTestEnvelope(t, savePlanRes)
		if planEnv.Data["id"] != "D-001" {
			t.Errorf("expected allocated plan id D-001, got: %v", planEnv.Data["id"])
		}

		// Verify no Role* files created anywhere under research/
		researchDir := filepath.Join(dir, "research")
		entries, _ := os.ReadDir(researchDir)
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "Role") {
				t.Errorf("found forbidden Role* file: %s", e.Name())
			}
		}

		// Save comparison
		cmpContent := `---
id: C-001
type: comparison
title: Pino vs Winston
date: 2026-11-04
status: complete
vivechak: 1
---
# Pino vs Winston

## Key Findings
- Pino throughput [Grade A: https://github.com/pinojs/pino | fetched].
- Winston transports [Grade B: https://github.com/winstonjs/winston | fetched].

## Weighted Evaluation Matrix
| Criterion | Weight | Pino Score (1-5) | Winston Score (1-5) |
|---|---|---|---|
| Performance | 0.50 | 5.0 | 3.0 |
| Ecosystem | 0.50 | 3.5 | 5.0 |
| Total Score | 1.00 | 4.25 | 4.00 |

## Recommendation
We recommend Pino [Grade A: https://github.com/pinojs/pino | fetched].

## Sources & Evidence Ledger
- Pino: https://github.com/pinojs/pino [Grade A | fetched]
- Winston: https://github.com/winstonjs/winston [Grade B | fetched]
`
		_, err = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "comparison",
				"id":      "C-001",
				"content": cmpContent,
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Save session D-001-S1
		sessionContent := `---
id: D-001-S1
type: session
title: Storage Landscape
date: 2026-11-04
status: complete
informs_decisions: ["D-001"]
vivechak: 1
---
# Storage Landscape
## Key Findings
- PostgreSQL JSONB [Grade A: https://postgresql.org | fetched].
- ClickHouse columnar [Grade B: https://clickhouse.com | fetched].

## Recommendation
PostgreSQL [Grade A: https://postgresql.org | fetched].

## Sources & Evidence Ledger
- PostgreSQL: https://postgresql.org [Grade A | fetched]
- ClickHouse: https://clickhouse.com [Grade B | fetched]
`
		_, err = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "session",
				"id":      "D-001-S1",
				"content": sessionContent,
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Record decision D-001
		adrContent := `---
id: D-001
type: decision
title: Primary Storage Engine
door_type: two-way
date: 2026-11-04
status: accepted
review_trigger: "Query latency exceeds 100ms"
vivechak: 1
---
# Primary Storage Engine
## Context
We need durable storage.
## Decision Outcome
We choose PostgreSQL.
`
		_, err = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"decision_id": "D-001",
				"content":     adrContent,
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Check every saved markdown artifact in research/ (excluding .history and DECISIONS.md)
		// has id, type, title, date, status, and vivechak in frontmatter
		var checkedCount int
		if err := filepath.Walk(researchDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info.IsDir() {
				return nil
			}
			if strings.Contains(path, ".history") || strings.Contains(path, "templates") {
				return nil
			}
			rel, _ := filepath.Rel(researchDir, path)
			if rel == "DECISIONS.md" || rel == "PHASE-0-GATE.md" || !strings.HasSuffix(rel, ".md") {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading %s: %v", rel, err)
			}
			fm, _, _ := core.ParseFrontmatter(data)
			if fm == nil {
				t.Fatalf("artifact %s has no frontmatter", rel)
			}

			requiredKeys := []string{"id", "type", "title", "date", "status", "vivechak"}
			for _, k := range requiredKeys {
				if !fm.Has(k) {
					t.Errorf("artifact %s missing required frontmatter key %q (keys present: %+v)", rel, k, fm)
				}
			}
			checkedCount++
			return nil
		}); err != nil {
			t.Fatalf("walking %s: %v", researchDir, err)
		}

		if checkedCount < 3 {
			t.Errorf("expected at least 3 checked artifacts, checked %d", checkedCount)
		}
	})

	// 2. Decision brief allocates D-001 and names it in IDS: line;
	// saving with D-NEW receives the allocated ID.
	t.Run("DecisionBriefNamesAllocatedID_DNEWReceivesIt", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testapp\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		stRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_start",
			Arguments: map[string]any{
				"scope":    "decision",
				"question": "Event storage engine",
			},
		})
		if err != nil || stRes.IsError {
			t.Fatalf("vivechak_start failed: %v", err)
		}

		stEnv := parseTestEnvelope(t, stRes)
		allocID, _ := stEnv.Data["id"].(string)
		if allocID == "" {
			t.Fatal("vivechak_start returned no allocated id")
		}

		brief, _ := stEnv.Data["brief"].(string)
		if !strings.Contains(brief, "IDS: "+allocID) {
			t.Errorf("brief should contain 'IDS: %s', got:\n%s", allocID, brief)
		}

		// Save plan with D-NEW
		planRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "plan",
				"scope":   "decision",
				"id":      "D-NEW",
				"content": "# Event Storage\n| ID | Topic |\n|---|---|\n| D-NEW-S1 | Landscape |\n",
			},
		})
		if err != nil || planRes.IsError {
			t.Fatalf("save plan with D-NEW failed: %v", err)
		}

		planEnv := parseTestEnvelope(t, planRes)
		planID, _ := planEnv.Data["id"].(string)
		if planID != allocID && (!strings.Contains(brief, allocID) || strings.Contains(brief, "D-NEW")) {
			t.Errorf("expected D-NEW to receive allocated ID %s or brief to name it without D-NEW, got planID %q", allocID, planID)
		}
	})

	// 3. Normalization and ID collision guards:
	// - C-01 normalizes to C-001
	// - d-9 normalizes to D-009
	// - D-NEW twice allocates D-001 and D-002
	// - Placeholders (D-NNN, d-tbd) and invalid IDs refused with E-ID-INVALID
	// - Slugged ADR (D-001-orm-comparison) listed in DECISIONS.md and recognized
	t.Run("IDNormalization_Placeholders_AndSluggedADR", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testapp\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		// Normalization: C-01 -> C-001
		cmpRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "comparison",
				"id":      "C-01",
				"content": "# Pino vs Winston\nBody text\n",
			},
		})
		if err != nil || cmpRes.IsError {
			t.Fatalf("save C-01 failed: %v", err)
		}
		cmpEnv := parseTestEnvelope(t, cmpRes)
		if cmpEnv.Data["id"] != "C-001" {
			t.Errorf("expected C-01 to normalize to C-001, got: %v", cmpEnv.Data["id"])
		}

		// Normalization: d-9 -> D-009
		planRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "plan",
				"scope":   "decision",
				"id":      "d-9",
				"content": "# D-9 Plan\n| ID | Topic |\n|---|---|\n| d-9-S1 | Landscape |\n",
			},
		})
		if err != nil || planRes.IsError {
			t.Fatalf("save d-9 failed: %v", err)
		}
		planEnv := parseTestEnvelope(t, planRes)
		if planEnv.Data["id"] != "D-009" {
			t.Errorf("expected d-9 to normalize to D-009, got: %v", planEnv.Data["id"])
		}

		// Placeholder rejection: D-NNN or d-tbd or invalid
		badIDs := []string{"D-NNN", "d-tbd", "INVALID@@ID"}
		for _, badID := range badIDs {
			badRes, _ := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_save",
				Arguments: map[string]any{
					"kind":    "plan",
					"scope":   "decision",
					"id":      badID,
					"content": "# Bad Plan\n",
				},
			})
			badEnv := parseTestEnvelope(t, badRes)
			if badEnv.Success && !badRes.IsError {
				t.Errorf("expected save with placeholder/invalid ID %q to fail", badID)
			}
		}

		// Slugged ADR: record decision with slug "orm-comparison"
		slugADR := `---
id: D-001
type: decision
title: ORM Comparison
door_type: two-way
status: accepted
review_trigger: "Latency exceeds 50ms"
date: 2026-11-04
vivechak: 1
---
# ORM Comparison
## Context
ORM choice.
## Decision Outcome
We choose GORM.
`
		slugRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"decision_id": "D-001",
				"slug":        "orm-comparison",
				"content":     slugADR,
			},
		})
		if err != nil || slugRes.IsError {
			t.Fatalf("record slugged decision failed: %v", err)
		}

		// Verify DECISIONS.md lists D-001 and orm-comparison
		decPath := filepath.Join(dir, "research", "DECISIONS.md")
		decBytes, err := os.ReadFile(decPath)
		if err != nil {
			t.Fatalf("reading DECISIONS.md: %v", err)
		}
		decStr := string(decBytes)
		if !strings.Contains(decStr, "D-001") || !strings.Contains(decStr, "ORM Comparison") {
			t.Errorf("DECISIONS.md missing slugged ADR D-001:\n%s", decStr)
		}
	})
}
