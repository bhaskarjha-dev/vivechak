package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMassive_FilesystemAndChaosMatrix tests filesystem edge cases, chaos file injection,
// special non-ADR file whitelisting, malformed YAML frontmatter, path traversal defense,
// and file isolation across the entire workspace.
func TestMassive_FilesystemAndChaosMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	wsDir := t.TempDir()

	// 1. Create directory structure
	_ = os.MkdirAll(filepath.Join(wsDir, core.TemplatesDir), 0o755)
	_ = os.MkdirAll(filepath.Join(wsDir, core.SessionsDir), 0o755)

	// 2. Populate all 6 templates
	for _, tmpl := range core.TemplatesToCopy {
		p := filepath.Join(wsDir, core.TemplatesDir, tmpl)
		if err := os.WriteFile(p, []byte("# Template "+tmpl), 0o644); err != nil {
			t.Fatalf("writing template %s: %v", tmpl, err)
		}
	}

	// 3. Inject chaos files into research/
	// These non-ADR markdown and text files MUST be safely ignored by IsSpecialResearchFile
	// and never break decision parsing or workspace validation.
	chaosFiles := []string{
		"NOTES.md",
		"TODO.md",
		"scratch.txt",
		"D-001-conflict-resolution.md",
		"D-002-plan.md",
		"FOUNDING-ARCHITECTURE.md",
		"PHASE-0-GATE.md",
		"DECISION-NOTES.md",
		"extra.json",
	}
	for _, f := range chaosFiles {
		p := filepath.Join(wsDir, core.ResearchDir, f)
		content := "# Chaos File " + f + "\nThis is random research scratch.\n"
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write chaos file %s: %v", f, err)
		}
	}

	// 4. Verify core.IsSpecialResearchFile recognizes special research artifacts vs standard ADRs
	specialFiles := []string{
		"D-001-conflict-resolution.md",
		"D-002-plan.md",
		"COMP-01-comparison.md",
		"FOUNDING-ARCHITECTURE.md",
		"PHASE-0-GATE.md",
		"DECISIONS.md",
		"FAD.md",
		"RESEARCH-PIPELINE.md",
	}
	for _, f := range specialFiles {
		if !core.IsSpecialResearchFile(f) {
			t.Errorf("expected IsSpecialResearchFile(%q) to be true", f)
		}
	}

	nonSpecialFiles := []string{
		"D-001.md",
		"D-002.md",
		"D-ARCH-01.md",
		"NOTES.md",
		"TODO.md",
		"scratch.txt",
		"extra.json",
	}
	for _, f := range nonSpecialFiles {
		if core.IsSpecialResearchFile(f) {
			t.Errorf("expected IsSpecialResearchFile(%q) to be false", f)
		}
	}

	// 5. Test session content validation with core.ValidateSession
	validSession := `---
session_id: T1-01
title: Foundation
status: complete
date: 2026-09-28
---
# Foundation
Refers to D-001. A (doc)
`
	resValid := core.ValidateSession([]byte(validSession))
	if resValid.HasBlocking() {
		t.Errorf("expected valid session to have no blocking issues: %v", resValid.Issues)
	}

	// 6. Test malformed YAML frontmatter resilience
	malformedSessions := map[string]string{
		"unclosed.md":    "---\nsession_id: S-01\ntitle: Unclosed\n# No closing dashes\n",
		"bad_yaml.md":    "---\nsession_id: [unclosed list\n---\n# Body\n",
		"empty_block.md": "---\n---\n# Empty Frontmatter\n",
	}
	for name, content := range malformedSessions {
		res := core.ValidateSession([]byte(content))
		if len(res.Issues) == 0 {
			t.Errorf("expected malformed session %s to have validation issues", name)
		}
	}

	// 7. Verify doctor ignores the chaos files and doesn't treat them as orphaned ADRs
	_ = os.WriteFile(filepath.Join(wsDir, core.PipelineFile), []byte("# Research Pipeline\n#### T1-01: Foundation\n"), 0o644)
	_ = os.WriteFile(filepath.Join(wsDir, core.DecisionsFile), []byte("# Decisions\n<!-- DECISION: D-001 -->\n"), 0o644)
	_ = os.WriteFile(filepath.Join(wsDir, core.SessionsDir, "T1-01.md"), []byte(validSession), 0o644)

	hasErrors, warnings, errs := checkWorkspace(wsDir)
	if hasErrors {
		t.Fatalf("expected workspace with chaos files to be healthy, but got errors: %v (warnings: %v)", errs, warnings)
	}
}

