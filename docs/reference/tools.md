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

---

## 6. `vivechak_review`
Executes non-blocking evidentiary quality coaching across session contents, detecting grade inflation, missing concerns, and delta confirmation bias.
- **Parameters:**
  - `session_id` (string, required): Session ID under review.
  - `content` (string, required): Session Markdown content to evaluate.

---

## 7. `vivechak_amend`
Amends an existing completed research session while archiving prior versions in `research/.history/`.
- **Parameters:**
  - `session_id` (string, required): Session ID to amend.
  - `content` (string, required): Updated Markdown content.
  - `reason` (string, optional): Justification for amendment.

---

## 8. `vivechak_replan`
Dynamically modifies the project dependency DAG, adding or pruning sessions based on discovered architectural forks.
- **Parameters:**
  - `action` (string, required): `"add_session"` or `"remove_session"`.
  - `session_id` (string, required): Target session identifier.
  - `session_def` (string, optional): Session definition table line when adding.

---

## 9. `vivechak_gate`
Mechanically evaluates Phase 0 Exit Gate criteria prior to writing code.
- **Parameters:**
  - `track` (string, optional): `"all"`, `"track_a"` (Two-Way Doors), or `"track_b"` (One-Way Doors).
  - `project_root` (string, optional): Absolute path to project workspace directory.
