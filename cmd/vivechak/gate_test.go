package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func TestRunGateCLI_EmptyWorkspace(t *testing.T) {
	dir := t.TempDir()
	// Init project metadata
	_ = os.MkdirAll(filepath.Join(dir, core.ResearchDir), 0o755)
	meta := core.WorkspaceMeta{Scope: core.ScopeProject}
	metaBytes, _ := json.Marshal(meta)
	_ = os.WriteFile(filepath.Join(dir, core.MetadataFile), metaBytes, 0o644)

	var stdout, stderr bytes.Buffer
	code := runGateWithArgs([]string{dir}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 on empty workspace, got %d", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "Phase 0 Gate: FAIL") {
		t.Errorf("expected Phase 0 Gate: FAIL in output, got: %s", out)
	}
}

func TestRunGateCLI_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, core.ResearchDir), 0o755)
	meta := core.WorkspaceMeta{Scope: core.ScopeComparison}
	metaBytes, _ := json.Marshal(meta)
	_ = os.WriteFile(filepath.Join(dir, core.MetadataFile), metaBytes, 0o644)

	var stdout, stderr bytes.Buffer
	code := runGateWithArgs([]string{"--format", "json", dir}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 on comparison workspace, got %d", code)
	}

	var res map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse JSON gate output: %v", err)
	}
	if res["status"] != "NOT_APPLICABLE" {
		t.Errorf("expected status NOT_APPLICABLE, got %v", res["status"])
	}
	if res["scope"] != "comparison" {
		t.Errorf("expected scope comparison, got %v", res["scope"])
	}
	if res["track"] != "all" {
		t.Errorf("expected track all, got %v", res["track"])
	}
}

func TestRunGateCLI_InvalidTrack(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := runGateWithArgs([]string{"--track", "invalid_track", dir}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 on invalid track, got %d", code)
	}
	errStr := stderr.String()
	if !strings.Contains(errStr, "invalid track filter") {
		t.Errorf("expected invalid track error, got: %s", errStr)
	}
}

