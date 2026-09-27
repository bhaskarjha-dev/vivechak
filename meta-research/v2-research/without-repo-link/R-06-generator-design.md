---
id: R-06
title: "Decision-Level and Comparison-Level Generator Design"
date: 2026-09-23
status: draft
topic: generator-design
informs_decisions: [D-002]
---

# R-06 — Decision-Level and Comparison-Level Generator Design

> **Read first — provenance of this session.** This session was scoped to build on `GENERATOR.md` (to be read at `d:\dev\pro\vivechak\GENERATOR.md`) and on R-03's multi-scope methodology findings. Neither was available: the file path is local to the requester's own machine, and nothing matching it was uploaded here (`/mnt/user-data/uploads` checked — empty). The brief's R-03 context block was left as the unfilled `[Inject key findings — ...]` template placeholder rather than actual content — that looks like a failed injection step in whatever assembles these session briefs, worth checking on your end rather than an intentional omission.
>
> Rather than inventing a fictitious R-03 and citing it as evidence, every claim below carries a provenance tag: **[BRIEF]** = stated directly in the R-06 brief itself, treated as primary source; **[PROPOSAL]** = this session's own design reasoning, not sourced from GENERATOR.md or R-03; **[UNVERIFIED]** = standing in for missing content, flagged for confirmation. Treat this whole document as a v0.1 draft pending that confirmation — see Open Questions for exactly what to check.

## Research Question

How should Vivechak's project-level methodology — implemented today in the single `GENERATOR.md` — decompose into two lighter generators, **GENERATOR-DECISION.md** (one decision → 1–3 research sessions + a proposed ADR) and **GENERATOR-COMPARISON.md** (fixed options + criteria → one Weighted Evaluation Protocol session), such that the methodology's non-negotiable properties (evidence grading, bounded exploration, structured falsification, self-containment) survive the drop in scope — and is a three-level Project / Decision / Comparison model actually the right decomposition, or does R-03 argue for something else?

## Key Findings

**F1 [process note].** The two inputs this session depends on most — `GENERATOR.md`'s actual text and R-03's actual findings — were unavailable (see provenance note above). This governs how every finding below should be read: F2 and F3 are grounded in the brief itself; the rest are this session's design proposals standing in for missing source material.

**F2 [BRIEF, high confidence].** The brief specifies the methodology core directly, independent of whatever R-03 says: prompts are "methodology-compliant" only if they implement evidence grading, bounded exploration, and structured falsification, and "self-contained" only if they carry no reference to external docs — so any frontier AI with web search can run them zero-shot. This session treats these four properties as invariant across every scope level (table below).

**F3 [BRIEF, high confidence].** The brief's two generator descriptions already imply a working scope model, independent of R-03's actual text:
- `GENERATOR.md` (existing): full vision → many-session pipeline + decision registry.
- `GENERATOR-DECISION.md` (this session): one decision context → 1–3 sessions + one proposed ADR.
- `GENERATOR-COMPARISON.md` (this session): fixed options + criteria → one WEP session, no ADR.

This session adopts Project ⊃ Decision ⊃ Comparison as its working hypothesis for "R-03's scope model recommendation," since it's the only scope model actually evidenced by an available source. **This is the single highest-priority thing to verify once R-03 is available** (Open Questions, #1).

**F4 [PROPOSAL].** The cleanest Decision-vs-Comparison boundary isn't session count, it's whether **the option set is already closed**. Comparison-level assumes 2–5 named options are already settled and just need rigorous, falsification-tested scoring. Decision-level exists for cases where that narrowing hasn't happened, or where the output needs to be a durable record (an ADR) rather than a one-off score. Test example 3 below shows this boundary is fuzzier in practice than in principle.

**F5 [PROPOSAL].** WEP is referenced in the brief as an already-named, presumably already-defined protocol — almost certainly specified inside `GENERATOR.md` [UNVERIFIED]. Self-containment means `GENERATOR-COMPARISON.md` can't just say "apply WEP" — it has to inline a complete, compact restatement. The draft below reconstructs a standard weighted-criteria structure (weighted criteria → graded per-option scores → sensitivity check → mandatory falsification pass → recommendation) consistent with the brief's other requirements. Treat this as a reconstruction, not a citation.

**F6 [PROPOSAL].** Complexity scoring can't mean the same thing at every level. At project scope [UNVERIFIED] it presumably sizes a whole roadmap. At decision scope, the only open question is session count (1–3) and split strategy for one decision. At comparison scope, sizing disappears entirely — always exactly one session, because supplying fixed options and criteria is precisely what makes sizing unnecessary.

