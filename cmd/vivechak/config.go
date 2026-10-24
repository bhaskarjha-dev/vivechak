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

	"github.com/tailscale/hujson"
)

const desktopShortcutsList = "cursor, vscode, claude-desktop (or claude), claude-code, mcp (or portable), opencode, copilot, void, continue, goose, windsurf, antigravity (or agy), zed, kiro, trae, omp, openhands, droid, cline, roo, devin"

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
	if p := findPreset(host); p != nil && p.ServerKey != "" {
		serverKey = p.ServerKey
	}

	configObj := map[string]any{
		serverKey: map[string]any{
			"vivechak": buildServerConfig(serverKey, exePath, host),
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
	if strings.HasSuffix(configPath, ".yaml") || strings.HasSuffix(configPath, ".yml") {
		b, _ := os.ReadFile(configPath)
		out = mergeGooseYAML(b, exePath)
	} else {
		b, err := os.ReadFile(configPath)
		if err == nil {
			// If existing file is valid JSON/JSONC, check its existing key
			if len(bytes.TrimSpace(b)) > 0 {
				var existing map[string]any
				if std, err := hujson.Standardize(bytes.Clone(b)); err == nil {
					dec := json.NewDecoder(bytes.NewReader(std))
					if dec.Decode(&existing) == nil {
						targetKey = determineServerKey(host, configPath, existing)
					}
				}
			}

			merged, mergeErr := mergeConfig(b, exePath, targetKey)
			if mergeErr != nil {
				fmt.Fprintf(stderr, "Warning: %v\n", mergeErr)
				targetPath := configPath
				if realPath, err := filepath.EvalSymlinks(configPath); err == nil && realPath != "" {
					targetPath = realPath
				}
				perm := os.FileMode(0644)
				if fi, err := os.Stat(targetPath); err == nil {
					perm = fi.Mode().Perm()
				}
				fmt.Fprintf(stderr, "Creating backup at %s.bak and writing fresh config\n", targetPath)
				if backupErr := os.WriteFile(targetPath+".bak", b, perm); backupErr != nil {
					fmt.Fprintf(stderr, "Warning: could not create backup: %v\n", backupErr)
				}
				_ = os.Chmod(targetPath+".bak", perm)
				if configPath != targetPath {
					_ = os.WriteFile(configPath+".bak", b, perm)
					_ = os.Chmod(configPath+".bak", perm)
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
	}

	if err := writeConfigFile(configPath, out); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}

	fmt.Fprintf(stderr, "Successfully wrote config to %s\n", configPath)
	return 0
}

// determineServerKey returns the appropriate top-level JSON key ("servers", "context_servers", "mcp", "extensions", or "mcpServers").
func determineServerKey(host, configPath string, existing map[string]any) string {
	if host == "opencode" || strings.Contains(strings.ToLower(configPath), "opencode") {
		return "mcp"
	}
	if host == "goose" || strings.Contains(strings.ToLower(configPath), "goose") {
		return "extensions"
	}
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
		if existing["mcp"] != nil && existing["mcpServers"] == nil {
			return "mcp"
		}
		if existing["extensions"] != nil && existing["mcpServers"] == nil {
			return "extensions"
		}
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
		if base == "opencode" || base == ".opencode" {
			return "mcp"
		}
		if base == "goose" || base == ".goose" {
			return "extensions"
		}
	}

	return "mcpServers"
}

// resolveLauncher detects ephemeral execution wrappers (such as npx cache paths
// or Homebrew Cellar temporary locations) and falls back to running via npx.
func resolveLauncher(exePath string) (cmd string, args []string) {
	clean := strings.ToLower(filepath.ToSlash(exePath))
	if strings.Contains(clean, "_npx") ||
		strings.Contains(clean, "node_modules/.cache") ||
		(strings.Contains(clean, "cellar") && strings.Contains(clean, "tmp")) {
		return "npx", []string{"-y", "vivechak", "serve"}
	}
	return exePath, []string{"serve"}
}

// buildServerConfig returns the appropriate configuration object for the given server key and optional host.
func buildServerConfig(serverKey, exePath string, host ...string) map[string]any {
	cmd, args := resolveLauncher(exePath)
	switch serverKey {
	case "mcp":
		fullCmd := append([]string{cmd}, args...)
		return map[string]any{
			"type":    "local",
			"command": fullCmd,
			"enabled": true,
		}
	case "extensions":
		return map[string]any{
			"name":    "vivechak",
			"type":    "stdio",
			"cmd":     cmd,
			"args":    args,
			"enabled": true,
			"timeout": 300,
		}
	default:
		cfg := map[string]any{
			"command": cmd,
			"args":    args,
		}
		var h string
		if len(host) > 0 {
			h = strings.ToLower(host[0])
		}
		if serverKey == "servers" || h == "cursor" || strings.Contains(h, "cursor") ||
			h == "vscode" || h == "code" || strings.Contains(h, "vscode") ||
			h == "windsurf" || strings.Contains(h, "windsurf") {
			cfg["env"] = map[string]string{
				"VCK_PROJECT_ROOT": "${workspaceFolder}",
			}
		}
		return cfg
	}
}

// mergeGooseYAML merges the vivechak extension into a Goose YAML configuration file.
func mergeGooseYAML(existingBytes []byte, exePath string) []byte {
	cmd, args := resolveLauncher(exePath)
	var argsYAML strings.Builder
	for _, a := range args {
		fmt.Fprintf(&argsYAML, "      - %s\n", a)
	}
	cleanExisting := strings.TrimSpace(string(existingBytes))
	vivechakBlock := fmt.Sprintf("  vivechak:\n    name: vivechak\n    type: stdio\n    cmd: %q\n    args:\n%s    enabled: true\n    timeout: 300", cmd, argsYAML.String())

	if cleanExisting == "" {
		return []byte(fmt.Sprintf("extensions:\n%s\n", vivechakBlock))
	}

	lines := strings.Split(cleanExisting, "\n")
	extensionsIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "extensions:" || strings.HasPrefix(trimmed, "extensions:") {
			extensionsIdx = i
			break
		}
	}

	if extensionsIdx == -1 {
		// No extensions: key found, append at the end
		return []byte(fmt.Sprintf("%s\n\nextensions:\n%s\n", cleanExisting, vivechakBlock))
	}

	// extensions: found, check if vivechak: already exists under it
	vivechakStart := -1
	vivechakEnd := -1

	for i := extensionsIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// If line has no indentation, it is a new top-level section
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			break
		}
		if trimmed == "vivechak:" || strings.HasPrefix(trimmed, "vivechak:") {
			vivechakStart = i
			vivechakEnd = len(lines)
			for j := i + 1; j < len(lines); j++ {
				jTrimmed := strings.TrimSpace(lines[j])
				if jTrimmed == "" {
					continue
				}
				// If a sibling key (2-space indent) or top-level key (0-space indent)
				if (len(lines[j]) >= 2 && lines[j][0] == ' ' && lines[j][1] == ' ' && lines[j][2] != ' ' && lines[j][2] != '\t') || (lines[j][0] != ' ' && lines[j][0] != '\t') {
					vivechakEnd = j
					break
				}
			}
			break
		}
	}

	if vivechakStart >= 0 && vivechakEnd > vivechakStart {
		// Replace existing vivechak block
		var newLines []string
		newLines = append(newLines, lines[:vivechakStart]...)
		newLines = append(newLines, vivechakBlock)
		newLines = append(newLines, lines[vivechakEnd:]...)
		return []byte(strings.Join(newLines, "\n") + "\n")
	}

	// Insert right after extensions: line
	var newLines []string
	newLines = append(newLines, lines[:extensionsIdx+1]...)
	newLines = append(newLines, vivechakBlock)
	newLines = append(newLines, lines[extensionsIdx+1:]...)
	return []byte(strings.Join(newLines, "\n") + "\n")
}

