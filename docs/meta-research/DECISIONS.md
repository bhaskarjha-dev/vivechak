# Vivechak — Architectural Decision Registry
### Decisions D-001 through D-014 · Status: All Locked
**Schema Version:** 1.1 · **Last Updated:** 2026-09-28

> This file documents every locked architectural decision for Vivechak.
> Decisions D-001–D-010 were made during v1.0 (methodology design, August 2026).
> Decisions D-011–D-014 were made during v1.1 (MCP server + multi-scope, September 2026).
> Raw research sessions are preserved in git history under `meta-research/sessions/`.

---

## Decision Index

| ID | Decision | Door | Status |
|---|---|---|---|
| D-001 | Context Architecture Law (P1) | One-Way | Accepted |
| D-002 | Reversibility-Calibrated Rigor (P2) | One-Way | Accepted |
| D-003 | Constrained DAG Pipeline Topology | Two-Way | Accepted |
| D-004 | Risk-Calibrated Scaling: Complexity Score, Not Tiers | One-Way | Accepted |
| D-005 | GRADE-Aligned Evidence Standard (A–E) | One-Way | Accepted |
| D-006 | Natural-Language Prompt Anatomy (No XML/Personas) | Two-Way | Accepted |
| D-007 | Hybrid Output Architecture (Markdown + YAML) | Two-Way | Accepted |
| D-008 | Commodity-Maximized Composition (P5) | One-Way | Accepted |
| D-009 | Two-Track Phase 0 Exit Gate | One-Way | Accepted |
| D-010 | Single Generator with Inline Method (No 5-Layer Stack) | One-Way | Accepted |
| D-011 | Guided Worker Pattern for MCP Server | One-Way | Accepted |
| D-012 | Three Scope Levels: Project / Decision / Comparison | Two-Way | Accepted |
| D-013 | 9 MCP Tools with `vivechak_` Prefix | One-Way | Accepted |
| D-014 | Binary-First Distribution | Two-Way | Accepted |

---

## Methodology Decisions (D-001 – D-010)

---

```yaml
---
id: D-001
title: "Context Architecture Law (P1)"
status: accepted
door_type: one-way
date: 2026-08-15
confidence: high
informed_by_sessions: [T1-02, T1-03, T2-01]
review_trigger: "Context windows exceed 2M tokens across all frontier models"
schema_version: "1.1"
---
```

# D-001: Context Architecture Law (P1)

## Context & Problem Statement

LLM performance degrades with input size. Research needs to be decomposed into focused sessions, but arbitrary decomposition wastes token budget and creates artificial coupling. The framework needs a principle for when and how to decompose.

## Evaluated Options

1. **Fixed session count per tier** — Rigid. Doesn't adapt to project complexity.
2. **Aspect isolation (one topic per session)** — Original v2.0 hypothesis. Too prescriptive.
3. **Context Architecture Law** — Decompose by attention budget and coupling. Mandate synthesis after decomposition. Prescriptive on WHAT to decompose, directional on HOW.

## Decision Outcome

**Chosen Option:** Context Architecture Law — decompose when a topic exceeds a single session's effective attention budget, always synthesize after decomposition.

### Rationale

Chroma Research 2025 demonstrated context rot across 18 frontier LLMs. MTI benchmarks showed +12.4% accuracy on coupled tasks when kept together. The principle balances decomposition benefits against synthesis costs. Refined from "Aspect Isolation" to "Context Architecture" to capture the coupling dimension, not just topic separation.

## Rejected Alternatives & Tradeoffs

- **Fixed session counts:** Rejected — projects vary too much in complexity.
- **Pure aspect isolation:** Rejected — ignores coupling between aspects.

## Failure Modes & Reversal Triggers

- If context windows exceed 2M tokens universally, the decomposition threshold shifts. Re-evaluate whether multi-session is still necessary for typical projects.

---

