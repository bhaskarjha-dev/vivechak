---
name: vivechak
description: Evidence-grounded research playbook for technical and architectural decisions. Activates when evaluating competing libraries or datastores (WEP matrix), resolving architectural dilemmas (ADRs), or planning system architectures (FAD). Works both with the Vivechak MCP server (vck) and standalone without installing any binary.
license: MIT
compatibility: macOS, Linux, Windows; works with any AI coding agent that has web search (MCP server optional)
metadata:
  version: "0.1.0"
  author: bhaskarjha-dev
  homepage: "https://vivechak.bhaskarjha.dev"
  mcp_server: "vck"
---

# Vivechak (विवेचक) — Autonomous Agent Operating Playbook

**Prove decisions before code.** You are operating as an epistemic analyst using **Vivechak**. Your objective is to eliminate hallucination, confirmation bias, and gut-feel assumptions from technical and architectural decisions — storing verifiable Markdown artifacts in `research/`.

---

## 1. Scope Inference & Workflow Routing

Whenever the user asks for a technical comparison, technology choice, architectural fork, or platform architecture, classify the request into one of **3 scopes**:

| Scope | Trigger | Artifact | Overhead | MCP Prompt / Tool |
|---|---|---|---|---|
| **1. Comparison** | Evaluate 2–5 named options (*"Pino vs Winston"*) | `research/C-NNN-comparison.md` (WEP) | 1 session (~5m), 0 DAG, 0 Gate | `/mcp__vivechak__compare` or `vivechak_start(scope="comparison", question="...")` |
| **2. Decision** | Architectural fork in active codebase (*"Migrate to ClickHouse?"*) | `research/D-NNN-decision.md` (ADR) + `D-NNN-R1.md` | 1–3 sessions (~30m), 0 Project DAG | `/mcp__vivechak__decision` or `vivechak_start(scope="decision", question="...")` |
| **3. Project** | Greenfield architecture or major rewrite | `research/FAD.md` (Founding Architecture) | Full DAG (4–30 sessions), Phase 0 Gate | `/mcp__vivechak__project` or `vivechak_start(scope="project", question="...")` |

---

## 2. Mode A — With Vivechak MCP Server (`vivechak_*` tools available)

When the `vivechak` (`vck`) MCP server is connected, follow the tool sequences below and honor `next` in every response:

### Scope 1: Comparison Scope (`ScopeComparison`)
Use when asked to compare, benchmark, or choose between 2–5 named technologies, libraries, or datastores:
1. **Trigger Start:** Call `vivechak_start(scope="comparison", question="<user prompt>")` (or `/mcp__vivechak__compare` in Claude Code).
2. **Execute Research:** Perform live web search for official documentation, reproducible benchmarks, and production postmortems. Grade every factual claim `A`–`E` with URLs, and include a Weighted Evaluation Protocol (WEP) matrix with a **±20% weight sensitivity check**.
3. **Persist Output:** Call `vivechak_save(kind="comparison", id="C-001", content="<matrix>")`. Saves `research/C-001-comparison.md`. Done!

### Scope 2: Decision Scope (`ScopeDecision`)
Use when resolving an architectural fork, migration, or dilemma in an active codebase:
1. **Trigger Start:** Call `vivechak_start(scope="decision", question="<dilemma>")` (or `/mcp__vivechak__decision` in Claude Code).
2. **Save Decision Plan:** Call `vivechak_save(kind="plan", id="D-001", content="<plan>")`. The plan needs one fenced `yaml vivechak-plan` block listing every session (`id`, `title`, `role`, `depends`) and one `prompt` fence per session under `### <id>: <title>` in `## Sessions`.
3. **Execute Sessions:** Run the 1–3 research sessions (`L` Landscape, `C` Comparison, `F` Falsification) via `vivechak_next()` and `vivechak_save(kind="session", id="D-001-S1", content="<session>")`.
4. **Record ADR:** Call `vivechak_record_decision(id="D-001", content="<adr>")` (or pass `auto_draft_from` with `auto_save=true`) to write the canonical ADR in `research/D-001-decision.md` and recompile `research/DECISIONS.md`.
5. **Red-Team & Gate (One-Way Doors):** Challenge One-Way Doors via `vivechak_review(mode="red_team", id="D-001")`, cite the returned `D-001-R1` findings (`>=50` chars) under `## Challenge Findings`, ask the user to run `vck signoff D-001 --reviewer "<name>"` in an interactive terminal, then verify with `vivechak_gate(decision="D-001")`. Never set `human_reviewed` yourself.

### Scope 3: Project Scope (`ScopeProject`)
Use when designing a greenfield platform or major subsystem rewrite from scratch:
1. **Trigger Start:** Call `vivechak_start(scope="project", question="<project vision>")` (or `/mcp__vivechak__project` in Claude Code).
2. **Save Pipeline:** Call `vivechak_save(kind="plan", id="DAG", content="<pipeline>")` to write `research/RESEARCH-PIPELINE.md`. Its one `yaml vivechak-plan` block lists every session with `depends` and exactly one `kind: synthesis` session; each session's prompt sits in one `prompt` fence under `### <id>: <title>` in `## Sessions`.
3. **Session Loop:** Loop through DAG sessions:
   - `vivechak_next()` → live web search → `vivechak_save(kind="session", id="<session_id>", content="<session>")` → `vivechak_record_decision(id="D-NNN", auto_draft_from="<session_id>", auto_save=true)`.
