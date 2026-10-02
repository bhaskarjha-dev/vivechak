# Vivechak v0.1.0 — Project-Scope Pipeline Generator
### Generate a Complete Research Pipeline from Your Project Vision

> **How to use:** Copy the generator prompt below into a fresh AI conversation
> (Claude, Gemini, or ChatGPT with web search enabled). Paste your project
> description where indicated. Send. You'll receive two ready-to-execute documents:
> a unified research pipeline with inline prompts, and a decision registry.
>
> **Smaller scope?** For a single decision, use [GENERATOR-DECISION.md](GENERATOR-DECISION.md).
> For a bounded comparison of named options, use [GENERATOR-COMPARISON.md](GENERATOR-COMPARISON.md).

---

## The Generator Prompt

````markdown
# GENERATE: Research Pipeline for a Technical Project

## BRIEF

Generate a complete, ready-to-execute research pipeline for
a new technical project. The output will be used to make evidence-grounded
architectural decisions before committing to irreversible implementation decisions.

The goal is to identify every critical architectural decision this project
requires, determine which are irreversible ("one-way doors" that need deep
research) vs. reversible ("two-way doors" that can be decided quickly), and
produce a unified research pipeline with copy-paste-ready prompts for every
session, plus a separate decision registry that evolves as research progresses.

You are not constrained to only the decisions and concerns explicitly
mentioned in the project vision. If your analysis reveals critical
architectural concerns, risks, dependencies, or opportunities that the
user did not mention, include research sessions for them. The project
vision defines the starting point, not the ceiling.

## PROJECT VISION

[PASTE YOUR PROJECT DESCRIPTION HERE]

Write everything you know about your project in natural language. Include:
- What you're building and why
- Who it's for (target users, market)
- What you believe makes it unique or different (if you know)
- Any constraints you're aware of (regulations, geography, budget, timeline)
- Any strong technical opinions or preferences you hold
- Anything you're uncertain about or want to explore

There is no required format. Write as much or as little as you have.
If you haven't decided something (project name, platform, tech stack),
say so — undecided elements become research questions in the pipeline.

## METHODOLOGY

Follow these rules when generating the pipeline:

### Step 1: Extract & Classify
From the project vision above:
- Infer the primary domain archetype (B2B SaaS, DevTools, FinTech, AI/ML,
  Consumer Mobile, Real-Time/IoT, or hybrid). If the project spans multiple
  domains, identify primary and secondary archetypes. If the project represents
  an unlisted or novel domain, synthesize a custom archetype defining its
  unique operational and compliance boundaries.
- Identify all stated constraints (regulatory, geographic, infrastructure).
- Identify what the user has decided vs. what remains open. Treat open
  elements as research questions to include in the pipeline.
- Assess regulatory exposure independently — don't trust self-reported risk
  labels. If the vision mentions payments, health data, children, financial
  transactions, or PII, flag elevated compliance requirements regardless
  of how the user describes the project.

### Step 2: Score Complexity (0–24)
Score the project across these 8 dimensions (0 = minimal, 3 = extreme):

| Dimension | 0 | 1 | 2 | 3 |
|---|---|---|---|---|
| Domain Novelty | Standard pattern (e.g., CRUD) | Established category | Unconventional workflow | New category |
| Technical Novelty | Known stack | New component | New paradigm | Unproven infra |
| Regulatory Exposure | No sensitive data | Internal data | PII/GDPR/SOC2 | HIPAA/PCI/KYC |
| Reversibility | Throwaway | Modular | Core schema/multi-tenant | Deep platform |
| Investment Horizon | Weekend spike | Lean MVP | Funded venture | Enterprise-critical |
| Coordination Complexity | Single decision-maker | Small team alignment | Cross-team | Multi-org |
| Expected Longevity | Days–weeks | Months | 1–3 years | 5+ years |
| Integration Complexity | Standalone | 1–2 external interfaces | Multiple complex integrations | Regulated rails/legacy systems |

Show the per-dimension score and rationale. Map total to tier:
- 0–4 → Tier 0 (1–3 sessions)
- 5–9 → Tier 1 (4–8 sessions)
- 10–15 → Tier 2 (9–16 sessions)
- 16–24 → Tier 3 (17–30 sessions)

HARD OVERRIDE: If Regulatory Exposure = 3, all intersecting decisions
automatically receive deep research regardless of total score.

