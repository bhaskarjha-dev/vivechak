---
id: R-04
title: "MCP Tool API Design for Vivechak"
date: 2026-09-23
status: draft
topic: api-design
tags: [mcp, api-design, tool-schema, evidence-grading, engine]
informs_decisions: [D-003]
confidence: high
---

# R-04: MCP Tool API Design for Vivechak
### Specifying the Tool Surface, Schemas, and Validation Contract for the Vivechak Engine's MCP Server

*API design research for Vivechak's Phase 5 "Engine" work (ROADMAP.md §Phase 5). Informs D-003. This is a one-way-door specification: tool names and schemas become a contract the moment an agent depends on them.*

---

## Research Question

What MCP tool surface — tool set, naming, input/output schemas, and validation strategy — should Vivechak expose so that AI agents can execute the framework's research-pipeline methodology (init → generate → save → execute sessions → decide → synthesize → gate) with evidence-grading and decision-tracking discipline structurally hard to skip, while staying simple enough for an agent to select the right tool on the first try, and stable enough to support agents built against it indefinitely?

---

## Key Findings

### What Vivechak itself already specifies (grounding, not invention)

The proposed 10-tool surface is not being designed in a vacuum — Vivechak's existing artifacts (`FRAMEWORK.md`, `GENERATOR.md`, `templates/*.md`, `ROADMAP.md`) already fix most of the content the MCP layer must enforce. This matters directly for validation design:

- **Evidence grading is already a closed spec**: Grade A–E by source class, three modifiers (`corroboration`, `recency`, `directness`), five `verification_method` values, and a hard rule that `recalled` evidence caps at Grade D and can never underpin a One-Way Door decision (FRAMEWORK.md §5, internal repo source — Grade A · corroborated · fresh · direct | fetched). The composite citation format (`Grade [A-E] · [corroboration] · [recency] · [directness] | [verification_method]`) is a parseable, regex-checkable string, which is what makes mechanical validation possible at all.
- **The session body skeleton and frontmatter fields are fixed**: seven required `##` sections, and eight required frontmatter keys (`id, title, date, status, topic, tags, informs_decisions, confidence`) per GENERATOR.md's DELIVERABLE spec (Grade A · corroborated · fresh · direct | fetched) — the same eight fields this document's own frontmatter uses, deliberately.
- **A research-output quality rubric with numeric floors already exists** (FRAMEWORK.md §4.3): ≥3 Grade A/B-backed claims, ≥2 options for comparison sessions, ≥1 failure mode per recommended option, explicit reversal triggers, source diversity, and at least one actively-sought disconfirming argument. This is the direct source for `vivechak_save_session`'s rubric checks below.
- **The Phase 0 Gate is already a 9-step checklist** (Track B) plus a 3-step fast track (Track A), with each step mechanically checkable against artifact metadata (`status: final`, `verification_method`, `review_trigger` presence, `human_reviewed`) except the premortem's and human sign-off's *substance*.
- **A fourth artifact type — Conflict Resolution (ACH matrix, `CHK-NN` IDs)** — exists in `templates/` but has no dedicated tool in the proposed 10. This is a genuine, if minor, gap in the original proposal (see Recommendation).
- **Vivechak already discovered, and reversed, the "reference a file" pattern this research would otherwise recommend for MCP Resources.** CHANGELOG.md v1.1.0 records that the ADR schema was inlined directly into `GENERATOR.md`, "replacing the unreachable file reference to `templates/DECISIONS.template.md` … the generating AI no longer needs filesystem access to produce correct decision records" (Grade A · corroborated · fresh · direct | fetched). This is first-party, pre-existing evidence for the Resources/Prompts recommendation below, arrived at independently of this session.
- **ROADMAP.md §Phase 5 already scopes the Engine's maturity ladder**: `vivechak init`, a YAML validator, and a pipeline executor are explicitly targeted at v0.1–v0.2; multi-model triangulation, checkpointing, and cost tracking are explicitly deferred to v1.0 (Grade A · corroborated · fresh · direct | fetched). The proposed 10-tool surface maps cleanly onto v0.1–v0.2 — none of the 10 tools call out to another LLM or execute research themselves, which matches the README's own positioning of Vivechak as "the orchestration layer above" deep-research tools, not a deep-research tool itself.

### Competitive landscape (context for D-003)

No tool was found that combines Vivechak's specific four properties — evidence grading with a verification-method hard cap, DAG-structured multi-session research pipelines, a Bezos-style reversibility gate, and Map-Reduce synthesis to a sealed architecture document — in one MCP surface. The nearest neighbors all operate one layer downstream, on decisions that already exist, rather than upstream on the research that produces them:

- **`mcp-adr-analysis-server`** (tosin2013) is the most feature-rich adjacent tool: 63 tools spanning ADR generation, Tree-sitter-based *drift detection* between ADRs and live code, content-safety scanning, and a persistent "decision memory" knowledge graph (Grade C · single · fresh · direct | fetched — project self-description, not independently benchmarked). It operates **during and after** code exists; Vivechak operates **before** any code exists. 63 tools is also useful calibration data: it shows the ADR-tooling niche tolerates much larger surfaces than 10 when the scope (code analysis + security + TDD + ADRs) is this broad — Vivechak's narrower, single-phase scope argues for staying well below it.
- **`AdrMcp`** (Atypical-Consulting, .NET) is the closest structural analogue for the decision-tracking half of Vivechak's surface: list/search, draft-from-template, validate structure and links, supersede-while-preserving-history — roughly 5–6 tools (Grade C · single · fresh · direct | fetched). Two design choices are directly reusable: its **preview-by-default, diff-before-write** posture, and its confirmation that a focused single-domain MCP surface this size is viable and shippable.
- **`sqlew`** stores ADRs in a queryable SQL database rather than git-native Markdown (Grade C · single · fresh · direct | fetched) — the opposite trade-off from Vivechak's git-diffable, human-readable Markdown+YAML choice (FRAMEWORK.md P6), which the framework's own meta-research separately investigated (E-023, git line-level diff mechanics across JSON/YAML/Markdown).
- **`kurdin` decision-cli** validates "missing sections, empty content" and normalizes formatting for clean diffs (Grade C · single · fresh · direct | fetched) — confirms that lightweight structural validation, independent of a save operation, is an expected capability in this space (supports keeping `vivechak_validate` as its own tool rather than folding it entirely into the save tools).
- The broader **MADR 4.0 ecosystem** (the community standard nearly every ADR tool above implements) captures `status`, `decision-drivers`, `considered-options`, and `consequences`, but has no equivalent of evidence grades, door-type classification, `review_trigger`/`review_date`/`prediction` calibration fields, or a verification-method hard cap (Grade B · corroborated · fresh · direct | fetched, across many independently-published MADR adoption pages). This is Vivechak's clearest point of differentiation and the reason `vivechak_record_decision`'s validation (below) is stricter than a generic ADR linter.

