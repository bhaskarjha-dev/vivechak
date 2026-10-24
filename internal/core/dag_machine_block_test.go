package core

import (
	"strings"
	"testing"
)

func TestParsePlanMachineBlock_IgnoresMarkdownTables(t *testing.T) {
	planWithTables := `---
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

` + "```yaml vivechak-plan\n" +
		`sessions:
  - id: T1-01
    title: Datastore Architecture
    door: one-way
    dependencies: []
    output_file: research/sessions/T1-01.md
    decision_id: D-001
  - id: T1-02
    title: Caching Strategy
    door: two-way
    dependencies: [T1-01]
    output_file: research/sessions/T1-02.md
decisions:
  - id: D-001
    title: Datastore Selection
    session_id: T1-01
    door: one-way
` + "```\n" + `

### T1-01: Datastore Architecture
Prompt content for T1-01.

### T1-02: Caching Strategy
Prompt content for T1-02.
`

	pipeline, err := ParsePipeline([]byte(planWithTables))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	// It should have exactly 2 sessions: T1-01 and T1-02
	if len(pipeline.Sessions) != 2 {
		t.Fatalf("expected exactly 2 sessions from machine block, got %d: %v", len(pipeline.Sessions), pipeline.Sessions)
	}

	if pipeline.Sessions[0].ID != "T1-01" || pipeline.Sessions[0].Title != "Datastore Architecture" {
		t.Errorf("session 0 mismatch: %+v", pipeline.Sessions[0])
	}
	if pipeline.Sessions[1].ID != "T1-02" || pipeline.Sessions[1].Title != "Caching Strategy" {
		t.Errorf("session 1 mismatch: %+v", pipeline.Sessions[1])
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
	origPlan := `# Top-Level Architectural Manifesto
This is a very specific manifesto written by the architect that MUST NOT be lost.

Here are some non-standard custom sections:
- Custom guideline 1
- Custom guideline 2

` + "```yaml vivechak-plan\n" +
		`sessions:
  - id: T1-01
    title: Initial Title
    door: two-way
    dependencies: []
    output_file: research/sessions/T1-01.md
` + "```\n" + `

### Session T1-01: Initial Title
Original prompt instructions.

### Bottom Section
Final notes and signing instructions.
`

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

	if !strings.Contains(resStr, "# Top-Level Architectural Manifesto") {
		t.Errorf("top-level manifesto was lost: %s", resStr)
	}
	if !strings.Contains(resStr, "Custom guideline 1") {
		t.Errorf("custom guideline was lost: %s", resStr)
	}
	if !strings.Contains(resStr, "### Bottom Section") {
		t.Errorf("bottom section was lost: %s", resStr)
	}
	if !strings.Contains(resStr, "Conduct in-depth caching strategy research") {
		t.Errorf("new session prompt was not inserted: %s", resStr)
	}
	if !strings.Contains(resStr, "T1-02") {
		t.Errorf("machine block was not updated with T1-02: %s", resStr)
	}
}
