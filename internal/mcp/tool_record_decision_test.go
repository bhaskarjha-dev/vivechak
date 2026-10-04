package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
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

	t.Run("replace unanchored decision in place (F-07)", func(t *testing.T) {
		existing := []byte(`# Architectural Decisions

---
id: D-001
title: Old Database Title
status: proposed
---

# D-001: Old Database Title
Initial hypothesis content.

---
id: D-002
title: Auth Strategy
status: proposed
---
`)
		res := updateOrAppendDecision(existing, "D-001", `---
id: D-001
title: New Database Title
status: accepted
---

# D-001: New Database Title
Accepted content.`)
		s := string(res)
		if strings.Contains(s, "Old Database Title") {
			t.Errorf("expected old unanchored content to be replaced, got: %s", s)
		}
		if !strings.Contains(s, "<!-- DECISION: D-001 -->") {
			t.Errorf("expected new anchored content, got: %s", s)
		}
		if !strings.Contains(s, "New Database Title") {
			t.Errorf("expected updated title, got: %s", s)
		}
		if !strings.Contains(s, "id: D-002") {
			t.Errorf("lost D-002 during update, got: %s", s)
		}
	})

	t.Run("replace unanchored D-002 preserving earlier D-001 [CRIT-01]", func(t *testing.T) {
		existing := []byte(`# Architectural Decisions

---
id: D-001
title: Database Selection
status: accepted
---

# D-001: Database Selection
PostgreSQL was chosen.

---
id: D-002
title: Old Auth Strategy
status: proposed
---

# D-002: Old Auth Strategy
Initial auth proposal.
`)
		res := updateOrAppendDecision(existing, "D-002", `---
id: D-002
title: New Auth Strategy
status: accepted
---

# D-002: New Auth Strategy
JWT and OAuth2 chosen.`)
		s := string(res)
		if !strings.Contains(s, "Database Selection") || !strings.Contains(s, "PostgreSQL was chosen.") {
			t.Errorf("CRIT-01 regression: wiped earlier D-001! got: %s", s)
		}
		if strings.Contains(s, "Old Auth Strategy") {
			t.Errorf("expected old D-002 content to be replaced, got: %s", s)
		}
		if !strings.Contains(s, "New Auth Strategy") || !strings.Contains(s, "JWT and OAuth2 chosen.") {
			t.Errorf("missing updated D-002, got: %s", s)
		}
	})

	t.Run("preserve internal thematic breaks (---) in decision body", func(t *testing.T) {
		existing := []byte(`# Architectural Decisions

---
id: D-001
title: Pipeline Design
status: proposed
---

# D-001: Pipeline Design
Section 1

---

Section 2 with internal horizontal rule

---
id: D-002
title: Next Decision
status: proposed
---

# D-002: Next Decision
Content of D-002
`)
		res := updateOrAppendDecision(existing, "D-001", `---
id: D-001
title: Updated Pipeline Design
status: accepted
---

# D-001: Updated Pipeline Design
Updated Section 1

---

Updated Section 2 with rule`)
		s := string(res)
		if !strings.Contains(s, "Updated Section 2 with rule") {
			t.Errorf("missing updated content, got: %s", s)
		}
		if !strings.Contains(s, "Next Decision") || !strings.Contains(s, "Content of D-002") {
			t.Errorf("lost D-002 due to internal horizontal rule, got: %s", s)
		}
	})

	t.Run("preserve subsequent decision with comments after delimiter [CRIT-01]", func(t *testing.T) {
		existing := []byte(`# Architectural Decisions

---
id: D-001
title: Old First Decision
status: proposed
---

# D-001: Old First Decision
Body of D-001.

---
# Architectural Decision Record Template — Vivechak v0.1.0
# Usage: Add one entry per decision to your project's DECISIONS.md registry file.
id: D-002
title: Second Decision With Comments
status: proposed
---

# D-002: Second Decision With Comments
Body of D-002.
`)
		res := updateOrAppendDecision(existing, "D-001", `---
id: D-001
title: New First Decision
status: accepted
---

# D-001: New First Decision
Updated Body of D-001.`)
		s := string(res)
		if strings.Contains(s, "Old First Decision") {
			t.Errorf("expected old D-001 to be replaced, got: %s", s)
		}
		if !strings.Contains(s, "New First Decision") {
			t.Errorf("missing updated D-001, got: %s", s)
		}
		if !strings.Contains(s, "D-002") || !strings.Contains(s, "Second Decision With Comments") || !strings.Contains(s, "Body of D-002.") {
			t.Fatalf("CRIT-01 regression: wiped subsequent D-002 with comments! got:\n%s", s)
		}
	})
}

