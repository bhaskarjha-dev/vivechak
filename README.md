# Universal Research Pipeline (URP) v3.0
### Evidence-Grounded Pre-Development Research Meta-Framework

> **Transform software architecture decisions from gut-feel and cached training data into structured, evidence-graded, empirically validated research — before a single line of application code is written.**

---

## What Is This?

The Universal Research Pipeline is a **production-grade meta-framework** for conducting systematic pre-development research on any software project. It generates customized research pipelines that investigate every critical architectural decision with the same rigor used in clinical medicine, intelligence analysis, and decision science.

**URP v3.0 is the first version validated by its own methodology.** 11 independent research sessions across frontier AI platforms tested every v2.0 assumption. Of 10 founding hypotheses, 0 survived unchanged — 3 were refuted, 4 refined, and 3 validated with structural enhancements. The result is an adaptive, risk-calibrated system grounded in 31 evidence nodes from peer-reviewed studies and industry standards.

### Key Capabilities

| Capability | Description |
|---|---|
| **Adaptive Scaling** | 8-dimension complexity scoring (0–24) → 4 tiers (1–30 sessions) instead of fixed 17–27 |
| **Constrained DAG** | Dependency-gated execution graph replaces rigid sequential staging |
| **5-Block Prompts** | BRIEF, SCOPE, APPROACH, DELIVERABLE, FORMAT — no personas, no hardcoded queries |
| **Staged Triangulation** | Single-model default → critique probe → full multi-model only for contested One-Way Doors |
| **GRADE-Aligned Evidence** | A–E grades + 3 modifiers (corroboration, recency, directness) + verification tracking |
| **Two-Track Gate** | Fast-track for reversible decisions; rigorous 9-step + premortem for irreversible ones |
| **5-Layer Generator** | Deterministic scaffold + scoped AI calls + validation gates |

---

## Quick Start

1. **Score your project:** Use [templates/COMPLEXITY-SCORING.template.md](templates/COMPLEXITY-SCORING.template.md) to assess complexity and determine your research tier.

2. **Generate your pipeline:** Use [META-PROMPT-GENERATOR.md](META-PROMPT-GENERATOR.md) to generate a customized RESEARCH-PIPELINE.md and PROMPT-LIBRARY.md for your project.

3. **Execute sessions:** Run each research prompt in a frontier AI deep research mode (Claude, Gemini, ChatGPT with search). Sessions default to independent and parallelizable.

4. **Record decisions:** Document each architectural verdict using [templates/DECISIONS.template.md](templates/DECISIONS.template.md) with evidence traceability.

5. **Synthesize:** Compile the Founding Architecture Document using [templates/FOUNDING-ARCHITECTURE.template.md](templates/FOUNDING-ARCHITECTURE.template.md).

6. **Gate:** Verify all exit criteria via [templates/PHASE-0-GATE.template.md](templates/PHASE-0-GATE.template.md) before writing code.

---

## Repository Structure

```
research-pipeline/
├── FRAMEWORK.md                 ← Complete v3.0 specification (start here)
├── PRINCIPLES.md                ← 8 evidence-grounded principles
├── EVIDENCE-GRADING.md          ← A–E grading spec with modifiers
├── META-PROMPT-GENERATOR.md     ← 5-layer generator + interim prompt
├── AGENTS.md                    ← Agent operating manual
├── CONTEXT.md                   ← Origin story & design insights
├── VISION.md                    ← Mission & philosophy
├── ROADMAP.md                   ← Evolution roadmap
│
├── schemas/                     ← JSON Schema validation
│   ├── session-frontmatter.schema.json
│   ├── adr-frontmatter.schema.json
│   └── evidence-record.schema.json
│
├── templates/                   ← Reusable v3.0 blueprints
│   ├── RESEARCH-PIPELINE.template.md
│   ├── PROMPT-LIBRARY.template.md
│   ├── DECISIONS.template.md
│   ├── EVIDENCE-RECORD.template.md
│   ├── COMPLEXITY-SCORING.template.md
│   ├── CONFLICT-RESOLUTION.template.md
│   ├── FOUNDING-ARCHITECTURE.template.md
│   └── PHASE-0-GATE.template.md
│
└── meta-research/               ← Empirical evidence base (v3.0 provenance)
    ├── DECISIONS.md              ← Sealed ADR corpus (10 verdicts, 31 evidence nodes)
    ├── RESEARCH-PIPELINE.md      ← The meta-research pipeline definition
    ├── PROMPT-LIBRARY.md         ← Prompts used for meta-research
    └── research/                 ← 14 primary research artifacts (639KB)
```

---

## Core Documents

| Document | Purpose | Read When |
|---|---|---|
| [FRAMEWORK.md](FRAMEWORK.md) | Complete v3.0 specification | First — understand the system |
| [PRINCIPLES.md](PRINCIPLES.md) | 8 evidence-grounded axioms | Understanding design philosophy |
| [EVIDENCE-GRADING.md](EVIDENCE-GRADING.md) | A–E grading with modifiers | Evaluating or grading claims |
| [META-PROMPT-GENERATOR.md](META-PROMPT-GENERATOR.md) | Generator architecture + prompt | Starting a new project pipeline |
| [AGENTS.md](AGENTS.md) | Operating manual for AI agents | Working on this repository |
| [CONTEXT.md](CONTEXT.md) | Origin story & tacit knowledge | Understanding why decisions were made |
| [VISION.md](VISION.md) | Mission & long-term philosophy | Understanding the bigger picture |
| [ROADMAP.md](ROADMAP.md) | Evolution plan | Understanding what's next |

---

## Empirical Provenance

URP v3.0 was produced by applying the framework to itself:

- **11 research sessions** (T1-01 through T3-01) across Claude, ChatGPT, and Gemini
- **31 evidence nodes** from peer-reviewed studies, industry standards, and empirical benchmarks
- **10 hypothesis verdicts** — every v2.0 assumption empirically tested
- **639KB** of primary research artifacts preserved in [meta-research/](meta-research/)

The meta-research directory is the empirical audit trail. The framework is fully self-contained without it.
