package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HostPreset represents a verified desktop or harness MCP configuration preset.
type HostPreset struct {
	Name          string
	DisplayName   string
	DocShortcut   string
	Aliases       []string
	ServerKey     string
	DocURL        string
	ConfigSummary string
	ResolvePath   func(homeDir, appData, cwd, goos string) (string, error)
}

// Shortcut returns the concise CLI shortcut used in setup documentation.
func (p HostPreset) Shortcut() string {
	if p.DocShortcut != "" {
		return p.DocShortcut
	}
	return p.Name
}

var canonicalPresets = []HostPreset{
	{
		// https://cursor.com/docs/context/mcp: project .cursor/mcp.json, global ~/.cursor/mcp.json; key mcpServers.
		Name:          "cursor",
		DisplayName:   "Cursor",
		ServerKey:     "mcpServers",
		DocURL:        "https://cursor.com/docs/context/mcp",
		ConfigSummary: "`.cursor/mcp.json` or `~/.cursor/mcp.json`",
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
		// https://code.visualstudio.com/docs/copilot/reference/mcp-configuration: workspace .vscode/mcp.json; key servers.
		Name:          "vscode",
		DisplayName:   "VS Code",
		Aliases:       []string{"code"},
		ServerKey:     "servers",
		DocURL:        "https://code.visualstudio.com/docs/copilot/reference/mcp-configuration",
		ConfigSummary: "`.vscode/mcp.json` (legacy file VS Code documents as deprecated; removal planned for February 2027)",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			return filepath.Join(cwd, ".vscode", "mcp.json"), nil
		},
	},
	{
		// https://modelcontextprotocol.io/docs/develop/connect-local-servers: OS-specific claude_desktop_config.json; key mcpServers.
		Name:          "claude-desktop",
		DisplayName:   "Claude Desktop",
		DocShortcut:   "claude",
		Aliases:       []string{"claude"},
		ServerKey:     "mcpServers",
		DocURL:        "https://modelcontextprotocol.io/docs/develop/connect-local-servers",
		ConfigSummary: "OS-specific `claude_desktop_config.json`",
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
		// https://antigravity.google/docs/mcp: global ~/.gemini/config/mcp_config.json, workspace .agents/mcp_config.json; key mcpServers.
		Name:          "antigravity",
		DisplayName:   "Google Antigravity",
		DocShortcut:   "agy",
		Aliases:       []string{"agy"},
		ServerKey:     "mcpServers",
		DocURL:        "https://antigravity.google/docs/mcp",
		ConfigSummary: "`~/.gemini/config/mcp_config.json` or `.agents/mcp_config.json`",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			if _, err := os.Stat(filepath.Join(cwd, ".agents")); err == nil {
				return filepath.Join(cwd, ".agents", "mcp_config.json"), nil
			}
			if homeDir != "" {
				return filepath.Join(homeDir, ".gemini", "config", "mcp_config.json"), nil
			}
			return filepath.Join(cwd, ".agents", "mcp_config.json"), nil
		},
	},
	{
		// https://docs.devin.ai/windsurf/plugins/cascade/mcp: ~/.codeium/windsurf/mcp_config.json; key mcpServers.
		Name:          "windsurf",
		DisplayName:   "Windsurf",
		ServerKey:     "mcpServers",
		DocURL:        "https://docs.devin.ai/windsurf/plugins/cascade/mcp",
		ConfigSummary: "`~/.codeium/windsurf/mcp_config.json` (read only when `chat.mcp.discovery.enabled` includes the windsurf source)",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for windsurf config")
			}
			return filepath.Join(homeDir, ".codeium", "windsurf", "mcp_config.json"), nil
		},
	},
	{
		// https://zed.dev/docs/configuring-zed and https://zed.dev/docs/ai/mcp: key context_servers.
		Name:          "zed",
		DisplayName:   "Zed",
		ServerKey:     "context_servers",
		DocURL:        "https://zed.dev/docs/ai/mcp",
		ConfigSummary: "`~/.config/zed/settings.json` (or AppData on Windows)",
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
		// https://docs.roocode.com/features/mcp/using-mcp-in-roo: project-level .roo/mcp.json; key mcpServers.
		Name:          "roo",
		DisplayName:   "Roo Code",
		ServerKey:     "mcpServers",
		DocURL:        "https://docs.roocode.com/features/mcp/using-mcp-in-roo",
		ConfigSummary: "`.roo/mcp.json` (project)",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
			return filepath.Join(cwd, ".roo", "mcp.json"), nil
		},
	},
	{
		// https://code.claude.com/docs/en/mcp: project .mcp.json, user ~/.claude.json; key mcpServers.
		Name:          "claude-code",
		DisplayName:   "Claude Code CLI",
		Aliases:       []string{"claudecode"},
		ServerKey:     "mcpServers",
		DocURL:        "https://code.claude.com/docs/en/mcp",
		ConfigSummary: "`.mcp.json` or `~/.claude.json`",
		ResolvePath: func(homeDir, appData, cwd, goos string) (string, error) {
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
		// https://opencode.ai/docs/config/: opencode.json, .opencode/opencode.json, ~/.config/opencode/opencode.json; key mcp.
		Name:          "opencode",
		DisplayName:   "OpenCode",
		Aliases:       []string{"open-code"},
		ServerKey:     "mcp",
		DocURL:        "https://opencode.ai/docs/config/",
		ConfigSummary: "`opencode.json`, `.opencode/opencode.json` or `~/.config/opencode/opencode.json`",
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
		// https://goose-docs.ai/docs/guides/config-files/: ~/.config/goose/config.yaml (macOS/Linux), %APPDATA%\Block\goose\config\config.yaml (Windows); key extensions.
		Name:          "goose",
		DisplayName:   "Block Goose",
		Aliases:       []string{"block-goose"},
		ServerKey:     "extensions",
		DocURL:        "https://goose-docs.ai/docs/guides/config-files/",
		ConfigSummary: "`~/.config/goose/config.yaml` (YAML format)",
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

// All returns a copy of the canonical verified host presets slice.
func All() []HostPreset {
	out := make([]HostPreset, len(canonicalPresets))
	copy(out, canonicalPresets)
	return out
}

// Find locates a HostPreset by name or alias (case-insensitive).
func Find(name string) *HostPreset {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == "" {
		return nil
	}
	for i := range canonicalPresets {
		if strings.EqualFold(canonicalPresets[i].Name, lower) {
			p := canonicalPresets[i]
			return &p
		}
		for _, alias := range canonicalPresets[i].Aliases {
			if strings.EqualFold(alias, lower) {
				p := canonicalPresets[i]
				return &p
			}
		}
	}
	return nil
}
