package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// gtPipeline is a three-session project pipeline in GENERATOR-PROJECT.md's shape, with the machine-readable plan
// block and a topology table; every session and the decision are two-way doors, so the gate's Track B rules stay out
// of the way of what these tests pin (targets, G2).
const gtPipeline = `# Research Pipeline: Event platform

` + "```yaml vivechak-plan" + `
plan: PROJECT
sessions:
  - {id: T1-01, title: Datastore, layer: 1, door: two-way, informs: [D-001], depends: []}
  - {id: T1-02, title: Queue, layer: 1, door: two-way, depends: []}
  - {id: T1-03, title: Auth, layer: 1, door: two-way, depends: []}
  - {id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}
decisions:
  - {id: D-001, title: Primary datastore, door: two-way}
` + "```" + `

| ID | Topic | Depends On | Output File |
|---|---|---|---|
| T1-01 | Datastore | — | research/sessions/T1-01.md |
| T1-02 | Queue | — | research/sessions/T1-02.md |
| T1-03 | Auth | — | research/sessions/T1-03.md |
| SYN-01 | Synthesis | T1-01, T1-02, T1-03 | research/FAD.md |

## Sessions

### T1-01: Datastore
` + "```prompt" + `
Research datastores. [UPSTREAM_FINDINGS]
` + "```" + `

### T1-02: Queue
` + "```prompt" + `
Research queues. [UPSTREAM_FINDINGS]
` + "```" + `

### T1-03: Auth
` + "```prompt" + `
Research auth. [UPSTREAM_FINDINGS]
` + "```" + `

### SYN-01: Synthesis
` + "```prompt" + `
Synthesize. [ALL_SESSION_FINDINGS]
` + "```" + `
`

// gtFAD is a FAD (SYN-01) with the four sections G3 requires and D-001 in its registry.
const gtFAD = `---
id: SYN-01
title: "Founding Architecture"
date: 2026-10-25
status: complete
---
# Founding Architecture
## Executive Summary
PostgreSQL is the primary datastore [A · single · fresh · direct · fetched](https://www.postgresql.org/docs/current/).
## Locked Decision Registry
| ID | Status |
|---|---|
| D-001 | accepted |
## Risk Register & Reversal Triggers
| Risk | Trigger |
|---|---|
| Write amplification | p95 write latency above 50 ms |
## Traceability Matrix
| Decision | Sessions |
|---|---|
| D-001 | T1-01 |
`

