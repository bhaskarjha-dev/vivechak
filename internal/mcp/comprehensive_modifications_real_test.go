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

// TestComprehensive_SavePlan_MultiDecisionRegistry_FullLifecycle verifies that save_plan
// handles a full multi-decision registry with proposed decisions without spurious warnings,
// while correctly surfacing genuine errors tagged by decision ID.
func TestComprehensive_SavePlan_MultiDecisionRegistry_FullLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	// 1. Init workspace
	_, envInit, err := handleInit(ctx, nil, InitInput{
		ProjectRoot: tmpDir,
		Scope:       "project",
	})
	if err != nil || !envInit.Success {
		t.Fatalf("init failed: %v, message: %s", err, envInit.Message)
	}

	// 2. Prepare generator prompt and check nextStep guidance
	_, envPrep, err := handlePrepareGenerator(ctx, nil, PrepareGeneratorInput{
		ProjectRoot: tmpDir,
		Scope:       "project",
		Context:     "Build high-performance real-time telemetry analytics platform in Go",
	})
	if err != nil || !envPrep.Success {
		t.Fatalf("prepare_generator failed: %v", err)
	}
	if !strings.Contains(envPrep.NextStep, "generate your research pipeline") {
		t.Errorf("expected NextStep to contain 'generate your research pipeline', got: %s", envPrep.NextStep)
	}
	if !strings.Contains(envPrep.NextStep, "Do NOT write to local artifact/brain files") {
		t.Errorf("expected NextStep to explicitly forbid brain artifact writes, got: %s", envPrep.NextStep)
	}

	// 3. Save plan with multi-decision registry containing valid proposed decisions and one invalid decision
	pipelineContent := `# Research Pipeline: Telemetry Platform
**Complexity Score:** 12 / 24 → **Tier 2 (9–16 Sessions)**

## Sessions
#### T0-01: Ingestion Protocols
| Field | Value |
|---|---|
| ID | T0-01 |
| Layer | 0 |
| Door Type | Two-Way |
| Dependencies | None |

` + "```prompt" + `
Investigate gRPC vs WebSocket for high-frequency telemetry.
` + "```" + `

---

#### T0-02: Storage Layer
| Field | Value |
|---|---|
| ID | T0-02 |
| Layer | 0 |
| Door Type | One-Way |
| Dependencies | None |

` + "```prompt" + `
Investigate ClickHouse vs TimescaleDB for telemetry timeseries.
` + "```" + `

---

#### SYN-01: Final Architecture Synthesis
| Field | Value |
|---|---|
| ID | SYN-01 |
| Layer | 1 |
| Door Type | One-Way |
| Dependencies | ALL prior sessions (hard) |

` + "```prompt" + `
Synthesize all telemetry architecture findings into FAD.
` + "```" + `
`

	decisionsContentWithDefect := `# Architectural Decision Records

<!-- DECISION: D-001 -->
---
id: D-001
title: "Ingestion Transport Protocol"
status: proposed
door_type: two-way
date: 2026-10-20
---

# D-001: Ingestion Transport Protocol
## Initial Context & Hypotheses
Evaluate gRPC streaming against WebSockets.
<!-- /DECISION: D-001 -->

<!-- DECISION: D-002 -->
---
id: D-002
title: "Timeseries Storage Engine"
status: proposed
door_type: one-way
date: 2026-10-20
---

# D-002: Timeseries Storage Engine
## Initial Context & Hypotheses
ClickHouse column-store vs TimescaleDB.
<!-- /DECISION: D-002 -->

<!-- DECISION: D-003 -->
---
id: D-003
status: proposed
door_type: one-way
date: 2026-10-20
---

# D-003: Query Engine Strategy
## Initial Context & Hypotheses
Missing title field in frontmatter above.
<!-- /DECISION: D-003 -->
`

	_, envSave, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot:      tmpDir,
		Scope:            "project",
		Content:          pipelineContent,
		DecisionsContent: decisionsContentWithDefect,
	})
	if err != nil {
		t.Fatalf("save_plan call error: %v", err)
	}
	if !envSave.Success {
		t.Fatalf("expected save_plan to succeed with warnings, got: %s", envSave.Message)
	}

	// Verify that D-001 and D-002 did not produce spurious warnings, and D-003 is tagged with ID
	foundD003MissingTitle := false
	for _, w := range envSave.Warnings {
		if strings.Contains(w, "DECISION D-003: [L2-BLOCK] V-MISSING-FIELD: Required field \"title\" missing") {
			foundD003MissingTitle = true
		}
		if strings.Contains(w, "DECISION D-001") || strings.Contains(w, "DECISION D-002") {
			t.Errorf("spurious warning on valid decisions: %s", w)
		}
		if strings.Contains(w, "W-NO-EVIDENCE-GRADES") {
			t.Errorf("spurious evidence grade warning on proposed decisions: %s", w)
		}
	}
	if !foundD003MissingTitle {
		t.Errorf("expected DECISION D-003 missing title warning, warnings were: %v", envSave.Warnings)
	}

	// 4. Now fix D-003 and save again — verify 0 warnings
	decisionsContentClean := strings.Replace(decisionsContentWithDefect, "id: D-003\nstatus: proposed", "id: D-003\ntitle: \"Query Engine Strategy\"\nstatus: proposed", 1)
	_, envSaveClean, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot:      tmpDir,
		Scope:            "project",
		Content:          pipelineContent,
		DecisionsContent: decisionsContentClean,
	})
	if err != nil {
		t.Fatalf("clean save_plan call error: %v", err)
	}
	if len(envSaveClean.Warnings) != 0 {
		t.Errorf("expected 0 warnings on clean save_plan, got: %v", envSaveClean.Warnings)
	}
}

