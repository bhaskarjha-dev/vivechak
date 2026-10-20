package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestComprehensive_PermissiveFrontmatterWithLeadingHeaders verifies that frontmatter parsing
// tolerates top-level titles, subheadings, and blank lines preceding the opening frontmatter block.
func TestComprehensive_PermissiveFrontmatterWithLeadingHeaders(t *testing.T) {
	input := `# System Architecture Document
### Subtitle: Data Storage Layer

---
session_id: T1-01
title: Storage Architecture
status: complete
date: 2026-10-07
door_type: one-way
---

## Executive Summary
This document outlines our storage strategy.
`
	fm, body, err := core.ParseFrontmatter([]byte(input))
	if err != nil {
		t.Fatalf("expected frontmatter to parse successfully, got error: %v", err)
	}
	if fm == nil {
		t.Fatalf("expected parsed frontmatter, got nil")
	}
	if fm.GetString("session_id") != "T1-01" {
		t.Errorf("expected session_id T1-01, got %q", fm.GetString("session_id"))
	}
	if fm.GetString("door_type") != "one-way" {
		t.Errorf("expected door_type one-way, got %q", fm.GetString("door_type"))
	}
	if !strings.Contains(string(body), "Executive Summary") {
		t.Errorf("expected body to contain Executive Summary, got: %s", string(body))
	}
}

// TestComprehensive_ADRSupersessionNoDuplicateFrontmatter verifies that superseding an ADR
// with a leading title never generates duplicate frontmatter blocks.
func TestComprehensive_ADRSupersessionNoDuplicateFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	researchDir := filepath.Join(tmpDir, core.ResearchDir)
	if err := os.MkdirAll(researchDir, 0o755); err != nil {
		t.Fatalf("mkdir research: %v", err)
	}

	d1Content := `# Architectural Decision: D-001 Storage Engine
---
id: D-001
title: Initial Datastore
status: accepted
door_type: one-way
date: 2026-09-01
---

# Context
We need a robust datastore for persistent transactional logs.
## Evaluated Options
- RocksDB: fast embedded storage.
## Decision
Adopt RocksDB. Grade A (benchmark)
## Rejected Alternatives
- SQLite: rejected due to write contention under high write frequency.
`
	if err := os.WriteFile(filepath.Join(researchDir, "D-001.md"), []byte(d1Content), 0o644); err != nil {
		t.Fatalf("writing D-001: %v", err)
	}

	d2Content := `---
id: D-002
title: High Throughput Storage
status: accepted
door_type: one-way
date: 2026-10-07
---

# Context
Workload has scaled to 100k writes/sec.
## Decision
Migrate to Pebble. Grade A (benchmark)
## Rejected Alternatives
- RocksDB: rejected due to cgo overhead and compaction stalls.
`
	ctx := context.Background()
	_, env, err := handleRecordDecision(ctx, nil, RecordDecisionInput{
		ProjectRoot: tmpDir,
		DecisionID:  "D-002",
		Supersedes:  "D-001",
		Content:     d2Content,
	})
	if err != nil || !env.Success {
		t.Fatalf("handleRecordDecision failed: %v, env: %+v", err, env)
	}

	oldBytes, err := os.ReadFile(filepath.Join(researchDir, "D-001.md"))
	if err != nil {
		t.Fatalf("read old D-001: %v", err)
	}
	oldStr := string(oldBytes)

	// Invariant: exactly 2 frontmatter fences (one opening, one closing)
	fenceCount := strings.Count(oldStr, "---")
	if fenceCount != 2 {
		t.Fatalf("expected exactly 2 frontmatter delimiters '---', got %d. Content:\n%s", fenceCount, oldStr)
	}
	if !strings.Contains(oldStr, "status: superseded") {
		t.Errorf("expected status: superseded in old D-001")
	}
	if !strings.Contains(oldStr, "superseded_by: D-002") {
		t.Errorf("expected superseded_by: D-002 in old D-001")
	}
	if !strings.HasPrefix(oldStr, "# Architectural Decision: D-001 Storage Engine") {
		t.Errorf("expected original leading title to be preserved")
	}
}

// TestComprehensive_DAGParsingASCIIArrowsAndUppercaseHeaders tests DAG complexity parsing
// across all ASCII arrow variants (->, -->, =>) and uppercase session headers (### FAD:).
func TestComprehensive_DAGParsingASCIIArrowsAndUppercaseHeaders(t *testing.T) {
	plan := `# Research Pipeline

**Complexity Score:** 15/25 => Tier 2 (8 sessions)

### S1: Foundation Landscape
| **ID** | S1 |
| **Dependencies** | None |

` + "```prompt" + `
Investigate landscape.
` + "```" + `

### FAD: Founding Architecture Synthesis
| **ID** | FAD |
| **Dependencies** | S1 |

` + "```prompt" + `
Synthesize all architecture findings.
` + "```" + `
`
	dag, err := core.ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	if dag.ComplexityScore != 15 {
		t.Errorf("expected complexity score 15, got %d", dag.ComplexityScore)
	}
	if !strings.HasPrefix(dag.Tier, "Tier 2") {
		t.Errorf("expected tier starting with Tier 2, got %q", dag.Tier)
	}
	if len(dag.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(dag.Sessions))
	}
	if dag.Sessions[0].ID != "S1" || dag.Sessions[1].ID != "FAD" {
		t.Errorf("unexpected session IDs: %s, %s", dag.Sessions[0].ID, dag.Sessions[1].ID)
	}
}

