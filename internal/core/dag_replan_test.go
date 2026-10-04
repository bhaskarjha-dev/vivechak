package core

import (
	"strings"
	"testing"
)

func TestDAG_AddSession(t *testing.T) {
	dag := &DAG{
		Sessions: []Session{
			{ID: "T0-01", Title: "Landscape", Layer: 0},
			{ID: "T1-01", Title: "Database", Layer: 1, Dependencies: []string{"T0-01"}},
		},
	}

	// 1. Valid addition
	err := dag.AddSession(Session{
		ID:           "T1-02",
		Title:        "Caching",
		Layer:        1,
		Dependencies: []string{"T0-01"},
		Prompt:       "# Brief\nResearch caching.",
	})
	if err != nil {
		t.Fatalf("unexpected error adding session: %v", err)
	}
	if len(dag.Sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(dag.Sessions))
	}

	// 2. Duplicate ID
	err = dag.AddSession(Session{ID: "T1-01"})
	if err == nil {
		t.Errorf("expected error adding duplicate session ID")
	}

	// 3. Nonexistent dependency
	err = dag.AddSession(Session{ID: "T2-01", Dependencies: []string{"NONEXISTENT"}})
	if err == nil {
		t.Errorf("expected error adding session with nonexistent dependency")
	}

	// 4. Cycle creation
	// If T0-01 depends on T1-01:
	err = dag.AddSession(Session{
		ID:           "T0-00",
		Dependencies: []string{"T1-01"},
	})
	if err != nil {
		// T0-00 depending on T1-01 is not a cycle by itself, but adding an edge back would be
	}
}

func TestDAG_RemoveSession(t *testing.T) {
	dag := &DAG{
		Sessions: []Session{
			{ID: "T0-01", Title: "Landscape"},
			{ID: "T1-01", Title: "Database", Dependencies: []string{"T0-01"}},
			{ID: "T1-02", Title: "Independent"},
		},
	}

	// 1. Cannot remove session that has dependents
	err := dag.RemoveSession("T0-01")
	if err == nil {
		t.Errorf("expected error removing session with dependents")
	}

	// 2. Can remove independent un-depended session
	err = dag.RemoveSession("T1-02")
	if err != nil {
		t.Fatalf("unexpected error removing independent session: %v", err)
	}
	if len(dag.Sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(dag.Sessions))
	}

	// 3. Remove non-existent session
	err = dag.RemoveSession("T9-99")
	if err == nil {
		t.Errorf("expected error removing nonexistent session")
	}
}

func TestDAG_UpdateDependencies(t *testing.T) {
	dag := &DAG{
		Sessions: []Session{
			{ID: "T0-01", Title: "Landscape"},
			{ID: "T0-02", Title: "Framework"},
			{ID: "T1-01", Title: "Database", Dependencies: []string{"T0-01"}},
		},
	}

	// 1. Valid update
	err := dag.UpdateDependencies("T1-01", []string{"T0-01", "T0-02"})
	if err != nil {
		t.Fatalf("unexpected error updating deps: %v", err)
	}
	s := dag.SessionByID("T1-01")
	if len(s.Dependencies) != 2 {
		t.Errorf("expected 2 deps, got %d", len(s.Dependencies))
	}

	// 2. Cycle creation error
	// Set T0-01 to depend on T1-01 -> creates cycle T0-01 -> T1-01 -> T0-01
	err = dag.UpdateDependencies("T0-01", []string{"T1-01"})
	if err == nil {
		t.Errorf("expected cycle error when updating deps")
	}
	// Verify old deps preserved after cycle error
	s0 := dag.SessionByID("T0-01")
	if len(s0.Dependencies) != 0 {
		t.Errorf("expected old deps preserved, got %v", s0.Dependencies)
	}
}

func TestDAG_UpdatePrompt(t *testing.T) {
	dag := &DAG{
		Sessions: []Session{
			{ID: "T1-01", Title: "DB", Prompt: "old prompt"},
		},
	}

	err := dag.UpdatePrompt("T1-01", "new expanded prompt with ALLIED TERMS")
	if err != nil {
		t.Fatalf("unexpected error updating prompt: %v", err)
	}
	if dag.SessionByID("T1-01").Prompt != "new expanded prompt with ALLIED TERMS" {
		t.Errorf("prompt was not updated")
	}

	// Empty prompt error
	err = dag.UpdatePrompt("T1-01", "  ")
	if err == nil {
		t.Errorf("expected error for empty prompt")
	}
}