// TestComprehensive_DAG_AllPriorSessions_ComplexExpansion verifies multi-layer dependency expansion
// and DAG scheduling invariance when multiple sessions use natural language dependencies.
func TestComprehensive_DAG_AllPriorSessions_ComplexExpansion(t *testing.T) {
	pipelineMD := `# Multi-Layer Research Pipeline
## Sessions
#### T0-01: Base 1
| Field | Value |
|---|---|
| ID | T0-01 |
| Layer | 0 |
| Dependencies | None |

` + "```prompt" + `
Prompt T0-01
` + "```" + `

---

#### T0-02: Base 2
| Field | Value |
|---|---|
| ID | T0-02 |
| Layer | 0 |
| Dependencies | None |

` + "```prompt" + `
Prompt T0-02
` + "```" + `

---

#### T1-01: Derived 1
| Field | Value |
|---|---|
| ID | T1-01 |
| Layer | 1 |
| Dependencies | T0-01 |

` + "```prompt" + `
Prompt T1-01
` + "```" + `

---

#### T2-01: Cross-Layer Bridge
| Field | Value |
|---|---|
| ID | T2-01 |
| Layer | 2 |
| Dependencies | all prior sessions |

` + "```prompt" + `
Prompt T2-01
` + "```" + `

---

#### SYN-01: Terminal Architecture
| Field | Value |
|---|---|
| ID | SYN-01 |
| Layer | 3 |
| Dependencies | ALL prior sessions (hard) |

` + "```prompt" + `
Prompt SYN-01
` + "```" + `
`

	dag, err := core.ParsePipeline([]byte(pipelineMD))
	if err != nil {
		t.Fatalf("ParsePipeline error: %v", err)
	}

	// T2-01 should depend on T0-01, T0-02, T1-01
	t2 := dag.SessionByID("T2-01")
	if t2 == nil {
		t.Fatal("T2-01 not found")
	}
	expectedT2 := []string{"T0-01", "T0-02", "T1-01"}
	if len(t2.Dependencies) != len(expectedT2) {
		t.Fatalf("T2-01 dependencies = %v; want %v", t2.Dependencies, expectedT2)
	}

	// SYN-01 should depend on all 4 preceding sessions
	syn := dag.SessionByID("SYN-01")
	if syn == nil {
		t.Fatal("SYN-01 not found")
	}
	expectedSYN := []string{"T0-01", "T0-02", "T1-01", "T2-01"}
	if len(syn.Dependencies) != len(expectedSYN) {
		t.Fatalf("SYN-01 dependencies = %v; want %v", syn.Dependencies, expectedSYN)
	}

	// Validate DAG structure
	if err := dag.ValidateDAG(); err != nil {
		t.Fatalf("ValidateDAG failed: %v", err)
	}

	// Initial NextSessions should return T0-01 and T0-02 only
	completed := make(map[string]bool)
	next := dag.NextSessions(completed)
	if len(next) != 2 {
		t.Fatalf("expected 2 initial sessions, got %d: %v", len(next), next)
	}
	if next[0].ID != "T0-01" && next[0].ID != "T0-02" {
		t.Errorf("unexpected initial session: %s", next[0].ID)
	}

	// Complete T0-01
	completed["T0-01"] = true
	next = dag.NextSessions(completed)
	// Now T0-02 and T1-01 should be actionable
	foundT1 := false
	for _, s := range next {
		if s.ID == "T1-01" {
			foundT1 = true
		}
	}
	if !foundT1 {
		t.Errorf("expected T1-01 to unlock after T0-01 completed, got: %v", next)
	}

	// T2-01 and SYN-01 must NOT unlock yet
	for _, s := range next {
		if s.ID == "T2-01" || s.ID == "SYN-01" {
			t.Errorf("session %s unlocked prematurely: %v", s.ID, next)
		}
	}

	// Complete all non-synthesis sessions
	completed["T0-02"] = true
	completed["T1-01"] = true
	completed["T2-01"] = true

	// Now SYN-01 must unlock as terminal session
	next = dag.NextSessions(completed)
	if len(next) != 1 || next[0].ID != "SYN-01" {
		t.Errorf("expected SYN-01 to unlock as final session, got: %v", next)
	}
}

