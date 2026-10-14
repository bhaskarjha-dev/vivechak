package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveConfigPath(t *testing.T) {
	homeDir := "/home/user"
	cwd := "/workspace"
	appData := "/appdata"

	// Windsurf
	p, err := resolveConfigPath("windsurf", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("windsurf: %v", err)
	}
	expectedWindsurf := filepath.Join(homeDir, ".codeium", "windsurf", "mcp_config.json")
	if p != expectedWindsurf {
		t.Errorf("windsurf path: got %s, want %s", p, expectedWindsurf)
	}

	// Cursor
	p, err = resolveConfigPath("cursor", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("cursor: %v", err)
	}
	expectedCursor := filepath.Join(homeDir, ".cursor", "mcp.json")
	if p != expectedCursor {
		t.Errorf("cursor path: got %s, want %s", p, expectedCursor)
	}

	// VSCode
	p, err = resolveConfigPath("vscode", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("vscode: %v", err)
	}
	expectedVSCode := filepath.Join(cwd, ".vscode", "mcp.json")
	if p != expectedVSCode {
		t.Errorf("vscode path: got %s, want %s", p, expectedVSCode)
	}

	// Claude Desktop on Windows
	p, err = resolveConfigPath("claude-desktop", homeDir, "C:\\Users\\user\\AppData\\Roaming", cwd, "windows")
	if err != nil {
		t.Fatalf("claude windows: %v", err)
	}
	expectedClaudeWin := filepath.Join("C:\\Users\\user\\AppData\\Roaming", "Claude", "claude_desktop_config.json")
	if p != expectedClaudeWin {
		t.Errorf("claude windows path: got %s, want %s", p, expectedClaudeWin)
	}

	// Claude Desktop on macOS (darwin)
	p, err = resolveConfigPath("claude-desktop", homeDir, appData, cwd, "darwin")
	if err != nil {
		t.Fatalf("claude darwin: %v", err)
	}
	expectedClaudeMac := filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	if p != expectedClaudeMac {
		t.Errorf("claude darwin path: got %s, want %s", p, expectedClaudeMac)
	}

	// Claude Desktop on Linux
	p, err = resolveConfigPath("claude-desktop", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("claude linux: %v", err)
	}
	expectedClaudeLinux := filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json")
	if p != expectedClaudeLinux {
		t.Errorf("claude linux path: got %s, want %s", p, expectedClaudeLinux)
	}

	// Zed on Linux
	p, err = resolveConfigPath("zed", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("zed linux: %v", err)
	}
	expectedZedLinux := filepath.Join(homeDir, ".config", "zed", "settings.json")
	if p != expectedZedLinux {
		t.Errorf("zed linux path: got %s, want %s", p, expectedZedLinux)
	}

	// Zed on Windows
	p, err = resolveConfigPath("zed", homeDir, "C:\\Users\\user\\AppData\\Roaming", cwd, "windows")
	if err != nil {
		t.Fatalf("zed windows: %v", err)
	}
	expectedZedWin := filepath.Join("C:\\Users\\user\\AppData\\Roaming", "Zed", "settings.json")
	if p != expectedZedWin {
		t.Errorf("zed windows path: got %s, want %s", p, expectedZedWin)
	}

	// Kiro (canonical ~/.kiro/settings/mcp.json by default)
	p, err = resolveConfigPath("kiro", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("kiro: %v", err)
	}
	expectedKiro := filepath.Join(homeDir, ".kiro", "settings", "mcp.json")
	if p != expectedKiro {
		t.Errorf("kiro path: got %s, want %s", p, expectedKiro)
	}

	// Kiro (workspace when .kiro exists in cwd)
	tempKiroWS := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tempKiroWS, ".kiro"), 0o755)
	p, err = resolveConfigPath("kiro", homeDir, appData, tempKiroWS, "linux")
	if err != nil {
		t.Fatalf("kiro workspace: %v", err)
	}
	expectedKiroWS := filepath.Join(tempKiroWS, ".kiro", "settings", "mcp.json")
	if p != expectedKiroWS {
		t.Errorf("kiro workspace path: got %s, want %s", p, expectedKiroWS)
	}

	// Kiro (legacy fallback when ~/.aws/.kiro exists)
	tempKiroHome := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tempKiroHome, ".aws", ".kiro"), 0o755)
	p, err = resolveConfigPath("kiro", tempKiroHome, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("kiro legacy: %v", err)
	}
	expectedKiroLegacy := filepath.Join(tempKiroHome, ".aws", ".kiro", "mcp.json")
	if p != expectedKiroLegacy {
		t.Errorf("kiro legacy path: got %s, want %s", p, expectedKiroLegacy)
	}

	// Antigravity (global default when no .agents in cwd)
	p, err = resolveConfigPath("antigravity", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("antigravity: %v", err)
	}
	expectedAGGlobal := filepath.Join(homeDir, ".gemini", "config", "mcp_config.json")
	if p != expectedAGGlobal {
		t.Errorf("antigravity global path: got %s, want %s", p, expectedAGGlobal)
	}

	// Antigravity (workspace when .agents exists in cwd)
	tempWS := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tempWS, ".agents"), 0o755)
	p, err = resolveConfigPath("antigravity", homeDir, appData, tempWS, "linux")
	if err != nil {
		t.Fatalf("antigravity workspace: %v", err)
	}
	expectedAGWS := filepath.Join(tempWS, ".agents", "mcp_config.json")
	if p != expectedAGWS {
		t.Errorf("antigravity workspace path: got %s, want %s", p, expectedAGWS)
	}

	// Trae
	p, err = resolveConfigPath("trae", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("trae: %v", err)
	}
	expectedTrae := filepath.Join(cwd, ".trae", "mcp.json")
	if p != expectedTrae {
		t.Errorf("trae path: got %s, want %s", p, expectedTrae)
	}

	// OMP (global)
	p, err = resolveConfigPath("omp", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("omp: %v", err)
	}
	expectedOMP := filepath.Join(homeDir, ".omp", "agent", "mcp.json")
	if p != expectedOMP {
		t.Errorf("omp path: got %s, want %s", p, expectedOMP)
	}

	// OpenHands (global)
	p, err = resolveConfigPath("openhands", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("openhands: %v", err)
	}
	expectedOH := filepath.Join(homeDir, ".openhands", "mcp.json")
	if p != expectedOH {
		t.Errorf("openhands path: got %s, want %s", p, expectedOH)
	}

	// Factory Droid (global)
	p, err = resolveConfigPath("droid", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("droid: %v", err)
	}
	expectedDroid := filepath.Join(homeDir, ".factory", "mcp.json")
	if p != expectedDroid {
		t.Errorf("droid path: got %s, want %s", p, expectedDroid)
	}

	// Cline (Linux, Mac, Windows)
	p, err = resolveConfigPath("cline", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("cline linux: %v", err)
	}
	expectedClineLinux := filepath.Join(homeDir, ".config", "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json")
	if p != expectedClineLinux {
		t.Errorf("cline linux path: got %s, want %s", p, expectedClineLinux)
	}
	p, err = resolveConfigPath("cline", homeDir, "C:\\Users\\user\\AppData\\Roaming", cwd, "windows")
	if err != nil {
		t.Fatalf("cline windows: %v", err)
	}
	expectedClineWin := filepath.Join("C:\\Users\\user\\AppData\\Roaming", "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json")
	if p != expectedClineWin {
		t.Errorf("cline windows path: got %s, want %s", p, expectedClineWin)
	}

	// Roo (Linux, Windows)
	p, err = resolveConfigPath("roo", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("roo linux: %v", err)
	}
	expectedRooLinux := filepath.Join(homeDir, ".config", "Code", "User", "globalStorage", "rooveterinaryinc.roo-cline", "settings", "cline_mcp_settings.json")
	if p != expectedRooLinux {
		t.Errorf("roo linux path: got %s, want %s", p, expectedRooLinux)
	}

	// Devin
	p, err = resolveConfigPath("devin", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("devin: %v", err)
	}
	expectedDevin := filepath.Join(cwd, ".devin", "mcp_config.local.json")
	if p != expectedDevin {
		t.Errorf("devin path: got %s, want %s", p, expectedDevin)
	}

	// Unsupported write clients return error
	for _, client := range []string{"chatgpt", "codex", "invalidclient"} {
		_, err = resolveConfigPath(client, homeDir, appData, cwd, "linux")
		if err == nil {
			t.Errorf("%s: expected error for unsupported write client, got nil", client)
		}
	}
}

