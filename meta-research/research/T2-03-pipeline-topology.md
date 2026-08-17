# Pipeline Topology for Multi-Session Research: Sequential, DAG, Iterative, or Adaptive

**Ticket:** T2-03
**Scope:** Systems architecture & workflow-optimization analysis of the research pipeline framework
**Subject:** Should the 3-tier sequential structure (Landscape → Architecture → Blueprints → SYN-01) be replaced, and with what?

---

## Executive Summary

The current framework enforces a rigid 3-tier sequential structure in which every session in Tier *N+1* waits for every session in Tier *N* to finish, regardless of whether it actually needs anything Tier *N* produced. The criticism in the brief is correct on the facts: most Tier 2 sessions depend on zero or one specific Tier 1 sessions, not all five. But the fix isn't to remove structure — it's to replace **stage-gating** with **dependency-gating**.

This document evaluates four topologies — rigid sequential, dependency graph (DAG), iterative/spiral, and fully adaptive — against a concrete example pipeline, and closes with a recommendation for a **constrained DAG with bounded adaptive checkpoints**. The Tier 1/2/3/SYN-01 labels survive as an organizing taxonomy; they simply stop functioning as execution gates.

---

## The Core Distinction: Information Dependency vs. Logical Dependency

Everything below rests on separating two things the current framework treats as one: *needing* a prior session's output versus merely *benefiting* from its context.

Project management already has a mature vocabulary for exactly this split. The standard PMBOK taxonomy of schedule dependencies distinguishes **mandatory dependencies** ("hard logic") — relationships inherent to the nature of the work itself, where the order isn't really a choice — from **discretionary dependencies** ("soft logic" or "preferred logic") — orderings that reflect best practice, domain experience, or team preference but aren't physically required. A mandatory dependency exists because one activity literally cannot begin until another produces what it needs (you can't crash-test a prototype that hasn't been built yet); a discretionary dependency exists because a team has learned that one order *tends to work better*, even though the reverse order is possible.

Mapped onto research sessions:

- **Information dependency (hard):** Session B cannot produce a defensible output without a specific fact, decision, or artifact that only Session A generates. Example: a "Component Architecture Blueprint" session cannot meaningfully exist before a framework has been chosen — there is nothing yet to blueprint. This is a genuine blocking relationship.
- **Logical dependency (soft):** Session B's output would be *better* — more consistent in terminology, less likely to duplicate effort, aware of constraints already surfaced — if it had Session A's context. But B can still run on its own, produce a valid and defensible answer, and be reconciled with A afterward. This is not a blocking relationship; it's a context-sharing opportunity.

The rigid 3-tier model encodes nearly everything as if it were a hard dependency, because a stage gate has no way to express "would help, but isn't required." That is the entire mechanism behind the problem described in the brief: Frontend Framework Selection is stuck behind Competitor Landscape not because a real information dependency exists between them, but because a linear model has no cheaper way to say "these are unrelated" than to make everything wait its turn. A topology that represents both dependency types separately — and routes each to the mechanism suited to it — removes the false wait without discarding the ordering constraints that are genuinely real.

---

## Four Topologies, Evaluated

### 1. Rigid Sequential Staging (the current model)

**How it works:** Everything in Stage *N+1* waits for all of Stage *N*. One linear chain from Landscape through Synthesis.

**Strengths**
- Trivial to reason about, schedule, and audit — there is exactly one valid execution order, and exactly one story of how the pipeline arrived at its conclusions.
- Guarantees every session has the maximum possible context, since every prior stage is complete before it starts — nothing is written in ignorance of something already known.
- Zero dependency-mapping investment up front: you never have to correctly identify which sessions need which others, because everything waits for everything anyway.
- Simple failure containment — a bad Tier 1 finding is caught before Tier 2 spends any effort building on it.

