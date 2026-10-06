package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectWorkspace_ScopePrecedence(t *testing.T) {
	t.Run("valid metadata scope takes precedence over pipeline presence", func(t *testing.T) {
		tmpDir := t.TempDir()
		researchDir := filepath.Join(tmpDir, ResearchDir)
		if err := os.MkdirAll(researchDir, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}

		// Write metadata with decision scope
		metaContent := `{"scope":"decision","created_at":"2026-09-29T00:00:00Z"}`
		if err := os.WriteFile(filepath.Join(tmpDir, MetadataFile), []byte(metaContent), 0o644); err != nil {
			t.Fatalf("write metadata: %v", err)
		}

		// Also create RESEARCH-PIPELINE.md (which previously would override metadata)
		if err := os.WriteFile(filepath.Join(tmpDir, PipelineFile), []byte("# Pipeline"), 0o644); err != nil {
			t.Fatalf("write pipeline: %v", err)
		}

		info := InspectWorkspace(tmpDir)
		if info.Scope != ScopeDecision {
			t.Errorf("expected ScopeDecision from metadata, got %q", info.Scope)
		}
		if !info.HasPipeline {
			t.Error("expected HasPipeline=true")
		}
	})

	t.Run("pipeline presence used as fallback when metadata scope is absent", func(t *testing.T) {
		tmpDir := t.TempDir()
		researchDir := filepath.Join(tmpDir, ResearchDir)
		if err := os.MkdirAll(researchDir, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}

		// Create pipeline without metadata file
		if err := os.WriteFile(filepath.Join(tmpDir, PipelineFile), []byte("# Pipeline"), 0o644); err != nil {
			t.Fatalf("write pipeline: %v", err)
		}

		info := InspectWorkspace(tmpDir)
		if info.Scope != ScopeProject {
			t.Errorf("expected ScopeProject inferred from pipeline, got %q", info.Scope)
		}
	})

	t.Run("invalid metadata scope falls back to pipeline presence", func(t *testing.T) {
		tmpDir := t.TempDir()
		researchDir := filepath.Join(tmpDir, ResearchDir)
		if err := os.MkdirAll(researchDir, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}

		// Metadata with invalid scope
		metaContent := `{"scope":"invalid-scope"}`
		if err := os.WriteFile(filepath.Join(tmpDir, MetadataFile), []byte(metaContent), 0o644); err != nil {
			t.Fatalf("write metadata: %v", err)
		}

		if err := os.WriteFile(filepath.Join(tmpDir, PipelineFile), []byte("# Pipeline"), 0o644); err != nil {
			t.Fatalf("write pipeline: %v", err)
		}

		info := InspectWorkspace(tmpDir)
		if info.Scope != ScopeProject {
			t.Errorf("expected ScopeProject fallback when metadata scope is invalid, got %q", info.Scope)
		}
	})

	t.Run("infers decision scope in manual workspace from plan file [HIGH-03]", func(t *testing.T) {
		tmpDir := t.TempDir()
		researchDir := filepath.Join(tmpDir, ResearchDir)
		if err := os.MkdirAll(researchDir, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}

		// Create D-001-plan.md without .vivechak.json
		if err := os.WriteFile(filepath.Join(researchDir, "D-001-plan.md"), []byte("# Decision Plan"), 0o644); err != nil {
			t.Fatalf("write decision plan: %v", err)
		}

		info := InspectWorkspace(tmpDir)
		if info.Scope != ScopeDecision {
			t.Errorf("expected ScopeDecision inferred from D-001-plan.md, got %q", info.Scope)
		}
		if !info.HasPipeline {
			t.Error("expected HasPipeline=true for decision plan")
		}
	})

	t.Run("infers comparison scope in manual workspace from comparison file [HIGH-03]", func(t *testing.T) {
		tmpDir := t.TempDir()
		researchDir := filepath.Join(tmpDir, ResearchDir)
		if err := os.MkdirAll(researchDir, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}

		// Create C-001-comparison.md without .vivechak.json
		if err := os.WriteFile(filepath.Join(researchDir, "C-001-comparison.md"), []byte("# Comparison"), 0o644); err != nil {
			t.Fatalf("write comparison: %v", err)
		}

		info := InspectWorkspace(tmpDir)
		if info.Scope != ScopeComparison {
			t.Errorf("expected ScopeComparison inferred from C-001-comparison.md, got %q", info.Scope)
		}
		if !info.HasPipeline {
			t.Error("expected HasPipeline=true for comparison plan")
		}
	})

	t.Run("infers decision scope in manual workspace from session prefix [HIGH-03]", func(t *testing.T) {
		tmpDir := t.TempDir()
		sessionsDir := filepath.Join(tmpDir, SessionsDir)
		if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}

		// Create sessions/S1-01.md
		if err := os.WriteFile(filepath.Join(sessionsDir, "S1-01.md"), []byte("# Session"), 0o644); err != nil {
			t.Fatalf("write session: %v", err)
		}

		info := InspectWorkspace(tmpDir)
		if info.Scope != ScopeDecision {
			t.Errorf("expected ScopeDecision inferred from sessions/S1-01.md, got %q", info.Scope)
		}
	})

	t.Run("resolves explicit research directory to parent workspace root", func(t *testing.T) {
		tmpDir := t.TempDir()
		researchDir := filepath.Join(tmpDir, ResearchDir)
		if err := os.MkdirAll(researchDir, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}

		resolved, err := ResolveWorkspace(researchDir)
		if err != nil {
			t.Fatalf("ResolveWorkspace failed: %v", err)
		}
		if filepath.Clean(resolved) != filepath.Clean(tmpDir) {
			t.Errorf("expected resolved path %q, got %q", tmpDir, resolved)
		}
	})
}

