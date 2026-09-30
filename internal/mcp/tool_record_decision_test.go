package mcputil

import (
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