**Weaknesses**
- Every session ends up on the critical path. In project-scheduling terms, the model gives every task zero float, even sessions with no real downstream dependents.
- Total wall-clock time is the *sum* of every stage's duration rather than the *longest genuine dependency chain*. A five-session Tier 1 and a six-session Tier 2 cost eleven sessions' worth of time even when only a couple of those eleven relationships are real.
- Change is expensive: a late Tier 1 revision invalidates work in every later tier, whether or not that later work actually used the revised finding.
- It manufactures the exact complaint in the brief — sessions with zero real dependency on the prior tier still wait, because the model has no way to express "unrelated."

**Best fit:** Small pipelines (roughly under 6–8 sessions total, where mapping a graph would cost more than it saves); genuinely linear problems where each step truly does need the last one's output; contexts where auditability and a single, uncontested reasoning chain matter more than speed.

---

### 2. Dependency Graph (DAG)

**How it works:** Each session is a node; edges represent only real information dependencies. Sessions with no incoming edges — or whose edges are already satisfied — run as soon as capacity allows. Execution order falls out of the graph rather than being declared up front.

This isn't a novel idea for orchestrating interdependent work; it's the default model wherever the underlying problem looks like this. Apache Airflow defines a workflow as literally a directed acyclic graph of tasks — tasks with no dependency between them run in parallel, and only tasks with an explicit upstream/downstream relationship run in sequence. Build systems take the same idea further: Bazel computes a project's dependency graph from declared inputs, executes independent targets in parallel, and on a re-run rebuilds only the targets whose actual dependencies changed, rather than everything downstream of a stage boundary.

**Strengths**
- Total time collapses to the *critical path* — the longest real dependency chain — instead of the sum of every session. Sessions off that path have slack and can run whenever capacity allows.
- Correctness is preserved exactly where it matters: a session with a genuine information dependency still waits for it.
- Mirrors tooling that already exists for precisely this class of problem (Airflow, Dagster, and Prefect for data pipelines; Bazel and Make for builds), so the pattern is well understood, and the graph itself doubles as the audit trail.
- Newer "asset-based" orchestrators such as Dagster reframe the same idea declaratively: instead of specifying steps and order, a team declares the deliverables it wants and what each depends on, and the orchestrator derives the plan. That maps unusually well onto research sessions — each one *is* a deliverable ("Database Technology Decision") with named inputs, not merely a task in a queue.

**Weaknesses**
- All the correctness lives in the edges. Under-specify a dependency and two sessions can silently contradict each other. Over-specify — add an edge "just to be safe" — and the sequential model's worst property quietly comes back; practitioners have documented real slowdowns from exactly this kind of over-cautious dependency declaration in production build graphs.
- The graph is normally fixed at design time. It handles "we didn't know we'd need this session" poorly unless paired with a mechanism to revise it mid-flight.
- A pure DAG only encodes hard dependencies. Stop there, and whatever benefit soft/logical dependencies were providing — shared vocabulary, already-surfaced constraints — disappears along with the false waits. Parallelism goes up, but so does the risk of duplicated effort or inconsistent assumptions, unless something else fills that gap.

**Best fit:** Pipelines where the *shape* of the problem is knowable up front — the sessions can be named and reasoned about, even if their specific findings aren't known yet. This describes a software-architecture research pipeline almost exactly.

---

### 3. Iterative / Spiral

**How it works:** Repeated passes over the whole problem, each pass going deeper, rather than one linear pass through fixed stages. The canonical version is Barry Boehm's spiral model: each loop moves through four quadrants — determine objectives and alternatives, evaluate the alternatives and resolve the highest risks, build and verify the next increment, then plan the following loop — repeating until risk is acceptably low. Agile/Scrum applies the same underlying logic at a faster cadence: fixed-length sprints, a backlog re-prioritized every cycle, and a retrospective that directly shapes what the next sprint covers.

**Strengths**
- Naturally suited to problems where the full topic list isn't known yet — a dependency graph can't be drawn for research topics that haven't been discovered, but another loop can always be run.
- Risk-first ordering: the spiral model explicitly tackles the riskiest unknowns earliest, when a wrong assumption is cheapest to correct. This is part of why NASA used it for Space Shuttle and Earth-observation software.
- Built-in course correction — every loop ends with a planning step, so a bad early finding doesn't silently propagate through several tiers before anyone revisits it.
- Low upfront modeling cost: unlike a DAG, there's no need to correctly enumerate every dependency before starting.

