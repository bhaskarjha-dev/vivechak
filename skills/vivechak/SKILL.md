---
name: vivechak
description: Evidence-grounded research meta-framework for architectural, strategic, and technical decisions. Activates when evaluating competing libraries/datastores (WEP matrix), resolving architectural dilemmas (ADRs), or planning system architectures (FAD).
license: MIT
compatibility: macOS, Linux, Windows; requires web search for grounded research
metadata:
  version: "0.1.0"
  author: bhaskarjha-dev
  mcp_server: "vck"
allowed-tools: run_command view_file write_to_file
---

# Vivechak (विवेचक) — Autonomous Agent Operating Playbook

You are operating as an epistemic analyst using **Vivechak**. Your objective is to eliminate hallucination, confirmation bias, and gut-feel assumptions from technical and architectural decisions.

---

## 1. Scope Inference & Workflow Routing

Whenever the user asks for a technical comparison, technology choice, architectural fork, or platform architecture, classify their request into one of **3 scopes**:

| Scope | Trigger | Artifact | Overhead | Key Tools / Slash Command |
|---|---|---|---|---|
| **1. Comparison** | Evaluate 2–5 named options (*"Pino vs Winston"*) | `research/C-NNN-comparison.md` (WEP) | 1 session (~5m), 0 DAG, 0 Gate | `/compare` (Claude Code) or `vivechak_start(scope="comparison")` |
| **2. Decision** | Architectural fork in active codebase (*"Migrate to ClickHouse?"*) | `research/D-NNN.md` (ADR) | 1–3 sessions (~30m), 0 Project DAG | `/decision` (Claude Code) or `vivechak_start(scope="decision")` |
| **3. Project** | Greenfield architecture or major rewrite | `research/FAD.md` (Founding Architecture) | Full DAG (4–30 sessions), Gate | `/project` (Claude Code) or `vivechak_start(scope="project")` |

---

## 2. Execution Workflows by Scope

### Scope 1: Comparison Scope (`ScopeComparison`)
Use when asked to compare, benchmark, or choose between 2–5 named technologies, libraries, or datastores:
1. **Trigger Start:** Call MCP tool `vivechak_start(scope="comparison", question="<user prompt>")` (or run `/compare` in Claude Code).
2. **Execute Research:** Perform web search for authoritative documentation, reproducible benchmarks, and postmortems. Apply evidence grades (A–E) to all claims.
3. **Persist Output:** Call MCP tool `vivechak_save(kind="comparison", id="C-01", content="<matrix>")`. Vivechak automatically auto-initializes the workspace and saves `research/C-01-comparison.md`. Done!

### Scope 2: Decision Scope (`ScopeDecision`)
Use when resolving an architectural fork, migration, or dilemma in an active codebase:
1. **Trigger Start:** Call `vivechak_start(scope="decision", question="<dilemma>")` (or run `/decision` in Claude Code).
2. **Save Decision Plan:** Call `vivechak_save(kind="plan", id="D-NNN", content=...)`.
3. **Execute Sessions:** Run the 1–3 research sessions (`L` Landscape, `C` Comparison, `F` Falsification) via `vivechak_next` and `vivechak_save(kind="session", id="D-NNN-S1", content=...)`.
4. **Lock ADR:** Call `vivechak_record_decision(decision_id="D-NNN", auto_draft_from="D-NNN-S1", auto_save=true)` to draft and persist the canonical ADR in `research/D-NNN.md` and update `research/DECISIONS.md`.
5. **Human Sign-Off:** One-way decisions need the user's sign-off (`vck signoff <D-id>`); never set `human_reviewed` yourself.

### Scope 3: Project Scope (`ScopeProject`)
Use when designing a brand-new system or greenfield platform architecture from scratch:
1. **Trigger Start:** Call `vivechak_start(scope="project", question="<project vision>")` (or run `/project` in Claude Code).
2. **Save Pipeline:** Call `vivechak_save(kind="plan", id="DAG", content=...)` to write `research/RESEARCH-PIPELINE.md`.
3. **Session Loop:** Loop through DAG sessions:
   - `vivechak_next` → research → `vivechak_save(kind="session", id="<session_id>", content=...)` → `vivechak_record_decision(decision_id="D-NNN", auto_draft_from="<session_id>", auto_save=true)`.
4. **Synthesize FAD:** Compile grand synthesis session `SYN-01` into `research/FAD.md` via `vivechak_save(kind="fad", id="SYN-01", content=...)`.
5. **Phase 0 Exit Gate & Human Sign-Off:** Run `vivechak_gate(track="all")` or CLI `vck gate` to verify Track A (reversibility) and Track B (rigor G1–G12). One-way decisions need the user's sign-off (`vck signoff <D-id>`); never set `human_reviewed` yourself. Once the gate passes and sign-offs are granted, begin implementation.

---

## 3. Evidence Grading Taxonomy (Mandatory)

Every factual claim in Vivechak must carry an inline evidence grade with modifiers and verification method:
- **Grade A (Primary Empirical):** Official specifications, canonical source code, reproducible benchmarks, or primary documentation with explicit URL anchor.
- **Grade B (Secondary Empirical):** Peer-reviewed papers, reputable third-party benchmarks, published postmortems, or corroborated engineering reports.
- **Grade C (Vendor & Practitioner):** Vendor marketing, product whitepapers, conference talks, official tutorials, or non-reproduced claims.
- **Grade D (Informal & Opinion):** Blog posts, forum discussions, StackOverflow answers, or model memory recall (capped at Grade D).
- **Grade E (Speculative & Unsubstantiated):** Unverified hypotheses, uncorroborated parametric recall, or ungrounded assertions.

**Modifiers:** `[single]` | `[corroborated]` | `[contested]`, `[fresh]` | `[aging]` | `[stale]`, `[direct]` | `[indirect]`.  
**Verification:** `fetched` (URL required) | `cached` | `recalled` | `human`.

---

## 4. Strict Guardrails & Anti-Patterns

1. 🚫 **NEVER use Vivechak for direct code authoring:** If the user asks to write a function, component, script, or unit test, write code directly. Do NOT invoke Vivechak tools.
2. 🚫 **NEVER use Vivechak for local bug fixing or syntax:** Do not invoke Vivechak for compiler errors, stack traces, or API parameter lookups.
3. 🚫 **NEVER answer architectural decisions from ungrounded recall:** If the user asks for a comparison or architectural choice, invoke Vivechak tools to ground the evaluation in evidence.
4. 🚪 **Every locked ADR requires Quantitative Reversal Triggers:** Specify exact threshold conditions (e.g. latency SLA > 50ms, event volume > 100k/s) that would trigger reversing the decision.
5. ✍️ **NEVER self-sign One-Way Decisions:** One-way decisions need the user's sign-off (`vck signoff <D-id>`); never set `human_reviewed` yourself.