// mergeConfig merges the vivechak MCP server entry into existing config JSON/JSONC,
// preserving comments and formatting using hujson.
func mergeConfig(existingBytes []byte, exePath string, preferredKey ...string) ([]byte, error) {
	var v hujson.Value
	var err error

	if len(bytes.TrimSpace(existingBytes)) > 0 {
		v, err = hujson.Parse(existingBytes)
		if err != nil {
			return nil, fmt.Errorf("existing config is not valid JSON: %w", err)
		}
	} else {
		v, err = hujson.Parse([]byte("{\n}\n"))
		if err != nil {
			return nil, fmt.Errorf("initializing empty config: %w", err)
		}
	}

	if v.Value == nil || v.Value.Kind() != '{' {
		return nil, fmt.Errorf("existing config root is not a JSON object")
	}

	key := "mcpServers"
	if len(preferredKey) > 0 && preferredKey[0] != "" {
		key = preferredKey[0]
	} else {
		if v.Find("/servers") != nil && v.Find("/mcpServers") == nil {
			key = "servers"
		} else if v.Find("/context_servers") != nil && v.Find("/mcpServers") == nil {
			key = "context_servers"
		} else if v.Find("/mcp") != nil && v.Find("/mcpServers") == nil {
			key = "mcp"
		} else if v.Find("/extensions") != nil && v.Find("/mcpServers") == nil {
			key = "extensions"
		}
	}

	// Check if key exists and is an object
	if targetVal := v.Find("/" + key); targetVal != nil {
		if targetVal.Value == nil || targetVal.Value.Kind() != '{' {
			return nil, fmt.Errorf("%s in config is not a JSON object", key)
		}
	} else {
		// Key does not exist; add it as an empty object
		patchKey := []map[string]any{
			{
				"op":    "add",
				"path":  "/" + key,
				"value": map[string]any{},
			},
		}
		patchBytes, err := json.Marshal(patchKey)
		if err != nil {
			return nil, fmt.Errorf("marshaling key patch: %w", err)
		}
		if err := v.Patch(patchBytes); err != nil {
			return nil, fmt.Errorf("patching %s: %w", key, err)
		}
	}

	var hostHint string
	if len(preferredKey) > 1 {
		hostHint = preferredKey[1]
	}

	serverCfg := buildServerConfig(key, exePath, hostHint)
	patchServer := []map[string]any{
		{
			"op":    "add",
			"path":  "/" + key + "/vivechak",
			"value": serverCfg,
		},
	}
	patchServerBytes, err := json.Marshal(patchServer)
	if err != nil {
		return nil, fmt.Errorf("marshaling server patch: %w", err)
	}
	if err := v.Patch(patchServerBytes); err != nil {
		return nil, fmt.Errorf("patching %s/vivechak: %w", key, err)
	}

	v.Format()
	out := v.Pack()
	if !bytes.HasSuffix(out, []byte("\n")) {
		out = append(out, '\n')
	}
	return out, nil
}