// TestComprehensive_StrictTrackBOneWayDoorGateBlocking tests that B3, B4, and B5 defects
// strictly BLOCK One-Way Doors from passing the Phase 0 Exit Gate.
func TestComprehensive_StrictTrackBOneWayDoorGateBlocking(t *testing.T) {
	ctx := context.Background()

	setupWorkspace := func(t *testing.T, sessionRec, adrBody string) (string, Envelope) {
		tmpDir := t.TempDir()
		_, envInit, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir})
		if err != nil || !envInit.Success {
			t.Fatalf("init failed: %v", err)
		}

		pipelineContent := `# Research Pipeline
### T1-01: Core Storage
| **ID** | T1-01 |
| **Dependencies** | None |
` + "```prompt" + `
storage prompt
` + "```" + `

### T1-02: Transport Protocol
| **ID** | T1-02 |
| **Dependencies** | None |
` + "```prompt" + `
transport prompt
` + "```" + `

### SYN-01: Founding Architecture Synthesis
| **ID** | SYN-01 |
| **Dependencies** | T1-01, T1-02 |
` + "```prompt" + `
synthesis prompt
` + "```" + `
`
		_, envPlan, err := handleSavePlan(ctx, nil, SavePlanInput{
			ProjectRoot:      tmpDir,
			Content:          pipelineContent,
			DecisionsContent: "# Decision Registry\n",
		})
		if err != nil || !envPlan.Success {
			t.Fatalf("save plan failed: %v", err)
		}

		// Save session T1-01
		sessContent := `---
session_id: T1-01
title: Core Storage
status: complete
date: 2026-10-07
---

# Key Findings
- Finding 1 [Grade A] https://sqlite.org
- Finding 2 [Grade A] https://tools.ietf.org/html/rfc7636
- Finding 3 [Grade B] Internal benchmark

## Discovered Concerns
Latency spike under heavy write contention.

## Delta
Contradicted assumption that LevelDB is simpler.

` + sessionRec

		_, envSess, err := handleSaveSession(ctx, nil, SaveSessionInput{
			ProjectRoot: tmpDir,
			SessionID:   "T1-01",
			Content:     sessContent,
		})
		if err != nil || !envSess.Success {
			t.Fatalf("save session T1-01 failed: %v", err)
		}

		// Save session T1-02
		sess2Content := `---
session_id: T1-02
title: Transport Protocol
status: complete
date: 2026-10-07
---

# Key Findings
- Transport 1 [Grade A] https://quicwg.org
- Transport 2 [Grade A] https://tools.ietf.org/html/rfc9000
- Transport 3 [Grade B] TCP comparison

## Discovered Concerns
Middlebox UDP blocking.

## Delta
Contradicted assumption that TCP is always faster.

## Recommendation
Use QUIC with TCP fallback. Grade A (https://quicwg.org)
`
		_, envSess2, err := handleSaveSession(ctx, nil, SaveSessionInput{
			ProjectRoot: tmpDir,
			SessionID:   "T1-02",
			Content:     sess2Content,
		})
		if err != nil || !envSess2.Success {
			t.Fatalf("save session T1-02 failed: %v", err)
		}

		// Save SYN-01 FAD
		fadContent := `---
session_id: SYN-01
title: Founding Architecture
status: complete
date: 2026-10-07
---
# Founding Architecture
System uses SQLite for local persistence and QUIC for transport.
Grade A (https://sqlite.org) verified.
`
		_, envFAD, err := handleSaveSession(ctx, nil, SaveSessionInput{
			ProjectRoot: tmpDir,
			SessionID:   "SYN-01",
			Content:     fadContent,
		})
		if err != nil || !envFAD.Success {
			t.Fatalf("save FAD failed: %v", err)
		}

		// Record decision D-001 (One-Way Door)
		adrContent := `---
id: D-001
title: Persistent Datastore
status: accepted
door_type: one-way
date: 2026-10-07
human_reviewed: true
informed_by_sessions:
  - T1-01
review_trigger: "Data exceeds 100 GB"
---
` + adrBody

		_, envADR, err := handleRecordDecision(ctx, nil, RecordDecisionInput{
			ProjectRoot: tmpDir,
			DecisionID:  "D-001",
			Content:     adrContent,
		})
		if err != nil || !envADR.Success {
			t.Fatalf("record decision failed: %v", err)
		}

		// Run gate
		_, envGate, err := handleRunGate(ctx, nil, RunGateInput{
			ProjectRoot: tmpDir,
		})
		if err != nil {
			t.Fatalf("run gate error: %v", err)
		}
		return tmpDir, envGate
	}

	t.Run("B5 missing rejected alternatives blocks one-way door", func(t *testing.T) {
		sessionRec := `## Recommendation
We adopt SQLite. Grade A (https://sqlite.org)
`
		// ADR without rejected alternatives section
		adrBody := `# Context
Embedded storage requirements for high concurrency operations.
# Decision
Adopt SQLite with WAL mode.
# Consequences
Simple deployment.
`
		_, envGate := setupWorkspace(t, sessionRec, adrBody)
		dataMap, _ := envGate.Data.(map[string]any)
		gatePassed, _ := dataMap["gate_passed"].(bool)
		if gatePassed {
			t.Errorf("expected gate_passed to be FALSE due to missing B5 rejected alternatives: %+v", envGate)
		}
		foundB5 := false
		for _, w := range envGate.Warnings {
			if strings.Contains(w, "B5-ALTERNATIVES") {
				foundB5 = true
				break
			}
		}
		if !foundB5 {
			t.Errorf("expected B5-ALTERNATIVES warning, got: %v", envGate.Warnings)
		}
	})

	t.Run("B3 low evidence grade in recommendation blocks one-way door", func(t *testing.T) {
		// Recommendation relies on Grade C evidence
		sessionRec := `## Recommendation
We adopt Custom Store based on Grade C (unverified vendor blog post).
`
		adrBody := `# Context
Storage requirements for high concurrency.
# Decision
Adopt Custom Store.
## Rejected Alternatives
- SQLite: rejected due to concurrency.
`
		_, envGate := setupWorkspace(t, sessionRec, adrBody)
		dataMap, _ := envGate.Data.(map[string]any)
		gatePassed, _ := dataMap["gate_passed"].(bool)
		if gatePassed {
			t.Errorf("expected gate_passed to be FALSE due to B3 Grade C evidence: %+v", envGate)
		}
		foundB3 := false
		for _, w := range envGate.Warnings {
			if strings.Contains(w, "B3-EVIDENTIARY") {
				foundB3 = true
				break
			}
		}
		if !foundB3 {
			t.Errorf("expected B3-EVIDENTIARY warning, got: %v", envGate.Warnings)
		}
	})

	t.Run("B4 recalled evidence in recommendation blocks one-way door", func(t *testing.T) {
		// Recommendation relies on recalled memory
		sessionRec := `## Recommendation
We adopt InfluxDB based on [recalled memory].
`
		adrBody := `# Context
Storage requirements for metrics telemetry.
# Decision
Adopt InfluxDB.
## Rejected Alternatives
- Prometheus: rejected due to retention policies.
`
		_, envGate := setupWorkspace(t, sessionRec, adrBody)
		dataMap, _ := envGate.Data.(map[string]any)
		gatePassed, _ := dataMap["gate_passed"].(bool)
		if gatePassed {
			t.Errorf("expected gate_passed to be FALSE due to B4 recalled evidence: %+v", envGate)
		}
		foundB4 := false
		for _, w := range envGate.Warnings {
			if strings.Contains(w, "B4-VERIFICATION") {
				foundB4 = true
				break
			}
		}
		if !foundB4 {
			t.Errorf("expected B4-VERIFICATION warning, got: %v", envGate.Warnings)
		}
	})

	t.Run("One-Way Door with Grade A evidence and rejected alternatives PASSES", func(t *testing.T) {
		sessionRec := `## Recommendation
We adopt SQLite WAL mode based on Grade A (https://sqlite.org/wal.html).
`
		adrBody := `# Context
Storage requirements for transactional records.
# Decision
Adopt SQLite with WAL mode. Grade A (https://sqlite.org/wal.html)
## Rejected Alternatives
- PostgreSQL: rejected due to daemon resource footprint.
`
		_, envGate := setupWorkspace(t, sessionRec, adrBody)
		dataMap, _ := envGate.Data.(map[string]any)
		gatePassed, _ := dataMap["gate_passed"].(bool)
		if !gatePassed {
			t.Errorf("expected gate_passed to be TRUE, got FALSE (warnings: %v)", envGate.Warnings)
		}
		if dataMap["gate_status"] != "PASS" {
			t.Errorf("expected gate_status to be PASS, got: %v", dataMap["gate_status"])
		}
	})
}

