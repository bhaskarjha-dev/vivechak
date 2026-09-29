package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectContext_Dependencies(t *testing.T) {
	// Create temp dir with session files
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	_ = os.MkdirAll(sessionsDir, 0o755)

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
	_ = os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(t101Content), 0o644)

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
	_ = os.MkdirAll(sessionsDir, 0o755)

	// Write two completed sessions
	_ = os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(`---
session_id: T1-01
title: Database Selection
date: 2026-09-27
---

## Recommendations
Use PostgreSQL. A (docs)
`), 0o644)

	_ = os.WriteFile(filepath.Join(sessionsDir, "T1-02.md"), []byte(`---
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
	_ = os.MkdirAll(sessionsDir, 0o755)

	_ = os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(`---
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
	_ = os.MkdirAll(sessionsDir, 0o755)

	_ = os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(`---
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

func TestReadSessionFile_Comprehensive(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. File with canonical ID and hyphenated title slug
	_ = os.WriteFile(filepath.Join(tmpDir, "T1-01-database-selection.md"), []byte("---\nsession_id: T1-01\ntitle: Database Selection\n---\n# Database Selection\n"), 0o644)
	// 2. File with canonical ID and no slug (exact match)
	_ = os.WriteFile(filepath.Join(tmpDir, "T1-02.md"), []byte("---\nsession_id: T1-02\ntitle: Auth\n---\n# Auth\n"), 0o644)
	// 3. File with no frontmatter at all and numeric sub-session stem
	_ = os.WriteFile(filepath.Join(tmpDir, "T1-03.md"), []byte("# Plain Markdown with no frontmatter\n"), 0o644)
	// 4. File whose filename prefix matches T1-04, but frontmatter explicitly declares T1-99
	_ = os.WriteFile(filepath.Join(tmpDir, "T1-04-old-name.md"), []byte("---\nsession_id: T1-99\ntitle: Renamed Session\n---\n# Stale name\n"), 0o644)
	// 5. File with underscore separator
	_ = os.WriteFile(filepath.Join(tmpDir, "T2-01_caching_layer.md"), []byte("---\nsession_id: T2-01\ntitle: Caching\n---\n# Caching\n"), 0o644)
	// 6. Sub-session of a decision (D-001 vs D-001-S1)
	_ = os.WriteFile(filepath.Join(tmpDir, "D-001-S1-investigation.md"), []byte("---\nsession_id: D-001-S1\ntitle: Investigation\n---\n# Investigation\n"), 0o644)

	t.Run("canonical ID matches file with title slug", func(t *testing.T) {
		content, filename, err := readSessionFile(tmpDir, "T1-01")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(content, "Database Selection") {
			t.Errorf("expected content from T1-01-database-selection.md, got %q", content)
		}
		if filename != "T1-01-database-selection.md" {
			t.Errorf("expected filename T1-01-database-selection.md, got %q", filename)
		}
	})

	t.Run("canonical ID matches exact filename", func(t *testing.T) {
		content, filename, err := readSessionFile(tmpDir, "T1-02")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(content, "Auth") {
			t.Errorf("expected content from T1-02.md, got %q", content)
		}
		if filename != "T1-02.md" {
			t.Errorf("expected filename T1-02.md, got %q", filename)
		}
	})

	t.Run("canonical ID matches underscore slug", func(t *testing.T) {
		content, filename, err := readSessionFile(tmpDir, "T2-01")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(content, "Caching") {
			t.Errorf("expected content from T2-01_caching_layer.md, got %q", content)
		}
		if filename != "T2-01_caching_layer.md" {
			t.Errorf("expected filename T2-01_caching_layer.md, got %q", filename)
		}
	})

	t.Run("sub-session ID matches sub-session file", func(t *testing.T) {
		content, filename, err := readSessionFile(tmpDir, "D-001-S1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(content, "Investigation") {
			t.Errorf("expected content from D-001-S1-investigation.md, got %q", content)
		}
		if filename != "D-001-S1-investigation.md" {
			t.Errorf("expected filename D-001-S1-investigation.md, got %q", filename)
		}
	})

	t.Run("parent decision ID does NOT falsely match sub-session file", func(t *testing.T) {
		// D-001 should NOT match D-001-S1-investigation.md because frontmatter is D-001-S1
		content, _, err := readSessionFile(tmpDir, "D-001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if content != "" {
			t.Errorf("D-001 must not match D-001-S1, got %q", content)
		}
	})

	t.Run("prefix substring T1 does NOT match compound IDs T1-01, T1-02, or T1-03", func(t *testing.T) {
		// T1 is a non-canonical prefix of T1-01, T1-02, T1-03.
		// It must never match any of them.
		content, _, err := readSessionFile(tmpDir, "T1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if content != "" {
			t.Errorf("T1 must not match any T1-XX files, got %q", content)
		}
	})

	t.Run("frontmatter authority rejects file with mismatched session_id", func(t *testing.T) {
		// T1-04-old-name.md has session_id: T1-99 in frontmatter.
		// Querying T1-04 must NOT return it because the file declared itself to be T1-99.
		content, _, err := readSessionFile(tmpDir, "T1-04")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if content != "" {
			t.Errorf("T1-04 must not match file declaring session_id T1-99, got %q", content)
		}

		// But querying T1-99 via prefix should NOT match T1-04-old-name because filename doesn't start with T1-99.
		// This protects integrity from both sides.
		content99, _, err := readSessionFile(tmpDir, "T1-99")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if content99 != "" {
			t.Errorf("T1-99 should not match T1-04-old-name without filename match, got %q", content99)
		}
	})

	t.Run("nonexistent session returns empty content without error", func(t *testing.T) {
		content, filename, err := readSessionFile(tmpDir, "NONEXISTENT-99")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if content != "" || filename != "" {
			t.Errorf("expected empty result, got content=%q filename=%q", content, filename)
		}
	})
}

func TestInjectContext_SizeWarning(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	_ = os.MkdirAll(sessionsDir, 0o755)

	// Create a large session output (>100KB)
	largeBody := strings.Repeat("Key finding with evidence: PostgreSQL scales linearly. A (benchmark)\n", 2000)
	sessionContent := "---\nsession_id: T1-01\ntitle: Heavy Session\ndate: 2026-09-29\nstatus: complete\n---\n# Findings\n" + largeBody
	_ = os.WriteFile(filepath.Join(sessionsDir, "T1-01.md"), []byte(sessionContent), 0o644)

	session := Session{
		ID:           "SYN-01",
		Title:        "Synthesis",
		Dependencies: []string{"T1-01"},
		Prompt:       "# Synthesis\n[ALL_SESSION_FINDINGS]",
	}

	completed := map[string]bool{"T1-01": true}
	injected, err := InjectContext(session, sessionsDir, completed)
	if err != nil {
		t.Fatalf("InjectContext: %v", err)
	}

	if injected.InjectedBytes <= MaxInjectedBytesWarning {
		t.Fatalf("expected injected bytes > %d, got %d", MaxInjectedBytesWarning, injected.InjectedBytes)
	}

	if injected.Warning == "" {
		t.Error("expected size warning when injected bytes exceed threshold, got empty string")
	}

	if !strings.Contains(injected.Warning, "W-INJECTION-SIZE") {
		t.Errorf("expected warning to contain 'W-INJECTION-SIZE', got %q", injected.Warning)
	}
}

func TestInjectContext_DecisionShortlistSlot(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	_ = os.MkdirAll(sessionsDir, 0o755)

	s1Content := `---
session_id: D-015-S1
title: Landscape
status: complete
---
# Landscape Findings
Shortlist: Redis, Dragonfly, KeyDB. A (official docs)
`
	_ = os.WriteFile(filepath.Join(sessionsDir, "D-015-S1.md"), []byte(s1Content), 0o644)

	session := Session{
		ID:           "D-015-S2",
		Title:        "Comparison",
		Dependencies: []string{"D-015-S1"},
		Prompt:       "# Comparison\n\nOptions to compare: [PASTE S1 SHORTLIST]\n\nEvaluate latency.",
	}

	completed := map[string]bool{"D-015-S1": true}
	injected, err := InjectContext(session, sessionsDir, completed)
	if err != nil {
		t.Fatalf("InjectContext: %v", err)
	}

	if strings.Contains(injected.InjectedPrompt, "[PASTE S1 SHORTLIST]") {
		t.Errorf("expected [PASTE S1 SHORTLIST] to be replaced, got: %s", injected.InjectedPrompt)
	}
	if !strings.Contains(injected.InjectedPrompt, "Redis, Dragonfly, KeyDB") {
		t.Errorf("expected shortlist findings in prompt, got: %s", injected.InjectedPrompt)
	}
}

func TestExtractFindings_Comprehensive(t *testing.T) {
	t.Run("headings exceeding 200 bytes do not starve raw body fallback (Fix H-01)", func(t *testing.T) {
		// Create a body with several lengthy headings that don't match relevant section keywords,
		// and technical body content below that lacks formal evidence grades.
		var sb strings.Builder
		for i := 1; i <= 8; i++ {
			fmt.Fprintf(&sb, "## Architectural Constraint Analysis Chapter %d: Operational Considerations\n", i)
		}
		sb.WriteString("Core finding: PostgreSQL with pgvector scales to 1M embeddings at 12ms latency.\n")
		sb.WriteString("Cluster deployment requires 3 nodes with 64GB RAM.\n")

		body := sb.String()
		extracted := extractFindings(body, "T1-01")

		// Must contain the actual technical findings, not just empty headings
		if !strings.Contains(extracted, "PostgreSQL with pgvector scales") {
			t.Errorf("extractFindings starved downstream prompt of technical content: %s", extracted)
		}
		if !strings.Contains(extracted, "Cluster deployment requires") {
			t.Errorf("extractFindings missing secondary technical content: %s", extracted)
		}
	})

	t.Run("extracts numeric evidence citations [E-NNN] and E-NNN (Fix X-01)", func(t *testing.T) {
		body := `# General Notes
Some introductory text.
Throughput exceeds 45,000 requests/sec under load [E-01].
Secondary latency target satisfied by Redis cluster E-002.
`
		extracted := extractFindings(body, "T1-02")
		if !strings.Contains(extracted, "[E-01]") {
			t.Errorf("expected numeric citation [E-01] to be extracted, got: %s", extracted)
		}
		if !strings.Contains(extracted, "E-002") {
			t.Errorf("expected numeric citation E-002 to be extracted, got: %s", extracted)
		}
	})

	t.Run("excludes rejected alternatives sections", func(t *testing.T) {
		body := `# Findings
## Recommendations
We recommend SQLite in WAL mode. A (official docs)

## Rejected Alternatives & Tradeoffs
We rejected RocksDB because Cgo compilation overhead is too high. A (benchmark)

## Alternatives Considered
DuckDB was ruled out due to concurrent write locking. B (docs)
`
		extracted := extractFindings(body, "T1-03")
		if !strings.Contains(extracted, "SQLite in WAL mode") {
			t.Errorf("expected SQLite recommendation to be present, got: %s", extracted)
		}
		if strings.Contains(extracted, "RocksDB because Cgo") {
			t.Errorf("expected rejected alternative RocksDB to be excluded, got: %s", extracted)
		}
		if strings.Contains(extracted, "DuckDB was ruled out") {
			t.Errorf("expected alternative considered DuckDB to be excluded, got: %s", extracted)
		}
	})

	t.Run("buffers headings so only headings with substantive content appear", func(t *testing.T) {
		body := `# Session Overview
## Unused Section With No Findings
## Section With Substantive Findings
Key recommendation: use Kafka for event sourcing. A (docs)
`
		extracted := extractFindings(body, "T1-04")
		if !strings.Contains(extracted, "Kafka for event sourcing") {
			t.Errorf("expected Kafka recommendation, got: %s", extracted)
		}
		if strings.Contains(extracted, "Unused Section With No Findings") {
			t.Errorf("expected empty heading to be suppressed, got: %s", extracted)
		}
	})

	t.Run("extracts Recommended Architecture and Shortlist headings without formal grades (F-04 and F-08)", func(t *testing.T) {
		body := `# Landscape & Architecture
## Recommended Architecture
Adopt distributed event logging with NATS JetStream.

### Shortlist of Candidates
1. NATS JetStream
2. Apache Kafka
3. RabbitMQ
`
		extracted := extractFindings(body, "T1-05")
		if !strings.Contains(extracted, "Adopt distributed event logging with NATS JetStream") {
			t.Errorf("expected Recommended Architecture to be extracted, got: %s", extracted)
		}
		if !strings.Contains(extracted, "1. NATS JetStream") {
			t.Errorf("expected shortlist candidate to be extracted, got: %s", extracted)
		}
	})

	t.Run("falls back to raw body if only stray grade matched and no relevant headings found (F-04)", func(t *testing.T) {
		body := `# Architectural Overview
Historical benchmark achieved 10k ops [Grade B].

System implementation requires distributed consensus with Raft and 3-node cluster.
Storage engine will use LSM trees for write throughput.
`
		extracted := extractFindings(body, "T1-06")
		// Must not starve the prompt of the rest of the text
		if !strings.Contains(extracted, "System implementation requires distributed consensus") {
			t.Errorf("expected raw body fallback when only minimal stray grade matched, got: %s", extracted)
		}
	})

	t.Run("code block comments with # do not trigger rejected heading logic", func(t *testing.T) {
		body := `## Recommended Architecture
Adopt custom microservice layout.

` + "```bash\n# rejected alternative: port 8080\nexport PORT=9090\n```\n" + `
This continues the recommended section and must be extracted.
`
		extracted := extractFindings(body, "T1-07")
		if !strings.Contains(extracted, "This continues the recommended section and must be extracted") {
			t.Errorf("expected content after code block to be retained, got: %s", extracted)
		}
	})
}