func TestDetermineServerKey(t *testing.T) {
	// VS Code default is servers
	if k := determineServerKey("vscode", "/ws/.vscode/mcp.json", nil); k != "servers" {
		t.Errorf("vscode default key: got %s, want servers", k)
	}

	// VS Code with existing mcpServers retains mcpServers
	existing := map[string]any{"mcpServers": map[string]any{}}
	if k := determineServerKey("vscode", "/ws/.vscode/mcp.json", existing); k != "mcpServers" {
		t.Errorf("vscode with mcpServers: got %s, want mcpServers", k)
	}

	// Zed is context_servers
	if k := determineServerKey("zed", "/home/user/.config/zed/settings.json", nil); k != "context_servers" {
		t.Errorf("zed key: got %s, want context_servers", k)
	}

	// Default client is mcpServers
	if k := determineServerKey("cursor", "/home/user/.cursor/mcp.json", nil); k != "mcpServers" {
		t.Errorf("cursor key: got %s, want mcpServers", k)
	}

	// Path heuristic for .vscode/mcp.json without host
	if k := determineServerKey("", filepath.Join("project", ".vscode", "mcp.json"), nil); k != "servers" {
		t.Errorf("path heuristic .vscode key: got %s, want servers", k)
	}

	// Path heuristic for zed settings.json without host (Unix and Windows)
	if k := determineServerKey("", filepath.Join("home", "user", ".config", "zed", "settings.json"), nil); k != "context_servers" {
		t.Errorf("path heuristic zed Unix key: got %s, want context_servers", k)
	}
	if k := determineServerKey("", filepath.Join("AppData", "Roaming", "Zed", "settings.json"), nil); k != "context_servers" {
		t.Errorf("path heuristic zed Windows key: got %s, want context_servers", k)
	}

	// Path containing 'zed' as substring (e.g. user jzed, or directory analyzed) should NOT trigger zed context_servers
	if k := determineServerKey("", filepath.Join("home", "jzed", "project", "mcp.json"), nil); k != "mcpServers" {
		t.Errorf("substring zed in username: got %s, want mcpServers", k)
	}
	if k := determineServerKey("", filepath.Join("workspace", "analyzed", "mcp.json"), nil); k != "mcpServers" {
		t.Errorf("substring zed in directory: got %s, want mcpServers", k)
	}
}

