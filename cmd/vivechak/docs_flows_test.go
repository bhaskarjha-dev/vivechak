package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	toolCallRegex   = regexp.MustCompile(`(?s)\*\*Tool Call:\*\*\s*` + "`" + `(\w+)` + "`" + `\s*` + "```" + `json\s*\n(.*?)\n` + "```" + `\s*\*\*Tool Response:\*\*\s*` + "```" + `json\s*\n(.*?)\n` + "```")
	pathRegex       = regexp.MustCompile(`research/[\w./-]+\.md`)
	inlineCallRegex = regexp.MustCompile(`\b(vivechak_[a-z0-9_]+)\(([^)\n]*)\)`)
	kvArgRegex      = regexp.MustCompile(`^\s*([a-zA-Z0-9_]+)\s*[:=]\s*(.+?)\s*$`)
)

type docToolStep struct {
	toolName string
	args     map[string]any
	rawArgs  string
	rawResp  string
}

type inlineCallSite struct {
	file     string
	line     int
	raw      string
	toolName string
	args     map[string]any
}

type liveToolSchema struct {
	required []string
	props    map[string]livePropSchema
}

type livePropSchema struct {
	typ  string
	enum []string
}

// TestDocsFlows_ReplaysExampleWalkthroughs verifies that every tool call documented in
// examples/*.md replays cleanly against the live compiled Vivechak binary over stdio MCP.
func TestDocsFlows_ReplaysExampleWalkthroughs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live docs flows replay test in short mode")
	}

	binPath := resolveOrBuildTestBinary(t)

	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}

	exampleFiles, err := filepath.Glob(filepath.Join(repoRoot, "examples", "*.md"))
	if err != nil || len(exampleFiles) == 0 {
		t.Fatalf("no example markdown files found in %s: %v", filepath.Join(repoRoot, "examples"), err)
	}

	for _, exFile := range exampleFiles {
		baseName := filepath.Base(exFile)
		t.Run(baseName, func(t *testing.T) {
			contentBytes, err := os.ReadFile(exFile)
			if err != nil {
				t.Fatalf("reading %s: %v", exFile, err)
			}
			content := string(contentBytes)

			matches := toolCallRegex.FindAllStringSubmatch(content, -1)
			if len(matches) == 0 {
				t.Skipf("no tool call pairs found in %s", baseName)
			}

			var steps []docToolStep
			for _, m := range matches {
				name := m[1]
				argsJSON := m[2]
				respJSON := m[3]

				var args map[string]any
				if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
					t.Fatalf("%s: invalid JSON in tool call arguments for %s: %v", baseName, name, err)
				}
				steps = append(steps, docToolStep{
					toolName: name,
					args:     args,
					rawArgs:  argsJSON,
					rawResp:  respJSON,
				})
			}

			projDir := t.TempDir()

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			session, cleanup := startLiveMCPServer(t, ctx, binPath, projDir)
			defer cleanup()

			for i, step := range steps {
				stepArgs := make(map[string]any, len(step.args)+1)
				for k, v := range step.args {
					stepArgs[k] = v
				}
				stepArgs["project_root"] = projDir

				res, err := session.CallTool(ctx, &mcp.CallToolParams{
					Name:      step.toolName,
					Arguments: stepArgs,
				})
				if err != nil {
					t.Fatalf("%s step %d (%s): CallTool transport error: %v", baseName, i+1, step.toolName, err)
				}
				if res.IsError {
					t.Fatalf("%s step %d (%s): returned isError: true. Response: %s", baseName, i+1, step.toolName, extractFirstText(res))
				}

				liveText := extractFirstText(res)
				documentedPaths := pathRegex.FindAllString(step.rawResp, -1)
				for _, p := range documentedPaths {
					if !strings.Contains(liveText, p) {
						t.Errorf("%s step %d (%s): documented path %q not found in real response:\n%s",
							baseName, i+1, step.toolName, p, liveText)
					}
				}
			}
		})
	}
}