// TestMassive_LiveMCP_DiamondDAG_And_DownstreamCascades exercises a complex Diamond DAG
// with parallel session discovery (parallelism_hint, ⚡ indicator), cascading downstream
// stale detection (W-STALE-DOWNSTREAM), single-source FAD persistence, root copy matching,
// and auto-persisted Phase 0 exit gate verification over real stdio MCP.
func TestMassive_LiveMCP_DiamondDAG_And_DownstreamCascades(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)

	// 1. Build live binary
	binDir := t.TempDir()
	binName := "vck_diamond_live"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\noutput: %s", err, string(out))
	}

	// 2. Connect to live MCP server
	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "diamond-dag-client", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to live MCP server: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	callAndParse := func(toolName string, args map[string]any) testEnvelope {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		})
		if err != nil {
			t.Fatalf("call %s failed: %v", toolName, err)
		}
		return parseTestEnvelope(t, res)
	}

	projectDir := t.TempDir()

	// 3. vivechak_init
	envInit := callAndParse("vivechak_init", map[string]any{
		"project_root": projectDir,
		"scope":        "project",
	})
	if !envInit.Success {
		t.Fatalf("init failed: %s", envInit.Message)
	}
	tmpls, ok := envInit.Data["templates_copied"].([]any)
	if !ok || len(tmpls) != 6 {
		t.Fatalf("expected 6 templates copied, got: %v", envInit.Data["templates_copied"])
	}

	// 4. Save Diamond DAG Pipeline:
	//           S-01 (Layer 0)
	//          /    \
	//     S-02        S-03 (Layer 1 - Parallel)
	//          \    /
	//           S-04 (Layer 2)
	//             |
	//            FAD (Layer 3 - Synthesis)
	diamondPlan := `# Diamond Research Pipeline

## Session DAG

### Session S-01 — Core Domain Model
| Field | Value |
|---|---|
| **ID** | S-01 |
| **Layer** | 0 |
| **Door Type** | one-way |
| **Dependencies** | None |
| **Output File** | ` + "`S-01-core.md`" + ` |

` + "```prompt\nDefine the core domain entities and invariants.\n```" + `

### Session S-02 — Persistent Storage Layer
| Field | Value |
|---|---|
| **ID** | S-02 |
| **Layer** | 1 |
| **Door Type** | one-way |
| **Dependencies** | S-01 |
| **Output File** | ` + "`S-02-storage.md`" + ` |

` + "```prompt\nEvaluate embedded storage engines for the domain.\n```" + `

### Session S-03 — Network Protocol & Wire Format
| Field | Value |
|---|---|
| **ID** | S-03 |
| **Layer** | 1 |
| **Door Type** | two-way |
| **Dependencies** | S-01 |
| **Output File** | ` + "`S-03-network.md`" + ` |

` + "```prompt\nEvaluate gRPC vs FlatBuffers vs Cap'n Proto.\n```" + `

### Session S-04 — Distributed Consensus
| Field | Value |
|---|---|
| **ID** | S-04 |
| **Layer** | 2 |
| **Door Type** | one-way |
| **Dependencies** | S-02, S-03 |
| **Output File** | ` + "`S-04-consensus.md`" + ` |

` + "```prompt\nEvaluate consensus protocol given storage and wire constraints.\n```" + `

### Session FAD — Founding Architecture Document
| Field | Value |
|---|---|
| **ID** | FAD |
| **Layer** | 3 |
| **Door Type** | one-way |
| **Dependencies** | S-01, S-02, S-03, S-04 |
| **Output File** | ` + "`FAD.md`" + ` |

` + "```prompt\nSynthesize founding architecture for the complete system.\n```" + `
`
	envPlan := callAndParse("vivechak_save_plan", map[string]any{
		"project_root": projectDir,
		"content":      diamondPlan,
	})
	if !envPlan.Success {
		t.Fatalf("save_plan failed: %s", envPlan.Message)
	}

	// 5. Step 1: Layer 0 -> S-01
	envNext1 := callAndParse("vivechak_next_session", map[string]any{
		"project_root": projectDir,
	})
	if envNext1.Data["session_id"] != "S-01" {
		t.Fatalf("expected S-01, got %v", envNext1.Data["session_id"])
	}

	// Save S-01
	s1Content := `---
session_id: S-01
title: Core Domain Model
date: 2026-10-02
status: complete
---
# Core Domain Model
## Recommendations
Adopt immutable event-sourced aggregate root model. Grade A (formal specification)
## Discovered Concerns & Failure Modes
- High event stream volume may cause replay latency.
## Delta
| Prior | Status | Impact |
|---|---|---|
| Mutable CRUD state | Contradicted | High |
`
	envSave1 := callAndParse("vivechak_save_session", map[string]any{
		"project_root": projectDir,
		"session_id":   "S-01",
		"content":      s1Content,
	})
	if !envSave1.Success {
		t.Fatalf("save S-01 failed: %s", envSave1.Message)
	}

	// Record ADR D-001 for S-01
	d1Content := `---
id: D-001
title: Event-Sourced Core Domain
status: accepted
door_type: one-way
human_reviewed: true
review_trigger: "Event replay latency exceeds 100ms"
review_date: 2026-12-01
---
# Decision
We adopt event sourcing. Grade A (formal proof)
`
	envD1 := callAndParse("vivechak_record_decision", map[string]any{
		"project_root": projectDir,
		"decision_id":  "D-001",
		"content":      d1Content,
	})
	if !envD1.Success {
		t.Fatalf("record D-001 failed: %s", envD1.Message)
	}

	// 6. Step 2: Layer 1 -> BOTH S-02 and S-03 are ready in parallel!
	envNext2 := callAndParse("vivechak_next_session", map[string]any{
		"project_root": projectDir,
	})
	// Check parallelism_hint
	parallelismHint, ok := envNext2.Data["parallelism_hint"].(string)
	if !ok || !strings.Contains(parallelismHint, "PARALLEL EXECUTION AVAILABLE: 2 independent sessions") {
		t.Errorf("expected parallelism_hint for 2 parallel sessions, got: %v", envNext2.Data["parallelism_hint"])
	}
	if !strings.Contains(envNext2.NextStep, "⚡ 2 sessions ready in parallel") {
		t.Errorf("expected ⚡ indicator in next_step, got: %s", envNext2.NextStep)
	}

	// Verify blocked sessions list includes S-04 and FAD
	blocked, ok := envNext2.Data["blocked_sessions"].([]any)
	if !ok || len(blocked) < 2 {
		t.Errorf("expected at least 2 blocked sessions (S-04, FAD), got: %v", envNext2.Data["blocked_sessions"])
	}

	// 7. Save S-02 (Storage)
	s2Content := `---
session_id: S-02
title: Persistent Storage Layer
date: 2026-10-02
status: complete
---
# Persistent Storage Layer
## Recommendations
Adopt RocksDB for persistent event log. Grade A (benchmark)
## Delta
| Prior | Status | Impact |
|---|---|---|
| SQL database | Contradicted | High |
`
	envSave2 := callAndParse("vivechak_save_session", map[string]any{
		"project_root": projectDir,
		"session_id":   "S-02",
		"content":      s2Content,
	})
	if !envSave2.Success {
		t.Fatalf("save S-02 failed: %s", envSave2.Message)
	}

	// 8. Next session must now return S-03 (S-04 is still blocked because it needs both S-02 AND S-03!)
	envNext3 := callAndParse("vivechak_next_session", map[string]any{
		"project_root": projectDir,
	})
	if envNext3.Data["session_id"] != "S-03" {
		t.Fatalf("expected S-03, got %v", envNext3.Data["session_id"])
	}

	// Save S-03 (Network)
	s3Content := `---
session_id: S-03
title: Network Protocol & Wire Format
date: 2026-10-02
status: complete
---
# Network Protocol
## Recommendations
Adopt FlatBuffers over gRPC for zero-copy deserialization. Grade B (benchmark)
## Delta
| Prior | Status | Impact |
|---|---|---|
| JSON over HTTP | Contradicted | Medium |
`
	envSave3 := callAndParse("vivechak_save_session", map[string]any{
		"project_root": projectDir,
		"session_id":   "S-03",
		"content":      s3Content,
	})
	if !envSave3.Success {
		t.Fatalf("save S-03 failed: %s", envSave3.Message)
	}

	// 9. Now S-04 (Consensus) is unblocked!
	envNext4 := callAndParse("vivechak_next_session", map[string]any{
		"project_root": projectDir,
	})
	if envNext4.Data["session_id"] != "S-04" {
		t.Fatalf("expected S-04, got %v", envNext4.Data["session_id"])
	}

	// Save S-04
	s4Content := `---
session_id: S-04
title: Distributed Consensus
date: 2026-10-02
status: complete
---
# Distributed Consensus
## Recommendations
Adopt Raft consensus engine. Grade A (formal TLA+ spec)
## Delta
| Prior | Status | Impact |
|---|---|---|
| Gossip protocol | Contradicted | High |
`
	envSave4 := callAndParse("vivechak_save_session", map[string]any{
		"project_root": projectDir,
		"session_id":   "S-04",
		"content":      s4Content,
	})
	if !envSave4.Success {
		t.Fatalf("save S-04 failed: %s", envSave4.Message)
	}

	// 10. Step 4: Synthesis Session FAD
	envNextSyn := callAndParse("vivechak_next_session", map[string]any{
		"project_root": projectDir,
	})
	if envNextSyn.Data["session_id"] != "FAD" {
		t.Fatalf("expected FAD, got %v", envNextSyn.Data["session_id"])
	}

	fadContent := `---
session_id: FAD
title: Founding Architecture Document
date: 2026-10-02
status: complete
---
# Founding Architecture Document
## 1. System Vision
Diamond DAG resilient high-throughput event processing platform.
## 2. Core Architecture
Event-sourced RocksDB storage with FlatBuffers wire format and Raft consensus. Grade A (spec analysis)
## 3. Technology Choices
- Storage: RocksDB. Grade A (empirical data)
- Protocol: FlatBuffers. Grade B (benchmark)
- Consensus: Raft. Grade A (formal verification)
## 4. Operational Model & Mitigations
Replay performance monitored with snapshotting. Grade B (operational runbook)
## 5. Rejected Alternatives
- SQL rejected due to schema mutation bottlenecks. Grade A (benchmark)
- JSON HTTP rejected due to serialization overhead. Grade B (docs)
`
	envSaveFAD := callAndParse("vivechak_save_session", map[string]any{
		"project_root": projectDir,
		"session_id":   "FAD",
		"content":      fadContent,
	})
	if !envSaveFAD.Success {
		t.Fatalf("save FAD failed: %s", envSaveFAD.Message)
	}

	// Verify single-source persistence & byte-exact root copy
	fadPath := filepath.Join(projectDir, core.FADFile)
	fadBytes, err := os.ReadFile(fadPath)
	if err != nil {
		t.Fatalf("reading research/FAD.md: %v", err)
	}
	rootFADPath := filepath.Join(projectDir, "FOUNDING-ARCHITECTURE.md")
	rootFADBytes, err := os.ReadFile(rootFADPath)
	if err != nil {
		t.Fatalf("reading root FOUNDING-ARCHITECTURE.md: %v", err)
	}
	if string(fadBytes) != string(rootFADBytes) {
		t.Errorf("root FOUNDING-ARCHITECTURE.md differs from research/FAD.md")
	}

	// Verify NO duplicate session file created in research/sessions/FAD.md
	if _, err := os.Stat(filepath.Join(projectDir, core.SessionsDir, "FAD.md")); !os.IsNotExist(err) {
		t.Errorf("research/sessions/FAD.md should NOT exist (duplicate synthesis file)")
	}

	// 11. Multi-hop Downstream Stale Cascading Test:
	// Amend S-01: Because S-02 and S-03 depend on S-01, S-04 depends on S-02/S-03, and FAD depends on all,
	// amending S-01 must detect completed downstream sessions and issue W-STALE-DOWNSTREAM.
	envAmend := callAndParse("vivechak_amend_session", map[string]any{
		"project_root":        projectDir,
		"session_id":          "S-01",
		"amending_session_id": "REV-01",
		"amendment":           "Architecture revision: Aggregate roots must support snapshotting every 10,000 events.",
	})
	if !envAmend.Success {
		t.Fatalf("amend S-01 failed: %s", envAmend.Message)
	}
	staleList, ok := envAmend.Data["potentially_stale_sessions"].([]any)
	if !ok || len(staleList) == 0 {
		t.Fatalf("expected downstream stale sessions detected, got: %v", envAmend.Data["potentially_stale_sessions"])
	}
	hasStaleWarning := false
	for _, w := range envAmend.Warnings {
		if strings.Contains(w, "W-STALE-DOWNSTREAM") {
			hasStaleWarning = true
			break
		}
	}
	if !hasStaleWarning {
		t.Errorf("expected W-STALE-DOWNSTREAM warning on amending upstream S-01, got: %v", envAmend.Warnings)
	}

	// 12. Run Phase 0 Gate: Must pass and auto-persist research/PHASE-0-GATE.md
	envGate := callAndParse("vivechak_run_gate", map[string]any{
		"project_root": projectDir,
		"verbose":      true,
	})
	if envGate.Data["gate_status"] != "PASS" || envGate.Data["gate_passed"] != true {
		t.Fatalf("expected gate PASS, got status=%v passed=%v warnings=%v message=%s",
			envGate.Data["gate_status"], envGate.Data["gate_passed"], envGate.Warnings, envGate.Message)
	}

	gatePath := filepath.Join(projectDir, core.ResearchDir, "PHASE-0-GATE.md")
	gateBytes, err := os.ReadFile(gatePath)
	if err != nil {
		t.Fatalf("reading research/PHASE-0-GATE.md: %v", err)
	}
	if !strings.Contains(string(gateBytes), "verdict: \"PASS\"") {
		t.Errorf("expected verdict: PASS in gate artifact, got:\n%s", string(gateBytes))
	}

	// 13. Verify CLI doctor on the diamond workspace passes cleanly
	docCmd := exec.CommandContext(ctx, binPath, "doctor", projectDir)
	docOut, docErr := docCmd.CombinedOutput()
	if docErr != nil {
		t.Fatalf("doctor failed on diamond workspace: %v\nOutput: %s", docErr, string(docOut))
	}
	if !strings.Contains(string(docOut), "✓ Workspace exists") {
		t.Errorf("expected doctor to pass workspace check, got: %s", string(docOut))
	}
}

