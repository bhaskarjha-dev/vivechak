# Vivechak Troubleshooting Guide

---

## 1. Quick Diagnostic: Run `vck doctor`

Run the workspace integrity checker before inspecting files manually:

```sh
vck doctor /path/to/your/project
# Or from inside your project repository:
vck doctor
```

`vck doctor` checks:
- `research/` directory structure and `.vivechak.json` metadata.
- Unregistered or orphaned session files in `research/sessions/`.
- Malformed or missing YAML frontmatter headers across `research/` and `research/sessions/`.
- Non-UTF-8 bytes in research artifacts.
- Stale decision references after ADR supersession and planned decisions lacking locked ADRs.

---

## 2. Error, Validation, Warning & Coaching Codes

Every operational error (`E-*`), structural or `vivechak-plan` validation error (`V-*`), warning (`W-*`), and quality coaching code (`Q-*`) emitted by Vivechak is generated from `internal/core/codes.go` in a single reference table:

- **Full diagnostic code catalogue (all emitted codes + `fix` instructions):** [`docs/reference/codes.md`](reference/codes.md)
- **Tool schemas & parameter constraints:** [`docs/reference/tools.md`](reference/tools.md)
- **Phase 0 Exit Gate rule specifications (G1–G12):** [`docs/reference/gate.md`](reference/gate.md)

---

## 3. Host & MCP Server Connection Issues

### Host reports "Server vivechak not found" or "Command failed"
- **macOS GUI App PATH:** Desktop GUI hosts (such as Claude Desktop or Cursor) often do not inherit shell `PATH` additions from `.zshrc` or `.bash_profile`. Run `vck setup <host>` (for example `vck setup claude` or `vck setup cursor`) to write the launcher entry cleanly.
- **Windows Path Escaping:** JSON files on Windows require escaped backslashes (`C:\\Users\\...\\vck.exe`); `vck setup <host>` writes valid JSON directly.

### Workspace resolution refused (`E-ROOT-UNRESOLVED`, `E-ROOT-SYSTEM`, or `E-ROOT-UNSAFE`)
- Vivechak refuses to create `research/` inside `/`, `$HOME`, temp folders, OS/application install directories, or subdirectories lacking a project root anchor.
- Pass `"project_root": "/abs/path/to/project"` in the tool call, or configure `"env": { "VCK_PROJECT_ROOT": "${workspaceFolder}" }` in your host's MCP config (see [`docs/hosts.md`](hosts.md)).

### Corrupt Host Configuration File on `vck setup`
- `vck setup` and `vck mcp-config --write` fail closed on malformed JSON/JSONC/YAML and never overwrite or truncate an unparseable config file. Fix the syntax error in the target config file (or pass `--dry-run` to preview the minimal unified diff).

### Interactive Protocol Inspection
```sh
npx @modelcontextprotocol/inspector vck serve
```

---

## 4. Phase 0 Exit Gate Troubleshooting

When `vivechak_gate` or `vck gate` reports `FAIL`, inspect `data.rules` (or terminal output) for the specific G1–G12 rule IDs:

> **Architectural Disclaimer:** The gate is necessary, not sufficient: it checks structure and the agent's own labels, not the truth of the sources. The sign-off is a typed attestation by the named reviewer; it raises the bar but is not a security boundary.

Key Track B checks on accepted One-Way Door ADRs (full table in [`docs/reference/gate.md`](reference/gate.md)):
1. **Grade A/B Depth (G6) & Provenance (G7):** Include ≥2 Grade A/B claims with `https://` URLs and `fetched` or `human` (`human-provided`) verification labels in the ADR or its informing sessions (`informed_by_sessions` / `informs_decisions`). `recalled`, `cached`, `verified`, and `unverified` do not satisfy G7.
2. **No Placeholders & Substantive Alternatives (G8):** Remove all bracketed template placeholders and document substantive rejected alternatives.
3. **Measurable Review Trigger (G9):** Set a concrete metric or calendar `review_trigger` in frontmatter.
4. **Adversarial Red-Team Challenge (G10):** Run `vivechak_review(mode="red_team", id="<decision id>")` (for example `id="D-001"`) on the decision itself, then cite the returned `record_id` (`D-001-R1`) with ≥50 characters of findings under `## Challenge Findings`. Re-run if the Decision Outcome changes later.
5. **Human Sign-Off (G11):** A human reviewer must run `vck signoff <D-id> --reviewer "<name>"` directly in an interactive terminal and type the decision ID when prompted; MCP tools reject agent-supplied sign-off fields with `E-HUMAN-ONLY-FIELD`.

Headless CLI evaluation:
```sh
vck gate
vck gate --track track_a
vck gate --track track_b
vck gate --format json
```

---

## 5. Concurrency & Advisory Lock Files (`*.lock`)

During mutating operations (`vivechak_save`, `vivechak_record_decision`, `vivechak_amend`, `vivechak_replan`), Vivechak uses OS advisory file locks under `research/`.
- Empty `0`-byte `.lock` files remain on disk after lock release by design (`flock` / `LockFileEx`) to avoid TOCTOU unlink races across concurrent processes, and are ignored via `research/.gitignore`.
- If no `vck` process is writing to the workspace, `research/*.lock` files can be removed safely at any time.
