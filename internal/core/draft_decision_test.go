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
}
