package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
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
		if !strings.Contains(stdout.String(), "Usage: vck <command>") || !strings.Contains(stdout.String(), "vivechak <command>") {
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
	if !strings.Contains(stderr.String(), "Usage: vck <command>") || !strings.Contains(stderr.String(), "vivechak <command>") {
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

func TestRunCLI_SetupAndInstall(t *testing.T) {
	// 1. setup --help
	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"vivechak", "setup", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for setup --help, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage: vivechak setup") {
		t.Errorf("expected setup usage, got %q", stdout.String())
	}

	// 2. install alias --help
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "install", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for install --help, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage: vivechak setup") {
		t.Errorf("expected setup usage via install alias, got %q", stdout.String())
	}

	// 3. setup directly to custom file
	targetFile := filepath.Join(t.TempDir(), "direct-setup.json")
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vivechak", "setup", targetFile}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for setup file, got %d (stderr: %s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Successfully registered Vivechak") {
		t.Errorf("expected success message, got: %s", stdout.String())
	}
	if _, err := os.Stat(targetFile); err != nil {
		t.Errorf("expected target file %s to be created", targetFile)
	}

	// 4. vck invocation works identically
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vck", "version"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for vck version, got %d", code)
	}
	if !strings.Contains(stdout.String(), version) {
		t.Errorf("expected version output for vck, got %q", stdout.String())
	}
}

func TestRunCLI_Gate(t *testing.T) {
	// 1. gate --help
	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"vivechak", "gate", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for gate --help, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage: vivechak gate") {
		t.Errorf("expected gate usage output, got stdout=%q, stderr=%q", stdout.String(), stderr.String())
	}

	// 2. gate with vck alias
	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vck", "gate", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 for vck gate --help, got %d", code)
	}

	// 3. gate on fresh comparison directory
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, core.ResearchDir), 0o755)
	meta := core.WorkspaceMeta{Scope: core.ScopeComparison}
	metaBytes, _ := json.Marshal(meta)
	_ = os.WriteFile(filepath.Join(dir, core.MetadataFile), metaBytes, 0o644)

	stdout.Reset()
	stderr.Reset()
	code = runCLI([]string{"vck", "gate", "--format", "json", dir}, &stdout, &stderr)
	if code != 0 {
		t.Errorf("expected exit code 0 on comparison workspace, got %d", code)
	}
	if !strings.Contains(stdout.String(), `"status": "PASS"`) {
		t.Errorf("expected status PASS in stdout, got: %s", stdout.String())
	}
}

func TestVersionParityAcrossEcosystem(t *testing.T) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(filepath.Join(root, "go.mod")); os.IsNotExist(err) {
		t.Skip("skipping version parity test in non-source environment")
		return
	}

	expected := version

	// 1. npm/vck/package.json
	vckPkgPath := filepath.Join(root, "npm", "vck", "package.json")
	if data, err := os.ReadFile(vckPkgPath); err == nil {
		var pkg struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(data, &pkg); err != nil {
			t.Errorf("parsing %s: %v", vckPkgPath, err)
		} else if pkg.Version != expected {
			t.Errorf("%s version %q != %q", vckPkgPath, pkg.Version, expected)
		}
	} else {
		t.Errorf("reading %s: %v", vckPkgPath, err)
	}

	// 2. npm/vivechak/package.json
	vivPkgPath := filepath.Join(root, "npm", "vivechak", "package.json")
	if data, err := os.ReadFile(vivPkgPath); err == nil {
		var pkg struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(data, &pkg); err != nil {
			t.Errorf("parsing %s: %v", vivPkgPath, err)
		} else if pkg.Version != expected {
			t.Errorf("%s version %q != %q", vivPkgPath, pkg.Version, expected)
		}
	} else {
		t.Errorf("reading %s: %v", vivPkgPath, err)
	}

	// 3. npm/vivechak/bin/runner.js
	runnerPath := filepath.Join(root, "npm", "vivechak", "bin", "runner.js")
	if data, err := os.ReadFile(runnerPath); err == nil {
		needle := "const VERSION = '" + expected + "';"
		if !strings.Contains(string(data), needle) {
			t.Errorf("%s does not contain %q", runnerPath, needle)
		}
	} else {
		t.Errorf("reading %s: %v", runnerPath, err)
	}

	// 4. pypi/vck/pyproject.toml
	pypiVckPath := filepath.Join(root, "pypi", "vck", "pyproject.toml")
	if data, err := os.ReadFile(pypiVckPath); err == nil {
		needle := `version = "` + expected + `"`
		if !strings.Contains(string(data), needle) {
			t.Errorf("%s does not contain %q", pypiVckPath, needle)
		}
	} else {
		t.Errorf("reading %s: %v", pypiVckPath, err)
	}

	// 5. pypi/vivechak/pyproject.toml
	pypiVivPath := filepath.Join(root, "pypi", "vivechak", "pyproject.toml")
	if data, err := os.ReadFile(pypiVivPath); err == nil {
		needle := `version = "` + expected + `"`
		if !strings.Contains(string(data), needle) {
			t.Errorf("%s does not contain %q", pypiVivPath, needle)
		}
	} else {
		t.Errorf("reading %s: %v", pypiVivPath, err)
	}

	// 6. Makefile
	makefilePath := filepath.Join(root, "Makefile")
	if data, err := os.ReadFile(makefilePath); err == nil {
		needle := "VERSION ?= " + expected
		if !strings.Contains(string(data), needle) {
			t.Errorf("%s does not contain %q", makefilePath, needle)
		}
	} else {
		t.Errorf("reading %s: %v", makefilePath, err)
	}

	// 7. pypi/vivechak/src/vivechak/cli.py
	pyCliPath := filepath.Join(root, "pypi", "vivechak", "src", "vivechak", "cli.py")
	if data, err := os.ReadFile(pyCliPath); err == nil {
		needle := `VERSION = "` + expected + `"`
		if !strings.Contains(string(data), needle) {
			t.Errorf("%s does not contain %q", pyCliPath, needle)
		}
	} else {
		t.Errorf("reading %s: %v", pyCliPath, err)
	}

	// 8. pypi/vivechak/src/vivechak/__init__.py
	pyVivInitPath := filepath.Join(root, "pypi", "vivechak", "src", "vivechak", "__init__.py")
	if data, err := os.ReadFile(pyVivInitPath); err == nil {
		needle := `__version__ = "` + expected + `"`
		if !strings.Contains(string(data), needle) {
			t.Errorf("%s does not contain %q", pyVivInitPath, needle)
		}
	} else {
		t.Errorf("reading %s: %v", pyVivInitPath, err)
	}

	// 9. pypi/vck/src/vck/__init__.py
	pyVckInitPath := filepath.Join(root, "pypi", "vck", "src", "vck", "__init__.py")
	if data, err := os.ReadFile(pyVckInitPath); err == nil {
		needle := `__version__ = "` + expected + `"`
		if !strings.Contains(string(data), needle) {
			t.Errorf("%s does not contain %q", pyVckInitPath, needle)
		}
	} else {
		t.Errorf("reading %s: %v", pyVckInitPath, err)
	}

	// 10. Dependency specs: npm/vck and pypi/vck must declare matching dependencies on vivechak
	if data, err := os.ReadFile(vckPkgPath); err == nil {
		depNeedle := `"vivechak": "^` + expected + `"`
		if !strings.Contains(string(data), depNeedle) {
			t.Errorf("%s does not contain dependency %q", vckPkgPath, depNeedle)
		}
	}
	if data, err := os.ReadFile(pypiVckPath); err == nil {
		depNeedle := `"vivechak~=` + expected + `"`
		if !strings.Contains(string(data), depNeedle) {
			t.Errorf("%s does not contain dependency %q", pypiVckPath, depNeedle)
		}
	}
}