**F7 [PROPOSAL].** The decision-level ADR is more useful drafted as a **falsifiable hypothesis** — a tentative, clearly-marked-Proposed decision the generator commits to up front — than as a passive TBD stub. The generated sessions can then be pointed specifically at falsifying that hypothesis, applying "structured falsification" at the decision level, not just within each session. This is this session's own synthesis, not sourced from GENERATOR.md.

### Methodology invariants (every scope level)

- **Evidence grading** — same A–D scale everywhere; every material claim carries a grade.
- **Bounded exploration** — always present; only the search budget changes.
- **Structured falsification** — always present; only its target changes (a claim, a leading option, or a proposed decision).
- **Self-containment** — every generated artifact is runnable with zero external context.
- **Placeholder IDs** — a generator never guesses a real `R-XX`/`D-XXX`; it emits a placeholder for the registry to resolve.

### Scope-dependent adaptations

| | `GENERATOR.md` (project) [UNVERIFIED] | `GENERATOR-DECISION.md` | `GENERATOR-COMPARISON.md` |
|---|---|---|---|
| Input | full project vision | Decision Context (7 fields) | options (2–5) + criteria + context |
| Complexity scoring answers | how many decisions & sessions, across a roadmap | how many sessions (1–3), and what split, for one decision | nothing — always 1 session |
| Option discovery | implicit, across many decisions | only when option clarity is low | never — options are a required input |
| Output | session pipeline + decision registry | 1–3 sessions + one Proposed ADR (falsifiable hypothesis) | 1 WEP session, no ADR |
| Search budget / session | [UNVERIFIED] | ~8–15 (illustrative) | ~5–10 (illustrative — no discovery phase) |

**Project-specific, correctly omitted below:** full-vision intake and stakeholder framing, multi-decision sequencing and cross-decision dependency mapping, decision-registry construction, and roadmap-level complexity scoring. These only make sense when sizing a whole program of work.

## GENERATOR-DECISION.md Draft

*Target ~5KB (measured: 5.3KB, 779 words). Voice: addressed to whichever frontier AI will run it — its job is to produce research sessions and a draft ADR, not to research the decision itself.*

```markdown
# Vivechak — Decision-Level Research Generator

You are a research planner for Vivechak, a structured-research framework for technical decisions. You do not answer the decision yourself. Given the **Decision Context** below, produce two things: (1) one to three self-contained research session prompts, and (2) one proposed ADR (Architecture Decision Record), drafted as a falsifiable hypothesis. Nothing you produce is final — it is scaffolding for research that hasn't happened yet.

## Required input: Decision Context

If any field below is missing or too vague to act on, ask for it before proceeding. Do not guess and continue.

- **Decision** — one sentence: what is being decided.
- **Why now** — the trigger. Why this needs deciding at this project stage.
- **Project stage** — early-exploration / active-build / pre-launch / scaling / maintenance.
- **Constraints** — hard limits: team skills/size, budget, timeline, existing stack, compliance. Anything that rules an option out regardless of merit.
- **Stakes / reversibility** — how costly, and how hard to reverse, if this turns out wrong.
- **Known candidate options** (optional) — leave blank if the option space itself is unclear. That is a legitimate reason to be here instead of at comparison-level.
- **Related decisions** (optional) — IDs of related decisions already on record.

## Step 1 — Size the decision

Judge three factors:
1. **Option clarity** — are the real candidates already known, or does the option space need discovery first?
2. **Investigation axes** — how many genuinely independent things determine the answer (technical feasibility, ongoing cost/ops burden, ecosystem trajectory, team fit, compliance)? Count only axes where evidence on one tells you nothing about another.
3. **Stakes / reversibility** — from the input.

Then choose a split and state it explicitly — don't just emit a number:
- **1 session** — options are known, and the decision turns on one dominant axis. (Close in shape to comparison-level; choose this generator anyway when the output needs to become an ADR, or the framing is genuinely "should we do this at all.")
- **2 sessions**, split one of two ways: (a) *discover-then-evaluate* — session 1 maps the real option space, session 2 evaluates the narrowed set; or (b) *parallel axes* — two axes are independent enough that mixing them in one session would blur the evidence.
- **3 sessions** — medium-high+ stakes, and (option discovery is needed OR three or more independent axes exist). Reserve the third session for dedicated risk/failure-mode investigation and structured falsification of the leading option — not more general research.

## Step 2 — Write the session prompt(s)

Each session prompt must be complete and independently runnable by any frontier AI with web search — no reference back to this generator or to any other Vivechak document. Each one must contain, inline:

- **Scope** — the specific question(s) this session, and only this session, answers. State what it explicitly does not need to resolve.
- **Bounded exploration** — a modest search budget (illustratively 8–15 targeted searches for a full session) and an explicit stop condition: stop once new searches stop changing the emerging picture. If the topic proves broader than expected, say so rather than silently expanding.
- **Evidence grading** (inline verbatim in every session):
  - **A** — primary source, official docs, or reproducible benchmark data, reasonably current.
  - **B** — credible secondary source with a named, accountable author or organization.
  - **C** — convergent consensus across multiple independent sources, no primary confirmation.
  - **D** — single, unverified, or promotional source. Usable only if flagged D, ideally after seeking a second source.
  Every material claim in the findings carries its grade.
- **Structured falsification** — before finalizing any finding, explicitly search for the strongest case against it (or for the alternative), and state whether it survives. Accumulating only confirming evidence does not satisfy this.
- **Required output** — graded findings, an explicit confidence level, and an "open questions this session could not resolve" list.

## Step 3 — Draft the proposed ADR

Write it as a falsifiable hypothesis, not a placeholder:

- **ID** — `D-XXX` (unassigned; the registry assigns the real number).
- **Status** — Proposed.
- **Context** — the Decision Context, restated in ADR form.
- **Hypothesis decision** — your own best tentative call from what's known now, stated plainly, not hedged into meaninglessness.
- **What would falsify this** — the specific findings that would change it. Point at least one generated session at testing this directly.
- **Options considered** — from the input, plus any discovery sessions are expected to surface.
- **Consequences** — TBD. Do not fabricate these before research exists.
- **Confidence & evidence basis** — TBD, to be filled from session grades once sessions return.
- **Review trigger** — what would make this worth revisiting later.

## Output

Emit the sessions first, clearly delimited and labeled Session 1 / 2 / 3, then the ADR. Keep them independently copy-pasteable — do not merge into prose.
```

