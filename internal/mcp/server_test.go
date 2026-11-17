package mcputil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMain(m *testing.M) {
	code := m.Run()
	_ = os.RemoveAll("research")
	os.Exit(code)
}

// testServer creates a Vivechak MCP server connected to in-memory transports.
func testServer(t *testing.T) *mcp.ClientSession {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	ctx := context.Background()

	server := NewServer("0.1.0-test", logger)

	st, ct := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "test-client", Version: "1.0.0"},
		nil,
	)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	return cs
}

// parseEnvelope extracts the Envelope from a CallToolResult's text content.
func parseEnvelope(t *testing.T, result *mcp.CallToolResult) Envelope {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("no text content in result")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var env Envelope
	if err := json.Unmarshal([]byte(tc.Text), &env); err != nil {
		t.Fatalf("parsing envelope: %v (text was: %s)", err, tc.Text)
	}
	return env
}

func envGuidance(env Envelope) string {
	if env.Fix != "" {
		return env.Fix
	}
	if env.Next != nil {
		return env.Next.Why
	}
	return ""
}

func envWarningStrings(env Envelope) []string {
	out := make([]string, len(env.Warnings))
	for i, w := range env.Warnings {
		out[i] = w.Code + ": " + w.Message
	}
	return out
}

func envWarningsJoined(env Envelope) string {
	return strings.Join(envWarningStrings(env), "\n")
}

// redTeamRecordID makes the real red-team call, vivechak_review(mode="red_team", id=decisionID), over the MCP client
// and returns the record_id it reports, after checking the challenge record it wrote. Gate tests cite this id in
// the ADR instead of writing a record with a hash they compute themselves (CRIT-13).
func redTeamRecordID(t *testing.T, ctx context.Context, cs *mcp.ClientSession, root, decisionID string) string {
	t.Helper()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_review",
		Arguments: map[string]any{"project_root": root, "mode": "red_team", "id": decisionID},
	})
	if err != nil {
		t.Fatalf("vivechak_review red_team %s: %v", decisionID, err)
	}
	return checkRedTeamRecord(t, root, decisionID, parseEnvelope(t, res))
}

// redTeamRecordIDDirect is redTeamRecordID for tests that drive the tool handlers directly.
func redTeamRecordIDDirect(t *testing.T, ctx context.Context, root, decisionID string) string {
	t.Helper()
	_, env, err := handleReview(ctx, nil, ReviewInput{ProjectRoot: root, Mode: "red_team", ID: decisionID})
	if err != nil {
		t.Fatalf("vivechak_review red_team %s: %v", decisionID, err)
	}
	return checkRedTeamRecord(t, root, decisionID, env)
}

func checkRedTeamRecord(t *testing.T, root, decisionID string, env Envelope) string {
	t.Helper()
	if !env.Ok {
		t.Fatalf("vivechak_review red_team %s failed: %s", decisionID, env.Message)
	}
	raw, err := json.Marshal(env.Data)
	if err != nil {
		t.Fatalf("marshal red_team data: %v", err)
	}
	var data struct {
		RecordID     string `json:"record_id"`
		DecisionHash string `json:"decision_hash"`
	}
	if err := json.Unmarshal(raw, &data); err != nil || data.RecordID == "" || data.DecisionHash == "" {
		t.Fatalf("vivechak_review red_team %s returned no record_id/decision_hash (err=%v): %s", decisionID, err, raw)
	}
	rec, err := os.ReadFile(filepath.Join(root, core.ResearchDir, data.RecordID+".md"))
	if err != nil {
		t.Fatalf("challenge record %s was not written: %v", data.RecordID, err)
	}
	if !strings.Contains(string(rec), "type: challenge") || !strings.Contains(string(rec), "decision_hash: "+data.DecisionHash) {
		t.Fatalf("challenge record %s lacks type: challenge or the reported decision_hash:\n%s", data.RecordID, rec)
	}
	return data.RecordID
}

// recordDecisionRevise records or revises an ADR through vivechak_record_decision and fails the test if it is refused.
func recordDecisionRevise(t *testing.T, ctx context.Context, cs *mcp.ClientSession, root, decisionID, content string) {
	t.Helper()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_record_decision",
		Arguments: map[string]any{"project_root": root, "id": decisionID, "content": content, "revise": true},
	})
	if err != nil {
		t.Fatalf("record_decision %s: %v", decisionID, err)
	}
	if env := parseEnvelope(t, res); !env.Ok {
		t.Fatalf("record_decision %s refused: %s", decisionID, env.Message)
	}
}

// TestToolListing verifies all 9 consolidated tools are registered.
func TestToolListing(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	var tools []*mcp.Tool
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		tools = append(tools, tool)
	}

	expected := []string{
		"vivechak_start",
		"vivechak_save",
		"vivechak_record_decision",
		"vivechak_next",
		"vivechak_status",
		"vivechak_review",
		"vivechak_amend",
		"vivechak_replan",
		"vivechak_gate",
	}

	if len(tools) != len(expected) {
		names := make([]string, len(tools))
		for i, t := range tools {
			names[i] = t.Name
		}
		t.Fatalf("expected %d tools, got %d: %v", len(expected), len(tools), names)
	}

	nameSet := map[string]bool{}
	for _, tool := range tools {
		nameSet[tool.Name] = true
	}

	for _, name := range expected {
		if !nameSet[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}

// TestP4_ContextBudget verifies the HIGH-06 wire context budgets and schema invariants:
// - ServerInstructions <= 900 bytes (and <= 600 per P4.4), no MANDATORY or ALWAYS
// - tools/list <= 7168 bytes (7 KB), no outputSchema, required list present on all 9 tools, no alias params
// - always-loaded context (instructions + tools/list + prompts/list + resources/list) <= 10240 bytes (10 KB)
// - unknown argument ignored with W-UNKNOWN-ARG
func TestP4_ContextBudget(t *testing.T) {
	instructionsLen := len(ServerInstructions)
	if instructionsLen > 600 {
		t.Errorf("ServerInstructions length is %d, want <= 600", instructionsLen)
	}
	if strings.Contains(ServerInstructions, "MANDATORY") || strings.Contains(ServerInstructions, "ALWAYS") {
		t.Errorf("ServerInstructions must not contain MANDATORY or ALWAYS")
	}

	cs := testServer(t)
	ctx := context.Background()

	listToolsRes, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	toolsBytes, err := json.Marshal(listToolsRes)
	if err != nil {
		t.Fatalf("marshaling ListTools: %v", err)
	}

	forbiddenAliases := map[string][]string{
		"vivechak_start":           {"context"},
		"vivechak_save":            {"session_id", "decision_id", "pipeline_content"},
		"vivechak_record_decision": {"decision_id"},
		"vivechak_next":            {"id"},
		"vivechak_status":          {"root"},
		"vivechak_review":          {"target", "session_id", "decision_id", "challenge_type"},
		"vivechak_amend":           {"session_id", "target", "amendment", "content", "append_content", "append"},
		"vivechak_replan":          {"operation", "new_prompt", "new_dependencies"},
	}

	for _, tool := range listToolsRes.Tools {
		if tool.OutputSchema != nil {
			t.Errorf("%s must not carry outputSchema", tool.Name)
		}
		if len(tool.Description) > 350 {
			t.Errorf("%s description length %d > 350", tool.Name, len(tool.Description))
		}
		if strings.Contains(tool.Description, "MANDATORY") || strings.Contains(tool.Description, "ALWAYS") {
			t.Errorf("%s description contains MANDATORY or ALWAYS", tool.Name)
		}
		rawSchema, _ := json.Marshal(tool.InputSchema)
		var schemaMap map[string]any
		_ = json.Unmarshal(rawSchema, &schemaMap)
		if _, hasReq := schemaMap["required"]; !hasReq {
			t.Errorf("%s inputSchema missing required list", tool.Name)
		}
		props, _ := schemaMap["properties"].(map[string]any)
		for _, alias := range forbiddenAliases[tool.Name] {
			if _, exists := props[alias]; exists {
				t.Errorf("%s inputSchema still exposes removed alias %q", tool.Name, alias)
			}
		}
		for pname, pval := range props {
			if pm, ok := pval.(map[string]any); ok {
				if desc, _ := pm["description"].(string); len(desc) > 100 {
					t.Errorf("%s.%s description length %d > 100", tool.Name, pname, len(desc))
				}
			}
		}
	}

	promptsRes, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}
	promptsBytes, _ := json.Marshal(promptsRes)

	resourcesRes, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources: %v", err)
	}
	resourcesBytes, _ := json.Marshal(resourcesRes)

	// Account for Python json.dumps ", " and ": " spacing (~8% larger than compact json.Marshal) by keeping strict headroom.
	if len(toolsBytes) > 7168 {
		t.Errorf("tools/list size %d exceeds 7 KB (7168 bytes)", len(toolsBytes))
	}
	alwaysLoaded := instructionsLen + len(toolsBytes) + len(promptsBytes) + len(resourcesBytes)
	t.Logf("Context budget (compact): instructions=%d, tools/list=%d, prompts/list=%d, resources/list=%d, always_loaded=%d",
		instructionsLen, len(toolsBytes), len(promptsBytes), len(resourcesBytes), alwaysLoaded)
	if alwaysLoaded > 10240 {
		t.Errorf("always-loaded context %d exceeds 10 KB (10240 bytes)", alwaysLoaded)
	}

	// Verify unknown argument emits W-UNKNOWN-ARG without failing
	extraRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_status",
		Arguments: map[string]any{"project_root": t.TempDir(), "bogus_extra": 1},
	})
	if err != nil || extraRes.IsError {
		t.Fatalf("vivechak_status(bogus_extra=1) failed: err=%v, isError=%v", err, extraRes.IsError)
	}
	extraEnv := parseEnvelope(t, extraRes)
	if !strings.Contains(envWarningsJoined(extraEnv), "W-UNKNOWN-ARG") {
		t.Errorf("expected W-UNKNOWN-ARG on extra argument, got warnings=%v", extraEnv.Warnings)
	}
}

// TestP4_UnregisteredToolRejection verifies calling retired/unregistered tool names returns an error.
func TestP4_UnregisteredToolRejection(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	retiredCalls := []string{
		"vivechak_init",
		"vivechak_prepare_generator",
		"vivechak_save_plan",
		"vivechak_save_session",
		"vivechak_next_session",
		"vivechak_validate",
		"vivechak_challenge",
		"vivechak_visualize",
		"vivechak_amend_session",
		"vivechak_run_gate",
	}

	for _, name := range retiredCalls {
		t.Run(name, func(t *testing.T) {
			_, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name:      name,
				Arguments: map[string]any{},
			})
			if err == nil {
				t.Errorf("calling unregistered tool %s should return error, got nil", name)
			}
		})
	}
}

