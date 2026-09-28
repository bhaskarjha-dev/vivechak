package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectContext_Dependencies(t *testing.T) {
	// Create temp dir with session files
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	os.MkdirAll(sessionsDir, 0o755)

	// Write upstream session T1-01
	t101Content := `---
session_id: T1-01
title: Database Selection
date: 2026-09-27
status: complete
---

# Database Selection

## Recommendations

PostgreSQL 16 is recommended for this use case. A (official docs)

## Findings

- PostgreSQL handles graph traversal up to 50k nodes at <50ms p95
- Neo4j adds operational overhead not justified at this scale
`
	os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(t101Content), 0o644)

	// Session T2-01 depends on T1-01
	session := Session{
		ID:           "T2-01",
		Title:        "Database Deep Dive",
		Dependencies: []string{"T1-01"},
		Prompt:       "# Deep Dive\n\nBased on upstream findings:\n\n[UPSTREAM_FINDINGS]\n\nInvestigate further.",
	}

	completed := map[string]bool{"T1-01": true}
	injected, err := InjectContext(session, sessionsDir, completed)
	if err != nil {
		t.Fatalf("InjectContext: %v", err)
	}

	// Verify upstream findings were injected
	if !strings.Contains(injected.InjectedPrompt, "PostgreSQL") {
		t.Error("injected prompt should contain PostgreSQL from upstream T1-01")
	}

	// Verify the placeholder was replaced
	if strings.Contains(injected.InjectedPrompt, "[UPSTREAM_FINDINGS]") {
		t.Error("placeholder [UPSTREAM_FINDINGS] should have been replaced")
	}

	// Verify metadata
	if len(injected.UpstreamSessions) != 1 || injected.UpstreamSessions[0] != "T1-01" {
		t.Errorf("expected upstream sessions [T1-01], got %v", injected.UpstreamSessions)
	}

	if injected.InjectedBytes == 0 {
		t.Error("injected bytes should be > 0")
	}
}

func TestInjectContext_Synthesis(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	os.MkdirAll(sessionsDir, 0o755)

	// Write two completed sessions
	os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(`---
session_id: T1-01
title: Database Selection
date: 2026-09-27
---

## Recommendations
Use PostgreSQL. A (docs)
`), 0o644)

	os.WriteFile(filepath.Join(sessionsDir, "T1-02.md"), []byte(`---
session_id: T1-02
title: Auth Strategy
date: 2026-09-27
---

## Findings
Use Clerk for auth. B (comparison)
`), 0o644)

	// SYN-01 session
	session := Session{
		ID:           "SYN-01",
		Title:        "Synthesis",
		Dependencies: []string{"T1-01", "T1-02"},
		Prompt:       "# Synthesis\n\nAggregate all findings:\n\n[ALL_SESSION_FINDINGS]\n\nProduce FAD.",
	}

	completed := map[string]bool{"T1-01": true, "T1-02": true}
	injected, err := InjectContext(session, sessionsDir, completed)
	if err != nil {
		t.Fatalf("InjectContext: %v", err)
	}

	// Should contain findings from BOTH sessions
	if !strings.Contains(injected.InjectedPrompt, "PostgreSQL") {
		t.Error("synthesis should contain T1-01 findings")
	}
	if !strings.Contains(injected.InjectedPrompt, "Clerk") {
		t.Error("synthesis should contain T1-02 findings")
	}

	// Should have both upstream sessions
	if len(injected.UpstreamSessions) != 2 {
		t.Errorf("expected 2 upstream sessions, got %d", len(injected.UpstreamSessions))
	}
}

func TestInjectContext_NoSlot(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	os.MkdirAll(sessionsDir, 0o755)

	os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(`---
session_id: T1-01
title: Test
date: 2026-09-27
---

## Summary
Key finding here.
`), 0o644)

	// Session WITHOUT any placeholder slot
	session := Session{
		ID:           "T2-01",
		Dependencies: []string{"T1-01"},
		Prompt:       "# Deep Dive\n\nInvestigate PostgreSQL extensions.",
	}

	completed := map[string]bool{"T1-01": true}
	injected, err := InjectContext(session, sessionsDir, completed)
	if err != nil {
		t.Fatalf("InjectContext: %v", err)
	}

	// Should append upstream context even without slot
	if !strings.Contains(injected.InjectedPrompt, "UPSTREAM RESEARCH CONTEXT") {
		t.Error("should append upstream context when no slot found")
	}
}

func TestInjectContext_NoDependencies(t *testing.T) {
	session := Session{
		ID:     "T1-01",
		Prompt: "# Research Brief\n\nInvestigate options.",
	}

	injected, err := InjectContext(session, "/nonexistent", nil)
	if err != nil {
		t.Fatalf("InjectContext: %v", err)
	}

	// Should return prompt unchanged
	if injected.InjectedPrompt != session.Prompt {
		t.Error("prompt should be unchanged when no dependencies")
	}
}

func TestInjectContext_SingleInjection(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	os.MkdirAll(sessionsDir, 0o755)

	os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(`---
session_id: T1-01
title: Database Selection
date: 2026-09-27
---

## Findings
PostgreSQL handles graph traversal up to 50k nodes at <50ms p95
`), 0o644)

	// Session WITH [ALL_SESSION_FINDINGS]
	session := Session{
		ID:           "SYN-01",
		Dependencies: []string{"T1-01"},
		Prompt:       "# Synthesis\n\n[ALL_SESSION_FINDINGS]\n\nAnd some more [ALL_SESSION_FINDINGS]",
	}

	completed := map[string]bool{"T1-01": true}
	injected, err := InjectContext(session, sessionsDir, completed)
	if err != nil {
		t.Fatalf("InjectContext: %v", err)
	}

	// Count occurrences of the injected text
	count := strings.Count(injected.InjectedPrompt, "PostgreSQL handles graph")
	if count != 1 {
		t.Errorf("expected exactly 1 injection of findings, got %d", count)
	}

	// Session with BOTH slots
	sessionBoth := Session{
		ID:           "SYN-02",
		Dependencies: []string{"T1-01"},
		Prompt:       "# Synthesis\n\n[ALL_SESSION_FINDINGS]\n\n[UPSTREAM_FINDINGS]",
	}

	injectedBoth, err := InjectContext(sessionBoth, sessionsDir, completed)
	if err != nil {
		t.Fatalf("InjectContext: %v", err)
	}

	countBoth := strings.Count(injectedBoth.InjectedPrompt, "PostgreSQL handles graph")
	if countBoth != 1 {
		t.Errorf("expected exactly 1 injection of findings when both slots present, got %d", countBoth)
	}
}
