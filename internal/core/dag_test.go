package core

import (
	"strings"
	"testing"
)

const samplePipeline = `# Research Pipeline: Test Project

**Complexity Score:** 8 / 24 → **Tier 1 (4–8 Sessions)**

**Execution DAG:**

` + "```mermaid" + `
graph TD
    T01[T1-01: Database Selection]
    T02[T1-02: Auth Strategy]
    T03[T2-01: Database Deep Dive]
    SYN[SYN-01: Founding Architecture Document]

    T01 -->|constrains| T03
    T03 --> SYN
    T02 --> SYN
` + "```" + `

## Sessions

#### T1-01: Database Selection — Primary Datastore

| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Layer** | 0 |
| **Door Type** | One-Way |
| **Decision** | D-001 |
| **Dependencies** | None (parallel) |
| **Output File** | ` + "`sessions/T1-01-database-selection.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-01 Database Selection

## BRIEF
Investigate database options for the project. This informs D-001 (One-Way Door).

## SCOPE
In-Scope: PostgreSQL vs MongoDB vs SQLite.

## DELIVERABLE
Concrete recommendation with evidence grades.
` + "```" + `

---

#### T1-02: Auth Strategy — Authentication Approach

| Field | Value |
|---|---|
| **ID** | T1-02 |
| **Layer** | 0 |
| **Door Type** | Two-Way |
| **Decision** | D-002 |
| **Dependencies** | None |
| **Output File** | ` + "`sessions/T1-02-auth-strategy.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T1-02 Auth Strategy

## BRIEF
Evaluate authentication approaches. Two-Way Door.

## DELIVERABLE
Recommendation with tradeoff matrix.
` + "```" + `

---

#### T2-01: Database Deep Dive — PostgreSQL Extensions

| Field | Value |
|---|---|
| **ID** | T2-01 |
| **Layer** | 1 |
| **Door Type** | One-Way |
| **Decision** | D-001 |
| **Dependencies** | T1-01 |
| **Output File** | ` + "`sessions/T2-01-database-deep-dive.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: T2-01 Database Deep Dive

## BRIEF
Deep investigation into PostgreSQL extensions based on T1-01 findings.

[UPSTREAM_FINDINGS]

## DELIVERABLE
Extension selection with benchmarks.
` + "```" + `

---

#### SYN-01: Founding Architecture Document

| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Layer** | 2 |
| **Door Type** | — |
| **Decision** | — |
| **Dependencies** | T2-01, T1-02 |
| **Output File** | ` + "`research/FAD.md`" + ` |

` + "```prompt" + `
# SYNTHESIS: Founding Architecture Document

Synthesize all research findings into the FAD.

[ALL_SESSION_FINDINGS]
` + "```" + `
`

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

func TestParseDependencies(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"None (parallel)", nil},
		{"none (independent)", nil},
		{"none()", nil},
		{"None", nil},
		{"none", nil},
		{"—", nil},
		{"-", nil},
		{"n/a", nil},
		{"N/A", nil},
		{"", nil},
		{"   ", nil},
		{"T1-01", []string{"T1-01"}},
		{"T1-01, T1-02", []string{"T1-01", "T1-02"}},
		{"T2-01, T1-02", []string{"T2-01", "T1-02"}},
		{"T1-01 (soft), T1-02", []string{"T1-01", "T1-02"}},
		{"T1-01, T1-02 (none-blocking)", []string{"T1-01", "T1-02"}},
		{"T1-01, none", []string{"T1-01"}},
		{"[T1-01, T1-02]", []string{"T1-01", "T1-02"}},
		{"`T1-01` and `T1-02`", []string{"T1-01", "T1-02"}},
		{"T1-01 & T1-02", []string{"T1-01", "T1-02"}},
		{"[None]", nil},
		{"`none`", nil},
		{"[R-01](./R-01.md), [R-02](#r-02)", []string{"R-01", "R-02"}},
		{"- R-01\n- R-02", []string{"R-01", "R-02"}},
		{"* R-01\r\n* R-02", []string{"R-01", "R-02"}},
	}


	for _, tc := range tests {
		got := parseDependencies(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("parseDependencies(%q): got %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("parseDependencies(%q)[%d]: got %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
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

func TestComplexityScore(t *testing.T) {
	dag, err := ParsePipeline([]byte(samplePipeline))
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}
	if dag.ComplexityScore != 8 {
		t.Errorf("expected ComplexityScore 8, got %d", dag.ComplexityScore)
	}
	if dag.Tier != "**Tier 1 (4–8 Sessions)**" {
		t.Errorf("expected Tier '**Tier 1 (4–8 Sessions)**', got %q", dag.Tier)
	}
}

func TestFlexibleSessionHeaders(t *testing.T) {
	pipeline := `# Pipeline