**Design notes**
- *Carries over from the core (F2):* evidence grading (verbatim A–D), bounded exploration (budget + explicit stop condition), structured falsification (now at two levels — per-session, and per F7, at the ADR/hypothesis level), self-containment, placeholder IDs.
- *Simplified:* complexity scoring drops from roadmap-sizing to a 3-factor / 3-outcome call for one decision (Step 1); input drops from a project vision to a 7-field schema.
- *Notable choice:* the missing-field instruction in the Decision Context section ("ask, don't guess") is load-bearing — test example 3 shows why, for realistically under-specified inputs.
- *Byte budget:* measured at 5.3KB against the ~5KB target. The four non-negotiable methodology elements plus the sizing logic and ADR template consume nearly all of it; nothing was cut from them to make room (Quality Risks).

## GENERATOR-COMPARISON.md Draft

*Target ~2KB (measured: 2.3KB, 341 words) — tight enough that every line has to earn its place.*

```markdown
# Vivechak — Comparison-Level Research Generator

You are a research planner for Vivechak. Given the **fixed options and criteria** below, produce exactly one self-contained research session prompt applying the Weighted Evaluation Protocol (WEP). You do not run the research yourself, and you do not draft an ADR — this generator's output feeds a decision; it is not a decision record.

**Use this generator** when 2–5 named options and the criteria to judge them by are already settled, and rigorous evidence-graded scoring is what's needed. **Use GENERATOR-DECISION.md instead** if the option space itself is unclear, if the real question is "should we do this at all," or if the outcome needs to become a recorded decision.

## Required input

- **Options** — 2–5 named candidates.
- **Criteria** — what matters, and relative importance if known. If weights aren't given, propose defaults and say so explicitly rather than assuming equal weight silently.
- **Context** — why this comparison, and any constraint that eliminates an option outright regardless of score.

## The session prompt you generate must instruct the researcher to:

1. Confirm or set criteria weights (sum to 100%); state the reasoning if you set them.
2. Score each option per criterion (1–5), citing evidence for every score and grading it: **A** primary/official/benchmark, **B** credible secondary, **C** convergent community consensus, **D** single/unverified (flag D explicitly).
3. Compute weighted totals and run a sensitivity check: would the ranking flip under a ±20% shift on the top-weighted criterion? Report the answer.
4. Run structured falsification: build the strongest honest case for the lowest-scoring *viable* option before finalizing a recommendation.
5. Stay bounded — a modest search budget (illustratively 5–10 targeted searches, narrower than decision-level since there is no discovery phase); stop once new evidence stops moving the scores.
6. Flag, but not unilaterally evaluate, any option or criterion the research surfaces that's missing from the input — evaluating it would break bounded exploration.
7. Close with a plain-language recommendation, its confidence level, and one explicit line: this is a scored recommendation, not a decision record — if adopted, promote it to an ADR.
```