func TestIsSystemOrAppDir(t *testing.T) {
	cases := []struct {
		path     string
		expected bool
	}{
		{`C:\Users\air\AppData\Local\Programs\Antigravity IDE`, true},
		{`C:\Users\test\AppData\Local\Programs\Cursor`, true},
		{`C:\Program Files\SomeApp`, true},
		{`C:\Program Files (x86)\SomeApp`, true},
		{`C:\Windows\System32`, true},
		{`/Applications/Antigravity.app`, true},
		{`/System/Library`, true},
		{`/usr/bin`, true},
		{`/usr/local/bin`, true},
		{`/bin`, true},
		{`/sbin`, true},
		{`/etc`, true},
		{`/opt`, true},
		{`/opt/`, true},
		// Legitimate project & sandbox workspaces MUST NOT be blocked:
		{`/opt/my-app`, false},
		{`/opt/company/service`, false},
		{`/testbed`, false},
		{`/workspace`, false},
		{`/workspaces/my-repo`, false},
		{`/var/www/html`, false},
		{`d:\dev\lab\personal-finance-dashboard`, false},
		{`/home/user/projects/my-app`, false},
		{t.TempDir(), false},
	}

	for _, tc := range cases {
		actual := IsSystemOrAppDir(tc.path)
		if actual != tc.expected {
			t.Errorf("IsSystemOrAppDir(%q) = %v; want %v", tc.path, actual, tc.expected)
		}
	}
}

func TestResolveWorkspace_EnvironmentAndMacroGuard(t *testing.T) {
	t.Run("resolves VIVECHAK_PROJECT_ROOT", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VIVECHAK_PROJECT_ROOT", tmpDir)
		t.Setenv("VIVECHAK_DEFAULT_ROOT", "")

		res, err := ResolveWorkspace("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Clean(res) != filepath.Clean(tmpDir) {
			t.Errorf("got %q, want %q", res, tmpDir)
		}
	})

	t.Run("resolves VIVECHAK_DEFAULT_ROOT as fallback", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VIVECHAK_PROJECT_ROOT", "")
		t.Setenv("VIVECHAK_DEFAULT_ROOT", tmpDir)

		res, err := ResolveWorkspace("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Clean(res) != filepath.Clean(tmpDir) {
			t.Errorf("got %q, want %q", res, tmpDir)
		}
	})

	t.Run("resolves WORKSPACE standard container environment variable", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VIVECHAK_PROJECT_ROOT", "")
		t.Setenv("VIVECHAK_DEFAULT_ROOT", "")
		t.Setenv("WORKSPACE", tmpDir)
		t.Setenv("PROJECT_ROOT", "")

		res, err := ResolveWorkspace("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Clean(res) != filepath.Clean(tmpDir) {
			t.Errorf("got %q, want %q", res, tmpDir)
		}
	})

	t.Run("resolves PROJECT_ROOT container environment variable", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VIVECHAK_PROJECT_ROOT", "")
		t.Setenv("VIVECHAK_DEFAULT_ROOT", "")
		t.Setenv("WORKSPACE", "")
		t.Setenv("PROJECT_ROOT", tmpDir)

		res, err := ResolveWorkspace("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Clean(res) != filepath.Clean(tmpDir) {
			t.Errorf("got %q, want %q", res, tmpDir)
		}
	})

	t.Run("ignores unexpanded macro variable like ${workspaceFolder}", func(t *testing.T) {
		t.Setenv("VIVECHAK_PROJECT_ROOT", "${workspaceFolder}")
		t.Setenv("VIVECHAK_DEFAULT_ROOT", "${workspaceRoot}")

		// Should not resolve to ${workspaceFolder}
		// If in a non-workspace dir, should return error or walk up without matching ${...}
		res, _ := ResolveWorkspace("")
		if res == "${workspaceFolder}" || res == "${workspaceRoot}" {
			t.Errorf("expected macro variable to be ignored, but got %q", res)
		}
	})
}

