package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func setupTestPipelineWorkspace(t *testing.T) string {
	tmpDir := t.TempDir()
	researchDir := filepath.Join(tmpDir, core.ResearchDir)
	sessionsDir := filepath.Join(tmpDir, core.SessionsDir)
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	pipelineContent := `# Research Pipeline

## Pipeline Topology

| ID | Topic | Dependencies | Output File |
|---|---|---|---|
| T0-01 | Landscape | none | sessions/T0-01.md |
| T1-01 | Database | T0-01 | sessions/T1-01.md |

---

## Session Prompts

### Session T0-01: Landscape

| **Field** | **Value** |
|---|---|
| **ID** | T0-01 |
| **Layer** | 0 |
| **Dependencies** | none |
| **Output** | sessions/T0-01.md |

` + "```prompt\n# T0-01\nBRIEF:\nLandscape research.\n```\n" + `

---

### Session T1-01: Database

| **Field** | **Value** |
|---|---|
| **ID** | T1-01 |
| **Layer** | 1 |
| **Dependencies** | T0-01 |
| **Output** | sessions/T1-01.md |

` + "```prompt\n# T1-01\nBRIEF:\nDatabase research.\n```\n"

	if err := os.WriteFile(filepath.Join(researchDir, "RESEARCH-PIPELINE.md"), []byte(pipelineContent), 0o644); err != nil {
		t.Fatalf("failed to write pipeline: %v", err)
	}

	return tmpDir
}

func TestHandleReplan_AddSession(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()

	res, env, err := handleReplan(ctx, nil, ReplanInput{
		ProjectRoot:  tmpDir,
		Operation:    "add_session",
		SessionID:    "T1-02",
		Topic:        "Caching Layer",
		Layer:        "1",
		Dependencies: "T0-01",
		Prompt:       "# T1-02: Caching\nBRIEF:\nResearch Redis vs Memcached.\n",
	})
	if err != nil {
		t.Fatalf("handleReplan failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %s", env.Message)
	}

	// Verify pipeline file was updated
	data, err := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	if err != nil {
		t.Fatalf("failed to read pipeline file: %v", err)
	}
	if !strings.Contains(string(data), "T1-02") {
		t.Errorf("expected pipeline file to contain T1-02")
	}
}

func TestHandleReplan_RemoveSession(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()

	// 1. Remove uncompleted session T1-01
	res, env, err := handleReplan(ctx, nil, ReplanInput{
		ProjectRoot: tmpDir,
		Operation:   "remove_session",
		SessionID:   "T1-01",
	})
	if err != nil {
		t.Fatalf("handleReplan failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %s", env.Message)
	}

	data, _ := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	if strings.Contains(string(data), "Session T1-01:") {
		t.Errorf("expected T1-01 prompt to be removed from pipeline file")
	}

	// 2. Mark T0-01 as completed by creating its session file
	sessionContent := "---\nid: T0-01\ntitle: Done\n---\n# Done\n"
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T0-01.md"), []byte(sessionContent), 0o644)

	// Attempt to remove completed session T0-01 -> must error
	resComp, _, errComp := handleReplan(ctx, nil, ReplanInput{
		ProjectRoot: tmpDir,
		Operation:   "remove_session",
		SessionID:   "T0-01",
	})
	if errComp != nil {
		t.Fatalf("unexpected handler err: %v", errComp)
	}
	if !resComp.IsError {
		t.Errorf("expected error removing completed session T0-01")
	}
}

func TestHandleReplan_UpdateDepsAndPrompt(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()

	// Update prompt
	res, env, err := handleReplan(ctx, nil, ReplanInput{
		ProjectRoot: tmpDir,
		Operation:   "update_prompt",
		SessionID:   "T1-01",
		NewPrompt:   "# T1-01 Updated\nBRIEF:\nUpdated brief.\n",
	})
	if err != nil {
		t.Fatalf("handleReplan failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %s", env.Message)
	}

	data, _ := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	if !strings.Contains(string(data), "Updated brief") {
		t.Errorf("expected updated prompt in pipeline")
	}
}
