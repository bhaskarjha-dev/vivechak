package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScopeFromString_And_GeneratorFile(t *testing.T) {
	// Test ScopeFromString
	cases := []struct {
		input    string
		expected Scope
		valid    bool
	}{
		{"project", ScopeProject, true},
		{"decision", ScopeDecision, true},
		{"comparison", ScopeComparison, true},
		{"invalid", "invalid", false},
		{"", "", false},
	}

	for _, tc := range cases {
		s, ok := ScopeFromString(tc.input)
		if ok != tc.valid {
			t.Errorf("ScopeFromString(%q): expected valid=%v, got %v", tc.input, tc.valid, ok)
		}
		if s != tc.expected {
			t.Errorf("ScopeFromString(%q): expected scope=%q, got %q", tc.input, tc.expected, s)
		}
	}

	// Test GeneratorFile
	if gen := GeneratorFile(ScopeDecision); gen != "GENERATOR-DECISION.md" {
		t.Errorf("expected GENERATOR-DECISION.md, got %q", gen)
	}
	if gen := GeneratorFile(ScopeComparison); gen != "GENERATOR-COMPARISON.md" {
		t.Errorf("expected GENERATOR-COMPARISON.md, got %q", gen)
	}
	if gen := GeneratorFile(ScopeProject); gen != "GENERATOR.md" {
		t.Errorf("expected GENERATOR.md, got %q", gen)
	}
	if gen := GeneratorFile("unknown"); gen != "" {
		t.Errorf("expected empty string for unknown generator, got %q", gen)
	}
}

func TestValidation_HelpersAndArtifact(t *testing.T) {
	// 1. ValidationLevel.String()
	levels := []struct {
		lvl  ValidationLevel
		name string
	}{
		{L1Construct, "L1-CONSTRUCT"},
		{L2Block, "L2-BLOCK"},
		{L3Warn, "L3-WARN"},
		{L4Gate, "L4-GATE"},
		{ValidationLevel(99), "L99"},
	}
	for _, l := range levels {
		if s := l.lvl.String(); s != l.name {
			t.Errorf("expected %s, got %s", l.name, s)
		}
	}

	// 2. ValidationIssue.String()
	issue := ValidationIssue{
		Level:   L3Warn,
		Code:    "W-TEST",
		Message: "Test warning message",
		FixHint: "Test hint",
	}
	issueStr := issue.String()
	if !strings.Contains(issueStr, "[L3-WARN]") || !strings.Contains(issueStr, "W-TEST") || !strings.Contains(issueStr, "fix: Test hint") {
		t.Errorf("unexpected issue string: %s", issueStr)
	}

	issueNoHint := ValidationIssue{
		Level:   L2Block,
		Code:    "E-BLOCK",
		Message: "Blocking error",
	}
	if strings.Contains(issueNoHint.String(), "fix:") {
		t.Errorf("expected no fix hint in string, got %s", issueNoHint.String())
	}

	// 3. ValidationResult methods
	var res ValidationResult
	res.AddIssue(L2Block, "E-ERR1", "Error 1")
	res.AddIssue(L3Warn, "W-WARN1", "Warning 1")
	if res.ErrorCount() != 1 {
		t.Errorf("expected 1 error, got %d", res.ErrorCount())
	}
	if res.WarningCount() != 1 {
		t.Errorf("expected 1 warning, got %d", res.WarningCount())
	}

	// 4. ValidateArtifact
	emptyArt := ValidateArtifact([]byte(""))
	if emptyArt.ErrorCount() == 0 {
		t.Errorf("expected blocking issue on empty artifact")
	}

	validArt := ValidateArtifact([]byte("---\ntitle: Artifact\n---\n# Architecture\nPostgreSQL is selected for database storage. Grade A (verified via official documentation). It supports relational integrity and ACID transactions at high scale.\n"))
	if validArt.ErrorCount() > 0 {
		t.Errorf("expected no blocking errors on valid artifact, got issues: %v", validArt.Issues)
	}
	if validArt.WarningCount() > 0 {
		t.Errorf("expected no warnings on valid artifact, got issues: %v", validArt.Issues)
	}
}

func TestResolveWorkspace_And_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	if canonical, err := filepath.EvalSymlinks(tmpDir); err == nil {
		tmpDir = canonical
	}

	// WorkspaceExists on nonexistent
	if WorkspaceExists(tmpDir) {
		t.Errorf("expected WorkspaceExists to be false before research dir created")
	}

	// Create research dir
	resDir := filepath.Join(tmpDir, ResearchDir)
	if err := os.MkdirAll(resDir, 0755); err != nil {
		t.Fatal(err)
	}

	if !WorkspaceExists(tmpDir) {
		t.Errorf("expected WorkspaceExists to be true after research dir created")
	}

	// 1. Explicit path
	res, err := ResolveWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("ResolveWorkspace explicit: %v", err)
	}
	if res != tmpDir {
		t.Errorf("expected %s, got %s", tmpDir, res)
	}

	// 2. VIVECHAK_PROJECT_ROOT env var
	t.Setenv("VIVECHAK_PROJECT_ROOT", tmpDir)
	resEnv, err := ResolveWorkspace("")
	if err != nil {
		t.Fatalf("ResolveWorkspace env var: %v", err)
	}
	if resEnv != tmpDir {
		t.Errorf("expected %s from env var, got %s", tmpDir, resEnv)
	}
	t.Setenv("VIVECHAK_PROJECT_ROOT", "")

	// 3. Walk up from CWD
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	subDir := filepath.Join(tmpDir, "sub", "deep")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(origCwd)
	}()

	resWalk, err := ResolveWorkspace("")
	if err != nil {
		t.Fatalf("ResolveWorkspace walk up: %v", err)
	}
	realWalk, err := filepath.EvalSymlinks(resWalk)
	if err != nil {
		realWalk = resWalk
	}
	if realWalk != tmpDir {
		t.Errorf("expected %s from walk up, got %s", tmpDir, resWalk)
	}

	// 4. Failure when not found
	emptyDir := t.TempDir()
	if err := os.Chdir(emptyDir); err != nil {
		t.Fatal(err)
	}
	_, err = ResolveWorkspace("")
	_ = os.Chdir(origCwd) // Restore immediately so emptyDir is not locked on Windows cleanup
	if err == nil {
		t.Errorf("expected error when no workspace found walking up from empty dir")
	}
}
