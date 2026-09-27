---
id: R-03
title: "Multi-Scope Research Methodology: Extending Vivechak Below Project Level"
date: 2026-09-23
status: draft
topic: multi-scope
informs_decisions: [D-002]
---

# R-03 — Multi-Scope Research Methodology

## Research Question

How should Vivechak extend its structured-research methodology below project level — to a **decision level** (1–3 sessions → grounded ADR) and a **comparison level** (1 session → weighted evaluation matrix) — without breaking the existing project-level pipeline (`GENERATOR.md` → 4–30 sessions → Founding Architecture Document), and without letting the smaller scopes quietly lose the rigor that makes the methodology worth using in the first place?

Four sub-questions frame the investigation:

1. Across medical evidence synthesis, intelligence analysis, and software RFC processes, what do structured methodologies hold constant across scope, and what do they deliberately vary?
2. Of three implementation patterns — (A) three fixed generators, (B) one scope-detecting generator, (C) a continuous 1–10 depth parameter — which is best supported by how scope variation actually works in the domains that have already solved this problem?
3. How should complexity scoring change *shape*, not just range, between project-level and decision-level?
4. Does the evidence support the three-level model as given, or does it reveal a level the brief hasn't considered?

## Key Findings

**⚠ Prominent finding — the three-level model may be missing a level.** Medical evidence synthesis draws a hard line between reviews that *answer a question* (systematic, rapid) and reviews that *map a space* (scoping) — and treats the second as a different task, not a lighter version of the first [A, S4–S6]. Vivechak's decision- and comparison-level scopes both assume the option set is already known ("MCP or custom plugin system," "Postgres vs CockroachDB"). Neither one covers "I don't yet know what my options are." See Finding 5 and the Recommendation for how to handle this without adding a fourth generator prematurely.

1. **Every domain studied varies rigor, not the underlying method.** The core epistemic moves — frame the question, gather evidence systematically, weigh alternatives explicitly, document reasoning, state confidence — are identical from the most exhaustive systematic review to the fastest rapid review. What changes is how much of each step happens and how formally it's recorded [A, S1–S3].

2. **Every domain uses named, discrete tiers — never a continuous dial.** Medicine: systematic / rapid / scoping. Software: ADR / RFC, or a *mini design doc* / full design doc. AI research products: Quick Search / Pro Search / Research / Labs (Perplexity), or standard / Max agent (Gemini). Nobody ships "rigor: 6.5" [A, S1–S3, S10–S17]. This is direct evidence against Option C.

3. **Scope is gated by stakes, reversibility, and blast radius — not by how big the topic sounds.** Cochrane triggers a rapid review by urgency and decision-need, not topic size [A, S2]. The Rust project routes only substantial changes through its RFC process; ordinary changes use the normal PR flow regardless of how much code they touch [A, S13]. Real engineering orgs gate ADR-vs-RFC by the same logic — a small internal choice needs only an ADR, while a broad, contested proposal goes through an RFC first, which then produces the ADR as its record [B, S20]. A one-line, hard-to-reverse decision should get more rigor than a large, cheaply-reversible one.

4. **Rigor reduction must be transparent and named, not silent.** Cochrane's rapid-review guidance requires every shortcut — single-reviewer screening, narrower search, omitted risk-of-bias steps — to be explicitly declared, not just applied [A, S2]. This is the single strongest methodology-integrity finding in this research: a decision-level or comparison-level Vivechak output should say what it *didn't* do, the same way a rapid review's methods section does.

5. **Scoping reviews answer a different kind of question, not a smaller one.** They map concepts, definitions, and the shape of a field rather than appraising evidence toward an answer, use a different reporting standard (PRISMA-ScR vs. PRISMA 2020), and explicitly skip risk-of-bias appraisal because it isn't the point [A, S4–S6]. This is the basis for the prominent finding above.

6. **Real AI research products converge on a hybrid of Option A and Option B — never a pure form of either.** Gemini's Deep Research agent ships two named model tiers (`deep-research-preview` vs. `deep-research-max-preview`) *and* a collaborative-planning mode that shows its research plan for approval before running [A, S15]. Perplexity explicitly renamed "Deep Research" to "Research" specifically to make its four tiers (Quick → Pro → Research → Labs) read as a clear, named progression [A, S17]. No product studied does silent, tier-less auto-detection (pure B) or a continuous depth slider (pure C).

7. **Every domain keeps a small, fixed set of tradecraft standards that survive every scope change.** ICD 203 names nine analytic tradecraft standards that apply to *all* IC products regardless of scale or urgency [A, S9]. PRISMA's checklist family (2020, ScR, P) all trace back to the same transparency purpose [A, S1, S4]. Nygard's ADR format insists every consequence be listed, not just the favorable ones, regardless of how small the decision is [A, S10]. This is the methodology core — see Finding 7 in Detailed Findings.

