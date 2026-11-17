package core

import (
	"strings"
	"testing"
)

// md turns <F> into a code fence (three backticks), so plan fixtures can be raw strings.
func md(s string) string { return strings.ReplaceAll(s, "<F>", "```") }

// samplePipeline is a project plan in the plan format: one vivechak-plan block (the only source of sessions and
// dependencies) and a "### <id>: <title>" section with one prompt fence per session under "## Sessions". Its
// tables and Mermaid diagram are prose for people.
var samplePipeline = md(`# Research Pipeline: Test Project

**Complexity Score:** 8 / 24 → **Tier 1 (4–8 Sessions)**

<F>yaml vivechak-plan
plan: PROJECT
sessions:
  - {id: T1-01, title: Database Selection, layer: 0, door: one-way, informs: [D-001], depends: []}
  - {id: T1-02, title: Auth Strategy, layer: 0, door: two-way, informs: [D-002], depends: []}
  - {id: T2-01, title: Database Deep Dive, layer: 1, door: one-way, informs: [D-001], depends: [T1-01]}
  - {id: SYN-01, title: Founding Architecture Document, layer: 2, kind: synthesis, depends: [T2-01, T1-02]}
decisions:
  - {id: D-001, title: Primary datastore, door: one-way}
  - {id: D-002, title: Authentication, door: two-way}
<F>

**Execution DAG:**

<F>mermaid
graph TD
    T01[T1-01: Database Selection]
    T02[T1-02: Auth Strategy]
    T03[T2-01: Database Deep Dive]
    SYN[SYN-01: Founding Architecture Document]

    T01 -->|constrains| T03
    T03 --> SYN
    T02 --> SYN
<F>

## Sessions

### T1-01: Database Selection — Primary Datastore

| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Layer** | 0 |
| **Door Type** | One-Way |
| **Decision** | D-001 |
| **Dependencies** | None (parallel) |
| **Output File** | <F>sessions/T1-01-database-selection.md<F> |

<F>prompt
# RESEARCH BRIEF: T1-01 Database Selection

## BRIEF
Investigate database options for the project. This informs D-001 (One-Way Door).

## SCOPE
In-Scope: PostgreSQL vs MongoDB vs SQLite.

## DELIVERABLE
Concrete recommendation with evidence grades.
<F>

---

### T1-02: Auth Strategy — Authentication Approach

<F>prompt
# RESEARCH BRIEF: T1-02 Auth Strategy

## BRIEF
Evaluate authentication approaches. Two-Way Door.

## DELIVERABLE
Recommendation with tradeoff matrix.
<F>

---

### T2-01: Database Deep Dive — PostgreSQL Extensions

<F>prompt
# RESEARCH BRIEF: T2-01 Database Deep Dive

## BRIEF
Deep investigation into PostgreSQL extensions based on T1-01 findings.

[UPSTREAM_FINDINGS]

## DELIVERABLE
Extension selection with benchmarks.
<F>

---

### SYN-01: Founding Architecture Document

<F>prompt
# SYNTHESIS: Founding Architecture Document

Synthesize all research findings into the FAD.

[ALL_SESSION_FINDINGS]
<F>
`)

// projectPlan builds a minimal project plan: block entries (one per line) and the sessions' prompt sections.
func projectPlan(entries []string, sections string) string {
	return md("# Research Pipeline\n\n<F>yaml vivechak-plan\nplan: PROJECT\nsessions:\n  - " +
		strings.Join(entries, "\n  - ") + "\n<F>\n\n## Sessions\n\n" + sections)
}

// section is one "### <id>: <title>" section with its prompt fence.
func section(id, title, prompt string) string {
	return md("### " + id + ": " + title + "\n\n<F>prompt\n" + prompt + "\n<F>\n\n")
}

func planCodes(t *testing.T, data string) []string {
	t.Helper()
	var codes []string
	for _, issue := range ValidatePlan([]byte(data)).Issues {
		codes = append(codes, issue.Code)
	}
	return codes
}

func TestParsePipeline(t *testing.T) {
	dag, err := ParsePipeline([]byte(samplePipeline))
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}

	if len(dag.Sessions) != 4 {
		t.Fatalf("expected 4 sessions, got %d", len(dag.Sessions))
	}

	// Verify session IDs
	expectedIDs := []string{"T1-01", "T1-02", "T2-01", "SYN-01"}
	for i, id := range expectedIDs {
		if dag.Sessions[i].ID != id {
			t.Errorf("session %d: expected ID %q, got %q", i, id, dag.Sessions[i].ID)
		}
	}

	// Verify T2-01 has T1-01 as dependency
	t201 := dag.SessionByID("T2-01")
	if t201 == nil {
		t.Fatal("T2-01 not found")
	}
	if len(t201.Dependencies) != 1 || t201.Dependencies[0] != "T1-01" {
		t.Errorf("T2-01 dependencies: got %v, want [T1-01]", t201.Dependencies)
	}

	// Verify SYN-01 has two dependencies
	syn := dag.SessionByID("SYN-01")
	if syn == nil {
		t.Fatal("SYN-01 not found")
	}
	if len(syn.Dependencies) != 2 {
		t.Errorf("SYN-01 dependencies: got %v, want [T2-01, T1-02]", syn.Dependencies)
	}

	// Verify prompts were extracted
	t101 := dag.SessionByID("T1-01")
	if t101 == nil {
		t.Fatal("T1-01 not found")
	}
	if t101.Prompt == "" {
		t.Error("T1-01 prompt should not be empty")
	}
	if !containsStr(t101.Prompt, "Database Selection") {
		t.Error("T1-01 prompt should contain 'Database Selection'")
	}

	// Verify T2-01 prompt has the upstream slot
	if !containsStr(t201.Prompt, "[UPSTREAM_FINDINGS]") {
		t.Error("T2-01 prompt should contain [UPSTREAM_FINDINGS]")
	}
}

