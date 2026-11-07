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

// TestWriteGuard_LiveSubprocess verifies that the stdio MCP server enforces
// write guards (E-REVISE-REQUIRED on duplicate saves, .history/ archival on revise: true,
// draft status on ADR auto_save with placeholders, and DECISIONS.md preamble preservation).
func TestWriteGuard_LiveSubprocess(t *testing.T) {
	binPath := resolveOrBuildTestBinary(t)

	// 1. Refuse second save of different content without revise: true
	t.Run("Refuse_SecondSave_DifferentContent", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testproject\n"), 0o644)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		startRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_start",
			Arguments: map[string]any{
				"scope":    "comparison",
				"question": "Pino vs Winston",
			},
		})
		if err != nil || startRes.IsError {
			t.Fatalf("vivechak_start failed: %v", err)
		}

		contentV1 := `---
id: C-001
type: comparison
title: Pino vs Winston
status: draft
---
# Pino vs Winston
Version 1 comparison body text.
`
		save1Res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "comparison",
				"id":      "C-001",
				"content": contentV1,
			},
		})
		if err != nil || save1Res.IsError {
			t.Fatalf("first save failed: %v", err)
		}

		filePath := filepath.Join(dir, "research", "C-001-comparison.md")
		readV1, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("reading saved file: %v", err)
		}
		if !strings.Contains(string(readV1), "Version 1") {
			t.Errorf("expected Version 1 in saved file, got: %s", string(readV1))
		}

		// Second save with differing content and NO revise flag -> must be refused with E-REVISE-REQUIRED
		contentV2 := `---
id: C-001
type: comparison
title: Pino vs Winston
status: draft
---
# Pino vs Winston
Version 2 differing content.
`
		save2Res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "comparison",
				"id":      "C-001",
				"content": contentV2,
			},
		})
		if err != nil {
			t.Fatalf("call error on second save: %v", err)
		}
		if !save2Res.IsError {
			env := parseTestEnvelope(t, save2Res)
			if env.Success {
				t.Fatalf("expected second save to fail with E-REVISE-REQUIRED, got success")
			}
			if !strings.Contains(env.Message, "E-REVISE-REQUIRED") {
				t.Errorf("expected E-REVISE-REQUIRED in message, got: %s", env.Message)
			}
		} else {
			// Transport error text
			env := parseTestEnvelope(t, save2Res)
			if !strings.Contains(env.Message, "E-REVISE-REQUIRED") {
				t.Errorf("expected E-REVISE-REQUIRED in error message, got: %s", env.Message)
			}
		}

		// Verify file remains unchanged
		readAfter, _ := os.ReadFile(filePath)
		if !strings.Contains(string(readAfter), "Version 1") {
			t.Errorf("file was overwritten despite refusal: %s", string(readAfter))
		}

		// 2. Saving with revise: true archives old version to research/.history/
		saveReviseRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "comparison",
				"id":      "C-001",
				"content": contentV2,
				"revise":  true,
			},
		})
		if err != nil || saveReviseRes.IsError {
			t.Fatalf("save with revise: true failed: %v", err)
		}

		readV2, _ := os.ReadFile(filePath)
		if !strings.Contains(string(readV2), "Version 2") {
			t.Errorf("expected Version 2 in revised file, got: %s", string(readV2))
		}

		// Check .history directory for archived version
		historyDir := filepath.Join(dir, "research", ".history")
		entries, hErr := os.ReadDir(historyDir)
		if hErr != nil || len(entries) == 0 {
			t.Fatalf("expected archived version in research/.history, got err=%v, count=%d", hErr, len(entries))
		}
	})

	// 3. auto_save draft ADR with placeholders saves status: proposed, response status: draft,
	// and subsequent auto_save archives previous version into research/.history/
	t.Run("AutoSave_DraftWithPlaceholders_ThenSecondAutoSave", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testproject\n"), 0o644)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		// Start decision scope
		_, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_start",
			Arguments: map[string]any{
				"scope":    "decision",
				"question": "Postgres vs MySQL",
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Save decision plan
		planContent := `# Plan for D-001
| ID | Topic |
|---|---|
| D-001-S1 | Datastore Landscape |
`
		_, err = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "plan",
				"id":      "D-001",
				"content": planContent,
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Save session D-001-S1
		sessionContent := `---
id: D-001-S1
session_id: D-001-S1
type: session
title: Datastore Landscape
status: complete
informs_decisions: ["D-001"]
vivechak: 1
---
# Datastore Landscape
## Key Findings
- PostgreSQL has advanced JSONB indexing [Grade A: https://postgresql.org | fetched].
- MySQL has simpler replication setup [Grade B: https://mysql.com | fetched].

## Recommendation
We select PostgreSQL [Grade A: https://postgresql.org | fetched].

## Discovered Concerns & Tradeoffs
- PostgreSQL vacuum tuning requires operational monitoring at high write volumes.

## Sources & Evidence Ledger
- PostgreSQL Docs: https://postgresql.org [Grade A | fetched]
- MySQL Docs: https://mysql.com [Grade B | fetched]
`
		_, err = cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_save",
			Arguments: map[string]any{
				"kind":    "session",
				"id":      "D-001-S1",
				"content": sessionContent,
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		// Call vivechak_record_decision with auto_draft_from and auto_save: true
		recRes1, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"decision_id":     "D-001",
				"auto_draft_from": "D-001-S1",
				"auto_save":       true,
			},
		})
		if err != nil || recRes1.IsError {
			envFail := parseTestEnvelope(t, recRes1)
			t.Fatalf("first auto_save record_decision failed (err=%v, isError=%v, msg=%s, data=%v)", err, recRes1.IsError, envFail.Message, envFail.Data)
		}

		env1 := parseTestEnvelope(t, recRes1)
		if !env1.Success {
			t.Fatalf("expected envelope success: true, got false (%s)", env1.Message)
		}
		if env1.Data["status"] != "draft" {
			t.Errorf("expected data.status == 'draft' when placeholders remain, got: %v", env1.Data["status"])
		}
		if phs, ok := env1.Data["placeholders"].([]any); !ok || len(phs) == 0 {
			t.Errorf("expected data.placeholders to have items, got: %v", env1.Data["placeholders"])
		}

		// Check saved file on disk: status must be proposed
		adrPath := filepath.Join(dir, "research", "D-001-decision.md")
		adrBytes, err := os.ReadFile(adrPath)
		if err != nil {
			t.Fatalf("reading saved ADR file: %v", err)
		}
		if !strings.Contains(string(adrBytes), "status: proposed") {
			t.Errorf("expected saved ADR to have status: proposed, got:\n%s", string(adrBytes))
		}

		// Call auto_save a second time on the existing ADR
		recRes2, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"decision_id":     "D-001",
				"auto_draft_from": "D-001-S1",
				"auto_save":       true,
			},
		})
		if err != nil || recRes2.IsError {
			t.Fatalf("second auto_save record_decision failed: %v", err)
		}

		// Check that research/.history has an archived file from the previous version
		histDir := filepath.Join(dir, "research", ".history")
		histEntries, hErr := os.ReadDir(histDir)
		if hErr != nil || len(histEntries) == 0 {
			t.Fatalf("expected archived version in research/.history, got err=%v, count=%d", hErr, len(histEntries))
		}
	})

	// 4. Text outside DECISIONS.md boundary markers survives vivechak_record_decision
	t.Run("DecisionsMDPreamblePreservation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testproject\n"), 0o644)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		_ = os.MkdirAll(filepath.Join(dir, "research"), 0o755)
		decisionsFile := filepath.Join(dir, "research", "DECISIONS.md")

		initialContent := `# Custom Architecture Team Decisions Handbook

This preamble contains important guidelines for all architectural decisions.
Do not overwrite these custom instructions.

<!-- VIVECHAK:START -->
# Decision Registry
| ID | Title | Status |
|---|---|---|
<!-- VIVECHAK:END -->

## Postamble Appendix
Signed off by Architecture Guild.
`
		if err := os.WriteFile(decisionsFile, []byte(initialContent), 0o644); err != nil {
			t.Fatal(err)
		}

		adrContent := `---
id: D-001
type: decision
title: Custom Framework Choice
door_type: two-way
status: accepted
review_trigger: "Throughput drops below 1000 req/s"
date: 2026-11-04
vivechak: 1
---
# Custom Framework Choice
## Context
We need a lightweight framework.
## Decision Outcome
We choose Go stdlib.
`
		recRes, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name: "vivechak_record_decision",
			Arguments: map[string]any{
				"decision_id": "D-001",
				"content":     adrContent,
			},
		})
		if err != nil || recRes.IsError {
			envFail := parseTestEnvelope(t, recRes)
			t.Fatalf("record_decision failed (err=%v, isError=%v, msg=%s, data=%v)", err, recRes.IsError, envFail.Message, envFail.Data)
		}

		afterBytes, err := os.ReadFile(decisionsFile)
		if err != nil {
			t.Fatalf("reading DECISIONS.md after record: %v", err)
		}
		afterStr := string(afterBytes)

		if !strings.Contains(afterStr, "Custom Architecture Team Decisions Handbook") {
			t.Errorf("preamble was lost in DECISIONS.md:\n%s", afterStr)
		}
		if !strings.Contains(afterStr, "This preamble contains important guidelines") {
			t.Errorf("preamble body was lost in DECISIONS.md:\n%s", afterStr)
		}
		if !strings.Contains(afterStr, "Postamble Appendix") {
			t.Errorf("postamble was lost in DECISIONS.md:\n%s", afterStr)
		}
		if !strings.Contains(afterStr, "D-001") {
			t.Errorf("D-001 entry missing inside DECISIONS.md markers:\n%s", afterStr)
		}
	})
}