func TestMergeConfig(t *testing.T) {
	exePath := "/path/to/vivechak"

	t.Run("empty config creates fresh vivechak entry", func(t *testing.T) {
		out, err := mergeConfig(nil, exePath)
		if err != nil {
			t.Fatalf("mergeConfig(nil): %v", err)
		}
		var parsed map[string]any
		if err := json.Unmarshal(out, &parsed); err != nil {
			t.Fatalf("unmarshal merged config: %v", err)
		}
		servers, ok := parsed["mcpServers"].(map[string]any)
		if !ok {
			t.Fatalf("mcpServers not a map: %v", parsed)
		}
		v, ok := servers["vivechak"].(map[string]any)
		if !ok || v["command"] != exePath {
			t.Errorf("expected command %s, got %v", exePath, v)
		}
	})

	t.Run("creates fresh servers entry for vscode", func(t *testing.T) {
		out, err := mergeConfig(nil, exePath, "servers")
		if err != nil {
			t.Fatalf("mergeConfig(nil, servers): %v", err)
		}
		var parsed map[string]any
		if err := json.Unmarshal(out, &parsed); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		servers, ok := parsed["servers"].(map[string]any)
		if !ok {
			t.Fatalf("servers key not found: %v", parsed)
		}
		if servers["vivechak"] == nil {
			t.Error("vivechak server missing")
		}
	})

	t.Run("creates fresh context_servers entry for zed", func(t *testing.T) {
		out, err := mergeConfig(nil, exePath, "context_servers")
		if err != nil {
			t.Fatalf("mergeConfig(nil, context_servers): %v", err)
		}
		var parsed map[string]any
		if err := json.Unmarshal(out, &parsed); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		servers, ok := parsed["context_servers"].(map[string]any)
		if !ok {
			t.Fatalf("context_servers key not found: %v", parsed)
		}
		if servers["vivechak"] == nil {
			t.Error("vivechak server missing")
		}
	})

	t.Run("preserves number fidelity with UseNumber", func(t *testing.T) {
		// Existing config with large integer and float fields
		input := []byte(`{
			"port": 8080,
			"timeout": 9223372036854775807,
			"mcpServers": {
				"other": {
					"command": "other-cmd"
				}
			}
		}`)
		out, err := mergeConfig(input, exePath)
		if err != nil {
			t.Fatalf("mergeConfig: %v", err)
		}

		// Decode with UseNumber to verify fidelity was not lost
		dec := json.NewDecoder(bytes.NewReader(out))
		dec.UseNumber()
		var parsed map[string]any
		if err := dec.Decode(&parsed); err != nil {
			t.Fatalf("decode: %v", err)
		}

		portNum, ok := parsed["port"].(json.Number)
		if !ok || portNum.String() != "8080" {
			t.Errorf("port number altered: %v", parsed["port"])
		}

		timeoutNum, ok := parsed["timeout"].(json.Number)
		if !ok || timeoutNum.String() != "9223372036854775807" {
			t.Errorf("timeout number altered: %v", parsed["timeout"])
		}

		servers := parsed["mcpServers"].(map[string]any)
		if servers["other"] == nil || servers["vivechak"] == nil {
			t.Errorf("lost other server or missing vivechak: %v", servers)
		}
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		_, err := mergeConfig([]byte("{invalid-json"), exePath)
		if err == nil {
			t.Error("expected error for invalid json, got nil")
		}
	})

	t.Run("non-object mcpServers returns error", func(t *testing.T) {
		input := []byte(`{"mcpServers": "not-an-object"}`)
		_, err := mergeConfig(input, exePath)
		if err == nil || !strings.Contains(err.Error(), "not a JSON object") {
			t.Errorf("expected error about JSON object, got %v", err)
		}
	})

	t.Run("jsonc with comments and trailing commas", func(t *testing.T) {
		input := []byte(`{
			// VS Code / Zed style settings comment
			"context_servers": {
				/* existing server */
				"custom": {
					"command": "node",
					"args": ["server.js",],
				},
			},
		}`)
		out, err := mergeConfig(input, exePath, "context_servers")
		if err != nil {
			t.Fatalf("failed to merge JSONC config: %v", err)
		}
		var parsed map[string]any
		if err := json.Unmarshal(out, &parsed); err != nil {
			t.Fatalf("merged output is not valid JSON: %v", err)
		}
		servers := parsed["context_servers"].(map[string]any)
		if servers["custom"] == nil {
			t.Error("lost custom server from JSONC")
		}
		if servers["vivechak"] == nil {
			t.Error("missing vivechak in merged JSONC")
		}
	})
}

