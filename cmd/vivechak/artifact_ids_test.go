package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
		planContent := wgDecisionPlan("D-NEW", "Event storage engine")
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
				"id":      "D-001",
				"content": adrContent,
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
				"content": wgDecisionPlan("D-NEW", "Event Storage"),
			},
		})
		if err != nil || planRes.IsError {
			t.Fatalf("save plan with D-NEW failed: %v", err)
		}

		planEnv := parseTestEnvelope(t, planRes)
		planID, _ := planEnv.Data["id"].(string)
		if planID != allocID {
			t.Errorf("expected the plan saved with D-NEW to receive the allocated ID %s, got planID %q", allocID, planID)
		}
	})

	// 3. Normalization and ID collision guards:
	// - C-01 normalizes to C-001
	// - d-9 normalizes to D-009
	// - D-NEW twice allocates D-001 and D-002
	// - Placeholders (D-NNN, d-tbd) get the next free id; invalid IDs are refused with E-ID-INVALID
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
				"content": wgDecisionPlan("D-009", "D-9 Plan"),
			},
		})
		if err != nil || planRes.IsError {
			t.Fatalf("save d-9 failed: %v", err)
		}
		planEnv := parseTestEnvelope(t, planRes)
		if planEnv.Data["id"] != "D-009" {
			t.Errorf("expected d-9 to normalize to D-009, got: %v", planEnv.Data["id"])
		}

		// Placeholders (D-NNN, d-tbd) get the next free id instead of failing (HIGH-15), and never become literal
		// file names; a malformed id is refused with E-ID-INVALID and nothing is written.
		for _, ph := range []string{"D-NNN", "d-tbd"} {
			phRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_save",
				Arguments: map[string]any{
					"kind":    "plan",
					"scope":   "decision",
					"id":      ph,
					"content": wgDecisionPlan("D-NEW", "Bad Plan"),
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			phEnv := parseTestEnvelope(t, phRes)
			phID, _ := phEnv.Data["id"].(string)
			if phRes.IsError || !strings.HasPrefix(phID, "D-") || len(phID) != 5 {
				t.Errorf("expected placeholder %q to get the next D-NNN id, got isError=%v id=%q (%s)", ph, phRes.IsError, phID, phEnv.Message)
			}
			if !strings.Contains(strings.Join(phEnv.WarningStrings(), "\n"), "W-ID-ALLOCATED: allocated decision ID "+phID) {
				t.Errorf("expected W-ID-ALLOCATED naming %s, got %v", phID, phEnv.Warnings)
			}
		}
		badRes, _ := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "plan",
				"scope":   "decision",
				"id":      "INVALID@@ID",
				"content": wgDecisionPlan("D-NEW", "Bad Plan"),
			},
		})
		if badEnv := parseTestEnvelope(t, badRes); !badRes.IsError || !strings.Contains(badEnv.Message, "E-ID-INVALID") {
			t.Errorf("expected save with invalid ID to fail with E-ID-INVALID, got isError=%v %q", badRes.IsError, badEnv.Message)
		}
		for _, f := range wgResearchFiles(t, dir) {
			if strings.Contains(strings.ToUpper(f), "NNN") || strings.Contains(strings.ToUpper(f), "TBD") || strings.Contains(f, "INVALID") {
				t.Errorf("a placeholder or invalid id became a file name: %s", f)
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
				"id":      "D-001",
				"slug":    "orm-comparison",
				"content": slugADR,
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

	// 4. One id rule (HIGH-15): the documented record_decision(decision_id="D-NNN") call after vivechak_start gets
	// the id the start reserved; further placeholders (any case, with or without prefix) get the next free id with
	// W-ID-ALLOCATED; a malformed id is refused with E-ID-INVALID and nothing is written; a comparison saved
	// without an id after vivechak_start(scope="comparison") gets the reserved C-NNN.
	t.Run("PlaceholdersAllocate_ReservedIDsUsed_MalformedRefused", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		st, _ := wgCall(t, ctx, cs, "vivechak_start", map[string]any{"scope": "decision", "question": "Primary datastore"})
		reserved, _ := st.Data["id"].(string)
		if reserved != "D-001" {
			t.Fatalf("expected vivechak_start to reserve D-001, got %q", reserved)
		}
		for i, tc := range []struct{ raw, want string }{{"D-NNN", "D-001"}, {"TBD", "D-002"}, {"d-xxx", "D-003"}, {"new", "D-004"}} {
			env, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{
				"id":      tc.raw,
				"content": wgADR(tc.raw, fmt.Sprintf("Decision %d", i+1), "proposed", "two-way"),
			})
			if isErr || env.Data["id"] != tc.want {
				t.Fatalf("placeholder %q: expected %s, got isError=%v id=%v (%s)", tc.raw, tc.want, isErr, env.Data["id"], env.Message)
			}
			if !strings.Contains(strings.Join(env.WarningStrings(), "\n"), "W-ID-ALLOCATED: allocated decision ID "+tc.want) {
				t.Errorf("placeholder %q: expected W-ID-ALLOCATED naming %s, got %v", tc.raw, tc.want, env.Warnings)
			}
			data, err := os.ReadFile(filepath.Join(dir, "research", tc.want+"-decision.md"))
			if err != nil {
				t.Fatalf("placeholder %q: ADR file missing: %v", tc.raw, err)
			}
			fm, _, _ := core.ParseFrontmatter(data)
			if fm.GetString("id") != tc.want || fm.GetString("type") != "decision" || fm.GetString("vivechak") != "1" {
				t.Errorf("placeholder %q: expected id %s, type decision, vivechak 1 in the ADR frontmatter, got %v", tc.raw, tc.want, fm)
			}
		}

		for _, bad := range []string{"Role", "D-1x", "../D-001"} {
			env, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{
				"id":      bad,
				"content": wgADR("D-777", "Not an id", "proposed", "two-way"),
			})
			if !isErr || !strings.Contains(env.Message, "E-ID-INVALID") || !strings.Contains(env.Message, "is not D-NNN or C-NNN") {
				t.Errorf("malformed id %q: expected E-ID-INVALID ... is not D-NNN or C-NNN, got isError=%v %q", bad, isErr, env.Message)
			}
		}
		if env, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "comparison", "id": "COMP-01", "content": wgComparison("A vs B", "Alpha", "Beta")}); !isErr || !strings.Contains(env.Message, "E-ID-INVALID") {
			t.Errorf("expected a comparison saved as COMP-01 to fail with E-ID-INVALID, got isError=%v %q", isErr, env.Message)
		}
		for _, f := range wgResearchFiles(t, dir) {
			upper := strings.ToUpper(f)
			if strings.Contains(f, "Role") || strings.Contains(upper, "NNN") || strings.Contains(upper, "TBD") || strings.Contains(upper, "XXX") ||
				strings.Contains(upper, "COMP-01") || strings.Contains(upper, "D-777") || strings.Contains(upper, "D-1X") {
				t.Errorf("a placeholder or refused id was written: %s", f)
			}
		}

		cst, _ := wgCall(t, ctx, cs, "vivechak_start", map[string]any{"scope": "comparison", "question": "Pino vs Winston"})
		cReserved, _ := cst.Data["id"].(string)
		env, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "comparison", "content": wgComparison("Pino vs Winston", "Pino", "Winston")})
		if isErr || cReserved == "" || env.Data["id"] != cReserved {
			t.Errorf("expected the comparison saved without an id to get the reserved %q, got isError=%v id=%v (%s)", cReserved, isErr, env.Data["id"], env.Message)
		}
	})

	// 5. ADRs carry type: decision and vivechak: 1 (also after auto_save), a seeded ADR gets the canonical name
	// research/<id>-decision.md, nothing is copied into research/templates/, and E-REVISE-REQUIRED names the title of
	// the file a save collides with.
	t.Run("SeededADRCanonicalName_AutoSaveKeepsType_NoTemplates_ReviseNamesTitle", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		wgCall(t, ctx, cs, "vivechak_start", map[string]any{"scope": "project", "question": "Event platform"})
		seed := strings.Replace(wgADR("D-NEW", "Event storage engine", "proposed", "one-way"), "human_reviewed: false\n", "human_reviewed: false\ntags: [storage, analytics]\n", 1)
		plan, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{
			"kind": "plan", "scope": "decision", "id": "D-NEW",
			"content":           wgDecisionPlan("D-NEW", "Event storage engine"),
			"decisions_content": seed,
		})
		if isErr || plan.Data["id"] != "D-001" {
			t.Fatalf("seeded plan save: expected D-001, got isError=%v id=%v (%s)", isErr, plan.Data["id"], plan.Message)
		}
		seeded := filepath.Join(dir, "research", "D-001-decision.md")
		data, err := os.ReadFile(seeded)
		if err != nil {
			t.Fatalf("expected the seeded ADR at research/D-001-decision.md (files %v): %v", wgResearchFiles(t, dir), err)
		}
		if _, err := os.Stat(filepath.Join(dir, "research", "D-001.md")); !os.IsNotExist(err) {
			t.Errorf("the seeded ADR must not use the old name research/D-001.md")
		}
		if strings.Contains(string(data), "D-NEW") {
			t.Errorf("the seeded ADR still holds the placeholder D-NEW:\n%s", data)
		}

		wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": "D-001-S1", "content": wgSession("D-001-S1", "D-001", "Landscape")})
		rec, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-001", "auto_draft_from": "D-001-S1", "auto_save": true})
		if isErr || rec.Data["file_path"] != "research/D-001-decision.md" {
			t.Fatalf("auto_save: expected research/D-001-decision.md, got isError=%v file_path=%v (%s)", isErr, rec.Data["file_path"], rec.Message)
		}
		data, _ = os.ReadFile(seeded)
		fm, _, _ := core.ParseFrontmatter(data)
		if fm.GetString("type") != "decision" || fm.GetString("vivechak") != "1" || fm.GetString("title") != "Event storage engine" {
			t.Errorf("auto_save must keep type: decision, vivechak: 1 and the seeded title, got %v", fm)
		}
		var adrs []string
		for _, f := range wgResearchFiles(t, dir) {
			if strings.HasPrefix(f, "research/D-001") && !strings.HasSuffix(f, "-plan.md") {
				adrs = append(adrs, f)
			}
		}
		if len(adrs) != 1 {
			t.Errorf("expected exactly one ADR file for D-001, got %v", adrs)
		}
		if _, err := os.Stat(filepath.Join(dir, "research", "templates")); !os.IsNotExist(err) {
			t.Errorf("expected no research/templates/ (templates are MCP resources), got err=%v", err)
		}

		env, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{
			"id":      "D-001",
			"content": wgADR("D-001", "Message queue", "proposed", "two-way"),
		})
		if !isErr || !strings.Contains(env.Message, "E-REVISE-REQUIRED") || !strings.Contains(env.Message, "Event storage engine") {
			t.Errorf("expected E-REVISE-REQUIRED naming the existing title \"Event storage engine\", got isError=%v %q", isErr, env.Message)
		}
	})

	// 6. A workspace made by the previous version seeded its ADR as research/D-001.md: later calls keep using that
	// file (auto_save, a plan re-save with decisions_content) and never create a second ADR for D-001.
	t.Run("LegacySeededADRStillUsed_NoDuplicate", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		research := filepath.Join(dir, "research")
		if err := os.MkdirAll(filepath.Join(research, "sessions"), 0o755); err != nil {
			t.Fatal(err)
		}
		legacyADR := strings.Replace(wgADR("D-001", "Event storage engine", "proposed", "one-way"), "human_reviewed: false\n", "human_reviewed: false\ntype: decision\nvivechak: 1\n", 1)
		if err := os.WriteFile(filepath.Join(research, "D-001.md"), []byte(legacyADR), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(research, "D-001-plan.md"), []byte(wgDecisionPlan("D-001", "Event storage engine")), 0o644); err != nil {
			t.Fatal(err)
		}
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": "D-001-S1", "content": wgSession("D-001-S1", "D-001", "Landscape")})
		rec, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-001", "auto_draft_from": "D-001-S1", "auto_save": true})
		if isErr || rec.Data["file_path"] != "research/D-001.md" {
			t.Fatalf("auto_save on a legacy seeded ADR: expected research/D-001.md, got isError=%v file_path=%v (%s)", isErr, rec.Data["file_path"], rec.Message)
		}
		if _, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{
			"kind": "plan", "id": "D-001", "revise": true,
			"content":           wgDecisionPlan("D-001", "Event storage engine"),
			"decisions_content": legacyADR,
		}); isErr {
			t.Fatalf("plan re-save with decisions_content failed")
		}
		for _, f := range wgResearchFiles(t, dir) {
			if f == "research/D-001-decision.md" {
				t.Errorf("a second ADR file was created next to the legacy research/D-001.md")
			}
		}
		gate, _ := wgCall(t, ctx, cs, "vivechak_gate", map[string]any{})
		if decs, ok := gate.Data["decisions"].([]any); ok {
			n := 0
			for _, d := range decs {
				if m, ok := d.(map[string]any); ok && m["id"] == "D-001" {
					n++
				}
			}
			if n != 1 {
				t.Errorf("expected the gate to see the legacy ADR D-001 exactly once, got %d in %v", n, decs)
			}
		}
	})

	// 7. The `decision` prompt never hard-codes the first id: after decision D-001 exists it does not say IDS: D-001.
	t.Run("DecisionPromptNamesNoFixedID", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "D-001", "content": wgDecisionPlan("D-001", "Primary datastore")})
		pr, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: "decision", Arguments: map[string]string{"question": "Message queue choice"}})
		if err != nil || pr == nil || len(pr.Messages) == 0 {
			t.Fatalf("prompts/get decision failed: %v", err)
		}
		text := pr.Messages[0].Content.(*mcp.TextContent).Text
		if strings.Contains(text, "D-001") {
			t.Errorf("the decision prompt still tells the agent to use D-001: %s", text)
		}
		if !strings.Contains(text, "vivechak_start") || len(text) > 600 {
			t.Errorf("expected the decision prompt (<=600 chars) to route through vivechak_start, got (%d chars): %s", len(text), text)
		}
	})

	// 8. The documented decision flow written with its placeholders (AGENTS.md: vivechak_save(kind="plan",
	// id="D-NNN") then vivechak_record_decision(decision_id="D-NNN", auto_draft_from="D-NNN-S1", auto_save=true))
	// lands plan, session and ADR on the one id vivechak_start reserved.
	t.Run("DocumentedPlaceholderFlow_LandsOnReservedDecision", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		st, _ := wgCall(t, ctx, cs, "vivechak_start", map[string]any{"scope": "decision", "question": "Event storage engine"})
		reserved, _ := st.Data["id"].(string)
		plan, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "D-NNN", "content": wgDecisionPlan("D-NEW", "Event storage engine")})
		if isErr || plan.Data["id"] != reserved {
			t.Fatalf("plan saved with D-NNN: expected the reserved %s, got isError=%v id=%v (%s)", reserved, isErr, plan.Data["id"], plan.Message)
		}
		next, _ := wgCall(t, ctx, cs, "vivechak_next", map[string]any{})
		sid, _ := next.Data["session_id"].(string)
		if sid != reserved+"-S1" {
			t.Fatalf("expected vivechak_next to serve %s-S1, got %q", reserved, sid)
		}
		wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": sid, "content": wgSession(sid, reserved, "Landscape")})
		rec, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-NNN", "auto_draft_from": "D-NNN-S1", "auto_save": true})
		if isErr || rec.Data["id"] != reserved || rec.Data["file_path"] != "research/"+reserved+"-decision.md" {
			t.Fatalf("record_decision(D-NNN, auto_draft_from=D-NNN-S1, auto_save): expected %s at research/%s-decision.md, got isError=%v id=%v file=%v (%s)",
				reserved, reserved, isErr, rec.Data["id"], rec.Data["file_path"], rec.Message)
		}
		for _, f := range wgResearchFiles(t, dir) {
			if strings.HasPrefix(f, "research/D-002") {
				t.Errorf("the documented flow split the decision across two ids: %s", f)
			}
		}
	})
	// 9. The literal AGENTS.md decision flow with every id written as the placeholder D-NNN (HIGH-15 follow-up):
	// vivechak_start(scope="decision") -> vivechak_save(kind="plan", id="D-NNN") -> vivechak_save(kind="session",
	// id="D-NNN-S1") -> vivechak_record_decision(decision_id="D-NNN", auto_draft_from="D-NNN-S1", auto_save=true).
	// The session and the ADR land on D-001 and nothing named after the placeholder is written.
	t.Run("LiteralAGENTSDecisionFlow_AllPlaceholders", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		st, _ := wgCall(t, ctx, cs, "vivechak_start", map[string]any{"scope": "decision", "question": "Event storage engine"})
		if st.Data["id"] != "D-001" {
			t.Fatalf("expected vivechak_start to reserve D-001, got %v", st.Data["id"])
		}
		plan, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "D-NNN", "content": wgDecisionPlan("D-NNN", "Event storage engine")})
		if isErr || plan.Data["id"] != "D-001" {
			t.Fatalf("plan(id=D-NNN): expected D-001, got isError=%v id=%v (%s)", isErr, plan.Data["id"], plan.Message)
		}
		sess, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": "D-NNN-S1", "content": wgSession("D-NNN-S1", "D-NNN", "Landscape")})
		if isErr || sess.Data["id"] != "D-001-S1" {
			t.Fatalf("session(id=D-NNN-S1): expected D-001-S1, got isError=%v id=%v (%s)", isErr, sess.Data["id"], sess.Message)
		}
		if !strings.Contains(strings.Join(sess.WarningStrings(), "\n"), "W-ID-ALLOCATED: session id D-NNN-S1 resolved to D-001-S1") {
			t.Errorf("expected W-ID-ALLOCATED naming D-001-S1, got %v", sess.Warnings)
		}
		sessData, err := os.ReadFile(filepath.Join(dir, "research", "sessions", "D-001-S1.md"))
		if err != nil {
			t.Fatalf("expected research/sessions/D-001-S1.md: %v", err)
		}
		if fm, _, _ := core.ParseFrontmatter(sessData); fm.GetString("id") != "D-001-S1" || strings.Contains(string(sessData), "D-NNN") {
			t.Errorf("the session must be D-001-S1 and informs D-001, got id %q:\n%s", fm.GetString("id"), sessData)
		}
		next, isErr := wgCall(t, ctx, cs, "vivechak_next", map[string]any{"session_id": "D-NNN-S1"})
		if isErr || next.Data["session_id"] != "D-001-S1" {
			t.Errorf("vivechak_next(session_id=D-NNN-S1): expected D-001-S1, got isError=%v %v (%s)", isErr, next.Data["session_id"], next.Message)
		}
		rec, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-NNN", "auto_draft_from": "D-NNN-S1", "auto_save": true})
		if isErr || rec.Data["id"] != "D-001" || rec.Data["file_path"] != "research/D-001-decision.md" {
			t.Fatalf("record_decision(D-NNN, auto_draft_from=D-NNN-S1): expected D-001 at research/D-001-decision.md, got isError=%v id=%v file=%v (%s)",
				isErr, rec.Data["id"], rec.Data["file_path"], rec.Message)
		}
		adr, _ := os.ReadFile(filepath.Join(dir, "research", "D-001-decision.md"))
		if fm, _, _ := core.ParseFrontmatter(adr); !strings.Contains(strings.Join(fm.GetStringSlice("informed_by_sessions"), ","), "D-001-S1") {
			t.Errorf("the ADR should be informed by D-001-S1, got %v", fm.GetStringSlice("informed_by_sessions"))
		}
		_ = filepath.Walk(filepath.Join(dir, "research"), func(path string, info os.FileInfo, err error) error {
			if err == nil && strings.Contains(strings.ToUpper(info.Name()), "NNN") {
				t.Errorf("a file or folder is named after the placeholder: %s", path)
			}
			return nil
		})
	})

	// 10. The same flow without vivechak_start resolves through the session's plan; a placeholder session id in a
	// session that has no decision plan is refused with E-ID-INVALID naming the fix, and nothing is written.
	t.Run("PlaceholderSessionIDs_ResolveToSessionPlan_OrAreRefused", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		fresh := wgNewProject(t)
		cs0, cleanup0 := startLiveMCPServer(t, ctx, binPath, fresh)
		defer cleanup0()
		env, isErr := wgCall(t, ctx, cs0, "vivechak_save", map[string]any{"kind": "session", "id": "D-NNN-S1", "content": wgSession("D-NNN-S1", "D-NNN", "Landscape")})
		if !isErr || !strings.Contains(env.Message, "E-ID-INVALID") || !strings.Contains(env.Guidance(), "vivechak_next") || !strings.Contains(env.Guidance(), "plan") {
			t.Errorf("expected E-ID-INVALID with the fix (save the plan first / use the id vivechak_next serves), got isError=%v %q / %q", isErr, env.Message, env.Guidance())
		}
		if entries, _ := os.ReadDir(filepath.Join(fresh, "research", "sessions")); len(entries) != 0 {
			t.Errorf("a refused placeholder session wrote %v", entries)
		}

		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()
		plan, _ := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "D-NNN", "content": wgDecisionPlan("D-NNN", "Message queue")})
		did, _ := plan.Data["id"].(string)
		sess, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": "d-nnn-s1", "content": wgSession("d-nnn-s1", "D-NNN", "Landscape")})
		if isErr || sess.Data["id"] != did+"-S1" {
			t.Fatalf("session d-nnn-s1 without vivechak_start: expected %s-S1, got isError=%v id=%v (%s)", did, isErr, sess.Data["id"], sess.Message)
		}
		amend, isErr := wgCall(t, ctx, cs, "vivechak_amend", map[string]any{"id": "D-NNN-S1", "note": "Load test confirmed the write-amplification concern."})
		if isErr || amend.Data["amended_file"] != "research/sessions/"+did+"-S1.md" || !strings.Contains(strings.Join(amend.WarningStrings(), "\n"), "resolved to "+did+"-S1") {
			t.Errorf("vivechak_amend(id=D-NNN-S1): expected research/sessions/%s-S1.md with W-ID-ALLOCATED, got isError=%v file=%v warnings=%v (%s)",
				did, isErr, amend.Data["amended_file"], amend.Warnings, amend.Message)
		}
		review, isErr := wgCall(t, ctx, cs, "vivechak_review", map[string]any{"mode": "red_team", "id": "D-NNN-S1"})
		if isErr || !strings.Contains(review.Message, did+"-S1") || !strings.Contains(strings.Join(review.WarningStrings(), "\n"), "resolved to "+did+"-S1") {
			t.Errorf("vivechak_review(red_team, id=D-NNN-S1): expected the challenge for %s-S1 with W-ID-ALLOCATED, got isError=%v %q %v", did, isErr, review.Message, review.Warnings)
		}
		replan, isErr := wgCall(t, ctx, cs, "vivechak_replan", map[string]any{"plan": did, "op": "update_prompt", "session_id": "D-NNN-S2", "prompt": "Find failure modes, reversal triggers and exit costs."})
		if isErr || !strings.Contains(strings.Join(replan.WarningStrings(), "\n"), "resolved to "+did+"-S2") {
			t.Errorf("vivechak_replan(session_id=D-NNN-S2): expected %s-S2 with W-ID-ALLOCATED, got isError=%v %q %v", did, isErr, replan.Message, replan.Warnings)
		}
		rec, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-NNN", "auto_draft_from": "D-NNN-S1", "auto_save": true})
		if isErr || rec.Data["id"] != did {
			t.Errorf("record_decision(D-NNN) after the session's plan %s: expected %s, got isError=%v id=%v (%s)", did, did, isErr, rec.Data["id"], rec.Message)
		}
	})

	// 11. D-NEW is a placeholder, never an id: two plans saved with id D-NEW (no vivechak_start) get D-001 and D-002,
	// an ADR recorded as D-NEW lands on the decision whose plan the session saved last (D-002), the next one on D-003,
	// a conflict resolution recorded as D-NEW gets an allocated id too, and no file name or content keeps "D-NEW".
	t.Run("DNEWTwice_DistinctIDs_NeverLiteral", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		for _, tc := range []struct{ title, want string }{{"Primary datastore", "D-001"}, {"Message queue", "D-002"}} {
			env, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "D-NEW", "content": wgDecisionPlan("D-NEW", tc.title)})
			if isErr || env.Data["id"] != tc.want || env.Data["file_path"] != "research/"+tc.want+"-plan.md" {
				t.Fatalf("plan %q saved with id D-NEW: expected %s at research/%s-plan.md, got isError=%v id=%v file=%v (%s)", tc.title, tc.want, tc.want, isErr, env.Data["id"], env.Data["file_path"], env.Message)
			}
			if !strings.Contains(strings.Join(env.WarningStrings(), "\n"), "W-ID-ALLOCATED: allocated decision ID "+tc.want) {
				t.Errorf("plan %q: expected W-ID-ALLOCATED naming %s, got %v", tc.title, tc.want, env.Warnings)
			}
		}
		for _, tc := range []struct{ title, want string }{{"Message queue", "D-002"}, {"Cache layer", "D-003"}} {
			env, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-NEW", "content": wgADR("D-NEW", tc.title, "proposed", "two-way")})
			if isErr || env.Data["id"] != tc.want || env.Data["file_path"] != "research/"+tc.want+"-decision.md" {
				t.Fatalf("ADR %q recorded as D-NEW: expected %s at research/%s-decision.md, got isError=%v id=%v file=%v (%s)", tc.title, tc.want, tc.want, isErr, env.Data["id"], env.Data["file_path"], env.Message)
			}
		}
		// A conflict resolution recorded as D-NEW gets an allocated decision id too.
		cr := "---\nid: D-NEW\ndecision_id: D-001\ntitle: Conflict Resolution for Storage Engine\nstatus: resolved\nconflicting_sources:\n  - Session D-001-S1\n  - Session D-001-S2\n" +
			"chosen_option: PostgreSQL\nresolution_method: ACH matrix\ndate: 2026-09-29\n---\n\n# Conflict Resolution: Storage Engine Divergence\n\n" +
			"## 1. Conflict Summary\nD-001-S1 recommended PostgreSQL for ACID compliance; D-001-S2 recommended MongoDB for document flexibility.\n\n" +
			"## 2. Analysis of Competing Hypotheses (ACH Matrix)\nPostgreSQL dominates on ACID transactions and relational joins. Grade A (official docs)\n\n" +
			"## 3. Resolution\nAdopt PostgreSQL with jsonb columns for flexible document attributes.\n"
		env, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-NEW", "artifact_type": "conflict-resolution", "content": cr})
		fp, _ := env.Data["file_path"].(string)
		if isErr || !regexp.MustCompile(`^research/D-\d{3}-conflict-resolution\.md$`).MatchString(fp) {
			t.Errorf("conflict resolution recorded as D-NEW: expected research/D-NNN-conflict-resolution.md, got isError=%v file=%q (%s)", isErr, fp, env.Message)
		}
		for _, f := range wgResearchFiles(t, dir) {
			data, _ := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f)))
			if strings.Contains(strings.ToUpper(f), "NEW") || strings.Contains(string(data), "D-NEW") {
				t.Errorf("the placeholder D-NEW survived in %s", f)
			}
		}

		// Session ids whose own number is a placeholder (no decision part to resolve) are refused, never file names.
		for _, sid := range []string{"T1-NNN", "T1-tbd"} {
			env, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": sid, "content": wgSession(sid, "D-001", "Landscape")})
			if !isErr {
				t.Errorf("session id %q: expected a refusal, got %q", sid, env.Message)
			}
		}
		if entries, _ := os.ReadDir(filepath.Join(dir, "research", "sessions")); len(entries) != 0 {
			t.Errorf("a refused placeholder session id was written: %v", entries)
		}
	})

	// 12. Comparisons saved without an id get C-001, then C-002 (the first file untouched); saving a different
	// comparison under the existing id C-001 is refused with E-REVISE-REQUIRED naming the file's title.
	t.Run("ComparisonsWithoutID_C001ThenC002_IDProtected", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		first := filepath.Join(dir, "research", "C-001-comparison.md")
		var firstContent []byte
		for _, tc := range []struct{ title, a, b, want string }{{"Pino vs Winston", "Pino", "Winston", "C-001"}, {"Zod vs Valibot", "Zod", "Valibot", "C-002"}} {
			env, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "comparison", "content": wgComparison(tc.title, tc.a, tc.b)})
			if isErr || env.Data["id"] != tc.want || env.Data["file_path"] != "research/"+tc.want+"-comparison.md" {
				t.Fatalf("comparison %q without id: expected %s at research/%s-comparison.md, got isError=%v id=%v file=%v (%s)", tc.title, tc.want, tc.want, isErr, env.Data["id"], env.Data["file_path"], env.Message)
			}
			if firstContent == nil {
				firstContent, _ = os.ReadFile(first)
			}
		}
		if now, _ := os.ReadFile(first); string(now) != string(firstContent) {
			t.Errorf("saving C-002 changed C-001")
		}
		env, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "comparison", "id": "C-001", "content": wgComparison("Prisma vs Drizzle", "Prisma", "Drizzle")})
		if !isErr || !strings.Contains(env.Message, "E-REVISE-REQUIRED") || !strings.Contains(env.Message, "Pino vs Winston") {
			t.Errorf("a different comparison under C-001: expected E-REVISE-REQUIRED naming \"Pino vs Winston\", got isError=%v %q", isErr, env.Message)
		}
		if now, _ := os.ReadFile(first); string(now) != string(firstContent) {
			t.Errorf("the refused save changed C-001")
		}
	})

	// 13. The server sets each artifact's id, type and vivechak keys itself: content that omits them (the ADR and plan
	// here) or states other values (a comparison claiming id C-01, type: decision and vivechak: 7; a session claiming
	// id landscape and type: comparison) is saved with the artifact's real id and type and vivechak: 1.
	t.Run("ServerSetsTypeAndVivechak_NotEchoedFromContent", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		cmp := strings.Replace(wgComparison("Pino vs Winston", "Pino", "Winston"), "scope_level: comparison\n", "scope_level: comparison\ntype: decision\nvivechak: 7\n", 1)
		sess := strings.Replace(wgSession("D-001-S1", "D-001", "Landscape"), "status: complete\n", "status: complete\ntype: comparison\n", 1)
		sess = strings.Replace(sess, "id: D-001-S1\n", "id: landscape\n", 1)
		adr := wgADR("D-001", "Primary datastore", "proposed", "two-way", "D-001-S1")
		if fm, _, _ := core.ParseFrontmatter([]byte(adr)); fm.Has("type") || fm.Has("vivechak") {
			t.Fatalf("fixture error: the ADR content must carry neither type nor vivechak")
		}
		calls := []struct {
			tool     string
			args     map[string]any
			file     string
			wantID   string
			wantType string
		}{
			{"vivechak_save", map[string]any{"kind": "comparison", "content": cmp}, "research/C-001-comparison.md", "C-001", "comparison"},
			{"vivechak_save", map[string]any{"kind": "plan", "id": "D-001", "content": wgDecisionPlan("D-001", "Primary datastore")}, "research/D-001-plan.md", "D-001", "plan"},
			{"vivechak_save", map[string]any{"kind": "session", "id": "D-001-S1", "content": sess}, "research/sessions/D-001-S1.md", "D-001-S1", "session"},
			{"vivechak_record_decision", map[string]any{"id": "D-001", "content": adr}, "research/D-001-decision.md", "D-001", "decision"},
		}
		for _, c := range calls {
			if env, isErr := wgCall(t, ctx, cs, c.tool, c.args); isErr || env.Data["file_path"] != c.file {
				t.Fatalf("%s %v: expected %s, got isError=%v file=%v (%s)", c.tool, c.args["kind"], c.file, isErr, env.Data["file_path"], env.Message)
			}
			data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(c.file)))
			if err != nil {
				t.Fatalf("reading %s: %v", c.file, err)
			}
			fm, _, _ := core.ParseFrontmatter(data)
			if fm.GetString("id") != c.wantID || fm.GetString("type") != c.wantType || fm.GetString("vivechak") != "1" {
				t.Errorf("%s: expected id %s, type %s, vivechak 1 set by the server, got id %q type %q vivechak %q", c.file, c.wantID, c.wantType, fm.GetString("id"), fm.GetString("type"), fm.GetString("vivechak"))
			}
		}
	})
}
