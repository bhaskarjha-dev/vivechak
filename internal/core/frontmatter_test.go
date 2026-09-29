package core

import (
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	input := `---
session_id: R-01
title: Test Session
date: 2026-09-27
status: complete
tags:
  - database
  - architecture
---

# Body Content

This is the body.
`

	fm, body, err := ParseFrontmatter([]byte(input))
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}

	if fm == nil {
		t.Fatal("frontmatter should not be nil")
	}

	if fm.GetString("session_id") != "R-01" {
		t.Errorf("session_id: got %q, want %q", fm.GetString("session_id"), "R-01")
	}
	if fm.GetString("title") != "Test Session" {
		t.Errorf("title: got %q, want %q", fm.GetString("title"), "Test Session")
	}
	if fm.GetString("status") != "complete" {
		t.Errorf("status: got %q, want %q", fm.GetString("status"), "complete")
	}

	tags := fm.GetStringSlice("tags")
	if len(tags) != 2 {
		t.Errorf("tags: got %d items, want 2", len(tags))
	}

	if !fm.Has("date") {
		t.Error("should have 'date' key")
	}
	if fm.Has("nonexistent") {
		t.Error("should not have 'nonexistent' key")
	}

	bodyStr := string(body)
	if len(bodyStr) == 0 {
		t.Error("body should not be empty")
	}
}

func TestParseFrontmatter_DashesInYAML(t *testing.T) {
	input := []byte(`---
title: My Document
description: |
  This has a line
  ---
  that looks like a delimiter
---

# Body here`)

	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter")
	}
	if fm.GetString("title") != "My Document" {
		t.Errorf("expected title 'My Document', got %q", fm.GetString("title"))
	}
	if !strings.Contains(string(body), "# Body here") {
		t.Errorf("body should contain '# Body here', got %q", string(body))
	}
}

func TestParseFrontmatter_None(t *testing.T) {
	input := `# Just a Markdown file

No frontmatter here.
`
	fm, body, err := ParseFrontmatter([]byte(input))
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	if fm != nil {
		t.Error("frontmatter should be nil when not present")
	}
	if len(body) == 0 {
		t.Error("body should contain the full content")
	}
}

func TestParseFrontmatter_Unclosed(t *testing.T) {
	input := `---
session_id: R-01
title: Unclosed Test
No closing dashes here...
`
	fm, _, err := ParseFrontmatter([]byte(input))
	if err == nil {
		t.Fatal("expected error for unclosed frontmatter block, got nil")
	}
	if !strings.Contains(err.Error(), "unclosed frontmatter block") {
		t.Errorf("expected 'unclosed frontmatter block' error, got %v", err)
	}
	if fm != nil {
		t.Errorf("expected nil frontmatter, got %v", fm)
	}
}

func TestParseFrontmatter_LeadingWhitespace(t *testing.T) {
	input := []byte("\n\n  ---\nsession_id: R-01\ntitle: Test\n---\n\n# Body\n")
	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter with leading whitespace")
	}
	if fm.GetString("session_id") != "R-01" {
		t.Errorf("session_id: got %q, want %q", fm.GetString("session_id"), "R-01")
	}
	if !strings.Contains(string(body), "# Body") {
		t.Errorf("body should contain '# Body', got %q", string(body))
	}
}

func TestParseFrontmatter_BOM(t *testing.T) {
	// UTF-8 BOM: EF BB BF
	input := append([]byte{0xEF, 0xBB, 0xBF}, []byte("---\ntitle: BOM Test\n---\n\n# Body\n")...)
	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter with BOM")
	}
	if fm.GetString("title") != "BOM Test" {
		t.Errorf("title: got %q, want %q", fm.GetString("title"), "BOM Test")
	}
	if !strings.Contains(string(body), "# Body") {
		t.Errorf("body should contain '# Body', got %q", string(body))
	}
}

func TestParseFrontmatter_CRLF(t *testing.T) {
	input := []byte("---\r\nsession_id: R-01\r\ntitle: CRLF Test\r\n---\r\n\r\n# Body\r\n")
	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter with CRLF")
	}
	if fm.GetString("session_id") != "R-01" {
		t.Errorf("session_id: got %q, want %q", fm.GetString("session_id"), "R-01")
	}
	if len(body) == 0 {
		t.Error("body should not be empty")
	}
}

func TestComposeFrontmatter(t *testing.T) {
	fm := Frontmatter{
		"session_id": "R-01",
		"title":      "Test",
		"status":     "draft",
	}
	body := []byte("# Body\n\nContent here.\n")

	result, err := ComposeFrontmatter(fm, body)
	if err != nil {
		t.Fatalf("ComposeFrontmatter: %v", err)
	}

	// Should start with ---
	if string(result[:3]) != "---" {
		t.Error("composed output should start with ---")
	}

	// Should contain the body
	resultStr := string(result)
	if !contains(resultStr, "# Body") {
		t.Error("composed output should contain the body")
	}

	// Round-trip: parse back
	fm2, body2, err := ParseFrontmatter(result)
	if err != nil {
		t.Fatalf("round-trip ParseFrontmatter: %v", err)
	}
	if fm2.GetString("session_id") != "R-01" {
		t.Errorf("round-trip session_id: got %q, want %q", fm2.GetString("session_id"), "R-01")
	}
	if len(body2) == 0 {
		t.Error("round-trip body should not be empty")
	}
}

