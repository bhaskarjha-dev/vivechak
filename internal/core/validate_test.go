package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvidenceGradePattern(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		// Valid evidence grades
		{"A (official documentation)", true},
		{"B (benchmarks)", true},
		{"C (community consensus)", true},
		{"D (recalled memory)", true},
		{"E (speculation)", true},
		{"[Grade A: verified]", true},
		{"[grade B: verified]", true},
		{"(Grade A · verified)", true},
		{"(grade B)", true},
		{"Grade A", true},
		{"Grade E", true},
		{"(A · corroborated · fresh · direct | fetched)", true},
		{"[A · official docs]", true},

		// False positives that must NOT match (H-05)
		{"this is a (note) about architecture", false},
		{"see section b (discussion below)", false},
		{"refer to e (email correspondence)", false},
		{"just a regular sentence.", false},
		{"Grade F", false},
		{"(Grade X)", false},
	}

	for _, tt := range tests {
		got := evidenceGradePattern.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("evidenceGradePattern.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestValidateSession_DateFormat(t *testing.T) {
	valid := `---
session_id: T1-01
title: DB Selection
date: 2026-09-29
---
PostgreSQL 16 is recommended. A (official docs)
`
	resValid := ValidateSession([]byte(valid))
	for _, issue := range resValid.Issues {
		if issue.Code == "W-INVALID-DATE-FORMAT" {
			t.Errorf("valid date should not trigger W-INVALID-DATE-FORMAT: %v", issue)
		}
	}

	invalid := `---
session_id: T1-01
title: DB Selection
date: invalid-date
---
PostgreSQL 16 is recommended. A (official docs)
`
	resInvalid := ValidateSession([]byte(invalid))
	found := false
	for _, issue := range resInvalid.Issues {
		if issue.Code == "W-INVALID-DATE-FORMAT" {
			found = true
			if issue.Level != L3Warn {
				t.Errorf("expected level L3Warn for date format, got %v", issue.Level)
			}
			break
		}
	}
	if !found {
		t.Error("expected W-INVALID-DATE-FORMAT for 'invalid-date'")
	}
}

func TestValidateSession_SessionIDHint(t *testing.T) {
	missingID := `---
title: DB Selection
date: 2026-09-29
---
Body content here.
`
	res := ValidateSession([]byte(missingID))
	found := false
	for _, issue := range res.Issues {
		if issue.Code == "V-MISSING-FIELD" && issue.Field == "session_id" {
			found = true
			if issue.FixHint != "Add 'session_id: <value>' (or 'id: <value>') to the frontmatter block" {
				t.Errorf("expected hint to mention 'id' alias, got: %q", issue.FixHint)
			}
			break
		}
	}
	if !found {
		t.Error("expected V-MISSING-FIELD for session_id")
	}
}

func TestEvidenceGradePattern_Citations(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"Backed by [E-1] and [E-02].", true},
		{"See findings in E-001 for details.", true},
		{"No evidence citation here.", false},
	}
	for _, tt := range tests {
		got := evidenceGradePattern.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("evidenceGradePattern.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestValidateConflictResolution(t *testing.T) {
	valid := `---
id: CHK-01
decision_id: D-001
title: Postgres vs Neo4j
status: resolved
door_type: one-way
---

# Conflict Resolution: CHK-01

## 1. Conflict Summary
The team is divided on whether to adopt Postgres or Neo4j.

## 2. ACH Matrix
Hypotheses evaluated against evidence. Postgres requires fewer operational dependencies.

## 3. Resolution
Adopt Postgres with recursive CTEs.
`
	res := ValidateConflictResolution([]byte(valid))
	if res.HasBlocking() {
		t.Errorf("expected valid conflict resolution, got blocking issues: %v", res.Issues)
	}

	missingFields := `---
title: Incomplete Conflict
---
Short body.
`
	resMissing := ValidateConflictResolution([]byte(missingFields))
	if !resMissing.HasBlocking() {
		t.Error("expected blocking issues for missing required fields in conflict resolution")
	}
}

func TestCanonicalTemplates_ParseAndValidate(t *testing.T) {
	tmplDir := filepath.Join("..", "..", "templates")
	templates := []string{
		"COMPARISON-SESSION.template.md",
		"CONFLICT-RESOLUTION.template.md",
		"DECISIONS.template.md",
		"FOUNDING-ARCHITECTURE.template.md",
		"PHASE-0-GATE.template.md",
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			path := filepath.Join(tmplDir, tmpl)
			data, err := os.ReadFile(path)
			if os.IsNotExist(err) {
				t.Skip("templates directory not found (running outside repo root)")
				return
			}
			if err != nil {
				t.Fatalf("reading template %s: %v", tmpl, err)
			}

			fm, body, err := ParseFrontmatter(data)
			if err != nil {
				t.Fatalf("ParseFrontmatter failed on %s: %v", tmpl, err)
			}
			if fm == nil {
				t.Fatalf("expected non-nil frontmatter in canonical template %s", tmpl)
			}
			if len(body) == 0 {
				t.Errorf("expected non-empty body in canonical template %s", tmpl)
			}

			// Verify specific validation functions never flag V-MISSING-FRONTMATTER or W-MISSING-FRONTMATTER
			switch tmpl {
			case "DECISIONS.template.md":
				res := ValidateDecision(data)
				for _, iss := range res.Issues {
					if iss.Code == "V-MISSING-FRONTMATTER" {
						t.Errorf("%s triggered V-MISSING-FRONTMATTER: %v", tmpl, iss)
					}
				}
			case "CONFLICT-RESOLUTION.template.md":
				res := ValidateConflictResolution(data)
				for _, iss := range res.Issues {
					if iss.Code == "V-MISSING-FRONTMATTER" {
						t.Errorf("%s triggered V-MISSING-FRONTMATTER: %v", tmpl, iss)
					}
				}
			case "FOUNDING-ARCHITECTURE.template.md":
				res := ValidateArtifact(data)
				for _, iss := range res.Issues {
					if iss.Code == "W-MISSING-FRONTMATTER" {
						t.Errorf("%s triggered W-MISSING-FRONTMATTER: %v", tmpl, iss)
					}
				}
				// Verify ValidateSession accepts synthesis_date as alias for date (Fix F-03)
				resSession := ValidateSession(data)
				for _, iss := range resSession.Issues {
					if iss.Code == "V-MISSING-FIELD" && iss.Field == "date" {
						t.Errorf("ValidateSession rejected synthesis_date as alias for date: %v", iss)
					}
				}
				// Verify ValidateFAD parses canonical template without frontmatter/date errors
				resFAD := ValidateFAD(data)
				for _, iss := range resFAD.Issues {
					if iss.Code == "V-MISSING-FRONTMATTER" || (iss.Code == "W-MISSING-FIELD" && iss.Field == "date") {
						t.Errorf("ValidateFAD error on canonical template %s: %v", tmpl, iss)
					}
				}
			case "COMPARISON-SESSION.template.md":
				res := ValidateSession(data)
				for _, iss := range res.Issues {
					if iss.Code == "V-MISSING-FRONTMATTER" {
						t.Errorf("%s triggered V-MISSING-FRONTMATTER: %v", tmpl, iss)
					}
				}
			}
		})
	}
}

