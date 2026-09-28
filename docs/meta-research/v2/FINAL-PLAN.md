# Vivechak: The Final Plan

> **What this is:** The single authoritative plan for Vivechak v2.0, distilled from ~900KB of research, three strategy documents, two synthesis artifacts, two audit reports, and fresh web verification.
>
> **What it replaces:** `what-next-old.md`, `WHATS-NEXT.md`, `ROADMAP-NEXT.md`, and the research synthesis. Those documents were the *process*. This document is the *result*.
>
> **Last audited & calibrated:** 2026-09-27 — Comprehensive accuracy audit against all 12 research sessions (~770KB). Calibrated to 100% evidentiary alignment: D-011..D-014 codified to continue meta-research registry (avoiding collision with D-001..D-010), generator byte budgets calibrated to empirical sizes (~8KB / ~4KB), comparison scope formalized as single-prompt WEP generator without routing, tool synthesis defined as artifact/prompt assembler, and code-signing fallback path established.
>
> **Date:** 2026-09-27
> **Evidence corpus:** 12 research sessions (~770KB), 4 pre-research strategy docs, 2 synthesis artifacts, 2 audit reports, 1 code-level MCP tool audit, fresh web verification of MCP spec / Go SDK / Agent Plugins / competitors

---

## Part 0: Audit Findings

Before planning, the full evidence corpus was audited for completeness and accuracy. **4 structural gaps and 10 content gaps found:**

### Structural Gaps

| # | Gap | Impact | Resolution |
|---|---|---|---|
| 1 | **DECISIONS.md still shows `status: proposed`** — D-001 through D-004 were never formally updated to `accepted` with evidence refs | Low — synthesis artifact contains the verdicts | Superseded by the Decision Log in this document (codified as D-011 through D-014 to avoid collision with repo meta-research D-001..D-010) |
| 2 | **SYN-01 was never executed as a standalone session** — synthesis was done by the main agent as an artifact, not as a pipeline session file | Low — substance is equivalent | Accept — synthesis artifact IS the SYN-01 deliverable |
| 3 | **Official ROADMAP.md has zero mention of MCP** — still describes Phase 5 as standalone Engine (LangGraph/ADK/CrewAI) | **High** — public roadmap contradicts decided direction | **Must update ROADMAP.md in Phase 0** |
| 4 | **README.md still says "pre-development"** — lines 7 and 17 use the old framing | Medium — blocks repositioning | Update in Phase 1 |

### Content Gaps (from gap audit)

| # | Gap | Where Found | Impact |
|---|---|---|---|
| 5 | **R/N/B/X profiling originated in R-06, not R-03** — R-03 proposed a 6-flag profile; R-06 refined to 4-dimension R/N/B/X | Provenance correction | None — R-06's version is what we adopt |
| 6 | **7 provenance misattributions in synthesis** — claims attributed to "both" sessions that actually came from only one run | Citation quality | None — all claims backed by at least one session |
| 7 | **Claude Code response window caps (10K warn, 25K hard cap)** — absent from synthesis | `without-repo-link/R-04` | **Added to Phase 2** — tool responses must use progressive disclosure |
| 8 | **Mechanical vs. substantive gate limitation** — MCP validation can check field presence, not semantic adequacy | `with-repo-link/R-04` | **Added to Phase 2** — `vivechak_run_gate` explicitly documents scope |
| 9 | **7 missing findings** from sessions not captured in synthesis (grill-me competitor, OrchestKit, Beads failure case, Cochrane 6.5–38.6% significance shift, Danish Mini-HTA 60-form sprawl, Claude Code token caps, gate limitations) | Gap audit cross-check | 2 of 7 affect plan (items 7–8 above); 5 of 7 validate existing decisions |
| 10 | **Session injection glitch in without-repo R-06** — session runner omitted GENERATOR.md and R-03 context block | Provenance correction | Use `with-repo-link/R-06` as sole authoritative draft for generator designs |

### Without-Repo-Link Sessions Were More Valuable Than Expected

The gap audit revealed that **without-repo-link sessions contributed MORE unique technical findings** than the targeted with-repo-link sessions: all 5 CVE references, the PATH truncation discovery, the 35% client limitation stat, and the Claude-Code-Deep-Research competitor threat all came exclusively from the broader, untargeted runs. The with-repo-link sessions were better at methodology design; the without-repo-link sessions were better at environmental discovery.

---

## Part 1: First Principles — The Hard Constraints

These are non-negotiable truths, validated by evidence:

### What IS True

