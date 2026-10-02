package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestInit_GitignoreCreation verifies .gitignore creation in research/
func TestInit_GitignoreCreation(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Fresh init creates research/.gitignore
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("init unsuccessful: %s", env.Message)
	}

	gitignorePath := filepath.Join(tmpDir, core.ResearchDir, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("reading .gitignore: %v", err)
	}
	if !strings.Contains(string(data), "*.lock") {
		t.Errorf("expected *.lock in .gitignore, got: %s", string(data))
	}

	// 2. Custom .gitignore is NOT overwritten on subsequent init
	customContent := "# Custom ignore\n*.tmp\n*.log\n"
	if err := os.WriteFile(gitignorePath, []byte(customContent), 0o644); err != nil {
		t.Fatalf("writing custom .gitignore: %v", err)
	}

	res2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("second init failed: %v", err)
	}
	_ = parseEnvelope(t, res2)

	data2, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("reading custom .gitignore: %v", err)
	}
	if string(data2) != customContent {
		t.Errorf("expected custom .gitignore to be preserved, got:\n%s", string(data2))
	}
}

// TestWrapFrontmatterForRegistry verifies frontmatter transformation into collapsible details
func TestWrapFrontmatterForRegistry(t *testing.T) {
	t.Run("converts frontmatter to details with summary", func(t *testing.T) {
		input := `---
id: D-001
title: "Database Strategy"
door_type: one-way
status: accepted
---

# D-001: Database Strategy
Decision content here.`

		wrapped := wrapFrontmatterForRegistry(input)
		if !strings.HasPrefix(wrapped, "<details>\n<summary>") {
			t.Errorf("expected wrapped output to start with <details><summary>, got:\n%s", wrapped)
		}
		if !strings.Contains(wrapped, "<strong>D-001</strong>") {
			t.Errorf("missing strong ID in summary: %s", wrapped)
		}
		if !strings.Contains(wrapped, "Database Strategy") {
			t.Errorf("missing title in summary: %s", wrapped)
		}
		if !strings.Contains(wrapped, "<code>one-way</code>") {
			t.Errorf("missing door_type in summary: %s", wrapped)
		}
		if !strings.Contains(wrapped, "```yaml") || !strings.Contains(wrapped, "```\n\n</details>") {
			t.Errorf("missing yaml code fence inside details: %s", wrapped)
		}
		if !strings.Contains(wrapped, "# D-001: Database Strategy\nDecision content here.") {
			t.Errorf("missing body content: %s", wrapped)
		}
	})

	t.Run("content without frontmatter is unchanged", func(t *testing.T) {
		input := "# Just Markdown Header\nSome regular body."
		wrapped := wrapFrontmatterForRegistry(input)
		if wrapped != input {
			t.Errorf("expected unchanged content, got:\n%s", wrapped)
		}
	})
}

