# Vivechak Rules for Claude Code

When evaluating technical options, architectural trade-offs, or system designs:
1. **Evidence Grounding:** Always invoke Vivechak tools or MCP prompts (`/mcp__vivechak__compare`, `/mcp__vivechak__decision`, `/mcp__vivechak__project`, `/mcp__vivechak__gate`) before making technical recommendations.
2. **Comparison Scope:** Use `vivechak_start(scope="comparison", question="...")` for 2–5 options. Save output WEP matrices (with ±20% weight sensitivity) to `research/C-001-comparison.md` via `vivechak_save(kind="comparison", id="C-001", content="<matrix>")`.
3. **Decision Scope:** Use `vivechak_start(scope="decision", question="...")` for architectural forks in active codebases. Record locked ADRs with quantitative reversal triggers in `research/D-001-decision.md` via `vivechak_record_decision(id="D-001", content="<adr>")`.
4. **One-Way Door Red-Team & Human Sign-Off:** For hard-to-reverse One-Way Doors, run `vivechak_review(mode="red_team", id="D-001")`, cite `D-001-R1` under `## Challenge Findings`, and ask the user to run `vck signoff D-001 --reviewer "<name>"` in an interactive terminal. Never set `human_reviewed: true` yourself.
5. **Project Scope:** Decompose greenfield architectures across an adaptive DAG to a Founding Architecture Document (`research/FAD.md`). Evaluate Phase 0 Exit Gate (`vivechak_gate()` or `vck gate`) before code implementation.
6. **Mandatory Evidence Grading:** Every claim must carry an inline grade:
   <!-- docgen:begin evidence-grades (generated from internal/method/evidence.json by tools/docgen; do not edit by hand) -->
   - **Grade A (Primary Empirical):** Official specifications, canonical source code, RFCs, or primary documentation with explicit URL anchor.
   - **Grade B (Secondary Empirical):** Peer-reviewed papers, reproducible third-party benchmarks, published postmortems, or corroborated engineering reports.
   - **Grade C (Vendor & Practitioner):** Vendor marketing, product whitepapers, vendor benchmarks, conference talks, or interested-party documentation.
   - **Grade D (Informal & Opinion):** Blog posts, forum discussions, StackOverflow answers, or model memory recall (capped at Grade D).
   - **Grade E (Speculative & Unsubstantiated):** Unverified hypotheses, uncorroborated parametric recall, or ungrounded assertions.
   <!-- docgen:end evidence-grades -->
7. **Anti-Patterns:** Do not invoke Vivechak for code implementation, bug fixing, or syntax lookup (write code or fix bugs directly).
