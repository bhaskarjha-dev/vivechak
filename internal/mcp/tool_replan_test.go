package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

// replanTopologyTable is planner prose in the test pipeline; replan leaves it as written (HIGH-07).
const replanTopologyTable = `| ID | Topic | Dependencies | Output File |
|---|---|---|---|
| T0-01 | Landscape | none | sessions/T0-01.md |
| T1-01 | Database | T0-01 | sessions/T1-01.md |
`

func setupTestPipelineWorkspace(t *testing.T) string {
	tmpDir := t.TempDir()
	researchDir := filepath.Join(tmpDir, core.ResearchDir)
	sessionsDir := filepath.Join(tmpDir, core.SessionsDir)
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("failed to create dirs: %v", err)
	}

	pipelineContent := `# Research Pipeline

## Pipeline Topology

` + replanTopologyTable + `
` + "```yaml vivechak-plan" + `
plan: PROJECT
sessions:
  - {id: T0-01, title: Landscape, layer: 0, depends: []}
  - {id: T1-01, title: Database, layer: 1, depends: [T0-01]}
  - {id: SYN-01, title: Synthesis, layer: 2, kind: synthesis, depends: [ALL]}
` + "```" + `

---

## Sessions

### T0-01: Landscape

| **Field** | **Value** |
|---|---|
| **ID** | T0-01 |
| **Layer** | 0 |
| **Dependencies** | none |
| **Output** | sessions/T0-01.md |

` + "```prompt\n# T0-01\nBRIEF:\nLandscape research.\n```\n" + `

---

### T1-01: Database

| **Field** | **Value** |
|---|---|
| **ID** | T1-01 |
| **Layer** | 1 |
| **Dependencies** | T0-01 |
| **Output** | sessions/T1-01.md |

` + "```prompt\n# T1-01\nBRIEF:\nDatabase research.\n```\n" + `

### SYN-01: Synthesis

` + "```prompt\n# SYN-01\nBRIEF:\nSynthesize. [ALL_SESSION_FINDINGS]\n```\n"

	if err := os.WriteFile(filepath.Join(researchDir, "RESEARCH-PIPELINE.md"), []byte(pipelineContent), 0o644); err != nil {
		t.Fatalf("failed to write pipeline: %v", err)
	}

	return tmpDir
}