func TestValidatePlan(t *testing.T) {
	// Plan with valid DAG but no YAML frontmatter or evidence grades
	plan := `# Research Pipeline: Test Pipeline

### Session T1-01: Architecture
| **ID** | T1-01 |
| **Dependencies** | None |

` + "````prompt" + `
Execute architecture research
` + "````" + `

### Session SYN-01: Synthesis
| **ID** | SYN-01 |
| **Dependencies** | T1-01 |

` + "````prompt" + `
Synthesize findings
` + "````" + `
`
	res := ValidatePlan([]byte(plan))
	if res.HasBlocking() {
		t.Errorf("expected valid plan, got blocking issues: %v", res.BlockingIssues())
	}
	for _, iss := range res.Issues {
		if iss.Code == "W-MISSING-FRONTMATTER" || iss.Code == "W-NO-EVIDENCE-GRADES" {
			t.Errorf("plan triggered false positive warning: %v", iss)
		}
	}

	// Empty plan should block
	resEmpty := ValidatePlan([]byte("   "))
	if !resEmpty.HasBlocking() {
		t.Error("expected blocking error for empty plan")
	}

	// Plan with DAG cycle should report invalid DAG
	cyclicPlan := `# Research Pipeline
### Session T1-01: First
| **ID** | T1-01 |
| **Dependencies** | T1-02 |
` + "````prompt" + `
Prompt
` + "````" + `
### Session T1-02: Second
| **ID** | T1-02 |
| **Dependencies** | T1-01 |
` + "````prompt" + `
Prompt
` + "````" + `
`
	resCycle := ValidatePlan([]byte(cyclicPlan))
	if !resCycle.HasBlocking() {
		t.Error("expected blocking error for cyclic plan")
	}
}