### Step 3: Build Session Matrix
For each architectural decision the project requires:
1. Classify as one-way door (irreversible: primary datastore, data model,
   authentication architecture, core language/runtime, public API contracts,
   wire protocols, regulatory compliance, hosting infrastructure,
   deployment architecture, hardware selection, inter-service communication
   patterns, event sourcing vs CRUD, monolith vs microservices,
   data partitioning strategy) or
   two-way door (reversible: UI framework, styling, CI tooling, utility
   libraries, IDE configuration, documentation format, logging providers,
   feature flag tooling).
2. Route using this matrix:

   |  | Known Pattern | Unknown/Novel |
   |---|---|---|
   | **Reversible (two-way)** | SKIP — decide by convention | FAST SPIKE — 1 session |
   | **Irreversible (one-way)** | CONFIRM & COMMIT — 1 session + ADR | DEEP RESEARCH — 2–5 sessions + ADR + premortem |

3. Organize sessions as a DAG (Directed Acyclic Graph):
   - Layer 0: Landscape & Discovery (fully parallel, no dependencies)
   - Layer 1: Architectural Decisions (sparse dependencies on Layer 0)
   - Layer 2: Blueprints & Specifications (depend on Layer 1 decisions)
   - Sink: Grand Synthesis (depends on all Layer 2)
4. Default every session to UNBLOCKED unless it literally cannot produce
   valid output without another session's artifact. Minimize dependencies
   aggressively — most Layer 0 and Layer 1 sessions can run in parallel.
   When a dependency IS required, specify:
   - **What specific information** from the upstream session is needed
     (not "attach the full research output")
   - Whether the upstream output **constrains** the downstream session
     (decision must be respected) or merely **informs** it (context that
     should be considered but not treated as a constraint)
5. For soft dependencies (contextual, non-blocking), note the upstream
   session as "Context: [session ID]" with a one-line summary of what
   to inject if available (e.g., "Context from T1-01: selected database
   technology and rationale"). The downstream session must produce valid
   output even without this context.
6. If classification reveals critical architectural concerns the user
   didn't mention (e.g., overlooked regulatory exposure, scaling
   bottlenecks, security implications, operational complexity), add
   sessions for them. The project vision defines the starting point
   for discovery, not its boundary.

### Step 4: Write Session Entries
For each session in the DAG, write a complete entry containing three parts:

**Part A — Session Metadata:**
A table with: session ID, title, layer, door type, decision ID, dependencies
(with type: hard/soft and what information is needed), and output filename
(`[Session-ID]-[slug].md`, e.g., `T1-01-primary-datastore-selection.md`).

**Part B — Decision Reference:**
For each architectural decision this session informs, note:
- Decision ID, title, and door type
- The competing hypotheses being evaluated
- Note: The full decision record lives in DECISIONS.md, not here

**Part C — Research Prompt (copy-paste ready):**
Write the complete research prompt using the 8-block structure:

**DECISION:** What architectural decision this session feeds, stated as a
neutral question. Include: the decision ID (D-NNN), the door type (one-way
or two-way), and what evidence would change the recommendation. This
anchors every session to its purpose.

**BRIEF:** What to investigate, typed by question kind (lookup, causal, or
evaluative). Who the audience is (a Principal Architect needing
production-grade tradeoffs, not summaries). Include project-specific
context from the vision.

**SCOPE:** Today's date as temporal anchor. In-scope / out-of-scope
boundaries. Source strategy BY CLAIM TYPE — e.g., performance claims need
reproducible benchmarks, API stability claims need changelogs, adoption
claims need download/usage data. Prefer official docs, RFCs, source code,
peer-reviewed benchmarks over blog posts and SEO content.

**KNOWN:** For sessions that depend on upstream sessions, include the
`[KNOWN_CONTEXT]` placeholder here. Tag upstream claims as `INHERITED` —
they are starting assumptions to stress-test, not settled facts. For
sessions with no upstream dependencies, state the key assumptions the
session starts with and note they are unverified.