// TestComprehensive_EagerReplanAndRootFADLocks tests eager replan lock acquisition
// and root FAD mirror advisory locking under concurrency.
func TestComprehensive_EagerReplanAndRootFADLocks(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	_, envInit, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir})
	if err != nil || !envInit.Success {
		t.Fatalf("init failed: %v", err)
	}

	pipelineContent := `# Research Pipeline
### T1-01: Session 1
| **ID** | T1-01 |
| **Dependencies** | None |
` + "```prompt" + `
prompt 1
` + "```" + `
`
	_, envPlan, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot:      tmpDir,
		Content:          pipelineContent,
		DecisionsContent: "# Decision Registry\n",
	})
	if err != nil || !envPlan.Success {
		t.Fatalf("save plan failed: %v", err)
	}

	// 1. Verify eager replan mutex prevents race when pipeline is locked
	pipelineFile := filepath.Join(tmpDir, core.PipelineFile)
	unlock, err := store.LockFile(ctx, pipelineFile, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("locking pipeline: %v", err)
	}

	// handleReplan should immediately fail or time out because lock is held externally
	ctxShort, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	_, envReplan, _ := handleReplan(ctxShort, nil, ReplanInput{
		ProjectRoot: tmpDir,
		Operation:   "add_session",
		SessionID:   "T1-02",
		Topic:       "Added Session",
		Prompt:      "prompt 2",
	})
	if envReplan.Success {
		t.Errorf("expected replan to fail while pipeline lock is held, but succeeded: %+v", envReplan)
	}
	_ = unlock()

	// 2. Now replan succeeds with lock released
	_, envReplan2, err := handleReplan(ctx, nil, ReplanInput{
		ProjectRoot: tmpDir,
		Operation:   "add_session",
		SessionID:   "T1-02",
		Topic:       "Added Session",
		Prompt:      "prompt 2",
	})
	if err != nil || !envReplan2.Success {
		t.Fatalf("expected replan to succeed when unlocked: %v, env: %+v", err, envReplan2)
	}

	// 3. Verify synthesis save mirrors FOUNDING-ARCHITECTURE.md to root
	fadContent := `---
session_id: SYN-01
title: Architecture
status: complete
date: 2026-10-07
---
# Founding Architecture
Complete spec.
`
	_, envSaveFAD, err := handleSaveSession(ctx, nil, SaveSessionInput{
		ProjectRoot: tmpDir,
		SessionID:   "SYN-01",
		Content:     fadContent,
	})
	if err != nil || !envSaveFAD.Success {
		t.Fatalf("save FAD failed: %v", err)
	}

	rootFADPath := filepath.Join(tmpDir, "FOUNDING-ARCHITECTURE.md")
	if _, err := os.Stat(rootFADPath); os.IsNotExist(err) {
		t.Fatalf("expected root FOUNDING-ARCHITECTURE.md to exist after synthesis save")
	}

	// 4. Verify amend_session updates root FOUNDING-ARCHITECTURE.md
	amendmentNote := `
### Amendment 2026-10-07
Founding Architecture Updated with additional telemetry notes.
`
	_, envAmendFAD, err := handleAmendSession(ctx, nil, AmendSessionInput{
		ProjectRoot: tmpDir,
		SessionID:   "SYN-01",
		Amendment:   amendmentNote,
	})
	if err != nil || !envAmendFAD.Success {
		t.Fatalf("amend FAD failed: %v", err)
	}

	rootBytes, err := os.ReadFile(rootFADPath)
	if err != nil {
		t.Fatalf("reading amended root FAD: %v", err)
	}
	if !strings.Contains(string(rootBytes), "Founding Architecture Updated") {
		t.Errorf("expected amended content in root FAD, got: %s", string(rootBytes))
	}
}

// TestComprehensive_QualityCoachingCitationDiversityAndAntiGoodhart tests Q-DENSITY and Q-BIAS.
func TestComprehensive_QualityCoachingCitationDiversityAndAntiGoodhart(t *testing.T) {
	// Substantive session (>1000 chars) with only 1 distinct URL triggers Q-DENSITY
	largeSessionSingleURL := strings.Repeat("Detailed domain analysis and architectural tradeoffs. ", 30) + `
## Prior
Initial assumption.
## Key Findings
- Finding 1 [Grade A] https://single-source.com/doc
- Finding 2 [Grade B] benchmark reports from internal test
- Finding 3 [Grade B] community feedback
## Discovered Concerns
Latency spike under contention.
## Delta
Contradicted old belief.
`
	obs := core.ObserveSessionQuality([]byte(largeSessionSingleURL))
	foundDensity := false
	for _, o := range obs {
		if o.Category == "evidence" && strings.Contains(o.Message, "fewer than 2 distinct external sources") {
			foundDensity = true
			break
		}
	}
	if !foundDensity {
		t.Errorf("expected Q-DENSITY citation diversity observation on substantive session with 1 URL")
	}

	// Delta table where all updates start from 'Unknown' triggers Q-BIAS (anti-Goodhart)
	sessionUnknownDelta := `
## Prior
All priors were unknown.
## Key Findings
- Finding 1 [Grade A] https://source1.com
- Finding 2 [Grade A] https://source2.com
- Finding 3 [Grade B]
## Discovered Concerns
Memory consumption.
## Delta
| Belief | Prior | Status | Post-Research |
|---|---|---|---|
| Query Latency | Unknown | Updated | 5ms p99 |
| Concurrency | Unknown | Updated | 1000 QPS |
`
	obs2 := core.ObserveSessionQuality([]byte(sessionUnknownDelta))
	foundBias := false
	for _, o := range obs2 {
		if o.Category == "bias" && strings.Contains(o.Message, "started from 'Unknown'") {
			foundBias = true
			break
		}
	}
	if !foundBias {
		t.Errorf("expected Q-BIAS observation when Delta only updates from 'Unknown' priors")
	}
}

// TestComprehensive_GuidedReplanAndReasonImpactMetadata tests vivechak_replan:
// - Empty operation returns guided overview without mutating pipeline
// - Valid operation with reason and impact modifies pipeline and logs history
func TestComprehensive_GuidedReplanAndReasonImpactMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	_, envInit, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "project"})
	if err != nil || !envInit.Success {
		t.Fatalf("init failed: %v", err)
	}

	pipelineContent := `# Research Pipeline
### T1-01: Session 1
| **ID** | T1-01 |
| **Dependencies** | None |
` + "```prompt\nprompt 1\n```\n"

	_, envPlan, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot:      tmpDir,
		Content:          pipelineContent,
		DecisionsContent: "# Decision Registry\n",
	})
	if err != nil || !envPlan.Success {
		t.Fatalf("save plan failed: %v", err)
	}

	// 1. Guided replan with empty operation
	_, envGuided, err := handleReplan(ctx, nil, ReplanInput{
		ProjectRoot: tmpDir,
	})
	if err != nil || !envGuided.Success {
		t.Fatalf("guided replan failed: %v, env: %+v", err, envGuided)
	}
	guidedData, ok := envGuided.Data.(map[string]any)
	if !ok || guidedData["status"] != "guided" {
		t.Errorf("expected status guided, got: %v", envGuided.Data)
	}
	if !strings.Contains(envGuided.NextStep, "Call vivechak_replan with operation") {
		t.Errorf("expected guided next_step to guide user, got: %q", envGuided.NextStep)
	}

	// Verify pipeline file content was NOT changed by guided overview
	pBytes, _ := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	if strings.Contains(string(pBytes), "REPLAN HISTORY") {
		t.Errorf("pipeline should not have replan history yet")
	}

	// 2. Concrete replan with reason and impact
	_, envReplan, err := handleReplan(ctx, nil, ReplanInput{
		ProjectRoot:  tmpDir,
		Operation:    "add_session",
		SessionID:    "T2-01",
		Topic:        "Security Hardening",
		Prompt:       "Security audit prompt",
		Reason:       "SOC2 compliance requirement",
		Impact:       "Adds required security research session",
		Dependencies: "T1-01",
	})
	if err != nil || !envReplan.Success {
		t.Fatalf("replan failed: %v, env: %+v", err, envReplan)
	}

	replanData := envReplan.Data.(map[string]any)
	if replanData["reason"] != "SOC2 compliance requirement" {
		t.Errorf("expected reason 'SOC2 compliance requirement' in replan data, got: %v", replanData["reason"])
	}
	if replanData["impact"] != "Adds required security research session" {
		t.Errorf("expected impact in replan data, got: %v", replanData["impact"])
	}

	pUpdatedBytes, err := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	if err != nil {
		t.Fatalf("reading updated pipeline: %v", err)
	}
	pUpdatedStr := string(pUpdatedBytes)
	if !strings.Contains(pUpdatedStr, "T2-01") {
		t.Errorf("expected T2-01 in updated pipeline")
	}
}