// TestComprehensive_ValidateDecision_StatusProposedVsAccepted verifies the evidentiary
// validation lifecycle across proposed, accepted, and superseded states.
func TestComprehensive_ValidateDecision_StatusProposedVsAccepted(t *testing.T) {
	// 1. Proposed ADR without evidence grades — must NOT produce W-NO-EVIDENCE-GRADES
	proposedADR := `---
id: D-001
title: "Data Bus Architecture"
status: proposed
door_type: one-way
date: 2026-10-20
---
# D-001: Data Bus Architecture
## Context & Hypotheses
We hypothesize Apache Pulsar or NATS JetStream will meet throughput requirements.
`
	vProposed := core.ValidateDecision([]byte(proposedADR))
	for _, iss := range vProposed.Issues {
		if iss.Code == "W-NO-EVIDENCE-GRADES" {
			t.Errorf("proposed ADR produced spurious W-NO-EVIDENCE-GRADES: %v", iss)
		}
	}

	// 2. Accepted ADR without evidence grades — MUST produce W-NO-EVIDENCE-GRADES
	acceptedADRNoGrades := `---
id: D-001
title: "Data Bus Architecture"
status: accepted
door_type: one-way
date: 2026-10-20
---
# D-001: Data Bus Architecture
## Decision Outcome
We choose NATS JetStream.
## Consequences
Low latency and operational simplicity.
`
	vAcceptedNoGrades := core.ValidateDecision([]byte(acceptedADRNoGrades))
	foundNoGrades := false
	for _, iss := range vAcceptedNoGrades.Issues {
		if iss.Code == "W-NO-EVIDENCE-GRADES" {
			foundNoGrades = true
			break
		}
	}
	if !foundNoGrades {
		t.Errorf("expected accepted ADR without evidence grades to produce W-NO-EVIDENCE-GRADES")
	}

	// 3. Accepted ADR with valid evidence grades — must be valid
	acceptedADRWithGrades := `---
id: D-001
title: "Data Bus Architecture"
status: accepted
door_type: one-way
date: 2026-10-20
---
# D-001: Data Bus Architecture
## Decision Outcome
We choose NATS JetStream. Grade A (https://nats.io/benchmarks)
## Consequences
Low latency and operational simplicity.
`
	vAcceptedWithGrades := core.ValidateDecision([]byte(acceptedADRWithGrades))
	for _, iss := range vAcceptedWithGrades.Issues {
		if iss.Code == "W-NO-EVIDENCE-GRADES" {
			t.Errorf("unexpected W-NO-EVIDENCE-GRADES for grounded accepted ADR: %v", iss)
		}
	}
}

// TestComprehensive_SavePlan_MultiDecision_MixedStatuses verifies that save_plan accurately
// isolates evidence warnings to only accepted decisions lacking grounding in a multi-ADR registry.
func TestComprehensive_SavePlan_MultiDecision_MixedStatuses(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	pipelineContent := `# Research Pipeline
### T0-01: Foundation
| **ID** | T0-01 |
| **Dependencies** | None |
` + "```prompt\nfoundation prompt\n```\n"

	multiADR := `# Decision Registry

### D-001: Initial Ingestion Architecture
---
id: D-001
title: "Initial Ingestion Architecture"
status: proposed
door_type: one-way
date: 2026-10-20
---
# Context
We hypothesize gRPC streaming will satisfy ingestion.

---

### D-002: Storage Layer Engine
---
id: D-002
title: "Storage Layer Engine"
status: accepted
door_type: one-way
date: 2026-10-20
---
# Context
We adopt PostgreSQL with TimescaleDB. (No evidence grades here)

---

### D-003: Messaging Backbone
---
id: D-003
title: "Messaging Backbone"
status: accepted
door_type: one-way
date: 2026-10-20
---
# Context
We adopt Apache Kafka. Grade A (https://kafka.apache.org/documentation)
`

	_, envSave, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot:      tmpDir,
		Scope:            "project",
		Content:          pipelineContent,
		DecisionsContent: multiADR,
	})
	if err != nil {
		t.Fatalf("handleSavePlan error: %v", err)
	}
	if !envSave.Success {
		t.Fatalf("expected handleSavePlan to succeed, got: %s", envSave.Message)
	}

	// Should have warning for D-002 only
	foundD2Warn := false
	foundD1Warn := false
	foundD3Warn := false
	for _, w := range envSave.Warnings {
		if strings.Contains(w, "DECISION D-002") && strings.Contains(w, "W-NO-EVIDENCE-GRADES") {
			foundD2Warn = true
		}
		if strings.Contains(w, "DECISION D-001") {
			foundD1Warn = true
		}
		if strings.Contains(w, "DECISION D-003") {
			foundD3Warn = true
		}
	}

	if !foundD2Warn {
		t.Errorf("expected DECISION D-002 warning for ungrounded accepted decision, got warnings: %v", envSave.Warnings)
	}
	if foundD1Warn {
		t.Errorf("unexpected warning for proposed D-001: %v", envSave.Warnings)
	}
	if foundD3Warn {
		t.Errorf("unexpected warning for grounded accepted D-003: %v", envSave.Warnings)
	}
}