**Weaknesses**
- Overhead scales with the number of loops, which makes it comparatively slow once a problem is already well understood — re-running "determine objectives" is wasted motion when objectives are already stable.
- Iteration by itself isn't a parallelism strategy. A spiral says *when* to reassess; it doesn't say which sessions within a given loop could have run simultaneously — that has to be layered on top.
- Needs an explicit stopping rule, or it risks looping indefinitely without a clear signal that risk is actually resolved.
- Coordination cost compounds with each loop — more loops mean more replanning and synthesis overhead.

**Best fit:** Genuinely exploratory research where the topic list itself is uncertain at the outset, fast-moving domains where a single upfront plan would be stale before it finished, or situations where early stakeholder feedback needs to reshape scope before more effort is committed.

---

### 4. Fully Adaptive

**How it works:** No topology is fixed up front. A controller inspects results as they land and decides in real time what runs next — spinning up new sessions where a finding revealed unexpected complexity, and closing or shrinking planned sessions where a topic turned out to be simple or already resolved elsewhere.

It's worth being precise about what this actually is: not a fourth independent structure, but a DAG whose node set and edges can mutate at runtime, governed by an iteration-style feedback loop. Modern orchestration tooling already builds this on top of a DAG rather than replacing it. Airflow's dynamic task mapping generates a variable number of parallel tasks at runtime from an upstream task's output, without the pipeline's author knowing that count in advance. Kubeflow Pipelines supports conditional branches and parallel loops driven by runtime data in the same spirit. A recent industry retrospective on ML-pipeline orchestration is blunt about the boundary here: DAG-based tools handle "the graph is known at compile time" workloads well, but open-ended, exploratory work needs execution paths that genuinely aren't knowable upfront — and that is a materially harder orchestration problem, not a free upgrade.

**Strengths**
- In principle, optimal use of research capacity: deep investment exactly where complexity turns out to be real, minimal investment where it doesn't.
- Responds to *actual* complexity discovered mid-pipeline, not complexity assumed at design time.
- Can make a simple project finish faster and a hard project get proportionally more scrutiny, automatically, without a human re-planning the whole pipeline.

**Weaknesses**
- Highest orchestration burden of the four options: something has to make good expand/contract judgment calls, and that something is unproven until it has a track record.
- Hardest to audit or reproduce — the path taken depends on the order findings arrived in, not only on what was found.
- Real risk of instability without guardrails: unbounded expansion (complexity justifying more sessions, which reveal more apparent complexity) or premature contraction (declaring a topic simple before its real complexity has actually surfaced).
- A cold-start problem: a controller needs some initial signal to react to, which means it can't safely run from session one with no groundwork already laid.

**Best fit:** Mature, frequently repeated pipeline templates with enough run history to calibrate what "unexpected complexity" looks like in that domain, and high-variance, high-value projects where getting the scope wrong in either direction is expensive enough to justify the extra orchestration cost.

---

## Mapping Real Dependencies: A Software Architecture Research Pipeline

Take a plausible session list for exactly the kind of pipeline described in the brief, and separate real (hard) dependencies from soft (logical) ones and from no dependency at all.