// TestComprehensive_MultiSessionDecisionChallengePreservation verifies that in multi-session decision scopes
// (e.g. D-001-S1, D-001-S2), challenging a specific session preserves the exact session ID without overwriting it.
func TestComprehensive_MultiSessionDecisionChallengePreservation(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	_, envInit, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "decision"})
	if err != nil || !envInit.Success {
		t.Fatalf("init failed: %v", err)
	}

	// Create session D-001-S1 in research/sessions/
	sessDir := filepath.Join(tmpDir, core.SessionsDir)
	_ = os.MkdirAll(sessDir, 0o755)

	s1Content := `---
session_id: D-001-S1
title: Auth Landscape
status: complete
date: 2026-10-18
---
## Key Findings
- OAuth2 with PKCE is standard [Grade A] https://datatracker.ietf.org/doc/html/rfc7636
## Discovered Concerns
Refresh token rotation complexity.
## Delta
Confirmed PKCE requirement.
`
	if err := os.WriteFile(filepath.Join(sessDir, "D-001-S1.md"), []byte(s1Content), 0o644); err != nil {
		t.Fatalf("writing D-001-S1: %v", err)
	}

	// Challenge D-001-S1
	_, envChal, err := handleChallenge(ctx, nil, ChallengeInput{
		ProjectRoot: tmpDir,
		SessionID:   "D-001-S1",
		Mode:        "red_team",
	})
	if err != nil || !envChal.Success {
		t.Fatalf("challenge D-001-S1 failed: %v, env: %+v", err, envChal)
	}
	chalResult, ok := envChal.Data.(*core.ChallengeResult)
	if !ok {
		t.Fatalf("invalid challenge data type: %T", envChal.Data)
	}
	if chalResult.SessionID != "D-001-S1" {
		t.Errorf("expected session_id 'D-001-S1', got: %v", chalResult.SessionID)
	}
}

// TestComprehensive_SpecialResearchFileExclusionFromADRScan verifies that comparison files
// (e.g. C-01-cmp-auth.md, COMPARISON.md) in research/ are never mistreated as ADRs by record_decision.
func TestComprehensive_SpecialResearchFileExclusionFromADRScan(t *testing.T) {
	tmpDir := t.TempDir()
	researchDir := filepath.Join(tmpDir, core.ResearchDir)
	_ = os.MkdirAll(researchDir, 0o755)
	ctx := context.Background()

	// Write a comparison matrix into research/
	compContent := `# Comparison: JWT vs Sessions
## Weighted Evaluation Protocol (WEP) Matrix
| Option | Score |
|---|---|
| JWT | 4.2 |
`
	if err := os.WriteFile(filepath.Join(researchDir, "C-01-cmp-auth.md"), []byte(compContent), 0o644); err != nil {
		t.Fatalf("write comparison file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(researchDir, "COMPARISON.md"), []byte(compContent), 0o644); err != nil {
		t.Fatalf("write COMPARISON.md: %v", err)
	}

	// Record legitimate decision D-001
	d1Content := `---
id: D-001
title: Auth Strategy
status: accepted
door_type: two-way
date: 2026-10-18
---
# Context
Auth strategy.
# Decision
Adopt session cookies. Grade A (https://owasp.org)
`
	_, envRec, err := handleRecordDecision(ctx, nil, RecordDecisionInput{
		ProjectRoot: tmpDir,
		DecisionID:  "D-001",
		Content:     d1Content,
	})
	if err != nil || !envRec.Success {
		t.Fatalf("record_decision D-001 failed: %v, env: %+v", err, envRec)
	}

	// Inspect DECISIONS.md
	regBytes, err := os.ReadFile(filepath.Join(tmpDir, core.DecisionsFile))
	if err != nil {
		t.Fatalf("read DECISIONS.md: %v", err)
	}
	regStr := string(regBytes)
	if !strings.Contains(regStr, "D-001") {
		t.Errorf("expected D-001 in registry")
	}
	if strings.Contains(regStr, "C-01") || strings.Contains(regStr, "COMPARISON") {
		t.Errorf("comparison files must not be recorded in DECISIONS.md registry: %s", regStr)
	}
}

// TestComprehensive_TrackFilterDecoupledGateClearance verifies that handleRunGate decouples Track A and Track B,
// respects track filtering, and comparison scope receives zero-gate clearance with advisory.
func TestComprehensive_TrackFilterDecoupledGateClearance(t *testing.T) {
	ctx := context.Background()

	t.Run("comparison scope receives zero-gate clearance", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, envInit, _ := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "comparison"})
		if !envInit.Success {
			t.Fatalf("init failed: %s", envInit.Message)
		}

		_, envGate, err := handleRunGate(ctx, nil, RunGateInput{
			ProjectRoot: tmpDir,
		})
		if err != nil {
			t.Fatalf("run_gate failed: %v", err)
		}
		gateData, ok := envGate.Data.(map[string]any)
		if !ok || gateData["gate_passed"] != true {
			t.Errorf("comparison scope must always pass gate, got: %+v", gateData)
		}
		foundZeroReq := false
		for _, adv := range envGate.Warnings {
			if strings.Contains(adv, "zero gate requirements") {
				foundZeroReq = true
				break
			}
		}
		if !foundZeroReq {
			t.Errorf("expected zero gate requirements advisory, got warnings: %v", envGate.Warnings)
		}
	})

	t.Run("project scope track filters", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, envInit, _ := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "project"})
		if !envInit.Success {
			t.Fatalf("init failed: %s", envInit.Message)
		}

		// Invalid track returns error envelope
		_, envInvalid, err := handleRunGate(ctx, nil, RunGateInput{
			ProjectRoot: tmpDir,
			Track:       "invalid_track",
		})
		if err == nil && envInvalid.Success {
			t.Errorf("expected error for invalid track, got success")
		}

		// Track A filter on empty workspace
		_, envTrackA, err := handleRunGate(ctx, nil, RunGateInput{
			ProjectRoot: tmpDir,
			Track:       "track_a",
		})
		if err != nil {
			t.Fatalf("track_a run gate error: %v", err)
		}
		tAData := envTrackA.Data.(map[string]any)
		if tAData["track_filter"] != "track_a" {
			t.Errorf("expected track_filter 'track_a', got: %v", tAData["track_filter"])
		}
	})
}

