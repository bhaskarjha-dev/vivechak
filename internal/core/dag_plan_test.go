package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

// Plans in the shape the Harness and the generators use (HIGH-07): the vivechak-plan block plus one prompt fence
// per session under "## Sessions"; the tables in between are prose.
var (
	harnessDecisionPlan = md(`# D-001-PLAN: Primary datastore

<F>yaml vivechak-plan
plan: D-001
sessions:
  - {id: D-001-S1, title: Landscape, role: L, depends: []}
  - {id: D-001-S2, title: Falsify, role: F, depends: [D-001-S1]}
<F>

## Routing
| Factor | Score | Rationale |
|---|---|---|
| R | 3 | data migration |

## Sessions

### D-001-S1: Landscape
| ID | Role | Depends | Filename |
|---|---|---|---|
| D-001-S1 | L | none | sessions/D-001-S1-landscape.md |

<F>prompt
# RESEARCH BRIEF: datastore landscape
Map the credible primary datastores.
<F>

### D-001-S2: Falsify
| ID | Role | Depends | Filename |
|---|---|---|---|
| D-001-S2 | F | D-001-S1 | sessions/D-001-S2-falsify.md |

<F>prompt
# RESEARCH BRIEF: how PostgreSQL could fail
Find failure modes and reversal triggers.
<F>
`)

	harnessPipeline = md(`# Research Pipeline: Demo

<F>yaml vivechak-plan
plan: PROJECT
sessions:
  - {id: T1-01, title: Datastore, layer: 1, door: one-way, informs: [D-001], depends: []}
  - {id: T1-02, title: Queue, layer: 1, door: two-way, depends: []}
  - {id: T1-03, title: Auth, layer: 1, door: two-way, depends: []}
  - {id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}
decisions:
  - {id: D-001, title: Primary datastore, door: one-way}
<F>

| ID | Topic | Depends On | Output File |
|---|---|---|---|
| T1-01 | Datastore | — | research/sessions/T1-01-datastore.md |
| T1-02 | Queue | — | research/sessions/T1-02-queue.md |
| T1-03 | Auth | — | research/sessions/T1-03-auth.md |
| SYN-01 | Synthesis | T1-01, T1-02, T1-03 | research/FAD.md |

## Sessions

### T1-01: Datastore
<F>prompt
Research datastores. [UPSTREAM_FINDINGS]
<F>

### T1-02: Queue
<F>prompt
Research queues. [UPSTREAM_FINDINGS]
<F>

### T1-03: Auth
<F>prompt
Research auth. [UPSTREAM_FINDINGS]
<F>

### SYN-01: Synthesis
<F>prompt
Synthesize. [ALL_SESSION_FINDINGS]
<F>
`)
)

type wantSession struct {
	id, deps, prompt string
}

func checkPlanSessions(t *testing.T, name, plan string, want []wantSession) *DAG {
	t.Helper()
	if res := ValidatePlan([]byte(plan)); len(res.Issues) != 0 {
		t.Fatalf("%s: expected no issues, got %+v", name, res.Issues)
	}
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatalf("%s: ParsePipeline: %v", name, err)
	}
	if len(dag.Sessions) != len(want) {
		t.Fatalf("%s: sessions %v, want %d", name, idsOf(dag.Sessions), len(want))
	}
	for i, w := range want {
		s := dag.Sessions[i]
		if s.ID != w.id || strings.Join(s.Dependencies, ",") != w.deps || s.Prompt != w.prompt {
			t.Errorf("%s: session %d = {%s deps %v prompt %q}, want {%s deps [%s] prompt %q}", name, i, s.ID, s.Dependencies, s.Prompt, w.id, w.deps, w.prompt)
		}
	}
	return dag
}

