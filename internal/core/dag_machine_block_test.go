package core

import (
	"strings"
	"testing"
)

func TestParsePlanMachineBlock_IgnoresMarkdownTables(t *testing.T) {
	planWithTables := md(`---
project: Test Machine Block
scope: project
---
# Research Pipeline: Test Machine Block

Here is an architectural options table that confused older parsers:

| Technology | Latency | Throughput | Notes |
|---|---|---|---|
| PostgreSQL | 2ms | High | Relational |
| Redis | 0.5ms | Ultra High | Cache |

And another overview table:

| Session | Focus |
|---|---|
| SEC-01 | Auth Security |
| PERF-02 | Performance Tuning |

The actual DAG is defined strictly in the machine block below:

<F>yaml vivechak-plan
plan: PROJECT
sessions:
  - id: T1-01
    title: Datastore Architecture
    door: one-way
    informs: [D-001]
    depends: []
  - id: T1-02
    title: Caching Strategy
    door: two-way
    depends: [T1-01]
  - id: T1-03
    title: Search Engine
    door: two-way
    depends: [T1-01]
  - id: T1-04
    title: Observability
    door: two-way
    depends: []
  - id: SYN-01
    title: Founding Architecture
    kind: synthesis
    depends: [ALL]
decisions:
  - id: D-001
    title: Datastore Selection
    door: one-way
<F>

## Sessions

### T1-01: Datastore Architecture
<F>prompt
Prompt content for T1-01.
<F>

### T1-02: Caching Strategy
<F>prompt
Prompt content for T1-02.
<F>

### T1-03: Search Engine
<F>prompt
Prompt content for T1-03.
<F>

### T1-04: Observability
<F>prompt
Prompt content for T1-04.
<F>

### SYN-01: Founding Architecture
<F>prompt
Prompt content for SYN-01. [ALL_SESSION_FINDINGS]
<F>
`)

	pipeline, err := ParsePipeline([]byte(planWithTables))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	// Exactly the block's five sessions, each with its own prompt (the prompts land on the returned sessions,
	// not on copies taken while the session slice grew).
	want := []struct{ id, title, prompt string }{
		{"T1-01", "Datastore Architecture", "Prompt content for T1-01."},
		{"T1-02", "Caching Strategy", "Prompt content for T1-02."},
		{"T1-03", "Search Engine", "Prompt content for T1-03."},
		{"T1-04", "Observability", "Prompt content for T1-04."},
		{"SYN-01", "Founding Architecture", "Prompt content for SYN-01. [ALL_SESSION_FINDINGS]"},
	}
	if len(pipeline.Sessions) != len(want) {
		t.Fatalf("expected exactly %d sessions from machine block, got %d: %v", len(want), len(pipeline.Sessions), idsOf(pipeline.Sessions))
	}
	for i, w := range want {
		s := pipeline.Sessions[i]
		if s.ID != w.id || s.Title != w.title {
			t.Errorf("session %d mismatch: %+v", i, s)
		}
		if s.Prompt != w.prompt {
			t.Errorf("session %s prompt = %q, want %q", w.id, s.Prompt, w.prompt)
		}
	}
	if deps := pipeline.SessionByID("SYN-01").Dependencies; strings.Join(deps, ",") != "T1-01,T1-02,T1-03,T1-04" {
		t.Errorf("SYN-01 dependencies = %v", deps)
	}
	if res := ValidatePlan([]byte(planWithTables)); len(res.Issues) != 0 {
		t.Errorf("expected no issues, got %+v", res.Issues)
	}

	// Verify decisions were parsed from machine block
	block, err := ParsePlanMachineBlock([]byte(planWithTables))
	if err != nil {
		t.Fatalf("ParsePlanMachineBlock failed: %v", err)
	}
	if block == nil || len(block.Decisions) != 1 || block.Decisions[0].ID != "D-001" {
		t.Errorf("decisions mismatch in machine block: %+v", block)
	}
}

