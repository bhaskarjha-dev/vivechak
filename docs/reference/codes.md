# Vivechak Diagnostic & Error Codes Reference

> Generated from `internal/core/codes.go` (`go run ./tools/docgen`). Do not edit by hand.

Vivechak emits 78 structured codes across its MCP envelope, structural/plan validator, warning diagnostics, and quality coach:

## 1. Operational Error Codes (`E-*`)

Returned on failed operations (`ok: false`, `isError: true`) with remediation in `fix`.

| Code | Message | Remediation (`fix`) |
|---|---|---|
| `E-CONFINEMENT` | Target artifact path escapes research/ or the workspace boundary. | Write artifacts only under research/ within the resolved project workspace. |
| `E-HUMAN-ONLY-FIELD` | Agent attempted to set a human-only sign-off field (human_reviewed, reviewed_by, reviewed_at, reviewed_hash). | Remove the sign-off fields and ask a human reviewer to run 'vck signoff <D-id> --reviewer "<name>"' in an interactive terminal. |
| `E-ID-CONFLICT` | Target file already exists on disk with a different artifact identifier. | Use a distinct id or slug for a new artifact, or match the existing file's id with revise=true. |
| `E-ID-INVALID` | Artifact or session identifier is malformed or does not match the artifact kind. | Use D-NNN for decisions, C-NNN for comparisons, or a valid session ID such as T1-01 (or D-NEW to allocate). |
| `E-INTERNAL` | Unexpected internal error while handling request. | Inspect server stderr diagnostics and retry the operation. |
| `E-INVALID-ARGUMENT` | Tool argument is missing, has the wrong type, or is outside its allowed values. | Pass arguments matching the tool's inputSchema in tools/list. |
| `E-IO` | Filesystem read, directory creation, or atomic write failed in the workspace. | Verify that the workspace directory exists and is writable, then retry. |
| `E-LOCK-TIMEOUT` | Timed out waiting to acquire the advisory file lock for an artifact. | Wait for concurrent writes to finish and retry the call. |
| `E-NOT-FOUND` | Requested plan, session, comparison, or decision file does not exist in research/. | Save the prerequisite artifact first or inspect available IDs with vivechak_status. |
| `E-NOT-INITIALIZED` | Workspace does not contain an initialized research/ directory yet. | Call vivechak_start with scope and question to initialize the workspace first. |
| `E-PLACEHOLDER-REMAINS` | Unexpanded template placeholders remain in an artifact that requires concrete content. | Replace all bracketed template placeholders with grounded findings before saving as complete or accepted. |
| `E-REVISE-REQUIRED` | Target artifact already exists with different content and revise=true was not set. | Pass revise=true to update that same artifact (archiving the old copy in research/.history/), or omit id for a new artifact. |
| `E-ROOT-AMBIGUOUS` | Client advertised multiple MCP workspace roots and no explicit project_root was given. | Pass project_root explicitly with the target project's absolute directory path. |
| `E-ROOT-INVALID` | Explicit or configured workspace root path does not exist or is not a directory. | Pass an existing project directory path in project_root. |
| `E-ROOT-SYSTEM` | Candidate directory is a system, home, or application directory rather than a project. | Pass project_root pointing to your project repository. |
| `E-ROOT-UNRESOLVED` | Could not resolve a safe project workspace root from arguments, client roots, environment, or cwd. | Pass project_root with the absolute path of the user's project directory. |
| `E-ROOT-UNSAFE` | Workspace path is empty or points to a restricted filesystem root, home, temp, or OS directory. | Choose a dedicated project directory inside your workspace. |
| `E-SESSION-COMPLETED` | Attempted to remove or alter the plan definition of an already completed research session. | Keep completed sessions intact and append post-hoc updates with vivechak_amend. |
| `E-TTY-REQUIRED` | vck signoff was invoked without an interactive terminal session. | Ask a human reviewer to run 'vck signoff <D-id> --reviewer "<name>"' directly in an interactive terminal. |

---

## 2. Structural & Plan Validation Codes (`V-*`)

Emitted during artifact or plan validation (`vivechak_save`, `vivechak_record_decision`, `vivechak_review`).

