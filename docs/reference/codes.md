# Vivechak Error & Diagnostic Codes Reference

Vivechak uses structured, deterministic codes across its MCP server envelope, validation engine, and Phase 0 Gate.

## 1. Operational Error Codes (`E-*`)

| Error Code | Meaning | Remediation |
|---|---|---|
| `E-ROOT-UNRESOLVED` | Workspace directory could not be discovered automatically | Pass `project_root='<absolute_path>'` or set `VCK_PROJECT_ROOT` |
| `E-INVALID-SCOPE` | Unknown research scope provided | Specify `"comparison"`, `"decision"`, or `"project"` |
| `E-INVALID-ARGUMENT` | Missing or malformed required argument | Check tool schema requirements |
| `E-SESSION-NOT-FOUND` | Target session ID does not exist in DAG or workspace | Verify session ID with `vivechak_status` |
| `E-TTY-REQUIRED` | Interactive TTY required for decision signoff | Run in an interactive terminal or provide `VCK_APPROVER_SECRET` (min 16 chars) in CI |
| `E-INTERNAL` | Internal handler panic caught cleanly without crashing stdio | Inspect stderr logs and report defect |
| `E-CONFLICT` | Conflicting session IDs or concurrent edit detected | Review revision history in `research/.history/` |
| `E-LOCKED` | Target workspace file is locked by a concurrent subagent | Wait for lock release or inspect `research/.locks/` |
| `E-FILE-EXISTS` | File already exists and overwrite flag was false | Use amend tool or confirm overwrite |

---

## 2. Evidentiary Quality Coaching Codes (`Q-*`)

Surfaced non-blockingly during `vivechak_save(kind="session")` and `vivechak_review`:

| Coaching Code | Issue Detected | Empirical Rationale |
|---|---|---|
| `Q-GRADE-INFLATION` | Session contains >60% Grade A claims without URL verification | Prevents rubber-stamping AI recall as authoritative documentation |
| `Q-CONCERNS-MISSING` | Discovered Concerns section is empty or <50 characters | Prevents hyper-literalism; ensures adjacent risks are documented |
| `Q-DELTA-BIAS` | 100% of prior assumptions were marked as 'Confirmed' | Guards against confirmation bias; real research surfaces surprises |
| `Q-BIAS` | All Delta updates started from 'Unknown' priors | Anti-Goodhart guard: ensures substantive prior beliefs are evaluated |
| `Q-DENSITY` | Substantive session cites fewer than 2 distinct external sources | Improves citation diversity across independent benchmarks and docs |
| `Q-LOW-FINDINGS` | Fewer than 2 key findings documented in session | Session lacks sufficient depth to inform architectural commitments |

---

## 3. Structural Validation Codes (`V-*`)

Evaluated during deliverable persistence:

| Code | Validation Check | Required Structure |
|---|---|---|
| `V-WEP-MISSING` | Missing Weighted Evaluation Protocol table | Comparison artifacts must include criteria, weights, and scored options |
| `V-FM-INVALID` | YAML frontmatter missing or malformed | Deliverables must begin with valid `---` YAML frontmatter |
| `V-REC-UNSUPPORTED` | Recommendation not grounded in findings | The chosen option must cite findings documented in the session |
| `V-GRADE-INVALID` | Evidence grade not in A–E range | All claims must carry recognized grade `[A-E]` |
| `V-PROVENANCE-INVALID` | Unrecognized provenance tag | Verification must be `fetched`, `cached`, `recalled`, or `human` |
