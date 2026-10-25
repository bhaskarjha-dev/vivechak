package mcputil

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ============================================================================
// T1-01: Scope-Aware Template Copying
// ============================================================================

// TestInit_DecisionScopeCopiesOnly2Templates verifies vivechak_init with
// decision scope copies only DECISIONS + CONFLICT-RESOLUTION templates.
func TestInit_DecisionScopeCopiesOnly2Templates(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "decision"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	tmplDir := filepath.Join(tmpDir, "research", "templates")

	// Should exist
	expected := core.TemplatesForScope(core.ScopeDecision)
	for _, tmpl := range expected {
		path := filepath.Join(tmplDir, tmpl)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected template %q to exist for decision scope", tmpl)
		}
	}

	// Should NOT exist
	excluded := []string{
		"COMPARISON-SESSION.template.md",
		"FOUNDING-ARCHITECTURE.template.md",
		"PHASE-0-GATE.template.md",
	}
	for _, tmpl := range excluded {
		path := filepath.Join(tmplDir, tmpl)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("template %q should NOT exist for decision scope", tmpl)
		}
	}
}

// TestInit_ComparisonScopeCopiesOnly1Template verifies vivechak_init with
// comparison scope copies only COMPARISON-SESSION template.
func TestInit_ComparisonScopeCopiesOnly1Template(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "comparison"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	tmplDir := filepath.Join(tmpDir, "research", "templates")

	// Should exist
	path := filepath.Join(tmplDir, "COMPARISON-SESSION.template.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("COMPARISON-SESSION.template.md should exist for comparison scope")
	}

	// Should NOT exist
	excluded := []string{
		"DECISIONS.template.md",
		"CONFLICT-RESOLUTION.template.md",
		"FOUNDING-ARCHITECTURE.template.md",
		"PHASE-0-GATE.template.md",
	}
	for _, tmpl := range excluded {
		path := filepath.Join(tmplDir, tmpl)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("template %q should NOT exist for comparison scope", tmpl)
		}
	}
}

// TestInit_ProjectScopeCopiesAll6Templates verifies vivechak_init with
// project scope copies all 6 templates.
func TestInit_ProjectScopeCopiesAll6Templates(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	tmplDir := filepath.Join(tmpDir, "research", "templates")
	for _, tmpl := range core.TemplatesToCopy {
		path := filepath.Join(tmplDir, tmpl)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("template %q should exist for project scope", tmpl)
		}
	}
}

// TestInit_RepairRestoresOnlyScopeTemplates verifies the re-init repair loop
// only restores templates appropriate for the workspace scope.
func TestInit_RepairRestoresOnlyScopeTemplates(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init with decision scope
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "decision"},
	})

	// Delete one template
	os.Remove(filepath.Join(tmpDir, "research", "templates", "DECISIONS.template.md"))

	// Re-init should repair only decision-scope templates
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "decision"},
	})
	if err != nil {
		t.Fatalf("re-init: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Success {
		t.Fatalf("re-init failed: %s", env.Message)
	}

	// Verify restored
	path := filepath.Join(tmpDir, "research", "templates", "DECISIONS.template.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("DECISIONS.template.md should have been restored")
	}

	// Verify project-only templates were NOT created during repair
	excluded := []string{"FOUNDING-ARCHITECTURE.template.md", "PHASE-0-GATE.template.md"}
	for _, tmpl := range excluded {
		path := filepath.Join(tmpDir, "research", "templates", tmpl)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("template %q should NOT be created during decision-scope repair", tmpl)
		}
	}
}

// ============================================================================
// T1-02 / T2-06: next_step Quality Bar Language + Blocked Sessions
// ============================================================================