func TestResolveDecisionFilename(t *testing.T) {
	tmpDir := t.TempDir()
	researchDir := filepath.Join(tmpDir, core.ResearchDir)
	if err := os.MkdirAll(researchDir, 0o755); err != nil {
		t.Fatalf("failed to create research dir: %v", err)
	}

	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("failed to open workspace: %v", err)
	}
	defer ws.Close()

	// 1. Explicit slug provided
	got := resolveDecisionFilename(ws, "D-001", "custom-database")
	if got != "D-001-custom-database.md" {
		t.Errorf("expected D-001-custom-database.md, got %q", got)
	}

	// 2. No slug, no existing file -> fallback to [ID]-decision.md
	gotFallback := resolveDecisionFilename(ws, "D-002", "")
	if gotFallback != "D-002-decision.md" {
		t.Errorf("expected D-002-decision.md, got %q", gotFallback)
	}

	// 3. No slug, but existing file D-002-auth-strategy.md exists -> matches existing name
	existingADR := filepath.Join(researchDir, "D-002-auth-strategy.md")
	if err := os.WriteFile(existingADR, []byte("adr"), 0o644); err != nil {
		t.Fatalf("failed to write adr: %v", err)
	}

	gotMatched := resolveDecisionFilename(ws, "D-002", "")
	if gotMatched != "D-002-auth-strategy.md" {
		t.Errorf("expected D-002-auth-strategy.md, got %q", gotMatched)
	}

	// 4. Ignore plan files like D-002-plan.md
	existingPlan := filepath.Join(researchDir, "D-003-plan.md")
	if err := os.WriteFile(existingPlan, []byte("plan"), 0o644); err != nil {
		t.Fatalf("failed to write plan: %v", err)
	}
	gotPlanIgnored := resolveDecisionFilename(ws, "D-003", "")
	if gotPlanIgnored != "D-003-decision.md" {
		t.Errorf("expected D-003-decision.md, got %q", gotPlanIgnored)
	}
}

func TestCompileDecisionsRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	researchDir := filepath.Join(tmpDir, core.ResearchDir)
	if err := os.MkdirAll(researchDir, 0o755); err != nil {
		t.Fatalf("failed to create research dir: %v", err)
	}

	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("failed to open workspace: %v", err)
	}
	defer ws.Close()

	// Write 3 ADR files out of order
	d002Content := `---
id: D-002
title: Auth Strategy
status: accepted
door_type: two-way
---
# D-002
Use Clerk for auth.
`
	d001Content := `---
id: D-001
title: Datastore Selection
status: accepted
door_type: one-way
review_trigger: "throughput > 50k"
---
# D-001
Use PostgreSQL.
`
	d003Content := `---
id: D-003
title: Hosting Architecture
status: accepted
door_type: two-way
---
# D-003
Deploy on Fly.io.
`
	// Also write non-ADR files that should be ignored
	planContent := `---
id: D-001-plan
title: Decision Plan
---
# Plan
`
	notesContent := "# Notes\nSome notes.\n"

	_ = os.WriteFile(filepath.Join(researchDir, "D-002-auth.md"), []byte(d002Content), 0o644)
	_ = os.WriteFile(filepath.Join(researchDir, "D-001-datastore.md"), []byte(d001Content), 0o644)
	_ = os.WriteFile(filepath.Join(researchDir, "D-003-hosting.md"), []byte(d003Content), 0o644)
	_ = os.WriteFile(filepath.Join(researchDir, "D-001-plan.md"), []byte(planContent), 0o644)
	_ = os.WriteFile(filepath.Join(researchDir, "FOUNDING-ARCHITECTURE.md"), []byte("---\nstatus: sealed\n---\n# FAD"), 0o644)
	_ = os.WriteFile(filepath.Join(researchDir, "NOTES.md"), []byte(notesContent), 0o644)

	ctx := t.Context()
	if err := compileDecisionsRegistry(ctx, ws, tmpDir); err != nil {
		t.Fatalf("compileDecisionsRegistry failed: %v", err)
	}

	decisionsData, err := os.ReadFile(filepath.Join(tmpDir, core.DecisionsFile))
	if err != nil {
		t.Fatalf("reading DECISIONS.md: %v", err)
	}
	decisionsStr := string(decisionsData)

	// Verify all 3 decisions are present in sorted order
	idx1 := strings.Index(decisionsStr, "<!-- DECISION: D-001 -->")
	idx2 := strings.Index(decisionsStr, "<!-- DECISION: D-002 -->")
	idx3 := strings.Index(decisionsStr, "<!-- DECISION: D-003 -->")

	if idx1 == -1 || idx2 == -1 || idx3 == -1 {
		t.Fatalf("missing one or more decision markers: D-001=%d, D-002=%d, D-003=%d", idx1, idx2, idx3)
	}

	if idx1 >= idx2 || idx2 >= idx3 {
		t.Errorf("decisions are not in sorted order: D-001=%d, D-002=%d, D-003=%d", idx1, idx2, idx3)
	}

	// Verify <details> formatting
	if !strings.Contains(decisionsStr, "<details>") || !strings.Contains(decisionsStr, "</details>") {
		t.Errorf("expected <details> blocks in DECISIONS.md, got:\n%s", decisionsStr)
	}

	// Verify non-ADRs are excluded
	if strings.Contains(decisionsStr, "D-001-plan") || strings.Contains(decisionsStr, "Decision Plan") {
		t.Errorf("plan file was mistakenly compiled into DECISIONS.md")
	}
	if strings.Contains(decisionsStr, "status: sealed") || strings.Contains(decisionsStr, "FOUNDING-ARCHITECTURE") {
		t.Errorf("FAD was mistakenly compiled into DECISIONS.md")
	}
	if strings.Contains(decisionsStr, "Some notes") {
		t.Errorf("NOTES.md was mistakenly compiled into DECISIONS.md")
	}
}

