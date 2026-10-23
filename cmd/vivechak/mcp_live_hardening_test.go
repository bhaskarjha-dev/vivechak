package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMCPLive_Hardening consolidates all edge cases, boundary conditions, defect regressions,
// and security confinement defenses over live stdio MCP against the compiled Vivechak binary.
func TestMCPLive_Hardening(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live MCP subprocess hardening test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	binPath := resolveOrBuildTestBinary(t)

	// =========================================================================
	// 1. CODE FENCE PARSER IMMUNITY (TILDE & 4-BACKTICK BLOCKS)
	// =========================================================================
	t.Run("01_CodeFence_Immunity", func(t *testing.T) {
		wsDir := t.TempDir()
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, wsDir)
		t.Cleanup(cleanup)

		callTool := func(name string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
			if err != nil {
				t.Fatalf("call to %s failed: %v", name, err)
			}
			return parseTestEnvelope(t, res)
		}

		// Init comparison workspace
		callTool("vivechak_init", map[string]any{"project_root": wsDir, "scope": "comparison"})

		// Save session containing nested ~~~~markdown and ````go blocks
		sessionMD := `---
session_id: C-001
title: BadgerDB vs Pebble
date: 2026-10-18
status: complete
---
## Prior
BadgerDB is pure Go, Pebble is CockroachDB's RocksDB replacement.

## Key Findings
- BadgerDB uses WiscKey separation. Grade A (https://github.com/dgraph-io/badger)
- Pebble provides iterator bounds. Grade A (https://github.com/cockroachdb/pebble)

Here is sample markdown inside tilde fences that must NOT prematurely end the section:
~~~~markdown
## Sources
| Grade | Status |
| Grade A | fetched |
~~~~

And sample code inside 4-backtick code blocks:
` + "````go" + `
// ## Evaluated Options
func main() {}
` + "````" + `

## Discovered Concerns
BadgerDB value log garbage collection creates noticeable I/O spikes under sustained write loads.

## Delta
| Prior Belief | Status | Evidence | Impact |
| BadgerDB is fast | Updated | Grade A (https://github.com/dgraph-io/badger) | High |

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Status | Supports |
| 1 | https://github.com/dgraph-io/badger | Grade A | fresh | fetched | BadgerDB |
| 2 | https://github.com/cockroachdb/pebble | Grade A | fresh | fetched | Pebble |
`
		saveEnv := callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "C-001",
			"content":      sessionMD,
		})
		if !saveEnv.Success {
			t.Fatalf("save_session failed: %s", saveEnv.Message)
		}

		// Verify session parsed completely
		valEnv := callTool("vivechak_validate", map[string]any{"project_root": wsDir})
		if !valEnv.Success {
			t.Fatalf("validate failed: %s", valEnv.Message)
		}
	})

	// =========================================================================
	// 2. DISCOVERED CONCERNS LENGTH THRESHOLD BOUNDARY (49 vs 50 CHARS)
	// =========================================================================
	t.Run("02_DiscoveredConcerns_LengthBoundary", func(t *testing.T) {
		wsDir := t.TempDir()
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, wsDir)
		t.Cleanup(cleanup)

		callTool := func(name string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
			if err != nil {
				t.Fatalf("call to %s failed: %v", name, err)
			}
			return parseTestEnvelope(t, res)
		}

		callTool("vivechak_init", map[string]any{"project_root": wsDir, "scope": "project"})

		// 49 characters under ## Discovered Concerns (must remain draft for One-Way Door)
		shortSess := `---
session_id: T1-01
title: Auth Strategy
date: 2026-10-18
door_type: one-way
status: complete
---
## Key Findings
- Finding Grade A (https://example.com)

## Discovered Concerns
1234567890123456789012345678901234567890123456789
`
		shortEnv := callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "T1-01",
			"content":      shortSess,
		})
		if shortEnv.Data["status"] != "draft" {
			t.Errorf("expected session with 49 chars of concerns to remain draft, got: %v", shortEnv.Data["status"])
		}

		// 50 characters under ## Discovered Concerns (meets threshold)
		validSess := strings.Replace(shortSess,
			"1234567890123456789012345678901234567890123456789",
			"12345678901234567890123456789012345678901234567890", 1)
		validEnv := callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "T1-01",
			"content":      validSess,
		})
		if validEnv.Data["status"] != "valid" && validEnv.Data["status"] != "valid-with-warnings" {
			t.Errorf("expected session with 50 chars of concerns to be valid, got: %v", validEnv.Data["status"])
		}
		if validEnv.Data["validation_passed"] != true {
			t.Errorf("expected validation_passed to be true, got: %v", validEnv.Data["validation_passed"])
		}
	})

	// =========================================================================
	// 3. IN-PROCESS WORKSPACE SESSION MEMORY (E-13)
	// =========================================================================
	t.Run("03_WorkspaceSessionMemory_EmptyProjectRoot", func(t *testing.T) {
		wsDir := t.TempDir()
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, wsDir)
		t.Cleanup(cleanup)

		callTool := func(name string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
			if err != nil {
				t.Fatalf("call to %s failed: %v", name, err)
			}
			return parseTestEnvelope(t, res)
		}

		// Initialize workspace with explicit path
		initEnv := callTool("vivechak_init", map[string]any{
			"project_root": wsDir,
			"scope":        "project",
		})
		if !initEnv.Success {
			t.Fatalf("init failed: %s", initEnv.Message)
		}

		// Call status with project_root explicitly omitted ("")
		statEnv := callTool("vivechak_status", map[string]any{})
		if !statEnv.Success {
			t.Fatalf("vivechak_status failed with session memory: %s", statEnv.Message)
		}
		resolvedRoot, _ := statEnv.Data["root"].(string)
		if !strings.EqualFold(filepath.Clean(resolvedRoot), filepath.Clean(wsDir)) {
			t.Errorf("expected session memory to resolve %q, got %q", wsDir, resolvedRoot)
		}
	})

	// =========================================================================
	// 4. TOPOLOGY VS PARSED SESSION PARITY & SINK GUARD (E-11 & E-12)
	// =========================================================================
	t.Run("04_ParserSessionMismatch_And_SinkGuard", func(t *testing.T) {
		wsDir := t.TempDir()
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, wsDir)
		t.Cleanup(cleanup)

		callTool := func(name string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
			if err != nil {
				t.Fatalf("call to %s failed: %v", name, err)
			}
			return parseTestEnvelope(t, res)
		}

		callTool("vivechak_init", map[string]any{"project_root": wsDir, "scope": "project"})

		// Plan declares T1-01 and T1-02 in Mermaid DAG, but only provides prompt for T1-01
		partialPlan := `# Research Pipeline: Partial Test
## Pipeline Overview
` + "```mermaid\ngraph TD\n  T1-01 --> SYN-01\n  T1-02 --> SYN-01\n```\n" + `
## Sessions
### T1-01: Engine Spike
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Dependencies** | None |

` + "```prompt\n# Brief for T1-01\n```\n"

		savePartial := callTool("vivechak_save_plan", map[string]any{
			"project_root": wsDir,
			"content":      partialPlan,
		})
		if !savePartial.Success {
			t.Fatalf("save_plan failed: %s", savePartial.Message)
		}
		hasMismatchWarning := false
		for _, w := range savePartial.Warnings {
			if strings.Contains(w, "W-PARSER-SESSION-MISMATCH") && strings.Contains(w, "T1-02") {
				hasMismatchWarning = true
				break
			}
		}
		if !hasMismatchWarning {
			t.Errorf("expected W-PARSER-SESSION-MISMATCH warning for T1-02, got: %v", savePartial.Warnings)
		}

		// Plan with unreachable sink node (T1-02 disconnected from SYN-01)
		sinkPlan := `# Research Pipeline: Full Test
## Execution DAG
` + "```mermaid\ngraph TD\n  T1-01 --> SYN-01\n  T1-02\n```\n" + `
## Sessions
### T1-01: Engine Spike
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Dependencies** | None |

` + "```prompt\n# Brief T1-01\n```\n" + `

### T1-02: Isolated Session
| Field | Value |
|---|---|
| **ID** | T1-02 |
| **Dependencies** | None |

` + "```prompt\n# Brief T1-02\n```\n" + `

### SYN-01: Synthesis
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Dependencies** | T1-01 |

` + "```prompt\n# Brief SYN-01\n```\n"

		saveSink := callTool("vivechak_save_plan", map[string]any{
			"project_root": wsDir,
			"content":      sinkPlan,
		})
		if !saveSink.Success {
			t.Fatalf("save_plan sink test failed: %s", saveSink.Message)
		}

		// Save completed session for T1-01
		s1Content := `---
session_id: T1-01
title: Engine Spike
date: 2026-10-18
status: complete
door_type: two-way
---
## Key Findings
- Storage finding Grade A (https://example.com)
`
		callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "T1-01",
			"content":      s1Content,
		})

		// Call next_session: Even though T1-01 (upstream of SYN-01) is complete,
		// SYN-01 MUST NOT be actionable because T1-02 (non-synthesis session) is still pending!
		nextEnv := callTool("vivechak_next_session", map[string]any{"project_root": wsDir})
		if !nextEnv.Success {
			t.Fatalf("next_session failed: %s", nextEnv.Message)
		}
		nextID, _ := nextEnv.Data["session_id"].(string)
		if nextID == "SYN-01" {
			t.Fatalf("Terminal synthesis sink guard failed: SYN-01 returned while T1-02 was still pending!")
		}
		if nextID != "T1-02" {
			t.Errorf("expected ready session T1-02, got %s", nextID)
		}
	})

	// =========================================================================
	// 5. NATURAL NUMERIC SORTING OF UNANCHORED ADRs (D-1, D-2, D-10)
	// =========================================================================
	t.Run("05_NaturalNumericSorting_UnanchoredADRs", func(t *testing.T) {
		wsDir := t.TempDir()
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, wsDir)
		t.Cleanup(cleanup)

		callTool := func(name string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
			if err != nil {
				t.Fatalf("call to %s failed: %v", name, err)
			}
			return parseTestEnvelope(t, res)
		}

		callTool("vivechak_init", map[string]any{"project_root": wsDir, "scope": "project"})

		// Save D-10, D-1, D-2 out of order
		for _, id := range []string{"D-10", "D-1", "D-2"} {
			adr := `---
decision_id: ` + id + `
title: Decision ` + id + `
date: 2026-10-18
status: accepted
door_type: two-way
---
# ` + id + `: Title
## Decision
Chose option for ` + id + `. Grade A (docs)
`
			callTool("vivechak_record_decision", map[string]any{
				"project_root": wsDir,
				"decision_id":  id,
				"content":      adr,
			})
		}

		// Read DECISIONS.md and verify natural sorting: D-1, D-2, D-10
		decData, err := os.ReadFile(filepath.Join(wsDir, core.ResearchDir, "DECISIONS.md"))
		if err != nil {
			t.Fatalf("reading DECISIONS.md: %v", err)
		}
		content := string(decData)

		idxD1 := strings.Index(content, "D-1")
		idxD2 := strings.Index(content, "D-2")
		idxD10 := strings.Index(content, "D-10")

		if idxD1 == -1 || idxD2 == -1 || idxD10 == -1 {
			t.Fatalf("missing decision entries in DECISIONS.md:\n%s", content)
		}
		if idxD1 >= idxD2 || idxD2 >= idxD10 {
			t.Errorf("expected natural order D-1 < D-2 < D-10, got positions: D-1=%d, D-2=%d, D-10=%d",
				idxD1, idxD2, idxD10)
		}
	})

	// =========================================================================
	// 6. DIAMOND DAG WITH MIXED-CASE DEPENDENCIES
	// =========================================================================
	t.Run("06_DiamondDAG_MixedCaseDependencies", func(t *testing.T) {
		wsDir := t.TempDir()
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, wsDir)
		t.Cleanup(cleanup)

		callTool := func(name string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
			if err != nil {
				t.Fatalf("call to %s failed: %v", name, err)
			}
			return parseTestEnvelope(t, res)
		}

		callTool("vivechak_init", map[string]any{"project_root": wsDir, "scope": "project"})

		// Diamond DAG: T0-01 -> t0-02 & T0-03 -> SYN-01
		diamondPlan := `# Diamond Pipeline
## Sessions
### T0-01: Root
| Field | Value |
|---|---|
| **ID** | T0-01 |
| **Dependencies** | None |
` + "```prompt\nRoot brief\n```\n" + `

### t0-02: Branch A
| Field | Value |
|---|---|
| **ID** | t0-02 |
| **Dependencies** | T0-01 |
` + "```prompt\nBranch A brief\n```\n" + `

### T0-03: Branch B
| Field | Value |
|---|---|
| **ID** | T0-03 |
| **Dependencies** | t0-01 |
` + "```prompt\nBranch B brief\n```\n" + `

### SYN-01: Synthesis
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Dependencies** | T0-02, t0-03 |
` + "```prompt\nSynthesis brief\n```\n"

		saveEnv := callTool("vivechak_save_plan", map[string]any{
			"project_root": wsDir,
			"content":      diamondPlan,
		})
		if !saveEnv.Success {
			t.Fatalf("diamond DAG save_plan failed: %s", saveEnv.Message)
		}

		// Next session must resolve T0-01 as the ready root
		nextEnv := callTool("vivechak_next_session", map[string]any{"project_root": wsDir})
		if !nextEnv.Success || !strings.EqualFold(nextEnv.Data["session_id"].(string), "T0-01") {
			t.Fatalf("expected next session T0-01, got: %+v", nextEnv)
		}
	})

	// =========================================================================
	// 7. PREMORTEM PLACEHOLDER & NEGATIVE GATE FAILURE MODES
	// =========================================================================
	t.Run("07_PremortemPlaceholder_And_NegativeGate", func(t *testing.T) {
		wsDir := t.TempDir()
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, wsDir)
		t.Cleanup(cleanup)

		callTool := func(name string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
			if err != nil {
				t.Fatalf("call to %s failed: %v", name, err)
			}
			return parseTestEnvelope(t, res)
		}

		callTool("vivechak_init", map[string]any{"project_root": wsDir, "scope": "decision"})

		// Save One-Way Door ADR missing review trigger, alternatives, and premortem with TODO
		badADR := `---
decision_id: D-001
title: Incomplete Decision
date: 2026-10-18
status: accepted
door_type: one-way
---
# D-001: Incomplete Decision
## Decision
Chose option without justification.

## Premortem
TODO: add premortem analysis before production release.
`
		callTool("vivechak_record_decision", map[string]any{
			"project_root": wsDir,
			"decision_id":  "D-001",
			"content":      badADR,
		})

		// Run gate on bad ADR: must fail Track B
		gateEnv := callTool("vivechak_run_gate", map[string]any{
			"project_root": wsDir,
			"track":        "track_b",
		})
		if gateEnv.Data["gate_passed"] == true {
			t.Errorf("expected gate to fail for incomplete One-Way Door ADR, but passed: %+v", gateEnv.Data)
		}
		if gateEnv.Data["gate_status"] != "FAIL" {
			t.Errorf("expected gate status 'FAIL', got: %v", gateEnv.Data["gate_status"])
		}
	})

	// =========================================================================
	// 8. SECURITY CONFINEMENT & PATH TRAVERSAL DEFENSES
	// =========================================================================
	t.Run("08_SecurityConfinement_PathTraversalDefenses", func(t *testing.T) {
		wsDir := t.TempDir()
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, wsDir)
		t.Cleanup(cleanup)

		callTool := func(name string, args map[string]any) testEnvelope {
			t.Helper()
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
			if err != nil {
				t.Fatalf("call to %s failed: %v", name, err)
			}
			return parseTestEnvelope(t, res)
		}

		callTool("vivechak_init", map[string]any{"project_root": wsDir, "scope": "project"})

		// Traversal in session_id
		env1 := callTool("vivechak_save_session", map[string]any{
			"project_root": wsDir,
			"session_id":   "../../evil_session",
			"content":      "# Malicious session",
		})
		if env1.Success {
			t.Error("expected traversal in session_id to be rejected")
		}

		// Traversal in decision_id
		env2 := callTool("vivechak_record_decision", map[string]any{
			"project_root": wsDir,
			"decision_id":  "../../../etc_passwd",
			"content":      "# Malicious ADR",
		})
		if env2.Success {
			t.Error("expected traversal in decision_id to be rejected")
		}
	})

	// =========================================================================
	// 9. SPECIAL RESEARCH FILES WHITELISTING
	// =========================================================================
	t.Run("09_SpecialResearchFiles_Whitelisting", func(t *testing.T) {
		// Verify core.IsSpecialResearchFile recognizes special files vs standard ADRs
		special := []string{
			"D-001-conflict-resolution.md",
			"D-002-plan.md",
			"C-001-comparison.md",
			"FOUNDING-ARCHITECTURE.md",
			"PHASE-0-GATE.md",
			"DECISIONS.md",
			"FAD.md",
			"RESEARCH-PIPELINE.md",
		}
		for _, f := range special {
			if !core.IsSpecialResearchFile(f) {
				t.Errorf("expected IsSpecialResearchFile(%q) to be true", f)
			}
		}

		nonSpecial := []string{
			"D-001.md",
			"D-002.md",
			"D-ARCH-01.md",
			"NOTES.md",
			"TODO.md",
			"scratch.txt",
		}
		for _, f := range nonSpecial {
			if core.IsSpecialResearchFile(f) {
				t.Errorf("expected IsSpecialResearchFile(%q) to be false", f)
			}
		}
	})

	// =========================================================================
	// 10. CLI SCOPE INFERENCE BATTERY
	// =========================================================================
	t.Run("10_ScopeInference_Battery", func(t *testing.T) {
		cases := []struct {
			input string
			want  core.Scope
		}{
			{"vck plan: Distributed Database Engine", core.ScopeProject},
			{"vck decide: BadgerDB vs Pebble", core.ScopeDecision},
			{"vck compare: RocksDB versus LevelDB", core.ScopeComparison},
			{"vck vs: Redis vs Memcached", core.ScopeComparison},
			{"should we choose PostgreSQL or MySQL?", core.ScopeDecision},
			{"compare grpc and rest", core.ScopeComparison},
			{"hey im confused bw winston and pino for my express app which one shud i use?? need fast json logging pls help", core.ScopeComparison},
			{"bro shuld i use react 19 server actions or tanstack react query for fetching data in my nextjs dashboard? im totally confused", core.ScopeDecision},
			{"full architecture for cloud native stream processing", core.ScopeProject},
		}

		for _, tc := range cases {
			got := core.InferScope(tc.input)
			if got != tc.want {
				t.Errorf("core.InferScope(%q) = %q, want %q", tc.input, got, tc.want)
			}
		}
	})

	// =========================================================================
	// 11. LIVE MCP PROMPTS DISCOVERY & WIRE-LEVEL ALIAS NORMALIZATION OVER STDIO
	// =========================================================================
	t.Run("11_LiveMCP_PromptsDiscovery_And_Aliases", func(t *testing.T) {
		wsDir := t.TempDir()
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, wsDir)
		t.Cleanup(cleanup)

		// 1. Verify prompts/list returns strictly the 4 canonical action verbs
		listRes, err := cs.ListPrompts(ctx, nil)
		if err != nil {
			t.Fatalf("ListPrompts failed: %v", err)
		}

		expectedPrompts := map[string]string{
			"compare":  "Compare Options (WEP Matrix)",
			"decision": "Architectural Decision (ADR)",
			"project":  "Project Pipeline (Full FAD)",
			"gate":     "Phase 0 Pre-Codebase Exit Gate",
		}

		if len(listRes.Prompts) != len(expectedPrompts) {
			t.Fatalf("expected %d prompts, got %d", len(expectedPrompts), len(listRes.Prompts))
		}

		for _, p := range listRes.Prompts {
			expectedTitle, ok := expectedPrompts[p.Name]
			if !ok {
				t.Errorf("unexpected prompt registered in live binary: %s", p.Name)
				continue
			}
			if p.Title != expectedTitle {
				t.Errorf("prompt %s: expected title %q, got %q", p.Name, expectedTitle, p.Title)
			}
		}

		// 2. Verify prompts/get with canonical, legacy (vivechak_*), and shorthand (vck_*) aliases
		triplets := []struct {
			canonical string
			legacy    string
			shorthand string
			args      map[string]string
		}{
			{
				canonical: "compare",
				legacy:    "vivechak_compare",
				shorthand: "vck_compare",
				args:      map[string]string{"options": "BadgerDB vs Pebble"},
			},
			{
				canonical: "decision",
				legacy:    "vivechak_decision",
				shorthand: "vck_decision",
				args:      map[string]string{"dilemma": "Migrate store to BadgerDB"},
			},
			{
				canonical: "project",
				legacy:    "vivechak_project",
				shorthand: "vck_project",
				args:      map[string]string{"vision": "Distributed Storage Engine"},
			},
			{
				canonical: "gate",
				legacy:    "vivechak_gate",
				shorthand: "vck_gate",
				args:      map[string]string{"track": "all"},
			},
		}

		for _, trip := range triplets {
			rCanon, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: trip.canonical, Arguments: trip.args})
			if err != nil {
				t.Fatalf("GetPrompt(%s) over live stdio failed: %v", trip.canonical, err)
			}
			rLeg, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: trip.legacy, Arguments: trip.args})
			if err != nil {
				t.Fatalf("GetPrompt(%s) over live stdio failed: %v", trip.legacy, err)
			}
			rShort, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: trip.shorthand, Arguments: trip.args})
			if err != nil {
				t.Fatalf("GetPrompt(%s) over live stdio failed: %v", trip.shorthand, err)
			}

			tcCanon, ok1 := rCanon.Messages[0].Content.(*mcp.TextContent)
			tcLeg, ok2 := rLeg.Messages[0].Content.(*mcp.TextContent)
			tcShort, ok3 := rShort.Messages[0].Content.(*mcp.TextContent)

			if !ok1 || !ok2 || !ok3 {
				t.Fatalf("expected TextContent across all prompt variations")
			}

			if tcCanon.Text != tcLeg.Text {
				t.Errorf("live wire prompt %s != %s", trip.canonical, trip.legacy)
			}
			if tcCanon.Text != tcShort.Text {
				t.Errorf("live wire prompt %s != %s", trip.canonical, trip.shorthand)
			}
		}
	})
}