None of the tools surveyed expose an **MCP interface upstream of decision-making** — at the research-session, evidence-grading, and pipeline-DAG layer. This is consistent with, not just asserted by, Vivechak's own README positioning.

### MCP ecosystem patterns (context for D-003)

- **Anthropic's own tool-design guidance is unambiguous on consolidation**: "instead of implementing `list_users`, `list_events`, and `create_event` tools, consider implementing a `schedule_event` tool" — and warns that tools which merely wrap an API 1:1, or that overlap in function, measurably degrade agent tool-selection accuracy (Grade A · corroborated · aging · direct | fetched — Anthropic Engineering, "Writing effective tools for AI agents," Sep 2025). This is the primary test applied to each of the 10 proposed tools below.
- **Namespacing tool names by domain prefix (`asana_search`, `jira_search`) is Anthropic's explicit recommendation** to help agents select the right tool among many simultaneously-connected servers (Grade A · corroborated · aging · direct | fetched, same source). Notably, this is **not universal practice**: GitHub's own official MCP server does not prefix its 35 tools with `github_` (`create_pull_request`, `get_file_contents`, `delete_file`), relying instead on server-level identity and client-side grouping (Grade B · corroborated · fresh · direct | fetched, cross-checked across three independent listings). This tension is resolved explicitly in the Recommendation section rather than papered over.
- **Real single-server tool counts in production span a wide range**: Splunk 2, Sentry 5, Grafana 5, Slack 11, GitHub 35 (grouped into 9 toggleable toolsets, "strongly recommended" to enable only what's needed to help tool selection), `mcp-adr-analysis-server` 63 (Grade B · corroborated · fresh · direct | fetched — Anthropic's own reported figures for the first five, cross-checked against GitHub's own toolset documentation for the sixth). 10 tools sits comfortably inside this range for a single, coherent domain.
- **Anthropic's November 2025 "Tool Search Tool" (progressive/deferred tool loading) reduced context consumption by ~85% and raised Opus 4 MCP-evaluation accuracy from 49% to 74%** in Anthropic's internal benchmark, as reported by one independent technical blog (Grade C · single · fresh · indirect | secondhand — I did not independently fetch Anthropic's original benchmark writeup; this number should be re-verified against Anthropic's own materials before being load-bearing for a scoping decision). Directionally it lowers the cost of a larger tool count for clients that support it, but client support is not yet universal, so it should not be used to justify tool proliferation today.
- **MCP fixes three distinct control surfaces, not one**: Tools are *model-controlled* (the agent decides when to call), Resources are *application-controlled* (the host/client decides what to surface, e.g., a picker UI), and Prompts are *user-controlled* (a human explicitly selects them, typically as a slash command) (Grade A · corroborated · fresh · direct | fetched — official MCP specification, tools/resources/prompts pages, cross-checked across the 2025-06-18 and 2025-11-25 protocol revisions). One independent analysis observes that "most catalog servers ship only tools, leaving two-thirds of the protocol unused" (Grade C · single · fresh · indirect | fetched) — a real adoption-pattern risk for anything Vivechak makes load-bearing on Resources or Prompts alone.
- **Tool annotations** (`readOnlyHint`, `destructiveHint`, `idempotentHint`, `openWorldHint`) are optional, self-reported, client-facing hints — not a security boundary — used by clients to decide auto-approval versus confirmation prompts (Grade B · corroborated · fresh · direct | fetched, corroborated across five independent SDK/vendor documentation pages). Every Vivechak tool below declares these explicitly.
- **MCP versioning operates on (at least) three independent axes that should not be conflated**: the wire protocol version (date-stamped, negotiated at `initialize`), the server's own deployable-artifact version (semver, declared in `server.json`, required by the MCP Registry), and each tool's own contract version (Grade B · corroborated · fresh · direct | fetched — cross-checked between a vendor analysis and an independent technical blog covering the same three-axis model, plus the official MCP Registry versioning page for the server-artifact axis specifically). The protocol axis is evolving quickly enough that it should be treated as SDK-maintenance surface, not something Vivechak's tool design manages directly — one report describes a stateless-architecture proposal (SEP-2575) and related breaking changes moving through the spec process around mid-2026 (Grade B · single · fresh · direct | fetched — independent technical blog; not independently cross-verified against a second source, flagged accordingly).

---

## Recommendation

**Keep the proposed 10-tool surface.** Evaluated tool-by-tool against Anthropic's consolidation test and against every competitor surveyed, none of the 10 wrap a single API call 1:1, none obviously duplicate another, and the total sits well inside the observed real-world range for a single-domain server. Two refinements, not a redesign, are warranted:

1. **Broaden `vivechak_record_decision`'s scope to cover Conflict Resolution artifacts** (`artifact_type: decision | conflict_resolution`) instead of leaving them uncovered or adding an eleventh tool. Decisions and conflict-resolutions are structurally the same operation from an agent's point of view — validate a structured judgment record against a template, assign/link an ID, persist it — and conflict-resolution is invoked rarely enough (only on genuine cross-session or cross-model contradiction) that a dedicated tool would mostly sit unused, diluting tool-selection signal for the other nine. This is a judgment call, not a certainty — see Open Questions.
2. **Clarify, rather than merge, two pairs of tools an agent could otherwise conflate**: `vivechak_validate` vs. the save/record tools (dry-run check vs. check-and-commit), and `vivechak_status` vs. `vivechak_next_session` (lightweight dashboard vs. heavyweight, execution-ready payload). Both distinctions are encoded directly into tool descriptions and output shapes below, not left implicit.

### Cross-cutting design principles

| # | Principle | Why |
|---|---|---|
| 1 | **State lives in the workspace, not the server.** Every tool takes `project_root`; DAG/session/decision state is derived from files under `<project_root>/research/` on every call. | MCP servers are not guaranteed to stay warm between agent turns; multiple different agents (Claude Code today, a different one tomorrow) must be able to resume. Extends Vivechak's own "100% self-contained workspace" principle (ROADMAP.md OVH-03) to the Engine layer. |
| 2 | **Tools are the reliable surface; Resources and Prompts mirror, never gate.** Nothing an agent needs to make progress is available *only* via a Resource or Prompt. | Tools are model-controlled and universally supported; Resources are application-controlled and Prompts are user-controlled — both invisible to a fully headless agent unless the host chooses to surface them (see MCP ecosystem findings above). Vivechak already independently rediscovered this exact risk when it inlined the ADR schema into `GENERATOR.md` instead of referencing a template file. |
| 3 | **Validation is layered by lifecycle stage, not maximal everywhere.** Session-save checks session completeness; decision-lock checks that decision's evidentiary bar; the gate checks project-wide readiness. | Mirrors P2, Reversibility-Calibrated Rigor, applied to the tool layer itself rather than only to research effort. |
| 4 | **Block only on objectively-checkable structure; warn, never block, on holistic quality.** Frontmatter completeness, citation-format validity, and the evidence-density floor are blocking. Options-evaluated depth, failure-mode coverage, and disconfirming-evidence presence are warnings. | Directly avoids reintroducing the "rigid output skeleton" anti-pattern FRAMEWORK.md's own meta-research refuted (Tam et al., EMNLP 2024 — format restrictions measurably impair model reasoning). A validator that rejects prose for organizing itself differently is the exact failure mode P4 exists to prevent. |
| 5 | **`force` can only save to draft, never to final.** A blocking save can always be checkpointed as work-in-progress; it can never be checkpointed as `status: final`. | This is the specific mechanism that answers the brief's central tension: an agent is never fully blocked from saving its work, but nothing incomplete can silently reach the artifacts (`synthesize`, `run_gate`) that treat `final` as a trust signal. |
| 6 | **Every tool returns structured JSON (`structuredContent`, matching a declared `outputSchema`) plus a short natural-language summary.** | Matches the MCP spec's own structured-output guidance and Anthropic's finding that low-signal, hard-to-parse tool responses increase agent error rates. |
| 7 | **`vivechak_` prefix on every tool name, despite GitHub's own server not doing this.** | GitHub relies on client-side namespacing and toolset grouping that is not guaranteed uniform across the many different hosts (Claude Code, Claude Desktop, Cursor, a custom orchestrator) Vivechak's AGENTS.md explicitly targets. A self-describing name costs a few characters and degrades gracefully everywhere; relying on the client is a bet Vivechak doesn't need to make. |
| 8 | **Each tool's output breadcrumbs the next call** (`next_step`, `save_instructions`, `informs_decisions_notified` fields). | Turns the lifecycle order (init → get prompt → save pipeline → {next session ⇄ save session ⇄ record decision} → synthesize → gate) into something discoverable from tool outputs, not just tool descriptions read in isolation. |

### Final tool list (summary)

| Tool | MCP primitive(s) | One-line purpose | Annotations |
|---|---|---|---|
| `vivechak_init` | Tool | Bootstrap the `research/` workspace and copy the 4 templates | not read-only · not destructive · idempotent · closed-world |
| `vivechak_get_generator_prompt` | Tool **+ Prompt** | Return the GENERATOR.md prompt with project vision inlined | read-only · idempotent · closed-world |
| `vivechak_save_pipeline` | Tool | Parse & persist a generated pipeline; supports amendments | not read-only · destructive (on `overwrite`) · not idempotent · closed-world |
| `vivechak_status` | Tool **+ Resource** | Lightweight progress dashboard | read-only · idempotent · closed-world |
| `vivechak_next_session` | Tool | DAG-aware, execution-ready next session(s) | read-only · idempotent · closed-world |
| `vivechak_save_session` | Tool | Validate and persist one research session output | not read-only · not destructive · idempotent · closed-world |
| `vivechak_record_decision` | Tool | Record/update/supersede a decision or conflict-resolution | not read-only · destructive · not idempotent · closed-world |
| `vivechak_validate` | Tool | Dry-run structural check for any of the 6 artifact types | read-only · idempotent · closed-world |
| `vivechak_synthesize` | Tool | Gather & pre-structure Map-Reduce synthesis input for the FAD | read-only · idempotent · closed-world |
| `vivechak_run_gate` | Tool | Execute the Phase 0 Gate; seals the FAD on a full pass | not read-only · destructive (on seal) · not idempotent · closed-world |

*(All ten are `openWorldHint: false` — every operation is local-filesystem, none call an external API. This is itself a finding worth stating plainly: Vivechak's MCP layer manages state and enforces schema; it deliberately does not execute research or call other model APIs, consistent with the README's positioning and ROADMAP's v0.1–v0.2 scoping.)*

### MCP Resources and Prompts: recommendation

- **Expose `vivechak_status`'s data as a Resource** (`vivechak://{project}/status`) in addition to the tool. Status is read-only with no side effects — the textbook Resource case — and lets a human using a Resource-aware client (e.g., attaching current project state to a fresh chat) get context without an explicit tool call. Keep the tool as the reliable path.
- **Register `vivechak_get_generator_prompt`'s content as an MCP Prompt** (e.g., `generate-pipeline`) *in addition to* the tool. Semantically this is close to a perfect fit — GENERATOR.md is literally described as something a human pastes and sends — but Prompts are user-selected, not model-callable, so a fully headless agent may have no path to it at all if it is the *only* surface. Dual-expose; never make it Prompt-only. This mirrors a pattern already used by GitHub's official MCP server, which bundles matching Resources and Prompts alongside Tools per functional area rather than relying on any one primitive alone.
- **Do not expose the 4 templates themselves as Resources as the primary distribution path.** Vivechak already tried the reference-a-file pattern for the ADR schema and reversed it (CHANGELOG.md v1.1.0) in favor of inlining. The MCP layer should inline the same way: `vivechak_get_generator_prompt` returns the full prompt text in-line, `vivechak_record_decision`'s validation errors quote the exact missing field, etc. — never "see `templates/DECISIONS.template.md`."
- **Do not build anything load-bearing on MCP Prompts alone**, given the practitioner finding that most deployed MCP servers implement Tools only. Revisit if client-side Prompt support becomes more consistent (see Open Questions).

### Tool-granularity verdict, stated plainly

Ten tools, single coherent domain, zero 1:1 API wraps, two genuine consolidations already applied by the original proposal (`get_customer_context`-style: `vivechak_next_session` compiles DAG state + injected context + the actual prompt text in one call rather than three; `vivechak_synthesize` compiles Filter+Group+Map rather than requiring the agent to fetch every session individually). Evidence from six competitor tools (2–63 tools) and five production single-domain servers (2–35 tools) does not support either shrinking below 10 (nothing left is a plausible merge without losing a distinct control point) or growing past it (nothing surveyed needs an 11th, given the conflict-resolution consolidation above).

---

## Detailed API Specification

Every tool below takes `project_root` (string, required, absolute path to the project) as its first parameter — omitted from the per-tool listings after the first for brevity, but present in every schema. Every output includes an implicit `schema_version: "1.1"` field mirroring Vivechak's own content-level schema version (see Versioning, below) — also omitted below for brevity except where shown explicitly.

### Shared types

Four tools (`save_pipeline`, `save_session`, `record_decision`, `validate`) return a `validation` field of the same shape, referenced below as `$ref: "#/definitions/ValidationResult"`. Defining it once keeps the four call sites — and any future tool that validates something — guaranteed identical, which matters directly for Design Principle 4 (validation logic is one source of truth exposed in two ways, not reimplemented per tool):

```json
// #/definitions/ValidationResult
{
  "type": "object",
  "properties": {
    "valid": { "type": "boolean", "description": "True iff issues contains zero severity:blocking entries." },
    "issues": { "type": "array", "items": { "$ref": "#/definitions/ValidationIssue" } },
    "rubric_score": {
      "type": ["object", "null"],
      "description": "Present only for session-type artifacts; null otherwise.",
      "properties": {
        "evidence_density": { "type": "object", "properties": { "met": {"type": "boolean"}, "count": {"type": "integer"}, "required": {"type": "integer"} } },
        "options_evaluated": { "type": "object", "properties": { "met": {"type": "boolean"}, "count": {"type": "integer"}, "required": {"type": "integer"} } },
        "failure_modes": { "type": "object", "properties": { "met": {"type": "boolean"} } },
        "reversal_triggers": { "type": "object", "properties": { "met": {"type": "boolean"} } },
        "source_diversity": { "type": "object", "properties": { "met": {"type": "boolean"} } },
        "disconfirming_evidence": { "type": "object", "properties": { "met": {"type": "boolean"} } }
      }
    }
  },
  "required": ["valid", "issues"]
}
```
```json
// #/definitions/ValidationIssue
{
  "type": "object",
  "properties": {
    "check_id": { "type": "string", "description": "e.g. 'SESSION-08.evidence_density' — namespaced by artifact type and check, per the Validation Strategy tables below." },
    "severity": { "type": "string", "enum": ["blocking", "warning", "info"] },
    "location": { "type": "string", "description": "e.g. 'frontmatter.status' or 'body.Sources & Evidence Ledger'." },
    "message": { "type": "string" },
    "fix_hint": { "type": ["string", "null"] }
  },
  "required": ["check_id", "severity", "location", "message"]
}
```

### 1. `vivechak_init`
**Title:** Initialize Vivechak workspace
**Description (as an agent would see it):** "Creates `<project_root>/research/` and copies the 4 operational templates (Decisions, Conflict Resolution, Founding Architecture, Phase 0 Gate) into it. Call this once, before `vivechak_get_generator_prompt`. Safe to re-call on an existing workspace without `force` — it will not overwrite existing sessions or decisions."

```json
// input
{
  "type": "object",
  "properties": {
    "project_root": { "type": "string", "description": "Absolute path to the project directory." },
    "project_name": { "type": "string", "description": "Human-readable project name used in generated document titles." },
    "id_prefix": { "type": "string", "default": "T", "description": "Prefix for this project's session IDs (e.g. 'T' for a standard tiered pipeline)." },
    "force": { "type": "boolean", "default": false, "description": "Re-initialize an existing workspace, refreshing templates to the current schema_version while preserving existing sessions/ and decisions." }
  },
  "required": ["project_root", "project_name"]
}
```
```json
// output
{
  "type": "object",
  "properties": {
    "workspace_path": { "type": "string" },
    "created_paths": { "type": "array", "items": { "type": "string" } },
    "templates_version": { "type": "string" },
    "already_initialized": { "type": "boolean" },
    "next_step": { "type": "string" }
  },
  "required": ["workspace_path", "already_initialized"]
}
```
Annotations: `readOnlyHint: false`, `destructiveHint: false`, `idempotentHint: true`, `openWorldHint: false`.

### 2. `vivechak_get_generator_prompt`
**Title:** Get pipeline generator prompt
**Description:** "Returns the complete GENERATOR.md prompt with your project vision already inlined, ready to execute — either reason through it yourself, or hand it to a separate deep-research session. Produces two documents (`RESEARCH-PIPELINE.md`, `DECISIONS.md`); save both with `vivechak_save_pipeline`. Also available as the `generate-pipeline` MCP Prompt for clients with slash-command support."

```json
// input
{
  "type": "object",
  "properties": {
    "project_root": { "type": "string" },
    "project_vision": { "type": "string", "description": "Open-ended natural-language description of what should be researched." },
    "domain_hint": {
      "type": "string",
      "enum": ["b2b-saas", "developer-tools", "fintech", "ai-ml-systems", "consumer-mobile", "real-time-iot", "non-software", "auto-detect"],
      "default": "auto-detect",
      "description": "Optional steer for archetype classification; 'auto-detect' lets the executing agent classify from the vision text per FRAMEWORK.md §8."
    }
  },
  "required": ["project_root", "project_vision"]
}
```
```json
// output
{
  "type": "object",
  "properties": {
    "prompt_text": { "type": "string" },
    "expected_deliverables": { "type": "array", "items": { "type": "string" } },
    "framework_version": { "type": "string" },
    "save_instructions": { "type": "string" }
  },
  "required": ["prompt_text", "expected_deliverables"]
}
```
Annotations: `readOnlyHint: true`, `destructiveHint: false`, `idempotentHint: true`, `openWorldHint: false`.

### 3. `vivechak_save_pipeline`
**Title:** Save or amend the research pipeline
**Description:** "Parses a generated RESEARCH-PIPELINE.md (+ DECISIONS.md) and persists it, building the session DAG. Use `mode: \"initial\"` once, right after `vivechak_get_generator_prompt`. Use `mode: \"amend\"` later if the project vision evolves mid-pipeline (Progressive Elaboration Protocol, FRAMEWORK.md §3.1) — this adds or updates sessions without discarding completed research."

```json
// input
{
  "type": "object",
  "properties": {
    "project_root": { "type": "string" },
    "mode": { "type": "string", "enum": ["initial", "amend"], "default": "initial" },
    "research_pipeline_md": { "type": "string", "description": "Full raw content of RESEARCH-PIPELINE.md." },
    "decisions_md": { "type": "string", "description": "Full raw content of DECISIONS.md. Optional on amend if no new decisions were introduced." },
    "amendment_note": { "type": "string", "description": "Required when mode='amend': what changed and why." },
    "overwrite": { "type": "boolean", "default": false, "description": "Allow mode='initial' to replace an existing pipeline. Prefer mode='amend' for an evolving project." }
  },
  "required": ["project_root", "research_pipeline_md"]
}
```
```json
// output
{
  "type": "object",
  "properties": {
    "sessions_registered": { "type": "array", "items": { "type": "object", "properties": {
      "id": {"type": "string"}, "title": {"type": "string"}, "layer": {"type": "integer"},
      "door_type": {"type": "string", "enum": ["one-way", "two-way"]}, "status": {"type": "string"} } } },
    "decisions_registered": { "type": "array", "items": { "type": "object", "properties": {
      "id": {"type": "string"}, "title": {"type": "string"}, "status": {"type": "string"}, "door_type": {"type": "string"} } } },
    "dag_summary": { "type": "object", "properties": {
      "total_sessions": {"type": "integer"}, "layers": {"type": "integer"}, "parallel_eligible_now": {"type": "array", "items": {"type": "string"}} } },
    "validation": { "$ref": "#/definitions/ValidationResult" },
    "saved_paths": { "type": "array", "items": { "type": "string" } }
  },
  "required": ["sessions_registered", "validation"]
}
```
Annotations: `readOnlyHint: false`, `destructiveHint: true`, `idempotentHint: false`, `openWorldHint: false`.

### 4. `vivechak_status`
**Title:** Pipeline status
**Description:** "Lightweight progress dashboard: session/decision counts, open contested flags, and gate readiness. Returns only IDs and titles for sessions eligible to run next — call `vivechak_next_session` for the full, execution-ready payload. Use `detail: \"full\"` for per-item tables."

```json
// input
{
  "type": "object",
  "properties": {
    "project_root": { "type": "string" },
    "detail": { "type": "string", "enum": ["summary", "full"], "default": "summary" }
  },
  "required": ["project_root"]
}
```
```json
// output
{
  "type": "object",
  "properties": {
    "tier": { "type": "string" },
    "sessions": { "type": "object", "properties": {
      "total": {"type": "integer"}, "final": {"type": "integer"}, "in_progress": {"type": "integer"},
      "not_started": {"type": "integer"}, "blocked": {"type": "integer"} } },
    "decisions": { "type": "object", "properties": {
      "total": {"type": "integer"}, "accepted": {"type": "integer"}, "proposed": {"type": "integer"}, "one_way_open": {"type": "integer"} } },
    "eligible_next_sessions": { "type": "array", "items": { "type": "object", "properties": { "id": {"type": "string"}, "title": {"type": "string"} } } },
    "contested_flags_open": { "type": "integer" },
    "gate_readiness": { "type": "object", "properties": {
      "track_a_ready": {"type": "boolean"}, "track_b_ready": {"type": "boolean"}, "blocking_reasons": {"type": "array", "items": {"type": "string"}} } },
    "sessions_detail": { "type": ["array", "null"] },
    "decisions_detail": { "type": ["array", "null"] }
  },
  "required": ["sessions", "decisions", "gate_readiness"]
}
```
Annotations: `readOnlyHint: true`, `destructiveHint: false`, `idempotentHint: true`, `openWorldHint: false`. Mirrored as Resource `vivechak://{project}/status`.

### 5. `vivechak_next_session`
**Title:** Get next eligible session(s)
**Description:** "Returns up to `max_sessions` DAG-eligible sessions, each with its complete 5-block prompt, target output filename, and any injected context from upstream sessions — ready to execute immediately. A session only appears here once every hard dependency has reached `status: final`."

```json
// input
{
  "type": "object",
  "properties": {
    "project_root": { "type": "string" },
    "max_sessions": { "type": "integer", "default": 3, "minimum": 1, "maximum": 10 },
    "layer": { "type": "integer", "description": "Optional: restrict to one DAG layer." }
  },
  "required": ["project_root"]
}
```
```json
// output
{
  "type": "object",
  "properties": {
    "eligible_sessions": { "type": "array", "items": { "type": "object", "properties": {
      "id": {"type": "string"}, "title": {"type": "string"}, "layer": {"type": "integer"},
      "door_type": {"type": "string"}, "informs_decision": {"type": "string"},
      "prompt_text": {"type": "string"}, "output_path": {"type": "string"},
      "injected_context": { "type": "array", "items": { "type": "object", "properties": {
        "from_session": {"type": "string"}, "type": {"type": "string", "enum": ["constrains", "informs"]}, "summary": {"type": "string"} } } }
    } } },
    "remaining_blocked": { "type": "array", "items": { "type": "object", "properties": {
      "id": {"type": "string"}, "title": {"type": "string"}, "blocked_on": {"type": "array", "items": {"type": "string"}} } } },
    "pipeline_complete": { "type": "boolean" }
  },
  "required": ["eligible_sessions", "pipeline_complete"]
}
```
Annotations: `readOnlyHint: true`, `destructiveHint: false`, `idempotentHint: true`, `openWorldHint: false`.

### 6. `vivechak_save_session`
**Title:** Validate and save a research session
**Description:** "Validates a completed research session against Vivechak's frontmatter, structure, and evidence-grading rules, then saves it. If blocking issues are found, the file is still written but downgraded to `status: draft` so work is never lost — it will not count toward gate readiness or synthesis until re-submitted clean. Use `force: true` to save a known-incomplete draft explicitly; `force` can never produce a `status: final` save while blocking issues remain. Prefer `vivechak_validate` first if you want to check without writing anything."

```json
// input
{
  "type": "object",
  "properties": {
    "project_root": { "type": "string" },
    "session_id": { "type": "string", "description": "Must match an id already registered by vivechak_save_pipeline, e.g. 'T1-01'." },
    "content_md": { "type": "string", "description": "Full raw Markdown, including YAML frontmatter." },
    "force": { "type": "boolean", "default": false, "description": "Persist despite blocking issues, as status: draft." }
  },
  "required": ["project_root", "session_id", "content_md"]
}
```
```json
// output
{
  "type": "object",
  "properties": {
    "status": { "type": "string", "enum": ["saved_final", "saved_draft", "rejected"] },
    "session_id": { "type": "string" },
    "saved_path": { "type": ["string", "null"] },
    "status_downgraded": { "type": "boolean", "description": "True if the incoming status: final was overridden to draft due to blocking issues." },
    "validation": { "$ref": "#/definitions/ValidationResult" },
    "informs_decisions_notified": { "type": "array", "items": { "type": "string" }, "description": "D-NNN ids this session informs — a reminder to call vivechak_record_decision, not an automatic update." }
  },
  "required": ["status", "session_id", "validation"]
}
```
Annotations: `readOnlyHint: false`, `destructiveHint: false`, `idempotentHint: true`, `openWorldHint: false`.

**Worked example** — an agent submits a session missing a reversal trigger and with only single-sourced evidence on one claim:

```json
// tools/call result (isError: false — file was written as a draft, not rejected outright)
{
  "structuredContent": {
    "status": "saved_draft",
    "session_id": "T1-01",
    "saved_path": "research/sessions/T1-01-primary-datastore-selection.md",
    "status_downgraded": true,
    "validation": {
      "valid": false,
      "issues": [
        {
          "check_id": "SESSION-08.evidence_density",
          "severity": "blocking",
          "location": "body.Sources & Evidence Ledger",
          "message": "Only 2 claims carry Grade A/B evidence; FRAMEWORK.md §4.3 requires ≥3.",
          "fix_hint": "Add one more corroborated Grade A or B citation, or upgrade an existing single-sourced claim."
        },
        {
          "check_id": "SESSION-11.reversal_triggers",
          "severity": "warning",
          "location": "body.Open Questions & Risks",
          "message": "No explicit reversal-trigger condition detected.",
          "fix_hint": "State the measurable condition under which this recommendation should be re-evaluated."
        }
      ],
      "rubric_score": {
        "evidence_density": { "met": false, "count": 2, "required": 3 },
        "options_evaluated": { "met": true, "count": 3, "required": 2 },
        "failure_modes": { "met": true },
        "reversal_triggers": { "met": false },
        "source_diversity": { "met": true },
        "disconfirming_evidence": { "met": true }
      }
    },
    "informs_decisions_notified": ["D-001"]
  },
  "content": [{ "type": "text", "text": "Saved T1-01 as a draft (not final) — one blocking issue: only 2 of the required 3 Grade A/B–backed claims. One non-blocking note on reversal triggers. Fix the evidence-density gap and resubmit to reach status: final." }]
}
```

### 7. `vivechak_record_decision`
**Title:** Record, update, or supersede a decision
**Description:** "Validates and persists an Architecture Decision Record, or — with `artifact_type: \"conflict_resolution\"` — an ACH-matrix conflict resolution. `operation: \"supersede\"` marks the prior decision superseded and links both records. One-Way Door decisions cannot be saved as `status: accepted` without `human_reviewed: true` and evidence whose `verification_method` is `fetched` or `cached` — this is enforced, not just requested (FRAMEWORK.md §5.4, §5.7)."

```json
// input
{
  "type": "object",
  "properties": {
    "project_root": { "type": "string" },
    "artifact_type": { "type": "string", "enum": ["decision", "conflict_resolution"], "default": "decision" },
    "id": { "type": "string", "description": "D-NNN for a decision, or CHK-NN for a conflict resolution." },
    "content_md": { "type": "string" },
    "operation": { "type": "string", "enum": ["create", "update", "supersede"], "default": "create" },
    "supersedes_id": { "type": "string", "description": "Required when operation='supersede'." }
  },
  "required": ["project_root", "id", "content_md"]
}
```
```json
// output
{
  "type": "object",
  "properties": {
    "status": { "type": "string", "enum": ["saved", "rejected"] },
    "id": { "type": "string" },
    "saved_path": { "type": ["string", "null"] },
    "validation": { "$ref": "#/definitions/ValidationResult" },
    "superseded": { "type": ["object", "null"], "properties": { "id": {"type": "string"}, "new_status": {"type": "string"} } },
    "human_review_required": { "type": "boolean" }
  },
  "required": ["status", "id", "validation", "human_review_required"]
}
```
Annotations: `readOnlyHint: false`, `destructiveHint: true`, `idempotentHint: false`, `openWorldHint: false`.

### 8. `vivechak_validate`
**Title:** Validate an artifact without saving
**Description:** "Dry-run structural check for any of session, decision, conflict_resolution, pipeline, synthesis, or gate content — never writes to disk. Use before you're ready to commit a draft, or for artifact types with no dedicated save tool (the FAD, the gate checklist)."

```json
// input
{
  "type": "object",
  "properties": {
    "project_root": { "type": "string" },
    "artifact_type": { "type": "string", "enum": ["session", "decision", "conflict_resolution", "pipeline", "synthesis", "gate"] },
    "content_md": { "type": "string" },
    "context_id": { "type": "string", "description": "Optional session_id or D-NNN to check referential integrity against the existing workspace." }
  },
  "required": ["project_root", "artifact_type", "content_md"]
}
```
```json
// output
{
  "type": "object",
  "properties": {
    "artifact_type": { "type": "string" },
    "validation": { "$ref": "#/definitions/ValidationResult" }
  },
  "required": ["artifact_type", "validation"]
}
```
Annotations: `readOnlyHint: true`, `destructiveHint: false`, `idempotentHint: true`, `openWorldHint: false`.

### 9. `vivechak_synthesize`
**Title:** Prepare FAD synthesis context
**Description:** "Performs the Filter, Group, and Map steps of the Map-Reduce synthesis workflow (FRAMEWORK.md §6.3): gathers every `status: final` session, clusters by topic, and extracts each one's Key Findings and Recommendation. Does not write the FAD — the calling agent performs Reduce/Reconcile/Trace and authors `FAD.md`, then should validate it with `vivechak_validate` before calling `vivechak_run_gate`."

```json
// input
{
  "type": "object",
  "properties": {
    "project_root": { "type": "string" },
    "scope": { "type": "string", "enum": ["all_final", "by_topic"], "default": "all_final" },
    "topic_filter": { "type": "string" }
  },
  "required": ["project_root"]
}
```
```json
// output
{
  "type": "object",
  "properties": {
    "sessions_ingested": { "type": "array", "items": { "type": "object", "properties": {
      "id": {"type": "string"}, "title": {"type": "string"}, "topic": {"type": "string"},
      "key_findings": {"type": "array", "items": {"type": "string"}}, "recommendation": {"type": "string"} } } },
    "decisions_locked": { "type": "array", "items": { "type": "object", "properties": {
      "id": {"type": "string"}, "title": {"type": "string"}, "chosen_option": {"type": "string"},
      "door_type": {"type": "string"}, "confidence": {"type": "string"}, "evidence_refs": {"type": "array", "items": {"type": "string"}} } } },
    "topic_clusters": { "type": "array", "items": { "type": "object", "properties": { "topic": {"type": "string"}, "session_ids": {"type": "array", "items": {"type": "string"}} } } },
    "open_questions_aggregated": { "type": "array", "items": { "type": "string" } },
    "unresolved_contradictions": { "type": "array", "items": { "type": "object", "properties": { "sessions": {"type": "array", "items": {"type": "string"}}, "summary": {"type": "string"} } } },
    "gate_precheck": { "type": "object", "properties": { "track_b_likely_ready": {"type": "boolean"}, "missing": {"type": "array", "items": {"type": "string"}} } },
    "not_yet_final": { "type": "array", "items": { "type": "object", "properties": { "id": {"type": "string"}, "status": {"type": "string"} } } }
  },
  "required": ["sessions_ingested", "decisions_locked", "gate_precheck"]
}
```
Annotations: `readOnlyHint: true`, `destructiveHint: false`, `idempotentHint: true`, `openWorldHint: false`.

### 10. `vivechak_run_gate`
**Title:** Execute Phase 0 Gate
**Description:** "Runs Track A (Two-Way Door fast-track) and Track B (One-Way Door 9-step) checks against current workspace state. Provide `fad_content_md`, `premortem_summary`, and `reviewer_name` to attempt a full pass; omitting them returns a pre-check of what's currently missing. On a full Track B pass, the FAD is sealed and persisted (immutable from that point — FRAMEWORK.md §6.3.1). Note: this tool can confirm a premortem and reviewer sign-off are *present*, not that they are substantively rigorous — that judgment stays with the human architect."

```json
// input
{
  "type": "object",
  "properties": {
    "project_root": { "type": "string" },
    "fad_content_md": { "type": "string", "description": "Draft FAD content. Required to attempt B9/full seal." },
    "premortem_summary": { "type": "string", "description": "Top-3 failure scenarios + mitigations. Required to pass B7 if any One-Way Door decision exists." },
    "reviewer_name": { "type": "string", "description": "Principal Architect name. Required to pass B8." }
  },
  "required": ["project_root"]
}
```
```json
// output
{
  "type": "object",
  "properties": {
    "track_a": { "type": "object", "properties": { "applicable": {"type": "boolean"}, "result": {"type": "string", "enum": ["pass","fail","n/a"]}, "checks": {"type": "array"} } },
    "track_b": { "type": "object", "properties": { "applicable": {"type": "boolean"}, "result": {"type": "string", "enum": ["pass","fail","n/a"]}, "checks": {"type": "array", "items": {"type": "object", "properties": {
      "id": {"type": "string"}, "label": {"type": "string"}, "pass": {"type": "boolean"}, "detail": {"type": "string"} } } } } },
    "overall": { "type": "string", "enum": ["pass", "fail"] },
    "blocking_issues": { "type": "array", "items": { "type": "string" } },
    "sealed_fad_path": { "type": ["string", "null"] }
  },
  "required": ["overall", "blocking_issues"]
}
```
Annotations: `readOnlyHint: false`, `destructiveHint: true`, `idempotentHint: false`, `openWorldHint: false`.

---

## Validation Strategy

This section answers the SCOPE's four questions directly: what is checked, how strictly, and why.

### What `vivechak_save_session` checks

**Blocking** (file is written as `status: draft`, never `final`, until resolved):

| Check ID | Checks | Source |
|---|---|---|
| SESSION-01 | Frontmatter parses as valid YAML | — |
| SESSION-02 | All 8 required keys present: `id, title, date, status, topic, tags, informs_decisions, confidence` | GENERATOR.md DELIVERABLE §2 |
| SESSION-03 | `id` matches the `session_id` parameter and an entry already registered by `vivechak_save_pipeline` | referential integrity |
| SESSION-04 | `status` ∈ valid enum; `confidence` ∈ {high, medium, low} | GENERATOR.md, FRAMEWORK.md §5.6 |
| SESSION-05 | All 7 required `##` sections present, in order (Research Question, Key Findings, Recommendation, Alternatives Considered, Detailed Findings, Open Questions & Risks, Sources & Evidence Ledger) — **presence only, never content or length** | FRAMEWORK.md §6.2 |
| SESSION-06 | Sources & Evidence Ledger contains ≥1 citation matching the composite format, with a `Grade` in A–E and a `verification_method` in the valid enum | FRAMEWORK.md §5.5 |
| SESSION-08 | Evidence density floor: ≥3 distinct claims carry Grade A or B evidence | FRAMEWORK.md §4.3 |
| SESSION-09 | `informs_decisions`, if non-empty, resolves to decision IDs that exist in the registry | referential integrity |

**Warning** (never blocks; surfaced in the response and left for the agent, synthesis step, or human to weigh):

| Check ID | Checks | Source |
|---|---|---|
| SESSION-10 | <2 options evaluated where the linked decision's `door_type` is one-way | FRAMEWORK.md §4.3 |
| SESSION-11 | No explicit reversal-trigger language detected | FRAMEWORK.md §4.3 |
| SESSION-12 | All evidence traceable to a single source/vendor | FRAMEWORK.md §4.3 |
| SESSION-13 | No disconfirming-evidence language detected | FRAMEWORK.md §4.3 |
| SESSION-14 | `verification_method: recalled` present on evidence feeding a One-Way Door decision (early warning; the hard block is at gate time, see below) | FRAMEWORK.md §5.4 |
| SESSION-15 | `stale` recency modifier used without surrounding context | FRAMEWORK.md §5.2 |
| SESSION-16 | `topic` or `tags` empty (hurts Map-Reduce clustering) | FRAMEWORK.md §6.3 |

**Why evidence density (SESSION-08) is the one rubric item promoted to blocking, and the rest are not**: it is the only rubric criterion that is (a) mechanically countable without judgment, (b) applicable to every session regardless of type (landscape scan, comparison, or blueprint), and (c) a direct expression of P3, Evidentiary Grounding — the framework's most load-bearing principle. "Options evaluated ≥2" only applies to comparison sessions and would false-positive on legitimate landscape sessions if made blocking; "failure modes" and "disconfirming evidence" require exactly the kind of holistic content judgment FRAMEWORK.md's own meta-research found format-based enforcement damages (Tam et al., EMNLP 2024, cited in FRAMEWORK.md §4.1 "What Was Removed and Why"). Blocking on those would risk reintroducing the rigid-output-skeleton failure mode the framework explicitly refuted.

### What `vivechak_record_decision` checks (decision artifacts)

**Blocking**: all 18 frontmatter fields present (FRAMEWORK.md §5.7); all 5 required body sections present; **if `door_type: one-way` and `status: accepted`, then `human_reviewed` must be `true`** (FRAMEWORK.md §5.7, explicit hard rule); **if `door_type: one-way`, every referenced `evidence_refs` citation's `verification_method` must be `fetched` or `cached`, never `recalled`** (FRAMEWORK.md §5.4, hard rule 2); `review_trigger` non-empty for any `status: accepted` record.

**Warning**: `confidence: high` claimed on single-sourced evidence; `prediction` field left null (optional but recommended for the calibration loop FRAMEWORK.md §5.7 and CHANGELOG.md v1.1.0 both added specifically to close).

### Strictness model

Validation strictness is **layered by lifecycle stage**, not uniform:

1. **Session save** enforces session-level completeness and the evidence-density floor only — it cannot know whether a claim will eventually underpin a One-Way Door, only warn.
2. **Decision record** enforces that specific decision's evidentiary bar — this is where the `recalled`-evidence hard rule actually blocks, because only here does `door_type` definitively apply.
3. **Gate** enforces project-wide readiness across every session and decision at once (DAG closure, zero unresolved `contested` flags, 100% `fetched`/`cached` verification on citations underpinning irreversible pillars, every locked ADR carrying a `review_trigger`).

This layering is itself the answer to "how should the server enforce methodology compliance without being annoying": each tool blocks exactly the thing only it can correctly judge, at exactly the moment that judgment becomes meaningful — never earlier (which would false-positive) and never only at the very end (which would make failures expensive to trace back).

### Error handling

Two tiers, matching the MCP specification directly:

- **Protocol errors** (unknown tool, malformed JSON-RPC, schema-invalid arguments) surface as standard JSON-RPC errors — handled by the SDK, not a Vivechak-specific design surface.
- **Tool execution errors** (validation failures, referential-integrity failures, filesystem errors) return `isError: true` with the full `ValidationResult` structure in `structuredContent` — never a bare string or opaque code. Every issue carries a `fix_hint`. Referential errors ("no session `T9-99` registered") use a distinct `check_id` namespace from content-validation errors, so an agent isn't left guessing whether it typed the wrong ID or wrote non-conformant content. A call to `vivechak_save_session` before `vivechak_init` has run returns a distinct, plainly-worded error ("workspace not initialized — call vivechak_init first") rather than surfacing as a generic validation failure.

---

## Open Questions & Risks

- **The session-level `status` enum's intermediate values are not spelled out in the source spec.** FRAMEWORK.md and GENERATOR.md reference `status: final` as the terminal, gate-eligible value, but never enumerate the full set the way they do for decisions (`proposed | accepted | rejected | deprecated | superseded`) or the FAD (`draft | sealed`). This document assumes `draft | final | deprecated` for the MCP layer's validation; this should be confirmed against — or added to — FRAMEWORK.md §6 before implementation, not silently decided by the tool layer alone.
- **Should Conflict Resolution get its own tool after all?** Folding it into `vivechak_record_decision` (Recommendation, refinement 1) keeps the count at 10 but is a genuine trade-off, not a clean win — revisit once real usage data shows how often conflict-resolution is actually invoked relative to decisions.
- **ROADMAP.md's own backlog item ES-03** ("Missing 'tool-generated' verification method... Add when Engine exists") is now directly relevant: now that an MCP layer exists in design form, should `verification_method` gain a sixth value for evidence produced by the calling agent's own tool use (e.g., running a benchmark) rather than fetched from the web? This is a change to the evidence schema itself, not just the MCP layer, and should go through Vivechak's own Change Propagation Map (CONTRIBUTING.md) rather than being decided unilaterally here.
- **ROADMAP.md's backlog item DA-03**: the dependency-injection context budget ("3–5 sentences") that populates `vivechak_next_session`'s `injected_context` field is still unquantified in tokens. Needs calibration against real Engine usage, per the same backlog item.
- **Multi-project ergonomics**: every tool requires `project_root` explicitly, which is correct for an agent juggling multiple projects but is mild repetitive friction for a single-project session. A server-side "default project" convenience was considered and deliberately left out of this design to avoid hidden state contradicting Principle 1 (state lives in the workspace) — worth user-testing before ruling out permanently.
- **The tool-contract deprecation window (below) is this document's own judgment call**, not sourced from a precedent specific to a project Vivechak's size — it should be revisited once real breaking-change history exists, consistent with FRAMEWORK.md's own "rubric calibration... should evolve with your team's experience" stance on the complexity-scoring tiers.
- **MCP Prompts adoption is currently thin** (one practitioner analysis found most deployed servers implement Tools only). The dual-expose recommendation for `vivechak_get_generator_prompt` should be revisited if that changes materially.
- **`vivechak_run_gate` can verify presence, not substance**, for the premortem and human-review steps (B7, B8). A minimal-effort premortem currently satisfies the mechanical check. Closing this gap would require either an LLM-as-judge pass inside the tool (a scope and reliability trade-off of its own) or accepting it as a known limit of mechanical gating and leaving the substantive judgment to the named reviewer, which is this document's provisional recommendation.
- **Cost tracking and multi-model triangulation orchestration are intentionally absent** from all 10 tools — ROADMAP.md explicitly scopes these to the v1.0 Engine target, not v0.1–v0.2. Their absence here is a scope decision, not an oversight, and should not be read as a gap in this design.
- **The GitHub MCP server precedent (no per-tool name prefix) was deliberately not followed** (Recommendation, principle 7); if evaluation data later shows client-side namespacing is reliable enough across Vivechak's actual target hosts (Claude Code, Cursor, Antigravity — per AGENTS.md), the prefix could be revisited as a (breaking) v2 naming change. Not recommended pre-emptively.

---

## Sources & Evidence Ledger

| # | Source | Type | Grade | Modifiers | Verification |
|---|---|---|---|---|---|
| 1 | Vivechak repo — `README.md`, `AGENTS.md`, `FRAMEWORK.md`, `GENERATOR.md`, `ROADMAP.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `VERSION`, `templates/*.md` (4), `examples/SAMPLE-PIPELINE.md`, `meta-research/DECISIONS.md`, `meta-research/README.md` | Primary repo (cloned, HEAD) | A | corroborated · fresh · direct | fetched |
| 2 | Anthropic, "Writing effective tools for AI agents — with agents," Sep 11 2025, anthropic.com/engineering/writing-tools-for-agents | Official vendor engineering blog | A | corroborated · aging · direct | fetched |
| 3 | Model Context Protocol specification — Tools, Resources, Prompts pages (2025-06-18, 2025-11-25, draft revisions), modelcontextprotocol.io | Official protocol spec | A | corroborated · fresh · direct | fetched |
| 4 | MCP Registry — Versioning Published MCP Servers, modelcontextprotocol.io/registry/versioning | Official registry docs | A | corroborated · fresh · direct | fetched |
| 5 | MCP tool annotations (`readOnlyHint`/`destructiveHint`/`idempotentHint`/`openWorldHint`) — AWS Labs, Outreach dev docs, praisonai docs, StackOne, PolicyLayer | Independent SDK/vendor docs, 5 sources | B | corroborated · fresh · direct | fetched |
| 6 | GitHub official MCP server toolset structure — ToolHive docs, lobehub, pkg.go.dev, claudelog | Independent mirrors of one official server, 4 sources | B | corroborated · fresh · direct | fetched |
| 7 | `mcp-adr-analysis-server` (tosin2013), `AdrMcp` (Atypical-Consulting), `sqlew`, `kurdin` decision-cli | Project self-description (READMEs/listings) | C | single · fresh · direct | fetched |
| 8 | "A Practical Guide to Building MCP Servers in the Code-Mode Era" (independent blog), re: Anthropic's Tool Search Tool benchmark | Secondary blog reporting a primary benchmark | C | single · fresh · indirect | secondhand |
| 9 | Gravitee, "Semantic Versioning Strategies for MCP Servers and Agent Tooling APIs" | Vendor blog | C | corroborated (with #10) · fresh · direct | fetched |
| 10 | codex.danielvaughan.com, "MCP Server Versioning and Backwards Compatibility," incl. SEP-2575 | Independent technical blog | B | single · fresh · direct | fetched |
| 11 | archestra.ai, "MCP tools, resources, prompts: the three primitives explained" | Independent practitioner blog | C | single · fresh · indirect | fetched |
| 12 | Shopware developer docs, "MCP Concepts" | Vendor documentation | B | corroborated · fresh · direct | fetched |
| 13 | MADR 4.0 ecosystem — multiple independently-published project ADR pages (Paxman, pyvista-wasm, opalmedapps, io-docs, harness-forge skill reference) | Independent project docs, 5+ sources | B | corroborated · fresh · direct | fetched |

*Coverage note: this document itself carries 9 distinct Grade A/B-backed claims in Key Findings alone, clearing the ≥3 floor SESSION-08 would enforce on any other Vivechak session — a deliberate, if slightly self-referential, consistency check.*