// TestRecordDecision_AutoDraft verifies auto-drafting from session findings
func TestRecordDecision_AutoDraft(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	sessionContent := `---
session_id: R-01
title: Key-Value Storage
date: 2026-10-01
status: complete
---

# Key-Value Storage

## Evaluated Options
- RocksDB: Embedded LSM-tree. Grade A (direct benchmark)
- BadgerDB: Pure Go alternative. Grade B (docs)

## Recommendations
We recommend RocksDB for production write-heavy workload. Grade A (benchmark)

## Discovered Concerns & Failure Modes
- Compaction latency spikes under heavy concurrent writes.
`
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "R-01",
			"content":      sessionContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session: %v", err)
	}

	// Call record_decision with auto_draft_from="R-01" and empty content
	resDraft, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":    tmpDir,
			"decision_id":     "D-001",
			"auto_draft_from": "R-01",
		},
	})
	if err != nil {
		t.Fatalf("record_decision auto_draft: %v", err)
	}
	envDraft := parseEnvelope(t, resDraft)
	if !envDraft.Success {
		t.Fatalf("expected auto_draft success, got: %s", envDraft.Message)
	}

	draftContent, ok := envDraft.Data.(map[string]any)["draft_content"].(string)
	if !ok || draftContent == "" {
		t.Fatalf("expected draft_content in response data, got: %v", envDraft.Data)
	}

	if !strings.Contains(draftContent, "id: D-001") {
		t.Errorf("draft missing id: %s", draftContent)
	}
	if !strings.Contains(draftContent, "RocksDB") {
		t.Errorf("draft missing recommendation content: %s", draftContent)
	}
	if !strings.Contains(draftContent, "Compaction latency") {
		t.Errorf("draft missing concerns: %s", draftContent)
	}
	if !strings.Contains(envDraft.NextStep, "Review the auto-drafted decision") {
		t.Errorf("expected NextStep to prompt review, got: %s", envDraft.NextStep)
	}

	// Now save the final decision with the content
	resSave, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"decision_id":  "D-001",
			"content":      draftContent,
		},
	})
	if err != nil {
		t.Fatalf("record_decision save: %v", err)
	}
	envSave := parseEnvelope(t, resSave)
	if !envSave.Success {
		t.Fatalf("save failed: %s", envSave.Message)
	}

	// Verify DECISIONS.md contains <details> wrapper
	decPath := filepath.Join(tmpDir, core.DecisionsFile)
	decData, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("reading DECISIONS.md: %v", err)
	}
	if !strings.Contains(string(decData), "<details>") || !strings.Contains(string(decData), "<strong>D-001</strong>") {
		t.Errorf("expected DECISIONS.md to contain collapsible details block, got:\n%s", string(decData))
	}
}

// TestAmendSession verifies vivechak_amend_session functionality
func TestAmendSession(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	sessionContent := `---
session_id: R-01
title: Initial Storage Analysis
date: 2026-10-01
status: complete
---

# Storage Findings
Initial conclusion was X. Grade A (direct analysis)
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "R-01",
			"content":      sessionContent,
		},
	})

	// 1. Amend existing session
	resAmend, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_amend_session",
		Arguments: map[string]any{
			"project_root":        tmpDir,
			"session_id":          "R-01",
			"amending_session_id": "R-03",
			"amendment":           "R-03 revealed that write-heavy workload requires WAL tuning. Revising IOPS estimate.",
		},
	})
	if err != nil {
		t.Fatalf("amend_session failed: %v", err)
	}
	envAmend := parseEnvelope(t, resAmend)
	if !envAmend.Success {
		t.Fatalf("amend_session unsuccessful: %s", envAmend.Message)
	}

	// Verify file content on disk
	filePath := filepath.Join(tmpDir, core.SessionsDir, "R-01.md")
	amendedData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("reading amended file: %v", err)
	}
	amendedStr := string(amendedData)
	if !strings.Contains(amendedStr, "Initial conclusion was X") {
		t.Errorf("original content was wiped! got: %s", amendedStr)
	}
	if !strings.Contains(amendedStr, "## Post-Hoc Amendment (appended by R-03)") {
		t.Errorf("missing amendment header: %s", amendedStr)
	}
	if !strings.Contains(amendedStr, "write-heavy workload requires WAL tuning") {
		t.Errorf("missing amendment text: %s", amendedStr)
	}
	if !strings.Contains(amendedStr, "*Added:") {
		t.Errorf("missing timestamp: %s", amendedStr)
	}

	// 2. Amend non-existent session fails
	resBad, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_amend_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "R-99",
			"amendment":    "Some amendment",
		},
	})
	envBad := parseEnvelope(t, resBad)
	if envBad.Success {
		t.Error("expected amending non-existent session to fail")
	}

	// 3. Amend with empty amendment fails
	resEmpty, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_amend_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "R-01",
			"amendment":    "   ",
		},
	})
	envEmpty := parseEnvelope(t, resEmpty)
	if envEmpty.Success {
		t.Error("expected empty amendment to fail")
	}
}

// TestNextSession_LayerTransitionReplan verifies layer transition replan advisory prompt
func TestNextSession_LayerTransitionReplan(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	multiLayerPlan := `# Research Pipeline
## Session DAG

### Session L0-01 — Foundation A
| Field | Value |
|---|---|
| **ID** | L0-01 |
| **Layer** | 0 |
| **Dependencies** | None |

` + "```prompt" + `
Foundation research prompt for L0-01.
` + "```" + `

### Session L1-01 — App Layer
| Field | Value |
|---|---|
| **ID** | L1-01 |
| **Layer** | 1 |
| **Dependencies** | L0-01 |

` + "```prompt" + `
App layer research prompt for L1-01.
[UPSTREAM_FINDINGS]
` + "```" + `
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"content":      multiLayerPlan,
		},
	})

	// Before L0-01 completed
	res1, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	env1 := parseEnvelope(t, res1)
	if env1.Data.(map[string]any)["session_id"] != "L0-01" {
		t.Fatalf("expected L0-01, got %v", env1.Data)
	}

	// Complete L0-01
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "L0-01",
			"content": `---
session_id: L0-01
title: Foundation A
date: 2026-10-01
status: complete
---
# Content
Done. Grade A (direct docs)
`,
		},
	})

	// Next session should be L1-01 and warnings should contain the layer transition replan prompt
	res2, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	env2 := parseEnvelope(t, res2)
	if env2.Data.(map[string]any)["session_id"] != "L1-01" {
		t.Fatalf("expected L1-01, got %v", env2.Data)
	}

	hasReplanWarning := false
	for _, w := range env2.Warnings {
		if strings.Contains(w, "LAYER-TRANSITION") && strings.Contains(w, "Layer 0 complete") {
			hasReplanWarning = true
			break
		}
	}
	if !hasReplanWarning {
		t.Errorf("expected layer replan prompt in warnings, got: %v", env2.Warnings)
	}
}