```
LAYER 0 — Tier 1 (sources: fully parallel, no dependencies on each other)
  Competitor & Market Landscape
  Technology & Tooling Trends
  Target User / Audience Research
  Regulatory & Data-Residency Scan
  Existing / Legacy Systems Audit

LAYER 1 — Tier 2 (0–2 real dependencies each — NOT "all of Tier 1")
  Frontend Framework Selection      <- (soft)  Target User Research
  API Design Paradigm               <- (none)
  Backend Framework Selection       <- (hard, conditional)  Legacy Systems Audit
  Database Technology Selection     <- (hard, conditional)  Legacy Systems Audit
  Cloud / Hosting Decision          <- (hard)  Regulatory & Data-Residency Scan
  Auth & Identity Strategy          <- (hard)  Regulatory & Data-Residency Scan

LAYER 2 — Tier 3 (real dependencies concentrate here)
  Component Architecture Blueprint  <- Frontend Framework + Backend Framework + API Paradigm
  Data Model / Schema Design        <- Database Technology Selection
  CI/CD Pipeline Design             <- Cloud/Hosting Decision + Backend Framework
  Security Architecture             <- Auth & Identity Strategy + Cloud/Hosting Decision
  Testing Strategy                  <- Component Architecture Blueprint + API Paradigm

LAYER 3 — Sink
  SYN-01 Grand Synthesis            <- everything above
                                        (partial synthesis can begin as branches close)
```

Tier 1 is already close to a perfect DAG source layer — nothing there should ever have waited on anything else in Tier 1, and the rigid model gets this layer right, essentially by accident, since there's only one stage boundary to enforce.

The interesting layer is Tier 2, laid out in full below:

| Session | Real (hard) dependency | Soft (logical) dependency | Can start immediately? |
|---|---|---|---|
| Frontend Framework Selection | *none* | Target User Research (mobile-heavy vs. desktop-heavy audience) | **Yes** |
| API Design Paradigm | *none* | *none* | **Yes** |
| Backend Framework Selection | Legacy Systems Audit — only if integrating with an existing backend | — | Conditional (yes if greenfield) |
| Database Technology Selection | Legacy Systems Audit — migration constraints, if any | Target User Research (scale expectations) | Conditional |
| Cloud / Hosting Decision | Regulatory & Data-Residency Scan | — | **No** — genuinely blocked |
| Auth & Identity Strategy | Regulatory & Data-Residency Scan | — | **No** — genuinely blocked |

Note the shape: none of these six sessions depends on *all* of Tier 1, and two of them (Frontend Framework Selection, API Design Paradigm) depend on *none* of it — the exact case named in the brief. That's the norm here, not the exception, once dependencies are checked individually instead of assumed at the tier level.

Tier 3 is where the rigid model's instinct is closest to correct. Blueprints genuinely can't exist before the decisions they operationalize — component architecture can't be designed for a framework nobody has picked. Real hard dependencies concentrate in Tier 3, which is exactly why the right fix is a DAG rather than a free-for-all: the goal is to remove *false* dependencies, not all dependencies.

SYN-01 genuinely depends on nearly everything — a true sink node. Even there, partial synthesis of a branch that has already closed (say, everything under the data layer) can begin before every other branch finishes, rather than waiting on 100% completion of every leaf session.

**Bottom line:** roughly two-thirds of the Tier 2 sessions above have at most one real Tier 1 dependency, not five. The rigid model isn't wrong about Tier 3 — real dependencies do concentrate there — it's wrong to treat the Tier 1→2 and Tier 2→3 boundaries as uniformly blocking when the actual graph underneath them is sparse.

---

## When Parallel Execution Helps — and When It Hurts

**Helps when:**

- **The topics are genuinely independent.** No shared state to corrupt, nothing to contradict.
- **Freshness or lack of bias is itself the goal.** This is why dual, independent screening is standard practice in systematic reviews: two or three reviewers screen the same records separately, specifically to control selection bias, and reconcile disagreements only afterward. The same logic is why a design sprint has each participant sketch solutions alone before anyone sees anyone else's idea — independent generation followed by reconciliation tends to produce different, and usually better, results than generation with visibility, because an early idea anchors everyone who hears it and groups tend to converge on the first plausible answer rather than the best one. (Structured "silent generation" techniques in group-decision research have been associated with meaningfully more — and more original — ideas than open group discussion.)
- **Time compression matters and the critical path allows it.** Running independent sessions concurrently reduces wall-clock time without touching correctness.
- **A slow or failed session shouldn't block unrelated work.** Parallel branches contain failure; sequential ones propagate it.

