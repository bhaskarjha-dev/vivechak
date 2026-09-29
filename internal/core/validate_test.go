package core

import (
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
