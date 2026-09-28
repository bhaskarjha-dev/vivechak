package mcputil

import (
	"strings"
	"testing"
)

func TestUpdateOrAppendDecision(t *testing.T) {
	t.Run("empty existing file", func(t *testing.T) {
		res := updateOrAppendDecision(nil, "D-001", "decision content")
		s := string(res)
		if !strings.HasPrefix(s, "# Architectural Decisions\n\n") {
			t.Errorf("expected header, got %q", s)
		}
		if !strings.Contains(s, "<!-- DECISION: D-001 -->\ndecision content\n<!-- /DECISION: D-001 -->") {
			t.Errorf("expected decision entry, got %q", s)
		}
	})

	t.Run("append to existing decisions", func(t *testing.T) {
		existing := []byte("# Architectural Decisions\n\n<!-- DECISION: D-001 -->\nFirst\n<!-- /DECISION: D-001 -->\n")
		res := updateOrAppendDecision(existing, "D-002", "Second")
		s := string(res)
		if !strings.Contains(s, "<!-- DECISION: D-001 -->") {
			t.Errorf("lost D-001: %q", s)
		}
		if !strings.Contains(s, "<!-- DECISION: D-002 -->\nSecond\n<!-- /DECISION: D-002 -->") {
			t.Errorf("missing D-002: %q", s)
		}
	})

	t.Run("update existing decision in place", func(t *testing.T) {
		existing := []byte("# Architectural Decisions\n\n<!-- DECISION: D-001 -->\nOld content\n<!-- /DECISION: D-001 -->\n\n---\n\n<!-- DECISION: D-002 -->\nKeep me\n<!-- /DECISION: D-002 -->\n")
		res := updateOrAppendDecision(existing, "D-001", "New content")
		s := string(res)
		if strings.Contains(s, "Old content") {
			t.Errorf("found old content: %q", s)
		}
		if !strings.Contains(s, "<!-- DECISION: D-001 -->\nNew content\n<!-- /DECISION: D-001 -->") {
			t.Errorf("missing updated D-001: %q", s)
		}
		if !strings.Contains(s, "<!-- DECISION: D-002 -->\nKeep me\n<!-- /DECISION: D-002 -->") {
			t.Errorf("lost D-002: %q", s)
		}
	})

	t.Run("malformed marker ordering (end marker before start marker)", func(t *testing.T) {
		// Corrupted or manually edited file where closing marker appears before opening marker
		existing := []byte("<!-- /DECISION: D-001 -->\nsome text\n<!-- DECISION: D-001 -->\nunclosed content")
		res := updateOrAppendDecision(existing, "D-001", "Fixed content")
		s := string(res)
		// Should not panic, should append safely rather than slice backwards
		if !strings.HasSuffix(strings.TrimSpace(s), "<!-- /DECISION: D-001 -->") {
			t.Errorf("expected closing marker at end, got %q", s)
		}
		if !strings.Contains(s, "Fixed content") {
			t.Errorf("expected fixed content, got %q", s)
		}
	})
}
