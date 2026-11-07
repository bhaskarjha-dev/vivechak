package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const verifiedPresetsList = "claude-code, claude-desktop (or claude), cursor, vscode, windsurf, zed, cline, roo, goose, agy (or antigravity), opencode"
const desktopShortcutsList = "claude-code, claude-desktop (or claude), cursor, vscode, windsurf, zed, cline, roo, goose, agy (or antigravity), opencode"

// HostPreset represents a supported desktop or harness MCP configuration preset.
type HostPreset struct {
	Name        string
	DisplayName string
	Aliases     []string
	ServerKey   string
	ResolvePath func(homeDir, appData, cwd, goos string) (string, error)
}

var hostPresets = []HostPreset{
	{
		Name:        "cursor",
		DisplayName: "Cursor",
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			if _, err := os.Stat(filepath.Join(cwd, ".cursor")); err == nil {
				return filepath.Join(cwd, ".cursor", "mcp.json"), nil
			}
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for cursor config")
			}
			return filepath.Join(homeDir, ".cursor", "mcp.json"), nil
		},
	},
	{
		Name:        "vscode",
		DisplayName: "VS Code",
		Aliases:     []string{"code"},
		ServerKey:   "servers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			return filepath.Join(cwd, ".vscode", "mcp.json"), nil
		},
	},
	{
		Name:        "claude-desktop",
		DisplayName: "Claude Desktop",
		Aliases:     []string{"claude"},
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			switch goos {
			case "windows":
				if appData == "" && homeDir != "" {
					appData = filepath.Join(homeDir, "AppData", "Roaming")
				}
				if appData == "" {
					return "", fmt.Errorf("could not determine AppData path for claude-desktop")
				}
				return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
			case "darwin":
				if homeDir == "" {
					return "", fmt.Errorf("could not determine home dir for claude-desktop")
				}
				return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
			default:
				if homeDir == "" {
					return "", fmt.Errorf("could not determine home dir for claude-desktop")
				}
				return filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json"), nil
			}
		},
	},
	{
		Name:        "antigravity",
		DisplayName: "Google Antigravity",
		Aliases:     []string{"agy"},
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			// If workspace has .agents directory or .agents/mcp_config.json, use workspace configuration
			if _, err := os.Stat(filepath.Join(cwd, ".agents")); err == nil {
				return filepath.Join(cwd, ".agents", "mcp_config.json"), nil
			}
			// Otherwise use canonical global Antigravity config in ~/.gemini/config/mcp_config.json
			// (shared across Antigravity IDE, Antigravity 2.0, and Antigravity CLI 'agy')
			if homeDir != "" {
				return filepath.Join(homeDir, ".gemini", "config", "mcp_config.json"), nil
			}
			return filepath.Join(cwd, ".agents", "mcp_config.json"), nil
		},
	},
	{
		Name:        "windsurf",
		DisplayName: "Windsurf",
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for windsurf config")
			}
			return filepath.Join(homeDir, ".codeium", "windsurf", "mcp_config.json"), nil
		},
	},
	{
		Name:        "zed",
		DisplayName: "Zed",
		ServerKey:   "context_servers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			if goos == "windows" {
				if appData == "" && homeDir != "" {
					appData = filepath.Join(homeDir, "AppData", "Roaming")
				}
				if appData == "" {
					return "", fmt.Errorf("could not determine AppData path for zed")
				}
				return filepath.Join(appData, "Zed", "settings.json"), nil
			}
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for zed config")
			}
			return filepath.Join(homeDir, ".config", "zed", "settings.json"), nil
		},
	},
	{
		Name:        "kiro",
		DisplayName: "AWS Kiro",
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			// Workspace-level Kiro configuration if .kiro directory exists
			if _, err := os.Stat(filepath.Join(cwd, ".kiro")); err == nil {
				return filepath.Join(cwd, ".kiro", "settings", "mcp.json"), nil
			}
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for kiro config")
			}
			// If legacy ~/.aws/.kiro directory exists, maintain backward compatibility
			if _, err := os.Stat(filepath.Join(homeDir, ".aws", ".kiro")); err == nil {
				return filepath.Join(homeDir, ".aws", ".kiro", "mcp.json"), nil
			}
			// Canonical official Kiro configuration path
			return filepath.Join(homeDir, ".kiro", "settings", "mcp.json"), nil
		},
	},
	{
		Name:        "trae",
		DisplayName: "ByteDance Trae",
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			// ByteDance Trae IDE: project .trae/mcp.json or global ~/.trae/mcp.json
			if _, err := os.Stat(filepath.Join(cwd, ".trae")); err == nil {
				return filepath.Join(cwd, ".trae", "mcp.json"), nil
			}
			if homeDir != "" {
				if _, err := os.Stat(filepath.Join(homeDir, ".trae")); err == nil {
					return filepath.Join(homeDir, ".trae", "mcp.json"), nil
				}
			}
			return filepath.Join(cwd, ".trae", "mcp.json"), nil
		},
	},
	{
		Name:        "omp",
		DisplayName: "Oh My Pi (OMP)",
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			// Oh My Pi (OMP): workspace .omp/mcp.json or global ~/.omp/agent/mcp.json
			if _, err := os.Stat(filepath.Join(cwd, ".omp")); err == nil {
				return filepath.Join(cwd, ".omp", "mcp.json"), nil
			}
			if homeDir != "" {
				return filepath.Join(homeDir, ".omp", "agent", "mcp.json"), nil
			}
			return filepath.Join(cwd, ".omp", "mcp.json"), nil
		},
	},
	{
		Name:        "openhands",
		DisplayName: "OpenHands",
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for openhands config")
			}
			return filepath.Join(homeDir, ".openhands", "mcp.json"), nil
		},
	},
	{
		Name:        "droid",
		DisplayName: "Factory Droid",
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			// Factory Droid: workspace .factory/mcp.json or global ~/.factory/mcp.json
			if _, err := os.Stat(filepath.Join(cwd, ".factory")); err == nil {
				return filepath.Join(cwd, ".factory", "mcp.json"), nil
			}
			if homeDir != "" {
				return filepath.Join(homeDir, ".factory", "mcp.json"), nil
			}
			return filepath.Join(cwd, ".factory", "mcp.json"), nil
		},
	},
	{
		Name:        "cline",
		DisplayName: "Cline",
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			// Cline VS Code extension: workspace .cline/mcp.json or global extension storage
			if _, err := os.Stat(filepath.Join(cwd, ".cline")); err == nil {
				return filepath.Join(cwd, ".cline", "mcp.json"), nil
			}
			return resolveVSCodeStoragePath("saoudrizwan.claude-dev", "cline_mcp_settings.json", homeDir, appData, goos)
		},
	},
	{
		Name:        "roo",
		DisplayName: "Roo Code",
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			// Roo Code VS Code extension: workspace .roo/mcp.json or global extension storage
			if _, err := os.Stat(filepath.Join(cwd, ".roo")); err == nil {
				return filepath.Join(cwd, ".roo", "mcp.json"), nil
			}
			return resolveVSCodeStoragePath("rooveterinaryinc.roo-cline", "mcp_settings.json", homeDir, appData, goos)
		},
	},
	{
		Name:        "devin",
		DisplayName: "Cognition Devin",
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			// Cognition Devin CLI / local config: .devin/mcp_config.local.json
			if _, err := os.Stat(filepath.Join(cwd, ".devin")); err == nil {
				return filepath.Join(cwd, ".devin", "mcp_config.local.json"), nil
			}
			if homeDir != "" {
				if _, err := os.Stat(filepath.Join(homeDir, ".devin")); err == nil {
					return filepath.Join(homeDir, ".devin", "mcp_config.local.json"), nil
				}
			}
			return filepath.Join(cwd, ".devin", "mcp_config.local.json"), nil
		},
	},
	{
		Name:        "mcp",
		DisplayName: "Universal Portable (.mcp.json)",
		Aliases:     []string{"portable", "universal"},
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			return filepath.Join(cwd, ".mcp.json"), nil
		},
	},
	{
		Name:        "claude-code",
		DisplayName: "Claude Code CLI",
		Aliases:     []string{"claudecode"},
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			// If repository has .mcp.json, prioritize project config
			if _, err := os.Stat(filepath.Join(cwd, ".mcp.json")); err == nil {
				return filepath.Join(cwd, ".mcp.json"), nil
			}
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for claude-code")
			}
			return filepath.Join(homeDir, ".claude.json"), nil
		},
	},
	{
		Name:        "opencode",
		DisplayName: "OpenCode",
		Aliases:     []string{"open-code"},
		ServerKey:   "mcp",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			if _, err := os.Stat(filepath.Join(cwd, "opencode.json")); err == nil {
				return filepath.Join(cwd, "opencode.json"), nil
			}
			if _, err := os.Stat(filepath.Join(cwd, ".opencode")); err == nil {
				return filepath.Join(cwd, ".opencode", "opencode.json"), nil
			}
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for opencode")
			}
			return filepath.Join(homeDir, ".config", "opencode", "opencode.json"), nil
		},
	},
	{
		Name:        "copilot",
		DisplayName: "GitHub Copilot",
		Aliases:     []string{"github-copilot"},
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			if envHome := os.Getenv("COPILOT_HOME"); envHome != "" {
				return filepath.Join(envHome, "mcp-config.json"), nil
			}
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for copilot")
			}
			return filepath.Join(homeDir, ".copilot", "mcp-config.json"), nil
		},
	},
	{
		Name:        "void",
		DisplayName: "Void Editor",
		Aliases:     []string{"void-editor"},
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			if _, err := os.Stat(filepath.Join(cwd, ".void")); err == nil {
				return filepath.Join(cwd, ".void", "mcp.json"), nil
			}
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for void")
			}
			return filepath.Join(homeDir, ".void-editor", "mcp.json"), nil
		},
	},
	{
		Name:        "continue",
		DisplayName: "Continue.dev",
		Aliases:     []string{"continuedev"},
		ServerKey:   "mcpServers",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			if _, err := os.Stat(filepath.Join(cwd, ".continue")); err == nil {
				return filepath.Join(cwd, ".continue", "mcpServers", "vivechak.json"), nil
			}
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for continue")
			}
			return filepath.Join(homeDir, ".continue", "mcpServers", "vivechak.json"), nil
		},
	},
	{
		Name:        "goose",
		DisplayName: "Block Goose",
		Aliases:     []string{"block-goose"},
		ServerKey:   "extensions",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			if goos == "windows" {
				if appData == "" && homeDir != "" {
					appData = filepath.Join(homeDir, "AppData", "Roaming")
				}
				if appData == "" {
					return "", fmt.Errorf("could not determine AppData path for goose")
				}
				return filepath.Join(appData, "Block", "goose", "config", "config.yaml"), nil
			}
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for goose")
			}
			return filepath.Join(homeDir, ".config", "goose", "config.yaml"), nil
		},
	},
}