### D-001-S1: Deep Dive Decision Session
| Field | Value |
|---|---|
| **ID** | D-001-S1 |
| **Dependencies** | None |

` + "```prompt" + `
Prompt content
` + "```" + `

### Session T1-01 — Database Architecture
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Dependencies** | None |

` + "```prompt" + `
Database prompt content
` + "```" + `
`
	dag, err := ParsePipeline([]byte(pipeline))
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}
	if len(dag.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(dag.Sessions))
	}
	if dag.Sessions[0].ID != "D-001-S1" {
		t.Errorf("expected ID D-001-S1, got %s", dag.Sessions[0].ID)
	}
	if dag.Sessions[1].ID != "T1-01" {
		t.Errorf("expected ID T1-01, got %s", dag.Sessions[1].ID)
	}
}

func TestParsePipeline_StrayPromptBlockDoesNotPanic(t *testing.T) {
	// A pipeline with a prompt code block before any session header must not crash
	pipeline := `# Pipeline with intro block

Here is an example prompt template:
` + "```prompt" + `
Stray intro prompt
` + "```" + `

### T1-01: First Real Session
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Dependencies** | None |

` + "```prompt" + `
Real session prompt
` + "```" + `
`
	dag, err := ParsePipeline([]byte(pipeline))
	if err != nil {
		t.Fatalf("ParsePipeline: %v", err)
	}
	if len(dag.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(dag.Sessions))
	}
	if dag.Sessions[0].ID != "T1-01" {
		t.Errorf("expected ID T1-01, got %s", dag.Sessions[0].ID)
	}
	if dag.Sessions[0].Prompt != "Real session prompt" {
		t.Errorf("expected prompt 'Real session prompt', got %q", dag.Sessions[0].Prompt)
	}
}

func TestParsePipeline_HorizontalTablePlan(t *testing.T) {
	plan := `# Research Plan: D-015 Cache Layer

| Session ID | Role | Depends | Filename |
|---|---|---|---|
| D-015-S1 | Landscape | none | ` + "`sessions/D-015-S1-landscape.md`" + ` |
| D-015-S2 | Comparison | D-015-S1 | ` + "`sessions/D-015-S2-comparison.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: D-015-S1 Landscape
Map caching options.
` + "```" + `

` + "```prompt" + `
# RESEARCH BRIEF: D-015-S2 Comparison
Compare shortlisted options.
` + "```" + `
`
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
	if dag.Sessions[0].OutputFile != "sessions/D-015-S1-landscape.md" {
		t.Errorf("expected output file sessions/D-015-S1-landscape.md, got %s", dag.Sessions[0].OutputFile)
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
	plan := `# Research Plan: D-016 Messaging Architecture

## Overview
| Session ID | Role | Depends | Filename |
|---|---|---|---|
| D-016-S1 | Landscape | none | ` + "`sessions/D-016-S1-landscape.md`" + ` |
| D-016-S2 | Comparison | D-016-S1 | ` + "`sessions/D-016-S2-comparison.md`" + ` |

## Detailed Briefs

### D-016-S1: Messaging Landscape
` + "```prompt" + `
# RESEARCH BRIEF: D-016-S1 Messaging Landscape
Investigate pub/sub brokers.
` + "```" + `