// TestComprehensive_ComparisonScopeValidateAndVisualize tests ValidateComparison and
// visualize for comparison workspaces.
func TestComprehensive_ComparisonScopeValidateAndVisualize(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	_, envInit, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "comparison"})
	if err != nil || !envInit.Success {
		t.Fatalf("init failed: %v", err)
	}

	wepMatrix := `---
id: C-01-cmp-framework-shootout
title: Fastify vs Express
date: 2026-10-18
status: complete
---

# COMPARISON: Fastify vs Express

## Summary Recommendation
Fastify is recommended for schema-validated routing and throughput based on benchmarks.

## Evaluated Options
- Option 1: Fastify (Grade A - https://fastify.io)
- Option 2: Express (Grade A - https://expressjs.com)

## Weighted Evaluation Protocol (WEP) Matrix
| Criterion | Weight | Fastify (1-5) | Express (1-5) | Evidence |
|---|---|---|---|---|
| Throughput | 0.50 | 5 | 3 | Grade B (benchmarks) |
| Middleware Ecosystem | 0.50 | 4 | 5 | Grade A (npm) |
| **Weighted Total** | **1.00** | **4.50** | **4.00** | |

## Sensitivity Analysis
Varying Throughput weight keeps Fastify ahead.
`
	_, envPlan, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot: tmpDir,
		Scope:       "comparison",
		DecisionID:  "C-01",
		Slug:        "framework-shootout",
		Content:     wepMatrix,
	})
	if err != nil || !envPlan.Success {
		t.Fatalf("save plan failed: %v, env: %+v", err, envPlan)
	}

	// 1. Validate comparison artifact via workspace validation
	_, envVal, err := handleValidate(ctx, nil, ValidateInput{
		ProjectRoot: tmpDir,
	})
	if err != nil || !envVal.Success {
		t.Fatalf("workspace validate failed: %v, env: %+v", err, envVal)
	}
	valData := envVal.Data.(map[string]any)
	if valData["comparisons"] == nil && valData["comparison"] == nil {
		t.Errorf("expected comparison validation in results: %+v", valData)
	}

	// 2. Direct ValidateComparison call
	valDirect := core.ValidateComparison([]byte(wepMatrix))
	if valDirect.HasBlocking() || (valDirect.Status != "valid" && valDirect.Status != "valid-with-warnings") {
		t.Errorf("expected ValidateComparison valid, got: %v (issues: %+v)", valDirect.Status, valDirect.Issues)
	}

	// 3. Visualize comparison workspace
	_, envVis, err := handleVisualize(ctx, nil, VisualizeInput{
		ProjectRoot: tmpDir,
		Format:      "table",
	})
	if err != nil || !envVis.Success {
		t.Fatalf("visualize table failed: %v, env: %+v", err, envVis)
	}
	visData := envVis.Data.(map[string]any)
	if visData["scope"] != core.ScopeComparison {
		t.Errorf("expected scope comparison in visualize data, got: %v", visData["scope"])
	}

	// 4. Visualize comparison workspace as mermaid
	_, envVisM, err := handleVisualize(ctx, nil, VisualizeInput{
		ProjectRoot: tmpDir,
		Format:      "mermaid",
	})
	if err != nil || !envVisM.Success {
		t.Fatalf("visualize mermaid failed: %v, env: %+v", err, envVisM)
	}
}