**CALIBRATION:** Include this block verbatim in every research session prompt:
```
1. Treat your memory as a hypothesis to test. Note initial beliefs;
   report what evidence confirmed, updated, or contradicted.
2. A significant claim is most valuable when grounded with a verifiable
   source. "Not found after searching X, Y, Z" is a valid, valued result.
3. An unsupported claim presented as fact costs more than an honest gap.
4. Before finishing: state the strongest objections an expert would raise
   and what evidence you found for or against each.
5. Distinguish what you found from what you recalled. Mark recalled claims
   honestly — they are starting points, not conclusions.
```

**APPROACH:** Directional, not prescriptive. Provide the research moves
vocabulary — the model picks which moves fit each question:
- DEEPEN: trace a claim to its primary source
- WIDEN: search with different vocabulary or source types
- CORROBORATE: find independent support for a load-bearing claim
- FALSIFY: actively search for evidence against your leading answer
- PIVOT: reframe the question when results suggest the framing is wrong

If you've only found confirming evidence, try FALSIFY before concluding.
Effort ceiling: spend effort proportional to the decision's reversibility.

Include this iterative guidance: "Research is iterative. After your initial
findings: Are there load-bearing claims with only one source? Seek a
second. Did you find contradictions you haven't resolved? Investigate.
Do your findings raise obvious follow-up questions? Pursue them.
Have you only found confirming evidence? Try to find disconfirming evidence.
When your evidence is sufficient to answer 'what would change this
recommendation?' — stop. Honest incompleteness with an attempt log
is more valuable than false completeness."

Add this instruction: "If your research reveals critical concerns,
dependencies, risks, or opportunities beyond the stated scope,
investigate and include them. The stated scope defines the minimum —
not the maximum — of what this session should cover. Justify any scope
expansion with evidence."

Do NOT prescribe specific search queries or set minimum search counts.

**DONE:** State 3-5 completion criteria that define when this session
is DONE. Frame each as "This session is complete when [X] is answered with
[evidence type]." These are quality goals, not method prescriptions — the
model decides HOW to meet them. Include criteria for: the primary question
answered with evidence, trade-offs grounded with verifiable data, at least
one failure mode per major option documented, and the ability to state what
would change the recommendation. For comparison sessions, also require a
weighted evaluation matrix with project-derived criteria, evidence
references per score, and a sensitivity check. Add a "Discovered Concerns"
criterion if research reveals material concerns beyond stated scope.

<!-- CORE:BEGIN — shared methodology kernel (must stay identical across GENERATOR.md, GENERATOR-DECISION.md, GENERATOR-COMPARISON.md) -->
For evidence grading, every factual claim should carry:
- Base grade: A (official docs/RFCs/peer-reviewed studies) | B (empirical/benchmarks) |
  C (vendor claims) | D (blog/tutorial/AI recall) | E (unverifiable)
- Modifiers: corroboration (single/corroborated/contested),
  recency (fresh/aging/stale), directness (direct/indirect)
- Verification: fetched | cached | recalled | secondhand | human-provided
  (recalled claims capped at Grade D regardless of apparent source)
<!-- CORE:END -->

**FORMAT:** Single complete Markdown file artifact with YAML frontmatter
(id, title, date, status, topic, tags, informs_decisions, confidence).
**Filename:** `[Session-ID]-[slug].md` (use the output filename from the
session metadata table).
Body sections: Prior (pre-research beliefs) → Research Question →
Key Findings (3–7 bullets) → Recommendation (isolated from rejected
options) → Alternatives Considered → Detailed Findings →
Open Questions & Risks → Delta (what research changed) →
Sources & Evidence Ledger.

The Prior section should contain 3-5 checkable propositions:
"I believe [X] because [reasoning]."

The Delta section is a table:
| Prior Belief | Status | Evidence | Impact |
|---|---|---|---|
| [X] | Confirmed/Updated/Contradicted | [source] | [high/med/low] |

These sections define required coverage areas, not rigid paragraph
templates. Within each section, let the AI organize content naturally.

IMPORTANT: Do NOT include expert personas, hardcoded search queries,
or minimum search counts in any prompt. Each prompt must be a complete,
front-loaded, single-turn brief.

**Context Injection Placeholders:**
- For any downstream session that depends on upstream sessions, include the exact placeholder `[UPSTREAM_FINDINGS]` inside the prompt where upstream findings should be injected. Also include `[KNOWN_CONTEXT]` in the KNOWN block for structured upstream context.
- For the final grand synthesis session (`SYN-01`), include the exact placeholder `[ALL_SESSION_FINDINGS]` inside the prompt where all session findings should be aggregated.

