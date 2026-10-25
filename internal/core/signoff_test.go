package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

func TestSignoffDecision_UpdatesBothTableAndDetailsBlock(t *testing.T) {
	tmpDir := t.TempDir()
	rDir := filepath.Join(tmpDir, ResearchDir)
	_ = os.MkdirAll(rDir, 0o755)

	// Standalone ADR D-001
	adrContent := `---
id: D-001
title: Datastore Architecture
status: accepted
door_type: one-way
human_reviewed: false
---
# Context
Storage choice.
`
	if err := os.WriteFile(filepath.Join(rDir, "D-001.md"), []byte(adrContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Consolidated registry with BOTH a table row and an anchored details block
	registryContent := `# Architectural Decisions Registry

| ID | Title | Status | Door | Human Review | File |
|---|---|---|---|---|---|
| D-001 | Datastore Architecture | accepted | one-way | false | D-001.md |

<!-- DECISION: D-001 -->
<details open>
<summary><b>D-001: Datastore Architecture</b></summary>

` + "```yaml" + `
id: D-001
title: Datastore Architecture
status: accepted
door_type: one-way
human_reviewed: false
` + "```" + `

### Context
Storage choice.
</details>
<!-- /DECISION: D-001 -->
`
	if err := os.WriteFile(filepath.Join(tmpDir, DecisionsFile), []byte(registryContent), 0o644); err != nil {
		t.Fatal(err)
	}

	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("OpenWorkspace failed: %v", err)
	}
	defer ws.Close()

	relPath, err := SignoffDecision(ws, "D-001", "Lead Architect")
	if err != nil {
		t.Fatalf("SignoffDecision failed: %v", err)
	}
	if filepath.ToSlash(relPath) != "research/D-001.md" {
		t.Errorf("expected relPath research/D-001.md, got %q", relPath)
	}

	// Verify D-001.md was updated
	adrUpdated, _ := os.ReadFile(filepath.Join(rDir, "D-001.md"))
	adrStr := string(adrUpdated)
	if !strings.Contains(adrStr, "human_reviewed: true") {
		t.Errorf("expected D-001.md to contain human_reviewed: true, got:\n%s", adrStr)
	}
	if !strings.Contains(adrStr, "reviewed_by: Lead Architect") {
		t.Errorf("expected D-001.md to contain reviewed_by: Lead Architect, got:\n%s", adrStr)
	}

	// Verify DECISIONS.md was updated in BOTH table and details block
	regUpdated, _ := os.ReadFile(filepath.Join(tmpDir, DecisionsFile))
	regStr := string(regUpdated)
	if !strings.Contains(regStr, "| D-001 | Datastore Architecture | accepted | one-way | true | D-001.md |") {
		t.Errorf("expected table row in DECISIONS.md to be updated with true, got:\n%s", regStr)
	}
	if !strings.Contains(regStr, "human_reviewed: true") {
		t.Errorf("expected <details> block in DECISIONS.md to contain human_reviewed: true, got:\n%s", regStr)
	}
	if !strings.Contains(regStr, "reviewed_by: Lead Architect") {
		t.Errorf("expected <details> block in DECISIONS.md to contain reviewed_by: Lead Architect, got:\n%s", regStr)
	}
}

func TestSignoffDecision_PreservesSpecialCharactersAndDollarSigns(t *testing.T) {
	tmpDir := t.TempDir()
	rDir := filepath.Join(tmpDir, ResearchDir)
	_ = os.MkdirAll(rDir, 0o755)

	adrContent := `---
id: D-002
title: Cloud Cost Model
status: accepted
door_type: one-way
human_reviewed: false
---
# Context
Estimated cost is $500/month with variable $RATE scaling.
`
	if err := os.WriteFile(filepath.Join(rDir, "D-002.md"), []byte(adrContent), 0o644); err != nil {
		t.Fatal(err)
	}

	registryContent := `# Architectural Decisions Registry

<!-- DECISION: D-002 -->
<details open>
<summary><b>D-002: Cloud Cost Model</b></summary>

` + "```yaml" + `
id: D-002
title: Cloud Cost Model
status: accepted
door_type: one-way
human_reviewed: false
` + "```" + `

### Context
Estimated cost is $500/month with variable $RATE scaling.
</details>
<!-- /DECISION: D-002 -->
`
	if err := os.WriteFile(filepath.Join(tmpDir, DecisionsFile), []byte(registryContent), 0o644); err != nil {
		t.Fatal(err)
	}

	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("OpenWorkspace failed: %v", err)
	}
	defer ws.Close()

	_, err = SignoffDecision(ws, "D-002", "Finance Reviewer ($USD)")
	if err != nil {
		t.Fatalf("SignoffDecision failed: %v", err)
	}

	regUpdated, _ := os.ReadFile(filepath.Join(tmpDir, DecisionsFile))
	regStr := string(regUpdated)
	if !strings.Contains(regStr, "$500/month with variable $RATE scaling.") {
		t.Errorf("expected dollar signs and body to be preserved verbatim, got:\n%s", regStr)
	}
	if !strings.Contains(regStr, "Finance Reviewer ($USD)") {
		t.Errorf("expected reviewer name with dollar signs to be preserved verbatim, got:\n%s", regStr)
	}
}

func TestSignoffDecision_CaseInsensitiveDecisionID(t *testing.T) {
	tmpDir := t.TempDir()
	rDir := filepath.Join(tmpDir, ResearchDir)
	_ = os.MkdirAll(rDir, 0o755)

	adrContent := `---
id: D-003
title: API Protocol
status: accepted
door_type: one-way
human_reviewed: false
---
# Context
API protocol.
`
	if err := os.WriteFile(filepath.Join(rDir, "D-003.md"), []byte(adrContent), 0o644); err != nil {
		t.Fatal(err)
	}

	registryContent := `# Architectural Decisions Registry

| ID | Title | Status | Door | Human Review | File |
|---|---|---|---|---|---|
| D-003 | API Protocol | accepted | one-way | false | D-003.md |

<!-- DECISION: D-003 -->
<details open>
<summary><b>D-003: API Protocol</b></summary>

` + "```yaml" + `
id: D-003
title: API Protocol
status: accepted
door_type: one-way
human_reviewed: false
` + "```" + `

### Context
API protocol.
</details>
<!-- /DECISION: D-003 -->
`
	if err := os.WriteFile(filepath.Join(tmpDir, DecisionsFile), []byte(registryContent), 0o644); err != nil {
		t.Fatal(err)
	}

	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("OpenWorkspace failed: %v", err)
	}
	defer ws.Close()

	// Call with LOWERCASE "d-003"
	_, err = SignoffDecision(ws, "d-003", "Protocol Lead")
	if err != nil {
		t.Fatalf("SignoffDecision failed on lowercase ID: %v", err)
	}

	// Verify DECISIONS.md was updated in BOTH table and details block
	regUpdated, _ := os.ReadFile(filepath.Join(tmpDir, DecisionsFile))
	regStr := string(regUpdated)
	if !strings.Contains(regStr, "| D-003 | API Protocol | accepted | one-way | true | D-003.md |") {
		t.Errorf("expected table row to be updated when signed off with lowercase ID, got:\n%s", regStr)
	}
	if !strings.Contains(regStr, "human_reviewed: true") {
		t.Errorf("expected details block to be updated when signed off with lowercase ID, got:\n%s", regStr)
	}
	if !strings.Contains(regStr, "reviewed_by: Protocol Lead") {
		t.Errorf("expected details block to contain reviewed_by, got:\n%s", regStr)
	}
}