| Code | Message | Remediation (`fix`) |
|---|---|---|
| `V-CMP-NO-MATRIX` | Comparison artifact lacks a Weighted Evaluation Protocol (WEP) Markdown table. | Include a criterion-by-option Markdown table with weights and 1-5 scores. |
| `V-CMP-OPTIONS` | Comparison matrix evaluates fewer than two distinct options. | Compare between 2 and 5 named options as columns in the WEP matrix. |
| `V-CMP-SCORES` | Comparison matrix cells lack valid numeric 1-5 option scores. | Score every option on every criterion from 1 to 5 backed by graded citations. |
| `V-CMP-WEIGHTS` | Comparison criterion weights are missing or do not normalize to 1.0 (or 100%). | Assign positive criterion weights that sum to 1.0 (e.g. 0.40, 0.35, 0.25). |
| `V-EMPTY` | Artifact content is empty. | Provide complete Markdown content matching the artifact's template. |
| `V-EMPTY-BODY` | Artifact contains YAML frontmatter but no Markdown body. | Write the research findings, trade-offs, and evidence ledger below the frontmatter fence. |
| `V-INVALID-FRONTMATTER` | YAML frontmatter block is syntactically malformed. | Fix YAML syntax errors between the opening and closing --- fences. |
| `V-MISSING-FIELD` | Required YAML frontmatter field is missing. | Add the missing required key to the YAML frontmatter block. |
| `V-MISSING-FRONTMATTER` | Artifact is missing its opening YAML frontmatter block. | Begin the document with a --- YAML frontmatter block containing required metadata. |
| `V-MISSING-SECTION` | Required Markdown section heading is missing from the artifact. | Add all required Markdown sections for the scope template. |
| `V-MISSING-STATUS` | Session frontmatter does not declare status. | Set status: complete (or draft) in the session's YAML frontmatter. |
| `V-NO-GRADED-CLAIM` | Artifact contains no inline evidence grades (A-E). | Grade factual claims A-E with modifiers, verification labels, and source URLs. |
| `V-OWD-NO-CONCERNS` | One-Way Door session lacks a substantive Discovered Concerns section. | Add a ## Discovered Concerns section (>=50 chars) detailing failure modes and disconfirming evidence. |
| `V-OWD-VAGUE-PROVENANCE` | One-Way Door Grade A/B claim lacks a concrete URL or primary standard anchor. | Attach a canonical https:// URL to every Grade A/B claim informing a One-Way Door. |
| `V-PLACEHOLDER` | Unexpanded template placeholder remains in the comparison output. | Replace all template placeholders with concrete option analysis and scores. |
| `V-PLAN-CYCLE` | Plan session dependency graph contains a cycle or depends on the synthesis session. | Remove circular depends entries so sessions form a DAG ending at kind: synthesis. |
| `V-PLAN-DUPLICATE-PROMPT` | Plan contains multiple ### <id>: prompt sections for the same session. | Keep exactly one ### <id>: <title> section with one prompt fence per session. |
| `V-PLAN-INVALID-BLOCK` | Plan's yaml vivechak-plan block is invalid, duplicated, or uses unknown keys. | Keep one valid yaml vivechak-plan block using plan, sessions, and decisions keys. |
| `V-PLAN-MISSING-BLOCK` | Plan lacks the required fenced yaml vivechak-plan block. | Add one ```yaml vivechak-plan block declaring plan and sessions. |
| `V-PLAN-MISSING-PROMPT` | Session declared in vivechak-plan has no prompt fence under ## Sessions. | Add ### <id>: <title> followed by one ```prompt fence under ## Sessions. |
| `V-PLAN-NO-SESSIONS` | Plan declares zero research sessions. | Declare at least one session in the vivechak-plan block's sessions list. |
| `V-PLAN-SYNTHESIS` | Project pipeline does not have exactly one kind: synthesis session. | Set kind: synthesis on the single terminal synthesis session in the project plan. |
| `V-PLAN-UNKNOWN-DEPENDENCY` | Session depends on a session ID not declared in the plan. | List only session IDs declared in the same vivechak-plan block (or [ALL] on synthesis). |
| `V-PLAN-UNLISTED-SESSION` | Prompt section under ## Sessions is not declared in the vivechak-plan block. | Add the session to the vivechak-plan sessions list or remove the extra section. |
| `V-TOO-SHORT` | Artifact body is shorter than the minimum substantive length. | Expand the evidence-graded analysis and trade-off discussion before saving. |

---

## 3. Operational & Advisory Warning Codes (`W-*`)

Returned in `warnings` (`[{code, message, fix}]`) without failing the request.