// TestNextSession_QualityBarLanguage verifies next_step uses quality bar language
// and does NOT contain "web search enabled".
func TestNextSession_QualityBarLanguage(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init + save pipeline with 2 sessions
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	pipeline := `# Pipeline
#### T1-01: Database Selection
| **ID** | T1-01 |
| **Dependencies** | None |
` + "```prompt\nResearch databases.\n```" + `

#### SYN-01: Synthesis
| **ID** | SYN-01 |
| **Dependencies** | T1-01 |
` + "```prompt\nSynthesize findings.\n[ALL_SESSION_FINDINGS]\n```"

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save_plan",
		Arguments: map[string]any{"project_root": tmpDir, "content": pipeline},
	})

	// Get next session (T1-01 - research session)
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}
	env := parseEnvelope(t, result)

	// P4: No "web search enabled"
	if strings.Contains(env.NextStep, "web search enabled") {
		t.Error("next_step should NOT contain 'web search enabled' (P4 violation)")
	}

	// Quality bar language for research session
	if !strings.Contains(env.NextStep, "evidence") {
		t.Error("next_step for research session should mention evidence quality")
	}
}

// TestNextSession_SynthesisLanguage verifies synthesis sessions get synthesis-specific guidance.
func TestNextSession_SynthesisLanguage(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	pipeline := `# Pipeline
#### T1-01: Database Selection
| **ID** | T1-01 |
| **Dependencies** | None |
` + "```prompt\nResearch.\n```" + `

#### SYN-01: Synthesis
| **ID** | SYN-01 |
| **Dependencies** | T1-01 |
` + "```prompt\nSynthesize.\n[ALL_SESSION_FINDINGS]\n```"

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save_plan",
		Arguments: map[string]any{"project_root": tmpDir, "content": pipeline},
	})

	// Complete T1-01
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-01",
			"content": `---
session_id: T1-01
title: DB Selection
date: 2026-09-29
status: complete
---
# DB Selection
PostgreSQL recommended. Grade A (docs)
`,
		},
	})

	// Get SYN-01 next_step
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}
	env := parseEnvelope(t, result)

	// Synthesis-specific language
	if !strings.Contains(env.NextStep, "ynthesize") {
		t.Error("SYN-01 next_step should contain synthesis-specific language")
	}
	if !strings.Contains(env.NextStep, "session evidence") || !strings.Contains(env.NextStep, "Conflicts") {
		t.Error("SYN-01 next_step should mention tracing to session evidence and surfacing conflicts")
	}
}

// TestNextSession_BlockedSessionsData verifies blocked_sessions appears in response data.
func TestNextSession_BlockedSessionsData(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	// Pipeline: T1-01 -> T2-01, T1-01 -> SYN-01
	pipeline := `# Pipeline
#### T1-01: Research
| **ID** | T1-01 |
| **Dependencies** | None |
` + "```prompt\nResearch.\n```" + `

#### T2-01: Deep Dive
| **ID** | T2-01 |
| **Dependencies** | T1-01 |
` + "```prompt\nDeep dive.\n[UPSTREAM_FINDINGS]\n```" + `

#### SYN-01: Synthesis
| **ID** | SYN-01 |
| **Dependencies** | T1-01, T2-01 |
` + "```prompt\nSynthesize.\n[ALL_SESSION_FINDINGS]\n```"

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save_plan",
		Arguments: map[string]any{"project_root": tmpDir, "content": pipeline},
	})

	// Get T1-01 — T2-01 and SYN-01 should be blocked
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}
	env := parseEnvelope(t, result)
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatal("expected data to be a map")
	}

	blocked, exists := data["blocked_sessions"]
	if !exists {
		t.Fatal("expected blocked_sessions in response data")
	}

	blockedList, ok := blocked.([]any)
	if !ok {
		t.Fatalf("blocked_sessions should be an array, got %T", blocked)
	}

	if len(blockedList) < 1 {
		t.Error("expected at least 1 blocked session")
	}

	// Verify structure: each has session_id and blocked_by
	for _, item := range blockedList {
		m, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("blocked session item should be a map, got %T", item)
		}
		if _, has := m["session_id"]; !has {
			t.Error("blocked session should have session_id")
		}
		if _, has := m["blocked_by"]; !has {
			t.Error("blocked session should have blocked_by")
		}
	}
}

