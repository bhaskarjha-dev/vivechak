# Vivechak (विवेचक) — Agent Operating Manual
### Vivechak v3.0 · Single Source of Truth for AI Agent Operations

> **Read this file COMPLETELY before modifying anything in this repository.**

---

## 1. What This Repository Is

The **Vivechak (विवेचक)** is a production-grade Meta-Framework for evidence-grounded pre-development research.

**Core Purpose:** Transform software architecture decisions from gut-feel, outdated training data, and hallucinated conclusions into structured, evidence-graded, risk-calibrated research — before a single line of application code is written.

**v3.0 Status:** Empirically validated via 14 meta-research sessions. All v2.0 axioms tested; none survived unchanged. See [meta-research/DECISIONS.md](meta-research/DECISIONS.md) for the complete evidentiary record.

---

## 2. Repository Structure

```
vivechak/
├── README.md                       ← Overview, quickstart, philosophy
├── GENERATOR.md                    ← The generator prompt (THE tool)
├── FRAMEWORK.md                    ← Complete v3.0 specification
├── AGENTS.md                       ← AI agent operating manual
├── ROADMAP.md                      ← Lineage, Phase 4 validation, frontiers
├── CHANGELOG.md                    ← Release history & spec evolution
├── CONTRIBUTING.md                 ← Evidence-grounding contribution rules
├── CODE_OF_CONDUCT.md              ← Contributor Covenant v2.1
├── LICENSE                         ← MIT License
├── VERSION                         ← Spec version (3.0.0)
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
    └── research/                   ← 14 primary research artifacts (639KB)
```

---

## 3. The 8 Core Principles (v3.0)

### P1: The Context Architecture Law
Decompose research by **attention budget and coupling**, not arbitrary session counts. Mandate explicit synthesis after every decomposition.

### P2: Reversibility-Calibrated Rigor
Scale research depth with decision reversibility — **deep for One-Way Doors, fast spikes for Two-Way Doors**.

### P3: Evidentiary Grounding & Verification Provenance
Every claim must carry an **evidence grade (A–E), contextual modifiers, and verification method**. Recalled AI claims capped at Grade D.

### P4: Prescriptive Scope, Dynamic Method
**Prescriptive** on WHAT/WHY/BOUNDARIES. **Directional** on HOW. No hardcoded queries, no personas, no rigid skeletons.

### P5: Commodity-Maximized Composition
**Compose 100%** of commodity infrastructure. **Build 100% custom** only for proprietary domain logic.

### P6: Dual-Audience Artifact Architecture
**Hybrid Markdown + YAML frontmatter** for human reading and machine synthesis.

### P7: Staged Triangulation
**Single-model default** → critique probe → full triangulation only for contested One-Way Doors.

### P8: Structured Falsification
**Prioritize disconfirming evidence**, document rejected alternatives, schedule review triggers, execute premortems.

---

## 4. Operational Workflow

When a session starts in this repo:

1. **Read [FRAMEWORK.md](FRAMEWORK.md)** — understand the complete v3.0 specification.
2. **If improving the meta-framework:** Enhance FRAMEWORK.md, GENERATOR.md, or templates/ following the principles above. All changes must be traceable to evidence.
3. **If generating a pipeline for a new project:** Use [GENERATOR.md](GENERATOR.md) to generate the customized pipeline documents (Single-Session for IDE agents, Split Generation for web chat).
4. **If executing research:** Follow the 5-block prompt anatomy (BRIEF, SCOPE, APPROACH, DELIVERABLE, FORMAT). Front-load everything in one turn.
5. **If recording decisions:** Use the DECISIONS template with YAML frontmatter, door_type classification, evidence_refs, and review_trigger.

---

## 5. Key Operational Rules

- **No v2.0 patterns:** Do not use 8-section XML prompts, expert personas, hardcoded search queries, minimum search counts, or rigid output skeletons. These are empirically refuted.
- **Front-load briefs:** Never drip-feed instructions across turns (39% performance drop documented).
- **Grade everything:** Every factual claim needs an inline evidence grade with modifiers and verification method.
- **Two-Way Doors move fast:** Don't over-research reversible decisions. Spike or decide by convention.
- **One-Way Doors move carefully:** Require corroborated Grade A/B evidence, locked ADR, and premortem before commitment.