## DELIVERABLE

Generate **two** complete Markdown file artifacts:

### 1. RESEARCH-PIPELINE.md

A unified research pipeline document with this structure:

```
# Research Pipeline: [Project Name]
## Generated by Vivechak v0.1.0

## Pipeline Overview
  ### Project Parameters         <- Extracted classifications, archetype, constraints
  ### Complexity Score           <- 8-dimension table, tier assignment
  ### Execution DAG              <- Mermaid diagram + parallel execution groups
  ### How to Execute             <- Template references, output conventions, gate instructions

## Research Sessions
  ### T1-01: [Title]
    Session Metadata table       <- ID, layer, door type, decision, dependencies, output filename
    Decision context             <- Which D-NNN this session informs (read-only reference)
    Research Prompt              <- Complete 8-block prompt in a `prompt code fence
  ### T1-02: [Title]             <- Same structure repeats for each session
    ...
  ### SYN-01: Grand Synthesis    <- Final synthesis session

## Phase 0 Exit Gate Criteria    <- Two-track criteria (Track A + Track B)
```

Requirements:
- Each session's research prompt must be inside a fenced code block
  (`prompt) to prevent Markdown header collision and to make it
  trivially copy-pasteable into a fresh AI session
- Downstream session prompts depending on prior sessions must include
  the `[UPSTREAM_FINDINGS]` placeholder; the final grand synthesis prompt
  (`SYN-01`) must include the `[ALL_SESSION_FINDINGS]` placeholder