// gtGateCLI runs `vck gate [args] <dir>` and returns its exit code and combined output.
func gtGateCLI(t *testing.T, binPath, dir string, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(binPath, append(append([]string{"gate"}, args...), dir)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return 0, string(out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("running vck gate: %v", err)
	}
	return exitErr.ExitCode(), string(out)
}

// gtRules returns a vivechak_gate result's rules as id -> status.
func gtRules(t *testing.T, env testEnvelope) map[string]string {
	t.Helper()
	raw, err := json.Marshal(env.Data["rules"])
	if err != nil {
		t.Fatalf("encoding rules: %v", err)
	}
	var rules []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(raw, &rules); err != nil {
		t.Fatalf("decoding rules %s: %v", raw, err)
	}
	out := make(map[string]string, len(rules))
	for _, r := range rules {
		out[r.ID] = r.Status
	}
	return out
}

// gtRuleReason returns the reason the gate gives for rule id.
func gtRuleReason(env testEnvelope, id string) string {
	rules, _ := env.Data["rules"].([]any)
	for _, r := range rules {
		if m, ok := r.(map[string]any); ok && m["id"] == id {
			reason, _ := m["reason"].(string)
			return reason
		}
	}
	return ""
}

// gtAssertGateCLIFails asserts that `vck gate` prints FAIL on its first line and exits 1 for every --track value.
func gtAssertGateCLIFails(t *testing.T, binPath, dir, when string) {
	t.Helper()
	for _, track := range [][]string{nil, {"--track", "track_a"}, {"--track", "track_b"}, {"--track", "all"}} {
		code, out := gtGateCLI(t, binPath, dir, track...)
		first := strings.TrimSpace(strings.SplitN(strings.TrimSpace(out), "\n", 2)[0])
		if code != 1 || first != "Phase 0 Gate: FAIL" {
			t.Errorf("%s: `vck gate %s` = exit %d, first line %q; want exit 1 and \"Phase 0 Gate: FAIL\"\n%s",
				when, strings.Join(track, " "), code, first, out)
		}
	}
}

func gtMustCall(t *testing.T, ctx context.Context, cs *mcp.ClientSession, name string, args map[string]any) testEnvelope {
	t.Helper()
	env, isErr := wgCall(t, ctx, cs, name, args)
	if isErr {
		t.Fatalf("%s %v failed: %s", name, args["kind"], env.Message)
	}
	return env
}

// TestGateTargets_LiveSubprocess covers CRIT-12 on the built binary over stdio and the `vck gate` CLI: the gate and
// vivechak_next decide from the files on disk, never from the last vivechak_start or a comparison saved on the side
// (HIGH-18). Each sequence is one a user hits in normal work: a quick comparison during a decision or a project, or a
// comparison before the first ADR.
func TestGateTargets_LiveSubprocess(t *testing.T) {
	binPath := resolveOrBuildTestBinary(t)

	// An unsigned one-way decision FAILs the gate (G10, G11) before and after a later comparison start and save;
	// the comparison never turns the FAIL into NOT_APPLICABLE or moves the target off the decision.
	t.Run("DecisionThenComparison_StillFails", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		gtMustCall(t, ctx, cs, "vivechak_start", map[string]any{"scope": "decision", "question": "Primary datastore"})
		gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "D-001", "content": wgDecisionPlan("D-001", "Primary datastore")})
		gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": "D-001-S1", "content": wgSession("D-001-S1", "D-001", "Landscape")})
		gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": "D-001-S2", "content": wgSession("D-001-S2", "D-001", "Falsify")})
		gtMustCall(t, ctx, cs, "vivechak_record_decision", map[string]any{
			"id":      "D-001",
			"content": wgADR("D-001", "Primary datastore", "accepted", "one-way", "D-001-S1", "D-001-S2"),
		})
		want := map[string]string{
			"G1": "PASS", "G2": "PASS", "G3": "n/a", "G4": "PASS", "G5": "PASS", "G6": "PASS",
			"G7": "PASS", "G8": "PASS", "G9": "PASS", "G10": "FAIL", "G11": "FAIL", "G12": "n/a",
		}
		check := func(when string) {
			t.Helper()
			gate := gtMustCall(t, ctx, cs, "vivechak_gate", map[string]any{})
			if gate.Data["verdict"] != "FAIL" || gate.Data["target"] != "D-001" {
				t.Errorf("%s: expected verdict FAIL for target D-001, got verdict %v target %v (%s)", when, gate.Data["verdict"], gate.Data["target"], gate.Message)
			}
			if got := gtRules(t, gate); !reflect.DeepEqual(got, want) {
				t.Errorf("%s: gate rules = %v, want %v", when, got, want)
			}
			gtAssertGateCLIFails(t, binPath, dir, when)
		}
		check("before the comparison")

		gtMustCall(t, ctx, cs, "vivechak_start", map[string]any{"scope": "comparison", "question": "Pino vs Winston"})
		gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "comparison", "content": wgComparison("Pino vs Winston", "Pino", "Winston")})
		check("after a comparison start and save")
	})

	// A project still FAILs (target project) and vivechak_next still serves its sessions after a comparison: started
	// mid-project, and (variant) when the workspace began as a comparison and the pipeline came later without a
	// project start, so the stored scope says comparison.
	t.Run("ProjectThenComparison_NextServesPipeline_GateFails", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		type step struct {
			tool string
			args map[string]any
		}
		for _, tc := range []struct {
			name  string
			steps []step
		}{
			{"project start, pipeline, then a comparison", []step{
				{"vivechak_start", map[string]any{"scope": "project", "question": "Event platform"}},
				{"vivechak_save", map[string]any{"kind": "plan", "id": "DAG", "content": gtPipeline}},
				{"vivechak_start", map[string]any{"scope": "comparison", "question": "Zod vs Valibot"}},
				{"vivechak_save", map[string]any{"kind": "comparison", "content": wgComparison("Zod vs Valibot", "Zod", "Valibot")}},
			}},
			{"comparison start and save, then the pipeline", []step{
				{"vivechak_start", map[string]any{"scope": "comparison", "question": "Zod vs Valibot"}},
				{"vivechak_save", map[string]any{"kind": "comparison", "content": wgComparison("Zod vs Valibot", "Zod", "Valibot")}},
				{"vivechak_save", map[string]any{"kind": "plan", "id": "DAG", "content": gtPipeline}},
			}},
		} {
			dir := wgNewProject(t)
			cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
			for _, s := range tc.steps {
				gtMustCall(t, ctx, cs, s.tool, s.args)
			}
			next := gtMustCall(t, ctx, cs, "vivechak_next", map[string]any{})
			if next.Data["session_id"] != "T1-01" {
				t.Errorf("%s: expected vivechak_next to serve T1-01, got %v (%s)", tc.name, next.Data["session_id"], next.Message)
			}
			gate := gtMustCall(t, ctx, cs, "vivechak_gate", map[string]any{})
			rules := gtRules(t, gate)
			if gate.Data["verdict"] != "FAIL" || gate.Data["target"] != "project" || rules["G1"] != "PASS" || rules["G2"] != "FAIL" {
				t.Errorf("%s: expected verdict FAIL for target project with G1 PASS and G2 FAIL, got verdict %v target %v rules %v (%s)",
					tc.name, gate.Data["verdict"], gate.Data["target"], rules, gate.Message)
			}
			cleanup()
			gtAssertGateCLIFails(t, binPath, dir, tc.name)
		}
	})

	// A comparison first, then an accepted one-way ADR recorded directly (no plan): the gate judges the ADR and FAILs;
	// the comparison does not make the workspace NOT_APPLICABLE. Variant: the same without any vivechak_start.
	t.Run("ComparisonThenOneWayADR_Fails", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		for _, withStart := range []bool{true, false} {
			dir := wgNewProject(t)
			cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
			if withStart {
				gtMustCall(t, ctx, cs, "vivechak_start", map[string]any{"scope": "comparison", "question": "Pino vs Winston"})
			}
			gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "comparison", "content": wgComparison("Pino vs Winston", "Pino", "Winston")})
			gtMustCall(t, ctx, cs, "vivechak_record_decision", map[string]any{
				"id":      "D-001",
				"content": wgADR("D-001", "Primary datastore", "accepted", "one-way"),
			})
			gate := gtMustCall(t, ctx, cs, "vivechak_gate", map[string]any{})
			want := map[string]string{
				"G1": "PASS", "G2": "PASS", "G3": "n/a", "G4": "PASS", "G5": "PASS", "G6": "FAIL",
				"G7": "FAIL", "G8": "PASS", "G9": "PASS", "G10": "FAIL", "G11": "FAIL", "G12": "n/a",
			}
			if gate.Data["verdict"] != "FAIL" || gate.Data["target"] != "D-001" {
				t.Errorf("start=%v: expected verdict FAIL for target D-001, got verdict %v target %v (%s)", withStart, gate.Data["verdict"], gate.Data["target"], gate.Message)
			}
			if got := gtRules(t, gate); !reflect.DeepEqual(got, want) {
				t.Errorf("start=%v: gate rules = %v, want %v", withStart, got, want)
			}
			cleanup()
			art, err := os.ReadFile(filepath.Join(dir, "research", "PHASE-0-GATE.md"))
			if fm, _, _ := core.ParseFrontmatter(art); err != nil || fm.GetString("verdict") != "FAIL" {
				t.Errorf("start=%v: expected research/PHASE-0-GATE.md with verdict FAIL (err %v)", withStart, err)
			}
			gtAssertGateCLIFails(t, binPath, dir, "comparison then a one-way ADR")
		}
	})

	// G2 judges each planned session's content, not the status it claims: a stub re-saved through vivechak_save with
	// status: complete and a TODO body (stored as complete: validation only warns), and a session edited on disk into a
	// note without findings, recommendation, ledger or grades, both FAIL G2 by name. The gate passed before the stubs.
	t.Run("G2StubSessions_FailNamingEachStub", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		dir := wgNewProject(t)
		cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
		defer cleanup()

		gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "DAG", "content": gtPipeline})
		paths := map[string]string{}
		for _, s := range []struct{ id, title string }{{"T1-01", "Datastore"}, {"T1-02", "Queue"}, {"T1-03", "Auth"}} {
			env := gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": s.id, "content": wgSession(s.id, "D-001", s.title)})
			paths[s.id], _ = env.Data["file_path"].(string)
		}
		gtMustCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-001", "content": wgADR("D-001", "Primary datastore", "accepted", "two-way", "T1-01")})
		gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "fad", "id": "SYN-01", "content": gtFAD})
		gate := gtMustCall(t, ctx, cs, "vivechak_gate", map[string]any{})
		wantPass := map[string]string{
			"G1": "PASS", "G2": "PASS", "G3": "PASS", "G4": "PASS", "G5": "PASS", "G6": "n/a",
			"G7": "n/a", "G8": "n/a", "G9": "n/a", "G10": "n/a", "G11": "n/a", "G12": "PASS",
		}
		if got := gtRules(t, gate); gate.Data["verdict"] != "PASS" || !reflect.DeepEqual(got, wantPass) {
			t.Fatalf("before the stubs: expected PASS with rules %v, got %v %v (%s)", wantPass, gate.Data["verdict"], got, gate.Message)
		}
		if code, out := gtGateCLI(t, binPath, dir); code != 0 {
			t.Fatalf("before the stubs: `vck gate` exit %d, want 0\n%s", code, out)
		}

		// T1-02 re-saved through the tool as a stub that claims status: complete. Validation only warns (no date, no
		// grades), so the file keeps status: complete and only G2's content check can catch it.
		gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": "T1-02", "revise": true,
			"content": "---\nid: T1-02\ntitle: Queue\nstatus: complete\n---\nTODO\n"})
		stub, _ := os.ReadFile(filepath.Join(dir, filepath.FromSlash(paths["T1-02"])))
		if fm, _, _ := core.ParseFrontmatter(stub); fm.GetString("status") != "complete" || !strings.Contains(string(stub), "\nTODO\n") {
			t.Fatalf("expected the re-saved T1-02 stub on disk with status complete, got:\n%s", stub)
		}
		// T1-03 edited on disk afterwards: status complete, but no findings, recommendation, ledger or grades.
		edited := "---\nid: T1-03\ntitle: Auth\nstatus: complete\n---\n## Notes\nWe looked at a few auth providers and they all seem fine for our needs.\n"
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(paths["T1-03"])), []byte(edited), 0o644); err != nil {
			t.Fatal(err)
		}

		gate = gtMustCall(t, ctx, cs, "vivechak_gate", map[string]any{})
		wantFail := map[string]string{}
		for k, v := range wantPass {
			wantFail[k] = v
		}
		wantFail["G2"] = "FAIL"
		if got := gtRules(t, gate); gate.Data["verdict"] != "FAIL" || !reflect.DeepEqual(got, wantFail) {
			t.Errorf("after the stubs: expected FAIL with rules %v, got %v %v (%s)", wantFail, gate.Data["verdict"], got, gate.Message)
		}
		if reason := gtRuleReason(gate, "G2"); reason != "Planned session(s) missing or incomplete: T1-02, T1-03" {
			t.Errorf("G2 must name exactly the two stub sessions, got %q", reason)
		}
		code, out := gtGateCLI(t, binPath, dir)
		if code != 1 || !strings.HasPrefix(out, "Phase 0 Gate: FAIL") || strings.Contains(out, "WARN") || !strings.Contains(out, "T1-03") {
			t.Errorf("after the stubs: `vck gate` must exit 1, print FAIL naming T1-03 and no WARN; got exit %d\n%s", code, out)
		}
	})
}