func TestNextSessions(t *testing.T) {
	dag, _ := ParsePipeline([]byte(samplePipeline))

	// Nothing completed → T1-01 and T1-02 should be ready (no deps)
	ready := dag.NextSessions(map[string]bool{})
	readyIDs := idsOf(ready)
	if !containsStr2(readyIDs, "T1-01") {
		t.Errorf("expected T1-01 in ready, got %v", readyIDs)
	}
	if !containsStr2(readyIDs, "T1-02") {
		t.Errorf("expected T1-02 in ready, got %v", readyIDs)
	}
	if len(ready) != 2 {
		t.Errorf("expected 2 ready sessions, got %d: %v", len(ready), readyIDs)
	}

	// T1-01 done → T2-01 should now be ready
	completed := map[string]bool{"T1-01": true}
	ready2 := dag.NextSessions(completed)
	readyIDs2 := idsOf(ready2)
	if !containsStr2(readyIDs2, "T1-02") {
		t.Errorf("T1-02 should still be ready, got %v", readyIDs2)
	}
	if !containsStr2(readyIDs2, "T2-01") {
		t.Errorf("T2-01 should now be ready, got %v", readyIDs2)
	}

	// T1-01 + T1-02 done → T2-01 ready, SYN-01 not yet (needs T2-01)
	completed2 := map[string]bool{"T1-01": true, "T1-02": true}
	ready3 := dag.NextSessions(completed2)
	readyIDs3 := idsOf(ready3)
	if !containsStr2(readyIDs3, "T2-01") {
		t.Errorf("T2-01 should be ready, got %v", readyIDs3)
	}
	if containsStr2(readyIDs3, "SYN-01") {
		t.Errorf("SYN-01 should NOT be ready yet, got %v", readyIDs3)
	}

	// All predecessors done → SYN-01 ready
	completed3 := map[string]bool{"T1-01": true, "T1-02": true, "T2-01": true}
	ready4 := dag.NextSessions(completed3)
	readyIDs4 := idsOf(ready4)
	if !containsStr2(readyIDs4, "SYN-01") {
		t.Errorf("SYN-01 should now be ready, got %v", readyIDs4)
	}

	// All complete → no more sessions
	completed4 := map[string]bool{"T1-01": true, "T1-02": true, "T2-01": true, "SYN-01": true}
	ready5 := dag.NextSessions(completed4)
	if len(ready5) != 0 {
		t.Errorf("expected 0 ready sessions, got %d", len(ready5))
	}
}

func idsOf(sessions []Session) []string {
	ids := make([]string, len(sessions))
	for i, s := range sessions {
		ids[i] = s.ID
	}
	return ids
}

func containsStr(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && contains(s, substr)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsStr2(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

func TestValidateDAG(t *testing.T) {
	// Valid DAG should pass
	dag, err := ParsePipeline([]byte(samplePipeline))
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}
	if err := dag.ValidateDAG(); err != nil {
		t.Errorf("expected valid DAG, got: %v", err)
	}

	// Dangling dependency should fail
	dangling := DAG{
		Sessions: []Session{
			{ID: "T1-01", Dependencies: []string{"NONEXISTENT"}},
		},
	}
	if err := dangling.ValidateDAG(); err == nil {
		t.Error("expected error for dangling dependency, got nil")
	}

	// Duplicate session ID should fail
	duplicate := DAG{
		Sessions: []Session{
			{ID: "T1-01"},
			{ID: "T1-01"},
		},
	}
	if err := duplicate.ValidateDAG(); err == nil {
		t.Error("expected error for duplicate session ID, got nil")
	}

	// Direct cycle (A -> B -> A) should fail
	cyclic := DAG{
		Sessions: []Session{
			{ID: "A", Dependencies: []string{"B"}},
			{ID: "B", Dependencies: []string{"A"}},
		},
	}
	if err := cyclic.ValidateDAG(); err == nil {
		t.Error("expected error for cyclic DAG, got nil")
	}
}

// The session's heading must start with its id: the flexible header forms the table parser guessed at
// ("### Session T1-01 — ...", "#### T1-01: ...") are refused with a fix that names the heading to write.
func TestPlanSessionHeadingMustStartWithID(t *testing.T) {
	plan := md(`# Pipeline

<F>yaml vivechak-plan
plan: D-001
sessions:
  - {id: D-001-S1, title: Deep Dive Decision Session, role: L, depends: []}
  - {id: D-001-S2, title: Database Architecture, role: C, depends: [D-001-S1]}
<F>

## Sessions

### D-001-S1: Deep Dive Decision Session
| Field | Value |
|---|---|
| **ID** | D-001-S1 |

<F>prompt
Prompt content
<F>

### Session D-001-S2 — Database Architecture

<F>prompt
Database prompt content
<F>
`)
	if _, err := ParsePipeline([]byte(plan)); err == nil || !strings.Contains(err.Error(), CodePlanMissingPrompt) {
		t.Fatalf("expected %s for a heading that does not start with the id, got %v", CodePlanMissingPrompt, err)
	}
	res := ValidatePlan([]byte(plan))
	if len(res.Issues) != 1 || !strings.Contains(res.Issues[0].FixHint, "### D-001-S2: <title>") {
		t.Fatalf("expected one issue whose fix names the heading ### D-001-S2: <title>, got %+v", res.Issues)
	}

	fixed := strings.Replace(plan, "### Session D-001-S2 — Database Architecture", "### D-001-S2: Database Architecture", 1)
	dag, err := ParsePipeline([]byte(fixed))
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}
	if len(dag.Sessions) != 2 || dag.Sessions[1].Prompt != "Database prompt content" || dag.Sessions[1].Role != "C" {
		t.Fatalf("unexpected sessions: %+v", dag.Sessions)
	}

	level4 := strings.Replace(fixed, "### D-001-S2:", "#### D-001-S2:", 1)
	if _, err := ParsePipeline([]byte(level4)); err == nil || !strings.Contains(err.Error(), CodePlanMissingPrompt) {
		t.Fatalf("expected %s for a level-4 session heading, got %v", CodePlanMissingPrompt, err)
	}
}

