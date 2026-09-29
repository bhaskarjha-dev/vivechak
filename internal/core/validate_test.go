package core

import (
	"os"
	"path/filepath"
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