// TestGateChallengeRecord_LiveSubprocess covers G10 on the built binary over stdio: a one-way decision passes G10 only
// with the challenge record vivechak_review(mode="red_team") wrote for its current outcome, cited in the ADR. The same
// record with its decision_hash removed (a hand-edited or hand-written record), or a record left behind by a later
// change to the decision outcome, is not current: G10 FAILs naming it.
func TestGateChallengeRecord_LiveSubprocess(t *testing.T) {
	binPath := resolveOrBuildTestBinary(t)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	dir := wgNewProject(t)
	cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
	defer cleanup()

	gtMustCall(t, ctx, cs, "vivechak_start", map[string]any{"scope": "decision", "question": "Primary datastore"})
	gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "plan", "id": "D-001", "content": wgDecisionPlan("D-001", "Primary datastore")})
	gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": "D-001-S1", "content": wgSession("D-001-S1", "D-001", "Landscape")})
	gtMustCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "session", "id": "D-001-S2", "content": wgSession("D-001-S2", "D-001", "Falsify")})
	adr := wgADR("D-001", "Primary datastore", "accepted", "one-way", "D-001-S1", "D-001-S2")
	gtMustCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-001", "content": adr})

	red := gtMustCall(t, ctx, cs, "vivechak_review", map[string]any{"mode": "red_team", "id": "D-001"})
	recID, _ := red.Data["record_id"].(string)
	recPath := filepath.Join(dir, "research", recID+".md")
	record, err := os.ReadFile(recPath)
	if recID != "D-001-R1" || err != nil || !strings.Contains(string(record), "\ndecision_hash: ") {
		t.Fatalf("expected the red-team record research/D-001-R1.md with a decision_hash, got id %q (err %v):\n%s", recID, err, record)
	}
	cited := adr + "## Challenge Findings\nRed-team challenge " + recID + " completed: write amplification at 5k writes/s was load-tested and stays within the latency budget.\n"
	gtMustCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-001", "content": cited, "revise": true})

	g10 := func(when string) (string, string) {
		t.Helper()
		gate := gtMustCall(t, ctx, cs, "vivechak_gate", map[string]any{})
		return gtRules(t, gate)["G10"], gtRuleReason(gate, "G10")
	}
	if status, reason := g10("with the current, cited record"); status != "PASS" {
		t.Fatalf("with the current, cited red-team record G10 must PASS, got %s: %s", status, reason)
	}

	var stripped []string
	for _, line := range strings.Split(string(record), "\n") {
		if !strings.HasPrefix(line, "decision_hash:") {
			stripped = append(stripped, line)
		}
	}
	if err := os.WriteFile(recPath, []byte(strings.Join(stripped, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	if status, reason := g10("record without decision_hash"); status != "FAIL" || !strings.Contains(reason, "stale challenge record (D-001-R1)") {
		t.Errorf("a challenge record without decision_hash must not count as current: expected G10 FAIL naming D-001-R1, got %s: %s", status, reason)
	}

	// Variant: the record is intact but the decision outcome changed after the red-team run.
	if err := os.WriteFile(recPath, record, 0o644); err != nil {
		t.Fatal(err)
	}
	if status, reason := g10("record restored"); status != "PASS" {
		t.Fatalf("with the record restored G10 must PASS again, got %s: %s", status, reason)
	}
	changed := strings.Replace(cited, "Chosen option for Primary datastore.", "Chosen option for Primary datastore: CockroachDB instead of PostgreSQL.", 1)
	gtMustCall(t, ctx, cs, "vivechak_record_decision", map[string]any{"id": "D-001", "content": changed, "revise": true})
	if status, reason := g10("outcome changed"); status != "FAIL" || !strings.Contains(reason, "stale challenge record (D-001-R1)") {
		t.Errorf("after the decision outcome changed G10 must FAIL naming the stale D-001-R1, got %s: %s", status, reason)
	}
}
