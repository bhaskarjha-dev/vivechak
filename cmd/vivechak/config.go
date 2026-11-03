package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tailscale/hujson"
)

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
	if envHome := os.Getenv("HOME"); envHome != "" {
		homeDir = envHome
	}
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
				fmt.Fprintf(stderr, "Error: could not merge Vivechak into existing configuration at %s: %v\n", configPath, mergeErr)
				fmt.Fprintf(stderr, "To protect your existing settings, Vivechak will not overwrite this file.\n")
				fmt.Fprintf(stderr, "Please verify the configuration file syntax or run without --write to inspect output.\n")
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
	if launcher := strings.TrimSpace(os.Getenv("VIVECHAK_LAUNCHER")); launcher != "" {
		if strings.EqualFold(launcher, "uvx") {
			return "uvx", []string{"vivechak", "serve"}
		}
		if strings.EqualFold(launcher, "npx") {
			return "npx", []string{"-y", "vivechak", "serve"}
		}
	}

	clean := strings.ToLower(filepath.ToSlash(exePath))

	// Under /nix/store/ write ~/.nix-profile/bin/vck when it exists
	if strings.Contains(clean, "/nix/store/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			nixProfileBin := filepath.Join(home, ".nix-profile", "bin", "vck")
			if fi, err := os.Stat(nixProfileBin); err == nil && !fi.IsDir() {
				return nixProfileBin, []string{"serve"}
			}
		}
	}

	// If vck or vivechak is found in PATH and points to the running binary (SameFile), prefer that PATH path
	if exeFi, err := os.Stat(exePath); err == nil {
		for _, name := range []string{"vck", "vivechak"} {
			if pathExe, err := exec.LookPath(name); err == nil {
				if pathFi, err := os.Stat(pathExe); err == nil && os.SameFile(exeFi, pathFi) {
					return pathExe, []string{"serve"}
				}
			}
		}
	}

	if strings.Contains(clean, "/vivechak/v") || strings.Contains(clean, `\vivechak\v`) ||
		strings.Contains(clean, "_npx") ||
		strings.Contains(clean, "node_modules/.cache") ||
		strings.Contains(clean, "cellar") || strings.Contains(clean, "caskroom") {
		return "npx", []string{"-y", "vivechak", "serve"}
	}
	if (strings.Contains(clean, "/uv/") || strings.Contains(clean, "\\uv\\")) &&
		(strings.Contains(clean, "cache") || strings.Contains(clean, "archive")) ||
		strings.Contains(clean, "uv-cache") {
		return "uvx", []string{"vivechak", "serve"}
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

	// Do not call v.Format() when existingBytes are present, to preserve original line-by-line whitespace and prevent secret leakage in dry-run diffs
	if len(bytes.TrimSpace(existingBytes)) == 0 {
		v.Format()
	}
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

	// 3. Detect and preserve existing permissions, or use 0600 for new files
	perm := os.FileMode(0600)
	var existing []byte
	if fi, err := os.Stat(targetPath); err == nil {
		perm = fi.Mode().Perm()
		if data, readErr := os.ReadFile(targetPath); readErr == nil {
			existing = data
		}
	} else if fi, err := os.Stat(configPath); err == nil {
		perm = fi.Mode().Perm()
		if data, readErr := os.ReadFile(configPath); readErr == nil {
			existing = data
		}
	}

	// If content is identical, do nothing (no backup, no rewrite)
	if existing != nil && bytes.Equal(existing, out) {
		return nil
	}

	if existing != nil {
		ts := time.Now().UTC().Format("20060102T150405Z")
		tsBak := fmt.Sprintf("%s.vivechak-bak-%s", targetPath, ts)
		if _, err := os.Stat(tsBak); os.IsNotExist(err) {
			_ = os.WriteFile(tsBak, existing, perm)
			_ = os.Chmod(tsBak, perm)
		}
		if configPath != targetPath {
			tsBakConfig := fmt.Sprintf("%s.vivechak-bak-%s", configPath, ts)
			if _, err := os.Stat(tsBakConfig); os.IsNotExist(err) {
				_ = os.WriteFile(tsBakConfig, existing, perm)
				_ = os.Chmod(tsBakConfig, perm)
			}
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