| # | Constraint | Evidence Grade |
|---|---|---|
| 1 | **The methodology is sound and unique.** No competitor replicates the full pipeline: complexity scoring → DAG → evidence grading → structured falsification → ADR → exit gate. | A (12 sessions, competitor analysis, cross-domain validation) |
| 2 | **The delivery is broken.** 106KB of documentation, copy-paste workflow, 0 stars, 0 forks. The framework is right; the packaging prevents anyone from discovering or using it. | A (GitHub metrics, behavioral analysis) |
| 3 | **MCP is the correct delivery vehicle.** Stateless protocol, universal agent support, Two-Way Door risk. Single binary Go server with embedded templates. | A (MCP 2026-07-28 spec, Go SDK Tier 1, Agent Plugins 1.0.0 spec) |
| 4 | **The server must be a Guided Worker, not a Guide or a Worker.** Atomic tools returning guidance in response data. No server-side FSM. No LLM API calls. | A (MCP spec removes sessions; P8 says "not rigid stage gates"; production MCP patterns) |
| 5 | **Three scope levels work.** Project (full pipeline → FAD), Decision (1–3 sessions → ADR), Comparison (1 session → WEP matrix). Comparison gets a template, not a generator. | A (Cochrane, ODNI ICD 203, live codebase analysis, R-03/R-06 cross-validation) |
| 6 | **Distribution is platform-native first, plugins second.** GitHub Releases → Homebrew/Scoop/winget → thin Agent Plugin. Claude Desktop and Antigravity don't support Agent Plugins. | A (Agent Plugins TSC membership, host testing, PATH truncation evidence) |
| 7 | **The moat is auditable rigor, not orchestration.** Evidence grading, falsification discipline, exit gates, ACH conflict resolution. Orchestration is being commoditized by deep research tools. | B (Spec Kit 110K+ stars proves demand; Claude-Code-Deep-Research copies A-E grading; deep research tools gaining multi-session) |

### What Is Settled (Closed Decisions)

| # | Decision | Why It's Closed |
|---|---|---|
| 1 | **Go over Python** | Single binary distribution, 5-20ms startup, `os.Root` security, Tier 1 SDK. Server does file I/O + validation, not AI. |
| 2 | **Monorepo** | Server reads GENERATOR.md and templates/ directly. Separate repos create sync problems. Reversible if scale demands it. |
| 3 | **9 tools with `vivechak_` prefix** | Code-level audit dropped `synthesize` (violates Guided Worker), renamed 2 tools. See tool table and `mcp-tool-audit.md`. |

---

## Part 2: The Plan

### Phase 0: Fix Live Bugs (~1–2 days)

These are methodological defects in the SHIPPED v1.1 framework. Every pipeline executed with the current templates has subtle quality issues. Fix before ANY new work.

| # | Fix | File(s) | What Exactly To Do |
|---|---|---|---|
| **0.1** | **ACH matrix orientation** | `templates/CONFLICT-RESOLUTION.template.md` | Transpose the matrix: hypotheses → ROWS, evidence → COLUMNS. Current layout (hypotheses in columns) induces confirmation bias per Dhami et al. 2024. Update template instructions to explain why. |
| **0.2** | **Evidence grade inconsistency** | `FRAMEWORK.md` §5.1 + `GENERATOR.md` §DELIVERABLE | FRAMEWORK.md is correct — peer-reviewed studies = Grade A (primary/authoritative). Update GENERATOR.md line ~181 to match: `A (official docs/RFCs/peer-reviewed studies)`. |
| **0.3** | **Update official ROADMAP.md** | `ROADMAP.md` Phase 5 section | Replace standalone Engine plan (LangGraph/ADK/CrewAI) with MCP server direction. Reference research evidence. Keep Phase 6 frontiers as-is (still gated). |
| **0.4** | **Initiate code signing enrollment** | Apple Developer ID / Azure Trusted Signing | Submit organization and individual identity verification Day 1. Empirical lead time is 1–20+ business days with reported stalls; must run asynchronously through Phases 0–2 so signing is unblocked for Phase 3. |

**What NOT to change in Phase 0:** Don't bump VERSION. Don't rewrite README. Don't touch FRAMEWORK.md beyond the grade fix. Minimal, surgical fixes only.

---

### Phase 1: Methodology Expansion (~2 weeks)

**Goal:** Vivechak works at 3 scope levels. All documentation reflects the new positioning.