// TestComprehensive_DAG_AllPriorSessions_SerializeAndVisualizeRoundTrip tests
// full serialization, round-trip re-parsing, and Mermaid visualization for pipelines
// containing natural language 'all prior sessions' dependency specifications.
func TestComprehensive_DAG_AllPriorSessions_SerializeAndVisualizeRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	pipelineContent := `# Multi-Phase Pipeline
### T0-01: Research Base 1
| **ID** | T0-01 |
| **Dependencies** | None |
` + "```prompt\nBase 1 prompt\n```\n" + `
---
### T0-02: Research Base 2
| **ID** | T0-02 |
| **Dependencies** | None |
` + "```prompt\nBase 2 prompt\n```\n" + `
---
### SYN-01: Architecture Synthesis
| **ID** | SYN-01 |
| **Dependencies** | ALL prior sessions (hard) |
` + "```prompt\nSynthesis prompt\n```\n"

	_, envSave, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot: tmpDir,
		Scope:       "project",
		Content:     pipelineContent,
	})
	if err != nil || !envSave.Success {
		t.Fatalf("save plan failed: %v", err)
	}

	// Visualize pipeline via handleVisualize
	resVis, envVis, err := handleVisualize(ctx, nil, VisualizeInput{
		ProjectRoot: tmpDir,
	})
	if err != nil || resVis.IsError {
		t.Fatalf("handleVisualize failed: %v, env: %+v", err, envVis)
	}

	visData, ok := envVis.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map in visualize envelope data, got: %T", envVis.Data)
	}
	mermaidDiagram, ok := visData["mermaid"].(string)
	if !ok {
		t.Fatalf("expected mermaid string in data: %v", visData)
	}

	// Verify diagram contains directed edges from each prior session to SYN-01
	if !strings.Contains(mermaidDiagram, "T0-01 --> SYN-01") {
		t.Errorf("expected edge T0-01 --> SYN-01 in Mermaid diagram, got:\n%s", mermaidDiagram)
	}
	if !strings.Contains(mermaidDiagram, "T0-02 --> SYN-01") {
		t.Errorf("expected edge T0-02 --> SYN-01 in Mermaid diagram, got:\n%s", mermaidDiagram)
	}
}

// TestComprehensive_PrepareGenerator_AntiLocalSpillGuidance_AllScopes tests that all three
// scopes include unequivocal guidance directing agents to call vivechak_save_plan rather
// than leaking raw plans to local scratchpads or brain files.
func TestComprehensive_PrepareGenerator_AntiLocalSpillGuidance_AllScopes(t *testing.T) {
	ctx := context.Background()

	scopes := []string{"project", "decision", "comparison"}
	for _, sc := range scopes {
		t.Run(sc, func(t *testing.T) {
			tmpDir := t.TempDir()
			_, envPrep, err := handlePrepareGenerator(ctx, nil, PrepareGeneratorInput{
				ProjectRoot: tmpDir,
				Scope:       sc,
				Context:     "High-throughput event streaming platform",
			})
			if err != nil || !envPrep.Success {
				t.Fatalf("prepare generator (%s) failed: %v, env: %+v", sc, err, envPrep)
			}

			if !strings.Contains(envPrep.NextStep, "vivechak_save_plan") {
				t.Errorf("scope %s next_step missing vivechak_save_plan: %s", sc, envPrep.NextStep)
			}
			if !strings.Contains(envPrep.NextStep, "Do NOT write to local artifact/brain files") {
				t.Errorf("scope %s next_step missing anti-local spill directive: %s", sc, envPrep.NextStep)
			}
		})
	}
}

