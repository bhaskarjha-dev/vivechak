package core

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testWorkspaceFS struct {
	root string
}

func (tfs *testWorkspaceFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(tfs.root, path))
}

func (tfs *testWorkspaceFS) ListDir(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(filepath.Join(tfs.root, path))
}

func (tfs *testWorkspaceFS) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(filepath.Join(tfs.root, path))
}

func setupTestWorkspace(t *testing.T, scope Scope) (string, *testWorkspaceFS) {
	t.Helper()
	dir := t.TempDir()
	rDir := filepath.Join(dir, ResearchDir)
	sDir := filepath.Join(rDir, "sessions")
	tDir := filepath.Join(rDir, "templates")

	if err := os.MkdirAll(sDir, 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}
	if err := os.MkdirAll(tDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}

	meta := WorkspaceMeta{Scope: scope}
	metaBytes, _ := json.Marshal(meta)
	_ = os.WriteFile(filepath.Join(dir, MetadataFile), metaBytes, 0o644)

	// Populate templates
	for _, tmpl := range TemplatesForScope(scope) {
		_ = os.WriteFile(filepath.Join(tDir, tmpl), []byte("# Template "+tmpl), 0o644)
	}

	return dir, &testWorkspaceFS{root: dir}
}

func TestEvaluateGate_ScopeProject_Empty(t *testing.T) {
	dir := t.TempDir()
	ws := &testWorkspaceFS{root: dir}
	info := InspectWorkspace(dir)

	res := EvaluateGate(ws, info, dir)
	if res.Passed {
		t.Errorf("expected empty project to fail gate, got passed=true")
	}
	if res.Status != "FAIL" {
		t.Errorf("expected FAIL, got %s", res.Status)
	}
	if res.StructuralChecks.Passed {
		t.Errorf("expected structural checks to fail")
	}
	if len(res.StructuralChecks.Issues) == 0 {
		t.Errorf("expected structural check issues")
	}
}

