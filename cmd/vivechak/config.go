package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func runMCPConfig() {
	var host string
	var write bool
	var customPath string

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		if args[i] == "--client" && i+1 < len(args) {
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

	configObj := map[string]any{
		"mcpServers": map[string]any{
			"vivechak": map[string]any{
				"command": exePath,
				"args":    []string{"serve"},
			},
		},
	}

	// Without --write, print universal MCP config JSON compatible with any MCP client
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
			fmt.Fprintf(os.Stderr, "%v\n", err)
			fmt.Fprintln(os.Stderr, "Supported client presets: cursor, vscode, claude-desktop, windsurf, antigravity, chatgpt, codex, kiro")
			fmt.Fprintln(os.Stderr, "Or provide a custom config path: vivechak mcp-config --path <filepath> --write")
			os.Exit(1)
		}
		if resolvedPath == "" {
			b, _ := json.MarshalIndent(configObj, "", "  ")
			fmt.Fprintln(os.Stdout, string(b))
			os.Exit(0)
		}
		configPath = resolvedPath
	} else {
		fmt.Fprintln(os.Stderr, "Error: --write requires a target file location.")
		fmt.Fprintln(os.Stderr, "Usage: vivechak mcp-config --client <preset> --write")
		fmt.Fprintln(os.Stderr, "   or: vivechak mcp-config --path <config_file_path> --write")
		fmt.Fprintln(os.Stderr, "Supported client presets: cursor, vscode, claude-desktop, windsurf, antigravity, chatgpt, codex, kiro")
		os.Exit(1)
	}

	// merge with existing
	var out []byte
	b, err := os.ReadFile(configPath)
	if err == nil {
		merged, mergeErr := mergeConfig(b, exePath)
		if mergeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", mergeErr)
			fmt.Fprintf(os.Stderr, "Creating backup at %s.bak and writing fresh config\n", configPath)
			if backupErr := os.WriteFile(configPath+".bak", b, 0644); backupErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not create backup: %v\n", backupErr)
			}
			fresh, freshErr := mergeConfig(nil, exePath)
			if freshErr != nil {
				fmt.Fprintf(os.Stderr, "failed to create fresh config: %v\n", freshErr)
				os.Exit(1)
			}
			out = fresh
		} else {
			out = merged
		}
	} else {
		fresh, freshErr := mergeConfig(nil, exePath)
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

// mergeConfig merges the vivechak MCP server entry into existing config JSON,
// preserving number fidelity with json.Number.
func mergeConfig(existingBytes []byte, exePath string) ([]byte, error) {
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

	if existing["mcpServers"] == nil {
		existing["mcpServers"] = make(map[string]any)
	}

	servers, ok := existing["mcpServers"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("mcpServers in config is not a JSON object")
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
func writeConfigFile(configPath string, out []byte) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tmpFile := configPath + ".tmp"
	f, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	if _, err := f.Write(out); err != nil {
		f.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	f.Close()

	if err := os.Rename(tmpFile, configPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
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