| # | Deliverable | Details | Door Type |
|---|---|---|---|
| **1.1** | `GENERATOR-DECISION.md` | **~7.9KB** (measured 7,908B) decision-level generator. Compressing to ~5KB cuts load-bearing methodology (WEP sensitivity, disconfirming evidence requirements). 4-step METHOD: reframe neutrally → profile R/N/B/X (Reversal cost, Novelty, Blast radius, Regulatory) → route to 4 lanes (SKIP / SPIKE / CONFIRM / DEEP) → compose 1–3 role-sessions (Landscape / Comparison / Falsify). Requester's LEANING withheld from L/C prompts; passed only to Falsify session. Hard cap of 3 sessions; escalate if more needed. Use R-06 with-repo-link draft as starting point. | Two-Way |
| **1.2** | `GENERATOR-COMPARISON.md` & `templates/COMPARISON-SESSION.template.md` | **~4KB** (measured 4,053B) single-prompt comparison generator and template. Pre-structured single-session prompt using WEP (no multi-decision routing/profiling logic). Minimum 3 options (avoid binary framing bias, P8). Includes SCOPE CHECK guard (redirects one-way/open-ended decisions to decision-level), criteria weighting, source-diversity guard. Use R-06 with-repo-link draft as starting point. | Two-Way |
| **1.3** | `FRAMEWORK.md` §9: Multi-Scope Research | Document the Scope × Depth model. 3 scope levels. 9 methodology invariants (I1–I9). Decision routing matrix (R/N/B/X → 4 lanes). 3 modes (evaluate / landscape / revalidate). Scope × Depth are independent axes. | Two-Way |
| **1.4** | Reposition all public-facing text | Change "pre-development research" → "evidence-grounded research for technical decisions" across README.md, AGENTS.md, GENERATOR.md BRIEF section. Update tagline. Add "Who is this for?" section targeting post-vibe developers. | Two-Way |
| **1.5** | Evidence grading lineage clarification | Clarify in FRAMEWORK.md §5: "Admiralty-Code-derived source grading with GRADE-inspired modifiers." Currently says "Cochrane/GRADE lineage" which overclaims — Vivechak's A–E grades sources by reliability (Admiralty Code), not bodies of evidence across endpoints (clinical GRADE). | Two-Way |
| **1.6** | Drift prevention mechanism | Add `CORE:BEGIN` / `CORE:END` marker blocks around shared methodology rules in GENERATOR.md, GENERATOR-DECISION.md, and COMPARISON-SESSION.template.md. Add CI check (diff/checksum) to detect when one file drifts from the others. Danish Mini-HTA 60-form sprawl (Kidholm/Ehlers) provides empirical evidence this is necessary. | Two-Way |
| **1.7** | Context engineering note | Add note to P1 in FRAMEWORK.md linking it to the "context engineering" paradigm. Address backlog item PE-01. | Two-Way |

**Validation:** Use the decision-level generator to research one real decision (e.g., a decision from the MCP server design). This simultaneously validates the new scope level AND produces dogfooding evidence.

---

### Phase 2: Go MCP Server (~3–4 weeks)

**Goal:** A working MCP server that any agent can use to run Vivechak pipelines.

#### Spike 0: Client Compatibility Canary (~1 day)

**Before building the full 10-tool surface**, build a throwaway 2-tool canary server and test it on all 7 target hosts. Verify:

| Probe | What to check | Why |
|---|---|---|
| Tool listing | Do both tools appear in each client's tool picker? | Claude Code defers tools; Cursor has a reported 40-tool cap |
| Annotations | Does `readOnlyHint: true` actually suppress approval prompts? | VS Code confirms every non-read-only tool |
| Response channels | Does `structuredContent` reach the model? Does plain `content` also work? | Cursor reportedly drops structuredContent-only; Gemini CLI rejects missing structuredContent |
| `alwaysLoad` | Does the flag work in Claude Code for the status tool? | Claude Code tool search hides tools by default |
| PATH resolution | Does an absolute-path `command` work? Does a bare command fail? | macOS GUI hosts truncate PATH to `/usr/bin:/bin:/usr/sbin:/sbin` |
| Workspace binding | Does `CLAUDE_PROJECT_DIR` / `VIVECHAK_PROJECT_ROOT` resolve correctly? | Each host sets different env vars for stdio servers |
| Output size | What happens when a response exceeds 10K tokens? 25K tokens? | Claude Code warns at 10K, hard caps at 25K |

**Kill criterion:** If more than 2 hosts fail basic tool listing + response handling, reassess the MCP approach before investing in full build. (Source: without-repo R-02 "Spike 0" concept.)

#### Architecture

```
cmd/vivechak/
├── main.go                  # Entry point: MCP server (stdio) + CLI subcommands
├── serve.go                 # MCP server startup
├── config.go                # `mcp-config --client <host> --write`
└── doctor.go                # `doctor` — detect corruption, orphans, stale refs

internal/
├── core/                    # Pure logic (no MCP dependency)
│   ├── workspace.go         # Init, discover, validate workspace (scope-aware)
│   ├── generator.go         # prepare_generator: fill context into generator prompts
│   ├── plan.go              # Plan parse/validate (DAG for project, routing for decision)
│   ├── dag.go               # DAG resolution, cycle detection, dependency ordering
│   ├── frontmatter.go       # YAML frontmatter read/write
│   ├── inject.go            # Context injection: extract upstream findings → fill slots
│   ├── validate.go          # 4-level validation ladder
│   └── gate.go              # Phase 0 exit gate (Track A + Track B)
│
├── store/                   # File I/O
│   ├── workspace.go         # os.Root path confinement (Go ≥ 1.25)
│   ├── atomic.go            # write-temp → fsync → rename
│   └── lock.go              # Advisory file locking (gofrs/flock)
│
├── mcp/                     # MCP adapter
│   ├── server.go            # Tool/Resource/Prompt registration
│   ├── tools.go             # 9 tool handlers
│   ├── resources.go         # Supplementary read-only mirrors
│   ├── prompts.go           # Optional entry shortcuts
│   └── envelope.go          # Standardized response envelope
│
└── embed/                   # Embedded assets
    ├── generators/           # GENERATOR.md, GENERATOR-DECISION.md, GENERATOR-COMPARISON.md
    ├── templates/            # All 5 templates
    └── embed.go              # go:embed declarations
```