// TestComprehensive_MultiDecisionRegistry_FullLifecycle_And_DetailsFrontmatter comprehensively verifies
// multi-decision registry validation, collapsible <details> blocks with both ```yaml and --- frontmatter,
// status synchronization, error reporting on malformed YAML, and workspace mode ADR discovery.
func TestComprehensive_MultiDecisionRegistry_FullLifecycle_And_DetailsFrontmatter(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Initialize project scope
	_, envInit, err := handleInit(ctx, nil, InitInput{
		ProjectRoot: tmpDir,
		Scope:       "project",
	})
	if err != nil || !envInit.Success {
		t.Fatalf("init failed: %v", err)
	}

	// 2. Prepare generator and verify Tier 3 scratch staging guidance
	_, envPrep, err := handlePrepareGenerator(ctx, nil, PrepareGeneratorInput{
		ProjectRoot: tmpDir,
		Scope:       "project",
		Context:     "Enterprise career recommendation engine with distributed graph store",
	})
	if err != nil || !envPrep.Success {
		t.Fatalf("prepare generator failed: %v", err)
	}
	if !strings.Contains(envPrep.NextStep, "For large Tier 2/3 pipelines where output exceeds single-turn token limits") {
		t.Errorf("expected Tier 2/3 scratch staging guidance in nextStep: %s", envPrep.NextStep)
	}

	// 3. Construct multi-decision registry with varied formats:
	// - D-01: Anchored <details> block with ```yaml
	// - D-02: Anchored <details> block with --- frontmatter
	// - D-03: Unanchored decision block with markdown header and --- frontmatter
	registryContent := `# Architectural Decisions Log
This log captures all founding architecture decisions for the enterprise career platform.

| Decision | Title | Door Type | Status | Date |
|---|---|---|---|---|
| D-01 | Primary Database | one-way | proposed | 2026-10-21 |
| D-02 | Cache & In-Memory Store | two-way | proposed | 2026-10-21 |
| D-03 | Graph Engine Selection | one-way | proposed | 2026-10-21 |

<!-- DECISION: D-01 -->
<details>
<summary><b>D-01: Primary Database</b></summary>

` + "```yaml" + `
id: D-01
title: Primary Database
door_type: one-way
status: proposed
date: 2026-10-21
` + "```" + `

### Context
We evaluated CockroachDB and PostgreSQL.
</details>
<!-- /DECISION: D-01 -->

---

<!-- DECISION: D-02 -->
<details>
<summary><b>D-02: Cache & In-Memory Store</b></summary>

---
id: D-02
title: Cache & In-Memory Store
door_type: two-way
status: proposed
date: 2026-10-21
---

### Context
Redis vs Dragonfly for ephemeral user sessions.
</details>
<!-- /DECISION: D-02 -->

---

### D-03: Graph Engine Selection
---
id: D-03
title: Graph Engine Selection
door_type: one-way
status: proposed
date: 2026-10-21
---

### Context
Evaluating Neo4j vs Memgraph for knowledge graph traversal.
`

	// 4. Test direct validation via handleValidate with artifact_type="decisions"
	_, envValDecisions, err := handleValidate(ctx, nil, ValidateInput{
		ProjectRoot:  tmpDir,
		ArtifactType: "decisions",
		Content:      registryContent,
	})
	if err != nil || !envValDecisions.Success {
		t.Fatalf("handleValidate with artifact_type=decisions failed: %v", err)
	}
	valData, _ := envValDecisions.Data.(map[string]any)
	if valData["status"] != "valid" {
		t.Errorf("expected status 'valid' for clean multi-decision registry, got: %v (warnings: %v)", valData["status"], envValDecisions.Warnings)
	}

	// 5. Test auto-detection via handleValidate with artifact_type="decision"
	_, envValDecisionAuto, err := handleValidate(ctx, nil, ValidateInput{
		ProjectRoot:  tmpDir,
		ArtifactType: "decision",
		Content:      registryContent,
	})
	if err != nil || !envValDecisionAuto.Success {
		t.Fatalf("handleValidate with artifact_type=decision auto-detect failed: %v", err)
	}
	valDataAuto, _ := envValDecisionAuto.Data.(map[string]any)
	if valDataAuto["status"] != "valid" {
		t.Errorf("expected status 'valid' under auto-detection, got: %v (warnings: %v)", valDataAuto["status"], envValDecisionAuto.Warnings)
	}

	// 6. Save plan with multi-decision registry into workspace
	pipelineContent := `# Research Pipeline: Career Platform
**Complexity Score:** 18 / 24 → **Tier 3 (17–30 Sessions)**

## Sessions
#### T0-01: Graph Schema Definition
| Field | Value |
|---|---|
| ID | T0-01 |
| Layer | 0 |
| Door Type | One-Way |
| Dependencies | None |

` + "```prompt" + `
Define ontology and vertex relationships for career skills.
` + "```" + `
`
	_, envSave, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot:      tmpDir,
		Scope:            "project",
		Content:          pipelineContent,
		DecisionsContent: registryContent,
	})
	if err != nil || !envSave.Success {
		t.Fatalf("handleSavePlan failed: %v", err)
	}
	if len(envSave.Warnings) != 0 {
		t.Errorf("expected 0 warnings on clean plan save, got: %v", envSave.Warnings)
	}

	// 7. Test workspace-mode validation discovering DECISIONS.md
	_, envWsVal, err := handleValidate(ctx, nil, ValidateInput{
		ProjectRoot: tmpDir,
	})
	if err != nil || !envWsVal.Success {
		t.Fatalf("workspace validate failed: %v", err)
	}
	wsData, _ := envWsVal.Data.(map[string]any)
	decGroup, ok := wsData["decisions"].(map[string]any)
	if !ok {
		t.Fatalf("expected decisions group in workspace validate data: %v", wsData)
	}
	// All 3 decisions should be discovered and reported as valid
	if decGroup["valid"] != 3 {
		t.Errorf("expected 3 valid decisions discovered from DECISIONS.md, got: %v", decGroup["valid"])
	}
	if decGroup["errors"] != 0 {
		t.Errorf("expected 0 errors in decisions group, got: %v", decGroup["errors"])
	}

	// 8. Test malformed decision detection in multi-decision registry
	registryWithMalformed := registryContent + `
---
id: D-04
title: [unclosed malformed bracket
door_type: one-way
status: proposed
date: 2026-10-21
---
Body for D-04.
`
	_, envValMalformed, err := handleValidate(ctx, nil, ValidateInput{
		ProjectRoot:  tmpDir,
		ArtifactType: "decisions",
		Content:      registryWithMalformed,
	})
	if err != nil || !envValMalformed.Success {
		t.Fatalf("handleValidate on malformed registry failed: %v", err)
	}
	malData, _ := envValMalformed.Data.(map[string]any)
	if malData["status"] != "draft" && malData["status"] != "invalid" {
		t.Errorf("expected draft or invalid status for malformed decision registry, got: %v", malData["status"])
	}
	foundD04Issue := false
	for _, w := range envValMalformed.Warnings {
		if strings.Contains(w, "DECISION D-04:") && strings.Contains(w, "V-INVALID-FRONTMATTER") {
			foundD04Issue = true
			break
		}
	}
	if !foundD04Issue {
		t.Errorf("expected DECISION D-04 V-INVALID-FRONTMATTER warning, got: %v", envValMalformed.Warnings)
	}
}

