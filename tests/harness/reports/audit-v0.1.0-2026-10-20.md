# Vivechak Release Audit Report: v0.1.0
* **Release Version:** `v0.1.0`
* **Date:** 2026-10-20
* **Total Runs Evaluated:** 18
* **Pass Rate:** 100.0% (18 / 18 PASS)

---

## 1. Execution Scorecard
| Run File | Scenario | Harness | Model | Scope | Status |
|---|---|---|---|---|---|
| `run-01-agy-gemini-3_8-flash-high.md` | **01** | `agy` | `gemini-3.8-flash-high` | `comparison` | ✅ PASS |
| `run-01-direct_mcp-deterministic.md` | **01** | `direct_mcp` | `deterministic` | `comparison` | ✅ PASS |
| `run-02-agy-gemini-3_8-flash-high.md` | **02** | `agy` | `gemini-3.8-flash-high` | `comparison` | ✅ PASS |
| `run-02-direct_mcp-deterministic.md` | **02** | `direct_mcp` | `deterministic` | `comparison` | ✅ PASS |
| `run-03-direct_mcp-deterministic.md` | **03** | `direct_mcp` | `deterministic` | `project` | ✅ PASS |
| `run-04-agy-gemini-3_8-flash-high.md` | **04** | `agy` | `gemini-3.8-flash-high` | `decision` | ✅ PASS |
| `run-04-direct_mcp-deterministic.md` | **04** | `direct_mcp` | `deterministic` | `decision` | ✅ PASS |
| `run-05-direct_mcp-deterministic.md` | **05** | `direct_mcp` | `deterministic` | `decision` | ✅ PASS |
| `run-06-direct_mcp-deterministic.md` | **06** | `direct_mcp` | `deterministic` | `project` | ✅ PASS |
| `run-07-direct_mcp-deterministic.md` | **07** | `direct_mcp` | `deterministic` | `project` | ✅ PASS |
| `run-08-agy-gemini-3_8-flash-high.md` | **08** | `agy` | `gemini-3.8-flash-high` | `decision` | ✅ PASS |
| `run-08-direct_mcp-deterministic.md` | **08** | `direct_mcp` | `deterministic` | `decision` | ✅ PASS |
| `run-12-agy-gemini-3_8-flash-high.md` | **12** | `agy` | `gemini-3.8-flash-high` | `project` | ✅ PASS |
| `run-12-direct_mcp-deterministic.md` | **12** | `direct_mcp` | `deterministic` | `project` | ✅ PASS |
| `run-13-direct_mcp-deterministic.md` | **13** | `direct_mcp` | `deterministic` | `project` | ✅ PASS |
| `run-14-direct_mcp-deterministic.md` | **14** | `direct_mcp` | `deterministic` | `project` | ✅ PASS |
| `run-16-direct_mcp-deterministic.md` | **16** | `direct_mcp` | `deterministic` | `comparison` | ✅ PASS |
| `run-21-direct_mcp-deterministic.md` | **21** | `direct_mcp` | `deterministic` | `boundary` | ✅ PASS |

---

## 2. Dynamic Turn 4 Meta-Improvements
### DX Friction Items (18)
* **agy (gemini-3.8-flash-high) on S-01:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Based on the end-to-end execution of this comparison session, here are concrete, high-signal suggestions to improve Vivechak’s developer experience, guidance clarity, and artifact ergonomics:

---

### 1. Developer Experience (DX) & Parameter Friction

#### Issue A: `project_root` Resolution Failed on Fresh Workspaces
* **What Happened:** When calling `vivechak_save_plan` without an explicit `project_root`, the tool threw:
  ```
  "could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set, 
   research/ not found walking up from ..., and no prior workspace initialized in this session. 
   Run vivechak_init first or pass project_root explicitly"
  ```
* **The Contradiction:** `instructions.md` explicitly documents:
  * *"Comparison ('A vs B', 1 session): No DAG/gate/init. 5m!"*
  * *"project_root defaults to CWD"*
* **Suggested Fix:** If `scope == "comparison"` (which bypasses `vivechak_init`), the backend should **automatically fall back to `process.cwd()`** and create the `research/` directory without failing or demanding `vivechak_init`.

#### Issue B: Parameter Proliferation (`id` vs. `decision_id` vs. `comparison_id`)
* **What Happened:** The schema for `vivechak_save_plan` defines three separate parameters that represent the exact same entity:
  * `id`: *"universal identifier"*
  * `decision_id`: *"identifier for decision scope or comparison scope"*
  * `comparison_id`: *"optional alias for decision_id in comparison scope"*
* **Suggested Fix:** Unify these into a single primary parameter `id` (e.g., `id: "C-01"`). Handle legacy aliases (`decision_id`, `comparison_id`) silently in the tool's backend adapter without exposing three overlapping fields in the JSON schema.

#### Issue C: Stale `.lock` File Left Behind
* **What Happened:** After `vivechak_save_plan` completed successfully, a 0-byte lock file remained on disk: `research/C-01-cmp-winston-vs-pino.md.lock`.
* **Suggested Fix:** Ensure the file persistence routine releases and removes file locks in a `finally` block so that empty lock files do not pollute the git working tree or cause concurrency conflicts on subsequent tool calls.

---

### 2. Prompt & Guidance Ambiguity

#### Issue A: The "Meta-Prompt" Indirection in `prepare_generator`
* **What Happened:** `vivechak_prepare_generator` returns instructions directing the agent to *generate a research prompt inside a `prompt` code block* with slot replacements (`{slots}`), rather than returning the actionable research brief directly.
* **The Friction:** In an automated agent workflow, this forces the agent into a meta-loop:
  1. Receive instructions on how to write a prompt.
  2. Parse the slot markers.
  3. Execute the research anyway.
* **Suggested Fix:** For automated AI agents, have `vivechak_prepare_generator` return the **already-interpolated research brief** directly. Tell the agent: *"Here is your research brief. Execute searches against these hypotheses and format the output according to the schema below."*

#### Issue B: Ambiguity Around `vivechak_validate` vs. `vivechak_save_plan`
* **What Happened:** `instructions.md` states: *"save_session auto-validates; vivechak_validate for dry-runs"*. It did not clarify whether `save_plan` also auto-validates. Agents worry that saving without validating will trigger a fatal schema rejection, leading them to call `vivechak_validate` first (doubling token usage and execution latency).
* **Suggested Fix:** Clarify in `instructions.md`:
  > *"Both `save_session` and `save_plan` auto-validate. Do NOT call `vivechak_validate` before `save_plan` unless testing malformed drafts."*

---

### 3. File & Artifact Layout Ergonomics