func TestEvaluateGate_ScopeProject_FullPass(t *testing.T) {
	dir, ws := setupTestWorkspace(t, ScopeProject)

	pipelineContent := `# Research Pipeline: MyProject

### T1-01: Session 1
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Output File** | ` + "`research/sessions/T1-01.md`" + ` |

### SYN-01: Synthesis
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Output File** | ` + "`research/FAD.md`" + ` |
`
	if err := os.WriteFile(filepath.Join(dir, PipelineFile), []byte(pipelineContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Session 1
	s1Content := `---
session_id: T1-01
status: complete
completed_at: "2026-10-13T00:00:00Z"
---
# Session Findings
Claim A [Grade A: verified by RFC 7519 https://tools.ietf.org/html/rfc7519]
`
	if err := os.WriteFile(filepath.Join(dir, "research", "sessions", "T1-01.md"), []byte(s1Content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Session 2
	s2Content := `---
session_id: T1-02
status: complete
completed_at: "2026-10-13T00:00:00Z"
---
# Session Findings 2
Claim B [Grade B: benchmark tests conducted locally]
`
	if err := os.WriteFile(filepath.Join(dir, "research", "sessions", "T1-02.md"), []byte(s2Content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Session 3
	s3Content := `---
session_id: T1-03
status: complete
completed_at: "2026-10-13T00:00:00Z"
---
# Session Findings 3
Claim C [Grade B: vendor documentation]
`
	if err := os.WriteFile(filepath.Join(dir, "research", "sessions", "T1-03.md"), []byte(s3Content), 0o644); err != nil {
		t.Fatal(err)
	}

	// FAD
	fadContent := `---
id: FAD-001
title: Founding Architecture Document
synthesis_date: "2026-10-13"
status: complete
---
# Founding Architecture Document
This comprehensive architectural blueprint synthesizes all session findings into a cohesive system design.
All major subsystems, data boundaries, and protocols are rigorously specified [Grade A: official docs https://golang.org/doc].
`
	if err := os.WriteFile(filepath.Join(dir, FADFile), []byte(fadContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Two-Way Door decision
	decContent := `---
decision_id: D-001
door_type: two-way
status: accepted
review_trigger: "Latency exceeds 50ms"
---
# D-001: Cache Strategy
Decision details.
## Reversal Triggers
- Latency > 50ms
`
	if err := os.WriteFile(filepath.Join(dir, "research", "D-001.md"), []byte(decContent), 0o644); err != nil {
		t.Fatal(err)
	}
	// Consolidated registry
	if err := os.WriteFile(filepath.Join(dir, DecisionsFile), []byte(decContent), 0o644); err != nil {
		t.Fatal(err)
	}

	info := InspectWorkspace(dir)
	res := EvaluateGate(ws, info, dir)

	if !res.Passed {
		t.Fatalf("expected gate to pass, issues:\nStructural: %v\nQuality: %v", res.StructuralChecks.Issues, res.QualityChecks.Issues)
	}
	if res.Status != "PASS" {
		t.Errorf("expected status PASS, got %s", res.Status)
	}
}

func setupOneWayDoorFixture(dir string) {
	fadContent := `---
id: FAD-001
title: Founding Architecture Document
synthesis_date: "2026-10-13"
status: complete
---
# Founding Architecture Document
This architectural blueprint synthesizes all session findings into an unified system design.
All major subsystems, interfaces, data layers, and security controls are specified with inline evidence grades [Grade A: official docs https://golang.org/doc].
`
	_ = os.WriteFile(filepath.Join(dir, PipelineFile), []byte("# Pipeline: TestApp"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, FADFile), []byte(fadContent), 0o644)

	for i := 1; i <= 3; i++ {
		sContent := fmt.Sprintf(`---
session_id: S-0%d
status: complete
---
# Session Findings
Factual claim verified by tests [Grade A: official documentation https://example.com/spec%d].
`, i, i)
		_ = os.WriteFile(filepath.Join(dir, "research", "sessions", fmt.Sprintf("S-0%d.md", i)), []byte(sContent), 0o644)
	}
}

func TestEvaluateGate_ScopeProject_OneWayDoorRigor(t *testing.T) {
	t.Run("fails when One-Way Door lacks rejected alternatives (B5)", func(t *testing.T) {
		dir, ws := setupTestWorkspace(t, ScopeProject)
		setupOneWayDoorFixture(dir)

		// D-001 is One-Way door with short/missing Rejected Alternatives
		dContent := `---
decision_id: D-001
door_type: one-way
status: accepted
review_trigger: "Storage cost exceeds budget by 50%"
---
# D-001: Storage Engine
## Context
Selecting the datastore.
## Decision Rationale
PostgreSQL is selected based on evaluation.
## Rejected Alternatives
None
## Reversal Triggers
- Storage cost exceeds budget by 50%
`
		_ = os.WriteFile(filepath.Join(dir, "research", "D-001.md"), []byte(dContent), 0o644)
		_ = os.WriteFile(filepath.Join(dir, DecisionsFile), []byte(dContent), 0o644)

		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)

		foundB5 := false
		for _, issue := range res.QualityChecks.Issues {
			if strings.Contains(issue, "B5-ALTERNATIVES") {
				foundB5 = true
				break
			}
		}
		if !foundB5 {
			t.Errorf("expected failure for B5-ALTERNATIVES, got issues: %v", res.QualityChecks.Issues)
		}
	})

	t.Run("fails when informing session has recalled verification for one-way door (B4)", func(t *testing.T) {
		dir, ws := setupTestWorkspace(t, ScopeProject)
		setupOneWayDoorFixture(dir)

		// S-01 informs D-001 and uses recalled verification in recommendation
		sContent := `---
session_id: S-01
status: complete
informs_decisions: ["D-001"]
---
# Session Findings
## Recommendation
We strongly recommend PostgreSQL based on parametric memory [Grade A: recalled].
`
		_ = os.WriteFile(filepath.Join(dir, "research", "sessions", "S-01.md"), []byte(sContent), 0o644)

		dContent := `---
decision_id: D-001
door_type: one-way
status: accepted
review_trigger: "Storage write latency exceeds 200ms"
informed_by_sessions: ["S-01"]
---
# D-001: Storage Engine
## Context
Selecting the primary relational engine for persistence.
## Decision Rationale
PostgreSQL is selected for compliance and reliability.
## Rejected Alternatives
- SQLite: does not provide the required multi-writer throughput under concurrent transaction load.
## Reversal Triggers
- Storage write latency exceeds 200ms at p99 for 3 consecutive days.
`
		_ = os.WriteFile(filepath.Join(dir, "research", "D-001.md"), []byte(dContent), 0o644)
		_ = os.WriteFile(filepath.Join(dir, DecisionsFile), []byte(dContent), 0o644)

		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)

		foundB4 := false
		for _, issue := range res.QualityChecks.Issues {
			if strings.Contains(issue, "B4-VERIFICATION") {
				foundB4 = true
				break
			}
		}
		if !foundB4 {
			t.Errorf("expected failure for B4-VERIFICATION, got issues: %v", res.QualityChecks.Issues)
		}
	})

	t.Run("fails when informing session has Grade D citations for one-way door (B3)", func(t *testing.T) {
		dir, ws := setupTestWorkspace(t, ScopeProject)
		setupOneWayDoorFixture(dir)

		// S-01 informs D-001 and uses Grade D citation in recommendation
		sContent := `---
session_id: S-01
status: complete
informs_decisions: ["D-001"]
---
# Session Findings
## Recommendation
We recommend SQLite based on blog speculation [Grade D: personal blog].
`
		_ = os.WriteFile(filepath.Join(dir, "research", "sessions", "S-01.md"), []byte(sContent), 0o644)

		dContent := `---
decision_id: D-001
door_type: one-way
status: accepted
review_trigger: "Storage write latency exceeds 200ms"
informed_by_sessions: ["S-01"]
---
# D-001: Storage Engine
## Context
Selecting the primary relational engine.
## Decision Rationale
PostgreSQL is chosen.
## Rejected Alternatives
- SQLite: does not provide the required multi-writer throughput under concurrent transaction load.
## Reversal Triggers
- Storage write latency exceeds 200ms at p99 for 3 consecutive days.
`
		_ = os.WriteFile(filepath.Join(dir, "research", "D-001.md"), []byte(dContent), 0o644)
		_ = os.WriteFile(filepath.Join(dir, DecisionsFile), []byte(dContent), 0o644)

		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)

		foundB3 := false
		for _, issue := range res.QualityChecks.Issues {
			if strings.Contains(issue, "B3-EVIDENTIARY") {
				foundB3 = true
				break
			}
		}
		if !foundB3 {
			t.Errorf("expected failure for B3-EVIDENTIARY, got issues: %v", res.QualityChecks.Issues)
		}
		if res.TrackBChecks.Passed {
			t.Errorf("expected Track B to fail for B3, got passed=true")
		}
		if res.TrackBChecks.Score == "3/3" || res.TrackBChecks.Count >= res.TrackBChecks.Total {
			t.Errorf("arithmetic contradiction: Track B failed but score is %s (count=%d, total=%d)",
				res.TrackBChecks.Score, res.TrackBChecks.Count, res.TrackBChecks.Total)
		}
	})

	t.Run("produces W-OWD-NO-CHALLENGE advisory when One-Way Door lacks red-team challenge (B11)", func(t *testing.T) {
		dir, ws := setupTestWorkspace(t, ScopeProject)
		setupOneWayDoorFixture(dir)

		dContent := `---
decision_id: D-001
door_type: one-way
status: accepted
review_trigger: "Storage cost exceeds budget by 50%"
---
# D-001: Storage Engine
## Context
Selecting the datastore.
## Decision Rationale
PostgreSQL is selected.
## Rejected Alternatives
- MySQL: does not support the required JSONB indexing performance.
## Reversal Triggers
- Storage cost exceeds budget by 50%
`
		_ = os.WriteFile(filepath.Join(dir, "research", "D-001.md"), []byte(dContent), 0o644)
		_ = os.WriteFile(filepath.Join(dir, DecisionsFile), []byte(dContent), 0o644)

		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)

		foundB11 := false
		for _, adv := range res.Advisories {
			if strings.Contains(adv, "W-OWD-NO-CHALLENGE") {
				foundB11 = true
				break
			}
		}
		if !foundB11 {
			t.Errorf("expected W-OWD-NO-CHALLENGE in advisories for un-challenged One-Way Door, got: %v", res.Advisories)
		}
	})
}

func TestEvaluateGate_ScopeDecision(t *testing.T) {
	dir, ws := setupTestWorkspace(t, ScopeDecision)

	// Decision plan
	_ = os.WriteFile(filepath.Join(dir, "research", "D-001-plan.md"), []byte("# Decision Plan"), 0o644)

	// Session
	sContent := `---
session_id: D-001-S01
status: complete
---
# Findings
Evaluation of choices [Grade A: https://tools.ietf.org/rfc/rfc1234].
`
	_ = os.WriteFile(filepath.Join(dir, "research", "sessions", "D-001-S01.md"), []byte(sContent), 0o644)

	// Decision
	dContent := `---
decision_id: D-001
door_type: two-way
status: accepted
review_trigger: "Library deprecation"
---
# D-001: Auth
Details.
## Reversal Triggers
- Library deprecation
`
	_ = os.WriteFile(filepath.Join(dir, "research", "D-001.md"), []byte(dContent), 0o644)

	info := InspectWorkspace(dir)
	res := EvaluateGate(ws, info, dir)

	if !res.Passed {
		t.Fatalf("expected decision scope gate to pass, got issues: %v", res.StructuralChecks.Issues)
	}
}

func TestEvaluateGate_ScopeComparison(t *testing.T) {
	dir, ws := setupTestWorkspace(t, ScopeComparison)

	// Comparison plan
	_ = os.WriteFile(filepath.Join(dir, "research", "C-001-plan.md"), []byte("# Comparison Plan"), 0o644)

	// Session
	sContent := `---
session_id: C-001-S01
status: complete
---
# Comparison Matrix
Option A vs Option B [Grade B: benchmark results].
`
	_ = os.WriteFile(filepath.Join(dir, "research", "sessions", "C-001-S01.md"), []byte(sContent), 0o644)

	info := InspectWorkspace(dir)
	res := EvaluateGate(ws, info, dir)

	if !res.Passed {
		t.Fatalf("expected comparison scope gate to pass, got issues: %v", res.StructuralChecks.Issues)
	}
}

func TestDetectProjectName(t *testing.T) {
	t.Run("from frontmatter", func(t *testing.T) {
		dir := t.TempDir()
		ws := &testWorkspaceFS{root: dir}
		_ = os.MkdirAll(filepath.Join(dir, ResearchDir), 0o755)
		_ = os.WriteFile(filepath.Join(dir, PipelineFile), []byte("---\nproject: QuantumCore\n---\n# Pipeline"), 0o644)

		name := DetectProjectName(ws, dir)
		if name != "QuantumCore" {
			t.Errorf("expected QuantumCore, got %s", name)
		}
	})

	t.Run("from heading with colon", func(t *testing.T) {
		dir := t.TempDir()
		ws := &testWorkspaceFS{root: dir}
		_ = os.MkdirAll(filepath.Join(dir, ResearchDir), 0o755)
		_ = os.WriteFile(filepath.Join(dir, PipelineFile), []byte("# Research Pipeline: Aurora Engine\nDetails"), 0o644)

		name := DetectProjectName(ws, dir)
		if name != "Aurora Engine" {
			t.Errorf("expected Aurora Engine, got %s", name)
		}
	})

	t.Run("fallback to folder name", func(t *testing.T) {
		dir := t.TempDir()
		ws := &testWorkspaceFS{root: dir}
		name := DetectProjectName(ws, dir)
		base := filepath.Base(dir)
		if name != base {
			t.Errorf("expected %s, got %s", base, name)
		}
	})
}

func TestRenderGateArtifact(t *testing.T) {
	dir := t.TempDir()
	ws := &testWorkspaceFS{root: dir}

	tmpl := `---
project: "[Project Name]"
date: "[YYYY-MM-DD]"
verdict: "[PENDING | PASS | FAIL]"
track_a_result: "[PENDING | PASS | FAIL]"
track_b_result: "[PENDING | PASS | FAIL]"
---

## Project: ` + "`[Project Name]`" + `
**Gate Date:** ` + "`[YYYY-MM-DD]`" + `

### Decision Routing Summary

| D-001 | ` + "`[Title]`" + ` | ` + "`[1-way/2-way]`" + ` | ` + "`[A/B]`" + ` | ` + "`[PASS/FAIL/PENDING]`" + ` |
| D-002 | ` + "`[Title]`" + ` | ` + "`[1-way/2-way]`" + ` | ` + "`[A/B]`" + ` | ` + "`[PASS/FAIL/PENDING]`" + ` |
| D-NNN | ` + "`[Title]`" + ` | ` + "`[1-way/2-way]`" + ` | ` + "`[A/B]`" + ` | ` + "`[PASS/FAIL/PENDING]`" + ` |

**Track A Result:** ` + "`[PASS / FAIL]`" + `
- [ ] Decision logged in ADR
- [ ] Reversibility confirmed
- [ ] At least one corroborated source

**Track B Result:** ` + "`[PASS / FAIL]`" + `

### Blocking Issues
- ` + "`[None — all checks passed]`" + `
`

	decisions := []GateDecisionItem{
		{
			ID:       "D-001",
			Title:    "PostgreSQL Database",
			DoorType: "two-way",
			Track:    "A",
			Status:   "PASS",
		},
	}

	rendered := RenderGateArtifact(
		ws,
		[]byte(tmpl),
		dir,
		"TestApp",
		"PASS",
		true,
		true,
		nil,
		nil,
		decisions,
	)

	out := string(rendered)
	if !strings.Contains(out, `project: "TestApp"`) {
		t.Errorf("missing rendered project in frontmatter")
	}
	if !strings.Contains(out, `verdict: "PASS"`) {
		t.Errorf("missing PASS verdict")
	}
	if !strings.Contains(out, `| D-001 | PostgreSQL Database | two-way | Track A | PASS |`) {
		t.Errorf("missing decision table entry:\n%s", out)
	}
	if !strings.Contains(out, "- [x] Decision logged in ADR") {
		t.Errorf("missing checked box for Track A")
	}
}

func TestRenderGateArtifact_B10_B11(t *testing.T) {
	dir := t.TempDir()
	ws := &testWorkspaceFS{root: dir}

	tmpl := `---
project: "[Project Name]"
date: "[YYYY-MM-DD]"
verdict: "[PENDING | PASS | FAIL]"
track_a_result: "[PENDING | PASS | FAIL]"
track_b_result: "[PENDING | PASS | FAIL]"
---

### B10. Substantive ADR Content
- [ ] Decision record contains substantive causal rationale, evaluated options, and failure modes (no placeholder stubs)

### B11. Adversarial Red-Team Challenge
- [ ] One-Way Door decisions subjected to adversarial challenge (` + "`" + `vivechak_review(mode="challenge")` + "`" + ` red-team probe or disconfirming evidence search)

**Track B Result:** ` + "`[PASS / FAIL]`" + `
`

	// Case 1: Track B passes with zero advisories -> both should be checked [x]
	renderedPass := RenderGateArtifact(
		ws,
		[]byte(tmpl),
		dir,
		"TestApp",
		"PASS",
		true,
		true,
		nil,
		nil,
		nil,
	)
	outPass := string(renderedPass)
	if !strings.Contains(outPass, "- [x] Decision record contains substantive causal rationale") {
		t.Errorf("expected B10 to be checked [x], got:\n%s", outPass)
	}
	if !strings.Contains(outPass, "- [x] One-Way Door decisions subjected to adversarial challenge") {
		t.Errorf("expected B11 to be checked [x], got:\n%s", outPass)
	}

	// Case 2: Advisories emitted for B10 and B11 -> checkboxes should remain [ ]
	renderedWithAdv := RenderGateArtifact(
		ws,
		[]byte(tmpl),
		dir,
		"TestApp",
		"PASS",
		true,
		true,
		nil,
		nil,
		nil,
		"B-DECISION-STUB: One-way door D-001 lacks mandatory structural sections",
		"W-OWD-NO-CHALLENGE: One-Way Door decision D-001 was not subjected to adversarial red-team challenge",
	)
	outAdv := string(renderedWithAdv)
	if !strings.Contains(outAdv, "- [ ] Decision record contains substantive causal rationale") {
		t.Errorf("expected B10 to remain unchecked [ ], got:\n%s", outAdv)
	}
	if !strings.Contains(outAdv, "- [ ] One-Way Door decisions subjected to adversarial challenge") {
		t.Errorf("expected B11 to remain unchecked [ ], got:\n%s", outAdv)
	}
}

func TestWorkspace_ScanAndCollect(t *testing.T) {
	dir, ws := setupTestWorkspace(t, ScopeProject)

	// Write 2 completed sessions
	s1 := `---
session_id: T1-01
status: complete
completed_at: "2026-10-12T10:00:00Z"
---
# Content 1
`
	s2 := `---
session_id: T1-02
status: complete
completed_at: "2026-10-12T11:00:00Z"
---
# Content 2
`
	_ = os.WriteFile(filepath.Join(dir, "research", "sessions", "T1-01.md"), []byte(s1), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "research", "sessions", "T1-02.md"), []byte(s2), 0o644)

	completed, err := ScanCompletedSessions(ws)
	if err != nil {
		t.Fatalf("ScanCompletedSessions failed: %v", err)
	}
	if len(completed) < 2 {
		t.Errorf("expected at least 2 completed sessions, got %d", len(completed))
	}
	if !completed["T1-01"] || !completed["T1-02"] {
		t.Errorf("expected sessions T1-01 and T1-02 in completed map, got %v", completed)
	}

	// Write decision files
	d1 := `---
decision_id: D-001
title: Auth Choice
door_type: one-way
status: accepted
---
# D-001
`
	_ = os.WriteFile(filepath.Join(dir, "research", "D-001.md"), []byte(d1), 0o644)

	decs := CollectGateDecisions(ws)
	if len(decs) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decs))
	}
	if decs[0].ID != "D-001" || decs[0].Track != "B" {
		t.Errorf("unexpected decision: %+v", decs[0])
	}
}

