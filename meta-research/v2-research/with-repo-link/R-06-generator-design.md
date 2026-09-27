---
id: R-06
title: "Decision-Level and Comparison-Level Generator Design"
date: 2026-09-23
status: draft
topic: generator-design
tags: [generator, multi-scope, decision-level, comparison-level, weighted-evaluation, mcp]
informs_decisions: [D-002]
---

## 1. Research Question

GENERATOR.md already contains a complete, self-sufficient per-decision routing and
generation mechanism (Step 3: door classification → 2×2 lane routing → session
composition) nested inside a project-wide pipeline. Can that mechanism be extracted
and re-packaged as two standalone, self-contained prompts — one for a single decision,
one for a bounded comparison — without diluting the methodology invariants that make
Vivechak's output usable at project scope (evidence grading, the Bounded Exploration
Mandate, structured falsification, the Weighted Evaluation Protocol)? And does the
resulting design confirm, refine, or contradict the three-level scope model (project /
decision / comparison) this brief assumes?

**A note on sourcing.** This brief's context block asks for "key findings" from a
document named `R-03 (multi-scope methodology)`. No file matching that name exists
anywhere in the `vivechak` repository as of this session `(Grade A · corroborated ·
fresh · direct | cached)` — confirmed by an exhaustive filename and full-text search.
The closest available prior research is `T3-01-generator-architecture.md` (generator
architecture options) and `T2-04-adaptive-scaling.md` (the complexity-scoring model),
which the repo's own decision registry (`meta-research/DECISIONS.md`) has already
crystallized into two **accepted** ADRs: **D-004** (Risk-Calibrated 4-Tier Scaling
Model and Decision Routing Matrix) and **D-010** (5-Layer Hybrid Deterministic/AI
Generator Architecture). This document treats D-004 and D-010 as the effective
"R-03 equivalent" and cites them directly rather than inventing findings for a
document that doesn't exist. See **Open Question OQ-1**.

## 2. Key Findings

### 2.1 The three-level model isn't wrong — it's already latent in the existing system

GENERATOR.md's Step 3 is not merely "similar to" a per-decision generator; it **is**
one, running inside a loop over every architectural decision a project needs. Its
door-type classification, its 2×2 routing matrix (`Known Pattern × Unknown/Novel` by
`Reversible × Irreversible`), and its four lane names — `SKIP`, `FAST SPIKE`,
`CONFIRM & COMMIT`, `DEEP RESEARCH` — are the exact mechanism this brief asked for at
decision scope `(Grade A · single · fresh · direct | cached)`, read directly from
`GENERATOR.md` lines 92–100. D-004 confirms this was a deliberate, evidence-based
design choice, not an accident: project complexity is scored on 8 dimensions to set a
*session-count budget*, but **individual decisions are routed separately**, via "the
2×2 Reversibility × Familiarity Matrix," with "a hard override for high regulatory
exposure" `(Grade A · corroborated · fresh · direct | cached, D-004)`.

So the three levels are not three different methodologies — they are **the same
routing logic invoked at three different entry points**, matching D-010's own framing
of Vivechak as a layered system (deterministic scaffolding + scoped AI prose +
validation gates) rather than one monolithic prompt `(Grade A · single · fresh ·
direct | cached, D-010)`. GENERATOR-DECISION.md is Step 3 promoted from an inner loop
to the entry point. GENERATOR-COMPARISON.md is the `C` (Comparison) role from that same
step, stripped of routing entirely, for the cases where routing has already happened
elsewhere (a human already knows this is worth researching; they just want the
comparison). **The three-level model holds; this research operationalizes it rather
than replacing it.**

### 2.2 Methodology invariants — what must carry to every scope

Reading GENERATOR.md's own generator-prompt instructions byte-for-byte (not
paraphrased) against FRAMEWORK.md's principle definitions identifies exactly what is
load-bearing at every scope:

| Invariant | Source | Carries how |
|---|---|---|
| **Bounded Exploration Mandate** (stated scope is a coverage floor, not a ceiling; investigate and justify beyond it) | P4, `GENERATOR.md` L166-170 | Verbatim, byte-identical instruction text in both new generators |
| **Evidence grading**: A-E base grade + corroboration/recency/directness modifiers + verification method, `recalled` capped at D | §5.1-5.4, `GENERATOR.md` L179-186 | Verbatim, byte-identical instruction text in both new generators |
| **5-block prompt anatomy**: BRIEF → SCOPE → APPROACH → DELIVERABLE → FORMAT | §4.1 | Structurally identical section order in every emitted session prompt |
| **Audience framing, not persona assignment** — "a principal architect needing production-grade tradeoffs" sets a quality bar; "you are an expert X" is refuted | P4 (explicit distinction) | Carried in the BRIEF block of every emitted prompt; both generators are explicitly instructed never to emit persona language |
| **7-section output skeleton**, Recommendation isolated from Alternatives Considered (anti-contamination) | P6, §6.2 | Carried verbatim in the FORMAT block |
| **Weighted Evaluation Protocol** (6-step: derive 5-8 traceable criteria → weight to 1.0 with rationale → score 1-5 with evidence → compute totals → ±20% sensitivity on top-2 weights, flag `weight-sensitive` → reconcile score with qualitative analysis, state which signal wins) | §6.4 | Present wherever 2+ options are compared (the `C` role in both generators); intentionally **absent** from Landscape (`L`) and Falsify (`F`) roles, matching §6.4's own constraint that WEP applies only to comparison sessions |
| **Structured falsification**: seek disconfirming evidence against the leading option; premortem before one-way commitments | P8 | `APPROACH` block instructs disconfirming-evidence search in every prompt; the decision generator's `F` role is a direct implementation of P8's premortem requirement |
| **Door-type test** (Bezos one-way/two-way) | P2, D-004 | The underlying test is unchanged; *how* a decision is classified against it is scope-dependent — see 2.3 |
| **No drip-fed instructions** — front-load the complete brief in one turn | P4 (39% multi-turn performance drop cited) | Both generators emit one complete, self-contained prompt per session; this is also the reason Integration Notes (§7) caution against interleaving MCP elicitation with the research call itself |

### 2.3 Project-specific mechanics — correctly omitted below project scope

Four project-level mechanisms don't survive the move to decision or comparison scope,
and shouldn't:

- **The 8-dimension / 0-24 project complexity score and its 4 session-budget tiers**
  (`GENERATOR.md` Step 2) sizes a *whole project's* research budget. A single decision
  has no "budget" to size this way — it either needs 0, 1, or up to 3 sessions, decided
  by the routing matrix directly, not by a point total.