func TestParsePipeline_StrayPromptBlockDoesNotPanic(t *testing.T) {
	// A prompt code block before the ## Sessions section must not crash or become a session's prompt
	pipeline := md(`# Pipeline with intro block

Here is an example prompt template:
<F>prompt
Stray intro prompt
<F>

<F>yaml vivechak-plan
plan: PROJECT
sessions:
  - {id: T1-01, title: First Real Session, depends: []}
  - {id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}
<F>

## Sessions

### T1-01: First Real Session
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Dependencies** | None |

<F>prompt
Real session prompt
<F>

### SYN-01: Synthesis

<F>prompt
Synthesize.
<F>
`)
	dag, err := ParsePipeline([]byte(pipeline))
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}
	if len(dag.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(dag.Sessions))
	}
	if dag.Sessions[0].ID != "T1-01" {
		t.Errorf("expected ID T1-01, got %s", dag.Sessions[0].ID)
	}
	if dag.Sessions[0].Prompt != "Real session prompt" {
		t.Errorf("expected prompt 'Real session prompt', got %q", dag.Sessions[0].Prompt)
	}
}

// A decision plan written as a horizontal session table with headerless prompts used to be guessed into sessions
// (prompts handed out in table order). Without the block it is refused; with it, the block gives the sessions and
// each prompt comes from its own heading.
func TestParsePipeline_HorizontalTablePlan(t *testing.T) {
	tableOnly := md(`# Research Plan: D-015 Cache Layer

| Session ID | Role | Depends | Filename |
|---|---|---|---|
| D-015-S1 | Landscape | none | <F>sessions/D-015-S1-landscape.md<F> |
| D-015-S2 | Comparison | D-015-S1 | <F>sessions/D-015-S2-comparison.md<F> |

<F>prompt
# RESEARCH BRIEF: D-015-S1 Landscape
Map caching options.
<F>

<F>prompt
# RESEARCH BRIEF: D-015-S2 Comparison
Compare shortlisted options.
<F>
`)
	if _, err := ParsePipeline([]byte(tableOnly)); err == nil || !strings.Contains(err.Error(), CodePlanMissingBlock) {
		t.Fatalf("expected %s for a table-only plan, got %v", CodePlanMissingBlock, err)
	}

	plan := md(`# Research Plan: D-015 Cache Layer

<F>yaml vivechak-plan
plan: D-015
sessions:
  - {id: D-015-S1, title: Landscape, role: L, door: one-way, informs: [D-015], depends: []}
  - {id: D-015-S2, title: Comparison, role: C, door: one-way, informs: [D-015], depends: [D-015-S1]}
<F>

| Session ID | Role | Depends | Filename |
|---|---|---|---|
| D-015-S1 | Landscape | none | <F>sessions/D-015-S1-landscape.md<F> |
| D-015-S2 | Comparison | D-015-S1 | <F>sessions/D-015-S2-comparison.md<F> |

## Sessions

### D-015-S1: Landscape

<F>prompt
# RESEARCH BRIEF: D-015-S1 Landscape
Map caching options.
<F>

### D-015-S2: Comparison

<F>prompt
# RESEARCH BRIEF: D-015-S2 Comparison
Compare shortlisted options.
<F>
`)
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePipeline horizontal table: %v", err)
	}
	if len(dag.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(dag.Sessions))
	}
	if dag.Sessions[0].ID != "D-015-S1" {
		t.Errorf("expected D-015-S1, got %s", dag.Sessions[0].ID)
	}
	if dag.Sessions[0].OutputFile != "research/sessions/D-015-S1.md" {
		t.Errorf("expected output file research/sessions/D-015-S1.md, got %s", dag.Sessions[0].OutputFile)
	}
	if dag.Sessions[1].ID != "D-015-S2" {
		t.Errorf("expected D-015-S2, got %s", dag.Sessions[1].ID)
	}
	if len(dag.Sessions[1].Dependencies) != 1 || dag.Sessions[1].Dependencies[0] != "D-015-S1" {
		t.Errorf("expected dependency D-015-S1, got %v", dag.Sessions[1].Dependencies)
	}
	if !strings.Contains(dag.Sessions[0].Prompt, "Map caching options.") {
		t.Errorf("expected D-015-S1 prompt to contain 'Map caching options.', got %q", dag.Sessions[0].Prompt)
	}
	if !strings.Contains(dag.Sessions[1].Prompt, "Compare shortlisted options.") {
		t.Errorf("expected D-015-S2 prompt to contain 'Compare shortlisted options.', got %q", dag.Sessions[1].Prompt)
	}
}

func TestParsePipeline_OverviewTableAndHeaders(t *testing.T) {
	plan := md(`# Research Plan: D-016 Messaging Architecture

<F>yaml vivechak-plan
plan: D-016
sessions:
  - {id: D-016-S1, title: Messaging Landscape, role: L, depends: []}
  - {id: D-016-S2, role: C, depends: [D-016-S1]}
<F>

## Overview
| Session ID | Role | Depends | Filename |
|---|---|---|---|
| D-016-S1 | Landscape | none | <F>sessions/D-016-S1-landscape.md<F> |
| D-016-S2 | Comparison | D-016-S1 | <F>sessions/D-016-S2-comparison.md<F> |

## Sessions

### D-016-S1: Messaging Landscape
<F>prompt
# RESEARCH BRIEF: D-016-S1 Messaging Landscape
Investigate pub/sub brokers.
<F>

### D-016-S2: NATS vs Kafka Comparison
<F>prompt
# RESEARCH BRIEF: D-016-S2 NATS vs Kafka Comparison
Compare latency and durability.
<F>
`)
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePipeline error: %v", err)
	}
	if len(dag.Sessions) != 2 {
		t.Fatalf("expected exactly 2 sessions, got %d", len(dag.Sessions))
	}
	s1 := dag.Sessions[0]
	if s1.ID != "D-016-S1" || s1.Title != "Messaging Landscape" || s1.OutputFile != "research/sessions/D-016-S1.md" {
		t.Errorf("unexpected s1 fields: %+v", s1)
	}
	if !strings.Contains(s1.Prompt, "Investigate pub/sub brokers.") {
		t.Errorf("unexpected s1 prompt: %q", s1.Prompt)
	}
	s2 := dag.Sessions[1]
	// A session without a title in the block takes its heading's title.
	if s2.ID != "D-016-S2" || s2.Title != "NATS vs Kafka Comparison" {
		t.Errorf("unexpected s2 fields: %+v", s2)
	}
	if len(s2.Dependencies) != 1 || s2.Dependencies[0] != "D-016-S1" {
		t.Errorf("unexpected s2 dependencies: %v", s2.Dependencies)
	}
	if !strings.Contains(s2.Prompt, "Compare latency and durability.") {
		t.Errorf("unexpected s2 prompt: %q", s2.Prompt)
	}
}

