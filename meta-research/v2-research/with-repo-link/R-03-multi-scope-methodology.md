---
id: R-03
title: "Multi-Scope Methodology — Extending Vivechak from Project-Level to Decision- and Comparison-Level Research"
date: 2026-09-23
status: draft
topic: multi-scope
tags: [scope-model, methodology-invariants, complexity-scoring, generator-architecture, rapid-review, structured-analytic-techniques, adr, deep-research-tools]
informs_decisions: [D-002]
confidence: medium
---

> **Reading guide.** Inline evidence tags use FRAMEWORK §5.5 in compact form: `[S## | Grade·corroboration·recency·directness | verification]`. Grades follow FRAMEWORK §5.1 with the conventions stated in Detailed Findings §0. **INFERENCE** marks reasoning that is mine rather than a sourced fact; its evidence basis is named. Source IDs resolve in *Sources & Evidence Ledger*. Assumption: D-002 is the decision "adopt a multi-scope model"; its text was not available to me.

## Research Question

How do mature structured-research methodologies — medical evidence synthesis, intelligence analysis, software design/decision processes, and AI deep-research tools — vary process weight across scope levels, and what scope model lets Vivechak add **decision-level** (1–3 sessions → ADR) and **comparison-level** (1 session → weighted matrix) research alongside the existing **project-level** pipeline (4–30 sessions → FAD) *without regressing the project-level workflow*?

Sub-questions: (1) what is invariant vs. scope-dependent; (2) how do generator options A/B/C compare; (3) how should complexity scoring adapt below project scope; (4) what output format fits each scope; (5) what are the risks of scope expansion.

## Key Findings

1. **⚠ PROMINENT — the three-level model is incomplete; it conflates two axes.** "Project / decision / comparison" is a *unit-of-analysis* axis. "How much rigor and effort" is a separate *depth* axis, and every methodology surveyed keeps them apart: medical review types could not be classified reliably from titles or from workflow/timeframe, only from what the review does with search, appraisal, synthesis and analysis [S12 | A·single·aging·indirect | fetched]; "rapid" is a modifier that composes with other types (rapid scoping reviews exist) [S07 | B·single·fresh·indirect | secondhand]; the US Army chooses among five named planning methodologies using three independent criteria — scope of the problem, time available, staff availability [S23 | A·corroborated·aging·indirect | fetched]. Inside Vivechak the overlap is already visible: **Tier 0 (1–3 sessions) equals the proposed decision-level budget**, and the per-decision routing matrix already emits decision-level outcomes [S01, S02 | A·corroborated·fresh·direct | fetched]. **Correction:** model *Scope × Depth (+ Mode)*. Keep three user-visible scope entry points, but add (a) single-option *viability/spike* handling inside decision scope (with a mandatory baseline option), (b) a **landscape** mode for when the option set is unknown, and (c) a **revalidate** mode for fired `review_trigger`s. A fourth *portfolio* level above project has no demand evidence and is parked (cf. OVH-05 [S03]).
2. **Mature methodologies fix the standards and scale the process weight.** ICD 203 makes its analytic standards universal but requires application to fit each product's purpose, source scope, timeline and consumer [S22 | A·single·aging·indirect | fetched]. Rapid reviews modify systematic-review methods to accelerate while staying systematic, transparent and reproducible [S06 | A·corroborated·fresh·indirect | fetched]; scoping reviews still need explicit search strategies and standardized extraction [S10, S11 | A·corroborated·aging·indirect | fetched]. A Google "mini design doc" keeps every step of the full doc, only terser [S24 | B·single·aging·indirect | fetched]; mini-HTA applies HTA methodology refined in time, scope and content [S16 | B·single·stale·indirect | fetched]. This yields the invariants I1–I9 in Recommendation R4.
3. **Safe shortcuts narrow the *question*; unsafe ones relax the *evidence bar* — and stakes gate both.** Limiting question scope, date range or study type is rated as expected to have no validity impact, whereas limiting quality assessment is not recommended [S08 | A·single·aging·indirect | fetched]. Simulated PubMed-only rapid searches changed the primary odds ratio by more than 20% in about 10% of 2,512 Cochrane meta-analyses and flipped significance in 6.5–38.6% of cases across methods; the authors judge that acceptable for scoping or urgent work but not for guidelines or regulatory decisions [S09 | A·single·aging·indirect | fetched]. Danish mini-HTAs — the local, lightweight form — often omitted literature selection/interpretation, and only 25% quantified clinical effect [S15 | A·single·stale·indirect | fetched]. **Implication:** door type (reversibility), not scope size, must set the evidence bar; light templates must make grade/provenance fields *structurally required*.
4. **Scope is selected by explicit categories and thresholds, not by continuous scales.** Rust requires an RFC only for "substantial" changes, with sub-team-specific overrides [S26 | A·corroborated·fresh·indirect | fetched]; Google uses five screening questions (write a doc at ≥3 "yes") [S24]; Amazon uses binary Type 1/Type 2 and warns against one-size-fits-all process [S27 | B·corroborated·aging·indirect | secondhand]; the Army uses named methodologies [S23]. AI research tools expose 2–4 discrete levels — Gemini Deep Research vs Max [S31 | A·single·fresh·direct | fetched]; Perplexity low/medium/high search context and minimal–high reasoning effort [S33, S34 | A·single·aging·direct | fetched]; OpenAI full vs mini models plus a `max_tool_calls` cap [S36 | A·single·aging·direct | fetched]; Anthropic three effort classes with numeric ranges [S30 | B·corroborated·aging·direct | fetched]. Numeric knobs are *resource caps*, not methodology levels. **No 1–10 methodological depth scale was found anywhere** (absence of evidence, not a tested failure). Vivechak's own backlog already flags tier-boundary cliffs and an uncalibrated rubric (SM-01/SM-02), and OVH-07 chose a coarse 1–5 scale to avoid false precision [S03 | A·single·fresh·direct | fetched].
5. **Multiple distinct processes carry documented integration and drift costs.** Army doctrine describes ADM, MDMP and TLP but does not explain how to integrate conceptual and detailed planning [S23 | A·single·aging·indirect | fetched]; Danish hospitals produced 60 different forms (49 mini-HTAs) [S14 | A·single·stale·indirect | fetched]; Vivechak needed a Change Propagation Map after FRAMEWORK improvements failed to reach GENERATOR [S03, S04 | A·corroborated·fresh·direct | fetched]. **A live instance found in this session:** FRAMEWORK §5.1 places peer-reviewed studies in Grade A, GENERATOR's grade list places "peer-reviewed/empirical" in Grade B [S01, S02]. Three hand-maintained generators would triple that surface.
6. **Most decision/comparison machinery already exists inside the project pipeline; what is missing is entry contracts, output contracts and scope-specific boundary language.** Existing: routing matrix (SKIP / FAST SPIKE / CONFIRM & COMMIT / DEEP), Tier 0, Weighted Evaluation Protocol for comparison sessions, Track A/B gates, staged triangulation, ≤2 child sessions [S01, S02]. Pasting a single decision into GENERATOR.md would still hit its required output structure (SYN-01 synthesis, Phase 0 gate) — INFERENCE from the required-structure text [S01]. Also **INFERENCE**: at comparison scope the *decomposition* a generator exists to perform is trivial (one session), so a fill-in prompt template (FRAMEWORK §4.2 + §6.4) likely suffices — testable (Open Question 1).
7. **Design-option result (Vivechak §6.4 protocol, self-applied):** A′ (*declared-scope entry points on one shared core*) 4.20 > A (three generators) 3.25 > B (single auto-detecting generator) 3.05 ≫ C (continuous 1–10 parameter) 1.85. Rank order was stable in 12/12 ±20% single-weight perturbations. **A vs B is within scoring error (Δ0.20) and should be treated as a tie**; C-last and "declared scope beats silent detection" are the robust conclusions. A′'s scores are prospective (unbuilt) and were assigned by the same model that proposed it — see Alternatives Considered.

