# Vivechak v1.0 — Pipeline Generator
### The Tool: Generate a Complete Research Pipeline from Your Project Vision

> **How to use:** Copy the generator prompt below into a fresh AI conversation
> (Claude, Gemini, or ChatGPT with web search enabled). Paste your project
> description where indicated. Send. You'll receive 3 ready-to-execute documents.

---

## The Generator Prompt

````markdown
# GENERATE: Pre-Development Research Pipeline

## BRIEF

Generate a complete, ready-to-execute pre-development research pipeline for
a new technical project. The output will be used to make evidence-grounded
architectural decisions before committing to irreversible implementation decisions.

The goal is to identify every critical architectural decision this project
requires, determine which are irreversible ("one-way doors" that need deep
research) vs. reversible ("two-way doors" that can be decided quickly), and
produce focused research prompts that investigate each decision with
appropriate rigor.

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
  domains, identify primary and secondary archetypes.
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
| Domain Novelty | Standard CRUD | Established SaaS | Unconventional workflow | New category |
| Technical Novelty | Known stack | New library | New paradigm | Unproven infra |
| Regulatory Exposure | No sensitive data | Internal data | PII/GDPR/SOC2 | HIPAA/PCI/KYC |
| Reversibility | Throwaway | Modular | Core schema/multi-tenant | Deep platform |
| Investment Horizon | Weekend spike | Lean MVP | Funded venture | Enterprise-critical |
| Coordination Complexity | Single decision-maker | Small team alignment | Cross-team | Multi-org |
| Expected Longevity | Days–weeks | Months | 1–3 years | 5+ years |
| Integration Complexity | Standalone | 1–2 APIs | Multiple complex APIs | Regulated rails/ERP |

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
   authentication architecture, public API contracts, wire protocols,
   regulatory compliance, hardware selection, core language/runtime) or
   two-way door (reversible: UI framework, styling, CI tooling, utility
   libraries, IDE configuration, documentation format).
2. Route using this matrix:

   |  | Known Pattern | Unknown/Novel |
   |---|---|---|
   | **Reversible (two-way)** | SKIP — decide by convention | FAST SPIKE — 1 session |
   | **Irreversible (one-way)** | CONFIRM — 1 session + decision record | DEEP RESEARCH — 2–5 sessions + decision record |

3. Organize sessions as a DAG (Directed Acyclic Graph):
   - Layer 0: Landscape & Discovery (fully parallel, no dependencies)
   - Layer 1: Architectural Decisions (sparse dependencies on Layer 0)
   - Layer 2: Blueprints & Specifications (depend on Layer 1 decisions)
   - Sink: Grand Synthesis (depends on all Layer 2)
4. Default every session to UNBLOCKED unless it literally cannot produce
   valid output without another session's artifact.
5. If classification reveals critical architectural concerns the user
   didn't mention (e.g., overlooked regulatory exposure, scaling
   bottlenecks, security implications, operational complexity), add
   sessions for them. The project vision defines the starting point
   for discovery, not its boundary.

### Step 4: Write Research Prompts
For each session, write a complete, copy-paste-ready prompt using this
5-block structure:

**BRIEF:** What to investigate, which decision it informs, who the audience
is (a Principal Architect needing production-grade tradeoffs, not summaries).
Include project-specific context from the vision.

**SCOPE:** Today's date as temporal anchor. In-scope / out-of-scope boundaries.
Source priorities: prefer official docs, RFCs, source code, peer-reviewed
benchmarks over blog posts and SEO content.

**APPROACH:** Directional, not prescriptive. Tell the AI to start with broad
landscape queries, then dynamically investigate specific tradeoffs, failure
modes, and benchmarks. Tell it to surface disagreements rather than smooth
them, and actively seek disconfirming evidence. Do NOT prescribe specific
search queries or set minimum search counts.

