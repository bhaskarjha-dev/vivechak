package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tailscale/hujson"
	"gopkg.in/yaml.v3"
)

// presetLocation is one config-file location a host documents (URL on the preset in presets.go).
type presetLocation struct {
	goos   string // "" = the same on linux, darwin and windows
	marker string // created in cwd before resolving; a trailing "/" makes it a directory
	want   func(home, appData, cwd string) string
}

// TestPresetFixtures asserts, for every kept preset, the exact file and key its host documents,
// then runs `vck setup <preset>` against a fixture home and checks where the entry landed.
func TestPresetFixtures(t *testing.T) {
	cases := []struct {
		preset    string
		key       string
		locations []presetLocation
	}{
		{"cursor", "mcpServers", []presetLocation{
			{"", "", func(h, a, c string) string { return filepath.Join(h, ".cursor", "mcp.json") }},
			{"", ".cursor/", func(h, a, c string) string { return filepath.Join(c, ".cursor", "mcp.json") }},
		}},
		{"vscode", "servers", []presetLocation{
			{"", "", func(h, a, c string) string { return filepath.Join(c, ".vscode", "mcp.json") }},
		}},
		{"claude-desktop", "mcpServers", []presetLocation{
			{"darwin", "", func(h, a, c string) string {
				return filepath.Join(h, "Library", "Application Support", "Claude", "claude_desktop_config.json")
			}},
			{"windows", "", func(h, a, c string) string { return filepath.Join(a, "Claude", "claude_desktop_config.json") }},
			{"linux", "", func(h, a, c string) string {
				return filepath.Join(h, ".config", "Claude", "claude_desktop_config.json")
			}},
		}},
		{"antigravity", "mcpServers", []presetLocation{
			{"", "", func(h, a, c string) string { return filepath.Join(h, ".gemini", "config", "mcp_config.json") }},
			{"", ".agents/", func(h, a, c string) string { return filepath.Join(c, ".agents", "mcp_config.json") }},
		}},
		{"windsurf", "mcpServers", []presetLocation{
			{"", "", func(h, a, c string) string { return filepath.Join(h, ".codeium", "windsurf", "mcp_config.json") }},
		}},
		{"zed", "context_servers", []presetLocation{
			{"linux", "", func(h, a, c string) string { return filepath.Join(h, ".config", "zed", "settings.json") }},
			{"darwin", "", func(h, a, c string) string { return filepath.Join(h, ".config", "zed", "settings.json") }},
			{"windows", "", func(h, a, c string) string { return filepath.Join(a, "Zed", "settings.json") }},
		}},
		{"roo", "mcpServers", []presetLocation{
			{"", "", func(h, a, c string) string { return filepath.Join(c, ".roo", "mcp.json") }},
		}},
		{"claude-code", "mcpServers", []presetLocation{
			{"", "", func(h, a, c string) string { return filepath.Join(h, ".claude.json") }},
			{"", ".mcp.json", func(h, a, c string) string { return filepath.Join(c, ".mcp.json") }},
		}},
		{"opencode", "mcp", []presetLocation{
			{"", "", func(h, a, c string) string { return filepath.Join(h, ".config", "opencode", "opencode.json") }},
			{"", "opencode.json", func(h, a, c string) string { return filepath.Join(c, "opencode.json") }},
			{"", ".opencode/", func(h, a, c string) string { return filepath.Join(c, ".opencode", "opencode.json") }},
		}},
		{"goose", "extensions", []presetLocation{
			{"linux", "", func(h, a, c string) string { return filepath.Join(h, ".config", "goose", "config.yaml") }},
			{"darwin", "", func(h, a, c string) string { return filepath.Join(h, ".config", "goose", "config.yaml") }},
			{"windows", "", func(h, a, c string) string { return filepath.Join(a, "Block", "goose", "config", "config.yaml") }},
		}},
	}

	if len(cases) != len(hostPresets) {
		t.Fatalf("%d fixture cases for %d presets: every preset needs its documented location here", len(cases), len(hostPresets))
	}

	for _, tc := range cases {
		t.Run(tc.preset, func(t *testing.T) {
			p := findPreset(tc.preset)
			if p == nil {
				t.Fatalf("preset %q is not accepted", tc.preset)
			}
			if p.ServerKey != tc.key {
				t.Errorf("%s: server key %q, want %q", tc.preset, p.ServerKey, tc.key)
			}

			// 1. Every documented location, on every OS the docs name.
			var writeWant func(home, appData, cwd string) string
			for _, loc := range tc.locations {
				oses := []string{loc.goos}
				if loc.goos == "" {
					oses = []string{"linux", "darwin", "windows"}
				}
				for _, goos := range oses {
					home, cwd := t.TempDir(), t.TempDir()
					appData := filepath.Join(home, "AppData", "Roaming")
					if loc.marker != "" {
						makeMarker(t, cwd, loc.marker)
					}
					got, err := resolveConfigPath(tc.preset, home, appData, cwd, goos)
					if want := loc.want(home, appData, cwd); err != nil || got != want {
						t.Errorf("%s on %s (marker %q): got %q (err %v), want %q", tc.preset, goos, loc.marker, got, err, want)
					}
				}
				if loc.marker == "" && (loc.goos == "" || loc.goos == runtime.GOOS) {
					writeWant = loc.want
				}
			}
			if writeWant == nil {
				t.Fatalf("%s: no documented location for %s", tc.preset, runtime.GOOS)
			}

			// 2. Fixture write on this OS: the entry lands in that file, under that key.
			home, cwd := t.TempDir(), t.TempDir()
			appData := filepath.Join(home, "AppData", "Roaming")
			prev := testHomeDir
			testHomeDir = home
			t.Cleanup(func() { testHomeDir = prev })
			t.Setenv("APPDATA", appData)
			t.Chdir(cwd)
			var stdout, stderr bytes.Buffer
			if code := runSetupWithArgs([]string{tc.preset}, &stdout, &stderr); code != 0 {
				t.Fatalf("setup %s exited %d: %s", tc.preset, code, stderr.String())
			}
			path := writeWant(home, appData, cwd)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("setup %s did not write %s: %v (stdout: %s)", tc.preset, path, err, stdout.String())
			}
			var root map[string]any
			if strings.HasSuffix(path, ".yaml") {
				err = yaml.Unmarshal(data, &root)
			} else {
				var std []byte
				if std, err = hujson.Standardize(data); err == nil {
					err = json.Unmarshal(std, &root)
				}
			}
			if err != nil {
				t.Fatalf("%s: %s is not parseable: %v\n%s", tc.preset, path, err, data)
			}
			servers, _ := root[tc.key].(map[string]any)
			if servers["vivechak"] == nil {
				t.Errorf("%s: %s has no %s.vivechak entry:\n%s", tc.preset, path, tc.key, data)
			}
		})
	}
}