#### Issue A: Empty DAG Directory Clutter in Comparison Scope
* **What Happened:** When `save_plan` saved the comparison file, it also automatically initialized:
  * `research/sessions/` (empty folder)
  * `research/templates/` (empty folder)
* **The Friction:** In a 1-session comparison (`scope='comparison'`), there is no DAG, no multi-session loop, and no template instantiation. Leaving empty folders makes the repository look like an incomplete project pipeline.
* **Suggested Fix:** Lazy-create `sessions/` and `templates/` only when running full `project` or multi-session `decision` pipelines. For standalone comparisons, keep the structure minimal:
  ```text
  research/
  ├── .vivechak.json
  └── C-01-cmp-winston-vs-pino.md
  ```

#### Issue B: Directory Partitioning for Scalability
* **Observation:** The generated file was placed directly at `research/C-01-cmp-winston-vs-pino.md`. 
* **Suggested Fix:** As a project evolves, teams will have ADRs, sessions, comparisons, and architecture docs. Grouping artifacts by subfolder keeps the workspace tidy:
  * `research/comparisons/C-01-winston-vs-pino.md` (or `research/decisions/` for ADRs).

#### Issue C: In-Document Frontmatter Duplication
* **Observation:** The required YAML frontmatter requires `id`, `informs_decisions`, and `status`, but the markdown body also requires `# Winston vs Pino ... ## Research Question ... ## Prior`.
* **Suggested Fix:** Allow the validator to auto-populate the document title (`# Title`) and ID from the frontmatter header to reduce boilerplate repetition inside the markdown body.


---



* **direct_mcp (deterministic) on S-01:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-02:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
### 1. Developer Experience (DX) Friction