// TestDocsFlows_InlineCallsAndScopeSequences extracts every inline vivechak_*(...) call from
// README.md, AGENTS.md, skills/vivechak/SKILL.md, templates/rules/*, docs/QUICKSTART.md,
// docs/MCP-TOOLS.md, and docs/TROUBLESHOOTING.md, validates every call against live tools/list
// schemas (rejecting missing required args, unknown/alias args, or invalid enum values), and
// replays each scope's documented sequence and inline calls over live stdio.
func TestDocsFlows_InlineCallsAndScopeSequences(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live inline docs flows test in short mode")
	}

	binPath := resolveOrBuildTestBinary(t)
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}

	docRelPaths := []string{
		"README.md",
		"AGENTS.md",
		"skills/vivechak/SKILL.md",
		"docs/QUICKSTART.md",
		"docs/MCP-TOOLS.md",
		"docs/TROUBLESHOOTING.md",
	}
	ruleMatches, err := filepath.Glob(filepath.Join(repoRoot, "templates", "rules", "*"))
	if err != nil || len(ruleMatches) == 0 {
		t.Fatalf("expected rule templates in templates/rules/*: %v", err)
	}
	for _, rm := range ruleMatches {
		rel, _ := filepath.Rel(repoRoot, rm)
		docRelPaths = append(docRelPaths, filepath.ToSlash(rel))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	schemaDir := t.TempDir()
	schemaSession, schemaCleanup := startLiveMCPServer(t, ctx, binPath, schemaDir)
	defer schemaCleanup()

	schemas := fetchLiveToolSchemas(t, ctx, schemaSession)

	var allCalls []inlineCallSite
	callsByFile := make(map[string][]inlineCallSite)
	for _, rel := range docRelPaths {
		absPath := filepath.Join(repoRoot, filepath.FromSlash(rel))
		rawBytes, rErr := os.ReadFile(absPath)
		if rErr != nil {
			t.Fatalf("reading %s: %v", rel, rErr)
		}
		calls, pErr := extractInlineCalls(rel, string(rawBytes))
		if pErr != nil {
			t.Fatalf("%v", pErr)
		}
		allCalls = append(allCalls, calls...)
		callsByFile[rel] = calls
	}
	if len(allCalls) == 0 {
		t.Fatal("extracted zero inline vivechak_*(...) calls from documentation")
	}

	// 1. Validate every inline call site against live tools/list schema.
	for _, c := range allCalls {
		if err := validateInlineCallAgainstSchema(c, schemas); err != nil {
			t.Errorf("%v", err)
		}
	}
	if t.Failed() {
		t.FailNow()
	}

	// 2. Replay every extracted inline call over live stdio against pre-seeded workspaces,
	// proving that the server accepts every documented call with zero W-UNKNOWN-ARG warnings.
	t.Run("LiveInlineCallsReplay", func(t *testing.T) {
		seedTemplate := setupSeededReplayWorkspace(t, ctx, binPath)
		replaySess, replayClean := startLiveMCPServer(t, ctx, binPath, seedTemplate)
		defer replayClean()

		for idx, c := range allCalls {
			wsDir := copyWorkspaceDir(t, seedTemplate)
			execArgs := materializeInlineArgs(c.toolName, c.args, wsDir)
			res, callErr := replaySess.CallTool(ctx, &mcp.CallToolParams{
				Name:      c.toolName,
				Arguments: execArgs,
			})
			if callErr != nil {
				t.Fatalf("inline call #%d (%s:%d %s): transport/schema error: %v", idx+1, c.file, c.line, c.raw, callErr)
			}
			txt := extractFirstText(res)
			if res.IsError {
				t.Fatalf("inline call #%d (%s:%d %s): returned isError=true: %s", idx+1, c.file, c.line, c.raw, txt)
			}
			if strings.Contains(txt, "W-UNKNOWN-ARG") {
				t.Fatalf("inline call #%d (%s:%d %s): emitted W-UNKNOWN-ARG (alias/unknown arg): %s", idx+1, c.file, c.line, c.raw, txt)
			}
		}
	})

	// 3. Replay each scope's documented sequence extracted from AGENTS.md and skills/vivechak/SKILL.md
	// end-to-end in fresh workspaces over stdio.
	for _, docFile := range []string{"AGENTS.md", "skills/vivechak/SKILL.md", "docs/QUICKSTART.md"} {
		fileCalls := callsByFile[docFile]
		for _, scope := range []string{"comparison", "decision", "project"} {
			seq := extractScopeSequence(fileCalls, scope)
			if len(seq) == 0 {
				t.Fatalf("%s: found no documented call sequence for scope %q", docFile, scope)
			}
			t.Run(fmt.Sprintf("Sequence_%s_%s", strings.ReplaceAll(docFile, "/", "_"), scope), func(t *testing.T) {
				wsDir := t.TempDir()
				sess, clean := startLiveMCPServer(t, ctx, binPath, wsDir)
				defer clean()

				for stepIdx, step := range seq {
					args := materializeSequenceStepArgs(scope, step.toolName, step.args, wsDir)
					res, callErr := sess.CallTool(ctx, &mcp.CallToolParams{
						Name:      step.toolName,
						Arguments: args,
					})
					if callErr != nil {
						t.Fatalf("%s (%s) step %d %s: transport error: %v", docFile, scope, stepIdx+1, step.raw, callErr)
					}
					txt := extractFirstText(res)
					if res.IsError {
						t.Fatalf("%s (%s) step %d %s: isError=true: %s", docFile, scope, stepIdx+1, step.raw, txt)
					}
					if strings.Contains(txt, "W-UNKNOWN-ARG") {
						t.Fatalf("%s (%s) step %d %s: W-UNKNOWN-ARG: %s", docFile, scope, stepIdx+1, step.raw, txt)
					}
				}
			})
		}
	}
}