## Recommendation

**Adopt Option A′: declared-scope entry points on one shared methodology core, with derived (not user-dialled) depth. Reject C. Defer B until usage data exists.** Keep GENERATOR.md behaviourally unchanged. Confidence: `medium` (principles: strong cross-domain convergence; Vivechak-specific mechanics: inference, unbuilt).

### R1. Scope model = Scope × Depth (+ Mode)

| Axis | Values | Set by | Selects |
|---|---|---|---|
| **Scope** | `project` · `decision` · `comparison` | Declared by requester (`SCOPE:`), validated by the Scope Fit Check (R3) | input contract, decomposition rule, output contract |
| **Depth** | project: Tier 0–3 · decision: SKIP / SPIKE / CONFIRM / DEEP · comparison: fixed 1 session (+ optional critique probe / triangulation) | **Derived** (8-dim score or Decision Profile) — never a user dial | session count, gates, triangulation |
| **Mode** | `evaluate` (default) · `landscape` · `revalidate` | Derived: `landscape` if option set is open; `revalidate` if input is an existing D-NNN whose `review_trigger` fired | job of the first session |

Illustration of why the axes must stay separate: the brief's comparison example (*PostgreSQL vs CockroachDB, write-heavy*) is **comparison scope** but concerns the primary datastore, a one-way door in GENERATOR's list [S01] — so it needs **Track B rigor** (fetched/cached evidence, corroborated A/B claims, human review, premortem; triangulation if contested). "1 session" means one prompt, not one model run: triangulation multiplies cost (1×/1.3×/3×) [S02] without changing scope.

### R2. Architecture and file plan

```
vivechak/
├── GENERATOR.md                        unchanged (project scope)
├── GENERATOR-DECISION.md               NEW  thin generator: decision scope
├── templates/
│   ├── COMPARISON-SESSION.template.md  NEW  fill-in prompt: comparison scope (no generator)
│   └── DECISIONS / CONFLICT-RESOLUTION / FAD / PHASE-0-GATE   unchanged
├── FRAMEWORK.md                        + Scope×Depth section, invariants I1–I9 stated normatively
└── AGENTS.md / README.md               + "Choose your entry point" table + Scope Fit Check
```

- **Shared core.** One `CORE:BEGIN … CORE:END` block (evidence grading §5.1–5.5, 5-block anatomy §4.1, 7-section skeleton §6.2, routing matrix, Weighted Evaluation Protocol summary, invariants I1–I9) maintained once in FRAMEWORK.md and copied byte-identically into each new entry file. Extend the Change Propagation Map with a diff/checksum step. Assembly by tooling is deferred to Engine v0.1 (ROADMAP) [S03].
- **Compatibility.** `GENERATOR-DECISION.md` emits the *same two file names* (RESEARCH-PIPELINE.md, DECISIONS.md) with `scope: decision` and no SYN-01/FAD, so the downstream workflow in README/AGENTS.md (execute → record decision → gate) is untouched. **INFERENCE** (from README/AGENTS workflow [S04]).
- **Exit ramp.** The decision generator may return *"no research warranted — decide by convention, log Track A"* for two-way/known-pattern decisions. Google, Rust and Amazon all encode an explicit "don't run the heavy process" path [S24, S26, S27].

### R3. Entry contracts, Scope Fit Check, escalation

| Scope | Input contract (analogy: PICO/PCC fix the question structure per review type [S10, S11]) | Produces |
|---|---|---|
| project | Free-form vision (unchanged) | RESEARCH-PIPELINE.md + DECISIONS.md → sessions → FAD → Phase 0 gate |
| decision | `DECISION:` one sentence · `CONTEXT:` 3–8 lines (constraints, scale, team, horizon) · `OPTIONS:` list or "unknown" · `PARENT:` project / D-NNN or "none" | ≤3 session prompts + 1 D-NNN seed |
| comparison | `OPTIONS:` 2–5 named · `WORKLOAD/CONSTRAINTS:` the outcome spec · `INFORMS:` D-NNN or "new" · `DOOR:` one-way / two-way / unknown | 1 filled session prompt |

**Scope Fit Check (Step 0; also a human/agent triage in AGENTS.md).** Surface form is *not* a reliable cue — the brief's own two examples are both "X vs Y" yet differ in scope (*MCP vs custom plugin system*: option space open, criteria unstated; *PostgreSQL vs CockroachDB for write-heavy*: closed pair, stated workload). Use structural questions:
1. More than one interdependent decision on the table? → `project`.
2. ≥2 named options **and** a stated evaluation basis? → `comparison`; otherwise → `decision` (add `landscape` session if options are open).
3. Door type known? Unknown → classify in-session, default to the stricter bar.
4. Does the decision constrain ≥2 *undecided* decisions? → escalate to `project` or add a coupling-check session (**provisional threshold, uncalibrated**).
5. Existing D-NNN with a fired `review_trigger`? → `revalidate` mode.

**Escalation is visible, never silent.** Every output opens with a *Scope Declaration* header (declared scope, detected scope, mode, depth profile, assumptions, override line). Comparison → decision: a materially better unlisted option, unstatable criteria, or dependence on an unmade upstream decision produces an *Escalation Notice* instead of an expanded analysis. Project → decision: score ≤4 with one dominant decision emits decision-scope output (resolves the Tier 0 overlap).

### R4. Methodology invariants (normative; identical at every scope)

| # | Invariant | Basis |
|---|---|---|
| I1 | Question-first: every session names the D-NNN it informs and its audience/quality bar | S02 P4; PICO/PCC analogy S10 |
| I2 | Prescriptive WHAT, directional HOW; 5-block prompt; front-loaded single turn; no personas, no hardcoded queries, no search-count floors | S02 P4 |
| I3 | Reversibility-calibrated rigor: **door type sets the evidence bar at any scope** | S02 P2; S09; S27 |
| I4 | Evidence grades A–E + modifiers + verification; recalled ≤ D; no recalled citations under one-way doors | S02 P3 |
| I5 | Falsification: disconfirming search, rejected alternatives with rationale, `review_trigger`; premortem for one-way doors | S02 P8; S19 |
| I6 | Dual-audience artifact: YAML frontmatter + 7-section skeleton; Recommendation isolated from Alternatives | S02 P6 |
| I7 | Regulatory Exposure = 3 hard override (→ Track B) survives every scope | S01 |
| I8 | **NEW — Declared omissions:** every reduced-scope output states what was *not* done and why (scope, sources not searched, checks skipped) | S06; S08; S22 |
| I9 | **NEW — Escalate, don't silently widen:** discovered coupling or better options become a visible Escalation Notice | S02 (Adaptive Checkpoints); S23 (ADM→MDMP hand-off) |

### R5. What adapts per scope