// TestP4_PanicRecovery verifies that an unhandled panic in a tool handler does not crash the server and returns E-INTERNAL.
func TestP4_PanicRecovery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := NewServer("0.1.0-test", logger)
	// Add a tool that deliberately panics to test recovery middleware
	mcp.AddTool(s, &mcp.Tool{
		Name:        "test_panic_tool",
		Description: "tool that panics",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, Envelope, error) {
		panic("deliberate test panic in handler")
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := s.Connect(context.Background(), serverTransport, nil); err != nil {
		t.Fatalf("connecting server: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	cs, err := client.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatalf("connecting client: %v", err)
	}
	defer cs.Close()

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "test_panic_tool",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("expected panic to be recovered without client transport error, got: %v", err)
	}
	if !res.IsError {
		t.Errorf("expected recovered panic to return IsError=true")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent from panic recovery, got %T", res.Content[0])
	}
	if !strings.Contains(tc.Text, "E-INTERNAL") {
		t.Errorf("expected panic recovery response to mention E-INTERNAL, got: %s", tc.Text)
	}
}

// TestAnnotations verifies all tools have appropriate annotations.
func TestAnnotations(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	readOnlyTools := map[string]bool{
		"vivechak_status": true,
		"vivechak_next":   true,
	}

	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		if tool.Annotations == nil {
			t.Errorf("%s: annotations are nil", tool.Name)
			continue
		}

		if readOnlyTools[tool.Name] {
			if !tool.Annotations.ReadOnlyHint {
				t.Errorf("%s: expected readOnlyHint=true", tool.Name)
			}
		} else {
			if tool.Annotations.ReadOnlyHint {
				t.Errorf("%s: expected readOnlyHint=false", tool.Name)
			}
		}
	}
}

// TestStatusNoWorkspace verifies vivechak_status works without an initialized workspace.
func TestStatusNoWorkspace(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_status",
		Arguments: map[string]any{
			"project_root": t.TempDir(),
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_status: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Ok {
		t.Error("status should succeed even without workspace")
	}
	if envGuidance(env) == "" {
		t.Error("status must always provide next_step (Guided Worker)")
	}
}

// TestInitAndStatus tests the init → status flow.
func TestInitAndStatus(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"question":     "test",
			"project_root": tmpDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_start: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Ok {
		t.Fatalf("init failed: %s", env.Message)
	}

	// Verify directory structure
	dirs := []string{"research", "research/sessions"}
	for _, dir := range dirs {
		path := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("missing directory: %s", dir)
		}
	}

	// Templates are served as MCP resources, never copied into research/templates/ (HIGH-15)
	if _, err := os.Stat(filepath.Join(tmpDir, "research", "templates")); !os.IsNotExist(err) {
		t.Errorf("expected no research/templates/, got err=%v", err)
	}

	// Status should now show initialized
	result2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_status",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_status: %v", err)
	}

	env2 := parseEnvelope(t, result2)
	if !env2.Ok {
		t.Error("status should succeed after init")
	}
}

// TestPrepareGenerator tests the prepare_generator tool.
func TestPrepareGenerator(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "project",
			"question":     "A web application for managing research projects",
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_start: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Ok {
		t.Fatalf("prepare_generator failed: %s", env.Message)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatal("data should be a map")
	}

	prompt, _ := data["brief"].(string)
	if prompt == "" {
		t.Error("brief should not be empty")
	}
	if len(prompt) < 100 {
		t.Errorf("brief seems too short: %d chars", len(prompt))
	}
}

func TestPrepareGenerator_ScopeAutoDetect(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Initialize decision-scope workspace
	_, envInit, err := handleStart(ctx, nil, StartInput{
		ProjectRoot: tmpDir,
		Scope:       "decision",
		Question:    "test",
	})
	if err != nil || !envInit.Ok {
		t.Fatalf("handleStart init: %v", err)
	}

	// Call handleStart without scope, providing project_root
	_, env, err := handleStart(ctx, nil, StartInput{
		ProjectRoot: tmpDir,
		Question:    "Choose between PostgreSQL and DynamoDB",
	})
	if err != nil || !env.Ok {
		t.Fatalf("handleStart: %v %s", err, env.Message)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatal("data should be a map")
	}

	if data["scope"] != core.ScopeDecision {
		t.Errorf("expected scope 'decision' auto-detected from workspace, got %v", data["scope"])
	}

	prompt, _ := data["brief"].(string)
	if strings.Contains(prompt, "[the question, as you'd ask it]") {
		t.Errorf("prompt still contains unreplaced DECISION placeholder")
	}
	if strings.Contains(prompt, "[what's being built; workload, team, constraints, what's already decided]") {
		t.Errorf("prompt still contains unreplaced CONTEXT placeholder")
	}
	if !strings.Contains(prompt, "Choose between PostgreSQL and DynamoDB") {
		t.Errorf("prompt does not contain injected context")
	}

	// Call handleStart without scope or project_root, using env var
	t.Setenv("VIVECHAK_PROJECT_ROOT", tmpDir)
	_, envFromEnv, err := handleStart(ctx, nil, StartInput{
		Question: "Choose between Redis and Memcached",
	})
	if err != nil || !envFromEnv.Ok {
		t.Fatalf("handleStart with env var: %v %s", err, envFromEnv.Message)
	}
	dataEnv, _ := envFromEnv.Data.(map[string]any)
	if dataEnv["scope"] != core.ScopeDecision {
		t.Errorf("expected scope 'decision' from env workspace, got %v", dataEnv["scope"])
	}
}

// TestSaveSessionValidation tests the validation ladder on session save.
func TestSaveSessionValidation(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace first
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	// Save a session with proper frontmatter
	content := `---
session_id: R-01
title: Test Session
date: 2026-09-27
status: complete
---

# Test Session

This is a test session with evidence grades.

Finding: The test framework works well. A (direct testing)

## Recommendations

Use this approach for production.
`
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "R-01",
			"content":      content,
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_save: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Ok {
		t.Fatalf("save_session failed: %s", env.Message)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatal("data should be a map")
	}
	if data["validation_passed"] != true {
		t.Errorf("expected validation_passed = true, got %v", data["validation_passed"])
	}

	// Verify file was saved
	sessionPath := filepath.Join(tmpDir, "research", "sessions", "R-01.md")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Error("session file not saved")
	}
}

// TestValidateDryRun tests the validate tool doesn't write anything.
func TestValidateDryRun(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	content := `---
session_id: R-01
title: Test
date: 2026-09-27
---

Short body.
`
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_review",
		Arguments: map[string]any{
			"artifact_type": "session",
			"content":       content,
			"mode":          "validate",
		},
	})
	if err != nil {
		t.Fatalf("calling vivechak_review: %v", err)
	}

	env := parseEnvelope(t, result)
	if !env.Ok {
		t.Fatalf("validate failed: %s", env.Message)
	}

	// Should have at least the W-NO-EVIDENCE-GRADES warning
	if len(env.Warnings) == 0 {
		t.Error("expected warnings for session without evidence grades")
	}
}

// TestGuidedWorkerPattern verifies every tool returns next_step.
func TestGuidedWorkerPattern(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Test a selection of tools
	calls := []struct {
		name string
		args map[string]any
	}{
		{"vivechak_status", map[string]any{"project_root": tmpDir}},
		{"vivechak_start", map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"}},
		{"vivechak_start", map[string]any{"project_root": tmpDir, "scope": "project", "question": "test project"}},
		{"vivechak_review", map[string]any{"mode": "validate", "content": "# test", "artifact_type": "session"}},
	}

	for _, tc := range calls {
		result, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      tc.name,
			Arguments: tc.args,
		})
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}

		env := parseEnvelope(t, result)
		if envGuidance(env) == "" {
			t.Errorf("%s: missing next/fix guidance — violates Guided Worker pattern", tc.name)
		}
		if env.Ok && (env.Next == nil || env.Next.Tool == "") {
			t.Errorf("%s: success envelope missing structured Next.Tool", tc.name)
		}
	}
}

// TestEndToEndPipelineFlow tests the full workflow:
// init → save_plan → next_session → save_session → next_session → run_gate
func TestEndToEndPipelineFlow(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Step 1: Init workspace
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	env := parseEnvelope(t, result)
	if !env.Ok {
		t.Fatalf("init failed: %s", env.Message)
	}

	// Step 2: Save a research pipeline
	pipeline := testPlan("Research Pipeline", "PROJECT", "", []string{
		"{id: T1-01, title: Database Selection, layer: 0, door: one-way, informs: [D-001], depends: []}",
		"{id: T1-02, title: Auth Strategy, layer: 0, door: two-way, informs: [D-002], depends: []}",
		"{id: SYN-01, title: Synthesis, layer: 1, kind: synthesis, depends: [T1-01, T1-02]}",
	}, testSection("T1-01", "Database Selection", "# RESEARCH BRIEF: T1-01\n\n## BRIEF\nInvestigate database options.\n\n## DELIVERABLE\nRecommendation with evidence grades.")+"---\n\n",
		testSection("T1-02", "Auth Strategy", "# RESEARCH BRIEF: T1-02\n\n## BRIEF\nEvaluate auth approaches.\n\n## DELIVERABLE\nComparison matrix.")+"---\n\n",
		testSection("SYN-01", "Synthesis", "# SYNTHESIS\n\nCompile findings.\n\n[ALL_SESSION_FINDINGS]"))
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "plan",
			"content":      pipeline,
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}
	env = parseEnvelope(t, result)
	if !env.Ok {
		t.Fatalf("save_plan failed: %s", env.Message)
	}

	// Step 3: Get next session (should return T1-01 or T1-02)
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}
	env = parseEnvelope(t, result)
	if !env.Ok {
		t.Fatalf("next_session failed: %s", env.Message)
	}

	data, _ := env.Data.(map[string]any)
	sessionID, _ := data["session_id"].(string)
	if sessionID != "T1-01" && sessionID != "T1-02" {
		t.Errorf("expected T1-01 or T1-02, got %s", sessionID)
	}

	// Check that parallel sessions are detected
	otherReady := data["other_ready_sessions"]
	if otherReady == nil {
		t.Log("No parallel sessions detected (may be correct if T1-01 returned first)")
	}

	// Verify prompt was included
	prompt, _ := data["prompt"].(string)
	if prompt == "" {
		t.Error("next_session should include the prompt")
	}

	// Step 4: Save T1-01 session output
	t101Output := `---
session_id: T1-01
title: Database Selection
date: 2026-09-27
status: complete
---

# Database Selection

## Findings

PostgreSQL 16 is the best choice. A (official documentation)

## Recommendations

Use PostgreSQL with pgvector for embeddings.

## Discovered Concerns

Found connection limits and memory spikes under high concurrency requiring connection pooling.
`
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "T1-01",
			"content":      t101Output,
		},
	})
	if err != nil {
		t.Fatalf("save_session T1-01: %v", err)
	}
	env = parseEnvelope(t, result)
	if !env.Ok {
		t.Fatalf("save_session T1-01 failed: %s", env.Message)
	}

	// Step 5: Save T1-02 session output
	t102Output := `---
session_id: T1-02
title: Auth Strategy
date: 2026-09-27
status: complete
---

# Auth Strategy

## Findings

Clerk provides the best developer experience. B (comparison analysis)

## Recommendations

Use Clerk for authentication.
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "T1-02",
			"content":      t102Output,
		},
	})
	if err != nil {
		t.Fatalf("save_session T1-02: %v", err)
	}

	// Step 6: Get next session — should be SYN-01 with findings injected
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("next_session for SYN-01: %v", err)
	}
	env = parseEnvelope(t, result)
	if !env.Ok {
		t.Fatalf("next_session SYN-01 failed: %s", env.Message)
	}

	data, _ = env.Data.(map[string]any)
	sessionID, _ = data["session_id"].(string)
	if sessionID != "SYN-01" {
		t.Errorf("expected SYN-01, got %s", sessionID)
	}

	// Verify context injection happened
	synPrompt, _ := data["prompt"].(string)
	if synPrompt == "" {
		t.Error("SYN-01 should have a prompt")
	}

	// Step 7: Run gate (should fail — no FAD yet)
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env = parseEnvelope(t, result)
	if !env.Ok {
		t.Fatalf("run_gate failed: %s", env.Message)
	}

	data, _ = env.Data.(map[string]any)
	if data["verdict"] != "FAIL" {
		t.Errorf("gate should FAIL without FAD, got verdict %v", data["verdict"])
	}
}

// TestFADSaveAndGatePass tests the full workflow:
// init → save_plan → save_session x3 (sessions + SYN) → save_session (FAD) → run_gate
func TestFADSaveAndGatePass(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	var result *mcp.CallToolResult
	// Step 1: Init workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Step 2: Save a research pipeline
	pipeline := testPlan("Research Pipeline", "PROJECT", "", []string{
		"{id: T1-01, title: Database Selection, layer: 0, door: two-way, informs: [D-001], depends: []}",
		"{id: T1-02, title: Auth Strategy, layer: 0, door: two-way, informs: [D-002], depends: []}",
		"{id: SYN-01, title: Synthesis, layer: 1, kind: synthesis, depends: [T1-01, T1-02]}",
	}, testSection("T1-01", "Database Selection", "# RESEARCH BRIEF: T1-01\n\n## BRIEF\nInvestigate database options.\n\n## DELIVERABLE\nRecommendation with evidence grades.")+"---\n\n",
		testSection("T1-02", "Auth Strategy", "# RESEARCH BRIEF: T1-02\n\n## BRIEF\nEvaluate auth approaches.\n\n## DELIVERABLE\nComparison matrix.")+"---\n\n",
		testSection("SYN-01", "Synthesis", "# SYNTHESIS\n\nCompile findings.\n\n[ALL_SESSION_FINDINGS]"))
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "plan",
			"content":      pipeline,
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}

	// Step 3 & 4: Save T1-01 & T1-02 session output
	t101Output := `---
