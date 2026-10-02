package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func TestCheckWorkspace_Nonexistent(t *testing.T) {
	hasErrors, _, errs := checkWorkspace(filepath.Join(t.TempDir(), "nonexistent"))
	if !hasErrors {
		t.Error("expected hasErrors=true for nonexistent workspace")
	}
	if len(errs) == 0 {
		t.Error("expected error message for nonexistent workspace")
	}
}

func TestCheckWorkspace_MissingTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, core.ResearchDir), 0o755)

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if !hasErrors {
		t.Error("expected hasErrors=true for missing templates")
	}
	found := false
	for _, e := range errs {
		if len(e) > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected missing templates error")
	}
}

func TestCheckWorkspace_Healthy(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directories
	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)

	// Create templates
	for _, tmpl := range core.TemplatesToCopy {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	// Create pipeline
	pipelineContent := `# Research Pipeline
#### T1-01: Foundation
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(pipelineContent), 0o644)

	// Create decisions
	decisionsContent := `# Decisions
<!-- DECISION: D-001 -->
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte(decisionsContent), 0o644)

	// Create session with valid frontmatter, no orphans, valid D-001 ref
	sessionContent := `---
session_id: T1-01
title: Foundation
status: complete
date: 2026-09-28
---
# Foundation
Refers to D-001. A (doc)
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T1-01-foundation.md"), []byte(sessionContent), 0o644)

	hasErrors, successes, errs := checkWorkspace(tmpDir)
	if hasErrors {
		t.Fatalf("expected healthy workspace, got errors: %v", errs)
	}
	if len(successes) == 0 {
		t.Error("expected successes messages")
	}
}

func TestCheckWorkspace_ScopeAwareTemplates_Decision(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)

	// Write metadata for decision scope
	metaJSON := `{"scope":"decision","created_at":"2026-10-02"}`
	_ = os.WriteFile(filepath.Join(tmpDir, core.MetadataFile), []byte(metaJSON), 0o644)

	// Decision scope only requires 2 templates
	for _, tmpl := range core.TemplatesForScope(core.ScopeDecision) {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if hasErrors {
		t.Fatalf("expected decision workspace to be healthy with only 2 templates, got errors: %v", errs)
	}
}

func TestCheckWorkspace_ScopeAwareTemplates_Comparison(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)

	// Write metadata for comparison scope
	metaJSON := `{"scope":"comparison","created_at":"2026-10-02"}`
	_ = os.WriteFile(filepath.Join(tmpDir, core.MetadataFile), []byte(metaJSON), 0o644)

	// Comparison scope only requires 1 template
	for _, tmpl := range core.TemplatesForScope(core.ScopeComparison) {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if hasErrors {
		t.Fatalf("expected comparison workspace to be healthy with only 1 template, got errors: %v", errs)
	}
}

func TestCheckWorkspace_OrphanSession(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	for _, tmpl := range core.TemplatesToCopy {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	// Pipeline has only T1-01
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte("# Pipeline\n#### T1-01: Foo\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte("# Decisions\n"), 0o644)

	// Session has UNKNOWN-01
	sessionContent := `---
session_id: UNKNOWN-01
title: Unknown
---
# Unknown
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "UNKNOWN-01.md"), []byte(sessionContent), 0o644)

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if !hasErrors {
		t.Error("expected hasErrors=true for orphaned session")
	}
	foundOrphan := false
	for _, e := range errs {
		if len(e) > 0 && (len(e) >= 8) {
			foundOrphan = true
			break
		}
	}
	if !foundOrphan {
		t.Error("expected orphan error message")
	}
}

