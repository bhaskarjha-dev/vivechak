package core

import (
	"strings"
	"testing"
)

func TestDraftDecisionFromSession(t *testing.T) {
	t.Run("empty session content returns error", func(t *testing.T) {
		_, err := DraftDecisionFromSession(nil, "D-001")
		if err == nil {
			t.Fatal("expected error for nil session content")
		}
		_, err = DraftDecisionFromSession([]byte(""), "D-001")
		if err == nil {
			t.Fatal("expected error for empty session content")
		}
	})

	t.Run("drafts decision with full sections and evidence grades", func(t *testing.T) {
		sessionContent := `---
session_id: R-01
title: Bank Connectivity Strategy
date: 2026-10-01
status: complete
---

# Session R-01: Bank Connectivity Strategy

## Evaluated Options & Alternatives
1. Plaid: Full coverage, higher API cost. Grade A (direct Plaid documentation)
2. SimpleFIN: Lightweight, read-only focus. Grade B (API analysis)
3. MX: Enterprise aggregator. Grade C (community reports)

## Recommendations & Verdict
We recommend starting with SimpleFIN for prototype velocity while maintaining an adapter interface for Plaid. Grade A (empirical spike)

## Discovered Concerns & Failure Modes
- Normalization burden across multiple regional banks.
- Token refresh failure causing silent sync degradation. Grade B (RFC analysis)
`

		draft, err := DraftDecisionFromSession([]byte(sessionContent), "D-001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify frontmatter fields
		fm, body, err := ParseFrontmatter([]byte(draft))
		if err != nil {
			t.Fatalf("draft frontmatter parsing error: %v", err)
		}
		if fm.GetString("id") != "D-001" {
			t.Errorf("expected id D-001, got %q", fm.GetString("id"))
		}
		if fm.GetString("title") != "Bank Connectivity Strategy" {
			t.Errorf("expected title 'Bank Connectivity Strategy', got %q", fm.GetString("title"))
		}
		if fm.GetString("status") != "proposed" {
			t.Errorf("expected status 'proposed', got %q", fm.GetString("status"))
		}
		if fm.GetString("door_type") != "one-way" {
			t.Errorf("expected door_type 'one-way', got %q", fm.GetString("door_type"))
		}
		sessions := fm.GetStringSlice("informed_by_sessions")
		if len(sessions) != 1 || sessions[0] != "R-01" {
			t.Errorf("expected informed_by_sessions [R-01], got %v", sessions)
		}

		// Verify body sections
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "# D-001: Bank Connectivity Strategy") {
			t.Errorf("missing title in body: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "SimpleFIN") {
			t.Errorf("missing recommendation content in body: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "Normalization burden") {
			t.Errorf("missing concerns content in body: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "## Challenge Findings & Red-Team Stress Test") {
			t.Errorf("missing challenge findings section in body: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "## Evidence Summary (from session)") {
			t.Errorf("missing evidence summary section: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "Grade A") {
			t.Errorf("missing extracted evidence grades: %s", bodyStr)
		}
	})

	t.Run("handles minimal session without headings gracefully", func(t *testing.T) {
		sessionContent := `---
id: S-99
---
Just raw notes without standard headings.
`
		draft, err := DraftDecisionFromSession([]byte(sessionContent), "D-099")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(draft, "id: D-099") {
			t.Errorf("expected D-099 in draft: %s", draft)
		}
		if !strings.Contains(draft, "informed_by_sessions: [\"S-99\"]") {
			t.Errorf("expected S-99 in informed_by_sessions: %s", draft)
		}
	})

	t.Run("preserves multi-option subheadings and multiple risks without premature truncation", func(t *testing.T) {
		sessionContent := `---
session_id: R-03
title: Primary Datastore Selection
date: 2026-10-02
status: complete
---

# Session R-03: Primary Datastore Selection

## Evaluated Options & Alternatives

### Option 1: SQLite with WAL
SQLite embedded engine. Grade A (docs)

### Option 2: PostgreSQL
Postgres relational database. Grade A (docs)

### Option 3: DuckDB
DuckDB analytical engine. Grade A (docs)

## Recommendations & Verdict
We recommend SQLite with WAL. Grade A (benchmark)

## Discovered Concerns & Failure Modes

### Risk 1: Concurrency Limits
Single writer limitation.

### Risk 2: Network Shares
Locking issues on NFS.
`
		draft, err := DraftDecisionFromSession([]byte(sessionContent), "D-003")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		bodyStr := draft
		if !strings.Contains(bodyStr, "Option 1: SQLite with WAL") {
			t.Errorf("missing Option 1 in draft: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "Option 2: PostgreSQL") {
			t.Errorf("missing Option 2 in draft: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "Option 3: DuckDB") {
			t.Errorf("missing Option 3 in draft: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "Risk 1: Concurrency Limits") {
			t.Errorf("missing Risk 1 in draft: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "Risk 2: Network Shares") {
			t.Errorf("missing Risk 2 in draft: %s", bodyStr)
		}
	})
}

func TestExtractSection_HeadingHierarchy(t *testing.T) {
	doc := `# Document Title
Intro text before any sections.

## Evaluated Options
Overview of options.

### Option 1: Embedded SQLite
SQLite details.
#### Option 1.1: WAL Configuration
WAL details here.

### Option 2: Server PostgreSQL
PostgreSQL details.

## Recommendations
Final verdict is Option 1.

## Other Notes
Final notes.
`

	t.Run("extracts options with child and grandchild headings", func(t *testing.T) {
		options := ExtractSection(doc, "option", "evaluated")
		if !strings.Contains(options, "Overview of options.") {
			t.Errorf("missing section overview: %s", options)
		}
		if !strings.Contains(options, "### Option 1: Embedded SQLite") {
			t.Errorf("missing Option 1 heading: %s", options)
		}
		if !strings.Contains(options, "#### Option 1.1: WAL Configuration") {
			t.Errorf("missing Option 1.1 grandchild heading: %s", options)
		}
		if !strings.Contains(options, "### Option 2: Server PostgreSQL") {
			t.Errorf("missing Option 2 heading: %s", options)
		}
		// Sibling heading ## Recommendations must NOT be included
		if strings.Contains(options, "Recommendations") || strings.Contains(options, "Final verdict") {
			t.Errorf("section extraction did not stop at sibling heading: %s", options)
		}
	})

	t.Run("extracts recommendations section terminated by another sibling", func(t *testing.T) {
		recs := ExtractSection(doc, "recommend")
		if !strings.Contains(recs, "Final verdict is Option 1.") {
			t.Errorf("missing recommendation content: %s", recs)
		}
		if strings.Contains(recs, "Final notes") {
			t.Errorf("recommendation did not stop at next sibling: %s", recs)
		}
	})

	t.Run("returns empty string when no keyword matches", func(t *testing.T) {
		res := ExtractSection(doc, "nonexistent-keyword")
		if res != "" {
			t.Errorf("expected empty string, got: %q", res)
		}
	})

	t.Run("returns empty string on empty body", func(t *testing.T) {
		res := ExtractSection("", "option")
		if res != "" {
			t.Errorf("expected empty string on empty body, got: %q", res)
		}
	})

	t.Run("preserves internal blank lines within section", func(t *testing.T) {
		multiPara := "## Details\nParagraph 1.\n\nParagraph 2.\n\n- Bullet 1\n- Bullet 2\n\n## Next\nOther."
		extracted := ExtractSection(multiPara, "details")
		expected := "Paragraph 1.\n\nParagraph 2.\n\n- Bullet 1\n- Bullet 2"
		if extracted != expected {
			t.Errorf("expected %q, got %q", expected, extracted)
		}
	})
}

// TestDraftDecision_NoADRYet_TakesPlanDecisionTitle covers HIGH-13: with no ADR yet, record_decision passes the
// decision's plan as the existing record, and the draft takes the plan's decision title (its heading, its plan
// block, or a real frontmatter title) over the session title; a fixed or placeholder title never wins.
func TestDraftDecision_NoADRYet_TakesPlanDecisionTitle(t *testing.T) {
	session := []byte("---\nid: D-001-S1\ntitle: Landscape\nstatus: complete\n---\n## Recommendation\nUse PostgreSQL. Grade A (docs)\n")
	cases := []struct{ name, plan, want string }{
		{"heading", "---\nid: D-001\ntype: plan\ntitle: \"[Decision title]\"\n---\n# D-001-PLAN: Event storage engine\n\n### D-001-S1: Landscape\n```prompt\n# RESEARCH BRIEF: datastore landscape\n```\n", "Event storage engine"},
		{"heading wins over a stale frontmatter title", "---\nid: D-001\ntype: plan\ntitle: Research Plan\n---\n# D-001-PLAN: Event storage engine\n", "Event storage engine"},
		{"frontmatter title written by the plan save", "---\nid: D-001\ntype: plan\ntitle: Event storage engine\n---\n## Routing\n", "Event storage engine"},
		{"plan block decision title", "---\nid: D-001\ntype: plan\ntitle: \"[Decision title]\"\n---\n```yaml vivechak-plan\nplan: D-001\nsessions:\n  - {id: D-001-S1, title: Landscape, depends: []}\ndecisions:\n  - {id: D-001, title: Primary datastore, door: one-way}\n```\n", "Primary datastore"},
		{"a heading inside a prompt fence is not the plan's title", "---\nid: D-001\ntype: plan\ntitle: \"[Decision title]\"\n---\n```prompt\n# RESEARCH BRIEF: datastore landscape\n```\n", "Landscape"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			draft, err := DraftDecisionFromSessionWithExisting([]byte(tc.plan), session, "D-001")
			if err != nil {
				t.Fatalf("draft: %v", err)
			}
			fm, _, _ := ParseFrontmatter([]byte(draft))
			if got := fm.GetString("title"); got != tc.want {
				t.Errorf("draft title = %q, want %q", got, tc.want)
			}
			if fm.GetString("status") != "proposed" {
				t.Errorf("a draft from a plan must be proposed, got %q", fm.GetString("status"))
			}
		})
	}
}

func TestPlanDecisionTitle(t *testing.T) {
	cases := []struct{ name, plan, want string }{
		{"id-plan heading", "# D-001-PLAN: Event storage engine\n", "Event storage engine"},
		{"plain heading", "# Event storage engine\n## Sessions\n", "Event storage engine"},
		{"heading without a title", "# D-001-PLAN\n", ""},
		{"placeholder heading", "# [Decision title]\n", ""},
		{"only fenced headings", "````markdown\n# Not the title\n```prompt\n# Nor this\n```\n````\n", ""},
		{"single decision in the plan block", "```yaml vivechak-plan\nplan: PROJECT\nsessions:\n  - {id: T1-01, title: Datastore, depends: []}\ndecisions:\n  - {id: D-001, title: Primary datastore, door: one-way}\n```\n", "Primary datastore"},
		{"nothing declared", "## Routing\n| R | 3 |\n", ""},
	}
	for _, tc := range cases {
		if got := PlanDecisionTitle([]byte(tc.plan)); got != tc.want {
			t.Errorf("%s: PlanDecisionTitle = %q, want %q", tc.name, got, tc.want)
		}
	}
}