**Hurts when:**

- **A shared foundation is genuinely unresolved.** If two sessions need the same unstated assumption — a budget ceiling, a non-negotiable constraint — and neither has it yet, running them in parallel doesn't save time; it produces two outputs that need reconciling later, and reconciliation is often costlier than the wait would have been.
- **Effort duplicates unknowingly.** Two sessions covering overlapping ground without visibility into each other waste research capacity that sequencing — or simply informing each session what else is in flight — would have avoided.
- **A dependency was undercounted.** If something assumed to be "logical" turns out to be closer to a hard dependency, the parallel session has to be partly redone once the real prerequisite lands, and that rework can exceed whatever time parallelism saved.
- **Narrative coherence is part of the deliverable's quality, not just its correctness.** A sequential chain doesn't only avoid factual contradictions — accumulated context often makes later sessions sharper: consistent terminology, awareness of what's already been ruled out. A naively parallel model can lose that unless it's deliberately reintroduced.

**The reconciling move:** hard dependencies become DAG edges — blocking, because they must be. Soft dependencies become a **shared context brief** — constraints, glossary, decisions already locked — injected into every session's starting context regardless of its position in the graph. That captures the benefit of accumulated context without turning it into a blocking gate, and it directly answers the "duplicated effort / inconsistent assumptions" risk without sacrificing the parallelism a correctly scoped DAG makes available.

---

## Can a Pipeline Expand and Contract Its Own Scope?

Yes — but treat it as an optional layer added *to* a DAG, activated at defined checkpoints, not as the base architecture. Precedent for both directions already exists.

**Expansion.** Airflow's dynamic task mapping is the clean version of this: a task can generate an arbitrary number of parallel child tasks at runtime based on what an upstream task actually returned, without the pipeline's author having to guess that number in advance. Translated to research sessions: a Tier 2 session that surfaces unexpected complexity — say, three genuinely different viable database candidates rather than one obvious choice — can spawn additional focused sub-sessions instead of either cramming all three into the original scope or silently under-covering the topic.

**Contraction.** The "rapid review" is the research-methodology version of scope-shrinking: it follows the same core process as a full systematic review but deliberately streamlines steps — a narrower search, lighter screening — when speed matters more than exhaustive coverage, while staying transparent about what was skipped. Applied here: if a planned Tier 2 session turns out to already be answered by something Tier 1 found — a "Cloud/Hosting Decision" session discovers the Compliance Scan already settled the question — that session should be allowed to close early rather than running to its originally planned depth out of habit.

**Guardrails this needs to actually work:**

- A stated rubric for what counts as "unexpected complexity," or every session will find a reason to expand.
- A cap on expansion depth per checkpoint — no more than *N* child sessions spawned from one node without an explicit review.
- Checkpoints at defined moments (end of a tier-batch, or when a session is flagged high-uncertainty) rather than continuous re-evaluation after every single session. Continuous adaptation is exactly the failure mode a recent review of ML-pipeline orchestration flags when comparing static DAGs to open-ended agent-style workflows: the latter genuinely needs different tooling because its shape isn't known upfront, and that is a harder, less auditable problem than DAG orchestration — not a strictly better one.
- A human or synthesis-step review before a scope change actually executes, at least until the pipeline has enough run history to calibrate automated judgment.

This is why Topology 4 should stay a checkpoint-triggered control layer rather than the default engine: unconstrained, continuous adaptivity is the hardest of the four options to audit, and both precedents above bound it deliberately — living and rapid reviews define *when* re-evaluation happens rather than re-evaluating after every record, and dynamic task mapping expands a known task type across a discovered input set rather than inventing new task types on the fly.

---

## What Established Methodologies Teach Us

Cross-cutting lessons, rather than a repeat of the topology write-ups above:

**Even the most sequential-looking processes contain a deliberate parallel step, and it's not an accident.** The GV design sprint is famously rigid — one phase per day, no reordering — yet Tuesday's sketching is done by each participant alone, specifically to avoid the anchoring and groupthink that group sketching would introduce, before Wednesday's convergence. Sequential *between* phases and parallel *within* a phase aren't in tension; the current framework's mistake is applying only the first pattern and never the second.

