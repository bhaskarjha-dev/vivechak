package core

import (
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"
)

func TestIsSynthesisSession_TightPattern(t *testing.T) {
	valid := []string{"FAD", "fad", "SYN", "syn", "SYN-01", "syn-01", "SYN-2", "SYN-99"}
	for _, id := range valid {
		if !IsSynthesisSession(id) {
			t.Errorf("expected IsSynthesisSession(%q) to be true", id)
		}
	}

	invalid := []string{"SYNC-01", "sync-01", "SYNTAX", "syn_01", "SESSION-01", "T0-01", "D-01", "C-01"}
	for _, id := range invalid {
		if IsSynthesisSession(id) {
			t.Errorf("expected IsSynthesisSession(%q) to be false", id)
		}
	}
}

func TestValidateComparison_Scenarios(t *testing.T) {
	// 1. Empty content rejected
	resEmpty := ValidateComparison([]byte(""))
	if resEmpty.Status != "invalid" || !resEmpty.HasBlocking() {
		t.Errorf("expected empty comparison to be blocked, got status %s", resEmpty.Status)
	}

	// 2. Short content rejected (<150 chars)
	resShort := ValidateComparison([]byte("# Short\nToo short"))
	if !resShort.HasBlocking() {
		t.Error("expected short comparison to have blocking issues")
	}

	// 3. Unreplaced placeholder rejected
	placeholderDoc := []byte(`---
id: C-01-cmp-test
title: [Comparison Title]
status: draft
---
# [Comparison Title]
| Criterion | [Option A] | [Option B] |
|---|---|---|
| Speed | 5 | 4 |
| Total | 5 | 4 |
## Recommendation
Option A is recommended based on [Grade A: https://example.com].
Additional text to ensure length exceeds one hundred and fifty characters easily.
`)
	resPh := ValidateComparison(placeholderDoc)
	if !resPh.HasBlocking() {
		t.Error("expected unfilled template placeholder to be blocked")
	}

	// 4. Missing WEP matrix table rejected
	noTableDoc := []byte(`---
id: C-01-cmp-test
title: Pino vs Winston
status: final
---
# Pino vs Winston
No table here just plain text comparing options.
## Recommendation
Pino is recommended based on [Grade A: https://github.com/pinojs/pino].
More text to ensure it exceeds the minimum character count of 150 characters for comparisons.
`)
	resNoTable := ValidateComparison(noTableDoc)
	if !resNoTable.HasBlocking() {
		t.Error("expected missing WEP matrix table to be blocked")
	}

	// 5. Valid comparison passes
	validDoc := []byte(`---
id: C-01-cmp-pino-vs-winston
title: Pino vs Winston
status: final
---
# Pino vs Winston

## Research Question
Which logging library provides the lowest overhead and best throughput for Node.js microservices?

## Weighted Evaluation Matrix

| Criterion | Weight | Pino | Winston |
|---|---|---|---|
| Throughput | 0.50 | 5 [Grade A: https://github.com/pinojs/pino] | 3 [Grade B: benchmark] |
| Bundle Size | 0.30 | 5 [Grade A: npmjs.com/package/pino] | 4 [Grade A: npmjs.com/package/winston] |
| API Flexibility | 0.20 | 4 | 5 |
| **Weighted Total** | **1.00** | **4.80** | **3.70** |

## Recommendation
Pino is recommended because its asynchronous sonic-boom pipeline provides 5x lower latency than Winston.
`)
	resValid := ValidateComparison(validDoc)
	if resValid.HasBlocking() {
		t.Errorf("expected valid comparison to have no blocking issues, got: %v", resValid.Issues)
	}
	if resValid.Status != "valid" {
		t.Errorf("expected valid status, got %s", resValid.Status)
	}
}