func TestReleaseWorkflowSedPatterns(t *testing.T) {
	root := filepath.Join("..", "..")
	releaseYmlPath := filepath.Join(root, ".github", "workflows", "release.yml")
	data, err := os.ReadFile(releaseYmlPath)
	if err != nil {
		t.Skip("skipping release workflow test in non-source environment")
		return
	}
	releaseYml := string(data)

	// Ensure release.yml uses the PEP 440 compatible ~= clause with [=><~] character class
	expectedPyPISed := `sed -i "s/\"vivechak[=><~].*\"/\"vivechak~=$TAG_VERSION\"/" pypi/vck/pyproject.toml`
	if !strings.Contains(releaseYml, expectedPyPISed) {
		t.Errorf(".github/workflows/release.yml does not contain expected PyPI sed replacement: %q", expectedPyPISed)
	}

	// Verify that the PyPI sed regex actually matches the line in pypi/vck/pyproject.toml
	pypiVckData, err := os.ReadFile(filepath.Join(root, "pypi", "vck", "pyproject.toml"))
	if err != nil {
		t.Fatalf("reading pypi/vck/pyproject.toml: %v", err)
	}
	rePyPI := regexp.MustCompile(`"vivechak[=><~].*"`)
	if !rePyPI.Match(pypiVckData) {
		t.Errorf("PyPI regex %s does not match content in pypi/vck/pyproject.toml", rePyPI.String())
	}

	// Verify that the npm sed regex actually matches the line in npm/vck/package.json
	npmVckData, err := os.ReadFile(filepath.Join(root, "npm", "vck", "package.json"))
	if err != nil {
		t.Fatalf("reading npm/vck/package.json: %v", err)
	}
	reNPM := regexp.MustCompile(`"vivechak":\s*".*"`)
	if !reNPM.Match(npmVckData) {
		t.Errorf("npm regex %s does not match content in npm/vck/package.json", reNPM.String())
	}
}

func TestRepositoryReleaseReadiness(t *testing.T) {
	root := filepath.Join("..", "..")

	// 1. Makefile must use LF line endings and be configured with eol=lf in .gitattributes
	makefilePath := filepath.Join(root, "Makefile")
	data, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Skip("skipping repository release readiness test in non-source environment")
		return
	}
	if strings.Contains(string(data), "\r") {
		t.Errorf("Makefile contains CRLF line endings (\\r), which breaks GNU Make on Linux/macOS")
	}

	gitAttrPath := filepath.Join(root, ".gitattributes")
	if data, err := os.ReadFile(gitAttrPath); err == nil {
		if !strings.Contains(string(data), "Makefile text eol=lf") {
			t.Errorf(".gitattributes missing 'Makefile text eol=lf' rule")
		}
	} else {
		t.Errorf("reading .gitattributes: %v", err)
	}

	// 2. README.md must use the dynamic GitHub Actions badge, not a static badge
	readmePath := filepath.Join(root, "README.md")
	if data, err := os.ReadFile(readmePath); err == nil {
		readme := string(data)
		if strings.Contains(readme, "tests-100%25%20passing-brightgreen.svg") {
			t.Errorf("README.md contains hardcoded static test badge")
		}
		expectedBadge := "actions/workflows/ci.yml/badge.svg"
		if !strings.Contains(readme, expectedBadge) {
			t.Errorf("README.md missing dynamic CI status badge %q", expectedBadge)
		}
	} else {
		t.Errorf("reading README.md: %v", err)
	}
}