func TestCheckWorkspace_CorruptedFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	for _, tmpl := range core.TemplatesToCopy {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte("# Pipeline\n#### T1-01: Foo\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte("# Decisions\n"), 0o644)

	// Malformed YAML frontmatter
	malformedContent := `---
session_id: [broken yaml
title: "unclosed
---
# Content
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T1-01.md"), []byte(malformedContent), 0o644)

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if !hasErrors {
		t.Error("expected hasErrors=true for corrupted frontmatter")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "Invalid frontmatter") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Invalid frontmatter' error, got %v", errs)
	}
}

func TestCheckWorkspace_StaleDecisionRef(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	for _, tmpl := range core.TemplatesToCopy {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte("# Pipeline\n#### T1-01: Foo\n"), 0o644)
	// DECISIONS.md only has D-001
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte("# Decisions\n<!-- DECISION: D-001 -->\n"), 0o644)

	// Session references non-existent D-999
	sessionContent := `---
session_id: T1-01
title: Foo
---
# Foo
Refers to D-999 which does not exist in DECISIONS.md. A (source)
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T1-01.md"), []byte(sessionContent), 0o644)

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if !hasErrors {
		t.Error("expected hasErrors=true for stale decision ref")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "Stale decision ref") && strings.Contains(e, "D-999") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Stale decision ref' for D-999, got %v", errs)
	}
}

