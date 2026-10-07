package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const desktopShortcutsList = "cursor, vscode, claude-desktop (or claude), windsurf, antigravity (or agy), zed, kiro, trae, omp, openhands, droid, cline, roo, devin"

func runMCPConfigWithArgs(args []string, stdout, stderr io.Writer) int {
	var host string
	var write bool
	var customPath string

	for i := 0; i < len(args); i++ {
		if args[i] == "--help" || args[i] == "-h" {
			fmt.Fprintln(stdout, "Usage: vivechak mcp-config [--path <filepath>] [--preset <shortcut>] [--write]  (or: vck mcp-config)")
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "Outputs universal Model Context Protocol (MCP) JSON configuration for Vivechak.")
			fmt.Fprintln(stdout, "Compatible with any MCP-compliant AI tool, agent harness, IDE, or runtime.")
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "Flags:")
			fmt.Fprintln(stdout, "  --write                 Write and merge configuration directly to target file")
			fmt.Fprintln(stdout, "  --path, --file <path>   Write to an arbitrary configuration file path (universal)")
			fmt.Fprintln(stdout, "  --preset <shortcut>     Optional desktop path shortcut (e.g. cursor, vscode, zed)")
			fmt.Fprintln(stdout, "  --client <shortcut>     Alias for --preset (backward compatible)")
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "Desktop Path Shortcuts:")
			fmt.Fprintf(stdout, "  %s\n", desktopShortcutsList)
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "Examples:")
			fmt.Fprintln(stdout, "  vck setup cursor                                        # Recommended: 1-second host setup")
			fmt.Fprintln(stdout, "  vck mcp-config                                          # Universal MCP JSON to stdout")
			fmt.Fprintln(stdout, "  vck mcp-config --path ~/.omp/agent/mcp.json --write     # Write directly to any agent config")
			fmt.Fprintln(stdout, "  vck mcp-config --preset cursor --write                  # Desktop shortcut for Cursor")
			fmt.Fprintln(stdout, "  vck mcp-config --preset agy --write                     # Shortcut for Antigravity (IDE, 2.0, CLI)")
			return 0
		} else if (args[i] == "--preset" || args[i] == "--client") && i+1 < len(args) {
			host = args[i+1]
			i++
		} else if (args[i] == "--path" || args[i] == "--file") && i+1 < len(args) {
			customPath = args[i+1]
			i++
		} else if args[i] == "--write" {
			write = true
		}
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(stderr, "failed to get executable path: %v\n", err)
		return 1
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to make path absolute: %v\n", err)
		return 1
	}

	if host != "" && !isSupportedPreset(host) {
		fmt.Fprintf(stderr, "Error: unknown desktop preset: %s (available shortcuts: %s)\n", host, desktopShortcutsList)
		return 1
	}

	serverKey := "mcpServers"
	switch host {
	case "vscode", "code":
		serverKey = "servers"
	case "zed":
		serverKey = "context_servers"
	}

	configObj := map[string]any{
		serverKey: map[string]any{
			"vivechak": map[string]any{
				"command": exePath,
				"args":    []string{"serve"},
			},
		},
	}

	// Without --write, print MCP config JSON compatible with the requested client or universal
	if !write {
		b, _ := json.MarshalIndent(configObj, "", "  ")
		fmt.Fprintln(stdout, string(b))
		return 0
	}

	// With --write, determine destination config path
	var configPath string
	cwd, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	appData := os.Getenv("APPDATA")

	if customPath != "" {
		configPath = customPath
	} else if host != "" {
		resolvedPath, err := resolveConfigPath(host, homeDir, appData, cwd, runtime.GOOS)
		if err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
		configPath = resolvedPath
	} else {
		fmt.Fprintln(stderr, "Error: --write requires a target file location.")
		fmt.Fprintln(stderr, "Usage: vivechak mcp-config --path <config_file_path> --write")
		fmt.Fprintln(stderr, "   or: vivechak mcp-config --preset <shortcut> --write")
		fmt.Fprintf(stderr, "Available desktop shortcuts: %s\n", desktopShortcutsList)
		return 1
	}

	// Determine serverKey based on host and target path
	targetKey := determineServerKey(host, configPath, nil)

	// merge with existing
	var out []byte
	b, err := os.ReadFile(configPath)
	if err == nil {
		// If existing file is valid JSON/JSONC, check its existing key
		if len(bytes.TrimSpace(b)) > 0 {
			var existing map[string]any
			cleaned := stripJSONComments(b)
			dec := json.NewDecoder(bytes.NewReader(cleaned))
			if dec.Decode(&existing) == nil {
				targetKey = determineServerKey(host, configPath, existing)
			}
		}

		merged, mergeErr := mergeConfig(b, exePath, targetKey)
		if mergeErr != nil {
			fmt.Fprintf(stderr, "Warning: %v\n", mergeErr)
			fmt.Fprintf(stderr, "Creating backup at %s.bak and writing fresh config\n", configPath)
			if backupErr := os.WriteFile(configPath+".bak", b, 0644); backupErr != nil {
				fmt.Fprintf(stderr, "Warning: could not create backup: %v\n", backupErr)
			}
			fresh, freshErr := mergeConfig(nil, exePath, targetKey)
			if freshErr != nil {
				fmt.Fprintf(stderr, "failed to create fresh config: %v\n", freshErr)
				return 1
			}
			out = fresh
		} else {
			out = merged
		}
	} else {
		fresh, freshErr := mergeConfig(nil, exePath, targetKey)
		if freshErr != nil {
			fmt.Fprintf(stderr, "failed to create fresh config: %v\n", freshErr)
			return 1
		}
		out = fresh
	}

	if err := writeConfigFile(configPath, out); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}

	fmt.Fprintf(stderr, "Successfully wrote config to %s\n", configPath)
	return 0
}