func TestParsePipeline_NestedCodeBlocks(t *testing.T) {
	plan := projectPlan([]string{"{id: T1-01, title: Session With Code Fence, depends: []}", "{id: SYN-01, kind: synthesis, depends: [ALL]}"},
		md(`### T1-01: Session With Code Fence
| Field | Value |
|---|---|
| **ID** | T1-01 |

<F>prompt
# RESEARCH BRIEF: T1-01
Output format required:
<F>yaml
id: T1-01
status: complete
<F>
Conclusion: must not be truncated.
<F>

`)+section("SYN-01", "Synthesis", "Synthesize."))
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePipeline error: %v", err)
	}
	if len(dag.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(dag.Sessions))
	}
	prompt := dag.Sessions[0].Prompt
	if !strings.Contains(prompt, "Conclusion: must not be truncated.") {
		t.Errorf("prompt was truncated by inner code fence: %q", prompt)
	}
	if !strings.Contains(prompt, "status: complete") {
		t.Errorf("prompt missing inner code block: %q", prompt)
	}
}

func TestValidateDAG_AlphanumericIDs(t *testing.T) {
	// Valid alphanumeric IDs without hyphens
	validDAG := DAG{
		Sessions: []Session{
			{ID: "T1", Dependencies: nil},
			{ID: "SYN", Dependencies: []string{"T1"}},
		},
	}
	if err := validDAG.ValidateDAG(); err != nil {
		t.Errorf("expected valid DAG with alphanumeric IDs, got: %v", err)
	}

	// Invalid ID with slash
	invalidDAG := DAG{
		Sessions: []Session{
			{ID: "invalid/id", Dependencies: nil},
		},
	}
	if err := invalidDAG.ValidateDAG(); err == nil {
		t.Error("expected error for ID with slash, got nil")
	}
}

func TestParsePipeline_SingleTokenAndHeaders(t *testing.T) {
	content := md(`# Research Pipeline

<F>yaml vivechak-plan
plan: PROJECT
sessions:
  - {id: S1, title: Landscape, depends: []}
  - {id: S2, depends: [S1]}
  - {id: SYN, kind: synthesis, depends: [S1, S2]}
<F>

## Sessions

### S1: Landscape
| **ID** | S1 |
| **Dependencies** | None |

<F>` + "`prompt" + `
Execute landscape research
<F>` + "`" + `

### S2: Deep Dive

<F>` + "`prompt" + `
Execute deep dive research
<F>` + "`" + `

### SYN: Synthesis

<F>` + "`prompt" + `
Synthesize all findings
<F>` + "`" + `
`)

	dag, err := ParsePipeline([]byte(content))
	if err != nil {
		t.Fatalf("ParsePipeline error: %v", err)
	}
	if len(dag.Sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(dag.Sessions))
	}
	if dag.Sessions[0].ID != "S1" || dag.Sessions[0].Title != "Landscape" || dag.Sessions[0].Prompt != "Execute landscape research" {
		t.Errorf("unexpected session 0: %+v", dag.Sessions[0])
	}
	if dag.Sessions[1].ID != "S2" || dag.Sessions[1].Title != "Deep Dive" || len(dag.Sessions[1].Dependencies) != 1 || dag.Sessions[1].Dependencies[0] != "S1" {
		t.Errorf("unexpected session 1: %+v", dag.Sessions[1])
	}
	if dag.Sessions[2].ID != "SYN" || len(dag.Sessions[2].Dependencies) != 2 || dag.Sessions[2].Dependencies[0] != "S1" || dag.Sessions[2].Dependencies[1] != "S2" {
		t.Errorf("unexpected session 2: %+v", dag.Sessions[2])
	}

	if err := dag.ValidateDAG(); err != nil {
		t.Errorf("expected valid DAG, got: %v", err)
	}
}

func TestParsePipeline_PromptHeaderWithLeadingLines(t *testing.T) {
	content := projectPlan([]string{"{id: T1-01, title: Database Selection, depends: []}", "{id: T1-02, title: Auth Strategy, depends: []}", "{id: SYN-01, kind: synthesis, depends: [ALL]}"},
		section("T1-01", "Database Selection", "<!-- Generated by Vivechak -->\n<!-- Priority: High -->\n\n# Session T1-02: Primary Datastore\nInvestigate database options.")+
			section("T1-02", "Auth Strategy", "<!-- Note: Two-way door -->\n\n# Session T1-01: Auth Strategy\nInvestigate auth options.")+
			section("SYN-01", "Synthesis", "Synthesize."))

	dag, err := ParsePipeline([]byte(content))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	if len(dag.Sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(dag.Sessions))
	}
	// A prompt belongs to the heading it sits under, whatever ids its own text mentions.
	if !strings.Contains(dag.Sessions[0].Prompt, "Primary Datastore") {
		t.Errorf("expected T1-01 prompt to contain 'Primary Datastore', got: %q", dag.Sessions[0].Prompt)
	}
	if !strings.Contains(dag.Sessions[1].Prompt, "Investigate auth options.") {
		t.Errorf("expected T1-02 prompt to contain 'Investigate auth options.', got: %q", dag.Sessions[1].Prompt)
	}
}