```yaml
---
id: D-002
title: "Reversibility-Calibrated Rigor (P2)"
status: accepted
door_type: one-way
date: 2026-08-15
confidence: high
informed_by_sessions: [T1-01, T2-04]
review_trigger: "Reversal cost drops below research cost for One-Way Doors"
schema_version: "1.1"
---
```

# D-002: Reversibility-Calibrated Rigor (P2)

## Context & Problem Statement

Not all decisions deserve equal research depth. Over-researching reversible choices wastes time; under-researching irreversible ones creates costly mistakes.

## Evaluated Options

1. **Uniform rigor for all decisions** — Simple but wasteful.
2. **Amazon Two-Way Door classification** — Binary: reversible vs irreversible.
3. **Reversibility-calibrated rigor** — Deep research for One-Way Doors, fast spikes for Two-Way Doors. Matches research investment to reversal cost.

## Decision Outcome

**Chosen Option:** Reversibility-calibrated rigor. One-Way Doors require corroborated Grade A/B evidence, locked ADR, and premortem. Two-Way Doors are decided by convention or single-session spike.

### Rationale

Amazon's Type 1/Type 2 framework (Bezos 2015) validated by Cynefin decision contexts (Snowden 1999). DORA data shows architectural failure rates concentrate in irreversible decisions. Research investment should correlate with reversal cost.

## Rejected Alternatives & Tradeoffs

- **Uniform rigor:** Rejected — most decisions are Two-Way Doors. Forcing deep research on all of them delays delivery without improving outcomes.

---

```yaml
---
id: D-003
title: "Constrained DAG Pipeline Topology"
status: accepted
door_type: two-way
date: 2026-08-15
confidence: high
informed_by_sessions: [T1-01, T2-03]
review_trigger: "Evidence that linear pipelines outperform DAGs for typical projects"
schema_version: "1.1"
---
```

# D-003: Constrained DAG Pipeline Topology

## Context & Problem Statement

Research sessions have dependencies. Some sessions can run in parallel; others must wait for upstream findings. The pipeline needs a topology that expresses these relationships.

## Decision Outcome

**Chosen Option:** Constrained DAG — sessions have explicit dependencies expressed as a directed acyclic graph. 67% of Tier 2 sessions have ≤1 Tier 1 dependency, enabling significant parallelism. Synthesis (SYN-01) is mandatory and always terminal.

### Rationale

PMBOK mandatory vs discretionary dependency classification validates the approach. Linear pipelines force unnecessary serialization. Full DAGs without constraints create unmanageable complexity. The constrained DAG (Tier 1 parallel → Tier 2 dependent → SYN terminal) balances expressiveness with simplicity.

---

```yaml
---
id: D-004
title: "Risk-Calibrated Scaling: Complexity Score, Not Rigid Tiers"
status: accepted
door_type: one-way
date: 2026-08-15
confidence: high
informed_by_sessions: [T1-01, T2-04]
review_trigger: "Calibration data from 20+ real pipelines available"
schema_version: "1.1"
---
```

# D-004: Risk-Calibrated Scaling: Complexity Score, Not Rigid Tiers

## Context & Problem Statement

The original v2.0 hypothesis proposed a rigid 4-tier scaling model (Tier 1: 4-8 sessions, Tier 2: 9-15, etc.). Meta-research refuted this as too rigid.

## Decision Outcome

**Chosen Option:** Continuous complexity score (6 dimensions, 0-4 each, max 24) mapped to session count ranges. The score is a tool for the generator prompt, not a rigid gate. Generator has latitude to adjust session count within ±2 of the scored range.

### Rationale

The original 4-tier model had a sharp cliff at the 15→16 boundary that couldn't be justified empirically. A continuous score with soft boundaries is more honest about uncertainty. Calibration data from real pipelines will refine the scoring over time.

---

```yaml
---
id: D-005
title: "GRADE-Aligned Evidence Standard (A–E)"
status: accepted
door_type: one-way
date: 2026-08-15
confidence: high
informed_by_sessions: [T1-01, T1-02, T2-05]
review_trigger: "Evidence that the grading system creates false precision"
schema_version: "1.1"
---
```