// TestWorkspaceValidate verifies workspace-level validation
func TestWorkspaceValidate(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	// Save valid session
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "R-01",
			"content": `---
session_id: R-01
title: Architecture
date: 2026-10-01
status: complete
---
# Architecture Findings
Recommendation: Microservices architecture with decoupled service boundaries. Grade A (direct benchmark)
`,
		},
	})

	// Record valid decision with substantive body > 100 characters
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"decision_id":  "D-001",
			"content": `---
id: D-001
title: Architecture Choice
status: accepted
door_type: two-way
---
# Architecture Choice
## Context
We need a scalable and resilient architecture layout for high-throughput distributed microservices.

## Decision
Decided to use event-driven microservices architecture. Grade A (empirical benchmark)

## Consequences
Consequences are minimal and manageable across all deployment targets and runtime platforms.
`,
		},
	})

	// Workspace-wide validate (empty content, project_root provided)
	resVal, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_validate",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("workspace validate failed: %v", err)
	}
	envVal := parseEnvelope(t, resVal)
	if !envVal.Success {
		t.Fatalf("validation unsuccessful: %s", envVal.Message)
	}

	dataMap, ok := envVal.Data.(map[string]any)
	if !ok {
		t.Fatalf("unexpected data structure: %v", envVal.Data)
	}

	sessionsMap, ok := dataMap["sessions"].(map[string]any)
	if !ok || sessionsMap["valid"].(float64) < 1 {
		t.Errorf("expected at least 1 valid session, got: %v", dataMap["sessions"])
	}

	decisionsMap, ok := dataMap["decisions"].(map[string]any)
	if !ok || (decisionsMap["valid"].(float64) < 1 && decisionsMap["warnings"].(float64) < 1) {
		t.Errorf("expected valid decisions, got: %v", dataMap["decisions"])
	}
	if decisionsMap["errors"].(float64) != 0 {
		t.Errorf("expected 0 decision errors, got: %v; warnings: %v", decisionsMap["errors"], envVal.Warnings)
	}
}

