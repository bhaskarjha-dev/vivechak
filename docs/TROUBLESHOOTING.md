# Vivechak Troubleshooting Guide
### Diagnosing and Resolving Common Workspace, MCP, and Research Pipeline Issues

---

## 1. Quick Diagnostic: Run `vck doctor`

Before manually inspecting configuration files, run the built-in workspace integrity checker:

```sh
vck doctor /path/to/your/project
# Or from inside your project repository:
vck doctor
```

`vck doctor` checks for:
- Correct `research/` directory layout and required templates.
- Unregistered or orphaned session files in `research/sessions/`.
- Malformed or missing YAML frontmatter headers.
- Stale decision references (`D-OLD`) after ADR supersession.
- In-flight decisions planned in `RESEARCH-PIPELINE.md` that lack locked ADRs.

---

## 2. Validation & Quality Error Codes

Vivechak employs a 4-level validation ladder (**L1–L4**). Below are the common error codes and their resolutions:

### `V-MISSING-FRONTMATTER` (Level 2 Blocking)
- **Cause:** A markdown file in `research/sessions/` or `research/` is missing the opening `---` YAML frontmatter delimiter, or contains irregular characters before the delimiter.
- **Resolution:** Ensure the document begins with valid YAML frontmatter:
  ```markdown
  ---
  session_id: T1-01
  title: "Primary Datastore Selection"
  date: 2026-10-13
  status: complete
  ---
  ```
  *(Note: Leading markdown `# Heading` titles followed immediately by `---` are supported, but frontmatter must enclose required fields like `session_id`, `title`, and `date`).*

### `V-OWD-NO-CONCERNS` (Level 2 Blocking on One-Way Doors)
- **Cause:** Principle P8 (Structured Falsification) mandates that irreversible architectural decisions (One-Way Doors) explicitly document non-trivial counter-arguments or failure modes.
- **Resolution:** Add a substantive `## Discovered Concerns` section with at least 50 characters detailing failure modes, scaling bottlenecks, or operational trade-offs uncovered during research.

### `V-PATH-ESCAPES-ROOT` (Level 1 Fatal)
- **Cause:** A tool argument attempted to read or write a path outside the workspace root (e.g. `../../etc/passwd` or absolute drive roots).
- **Resolution:** Supply paths that are local to the workspace root or pass an explicit valid project root directory to `vivechak_init`.

### `W-STALE-DOWNSTREAM` (Level 3 Advisory Warning)
- **Cause:** You ran `vivechak_amend_session` on an upstream session (e.g. `T1-01`), but downstream sessions (e.g. `T2-01`) were completed prior to the amendment and may rely on superseded findings.
- **Resolution:** Review the downstream session listed in the warning. If necessary, execute a targeted critique or amend the downstream session with `vivechak_amend_session`.

### `W-STALE-DECISION-REFERENCE` (Level 3 Advisory Warning)
- **Cause:** An ADR was marked as superseded via `vivechak_record_decision(supersedes="D-001")`, but an existing session references the old `D-001` decision.
- **Resolution:** Update references in downstream sessions to the new superseding decision ID (`D-NEW`), or verify that the historical citation is intentional.

### `W-PARSER-SESSION-MISMATCH` (Level 3 Advisory Warning)
- **Cause:** Declared sessions in the pipeline overview table/topology are missing parsed prompt specification blocks in `RESEARCH-PIPELINE.md`. This typically occurs when an LLM generator encounters token output limits while generating large Tier 2 (9–16 sessions) or Tier 3 (17–30 sessions) pipelines.
- **Resolution:** Use phased prompting. Prompt the LLM in a fresh turn to emit the prompt blocks for the remaining layers, and append the missing `### Session ID: ...` sections into `research/RESEARCH-PIPELINE.md`.

### `W-OWD-NO-CHALLENGE` (Level 3 Advisory Warning)
- **Cause:** Track B Check B11 detected that an irreversible One-Way Door decision was not subjected to adversarial red-team stress testing prior to commitment.
- **Resolution:** Execute `vivechak_challenge(session_id="...", mode="red_team")` on the informing session, search for disconfirming evidence, and record findings in the decision record under `## Challenge Findings & Red-Team Stress Test`.

### `W-NO-EVIDENCE-GRADES` (Level 3 Advisory Warning)
- **Cause:** A completed research session or locked ADR does not contain inline evidence citations (e.g. `[Grade: A]`, `[Grade: B (source)]`). Under Principle P3 (Evidentiary Grounding), every substantive factual claim must carry an inline evidence grade.
- **Resolution:** Add inline evidence grades to key claims. *(Note: Initial candidate decisions in `research/DECISIONS.md` with `status: proposed` are intentionally exempt during pipeline planning, as evidence has not yet been gathered).*

### `W-RECALLED-GRADE-CAP` (Level 3 Advisory Warning)
- **Cause:** A claim marked with verification method `recalled` (from model parametric memory) was assigned Grade A, B, or C. Under Principle P3, unverified model recall is capped at Grade D.
- **Resolution:** Verify the claim against canonical primary sources (documentation, source code, benchmarks) using web search and tag it as `fetched` or `cached`, or downgrade the claim to `[Grade: D (recalled)]`.

---

## 3. Cognitive Quality Coaching Advisories (`Q-*`)

During `vivechak_save_session`, the engine surfaces non-blocking `[Q-*]` advisory observations in `next_step` to help agents avoid common research biases:

| Code | Meaning | Recommended Action |
|---|---|---|
| **`Q-BIAS`** | Performative belief delta; prior assumptions were accepted without investigation or started from "Unknown". | Review Principle P8 (Structured Falsification). Explicitly state what prior beliefs were contradicted or refined. |
| **`Q-DENSITY`** | Session cites fewer than 2 distinct independent sources. | Ensure substantive architectural research cites multiple independent sources (official docs, independent benchmarks). |
| **`Q-INFLATION`** | Grade A claims detected without qualified provenance anchors. | Provide canonical URLs (`https://`), formal RFCs (`RFC 9110`), or source repository links for top-tier claims. |
| **`Q-FRESH`** | Date citations reference years older than 2 years from today. | Verify that APIs or benchmarks have not changed in subsequent major library releases. |

---

## 4. Host & MCP Server Connection Issues

### Problem: Host reports "Server vivechak not found" or "Command failed"
- **Cause 1 (macOS GUI App Truncated Path):** Desktop GUI applications (such as Claude Desktop or Cursor) do not inherit shell `$PATH` configurations set in `.zshrc` or `.bash_profile`.
  - **Fix:** Use `vck setup <host>` (e.g. `vck setup claude` or `vck setup cursor`). It automatically resolves and writes the **absolute binary path** to your editor's MCP configuration.
- **Cause 2 (Windows Path Escaping):** JSON files on Windows require escaped backslashes (`C:\\Users\\...\\vivechak.exe`).
  - **Fix:** Running `vck setup <host>` automatically escapes Windows backslashes correctly.

### Problem: Workspace initialized in Application Program Folders
- **Cause:** If an editor launches an agent with CWD pointing to the application installation folder (e.g. `AppData\Local\Programs\...` or `/usr/bin`), Vivechak actively rejects workspace creation to prevent polluting application directories.
- **Fix:** Always provide `project_root` explicitly when initializing:
  ```json
  {
    "scope": "project",
    "project_root": "d:/my-project"
  }
  ```
  Or set the environment variable `"VIVECHAK_PROJECT_ROOT": "${workspaceFolder}"` in your host's MCP server configuration.

### Problem: Verifying MCP Connection Interactively
To test that your Vivechak binary responds to JSON-RPC over `stdio` without an editor:
```sh
npx @modelcontextprotocol/inspector vck serve
```
This launches the official web-based MCP Inspector at `http://localhost:5173`, allowing you to inspect tool schemas and execute live test calls interactively.

---

## 5. Phase 0 Exit Gate Troubleshooting

If `vivechak_run_gate` fails with status `FAIL`:

```
Track A (Reversibility & Bounds): PASS
Track B (Rigor & Evidentiary Checks): FAIL
Issues:
  - Decision D-001 (One-Way): evidentiary threshold not met (no corroborated Grade A/B evidence)
  - Decision D-001 (One-Way): rejected alternatives analysis missing or insufficient (<200 chars)
```

### Remediation for Track B Failures:
1. **Evidentiary Threshold (Check B3):** Ensure One-Way Door decisions cite at least one Grade A source with a canonical URL (`https://...`) or formal standard (`RFC ...`), or reproducible Grade B benchmark.
2. **Verification Method (Check B4):** Ensure factual claims are marked as `fetched` or `cached`. Model parametric memory (`recalled`) is capped at Grade D and cannot anchor irreversible decisions.
3. **Rejected Alternatives (Check B5 & B10):** Ensure the ADR includes a `## Rejected Alternatives & Tradeoffs` section explaining why candidates were dismissed, with at least 200 characters of substantive rationale.
4. **Reversal Triggers (Check B6):** Ensure frontmatter defines a measurable `review_trigger` (e.g. `"Throughput exceeds 50k req/sec or p99 latency > 200ms"`) and scheduled `review_date`.
5. **Adversarial Red-Team Challenge (Check B11):** Run `vivechak_challenge(session_id="...", mode="red_team")` on the informing session to stress-test your conclusions. Document the disconfirming search results, edge case probes, and mitigations in a dedicated `## Challenge Findings & Red-Team Stress Test` section in the ADR.

### Direct Terminal Evaluation via `vck gate`:
You can also run gate evaluations directly from your terminal or CI pipeline without an active MCP agent:
```sh
# Run Phase 0 Gate on the current workspace:
vck gate

# Filter by track (e.g. focus solely on Two-Way Doors Track A or One-Way Doors Track B):
vck gate --track track_a
vck gate --track track_b

# Output CI-ready JSON with machine-parseable diagnostics:
vck gate --format json
```

---

## 6. Concurrency & Advisory Lock Files (`*.lock`)

### Why do 0-byte `.lock` files remain in `research/`?
During mutating operations (such as `vivechak_save_plan`, `vivechak_replan`, and `vivechak_save_session`), Vivechak acquires OS-level advisory file locks (`RESEARCH-PIPELINE.md.lock`, `fad_root.lock`) to prevent concurrent subagents or parallel swarm workers from clobbering research artifacts simultaneously.

- **Advisory Lock Lifecycle:** When a tool call finishes, the process automatically releases the lock. However, the empty 0-byte `.lock` file remains on disk. This is **intentional by design** in standard operating system locking patterns (`flock` on POSIX and `LockFileEx` on Windows) to prevent Time-of-Check to Time-of-Use (TOCTOU) file deletion race conditions across concurrent processes.
- **Git Tracking:** Vivechak automatically writes `research/.gitignore` containing `*.lock` upon workspace initialization, ensuring lock files are never committed to git history.
- **Deadlock Diagnostics:** A 0-byte `.lock` file does **not** mean a lock is currently active or stuck. OS advisory locks are tied to active file descriptors and are freed instantly by the kernel when an MCP tool finishes or when a process exits.
- **Safe Cleanup:** If no MCP server or agent is actively writing to the workspace, `.lock` files can be safely deleted at any time with zero impact on research state:
  ```sh
  # Optional cleanup (bash/zsh):
  rm -f research/*.lock
  # Optional cleanup (PowerShell):
  Remove-Item -Path research/*.lock -Force
  ```