session_id: T1-01
title: Database Selection
date: 2026-09-27
status: complete
---
# Database Selection

## Recommendation
Use PostgreSQL 16 with pgvector. [Grade A: https://postgresql.org] [fetched]

## Key Findings
PostgreSQL 16 provides the best developer experience and native vector indexing. [Grade A: https://postgresql.org] [fetched]

## Discovered Concerns
Connection pool saturation under burst traffic can exhaust file descriptors in high-concurrency environments.

## Sources & Evidence Ledger
- https://postgresql.org
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save",
		Arguments: map[string]any{"project_root": tmpDir, "kind": "session", "id": "T1-01", "content": t101Output},
	})

	t102Output := `---
session_id: T1-02
title: Auth Strategy
date: 2026-09-27
status: complete
---
# Auth Strategy

## Recommendation
Use Clerk for authentication. [Grade B: https://clerk.com] [fetched]

## Key Findings
Clerk provides comprehensive multi-tenant auth out of the box. [Grade B: https://clerk.com] [fetched]

## Sources & Evidence Ledger
- https://clerk.com
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save",
		Arguments: map[string]any{"project_root": tmpDir, "kind": "session", "id": "T1-02", "content": t102Output},
	})

	// Step 5: Save SYN-01 output (writes to research/FAD.md and mirrors to FOUNDING-ARCHITECTURE.md)
	syn01Output := `---
session_id: SYN-01
title: Founding Architecture Document
synthesis_date: 2026-09-27
status: complete
---
# Founding Architecture Document

## Executive Summary
This document synthesizes findings across sessions T1-01 and T1-02 to establish the project's foundation.
PostgreSQL 16 selected as primary datastore. [Grade A: https://postgresql.org] [fetched]

## Architecture Overview
PostgreSQL 16 selected as primary datastore. [Grade A: https://postgresql.org] [fetched]

## Technology Decisions
1. Database: PostgreSQL 16 with pgvector — [Grade A: https://postgresql.org] [fetched]
2. Auth: Clerk — [Grade B: https://clerk.com] [fetched]

## Locked Decision Registry
- D-001: Primary Datastore (accepted, two-way)
- D-002: Auth Strategy (accepted, two-way)

## Risk Register
1. Connection pool saturation under burst traffic mitigated via PgBouncer.

## Traceability Matrix
- T1-01 -> D-001: PostgreSQL selection
- T1-02 -> D-002: Clerk auth
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save",
		Arguments: map[string]any{"project_root": tmpDir, "kind": "fad", "id": "SYN-01", "content": syn01Output},
	})
	if err != nil {
		t.Fatalf("save_session SYN-01: %v", err)
	}

	// Step 5.5: Record a decision (required by gate Track A check 4)
	decisionContent := `---
decision_id: D-001
title: Primary Datastore
status: accepted
door_type: two-way
review_trigger: "Throughput exceeds 50k ops"
date: 2026-09-27
---
# Decision: Primary Datastore

## Context
We need a database. [Grade A: https://postgresql.org] [fetched]

## Decision
Use PostgreSQL 16. [Grade A: https://postgresql.org] [fetched]

## Evaluated Alternatives
We evaluated MySQL and MongoDB; rejected due to lack of native vector indexing support.
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"id":            "D-001",
			"artifact_type": "decision",
			"content":       decisionContent,
		},
	})

	// Step 6.5: Record a decision to create DECISIONS.md (needed for gate PASS)
	decisionOutput := `---
decision_id: D-002
title: Auth Strategy
status: accepted
door_type: two-way
review_trigger: "Vendor deprecation or cost exceeds $500/mo"
---
# Context
We need auth.
# Decision
Use Clerk.
# Consequences
Tied to Clerk.
`
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "decision",
			"id":            "D-002",
			"content":       decisionOutput,
		},
	})
	if err != nil {
		t.Fatalf("record_decision: %v", err)
	}
	decEnv := parseEnvelope(t, result)
	if !decEnv.Ok {
		t.Fatalf("record_decision failed: %s", decEnv.Message)
	}

	// Step 7: Run gate
	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env := parseEnvelope(t, result)
	if !env.Ok {
		t.Fatalf("run_gate failed: %s", env.Message)
	}

	data, _ := env.Data.(map[string]any)
	if data["verdict"] != "PASS" {
		t.Errorf("expected gate to PASS with FAD, got %v. Issues: %v", data["verdict"], data)
	}
}

func TestSaveSession_FADCaseInsensitive(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	fadContent := `---
id: FAD
title: Founding Architecture Document
synthesis_date: 2026-09-30
status: complete
---
# Founding Architecture
Synthesis of all research findings. Grade A (official documentation).
` + strings.Repeat("Architectural blueprints and component definitions. ", 10)

	// Save session with lowercase "fad"
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "fad",
			"id":           "fad",
			"content":      fadContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session fad: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("expected success saving fad session: %s", env.Message)
	}

	// Verify research/FAD.md was written
	if _, err := os.Stat(filepath.Join(tmpDir, "research", "FAD.md")); err != nil {
		t.Errorf("expected research/FAD.md to exist: %v", err)
	}

	// Verify research/sessions/fad.md was NOT written because session_id is FAD (case-insensitive)
	if _, err := os.Stat(filepath.Join(tmpDir, "research", "sessions", "fad.md")); !os.IsNotExist(err) {
		t.Errorf("expected research/sessions/fad.md to NOT exist, but stat returned %v", err)
	}
}

func TestRecordDecision_ConflictResolution(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Initialize workspace
	initRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init call: %v", err)
	}
	initEnv := parseEnvelope(t, initRes)
	if !initEnv.Ok {
		t.Fatalf("init failed: %s", initEnv.Message)
	}

	// 2. Call vivechak_record_decision with artifact_type = "conflict-resolution"
	crContent := `---
id: CHK-01
decision_id: D-001
title: Conflict Resolution for Storage Engine
status: resolved
door_type: one-way
conflicting_sources:
  - Session T1-01
  - Session T1-02
chosen_option: PostgreSQL
resolution_method: ACH matrix
date: 2026-09-29
schema_version: "0.1.0"
---

# Conflict Resolution: CHK-01 — Storage Engine Divergence

## 1. Conflict Summary
T1-01 recommended PostgreSQL based on ACID compliance. T1-02 recommended MongoDB based on document flexibility.

## 2. Analysis of Competing Hypotheses (ACH Matrix)
Evaluated both options against data integrity, query latency, and operational complexity.
PostgreSQL dominates on ACID transactions and relational joins. Grade A (official docs)

## 3. Resolution
Adopt PostgreSQL with jsonb columns for flexible document attributes.
`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"id":            "CHK-01",
			"artifact_type": "conflict-resolution",
			"content":       crContent,
		},
	})
	if err != nil {
		t.Fatalf("record_decision conflict-resolution: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("expected success, got message: %s, warnings: %v", env.Message, env.Warnings)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected Data map, got %T", env.Data)
	}
	if data["artifact_type"] != "conflict-resolution" {
		t.Errorf("expected artifact_type 'conflict-resolution', got %v", data["artifact_type"])
	}
	if data["status"] != "valid" {
		t.Errorf("expected status 'valid', got %v (warnings: %v)", data["status"], env.Warnings)
	}

	// Verify file was written to disk
	filePath := filepath.Join(tmpDir, "research", "CHK-01-conflict-resolution.md")
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("expected file %s to exist on disk: %v", filePath, err)
	}
	if !strings.Contains(string(contentBytes), "ACH Matrix") {
		t.Errorf("expected saved file to contain ACH Matrix")
	}
}

func TestIDValidationRejectsTraversal(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	invalidIDs := []string{"../../etc/passwd", "../evil"}
	for _, id := range invalidIDs {
		result, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"project_root": tmpDir,
				"kind":         "session",
				"id":           id,
				"content":      "content",
			},
		})
		if err != nil {
			t.Fatalf("save_session with %s: %v", id, err)
		}
		// If the SDK itself rejected it due to schema, it might be an error not JSON.
		// We just ensure it didn't succeed.
		if len(result.Content) > 0 {
			if tc, ok := result.Content[0].(*mcp.TextContent); ok {
				if strings.HasPrefix(tc.Text, "validation error") {
					continue
				}
				env := parseEnvelope(t, result)
				if env.Ok {
					t.Errorf("expected save_session to fail for ID %s", id)
				}
			}
		}
	}

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "T1-01",
			"content":      "---\nsession_id: T1-01\ntitle: test\ndate: 2026-09-27\nstatus: complete\n---\n# test\nA (doc)",
		},
	})
	if err != nil {
		t.Fatalf("save_session with T1-01: %v", err)
	}
	env := parseEnvelope(t, result)
	if !env.Ok {
		t.Errorf("expected save_session to succeed for T1-01, got: %s", env.Message)
	}

	result, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "decision",
			"id":            "../../../hack",
			"content":       "content",
		},
	})
	if err != nil {
		t.Fatalf("record_decision with hack: %v", err)
	}
	if len(result.Content) > 0 {
		if tc, ok := result.Content[0].(*mcp.TextContent); ok {
			if !strings.HasPrefix(tc.Text, "validation error") {
				env = parseEnvelope(t, result)
				if env.Ok {
					t.Errorf("expected record_decision to fail for ../../../hack")
				}
			}
		}
	}
}

func TestContentSizeLimit(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	// 10MB + 1 string
	hugeContent := strings.Repeat("A", 10*1024*1024+1)

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "T1-01",
			"content":      hugeContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session with huge content: %v", err)
	}
	env := parseEnvelope(t, result)
	if env.Ok {
		t.Error("expected save_session to fail for huge content")
	}
}

func TestInit_RejectsFilesystemRoot(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	rootPaths := []string{"/"}
	vol := filepath.VolumeName(t.TempDir())
	if vol != "" {
		rootPaths = append(rootPaths, vol+"\\", vol+"/")
	}

	for _, root := range rootPaths {
		result, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_start",
			Arguments: map[string]any{
				"question":     "test",
				"scope":        "project",
				"project_root": root,
			},
		})
		if err != nil {
			t.Fatalf("call init with %s: %v", root, err)
		}
		env := parseEnvelope(t, result)
		if env.Ok {
			t.Errorf("expected vivechak_start to reject filesystem root %q", root)
		}
		if !strings.Contains(env.Message, "cannot be a filesystem root") && !strings.Contains(env.Message, "cannot be used as workspace root") {
			t.Errorf("expected error message to mention filesystem root, got: %s", env.Message)
		}
	}
}

func TestNextSession_InjectionSizeWarning(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Init workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// 2. Save pipeline with T1-01 and SYN-01
	pipeline := testPlan("Research Pipeline", "PROJECT", "", []string{"{id: T1-01, depends: []}", "{id: T2-01, depends: [T1-01]}",
		"{id: SYN-01, kind: synthesis, depends: [ALL]}"},
		testSection("T1-01", "Session T1-01", "Prompt T1-01"), testSection("T2-01", "Session T2-01", "Investigate based on:\n[UPSTREAM_FINDINGS]"),
		testSection("SYN-01", "Synthesis", "Synthesize."))
	saveRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "plan",
			"content":      pipeline,
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}
	saveEnv := parseEnvelope(t, saveRes)
	if !saveEnv.Ok {
		t.Fatalf("save_plan failed: %s", saveEnv.Message)
	}

	// 3. Save huge T1-01 session (>100KB)
	largeBody := strings.Repeat("Finding: High-performance caching layer delivers 10x throughput. A (benchmark)\n", 2000)
	t101Content := `---
session_id: T1-01
title: Database Selection
date: 2026-09-29
status: complete
---
# Findings
` + largeBody
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "T1-01",
			"content":      t101Content,
		},
	})
	if err != nil {
		t.Fatalf("save_session: %v", err)
	}

	// 4. Request T2-01 via next_session
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "T2-01",
		},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}

	env := parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("next_session failed: %s", env.Message)
	}

	foundWarning := false
	for _, w := range envWarningStrings(env) {
		if strings.Contains(w, "W-INJECTION-SIZE") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Errorf("expected W-INJECTION-SIZE warning in env.Warnings, got: %v", env.Warnings)
	}
}

func TestDecisionScope_EndToEnd(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	// 1. Init with scope decision
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"question":     "test",
			"project_root": tmpDir,
			"scope":        "decision",
		},
	})
	if err != nil {
		t.Fatalf("init decision scope: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("init failed: %s", env.Message)
	}

	// 2. Save Plan with DecisionsContent (ADR)
	planContent := testPlan("Research Plan: D-015 Cache Layer", "D-015", "", []string{"{id: D-015-S1, title: Cache Layer Comparison, role: C, door: two-way, informs: [D-015], depends: []}"},
		testSection("D-015-S1", "Cache Layer Comparison", "# RESEARCH BRIEF: Cache Layer\nCompare Redis vs Dragonfly. A (benchmark)"))
	adrContent := `---
id: D-015
title: Cache Layer Selection
status: proposed
door_type: two-way
date: 2026-09-29
human_reviewed: false
review_trigger: Traffic exceeds 50k QPS
tags: [caching, database]
---
# Context
We need a distributed cache. A (official docs)

# Decision
Pending research.
`
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root":      tmpDir,
			"kind":              "plan",
			"id":                "D-015",
			"slug":              "cache-layer",
			"content":           planContent,
			"decisions_content": adrContent,
		},
	})
	if err != nil {
		t.Fatalf("save_plan: %v", err)
	}
	env = parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("save_plan failed: %s", env.Message)
	}

	// 3. Next session returns D-015-S1
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}
	env = parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("next_session failed: %s", env.Message)
	}
	data, _ := env.Data.(map[string]any)
	if data["session_id"] != "D-015-S1" {
		t.Fatalf("expected session_id D-015-S1, got %v", data["session_id"])
	}

	// 4. Save session output
	sessionContent := `---
session_id: D-015-S1
title: Cache Comparison
status: complete
date: 2026-09-29
---
# Cache Comparison

## Recommendation
Redis chosen for cluster stability. [Grade A: https://redis.io] [fetched]

## Key Findings
Redis provides cluster mode and automatic failover for high availability. [Grade A: https://redis.io] [fetched]

## Sources & Evidence Ledger
- https://redis.io
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "D-015-S1",
			"content":      sessionContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session: %v", err)
	}

	// 4b. Record finalized decision as accepted
	acceptedADR := `---
id: D-015
title: Cache Layer Selection
status: accepted
door_type: two-way
date: 2026-09-29
human_reviewed: false
review_trigger: Traffic exceeds 50k QPS
tags: [caching, database]
---
# Context
We need a distributed cache. A (official docs)

# Decision
Redis chosen for cluster stability. A (official documentation)
`
	resDec, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "decision",
			"id":            "D-015",
			"content":       acceptedADR,
			"revise":        true,
		},
	})
	if err != nil {
		t.Fatalf("record_decision: %v", err)
	}
	envDec := parseEnvelope(t, resDec)
	if !envDec.Ok {
		t.Fatalf("record_decision failed: %s", envDec.Message)
	}

	// 5. Run gate - should pass for decision scope
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"verbose":      true,
		},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env = parseEnvelope(t, res)
	gateData, _ := env.Data.(map[string]any)
	if gateData["verdict"] != "PASS" {
		t.Fatalf("expected gate PASS, got verdict=%v (message=%s)", gateData["verdict"], env.Message)
	}
}

