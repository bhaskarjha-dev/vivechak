package mcputil

import (
	"context"
	"fmt"
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

	// 4. Amend synthesis session FAD (targeting research/FAD.md)
	fadContent := `---
session_id: SYN-01
title: Founding Architecture Document
date: 2026-10-01
status: complete
---
# Founding Architecture Document
Core architecture findings here. Grade A (empirical benchmark)
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "SYN-01",
			"content":      fadContent,
		},
	})

	resAmendFAD, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_amend_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "FAD",
			"amendment":    "Post-synthesis review: Added rate limiting layer before ingress.",
		},
	})
	if err != nil {
		t.Fatalf("amend_session FAD failed: %v", err)
	}
	envAmendFAD := parseEnvelope(t, resAmendFAD)
	if !envAmendFAD.Success {
		t.Fatalf("amend_session FAD unsuccessful: %s", envAmendFAD.Message)
	}

	fadDiskData, err := os.ReadFile(filepath.Join(tmpDir, core.FADFile))
	if err != nil {
		t.Fatalf("reading amended FAD.md: %v", err)
	}
	if !strings.Contains(string(fadDiskData), "Added rate limiting layer before ingress") {
		t.Errorf("expected amendment in FAD.md, got:\n%s", string(fadDiskData))
	}

	// 5. Test amending upstream session with completed downstream session emits W-STALE-DOWNSTREAM
	pipelinePlan := `# Research Pipeline
## Execution DAG
### Session T0-01: Foundation
| Field | Value |
|---|---|
| **ID** | T0-01 |
| **Dependencies** | None |
| **Output File** | sessions/T0-01.md |

` + "```prompt" + `
Investigate foundation.
` + "```" + `

### Session T1-01: Deep Dive
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Dependencies** | [T0-01] |
| **Output File** | sessions/T1-01.md |

` + "```prompt" + `
Investigate deep dive.
` + "```" + `
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(pipelinePlan), 0o644)

	// Save T0-01 and T1-01 as completed sessions
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T0-01.md"), []byte(`---
id: T0-01
status: complete
---
# T0-01 Foundation
`), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T1-01.md"), []byte(`---
id: T1-01
status: complete
---
# T1-01 Deep Dive
`), 0o644)

	resStale, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_amend_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T0-01",
			"amendment":    "Found flaw in foundation assumptions.",
		},
	})
	if err != nil {
		t.Fatalf("amend_session failed: %v", err)
	}
	envStale := parseEnvelope(t, resStale)
	if !envStale.Success {
		t.Fatalf("amend_session failed: %s", envStale.Message)
	}
	staleFound := false
	for _, w := range envStale.Warnings {
		if strings.Contains(w, "W-STALE-DOWNSTREAM") && strings.Contains(w, "T1-01") {
			staleFound = true
			break
		}
	}
	if !staleFound {
		t.Errorf("expected W-STALE-DOWNSTREAM warning mentioning T1-01, got: %v", envStale.Warnings)
	}
	dataStale, ok := envStale.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got: %T", envStale.Data)
	}
	staleList, ok := dataStale["potentially_stale_sessions"].([]any)
	if !ok || len(staleList) == 0 || staleList[0] != "T1-01" {
		t.Errorf("expected potentially_stale_sessions to contain T1-01, got: %v", dataStale["potentially_stale_sessions"])
	}

	// 6. Test amending upstream session when downstream session is NOT completed (no warning)
	_ = os.Remove(filepath.Join(tmpDir, core.SessionsDir, "T1-01.md"))
	resNotStale, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_amend_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T0-01",
			"amendment":    "Another amendment while T1-01 is still pending.",
		},
	})
	if err != nil {
		t.Fatalf("amend_session failed: %v", err)
	}
	envNotStale := parseEnvelope(t, resNotStale)
	for _, w := range envNotStale.Warnings {
		if strings.Contains(w, "W-STALE-DOWNSTREAM") {
			t.Errorf("unexpected W-STALE-DOWNSTREAM warning when downstream session is not completed: %v", w)
		}
	}
	dataNotStale, ok := envNotStale.Data.(map[string]any)
	if ok && dataNotStale["potentially_stale_sessions"] != nil {
		t.Errorf("expected potentially_stale_sessions to be nil, got: %v", dataNotStale["potentially_stale_sessions"])
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

// TestGate_WhitelistScanning verifies non-ADR artifacts in research/ do not fail the gate
func TestGate_WhitelistScanning(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "decision"},
	})

	// Save decision plan
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "decision",
			"decision_id":  "D-001",
			"pipeline_content": `# Decision Plan: D-001