func TestStripJSONComments(t *testing.T) {
	raw := `{
		// Line comment
		"url": "https://example.com/test//not-a-comment",
		"comment_in_str": "/* also not comment */",
		"escaped_quote": "test \"with\" quotes",
		/* Multi
		   line
		   comment */
		"items": [
			1,
			2, // trailing comment
		],
		"enabled": true,
	}`

	cleaned := stripJSONComments([]byte(raw))
	var val map[string]any
	if err := json.Unmarshal(cleaned, &val); err != nil {
		t.Fatalf("unmarshal cleaned json failed: %v\nCleaned content:\n%s", err, string(cleaned))
	}

	if val["url"] != "https://example.com/test//not-a-comment" {
		t.Errorf("string with // was corrupted: %v", val["url"])
	}
	if val["comment_in_str"] != "/* also not comment */" {
		t.Errorf("string with /* was corrupted: %v", val["comment_in_str"])
	}
	items := val["items"].([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if val["enabled"] != true {
		t.Errorf("expected enabled true, got %v", val["enabled"])
	}
}

func TestWriteConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "nested", "sub", "config.json")
	content := []byte("{\"hello\": \"world\"}\n")

	if err := writeConfigFile(targetPath, content); err != nil {
		t.Fatalf("writeConfigFile failed: %v", err)
	}

	readBack, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("reading back file: %v", err)
	}
	if !bytes.Equal(readBack, content) {
		t.Errorf("got %q, want %q", readBack, content)
	}

	// Overwrite existing file
	newContent := []byte("{\"updated\": true}\n")
	if err := writeConfigFile(targetPath, newContent); err != nil {
		t.Fatalf("overwrite failed: %v", err)
	}
	readBack2, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("reading back updated file: %v", err)
	}
	if !bytes.Equal(readBack2, newContent) {
		t.Errorf("got %q, want %q", readBack2, newContent)
	}
}

func TestIsSupportedPreset(t *testing.T) {
	valid := []string{
		"cursor", "vscode", "code", "claude-desktop", "claude", "claude-code", "claudecode",
		"mcp", "portable", "universal", "opencode", "open-code", "copilot", "github-copilot",
		"void", "void-editor", "continue", "continuedev", "goose", "block-goose",
		"windsurf", "antigravity", "agy", "zed", "kiro", "trae", "omp", "openhands", "droid", "cline", "roo", "devin",
	}
	for _, v := range valid {
		if !isSupportedPreset(v) {
			t.Errorf("expected %s to be recognized as supported preset", v)
		}
	}

	invalid := []string{"", "random", "cursro", "vs-code", "sublime", "chatgpt", "codex"}
	for _, inv := range invalid {
		if isSupportedPreset(inv) {
			t.Errorf("expected %s to NOT be recognized as supported preset", inv)
		}
	}
}

func TestPresetAndClientCompatibility(t *testing.T) {
	home := "/test/home"
	appData := "/test/appdata"
	cwd := "/test/cwd"

	presets := []string{
		"cursor", "vscode", "code", "claude-desktop", "claude", "claude-code", "claudecode",
		"mcp", "portable", "universal", "opencode", "open-code", "copilot", "github-copilot",
		"void", "void-editor", "continue", "continuedev", "goose", "block-goose",
		"windsurf", "antigravity", "agy", "zed", "kiro", "trae", "omp", "openhands", "droid", "cline", "roo", "devin",
	}
	for _, p := range presets {
		pathFromPreset, err1 := resolveConfigPath(p, home, appData, cwd, "linux")
		if err1 != nil {
			t.Fatalf("resolveConfigPath(%q) failed: %v", p, err1)
		}
		if pathFromPreset == "" {
			t.Fatalf("expected non-empty path for preset %q", p)
		}
	}
}

