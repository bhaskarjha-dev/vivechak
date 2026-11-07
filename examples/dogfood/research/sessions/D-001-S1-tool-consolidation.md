---
id: D-001-S1
title: "Tool Consolidation Analysis"
date: "2026-10-24"
status: complete
topic: "tool-consolidation"
tags: ["mcp", "tools"]
informs_decisions: ["D-001"]
confidence: high
schema_version: "0.1.0"
---

# D-001-S1: Tool Consolidation Analysis

## Prior

1. I believe having 13 specialized tools provides precise functional demarcation for host agents.
2. I believe smaller models might struggle with overloaded parameters if tools are merged.

## Research Question

Does consolidating the MCP tool surface from 13 tools down to 9 cohesive tools reduce agent tool-selection failures while staying within the context budget?

## Key Findings

- **Context Budget:** The 13-tool schema required ~18 KB of prompt tokens. Consolidating to 9 tools reduces schema footprint to 8.4 KB, yielding a 53% token reduction [Grade A: source inspect | verified].
- **Agent Accuracy:** Autonomous multi-harness benchmarks showed an initial 22% failure rate where smaller models confused `save_plan` vs `save_session`. Polymorphic `vivechak_save(kind=...)` eliminates this confusion [Grade B: benchmark runs | cached].

## Recommendation

Consolidate tool surface into 9 tools: `vivechak_start`, `vivechak_save`, `vivechak_record_decision`, `vivechak_next`, `vivechak_status`, `vivechak_review`, `vivechak_amend`, `vivechak_replan`, and `vivechak_gate`.

## Alternatives Considered

| Option | Verdict | Key Tradeoff |
|---|---|---|
| **9 Cohesive Tools** | **Recommended** | Optimal balance between token economy and explicit semantic discovery |
| **13 Granular Tools** | **Rejected** | Triggers tool confusion in smaller models and consumes >18 KB context |

## Open Questions & Risks

| Risk / Question | Impact | Mitigation |
|---|---|---|
| Legacy clients calling deprecated tool names | Low | Return clear deprecation guidance mapping old tool names to consolidated ones |

## Discovered Concerns

During multi-harness testing with smaller parameter models, we discovered that models frequently enter recursive help loops when tool descriptions exceed 500 characters. Keeping server instructions ≤600 characters and tool schemas concise is critical to avoid context poisoning.

## Delta

| Prior Belief | Status | Evidence | Impact |
|---|---|---|---|
| 13 granular tools provide better precision | Contradicted | Small models suffer 22% tool selection error rate with 13 tools | High |

## Sources & Evidence Ledger

| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | https://modelcontextprotocol.io/specification | A | corroborated · fresh | fetched | MCP tool schema budget and capabilities |
| 2 | tests/harness/reports/audit-v0.1.0-2026-10-20.md | B | direct · fresh | cached | Empirical multi-harness failure rates |
