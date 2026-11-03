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
	homeDir, _ := os.UserHomeDir()
	if envHome := os.Getenv("HOME"); envHome != "" {
		homeDir = envHome
	}
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
	fmt.Fprintf(w, "  %s\n", desktopShortcutsList)
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
	line = secretValueRegex.ReplaceAllString(line, `$1"[REDACTED]"`)
	if strings.Contains(line, "FAKE-REDACTED") {
		line = strings.ReplaceAll(line, "FAKE-REDACTED", "[REDACTED]")
	}
	return line
}

// printConfigDiff prints unified diff of old versus new lines (+ and -) for safe dry-run preview.
func printConfigDiff(w io.Writer, oldBytes, newBytes []byte, filePath string) {
	fmt.Fprintf(w, "--- %s\n+++ %s\n", filePath, filePath)
	if len(oldBytes) == 0 {
		for _, line := range strings.Split(string(newBytes), "\n") {
			if line != "" {
				fmt.Fprintf(w, "+%s\n", redactConfigSecret(line))
			}
		}
		return
	}

	oldLines := strings.Split(strings.ReplaceAll(string(oldBytes), "\r\n", "\n"), "\n")
	newLines := strings.Split(strings.ReplaceAll(string(newBytes), "\r\n", "\n"), "\n")

	n := len(oldLines)
	m := len(newLines)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			if strings.TrimSpace(oldLines[i]) == strings.TrimSpace(newLines[j]) {
				dp[i+1][j+1] = dp[i][j] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i+1][j+1] = dp[i][j]
			} else {
				dp[i+1][j+1] = dp[i][j+1]
			}
		}
	}

	type diffEntry struct {
		op   byte
		line string
	}
	var diffOps []diffEntry
	i, j := n, m
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && strings.TrimSpace(oldLines[i-1]) == strings.TrimSpace(newLines[j-1]) {
			diffOps = append(diffOps, diffEntry{op: ' ', line: oldLines[i-1]})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			diffOps = append(diffOps, diffEntry{op: '+', line: newLines[j-1]})
			j--
		} else if i > 0 && (j == 0 || dp[i][j-1] < dp[i-1][j]) {
			diffOps = append(diffOps, diffEntry{op: '-', line: oldLines[i-1]})
			i--
		}
	}

	for k := len(diffOps) - 1; k >= 0; k-- {
		op := diffOps[k]
		switch op.op {
		case '+':
			fmt.Fprintf(w, "+%s\n", redactConfigSecret(op.line))
		case '-':
			fmt.Fprintf(w, "-%s\n", redactConfigSecret(op.line))
		}
	}
}