func TestRunSetupWithArgs_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runSetupWithArgs([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage: vivechak setup") {
		t.Errorf("expected usage output, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "vck setup cursor") {
		t.Errorf("expected vck example, got: %s", stdout.String())
	}
}

func TestRunSetupWithArgs_DirectFileAndDryRun(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "test-mcp.json")

	// 1. Dry run should NOT create the file
	var stdoutDry, stderrDry bytes.Buffer
	code := runSetupWithArgs([]string{tempFile, "--dry-run"}, &stdoutDry, &stderrDry)
	if code != 0 {
		t.Fatalf("dry-run failed with code %d: %s", code, stderrDry.String())
	}
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create file on disk")
	}
	if !strings.Contains(stdoutDry.String(), "[dry-run]") {
		t.Errorf("expected [dry-run] marker, got: %s", stdoutDry.String())
	}

	// 2. Real setup should write the file
	var stdoutReal, stderrReal bytes.Buffer
	code = runSetupWithArgs([]string{tempFile}, &stdoutReal, &stderrReal)
	if code != 0 {
		t.Fatalf("setup failed with code %d: %s", code, stderrReal.String())
	}
	if _, err := os.Stat(tempFile); err != nil {
		t.Fatalf("setup should have created file: %v", err)
	}
	if !strings.Contains(stdoutReal.String(), "Successfully registered Vivechak") {
		t.Errorf("expected success message, got: %s", stdoutReal.String())
	}

	// Verify content
	data, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON written: %v", err)
	}
	servers, ok := parsed["mcpServers"].(map[string]any)
	if !ok || servers["vivechak"] == nil {
		t.Errorf("expected vivechak entry under mcpServers, got: %v", parsed)
	}
}

func TestRunSetupWithArgs_JSONC_PreservesUserConfig(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "zed-settings.json")
	initialJSONC := `{
		// User editor preferences
		"theme": "Nord",
		"tab_size": 4,
		"context_servers": {
			// Existing custom tool
			"custom-tool": {
				"command": "custom-binary",
				"args": ["run",],
			},
		},
	}`
	if err := os.WriteFile(tempFile, []byte(initialJSONC), 0o644); err != nil {
		t.Fatalf("failed to write initial JSONC: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := runSetupWithArgs([]string{tempFile}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("setup failed with code %d (stderr: %s)", code, stderr.String())
	}

	// Verify no backup was created since config was valid JSONC
	if _, err := os.Stat(tempFile + ".bak"); !os.IsNotExist(err) {
		t.Errorf("unexpected backup file created for valid JSONC: %s.bak", tempFile)
	}

	data, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("reading merged file: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("merged file is not valid JSON: %v", err)
	}

	if parsed["theme"] != "Nord" {
		t.Errorf("lost user theme: %v", parsed["theme"])
	}
	ctxServers, ok := parsed["context_servers"].(map[string]any)
	if !ok {
		t.Fatalf("context_servers missing or not a map: %v", parsed)
	}
	if ctxServers["custom-tool"] == nil {
		t.Errorf("lost existing custom-tool from context_servers")
	}
	if ctxServers["vivechak"] == nil {
		t.Errorf("missing vivechak in context_servers")
	}
}

func TestRunSetupWithArgs_AutoDetectWorkspace(t *testing.T) {
	tempWS := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tempWS, ".cursor"), 0o755)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempWS)
	defer func() { _ = os.Chdir(oldWd) }()

	var stdout, stderr bytes.Buffer
	code := runSetupWithArgs([]string{"--dry-run"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected 0, got %d; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Detected Cursor workspace") {
		t.Errorf("expected auto-detection of Cursor, got: %s", stdout.String())
	}
}

