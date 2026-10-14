package main

import (
	"bytes"
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

// TestLiveHarness_EndToEnd_All21Hosts_Simulation validates Level 3 (idempotence, non-destructive merging,
// state preservation) and Level 4 (live launch round-trip via extracted config) across ALL 21 supported AI harnesses.
func TestLiveHarness_EndToEnd_All21Hosts_Simulation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping exhaustive harness simulation in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	// Build the live vivechak binary
	binDir := t.TempDir()
	binName := "vivechak_harness_tester"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build test binary: %v\noutput: %s", err, string(out))
	}

	all21Presets := []struct {
		preset        string
		serverKey     string
		isWorkspace   bool
		isYAML        bool
		isOpenCode    bool
		preExisting   string
		verifyContent func(t *testing.T, content []byte) (cmd string, args []string)
	}{
		{
			preset:      "claude-code",
			serverKey:   "mcpServers",
			isWorkspace: false,
			preExisting: `{"numStartups": 1337, "userEmail": "engineer@example.com", "approvedTools": ["bash", "git"], "modelPref": "claude-3-7-sonnet"}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				dec := json.NewDecoder(bytes.NewReader(content))
				dec.UseNumber()
				if err := dec.Decode(&parsed); err != nil {
					t.Fatalf("failed to parse claude-code JSON: %v", err)
				}
				if parsed["numStartups"].(json.Number).String() != "1337" {
					t.Errorf("claude-code numStartups clobbered: %v", parsed["numStartups"])
				}
				if parsed["userEmail"] != "engineer@example.com" {
					t.Errorf("claude-code userEmail clobbered: %v", parsed["userEmail"])
				}
				mcpServers, ok := parsed["mcpServers"].(map[string]any)
				if !ok {
					t.Fatalf("missing mcpServers in claude-code config: %v", parsed)
				}
				vivechak, ok := mcpServers["vivechak"].(map[string]any)
				if !ok {
					t.Fatalf("missing mcpServers.vivechak in claude-code: %v", mcpServers)
				}
				cmd := vivechak["command"].(string)
				var args []string
				for _, a := range vivechak["args"].([]any) {
					args = append(args, a.(string))
				}
				return cmd, args
			},
		},
		{
			preset:      "mcp",
			serverKey:   "mcpServers",
			isWorkspace: true,
			preExisting: `{"mcpServers": {"existing_fetch": {"command": "fetch-mcp", "args": ["--port", "8080"]}}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse universal .mcp.json: %v", err)
				}
				mcpServers := parsed["mcpServers"].(map[string]any)
				if mcpServers["existing_fetch"] == nil {
					t.Errorf("universal .mcp.json clobbered existing_fetch server")
				}
				vivechak := mcpServers["vivechak"].(map[string]any)
				cmd := vivechak["command"].(string)
				var args []string
				for _, a := range vivechak["args"].([]any) {
					args = append(args, a.(string))
				}
				return cmd, args
			},
		},
		{
			preset:      "opencode",
			serverKey:   "mcp",
			isWorkspace: true,
			isOpenCode:  true,
			preExisting: `{"providers": [{"name": "anthropic", "apiKey": "sk-secret-key"}]}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse opencode.json: %v", err)
				}
				if parsed["providers"] == nil {
					t.Errorf("opencode providers clobbered")
				}
				mcpBlock, ok := parsed["mcp"].(map[string]any)
				if !ok {
					t.Fatalf("opencode config missing root 'mcp' key: %v", parsed)
				}
				vivechak, ok := mcpBlock["vivechak"].(map[string]any)
				if !ok {
					t.Fatalf("missing mcp.vivechak in opencode: %v", mcpBlock)
				}
				if vivechak["type"] != "local" {
					t.Errorf("opencode expected type 'local', got %v", vivechak["type"])
				}
				if vivechak["enabled"] != true {
					t.Errorf("opencode expected enabled true, got %v", vivechak["enabled"])
				}
				cmdArray, ok := vivechak["command"].([]any)
				if !ok || len(cmdArray) < 2 {
					t.Fatalf("opencode expected array command, got %v", vivechak["command"])
				}
				cmd := cmdArray[0].(string)
				var args []string
				for _, a := range cmdArray[1:] {
					args = append(args, a.(string))
				}
				return cmd, args
			},
		},
		{
			preset:      "continue",
			serverKey:   "mcpServers",
			isWorkspace: true,
			preExisting: ``,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse continue vivechak.json: %v", err)
				}
				mcpServers := parsed["mcpServers"].(map[string]any)
				vivechak := mcpServers["vivechak"].(map[string]any)
				cmd := vivechak["command"].(string)
				var args []string
				for _, a := range vivechak["args"].([]any) {
					args = append(args, a.(string))
				}
				return cmd, args
			},
		},
		{
			preset:      "goose",
			serverKey:   "extensions",
			isWorkspace: false,
			isYAML:      true,
			preExisting: "GOOSE_PROVIDER: openai\nGOOSE_MODEL: gpt-4o\n",
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				str := string(content)
				if !strings.Contains(str, "GOOSE_PROVIDER: openai") {
					t.Errorf("goose pre-existing GOOSE_PROVIDER clobbered")
				}
				if !strings.Contains(str, "extensions:") {
					t.Fatalf("goose missing extensions key in YAML:\n%s", str)
				}
				if !strings.Contains(str, "vivechak:") {
					t.Fatalf("goose missing vivechak extension in YAML:\n%s", str)
				}
				if !strings.Contains(str, "type: stdio") {
					t.Errorf("goose missing type: stdio in YAML:\n%s", str)
				}
				if !strings.Contains(str, "enabled: true") {
					t.Errorf("goose missing enabled: true in YAML:\n%s", str)
				}
				return binPath, []string{"serve"}
			},
		},
		{
			preset:      "cursor",
			serverKey:   "mcpServers",
			isWorkspace: true,
			preExisting: `{"mcpServers": {"git_helper": {"command": "git-mcp"}}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse cursor mcp.json: %v", err)
				}
				mcpServers := parsed["mcpServers"].(map[string]any)
				if mcpServers["git_helper"] == nil {
					t.Errorf("cursor git_helper clobbered")
				}
				vivechak := mcpServers["vivechak"].(map[string]any)
				cmd := vivechak["command"].(string)
				var args []string
				for _, a := range vivechak["args"].([]any) {
					args = append(args, a.(string))
				}
				return cmd, args
			},
		},
		{
			preset:      "vscode",
			serverKey:   "servers",
			isWorkspace: true,
			preExisting: `{"servers": {"workspace_linter": {"command": "linter"}}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse vscode mcp.json: %v", err)
				}
				servers := parsed["servers"].(map[string]any)
				if servers["workspace_linter"] == nil {
					t.Errorf("vscode workspace_linter clobbered")
				}
				vivechak := servers["vivechak"].(map[string]any)
				cmd := vivechak["command"].(string)
				var args []string
				for _, a := range vivechak["args"].([]any) {
					args = append(args, a.(string))
				}
				return cmd, args
			},
		},
		{
			preset:      "zed",
			serverKey:   "context_servers",
			isWorkspace: false,
			preExisting: `{"theme": "One Dark", "ui_font_size": 15, "context_servers": {"custom_rag": {"command": "rag-server"}}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse zed settings.json: %v", err)
				}
				if parsed["theme"] != "One Dark" {
					t.Errorf("zed theme clobbered")
				}
				servers := parsed["context_servers"].(map[string]any)
				if servers["custom_rag"] == nil {
					t.Errorf("zed custom_rag clobbered")
				}
				vivechak := servers["vivechak"].(map[string]any)
				cmd := vivechak["command"].(string)
				var args []string
				for _, a := range vivechak["args"].([]any) {
					args = append(args, a.(string))
				}
				return cmd, args
			},
		},
		{
			preset:      "copilot",
			serverKey:   "mcpServers",
			isWorkspace: false,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse copilot mcp.json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "void",
			serverKey:   "mcpServers",
			isWorkspace: true,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse void mcp.json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "claude-desktop",
			serverKey:   "mcpServers",
			isWorkspace: false,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse claude-desktop json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "windsurf",
			serverKey:   "mcpServers",
			isWorkspace: false,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse windsurf json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "antigravity",
			serverKey:   "mcpServers",
			isWorkspace: false,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse antigravity json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "kiro",
			serverKey:   "mcpServers",
			isWorkspace: false,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse kiro json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "trae",
			serverKey:   "mcpServers",
			isWorkspace: true,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse trae json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "omp",
			serverKey:   "mcpServers",
			isWorkspace: false,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse omp json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "openhands",
			serverKey:   "mcpServers",
			isWorkspace: false,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse openhands json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "droid",
			serverKey:   "mcpServers",
			isWorkspace: true,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse droid json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "cline",
			serverKey:   "mcpServers",
			isWorkspace: false,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse cline json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "roo",
			serverKey:   "mcpServers",
			isWorkspace: false,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse roo json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
		{
			preset:      "devin",
			serverKey:   "mcpServers",
			isWorkspace: true,
			preExisting: `{"mcpServers": {}}`,
			verifyContent: func(t *testing.T, content []byte) (string, []string) {
				var parsed map[string]any
				if err := json.Unmarshal(content, &parsed); err != nil {
					t.Fatalf("failed to parse devin json: %v", err)
				}
				vivechak := parsed["mcpServers"].(map[string]any)["vivechak"].(map[string]any)
				return vivechak["command"].(string), []string{"serve"}
			},
		},
	}

	for _, tc := range all21Presets {
		tc := tc
		t.Run("Harness_"+tc.preset, func(t *testing.T) {
			mockHome := t.TempDir()
			mockWs := t.TempDir()
			mockAppData := filepath.Join(mockHome, "AppData", "Roaming")
			_ = os.MkdirAll(mockAppData, 0o755)

			p := findPreset(tc.preset)
			if p == nil {
				t.Fatalf("findPreset(%q) returned nil", tc.preset)
			}
			cfgPath, err := p.ResolvePath(mockHome, mockAppData, mockWs, runtime.GOOS)
			if err != nil {
				t.Fatalf("ResolvePath failed for %q: %v", tc.preset, err)
			}

			// Pre-seed config file if preExisting is specified
			if tc.preExisting != "" {
				_ = os.MkdirAll(filepath.Dir(cfgPath), 0o755)
				if err := os.WriteFile(cfgPath, []byte(tc.preExisting), 0o644); err != nil {
					t.Fatalf("failed to write preExisting: %v", err)
				}
			}

			// Helper to run vck setup in the mock environment
			runSetupSubprocess := func() ([]byte, error) {
				cmd := exec.CommandContext(ctx, binPath, "setup", tc.preset)
				cmd.Dir = mockWs
				// Filter current env to isolate HOME, USERPROFILE, APPDATA
				var cleanEnv []string
				for _, e := range os.Environ() {
					key := strings.SplitN(e, "=", 2)[0]
					if strings.EqualFold(key, "HOME") || strings.EqualFold(key, "USERPROFILE") || strings.EqualFold(key, "APPDATA") {
						continue
					}
					cleanEnv = append(cleanEnv, e)
				}
				cmd.Env = append(cleanEnv,
					"HOME="+mockHome,
					"USERPROFILE="+mockHome,
					"APPDATA="+mockAppData,
				)
				return cmd.CombinedOutput()
			}

			// 1. First Run: setup
			out1, err1 := runSetupSubprocess()
			if err1 != nil {
				t.Fatalf("setup %s failed: %v\noutput: %s", tc.preset, err1, string(out1))
			}

			// 2. Read and verify content
			updated, err := os.ReadFile(cfgPath)
			if err != nil {
				t.Fatalf("failed to read updated config %s: %v", cfgPath, err)
			}
			serverCmd, serverArgs := tc.verifyContent(t, updated)

			// 3. Idempotence Run: run setup again
			out2, err2 := runSetupSubprocess()
			if err2 != nil {
				t.Fatalf("second setup %s failed: %v\noutput: %s", tc.preset, err2, string(out2))
			}
			updatedAgain, _ := os.ReadFile(cfgPath)
			if tc.isYAML {
				// Count occurrences of 'vivechak:' in YAML to ensure no duplicates
				if strings.Count(string(updatedAgain), "vivechak:") != 1 {
					t.Errorf("expected exactly 1 'vivechak:' block after idempotent run, got:\n%s", string(updatedAgain))
				}
			} else {
				var parsedTwice map[string]any
				if err := json.Unmarshal(updatedAgain, &parsedTwice); err != nil {
					t.Fatalf("failed to parse json after second run: %v", err)
				}
				// Verify pre-existing and single vivechak key
				sk := tc.serverKey
				subMap, ok := parsedTwice[sk].(map[string]any)
				if !ok || subMap["vivechak"] == nil {
					t.Errorf("second run corrupted %s.%s: %v", sk, "vivechak", parsedTwice)
				}
			}

			// 4. Live Launch Round-Trip: Connect to Vivechak using the exact command & args from config
			t.Run("LiveMCPHandshake", func(t *testing.T) {
				srvCtx, srvCancel := context.WithTimeout(ctx, 15*time.Second)
				defer srvCancel()

				mcpCmd := exec.CommandContext(srvCtx, serverCmd, serverArgs...)
				transport := &mcp.CommandTransport{Command: mcpCmd}
				client := mcp.NewClient(
					&mcp.Implementation{Name: "test-" + tc.preset, Version: "1.0.0"},
					nil,
				)

				session, err := client.Connect(srvCtx, transport, nil)
				if err != nil {
					t.Fatalf("failed to connect to server spawned from %s config: %v", tc.preset, err)
				}
				defer session.Close()

				// Call tools/list and verify all 13 tools
				toolsRes, err := session.ListTools(srvCtx, nil)
				if err != nil {
					t.Fatalf("ListTools failed: %v", err)
				}
				if len(toolsRes.Tools) != 13 {
					t.Errorf("expected 13 tools, got %d", len(toolsRes.Tools))
				}

				foundInit := false
				foundGate := false
				for _, tool := range toolsRes.Tools {
					if tool.Name == "vivechak_init" {
						foundInit = true
					}
					if tool.Name == "vivechak_run_gate" {
						foundGate = true
					}
				}
				if !foundInit || !foundGate {
					t.Errorf("missing core tools in list: %v", toolsRes.Tools)
				}
			})
		})
	}
}

