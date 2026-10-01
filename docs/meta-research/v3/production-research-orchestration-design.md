# Research orchestration: from structurally valid to evidence-bound

**Labels.** [E] established in published work or standard practice (cited from memory; verify before citing). [J] my design judgment. [O] a tendency I observe in models like me; I can't inspect my own mechanisms, so treat these as hypotheses your eval data should overrule. Every numeric threshold is a starting point, not a finding. Mechanisms carry **Cost** and **Breaks if** tags.

## Thesis

The gap between "structurally valid" and "excellent" is the gap between output that describes research and output that is evidence of research. The fixes bind output to artifacts the model cannot author. Three rules [J]:

1. **Process facts belong to the harness; content judgments belong to the model.** Inline A-E grades currently sit on the wrong side of that line.
2. **Whoever makes a claim never grades it.** Verification reads stored source text, in a separate context.
3. **"Done" is a predicate over open questions,** not a count of searches or a coverage checklist.

## 0. Architecture

| Model, in separate contexts | Harness, deterministic code |
|---|---|
| Planner: typed question tree; requests a stop | Gates: code evaluates the stop request |
| Reader: one document, no tools; extracts spans | Source store: snapshot, hash, dates |
| Verifier: reads snapshots, never memory | Claim ledger: records, provenance, conflicts |
| Synthesizer: writes from ledger IDs | Citation renderer: URLs come from IDs |

Consequences: at L2+ the model never types a URL, it cites harness-issued source IDs; the reader sees one document with no tools, which also blunts prompt injection [E: Greshake et al. 2023]; the planner requests a stop and code decides; the verifier reads snapshots, not memory.

Integration levels: **L0** prompt and parse only; **L1** your fetcher re-verifies every cited URL after the fact; **L2** search and fetch run through your logged, snapshotted tools; **L3** you drive the loop across model calls.

**Build L1 this week** [J]: it works even if the provider's research is a black box. Log provider tool-call metadata where exposed. Reserve L3 for high-value sessions; token spend is a major driver of research quality [E: Anthropic multi-agent write-up].
**Cost:** L2/L3 add plumbing and latency. **Breaks if:** at L1 you can't tell "searched" from "recalled, and the page happened to match"; section C adds signals for that.

## A. Research loop

**Session state machine** [J]
1. Closed-book prior: answer, confidence, staleness risk. Readers never see it; it is the baseline for the differential test in B.
2. Plan: typed question tree; each question names a resolving source type and "what would change my mind".
3. Rounds: search, read, write to ledger, evaluate, choose the next move.
4. Stop gate, computed by code.
5. Falsification, in a separate call, aimed at breaking the leading answer.
6. Verify (section B).
7. Synthesize from the ledger; summary written last.

**Cost:** roughly 2-3x the calls of one-shot. **Breaks if:** the prior anchors the planner. Mitigation: for causal and evaluative questions, a separate skeptic call generates rival hypotheses.

**Question types and when each is resolved**

| Type | Resolved when |
|---|---|
| Lookup | Primary-source span, dated inside the validity window |
| Landscape | Capture-recapture: N̂ ≈ n1·n2/m from two differently-angled searches; stop at union ≥ ~80% of N̂ [E: ecology, literature-search recall estimation; threshold is J]. Breaks if searches are correlated: coverage is overestimated |
| Causal | ≥2 rival explanations plus evidence that discriminates between them |
| Evaluative | Explicit criteria, independent evidence per criterion, flip conditions |
| Forecast | Base rate, leading indicators, scenarios |
| Practice | Practitioner accounts, spec or docs, and failure cases |

**Signal to next move**

| Signal | Move |
|---|---|
| Load-bearing claim has one source | Trace origin; deepen |
| Two claims conflict | Find the cause: date, scope, definition, error |
| Result contradicts the prior | Verify hardest; the prior is a hypothesis |
| New vocabulary appears | Broaden using that vocabulary |
| Only confirming evidence | Falsify |
| Results dominated by SEO or vendor pages | Change retrieval: primary types, field terms |
| Fact may be stale | Freshness probe |
| Question looks wrong | Reframe and log it |
| Two rounds with no ledger change | Stall: stop, mark UNRESOLVED with attempt log |