func TestRunSetupWithArgs_InvalidTarget(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runSetupWithArgs([]string{"nonexistent-preset"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("expected failure for invalid preset, got 0")
	}
	if !strings.Contains(stderr.String(), "unrecognized host preset or file path") {
		t.Errorf("expected error message, got: %s", stderr.String())
	}
}

func TestFormatHostName(t *testing.T) {
	tests := map[string]string{
		"cursor":          "Cursor",
		"vscode":          "VS Code",
		"code":            "VS Code",
		"claude-desktop":  "Claude Desktop",
		"claude":          "Claude Desktop",
		"claude-code":     "Claude Code CLI",
		"mcp":             "Universal Portable (.mcp.json)",
		"opencode":        "OpenCode",
		"copilot":         "GitHub Copilot",
		"void":            "Void Editor",
		"continue":        "Continue.dev",
		"goose":           "Block Goose",
		"antigravity":     "Google Antigravity",
		"agy":             "Google Antigravity",
		"zed":             "Zed",
		"kiro":            "AWS Kiro",
		"trae":            "ByteDance Trae",
		"omp":             "Oh My Pi (OMP)",
		"unknown":         "unknown",
	}
	for in, want := range tests {
		if got := formatHostName(in); got != want {
			t.Errorf("formatHostName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRunSetupWithArgs_CorruptConfigFailsFastWithoutOverwrite(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "corrupt_config.json")
	corruptContent := `{"unclosed_json": `
	if err := os.WriteFile(tempFile, []byte(corruptContent), 0o644); err != nil {
		t.Fatalf("failed to write corrupt config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := runSetupWithArgs([]string{tempFile}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for corrupt config, got %d", code)
	}
	if !strings.Contains(stderr.String(), "could not merge Vivechak into existing configuration") {
		t.Errorf("expected merge failure error message in stderr, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "will not overwrite this file") {
		t.Errorf("expected protection message in stderr, got: %s", stderr.String())
	}

	// Verify the original file was NOT overwritten
	after, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("reading file after setup: %v", err)
	}
	if string(after) != corruptContent {
		t.Errorf("corrupt file was overwritten! got: %s, want: %s", string(after), corruptContent)
	}
}

func TestResolveConfigPath_ExpandedHarnesses(t *testing.T) {
	homeDir := "/home/testuser"
	cwd := "/workspace/testproject"
	appData := "C:\\Users\\testuser\\AppData\\Roaming"

	// 1. Universal Portable .mcp.json
	p, err := resolveConfigPath("mcp", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("mcp: %v", err)
	}
	if p != filepath.Join(cwd, ".mcp.json") {
		t.Errorf("mcp got %s, want %s", p, filepath.Join(cwd, ".mcp.json"))
	}

	// 2. Claude Code (global default when no .mcp.json in cwd)
	p, err = resolveConfigPath("claude-code", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("claude-code: %v", err)
	}
	if p != filepath.Join(homeDir, ".claude.json") {
		t.Errorf("claude-code got %s, want %s", p, filepath.Join(homeDir, ".claude.json"))
	}

	// Claude Code (project when .mcp.json exists in cwd)
	tempClaudeWS := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempClaudeWS, ".mcp.json"), []byte("{}"), 0o644)
	p, err = resolveConfigPath("claude-code", homeDir, appData, tempClaudeWS, "linux")
	if err != nil {
		t.Fatalf("claude-code project: %v", err)
	}
	if p != filepath.Join(tempClaudeWS, ".mcp.json") {
		t.Errorf("claude-code project got %s, want %s", p, filepath.Join(tempClaudeWS, ".mcp.json"))
	}

	// 3. OpenCode (global default)
	p, err = resolveConfigPath("opencode", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("opencode: %v", err)
	}
	if p != filepath.Join(homeDir, ".config", "opencode", "opencode.json") {
		t.Errorf("opencode got %s, want %s", p, filepath.Join(homeDir, ".config", "opencode", "opencode.json"))
	}

	// OpenCode (workspace when opencode.json exists)
	tempOpenCodeWS := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempOpenCodeWS, "opencode.json"), []byte("{}"), 0o644)
	p, err = resolveConfigPath("opencode", homeDir, appData, tempOpenCodeWS, "linux")
	if err != nil {
		t.Fatalf("opencode workspace: %v", err)
	}
	if p != filepath.Join(tempOpenCodeWS, "opencode.json") {
		t.Errorf("opencode workspace got %s, want %s", p, filepath.Join(tempOpenCodeWS, "opencode.json"))
	}

	// 4. GitHub Copilot CLI
	p, err = resolveConfigPath("copilot", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("copilot: %v", err)
	}
	if p != filepath.Join(homeDir, ".copilot", "mcp-config.json") {
		t.Errorf("copilot got %s, want %s", p, filepath.Join(homeDir, ".copilot", "mcp-config.json"))
	}

	// Copilot CLI with COPILOT_HOME override
	tempCopilotHome := t.TempDir()
	t.Setenv("COPILOT_HOME", tempCopilotHome)
	p, err = resolveConfigPath("copilot", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("copilot with COPILOT_HOME: %v", err)
	}
	if p != filepath.Join(tempCopilotHome, "mcp-config.json") {
		t.Errorf("copilot COPILOT_HOME got %s, want %s", p, filepath.Join(tempCopilotHome, "mcp-config.json"))
	}

	// 5. Void Editor (global default)
	p, err = resolveConfigPath("void", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("void: %v", err)
	}
	if p != filepath.Join(homeDir, ".void-editor", "mcp.json") {
		t.Errorf("void got %s, want %s", p, filepath.Join(homeDir, ".void-editor", "mcp.json"))
	}

	// Void Editor (workspace when .void exists)
	tempVoidWS := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tempVoidWS, ".void"), 0o755)
	p, err = resolveConfigPath("void", homeDir, appData, tempVoidWS, "linux")
	if err != nil {
		t.Fatalf("void workspace: %v", err)
	}
	if p != filepath.Join(tempVoidWS, ".void", "mcp.json") {
		t.Errorf("void workspace got %s, want %s", p, filepath.Join(tempVoidWS, ".void", "mcp.json"))
	}

	// 6. Continue.dev (global default)
	p, err = resolveConfigPath("continue", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("continue: %v", err)
	}
	if p != filepath.Join(homeDir, ".continue", "mcpServers", "vivechak.json") {
		t.Errorf("continue got %s, want %s", p, filepath.Join(homeDir, ".continue", "mcpServers", "vivechak.json"))
	}

	// Continue.dev (workspace when .continue exists)
	tempContinueWS := t.TempDir()
	_ = os.MkdirAll(filepath.Join(tempContinueWS, ".continue"), 0o755)
	p, err = resolveConfigPath("continue", homeDir, appData, tempContinueWS, "linux")
	if err != nil {
		t.Fatalf("continue workspace: %v", err)
	}
	if p != filepath.Join(tempContinueWS, ".continue", "mcpServers", "vivechak.json") {
		t.Errorf("continue workspace got %s, want %s", p, filepath.Join(tempContinueWS, ".continue", "mcpServers", "vivechak.json"))
	}

	// 7. Block Goose (Linux/macOS)
	p, err = resolveConfigPath("goose", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("goose linux: %v", err)
	}
	if p != filepath.Join(homeDir, ".config", "goose", "config.yaml") {
		t.Errorf("goose linux got %s, want %s", p, filepath.Join(homeDir, ".config", "goose", "config.yaml"))
	}

	// Block Goose (Windows)
	p, err = resolveConfigPath("goose", homeDir, appData, cwd, "windows")
	if err != nil {
		t.Fatalf("goose windows: %v", err)
	}
	expectedGooseWin := filepath.Join(appData, "Block", "goose", "config", "config.yaml")
	if p != expectedGooseWin {
		t.Errorf("goose windows got %s, want %s", p, expectedGooseWin)
	}
}