// TestComprehensive_SavePlan_AutoNormalizesDecisionsRegistry verifies that save_plan
// intercepts raw frontmatter decision blocks in in.DecisionsContent and auto-normalizes
// them into collapsible <details> blocks on disk, guaranteeing markdown preview readability.
func TestComprehensive_SavePlan_AutoNormalizesDecisionsRegistry(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Initialize project scope
	_, envInit, err := handleInit(ctx, nil, InitInput{
		ProjectRoot: tmpDir,
		Scope:       "project",
	})
	if err != nil || !envInit.Success {
		t.Fatalf("init failed: %v", err)
	}

	pipelineContent := `# Research Pipeline
### T0-01: Foundation
| **ID** | T0-01 |
| **Dependencies** | None |
` + "```prompt\nfoundation prompt\n```\n"

	// Raw un-wrapped decision blocks as generated by smaller models
	rawDecisions := `# Architectural Decisions Log
This log captures all founding architecture decisions.

| Decision | Title | Door Type | Status | Date |
|---|---|---|---|---|
| D-01 | Database Engine | one-way | proposed | 2026-10-21 |
| D-02 | In-Memory Cache | two-way | proposed | 2026-10-21 |

<!-- DECISION: D-01 -->
---
id: D-01
title: Database Engine
door_type: one-way
status: proposed
date: 2026-10-21
---

### Context
Evaluating PostgreSQL vs CockroachDB.
<!-- /DECISION: D-01 -->

---

### D-02: In-Memory Cache
---
id: D-02
title: In-Memory Cache
door_type: two-way
status: proposed
date: 2026-10-21
---

### Context
Evaluating Redis vs Dragonfly.
`

	// 2. Save plan with raw decisions
	_, envSave, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot:      tmpDir,
		Scope:            "project",
		Content:          pipelineContent,
		DecisionsContent: rawDecisions,
	})
	if err != nil || !envSave.Success {
		t.Fatalf("handleSavePlan failed: %v, warnings: %v", err, envSave.Warnings)
	}

	// 3. Inspect research/DECISIONS.md on disk
	decisionsDiskPath := filepath.Join(tmpDir, core.ResearchDir, "DECISIONS.md")
	diskBytes, err := os.ReadFile(decisionsDiskPath)
	if err != nil {
		t.Fatalf("failed to read DECISIONS.md from disk: %v", err)
	}
	diskContent := string(diskBytes)

	// Verify that <details><summary>... was injected for D-01
	if !strings.Contains(diskContent, "<details>\n<summary><strong>D-01</strong> · Database Engine · <code>one-way</code> · proposed</summary>") {
		t.Errorf("expected D-01 details block on disk, got:\n%s", diskContent)
	}
	// Verify that D-02 unanchored was also normalized into <details>
	if !strings.Contains(diskContent, "<details>\n<summary><strong>D-02</strong> · In-Memory Cache · <code>two-way</code> · proposed</summary>") {
		t.Errorf("expected D-02 details block on disk, got:\n%s", diskContent)
	}
	// Verify ```yaml code fence is present
	if !strings.Contains(diskContent, "```yaml\nid: D-01") {
		t.Errorf("expected ```yaml code fence for D-01 on disk, got:\n%s", diskContent)
	}

	// 4. Test idempotency: saving the already-normalized content again does not double-wrap
	_, envSave2, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot:      tmpDir,
		Scope:            "project",
		Content:          pipelineContent,
		DecisionsContent: diskContent,
	})
	if err != nil || !envSave2.Success {
		t.Fatalf("second handleSavePlan failed: %v", err)
	}
	diskBytes2, err := os.ReadFile(decisionsDiskPath)
	if err != nil {
		t.Fatalf("failed to read DECISIONS.md after second save: %v", err)
	}
	diskContent2 := string(diskBytes2)
	if strings.Contains(diskContent2, "<details>\n<details>") || strings.Count(diskContent2, "<details>") != 2 {
		t.Errorf("idempotency violated: duplicate <details> found in:\n%s", diskContent2)
	}

	// 5. Run workspace validation and ensure decisions are valid
	_, envVal, err := handleValidate(ctx, nil, ValidateInput{
		ProjectRoot: tmpDir,
	})
	if err != nil || !envVal.Success {
		t.Fatalf("workspace validate failed: %v", err)
	}
	valData, _ := envVal.Data.(map[string]any)
	decGroup, ok := valData["decisions"].(map[string]any)
	if !ok {
		t.Fatalf("expected decisions group in validate data: %v", valData)
	}
	if decGroup["valid"] != 2 {
		t.Errorf("expected 2 valid decisions discovered, got: %v", decGroup["valid"])
	}
}

