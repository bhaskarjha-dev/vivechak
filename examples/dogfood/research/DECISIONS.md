# Architectural Decisions Registry

This file records all architectural decisions accepted during Vivechak development. Canonical definitions live in individual `research/D-NNN-decision.md` files.

---

<!-- DECISION: D-001 -->
<details open>
<summary><b>D-001: Consolidate MCP Tool Surface from 13 to 9 Tools</b> · <code>two-way</code> · <i>accepted</i></summary>

```yaml
id: D-001
title: "Consolidate MCP Tool Surface from 13 to 9 Tools"
status: accepted
door_type: two-way
date: "2026-10-24"
confidence: high
evidence_refs: ["E-001", "E-002"]
review_trigger: "Re-evaluate if frontier LLMs demonstrate zero tool-confusion on >12 distinct MCP tool schemas"
review_date: "2027-04-24"
prediction: "Consolidation to 9 tools reduces tool-selection errors and prompt tokens by >35% across multi-harness benchmarks"
tags: ["mcp", "architecture", "tool-consolidation"]
authored_by: "agent-vivechak"
human_reviewed: false
schema_version: "0.1.0"
```

### Context & Problem Statement
Prior to v0.1.0, Vivechak exposed 13 distinct MCP tools. In multi-harness autonomous testing across client hosts, smaller models were unsure whether `init` had to run before saving and used decision ids for comparisons (tests/harness/DEFECTS.md); the schema size was not recorded.

### Decision Outcome
**Chosen Option:** Consolidate the tool surface from 13 to 9 cohesive MCP tools (`vivechak_start`, `vivechak_save`, `vivechak_record_decision`, `vivechak_next`, `vivechak_status`, `vivechak_review`, `vivechak_amend`, `vivechak_replan`, `vivechak_gate`).

### Rationale
- Guided Worker protocol with structured `next` actions
- Zero-ceremony transparent initialization on the fly
- Context budget met after schema trimming: `tools/list` is 6,115 bytes and always-loaded context is 9,063 bytes (down from 19,419 bytes at commit 12b768f; budget ≤10 KB)

### Reversal Trigger
Re-evaluate if future benchmarks across ≥3 major agent hosts demonstrate that separate `save_plan` and `save_session` tools outperform polymorphic `vivechak_save`.

</details>
<!-- /DECISION: D-001 -->

---

<!-- DECISION: D-002 -->
<details open>
<summary><b>D-002: Workspace Storage Model: Confined File-Based Store vs SQLite</b> · <code>two-way</code> · <i>accepted</i></summary>

```yaml
id: D-002
title: "Workspace Storage Model: Confined File-Based Store vs SQLite"
status: accepted
door_type: two-way
date: "2026-10-24"
confidence: high
evidence_refs: ["E-003", "E-004"]
review_trigger: "Re-evaluate if concurrent multi-agent write contention causes measurable lock latency >500ms"
review_date: "2027-04-24"
prediction: "Confined markdown/JSON store preserves human inspectability, git diff fidelity, and zero external CGO runtime dependencies"
tags: ["storage", "cgo-free", "architecture"]
authored_by: "agent-vivechak"
human_reviewed: false
schema_version: "0.1.0"
```

### Context & Problem Statement
Vivechak persists research artifacts across multiple interactive sessions and subagents. Requirements: zero external CGO runtime dependencies, human readability in Git PRs, and strict directory confinement.

### Decision Outcome
**Chosen Option:** Confined File-Based Store (`internal/store`) using Go 1.24+ `os.Root` confinement, atomic temporary file replacement, and advisory file locking.

### Rationale
- 100% CGO-free static binary distribution across 7 OS/architecture targets
- First-class version-controlled Markdown artifacts for human code reviews
- Writes confined to `research/` in the resolved project root (`os.Root`); the agent chooses that root

### Reversal Trigger
If multi-agent research workloads with >10 concurrent subagents experience lock contention latency exceeding 500ms, evaluate introducing an embedded pure-Go transactional engine for intermediate index state.

</details>
<!-- /DECISION: D-002 -->