| Field | Value |
|---|---|
| **Decision** | D-001 |
| **Door Type** | Two-Way |
| **Output File** | sessions/S1-storage.md |
` + "```prompt\n# Brief\n```\n",
		},
	})

	// Save session S1-storage
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "S1-storage",
			"content": `---
session_id: S1-storage
title: Storage Investigation
date: 2026-10-01
status: complete
---
# Storage Investigation
Recommendation: SQLite with WAL. Grade A (official documentation)
`,
		},
	})

	// Record valid accepted ADR
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"decision_id":  "D-001",
			"content": `---
id: D-001
title: Storage Architecture
status: accepted
door_type: two-way
review_trigger: "Scale exceeds 500k writes/sec"
---
# D-001: Storage Architecture
## Context
Local-first embedded database requirement for high throughput.
## Decision
We choose SQLite WAL mode. Grade A (official documentation)
## Consequences
Single-writer constraint handled via connection queue.
`,
		},
	})

	// Plant non-ADR markdown files in research/ that should be ignored by whitelist
	notesPath := filepath.Join(tmpDir, core.ResearchDir, "NOTES.md")
	_ = os.WriteFile(notesPath, []byte(`---
status: sealed
random_field: unexpected
---
# General Research Notes
This is just notes, not an ADR.
`), 0o644)

	resGate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate failed: %v", err)
	}
	envGate := parseEnvelope(t, resGate)
	dataMap := envGate.Data.(map[string]any)

	gatePassed, _ := dataMap["gate_passed"].(bool)
	if !gatePassed {
		t.Errorf("expected gate to pass with whitelist scanning, got status %v; warnings: %v; data: %v",
			dataMap["gate_status"], envGate.Warnings, dataMap)
	}
}

// TestGate_DiagnosticErrors verifies actionable diagnostics when ADR is not accepted
func TestGate_DiagnosticErrors(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "decision"},
	})

	// Save decision plan
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "decision",
			"decision_id":  "D-001",
			"pipeline_content": `# Decision Plan: D-001
| Field | Value |
|---|---|
| **Decision** | D-001 |
| **Door Type** | Two-Way |
| **Output File** | sessions/S1-storage.md |
` + "```prompt\n# Brief\n```\n",
		},
	})

	// Record ADR with status 'proposed' (not accepted)
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"decision_id":  "D-001",
			"content": `---
id: D-001
title: Unsettled Choice
status: proposed
door_type: two-way
---
# D-001: Unsettled Choice
## Context
Need to pick something with adequate rationale.
## Decision
Draft choice. Grade B (estimate)
## Consequences
Pending final review.
`,
		},
	})

	resGate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate failed: %v", err)
	}
	envGate := parseEnvelope(t, resGate)

	foundDiagnostic := false
	for _, warning := range envGate.Warnings {
		if strings.Contains(warning, "D-001") &&
			strings.Contains(warning, "proposed") &&
			strings.Contains(warning, "status to 'accepted'") {
			foundDiagnostic = true
			break
		}
	}
	if !foundDiagnostic {
		t.Errorf("expected actionable diagnostic error with file/id/status/hint in warnings, got warnings: %v", envGate.Warnings)
	}
}

// TestWorkspaceValidate_DecisionSingleSource verifies exact count and no double-counting of DECISIONS.md
func TestWorkspaceValidate_DecisionSingleSource(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	// Record two decisions
	for _, id := range []string{"D-001", "D-002"} {
		_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"project_root": tmpDir,
				"decision_id":  id,
				"content": fmt.Sprintf(`---
id: %s
title: Decision %s
status: accepted
door_type: two-way
---
# %s: Decision %s
## Context
Context description for %s with sufficient length.
## Decision
Recommendation for %s. Grade A (empirical benchmark)
## Consequences
Manageable consequences for %s.
`, id, id, id, id, id, id, id),
			},
		})
	}

	// Verify DECISIONS.md was compiled
	if _, err := os.Stat(filepath.Join(tmpDir, core.DecisionsFile)); err != nil {
		t.Fatalf("DECISIONS.md should exist after record_decision: %v", err)
	}

	// Plant non-ADR files
	_ = os.WriteFile(filepath.Join(tmpDir, core.ResearchDir, "NOTES.md"), []byte("# Notes\nNot an ADR"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.ResearchDir, "D-001-plan.md"), []byte("# Plan\nNot an ADR"), 0o644)

	resVal, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_validate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("workspace validate failed: %v", err)
	}
	envVal := parseEnvelope(t, resVal)
	dataMap := envVal.Data.(map[string]any)
	decisionsMap := dataMap["decisions"].(map[string]any)

	validCount := int(decisionsMap["valid"].(float64))
	errorCount := int(decisionsMap["errors"].(float64))

	// Exactly 2 valid decisions (D-001.md, D-002.md), not 3 (double-counting DECISIONS.md) or 5 (non-ADR files)
	if validCount != 2 {
		t.Errorf("expected exactly 2 valid decisions, got %d (data: %v, warnings: %v)", validCount, decisionsMap, envVal.Warnings)
	}
	if errorCount != 0 {
		t.Errorf("expected 0 decision errors, got %d", errorCount)
	}
}