func TestParsePlan_HarnessShapes(t *testing.T) {
	dag := checkPlanSessions(t, "decision plan", harnessDecisionPlan, []wantSession{
		{"D-001-S1", "", "# RESEARCH BRIEF: datastore landscape\nMap the credible primary datastores."},
		{"D-001-S2", "D-001-S1", "# RESEARCH BRIEF: how PostgreSQL could fail\nFind failure modes and reversal triggers."},
	})
	if dag.Plan != "D-001" || dag.IsProject() || dag.Sessions[0].Role != "L" || dag.Sessions[1].Role != "F" {
		t.Errorf("decision plan fields: plan %q, roles %q %q", dag.Plan, dag.Sessions[0].Role, dag.Sessions[1].Role)
	}

	dag = checkPlanSessions(t, "project plan", harnessPipeline, []wantSession{
		{"T1-01", "", "Research datastores. [UPSTREAM_FINDINGS]"},
		{"T1-02", "", "Research queues. [UPSTREAM_FINDINGS]"},
		{"T1-03", "", "Research auth. [UPSTREAM_FINDINGS]"},
		{"SYN-01", "T1-01,T1-02,T1-03", "Synthesize. [ALL_SESSION_FINDINGS]"},
	})
	if !dag.IsProject() || !dag.IsSynthesisID("SYN-01") || dag.SessionByID("T1-01").DecisionRef != "D-001" || dag.SessionByID("T1-01").DoorType != "one-way" {
		t.Errorf("project plan fields: %+v", dag.Sessions)
	}
}

// The plans earlier audit rounds used as repros (round 1, HIGH-07), written with the block: each parses to exactly
// its declared sessions and dependencies, with no issue, and every session has its own prompt.
func TestParsePlan_EarlierReproPlans(t *testing.T) {
	block := func(entries ...string) string {
		return "<F>yaml vivechak-plan\nplan: PROJECT\nsessions:\n  - " + strings.Join(entries, "\n  - ") + "\n<F>\n"
	}
	sessions := func(ids ...string) string {
		var b strings.Builder
		b.WriteString("## Sessions\n\n")
		for _, id := range ids {
			b.WriteString("### " + id + ": Topic " + id + "\n\n<F>prompt\nBrief for " + id + ".\n<F>\n\n")
		}
		return b.String()
	}
	three := []wantSession{{"T1-01", "", "Brief for T1-01."}, {"T1-02", "T1-01", "Brief for T1-02."}, {"SYN-01", "T1-01,T1-02", "Brief for SYN-01."}}
	entries := []string{"{id: T1-01, title: Datastore, layer: 1, door: one-way, informs: [D-001], depends: []}",
		"{id: T1-02, title: Cache, layer: 1, door: two-way, depends: [T1-01]}", "{id: SYN-01, title: Synthesis, layer: 2, kind: synthesis, depends: [ALL]}"}

	// R1: a Phase 0 Exit Gate table (B1 used to become a session depending on a phantom "B").
	r1 := md("# Research Pipeline: R1\n\n" + block(entries...) + "\n" + sessions("T1-01", "T1-02", "SYN-01") + `## Phase 0 Exit Gate Criteria

| Check | Criterion | Track | Evidence |
|---|---|---|---|
| B1 | DAG closure | B | all sessions final |
| B2 | Falsification record | B | premortem written |
`)
	checkPlanSessions(t, "R1 gate table", r1, three)

	// R2: a risk register row (R1 used to become a phantom session blocked on T1-01).
	r2 := md("# Research Pipeline: R2\n\n" + block(entries...) + `
## Risk Register

| ID | Risk | Session | Mitigation |
|---|---|---|---|
| R1 | Vendor lock-in | T1-01 | keep the data model portable |

` + sessions("T1-01", "T1-02", "SYN-01"))
	checkPlanSessions(t, "R2 risk register", r2, three)

	// R3: a six-column session table (it used to drop every dependency, so SYN-01 showed Ready first).
	r3 := md("# Research Pipeline: R3\n\n" + block(entries...) + `
## Session Matrix

| ID | Title | Layer | Door Type | Decision ID | Dependencies |
|---|---|---|---|---|---|
| T1-01 | Datastore | 1 | one-way | D-001 | none |
| T1-02 | Cache | 1 | two-way | — | T1-01 |
| SYN-01 | Synthesis | 2 | — | — | T1-01, T1-02 |

` + sessions("T1-01", "T1-02", "SYN-01"))
	dag := checkPlanSessions(t, "R3 six-column table", r3, three)
	if ready := idsOf(dag.NextSessions(map[string]bool{})); strings.Join(ready, ",") != "T1-01" {
		t.Errorf("R3: first ready sessions %v, want [T1-01]", ready)
	}

	// R4: a decision plan headed "# D-01-PLAN:" (it used to warn W-PARSER-SESSION-MISMATCH for D-01-PLAN).
	r4 := md(`# D-01-PLAN: Event storage engine

<F>yaml vivechak-plan
plan: D-01
sessions:
  - {id: D-01-S1, title: Landscape, role: L, door: one-way, informs: [D-01], depends: []}
  - {id: D-01-S2, title: Comparison, role: C, door: one-way, informs: [D-01], depends: [D-01-S1]}
<F>

` + sessions("D-01-S1", "D-01-S2"))
	checkPlanSessions(t, "R4 D-01-PLAN heading", r4, []wantSession{{"D-01-S1", "", "Brief for D-01-S1."}, {"D-01-S2", "D-01-S1", "Brief for D-01-S2."}})

	// R6 and R8: SYNC-01 without kind is ordinary; FINAL-01 with kind: synthesis is the synthesis session.
	r6 := md("# Research Pipeline: R6\n\n" + block("{id: SYNC-01, title: Sync protocol, depends: []}", "{id: FINAL-01, title: Founding architecture, kind: synthesis, depends: [ALL]}") +
		"\n" + sessions("SYNC-01", "FINAL-01"))
	dag = checkPlanSessions(t, "R6/R8 synthesis by kind", r6, []wantSession{{"SYNC-01", "", "Brief for SYNC-01."}, {"FINAL-01", "SYNC-01", "Brief for FINAL-01."}})
	if dag.IsSynthesisID("SYNC-01") || !dag.IsSynthesisID("FINAL-01") || dag.SessionByID("FINAL-01").OutputFile != FADFile {
		t.Errorf("R6/R8: SYNC-01 must be ordinary and FINAL-01 the synthesis session: %+v", dag.Sessions)
	}
}