func TestComparisonScope_EndToEnd(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	// 1. Init with scope comparison
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"question":     "test",
			"project_root": tmpDir,
			"scope":        "comparison",
		},
	})
	if err != nil {
		t.Fatalf("init comparison scope: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("init failed: %s", env.Message)
	}

	// 2. Save the comparison (kind="comparison"; kind="plan" never saves a comparison)
	comparisonPrompt := `# RESEARCH BRIEF: Queue Selection
Compare Kafka vs RabbitMQ. A (benchmark)
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "comparison",
			"slug":         "queue-tech",
			"content":      comparisonPrompt,
		},
	})
	if err != nil {
		t.Fatalf("save_plan comparison: %v", err)
	}

	// 3. Next session returns comparison prompt
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}
	env = parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("next_session comparison failed: %s", env.Message)
	}

	// 4. Save session output for comparison
	cmpOutput := `---
session_id: D-020-cmp-queue-tech
title: Queue Comparison
status: complete
date: 2026-09-29
---
# Queue Comparison Matrix
Kafka chosen for stream retention. A (official docs)
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "D-020-cmp-queue-tech",
			"content":      cmpOutput,
		},
	})
	if err != nil {
		t.Fatalf("save_session comparison: %v", err)
	}

	// 5. Run gate for comparison scope
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_gate",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"verbose":      true,
		},
	})
	if err != nil {
		t.Fatalf("run_gate comparison: %v", err)
	}
	env = parseEnvelope(t, res)
	gateData, _ := env.Data.(map[string]any)
	// A comparison-only workspace has no architecture gate: exactly not_applicable, never PASS (CRIT-12). The
	// comparison is saved with kind="comparison" since CRIT-19, so no decision target exists.
	if gateData["verdict"] != "not_applicable" {
		t.Fatalf("expected gate verdict not_applicable for a comparison-only workspace, got verdict=%v (message=%s)", gateData["verdict"], env.Message)
	}
}

func TestNextSession_NotFound_SessionIDList(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	// Init
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	pipeline := oneSessionTestPlan("Session One", "Prompt one")
	// Save pipeline
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "plan",
			"content":      pipeline,
		},
	})

	// Request nonexistent session ID
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"session_id":   "NONEXISTENT-99",
		},
	})
	if err != nil {
		t.Fatalf("unexpected call error: %v", err)
	}
	env := parseEnvelope(t, res)
	if env.Ok {
		t.Errorf("expected error for nonexistent session ID")
	}
	if !strings.Contains(envGuidance(env), "T1-01") {
		t.Errorf("expected available sessions list in NextStep, got: %s", envGuidance(env))
	}
}

func TestValidateArtifact_PlanAndConflict(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)

	// 1. Conflict resolution
	conflict := `---
id: CR-001
decision_id: D-001
title: Database Selection Conflict Resolution
status: complete
door_type: one-way
---
# Conflict Resolution: Database Selection

## Conflict Summary
Session T1-01 recommended PostgreSQL. Session T1-02 recommended ScyllaDB.

## ACH Matrix Analysis
| Hypothesis | Evidence 1 | Evidence 2 | Verdict |
|---|---|---|---|
| PostgreSQL | Consistent | Consistent | Strongly Supported |
| ScyllaDB | Inconsistent | Consistent | Rejected |

## Final Resolution
Adopt PostgreSQL for primary ACID datastore. A (benchmarks)
`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_review",
		Arguments: map[string]any{
			"artifact_type": "conflict-resolution",
			"content":       conflict,
			"mode":          "validate",
		},
	})
	if err != nil {
		t.Fatalf("validate conflict-resolution: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("expected validate success: %s", env.Message)
	}
	data, _ := env.Data.(map[string]any)
	if data["status"] != "valid" && data["status"] != "valid-with-warnings" {
		t.Errorf("expected valid status, got %v", data["status"])
	}

	// 2. Plan artifact
	plan := oneSessionTestPlan("Benchmark", "Execute benchmark")
	resPlan, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_review",
		Arguments: map[string]any{
			"artifact_type": "plan",
			"content":       plan,
			"mode":          "validate",
		},
	})
	if err != nil {
		t.Fatalf("validate plan: %v", err)
	}
	envPlan := parseEnvelope(t, resPlan)
	if !envPlan.Ok {
		t.Fatalf("expected validate plan success: %s", envPlan.Message)
	}
	dataPlan, _ := envPlan.Data.(map[string]any)
	if errCount, ok := dataPlan["error_count"].(float64); !ok || errCount != 0 {
		t.Errorf("expected 0 errors on valid plan, got %v", dataPlan["error_count"])
	}
}