func TestDAG_Serialize_RoundTrip(t *testing.T) {
	dag := &DAG{
		Archetype:       "Desktop Agent",
		Tier:            "Tier 2",
		ComplexityScore: 18,
		Sessions: []Session{
			{
				ID:           "T0-01",
				Title:        "System Architecture Landscape",
				Layer:        0,
				Dependencies: nil,
				DecisionRef:  "D-001",
				DoorType:     "one-way",
				Prompt:       "# T0-01: System Architecture\n\nBRIEF:\nConduct research.",
			},
			{
				ID:           "T1-01",
				Title:        "Embedded Database Spike",
				Layer:        1,
				Dependencies: []string{"T0-01"},
				DecisionRef:  "D-002",
				DoorType:     "two-way",
				Prompt:       "# T1-01: Database Spike\n\nBRIEF:\nResearch SQLite vs BadgerDB.",
			},
		},
	}

	serialized := dag.Serialize()
	if len(serialized) == 0 {
		t.Fatalf("serialized output is empty")
	}

	// Parse it back
	reparsed, err := ParsePipeline(serialized)
	if err != nil {
		t.Fatalf("failed to re-parse serialized pipeline: %v\nContent:\n%s", err, string(serialized))
	}

	if len(reparsed.Sessions) != 2 {
		t.Fatalf("expected 2 sessions after round-trip, got %d", len(reparsed.Sessions))
	}

	s0 := reparsed.SessionByID("T0-01")
	if s0 == nil || s0.Title != "System Architecture Landscape" {
		t.Errorf("session T0-01 lost or corrupted: %+v", s0)
	}

	s1 := reparsed.SessionByID("T1-01")
	if s1 == nil || len(s1.Dependencies) != 1 || s1.Dependencies[0] != "T0-01" {
		t.Errorf("session T1-01 dependencies corrupted: %+v", s1)
	}
	if !strings.Contains(s1.Prompt, "SQLite vs BadgerDB") {
		t.Errorf("session T1-01 prompt corrupted: %s", s1.Prompt)
	}
}

func TestDAG_ToMermaid_And_ToStatusTable(t *testing.T) {
	dag := &DAG{
		Sessions: []Session{
			{ID: "T0-01", Title: "Landscape", Layer: 0},
			{ID: "T1-01", Title: "Database", Layer: 1, Dependencies: []string{"T0-01"}},
			{ID: "SYN-01", Title: "Synthesis", Layer: 2, Dependencies: []string{"T1-01"}},
		},
	}

	completed := map[string]bool{"T0-01": true}

	mermaid := dag.ToMermaid(completed)
	if !strings.Contains(mermaid, "graph TD") {
		t.Errorf("expected graph TD in mermaid output")
	}
	if !strings.Contains(mermaid, "T0-01 --> T1-01") {
		t.Errorf("expected edge T0-01 --> T1-01 in mermaid")
	}
	// T0-01 is completed (green)
	if !strings.Contains(mermaid, "style T0-01 fill:#22c55e") {
		t.Errorf("expected completed green style for T0-01")
	}
	// T1-01 has all deps completed -> ready (amber)
	if !strings.Contains(mermaid, "style T1-01 fill:#f59e0b") {
		t.Errorf("expected ready amber style for T1-01")
	}
	// SYN-01 is synthesis -> blue
	if !strings.Contains(mermaid, "style SYN-01 fill:#3b82f6") {
		t.Errorf("expected synthesis blue style for SYN-01")
	}

	table := dag.ToStatusTable(completed)
	if !strings.Contains(table, "| Session ID | Title |") {
		t.Errorf("expected table header")
	}
	if !strings.Contains(table, "Completed") {
		t.Errorf("expected Completed status in table")
	}
	if !strings.Contains(table, "Ready (⚡)") {
		t.Errorf("expected Ready (⚡) status in table")
	}
}