func TestUnicodeTruncation_RuneBoundaries(t *testing.T) {
	// Hindi / Devanagari text
	devanagari := "विवेचक एक साक्ष्य-आधारित वास्तुकला अनुसंधान उपकरण है।"
	for maxR := 1; maxR <= len([]rune(devanagari)); maxR++ {
		truncated := TruncateRunes(devanagari, maxR)
		if !utf8.ValidString(truncated) {
			t.Fatalf("TruncateRunes produced invalid UTF-8 at maxRunes %d: %q", maxR, truncated)
		}
		runeLen := len([]rune(truncated))
		if runeLen > maxR {
			t.Errorf("expected at most %d runes, got %d", maxR, runeLen)
		}
	}

	// Japanese Kanji and Hiragana
	japanese := "こんにちは世界、システムアーキテクチャの評価を実施します。"
	truncatedJp := TruncateRunes(japanese, 10)
	if !utf8.ValidString(truncatedJp) {
		t.Fatalf("invalid UTF-8 in truncated Japanese: %q", truncatedJp)
	}
	if len([]rune(truncatedJp)) > 10 {
		t.Errorf("expected <= 10 runes, got %d", len([]rune(truncatedJp)))
	}
}

func TestGate_G1_Through_G12_Corpus(t *testing.T) {
	// Helper to create passing base workspace
	setupPassingBase := func(t *testing.T) (string, *testWorkspaceFS) {
		dir := t.TempDir()
		ws := &testWorkspaceFS{root: dir}
		_ = os.MkdirAll(filepath.Join(dir, ResearchDir, "templates"), 0o755)
		_ = os.MkdirAll(filepath.Join(dir, ResearchDir, "sessions"), 0o755)
		_ = os.MkdirAll(filepath.Join(dir, ResearchDir, "decisions"), 0o755)

		// 6 templates (G3)
		for _, tmpl := range TemplatesToCopy {
			_ = os.WriteFile(filepath.Join(dir, TemplatesDir, tmpl), []byte("# Template"), 0o644)
		}

		// Metadata
		_ = os.WriteFile(filepath.Join(dir, MetadataFile), []byte(`{"scope":"project"}`), 0o644)

		// Pipeline (G1)
		pipeline := `# Pipeline
### T0-01: Foundation
Dependencies: none
` + "```prompt\nResearch\n```\n" + `
### T1-01: Backend Architecture
Dependencies: [T0-01]
` + "```prompt\nResearch\n```\n" + `
### T2-01: Data Tier Architecture
Dependencies: [T0-01]
` + "```prompt\nResearch\n```\n" + `
### SYN-01: Synthesis
Dependencies: [T1-01, T2-01]
` + "```prompt\nSynthesize\n```\n"
		_ = os.WriteFile(filepath.Join(dir, PipelineFile), []byte(pipeline), 0o644)

		// Sessions T0-01, T1-01, T2-01 (G2)
		s1 := `---
session_id: T0-01
status: complete
informs_decisions: ["D-001", "D-002"]
---
# Foundation
## Recommendation
We select PostgreSQL [Grade A: https://www.postgresql.org/docs/] and Redis [Grade B: https://redis.io/benchmarks].
`
		_ = os.WriteFile(filepath.Join(dir, SessionsDir, "T0-01.md"), []byte(s1), 0o644)

		s2 := `---
session_id: T1-01
status: complete
---
# Backend Architecture
## Recommendation
Go microservices architecture backed by official specs [Grade A: https://go.dev].
`
		_ = os.WriteFile(filepath.Join(dir, SessionsDir, "T1-01.md"), []byte(s2), 0o644)

		s3 := `---
session_id: T2-01
status: complete
---
# Data Tier Architecture
## Recommendation
Replication with read replicas verified against vendor specs [Grade A: https://aws.amazon.com/rds/].
`
		_ = os.WriteFile(filepath.Join(dir, SessionsDir, "T2-01.md"), []byte(s3), 0o644)

		// FAD.md (G2 synthesis)
		fad := `---
id: FAD-001
session_id: FAD
title: Founding Architecture Document
synthesis_date: 2026-10-24
status: complete
schema_version: "0.1.0"
---
# Founding Architecture Document

## Executive Summary
This document provides the foundational architecture synthesized across all research sessions for the platform.
The architecture combines Go microservices with PostgreSQL [Grade A: https://www.postgresql.org/docs/] and Redis [Grade B: https://redis.io/benchmarks].
`
		_ = os.WriteFile(filepath.Join(dir, FADFile), []byte(fad), 0o644)

		// ADR D-001 (One-Way Door: G5, G6, G7, G8, G9, G10, G11)
		d1 := `---
id: D-001
title: Primary Datastore Selection
door_type: one-way
status: accepted
review_trigger: "Query latency p99 exceeds 50ms under peak load"
human_reviewed: true
informed_by_sessions: ["T0-01"]
---
# Context
We need a durable, transactional datastore.

# Decision
We adopt PostgreSQL backed by official documentation [Grade A: https://www.postgresql.org/docs/] and empirical benchmark analysis [Grade B: https://tech.benchmarks.org/pg].

# Alternatives Considered
- MySQL: Viable option but lacks advanced JSONB indexing required for metadata.
- MongoDB: Document model flexible but lacks ACID multi-table transactions required for ledger.

# Red-Team Stress Test
## Challenge Findings
Evaluated failover latency and split-brain risks under network partitions.
`
		_ = os.WriteFile(filepath.Join(dir, DecisionsDir, "D-001.md"), []byte(d1), 0o644)

		// ADR D-002 (Two-Way Door: G12)
		d2 := `---
id: D-002
title: Caching Tier
door_type: two-way
status: accepted
review_trigger: "Cache hit ratio drops below 80%"
---
# Context
In-memory caching.
# Decision
We use Redis.
`
		_ = os.WriteFile(filepath.Join(dir, DecisionsDir, "D-002.md"), []byte(d2), 0o644)

		// Consolidated DECISIONS.md (G4)
		_ = os.WriteFile(filepath.Join(dir, DecisionsFile), []byte("# Decisions\n| D-001 | Datastore |\n| D-002 | Cache |\n"), 0o644)

		return dir, ws
	}

	// 1. Positive case: full pass
	t.Run("Golden_Pass_FullProject", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)
		if !res.Passed {
			t.Fatalf("expected golden base to pass, got warnings: %v", res.Warnings)
		}
	})

	// 2. Negative G1: missing pipeline
	t.Run("Fail_G1_MissingPipeline", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		_ = os.Remove(filepath.Join(dir, PipelineFile))
		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)
		if res.Passed || res.StructuralChecks.Passed {
			t.Error("expected G1 failure when pipeline file missing")
		}
	})

	// 3. Negative G2: session still in draft status
	t.Run("Fail_G2_SessionDraft", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		sDraft := `---
session_id: T0-01
status: draft
---
# Incomplete
`
		_ = os.WriteFile(filepath.Join(dir, SessionsDir, "T0-01.md"), []byte(sDraft), 0o644)
		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)
		if res.Passed || res.StructuralChecks.Passed {
			t.Error("expected G2 failure when session is draft")
		}
	})

	// 4. Negative G3: missing template
	t.Run("Fail_G3_MissingTemplates", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		_ = os.Remove(filepath.Join(dir, TemplatesDir, "DECISIONS.template.md"))
		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)
		if res.Passed || res.StructuralChecks.Passed {
			t.Error("expected G3 failure when template is missing")
		}
	})

	// 5. Negative G4: missing ADR parity (informs D-999 which does not exist)
	t.Run("Fail_G4_MissingADRParity", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		s1 := `---
session_id: T0-01
status: complete
informs_decisions: ["D-999"]
---
# Informs missing D-999
`
		_ = os.WriteFile(filepath.Join(dir, SessionsDir, "T0-01.md"), []byte(s1), 0o644)
		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)
		// Track B checks evidentiary integrity of informing sessions
		if res.Passed && res.TrackBChecks.Passed {
			t.Error("expected failure or advisory when informing unknown decision")
		}
	})

	// 6. Negative G5: ADR has proposed status (not finalized)
	t.Run("Fail_G5_ADRStatusProposed", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		d1 := `---
id: D-001
title: Datastore
door_type: one-way
status: proposed
---
# Decision
`
		_ = os.WriteFile(filepath.Join(dir, DecisionsDir, "D-001.md"), []byte(d1), 0o644)
		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)
		if res.Passed || res.TrackBChecks.Passed {
			t.Error("expected G5 failure when ADR has status: proposed")
		}
	})

	// 7. Negative G6: One-Way Door lacks Grade A/B evidence
	t.Run("Fail_G6_LackGradeABEvidence", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		s1 := `---
session_id: T0-01
status: complete
informs_decisions: ["D-001"]
---
## Recommendation
We choose Postgres based on [Grade D: personal blog post].
`
		_ = os.WriteFile(filepath.Join(dir, SessionsDir, "T0-01.md"), []byte(s1), 0o644)
		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)
		if res.Passed || res.TrackBChecks.Passed {
			t.Error("expected G6 failure when One-Way Door has low-grade citations")
		}
	})

	// 8. Negative G7: One-Way Door has recalled provenance
	t.Run("Fail_G7_RecalledProvenance", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		s1 := `---
session_id: T0-01
status: complete
informs_decisions: ["D-001"]
---
## Recommendation
Postgres chosen based on [Grade A | recalled memory].
`
		_ = os.WriteFile(filepath.Join(dir, SessionsDir, "T0-01.md"), []byte(s1), 0o644)
		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)
		if res.Passed || res.TrackBChecks.Passed {
			t.Error("expected G7 failure when One-Way Door has recalled verification")
		}
	})

	// 9. Negative G8: One-Way Door lacks rejected alternatives
	t.Run("Fail_G8_MissingRejectedAlternatives", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		d1 := `---
id: D-001
title: Datastore
door_type: one-way
status: accepted
review_trigger: "Latency > 50ms"
human_reviewed: true
informed_by_sessions: ["T0-01"]
---
# Context
Datastore
# Decision
Postgres [Grade A: https://postgresql.org].
`
		_ = os.WriteFile(filepath.Join(dir, DecisionsDir, "D-001.md"), []byte(d1), 0o644)
		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)
		if res.Passed || res.TrackBChecks.Passed {
			t.Error("expected G8 failure when One-Way Door lacks rejected alternatives")
		}
	})

	// 10. Negative G9: One-Way Door lacks review trigger
	t.Run("Fail_G9_MissingReviewTrigger", func(t *testing.T) {
		dir, ws := setupPassingBase(t)
		d1 := `---
id: D-001
title: Datastore
door_type: one-way
status: accepted
human_reviewed: true
informed_by_sessions: ["T0-01"]
---
# Context
Datastore
# Decision
Postgres [Grade A: https://postgresql.org].
# Alternatives Considered
- MySQL: lacked JSONB support required for fast document query.
- MongoDB: lacks relational joins.
`
		_ = os.WriteFile(filepath.Join(dir, DecisionsDir, "D-001.md"), []byte(d1), 0o644)
		info := InspectWorkspace(dir)
		res := EvaluateGate(ws, info, dir)
		if res.Passed || res.TrackBChecks.Passed {
			t.Error("expected G9 failure when One-Way Door lacks review trigger")
		}
	})
}

func FuzzFrontmatter(f *testing.F) {
	seeds := [][]byte{
		[]byte("---\nid: D-001\ntitle: Test\n---\nBody"),
		[]byte("---\nstatus: complete\n---\n"),
		[]byte("No frontmatter\n"),
		[]byte("---\ninvalid yaml: : : : [[\n---\n"),
		[]byte("\xef\xbb\xbf---\nid: UTF8-BOM\n---\n"),
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Should not panic on arbitrary input
		fm, body, _ := ParseFrontmatter(data)
		_ = fm
		_ = body
	})
}
