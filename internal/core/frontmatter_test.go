package core

import (
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