// TestComprehensive_TrackBArithmeticAndAdversarialChallenge verifies:
// 1. In Project Scope, Track B has 4 indicators (B3, B4, B5, B10).
//    - When all pass, Score is "4/4", Count=4, Total=4, Passed=true.
//    - When B5 (rejected alternatives) is missing, Count=3, Total=4, Score="3/4", Passed=false (Count < Total).
//    - When B3 (evidence grade) is weak, Count=3, Total=4, Score="3/4", Passed=false (Count < Total).
// 2. In Decision Scope, Track B has 3 indicators (B3, B4, B5).
//    - When all pass, Score is "3/3", Count=3, Total=3.
//    - When B3 is weak, Count=2, Total=3, Score="2/3", Passed=false (Count < Total).
// 3. Adversarial Red-Team Challenge (B11):
//    - Emits W-OWD-NO-CHALLENGE advisory when a One-Way Door has not undergone challenge.
//    - Advisory is cleared when the One-Way Door ADR includes red-team challenge sections.
func TestComprehensive_TrackBArithmeticAndAdversarialChallenge(t *testing.T) {
	ctx := context.Background()

	t.Run("Project Scope - Full Pass with Adversarial Challenge produces 4/4 and clears B11", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, envInit, err := handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "project"})
		if err != nil || !envInit.Success {
			t.Fatalf("init failed: %v", err)
		}

		pipelineContent := `# Research Pipeline
### T1-01: Storage
| **ID** | T1-01 |
| **Dependencies** | None |
` + "```prompt\nstorage prompt\n```\n" + `
### T1-02: Transport
| **ID** | T1-02 |
| **Dependencies** | None |
` + "```prompt\ntransport prompt\n```\n" + `
### SYN-01: Synthesis
| **ID** | SYN-01 |
| **Dependencies** | T1-01, T1-02 |
` + "```prompt\nsynthesis prompt\n```\n"

		_, envPlan, err := handleSavePlan(ctx, nil, SavePlanInput{
			ProjectRoot:      tmpDir,
			Content:          pipelineContent,
			DecisionsContent: "# Decisions\n",
		})
		if err != nil || !envPlan.Success {
			t.Fatalf("save plan failed: %v", err)
		}

		sess1Content := `---
session_id: T1-01
title: Storage
status: complete
date: 2026-10-18
---
## Key Findings
- SQLite WAL mode is high performance [Grade A] https://sqlite.org/wal.html
## Discovered Concerns
Concurrency limit on multi-writer transactions.
## Delta
Confirmed WAL mode suitability.
## Recommendation
Adopt SQLite WAL mode based on Grade A evidence (https://sqlite.org/wal.html).
`
		_, envSess1, err := handleSaveSession(ctx, nil, SaveSessionInput{
			ProjectRoot: tmpDir,
			SessionID:   "T1-01",
			Content:     sess1Content,
		})
		if err != nil || !envSess1.Success {
			t.Fatalf("save session T1-01: %v", err)
		}

		sess2Content := `---
session_id: T1-02
title: Transport
status: complete
date: 2026-10-18
---
## Key Findings
- QUIC transport standard [Grade A] https://quicwg.org
## Discovered Concerns
UDP blocking.
## Delta
Confirmed QUIC suitability.
## Recommendation
Use QUIC based on Grade A (https://quicwg.org).
`
		_, envSess2, err := handleSaveSession(ctx, nil, SaveSessionInput{
			ProjectRoot: tmpDir,
			SessionID:   "T1-02",
			Content:     sess2Content,
		})
		if err != nil || !envSess2.Success {
			t.Fatalf("save session T1-02: %v", err)
		}

		fadContent := `---
id: FAD-001
session_id: SYN-01
title: Founding Architecture
synthesis_date: "2026-10-18"
date: "2026-10-18"
status: complete
---
# Founding Architecture Document
This comprehensive architectural blueprint synthesizes all session findings into a unified system design.
All major subsystems, data boundaries, and protocols are rigorously specified [Grade A: official docs https://sqlite.org].
System uses SQLite for local persistence and QUIC for high throughput event transport.
`
		_, envFAD, err := handleSaveSession(ctx, nil, SaveSessionInput{
			ProjectRoot: tmpDir,
			SessionID:   "SYN-01",
			Content:     fadContent,
		})
		if err != nil || !envFAD.Success {
			t.Fatalf("save FAD: %v", err)
		}

		// One-Way Door with Grade A evidence, substantive rejected alternatives, and Red-Team Challenge section
		adrContent := `---
id: D-001
title: Storage Engine
status: accepted
door_type: one-way
date: 2026-10-18
human_reviewed: true
informed_by_sessions:
  - T1-01
review_trigger: "Data exceeds 100 GB"
---
# Context
High performance embedded transactional store.
# Decision
Adopt SQLite with WAL mode. Grade A (https://sqlite.org/wal.html)
## Rejected Alternatives
- PostgreSQL: rejected due to operational daemon footprint and excessive RAM usage for edge devices.
## Red-Team Stress Test
Challenged lock contention under 100 concurrent writers. WAL mode sustains required throughput with busy handler retry.
## Reversal Triggers
- Data size exceeds 100 GB.
`
		_, envADR, err := handleRecordDecision(ctx, nil, RecordDecisionInput{
			ProjectRoot: tmpDir,
			DecisionID:  "D-001",
			Content:     adrContent,
		})
		if err != nil || !envADR.Success {
			t.Fatalf("record decision: %v", err)
		}

		_, envGate, err := handleRunGate(ctx, nil, RunGateInput{ProjectRoot: tmpDir})
		if err != nil || !envGate.Success {
			t.Fatalf("run gate failed: %v", err)
		}

		gateData := envGate.Data.(map[string]any)
		if gateData["gate_passed"] != true {
			t.Fatalf("expected gate_passed true, got false. Warnings: %v", envGate.Warnings)
		}
		if gateData["gate_status"] != "PASS" {
			t.Errorf("expected gate_status 'PASS', got: %v", gateData["gate_status"])
		}

		trackBData := gateData["track_b"].(map[string]any)
		if trackBData["score"] != "4/4" {
			t.Errorf("expected Track B score '4/4', got: %v", trackBData["score"])
		}
		if trackBData["passed"] != true {
			t.Errorf("expected Track B passed true, got: %v", trackBData["passed"])
		}

		// Verify B11 advisory is NOT present because Red-Team Stress Test exists
		for _, w := range envGate.Warnings {
			if strings.Contains(w, "W-OWD-NO-CHALLENGE") {
				t.Errorf("unexpected W-OWD-NO-CHALLENGE advisory when red-team challenge was documented: %s", w)
			}
		}
	})

	t.Run("Project Scope - Missing Challenge produces W-OWD-NO-CHALLENGE advisory but preserves 4/4 score", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, _, _ = handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "project"})

		pContent := "# Research Pipeline\n### T1-01: Storage\n| **ID** | T1-01 |\n| **Dependencies** | None |\n```prompt\nstorage\n```\n### T1-02: Transport\n| **ID** | T1-02 |\n| **Dependencies** | None |\n```prompt\ntransport\n```\n### SYN-01: Synthesis\n| **ID** | SYN-01 |\n| **Dependencies** | T1-01, T1-02 |\n```prompt\nsynthesis\n```\n"
		_, _, _ = handleSavePlan(ctx, nil, SavePlanInput{ProjectRoot: tmpDir, Content: pContent, DecisionsContent: "# Decisions\n"})

		s1Content := "---\nsession_id: T1-01\ntitle: Storage\nstatus: complete\n---\n## Key Findings\n- Finding [Grade A] https://sqlite.org\n## Discovered Concerns\nLocking.\n## Delta\nUpdated.\n## Recommendation\nUse SQLite. Grade A (https://sqlite.org)\n"
		_, _, _ = handleSaveSession(ctx, nil, SaveSessionInput{ProjectRoot: tmpDir, SessionID: "T1-01", Content: s1Content})

		s2Content := "---\nsession_id: T1-02\ntitle: Transport\nstatus: complete\n---\n## Key Findings\n- Transport [Grade A] https://quicwg.org\n## Discovered Concerns\nUDP blocking.\n## Delta\nUpdated.\n## Recommendation\nUse QUIC. Grade A (https://quicwg.org)\n"
		_, _, _ = handleSaveSession(ctx, nil, SaveSessionInput{ProjectRoot: tmpDir, SessionID: "T1-02", Content: s2Content})

		fadContent := `---
id: FAD-001
session_id: SYN-01
title: Founding Architecture
synthesis_date: "2026-10-18"
date: "2026-10-18"
status: complete
---
# Founding Architecture Document
This comprehensive architectural blueprint synthesizes all session findings into an unified system design.
All major subsystems, interfaces, data layers, and security controls are specified with inline evidence grades [Grade A: official docs https://sqlite.org].
`
		_, _, _ = handleSaveSession(ctx, nil, SaveSessionInput{ProjectRoot: tmpDir, SessionID: "SYN-01", Content: fadContent})

		// One-Way Door WITHOUT Red-Team Challenge
		adrContent := `---
id: D-001
title: Storage Engine
status: accepted
door_type: one-way
date: 2026-10-18
human_reviewed: true
informed_by_sessions:
  - T1-01
review_trigger: "Data exceeds 100 GB"
---
# Context
High performance storage.
# Decision
Adopt SQLite with WAL mode. Grade A (https://sqlite.org/wal.html)
## Rejected Alternatives
- PostgreSQL: rejected due to operational daemon footprint.
## Reversal Triggers
- Data size exceeds 100 GB.
`
		_, _, _ = handleRecordDecision(ctx, nil, RecordDecisionInput{ProjectRoot: tmpDir, DecisionID: "D-001", Content: adrContent})

		_, envGate, err := handleRunGate(ctx, nil, RunGateInput{ProjectRoot: tmpDir})
		if err != nil || !envGate.Success {
			t.Fatalf("run gate failed: %v", err)
		}

		gateData := envGate.Data.(map[string]any)
		trackBData := gateData["track_b"].(map[string]any)
		// Core 4 indicators still pass
		if trackBData["score"] != "4/4" {
			t.Errorf("expected score 4/4, got: %v", trackBData["score"])
		}

		// Adversarial challenge advisory MUST be emitted
		foundB11 := false
		for _, w := range envGate.Warnings {
			if strings.Contains(w, "W-OWD-NO-CHALLENGE") {
				foundB11 = true
				break
			}
		}
		if !foundB11 {
			t.Errorf("expected W-OWD-NO-CHALLENGE advisory in warnings: %v", envGate.Warnings)
		}
	})

	t.Run("Decision Scope - Track B Arithmetic shows 3/3 on pass and Count < Total on failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, _, _ = handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir, Scope: "decision"})

		// Session with weak Grade C evidence
		sContent := "---\nsession_id: D-001-S1\ntitle: Auth\nstatus: complete\n---\n## Key Findings\n- Finding [Grade C] personal blog\n## Discovered Concerns\nSecurity.\n## Delta\nUpdated.\n## Recommendation\nUse Custom Auth. Grade C (unverified blog)\n"
		sDir := filepath.Join(tmpDir, core.SessionsDir)
		_ = os.MkdirAll(sDir, 0o755)
		_ = os.WriteFile(filepath.Join(sDir, "D-001-S1.md"), []byte(sContent), 0o644)

		// Decision plan
		_ = os.WriteFile(filepath.Join(tmpDir, core.ResearchDir, "D-001-plan.md"), []byte("# Plan\n"), 0o644)

		// One-Way Door ADR
		adrContent := `---
id: D-001
title: Auth Architecture
status: accepted
door_type: one-way
date: 2026-10-18
informed_by_sessions:
  - D-001-S1
review_trigger: "Breach event occurs"
---
# Context
Authentication.
# Decision
Custom Auth.
## Rejected Alternatives
- Auth0: rejected due to monthly active user pricing curve.
## Reversal Triggers
- Breach event occurs.
`
		_ = os.WriteFile(filepath.Join(tmpDir, core.ResearchDir, "D-001.md"), []byte(adrContent), 0o644)

		ws, err := store.OpenWorkspace(tmpDir)
		if err != nil {
			t.Fatalf("open workspace: %v", err)
		}
		defer ws.Close()

		info := core.InspectWorkspace(tmpDir)
		res := core.EvaluateGate(ws, info, tmpDir)

		// Verification: Track B Total in Decision Scope must be 3 (B3, B4, B5)
		if res.TrackBChecks.Total != 3 {
			t.Errorf("expected TrackB Total 3 in Decision Scope, got: %d", res.TrackBChecks.Total)
		}
		// Because B3 fails (Grade C), Track B must NOT pass, and Count must be strictly less than Total
		if res.TrackBChecks.Passed {
			t.Errorf("expected Track B to fail due to Grade C evidence")
		}
		if res.TrackBChecks.Count >= res.TrackBChecks.Total {
			t.Errorf("arithmetic contradiction: Track B failed but Count (%d) >= Total (%d), Score: %s",
				res.TrackBChecks.Count, res.TrackBChecks.Total, res.TrackBChecks.Score)
		}
		if res.TrackBChecks.Score != "2/3" {
			t.Errorf("expected Track B score '2/3', got: %s", res.TrackBChecks.Score)
		}

		// Now fix the session to Grade A evidence and re-evaluate
		sContentFixed := "---\nsession_id: D-001-S1\ntitle: Auth\nstatus: complete\n---\n## Key Findings\n- Finding [Grade A] https://tools.ietf.org/rfc/rfc7636\n## Discovered Concerns\nPKCE rotation complexity.\n## Delta\nUpdated.\n## Recommendation\nUse OAuth2 with PKCE. Grade A (https://tools.ietf.org/rfc/rfc7636)\n"
		_ = os.WriteFile(filepath.Join(sDir, "D-001-S1.md"), []byte(sContentFixed), 0o644)

		resFixed := core.EvaluateGate(ws, info, tmpDir)
		if !resFixed.TrackBChecks.Passed {
			t.Errorf("expected Track B to pass after fixing evidence, got issues: %v", resFixed.TrackBChecks.Issues)
		}
		if resFixed.TrackBChecks.Score != "3/3" {
			t.Errorf("expected Track B score '3/3' on full pass, got: %s", resFixed.TrackBChecks.Score)
		}
		if resFixed.TrackBChecks.Count != 3 || resFixed.TrackBChecks.Total != 3 {
			t.Errorf("expected Count=3 and Total=3, got Count=%d, Total=%d", resFixed.TrackBChecks.Count, resFixed.TrackBChecks.Total)
		}
	})
}