func TestRunGateCLI_SystemDirRejection(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runGateWithArgs([]string{"C:\\Windows"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 on system dir, got %d", code)
	}
	errStr := stderr.String()
	if !strings.Contains(errStr, "application or system directory") {
		t.Errorf("expected system directory error, got: %s", errStr)
	}
}

func TestRunGateCLI_UninitializedWorkspace(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nonexistent_dir")
	var stdout, stderr bytes.Buffer
	code := runGateWithArgs([]string{dir}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 on uninitialized workspace, got %d", code)
	}
	errStr := stderr.String()
	if !strings.Contains(errStr, "workspace not initialized") {
		t.Errorf("expected workspace not initialized error, got: %s", errStr)
	}
}

func TestRunGateCLI_TrackFiltersAndPassingWorkspace(t *testing.T) {
	// Scenario 1: Decision Scope workspace that passes all gate checks (0 DAG, 0 FAD, 2 templates)
	decDir := t.TempDir()
	decResearch := filepath.Join(decDir, core.ResearchDir)
	decSessions := filepath.Join(decDir, core.SessionsDir)
	decTpls := filepath.Join(decResearch, "templates")
	_ = os.MkdirAll(decSessions, 0o755)
	_ = os.MkdirAll(decTpls, 0o755)

	decMeta := core.WorkspaceMeta{Scope: core.ScopeDecision}
	decMetaBytes, _ := json.Marshal(decMeta)
	_ = os.WriteFile(filepath.Join(decDir, core.MetadataFile), decMetaBytes, 0o644)

	for _, tpl := range core.TemplatesForScope(core.ScopeDecision) {
		_ = os.WriteFile(filepath.Join(decTpls, tpl), []byte("# Template "+tpl), 0o644)
	}

	// 1 session with evidence grades
	s1 := `---
session_id: D-001-S1
title: S1
status: complete
date: 2026-10-18
---
## Key Findings
- Finding 1 [Grade A] https://example.com/spec
## Discovered Concerns
Latency spike under heavy loads.
## Delta
Updated belief.
`
	_ = os.WriteFile(filepath.Join(decSessions, "D-001-S1.md"), []byte(s1), 0o644)

	// ADR D-001 (Two-Way Door)
	decisions := `# Architectural Decisions

| ID | Title | Status | Door | Date |
|---|---|---|---|---|
| D-001 | Logging Strategy | accepted | two-way | 2026-10-18 |
`
	_ = os.WriteFile(filepath.Join(decResearch, core.DecisionsFile), []byte(decisions), 0o644)

	adr1 := `---
id: D-001
title: Logging Strategy
status: accepted
door_type: two-way
date: 2026-10-18
review_trigger: "p95 latency above 50 ms"
---
# Context
Logging choice.
# Decision
Use slog. Grade A (https://golang.org)
`
	_ = os.WriteFile(filepath.Join(decResearch, "D-001.md"), []byte(adr1), 0o644)

	// Test 1: Full Gate on Decision Workspace passes with exit code 0
	var stdout, stderr bytes.Buffer
	code := runGateWithArgs([]string{"--verbose", decDir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0 for passing decision workspace, got %d, out: %s, err: %s", code, stdout.String(), stderr.String())
	}
	outStr := stdout.String()
	if !strings.Contains(outStr, "Phase 0 Gate: PASS") {
		t.Errorf("expected Phase 0 Gate: PASS in output, got: %s", outStr)
	}
	if !strings.Contains(outStr, "Track A (Two-Way Doors): PASS") {
		t.Errorf("expected Track A PASS, got: %s", outStr)
	}
	if !strings.Contains(outStr, "Track B (One-Way Doors): N/A") {
		t.Errorf("expected Track B N/A (no one-way decisions, so none of G6-G11 ran), got: %s", outStr)
	}

	// Test 2: Filter by track_a on decision workspace
	stdout.Reset()
	stderr.Reset()
	codeA := runGateWithArgs([]string{"--track", "track_a", "--format", "json", decDir}, &stdout, &stderr)
	if codeA != 0 {
		t.Fatalf("expected exit code 0 for track_a filter, got %d", codeA)
	}
	var resA map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &resA); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if resA["status"] != "PASS" || resA["passed"] != true {
		t.Errorf("expected status PASS and passed true, got: %v", resA)
	}

	// Scenario 2: Project Scope with partial progress - Track A passes, Track B fails
	projDir := t.TempDir()
	projResearch := filepath.Join(projDir, core.ResearchDir)
	_ = os.MkdirAll(projResearch, 0o755)
	projMeta := core.WorkspaceMeta{Scope: core.ScopeProject}
	projMetaBytes, _ := json.Marshal(projMeta)
	_ = os.WriteFile(filepath.Join(projDir, core.MetadataFile), projMetaBytes, 0o644)
	_ = os.WriteFile(filepath.Join(projResearch, core.DecisionsFile), []byte(decisions), 0o644)
	_ = os.WriteFile(filepath.Join(projResearch, "D-001.md"), []byte(adr1), 0o644)

	stdout.Reset()
	stderr.Reset()
	codeProjA := runGateWithArgs([]string{"--track", "track_a", projDir}, &stdout, &stderr)
	if codeProjA != 1 {
		t.Errorf("expected exit code 1 for track_a on failing project workspace, got %d", codeProjA)
	}

	stdout.Reset()
	stderr.Reset()
	codeProjB := runGateWithArgs([]string{"--track", "track_b", projDir}, &stdout, &stderr)
	if codeProjB != 1 {
		t.Errorf("expected exit code 1 for track_b on incomplete project workspace, got %d", codeProjB)
	}
}

// A track summary must not say PASS for rules that did not run: with only a proposed one-way ADR,
// G6–G11 are n/a, so Track B reads N/A, the n/a reason names the ADR that will be checked once it is
// accepted, and the verdict is still FAIL (G5 reports the status).
func TestRunGateCLI_TrackSummaryIsNAWhenNoRuleRan(t *testing.T) {
	dir := t.TempDir()
	research := filepath.Join(dir, core.ResearchDir)
	sessions := filepath.Join(dir, core.SessionsDir)
	_ = os.MkdirAll(sessions, 0o755)
	s1 := `---
session_id: D-001-S1
title: S1
status: complete
date: 2026-10-18
informs_decisions: [D-001]
---
## Key Findings
- Finding 1 [Grade A] https://example.com/spec
## Discovered Concerns
Latency spike under heavy loads.
## Delta
Updated belief.
`
	_ = os.WriteFile(filepath.Join(sessions, "D-001-S1.md"), []byte(s1), 0o644)
	decisions := `# Architectural Decisions

| ID | Title | Status | Door | Date |
|---|---|---|---|---|
| D-001 | Event store | proposed | one-way | 2026-10-18 |
`
	_ = os.WriteFile(filepath.Join(research, core.DecisionsFile), []byte(decisions), 0o644)
	adr := `---
id: D-001
type: decision
title: Event store
status: proposed
door_type: one-way
date: 2026-10-18
review_trigger: "p95 write latency above 50 ms"
---
# Context
Storage choice.
# Decision
Use PostgreSQL. [Grade A](https://www.postgresql.org/docs/)
`
	_ = os.WriteFile(filepath.Join(research, "D-001-decision.md"), []byte(adr), 0o644)

	var stdout, stderr bytes.Buffer
	code := runGateWithArgs([]string{dir}, &stdout, &stderr)
	out := stdout.String()
	if code != 1 || !strings.Contains(out, "Phase 0 Gate: FAIL") {
		t.Fatalf("expected FAIL with exit code 1 for a proposed one-way ADR, got %d: %s", code, out)
	}
	if !strings.Contains(out, "Track B (One-Way Doors): N/A") {
		t.Errorf("Track B must read N/A when none of G6-G11 ran, got: %s", out)
	}
	if !strings.Contains(out, "checked once accepted: D-001") {
		t.Errorf("the n/a reason must name the one-way ADR that is checked once accepted, got: %s", out)
	}
}