// TestRunGate_AutoPersistGateArtifact verifies research/PHASE-0-GATE.md auto-persistence on run_gate
func TestRunGate_AutoPersistGateArtifact(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Initialize decision scope workspace
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "decision"},
	})

	// 2. Run gate on fresh workspace (gate fails due to missing sessions/decisions)
	resFail, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("run_gate failed: %v", err)
	}
	envFail := parseEnvelope(t, resFail)
	failData := envFail.Data.(map[string]any)
	if failData["gate_artifact"] != core.GateFile {
		t.Errorf("expected gate_artifact to be %q, got: %v", core.GateFile, failData["gate_artifact"])
	}

	gatePath := filepath.Join(tmpDir, core.GateFile)
	failBytes, err := os.ReadFile(gatePath)
	if err != nil {
		t.Fatalf("expected %s to be created on gate failure: %v", core.GateFile, err)
	}
	failContent := string(failBytes)
	if !strings.Contains(failContent, `verdict: "FAIL"`) {
		t.Errorf("expected verdict: FAIL in frontmatter of failed gate, got:\n%s", failContent)
	}
	if !strings.Contains(failContent, "**Overall** | **FAIL**") {
		t.Errorf("expected **Overall** | **FAIL** in verdict table of failed gate")
	}

	// 3. Save completed decision session
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "S1-01",
			"content": `---
session_id: S1-01
title: Auth Strategy Comparison
date: 2026-10-01
status: complete
tags: [auth]
---
# Auth Strategy Comparison
## Recommendation
Adopt OAuth2 with PKCE. Grade A (RFC 7636)
## Key Findings
- OAuth2 with PKCE is standard for SPAs and native apps. Grade A (RFC 7636)
`,
		},
	})

	// 4. Record accepted decision with review trigger
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"decision_id":  "D-001",
			"content": `---
id: D-001
title: Auth Strategy
status: accepted
door_type: one-way
review_trigger: 6 months or major breaking RFC change
human_reviewed: true
---
# D-001: Auth Strategy
## Context
We need a secure authentication architecture for the platform.
## Decision
We select OAuth2 with PKCE. Grade A (RFC 7636)
## Consequences
Provides robust security with zero client secret exposure.
`,
		},
	})

	// 5. Run gate again — should PASS
	resPass, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate pass failed: %v", err)
	}
	envPass := parseEnvelope(t, resPass)
	passData := envPass.Data.(map[string]any)
	if passData["gate_status"] != "PASS" || passData["gate_passed"] != true {
		t.Fatalf("expected gate PASS, got status=%v passed=%v (warnings: %v)", passData["gate_status"], passData["gate_passed"], envPass.Warnings)
	}
	if passData["gate_artifact"] != core.GateFile {
		t.Errorf("expected gate_artifact to be %q, got: %v", core.GateFile, passData["gate_artifact"])
	}

	passBytes, err := os.ReadFile(gatePath)
	if err != nil {
		t.Fatalf("expected %s to exist on gate pass: %v", core.GateFile, err)
	}
	passContent := string(passBytes)
	if !strings.Contains(passContent, `verdict: "PASS"`) {
		t.Errorf("expected verdict: PASS in frontmatter of passed gate, got:\n%s", passContent)
	}
	if !strings.Contains(passContent, `track_b_result: "PASS"`) {
		t.Errorf("expected track_b_result: PASS in passed gate, got:\n%s", passContent)
	}
	if !strings.Contains(passContent, "**Overall** | **PASS**") {
		t.Errorf("expected **Overall** | **PASS** in verdict table of passed gate")
	}
	if !strings.Contains(passContent, "| D-001 | Auth Strategy | one-way | Track B | PASS |") {
		t.Errorf("expected D-001 in Decision Routing Summary table, got:\n%s", passContent)
	}
}