// TestComprehensive_ScopeInferenceAndPrepositionStrippingMCP verifies that live MCP tool calls
// accurately infer scope using 3-tier precedence and cleanly strip prepositional phrases.
func TestComprehensive_ScopeInferenceAndPrepositionStrippingMCP(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	cases := []struct {
		name           string
		inputContext   string
		expectedScope  string
		checkPromptSub string
	}{
		{
			name:           "Pino vs Winston with 'on' preposition phrase",
			inputContext:   "Pino vs Winston on Node 22",
			expectedScope:  "comparison",
			checkPromptSub: "OPTIONS: Pino, Winston",
		},
		{
			name:           "Architectural dilemma with 'vs' routes to Decision scope",
			inputContext:   "should we use kafka vs rabbitmq for our async notifications pipeline",
			expectedScope:  "decision",
			checkPromptSub: "Decision Research Plan",
		},
		{
			name:           "Colloquial prompt with slang/typos routes to Comparison scope",
			inputContext:   "hey im confused bw winston and pino for my express app which one shud i use?? need fast json logging pls help",
			expectedScope:  "comparison",
			checkPromptSub: "Comparison Research Prompt",
		},
		{
			name:           "Migration question with slang routes to Decision scope",
			inputContext:   "shuld we migrate our event storage from mysql to clickhouse",
			expectedScope:  "decision",
			checkPromptSub: "Decision Research Plan",
		},
		{
			name:           "Frontend framework dilemma with slang routes to Decision scope",
			inputContext:   "bro shuld i use react 19 server actions or tanstack react query for fetching data in my nextjs dashboard? im totally confused",
			expectedScope:  "decision",
			checkPromptSub: "Decision Research Plan",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_prepare_generator",
				Arguments: map[string]any{
					"project_root": tmpDir,
					"context":      tc.inputContext,
				},
			})
			if err != nil {
				t.Fatalf("prepare_generator failed: %v", err)
			}
			env := parseEnvelope(t, res)
			if !env.Success {
				t.Fatalf("prepare_generator returned error: %s", env.Message)
			}
			data := env.Data.(map[string]any)
			if data["scope"] != tc.expectedScope {
				t.Errorf("context %q: expected scope %q, got %q", tc.inputContext, tc.expectedScope, data["scope"])
			}
			prompt, _ := data["prompt"].(string)
			if !strings.Contains(prompt, tc.checkPromptSub) {
				t.Errorf("context %q: expected prompt to contain %q, got: %s", tc.inputContext, tc.checkPromptSub, prompt[:300])
			}
		})
	}
}

// TestComprehensive_WorkspaceAnchoringAcrossCalls verifies that when project_root is passed
// once to vivechak_prepare_generator, subsequent tool calls that omit project_root inherit
// the anchored workspace from session memory.
func TestComprehensive_WorkspaceAnchoringAcrossCalls(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Prepare generator with explicit project_root
	res1, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_prepare_generator",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "comparison",
			"context":      "Pino vs Winston on Node 22",
		},
	})
	if err != nil {
		t.Fatalf("prepare_generator: %v", err)
	}
	env1 := parseEnvelope(t, res1)
	if !env1.Success {
		t.Fatalf("prepare_generator returned error: %s", env1.Message)
	}

	// Verify project_root was anchored in session memory
	lastResolved := core.GetLastResolvedWorkspace()
	if !strings.EqualFold(lastResolved, tmpDir) {
		t.Fatalf("core.GetLastResolvedWorkspace() = %q, want %q", lastResolved, tmpDir)
	}

	// 2. Call vivechak_save_plan with project_root OMITTED
	wepContent := `# Comparison: Pino vs Winston
| Metric | Pino | Winston |
|---|---|---|
| Speed | 5 | 3 |
`
	res2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"scope":       "comparison",
			"decision_id": "C-01",
			"slug":        "pino-vs-winston",
			"content":     wepContent,
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}
	env2 := parseEnvelope(t, res2)
	if !env2.Success {
		t.Fatalf("save_plan without project_root failed: %s", env2.Message)
	}

	// Verify plan was saved inside the anchored tmpDir
	expectedPlanPath := filepath.Join(tmpDir, core.ResearchDir, "C-01-cmp-pino-vs-winston.md")
	if _, err := os.Stat(expectedPlanPath); os.IsNotExist(err) {
		t.Fatalf("expected plan file to exist at anchored path %s, but not found", expectedPlanPath)
	}

	// 3. Call vivechak_status with project_root OMITTED
	res3, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_status",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	env3 := parseEnvelope(t, res3)
	if !env3.Success {
		t.Fatalf("status without project_root failed: %s", env3.Message)
	}
	data3 := env3.Data.(map[string]any)
	if data3["scope"] != "comparison" {
		t.Errorf("expected scope 'comparison' in status data, got: %v", data3["scope"])
	}
}