// TestMassive_LiveMCP_HighConcurrencyStress launches multiple goroutines issuing concurrent
// tool calls (saves, amends, records, validates) against a single live MCP subprocess
// to verify race immunity, advisory file locking atomicity, and zero workspace corruptions.
func TestMassive_LiveMCP_HighConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)

	binDir := t.TempDir()
	binName := "vck_conc_bin"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\noutput: %s", err, string(out))
	}

	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "concurrency-client", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	projectDir := t.TempDir()

	// Initialize workspace
	resInit, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_init",
		Arguments: map[string]any{
			"project_root": projectDir,
			"scope":        "project",
		},
	})
	if err != nil {
		t.Fatalf("init call failed: %v", err)
	}
	envInit := parseTestEnvelope(t, resInit)
	if !envInit.Success {
		t.Fatalf("init unsuccessful: %s", envInit.Message)
	}

	// Save a baseline decision D-001
	initD1 := `---
id: D-001
title: Baseline Architectural Decision
status: accepted
door_type: two-way
human_reviewed: false
---
# Baseline Decision
Initial baseline.
`
	_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_record_decision",
		Arguments: map[string]any{
			"project_root": projectDir,
			"decision_id":  "D-001",
			"content":      initD1,
		},
	})

	// Launch concurrent operations:
	// 5 workers writing distinct sessions
	// 5 workers recording distinct decisions
	// 2 workers calling validate
	var wg sync.WaitGroup
	errCount := 0
	var mu sync.Mutex

	recordErr := func(msg string) {
		mu.Lock()
		defer mu.Unlock()
		errCount++
		t.Logf("concurrency notice: %s", msg)
	}

	// 5 session writers
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sessID := fmt.Sprintf("S-%02d", id)
			sessBody := fmt.Sprintf(`---
session_id: %s
title: Concurrent Session %d
date: 2026-10-02
status: complete
---
# Concurrent Findings for Worker %d
Finding details for session %d. Grade B (automated run)
`, sessID, id, id, id)

			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_save_session",
				Arguments: map[string]any{
					"project_root": projectDir,
					"session_id":   sessID,
					"content":      sessBody,
				},
			})
			if err != nil {
				recordErr(fmt.Sprintf("save_session %s err: %v", sessID, err))
				return
			}
			env := parseTestEnvelope(t, res)
			if !env.Success {
				recordErr(fmt.Sprintf("save_session %s fail: %s", sessID, env.Message))
			}
		}(i)
	}

	// 5 decision writers
	for i := 2; i <= 6; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			decID := fmt.Sprintf("D-%03d", id)
			decBody := fmt.Sprintf(`---
id: %s
title: Concurrent Decision %d
status: accepted
door_type: two-way
human_reviewed: false
---
# Decision %d
Decision details for worker %d.
`, decID, id, id, id)

			res, err := cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_record_decision",
				Arguments: map[string]any{
					"project_root": projectDir,
					"decision_id":  decID,
					"content":      decBody,
				},
			})
			if err != nil {
				recordErr(fmt.Sprintf("record_decision %s err: %v", decID, err))
				return
			}
			env := parseTestEnvelope(t, res)
			if !env.Success {
				recordErr(fmt.Sprintf("record_decision %s fail: %s", decID, env.Message))
			}
		}(i)
	}

	// 2 concurrent validators
	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			_, _ = cs.CallTool(ctx, &mcp.CallToolParams{
				Name: "vivechak_validate",
				Arguments: map[string]any{
					"project_root": projectDir,
				},
			})
		}(i)
	}

	wg.Wait()

	if errCount > 0 {
		t.Fatalf("%d concurrent operations failed", errCount)
	}

	// Post-concurrency sanity check: Verify DECISIONS.md is not corrupted
	decData, err := os.ReadFile(filepath.Join(projectDir, core.DecisionsFile))
	if err != nil {
		t.Fatalf("reading DECISIONS.md after concurrency: %v", err)
	}
	decStr := string(decData)
	for i := 2; i <= 6; i++ {
		expectedID := fmt.Sprintf("D-%03d", i)
		if !strings.Contains(decStr, expectedID) {
			t.Errorf("DECISIONS.md missing concurrent decision %s", expectedID)
		}
	}

	// Verify all 5 sessions exist on disk and are readable
	for i := 1; i <= 5; i++ {
		sessPath := filepath.Join(projectDir, core.SessionsDir, fmt.Sprintf("S-%02d.md", i))
		if _, err := os.Stat(sessPath); err != nil {
			t.Errorf("missing session S-%02d on disk: %v", i, err)
		}
	}
}

