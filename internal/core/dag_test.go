package core

import (
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
		{"None", nil},
		{"—", nil},
		{"T1-01", []string{"T1-01"}},
		{"T1-01, T1-02", []string{"T1-01", "T1-02"}},
		{"T2-01, T1-02", []string{"T2-01", "T1-02"}},
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