// TestComprehensive_DogfoodingHardening_E2ELifecycle exercises all Run 30 modifications:
// 1. PipelineContent alias in handleSavePlan
// 2. Proper noun immunity in lowGradeRe during gate evaluation
// 3. Proper noun immunity in gradeALineRe during quality evaluation
// 4. Single-turn auto_draft_from + auto_save in handleRecordDecision
// 5. Successful Phase 0 Gate clearance
func TestComprehensive_DogfoodingHardening_E2ELifecycle(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Initialize project workspace
	_, envInit, err := handleInit(ctx, nil, InitInput{
		ProjectRoot: tmpDir,
		Scope:       "project",
	})
	if err != nil || !envInit.Success {
		t.Fatalf("handleInit failed: %v", err)
	}

	// 2. Persist project plan using pipeline_content alias (instead of content)
	pipelineContent := `# Research Pipeline
### Session T1-01: Ingestion Engine
| **ID** | T1-01 |
| **Dependencies** | None |
` + "```prompt\nIngestion engine research\n```\n"

	decisionsContent := `# Decisions
<!-- DECISION: D-001 -->
<details>
<summary><strong>D-001</strong> · Ingestion Engine · <code>one-way</code> · proposed</summary>

` + "```yaml\nid: D-001\ntitle: Ingestion Engine\ndoor_type: one-way\nstatus: proposed\nreview_trigger: Throughput > 2M eps\n```\n" + `
</details>
<!-- /DECISION: D-001 -->
`

	_, envSavePlan, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot:      tmpDir,
		Scope:            "project",
		PipelineContent:  pipelineContent, // Using the alias!
		DecisionsContent: decisionsContent,
	})
	if err != nil || !envSavePlan.Success {
		t.Fatalf("handleSavePlan with PipelineContent failed: %v", err)
	}

	// Verify RESEARCH-PIPELINE.md was persisted
	pipePath := filepath.Join(tmpDir, core.PipelineFile)
	if _, err := os.Stat(pipePath); err != nil {
		t.Fatalf("RESEARCH-PIPELINE.md was not written: %v", err)
	}

	// 3. Save completed session T1-01 with proper nouns in parentheses
	sessionContent := `---
session_id: T1-01
title: Ingestion Engine Research
date: 2026-10-21
status: complete
door_type: one-way
informs_decisions: ["D-001"]
---
# Ingestion Engine Research

## Prior
We believed Kafka was the only feasible choice.

## Delta
| Prior | Research Finding | Status |
|---|---|---|
| Kafka is required | Redpanda handles 500k eps with zero JVM GC | Contradicted |

## Recommendation
We strongly recommend Redpanda (Decision D-001) for the telemetry bus.
Data is flushed to ClickHouse (ClickHouse cluster) for analytical queries.
Edge clients buffer locally in memory (Edge buffer caching) before streaming.
Secondary streaming uses Arroyo (Arroyo streaming engine).
Verified by official documentation [Grade A: official docs](https://redpanda.com/docs).
Corroborated by independent benchmark [Grade B: benchmark](https://benchmark.example.com).

## Alternatives Considered
- Apache Kafka: rejected due to JVM GC pause spikes under heavy load.
- NATS JetStream: rejected due to persistence and compaction bottlenecks.

## Discovered Concerns
Memory footprint increases under partition imbalances (>500 partitions per broker).

## Challenge Findings
Adversarial red-team stress test completed using Gary Klein premortem method.
`
	_, envSaveSess, err := handleSaveSession(ctx, nil, SaveSessionInput{
		ProjectRoot: tmpDir,
		SessionID:   "T1-01",
		Content:     sessionContent,
	})
	if err != nil || !envSaveSess.Success {
		t.Fatalf("handleSaveSession failed: %v, env: %+v", err, envSaveSess)
	}

	// 4. Auto-draft and auto-save the decision in a single turn using auto_save: true
	_, envRecDec, err := handleRecordDecision(ctx, nil, RecordDecisionInput{
		ProjectRoot:   tmpDir,
		DecisionID:    "D-001",
		AutoDraftFrom: "T1-01",
		AutoSave:      true,
	})
	if err != nil || !envRecDec.Success {
		t.Fatalf("handleRecordDecision with auto_save failed: %v, env: %+v", err, envRecDec)
	}

	// Verify ADR was written directly to disk
	adrPath := filepath.Join(tmpDir, core.ResearchDir, "D-001-decision.md")
	adrBytes, err := os.ReadFile(adrPath)
	if err != nil {
		t.Fatalf("reading created ADR: %v", err)
	}
	if !strings.Contains(string(adrBytes), "Redpanda") {
		t.Errorf("expected ADR to contain Redpanda, got:\n%s", string(adrBytes))
	}

	// 5. Evaluate Phase 0 Gate on the workspace
	_, envGate, err := handleRunGate(ctx, nil, RunGateInput{
		ProjectRoot: tmpDir,
		Track:       "all",
	})
	if err != nil || !envGate.Success {
		t.Fatalf("handleRunGate failed: %v, env: %+v", err, envGate)
	}
	gateData, _ := envGate.Data.(map[string]any)

	// Verify that B3-EVIDENTIARY did NOT falsely trigger on (ClickHouse), (Decision D-001), (Edge buffer caching)
	blockingIssues, _ := gateData["blocking_issues"].([]string)
	for _, issue := range blockingIssues {
		if strings.Contains(issue, "B3-EVIDENTIARY") {
			t.Errorf("unexpected false-positive B3-EVIDENTIARY blocking issue: %s", issue)
		}
	}
}

