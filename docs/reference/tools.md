# MCP Tools Reference — Vivechak v0.1.0

> Generated from `internal/mcp` tool registrations (`go run ./tools/docgen`). Do not edit by hand.

Vivechak exposes 9 Model Context Protocol (MCP) tools:

| # | Tool | Required Parameters | Description |
|---|---|---|---|
| 1 | `vivechak_start` | `scope`, `question` | Start technical research (comparison, decision, or project). Returns a research brief to run with web search and an ID. Not for coding. |
| 2 | `vivechak_save` | `kind`, `content` | Validate and persist research under research/ (comparison, plan, session, or fad). Needs revise=true to overwrite; invalid saves as draft. |
| 3 | `vivechak_record_decision` | `id` | Create or update an ADR or conflict resolution and compile DECISIONS.md. Supports auto_draft_from, auto_save, status, and supersedes. |
| 4 | `vivechak_next` | *(none)* | Return the next ready session prompt from the plan DAG with upstream findings injected as data. |
| 5 | `vivechak_status` | *(none)* | Show workspace status, sessions, decisions, and progress (summary, table, or mermaid). Read-only. |
| 6 | `vivechak_review` | `mode` | Check research: validate (content or id), evidence (citation audit), red_team (writes challenge record), or consistency. |
| 7 | `vivechak_amend` | `id`, `note` | Append a dated amendment to a saved session, ADR, FAD, or comparison and flag stale dependents. |
| 8 | `vivechak_replan` | `plan`, `op` | Edit a plan DAG in place (add_session, remove_session, update_deps, update_prompt); protects completed sessions. |
| 9 | `vivechak_gate` | *(none)* | Evaluate Phase 0 exit criteria (G1-G12) for the workspace or one ADR and write PHASE-0-GATE.md (one-way needs vck signoff). |

---

## 1. `vivechak_start`

Start technical research (comparison, decision, or project). Returns a research brief to run with web search and an ID. Not for coding.

- **Annotations:** `readOnlyHint=false`, `idempotentHint=false`, `destructiveHint=true`, `openWorldHint=false`
- **Parameters:**
  - `scope` (string, required): Research scope: comparison (2-5 options), decision (ADR), or project (pipeline -> FAD). Allowed: `"comparison"`, `"decision"`, `"project"`.
  - `question` (string, required): Technical dilemma, options to compare, or project vision.
  - `options` (string[], optional): Named options for comparison scope.
  - `criteria` (string, optional): Optional evaluation criteria or constraints.
  - `part` (integer, optional): 1-based part index for multi-part project briefs.
  - `project_root` (string, optional): Absolute path to the target project directory.

---

## 2. `vivechak_save`

Validate and persist research under research/ (comparison, plan, session, or fad). Needs revise=true to overwrite; invalid saves as draft.

- **Annotations:** `readOnlyHint=false`, `idempotentHint=false`, `destructiveHint=true`, `openWorldHint=false`
- **Parameters:**
  - `kind` (string, required): Artifact kind: plan, session, comparison, or fad. Allowed: `"plan"`, `"session"`, `"comparison"`, `"fad"`.
  - `content` (string, required): Markdown content to validate and save.
  - `id` (string, optional): Artifact or session ID (allocated when omitted for comparison/plan).
  - `scope` (string, optional): Optional scope hint when saving a plan. Allowed: `"comparison"`, `"decision"`, `"project"`.
  - `slug` (string, optional): Optional URL-safe filename slug.
  - `revise` (boolean, optional): Update an existing artifact, archiving the old version.
  - `decisions_content` (string, optional): Initial ADR stub or registry Markdown when saving a plan.
  - `project_root` (string, optional): Absolute path to the target project directory.

---

## 3. `vivechak_record_decision`

Create or update an ADR or conflict resolution and compile DECISIONS.md. Supports auto_draft_from, auto_save, status, and supersedes.

- **Annotations:** `readOnlyHint=false`, `idempotentHint=false`, `destructiveHint=true`, `openWorldHint=false`
- **Parameters:**
  - `id` (string, required): Decision ID (e.g. D-001, or D-NEW to allocate the next free ID).
  - `content` (string, optional): Decision or conflict-resolution Markdown (omit when using auto_draft_from).
  - `auto_draft_from` (string, optional): Session ID (e.g. D-001-S1) to draft the ADR from.
  - `auto_save` (boolean, optional): Save the drafted ADR immediately when using auto_draft_from.
  - `status` (string, optional): Lifecycle status written into frontmatter before validation. Allowed: `"proposed"`, `"accepted"`, `"rejected"`, `"superseded"`, `"deprecated"`.
  - `slug` (string, optional): Optional URL-safe filename slug.
  - `supersedes` (string, optional): Existing decision ID superseded by this ADR (e.g. D-001).
  - `revise` (boolean, optional): Overwrite an existing ADR with history archiving.
  - `artifact_type` (string, optional): Artifact type: decision (default) or conflict-resolution. Allowed: `"decision"`, `"conflict-resolution"`.
  - `project_root` (string, optional): Absolute path to the target project directory.

