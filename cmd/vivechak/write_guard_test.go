package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Fixtures in the shapes the generators ask for (the same shapes the audit harness uses).

func wgComparison(title, a, b string) string {
	return `---
id: C-01
title: "` + title + `"
date: 2026-10-25
status: complete
scope_level: comparison
---
## Prior
- I believe ` + a + ` is faster because it does less work per log line.

## Research Question
Which option fits a Node.js API that logs 5k lines/s?

## Key Findings
- ` + a + ` sustains the target rate [A · single · fresh · direct · fetched](https://example.com/a/bench)
- ` + b + ` has the larger plugin ecosystem [B · single · fresh · direct · fetched](https://example.com/b/plugins)

## Recommendation
Use ` + a + ` [A · single · fresh · direct · fetched](https://example.com/a/docs).

## Weighted Evaluation Matrix
| Criterion | Weight | ` + a + ` | ` + b + ` |
|---|---|---|---|
| Throughput | 0.6 | 5 | 3 |
| Ecosystem | 0.4 | 3 | 5 |
| **Weighted total** | 1.0 | 4.2 | 3.8 |

## Alternatives Considered
` + b + `: rejected for throughput.

## Open Questions & Risks
Throughput under synchronous transports is untested.

## Delta
| Prior Belief | Status | Evidence | Impact |
|---|---|---|---|
| ` + a + ` is faster | Confirmed | benchmark | none |

## Sources & Evidence Ledger
| # | Source | Grade |
|---|---|---|
| 1 | https://example.com/a/docs | A |
`
}

func wgDecisionPlan(id, title string) string {
	return `# ` + id + `-PLAN: ` + title + `

` + "```yaml vivechak-plan" + `
plan: ` + id + `
sessions:
  - {id: ` + id + `-S1, title: Landscape, role: L, depends: []}
  - {id: ` + id + `-S2, title: Falsify, role: F, depends: [` + id + `-S1]}
` + "```" + `

## Routing
| Factor | Score | Rationale |
|---|---|---|
| R | 3 | data migration |

## Sessions

### ` + id + `-S1: Landscape
| ID | Role | Depends | Filename |
|---|---|---|---|
| ` + id + `-S1 | L | none | sessions/` + id + `-S1-landscape.md |

` + "```prompt" + `
# RESEARCH BRIEF: landscape
Map the credible options.
` + "```" + `

### ` + id + `-S2: Falsify
| ID | Role | Depends | Filename |
|---|---|---|---|
| ` + id + `-S2 | F | ` + id + `-S1 | sessions/` + id + `-S2-falsify.md |

` + "```prompt" + `
# RESEARCH BRIEF: how the leading option could fail
Find failure modes and reversal triggers.
` + "```" + `
`
}

func wgSession(id, informs, title string) string {
	return `---
id: ` + id + `
title: "` + title + `"
date: 2026-10-25
status: complete
informs_decisions: [` + informs + `]
---
## Prior
- I believe PostgreSQL fits because the team knows it.

## Research Question
Which primary datastore should we use?

## Key Findings
- PostgreSQL indexes JSONB documents with GIN [A · single · fresh · direct · fetched](https://www.postgresql.org/docs/current/datatype-json.html)
- Logical replication supports zero-downtime upgrades [A · single · fresh · direct · fetched](https://www.postgresql.org/docs/current/logical-replication.html)

## Recommendation
Use PostgreSQL as the primary datastore [A · single · fresh · direct · fetched](https://www.postgresql.org/docs/current/).

## Alternatives Considered
MongoDB: rejected because the team has no operational experience with it.

## Open Questions & Risks
Write amplification under heavy update load.

## Discovered Concerns
Write amplification under heavy update workloads needs a load test before launch.

## Delta
| Prior Belief | Status | Evidence | Impact |
|---|---|---|---|
| PostgreSQL fits | Confirmed | docs | none |

## Sources & Evidence Ledger
| # | Source | Grade |
|---|---|---|
| 1 | https://www.postgresql.org/docs/current/ | A |
`
}