// TestNextSession_ParallelSessionCount verifies the parallel session count phrasing.
func TestNextSession_ParallelSessionCount(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})

	// 3 independent sessions
	pipeline := `# Pipeline
#### T1-01: Research A
| **ID** | T1-01 |
| **Dependencies** | None |
` + "```prompt\nA.\n```" + `

#### T1-02: Research B
| **ID** | T1-02 |
| **Dependencies** | None |
` + "```prompt\nB.\n```" + `

#### T1-03: Research C
| **ID** | T1-03 |
| **Dependencies** | None |
` + "```prompt\nC.\n```"

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save_plan",
		Arguments: map[string]any{"project_root": tmpDir, "content": pipeline},
	})

	result, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	env := parseEnvelope(t, result)
	// Should mention number of parallel sessions with ⚡ prefix
	if !strings.HasPrefix(env.NextStep, "⚡ 3 sessions ready in parallel") {
		t.Errorf("expected '⚡ 3 sessions ready in parallel' prefix in next_step, got: %s", env.NextStep)
	}

	dataMap, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got: %T", env.Data)
	}
	if hint, ok := dataMap["parallelism_hint"].(string); !ok || !strings.Contains(hint, "PARALLEL EXECUTION AVAILABLE: 3 independent sessions") {
		t.Errorf("expected parallelism_hint in data, got: %v", dataMap["parallelism_hint"])
	}
}

// ============================================================================
// T2-03: Advisory Validation Language + Observations
// ============================================================================

// TestValidate_AdvisoryLanguage verifies validation uses advisory language.
func TestValidate_AdvisoryLanguage(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	// Content with blocking issues
	content := `Short body without frontmatter.`

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_validate",
		Arguments: map[string]any{
			"content":       content,
			"artifact_type": "session",
		},
	})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	env := parseEnvelope(t, result)

	// Should say "Review" not "Fix"
	if strings.Contains(env.NextStep, "Fix the blocking") {
		t.Error("next_step should NOT say 'Fix the blocking' — use advisory language")
	}
	if !strings.Contains(env.NextStep, "Review") {
		t.Error("next_step should use advisory 'Review' language")
	}
}

// TestValidate_ObservationsInResponse verifies observations appear in response data.
func TestValidate_ObservationsInResponse(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	// Valid session content without Prior/Delta sections
	content := `---
session_id: T1-01
title: Test Session
date: 2026-09-29
---
# Test Session

## Recommendation
Use PostgreSQL. Grade A (official docs)

## Key Findings
- Handles 100k writes/sec
`

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_validate",
		Arguments: map[string]any{
			"content":       content,
			"artifact_type": "session",
		},
	})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	env := parseEnvelope(t, result)

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatal("expected data to be a map")
	}

	observations, exists := data["observations"]
	if !exists {
		t.Fatal("expected observations in validation response data")
	}

	obsList, ok := observations.([]any)
	if !ok {
		t.Fatalf("observations should be an array, got %T", observations)
	}

	// Should have suggestions for missing Prior and Delta
	foundPrior := false
	foundDelta := false
	for _, obs := range obsList {
		m, ok := obs.(map[string]any)
		if !ok {
			continue
		}
		msg, _ := m["message"].(string)
		if strings.Contains(msg, "Prior") {
			foundPrior = true
		}
		if strings.Contains(msg, "Delta") {
			foundDelta = true
		}
	}
	if !foundPrior {
		t.Error("expected observation about missing Prior section")
	}
	if !foundDelta {
		t.Error("expected observation about missing Delta section")
	}
}

