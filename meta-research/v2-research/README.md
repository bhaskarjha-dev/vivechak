# v2-Research — MCP Server Evidence Base

> **Sealed:** 2026-09-27. These files are historical records of the research that justified v2.0 decisions.

## What's Here

This directory contains the empirical evidence base for Vivechak's MCP server (Phase 2), distribution strategy (Phase 3), and methodology expansion (Phase 1). It was produced using Vivechak's own methodology — a 6-session research pipeline executed twice (with and without repository context).

### Research Pipeline

| File | Purpose |
|---|---|
| `RESEARCH-PIPELINE.md` | The pipeline definition (6 sessions + SYN-01) that drove all 12 sessions |
| `DECISIONS.md` | Research decisions D-011 through D-014 (MCP pattern, scope model, tool API, distribution) |

### Research Sessions (~770KB)

Each session was executed twice:
- **`with-repo-link/`** — AI had access to the Vivechak GitHub repository. Produced more targeted, internally consistent findings.
- **`without-repo-link/`** — AI worked from the brief alone. Discovered environmental findings (PATH truncation, CVEs, client-specific bugs) that the targeted sessions missed.

**Both runs are preserved because they contain genuinely different findings.** The divergence itself is evidence for the value of running research from multiple perspectives.

| Session | Topic | Key Decision Informed |
|---|---|---|
| R-01 | Competitor deep dive | Market positioning, threat assessment |
| R-02 | MCP server architecture | D-011 (Guided Worker), state management, Go SDK patterns |
| R-03 | Multi-scope methodology | D-012 (3 scope levels), Scope × Depth independence |
| R-04 | MCP tool API design | D-013 (9 tools), validation ladder, response envelope |
| R-05 | Agent Plugins & distribution | D-014 (binary-first), per-host reality matrix |
| R-06 | Generator design | GENERATOR-DECISION.md, COMPARISON-SESSION.template.md |

### Strategy Documents

| File | Purpose |
|---|---|
| `FINAL-PLAN.md` | The definitive v2.0 plan — 5 phases, 11 locked decisions, risk register |
| `PYTHON-VS-GO.md` | Evaluation of Python vs Go for the MCP server (verdict: Go) |
| `what-next-old.md` | Pre-research strategic analysis (historical — ~70% accurate) |
| `WHATS-NEXT.md` | Updated pre-research analysis (historical — ~80% accurate) |
| `ROADMAP-NEXT.md` | Pre-research roadmap (historical — superseded by FINAL-PLAN) |

## Why Two Runs?

Running research both with and without repository context was not planned redundancy — it was an accidental experiment that produced a valuable finding:

- **With-repo sessions** found internal consistency issues, generated draft artifacts (GENERATOR-DECISION.md, COMPARISON-SESSION.template.md), and produced tighter recommendations aligned with existing terminology.
- **Without-repo sessions** discovered environmental constraints (macOS PATH truncation, 5 CVEs in Anthropic MCP servers, Claude Code tool-search deferral) that the targeted sessions missed entirely.

This dual-context approach is now recognized as a legitimate research strategy for future Vivechak pipelines.