// writeConfigFile atomically writes config content to configPath.
// If configPath is a symlink, it resolves the symlink target via filepath.EvalSymlinks
// and writes to the underlying target file directly so that the symlink remains intact.
// If the destination file already exists, it preserves its original file permissions
// for both the temporary file and the <file>.bak backup file.
func writeConfigFile(configPath string, out []byte) (retErr error) {
	// 1. Resolve symlink destination if configPath is a symlink
	targetPath := configPath
	if realPath, err := filepath.EvalSymlinks(configPath); err == nil && realPath != "" {
		targetPath = realPath
	} else if lfi, lerr := os.Lstat(configPath); lerr == nil && (lfi.Mode()&os.ModeSymlink != 0) {
		if linkDst, rerr := os.Readlink(configPath); rerr == nil && linkDst != "" {
			if !filepath.IsAbs(linkDst) {
				linkDst = filepath.Join(filepath.Dir(configPath), linkDst)
			}
			targetPath = filepath.Clean(linkDst)
		}
	}

	// 2. Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(configPath), err)
	}
	if filepath.Dir(targetPath) != filepath.Dir(configPath) {
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create target directory %s: %w", filepath.Dir(targetPath), err)
		}
	}

	// 3. Detect and preserve existing permissions
	perm := os.FileMode(0644)
	if fi, err := os.Stat(targetPath); err == nil {
		perm = fi.Mode().Perm()
		if existing, readErr := os.ReadFile(targetPath); readErr == nil {
			_ = os.WriteFile(targetPath+".bak", existing, perm)
			_ = os.Chmod(targetPath+".bak", perm)
		}
	} else if fi, err := os.Stat(configPath); err == nil {
		perm = fi.Mode().Perm()
		if existing, readErr := os.ReadFile(configPath); readErr == nil {
			_ = os.WriteFile(configPath+".bak", existing, perm)
			_ = os.Chmod(configPath+".bak", perm)
		}
	}

	if configPath != targetPath {
		if existing, readErr := os.ReadFile(targetPath); readErr == nil {
			_ = os.WriteFile(configPath+".bak", existing, perm)
			_ = os.Chmod(configPath+".bak", perm)
		}
	}

	// 4. Write to temp file alongside the targetPath
	tmpFile := targetPath + ".tmp"
	f, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	_ = os.Chmod(tmpFile, perm)
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

	// 5. Atomically replace targetPath (preserving configPath symlink)
	err = os.Rename(tmpFile, targetPath)
	if err != nil && runtime.GOOS == "windows" {
		backoff := 50 * time.Millisecond
		for attempt := 0; attempt < 3; attempt++ {
			time.Sleep(backoff)
			backoff *= 2
			if err = os.Rename(tmpFile, targetPath); err == nil {
				break
			}
		}
	}
	if err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	_ = os.Chmod(targetPath, perm)
	return nil
}

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
			return resolveVSCodeStoragePath("saoudrizwan.claude-dev", homeDir, appData, goos)
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
			return resolveVSCodeStoragePath("rooveterinaryinc.roo-cline", homeDir, appData, goos)
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
			// No target and no workspace detected: print informative guidance
			fmt.Fprintln(stdout, "No supported AI editor or harness workspace detected in the current directory.")
			fmt.Fprintln(stdout, "Specify a target host preset or configuration file path.")
			fmt.Fprintln(stdout, "")
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
	} else if strings.Contains(target, "/") || strings.Contains(target, "\\") || strings.HasSuffix(target, ".json") || strings.HasSuffix(target, ".jsonc") || strings.HasSuffix(target, ".yaml") || strings.HasSuffix(target, ".yml") {
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
	hostHint := target
	if !isSupportedPreset(target) {
		cleanLow := strings.ToLower(filepath.Clean(configPath))
		if strings.Contains(cleanLow, "cursor") {
			hostHint = "cursor"
		} else if strings.Contains(cleanLow, "windsurf") || strings.Contains(cleanLow, "codeium") {
			hostHint = "windsurf"
		} else if strings.Contains(cleanLow, "vscode") || strings.Contains(cleanLow, ".vscode") {
			hostHint = "vscode"
		}
	}

	var out []byte
	if strings.HasSuffix(configPath, ".yaml") || strings.HasSuffix(configPath, ".yml") {
		b, _ := os.ReadFile(configPath)
		out = mergeGooseYAML(b, exePath)
	} else {
		b, err := os.ReadFile(configPath)
		if err == nil {
			if len(bytes.TrimSpace(b)) > 0 {
				var existing map[string]any
				if std, err := hujson.Standardize(bytes.Clone(b)); err == nil {
					dec := json.NewDecoder(bytes.NewReader(std))
					if dec.Decode(&existing) == nil {
						targetKey = determineServerKey(target, configPath, existing)
					}
				}
			}
			merged, mergeErr := mergeConfig(b, exePath, targetKey, hostHint)
			if mergeErr != nil {
				fmt.Fprintf(stderr, "Error: could not merge Vivechak into existing configuration at %s: %v\n", configPath, mergeErr)
				fmt.Fprintf(stderr, "To protect your existing settings, Vivechak will not overwrite this file.\n")
				fmt.Fprintf(stderr, "Please verify the configuration file syntax or configure manually using 'vck mcp-config'.\n")
				return 1
			}
			out = merged
		} else {
			fresh, freshErr := mergeConfig(nil, exePath, targetKey, hostHint)
			if freshErr != nil {
				fmt.Fprintf(stderr, "failed to create fresh config: %v\n", freshErr)
				return 1
			}
			out = fresh
		}
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
