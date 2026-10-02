package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestRealComprehensive_AllScopes_AllModifications conducts an exhaustive, real-world
// trial testing every scope, tool, parser, and guardrail via real stdio JSON-RPC MCP
// connections against the live compiled binary.
func TestRealComprehensive_AllScopes_AllModifications(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	// 1. Build the live binary
	binDir := t.TempDir()
	binName := "vck_real_trial"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	// 2. Connect to live MCP server over stdio
	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "real-trial-client", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server over stdio: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	// Helper to parse MCP CallToolResult into Envelope
	callAndParse := func(toolName string, args map[string]any) testEnvelope {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		})
		if err != nil {
			t.Fatalf("call to %s failed: %v", toolName, err)
		}
		return parseTestEnvelope(t, res)
	}

	// =========================================================================
	// SCENARIO 1: FULL PROJECT SCOPE LIFECYCLE (WITH ALL AUDITED HARDENINGS)
	// =========================================================================
	t.Run("ProjectScope_FullLifecycle_And_HardenedGuardrails", func(t *testing.T) {
		projectDir := t.TempDir()

		// 1. vivechak_init
		envInit := callAndParse("vivechak_init", map[string]any{
			"project_root": projectDir,
			"scope":        "project",
		})
		if !envInit.Success {
			t.Fatalf("vivechak_init unsuccessful: %s", envInit.Message)
		}

		// Verify research/.gitignore
		giData, err := os.ReadFile(filepath.Join(projectDir, core.ResearchDir, ".gitignore"))
		if err != nil || !strings.Contains(string(giData), "*.lock") {
			t.Errorf("expected .gitignore with *.lock, got err=%v content=%s", err, string(giData))
		}

		// Verify 6 templates copied, including SESSION.template.md
		tmplCopied, ok := envInit.Data["templates_copied"].([]any)
		if !ok || len(tmplCopied) != 6 {
			t.Fatalf("expected 6 templates copied, got: %v", envInit.Data["templates_copied"])
		}
		if _, err := os.Stat(filepath.Join(projectDir, core.TemplatesDir, "SESSION.template.md")); err != nil {
			t.Errorf("SESSION.template.md does not exist in templates/: %v", err)
		}

		// 2. vivechak_prepare_generator
		envPrep := callAndParse("vivechak_prepare_generator", map[string]any{
			"project_root": projectDir,
			"context":      "High-performance local financial intelligence engine with local LLM and encrypted datastore.",
		})
		if !envPrep.Success {
			t.Fatalf("vivechak_prepare_generator failed: %s", envPrep.Message)
		}
		prepPrompt := envPrep.Data["prompt"].(string)
		if !strings.Contains(prepPrompt, "High-performance local financial intelligence") {
			t.Errorf("expected brief in generator prompt")
		}

		// 3. vivechak_save_plan: test hardened dependency parsing (markdown links & bullet lists)
		pipelineContent := `# Research Pipeline: Katha Ledger

**Complexity Score:** 14 / 24 → **Tier 2 (8–12 Sessions)**

**Execution DAG:**
` + "```mermaid" + `
graph TD
    R01[R-01: Ingestion Engine]
    R02[R-02: Local Embedding Model]
    R03[R-03: Encrypted Storage Engine]
    SYN[SYN-01: Founding Architecture Document]

    R01 --> R02
    R01 --> R03
    R02 --> SYN
    R03 --> SYN
` + "```" + `

## Sessions

#### R-01: Ingestion Engine — Bank Data Parsing

| Field | Value |
|---|---|
| **ID** | R-01 |
| **Layer** | 0 |
| **Door Type** | Two-Way |
| **Decision** | D-001 |
| **Dependencies** | None (parallel) |
| **Output File** | ` + "`sessions/R-01.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: R-01 Ingestion Engine
## BRIEF
Research bank sync options.
## SCOPE
SimpleFIN vs Plaid.
## DELIVERABLE
Concrete recommendation with Grade A evidence.
` + "```" + `

---

#### R-01b: Auth Strategy — Master Key Derivation

| Field | Value |
|---|---|
| **ID** | R-01b |
| **Layer** | 0 |
| **Door Type** | Two-Way |
| **Dependencies** | None (parallel) |
| **Output File** | ` + "`sessions/R-01b.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: R-01b Auth Strategy
## BRIEF
Argon2id vs PBKDF2 for key derivation.
## DELIVERABLE
Key derivation recommendation with Grade A evidence.
` + "```" + `

---

#### R-02: Local Embedding Model — Semantic Classification

| Field | Value |
|---|---|
| **ID** | R-02 |
| **Layer** | 1 |
| **Door Type** | One-Way |
| **Decision** | D-002 |
| **Dependencies** | [R-01](./sessions/R-01.md) |
| **Output File** | ` + "`sessions/R-02.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: R-02 Local Embedding Model
## BRIEF
Research local embedding models.
[UPSTREAM_FINDINGS]
## DELIVERABLE
Model selection with benchmark evidence.
` + "```" + `

---

#### R-03: Encrypted Storage Engine — Primary Datastore

| Field | Value |
|---|---|
| **ID** | R-03 |
| **Layer** | 1 |
| **Door Type** | One-Way |
| **Decision** | D-003 |
| **Dependencies** | ` + "`R-01`" + ` |
| **Output File** | ` + "`sessions/R-03.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: R-03 Encrypted Storage Engine
## BRIEF
Investigate embedded encrypted databases.
[UPSTREAM_FINDINGS]
## DELIVERABLE
Storage selection with Grade A evidence.
` + "```" + `

---

#### SYN-01: Founding Architecture Document — Katha Ledger

| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Layer** | 2 |
| **Door Type** | One-Way |
| **Decision** | All |
| **Dependencies** | - R-02\n- R-03 |
| **Output File** | ` + "`research/FAD.md`" + ` |

` + "```prompt" + `
# RESEARCH BRIEF: SYN-01 Founding Architecture Document
## BRIEF
Synthesize all upstream findings into FAD.
[ALL_SESSION_FINDINGS]
## DELIVERABLE
Founding architecture document.
` + "```" + `
`
		envPlan := callAndParse("vivechak_save_plan", map[string]any{
			"project_root":     projectDir,
			"content": pipelineContent,
		})
		if !envPlan.Success {
			t.Fatalf("vivechak_save_plan failed: %s", envPlan.Message)
		}
		if _, err := os.Stat(filepath.Join(projectDir, core.PipelineFile)); err != nil {
			t.Fatalf("RESEARCH-PIPELINE.md does not exist: %v", err)
		}

		// 4. vivechak_status: verify Layer 0 is R-01 and next session is R-01
		envStatus := callAndParse("vivechak_status", map[string]any{
			"project_root": projectDir,
		})
		if !envStatus.Success {
			t.Fatalf("status failed: %s", envStatus.Message)
		}
		if envStatus.Data["has_pipeline"] != true {
			t.Errorf("expected has_pipeline true, got: %v", envStatus.Data["has_pipeline"])
		}
		if envStatus.Data["session_count"].(float64) != 0 {
			t.Errorf("expected 0 session count, got: %v", envStatus.Data["session_count"])
		}

		// 5. vivechak_next_session: get R-01 prompt, verify parallelism_hint for multiple Layer 0 sessions
		envNextR1 := callAndParse("vivechak_next_session", map[string]any{
			"project_root": projectDir,
		})
		if !envNextR1.Success || envNextR1.Data["session_id"] != "R-01" {
			t.Fatalf("expected next session R-01, got: %v", envNextR1.Data)
		}
		parallelHint, hasHint := envNextR1.Data["parallelism_hint"].(string)
		if !hasHint || !strings.Contains(parallelHint, "PARALLEL EXECUTION AVAILABLE: 2 independent sessions") {
			t.Errorf("expected parallelism_hint in data map, got: %v", envNextR1.Data["parallelism_hint"])
		}
		if !strings.HasPrefix(envNextR1.NextStep, "⚡ 2 sessions ready in parallel") {
			t.Errorf("expected NextStep to start with parallel hint, got: %s", envNextR1.NextStep)
		}

		// Save R-01b session
		r01bContent := `---
session_id: R-01b
title: Auth Strategy
date: 2026-10-01
status: complete
tags: [auth, argon2, security]
---
# Auth Strategy
## Recommendation
Adopt Argon2id for master key derivation. Grade A (OWASP 2024 guidance)
`
		envSaveR1b := callAndParse("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "R-01b",
			"content":      r01bContent,
		})
		if !envSaveR1b.Success {
			t.Fatalf("save R-01b failed: %s", envSaveR1b.Message)
		}

		// 6. vivechak_save_session for R-01: include Delta with contradicted belief (testing P8 confirmation bias check)
		r01Content := `---
session_id: R-01
title: Ingestion Engine
date: 2026-10-01
status: complete
tags: [ingestion, plaid, simplefin]
---
# Ingestion Engine

## Recommendation
SimpleFIN Bridge instead of Plaid which was rejected due to recurring per-user API licensing. Grade A (direct pricing and API review)

## Prior
- Believed Plaid OAuth was mandatory for regional credit unions.

## Delta
| Belief | Status | Impact |
|---|---|---|
| Plaid is mandatory | Contradicted — SimpleFIN covers OFX direct connect | High cost reduction |

## Evaluated Options & Alternatives

### Option 1: SimpleFIN Bridge
Read-only financial data bridge. Grade A (official documentation)

### Option 2: Plaid API
Enterprise banking aggregator. Grade B (benchmarks)

## Discovered Concerns & Failure Modes

### Risk 1: Sync Lag
Polling intervals may delay real-time balance refresh. Grade A (API specification)
`
		envSaveR1 := callAndParse("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "R-01",
			"content":      r01Content,
		})
		if !envSaveR1.Success {
			t.Fatalf("save R-01 failed: %s", envSaveR1.Message)
		}

		// 7a. Test auto_draft_from: generate draft from R-01 (without content)
		envDraft := callAndParse("vivechak_record_decision", map[string]any{
			"project_root":    projectDir,
			"decision_id":     "D-001",
			"auto_draft_from": "R-01",
		})
		if !envDraft.Success {
			t.Fatalf("auto_draft_from failed: %s", envDraft.Message)
		}
		draftText := envDraft.Data["draft_content"].(string)
		if !strings.Contains(draftText, "D-001") || !strings.Contains(draftText, "SimpleFIN") {
			t.Errorf("auto draft missing expected fields: %s", draftText)
		}

		// 7b. Save finalized D-001 decision
		envRecD1 := callAndParse("vivechak_record_decision", map[string]any{
			"project_root": projectDir,
			"decision_id":  "D-001",
			"content": `---
id: D-001
title: Bank Connectivity Strategy
status: accepted
door_type: two-way
review_trigger: "SimpleFIN downtime exceeds 1% monthly SLA"
---
# D-001: Bank Connectivity Strategy
## Context
We require reliable bank feed aggregation for personal accounting.
## Decision
Adopt SimpleFIN Bridge instead of Plaid. Grade A (direct pricing review)
## Consequences
Saves API costs while maintaining read-only access.
`,
		})
		if !envRecD1.Success {
			t.Fatalf("record D-001 failed: %s", envRecD1.Message)
		}

		// Verify DECISIONS.md was compiled and wrapped in <details>
		decData, err := os.ReadFile(filepath.Join(projectDir, core.DecisionsFile))
		if err != nil {
			t.Fatalf("reading DECISIONS.md: %v", err)
		}
		if !strings.Contains(string(decData), "<details>") || !strings.Contains(string(decData), "<!-- DECISION: D-001 -->") {
			t.Errorf("DECISIONS.md missing <details> or marker: %s", string(decData))
		}

		// 8. vivechak_next_session for R-02 and R-03: verify upstream R-01 injected
		envNextR2 := callAndParse("vivechak_next_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "R-02",
		})
		promptR2 := envNextR2.Data["prompt"].(string)
		if !strings.Contains(promptR2, "SimpleFIN") {
			t.Fatalf("expected R-01 upstream findings injected into R-02 prompt, got:\n%s", promptR2)
		}

		// Save R-02 session
		r02Content := `---
session_id: R-02
title: Local Embedding Model
date: 2026-10-01
status: complete
tags: [ai, onnx, embeddings]
---
# Local Embedding Model
## Recommendation
ONNX Runtime with all-MiniLM-L6-v2 instead of OpenAI embeddings. Grade A (local benchmark)
## Evaluated Options
- all-MiniLM-L6-v2: 384 dimensions, 15ms latency. Grade A (benchmark)
- BGE-small: 384 dimensions, 22ms latency. Grade B (benchmark)
## Concerns
Memory consumption under concurrent categorization threads.
`
		envSaveR2 := callAndParse("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "R-02",
			"content":      r02Content,
		})
		if !envSaveR2.Success {
			t.Fatalf("save R-02 failed: %s", envSaveR2.Message)
		}

		// Save R-03 session
		r03Content := `---
session_id: R-03
title: Encrypted Storage Engine
date: 2026-10-01
status: complete
tags: [storage, sqlite, cipher]
---
# Encrypted Storage Engine
## Recommendation
SQLite WAL mode with SQLCipher instead of PostgreSQL. Grade A (encryption benchmark)
## Evaluated Options
- SQLite + SQLCipher: embedded, zero external daemon. Grade A (official docs)
- PostgreSQL + pgcrypto: requires background service. Grade A (official docs)
## Concerns
Write concurrency locks when batch indexing.
`
		envSaveR3 := callAndParse("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "R-03",
			"content":      r03Content,
		})
		if !envSaveR3.Success {
			t.Fatalf("save R-03 failed: %s", envSaveR3.Message)
		}

		// Record D-002 (One-Way Door)
		envRecD2 := callAndParse("vivechak_record_decision", map[string]any{
			"project_root": projectDir,
			"decision_id":  "D-002",
			"content": `---
id: D-002
title: Primary Storage Engine
status: accepted
door_type: one-way
review_trigger: "Database file size exceeds 50 GB"
---
# D-002: Primary Storage Engine
## Context
Local-first application requires embedded datastore with at-rest encryption.
## Decision
We choose SQLite WAL mode with SQLCipher. Grade A (benchmark)
## Consequences
Requires strict serialization of write transactions.
`,
		})
		if !envRecD2.Success {
			t.Fatalf("record D-002 failed: %s", envRecD2.Message)
		}

		// Verify DECISIONS.md has D-001 and D-002 deterministically sorted
		decData2, _ := os.ReadFile(filepath.Join(projectDir, core.DecisionsFile))
		strDec := string(decData2)
		idxD1 := strings.Index(strDec, "<!-- DECISION: D-001 -->")
		idxD2 := strings.Index(strDec, "<!-- DECISION: D-002 -->")
		if idxD1 == -1 || idxD2 == -1 || idxD1 >= idxD2 {
			t.Fatalf("DECISIONS.md markers not in deterministic sorted order: idxD1=%d, idxD2=%d", idxD1, idxD2)
		}

		// 9. vivechak_amend_session: amend R-01 which has completed downstream sessions (R-02, R-03)
		// Verifies W-STALE-DOWNSTREAM warning and potentially_stale_sessions in data envelope
		envAmendR1 := callAndParse("vivechak_amend_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "R-01",
			"amendment":    "POST-HOC UPDATE: SimpleFIN now offers direct webhook push.",
		})
		if !envAmendR1.Success {
			t.Fatalf("amend R-01 failed: %s", envAmendR1.Message)
		}
		foundStale := false
		for _, w := range envAmendR1.Warnings {
			if strings.Contains(w, "W-STALE-DOWNSTREAM") && strings.Contains(w, "R-02") && strings.Contains(w, "R-03") {
				foundStale = true
				break
			}
		}
		if !foundStale {
			t.Errorf("expected W-STALE-DOWNSTREAM warning mentioning R-02 and R-03, got: %v", envAmendR1.Warnings)
		}
		staleSessions, ok := envAmendR1.Data["potentially_stale_sessions"].([]any)
		if !ok || len(staleSessions) < 2 {
			t.Errorf("expected potentially_stale_sessions to contain R-02 and R-03, got: %v", envAmendR1.Data["potentially_stale_sessions"])
		}

		// Also amend R-02 with a post-hoc note for context injection testing
		envAmendR2 := callAndParse("vivechak_amend_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "R-02",
			"amendment":    "POST-HOC CORRECTION: Validated zero-allocation quantization for INT8 inference.",
		})
		if !envAmendR2.Success {
			t.Fatalf("amend R-02 failed: %s", envAmendR2.Message)
		}

		// 10. vivechak_next_session for SYN-01 (synthesis): verify technology matrix, belief evolution, amendment preservation
		envSyn := callAndParse("vivechak_next_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "SYN-01",
			"verbose":      false,
		})
		if !envSyn.Success {
			t.Fatalf("fetch synthesis session failed: %s", envSyn.Message)
		}
		synPrompt := envSyn.Data["prompt"].(string)

		// 10a. Prompt must NOT be truncated even with verbose=false
		if strings.Contains(synPrompt, "... [truncated") {
			t.Errorf("synthesis prompt was truncated with verbose=false")
		}
		// 10b. Technology matrix must be injected
		if !strings.Contains(synPrompt, "## TECHNOLOGY CHOICES ACROSS SESSIONS") {
			t.Errorf("missing technology choices matrix in synthesis prompt:\n%s", synPrompt)
		}
		if !strings.Contains(synPrompt, "R-01") || !strings.Contains(synPrompt, "R-02") || !strings.Contains(synPrompt, "R-03") {
			t.Errorf("technology matrix missing sessions: %s", synPrompt)
		}
		// 10c. Belief evolution must be injected
		if !strings.Contains(synPrompt, "## BELIEF EVOLUTION ACROSS SESSIONS") || !strings.Contains(synPrompt, "Plaid is mandatory") {
			t.Errorf("missing belief evolution in synthesis prompt:\n%s", synPrompt)
		}
		// 10d. Post-hoc amendment must be preserved
		if !strings.Contains(synPrompt, "POST-HOC CORRECTION: Validated zero-allocation") {
			t.Errorf("post-hoc amendment was not extracted into upstream context:\n%s", synPrompt)
		}

		// 11. vivechak_save_session for SYN-01: verify single-write to research/FAD.md (NO duplicate in sessions/)
		fadContent := `---
id: FAD-01
session_id: SYN-01
title: Katha Ledger Architecture
synthesis_date: 2026-10-01
status: complete
---
# Founding Architecture Document: Katha Ledger

## Executive Summary
Katha Ledger is an embedded, encrypted personal finance intelligence system. Grade A (evidence synthesis)

## Technology Coherence
- Bank Ingestion: SimpleFIN Bridge. Grade A (API review)
- AI Categorization: ONNX Runtime all-MiniLM-L6-v2. Grade A (benchmark)
- Primary Datastore: SQLite WAL + SQLCipher. Grade A (benchmark)

## Component Architecture
Components communicate through in-memory Go channels with strict data isolation.
`
		envSaveSyn := callAndParse("vivechak_save_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "SYN-01",
			"content":      fadContent,
		})
		if !envSaveSyn.Success {
			t.Fatalf("save SYN-01 failed: %s", envSaveSyn.Message)
		}

		// Verify root_copy returned in data and mirrored to project root
		if envSaveSyn.Data["root_copy"] != "FOUNDING-ARCHITECTURE.md" {
			t.Errorf("expected root_copy 'FOUNDING-ARCHITECTURE.md', got: %v", envSaveSyn.Data["root_copy"])
		}
		rootFADPath := filepath.Join(projectDir, "FOUNDING-ARCHITECTURE.md")
		rootFADBytes, err := os.ReadFile(rootFADPath)
		if err != nil {
			t.Fatalf("expected FOUNDING-ARCHITECTURE.md at project root: %v", err)
		}
		fadDiskBytes, _ := os.ReadFile(filepath.Join(projectDir, core.FADFile))
		if string(rootFADBytes) != string(fadDiskBytes) {
			t.Errorf("root FOUNDING-ARCHITECTURE.md does not match research/FAD.md")
		}

		// Verify research/FAD.md exists
		if _, err := os.Stat(filepath.Join(projectDir, core.FADFile)); err != nil {
			t.Fatalf("research/FAD.md does not exist: %v", err)
		}
		// Verify sessions/SYN-01.md DOES NOT EXIST (zero duplicate / single source of truth)
		if _, err := os.Stat(filepath.Join(projectDir, core.SessionsDir, "SYN-01.md")); err == nil {
			t.Fatalf("sessions/SYN-01.md exists! Dual-write bug detected.")
		}

		// 12. vivechak_amend_session on synthesis session (targeting research/FAD.md)
		envAmendFAD := callAndParse("vivechak_amend_session", map[string]any{
			"project_root": projectDir,
			"session_id":   "FAD",
			"amendment":    "Added security guardrails for IPC transport encryption.",
		})
		if !envAmendFAD.Success {
			t.Fatalf("amend FAD failed: %s", envAmendFAD.Message)
		}
		fadDisk, _ := os.ReadFile(filepath.Join(projectDir, core.FADFile))
		if !strings.Contains(string(fadDisk), "Added security guardrails for IPC transport") {
			t.Fatalf("amendment missing from FAD.md: %s", string(fadDisk))
		}

		// 13. Deliberate Failure Gate Test: change D-002 status to 'proposed' on disk
		d002Path := filepath.Join(projectDir, core.ResearchDir, "D-002-decision.md")
		d002Data, err := os.ReadFile(d002Path)
		if err != nil {
			// Find actual D-002 file
			dEntries, _ := os.ReadDir(filepath.Join(projectDir, core.ResearchDir))
			for _, de := range dEntries {
				if strings.HasPrefix(de.Name(), "D-002") {
					d002Path = filepath.Join(projectDir, core.ResearchDir, de.Name())
					d002Data, _ = os.ReadFile(d002Path)
					break
				}
			}
		}
		tamperedD002 := strings.Replace(string(d002Data), "status: accepted", "status: proposed", 1)
		_ = os.WriteFile(d002Path, []byte(tamperedD002), 0o644)

		envGateFail := callAndParse("vivechak_run_gate", map[string]any{
			"project_root": projectDir,
			"verbose":      true,
		})
		if envGateFail.Data["gate_passed"] == true {
			t.Fatalf("gate should have failed when D-002 is proposed")
		}
		// Verify actionable diagnostic error
		foundDiag := false
		for _, w := range envGateFail.Warnings {
			if strings.Contains(w, "D-002") && strings.Contains(w, "proposed") && strings.Contains(w, "status to 'accepted'") {
				foundDiag = true
				break
			}
		}
		if !foundDiag {
			t.Errorf("expected actionable diagnostic error for proposed D-002, got warnings: %v", envGateFail.Warnings)
		}

		// Restore D-002 to accepted
		restoredD002 := strings.Replace(tamperedD002, "status: proposed", "status: accepted", 1)
		_ = os.WriteFile(d002Path, []byte(restoredD002), 0o644)

		// 14. Plant non-ADR markdown files in research/ to test whitelist scanning and core.IsSpecialResearchFile
		_ = os.WriteFile(filepath.Join(projectDir, core.ResearchDir, "NOTES.md"), []byte("# Notes\nRandom notes with status: sealed"), 0o644)
		_ = os.WriteFile(filepath.Join(projectDir, core.ResearchDir, "D-001-plan.md"), []byte("# Plan\nNot an ADR"), 0o644)
		_ = os.WriteFile(filepath.Join(projectDir, core.ResearchDir, "D-001-conflict-resolution.md"), []byte("# Conflict\nNot an ADR"), 0o644)
		_ = os.WriteFile(filepath.Join(projectDir, core.ResearchDir, "FOUNDING-ARCHITECTURE.md"), []byte("# Copy\nNot an ADR"), 0o644)

		// 15. Workspace validation: verify exact 2 decisions reported (no double counting)
		envValWs := callAndParse("vivechak_validate", map[string]any{
			"project_root": projectDir,
		})
		if !envValWs.Success {
			t.Fatalf("workspace validate failed: %s", envValWs.Message)
		}
		decMap := envValWs.Data["decisions"].(map[string]any)
		validDecCount := int(decMap["valid"].(float64))
		errorDecCount := int(decMap["errors"].(float64))
		if validDecCount != 2 {
			t.Errorf("expected exactly 2 valid decisions, got %d (data: %v, warnings: %v)", validDecCount, decMap, envValWs.Warnings)
		}
		if errorDecCount != 0 {
			t.Errorf("expected 0 decision errors, got %d", errorDecCount)
		}

		// 16. vivechak_run_gate: verify gate PASSES and auto-persists research/PHASE-0-GATE.md
		envGatePass := callAndParse("vivechak_run_gate", map[string]any{
			"project_root": projectDir,
			"verbose":      true,
		})
		if envGatePass.Data["gate_passed"] != true || envGatePass.Data["gate_status"] != "PASS" {
			t.Fatalf("expected gate PASS, got status=%v passed=%v warnings=%v",
				envGatePass.Data["gate_status"], envGatePass.Data["gate_passed"], envGatePass.Warnings)
		}
		if envGatePass.Data["gate_artifact"] != "research/PHASE-0-GATE.md" {
			t.Errorf("expected gate_artifact 'research/PHASE-0-GATE.md', got: %v", envGatePass.Data["gate_artifact"])
		}
		gateData, err := os.ReadFile(filepath.Join(projectDir, core.GateFile))
		if err != nil {
			t.Fatalf("expected research/PHASE-0-GATE.md to exist: %v", err)
		}
		gateStr := string(gateData)
		if !strings.Contains(gateStr, `verdict: "PASS"`) || !strings.Contains(gateStr, `track_a_result: "PASS"`) {
			t.Errorf("gate artifact missing expected frontmatter: %s", gateStr)
		}
		if !strings.Contains(gateStr, "| D-001 |") || !strings.Contains(gateStr, "| D-002 |") {
			t.Errorf("gate artifact missing decision routing table rows: %s", gateStr)
		}
		if !strings.Contains(gateStr, "- [x] Decision logged in ADR") {
			t.Errorf("gate artifact missing checked checkboxes: %s", gateStr)
		}

		// 17. Run CLI doctor on the project directory
		docCmd := exec.CommandContext(ctx, binPath, "doctor", projectDir)
		docOut, docErr := docCmd.CombinedOutput()
		if docErr != nil {
			t.Fatalf("vck doctor failed: %v\noutput: %s", docErr, string(docOut))
		}
		if !strings.Contains(string(docOut), "✓ Workspace exists") {
			t.Errorf("doctor output missing expected check: %s", string(docOut))
		}
		if !strings.Contains(string(docOut), "Workspace is healthy") {
			t.Errorf("doctor output should report healthy workspace: %s", string(docOut))
		}
	})

	// =========================================================================
	// SCENARIO 2: DECISION SCOPE END-TO-END TRIAL
	// =========================================================================
	t.Run("DecisionScope_FullLifecycle", func(t *testing.T) {
		decDir := t.TempDir()

		// 1. init
		callAndParse("vivechak_init", map[string]any{
			"project_root": decDir,
			"scope":        "decision",
		})

		// 2. prepare_generator
		callAndParse("vivechak_prepare_generator", map[string]any{
			"project_root": decDir,
			"scope":        "decision",
			"context":      "Choose between badger and bolt for embedded key-value storage.",
		})

		// 3. save_plan
		planContent := `# Decision Research Plan: D-001

#### S1-kv-eval: Key-Value Storage Evaluation

| Field | Value |
|---|---|
| **ID** | S1-kv-eval |
| **Decision** | D-001 |
| **Door Type** | Two-Way |
| **Output File** | sessions/S1-kv-eval.md |

` + "```prompt\n# Brief\nEvaluate Badger vs Bolt.\n```\n"

		callAndParse("vivechak_save_plan", map[string]any{
			"project_root":     decDir,
			"scope":            "decision",
			"decision_id":      "D-001",
			"content": planContent,
		})

		// 4. next_session
		envNext := callAndParse("vivechak_next_session", map[string]any{
			"project_root": decDir,
		})
		if envNext.Data["session_id"] != "S1-kv-eval" {
			t.Errorf("expected session S1-kv-eval, got: %v", envNext.Data["session_id"])
		}

		// 5. save_session
		callAndParse("vivechak_save_session", map[string]any{
			"project_root": decDir,
			"session_id":   "S1-kv-eval",
			"content": `---
session_id: S1-kv-eval
title: KV Storage Evaluation
date: 2026-10-01
status: complete
---
# KV Storage Evaluation
Recommendation: BadgerDB due to LSM tree write throughput. Grade A (benchmark)
Evaluated: BoltDB, BadgerDB.
Concerns: Memory tuning for LSM levels.
`,
		})

		// 6. record_decision
		callAndParse("vivechak_record_decision", map[string]any{
			"project_root": decDir,
			"decision_id":  "D-001",
			"content": `---
id: D-001
title: Key-Value Storage
status: accepted
door_type: two-way
review_trigger: "Read-to-write ratio drops below 2:1"
---
# D-001: Key-Value Storage
## Context
Need fast local key-value datastore.
## Decision
Adopt BadgerDB. Grade A (benchmark)
## Consequences
Requires write-buffer configuration.
`,
		})

		// 7. Plant non-ADR file
		_ = os.WriteFile(filepath.Join(decDir, core.ResearchDir, "NOTES.md"), []byte("# Notes"), 0o644)

		// 8. validate
		envVal := callAndParse("vivechak_validate", map[string]any{"project_root": decDir})
		decVal := envVal.Data["decisions"].(map[string]any)
		if decVal["valid"].(float64) != 1 || decVal["errors"].(float64) != 0 {
			t.Errorf("expected 1 valid decision, 0 errors, got: %v", decVal)
		}

		// 9. run_gate
		envGate := callAndParse("vivechak_run_gate", map[string]any{
			"project_root": decDir,
			"verbose":      true,
		})
		if envGate.Data["gate_passed"] != true {
			t.Fatalf("expected decision scope gate PASS, got: %v warnings=%v", envGate.Data, envGate.Warnings)
		}

		// 10. CLI doctor
		docCmd := exec.CommandContext(ctx, binPath, "doctor", decDir)
		if docOut, err := docCmd.CombinedOutput(); err != nil {
			t.Fatalf("doctor failed on decision scope: %v\noutput: %s", err, string(docOut))
		}
	})

	// =========================================================================
	// SCENARIO 3: COMPARISON SCOPE END-TO-END TRIAL
	// =========================================================================
	t.Run("ComparisonScope_FullLifecycle", func(t *testing.T) {
		cmpDir := t.TempDir()

		// 1. init
		callAndParse("vivechak_init", map[string]any{
			"project_root": cmpDir,
			"scope":        "comparison",
		})

		// 2. save_plan
		cmpPlan := `# Comparison Plan: CMP-01

| Field | Value |
|---|---|
| **ID** | CMP-01 |
| **Output File** | sessions/CMP-01.md |

` + "```prompt\n# Brief\nCompare RabbitMQ vs NATS.\n```\n"

		callAndParse("vivechak_save_plan", map[string]any{
			"project_root":     cmpDir,
			"scope":            "comparison",
			"content": cmpPlan,
		})

		// 3. next_session
		envNext := callAndParse("vivechak_next_session", map[string]any{
			"project_root": cmpDir,
		})
		if envNext.Data["session_id"] != "CMP-01" {
			t.Errorf("expected session CMP-01, got: %v", envNext.Data["session_id"])
		}

		// 4. save_session with WEP matrix and Grade A evidence
		callAndParse("vivechak_save_session", map[string]any{
			"project_root": cmpDir,
			"session_id":   "CMP-01",
			"content": `---
session_id: CMP-01
title: Message Broker Comparison
date: 2026-10-01
status: complete
---
# Message Broker Comparison

## Recommendation
NATS Core is chosen over RabbitMQ due to minimal operational overhead and 10x throughput. Grade A (benchmark)

## Weighted Evaluation Protocol (WEP) Matrix
| Criteria | Weight | NATS | RabbitMQ |
|---|---|---|---|
| Throughput | 0.4 | 9 (Grade A) | 6 (Grade A) |
| Simplicity | 0.3 | 9 (Grade B) | 5 (Grade B) |
| Latency | 0.3 | 9 (Grade A) | 7 (Grade A) |
| **Weighted Total** | **1.0** | **9.0** | **6.0** |

## Concerns
NATS at-most-once delivery requires JetStream for durability. Grade A (docs)
`,
		})

		// 5. validate
		envVal := callAndParse("vivechak_validate", map[string]any{"project_root": cmpDir})
		if !envVal.Success {
			t.Fatalf("comparison validate failed: %s", envVal.Message)
		}

		// 6. run_gate
		envGate := callAndParse("vivechak_run_gate", map[string]any{
			"project_root": cmpDir,
			"verbose":      true,
		})
		if envGate.Data["gate_passed"] != true {
			t.Fatalf("expected comparison gate PASS, got: %v warnings=%v", envGate.Data, envGate.Warnings)
		}

		// 7. CLI doctor
		docCmd := exec.CommandContext(ctx, binPath, "doctor", cmpDir)
		if docOut, err := docCmd.CombinedOutput(); err != nil {
			t.Fatalf("doctor failed on comparison scope: %v\noutput: %s", err, string(docOut))
		}
	})

	// =========================================================================
	// SCENARIO 4: SYNTHESIS CONTEXT DENSITY & BUDGETING STRESS TEST
	// =========================================================================
	t.Run("SynthesisContextDensity_BudgetingStressTest", func(t *testing.T) {
		stressDir := t.TempDir()

		// 1. Init project
		callAndParse("vivechak_init", map[string]any{
			"project_root": stressDir,
			"scope":        "project",
		})

		// 2. Generate 10 completed session files, each with 5KB findings (50KB total > 20KB budget)
		sessionsDir := filepath.Join(stressDir, core.SessionsDir)
		var sessionHeaders strings.Builder
		for i := 1; i <= 10; i++ {
			id := fmt.Sprintf("R-%02d", i)
			sessionHeaders.WriteString(fmt.Sprintf(`#### %s: Topic %d
| Field | Value |
|---|---|
| **ID** | %s |
| **Layer** | 0 |
| **Output File** | sessions/%s.md |
`+"```prompt\n# Brief\n```\n\n", id, i, id, id))

			// Write session file
			findingBlock := strings.Repeat(fmt.Sprintf("Findings and benchmarks for session %s. Grade A (empirical measurement). ", id), 50)
			content := fmt.Sprintf(`---
session_id: %s
title: Research Topic %d
date: 2026-10-01
status: complete
tags: [tag%d, benchmark]
---
# Research Topic %d
## Recommendation
Adopt Technology %d. Grade A (benchmark)

## Alternatives Considered
%s

## Findings
%s
`, id, i, i, i, i, findingBlock, findingBlock)

			_ = os.WriteFile(filepath.Join(sessionsDir, id+".md"), []byte(content), 0o644)
		}

		// Add SYN-01
		sessionHeaders.WriteString(`#### SYN-01: Synthesis
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Layer** | 1 |
| **Output File** | research/FAD.md |
` + "```prompt\n# Brief\n[ALL_SESSION_FINDINGS]\n```\n")

		pipelineStr := "# Stress Pipeline\n\n## Sessions\n\n" + sessionHeaders.String()
		callAndParse("vivechak_save_plan", map[string]any{
			"project_root":     stressDir,
			"content": pipelineStr,
		})

		// 3. Call vivechak_next_session for synthesis
		envSyn := callAndParse("vivechak_next_session", map[string]any{
			"project_root": stressDir,
			"session_id":   "SYN-01",
		})
		if !envSyn.Success {
			t.Fatalf("next_session synthesis failed: %s", envSyn.Message)
		}

		prompt := envSyn.Data["prompt"].(string)

		// 4. Verify technology matrix contains all 10 sessions
		for i := 1; i <= 10; i++ {
			id := fmt.Sprintf("R-%02d", i)
			if !strings.Contains(prompt, id) {
				t.Errorf("technology matrix missing %s", id)
			}
		}

		// 5. Verify trimming notices exist due to 50KB > 20KB budget
		trimmedNotices := strings.Count(prompt, "trimmed for synthesis context budget")
		if trimmedNotices == 0 {
			t.Errorf("expected trimming notices when 50KB findings injected into 20KB budget")
		}
	})

	// =========================================================================
	// SCENARIO 5: CLI SUBCOMMANDS & INTEGRITY DIAGNOSTICS
	// =========================================================================
	t.Run("CLISubcommands_And_Diagnostics", func(t *testing.T) {
		// 1. version
		verCmd := exec.CommandContext(ctx, binPath, "version")
		verOut, err := verCmd.CombinedOutput()
		if err != nil || !strings.Contains(string(verOut), "vivechak") {
			t.Fatalf("version failed: %v, out: %s", err, string(verOut))
		}

		// 2. mcp-config
		cfgCmd := exec.CommandContext(ctx, binPath, "mcp-config")
		cfgOut, err := cfgCmd.CombinedOutput()
		if err != nil || !strings.Contains(string(cfgOut), "mcpServers") {
			t.Fatalf("mcp-config failed: %v, out: %s", err, string(cfgOut))
		}

		// 3. setup --help
		setupCmd := exec.CommandContext(ctx, binPath, "setup", "--help")
		setupOut, err := setupCmd.CombinedOutput()
		if err != nil || !strings.Contains(string(setupOut), "Available Presets") {
			t.Fatalf("setup --help failed: %v, out: %s", err, string(setupOut))
		}

		// 4. doctor on uninitialized workspace
		emptyDir := t.TempDir()
		docEmptyCmd := exec.CommandContext(ctx, binPath, "doctor", emptyDir)
		docEmptyOut, docErr := docEmptyCmd.CombinedOutput()
		if docErr == nil {
			t.Fatalf("doctor should fail on uninitialized workspace, got: %s", string(docEmptyOut))
		}
		if !strings.Contains(string(docEmptyOut), "does not exist") {
			t.Errorf("doctor missing expected error: %s", string(docEmptyOut))
		}
	})
}