// TestDocsFlows_FailsOnMissingRequiredArg verifies negative schema rejection on missing required
// parameters and alias parameters.
func TestDocsFlows_FailsOnMissingRequiredArg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live test in short mode")
	}

	binPath := resolveOrBuildTestBinary(t)
	projDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	session, cleanup := startLiveMCPServer(t, ctx, binPath, projDir)
	defer cleanup()

	schemas := fetchLiveToolSchemas(t, ctx, session)

	// 1. Missing required argument in inline doc snippet must fail validateInlineCallAgainstSchema
	missingCalls, err := extractInlineCalls("scratch.md", `Call vivechak_save(kind="comparison")`)
	if err != nil || len(missingCalls) != 1 {
		t.Fatalf("extractInlineCalls failed: %v", err)
	}
	if err := validateInlineCallAgainstSchema(missingCalls[0], schemas); err == nil {
		t.Fatal("expected validateInlineCallAgainstSchema to reject vivechak_save missing required content")
	}

	// 2. Alias parameter in inline doc snippet must fail validateInlineCallAgainstSchema
	aliasCalls, err := extractInlineCalls("scratch.md", `Call vivechak_record_decision(decision_id="D-001", content="x")`)
	if err != nil || len(aliasCalls) != 1 {
		t.Fatalf("extractInlineCalls failed: %v", err)
	}
	if err := validateInlineCallAgainstSchema(aliasCalls[0], schemas); err == nil {
		t.Fatal("expected validateInlineCallAgainstSchema to reject retired alias decision_id")
	}

	// 3. Live stdio server must also reject vivechak_save missing required kind
	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": projDir,
			"content":      "# Test",
		},
	})
	if err == nil && !res.IsError {
		t.Fatalf("expected vivechak_save without required kind to fail, but got success")
	}
}

func extractFirstText(res *mcp.CallToolResult) string {
	if res == nil || len(res.Content) == 0 {
		return ""
	}
	if tc, ok := res.Content[0].(*mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}

func fetchLiveToolSchemas(t *testing.T, ctx context.Context, session *mcp.ClientSession) map[string]liveToolSchema {
	t.Helper()
	lt, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
	out := make(map[string]liveToolSchema, len(lt.Tools))
	for _, tool := range lt.Tools {
		rawSchema, mErr := json.Marshal(tool.InputSchema)
		if mErr != nil {
			t.Fatalf("marshaling inputSchema for %s: %v", tool.Name, mErr)
		}
		var parsed struct {
			Required   []string `json:"required"`
			Properties map[string]struct {
				Type string   `json:"type"`
				Enum []string `json:"enum"`
			} `json:"properties"`
		}
		if uErr := json.Unmarshal(rawSchema, &parsed); uErr != nil {
			t.Fatalf("unmarshaling inputSchema for %s: %v", tool.Name, uErr)
		}
		props := make(map[string]livePropSchema, len(parsed.Properties))
		for pn, pv := range parsed.Properties {
			props[pn] = livePropSchema{typ: pv.Type, enum: pv.Enum}
		}
		out[tool.Name] = liveToolSchema{
			required: parsed.Required,
			props:    props,
		}
	}
	return out
}

func extractInlineCalls(relFile, content string) ([]inlineCallSite, error) {
	var out []inlineCallSite
	lines := strings.Split(content, "\n")
	for idx, line := range lines {
		matches := inlineCallRegex.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			raw := m[0]
			name := m[1]
			argList := strings.TrimSpace(m[2])
			args, err := parseInlineArgList(argList)
			if err != nil {
				return nil, fmt.Errorf("%s:%d: invalid inline call %s: %w", relFile, idx+1, raw, err)
			}
			out = append(out, inlineCallSite{
				file:     relFile,
				line:     idx + 1,
				raw:      raw,
				toolName: name,
				args:     args,
			})
		}
	}
	return out, nil
}