func TestCompileDecisionsRegistry_UnanchoredAndNaturalSort(t *testing.T) {
	tmpDir := t.TempDir()
	researchDir := filepath.Join(tmpDir, "research")
	if err := os.MkdirAll(researchDir, 0o755); err != nil {
		t.Fatalf("creating research dir: %v", err)
	}

	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("opening workspace: %v", err)
	}
	defer ws.Close()

	// 1. Write DECISIONS.md containing UNANCHORED frontmatter blocks with out-of-order IDs:
	// D-1, D-10, D-2
	decisionsPreamble := `# Architecture Decisions

Registry of decisions made during research.

---
id: D-1
title: Language Selection
status: proposed
door_type: one-way
---
# D-1: Language Selection
We chose Go.

---
id: D-10
title: Metrics Framework
status: proposed
door_type: two-way
---
# D-10: Metrics
We chose Prometheus.

---
id: D-2
title: Datastore Selection
status: proposed
door_type: one-way
---
# D-2: Datastore
We chose SQLite.
`
	if err := os.WriteFile(filepath.Join(researchDir, "DECISIONS.md"), []byte(decisionsPreamble), 0o644); err != nil {
		t.Fatalf("writing DECISIONS.md: %v", err)
	}

	// 2. Also write an individual ADR file in research/ D-3.md
	d3Content := `---
id: D-3
title: API Protocol
status: accepted
door_type: one-way
review_trigger: "protocol deprecation"
---
# D-3: API Protocol
We chose JSON-RPC over MCP.
`
	if err := os.WriteFile(filepath.Join(researchDir, "D-3.md"), []byte(d3Content), 0o644); err != nil {
		t.Fatalf("writing D-3.md: %v", err)
	}

	// 3. Compile registry
	ctx := context.Background()
	if err := compileDecisionsRegistry(ctx, ws, tmpDir); err != nil {
		t.Fatalf("compileDecisionsRegistry failed: %v", err)
	}

	// 4. Verify DECISIONS.md has all 4 decisions wrapped in anchors and sorted naturally:
	// D-1 -> D-2 -> D-3 -> D-10 (natural numeric sorting!)
	data, err := os.ReadFile(filepath.Join(researchDir, "DECISIONS.md"))
	if err != nil {
		t.Fatalf("reading compiled DECISIONS.md: %v", err)
	}
	content := string(data)

	idx1 := strings.Index(content, "<!-- DECISION: D-1 -->")
	idx2 := strings.Index(content, "<!-- DECISION: D-2 -->")
	idx3 := strings.Index(content, "<!-- DECISION: D-3 -->")
	idx10 := strings.Index(content, "<!-- DECISION: D-10 -->")

	if idx1 == -1 || idx2 == -1 || idx3 == -1 || idx10 == -1 {
		t.Fatalf("missing anchor in compiled registry: D-1=%d, D-2=%d, D-3=%d, D-10=%d\ncontent:\n%s", idx1, idx2, idx3, idx10, content)
	}

	// Natural sort check: D-1 < D-2 < D-3 < D-10
	if idx1 >= idx2 || idx2 >= idx3 || idx3 >= idx10 {
		t.Errorf("expected natural order D-1 < D-2 < D-3 < D-10; got indices: D-1=%d, D-2=%d, D-3=%d, D-10=%d", idx1, idx2, idx3, idx10)
	}
}