func TestDraftDecisionFromSessionWithExisting_PreservesUserEdits(t *testing.T) {
	sessionContent := `---
session_id: T1-01
title: Raw Session Datastore Evaluation
status: complete
date: 2026-10-15
---
# Datastore Evaluation

## Executive Summary
Chose PostgreSQL for transactional integrity. Grade A (canonical docs: https://postgresql.org).

## Discovered Concerns
Connection scaling requires connection pooler under high load.
`

	existingADR := `---
decision_id: D-001
title: Custom Curated Datastore Architecture Title
status: draft
door_type: one-way
review_trigger: "When write TPS exceeds 20,000"
date: 2026-10-14
---
# Decision: Custom Curated Datastore Architecture Title

## Context
Initial context notes.
`

	merged, err := DraftDecisionFromSessionWithExisting([]byte(existingADR), []byte(sessionContent), "D-001")
	if err != nil {
		t.Fatalf("DraftDecisionFromSessionWithExisting failed: %v", err)
	}

	mergedStr := string(merged)

	// User-edited title must be preserved
	if !strings.Contains(mergedStr, "Custom Curated Datastore Architecture Title") {
		t.Errorf("user-edited title was overwritten: %s", mergedStr)
	}

	// User-edited review_trigger must be preserved
	if !strings.Contains(mergedStr, "When write TPS exceeds 20,000") {
		t.Errorf("user-edited review trigger was overwritten: %s", mergedStr)
	}

	// Session content / findings must be incorporated
	if !strings.Contains(mergedStr, "PostgreSQL for transactional integrity") {
		t.Errorf("session findings were not incorporated: %s", mergedStr)
	}
}

func TestReplanInPlace_PreservesCustomProse(t *testing.T) {
	origPlan := md(`# Top-Level Architectural Manifesto
This is a very specific manifesto written by the architect that MUST NOT be lost.

Here are some non-standard custom sections:
- Custom guideline 1
- Custom guideline 2

<F>yaml vivechak-plan
plan: PROJECT
sessions:
  - id: T1-01
    title: Initial Title
    door: two-way
    depends: []
  - id: SYN-01
    title: Synthesis
    kind: synthesis
    depends: [ALL]
<F>

## Sessions

### T1-01: Initial Title
Original instructions for the agent, written by the planner.

<F>prompt
Original prompt instructions.
<F>

### SYN-01: Synthesis
<F>prompt
Synthesize. [ALL_SESSION_FINDINGS]
<F>

## Bottom Section
Final notes and signing instructions.
`)

	params := ReplanParams{
		SessionID:    "T1-02",
		Topic:        "Caching Strategy",
		DoorType:     "two-way",
		Dependencies: []string{"T1-01"},
		Prompt:       "Conduct in-depth caching strategy research.",
	}

	result, err := ReplanInPlace([]byte(origPlan), "add_session", params)
	if err != nil {
		t.Fatalf("ReplanInPlace failed: %v", err)
	}

	resStr := string(result)
	for _, kept := range []string{"# Top-Level Architectural Manifesto\nThis is a very specific manifesto written by the architect that MUST NOT be lost.",
		"- Custom guideline 1\n- Custom guideline 2", "Original instructions for the agent, written by the planner.",
		"## Bottom Section\nFinal notes and signing instructions.\n"} {
		if !strings.Contains(resStr, kept) {
			t.Errorf("planner prose %q was lost:\n%s", kept, resStr)
		}
	}
	// The new block entry (block style, like its siblings) and section go before the synthesis session.
	if !strings.Contains(resStr, "  - id: T1-02\n    title: Caching Strategy\n    door: two-way\n    depends: [T1-01]\n  - id: SYN-01") {
		t.Errorf("block entry for T1-02 missing or not in the planner's style:\n%s", resStr)
	}
	if !strings.Contains(resStr, md("### T1-02: Caching Strategy\n\n<F>prompt\nConduct in-depth caching strategy research.\n<F>\n\n### SYN-01: Synthesis")) {
		t.Errorf("prompt section for T1-02 missing or misplaced:\n%s", resStr)
	}
	dag, err := ParsePipeline(result)
	if err != nil {
		t.Fatalf("replanned plan no longer parses: %v", err)
	}
	if got := dag.SessionByID("SYN-01").Dependencies; strings.Join(got, ",") != "T1-01,T1-02" {
		t.Errorf("SYN-01 (depends: [ALL]) should now wait for T1-02 too, got %v", got)
	}
}