**Design notes**
- *Carries over:* the same A–D grading scale (one line instead of a block), bounded exploration (narrower budget — no discovery to fund), structured falsification (reframed as "steelman the lowest-scoring viable option"), self-containment (WEP inlined in full, not cited by name — F5).
- *Omitted entirely:* no complexity-scoring step (always 1 session, per F6), no ADR (comparison output is an input to a decision, not a decision record).
- *Notable choice:* step 6 lets the researcher flag a missing option/criterion without unilaterally evaluating it — preserves bounded exploration even when the input turns out to be incomplete.
- *Byte budget:* measured at 2.3KB, slightly over the ~2KB target — WEP's seven scoring steps plus the four invariants leave almost no slack (Quality Risks).

## Test Results

Design-level dry runs — how each generator's logic would apply to a real input, worked through by hand. Not empirical output from actually executing a generated session prompt against a live research AI.

| Example | Generator | Shape | Key insight |
|---|---|---|---|
| 1. MCP vs. custom plugin system | Decision-level | 3 sessions | "Custom" isn't a concrete option yet — needs discovery, not scoring |
| 2. PostgreSQL vs. CockroachDB, write-heavy SaaS | Comparison-level | 1 WEP session | Clean fit; falsification step exists specifically to catch "write-heavy" framing bias |
| 3. Adopt SSR for a Next.js app | Ambiguous — 1 session either way | 1 session | Input is under-specified; the boundary between the two generators genuinely blurs here |

**Example 1 — "Should Vivechak use MCP or build a custom plugin system?"**
Decision-level, not comparison-level: "custom plugin system" isn't an existing, scoreable thing, it's a design space, so evaluating it fairly requires first sketching what it would look like — that's option discovery, out of scope for the comparison generator by design. Sizing: option clarity medium (MCP is well-specified; "custom" isn't), at least three independent axes (what MCP actually supports/limits for Vivechak's needs; what a custom system would cost to build and maintain; ecosystem/adoption trajectory), stakes medium-high (an integration-layer choice that's expensive to reverse later) → **3 sessions**: (1) MCP capability/limits for this use case, (2) custom-system cost/design sketch, (3) ecosystem trajectory + structured falsification of whichever way sessions 1–2 lean. Worth flagging: this session's own Integration Notes (below) refer to "the MCP server" as if Vivechak already has one — if that's existing infrastructure, a real run of this generator should have it captured in the Decision Context's Constraints/Related-decisions fields rather than re-litigated from scratch. That's the schema working as intended, not a flaw in the example.

**Example 2 — "PostgreSQL vs. CockroachDB for a write-heavy SaaS workload"**
Comparison-level, cleanly: two concrete, well-documented, existing options; no discovery needed. Criteria the generated session would confirm or default to: write throughput/latency at target scale (high weight, given the "write-heavy" framing), horizontal scaling story, operational complexity and team familiarity, consistency guarantees under concurrent writes, managed-hosting cost, ecosystem/tooling maturity. This example is a good stress test of the falsification step specifically: "write-heavy" framing invites scoring CockroachDB's distributed story favorably by default, and the mandatory steelman-the-lower-scorer pass exists to ask, e.g., whether the actual write volume ever approaches what a well-tuned single Postgres primary can handle before distributed overhead is worth it. This session deliberately doesn't resolve the actual Postgres-vs-CockroachDB question — running the generated session for real, against Vivechak's actual numbers, is what that's for.

**Example 3 — "Should we adopt server-side rendering for our Next.js app?"**
The interesting case. As posed, it's missing every required Decision Context field (why now, constraints, stage, stakes) — a real run should trigger the missing-field check and ask before proceeding. It's also not cleanly a two-option comparison: Next.js's rendering strategies (SSR/SSG/ISR/CSR/RSC) are a closed, well-documented set, which pulls toward comparison-level, but the question is framed as adopt/don't-adopt rather than "score these named options," which pulls toward decision-level. If reasonable defaults are filled in (currently CSR, seeing SEO/load-time pain, pre-launch, moderate/reversible stakes since Next.js allows mixed strategies per route), sizing lands at just 1 session either way — meaning the choice of generator here is really about whether the framework designer wants an ADR out of it, not about the research differing. That's a genuine soft spot in the three-level model, not a contrived edge case (Open Questions, #6).