func TestValidateFAD(t *testing.T) {
	validFAD := `---
id: FAD-001
title: Founding Architecture Document
synthesis_date: 2026-09-29
status: complete
---
# Founding Architecture Document

## Executive Summary
This document synthesizes findings across all research sessions to establish the technical foundation.
The datastore uses PostgreSQL 16 with read-replicas. A (docs)

## Key Architectural Decisions
1. Storage: NVMe direct volumes. A (benchmarks)
2. Network: Envoy mesh with mTLS. B (verified)
`
	resValid := ValidateFAD([]byte(validFAD))
	if resValid.HasBlocking() {
		t.Errorf("expected valid FAD, got blocking issues: %v", resValid.BlockingIssues())
	}
	if resValid.WarningCount() > 0 {
		t.Errorf("expected 0 warnings for valid FAD, got: %v", resValid.Issues)
	}

	// Missing frontmatter
	resNoFM := ValidateFAD([]byte("# Title without frontmatter\n" + strings.Repeat("body text ", 50)))
	foundNoFM := false
	for _, iss := range resNoFM.Issues {
		if iss.Code == "W-MISSING-FRONTMATTER" {
			foundNoFM = true
			break
		}
	}
	if !foundNoFM {
		t.Error("expected W-MISSING-FRONTMATTER for FAD without frontmatter")
	}

	// Short body
	resShort := ValidateFAD([]byte("---\nid: FAD\ntitle: FAD\ndate: 2026-09-29\n---\nShort body. A (doc)"))
	foundShort := false
	for _, iss := range resShort.Issues {
		if iss.Code == "W-SHORT-BODY" {
			foundShort = true
			break
		}
	}
	if !foundShort {
		t.Error("expected W-SHORT-BODY for short FAD")
	}

	// Missing evidence grades
	resNoGrades := ValidateFAD([]byte("---\nid: FAD\ntitle: FAD\ndate: 2026-09-29\n---\n" + strings.Repeat("Long substantive body without any evidence grades. ", 20)))
	foundNoGrades := false
	for _, iss := range resNoGrades.Issues {
		if iss.Code == "W-NO-EVIDENCE-GRADES" {
			foundNoGrades = true
			break
		}
	}
	if !foundNoGrades {
		t.Error("expected W-NO-EVIDENCE-GRADES for FAD without grades")
	}
}

func TestRecalledHighGradePattern(t *testing.T) {
	positives := []string{
		"A (recalled)",
		"Grade A (recalled)",
		"Grade B (parametric memory)",
		"[Grade A · recalled]",
		"(Grade B · recalled)",
		"[B: recalled]",
		"(A · memory)",
		"Grade A [recalled]",
		"PostgreSQL offers MVCC. [Grade A · recalled · unverified]",
	}
	for _, s := range positives {
		if !recalledHighGradePattern.MatchString(s) {
			t.Errorf("expected recalledHighGradePattern to match %q", s)
		}
	}

	negatives := []string{
		"Grade D (recalled)",
		"Grade C (memory)",
		"Grade A (verified)",
		"B (benchmarks)",
		"[A · official docs]",
		"Grade A is accepted. We recalled the earlier meeting during synthesis.",
	}
	for _, s := range negatives {
		if recalledHighGradePattern.MatchString(s) {
			t.Errorf("expected recalledHighGradePattern NOT to match %q", s)
		}
	}
}

func TestValidateSession_RecalledGradeCap(t *testing.T) {
	// Session with properly capped Grade D recalled citation
	validRecalled := `---
session_id: T1-01
title: Capped Session
date: 2026-09-29
---
Claim verified from docs: Grade A (official docs).
Unverified claim from memory: Grade D (recalled).
`
	resValid := ValidateSession([]byte(validRecalled))
	for _, iss := range resValid.Issues {
		if iss.Code == "W-RECALLED-GRADE-CAP" {
			t.Errorf("unexpected W-RECALLED-GRADE-CAP for Grade D recalled: %v", iss)
		}
	}

	// Session with violation: Grade A recalled
	invalidRecalled := `---
session_id: T1-01
title: Uncapped Session
date: 2026-09-29
---
Claim: SQLite has WAL mode. A (recalled)
`
	resInvalid := ValidateSession([]byte(invalidRecalled))
	found := false
	for _, iss := range resInvalid.Issues {
		if iss.Code == "W-RECALLED-GRADE-CAP" {
			found = true
			if iss.Level != L3Warn {
				t.Errorf("expected level L3Warn, got %v", iss.Level)
			}
			break
		}
	}
	if !found {
		t.Error("expected W-RECALLED-GRADE-CAP when Grade A is recalled")
	}
}