// determineServerKey returns the appropriate top-level JSON key ("servers", "context_servers", or "mcpServers").
func determineServerKey(host, configPath string, existing map[string]any) string {
	if host == "vscode" || host == "code" {
		if existing != nil && existing["mcpServers"] != nil && existing["servers"] == nil {
			return "mcpServers"
		}
		return "servers"
	}
	if host == "zed" {
		return "context_servers"
	}

	if existing != nil {
		if existing["servers"] != nil && existing["mcpServers"] == nil {
			return "servers"
		}
		if existing["context_servers"] != nil && existing["mcpServers"] == nil {
			return "context_servers"
		}
	}

	clean := filepath.Clean(configPath)
	if filepath.Base(filepath.Dir(clean)) == ".vscode" || (filepath.Base(clean) == "mcp.json" && strings.Contains(clean, ".vscode")) {
		return "servers"
	}
	for d := filepath.Dir(clean); d != "" && d != "." && d != filepath.Dir(d); d = filepath.Dir(d) {
		base := strings.ToLower(filepath.Base(d))
		if base == "zed" || base == ".zed" {
			return "context_servers"
		}
	}

	return "mcpServers"
}

// stripJSONComments removes C-style single-line (//) and multi-line (/* */) comments
// as well as trailing commas before closing braces/brackets, making JSONC valid standard JSON.
// It preserves whitespace and newlines so character offsets and line structure are retained.
func stripJSONComments(data []byte) []byte {
	out := make([]byte, len(data))
	copy(out, data)

	inString := false
	escaped := false
	n := len(out)

	for i := 0; i < n; i++ {
		c := out[i]

		if inString {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
			continue
		}

		if c == '"' {
			inString = true
			continue
		}

		// Line comment //
		if c == '/' && i+1 < n && out[i+1] == '/' {
			out[i] = ' '
			out[i+1] = ' '
			i += 2
			for i < n && out[i] != '\n' && out[i] != '\r' {
				out[i] = ' '
				i++
			}
			i--
			continue
		}

		// Block comment /* ... */
		if c == '/' && i+1 < n && out[i+1] == '*' {
			out[i] = ' '
			out[i+1] = ' '
			i += 2
			for i < n {
				if out[i] == '*' && i+1 < n && out[i+1] == '/' {
					out[i] = ' '
					out[i+1] = ' '
					i++
					break
				}
				if out[i] != '\n' && out[i] != '\r' {
					out[i] = ' '
				}
				i++
			}
			continue
		}
	}

	// Pass 2: Trailing commas before } or ]
	inString = false
	escaped = false
	for i := 0; i < n; i++ {
		c := out[i]

		if inString {
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				inString = false
			}
			continue
		}

		if c == '"' {
			inString = true
			continue
		}

		if c == ',' {
			j := i + 1
			for j < n && (out[j] == ' ' || out[j] == '\t' || out[j] == '\n' || out[j] == '\r') {
				j++
			}
			if j < n && (out[j] == '}' || out[j] == ']') {
				out[i] = ' '
			}
		}
	}

	return out
}

