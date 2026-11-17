# MCP Tools Overview

The Model Context Protocol (MCP) server for [Vivechak](../README.md) (`vck serve`) exposes 9 tools, 4 prompts, and 8 read-only resources over `stdio`.

- **Complete generated tool schemas & parameters:** [`docs/reference/tools.md`](reference/tools.md)
- **Complete generated diagnostic & error codes:** [`docs/reference/codes.md`](reference/codes.md)
- **Complete generated Phase 0 Exit Gate rules (G1–G12):** [`docs/reference/gate.md`](reference/gate.md)

---

## 1. Canonical Response Envelope

Every tool call returns a single JSON envelope in both `TextContent` and `StructuredContent`:

```json
{
  "ok": true,
  "message": "Session T1-01 saved to research/sessions/T1-01.md.",
  "data": {
    "workspace": { "root": "/path/to/project", "rule": "arg" },
    "id": "T1-01",
    "file_path": "research/sessions/T1-01.md"
  },
  "warnings": [
    {
      "code": "Q-DEPTH",
      "message": "Fewer than 3 key findings documented in session.",
      "fix": "Expand Key Findings with additional graded primary or secondary sources."
    }
  ],
  "next": {
    "tool": "vivechak_next",
    "args": {},
    "why": "Session T1-01 complete. Retrieve the next ready session in the plan DAG."
  }
}
```

When a call fails (`isError: true`), `ok` is `false` and the envelope carries `code` (from [`docs/reference/codes.md`](reference/codes.md)), `message`, and `fix`. Unknown input keys are ignored with warning `W-UNKNOWN-ARG`.

---

## 2. Workspace Resolution Ladder

Every filesystem tool resolves the target project root using the same fail-closed ladder:

1. **Explicit `project_root` argument** passed to the tool call.
2. **Single MCP client root** (`roots/list`), requested only when advertised by the client at `initialize` (at most once per session, refreshed on `notifications/roots/list_changed`).
3. **Namespaced environment variables**: `VIVECHAK_PROJECT_ROOT`, then `VCK_PROJECT_ROOT` (ignoring unexpanded `${...}` macros and unnamespaced variables such as `WORKSPACE` or `PROJECT_ROOT`).
4. **Session memory**: the workspace already resolved earlier in the same MCP session.
5. **Marker walk-up from `cwd`**: an ancestor containing `research/.vivechak.json`.
6. **Guarded `cwd`**: `cwd` itself when it carries a project root manifest (`.git`, `go.mod`, `package.json`, `Cargo.toml`, `pyproject.toml`) and is not a home, temp, or system directory.
7. **Refusal (`E-ROOT-UNRESOLVED`)**: returns an error asking for `project_root` without writing to disk.

---

## 3. Tool Call Sequences by Scope

### Comparison Scope (1 Session → WEP Matrix)
1. `vivechak_start(scope="comparison", question="<options and context>")`
2. Run the returned brief with live web search and grade claims `A`–`E`.
3. `vivechak_save(kind="comparison", id="C-001", content="<WEP matrix markdown>")`

### Decision Scope (1–3 Sessions → Locked ADR)
1. `vivechak_start(scope="decision", question="<architectural dilemma>")`
2. `vivechak_save(kind="plan", scope="decision", id="D-001", content="<decision plan with vivechak-plan block>")`
3. Loop `vivechak_next()` and `vivechak_save(kind="session", id="D-001-S1", content="<session markdown>")`.
4. For One-Way Doors, record and stress-test the ADR with `vivechak_record_decision(id="D-001", content="<ADR markdown>")` and `vivechak_review(mode="red_team", id="D-001")`, then ask the human reviewer to run `vck signoff D-001 --reviewer "<name>"` in a terminal.
5. `vivechak_gate(decision="D-001")`

### Project Scope (4–30 DAG Sessions → FAD)
1. `vivechak_start(scope="project", question="<project vision>")`
2. `vivechak_save(kind="plan", scope="project", id="DAG", content="<pipeline markdown with vivechak-plan block>")` *(stop here on planning-only requests)*.
3. Loop `vivechak_next()`, `vivechak_save(kind="session", id="<session_id>", content="<session markdown>")`, and `vivechak_record_decision(id="D-NNN", auto_draft_from="<session_id>", auto_save=true)`.
4. Synthesize `research/FAD.md` with `vivechak_save(kind="fad", id="SYN-01", content="<FAD markdown>")`.
5. `vivechak_gate()`

Skip Vivechak tools for direct coding, syntax questions, local debugging, and micro-refactors.
