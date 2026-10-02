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
		options := extractSection(doc, "option", "evaluated")
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
		recs := extractSection(doc, "recommend")
		if !strings.Contains(recs, "Final verdict is Option 1.") {
			t.Errorf("missing recommendation content: %s", recs)
		}
		if strings.Contains(recs, "Final notes") {
			t.Errorf("recommendation did not stop at next sibling: %s", recs)
		}
	})

	t.Run("returns empty string when no keyword matches", func(t *testing.T) {
		res := extractSection(doc, "nonexistent-keyword")
		if res != "" {
			t.Errorf("expected empty string, got: %q", res)
		}
	})

	t.Run("returns empty string on empty body", func(t *testing.T) {
		res := extractSection("", "option")
		if res != "" {
			t.Errorf("expected empty string on empty body, got: %q", res)
		}
	})
}