// Every way a plan can be unusable is refused with its V-PLAN-* code and a one-sentence fix.
func TestValidatePlan_RefusesEachPlanProblem(t *testing.T) {
	good := []string{"{id: T1-01, title: Datastore, depends: []}", "{id: T1-02, title: Queue, depends: [T1-01]}", "{id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}"}
	prompts := section("T1-01", "Datastore", "Research datastores.") + section("T1-02", "Queue", "Research queues.") + section("SYN-01", "Synthesis", "Synthesize.")
	with := func(i int, entry string) []string {
		out := append([]string{}, good...)
		out[i] = entry
		return out
	}
	cases := []struct {
		name string
		plan string
		want string
	}{
		{"table-only plan", md("# Research Pipeline\n\n| ID | Topic | Dependencies | Output File |\n|---|---|---|---|\n| T1-01 | Datastore | none | sessions/T1-01.md |\n\n## Sessions\n\n") + prompts, CodePlanMissingBlock},
		{"two blocks", projectPlan(good, prompts) + md("\n<F>yaml vivechak-plan\nplan: PROJECT\nsessions: []\n<F>\n"), CodePlanInvalidBlock},
		{"unclosed block", md("# P\n\n<F>yaml vivechak-plan\nplan: PROJECT\nsessions:\n  - {id: T1-01}\n"), CodePlanInvalidBlock},
		{"invalid YAML", projectPlan(with(1, "{id: T1-02, depends: [T1-01}"), prompts), CodePlanInvalidBlock},
		{"unknown session key", projectPlan(with(1, "{id: T1-02, title: Queue, depends_on: [T1-01]}"), prompts), CodePlanInvalidBlock},
		{"unknown top-level key", strings.Replace(projectPlan(good, prompts), "plan: PROJECT", "plan: PROJECT\nschema_version: \"0.1.0\"", 1), CodePlanInvalidBlock},
		{"no plan field", strings.Replace(projectPlan(good, prompts), "plan: PROJECT\n", "", 1), CodePlanInvalidBlock},
		{"bad plan field", strings.Replace(projectPlan(good, prompts), "plan: PROJECT", "plan: Demo", 1), CodePlanInvalidBlock},
		{"bad door", projectPlan(with(0, "{id: T1-01, title: Datastore, door: oneway, depends: []}"), prompts), CodePlanInvalidBlock},
		{"bad kind", projectPlan(with(0, "{id: T1-01, title: Datastore, kind: research, depends: []}"), prompts), CodePlanInvalidBlock},
		{"bad role", projectPlan(with(0, "{id: T1-01, title: Datastore, role: X, depends: []}"), prompts), CodePlanInvalidBlock},
		{"layer not a number", projectPlan(with(0, "{id: T1-01, title: Datastore, layer: L1, depends: []}"), prompts), CodePlanInvalidBlock},
		{"depends not a list", projectPlan(with(1, "{id: T1-02, title: Queue, depends: T1-01}"), prompts), CodePlanInvalidBlock},
		{"session without id", projectPlan(with(1, "{title: Queue, depends: [T1-01]}"), prompts), CodePlanInvalidBlock},
		{"duplicate session", projectPlan(append(append([]string{}, good...), "{id: T1-01, title: Again}"), prompts), CodePlanInvalidBlock},
		{"no sessions", md("# P\n\n<F>yaml vivechak-plan\nplan: PROJECT\nsessions: []\n<F>\n\n## Sessions\n"), CodePlanNoSessions},
		{"sessions key missing", md("# P\n\n<F>yaml vivechak-plan\nplan: D-001\n<F>\n"), CodePlanNoSessions},
		{"no ## Sessions section", strings.Replace(projectPlan(good, prompts), "## Sessions", "## Session Prompts", 1), CodePlanMissingPrompt},
		{"session without a heading", projectPlan(good, section("T1-01", "Datastore", "a")+section("SYN-01", "Synthesis", "s")), CodePlanMissingPrompt},
		{"heading without a prompt fence", projectPlan(good, section("T1-01", "Datastore", "a")+"### T1-02: Queue\n\nResearch queues.\n\n"+section("SYN-01", "Synthesis", "s")), CodePlanMissingPrompt},
		{"empty prompt fence", projectPlan(good, section("T1-01", "Datastore", "a")+section("T1-02", "Queue", "  ")+section("SYN-01", "Synthesis", "s")), CodePlanMissingPrompt},
		{"two prompt fences", projectPlan(good, section("T1-01", "Datastore", "a")+md("### T1-02: Queue\n\n<F>prompt\nb\n<F>\n\n<F>prompt\nc\n<F>\n\n")+section("SYN-01", "Synthesis", "s")), CodePlanDuplicatePrompt},
		{"two headings for a session", projectPlan(good, prompts+section("T1-02", "Queue again", "b")), CodePlanDuplicatePrompt},
		{"prompt for an unlisted session", projectPlan(good, prompts+section("T9-99", "Forgotten", "x")), CodePlanUnlistedSession},
		{"unknown dependency", projectPlan(with(1, "{id: T1-02, title: Queue, depends: [T9-99]}"), prompts), CodePlanUnknownDependency},
		{"self dependency", projectPlan(with(1, "{id: T1-02, title: Queue, depends: [T1-02]}"), prompts), CodePlanCycle},
		{"cycle", projectPlan(with(0, "{id: T1-01, title: Datastore, depends: [T1-02]}"), prompts), CodePlanCycle},
		{"depends on the synthesis session", projectPlan(with(1, "{id: T1-02, title: Queue, depends: [SYN-01]}"), prompts), CodePlanCycle},
		{"project plan without synthesis", projectPlan(with(2, "{id: SYN-01, title: Synthesis, depends: [ALL]}"), prompts), CodePlanSynthesis},
		{"project plan with two synthesis sessions", projectPlan(with(1, "{id: T1-02, title: Queue, kind: synthesis, depends: [T1-01]}"), prompts), CodePlanSynthesis},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := ValidatePlan([]byte(tc.plan))
			if !res.HasBlocking() || res.BlockingIssues()[0].Code != tc.want {
				t.Fatalf("expected first issue %s, got %+v", tc.want, res.Issues)
			}
			for _, issue := range res.Issues {
				if issue.Level != L2Block || !strings.HasPrefix(issue.Code, "V-") {
					t.Errorf("plan problems are blocking V- codes, got %+v", issue)
				}
				if fix := issue.FixHint; fix == "" || strings.Contains(fix, "\n") || !strings.HasSuffix(fix, ".") {
					t.Errorf("%s needs a one-sentence fix, got %q", issue.Code, fix)
				}
			}
			if _, err := ParsePipeline([]byte(tc.plan)); tc.want != CodePlanSynthesis && (err == nil || PlanFix(err) == "") {
				t.Errorf("ParsePipeline must fail with a fix for %s, got %v", tc.want, err)
			}
		})
	}

	// The valid plan the cases start from has no issue at all.
	if res := ValidatePlan([]byte(projectPlan(good, prompts))); len(res.Issues) != 0 {
		t.Fatalf("base plan: %+v", res.Issues)
	}
	// An unknown key the planner commonly writes names the key to use.
	res := ValidatePlan([]byte(projectPlan(with(1, "{id: T1-02, title: Queue, dependencies: [T1-01]}"), prompts)))
	if len(res.Issues) != 1 || res.Issues[0].FixHint != "Write depends instead of dependencies." {
		t.Fatalf("expected the depends hint, got %+v", res.Issues)
	}
	// The scope a plan is saved as must agree with its block.
	if res := ValidatePlanForScope([]byte(harnessDecisionPlan), "project"); !res.HasBlocking() || res.Issues[0].Code != CodePlanInvalidBlock {
		t.Fatalf("a decision plan saved as the project pipeline: %+v", res.Issues)
	}
	if res := ValidatePlanForScope([]byte(harnessPipeline), "decision"); !res.HasBlocking() || res.Issues[0].Code != CodePlanInvalidBlock {
		t.Fatalf("the project pipeline saved as a decision plan: %+v", res.Issues)
	}
	if res := ValidatePlanForScope([]byte(harnessDecisionPlan), "decision"); len(res.Issues) != 0 {
		t.Fatalf("a decision plan needs no synthesis session: %+v", res.Issues)
	}
}

