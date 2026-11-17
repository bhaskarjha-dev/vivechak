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

	// 7-12. CRIT-13: G7 reads each claim's own verification label, so labels other than fetched/human and prose that
	// merely mentions "human" or "fetched" never count, while G6 still counts the same claims; the formats the
	// templates teach (composite citation, ledger row) count when labelled fetched/human.
	labelCases := []struct {
		name     string
		findings string
		ledger   string
		g7Pass   bool
	}{
		{"Fail_G7_UnverifiedLabels", `- PostgreSQL indexes JSONB documents with GIN [A · single · fresh · direct · unverified](https://www.postgresql.org/docs/current/datatype-json.html)
- Logical replication supports zero-downtime upgrades [A · single · fresh · direct · unverified](https://www.postgresql.org/docs/current/logical-replication.html)
`, "- PostgreSQL manual (see findings).\n", false},
		{"Fail_G7_CachedLabels", `- PostgreSQL is best suited for durable ACID multi-table transactions [Grade A: https://www.postgresql.org/docs/ | cached].
- Redis handles high-throughput in-memory caching [Grade B: https://redis.io/benchmarks | cached].
`, "- PostgreSQL manual and Redis benchmarks (see findings).\n", false},
		{"Fail_G7_HumanWrittenProse", `- A human architect fetched and read the PostgreSQL manual [Grade A: https://www.postgresql.org/docs/].
- Human-provided benchmark notes were verified by hand [Grade B: https://redis.io/benchmarks].
`, "- PostgreSQL manual and Redis benchmarks (see findings).\n", false},
		{"Fail_G7_VerifiedLabels", `- PostgreSQL is best suited for durable ACID multi-table transactions (Grade A · corroborated · fresh | verified) https://www.postgresql.org/docs/
- Redis handles high-throughput in-memory caching (Grade B · single · fresh | verified) https://redis.io/benchmarks
`, "- PostgreSQL manual and Redis benchmarks (see findings).\n", false},
		{"Fail_G7_LedgerRowsUnverified", `- **JSONB indexing:** GIN indexes JSONB documents (Grade A · corroborated · fresh · direct | unverified).
- **Caching:** Redis Streams keep bounded memory (Grade B · single · fresh · direct | cached).
`, `| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | PostgreSQL manual https://www.postgresql.org/docs/current/datatype-json.html | A | corroborated · fresh | unverified | JSONB indexing |
| 2 | Redis Streams https://redis.io/docs/data-types/streams/ | B | single · fresh | cached | Caching |
`, false},
		// SESSION.template.md exactly: inline findings carry the composite citation, the ledger carries the URLs.
		{"Pass_G7_SessionTemplateLedger", `- **JSONB indexing:** GIN indexes JSONB documents (Grade A · corroborated · fresh · direct | fetched).
- **Caching:** Redis Streams keep bounded memory (Grade B · single · fresh · direct | human-provided).
`, `| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | PostgreSQL manual https://www.postgresql.org/docs/current/datatype-json.html | A | corroborated · fresh | fetched | JSONB indexing |
| 2 | Redis Streams https://redis.io/docs/data-types/streams/ | B | single · fresh | human-provided | Caching |
`, true},
		// FRAMEWORK.md §5.5's composite citation with the URL on the claim's line.
		{"Pass_G7_CompositeCitationWithURL", `- GIN indexes JSONB documents https://www.postgresql.org/docs/current/datatype-json.html (Grade A · corroborated · fresh · direct | fetched).
- Redis Streams keep bounded memory https://redis.io/docs/data-types/streams/ (Grade B · single · fresh · direct | human).
`, "- PostgreSQL manual and Redis docs (see findings).\n", true},
	}
	for _, tc := range labelCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, ws := setupPassingBase(t)
			s1 := "---\nsession_id: T0-01\nstatus: complete\ninforms_decisions: [\"D-001\"]\n---\n## Key Findings\n" + tc.findings +
				"\n## Recommendation\nAdopt PostgreSQL as the primary datastore with Redis Streams for caching.\n\n## Sources & Evidence Ledger\n" + tc.ledger
			if err := os.WriteFile(filepath.Join(dir, SessionsDir, "T0-01.md"), []byte(s1), 0o644); err != nil {
				t.Fatal(err)
			}
			res := EvaluateGate(ws, InspectWorkspace(dir), dir)
			status := map[string]GateRuleResult{}
			for _, r := range res.Rules {
				status[r.ID] = r
				if r.ID != "G7" && r.Status == "FAIL" {
					t.Errorf("expected only G7 to depend on the labels, but %s FAILs: %s", r.ID, r.Reason)
				}
			}
			if status["G6"].Status != "PASS" {
				t.Errorf("expected G6 to count these Grade A/B claims with URLs, got %s: %s", status["G6"].Status, status["G6"].Reason)
			}
			if tc.g7Pass {
				if status["G7"].Status != "PASS" || !res.Passed {
					t.Errorf("expected G7 and the gate to PASS for fetched/human labels, got G7 %s: %s (passed=%v)", status["G7"].Status, status["G7"].Reason, res.Passed)
				}
				return
			}
			if status["G7"].Status != "FAIL" || res.Passed {
				t.Fatalf("expected G7 to FAIL, got G7 %s: %s (passed=%v)", status["G7"].Status, status["G7"].Reason, res.Passed)
			}
			for _, want := range []string{"D-001: 0 of 2 Grade A/B claims with URLs carry a fetched or human label", "labels as reported", "exactly fetched, human or human-provided"} {
				if !strings.Contains(status["G7"].Reason, want) {
					t.Errorf("expected G7 reason to contain %q, got: %s", want, status["G7"].Reason)
				}
			}
		})
	}
}