Add this instruction to each prompt's APPROACH block: `If your research
reveals critical concerns, dependencies, risks, or opportunities not
listed in the coverage checklist, investigate and include them. The
stated scope defines the minimum --- not the maximum --- of what this
session should cover. Justify any scope expansion with evidence."

**DELIVERABLE:** Coverage checklist of what the output must address.
Include: recommendation, options evaluation, deep analysis of top contenders,
inline evidence grades, open risks with reversal triggers, and a
"Discovered Concerns" section if research reveals material concerns
beyond the stated scope (omit if nothing emerged).

For evidence grading, every factual claim should carry:
- Base grade: A (official docs/RFCs) | B (peer-reviewed/empirical) |
  C (vendor claims) | D (blog/tutorial/AI recall) | E (unverifiable)
- Modifiers: corroboration (single/corroborated/contested),
  recency (fresh/aging/stale), directness (direct/indirect)
- Verification: fetched | cached | recalled | secondhand | human-provided
  (recalled claims capped at Grade D regardless of apparent source)

**FORMAT:** Single complete Markdown file artifact with YAML frontmatter
(id, title, date, status, topic, tags, informs_decisions, confidence).
Body sections: Research Question → Key Findings (3–7 bullets) →
Recommendation (isolated from rejected options) → Alternatives Considered
→ Detailed Findings → Open Questions & Risks → Sources & Evidence Ledger.

IMPORTANT: Do NOT include expert personas, hardcoded search queries,
minimum search counts, or rigid output skeletons in any prompt. Each
prompt must be a complete, front-loaded, single-turn brief.

### Step 5: Seed Decision Registry
For each architectural decision identified, create a D-NNN entry with:
- Status: proposed
- Door type: one-way or two-way
- Initial competing hypotheses
- Sessions that will inform the verdict
- Review trigger (condition for future re-evaluation)

## DELIVERABLE

Generate THREE complete Markdown file artifacts:

1. **RESEARCH-PIPELINE.md** — Must include:
   - **How to Execute This Pipeline** section at the top with:
     - Where to save research outputs (`sessions/T#-##-[slug].md`)
     - How to record decisions (reference `templates/DECISIONS.template.md`)
     - How to resolve conflicts (reference `templates/CONFLICT-RESOLUTION.template.md`)
     - How to compile the FAD (reference `templates/FOUNDING-ARCHITECTURE.template.md`)
     - How to run the gate (reference `templates/PHASE-0-GATE.template.md`)
   - Extracted project parameters and inferred classifications
   - Complexity score with per-dimension rationale and tier assignment
   - Session matrix (DAG) with IDs, dependencies, door types
   - Execution plan (parallel groups + gated dependencies)
   - Phase 0 exit gate criteria (two-track: Track A for two-way, Track B for one-way)

2. **PROMPT-LIBRARY.md** — Complete, copy-paste-ready research prompts
   for every session. Each prompt self-contained with 5-block anatomy.

3. **DECISIONS.md** — Initial decision registry with proposed hypotheses,
   door types, and session links.

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

## Choose Your Approach

| Approach | Best For | Sessions |
|---|---|---|
| **Single Session** (above) | Agentic IDEs (Antigravity, Cursor, Claude Code), API agents, Tier 0–1 projects | 1 |
| **Split Generation** (below) | Web chat interfaces, credit-limited sessions, Tier 2–3 projects | 2 |

The PROMPT-LIBRARY scales linearly with session count — a Tier 3 project with 20 sessions
produces 15,000+ words of prompts alone. If your session can handle all 3 files at once, use
the single prompt above. If it can't, use the split approach below.

> **Why not 3 parallel sessions?** The 3 files are NOT independent. PROMPT-LIBRARY needs the
> session matrix from RESEARCH-PIPELINE. DECISIONS needs the session and decision IDs. Three
> independent sessions will produce inconsistent session lists and divergent decisions.
> The correct split is sequential: **foundation first, then prompts.**

---

## Split Generation (Alternative)

### Step 1: Generate Pipeline + Decisions

Use the **same generator prompt** above, but replace the `## DELIVERABLE` section with:

````markdown
## DELIVERABLE

Generate TWO complete Markdown file artifacts:

1. **RESEARCH-PIPELINE.md** — Must include:
   - **How to Execute This Pipeline** section at the top with:
     - Where to save research outputs (`sessions/T#-##-[slug].md`)
     - How to record decisions (reference `templates/DECISIONS.template.md`)
     - How to resolve conflicts (reference `templates/CONFLICT-RESOLUTION.template.md`)
     - How to compile the FAD (reference `templates/FOUNDING-ARCHITECTURE.template.md`)
     - How to run the gate (reference `templates/PHASE-0-GATE.template.md`)
   - Extracted project parameters and inferred classifications
   - Complexity score with per-dimension rationale and tier assignment
   - Session matrix (DAG) with IDs, titles, dependencies, door types
   - Execution plan (parallel groups + gated dependencies)
   - Phase 0 exit gate criteria (two-track: Track A for two-way, Track B for one-way)

2. **DECISIONS.md** — Initial decision registry with proposed hypotheses,
   door types, and session links.

Do NOT generate PROMPT-LIBRARY.md — it will be generated in a follow-up session.
````

### Step 2: Generate Prompts from Pipeline

Open a **fresh AI session** and send this prompt with both paste sections filled in:

````markdown
# GENERATE: Research Prompts from Pipeline

## BRIEF

Generate the complete PROMPT-LIBRARY.md for a pre-development research pipeline.
The Research Pipeline and Decision Registry have already been generated (provided
below). Your job is to write a complete, copy-paste-ready research prompt for
every session listed in the pipeline.

## PROJECT VISION

[PASTE YOUR ORIGINAL PROJECT DESCRIPTION HERE]

## RESEARCH PIPELINE

[PASTE YOUR GENERATED RESEARCH-PIPELINE.md HERE]

## PROMPT GENERATION RULES

For each session in the Research Pipeline above, write a self-contained prompt
using this 5-block structure:

**BRIEF:** What to investigate, which decision it informs, who the audience
is (a Principal Architect needing production-grade tradeoffs, not summaries).
Include project-specific context from the vision.

**SCOPE:** Today's date as temporal anchor. In-scope / out-of-scope boundaries.
Source priorities: prefer official docs, RFCs, source code, peer-reviewed
benchmarks over blog posts and SEO content.

**APPROACH:** Directional, not prescriptive. Tell the AI to start with broad
landscape queries, then dynamically investigate specific tradeoffs, failure
modes, and benchmarks. Tell it to surface disagreements rather than smooth
them, and actively seek disconfirming evidence. Do NOT prescribe specific
search queries or set minimum search counts.

Add this instruction to each prompt's APPROACH block: `If your research
reveals critical concerns, dependencies, risks, or opportunities not
listed in the coverage checklist, investigate and include them. The
stated scope defines the minimum --- not the maximum --- of what this
session should cover. Justify any scope expansion with evidence."

**DELIVERABLE:** Coverage checklist of what the output must address.
Include: recommendation, options evaluation, deep analysis of top contenders,
inline evidence grades, open risks with reversal triggers, and a
"Discovered Concerns" section if research reveals material concerns
beyond the stated scope (omit if nothing emerged).

For evidence grading, every factual claim should carry:
- Base grade: A (official docs/RFCs) | B (peer-reviewed/empirical) |
  C (vendor claims) | D (blog/tutorial/AI recall) | E (unverifiable)
- Modifiers: corroboration (single/corroborated/contested),
  recency (fresh/aging/stale), directness (direct/indirect)
- Verification: fetched | cached | recalled | secondhand | human-provided
  (recalled claims capped at Grade D regardless of apparent source)

**FORMAT:** Single complete Markdown file artifact with YAML frontmatter
(id, title, date, status, topic, tags, informs_decisions, confidence).
Body sections: Research Question → Key Findings (3–7 bullets) →
Recommendation → Alternatives Considered → Detailed Findings →
Open Questions & Risks → Sources & Evidence Ledger.

