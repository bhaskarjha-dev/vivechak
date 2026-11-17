package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bhaskarjha-dev/vivechak/internal/hosts"
)

const verifiedPresetsList = "claude-code, claude-desktop (or claude), cursor, vscode, windsurf, zed, roo, goose, agy (or antigravity), opencode"
const desktopShortcutsList = "claude-code, claude-desktop (or claude), cursor, vscode, windsurf, zed, roo, goose, agy (or antigravity), opencode"

// HostPreset represents a supported desktop or harness MCP configuration preset.
// Canonical preset definitions live in internal/hosts/presets.go (shared with tools/docgen):
// - cursor: https://cursor.com/docs/context/mcp (.cursor/mcp.json or ~/.cursor/mcp.json; key mcpServers)
// - vscode: https://code.visualstudio.com/docs/copilot/reference/mcp-configuration (.vscode/mcp.json; key servers)
// - claude-desktop: https://modelcontextprotocol.io/docs/develop/connect-local-servers (claude_desktop_config.json; key mcpServers)
// - antigravity: https://antigravity.google/docs/mcp (~/.gemini/config/mcp_config.json or .agents/mcp_config.json; key mcpServers)
// - windsurf: https://docs.devin.ai/windsurf/plugins/cascade/mcp (~/.codeium/windsurf/mcp_config.json; key mcpServers)
// - zed: https://zed.dev/docs/configuring-zed and https://zed.dev/docs/ai/mcp (settings.json; key context_servers)
// - roo: https://docs.roocode.com/features/mcp/using-mcp-in-roo (.roo/mcp.json; key mcpServers)
// - claude-code: https://code.claude.com/docs/en/mcp (.mcp.json or ~/.claude.json; key mcpServers)
// - opencode: https://opencode.ai/docs/config/ (opencode.json, .opencode/opencode.json, or ~/.config/opencode/opencode.json; key mcp)
// - goose: https://goose-docs.ai/docs/guides/config-files/ (~/.config/goose/config.yaml; key extensions)
type HostPreset = hosts.HostPreset

// hostPresets lists only hosts whose MCP config file and key are verified against the
// host's own documentation; every other host is configured with `vck setup <path>`.
// TestPresetFixtures asserts each file and key.
var hostPresets = hosts.All()

// findPreset locates a HostPreset by name or alias (case-insensitive).
func findPreset(name string) *HostPreset {
	return hosts.Find(name)
}

// resolveConfigPath returns the destination configuration file path for the host client.
func resolveConfigPath(host, homeDir, appData, cwd, goos string) (string, error) {
	preset := findPreset(host)
	if preset == nil {
		return "", fmt.Errorf("unknown desktop preset: %s (available shortcuts: %s)", host, desktopShortcutsList)
	}
	return preset.ResolvePath(homeDir, appData, cwd, goos)
}

// isSupportedPreset checks whether the given host name is a recognized desktop preset.
func isSupportedPreset(host string) bool {
	return findPreset(host) != nil
}

// detectWorkspaceHost checks the current working directory for known workspace indicators.
func detectWorkspaceHost(cwd string) string {
	if cwd == "" {
		return ""
	}
	checks := []struct {
		target string
		host   string
		isFile bool
	}{
		{target: ".cursor", host: "cursor"},
		{target: ".agents", host: "antigravity"},
		{target: ".vscode", host: "vscode"},
		{target: ".mcp.json", host: "claude-code", isFile: true},
		{target: "opencode.json", host: "opencode", isFile: true},
		{target: ".opencode", host: "opencode"},
		{target: ".goose", host: "goose"},
		{target: ".roo", host: "roo"},
	}
	for _, c := range checks {
		fi, err := os.Stat(filepath.Join(cwd, c.target))
		if err == nil {
			if c.isFile && !fi.IsDir() {
				return c.host
			}
			if !c.isFile && fi.IsDir() {
				return c.host
			}
		}
	}
	return ""
}

// formatHostName returns a human-friendly name for a host preset.
func formatHostName(h string) string {
	if p := findPreset(h); p != nil {
		return p.DisplayName
	}
	return h
}