8. **Software architecture already contains an internal escalation pattern that is a near-exact precedent for Vivechak's hybrid.** Google's own design-doc culture has a *mini design doc* mode for small-but-non-obvious changes that keeps the same section structure in compressed form [B, S11]. Several real engineering orgs run a formal ADR ⇄ RFC decision tree keyed on impact and consensus [B, S20]. This is direct precedent for recommending A-shaped fixed artifacts with a B-shaped router in front of them, rather than inventing a novel pattern.

**Note on evidence grades.** Grades below are adapted from the logic of GRADE's certainty-of-evidence system [A, S19], simplified for a design document rather than a clinical one: **[A] Strong** — primary or standard-setting source, typically corroborated independently. **[B] Moderate** — reputable secondary source or a pattern corroborated across several independent practitioner sources without one canonical primary document. **[C] Limited** — single source, community-level source, or an inference by analogy rather than direct evidence. Grades are cited inline as `[grade, source-id]`; the full ledger is at the end.

## Recommendation

**Adopt Option A for the outer shape, with a thin, shared Option B-style intake step in front of it. Reject Option C outright.**

Concretely:

1. **Keep three fixed, named generators** — `GENERATOR.md` (project), `GENERATOR-DECISION.md` (decision), `GENERATOR-COMPARISON.md` (comparison) — each with its own fixed input contract, session-count logic, and output artifact. This mirrors every precedent found: PRISMA has separate checklists per review type rather than one adaptive checklist [A, S1, S4]; software orgs keep ADRs and RFCs as separate artifact types rather than one shape-shifting document [B, S20]. Fixed, named artifacts are auditable — a reader can check a decision-level output against `GENERATOR-DECISION.md`'s contract the same way a reviewer checks a rapid review against Cochrane's rapid-review guidance, rather than having to reconstruct what a single adaptive prompt decided to do on the fly.

2. **Factor out the invariant core into one shared file** (see Finding 7) that all three generators `include` rather than restate. This is the direct fix for the biggest structural risk of Option A: drift between the three generators as the methodology evolves. Concretely, this shared core should own: the evidence-grading convention, the "state what you didn't do" transparency rule, the requirement to document rejected alternatives, and the citation/sourcing standard. Individual generators own: input format, session count logic, search breadth, and output template.

3. **Add one small, shared "scope intake" step — not a fourth generator.** Before any of the three generators commits to a pipeline, run a short classification pass that (a) states which of the three scopes it thinks the input matches and why, (b) flags if the option set looks undefined rather than merely uncompared (the scoping-review gap), and (c) lets the user or calling agent confirm or override before work begins. This mirrors Gemini's collaborative-planning pattern [A, S15] and Cochrane's practice of agreeing the review type with stakeholders before starting [A, S2] — assistive, visible, and overridable, never silently authoritative. This is the "B" half of the hybrid, and it is the only place Option B's spirit belongs: as a front door, not as the engine.

4. **Do not build Option C.** No domain studied uses a continuous depth parameter, and the reason is structural, not just habitual: output *format* doesn't vary continuously either. A Founding Architecture Document, an ADR, and a weighted matrix are qualitatively different artifacts; there is no well-formed "6.5" between an ADR and a matrix. A continuous parameter would also triple the testing surface (a range of behaviors instead of three) for a flexibility gain that doesn't correspond to how the adjacent domains — or the AI research products most similar to Vivechak — actually behave.

5. **Handle the scoping-review-shaped gap inside decision-level, not as a new generator yet.** Give `GENERATOR-DECISION.md` a built-in escape valve: if the scope-intake step finds the option set isn't actually known, session 1 becomes a short landscape-mapping pass (map the plausible options) before the decision frame is set, rather than forcing a premature ADR. Track how often this fires. If it fires often enough to be a real workload rather than an edge case, that's the evidence to promote it to a fourth generator later — this keeps the system from adding complexity ahead of demonstrated need, matching how scoping reviews themselves grew out of teams needing to decide *whether* a systematic review was even feasible before committing to one [A, S4].

This recommendation is evaluated against the alternatives in the matrix in Detailed Findings §6.

## Alternatives Considered

### Option A (pure) — Three Fixed Generators, No Shared Intake

**For:** Matches the medical (PRISMA / PRISMA-ScR) and software (ADR / RFC) precedent almost exactly — named, independently versioned artifacts with clear, auditable contracts [A, S1, S4, S10, S13]. Lowest risk that a smaller tier silently inherits unreviewed logic from the project-level generator, because each file is self-contained. Simplest to test in isolation.

**Against:** Tier selection is left entirely to the user or calling agent. No domain studied leaves this step unassisted — Cochrane's guidance opens with stakeholder agreement on review type [A, S2]; real ADR-vs-RFC practice is commonly captured as an explicit decision tree, not left to intuition [B, S20]. Pure A also has no defense against the three files drifting apart over time unless something outside the pattern itself enforces consistency — which is exactly why the recommendation adds a shared core file rather than accepting pure A as-is.