func TestHandleReplan_AddSession(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()

	res, env, err := handleReplan(ctx, nil, ReplanInput{
		Plan:         "PROJECT",
		ProjectRoot:  tmpDir,
		Op:           "add_session",
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

func TestHandleReplan_AddSession_AllPriorSessions(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()

	res, env, err := handleReplan(ctx, nil, ReplanInput{
		Plan:         "PROJECT",
		ProjectRoot:  tmpDir,
		Op:           "add_session",
		SessionID:    "T2-01",
		Topic:        "Integration",
		Layer:        "2",
		Dependencies: "ALL prior sessions (hard)",
		Prompt:       "# T2-01: Integration\nBRIEF:\nIntegrate the architecture.\n",
	})
	if err != nil {
		t.Fatalf("handleReplan failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %s", env.Message)
	}

	// Read updated pipeline and parse DAG
	data, err := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	if err != nil {
		t.Fatalf("failed to read pipeline file: %v", err)
	}
	dag, err := core.ParsePipeline(data)
	if err != nil {
		t.Fatalf("failed to parse updated pipeline: %v", err)
	}
	added := dag.SessionByID("T2-01")
	if added == nil {
		t.Fatal("T2-01 not found in parsed pipeline")
	}
	// setupTestPipelineWorkspace has T0-01 and T1-01 before its synthesis session SYN-01; the new session goes
	// before SYN-01, so ALL is those two and SYN-01 (depends: [ALL]) now waits for T2-01 too.
	if len(added.Dependencies) != 2 || added.Dependencies[0] != "T0-01" || added.Dependencies[1] != "T1-01" {
		t.Fatalf("expected dependencies [T0-01, T1-01], got %v", added.Dependencies)
	}
	if got := dag.SessionByID("SYN-01").Dependencies; strings.Join(got, ",") != "T0-01,T1-01,T2-01" {
		t.Fatalf("expected SYN-01 to depend on [T0-01 T1-01 T2-01], got %v", got)
	}
}

func TestHandleReplan_RemoveSession(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()

	// 1. Remove uncompleted session T1-01
	res, env, err := handleReplan(ctx, nil, ReplanInput{
		Plan:        "PROJECT",
		ProjectRoot: tmpDir,
		Op:          "remove_session",
		SessionID:   "T1-01",
	})
	if err != nil {
		t.Fatalf("handleReplan failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got: %s", env.Message)
	}

	data, _ := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	if strings.Contains(string(data), "### T1-01:") || strings.Contains(string(data), "{id: T1-01,") {
		t.Errorf("expected T1-01's prompt section and block entry to be removed from pipeline file")
	}

	// 2. Mark T0-01 as completed by creating its session file
	sessionContent := "---\nid: T0-01\ntitle: Done\n---\n# Done\n"
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T0-01.md"), []byte(sessionContent), 0o644)

	// Attempt to remove completed session T0-01 -> must error
	resComp, _, errComp := handleReplan(ctx, nil, ReplanInput{
		Plan:        "PROJECT",
		ProjectRoot: tmpDir,
		Op:          "remove_session",
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
		Plan:        "PROJECT",
		ProjectRoot: tmpDir,
		Op:          "update_prompt",
		SessionID:   "T1-01",
		Prompt:      "# T1-01 Updated\nBRIEF:\nUpdated brief.\n",
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

func TestHandleReplan_RemoveSession_SluggedWithoutFrontmatterID(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()

	// Write session as slugged file T0-01-landscape.md without frontmatter id
	sessionContent := "---\ntitle: Done\nstatus: complete\ndate: \"2026-10-14\"\n---\n# Done\n"
	_ = os.WriteFile(filepath.Join(tmpDir, core.SessionsDir, "T0-01-landscape.md"), []byte(sessionContent), 0o644)

	// Attempt to remove T0-01 -> should be recognized as completed via knownIDs and rejected
	resComp, envComp, errComp := handleReplan(ctx, nil, ReplanInput{
		Plan:        "PROJECT",
		ProjectRoot: tmpDir,
		Op:          "remove_session",
		SessionID:   "T0-01",
	})
	if errComp != nil {
		t.Fatalf("unexpected handler err: %v", errComp)
	}
	if !resComp.IsError {
		t.Errorf("expected error removing completed session T0-01 via slugged file, got env: %+v", envComp)
	}
	if !strings.Contains(envComp.Message, "cannot remove completed session") {
		t.Errorf("expected 'cannot remove completed session' message, got: %s", envComp.Message)
	}
}

func TestHandleReplan_MissingOperationRefused(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()

	res, env, err := handleReplan(ctx, nil, ReplanInput{
		Plan:        "PROJECT",
		ProjectRoot: tmpDir,
		Reason:      "Company directive: no MongoDB allowed",
	})
	if err != nil {
		t.Fatalf("unexpected handler err: %v", err)
	}
	if !res.IsError || env.Ok || env.Code != "E-INVALID-ARGUMENT" {
		t.Fatalf("expected E-INVALID-ARGUMENT for missing op, got: %+v", env)
	}
}

func TestHandleReplan_WithReason(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()

	res, env, err := handleReplan(ctx, nil, ReplanInput{
		Plan:        "PROJECT",
		ProjectRoot: tmpDir,
		Op:          "update_prompt",
		SessionID:   "T1-01",
		Reason:      "Strict enterprise compliance requirements",
		Prompt:      "# T1-01 Oracle Evaluation\nBRIEF:\nResearch Oracle.\n",
	})
	if err != nil {
		t.Fatalf("unexpected handler err: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got error: %s", env.Message)
	}
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map data in envelope")
	}
	if data["reason"] != "Strict enterprise compliance requirements" {
		t.Errorf("expected reason in resData, got %v", data["reason"])
	}
}

// Replan edits the session's block entry and its one prompt section in place; the planner's topology table is
// prose and stays exactly as written (HIGH-07).
func TestHandleReplan_EditsBlockAndPromptSectionInPlace(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	ctx := context.Background()
	original, _ := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))

	// 1. Update deps for T1-01 to "none"
	res, env, err := handleReplan(ctx, nil, ReplanInput{
		Plan:         "PROJECT",
		ProjectRoot:  tmpDir,
		Op:           "update_deps",
		SessionID:    "T1-01",
		Dependencies: "none",
	})
	if err != nil || res.IsError {
		t.Fatalf("update_deps failed: %v (%s)", err, env.Message)
	}

	data, _ := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	pStr := string(data)
	want := strings.Replace(string(original), "{id: T1-01, title: Database, layer: 1, depends: [T0-01]}", "{id: T1-01, title: Database, layer: 1, depends: []}", 1)
	if pStr != want {
		t.Errorf("update_deps must change only T1-01's depends value, got:\n%s", pStr)
	}

	// 2. Add session T2-01
	resAdd, envAdd, errAdd := handleReplan(ctx, nil, ReplanInput{
		Plan:         "PROJECT",
		ProjectRoot:  tmpDir,
		Op:           "add_session",
		SessionID:    "T2-01",
		Topic:        "Observability Tier",
		Layer:        "2",
		Dependencies: "T1-01",
	})
	if errAdd != nil || resAdd.IsError {
		t.Fatalf("add_session failed: %v (%s)", errAdd, envAdd.Message)
	}

	dataAdd, _ := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	pAddStr := string(dataAdd)
	if !strings.Contains(pAddStr, "  - {id: T2-01, title: Observability Tier, layer: 2, door: two-way, depends: [T1-01]}\n  - {id: SYN-01") {
		t.Errorf("expected a T2-01 block entry before SYN-01, got:\n%s", pAddStr)
	}
	if !strings.Contains(pAddStr, "### T2-01: Observability Tier\n\n```prompt\n") || !strings.Contains(pAddStr, replanTopologyTable) {
		t.Errorf("expected a T2-01 prompt section and the topology table as written, got:\n%s", pAddStr)
	}

	// 3. Remove session T2-01
	resDel, envDel, errDel := handleReplan(ctx, nil, ReplanInput{
		Plan:        "PROJECT",
		ProjectRoot: tmpDir,
		Op:          "remove_session",
		SessionID:   "T2-01",
	})
	if errDel != nil || resDel.IsError {
		t.Fatalf("remove_session failed: %v (%s)", errDel, envDel.Message)
	}

	dataDel, _ := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	if string(dataDel) != want {
		t.Errorf("removing T2-01 must restore the plan as it was before it was added, got:\n%s", string(dataDel))
	}
}

// Removing the synthesis session of a project plan would leave it without one: refused with V-PLAN-SYNTHESIS.
func TestHandleReplan_RemoveSynthesisSessionRefused(t *testing.T) {
	tmpDir := setupTestPipelineWorkspace(t)
	before, _ := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	res, env, err := handleReplan(context.Background(), nil, ReplanInput{Plan: "PROJECT", ProjectRoot: tmpDir, Op: "remove_session", SessionID: "SYN-01"})
	if err != nil {
		t.Fatalf("handleReplan: %v", err)
	}
	if !res.IsError || !strings.Contains(env.Message, core.CodePlanSynthesis) || envGuidance(env) == "" {
		t.Fatalf("expected %s with a fix, got: %+v", core.CodePlanSynthesis, env)
	}
	after, _ := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	if string(after) != string(before) {
		t.Fatal("a refused replan must leave the plan unchanged")
	}
}