func TestDetermineServerKey_ExpandedHarnesses(t *testing.T) {
	if got := determineServerKey("opencode", "/path/to/opencode.json", nil); got != "mcp" {
		t.Errorf("opencode serverKey got %q, want %q", got, "mcp")
	}
	if got := determineServerKey("goose", "/path/to/config.yaml", nil); got != "extensions" {
		t.Errorf("goose serverKey got %q, want %q", got, "extensions")
	}
	if got := determineServerKey("mcp", "/workspace/.mcp.json", nil); got != "mcpServers" {
		t.Errorf("mcp serverKey got %q, want %q", got, "mcpServers")
	}
	if got := determineServerKey("claude-code", "/home/.claude.json", nil); got != "mcpServers" {
		t.Errorf("claude-code serverKey got %q, want %q", got, "mcpServers")
	}
	if got := determineServerKey("copilot", "/home/.copilot/mcp-config.json", nil); got != "mcpServers" {
		t.Errorf("copilot serverKey got %q, want %q", got, "mcpServers")
	}
	if got := determineServerKey("void", "/workspace/.void/mcp.json", nil); got != "mcpServers" {
		t.Errorf("void serverKey got %q, want %q", got, "mcpServers")
	}
	if got := determineServerKey("continue", "/workspace/.continue/mcpServers/vivechak.json", nil); got != "mcpServers" {
		t.Errorf("continue serverKey got %q, want %q", got, "mcpServers")
	}
}

func TestBuildServerConfig(t *testing.T) {
	exe := "/path/to/vivechak"

	// Standard
	std := buildServerConfig("mcpServers", exe)
	if std["command"] != exe {
		t.Errorf("expected command %q, got %v", exe, std["command"])
	}

	// OpenCode (mcp)
	opencode := buildServerConfig("mcp", exe)
	if opencode["type"] != "local" {
		t.Errorf("expected type 'local', got %v", opencode["type"])
	}
	cmdList, ok := opencode["command"].([]string)
	if !ok || len(cmdList) != 2 || cmdList[0] != exe || cmdList[1] != "serve" {
		t.Errorf("expected command array [%q, 'serve'], got %v", exe, opencode["command"])
	}
	if opencode["enabled"] != true {
		t.Errorf("expected enabled true, got %v", opencode["enabled"])
	}

	// Goose (extensions)
	goose := buildServerConfig("extensions", exe)
	if goose["type"] != "stdio" {
		t.Errorf("expected type 'stdio', got %v", goose["type"])
	}
	if goose["cmd"] != exe {
		t.Errorf("expected cmd %q, got %v", exe, goose["cmd"])
	}
	if goose["enabled"] != true {
		t.Errorf("expected enabled true, got %v", goose["enabled"])
	}
}

func TestMergeGooseYAML(t *testing.T) {
	exe := "/custom/bin/vivechak"

	// 1. Empty file
	outEmpty := mergeGooseYAML(nil, exe)
	strEmpty := string(outEmpty)
	if !strings.Contains(strEmpty, "extensions:") || !strings.Contains(strEmpty, "vivechak:") || !strings.Contains(strEmpty, exe) {
		t.Fatalf("mergeGooseYAML empty failed:\n%s", strEmpty)
	}

	// 2. Existing YAML without extensions
	existingNoExt := "version: 1.0\nprovider: anthropic\n"
	outNoExt := mergeGooseYAML([]byte(existingNoExt), exe)
	strNoExt := string(outNoExt)
	if !strings.Contains(strNoExt, "provider: anthropic") || !strings.Contains(strNoExt, "extensions:") || !strings.Contains(strNoExt, exe) {
		t.Fatalf("mergeGooseYAML without extensions failed:\n%s", strNoExt)
	}

	// 3. Existing YAML with other extensions
	existingWithExt := "extensions:\n  developer:\n    enabled: true\n"
	outWithExt := mergeGooseYAML([]byte(existingWithExt), exe)
	strWithExt := string(outWithExt)
	if !strings.Contains(strWithExt, "developer:") || !strings.Contains(strWithExt, "vivechak:") || !strings.Contains(strWithExt, exe) {
		t.Fatalf("mergeGooseYAML with existing extensions failed:\n%s", strWithExt)
	}

	// 4. Existing YAML with pre-existing vivechak (update in place)
	existingWithVivechak := "extensions:\n  developer:\n    enabled: true\n  vivechak:\n    type: stdio\n    cmd: \"/old/path\"\n"
	outUpdated := mergeGooseYAML([]byte(existingWithVivechak), exe)
	strUpdated := string(outUpdated)
	if strings.Contains(strUpdated, "/old/path") {
		t.Errorf("old path was not replaced:\n%s", strUpdated)
	}
	if !strings.Contains(strUpdated, exe) || !strings.Contains(strUpdated, "developer:") {
		t.Errorf("updated YAML missing new path or sibling extension:\n%s", strUpdated)
	}
}

