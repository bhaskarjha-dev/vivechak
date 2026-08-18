# Universal Research Pipeline (URP) v3.0
### Evidence-Grounded Pre-Development Research Meta-Framework

> **Transform software architecture decisions from gut-feel and cached training data into structured, evidence-graded, empirically validated research — before a single line of application code is written.**

---

## What Is This?

You have a software project idea. Before coding, you need to make architectural decisions — database, auth, hosting, data model, APIs. Some of these decisions are **irreversible** (one-way doors). If you get them wrong, you're looking at a rewrite.

**URP generates a customized research pipeline for your project** that investigates every critical decision with appropriate rigor. It produces focused research prompts, executes them across frontier AI platforms, grades every claim with traceable evidence, and synthesizes the results into a sealed Founding Architecture Document (FAD) — your architectural source of truth before writing code.

---

## Quick Start (5 Steps)

### Step 1: Generate Your Pipeline
Open [GENERATOR.md](GENERATOR.md), copy the generator prompt, paste your project description where indicated, and send it to a frontier AI (Claude, Gemini, or ChatGPT with web search). You'll receive 3 files:

- **RESEARCH-PIPELINE.md** — your project's complexity score, session DAG, and execution plan
- **PROMPT-LIBRARY.md** — copy-paste research prompts for each session
- **DECISIONS.md** — initial hypothesis registry

Save these in your new project's `research/` directory.

### Step 2: Execute Research Sessions
Take each prompt from your generated PROMPT-LIBRARY.md and run it in an independent AI deep research session. Sessions are designed to be parallel and independent — run as many simultaneously as you want.

Save each output as `research/sessions/T#-##-[slug].md` in your project directory.

### Step 3: Record Decisions
After key research sessions, lock architectural decisions using the [DECISIONS template](templates/DECISIONS.template.md). Classify each as one-way or two-way door. If models disagree on a contested decision, use the [CONFLICT-RESOLUTION template](templates/CONFLICT-RESOLUTION.template.md).

### Step 4: Synthesize & Gate
Compile all findings into a Founding Architecture Document using the [FAD template](templates/FOUNDING-ARCHITECTURE.template.md). Verify exit criteria with the [PHASE-0-GATE template](templates/PHASE-0-GATE.template.md).

### Step 5: Build
Start development with the FAD as your architectural source of truth. The research pipeline's job is done.

### What Your New Project Directory Looks Like

```
my-new-project/
├── research/
│   ├── RESEARCH-PIPELINE.md         ← Generated in Step 1
│   ├── PROMPT-LIBRARY.md            ← Generated in Step 1
│   ├── DECISIONS.md                 ← Seeded in Step 1, filled in Step 3
│   ├── sessions/
│   │   ├── T1-01-market-landscape.md
│   │   ├── T2-01-datastore-selection.md
│   │   └── ...
│   ├── FAD.md                       ← Created in Step 4
│   └── PHASE-0-GATE.md              ← Completed in Step 4
├── src/                             ← Development starts after gate passes
└── ...
```

---

## Repository Structure

```
research-pipeline/
├── README.md              ← You are here
├── GENERATOR.md           ← THE TOOL: generator prompt (start here)
├── FRAMEWORK.md           ← Complete v3.0 methodology specification
├── AGENTS.md              ← AI agent operating manual
│
├── templates/             ← Operational templates (used during research)
│   ├── DECISIONS.template.md
│   ├── CONFLICT-RESOLUTION.template.md
│   ├── FOUNDING-ARCHITECTURE.template.md
│   └── PHASE-0-GATE.template.md
│
└── meta-research/         ← Empirical evidence base (provenance, not operational)
    ├── README.md
    ├── DECISIONS.md        ← 10 hypothesis verdicts, 31 evidence nodes
    └── research/           ← 14 primary research artifacts (639KB)
```

**4 operational files** (GENERATOR + 3 docs) + **4 templates** + FRAMEWORK for reference. That's the whole system.

---

## Key Capabilities

| Capability | How It Works |
|---|---|
| **Adaptive Scaling** | 8-dimension complexity scoring (0–24) maps to 4 tiers (1–30 sessions) |
| **Open-Ended Input** | Accepts natural language vision dumps; AI extracts parameters and classifies |
| **Constrained DAG** | Sessions run when dependencies are met, not rigid stage gates |
| **5-Block Prompts** | BRIEF, SCOPE, APPROACH, DELIVERABLE, FORMAT — no personas, no hardcoded queries |
| **Staged Triangulation** | Single-model default → multi-model only for contested one-way doors |
| **GRADE-Aligned Evidence** | A–E grades + modifiers (corroboration, recency, directness) + verification |
| **Two-Track Gate** | Fast-track for reversible decisions; 9-step + premortem for irreversible |

---

## Origin & Philosophy

### Why This Exists

Software venture failures almost never stem from bad code — they stem from **premature architectural decisions made on unverified assumptions.** The cost of wrong decisions compounds: a bad database choice costs 10× more to fix at month 6 than at month 0.

URP was born from 7 real project research pipelines (2024–2026), each independently discovering pieces of the same methodology: unbiased landscape cataloging, evidence grading, aspect isolation, decision registries, and synthesis protocols. URP v1.0/v2.0 consolidated these patterns.

### The v3.0 Transformation

In August 2026, URP was subjected to its own methodology. 11 independent deep research sessions tested every foundational assumption. Of 10 hypotheses, **0 survived unchanged**:

| v2.0 Dogma | Verdict | v3.0 Resolution |
|---|---|---|
| Absolute aspect-isolation | Refined | Context Architecture Law — conditional decomposition + synthesis |
| Mandatory 3-model triangulation | Refined | Staged, risk-triggered protocol |
| Fixed 17–27 sessions | Refuted | 4-tier adaptive scaling (1–30) |
| 8-section XML prompts | Refuted | 5-block prompt anatomy |
| Expert personas improve research | Refuted | Personas debunked for factual accuracy |
| Fixed ~40/60 compose/build | Refined | Wardley evolution mapping |

The most load-bearing finding: **how you frame a research question measurably biases what "evidence" a model reports back.** This single finding justifies the entire framework — unstructured "just ask the AI" research is systematically vulnerable to confirmation bias.

### The Self-Referential Validation

URP v3.0's most distinctive property: it was validated by the methodology it prescribes. The meta-research discovered that several of its own axioms needed refinement. This recursive self-correction is built into v3.0's DNA through mandatory review triggers and decay conditions on every locked ADR.

---

## Roadmap

| Phase | Status | Description |
|---|---|---|
| 1: Foundation | ✅ | 7 project methodologies consolidated into v1.0/v2.0 |
| 2: Self-Validation | ✅ | 11 meta-research sessions → 10 verdicts → v3.0 spec |
| 3: v3.0 Overhaul | ✅ | Framework, generator, templates rewritten from evidence |
| 4: Code-Based Generator | Next | 5-layer deterministic/AI hybrid CLI tool |
| 5: Operational Hardening | Future | CI validation, blast-radius tracking, review scheduling |
| 6: v4.0 Frontiers | Future | DSPy optimization, multi-agent debate, longitudinal calibration |