| Dimension | Project | Decision | Comparison |
|---|---|---|---|
| Entry artifact | GENERATOR.md | GENERATOR-DECISION.md | COMPARISON-SESSION.template.md |
| Decomposition | DAG layers 0–2 + sink | ≤3 planned sessions (+≤2 child via Adaptive Checkpoints = ≤5, matching the routing matrix's DEEP 2–5) | 1 session |
| Depth instrument | 8-dim score → Tier 0–3 | Decision Profile → routing outcome (R6) | none; inherits door type + regulatory flag |
| Sessions | 4–30 (Tier 0: 1–3) | 0–3 (+≤2) | 1 |
| Triangulation | staged (P7) | staged; one-way ∧ novel ∧ contested → full | staged; one-way door → at least a critique probe |
| Gate | Phase 0: Track A/B | Track A, or Track B steps 1–8 for the single decision (**INFERENCE**: DAG closure trivial, no FAD) | Track A (two-way); *proposal only* until Track B if one-way |
| Output artifacts | pipeline, registry, sessions, FAD, gate | sessions + one ADR (+ premortem note) | one session file (weights, matrix, sensitivity) + Track A log line or ADR seed |
| Frontmatter additions | none | `scope`, `mode`, `depth_profile`, `parent`, `omitted` | same |
| Session ID (proposal) | `T#-##`, `SYN-##` | `DS-NN` | `CS-NN` |

### R6. Complexity scoring below project scope

Keep the 8-dimension 0–24 score at project scope. **Do not reuse the total at decision scope.** Use the routing matrix as the instrument, plus a Decision Profile (≤6 flags). **INFERENCE** grounded in categorical routing elsewhere (KF4), stakes-gating (KF3) and SM-01/SM-02 [S03].

| Project dimension | Decision-level treatment |
|---|---|
| Domain Novelty + Technical Novelty | **Merge** into one *Novelty* flag (known / novel) — the routing matrix's column; sidesteps the SM-01 double-count risk at this scope |
| Regulatory Exposure | **Inherit** from parent or infer from CONTEXT; keep the =3 hard override (I7) |
| Reversibility | **Promote** to primary axis (door type) with a cost-to-reverse note |
| Investment Horizon | **Inherit**; not scored |
| Coordination Complexity | **Re-anchor → Contestation** (single owner / agreed / disputed); disputed → critique probe or ACH per P7 (also addresses backlog GA-02) |
| Expected Longevity | **Re-anchor →** horizon for `review_trigger`/`review_date` and evidence half-life |
| Integration Complexity | **Re-anchor → Coupling / fan-out** (0 / 1 / ≥2 undecided decisions constrained) |
| *(new)* Evidence availability | Grade A/B sources for *our* workload exist? no → recommend a spike/benchmark rather than more reading (Cynefin-Complex guidance in FRAMEWORK [S02]) |

Routing: two-way ∧ known → **SKIP** (0 sessions); two-way ∧ novel → **SPIKE** (1); one-way ∧ known → **CONFIRM** (1 + ADR + Track B); one-way ∧ novel → **DEEP** (2–3 + ≤2 child + ADR + premortem + human review). Comparison scope: no score; only fit check, door type, optional critique probe.

### R7. P4-compatible effort and boundary language

Anthropic found agents misjudge effort and over-invested in simple queries until explicit class-based scaling rules were added [S30]; Vivechak's P4 bans search-count floors [S02]. Reconcile at the *scope* layer with **qualitative** effort/boundary sentences, no counts: comparison — "This is a comparison-level session: evaluate the named options against the stated workload; record material adjacent concerns and better unlisted options under Discovered Concerns and recommend escalation rather than widening the analysis." Decision scope: same, with the option-discovery allowance. This *bounds* the Bounded Exploration Mandate (OVH-06 [S03]) at small scopes instead of repealing it.

### R8. Rollout, validation gates, reversal triggers

1. **Docs only:** add Scope × Depth and I1–I9 to FRAMEWORK; add the entry-point table to AGENTS/README.
2. **Comparison template:** run ≥3 real comparisons; score against FRAMEWORK §4.3 rubric.
3. **Decision generator:** run ≥3 real decisions, ≥1 open-option and ≥1 one-way.
4. **Review:** measure declared-vs-fit-check disagreement (provisional target ≤20%, uncalibrated), 100% presence of grade+verification fields, rubric pass-rate parity with project sessions, effort.

**Reversal triggers.** Move toward B (auto-detect with declared override) if the fit check overrides >25% of declared scopes or users ask for one entry point. Create `GENERATOR-COMPARISON.md` if template outputs fail the §4.3 rubric measurably more than generator outputs. Adopt assembled prompts (Engine) or merge files if core-block drift recurs in ≥2 releases. Revisit C only after SM-01/SM-02 calibration data exist.

### R9. D-002 hand-off

`status: proposed` · `door_type: two-way` for generator content, **one-way-ish for schema keys and file names** (adopters copy them — keep frontmatter additions to the five listed) · `confidence: medium` · `review_trigger:` "after 3 comparison + 3 decision runs, or on any reversal trigger in R8".

## Alternatives Considered

Evaluated with the Weighted Evaluation Protocol (FRAMEWORK §6.4) — Vivechak applied to its own design question. Criteria derive from the brief's four evaluation dimensions plus its "do not break project-level" constraint.

| Criterion | Weight | Rationale |
|---|---|---|
| Backward compatibility with project-level workflow | 0.20 | Explicit constraint in the brief |
| Methodology integrity (invariants cannot be silently dropped) | 0.25 | Highest: rigor erosion at small scope is the best-documented failure (S09, S15, S21) |
| Usability (right pick by developer *or agent*, low cognitive load) | 0.20 | Brief's first evaluation dimension |
| Implementation & maintenance complexity (incl. drift) | 0.20 | Single maintainer; drift already occurred once (S03) |
| Cross-domain precedent | 0.10 | Evidence weight, but analogical (indirect) |
| Extensibility (modes, Engine compatibility) | 0.05 | Roadmap Phase 5 (S03) |

| Criterion (weight) | **A** three fixed generators | **B** one auto-detecting generator | **C** continuous 1–10 parameter | **A′** declared scope + shared core (recommended) |
|---|---|---|---|---|
| Backward compat (0.20) | **5** — GENERATOR.md untouched; additive | **3** — edits the project generator; regression risk unmeasured | **2** — re-maps tiers, gates and triangulation onto a new scale | **5** — GENERATOR.md behaviourally unchanged |
| Integrity (0.25) | **3** — each file can hard-code invariants, but copies drift (S03; S01≠S02 grade lists) and no hand-off between generators is defined (cf. S23) | **3** — one copy of invariants, but silent under-scoping: LLM agents misjudge effort (S30); light modes erode rigor (S09, S15, S21) | **2** — a linear dial can interpolate invariants away; shortcut impacts are non-linear and method-specific (S09); appraisal must not be cut (S08) | **4** — invariants in shared core, required output fields, fit check; residual copy-drift |
| Usability (0.20) | **3** — explicit contracts, deterministic dispatch; but user must classify and the decision/comparison boundary is fuzzy (R3) | **3** — one entry; misclassification silent; no per-scope input contract | **2** — no defined meaning for "6 vs 7"; effort/timeframe cues are unreliable classifiers (S12) | **4** — declared scope + structural fit check + visible header; one extra step |
| Implementation (0.20) | **2** — three self-contained prompts, duplicated core; drift ×3 | **3** — one file, larger router; test matrix ×3 | **1** — needs depth→sessions/gates mappings and calibration data that do not exist (S03 SM-01/02) | **4** — one thin generator + one template + core check |
| Precedent (0.10) | **4** — per-type input contracts (PICO/PCC, S10/S11), type-specific PRISMA checklists 27 vs 20+2 (S13), named Army methodologies (S23), MADR variants (S28) | **3** — effort rules embedded in a lead-agent prompt (S30); ICD 203 judgment-based application (S22); plan/clarify gates (S31, S36) | **2** — numeric knobs are resource caps or 3–4 levels (S34, S36); no 1–10 methodology scale found | **4** — A's precedent plus confirm-step precedent (S31, S36) |
| Extensibility (0.05) | **2** — each mode = new file | **4** — add a branch | **3** — parameter-agnostic, cannot express modes | **4** — mode modules |
| **Weighted total** | **3.25** | **3.05** | **1.85** | **4.20** |

**Sensitivity check** (each weight ±20%, others renormalised; 12 runs): A 3.16–3.34, B 3.04–3.06, C 1.81–1.89, A′ 4.16–4.24. **Rank order A′ > A > B > C unchanged in 12/12.** Pessimistic A′ (−1 on integrity, usability, implementation) = 3.55, still above A. **Score-sensitive, not weight-sensitive:** if B earned a 4 on usability (auto-detect *with* a declared override), B ties A at 3.25 — hence "A vs B = tie within scoring error."

**Qualitative/quantitative agreement:** they agree. **Bias warnings:** (i) A′ was synthesised by the model that scored it, *after* seeing A/B/C's weaknesses — its scores are prospective hypotheses, not measurements; (ii) all scores are judgment, referenced to evidence but not measured on Vivechak; (iii) a maintainer weighting compatibility above 0.5 would still not rank A above A′ (A′ ties A on that criterion).

**Option notes**
- **A — three fixed generators.** Right instinct (explicit input contracts mirror how medicine fixes question structure per review type), wrong granularity. Comparison scope needs no *generator* (KF6 inference); decision scope does. Three copies of the core multiply a drift problem Vivechak has already had. *Taken forward:* entry contracts; two new artifacts instead of three.
- **B — single auto-detecting generator.** Fits the "open-ended input, AI-driven classification" principles [S01, S02], but those principles are justified for *risk* (users under-estimate it, OVH-02 [S03]); scope is requester intent plus problem structure. Surface-form detection cannot separate the brief's own two "X vs Y" examples. The vendors that automate research still offer a human confirm step (Gemini's optional plan review [S31]; ChatGPT's clarifying questions [S36]). A longer, branching 18.4 KB monolith [S01] also raises an unmeasured instruction-following risk (the 39% figure in FRAMEWORK concerns multi-turn drip-feeding, not prompt length). *Taken forward:* the fit check, as an assistant to — not a substitute for — declaration.
- **C — continuous 1–10 parameter.** Rejected: no methodological precedent (KF4); classification by timeframe/effort is unreliable [S12]; conflicts with Vivechak's own coarse-scale principle (OVH-07) and the uncalibrated-rubric backlog (SM-01/02) [S03]; invites erosion of invariants at low values (S08, S09). A resource cap analogous to `max_tool_calls` [S36] is a legitimate *Engine-layer* control, separate from methodology scope.
- **E — patch Tier 0 into GENERATOR.md** ("if input is a single decision, omit SYN-01/FAD"). Cheapest; keeps one file. Rejected *for now*: edits the project generator (the brief's protected surface), supplies no input contract. **Kept as fallback** if `GENERATOR-DECISION.md` proves redundant (reversal trigger).
- **F — status quo (paste decisions into GENERATOR.md).** The required output structure includes a synthesis session and Phase 0 gate [S01], so single decisions get over-scoped output. INFERENCE.
- **G — Mode-first model** (landscape/evaluate/revalidate as the primary axis, mirroring purpose-based medical typology [S10]). Rejected as primary: adds a choice to the ~90% "evaluate" case; retained as *derived* mode.
- **H — portfolio level above project.** No demand evidence; OVH-05 rejected a comparable expansion on that ground [S03]. Parked.

## Detailed Findings

### 0. Method and grading conventions
- Executed 2026-09-24 against a brief dated 2026-09-23 (frontmatter follows the brief). Single-model research (Claude) with live web tools, ≈33 retrieval attempts (22 searches, 8 successful page fetches, 3 blocked or rejected) across four domains plus Vivechak's README, GENERATOR, FRAMEWORK, ROADMAP and AGENTS files. No triangulation. GitHub directory pages were robots-blocked and PMC full text returned a CAPTCHA, so several medical sources are **abstract- or excerpt-only** (marked in the ledger).
- **Grade conventions (my judgment calls):** peer-reviewed studies and official directives/docs = A (FRAMEWORK §5.1 — GENERATOR disagrees, see DC-2); first-party engineering or practitioner accounts of their own practice, and think-tank empirical reports = B; secondary summaries, blogs, third-party catalogues = D; vendor performance claims = C. `stale` follows FRAMEWORK domain half-lives (foundational methods ~5+ years; AI-tool facts ~6 months); stale-but-real sources keep their grade.
- **Verification:** `fetched` = retrieved live this session; `secondhand` = seen through a summary of a primary source. **No recommendation rests on a recalled claim.**

### 1. Medical evidence synthesis — what changes with review type, and what stays
**Types differ by purpose and question shape, not by size.** Scoping reviews fit gap-finding, mapping a literature, clarifying concepts, or acting as precursors that test inclusion criteria and questions; both they and systematic reviews still require rigorous, transparent methods [S10 | A·corroborated·aging·indirect | fetched]. Library comparisons summarise the split: map vs. synthesise, broad vs. focused question, PCC vs. PICO framing, mandatory risk-of-bias appraisal only for systematic reviews, while explicit search strategy and standardised extraction are common to both [S11 | D·single·fresh·indirect | secondhand]. *The question template is itself the scope declaration.*

**Classification by effort fails.** Grant & Booth analysed 14 review types with the SALSA framework after concluding that titles, descriptions, workflow and timeframe were unreliable cues; they also report that few types have prescribed methods and many are not mutually exclusive [S12 | A·single·aging·indirect | fetched]. Design for overlap (derived fields) rather than hard partitions.

**"Rapid" is a lever set, not a type.** Ottawa's comparison: a full review ≈ 1 year vs. rapid 2 weeks–4 months; several broad questions vs. fewer, clearly specified ones; exhaustive vs. limited searches; in-depth extraction vs. extraction tailored to the commissioner's decision; detailed publication vs. limited grey-literature output [S07 | B·single·fresh·indirect | secondhand]. Cochrane's RRMG offers 26 method recommendations (from a 63-respondent survey) and says best practice is limited by missing evidence on some shortcuts [S05 | A·single·fresh·indirect | fetched (abstract)]. Its series proposes: dual-screen ~20% until agreement is good, then single-screen; single extraction of key data verified by a second person; single risk-of-bias on key outcomes, verified — and warns that unreflective shortcuts can introduce bias [S06 | A·corroborated·fresh·indirect | fetched]. Single screening missed 13% of relevant studies in one study [S07 | B·corroborated·fresh·indirect | secondhand].

| Medical lever | Documented validity impact [S08 \| A·single·aging·indirect \| fetched] | Vivechak analogue | Verdict |
|---|---|---|---|---|---|
| Narrow question / dates / study types | none expected | fewer decisions, options, criteria | **Safe — the primary scope lever** |
| Fewer databases / no grey literature | varies: fewer databases can be harmless with supplementary searching; omitting grey literature may add publication bias (evidence mixed) | narrower landscape; fewer source classes | Acceptable if declared (I8) |
| Single reviewer | may miss up to ~9% of RCTs | single-model default (P7) | Acceptable + critique probe for medium stakes |
| Omit/limit quality assessment | **not recommended** | drop evidence grading | **Never** (I4) |
| Omit protocol / narrative-only synthesis | unknown | skip declared scope / skip matrix | Declare; matrix optional only for two-way |

**Shortcut damage is stakes-dependent.** Simulating rapid methods on 2,512 Cochrane reviews (16,088 studies): PubMed-only searching changed the primary odds ratio by >20% in ~10% of meta-analyses; significance changed in 6.5–38.6% across methods; no systematic bias was seen; the authors accept this for scoping, resource-limited or urgent work and favour comprehensive methods for guidelines and regulatory decisions [S09 | A·single·aging·indirect | fetched]. → Rigor should follow *reversibility/stakes* (I3), exactly as P2 states.

**The reporting core is shared; checklists are per type.** PRISMA 2020 has 27 items; the scoping extension has 20 essential + 2 optional [S13 | A·corroborated·aging·indirect | fetched (abstracts)].

**Mini-HTA is the closest decision-level analogue — and a warning.** It is a form/checklist covering efficacy, safety, cost and organisational consequences, using HTA methodology at hospital level [S16 | B·single·stale·indirect | fetched]; 65% of surveyed managers at one hospital reported 2–15 hours to produce one (excluding literature work) [S15]. A national survey found 60 distinct local forms [S14 | A·single·stale·indirect | fetched (abstract)] and a review of 52 mini-HTAs found literature selection/interpretation often missing and only 25% quantifying clinical effect [S15 | A·single·stale·indirect | fetched (abstract+excerpt)]. The authors found local involvement in producing the assessment likely to matter for use of the results, and judged the format a reasonable balance of depth and resources [S14]. → Light decision tooling gets used *and* quietly loses its evidence core unless the core is structurally mandatory.

**Update mode exists as a first-class idea.** In a sample of 100 systematic reviews the median time to needing an update was 5.5 years, but 23% were outdated within 2 years and 7% at publication [S17 | A·single·stale·indirect | fetched (excerpt)]. Living reviews suit important questions with frequent, impactful new evidence; proposed criteria to start are decision importance, ongoing publication and low certainty, and to retire are conclusive (GRADE moderate/high) evidence, declining relevance, no expected studies, or lost resources [S18 | A·single·fresh·indirect | fetched]. → A `revalidate` mode with an explicit delta scope and retirement conditions.

### 2. Intelligence analysis — what scope does to the analytic framework
- **Standards constant, application scaled.** ICD 203's standards are the common foundation across the community, applied to each product in a manner fitting its purpose, source scope, timeline and customers, with elements free to add mission-specific supplements [S22 | A·corroborated·aging·indirect | fetched]. The nine tradecraft standards cover source quality, uncertainty, information-vs-assumption, alternatives, relevance, argument logic, consistency, accuracy and visuals [S22 (dni.gov; Frontiers listing)].
- **ACH is reserved for weight.** Heuer positions ACH as an eight-step aid for important, controversial issues where an audit trail matters, and notes people are poor at generating the full hypothesis set [S19 | A·single·stale·indirect | fetched (mirror)]. I found **no source that scales ACH itself from tactical to strategic**; lighter needs are met by *different* techniques (e.g., a key-assumptions check) and by three technique families — diagnostic, contrarian, imaginative [S21, S20 | A·corroborated·stale·indirect | fetched (excerpts)].
- **Technique choice keys on problem type and product format.** RAND cites puzzles / mysteries / complexities and formats from current intelligence to estimates as candidate typologies; Marine Corps intelligence built 28 structured techniques for *tactical* situations [S21 | B·single·stale·indirect | fetched]. Tactical adaptation = a different toolkit, not a smaller dial.
- **Time pressure suppresses structure.** Analysts often judge structured techniques too slow for short-deadline products; only 6 of 29 sampled CIA assessments used them explicitly; techniques risk becoming box-checking [S21]. Effectiveness evidence is thin: no systematic evaluation; a 2004 MITRE experiment found ACH reduced confirmation bias only in participants without professional intelligence background [S21].
- **Implication (INFERENCE):** keep full ACH/CONFLICT-RESOLUTION as an *escalation tool for contested one-way doors* (already P7); at comparison scope use the weighted matrix plus explicit key-assumption lines; at decision scope add the premortem for one-way doors. Evidence gap: Heuer & Pherson's technique-selection guidance was not accessed.

### 3. Software design and decision processes — org-level to single-feature
- **Google design docs.** Written where trade-offs are ambiguous; ~10–20 pages for a larger project, a 1–3 page *mini* doc for incremental work with "the same steps, terser"; the write/skip call is a benefit-vs-overhead judgment, with a five-question screen (≥3 yes → write); "implementation manual" docs signal no trade-offs; review weight ranges from a mailing to formal meetings and central review became infeasible as the company scaled [S24 | B·single·aging·indirect | fetched].
- **Spotify.** ADRs record significant decisions; the team reaches them through RFC discussion or meetings, and for large-impact changes (e.g., API-breaking) the ADR is the *final step after the RFC*; small decisions merit ADRs too because they are cheap [S25 | B·corroborated·aging·indirect | fetched]. → **The decision record is scope-invariant; the deliberation artifact scales** — mirrored in Vivechak (sessions → D-NNN).
- **Rust.** RFCs only for "substantial" changes; the definition varies by area with sub-team guidelines; explicit exemptions (reshaping without changing meaning, objectively measurable improvements, developer-invisible changes); un-RFC'd features may be bounced [S26 | A·corroborated·fresh·indirect | fetched].
- **Amazon.** Type 1 (hard to reverse) decisions deserve deliberate process; Type 2 should be fast and light; as organisations grow they misapply heavyweight process to Type 2, causing slowness and risk aversion; most decisions should proceed near ~70% of desired information [S27 | B·corroborated·aging·indirect | secondhand]. → Over-heavy process is a documented failure; a **SKIP** outcome is legitimate output.
- **ADR templates.** MADR-style tooling names four template variants — full, minimal, bare, bare-minimal — along two axes (section count × guidance text) [S28 | D·single·fresh·indirect | secondhand]. Third-party summaries attribute a five-part "definition of done" and an architectural-significance test to Zimmermann, but two summaries expand the acronym differently [S29 | D·contested·fresh·indirect | secondhand] — leads only.
- **Army planning (cross-check).** Doctrine lists ADM, MDMP, TLP, RDSP and problem-solving; the mix follows problem scope, time and staff; MDMP serves staffed headquarters, TLP small units, "similar but not identical," linked by common problem-solving; ADM can frame the problem and an operational approach before MDMP when the situation is complex; MDMP contains an explicit course-of-action *comparison* step; a DTIC handbook notes doctrine does not explain how to integrate conceptual and detailed planning [S23 | A·corroborated·aging·indirect | fetched]. → Same nesting as Vivechak (comparison inside decision inside project) and the same hand-off risk.

### 4. AI research tools — how they handle depth
| Tool | Depth levels | Scope negotiation | Resource control |
|---|---|---|---|
| Gemini Deep Research (API, 04-2026 preview) | Deep Research (speed) vs Max (max comprehensiveness) [S31 \| A·single·fresh·direct \| fetched] | Collaborative planning: agent proposes a plan to review/refine/approve before running [S31] | Secondary reports ≈80 vs ≈160 queries, ≈250k vs ≈900k input tokens, 60-min cap [S32 \| D·single·fresh·direct \| secondhand] |
| Perplexity | Sonar tiers; low/medium/high search context [S33 \| A·single·aging·direct \| fetched] | Multi-model "Model Council" surfaces convergence/divergence (Feb 2026) [S35 \| D·single·fresh·direct \| secondhand] | Deep-research `reasoning_effort` minimal/low/medium/high, async mode [S34 \| A·single·aging·direct \| fetched] |
| OpenAI | Full vs mini deep-research models [S36 \| A·single·aging·direct \| fetched] | ChatGPT asks clarifying questions before researching [S36] | `max_tool_calls` is the primary lever for cost/latency [S36] |
| Anthropic Research | 3 effort classes: fact-finding (1 agent, 3–10 calls), direct comparison (2–4 subagents, 10–15 calls each), complex (>10 subagents) [S30 \| B·corroborated·aging·direct \| fetched] | Rules embedded in the lead agent's prompt; early versions over-invested in simple queries [S30] | Multi-agent runs reportedly ≈15× chat tokens, ≈90% better than single-agent on internal evals [S30 secondary \| D·corroborated \| secondhand] |

**Patterns.** (1) 2–4 discrete levels; none of the sources reviewed used a 1–10 scale. (2) Numeric controls cap *resources*, not methodology. (3) Scope can be negotiated through optional plan/clarify gates rather than silently inferred. (4) Tools answer single questions; Vivechak's scopes decide *which questions, how many runs, which gates* — its README already positions it above these tools [S04 | A·single·fresh·direct | fetched]. (5) Vendor semantics churn (a third-party listing shows three effort levels where the vendor announcement lists four [S34]) → keep effort language vendor-neutral and log the tier used in session metadata. **INFERENCE for mapping:** spike/comparison sessions → standard tier; DEEP sessions → maximum-depth tier.

### 5. Cross-domain support for the invariants
| Invariant | Medical | Intelligence | Software | AI tools |
|---|---|---|---|---|
| I3 Rigor by stakes | guidelines/regulatory need full methods [S09] | ACH for important/controversial [S19] | Type 1/2 [S27]; "substantial" [S26]; trade-off test [S24] | value-vs-cost framing for multi-agent runs [S30] |
| I4 Evidence quality kept | appraisal not to be cut [S08] | source-quality standard [S22] | alternatives with trade-offs [S24] | — |
| I5 Alternatives/falsification | comparator in PICO [S11] | ACH; alternatives standard [S19, S22] | alternatives section [S24]; options in MADR [S28] | — |
| I6 Record vs deliberation | PRISMA reporting [S13] | standards vs product format [S22] | RFC → ADR [S25] | plan → report [S31] |
| I8 Declared omissions | shortcuts must be transparent [S06] | uncertainty and method disclosure [S22] | non-goals [S24] | plan review [S31] |
| I9 Escalation | scoping → systematic review precursor [S10] | — | RFC for substantial change [S26] | — |

### 6. Discovered Concerns (beyond the stated coverage checklist)
- **DC-1 Vocabulary collision.** "Tier 0 (1–3 sessions)" and "decision-level (1–3 sessions)" name overlapping things. Fix: *Tier* = project depth; *Scope* = unit; *Depth profile* = decision-level routing outcome.
- **DC-2 Grade-definition drift.** Peer-reviewed = A in FRAMEWORK §5.1 but B in GENERATOR's list [S01, S02]. Decide and propagate; it changes the grade of every peer-reviewed citation in every session.
- **DC-3 Budget mismatch.** Decision-level 1–3 vs. DEEP 2–5 in the routing matrix. Reconcile as ≤3 planned + ≤2 child (R5).
- **DC-4 No anchoring guard for single-option sessions.** CONFIRM & COMMIT and FAST SPIKE evaluate one option; the ≥2-options rule applies only to comparison sessions (§4.3). Require a baseline (status quo / do-nothing) at decision scope. INFERENCE from [S19, S21] (satisficing; poor hypothesis generation).
- **DC-5 Unbounded exploration at small scope.** "Minimum, not maximum" can inflate a 1-session comparison into a landscape survey; add the escalation clause (R7).
- **DC-6 Weight provenance.** The matrix's weights come from the same model that scores; FRAMEWORK flags criteria-selection bias but not this [S02]. For one-way doors consider human-supplied or human-ratified weights.
- **DC-7 P4 vs effort scaling.** Reconciled qualitatively at the scope layer (R7); numeric caps belong to the Engine.
- **DC-8 Demand evidence.** OVH-05 rejected an expansion for lack of demand evidence [S03]. Here the signal is the maintainer's own hand-authored decision/comparison-style 5-block sessions (this brief is one — INFERENCE); public adoption signals are nil (0 stars, 0 forks at retrieval [S04]). Promote through the R8 validation gates, not on argument alone.

## Open Questions & Risks

### Open questions
1. **Template vs. generator at comparison scope.** Does a filled `COMPARISON-SESSION.template.md` match generator-written prompts on the FRAMEWORK §4.3 rubric (≥3 Grade A/B claims, ≥2 options, failure modes, reversal triggers, disconfirming evidence)? *Falsification test for KF6.*
2. **Auto-detection accuracy.** On a labelled set (~30 mixed inputs, including "X vs Y" phrasings that differ in scope), how often would option B pick correctly, and would a declared override plus the Scope Declaration header make it safe?
3. **Threshold calibration.** Fan-out ≥2, override rate >25%, disagreement ≤20% are placeholders; no source calibrates them (cf. SM-01/SM-02).
4. **Mode visibility.** Should `landscape` and `revalidate` get user-visible entry points now, or stay derived until usage shows demand?
5. **Weight provenance.** For one-way doors, should criteria weights be human-supplied or human-ratified (DC-6)?
6. **Baseline rule.** Is "status quo / do-nothing" always the right baseline for single-option decision sessions (DC-4)?
7. **Schema stickiness.** Are `scope`, `mode`, `depth_profile`, `parent`, `omitted` and the `DS-`/`CS-` ID prefixes the right minimal, forward-compatible set? Bump session `schema_version` to 1.2?
8. **Primary sources not fetched** (findings could shift if they differ): MADR project site; Zimmermann's ADR papers (ecADR expansion is *contested* across summaries); Amazon letters and the 6-pager; Heuer & Pherson's technique-selection guidance; Cochrane Handbook ch. 1–2; Army IPB/ATP 5-0.1 and the RDSP text; Kubernetes KEP, Python PEP 1, Oxide RFDs; ATAM-family architecture evaluation; GRADE Evidence-to-Decision frameworks (which come in scope-tailored variants and may be a direct analogue).
9. **ACH scaling.** Is there any source that scales ACH itself by echelon? None found (§2).
10. **What does D-002 actually say?** Its options and constraints were not visible; reconcile R9 against its text.
11. **Portfolio level.** Any real demand for research above project scope (Option H)?
12. **Engine fit.** Should Phase 5 expose `scope` as a router parameter (B-like) so A′ compiles into it, and where should resource caps (à la `max_tool_calls`) live?

### Risk register
| ID | Risk | Evidence | L / I | Mitigation |
|---|---|---|---|---|
| RK-1 | **Scope laundering** — a project sliced into many comparison sessions to avoid heavier gates; cross-decision effects missed | P1 over-isolation misses cross-cutting trade-offs [S02]; design review's main value is cross-cutting concerns [S24] | M / H | Fit-check Q4 coupling test; `parent` field; prompt to synthesise when ≥3 sessions share a parent |
| RK-2 | **Evidence-bar erosion at light scope** | Mini-HTA omissions [S15]; shortcut impacts [S09]; SAT box-checking [S21] | H / H | I3/I4 as structurally required fields; door type gates the bar; comparison of a one-way door = *proposal only* until Track B |
| RK-3 | **Misclassification** (A: user error; B: silent) | Surface-form cues fail [S12]; brief's two "X vs Y" examples | M / M | Structural fit check; visible Scope Declaration header with override |
| RK-4 | **Drift across entry files** | Propagation incident [S03]; live grade inconsistency [S01, S02]; Army integration gap [S23]; 60 mini-HTA forms [S14] | H / M | One shared core block + diff/checksum step; Engine assembly later |
| RK-5 | **Option-set anchoring** — user-supplied options omit the best one | Poor hypothesis generation; satisficing [S19, S21] | M / H | Baseline option; Discovered Alternatives; Escalation Notice |
| RK-6 | **False precision / self-scored weights** | FRAMEWORK §6.4 caveats [S02]; DC-6 | M / M | Sensitivity check; human-ratified weights for one-way doors |
| RK-7 | **Over-formalising small decisions** | Heavy-process creep [S27]; "implementation manual" docs [S24]; exemptions [S26] | M / M | SKIP outcome and exit ramp are first-class |
| RK-8 | **Ritualisation** — omissions/escalation blocks filled but ignored | SAT box-checking [S21] | M / M | Review declared omissions in retrospectives; count Escalation Notices |
| RK-9 | **Stale quick decisions** | 23% of reviews outdated within 2 years [S17]; FRAMEWORK half-lives [S02] | M / M | Every scope emits `review_trigger`; `revalidate` mode with retirement conditions [S18] |
| RK-10 | **Vendor-tier churn** | Conflicting effort-level listings [S34]; preview status [S31] | H / L | Vendor-neutral effort language; log tool tier in metadata |
| RK-11 | **Vocabulary/number collisions** | DC-1, DC-3 | H / L | Rename and reconcile before publishing |
| RK-12 | **Unvalidated design** — A′ is unbuilt and self-scored | Alternatives Considered bias warnings | H / M | R8 validation gates; reversal triggers |
| RK-13 | **Regulatory override lost at small scope** | Hard override exists only in the project rubric text [S01] | L / H | I7 in the shared core; Decision Profile inherits it |

## Sources & Evidence Ledger

Tag key: `Grade · corroboration · recency · directness | verification`. **36 sources: 24 Grade A (S33: A for features, C for vendor performance claims), 7 B, 5 D, 0 E; 29 fetched (many abstract- or excerpt-only), 7 secondhand, 0 recalled.** Retrieved 2026-09-24. Many `directness: indirect` ratings reflect that cross-domain evidence is analogical to Vivechak.

| ID | Source | URL | Grade · corr · recency · directness | Verification | Limits / notes |
|---|---|---|---|---|---|
| S01 | Vivechak GENERATOR.md v1.1 | https://github.com/bhaskarjha-dev/vivechak/blob/main/GENERATOR.md | A · single · fresh · direct | fetched | 373 lines / 18.4 KB per page metadata |
| S02 | Vivechak FRAMEWORK.md v1.1 | https://github.com/bhaskarjha-dev/vivechak/blob/main/FRAMEWORK.md | A · single · fresh · direct | fetched | Specification; §5.1 grades, §6.4 protocol |
| S03 | Vivechak ROADMAP.md (OVH-01–07; backlog SM/GA/OP) | https://github.com/bhaskarjha-dev/vivechak/blob/main/ROADMAP.md | A · single · fresh · direct | fetched | |
| S04 | Vivechak README.md and AGENTS.md | https://github.com/bhaskarjha-dev/vivechak ; …/blob/main/AGENTS.md | A · corroborated · fresh · direct | fetched | 0 stars / 0 forks at retrieval |
| S05 | Garritty et al. 2021, J Clin Epidemiol 130:13–22 (Cochrane RRMG interim guidance) | https://pubmed.ncbi.nlm.nih.gov/33068715/ | A · single · fresh · indirect | fetched (abstract) | PMC full text blocked by CAPTCHA |
| S06 | Nussbaumer-Streit et al. 2023, BMJ Evid Based Med (RRMG series: team, screening, RoB, extraction) | https://ora.ox.ac.uk/objects/uuid:fc69b831-b4eb-4831-82c2-3d988ecdb263/files/rfq977x213 ; https://www.rti.org/publication/rapid-reviews-methods-series-0 | A · corroborated · fresh · indirect | fetched (abstract/excerpt) | |
| S07 | Cochrane training slides: "How to do a rapid scoping review" (cites Khangura 2012; Gartlehner 2020) and RRMG interim-recommendations slides | https://training.cochrane.org/sites/training.cochrane.org/files/public/uploads/How%20to%20do%20a%20rapid%20scoping%20review.pdf ; …/Part%202%2C%20Rapid%20Reviews%20methods.pdf | B · corroborated · fresh · indirect | secondhand | Summaries of primaries not fetched |
| S08 | Implementation Science 2016, Table 1 "shortcuts" (article s13012-016-0472-9) | https://implementationscience.biomedcentral.com/articles/10.1186/s13012-016-0472-9/tables/1 | A · single · aging · indirect | fetched | Authors not verified in-session |
| S09 | Marshall et al. 2019, J Clin Epidemiol 109:30–41 | https://discovery-pp.ucl.ac.uk/id/eprint/10066545 ; https://pmc.ncbi.nlm.nih.gov/articles/PMC6524137 | A · single · aging · indirect | fetched (abstract/key points) | Simulation on binary outcomes in Cochrane reviews |
| S10 | Munn et al. 2018, BMC Med Res Methodol 18:143 | https://www.ncbi.nlm.nih.gov/pmc/articles/PMC6245623/ | A · corroborated · aging · indirect | fetched (abstract) | |
| S11 | University library comparison guides (scoping vs systematic; PCC/PICO) | https://apu.libguides.com/reviews/overview | D · single · fresh · indirect | secondhand | Summarises JBI/Munn |
| S12 | Grant & Booth 2009, Health Info Libr J 26:91–108 (via abstract and Price 2022 review) | https://salford-repository.worktribe.com/output/1454519/a-typology-of-reviews-an-analysis-of-14-review-types-and-associated-methodologies ; https://journals.library.ualberta.ca/eblip/index.php/EBLIP/article/view/30093 | A · single · aging · indirect | fetched (abstracts) | Newer typologies not checked |
| S13 | PRISMA 2020 statement; Tricco et al. 2018 PRISMA-ScR | https://www.scienceopen.com/document/vid/cfb22a3f-d475-48b1-b4b2-510ac3bcd61f ; https://scholarworks.aub.edu.lb/items/6e701708-5ea2-4543-993f-8d9ab5c55973 | A · corroborated · aging · indirect | fetched (abstracts) | |
| S14 | Ehlers et al. 2006, Int J Technol Assess Health Care (mini-HTA in hospitals) | https://pubmed.ncbi.nlm.nih.gov/16984056/ | A · single · stale · indirect | fetched (abstract) | Danish context |
| S15 | Kidholm et al. 2009, IJTAHC 25:42–48 (quality of mini-HTA) | https://core.ac.uk/works/100971865 | A · single · stale · indirect | fetched (abstract + excerpt) | 2008 data; 52 documents |
| S16 | NOKC/FHI 2010, survey of mini-HTA systems | https://www.fhi.no/en/publ/2010/survey-and-discussion-of-existing-mini-hta-systems-internationally-/ | B · single · stale · indirect | fetched (excerpt) | |
| S17 | Cochrane Handbook, ch. IV (updating; cites Shojania 2007) | https://www.cochrane.org/authors/handbooks-and-manuals/handbook/current/chapter-iv | A · single · stale · indirect | fetched (excerpt) | Sample: reviews published 1995–2005 |
| S18 | Murad et al. 2023, BMJ Evid Based Med 28:348 (living-review retirement triggers) | https://ebm.bmj.com/content/28/5/348 | A · single · fresh · indirect | fetched (abstract) | |
| S19 | Heuer, *Psychology of Intelligence Analysis* (1999), ch. 8 — mirror copy | https://corpora.tika.apache.org/base/docs/govdocs1/269/269871.html | A · single · stale · indirect | fetched (excerpt) | Mirror on a corpus site; CIA page not fetched |
| S20 | US Government, *A Tradecraft Primer* (2009) | https://www.cia.gov/resources/csi/static/955180a45afe3f5013772c313b16face/Tradecraft-Primer-apr09.pdf | A · single · stale · indirect | fetched (excerpt) | |
| S21 | RAND RR-1408, Artner, Girven & Bruce (2016) | https://apps.dtic.mil/sti/tr/pdf/AD1024447.pdf | B · single · stale · indirect | fetched (full text) | Pilot; authors caution against generalising |
| S22 | ICD 203 (2015 revision); dni.gov objectivity page; Frontiers 2018 listing of the nine standards | https://www.dni.gov/files/documents/ICD/ICD-203.pdf ; https://www.dni.gov/index.php/how-we-work/objectivity | A · corroborated · aging · indirect | fetched (excerpts) | Later revisions not checked |
| S23 | US Army planning doctrine excerpt (hosted copy); Lightning Press MDMP and TLP summaries; DTIC handbooks | https://www.mnd.gov.tw/File/44288 ; https://www.thelightningpress.com/troop-leading-procedures-tlp/ ; https://www.thelightningpress.com/military-decision-making-process-mdmp/ ; https://apps.dtic.mil/sti/pdfs/AD1055096.pdf ; https://apps.dtic.mil/sti/tr/pdf/AD1018227.pdf | A · corroborated · aging · indirect | fetched (excerpts) | Doctrine edition not verified; secondary summaries for MDMP/TLP wording |
| S24 | Malte Ubl, "Design Docs at Google" (2020) | https://www.industrialempathy.com/posts/design-docs-at-google/ | B · single · aging · indirect | fetched | First-party practitioner account |
| S25 | Spotify Engineering, "When Should I Write an Architecture Decision Record" (2020); InfoQ summary | https://engineering.atspotify.com/2020/4/when-should-i-write-an-architecture-decision-record ; https://www.infoq.com/news/2020/04/architecture-decision-records/ | B · corroborated · aging · indirect | fetched (excerpts) | |
| S26 | rust-lang/rfcs README | https://github.com/rust-lang/rfcs | A · corroborated · fresh · indirect | fetched (excerpt) | Corroborated by forks' copies |
| S27 | Bezos shareholder letters (2015/2016) as reproduced | https://bytepawn.com/the-best-parts-of-jeff-bezos-invent-and-wander.html ; https://rufuspollock.com/post/jeff-bezos-fast-high-quality-decision-making ; https://blog.matt-rickard.com/p/high-velocity-decision-making | B · corroborated · aging · indirect | secondhand | Primary letters not fetched |
| S28 | `adrs-core` crate source documenting MADR 4.0 template variants | https://docs.rs/adrs-core/0.7.6/src/adrs_core/template.rs.html | D · single · fresh · indirect | secondhand | MADR project site not fetched |
| S29 | Third-party ADR skill summaries citing Zimmermann (ecADR; ASR test) | https://claudeskills.info/skills/ncklrs/startup-os-skills/adr/ ; https://claudeskills.info/skills/existential-birds/beagle/adr-writing/ | D · contested · fresh · indirect | secondhand | Acronym expansion differs; leads only |
| S30 | Anthropic Engineering, "How we built our multi-agent research system" (2025); ByteByteGo and aktagon summaries for the ≈15× / ≈90% figures | https://www.anthropic.com/engineering/multi-agent-research-system ; https://blog.bytebytego.com/p/how-anthropic-built-a-multi-agent | B · corroborated · aging · direct | fetched (excerpts) | Architecture/effort rules only; no pricing or limits used |
| S31 | Google Gemini API docs — Deep Research agent | https://ai.google.dev/gemini-api/docs/deep-research | A · single · fresh · direct | fetched (excerpts) | Preview features |
| S32 | Blog coverage of Deep Research vs Max figures | https://pasqualepillitteri.it/en/news/1191/google-deep-research-max-gemini-3-1-pro-ai-agents | D · single · fresh · direct | secondhand | Not verified against Google docs |
| S33 | Perplexity blog — Sonar search modes | https://www.perplexity.ai/hub/blog/new-sonar-search-modes-outperform-openai-in-cost-and-performance | A (features) / C (performance) · single · aging · direct | fetched (excerpt) | |
| S34 | Perplexity community announcement — async mode and `reasoning_effort`; contrast: third-party listing | https://community.perplexity.ai/t/sonar-deep-research-async-mode-and-reasoning-effort-now-live/4736 ; https://www.llmreference.com/model/sonar-deep-research/perplexity-api | A · single · aging · direct | fetched (excerpt) | Third-party listing shows 3 levels vs vendor's 4 |
| S35 | AI Weekly — Perplexity guide (Model Council) | https://aiweekly.co/learning-ai/generative-ai/how-to-use-perplexity | D · single · fresh · direct | secondhand | |
| S36 | OpenAI developer guide — deep research (ja-JP page) | https://developers.openai.com/ja-JP/api/docs/guides/deep-research | A · single · aging · direct | fetched (excerpt) | Excerpt read in Japanese |

**Retrieved but not used as evidence:** a third-party Perplexity API tutorial whose model-naming claims contradict the vendor announcement (unreliable); a dictionary-style page asserting the "single most common" PRISMA-ScR compliance error (unsupported).
**Coverage against the brief's checklist:** cross-domain analysis (§1–4) · option matrix (Alternatives Considered) · recommendation (R1–R9) · invariants (R4, §5) · adaptations (R5) · complexity scoring (R6) · risks (register) · inline grades (throughout) · open questions (above).
