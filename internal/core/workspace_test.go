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
}