IMPORTANT: Do NOT include expert personas, hardcoded search queries,
minimum search counts, or rigid output skeletons in any prompt. Each
prompt must be a complete, front-loaded, single-turn brief.

## DELIVERABLE

Generate a single complete Markdown file artifact:

**PROMPT-LIBRARY.md** — All research prompts for every session listed in
the Research Pipeline, organized by layer (Layer 0 → Layer 1 → Layer 2 → Sink).
Each prompt must be copy-paste-ready into a fresh AI session.

## SCOPE
- Today's date: [INSERT DATE]
- Generate prompts appropriate for frontier AI with deep research/web
  search capabilities (Claude, Gemini, ChatGPT)
- Match session IDs and titles exactly as they appear in the Research Pipeline
- Each prompt must be self-contained (usable without reading other prompts)
````

---

## Set Up Your Project

After the AI returns the generated documents, set up your new project workspace:

### 1. Create the research directory
```
my-project/
└── research/
    ├── RESEARCH-PIPELINE.md         ← Generated (paste here)
    ├── PROMPT-LIBRARY.md            ← Generated (paste here)
    ├── DECISIONS.md                 ← Generated (paste here)
    ├── sessions/                    ← Create empty folder for research outputs
    └── templates/                   ← Copy from Vivechak (see step 2)
```

### 2. Copy the operational templates
Copy the 4 templates from this repository into your project's `research/templates/` directory:

```
templates/DECISIONS.template.md
templates/CONFLICT-RESOLUTION.template.md
templates/FOUNDING-ARCHITECTURE.template.md
templates/PHASE-0-GATE.template.md
```

**Why:** These templates are the contracts for recording decisions, resolving conflicts, compiling the final architecture, and running the exit gate. With them in your workspace, any AI agent (Antigravity, Claude Code, Cursor) can autonomously execute the full research workflow without referencing the meta-repo.

### 3. Execute the pipeline
See the generated RESEARCH-PIPELINE.md for the complete execution guide, or follow the [README Quick Start](README.md).

---

## Using with AI Agents (Antigravity, Claude Code, etc.)

With templates in your workspace, you can delegate research steps directly to an AI agent:

**Execute a research session:**
> *Read session T2-01 from `research/PROMPT-LIBRARY.md`. Run the deep research with web search, grade all evidence, and save the result to `research/sessions/T2-01-datastore-selection.md`.*

**Record a decision:**
> *Read `research/sessions/T2-01-datastore-selection.md`. Formulate the verdict for D-001 in `research/DECISIONS.md` following the format in `research/templates/DECISIONS.template.md`.*

**Compile the FAD:**
> *Read all finalized sessions in `research/sessions/` and locked decisions in `research/DECISIONS.md`. Synthesize them into `FAD.md` following `research/templates/FOUNDING-ARCHITECTURE.template.md`.*

**Run the gate:**
> *Audit `FAD.md` and `research/DECISIONS.md` against `research/templates/PHASE-0-GATE.template.md`. Conduct the premortem for all One-Way Doors and emit `PHASE-0-GATE.md`.*

---

## Design Notes

This generator prompt embodies Vivechak v1.0 principles:

- **No persona** — task framing, not role assignment (personas debunked: Zheng et al. EMNLP 2024)
- **Open-ended input** — accepts natural language vision dumps; the AI extracts structure
- **AI-driven classification** — domain, risk, and complexity inferred from vision, not self-reported
- **Self-contained** — all methodology operationalized inline; receiving AI needs no other Vivechak files
- **Self-documenting output** — generated RESEARCH-PIPELINE.md includes its own execution guide
- **Uncertainty-native** — undecided elements become research questions, not blockers
- **Front-loaded** — complete brief in one turn (39% drop from drip-feeding: Laban et al. ICLR 2026)

For the complete methodology specification, see [FRAMEWORK.md](FRAMEWORK.md).