### Option B (pure) — One Adaptive Generator

**For:** Single artifact to maintain. Matches the surface appeal of a "smart" tool that just figures out what's needed.

**Against:** No domain researched — medical, intelligence, software, or the current AI-research-tool market — ships a fully silent, single, self-adapting tool for this class of problem. Even the most adaptive product studied, Gemini Deep Research, always surfaces its plan for explicit approval before running long research rather than silently deciding depth [A, S15]. A single adaptive Vivechak generator would make the rigor contract un-auditable: a reader of a decision-level output couldn't tell whether steps were skipped deliberately (as Cochrane requires rapid reviews to declare) or simply skipped by the model in that run, without diffing prompts. That opacity is precisely what Cochrane's guidance warns rapid reviews to avoid [A, S2].

### Option C — Continuous Scope Parameter (1–10)

**For:** Maximally flexible in principle; avoids drawing arbitrary tier boundaries.

**Against:** Zero precedent across all four domains researched. Every one converges on a small number of qualitatively different, named tiers rather than a smooth dial [A, S1–S3, S10–S17]. A numeric parameter creates a false-precision problem — a user has to decide whether their question is a "4" or a "6," a distinction that maps to nothing real, mirroring how nobody using Perplexity picks "Research depth 6"; they pick Quick, Pro, Research, or Labs [A, S17]. Output format doesn't scale continuously either (ADR vs. matrix vs. Founding Architecture Document are discrete shapes), so a continuous input parameter would still have to snap to one of a few discrete output templates somewhere in the pipeline — meaning Option C doesn't actually avoid building tiers, it just hides them one layer down while adding interface and testing complexity on top.

## Detailed Findings

### 1. Medical research: full systematic reviews, rapid reviews, and scoping reviews

PRISMA 2020 is the current reporting standard for systematic reviews: a 27-item checklist covering rationale, eligibility criteria, search strategy, study selection, risk-of-bias assessment, and synthesis, designed so that a review's methodology is transparent and replicable [A, S1]. A full Cochrane systematic review pairs this with comprehensive multi-database searching, dual independent screening, formal risk-of-bias tools, and GRADE certainty ratings — commonly over a thousand hours of work [B, S1, S19].

Cochrane's Rapid Reviews Methods Group defines a rapid review as a deliberately abbreviated systematic review — one that speeds up the standard process by simplifying or dropping specific methodological steps so stakeholders get usable evidence on a resource-efficient timeline [A, S2]. Its interim guidance is unusually specific about *which* methods bend and which don't, stage by stage [A, S2]:

- **Question-setting:** still requires a protocol with PICOS and inclusion/exclusion criteria, agreed with stakeholders — this step is not cut.
- **Search:** always search Cochrane CENTRAL, MEDLINE, and Embase; specialized databases are capped at 1–2 or dropped if time is short; language and date limits are permitted *with justification*.
- **Screening:** dual-reviewer screening is required for only a sample (≥20%) of abstracts, calibrated first with a shared pilot exercise; one reviewer handles the rest, with a second reviewer checking only the excluded set.
- **Data extraction and risk of bias:** single-reviewer with second-reviewer verification, not independent duplication; risk-of-bias ratings are limited to the most important outcomes.
- **Synthesis:** narrative by default; meta-analysis only if it was already going to be appropriate for a full review.
- **Timeline:** explicitly "tailored," ranging from one week to six months, chosen based on urgency and available resources — this is the closest existing analogue to a session-count decision.

Critically, every one of these is a *named, declared* reduction, not a silent one, and the guidance is explicit that this remains a live methodological risk rather than a settled question: several of the shortcuts it recommends still lack a solid evidence base of their own, and the guidance says so openly rather than presenting them as cost-free [A, S2] — Cochrane itself treats rigor reduction as a monitored risk, not a solved problem.

Scoping reviews are not a third point on the same rigor scale. They exist to map the breadth and shape of the evidence on a topic, clarify concepts, or test whether a future systematic review is even feasible — replacing PICO with PCC (Population, Concept, Context) and using a separate reporting standard, PRISMA-ScR, rather than a shortened PRISMA 2020 [A, S4–S6]. They typically skip formal risk-of-bias appraisal entirely: judging each study's individual quality isn't the point of a mapping exercise, so it's treated as optional rather than core [B, S4]. Guidance in this space also names a specific, on-point risk directly relevant to Vivechak: picking the wrong checklist for a review's actual purpose is described as the single most common compliance error in the field [B, S4] — direct precedent for the misuse risk discussed in Open Questions & Risks.

