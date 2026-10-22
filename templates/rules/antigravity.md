# Vivechak Skill for Google Antigravity

This skill operationalizes Vivechak's epistemic grounding meta-framework inside Google Antigravity agents.

## When to Activate
Activate this skill whenever the user asks to:
- Compare 2–5 competing technologies, libraries, algorithms, or datastores.
- Resolve an architectural dilemma or migration fork in an active codebase.
- Architect a greenfield platform, system, or comprehensive subsystem rewrite.
- Evaluate pre-codebase exit criteria via the Phase 0 Gate.

## Scope Routing & Execution
1. **Comparison Scope (`ScopeComparison`):**
   - Call `vivechak_prepare_generator(scope="comparison", context="...")`.
   - Run grounded web search.
   - Save output via `vivechak_save_plan(content=..., scope="comparison")`.
2. **Decision Scope (`ScopeDecision`):**
   - Call `vivechak_prepare_generator(scope="decision", context="...")`.
   - Save decision plan via `vivechak_save_plan(scope="decision", content=...)`.
   - Execute 1–3 sessions (`vivechak_next_session` → `vivechak_save_session`).
   - Lock ADR via `vivechak_record_decision(decision_id="D-NNN", auto_draft_from="...", auto_save=true)`.
3. **Project Scope (`ScopeProject`):**
   - Call `vivechak_prepare_generator(scope="project", context="...")`.
   - Save DAG via `vivechak_save_plan(pipeline_content=..., scope="project")`.
   - Loop sessions to `FAD.md`.
   - Evaluate pre-codebase exit gate via `vivechak_run_gate`.

## Anti-Pattern Boundary
Do NOT invoke Vivechak for:
- Writing functions, components, API endpoints, or scripts.
- Debugging stack traces, fixing syntax errors, or repairing unit tests.
- Local refactoring or code cleanup.