// TestComprehensive_GateStatusTextAndTransitions tests that gate_status and next_step
// accurately transition through the state machine:
// FAIL (structural failure) -> WARN (structural pass, quality failure) -> PASS (both pass).
func TestComprehensive_GateStatusTextAndTransitions(t *testing.T) {
	ctx := context.Background()

	// State 1: FAIL — Empty workspace has structural failures
	tmpDir1 := t.TempDir()
	_, _, _ = handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir1, Scope: "project"})
	_, envGate1, _ := handleRunGate(ctx, nil, RunGateInput{ProjectRoot: tmpDir1})
	data1 := envGate1.Data.(map[string]any)
	if data1["gate_status"] != "FAIL" {
		t.Errorf("State 1: expected gate_status 'FAIL', got: %v", data1["gate_status"])
	}
	if data1["gate_passed"] != false {
		t.Errorf("State 1: expected gate_passed false, got: %v", data1["gate_passed"])
	}
	if !strings.Contains(envGate1.NextStep, "⛔ DO NOT PROCEED TO CODE") {
		t.Errorf("State 1: expected blocking next_step, got: %q", envGate1.NextStep)
	}

	// State 2: WARN — Structural pass, but Quality checks fail (One-Way Door missing B5 rejected alternatives)
	tmpDir2 := t.TempDir()
	_, _, _ = handleInit(ctx, nil, InitInput{ProjectRoot: tmpDir2, Scope: "project"})
	pContent := "# Research Pipeline\n### T1-01: S1\n| **ID** | T1-01 |\n| **Dependencies** | None |\n```prompt\np1\n```\n### T1-02: S2\n| **ID** | T1-02 |\n| **Dependencies** | None |\n```prompt\np2\n```\n### SYN-01: FAD\n| **ID** | SYN-01 |\n| **Dependencies** | T1-01, T1-02 |\n```prompt\np3\n```\n"
	_, _, _ = handleSavePlan(ctx, nil, SavePlanInput{ProjectRoot: tmpDir2, Content: pContent, DecisionsContent: "# Decisions\n"})
	s1Content := "---\nsession_id: T1-01\ntitle: S1\nstatus: complete\n---\n## Key Findings\n- Finding [Grade A] https://sqlite.org\n## Discovered Concerns\nWrite contention.\n## Delta\nUpdated.\n## Recommendation\nUse SQLite. Grade A (https://sqlite.org)\n"
	_, _, _ = handleSaveSession(ctx, nil, SaveSessionInput{ProjectRoot: tmpDir2, SessionID: "T1-01", Content: s1Content})
	s2Content := "---\nsession_id: T1-02\ntitle: S2\nstatus: complete\n---\n## Key Findings\n- Finding 2 [Grade A] https://quicwg.org\n## Discovered Concerns\nUDP blocking.\n## Delta\nUpdated.\n## Recommendation\nUse QUIC. Grade A (https://quicwg.org)\n"
	_, _, _ = handleSaveSession(ctx, nil, SaveSessionInput{ProjectRoot: tmpDir2, SessionID: "T1-02", Content: s2Content})
	fadContent := `---
id: FAD-001
session_id: SYN-01
title: Founding Architecture Document
synthesis_date: "2026-10-18"
date: "2026-10-18"
status: complete
---
# Founding Architecture Document
This comprehensive architectural blueprint synthesizes all session findings into an unified system design.
All major subsystems, data boundaries, and protocols are rigorously specified [Grade A: official docs https://sqlite.org].
`
	_, _, _ = handleSaveSession(ctx, nil, SaveSessionInput{ProjectRoot: tmpDir2, SessionID: "SYN-01", Content: fadContent})
	// One-Way Door missing B5 (rejected alternatives)
	adrBad := `---
id: D-001
title: Storage
status: accepted
door_type: one-way
date: 2026-10-18
informed_by_sessions:
  - T1-01
review_trigger: "Data > 100 GB"
---
# Context
Storage choice.
# Decision
Adopt SQLite. Grade A (https://sqlite.org)
## Reversal Triggers
- Data > 100 GB.
`
	_, _, _ = handleRecordDecision(ctx, nil, RecordDecisionInput{ProjectRoot: tmpDir2, DecisionID: "D-001", Content: adrBad})

	ws2, err := store.OpenWorkspace(tmpDir2)
	if err != nil {
		t.Fatalf("open workspace: %v", err)
	}
	defer ws2.Close()
	info2 := core.InspectWorkspace(tmpDir2)
	res2 := core.EvaluateGate(ws2, info2, tmpDir2)

	if !res2.StructuralChecks.Passed {
		t.Errorf("State 2: expected StructuralChecks to pass, got: %v", res2.StructuralChecks.Issues)
	}
	if res2.TrackBChecks.Passed {
		t.Errorf("State 2: expected TrackBChecks to fail due to missing B5")
	}
	if res2.Status != "WARN" {
		t.Errorf("State 2: expected core GateResult.Status to be 'WARN', got: %s", res2.Status)
	}
	if res2.Passed != false {
		t.Errorf("State 2: expected core GateResult.Passed to be false, got: %v", res2.Passed)
	}
	if !strings.Contains(res2.NextStep, "Gate quality checks incomplete") {
		t.Errorf("State 2: expected NextStep to warn about quality checks, got: %q", res2.NextStep)
	}

	// State 3: PASS — Both Structural and Quality pass
	adrGood := `---
id: D-001
title: Storage
status: accepted
door_type: one-way
date: 2026-10-18
informed_by_sessions:
  - T1-01
review_trigger: "Data > 100 GB"
---
# Context
Storage choice.
# Decision
Adopt SQLite. Grade A (https://sqlite.org)
## Rejected Alternatives
- PostgreSQL: rejected due to daemon resource footprint and high operational complexity.
## Reversal Triggers
- Data > 100 GB.
`
	_, _, _ = handleRecordDecision(ctx, nil, RecordDecisionInput{ProjectRoot: tmpDir2, DecisionID: "D-001", Content: adrGood})
	res3 := core.EvaluateGate(ws2, info2, tmpDir2)
	if !res3.Passed {
		t.Errorf("State 3: expected core GateResult.Passed to be true, got false. Issues: %v", res3.QualityChecks.Issues)
	}
	if res3.Status != "PASS" {
		t.Errorf("State 3: expected core GateResult.Status to be 'PASS', got: %s", res3.Status)
	}
	if !strings.Contains(res3.NextStep, "Gate PASSED!") {
		t.Errorf("State 3: expected NextStep to celebrate Gate PASSED, got: %q", res3.NextStep)
	}
}