4. **Synthesize FAD:** Compile the grand synthesis session (`kind: synthesis`, e.g. `SYN-01`) into `research/FAD.md` via `vivechak_save(kind="fad", id="SYN-01", content="<fad>")`.
5. **Phase 0 Exit Gate & Human Sign-Off:** Run `vivechak_gate(track="all")` (or CLI `vck gate`) to verify Track A (reversibility) and Track B (rigor `G1`–`G12`). One-way decisions require `vck signoff <D-id> --reviewer "<name>"`; never set `human_reviewed` yourself. *(The gate is necessary, not sufficient: it checks structure and reported labels, not the ground-truth accuracy of external pages.)*

---

## 3. Mode B — Standalone Skill-Only Mode (Zero MCP Server / Zero Binary)

If `vivechak_*` MCP tools are **not** available in your harness, execute the exact same epistemic rigor directly using your native web search and file-write tools. Files written under `research/` remain 100% compatible with `vck doctor` and `vck gate` if the user installs the CLI later:

1. **Comparison Scope (`research/C-001-comparison.md`):**
   - Start with YAML frontmatter (`id: C-001`, `title`, `date: YYYY-MM-DD`, `status: complete`, `scope_level: comparison`).
   - Run live web search and include: `## Prior`, `## Recommendation` (isolated from alternatives), `## Weighted Evaluation Matrix` (2–5 options, weights summing to `1.0`, scores `1`–`5` with inline graded citations), `## Sensitivity Analysis (±20%)` (verify whether ±20% shifts on top criteria change the winner), `## Failure Modes & Disconfirming Evidence`, `## Delta`, and `## Sources & Evidence Ledger`.
2. **Decision Scope (`research/D-001-decision.md` + `research/D-001-R1.md`):**
   - Classify `door_type: two-way` (fast reversibility check) or `door_type: one-way` (hard to reverse).
   - Write YAML frontmatter (`id`, `type: decision`, `title`, `status: proposed`, `door_type`, `date`, `review_trigger`, `human_reviewed: false`) and body sections: `## Context & Problem Statement`, `## Decision Outcome` (>=2 URL-anchored Grade A/B `fetched` claims for One-Way Doors), `## Rejected Alternatives & Tradeoffs`, `## Challenge Findings`, and `## Failure Modes & Reversal Triggers`.
   - For One-Way Doors, actively search for disconfirming evidence and production postmortems, record them in `research/D-001-R1.md`, summarize `>=50` chars in `## Challenge Findings`, and **stop to ask the human engineer to review and approve** (never mark `human_reviewed: true` yourself).
3. **Project Scope (`research/RESEARCH-PIPELINE.md` → `research/sessions/*.md` → `research/FAD.md`):**
   - Decompose into a `yaml vivechak-plan` DAG with one terminal `kind: synthesis` node (`SYN-01`), execute independent layers with web search, record standalone `D-NNN-decision.md` ADRs, compile `research/FAD.md` (`Executive Summary`, `Locked Decision Registry`, `Risk Register`, `Traceability Matrix`), and surface a pre-coding gate summary before authoring code.

---

## 4. Evidence Grading Taxonomy (Mandatory)

Every factual claim in Vivechak must carry an inline evidence grade with modifiers and verification method — e.g. `[A · corroborated · fresh · direct · fetched](https://docs.example.org/spec)`:
<!-- docgen:begin evidence-skill (generated from internal/method/evidence.json by tools/docgen; do not edit by hand) -->
- **Grade A (Primary Empirical):** Official specifications, canonical source code, RFCs, or primary documentation with explicit URL anchor.
- **Grade B (Secondary Empirical):** Peer-reviewed papers, reproducible third-party benchmarks, published postmortems, or corroborated engineering reports.
- **Grade C (Vendor & Practitioner):** Vendor marketing, product whitepapers, vendor benchmarks, conference talks, or interested-party documentation.
- **Grade D (Informal & Opinion):** Blog posts, forum discussions, StackOverflow answers, or model memory recall (capped at Grade D).
- **Grade E (Speculative & Unsubstantiated):** Unverified hypotheses, uncorroborated parametric recall, or ungrounded assertions.

**Modifiers:** corroboration (`single` | `corroborated` | `contested`), recency (`fresh` | `aging` | `stale`), directness (`direct` | `indirect`).

**Verification:** `fetched` | `cached` | `recalled` | `secondhand` | `human-provided` (or `human`).

**Rules:** Parametric memory recall is capped at **Grade D**. Grade A requires canonical URL or specification anchor.
<!-- docgen:end evidence-skill -->

---

## 5. Strict Guardrails & Anti-Patterns

1. 🚫 **NEVER use Vivechak for direct code authoring:** If the user asks to write a function, component, script, or unit test, write code directly. Do NOT invoke Vivechak tools.
2. 🚫 **NEVER use Vivechak for local bug fixing or syntax:** Do not invoke Vivechak for compiler errors, stack traces, or API parameter lookups.
3. 🚫 **NEVER answer architectural decisions from ungrounded recall:** If the user asks for a comparison or architectural choice, ground the evaluation in live fetched evidence.
4. 🚪 **Every locked ADR requires Quantitative Reversal Triggers:** Specify exact measurable thresholds (e.g. `p95 latency > 50ms`, `event volume > 150k/s`, or a calendar date) that trigger revisiting the decision.
5. ✍️ **NEVER self-sign One-Way Decisions:** One-way decisions require human review (`vck signoff <D-id> --reviewer "<name>"` when using CLI/MCP, or explicit human confirmation in standalone mode); never set `human_reviewed: true` yourself.
