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
