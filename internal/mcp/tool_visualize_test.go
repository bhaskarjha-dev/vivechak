package mcputil

import (
	"context"
	"strings"
	"testing"
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
