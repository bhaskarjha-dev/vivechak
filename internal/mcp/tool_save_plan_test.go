package mcputil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func TestSavePlan_PipelineContentAlias(t *testing.T) {
	tmpDir := t.TempDir()

	pipelineContent := `# Research Pipeline
### Session T1-01: Storage Selection
| **ID** | T1-01 |
| **Dependencies** | None |
` + "```prompt\nBrief\n```\n"

	decisionsContent := `# Decisions
<!-- DECISION: D-001 -->
<details>
<summary><strong>D-001</strong> · Storage Selection · <code>one-way</code> · proposed</summary>

` + "```yaml\nid: D-001\ntitle: Storage Selection\ndoor_type: one-way\nstatus: proposed\n```\n" + `
</details>
<!-- /DECISION: D-001 -->
`

	ctx := context.Background()

	// Pass pipeline content via PipelineContent instead of Content
	_, env, err := handleSavePlan(ctx, nil, SavePlanInput{
		ProjectRoot:      tmpDir,
		Scope:            "project",
		PipelineContent:  pipelineContent,
		DecisionsContent: decisionsContent,
	})
	if err != nil || !env.Success {
		t.Fatalf("handleSavePlan with PipelineContent failed: err=%v, env=%+v", err, env)
	}

	// Verify RESEARCH-PIPELINE.md was persisted
	pData, err := os.ReadFile(filepath.Join(tmpDir, core.PipelineFile))
	if err != nil {
		t.Fatalf("reading pipeline file: %v", err)
	}
	if !strings.Contains(string(pData), "T1-01") {
		t.Errorf("expected pipeline file to contain T1-01, got:\n%s", string(pData))
	}

	// Verify DECISIONS.md was persisted
	dData, err := os.ReadFile(filepath.Join(tmpDir, core.DecisionsFile))
	if err != nil {
		t.Fatalf("reading decisions file: %v", err)
	}
	if !strings.Contains(string(dData), "D-001") {
		t.Errorf("expected decisions file to contain D-001, got:\n%s", string(dData))
	}
}