// mergeConfig merges the vivechak MCP server entry into existing config JSON,
// preserving number fidelity with json.Number.
func mergeConfig(existingBytes []byte, exePath string, preferredKey ...string) ([]byte, error) {
	var existing map[string]any
	if len(bytes.TrimSpace(existingBytes)) > 0 {
		cleaned := stripJSONComments(existingBytes)
		dec := json.NewDecoder(bytes.NewReader(cleaned))
		dec.UseNumber()
		if err := dec.Decode(&existing); err != nil {
			return nil, fmt.Errorf("existing config is not valid JSON: %w", err)
		}
	}
	if existing == nil {
		existing = make(map[string]any)
	}

	key := "mcpServers"
	if len(preferredKey) > 0 && preferredKey[0] != "" {
		key = preferredKey[0]
	} else if existing != nil {
		if existing["servers"] != nil && existing["mcpServers"] == nil {
			key = "servers"
		} else if existing["context_servers"] != nil && existing["mcpServers"] == nil {
			key = "context_servers"
		}
	}

	if existing[key] == nil {
		existing[key] = make(map[string]any)
	}

	servers, ok := existing[key].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s in config is not a JSON object", key)
	}

	servers["vivechak"] = map[string]any{
		"command": exePath,
		"args":    []string{"serve"},
	}

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling merged config: %w", err)
	}
	out = append(out, '\n')
	return out, nil
}

// writeConfigFile atomically writes config content to configPath.
func writeConfigFile(configPath string, out []byte) (retErr error) {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tmpFile := configPath + ".tmp"
	f, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	defer func() {
		if retErr != nil {
			_ = os.Remove(tmpFile)
		}
	}()

	if _, err := f.Write(out); err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	err = os.Rename(tmpFile, configPath)
	if err != nil && runtime.GOOS == "windows" {
		backoff := 50 * time.Millisecond
		for attempt := 0; attempt < 3; attempt++ {
			time.Sleep(backoff)
			backoff *= 2
			if err = os.Rename(tmpFile, configPath); err == nil {
				break
			}
		}
	}
	if err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}

// resolveConfigPath returns the destination configuration file path for the host client.
func resolveConfigPath(host, homeDir, appData, cwd, goos string) (string, error) {
	switch host {
	case "cursor":
		if _, err := os.Stat(filepath.Join(cwd, ".cursor")); err == nil {
			return filepath.Join(cwd, ".cursor", "mcp.json"), nil
		}
		if homeDir == "" {
			return "", fmt.Errorf("could not determine home dir for cursor config")
		}
		return filepath.Join(homeDir, ".cursor", "mcp.json"), nil
	case "vscode", "code":
		return filepath.Join(cwd, ".vscode", "mcp.json"), nil
	case "claude-desktop", "claude":
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
		default: // linux and others
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for claude-desktop")
			}
			return filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json"), nil
		}
	case "antigravity", "agy":
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
	case "windsurf":
		if homeDir == "" {
			return "", fmt.Errorf("could not determine home dir for windsurf config")
		}
		return filepath.Join(homeDir, ".codeium", "windsurf", "mcp_config.json"), nil
	case "zed":
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
	case "kiro":
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
	case "trae":
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
	case "omp":
		// Oh My Pi (OMP): workspace .omp/mcp.json or global ~/.omp/agent/mcp.json
		if _, err := os.Stat(filepath.Join(cwd, ".omp")); err == nil {
			return filepath.Join(cwd, ".omp", "mcp.json"), nil
		}
		if homeDir != "" {
			return filepath.Join(homeDir, ".omp", "agent", "mcp.json"), nil
		}
		return filepath.Join(cwd, ".omp", "mcp.json"), nil
	case "openhands":
		if homeDir == "" {
			return "", fmt.Errorf("could not determine home dir for openhands config")
		}
		return filepath.Join(homeDir, ".openhands", "mcp.json"), nil
	case "droid":
		// Factory Droid: workspace .factory/mcp.json or global ~/.factory/mcp.json
		if _, err := os.Stat(filepath.Join(cwd, ".factory")); err == nil {
			return filepath.Join(cwd, ".factory", "mcp.json"), nil
		}
		if homeDir != "" {
			return filepath.Join(homeDir, ".factory", "mcp.json"), nil
		}
		return filepath.Join(cwd, ".factory", "mcp.json"), nil
	case "cline":
		// Cline VS Code extension: workspace .cline/mcp.json or global extension storage
		if _, err := os.Stat(filepath.Join(cwd, ".cline")); err == nil {
			return filepath.Join(cwd, ".cline", "mcp.json"), nil
		}
		return resolveVSCodeStoragePath("saoudrizwan.claude-dev", homeDir, appData, goos)
	case "roo":
		// Roo Code VS Code extension: workspace .roo/mcp.json or global extension storage
		if _, err := os.Stat(filepath.Join(cwd, ".roo")); err == nil {
			return filepath.Join(cwd, ".roo", "mcp.json"), nil
		}
		return resolveVSCodeStoragePath("rooveterinaryinc.roo-cline", homeDir, appData, goos)
	case "devin":
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
	default:
		return "", fmt.Errorf("unknown desktop preset: %s (available shortcuts: %s)", host, desktopShortcutsList)
	}
}

