package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func TestRunCLI_Version(t *testing.T) {
	for _, arg := range []string{"version", "--version", "-v"} {
		var stdout, stderr bytes.Buffer
		code := runCLI([]string{"vivechak", arg}, &stdout, &stderr)
		if code != 0 {
			t.Errorf("expected exit code 0 for %s, got %d", arg, code)
		}
		if !strings.Contains(stdout.String(), version) {
			t.Errorf("expected version output for %s to contain %q, got %q", arg, version, stdout.String())
		}
	}
}

func TestRunCLI_Help(t *testing.T) {
	for _, arg := range []string{"help", "--help", "-h"} {
		var stdout, stderr bytes.Buffer
		code := runCLI([]string{"vivechak", arg}, &stdout, &stderr)
		if code != 0 {
			t.Errorf("expected exit code 0 for %s, got %d", arg, code)
		}
		if !strings.Contains(stdout.String(), "Usage: vivechak <command>") {
			t.Errorf("expected help output for %s, got %q", arg, stdout.String())
		}
	}
}

func TestRunCLI_UnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"vivechak", "unknown-subcommand"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 for unknown command, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Usage: vivechak <command>") {
		t.Errorf("expected usage in stderr, got %q", stderr.String())
	}
}

func TestRunCLI_Doctor(t *testing.T) {
	// 1. Doctor --help
	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"vivechak", "doctor", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for doctor --help, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage: vivechak doctor") {
		t.Errorf("expected doctor usage, got %q", stdout.String())
	}

	// 2. Doctor on non-existent directory
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "doctor", filepath.Join(t.TempDir(), "nonexistent")}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 for non-existent workspace, got %d", code)
	}

	// 3. Doctor on healthy workspace
	tmpDir := t.TempDir()
	for _, dir := range []string{core.ResearchDir, core.SessionsDir, core.TemplatesDir} {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, tmpl := range core.TemplatesToCopy {
		if err := os.WriteFile(filepath.Join(tmpDir, core.TemplatesDir, tmpl), []byte("# Template\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "doctor", tmpDir}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for healthy workspace, got %d (stderr: %s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Workspace is healthy.") {
		t.Errorf("expected 'Workspace is healthy.', got %q", stdout.String())
	}
}

func TestRunCLI_MCPConfig(t *testing.T) {
	// 1. mcp-config --help
	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"vivechak", "mcp-config", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for mcp-config --help, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage: vivechak mcp-config") {
		t.Errorf("expected mcp-config usage, got %q", stdout.String())
	}

	// 2. mcp-config default output
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "mcp-config"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for mcp-config default, got %d", code)
	}
	var defaultCfg map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &defaultCfg); err != nil {
		t.Fatalf("failed to parse mcp-config JSON: %v", err)
	}
	if _, ok := defaultCfg["mcpServers"]; !ok {
		t.Errorf("expected mcpServers key in default config, got: %v", defaultCfg)
	}

	// 3. mcp-config presets
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "mcp-config", "--preset", "vscode"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for vscode preset, got %d", code)
	}
	var vscodeCfg map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &vscodeCfg); err != nil {
		t.Fatalf("failed to parse vscode config JSON: %v", err)
	}
	if _, ok := vscodeCfg["servers"]; !ok {
		t.Errorf("expected servers key in vscode config, got: %v", vscodeCfg)
	}

	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "mcp-config", "--preset", "zed"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for zed preset, got %d", code)
	}
	var zedCfg map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &zedCfg); err != nil {
		t.Fatalf("failed to parse zed config JSON: %v", err)
	}
	if _, ok := zedCfg["context_servers"]; !ok {
		t.Errorf("expected context_servers key in zed config, got: %v", zedCfg)
	}

	// 4. mcp-config invalid preset
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "mcp-config", "--preset", "invalid-preset"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 for invalid preset, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown desktop preset") {
		t.Errorf("expected unknown desktop preset error, got %q", stderr.String())
	}

	// 5. mcp-config --write with missing target
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "mcp-config", "--write"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 for --write without target, got %d", code)
	}
	if !strings.Contains(stderr.String(), "--write requires a target file location") {
		t.Errorf("expected target requirement error, got %q", stderr.String())
	}

	// 6. mcp-config --path <file> --write
	customPath := filepath.Join(t.TempDir(), "custom-agent-config.json")
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "mcp-config", "--path", customPath, "--write"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for --path --write, got %d (stderr: %s)", code, stderr.String())
	}
	data, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("failed to read written config file: %v", err)
	}
	var customCfg map[string]any
	if err := json.Unmarshal(data, &customCfg); err != nil {
		t.Fatalf("written config is not valid JSON: %v", err)
	}

	// 7. Test aliases --client and --file with merging into existing file
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "mcp-config", "--file", customPath, "--client", "cursor", "--write"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for alias merge, got %d (stderr: %s)", code, stderr.String())
	}

	// 8. Corrupted existing config triggers backup (.bak)
	corruptPath := filepath.Join(t.TempDir(), "corrupt-config.json")
	if err := os.WriteFile(corruptPath, []byte("NOT_JSON{"), 0644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "mcp-config", "--path", corruptPath, "--write"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 on corrupt merge with backup, got %d", code)
	}
	if _, err := os.Stat(corruptPath + ".bak"); err != nil {
		t.Errorf("expected backup file %s.bak to exist", corruptPath)
	}
}

func TestRunServeWithContext_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	code := runServeWithContext(ctx, logger)
	if code != 0 {
		t.Errorf("expected exit code 0 on canceled context, got %d", code)
	}
}

func TestRunCLI_Serve(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var stdout, stderr bytes.Buffer
	// Test explicit serve
	code := runCLIWithContext(ctx, []string{"vivechak", "serve"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for serve with canceled context, got %d", code)
	}

	// Test default empty subcommand
	stdout.Reset()
	stderr.Reset()
	code = runCLIWithContext(ctx, []string{"vivechak"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for default command with canceled context, got %d", code)
	}
}