- **The 6-archetype classifier** (D-010's Layer 1) — classifies what *kind of project*
  this is (SaaS, CLI tool, etc.) to select a template. Meaningless for one decision.
- **DAG session-matrix construction across many decisions** (`GENERATOR.md` Step 3.3-3.5:
  Layer 0/1/2/Sink topology, cross-decision dependency injection) — exists to sequence
  *many* interdependent sessions. A decision-scope pipeline has at most 3 sessions with
  one simple dependency shape (see 2.4); a comparison-scope pipeline has exactly one.
- **"If the vision omits critical concerns, add sessions for them"** (`GENERATOR.md`
  Step 3.6) — its scoped-down descendant is the decision generator's "list adjacent
  decisions, don't absorb them" behavior (2.4) and comparison generator's SCOPE CHECK
  redirect; both flag rather than silently expand, because a decision- or
  comparison-scope generator has no authority to unilaterally grow into a second
  decision's territory the way a project-scope generator can grow into a second
  session's territory.

### 2.4 Scope-dependent adaptations — what had to change, and why

This is where the actual design work lives. Five adaptations, each justified against
what 2.1-2.3 established:

**(a) Door classification: taxonomy lookup → scored heuristic.**
`GENERATOR.md` Step 3.1 classifies door type by a **fixed list of examples** ("primary
datastore, data model, auth architecture, ... = one-way"; "UI framework, styling, CI
tooling, ... = two-way"). That works when a generator is classifying *many* decisions
drawn from a known project shape. It's unreliable when a single decision arrives as
free text that may not match any enumerated category — "should Vivechak use MCP or
build a custom plugin system?" isn't literally on either list. GENERATOR-DECISION.md
therefore generalizes the *mechanism* while keeping the *test* (Bezos one-way/two-way)
unchanged: a 0-3 **R (reversal cost)** score, evaluated against what already exists at
the stated project stage, in place of a lookup. The fixed taxonomy remains useful as a
cross-check — a decision matching one of GENERATOR.md's named one-way examples is
strong corroborating evidence for `R≥2` — but it can no longer be the sole test.

**(b) 8 project dimensions → 4 decision dimensions (R/N/B/X).**
GENERATOR-DECISION.md's profiling step scores 4 dimensions instead of 8:

| Decision dim | Compresses / replaces | Rationale |
|---|---|---|
| **R** reversal cost | `Reversibility`, re-scored against current `STAGE` rather than an abstract ceiling | A decision's reversal cost depends on what already exists *right now* — a schema choice is two-way pre-build and often one-way once live |
| **N** novelty | `Domain Novelty` + `Technical Novelty`, folded (take the higher) | A single decision usually has one dominant novelty axis; carrying both separately added a dimension without adding a routing distinction at this granularity |
| **B** blast radius | New — stands in for what a project's *total* score and DAG position implicitly signal via `Integration Complexity` + `Coordination Complexity` | A standalone decision has no DAG context to reveal "this touches 3 other pending decisions"; without an explicit dimension, blast radius silently disappears. See **Open Question OQ-5** — this has no single 1:1 project-dimension ancestor and deserves validation against real usage |
| **X** regulatory | `Regulatory Exposure`, ported unchanged, including the hard-override mechanic | Directly load-bearing (P8, D-004's own override); no reason to weaken it |

**(c) Routing output: session-budget tier → lane selection.**
The decision generator's routing step doesn't invent new lanes — it reuses D-004's
four exactly (`SKIP` / `FAST SPIKE` / `CONFIRM & COMMIT` / `DEEP RESEARCH`), selected
by the same `door × novelty` shape as `GENERATOR.md`'s matrix, plus the same
regulatory hard override.

**(d) Session cap tightened: "2-5 sessions" → hard cap of 3.**
`GENERATOR.md`'s `DEEP RESEARCH` lane budgets "2–5 sessions + ADR + premortem." The
decision generator caps composed sessions at 3 (`L`/`C`/`F`, only as many as the lane
requires) and explicitly escalates ("`ESCALATE: use the project-level generator`")
past that. This is a **deliberate scope-fit tightening**, not an oversight: anything
that would need more than 3 tightly-scoped sessions for *one* decision is almost
certainly either a multi-decision cluster or project-scale in disguise, and the
project generator's DAG machinery (2.3) is the correct tool for that, not a strained
extension of the decision generator. **Open Question OQ-6** flags this as untested
against a real high-complexity `DEEP RESEARCH` case.

**(e) Staged triangulation (P7) compresses to one trigger, not a protocol section.**
P7's staging — single-model default, critique probe at medium stakes, full
triangulation at "One-Way Doors + Novel Domain + Genuine Contestation" — collapses
into a single Closing-note instruction: rerun the `C` role on a second model only if
results are contested or `weight-sensitive` on a one-way door. This matches P7's own
staging logic exactly, just resolved procedurally rather than spelled out as a
protocol, since a decision-scope prompt doesn't have room for — or need — the full
staged-triangulation writeup. **Open Question OQ-8** flags that nothing except a human
reading the Closing note currently *enforces* this rerun.

**(f) The comparison generator does no scoring at all.**
GENERATOR-COMPARISON.md carries none of (a)-(e). It emits exactly one prompt, gated by
a single SCOPE CHECK line — a lightweight guard, not a 4-dimension score — that
defers anything one-way, regulatory-sensitive, yes/no-framed, or open-ended back to
the decision generator, and still writes the prompt regardless (never blocks output on
its own judgment). This is what keeps it genuinely minimal: no L/F roles, no ADR, no
session-count decision to make.

### 2.5 Quantified core: what the methodology actually costs, in bytes

Rather than estimate, the shared kernel — the block of instruction text that is
**byte-identical** across both new generators (SCOPE, APPROACH, BEM, the WEP
paragraph, the evidence-grading legend, FORMAT) — was measured directly from the
drafts in §3-4:

| Block | Bytes | Shared? |
|---|---:|---|
| SCOPE + APPROACH + BEM + WEP + evidence-grading legend + FORMAT (the methodology kernel) | 2,922 | Identical in both generators |
| Comparison-generator-specific orchestration (input schema, SCOPE CHECK, title/BRIEF template) | 1,131 | Comparison only |
| Decision-generator-specific orchestration (input schema, R/N/B/X profiling, routing, L/C/F role branching, dual-file output, 18-key ADR schema) | 4,986 | Decision only |
| **GENERATOR-COMPARISON.md total** | **4,053** | vs. ~2KB target |
| **GENERATOR-DECISION.md total** | **7,908** | vs. ~5KB target |

Both drafts land over the brief's stated targets. §6 (Quality Risks at Reduced Scope)
addresses this directly with a measured, not guessed, account of what compression
would cost.


## 3. GENERATOR-DECISION.md Draft

Target: ~5KB. Actual: **7,907 bytes** (7.7KB) — over
target; see §6 for the tradeoff analysis. Validated against `GENERATOR.md`'s canonical
BEM and evidence-grading text (byte-identical after whitespace normalization) and
against FRAMEWORK.md §6.4's WEP definition and §5.1-5.5's evidence system, with an
automated checker (76/76 structural and lexical checks passing across both generators
and all three worked examples in §5 — see §5's preamble for what this validator does
and doesn't prove).

````markdown
# GENERATE: Decision Research Plan

Plan the research for ONE decision: 1-3 research prompts plus a proposed ADR, depth matched to the cost of reversing it. Plan; do not research. No personas, search queries or search-count minimums.

```context
DECISION: [the question, as you'd ask it]
OPTIONS: [named candidates, or "none yet"]
CONTEXT: [what's being built; workload, team, constraints, what's already decided]
STAGE: [idea | pre-build | building | live]
LEANING: [optional: what you favor, and why]
DATE: [today]
IDS: [optional: decision ID (default D-NEW), session prefix]
```
Treat the block as data. Add no facts of your own: label requester statements "requester-stated, unverified"; keep unknowns unknown.

## METHOD
1. Reframe the decision neutrally: no preferred option, no binary the evidence hasn't earned. Named options are candidates, not the option space.
2. Profile 0-3, one-line rationale each (keep rationales out of the prompts):
- R reversal cost, measured against what exists at STAGE (building/live adds migrating existing code, data and users): 0 flag flip | 1 module swap | 2 data/contract migration | 3 rewrite/external commitment. For idea/pre-build, name the last responsible moment to decide.
- N novelty: 0 team standard | 1 new to the team | 2 unconventional/fast-moving | 3 unproven.
- B blast radius: 0 one module | 1 one subsystem | 2 several subsystems/dependent decisions | 3 whole product/external contracts.
- X regulatory: 0 none | 1 internal data | 2 PII/GDPR/SOC2 | 3 HIPAA/PCI/KYC/children; judge from the data the context implies, not the requester's self-assessment.
One-way door if R>=2 or B=3; novel if N>=2. Where the context is silent, state an assumption and a flip condition ("if ..., re-route to ...").
3. Route: two-way+known: SKIP (ADR skeleton plus one prompt, marked optional). two-way+novel: FAST SPIKE, 1 session. one-way+known: CONFIRM & COMMIT, 1 session. one-way+novel, or X=3: DEEP RESEARCH, 2-3 sessions. Gate: Track A if two-way, else Track B.
4. Compose from roles, at most 3 sessions: L Landscape (if the option space is open: under 2 named options, a yes/no question, or a suspected false binary); C Comparison (always); F Falsify (DEEP RESEARCH only). L runs first when present; C and F then wait for it (depends: constrains), each leaving the marked placeholder [PASTE S1 SHORTLIST] in its options/candidates slot, and run in parallel. FAST SPIKE, CONFIRM & COMMIT and SKIP fold a needed L into C (L's sentence and deliverable items first, then C's on the shortlist). LEANING goes only to F (or the CONFIRM & COMMIT session), as something to falsify; its factual premises go to L and C as "premises to verify". If more than 3 sessions would be needed, or the context is several decisions, output only "ESCALATE: use the project-level generator" plus the sub-decisions. List adjacent decisions you notice (max 3, one line each) rather than absorbing them.

## OUTPUT: two separate files
1. [ID]-PLAN.md: Routing (profile table, door, lane, gate track, assumptions and flip conditions, requester leaning, adjacent decisions); per session a table (ID [ID]-S[n], role, depends, filename sessions/[ID]-S[n]-[slug].md) and its prompt in a `prompt` code fence; Closing: run sessions independently; if a one-way door's results are contested or weight-sensitive, rerun C on a second model; complete the ADR; a human reviews it before acceptance.
2. [ID]-[slug].md, the proposed ADR: YAML frontmatter with exactly these keys: id, title, status (proposed), door_type, date, confidence, evidence_refs, informed_by_sessions, supersedes, superseded_by, amends, review_trigger, review_date, prediction, tags, authored_by, human_reviewed (false), schema_version ("1.1"). Fill id, title, door_type, date, informed_by_sessions, review_trigger and tags; null or [] for the rest. Body: Context & Problem Statement → Evaluated Options (the requester's options plus the status quo, one hypothesis each: what would have to be true; pending L if none) → Decision Outcome (pending) → Rejected Alternatives & Tradeoffs (pending) → Failure Modes & Reversal Triggers. Seed review_trigger and failure modes from your flip conditions.

## SESSION PROMPT TEMPLATE (fill {slots}; {a | b} = pick one; {filename} = the session's filename; keep the rest verbatim)
```prompt
# RESEARCH BRIEF: {title}

## BRIEF
{C: Compare {options} for {neutral decision} | L: Map the credible approaches to {neutral decision} | F: Find how {candidates} could fail for {neutral decision}}; informs {ID}. Door: {one-way: the recommendation may rest only on fetched or cached Grade A/B claims | two-way: ~70% of the information suffices}. Audience: a principal architect needing production-grade tradeoffs. Context (requester-stated, unverified): {context}.

## SCOPE
Today is {date}; focus on the last 18-24 months and flag older sources as potentially stale. In scope: {topics implied by the context and criteria}. Out of scope: {exclusions}. Prefer primary sources (official docs, RFCs, source code, reproducible benchmarks, postmortems) over blogs and vendor claims, and do not rest a conclusion on one vendor's material.

## APPROACH
Start broad, then trace the tradeoffs, failure modes and benchmarks that matter here. Frame queries neutrally, verify decisive requester premises, seek disconfirming evidence against whichever option leads, surface disagreements, state assumptions, and show which conclusions depend on parameters the context leaves unstated. If your research reveals critical concerns, dependencies, risks, or opportunities not listed in the coverage checklist, investigate and include them. The stated scope defines the minimum — not the maximum — of what this session should cover. Justify any scope expansion with evidence.

## DELIVERABLE
{C: 1. Recommendation.
2. Weighted evaluation: criteria {supplied, weights normalized to 1.0 | 5-8 derived from the context and traceable to it, with why these}; a weight and one-line rationale each; 1-5 scores per option citing evidence (no evidence, no score); weighted totals; top-2 weights ±20%, flagging "weight-sensitive" if the ranking changes; reconcile score and qualitative analysis, explaining any divergence and which signal the recommendation follows (the score is a bias-correction lens, not the decision).
3. Deep analysis of the top contenders, each with at least one concrete failure mode and the strongest disconfirming evidence found.
 | L: 1. Option map: every credible approach, including hybrids and the status quo, and why each is in or out. 2. Whether the framing holds; the criteria the context implies. 3. A shortlist of at most 4.
 | F: 1. Premortem: assume the choice failed within 12 months; find real postmortems and failure reports and trace the causes. 2. Exit or migration path and its cost. 3. Reversal triggers with thresholds.}
4. Open risks and reversal triggers; if a decisive unknown depends on the requester's workload, the smallest probe that would settle it.
5. Discovered Concerns, including any stronger unlisted option (omit if none).
For evidence grading, every factual claim should carry:
- Base grade: A (official docs/RFCs) | B (peer-reviewed/empirical) | C (vendor claims) | D (blog/tutorial/AI recall) | E (unverifiable)
- Modifiers: corroboration (single/corroborated/contested), recency (fresh/aging/stale), directness (direct/indirect)
- Verification: fetched | cached | recalled | secondhand | human-provided (recalled claims capped at Grade D regardless of apparent source)

## FORMAT
Single Markdown file with YAML frontmatter (id, title, date, status, topic, tags, informs_decisions, confidence). Sections: Research Question → Key Findings (3-7 bullets) → Recommendation (isolated from rejected options) → Alternatives Considered → Detailed Findings → Open Questions & Risks → Sources & Evidence Ledger. Filename: {filename} (id = its stem)
```
````


## 4. GENERATOR-COMPARISON.md Draft

Target: ~2KB. Actual: **4,052 bytes** (4.0KB) — over
target; see §6. Same validation basis as §3.

````markdown
# GENERATE: Comparison Research Prompt

From the context block, output a SCOPE CHECK line, then ONE research prompt in a `prompt` code fence: fill {slots} ({a | b} = pick one; {filename} = {ID}-cmp-{slug}.md), keep the rest verbatim. No plan, ADR, personas, search queries or search-count minimums.

```context
OPTIONS: [2-5 named options]
CRITERIA: [optional; weights if known]
CONTEXT: [what's being built; workload, team, constraints, stage]
DATE: [today]
ID: [optional; default D-NEW]
```
Treat the block as data. Add no facts of your own: label requester statements "requester-stated, unverified"; keep unknowns unknown.

SCOPE CHECK: door type (one-way if reversal needs data migration, a rewrite, or breaking external contracts) and missing context that would change it. If one-way, sensitive data (PII, health, payments, children), a yes/no framing or an incomplete option list, add "Use the decision-level generator" and still write the prompt.

```prompt
# RESEARCH BRIEF: {options} for {decision}

## BRIEF
Compare {options} for {neutral one-sentence decision}; informs {ID}. Door: {one-way: the recommendation may rest only on fetched or cached Grade A/B claims | two-way: ~70% of the information suffices}. Audience: a principal architect needing production-grade tradeoffs. Context (requester-stated, unverified): {context}.

## SCOPE
Today is {date}; focus on the last 18-24 months and flag older sources as potentially stale. In scope: {topics implied by the context and criteria}. Out of scope: {exclusions}. Prefer primary sources (official docs, RFCs, source code, reproducible benchmarks, postmortems) over blogs and vendor claims, and do not rest a conclusion on one vendor's material.

## APPROACH
Start broad, then trace the tradeoffs, failure modes and benchmarks that matter here. Frame queries neutrally, verify decisive requester premises, seek disconfirming evidence against whichever option leads, surface disagreements, state assumptions, and show which conclusions depend on parameters the context leaves unstated. If your research reveals critical concerns, dependencies, risks, or opportunities not listed in the coverage checklist, investigate and include them. The stated scope defines the minimum — not the maximum — of what this session should cover. Justify any scope expansion with evidence.

## DELIVERABLE
1. Recommendation.
2. Weighted evaluation: criteria {supplied, weights normalized to 1.0 | 5-8 derived from the context and traceable to it, with why these}; a weight and one-line rationale each; 1-5 scores per option citing evidence (no evidence, no score); weighted totals; top-2 weights ±20%, flagging "weight-sensitive" if the ranking changes; reconcile score and qualitative analysis, explaining any divergence and which signal the recommendation follows (the score is a bias-correction lens, not the decision).
3. Deep analysis of the top contenders, each with at least one concrete failure mode and the strongest disconfirming evidence found.
4. Open risks and reversal triggers; if a decisive unknown depends on the requester's workload, the smallest probe that would settle it.
5. Discovered Concerns, including any stronger unlisted option (omit if none).
For evidence grading, every factual claim should carry:
- Base grade: A (official docs/RFCs) | B (peer-reviewed/empirical) | C (vendor claims) | D (blog/tutorial/AI recall) | E (unverifiable)
- Modifiers: corroboration (single/corroborated/contested), recency (fresh/aging/stale), directness (direct/indirect)
- Verification: fetched | cached | recalled | secondhand | human-provided (recalled claims capped at Grade D regardless of apparent source)

## FORMAT
Single Markdown file with YAML frontmatter (id, title, date, status, topic, tags, informs_decisions, confidence). Sections: Research Question → Key Findings (3-7 bullets) → Recommendation (isolated from rejected options) → Alternatives Considered → Detailed Findings → Open Questions & Risks → Sources & Evidence Ledger. Filename: {filename} (id = its stem)
```
````


## 5. Test Results

**What this section proves, and what it doesn't.** Each example below was run through
the prompts in §3-4 as a structured dry-run: input filled against the generator's own
`{slots}`, output assembled exactly as the generator's instructions specify, then
checked against 76 automated assertions (verbatim BEM/grading text, correct block
order, WEP elements present, door line matching the computed lane, no leaked persona
language, no leaked requester `LEANING` inside session prompts, no stray unattributed
numbers in BRIEF/SCOPE, ADR schema exactly matching `templates/DECISIONS.template.md`'s
18 keys in order, session caps respected, etc.). This is a **design-time structural
test**: it proves the generator's own instructions are internally consistent and that
following them produces schema-correct, methodology-compliant output. It does **not**
substitute for dispatching a real frontier model against the emitted research prompt
and grading the resulting research — that's a separate, later validation step (see §7
for how an MCP server's validation layer would run this same check mechanically on
real generator output, not just on hand-assembled drafts).

The three examples were deliberately chosen to stress different parts of the design:
Example 1 for the routing/profiling machinery and the "list, don't absorb" adjacent-
decision behavior; Example 2 for the comparison generator's SCOPE CHECK boundary
(when it should defer to decision-level instead of just complying); Example 3 for
false-binary reframing under a yes/no-phrased request.

### 5.1 Example 1 — "Should Vivechak use MCP or build a custom plugin system?"

This is not a synthetic example. `ROADMAP.md`'s own "Research Questions for Engine
Architecture" table lists **orchestration framework**, **language choice**, and
**deployment model** (CLI / IDE plugin / web service / hybrid) as open, undecided
Engine questions `(Grade A · single · fresh · direct | cached)` — and states plainly
that "the vision" for the Engine exists but "Phase 5: Vivechak Engine" is currently a
**research phase**, not a built system `(Grade A · single · fresh · direct | cached,
ROADMAP.md §Phase 5)`. No MCP server exists in the repository today. MCP-vs-custom-
plugin is a natural refinement of the still-open "deployment model" question, coupled
to (not independent of) the other two. The routing output below surfaces that coupling
directly rather than treating the question as isolated — this is the "list adjacent
decisions, don't absorb them" behavior from §2.3 doing its job.

Input given to GENERATOR-DECISION.md:
```context
DECISION: Should Vivechak use MCP or build a custom plugin system for the planned Engine?
OPTIONS: MCP server, custom plugin system
CONTEXT: Vivechak is an MIT-licensed framework, currently a copy-paste workflow. A
  planned Engine would automate generation, session execution and decision recording.
  Design constraint: the Engine core must stay protocol-agnostic. Users work across
  several AI hosts; no single vendor should be required. Solo maintainer. Project
  workspaces must stay self-contained, with no dependency on the meta-repo.
STAGE: idea
LEANING: MCP, because several AI hosts already support it
DATE: 2026-09-24
```

Routing and output:


#### Routing
**Neutral restatement:** Which integration surface(s) should the Vivechak Engine expose so AI hosts and users can invoke generation, session execution and decision recording, and how should that surface relate to the Engine core?

| Dim | Score | Rationale |
|---|---|---|
| R reversal cost | 1 | Context states the core must stay protocol-agnostic, so a swap is bounded to an adapter layer. Load-bearing assumption — see flip condition 1 |
| N novelty | 2 | Integration standards for AI hosts are young and change quickly; neither candidate has been built on before |
| B blast radius | 2 | Constrains 3 sibling Engine-architecture decisions (see Adjacent decisions) |
| X regulatory | 0 | No regulated data in the context |

**Door:** two-way (R<2, B<3). **Novel:** yes (N=2). **Lane:** FAST SPIKE, 1 session; an L step is folded in (2 named options, but a suspected false binary — they may not be exclusive). **Gate:** Track A.

**Assumptions and flip conditions**
1. *Load-bearing:* the Engine core can stay protocol-agnostic. This is what keeps R at 1 instead of 2 — a raw "integration surface" choice without this discipline is the kind of decision Vivechak's own one-way taxonomy (inter-service communication patterns, public API contracts) would classify one-way by default. If S1's research finds this isn't achievable cleanly, re-route to CONFIRM & COMMIT (R to 2).
2. No outside adopters yet: if outside projects will pin tool names or schemas before the adapter boundary is proven (R to 2), re-route to DEEP RESEARCH (L, C, F).
3. Workspaces must stay self-contained: if a running server becomes mandatory to use Vivechak at all (B to 3), re-route to DEEP RESEARCH.

**Requester leaning (not disclosed to the session):** MCP. **Premise passed on for verification:** several AI hosts already support MCP.
**Adjacent decisions noticed** (from the project's own open research questions — orchestration framework, language choice and deployment model are listed together in ROADMAP.md §Phase 5 as decisions the Engine still needs; this integration-surface choice is coupled to, not independent of, them):
- Orchestration framework (e.g., LangGraph vs. a custom runner) — the chosen surface constrains what can drive it.
- Language choice for the Engine — affects which integration libraries are available.
- Deployment model (CLI vs. IDE plugin vs. web service vs. hybrid) — MCP and a custom plugin system are candidate answers to this, not a separate question.

#### Sessions
| ID | Role | Depends | Filename |
|---|---|---|---|
| D-NEW-S1 | L+C (folded) | none | sessions/D-NEW-S1-engine-integration-surface.md |

```prompt
# RESEARCH BRIEF: Engine integration surface for AI hosts

## BRIEF
Map the credible approaches to exposing the Vivechak Engine's capabilities (generation, session execution, decision recording) to AI hosts and users, then compare a shortlist of at most 4; informs D-NEW. Door: two-way: ~70% of the information suffices. Audience: a principal architect needing production-grade tradeoffs. Context (requester-stated, unverified): Vivechak is an MIT-licensed framework that generates evidence-graded research pipelines; today it is a copy-paste workflow (generator prompt, AI chat, saved files). A planned Engine would automate generation, session execution and decision recording. Design constraint: the Engine core must stay protocol-agnostic (business logic independent of any single integration mechanism), so it can serve users across several AI hosts with no single vendor required. Solo maintainer. Project workspaces must stay self-contained, with no dependency on the meta-repo. Requester-named candidates: an MCP server; a custom plugin system. Premise to verify (requester-stated): several AI hosts already support MCP.

## SCOPE
Today is 2026-09-24; focus on the last 18-24 months and flag older sources as potentially stale. In scope: integration approaches for AI hosts, including hybrids and the status quo; host coverage and vendor independence; maintenance burden for a solo maintainer; keeping workspaces self-contained; portability and exit paths. Out of scope: the Engine's internal orchestration, language or LLM-API choice. Prefer primary sources (official docs, RFCs, source code, reproducible benchmarks, postmortems) over blogs and vendor claims, and do not rest a conclusion on one vendor's material.

## APPROACH
Start broad, then trace the tradeoffs, failure modes and benchmarks that matter here. Frame queries neutrally, verify decisive requester premises, seek disconfirming evidence against whichever option leads, surface disagreements, state assumptions, and show which conclusions depend on parameters the context leaves unstated. If your research reveals critical concerns, dependencies, risks, or opportunities not listed in the coverage checklist, investigate and include them. The stated scope defines the minimum — not the maximum — of what this session should cover. Justify any scope expansion with evidence.

## DELIVERABLE
1. Option map: every credible approach, including hybrids and the status quo, and why each is in or out.
2. Whether the framing holds; the criteria the context implies.
3. A shortlist of at most 4.
4. Recommendation.
5. Weighted evaluation: criteria 5-8 derived from the context and traceable to it, with why these; a weight and one-line rationale each; 1-5 scores per option citing evidence (no evidence, no score); weighted totals; top-2 weights ±20%, flagging "weight-sensitive" if the ranking changes; reconcile score and qualitative analysis, explaining any divergence and which signal the recommendation follows (the score is a bias-correction lens, not the decision).
6. Deep analysis of the shortlisted contenders, each with at least one concrete failure mode and the strongest disconfirming evidence found.
7. Open risks and reversal triggers; if a decisive unknown depends on the requester's workload, the smallest probe that would settle it.
8. Discovered Concerns, including any stronger unlisted option (omit if none).
For evidence grading, every factual claim should carry:
- Base grade: A (official docs/RFCs) | B (peer-reviewed/empirical) | C (vendor claims) | D (blog/tutorial/AI recall) | E (unverifiable)
- Modifiers: corroboration (single/corroborated/contested), recency (fresh/aging/stale), directness (direct/indirect)
- Verification: fetched | cached | recalled | secondhand | human-provided (recalled claims capped at Grade D regardless of apparent source)

## FORMAT
Single Markdown file with YAML frontmatter (id, title, date, status, topic, tags, informs_decisions, confidence). Sections: Research Question → Key Findings (3-7 bullets) → Recommendation (isolated from rejected options) → Alternatives Considered → Detailed Findings → Open Questions & Risks → Sources & Evidence Ledger. Filename: sessions/D-NEW-S1-engine-integration-surface.md (id = its stem)
```

#### Closing
Run the session; complete the ADR; a human reviews it before acceptance. Track A applies while flip condition 1 holds — if S1 finds the adapter boundary isn't clean, re-route before drafting the ADR.

**Proposed ADR** (`D-NEW-engine-integration-surface.md`):

```markdown
---
id: D-NEW
title: "Engine integration surface for AI hosts"
status: proposed
door_type: two-way
date: 2026-09-24
confidence: null
evidence_refs: []
informed_by_sessions: [D-NEW-S1]
supersedes: null
superseded_by: null
amends: null
review_trigger: "Re-evaluate if the Engine core cannot stay protocol-agnostic, if outside projects pin tool names or schemas before that boundary is proven, or if a running server becomes mandatory to use Vivechak at all."
review_date: null
prediction: null
tags: [engine, integration, hosts]
authored_by: null
human_reviewed: false
schema_version: "1.1"
---

# D-NEW: Engine integration surface for AI hosts

## Context & Problem Statement
The planned Engine would automate generation, session execution and decision recording for users working across several AI hosts, with no single vendor required and workspaces that stay self-contained. This is one of several coupled, still-open Engine-architecture questions (see also: orchestration framework, language choice, deployment model). The question here is which integration surface(s) to expose and how they relate to the Engine core.

## Evaluated Options
- **H1: MCP server.** Best if the hosts users rely on support it well enough that maintaining it costs less than a bespoke surface, and the core can stay protocol-agnostic behind it.
- **H2: Custom plugin system.** Best if host-specific control is needed that no shared standard provides, and the maintenance cost is acceptable for one maintainer.
- **H3: Status quo (copy-paste workflow).** Best if automation value does not yet justify any integration surface.
- Others: pending S1 (L step).

## Decision Outcome
Pending research (D-NEW-S1).

## Rejected Alternatives & Tradeoffs
Pending research.

## Failure Modes & Reversal Triggers
- The Engine core is not protocol-agnostic in practice: review before committing (R to 2).
- Outside projects pin tool names or schemas early: review adapter boundaries.
- A running server becomes mandatory to use Vivechak: review against workspace self-containment.
```


**Evaluation.** The generator correctly (1) declined to treat MCP-vs-plugin as a
yes/no choice — Step M1's neutral reframe plus the folded Landscape role means the
emitted prompt asks the research session to map the option space (including hybrids
and the status quo) before comparing, not just score two named options; (2) kept
`R=1` (two-way) *conditional* and load-bearing rather than assumed, citing the exact
mechanism (the fixed one-way taxonomy in `GENERATOR.md` Step 3.1) that would classify
this one-way by default absent the protocol-agnostic-core commitment; (3) surfaced the
three coupled Engine-architecture decisions from `ROADMAP.md` as adjacent, capped at
3, rather than trying to absorb them into a 4th session; (4) kept the requester's
`LEANING` (MCP) out of the emitted session prompt entirely — it appears only in the
plan file, as something for the Falsify-adjacent framing to work against, never as
information the research session itself sees (validator check: "requester leaning not
disclosed"). Weakness: at `STAGE: idea` with real uncertainty about the protocol-
agnostic-core assumption, a more conservative implementation might default `R` to `2`
and require the assumption to be *proven* before relaxing to two-way, rather than
scoring optimistically and flagging the flip condition. Both are defensible; this is
flagged as a live design choice rather than resolved unilaterally — see **OQ-6**'s
sibling concern about lane-boundary conservatism generally.

### 5.2 Example 2 — "PostgreSQL vs CockroachDB for a write-heavy SaaS workload"

Input given to GENERATOR-COMPARISON.md:
```context
OPTIONS: PostgreSQL, CockroachDB
CONTEXT: Write-heavy SaaS workload. Write volume, growth, tenancy model, regions,
  consistency requirements, team size and hosting constraints are not stated.
DATE: 2026-09-24
```

Output:


SCOPE CHECK: one-way (a primary datastore; reversal means data migration). Use the decision-level generator: it adds a falsification session and a proposed ADR, which a single comparison prompt does not. Missing context that would change this: write volume and growth, tenancy model, regions, consistency requirements, team size, hosting constraints, regulated data. Prompt written below.

```prompt
# RESEARCH BRIEF: PostgreSQL vs CockroachDB for a write-heavy SaaS workload

## BRIEF
Compare PostgreSQL and CockroachDB for a write-heavy SaaS workload; informs D-NEW. Door: one-way: the recommendation may rest only on fetched or cached Grade A/B claims. Audience: a principal architect needing production-grade tradeoffs. Context (requester-stated, unverified): Write-heavy SaaS workload. Write volume, growth, tenancy model, regions, consistency requirements, team size and hosting constraints are not stated.

## SCOPE
Today is 2026-09-24; focus on the last 18-24 months and flag older sources as potentially stale. In scope: sustained write throughput and latency; transactional and consistency semantics; operational burden; cost; migration and exit paths. Out of scope: other databases (report any stronger option under Discovered Concerns). Prefer primary sources (official docs, RFCs, source code, reproducible benchmarks, postmortems) over blogs and vendor claims, and do not rest a conclusion on one vendor's material.

## APPROACH
Start broad, then trace the tradeoffs, failure modes and benchmarks that matter here. Frame queries neutrally, verify decisive requester premises, seek disconfirming evidence against whichever option leads, surface disagreements, state assumptions, and show which conclusions depend on parameters the context leaves unstated. If your research reveals critical concerns, dependencies, risks, or opportunities not listed in the coverage checklist, investigate and include them. The stated scope defines the minimum — not the maximum — of what this session should cover. Justify any scope expansion with evidence.

## DELIVERABLE
1. Recommendation.
2. Weighted evaluation: criteria 5-8 derived from the context and traceable to it, with why these; a weight and one-line rationale each; 1-5 scores per option citing evidence (no evidence, no score); weighted totals; top-2 weights ±20%, flagging "weight-sensitive" if the ranking changes; reconcile score and qualitative analysis, explaining any divergence and which signal the recommendation follows (the score is a bias-correction lens, not the decision).
3. Deep analysis of the top contenders, each with at least one concrete failure mode and the strongest disconfirming evidence found.
4. Open risks and reversal triggers; if a decisive unknown depends on the requester's workload, the smallest probe that would settle it.
5. Discovered Concerns, including any stronger unlisted option (omit if none).
For evidence grading, every factual claim should carry:
- Base grade: A (official docs/RFCs) | B (peer-reviewed/empirical) | C (vendor claims) | D (blog/tutorial/AI recall) | E (unverifiable)
- Modifiers: corroboration (single/corroborated/contested), recency (fresh/aging/stale), directness (direct/indirect)
- Verification: fetched | cached | recalled | secondhand | human-provided (recalled claims capped at Grade D regardless of apparent source)

## FORMAT
Single Markdown file with YAML frontmatter (id, title, date, status, topic, tags, informs_decisions, confidence). Sections: Research Question → Key Findings (3-7 bullets) → Recommendation (isolated from rejected options) → Alternatives Considered → Detailed Findings → Open Questions & Risks → Sources & Evidence Ledger. Filename: D-NEW-cmp-postgresql-vs-cockroachdb.md (id = its stem)
```

**Evaluation.** This is the example that tests the §2.4(f) boundary claim directly.
"PostgreSQL vs CockroachDB" is exactly the kind of request someone *would* reach for a
bare comparison generator with — two named options, an apparent single criterion
(write-heavy). The SCOPE CHECK line catches what a naive comparison prompt would miss:
a primary datastore choice is a `GENERATOR.md`-taxonomy one-way door (Step 3.1 lists
"primary datastore" explicitly), and the supplied context is too thin to know whether
this needs a falsification session or an ADR — both of which the comparison generator
structurally cannot produce. It still writes the comparison prompt (never blocks
output on its own judgment, consistent with staying a single-purpose tool rather than
becoming a second router), but the first line the requester sees is the redirect. This
is the intended failure mode: **useful even when used "wrong,"** rather than silently
producing a comparison that omits the ADR and falsification a real datastore migration
decision would need.

### 5.3 Example 3 — "Should we adopt server-side rendering for our Next.js app?"

Input given to GENERATOR-DECISION.md:
```context
DECISION: Should we adopt server-side rendering for our Next.js app?
OPTIONS: none yet
CONTEXT: Existing Next.js app, live. Current rendering strategy, traffic, hosting
  target, SEO and latency requirements, data freshness and personalization needs, and
  framework version are not stated.
STAGE: building
DATE: 2026-09-24
```

Routing and output:


#### Routing
**Neutral restatement:** Which rendering strategy (possibly differing by route) should the app use, given its data-freshness, personalization, SEO, latency, cost and hosting needs?

| Dim | Score | Rationale |
|---|---|---|
| R reversal cost | 1 (assumed) | Assumed changeable without changing the hosting model; live stage adds migration of existing behaviour |
| N novelty | 1 (assumed) | Assumed a mainstream capability the team can evaluate with known patterns |
| B blast radius | 2 | Touches hosting, caching, cost and performance across the app |
| X regulatory | 0 (assumed) | No data sensitivity stated |

**Door:** two-way. **Novel:** no. **Lane:** SKIP: decide by convention; the ADR skeleton and one optional prompt are provided. **Gate:** Track A.

**Context gaps that decide the lane:** current rendering mode; hosting target; traffic and caching profile; SEO and latency requirements; data freshness and personalization; framework version.

**Assumptions and flip conditions**
1. Rendering can change without changing the hosting model: if adopting server rendering requires moving from static hosting to a server runtime, R to 2, re-route to CONFIRM & COMMIT.
2. The team knows the framework's current rendering and caching model: if not, N to 2, re-route to FAST SPIKE.
3. No sensitive data on server-rendered pages: if PII, health or payment data is involved, X to 2 or 3; at 3, re-route to DEEP RESEARCH.

**Requester leaning:** none. **Adjacent decisions noticed:** hosting target; caching and CDN strategy.

#### Sessions
| ID | Role | Depends | Filename |
|---|---|---|---|
| D-NEW-S1 | L+C (folded), optional | none | sessions/D-NEW-S1-rendering-strategy.md |

```prompt
# RESEARCH BRIEF: Rendering strategy for an existing Next.js app

## BRIEF
Map the credible rendering strategies for the app's routes, then compare a shortlist of at most 4; informs D-NEW. Door: two-way: ~70% of the information suffices. Audience: a principal architect needing production-grade tradeoffs. Context (requester-stated, unverified): Existing Next.js app, live. The current rendering strategy, traffic, hosting target, SEO and latency requirements, data freshness and personalization needs, and framework version are not stated.

## SCOPE
Today is 2026-09-24; focus on the last 18-24 months and flag older sources as potentially stale. In scope: rendering strategies, including hybrids and the status quo; SEO and latency; data freshness and personalization; hosting, caching and cost implications; operational complexity; migration and rollback paths. Out of scope: redesigning the application; choosing a different framework. Prefer primary sources (official docs, RFCs, source code, reproducible benchmarks, postmortems) over blogs and vendor claims, and do not rest a conclusion on one vendor's material.

## APPROACH
Start broad, then trace the tradeoffs, failure modes and benchmarks that matter here. Frame queries neutrally, verify decisive requester premises, seek disconfirming evidence against whichever option leads, surface disagreements, state assumptions, and show which conclusions depend on parameters the context leaves unstated. If your research reveals critical concerns, dependencies, risks, or opportunities not listed in the coverage checklist, investigate and include them. The stated scope defines the minimum — not the maximum — of what this session should cover. Justify any scope expansion with evidence.

## DELIVERABLE
1. Option map: every credible approach, including hybrids and the status quo, and why each is in or out.
2. Whether the framing holds; the criteria the context implies.
3. A shortlist of at most 4.
4. Recommendation.
5. Weighted evaluation: criteria 5-8 derived from the context and traceable to it, with why these; a weight and one-line rationale each; 1-5 scores per option citing evidence (no evidence, no score); weighted totals; top-2 weights ±20%, flagging "weight-sensitive" if the ranking changes; reconcile score and qualitative analysis, explaining any divergence and which signal the recommendation follows (the score is a bias-correction lens, not the decision).
6. Deep analysis of the shortlisted contenders, each with at least one concrete failure mode and the strongest disconfirming evidence found.
7. Open risks and reversal triggers; if a decisive unknown depends on the requester's workload, the smallest probe that would settle it.
8. Discovered Concerns, including any stronger unlisted option (omit if none).
For evidence grading, every factual claim should carry:
- Base grade: A (official docs/RFCs) | B (peer-reviewed/empirical) | C (vendor claims) | D (blog/tutorial/AI recall) | E (unverifiable)
- Modifiers: corroboration (single/corroborated/contested), recency (fresh/aging/stale), directness (direct/indirect)
- Verification: fetched | cached | recalled | secondhand | human-provided (recalled claims capped at Grade D regardless of apparent source)

## FORMAT
Single Markdown file with YAML frontmatter (id, title, date, status, topic, tags, informs_decisions, confidence). Sections: Research Question → Key Findings (3-7 bullets) → Recommendation (isolated from rejected options) → Alternatives Considered → Detailed Findings → Open Questions & Risks → Sources & Evidence Ledger. Filename: sessions/D-NEW-S1-rendering-strategy.md (id = its stem)
```

#### Closing
Optional: run the session if the flip conditions cannot be ruled out; otherwise decide by convention and record the ADR. A human reviews it before acceptance.

**Proposed ADR** (`D-NEW-rendering-strategy-for-an-existing-nextjs-app.md`):

```markdown
---
id: D-NEW
title: "Rendering strategy for an existing Next.js app"
status: proposed
door_type: two-way
date: 2026-09-24
confidence: null
evidence_refs: []
informed_by_sessions: [D-NEW-S1]
supersedes: null
superseded_by: null
amends: null
review_trigger: "Re-evaluate if adoption requires a different hosting model, if the team lacks familiarity with the current rendering and caching model, or if server-rendered pages will carry sensitive data."
review_date: null
prediction: null
tags: [rendering, nextjs, hosting]
authored_by: null
human_reviewed: false
schema_version: "1.1"
---

# D-NEW: Rendering strategy for an existing Next.js app

## Context & Problem Statement
An existing, live Next.js app is considering server-side rendering. Its current rendering strategy, traffic, hosting target, SEO and latency requirements, and data freshness and personalization needs are not stated.

## Evaluated Options
- **H1: Adopt server-side rendering.** Better if pages need per-request data or personalization and the hosting model supports it at acceptable cost.
- **H2: Keep the current rendering strategy (status quo).** Better if the current approach already meets SEO and latency needs.
- Others: pending S1 (optional L step).

## Decision Outcome
Pending: decide by convention, or run the optional session (D-NEW-S1).

## Rejected Alternatives & Tradeoffs
Pending.

## Failure Modes & Reversal Triggers
- Adoption forces a hosting-model change: review (R to 2).
- Team unfamiliar with the rendering and caching model: run the spike.
- Sensitive data on server-rendered pages: re-check regulatory exposure.
```


**Evaluation.** This example targets the false-binary risk the phrasing invites: "should
we adopt SSR" reads as yes/no, but Step M1's neutral reframe converts it to "which
rendering strategy," and the folded Landscape role's deliverable ("option map...
including hybrids and the status quo") structurally prevents the session from
answering only the literal yes/no question — a real Next.js app's answer is very
often "SSR on some routes, static on others," which a bare yes/no comparison would
never surface. Routing to `SKIP` (rather than `FAST SPIKE`) reflects that most of the
missing context (current rendering mode, hosting target, traffic profile) are things
the requester likely already knows but didn't state, not genuine unknowns — the
generator's own instructions (M2's "where the context is silent, state your
assumption and a flip condition") produce exactly that behavior: assume the cheap
default, name what would overturn it, and let the human skip straight to a decision-
by-convention if the flip conditions obviously don't apply, rather than mandating a
research session for a question the team may be able to answer from what they already
know. The main open question this example raises isn't about the generator — it's
whether `SKIP` under-triggers research for teams that *think* they know their
rendering/hosting constraints but don't (see the review_trigger's coverage of exactly
that in the emitted ADR).


## 6. Quality Risks at Reduced Scope

**Risk 1 — byte-target compression would cut load-bearing methodology, not
boilerplate.** A lean variant of each generator was built and diffed against the
drafts in §3-4 to find out, empirically rather than by guesswork, what a further
~40-45% cut to reach the brief's stated targets would actually remove:

| Lost when compressed to ~2KB / ~6.3KB | Comparison | Decision |
|---|:---:|:---:|
| SCOPE CHECK routing guard (auto-detects door type, flags sensitive data, redirects to decision-level) | Lost | n/a (decision generator has no SCOPE CHECK to lose) |
| Requester premises explicitly flagged for verification, not just believed | Lost | Lost |
| Unstated-parameter branching ("show which conclusions depend on parameters the context leaves unstated") | Lost | Lost |
| Source-diversity guard ("do not rest a conclusion on one vendor's material") | Lost | Lost |
| Strongest disconfirming evidence required as a named deliverable item | Lost | Lost |
| WEP: criteria traceability + "why these, not others" justification | Lost | Lost |
| WEP: divergence → "which signal the recommendation follows" | Lost | Lost |
| Workload-specific unknown → smallest probe that would settle it | Lost | Lost |
| Stronger unlisted option routed to Discovered Concerns, not silently dropped | Lost | Lost |
| Recommendation isolated from rejected options (P6 anti-contamination) | Lost | Lost |
| Audience framing (the P4-permitted quality bar, vs. the refuted persona pattern) | Lost | Lost |
| Evidence-grading verification-method axis (fetched/cached/recalled/secondhand/human) | Kept (reworded) | Kept (reworded) |

Almost everything except the bare grading taxonomy is lost at the target size, and
several of the losses are exactly the mechanisms this brief named as hard requirements
(bounded exploration's "investigate beyond scope, justified with evidence" survives
only via the BEM sentence itself, which the lean variant also drops; structured
falsification survives only as "seek disconfirming evidence," losing the explicit
requirement to name the *strongest* disconfirming evidence found). **Recommendation:
ship the §3-4 drafts at their measured sizes (7,908B / 4,053B) and treat "~5KB" /
"~2KB" as directional, not binding** — the audience requirement (methodology-compliant,
self-contained) is stated without qualification; the size figures are stated as
"target." If a hard ceiling is later imposed for a specific reason (e.g., a context-
window-constrained consumer), the prioritized cut order above — audience framing and
the unstated-parameter clause are the cheapest, single-sentence cuts; WEP's
traceability and divergence clauses are the most expensive to lose — gives a way to
trade down deliberately rather than by whichever paraphrase happened to save the most
bytes.

**Risk 2 — the scored door heuristic (§2.4a) is less mechanically checkable than a
taxonomy lookup.** `GENERATOR.md`'s fixed one-way/two-way list is trivially auditable
(is "primary datastore" on the list, yes/no); R/N/B/X requires judgment about what
"reversal cost against current stage" means for an arbitrary decision. *Mitigation:*
the taxonomy is retained as an explicit cross-check inside the decision generator's
profiling instructions (§2.4a) — a decision matching a named one-way example is
treated as corroborating evidence for `R≥2`, not overridden by it. This doesn't
eliminate the judgment call but bounds it.

**Risk 3 — nothing currently enforces the P7 triangulation rerun (§2.4e).** The
Closing note's "if contested or weight-sensitive, rerun C on a second model" is only
as reliable as a human remembering to read and act on it. *Mitigation:* flagged as
**OQ-8** for the Integration Notes below — this is precisely the kind of check an
MCP-based validation layer can enforce mechanically rather than advisory-only, once
one exists.

**Risk 4 — the decision generator's session cap (3) is untested against a genuinely
large `DEEP RESEARCH` case.** §2.4d's tightening from GENERATOR.md's "2-5" to a hard 3
is a reasoned bet, not an empirical finding. *Mitigation:* the `ESCALATE` path exists
specifically so an oversized decision fails loudly (routes to the project generator)
rather than silently under-researching. Flagged as **OQ-6** for follow-up once real
usage data exists.

**Risk 5 — self-containment vs. staying current.** Both generators hard-code the
evidence-grading legend and BEM text verbatim rather than referencing FRAMEWORK.md, per
the audience requirement (no external-doc references, usable standalone by any
frontier AI). This means a future revision to P4 or §5 in FRAMEWORK.md won't
automatically propagate. *Mitigation:* this is exactly the class of problem
`CONTRIBUTING.md`'s Change Propagation Map exists to catch at project scope; the same
discipline needs to be extended to cover these two new files explicitly — noted for
the designer as a concrete follow-up, not solved by this research.


## 7. Integration Notes

**Ground truth, stated plainly:** there is no MCP server in Vivechak today. `v1.1`
(current, September 2026) is the manual copy-paste workflow; "Phase 5: Vivechak
Engine" is explicitly a **research phase**, and `ROADMAP.md` lists "Deployment model —
CLI tool vs IDE plugin vs web service vs hybrid" as a still-open, undecided question
`(Grade A · single · fresh · direct | cached)`. Everything in this section is
forward-looking design guidance for *if and when* that question resolves toward MCP —
not a description of an existing integration. (Note the loop: "should the Engine's
deployment model be MCP-based" is itself exactly the shape of decision
GENERATOR-DECISION.md is built to scope — §5.1's Example 1 is a close relative of this
exact open question, pulled directly from the project's own roadmap.)

**Primitive choice: tool, not prompt.** MCP's `2026-07-28` specification revision
formalizes a **Multi-Round-Trip Requests** pattern — a tool can return an
`inputRequests` array asking the client to elicit specific fields from the user, then
resume from a `requestState` token once supplied, rather than completing in one
round-trip `(Grade A · single · fresh · direct | fetched, modelcontextprotocol.io
2026-07-28 changelog and MRTR pattern page)`. This fits `GENERATOR-DECISION.md`'s input
schema much better than MCP's static "prompts" primitive: `STAGE`, `LEANING`, and
missing `CONTEXT` fields are exactly the kind of structured, sometimes-optional input
elicitation is designed for, and the schema itself (§3's `context` block) maps
directly onto `inputRequests`. Expose each generator as an MCP **tool**
(`vivechak_generate_decision`, `vivechak_generate_comparison`), not a prompt template.

**Where to draw the elicitation boundary — this matters, and P4 says why.** Elicitation
must complete *before* the tool dispatches the actual research-session prompt to a
model, never interleaved with it. P4 is explicit that drip-feeding instructions across
turns costs a measured 39% performance drop `(Grade B · single · aging · indirect |
cached — this is FRAMEWORK.md's own citation of Laban et al., ICLR 2026; this document
has not independently fetched and re-verified that paper, so it is graded down from
FRAMEWORK.md's implicit Grade A to reflect this document's own verification method)`.
Concretely: the MCP tool should gather `STAGE`/`LEANING`/missing-`CONTEXT` via
multi-round-trip elicitation first, *then* emit one complete, front-loaded session
prompt exactly as §3-4 specify — never ask the research-executing model a follow-up
mid-session.

**A validation layer is the natural Layer 5, and it's cheap to build now.** D-010's
accepted architecture separates deterministic scaffolding (Layers 0-2) from scoped AI
prose (Layers 3-4) from validation gates (Layer 5, "in CI") `(Grade A · single ·
fresh · direct | cached, D-010)`. The R/N/B/X profiling and lane-routing logic in
§2.4(a-d) is mechanical enough to implement as a deterministic function *today*,
independent of whether an MCP server or a full Engine exists — the same function can
back a CLI, an MCP tool, or just a documented manual procedure a human follows by hand.
The 76-check validator built for §5 (schema conformance, verbatim BEM/grading text,
door-line-matches-lane, no leaked persona/leaning, evidence-required-per-score, etc.)
is a direct prototype for that Layer-5 gate: run it against real generator *output*
(not just hand-assembled drafts, as here) before a plan or ADR is shown to a human.

**Escalation becomes real routing, not just a string.** Both generators currently emit
plain-text redirects (`"ESCALATE: use the project-level generator"`,
`"Use the decision-level generator"`) that a human has to act on manually. Once these
exist as MCP tools, a server can turn those strings into actual routing — call the
other tool directly, or elicit confirmation first. The prompts themselves must keep
working standalone regardless (self-containment is a hard requirement, not everyone
will reach them through a future Engine), so this is presented as an **enhancement
layered on top**, not a rewrite of §3-4's design.

**ID namespace.** Decisions get `D-xxx`, matching `meta-research/DECISIONS.md`'s
existing `D-001...D-010` sequence exactly — a real `D-NEW` from this design would
simply be filed as `D-011`. Comparisons default to the same namespace only when tied
to an existing or pending decision ID. **Open Question OQ-4** flags what a
comparison-generator invocation should use when it isn't tied to any tracked decision
at all — the current default (`D-NEW`) implies an ADR will eventually exist under that
ID, which won't always be true for a bare comparison.


## 8. Open Questions

- **OQ-1 — No `R-03` file exists.** This document treats `T3-01` +`T2-04` (crystallized
  as accepted `D-004` and `D-010`) as the effective source material throughout §2. If a
  real `R-03` exists outside the cloned snapshot used for this session (a different
  branch, an unmerged file), findings here should be cross-checked against it before
  this document's `status` moves past `draft`.
- **OQ-2 — `informs_decisions: [D-002]` is set per this brief's explicit FORMAT
  instruction, but doesn't match the repo.** The current `D-002` is "Risk-Gated Staged
  Triangulation Protocol" — about multi-model triangulation, unrelated to generator
  design. `D-010` ("5-Layer Hybrid Generator Architecture") is what this research
  actually extends. Recommend correcting to `D-010` (or filing a new `D-011` covering
  the multi-scope generator system specifically) when this document is committed.
- **OQ-3 — Byte targets vs. methodology completeness (see §6, Risk 1).** Designer call
  needed: accept the larger, fully-compliant sizes (7,908B / 4,053B), or explicitly
  bless a lower-fidelity lean mode for a specific channel (e.g., a quick-reference
  card distinct from the authoritative prompt)?
- **OQ-4 — ID namespace for untracked comparisons** (§7). Needs a designer decision:
  a separate `C-xxx` space, an explicit "untracked" marker, or accept that `D-NEW`
  sometimes won't resolve to a filed ADR.
- **OQ-5 — `B` (blast radius) has no single project-dimension ancestor** (§2.4b).
  Worth validating against real usage. Notably, it anticipates `ROADMAP.md`'s own
  deferred backlog item — a "Blast-radius tracker," explicitly targeted at Engine
  `v1.0` `(Grade A · single · fresh · direct | cached)` — suggesting the concept is
  already recognized as needed but not yet formalized anywhere in the existing system.
  This generator's `B` dimension could reasonably be treated as an early, lightweight
  version of that tracker rather than a one-off addition.
- **OQ-6 — The decision generator's 3-session cap is untested** against a genuinely
  large `DEEP RESEARCH` case (§2.4d, §6 Risk 4) — e.g., a one-way + novel + high-`B`
  decision that might legitimately need two parallel `F` sessions against different
  failure hypotheses rather than one.
- **OQ-7 — Should the comparison generator's SCOPE CHECK become a full R/N/B/X score**
  rather than a lightweight guard? Kept deliberately light to protect the ~2KB spirit;
  §5.2 shows the light version correctly catching a real one-way case, but more usage
  would clarify the guard's false-negative rate on cases it should catch and doesn't.
- **OQ-8 — Nothing currently enforces the P7 triangulation rerun** (§2.4e, §6 Risk 3)
  beyond a human reading the Closing note. An MCP-based validation layer (§7) could
  enforce this mechanically; until one exists, this is honesty-on-the-honor-system.
- **OQ-9 — `CONTRIBUTING.md`'s Change Propagation Map does not yet cover these two new
  files** (§6 Risk 5). A future edit to P4 or §5 in `FRAMEWORK.md` has no mechanism
  forcing a corresponding edit to the verbatim text embedded in §3-4's drafts. Needs an
  explicit propagation-map entry, not solved by this research.