// isSupportedPreset checks whether the given host name is a recognized desktop preset.
func isSupportedPreset(host string) bool {
	switch host {
	case "cursor", "vscode", "code", "claude-desktop", "claude", "windsurf", "antigravity", "agy", "zed", "kiro",
		"trae", "omp", "openhands", "droid", "cline", "roo", "devin":
		return true
	default:
		return false
	}
}

// resolveVSCodeStoragePath resolves the globalStorage configuration path for VS Code extension agents.
func resolveVSCodeStoragePath(extID, homeDir, appData, goos string) (string, error) {
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
	return filepath.Join(baseDir, "globalStorage", extID, "settings", "cline_mcp_settings.json"), nil
}

// runSetupWithArgs handles the "setup" (and "install") subcommand.
// Usage: vivechak setup [host|filepath] [--dry-run]
func runSetupWithArgs(args []string, stdout, stderr io.Writer) int {
	var target string
	var dryRun bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--help" || arg == "-h" {
			printSetupUsage(stdout)
			return 0
		} else if arg == "--dry-run" || arg == "-n" {
			dryRun = true
		} else if !strings.HasPrefix(arg, "-") && target == "" {
			target = arg
		}
	}

	cwd, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	appData := os.Getenv("APPDATA")

	// If no target provided, attempt smart workspace auto-detection
	if target == "" {
		detectedHost := detectWorkspaceHost(cwd)
		if detectedHost != "" {
			fmt.Fprintf(stdout, "Detected %s workspace in current directory.\n", formatHostName(detectedHost))
			target = detectedHost
		} else {
			// No target and no workspace detected: print concise guidance
			printSetupUsage(stdout)
			return 0
		}
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(stderr, "failed to get executable path: %v\n", err)
		return 1
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to make path absolute: %v\n", err)
		return 1
	}

	var configPath string
	var hostName string

	if isSupportedPreset(target) {
		hostName = formatHostName(target)
		resolvedPath, err := resolveConfigPath(target, homeDir, appData, cwd, runtime.GOOS)
		if err != nil {
			fmt.Fprintf(stderr, "Error resolving path for %s: %v\n", target, err)
			return 1
		}
		configPath = resolvedPath
	} else if strings.Contains(target, "/") || strings.Contains(target, "\\") || strings.HasSuffix(target, ".json") || strings.HasSuffix(target, ".jsonc") {
		if fi, err := os.Stat(target); err == nil && fi.IsDir() {
			configPath = filepath.Join(target, "mcp.json")
		} else {
			configPath = target
		}
		hostName = filepath.Base(configPath)
	} else {
		fmt.Fprintf(stderr, "Error: unrecognized host preset or file path: %q\n", target)
		fmt.Fprintf(stderr, "Available presets: %s\n", desktopShortcutsList)
		fmt.Fprintf(stderr, "Or specify a direct configuration file path: vck setup ./mcp.json\n")
		return 1
	}

	targetKey := determineServerKey(target, configPath, nil)

	var out []byte
	b, err := os.ReadFile(configPath)
	if err == nil {
		if len(bytes.TrimSpace(b)) > 0 {
			var existing map[string]any
			cleaned := stripJSONComments(b)
			dec := json.NewDecoder(bytes.NewReader(cleaned))
			if dec.Decode(&existing) == nil {
				targetKey = determineServerKey(target, configPath, existing)
			}
		}
		merged, mergeErr := mergeConfig(b, exePath, targetKey)
		if mergeErr != nil {
			fmt.Fprintf(stderr, "Error: could not merge Vivechak into existing configuration at %s: %v\n", configPath, mergeErr)
			fmt.Fprintf(stderr, "To protect your existing settings, Vivechak will not overwrite this file.\n")
			fmt.Fprintf(stderr, "Please verify the configuration file syntax or configure manually using 'vck mcp-config'.\n")
			return 1
		}
		out = merged
	} else {
		fresh, freshErr := mergeConfig(nil, exePath, targetKey)
		if freshErr != nil {
			fmt.Fprintf(stderr, "failed to create fresh config: %v\n", freshErr)
			return 1
		}
		out = fresh
	}

	if dryRun {
		fmt.Fprintf(stdout, "[dry-run] Would write Vivechak configuration to %s (%s):\n", configPath, targetKey)
		fmt.Fprintln(stdout, string(out))
		return 0
	}

	if err := writeConfigFile(configPath, out); err != nil {
		fmt.Fprintf(stderr, "Error writing configuration: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "✓ Successfully registered Vivechak in %s (%s)\n", hostName, configPath)
	return 0
}

// detectWorkspaceHost checks the current working directory for known workspace indicators.
func detectWorkspaceHost(cwd string) string {
	if cwd == "" {
		return ""
	}
	checks := []struct {
		dir  string
		host string
	}{
		{".cursor", "cursor"},
		{".agents", "antigravity"},
		{".vscode", "vscode"},
		{".trae", "trae"},
		{".omp", "omp"},
		{".factory", "droid"},
		{".kiro", "kiro"},
		{".devin", "devin"},
		{".cline", "cline"},
		{".roo", "roo"},
	}
	for _, c := range checks {
		if fi, err := os.Stat(filepath.Join(cwd, c.dir)); err == nil && fi.IsDir() {
			return c.host
		}
	}
	return ""
}

// formatHostName returns a human-friendly name for a host preset.
func formatHostName(h string) string {
	switch h {
	case "cursor":
		return "Cursor"
	case "vscode", "code":
		return "VS Code"
	case "claude-desktop", "claude":
		return "Claude Desktop"
	case "windsurf":
		return "Windsurf"
	case "antigravity", "agy":
		return "Google Antigravity"
	case "zed":
		return "Zed"
	case "kiro":
		return "AWS Kiro"
	case "trae":
		return "ByteDance Trae"
	case "omp":
		return "Oh My Pi (OMP)"
	case "openhands":
		return "OpenHands"
	case "droid":
		return "Factory Droid"
	case "cline":
		return "Cline"
	case "roo":
		return "Roo Code"
	case "devin":
		return "Cognition Devin"
	default:
		return h
	}
}

// printSetupUsage prints user guidance for the setup subcommand.
func printSetupUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: vivechak setup [host|filepath] [--dry-run]")
	fmt.Fprintln(w, "Alias: vivechak install (or: vck setup, vck install)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Set up Vivechak Model Context Protocol (MCP) in your AI editor, IDE, or agent harness.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Arguments:")
	fmt.Fprintln(w, "  <host>        Desktop shortcut (e.g. cursor, agy, vscode, zed, kiro, trae, omp, etc.)")
	fmt.Fprintln(w, "  <filepath>    Target configuration file path to write and merge into")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --dry-run, -n  Preview configuration changes without writing to disk")
	fmt.Fprintln(w, "  --help, -h     Show this help message")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Available Presets:")
	fmt.Fprintf(w, "  %s\n", desktopShortcutsList)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  vck setup cursor                       # Configure Cursor")
	fmt.Fprintln(w, "  vck setup agy                          # Configure Google Antigravity (IDE, 2.0, CLI)")
	fmt.Fprintln(w, "  vck setup vscode                       # Configure VS Code workspace")
	fmt.Fprintln(w, "  vck setup ./custom-mcp.json            # Write directly to any agent config file")
	fmt.Fprintln(w, "  vck setup cursor --dry-run             # Preview configuration without writing")
}