# D-005: GRADE-Aligned Evidence Standard (A–E)

## Context & Problem Statement

AI-generated research mixes facts and confabulation without distinction. Every claim needs a quality signal.

## Decision Outcome

**Chosen Option:** 5-level evidence grading (A–E) inspired by GRADE (Guyatt et al.) and critical review of the Admiralty Code (87% diagonal collapse). Every factual claim carries an inline grade, modifiers (corroborated/single, fresh/aging, direct/indirect), and verification method. Recalled/confabulated claims are capped at Grade D.

### Rationale

The Admiralty Code's 6×6 matrix collapses to the diagonal 87% of the time — a simpler linear scale is more honest. GRADE's certainty levels map cleanly to the AI research context. The modifier system captures dimensions that a single letter cannot.

---

```yaml
---
id: D-006
title: "Natural-Language Prompt Anatomy (No XML/Personas)"
status: accepted
door_type: two-way
date: 2026-08-15
confidence: very-high
informed_by_sessions: [T1-03, T2-06]
review_trigger: "Empirical evidence that structured XML outperforms natural language"
schema_version: "1.1"
---
```

# D-006: Natural-Language Prompt Anatomy (No XML/Personas)

## Context & Problem Statement

The original v2.0 hypothesis proposed an 8-section XML prompt template with expert personas and hardcoded search queries. Meta-research refuted all three elements.

## Decision Outcome

**Chosen Option:** Natural-language brief with 5 flexible blocks (Brief, Scope, Constraints, Deliverable, Context). No XML structure, no expert personas, no hardcoded queries. Dynamic method selection by the AI, not prescribed search patterns.

### Rationale

- **Personas:** Zheng et al. (EMNLP 2024) showed persona accuracy degradation. Expert personas don't improve output quality and can degrade it.
- **XML structure:** Tam et al. (EMNLP 2024) demonstrated format restrictions impair reasoning.
- **Hardcoded queries:** ReAct (Yao et al. 2023) interleaved reasoning outperforms static query lists.
- **Front-loaded briefs:** Laban et al. (ICLR 2026, Outstanding Paper) documented 39% performance drop from drip-feeding instructions across turns.

---

```yaml
---
id: D-007
title: "Hybrid Output Architecture (Markdown + YAML Frontmatter)"
status: accepted
door_type: two-way
date: 2026-08-15
confidence: high
informed_by_sessions: [T1-01, T2-07]
review_trigger: "Git line-level diffing no longer relevant for artifact management"
schema_version: "1.1"
---
```

# D-007: Hybrid Output Architecture (Markdown + YAML Frontmatter)

## Context & Problem Statement

Research outputs need to serve two audiences: human readers (prose) and machine synthesis (structured data). Pure JSON is unreadable; pure Markdown is unparseable.

## Decision Outcome

**Chosen Option:** Markdown body with YAML frontmatter. Human-readable prose in Markdown. Machine-parseable metadata (session_id, status, evidence_refs, door_type) in YAML frontmatter. Git line-level diffs work on both.

---

```yaml
---
id: D-008
title: "Commodity-Maximized Composition (P5)"
status: accepted
door_type: one-way
date: 2026-08-15
confidence: high
informed_by_sessions: [T1-01, T2-04, SYN-01]
review_trigger: "Proprietary infrastructure creates measurable competitive advantage"
schema_version: "1.1"
---
```

# D-008: Commodity-Maximized Composition (P5)

## Decision Outcome

**Chosen Option:** Compose 100% commodity infrastructure. Build 100% custom only for proprietary domain logic (methodology, evidence grading, validation). The framework's value is the methodology and rigor, not the plumbing. Wardley evolution mapping confirms infrastructure components are commodity-phase.

---