// TestValidate_NoObservationsWhenComplete verifies no structure observations when Prior+Delta present.
func TestValidate_NoObservationsWhenComplete(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	content := `---
session_id: T1-01
title: Complete Session
date: 2026-09-29
---
## Prior
- I believe PostgreSQL is best because of JSON support

## Research Question
Should we use PostgreSQL or MySQL?

## Key Findings
- PostgreSQL handles JSON natively. Grade A (https://www.postgresql.org/docs/current/datatype-json.html)
- MySQL lacks recursive CTEs. Grade B (benchmarks)
- PostgreSQL has better indexing. Grade A (RFC 7636)
- Connection pooling needed. Grade C (vendor docs)

## Delta
| Prior Belief | Status | Evidence | Impact |
|---|---|---|---|
| PostgreSQL best for JSON | Confirmed | Grade A (https://www.postgresql.org/docs/) | high |
`

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_validate",
		Arguments: map[string]any{
			"content":       content,
			"artifact_type": "session",
		},
	})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	env := parseEnvelope(t, result)
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatal("expected data to be a map")
	}

	// With Prior, Delta, and 4+ evidence grades — should have no structure/evidence observations
	if obs, exists := data["observations"]; exists {
		obsList, _ := obs.([]any)
		for _, o := range obsList {
			m, _ := o.(map[string]any)
			cat, _ := m["category"].(string)
			if cat == "structure" || cat == "evidence" {
				t.Errorf("unexpected %s observation when content is complete: %v", cat, m)
			}
		}
	}
}

// TestValidate_WarningCountLanguage verifies advisory phrasing for warnings.
func TestValidate_WarningCountLanguage(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	// Valid content with a warning (no evidence grades)
	content := `---
session_id: T1-01
title: Session
date: 2026-09-29
---
# Test Session
This session has no evidence grades but is otherwise valid with enough body content.
` + strings.Repeat("Additional substantive content. ", 20)

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_validate",
		Arguments: map[string]any{
			"content":       content,
			"artifact_type": "session",
		},
	})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	env := parseEnvelope(t, result)

	if strings.Contains(env.NextStep, "You can save it using the appropriate save tool") {
		t.Error("old language found — should use 'observations' framing")
	}
}

// ============================================================================
// P4: No "web search enabled" in any tool next_step
// ============================================================================

// TestNoWebSearchEnabledInAnyNextStep runs every MCP tool and verifies none say "web search enabled".
func TestNoWebSearchEnabledInAnyNextStep(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init
	res, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "project"},
	})
	env := parseEnvelope(t, res)
	if strings.Contains(env.NextStep, "web search enabled") {
		t.Error("vivechak_init next_step contains 'web search enabled'")
	}

	// Prepare generator
	res, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_prepare_generator",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"context":      "Test project",
		},
	})
	env = parseEnvelope(t, res)
	if strings.Contains(env.NextStep, "web search enabled") {
		t.Error("vivechak_prepare_generator next_step contains 'web search enabled'")
	}

	// Save plan
	pipeline := `# Pipeline
#### T1-01: Research
| **ID** | T1-01 |
| **Dependencies** | None |
` + "```prompt\nResearch.\n```"

	res, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save_plan",
		Arguments: map[string]any{"project_root": tmpDir, "content": pipeline},
	})
	env = parseEnvelope(t, res)
	if strings.Contains(env.NextStep, "web search enabled") {
		t.Error("vivechak_save_plan next_step contains 'web search enabled'")
	}

	// Next session
	res, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	env = parseEnvelope(t, res)
	if strings.Contains(env.NextStep, "web search enabled") {
		t.Error("vivechak_next_session next_step contains 'web search enabled'")
	}

	// Validate
	res, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_validate",
		Arguments: map[string]any{"content": "test", "artifact_type": "session"},
	})
	env = parseEnvelope(t, res)
	if strings.Contains(env.NextStep, "web search enabled") {
		t.Error("vivechak_validate next_step contains 'web search enabled'")
	}

	// Status
	res, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_status",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	env = parseEnvelope(t, res)
	if strings.Contains(env.NextStep, "web search enabled") {
		t.Error("vivechak_status next_step contains 'web search enabled'")
	}
}

// ============================================================================
// T2-01: Generator Content Verification (8-block structure)
// ============================================================================