**Structure and adaptability aren't opposites — the question is where the guardrail sits.** Scrum re-plans every sprint through backlog grooming: bounded adaptivity, reassessed on a fixed cadence. Kanban re-prioritizes continuously but caps how much can be in flight via WIP limits: bounded adaptivity, reassessed anytime but only within capacity. Neither is "no structure" — both bound *when* or *how much* can change, which is the same move recommended below for adaptive checkpoints.

**The double diamond justifies parallel discovery and sequential decision as two different modes of one pipeline, not a single choice.** Diverge phases (Discover, Develop) are explicitly where broad, low-commitment exploration belongs; converge phases (Define, Deliver) are where narrowing and commitment belong. That maps directly onto Tier 1 (diverge — stay parallel) and Tier 3 (converge — accept more real sequencing), with Tier 2 as the transition, where dependency-mapping — not blanket sequencing — decides which sessions are still diverging and which are already converging.

**Rigor and speed are a dial, not a binary, and the research-methodology world has already built the dial.** A full systematic review, a rapid review, and a living review all use the same underlying method at different levels of streamlining and update frequency, chosen deliberately based on how much the answer is worth and how fast it's needed. A pipeline framework should offer the equivalent per-session choice rather than applying one fixed depth to every topic regardless of how settled or contested it already is.

**Risk-first ordering beats topic-first ordering when uncertainty is the dominant cost.** Boehm's spiral model doesn't ask "what comes next in the natural order of things" — it asks "what's the riskiest unresolved assumption" and tackles that first, every loop. For a research pipeline, this suggests sequencing — within whatever slack the DAG allows — by how much a wrong answer would cost downstream, not only by topical adjacency.

---

## Comparative Summary

| | Rigid Sequential | DAG | Iterative / Spiral | Fully Adaptive |
|---|---|---|---|---|
| **Speed (typical)** | Slowest — sum of all stages | Fast — critical path only | Moderate — overhead per loop | Fastest in principle, unpredictable in practice |
| **Correctness guarantee** | Highest (max context always) | High (real dependencies enforced) | Moderate (risk-driven, not dependency-driven) | Depends entirely on controller quality |
| **Handles unknown scope** | Poorly | Poorly (graph fixed at design time) | Well | Best, by design |
| **Upfront modeling cost** | None | Moderate — dependencies must be mapped | Low | High — needs a calibrated controller |
| **Auditability / reproducibility** | Highest | High — the graph is the audit trail | Moderate | Lowest |
| **Best-fit scale** | Small pipelines | Known-shape, medium–large pipelines | Exploratory or ambiguous scope | Mature, high-variance pipelines only |

---

## Recommendation: A Constrained DAG With Adaptive Checkpoints

Replace stage-gating with dependency-gating. Concretely:

1. **Map real dependencies once, per pipeline template.** Every session gets an explicit, named list of hard (information) dependencies. No session inherits a wait just because of which tier it's labeled — the tier becomes metadata, not a gate.
2. **Invert the default.** Today a session is sequential unless someone proves otherwise. Flip it: every new session defaults to *no dependency* — eligible to start immediately — unless a specific, named, genuine information dependency is identified during pipeline design. This is the direct fix for the Frontend-Framework-vs-Competitor-Landscape problem: the burden of proof moves to whoever wants to add a blocking edge, not to whoever wants to remove one.
3. **Handle soft dependencies with a shared context brief, not an edge.** A short, living document — locked decisions, constraints, glossary — gets injected into every session's starting context regardless of its graph position. This is what prevents more parallelism from turning into more contradictions.
4. **Add bounded adaptive checkpoints, not continuous adaptivity.** At defined points — end of a tier-batch, or when a session is flagged high-uncertainty — a short triage step can spawn additional sessions on a branch that revealed real complexity, or close a session early if it's already answered. Cap expansions per checkpoint, require a one-line justification for any scope change, and keep a human or synthesis-step review in the loop until the pipeline has enough history to trust an automated version.
5. **Keep SYN-01 as a true sink, but let synthesis start incrementally** on branches that have already closed, instead of blocking on 100% completion of every leaf session.
6. **Reserve unconstrained "fully adaptive" mode for mature, repeated templates only.** It's the highest-value option once there's calibration data on what "unexpected complexity" actually looks like in this domain, and the least trustworthy option before that data exists.

