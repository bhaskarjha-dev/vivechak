package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
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
}