func TestParsePipeline_UnclosedPromptAtEOFWithLeadingLines(t *testing.T) {
	// A plan whose final prompt is unclosed at EOF (no closing fence) and starts with comment lines.
	content := projectPlan([]string{"{id: T1-01, title: Database Selection, depends: []}", "{id: T1-02, title: Auth Strategy, depends: []}"},
		section("T1-01", "Database Selection", "# Session T1-01: Primary Datastore\nInvestigate database options.")+
			md("### T1-02: Auth Strategy\n\n<F>prompt\n<!-- Note: Two-way door -->\n<!-- Generated by Vivechak -->\n\n# Session T1-02: Auth Strategy\nInvestigate auth options at EOF."))

	dag, err := ParsePipeline([]byte(content))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	if len(dag.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(dag.Sessions))
	}
	if !strings.Contains(dag.Sessions[1].Prompt, "Investigate auth options at EOF.") {
		t.Errorf("expected T1-02 prompt to contain EOF prompt body, got: %q", dag.Sessions[1].Prompt)
	}
	if dag.Sessions[1].ID != "T1-02" {
		t.Errorf("expected session ID T1-02, got: %s", dag.Sessions[1].ID)
	}
}

// Synthesis is identified only by kind: synthesis (HIGH-07): FINAL-01 with the kind is the synthesis session
// (held back until every other session completes, output research/FAD.md); SYNC-01, SYN-01 and FAD without it are
// ordinary sessions.
func TestPlanSynthesisIsKindOnly(t *testing.T) {
	plan := projectPlan([]string{
		"{id: SYNC-01, title: Sync protocol, depends: []}",
		"{id: SYN-01, title: Syndication feeds, depends: []}",
		"{id: FAD, title: Feature audit, depends: [SYNC-01]}",
		"{id: FINAL-01, title: Founding architecture, kind: synthesis, depends: [SYNC-01]}",
	}, section("SYNC-01", "Sync protocol", "Research sync.")+section("SYN-01", "Syndication feeds", "Research feeds.")+
		section("FAD", "Feature audit", "Audit features.")+section("FINAL-01", "Founding architecture", "Synthesize. [ALL_SESSION_FINDINGS]"))

	if codes := planCodes(t, plan); len(codes) != 0 {
		t.Fatalf("expected a valid plan, got %v", codes)
	}
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}
	for _, id := range []string{"SYNC-01", "SYN-01", "FAD"} {
		s := dag.SessionByID(id)
		if s.IsSynthesis() || dag.IsSynthesisID(id) || s.OutputFile != "research/sessions/"+id+".md" {
			t.Errorf("%s without kind: synthesis must be an ordinary session: %+v", id, s)
		}
	}
	final := dag.SessionByID("FINAL-01")
	if !final.IsSynthesis() || !dag.IsSynthesisID("final-01") || final.OutputFile != FADFile {
		t.Fatalf("FINAL-01 with kind: synthesis must be the synthesis session: %+v", final)
	}
	if got := dag.SynthesisSessions(); len(got) != 1 || got[0].ID != "FINAL-01" {
		t.Fatalf("SynthesisSessions = %v, want [FINAL-01]", got)
	}

	// FINAL-01 depends only on SYNC-01 but waits for every other session; SYN-01 and FAD are served like any other.
	ready := idsOf(dag.NextSessions(map[string]bool{"SYNC-01": true}))
	if containsStr2(ready, "FINAL-01") || !containsStr2(ready, "SYN-01") || !containsStr2(ready, "FAD") {
		t.Fatalf("after SYNC-01: ready %v, want SYN-01 and FAD but not FINAL-01", ready)
	}
	ready = idsOf(dag.NextSessions(map[string]bool{"SYNC-01": true, "SYN-01": true, "FAD": true}))
	if len(ready) != 1 || ready[0] != "FINAL-01" {
		t.Fatalf("after every research session: ready %v, want [FINAL-01]", ready)
	}

	// A project plan whose only synthesis-looking session lacks the kind is refused.
	noKind := strings.Replace(plan, ", kind: synthesis", "", 1)
	if codes := planCodes(t, noKind); len(codes) != 1 || codes[0] != CodePlanSynthesis {
		t.Fatalf("expected [%s] without any kind: synthesis, got %v", CodePlanSynthesis, codes)
	}
}

// depends is a YAML list of the block's session ids; ALL (in any of its spellings) is every non-synthesis session
// listed before the one that says it.
func TestPlanDependsListsAndALL(t *testing.T) {
	for _, all := range []string{"ALL", "all", "ALL prior sessions (hard)", "all previous sessions", "all preceding sessions",
		"all other sessions", "all non-synthesis sessions", "all sessions", "__all_prior__", "__all__", "'*'"} {
		plan := projectPlan([]string{
			"{id: T0-01, depends: []}",
			"{id: T0-02}",
			"{id: T1-01, depends: [" + all + "]}",
			"{id: SYN-01, kind: synthesis, depends: [" + all + "]}",
		}, section("T0-01", "A", "a")+section("T0-02", "B", "b")+section("T1-01", "C", "c")+section("SYN-01", "S", "s"))
		dag, err := ParsePipeline([]byte(plan))
		if err != nil {
			t.Fatalf("depends: [%s]: %v", all, err)
		}
		if got := dag.SessionByID("T1-01").Dependencies; strings.Join(got, ",") != "T0-01,T0-02" {
			t.Errorf("depends: [%s] on T1-01 = %v, want [T0-01 T0-02]", all, got)
		}
		if got := dag.SessionByID("SYN-01").Dependencies; strings.Join(got, ",") != "T0-01,T0-02,T1-01" {
			t.Errorf("depends: [%s] on SYN-01 = %v, want [T0-01 T0-02 T1-01]", all, got)
		}
		if !isAllPriorSessions(strings.Trim(all, "'")) {
			t.Errorf("isAllPriorSessions(%q) = false", all)
		}
	}

	// ALL never pulls in the synthesis session, wherever the planner lists it.
	plan := projectPlan([]string{"{id: SYN-01, kind: synthesis, depends: []}", "{id: T0-01}", "{id: T1-01, depends: [ALL]}"},
		section("SYN-01", "S", "s")+section("T0-01", "A", "a")+section("T1-01", "B", "b"))
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}
	if got := dag.SessionByID("T1-01").Dependencies; strings.Join(got, ",") != "T0-01" {
		t.Errorf("ALL after the synthesis session = %v, want [T0-01]", got)
	}

	if isAllPriorSessions("T1-01") || isAllPriorSessions("allocation") {
		t.Error("isAllPriorSessions must not match ids or words that merely start with 'all'")
	}
}

