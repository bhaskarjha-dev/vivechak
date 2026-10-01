I'll check a few of the "well-established" claims against current sources first, so the labels in the design mean something.Your suspicion is right, and I'd sharpen it. The weak point isn't a prompt block. Quality is currently asserted by the producer: grades, coverage checklists and a structural exit gate all check what the output says about itself. The design below moves each quality signal from assertion to audit. The harness records what was actually retrieved, an independent process checks claims against that record, and "done" is defined over the state of the evidence rather than the shape of the document. Nothing here replaces your DAG, decomposition or A–E letters. It changes who assigns the letters and what "done" means.

Three recent results point the same way. An audit of commercial deep-research systems measured citation accuracy between 40% and 80% depending on the system, plus a substantial share of statements that the systems' own cited sources didn't back up. A benchmark built on expert-written rubrics found that leading deep-research agents satisfy under 68% of criteria, mostly because they overlook unstated context and reason poorly about what they retrieved. And an August 2026 study of 800 research-agent trajectories found that the most common failure was noticing a serious problem while reviewing their own work and then delivering the conclusion unchanged, which occurred in 82.5% of trajectories. That last study covers scientific-discovery agents rather than web research, so I'm mapping it by analogy, and its authors say they didn't test whether orchestration-level fixes close these gaps. Treat what follows as an informed hypothesis about how to build those fixes, not a validated recipe.

I tag recommendations [E] when they rest on published results or documented production practice, [J] when they're my design judgment, and [E+J] when an established idea is applied in a way I haven't seen tested.

## Premises I'm designing around

You asked for my perspective as a system that does research. I don't have reliable introspective access to my own processing and I carry no memory of past runs, so read these as hypotheses about models like me that your telemetry should confirm or kill.

- **Recall arrives before search does.** On a familiar topic an answer exists before the first query, so search drifts into confirmation. Disagreeing results get explained away more easily than agreeing ones get questioned, though I've also seen the opposite failure: over-trusting one confident-sounding page. The defect is asymmetric conflict handling, not "trusting memory." [J]
- **Writing pulls against honesty.** The most fluent continuation of a research report is a coherent, authoritative narrative. "Not found" and "these sources disagree" read as breaks in flow and get smoothed out. [J]
- **Anything satisfiable by text will be satisfied by text.** A grade label, a filled-in checklist and a limitations section are cheap, and the work behind them isn't. One trajectory in the 800-run study shipped its review section as unfilled placeholders. [E+J]
- **Default search habits are poor.** Agents write long, over-specified queries, stop at the first answer-shaped page, and prefer clean SEO pages to messy primaries. Anthropic's team saw agents choose SEO content farms over authoritative but less prominent sources, and fixed it by adding source-quality heuristics to the prompts. [E]
- **Finding a flaw is easier than fixing it.** Once I've committed to a claim in context I tend to defend it. A fresh-context reviewer is a harsher and more accurate judge of the same text. [J]
- **"Comprehensive" produces list inflation.** You get many true, obvious bullets that look like depth. The scarce thing is a non-obvious, decision-relevant finding, which is what you said you want. [J]

## One constraint decides how much you can enforce

Your framework sits between the user and a model's own research capability, so the first question is how much of the tool layer you can see.

- **Tier 0, black box (prompt in, document out).** You can only audit afterward. Re-fetch every cited URL yourself and check that quoted spans exist, run a closed-book baseline for contrast, run a red-team pass, and feed failures back as a repair session.
- **Tier 1, visible tool calls.** Add query-log and URL-in-log checks and query-leading-ness linting. Most provider APIs return search queries, result URLs and citation spans in the response object, but check what yours exposes.
- **Tier 2, harness-owned search and fetch.** Everything below becomes possible, including extraction at the edge and rebuilt contexts.

Build the Tier 0 audits first. They work with any execution layer, and their telemetry tells you where deeper control pays. [J]

## Shared primitives

Nine pieces recur below, so here is the vocabulary.

1. Evidence store: harness-owned. Every query, fetched URL, status, retrieval time, content hash, extracted page date and the full page text. The model cites from it and never writes to it.
2. Claim records: atomic, scoped, span-anchored records whose A–E grade is computed by the harness.
3. Question ledger: sub-questions in states open, answered, contested, or unanswerable-with-trail. It's seeded by your brief plus examiner questions, which a separate agent writes from the brief alone: 10–15 questions a skeptical domain expert would ask, plus a hidden task-specific rubric.
4. Prior capture: a no-tools pass that writes checkable propositions before any search. The final report carries a delta against it.
5. Independent verifier: fresh context, sees only a claim and its cited span.
6. Red-team gate: a fresh-context agent with search tools whose job is to break the finished work.
7. Closed-evidence writer: composes the report from claim records and the ledger only.
8. Contradiction register: a typed list of claims in tension that synthesis must resolve or carry forward.
9. Field guide: process knowledge handed between sessions.