```yaml
---
id: D-009
title: "Two-Track Phase 0 Exit Gate"
status: accepted
door_type: one-way
date: 2026-08-15
confidence: high
informed_by_sessions: [T1-01, T2-05, SYN-01]
review_trigger: "Gate false-positive rate exceeds 20%"
schema_version: "1.1"
---
```

# D-009: Two-Track Phase 0 Exit Gate

## Decision Outcome

**Chosen Option:** Two-track gate: Track A checks structural completeness (all sessions complete, FAD exists, decisions locked). Track B checks quality indicators (evidence grades present, premortem substance). Gate performs MECHANICAL checks only — semantic quality is the human/agent's responsibility. Inspired by WHO Surgical Safety Checklist (Gawande et al., 47% mortality reduction from simple checklists).

---

```yaml
---
id: D-010
title: "Single Generator with Inline Method (No 5-Layer Stack)"
status: accepted
door_type: one-way
date: 2026-08-15
confidence: high
informed_by_sessions: [T1-03, T3-01]
review_trigger: "Generator prompt exceeds 30K tokens"
schema_version: "1.1"
---
```

# D-010: Single Generator with Inline Method (No 5-Layer Stack)

## Context & Problem Statement

The original hypothesis proposed a 5-layer generator architecture (system prompt → method → scaling → session factory → output). Meta-research refuted this as over-engineered.

## Decision Outcome

**Chosen Option:** Single self-contained generator prompt with methodology inlined. All methodology is operationalized directly in the generator prompt and baked into every generated session prompt. No layered architecture, no method injection, no session factory. GENERATOR.md is the implementation; FRAMEWORK.md is the specification.

---

## MCP Server Decisions (D-011 – D-014)

---

```yaml
---
id: D-011
title: "Guided Worker Pattern for MCP Server"
status: accepted
door_type: one-way
date: 2026-09-25
confidence: high
informed_by_sessions: [R-02, R-04]
review_trigger: "MCP spec adds stateful session management"
schema_version: "1.1"
---
```

# D-011: Guided Worker Pattern for MCP Server

## Context & Problem Statement

The MCP server needs to guide the host agent through the research workflow without maintaining server-side state. The MCP 2026-07-28 spec is stateless — sessions were removed.

## Evaluated Options

1. **State Machine on server** — Server tracks workflow stage, rejects out-of-order calls.
2. **Guided Worker** — Every response includes `next_step` telling the agent what to do next. Hard refusals ONLY for physically impossible operations, never for "wrong order."
3. **Unguided (raw API)** — Agent must discover the workflow on its own.

## Decision Outcome

**Chosen Option:** Guided Worker pattern. Every tool response includes `next_step`. The server never rejects calls because they're "out of order" — it simply tells the agent what would be most productive next. This respects MCP's stateless design and P8 ("not rigid stage gates").

### Rationale

MCP is stateless by spec. A state machine creates a shadow state that contradicts the protocol's design. The Guided Worker pattern provides the same UX (agent knows what to do next) without server-side state.

## Rejected Alternatives & Tradeoffs

- **State Machine:** Rejected — creates shadow state, contradicts MCP spec. Beads project's JSONL+SQLite had integrity/concurrency bugs from shadow state.
- **Unguided:** Rejected — agents consistently fail to discover multi-step workflows without guidance.

---

```yaml
---
id: D-012
title: "Three Scope Levels: Project / Decision / Comparison"
status: accepted
door_type: two-way
date: 2026-09-25
confidence: high
informed_by_sessions: [R-03, R-06]
review_trigger: "Demand signal for additional scope levels"
schema_version: "1.1"
---
```

# D-012: Three Scope Levels: Project / Decision / Comparison

## Context & Problem Statement

Not every use of Vivechak is a full project pipeline. Developers need to make single decisions or compare options without the overhead of a multi-session pipeline.

## Decision Outcome

**Chosen Option:** Three scope levels with independent generators and templates:
- **Project** (GENERATOR.md → full pipeline → FAD) — multi-session DAG
- **Decision** (GENERATOR-DECISION.md → 1-3 sessions → ADR) — single decision
- **Comparison** (GENERATOR-COMPARISON.md → 1 session → WEP matrix) — bounded comparison