#### 9 Tools (Build Order)

> **Why 9, not 10:** Code-level audit (2026-09-27) revealed `vivechak_synthesize` can't actually synthesize without calling an LLM (violates Guided Worker pattern). Synthesis is a session — the agent calls `vivechak_next_session` for the SYN-01 node, gets findings injected, executes it, saves via `vivechak_save_session`. Additionally, `get_generator_prompt` was renamed to `prepare_generator` (returns prompt with context filled in, not raw template text) and `save_pipeline` was renamed to `save_plan` (handles all 3 scope levels, not just project DAGs).

| # | Tool | What | Workflow Step | Read-Only? |
|---|---|---|---|---|
| 1 | `vivechak_init` | Create workspace structure, copy templates. Accepts `scope` param (project \| decision \| comparison) to adjust directory layout. | Project §3, Decision §setup, Comparison §setup | No |
| 2 | `vivechak_prepare_generator` | Accept user's context (project vision / decision context / comparison context) + scope level. Return the appropriate generator prompt WITH context slots filled in, ready for the agent to execute or the user to paste into a deep research session. | Project §1-2, Decision §1-2, Comparison §1-2 | Yes |
| 3 | `vivechak_save_plan` | Validate + persist agent-generated plan. At project scope: validate DAG + save RESEARCH-PIPELINE.md + DECISIONS.md. At decision scope: validate routing + save [ID]-PLAN.md + [ID]-[slug].md (proposed ADR). At comparison scope: save the generated prompt. | Project §2 (save), Decision §2 (save), Comparison §2 (save) | No |
| 4 | `vivechak_status` | Scan workspace, report progress at current scope. Project: DAG completion, blocked sessions, decision lock status. Decision: which of 1-3 sessions are done, ADR status. | Orientation (any time) | Yes |
| 5 | `vivechak_next_session` | Return next actionable session(s) with full prompts AND upstream findings injected into context slots. For SYN-01 (synthesis), returns the synthesis prompt with ALL session findings aggregated. This is the core value proposition. | Project §4a, Decision §3 | Yes |
| 6 | `vivechak_save_session` | Validate (4-level ladder) + persist completed session output. Checks YAML frontmatter, evidence grades, required sections. Does NOT reject for formatting variation — validates structure, not style. | Project §4d, Decision §3 (save) | No |
| 7 | `vivechak_record_decision` | Save ADR or conflict resolution via `artifact_type` field. Validates against DECISIONS.template.md or CONFLICT-RESOLUTION.template.md schema. | Project §5-6, Decision §4 | No |
| 8 | `vivechak_validate` | Dry-run validation on any artifact (session, plan, decision, FAD). No side effects. | Pre-save check (any time) | Yes |
| 9 | `vivechak_run_gate` | Phase 0 Gate (Track A + Track B). Mechanical checks only: ADR lock status, reversal triggers present, evidence density, FAD completeness. Semantic quality (is the premortem substantive?) is the host agent's job. | Project §8 | No |

#### Key Design Rules

| Rule | Detail |
|---|---|
| **Guided Worker** | Every response includes `next_step` telling the agent what to do next. Hard refusals ONLY for physically impossible operations (missing files), never for "wrong order." |
| **Workspace as database** | Parse/write `research/` Markdown+YAML directly. No `.vivechak/state.json`. Workspace path as explicit argument on every call. |
| **Workspace resolution** | Explicit `project_root` argument > `VIVECHAK_PROJECT_ROOT` env > discover `research/` directory walking up from CWD > error. Without-repo R-02 proposed a 6-step precedence chain — adopt the 4-step version. |
| **Concurrent writes** | Advisory file locking (`gofrs/flock`) with timeout on every write operation. Never silently overwrite. Atomic writes via `write-temp → fsync → rename`. Two agents on one repo = two stdio server processes; locking prevents data loss. |
| **`vivechak doctor`** | CLI subcommand to detect corruption, schema-invalid frontmatter, orphaned sessions, stale cross-references. Never silently repair — report and let the user decide. |
| **Stateless** | MCP 2026-07-28 mandates no protocol sessions. Tools receive full context on every call. Multi-round elicitation (MRTR) conversational state is preserved in client `requestState` tokens or transient workspace drafts (`.vivechak/draft-state.json`), maintaining strict process-level statelessness. |
| **Dual content** | Return both `structuredContent` AND text `content` in every response (Cursor drops structuredContent-only; Gemini CLI rejects missing structuredContent). |
| **Response size limit** | Claude Code warns at 10K tokens, hard caps at 25K tokens. `vivechak_synthesize` and `vivechak_run_gate` MUST use progressive disclosure (summary → detail on request via `verbose: true` parameter), not full dumps. Include `meta.truncated: true` when responses are shortened. |
| **Validation ladder** | L1 Construct (autofill) → L2 Block (save as `status: draft`) → L3 Warn (save but flag) → L4 Gate (project-wide structural completeness). |
| **Gate scope** | `vivechak_run_gate` performs **structural/mechanical validation only** (field presence, evidence density counts, decision lock status). Semantic quality assessment (is the premortem substantive? are alternatives genuine?) is the host agent's responsibility. Document this limitation explicitly. |
| **Security** | `os.Root` for path traversal protection (5 CVEs in Anthropic reference MCP servers in 10 months). Pin Go ≥ 1.25. |
| **Tool descriptions** | Front-load decisive action in first 300–500 chars (Claude Code truncates at 2,048). Mark `vivechak_status` with `alwaysLoad: true` for Claude Code so it's always visible for agent orientation. |
| **Templates embedded** | `embed.FS` — single binary, no external file dependencies. |

