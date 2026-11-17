package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestConfigSafety(t *testing.T) {
	binPath := resolveOrBuildTestBinary(t)

	t.Run("ZedDryRunPrintsDiffWithoutSecrets", func(t *testing.T) {
		tempHome := t.TempDir()
		zedDir := filepath.Join(tempHome, ".config", "zed")
		if runtime.GOOS == "windows" {
			zedDir = filepath.Join(tempHome, "AppData", "Zed")
		}
		if err := os.MkdirAll(zedDir, 0o755); err != nil {
			t.Fatal(err)
		}
		zsrc := "// Zed settings\n{\n  // my theme\n  \"theme\": \"One Dark\",\n  \"context_servers\": {\n" +
			"    \"github\": { \"command\": { \"path\": \"gh-mcp\", \"args\": [\"--q\", \"a&b<c>\"], \"env\": { \"GITHUB_TOKEN\": \"FAKE-REDACTED\" } } }\n  }\n}\n"
		zPath := filepath.Join(zedDir, "settings.json")
		if err := os.WriteFile(zPath, []byte(zsrc), 0o600); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(binPath, "setup", "zed", "--dry-run")
		cmd.Env = append(os.Environ(),
			"HOME="+tempHome,
			"XDG_CONFIG_HOME="+filepath.Join(tempHome, ".config"),
			"APPDATA="+filepath.Join(tempHome, "AppData"),
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("setup zed --dry-run failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !regexp.MustCompile(`(?m)^\+.*vivechak`).MatchString(out) {
			t.Errorf("expected '+' line with 'vivechak' in dry-run diff, got:\n%s", out)
		}
		if strings.Contains(out, "FAKE-REDACTED") {
			t.Errorf("dry-run diff leaked secret token 'FAKE-REDACTED', got:\n%s", out)
		}

		// Ensure file was not modified
		after, err := os.ReadFile(zPath)
		if err != nil || string(after) != zsrc {
			t.Errorf("setup --dry-run modified the target file")
		}
	})

	t.Run("ClaudeCodeIdempotencyAndSingleBackup", func(t *testing.T) {
		tempHome := t.TempDir()
		orig := "{\n  \"zeta\": 1,\n  \"primaryApiKey\": \"FAKE-REDACTED\",\n  \"alpha\": 2,\n  \"mcpServers\": {}\n}\n"
		cPath := filepath.Join(tempHome, ".claude.json")
		if err := os.WriteFile(cPath, []byte(orig), 0o600); err != nil {
			t.Fatal(err)
		}

		runSetup := func() {
			cmd := exec.Command(binPath, "setup", "claude-code")
			cmd.Dir = tempHome
			cmd.Env = append(os.Environ(),
				"HOME="+tempHome,
				"USERPROFILE="+tempHome,
				"XDG_CONFIG_HOME="+filepath.Join(tempHome, ".config"),
				"APPDATA="+filepath.Join(tempHome, "AppData"),
			)
			if err := cmd.Run(); err != nil {
				t.Fatalf("setup claude-code failed: %v", err)
			}
		}

		// First run: modifies .claude.json, creates one backup
		runSetup()

		entries, err := os.ReadDir(tempHome)
		if err != nil {
			t.Fatal(err)
		}
		var baks []string
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".claude.json") && strings.Contains(e.Name(), "bak") {
				baks = append(baks, e.Name())
			}
		}
		if len(baks) != 1 {
			t.Fatalf("expected exactly 1 backup after first run, got %d: %v", len(baks), baks)
		}
		bakContent, err := os.ReadFile(filepath.Join(tempHome, baks[0]))
		if err != nil || string(bakContent) != orig {
			t.Errorf("backup does not match original content")
		}

		// Second run: file already has vivechak and is identical -> no-op, no new backup
		runSetup()

		entries2, err := os.ReadDir(tempHome)
		if err != nil {
			t.Fatal(err)
		}
		var baks2 []string
		for _, e := range entries2 {
			if strings.HasPrefix(e.Name(), ".claude.json") && strings.Contains(e.Name(), "bak") {
				baks2 = append(baks2, e.Name())
			}
		}
		if len(baks2) != 1 {
			t.Fatalf("expected still exactly 1 backup after second run, got %d: %v", len(baks2), baks2)
		}
		bakContent2, err := os.ReadFile(filepath.Join(tempHome, baks2[0]))
		if err != nil || string(bakContent2) != orig {
			t.Errorf("backup was corrupted after second run")
		}
	})

	t.Run("WindsurfFreshFileMode600", func(t *testing.T) {
		tempHome := t.TempDir()
		cmd := exec.Command(binPath, "setup", "windsurf")
		cmd.Dir = tempHome
		cmd.Env = append(os.Environ(),
			"HOME="+tempHome,
			"USERPROFILE="+tempHome,
			"XDG_CONFIG_HOME="+filepath.Join(tempHome, ".config"),
			"APPDATA="+filepath.Join(tempHome, "AppData"),
		)
		if err := cmd.Run(); err != nil {
			t.Fatalf("setup windsurf failed: %v", err)
		}

		// Path resolution varies by platform (e.g. .codeium/windsurf/mcp_config.json on Linux/macOS)
		var wPath string
		_ = filepath.Walk(tempHome, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && info.Name() == "mcp_config.json" {
				wPath = path
			}
			return nil
		})

		if wPath == "" {
			t.Fatalf("expected mcp_config.json to be created under tempHome")
		}

		if runtime.GOOS != "windows" {
			info, err := os.Stat(wPath)
			if err != nil {
				t.Fatal(err)
			}
			perm := info.Mode().Perm()
			if perm != 0o600 {
				t.Errorf("expected file mode 0600, got %o", perm)
			}
		}
	})

	// HIGH-11: a dry run over a very large ~/.claude.json must stay fast and print only the
	// vivechak entry (a whole-file diff ran out of memory on a 22,506-line file at cfe8877).
	t.Run("ClaudeCodeDryRunOn50000LineConfigIsFastAndFocused", func(t *testing.T) {
		tempHome := t.TempDir()
		var b strings.Builder
		b.WriteString("{\n  \"numStartups\": 412,\n  \"projects\": {\n")
		const projects = 10000 // 5 lines each, plus 6 around them: 50,006 lines
		for i := 0; i < projects; i++ {
			fmt.Fprintf(&b, "    \"/home/dev/project-%05d\": {\n", i)
			b.WriteString("      \"allowedTools\": [],\n")
			fmt.Fprintf(&b, "      \"history\": \"FILLER-LINE-%05d\",\n", i)
			b.WriteString("      \"lastCost\": 0.25\n")
			if i < projects-1 {
				b.WriteString("    },\n")
			} else {
				b.WriteString("    }\n")
			}
		}
		b.WriteString("  },\n  \"mcpServers\": {}\n}\n")
		src := b.String()
		if n := strings.Count(src, "\n"); n < 50000 {
			t.Fatalf("fixture has %d lines, want at least 50,000", n)
		}
		cPath := filepath.Join(tempHome, ".claude.json")
		if err := os.WriteFile(cPath, []byte(src), 0o600); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(binPath, "setup", "claude-code", "--dry-run")
		cmd.Dir = tempHome
		cmd.Env = append(os.Environ(),
			"HOME="+tempHome,
			"USERPROFILE="+tempHome,
			"XDG_CONFIG_HOME="+filepath.Join(tempHome, ".config"),
			"APPDATA="+filepath.Join(tempHome, "AppData"),
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		start := time.Now()
		err := cmd.Run()
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("setup claude-code --dry-run failed: %v, stderr: %s", err, stderr.String())
		}
		// The 2 s bound is on the dry run's own CPU time (user + system): it measures the work the dry run
		// does (the whole-file diff it replaced ran out of memory), and unlike wall time it doesn't grow
		// when a loaded CI machine delays the process. A hang still fails through the wall-clock cap.
		cpu := cmd.ProcessState.UserTime() + cmd.ProcessState.SystemTime()
		if cpu >= 2*time.Second || elapsed >= 60*time.Second {
			t.Errorf("dry run over a 50,000-line ~/.claude.json used %v CPU (%v wall), want under 2s CPU", cpu, elapsed)
		}
		out := stdout.String()
		if !regexp.MustCompile(`(?m)^\+.*vivechak`).MatchString(out) {
			t.Errorf("expected a '+' line with 'vivechak' in the dry-run diff, got:\n%.2000s", out)
		}
		if strings.Contains(out, "FILLER-LINE-") || strings.Count(out, "\n") > 200 {
			t.Errorf("dry run printed the untouched lines of the file (%d lines), want only the vivechak entry:\n%.2000s",
				strings.Count(out, "\n"), out)
		}
		if after, err := os.ReadFile(cPath); err != nil || string(after) != src {
			t.Errorf("setup --dry-run modified ~/.claude.json")
		}
	})

	// HIGH-11: a dry run prints no value of another server, whatever its key is called;
	// redaction by key name alone would leak all three values below.
	t.Run("DryRunPrintsNoValueOfOtherServers", func(t *testing.T) {
		secrets := []string{"FAKE-ARGS-TOKEN-1111", "FAKE-DB-PASSWORD-2222", "FAKE-PLAIN-VALUE-3333"}
		other := `{"command":"db-mcp","args":["--token","FAKE-ARGS-TOKEN-1111"],` +
			`"env":{"DATABASE_URL":"postgres://admin:FAKE-DB-PASSWORD-2222@db.internal:5432/prod","REGION_HINT":"FAKE-PLAIN-VALUE-3333"}}`
		cases := []struct {
			host, file, src string
		}{
			{"claude-code", ".claude.json", `{"numStartups":3,"mcpServers":{"db":` + other + `}}` + "\n"},
			{"vscode", filepath.Join(".vscode", "mcp.json"), `{"servers":{"db":` + other + `}}` + "\n"},
		}
		for _, tc := range cases {
			tempHome := t.TempDir()
			cfg := filepath.Join(tempHome, tc.file)
			if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(cfg, []byte(tc.src), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(binPath, "setup", tc.host, "--dry-run")
			cmd.Dir = tempHome
			cmd.Env = append(os.Environ(),
				"HOME="+tempHome,
				"USERPROFILE="+tempHome,
				"XDG_CONFIG_HOME="+filepath.Join(tempHome, ".config"),
				"APPDATA="+filepath.Join(tempHome, "AppData"),
			)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("setup %s --dry-run failed: %v, stderr: %s", tc.host, err, stderr.String())
			}
			out := stdout.String()
			if !regexp.MustCompile(`(?m)^\+.*vivechak`).MatchString(out) {
				t.Errorf("%s: expected a '+' line with 'vivechak' in the dry-run diff, got:\n%s", tc.host, out)
			}
			for _, s := range secrets {
				if strings.Contains(out, s) || strings.Contains(stderr.String(), s) {
					t.Errorf("%s: dry run printed another server's value %q:\nstdout:\n%s\nstderr:\n%s", tc.host, s, out, stderr.String())
				}
			}
		}
	})

	// HIGH-11: the dry run still extracts the vivechak entry from a JSON config that keeps its
	// servers under Goose-style "extensions", and from a YAML config (where it falls back to the
	// whole file, the other extension's value would be printed).
	t.Run("DryRunExtractsEntryFromJSONExtensionsAndYAML", func(t *testing.T) {
		const otherValue = "FAKE-EXT-VALUE-4444"
		cases := []struct {
			name, file, src string
		}{
			{"json-extensions", "agent.json",
				`{"extensions":{"vivechak":{"name":"vivechak","type":"stdio","cmd":"/old/vck","args":["serve"],"enabled":true},` +
					`"other":{"name":"other","type":"stdio","cmd":"other-mcp","envs":{"REGION_HINT":"` + otherValue + `"}}}}` + "\n"},
			{"yaml", filepath.Join("goose", "config.yaml"),
				"extensions:\n  other:\n    name: other\n    type: stdio\n    cmd: other-mcp\n    envs:\n      REGION_HINT: " + otherValue + "\n"},
		}
		for _, tc := range cases {
			tempHome := t.TempDir()
			cfg := filepath.Join(tempHome, tc.file)
			if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(cfg, []byte(tc.src), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(binPath, "setup", cfg, "--dry-run")
			cmd.Dir = tempHome
			cmd.Env = append(os.Environ(),
				"HOME="+tempHome,
				"USERPROFILE="+tempHome,
				"XDG_CONFIG_HOME="+filepath.Join(tempHome, ".config"),
				"APPDATA="+filepath.Join(tempHome, "AppData"),
			)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("%s: setup <path> --dry-run failed: %v, stderr: %s", tc.name, err, stderr.String())
			}
			out := stdout.String()
			if !regexp.MustCompile(`(?m)^\+.*vivechak`).MatchString(out) {
				t.Errorf("%s: expected a '+' line with 'vivechak' in the dry-run diff, got:\n%s", tc.name, out)
			}
			if strings.Contains(out, otherValue) {
				t.Errorf("%s: dry run printed the other extension's value, so the vivechak entry was not extracted:\n%s", tc.name, out)
			}
			if tc.name == "json-extensions" &&
				(strings.Contains(out, "(vivechak not configured)") || !regexp.MustCompile(`(?m)^- .*vivechak`).MatchString(out)) {
				t.Errorf("%s: expected the existing vivechak entry under \"extensions\" as '-' lines, got:\n%s", tc.name, out)
			}
		}
	})
}