// TestLiveHarness_Adversarial_EdgeCases validates resilient handling of malformed JSON,
// comments, trailing commas, and uncorrupted file state on failure.
func TestLiveHarness_Adversarial_EdgeCases(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)

	binDir := t.TempDir()
	binName := "vivechak_edge_bin"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build test binary: %v\noutput: %s", err, string(out))
	}

	// 1. Comments and trailing commas in JSON (Cursor / VS Code style JSONC)
	t.Run("JSONC_Comments_And_Trailing_Commas", func(t *testing.T) {
		ws := t.TempDir()
		cursorDir := filepath.Join(ws, ".cursor")
		_ = os.MkdirAll(cursorDir, 0o755)
		cfgFile := filepath.Join(cursorDir, "mcp.json")

		jsoncContent := `{
			// User comments at top
			"mcpServers": {
				/* Multiline block comment */
				"legacy_tool": {
					"command": "tool",
					"args": ["serve"], // trailing comma
				},
			},
		}`
		if err := os.WriteFile(cfgFile, []byte(jsoncContent), 0o644); err != nil {
			t.Fatalf("failed to write jsonc: %v", err)
		}

		cmd := exec.CommandContext(ctx, binPath, "setup", "cursor")
		cmd.Dir = ws
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("setup cursor on JSONC failed: %v\noutput: %s", err, string(out))
		}

		resBytes, _ := os.ReadFile(cfgFile)
		var parsed map[string]any
		if err := json.Unmarshal(resBytes, &parsed); err != nil {
			t.Fatalf("resulting config is not valid standard JSON: %v\ncontent: %s", err, string(resBytes))
		}
		mcpServers := parsed["mcpServers"].(map[string]any)
		if mcpServers["legacy_tool"] == nil || mcpServers["vivechak"] == nil {
			t.Errorf("expected both legacy_tool and vivechak, got: %v", mcpServers)
		}
	})

	// 2. Corrupt JSON syntax must fail cleanly without wiping existing file
	t.Run("Corrupt_JSON_Safety", func(t *testing.T) {
		ws := t.TempDir()
		cursorDir := filepath.Join(ws, ".cursor")
		_ = os.MkdirAll(cursorDir, 0o755)
		cfgFile := filepath.Join(cursorDir, "mcp.json")

		corruptContent := `{"mcpServers": { UNCLOSED BRACE...`
		_ = os.WriteFile(cfgFile, []byte(corruptContent), 0o644)

		cmd := exec.CommandContext(ctx, binPath, "setup", "cursor")
		cmd.Dir = ws
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Errorf("expected setup to fail on corrupt JSON, but succeeded! output: %s", string(out))
		}

		// Ensure original corrupt content was NOT wiped
		afterBytes, _ := os.ReadFile(cfgFile)
		if string(afterBytes) != corruptContent {
			t.Errorf("corrupted file was modified or deleted on error! got: %s", string(afterBytes))
		}
	})

	// 3. Workspace auto-detection with portable .mcp.json
	t.Run("AutoDetect_Portable_MCP_JSON", func(t *testing.T) {
		ws := t.TempDir()
		mcpFile := filepath.Join(ws, ".mcp.json")
		_ = os.WriteFile(mcpFile, []byte(`{"mcpServers": {"existing": {"command": "echo"}}}`), 0o644)

		// Run setup with NO target argument inside ws
		cmd := exec.CommandContext(ctx, binPath, "setup")
		cmd.Dir = ws
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("setup with zero args on .mcp.json failed: %v\noutput: %s", err, string(out))
		}
		if !strings.Contains(string(out), "Universal Portable MCP") && !strings.Contains(string(out), ".mcp.json") {
			t.Errorf("expected detection banner in output, got: %s", string(out))
		}

		resBytes, _ := os.ReadFile(mcpFile)
		var parsed map[string]any
		if err := json.Unmarshal(resBytes, &parsed); err != nil {
			t.Fatalf("failed to parse updated .mcp.json: %v", err)
		}
		servers := parsed["mcpServers"].(map[string]any)
		if servers["existing"] == nil || servers["vivechak"] == nil {
			t.Errorf("missing existing or vivechak in .mcp.json: %v", servers)
		}
	})

	// 4. Workspace auto-detection with opencode.json
	t.Run("AutoDetect_OpenCode_JSON", func(t *testing.T) {
		ws := t.TempDir()
		opencodeFile := filepath.Join(ws, "opencode.json")
		_ = os.WriteFile(opencodeFile, []byte(`{"model": "gpt-4o"}`), 0o644)

		// Run setup with NO target argument inside ws
		cmd := exec.CommandContext(ctx, binPath, "setup")
		cmd.Dir = ws
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("setup with zero args on opencode.json failed: %v\noutput: %s", err, string(out))
		}
		if !strings.Contains(string(out), "OpenCode") {
			t.Errorf("expected OpenCode detection banner in output, got: %s", string(out))
		}

		resBytes, _ := os.ReadFile(opencodeFile)
		var parsed map[string]any
		if err := json.Unmarshal(resBytes, &parsed); err != nil {
			t.Fatalf("failed to parse updated opencode.json: %v", err)
		}
		mcpBlock := parsed["mcp"].(map[string]any)
		if mcpBlock["vivechak"] == nil {
			t.Errorf("missing vivechak in opencode.json: %v", parsed)
		}
	})
}