func makeMarker(t *testing.T, dir, marker string) {
	t.Helper()
	if strings.HasSuffix(marker, "/") {
		if err := os.MkdirAll(filepath.Join(dir, strings.TrimSuffix(marker, "/")), 0o755); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := os.WriteFile(filepath.Join(dir, marker), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// isolateLauncherEnv empties PATH and points every directory resolveLauncher consults at a
// fresh temporary home, so only the binary's own location decides the launcher.
func isolateLauncherEnv(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("PATH", t.TempDir())
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("VIVECHAK_LAUNCHER", "")
	return home
}

func writeFakeBinary(t *testing.T, path string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// A source build that is not on PATH is written by its absolute path (never npx), with a warning.
func TestResolveLauncher_SourceBuildNotOnPATH(t *testing.T) {
	home := isolateLauncherEnv(t)
	exe := writeFakeBinary(t, filepath.Join(home, "code", "vivechak", "vck"))
	cmd, args := resolveLauncher(exe)
	if cmd != exe || len(args) != 1 || args[0] != "serve" {
		t.Errorf("resolveLauncher(%q) = %q %v, want the absolute path and [serve]", exe, cmd, args)
	}
	if w := launcherWarning(exe); !strings.Contains(w, exe) || !strings.Contains(w, "not on your PATH") {
		t.Errorf("expected a not-on-PATH warning naming %s, got %q", exe, w)
	}
}

// A Homebrew Cellar or Caskroom binary that is not on PATH gets its absolute path plus a
// warning that upgrades may break it (it used to get `npx -y vivechak serve`).
func TestResolveLauncher_HomebrewNotOnPATH(t *testing.T) {
	root := isolateLauncherEnv(t)
	for _, rel := range [][]string{
		{"Cellar", "vivechak", "0.1.0", "bin", "vck"},
		{"opt", "homebrew", "Cellar", "vivechak", "0.1.0", "bin", "vck"},
		{"usr", "local", "Caskroom", "vivechak", "0.1.0", "vck"},
	} {
		exe := writeFakeBinary(t, filepath.Join(append([]string{root}, rel...)...))
		cmd, args := resolveLauncher(exe)
		if cmd != exe || len(args) != 1 || args[0] != "serve" {
			t.Errorf("resolveLauncher(%q) = %q %v, want the absolute path and [serve]", exe, cmd, args)
		}
		if w := launcherWarning(exe); !strings.Contains(w, exe) || !strings.Contains(w, "Upgrading") {
			t.Errorf("expected an upgrade warning naming %s, got %q", exe, w)
		}
	}
}

// A Homebrew binary reached through the brew symlink on PATH is written by that stable PATH
// entry, with no warning.
func TestResolveLauncher_HomebrewOnPATH(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks needs extra privileges on Windows")
	}
	root := isolateLauncherEnv(t)
	exe := writeFakeBinary(t, filepath.Join(root, "opt", "homebrew", "Cellar", "vivechak", "0.1.0", "bin", "vck"))
	binDir := filepath.Join(root, "opt", "homebrew", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(binDir, "vck")
	if err := os.Symlink(exe, link); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	if cmd, args := resolveLauncher(exe); cmd != link || len(args) != 1 || args[0] != "serve" {
		t.Errorf("resolveLauncher(%q) = %q %v, want %q [serve]", exe, cmd, args, link)
	}
	if w := launcherWarning(exe); w != "" {
		t.Errorf("expected no warning for a binary on PATH, got %q", w)
	}

	// The same under the other lookup name: a "vivechak" symlink on PATH to the "vck" binary.
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	altLink := filepath.Join(binDir, "vivechak")
	if err := os.Symlink(exe, altLink); err != nil {
		t.Fatal(err)
	}
	if cmd, _ := resolveLauncher(exe); cmd != altLink {
		t.Errorf("resolveLauncher(%q) = %q, want the PATH entry %q", exe, cmd, altLink)
	}
	if w := launcherWarning(exe); w != "" {
		t.Errorf("expected no warning for a binary on PATH under the name vivechak, got %q", w)
	}
}

// A binary that is on PATH under its own name (as after install.sh copies vck into a PATH
// directory) is written by that PATH entry, which here equals its own absolute path, with no
// warning: the warning used to fire whenever the command equalled the binary's path.
func TestResolveLauncher_OnPATHUnderOwnName(t *testing.T) {
	for _, name := range []string{"vck", "vivechak"} {
		root := isolateLauncherEnv(t)
		binDir := filepath.Join(root, "usr", "local", "bin")
		exe := writeFakeBinary(t, filepath.Join(binDir, name))
		t.Setenv("PATH", binDir)
		cmd, args, absoluteFallback := resolveLauncherDetail(exe)
		if cmd != exe || len(args) != 1 || args[0] != "serve" || absoluteFallback {
			t.Errorf("%s: resolveLauncherDetail(%q) = %q %v fallback=%v, want the PATH entry %q [serve] and no fallback",
				name, exe, cmd, args, absoluteFallback, exe)
		}
		if w := launcherWarning(exe); w != "" {
			t.Errorf("%s: expected no warning for a binary on PATH, got %q", name, w)
		}
	}
}

// The built binary on PATH under its own name runs `setup claude-code` without the
// not-on-PATH warning and writes its PATH entry (the README quickstart path).
func TestLauncher_BinaryOnPATH_LiveSubprocess(t *testing.T) {
	src := resolveOrBuildTestBinary(t)
	home := t.TempDir()
	binDir := filepath.Join(home, "bin")
	name := "vck"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	exe := filepath.Join(binDir, name)
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe, data, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(exe, "setup", "claude-code")
	cmd.Dir = home
	cmd.Env = isolatedSetupEnv(home, binDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("setup claude-code failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	if strings.Contains(stderr.String(), "not on your PATH") {
		t.Errorf("setup warned that a binary on PATH is not on PATH: %q", stderr.String())
	}
	cfg, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		t.Fatalf("reading ~/.claude.json: %v", err)
	}
	var parsed struct {
		MCPServers map[string]struct {
			Command string `json:"command"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(cfg, &parsed); err != nil {
		t.Fatalf("~/.claude.json is not JSON: %v\n%s", err, cfg)
	}
	gotPath, _ := filepath.EvalSymlinks(parsed.MCPServers["vivechak"].Command)
	wantPath, _ := filepath.EvalSymlinks(exe)
	if gotPath == "" || gotPath != wantPath {
		t.Errorf("vivechak command = %q, want the PATH entry %q\n%s", parsed.MCPServers["vivechak"].Command, exe, cfg)
	}
}

// isolatedSetupEnv is os.Environ with PATH set to pathDir and every home/config/cache
// directory pointed into home.
func isolatedSetupEnv(home, pathDir string) []string {
	var env []string
	for _, kv := range os.Environ() {
		k := strings.SplitN(kv, "=", 2)[0]
		switch strings.ToUpper(k) {
		case "PATH", "HOME", "USERPROFILE", "APPDATA", "LOCALAPPDATA", "XDG_CONFIG_HOME", "XDG_CACHE_HOME", "VIVECHAK_LAUNCHER":
			continue
		}
		env = append(env, kv)
	}
	return append(env,
		"PATH="+pathDir,
		"HOME="+home,
		"USERPROFILE="+home,
		"APPDATA="+filepath.Join(home, "AppData", "Roaming"),
		"LOCALAPPDATA="+filepath.Join(home, "AppData", "Local"),
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"XDG_CACHE_HOME="+filepath.Join(home, ".cache"),
	)
}

// The built binary, copied to a Homebrew Cellar path that is not on PATH, writes that absolute
// path into the Cursor config and warns on stderr.
func TestLauncher_HomebrewBinaryNotOnPATH_LiveSubprocess(t *testing.T) {
	src := resolveOrBuildTestBinary(t)
	home := t.TempDir()
	name := "vck"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	exe := filepath.Join(home, "Cellar", "vivechak", "0.1.0", "bin", name)
	if err := os.MkdirAll(filepath.Dir(exe), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe, data, 0o755); err != nil {
		t.Fatal(err)
	}

	var env []string
	for _, kv := range os.Environ() {
		k := strings.SplitN(kv, "=", 2)[0]
		switch strings.ToUpper(k) {
		case "PATH", "HOME", "USERPROFILE", "APPDATA", "LOCALAPPDATA", "XDG_CONFIG_HOME", "XDG_CACHE_HOME", "VIVECHAK_LAUNCHER":
			continue
		}
		env = append(env, kv)
	}
	env = append(env,
		"PATH="+t.TempDir(),
		"HOME="+home,
		"USERPROFILE="+home,
		"APPDATA="+filepath.Join(home, "AppData", "Roaming"),
		"LOCALAPPDATA="+filepath.Join(home, "AppData", "Local"),
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"XDG_CACHE_HOME="+filepath.Join(home, ".cache"),
	)
	cmd := exec.Command(exe, "setup", "cursor")
	cmd.Dir = home
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("setup cursor failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	cfg, err := os.ReadFile(filepath.Join(home, ".cursor", "mcp.json"))
	if err != nil {
		t.Fatalf("reading ~/.cursor/mcp.json: %v", err)
	}
	var parsed struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(cfg, &parsed); err != nil {
		t.Fatalf("~/.cursor/mcp.json is not JSON: %v\n%s", err, cfg)
	}
	entry := parsed.MCPServers["vivechak"]
	gotPath, _ := filepath.EvalSymlinks(entry.Command)
	wantPath, _ := filepath.EvalSymlinks(exe)
	if gotPath == "" || gotPath != wantPath || len(entry.Args) != 1 || entry.Args[0] != "serve" {
		t.Errorf("vivechak entry = %q %v, want %q [serve]\n%s", entry.Command, entry.Args, exe, cfg)
	}
	if !strings.Contains(stderr.String(), "not on your PATH") || !strings.Contains(stderr.String(), "Upgrading") {
		t.Errorf("expected the upgrade warning on stderr, got: %q", stderr.String())
	}
}
