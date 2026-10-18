package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func TestHandleVisualize_MermaidAndTable(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()

	// 1. Mermaid format (default)
	res, env, err := handleVisualize(ctx, nil, VisualizeInput{
		ProjectRoot: tmpDir,
	})
	if err != nil {
		t.Fatalf("handleVisualize failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got error: %s", env.Message)
	}

	dataMap, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", env.Data)
	}
	rendered, _ := dataMap["rendered"].(string)
	if !strings.Contains(rendered, "graph TD") {
		t.Errorf("expected graph TD in mermaid output, got:\n%s", rendered)
	}

	// 2. Table format
	resTable, envTable, err := handleVisualize(ctx, nil, VisualizeInput{
		ProjectRoot: tmpDir,
		Format:      "table",
	})
	if err != nil {
		t.Fatalf("handleVisualize table failed: %v", err)
	}
	if resTable.IsError {
		t.Fatalf("expected success, got error: %s", envTable.Message)
	}

	dataMapTable := envTable.Data.(map[string]any)
	renderedTable, _ := dataMapTable["rendered"].(string)
	if !strings.Contains(renderedTable, "| Session ID | Title |") {
		t.Errorf("expected table header in table output, got:\n%s", renderedTable)
	}

	// 3. Unsupported format
	resBad, envBad, _ := handleVisualize(ctx, nil, VisualizeInput{
		ProjectRoot: tmpDir,
		Format:      "invalid_format",
	})
	if !resBad.IsError {
		t.Errorf("expected error for unsupported format")
	}
	if envBad.Message == "" {
		t.Errorf("expected error message in envelope")
	}
}

func TestHandleVisualize_SluggedCompletedSessionDetection(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()

	// Write a session file with slug in name and NO id/session_id in frontmatter
	sessContent := `---
title: Landscape Research
status: complete
date: "2026-10-14"
---
# Landscape Research
Findings here.
`
	err := os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T0-01-landscape-analysis.md"), []byte(sessContent), 0o644)
	if err != nil {
		t.Fatalf("failed to write slugged session file: %v", err)
	}

	res, env, err := handleVisualize(ctx, nil, VisualizeInput{
		ProjectRoot: tmpDir,
		Format:      "table",
	})
	if err != nil || res.IsError {
		t.Fatalf("handleVisualize failed: err=%v, env=%+v", err, env)
	}
	dataMap := env.Data.(map[string]any)
	renderedTable := dataMap["rendered"].(string)

	if !strings.Contains(renderedTable, "Completed") {
		t.Errorf("expected T0-01 to be recognized as Completed via knownIDs slug matching, got:\n%s", renderedTable)
	}
}

func TestHandleVisualize_ComparisonScope(t *testing.T) {
	tmpDir := t.TempDir()
	rDir := filepath.Join(tmpDir, core.ResearchDir)
	_ = os.MkdirAll(rDir, 0o755)

	// Save comparison artifact
	cmp := `---
id: C-01-cmp-pino-vs-winston
title: Pino vs Winston Comparison
date: 2026-10-18
status: final
scope_level: comparison
---
# Pino vs Winston Comparison

## Recommendation
Pino is recommended.

## Weighted Evaluation Matrix
| Criterion | Weight | Pino | Winston |
|---|---|---|---|
| Latency | 0.6 | 5 | 3 |
| Ecosystem | 0.4 | 4 | 5 |
`
	_ = os.WriteFile(filepath.Join(rDir, "C-01-comparison.md"), []byte(cmp), 0o644)
	meta := `{"scope":"comparison"}`
	_ = os.WriteFile(filepath.Join(tmpDir, core.MetadataFile), []byte(meta), 0o644)

	ctx := context.Background()

	// 1. Mermaid format
	resM, envM, err := handleVisualize(ctx, nil, VisualizeInput{
		ProjectRoot: tmpDir,
		Format:      "mermaid",
	})
	if err != nil || resM.IsError {
		t.Fatalf("comparison visualize mermaid failed: err=%v, env=%+v", err, envM)
	}
	dataM := envM.Data.(map[string]any)
	renderedM := dataM["rendered"].(string)
	if !strings.Contains(renderedM, "graph TD") {
		t.Errorf("expected graph TD in comparison mermaid output, got: %s", renderedM)
	}

	// 2. Table format
	resT, envT, err := handleVisualize(ctx, nil, VisualizeInput{
		ProjectRoot: tmpDir,
		Format:      "table",
	})
	if err != nil || resT.IsError {
		t.Fatalf("comparison visualize table failed: err=%v, env=%+v", err, envT)
	}
	dataT := envT.Data.(map[string]any)
	renderedT := dataT["rendered"].(string)
	if !strings.Contains(renderedT, "| Criterion |") && !strings.Contains(renderedT, "| Option |") {
		t.Errorf("expected table output, got: %s", renderedT)
	}
}