func TestRunGate_Mechanics(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	// 1. Initialize
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	// 2. Save pipeline with 3 sessions (meeting recommended minimum for project scope)
	pipe := threeSessionTestPlan("Core Architecture", "Brief 1", "Storage Layer", "Brief 2", "Synthesis", "Synthesis brief")
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "plan",
			"content":      pipe,
		},
	})

	// 3. Save only session T1-01
	resT1, errT1 := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "T1-01",
			"content": `---
session_id: T1-01
title: Core Architecture
date: 2026-09-29
status: complete
---
# Architecture Findings

## Recommendation
Use SQLite in WAL mode. [Grade A: https://sqlite.org] [fetched]

## Key Findings
Recommended SQLite in WAL mode for lightweight single-node performance. [Grade A: https://sqlite.org] [fetched]

## Sources & Evidence Ledger
- https://sqlite.org
`,
		},
	})
	if errT1 != nil {
		t.Fatalf("save T1-01 failed: %v", errT1)
	}
	envT1 := parseEnvelope(t, resT1)
	if !envT1.Ok {
		t.Fatalf("save T1-01 envelope failed: %s", envT1.Message)
	}

	// 4. Run gate while DAG is incomplete (T1-02 and SYN-01 missing)
	resGate1, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("gate call 1: %v", err)
	}
	envGate1 := parseEnvelope(t, resGate1)
	gate1Data, _ := envGate1.Data.(map[string]any)
	if gate1Data["verdict"] != "FAIL" {
		t.Errorf("gate should fail when DAG sessions are incomplete, got verdict %v", gate1Data["verdict"])
	}

	// 5. Complete T1-02
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "T1-02",
			"content": `---
session_id: T1-02
title: Storage Layer
date: 2026-09-29
status: complete
---
# Storage Findings

## Recommendation
Deploy NVMe direct-attached volume. [Grade B: https://fio.benchmarks.org] [fetched]

## Key Findings
Raw disk benchmarks demonstrate 15k IOPS sustained under load. [Grade B: https://fio.benchmarks.org] [fetched]

## Sources & Evidence Ledger
- https://fio.benchmarks.org
`,
		},
	})

	// 6. Complete SYN-01 and FAD with substantive body (> 100 chars and evidence grades)
	fadContent := `---
id: SYN-01
title: Founding Architecture Document
synthesis_date: 2026-09-29
status: complete
---
# Founding Architecture Document

## Executive Summary
This document synthesizes findings across sessions T1-01 and T1-02 to establish the project's foundation.
The architecture utilizes SQLite with WAL mode backed by high-IOPS storage. [Grade A: https://sqlite.org] [fetched]

## Technology Decisions
1. Datastore: SQLite in WAL mode. [Grade A: https://sqlite.org] [fetched]
2. Storage: NVMe direct-attached volume. [Grade B: https://fio.benchmarks.org] [fetched]

## Locked Decision Registry
- D-001: Database Selection (accepted, one-way)

## Risk Register
1. Concurrent write contention: mitigated via retry loops and short transactions.

## Traceability Matrix
- T1-01 -> D-001: SQLite selection
- T1-02 -> Storage layer NVMe validation
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "SYN-01",
			"content":      fadContent,
		},
	})

	// 7. Record one-way door ADR without reversal trigger
	untriggeredADR := `<!-- DECISION: D-001 -->
---
decision_id: D-001
title: Database Selection
status: accepted
door_type: one-way
---
# Decision
Use SQLite in WAL mode. [Grade A: https://sqlite.org] [fetched]
<!-- /DECISION: D-001 -->
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "decision",
			"id":            "D-001",
			"content":       untriggeredADR,
		},
	})

	// 8. Run gate with one-way door lacking reversal trigger
	resGate2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("gate call 2: %v", err)
	}
	envGate2 := parseEnvelope(t, resGate2)
	gate2Data, _ := envGate2.Data.(map[string]any)
	if gate2Data["verdict"] != "FAIL" {
		t.Errorf("gate should fail when one-way door decision lacks reversal triggers, got verdict %v", gate2Data["verdict"])
	}

	// 9. Add reversal trigger to the one-way door
	triggeredADR := func(challenge string) string {
		return `<!-- DECISION: D-001 -->
---
decision_id: D-001
title: Database Selection
status: accepted
door_type: one-way
review_trigger: "Throughput exceeds 50k ops"
---
# D-001: Database Selection

## Context
Embedded operational storage selection for local client execution.

## Decision
Use SQLite in WAL mode. [Grade A: https://sqlite.org | fetched]
WAL mode lets readers proceed during writes. [Grade B: https://sqlite.org/wal.html | fetched]

## Evaluated Alternatives
We evaluated DuckDB and PostgreSQL; rejected due to client-side embedded deployment requirements.
` + challenge + `<!-- /DECISION: D-001 -->
`
	}
	recordDecisionRevise(t, ctx, cs, tmpDir, "D-001", triggeredADR(""))

	// The red-team record comes from the real tool; the ADR then cites the record id it returns.
	recID := redTeamRecordID(t, ctx, cs, tmpDir, "D-001")
	recordDecisionRevise(t, ctx, cs, tmpDir, "D-001", triggeredADR(fmt.Sprintf(`
## Challenge Findings & Red-Team Stress Test
Evaluated %s adversarial red-team probe confirms single-writer lock contention during bulk transaction processing.
`, recID)))

	// Perform human signoff on D-001
	wsSign, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("open workspace for signoff: %v", err)
	}
	if _, err := core.SignoffDecision(wsSign, "D-001", "Principal Architect"); err != nil {
		wsSign.Close()
		t.Fatalf("signoff D-001: %v", err)
	}
	wsSign.Close()

	// 10. Run gate -> MUST PASS!
	resGate3, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("gate call 3: %v", err)
	}
	envGate3 := parseEnvelope(t, resGate3)
	gate3Data, _ := envGate3.Data.(map[string]any)
	if gate3Data["verdict"] != "PASS" {
		t.Fatalf("expected gate PASS after human signoff, got verdict=%v (message=%s, warnings=%v, rules=%v)",
			gate3Data["verdict"], envGate3.Message, envGate3.Warnings, gate3Data["rules"])
	}
}

func TestRunGate_DecisionScope_ADRValidation(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	// 1. Initialize decision scope
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "decision"},
	})

	// 2. Save decision plan with 1 session
	pipe := `# Decision Pipeline
#### S1-01: Auth Strategy Comparison
| **ID** | S1-01 |
| **Dependencies** | None |
| **Output File** | sessions/S1-01.md |
` + "```prompt" + `
Compare auth options.
` + "```" + `
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "plan",
			"scope":        "decision",
			"id":           "D-001",
			"content":      pipe,
		},
	})

	// 3. Save completed session
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "S1-01",
			"content": `---
session_id: S1-01
title: Auth Strategy Comparison
date: 2026-09-29
status: complete
---
# Auth Findings