func TestCollectGateDecisions_DualSource(t *testing.T) {
	dir := t.TempDir()
	ws := &testWorkspaceFS{root: dir}
	_ = os.MkdirAll(filepath.Join(dir, "research"), 0o755)

	// Standalone D-01
	_ = os.WriteFile(filepath.Join(dir, "research", "D-01.md"), []byte("---\nid: D-01\ntitle: Standalone\ndoor_type: two-way\nstatus: accepted\n---\nBody"), 0o644)
	// Consolidated DECISIONS.md with D-02
	_ = os.WriteFile(filepath.Join(dir, "research", "DECISIONS.md"), []byte("---\nid: D-02\ntitle: Registry\ndoor_type: two-way\nstatus: accepted\n---\nBody"), 0o644)

	decisions := CollectGateDecisions(ws)
	if len(decisions) != 2 {
		t.Fatalf("expected 2 decisions from dual sources, got %d: %+v", len(decisions), decisions)
	}
}

func TestCollectGateDecisions_DetailsBlocks(t *testing.T) {
	dir := t.TempDir()
	ws := &testWorkspaceFS{root: dir}
	_ = os.MkdirAll(filepath.Join(dir, "research"), 0o755)

	decisionsMD := `# Architectural Decisions Log
| ID | Title |
|---|---|
| D-01 | Primary DB |

---

<details>
<summary><strong>D-01</strong> · Primary DB · <code>one-way</code> · proposed</summary>

` + "```yaml" + `
id: D-01
title: Primary DB
door_type: one-way
status: proposed
date: 2026-10-21
` + "```" + `

</details>

# D-01: Primary DB
Context.
`
	_ = os.WriteFile(filepath.Join(dir, "research", "DECISIONS.md"), []byte(decisionsMD), 0o644)

	decisions := CollectGateDecisions(ws)
	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision from details block, got %d: %+v", len(decisions), decisions)
	}
	if decisions[0].ID != "D-01" {
		t.Errorf("expected ID D-01, got: %s", decisions[0].ID)
	}
	if decisions[0].DoorType != "one-way" || decisions[0].Track != "B" {
		t.Errorf("expected Track B one-way, got: %+v", decisions[0])
	}
	if decisions[0].Status != "PROPOSED" {
		t.Errorf("expected status PROPOSED, got: %s", decisions[0].Status)
	}
}