#### Standardized Response Envelope

Every tool returns:

```json
{
  "success": true,
  "message": "Session R-01 saved as draft (2 warnings, 0 errors)",
  "data": { "..." },
  "warnings": ["W-LOW-OPTIONS"],
  "next_step": "Record decisions using vivechak_record_decision, then run vivechak_next_session."
}
```

#### Testing Strategy

| Tier | What | When |
|---|---|---|
| **1** | Unit tests: DAG resolution, frontmatter parsing, validation ladder, path traversal attack vectors | Every PR |
| **2** | In-memory wire tests: `mcp.NewInMemoryTransports()` — full tool call sequences without subprocess | Every PR |
| **3** | CI smoke: build binary → `npx @modelcontextprotocol/inspector` → verify zero stdout leakage | Every release |
| **4** | Agent evals: execute full pipeline via MCP on 3+ prompts across Claude Code / Cursor / Antigravity | Pre-release |

---

### Phase 3: Distribution (~1–2 weeks)

**Goal:** One-command install from any agent host.

| Timing | Channel | Details |
|---|---|---|
| **Day 1** | GitHub Releases | GoReleaser: 6 platforms (darwin/{amd64,arm64} → lipo universal, windows/{amd64,arm64}, linux/{amd64,arm64}). Checksums + SBOMs. |
| **Day 1** | Shell installers | `install.sh` + `install.ps1` one-liners that download correct binary |
| **Day 1** | `vivechak mcp-config` | CLI subcommand: `--client {cursor,vscode,antigravity,claude-desktop,chatgpt,codex,kiro} --write` generates correct absolute-path config for each host |
| **Day 1** | Per-host docs | Copy-paste MCP config blocks for all 7 hosts |
| **Week 2** | Homebrew tap | macOS + Linuxbrew. Requires Apple Developer ID signing (start enrollment Day 1 in Phase 0.4). *Contingency:* If identity verification stalls, release initial community binary via custom tap with Gatekeeper quarantine-strip hook (`xattr -dr com.apple.quarantine`), plus Scoop and winget portable zips. |
| **Week 2** | Scoop bucket | Windows, no UAC required |
| **Week 2** | winget manifest | Pre-installed on Windows 11 |
| **Week 3** | Thin Agent Plugin | `plugin.json` + `mcp.json` + `skills/` quick-start. For Cursor, VS Code, GitHub Copilot, Kiro. |
| **Month 2** | Claude Desktop `.mcpb` | Needs universal binary. Only one-click path for Claude Desktop. |
| **Month 2** | MCP Registry `server.json` | Discovery aggregator. CI step recalculating `fileSha256` per release. |

#### Cross-Compilation Matrix

```
darwin/amd64  ──┐
                ├──▷ lipo → vivechak_darwin_universal (Homebrew, .mcpb)
darwin/arm64  ──┘
windows/amd64 ──────▷ vivechak_windows_amd64.exe (winget, Scoop)
windows/arm64 ──────▷ vivechak_windows_arm64.exe (P1)
linux/amd64   ──────▷ vivechak_linux_amd64 (tar.gz, deb, rpm)
linux/arm64   ──────▷ vivechak_linux_arm64 (tar.gz, deb, rpm)
```

---

### Phase 4: Demand Proof (~2 weeks + ongoing)

**Goal:** Answer the only question that matters — *will anyone use this?*

| Deliverable | What | Why |
|---|---|---|
| **Real case study** | Run full Vivechak pipeline (via MCP server) on a real project. Publish complete `research/` directory. | Current SAMPLE-PIPELINE.md is synthetic. Real artifacts are 10× more compelling. |
| **Free wedge skill** | Single-file `SKILL.md` for Cursor skills / skills marketplaces. Top-of-funnel. | Zero-cost distribution through existing marketplaces. `grill-me` (12.9K–460K installs) proves skills marketplace is a viable discovery channel. |
| **Blog post #1** | "Why your AI architecture decisions are wrong (and how to fix them)" | Content marketing targeting post-vibe developers |
| **README overhaul** | Show, don't tell. Before/after comparison. Time-to-value metric. "Install → first research in 5 minutes." | Current README is documentation pitch. Needs to be a product pitch. |
| **GitHub Discussions** | Enable. Categories: "Show & Tell", "Q&A", "Feature Requests" | Zero feedback channels = zero adoption signal |
| **Metrics tracking** | GitHub stars, GoReleaser download counts, MCP Registry installs | Must measure what you intend to improve |

