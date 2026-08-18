# URP v3.0 — Pipeline Generator
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
a new software project. The output will be used to make evidence-grounded
architectural decisions before writing any application code.

The goal is to identify every critical architectural decision this project
requires, determine which are irreversible ("one-way doors" that need deep
research) vs. reversible ("two-way doors" that can be decided quickly), and
produce focused research prompts that investigate each decision with
appropriate rigor.

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
1. Classify as one-way door (irreversible: database, data model, auth,
   public APIs, regulatory compliance) or two-way door (reversible:
   UI framework, styling, CI tooling, utility libraries).
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

**DELIVERABLE:** Coverage checklist of what the output must address.
Include: recommendation, options evaluation, deep analysis of top contenders,
inline evidence grades, and open risks with reversal triggers.

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

1. **RESEARCH-PIPELINE.md** — Extracted parameters, complexity score with
   rationale, session matrix (DAG), execution plan (parallel groups +
   gated dependencies), and Phase 0 exit gate checklist.

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

## What You Get Back

The AI will return 3 documents. Save them in your new project's `research/` directory:

```
my-project/
└── research/
    ├── RESEARCH-PIPELINE.md    ← Your project's complexity score, session DAG, execution plan
    ├── PROMPT-LIBRARY.md       ← Copy-paste research prompts for each session
    └── DECISIONS.md            ← Initial hypothesis registry (fill in during research)
```

Then execute each prompt from PROMPT-LIBRARY.md in independent AI sessions. See [README.md](README.md) for the complete workflow.

---

## Design Notes

This generator prompt embodies URP v3.0 principles:

- **No persona** — task framing, not role assignment (personas debunked for factual accuracy: Zheng et al. EMNLP 2024)
- **Open-ended input** — accepts natural language vision dumps; the AI extracts structure
- **AI-driven classification** — domain, risk, and complexity are inferred from the vision, not self-reported by the user
- **Self-contained** — all methodology concepts are operationalized inline; the receiving AI doesn't need access to any other URP files
- **Uncertainty-native** — undecided elements become research questions, not blockers
- **Front-loaded** — complete brief in one turn (39% performance drop from drip-feeding: Laban et al. ICLR 2026)

For the complete methodology specification, see [FRAMEWORK.md](FRAMEWORK.md).