func TestRecordDecision_Supersede(t *testing.T) {
	tmpDir := t.TempDir()
	researchDir := filepath.Join(tmpDir, core.ResearchDir)
	sessionsDir := filepath.Join(tmpDir, core.SessionsDir)
	if err := os.MkdirAll(researchDir, 0o755); err != nil {
		t.Fatalf("mkdir research: %v", err)
	}
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}

	// 1. Create original decision D-001
	d1Content := `---
id: D-001
title: Old Database Engine
status: accepted
door_type: two-way
---
# D-001: Old Database Engine
We chose SQLite with WAL mode. Sufficient body text to satisfy validation length requirements.
`
	if err := os.WriteFile(filepath.Join(researchDir, "D-001-database.md"), []byte(d1Content), 0o644); err != nil {
		t.Fatalf("writing D-001: %v", err)
	}

	// 2. Create a session referencing D-001
	s1Content := `---
session_id: T-01
title: Auth and DB Investigation
date: 2026-10-01
status: complete
---
## Key Findings
- Evaluated against D-001 architecture (Grade A · direct docs | official docs).
`
	if err := os.WriteFile(filepath.Join(sessionsDir, "T-01.md"), []byte(s1Content), 0o644); err != nil {
		t.Fatalf("writing session T-01: %v", err)
	}

	// 3. Test self-supersede error
	ctx := context.Background()
	_, envSelf, _ := handleRecordDecision(ctx, nil, RecordDecisionInput{
		ProjectRoot: tmpDir,
		DecisionID:  "D-001",
		Supersedes:  "D-001",
		Content:     d1Content,
	})
	if envSelf.Success {
		t.Errorf("expected failure when superseding itself")
	}

	// 4. Test supersede non-existent decision error
	_, envNonExistent, _ := handleRecordDecision(ctx, nil, RecordDecisionInput{
		ProjectRoot: tmpDir,
		DecisionID:  "D-002",
		Supersedes:  "D-999",
		Content:     d1Content,
	})
	if envNonExistent.Success {
		t.Errorf("expected failure when superseding non-existent decision")
	}

	// 5. Successfully supersede D-001 with D-002
	d2Content := `---
id: D-002
title: Embedded DuckDB Migration
status: accepted
door_type: two-way
---
# D-002: Embedded DuckDB Migration
Migrated away from SQLite to DuckDB for analytical query performance and embedded vector extension capabilities.
`
	_, envSuccess, err := handleRecordDecision(ctx, nil, RecordDecisionInput{
		ProjectRoot: tmpDir,
		DecisionID:  "D-002",
		Supersedes:  "D-001",
		Content:     d2Content,
	})
	if err != nil || !envSuccess.Success {
		t.Fatalf("handleRecordDecision supersede failed: err=%v, env=%+v", err, envSuccess)
	}

	// Verify old decision was marked superseded
	oldData, err := os.ReadFile(filepath.Join(researchDir, "D-001-database.md"))
	if err != nil {
		t.Fatalf("reading old decision: %v", err)
	}
	oldStr := string(oldData)
	if !strings.Contains(oldStr, "status: superseded") {
		t.Errorf("expected old decision status to be superseded, got:\n%s", oldStr)
	}
	if !strings.Contains(oldStr, "superseded_by: D-002") {
		t.Errorf("expected old decision superseded_by: D-002, got:\n%s", oldStr)
	}

	// Verify new decision has supersedes: D-001 in file
	newData, err := os.ReadFile(filepath.Join(researchDir, "D-002-decision.md"))
	if err != nil {
		t.Fatalf("reading new decision: %v", err)
	}
	newStr := string(newData)
	if !strings.Contains(newStr, "supersedes: D-001") {
		t.Errorf("expected new decision supersedes: D-001, got:\n%s", newStr)
	}

	// Verify warnings contain W-STALE-DECISION-REFERENCE for session T-01
	foundWarning := false
	for _, w := range envSuccess.Warnings {
		if strings.Contains(w, "W-STALE-DECISION-REFERENCE") && strings.Contains(w, "T-01") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Errorf("expected W-STALE-DECISION-REFERENCE for session T-01, got warnings: %v", envSuccess.Warnings)
	}

	// Verify DECISIONS.md registry contains both decisions
	decData, err := os.ReadFile(filepath.Join(researchDir, "DECISIONS.md"))
	if err != nil {
		t.Fatalf("reading DECISIONS.md: %v", err)
	}
	decStr := string(decData)
	if !strings.Contains(decStr, "<!-- DECISION: D-001 -->") || !strings.Contains(decStr, "<!-- DECISION: D-002 -->") {
		t.Errorf("expected both decisions in DECISIONS.md, got:\n%s", decStr)
	}
}



