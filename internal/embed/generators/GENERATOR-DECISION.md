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
1. [ID]-PLAN.md: Routing (profile table, door, lane, gate track, assumptions and flip conditions, requester leaning, adjacent decisions); per session: a header (### [ID]-S[n]: [Title]), a table (ID [ID]-S[n], role, depends, filename sessions/[ID]-S[n]-[slug].md) and its prompt in a `prompt` code fence; Closing: run sessions independently; if a one-way door's results are contested or weight-sensitive, rerun C on a second model; complete the ADR; a human reviews it before acceptance.
2. [ID]-[slug].md, the proposed ADR: YAML frontmatter with exactly these keys: id, title, status (proposed), door_type, date, confidence, evidence_refs, informed_by_sessions, supersedes, superseded_by, amends, review_trigger, review_date, prediction, tags, authored_by, human_reviewed (false), schema_version ("0.1.0"). Fill id, title, door_type, date, informed_by_sessions, review_trigger and tags; null or [] for the rest. Body: Context & Problem Statement → Evaluated Options (the requester's options plus the status quo, one hypothesis each: what would have to be true; pending L if none) → Decision Outcome (pending) → Rejected Alternatives & Tradeoffs (pending) → Failure Modes & Reversal Triggers. Seed review_trigger and failure modes from your flip conditions.

<!-- CORE:BEGIN — shared methodology kernel (must stay identical across GENERATOR.md, GENERATOR-DECISION.md, GENERATOR-COMPARISON.md) -->
## SESSION PROMPT TEMPLATE (fill {slots}; {a | b} = pick one; {filename} = the session's filename; keep the rest verbatim)
```prompt
# RESEARCH BRIEF: {title}

## DECISION
{neutral decision question}; informs {ID}. Door: {one-way | two-way}. What would change the recommendation: {reversal condition}.

## BRIEF
{C: Compare {options} for {neutral decision} | L: Map the credible approaches to {neutral decision} | F: Find how {candidates} could fail for {neutral decision}}. Question type: {lookup | causal | evaluative}. Audience: a principal architect needing production-grade tradeoffs. Context (requester-stated, unverified): {context}.

## SCOPE
Today is {date}; focus on the last 18-24 months and flag older sources as potentially stale. In scope: {topics implied by the context and criteria}. Out of scope: {exclusions}. Source strategy by claim type: performance claims need reproducible benchmarks, stability claims need changelogs, adoption claims need download/usage data. Prefer primary sources (official docs, RFCs, source code, reproducible benchmarks, postmortems) over blogs and vendor claims, and do not rest a conclusion on one vendor's material.

## KNOWN
{If upstream sessions exist: [KNOWN_CONTEXT] — treat inherited claims as assumptions to stress-test, not settled facts. | If no upstream: Key assumptions this session starts with (requester-stated, unverified): {assumptions}.}

## CALIBRATION
1. Treat your memory as a hypothesis to test. Note initial beliefs; report what evidence confirmed, updated, or contradicted.
2. A significant claim is most valuable when grounded with a verifiable source. "Not found after searching X, Y, Z" is a valid, valued result.
3. An unsupported claim presented as fact costs more than an honest gap.
4. Before finishing: state the strongest objections an expert would raise and what evidence you found for or against each.
5. Distinguish what you found from what you recalled. Mark recalled claims honestly — they are starting points, not conclusions.

## APPROACH
Use the research move that fits each question:
- DEEPEN: trace a claim to its primary source
- WIDEN: search with different vocabulary or source types
- CORROBORATE: find independent support for a load-bearing claim
- FALSIFY: actively search for evidence against your leading answer
- PIVOT: reframe the question when results suggest the framing is wrong

If you've only found confirming evidence, try FALSIFY before concluding. Effort ceiling: spend effort proportional to the decision's reversibility.

Research is iterative. After your initial findings: Are there load-bearing claims with only one source? Seek a second. Did you find contradictions you haven't resolved? Investigate. Do your findings raise obvious follow-up questions? Pursue them. Have you only found confirming evidence? Try to find disconfirming evidence. When your evidence is sufficient to answer "what would change this recommendation?" — stop. Honest incompleteness with an attempt log is more valuable than false completeness.

Start broad, then trace the tradeoffs, failure modes and benchmarks that matter here. Frame queries neutrally, verify decisive requester premises, seek disconfirming evidence against whichever option leads, surface disagreements, state assumptions, and show which conclusions depend on parameters the context leaves unstated. If your research reveals critical concerns, dependencies, risks, or opportunities beyond the stated scope, investigate and include them. The stated scope defines the minimum — not the maximum — of what this session should cover. Justify any scope expansion with evidence.

## DONE
{C: This session is complete when:
1. A recommendation is stated with evidence for the primary question.
2. A weighted evaluation matrix is produced: criteria {supplied, weights normalized to 1.0 | 5-8 derived from the context and traceable to it, with why these}; a weight and one-line rationale each; 1-5 scores per option citing evidence (no evidence, no score); weighted totals; top-2 weights ±20%, flagging "weight-sensitive" if the ranking changes; reconcile score and qualitative analysis, explaining any divergence and which signal the recommendation follows (the score is a bias-correction lens, not the decision).
3. At least one concrete failure mode per top contender is documented with the strongest disconfirming evidence found.
4. "What would change this recommendation" is answered.
 | L: This session is complete when:
1. Every credible approach (including hybrids and the status quo) is mapped with in/out rationale.
2. Whether the framing holds and the criteria the context implies are stated.
3. A shortlist of at most 4 is produced.
4. "What would change this recommendation" is answered.
 | F: This session is complete when:
1. A premortem assuming the choice failed within 12 months is grounded with real postmortems and failure reports.
2. An exit or migration path and its cost are documented.
3. Reversal triggers with thresholds are stated.
4. "What would change this recommendation" is answered.}
Open risks and reversal triggers; if a decisive unknown depends on the requester's workload, the smallest probe that would settle it.
Discovered Concerns, including any stronger unlisted option (omit if none).
For evidence grading, every factual claim should carry:
- Base grade: A (official docs/RFCs/peer-reviewed studies) | B (empirical/benchmarks) | C (vendor claims) | D (blog/tutorial/AI recall) | E (unverifiable)
- Modifiers: corroboration (single/corroborated/contested), recency (fresh/aging/stale), directness (direct/indirect)
- Verification: fetched | cached | recalled | secondhand | human-provided (recalled claims capped at Grade D regardless of apparent source)

## FORMAT
Single Markdown file with YAML frontmatter (id, title, date, status, topic, tags, informs_decisions, confidence). Sections: Prior (pre-research beliefs: 3-5 checkable propositions "I believe [X] because [reasoning]") → Research Question → Key Findings (3-7 bullets) → Recommendation (isolated from rejected options) → Alternatives Considered → Detailed Findings → Open Questions & Risks → Delta (what research changed: table with Prior Belief | Status | Evidence | Impact) → Sources & Evidence Ledger. Filename: {filename} (id = its stem)
```
<!-- CORE:END -->