## Key Findings
Recommended OAuth2 with PKCE. Grade A (https://auth0.com) [fetched]

## Sources & Evidence Ledger
- https://auth0.com
`},
	})

	// 4. Record decision with status 'proposed' (not yet accepted)
	proposedADR := `---
id: D-001
title: Auth Strategy
status: proposed
door_type: two-way
---
# D-001: Auth Strategy
We propose OAuth2 with PKCE.`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "decision",
			"id":            "D-001",
			"content":       proposedADR,
		},
	})

	// 5. Gate MUST FAIL because status is proposed
	res1, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	env1 := parseEnvelope(t, res1)
	data1, _ := env1.Data.(map[string]any)
	if data1["verdict"] != "FAIL" {
		t.Errorf("gate should fail when decision status is proposed, got verdict %v", data1["verdict"])
	}

	// 6. Record decision as one-way door without review trigger
	untriggeredOneWay := `---
id: D-001
title: Auth Strategy
status: accepted
door_type: one-way
---
# D-001: Auth Strategy
Accepted OAuth2 with PKCE.`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "decision",
			"id":            "D-001",
			"content":       untriggeredOneWay,
			"revise":        true,
		},
	})

	// 7. Gate MUST FAIL because one-way door lacks review trigger
	res2, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	env2 := parseEnvelope(t, res2)
	data2, _ := env2.Data.(map[string]any)
	if data2["verdict"] != "FAIL" {
		t.Errorf("gate should fail when one-way door decision lacks review trigger, got verdict %v", data2["verdict"])
	}

	// 8. Record decision as accepted with review trigger
	validOneWay := func(challenge string) string {
		return `---
id: D-001
title: Auth Strategy
status: accepted
door_type: one-way
review_trigger: 6 months or major breaking RFC change
---
# D-001: Auth Strategy

## Context
We need a robust, secure authentication mechanism for client-facing identity federation across multi-tenant cloud environments.

## Decision Outcome
[Grade A: https://auth0.com | fetched] Accepted OAuth2 with PKCE.
[Grade B: https://oauth.net | fetched] Verified token exchange semantics.

## Rejected Alternatives
We evaluated basic auth and API tokens; rejected due to security considerations.
` + challenge
	}
	recordDecisionRevise(t, ctx, cs, tmpDir, "D-001", validOneWay(""))

	// The red-team record comes from the real tool; the ADR then cites the record id it returns.
	recID := redTeamRecordID(t, ctx, cs, tmpDir, "D-001")
	recordDecisionRevise(t, ctx, cs, tmpDir, "D-001", validOneWay(fmt.Sprintf(`
## Challenge Findings & Red-Team Stress Test
Evaluated %s adversarial red-team probe confirms token leakage vectors mitigated with PKCE code challenges.
`, recID)))

	// Perform human signoff on D-001
	wsSign, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("open workspace for signoff: %v", err)
	}
	if _, err := core.SignoffDecision(wsSign, "D-001", "Security Architect"); err != nil {
		wsSign.Close()
		t.Fatalf("signoff D-001: %v", err)
	}
	wsSign.Close()

	// 9. Gate MUST PASS
	res3, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	env3 := parseEnvelope(t, res3)
	data3, _ := env3.Data.(map[string]any)
	if data3["verdict"] != "PASS" {
		t.Errorf("gate should pass with valid accepted ADR, got verdict=%v (issues: %v)", data3["verdict"], data3)
	}
}

func TestSavePlan_Validation(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	// 1. Pipeline with dependency cycle (T1-01 -> T1-02 -> T1-01)
	cyclePipeline := testPlan("Pipeline", "PROJECT", "", []string{"{id: T1-01, title: Session 1, depends: [T1-02]}",
		"{id: T1-02, title: Session 2, depends: [T1-01]}", "{id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}"},
		testSection("T1-01", "Session 1", "Prompt 1"), testSection("T1-02", "Session 2", "Prompt 2"), testSection("SYN-01", "Synthesis", "Synthesize."))
	res1, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "plan",
			"content":      cyclePipeline,
		},
	})
	env1 := parseEnvelope(t, res1)
	if env1.Ok {
		t.Error("expected save_plan to fail on pipeline with dependency cycle")
	}
	if !strings.Contains(env1.Message, core.CodePlanCycle) || !strings.Contains(env1.Message, "T1-01 -> T1-02 -> T1-01") {
		t.Errorf("expected a %s error naming the cycle, got: %s", core.CodePlanCycle, env1.Message)
	}

	// Verify file was NOT written
	if _, err := os.Stat(filepath.Join(tmpDir, core.PipelineFile)); !os.IsNotExist(err) {
		t.Errorf("pipeline file should not have been created for invalid DAG")
	}

	// 2. Valid pipeline
	validPipeline := oneSessionTestPlan("Session 1", "Prompt 1")
	res2, _ := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "plan",
			"content":      validPipeline,
		},
	})
	env2 := parseEnvelope(t, res2)
	if !env2.Ok {
		t.Fatalf("expected valid pipeline to succeed, got: %s", env2.Message)
	}
	data2, _ := env2.Data.(map[string]any)
	if data2["validation"] == nil {
		t.Error("expected validation result in save_plan response data")
	}
}

func TestSavePlan_MultiDecisionRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	cs := testServer(t)
	ctx := context.Background()

	// 1. Init workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"question":     "test",
			"project_root": tmpDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	pipelineMD := testPlan("Research Pipeline: Multi Decision Test", "PROJECT", "", []string{"{id: T0-01, title: Session 1, depends: []}",
		"{id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}"},
		testSection("T0-01", "Session 1", "Prompt 1"), testSection("SYN-01", "Synthesis", "Synthesize."))

	decisionsMD := `# Architectural Decision Records

<!-- DECISION: D-001 -->
---
id: D-001
title: "Repository Information Architecture"
status: proposed
door_type: one-way
date: 2026-10-20
---

# D-001: Repository Information Architecture
## Initial Context & Hypotheses
Career-stage-first vs role-first.
<!-- /DECISION: D-001 -->

<!-- DECISION: D-002 -->
---
id: D-002
title: "Taxonomy Classification Pipeline"
status: proposed
door_type: two-way
date: 2026-10-20
---

# D-002: Taxonomy Classification Pipeline
## Initial Context & Hypotheses
Zero-shot LLM vs BERT.
<!-- /DECISION: D-002 -->
`

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root":      tmpDir,
			"kind":              "plan",
			"content":           pipelineMD,
			"decisions_content": decisionsMD,
		},
	})
	if err != nil {
		t.Fatalf("save_plan failed: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("expected save_plan to succeed, got: %s", env.Message)
	}

	// Verify that decisions_validation status is valid and no false missing-field warnings are present
	data, _ := env.Data.(map[string]any)
	decVal, ok := data["decisions_validation"].(map[string]any)
	if !ok {
		t.Fatalf("expected decisions_validation in data, got: %v", data["decisions_validation"])
	}
	if decVal["status"] != "valid" {
		t.Errorf("decisions_validation status = %v; want 'valid'", decVal["status"])
	}
	for _, w := range envWarningStrings(env) {
		if strings.Contains(w, "V-MISSING-FIELD") {
			t.Errorf("spurious V-MISSING-FIELD warning found in save_plan: %s", w)
		}
		if strings.Contains(w, "W-NO-EVIDENCE-GRADES") {
			t.Errorf("spurious W-NO-EVIDENCE-GRADES warning found in save_plan: %s", w)
		}
	}
}

func TestInit_HomeDirGuard(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("user home directory not available")
	}
	if info, err := os.Stat(home); err != nil || !info.IsDir() {
		t.Skip("user home directory does not exist on filesystem")
	}

	t.Chdir(home)
	cs := testServer(t)
	ctx := context.Background()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "scope": "project"}, // project_root omitted
	})
	if err != nil {
		t.Fatalf("unexpected call error: %v", err)
	}
	env := parseEnvelope(t, res)
	if env.Ok {
		t.Error("expected error when initializing directly in home directory with omitted project_root")
	}
	if !strings.Contains(env.Message, "user home directory") && !strings.Contains(env.Message, "E-ROOT-UNRESOLVED") {
		t.Errorf("expected home directory or unresolved error message, got: %s", env.Message)
	}

	// project_root: "." must also be blocked (MED-06)
	resDot, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "scope": "project", "project_root": "."},
	})
	if err != nil {
		t.Fatalf("unexpected call error: %v", err)
	}
	envDot := parseEnvelope(t, resDot)
	if envDot.Ok {
		t.Error("expected error when initializing directly in home directory with project_root='.'")
	}
	if !strings.Contains(envDot.Message, "user home directory") {
		t.Errorf("expected home directory error message for project_root='.', got: %s", envDot.Message)
	}
}

func TestInit_SystemOrAppDirGuard(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	testPaths := []string{
		`C:\Users\air\AppData\Local\Programs\Antigravity IDE`,
		`C:\Program Files\SomeApp`,
		`/Applications/Antigravity.app`,
		`/usr/bin`,
	}

	for _, p := range testPaths {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_start",
			Arguments: map[string]any{"question": "test", "scope": "project", "project_root": p},
		})
		if err != nil {
			t.Fatalf("unexpected call error for %s: %v", p, err)
		}
		env := parseEnvelope(t, res)
		if env.Ok {
			t.Errorf("expected error when initializing in system/app dir %q, got success", p)
		}
		if !strings.Contains(env.Message, "application or system directory") {
			t.Errorf("expected application or system directory error, got: %s", env.Message)
		}
	}
}

func TestInit_EnvironmentVariables(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	t.Run("resolves VIVECHAK_PROJECT_ROOT when project_root omitted", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VIVECHAK_PROJECT_ROOT", tmpDir)
		t.Setenv("VIVECHAK_DEFAULT_ROOT", "")

		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_start",
			Arguments: map[string]any{"question": "test", "scope": "project"},
		})
		if err != nil {
			t.Fatalf("call error: %v", err)
		}
		env := parseEnvelope(t, res)
		if !env.Ok {
			t.Fatalf("expected success, got error: %s", env.Message)
		}
		if !core.WorkspaceExists(tmpDir) {
			t.Errorf("expected workspace to be created at %s", tmpDir)
		}
	})

	t.Run("resolves VCK_PROJECT_ROOT when project_root omitted", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VIVECHAK_PROJECT_ROOT", "")
		t.Setenv("VCK_PROJECT_ROOT", tmpDir)
		t.Setenv("VIVECHAK_DEFAULT_ROOT", "")
		t.Setenv("VCK_DEFAULT_ROOT", "")

		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_start",
			Arguments: map[string]any{"question": "test", "scope": "project"},
		})
		if err != nil {
			t.Fatalf("call error: %v", err)
		}
		env := parseEnvelope(t, res)
		if !env.Ok {
			t.Fatalf("expected success, got error: %s", env.Message)
		}
		if !core.WorkspaceExists(tmpDir) {
			t.Errorf("expected workspace to be created at %s", tmpDir)
		}
	})

	t.Run("ignores unsupported VCK_DEFAULT_ROOT", func(t *testing.T) {
		isolatedDir := t.TempDir()
		t.Chdir(isolatedDir)
		tmpDir := t.TempDir()
		t.Setenv("VIVECHAK_PROJECT_ROOT", "")
		t.Setenv("VCK_PROJECT_ROOT", "")
		t.Setenv("VIVECHAK_DEFAULT_ROOT", "")
		t.Setenv("VCK_DEFAULT_ROOT", tmpDir)

		_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_start",
			Arguments: map[string]any{"question": "test", "scope": "project"},
		})
		if core.WorkspaceExists(tmpDir) {
			t.Errorf("unsupported VCK_DEFAULT_ROOT should not create workspace at %s", tmpDir)
		}
	})

	t.Run("ignores unnamespaced WORKSPACE", func(t *testing.T) {
		isolatedDir := t.TempDir()
		t.Chdir(isolatedDir)
		tmpDir := t.TempDir()
		t.Setenv("VIVECHAK_PROJECT_ROOT", "")
		t.Setenv("VCK_PROJECT_ROOT", "")
		t.Setenv("WORKSPACE", tmpDir)

		_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_start",
			Arguments: map[string]any{"question": "test", "scope": "project"},
		})
		if core.WorkspaceExists(tmpDir) {
			t.Errorf("unnamespaced WORKSPACE should not create workspace at %s", tmpDir)
		}
	})

	t.Run("ignores unnamespaced PROJECT_ROOT", func(t *testing.T) {
		isolatedDir := t.TempDir()
		t.Chdir(isolatedDir)
		tmpDir := t.TempDir()
		t.Setenv("VIVECHAK_PROJECT_ROOT", "")
		t.Setenv("VCK_PROJECT_ROOT", "")
		t.Setenv("PROJECT_ROOT", tmpDir)

		_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_start",
			Arguments: map[string]any{"question": "test", "scope": "project"},
		})
		if core.WorkspaceExists(tmpDir) {
			t.Errorf("unnamespaced PROJECT_ROOT should not create workspace at %s", tmpDir)
		}
	})

	t.Run("allows custom sandbox project subfolder path", func(t *testing.T) {
		sandboxDir := filepath.Join(t.TempDir(), "testbed-sandbox", "my-app")
		if err := os.MkdirAll(sandboxDir, 0o755); err != nil {
			t.Fatalf("mkdir sandbox: %v", err)
		}
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      "vivechak_start",
			Arguments: map[string]any{"question": "test", "scope": "project", "project_root": sandboxDir},
		})
		if err != nil {
			t.Fatalf("call error: %v", err)
		}
		env := parseEnvelope(t, res)
		if !env.Ok {
			t.Fatalf("expected success for sandbox path %s, got: %s", sandboxDir, env.Message)
		}
		if !core.WorkspaceExists(sandboxDir) {
			t.Errorf("expected workspace created at %s", sandboxDir)
		}
	})
}

func TestStatus_TerminalSynthesisGuidance(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Create sessions and FAD.md
	sessPath := filepath.Join(tmpDir, "research", "sessions", "T1-01.md")
	_ = os.WriteFile(sessPath, []byte("# Session\nA (doc)"), 0o644)
	fadPath := filepath.Join(tmpDir, "research", "FAD.md")
	_ = os.WriteFile(fadPath, []byte("# FAD\nSynthesis complete."), 0o644)

	// Call status
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_status",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("status call: %v", err)
	}
	env := parseEnvelope(t, res)
	if !strings.Contains(envGuidance(env), "vivechak_gate") {
		t.Errorf("expected status next_step to direct to vivechak_gate, got: %s", envGuidance(env))
	}
}

func TestRunGate_EmptyDecisionsRegistryClarifiedMessage(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Create pipeline with empty decisions
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "RESEARCH-PIPELINE.md"), []byte("# Pipeline\n#### T1-01: Foundation\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "DECISIONS.md"), []byte("# Architectural Decision Log\n\n| Decision | Title | Door Type | Status | Date |\n|---|---|---|---|---|\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "sessions", "T1-01.md"), []byte("# Findings\nA (docs)"), 0o644)
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "FAD.md"), []byte("# FAD\nFindings... A (docs)"), 0o644)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env := parseEnvelope(t, res)
	foundClarified := false
	for _, w := range envWarningStrings(env) {
		if strings.Contains(w, "No architectural decisions recorded yet — record decisions using vivechak_record_decision before running the gate") {
			foundClarified = true
			break
		}
	}
	if !foundClarified {
		t.Errorf("expected clarified empty decisions warning, got warnings: %v", env.Warnings)
	}
}

func TestRunGate_EnvelopeKeyParity(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env := parseEnvelope(t, res)
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got: %T", env.Data)
	}

	requiredKeys := []string{
		"structural_checks",
		"rules",
		"track_a",
		"track_filter",
		"scope_note",
		"verdict",
	}
	for _, key := range requiredKeys {
		if data[key] == nil {
			t.Errorf("expected gate data key %q to be present", key)
		}
	}

	// Verify no duplicate track keys and no copies of the verdict (HIGH-17)
	forbiddenKeys := []string{
		"quality_checks",
		"structural_completeness",
		"mechanical_quality",
		"track_b",
		"track_b_checks",
		"gate_passed",
		"gate_status",
		"track_a_result",
		"track_b_result",
	}
	for _, key := range forbiddenKeys {
		if data[key] != nil {
			t.Errorf("duplicate gate data key %q must not be present (HIGH-17)", key)
		}
	}
}

func TestRunGate_IgnoresConflictResolutionArtifacts(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Init decision scope
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "decision"},
	})

	// Add a valid session
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "sessions", "S1.md"), []byte("# S1\nA (doc)"), 0o644)

	// Add an accepted ADR
	adr := `---
decision_id: D-001
title: DB Selection
status: accepted
door_type: two-way
---
# Decision
Use SQLite. A (doc)`
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "D-001-db.md"), []byte(adr), 0o644)

	// Add a conflict resolution record in research/
	cr := `---
id: CHK-01
decision_id: D-001
title: Conflict Resolution
status: resolved
door_type: two-way
conflicting_sources: [S1]
---
# Conflict Resolution
Resolved in favor of SQLite. A (doc)`
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "CHK-01-conflict-resolution.md"), []byte(cr), 0o644)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env := parseEnvelope(t, res)
	// Must not complain that CHK-01 is missing decision_id or is an invalid ADR
	for _, w := range envWarningStrings(env) {
		if strings.Contains(w, "CHK-01") {
			t.Errorf("unexpected gate warning for conflict resolution: %s", w)
		}
	}
}

func TestServerInstructions(t *testing.T) {
	instrLen := len(ServerInstructions)
	if instrLen > 600 {
		t.Errorf("expected ServerInstructions length <= 600 bytes per P4.4, got %d", instrLen)
	}
	requiredSubstrings := []string{
		"research/",
		"A-E",
		"vivechak_start",
	}
	for _, sub := range requiredSubstrings {
		if !strings.Contains(ServerInstructions, sub) {
			t.Errorf("ServerInstructions missing required substring %q", sub)
		}
	}
}

func TestToolDescriptions(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	toolsToTest := map[string]string{
		"vivechak_start":  "brief",
		"vivechak_save":   "persist",
		"vivechak_gate":   "exit criteria",
		"vivechak_review": "validate",
	}

	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		expectedPhrase, ok := toolsToTest[tool.Name]
		if !ok {
			continue
		}
		if len(tool.Description) < 100 {
			t.Errorf("tool %q description too short: %d bytes", tool.Name, len(tool.Description))
		}
		if !strings.Contains(tool.Description, expectedPhrase) {
			t.Errorf("tool %q description missing phrase %q", tool.Name, expectedPhrase)
		}
	}
}

func TestSaveSession_QualityObservationsAndProgress(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	sessContent := `---
id: T0-01
title: Initial Research
date: 2026-10-14
status: complete
topic: architecture
informs_decisions: [D-001]
confidence: high
---
# Prior
Initial belief here.

# Research Question
What db?

# Key Findings
- Finding 1 [Grade A · direct | fetched https://example.com]
- Finding 2 [Grade B · direct | cached]
- Finding 3 [Grade B · indirect | cached]

# Recommendation
Use Postgres.

# Alternatives Considered
MySQL was considered.

# Open Questions & Risks
None.

# Sources & Evidence Ledger
| Claim | Grade | Modifiers | Method | Source |
|---|---|---|---|---|
| Finding 1 | Grade A | direct | fetched | https://example.com |
`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "T0-01",
			"content":      sessContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("save_session expected success, got error: %s", env.Message)
	}

	dataMap, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected env.Data to be map[string]any, got %T", env.Data)
	}

	obsRaw, ok := dataMap["quality_observations"]
	if !ok {
		t.Fatalf("expected quality_observations in save_session response data")
	}
	obsSlice, ok := obsRaw.([]any)
	if !ok || len(obsSlice) == 0 {
		t.Fatalf("expected non-empty quality_observations slice, got %v", obsRaw)
	}

	progressRaw, ok := dataMap["progress"]
	if !ok {
		t.Fatalf("expected progress in save_session response data")
	}
	progMap, ok := progressRaw.(map[string]any)
	if !ok {
		t.Fatalf("expected progress to be map[string]any, got %T", progressRaw)
	}
	if progMap["sessions_completed"] != float64(1) && progMap["sessions_completed"] != 1 {
		t.Errorf("expected sessions_completed=1, got %v", progMap["sessions_completed"])
	}
}

func TestRecordDecision_ProgressAndDoorTypeAwareNextStep(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	adr := `---
id: D-001
title: Primary Datastore
status: accepted
door_type: one-way
review_trigger: "latency > 100ms"
informed_by_sessions: [T0-01]
confidence: high
---
# Context
Storage requirement.
# Evaluated Options
1. Postgres
# Decision Outcome
Chosen Option 1.
# Rejected Alternatives & Tradeoffs
MySQL rejected due to JSON querying needs.
`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"id":           "D-001",
			"content":      adr,
		},
	})
	if err != nil {
		t.Fatalf("record_decision: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("record_decision expected success, got: %s", env.Message)
	}

	recDataMap, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected env.Data to be map[string]any, got %T", env.Data)
	}
	progRaw, ok := recDataMap["progress"]
	if !ok {
		t.Fatalf("expected progress in record_decision data")
	}
	progMap, ok := progRaw.(map[string]any)
	if !ok {
		t.Fatalf("expected progress to be map, got %T", progRaw)
	}
	if progMap["decisions_recorded"] != float64(1) && progMap["decisions_recorded"] != 1 {
		t.Errorf("expected decisions_recorded=1, got %v", progMap["decisions_recorded"])
	}
}

func TestGate_EvidentiaryIntegrity_B3_B4_B5(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	pipelineContent := testPlan("Research Pipeline", "PROJECT", "", []string{"{id: T0-01, title: Session 1, door: one-way, depends: []}",
		"{id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}"},
		testSection("T0-01", "Session 1", "Research Session 1."), testSection("SYN-01", "Synthesis", "Synthesize."))
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "RESEARCH-PIPELINE.md"), []byte(pipelineContent), 0o644)

	sessContent := `---
id: T0-01
title: Session 1
status: complete
informs_decisions: [D-001]
---
# Key Findings
- Finding 1 [Grade C · direct | recalled memory]
# Sources & Evidence Ledger
| Claim | Grade | Modifiers | Method | Source |
|---|---|---|---|---|
| Finding 1 | Grade C | direct | recalled | memory |
`
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "sessions", "T0-01.md"), []byte(sessContent), 0o644)

	decContent := `---
id: D-001
title: Decision 1
status: accepted
door_type: one-way
review_trigger: "latency > 100ms"
informed_by_sessions: [T0-01]
---
# Context
Context here.
# Decision Outcome
We decided to use Postgres.
`
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "D-001-db.md"), []byte(decContent), 0o644)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env := parseEnvelope(t, res)

	hasB3 := false
	hasB4 := false
	hasB5 := false
	for _, w := range envWarningStrings(env) {
		if strings.Contains(w, "B3-EVIDENTIARY") {
			hasB3 = true
		}
		if strings.Contains(w, "B4-VERIFICATION") {
			hasB4 = true
		}
		if strings.Contains(w, "B5-ALTERNATIVES") {
			hasB5 = true
		}
	}
	if !hasB3 {
		t.Errorf("expected B3-EVIDENTIARY advisory for one-way door with Grade C, got warnings: %v", env.Warnings)
	}
	if !hasB4 {
		t.Errorf("expected B4-VERIFICATION advisory for one-way door with recalled method, got warnings: %v", env.Warnings)
	}
	if !hasB5 {
		t.Errorf("expected B5-ALTERNATIVES advisory for one-way door lacking alternatives, got warnings: %v", env.Warnings)
	}
}

func TestGate_EvidentiaryIntegrity_RAMNotRecalled(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	pipelineContent := testPlan("Research Pipeline", "PROJECT", "", []string{"{id: T0-01, title: In-Memory Datastore Evaluation, depends: []}",
		"{id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}"},
		testSection("T0-01", "In-Memory Datastore Evaluation", "Research In-Memory Datastore Evaluation."), testSection("SYN-01", "Synthesis", "Synthesize."))
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "RESEARCH-PIPELINE.md"), []byte(pipelineContent), 0o644)

	// Session text discussing RAM, heap memory, in-memory caching - NOT unverified recall citations!
	sessContent := `---
id: T0-01
title: In-Memory Datastore Evaluation
status: complete
informs_decisions: [D-001]
---
# Key Findings
- Redis keeps data in system memory (RAM). Grade A (primary docs: https://redis.io)
- Memory limits were evaluated under load. Grade B (benchmark)
- Heap allocation remained under 200MB during spikes. Grade A (direct telemetry)
# Sources & Evidence Ledger
| Claim | Grade | Modifiers | Method | Source |
|---|---|---|---|---|
| RAM usage | Grade A | direct | fetched | https://redis.io |
| Heap memory | Grade A | fresh | fetched | https://redis.io/docs |
`
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "sessions", "T0-01.md"), []byte(sessContent), 0o644)

	decContent := `---
id: D-001
title: Cache Store
status: accepted
door_type: one-way
review_trigger: "memory usage > 512MB"
informed_by_sessions: [T0-01]
---
# Context
In-memory caching is needed.
# Decision Outcome
We decided on Redis.
# Rejected Alternatives
Memcached was rejected due to lack of persistence data structures.
`
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "D-001.md"), []byte(decContent), 0o644)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env := parseEnvelope(t, res)

	// Verify that B4-VERIFICATION does NOT fire for system memory/RAM discussions!
	for _, w := range envWarningStrings(env) {
		if strings.Contains(w, "B4-VERIFICATION") {
			t.Errorf("unexpected B4-VERIFICATION advisory triggered by technical RAM/memory keywords: %s", w)
		}
	}
}

func TestGate_EvidentiaryIntegrity_FallbackToDecisionsMD(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	pipelineContent := testPlan("Research Pipeline", "PROJECT", "", []string{"{id: T0-01, title: Session 1, depends: []}",
		"{id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}"},
		testSection("T0-01", "Session 1", "Research Session 1."), testSection("SYN-01", "Synthesis", "Synthesize."))
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "RESEARCH-PIPELINE.md"), []byte(pipelineContent), 0o644)

	sessContent := `---
id: T0-01
title: Session 1
status: complete
informs_decisions: [D-001]
---
# Key Findings
- Finding 1 [Grade C · direct | recalled memory]
# Sources & Evidence Ledger
| Claim | Grade | Modifiers | Method | Source |
|---|---|---|---|---|
| Finding 1 | Grade C | direct | recalled | memory |
`
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "sessions", "T0-01.md"), []byte(sessContent), 0o644)

	// Note: DO NOT create research/D-001.md! Only place D-001 inside DECISIONS.md
	decisionsContent := `# Architecture Decisions

<!-- DECISION: D-001 -->
---
id: D-001
title: Datastore Decision
status: accepted
door_type: one-way
review_trigger: "latency > 100ms"
informed_by_sessions: [T0-01]
---
# Context
Context here.
# Decision Outcome
Decision outcome.
<!-- /DECISION: D-001 -->
`
	_ = os.WriteFile(filepath.Join(tmpDir, "research", "DECISIONS.md"), []byte(decisionsContent), 0o644)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	env := parseEnvelope(t, res)

	// Fallback to DECISIONS.md should detect D-001, find T0-01 has Grade C & recalled, and missing rejected alternatives
	hasB3 := false
	hasB4 := false
	hasB5 := false
	for _, w := range envWarningStrings(env) {
		if strings.Contains(w, "B3-EVIDENTIARY") {
			hasB3 = true
		}
		if strings.Contains(w, "B4-VERIFICATION") {
			hasB4 = true
		}
		if strings.Contains(w, "B5-ALTERNATIVES") {
			hasB5 = true
		}
	}
	if !hasB3 {
		t.Errorf("expected B3-EVIDENTIARY via DECISIONS.md fallback, got warnings: %v", env.Warnings)
	}
	if !hasB4 {
		t.Errorf("expected B4-VERIFICATION via DECISIONS.md fallback, got warnings: %v", env.Warnings)
	}
	if !hasB5 {
		t.Errorf("expected B5-ALTERNATIVES via DECISIONS.md fallback, got warnings: %v", env.Warnings)
	}
}

func TestValidate_PlanDoesNotTriggerSessionObservations(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	planContent := testPlan("Research Plan: Distributed Tracing", "PROJECT", "## Strategy\nInvestigate OpenTelemetry vs OpenTracing.",
		[]string{"{id: T0-01, title: Tracing protocol, depends: []}", "{id: T0-02, title: Collector performance, depends: []}",
			"{id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}"},
		testSection("T0-01", "Tracing protocol", "Compare OpenTelemetry and OpenTracing protocols."),
		testSection("T0-02", "Collector performance", "Measure collector performance."), testSection("SYN-01", "Synthesis", "Synthesize."))
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_review",
		Arguments: map[string]any{
			"project_root":  tmpDir,
			"artifact_type": "plan",
			"content":       planContent,
			"mode":          "validate",
		},
	})
	if err != nil {
		t.Fatalf("validate plan: %v", err)
	}
	env := parseEnvelope(t, res)
	if !env.Ok {
		t.Fatalf("expected validation success for plan, got: %s", env.Message)
	}

	// Verify that dry-running a plan does NOT emit session quality observations (Prior / Delta)
	for _, w := range envWarningStrings(env) {
		if strings.Contains(w, "No Prior section") || strings.Contains(w, "No Delta section") {
			t.Errorf("unexpected session observation leaked into plan validation: %s", w)
		}
	}
}

func TestNextSession_CaseInsensitive(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_start",
		Arguments: map[string]any{"question": "test", "project_root": tmpDir, "scope": "project"},
	})

	pipelineContent := testPlan("Research Pipeline", "PROJECT", "", []string{"{id: T0-01, title: Session 1, depends: []}",
		"{id: T0-02, title: Session 2, depends: [T0-01]}", "{id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}"},
		testSection("T0-01", "Session 1", "# Brief"), testSection("T0-02", "Session 2", "# Brief"), testSection("SYN-01", "Synthesis", "Synthesize."))

	_ = os.WriteFile(filepath.Join(tmpDir, "research", "RESEARCH-PIPELINE.md"), []byte(pipelineContent), 0o644)

	// Save session 1 with lowercase filename / id "t0-01"
	sessContent := `---
id: t0-01
title: Session 1
date: 2026-10-14
status: complete
---
# Key Findings
- Finding 1 [Grade A]
## Recommendation
Recommended Option A.
`
	saveRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"kind":         "session",
			"id":           "t0-01",
			"content":      sessContent,
		},
	})
	if err != nil {
		t.Fatalf("save_session: %v", err)
	}
	saveEnv := parseEnvelope(t, saveRes)
	if !saveEnv.Ok {
		t.Fatalf("save_session failed: %s", saveEnv.Message)
	}

	// Now ask for next_session: T0-02 should be ready because t0-01 completed!
	nextRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next",
		Arguments: map[string]any{
			"project_root": tmpDir,
		},
	})
	if err != nil {
		t.Fatalf("next_session: %v", err)
	}
	nextEnv := parseEnvelope(t, nextRes)
	if !nextEnv.Ok {
		t.Fatalf("expected next_session success, got: %s", nextEnv.Message)
	}
	dataMap, ok := nextEnv.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected nextEnv.Data map, got %T", nextEnv.Data)
	}
	if dataMap["session_id"] != "T0-02" {
		t.Errorf("expected next ready session to be T0-02, got %v", dataMap["session_id"])
	}
}

func TestRunGate_OneWayDoor_MissingAlternatives_FailsTrackB(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	// Initialize decision-scope workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"question":     "test",
			"project_root": tmpDir,
			"scope":        "decision",
		},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Record a One-Way Door ADR that deliberately lacks an alternatives section
	incompleteADR := `---
id: D-001
title: Core Architecture
status: accepted
door_type: one-way
review_trigger: 6 months
human_reviewed: false
---
# D-001: Core Architecture
## Context
We are deciding the core database layer.
## Decision
We select PostgreSQL. Grade A (RFC docs)
`
	_, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"id":           "D-001",
			"content":      incompleteADR,
		},
	})
	if err != nil {
		t.Fatalf("record_decision: %v", err)
	}

	// Run gate: must NOT be PASS! Must fail Track B due to missing B5 alternatives.
	resGate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	envGate := parseEnvelope(t, resGate)
	gateData := envGate.Data.(map[string]any)
	if gateData["verdict"] != "FAIL" {
		t.Fatalf("expected gate to FAIL on One-Way Door with missing alternatives, got verdict %v", gateData["verdict"])
	}

	if !strings.Contains(envGate.Message, "Phase 0 Gate: FAIL") {
		t.Errorf("expected failure message 'Phase 0 Gate: FAIL', got: %s", envGate.Message)
	}
	if !strings.Contains(envGuidance(envGate), "DO NOT START IMPLEMENTATION") {
		t.Errorf("expected failure next.why stop-signal 'DO NOT START IMPLEMENTATION', got: %s", envGuidance(envGate))
	}

	foundB5 := false
	for _, w := range envWarningStrings(envGate) {
		if strings.Contains(w, "B5-ALTERNATIVES") {
			foundB5 = true
			break
		}
	}
	if !foundB5 {
		t.Errorf("expected B5-ALTERNATIVES in gate warnings, got: %v", envGate.Warnings)
	}

	foundChallengeWarn := false
	for _, w := range envWarningStrings(envGate) {
		if strings.Contains(w, "W-OWD-NO-CHALLENGE") {
			foundChallengeWarn = true
			break
		}
	}
	if !foundChallengeWarn {
		t.Errorf("expected W-OWD-NO-CHALLENGE in gate warnings for unchallenged One-Way Door, got: %v", envGate.Warnings)
	}
}

func TestRunGate_ChallengeWarning_Resolved(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)
	tmpDir := t.TempDir()

	// Initialize decision-scope workspace
	_, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"question":     "test",
			"project_root": tmpDir,
			"scope":        "decision",
		},
	})
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Record a One-Way Door ADR, run the real red-team review, then cite the record it returns
	adrWithChallenge := func(challenge string) string {
		return `---
id: D-001
title: Core Storage Engine
status: accepted
door_type: one-way
review_trigger: 6 months
human_reviewed: false
---
# D-001: Core Storage Engine
## Context
We are choosing an embedded transactional datastore.
## Evaluated Options
- SQLite
- DuckDB
## Decision Outcome
We select SQLite. Grade A (https://sqlite.org)
## Rejected Alternatives & Tradeoffs
- DuckDB: Rejected due to OLTP transaction locks.
` + challenge + `## Failure Modes & Reversal Triggers
Review if write transactions exceed 500 ops/sec.
`
	}
	recordDecisionRevise(t, ctx, cs, tmpDir, "D-001", adrWithChallenge(""))
	recID := redTeamRecordID(t, ctx, cs, tmpDir, "D-001")
	recordDecisionRevise(t, ctx, cs, tmpDir, "D-001", adrWithChallenge(fmt.Sprintf(`## Challenge Findings & Red-Team Stress Test
Evaluated %s adversarial probe confirms concurrency limit at 1 writer per database file.
`, recID)))

	// Run gate: W-OWD-NO-CHALLENGE should NOT be emitted, and G10 passes on the tool's own, cited record
	resGate, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_gate",
		Arguments: map[string]any{"project_root": tmpDir, "verbose": true},
	})
	if err != nil {
		t.Fatalf("run_gate: %v", err)
	}
	envGate := parseEnvelope(t, resGate)
	for _, w := range envWarningStrings(envGate) {
		if strings.Contains(w, "W-OWD-NO-CHALLENGE") {
			t.Errorf("unexpected W-OWD-NO-CHALLENGE warning when Challenge Findings section is present: %s", w)
		}
	}
	gateData, _ := envGate.Data.(map[string]any)
	rulesList, _ := gateData["rules"].([]any)
	g10 := ""
	for _, r := range rulesList {
		if rm, ok := r.(map[string]any); ok && rm["id"] == "G10" {
			g10, _ = rm["status"].(string)
		}
	}
	if g10 != "PASS" {
		t.Errorf("expected G10 PASS with the red-team record %s cited, got %q (rules: %v)", recID, g10, rulesList)
	}
}

func TestValidate_DecisionRegistryWithPreamble(t *testing.T) {
	ctx := context.Background()
	cs := testServer(t)

	decisionsContent := `# Architectural Decision Records

This log documents all technical choices for the platform.

| Decision | Title | Door Type | Status | Date |
|---|---|---|---|---|
| D-01 | Primary Database | one-way | proposed | 2026-10-21 |

<!-- DECISION: D-01 -->
<details>
<summary><b>D-01: Primary Database</b></summary>

---
id: D-01
title: Primary Database
status: proposed
door_type: one-way
date: 2026-10-21
---
Initial proposed ADR skeleton for D-01.
</details>
<!-- /DECISION: D-01 -->
`

	// 1. Validate with artifact_type="decision"
	res1, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_review",
		Arguments: map[string]any{
			"artifact_type": "decision",
			"content":       decisionsContent,
			"mode":          "validate",
		},
	})
	if err != nil {
		t.Fatalf("call validate: %v", err)
	}
	env1 := parseEnvelope(t, res1)
	if !env1.Ok {
		t.Fatalf("expected success, got: %s", env1.Message)
	}
	data1, _ := env1.Data.(map[string]any)
	if data1["status"] != "valid" {
		t.Errorf("expected valid status, got %v (warnings: %v)", data1["status"], env1.Warnings)
	}

	// 2. Validate with artifact_type="decisions"
	res2, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_review",
		Arguments: map[string]any{
			"artifact_type": "decision",
			"content":       decisionsContent,
			"mode":          "validate",
		},
	})
	if err != nil {
		t.Fatalf("call validate: %v", err)
	}
	env2 := parseEnvelope(t, res2)
	if !env2.Ok {
		t.Fatalf("expected success, got: %s", env2.Message)
	}
	data2, _ := env2.Data.(map[string]any)
	if data2["status"] != "valid" {
		t.Errorf("expected valid status for artifact_type=decisions, got %v (warnings: %v)", data2["status"], env2.Warnings)
	}
}

// TestMCPProtocolTriumvirate_ListingAndExecution verifies that the server exposes
// the complete MCP Protocol Triumvirate: exactly 9 Tools, 4 Prompts, and 8 Resources,
// and that they can be listed and invoked across a single live MCP client session.
func TestMCPProtocolTriumvirate_ListingAndExecution(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	// 1. Verify 9 Tools
	toolsRes, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
	if len(toolsRes.Tools) != 9 {
		t.Errorf("expected exactly 9 tools, got %d", len(toolsRes.Tools))
	}

	// 2. Verify 4 Prompts
	promptsRes, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("ListPrompts failed: %v", err)
	}
	if len(promptsRes.Prompts) != 4 {
		t.Errorf("expected exactly 4 prompts, got %d", len(promptsRes.Prompts))
	}

	// 3. Verify 8 Resources
	resourcesRes, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}
	if len(resourcesRes.Resources) != 8 {
		t.Errorf("expected exactly 8 resources, got %d", len(resourcesRes.Resources))
	}

	// 4. Sequential execution of one primitive from each triumvirate pillar
	// Pillar A: Execute a prompt
	pRes, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "compare",
		Arguments: map[string]string{
			"options": "Option A vs Option B",
		},
	})
	if err != nil {
		t.Fatalf("GetPrompt failed: %v", err)
	}
	if len(pRes.Messages) == 0 {
		t.Fatal("expected prompt messages, got none")
	}

	// Pillar B: Read a canonical resource
	rRes, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "vivechak://methodology/grades",
	})
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}
	if len(rRes.Contents) == 0 || !strings.Contains(rRes.Contents[0].Text, "Grade A") {
		t.Errorf("unexpected resource content: %v", rRes.Contents)
	}

	// Pillar C: Invoke an orchestration tool
	tmpDir := t.TempDir()
	tRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_start",
		Arguments: map[string]any{
			"project_root": tmpDir,
			"scope":        "comparison",
			"question":     "Option A vs Option B",
		},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	env := parseEnvelope(t, tRes)
	if !env.Ok {
		t.Errorf("vivechak_start tool invocation failed: %s", env.Message)
	}
}