## Integration Notes

No MCP server tool surface or schema was provided to this session, so these are proposed contracts the generators' output should satisfy, not a specific call sequence. [PROPOSAL] throughout.

- **Placeholder IDs, not guesses.** Neither generator can know the registry's next free `R-XX`/`D-XXX` at generation time, so both emit placeholders; the server (or whatever holds registry state) assigns real numbers at commit time.
- **A thin machine-parseable header per artifact.** Each generated session prompt and the ADR draft should carry minimal metadata — `scope: decision|comparison`, `parent_decision`, `status: pending|proposed`, `session_count` — so the server can register them without a human re-typing structure by hand.
- **Comparison output still needs a link back.** It produces no ADR, but it still needs enough metadata for the server to connect "we scored this" to "we later decided this" if a human promotes the result — otherwise that audit trail breaks silently.
- **Open dispatch question [UNVERIFIED].** Whether the server can itself hand generated session prompts to a research AI, or whether that stays a manual paste-into-a-fresh-conversation step, is unknown here. The contract above holds either way.
- **Drift risk.** WEP and the A–D grading scale are now duplicated across at least three files for self-containment. A natural role for the server is owning the canonical definitions and linting (or regenerating) the inlined copies whenever they change — flagged again in Open Questions, since this session can't confirm the server is positioned to do that.

## Quality Risks at Reduced Scope

| Risk | Mitigation designed in |
|---|---|
| Narrower context misses existing project-level constraints/decisions | Decision Context requires Constraints + Related-decisions fields; generator flags apparent conflicts rather than ignoring them |
| Fewer sessions → anchoring on the first plausible answer | Structured falsification stays mandatory at every session count, including 1; elevated to the ADR/hypothesis level too (F7) |
| Evidence-grading rigor gets cut under "just answer fast" pressure | The grading scale is compact (one block or one line) and was never on the list of things trimmed to hit the byte targets |
| ADR from 1–3 sessions overstates confidence vs. a full pipeline | Status defaults to Proposed, never Accepted; Consequences/Confidence stay TBD until real findings exist; hypothesis framing makes the tentativeness explicit |
| Comparison output mistaken for a final decision (no ADR exists to signal otherwise) | Generated session output must close with an explicit "not a decision record" line |
| No discovery phase at comparison-level → bad input produces a confident, wrong-shaped answer | Researcher may flag a missing option/criterion but may not unilaterally expand scope to cover it |
| Self-containment forces WEP + grading-scale duplication across files → drift | Proposed server-owned canonical copy + lint/regen step (Integration Notes); unresolved without server details |
| This entire document rests on an unverified reconstruction of `GENERATOR.md`/R-03 | Provenance tags throughout; explicit recommendation to re-run or revise once source docs are available — the biggest risk here is in this session, not in the design itself |

## Open Questions

1. **[Highest priority]** Does R-03 actually recommend the three-level Project/Decision/Comparison model this session assumed (F3), or something else — e.g., two levels, or comparison as a mode of decision-level rather than a separate generator? Needs the real R-03 document.
2. What are `GENERATOR.md`'s actual evidence-grading scale, bounded-exploration mechanic, and WEP definition? This session proposed reconstructions for all three (F2, F5, F6) that need checking against — and likely replacing with — the originals.
3. Does the falsifiable-hypothesis ADR framing (F7) match how `GENERATOR.md`/the decision registry expects proposed ADRs to look, or does it expect a neutral TBD stub instead?
4. Who or what turns session findings back into a completed (non-Proposed) ADR — a synthesis step in this same generator, a separate one, or something the MCP server automates? This draft is silent on that step.
5. Should `GENERATOR-COMPARISON.md` output ever auto-promote into an ADR, or is that always a deliberate human call? This draft assumed always-manual.
6. Example 3 surfaced a real boundary fuzziness for "adopt/don't-adopt a known, bounded option set" questions — should intake itself detect this shape and suggest the other generator, rather than leaving the choice entirely to the framework designer?
7. What's the MCP server's actual tool surface — can it dispatch generated child sessions itself, or does that stay a manual step indefinitely?
8. Given self-containment forces duplication of WEP and the grading scale across at least three files, what mechanism (if any) is intended to keep them in sync as `GENERATOR.md` evolves?