// TestGenerator_8BlockStructure verifies generators contain the 8-block prompt structure.
func TestGenerator_8BlockStructure(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	scopes := []string{"project", "decision", "comparison"}
	requiredBlocks := []string{"DECISION", "BRIEF", "SCOPE", "KNOWN", "CALIBRATION", "APPROACH", "DONE", "FORMAT"}

	for _, scope := range scopes {
		t.Run(scope, func(t *testing.T) {
			tmpDir := t.TempDir()
			result, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_prepare_generator",
				Arguments: map[string]any{
					"project_root": tmpDir,
					"scope":        scope,
					"context":      "A test application for research",
				},
			})
			if err != nil {
				t.Fatalf("prepare_generator(%s): %v", scope, err)
			}

			env := parseEnvelope(t, result)
			data, _ := env.Data.(map[string]any)
			prompt, _ := data["prompt"].(string)

			for _, block := range requiredBlocks {
				// For project scope, these blocks are in the instructions, not the prompt itself
				if scope == "project" {
					// Project generator describes how to BUILD prompts with 8 blocks
					if block == "DONE" {
						// The generator's instructions should reference the DONE block
						// (not the old DELIVERABLE block) for session prompts
						if !strings.Contains(prompt, "**DONE:**") {
							t.Error("project generator should reference **DONE:** block for session prompts")
						}
					}
				} else {
					// Decision and Comparison generators include the actual prompt template
					header := "## " + block
					if !strings.Contains(prompt, header) {
						t.Errorf("%s generator missing block %q", scope, header)
					}
				}
			}

			// Verify the session prompt template uses DONE not DELIVERABLE
			// The generator itself may have a ## DELIVERABLE section for its output spec,
			// but the prompt-writing instructions should reference **DONE:** not **DELIVERABLE:**
			if strings.Contains(prompt, "**DELIVERABLE:**") {
				t.Errorf("%s generator session prompt instructions should use **DONE:** not **DELIVERABLE:**", scope)
			}
		})
	}
}

// TestGenerator_NoCoverageChecklist verifies generators don't use "coverage checklist" phrasing.
func TestGenerator_NoCoverageChecklist(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	scopes := []string{"project", "decision", "comparison"}
	for _, scope := range scopes {
		t.Run(scope, func(t *testing.T) {
			tmpDir := t.TempDir()
			result, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_prepare_generator",
				Arguments: map[string]any{
					"project_root": tmpDir,
					"scope":        scope,
					"context":      "Test",
				},
			})
			if err != nil {
				t.Fatalf("prepare_generator(%s): %v", scope, err)
			}
			env := parseEnvelope(t, result)
			data, _ := env.Data.(map[string]any)
			prompt, _ := data["prompt"].(string)

			if strings.Contains(strings.ToLower(prompt), "coverage checklist") {
				t.Errorf("%s generator should not contain 'coverage checklist' (T1-05)", scope)
			}
		})
	}
}

// ============================================================================
// T2-02: Prior/Delta in Generator FORMAT Blocks
// ============================================================================

// TestGenerator_PriorDeltaInFormat verifies FORMAT sections include Prior and Delta.
func TestGenerator_PriorDeltaInFormat(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	scopes := []string{"project", "decision", "comparison"}
	for _, scope := range scopes {
		t.Run(scope, func(t *testing.T) {
			tmpDir := t.TempDir()
			result, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_prepare_generator",
				Arguments: map[string]any{
					"project_root": tmpDir,
					"scope":        scope,
					"context":      "Test",
				},
			})
			if err != nil {
				t.Fatalf("prepare_generator(%s): %v", scope, err)
			}
			env := parseEnvelope(t, result)
			data, _ := env.Data.(map[string]any)
			prompt, _ := data["prompt"].(string)

			if !strings.Contains(prompt, "Prior") {
				t.Errorf("%s generator should mention Prior section", scope)
			}
			if !strings.Contains(prompt, "Delta") {
				t.Errorf("%s generator should mention Delta section", scope)
			}
		})
	}
}

// ============================================================================
// T3-01: Iterative Research Guidance in Generators
// ============================================================================