- Sessions must be ordered by layer (Layer 0 -> Layer 1 -> Layer 2 -> Sink)
- The Pipeline Overview section must include a "How to Execute" guide
  referencing: templates/DECISIONS.template.md, templates/CONFLICT-RESOLUTION.template.md,
  templates/COMPARISON-SESSION.template.md, templates/FOUNDING-ARCHITECTURE.template.md,
  and templates/PHASE-0-GATE.template.md.
  The guide must also include:
  - **Session normalization note:** A session is one independent deep
    research execution (a single prompt sent to a frontier AI's deep
    research mode). If using standard chat, multiple turns may be needed
    per session. Tier budgets are relative proportions, not absolute counts.
  - **Troubleshooting guidance:** What to do when a session produces
    vague or off-topic output (try a different model, split into focused
    sub-sessions), when sessions produce contradictory recommendations
    (use CONFLICT-RESOLUTION template), and when the exit gate reveals
    gaps (spawn targeted follow-up sessions, don't re-run entire pipeline).
- Each session includes a reference to the decision(s) it informs, but
  decisions are NOT recorded inline -- they live in DECISIONS.md
- Output filenames for each session must follow the convention:
  sessions/[Session-ID]-[slug].md

### 2. DECISIONS.md

A decision registry seeded with initial hypotheses. This is a **living
document** that evolves as research sessions are executed:

- For each architectural decision identified, create a D-NNN entry wrapped in HTML decision comments (`<!-- DECISION: D-NNN --> ... <!-- /DECISION: D-NNN -->`) with YAML frontmatter using this schema:

  ```markdown
  <!-- DECISION: D-NNN -->
  ---
  id: D-NNN
  title: "[Decision Title]"
  status: proposed           # proposed | accepted | rejected | deprecated | superseded
  door_type: one-way         # one-way | two-way
  date: YYYY-MM-DD
  confidence: medium         # high | medium | low
  evidence_refs: []          # Optional: E-NNN IDs if you maintain a separate evidence ledger
  informed_by_sessions: []   # Session IDs that informed this decision (primary traceability)
  supersedes: null           # D-NNN ID this supersedes, or null
  superseded_by: null        # D-NNN ID that supersedes this, or null
  amends: null               # D-NNN ID this partially updates, or null
  review_trigger: "[Condition or date for mandatory re-evaluation]"
  review_date: null          # YYYY-MM-DD — set when decision is locked, BEFORE outcome is known
  prediction: null           # Optional: predicted outcome at decision time (for calibration)
  tags: []
  authored_by: "[agent-id or human name]"
  human_reviewed: false      # Mandatory true for one-way doors before acceptance
  schema_version: "0.1.0"
  ---

  # D-NNN: [Decision Title]

  ## Initial Context & Hypotheses
  [Initial rationale, competing options, and research direction]
  <!-- /DECISION: D-NNN -->
  ```

  At generation time, populate: `id`, `title`, `status: proposed`,
  `door_type`, `date`, `informed_by_sessions`, `review_trigger`, and
  initial competing hypotheses in the body. Leave remaining fields as
  null/empty — they are populated during pipeline execution.
- Group entries by decision domain (data, auth, infra, etc.)

**Why a separate file:** Decisions evolve throughout the research pipeline
(from proposed -> researched -> accepted/rejected). The pipeline is a static
execution plan; decisions are a living registry. Coupling them would require
editing the pipeline mid-execution.
## SCOPE
- Today's date: [INSERT DATE]
- Generate prompts appropriate for frontier AI with deep research/web
  search capabilities (Claude, Gemini, ChatGPT)
- Scale session count strictly to the tier determined by complexity scoring
- If the project vision mentions undecided elements, include research
  sessions to investigate those open questions
- Compose 100% of commodity infrastructure from established providers;
  build 100% custom only for proprietary domain logic
````

---

## Set Up Your Project

After the AI returns the generated pipeline, set up your new project workspace:

### 1. Create the research directory
```
my-project/
└── research/
    ├── RESEARCH-PIPELINE.md         ← Generated (paste here)
    ├── DECISIONS.md                 ← Generated (paste here)
    ├── sessions/                    ← Create empty folder for research outputs
    └── templates/                   ← Copy from Vivechak (see step 2)
```

### 2. Copy the operational templates
Copy the 5 templates from this repository into your project's `research/templates/` directory:

```
templates/DECISIONS.template.md
templates/CONFLICT-RESOLUTION.template.md
templates/COMPARISON-SESSION.template.md
templates/FOUNDING-ARCHITECTURE.template.md
templates/PHASE-0-GATE.template.md
```

**Why:** These templates are the contracts for recording decisions, resolving conflicts, structuring comparisons, compiling the final architecture, and running the exit gate. With them in your workspace, any AI agent (Antigravity, Claude Code, Cursor) can autonomously execute the full research workflow without referencing the meta-repo.

### 3. Execute the pipeline
See the generated RESEARCH-PIPELINE.md for the complete execution guide, or follow the [README Quick Start](README.md).

---

## Using with AI Agents (Antigravity, Claude Code, etc.)

With templates in your workspace, you can delegate research steps directly to an AI agent:

**Execute a research session:**
> *Read session T2-01 from `research/RESEARCH-PIPELINE.md`. Copy the research prompt and run the deep research with web search. Grade all evidence and save the result to `research/sessions/T2-01-datastore-selection.md`.*

**Record a decision:**
> *Read `research/sessions/T2-01-datastore-selection.md`. Formulate the verdict for D-001 following `research/templates/DECISIONS.template.md` and record it in `research/DECISIONS.md`.*

**Compile the FAD:**
> *Read all finalized sessions in `research/sessions/` and all locked decisions. Synthesize them into `FAD.md` following `research/templates/FOUNDING-ARCHITECTURE.template.md`.*

**Run the gate:**
> *Audit `FAD.md` and all decisions against `research/templates/PHASE-0-GATE.template.md`. Conduct the premortem for all One-Way Doors and emit `PHASE-0-GATE.md`.*

---

## Design Notes

This generator prompt embodies Vivechak v0.1.0 principles:

- **No persona** — task framing, not role assignment (personas debunked: Zheng et al. EMNLP 2024)
- **Open-ended input** — accepts natural language vision dumps; the AI extracts structure
- **AI-driven classification** — domain, risk, and complexity inferred from vision, not self-reported
- **Self-contained** — all methodology operationalized inline; receiving AI needs no other Vivechak files
- **Lifecycle-aware output** — pipeline + prompts are static (read-only plan); decisions are a separate living registry that evolves during execution
- **Minimum dependencies** — sessions default to unblocked; dependencies specify what information is needed and whether it constrains or merely informs
- **Named outputs** — each session specifies its output filename for immediate artifact naming
- **Self-documenting** — generated pipeline includes its own execution guide
- **Uncertainty-native** — undecided elements become research questions, not blockers
- **Front-loaded** — complete brief in one turn (39% drop from drip-feeding: Laban et al. ICLR 2026)

For the complete methodology specification, see [FRAMEWORK.md](FRAMEWORK.md).