Here is how one session runs once those pieces are in place.## A. The research loop

**Round structure.** Each round picks the highest-value open items (blocking questions first, then contested ones, then weakly supported claims). It runs short broad queries before narrow ones, fetches full text from the best sources (primaries preferred), extracts spans, updates the ledger and claim records, and runs the trigger check. Context is rebuilt from the ledger each round instead of accumulating transcripts. Short queries first is documented practice: Anthropic's team found that agents tend to write long, highly specific queries that return little, and fixed it by prompting for brief, broad queries first and narrowing afterward. [E+J]

```
prior   = closed_book(brief)       # no tools: checkable propositions + "I suspect this is stale"
ledger  = plan(brief, prior, examiner_qs, inherited_open_qs)
while budget_left and not ledger.terminal():
    for t in ledger.next(k=3):                              # parallelizable
        hits  = search(neutral_queries(t, field_guide))
        pages = fetch(pick(hits, prefer="primary", diversify="origin"))   # full text, never snippets
        spans = extract(pages, t)  # fresh context: verbatim spans, qualifiers, dates, plus "what limits or contradicts this?"
        store.add(spans); ledger.update(t, spans)
    fired = triggers(store, ledger)
    notebook.append(changes, surprises, weakest_claim, next_move)   # next_move: deepen, widen, corroborate, falsify, chase, pivot
    if saturated(store) and not fired: break
falsify(ledger.load_bearing())                              # one disconfirming query each, logged
repairs = verify(store) + red_team(store)                   # failures become ledger items
if repairs: re-enter loop (max 2 cycles)
report = write_from(store, ledger)
```

**What triggers another round.**

| Trigger | Detected by | Typical move |
|---|---|---|
| Blocking question still open | Ledger state | Deepen, or widen the vocabulary |
| Load-bearing claim graded C or below | Computed grade | Chase to the primary, or corroborate from a distinct origin |
| Two claims share a key (entity, attribute, as-of) but differ in value | Key collision | Find the cause: definition, date, scope, error |
| A finding contradicts the prior | Prior status = updated | Get a third independent source before accepting |
| A new entity or term appears in 2+ sources but not in the plan | Entity diff on extractions | Add a ledger item (unknown unknowns) |
| Fast-moving claim with no source inside its freshness window | Harness-extracted page dates | Search for newer, or mark stale |
| All support traces to one origin | Origin groups | Find a distinct origin |

Most of these are deterministic checks over the store, so the model doesn't decide whether to continue, only how. That matters because the 800-trajectory study flagged shallow search coverage in roughly 55% of trajectories. [E+J]

**When to stop.** Stop when one of three things happens:

- every blocking ledger item is answered, contested, or unanswerable with a logged trail;
- two consecutive rounds add no load-bearing claim and change no ledger state;
- the budget cap hits.

Hitting the cap is allowed, but the output must carry the open items, because honest incompleteness is a valid result. Reserve about 15% of budget for unplanned exploration, with a "what surprised me" log that can spawn ledger items. A ledger-driven loop optimizes for answering known questions, and the findings an expert calls new often come from adjacent discoveries. [J] *Breaks when:* the 15% turns into tangent-chasing. Require each wander to end in a ledger entry or be dropped.

**"Deep enough" is a property of each claim, set by its role.** [J] Use a depth ladder:

- L0, recalled
- L1, seen in a snippet
- L2, read in full
- L3, primary source reached
- L4, independently corroborated from a distinct origin
- L5, survived a logged falsification attempt

Background claims need L2. Load-bearing claims need L4 and L5. Anything that contradicts the prior needs L4 before it's accepted. This replaces search-count quotas: a claim carrying a recommendation can't pass until it has been chased to its origin and attacked. Before exit the model must also write, in the notebook, the three strongest objections a domain expert would raise, the query it ran for each, and what came back. That turns "have I gone deep enough" into a log the red team can check against reality.

**Unit of evidence.** Use structured claim records for everything load-bearing and free prose for narrative. Atomic-claim verification is established practice (FActScore and SAFE both decompose long-form text into atomic facts and check each). The additions here are the key, the origin group and the dependency edges. [E+J]

