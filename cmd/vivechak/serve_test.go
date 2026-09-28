package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type testEnvelope struct {
	Success  bool           `json:"success"`
	Message  string         `json:"message"`
	Data     map[string]any `json:"data,omitempty"`
	Warnings []string       `json:"warnings,omitempty"`
	NextStep string         `json:"next_step,omitempty"`
	Meta     map[string]any `json:"meta,omitempty"`
}

func parseTestEnvelope(t *testing.T, res *mcp.CallToolResult) testEnvelope {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatal("result has no content")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", res.Content[0])
	}
	var env testEnvelope
	if err := json.Unmarshal([]byte(tc.Text), &env); err != nil {
		t.Fatalf("parsing envelope: %v; raw: %s", err, tc.Text)
	}
	return env
}

// TestLiveMCPServer_EndToEnd executes a full end-to-end integration test
// by compiling the actual vivechak binary and running it as an external subprocess
// communicating over standard MCP JSON-RPC protocol via stdio.
func TestLiveMCPServer_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live MCP subprocess test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)

	// 1. Build vivechak binary in a temporary directory
	binDir := t.TempDir()
	binName := "vivechak_test_bin"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build vivechak binary for test: %v\noutput: %s", err, string(out))
	}

	// 2. Setup temporary workspace
	workDir := t.TempDir()

	// 3. Launch subprocess using CommandTransport
	cmd := exec.CommandContext(ctx, binPath, "serve")
	cmd.Stderr = os.Stderr // preserve log visibility on test failure

	transport := &mcp.CommandTransport{
		Command: cmd,
	}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "live-e2e-test", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server over stdio: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	// 4. Introspect registered tools
	var toolNames []string
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		toolNames = append(toolNames, tool.Name)
	}
	if len(toolNames) != 9 {
		t.Fatalf("expected 9 tools, got %d: %v", len(toolNames), toolNames)
	}

	// 5. Tool: vivechak_init
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": workDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("vivechak_init call: %v", err)
	}
	env := parseTestEnvelope(t, res)
	if !env.Success {
		t.Fatalf("vivechak_init failed: %s", env.Message)
	}
	metaFile := filepath.Join(workDir, "research", ".vivechak.json")
	if _, err := os.Stat(metaFile); err != nil {
		t.Fatalf(".vivechak.json was not created on disk: %v", err)
	}

	// 6. Tool: vivechak_prepare_generator
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_prepare_generator",
		Arguments: map[string]any{
			"project_root": workDir,
			"scope":        "project",
			"context":      "Building high-throughput ledger system.",
		},
	})
	if err != nil {
		t.Fatalf("vivechak_prepare_generator call: %v", err)
	}
	env = parseTestEnvelope(t, res)
	if !env.Success {
		t.Fatalf("vivechak_prepare_generator failed: %s", env.Message)
	}
	prompt, _ := env.Data["prompt"].(string)
	if !strings.Contains(prompt, "high-throughput ledger system") {
		t.Error("prompt should contain injected context")
	}

	// 7. Tool: vivechak_save_plan
	pipeline := `# Research Pipeline

**Complexity Score**: 12/25 → Medium (Tier 2)

## Session DAG

### Session T1-01 — Database Selection
- **ID**: T1-01
- **Layer**: 1
- **Door Type**: one-way
- **Dependencies**: None
- **Output File**: ` + "`T1-01-database.md`" + `

` + "```prompt" + `
Investigate database engines.
` + "```" + `

### Session T1-02 — Consensus
- **ID**: T1-02
- **Layer**: 1
- **Door Type**: one-way
- **Dependencies**: None
- **Output File**: ` + "`T1-02-consensus.md`" + `

` + "```prompt" + `
Investigate consensus protocols.
` + "```" + `

### Session SYN-01 — Synthesis
- **ID**: SYN-01
- **Layer**: 2
- **Door Type**: one-way
- **Dependencies**: T1-01, T1-02
- **Output File**: ` + "`FAD.md`" + `

` + "```prompt" + `
Synthesize findings:
[ALL_SESSION_FINDINGS]
` + "```" + `
`
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_plan",
		Arguments: map[string]any{
			"project_root": workDir,
			"scope":        "project",
			"content":      pipeline,
		},
	})
	if err != nil {
		t.Fatalf("vivechak_save_plan call: %v", err)
	}
	env = parseTestEnvelope(t, res)
	if !env.Success {
		t.Fatalf("vivechak_save_plan failed: %s", env.Message)
	}

	// 8. Tool: vivechak_status
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_status",
		Arguments: map[string]any{"project_root": workDir},
	})
	if err != nil {
		t.Fatalf("vivechak_status call: %v", err)
	}
	env = parseTestEnvelope(t, res)
	if !env.Success {
		t.Fatalf("vivechak_status failed: %s", env.Message)
	}

	// 9. Tool: vivechak_next_session -> T1-01
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": workDir},
	})
	if err != nil {
		t.Fatalf("vivechak_next_session call: %v", err)
	}
	env = parseTestEnvelope(t, res)
	if env.Data["session_id"] != "T1-01" {
		t.Fatalf("expected next session T1-01, got %v", env.Data["session_id"])
	}

	// 10. Tool: vivechak_save_session (T1-01)
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": workDir,
			"session_id":   "T1-01",
			"content": `---
session_id: T1-01
title: Database Selection
status: complete
---
# Findings
Use RocksDB. A (benchmark)
`,
		},
	})
	if err != nil {
		t.Fatalf("vivechak_save_session call: %v", err)
	}
	env = parseTestEnvelope(t, res)
	if !env.Success {
		t.Fatalf("save_session failed: %s", env.Message)
	}

	// 11. Tool: vivechak_next_session -> T1-02
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_next_session",
		Arguments: map[string]any{"project_root": workDir},
	})
	if err != nil {
		t.Fatalf("vivechak_next_session call: %v", err)
	}
	env = parseTestEnvelope(t, res)
	if env.Data["session_id"] != "T1-02" {
		t.Fatalf("expected next session T1-02, got %v", env.Data["session_id"])
	}

	// 12. Save T1-02
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save_session",
		Arguments: map[string]any{
			"project_root": workDir,
			"session_id":   "T1-02",
			"content": `---
session_id: T1-02
title: Consensus
status: complete
---
# Findings
Use Raft. A (official spec)
`,
		},
	})
	if err != nil {
		t.Fatalf("vivechak_save_session call: %v", err)
	}
	env = parseTestEnvelope(t, res)
	if !env.Success {
		t.Fatalf("save_session failed: %s", env.Message)
	}

	// 13. Verify upstream context injection in SYN-01
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_next_session",
		Arguments: map[string]any{
			"project_root": workDir,
			"session_id":   "SYN-01",
		},
	})
	if err != nil {
		t.Fatalf("next_session SYN-01 call: %v", err)
	}
	env = parseTestEnvelope(t, res)
	synPrompt, _ := env.Data["prompt"].(string)
	if !strings.Contains(synPrompt, "RocksDB") || !strings.Contains(synPrompt, "Raft") {
		t.Errorf("expected synthesized findings from both upstream sessions, got %q", synPrompt)
	}

	// 14. Tool: vivechak_record_decision (D-001)
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root":  workDir,
			"decision_id":   "D-001",
			"artifact_type": "decision",
			"content": `---
decision_id: D-001
title: Storage Engine
status: accepted
door_type: one-way
reversal_triggers:
  - Latency exceeds 100ms
---
# Context
We need storage.
# Decision
RocksDB. A (benchmark)
# Consequences
Fast writes.
`,
		},
	})
	if err != nil {
		t.Fatalf("vivechak_record_decision call: %v", err)
	}
	env = parseTestEnvelope(t, res)
	if !env.Success {
		t.Fatalf("record_decision failed: %s", env.Message)
	}

	// 15. Tool: vivechak_validate (dry-run)
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_validate",
		Arguments: map[string]any{
			"artifact_type": "decision",
			"content": `---
decision_id: D-001
title: Storage Engine
status: accepted
---
# Context
Dry run test.
`,
		},
	})
	if err != nil {
		t.Fatalf("vivechak_validate call: %v", err)
	}
	env = parseTestEnvelope(t, res)
	if !env.Success {
		t.Fatalf("validate failed: %s", env.Message)
	}

	// 16. Save FAD & SYN-01
	fadContent := `---
session_id: SYN-01
title: Founding Architecture Document
status: accepted
---
# FAD
## Summary
Ledger architecture. A (pipeline)
## Decisions
Storage and consensus accepted. A (formal proof)
`
	cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save_session",
		Arguments: map[string]any{"project_root": workDir, "session_id": "FAD", "content": fadContent},
	})
	cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "vivechak_save_session",
		Arguments: map[string]any{"project_root": workDir, "session_id": "SYN-01", "content": fadContent},
	})

	// 17. Tool: vivechak_run_gate
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_run_gate",
		Arguments: map[string]any{
			"project_root": workDir,
			"verbose":      true,
		},
	})
	if err != nil {
		t.Fatalf("vivechak_run_gate call: %v", err)
	}
	env = parseTestEnvelope(t, res)
	if env.Data["gate_status"] != "PASS" || env.Data["gate_passed"] != true {
		t.Fatalf("expected gate PASS, got status=%v passed=%v", env.Data["gate_status"], env.Data["gate_passed"])
	}
}