// TestSaveSession_FADRootCopy verifies Task 5: save SYN-01 writes research/FAD.md and mirrors to FOUNDING-ARCHITECTURE.md
func TestSaveSession_FADRootCopy(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	fadContent := `---
title: Founding Architecture Document
date: 2026-10-01
status: complete
---
# Founding Architecture Document
## Section 1: Executive Summary
Adopt Solution X. Grade A (official docs)
`

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "SYN-01",
			"content":      fadContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session SYN-01 failed: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("save_session SYN-01 unsuccessful: %s", env.Message)
	}
	dataMap, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got: %T", env.Data)
	}
	if dataMap["root_copy"] != "FOUNDING-ARCHITECTURE.md" {
		t.Errorf("expected root_copy to be 'FOUNDING-ARCHITECTURE.md', got: %v", dataMap["root_copy"])
	}

	// 1. research/FAD.md exists
	internalFADPath := filepath.Join(tmpDir, core.FADFile)
	internalBytes, err := os.ReadFile(internalFADPath)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", core.FADFile, err)
	}

	// 2. FOUNDING-ARCHITECTURE.md exists at project root
	rootFADPath := filepath.Join(tmpDir, "FOUNDING-ARCHITECTURE.md")
	rootBytes, err := os.ReadFile(rootFADPath)
	if err != nil {
		t.Fatalf("expected FOUNDING-ARCHITECTURE.md at root to exist: %v", err)
	}

	// 3. Both contents are identical
	if string(internalBytes) != string(rootBytes) {
		t.Errorf("internal FAD and root FAD content mismatch:\nInternal:\n%s\nRoot:\n%s", string(internalBytes), string(rootBytes))
	}

	// 4. sessions/SYN-01.md does NOT exist
	unwantedPath := filepath.Join(tmpDir, core.SessionsDir, "SYN-01.md")
	if _, err := os.Stat(unwantedPath); err == nil {
		t.Errorf("sessions/SYN-01.md should NOT exist (dual-write was removed)")
	}
}

// TestAmendSession_FADRootCopy verifies that amending SYN-01 mirrors updates to root FOUNDING-ARCHITECTURE.md
func TestAmendSession_FADRootCopy(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	fadContent := `---
title: Founding Architecture Document
date: 2026-10-01
status: complete
---
# Founding Architecture Document
## Section 1: Executive Summary
Adopt Solution X. Grade A (https://solutionx.org/docs | fetched)
`

	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "SYN-01",
			"content":      fadContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session SYN-01 failed: %v", err)
	}

	// Now amend SYN-01
	amendRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_amend_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "SYN-01",
			"amendment":    "Addendum: Disaster recovery failover testing completed successfully.",
		},
	})
	if err != nil {
		t.Fatalf("amend_session SYN-01 failed: %v", err)
	}
	env := parseEnvelope(t, amendRes)
	if !env.Success {
		t.Fatalf("amend_session SYN-01 unsuccessful: %s", env.Message)
	}
	dataMap, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got: %T", env.Data)
	}
	if dataMap["root_copy"] != "FOUNDING-ARCHITECTURE.md" {
		t.Errorf("expected root_copy to be 'FOUNDING-ARCHITECTURE.md', got: %v", dataMap["root_copy"])
	}

	// 1. Verify research/FAD.md contains amendment
	internalBytes, err := os.ReadFile(filepath.Join(tmpDir, core.FADFile))
	if err != nil {
		t.Fatalf("expected %s to exist: %v", core.FADFile, err)
	}
	if !strings.Contains(string(internalBytes), "Disaster recovery failover testing") {
		t.Errorf("internal FAD does not contain amendment text")
	}

	// 2. Verify root FOUNDING-ARCHITECTURE.md contains amendment
	rootBytes, err := os.ReadFile(filepath.Join(tmpDir, "FOUNDING-ARCHITECTURE.md"))
	if err != nil {
		t.Fatalf("expected root FOUNDING-ARCHITECTURE.md to exist: %v", err)
	}
	if !strings.Contains(string(rootBytes), "Disaster recovery failover testing") {
		t.Errorf("root FOUNDING-ARCHITECTURE.md does not contain amendment text")
	}

	// 3. Both are identical
	if string(internalBytes) != string(rootBytes) {
		t.Errorf("internal FAD and root FAD mismatch after amendment")
	}
}