// TestGenerator_IterativeGuidance verifies iterative research language in generators.
func TestGenerator_IterativeGuidance(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	scopes := []string{"project", "decision", "comparison"}
	for _, scope := range scopes {
		t.Run(scope, func(t *testing.T) {
			tmpDir := t.TempDir()
			result, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_prepare_generator",
				Arguments: map[string]any{
					"project_root": tmpDir,
					"scope":        scope,
					"context":      "Test",
				},
			})
			if err != nil {
				t.Fatalf("prepare_generator(%s): %v", scope, err)
			}
			env := parseEnvelope(t, result)
			data, _ := env.Data.(map[string]any)
			prompt, _ := data["prompt"].(string)

			// Should have research moves vocabulary
			moves := []string{"DEEPEN", "WIDEN", "CORROBORATE", "FALSIFY", "PIVOT"}
			for _, move := range moves {
				if !strings.Contains(prompt, move) {
					t.Errorf("%s generator missing research move %q", scope, move)
				}
			}

			// Should have iterative guidance
			if !strings.Contains(prompt, "iterative") {
				t.Errorf("%s generator missing 'iterative' research guidance", scope)
			}
		})
	}
}

// ============================================================================
// T2-01: CALIBRATION block in generators
// ============================================================================

// TestGenerator_CalibrationBlock verifies CALIBRATION block in decision/comparison generators.
func TestGenerator_CalibrationBlock(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	scopes := []string{"decision", "comparison"}
	for _, scope := range scopes {
		t.Run(scope, func(t *testing.T) {
			tmpDir := t.TempDir()
			result, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_prepare_generator",
				Arguments: map[string]any{
					"project_root": tmpDir,
					"scope":        scope,
					"context":      "Test",
				},
			})
			if err != nil {
				t.Fatalf("prepare_generator(%s): %v", scope, err)
			}
			env := parseEnvelope(t, result)
			data, _ := env.Data.(map[string]any)
			prompt, _ := data["prompt"].(string)

			if !strings.Contains(prompt, "## CALIBRATION") {
				t.Errorf("%s generator missing CALIBRATION block", scope)
			}
			if !strings.Contains(prompt, "hypothesis to test") {
				t.Errorf("%s generator CALIBRATION should include 'hypothesis to test'", scope)
			}
		})
	}
}

// ============================================================================
// Gate template count — scope aware
// ============================================================================

// TestRunGate_ScopeAwareTemplateCount verifies gate uses scope-specific template counts.
func TestRunGate_ScopeAwareTemplateCount(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	// Decision scope: 2 templates should be enough
	tmpDir := t.TempDir()
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "decision"},
	})

	// Save a session
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "S1-01",
			"content": `---
session_id: S1-01
title: Decision Research
date: 2026-09-29
status: complete
---
# Research
Found X. Grade A (docs)
`,
		},
	})

	// Record decision
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "decision",
			"decision_id":   "D-001",
			"content": `---
decision_id: D-001
title: Test Decision
status: accepted
door_type: two-way
---
# Context
Need X.
# Decision
Use Y. Grade A (docs)
# Consequences
None significant.
`,
		},
	})

	// Gate should pass with only 2 templates
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_run_gate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env := parseEnvelope(t, res)
	data, _ := env.Data.(map[string]any)

	// Should not complain about template count
	msg := env.Message
	if strings.Contains(msg, "templates found") {
		t.Errorf("gate should not complain about template count for decision scope: %s", msg)
	}

	// Verify gate passes
	if data["gate_status"] != "PASS" {
		// Dump the full response for debugging
		raw, _ := json.MarshalIndent(data, "", "  ")
		t.Errorf("gate should PASS with scope-appropriate templates, got: %s\nFull data: %s", data["gate_status"], string(raw))
	}
}

// ============================================================================
// Comparison scope next_step quality guidance
// ============================================================================

