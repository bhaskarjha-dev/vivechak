# Vivechak (विवेचक) — Agent Operating Manual
### Vivechak v1.1 · Single Source of Truth for AI Agent Operations

> **Read this file COMPLETELY before doing anything in this repository.**

---

## 1. What This Repository Is

**Vivechak** is a meta-framework that generates evidence-grounded research pipelines for software projects. It transforms a raw project idea into a sealed Founding Architecture Document (FAD) before coding begins.

**The tool is [GENERATOR.md](GENERATOR.md).** Everything else supports it.

---

## 2. Using Vivechak for a New Project

This is the common case — an agent is asked to generate and/or execute a research pipeline for a project.

### Generate the Pipeline

1. Open [GENERATOR.md](GENERATOR.md) and extract the prompt from inside the ```` ````markdown ```` code fences.
2. Replace `[PASTE YOUR PROJECT DESCRIPTION HERE]` with the project vision.
3. Execute the prompt in a fresh AI session with web search enabled.
4. You will receive **two files**: `RESEARCH-PIPELINE.md` (session DAG + prompts) and `DECISIONS.md` (initial decision registry).

### Set Up the Project Workspace

5. Save both generated files to the project's `research/` directory.
6. Copy the **4 templates** from `templates/` into `research/templates/`:
   - `DECISIONS.template.md` — for recording architectural decisions
   - `CONFLICT-RESOLUTION.template.md` — for resolving conflicting findings
   - `FOUNDING-ARCHITECTURE.template.md` — for compiling the final FAD
   - `PHASE-0-GATE.template.md` — for the pre-coding exit gate

```
project/
└── research/
    ├── RESEARCH-PIPELINE.md         ← Generated
    ├── DECISIONS.md                 ← Generated
    ├── sessions/                    ← Empty folder for session outputs
    └── templates/                   ← Copied from Vivechak
        ├── DECISIONS.template.md
        ├── CONFLICT-RESOLUTION.template.md
        ├── FOUNDING-ARCHITECTURE.template.md
        └── PHASE-0-GATE.template.md
```

**The project workspace is now 100% self-contained.** No further reference to this repository is needed.

### Execute the Pipeline

7. Take each prompt from the generated `RESEARCH-PIPELINE.md` and run it in an independent AI research session. Save outputs to `research/sessions/`.
8. After key sessions, record decisions in `DECISIONS.md` using the `DECISIONS.template.md` format. Classify each as one-way or two-way door.
9. If sessions produce conflicting recommendations, resolve using `CONFLICT-RESOLUTION.template.md`.
10. Synthesize all findings into `FAD.md` using `FOUNDING-ARCHITECTURE.template.md`.
11. Verify exit criteria with `PHASE-0-GATE.template.md`. Once the gate passes — start coding.

> **You do NOT need to read FRAMEWORK.md to use Vivechak.** GENERATOR.md is self-contained — all methodology is operationalized inline in the generator prompt and baked into every generated session prompt.

---

## 3. Developing Vivechak Itself

This section applies only when improving the meta-framework — editing FRAMEWORK.md, GENERATOR.md, or templates/.

1. **Read [FRAMEWORK.md](FRAMEWORK.md)** — the complete methodology specification (8 principles, evidence grading, pipeline topology, prompt anatomy).
2. **Read [CONTRIBUTING.md](CONTRIBUTING.md)** — all changes must be traceable to evidence.
3. **⚠️ Run the Change Propagation Map** (CONTRIBUTING.md § Change Propagation Map) before committing. FRAMEWORK.md is the specification; GENERATOR.md is the implementation. Changing the spec without updating the implementation means improvements exist in documentation but never affect generated pipelines. Every change to FRAMEWORK.md, GENERATOR.md, or templates/ MUST be checked against the propagation map.
4. All changes must follow the 8 Core Principles (summarized below).
5. The sealed `meta-research/` directory contains the empirical evidence base — 15 research artifacts validating every design decision.

### The 8 Core Principles

| # | Principle | Summary |
|---|---|---|
| P1 | Context Architecture Law | Decompose by attention budget and coupling, not arbitrary counts. Mandate synthesis after decomposition. |
| P2 | Reversibility-Calibrated Rigor | Deep research for One-Way Doors, fast spikes for Two-Way Doors. |
| P3 | Evidentiary Grounding | Every claim carries evidence grade (A–E), modifiers, and verification method. Recalled = Grade D cap. |
| P4 | Prescriptive Scope, Dynamic Method | Prescriptive on WHAT/WHY/BOUNDARIES. Directional on HOW. No hardcoded queries or personas. |
| P5 | Commodity-Maximized Composition | Compose 100% commodity infra. Build 100% custom only for proprietary domain logic. |
| P6 | Dual-Audience Artifacts | Hybrid Markdown + YAML frontmatter for human reading and machine synthesis. |
| P7 | Staged Triangulation | Single-model default → critique probe → full triangulation only for contested One-Way Doors. |
| P8 | Structured Falsification | Prioritize disconfirming evidence, document rejected alternatives, schedule review triggers, premortems. |

---

## 4. Repository Structure

```
vivechak/
├── README.md                       ← Overview + 5-step quickstart
├── GENERATOR.md                    ← THE TOOL — generator prompt (self-contained)
├── FRAMEWORK.md                    ← Complete methodology spec (for development only)
├── AGENTS.md                       ← This file
├── ROADMAP.md                      ← Project history + future plans
├── CHANGELOG.md                    ← Release history
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

## 5. Key Operational Rules

- **No legacy dogma:** Do not use 8-section XML prompts, expert personas, hardcoded search queries, minimum search counts, or rigid output skeletons. These are empirically refuted.
- **Front-load briefs:** Never drip-feed instructions across turns (39% performance drop documented).
- **Grade everything:** Every factual claim needs an inline evidence grade with modifiers and verification method.
- **Two-Way Doors move fast:** Don't over-research reversible decisions. Spike or decide by convention.
- **One-Way Doors move carefully:** Require corroborated Grade A/B evidence, locked ADR, and premortem before commitment.
