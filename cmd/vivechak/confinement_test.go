package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestConfinement_LiveSubprocess verifies store confinement invariant across live stdio MCP interactions.
func TestConfinement_LiveSubprocess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live confinement subprocess test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)

	binPath := resolveOrBuildTestBinary(t)
	projDir := t.TempDir()

	// Create root README.md to ensure it is never touched or overwritten
	readmePath := filepath.Join(projDir, "README.md")
	readmeContent := []byte("# Safe Project README\n")
	if err := os.WriteFile(readmePath, readmeContent, 0644); err != nil {
		t.Fatal(err)
	}

	cs, cleanup := startLiveMCPServer(t, ctx, binPath, projDir)
	t.Cleanup(cleanup)

	callTool := func(name string, args map[string]any) testEnvelope {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			t.Fatalf("call to %s failed: %v", name, err)
		}
		return parseTestEnvelope(t, res)
	}

	// 1. Initialize decision scope
	callTool("vivechak_start", map[string]any{"project_root": projDir, "scope": "decision"})

	// 2. Traversal slug attempts via record_decision
	for _, slug := range []string{"../../../README", "x/../../AGENTS"} {
		env := callTool("vivechak_record_decision", map[string]any{
			"project_root": projDir,
			"decision_id":  "D-001",
			"slug":         slug,
			"content":      "---\nid: D-001\ntitle: Storage Engine\nstatus: proposed\ndoor_type: one-way\n---\n# D-001\n",
		})
		filePath, _ := env.Data["file_path"].(string)
		if env.Success && filePath != "" && !strings.HasPrefix(filePath, "research/") {
			t.Errorf("slug %q escaped research: got file_path %s", slug, filePath)
		}
	}

	// 3. Traversal plan ID attempts via save
	envPlan1 := callTool("vivechak_save", map[string]any{
		"project_root": projDir,
		"kind":         "plan",
		"scope":        "decision",
		"id":           "../AGENTS",
		"content":      "# Plan with traversal\n",
	})
	_ = envPlan1

	envPlan2 := callTool("vivechak_save", map[string]any{
		"project_root": projDir,
		"kind":         "plan",
		"scope":        "decision",
		"id":           "x/../../CLAUDE",
		"content":      "# Plan with traversal 2\n",
	})
	_ = envPlan2

	// 4. Verify no files written outside research/ and README.md intact
	readBack, err := os.ReadFile(readmePath)
	if err != nil || string(readBack) != string(readmeContent) {
		t.Fatalf("README.md was corrupted or overwritten!")
	}
	if _, err := os.Stat(filepath.Join(projDir, "AGENTS.md")); err == nil {
		t.Fatalf("AGENTS.md was created outside research/!")
	}

	// 5. Comparison, decision, and project journeys
	callTool("vivechak_save", map[string]any{
		"project_root": projDir,
		"kind":         "comparison",
		"id":           "C-001",
		"content":      "---\nid: C-001\ntitle: Comparison\ndate: 2026-10-25\nstatus: complete\n---\n## Key Findings\n- Sustains load [E-A: https://example.com/bench]\n## Recommendation\nUse A [E-A: https://example.com/docs]\n",
	})

	callTool("vivechak_record_decision", map[string]any{
		"project_root": projDir,
		"decision_id":  "D-001",
		"content":      "---\nid: D-001\ntitle: Primary Datastore\nstatus: accepted\ndoor_type: two-way\ndate: 2026-10-25\nreview_trigger: \"p95 > 50ms\"\n---\n# D-001: Primary Datastore\n## Decision Outcome\nPostgreSQL\n## Rejected Alternatives\nMySQL\n## Reversal Triggers\np95 > 50ms\n",
	})

	// Walk project directory and verify confinement
	err = filepath.Walk(projDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		rel, rErr := filepath.Rel(projDir, path)
		if rErr != nil {
			return rErr
		}
		relSlash := filepath.ToSlash(rel)
		if relSlash == "README.md" {
			return nil
		}
		if !strings.HasPrefix(relSlash, "research/") {
			t.Errorf("file created outside research/: %s", relSlash)
		}
		if strings.HasSuffix(relSlash, ".lock") && !strings.HasPrefix(relSlash, "research/.locks/") {
			t.Errorf("lock file created outside research/.locks/: %s", relSlash)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking project directory: %v", err)
	}
}