| Code | Message | Remediation (`fix`) |
|---|---|---|
| `W-CMP-GATE-NA` | Comparison scope has no Phase 0 architecture gate requirements. | Proceed to implementation, or start decision scope if locking an ADR. |
| `W-DECISIONS-COMPILE` | Individual ADR was saved, but recompiling research/DECISIONS.md failed. | Check write permissions on research/DECISIONS.md and re-run vivechak_record_decision. |
| `W-GATE-RULE-ISSUE` | Phase 0 Gate surfaced a rule failure or structural check issue. | Inspect data.rules and resolve each failing check before coding. |
| `W-GATE-WRITE` | Phase 0 Gate ran, but persisting research/PHASE-0-GATE.md failed. | Check write permissions under research/ and re-run vivechak_gate. |
| `W-ID-ALLOCATED` | Omitted or placeholder identifier was resolved to a server-allocated ID. | Use the allocated ID returned in data.id on subsequent tool calls. |
| `W-INJECT-FAILED` | Upstream session context could not be injected; serving raw session prompt. | Verify upstream session files under research/sessions/ are readable. |
| `W-INJECTION-SIZE` | Injected upstream findings exceed the recommended context budget. | Keep upstream Key Findings and Recommendation sections concise. |
| `W-INVALID-DATE-FORMAT` | Frontmatter date is not in YYYY-MM-DD format. | Format frontmatter date as ISO YYYY-MM-DD. |
| `W-LAYER-TRANSITION` | All sessions in the prior DAG layer are complete; advancing to the next layer. | Check whether findings warrant vivechak_replan before running downstream sessions. |
| `W-MARKER-CONFLICT` | ADR body contains nested HTML decision registry markers. | Remove manual <!-- DECISION: ... --> markers from individual ADR files. |
| `W-MISSING-DATE` | Frontmatter omits the date field. | Add date: YYYY-MM-DD to the YAML frontmatter. |
| `W-MISSING-DOOR-TYPE` | Artifact omits door_type (one-way or two-way). | Set door_type: one-way or door_type: two-way in YAML frontmatter. |
| `W-MISSING-FIELD` | Recommended YAML frontmatter field is missing. | Add the recommended field to the YAML frontmatter. |
| `W-MISSING-FRONTMATTER` | Artifact omits optional YAML frontmatter. | Add YAML frontmatter with id, title, date, and status. |
| `W-NO-CONCERNS` | Two-Way Door session omits a Discovered Concerns section. | Add ## Discovered Concerns noting caveats or operational risks. |
| `W-NO-EVIDENCE-GRADES` | Artifact body contains no inline evidence grade markers. | Tag factual claims with grades A-E (see vivechak://methodology/grades). |
| `W-OUTPUT-SIZE` | Session prompt exceeded 10K estimated tokens and was truncated. | Call vivechak_next with verbose=true to fetch the full prompt. |
| `W-OWD-NO-CHALLENGE` | One-Way Door ADR lacks a current adversarial red-team challenge record. | Run vivechak_review(mode="red_team", id="<D-id>") and cite the returned record_id. |
| `W-PLACEHOLDER-REMAINS` | Proposed draft ADR still contains unexpanded template placeholders. | Replace every placeholder before marking status: accepted. |
| `W-PRE-SYNTHESIS-GAP` | Completed One-Way Door session has evidentiary or challenge gaps before synthesis. | Amend weak One-Way Door sessions via vivechak_amend before running vivechak_gate. |
| `W-RECALLED-GRADE-CAP` | Claim labelled recalled or parametric is graded higher than Grade D. | Cap model-recall claims at Grade D or verify with web search and cite the source URL. |
| `W-SHORT-BODY` | Artifact Markdown body is unusually short. | Expand empirical findings, trade-offs, and citations if incomplete. |
| `W-SHORT-DECISION` | Decision record body is unusually short. | Document context, decision outcome, rejected alternatives, and reversal triggers. |
| `W-STALE-DECISION-REFERENCE` | Saved session references an ADR that has been superseded. | Review the session and append an update with vivechak_amend if needed. |
| `W-STALE-DOWNSTREAM` | Completed downstream sessions depend on an artifact that was amended. | Review listed downstream sessions and amend them if conclusions changed. |
| `W-UNKNOWN-ARG` | Tool call included unknown argument name(s), which were ignored. | Pass only parameter names declared in the tool's inputSchema. |
| `W-VAGUE-PROVENANCE` | Grade A/B claim lacks an explicit URL or standard identifier. | Include a canonical https:// URL on the claim line or ledger row. |

---

## 4. Cognitive Quality Coaching Codes (`Q-*`)

Surfaced non-blockingly on session saves and reviews to highlight evidentiary gaps.

| Code | Message | Remediation (`fix`) |
|---|---|---|
| `Q-BIAS` | Session Delta lists no contradicted or updated beliefs, suggesting possible confirmation bias. | Actively test disconfirming evidence and record how initial assumptions changed in ## Delta. |
| `Q-CONCERNS` | Session omits or under-develops the ## Discovered Concerns section. | Document concrete failure modes, operational caveats, and limitations in ## Discovered Concerns. |
| `Q-DEPTH` | Session contains fewer than the recommended 4-7 graded key findings. | Expand ## Key Findings with additional empirically graded claims where warranted. |
| `Q-EVIDENCE` | Session evidence count, URL anchoring, or grade mix warrants review. | Anchor Grade A claims with primary URLs and cite multiple independent sources. |
| `Q-FRESHNESS` | Session cites potentially stale year references (>2 years old). | Verify version-sensitive claims against current upstream documentation. |
| `Q-RIGOR` | Session recommends an option without documenting evaluated alternatives. | Add an ## Alternatives Considered section weighing rejected candidates and trade-offs. |
| `Q-STRUCTURE` | Session omits ## Prior or ## Delta belief-tracking sections. | Include both ## Prior and ## Delta to show how evidence updated initial assumptions. |
