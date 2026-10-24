package workspace

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSafeRoot(t *testing.T) {
	// 1. Empty path rejected
	if err := SafeRoot(""); err == nil {
		t.Error("expected empty path to be rejected by SafeRoot")
	}

	// 2. Root directory rejected
	if err := SafeRoot("/"); err == nil {
		t.Error("expected '/' to be rejected by SafeRoot")
	}

	// 3. Drive root rejected on Windows
	if runtime.GOOS == "windows" {
		for _, drive := range []string{"C:\\", "D:\\", "c:", "C:"} {
			if err := SafeRoot(drive); err == nil {
				t.Errorf("expected drive root %q to be rejected", drive)
			}
		}
	}

	// 4. User home directory rejected directly
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if err := SafeRoot(home); err == nil {
			t.Errorf("expected user home %q to be rejected directly", home)
		}
	}

	// 5. System directories rejected
	for _, sys := range []string{
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\Program Files (x86)",
		"C:\\ProgramData",
		"/usr/bin",
		"/etc",
		"/bin",
		"/sbin",
	} {
		if err := SafeRoot(sys); err == nil {
			t.Errorf("expected system dir %q to be rejected", sys)
		}
	}

	// 6. Valid project directories accepted
	tmpDir := t.TempDir()
	validProject := filepath.Join(tmpDir, "my-safe-project")
	_ = os.MkdirAll(validProject, 0o755)
	if err := SafeRoot(validProject); err != nil {
		t.Errorf("expected valid project dir %q to pass SafeRoot, got: %v", validProject, err)
	}
}

func TestResolveWorkspace_Ladder(t *testing.T) {
	// Clean slate
	ResetLastResolvedWorkspace()

	// 1. Explicit path
	tmpDir := t.TempDir()
	explicitDir := filepath.Join(tmpDir, "explicit-proj")
	_ = os.MkdirAll(explicitDir, 0o755)

	res, err := ResolveWorkspace(explicitDir)
	if err != nil {
		t.Fatalf("explicit resolution failed: %v", err)
	}
	if filepath.Clean(res) != filepath.Clean(explicitDir) {
		t.Errorf("got %q, want %q", res, explicitDir)
	}

	// 2. MCP Roots protocol
	ResetLastResolvedWorkspace()
	mcpDir := filepath.Join(tmpDir, "mcp-proj")
	_ = os.MkdirAll(filepath.Join(mcpDir, "research"), 0o755)
	_ = os.WriteFile(filepath.Join(mcpDir, "research", ".vivechak.json"), []byte(`{"scope":"comparison"}`), 0o644)

	res, err = ResolveWorkspaceWithRoots(context.Background(), "", []string{mcpDir})
	if err != nil {
		t.Fatalf("MCP roots resolution failed: %v", err)
	}
	if filepath.Clean(res) != filepath.Clean(mcpDir) {
		t.Errorf("got %q, want %q", res, mcpDir)
	}

	// 3. Environment variable ladder
	ResetLastResolvedWorkspace()
	envDir := filepath.Join(tmpDir, "env-proj")
	_ = os.MkdirAll(envDir, 0o755)
	t.Setenv("VIVECHAK_PROJECT_ROOT", envDir)

	res, err = ResolveWorkspace("")
	if err != nil {
		t.Fatalf("env resolution failed: %v", err)
	}
	if filepath.Clean(res) != filepath.Clean(envDir) {
		t.Errorf("got %q, want %q", res, envDir)
	}

	// 4. Marker walk-up
	ResetLastResolvedWorkspace()
	t.Setenv("VIVECHAK_PROJECT_ROOT", "")
	t.Setenv("VCK_PROJECT_ROOT", "")
	t.Setenv("VIVECHAK_DEFAULT_ROOT", "")
	t.Setenv("VCK_DEFAULT_ROOT", "")

	walkDir := filepath.Join(tmpDir, "walk-proj")
	nested := filepath.Join(walkDir, "sub1", "sub2")
	_ = os.MkdirAll(filepath.Join(walkDir, "research"), 0o755)
	_ = os.WriteFile(filepath.Join(walkDir, "research", ".vivechak.json"), []byte(`{"scope":"project"}`), 0o644)
	_ = os.MkdirAll(nested, 0o755)

	origWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWd) }()
	_ = os.Chdir(nested)

	res, err = ResolveWorkspace("")
	if err != nil {
		t.Fatalf("walk-up resolution failed: %v", err)
	}
	if filepath.Clean(res) != filepath.Clean(walkDir) {
		t.Errorf("got %q, want %q", res, walkDir)
	}

	// 5. Guarded CWD (project manifest anchor)
	ResetLastResolvedWorkspace()
	anchorDir := filepath.Join(tmpDir, "anchor-proj")
	_ = os.MkdirAll(anchorDir, 0o755)
	_ = os.WriteFile(filepath.Join(anchorDir, "go.mod"), []byte("module test"), 0o644)
	_ = os.Chdir(anchorDir)

	res, err = ResolveWorkspace("")
	if err != nil {
		t.Fatalf("anchor resolution failed: %v", err)
	}
	if filepath.Clean(res) != filepath.Clean(anchorDir) {
		t.Errorf("got %q, want %q", res, anchorDir)
	}

	// 6. Fail-closed without anchors or session memory
	ResetLastResolvedWorkspace()
	plainDir := filepath.Join(tmpDir, "plain-empty-dir")
	_ = os.MkdirAll(plainDir, 0o755)
	_ = os.Chdir(plainDir)

	_, err = ResolveWorkspace("")
	if err == nil || !strings.Contains(err.Error(), "E-ROOT-UNRESOLVED") {
		t.Errorf("expected E-ROOT-UNRESOLVED error, got: %v", err)
	}
}