// TestNextSession_SingleReadySession_NoParallelHint verifies that a single ready session has no parallelism_hint
func TestNextSession_SingleReadySession_NoParallelHint(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	singlePipeline := `# Pipeline
## Execution DAG
### Session T0-01: Only One Ready
| Field | Value |
|---|---|
| **ID** | T0-01 |
| **Dependencies** | None |
| **Output File** | sessions/T0-01.md |

` + "```prompt" + `
Investigate single session.
` + "```" + `

### Session T1-01: Blocked
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Dependencies** | [T0-01] |
| **Output File** | sessions/T1-01.md |

` + "```prompt" + `
Investigate blocked session.
` + "```" + `
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(singlePipeline), 0o644)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("next_session failed: %v", err)
	}
	env := parseEnvelope(t, res)
	dataMap, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got: %T", env.Data)
	}
	if dataMap["parallelism_hint"] != nil {
		t.Errorf("expected no parallelism_hint for single ready session, got: %v", dataMap["parallelism_hint"])
	}
	if strings.Contains(env.NextStep, "⚡") {
		t.Errorf("expected no ⚡ prefix when only 1 session is ready, got: %s", env.NextStep)
	}
}

// TestWorkspaceValidate_DAGContext verifies that workspace validation enforces one-way door rules from DAG
func TestWorkspaceValidate_DAGContext(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	pipeline := `# Pipeline
## Execution DAG
### Session T1-01: Consensus Protocol
| Field | Value |
|---|---|
| **ID** | T1-01 |
| **Door Type** | one-way |
| **Dependencies** | None |
| **Output File** | sessions/T1-01.md |

` + "```prompt" + `
Investigate consensus.
` + "```" + `
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(pipeline), 0o644)

	// Save session without substantive concerns directly to disk
	badSession := `---
session_id: T1-01
title: Consensus Protocol
date: 2026-09-29
status: complete
---
## Key Findings
- Raft protocol is suitable. Grade A (https://raft.github.io)

## Discovered Concerns
None.

## Sources & Evidence Ledger
| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | https://raft.github.io | Grade A | fresh | fetched | Consensus |
`
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T1-01.md"), []byte(badSession), 0o644)

	// Run workspace validate: should detect blocking error because DAG marks T1-01 as one-way
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_validate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("workspace validate failed: %v", err)
	}
	env := parseEnvelope(t, res)
	dataMap := env.Data.(map[string]any)
	sessMap := dataMap["sessions"].(map[string]any)
	if int(sessMap["errors"].(float64)) != 1 {
		t.Errorf("expected 1 session error for missing concerns on one-way door, got: %v", sessMap["errors"])
	}

	// Fix concerns
	fixedSession := strings.Replace(badSession, "None.",
		"Split-brain partitioning during WAN link flaps requires strict majority quorums across three availability zones.", 1)
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T1-01.md"), []byte(fixedSession), 0o644)

	res2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_validate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("workspace validate 2 failed: %v", err)
	}
	env2 := parseEnvelope(t, res2)
	dataMap2 := env2.Data.(map[string]any)
	sessMap2 := dataMap2["sessions"].(map[string]any)
	if int(sessMap2["errors"].(float64)) != 0 {
		t.Errorf("expected 0 session errors after fixing concerns, got: %v", sessMap2["errors"])
	}
}

