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

