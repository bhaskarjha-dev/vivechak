package core

import (
	"strings"
	"testing"
)

func TestGenerateChallenges_RedTeam(t *testing.T) {
	sessionContent := `---
id: "T1-01"
title: "Database Engine Selection"
door_type: "one-way"
status: "completed"
---

# T1-01: Database Engine Selection

## Research Question
Which embedded database satisfies high concurrency requirements?

## Key Findings
- SQLite WAL mode handles up to 10k QPS with concurrent readers. Grade A (https://sqlite.org/wal.html)
- BadgerDB exhibits high SSD wear under write-heavy workloads. Grade B (benchmark)

## Recommendation
SQLite in WAL mode is recommended as the embedded database for our desktop agent.

## Alternatives Considered
| Option | Verdict | Key Tradeoff |
|---|---|---|
| SQLite WAL | Recommended | Zero-dependency, rock solid |
| BadgerDB | Rejected | Pure Go but high write amplification |

## Open Questions & Risks
- Concurrency contention during heavy background sync operations.
`

	res, err := GenerateChallenges([]byte(sessionContent), "T1-01", ModeRedTeam, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.SessionID != "T1-01" {
		t.Errorf("expected session_id T1-01, got %s", res.SessionID)
	}
	if res.Mode != ModeRedTeam {
		t.Errorf("expected mode red_team, got %s", res.Mode)
	}
	if len(res.Prompts) != 4 {
		t.Fatalf("expected 4 red_team prompts, got %d", len(res.Prompts))
	}

	types := map[string]bool{}
	for _, p := range res.Prompts {
		types[p.Type] = true
		if p.Prompt == "" {
			t.Errorf("empty prompt text for type %s", p.Type)
		}
		if p.Rationale == "" {
			t.Errorf("empty rationale for type %s", p.Type)
		}
	}

	expectedTypes := []string{"disconfirming_search", "assumption_audit", "steel_man_alternative", "premortem"}
	for _, et := range expectedTypes {
		if !types[et] {
			t.Errorf("missing expected prompt type %s", et)
		}
	}

	// Verify rejected alternative was extracted
	alts, ok := res.Context["rejected_alts"].([]string)
	if !ok || len(alts) == 0 || alts[0] != "BadgerDB" {
		t.Errorf("expected rejected alternative BadgerDB, got %v", res.Context["rejected_alts"])
	}
}

func TestGenerateChallenges_EvidenceAudit(t *testing.T) {
	sessionContent := `---
id: "T1-02"
title: "Sync Architecture"
---

# T1-02: Sync Architecture

## Key Findings
- WebSockets provide sub-50ms latency for real-time state sync. Grade A (https://tools.ietf.org/rfc6455)
- Server-Sent Events require HTTP/2 for multiplexing. Grade A (unverified RFC draft)
- SQLite sync conflicts can be resolved using CRDTs. Grade D (recalled memory)
`

	res, err := GenerateChallenges([]byte(sessionContent), "T1-02", ModeEvidenceAudit, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Prompts) == 0 {
		t.Fatalf("expected evidence audit prompts, got 0")
	}

	foundVerifyURL := false
	foundMissingURL := false
	foundUpgradeSecondary := false

	for _, p := range res.Prompts {
		if p.Type == "verify_primary_url" {
			foundVerifyURL = true
		}
		if p.Type == "missing_grade_a_url" {
			foundMissingURL = true
		}
		if p.Type == "upgrade_secondary_citation" {
			foundUpgradeSecondary = true
		}
	}

	if !foundVerifyURL {
		t.Errorf("missing verify_primary_url prompt")
	}
	if !foundMissingURL {
		t.Errorf("missing missing_grade_a_url prompt")
	}
	if !foundUpgradeSecondary {
		t.Errorf("missing upgrade_secondary_citation prompt")
	}
}

func TestGenerateChallenges_CrossSession(t *testing.T) {
	s1 := []byte("# T1-01\nRecommendation: Use SQLite WAL")
	s2 := []byte("# T1-02\nRecommendation: Use pure Go sync engine")
	s3 := []byte("# T1-03\nRecommendation: Use BadgerDB cache")

	allSessions := map[string][]byte{
		"T1-01": s1,
		"T1-02": s2,
		"T1-03": s3,
	}

	// Test with 3 sessions (should generate full cross-session prompts)
	res, err := GenerateChallenges(s1, "T1-01", ModeCrossSession, allSessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Prompts) < 3 {
		t.Errorf("expected at least 3 cross_session prompts, got %d", len(res.Prompts))
	}

	// Test with < 3 sessions (should generate prerequisite guidance)
	fewSessions := map[string][]byte{
		"T1-01": s1,
		"T1-02": s2,
	}
	resFew, err := GenerateChallenges(s1, "T1-01", ModeCrossSession, fewSessions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resFew.Prompts) == 0 || resFew.Prompts[0].Type != "cross_session_prerequisite" {
		t.Errorf("expected cross_session_prerequisite prompt when < 3 sessions")
	}
}

func TestGenerateChallenges_EmptyContent(t *testing.T) {
	_, err := GenerateChallenges([]byte(""), "T1-01", ModeRedTeam, nil)
	if err == nil {
		t.Errorf("expected error for empty session content")
	}
}

func TestGenerateChallenges_FallbackRecommendation(t *testing.T) {
	// Session with no Recommendation section — should fallback to title or session ID
	sessionContent := `---
id: "T1-03"
title: "Fallback Title"
---
# Just text
Some findings without explicit recommendation header.
`
	res, err := GenerateChallenges([]byte(sessionContent), "T1-03", ModeRedTeam, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rec, _ := res.Context["recommendation"].(string)
	if !strings.Contains(rec, "Fallback Title") {
		t.Errorf("expected recommendation to fallback to title, got %q", rec)
	}
}