// TestRunGate_OneWayHumanReviewedAdvisory verifies advisory warning for unreviewed one-way doors
func TestRunGate_OneWayHumanReviewedAdvisory(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "decision"},
	})

	// Save required decision session
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "S1-01",
			"content": `---
session_id: S1-01
title: Consensus Research
date: 2026-10-01
status: complete
---
# Consensus Findings
Raft protocol chosen. Grade A (formal proof)
`,
		},
	})

	// 1. One-way door with human_reviewed: false
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"decision_id":  "D-001",
			"content": `---
id: D-001
title: Irreversible Ledger
status: accepted
door_type: one-way
human_reviewed: false
review_trigger: "High contention"
---
# Context
Distributed ledger persistence under high volume transactions.
# Decision
Adopt custom Raft implementation. Grade A (direct benchmark)
# Consequences
High write throughput achieved across regions.
`,
		},
	})

	resGate1, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	envGate1 := parseEnvelope(t, resGate1)

	// Check advisory warning exists in warnings
	foundAdvisory := false
	for _, w := range envGate1.Warnings {
		if strings.Contains(w, "lack human review") && strings.Contains(w, "D-001") {
			foundAdvisory = true
			break
		}
	}
	if !foundAdvisory {
		t.Errorf("expected advisory warning for unreviewed one-way door, got warnings: %v", envGate1.Warnings)
	}

	// 2. Update with human_reviewed: true
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"decision_id":  "D-001",
			"content": `---
id: D-001
title: Irreversible Ledger
status: accepted
door_type: one-way
human_reviewed: true
review_trigger: "High contention"
---
# Context
Distributed ledger persistence under high volume transactions.
# Decision
Adopt custom Raft implementation. Grade A (direct benchmark)
# Consequences
High write throughput achieved across regions.
`,
		},
	})

	resGate2, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	envGate2 := parseEnvelope(t, resGate2)

	for _, w := range envGate2.Warnings {
		if strings.Contains(w, "lack human review") {
			t.Errorf("advisory warning should be gone when human_reviewed: true, got: %s", w)
		}
	}
}

// TestNextSession_SynthesisExemptionAndMessaging tests Task 1 and Task 12
func TestNextSession_SynthesisExemptionAndMessaging(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	// Create a large prompt (> 40,000 chars)
	largeBody := strings.Repeat("Detailed research guidelines and extensive domain constraints. ", 800)

	pipeline := `# Research Pipeline
## Session DAG

### Session R-01 — Regular Session
| Field | Value |
|---|---|
| **ID** | R-01 |
| **Dependencies** | None |

` + "```prompt\n" + largeBody + "\n```" + `

### Session SYN-01 — Synthesis Session
| Field | Value |
|---|---|
| **ID** | SYN-01 |
| **Dependencies** | R-01 |

` + "```prompt\n" + largeBody + "\n```" + `
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"content":      pipeline,
		},
	})

	// 1. Regular session R-01 with verbose=false: prompt should be truncated
	resR1, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": false},
	})
	envR1 := parseEnvelope(t, resR1)
	promptR1 := envR1.Data.(map[string]any)["prompt"].(string)
	if !strings.Contains(promptR1, "[truncated") {
		t.Fatalf("expected regular session prompt > 10K tokens to be truncated when verbose=false, got len=%d prompt=%q data=%v", len(promptR1), promptR1, envR1.Data)
	}

	// Complete R-01 with Delta section
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "R-01",
			"content": `---
session_id: R-01
title: Regular Research
date: 2026-10-01
status: complete
---
# Findings
PostgreSQL recommended. Grade A (direct docs)

## Delta
| Prior | Status | Impact |
|---|---|---|
| MongoDB was preferred | Contradicted | High |
`,
		},
	})

	// 2. Synthesis session SYN-01 with verbose=false: prompt MUST NOT be truncated (Task 1)
	resSyn, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": false},
	})
	envSyn := parseEnvelope(t, resSyn)
	promptSyn := envSyn.Data.(map[string]any)["prompt"].(string)
	if strings.Contains(promptSyn, "... [truncated]") {
		t.Errorf("synthesis session prompt should NEVER be truncated even when verbose=false")
	}

	// 3. Verify belief evolution section injected (Task 5)
	if !strings.Contains(promptSyn, "BELIEF EVOLUTION ACROSS SESSIONS") {
		t.Errorf("expected synthesis prompt to include aggregated belief evolution, got:\n%s", promptSyn)
	}
	if !strings.Contains(promptSyn, "MongoDB was preferred") {
		t.Errorf("expected delta table content to appear in belief evolution section")
	}

	// 4. Verify Task 12 NextStep guidance for synthesis
	if !strings.Contains(envSyn.NextStep, "Belief Evolution section") || !strings.Contains(envSyn.NextStep, "strongest and weakest signals") {
		t.Errorf("expected NextStep to include synthesis guidance, got: %s", envSyn.NextStep)
	}
}