**What changes in practice, using the example above:** Tier 1 stays exactly as parallel as it already is. Frontend Framework Selection and API Design Paradigm start immediately, alongside Tier 1, because neither has a real dependency on anything in it. Cloud/Hosting Decision and Auth & Identity Strategy wait — correctly — on the Compliance Scan, because that dependency is real. Tier 3 sessions get built one by one as their specific Tier 2 prerequisites land, rather than waiting for all of Tier 2 to close. SYN-01 begins assembling whichever branches finish first. Nothing here requires abandoning structure. It requires making the structure answer to the actual shape of the work, rather than the shape of the tier list.

---

## Sources & Further Reading

- [PMTI — Mandatory vs. Discretionary Dependencies](https://www.4pmti.com/learn/mandatory-vs-discretionary-dependencies/)
- [Kanban Zone — Project Dependencies and How to Deal With Them](https://kanbanzone.com/2019/project-dependencies-dealing-with-them-efficiently/)
- [Macquarie University — PRISMA Flow Diagram & Screening](https://libguides.mq.edu.au/systematic_reviews/prisma_screen)
- [AJE — How to Create an Effective PRISMA Flow Diagram](https://www.aje.com/arc/how-to-create-prisma-flow-diagram)
- [UXPin — Double Diamond Design Process Explained](https://www.uxpin.com/studio/blog/double-diamond-design-process/)
- [GV — The Design Sprint (official 5-day guide)](https://www.gv.com/sprint/)
- [Wikipedia — Spiral Model](https://en.wikipedia.org/wiki/Spiral_model)
- [TeachingAgile — Spiral Model: Phases, Advantages & Disadvantages](https://teachingagile.com/sdlc/models/spiral)
- [Apache Airflow — Architecture Overview (DAG concepts)](https://airflow.apache.org/docs/apache-airflow/stable/core-concepts/overview.html)
- [Apache Airflow — Dynamic Task Mapping](https://airflow.apache.org/docs/apache-airflow/stable/authoring-and-scheduling/dynamic-task-mapping.html)
- [Wrike — The Critical Path Method in Project Management](https://www.wrike.com/blog/critical-path-is-easy-as-123/)
- [ScienceDirect — Living Systematic Review: the why, what, when, and how](https://www.sciencedirect.com/science/article/abs/pii/S0895435617306364)
- [Atlassian — Kanban vs. Scrum](https://www.atlassian.com/agile/kanban/kanban-vs-scrum)
- [Bazel — Dependencies](https://bazel.build/concepts/dependencies)
- [brentley.dev — Closing the Gap on Bazel's iOS Incremental Compilation](https://brentley.dev/xcodebuild-vs-bazel-incremental/)
- [Kubeflow — Control Flow in Pipelines](https://www.kubeflow.org/docs/components/pipelines/user-guides/core-functions/control-flow/)
- [ZenML — From Pipelines to Agents: How Orchestration Is Being Rewritten](https://www.zenml.io/blog/from-pipelines-to-agents)
- [Covidence — Systematic Reviews vs. Rapid Reviews](https://www.covidence.org/blog/the-difference-between-systematic-reviews-and-rapid-reviews/)
- [Dagster — What Is a Software-Defined Asset](https://dagster.io/glossary/software-defined-assets)
- [Medium (S. Gioia) — The New Brainstorming: Six Principles to Redeem Group Ideation](https://medium.com/@stephgioia/the-new-brainstorming-45a9018fc4b3)