**What stays the same:** protocol-first design, systematic (not ad hoc) search, transparent reporting against a named checklist, explicit inclusion/exclusion criteria, at least two people involved in screening. **What changes:** database count, reviewer independence, risk-of-bias depth, synthesis method, and — for scoping reviews — the fundamental aim, moving from *answering* to *mapping*.

### 2. Intelligence analysis: ACH and technique selection across scale

Analysis of Competing Hypotheses (ACH), developed by Richards Heuer at the CIA in the 1970s, is an eight-step method: list every plausible (and mutually exclusive) hypothesis, list significant evidence for and against each, build a matrix scoring each evidence item's consistency with each hypothesis, refine, and favor the hypothesis with the *least* evidence against it rather than the most evidence for it — deliberately structured to counter confirmation bias [A, S7]. The method closes by requiring analysts to report the relative likelihood of *all* hypotheses considered, not just the winner, and to name in advance what future evidence would change the conclusion [A, S7].

ACH is one of 66 techniques catalogued in Heuer and Pherson's *Structured Analytic Techniques for Intelligence Analysis*, organized into six families (organizing data, exploration, diagnostic reasoning, reframing, foresight, decision support), each technique documented with explicit guidance on when to use it [A, S8]. This matters for Vivechak's design question specifically: intelligence analysis's answer to how you scale rigor is a **toolkit of named techniques selected for the situation**, not one technique that dials itself up or down. The catalogue includes a lightweight "Getting Started Checklist" for framing any project before deeper work begins [A, S8] — functionally similar to the scope-intake step recommended above.

Underneath all of it sits ICD 203, the Office of the Director of National Intelligence's Analytic Standards, which require every analytic product — regardless of urgency, classification, or scale — to (among nine tradecraft standards) properly describe source quality, explicitly express uncertainty, distinguish assumptions from underlying intelligence, and incorporate analysis of alternatives [A, S9]. These standards don't scale down for a fast tactical assessment; they're the fixed floor a fast assessment still has to clear. This is the cleanest cross-domain example of an *invariant core* separated from *which technique you pick*, and it maps directly onto the recommendation to factor Vivechak's evidence-grading and alternatives-documentation rules into one shared file all three generators inherit.

**What stays the same:** the nine ICD 203 tradecraft standards, the discipline of stating alternatives and uncertainty explicitly. **What changes:** which of 66 named techniques gets used, driven by available time, stakes, and how contested the issue is [A, S8, S9].

### 3. Software architecture: ADRs, design docs, 6-pagers, and RFCs at different scales

Michael Nygard's 2011 Architecture Decision Record format is deliberately minimal: Title, Status (proposed / accepted / deprecated / superseded), Context, Decision, and Consequences — with an explicit instruction to list *all* consequences, "not just the positive ones" [A, S10]. One ADR covers one decision, kept to roughly two pages, stored as a short markdown file next to the code it affects, and never edited after acceptance — a later change gets a new ADR that supersedes the old one, preserving history rather than overwriting it [A, S10]. This immutable-log pattern is itself a useful invariant for Vivechak's decision-level output: a decision-level session shouldn't edit a prior ADR in place if the decision changes; it should produce a new one that supersedes it.

Google's design-doc culture, documented publicly by a former Google engineer, uses a consistent structure — Context and Scope, **Goals and Non-Goals**, the actual design, Alternatives Considered, Cross-Cutting Concerns — for genuinely significant decisions [A, S11]. Its definition of a non-goal is unusually precise and directly reusable for Vivechak: a non-goal is not simply a negated goal (a rule like requiring a system not to crash) but something that could reasonably have been a goal and was deliberately left out instead — the source's own illustration is a database design doc that explicitly marks full ACID compliance as a conscious non-goal, so anyone reading it knows it was considered and set aside rather than overlooked [A, S11] — structurally identical to how this brief's own SCOPE section separates what's in scope from what's out. Notably, Google's own culture already contains an internal scope adaptation: a *mini design doc* mode for changes that are small but not obvious, keeping the same reasoning structure in compressed form rather than switching to a different one [B, S11] — direct precedent for Vivechak keeping the same evidence discipline across scope levels while compressing its expression.

Amazon's six-page narrative memo sits at the opposite end: no slides, no bullet points, full prose, read silently for up to 30 minutes at the start of a meeting before any discussion happens [A, S12]. It is reserved for higher-stakes, leadership-level decisions specifically because the format is expensive to produce and consume — direct evidence that format should track stakes, not just topic size, matching Finding 3.

At the process-gating level, the Rust project requires its formal RFC process only for substantial changes to the language, compiler, or ecosystem; ordinary changes, including large ones, go through the normal pull-request workflow [A, S13]. Spotify's engineering blog documents individual teams adopting ADRs at team level, with decisions often surfacing through informal RFC-style discussion or regular engineering meetings before being written down [A, S14]. Multiple independent organizational practices converge on an explicit ADR-vs-RFC decision tree keyed on impact and existing consensus: a small, patterned decision gets only an ADR; a large, contested, or precedent-setting one goes through an RFC first, which then *produces* an ADR as its durable record [B, S20]. One practitioner framing draws the line simply: an ADR records a decision that's already been made and is ready to act on, while an RFC exists to gather feedback on a proposal that may still need real time and debate before it becomes a decision at all [C, S20].

