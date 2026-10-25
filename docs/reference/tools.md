# MCP Tools Reference — Vivechak v0.1.0

Vivechak exposes 9 cohesive Model Context Protocol (MCP) tools for autonomous and pair-programming research workflows:

| # | Tool | Purpose | Primary Scope | Guided Action Returned |
|---|---|---|---|---|
| 1 | `vivechak_start` | Initialize engagement and return direct actionable research brief | All Scopes | Search & populate research deliverable |
| 2 | `vivechak_save` | Atomically persist comparison matrix, pipeline DAG, research session, or FAD | All Scopes | `vivechak_save` or `vivechak_next` |
| 3 | `vivechak_record_decision` | Persist canonical Architectural Decision Record (`D-NNN.md`) and update registry | Decision & Project | `vivechak_next` or `vivechak_gate` |
| 4 | `vivechak_next` | Topologically resolve next unblocked research session in dependency DAG | Project | Search & save session |
| 5 | `vivechak_status` | Inspect workspace state, session counts, progress percentage, and DAG blockers | All Scopes | Next appropriate action |
| 6 | `vivechak_review` | Evaluate session quality and surface real-time non-blocking quality coaching (`Q-*`) | Decision & Project | Address quality observations |
| 7 | `vivechak_amend` | Amend completed research session or record revised findings | All Scopes | Re-evaluate dependent ADRs |
| 8 | `vivechak_replan` | Dynamically modify dependency DAG (insert/prune sessions) based on discoveries | Project | `vivechak_next` |
| 9 | `vivechak_gate` | Mechanically evaluate Track A (reversibility) and Track B (rigor G1–G12) exit criteria | All Scopes | Proceed to implementation or remediate |

---

## 1. `vivechak_start`
Initializes a Vivechak research workspace and generates a directly executable research brief.
- **Parameters:**
  - `scope` (string, required): `"comparison"`, `"decision"`, or `"project"`.
  - `question` (string, required): Technical question, comparison dilemma, or system architecture description.
  - `project_root` (string, optional): Absolute path to project workspace directory.
  - `part` (integer, optional): 1-based part index for multi-part delivery of large briefs (scope="project").
  - `options` (array of strings, optional): Named options to evaluate in comparison scope.
  - `criteria` (string, optional): Key evaluation criteria or constraints.
- **Guided Action:** Returns directly executable research brief with required hypotheses and evidence contracts.

---

## 2. `vivechak_save`
Polymorphically persists research deliverables with automatic atomic writes and validation.
- **Parameters:**
  - `kind` (string, required): `"comparison"`, `"plan"`, `"session"`, or `"fad"`.
  - `id` (string, required): Session ID (`C-01`, `D-001`, `T1-01`, `SYN-01`).
  - `content` (string, required): Complete Markdown artifact content.
  - `project_root` (string, optional): Absolute path to project workspace directory.

---

## 3. `vivechak_record_decision`
Records and locks an Architectural Decision Record (ADR) as canonical `research/D-NNN.md` and compiles consolidated `research/DECISIONS.md`.
- **Parameters:**
  - `decision_id` (string, required): Canonical ADR identifier (`D-001`, `D-002`).
  - `auto_draft_from` (string, optional): Session ID to automatically extract rationale, options, and findings from.
  - `auto_save` (boolean, optional): Atomically persist the draft ADR to disk immediately.
  - `project_root` (string, optional): Absolute path to project workspace directory.

---

## 4. `vivechak_next`
Resolves the next unblocked research session in a project dependency DAG using Kahn's topological sort.
- **Parameters:**
  - `project_root` (string, optional): Absolute path to project workspace directory.
- **Guided Action:** Returns the complete session brief for the highest-priority unblocked session.

---

## 5. `vivechak_status`
Provides workspace health, detected scope, completed sessions count, and current DAG topology state.
- **Parameters:**
  - `project_root` (string, optional): Absolute path to project workspace directory.
  - `format` (string, optional): Output format: `"text"` (default) or `"mermaid"` (Mermaid DAG flowchart diagram).

---

## 6. `vivechak_review`
Inspects research deliverables for quality validation or adversarial stress testing.
- **Parameters:**
  - `mode` (string, optional): Review mode: `"validate"` (dry-run 4-level quality checks on content or workspace) or `"challenge"` (generates adversarial red-team critique prompts for One-Way Doors).
  - `target` (string, optional): Target session ID (e.g. `T1-01`), decision ID (e.g. `D-001`), or artifact path.
  - `artifact_type` (string, optional): What to validate when `mode="validate"`: `"session"`, `"decision"`, `"plan"`, `"comparison"`, or `"fad"`. Omit to validate entire workspace.
  - `content` (string, optional): Markdown content to validate in memory (dry run); omit to inspect saved files.
  - `project_root` (string, optional): Absolute path to project workspace directory.

---

## 7. `vivechak_amend`
Amends an existing completed research session or decision while archiving prior revisions in `research/.history/`.
- **Parameters:**
  - `session_id` (string, required): Session ID (`T1-01`), decision ID (`D-001`), or comparison ID (`C-01`) to amend (alias: `target`).
  - `amendment` (string, required): Amendment Markdown content to append (alias: `content`).
  - `reason` (string, optional): Rationale or summary for the amendment.
  - `amending_session_id` (string, optional): Session ID that discovered the amendment (for traceability).
  - `project_root` (string, optional): Absolute path to project workspace directory.

---

## 8. `vivechak_replan`
Dynamically modifies the project dependency DAG, adding, removing, or updating sessions based on discovered architectural forks. Call with empty arguments for a guided overview.
- **Parameters:**
  - `operation` (string, optional): Replan operation: `"add_session"`, `"remove_session"`, `"update_deps"`, or `"update_prompt"`.
  - `session_id` (string, optional): Target session identifier (e.g. `T1-02`).
  - `reason` (string, optional): Architectural finding or requirement prompting the replan.
  - `impact` (string, optional): Intended effect on DAG topology or affected downstream sessions.
  - `layer` (string, optional): Layer number when adding a session (e.g. `"1"`).
  - `topic` (string, optional): Session title/topic when adding a session.
  - `prompt` (string, optional): Research prompt text when adding a session.
  - `dependencies` (string, optional): Comma-separated dependency session IDs when adding a session.
  - `door_type` (string, optional): Decision reversibility for the session: `"one-way"` or `"two-way"`.
  - `decision_ref` (string, optional): Associated decision ID (e.g. `"D-001"`).
  - `new_prompt` (string, optional): Replacement prompt markdown when `operation="update_prompt"`.
  - `new_dependencies` (string, optional): Replacement comma-separated dependencies when `operation="update_deps"`.
  - `project_root` (string, optional): Absolute path to project workspace directory.

---

## 9. `vivechak_gate`
Mechanically evaluates Phase 0 Exit Gate criteria prior to writing code.
- **Parameters:**
  - `track` (string, optional): Gate evaluation track: `"all"` (default), `"track_a"` (Two-Way Doors), or `"track_b"` (One-Way Doors).
  - `project_root` (string, optional): Absolute path to project workspace directory.
