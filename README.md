<div align="center">

<img src="docs/assets/logo.jpg" alt="Vivechak" width="180" />

# Vivechak (विवेचक)

**Evidence-grounded pre-development research meta-framework.**

*vi- (apart) + √vic (to sift/separate) + -aka (the agent who does) = "the discerning analyst"*

</div>

---

### v1.1 — Separating evidence from assumption, truth from bias

> **Transform software architecture decisions from gut-feel and cached training data into structured, evidence-graded, empirically validated research — before a single line of application code is written.**

---

## What Is This?

You have a software project idea. Before coding, you need to make architectural decisions — database, auth, hosting, data model, APIs. Some of these decisions are **irreversible** (one-way doors). If you get them wrong, you're looking at a rewrite.

**Vivechak generates a customized research pipeline for your project** that investigates every critical decision with appropriate rigor. It produces focused research prompts, executes them across frontier AI platforms, grades every claim with traceable evidence, and synthesizes the results into a sealed Founding Architecture Document (FAD) — your architectural source of truth before writing code.

**Vivechak is not a deep-research tool** — it is the **orchestration layer above** tools like Gemini Deep Research, ChatGPT Deep Research, and Perplexity. Those tools execute individual research sessions; Vivechak structures which sessions to run, how evidence is graded consistently across sessions, how decisions are tracked with reversal triggers, and how findings are synthesized into a coherent architecture.

---

## Quick Start (5 Steps)

### Step 1: Generate Your Pipeline
Open [GENERATOR.md](GENERATOR.md), paste your project description into the generator prompt, and send it to a frontier AI (Claude, Gemini, or ChatGPT with web search). You'll receive two files:
- **RESEARCH-PIPELINE.md** — your project's complexity score, session DAG, and copy-paste-ready research prompts for every session
- **DECISIONS.md** — initial decision registry with proposed hypotheses (evolves as you execute research)

### Step 2: Set Up Your Project Workspace
Create your project's `research/` directory, paste in the generated files, and **copy the 4 templates** from this repository:

```
my-project/
└── research/
    ├── RESEARCH-PIPELINE.md         ← Generated (paste here)
    ├── DECISIONS.md                 ← Generated (paste here)
    ├── sessions/                    ← Create empty folder for research outputs
    └── templates/                   ← Copy from Vivechak (see templates/)
        ├── DECISIONS.template.md
        ├── CONFLICT-RESOLUTION.template.md
        ├── FOUNDING-ARCHITECTURE.template.md
        └── PHASE-0-GATE.template.md
```

Your project is now **100% self-contained**. You never need to return to this meta-repo.