func TestValidateSession_Valid(t *testing.T) {
	input := `---
session_id: R-01
title: Database Selection
date: 2026-09-27
status: complete
---

# Database Selection

PostgreSQL is recommended. A (official documentation)

## Recommendations

Use PostgreSQL 16 with pgvector.
`
	result := ValidateSession([]byte(input))
	if result.HasBlocking() {
		t.Errorf("valid session should not have blocking issues: %v", result.Issues)
	}
}

func TestValidateSession_MissingFrontmatter(t *testing.T) {
	input := `# No Frontmatter Session

Just body content.
`
	result := ValidateSession([]byte(input))
	if !result.HasBlocking() {
		t.Error("missing frontmatter should produce blocking issues")
	}
	if result.Status != "draft" {
		t.Errorf("status should be 'draft', got %q", result.Status)
	}
}

func TestValidateSession_NoEvidenceGrades(t *testing.T) {
	input := `---
session_id: R-01
title: Test
date: 2026-09-27
---

# Test Session

Some findings without any evidence grades.
`
	result := ValidateSession([]byte(input))
	if result.WarningCount() == 0 {
		t.Error("session without evidence grades should have warnings")
	}
}

func TestValidateDecision(t *testing.T) {
	input := `---
decision_id: D-001
title: Primary Datastore Selection
status: accepted
door_type: one-way
---

# Context

We need a database for our application.

# Decision

Use PostgreSQL 16 with pgvector extension.

# Consequences

Single operational surface. Need to monitor graph traversal performance.
`
	result := ValidateDecision([]byte(input))
	if result.HasBlocking() {
		t.Errorf("valid decision should not have blocking issues: %v", result.Issues)
	}
}

func TestValidateSession_WithIdAlias(t *testing.T) {
	inputSession := `---
id: T1-01
title: Test
date: 2026-09-27
status: complete
---
# Test Session
Findings... A (doc)
`
	resultSession := ValidateSession([]byte(inputSession))
	for _, issue := range resultSession.Issues {
		if issue.Code == "V-MISSING-FIELD" && issue.Field == "session_id" {
			t.Errorf("should not have V-MISSING-FIELD for session_id when 'id' is present")
		}
	}
	if resultSession.HasBlocking() {
		t.Errorf("session with 'id' alias should be valid, got issues: %v", resultSession.Issues)
	}

	inputDecision := `---
id: D-001
title: Primary Datastore Selection
status: accepted
door_type: one-way
---
# Context
We need a database. We evaluated several databases, and PostgreSQL seems to be the best for our needs, given pgvector support.
# Decision
Use PostgreSQL.
# Consequences
Single operational surface. It is very easy to use and well documented. A (doc)
`
	resultDecision := ValidateDecision([]byte(inputDecision))
	for _, issue := range resultDecision.Issues {
		if issue.Code == "V-MISSING-FIELD" && issue.Field == "decision_id" {
			t.Errorf("should not have V-MISSING-FIELD for decision_id when 'id' is present")
		}
	}
	if resultDecision.HasBlocking() || resultDecision.WarningCount() > 0 {
		t.Errorf("decision with 'id' alias should be valid, got issues: %v", resultDecision.Issues)
	}
}

func TestExtractSessionID(t *testing.T) {
	t.Run("session_id takes precedence", func(t *testing.T) {
		fm := Frontmatter{"session_id": "T1-01", "id": "other"}
		if got := ExtractSessionID("T1-01-database.md", fm); got != "T1-01" {
			t.Errorf("expected T1-01, got %q", got)
		}
	})

	t.Run("id alias used when session_id absent", func(t *testing.T) {
		fm := Frontmatter{"id": "T1-02"}
		if got := ExtractSessionID("T1-02-auth.md", fm); got != "T1-02" {
			t.Errorf("expected T1-02, got %q", got)
		}
	})

	t.Run("fallback to full filename stem without SplitN truncation", func(t *testing.T) {
		if got := ExtractSessionID("T1-01-database-selection.md", nil); got != "T1-01-database-selection" {
			t.Errorf("expected T1-01-database-selection, got %q", got)
		}
		emptyFm := Frontmatter{}
		if got := ExtractSessionID("T1-01-database-selection.md", emptyFm); got != "T1-01-database-selection" {
			t.Errorf("expected T1-01-database-selection, got %q", got)
		}
	})
}

func TestParseFrontmatter_HorizontalRule(t *testing.T) {
	input := `----------------------------------------
# Document Title

This document starts with a horizontal rule.
----------------------------------------
`
	fm, body, err := ParseFrontmatter([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error on horizontal rule: %v", err)
	}
	if fm != nil {
		t.Errorf("expected nil frontmatter for document starting with HR, got: %v", fm)
	}
	if !strings.Contains(string(body), "Document Title") {
		t.Errorf("expected body to contain document title")
	}
}

func TestFrontmatter_Set(t *testing.T) {
	fm := Frontmatter{"session_id": "T1-01"}
	fm.Set("status", "draft")
	if fm.GetString("status") != "draft" {
		t.Errorf("expected status 'draft', got %q", fm.GetString("status"))
	}
	var nilFm Frontmatter
	nilFm.Set("status", "draft") // Should not panic
}

