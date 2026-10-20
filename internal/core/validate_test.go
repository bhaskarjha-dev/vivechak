package core

import (
	"fmt"
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
		{"[Grade: A]", true},
		{"[Grade: A] (https://github.com/pinojs/pino)", true},
		{"(Grade: B)", true},
		{"Grade: C", true},
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
		"Grade C (memory)",
		"[Grade A · recalled]",
		"(Grade B · recalled)",
		"[Grade C · recalled]",
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
		"Grade D (memory)",
		"Grade E (recalled)",
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

func TestValidateDecision_EvidentiaryGrounding(t *testing.T) {
	// Missing evidence grades should produce W-NO-EVIDENCE-GRADES
	noGrades := `---
decision_id: D-001
title: Decision Without Grades
status: accepted
door_type: one-way
---
# Context
We need a datastore for high volume events.
# Decision
We will use Kafka.
# Consequences
High operational overhead.
`
	res := ValidateDecision([]byte(noGrades))
	foundNoGrades := false
	for _, iss := range res.Issues {
		if iss.Code == "W-NO-EVIDENCE-GRADES" {
			foundNoGrades = true
			break
		}
	}
	if !foundNoGrades {
		t.Error("expected W-NO-EVIDENCE-GRADES for decision without inline grades")
	}

	// Recalled Grade A should produce W-RECALLED-GRADE-CAP
	recalledA := `---
decision_id: D-002
title: Decision With Recalled Grade A
status: accepted
door_type: one-way
---
# Context
We evaluated storage backends.
# Decision
PostgreSQL handles 100k writes/sec. A (recalled)
# Consequences
No issues anticipated.
`
	resRecalled := ValidateDecision([]byte(recalledA))
	foundRecalledCap := false
	for _, iss := range resRecalled.Issues {
		if iss.Code == "W-RECALLED-GRADE-CAP" {
			foundRecalledCap = true
			break
		}
	}
	if !foundRecalledCap {
		t.Error("expected W-RECALLED-GRADE-CAP for decision with recalled Grade A")
	}
}

func TestValidateDecision_StatusProposed_SuppressesEvidenceWarning(t *testing.T) {
	proposed := `---
id: D-001
title: Initial Repository Information Architecture
status: proposed
door_type: one-way
---
# D-001: Initial Repository Information Architecture
## Initial Context & Hypotheses
This is a seed hypothesis to be researched during pipeline execution.
`
	res := ValidateDecision([]byte(proposed))
	for _, iss := range res.Issues {
		if iss.Code == "W-NO-EVIDENCE-GRADES" {
			t.Errorf("spurious W-NO-EVIDENCE-GRADES for decision with status: proposed: %v", iss)
		}
	}
}

func TestValidateConflictResolution_EvidentiaryGrounding(t *testing.T) {
	// Missing evidence grades should produce W-NO-EVIDENCE-GRADES
	noGrades := `---
id: CR-001
decision_id: D-001
title: Conflict Resolution Without Grades
status: resolved
door_type: one-way
---
# Conflict Summary
Model A recommended Redis while Model B recommended Memcached.
# ACH Matrix
We weighed latency and persistence.
# Resolution
We selected Redis.
`
	res := ValidateConflictResolution([]byte(noGrades))
	foundNoGrades := false
	for _, iss := range res.Issues {
		if iss.Code == "W-NO-EVIDENCE-GRADES" {
			foundNoGrades = true
			break
		}
	}
	if !foundNoGrades {
		t.Error("expected W-NO-EVIDENCE-GRADES for conflict resolution without inline grades")
	}

	// Recalled Grade B should produce W-RECALLED-GRADE-CAP
	recalledB := `---
id: CR-002
decision_id: D-002
title: Conflict Resolution With Recalled Grade B
status: resolved
door_type: one-way
---
# Conflict Summary
Disagreement on query latency.
# ACH Matrix
Benchmarked throughput: Redis is faster. B (recalled)
# Resolution
Resolved in favor of Redis.
`
	resRecalled := ValidateConflictResolution([]byte(recalledB))
	foundRecalledCap := false
	for _, iss := range resRecalled.Issues {
		if iss.Code == "W-RECALLED-GRADE-CAP" {
			foundRecalledCap = true
			break
		}
	}
	if !foundRecalledCap {
		t.Error("expected W-RECALLED-GRADE-CAP for conflict resolution with recalled Grade B")
	}
}

func TestValidateFAD_RecalledGradeCap(t *testing.T) {
	fadContent := `---
id: FAD
title: Founding Architecture Document
synthesis_date: 2026-09-30
status: complete
---
# Architecture Blueprint
Primary database is CockroachDB for multi-region replication. Grade A (parametric memory).
` + strings.Repeat("Extensive architectural details and component boundaries. ", 10)

	res := ValidateFAD([]byte(fadContent))
	found := false
	for _, iss := range res.Issues {
		if iss.Code == "W-RECALLED-GRADE-CAP" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected W-RECALLED-GRADE-CAP for FAD with recalled Grade A")
	}
}

func TestValidateArtifact_RecalledGradeCap(t *testing.T) {
	artContent := `---
title: Research Synthesis
---
# Findings
Selected tool is PostgreSQL. Grade B (recalled from memory).
` + strings.Repeat("Additional details on data storage and retention policy. ", 10)

	res := ValidateArtifact([]byte(artContent))
	found := false
	for _, iss := range res.Issues {
		if iss.Code == "W-RECALLED-GRADE-CAP" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected W-RECALLED-GRADE-CAP for generic artifact with recalled Grade B")
	}
}

func TestValidateSession_TemplateFile(t *testing.T) {
	tmplPath := filepath.Join("..", "..", "templates", "SESSION.template.md")
	data, err := os.ReadFile(tmplPath)
	if err != nil {
		t.Fatalf("failed to read SESSION.template.md: %v", err)
	}

	filled := string(data)
	filled = strings.ReplaceAll(filled, `"[SESSION-ID]"`, `"T1-01"`)
	filled = strings.ReplaceAll(filled, `"[Session Title]"`, `"Database Selection"`)
	filled = strings.ReplaceAll(filled, `"[YYYY-MM-DD]"`, `"2026-10-01"`)
	filled = strings.ReplaceAll(filled, `"[topic-slug]"`, `"database-selection"`)
	filled = strings.ReplaceAll(filled, `"[high|medium|low]"`, `"high"`)
	filled = strings.ReplaceAll(filled, `"[A-E]"`, `"A"`)
	filled = strings.ReplaceAll(filled, `"[modifiers]"`, `"corroborated · fresh"`)
	filled = strings.ReplaceAll(filled, `"[verification]"`, `"fetched"`)
	filled = strings.ReplaceAll(filled, `status: draft`, `status: complete`)

	res := ValidateSession([]byte(filled))
	if res.ErrorCount() > 0 {
		t.Errorf("filled SESSION.template.md produced validation errors (%d): %v", res.ErrorCount(), res.Issues)
	}
}

func TestValidateSession_TechnicalMemoryNotRecalled(t *testing.T) {
	content := `---
session_id: T1-01
title: Cache Selection
date: 2026-09-29
status: complete
---
# Key Findings
- Redis provides microsecond in-memory performance. Grade A (official docs: in-memory caching)
- Shared memory buffer limits are documented in RFC 9110. Grade A (RFC 9110: shared memory)
`
	res := ValidateSession([]byte(content))
	for _, iss := range res.Issues {
		if iss.Code == "W-RECALLED-GRADE-CAP" {
			t.Errorf("technical computing memory was falsely flagged with W-RECALLED-GRADE-CAP: %v", iss)
		}
	}
}

func TestValidateSession_QualifiedProvenance(t *testing.T) {
	vagueOneWay := `---
session_id: T1-01
title: Database Selection
date: 2026-09-29
door_type: one-way
status: complete
---
## Key Findings
- PostgreSQL has JSONB support. Grade A (fetched)

## Discovered Concerns
- High write amplification in heavy OLTP workloads could degrade NVMe SSD lifespan significantly.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | official docs | Grade A | fresh | fetched | JSONB support |
`
	res := ValidateSessionWithContext([]byte(vagueOneWay), true)
	foundBlock := false
	for _, iss := range res.Issues {
		if iss.Code == "V-OWD-VAGUE-PROVENANCE" && iss.Level == L2Block {
			foundBlock = true
			break
		}
	}
	if !foundBlock {
		t.Errorf("expected V-OWD-VAGUE-PROVENANCE block for one-way door with vague source, got issues: %v", res.Issues)
	}

	// Two-way door should produce warning, not block
	vagueTwoWay := strings.Replace(vagueOneWay, "door_type: one-way", "door_type: two-way", 1)
	resTwoWay := ValidateSessionWithContext([]byte(vagueTwoWay), false)
	foundWarn := false
	for _, iss := range resTwoWay.Issues {
		if iss.Code == "W-VAGUE-PROVENANCE" && iss.Level == L3Warn {
			foundWarn = true
			break
		}
	}
	if !foundWarn {
		t.Errorf("expected W-VAGUE-PROVENANCE warning for two-way door with vague source, got: %v", resTwoWay.Issues)
	}
}

func TestValidateSession_DiscoveredConcerns(t *testing.T) {
	trivialConcerns := `---
session_id: T1-01
title: Auth Selection
date: 2026-09-29
door_type: one-way
status: complete
---
## Key Findings
- Keycloak supports OIDC. Grade A (https://keycloak.org/docs)

## Discovered Concerns
None.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | https://keycloak.org/docs | Grade A | fresh | fetched | OIDC support |
`
	res := ValidateSessionWithContext([]byte(trivialConcerns), true)
	foundBlock := false
	for _, iss := range res.Issues {
		if iss.Code == "V-OWD-NO-CONCERNS" && iss.Level == L2Block {
			foundBlock = true
			break
		}
	}
	if !foundBlock {
		t.Errorf("expected V-OWD-NO-CONCERNS block for one-way door with trivial concerns, got: %v", res.Issues)
	}
}

func TestValidateEvidenceProvenance_QualifiedTypes(t *testing.T) {
	qualifiedSources := []string{
		"https://w3.org/TR/webauthn-2",
		"wails.io documentation",
		"RFC 9110 HTTP Semantics",
		"ISO 27001 standard",
		"github.com/mattn/go-sqlite3",
		"gitlab.com/group/repo",
	}

	for _, src := range qualifiedSources {
		session := fmt.Sprintf(`---
session_id: T1-01
title: Auth Selection
date: 2026-09-29
door_type: one-way
status: complete
---
## Key Findings
- Finding supported. Grade A (fetched)

## Discovered Concerns
- High computational overhead under sustained burst load conditions.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | %s | Grade A | fresh | fetched | Auth analysis |
`, src)

		res := ValidateSessionWithContext([]byte(session), true)
		for _, iss := range res.Issues {
			if iss.Code == "V-OWD-VAGUE-PROVENANCE" || iss.Code == "W-VAGUE-PROVENANCE" {
				t.Errorf("source %q was falsely flagged as vague: %v", src, iss)
			}
		}
	}
}

func TestValidateEvidenceProvenance_FiveColumnLedger(t *testing.T) {
	session := `---
session_id: T1-01
title: Memory Buffering
date: 2026-10-01
door_type: one-way
status: complete
---
## Key Findings
- In-memory ring buffer keeps allocations bounded. Grade A (fetched)

## Discovered Concerns
Ring buffer overflows drop oldest frames without disk backpressure.

## Sources & Evidence Ledger
| Claim | Grade | Modifiers | Method | Source |
|---|---|---|---|---|
| RAM allocation | Grade A | direct | fetched | https://golang.org/pkg/sync |
| Heap pause | Grade A | fresh | fetched | https://golang.org/doc/gc |
`
	res := ValidateSessionWithContext([]byte(session), true)
	for _, iss := range res.Issues {
		if iss.Code == "V-OWD-VAGUE-PROVENANCE" || iss.Code == "W-VAGUE-PROVENANCE" {
			t.Errorf("5-column ledger falsely flagged vague provenance: %v", iss)
		}
	}
}

func TestValidateEvidenceProvenance_MarkdownLinksAndColumnVariants(t *testing.T) {
	t.Run("4-column table with Citation header and markdown link", func(t *testing.T) {
		session := `---
session_id: T1-01
title: Transport
date: 2026-10-01
door_type: one-way
status: complete
---
## Key Findings
- QUIC transport delivers low latency connection migration. Grade A (fetched)

## Discovered Concerns
UDP blocking by intermediate enterprise firewalls requires TCP fallback paths.

## Sources & Evidence Ledger
| Claim | Grade | Method | Citation |
|---|---|---|---|
| QUIC migration | Grade A | fetched | [RFC 9000 Specification](https://www.rfc-editor.org/rfc/rfc9000) |
`
		res := ValidateSessionWithContext([]byte(session), true)
		for _, iss := range res.Issues {
			if iss.Code == "V-OWD-VAGUE-PROVENANCE" || iss.Code == "W-VAGUE-PROVENANCE" {
				t.Errorf("4-column ledger with markdown link falsely flagged: %v", iss)
			}
		}
	})

	t.Run("5-column table with Reference header and domain anchor", func(t *testing.T) {
		session := `---
session_id: T1-02
title: Framework Selection
date: 2026-10-01
door_type: one-way
status: complete
---
## Key Findings
- Framework provides native desktop bindings. Grade A (fetched)

## Discovered Concerns
Cross-compilation requires platform-specific toolchains installed on CI agents.

## Sources & Evidence Ledger
| Claim | Grade | Modifiers | Method | Reference |
|---|---|---|---|---|
| Desktop bindings | Grade A | fresh | fetched | wails.io documentation |
`
		res := ValidateSessionWithContext([]byte(session), true)
		for _, iss := range res.Issues {
			if iss.Code == "V-OWD-VAGUE-PROVENANCE" || iss.Code == "W-VAGUE-PROVENANCE" {
				t.Errorf("5-column ledger with Reference header falsely flagged: %v", iss)
			}
		}
	})

	t.Run("6-column table with Source / Anchor header and canonical URL", func(t *testing.T) {
		session := `---
session_id: T1-03
title: Consensus Protocol
date: 2026-10-01
door_type: one-way
status: complete
---
## Key Findings
- Raft leader election converges in bounded time. Grade A (fetched)

## Discovered Concerns
Split-brain partition in 2-node cluster prevents quorum without witness node.

## Sources & Evidence Ledger
| # | Claim | Grade | Modifiers | Method | Source / Anchor |
|---|---|---|---|---|---|
| 1 | Raft convergence | Grade A | direct | fetched | [Raft Paper](https://raft.github.io/raft.pdf) |
`
		res := ValidateSessionWithContext([]byte(session), true)
		for _, iss := range res.Issues {
			if iss.Code == "V-OWD-VAGUE-PROVENANCE" || iss.Code == "W-VAGUE-PROVENANCE" {
				t.Errorf("6-column ledger with Source / Anchor header falsely flagged: %v", iss)
			}
		}
	})

	t.Run("numbered H2 and H3 Sources headers parsed accurately", func(t *testing.T) {
		session := `---
session_id: T1-04
title: Numbered Header
date: 2026-10-01
door_type: one-way
status: complete
---
## Key Findings
- Distributed tracing overhead is bounded. Grade A (fetched)

## Discovered Concerns
Sampling rate trade-offs under high volume require dynamic rate limiting.

## 5. Sources & Evidence Ledger
| # | Claim | Grade | Method | Source |
|---|---|---|---|---|
| 1 | Tracing overhead | Grade A | fetched | https://opentelemetry.io/docs/specs/otel/ |
`
		res := ValidateSessionWithContext([]byte(session), true)
		for _, iss := range res.Issues {
			if iss.Code == "V-OWD-VAGUE-PROVENANCE" || iss.Code == "W-VAGUE-PROVENANCE" {
				t.Errorf("numbered heading ledger falsely flagged: %v", iss)
			}
		}
	})
}

func TestValidateDiscoveredConcerns_LengthBoundary(t *testing.T) {
	// 49 characters: below 50 char threshold
	shortConcerns := `---
session_id: T1-01
title: Auth Selection
date: 2026-09-29
door_type: one-way
status: complete
---
## Key Findings
- Finding text. Grade A (https://example.com)

## Discovered Concerns
1234567890123456789012345678901234567890123456789

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | https://example.com | Grade A | fresh | fetched | Auth analysis |
`
	resShort := ValidateSessionWithContext([]byte(shortConcerns), true)
	foundBlock := false
	for _, iss := range resShort.Issues {
		if iss.Code == "V-OWD-NO-CONCERNS" {
			foundBlock = true
			break
		}
	}
	if !foundBlock {
		t.Errorf("expected V-OWD-NO-CONCERNS for 49 characters, got: %v", resShort.Issues)
	}

	// 50 characters: meets threshold
	validConcerns := strings.Replace(shortConcerns,
		"1234567890123456789012345678901234567890123456789",
		"12345678901234567890123456789012345678901234567890", 1)
	resValid := ValidateSessionWithContext([]byte(validConcerns), true)
	for _, iss := range resValid.Issues {
		if iss.Code == "V-OWD-NO-CONCERNS" {
			t.Errorf("50 characters was falsely blocked: %v", iss)
		}
	}

	// Two-way door with trivial concerns triggers W-NO-CONCERNS warning
	twoWayTrivial := strings.Replace(shortConcerns, "door_type: one-way", "door_type: two-way", 1)
	twoWayTrivial = strings.Replace(twoWayTrivial, "1234567890123456789012345678901234567890123456789", "None.", 1)
	resTwoWay := ValidateSessionWithContext([]byte(twoWayTrivial), false)
	foundWarn := false
	for _, iss := range resTwoWay.Issues {
		if iss.Code == "W-NO-CONCERNS" && iss.Level == L3Warn {
			foundWarn = true
			break
		}
	}
	if !foundWarn {
		t.Errorf("expected W-NO-CONCERNS warning for two-way door with trivial concerns, got: %v", resTwoWay.Issues)
	}
}

func TestValidateEvidenceProvenance_IgnoresCodeBlocks(t *testing.T) {
	body := `
## Key Findings
- Primary finding with Grade A (official documentation).

Here is an example code block with sample evidence:
` + "```markdown" + `
## Sources
| Grade | Modifiers | Status | Source |
| Grade A | fetched | valid | missing-url-here |
` + "```" + `

## 4. Sources & Evidence Ledger
| Grade | Modifiers | Status | Source |
| Grade A | fetched | valid | https://docs.example.com/api |
`
	res := &ValidationResult{Status: "valid"}
	ValidateEvidenceProvenance(body, true, res)
	for _, issue := range res.Issues {
		if issue.Code == "V-MISSING-PROVENANCE" {
			t.Errorf("code block sources table falsely triggered provenance blocking: %v", issue)
		}
	}
}

func TestValidateComparison(t *testing.T) {
	validCmp := `---
id: C-01-cmp-pino-vs-winston
title: Pino vs Winston Comparison
date: 2026-10-18
status: final
scope_level: comparison
---
# Pino vs Winston Comparison

## Recommendation
Pino is recommended for high-throughput microservices.

## Weighted Evaluation Matrix
| Criterion | Weight | Pino | Winston |
|---|---|---|---|
| Latency | 0.6 | 5 (Grade A: official benchmarks) | 3 (Grade B: local tests) |
| Ecosystem | 0.4 | 4 (Grade A: npm stats) | 5 (Grade A: npm stats) |
| **Weighted Total** | 1.0 | 4.6 | 3.8 |
`
	res := ValidateComparison([]byte(validCmp))
	if res.Status != "valid" {
		t.Errorf("expected valid comparison, got: %s (issues: %v)", res.Status, res.Issues)
	}

	// Missing WEP matrix should produce W-MISSING-WEP-MATRIX warning
	noWEP := `---
id: C-02-cmp-algo
title: Algorithm Comparison
date: 2026-10-18
status: draft
---
# Algorithm Comparison
Discussion of B-Tree vs LSM Tree without any evaluation matrix table or weights.
Grade A: fetched from official paper.
`
	resNoWEP := ValidateComparison([]byte(noWEP))
	foundWarn := false
	for _, iss := range resNoWEP.Issues {
		if iss.Code == "W-MISSING-WEP-MATRIX" {
			foundWarn = true
			break
		}
	}
	if !foundWarn {
		t.Errorf("expected W-MISSING-WEP-MATRIX for comparison lacking matrix table, got: %v", resNoWEP.Issues)
	}
}

func TestValidateDecisionRegistry(t *testing.T) {
	// 1. Registry with markdown preamble and anchored blocks in status: proposed
	content := `# Architectural Decision Log

| Decision | Title | Status |
|---|---|---|
| D-01 | DB Choice | proposed |
| D-02 | Cache Choice | proposed |

<!-- DECISION: D-01 -->
<details>
<summary><b>D-01: DB Choice</b></summary>

---
id: D-01
title: DB Choice
status: proposed
door_type: one-way
date: 2026-10-21
---
Short skeleton body for seed decision.
</details>
<!-- /DECISION: D-01 -->

<!-- DECISION: D-02 -->
<details>
<summary><b>D-02: Cache Choice</b></summary>

---
id: D-02
title: Cache Choice
status: proposed
door_type: two-way
date: 2026-10-21
---
Another short skeleton body.
</details>
<!-- /DECISION: D-02 -->
`
	res := ValidateDecisionRegistry([]byte(content))
	if res.Status != "valid" {
		t.Fatalf("expected valid status for multi-decision registry, got: %s (issues: %v)", res.Status, res.Issues)
	}
	if res.ErrorCount() != 0 {
		t.Errorf("expected 0 errors, got: %d (%v)", res.ErrorCount(), res.Issues)
	}
	if res.WarningCount() != 0 {
		t.Errorf("expected 0 warnings for proposed seed decisions, got: %d (%v)", res.WarningCount(), res.Issues)
	}

	// 2. Registry with unanchored decision blocks
	unanchored := `# Decision Registry
### D-01: Storage Engine
---
id: D-01
title: Storage Engine
status: proposed
door_type: one-way
date: 2026-10-21
---
Context for D-01.

### D-02: Caching Layer
---
id: D-02
title: Caching Layer
status: proposed
door_type: two-way
date: 2026-10-21
---
Context for D-02.
`
	resUnanchored := ValidateDecisionRegistry([]byte(unanchored))
	if resUnanchored.Status != "valid" {
		t.Fatalf("expected valid for unanchored registry, got: %s (issues: %v)", resUnanchored.Status, resUnanchored.Issues)
	}

	// 3. Registry with a defective decision block (missing required title)
	defective := `# Decisions Log
<!-- DECISION: D-01 -->
---
id: D-01
status: proposed
door_type: one-way
date: 2026-10-21
---
Body without title.
<!-- /DECISION: D-01 -->
`
	resDefective := ValidateDecisionRegistry([]byte(defective))
	if !resDefective.HasBlocking() {
		t.Fatalf("expected blocking error for decision missing title, got status: %s", resDefective.Status)
	}
	foundPrefixed := false
	for _, iss := range resDefective.Issues {
		if strings.HasPrefix(iss.Prefix, "DECISION D-01:") || strings.HasPrefix(iss.String(), "DECISION D-01:") {
			foundPrefixed = true
			break
		}
	}
	if !foundPrefixed {
		t.Errorf("expected issue prefixed with 'DECISION D-01:', got: %v", resDefective.Issues)
	}

	// 4. Empty registry content
	resEmpty := ValidateDecisionRegistry([]byte(""))
	if !resEmpty.HasBlocking() {
		t.Errorf("expected blocking error for empty registry, got: %s", resEmpty.Status)
	}
}