func TestMergeConfig_PreservesClaudeCodeState(t *testing.T) {
	existingClaude := []byte(`{
  "numStartups": 42,
  "theme": "dark",
  "allowedTools": ["bash", "read_file"],
  "mcpServers": {
    "existing-server": {
      "command": "node",
      "args": ["server.js"]
    }
  }
}`)
	exe := "/path/to/vivechak"
	merged, err := mergeConfig(existingClaude, exe, "mcpServers")
	if err != nil {
		t.Fatalf("mergeConfig failed: %v", err)
	}

	var parsed map[string]any
	dec := json.NewDecoder(bytes.NewReader(merged))
	dec.UseNumber()
	if err := dec.Decode(&parsed); err != nil {
		t.Fatalf("merged is invalid JSON: %v", err)
	}

	if num, ok := parsed["numStartups"].(json.Number); !ok || num.String() != "42" {
		t.Errorf("numStartups was corrupted: %v", parsed["numStartups"])
	}
	if parsed["theme"] != "dark" {
		t.Errorf("theme was corrupted: %v", parsed["theme"])
	}

	servers := parsed["mcpServers"].(map[string]any)
	if servers["existing-server"] == nil {
		t.Errorf("existing-server was dropped")
	}
	if servers["vivechak"] == nil {
		t.Errorf("vivechak was not added")
	}
}

func TestDetectWorkspaceHost_ExpandedIndicators(t *testing.T) {
	// 1. .mcp.json file
	dirMCP := t.TempDir()
	_ = os.WriteFile(filepath.Join(dirMCP, ".mcp.json"), []byte("{}"), 0o644)
	if got := detectWorkspaceHost(dirMCP); got != "mcp" {
		t.Errorf("expected 'mcp' for .mcp.json, got %q", got)
	}

	// 2. opencode.json file
	dirOpenCode := t.TempDir()
	_ = os.WriteFile(filepath.Join(dirOpenCode, "opencode.json"), []byte("{}"), 0o644)
	if got := detectWorkspaceHost(dirOpenCode); got != "opencode" {
		t.Errorf("expected 'opencode' for opencode.json, got %q", got)
	}

	// 3. .void directory
	dirVoid := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dirVoid, ".void"), 0o755)
	if got := detectWorkspaceHost(dirVoid); got != "void" {
		t.Errorf("expected 'void' for .void, got %q", got)
	}

	// 4. .continue directory
	dirContinue := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dirContinue, ".continue"), 0o755)
	if got := detectWorkspaceHost(dirContinue); got != "continue" {
		t.Errorf("expected 'continue' for .continue, got %q", got)
	}

	// 5. .copilot directory
	dirCopilot := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dirCopilot, ".copilot"), 0o755)
	if got := detectWorkspaceHost(dirCopilot); got != "copilot" {
		t.Errorf("expected 'copilot' for .copilot, got %q", got)
	}

	// 6. .goose directory
	dirGoose := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dirGoose, ".goose"), 0o755)
	if got := detectWorkspaceHost(dirGoose); got != "goose" {
		t.Errorf("expected 'goose' for .goose, got %q", got)
	}
}

func TestRunSetup_DryRun_All21Presets(t *testing.T) {
	presets := []string{
		"cursor", "vscode", "claude-desktop", "claude-code", "mcp", "opencode",
		"copilot", "void", "continue", "goose", "windsurf", "antigravity",
		"zed", "kiro", "trae", "omp", "openhands", "droid", "cline", "roo", "devin",
	}

	for _, p := range presets {
		var stdout, stderr bytes.Buffer
		code := runSetupWithArgs([]string{"--dry-run", p}, &stdout, &stderr)
		if code != 0 {
			t.Errorf("setup --dry-run %s failed with code %d: %s", p, code, stderr.String())
			continue
		}
		if !strings.Contains(stdout.String(), "[dry-run]") {
			t.Errorf("setup --dry-run %s missing [dry-run] marker in output: %s", p, stdout.String())
		}
	}
}
