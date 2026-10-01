package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInjectIntoPrompt_KnownContextSlot verifies [KNOWN_CONTEXT] placeholder injection.
func TestInjectIntoPrompt_KnownContextSlot(t *testing.T) {
	t.Run("fills KNOWN_CONTEXT when primary slot not present", func(t *testing.T) {
		prompt := "## KNOWN\n[KNOWN_CONTEXT]\n\n## APPROACH\nResearch here."
		content := "Upstream finding: PostgreSQL is recommended."
		result := injectIntoPrompt(prompt, UpstreamFindingsSlot, content)

		if strings.Contains(result, "[KNOWN_CONTEXT]") {
			t.Error("KNOWN_CONTEXT placeholder should have been replaced")
		}
		if !strings.Contains(result, "PostgreSQL") {
			t.Error("content should have been injected into KNOWN_CONTEXT slot")
		}
	})

	t.Run("fills both UPSTREAM_FINDINGS and KNOWN_CONTEXT when both present", func(t *testing.T) {
		prompt := "Based on upstream:\n[UPSTREAM_FINDINGS]\n\n## KNOWN\n[KNOWN_CONTEXT]\n\nInvestigate."
		content := "Upstream finding: Redis recommended."
		result := injectIntoPrompt(prompt, UpstreamFindingsSlot, content)

		if strings.Contains(result, "[UPSTREAM_FINDINGS]") {
			t.Error("UPSTREAM_FINDINGS placeholder should have been replaced")
		}
		if strings.Contains(result, "[KNOWN_CONTEXT]") {
			t.Error("KNOWN_CONTEXT placeholder should also have been replaced")
		}
		// Content should appear in both locations
		count := strings.Count(result, "Redis recommended")
		if count < 2 {
			t.Errorf("expected content in both slots, found %d occurrences", count)
		}
	})

	t.Run("falls back to KNOWN_CONTEXT when UPSTREAM_FINDINGS absent", func(t *testing.T) {
		prompt := "## KNOWN\n[KNOWN_CONTEXT]\n\nResearch here."
		content := "Prior research shows X."
		result := injectIntoPrompt(prompt, UpstreamFindingsSlot, content)

		if strings.Contains(result, "[KNOWN_CONTEXT]") {
			t.Error("should have used KNOWN_CONTEXT as fallback")
		}
		if !strings.Contains(result, "Prior research shows X") {
			t.Error("content should have been injected")
		}
	})

	t.Run("appends when neither slot present", func(t *testing.T) {
		prompt := "## APPROACH\nResearch here."
		content := "Upstream context."
		result := injectIntoPrompt(prompt, UpstreamFindingsSlot, content)

		if !strings.Contains(result, "UPSTREAM RESEARCH CONTEXT") {
			t.Error("should have appended with UPSTREAM RESEARCH CONTEXT header")
		}
		if !strings.Contains(result, "Upstream context") {
			t.Error("content should have been appended")
		}
	})

	t.Run("empty content returns prompt unchanged", func(t *testing.T) {
		prompt := "## KNOWN\n[KNOWN_CONTEXT]\n\nResearch."
		result := injectIntoPrompt(prompt, UpstreamFindingsSlot, "")

		if result != prompt {
			t.Errorf("empty content should return original prompt unchanged, got:\n%s", result)
		}
	})
}

// TestKnownContextSlotConstant verifies the constant value.
func TestKnownContextSlotConstant(t *testing.T) {
	if KnownContextSlot != "[KNOWN_CONTEXT]" {
		t.Errorf("KnownContextSlot = %q, want [KNOWN_CONTEXT]", KnownContextSlot)
	}
}

// TestInjectContext_KnownContextIntegration tests real file-based injection with KNOWN_CONTEXT.
func TestInjectContext_KnownContextIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	_ = os.MkdirAll(sessionsDir, 0o755)

	// Write upstream session
	upstream := `---
session_id: T1-01
title: Database Selection
date: 2026-09-27
status: complete
---

## Recommendation

Use PostgreSQL for this project. Grade A (official docs).

## Key Findings

- PostgreSQL 16 supports recursive CTEs efficiently
- Connection pooling via PgBouncer recommended
`
	_ = os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(upstream), 0o644)

	// Session T2-01 uses KNOWN_CONTEXT placeholder (8-block format)
	session := Session{
		ID:           "T2-01",
		Title:        "Schema Design",
		Dependencies: []string{"T1-01"},
		Prompt: `# RESEARCH BRIEF: Schema Design

## DECISION
How should we structure the database schema?

## KNOWN
[KNOWN_CONTEXT]

## APPROACH
Investigate schema patterns.`,
	}

	completed := map[string]bool{"T1-01": true}
	injected, err := InjectContext(session, sessionsDir, completed)
	if err != nil {
		t.Fatalf("InjectContext: %v", err)
	}

	// KNOWN_CONTEXT should be filled
	if strings.Contains(injected.InjectedPrompt, "[KNOWN_CONTEXT]") {
		t.Error("KNOWN_CONTEXT placeholder should have been replaced")
	}
	if !strings.Contains(injected.InjectedPrompt, "PostgreSQL") {
		t.Error("upstream findings should appear in injected prompt")
	}
}

// TestInjectContext_BothPlaceholders tests prompts with both UPSTREAM_FINDINGS and KNOWN_CONTEXT.
func TestInjectContext_BothPlaceholders(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	_ = os.MkdirAll(sessionsDir, 0o755)

	upstream := `---
session_id: T1-01
title: API Research
date: 2026-09-27
status: complete
---

## Recommendation

REST API with versioning. Grade B (benchmarks).
`
	_ = os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(upstream), 0o644)

	session := Session{
		ID:           "T2-01",
		Title:        "API Deep Dive",
		Dependencies: []string{"T1-01"},
		Prompt: `# RESEARCH BRIEF

## KNOWN
[KNOWN_CONTEXT]

Based on prior research:
[UPSTREAM_FINDINGS]

## APPROACH
Deep dive.`,
	}

	completed := map[string]bool{"T1-01": true}
	injected, err := InjectContext(session, sessionsDir, completed)
	if err != nil {
		t.Fatalf("InjectContext: %v", err)
	}

	if strings.Contains(injected.InjectedPrompt, "[UPSTREAM_FINDINGS]") {
		t.Error("UPSTREAM_FINDINGS should have been replaced")
	}
	if strings.Contains(injected.InjectedPrompt, "[KNOWN_CONTEXT]") {
		t.Error("KNOWN_CONTEXT should also have been replaced")
	}
}
