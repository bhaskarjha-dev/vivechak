package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGate_OneWay_NegativeCases verifies that Track B rules G6, G7, G10, and G11
// fail rigorously when their prerequisites are absent, recalled, or stale.
func TestGate_OneWay_NegativeCases(t *testing.T) {
	// 1. Negative G7: Informing claims carry only 'recalled' provenance (no fetched or human)
	t.Run("Fail_G7_RecalledClaims", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		s1 := `---
session_id: T0-01
status: complete
informs_decisions: ["D-001"]
---
## Recommendation
- Postgres chosen based on [Grade A: https://www.postgresql.org/docs/ | recalled memory].
- Benchmarks from [Grade B: https://tech.benchmarks.org/pg | recalled memory].
`
		if err := os.WriteFile(filepath.Join(dir, SessionsDir, "T0-01.md"), []byte(s1), 0o644); err != nil {
			t.Fatal(err)
		}
		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)

		if res.Passed {
			t.Error("expected gate failure when One-Way Door informing claims are only recalled")
		}
		foundG7Fail := false
		for _, r := range res.Rules {
			if r.ID == "G7" && r.Status == "FAIL" {
				foundG7Fail = true
				if !strings.Contains(r.Reason, "fetched or human") && !strings.Contains(r.Reason, "unverified") {
					t.Errorf("expected G7 reason to mention fetched/human or unverified, got: %s", r.Reason)
				}
				break
			}
		}
		if !foundG7Fail {
			t.Errorf("expected G7 to FAIL, rules: %+v", res.Rules)
		}
	})

	// 2. Negative G10: Challenge text present in ADR body, but challenge record file is missing
	t.Run("Fail_G10_MissingChallengeRecord", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		// Delete the challenge record file
		_ = os.Remove(filepath.Join(dir, ResearchDir, "D-001-R1.md"))

		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)

		if res.Passed {
			t.Error("expected gate failure when challenge record file is missing")
		}
		foundG10Fail := false
		for _, r := range res.Rules {
			if r.ID == "G10" && r.Status == "FAIL" {
				foundG10Fail = true
				if !strings.Contains(r.Reason, "lacks an adversarial red-team challenge record") {
					t.Errorf("expected G10 reason to mention lacking challenge record, got: %s", r.Reason)
				}
				break
			}
		}
		if !foundG10Fail {
			t.Errorf("expected G10 to FAIL, rules: %+v", res.Rules)
		}
	})

	// 3. Negative G10: Challenge record exists, but decision_hash is stale
	t.Run("Fail_G10_StaleDecisionHash", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		staleChRecord := `---
decision_id: D-001
decision_hash: deadbeef00000000000000000000000000000000000000000000000000000000
date: 2026-10-24
---
# Adversarial Challenge D-001-R1
Substantive challenge findings for D-001.
`
		if err := os.WriteFile(filepath.Join(dir, ResearchDir, "D-001-R1.md"), []byte(staleChRecord), 0o644); err != nil {
			t.Fatal(err)
		}

		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)

		if res.Passed {
			t.Error("expected gate failure when challenge record decision_hash is stale")
		}
		foundG10Stale := false
		for _, r := range res.Rules {
			if r.ID == "G10" && r.Status == "FAIL" {
				foundG10Stale = true
				if !strings.Contains(r.Reason, "stale challenge record") {
					t.Errorf("expected G10 reason to mention stale challenge record, got: %s", r.Reason)
				}
				break
			}
		}
		if !foundG10Stale {
			t.Errorf("expected G10 to FAIL due to stale challenge record, rules: %+v", res.Rules)
		}
	})

	// 4. Negative G11: ADR has human_reviewed: true, but reviewed_hash is stale
	t.Run("Fail_G11_StaleReviewedHash", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		// Modify the ADR text directly so reviewed_hash is out of date
		d1Stale := `---
id: D-001
title: Datastore Architecture
door_type: one-way
status: accepted
review_trigger: "Database write throughput > 10000 ops"
human_reviewed: true
reviewed_by: "Principal Architect"
reviewed_hash: "stalehash1234567890abcdef"
reviewed_at: 2026-10-24T12:00:00Z
informed_by_sessions: ["T0-01"]
---
# Context
We need a durable, transactional datastore.

## Decision Outcome
We adopt PostgreSQL and Redis as researched in T0-01.

## Rejected Alternatives
- MySQL: Viable option but lacks advanced JSONB indexing required for metadata.
- MongoDB: Document model flexible but lacks ACID multi-table transactions required for ledger.

## Red-Team Stress Test
### Challenge Findings
Evaluated D-001-R1 failover latency and split-brain risks under network partitions with Gary Klein premortem substantive findings.
`
		if err := os.WriteFile(filepath.Join(dir, DecisionsDir, "D-001.md"), []byte(d1Stale), 0o644); err != nil {
			t.Fatal(err)
		}

		// Also update the challenge record hash to match D-001's new outcome so G10 doesn't mask G11
		newOutcomeHash := ComputeDecisionOutcomeHash([]byte(d1Stale))
		chRecord := fmt.Sprintf(`---
decision_id: D-001
decision_hash: %s
date: 2026-10-24
---
# Adversarial Challenge D-001-R1
Substantive challenge findings for D-001.
`, newOutcomeHash)
		_ = os.WriteFile(filepath.Join(dir, ResearchDir, "D-001-R1.md"), []byte(chRecord), 0o644)

		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)

		if res.Passed {
			t.Error("expected gate failure when reviewed_hash does not match current ADR content")
		}
		foundG11Stale := false
		for _, r := range res.Rules {
			if r.ID == "G11" && r.Status == "FAIL" {
				foundG11Stale = true
				if !strings.Contains(r.Reason, "stale sign-off") && !strings.Contains(r.Reason, "content changed after review") && !strings.Contains(r.Reason, "modified after human review") {
					t.Errorf("expected G11 reason to mention stale sign-off or modified after human review, got: %s", r.Reason)
				}
				break
			}
		}
		if !foundG11Stale {
			t.Errorf("expected G11 to FAIL due to stale reviewed_hash, rules: %+v", res.Rules)
		}
	})

	// 5. HIGH-16: G6 and G7 count claims of session that lists ADR in informs_decisions
	// even when ADR lacks informed_by_sessions
	t.Run("Pass_G6_G7_InformsDecisionsLinking", func(t *testing.T) {
		dir, ws := setupPassingBase(t)

		// Create ADR without informed_by_sessions
		d1NoInformedBy := `---
id: D-001
title: Datastore Architecture
door_type: one-way
status: accepted
review_trigger: "Database write throughput > 10000 ops"
---
# Context
We need a durable, transactional datastore.

## Decision Outcome
We adopt PostgreSQL and Redis as researched in T0-01.

## Rejected Alternatives
- MySQL: Viable option but lacks advanced JSONB indexing required for metadata.
- MongoDB: Document model flexible but lacks ACID multi-table transactions required for ledger.

## Red-Team Stress Test
### Challenge Findings
Evaluated D-001-R1 failover latency and split-brain risks under network partitions with Gary Klein premortem substantive findings.
`
		if err := os.WriteFile(filepath.Join(dir, DecisionsDir, "D-001.md"), []byte(d1NoInformedBy), 0o644); err != nil {
			t.Fatal(err)
		}

		// Session T0-01 has informs_decisions: ["D-001"] and 2 verified Grade A/B claims on separate lines
		s1 := `---
session_id: T0-01
status: complete
informs_decisions: ["D-001"]
---
## Recommendation
- PostgreSQL chosen for ACID multi-table transactions [Grade A: https://www.postgresql.org/docs/ | fetched].
- Redis chosen for caching layer [Grade B: https://tech.benchmarks.org/pg | fetched].
`
		if err := os.WriteFile(filepath.Join(dir, SessionsDir, "T0-01.md"), []byte(s1), 0o644); err != nil {
			t.Fatal(err)
		}

		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)

		for _, r := range res.Rules {
			if r.ID == "G6" && r.Status != "PASS" {
				t.Errorf("expected G6 to PASS via informs_decisions link, got status %s: %s", r.Status, r.Reason)
			}
			if r.ID == "G7" && r.Status != "PASS" {
				t.Errorf("expected G7 to PASS via informs_decisions link, got status %s: %s", r.Status, r.Reason)
			}
		}
	})

	// 6. HIGH-16: When no sessions link, G6 failure reason indicates no linked sessions
	t.Run("Fail_G6_G7_NoInformingLink", func(t *testing.T) {
		dir, ws := setupPassingBase(t)

		// Create ADR without informed_by_sessions
		d1NoInformedBy := `---
id: D-001
title: Datastore Architecture
door_type: one-way
status: accepted
review_trigger: "Database write throughput > 10000 ops"
---
# Context
We need a durable, transactional datastore.

## Decision Outcome
We adopt PostgreSQL and Redis.

## Rejected Alternatives
- MySQL: Viable option but lacks advanced JSONB indexing required for metadata.
- MongoDB: Document model flexible but lacks ACID multi-table transactions required for ledger.

## Red-Team Stress Test
### Challenge Findings
Evaluated D-001-R1 failover latency and split-brain risks under network partitions with Gary Klein premortem substantive findings.
`
		if err := os.WriteFile(filepath.Join(dir, DecisionsDir, "D-001.md"), []byte(d1NoInformedBy), 0o644); err != nil {
			t.Fatal(err)
		}

		// Session T0-01 informs a different ADR, not D-001
		s1 := `---
session_id: T0-01
status: complete
informs_decisions: ["D-999"]
---
## Recommendation
Postgres chosen based on [Grade A: https://www.postgresql.org/docs/ | fetched docs].
`
		if err := os.WriteFile(filepath.Join(dir, SessionsDir, "T0-01.md"), []byte(s1), 0o644); err != nil {
			t.Fatal(err)
		}

		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)

		foundG6Fail := false
		for _, r := range res.Rules {
			if r.ID == "G6" && r.Status == "FAIL" {
				foundG6Fail = true
				if !strings.Contains(r.Reason, "no sessions linked to D-001 via informed_by_sessions or informs_decisions") {
					t.Errorf("expected G6 reason to mention no sessions linked via informed_by_sessions or informs_decisions, got: %s", r.Reason)
				}
				break
			}
		}
		if !foundG6Fail {
			t.Errorf("expected G6 to FAIL, rules: %+v", res.Rules)
		}
	})
}