* **CWD Fallback Failure for `project_root`**:
  * *Issue*: [`instructions.md`](file:///C:/Users/ai2/.gemini/antigravity-cli/mcp/vivechak/instructions.md#L15) specifies that `project_root defaults to CWD`. However, calling `vivechak_save_plan` without an explicit `project_root` resulted in a hard failure:
    > `could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set, research/ not found walking up from ...`
  * *Fix*: If `research/` does not exist yet, default to `process.cwd()` automatically rather than aborting. This is especially critical for the `comparison` scope, which is documented as `"5m! No DAG/gate/init."`
* **Identifier Parameter Redundancy**:
  * *Issue*: The [`vivechak_save_plan.json`](file:///C:/Users/ai2/.gemini/antigravity-cli/mcp/vivechak/vivechak_save_plan.json) schema defines three overlapping ID fields: `id`, `comparison_id`, and `decision_id`.
  * *Fix*: Consolidate onto a single canonical `id` field (with backward-compatible aliasing), or automatically parse the ID from the document header (e.g., `# Comparison Plan: C-01` in `content`).
* **Unclear Role of `slug`**:
  * *Issue*: The `slug` parameter is present in the schema, but in comparison mode the generated file was named generic `C-01-comparison.md` regardless.
  * *Fix*: Clarify in the schema if `slug` is used to name comparison artifacts (e.g., `research/C-01-pino-vs-winston.md`) or if it only applies to `decision` scope.

---

### 2. Prompt & Guidance Ambiguity

* **Rule Conflict on Generator Invocation**:
  * *Issue*: Under *Mandatory Trigger Rules*, line 4 states *"ALWAYS invoke vivechak_prepare_generator for comparisons"*, while line 5 states *"User-provided plans/matrices save directly via save_plan."*
  * *Fix*: Clarify rule 4 to read: *"ALWAYS invoke prepare_generator unless the user has already supplied a complete plan/matrix (in which case proceed directly to save_plan)."*
* **Actionable Parameters in `next_step`**:
  * *Issue*: The returned `next_step` suggested: `"(Optional: run vivechak_challenge to audit empirical claims)"`, but did not specify what arguments `vivechak_challenge` expects (e.g., target file path, claim string, or ID).
  * *Fix*: Include concrete call suggestions in `next_step`, such as `vivechak_challenge(target="research/C-01-comparison.md")`.

---

### 3. File & Artifact Layout

* **Ephemeral Lockfiles Left in Workspace**:
  * *Issue*: A 0-byte lockfile ([`C-01-comparison.md.lock`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-02-agy-1792508689572/research/C-01-comparison.md.lock)) remained on disk after the write completed.
  * *Fix*: Ephemeral locks should be removed upon transaction completion, or `*.lock` should be included in the auto-generated [`.gitignore`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-02-agy-1792508689572/research/.gitignore) to avoid dirtying git status.
* **Extraneous Scaffolded Directories**:
  * *Issue*: Even though comparison scope requires no DAG or sessions, an empty [research/sessions/](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-02-agy-1792508689572/research/sessions) directory and session templates were created.
  * *Fix*: Keep comparison scope minimal. Avoid scaffolding project-level DAG folders unless migrating from a comparison to a full project pipeline.
* **Semantic Artifact Filenames**:
  * *Issue*: The file was saved as `C-01-comparison.md`. In a workspace with multiple comparisons, searching for specific technologies requires full-text search.
  * *Fix*: Incorporate the compared topics or slug into the filename (e.g., `research/C-01-pino-vs-winston.md`).


---



* **direct_mcp (deterministic) on S-02:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-03:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-04:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
(No output captured)

---



* **direct_mcp (deterministic) on S-04:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-05:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-06:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-07:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-08:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Based on the hands-on execution in this session, here are concrete, actionable suggestions for improving Vivechak:

---

### 1. Developer Experience (DX) Friction

#### 1.1 `vivechak_validate` Lacks a Direct `file_path` / `path` Argument
* **What happened:** Attempting to validate a single file on disk via `vivechak_validate(file_path="research/D-001-...")` threw an unhandled schema error:  
  `unexpected additional properties ["file_path"]`.
* **The issue:** The tool only accepts `artifact_type`, `content`, and `project_root`. To validate a specific file on disk, you must either read the entire file into memory and pass it to `content`, or omit `content` to validate the *entire* workspace.
* **Suggestion:** Add a `path` or `file_path` parameter to `vivechak_validate` so agents and users can validate individual files directly without loading them into the conversation context.

#### 1.2 No "Stateless / Ad-Hoc" Mode for `vivechak_challenge`
* **What happened:** Calling `vivechak_challenge` on an idea or proposal failed because no workspace existed and no prior session/ADR was recorded on disk (`session or decision D-001 not found`).
* **The issue:** Often a developer wants an immediate red-team critique of an idea, proposal, or PR snippet without having to first run `vivechak_init`, format an ADR, and persist files to disk.
* **Suggestion:** Support an ad-hoc mode in `vivechak_challenge`:
  ```json
  {
    "mode": "red_team",
    "inline_proposal": "Microservices with Kafka for a 3-person team MVP"
  }
  ```
  If `session_id`/`decision_id` is omitted but `inline_proposal` or `text` is provided, generate the 4-part challenge prompts directly against that input.

#### 1.3 Redundant `project_root` Passing Across Every Tool
* **What happened:** Every single tool call (`init`, `record_decision`, `challenge`, `validate`) required explicitly passing the full absolute path `project_root: "C:/Users/.../s-08-agy-1792509182375"`, otherwise tools errored with:  
  `"could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set..."`.
* **Suggestion:** Once `vivechak_init` initializes a workspace within an active MCP server session, the server should cache that `project_root` as the active session default so subsequent tool calls can omit it.

---

### 2. Guidance & `next_step` Ambiguity

#### 2.1 Target-Blind `next_step` After Challenging a Decision
* **What happened:** After challenging decision `D-001`, `vivechak_challenge` returned:
  > `"next_step": "Execute these challenge prompts in your research session. If findings change or new risks emerge, save the updates via vivechak_amend_session."`
* **The issue:** The target was a **Decision Record (ADR)**, not a research session. Recommending `vivechak_amend_session` for a decision is misleading because `vivechak_amend_session` expects session files in `research/sessions/`, not ADR files in `research/`.
* **Suggestion:** Make `next_step` aware of the target type:
  * For a session: `"save the updates via vivechak_amend_session."`
  * For an ADR: `"update the ADR via vivechak_record_decision or edit the decision file."`

#### 2.2 Premature Evidence Warnings on `proposed` Decisions
* **What happened:** Saving the unvetted decision with `status: proposed` triggered:
  `[L3-WARN] W-NO-EVIDENCE-GRADES: No inline evidence grades found (expected A-E grades per P3)`.
* **The issue:** When capturing a team's raw, unvalidated initial proposal to be challenged, it inherently lacks evidence. Flagging this as a warning before the red-team challenge or research session has occurred penalizes recording raw ideas.
* **Suggestion:** Suppress `W-NO-EVIDENCE-GRADES` when `status: proposed`, and only enforce it when transitioning an ADR to `status: accepted`.

---

### 3. File & Artifact Layout Ergonomics

#### 3.1 Single Source of Truth (SSOT) Violation: Dual ADR Storage
* **What happened:** `vivechak_record_decision` created **both**:
  1. [`research/D-001-microservices-kafka-3-person-mvp.md`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-08-agy-1792509182375/research/D-001-microservices-kafka-3-person-mvp.md) (the standalone ADR).
  2. [`research/DECISIONS.md`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-08-agy-1792509182375/research/DECISIONS.md) (where the entire markdown body was duplicated).
* **The issue:** When amending or updating the decision post-challenge, we had to edit both files to prevent them from drifting out of sync.
* **Suggestion:**
  * Keep [`research/DECISIONS.md`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-08-agy-1792509182375/research/DECISIONS.md) strictly as a **registry/index table** (ID, Title, Status, Door Type, Date, Summary, and link to file).
  * Keep the full decision body exclusively in `research/decisions/D-NNN-[slug].md`.

#### 3.2 Template Pollution in Workspace Root
* **What happened:** `vivechak_init` copied `research/templates/DECISIONS.template.md` and `research/templates/CONFLICT-RESOLUTION.template.md` directly into the project repository.
* **The issue:** In production codebases, teams rarely want static tool templates checked into their Git history alongside application code.
* **Suggestion:** Store templates internally within the MCP server or under a hidden metadata folder (e.g. `.vivechak/templates/`) rather than polluting the user-facing `research/` directory.


---



* **direct_mcp (deterministic) on S-08:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-12:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
### Meta-Improvement Recommendations for Vivechak

Based on this execution session, here are concrete observations and suggested refinements:

---

### 1. Developer Experience (DX) & Parameter Handling

* **Automatic Fallback to Current Working Directory (CWD)**:
  - *Observation:* When [`vivechak_run_gate`](file:///C:/Users/ai2/.gemini/antigravity-cli/mcp/vivechak/vivechak_run_gate.json) was called with `{ "verbose": true }` (omitting `project_root`), the resolver failed with:
    > `could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set, research/ not found walking up from ... and no prior workspace initialized in this session.`
    Only after explicitly providing `project_root: "C:\\path\\to\\dir"` did it cleanly evaluate the target as an uninitialized workspace.
  - *Improvement:* If `research/` is not found walking upward, the resolver should automatically treat `os.Getwd()` as the target workspace and return `"initialized": false` / `"workspace not initialized at <cwd>"`, rather than failing early at path resolution.
* **Universal Parameter Aliasing**:
  - Ensure `root` is consistently accepted as a universal alias for `project_root` across all tools (`run_gate`, `status`, `visualize`, `validate`).

---

### 2. Guidance & `next_step` Precision

* **Scope-Aware Initialization Guidance**:
  - *Observation:* When the gate detected an uninitialized directory, `next_step` simply returned `"Run vivechak_init first."`.
  - *Improvement:* Since `run_gate` is specifically part of the **Project Scope** workflow, `next_step` should guide the model with exact parameterization:
    > `"Run vivechak_init(scope='project') to create your research pipeline templates, complete your research DAG sessions, and compile the FAD before re-running the gate."`
  - This removes any ambiguity for smaller models about whether to initialize in `comparison`, `decision`, or `project` scope.

---

### 3. File & Artifact Diagnostics

* **Pre-Gate Gap Punch List in Uninitialized / Incomplete Workspaces**:
  - *Improvement:* When `run_gate` fails due to an uninitialized workspace or missing artifacts, include a lightweight `readiness_checklist` directly in the JSON response:
    ```json
    {
      "success": false,
      "missing_prerequisites": [
        "research/.vivechak.json",
        "research/sessions/*.md",
        "research/DECISIONS.md",
        "research/FOUNDING-ARCHITECTURE.md"
      ],
      "next_step": "Run vivechak_init(scope='project') to initialize."
    }
    ```
  - This immediately informs agents and developers of the exact files required before the gate can be satisfied.


---



* **direct_mcp (deterministic) on S-12:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-13:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-14:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-16:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-21:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



### Prompt / Brief Clarifications (18)
* **agy (gemini-3.8-flash-high) on S-01:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Based on the end-to-end execution of this comparison session, here are concrete, high-signal suggestions to improve Vivechak’s developer experience, guidance clarity, and artifact ergonomics:

---

### 1. Developer Experience (DX) & Parameter Friction

#### Issue A: `project_root` Resolution Failed on Fresh Workspaces
* **What Happened:** When calling `vivechak_save_plan` without an explicit `project_root`, the tool threw:
  ```
  "could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set, 
   research/ not found walking up from ..., and no prior workspace initialized in this session. 
   Run vivechak_init first or pass project_root explicitly"
  ```
* **The Contradiction:** `instructions.md` explicitly documents:
  * *"Comparison ('A vs B', 1 session): No DAG/gate/init. 5m!"*
  * *"project_root defaults to CWD"*
* **Suggested Fix:** If `scope == "comparison"` (which bypasses `vivechak_init`), the backend should **automatically fall back to `process.cwd()`** and create the `research/` directory without failing or demanding `vivechak_init`.

#### Issue B: Parameter Proliferation (`id` vs. `decision_id` vs. `comparison_id`)
* **What Happened:** The schema for `vivechak_save_plan` defines three separate parameters that represent the exact same entity:
  * `id`: *"universal identifier"*
  * `decision_id`: *"identifier for decision scope or comparison scope"*
  * `comparison_id`: *"optional alias for decision_id in comparison scope"*
* **Suggested Fix:** Unify these into a single primary parameter `id` (e.g., `id: "C-01"`). Handle legacy aliases (`decision_id`, `comparison_id`) silently in the tool's backend adapter without exposing three overlapping fields in the JSON schema.

#### Issue C: Stale `.lock` File Left Behind
* **What Happened:** After `vivechak_save_plan` completed successfully, a 0-byte lock file remained on disk: `research/C-01-cmp-winston-vs-pino.md.lock`.
* **Suggested Fix:** Ensure the file persistence routine releases and removes file locks in a `finally` block so that empty lock files do not pollute the git working tree or cause concurrency conflicts on subsequent tool calls.

---

### 2. Prompt & Guidance Ambiguity

#### Issue A: The "Meta-Prompt" Indirection in `prepare_generator`
* **What Happened:** `vivechak_prepare_generator` returns instructions directing the agent to *generate a research prompt inside a `prompt` code block* with slot replacements (`{slots}`), rather than returning the actionable research brief directly.
* **The Friction:** In an automated agent workflow, this forces the agent into a meta-loop:
  1. Receive instructions on how to write a prompt.
  2. Parse the slot markers.
  3. Execute the research anyway.
* **Suggested Fix:** For automated AI agents, have `vivechak_prepare_generator` return the **already-interpolated research brief** directly. Tell the agent: *"Here is your research brief. Execute searches against these hypotheses and format the output according to the schema below."*

#### Issue B: Ambiguity Around `vivechak_validate` vs. `vivechak_save_plan`
* **What Happened:** `instructions.md` states: *"save_session auto-validates; vivechak_validate for dry-runs"*. It did not clarify whether `save_plan` also auto-validates. Agents worry that saving without validating will trigger a fatal schema rejection, leading them to call `vivechak_validate` first (doubling token usage and execution latency).
* **Suggested Fix:** Clarify in `instructions.md`:
  > *"Both `save_session` and `save_plan` auto-validate. Do NOT call `vivechak_validate` before `save_plan` unless testing malformed drafts."*

---

### 3. File & Artifact Layout Ergonomics

#### Issue A: Empty DAG Directory Clutter in Comparison Scope
* **What Happened:** When `save_plan` saved the comparison file, it also automatically initialized:
  * `research/sessions/` (empty folder)
  * `research/templates/` (empty folder)
* **The Friction:** In a 1-session comparison (`scope='comparison'`), there is no DAG, no multi-session loop, and no template instantiation. Leaving empty folders makes the repository look like an incomplete project pipeline.
* **Suggested Fix:** Lazy-create `sessions/` and `templates/` only when running full `project` or multi-session `decision` pipelines. For standalone comparisons, keep the structure minimal:
  ```text
  research/
  ├── .vivechak.json
  └── C-01-cmp-winston-vs-pino.md
  ```

#### Issue B: Directory Partitioning for Scalability
* **Observation:** The generated file was placed directly at `research/C-01-cmp-winston-vs-pino.md`. 
* **Suggested Fix:** As a project evolves, teams will have ADRs, sessions, comparisons, and architecture docs. Grouping artifacts by subfolder keeps the workspace tidy:
  * `research/comparisons/C-01-winston-vs-pino.md` (or `research/decisions/` for ADRs).

#### Issue C: In-Document Frontmatter Duplication
* **Observation:** The required YAML frontmatter requires `id`, `informs_decisions`, and `status`, but the markdown body also requires `# Winston vs Pino ... ## Research Question ... ## Prior`.
* **Suggested Fix:** Allow the validator to auto-populate the document title (`# Title`) and ID from the frontmatter header to reduce boilerplate repetition inside the markdown body.


---



* **direct_mcp (deterministic) on S-01:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-02:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
### 1. Developer Experience (DX) Friction

* **CWD Fallback Failure for `project_root`**:
  * *Issue*: [`instructions.md`](file:///C:/Users/ai2/.gemini/antigravity-cli/mcp/vivechak/instructions.md#L15) specifies that `project_root defaults to CWD`. However, calling `vivechak_save_plan` without an explicit `project_root` resulted in a hard failure:
    > `could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set, research/ not found walking up from ...`
  * *Fix*: If `research/` does not exist yet, default to `process.cwd()` automatically rather than aborting. This is especially critical for the `comparison` scope, which is documented as `"5m! No DAG/gate/init."`
* **Identifier Parameter Redundancy**:
  * *Issue*: The [`vivechak_save_plan.json`](file:///C:/Users/ai2/.gemini/antigravity-cli/mcp/vivechak/vivechak_save_plan.json) schema defines three overlapping ID fields: `id`, `comparison_id`, and `decision_id`.
  * *Fix*: Consolidate onto a single canonical `id` field (with backward-compatible aliasing), or automatically parse the ID from the document header (e.g., `# Comparison Plan: C-01` in `content`).
* **Unclear Role of `slug`**:
  * *Issue*: The `slug` parameter is present in the schema, but in comparison mode the generated file was named generic `C-01-comparison.md` regardless.
  * *Fix*: Clarify in the schema if `slug` is used to name comparison artifacts (e.g., `research/C-01-pino-vs-winston.md`) or if it only applies to `decision` scope.

---

### 2. Prompt & Guidance Ambiguity

* **Rule Conflict on Generator Invocation**:
  * *Issue*: Under *Mandatory Trigger Rules*, line 4 states *"ALWAYS invoke vivechak_prepare_generator for comparisons"*, while line 5 states *"User-provided plans/matrices save directly via save_plan."*
  * *Fix*: Clarify rule 4 to read: *"ALWAYS invoke prepare_generator unless the user has already supplied a complete plan/matrix (in which case proceed directly to save_plan)."*
* **Actionable Parameters in `next_step`**:
  * *Issue*: The returned `next_step` suggested: `"(Optional: run vivechak_challenge to audit empirical claims)"`, but did not specify what arguments `vivechak_challenge` expects (e.g., target file path, claim string, or ID).
  * *Fix*: Include concrete call suggestions in `next_step`, such as `vivechak_challenge(target="research/C-01-comparison.md")`.

---

### 3. File & Artifact Layout

* **Ephemeral Lockfiles Left in Workspace**:
  * *Issue*: A 0-byte lockfile ([`C-01-comparison.md.lock`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-02-agy-1792508689572/research/C-01-comparison.md.lock)) remained on disk after the write completed.
  * *Fix*: Ephemeral locks should be removed upon transaction completion, or `*.lock` should be included in the auto-generated [`.gitignore`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-02-agy-1792508689572/research/.gitignore) to avoid dirtying git status.
* **Extraneous Scaffolded Directories**:
  * *Issue*: Even though comparison scope requires no DAG or sessions, an empty [research/sessions/](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-02-agy-1792508689572/research/sessions) directory and session templates were created.
  * *Fix*: Keep comparison scope minimal. Avoid scaffolding project-level DAG folders unless migrating from a comparison to a full project pipeline.
* **Semantic Artifact Filenames**:
  * *Issue*: The file was saved as `C-01-comparison.md`. In a workspace with multiple comparisons, searching for specific technologies requires full-text search.
  * *Fix*: Incorporate the compared topics or slug into the filename (e.g., `research/C-01-pino-vs-winston.md`).


---



* **direct_mcp (deterministic) on S-02:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-03:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-04:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
(No output captured)

---



* **direct_mcp (deterministic) on S-04:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-05:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-06:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-07:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-08:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Based on the hands-on execution in this session, here are concrete, actionable suggestions for improving Vivechak:

---

### 1. Developer Experience (DX) Friction

#### 1.1 `vivechak_validate` Lacks a Direct `file_path` / `path` Argument
* **What happened:** Attempting to validate a single file on disk via `vivechak_validate(file_path="research/D-001-...")` threw an unhandled schema error:  
  `unexpected additional properties ["file_path"]`.
* **The issue:** The tool only accepts `artifact_type`, `content`, and `project_root`. To validate a specific file on disk, you must either read the entire file into memory and pass it to `content`, or omit `content` to validate the *entire* workspace.
* **Suggestion:** Add a `path` or `file_path` parameter to `vivechak_validate` so agents and users can validate individual files directly without loading them into the conversation context.

#### 1.2 No "Stateless / Ad-Hoc" Mode for `vivechak_challenge`
* **What happened:** Calling `vivechak_challenge` on an idea or proposal failed because no workspace existed and no prior session/ADR was recorded on disk (`session or decision D-001 not found`).
* **The issue:** Often a developer wants an immediate red-team critique of an idea, proposal, or PR snippet without having to first run `vivechak_init`, format an ADR, and persist files to disk.
* **Suggestion:** Support an ad-hoc mode in `vivechak_challenge`:
  ```json
  {
    "mode": "red_team",
    "inline_proposal": "Microservices with Kafka for a 3-person team MVP"
  }
  ```
  If `session_id`/`decision_id` is omitted but `inline_proposal` or `text` is provided, generate the 4-part challenge prompts directly against that input.

#### 1.3 Redundant `project_root` Passing Across Every Tool
* **What happened:** Every single tool call (`init`, `record_decision`, `challenge`, `validate`) required explicitly passing the full absolute path `project_root: "C:/Users/.../s-08-agy-1792509182375"`, otherwise tools errored with:  
  `"could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set..."`.
* **Suggestion:** Once `vivechak_init` initializes a workspace within an active MCP server session, the server should cache that `project_root` as the active session default so subsequent tool calls can omit it.

---

### 2. Guidance & `next_step` Ambiguity

#### 2.1 Target-Blind `next_step` After Challenging a Decision
* **What happened:** After challenging decision `D-001`, `vivechak_challenge` returned:
  > `"next_step": "Execute these challenge prompts in your research session. If findings change or new risks emerge, save the updates via vivechak_amend_session."`
* **The issue:** The target was a **Decision Record (ADR)**, not a research session. Recommending `vivechak_amend_session` for a decision is misleading because `vivechak_amend_session` expects session files in `research/sessions/`, not ADR files in `research/`.
* **Suggestion:** Make `next_step` aware of the target type:
  * For a session: `"save the updates via vivechak_amend_session."`
  * For an ADR: `"update the ADR via vivechak_record_decision or edit the decision file."`

#### 2.2 Premature Evidence Warnings on `proposed` Decisions
* **What happened:** Saving the unvetted decision with `status: proposed` triggered:
  `[L3-WARN] W-NO-EVIDENCE-GRADES: No inline evidence grades found (expected A-E grades per P3)`.
* **The issue:** When capturing a team's raw, unvalidated initial proposal to be challenged, it inherently lacks evidence. Flagging this as a warning before the red-team challenge or research session has occurred penalizes recording raw ideas.
* **Suggestion:** Suppress `W-NO-EVIDENCE-GRADES` when `status: proposed`, and only enforce it when transitioning an ADR to `status: accepted`.

---

### 3. File & Artifact Layout Ergonomics

#### 3.1 Single Source of Truth (SSOT) Violation: Dual ADR Storage
* **What happened:** `vivechak_record_decision` created **both**:
  1. [`research/D-001-microservices-kafka-3-person-mvp.md`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-08-agy-1792509182375/research/D-001-microservices-kafka-3-person-mvp.md) (the standalone ADR).
  2. [`research/DECISIONS.md`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-08-agy-1792509182375/research/DECISIONS.md) (where the entire markdown body was duplicated).
* **The issue:** When amending or updating the decision post-challenge, we had to edit both files to prevent them from drifting out of sync.
* **Suggestion:**
  * Keep [`research/DECISIONS.md`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-08-agy-1792509182375/research/DECISIONS.md) strictly as a **registry/index table** (ID, Title, Status, Door Type, Date, Summary, and link to file).
  * Keep the full decision body exclusively in `research/decisions/D-NNN-[slug].md`.

#### 3.2 Template Pollution in Workspace Root
* **What happened:** `vivechak_init` copied `research/templates/DECISIONS.template.md` and `research/templates/CONFLICT-RESOLUTION.template.md` directly into the project repository.
* **The issue:** In production codebases, teams rarely want static tool templates checked into their Git history alongside application code.
* **Suggestion:** Store templates internally within the MCP server or under a hidden metadata folder (e.g. `.vivechak/templates/`) rather than polluting the user-facing `research/` directory.


---



* **direct_mcp (deterministic) on S-08:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-12:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
### Meta-Improvement Recommendations for Vivechak

Based on this execution session, here are concrete observations and suggested refinements:

---

### 1. Developer Experience (DX) & Parameter Handling

* **Automatic Fallback to Current Working Directory (CWD)**:
  - *Observation:* When [`vivechak_run_gate`](file:///C:/Users/ai2/.gemini/antigravity-cli/mcp/vivechak/vivechak_run_gate.json) was called with `{ "verbose": true }` (omitting `project_root`), the resolver failed with:
    > `could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set, research/ not found walking up from ... and no prior workspace initialized in this session.`
    Only after explicitly providing `project_root: "C:\\path\\to\\dir"` did it cleanly evaluate the target as an uninitialized workspace.
  - *Improvement:* If `research/` is not found walking upward, the resolver should automatically treat `os.Getwd()` as the target workspace and return `"initialized": false` / `"workspace not initialized at <cwd>"`, rather than failing early at path resolution.
* **Universal Parameter Aliasing**:
  - Ensure `root` is consistently accepted as a universal alias for `project_root` across all tools (`run_gate`, `status`, `visualize`, `validate`).

---

### 2. Guidance & `next_step` Precision

* **Scope-Aware Initialization Guidance**:
  - *Observation:* When the gate detected an uninitialized directory, `next_step` simply returned `"Run vivechak_init first."`.
  - *Improvement:* Since `run_gate` is specifically part of the **Project Scope** workflow, `next_step` should guide the model with exact parameterization:
    > `"Run vivechak_init(scope='project') to create your research pipeline templates, complete your research DAG sessions, and compile the FAD before re-running the gate."`
  - This removes any ambiguity for smaller models about whether to initialize in `comparison`, `decision`, or `project` scope.

---

### 3. File & Artifact Diagnostics

* **Pre-Gate Gap Punch List in Uninitialized / Incomplete Workspaces**:
  - *Improvement:* When `run_gate` fails due to an uninitialized workspace or missing artifacts, include a lightweight `readiness_checklist` directly in the JSON response:
    ```json
    {
      "success": false,
      "missing_prerequisites": [
        "research/.vivechak.json",
        "research/sessions/*.md",
        "research/DECISIONS.md",
        "research/FOUNDING-ARCHITECTURE.md"
      ],
      "next_step": "Run vivechak_init(scope='project') to initialize."
    }
    ```
  - This immediately informs agents and developers of the exact files required before the gate can be satisfied.


---



* **direct_mcp (deterministic) on S-12:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-13:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-14:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-16:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-21:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



### File & Layout Ergonomics (18)
* **agy (gemini-3.8-flash-high) on S-01:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Based on the end-to-end execution of this comparison session, here are concrete, high-signal suggestions to improve Vivechak’s developer experience, guidance clarity, and artifact ergonomics:

---

### 1. Developer Experience (DX) & Parameter Friction

#### Issue A: `project_root` Resolution Failed on Fresh Workspaces
* **What Happened:** When calling `vivechak_save_plan` without an explicit `project_root`, the tool threw:
  ```
  "could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set, 
   research/ not found walking up from ..., and no prior workspace initialized in this session. 
   Run vivechak_init first or pass project_root explicitly"
  ```
* **The Contradiction:** `instructions.md` explicitly documents:
  * *"Comparison ('A vs B', 1 session): No DAG/gate/init. 5m!"*
  * *"project_root defaults to CWD"*
* **Suggested Fix:** If `scope == "comparison"` (which bypasses `vivechak_init`), the backend should **automatically fall back to `process.cwd()`** and create the `research/` directory without failing or demanding `vivechak_init`.

#### Issue B: Parameter Proliferation (`id` vs. `decision_id` vs. `comparison_id`)
* **What Happened:** The schema for `vivechak_save_plan` defines three separate parameters that represent the exact same entity:
  * `id`: *"universal identifier"*
  * `decision_id`: *"identifier for decision scope or comparison scope"*
  * `comparison_id`: *"optional alias for decision_id in comparison scope"*
* **Suggested Fix:** Unify these into a single primary parameter `id` (e.g., `id: "C-01"`). Handle legacy aliases (`decision_id`, `comparison_id`) silently in the tool's backend adapter without exposing three overlapping fields in the JSON schema.

#### Issue C: Stale `.lock` File Left Behind
* **What Happened:** After `vivechak_save_plan` completed successfully, a 0-byte lock file remained on disk: `research/C-01-cmp-winston-vs-pino.md.lock`.
* **Suggested Fix:** Ensure the file persistence routine releases and removes file locks in a `finally` block so that empty lock files do not pollute the git working tree or cause concurrency conflicts on subsequent tool calls.

---

### 2. Prompt & Guidance Ambiguity

#### Issue A: The "Meta-Prompt" Indirection in `prepare_generator`
* **What Happened:** `vivechak_prepare_generator` returns instructions directing the agent to *generate a research prompt inside a `prompt` code block* with slot replacements (`{slots}`), rather than returning the actionable research brief directly.
* **The Friction:** In an automated agent workflow, this forces the agent into a meta-loop:
  1. Receive instructions on how to write a prompt.
  2. Parse the slot markers.
  3. Execute the research anyway.
* **Suggested Fix:** For automated AI agents, have `vivechak_prepare_generator` return the **already-interpolated research brief** directly. Tell the agent: *"Here is your research brief. Execute searches against these hypotheses and format the output according to the schema below."*

#### Issue B: Ambiguity Around `vivechak_validate` vs. `vivechak_save_plan`
* **What Happened:** `instructions.md` states: *"save_session auto-validates; vivechak_validate for dry-runs"*. It did not clarify whether `save_plan` also auto-validates. Agents worry that saving without validating will trigger a fatal schema rejection, leading them to call `vivechak_validate` first (doubling token usage and execution latency).
* **Suggested Fix:** Clarify in `instructions.md`:
  > *"Both `save_session` and `save_plan` auto-validate. Do NOT call `vivechak_validate` before `save_plan` unless testing malformed drafts."*

---

### 3. File & Artifact Layout Ergonomics

#### Issue A: Empty DAG Directory Clutter in Comparison Scope
* **What Happened:** When `save_plan` saved the comparison file, it also automatically initialized:
  * `research/sessions/` (empty folder)
  * `research/templates/` (empty folder)
* **The Friction:** In a 1-session comparison (`scope='comparison'`), there is no DAG, no multi-session loop, and no template instantiation. Leaving empty folders makes the repository look like an incomplete project pipeline.
* **Suggested Fix:** Lazy-create `sessions/` and `templates/` only when running full `project` or multi-session `decision` pipelines. For standalone comparisons, keep the structure minimal:
  ```text
  research/
  ├── .vivechak.json
  └── C-01-cmp-winston-vs-pino.md
  ```

#### Issue B: Directory Partitioning for Scalability
* **Observation:** The generated file was placed directly at `research/C-01-cmp-winston-vs-pino.md`. 
* **Suggested Fix:** As a project evolves, teams will have ADRs, sessions, comparisons, and architecture docs. Grouping artifacts by subfolder keeps the workspace tidy:
  * `research/comparisons/C-01-winston-vs-pino.md` (or `research/decisions/` for ADRs).

#### Issue C: In-Document Frontmatter Duplication
* **Observation:** The required YAML frontmatter requires `id`, `informs_decisions`, and `status`, but the markdown body also requires `# Winston vs Pino ... ## Research Question ... ## Prior`.
* **Suggested Fix:** Allow the validator to auto-populate the document title (`# Title`) and ID from the frontmatter header to reduce boilerplate repetition inside the markdown body.


---



* **direct_mcp (deterministic) on S-01:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-02:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
### 1. Developer Experience (DX) Friction

* **CWD Fallback Failure for `project_root`**:
  * *Issue*: [`instructions.md`](file:///C:/Users/ai2/.gemini/antigravity-cli/mcp/vivechak/instructions.md#L15) specifies that `project_root defaults to CWD`. However, calling `vivechak_save_plan` without an explicit `project_root` resulted in a hard failure:
    > `could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set, research/ not found walking up from ...`
  * *Fix*: If `research/` does not exist yet, default to `process.cwd()` automatically rather than aborting. This is especially critical for the `comparison` scope, which is documented as `"5m! No DAG/gate/init."`
* **Identifier Parameter Redundancy**:
  * *Issue*: The [`vivechak_save_plan.json`](file:///C:/Users/ai2/.gemini/antigravity-cli/mcp/vivechak/vivechak_save_plan.json) schema defines three overlapping ID fields: `id`, `comparison_id`, and `decision_id`.
  * *Fix*: Consolidate onto a single canonical `id` field (with backward-compatible aliasing), or automatically parse the ID from the document header (e.g., `# Comparison Plan: C-01` in `content`).
* **Unclear Role of `slug`**:
  * *Issue*: The `slug` parameter is present in the schema, but in comparison mode the generated file was named generic `C-01-comparison.md` regardless.
  * *Fix*: Clarify in the schema if `slug` is used to name comparison artifacts (e.g., `research/C-01-pino-vs-winston.md`) or if it only applies to `decision` scope.

---

### 2. Prompt & Guidance Ambiguity

* **Rule Conflict on Generator Invocation**:
  * *Issue*: Under *Mandatory Trigger Rules*, line 4 states *"ALWAYS invoke vivechak_prepare_generator for comparisons"*, while line 5 states *"User-provided plans/matrices save directly via save_plan."*
  * *Fix*: Clarify rule 4 to read: *"ALWAYS invoke prepare_generator unless the user has already supplied a complete plan/matrix (in which case proceed directly to save_plan)."*
* **Actionable Parameters in `next_step`**:
  * *Issue*: The returned `next_step` suggested: `"(Optional: run vivechak_challenge to audit empirical claims)"`, but did not specify what arguments `vivechak_challenge` expects (e.g., target file path, claim string, or ID).
  * *Fix*: Include concrete call suggestions in `next_step`, such as `vivechak_challenge(target="research/C-01-comparison.md")`.

---

### 3. File & Artifact Layout

* **Ephemeral Lockfiles Left in Workspace**:
  * *Issue*: A 0-byte lockfile ([`C-01-comparison.md.lock`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-02-agy-1792508689572/research/C-01-comparison.md.lock)) remained on disk after the write completed.
  * *Fix*: Ephemeral locks should be removed upon transaction completion, or `*.lock` should be included in the auto-generated [`.gitignore`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-02-agy-1792508689572/research/.gitignore) to avoid dirtying git status.
* **Extraneous Scaffolded Directories**:
  * *Issue*: Even though comparison scope requires no DAG or sessions, an empty [research/sessions/](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-02-agy-1792508689572/research/sessions) directory and session templates were created.
  * *Fix*: Keep comparison scope minimal. Avoid scaffolding project-level DAG folders unless migrating from a comparison to a full project pipeline.
* **Semantic Artifact Filenames**:
  * *Issue*: The file was saved as `C-01-comparison.md`. In a workspace with multiple comparisons, searching for specific technologies requires full-text search.
  * *Fix*: Incorporate the compared topics or slug into the filename (e.g., `research/C-01-pino-vs-winston.md`).


---



* **direct_mcp (deterministic) on S-02:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-03:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-04:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
(No output captured)

---



* **direct_mcp (deterministic) on S-04:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-05:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-06:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-07:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-08:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Based on the hands-on execution in this session, here are concrete, actionable suggestions for improving Vivechak:

---

### 1. Developer Experience (DX) Friction

#### 1.1 `vivechak_validate` Lacks a Direct `file_path` / `path` Argument
* **What happened:** Attempting to validate a single file on disk via `vivechak_validate(file_path="research/D-001-...")` threw an unhandled schema error:  
  `unexpected additional properties ["file_path"]`.
* **The issue:** The tool only accepts `artifact_type`, `content`, and `project_root`. To validate a specific file on disk, you must either read the entire file into memory and pass it to `content`, or omit `content` to validate the *entire* workspace.
* **Suggestion:** Add a `path` or `file_path` parameter to `vivechak_validate` so agents and users can validate individual files directly without loading them into the conversation context.

#### 1.2 No "Stateless / Ad-Hoc" Mode for `vivechak_challenge`
* **What happened:** Calling `vivechak_challenge` on an idea or proposal failed because no workspace existed and no prior session/ADR was recorded on disk (`session or decision D-001 not found`).
* **The issue:** Often a developer wants an immediate red-team critique of an idea, proposal, or PR snippet without having to first run `vivechak_init`, format an ADR, and persist files to disk.
* **Suggestion:** Support an ad-hoc mode in `vivechak_challenge`:
  ```json
  {
    "mode": "red_team",
    "inline_proposal": "Microservices with Kafka for a 3-person team MVP"
  }
  ```
  If `session_id`/`decision_id` is omitted but `inline_proposal` or `text` is provided, generate the 4-part challenge prompts directly against that input.

#### 1.3 Redundant `project_root` Passing Across Every Tool
* **What happened:** Every single tool call (`init`, `record_decision`, `challenge`, `validate`) required explicitly passing the full absolute path `project_root: "C:/Users/.../s-08-agy-1792509182375"`, otherwise tools errored with:  
  `"could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set..."`.
* **Suggestion:** Once `vivechak_init` initializes a workspace within an active MCP server session, the server should cache that `project_root` as the active session default so subsequent tool calls can omit it.

---

### 2. Guidance & `next_step` Ambiguity

#### 2.1 Target-Blind `next_step` After Challenging a Decision
* **What happened:** After challenging decision `D-001`, `vivechak_challenge` returned:
  > `"next_step": "Execute these challenge prompts in your research session. If findings change or new risks emerge, save the updates via vivechak_amend_session."`
* **The issue:** The target was a **Decision Record (ADR)**, not a research session. Recommending `vivechak_amend_session` for a decision is misleading because `vivechak_amend_session` expects session files in `research/sessions/`, not ADR files in `research/`.
* **Suggestion:** Make `next_step` aware of the target type:
  * For a session: `"save the updates via vivechak_amend_session."`
  * For an ADR: `"update the ADR via vivechak_record_decision or edit the decision file."`

#### 2.2 Premature Evidence Warnings on `proposed` Decisions
* **What happened:** Saving the unvetted decision with `status: proposed` triggered:
  `[L3-WARN] W-NO-EVIDENCE-GRADES: No inline evidence grades found (expected A-E grades per P3)`.
* **The issue:** When capturing a team's raw, unvalidated initial proposal to be challenged, it inherently lacks evidence. Flagging this as a warning before the red-team challenge or research session has occurred penalizes recording raw ideas.
* **Suggestion:** Suppress `W-NO-EVIDENCE-GRADES` when `status: proposed`, and only enforce it when transitioning an ADR to `status: accepted`.

---

### 3. File & Artifact Layout Ergonomics

#### 3.1 Single Source of Truth (SSOT) Violation: Dual ADR Storage
* **What happened:** `vivechak_record_decision` created **both**:
  1. [`research/D-001-microservices-kafka-3-person-mvp.md`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-08-agy-1792509182375/research/D-001-microservices-kafka-3-person-mvp.md) (the standalone ADR).
  2. [`research/DECISIONS.md`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-08-agy-1792509182375/research/DECISIONS.md) (where the entire markdown body was duplicated).
* **The issue:** When amending or updating the decision post-challenge, we had to edit both files to prevent them from drifting out of sync.
* **Suggestion:**
  * Keep [`research/DECISIONS.md`](file:///C:/Users/ai2/AppData/Local/Temp/vivechak-sandboxes/s-08-agy-1792509182375/research/DECISIONS.md) strictly as a **registry/index table** (ID, Title, Status, Door Type, Date, Summary, and link to file).
  * Keep the full decision body exclusively in `research/decisions/D-NNN-[slug].md`.

#### 3.2 Template Pollution in Workspace Root
* **What happened:** `vivechak_init` copied `research/templates/DECISIONS.template.md` and `research/templates/CONFLICT-RESOLUTION.template.md` directly into the project repository.
* **The issue:** In production codebases, teams rarely want static tool templates checked into their Git history alongside application code.
* **Suggestion:** Store templates internally within the MCP server or under a hidden metadata folder (e.g. `.vivechak/templates/`) rather than polluting the user-facing `research/` directory.


---



* **direct_mcp (deterministic) on S-08:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **agy (gemini-3.8-flash-high) on S-12:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
### Meta-Improvement Recommendations for Vivechak

Based on this execution session, here are concrete observations and suggested refinements:

---

### 1. Developer Experience (DX) & Parameter Handling

* **Automatic Fallback to Current Working Directory (CWD)**:
  - *Observation:* When [`vivechak_run_gate`](file:///C:/Users/ai2/.gemini/antigravity-cli/mcp/vivechak/vivechak_run_gate.json) was called with `{ "verbose": true }` (omitting `project_root`), the resolver failed with:
    > `could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set, research/ not found walking up from ... and no prior workspace initialized in this session.`
    Only after explicitly providing `project_root: "C:\\path\\to\\dir"` did it cleanly evaluate the target as an uninitialized workspace.
  - *Improvement:* If `research/` is not found walking upward, the resolver should automatically treat `os.Getwd()` as the target workspace and return `"initialized": false` / `"workspace not initialized at <cwd>"`, rather than failing early at path resolution.
* **Universal Parameter Aliasing**:
  - Ensure `root` is consistently accepted as a universal alias for `project_root` across all tools (`run_gate`, `status`, `visualize`, `validate`).

---

### 2. Guidance & `next_step` Precision

* **Scope-Aware Initialization Guidance**:
  - *Observation:* When the gate detected an uninitialized directory, `next_step` simply returned `"Run vivechak_init first."`.
  - *Improvement:* Since `run_gate` is specifically part of the **Project Scope** workflow, `next_step` should guide the model with exact parameterization:
    > `"Run vivechak_init(scope='project') to create your research pipeline templates, complete your research DAG sessions, and compile the FAD before re-running the gate."`
  - This removes any ambiguity for smaller models about whether to initialize in `comparison`, `decision`, or `project` scope.

---

### 3. File & Artifact Diagnostics

* **Pre-Gate Gap Punch List in Uninitialized / Incomplete Workspaces**:
  - *Improvement:* When `run_gate` fails due to an uninitialized workspace or missing artifacts, include a lightweight `readiness_checklist` directly in the JSON response:
    ```json
    {
      "success": false,
      "missing_prerequisites": [
        "research/.vivechak.json",
        "research/sessions/*.md",
        "research/DECISIONS.md",
        "research/FOUNDING-ARCHITECTURE.md"
      ],
      "next_step": "Run vivechak_init(scope='project') to initialize."
    }
    ```
  - This immediately informs agents and developers of the exact files required before the gate can be satisfied.


---



* **direct_mcp (deterministic) on S-12:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-13:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-14:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-16:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



* **direct_mcp (deterministic) on S-21:**
## 4. Turn 4: Meta-Improvement Probe
**Prompt:**
```text
Based on your execution experience in this session, suggest all potential improvements to Vivechak:
1. DX Friction: Were any tool parameters confusing, redundant, or missing aliases?
2. Prompt / Guidance Ambiguity: Did any instructions in the generator brief or next_step cause hesitation or sub-optimal formatting?
3. File & Artifact Layout: Could the generated markdown artifacts or directory structure be cleaner or more ergonomic?
```

**Model Response:**
Turn 4 Suggestions: Engine interfaces were clean, parameter aliases resolved correctly, and workspace validation succeeded.

---