func TestDAG_TransitiveDependents(t *testing.T) {
	// Linear DAG: T0-01 -> T1-01 -> SYN-01
	dag := &DAG{
		Sessions: []Session{
			{ID: "T0-01", Dependencies: nil},
			{ID: "T1-01", Dependencies: []string{"T0-01"}},
			{ID: "SYN-01", Dependencies: []string{"T1-01"}},
			{ID: "T0-02", Dependencies: nil}, // Independent
		},
	}

	// Dependents of T0-01 should be T1-01 and SYN-01
	deps01 := dag.TransitiveDependents("T0-01")
	expected01 := []string{"SYN-01", "T1-01"}
	if len(deps01) != len(expected01) || deps01[0] != expected01[0] || deps01[1] != expected01[1] {
		t.Errorf("TransitiveDependents(T0-01) = %v; want %v", deps01, expected01)
	}

	// Dependents of T1-01 should be only SYN-01
	deps11 := dag.TransitiveDependents("T1-01")
	expected11 := []string{"SYN-01"}
	if len(deps11) != len(expected11) || deps11[0] != expected11[0] {
		t.Errorf("TransitiveDependents(T1-01) = %v; want %v", deps11, expected11)
	}

	// Dependents of leaf SYN-01 should be empty
	depsSyn := dag.TransitiveDependents("SYN-01")
	if len(depsSyn) != 0 {
		t.Errorf("TransitiveDependents(SYN-01) = %v; want empty", depsSyn)
	}

	// Dependents of independent session T0-02 should be empty
	deps02 := dag.TransitiveDependents("T0-02")
	if len(deps02) != 0 {
		t.Errorf("TransitiveDependents(T0-02) = %v; want empty", deps02)
	}

	// Nil DAG should return nil
	var nilDAG *DAG
	if nilDAG.TransitiveDependents("T0-01") != nil {
		t.Errorf("nil DAG should return nil")
	}

	// Diamond DAG: A -> B, C -> D
	diamond := &DAG{
		Sessions: []Session{
			{ID: "A"},
			{ID: "B", Dependencies: []string{"A"}},
			{ID: "C", Dependencies: []string{"A"}},
			{ID: "D", Dependencies: []string{"B", "C"}},
		},
	}
	depsA := diamond.TransitiveDependents("A")
	expectedA := []string{"B", "C", "D"}
	if len(depsA) != len(expectedA) {
		t.Fatalf("TransitiveDependents(A) len = %d; want %d (%v)", len(depsA), len(expectedA), depsA)
	}
	for i := range depsA {
		if depsA[i] != expectedA[i] {
			t.Errorf("TransitiveDependents(A)[%d] = %q; want %q", i, depsA[i], expectedA[i])
		}
	}
}

func TestDAG_CaseInsensitiveOperations(t *testing.T) {
	dag := &DAG{
		Sessions: []Session{
			{ID: "T0-01", Dependencies: nil},
			{ID: "t0-02", Dependencies: []string{"T0-01"}},
			{ID: "T1-01", Dependencies: []string{"t0-01", "T0-02"}},
		},
	}

	// 1. SessionByID with various casings
	if s := dag.SessionByID("t0-01"); s == nil || s.ID != "T0-01" {
		t.Errorf("expected to find T0-01 with lowercase query, got: %v", s)
	}
	if s := dag.SessionByID("T0-02"); s == nil || s.ID != "t0-02" {
		t.Errorf("expected to find t0-02 with uppercase query, got: %v", s)
	}

	// 2. ValidateDAG with mixed casing dependencies
	if err := dag.ValidateDAG(); err != nil {
		t.Errorf("expected ValidateDAG to succeed with mixed-case dependencies, got: %v", err)
	}

	// 3. NextSessions with mixed-case completed IDs
	completed := map[string]bool{
		"t0-01": true, // lowercase completion of T0-01
	}
	ready := dag.NextSessions(completed)
	if len(ready) != 1 || !strings.EqualFold(ready[0].ID, "t0-02") {
		t.Errorf("expected t0-02 to be ready when t0-01 completed, got: %+v", ready)
	}

	// Complete t0-02 with uppercase ID
	completed["T0-02"] = true
	ready = dag.NextSessions(completed)
	if len(ready) != 1 || !strings.EqualFold(ready[0].ID, "t1-01") {
		t.Errorf("expected T1-01 to be ready when T0-02 completed, got: %+v", ready)
	}
}

func TestDAG_ParsePipeline_UntaggedInnerCodeFence(t *testing.T) {
	input := projectPlan([]string{"{id: S1, title: Test, depends: []}", "{id: SYN-01, kind: synthesis, depends: [ALL]}"},
		md(`### S1: Test

| **Field** | **Value** |
|---|---|
| **ID** | S1 |

<F>prompt
# Prompt Brief

Here is an example code block:
<F>

foo := 123
bar := 456
<F>

This instruction must not be truncated!
<F>

---

`)+section("SYN-01", "Synthesis", "Synthesize."))

	dag, err := ParsePipeline([]byte(input))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	s := dag.SessionByID("S1")
	if s == nil {
		t.Fatalf("session S1 not found")
	}
	if !strings.Contains(s.Prompt, "This instruction must not be truncated!") {
		t.Errorf("inner untagged code fence caused prompt truncation. Prompt was:\n%s", s.Prompt)
	}
}