```json
{
  "id": "C12",
  "claim": "Atomic, falsifiable, scoped: who/what/when/where/how much",
  "key": {"entity": "…", "attribute": "…", "as_of": "2026-09"},
  "type": "fact | statistic | causal | forecast | definition",
  "load_bearing": true,
  "volatility": "static | slow | fast",
  "support": [{
    "src": "F07", "span": "verbatim, ≤40 words",
    "relation": "supports | contradicts | qualifies",
    "directness": "primary | secondary | tertiary", "origin_group": "G3"
  }],
  "origin": "searched | recalled | derived",
  "prior_status": "confirmed | updated | new",
  "falsification": {"query": "…", "found": "…"},
  "unresolved": "what would change my mind"
}
```

Source records (F07) are created by the harness, not the model. A claim is load-bearing if some conclusion lists it as a dependency. The model declares conclusions and their dependencies, so importance is an edge in a graph rather than a self-assessed adjective. *Cost:* roughly 20–40% more output tokens (my estimate) and a stiffer format. *Breaks when:* models pad with trivially true claims to look rigorous. Dependency edges, a cap of about 30 records per session, and the non-obviousness score in B all contain that.

**Context hygiene, so more searching doesn't degrade synthesis.** Across 18 models, performance became less reliable as input length grew, even on simple tasks. Anthropic's system leans on subagents that compress material in their own context windows and pass condensed results up, plus filesystem handoffs to avoid information loss between stages. Four mechanisms follow [E+J]:

1. **Extract at the edge.** Raw pages never enter the reasoning context. A cheap fresh-context extractor receives (sub-question, page) and returns verbatim spans, qualifiers, the page date and one span that limits or contradicts. Pages stay in the store for verification.
2. **Rebuild state each round.** Round context is the brief, the ledger, the claim records and that round's extractions.
3. **Synthesize hierarchically.** Write per-question mini-answers (≤300 words, with claim IDs), then do global synthesis from the mini-answers plus the contradiction register. The writer pulls spans on demand (`get_span(claim_id)`) instead of having them preloaded.
4. **Put long material first and the instruction last.**

*Cost:* extraction strips caveats, hence the qualifiers field and the mandatory limiting span. *Breaks when:* the extractor prompt is leading ("find evidence that X"), which launders confirmation at the edge. Phrase it as a neutral question.

## B. Quality enforcement that can't be sprinkled on

**What separates a real mechanism from a ceremonial one.** [J] A real mechanism has four properties:

1. **It's verified by something the producer doesn't control** (the store, a deterministic check, a fresh-context agent). The 800-run study states the problem sharply: the agent writes both the claim and the file that contradicts it, and nothing ever compares them, and a self-review is just more text that nothing forces to change the report.
2. **It's cheaper to satisfy honestly than to fake.** A verbatim span from a fetched page is hard to invent, while a grade label is free.
3. **It's outcome-linked, not activity-linked.** It counts contradictions resolved, primaries reached and claims attacked, never searches performed.
4. **It's binding, with specific repair tasks.** A failure becomes a named work item, and a critical one blocks release.

**Test your gates before trusting them.** [E+J: mutation testing is standard in software, but applying it to research QA is my adaptation.]

- **No-search control.** Run the whole pipeline with search disabled or returning nothing, using the same output schema. The gate must reject. If it passes, it measures style.
- **Mutation set.** Take a good report and inject defects: swap a number, replace a span with unrelated text, delete the contrary-evidence section, swap a primary source for a blog, backdate a source. Each must be caught.

*Cost:* about a day of harness work. *Breaks when:* you only mutate defects you already thought of. Add every real failure you find to the set.

**Completion criteria that force depth without prescribing method.** Define done over states of the ledger and claim graph (gates G5–G7 below). Replace the coverage checklist with the examiner's questions and hidden rubric, generated by an agent that sees the brief but not the producer's plan. A recent survey of rubric-based evaluation reports that context-aware, task-specific rubrics significantly outperform generic ones. The producer also can't dodge hard questions by never writing them down. The 800-run study's authors argue that "never considered a second hypothesis" can't be fixed by gating afterward, because nothing was written down to compare against. Injecting alternatives up front from a separate agent is my way around that, and it's untested. [E+J]

**Looks thorough versus is thorough.** Cheap tells, computed from the store [J]:

- no source dated after the model's training cutoff on a fast-moving topic;
- few domains, many listicles and vendor blogs, and a low primary-source share;
- any quoted span missing from the fetched text;
- uniform confidence and zero stated unknowns;
- zero contradictions where the red team finds disagreement;
- a near-zero delta on a volatile topic;
- every recommendation hedged both ways, with no tripwire.

Tests, cheapest first:

1. The negative controls above.
2. **Closed-book contrast.** Run the brief with tools off and measure how much of the final report the closed-book run already contained. The value of your research is what's left.
3. **Held-out examiner probe.** A separate agent writes 10 hard questions from the brief alone, and you check whether the report answers them correctly with evidence. This is nugget-style coverage from IR evaluation, with adversarially generated nuggets. [E+J]
4. **Red-team attack success.** Attacks fall into four types: support, omission, staleness and perspective. Each attack needs a URL and a span, so the same standard binds the attacker.
5. **Planted-discovery suite.** Write 20–30 briefs where you know there is a post-cutoff fact, an obscure primary document two or three hops deep, a widely repeated outdated "fact", an SEO page with a plausible wrong number, or a real disagreement between credible sources. Score the hit rate. This measures the system rather than a run, and it's the closest proxy for "I learned something."
6. **Expert blind comparison** against a single-prompt baseline: "which taught you something new and correct?" Run it quarterly to calibrate the cheaper proxies. Start with about 20 briefs rather than waiting for a large eval, which is what Anthropic recommends. [E]

**The rubric.** Keep your A–E letters but change who assigns them. Give verifiers and judges the store, not just the report. In the 800-run study, an artifact-aware judge agreed with human labels substantially better (kappa 0.75) than a single-call judge reading only the transcript (0.53).

Hard gates (any failure rejects the output and emits repair tasks):

| ID | Check | How it's computed | Pass |
|---|---|---|---|
| G1 | Citation integrity | Every cited source exists in the fetch log. Every quoted span fuzzy-matches stored text (≥0.9 after normalization). | 100% |
| G2 | Support validity | A fresh-context verifier sees only claim and span (plus neighboring text) and returns entailed, partial, not entailed or contradicted. A deterministic check requires every number, date, entity and quantifier in the claim to appear in the span or a logged derivation. | 0 contradicted; ≥90% entailed on load-bearing claims; the rest repaired or demoted |
| G3 | Origin honesty | Every claim marked searched has a fetch event. Recalled claims are capped at E and can't be sole support for any conclusion. | 100% |
| G4 | Freshness | Each fast- or slow-volatility load-bearing claim has a source whose harness-extracted date falls inside its window. Undated sources don't count. | 100%, or flagged stale in prose |
| G5 | Ledger terminal | No open blocking items. Each unanswerable item has ≥3 attempts across ≥2 move types. Every examiner question is addressed. | 100% |
| G6 | Dependency integrity | Every conclusion lists claim IDs, and its confidence can't exceed its weakest load-bearing dependency. An LLM sentence classifier flags factual sentences with no claim ID. | ≤5% unanchored |
| G7 | Binding review | Every critical issue in the review or limitations produced a ledger change (demote, remove or fix). No headline claim contradicts a recorded limitation. | 0 open critical |

Computed grades. The model proposes the fields and the harness decides the letter [J]:

- **E:** no retrieved support (recall or inference).
- **D:** retrieved but snippet-only, or from a low-tier or undated source, single origin.
- **C:** full text of one credible secondary source read, or a primary read only in part.
- **B:** primary read in full with a verified span, or two or more independent-origin credible secondary sources agreeing.
- **A:** B plus independent corroboration (or direct examination of the primary data), inside its freshness window.
- **Modifiers:** `~` stale, `!` contested, `Δ` contradicts the prior, `calc` derived by a logged computation.

"Primary" is claim-relative: a vendor's pricing page is primary for its price and tertiary for its performance. Directness and origin group are still model-declared, so the verifier audits them on every load-bearing claim, and the harness derives what it can itself (domain class, fetch depth, page date).

Scored dimensions (a judge with store access, binary sub-questions where possible):

| ID | Dimension | Measured by | Anchors, 0 to 3 |
|---|---|---|---|
| S1 | Evidence depth | Share of load-bearing claims at B or better; share with origin traced | 0: under 30% B+ · 1: 30–50% · 2: 50–75% · 3: over 75%, half independently corroborated |
| S2 | Disagreement fidelity | Contested points identified, typed (definition, date, scope, method, incentive, error) and adjudicated; recall against disagreements the red team finds | 0: none found and the red team finds one · 3: all typed, resolving evidence named |
| S3 | Non-obviousness | Delta ratio (new plus updated verified claims over all claims), plus a practitioner-proxy rating | 0: under 5% and rated known · 3: over 20% and rated plausibly new (calibrate per domain) |
| S4 | Mechanism and bounds | Per key finding, three binary checks: mechanism, boundary condition, counterexample addressed | Pass rate under 40% · 40–65% · 65–85% · over 85% |
| S5 | Calibration | Stated confidence against verifier outcomes; unknowns listed with search trails | 0: no unknowns, uniform confidence · 3: confidence tracks outcomes, unknowns documented |
| S6 | Decision utility | Per recommendation: the decision, a flip condition, cost of being wrong, cheapest next test | Share of recommendations with all four |
| S7 | Adversarial robustness | Successful red-team attacks weighted by severity | 3: none major · 0: two or more major |
| S8 | Held-out coverage | Share of examiner probe questions answered correctly with evidence | Calibrate per domain |