func parseInlineArgList(s string) (map[string]any, error) {
	out := make(map[string]any)
	if s == "" {
		return out, nil
	}
	parts := splitTopLevelCommas(s)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		m := kvArgRegex.FindStringSubmatch(p)
		if len(m) != 3 {
			return nil, fmt.Errorf("expected key=value argument, got %q", p)
		}
		key := m[1]
		valRaw := strings.TrimSpace(m[2])
		out[key] = parseInlineValue(valRaw)
	}
	return out, nil
}

func splitTopLevelCommas(s string) []string {
	var parts []string
	var cur strings.Builder
	var quote rune
	bracketDepth := 0
	for _, r := range s {
		switch {
		case quote != 0:
			cur.WriteRune(r)
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
			cur.WriteRune(r)
		case r == '[':
			bracketDepth++
			cur.WriteRune(r)
		case r == ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			cur.WriteRune(r)
		case r == ',' && bracketDepth == 0:
			parts = append(parts, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

func parseInlineValue(raw string) any {
	if (strings.HasPrefix(raw, `"`) && strings.HasSuffix(raw, `"`)) ||
		(strings.HasPrefix(raw, `'`) && strings.HasSuffix(raw, `'`)) {
		return raw[1 : len(raw)-1]
	}
	if strings.EqualFold(raw, "true") {
		return true
	}
	if strings.EqualFold(raw, "false") {
		return false
	}
	return raw
}

func isDocPlaceholder(s string) bool {
	trimmed := strings.TrimSpace(s)
	return trimmed == "..." ||
		(strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">")) ||
		strings.Contains(trimmed, "<") ||
		strings.Contains(trimmed, "NNN")
}

func validateInlineCallAgainstSchema(c inlineCallSite, schemas map[string]liveToolSchema) error {
	sch, ok := schemas[c.toolName]
	if !ok {
		return fmt.Errorf("%s:%d: unknown or retired tool %q in %s", c.file, c.line, c.toolName, c.raw)
	}
	for k, v := range c.args {
		prop, exists := sch.props[k]
		if !exists {
			return fmt.Errorf("%s:%d: unknown or alias argument %q in %s", c.file, c.line, k, c.raw)
		}
		if len(prop.enum) > 0 {
			if sv, ok := v.(string); ok && !isDocPlaceholder(sv) && !slices.Contains(prop.enum, sv) {
				return fmt.Errorf("%s:%d: argument %q=%q not in enum %v in %s", c.file, c.line, k, sv, prop.enum, c.raw)
			}
		}
	}
	for _, req := range sch.required {
		if _, present := c.args[req]; !present {
			return fmt.Errorf("%s:%d: missing required argument %q in %s", c.file, c.line, req, c.raw)
		}
	}
	return nil
}

const (
	docsFlowComparisonFixture   = "---\nid: C-001\ntitle: \"Pino vs Winston\"\ndate: 2026-10-25\nstatus: complete\nscope_level: comparison\n---\n## Prior\n- Pino is faster.\n\n## Research Question\nWhich JSON logger fits Node.js?\n\n## Key Findings\n- Pino achieves 5x throughput [A · single · fresh · direct · fetched](https://getpino.io)\n- Winston has broader transports [B · single · fresh · direct · fetched](https://github.com/winstonjs/winston)\n\n## Recommendation\nUse Pino [A · single · fresh · direct · fetched](https://getpino.io).\n\n## Weighted Evaluation Matrix\n| Criterion | Weight | Pino | Winston |\n|---|---|---|---|\n| Throughput | 0.6 | 5 | 3 |\n| Ecosystem | 0.4 | 4 | 5 |\n| **Weighted total** | 1.0 | 4.6 | 3.8 |\n\n## Sources & Evidence Ledger\n| # | Source | Grade |\n|---|---|---|\n| 1 | https://getpino.io | A |\n"
	docsFlowDecisionPlanFixture = "# D-001-PLAN: Event storage\n\n```yaml vivechak-plan\nplan: D-001\nsessions:\n  - {id: D-001-S1, title: Landscape, role: L, depends: []}\n```\n\n## Sessions\n\n### D-001-S1: Landscape\n```prompt\nMap credible event storage engines.\n```\n"
	docsFlowProjectPlanFixture  = "# Research Pipeline: Demo\n\n```yaml vivechak-plan\nplan: PROJECT\nsessions:\n  - {id: T1-01, title: Storage, layer: 1, door: two-way, informs: [D-001], depends: []}\n  - {id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}\ndecisions:\n  - {id: D-001, title: Primary storage, door: two-way}\n```\n\n## Sessions\n\n### T1-01: Storage\n```prompt\nEvaluate storage options. [UPSTREAM_FINDINGS]\n```\n\n### SYN-01: Synthesis\n```prompt\nSynthesize findings. [ALL_SESSION_FINDINGS]\n```\n"
	docsFlowSessionFixture      = "---\nid: T1-01\ntitle: \"Storage\"\ndate: 2026-10-25\nstatus: complete\ninforms_decisions: [D-001]\n---\n## Prior\n- PostgreSQL fits.\n\n## Research Question\nWhich datastore fits?\n\n## Key Findings\n- PostgreSQL supports JSONB indexing [A · single · fresh · direct · fetched](https://www.postgresql.org/docs/current/datatype-json.html)\n- Replication supports upgrades [A · single · fresh · direct · fetched](https://www.postgresql.org/docs/current/logical-replication.html)\n\n## Recommendation\nUse PostgreSQL [A · single · fresh · direct · fetched](https://www.postgresql.org/docs/current/).\n\n## Discovered Concerns\nWrite amplification under heavy update workloads requires benchmarking before launch.\n\n## Delta\n| Prior Belief | Status | Evidence | Impact |\n|---|---|---|---|\n| PostgreSQL fits | Confirmed | docs | none |\n\n## Sources & Evidence Ledger\n| # | Source | Grade |\n|---|---|---|\n| 1 | https://www.postgresql.org/docs/current/ | A |\n"
	docsFlowADRFixture          = "---\nid: D-001\ntitle: \"Primary storage\"\nstatus: accepted\ndoor_type: two-way\ndate: 2026-10-25\nreview_trigger: \"p95 latency > 50 ms for 7 days\"\ninformed_by_sessions: [T1-01]\n---\n# D-001: Primary storage\n## Context & Problem Statement\nSelect primary datastore.\n## Decision Outcome\nAdopt PostgreSQL [A · single · fresh · direct · fetched](https://www.postgresql.org/docs/current/).\n## Rejected Alternatives & Tradeoffs\nMongoDB rejected due to higher operational overhead for relational queries.\n## Failure Modes & Reversal Triggers\nRevisit if p95 latency stays above 50 ms for 7 days.\n"
	docsFlowFADFixture          = "---\nid: SYN-01\ntitle: \"Founding Architecture\"\ndate: 2026-10-25\nstatus: complete\n---\n# Founding Architecture Document\n\n## Executive Summary\nPostgreSQL-backed event service [A · single · fresh · direct · fetched](https://www.postgresql.org/docs/current/).\n\n## System Architecture\nStateless Go API over PostgreSQL.\n\n## Locked Decisions Registry\n| ID | Status |\n|---|---|\n| D-001 | accepted |\n\n## Cross-Session Conflict Resolutions\nNo unresolved conflicts across sessions.\n\n## Risks & Review Schedule\nRevisit D-001 if p95 latency exceeds 50 ms.\n"
)

func setupSeededReplayWorkspace(t *testing.T, ctx context.Context, binPath string) string {
	t.Helper()
	wsDir := t.TempDir()
	sess, clean := startLiveMCPServer(t, ctx, binPath, wsDir)
	defer clean()

	mustCall := func(name string, args map[string]any) {
		args["project_root"] = wsDir
		res, err := sess.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil || res.IsError {
			t.Fatalf("seed %s failed: err=%v text=%s", name, err, extractFirstText(res))
		}
	}
	mustCall("vivechak_start", map[string]any{"scope": "project", "question": "Demo telemetry service"})
	mustCall("vivechak_save", map[string]any{"kind": "plan", "scope": "project", "id": "DAG", "content": docsFlowProjectPlanFixture})
	mustCall("vivechak_save", map[string]any{"kind": "session", "id": "T1-01", "content": docsFlowSessionFixture})
	mustCall("vivechak_save", map[string]any{"kind": "session", "id": "D-001-S1", "content": strings.ReplaceAll(docsFlowSessionFixture, "id: T1-01", "id: D-001-S1")})
	mustCall("vivechak_record_decision", map[string]any{"id": "D-001", "content": docsFlowADRFixture})
	return wsDir
}

func copyWorkspaceDir(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copyWorkspaceDir: %v", err)
	}
	return dst
}

func materializeInlineArgs(toolName string, rawArgs map[string]any, wsDir string) map[string]any {
	out := make(map[string]any, len(rawArgs)+2)
	for k, v := range rawArgs {
		out[k] = v
	}
	out["project_root"] = wsDir

	if sv, ok := out["question"].(string); ok && isDocPlaceholder(sv) {
		out["question"] = "Evaluate Pino vs Winston for JSON logging"
	}
	if idStr, ok := out["id"].(string); ok {
		switch {
		case strings.Contains(idStr, "NNN-S1") || idStr == "<session_id>":
			out["id"] = "D-001-S1"
		case strings.Contains(idStr, "NNN") || isDocPlaceholder(idStr):
			if _, hasDraft := out["auto_draft_from"]; hasDraft {
				out["id"] = "D-002"
			} else {
				out["id"] = "D-001"
			}
		}
	}
	if dfStr, ok := out["auto_draft_from"].(string); ok && isDocPlaceholder(dfStr) {
		out["auto_draft_from"] = "D-001-S1"
	}
	if decStr, ok := out["decision"].(string); ok && isDocPlaceholder(decStr) {
		out["decision"] = "D-001"
	}

	switch toolName {
	case "vivechak_save":
		kind, _ := out["kind"].(string)
		out["revise"] = true
		switch kind {
		case "comparison":
			out["content"] = docsFlowComparisonFixture
		case "plan":
			idVal, _ := out["id"].(string)
			if strings.EqualFold(idVal, "DAG") || out["scope"] == "project" {
				out["content"] = docsFlowProjectPlanFixture
			} else {
				out["id"] = "D-001"
				out["content"] = docsFlowDecisionPlanFixture
			}
		case "session":
			out["content"] = docsFlowSessionFixture
		case "fad":
			out["content"] = docsFlowFADFixture
		}
	case "vivechak_record_decision":
		out["revise"] = true
		if _, hasContent := out["content"]; hasContent {
			out["content"] = docsFlowADRFixture
		}
	}
	return out
}

func extractScopeSequence(calls []inlineCallSite, scope string) []inlineCallSite {
	startIdx := -1
	for i, c := range calls {
		if c.toolName == "vivechak_start" && c.args["scope"] == scope {
			// Pick the first sequence that actually has follow-up calls (not a standalone start bullet).
			if i+1 < len(calls) && calls[i+1].toolName != "vivechak_start" {
				startIdx = i
				break
			}
		}
	}
	if startIdx == -1 {
		return nil
	}
	var seq []inlineCallSite
	seq = append(seq, calls[startIdx])
	for i := startIdx + 1; i < len(calls); i++ {
		if calls[i].toolName == "vivechak_start" {
			break
		}
		seq = append(seq, calls[i])
	}
	return seq
}

func materializeSequenceStepArgs(scope, toolName string, rawArgs map[string]any, wsDir string) map[string]any {
	out := materializeInlineArgs(toolName, rawArgs, wsDir)
	if scope == "decision" {
		if toolName == "vivechak_save" && out["kind"] == "session" {
			out["id"] = "D-001-S1"
			out["content"] = strings.ReplaceAll(docsFlowSessionFixture, "id: T1-01", "id: D-001-S1")
		}
		if toolName == "vivechak_record_decision" {
			if _, hasDraft := out["auto_draft_from"]; hasDraft {
				out["auto_draft_from"] = "D-001-S1"
			} else {
				out["content"] = strings.ReplaceAll(docsFlowADRFixture, "[T1-01]", "[D-001-S1]")
			}
		}
	}
	if scope == "project" {
		if toolName == "vivechak_save" && out["kind"] == "session" {
			out["id"] = "T1-01"
			out["content"] = docsFlowSessionFixture
		}
		if toolName == "vivechak_record_decision" {
			if _, hasDraft := out["auto_draft_from"]; hasDraft {
				out["auto_draft_from"] = "T1-01"
			}
		}
	}
	return out
}
