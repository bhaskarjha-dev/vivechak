package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func runMCPConfig() {
	var host string
	var write bool

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		if args[i] == "--client" && i+1 < len(args) {
			host = args[i+1]
			i++
		} else if args[i] == "--write" {
			write = true
		}
	}

	if host == "" {
		fmt.Fprintln(os.Stderr, "Usage: vivechak mcp-config --client <host> [--write]")
		fmt.Fprintln(os.Stderr, "Supported clients: cursor, vscode, claude-desktop, windsurf, antigravity, chatgpt, codex, kiro")
		os.Exit(1)
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

	configObj := map[string]any{
		"mcpServers": map[string]any{
			"vivechak": map[string]any{
				"command": exePath,
				"args":    []string{"serve"},
			},
		},
	}

	supportedClients := map[string]bool{
		"cursor":         true,
		"vscode":         true,
		"claude-desktop": true,
		"windsurf":       true,
		"antigravity":    true,
		"chatgpt":        true,
		"codex":          true,
		"kiro":           true,
	}

	if !supportedClients[host] {
		fmt.Fprintf(os.Stderr, "unknown client: %s\n", host)
		fmt.Fprintln(os.Stderr, "Supported clients: cursor, vscode, claude-desktop, windsurf, antigravity, chatgpt, codex, kiro")
		os.Exit(1)
	}

	if !write {
		b, _ := json.MarshalIndent(configObj, "", "  ")
		fmt.Fprintln(os.Stdout, string(b))
		os.Exit(0)
	}

	var configPath string
	cwd, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	appData := os.Getenv("APPDATA")

	resolvedPath, err := resolveConfigPath(host, homeDir, appData, cwd, runtime.GOOS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if resolvedPath == "" {
		b, _ := json.MarshalIndent(configObj, "", "  ")
		fmt.Fprintln(os.Stdout, string(b))
		os.Exit(0)
	}
	configPath = resolvedPath

	// merge with existing
	var existing map[string]any
	b, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(b, &existing)
	}
	if existing == nil {
		existing = make(map[string]any)
	}

	if existing["mcpServers"] == nil {
		existing["mcpServers"] = make(map[string]any)
	}

	servers, ok := existing["mcpServers"].(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "mcpServers in config is not a JSON object")
		os.Exit(1)
	}

	servers["vivechak"] = map[string]any{
		"command": exePath,
		"args":    []string{"serve"},
	}

	out, _ := json.MarshalIndent(existing, "", "  ")
	out = append(out, '\n')

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create directory %s: %v\n", dir, err)
		os.Exit(1)
	}

	tmpFile := configPath + ".tmp"
	f, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open temp file: %v\n", err)
		os.Exit(1)
	}
	if _, err := f.Write(out); err != nil {
		f.Close()
		fmt.Fprintf(os.Stderr, "failed to write temp file: %v\n", err)
		os.Exit(1)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		fmt.Fprintf(os.Stderr, "failed to sync temp file: %v\n", err)
		os.Exit(1)
	}
	f.Close()

	if err := os.Rename(tmpFile, configPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to rename temp file: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Successfully wrote config to %s\n", configPath)
}

// resolveConfigPath returns the destination configuration file path for the host client.
// Returns an empty string for cloud/extension clients that do not write local files.
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
	case "chatgpt", "codex", "kiro":
		return "", nil // cloud or extension-configured clients
	default:
		return "", fmt.Errorf("unknown client: %s", host)
	}
}