### D-016-S2: NATS vs Kafka Comparison
` + "```prompt" + `
# RESEARCH BRIEF: D-016-S2 NATS vs Kafka Comparison
Compare latency and durability.
` + "```" + `
`
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePipeline error: %v", err)
	}
	if len(dag.Sessions) != 2 {
		t.Fatalf("expected exactly 2 deduplicated sessions, got %d", len(dag.Sessions))
	}
	s1 := dag.Sessions[0]
	if s1.ID != "D-016-S1" || s1.Title != "Messaging Landscape" || s1.OutputFile != "sessions/D-016-S1-landscape.md" {
		t.Errorf("unexpected s1 fields: %+v", s1)
	}
	if !strings.Contains(s1.Prompt, "Investigate pub/sub brokers.") {
		t.Errorf("unexpected s1 prompt: %q", s1.Prompt)
	}
	s2 := dag.Sessions[1]
	if s2.ID != "D-016-S2" || s2.Title != "NATS vs Kafka Comparison" || s2.OutputFile != "sessions/D-016-S2-comparison.md" {
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
	plan := `# Pipeline
#### T1-01: Session With Code Fence
| Field | Value |
|---|---|
| **ID** | T1-01 |

` + "```prompt" + `
# RESEARCH BRIEF: T1-01
Output format required:
` + "```yaml" + `
id: T1-01
status: complete
` + "```" + `
Conclusion: must not be truncated.
` + "```" + `
`
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePipeline error: %v", err)
	}
	if len(dag.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(dag.Sessions))
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
	content := `# Research Pipeline

### Session S1: Landscape
| **ID** | S1 |
| **Dependencies** | None |

` + "````prompt" + `
Execute landscape research
` + "````" + `

#### S2: Deep Dive
| **ID** | S2 |
| **Dependencies** | [S1] |

` + "````prompt" + `
Execute deep dive research
` + "````" + `

#### SYN: Synthesis
| **ID** | SYN |
| **Dependencies** | S1 and S2 |

` + "````prompt" + `
Synthesize all findings
` + "````" + `
`

	dag, err := ParsePipeline([]byte(content))
	if err != nil {
		t.Fatalf("ParsePipeline error: %v", err)
	}
	if len(dag.Sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(dag.Sessions))
	}
	if dag.Sessions[0].ID != "S1" || dag.Sessions[0].Title != "Landscape" {
		t.Errorf("unexpected session 0: %+v", dag.Sessions[0])
	}
	if dag.Sessions[1].ID != "S2" || len(dag.Sessions[1].Dependencies) != 1 || dag.Sessions[1].Dependencies[0] != "S1" {
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
	content := `# Research Pipeline

#### T1-01: Database Selection
| **ID** | T1-01 |
| **Dependencies** | None |

` + "````prompt" + `
<!-- Generated by Vivechak -->
<!-- Priority: High -->

# Session T1-01: Primary Datastore
Investigate database options.
` + "````" + `

#### T1-02: Auth Strategy
| **ID** | T1-02 |
| **Dependencies** | None |

` + "````prompt" + `
<!-- Note: Two-way door -->

# Session T1-02: Auth Strategy
Investigate auth options.
` + "````" + `
`

	dag, err := ParsePipeline([]byte(content))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	if len(dag.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(dag.Sessions))
	}
	if !strings.Contains(dag.Sessions[0].Prompt, "Primary Datastore") {
		t.Errorf("expected T1-01 prompt to contain 'Primary Datastore', got: %q", dag.Sessions[0].Prompt)
	}
	if !strings.Contains(dag.Sessions[1].Prompt, "Auth Strategy") {
		t.Errorf("expected T1-02 prompt to contain 'Auth Strategy', got: %q", dag.Sessions[1].Prompt)
	}
}

func TestParsePipeline_UnclosedPromptAtEOFWithLeadingLines(t *testing.T) {
	// A pipeline where the final prompt is unclosed at EOF (no closing fence)
	// and has comment/header lines before the session header line.
	content := `# Research Pipeline

#### T1-01: Database Selection
| **ID** | T1-01 |
| **Dependencies** | None |

` + "````prompt" + `
# Session T1-01: Primary Datastore
Investigate database options.
` + "````" + `

#### T1-02: Auth Strategy
| **ID** | T1-02 |
| **Dependencies** | None |

` + "````prompt" + `
<!-- Note: Two-way door -->
<!-- Generated by Vivechak -->