// research/FAD.md completes the plan's kind: synthesis session whatever id it carries, and only that session.
func TestPlanCompletedSessionsFADCompletesSynthesis(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, SessionsDir), 0o755); err != nil {
		t.Fatal(err)
	}
	ws, err := store.OpenWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	plan := projectPlan([]string{"{id: SYN-01, title: Syndication}", "{id: FINAL-01, kind: synthesis, depends: [ALL]}"},
		section("SYN-01", "Syndication", "a")+section("FINAL-01", "Founding architecture", "s"))
	dag, err := ParsePipeline([]byte(plan))
	if err != nil {
		t.Fatal(err)
	}

	write := func(content string) {
		if err := os.WriteFile(filepath.Join(dir, FADFile), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("---\nid: FAD\nstatus: draft\n---\n# FAD\n")
	if done, _ := dag.CompletedSessions(ws); done["FINAL-01"] {
		t.Fatal("a draft FAD must not complete the synthesis session")
	}
	write("---\nid: FAD\nstatus: complete\n---\n# FAD\n")
	done, err := dag.CompletedSessions(ws)
	if err != nil {
		t.Fatal(err)
	}
	if !done["FINAL-01"] || done["SYN-01"] {
		t.Fatalf("FAD must complete FINAL-01 (kind: synthesis) and not SYN-01 (ordinary): %v", done)
	}
}