// TestGate_OneWay_ClaimRecognizer pins the claim formats G6 counts and the labels G7 accepts (CRIT-13): G7 counts
// only claims G6 counts, and only when the claim's own label is exactly fetched, human or human-provided.
func TestGate_OneWay_ClaimRecognizer(t *testing.T) {
	cases := []struct {
		line             string
		graded, verified bool
	}{
		// Bracketed tags (the validator's hint, the Harness fixture, the gate tests' own form).
		{"- GIN indexes JSONB [A · single · fresh · direct · fetched](https://www.postgresql.org/docs/current/datatype-json.html)", true, true},
		{"- GIN indexes JSONB [a · single · fresh · direct · Fetched](https://www.postgresql.org/docs/)", true, true},
		{"- Redis benchmark [Grade B: https://redis.io/benchmarks | fetched]", true, true},
		{"- PostgreSQL docs: https://www.postgresql.org/docs/ [Grade A | fetched]", true, true},
		{"- Raft safety [E-A: https://raft.github.io · fetched]", true, true},
		{"- Raft safety [E-B: https://raft.github.io | human]", true, true},
		{"- Operator runbook [Grade A: https://example.com/runbook | human-provided]", true, true},
		// The parenthesized composite citation (SESSION.template.md, DECISIONS.template.md, FRAMEWORK.md §5.5).
		{"- **JSONB:** GIN https://www.postgresql.org/docs/ (Grade A · corroborated · fresh · direct | fetched).", true, true},
		{"- Streams https://redis.io/docs/ (B · single · fresh | human-provided)", true, true},
		{"- Vendor benchmark https://example.com/bench (Grade: B | fetched)", true, true},
		{"- Raft https://raft.github.io (E-A · fresh · human)", true, true},
		// The Sources & Evidence Ledger row: grade cell, URL and label cell in the same row.
		{"| 1 | PostgreSQL manual https://www.postgresql.org/docs/ | A | corroborated · fresh | fetched | JSONB |", true, true},
		{"| 2 | https://redis.io/benchmarks | Grade B | fresh | human | cache |", true, true},
		{"| 3 | https://raft.github.io | [A] | single | `fetched` | consensus |", true, true},
		// Claims G6 counts whose label is not fetched or human: G7 must not count them.
		{"- GIN [A · single · fresh · direct · unverified](https://www.postgresql.org/docs/)", true, false},
		{"- Redis [Grade B: https://redis.io/benchmarks | cached]", true, false},
		{"- PostgreSQL [Grade A: https://www.postgresql.org/docs/ | recalled]", true, false},
		{"- PostgreSQL [Grade A: https://www.postgresql.org/docs/ | recalled memory]", true, false},
		{"- PostgreSQL [Grade A: https://www.postgresql.org/docs/ | verified]", true, false},
		{"- Binary size https://go.dev/doc (Grade A · source inspection | verified)", true, false},
		{"- PostgreSQL [Grade A: https://www.postgresql.org/docs/ | secondhand]", true, false},
		{"- PostgreSQL [Grade A: https://www.postgresql.org/docs/ | fetched docs]", true, false},
		{"- PostgreSQL [Grade A: https://www.postgresql.org/docs/ | canonical source code]", true, false},
		{"- A human architect fetched the manual [Grade A: https://www.postgresql.org/docs/]", true, false},
		{"- Human-provided notes, verified and fetched [Grade B: https://redis.io/benchmarks] (unverified elsewhere)", true, false},
		{"- PostgreSQL [Grade A: https://www.postgresql.org/docs/] [fetched]", true, false},
		{"| 1 | https://www.postgresql.org/docs/ | A | fresh | fetched docs | JSONB |", true, false},
		{"| 1 | https://www.postgresql.org/docs/ | A | fresh |  | JSONB |", true, false},
		{"| 1 | https://www.postgresql.org/docs/ | A | fresh | unverified | JSONB |", true, false},
		// Not Grade A/B claims with a URL: neither rule counts them.
		{"- GIN indexes JSONB [A · single · fresh · direct · fetched]", false, false},
		{"- (Grade A · corroborated · fresh | fetched) with no link", false, false},
		{"- We chose option (A) after reading https://example.com/fetched", false, false},
		{"- Grade A (https://sqlite.org) fetched by a human", false, false},
		{"- [Benchmarks](https://example.com/bench) were fetched", false, false},
		{"- Vendor whitepaper [Grade C: https://example.com/whitepaper | fetched]", false, false},
		{"| # | Source | Grade | Modifiers | Verification | Used For |", false, false},
		{"|---|---|---|---|---|---|", false, false},
		{"| 1 | https://example.com/blog | C | fresh | fetched | blog |", false, false},
		{"| Option A | https://example.com | fetched |", false, false},
		{"| fetched | human |", false, false},
	}
	for _, tc := range cases {
		graded, verified := classifyClaimLine(tc.line)
		if graded != tc.graded || verified != tc.verified {
			t.Errorf("classifyClaimLine(%q) = (graded %v, verified %v), want (%v, %v)", tc.line, graded, verified, tc.graded, tc.verified)
		}
		if verified && !graded {
			t.Errorf("classifyClaimLine(%q): G7 counted a claim G6 does not count", tc.line)
		}
	}
}