// TestWorkspaceValidate_FallbackToDecisionsRegistry verifies fallback parsing of DECISIONS.md
func TestWorkspaceValidate_FallbackToDecisionsRegistry(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	decisionsMD := `# Architectural Decisions

<!-- DECISION: D-001 -->
---
id: D-001
title: Consensus Engine
status: accepted
door_type: two-way
---
# D-001: Consensus Engine
We choose Raft consensus for deterministic state replication. Grade A (https://raft.github.io)
<!-- /DECISION: D-001 -->
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte(decisionsMD), 0o644)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_validate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	env := parseEnvelope(t, res)
	dataMap := env.Data.(map[string]any)
	decMap := dataMap["decisions"].(map[string]any)
	if int(decMap["valid"].(float64)) != 1 {
		t.Errorf("expected 1 valid decision from fallback DECISIONS.md parsing, got: %v", decMap["valid"])
	}
}

// TestRunGate_ZeroTwoWayDoors_TrackAPasses verifies Track A handles 0 two-way doors cleanly
func TestRunGate_ZeroTwoWayDoors_TrackAPasses(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "decision"},
	})

	pipe := `# Pipeline
#### D-AUTH-S01: Auth Strategy
| **ID** | D-AUTH-S01 |
| **Dependencies** | None |
| **Output File** | sessions/D-AUTH-S01.md |
` + "```prompt" + `
Auth prompt
` + "```" + `
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(pipe), 0o644)

	sessionContent := `---
session_id: D-AUTH-S01
title: Auth Strategy
date: 2026-09-29
door_type: one-way
status: complete
---
# Findings
WebAuthn is standard. Grade A (https://w3.org/TR/webauthn-2)

## Discovered Concerns
Device registration complexity requires robust fallback credentials.

## Sources
| # | Source | Grade | Verification |
|---|---|---|---|
| 1 | https://w3.org/TR/webauthn-2 | Grade A | fetched |
`
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "D-AUTH-S01.md"), []byte(sessionContent), 0o644)

	// Record one-way ADR
	adrContent := `<!-- DECISION: D-001 -->
---
decision_id: D-001
title: WebAuthn Adoption
status: accepted
door_type: one-way
review_trigger: Passkey adoption drops below 80%
informing_sessions: [D-AUTH-S01]
---
# D-001: WebAuthn Adoption

## Context & Rationale
We adopt WebAuthn as the primary authentication strategy for enterprise identity assurance.
This is a critical architectural commitment verified across standards documents and browser implementations. Grade A (https://w3.org/TR/webauthn-2)

## Rejected Alternatives
1. Password-only authentication: high credential stuffing vulnerability.
2. SMS 2FA: SIM swapping vulnerabilities documented by NIST SP 800-63B.
<!-- /DECISION: D-001 -->
`
	_ = os.WriteFile(filepath.Join(tmpDir, filepath.Join(core.ResearchDir, "D-001.md")), []byte(adrContent), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte(adrContent), 0o644)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate failed: %v", err)
	}
	env := parseEnvelope(t, res)
	dataMap := env.Data.(map[string]any)
	if dataMap["gate_status"] != "PASS" {
		t.Fatalf("expected gate PASS, got: %v (warnings: %v)", dataMap["gate_status"], env.Warnings)
	}

	gateArtifact, err := os.ReadFile(filepath.Join(tmpDir, filepath.Join(core.ResearchDir, "PHASE-0-GATE.md")))
	if err != nil {
		t.Fatalf("reading PHASE-0-GATE.md: %v", err)
	}
	gateStr := string(gateArtifact)
	if !strings.Contains(gateStr, "Track A Result:** PASS (0 two-way doors)") {
		t.Errorf("expected 'Track A Result: PASS (0 two-way doors)', got in gate artifact:\n%s", gateStr)
	}
}

// TestRunGate_ADRStub_AdvisoryWarning verifies B10 stub check triggers advisory on small ADR body
func TestRunGate_ADRStub_AdvisoryWarning(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "decision"},
	})

	pipe := `# Pipeline
#### D-01: Auth
| **ID** | D-01 |
| **Dependencies** | None |
| **Output File** | sessions/D-01.md |
` + "```prompt" + `
Auth
` + "```" + `
`
	_ = os.WriteFile(filepath.Join(tmpDir, core.PipelineFile), []byte(pipe), 0o644)
	_ = os.MkdirAll(filepath.Join(tmpDir, core.SessionsDir), 0o755)
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "D-01.md"), []byte(`---
session_id: D-01
title: Auth
date: 2026-09-29
door_type: one-way
status: complete
---
Findings. Grade A (https://example.com)
## Discovered Concerns
Significant integration complexity discovered with legacy IdPs.
## Sources
| # | Source | Grade | Verification |
|---|---|---|---|
| 1 | https://example.com | Grade A | fetched |
`), 0o644)

	// Short ADR body (< 200 characters)
	shortADR := `---
decision_id: D-001
title: Short Decision
status: accepted
door_type: one-way
review_trigger: Trigger text
---
# D-001
Use tool X. Grade A (https://example.com)
## Rejected Alternatives
None.
`
	_ = os.WriteFile(filepath.Join(tmpDir, filepath.Join(core.ResearchDir, "D-001.md")), []byte(shortADR), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, core.DecisionsFile), []byte(shortADR), 0o644)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate failed: %v", err)
	}
	env := parseEnvelope(t, res)
	foundStub := false
	for _, w := range env.Warnings {
		if strings.Contains(w, "B-DECISION-STUB") {
			foundStub = true
			break
		}
	}
	if !foundStub {
		t.Errorf("expected B-DECISION-STUB advisory for < 200 char body, got warnings: %v", env.Warnings)
	}
}