// replan edits only the session's block entry (in the style and with the keys the planner wrote) and its one prompt
// section; every other byte, including planner prose and tables, stays as written (HIGH-07).
func TestReplanInPlace_EditsOnlyTheSessionsEntryAndPromptSection(t *testing.T) {
	plan := md(`# Research Pipeline: Styles

Planner note: T1-02 waits for T1-01 because the queue must match the datastore's consistency model.

<F>yaml vivechak-plan
plan: PROJECT
sessions:
  - {id: T1-01, title: Datastore, layer: 1, door: one-way, informs: [D-001], depends: []}
  - id: T1-02
    title: Queue
    door: two-way
    depends:
      - T1-01
  - {id: T1-03, title: Auth}
  - {id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}
decisions:
  - {id: D-001, title: Primary datastore, door: one-way}
<F>

| ID | Depends |
|---|---|
| T1-02 | T1-01 |

## Sessions

### T1-01: Datastore
Context: pick the primary store.

<F>prompt
Research datastores.
<F>

### T1-02: Queue
<F>prompt
Research queues.
<F>

### T1-03: Auth
<F>prompt
Research auth.
<F>

### SYN-01: Synthesis
<F>prompt
Synthesize. [ALL_SESSION_FINDINGS]
<F>
`)
	replan := func(t *testing.T, in, op string, p ReplanParams) string {
		t.Helper()
		out, err := ReplanInPlace([]byte(in), op, p)
		if err != nil {
			t.Fatalf("%s %s: %v", op, p.SessionID, err)
		}
		return string(out)
	}
	expect := func(t *testing.T, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("replanned plan differs from the expected edit:\n--- got ---\n%s\n--- want ---\n%s", got, want)
		}
	}
	swap := func(old, new string) string {
		if strings.Count(plan, old) != 1 {
			t.Fatalf("fixture must contain %q exactly once", old)
		}
		return strings.Replace(plan, old, new, 1)
	}

	t.Run("update_prompt changes only the prompt fence", func(t *testing.T) {
		got := replan(t, plan, "update_prompt", ReplanParams{SessionID: "T1-01", Prompt: "Evaluate Oracle for enterprise compliance."})
		expect(t, got, swap("Research datastores.", "Evaluate Oracle for enterprise compliance."))
	})
	t.Run("update_prompt lengthens the fence for a prompt with a fence", func(t *testing.T) {
		prompt := md("Example:\n<F>sql\nSELECT 1;\n<F>")
		got := replan(t, plan, "update_prompt", ReplanParams{SessionID: "T1-03", Prompt: prompt})
		expect(t, got, swap(md("<F>prompt\nResearch auth.\n<F>"), "````prompt\n"+prompt+"\n````"))
		if s := mustParse(t, got).SessionByID("T1-03"); s.Prompt != prompt {
			t.Errorf("T1-03 prompt = %q, want %q", s.Prompt, prompt)
		}
	})
	t.Run("update_deps on a flow entry without depends adds the key in flow style", func(t *testing.T) {
		got := replan(t, plan, "update_deps", ReplanParams{SessionID: "T1-03", Dependencies: []string{"T1-01"}})
		expect(t, got, swap("  - {id: T1-03, title: Auth}", "  - {id: T1-03, title: Auth, depends: [T1-01]}"))
	})
	t.Run("update_deps on a flow list keeps the other keys", func(t *testing.T) {
		got := replan(t, plan, "update_deps", ReplanParams{SessionID: "T1-01", Dependencies: []string{"T1-03"}})
		expect(t, got, swap("informs: [D-001], depends: []}", "informs: [D-001], depends: [T1-03]}"))
	})
	t.Run("update_deps on a block list rewrites its items in place", func(t *testing.T) {
		got := replan(t, plan, "update_deps", ReplanParams{SessionID: "T1-02", Dependencies: []string{"T1-01", "T1-03"}})
		expect(t, got, swap("    depends:\n      - T1-01\n", "    depends:\n      - T1-01\n      - T1-03\n"))
		got = replan(t, plan, "update_deps", ReplanParams{SessionID: "T1-02"})
		expect(t, got, swap("    depends:\n      - T1-01\n", "    depends: []\n"))
	})
	t.Run("remove_session drops only its entry and its section", func(t *testing.T) {
		got := replan(t, plan, "remove_session", ReplanParams{SessionID: "T1-03"})
		want := strings.Replace(swap("  - {id: T1-03, title: Auth}\n", ""), md("### T1-03: Auth\n<F>prompt\nResearch auth.\n<F>\n\n"), "", 1)
		expect(t, got, want)
	})
	t.Run("add_session goes before the synthesis entry and section", func(t *testing.T) {
		got := replan(t, plan, "add_session", ReplanParams{SessionID: "T1-04", Topic: "Cache", Layer: 1, Dependencies: []string{"T1-01"}, Prompt: "Research caches."})
		want := strings.Replace(swap("  - {id: SYN-01,", "  - {id: T1-04, title: Cache, layer: 1, door: two-way, depends: [T1-01]}\n  - {id: SYN-01,"),
			"### SYN-01: Synthesis", md("### T1-04: Cache\n\n<F>prompt\nResearch caches.\n<F>\n\n### SYN-01: Synthesis"), 1)
		expect(t, got, want)
	})
	t.Run("removing the synthesis session is refused", func(t *testing.T) {
		if _, err := ReplanInPlace([]byte(plan), "remove_session", ReplanParams{SessionID: "SYN-01"}); err == nil || !strings.Contains(err.Error(), CodePlanSynthesis) {
			t.Errorf("expected %s, got %v", CodePlanSynthesis, err)
		}
	})
	t.Run("a CRLF plan keeps CRLF line endings", func(t *testing.T) {
		crlf := strings.ReplaceAll(plan, "\n", "\r\n")
		for _, tc := range []struct {
			op string
			p  ReplanParams
		}{
			{"update_prompt", ReplanParams{SessionID: "T1-01", Prompt: "Evaluate Oracle.\nThen PostgreSQL."}},
			{"update_deps", ReplanParams{SessionID: "T1-02", Dependencies: []string{"T1-01", "T1-03"}}},
			{"update_deps", ReplanParams{SessionID: "T1-03", Dependencies: []string{"T1-01"}}},
			{"add_session", ReplanParams{SessionID: "T1-04", Topic: "Cache", Prompt: "Research caches.\nCompare two."}},
		} {
			got := replan(t, crlf, tc.op, tc.p)
			if strings.Count(got, "\n") != strings.Count(got, "\r\n") {
				t.Errorf("%s %s: the edit added LF-only lines to a CRLF plan:\n%q", tc.op, tc.p.SessionID, got)
			}
			if lf := replan(t, plan, tc.op, tc.p); strings.ReplaceAll(got, "\r\n", "\n") != lf {
				t.Errorf("%s %s: the CRLF edit differs from the LF edit", tc.op, tc.p.SessionID)
			}
		}
	})
}