// TestNextSession_ComparisonQualityGuidance verifies comparison scope next_step
// includes evidence quality language.
func TestNextSession_ComparisonQualityGuidance(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_init",
		Arguments: map[string]any{"project_root": tmpDir, "scope": "comparison"},
	})

	compPrompt := `# RESEARCH BRIEF: Queue Selection
Compare Kafka vs RabbitMQ. A (benchmark)
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "comparison",
			"decision_id":  "D-001",
			"slug":         "queue",
			"content":      compPrompt,
		},
	})

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}
	env := parseEnvelope(t, res)

	if !strings.Contains(env.NextStep, "evidence") {
		t.Errorf("comparison next_step should mention evidence quality, got: %s", env.NextStep)
	}
	if strings.Contains(env.NextStep, "web search enabled") {
		t.Error("comparison next_step should NOT say 'web search enabled'")
	}
}

// TestPrepareGenerator_AutoInfersScope verifies that prepare_generator automatically infers
// comparison scope from "vs" in the context when scope is omitted.
func TestPrepareGenerator_AutoInfersScope(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	tmpDir := t.TempDir()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_prepare_generator",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"context":      "Compare Winston vs Pino for high throughput logging in Node.js",
		},
	})
	if err != nil {
		t.Fatalf("prepare_generator: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("prepare_generator failed: %s", env.Message)
	}
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatal("expected data map in envelope")
	}
	if data["scope"] != "comparison" {
		t.Errorf("expected scope 'comparison', got: %v", data["scope"])
	}
	prompt, _ := data["prompt"].(string)
	if !strings.Contains(prompt, "COMPARISON") {
		t.Errorf("expected comparison prompt content")
	}
}

// TestSavePlan_AutoInitializesWorkspace verifies that save_plan auto-initializes the
// workspace structure and templates when called on a fresh directory.
func TestSavePlan_AutoInitializesWorkspace(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	compContent := `# RESEARCH BRIEF: Logger Shootout
Compare Winston vs Pino. Evidence Grade A (official docs).
`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "comparison",
			"decision_id":  "C-01",
			"slug":         "logging",
			"content":      compContent,
		},
	})
	if err != nil {
		t.Fatalf("save_plan failed: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Success {
		t.Fatalf("save_plan returned unsuccessful: %s", env.Message)
	}

	// Verify workspace structure was auto-initialized
	if _, err := os.Stat(filepath.Join(tmpDir, "research", "templates", "COMPARISON-SESSION.template.md")); err != nil {
		t.Errorf("auto-init failed to copy comparison template: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "research", "C-01-cmp-logging.md")); err != nil {
		t.Errorf("comparison plan file not found: %v", err)
	}
}

// TestPrepareGenerator_ScopeAwareNextStep verifies that prepare_generator returns
// scope-appropriate next_step guidance for comparison, decision, and project scopes.
func TestPrepareGenerator_ScopeAwareNextStep(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	cases := []struct {
		name        string
		scope       string
		context     string
		expectedSub string
	}{
		{
			name:        "comparison scope",
			scope:       "comparison",
			context:     "Pino vs Winston for logging",
			expectedSub: "WEP comparison matrix",
		},
		{
			name:        "decision scope",
			scope:       "decision",
			context:     "Migrate PostgreSQL to ClickHouse for analytics",
			expectedSub: "decision plan and ADR skeleton",
		},
		{
			name:        "project scope",
			scope:       "project",
			context:     "Build a distributed streaming architecture in Go",
			expectedSub: "generate your research pipeline",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_prepare_generator",
				Arguments: map[string]any{
					"project_root": t.TempDir(),
					"scope":        tc.scope,
					"context":      tc.context,
				},
			})
			if err != nil {
				t.Fatalf("prepare_generator(%s): %v", tc.scope, err)
			}
			env := parseEnvelope(t, res)
			if !strings.Contains(env.NextStep, tc.expectedSub) {
				t.Errorf("expected next_step to contain %q, got: %q", tc.expectedSub, env.NextStep)
			}
		})
	}
}

// TestPrepareGenerator_ExtractComparisonOptions verifies that Comparison scope auto-populates candidate options from context.
func TestPrepareGenerator_ExtractComparisonOptions(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	cases := []struct {
		context  string
		expected string
	}{
		{
			context:  "Compare Pino vs Winston for high-throughput microservices",
			expected: "OPTIONS: Pino, Winston",
		},
		{
			context:  "compare Gemni 4 Argon vs GPT 6 Astra vs Opus 5.5 for CI pipeline",
			expected: "OPTIONS: Gemni 4 Argon, GPT 6 Astra, Opus 5.5",
		},
		{
			context:  "Auth0 vs Clerk versus Supabase for mobile authentication",
			expected: "OPTIONS: Auth0, Clerk, Supabase",
		},
	}

	for _, tc := range cases {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_prepare_generator",
			Arguments: map[string]any{
				"project_root": t.TempDir(),
				"scope":        "comparison",
				"context":      tc.context,
			},
		})
		if err != nil {
			t.Fatalf("prepare_generator(%q): %v", tc.context, err)
		}

		env := parseEnvelope(t, res)
		data, ok := env.Data.(map[string]any)
		if !ok {
			t.Fatalf("expected env.Data map, got %T", env.Data)
		}
		prompt, _ := data["prompt"].(string)
		if !strings.Contains(prompt, tc.expected) {
			t.Errorf("for context %q, expected prompt to contain %q, got:\n%s", tc.context, tc.expected, prompt)
		}
	}
}

// TestServerInstructions_MultiScopeAndOutofScope verifies the instruction prompt includes
// multi-scope workflows, out-of-scope boundaries, and fits within size limits.
func TestServerInstructions_MultiScopeAndOutofScope(t *testing.T) {
	instrLen := len(ServerInstructions)
	if instrLen > 600 {
		t.Errorf("expected ServerInstructions length <= 600 bytes per P4.4, got %d", instrLen)
	}

	mandatoryPhrases := []string{
		"research/",
		"A-E",
		"vivechak_start",
	}

	for _, phrase := range mandatoryPhrases {
		if !strings.Contains(ServerInstructions, phrase) {
			t.Errorf("ServerInstructions missing required phrase: %q", phrase)
		}
	}
}

// TestSaveSession_CustomOutputFileFromDAG verifies that vivechak_save_session respects
// the custom OutputFile path configured in the DAG pipeline.
func TestSaveSession_CustomOutputFileFromDAG(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Initialize project scope
	initRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("vivechak_init failed: %v", err)
	}
	_ = parseEnvelope(t, initRes)

	// 2. Save pipeline with custom OutputFile
	pipelineContent := `# RESEARCH PIPELINE

