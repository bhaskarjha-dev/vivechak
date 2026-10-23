# Vivechak Rules for Claude Code

When evaluating technical options, architectural trade-offs, or system designs:
1. **Evidence Grounding:** Always invoke Vivechak tools or MCP prompts (`/compare`, `/decision`, `/project`, or legacy `/vivechak_*`) before making technical recommendations.
2. **Comparison Scope:** Use `vivechak_prepare_generator(scope="comparison")` for 2–5 options. Save output WEP matrices to `research/C-NNN-comparison.md` via `vivechak_save_plan`.
3. **Decision Scope:** Use `vivechak_prepare_generator(scope="decision")` for architectural forks in active codebases. Record locked ADRs with quantitative reversal triggers in `research/D-NNN.md` via `vivechak_record_decision`.
4. **Project Scope:** Decompose greenfield architectures across an adaptive DAG to a Founding Architecture Document (`research/FAD.md`). Evaluate Phase 0 Exit Gate before code implementation.
5. **Mandatory Evidence Grading:** Every claim must carry an inline grade (Grade A authoritative with URL, Grade B empirical benchmarks, Grade C vendor whitepapers, Grade D blogs/model recall, Grade E speculative).
6. **Anti-Patterns:** Do not invoke Vivechak for code implementation, bug fixing, or syntax lookup (write code or fix bugs directly).