**What stays the same:** documenting context, decision, and consequences (including negative ones); explicitly naming alternatives and non-goals; storing decisions as durable, version-controlled artifacts rather than ephemeral chat or slides. **What changes:** document length, review formality, number of required stakeholders, and — as Amazon's format shows — even the writing medium itself, gated by stakes and reversibility rather than topic size [A, S10–S13].

### 4. AI research tools: how Gemini Deep Research and Perplexity handle depth

Google's Gemini Deep Research agent begins every task by generating a multi-step research plan and — in its collaborative-planning mode — showing that plan for the user to edit or approve before any searching happens; only once the plan is explicitly confirmed does it execute [A, S15, S16]. The underlying API ships two named agent tiers rather than a depth number: `deep-research-preview`, tuned for speed and streaming to a UI, and `deep-research-max-preview`, tuned for maximum comprehensiveness [A, S15]. A typical moderate-depth run uses roughly 80 search queries and reads dozens to over a hundred sources before synthesizing a cited, multi-page report [A, S15].

Perplexity's product structure is the clearest evidence in this entire research set for named, discrete tiers over a continuous dial. As of 2026 it runs four modes in an explicit progression: **Quick Search** (fast, small source set, for simple lookups), **Pro Search** (multi-step reasoning over more sources), **Research** — deliberately renamed from "Deep Research" specifically, in Perplexity's own words, "to reflect its central role between Perplexity's Search and Lab modes" [A, S17] — and **Labs**, which spends ten-plus minutes producing full deliverables (reports, spreadsheets, dashboards, mini-apps) rather than a single report [A, S17; B, S18]. The rename itself is telling: Perplexity had *already* built adaptive depth-selection internally, and still chose to expose it to users as a small set of named, orderable tiers rather than hide it behind one adaptive mode — the same shape this report recommends for Vivechak.

Neither product ships pure silent auto-detection (Option B) or a numeric depth slider (Option C). Both ship named tiers (Option A's shape) with a thin, visible, overridable planning or selection layer in front (Option B's spirit) — precisely the hybrid recommended above, arrived at independently by two different companies solving an adjacent problem.

**What stays the same:** citations back to sources, a visible plan or mode choice before deep work begins. **What changes:** number of sources read, wall-clock time invested, and output shape — a short answer, a cited report, or a full multi-artifact deliverable [A, S15–S17].

### 5. Cross-domain pattern: what actually gates scope

Pulling the four domains together, one variable recurs as the real gate, under different names in each field: **medicine** calls it urgency and decision-need [A, S2]; **intelligence** calls it stakes and contention [A, S8, S9]; **software** calls it blast radius and reversibility cost [A, S11, S13]; **AI research products** call it query complexity and desired deliverable [A, S15, S17]. None of them gate primarily on "how big does this topic sound." A single, hard-to-reverse technology choice with real disagreement behind it (a plausible "decision-level" Vivechak input) can legitimately warrant more rigor than a sprawling but low-stakes, easily-revisited project. Vivechak's complexity scoring should encode this directly — see §9.

A second recurring pattern: **the rigor floor is fixed; only the rigor ceiling moves.** ICD 203's nine standards, PRISMA's transparency purpose, and Nygard's rule that every consequence gets listed, not just the convenient ones, don't get cheaper at smaller scope — they're binary, met-or-not requirements. What scales is search breadth, reviewer count, document length, and time invested. This is the strongest single argument for factoring Vivechak's invariant core out of all three generators rather than letting each one restate (and potentially under-state) it independently.

### 6. Design option evaluation matrix

| Criterion | (A) Three fixed generators | (B) One adaptive generator | (C) Continuous 1–10 parameter | **Recommended: A + thin B intake** |
|---|---|---|---|---|
| **Usability — picking the right tool** | Weak alone; no domain leaves this unassisted [A, S2, S20] | Strong in principle, but no real product ships this silently [A, S15, S17] | Weak — forces a numeric guess with no real-world anchor [A, S17] | Strong — matches Gemini's plan-then-approve and Perplexity's named-tier UX exactly [A, S15, S17] |
| **Methodology integrity** | Strong if the shared core is factored out; weak if each file restates rules independently [A, S1, S9, S10] | Weakest — rigor contract becomes un-auditable, exactly what Cochrane's transparency rule warns against [A, S2] | Weak — no discrete output shape to audit against [A, S1, S10] | Strong — fixed, auditable per-tier contract plus a shared, enforced invariant core |
| **Implementation complexity** | Low–medium; three independent surfaces to build and test | Medium–high; one system that must correctly branch on ambiguous signals with no visible checkpoint | High; ten behaviors to validate, plus a second discretization step to map the number onto one of a few real output shapes anyway | Low–medium; three surfaces plus one small, shared classifier step |
| **Precedent across domains researched** | Strong (PRISMA family, ADR/RFC split) [A, S1, S4, S10, S13] | None found in pure form | None found in any form | Strong (Gemini, Perplexity both converge here independently) [A, S15, S17] |
| **Drift / maintenance risk** | Real, unless mitigated (see recommendation) | Low (one file) but at the cost of auditability | Real — a wide behavior space is harder to keep consistent than three fixed ones | Low — invariant core is centralized; only tier-specific logic is duplicated by design |

