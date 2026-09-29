package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func runMCPConfig() {
	var host string
	var write bool
	var customPath string

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		if args[i] == "--help" || args[i] == "-h" {
			fmt.Fprintln(os.Stdout, "Usage: vivechak mcp-config [--path <filepath>] [--preset <shortcut>] [--write]")
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, "Outputs universal Model Context Protocol (MCP) JSON configuration for Vivechak.")
			fmt.Fprintln(os.Stdout, "Compatible with any MCP-compliant AI tool, agent harness, IDE, or runtime.")
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, "Flags:")
			fmt.Fprintln(os.Stdout, "  --write                 Write and merge configuration directly to target file")
			fmt.Fprintln(os.Stdout, "  --path, --file <path>   Write to an arbitrary configuration file path (universal)")
			fmt.Fprintln(os.Stdout, "  --preset <shortcut>     Optional desktop path shortcut (e.g. cursor, vscode, zed)")
			fmt.Fprintln(os.Stdout, "  --client <shortcut>     Alias for --preset (backward compatible)")
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, "Desktop Path Shortcuts:")
			fmt.Fprintln(os.Stdout, "  cursor, vscode, claude-desktop, windsurf, antigravity, zed, kiro")
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, "Examples:")
			fmt.Fprintln(os.Stdout, "  vivechak mcp-config                                     # Universal MCP JSON to stdout")
			fmt.Fprintln(os.Stdout, "  vivechak mcp-config --path ~/.omp/agent/mcp.json --write # Write directly to any agent config")
			fmt.Fprintln(os.Stdout, "  vivechak mcp-config --preset cursor --write             # Desktop shortcut for Cursor")
			os.Exit(0)
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
		fmt.Fprintf(os.Stderr, "failed to get executable path: %v\n", err)
		os.Exit(1)
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to make path absolute: %v\n", err)
		os.Exit(1)
	}

	if host != "" && !isSupportedPreset(host) {
		fmt.Fprintf(os.Stderr, "Error: unknown desktop preset: %s (available shortcuts: cursor, vscode, claude-desktop, windsurf, antigravity, zed, kiro)\n", host)
		os.Exit(1)
	}

	serverKey := "mcpServers"
	if host == "vscode" {
		serverKey = "servers"
	} else if host == "zed" {
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
		fmt.Fprintln(os.Stdout, string(b))
		os.Exit(0)
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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		configPath = resolvedPath
	} else {
		fmt.Fprintln(os.Stderr, "Error: --write requires a target file location.")
		fmt.Fprintln(os.Stderr, "Usage: vivechak mcp-config --path <config_file_path> --write")
		fmt.Fprintln(os.Stderr, "   or: vivechak mcp-config --preset <shortcut> --write")
		fmt.Fprintln(os.Stderr, "Available desktop shortcuts: cursor, vscode, claude-desktop, windsurf, antigravity, zed, kiro")
		os.Exit(1)
	}

	// Determine serverKey based on host and target path
	targetKey := determineServerKey(host, configPath, nil)

	// merge with existing
	var out []byte
	b, err := os.ReadFile(configPath)
	if err == nil {
		// If existing file is valid JSON, check its existing key
		if len(bytes.TrimSpace(b)) > 0 {
			var existing map[string]any
			dec := json.NewDecoder(bytes.NewReader(b))
			if dec.Decode(&existing) == nil {
				targetKey = determineServerKey(host, configPath, existing)
			}
		}

		merged, mergeErr := mergeConfig(b, exePath, targetKey)
		if mergeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", mergeErr)
			fmt.Fprintf(os.Stderr, "Creating backup at %s.bak and writing fresh config\n", configPath)
			if backupErr := os.WriteFile(configPath+".bak", b, 0644); backupErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not create backup: %v\n", backupErr)
			}
			fresh, freshErr := mergeConfig(nil, exePath, targetKey)
			if freshErr != nil {
				fmt.Fprintf(os.Stderr, "failed to create fresh config: %v\n", freshErr)
				os.Exit(1)
			}
			out = fresh
		} else {
			out = merged
		}
	} else {
		fresh, freshErr := mergeConfig(nil, exePath, targetKey)
		if freshErr != nil {
			fmt.Fprintf(os.Stderr, "failed to create fresh config: %v\n", freshErr)
			os.Exit(1)
		}
		out = fresh
	}

	if err := writeConfigFile(configPath, out); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Successfully wrote config to %s\n", configPath)
}

// determineServerKey returns the appropriate top-level JSON key ("servers", "context_servers", or "mcpServers").
func determineServerKey(host, configPath string, existing map[string]any) string {
	if host == "vscode" {
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
	if filepath.Base(filepath.Dir(clean)) == ".vscode" || filepath.Base(clean) == "mcp.json" && strings.Contains(clean, ".vscode") {
		return "servers"
	}
	if strings.Contains(strings.ToLower(clean), "zed") {
		return "context_servers"
	}

	return "mcpServers"
}

// mergeConfig merges the vivechak MCP server entry into existing config JSON,
// preserving number fidelity with json.Number.
func mergeConfig(existingBytes []byte, exePath string, preferredKey ...string) ([]byte, error) {
	var existing map[string]any
	if len(bytes.TrimSpace(existingBytes)) > 0 {
		dec := json.NewDecoder(bytes.NewReader(existingBytes))
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

	if err := os.Rename(tmpFile, configPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}

// resolveConfigPath returns the destination configuration file path for the host client.
func resolveConfigPath(host, homeDir, appData, cwd, goos string) (string, error) {
	switch host {
	case "cursor":
		if homeDir == "" {
			return "", fmt.Errorf("could not determine home dir for cursor config")
		}
		return filepath.Join(homeDir, ".cursor", "mcp.json"), nil
	case "vscode":
		return filepath.Join(cwd, ".vscode", "mcp.json"), nil
	case "claude-desktop":
		if goos == "windows" {
			if appData == "" && homeDir != "" {
				appData = filepath.Join(homeDir, "AppData", "Roaming")
			}
			if appData == "" {
				return "", fmt.Errorf("could not determine AppData path for claude-desktop")
			}
			return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
		} else if goos == "darwin" {
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for claude-desktop")
			}
			return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
		} else { // linux and others
			if homeDir == "" {
				return "", fmt.Errorf("could not determine home dir for claude-desktop")
			}
			return filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json"), nil
		}
	case "antigravity":
		return filepath.Join(cwd, ".gemini", "settings.json"), nil
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
		if homeDir == "" {
			return "", fmt.Errorf("could not determine home dir for kiro config")
		}
		return filepath.Join(homeDir, ".aws", ".kiro", "mcp.json"), nil
	default:
		return "", fmt.Errorf("unknown desktop preset: %s (available shortcuts: cursor, vscode, claude-desktop, windsurf, antigravity, zed, kiro)", host)
	}
}

// isSupportedPreset checks whether the given host name is a recognized desktop preset.
func isSupportedPreset(host string) bool {
	switch host {
	case "cursor", "vscode", "claude-desktop", "windsurf", "antigravity", "zed", "kiro":
		return true
	default:
		return false
	}
}