Scope × Depth are independent dimensions. Each scope has its own generator but shares the same evidence grading, frontmatter format, and validation ladder.

### Rationale

Cross-domain validation: Cochrane systematic reviews use full/scoping/rapid review hierarchy. ICD 203 has multi-level assessment. PRISMA has full/ScR variants. The pattern recurs across every evidence-based discipline.

---

```yaml
---
id: D-013
title: "9 MCP Tools with vivechak_ Prefix"
status: accepted
door_type: one-way
date: 2026-09-27
confidence: high
informed_by_sessions: [R-02, R-04]
review_trigger: "MCP spec adds tool namespacing that conflicts with prefix convention"
schema_version: "1.1"
---
```

# D-013: 9 MCP Tools with `vivechak_` Prefix

## Context & Problem Statement

The tool surface defines what agents can do. Too many tools increases cognitive load and discovery cost. Too few makes the server useless. The original plan had 10 tools.

## Decision Outcome

**Chosen Option:** 9 tools after code-level audit (2026-09-27) dropped `vivechak_synthesize`:

| Tool | Purpose | Read-Only |
|---|---|---|
| `vivechak_init` | Create workspace | No |
| `vivechak_prepare_generator` | Get generator prompt with context | Yes |
| `vivechak_save_plan` | Persist pipeline/decisions files | No |
| `vivechak_status` | Scan workspace progress | Yes |
| `vivechak_next_session` | Get next actionable session with context injection | Yes |
| `vivechak_save_session` | Validate and persist session output | No |
| `vivechak_record_decision` | Persist ADR or conflict resolution | No |
| `vivechak_validate` | Dry-run validation (no side effects) | Yes |
| `vivechak_run_gate` | Phase 0 exit gate (mechanical checks) | Yes |

### Rationale

`vivechak_synthesize` was dropped because it can't synthesize without an LLM, and the server must not call LLM APIs. SYN-01 is handled by `vivechak_next_session` (which returns the synthesis prompt with all findings injected) + `vivechak_save_session` (which persists the result).

`get_generator_prompt` → `prepare_generator` and `save_pipeline` → `save_plan` were renamed for clarity.

---

```yaml
---
id: D-014
title: "Binary-First Distribution"
status: accepted
door_type: two-way
date: 2026-09-25
confidence: high
informed_by_sessions: [R-05]
review_trigger: "Agent Plugins become universal across all major hosts"
schema_version: "1.1"
---
```

# D-014: Binary-First Distribution

## Context & Problem Statement

The MCP server needs to reach developers across 7+ host environments (Cursor, VS Code, Claude Desktop, Antigravity, ChatGPT, Codex, Kiro). The primary distribution question is binary vs npm vs Agent Plugin.

## Evaluated Options

1. **npm package** — Easy install but requires Node.js. Adds a runtime dependency.
2. **Agent Plugin** — One-click install in some hosts. But Claude Desktop and Antigravity don't support it.
3. **Binary-first** — Single compiled binary. No runtime dependencies. Works everywhere. Absolute path in MCP config avoids PATH truncation.

## Decision Outcome

**Chosen Option:** Binary-first distribution via GoReleaser (6 platforms), shell installers, Homebrew, Scoop, winget. `vivechak mcp-config --client <host> --write` generates correct absolute-path config for each host.

### Rationale

macOS GUI hosts truncate PATH to `/usr/bin:/bin:/usr/sbin:/sbin`. npm requires Node.js installation. Agent Plugins don't work on Claude Desktop or Antigravity. A single binary with absolute-path config is the only approach that works everywhere.

## Rejected Alternatives & Tradeoffs

- **npm:** Rejected — adds Node.js dependency, doesn't solve PATH truncation.
- **Agent Plugin:** Not rejected but demoted to supplementary channel. Can't be primary when 2 major hosts don't support it.
