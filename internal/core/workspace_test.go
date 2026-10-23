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
		// Linux runners where filepath.Abs prepends CWD to Windows paths:
		{`/home/runner/work/vivechak/vivechak/internal/mcp/C:\Users\air\AppData\Local\Programs\Antigravity IDE`, true},
		{`/home/runner/work/vivechak/vivechak/internal/mcp/C:\Program Files\SomeApp`, true},
		{`/home/runner/work/vivechak/vivechak/internal/mcp/C:/Program Files/SomeApp`, true},
		// Legitimate project & sandbox workspaces MUST NOT be blocked:
		{`/opt/my-app`, false},
		{`/opt/company/service`, false},
		{`/testbed`, false},
		{`/workspace`, false},
		{`/workspaces/my-repo`, false},
		{`/var/www/html`, false},
		{`d:\dev\lab\personal-finance-dashboard`, false},
		{`d:\windows-tools`, false},
		{`c:\binary-tree`, false},
		{`/library-books`, false},
		{`/system-monitor`, false},
		{`/etc-hosts-manager`, false},
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

func TestResolveWorkspace_LastResolvedMemory(t *testing.T) {
	wsDir := t.TempDir()
	researchDir := filepath.Join(wsDir, ResearchDir)
	if err := os.MkdirAll(researchDir, 0o755); err != nil {
		t.Fatalf("mkdir research failed: %v", err)
	}

	// 1. Resolve explicitly to prime in-process memory
	resolved, err := ResolveWorkspace(wsDir)
	if err != nil {
		t.Fatalf("ResolveWorkspace explicit: %v", err)
	}
	if filepath.Clean(resolved) != filepath.Clean(wsDir) {
		t.Fatalf("got %s, want %s", resolved, wsDir)
	}

	// 2. Switch CWD to an outside directory with no research/ dir
	outsideDir := t.TempDir()
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(outsideDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(origCwd)
		SetLastResolvedWorkspace("")
	}()

	t.Setenv("VIVECHAK_PROJECT_ROOT", "")
	t.Setenv("VIVECHAK_DEFAULT_ROOT", "")
	t.Setenv("WORKSPACE", "")
	t.Setenv("PROJECT_ROOT", "")

	// 3. Resolve with empty string from outside CWD -> should hit in-process memory
	resFromMemory, err := ResolveWorkspace("")
	if err != nil {
		t.Fatalf("expected ResolveWorkspace to succeed via in-process memory, got error: %v", err)
	}
	if filepath.Clean(resFromMemory) != filepath.Clean(wsDir) {
		t.Errorf("expected %s from in-process memory, got %s", wsDir, resFromMemory)
	}

	// 4. When in-process memory is cleared, resolution should fail
	SetLastResolvedWorkspace("")
	_, err = ResolveWorkspace("")
	if err == nil {
		t.Errorf("expected error after clearing session memory when CWD is outside workspace")
	}
}

func TestScanCompletedSessions_IgnoresDraft(t *testing.T) {
	dir := t.TempDir()
	ws := &testWorkspaceFS{root: dir}
	_ = os.MkdirAll(filepath.Join(dir, SessionsDir), 0o755)

	// Save completed session
	completedContent := "---\nsession_id: T0-01\nstatus: completed\n---\n# Findings"
	_ = os.WriteFile(filepath.Join(dir, SessionsDir, "T0-01.md"), []byte(completedContent), 0o644)

	// Save draft session with blocking issues
	draftContent := "---\nsession_id: T0-02\nstatus: draft\n---\n# Draft"
	_ = os.WriteFile(filepath.Join(dir, SessionsDir, "T0-02.md"), []byte(draftContent), 0o644)

	completed, err := ScanCompletedSessions(ws, "T0-01", "T0-02")
	if err != nil {
		t.Fatalf("ScanCompletedSessions failed: %v", err)
	}

	if !completed["T0-01"] {
		t.Errorf("expected T0-01 to be completed")
	}
	if completed["T0-02"] {
		t.Errorf("expected draft session T0-02 NOT to be marked as completed")
	}

	// Draft FAD test
	draftFAD := "---\nstatus: draft\n---\n# FAD Draft"
	_ = os.WriteFile(filepath.Join(dir, FADFile), []byte(draftFAD), 0o644)
	completedFAD, err := ScanCompletedSessions(ws, "SYN-01")
	if err != nil {
		t.Fatalf("ScanCompletedSessions failed: %v", err)
	}
	if completedFAD["SYN-01"] || completedFAD["FAD"] {
		t.Errorf("expected draft FAD NOT to be marked as completed")
	}
}

func TestDirFS_ResolveAndOperations(t *testing.T) {
	tmpDir := t.TempDir()
	fs := DirFS(tmpDir)

	// 1. Stat on empty and dot paths
	fiEmpty, err := fs.Stat("")
	if err != nil {
		t.Fatalf("Stat('') failed: %v", err)
	}
	fiDot, err := fs.Stat(".")
	if err != nil {
		t.Fatalf("Stat('.') failed: %v", err)
	}
	if fiEmpty.Name() != fiDot.Name() {
		t.Errorf("expected same FileInfo for '' and '.', got %s vs %s", fiEmpty.Name(), fiDot.Name())
	}

	// 2. Write a test file and verify ReadFile and Stat
	sampleFile := "test-sample.md"
	samplePath := filepath.Join(tmpDir, sampleFile)
	if err := os.WriteFile(samplePath, []byte("# Test Content\n"), 0o644); err != nil {
		t.Fatalf("writing sample file: %v", err)
	}

	data, err := fs.ReadFile(sampleFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != "# Test Content\n" {
		t.Errorf("unexpected content: %s", string(data))
	}

	fiFile, err := fs.Stat(sampleFile)
	if err != nil {
		t.Fatalf("Stat on file failed: %v", err)
	}
	if fiFile.Size() != int64(len("# Test Content\n")) {
		t.Errorf("unexpected file size: %d", fiFile.Size())
	}

	// 3. ListDir on empty and dot paths
	entriesEmpty, err := fs.ListDir("")
	if err != nil {
		t.Fatalf("ListDir('') failed: %v", err)
	}
	if len(entriesEmpty) != 1 || entriesEmpty[0].Name() != sampleFile {
		t.Errorf("unexpected ListDir('') entries: %+v", entriesEmpty)
	}

	entriesDot, err := fs.ListDir(".")
	if err != nil {
		t.Fatalf("ListDir('.') failed: %v", err)
	}
	if len(entriesDot) != 1 || entriesDot[0].Name() != sampleFile {
		t.Errorf("unexpected ListDir('.') entries: %+v", entriesDot)
	}

	// 4. Non-existent path returns os.ErrNotExist
	_, err = fs.ReadFile("missing-file.md")
	if !os.IsNotExist(err) {
		t.Errorf("expected IsNotExist for missing file, got: %v", err)
	}

	// 5. dirFS with empty dir
	fsEmptyDir := DirFS("")
	fiCurr, err := fsEmptyDir.Stat(".")
	if err != nil {
		t.Fatalf("DirFS('').Stat('.') failed: %v", err)
	}
	if !fiCurr.IsDir() {
		t.Errorf("expected current dir to be directory")
	}
}


