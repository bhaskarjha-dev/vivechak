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
   - Call `vivechak_start(scope="comparison", question="...")`.
   - Run grounded web search.
   - Save output via `vivechak_save(kind="comparison", id="C-001", content="<matrix>")`.
2. **Decision Scope (`ScopeDecision`):**
   - Call `vivechak_start(scope="decision", question="...")`.
   - Save decision plan via `vivechak_save(kind="plan", id="D-NNN", content="<plan>")`.
   - Execute 1–3 sessions (`vivechak_next()` → `vivechak_save(kind="session", id="D-NNN-S1", content="<session>")`).
   - Lock ADR via `vivechak_record_decision(id="D-NNN", auto_draft_from="D-NNN-S1", auto_save=true)`.
3. **Project Scope (`ScopeProject`):**
   - Call `vivechak_start(scope="project", question="...")`.
   - Save DAG via `vivechak_save(kind="plan", id="DAG", content="<pipeline>")`.
   - Loop sessions to `research/FAD.md` via `vivechak_save(kind="fad", id="SYN-01", content="<fad>")`.
   - Evaluate pre-codebase exit gate via `vivechak_gate()` or `vck gate`.

## Evidence Grades
Grade every factual claim:
<!-- docgen:begin evidence-grades (generated from internal/method/evidence.json by tools/docgen; do not edit by hand) -->
- **Grade A (Primary Empirical):** Official specifications, canonical source code, RFCs, or primary documentation with explicit URL anchor.
- **Grade B (Secondary Empirical):** Peer-reviewed papers, reproducible third-party benchmarks, published postmortems, or corroborated engineering reports.
- **Grade C (Vendor & Practitioner):** Vendor marketing, product whitepapers, vendor benchmarks, conference talks, or interested-party documentation.
- **Grade D (Informal & Opinion):** Blog posts, forum discussions, StackOverflow answers, or model memory recall (capped at Grade D).
- **Grade E (Speculative & Unsubstantiated):** Unverified hypotheses, uncorroborated parametric recall, or ungrounded assertions.
<!-- docgen:end evidence-grades -->

## Anti-Pattern Boundary
Do NOT invoke Vivechak for:
- Writing functions, components, API endpoints, or scripts.
- Debugging stack traces, fixing syntax errors, or repairing unit tests.
- Local refactoring or code cleanup.