func wgADR(id, title, status, door string, informed ...string) string {
	inf := ""
	if len(informed) > 0 {
		inf = "informed_by_sessions: [" + strings.Join(informed, ", ") + "]\n"
	}
	return `---
id: ` + id + `
title: "` + title + `"
status: ` + status + `
door_type: ` + door + `
date: 2026-10-25
review_trigger: "p95 latency above 50 ms for 7 days"
human_reviewed: false
` + inf + `---
# ` + id + `: ` + title + `
## Context & Problem Statement
We need to choose: ` + title + `.
## Decision Outcome
Chosen option for ` + title + `.
## Rejected Alternatives & Tradeoffs
The main alternative was rejected because it doubles operating cost for this workload.
## Failure Modes & Reversal Triggers
Revisit if p95 latency stays above 50 ms for 7 days.
`
}

const wgPipeline = `# Research Pipeline: Demo

` + "```yaml vivechak-plan" + `
plan: PROJECT
sessions:
  - {id: T1-01, title: Datastore, layer: 1, door: one-way, informs: [D-001], depends: []}
  - {id: T1-02, title: Queue, layer: 1, door: two-way, depends: []}
  - {id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}
decisions:
  - {id: D-001, title: Primary datastore, door: one-way}
` + "```" + `

| ID | Topic | Depends On | Output File |
|---|---|---|---|
| T1-01 | Datastore | — | research/sessions/T1-01-datastore.md |
| T1-02 | Queue | — | research/sessions/T1-02-queue.md |
| SYN-01 | Synthesis | T1-01, T1-02 | research/FAD.md |

## Sessions

### T1-01: Datastore
` + "```prompt" + `
Research datastores. [UPSTREAM_FINDINGS]
` + "```" + `

### T1-02: Queue
` + "```prompt" + `
Research queues. [UPSTREAM_FINDINGS]
` + "```" + `

### SYN-01: Synthesis
` + "```prompt" + `
Synthesize. [ALL_SESSION_FINDINGS]
` + "```" + `
`

// wgCall calls a tool on the live server and returns its envelope and isError flag.
func wgCall(t *testing.T, ctx context.Context, cs *mcp.ClientSession, name string, args map[string]any) (testEnvelope, bool) {
	t.Helper()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: transport error: %v", name, err)
	}
	return parseTestEnvelope(t, res), res.IsError
}

// wgResearchFiles lists the files under dir/research (relative, slash-separated), skipping .history and .locks.
func wgResearchFiles(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	root := filepath.Join(dir, "research")
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		rel = filepath.ToSlash(rel)
		if strings.Contains(rel, "/.history/") || strings.Contains(rel, "/.locks/") {
			return nil
		}
		out = append(out, rel)
		return nil
	})
	return out
}

func wgNewProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testproject\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestWriteGuard_PlanAfterComparison_LiveSubprocess covers CRIT-19 on the built binary over stdio: after a comparison,
// the documented plan save (kind="plan", id, content; no scope) stores a plan, never a comparison, the ADR can be
// recorded, and the gate fails for the unsigned one-way decision (MCP and `vck gate`).
func TestWriteGuard_PlanAfterComparison_LiveSubprocess(t *testing.T) {
	binPath := resolveOrBuildTestBinary(t)

	t.Run("DecisionPlanAfterComparison_ADRRecorded_GateFails", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		wgCall(t, ctx, cs, "vivechak_start", map[string]any{"scope": "comparison", "question": "Pino vs Winston"})
		if env, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "comparison", "id": "C-001", "content": wgComparison("Pino vs Winston", "Pino", "Winston")}); isErr {
			t.Fatalf("comparison save failed: %s", env.Message)
		}
		wgCall(t, ctx, cs, "vivechak_start", map[string]any{"scope": "decision", "question": "Primary datastore"})

		env, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "D-001", "content": wgDecisionPlan("D-001", "Primary datastore")})
		if isErr {
			t.Fatalf("plan save failed: %s", env.Message)
		}
		files := wgResearchFiles(t, dir)
		if !containsString(files, "research/D-001-plan.md") {
			t.Fatalf("expected research/D-001-plan.md, got %v (message %q)", files, env.Message)
		}
		for _, f := range files {
			if strings.Contains(f, "D-001") && strings.HasSuffix(f, "-comparison.md") {
				t.Fatalf("the decision plan was stored as a comparison: %s", f)
			}
		}

		next, _ := wgCall(t, ctx, cs, "vivechak_next", map[string]any{})
		if next.Data["session_id"] != "D-001-S1" {
			t.Fatalf("expected vivechak_next to serve D-001-S1, got %v (%s)", next.Data["session_id"], next.Message)
		}
		for _, s := range []struct{ id, title string }{{"D-001-S1", "Landscape"}, {"D-001-S2", "Falsify"}} {
			if e, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": s.id, "content": wgSession(s.id, "D-001", s.title)}); isErr {
				t.Fatalf("session %s save failed: %s", s.id, e.Message)
			}
		}
		rec, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{
			"id":      "D-001",
			"content": wgADR("D-001", "Primary datastore", "accepted", "one-way", "D-001-S1", "D-001-S2"),
		})
		if isErr {
			t.Fatalf("record_decision failed: %s", rec.Message)
		}
		if fp, _ := rec.Data["file_path"].(string); fp != "research/D-001-decision.md" {
			t.Errorf("expected the ADR at research/D-001-decision.md, got %q", fp)
		}
		if _, err := os.Stat(filepath.Join(dir, "research", "D-001-decision.md")); err != nil {
			t.Fatalf("ADR file missing: %v", err)
		}

		gate, _ := wgCall(t, ctx, cs, "vivechak_gate", map[string]any{})
		if gate.Data["verdict"] != "FAIL" {
			t.Errorf("expected the MCP gate to FAIL for the unsigned one-way decision, got verdict %v (%s)", gate.Data["verdict"], gate.Message)
		}
		cmd := exec.Command(binPath, "gate", dir)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			t.Errorf("expected `vck gate` to exit 1, got err=%v\n%s", err, out)
		}
	})

	t.Run("ProjectPlanDAGAfterComparison", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "comparison", "content": wgComparison("Zod vs Valibot", "Zod", "Valibot")})
		env, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "DAG", "content": wgPipeline})
		if isErr {
			t.Fatalf("pipeline save failed: %s", env.Message)
		}
		if _, err := os.Stat(filepath.Join(dir, "research", "RESEARCH-PIPELINE.md")); err != nil {
			t.Fatalf("expected research/RESEARCH-PIPELINE.md: %v (files %v)", err, wgResearchFiles(t, dir))
		}
		next, _ := wgCall(t, ctx, cs, "vivechak_next", map[string]any{})
		if sid, _ := next.Data["session_id"].(string); !strings.HasPrefix(sid, "T1-0") {
			t.Fatalf("expected vivechak_next to serve a T1 session, got %v (%s)", next.Data["session_id"], next.Message)
		}
	})

	// Variant the audit check does not sample: no id at all after a comparison (the `decision` and `project` prompts
	// tell the agent to call vivechak_save with kind="plan" and content only); the plan's own shape decides.
	t.Run("PlanWithoutIDAfterComparison_RoutedByShape", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "comparison", "content": wgComparison("Pino vs Winston", "Pino", "Winston")})
		env, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "content": wgDecisionPlan("D-NEW", "Message queue")})
		if isErr {
			t.Fatalf("decision plan save without id failed: %s", env.Message)
		}
		id, _ := env.Data["id"].(string)
		if !strings.HasPrefix(id, "D-") || !containsString(wgResearchFiles(t, dir), "research/"+id+"-plan.md") {
			t.Fatalf("expected a decision plan research/D-NNN-plan.md, got id %q, files %v", id, wgResearchFiles(t, dir))
		}

		env, isErr = wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "content": wgPipeline})
		if isErr {
			t.Fatalf("pipeline save without id failed: %s", env.Message)
		}
		if _, err := os.Stat(filepath.Join(dir, "research", "RESEARCH-PIPELINE.md")); err != nil {
			t.Fatalf("expected research/RESEARCH-PIPELINE.md: %v (files %v)", err, wgResearchFiles(t, dir))
		}
		for _, f := range wgResearchFiles(t, dir) {
			if strings.HasSuffix(f, "-comparison.md") && f != "research/C-001-comparison.md" {
				t.Errorf("a plan was stored as a comparison: %s", f)
			}
		}
	})

	t.Run("ComparisonIDOrScopeWithKindPlanRefused", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		env, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "C-001", "content": wgComparison("A vs B", "Alpha", "Beta")})
		if !isErr || !strings.Contains(env.Message, "E-ID-INVALID") {
			t.Errorf("expected kind=plan with a comparison id to fail with E-ID-INVALID, got isError=%v %q", isErr, env.Message)
		}
		env, isErr = wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "scope": "comparison", "content": wgComparison("A vs B", "Alpha", "Beta")})
		if !isErr || (!strings.Contains(env.Message, "kind=\"comparison\"") && !strings.Contains(env.Guidance(), "kind=\"comparison\"")) {
			t.Errorf("expected kind=plan with scope=comparison to be refused and point at kind=\"comparison\", got isError=%v %q / %q", isErr, env.Message, env.Guidance())
		}
		for _, f := range wgResearchFiles(t, dir) {
			if strings.HasSuffix(f, ".md") {
				t.Errorf("a refused plan save wrote %s", f)
			}
		}
	})

	// resolveDecisionFilename never adopts a comparison or a plan that happens to start with the decision id
	// (the file the 12b768f build wrote for the repro above).
	t.Run("ComparisonNamedLikeDecisionIsNotAdopted", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		if err := os.MkdirAll(filepath.Join(dir, "research"), 0o755); err != nil {
			t.Fatal(err)
		}
		misSaved := strings.Replace(wgComparison("Primary datastore", "PostgreSQL", "MongoDB"), "id: C-01\n", "id: D-001\ntype: comparison\nvivechak: 1\n", 1)
		misPath := filepath.Join(dir, "research", "D-001-comparison.md")
		if err := os.WriteFile(misPath, []byte(misSaved), 0o644); err != nil {
			t.Fatal(err)
		}
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		rec, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{
			"id":      "D-001",
			"content": wgADR("D-001", "Primary datastore", "proposed", "two-way"),
		})
		if isErr {
			t.Fatalf("record_decision failed: %s", rec.Message)
		}
		if fp, _ := rec.Data["file_path"].(string); fp != "research/D-001-decision.md" {
			t.Errorf("expected research/D-001-decision.md, got %q", fp)
		}
		if got, _ := os.ReadFile(misPath); string(got) != misSaved {
			t.Errorf("the comparison file was modified by record_decision")
		}
	})
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// wgHistory lists the files in dir/research/.history.
func wgHistory(t *testing.T, dir string) []string {
	t.Helper()
	entries, _ := os.ReadDir(filepath.Join(dir, "research", ".history"))
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// wgHasArchive reports whether research/.history holds an archived copy of research/<name> with exactly content.
func wgHasArchive(t *testing.T, dir, name, content string) bool {
	t.Helper()
	for _, h := range wgHistory(t, dir) {
		if !strings.HasPrefix(h, name+"@") {
			continue
		}
		if data, err := os.ReadFile(filepath.Join(dir, "research", ".history", h)); err == nil && string(data) == content {
			return true
		}
	}
	return false
}

// TestWriteGuard_LiveSubprocess verifies that the stdio MCP server enforces
// write guards (E-REVISE-REQUIRED on duplicate saves, .history/ archival on revise: true,
// draft status on ADR auto_save with placeholders, and DECISIONS.md preamble preservation).
func TestWriteGuard_LiveSubprocess(t *testing.T) {
	binPath := resolveOrBuildTestBinary(t)

	// 1. Refuse second save of different content without revise: true
	t.Run("Refuse_SecondSave_DifferentContent", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testproject\n"), 0o644)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		startRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_start",
			Arguments: map[string]any{
				"scope":    "comparison",
				"question": "Pino vs Winston",
			},
		})
		if err != nil || startRes.IsError {
			t.Fatalf("vivechak_start failed: %v", err)
		}

		contentV1 := `---
id: C-001
type: comparison
title: Pino vs Winston
status: draft
---
# Pino vs Winston
Version 1 comparison body text.
`
		save1Res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "comparison",
				"id":      "C-001",
				"content": contentV1,
			},
		})
		if err != nil || save1Res.IsError {
			t.Fatalf("first save failed: %v", err)
		}

		filePath := filepath.Join(dir, "research", "C-001-comparison.md")
		readV1, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("reading saved file: %v", err)
		}
		if !strings.Contains(string(readV1), "Version 1") {
			t.Errorf("expected Version 1 in saved file, got: %s", string(readV1))
		}

		// Second save with differing content and NO revise flag -> must be refused with E-REVISE-REQUIRED
		contentV2 := `---
id: C-001
type: comparison
title: Pino vs Winston
status: draft
---
# Pino vs Winston
Version 2 differing content.
`
		save2Res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "comparison",
				"id":      "C-001",
				"content": contentV2,
			},
		})
		if err != nil {
			t.Fatalf("call error on second save: %v", err)
		}
		if !save2Res.IsError {
			env := parseTestEnvelope(t, save2Res)
			if env.Ok {
				t.Fatalf("expected second save to fail with E-REVISE-REQUIRED, got success")
			}
			if !strings.Contains(env.Message, "E-REVISE-REQUIRED") {
				t.Errorf("expected E-REVISE-REQUIRED in message, got: %s", env.Message)
			}
		} else {
			// Transport error text
			env := parseTestEnvelope(t, save2Res)
			if !strings.Contains(env.Message, "E-REVISE-REQUIRED") {
				t.Errorf("expected E-REVISE-REQUIRED in error message, got: %s", env.Message)
			}
		}

		// Verify file remains unchanged
		readAfter, _ := os.ReadFile(filePath)
		if !strings.Contains(string(readAfter), "Version 1") {
			t.Errorf("file was overwritten despite refusal: %s", string(readAfter))
		}

		// 2. Saving with revise: true archives old version to research/.history/
		saveReviseRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "comparison",
				"id":      "C-001",
				"content": contentV2,
				"revise":  true,
			},
		})
		if err != nil || saveReviseRes.IsError {
			t.Fatalf("save with revise: true failed: %v", err)
		}

		readV2, _ := os.ReadFile(filePath)
		if !strings.Contains(string(readV2), "Version 2") {
			t.Errorf("expected Version 2 in revised file, got: %s", string(readV2))
		}

		// Check .history directory for archived version
		historyDir := filepath.Join(dir, "research", ".history")
		entries, hErr := os.ReadDir(historyDir)
		if hErr != nil || len(entries) == 0 {
			t.Fatalf("expected archived version in research/.history, got err=%v, count=%d", hErr, len(entries))
		}
	})

	// 3. auto_save draft ADR with placeholders saves status: proposed, response status: draft,
	// and subsequent auto_save archives previous version into research/.history/
	t.Run("AutoSave_DraftWithPlaceholders_ThenSecondAutoSave", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testproject\n"), 0o644)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		// Start decision scope
		_, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_start",
			Arguments: map[string]any{
				"scope":    "decision",
				"question": "Postgres vs MySQL",
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Save decision plan
		planContent := livePlan("Plan for D-001", "D-001", "", []string{
			"{id: D-001-S1, title: Datastore Landscape, role: L, depends: []}",
		}, liveSection("D-001-S1", "Datastore Landscape", "Map the credible datastores."))
		planRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "plan",
				"id":      "D-001",
				"content": planContent,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if planRes.IsError {
			t.Fatalf("decision plan save refused: %+v", planRes.Content)
		}

		// Save session D-001-S1
		sessionContent := `---
id: D-001-S1
session_id: D-001-S1
type: session
title: Datastore Landscape
status: complete
informs_decisions: ["D-001"]
vivechak: 1
---
# Datastore Landscape
## Key Findings
- PostgreSQL has advanced JSONB indexing [Grade A: https://postgresql.org | fetched].
- MySQL has simpler replication setup [Grade B: https://mysql.com | fetched].

## Recommendation
We select PostgreSQL [Grade A: https://postgresql.org | fetched].

## Discovered Concerns & Tradeoffs
- PostgreSQL vacuum tuning requires operational monitoring at high write volumes.

## Sources & Evidence Ledger
- PostgreSQL Docs: https://postgresql.org [Grade A | fetched]
- MySQL Docs: https://mysql.com [Grade B | fetched]
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

		// Call vivechak_record_decision with auto_draft_from and auto_save: true
		recRes1, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"id":              "D-001",
				"auto_draft_from": "D-001-S1",
				"auto_save":       true,
			},
		})
		if err != nil || recRes1.IsError {
			envFail := parseTestEnvelope(t, recRes1)
			t.Fatalf("first auto_save record_decision failed (err=%v, isError=%v, msg=%s, data=%v)", err, recRes1.IsError, envFail.Message, envFail.Data)
		}

		env1 := parseTestEnvelope(t, recRes1)
		if !env1.Ok {
			t.Fatalf("expected envelope ok: true, got false (%s)", env1.Message)
		}
		if env1.Data["status"] != "draft" {
			t.Errorf("expected data.status == 'draft' when placeholders remain, got: %v", env1.Data["status"])
		}
		if phs, ok := env1.Data["placeholders"].([]any); !ok || len(phs) == 0 {
			t.Errorf("expected data.placeholders to have items, got: %v", env1.Data["placeholders"])
		}

		// Check saved file on disk: status must be proposed
		adrPath := filepath.Join(dir, "research", "D-001-decision.md")
		adrBytes, err := os.ReadFile(adrPath)
		if err != nil {
			t.Fatalf("reading saved ADR file: %v", err)
		}
		if !strings.Contains(string(adrBytes), "status: proposed") {
			t.Errorf("expected saved ADR to have status: proposed, got:\n%s", string(adrBytes))
		}

		// Call auto_save a second time on the existing ADR
		recRes2, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"id":              "D-001",
				"auto_draft_from": "D-001-S1",
				"auto_save":       true,
			},
		})
		if err != nil || recRes2.IsError {
			t.Fatalf("second auto_save record_decision failed: %v", err)
		}

		// The ADR's own previous version is archived in research/.history/ (the DECISIONS.md archive written by the
		// registry recompile does not count).
		if !wgHasArchive(t, dir, "D-001-decision.md", string(adrBytes)) {
			t.Fatalf("expected research/.history to hold the previous D-001-decision.md, got %v", wgHistory(t, dir))
		}
	})

	// 4. Text outside DECISIONS.md boundary markers survives vivechak_record_decision
	t.Run("DecisionsMDPreamblePreservation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testproject\n"), 0o644)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		_ = os.MkdirAll(filepath.Join(dir, "research"), 0o755)
		decisionsFile := filepath.Join(dir, "research", "DECISIONS.md")

		initialContent := `# Custom Architecture Team Decisions Handbook

This preamble contains important guidelines for all architectural decisions.
Do not overwrite these custom instructions.

<!-- VIVECHAK:START -->
# Decision Registry
| ID | Title | Status |
|---|---|---|
<!-- VIVECHAK:END -->

## Postamble Appendix
Signed off by Architecture Guild.
`
		if err := os.WriteFile(decisionsFile, []byte(initialContent), 0o644); err != nil {
			t.Fatal(err)
		}

		adrContent := `---
id: D-001
type: decision
title: Custom Framework Choice
door_type: two-way
status: accepted
review_trigger: "Throughput drops below 1000 req/s"
date: 2026-11-04
vivechak: 1
---
# Custom Framework Choice
## Context
We need a lightweight framework.
## Decision Outcome
We choose Go stdlib.
`
		recRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"id":      "D-001",
				"content": adrContent,
			},
		})
		if err != nil || recRes.IsError {
			envFail := parseTestEnvelope(t, recRes)
			t.Fatalf("record_decision failed (err=%v, isError=%v, msg=%s, data=%v)", err, recRes.IsError, envFail.Message, envFail.Data)
		}

		afterBytes, err := os.ReadFile(decisionsFile)
		if err != nil {
			t.Fatalf("reading DECISIONS.md after record: %v", err)
		}
		afterStr := string(afterBytes)

		if !strings.Contains(afterStr, "Custom Architecture Team Decisions Handbook") {
			t.Errorf("preamble was lost in DECISIONS.md:\n%s", afterStr)
		}
		if !strings.Contains(afterStr, "This preamble contains important guidelines") {
			t.Errorf("preamble body was lost in DECISIONS.md:\n%s", afterStr)
		}
		if !strings.Contains(afterStr, "Postamble Appendix") {
			t.Errorf("postamble was lost in DECISIONS.md:\n%s", afterStr)
		}
		if !strings.Contains(afterStr, "D-001") {
			t.Errorf("D-001 entry missing inside DECISIONS.md markers:\n%s", afterStr)
		}
	})

	// 5. The documented auto_save call over an existing ADR whose content differs (a later session arrives after the
	// agent edited the ADR, CRIT-18): it saves the redraft as a revision, never E-REVISE-REQUIRED, and archives that
	// ADR's previous version (the edited one) in research/.history/.
	t.Run("AutoSave_OverExistingADR_DifferentContent_ArchivesThatADR", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		mustCall := func(name string, args map[string]any) testEnvelope {
			t.Helper()
			env, isErr := wgCall(t, ctx, cs, name, args)
			if isErr {
				t.Fatalf("%s failed: %s", name, env.Message)
			}
			return env
		}
		mustCall("vivechak_start", map[string]any{"scope": "decision", "question": "Primary datastore"})
		mustCall("vivechak_save", map[string]any{"kind": "plan", "id": "D-001", "content": wgDecisionPlan("D-001", "Primary datastore")})
		mustCall("vivechak_save", map[string]any{"kind": "session", "id": "D-001-S1", "content": wgSession("D-001-S1", "D-001", "Landscape")})
		mustCall("vivechak_record_decision", map[string]any{"id": "D-001", "auto_draft_from": "D-001-S1", "auto_save": true})
		adrPath := filepath.Join(dir, "research", "D-001-decision.md")
		draft, err := os.ReadFile(adrPath)
		if err != nil {
			t.Fatalf("auto-saved ADR missing: %v", err)
		}
		edited := strings.Replace(string(draft), "[Condition or date for mandatory re-evaluation]", "p95 write latency above 50 ms for 7 days", 1)
		if edited == string(draft) {
			t.Fatalf("the auto-saved draft has no review-trigger placeholder to edit:\n%s", draft)
		}
		mustCall("vivechak_record_decision", map[string]any{"id": "D-001", "content": edited, "revise": true})
		editedOnDisk, _ := os.ReadFile(adrPath)

		mustCall("vivechak_save", map[string]any{"kind": "session", "id": "D-001-S2", "content": wgSession("D-001-S2", "D-001", "Falsify")})
		rec, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-001", "auto_draft_from": "D-001-S2", "auto_save": true})
		if isErr || rec.Data["file_path"] != "research/D-001-decision.md" {
			t.Fatalf("auto_save over the existing, edited ADR: expected a revision of research/D-001-decision.md, got isError=%v file=%v (%s)", isErr, rec.Data["file_path"], rec.Message)
		}
		after, _ := os.ReadFile(adrPath)
		if string(after) == string(editedOnDisk) || !strings.Contains(string(after), "D-001-S2") {
			t.Errorf("expected the ADR redrafted from D-001-S2, got:\n%s", after)
		}
		if !wgHasArchive(t, dir, "D-001-decision.md", string(editedOnDisk)) {
			t.Errorf("expected research/.history to hold the edited D-001-decision.md the auto_save replaced, got %v", wgHistory(t, dir))
		}
	})

	// 6. `kind` is required: vivechak_save's input schema lists it, and a save without it is refused (here one whose
	// C- id would otherwise let the handler guess a comparison) and writes nothing.
	t.Run("KindRequired_SchemaAndCall", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		var required []string
		for tool, err := range cs.Tools(ctx, nil) {
			if err != nil {
				t.Fatalf("listing tools: %v", err)
			}
			if tool.Name != "vivechak_save" {
				continue
			}
			raw, _ := json.Marshal(tool.InputSchema)
			var schema struct {
				Required []string `json:"required"`
			}
			if err := json.Unmarshal(raw, &schema); err != nil {
				t.Fatalf("decoding vivechak_save's input schema %s: %v", raw, err)
			}
			required = schema.Required
		}
		if !containsString(required, "kind") || !containsString(required, "content") {
			t.Errorf("vivechak_save's input schema must require kind and content, got required=%v", required)
		}

		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "vivechak_save", Arguments: map[string]any{
			"id":      "C-001",
			"content": wgComparison("Pino vs Winston", "Pino", "Winston"),
		}})
		if err != nil {
			t.Fatalf("transport error: %v", err)
		}
		text := ""
		if len(res.Content) > 0 {
			if tc, ok := res.Content[0].(*mcp.TextContent); ok {
				text = tc.Text
			}
		}
		if !res.IsError || !strings.Contains(text, "kind") {
			t.Errorf("a save without kind must be refused naming kind, got isError=%v %q", res.IsError, text)
		}
		if files := wgResearchFiles(t, dir); len(files) != 0 {
			t.Errorf("a refused save wrote %v", files)
		}
	})
}

// TestWriteGuard_DraftTakesPlanTitle_LiveSubprocess covers HIGH-13 on the built binary over stdio: with no ADR yet,
// the auto-drafted ADR takes the plan's decision title (never the fixed "Research Plan"), and the plan's own title
// comes from its heading or plan block, else stays the visible placeholder "[Decision title]".
func TestWriteGuard_DraftTakesPlanTitle_LiveSubprocess(t *testing.T) {
	binPath := resolveOrBuildTestBinary(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dir := wgNewProject(t)
	cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
	defer cleanup()

	planTitle := func(id string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(dir, "research", id+"-plan.md"))
		if err != nil {
			t.Fatalf("reading %s plan: %v", id, err)
		}
		fm, _, _ := core.ParseFrontmatter(data)
		return fm.GetString("title")
	}
	draftTitle := func(id string) string {
		t.Helper()
		wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": id + "-S1", "content": wgSession(id+"-S1", id, "Landscape")})
		env, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": id, "auto_draft_from": id + "-S1"})
		if isErr {
			t.Fatalf("auto-draft %s failed: %s", id, env.Message)
		}
		draft, _ := env.Data["draft_content"].(string)
		fm, _, _ := core.ParseFrontmatter([]byte(draft))
		return fm.GetString("title")
	}

	// The audit's case: a plan headed "# D-001-PLAN: Event storage engine".
	if _, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "D-001", "content": wgDecisionPlan("D-001", "Event storage engine")}); isErr {
		t.Fatal("plan D-001 save failed")
	}
	if got := planTitle("D-001"); got != "Event storage engine" {
		t.Errorf("plan D-001 title = %q, want the heading's title", got)
	}
	if got := draftTitle("D-001"); got != "Event storage engine" {
		t.Errorf("with no ADR yet the draft title = %q, want the plan's title", got)
	}
	rec, isErr := wgCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-001", "auto_draft_from": "D-001-S1", "auto_save": true})
	if isErr {
		t.Fatalf("auto_save failed: %s", rec.Message)
	}
	adr, _ := os.ReadFile(filepath.Join(dir, "research", "D-001-decision.md"))
	if fm, _, _ := core.ParseFrontmatter(adr); fm.GetString("title") != "Event storage engine" {
		t.Errorf("auto-saved ADR title = %q, want the plan's title", fm.GetString("title"))
	}

	// Variant: no heading, the decision title only in the plan block.
	blockOnly := strings.Replace(wgDecisionPlan("D-002", "unused"), "# D-002-PLAN: unused\n", "", 1)
	blockOnly = strings.Replace(blockOnly, "plan: D-002\n", "plan: D-002\ndecisions:\n  - {id: D-002, title: Message queue, door: two-way}\n", 1)
	if _, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "D-002", "content": blockOnly}); isErr {
		t.Fatal("plan D-002 save failed")
	}
	if got := planTitle("D-002"); got != "Message queue" {
		t.Errorf("plan D-002 title = %q, want the plan block's decision title", got)
	}
	if got := draftTitle("D-002"); got != "Message queue" {
		t.Errorf("D-002 draft title = %q, want the plan block's decision title", got)
	}

	// Variant: the plan declares no title at all: a visible placeholder, never a fixed string.
	untitled := strings.Replace(wgDecisionPlan("D-003", "unused"), "# D-003-PLAN: unused\n", "", 1)
	if _, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "D-003", "content": untitled}); isErr {
		t.Fatal("plan D-003 save failed")
	}
	if got := planTitle("D-003"); got != "[Decision title]" {
		t.Errorf("an untitled plan's title = %q, want the placeholder [Decision title]", got)
	}
	if got := draftTitle("D-003"); got == "Research Plan" || got == "" {
		t.Errorf("D-003 draft title = %q; a fixed or empty title must never be used", got)
	}
}
