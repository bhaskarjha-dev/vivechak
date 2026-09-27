---
id: R-02
title: "MCP server architecture for multi-step workflow servers (Vivechak MCP server, Go)"
date: 2026-09-23
status: draft
topic: mcp-architecture
informs_decisions: [D-001]
---

# R-02 — MCP Server Architecture for Multi-Step Workflows (Vivechak, Go)

> **How to read the evidence tags.** Every non-trivial claim carries `[grade · source-id]`; source ids (S1…S66) resolve in §7.
> **A** = normative or first-party primary (spec text, official SDK docs/source, a vendor's own docs for its own product).
> **B** = credible secondary or production evidence (implementer's engineering blog, production repos/PRs/CI, compatibility trackers that cite primaries).
> **C** = single-source practitioner report, community issue, or unverified vendor claim.
> **D** = this report's own inference, or knowledge not re-verified in this session.
> Recommendations carry a separate *confidence* level (High / Med / Low).

## 1. Research Question

**Primary question.** Which architecture should the Vivechak MCP server use so that AI agents in Claude (Desktop and Code), Cursor, VS Code and Antigravity can reliably run a structured research pipeline — initialise a workspace, track session-DAG state, submit and validate research outputs, and be guided through the process — as a production-grade Go server?

| # | Sub-question (from the brief) | Answered in |
|---|---|---|
| Q1 | Workflow state: server memory vs files vs database | §3.3, §5.3 |
| Q2 | Resources vs Tools for exposing the pipeline | §3.6, §5.4 |
| Q3 | Should the server ship Prompts | §3.6, §5.4 |
| Q4 | Long-running operations (e.g. validating a large output) | §3.10, §5.7 |
| Q5 | Naming and description conventions for agent discoverability | §3.5, §3.7, §5.5 |
| Q6 | How to test MCP servers | §3.12, §5.11 |
| Q7 | "Guide" (Approach B) vs "Worker", with disconfirming evidence | §3.1, §5.8 |
| Q8 | How Claude Desktop, Cursor, VS Code (and Antigravity) discover/configure servers | §3.11, §5.10 |
| Q9 | Go SDK idioms; production Go servers | §5.1, §5.2, §5.9 |
| Q10 | Concerns the brief did not list | §6.2 |

**Working definitions.** The brief names *Guide* as Approach B and contrasts it with *Worker* but does not define A or C, so this report assumes:

- **Worker (A)** — the server executes pipeline steps itself and hands the agent finished results. LLM-calling workers are out of scope, so "worker" here means deterministic execution: parsing, scaffolding, validating, file operations.
- **Guide (B)** — the server is the authority on *process*: it stores DAG state, says what is ready and what "done" means, validates submissions and refuses illegal transitions. The agent performs all judgement work with its own tools.
- **Passive toolbox (C, added by this report)** — stateless helpers/validators; process knowledge lives in AGENTS.md, Skills or prompts and the agent orchestrates.

If D-001 labels these differently, the findings still transfer; only the definition of "Guide" matters for the verdict in §3.1.

**Assumptions about Vivechak (not verified, but they drive the design).** (1) The pipeline is a DAG of steps whose outputs are files (Markdown/JSON/YAML) in a repo or workspace. (2) It is local-first with one human, but several agents/clients may touch one workspace at once (e.g. Cursor and Claude Code open on the same repo). (3) The server never calls an LLM. (4) Research outputs may contain untrusted, web-derived text.

**Evidence window.** Sources were retrieved on 2026-09-24 (the brief is dated 2026-09-23). Latest spec: **2026-07-28**. Latest Go SDK release seen: **v1.8.0 (2026-09-04)**. Client behaviour changes monthly, so client-specific claims should be re-confirmed by "Spike 0" (§3.14).

### Coverage checklist (from the brief)

| Deliverable item | Status | Where |
|---|---|---|
| Best practices for multi-step workflow MCP servers | ✅ | §3, §5.2, §5.5–5.8 |
| State-management recommendation with evidence | ✅ | §3.3, §5.3 |
| MCP primitive usage guide (Resources / Tools / Prompts) | ✅ | §3.6, §5.4 |
| Tool naming and description patterns | ✅ | §3.7, §5.5 |
| Go MCP SDK patterns and idioms | ✅ | §5.1, §5.9 |
| Testing strategy | ✅ | §3.12, §5.11 |
| Production MCP servers analysed | ✅ 8 servers; the census is **not** a representative sample | §5.2 |
| "Guide" vs "Worker" validation, incl. disconfirming evidence | ✅ but **no controlled comparison exists** | §5.8 |
| Inline evidence grades | ✅ | throughout |
| Open risks and Discovered Concerns | ✅ | §6 |

## 2. Key Findings

**Bottom line.** Build a **stateless, stdio-first Go server on the official SDK (v1.8.x) that acts as the process authority** for the pipeline: durable state in workspace files, deterministic validation, *enforced* transitions, and a "what to do next" answer on every call — while the agent does all judgement work. Keep the surface to about seven workflow-shaped tools, treat Resources and Prompts as optional mirrors, and deliver the process narrative through *always-on* channels (server instructions + an AGENTS.md snippet + an Agent Skill), because tools that the agent must decide to discover are not reliably used. Validate the Guide pattern with your own agent evals before freezing D-001: the published evidence is convergent but indirect.

**F1 — The protocol just changed underneath you, and clients have not caught up.**
Spec 2026-07-28 removes the `initialize` handshake and protocol-level sessions, puts version/capabilities in per-request `_meta`, adds `server/discover`, and says cross-call state must travel as server-minted handles passed as ordinary tool arguments [A · S1]. List endpoints must not vary per connection or as a side effect of other requests; they may vary only by the authorization presented, which does not help a workflow phase [A · S3]. Roots, sampling and logging are deprecated; Tasks became an extension; server-initiated requests were replaced by multi-round-trip requests [A · S1]. Go SDK v1.7.0+ implements the new revision and stays compatible back to 2024-11-05; v1.8.0 is current [A · S11, S12, S13]. Client adoption is partial: Claude Code keeps stdio servers on the legacy handshake by default [A · S22]; Cursor 3.15/3.19 were reported still requesting 2025-11-25 [C · S34]. **Implication:** no in-connection workflow state, a static tool list, explicit handles — correct in both eras.

**F2 — Workflow state belongs in files inside the workspace.**
An authoritative per-session file, replaced atomically under an advisory lock with a monotonically increasing `rev`, beats server memory (incompatible with F1 and with one stdio process per client) and beats a database for a local-first tool [D, from A · S1 and B · S46–S48]. Every surveyed workflow server that persists state does so in a dot-directory in the project (Task Master `.taskmaster/`, Shrimp `DATA_DIR`, Beads `.beads/`) [B · S46, S47, S48]. Beads is the cautionary tale in both directions: its MCP layer misrouted writes across repositories until every call accepted an explicit workspace path, and its dual JSONL + SQLite storage had integrity bugs before it moved to Dolt [B · S48].

**F3 — Tools are the control plane; Resources and Prompts are optional accelerators.**
The protocol assigns Tools to model control, Resources to application control, Prompts to user control [A · S4]. Vendor docs show Claude Code, Cursor and VS Code all surface Tools, Prompts and Resources [A · S22, S23, S24], but Antigravity's prompt support is reported absent [C · S31] and Claude Desktop details rest on third-party claims [C · S64]. Older third-party matrices that say Cursor lacks Resources/Prompts contradict Cursor's own docs [D · S58 vs A · S23]. **Implication:** nothing critical may live only in a Resource or Prompt.

**F4 — Guide vs Worker: the evidence favours "Guide with server-side enforcement" (B′), and it disfavours prose-only guidance and chatty step-by-step tools.**
For: a practitioner report that models skipped and reordered steps even under emphatic instructions until the sequence moved into a server-side state machine [C · S51]; enforcement-first servers (work-item graph with gates; FSM with refusal-plus-legal-moves) [C · S49, S50]; Beads' DAG "ready" semantics [B · S48]; Anthropic and Block guidance to shape tools around whole workflows and consolidate chains [A · S36; B · S38]; MCP-Bench finding that multi-step orchestration remains a persistent weakness across 20 LLMs [B · S59]. Against (kept as design constraints, §5.8): trigger-layer fragility, approval friction, the 2026-07-28 ban on per-connection tool lists, round-trip cost, the CLI + AGENTS.md alternative, and sampling being a dead end for "Worker via the client's LLM". **No controlled comparison of Guide vs Worker vs Passive was found** [D].

**F5 — Discoverability is a first-class design constraint, not polish.**
In Claude Code, tool search is on by default: only tool names and server instructions load at session start, and descriptions and instructions are truncated at 2,048 characters [A · S22]. Claude Code cuts tool descriptions *silently* (its `/mcp` view shows the untruncated text), so an author can be misled about what the model sees [C · S33]. claude.ai and Claude Desktop were reported to ignore the `instructions` field (Mar–Apr 2026) [C · S33]; a March report of ~500-character description truncation in claude.ai appears superseded by an August comparison in which claude.ai received full descriptions [C · S33]. Handling of `instructions` in Cursor and Antigravity is **unverified**; VS Code's own developer docs list "Server instructions" among the MCP features it supports, alongside Tools, Prompts, Resources and Elicitation [A · S25]. Vercel's Next.js-doc eval found: no-docs baseline 53%; a Skill, left uninvoked in 56% of runs, also scored 53% (worse than baseline on some sub-metrics — an unused skill can add noise); explicit "use this skill" instructions lifted triggering to 95%+ but pass rate to only 79%, and were sensitive to exact wording; a compressed, always-loaded AGENTS.md index reached 100% [B · S40]. Vercel's own caveat: skills still win for narrow, explicitly-requested vertical workflows — arguably close to what a Vivechak Skill is, provided invocation is made explicit rather than left implicit [B · S40]. **Implication:** three always-on channels (instructions, AGENTS.md snippet, Skill with an explicit trigger line) plus front-loaded 3–4-sentence descriptions; verify with a canary in each client.

**F6 — Keep the tool surface small, workflow-shaped and fully annotated.**
Task Master's full 36-tool set costs roughly 21k tokens versus roughly 5k for its 7-tool "core" mode [B · S46]. VS Code shows a confirmation dialog for every tool not marked `readOnlyHint` [A · S25]; a Go workflow-gate server discovered that 21 of its 28 tools lacked annotations, producing three write-confirmation prompts before the first result [B · S53]. Use about 7 tools, consolidate `validate + persist + advance` into one call, annotate every tool, and enforce both with a registration-time check.

**F7 — Errors are part of the guidance protocol.**
Input-validation failures should be tool-execution errors (`isError: true`) so the model can self-correct [A · S5, S6]; the Go SDK now converts schema-validation failures that way (PR #863, merged 2026-03-31) [A · S16]. Make errors actionable (code, cause, remediation, legal next tools) and put decision-critical text in `content`, not only in `structuredContent` [A · S36; D]. Report *content-validation outcomes* (the artifact is not good enough yet) as normal results with `status: needs_revision`, and reserve `isError` for calls that could not do their job — a judgement call on which sources disagree (§5.13).

**F8 — You do not need the Tasks extension.**
Tasks is now a Final extension (`io.modelcontextprotocol/tasks`) [A · S7], but the v1.8.0 Go SDK API index shows no Tasks types [A · S12], and whether any Go runtime for it exists outside the SDK was not confirmed [D]. Sources disagree on Tasks support even in other languages: a FastMCP maintainer says no SDK ships a runtime, while the C# SDK documents one [B/A · S54]. Client support is unverified. Claude Code resets its idle timer on progress notifications and auto-backgrounds long calls [A · S22]. Make validation fast and deterministic; for slow checks use checkpointed, idempotent, resumable calls with progress notifications and a pollable status.

**F9 — Go SDK: adopt v1.8.x and plan for churn.**
`mcp.AddTool` gives typed handlers with schema inference, input validation, structured output and error-to-`isError` conversion [A · S12, S15]. `NewInMemoryTransports` makes protocol-level tests cheap [A · S12]. GitHub's Go MCP server migrated from `mark3labs/mcp-go` to the official SDK and reported v1.7.0-pre.3 serving over half a million users [B · S43; A · S13]. Pitfalls are documented: zero-value output validated on error paths, `json.RawMessage` inference, custom marshallers panicking at registration [C · S19], `additionalProperties:false` strictness [B · S17]. The `MCPGODEBUG` escape hatches are scheduled for removal in v1.9.0 [A · S13, S14].

**F10 — Client discovery is five different config surfaces; workspace discovery is the real integration problem.**
Claude Code: `.mcp.json`/`claude mcp add` [A · S22]. Cursor: `.cursor/mcp.json` with `${workspaceFolder}` [A · S23]. VS Code: `.vscode/mcp.json` (`servers`) or portable `.mcp.json` (`mcpServers`), `code --add-mcp`, gallery [A · S24]. Antigravity: `.agents/mcp_config.json` or `~/.gemini/config/mcp_config.json` [B · S29]. Claude Desktop: `claude_desktop_config.json` or `.mcpb` bundle [A · S26, S27, S28]. Roots are deprecated [A · S1], Claude Code exposes `CLAUDE_PROJECT_DIR` [A · S22], and Claude Desktop has no project concept [C · S29]. **Implication:** resolve the workspace through a documented precedence chain, accept an explicit `workspace_root` argument guarded by an allow-list, and echo the resolved path in every response.

**F11 — Test in layers, and snapshot the tool surface.**
Pure-core tests → in-memory MCP tests (both protocol eras) → tool-schema snapshots (GitHub's "toolsnaps") → real-process stdio test → Inspector CLI in CI → optional conformance suite → agent evals [A · S8, S12; B · S9, S44; A · S36]. Enforce annotation/description invariants at registration time rather than by inspecting the wire.

**F12 — Path confinement and untrusted content are the realistic security risks, and they recur in Anthropic's own reference servers.**
The filesystem server needed two CVE fixes (CVSS 8.4 and 7.3) for a bypassable prefix check and a symlink escape; the Git server separately shipped three more path-related CVEs in January 2026 — five in total across two reference implementations in about ten months [B · S55]. Go 1.24+ has traversal-resistant `os.Root`, expanded in 1.25 [A · S20; B · S21]. Separately, the MCP Inspector shipped an RCE from a missing session token and origin check (CVE-2025-49596, fixed v0.14.1) and the Python SDK shipped with DNS-rebinding protection off by default until v1.23.0 (CVE-2025-66414) — relevant if HTTP is ever added [B · S56]. Research artifacts may carry prompt-injection text, and annotations are untrusted hints [A · S5; B · S35].

### Where sources disagree (full register in §5.13)

1. **Cursor's supported primitives** — vendor docs vs a third-party matrix (trust the vendor; test).
2. **Whether schema-validation failures are protocol errors or tool results** in Go — SDK PR #863 vs a 2026 blog (trust the merged PR; pin a regression test).
3. **Whether a failed *content* validation should set `isError`** — spec wording vs a strict-client bug report vs practitioner rules of thumb (§3.8 picks one and says why).
4. **Tasks runtime availability** — one source says no SDK ships it, another documents one; for Go it is absent.
5. **Tool-count caps** — Cursor 40 (2025 forum, undocumented now) vs possibly 80; Antigravity 100 vs 512/800 by version. Moot at about 7 tools.
6. **Beads' own advice** on MCP vs CLI changed between versions.
7. **How much of a tool description reaches the model** — a March report of ~500 characters in claude.ai vs an August comparison showing full descriptions there and a silent 2,048-character cut in Claude Code (design for the smaller number).

### Where the evidence is weakest

- No head-to-head study of Guide vs Worker vs Passive exists; the case for B′ is convergent practitioner evidence plus design-principle sources.
- Antigravity and Claude Desktop behaviour rests on issues and blog posts (grade C).
- Several "workflow server" data points are single-author reports (grade C); the census in §5.2 may suffer survivorship bias.

## 3. Recommendation

### 3.1 Decision for D-001

**Adopt Approach B′ — "Guide with server-side enforcement" — implemented as a stateless, stdio-first Go MCP server on the official SDK (pin v1.8.x), backed by file-based state in the workspace.**
Confidence: **High** for the architectural shape (stateless server, external state, small tool surface, enforcement); **Medium** that Guide beats a Passive toolbox on real agents (needs the evals in §3.12); **Low** on any client-specific behaviour until Spike 0 (§3.14).

B′ is five commitments:

1. **Process authority vs judgement worker.** The server owns state, transitions, validation and "what next"; the agent owns research and writing.
2. **Enforced, not advisory.** Illegal transitions fail with a structured, self-correcting error; nothing depends on the agent remembering prose [C · S51; C · S50].
3. **Coarse-grained deterministic workers** wherever a chain of calls is purely mechanical — `step_submit` = validate + persist + advance + next-actions [A · S36; B · S38].
4. **Always-on guidance channels** (server instructions, AGENTS.md snippet, Skill) because optional discovery is unreliable [B · S40; A · S22; C · S33].
5. **One core, two front-ends.** The same Go library backs the MCP adapter and a CLI, so IDE agents with a shell have a fallback and the core is testable without MCP [B · S48; D].

### 3.2 Architecture

```
Claude Desktop · Claude Code · Cursor · VS Code · Antigravity
      │ MCP over stdio                  │ shell (optional)
      ▼                                 ▼
cmd/vivechak     `mcp`  |  `init` `status` `submit` `check` `install`
      │                                 │
internal/mcpx  (thin adapter)    internal/cli (thin adapter)
      └───────────────┬─────────────────┘
internal/core      pipeline · DAG · transitions · next-actions   (pure)
internal/validate  rules → report                                 (pure)
internal/store     os.Root · advisory lock · atomic replace · rev
      │
<workspace>/.vivechak/sessions/<session_id>/
      session.json   journal.jsonl   reports/<step>.<rev>.json
internal/guide     instructions text · AGENTS snippet · Skill (embed.FS)
```

- **Transport:** stdio in v1 — every target client supports local stdio [A · S22, S23, S24, S26; B · S29]. If HTTP is added later, the Go SDK accepts the 2026-07-28 protocol over HTTP only with `StreamableHTTPOptions.Stateless = true` [A · S13].
- **Registration discipline:** every tool is registered through one wrapper (`mcpx.Add`) that refuses nil annotations, short descriptions, or a non-object root schema (§3.12).
- **Pipeline definition** (steps, dependencies, acceptance rules, per-step instruction text) lives in versioned YAML embedded in the binary and overridable per workspace — not in Go strings. Shrimp's users customise prompt templates through environment variables, which suggests this is wanted [B · S47; D].

### 3.3 State design (Q1)

**Authority:** one `session.json` per session under `.vivechak/sessions/<id>/`. **Write protocol:** take an advisory lock → read → verify the transition is legal → apply → write a temp file in the same directory → fsync → rename over `session.json` → increment `rev` → unlock. Never hold state in process memory beyond a single call.

- **Optimistic concurrency:** mutating tools accept an optional `expected_rev`; a mismatch returns `stale_revision` with the current status. The lock protects correctness; `expected_rev` protects intent across multiple agents [D].
- **Content addressing:** each artifact reference stores path + SHA-256; each validation report is keyed by artifact hash. If the file's hash later differs, the step becomes `stale` and dependents are blocked until re-validation [D].
- **Journal:** `journal.jsonl` is an append-only audit trail (who/when/what, using client info from the request). It is *not* a second authority — Beads' dual-store design produced data-duplication and data-loss bugs [B · S48].
- **IDs:** server-minted session IDs (short, readable, random suffix) and semantic step IDs (`n2-sources`) rather than UUIDs, since natural-language identifiers reduce agent errors [A · S36].
- **Schema evolution:** `schema_version` in every file; refuse to write files from a newer version.
- **Git:** commit `session.json` and `reports/`; `workspace_init` writes a `.gitignore` for lock and temp files.
- **When to graduate to a database:** measured lock contention, more than a handful of concurrent writers, cross-workspace queries, or DAGs in the thousands of nodes. Beads went to Dolt for merge/sync needs, which is heavy for this use case [B · S48].

Minimal shape:

```json
{
  "schema_version": 1, "session_id": "s-0924-a3f9", "rev": 12,
  "brief": {"question": "…", "profile": "default"},
  "pipeline": {"id": "vivechak.default", "version": "1.3.0", "digest": "sha256:…"},
  "steps": {
    "n1-scope":    {"status": "done",  "deps": [], "artifacts": [{"path": "research/n1-scope.md", "sha256": "…", "report": "reports/n1-scope.4.json"}]},
    "n2-sources":  {"status": "ready", "deps": ["n1-scope"]},
    "n3-analysis": {"status": "pending", "deps": ["n2-sources"]}
  }
}
```

**Transitions:** `pending → ready` (all deps `done`) · `ready|in_progress → done` (submission passes) · `ready|in_progress → needs_revision` (fails) · `done → stale` (artifact hash changed or upstream reopened) · `stale → ready` (re-validate or reopen). Invariants: acyclic; a step is `done` only if all deps are `done` and not `stale`; `rev` strictly increases; identical resubmission is a no-op.

### 3.4 Tool surface (about seven tools)

| Tool | Purpose | Annotations | Notes |
|---|---|---|---|
| `workspace_init` | Create/attach `.vivechak/`; optionally write the AGENTS.md snippet and Skill files | writes; idempotent; non-destructive | Never overwrites user files; returns resolved workspace and pipeline version |
| `session_start` | Create a session from a research brief; returns `session_id`, DAG summary, ready steps | writes; not idempotent | Accepts `request_id` so retries do not create duplicates |
| `session_list` | List sessions in the workspace (paged) | read-only; idempotent | |
| `session_status` | Re-orient: DAG state, ready/blocked/stale steps, `next_actions`, `rev` | read-only; idempotent | `response_format: concise\|detailed`; entry point after any context loss; mark always-load in Claude Code |
| `step_submit` | Validate + persist + advance for a ready step | writes; idempotent on identical hashes | Returns `done` or `needs_revision`, issues with rule ids and fixes, next actions |
| `artifact_check` | Dry-run validation, no state change | read-only; idempotent | Same validators as `step_submit` |
| `step_claim` *(optional, v1.1)* | Advisory lease for parallel agents | writes | Only if parallel sub-agents are used |

Rationale for the shape: consolidation and workflow-shaped tools [A · S36; B · S38]; small surface because tool definitions cost context [B · S46; A · S22]; explicit read-only marking to avoid confirmation fatigue [A · S25; B · S53]. Do **not** unlock tools per phase — see §3.6 and F1.

### 3.5 Always-on guidance (Q5)

Optional discovery is the weakest link, so deliver the process through three independent channels [B · S40; A · S22; C · S33]:

1. **Server `instructions`** (short; front-loaded; ≤ ~1.2 KB, well under Claude Code's 2,048-character cut-off [A · S22]):

   ```text
   Vivechak runs a structured research pipeline as a DAG. You do the research; this server tracks state, checks your outputs and says what to do next.
   Flow: workspace_init (once per repo) → session_start (research question) → session_status (see ready steps) → do the step and write the artifact where the step says → step_submit. If step_submit returns needs_revision, fix the listed issues and resubmit. Repeat until session_status reports complete.
   Rules: never edit files under .vivechak/ directly; use artifact_check to dry-run a draft; on any error read error.remediation and error.allowed_next; after losing context, call session_status.
   ```

2. **AGENTS.md / CLAUDE.md snippet** written by `workspace_init` (3–5 lines: "for research tasks, call `session_status` first"). In Vercel's eval, an always-loaded, compressed AGENTS.md index reached a 100% pass rate versus a 53% no-docs baseline; an available Skill matched the 53% baseline because it went uninvoked in 56% of runs [B · S40]. Beads uses the same trick — AGENTS.md plus a `prime` command that emits workflow context [B · S48].
3. **An Agent Skill** with the long-form procedure, worked examples and edge cases. The standard is read by Claude Code, Cursor, Codex, Gemini CLI, Antigravity and VS Code according to one survey [C · S41]; skill directory paths differ per client, so let `workspace_init` copy it to each client's location (verify in Spike 0). Vercel's caveat is worth keeping: skills still won for narrow, explicitly-requested vertical workflows, and adding an explicit "call this for research tasks" instruction raised their own skill's trigger rate to 95%+ (pass rate 79%) — better than the unprompted 53%, though still short of the AGENTS.md index, and sensitive to exact wording [B · S40]. Treat the Skill as a detail-rich complement invoked via the AGENTS.md trigger line, not as the primary channel.

Claude Code extra: mark `session_status` always-load (`alwaysLoad` in the server entry, or `anthropic/alwaysLoad` in the tool's `_meta`) so it is not hidden behind tool search [A · S22].

### 3.6 Resources and Prompts (Q2, Q3)

**Resources — read-only mirrors, never the only path.** Use resource *templates* (static list, no per-connection variance [A · S3]):

| URI | Content |
|---|---|
| `vivechak://guide` | Long-form process guide (Markdown, static) |
| `vivechak://sessions/{session_id}` | Session status (JSON) |
| `vivechak://sessions/{session_id}/steps/{step_id}` | Step spec + acceptance criteria (Markdown) |
| `vivechak://sessions/{session_id}/reports/{step_id}` | Latest validation report (JSON) |

Tools return a compact inline summary **plus** a `resource_link` to the full report; links returned by tools need not appear in `resources/list` [A · S5]. In Claude Code, resources are `@`-mentionable and the client auto-provides list/read tools for them [A · S22]; VS Code adds them via *Add Context* [A · S24]; Cursor lists Resources as supported [A · S23]. Set short, private cache hints on dynamic resources under 2026-07-28 (the SDK exposes `SetCacheable`) [A · S1; C · S19].

**Prompts — user-invoked shortcuts, three at most:** `start_research(question)`, `resume_research(session_id?)`, `review_session(session_id)`. Each is a short user message telling the agent to call `session_start` / `session_status`; no logic lives in them. They appear as `/mcp__vivechak__start_research` in Claude Code [A · S22] and `/vivechak.start_research` in VS Code [A · S24]. Antigravity reportedly does not surface MCP prompts [C · S31], which is one more reason nothing depends on them.

**Do not use dynamic per-phase tool exposure** (`list_changed`): 2026-07-28 forbids per-connection variance, with the sole exception of variance driven by the authorization presented — not by workflow phase [A · S3], and Cursor staff describe list-changed handling as a known limitation [B · S32]. This is a live tension worth flagging rather than smoothing over: VS Code's own developer docs advertise "dynamic tool discovery," where "a server can provide different tools based on the framework or language detected in the workspace, or in response to the user's chat prompt" [A · S25] — exactly the phase-gating pattern the spec's list-invariance rule appears to rule out. The likely reconciliation is that VS Code's dynamic discovery varies the list once per session based on workspace context rather than repeatedly mid-connection in response to tool results, which may or may not satisfy "MUST NOT vary... as a side effect of other requests" depending on interpretation — this was not resolved in this pass. Vivechak's design sidesteps the ambiguity either way: gate by phase inside handlers, and return the legal next actions in the error, rather than depending on any client's list-changed behaviour.

### 3.7 Naming and description rules (Q5)

1. **`resource_verb`, snake_case, `[a-z0-9_]` only.** The spec allows `.` and `-` [A · S5], but Claude Code composes `mcp__<server>__<tool>` names and rewrites unusual characters in some names [A · S22]; the conservative charset avoids surprises [D]. Namespace by the noun the agent operates on (`session_*`, `step_*`, `artifact_*`, `workspace_*`); Anthropic recommends resource-based namespacing and says prefix vs suffix choices measurably affect tool-use, so evaluate [A · S36].
2. **Do not repeat the server name** in tool names (clients already prefix it in Claude Code [A · S22]), but keep a distinctive noun in the entry tool so tool search can find it [D].
3. **One verb per meaning; no generic verbs.** `start`, `status`, `list`, `submit`, `check`, `claim`, `init`. Avoid `run`, `process`, `handle`, and near-synonym pairs (`check` vs `validate` vs `verify`). Shrimp-style chains of similar tools (`analyze` / `reflect` / `verify`) were flagged by an external audit for lacking when-to-use-vs-alternative guidance [B · S47; C · S42].
4. **Unambiguous, fully qualified parameters** (`session_id`, `step_id`, `artifact_paths`) [A · S36].
5. **Descriptions of at least 3–4 sentences** [A · S37] that say: what it does; when to use it and when *not* (name the sibling); preconditions and side effects; what it returns and what to do next. Put the decision-critical part in the first ~500 characters [C · S33; A · S22].
6. **Two response shapes.** `content` = concise, agent-oriented text (default ≤ ~1.5k tokens); `structuredContent` = full data; read tools take `response_format` (`concise|detailed`) [A · S36].
7. **Budget:** about 7 tools and roughly 3–4k tokens of tool definitions in total [D, sized against B · S46].

Template (`step_submit`):

```text
Submit the artifact you wrote for a ready research step. The server validates it against the step's acceptance rules, records the result, and unlocks dependent steps when it passes. Use it after writing the file at the path shown by session_status; use artifact_check instead to dry-run a draft without changing state. Fails with precondition_failed if the step is not ready. Returns status (done | needs_revision), issues with rule ids and fixes, and next_actions.
```

### 3.8 Error and validation contract (Q5, Q6)

| Class | Examples | Wire form | Agent experience |
|---|---|---|---|
| Protocol error | unknown tool, malformed request | JSON-RPC error (SDK-generated) | Rare; not used for workflow logic |
| Tool execution error | `invalid_argument`, `workspace_unresolved`, `workspace_not_initialized`, `path_outside_workspace`, `not_found`, `precondition_failed`, `stale_revision`, `locked`, `internal` | `isError: true` | Model sees text and can self-correct [A · S5, S6] |
| Validation outcome | artifact fails acceptance rules | normal result, `status: needs_revision`, `issues[]` | The report is the deliverable |

Every error carries `code`, `message`, `remediation`, `allowed_next[]` (tool names), `retryable`, and `state{session_id, rev}` — in `content` text always, in `structuredContent` too. `precondition_failed` must list the currently ready steps and the tool that would move things forward; that is what lets an agent recover without a human (the refusal-plus-reachable-actions pattern [C · S50]).

**Why validation outcomes are not `isError` (a judgement call [D]):** the report is what the tool exists to produce; strict clients in some SDKs validate `structuredContent` against `outputSchema` even on error results, which makes rich error-path payloads fragile [C · S61]; and separating "you called me wrongly" from "your work isn't good enough yet" keeps client error UI meaningful. The opposing view — the spec lists business-logic failures among tool-execution errors [A · S5], and practitioner guidance reserves protocol errors for the framework [C · S60] — is legitimate. Settle it with the A/B agent eval in §3.12.

### 3.9 Workspace resolution

Resolve in this order, canonicalising with `EvalSymlinks`, and **echo the resolved absolute path in every response** so a wrong-workspace call is visible:

1. Tool argument `workspace_root` — accepted only inside the allow-list (default: the launch workspace; widen with `--allow-root`).
2. `--workspace <dir>` — config templates pass `${workspaceFolder}` where supported (Cursor, VS Code) [A · S23, S24].
3. `CLAUDE_PROJECT_DIR` — set by Claude Code in the server's environment [A · S22].
4. `VIVECHAK_WORKSPACE`.
5. Client roots — legacy protocol only; deprecated in 2026-07-28 [A · S1]; Claude Code answers with its launch directory [A · S22].
6. Process cwd, then walk up to the nearest directory containing `.vivechak/` or `.git`.

If nothing resolves: `workspace_unresolved`, telling the agent to pass `workspace_root`. Evidence for the explicit argument and the echo: Beads' MCP server misrouted operations across repositories until every tool took an optional workspace path with canonicalised routing; it also flags one-server-per-project setups as error-prone because the agent may pick the wrong server [B · S48].

### 3.10 Validation and long-running work (Q4)

- Validators are pure, deterministic functions of (artifact bytes, step spec) with stable rule IDs; a context deadline (default ≈30 s) and a size cap (`artifact_too_large` with remediation). The SDK already depends on a JSON Schema library; reuse it for structured artifacts [D].
- Keep the fast path synchronous. If network-bound checks (link liveness, citation lookup) are added, expose them as a separate tool (v1.1) that persists per-item results as it goes, emits progress notifications, and is resumable — re-calling continues from the checkpoint and returns `complete:false` plus a resume hint. Claude Code resets its idle timer on progress notifications, and auto-backgrounds calls running past two minutes [A · S22]. In Go: `req.Params.GetProgressToken()` and `req.Session.NotifyProgress(...)` [A · S12].
- Keep responses under about 10k tokens (Claude Code warns beyond that and caps at 25k by default) [A · S22]; return summary + `resource_link` for anything bigger.
- Do not adopt the Tasks extension until the official Go SDK ships it and the target clients support it [A · S7, S12; C · S54].

### 3.11 Client configuration and distribution (Q8)

Detail and snippets in §5.10. In short:

- Commit **project-scoped config** for each client: `.mcp.json` (Claude Code; also read by VS Code as a portable format), `.vscode/mcp.json`, `.cursor/mcp.json`, `.agents/mcp_config.json` (Antigravity) [A · S22, S23, S24; B · S29].
- Ship `vivechak mcp install --client <claude-code|cursor|vscode|antigravity|claude-desktop> [--scope project|user]`, using each client's own mechanism where one exists (`claude mcp add`, `code --add-mcp`) and a merge-safe file edit otherwise [A · S22, S24; D].
- Claude Desktop: document `claude_desktop_config.json` with an explicit `--allow-root`, and offer a `.mcpb` bundle whose user-config prompt collects the workspace directory [A · S26, S27, S28].
- The official MCP Registry is metadata-only and was still described as preview in its own quickstart at retrieval; treat it as an optional discoverability channel after v1 [A · S10].

### 3.12 Test plan (Q6)

1. **Core tests (no MCP):** table-driven transitions; property tests for invariants (acyclic, monotonic `rev`, idempotent resubmission); crash-consistency tests (kill between temp write and rename); multi-process contention under `-race`; golden validation reports.
2. **In-memory MCP tests:** every tool through `NewInMemoryTransports`, run once per protocol era (default and `2025-11-25`); assert *tool errors, not protocol errors*, for every failure class [A · S12, S16].
3. **Tool-surface snapshots:** one golden JSON per tool, GitHub "toolsnaps"-style, with an explicit update flag and CI failure on drift [B · S44]; plus registration-time invariants (annotations present, description length band, root schema `type: object`, no root `$ref`/combinators, property-name rules) [A · S22; B · S48, S53].
4. **Process test:** spawn the real binary with `CommandTransport`; a stray stdout write breaks it, which is the point [A · S12].
5. **Inspector CLI smoke test** in CI (`tools/list`, one `tools/call`) [A · S8]; keep Inspector patched to at least v0.14.1, which closed an RCE in its proxy [B · S56].
6. **Conformance suite** — optional; it targets a server URL, so it matters only if HTTP is shipped [B · S9].
7. **Agent evals** — realistic multi-call tasks, verifiable outcomes, and tracked tool-call counts, errors and tokens, as Anthropic advises [A · S36]. Scenarios: happy path; skip-ahead temptation; invalid artifact; stale `rev`; two agents; context loss and resume; path-traversal attempt; injected instructions inside an artifact. Conditions: tools only vs +instructions vs +AGENTS.md +Skill. Run per target model/client. Include the `isError` vs `needs_revision` A/B.

### 3.13 Security checklist

1. All filesystem access through an `os.Root` opened at the resolved workspace (Go ≥ 1.25 so `Rename`, `WriteFile`, `MkdirAll` work on it) [A · S20; B · S21].
2. Accept workspace-relative artifact paths only; refuse `..` and escaping symlinks. Anthropic's own reference servers have a poor track record here: the filesystem server's naive prefix check was defeated by symlinks and by sibling directories sharing a string prefix (CVE-2025-53109/-53110), and the official Git server separately shipped an unconstrained `git_init` path, a bypassable repository-boundary check, and a `git_diff` misuse that could overwrite arbitrary files (CVE-2025-68143/-68144/-68145) — five path-confinement CVEs across two reference servers in about ten months [B · S55]. A security vendor's scan put path-traversal bugs in a large share of MCP servers surveyed; treat that figure as a single vendor's own study rather than a settled industry number [C · S55].
3. Treat artifact content as untrusted data: never echo whole files; excerpt with length caps and fencing; content never changes state except through validators [D].
4. No shell execution and no network by default; if source verification is added, add an allow-list and SSRF protection [D].
5. Annotations are hints, not controls [A · S5; B · S35]; enforce the state machine server-side.
6. stdout carries JSON-RPC only; log with `slog` to stderr (the logging capability is deprecated) [A · S1].
7. If HTTP is ever added: bind localhost, and explicitly configure DNS-rebinding protection rather than assume it is on by default — the MCP Inspector's own proxy shipped an RCE (CVE-2025-49596, CVSS 9.4, fixed in v0.14.1) from missing session auth and origin checks, and the Python SDK separately shipped with DNS-rebinding protection off by default for unauthenticated localhost HTTP servers until v1.23.0 (CVE-2025-66414) [B · S56]. Whether the Go SDK's HTTP transport had the same default gap was not independently confirmed in this pass — verify against the pinned version before relying on it.
8. Supply chain: pin the SDK, run `govulncheck` [D], sign releases, and let snapshot tests catch description drift [C · S57].
9. For override-style tools in Claude Code, set `anthropic/requiresUserInteraction` [A · S22].

### 3.14 Rollout and exit criteria

**Spike 0 (½–1 day, before freezing D-001).** Build a throwaway Go server with a canary token in `instructions`, one read-only tool, one prompt, one resource and one slow tool. Run it in all five clients and fill this matrix:

| Probe | Claude Desktop | Claude Code | Cursor | VS Code | Antigravity |
|---|---|---|---|---|---|
| Model sees `instructions` canary? | | | | | |
| Agent finds `session_status` unprompted (tool search / deferral)? | | | | | |
| Prompts listed? Resources listable/attachable? | | | | | |
| Negotiated protocol version / handshake type | | | | | |
| cwd, `CLAUDE_PROJECT_DIR`, `${workspaceFolder}`, roots as seen by the server | | | | | |
| Approval prompts with vs without `readOnlyHint` | | | | | |
| Progress display; timeout behaviour at 60 s / 3 min | | | | | |
| Behaviour with a 12k-token result | | | | | |
| Skill directory path and auto-discovery | | | | | |

**Phases.** P1 core + CLI (state, validators, crash/property tests). P2 MCP adapter (six tools, in-memory and stdio tests, Inspector CI). P3 guidance assets + evals (iterate descriptions and instructions). P4 hardening and distribution (install helper, MCPB, optional registry entry, optional HTTP).

**Proposed exit criteria (thresholds are D, to be tuned).** Zero state corruption across crash and contention suites; ≥ 95% happy-path completion on each target model; ≥ 90% of out-of-order attempts recovered using only the error's remediation; ≤ 1 confirmation prompt per read in VS Code with annotations; tool-definition budget ≤ 4k tokens.

### 3.15 What would change this recommendation

- Spike 0 shows `instructions`, AGENTS.md *and* Skills are all ignored in ≥ 3 of 5 clients → make tool responses more prescriptive and consider a CLI-first approach for shell-capable agents.
- Evals show B′ no better than Passive on the target models → drop enforcement complexity, keep validators and state.
- Measured contention or multi-user needs → move authority to SQLite (or a remote service with Stateless HTTP).
- The official Go SDK ships Tasks *and* target clients support it → revisit long-running design.
- Target clients start honouring per-connection list changes safely → phase-based tool exposure becomes possible, but it is still not required.

## 4. Alternatives Considered

| Alternative | Verdict | Why | Revisit if |
|---|---|---|---|
| **In-memory session state** (single-process, per connection) | Rejected | Contradicts the 2026-07-28 sessionless model [A · S1]; lost on restart; two stdio processes on one repo diverge; Beads' routing bugs are the same failure class [B · S48] | Never for workflow state; fine for caches |
| **Embedded database (SQLite/bbolt) as authority** | Deferred | Transactions are attractive, but the binary is opaque to git review and adds cgo or pure-Go driver choices [D]; Beads' hybrids show integrity risks [B · S48] | Contention or cross-workspace queries are measured |
| **Event-sourced journal as authority** | Deferred | Robust to partial writes, but needs folding, compaction and snapshotting; more code for a small DAG [D] | Audit or replay becomes a product requirement |
| **Worker/executor server** (server performs steps) | Rejected for judgement steps; adopted for mechanical ones | Sampling is deprecated, so borrowing the client's LLM is a dead end [A · S1]; LLM-calling servers are out of scope; mechanical chains are folded into `step_submit` [A · S36] | Steps become deterministic transformations |
| **Passive toolbox + AGENTS.md/Skill only** | Strong contender; retained as complement and fallback | Cheapest; keeps process in prose that models may skip [C · S51]; always-loaded context is reliable [B · S40] | Evals show no gain from enforcement |
| **CLI only (no MCP)** | Retained as second front-end | Beads recommends CLI + AGENTS.md for shell-capable agents [B · S48]; Claude Desktop chat lacks a shell [D] | MCP support proves unreliable in target clients |
| **Fine-grained per-step tools** (`plan_task`, `analyze_task`, `reflect_task`, …) | Rejected | Round trips, approval fatigue and disambiguation cost [A · S36; A · S25; C · S42] | A tool-search-first world makes large surfaces cheap |
| **Dynamic tool exposure per phase** | Rejected | Forbidden per connection in 2026-07-28 [A · S3]; weakly supported in Cursor [B · S32] | Never as a requirement |
| **Elicitation/sampling/roots-driven flows** | Avoided | Sampling, roots and logging are deprecated; server-initiated requests replaced by MRTR [A · S1]; uneven client support [C · S64] | Clients adopt MRTR uniformly |
| **Tasks extension for long validations** | Deferred | Final [A · S7] but no official Go runtime [A · S12] and unverified client support | Go SDK ships it and clients support it |
| **`mark3labs/mcp-go` instead of the official SDK** | Not chosen; not evaluated in depth | Official SDK is the brief's Tier 1 choice; GitHub migrated to it [B · S43]; this report did not check whether mark3labs supports 2026-07-28 [D] | Official SDK blocks a requirement |
| **Remote Streamable-HTTP service** | Later, optional | Needs `Stateless = true` for 2026-07-28 [A · S13], auth, and multi-tenant state; unnecessary for local-first use | Multi-user or hosted use appears |

## 5. Detailed Findings

### 5.1 Protocol and Go SDK landscape

**Spec 2026-07-28 (final on 2026-07-28 after a release-candidate period)** [A · S1, S2]. The steering announcement frames the cut as removing protocol-level session state so a server can be scaled and load-balanced like an ordinary stateless API, with any needed continuity carried by the application layer instead of the transport [A · S2]:

- Stateless model: no `initialize`/`initialized` handshake; each request carries protocol version, client info and capabilities in `_meta`; `server/discover` lets a client learn versions and capabilities up front [A · S1, S13].
- Sessionless: servers needing cross-call state use explicit, server-minted handles passed as ordinary tool arguments [A · S1].
- Server-initiated requests (elicitation, sampling) are replaced by multi-round-trip requests (MRTR); change notifications are unified under one subscription stream [A · S1, S13].
- Roots, sampling and logging are deprecated. The spec's migration advice: pass directories through tool parameters, resource URIs or server configuration; use direct LLM-provider APIs instead of sampling; log to stderr on stdio [A · S1].
- Tasks moved into an extension (`io.modelcontextprotocol/tasks`), formalised as **SEP-2663**, status **Final** [A · S1, S7].
- Tasks moved into an extension (`io.modelcontextprotocol/tasks`) [A · S1, S7].
- List results gain cache hints (`ttlMs`, `cacheScope`); tool ordering SHOULD be deterministic; list endpoints MUST NOT vary per connection or as a side effect of other requests — the one sanctioned exception is variance by the authorization presented with the request, which does not help a workflow-phase design [A · S1, S3].

A rule that predates 2026-07-28 but matters just as much: since the **2025-11-25** revision, input-validation failures are Tool Execution Errors rather than Protocol Errors (**SEP-1303**, Final), specifically so a model sees the message and can self-correct on retry [A · S6]. The official Go SDK only caught up to this on 2026-03-31 via PR #863 (§5.6) [A · S16].

**Go SDK** (`github.com/modelcontextprotocol/go-sdk`):

| Version | Date | Relevance |
|---|---|---|
| v1.5.0 | 2026-04-07 | Date from a downstream review; includes the input-validation-as-tool-error change merged 2026-03-31 [B · S18; A · S16] |
| v1.7.0 | mid-2026 | Full 2026-07-28 support; negotiates the highest mutual version and keeps compatibility with 2025-11-25 and earlier; HTTP handler accepts the new protocol only with `Stateless = true`; adds `MCPGODEBUG` escape hatches for behaviour changes [A · S13] |
| v1.8.0 | 2026-09-04 | Transport hardening, bounded decoding, `ServerOptions.SupportedProtocolVersions`, cache controls; more `MCPGODEBUG` flags [A · S12; B · S14; C · S19] |
| v1.9.0 | not yet released | Scheduled to remove all escape-hatch flags [A · S13, S14] |

The README documents a version-compatibility table and the roots/sampling/logging deprecation [A · S11]. GitHub reported its Go MCP server on v1.7.0-pre.3 serving more than half a million users [A · S13].

**Client adoption of the new revision.** Claude Code's newer runtime asks HTTP servers about 2026-07-28 but, by default, keeps stdio servers on the earlier handshake unless negotiation is set to `auto` [A · S22]. The AI SDK's stdio client probes `server/discover` and falls back to legacy `initialize` [B · S35]. Cursor publishes no protocol revision [B · S32] and was reported still requesting 2025-11-25 [C · S34]. VS Code, Antigravity and Claude Desktop: unknown [D]. For a stdio server, the legacy handshake is therefore the path that matters first, and both must work.

### 5.2 Multi-step workflow servers in the wild (census)

| Server | Lang | State | Pattern | Notable facts |
|---|---|---|---|---|
| **GitHub MCP Server** | Go | none (API facade) | Worker/facade with toolsets | Migrated from `mark3labs/mcp-go` to the official SDK [B · S43]; schema snapshot tests ("toolsnaps") [B · S44]; dynamic toolset tools introduced to manage surface size [B · S45]; toolset curation driven by production usage data [B · S43] |
| **Task Master** | Node | `.taskmaster/` (`tasks.json`, config) | Guide + LLM-calling worker | 36 tools ≈ 21k tokens; 15-tool "standard" ≈ 10k; 7-tool "core" ≈ 5k [B · S46] |
| **Shrimp Task Manager** | Node | `DATA_DIR` JSON files | Guide via prompt templates | About 15 tools whose responses steer the next step; templates customisable [B · S47]; an external audit found descriptions lacking usage guidance [C · S42] |
| **Beads** (+ `beads-mcp`) | Go core; Python MCP wrapper | `.beads/` JSONL + SQLite, later Dolt | DAG state oracle ("ready" work); CLI-first | Per-call `workspace_root`, canonical routing, `set_context` kept for compatibility; misrouting incidents; root-level `$ref` in an output schema broke tool loading in Claude Code until fixed; per-project MCP instances discouraged [B · S48] |
| **MCP Task Orchestrator** | not verified | work-item graph (store not verified) | Guide + gates | 13–14 tools; advancing an item fails when required notes are missing; "server owns guardrails, agent owns the rest" [C · S49] |
| **Theodosia / deploy-gate-agent** | Python | Burr state machine | Enforce via one step tool | Out-of-order calls return `invalid_transition` plus reachable actions; agent self-corrects [C · S50] |
| **txn2/mcp-data-platform** | Go (official SDK) | platform | Guide via instructions; mandatory opening sequence | 28 tools; 21 unannotated; three write-confirmation prompts before first result; proposes a table test over registered tools [B · S53] |
| **Anthropic filesystem server** | TypeScript | none | Facade | Two path-confinement CVEs (prefix check, symlink) [B · S55] |

**Lessons distilled**

1. Workflow servers cluster around *state + next + validate* (Guide); Worker style dominates API facades [D, from the census].
2. Project-local file state is the norm [B · S46, S47, S48].
3. Surface size is a real cost; ship profiles or stay tiny [B · S46; A · S22].
4. Explicit workspace routing, with the resolved path echoed back, prevents cross-repo damage [B · S48].
5. Schema-hygiene bugs can make tools vanish silently — Claude Code drops tools whose schemas are invalid and flattens root-level combinators [A · S22; B · S48].
6. Annotation discipline needs a regression test, not good intentions [B · S53].
7. Snapshot tool schemas in CI [B · S44].
8. A CLI plus AGENTS.md is a serious alternative for shell-capable agents [B · S48].
9. Enforcement belongs on the server side of the trust boundary [C · S49, S50, S51, S52].

*Caveats.* Survivorship bias (the census was built by searching for guide-like servers); several entries are single-author reports; the Task Orchestrator entry is a directory listing, not a code review [C]. GitHub's dynamic-toolset design relies on per-session list changes, which 2026-07-28 removes — how GitHub reconciles this was not verified [D].

### 5.3 State management (Q1)

| Option | Survives restart | Safe with several processes | Inspectable / git-friendly | Fits sessionless spec | Complexity | Verdict |
|---|---|---|---|---|---|---|
| Server memory | No | No | No | No — needs handles anyway [A · S1] | Low | Reject |
| **Files: one authoritative `session.json`, lock + atomic replace + `rev`** | Yes | Yes (lock) | Yes | Yes | Low–Med | **Recommend** |
| Files + derived SQLite index | Yes | Yes | Yes (index derived) | Yes | Med | Optional later |
| SQLite as authority | Yes | Yes | Opaque binary | Yes | Med | Defer |
| Dolt / remote DB | Yes | Yes | SQL diffs | Yes | High | Reject for v1 |

Evidence and reasoning:

- The spec's direction makes in-connection state a dead end [A · S1]; stdio also means one server process per client session, so two clients on one repo means two processes [A · S22; D].
- Precedent: Task Master, Shrimp and Beads all keep project-local files [B · S46, S47, S48].
- Beads hazards: MCP-layer misrouting across repositories, a global daemon removed to stop cross-project pollution, early JSONL + SQLite versions with data-duplication and data-loss warnings, and an eventual move to Dolt [B · S48]. Lessons: one authority, explicit routing, no ambient "current project".
- Go primitives: `os.Root` (Go 1.24) confines file access and defeats symlink escapes [A · S20]; Go 1.25 added `Rename`, `WriteFile`, `MkdirAll`, `RemoveAll` and more [B · S21]. Helper packages written for Go 1.24 note the missing methods, so pin Go ≥ 1.25 [C · search result for `osroot` helpers].
- Locking and atomic-write libraries — `gofrs/flock`, `google/renameio/v2`, and the Go toolchain's `lockedfile` (public copy in `rogpeppe/go-internal`) — are candidates from prior knowledge and were **not re-verified** this session [D]. Advisory locks behave differently on network filesystems and Windows [D]; add a lock timeout and a `locked` (retryable) error.
- Hand edits and corruption: refuse to operate on a file that fails schema validation and offer `vivechak doctor`; never silently repair [D].
- Location: the workspace, so state travels with the repo; offer `--state-dir` for research spanning repositories [D].

### 5.4 MCP primitives (Q2, Q3)

| Primitive | Control model | Use in Vivechak | Client evidence | Caveats |
|---|---|---|---|---|
| **Tools** | Model-controlled [A · S4] | All state changes and every action the agent must take | Universal | Tool results are the only channel guaranteed to reach the model everywhere |
| **Resources** | Application-driven [A · S4] | Read-only mirrors: guide, session status, step spec, latest report | Claude Code (@-mention plus auto list/read tools) [A · S22]; Cursor [A · S23]; VS Code (Add Context) [A · S24]; Claude Desktop [C · S64]; Antigravity unverified [D] | Tool-returned `resource_link`s need not appear in `resources/list` [A · S5] |
| **Prompts** | User-controlled [A · S4] | Three slash-command shortcuts | Claude Code `/mcp__server__prompt` [A · S22]; Cursor [A · S23]; VS Code `/server.prompt` [A · S24]; Antigravity reported unsupported [C · S31] | Never carry logic |
| **Server `instructions`** | Injected by client | Process narrative, ≤ ~1.2 KB | Claude Code yes [A · S22]; VS Code lists it as a supported feature [A · S25]; claude.ai/Desktop reported no as of Mar–Apr 2026 [C · S33]; Cursor and Antigravity unverified [D] | Front-load; verify with a canary in each client |

Additional constraints from 2026-07-28: list results carry cache hints; lists must not vary per connection (except by authorization) [A · S1, S3]; resource-not-found uses the invalid-params code [A · S1]. Separately, since 2025-11-25, input-validation failures are tool-execution errors under **SEP-1303** [A · S6]; long-running work is covered by the **SEP-2663** Tasks extension, Final status, but absent from the Go SDK (§5.7) [A · S7; A · S12]. Structured tool output should also be serialised as text for older clients — the Go SDK does this automatically when `Content` is unset [A · S5, S15].

**Decision rule.** Needs to *act* or *change state* → Tool. *Large, stable, read-only* and worth attaching → Resource plus a link returned by a tool. A *human* should trigger a canned start → Prompt. The agent needs to *know the process* → instructions + AGENTS.md + Skill, not Prompt or Resource.

### 5.5 Tool naming, description and discoverability (Q5)

**What the clients do to your tools**

- Claude Code: tool search is on by default; only tool names and server instructions load at session start; descriptions and instructions are cut at 2,048 characters; `alwaysLoad` and `anthropic/alwaysLoad` opt tools out of deferral; tools with invalid schemas are excluded, root-level combinators flattened, and property names limited to 1–64 characters of ASCII letters, digits, `_`, `.` and `-` [A · S22].
- claude.ai reportedly truncates descriptions near 500 characters and, with Claude Desktop, ignored `instructions` as of Mar–Apr 2026 [C · S33]. Re-test; these may have changed.
- VS Code shows the description in the tool picker and confirmation dialog, and confirms every tool without `readOnlyHint` [A · S25].
- Cursor asks for approval by default [A · S23]. Its 40-tool cap is repeated in 2025 forum threads but not in current docs [B · S32]; a March 2026 forum question asks whether it became 80 [C · S32 context].
- Antigravity reports of a tool cap conflict (100 vs 512/800 by version) [C · S30].

**Design guidance and evidence**

- Anthropic: build a few thoughtful tools around high-impact workflows; consolidate chained operations; namespace by resource and test prefix vs suffix; return meaningful, high-signal context and support `concise|detailed`; make errors actionable; write descriptions as for a new hire; iterate with evals [A · S36]. Claude docs add at least 3–4 sentences per description and consolidation via an `action` parameter where sensible [A · S37].
- Block: design top-down from the workflow, not bottom-up from endpoints; where chaining is unavoidable, spell out steps and dependencies and keep intermediate outputs concise [B · S38].
- Phil Schmid: treat MCP as a user interface for agents; pair it with Skills that teach when and how to combine tools [B · S39].
- Vercel (Next.js doc-access evals, four conditions): no-docs baseline 53%; an available Skill with default invocation also 53%, because it went uninvoked in 56% of runs (and underperformed baseline on some sub-metrics, suggesting an unused skill can add noise); explicit "use this skill" instructions raised the trigger rate to 95%+ but pass rate to only 79%, and were sensitive to exact wording; a compressed, always-loaded AGENTS.md index reached 100% [B · S40]. The trigger step is the weak link — the same step an MCP guide depends on.

**Numbers to design against**

| Constraint | Value | Source |
|---|---|---|
| Claude Code description / instruction cut-off | 2,048 chars | [A · S22] |
| claude.ai description cut-off (claimed) | ~500 chars | [C · S33] |
| Claude Code tool-result warning / default cap | 10k / 25k tokens (raisable via `_meta` up to 500k chars) | [A · S22] |
| Tool-definition cost, Task Master | 36 tools ≈ 21k tokens; 7 tools ≈ 5k | [B · S46] |
| Tool caps | Cursor folklore 40 (undocumented); Antigravity 100–800; VS Code cap unverified | [B · S32; C · S30; D] |

Conclusion: about seven tools, 3–4-sentence descriptions with the decisive content first, and three always-on guidance channels (§3.5).

### 5.6 Error handling and validation (Q5, Q6)

- **Spec.** Two error kinds; input-validation failures are tool-execution errors (**SEP-1303**), which give the model feedback to self-correct and retry [A · S5, S6].
- **Go SDK contract.** Invalid input is rejected before the handler runs; the output schema is inferred from `Out`; `Out` populates `StructuredContent`; unset `Content` is auto-filled with the JSON; a returned Go error becomes an `isError` result rather than a protocol error [A · S15]. PR #863 (merged 2026-03-31) makes schema-validation failures tool results [A · S16]. A May 2026 Go-focused article still describes them as protocol errors [C · S60] — treat as stale or version-specific and pin a regression test (§5.13).
- **Gotchas.** Zero values are validated on error paths, so a map field without `omitempty` can turn every error return into an output-validation failure [C · S19]. `json.RawMessage` is inferred as a byte array [C · S19]. Custom marshallers can panic at registration [C · S19]. Generated object schemas forbid extra properties, so an unexpected argument fails validation [B · S17]. Some strict clients validate `structuredContent` against `outputSchema` even for error results [C · S61].
- **Guidance value of errors.** Actionable messages steer agents; opaque codes and tracebacks do not [A · S36].
- **Contract:** §3.8. Example payload:

```json
{
  "ok": false,
  "error": {
    "code": "precondition_failed",
    "message": "Step n3-analysis is pending: n2-sources is not done.",
    "remediation": "Finish and submit n2-sources first, or call session_status to see what is ready.",
    "allowed_next": ["session_status", "step_submit"],
    "retryable": false
  },
  "state": {"session_id": "s-0924-a3f9", "rev": 12},
  "workspace": "/home/me/proj"
}
```

- **Approval friction.** VS Code confirms every tool not marked `readOnlyHint` [A · S25]. A Go workflow-gate server's session-opening sequence produced three write prompts until its read tools were annotated; the maintainer also noted that approval memory is per conversation on at least one major client and that constant prompts teach users to approve reflexively [B · S53]. Annotations are untrusted hints and never replace server-side checks [A · S5; B · S35].
- **Size.** Keep responses under about 10k tokens; use summary + `resource_link` beyond that [A · S22].

### 5.7 Long-running operations (Q4)

- **Protocol.** Progress notifications and cancellation exist in every era. Tasks is now a Final extension (**SEP-2663**) with polling via `tasks/get` [A · S1, S7].
- **Go SDK.** The v1.8.0 API index has no Tasks types, and no third-party Go runtime for the extension was identified in this research [A · S12]. Sources disagree on whether *any* SDK ships a Tasks runtime at all: a FastMCP maintainer's notes say none does, while the C# SDK's own documentation describes a package implementing SEP-2663 [B/A · S54] (§5.13) — for Go specifically, the answer is no.
- **Client behaviour.** Claude Code: per-server timeout and `MCP_TOOL_TIMEOUT` settings; idle timeouts of 5 minutes (HTTP) and 30 minutes (stdio) that progress notifications reset; calls exceeding two minutes are auto-backgrounded [A · S22]. Cursor, VS Code, Antigravity and Claude Desktop timeouts: unverified [D].
- **Pattern for Vivechak.** (1) Make validation deterministic and fast — parsing megabyte-scale Markdown in Go should be sub-second, but measure [D]. (2) Anything slower becomes a resumable, checkpointed tool that persists per-item results, reports progress, is idempotent by (artifact hash, check id) and returns `complete:false` with a resume hint. (3) Expose the last report as a resource. (4) Do not build on Tasks yet.

### 5.8 "Guide" vs "Worker" validation (Q7)

**Search for disconfirming evidence.** I looked specifically for: (a) reports of guide-style servers failing or being ignored; (b) guidance that argues for consolidation instead of step-by-step tools; (c) friction costs of chatty protocols; (d) protocol features that break phase-gated designs; (e) alternatives to MCP for the same job; (f) evidence that models discount instructions embedded in tool results. Findings (a)–(e) produced real qualifiers; (f) produced **no direct evidence** in either direction.

**Evidence for B′ (Guide + enforcement)**

| # | Finding | Grade | Design consequence |
|---|---|---|---|
| 1 | Models skipped, merged and reordered steps of a ten-step workflow even with emphatic wording; moving the sequence into a stateful MCP tool fixed it | C · S51 | Enforce transitions server-side |
| 2 | Enforcement-first servers refuse out-of-order calls with the legal moves, or block advancing until required notes exist | C · S49, S50 | Structured refusal with `allowed_next` |
| 3 | Prompt-only orchestration degrades as the context fills | C · S52 | Keep state outside the context; `session_status` re-orients |
| 4 | The "ready work" query over a dependency graph is the core primitive of a widely used agent issue tracker, and the graph persists across sessions | B · S48 | `session_status` returns ready steps and next actions |
| 5 | Workflow-shaped, consolidated tools are the recommended design | A · S36; B · S38 | `step_submit` = validate + persist + advance |
| 6 | Multi-step orchestration remains a persistent weakness across 20 LLMs; scores vary with the number of servers and dependency mix | B · S59 | Lower the orchestration burden on the model |

**Disconfirming evidence and qualifiers**

| # | Finding | Grade | Design consequence |
|---|---|---|---|
| 1 | *Trigger fragility.* An unprompted Skill went uninvoked in 56% of runs (53% pass, no better than a no-docs baseline); an explicit trigger instruction lifted invocation to 95%+ but pass rate to only 79% and was wording-sensitive; always-loaded AGENTS.md hit 100%. Claude Code defers MCP tool loading by default; `instructions` reported ignored in claude.ai/Desktop | B · S40; A · S22; C · S33 | Never rely on the agent choosing to call the guide; use always-on channels; if using a Skill, give it an explicit trigger line and expect it to under-perform AGENTS.md |
| 2 | *Approval friction.* VS Code confirms every non-read-only tool; a Go workflow-gate server produced three prompts before the first result | A · S25; B · S53 | Annotate every tool; consolidate; keep reads read-only |
| 3 | *Round trips.* Guidance favours fewer, coarser calls over chains | A · S36 | No `step_start`/`step_finish` pairs; return next actions inline |
| 4 | *Per-phase tool exposure is barred or weakly supported* | A · S3; B · S32 | Static tool list; gate inside handlers |
| 5 | *CLI + AGENTS.md alternative.* Beads has advised shell use over MCP for agents with a shell | B · S48 | Dual front-end; keep Passive as the eval control arm |
| 6 | *Sampling is deprecated*, so a Worker that borrows the client's LLM is a dead end | A · S1 | No LLM-dependent worker steps |
| 7 | *Enforcement can trap the agent* (dead ends, unreachable states, unhelpful refusals) | D | `allowed_next`, `stale` recovery, a human-run out-of-band `override` command |
| 8 | *Guide text rides the channel used by indirect prompt injection*; clients or models may discount instruction-like tool output | D | Keep in-result guidance short and factual; duplicate the process in always-on channels; measure |
| 9 | *Survivorship bias and single-author sources* in the census | D | Run your own evals |

**Steelman — Worker.** Deterministic execution reduces agent decisions and variance. But judgement steps cannot be executed without an LLM, sampling is deprecated, and LLM-calling servers are out of scope; the mechanical parts are already folded into `step_submit`.

**Steelman — Passive.** Cheapest to build, no protocol dependency, works in any client with a shell, and Skills plus AGENTS.md are now a cross-client standard [C · S41]. It gives up enforcement, deterministic validation and cross-session state — exactly the properties the practitioner reports value [C · S51].

**Verdict.** B′ is best supported, with three conditions attached: enforcement with escape hatches, always-on guidance, and evidence from your own evals. The Passive arm should be built anyway — it is nearly free once the core exists.

**What would falsify B′:** Passive + Skill matching B′ on completion and error rates across target models; enforcement causing dead-ends above an agreed rate; or approval friction leading users to disable the server.

### 5.9 Go SDK patterns and idioms

Snippets are illustrative and were composed from the API surface documented in S12/S15; compile against your pinned version.

**Bootstrap (stdio; stdout reserved for JSON-RPC)**

```go
func newServer(c *core.Core) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{Name: "vivechak", Version: version},
		&mcp.ServerOptions{Instructions: guide.Instructions}, // ≤ ~1.2 KB, see §3.5
	)
	mcpx.RegisterTools(s, c)     // ~7 typed tools, all via mcpx.Add
	mcpx.RegisterResources(s, c) // resource templates
	mcpx.RegisterPrompts(s, c)   // start / resume / review
	return s
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil))) // never stdout
	c, err := core.Open(resolveWorkspace())
	if err != nil { slog.Error("open", "err", err); os.Exit(1) }
	if err := newServer(c).Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		slog.Error("server exited", "err", err); os.Exit(1)
	}
}
```

**One registration path that enforces surface invariants**

```go
func Add[In, Out any](s *mcp.Server, t *mcp.Tool, h mcp.ToolHandlerFor[In, Out]) {
	switch {
	case t.Annotations == nil:
		panic(t.Name + ": annotations required (set ReadOnlyHint explicitly)")
	case len(t.Description) < 200 || len(t.Description) > 1500:
		panic(t.Name + ": description must be 200–1500 chars")
	case !toolName.MatchString(t.Name): // ^[a-z][a-z0-9]*(_[a-z0-9]+)+$
		panic(t.Name + ": use resource_verb snake_case")
	}
	mcp.AddTool(s, t, h) // infers schemas, validates input, packs errors
	Registered = append(Registered, t)
}
```

**A consolidated, typed tool (validate + persist + advance)**

```go
type StepSubmitIn struct {
	SessionID     string   `json:"session_id" jsonschema:"session id from session_start or session_list"`
	StepID        string   `json:"step_id" jsonschema:"step id from session_status, e.g. n2-sources"`
	ArtifactPaths []string `json:"artifact_paths" jsonschema:"workspace-relative paths of the files you wrote for this step"`
	ExpectedRev   *int     `json:"expected_rev,omitempty" jsonschema:"rev from your last session_status; omit to skip the stale check"`
	WorkspaceRoot string   `json:"workspace_root,omitempty" jsonschema:"only if the server was not launched inside the target workspace"`
}

type StepSubmitOut struct {
	Status      string       `json:"status" jsonschema:"done or needs_revision"`
	Rev         int          `json:"rev"`
	Workspace   string       `json:"workspace"`
	Issues      []Issue      `json:"issues,omitempty"`      // omitempty: zero Out must stay schema-valid
	NextActions []NextAction `json:"next_actions,omitempty"`
	ReportURI   string       `json:"report_uri,omitempty"`
}

mcpx.Add(s, &mcp.Tool{
	Name:        "step_submit",
	Title:       "Submit a research step for validation",
	Description: stepSubmitDescription, // see §3.7 template
	Annotations: &mcp.ToolAnnotations{IdempotentHint: true, ReadOnlyHint: false},
	// Set DestructiveHint explicitly false for additive writes; the spec default is true (verify field type on compile).
}, func(ctx context.Context, req *mcp.CallToolRequest, in StepSubmitIn) (*mcp.CallToolResult, StepSubmitOut, error) {
	res, err := c.SubmitStep(ctx, in.toCore()) // lock → validate → persist → advance
	var pe *core.PreconditionError
	switch {
	case errors.As(err, &pe):
		return toolError(pe), StepSubmitOut{}, nil // isError: true, text-first
	case err != nil:
		return nil, StepSubmitOut{}, err // SDK converts to an isError result
	}
	return summaryResult(res), fromCore(res), nil // content: concise text; structured: full data
})
```

**Progress for slow checks**

```go
if tok := req.Params.GetProgressToken(); tok != nil {
	_ = req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
		ProgressToken: tok, Progress: float64(i), Total: float64(n), Message: "checking sources",
	})
}
```

**Idioms and policy**

- Typed handlers (`mcp.AddTool` with `In`/`Out` structs) give schema inference, validation and structured output for free [A · S12, S15]; keep `Out` types plain (no `json.RawMessage`, no custom marshallers, `omitempty` on maps and slices) [C · S19].
- Return decision-critical text in `Content`; use `StructuredContent` for programmatic clients [A · S5, S15].
- Reuse the SDK's JSON Schema library to validate structured research artifacts against per-step schemas [D].
- Pin `v1.8.x`, run CI against both protocol eras, and treat any `MCPGODEBUG` flag as a temporary crutch — all are slated for removal in v1.9.0 [A · S13, S14]. `ServerOptions.SupportedProtocolVersions` lets you restrict what is advertised [C · S19].
- For HTTP later: `StreamableHTTPOptions{Stateless: true}` for the new protocol [A · S13]; a schema cache can be shared across servers in stateless mode [B · S12 example via third-party wrapper; D].

### 5.10 Client discovery and configuration (Q8)

| Client | Where config lives | Root key | Workspace signal | Install UX | Notes |
|---|---|---|---|---|---|
| **Claude Code** | Project `.mcp.json` (approval required), local and user scopes; `claude mcp add` / `add-json` | `mcpServers` | `CLAUDE_PROJECT_DIR` in the server's env; roots answered with launch dir; `${VAR:-default}` expansion | CLI; plugin-bundled servers | Prompts as `/mcp__server__prompt`; resources via `@`; tool search default; honours `instructions` [A · S22] |
| **Cursor** | `.cursor/mcp.json` (project), `~/.cursor/mcp.json` (global) | `mcpServers` | `${workspaceFolder}` and other variables in command, args, env, URL, headers | Deeplinks, marketplace | Tools, Prompts, Resources, Roots, Elicitation, Apps; approval by default [A · S23] |
| **VS Code** | `.vscode/mcp.json` (`servers`); portable `.mcp.json` (`mcpServers`); user-profile `mcp.json`; auto-discovery of other tools' configs | `servers` / `mcpServers` | `${workspaceFolder}` (used in sandbox examples) | UI, `code --add-mcp`, gallery | Resources via *Add Context*; prompts `/server.prompt`; stdio sandbox on macOS/Linux; workspace trust; confirms non-read-only tools [A · S24, S25] |
| **Antigravity** | `.agents/mcp_config.json` (workspace); `~/.gemini/config/mcp_config.json` (global; an earlier path was reported in May 2026) | `mcpServers`; remote entries use `serverUrl` | Not documented | "Manage MCP Servers" UI | Tool cap reports conflict; prompts reported unsupported; skills native [B · S29; C · S30, S31] |
| **Claude Desktop** | `claude_desktop_config.json`; `.mcpb` bundles with a manifest and user-config prompts | `mcpServers` | None — no project concept | Settings; bundle install | `instructions` reported unread; primitives beyond tools unverified [A · S26, S27, S28; C · S29, S33, S64] |

Starter files (server name `vivechak`; the binary reads `CLAUDE_PROJECT_DIR` itself):

```jsonc
// .mcp.json  (Claude Code; VS Code also reads this portable format)
{ "mcpServers": { "vivechak": { "type": "stdio", "command": "vivechak", "args": ["mcp"] } } }

// .cursor/mcp.json
{ "mcpServers": { "vivechak": { "command": "vivechak", "args": ["mcp", "--workspace", "${workspaceFolder}"] } } }

// .vscode/mcp.json
{ "servers": { "vivechak": { "type": "stdio", "command": "vivechak", "args": ["mcp", "--workspace", "${workspaceFolder}"] } } }

// .agents/mcp_config.json  (Antigravity — interpolation undocumented; verify cwd in Spike 0)
{ "mcpServers": { "vivechak": { "command": "vivechak", "args": ["mcp"] } } }

// claude_desktop_config.json  (no workspace concept → explicit allow-list)
{ "mcpServers": { "vivechak": { "command": "/usr/local/bin/vivechak", "args": ["mcp", "--allow-root", "/Users/me/research"] } } }
```

**Notes.** Project-scoped Claude Code entries require user approval [A · S22]. VS Code can pick up configs written for other tools [A · S24]. MCPB bundles can prompt the user for configuration values, which is the natural place to collect a Claude Desktop workspace directory [A · S27]. Third-party matrices that contradict vendor docs (e.g. Cursor lacking Resources/Prompts) should be discounted [D · S58 vs A · S23].

### 5.11 Testing MCP servers (Q6)

| Layer | What it proves | Tools | Evidence |
|---|---|---|---|
| Core | Transitions, invariants, crash safety, contention | table tests; property tests (e.g. `pgregory.net/rapid` [D]); fault-injection hook in the store; `go test -race` | [D] |
| In-memory MCP | Every tool, every failure class, both protocol eras | `mcp.NewInMemoryTransports`, real `Client` | [A · S12] |
| Tool-surface snapshot | No accidental schema/description drift | golden JSON per tool, explicit update flag, CI fails on missing snapshot | [B · S44]; diff tooling [C · S57] |
| Registration invariants | Annotations, description band, name pattern, root schema object, no root `$ref` | `mcpx.Add` panics; test iterates `Registered` | [B · S53; A · S22; B · S48] |
| Process / stdio | Real binary speaks clean JSON-RPC | `mcp.CommandTransport{Command: exec.Command(bin, "mcp")}` | [A · S11] |
| Inspector CLI | Independent client sees the same tools | `npx @modelcontextprotocol/inspector --cli …` (flags vary by version) | [A · S8; B · S56] |
| Conformance | Protocol behaviour over HTTP | `@modelcontextprotocol/conformance` against a server URL; baseline of expected failures | [B · S9] |
| Agent evals | The agent follows the pipeline | multi-call tasks, verifiable outcomes, tool-call / error / token metrics | [A · S36] |

```go
func connect(t *testing.T, s *mcp.Server, proto string) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()
	if _, err := s.Connect(ctx, st, nil); err != nil { t.Fatal(err) }
	var opts *mcp.ClientSessionOptions
	if proto != "" { opts = &mcp.ClientSessionOptions{ProtocolVersion: proto} }
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).Connect(ctx, ct, opts)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { cs.Close() })
	return cs
}

func TestStepSubmit_OutOfOrderIsAToolError(t *testing.T) {
	for _, proto := range []string{"", "2025-11-25"} { // newest era, then legacy handshake
		t.Run("proto="+proto, func(t *testing.T) {
			cs := connect(t, newTestServer(t), proto)
			res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
				Name: "step_submit",
				Arguments: map[string]any{"session_id": sid, "step_id": "n3-analysis", "artifact_paths": []string{"research/n3.md"}},
			})
			if err != nil { t.Fatalf("protocol error, want tool error: %v", err) }
			if !res.IsError { t.Fatal("want isError=true") }
			text := res.Content[0].(*mcp.TextContent).Text
			for _, want := range []string{"precondition_failed", "allowed_next"} {
				if !strings.Contains(text, want) { t.Errorf("missing %q in %q", want, text) }
			}
		})
	}
}
```

Also test: malformed arguments (schema violation) returns a tool error, not a protocol error — guarding the §5.13 disagreement [A · S16 vs C · S60]; two runs of `tools/list` return identical order [A · S1]; the process test fails if anything other than JSON-RPC reaches stdout [A · S11].

**Agent-eval design** (Anthropic's method [A · S36]). 12–20 tasks; success measured by final workspace state, not by transcript; record calls, errors, tokens and wall-clock; hold out a test split so descriptions are not overfit; re-run per target model and client. Conditions: tools only · +instructions · +AGENTS.md +Skill · Passive control. Adversarial cases: skip-ahead, invalid artifact, stale `rev`, two agents, context loss, path traversal, injected instructions in an artifact.

### 5.12 Security and trust boundaries

- **Path confinement is the proven failure class, and it recurs.** Anthropic's filesystem server needed CVE fixes for a bypassable directory-prefix check and a symlink escape (CVE-2025-53109, CVSS 8.4; CVE-2025-53110, CVSS 7.3; fixed in the npm package's 2025.7.1) [B · S55]. Separately, Anthropic's official Git MCP server — "the canonical Git MCP server, the one developers are expected to copy," as one researcher put it — shipped three more path-related flaws disclosed in January 2026: an unconstrained `git_init` path argument, a repository-boundary check that didn't actually enforce its own restriction, and a `git_diff`-based primitive that could overwrite or empty arbitrary files (CVE-2025-68143/-68144/-68145; fixed across the 2025.9.25 and 2025.12.18 releases) [B · S55]. In Go, open `os.Root` at the resolved workspace and never build paths by string concatenation [A · S20; B · S21].
- **Untrusted content.** Claude Code warns that servers which fetch external content can expose users to prompt injection [A · S22]. Research artifacts are such content; excerpt and fence them, never execute them, and never let them alter state outside validators [D].
- **Annotations are hints**, not authorisation; clients must treat them as untrusted unless the server is trusted [A · S5; B · S35].
- **Transport.** stdio: keep stdout clean. HTTP, if ever added: bind localhost and configure DNS-rebinding protection explicitly. Two concrete precedents argue against assuming safe defaults: the MCP Inspector's proxy had no session token or origin check before v0.14.1, letting a public webpage reach a local Inspector instance via `0.0.0.0` and execute code (CVE-2025-49596, CVSS 9.4, disclosed by Oligo Security, fixed June 2025) [B · S56]; and the Python MCP SDK shipped with DNS-rebinding protection off by default for unauthenticated localhost HTTP servers until v1.23.0 (CVE-2025-66414) [B · S56]. A secondary source additionally names the Go SDK among historically-affected implementations, but that specific claim was not independently confirmed against a primary advisory in this pass [C · S56] — verify the pinned Go SDK version's HTTP defaults directly before relying on them.
- **Tooling.** Keep MCP Inspector patched to at least v0.14.1 for the reason above [B · S56].
- **Drift and rug-pulls.** Snapshot tests and signed releases make silent description changes visible [C · S57].
- **Human gates.** Where a human must confirm an override, Claude Code supports a per-tool interaction requirement [A · S22]; elsewhere use an out-of-band CLI command rather than relying on elicitation, whose support is uneven [A · S23; C · S64].

### 5.13 Disagreements register

| Topic | Position 1 | Position 2 | Resolution here | Confidence |
|---|---|---|---|---|
| Cursor primitives | Own docs: Tools, Prompts, Resources, Roots, Elicitation, Apps [A · S23] | Third-party matrix: no Resources/Prompts [D · S58] | Trust the vendor; verify in Spike 0 | High |
| Schema-validation failures in Go | Tool result via PR #863 [A · S16] | "Protocol error" per a May 2026 article [C · S60] | Trust the merged PR; add regression test for the pinned version | Med-High |
| `isError` for failed content validation | Spec lists business-logic failures as tool errors [A · S5]; practitioner rule of thumb [C · S60] | Strict clients may validate structured payloads on errors [C · S61] | `needs_revision` as a normal result; A/B in evals | Low-Med |
| Tasks runtime availability | "No SDK ships a runtime" — FastMCP maintainer notes [B · S54] | A C# SDK package implementing SEP-2663 is documented [A · S54] | Absent for Go specifically either way | Med |
| Tool-count caps | Cursor 40 (2025 forums) [B · S32] | Possibly 80 (Mar 2026 question) [C · S66]; Antigravity 100 vs 512/800 [C · S30] | Moot at ~7 tools; keep the budget test | High |
| `instructions` field | Loaded by Claude Code [A · S22] | Ignored by claude.ai/Desktop (Mar–Apr 2026) [C · S33] | Use three channels; canary per client | Med |
| Beads on MCP vs CLI | Earlier versions advised shell over MCP [B · S48] | Later README documents a single MCP server with per-call `workspace_root` [B · S48] | Dual front-end | Med |
| Guide compliance | Prose steering is enough | Prose gets skipped; enforce [C · S51] | Enforce plus always-on prose | Med |
| Dynamic tool lists | 2026-07-28: lists MUST NOT vary per connection except by authorization [A · S3] | VS Code's own docs advertise "dynamic tool discovery" varying tools by workspace/prompt context [A · S25] | Unresolved; Vivechak avoids depending on list-changed either way | Low |

## 6. Open Questions & Risks

### 6.1 Risk register

Likelihood (L) and impact (I) are this report's own estimates [D].

| ID | Risk | L | I | Mitigation | Next step |
|---|---|---|---|---|---|
| R1 | Clients ignore instructions/AGENTS.md/Skill, so the Guide is never invoked | Med | High | Three channels; `next_actions` in every result; `alwaysLoad`; canary | Spike 0 |
| R2 | Enforcement traps an agent in a dead end | Med | High | `allowed_next`; `stale` recovery; human-run `override`; `doctor` | Eval scenario; design override |
| R3 | State corruption from concurrent processes, crashes or hand edits | Low-Med | High | Lock + atomic replace + `rev`; schema validation; crash tests | Fault-injection suite |
| R4 | Writes land in the wrong workspace | Med | High | Allow-list; echo resolved path; `os.Root`; no ambient "current project" | Test with two workspaces |
| R5 | Approval fatigue leads users to disable or blanket-approve | Med | Med | Annotate all tools; consolidate; keep reads read-only | Measure prompts per session |
| R6 | Protocol or SDK churn (2026-07-28 rollout; v1.9.0 flag removal) | High | Med | Dual-era tests; pin; upgrade cadence | CI matrix |
| R7 | Schema or description rules make tools vanish or truncate on some client | Med | High | Registration invariants; Inspector; per-client smoke test | Spike 0 |
| R8 | Prompt injection via research artifacts steers the agent | Med | Med-High | Excerpt/fence; no exec; validators only; eval case | Add adversarial evals |
| R9 | Description or guidance drift, including a malicious update | Low | Med | Snapshot tests; signed releases; pinned binary | CI |
| R10 | Tool-definition and response bloat erode context | Med | Med | Token-budget test; `concise` default; resource links | Budget assertion |
| R11 | Vivechak semantics unknown (who edits the DAG, artifact formats, multi-writer) | High | Med | Design review with the Vivechak owner | See §6.4 |
| R12 | No controlled evidence on Guide vs Passive | High | Med | Own evals with a Passive control arm | Build harness in P3 |

### 6.2 Discovered Concerns (not in the brief)

- **DC-1 — The spec went stateless.** In-connection state and per-phase tool lists are incompatible with 2026-07-28; the whole design must keep state external and the tool list static [A · S1, S3].
- **DC-2 — The discovery trigger is the weakest link.** Deferred tool loading, ignored `instructions` and optional skills all reduce the chance the agent ever starts the workflow [A · S22; B · S40; C · S33].
- **DC-3 — Three "obvious" mechanisms are deprecated.** Roots (workspace discovery), sampling (server-side LLM) and the logging capability are all deprecated [A · S1].
- **DC-4 — Path confinement.** Symlinks and prefix checks have bitten Anthropic's own reference servers five separate times across two servers (filesystem, then Git) [B · S55]; use `os.Root`.
- **DC-5 — Schema hygiene can silently remove tools.** Claude Code drops tools with invalid schemas and flattens root combinators; Beads shipped a root-level `$ref` that broke loading [A · S22; B · S48]. Whether the Go SDK's inference can emit a root `$ref` for recursive types was **not verified** [D].
- **DC-6 — Multi-workspace misrouting** in stateful MCP servers [B · S48].
- **DC-7 — Approval friction and annotation drift** [A · S25; B · S53].
- **DC-8 — SDK churn and error-path gotchas** [A · S13; C · S19].
- **DC-9 — Opportunity: Agent Skills + AGENTS.md** as a cross-client procedural layer that partly overlaps the Guide. Evidence favours AGENTS.md for passive framework literacy, but Vercel's own reading is that skills still win for narrow, explicitly-triggered vertical workflows — plausibly a good description of "run the Vivechak pipeline" if the trigger is made explicit rather than left to the model's discretion [B · S40; C · S41].
- **DC-10 — Governance of DAG mutation.** Who may add or reorder steps, and how does a human override without elicitation? Undecided [D].
- **DC-11 — Opportunity: CLI-first fallback** reusing the same core [B · S48].
- **DC-12 — Agent evals are the only real test of B′**, and need a harness per model and client [A · S36].

### 6.3 Unverified items (resolve in Spike 0 or by reading source)

1. `instructions` handling in Cursor, VS Code and Antigravity; current state in claude.ai and Claude Desktop.
2. Antigravity: Resources support, variable interpolation, working directory, tool cap, canonical global config path.
3. Claude Desktop: Prompts and Resources support; how workspace is conveyed.
4. VS Code: any tool cap; `${workspaceFolder}` expansion inside `.mcp.json`.
5. Timeouts and progress display in every client except Claude Code.
6. Whether Go schema inference can emit a root `$ref` for recursive types in v1.8.
7. Exact field types on `mcp.ToolAnnotations` (`DestructiveHint`, `OpenWorldHint`) and the `Tool` `_meta` field name.
8. Any Tasks-related API in the Go SDK beyond the index reviewed.
9. Locking and atomic-write libraries; Windows and network-filesystem lock semantics.
10. How GitHub's dynamic toolsets behave under 2026-07-28.
11. Per-client Skill directories and auto-discovery rules.
12. Whether `mark3labs/mcp-go` supports 2026-07-28 (only relevant if the official SDK is ever reconsidered).
13. Exact Inspector v2 CLI flags.
14. MCP Registry package types suitable for a Go binary.

### 6.4 Questions for the Vivechak owner (inputs to D-001)

1. May agents propose changes to the DAG (add, split, reorder steps)? Under what approval?
2. What are the artifact formats, and are there existing schemas or linters to reuse as validators?
3. Is multi-user or remote operation ever needed, or is this strictly local-first?
4. Which clients are P0? (This decides how much Antigravity and Claude Desktop uncertainty matters.)
5. Are sessions and reports committed to git, or ignored?
6. Do any validation checks need the network (link liveness, citation lookups)?
7. Where must a human be in the loop (overrides, publishing)?
8. Expected DAG size and expected concurrency (agents per workspace)?

## 7. Sources & Evidence Ledger

**Method note.** Research ran in two passes: an initial broad sweep (official docs, SDK source, client docs, production repos, engineering blogs, security advisories) followed by a targeted verification pass that re-confirmed every load-bearing claim against a live search and recorded the resolving URL below. Grades follow §0's scale (A = normative/first-party primary; B = credible secondary/production evidence; C = single-source practitioner report or unverified vendor claim; D = this report's own inference). Where the verification pass could not re-surface an exact URL for a claim carried from the first pass, that is stated plainly in the note rather than papered over with an invented link — those entries are graded C or D and their claims are treated as correspondingly less certain throughout the report (see §5.13, §6.3).

### Protocol and spec

| ID | Grade | Source | Note |
|---|---|---|---|
| S1 | A | MCP Specification, "2026-07-28" revision — Changelog. `modelcontextprotocol.io/specification/2026-07-28/changelog` (mirrored at `github.com/modelcontextprotocol/modelcontextprotocol/blob/main/docs/specification/2026-07-28/changelog.mdx`) | Primary source for the stateless/sessionless model, removal of `initialize`/`initialized`, `_meta`-carried version/capabilities, `server/discover`, MRTR replacing server-initiated requests, deprecation of Roots/Sampling/Logging, Tasks moved to an extension, list-endpoint invariance, cache hints |
| S2 | A | MCP project blog, announcement of the 2026-07-28 specification. `blog.modelcontextprotocol.io/posts/2026-07-28/` | States the rationale for removing protocol-level session state: servers should scale and load-balance like ordinary stateless APIs, with continuity carried by the application layer |
| S3 | A | Same changelog document as S1, specifically the clause that list results (`tools/list`, `resources/list`, `prompts/list`) must not vary per connection or as a side effect of other requests, with variance by presented authorization as the sole exception | Direct textual basis for "no dynamic per-phase tool exposure" |
| S4 | A | MCP Specification, base server-primitives description (`modelcontextprotocol.io/specification/2025-11-25/server/tools`, read together with the parallel resources/prompts sections of the same doc family) | Source for the model-controlled (Tools) / application-driven (Resources) / user-controlled (Prompts) framing |
| S5 | A | MCP Specification, `server/tools` page (2025-11-25 revision, carried forward) | Tool-execution vs protocol error split; annotations (`readOnlyHint` etc.) as untrusted hints; `resource_link` content type not required to appear in `resources/list` |
| S6 | A | SEP-1303, "Input Validation Errors as Tool Execution Errors." `modelcontextprotocol.io/community/seps/1303-input-validation-errors-as-tool-execution-errors`; tracking issue `github.com/modelcontextprotocol/modelcontextprotocol/issues/1303` | Status: Final. Adopted into the **2025-11-25** specification changelog (not 2026-07-28); rationale given is explicitly to let a model see and self-correct from a validation failure |
| S7 | A | SEP-2663, "Tasks Extension." `modelcontextprotocol.io/seps/2663-tasks-extension`; merged PR `github.com/modelcontextprotocol/modelcontextprotocol/pull/2663` | Status: Final, merged 2026-05-15. Defines `io.modelcontextprotocol/tasks` as an extension, not core protocol, with `tasks/get` polling |
| S8 | A | MCP Inspector, official repository. `github.com/modelcontextprotocol/inspector` | CLI mode for scripted `tools/list`/`tools/call` checks in CI |
| S9 | B | MCP conformance test suite (`@modelcontextprotocol/conformance`, referenced from the MCP org's tooling) | Targets a server URL; relevant only if/when an HTTP transport is added. First-pass finding, not re-fetched in the verification pass — treat the package name as indicative rather than confirmed-current |
| S10 | A | MCP Registry. `registry.modelcontextprotocol.io`; quickstart docs describing it as a preview, metadata-only service | Basis for treating the Registry as an optional, non-load-bearing discoverability channel |

### Go SDK

| ID | Grade | Source | Note |
|---|---|---|---|
| S11 | A | Official Go SDK repository and README. `github.com/modelcontextprotocol/go-sdk` | Version-compatibility table; "maintained in collaboration with Google"; documents the Roots/Sampling/Logging deprecation from the server author's side |
| S12 | A | Go SDK API reference. `pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp` | Confirmed exact shapes used in this report's code: `mcp.AddTool[In, Out any](s *Server, t *Tool, h ToolHandlerFor[In, Out])`; handler signature `func(ctx, req *CallToolRequest, in In) (*CallToolResult, Out, error)`; `mcp.NewInMemoryTransports()`; `mcp.CommandTransport{Command: exec.Command(...)}`; `mcp.NewServer(&Implementation{...}, opts)`; schema inference via `jsonschema` struct tags, backed internally by `google/jsonschema-go`; no Tasks-related types in the index reviewed |
| S13 | A | Go SDK releases page. `github.com/modelcontextprotocol/go-sdk/releases` | v1.7.0: full 2026-07-28 support, backward-compatible to 2024-11-05, `StreamableHTTPOptions.Stateless = true` required for the new protocol over HTTP, `MCPGODEBUG` escape-hatch flags introduced; notes GitHub's production Go MCP server running v1.7.0-pre.3 with over half a million users; v1.8.0 (2026-09-04) adds transport hardening, bounded decoding, `ServerOptions.SupportedProtocolVersions`, cache controls |
| S14 | A | Same releases page as S13 | States `MCPGODEBUG` flags are a transitional mechanism slated for removal once the ecosystem has migrated, targeted for v1.9.0 |
| S15 | A | Same API reference as S12 | Typed-handler behavior: input validated and rejected before the handler runs; `Out` populates `StructuredContent`; unset `Content` auto-filled with the JSON of `Out`; a returned Go `error` becomes an `isError` tool result rather than a protocol error |
| S16 | A | Go SDK pull request #863. `github.com/modelcontextprotocol/go-sdk/pull/863` | "mcp: return input validation errors as tool results, not JSON-RPC errors." Merged 2026-03-31; brings the SDK into line with SEP-1303 (S6) |
| S17 | B | Practitioner report on Go SDK schema generation strictness (generated object schemas set `additionalProperties: false`, so an unexpected argument fails validation) | First-pass finding; exact URL not re-confirmed in the verification pass. Treated as credible because it matches `jsonschema-go`'s documented default behavior, but flagged for the reader to re-check against the pinned SDK version |
| S18 | B | Third-party review citing Go SDK v1.5.0 as dated 2026-04-07 | First-pass finding, not re-confirmed; used only to sanity-check the release cadence implied by S13, not as a load-bearing claim |
| S19 | C | Practitioner deep-dive on Go SDK edge cases: zero Go values still validated against the output schema on error-return paths (so a non-`omitempty` field can turn an error path into an output-validation failure), `json.RawMessage` inferred as a byte array, custom marshallers panicking at tool registration | Single-author technical blog from the first pass; not re-confirmed verbatim in the verification pass. The underlying Go behaviors described (JSON reflection over zero values, `RawMessage`'s `[]byte` kind) are consistent with ordinary `encoding/json`/`reflect` semantics, which raises plausibility, but the specific SDK-level manifestation should be checked against the pinned version before relying on it |
| S20 | A | Go blog, "Traversal-resistant file APIs." `go.dev/blog/osroot` | Introduces `os.Root`/`os.OpenRoot` in Go 1.24: `Open`, `OpenFile`, `Create`, `Stat`, `Lstat`, `Mkdir`, `Remove` scoped to a root, defeating symlink escapes |
| S21 | B | Go package documentation corroborating Go 1.25 additions to `*os.Root` (native `ReadFile`, `WriteFile`, `MkdirAll`, `Rename`, `RemoveAll`), e.g. `pkg.go.dev/github.com/entireio/cli/cmd/entire/cli/osroot` and `pkg.go.dev/github.com/GrayCodeAI/trace/cli/osroot`, which document their own wrapper helpers as predating and filling gaps left by the pre-1.25 API | Indirect (third-party package docs describing what the pre-1.25 stdlib lacked and what 1.25 added) rather than the Go 1.25 release notes directly; consistent across two independent packages |

### Client documentation and configuration

| ID | Grade | Source | Note |
|---|---|---|---|
| S22 | A / B | Claude Code MCP documentation, `code.claude.com/docs/en/mcp`, corroborated by GitHub issues on the product's own repository (`github.com/anthropics/claude-code`, e.g. #87650, #43474 on description/instruction truncation) | Docs (A): `.mcp.json` at project/local/user scope, `claude mcp add`, `CLAUDE_PROJECT_DIR` env var, tool search on by default with only names + server `instructions` loaded at session start, `alwaysLoad`/`anthropic/alwaysLoad`, resources via `@`-mention with auto-provided list/read tools, prompts as `/mcp__server__prompt`, per-server timeouts and `MCP_TOOL_TIMEOUT`, 5-minute (HTTP) / 30-minute (stdio) idle timeouts reset by progress notifications, auto-backgrounding past ~2 minutes, response-size warning/cap behavior, `anthropic/requiresUserInteraction`, schema-validity requirements (root `type: object`, property names limited to 1–64 chars of `[A-Za-z0-9_.-]`), dropping of tools with invalid schemas, flattening of root-level combinators. Issues (B): silent 2,048-character truncation of **both** tool descriptions and server `instructions` (the `/mcp` UI shows the untruncated text, so an author can be misled about what the model actually sees); server `instructions` is specifically what the tool-search mechanism reads to decide whether to search a server at all |
| S23 | A | Cursor MCP documentation. `cursor.com/docs/mcp` (also served at `cursor.com/docs/mcp.md`) | States Cursor supports Tools, Prompts, Resources, Roots, Elicitation and "Apps"; `~/.cursor/mcp.json` (global) and project-level `.cursor/mcp.json`; approval required by default for tool calls |
| S24 | A | VS Code MCP documentation, spanning `code.visualstudio.com/docs/copilot/chat/mcp-servers`, `code.visualstudio.com/api/extension-guides/ai/mcp`, `code.visualstudio.com/docs/agent-customization/mcp-servers`, `code.visualstudio.com/docs/agents/reference/mcp-configuration` | `.vscode/mcp.json` (`servers` key) and portable `.mcp.json` (`mcpServers` key, shared with other tools); `code --add-mcp`; MCP server gallery; resources surfaced via "Add Context"; prompts as `/server.prompt`; auto-discovery of configs written for other clients (e.g. Claude Desktop); workspace-trust gating; stdio sandboxing on macOS/Linux |
| S25 | A | Same VS Code documentation family as S24 | Explicitly lists supported MCP features including Tools, Prompts, Resources, Elicitation, Sampling, Authentication, **Server instructions**, Roots and MCP Apps; states VS Code does not prompt for confirmation on tools annotated `readOnlyHint`, and does prompt for every other tool; separately advertises "dynamic tool discovery" where a server may vary the tools it returns based on workspace or prompt context — in tension with the 2026-07-28 list-invariance rule (S1, S3; see §5.13) |
| S26 | A | Anthropic support documentation, "Getting started with local MCP servers on Claude Desktop." `support.claude.com/en/articles/10949351-getting-started-with-local-mcp-servers-on-claude-desktop` | `claude_desktop_config.json` locations (`~/Library/Application Support/Claude/` on macOS, `%APPDATA%\Claude\` on Windows); Settings → Developer → Edit Config |
| S27 | A | Anthropic documentation on Desktop Extensions / MCPB. `claude.com/docs/connectors/building/mcpb`; `support.claude.com/en/articles/12922929-building-desktop-extensions-with-mcpb` | `.mcpb` bundle install via Settings → Extensions → Advanced settings; a bundle's `manifest.json` `user_config` block drives an auto-generated settings UI — the mechanism this report proposes for collecting a Claude Desktop workspace directory |
| S28 | A | MCPB format specification and reference implementation. `github.com/modelcontextprotocol/mcpb` | Confirms `.mcpb` as a defined, versioned bundle format rather than a Claude-only convention |
| S29 | B | Antigravity MCP configuration, corroborated across Google's own documentation (`docs.cloud.google.com`, a Google Cloud codelab), a third-party integration's docs (`docs.sonarsource.com`) and community write-ups (Medium, DEV.to) | Converges on: project/workspace config at `.agents/mcp_config.json`; global config at `~/.gemini/config/mcp_config.json`; remote server entries keyed by `serverUrl` rather than `url`/`httpUrl` (a stricter schema than most MCP clients); native Agent Skills at `.agents/skills/<name>/SKILL.md`. One integrator's docs claim Antigravity "supports only one MCP config location," which conflicts with the two-level picture above from Google's own material — noted, not resolved, since it may reflect that integrator's own choice of where to write rather than a platform limit |
| S30 | C | First-pass finding on Antigravity tool-count limits, with conflicting numbers (100 vs 512/800) reported across sources by version | Not re-confirmed in the verification pass; kept only as a "moot at ~7 tools" data point, not a design driver |
| S31 | C | First-pass finding reporting Antigravity does not surface MCP Prompts | Not re-confirmed in the verification pass; treated as one more reason the design places nothing load-bearing in Prompts |
| S32 | B | Cursor community/forum reports and a third-party MCP-client compatibility page (e.g. `zuplo.com/learn/mcp/compatibility/clients/cursor.md`) describing a historical ~40-tool soft limit and known limitations in `list_changed` handling | Forum-sourced folklore rather than current official documentation; Cursor's own docs (S23) do not state a numeric cap |
| S33 | B / C | GitHub issues on `anthropics/claude-code` regarding truncation (folded into S22, grade B) plus separate, lower-confidence first-pass reports that claude.ai truncates tool descriptions near ~500 characters (March 2026) and that claude.ai/Claude Desktop ignore the `instructions` field (Mar–Apr 2026), with a later (August 2026) comparison reportedly showing claude.ai receiving full descriptions | The claude.ai-specific truncation figure and its apparent reversal are single-source and not independently re-confirmed; presented in the report with that uncertainty intact rather than as settled fact |
| S34 | C | First-pass report that Cursor versions 3.15/3.19 were still requesting protocol revision 2025-11-25 | Not re-confirmed in the verification pass; Cursor publishes no protocol-revision number in its own docs, so this remains the only (unverified) data point on Cursor's protocol-negotiation behavior |
| S35 | B | First-pass finding on an AI SDK's stdio client probing `server/discover` before falling back to legacy `initialize`, and on annotations being explicitly untrusted-by-default in at least one SDK's client implementation | Not re-fetched verbatim in the verification pass; consistent with the spec's own framing in S1/S5 |

### Tool design and agent-facing guidance

| ID | Grade | Source | Note |
|---|---|---|---|
| S36 | A | Anthropic engineering blog, "Writing effective tools for agents." `anthropic.com/engineering/writing-tools-for-agents` | Consolidate functionality into fewer, higher-level tools; namespace tools by resource; test prefix vs suffix naming choices empirically; return meaningful, high-signal context and support a concise/detailed response mode; make error messages actionable; use natural-language identifiers over opaque IDs; write descriptions as if for a new hire; iterate against evals rather than intuition alone |
| S37 | A | Anthropic/Claude platform documentation on tool use, `platform.claude.com/docs/en/agents-and-tools/tool-use/implement-tool-use` (and related pages) | Recommends descriptions of at least 3–4 sentences; documents an optional `input_examples` field for illustrating complex tool calls |
| S38 | B | Block engineering blog, "Block's playbook for designing MCP servers." `engineering.block.xyz/blog/blocks-playbook-for-designing-mcp-servers` | Design top-down from the workflow rather than bottom-up from API endpoints; combine multiple internal calls into one high-level tool; where chaining is unavoidable, spell out steps/dependencies and keep intermediate outputs concise; notes that Block's own agent (Goose) uses server `instructions` to help build its system prompt and uses tool annotations (e.g. `readOnlyHint`) for automatic approval decisions — a second, independent confirmation of the pattern documented for VS Code in S25; states Block has built more than 60 MCP servers |
| S39 | B | First-pass finding: a named AI practitioner's blog post framing MCP as "a UI for agents" and recommending it be paired with Skills for guidance on when/how to combine tools | Not re-fetched verbatim in the verification pass; used only as a secondary echo of the AGENTS.md/Skill pairing already independently supported by S40 |
| S40 | B | Vercel engineering blog, "AGENTS.md outperforms Skills in our agent evals." `vercel.com/blog/agents-md-outperforms-skills-in-our-agent-evals` | Four-condition Next.js-documentation eval: no-docs baseline 53% pass; an available Skill with default invocation also 53%, because it went uninvoked in 56% of runs (and underperformed the baseline on some sub-metrics, suggesting an unused skill can add noise rather than being neutral); adding an explicit "use this skill" instruction raised invocation to 95%+ but pass rate to only 79%, and was sensitive to exact wording; a compressed, always-loaded AGENTS.md index reached 100%. Vercel's own stated caveat: skills still win for narrow, explicitly-requested vertical workflows |
| S41 | C | First-pass finding: a survey claiming the Agent Skills convention (`SKILL.md`) is read by Claude Code, Cursor, Codex, Gemini CLI, Antigravity and VS Code | Not re-fetched verbatim in the verification pass; partially corroborated independently by S29's confirmation of Antigravity's own `.agents/skills/<name>/SKILL.md` convention |
| S42 | C | First-pass finding: an external audit criticizing Shrimp Task Manager's tool descriptions for not explaining when to use one tool over a similarly-named sibling | Not re-confirmed in the verification pass |

### Production and reference MCP servers

| ID | Grade | Source | Note |
|---|---|---|---|
| S43 | A | GitHub Changelog, announcing the GitHub MCP Server's migration to the official Go SDK plus tool-specific configuration and resource completions. `github.blog/changelog/2025-12-10-the-github-mcp-server-adds-support-for-tool-specific-configuration-and-more/` | Confirms the mark3labs/mcp-go → official-SDK migration completed; a general-availability, first-party announcement |
| S44 | B | `github.com/github/github-mcp-server` pull requests referencing "toolsnaps" (schema/description snapshot tests), e.g. #1429, #1430, #1432, #1433, #1440, #1468, and issue #1184 | Confirms an actual, running snapshot-testing practice guarding the tool surface against silent drift |
| S45 | B | Same repository as S44 | Confirms dynamic/toolset-scoped tool grouping (curated by production usage data) as GitHub's answer to a large tool count |
| S46 | B | `github.com/eyaltoledano/claude-task-master`, pull request #1181 | Project's own measured numbers: "all" tool-loading mode = 36 tools ≈ 21,000 tokens (default); "standard" mode ≈ 52% token reduction from that baseline; "core" mode ≈ 76% reduction (≈ 5,000 tokens) |
| S47 | C | First-pass finding on mcp-shrimp-task-manager's architecture: a `DATA_DIR`-rooted JSON store, roughly 15 tools, and prompt templates customizable via environment variables | Not re-confirmed verbatim in the verification pass |
| S48 | B | `github.com/steveyegge/beads` (main README and `integrations/beads-mcp/README.md`); `pypi.org/project/beads-mcp/`; a live example issue, `github.com/gastownhall/beads/issues/4036` | CLI-first design ("you don't use Beads directly as a human... your coding agent will"); `AGENTS.md`/`CLAUDE.md` integration plus a `bd prime` command that emits workflow context; `bd ready` as the core "what can I work on" query; storage evolution from JSONL+SQLite to Dolt (version-controlled SQL with cell-level merge); hash-based issue IDs for zero-conflict multi-agent/multi-branch use; every MCP tool accepts an optional `workspace_root`, canonicalized (symlinks resolved, git top-level detected) and pooled by resolved path; `set_context` retained only for backward compatibility; a resolved bug (issue #346) in which self-referential Pydantic models produced an output schema with a root-level `$ref`, which broke MCP tool loading in Claude Code until the server was changed to guarantee `type: object` at the schema root; per-project MCP server instances explicitly documented as "not recommended" in favor of one server routed by argument |
| S49 | C | First-pass finding on an MCP server described as a work-item dependency graph with gates (roughly 13–14 tools; advancing an item fails when required notes are missing) | Not re-confirmed verbatim in the verification pass |
| S50 | C | First-pass finding on a small deployment-gate MCP server built on a Python state-machine library (Burr), returning `invalid_transition` plus the currently reachable actions on an out-of-order call | Not re-confirmed verbatim in the verification pass |
| S51 | C | First-pass finding: a practitioner report describing a ten-step workflow where models skipped, merged and reordered steps despite emphatic prose instructions, resolved only once the sequence was enforced inside a stateful MCP tool | Not re-confirmed verbatim in the verification pass. This is the single most load-bearing C-grade claim behind the Guide-with-enforcement recommendation (§3.1, §5.8); it is corroborated in direction (not in specifics) by the general, better-evidenced finding that trigger-dependent guidance is unreliable (S40) and that orchestration itself is weak across models (S59), but the specific ten-step anecdote should be treated as illustrative, not as a controlled result |
| S52 | C | First-pass finding on prompt-only orchestration degrading as an agent's context fills | Not re-confirmed verbatim; treated as a plausible but unverified qualifier |
| S53 | B | `github.com/txn2/mcp-data-platform`, its issue tracker (e.g. issue #1692) | A Go, official-SDK-based server with roughly 28 tools; a reported finding that 21 lacked annotations, producing three write-confirmation prompts before the first usable result; the source's own recommendation of a registration-time table test over every tool's annotations, which this report adopts directly (§3.12, §5.9) |
| S54 | B / A | FastMCP maintainer notes, `gofastmcp.com/development/v4-notes/background-tasks` (B, a maintainer's own technical note, not an MCP-project primary); C# SDK documentation, `csharp.sdk.modelcontextprotocol.io/v2/concepts/tasks/tasks.html` (A for the C# SDK specifically) | The disagreement itself is the finding: the FastMCP note states no SDK, in any language, yet ships a Tasks runtime; the C# SDK's own docs describe a `ModelContextProtocol.Extensions.Tasks` package implementing SEP-2663. Both can be true if the landscape is uneven across languages — which is exactly why this report treats Tasks as unavailable for Go specifically (confirmed absent from the v1.8.0 API surface reviewed, S12) without generalizing to "no SDK anywhere" |
| S55 | B | Multiple corroborating security write-ups: `cymulate.com/blog/cve-2025-53109-53110-escaperoute-anthropic/`, `cybersecuritynews.com` coverage of the same, `thehackernews.com/2026/01/three-flaws-in-anthropic-mcp-git-server.html`, `darkreading.com` coverage of the Git-server flaws, and `endorlabs.com/learn/classic-vulnerabilities-meet-ai-infrastructure-why-mcp-needs-appsec` for the aggregate industry figure | CVE-2025-53109 (CVSS 8.4) and CVE-2025-53110 (CVSS 7.3): Anthropic's reference filesystem MCP server's directory-prefix containment check could be bypassed by symlinks and by sibling directories sharing a string prefix; fixed in the npm package's 2025.7.1. CVE-2025-68143, -68144, -68145 (disclosed January 2026): Anthropic's reference Git MCP server separately shipped an unconstrained `git_init` path argument, a repository-boundary check that did not actually enforce its stated restriction, and a `git_diff`-based primitive usable to overwrite or empty arbitrary files; fixed across the 2025.9.25 and 2025.12.18 releases. The Endor Labs figure (a large share of surveyed MCP servers having path-traversal issues) is that vendor's own research and is presented in this report as a single vendor's claim, not a settled industry statistic |
| S56 | B | Security-industry coverage of two distinct CVEs: Oligo Security's own disclosure write-up plus `thehackernews.com`/`socradar.io`/Recorded Future coverage of CVE-2025-49596 (MCP Inspector); OSV/PyPA advisory data (`osv.dev`, PyPI security advisories) for CVE-2025-66414 (Python MCP SDK); a secondary blog (AgenticWire) additionally naming the Go SDK among historically affected implementations | CVE-2025-49596 (CVSS 9.4): the MCP Inspector's browser-facing proxy lacked a session token and origin validation before v0.14.1 (fixed June 2025), letting an untrusted webpage reach a local Inspector instance and execute code. CVE-2025-66414: the Python MCP SDK shipped with DNS-rebinding protection off by default for unauthenticated localhost HTTP servers until v1.23.0. The claim that the Go SDK shared this specific default gap comes only from the secondary source and is explicitly flagged in the report (§3.13, §5.12) as unconfirmed against a primary advisory — treat it as a prompt to verify the pinned Go SDK's HTTP defaults directly, not as an established fact |
| S57 | C | General security-community usage of "tool poisoning" / "MCP rug pull" to describe undetected changes to a previously-approved tool's description or schema after installation | First-pass finding; the underlying concept is real and named in MCP security discourse, but the specific source was not re-confirmed verbatim in the verification pass. Used only to motivate snapshot testing and signed releases (§3.12, §3.13), which are good practice independent of this citation |
| S58 | D | A lower-quality third-party feature-comparison matrix, encountered in the first pass, asserting Cursor does not support Resources/Prompts | Superseded in this report by Cursor's own documentation (S23); kept in the ledger only to document the specific disagreement recorded in §5.13, not as a source this report relies on |
| S59 | B | MCP-Bench benchmark paper (`arxiv.org/abs/2508.20453`; code at `github.com/Accenture/mcp-bench`; peer-reviewed listing at `openreview.net/forum?id=fe8mzHwMxN`, accepted as an ICLR 2026 poster and a NeurIPS 2025 workshop paper) | 250 tools across 28 realistic MCP servers; evaluation of 20 frontier LLMs finds persistent, significant weaknesses in multi-step tool orchestration rather than a solved problem |
| S60 | C | First-pass finding: a May 2026 Go-focused technical article describing MCP schema-validation failures as protocol (JSON-RPC) errors | Appears to predate, or not account for, Go SDK PR #863 (S16, merged 2026-03-31), which changed this behavior; kept in the ledger specifically as the losing side of the disagreement recorded in §5.13 rather than as current guidance |
| S61 | C | First-pass finding: a report of at least one MCP client validating `structuredContent` against a tool's declared `outputSchema` even when the result carries `isError: true` | Not re-confirmed verbatim in the verification pass; used only to explain why this report avoids putting rich, schema-bound payloads on error paths (§3.8) |
| S64 | C | General compatibility-tracker and community commentary on Claude Desktop's support for MCP primitives beyond Tools (Resources, Prompts) and on its handling of Roots/Elicitation/MRTR-style flows | No single authoritative Anthropic document was found in either pass stating Claude Desktop's full primitive support matrix; treated throughout as an open item for Spike 0 (§3.14) rather than a settled fact |
| S66 | C | A single March 2026 community forum question asking whether Cursor's tool-count ceiling had moved from 40 to 80 | Not independently confirmed; included only to show the number itself is unstable and therefore not worth designing precisely against (§5.5, §5.13) |

### On the D-grade claims used throughout

Claims marked **D** are this report's own inferences (e.g., specific numeric thresholds for the eval exit criteria in §3.14, the recommended file-locking libraries in §5.3, the token-budget target in §3.7) or knowledge the model held prior to this session's searches that was not independently re-verified against a live source in either research pass (e.g., the general behavior of `gofrs/flock` and `google/renameio/v2`, which are real, commonly-used Go packages for advisory locking and atomic file replacement, but whose specific API details were not re-checked here). These are lower-confidence than any A/B/C entry above and are flagged individually in §6.3 ("Unverified items") where they matter enough to affect the recommendation.