// Session ids are only names: FAD and SYNTHESIS are ordinary sessions unless the block marks one kind: synthesis.
func TestParsePipeline_StandaloneIdentifiers(t *testing.T) {
	input := md(`# Pipeline

<F>yaml vivechak-plan
plan: D-001
sessions:
  - {id: FAD, depends: []}
  - {id: SYNTHESIS, depends: [FAD]}
<F>

## Sessions

### FAD: Founding Architecture Document
<F>prompt
Write it.
<F>

### SYNTHESIS: Final System Synthesis
<F>prompt
Synthesize it.
<F>
`)
	dag, err := ParsePipeline([]byte(input))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	if len(dag.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(dag.Sessions))
	}
	if dag.Sessions[0].ID != "FAD" || dag.Sessions[0].Title != "Founding Architecture Document" || dag.Sessions[0].IsSynthesis() {
		t.Errorf("unexpected session 0: %+v", dag.Sessions[0])
	}
	if dag.Sessions[1].ID != "SYNTHESIS" || dag.Sessions[1].Title != "Final System Synthesis" || dag.Sessions[1].IsSynthesis() {
		t.Errorf("unexpected session 1: %+v", dag.Sessions[1])
	}
}

// A plan the generator cut short (sessions listed, prompt sections missing) used to be caught by a regex scan of
// tables and diagrams (W-PARSER-SESSION-MISMATCH). The block now lists the sessions, so each missing prompt is a
// blocking V-PLAN-MISSING-PROMPT, and Mermaid diagrams or tables never yield sessions or warnings.
func TestValidatePlanRefusesTruncatedPlan(t *testing.T) {
	head := md(`# Research Pipeline

<F>yaml vivechak-plan
plan: PROJECT
sessions:
  - {id: T0-01, title: Storage Landscape, layer: 0, depends: []}
  - {id: T0-02, title: Auth Landscape, layer: 0, depends: []}
  - {id: T1-01, title: SQLite Engine, layer: 1, depends: [T0-01, T0-02]}
  - {id: SYN-01, title: Architecture Synthesis, layer: 2, kind: synthesis, depends: [ALL]}
<F>

## Pipeline Topology

| Session | Layer | Focus | Dependencies | Est. Calls |
|---|---|---|---|---|
| T0-01 | 0 | Storage Landscape | None | 3 |
| T0-02 | 0 | Auth Landscape | None | 3 |
| T1-01 | 1 | SQLite Engine | T0-01 | 5 |
| SYN-01 | 2 | Architecture Synthesis | T1-01 | 4 |

## Execution DAG

<F>mermaid
graph TD
    T0_01[T0-01 Storage] --> T1_01[T1-01 SQLite]
    T0_02 --> T1_01
    T1_01 --> SYN_01[SYN-01 Grand Synthesis]
    SYN_01 --> FAD
<F>

## Sessions

`)
	truncated := head + section("T0-01", "Storage Landscape", "# T0-01 Brief")
	res := ValidatePlan([]byte(truncated))
	var missing []string
	for _, issue := range res.Issues {
		if issue.Code != CodePlanMissingPrompt || issue.Level != L2Block {
			t.Errorf("unexpected issue %+v", issue)
		}
		missing = append(missing, issue.Message)
	}
	if len(missing) != 3 || !strings.Contains(strings.Join(missing, "|"), "T0-02") || !strings.Contains(strings.Join(missing, "|"), "SYN-01") {
		t.Fatalf("expected a blocking %s for T0-02, T1-01 and SYN-01, got %v", CodePlanMissingPrompt, missing)
	}

	complete := truncated + section("T0-02", "Auth Landscape", "b") + section("T1-01", "SQLite Engine", "c") + section("SYN-01", "Architecture Synthesis", "s")
	if res := ValidatePlan([]byte(complete)); len(res.Issues) != 0 {
		t.Fatalf("expected no issues once every session has its prompt, got %+v", res.Issues)
	}
}

func TestNextSessions_TerminalSynthesisGuard(t *testing.T) {
	dag := &DAG{
		Sessions: []Session{
			{ID: "T0-01", Title: "Discovery", Dependencies: nil},
			{ID: "T1-01", Title: "Database Choice", Dependencies: []string{"T0-01"}},
			{ID: "SYN-01", Title: "System Synthesis", Kind: SynthesisKind, Dependencies: []string{"T0-01"}}, // Note: artificially weak dependency on T0-01
		},
	}

	// Case 1: Nothing completed -> only T0-01 ready
	ready1 := dag.NextSessions(map[string]bool{})
	if len(ready1) != 1 || ready1[0].ID != "T0-01" {
		t.Fatalf("expected only T0-01 ready, got %+v", ready1)
	}

	// Case 2: T0-01 completed -> T1-01 ready, but SYN-01 MUST be held back
	// because T1-01 (a non-synthesis research session) is not completed!
	ready2 := dag.NextSessions(map[string]bool{"T0-01": true})
	for _, s := range ready2 {
		if s.ID == "SYN-01" {
			t.Fatalf("SYN-01 should NOT be actionable while T1-01 remains uncompleted!")
		}
	}
	foundT1 := false
	for _, s := range ready2 {
		if s.ID == "T1-01" {
			foundT1 = true
		}
	}
	if !foundT1 {
		t.Fatalf("expected T1-01 to be actionable once T0-01 is complete")
	}

	// Case 3: Both T0-01 and T1-01 completed -> SYN-01 is now actionable
	ready3 := dag.NextSessions(map[string]bool{"T0-01": true, "T1-01": true})
	foundSYN := false
	for _, s := range ready3 {
		if s.ID == "SYN-01" {
			foundSYN = true
		}
	}
	if !foundSYN {
		t.Fatalf("expected SYN-01 to be actionable once all non-synthesis sessions are completed")
	}
}