// TestComprehensive_PromptActionVerbsAndMiddlewareEquivalence validates that:
// 1. prompts/list returns strictly the 4 canonical action verbs (compare, decision, project, gate).
// 2. Calling prompts/get with canonical, legacy (vivechak_*), or shorthand (vck_*) produces identical results.
// 3. Prompt arguments are accurately interpolated and anti-spill persistence directives are attached.
func TestComprehensive_PromptActionVerbsAndMiddlewareEquivalence(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	listRes, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}

	expectedPrompts := map[string]string{
		"compare":  "Compare Options (WEP Matrix)",
		"decision": "Architectural Decision (ADR)",
		"project":  "Project Pipeline (Full FAD)",
		"gate":     "Phase 0 Pre-Codebase Exit Gate",
	}

	if len(listRes.Prompts) != len(expectedPrompts) {
		t.Fatalf("expected exactly %d prompts, got %d", len(expectedPrompts), len(listRes.Prompts))
	}

	for _, p := range listRes.Prompts {
		expTitle, ok := expectedPrompts[p.Name]
		if !ok {
			t.Errorf("unexpected prompt registered: %s", p.Name)
			continue
		}
		if p.Title != expTitle {
			t.Errorf("prompt %s: expected title %q, got %q", p.Name, expTitle, p.Title)
		}
		if p.Description == "" {
			t.Errorf("prompt %s: empty description", p.Name)
		}
	}

	// Matrix of prompt alias calls
	tests := []struct {
		name       string
		canonical  string
		legacy     string
		shorthand  string
		args       map[string]string
		mustSubstr []string
	}{
		{
			name:      "Comparison",
			canonical: "compare",
			legacy:    "vivechak_compare",
			shorthand: "vck_compare",
			args: map[string]string{
				"options":  "NATS vs Kafka vs RabbitMQ",
				"criteria": "Throughput, Multi-Tenancy",
				"context":  "High-throughput messaging fabric",
			},
			mustSubstr: []string{
				"OPTIONS: NATS vs Kafka vs RabbitMQ",
				"CRITERIA: Throughput, Multi-Tenancy",
				"CONTEXT: High-throughput messaging fabric",
				"vivechak_save_plan",
				"C-NNN-comparison.md",
			},
		},
		{
			name:      "Decision",
			canonical: "decision",
			legacy:    "vivechak_decision",
			shorthand: "vck_decision",
			args: map[string]string{
				"dilemma": "Migrate from REST to gRPC for inter-service communication",
				"context": "Low-latency microservices cluster",
			},
			mustSubstr: []string{
				"DECISION: Migrate from REST to gRPC for inter-service communication",
				"CONTEXT: Low-latency microservices cluster",
				"vivechak_save_plan",
				"vivechak_record_decision",
			},
		},
		{
			name:      "Project",
			canonical: "project",
			legacy:    "vivechak_project",
			shorthand: "vck_project",
			args: map[string]string{
				"vision":  "Build next-generation distributed vector search engine",
				"context": "Rust, SIMD optimizations, AWS Graviton3",
			},
			mustSubstr: []string{
				"Build next-generation distributed vector search engine",
				"Rust, SIMD optimizations",
				"RESEARCH-PIPELINE.md",
				"pipeline_content",
			},
		},
		{
			name:      "Gate",
			canonical: "gate",
			legacy:    "vivechak_gate",
			shorthand: "vck_gate",
			args: map[string]string{
				"track":        "track_b",
				"project_root": "/opt/workspace",
			},
			mustSubstr: []string{
				`track="track_b"`,
				`project_root="/opt/workspace"`,
				"vivechak_run_gate",
				"track_a_result",
				"track_b_result",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rCanon, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: tc.canonical, Arguments: tc.args})
			if err != nil {
				t.Fatalf("canonical prompt %s: %v", tc.canonical, err)
			}
			rLeg, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: tc.legacy, Arguments: tc.args})
			if err != nil {
				t.Fatalf("legacy prompt %s: %v", tc.legacy, err)
			}
			rShort, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: tc.shorthand, Arguments: tc.args})
			if err != nil {
				t.Fatalf("shorthand prompt %s: %v", tc.shorthand, err)
			}

			txtCanon := rCanon.Messages[0].Content.(*mcp.TextContent).Text
			txtLeg := rLeg.Messages[0].Content.(*mcp.TextContent).Text
			txtShort := rShort.Messages[0].Content.(*mcp.TextContent).Text

			if txtCanon != txtLeg {
				t.Errorf("expected canonical %q == legacy %q", tc.canonical, tc.legacy)
			}
			if txtCanon != txtShort {
				t.Errorf("expected canonical %q == shorthand %q", tc.canonical, tc.shorthand)
			}

			for _, sub := range tc.mustSubstr {
				if !strings.Contains(txtCanon, sub) {
					t.Errorf("prompt %s missing expected substring %q in:\n%s", tc.canonical, sub, txtCanon)
				}
			}
		})
	}
}