### 7. Methodology invariants — what stays constant across all scopes

Synthesized from every domain above, these are the rules that should live in Vivechak's shared core, inherited by `GENERATOR.md`, `GENERATOR-DECISION.md`, and `GENERATOR-COMPARISON.md` alike, regardless of session count:

- **The question is framed explicitly before work begins**, including what's deliberately out of scope — Google's Goals/Non-Goals split [A, S11], Cochrane's protocol-first rule [A, S2].
- **Alternatives are documented, not just the chosen option** — ACH's requirement to report likelihood across *all* hypotheses [A, S7]; design docs' mandatory Alternatives Considered section [A, S11]; MADR-style ADRs' "Considered Options" [B, S21].
- **Consequences and uncertainty are stated explicitly, including unfavorable ones** — Nygard's ADR format requires every consequence to be listed, not just the favorable ones [A, S10]; ICD 203's requirement to "properly express and explain uncertainties" [A, S9].
- **Source and evidence quality is characterized, not just cited** — ICD 203's first tradecraft standard [A, S9]; PRISMA's risk-of-bias and certainty requirements [A, S1].
- **Any rigor that was skipped is named, with a reason** — Cochrane's core transparency rule for rapid reviews [A, S2]. This is the one most likely to be silently dropped as Vivechak shrinks in scope, and the one worth protecting most deliberately.
- **The output is a durable, versioned artifact, not an ephemeral answer** — ADRs live in the repo [A, S10]; Cochrane rapid reviews are still registered and published [A, S2].

### 8. Scope-dependent adaptations — what changes per level

| Dimension | Project-level (`GENERATOR.md`, existing) | Decision-level (`GENERATOR-DECISION.md`, new) | Comparison-level (`GENERATOR-COMPARISON.md`, new) |
|---|---|---|---|
| Input | Full project vision document | A framed question with a known-but-uncompared option set (e.g., "MCP or custom plugin system?") | Two or more named, already-identified options plus the axis they're being weighed on |
| Session count | 4–30, scored (existing logic) | 1–3, gated by a small factor set (§9) | 1, fixed |
| Evidence breadth | Broad, multi-source, spans the whole system | Narrow, focused on the specific fork in the road | Narrowest — criteria-driven, not exploratory |
| Alternatives considered | Many, at the architecture level | The named option plus at least one credible alternative, explicitly | Exactly the options given, scored against explicit weighted criteria |
| Output artifact | Founding Architecture Document | Grounded ADR (Nygard shape + explicit alternatives/evidence, MADR-influenced) [A, S10; B, S21] | Weighted evaluation matrix (ACH-matrix shape with added weights) [A, S7, S8] |
| Primary audience | Whole team / future maintainers | Whoever owns the specific decision | Whoever needs to pick, fast |
| Review / approval step | Implied by pipeline length itself | Scope-intake confirmation (§Recommendation, item 3) | Scope-intake confirmation, lightweight — mostly confirming the criteria/weights before scoring |

### 9. Complexity scoring: decision-level vs. project-level

**Caveat, stated plainly:** this research was not given Vivechak v1.1's actual project-level scoring internals, so what follows is a design proposal grounded in cross-domain evidence, not a description of the existing system. Confirming it against the real `GENERATOR.md` scoring logic is a prerequisite before implementation — see Open Questions.

By analogy with every domain studied, project-level scoring is presumably multi-factor and wide-range (subsystem count, integration points, stakeholder count, unknowns) mapping onto the given 4–30 session range — closer to how a full systematic review's scope is set by PICOS breadth and the number of comparators [A, S2].

Decision-level scoring should not be a smaller version of that same wide scorer; it should be a narrow **gate**, because the problem itself is narrower — this mirrors how Cochrane's rapid-review timeline (1 week to 6 months) is set by a handful of factors, not the same multi-database, multi-outcome scoring used for a full review [A, S2]. A concrete, testable proposal, keeping the given 1–3 range:

