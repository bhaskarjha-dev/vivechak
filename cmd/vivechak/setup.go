package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/tailscale/hujson"
	"gopkg.in/yaml.v3"
)

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
	homeDir, _ := getUserHome()
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
	var oldBytes []byte
	if strings.HasSuffix(configPath, ".yaml") || strings.HasSuffix(configPath, ".yml") {
		b, _ := os.ReadFile(configPath)
		oldBytes = b
		out = mergeGooseYAML(b, exePath)
	} else {
		b, err := os.ReadFile(configPath)
		if err == nil {
			oldBytes = b
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
		printConfigDiff(stdout, oldBytes, out, configPath)
		return 0
	}

	if err := writeConfigFile(configPath, out); err != nil {
		fmt.Fprintf(stderr, "Error writing configuration: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "✓ Successfully registered Vivechak in %s (%s)\n", hostName, configPath)
	return 0
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
	fmt.Fprintf(w, "  %s\n", verifiedPresetsList)
	fmt.Fprintln(w, "  (unverified hosts: use 'vck setup <path>')")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  vck setup cursor                       # Configure Cursor")
	fmt.Fprintln(w, "  vck setup agy                          # Configure Google Antigravity (IDE, 2.0, CLI)")
	fmt.Fprintln(w, "  vck setup vscode                       # Configure VS Code workspace")
	fmt.Fprintln(w, "  vck setup ./custom-mcp.json            # Write directly to any agent config file")
	fmt.Fprintln(w, "  vck setup cursor --dry-run             # Preview configuration without writing")
}

var secretValueRegex = regexp.MustCompile(`(?i)("(?:[^"]*(?:token|secret|key|password|credential|auth)[^"]*)"\s*:\s*)"(?:[^"]+)"`)

// redactConfigSecret replaces sensitive token/key/password values in diff lines to prevent leakage.
func redactConfigSecret(line string) string {
	return secretValueRegex.ReplaceAllString(line, `$1"[REDACTED]"`)
}

// extractServerEntry locates and extracts the configuration object for serverName
// (e.g. "vivechak") from JSON, JSONC, HuJSON, or YAML config files.
func extractServerEntry(data []byte, serverName string) string {
	if len(bytes.TrimSpace(data)) == 0 {
		return ""
	}

	// Try JSON/HuJSON first
	if std, err := hujson.Standardize(data); err == nil {
		var root map[string]any
		if err := json.Unmarshal(std, &root); err == nil {
			for _, key := range []string{"mcpServers", "context_servers", "servers", "mcp"} {
				if srvMap, ok := root[key].(map[string]any); ok {
					if entry, found := srvMap[serverName]; found {
						b, _ := json.MarshalIndent(map[string]any{serverName: entry}, "", "  ")
						return string(b)
					}
				}
			}
			if entry, found := root[serverName]; found {
				b, _ := json.MarshalIndent(map[string]any{serverName: entry}, "", "  ")
				return string(b)
			}
		}
	}

	// Try YAML (e.g. Goose config)
	var yRoot map[string]any
	if err := yaml.Unmarshal(data, &yRoot); err == nil {
		for _, key := range []string{"extensions", "mcpServers", "servers"} {
			if srvMap, ok := yRoot[key].(map[string]any); ok {
				if entry, found := srvMap[serverName]; found {
					b, _ := yaml.Marshal(map[string]any{serverName: entry})
					return string(b)
				}
			}
		}
	}

	return ""
}

// printConfigDiff prints focused diff of old versus new server entries (+ and -) for safe dry-run preview.
func printConfigDiff(w io.Writer, oldBytes, newBytes []byte, filePath string) {
	fmt.Fprintf(w, "--- %s\n+++ %s\n", filePath, filePath)
	oldEntry := extractServerEntry(oldBytes, "vivechak")
	newEntry := extractServerEntry(newBytes, "vivechak")

	if oldEntry == "" {
		fmt.Fprintf(w, "- (vivechak not configured)\n")
	} else {
		for _, line := range strings.Split(oldEntry, "\n") {
			if strings.TrimSpace(line) != "" {
				fmt.Fprintf(w, "- %s\n", redactConfigSecret(line))
			}
		}
	}

	if newEntry == "" && len(newBytes) > 0 {
		for _, line := range strings.Split(string(newBytes), "\n") {
			if strings.TrimSpace(line) != "" {
				fmt.Fprintf(w, "+ %s\n", redactConfigSecret(line))
			}
		}
	} else {
		for _, line := range strings.Split(newEntry, "\n") {
			if strings.TrimSpace(line) != "" {
				fmt.Fprintf(w, "+ %s\n", redactConfigSecret(line))
			}
		}
	}
}