func TestCheckWorkspace_SessionWithoutFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	for _, tmpl := range core.TemplatesToCopy {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	// Pipeline contains T1-01-database
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte("# Pipeline\n#### T1-01-database: Full name\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte("# Decisions\n"), 0o644)

	// Session without frontmatter named T1-01-database.md
	// With core.ExtractSessionID, id is T1-01-database (not truncated to T1-01 by SplitN)
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T1-01-database.md"), []byte("# Content without frontmatter"), 0o644)

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if hasErrors {
		t.Errorf("expected no error since T1-01-database matches pipeline, got errors: %v", errs)
	}
}

func TestCheckWorkspace_CasualTextDoesNotTriggerStaleRef(t *testing.T) {
	tmpDir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	for _, tmpl := range core.TemplatesToCopy {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte("# Pipeline\n#### T1-01: Foo\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte("# Decisions\n<!-- DECISION: D-001 -->\n"), 0o644)

	// Session casually mentions D-100 tier instances, D-25 connectors, but no decision references
	sessionContent := `---
session_id: T1-01
title: Foo
informs_decisions: [D-001]
---
# Hardware & Deployment
Deploy on AWS D-100 tier servers with D-25 cable connectors. A (spec)
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T1-01.md"), []byte(sessionContent), 0o644)

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if hasErrors {
		t.Errorf("expected no errors for casual D-100/D-25 text, got errors: %v", errs)
	}
}

func TestCheckWorkspace_MultipleDecisionPlans_NoFalseOrphans(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	for _, tmpl := range core.TemplatesToCopy {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	// Create two decision plans
	plan1 := "# Plan 1\n#### S1-01: First Decision Session\n"
	plan2 := "# Plan 2\n#### S2-01: Second Decision Session\n"
	_ = os.WriteFile(filepath.Join(tmpDir, core.ResearchDir, "D-001-plan.md"), []byte(plan1), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.ResearchDir, "D-002-plan.md"), []byte(plan2), 0o644)

	// Create standalone ADRs
	_ = os.WriteFile(filepath.Join(tmpDir, core.ResearchDir, "D-001-decision.md"), []byte("---\nid: D-001\nstatus: accepted\n---\n# D-001"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.ResearchDir, "D-002-decision.md"), []byte("---\nid: D-002\nstatus: accepted\n---\n# D-002"), 0o644)

	// Create sessions corresponding to each plan
	s1 := "---\nsession_id: S1-01\ntitle: S1\nstatus: complete\ndate: 2026-09-29\n---\nRefers to D-001. A (doc)"
	s2 := "---\nsession_id: S2-01\ntitle: S2\nstatus: complete\ndate: 2026-09-29\n---\nRefers to D-002. A (doc)"
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "S1-01.md"), []byte(s1), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "S2-01.md"), []byte(s2), 0o644)

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if hasErrors {
		t.Fatalf("expected healthy workspace with multiple decision plans, got errors: %v", errs)
	}
}

func TestCheckWorkspace_AlphanumericDecisionRef(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	for _, tmpl := range core.TemplatesToCopy {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte("# Pipeline\n#### T1-01: Auth Session\n"), 0o644)

	// Decision registry with alphanumeric IDs
	decisions := `# Decisions
<!-- DECISION: D-AUTH-01 -->
---
id: D-AUTH-01
status: accepted
---
# D-AUTH-01

<!-- DECISION: D-NEW -->
---
id: D-NEW
status: accepted
---
# D-NEW
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte(decisions), 0o644)

	// Session referencing D-AUTH-01 and D-NEW
	sessionContent := `---
session_id: T1-01
title: Auth Session
status: complete
date: 2026-09-29
---
# Auth Findings
Refers to [D-AUTH-01] and decision D-NEW. A (doc)
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T1-01.md"), []byte(sessionContent), 0o644)

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if hasErrors {
		t.Fatalf("expected healthy workspace for alphanumeric ADR refs, got errors: %v", errs)
	}
}

func TestCheckWorkspace_InFlightPlannedDecision(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	for _, tmpl := range core.TemplatesToCopy {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	// Pipeline declares planned decision D-001
	pipeline := `# Research Pipeline
#### T1-01: Database Selection
This session informs D-001 (One-Way Door).
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(pipeline), 0o644)

	// DECISIONS.md is still empty or initial registry (D-001 not locked yet)
	decisions := `# Decisions
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte(decisions), 0o644)

	// Session references D-001 as its target decision
	sessionContent := `---
session_id: T1-01
title: Database Selection
status: complete
date: 2026-09-29
informs_decisions: [D-001]
---
# Database Selection
Recommends PostgreSQL. This informs D-001. A (doc)
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T1-01.md"), []byte(sessionContent), 0o644)

	hasErrors, successes, errs := checkWorkspace(tmpDir)
	if hasErrors {
		t.Fatalf("expected in-flight planned decision not to cause error, got errors: %v", errs)
	}
	foundInFlight := false
	for _, s := range successes {
		if strings.Contains(s, "In-flight planned decision") && strings.Contains(s, "D-001") {
			foundInFlight = true
			break
		}
	}
	if !foundInFlight {
		t.Errorf("expected informational in-flight message in successes: %v", successes)
	}
}

func TestCheckWorkspace_SluggedFilenameMatchingPipelinePrefix(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	for _, tmpl := range core.TemplatesToCopy {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	// Pipeline only lists T1-01
	pipeline := `# Research Pipeline
#### T1-01: Database Selection
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(pipeline), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte("# Decisions\n"), 0o644)

	// Session filename has slug, but no explicit frontmatter session_id
	sessionContent := `# Database Selection Findings
Findings... A (doc)
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T1-01-database-selection.md"), []byte(sessionContent), 0o644)

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if hasErrors {
		t.Fatalf("expected slugged filename matching T1-01 prefix not to be marked orphan, got errors: %v", errs)
	}
}

func TestCheckWorkspace_SingleTokenAndCompoundSluggedFilenames(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	for _, tmpl := range core.TemplatesToCopy {
		_ = os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	// Pipeline contains S1 and D-001-S2
	pipeline := `# Research Pipeline
#### S1: Fast Spike
#### D-001-S2: Deep Research
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(pipeline), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte("# Decisions\n"), 0o644)

	// Single token prefix: S1-spike.md
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "S1-spike.md"), []byte("# Content A (doc)"), 0o644)
	// Compound prefix: D-001-S2-benchmark.md
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "D-001-S2-benchmark.md"), []byte("# Content A (doc)"), 0o644)

	hasErrors, _, errs := checkWorkspace(tmpDir)
	if hasErrors {
		t.Fatalf("expected S1 and D-001-S2 slugged filenames not to be marked orphan, got errors: %v", errs)
	}
}




