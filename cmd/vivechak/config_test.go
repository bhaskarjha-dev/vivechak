package main

import (
	"path/filepath"
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

	// Cloud / extension hosts (no local file)
	for _, client := range []string{"chatgpt", "codex", "kiro"} {
		p, err = resolveConfigPath(client, homeDir, appData, cwd, "linux")
		if err != nil {
			t.Errorf("%s: unexpected error %v", client, err)
		}
		if p != "" {
			t.Errorf("%s: expected empty path, got %s", client, p)
		}
	}

	// Unknown client
	_, err = resolveConfigPath("invalidclient", homeDir, appData, cwd, "linux")
	if err == nil {
		t.Error("expected error for unknown client, got nil")
	}
}