## Overview
| ID | Topic | Dependencies | OutputFile |
|---|---|---|---|
| T1-01 | Database Architecture | none | sessions/T1-01-database-selection.md |
| T1-02 | Auth System | none | |

### T1-01: Database Architecture
` + "```prompt\nResearch DB\n```" + `

### T1-02: Auth System
` + "```prompt\nResearch Auth\n```" + `
`
	savePlanRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"content":      pipelineContent,
		},
	})
	if err != nil {
		t.Fatalf("vivechak_save_plan failed: %v", err)
	}
	_ = parseEnvelope(t, savePlanRes)

	// 3. Save session T1-01 with custom OutputFile
	sessionContent := `---
id: T1-01
title: Database Architecture
date: 2026-10-23
status: complete
---
# Database Architecture
## Key Findings
- Finding 1: PostgreSQL handles high concurrency. Grade A (official docs)
`
	saveSessRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-01",
			"content":      sessionContent,
		},
	})
	if err != nil {
		t.Fatalf("vivechak_save_session failed: %v", err)
	}
	env := parseEnvelope(t, saveSessRes)
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map in save_session response, got %T", env.Data)
	}
	filePath, _ := data["file_path"].(string)
	expectedPath := "research/sessions/T1-01-database-selection.md"
	if filePath != expectedPath {
		t.Errorf("file_path = %q; want %q", filePath, expectedPath)
	}

	// Verify file actually exists at custom output path
	fullPath := filepath.Join(tmpDir, filePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("file was not created at expected custom output path: %s", fullPath)
	}

	// 4. Save session T1-02 without custom OutputFile (should use default T1-02.md)
	saveSess2Res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T1-02",
			"content":      sessionContent,
		},
	})
	if err != nil {
		t.Fatalf("vivechak_save_session T1-02 failed: %v", err)
	}
	env2 := parseEnvelope(t, saveSess2Res)
	data2, _ := env2.Data.(map[string]any)
	filePath2, _ := data2["file_path"].(string)
	expectedDefaultPath := "research/sessions/T1-02.md"
	if filePath2 != expectedDefaultPath {
		t.Errorf("file_path for T1-02 = %q; want %q", filePath2, expectedDefaultPath)
	}
}