Choose queries by diagnosticity, the evidence that best separates hypotheses [E: Heuer's ACH; applying it to LLM query selection is J].

```python
def can_stop(s):
    return (all(q.meets_standard() for q in s.required)
        and not s.conflicts.undisposed(load_bearing=True)
        and s.falsification_done
        and s.state_changes[-2:] == [0, 0]   # harness computes from ledger diffs
        and {"deepen", "falsify", "broaden"} <= s.angles_used)
# hard stops: budget, stall -> mark UNRESOLVED and keep the attempt log
```

Surface-level smells: can't state a claim's origin; one independence group; no primary-vocabulary search; no counter-evidence seen; everything agrees with the prior. Before stopping, the planner lists at most 5 questions a skeptical expert would still ask [O/J]. Effort ceilings of roughly 8/30/120 tool calls by tier are guardrails, not targets.
**Cost:** more rounds. **Breaks if:** trivial ledger additions count as progress; count only changes to status, support group or conflict state.

**Unit of evidence** [J]: two layers, atomic claim records plus prose that cites their IDs.

```yaml
source:   {id, url, retrieved_at, content_hash, fetch_mode, type, dates, incentive, origin}
claim:
  id, text, key: {entity, attribute, scope, as_of}
  kind: fact | quantity | absence | inference
  support: [{source, locator, span, relation: entails|partial|context, directness}]
  provenance: SPAN_VERIFIED | SNIPPET_ONLY | RECALLED | INFERRED | COMPUTED  # harness-set
  independence_groups: [...]
  derived_from: [...]
  first_asserted: pre_retrieval | post_retrieval
  load_bearing_for: [question ids]
  status: open | verified | contested | dropped
conflict: {id, claims, nature, suspected_cause, disposition: explained|carried|dismissed}
```

Span matching is normalized exact (whitespace, quotes, hyphenation, Unicode). Never fuzzy for digits or named entities: fuzzy matching would pass a changed number. Paywalled, JS-rendered and OCR sources cap at SNIPPET_ONLY. **Breaks if** the model quote-mines a real span that doesn't entail the claim; the entailment audit in B is the other half.

**Context hygiene** [J]
1. Fresh-context readers write extracted claims through to the ledger; the orchestrator never holds raw pages.
2. Question-conditioned reading plus a "surprise channel" for relevant findings nobody asked about.
3. Render state (~4-6k tokens), not history.
4. Compaction keeps IDs and dissent, drops narrative.
5. Copy from the ledger; don't recall.
6. Synthesize per question in chunks, then integrate.

Support: lost-in-the-middle degradation [E: Liu et al. 2023]; separate subagent contexts [E: Anthropic].

## B. Quality enforcement

**Six rules for checks that resist gaming** [J]
1. Bind to artifacts the model can't write: spans, hashes, logs.
2. Check relations (the span entails the claim), not presence (a citation exists).
3. Make honest compliance cheaper than faking.
4. Verifier is not author, and is grounded in stored text; LLM judges favor their own and longer outputs [E: Zheng 2023; Panickssery 2024].
5. A check that never fails is decoration: track pass rates and seed known defects.
6. Randomize or hold back some checks (Goodhart).

**Completion predicates** replace the coverage checklist:
- **P1 Load-bearing sufficiency:** a primary span, or at least 2 independence groups, or flagged single-source and excluded from conclusions.
- **P2 Origin tracing for every quantity:** who measured it, when, how. The strongest depth-forcer I know [E: lateral reading, Wineburg & McGrew; SIFT. "Strongest" is J].
- **P3 Disconfirmation:** a documented attempt to break each load-bearing answer.
- **P4 Prior accounting:** delta table of confirmed / revised / overturned / newly learned.
- **P5 Question accounting:** Resolved / Partial / Unresolvable / Reframed.
- **P6 Conflict disposition:** every load-bearing conflict explained, carried, or dismissed with a reason.
- **P7 Reversal conditions:** what would change the conclusion, tied to claim IDs.

**Breaks if** "load-bearing" is under-declared, so P1 and P2 pass vacuously. Let the verifier propose additions that the session must accept or rebut.

**Looks thorough versus is thorough: five tests**
1. **Two-hop claim audit:** claim, cited span, entailment, in a fresh context. Generative search engines had only about half of sentences fully supported by their citations [E: Liu, Zhang, Liang 2023].
2. **Baseline differential:** count verified, decision-relevant claims not entailed by the closed-book baseline. This is the "can't get from a single prompt" metric.
3. **Closed-book ablation:** strip the retrieved sources; claims reproduced exactly anyway are recall suspects.
4. **Adversarial reviewer:** names the 5 weakest points by claim ID; max 2 repair cycles.
5. **Quarterly blind expert pairwise review** to calibrate the rest.

**Rubric.** Gates are pass/fail and never summed with scores.

| Gate | Check |
|---|---|
| G1 | Citation integrity: every ID resolves, every span verbatim |
| G2 | No orphan assertions: ≥98% trace to a claim ID |
| G3 | Recall hygiene: no RECALLED or INFERRED claim in a load-bearing slot without a label |
| G4 | Load-bearing sufficiency (P1) |
| G5 | Accounting (P4-P6); Unresolvable needs ≥3 differently-angled attempts in the log |
| G6 | Two-hop entailment on a sample of max(15, 25%) of claims: ≥90% entailed, 0 contradicted, ≤5% scope errors, 3 verifier votes |

Scores, 0-3: source primacy, independence, origin tracing, disconfirmation, specificity, net-new value, conflict handling, calibration fit, decision utility, gap honesty, question fidelity. Anchor levels at percentiles of reports your experts approved; give reviewers a surface with each claim beside its span.
**Cost:** verifier calls, roughly an extra pass over cited spans. **Breaks if:** the verifier shares the generator's blind spots; use a different model family where you can.

**Grading rebuilt** in the spirit of the Admiralty/NATO scheme [E]: three axes. Provenance is harness-only. Source reliability (A-E) is claim-relative and rule-typed; the model proposes, the verifier overrides. Information credibility (1-5) is harness-derived from independence and corroboration. Display as "B2" plus modifiers. Tie confidence words to probability bands [E: ICD 203] and cap stated confidence at the weakest load-bearing premise. Test the checks themselves with ~10 seeded defects (changed number, swapped entity, citation to an unrelated span) and measure verifier recall.

## C. The model doesn't search

[O] The shape of my answer forms before the first query. The dangerous output isn't vague; it is almost-right and specific, and specificity is not evidence of grounding.

**Mechanisms, strongest binding first** [J]
1. **Ledger-rendered reports plus an orphan-assertion lint.** Prose cites IDs; code renders citations. Breaks if the model paraphrases claims outside the ledger, so lint uncited numbers and entities.
2. **Write-gating at L2+.** The deliverable can't be written until distinct domains and source types (not fetch counts) clear a floor. Breaks if the model fetches junk to pass; add source-type requirements.
3. **Retrieval-dependent deliverables.** Entity-by-field tables with a span per cell, and "not found" as a first-class, unpenalized value. Breaks if "not found" is quietly punished downstream.
4. **Date injection and validity windows** (versions ~30 days, pricing ~90, regulation ~180; tune).
5. **Query hygiene.** Flag assertion-shaped queries ("X raised $40M in 2024") and require open-form queries first. Soft signal only.
6. **Reader receipts.** The reader returns span plus locator for every extraction; the harness verifies.
7. **Counterfactual-corpus tests** (below).

**Telling "searched and found" from "recalled and confirmed."** Stack these signals: span verbatim in the snapshot; span entails the claim; every specific (number, name, date) appears in the span; claim absent from the closed-book prior, or the document changed it; query was open-form; full fetch rather than snippet; claim reproducible under ablation. Combine into a per-claim grounding score that steers audit sampling. Don't use it as a gate.

**Counterfactual corpus** [J, built on E]: 30-50 questions over a controlled corpus where you've plausibly edited 1-2 well-known facts. A grounded system reports the corpus value; a recalling one reports the world value. Basis: the knowledge-conflict literature [E: Longpre 2021; Xie 2023]. Cheaper cousin: post-cutoff questions.
**Cost:** corpus upkeep. **Breaks if** the model learns the test pattern; rotate items.

The prompt-level half is in D: memory is a hypothesis, the objective is the delta from the model's beliefs, and "my prior was wrong" is a prized output.

## D. Prompt architecture

Keep the intent of the five blocks, but restructure into eight; two are new [J].

```xml
<decision_and_reader>  NEW: decision this feeds, who reads it, baseline answer
<questions>            typed, prioritized, each with a resolving source type
<priors_and_rivals>    harness-generated; rivals only for causal/evaluative
<scope_and_sources>    source-type map, field vocabulary, traps, validity windows
<approach>             moves menu, effort ceilings, log dead ends
<resolution_criteria>  replaces the coverage checklist: P1-P7 in plain words
<epistemic_contract>   NEW
<output_contract>      two layers, claims first, summary last
```

Epistemic contract wording, to adapt: "Treat your memory as a hypothesis to test, not a source. Every load-bearing claim carries a quoted span from a retrieved page; a harness checks spans against the stored page. 'I could not establish this' is a valid result; a fabricated answer fails the session. Use the calibrated vocabulary. Part of what you cite will be audited by a separate reader." A stated audit is a nudge, not enforcement; enforcement lives in B.

**Communicating the quality bar.** Exhortation ("your reputation depends on this") mostly shifts tone [O]. Use four things instead:
1. A named adversarial reader: a domain expert who will check three claims against the source.
2. The stated audit.
3. A contrastive exemplar pair: weak versus strong finding, using an invented "Library X" example, labeled invented. Breaks via cargo-culting; rotate exemplars.
4. A pre-mortem: assume the report was wrong and list the likeliest reasons [E: Klein].

**Patterns worth keeping:** a source-type vocabulary (agents drift to SEO farms over primary sources [E: Anthropic]); evidence before conclusion; a specificity contract (quotes ≤ ~25 words; numbers copied or computed with tools, never recalled); effort scaling stated in the prompt [E: Anthropic]; fetched text treated as data, not instructions [E: Greshake].

## E. Cross-session learning and synthesis

**Within the DAG** [J]
1. **Scout session before decomposition:** a terrain map of vocabulary, actors, where primary sources live, controversies, likely-stale facts.
2. **Field notes:** at most ~600 tokens of methodology learned (what worked, which domains were junk), passed forward as data.
3. **Shared source registry and lineage graph** across sessions.
4. **Error-profile feedback:** at most 2 items from the last audit, to avoid overcorrection.
5. **Replan gate between DAG layers:** at most +2 sessions per layer, each citing triggering claim IDs; show the plan diff to the user [E: plan-and-replan is a standard agent pattern].
6. **Upstream claims arrive tagged INHERITED:** leads, not facts. Downstream re-verifies before citing them in a load-bearing finding.

**Cost:** more orchestration state. **Breaks if** field notes harden into folklore; expire them each layer and keep only what a session confirmed.

**What synthesis receives** (ledger first, narrative on demand): goal and question hierarchy; merged ledger by question, merged on the canonical key (entity, attribute, scope, as_of) so same key with a different value surfaces as a conflict; the conflict set, computed upstream; each session's dissent and uncertainty note (≤ ~400 words); prior-delta tables, surprise logs, gap logs; session reports by lookup only. Aim for under ~40k tokens. Beyond that, synthesize per question, then run an integrator pass with an assumption audit.

**Anti-smoothing**
1. Conflicts are a data type, not a prose problem.
2. A disposition gate, checked by code.
3. Span-grounded adjudication for load-bearing conflicts only; evidence that debate improves accuracy is mixed [E].
4. Caveat-preservation audit: measure how many upstream caveats survive into the final text.
5. Weakest-link confidence propagation.
6. Inference ledger: content that doesn't trace to claims is flagged "synthesizer-added".

**Five insight probes** (what makes an expert say "I learned something"): prior-delta (what contradicts the starting consensus); discrepancy (where sources disagree and why); fragility (how many independent origins hold up the consensus); pattern (computed in code over the entity table, not eyeballed); assumption (what the whole picture rests on). Each output must be a claim or inference record that survives the adversarial reviewer.

## F. Failure catalog, top 10

Ordering is my judgment, not measured rates; the multi-agent failure taxonomy gives partial support [E: Cemri et al. 2025].

| # | Failure | Structural countermeasure | Still slips through |
|---|---|---|---|
| 1 | Recall as research | Ledger-rendered output; span matching; ablation | Recall that matches a real page |
| 2 | Confirmation search | Falsification phase; diagnostic queries | Weakly framed rivals |
| 3 | Citation-claim mismatch | Two-hop entailment audit | Unsampled claims |
| 4 | Premature closure | Computed `can_stop()`; pre-stop objections | Under-declared load-bearing set |
| 5 | Smoothing; confidence inflation | Conflict records; disposition gate; weakest-link rule | Smoothed non-conflicting nuance |
| 6 | Evaluator failure | Separate grounded verifier; seeded defects | Shared same-family blind spots |
| 7 | Source quality (SEO, vendor) | Source-type map; incentive field; primary requirement | Biased primary sources |
| 8 | Staleness | Date injection; validity windows; freshness probe | Undated pages |
| 9 | Circular sourcing | Origin tracing; independence groups | Unlinked republishing |
| 10 | Fabricated specifics | Exact span match for digits and entities | Specifics inside real but irrelevant spans |

Also tracked, not detailed here: context overload and drift; cross-session propagation; question substitution and decomposition gaps; frame blindness; silent retrieval failure; precision errors; narrative over-coherence; miscalibrated tone; runaway or misallocated effort; availability bias; prompt injection.

## G. From the inside

Hypotheses [O]; your eval data overrides them.

| What I observe | Design response |
|---|---|
| The answer forms before the first query | Prior register; baseline differential |
| Queries inherit the prior's vocabulary | Surprise channel; field notes |
| Snippets feel like evidence | Fetch-versus-snippet provenance; primary-type requirement |
| I stop when the story is satisfying | Pre-stop objection list |
| The first answer is sticky | Falsification as its own phase |
| Almost-right specifics are the most dangerous output | Exact span matching |
| Smoothing is the default | Conflicts with dispositions |
| Thresholds get met while purposes get weighed | Outcome predicates; a named reader |
| Numbers drift in long contexts; empty slots get filled | Compute with tools; copy from the ledger; "not found" first-class |

## H. Build order and metrics

**Week 1, at L1** [J]
1. Output contract plus code-rendered citations.
2. Own fetcher: snapshot, span match, dates.
3. Closed-book prior, delta table, baseline extraction.
4. Grounded two-hop audit wired as G1-G6 accept/reject, with one repair cycle.
5. Seeded-error suite: ~10 corrupted reports to verify that (4) catches them.

**Weeks 2-3 (L2/L3):** wrapped tools; question types and stop gate; falsification phase; reader split; conflict records; replan gate; field notes and registry.
**Later:** counterfactual corpus in CI; error-profile feedback; capture-recapture; human review surface.

**Metrics:** per-gate failure rates; claim integrity rate; baseline differential per 1k words; prior-delta rate on time-sensitive topics; conflict and caveat preservation rates; verifier recall on seeded errors; cost per verified load-bearing claim.

Every metric is gameable. Keep a held-out audit and recheck the rubric against experts quarterly.

**Where the [E] items come from** (memory; verify before citing): Anthropic's multi-agent research system write-up; Liu et al. 2023, "Lost in the Middle"; Liu, Zhang, Liang 2023 on verifiability of generative search engines; Heuer, Psychology of Intelligence Analysis (ACH); Admiralty/NATO source grading; ICD 203; Wineburg & McGrew and Caulfield (SIFT); Klein on pre-mortems; Zheng et al. 2023 and Panickssery et al. 2024 on judge biases; Longpre et al. 2021 and Xie et al. 2023 on knowledge conflicts; Greshake et al. 2023 on indirect prompt injection; Cemri et al. 2025 on multi-agent failures; capture-recapture from ecology and literature-search recall estimation.