func TestVerifyDecisionsMechanical_CaseInsensitiveDeduplication(t *testing.T) {
	dir := t.TempDir()
	ws := &testWorkspaceFS{root: dir}
	_ = os.MkdirAll(filepath.Join(dir, "research"), 0o755)

	// Standalone D-01 with lowercase id "d-01"
	d1 := `---
id: d-01
door_type: two-way
status: accepted
---
# D-01
`
	_ = os.WriteFile(filepath.Join(dir, "research", "D-01.md"), []byte(d1), 0o644)

	// Consolidated DECISIONS.md with uppercase ID "D-01"
	decisionsContent := `# Decisions
<!-- DECISION: D-01 -->
---
id: D-01
door_type: two-way
status: accepted
---
# D-01
<!-- /DECISION: D-01 -->
`
	_ = os.WriteFile(filepath.Join(dir, DecisionsFile), []byte(decisionsContent), 0o644)

	info := InspectWorkspace(dir)
	ok, issues, _ := verifyDecisionsMechanical(ws, info)
	if !ok {
		t.Fatalf("expected verifyDecisionsMechanical to pass, got issues: %v", issues)
	}
}

func TestEvaluateGate_TrackA_DecoupledFromFAD(t *testing.T) {
	dir := t.TempDir()
	ws := &testWorkspaceFS{root: dir}
	_ = os.MkdirAll(filepath.Join(dir, "research"), 0o755)

	// Two-way door decision
	d1 := `---
id: D-01
title: Logging Library
door_type: two-way
status: accepted
review_trigger: Traffic > 100k req/s
---
# Logging Library
`
	_ = os.WriteFile(filepath.Join(dir, "research", "D-01.md"), []byte(d1), 0o644)

	info := InspectWorkspace(dir)
	res := EvaluateGate(ws, info, dir)

	// Track A must pass because D-01 is accepted two-way door, even with zero FAD or DAG
	if !res.TrackAChecks.Passed {
		t.Errorf("expected Track A to pass independently of FAD, got issues: %v", res.TrackAChecks.Issues)
	}
	// Structural checks should fail because FAD and DAG are missing
	if res.StructuralChecks.Passed {
		t.Errorf("expected StructuralChecks to fail due to missing FAD/DAG")
	}
}