// TestMassive_LiveMCP_MultiScopeFullRoundtrips tests decision-scope and comparison-scope
// pipelines end-to-end via real live subprocess stdio MCP connections.
func TestMassive_LiveMCP_MultiScopeFullRoundtrips(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)

	binDir := t.TempDir()
	binName := "vck_multiscope_bin"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\noutput: %s", err, string(out))
	}

	serverCmd := exec.CommandContext(ctx, binPath, "serve")
	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "multiscope-client", Version: "1.0.0"},
		nil,
	)

	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	callAndParse := func(toolName string, args map[string]any) testEnvelope {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		})
		if err != nil {
			t.Fatalf("call %s failed: %v", toolName, err)
		}
		return parseTestEnvelope(t, res)
	}

	// =========================================================================
	// Scope 1: Decision Scope Roundtrip
	// =========================================================================
	t.Run("DecisionScope_Roundtrip", func(t *testing.T) {
		dDir := t.TempDir()

		// 1. Init decision scope (must copy only 2 templates)
		envInit := callAndParse("vivechak_init", map[string]any{
			"project_root": dDir,
			"scope":        "decision",
		})
		if !envInit.Success {
			t.Fatalf("decision init failed: %s", envInit.Message)
		}
		tmpls, ok := envInit.Data["templates_copied"].([]any)
		if !ok || len(tmpls) != 2 {
			t.Fatalf("expected 2 templates copied for decision scope, got: %v", envInit.Data["templates_copied"])
		}

		// 2. Prepare generator (decision scope)
		envPrep := callAndParse("vivechak_prepare_generator", map[string]any{
			"project_root": dDir,
			"scope":        "decision",
			"context":      "Choose between CockroachDB vs TiDB for multi-region transactional SQL.",
		})
		if !envPrep.Success {
			t.Fatalf("prepare generator failed: %s", envPrep.Message)
		}
		prompt := envPrep.Data["prompt"].(string)
		if !strings.Contains(prompt, "CockroachDB vs TiDB") {
			t.Errorf("expected context injected into decision prompt")
		}

		// 3. Save decision plan (D-001-plan.md)
		planContent := `# Decision Research Pipeline: D-001

## Session DAG

### Session D-001-S1 — Distributed SQL Evaluation
| Field | Value |
|---|---|
| **ID** | D-001-S1 |
| **Layer** | 0 |
| **Door Type** | one-way |
| **Dependencies** | None |
| **Output File** | ` + "`D-001-S1.md`" + ` |

` + "```prompt\nEvaluate CockroachDB and TiDB against regional latency requirements.\n```" + `
`
		envPlan := callAndParse("vivechak_save_plan", map[string]any{
			"project_root": dDir,
			"content":      planContent,
			"scope":        "decision",
			"decision_id":  "D-001",
		})
		if !envPlan.Success {
			t.Fatalf("save decision plan failed: %s", envPlan.Message)
		}

		// 4. Next session -> D-001-S1
		envNext := callAndParse("vivechak_next_session", map[string]any{
			"project_root": dDir,
		})
		if envNext.Data["session_id"] != "D-001-S1" {
			t.Fatalf("expected session D-001-S1, got: %v", envNext.Data["session_id"])
		}

		// 5. Save session D-001-S1
		s1Content := `---
session_id: D-001-S1
title: Distributed SQL Evaluation
date: 2026-10-02
status: complete
---
# Evaluation Findings
## Recommendations
We recommend CockroachDB. Grade A (regional multi-cloud benchmark)
## Discovered Concerns & Failure Modes
- Range lease movement during WAN partition causes query timeout.
## Delta
| Prior | Status | Impact |
|---|---|---|
| Aurora Global DB | Contradicted | High |
`
		envSave := callAndParse("vivechak_save_session", map[string]any{
			"project_root": dDir,
			"session_id":   "D-001-S1",
			"content":      s1Content,
		})
		if !envSave.Success {
			t.Fatalf("save session D-001-S1 failed: %s", envSave.Message)
		}

		// 6. Record decision D-001
		d1Content := `---
id: D-001
title: Adopt CockroachDB for Multi-Region SQL
status: accepted
door_type: one-way
human_reviewed: true
review_trigger: "WAN partition query latency > 500ms"
review_date: 2026-12-01
---
# Decision
We adopt CockroachDB. Grade A (benchmark)
`
		envRec := callAndParse("vivechak_record_decision", map[string]any{
			"project_root": dDir,
			"decision_id":  "D-001",
			"content":      d1Content,
		})
		if !envRec.Success {
			t.Fatalf("record decision failed: %s", envRec.Message)
		}

		// 7. Validate: D-001-plan.md must be whitelisted by IsSpecialResearchFile
		envVal := callAndParse("vivechak_validate", map[string]any{
			"project_root": dDir,
		})
		if !envVal.Success {
			t.Fatalf("validate failed: %s", envVal.Message)
		}

		// 8. Run gate for decision scope
		envGate := callAndParse("vivechak_run_gate", map[string]any{
			"project_root": dDir,
			"verbose":      true,
		})
		if envGate.Data["gate_status"] != "PASS" || envGate.Data["gate_passed"] != true {
			t.Fatalf("expected gate PASS, got status=%v passed=%v warnings=%v message=%s",
				envGate.Data["gate_status"], envGate.Data["gate_passed"], envGate.Warnings, envGate.Message)
		}
	})

	// =========================================================================
	// Scope 2: Comparison Scope Roundtrip
	// =========================================================================
	t.Run("ComparisonScope_Roundtrip", func(t *testing.T) {
		cDir := t.TempDir()

		// 1. Init comparison scope (must copy only 1 template)
		envInit := callAndParse("vivechak_init", map[string]any{
			"project_root": cDir,
			"scope":        "comparison",
		})
		if !envInit.Success {
			t.Fatalf("comparison init failed: %s", envInit.Message)
		}
		tmpls, ok := envInit.Data["templates_copied"].([]any)
		if !ok || len(tmpls) != 1 {
			t.Fatalf("expected 1 template copied for comparison scope, got: %v", envInit.Data["templates_copied"])
		}

		// 2. Save comparison plan
		compPlan := `# Comparison Pipeline: Kafka vs Redpanda

## Session DAG

### Session COMP-01 — Streaming Broker Comparison
| Field | Value |
|---|---|
| **ID** | COMP-01 |
| **Layer** | 0 |
| **Door Type** | two-way |
| **Dependencies** | None |
| **Output File** | ` + "`COMP-01.md`" + ` |

` + "```prompt\nCompare Kafka vs Redpanda using WEP matrix.\n```" + `
`
		envPlan := callAndParse("vivechak_save_plan", map[string]any{
			"project_root": cDir,
			"content":      compPlan,
			"scope":        "comparison",
			"decision_id":  "COMP-01",
		})
		if !envPlan.Success {
			t.Fatalf("save comparison plan failed: %s", envPlan.Message)
		}

		// 3. Next session -> CMP-01
		envNext := callAndParse("vivechak_next_session", map[string]any{
			"project_root": cDir,
		})
		if envNext.Data["session_id"] != "CMP-01" {
			t.Fatalf("expected CMP-01, got: %v", envNext.Data["session_id"])
		}

		// 4. Save comparison session with WEP matrix
		compContent := `---
session_id: CMP-01
title: Kafka vs Redpanda Comparison
date: 2026-10-02
status: complete
---
# Comparison Session: Kafka vs Redpanda

## Weighted Evaluation Matrix (WEP)
| Criterion | Weight | Kafka | Redpanda |
|---|---|---|---|
| Latency p99 | 0.40 | 7 (Grade B) | 9 (Grade A) |
| Operational Simplicity | 0.30 | 5 (Grade B) | 9 (Grade A) |
| Ecosystem & Connectors | 0.30 | 10 (Grade A) | 7 (Grade B) |
| **Weighted Score** | **1.00** | **7.3** | **8.4** |

## Recommendation
Adopt Redpanda for low operational complexity and zero JVM overhead. Grade A (benchmark)
`
		envSave := callAndParse("vivechak_save_session", map[string]any{
			"project_root": cDir,
			"session_id":   "CMP-01",
			"content":      compContent,
		})
		if !envSave.Success {
			t.Fatalf("save CMP-01 failed: %s", envSave.Message)
		}

		// 5. Validate comparison workspace
		envVal := callAndParse("vivechak_validate", map[string]any{
			"project_root": cDir,
		})
		if !envVal.Success {
			t.Fatalf("validate comparison failed: %s", envVal.Message)
		}

		// 6. Run gate for comparison scope
		envGate := callAndParse("vivechak_run_gate", map[string]any{
			"project_root": cDir,
			"verbose":      true,
		})
		if envGate.Data["gate_status"] != "PASS" || envGate.Data["gate_passed"] != true {
			t.Fatalf("expected gate PASS for comparison, got status=%v passed=%v warnings=%v message=%s",
				envGate.Data["gate_status"], envGate.Data["gate_passed"], envGate.Warnings, envGate.Message)
		}
	})
}