- **Base: 1 session.**
- **+1 if reversibility cost is high** — expensive migration, data lock-in, contractual or compliance switching cost. Mirrors blast-radius gating in the Rust RFC process and ADR-vs-RFC practice [A, S13; B, S20].
- **+1 if there is active, named disagreement** about the right answer among stakeholders. Mirrors SAT selection guidance, where contested issues warrant more rigorous technique [A, S8].
- **+1 if the option set itself is not yet well-defined**, triggering the landscape-mapping sub-step from the Recommendation rather than a premature ADR. This is where the scoping-review gap gets absorbed without a new generator.
- **Cap at 3**, per the given range.

Worked example, using the brief's own sample question: *"Should we use MCP or build a custom plugin system?"* — base 1, likely +1 for reversibility (a plugin architecture is costly to reverse once built on), possibly +1 if the team is actively split — plausibly lands at 2 sessions, which feels like the right order of magnitude for that question without needing to invent a numeric formula the framework designer would have to defend line-by-line.

**Comparison-level scoring is not a session-count problem at all — session count is fixed at 1 by definition.** The free variable moves to a different dimension entirely: how many weighted criteria go into the matrix. This is a genuinely different kind of "complexity" than either of the other two levels, closer to how ACH's matrix depth is set by how many pieces of evidence are diagnostic, not by a session budget [A, S7]. As a design heuristic rather than a citation-backed rule [C, self], matrices stay legible up to roughly six to eight criteria in practice; beyond that, the output stops being a quick decision aid and starts resembling a decision-level analysis in disguise — which is itself a useful trigger: if a comparison-level session keeps wanting more than ~8 criteria, that's a signal to escalate to decision-level rather than to keep adding matrix rows.

## Open Questions & Risks

### Risks of scope expansion

- **Dilution.** If the invariant core (§7) isn't literally factored into one shared file that all three generators include, each generator will restate the rules from memory over time, and restatements drift. Cochrane's own guidance names this exact failure mode for rapid reviews: shortcuts taken without evidence behind them, evaluated only after the fact [A, S2].
- **Misuse — wrong-tier selection.** Running a comparison-level matrix on a question that's actually high-stakes and contested (under-scoping), or running a full project-level pipeline on a simple binary choice (over-scoping and wasting the person's time). The medical-literature finding that using the wrong checklist for the wrong review type is "the single most common compliance error" in that field [B, S4] is a direct, evidenced precedent for this exact risk in Vivechak.
- **Silent quality degradation.** The failure mode Cochrane worries about most: a shortcut becomes routine and its cost is never measured [A, S2]. Mitigation is the transparency invariant in §7 — every skipped step must be named in the output itself, so degradation is at least visible to the reader even if not fully solved.
- **Generator drift.** Three files maintained independently will diverge in tone, citation standards, or rigor unless the shared-core mechanism from the Recommendation is actually enforced (e.g., in review/CI, not just in intent).
- **Escalation handling is currently undesigned.** What happens when a decision-level session, mid-flight, reveals project-level complexity (the "this is bigger than we thought" moment)? None of the domains researched hand this off perfectly — a scoping review escalating to a systematic review still means starting a new, separate review [B, S4] — but Vivechak could do better than that by design. Worth a dedicated design pass, not covered further here.
- **Floor risk.** Using Vivechak at all for a question too trivial to need structured research — below even comparison-level — wastes session budget and could make the tool feel like overkill, eroding trust in it for the cases where it genuinely earns its cost.

### Open questions