### Step 3: Execute Research Sessions
Take each prompt from your generated RESEARCH-PIPELINE.md (inside the ```` ```prompt ```` code fences) and run it in an independent AI deep research session. Layer 0 sessions are designed to be parallel — run as many simultaneously as you want.

Save each output using the filename specified in the session's metadata table (e.g., `research/sessions/T1-01-primary-datastore-selection.md`).

### Step 4: Record Decisions
After key research sessions, update `DECISIONS.md` — lock architectural decisions using the format from `templates/DECISIONS.template.md`. Classify each as one-way or two-way door. If models disagree, use `templates/CONFLICT-RESOLUTION.template.md`.

### Step 5: Synthesize, Gate & Build
Compile all findings into a Founding Architecture Document using `templates/FOUNDING-ARCHITECTURE.template.md`. Verify exit criteria with `templates/PHASE-0-GATE.template.md`. Once the gate passes — start coding with the FAD as your architectural source of truth.

---

## Repository Structure

> **For users:** You only need **GENERATOR.md** (to generate your pipeline) and **templates/** (to copy into your project). Everything else is either for learning or for developing Vivechak itself.

```
vivechak/
├── README.md                       ← You are here
├── GENERATOR.md                    ← THE TOOL — generator prompt (start here)
├── FRAMEWORK.md                    ← Deep methodology reference (not required for use)
├── AGENTS.md                       ← AI agent operating manual
├── ROADMAP.md                      ← Project history & future plans
├── CHANGELOG.md                    ← Release history & spec evolution
├── CONTRIBUTING.md                 ← Evidence-grounding contribution rules
├── CODE_OF_CONDUCT.md              ← Contributor Covenant v2.1
├── LICENSE                         ← MIT License
├── VERSION                         ← Release version (1.1.0)
│
├── templates/                      ← 4 operational contracts (copy to projects)
│   ├── DECISIONS.template.md       ← YAML frontmatter ADR format
│   ├── CONFLICT-RESOLUTION.template.md ← Analysis of Competing Hypotheses
│   ├── FOUNDING-ARCHITECTURE.template.md ← Map-Reduce synthesis to FAD
│   └── PHASE-0-GATE.template.md    ← Two-track pre-codebase exit gate
│
├── examples/                       ← Concrete adoption walkthroughs
│   └── SAMPLE-PIPELINE.md          ← End-to-end Tier 1 sample project
│
├── docs/assets/                    ← Visual branding & diagrams
│   └── logo.jpg                    ← Minimalist prism logo
│
└── meta-research/                  ← Empirical evidence base (sealed provenance)
    ├── README.md                   ← Provenance index
    ├── DECISIONS.md                ← 10 hypothesis verdicts & 31 evidence nodes
    ├── RESEARCH-PIPELINE.md        ← Meta-research execution DAG
    ├── PROMPT-LIBRARY.md           ← 14 meta-research prompts
    └── research/                   ← 15 primary research artifacts (642KB)
```

---

## Using with AI Agents (Antigravity, Claude Code, Cursor, etc.)

Because your project workspace contains both the generated files AND the operational templates, AI agents can autonomously execute every step. Example commands:

| Step | Agent Prompt |
|---|---|
| **Research** | *"Read T2-01's prompt from `research/RESEARCH-PIPELINE.md`, run deep research with web search, save to `research/sessions/T2-01-datastore-selection.md`"* |
| **Decide** | *"Read `research/sessions/T2-01-*.md`, formulate D-001 in `research/DECISIONS.md` following `research/templates/DECISIONS.template.md`"* |
| **Synthesize** | *"Read all sessions and decisions, compile `FAD.md` following `research/templates/FOUNDING-ARCHITECTURE.template.md`"* |
| **Gate** | *"Audit `FAD.md` against `research/templates/PHASE-0-GATE.template.md`, run premortem for One-Way Doors, emit `PHASE-0-GATE.md`"* |

The templates act as **contracts** — the agent reads them and follows the exact format, methodology, and checklist without hallucinating structure.

---

## Key Capabilities

| Capability | How It Works |
|---|---|
| **Flexible Generation** | Single-session for IDE/API agents, or 2-step split for web chat output limits |
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

Vivechak was born from 7 real project research pipelines (2024–2026), each independently discovering pieces of the same methodology: unbiased landscape cataloging, evidence grading, aspect isolation, decision registries, and synthesis protocols.

### The Empirical Self-Validation

In August 2026, the framework was subjected to its own methodology. 11 independent deep research sessions (producing 15 artifacts across multiple AI platforms) tested every foundational assumption. Of 10 hypotheses, **0 survived unchanged**:

| Legacy Dogma | Verdict | Resolution |
|---|---|---|
| Absolute aspect-isolation | Refined | Context Architecture Law — conditional decomposition + synthesis |
| Mandatory 3-model triangulation | Refined | Staged, risk-triggered protocol |
| Fixed 17–27 sessions | Refuted | 4-tier adaptive scaling (1–30) |
| 8-section XML prompts | Refuted | 5-block prompt anatomy |
| Expert personas improve research | Refuted | Personas debunked for factual accuracy |
| Fixed ~40/60 compose/build | Refined | Wardley evolution mapping |

The most load-bearing finding: **how you frame a research question measurably biases what "evidence" a model reports back.** This single finding justifies the entire framework — unstructured "just ask the AI" research is systematically vulnerable to confirmation bias.

### The Self-Referential Validation

Vivechak's most distinctive property: it was validated by the methodology it prescribes. The meta-research discovered that several of its own axioms needed refinement. This recursive self-correction is built into Vivechak's DNA through mandatory review triggers and decay conditions on every locked ADR.

### FRAMEWORK.md — The Deep Reference

[FRAMEWORK.md](FRAMEWORK.md) contains the complete methodology specification — 8 principles, evidence grading system, pipeline topology, prompt anatomy, and synthesis workflows. **You do not need to read it to use Vivechak.** GENERATOR.md is self-contained; all methodology is operationalized inline in the generator prompt. FRAMEWORK.md is the definitive reference for understanding *why* the methodology works and for anyone improving the framework itself.




## Roadmap

| Phase | Status | Description |
|---|---|---|
| 1: Foundation (Gen 1) | ✅ | 7 project methodologies consolidated into initial pre-development patterns |
| 2: Self-Validation (Gen 2) | ✅ | 11 meta-research sessions → 10 verdicts → Gen 3 spec |
| 3: Framework Release (v1.0) | ✅ | Framework, generator, templates launch as Vivechak v1.0.0 |
| 3b: Bias Correction (v1.1) | ✅ | Bounded Exploration Mandate, Weighted Evaluation Protocol, SaaS de-anchoring |
| 4: Real-World Validation | Next | Battle-test on 2–3 real projects, refine generator from output quality |
| 5: Tooling | Future | CLI initializer, YAML validator, code-based generator (if validated) |
| 6: v2.0 Frontiers | Future | DSPy optimization, multi-agent debate, longitudinal calibration |

See [ROADMAP.md](ROADMAP.md) for detailed plans, overhaul decisions, and source material status.

