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
| **1. Comparison** | Evaluate 2–5 named options (*"Pino vs Winston"*) | `research/C-NNN-comparison.md` (WEP) | 1 session (~5m), 0 DAG, 0 Gate | `/compare` *(or `/vivechak_compare`)* or `vivechak_prepare_generator(scope="comparison")` |
| **2. Decision** | Architectural fork in active codebase (*"Migrate to ClickHouse?"*) | `research/D-NNN.md` (ADR) | 1–3 sessions (~30m), 0 Project DAG | `/decision` *(or `/vivechak_decision`)* or `vivechak_prepare_generator(scope="decision")` |
| **3. Project** | Greenfield architecture or major rewrite | `research/FAD.md` (Founding Architecture) | Full DAG (4–30 sessions), Gate | `/project` *(or `/vivechak_project`)* or `vivechak_prepare_generator(scope="project")` |

---

## 2. Execution Workflows by Scope

### Scope 1: Comparison Scope (`ScopeComparison`)
Use when asked to compare, benchmark, or choose between 2–5 named technologies, libraries, or datastores:
1. **Trigger Generator:** Call MCP tool `vivechak_prepare_generator(scope="comparison", context="<user prompt>")` or run `/compare` *(or `/vivechak_compare`)*.
2. **Execute Research:** Perform web search for authoritative documentation, reproducible benchmarks, and postmortems. Apply evidence grades (A–E) to all claims.
3. **Persist Output:** Call MCP tool `vivechak_save_plan(content=<matrix>, scope="comparison")`. Vivechak automatically auto-initializes the workspace and saves `research/C-NNN-comparison.md`. Done!

### Scope 2: Decision Scope (`ScopeDecision`)
Use when resolving an architectural fork, migration, or dilemma in an active codebase:
1. **Trigger Generator:** Call `vivechak_prepare_generator(scope="decision", context="<dilemma>")` or run `/decision` *(or `/vivechak_decision`)*.
2. **Save Decision Plan:** Call `vivechak_save_plan(scope="decision", content=...)`.
3. **Execute Sessions:** Run the 1–3 research sessions (`L` Landscape, `C` Comparison, `F` Falsification) via `vivechak_next_session` and `vivechak_save_session`.
4. **Lock ADR:** Call `vivechak_record_decision(decision_id="D-NNN", auto_draft_from="<session_id>", auto_save=true)` to draft and persist the canonical ADR in `research/D-NNN.md` and update `research/DECISIONS.md`.

### Scope 3: Project Scope (`ScopeProject`)
Use when designing a brand-new system or greenfield platform architecture from scratch:
1. **Prepare Pipeline:** Call `vivechak_prepare_generator(scope="project", context="<project vision>")` or run `/project` *(or `/vivechak_project`)*.
2. **Save Pipeline:** Call `vivechak_save_plan(pipeline_content=..., scope="project")` to write `research/RESEARCH-PIPELINE.md`.
3. **Session Loop:** Loop through DAG sessions:
   - `vivechak_next_session` → research → `vivechak_save_session` → `vivechak_record_decision(auto_draft_from=..., auto_save=true)`.
4. **Synthesize FAD:** Compile grand synthesis session `SYN-01` into `research/FAD.md`.
5. **Phase 0 Exit Gate:** Run `vivechak_run_gate(track="all")` or run `/vivechak_gate` to verify Track A (reversibility) and Track B (rigor B1–B11) before writing codebase implementation.

---

## 3. Evidence Grading Taxonomy (Mandatory)

Every factual claim in Vivechak must carry an inline evidence grade with modifiers and verification method:
- **Grade A (Primary / Authoritative):** Official vendor docs, RFC specs, source code, reproducible benchmarks. Requires URL or title.
- **Grade B (Empirical / Experimental):** Independent third-party reproducible benchmarks, production load tests.
- **Grade C (Vendor / Motivated):** Marketing whitepapers, sponsored posts.
- **Grade D (Secondary / Opinion / Model Recall):** Community blogs, tutorials, StackOverflow, and **any unverified LLM recall** (capped at Grade D).
- **Grade E (Speculative):** Unverified assertions.

**Modifiers:** `[single]` | `[corroborated]` | `[contested]`, `[fresh]` | `[aging]` | `[stale]`, `[direct]` | `[indirect]`.  
**Verification:** `fetched` (URL required) | `cached` | `recalled` | `human`.

---

## 4. Strict Guardrails & Anti-Patterns

1. 🚫 **NEVER use Vivechak for direct code authoring:** If the user asks to write a function, component, script, or unit test, write code directly. Do NOT invoke Vivechak tools.
2. 🚫 **NEVER use Vivechak for local bug fixing or syntax:** Do not invoke Vivechak for compiler errors, stack traces, or API parameter lookups.
3. 🚫 **NEVER answer architectural decisions from ungrounded recall:** If the user asks for a comparison or architectural choice, invoke Vivechak tools or slash commands to ground the evaluation in evidence.
4. 🚪 **Every locked ADR requires Quantitative Reversal Triggers:** Specify exact threshold conditions (e.g. latency SLA > 50ms, event volume > 100k/s) that would trigger reversing the decision.