func TestParsePipeline_NoPhantomSessionFromPlanTitle(t *testing.T) {
	planContent := md(`# D-001-PLAN.md — Decision Research Plan

<F>yaml vivechak-plan
plan: D-001
sessions:
  - {id: D-001-S1, title: Landscape, role: L, depends: []}
  - {id: D-001-S2, title: Comparison, role: C, depends: [D-001-S1]}
<F>

## Routing
| Profile | Value | Rationale |
|---|---|---|
| R | 3 | High reversal cost |

## Sessions

### D-001-S1: Landscape — Architecture Options
| Field | Value |
|---|---|
| ID | D-001-S1 |
| Role | L |
| Depends | none |

<F>prompt
Brief for S1
<F>

### D-001-S2: Comparison — Microservices vs Monolith
| Field | Value |
|---|---|
| ID | D-001-S2 |
| Role | C |
| Depends | D-001-S1 (constrains) |

<F>prompt
Brief for S2
<F>
`)

	dag, err := ParsePipeline([]byte(planContent))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	// Must contain exactly 2 sessions: D-001-S1 and D-001-S2 (no phantom D-001 or D-001-PLAN)
	if len(dag.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d: %v", len(dag.Sessions), idsOf(dag.Sessions))
	}
	s2 := dag.SessionByID("D-001-S2")
	if s2 == nil || dag.SessionByID("D-001-S1") == nil {
		t.Fatalf("expected D-001-S1 and D-001-S2, got %v", idsOf(dag.Sessions))
	}
	if len(s2.Dependencies) != 1 || s2.Dependencies[0] != "D-001-S1" {
		t.Fatalf("expected D-001-S2 dependency to be ['D-001-S1'], got: %v", s2.Dependencies)
	}
	if res := ValidatePlan([]byte(planContent)); len(res.Issues) != 0 {
		t.Fatalf("a # D-001-PLAN heading must not warn: %+v", res.Issues)
	}
}

func TestParsePipeline_AllPriorSessionsExpansion(t *testing.T) {
	pipelineMD := projectPlan([]string{"{id: T0-01, title: Session 1, depends: []}", "{id: T0-02, title: Session 2, depends: []}",
		"{id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL prior sessions (hard)]}"},
		section("T0-01", "Session 1", "Prompt 1")+"---\n\n"+section("T0-02", "Session 2", "Prompt 2")+"---\n\n"+section("SYN-01", "Synthesis", "Synthesis prompt"))

	dag, err := ParsePipeline([]byte(pipelineMD))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	syn := dag.SessionByID("SYN-01")
	if syn == nil {
		t.Fatal("SYN-01 not found in parsed pipeline")
	}

	expectedDeps := []string{"T0-01", "T0-02"}
	if len(syn.Dependencies) != len(expectedDeps) {
		t.Fatalf("SYN-01 dependencies = %v; want %v", syn.Dependencies, expectedDeps)
	}
	for i, dep := range syn.Dependencies {
		if dep != expectedDeps[i] {
			t.Errorf("dep[%d] = %q; want %q", i, dep, expectedDeps[i])
		}
	}

	if err := dag.ValidateDAG(); err != nil {
		t.Errorf("ValidateDAG() error = %v; want nil", err)
	}
}

// Tables are prose: bolded, backticked, four-column or arbitrary tables never become sessions, dependencies or
// titles, and a plan with only tables is refused with V-PLAN-MISSING-BLOCK.
func TestParsePipeline_IgnoreArbitraryFourColumnTables(t *testing.T) {
	tables := md(`| Step | Action | Description | Notes |
|---|---|---|---|
| 1 | Scaffolding | Initialize repo | First priority |
| 2 | Dependencies | Install libs | Use npm |

| **ID** | **Topic** | **Dependencies** | **OutputFile** |
|---|---|---|---|
| **T9-01** | **Phantom Baseline** | None | <F>sessions/T9-01.md<F> |
| <F>T9-02<F> | *Phantom Engine* | <F>T9-01<F> | "sessions/T9-02.md" |

| **Field** | **Value** |
|---|---|
| **Session ID** | <F>T9-03<F> |
| **Dependencies** | <F>T9-02<F> |
`)
	if _, err := ParsePipeline([]byte("# Project Architecture\n\n" + tables)); err == nil || !strings.Contains(err.Error(), CodePlanMissingBlock) {
		t.Fatalf("expected %s for a plan made of tables, got %v", CodePlanMissingBlock, err)
	}

	plan := projectPlan([]string{"{id: T0-01, title: System Architecture, depends: []}", "{id: SYN-01, kind: synthesis, depends: [ALL]}"},
		tables+"\n"+section("T0-01", "System Architecture", "Investigate architecture.")+section("SYN-01", "Synthesis", "Synthesize."))
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	if len(dag.Sessions) != 2 || dag.SessionByID("T0-01") == nil {
		t.Fatalf("expected exactly T0-01 and SYN-01, got %v", idsOf(dag.Sessions))
	}
	for _, phantom := range []string{"1", "Option A", "T9-01", "T9-02", "T9-03"} {
		if dag.SessionByID(phantom) != nil {
			t.Errorf("table row %q was parsed as a session", phantom)
		}
	}
	if res := ValidatePlan([]byte(plan)); len(res.Issues) != 0 {
		t.Fatalf("tables must not produce issues: %+v", res.Issues)
	}
}

func TestDAG_ParsePipeline_MachineBlockAllPriorExpansion(t *testing.T) {
	plan := md("<F>yaml vivechak-plan\n" +
		"plan: PROJECT\n" +
		"sessions:\n" +
		"  - id: T0-01\n" +
		"    title: Core Framework\n" +
		"    layer: 0\n" +
		"    door: two-way\n" +
		"    depends: []\n" +
		"  - id: T1-01\n" +
		"    title: Datastore Selection\n" +
		"    layer: 1\n" +
		"    door: one-way\n" +
		"    depends: [\"T0-01\"]\n" +
		"  - id: SYN-01\n" +
		"    title: Architecture Synthesis\n" +
		"    layer: 2\n" +
		"    kind: synthesis\n" +
		"    depends: [\"all prior\"]\n" +
		"<F>\n\n# Research Pipeline\n\n## Sessions\n\n" +
		section("T0-01", "Core Framework", "a") + section("T1-01", "Datastore Selection", "b") + section("SYN-01", "Architecture Synthesis", "s"))
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	syn := dag.SessionByID("SYN-01")
	if syn == nil {
		t.Fatalf("SYN-01 not found")
	}
	if len(syn.Dependencies) != 2 {
		t.Fatalf("expected 2 expanded dependencies for SYN-01, got %v", syn.Dependencies)
	}
	if syn.Dependencies[0] != "T0-01" || syn.Dependencies[1] != "T1-01" {
		t.Fatalf("expected [T0-01, T1-01], got %v", syn.Dependencies)
	}
}