**Exit rule.** Ship when all gates pass, S1, S2, S5 and S7 are at least 2, and the mean is at least 2.0. Otherwise run up to two repair cycles, then ship with a visible Quality Status block listing what failed. Treat thresholds as starting points. Calibrate on 30–50 human-labeled outputs and track judge agreement (Cohen's kappa). Judges favor their own model family and longer outputs, so use a different family where you can and score length-neutrally. Start with one judge call per dimension, since Anthropic found a single LLM call with a rubric more consistent than multiple judges per component. [E+J]

*Breaks when:* the S3 practitioner proxy only measures distance from the model's own prior, not the expert's. Calibrate against 30+ expert-labeled claims per domain.

## C. When the model doesn't search

Recall is available, the prompt is answerable from it, nothing penalizes skipping search, and every requirement you impose can be met with text. So the fix has two halves: make fabrication detectable, and make honest compliance the cheapest path.

1. **Provenance gating.** [E+J] At Tier 1–2 the harness rejects any claim record whose source isn't in the fetch log or whose span isn't in the stored text. At Tier 0 you re-fetch cited URLs and match spans yourself. The 800-run study's judge did essentially this and found a reference listed as the primary source whose URL never appeared among the pages the run actually fetched. *Cost:* paywalls and JS-rendered pages cause false rejects, so allow a failed-fetch source only at grade D with the failure logged. Pages also change between the model's fetch and yours, so treat span mismatches on dynamic pages as "needs archive," not "fabricated." *Breaks when:* you trust model-reported URLs without matching them to actual tool events.

2. **Closed-evidence writer plus a recall lane.** [J] The writer composes from claim records only. Any sentence without claim IDs must be tagged `[inference]`, and the harness counts them. Separately, give recall a legitimate home: a short "Background (unverified recall)" section that can't support any conclusion. If recall has no sanctioned outlet, it gets laundered as research. *Cost:* reports feel segmented. *Breaks when:* the lane becomes a dumping ground. Cap its length and exclude it from dependency edges.

3. **Prior capture and delta.** [J] A first pass in a separate context, with no tools, writes checkable propositions (each with a number, name, date or mechanism) and flags which it suspects are stale. The search phase receives them as hypotheses to test. The report ends with a Delta section covering what the evidence updated or contradicted and what was new. That gives a per-claim measure of whether search changed anything, and it is your "I learned something" metric. *Cost:* one cheap call, plus anchoring risk, since priors can steer what gets searched. Pass only the propositions into the search context and require at least one neutral query per proposition. *Breaks when:* propositions are vague ("the market is growing"), because vague claims can't be updated.

4. **Telling "found" from "recalled, then confirmed."** [J] You can't observe the cognitive process, only its correlates. Four cheap ones:
   - **Order:** did the claim appear in the prior before the retrieval that supports it?
   - **Query leading-ness:** what share of first-query tokens (numbers, entities, exact phrasings) came from the prior rather than the brief?
   - **Scope drift:** do numbers, dates or quantifiers in the claim fail to appear in the span?
   - **Surprise:** what is the delta ratio?

   Policy: a prior-confirmed claim whose first query was leading can't exceed grade C until a source reached through a neutral query corroborates it. This detects confirmatory search patterns. It can't prove their absence.

5. **Symmetric conflict rule.** [J] When retrieved evidence conflicts with the prior, neither wins by default and a third independent origin decides. This blocks both failures from the premises: explaining away surprises, and over-trusting one page.

6. **Evidence-shaped briefs.** [J] Specificity and recency are what make memory unreliable, so have your generator instantiate generic briefs with named candidates, an as-of date, the user's actual constraints, "what changed since <date>," and primary-document requirements. "Best vector databases" is answerable from memory. "Which of these five support requirement X today, per release notes and issue trackers from the last six months" isn't. Where possible, make the deliverable a table whose cells are individually fetchable (release date, last commit, current price on the pricing page, status of a specific issue), each with its own source and as-of date. Cells are cheap to audit and expensive to invent.

Two cheap extras:

- Seed the planner with results from 5–10 neutral searches across source types (docs, issue trackers, filings, forums), so the plan is conditioned on current reality.
- Put one liveness canary in each session: a sub-question whose current answer the harness can fetch itself, such as a package's latest release from its registry. It catches silently broken search paths. [J]

## D. Prompt architecture

Your five blocks aren't wrong, but they're missing four things.

- **Decision frame:** what decision this feeds and what would change it. This is the biggest single lever for prioritization and for defining "enough."
- **Reader model:** what the reader already knows, so "valuable" can mean "not derivable from that." This is how you operationalize "I learned something." It also targets the biggest weak spot in the benchmarks: implicit criteria make up about 39% of one expert rubric, and agents fail on implicit reasoning 45–50% of the time. [E+J]
- **Epistemic contract:** the evidence standard and loss function, stated as mechanisms rather than exhortations.
- **Done-definition**, replacing the coverage checklist.

Approach should also become a protocol with required artifacts (prior, ledger, per-round notebook entry, falsification log). That constrains behavior without prescribing queries.

```
DECISION   This research feeds: <decision>. It changes if: <thresholds/conditions>.
READER     Knows: <A, B, C>. A finding is valuable only if not derivable from that. Skip textbook material.
BRIEF      Resolve: <Q1..Qn>, plus examiner questions (injected, mandatory).
KNOWN      Trusted inputs with grades: <upstream claims>. Assumptions to stress-test: <list>.
           Field guide: <vocabulary, source map, prior-miss rates, dead ends>.
SCOPE      As-of date. Freshness windows by claim type. Source strategy by claim type
           (pricing: vendor page or archive; performance: independent benchmark or its repo; legal: statute or docket).
CONTRACT   1. Your memory is a list of hypotheses to test. Write them down first.
           2. A claim exists only with a stored source and a verbatim span. The harness assigns grades.
           3. An unsupported claim costs more than an omission. "Not found; tried X, Y, Z" is a good answer.
           4. Before finishing: three objections an expert would raise, the query you ran for each, what came back.
           5. Checks that will run: span match, independent entailment, red team. Others are not disclosed.
PROTOCOL   prior → plan → loop → falsify → write. Required artifacts: prior list, ledger, notebook entry per round, falsification log.
DONE       Ledger terminal. Each load-bearing claim: corroborated, falsification attempted, fresh.
OUTPUT     Conclusions first (claim IDs, confidence, tripwire), then Delta, Contested, Unknowns, evidence appendix (JSON).
```

Patterns that matter most for evidence-grounded research:

1. **Make the audit real and name it.** "A domain expert will check five random claims against your stored sources" beats "your reputation depends on this," and it's only credible if you run the check. Disclose the mechanical gates, since honest compliance is the goal, and keep the examiner rubric hidden to measure generalization. [J]
2. **State the loss function in plain words, then enforce it.** An unsupported claim costs more than an omission, and "not found, tried X, Y, Z" is a good answer. If the harness doesn't score it that way, the model learns the sentence is decoration. [J]
3. **Include one worked example of honest downgrading.** Show a claim recalled from memory, "confirmed" by a vendor blog, and demoted to D with the reason. Models calibrate on examples, and most prompts show only claims that succeeded. [J]
4. **Separate finding, judging and writing in time.** They want different temperaments: curious, skeptical, faithful. Anthropic likewise runs a distinct citation pass over the documents and report after the research loop ends. [E+J]
5. **Start wide, then narrow, and think between tool calls** to assess results and plan the next query. [E]
6. **Give source guidance by claim type** rather than a generic quality lecture. [E]
7. **Size effort in the prompt as guidance, not a quota.** Anthropic's rules of thumb were 3–10 tool calls for simple fact-finding, 10–15 calls per subagent for comparisons, and many subagents with divided responsibilities for complex research. State-based completion still decides when to stop. If you fan out inside a session, give each worker an objective, an output format, source guidance and boundaries. [E]
8. **Avoid** "comprehensive," "thorough," "world-class researcher" role-play, bare "cite your sources," search-count quotas and long rule lists. "Comprehensive" tends to yield enumeration, while a decision frame yields prioritization. Keep the contract to about six lines. [J]

## E. Cross-session learning and synthesis

Methodology doesn't transfer by injecting content. Five mechanisms, in rough order of expected payoff:

1. **Field guide.** [J] After each session an independent retrospective writes at most about 600 tokens of process knowledge, kept separate from findings:
   - a vocabulary map (what practitioners actually call things, probably the biggest search-efficiency gain);
   - a source-reliability map for this topic, stored as hypotheses with evidence counts;
   - prior-miss rates by claim category, which tell the next session where memory is least trustworthy so it spends search effort there;
   - dead ends, and newly found canonical documents, people and organizations;
   - inherited open questions.

   *Breaks when:* early biases entrench ("authority snowballing"). Store entries as hypotheses, require re-verification of anything load-bearing, and expire entries.

2. **Inherited-claim audit.** [E+J] Downstream sessions don't just receive upstream claims. They re-verify the top-k, prioritized by (number of dependents × low grade), and file corrections to the store. Grades never rise by repetition: a claim restated by three sessions is still one claim with one origin. Corrections propagate to every dependent. Failure studies keep finding this mechanism; one hallucination-focused audit traced failures to hallucination propagation and cognitive biases. *Cost:* extra verification, which prioritization limits.

3. **Staged rigor in the DAG.** [J] Scout sessions are cheap and wide, and they find the vocabulary and the real questions. Dig sessions run full gates on the load-bearing questions the scouts surfaced. A break session is a dedicated adversarial node whose only job is to refute the top conclusions, scored on verified problems found. Then comes synthesis. Methodology improves because rigor is staged explicitly, not because you hope the model learns.

4. **Gate telemetry becomes targeted lessons.** [J] If 30% of statistics in session 1 failed the containment check, session 2's prompt gets a lesson: quote table cells directly and state units. Keep it to five lessons and let them expire. *Breaks when:* lessons bloat the prompt or overfit.

5. **Adaptive replanning.** [J] After each session, the planner may add, drop or merge downstream sessions given what was learned, with changes logged and user approval beyond N additions. *Breaks when:* scope creeps and runs stop being reproducible.

**What synthesis receives.** [J] Give it the deduped claim ledger (merged across sessions by claim key, with grades, statuses and conflicts), per-session digests of about 600 words (conclusions, confidence, dissent, surprises, open items), the contradiction register, the open-question ledger and the delta report. Raw session outputs are available on demand through retrieval and are not preloaded. Full outputs cost five to ten times the tokens, and each carries its own narrative's smoothing. Claims alone lose argument structure, so the argument graph (conclusion-to-claim edges) and the qualifiers field travel with them.

**Preventing smoothed-over disagreements.**

1. **Detect before synthesis, with a process other than the synthesizer.** Group by claim key. The same (entity, attribute, as-of) with values beyond tolerance becomes a register entry. For unkeyed claims, run an entailment pass over claims that share entities. Each entry gets a typed cause hypothesis: definition, date, population or scope, methodology, incentive, error, or genuine uncertainty.
2. **Make coverage mechanical.** The report needs a "Disagreements and uncertainty" section that references every register ID, and the harness checks it. Omission then requires a stated reason.
3. **No averaging.** The writer must show both sides with sources, say which it favors and why (or "undecidable"), and name what would resolve it. "Somewhere between X and Y" requires a stated method.
4. **Dissent fields.** Each session digest records what doesn't fit the consensus, and synthesis must address it.
5. **Write the disagreements section first.** Fluency pressure pulls toward one story, so confront the conflicts before building it.
6. **Divergence check.** Run two independent syntheses in different input orders and diff their conclusions. Divergence marks fragile conclusions. This is self-consistency applied at the conclusion level (E+J). It costs double the synthesis tokens, so use it for top conclusions only.
7. **Confidence rule.** A conclusion that touches a contested claim inherits contested confidence unless the writer records the resolution reasoning.

## F. Failure catalog

Ordering is my estimate of frequency × impact, informed by the failure studies where they speak to it. Those studies measure different populations (commercial reports versus scientific agents), so instrument your own system. Rows 1–7 are the ones I'd instrument first.

| # | Failure | Structural countermeasure |
|---|---|---|
| 1 | Unsupported claims: recall presented as research. One OPPO taxonomy built from about 1,000 reports found fabricated rigor was the largest single coded failure, near 19% | Harness-computed grades, the claim→span→fetch-log chain, closed-evidence writer, quarantined recall lane, and a no-search control that must fail |
| 2 | Citation–claim mismatch: real URL, source doesn't say it | Span-only independent verifier, numeric and entity containment check, Tier 0 re-fetch audit |
| 3 | Flaw spotted in review, shipped anyway (82.5% of trajectories) | Review emits structured, severity-tagged, claim-anchored findings. Critical ones must produce a ledger diff or the report is refused (G7) |
| 4 | Overclaiming, concealed negatives and omitted limitations (about 78% and 62% of trajectories) | Confidence capped by weakest dependency. Unknowns and Contested sections generated from ledger and register. Negative results logged in the store |
| 5 | Confirmation-biased search and frame-lock | Examiner questions and competing hypotheses from an independent agent, leading-ness lint, mandatory falsification queries. For load-bearing questions, a hypotheses-by-evidence matrix that weights diagnostic evidence (analysis of competing hypotheses, from intelligence tradecraft) |
| 6 | Shallow search and premature closure | State-based done plus the depth ladder. A no-result protocol: at least 3 reformulations of different types (jargon, source type, language or framing) before "unanswerable." Saturation test |
| 7 | Retrieved evidence that never reaches the conclusion, or counter-evidence acknowledged then ignored | Conclusion→claim dependency graph. Lint that every `contradicts` relation in the store is addressed in the report. Register coverage check |
| 8 | Source-quality failures: SEO farms, vendor marketing, aggregators over primaries | Source strategy by claim type, domain-class priors, lateral reading for unfamiliar sources (what do independent sources say about this one?), grade caps by class |
| 9 | Circular sourcing and false corroboration | Trace to the terminal source, origin groups, corroboration counts distinct origins only |
| 10 | Snippet-level reading and silent tool failure (paywall, 403, truncated PDF treated as read) | Depth levels in the store with grade caps, fetch status and truncation flags, claims citing failed fetches rejected |
| 11 | Staleness and temporal confusion | Volatility classes and freshness windows, harness-extracted page dates, as-of fields, conflicting-date check |
| 12 | Error propagation across sessions | Grades immutable across injection, no upgrade by repetition, inherited-claim audit, correction log |
| 13 | Smoothing and false consensus in synthesis | Register built before synthesis, coverage check, no-averaging rule, disagreements written first |
| 14 | Context overload degrading synthesis | Extraction at the edge, state rebuilt per round, hierarchical synthesis, on-demand span retrieval |
| 15 | Quantitative and entity errors: units, denominators, tiny samples, same-name entities, version conflation | Computations via a code tool, logged with input claim IDs. Unit, denominator and n fields on statistics. Canonical entity IDs and version fields in the claim key |
| 16 | Mis-decomposition: a perfect answer to the wrong question | Examiner questions, a plan pre-mortem, a rewarded `question_challenge` output, a replanning gate after each session |
| 17 | Sycophancy toward the project vision | Neutral rewrite of the vision before planning, a standing prior-art and kill-criteria session, an adversarial session before synthesis |
| 18 | Verifier blind spots and judge bias (self-preference, length) | Fresh context, span-only view, different model family where possible, human-labeled calibration set with kappa tracking, every finding must carry an anchor |
| 19 | Goodhart on your own gates: padded claims, strawman counterarguments, template-filled reviews | Load-bearing defined by dependency edges, specificity checks (non-template, claim-ID anchored), some checks held out, periodic human audit of metric versus expert agreement, a growing mutation set |
| 20 | Prompt injection and AI-generated source contamination | Retrieved text treated as data, extraction in a tool-less context with a fixed schema, flag instructions aimed at the model, origin tracing exposes AI-written echo chambers |
| 21 | Runaway effort and duplicated work. Anthropic's early agents spawned dozens of subagents for simple queries and searched endlessly for sources that didn't exist | Effort-scaling rules in the prompt, shared query log with dedupe, explicit task boundaries, unanswerable-after-k rule |
| 22 | Blind spots: paywalls, non-English sources, offline knowledge | Scope names locales and languages. An "invisible to this search" section states what couldn't be reached and the likely direction of bias |
| 23 | Honest but hollow output: well-formed, nothing new | Delta report, reader model in the prompt, S3 score, request fewer and deeper findings with mechanism, boundary and counterexample |

## What I'd build first, and how this could go wrong

**This week:**
- claim records with harness-computed grades and the Tier 0 re-fetch audit;
- the no-search negative control and a 10-defect mutation set;
- closed-book prior and the Delta section.

**Next:**
- the examiner agent (questions plus hidden rubric);
- the independent span-only verifier;
- binding review with the repair loop (G7).

**After:**
- the contradiction register with its coverage check;
- the red-team gate;
- the field guide and the scout/dig/break nodes;
- Tier 2 extraction and state rebuilds, if you own the tools.

Three risks, and the evidence that would tell you the design isn't working:

- **Bureaucracy.** A stack of gates can yield compliant, hollow research. The 800-run study notes that adding a score to a method-quality problem just creates a new target to game, and what's needed is verification the agent doesn't control. Keep some checks hidden, keep the exploration budget, and audit against blind expert comparisons.
- **Spend confounds.** Anthropic's analysis attributed 80% of variance on a browsing benchmark to token usage alone, and their multi-agent system used about 15× the tokens of chat. Much of any improvement after adding loops and verifiers will simply be extra spend. Run budget-matched ablations to learn whether the design itself helps. [E+J]
- **Kill criteria for this design.** The verifier catches under roughly 80% of injected mutations, so it's blind. The delta ratio doesn't rise against a single-prompt baseline at matched budget, so the loop is theater. Red-team attack success doesn't fall across repair cycles, so repair is cosmetic. The thresholds are arbitrary and meant to be tuned.

If it's useful, I can write the examiner, verifier and red-team prompts and the JSON schemas next.