---

## Part 3: Decision Log (Finalized)

All 4 core architectural decisions codified as D-011 through D-014 (continuing the repository registry sequence after D-001..D-010 in `meta-research/DECISIONS.md`). Total of 11 locked decisions:

| # | Decision | Type | Status | Key Evidence |
|---|---|---|---|---|
| 1 | **MCP server replaces standalone Engine** | One-Way → Two-Way | **Locked** | MCP universality, compound failure risk of Engine, Two-Way Door economics |
| 2 | **Go for server language** | Two-Way | **Locked** | Distribution (single binary), Go SDK Tier 1, startup time, `os.Root` security |
| 3 | **Guided Worker pattern** (D-011) | One-Way | **Locked** | MCP stateless spec, P8 ("not rigid stage gates"), production MCP patterns |
| 4 | **Workspace files as canonical state** | One-Way | **Locked** | P6 (dual-audience), no shadow state sync bugs, Git compatibility. Beads failure case study validates: JSONL+SQLite had integrity/concurrency bugs |
| 5 | **9 tools with `vivechak_` prefix** (D-013) | One-Way | **Locked** | Code-level audit (2026-09-27) dropped `vivechak_synthesize` (violates Guided Worker), renamed `get_generator_prompt` → `prepare_generator`, `save_pipeline` → `save_plan`. See `mcp-tool-audit.md`. |
| 6 | **3 scope levels: Project / Decision / Comparison** (D-012) | Two-Way | **Locked** | Cross-domain validation (Cochrane, ICD 203, PRISMA), R-03/R-06 evidence |
| 7 | **Comparison = single-prompt WEP generator/template, not routing engine** | Two-Way | **Locked** | No decomposition needed; always 1 session; generator routing overhead unjustified |
| 8 | **Decision profiling: R/N/B/X (4 dimensions)** | Two-Way | **Locked** | 8-dim project scoring nonsensical for single decisions; R-06 refinement of R-03's 6-flag profile |
| 9 | **Binary-first distribution** (D-014) | Two-Way | **Locked** | Agent Plugins not universal (Claude Desktop, Antigravity excluded); PATH truncation |
| 10 | **Monorepo** | Two-Way | **Locked** | Server reads generators/templates directly; sync overhead at current scale |
| 11 | **Don't bump VERSION for Phases 0–3** | Two-Way | **Active** | Delivery improvements, not methodology principle changes |

---

## Part 4: The Anti-Plan — What NOT To Do

| # | Don't | Why |
|---|---|---|
| 1 | **Don't build a standalone Engine** | Replaced by MCP server. Framework choice (LangGraph/ADK/CrewAI) is volatile One-Way Door. |
| 2 | **Don't have the server call LLM APIs** | Host agent already has LLM access. Server managing API keys = friction, security risk, duplicated capability. |
| 3 | **Don't use dynamic tool registration** | MCP 2026-07-28 mandates static `tools/list`. GitHub deleted dynamic discovery as tech debt (May 2026). |
| 4 | **Don't maintain shadow state** | `.vivechak/state.json` creates sync bugs with `research/` directory. Workspace files ARE the state. |
| 5 | **Don't add more methodology documentation** | 106KB is already too much. New generators ADD entry points but don't add documentation volume. |
| 6 | **Don't chase Phase 6 frontiers** | DSPy, knowledge graphs, multi-agent debate — all gate conditions unmet. Require MCP usage data. |
| 7 | **Don't actively pursue non-software domains** | Keep "technical projects" language but don't build domain-specific generators. Zero demand signal. |
| 8 | **Don't over-invest in Agent Plugins packaging** | Claude Desktop and Antigravity don't support it. It's supplementary, not primary. |

---

## Part 5: Kill Criteria

| Signal | Timeframe | Response |
|---|---|---|
| **MCP server shipped, < 50 downloads after 3 months** | Phase 4 + 3 months | Serious reassessment. Is it visibility or product-market fit? Try one pivot (different positioning, different channel). If still < 50 after 6 months total, **archive the project.** |
| **Deep research tool ships native multi-session DAG with evidence grading** | Any time | Pivot moat messaging entirely to "auditable rigor + exit gates." Accelerate MCP server if not shipped. |
| **Go MCP SDK has breaking v2.0 release** | Any time | Pin v1.x. Evaluate migration cost. Don't block shipping. |
| **Agent Plugins 1.1.0 adds Claude Desktop support** | Any time | Elevate Agent Plugin from Week 3 to Day 1 channel. |
| **0 GitHub Discussions after 1 month of enablement** | Phase 4 + 1 month | Marketing problem, not product problem. Invest in content before building more features. |