---

## 4. `vivechak_next`

Return the next ready session prompt from the plan DAG with upstream findings injected as data.

- **Annotations:** `readOnlyHint=true`, `idempotentHint=true`, `destructiveHint=false`, `openWorldHint=false`
- **Parameters:**
  - `session_id` (string, optional): Specific session ID to retrieve (default: next ready session).
  - `verbose` (boolean, optional): Return full prompt without 10K-token truncation.
  - `project_root` (string, optional): Absolute path to the target project directory.

---

## 5. `vivechak_status`

Show workspace status, sessions, decisions, and progress (summary, table, or mermaid). Read-only.

- **Annotations:** `readOnlyHint=true`, `idempotentHint=true`, `destructiveHint=false`, `openWorldHint=false`
- **Parameters:**
  - `format` (string, optional): Output format: summary (default), table, or mermaid. Allowed: `"summary"`, `"table"`, `"mermaid"`.
  - `project_root` (string, optional): Absolute path to the target project directory.

---

## 6. `vivechak_review`

Check research: validate (content or id), evidence (citation audit), red_team (writes challenge record), or consistency.

- **Annotations:** `readOnlyHint=false`, `idempotentHint=false`, `destructiveHint=false`, `openWorldHint=false`
- **Parameters:**
  - `mode` (string, required): Review mode: validate, evidence, red_team, or consistency. Allowed: `"validate"`, `"evidence"`, `"red_team"`, `"consistency"`.
  - `id` (string, optional): Target session, decision, comparison, or plan ID.
  - `content` (string, optional): Inline Markdown content for dry-run validate mode.
  - `artifact_type` (string, optional): Artifact type when validating inline content. Allowed: `"session"`, `"decision"`, `"comparison"`, `"plan"`, `"fad"`, `"conflict-resolution"`.
  - `project_root` (string, optional): Absolute path to the target project directory.

---

## 7. `vivechak_amend`

Append a dated amendment to a saved session, ADR, FAD, or comparison and flag stale dependents.

- **Annotations:** `readOnlyHint=false`, `idempotentHint=false`, `destructiveHint=true`, `openWorldHint=false`
- **Parameters:**
  - `id` (string, required): Session, decision, comparison, or synthesis ID to amend.
  - `note` (string, required): Amendment Markdown note to append.
  - `reason` (string, optional): Short summary of why the artifact is being amended.
  - `amending_session_id` (string, optional): Session ID that discovered the amendment.
  - `project_root` (string, optional): Absolute path to the target project directory.

---

## 8. `vivechak_replan`

Edit a plan DAG in place (add_session, remove_session, update_deps, update_prompt); protects completed sessions.

- **Annotations:** `readOnlyHint=false`, `idempotentHint=false`, `destructiveHint=true`, `openWorldHint=false`
- **Parameters:**
  - `plan` (string, required): Target plan: PROJECT or decision ID (e.g. D-001).
  - `op` (string, required): Replan operation to perform. Allowed: `"add_session"`, `"remove_session"`, `"update_deps"`, `"update_prompt"`.
  - `session_id` (string, optional): Target session ID (e.g. T1-02).
  - `reason` (string, optional): Rationale for the plan change.
  - `layer` (string, optional): DAG layer number for add_session.
  - `topic` (string, optional): Session title for add_session.
  - `prompt` (string, optional): Session prompt text for add_session or update_prompt.
  - `dependencies` (string, optional): Comma-separated dependency session IDs for add_session or update_deps.
  - `door_type` (string, optional): Door classification for add_session. Allowed: `"one-way"`, `"two-way"`.
  - `decision_ref` (string, optional): Informed decision ID for add_session.
  - `project_root` (string, optional): Absolute path to the target project directory.

---

## 9. `vivechak_gate`

Evaluate Phase 0 exit criteria (G1-G12) for the workspace or one ADR and write PHASE-0-GATE.md (one-way needs vck signoff).

- **Annotations:** `readOnlyHint=false`, `idempotentHint=false`, `destructiveHint=true`, `openWorldHint=false`
- **Parameters:**
  - `decision` (string, optional): Specific decision ID to evaluate (e.g. D-001).
  - `track` (string, optional): Track filter: all (default), track_a, or track_b. Allowed: `"all"`, `"track_a"`, `"track_b"`.
  - `verbose` (boolean, optional): Include verbose rule diagnostics.
  - `project_root` (string, optional): Absolute path to the target project directory.