func mustParse(t *testing.T, plan string) *DAG {
	t.Helper()
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}
	return dag
}

func TestReplanInPlace_DollarSignImmunity(t *testing.T) {
	origPlan := projectPlan([]string{"{id: T1-01, title: Initial Title, door: two-way, depends: []}", "{id: SYN-01, kind: synthesis, depends: [ALL]}"},
		md("### T1-01: Initial Title\n\n| **Field** | **Value** |\n|---|---|\n| **ID** | T1-01 |\n\n<F>`prompt\n# T1-01: Initial Prompt\nInitial instructions.\n<F>`\n\n")+
			section("SYN-01", "Synthesis", "Synthesize."))

	// 1. update_prompt with dollar signs ($1, $VAR, $50/mo, $0)
	dollarPrompt := "# T1-01: Cost Analysis\nCost target is $50/mo. Shell command: echo $VAR. Regex group test: $1 and $2."
	res1, err := ReplanInPlace([]byte(origPlan), "update_prompt", ReplanParams{SessionID: "T1-01", Prompt: dollarPrompt})
	if err != nil {
		t.Fatalf("update_prompt failed: %v", err)
	}
	dag, err := ParsePipeline(res1)
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}
	if got := dag.SessionByID("T1-01").Prompt; got != dollarPrompt {
		t.Errorf("expected the prompt preserved literally, got:\n%s", got)
	}

	// 2. add_session with dollar signs in the title and prompt
	res2, err := ReplanInPlace([]byte(origPlan), "add_session", ReplanParams{SessionID: "T1-02", Topic: "Cost $50/mo & $VAR", Prompt: "Price $1 vs $2.", Dependencies: []string{"T1-01"}})
	if err != nil {
		t.Fatalf("add_session failed: %v", err)
	}
	dag, err = ParsePipeline(res2)
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}
	if s := dag.SessionByID("T1-02"); s == nil || s.Title != "Cost $50/mo & $VAR" || s.Prompt != "Price $1 vs $2." {
		t.Errorf("expected title and prompt with dollar signs preserved literally, got %+v", s)
	}
}