---

## Part 6: The First 72 Hours

### Day 1 (Phase 0 — Bug Fixes)
1. Open `templates/CONFLICT-RESOLUTION.template.md`. Transpose ACH matrix: hypotheses → rows, evidence → columns. Update explanatory text.
2. Open `GENERATOR.md` line ~181. Change `A (official docs/RFCs)` to `A (official docs/RFCs/peer-reviewed studies)` to match FRAMEWORK.md §5.1.
3. Open `ROADMAP.md` Phase 5 section. Replace Engine plan with MCP server direction. Reference this plan and research corpus.
4. Submit Apple Developer ID and Azure Trusted Signing enrollment (lead time 1–20 business days).
5. Commit: `fix: ACH matrix orientation, evidence grade consistency, roadmap alignment`

### Day 2 (Phase 1 — Start Methodology)
6. Begin drafting `GENERATOR-DECISION.md` — use R-06 with-repo-link session as starting draft (near-complete 7.9KB prompt).
7. Begin drafting `templates/COMPARISON-SESSION.template.md` and `GENERATOR-COMPARISON.md` — use R-06 with-repo-link for structure.
8. Start `FRAMEWORK.md` §9 outline: Scope × Depth model, invariants I1–I9, routing matrix.

### Day 3 (Phase 1 — Continue + Logistics)
9. Verify Apple Developer ID and Azure Trusted Signing enrollment tickets; monitor verification status.
10. Continue generator and template drafts.
11. Test GENERATOR-DECISION.md by running it on a real decision (dogfooding).

---

## Part 7: Risk Register

### Existential Risks

| ID | Risk | Mitigation |
|---|---|---|
| **EX-1** | Deep research tools gain native multi-session orchestration within 6–12 months | Compete on rigor-as-discipline, not orchestration-as-mechanism. The moat is auditable methodology, not "we coordinate multiple sessions." |
| **EX-2** | Zero adoption — 0 stars, 0 forks, 0 watchers despite months of availability | Ship MCP server + free wedge skill + case study. Set 3-month kill criterion. |
| **EX-3** | Claude Code Skills marketplace produces a "good enough" clone (Claude-Code-Deep-Research already has A–E grading) | Differentiate on exit gate, locked workflow integrity, and MADR-compatible ADR output. |

### High-Priority Risks

| ID | Risk | Mitigation |
|---|---|---|
| **HI-1** | Generator drift across 3 files | `CORE:BEGIN/END` blocks verified by CI diff/checksum. Danish Mini-HTA 60-form sprawl provides empirical evidence this is necessary. |
| **HI-2** | macOS GUI PATH truncation breaks bare commands | `vivechak mcp-config --client <host> --write` generates absolute path configs. |
| **HI-3** | Claude Code silently truncates tool descriptions at 2,048 chars | Front-load decisive action in first 300–500 chars. |
| **HI-4** | Validation strictness kills adoption | Severity-based validation: Block on structure, Warn on quality, Comply-or-Explain at gate. |
| **HI-5** | Path traversal attacks on workspace files | Go 1.25 `os.Root` for traversal-resistant access. 5 CVEs in Anthropic reference servers in 10 months. |

### Medium Risks

| ID | Risk | Mitigation |
|---|---|---|
| **MD-1** | MCP spec continues evolving rapidly | Pin Go SDK v1.x; implement CI "toolsnaps" |
| **MD-2** | 63% of ADRs rubber-stamped (ICSA 2026) | Vivechak's gate enforcement makes this a feature |
| **MD-3** | Apple Developer ID required for macOS Homebrew cask | Start enrollment immediately on Day 1 (Phase 0.4) |
| **MD-4** | Registry stagnation (62% of MCP servers published once) | CI step recalculating `fileSha256` on every release |
| **MD-5** | Code signing identity verification stalls (1–20+ business days) | Ship unsigned community binary with Homebrew quarantine-strip hook and Scoop/winget portable zip as v1.2 fallback |

---

## Part 8: Open Backlog (Carried Forward)

Items from the existing ROADMAP backlog, updated with current analysis:

| ID | What | Status | Target |
|---|---|---|---|
| SM-01 | Domain/Technical Novelty dimension overlap risk | Open — needs calibration data | Post-Phase 4 |
| SM-02 | Sharp tier boundary cliff at 15→16 | Open — needs calibration data | Post-Phase 4 |
| GA-02 | No protocol for human stakeholder disagreement | Open — low urgency | Post-Phase 4 |
| GA-04 | Quality rubric evaluation process undefined | Partially addressed by `vivechak_validate` | Phase 2 |
| ES-03 | Missing "tool-generated" verification method | Open | Post-Phase 4 |
| OP-02 | No worked example of a failed pipeline | Open — create when failure data available | Phase 4 (if failure occurs) |
| OP-03 | Positioning vs Structured MADR 1.0 | Scheduled | Phase 4 |
| DA-03 | Context injection token budget not quantified | Open — quantify from MCP usage | Post-Phase 4 |
| PE-01 | "Context engineering" terminology | Scheduled | Phase 1.7 |
| NEW-01 | FAD Amendment Protocol | Open | Post-Phase 4 |
| NEW-02 | Narrow "10× cost at month 6" claim | Scheduled | Phase 1.4 (README update) |
| NEW-03 | Re-evaluate P1 decomposition threshold (1M+ context) | Open — evidence-gated | Post-Phase 4 |
| NEW-04 | Re-evaluate P7 triangulation value (model diversity) | Open — evidence-gated | Post-Phase 4 |
| NEW-05 | Monitor deep research tools for multi-session capability | Ongoing | Ongoing |
| NEW-06 | ~~**9 vs 10 tools**~~ | **Resolved** — code-level audit (2026-09-27) confirmed: `vivechak_synthesize` can't synthesize without LLM, violates Guided Worker. Dropped. SYN-01 handled by `next_session` + `save_session`. | Closed |
| NEW-07 | **SHA-256 content-addressing** for artifact staleness detection (without-repo R-02). Enables automatic "this session's cited sources have changed" warnings. | Open — from full audit | Post-Phase 4 |
| NEW-08 | **PRISMA-ScR-style coverage checklist** for research sessions (without-repo R-01). 27 items for full, 20+2 for scoping reviews. | Open — from full audit | Post-Phase 4 |
| NEW-09 | **Public calibration tracking** — report % of `prediction` fields that held up at `review_date`. Genuinely unique feature no competitor has (with-repo R-01). | Open — from full audit | Phase 4 |
| NEW-10 | **Stale-reference detection** — when `vivechak_record_decision` supersedes D-NNN, warn about sessions/decisions referencing the old D-NNN (with-repo R-01). | Open — from full audit | Post-Phase 2 |
| NEW-11 | **`C-NNN` ID namespace** for standalone comparisons not tied to a decision (R-06 OQ-4). | Open — from full audit | Phase 1 |
| NEW-12 | **Response envelope evaluation** — consider without-repo R-04's richer `{ok, summary, issues, next_actions, data, meta}` envelope with `fix_hint` per issue and plural `next_actions`. | Open — from full audit | Phase 2 |

---

## Appendix A: Document Lineage

This plan supersedes and consolidates:

| Document | Location | Status |
|---|---|---|
| `what-next-old.md` | `temp/` | **Superseded** — strategic direction confirmed (~70% accuracy); Worker tools killed |
| `WHATS-NEXT.md` | `temp/` | **Superseded** — best pre-research doc (~80% accuracy); comparison delivery changed |
| `ROADMAP-NEXT.md` | `temp/` | **Superseded** — 4-wave structure confirmed; every tactical detail changed |
| `PYTHON-VS-GO.md` | `temp/` | **Superseded** — verdict (Go) incorporated |
| `DECISIONS.md` | `temp/research/` | **Superseded** — Codified as D-011→D-014 in Decision Log above to avoid collision with repository meta-research D-001..D-010 |
| `research-synthesis.md` | artifact directory | **Reference** — evidence base with full decision verdicts |
| `research-impact.md` | artifact directory | **Reference** — before/after comparison |
| 12 research sessions | `temp/research/sessions/` | **Archived evidence** — ~770KB primary corpus |
| `RESEARCH-PIPELINE.md` | `temp/research/` | **Reference** — pipeline definition that drove the 12 sessions |

## Appendix B: Evidence Sources (Highest Confidence)

| Source | Grade | Used For |
|---|---|---|
| MCP 2026-07-28 specification | A | Stateless protocol, static tool lists, caching |
| Go MCP SDK (Tier 1, Google co-maintained) | A | Server implementation foundation |
| Agent Plugins 1.0.0 specification (Aug 2026) | A | Distribution packaging format |
| Dhami et al. 2024 (ACH matrix orientation) | A | CONFLICT-RESOLUTION template fix |
| Zheng et al. EMNLP 2024 (persona debunking) | A | P4 stability validation |
| Laban et al. ICLR 2026 (drip-feed penalty) | A | Prompt anatomy stability |
| Kim et al. ICML 2025 (correlated LLM errors) | A | P7 evaluation |
| Cochrane/Marshall 2019 / Nussbaumer-Streit 2023 | A | Scope reduction quality risks (6.5–38.6% significance shift) |
| ODNI ICD 203 (intelligence analysis standards) | A | Cross-domain scope invariants |
| Kidholm/Ehlers (Danish Mini-HTA) | B | Multi-generator drift empirical evidence |
| Spec Kit GitHub (~110K–120K+ stars) | B | Market demand validation |
| Claude-Code-Deep-Research (liangdabiao) | B | Commoditization threat assessment |
| Stack Overflow 2025 (84% AI adoption, 33% trust) | B | Target audience validation |
| ICSA 2026 (63% ADR rubber-stamping) | B | Gate enforcement value proposition |
| 5 CVEs in Anthropic reference MCP servers | B | Security requirement justification |