- What does `GENERATOR.md` v1.1's actual complexity scoring weight today? The decision-level proposal in §9 is built by cross-domain analogy, not by inspecting the existing scorer, and should be reconciled with it before implementation.
- Where exactly is the boundary between decision-level and comparison-level when a "decision" happens to have only two candidate options? The brief's own examples (MCP vs. custom plugin; Postgres vs. CockroachDB) sit close enough to this line that the scope-intake step needs a clear tie-breaker, likely reversibility cost per Finding 3.
- Should the scoping-review-shaped gap (prominent finding) become a fourth generator, or stay a built-in sub-step of decision-level indefinitely? §Recommendation proposes deferring this and deciding from usage data.
- Should the scope-intake classification be visible and user-editable by default (Gemini's collaborative-planning pattern [A, S15]), or only surfaced when its confidence is low?
- What's the actual mechanism — not just the intent — that keeps the shared invariant core in sync across three generator files as the methodology evolves? A written rule isn't self-enforcing.
- How should a comparison-level session handle discovering, mid-session, that the two named options aren't actually comparable on the criteria given (e.g., missing data for one option)? Not addressed by any domain studied directly.

## Sources & Evidence Ledger

| # | Source | Domain | Grade | Link |
|---|---|---|---|---|
| S1 | Page MJ, et al., "The PRISMA 2020 statement: an updated guideline for reporting systematic reviews," BMJ / PMC | Medical | A | https://www.ncbi.nlm.nih.gov/pmc/articles/PMC8005924/ |
| S2 | Cochrane Rapid Reviews Methods Group, "Cochrane Rapid Reviews: Interim Guidance," March 2020 | Medical | A | https://methods.cochrane.org/sites/methods.cochrane.org.rapidreviews/files/uploads/cochrane_rr_-_guidance-23mar2020-final.pdf |
| S3 | Garritty C, et al., "Cochrane Rapid Reviews Methods Group offers evidence-informed guidance to conduct rapid reviews," J Clin Epidemiol 2021 | Medical | A | https://pmc.ncbi.nlm.nih.gov/articles/PMC7557165 |
| S4 | CASRAI, "Systematic Review vs Scoping Review: PRISMA Guide" and "PRISMA-ScR" entry | Medical | B | https://casrai.org/guides/systematic-review-vs-scoping-review-prisma-guide |
| S5 | Northwestern Galter Health Sciences Library, "Systematic or Scoping Review? Choosing the Best Review for You" | Medical | B | https://galter.northwestern.edu/News/systematic-or-scoping-review-choosing-the-best-review-for-you |
| S6 | Binghamton University Libraries, "Scoping Reviews" subject guide | Medical | B | https://libraryguides.binghamton.edu/scopingreview |
| S7 | Heuer RJ Jr., *Psychology of Intelligence Analysis*, CIA Center for the Study of Intelligence, 1999 | Intelligence | A | https://archive.org/details/PsychologyOfIntelligenceAnalysis |
| S8 | Heuer RJ, Pherson RH, *Structured Analytic Techniques for Intelligence Analysis*, 3rd ed., SAGE/CQ Press | Intelligence | A | https://us.sagepub.com/en-us/nam/structured-analytic-techniques-for-intelligence-analysis/book255432 |
| S9 | Office of the Director of National Intelligence, ICD 203 "Analytic Standards" | Intelligence | A | https://www.dni.gov/index.php/how-we-work/objectivity ; primary text: https://irp.fas.org/dni/icd/icd-203.pdf |
| S10 | Nygard M., "Documenting Architecture Decisions" (2011), as reproduced by Agile Alliance and ADR implementers | Software | A | https://www.agilealliance.org/?p=8043624 |
| S11 | Ubl M., "Design Docs at Google" | Software | A | https://www.industrialempathy.com/posts/design-docs-at-google/ |
| S12 | Bezos J., Amazon shareholder letters (1997, 2017), as quoted by Office Watch and UsingData | Software/Org process | A | https://office-watch.com/2018/powerpoint-banned-at-amazon-jeff-bezos/ ; https://usingdata.com/usingdata/2020/8/11/six-page-narratives |
| S13 | rust-lang/rfcs, "Rust RFCs" README | Software | A | https://github.com/rust-lang/rfcs |
| S14 | Blake J., "When Should I Write an Architecture Decision Record," Spotify Engineering | Software | A | https://engineering.atspotify.com/2020/4/when-should-i-write-an-architecture-decision-record |
| S15 | Google AI for Developers, "Gemini Deep Research Agent" / developer guide | AI tools | A | https://ai.google.dev/gemini-api/docs/interactions/deep-research ; https://aistudio.google.com/learn/deep-research-developer-guide |
| S16 | Google, "Gemini Deep Research — your personal research assistant" | AI tools | A | https://gemini.google.com/overview/deep-research/ |
| S17 | Perplexity AI, "Introducing Perplexity Labs" | AI tools | A | https://www.perplexity.ai/hub/blog/introducing-perplexity-labs |
| S18 | Independent 2026 product-comparison writeups on Perplexity's mode tiers (Turion.ai, Nimble Cyber, ToolChase, AI Weekly) | AI tools | B | https://turion.ai/blog/perplexity-ai-complete-guide/ ; https://nimblecyber.com/how-to-use-perplexity-ai-for-research-2026/ ; https://toolchase.com/blog/perplexity-ai-review/ |
| S19 | BMJ Best Practice, "What is GRADE?"; Wikipedia, "GRADE approach" | Medical / evidence grading | A/B | https://bestpractice.bmj.com/info/evidence/learn-ebm/what-is-grade/ |
| S20 | Practitioner corroboration on ADR-vs-RFC tiering: candost.blog; SciLifeLab architecture governance doc; AM Digital platform playbook | Software | B/C | https://candost.blog/adrs-rfcs-differences-when-which/ ; https://github.com/SciLifeLab/architecture/pull/3 ; https://playbook.platformdev.amdigital.co.uk/Ways-of-Working/Toolkit/Architectural-Design/Documenting-Decision/ |
| S21 | MADR (Markdown Architectural Decision Records) format, referenced via ADR practitioner sources as the dominant current ADR template with an explicit "Considered Options" section | Software | B | https://scrapbox.io/iki-iki/ADR |