// TestMassive_CLI_SubcommandMatrix tests the live compiled binary against all CLI
// subcommands, presets, arguments, and dry-run installations.
func TestMassive_CLI_SubcommandMatrix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)

	binDir := t.TempDir()
	binName := "vck_cli_matrix"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\noutput: %s", err, string(out))
	}

	// 1. Test version
	cmdVer := exec.CommandContext(ctx, binPath, "version")
	outVer, errVer := cmdVer.CombinedOutput()
	if errVer != nil {
		t.Fatalf("version failed: %v\noutput: %s", errVer, string(outVer))
	}
	if !strings.Contains(string(outVer), "vivechak") {
		t.Errorf("version output unexpected: %s", string(outVer))
	}

	// 2. Test setup --help
	cmdHelp := exec.CommandContext(ctx, binPath, "setup", "--help")
	outHelp, errHelp := cmdHelp.CombinedOutput()
	if errHelp != nil {
		t.Fatalf("setup --help failed: %v\noutput: %s", errHelp, string(outHelp))
	}
	if !strings.Contains(string(outHelp), "Available Presets:") {
		t.Errorf("setup --help output missing Available Presets:\n%s", string(outHelp))
	}

	// 3. Test mcp-config on all 14 presets: each must output valid JSON
	allPresets := []string{
		"claude-desktop", "cursor", "windsurf", "vscode", "gemini-cli",
		"claude-code", "codewhisperer", "cline", "continue", "roo-cline",
		"zed", "neovim", "jetbrains", "emacs",
	}
	for _, p := range allPresets {
		cmdCfg := exec.CommandContext(ctx, binPath, "mcp-config", p)
		outCfg, errCfg := cmdCfg.CombinedOutput()
		if errCfg != nil {
			t.Errorf("mcp-config %s failed: %v\noutput: %s", p, errCfg, string(outCfg))
			continue
		}
		var parsed map[string]any
		if err := json.Unmarshal(outCfg, &parsed); err != nil {
			t.Errorf("mcp-config %s produced invalid JSON: %v\noutput: %s", p, err, string(outCfg))
		}
	}

	// 4. Test setup --dry-run for key hosts
	for _, host := range []string{"vscode", "cursor", "zed", "claude-desktop"} {
		cmdDry := exec.CommandContext(ctx, binPath, "setup", "--dry-run", host)
		outDry, errDry := cmdDry.CombinedOutput()
		if errDry != nil {
			t.Errorf("setup --dry-run %s failed: %v\noutput: %s", host, errDry, string(outDry))
		}
		if !strings.Contains(string(outDry), "[dry-run]") {
			t.Errorf("setup --dry-run %s expected dry-run banner, got:\n%s", host, string(outDry))
		}
	}

	// 5. Test doctor on non-existent directory -> must fail
	cmdDocFail := exec.CommandContext(ctx, binPath, "doctor", filepath.Join(binDir, "non_existent_workspace_dir"))
	outDocFail, errDocFail := cmdDocFail.CombinedOutput()
	if errDocFail == nil {
		t.Errorf("doctor on non-existent directory should have failed, but passed: %s", string(outDocFail))
	}
}
