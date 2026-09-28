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

	// Kiro
	p, err = resolveConfigPath("kiro", homeDir, appData, cwd, "linux")
	if err != nil {
		t.Fatalf("kiro: %v", err)
	}
	expectedKiro := filepath.Join(homeDir, ".aws", ".kiro", "mcp.json")
	if p != expectedKiro {
		t.Errorf("kiro path: got %s, want %s", p, expectedKiro)
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
	valid := []string{"cursor", "vscode", "claude-desktop", "windsurf", "antigravity", "zed", "kiro"}
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

	presets := []string{"cursor", "vscode", "claude-desktop", "windsurf", "antigravity", "zed", "kiro"}
	for _, p := range presets {
		pathFromPreset, err1 := resolveConfigPath(p, home, appData, cwd, "linux")
		if err1 != nil {
			t.Fatalf("resolveConfigPath(%q) failed: %v", p, err1)
		}
		if pathFromPreset == "" {
			t.Errorf("empty path for preset %q", p)
		}
	}
}