# Session T1-02: Auth Strategy
Investigate auth options at EOF.`

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

func TestIsSynthesisSession(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"FAD", true},
		{"fad", true},
		{"SYN", true},
		{"syn", true},
		{"SYN-01", true},
		{"syn-01", true},
		{"SYN-02", true},
		{"SYNTHESIS", true},
		{"T1-01", false},
		{"S1-01", false},
		{"D-001", false},
		{"COMP-01", false},
		{"", false},
		{"   ", false},
		{"  SYN-01  ", true},
		{"  FAD  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := IsSynthesisSession(tt.id)
			if got != tt.want {
				t.Errorf("IsSynthesisSession(%q) = %v; want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestParseDependencies_Comprehensive(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"None", nil},
		{"None (parallel)", nil},
		{"none", nil},
		{"—", nil},
		{"-", nil},
		{"N/A", nil},
		{"", nil},
		{"T1-01", []string{"T1-01"}},
		{"T1-01, T1-02", []string{"T1-01", "T1-02"}},
		{"T1-01 (soft), T1-02 (hard)", []string{"T1-01", "T1-02"}},
		{"[T1-01, T1-02]", []string{"T1-01", "T1-02"}},
		{"`T1-01` and `T1-02`", []string{"T1-01", "T1-02"}},
		{"T1-01 & T1-02", []string{"T1-01", "T1-02"}},
		{"[R-01](./R-01.md), [R-02](./R-02.md)", []string{"R-01", "R-02"}},
		{"- R-01\n- R-02\n- R-03", []string{"R-01", "R-02", "R-03"}},
		{"* `R-01` and `R-02`", []string{"R-01", "R-02"}},
		{"• R-01, • R-02", []string{"R-01", "R-02"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseDependencies(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseDependencies(%q) = %v (len %d); want %v (len %d)", tt.input, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseDependencies(%q)[%d] = %q; want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
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

func TestDAG_SerializeRoundTrip(t *testing.T) {
	inputPipeline := `# Research Pipeline: Katha Interactive Platform
### Project Parameters
Archetype: Interactive Web App
Constraints: Low Latency LLM Streaming

## Pipeline Topology

| ID | Topic | Dependencies | Output File |
|---|---|---|---|
| S1 | Persistence Landscape | none | sessions/S1.md |
| S2 | Real-time Sync | S1 | sessions/S2.md |

---

## Session Prompts

### Session S1: Persistence Landscape

| **Field** | **Value** |
|---|---|
| **ID** | S1 |
| **Layer** | 0 |
| **Dependencies** | none |
| **Output** | sessions/S1.md |
| **Door Type** | one-way |

` + "````prompt" + `
# RESEARCH BRIEF: S1
Investigate database options.
` + "````" + `

---

### Session S2: Real-time Sync

| **Field** | **Value** |
|---|---|
| **ID** | S2 |
| **Layer** | 1 |
| **Dependencies** | S1 |
| **Output** | sessions/S2.md |
| **Door Type** | two-way |

` + "````prompt" + `
# RESEARCH BRIEF: S2
Investigate synchronization options.
` + "````" + `

---

## Phase 0 Exit Gate Criteria
- [ ] B1: Staged Triangulation
- [ ] B2: Falsification Record
`

	dag, err := ParsePipeline([]byte(inputPipeline))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	if len(dag.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(dag.Sessions))
	}
	if !strings.Contains(dag.Preamble, "Archetype: Interactive Web App") {
		t.Errorf("expected Preamble to retain project parameters, got: %q", dag.Preamble)
	}
	if !strings.Contains(dag.Epilogue, "Phase 0 Exit Gate Criteria") {
		t.Errorf("expected Epilogue to retain exit gate criteria, got: %q", dag.Epilogue)
	}

	serialized := string(dag.Serialize())
	if !strings.Contains(serialized, "Archetype: Interactive Web App") {
		t.Errorf("serialized output lost project parameters")
	}
	if !strings.Contains(serialized, "Phase 0 Exit Gate Criteria") {
		t.Errorf("serialized output lost exit gate criteria")
	}
	if !strings.Contains(serialized, "````prompt") {
		t.Errorf("serialized output should use 4-backtick code fences")
	}
}

func TestDAG_ParsePipeline_UntaggedInnerCodeFence(t *testing.T) {
	input := `# Pipeline

## Session Prompts

### Session S1: Test

| **Field** | **Value** |
|---|---|
| **ID** | S1 |

` + "```prompt" + `
# Prompt Brief

Here is an example code block:
` + "```\n" + `
foo := 123
bar := 456
` + "```\n" + `
This instruction must not be truncated!
` + "```" + `

---
`

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