func TestEvaluateGate_ComparisonScopeZeroGate(t *testing.T) {
	dir := t.TempDir()
	ws := &testWorkspaceFS{root: dir}
	_ = os.MkdirAll(filepath.Join(dir, "research"), 0o755)

	// Save comparison artifact
	cmp := `---
id: C-01-cmp-pino-vs-winston
title: Pino vs Winston
date: 2026-10-18
status: final
scope_level: comparison
---
# Pino vs Winston
| Criterion | Weight |
|---|---|
| Speed | 1.0 |
`
	_ = os.WriteFile(filepath.Join(dir, "research", "C-01-comparison.md"), []byte(cmp), 0o644)

	meta := WorkspaceMeta{Scope: ScopeComparison}
	metaBytes, _ := json.Marshal(meta)
	_ = os.WriteFile(filepath.Join(dir, MetadataFile), metaBytes, 0o644)

	info := InspectWorkspace(dir)
	res := EvaluateGate(ws, info, dir)

	if !res.Passed {
		t.Errorf("expected Comparison Scope to pass gate unconditionally with zero gate overhead, got: %v", res.Status)
	}
	if res.Status != "PASS" {
		t.Errorf("expected status PASS, got %s", res.Status)
	}
}

func TestVerifyEvidentiaryIntegrity_FalsePositiveParentheses(t *testing.T) {
	dir, ws := setupTestWorkspace(t, ScopeProject)

	// One-way door ADR
	adr := `---
id: D-001
title: Message Bus Selection
door_type: one-way
status: accepted
informed_by_sessions: ["T1-01"]
review_trigger: Throughput > 1M eps
---
## Context
Selecting message bus for 500k eps telemetry ingestion.

## Alternatives Considered
- Apache Kafka: rejected due to JVM GC pause spikes.
- NATS JetStream: rejected due to Raft compaction bottleneck.

## Reversal Triggers
Throughput exceeds 1.5M events/second.
`
	if err := os.WriteFile(filepath.Join(dir, "research", "D-001.md"), []byte(adr), 0o644); err != nil {
		t.Fatal(err)
	}

	// Session T1-01 with English parenthetical phrases that previously triggered false-positives
	s1 := `---
session_id: T1-01
status: complete
informs_decisions: ["D-001"]
---
## Recommendation
We recommend Redpanda (Decision D-001) as the primary streaming bus.
Downstream hot storage uses ClickHouse (ClickHouse vs TimescaleDB).
Edge telemetry buffers in memory (Edge buffer caching) before dispatch.
Hardware references [ClickHouse] and [D-001] are strictly parenthetical.
Backing empirical benchmark: [Grade A: official docs](https://redpanda.com/docs).

## Challenge Findings
Red-team stress test completed with Gary Klein premortem.
`
	if err := os.WriteFile(filepath.Join(dir, "research", "sessions", "T1-01.md"), []byte(s1), 0o644); err != nil {
		t.Fatal(err)
	}

	info := InspectWorkspace(dir)
	blocking, _ := verifyEvidentiaryIntegrity(ws, info)

	for _, issue := range blocking {
		if strings.Contains(issue, "B3-EVIDENTIARY") {
			t.Errorf("unexpected false-positive B3-EVIDENTIARY failure: %s", issue)
		}
	}

	// Now add a genuine low-grade citation inside Recommendation section and verify B3 is properly caught
	s1LowGrade := strings.Replace(s1, "Backing empirical benchmark:", "Noted [Grade C: vendor blog claim] without verification.\nBacking empirical benchmark:", 1)
	_ = os.WriteFile(filepath.Join(dir, "research", "sessions", "T1-01.md"), []byte(s1LowGrade), 0o644)

	blockingWithLow, _ := verifyEvidentiaryIntegrity(ws, info)
	foundB3 := false
	for _, issue := range blockingWithLow {
		if strings.Contains(issue, "B3-EVIDENTIARY") {
			foundB3 = true
			break
		}
	}
	if !foundB3 {
		t.Errorf("expected B3-EVIDENTIARY to trigger for [Grade C: vendor blog claim]")
	}
}

