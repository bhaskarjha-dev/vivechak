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

- **Context Budget:** The size of the 13-tool schema was not recorded. After consolidation, the nine tool definitions in the `tools/list` response are 19,419 bytes at commit 12b768f (measured 2026-11-12), so consolidation alone does not meet the ≤10 KB budget [Grade A: tools/list response of the built binary | fetched].
- **Agent Accuracy:** Harness runs recorded agents unsure whether `vivechak_init` was required before saving (D-10-01) and passing decision ids for comparisons (D-01-01, D-02-01); no failure rate was measured. Polymorphic `vivechak_save(kind=...)` removes the separate init step and the plan/session split [Grade C: tests/harness/DEFECTS.md | cached].

## Recommendation

Consolidate tool surface into 9 tools: `vivechak_start`, `vivechak_save`, `vivechak_record_decision`, `vivechak_next`, `vivechak_status`, `vivechak_review`, `vivechak_amend`, `vivechak_replan`, and `vivechak_gate`.

## Alternatives Considered

| Option | Verdict | Key Tradeoff |
|---|---|---|
| **9 Cohesive Tools** | **Recommended** | Optimal balance between token economy and explicit semantic discovery |
| **13 Granular Tools** | **Rejected** | Recorded agent friction with the init step and id prefixes; schema size not recorded |

## Open Questions & Risks

| Risk / Question | Impact | Mitigation |
|---|---|---|
| Legacy clients calling deprecated tool names | Low | Return clear deprecation guidance mapping old tool names to consolidated ones |

## Discovered Concerns

During multi-harness testing with smaller parameter models, we discovered that models frequently enter recursive help loops when tool descriptions exceed 500 characters. Keeping server instructions ≤600 characters and tool schemas concise is critical to avoid context poisoning.

## Delta

| Prior Belief | Status | Evidence | Impact |
|---|---|---|---|
| 13 granular tools provide better precision | Contradicted | Harness runs recorded agents unsure about `vivechak_init` and id prefixes with 13 tools (tests/harness/DEFECTS.md); no error rate was measured | High |

## Sources & Evidence Ledger

| # | Source | Grade | Modifiers | Verification | Used For |
|---|---|---|---|---|---|
| 1 | https://modelcontextprotocol.io/specification | A | corroborated · fresh | fetched | MCP tool schemas and capabilities |
| 2 | tests/harness/DEFECTS.md | C | corroborated · aging · direct | cached | Recorded agent friction with the 13-tool surface (D-01-01, D-02-01, D-10-01) |
| 3 | `tools/list` response of the binary built at commit 12b768f | A | single · fresh · direct | fetched | Size of the nine tool definitions (19,419 bytes, measured 2026-11-12) |