// findPreset locates a HostPreset by name or alias (case-insensitive).
func findPreset(name string) *HostPreset {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == "" {
		return nil
	}
	for i := range hostPresets {
		if strings.EqualFold(hostPresets[i].Name, lower) {
			return &hostPresets[i]
		}
		for _, alias := range hostPresets[i].Aliases {
			if strings.EqualFold(alias, lower) {
				return &hostPresets[i]
			}
		}
	}
	return nil
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

// resolveVSCodeStoragePath resolves the globalStorage configuration path for VS Code extension agents.
func resolveVSCodeStoragePath(extID, settingsFileName, homeDir, appData, goos string) (string, error) {
	if settingsFileName == "" {
		settingsFileName = "mcp_settings.json"
	}
	var baseDir string
	switch goos {
	case "windows":
		if appData == "" && homeDir != "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		if appData == "" {
			return "", fmt.Errorf("could not determine AppData path for VS Code extension storage")
		}
		baseDir = filepath.Join(appData, "Code", "User")
	case "darwin":
		if homeDir == "" {
			return "", fmt.Errorf("could not determine home dir for VS Code extension storage")
		}
		baseDir = filepath.Join(homeDir, "Library", "Application Support", "Code", "User")
	default:
		if homeDir == "" {
			return "", fmt.Errorf("could not determine home dir for VS Code extension storage")
		}
		baseDir = filepath.Join(homeDir, ".config", "Code", "User")
	}
	return filepath.Join(baseDir, "globalStorage", extID, "settings", settingsFileName), nil
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
		{target: ".mcp.json", host: "mcp", isFile: true},
		{target: "opencode.json", host: "opencode", isFile: true},
		{target: ".opencode", host: "opencode"},
		{target: ".void", host: "void"},
		{target: ".continue", host: "continue"},
		{target: ".goose", host: "goose"},
		{target: ".copilot", host: "copilot"},
		{target: ".trae", host: "trae"},
		{target: ".omp", host: "omp"},
		{target: ".factory", host: "droid"},
		{target: ".kiro", host: "kiro"},
		{target: ".devin", host: "devin"},
		{target: ".cline", host: "cline"},
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