func TestGate_DecisionScope_TwoWayDoorPasses(t *testing.T) {
	dir := t.TempDir()
	ws := DirFS(dir)

	// Create workspace in ScopeDecision
	tmplDir := filepath.Join(dir, "research", "templates")
	if err := os.MkdirAll(tmplDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "research", "sessions"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, MetadataFile), []byte(`{"scope":"decision","created_at":"2026-10-23T00:00:00Z"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Scaffold required templates
	for _, tmpl := range TemplatesForScope(ScopeDecision) {
		if err := os.WriteFile(filepath.Join(tmplDir, tmpl), []byte("# Template "+tmpl), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a completed session
	sessContent := `---
session_id: S-01
status: complete
door_type: two-way
---
# S-01: Research
Investigated pino. Reversible fast spike.
`
	if err := os.WriteFile(filepath.Join(dir, "research", "sessions", "S-01.md"), []byte(sessContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a Two-Way Door decision record
	adr := `---
id: D-001
decision_id: D-001
title: Logging Library Selection
status: accepted
door_type: two-way
reversibility:
  type: two-way
  reversal_cost: low
  fallback_strategy: "Swap wrapper to winston if pino lacks transport"
---
# D-001: Logging Library Selection

## Context
Need structured JSON logging.

## Decision
Use Pino for speed.
`
	if err := os.WriteFile(filepath.Join(dir, "research", "D-001-logging.md"), []byte(adr), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "research", "DECISIONS.md"), []byte("# Architectural Decisions\n\n"+adr), 0644); err != nil {
		t.Fatal(err)
	}

	info := InspectWorkspace(dir)
	result := EvaluateGate(ws, info, dir)
	if result.Status != "PASS" {
		t.Fatalf("expected gate status PASS for two-way door in ScopeDecision, got %s (warnings: %v, advisories: %v)", result.Status, result.Warnings, result.Advisories)
	}
	if !result.Passed {
		t.Errorf("expected result.Passed = true")
	}
}

