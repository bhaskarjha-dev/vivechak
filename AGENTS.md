# Vivechak (विवेचक) — Agent Operating Manual
### Vivechak v0.1.0 · Single Source of Truth for AI Agent Operations

> **Read this file COMPLETELY before doing anything in this repository.**

---

## 1. What This Repository Is

**Vivechak** is an epistemic grounding meta-framework that generates evidence-graded research for technical decisions across all stages of a project's lifecycle. It operates across three distinct scope levels:
- **Comparison Scope (`ScopeComparison`):** Rapid evaluation between 2–5 options (→ WEP matrix in 1 session, ~5 minutes, 0 DAG, 0 gate).
- **Decision Scope (`ScopeDecision`):** Architectural dilemma in an active codebase (→ ADR in 1–3 sessions, ~30 minutes, 0 project DAG, 0 FAD).
- **Project Scope (`ScopeProject`):** Greenfield architecture or major rewrite (→ FAD across 4–30 DAG sessions, Phase 0 Gate).

**The methodology generators are [GENERATOR.md](GENERATOR.md), [GENERATOR-DECISION.md](GENERATOR-DECISION.md), and [GENERATOR-COMPARISON.md](GENERATOR-COMPARISON.md).** The MCP server (`cmd/vivechak/`, binary alias `vck`) automates orchestration — workspace initialization, scope inference, DAG resolution, context injection, validation, and exit gates. Everything else supports them.

---

## 2. Using Vivechak Across the 3 Scopes

An agent uses Vivechak by choosing the scope level that matches the decision task:

| Dimension | 1. Comparison Scope (`ScopeComparison`) | 2. Decision Scope (`ScopeDecision`) | 3. Project Scope (`ScopeProject`) | 4. OUT OF SCOPE |
|---|---|---|---|---|
| **Objective** | Compare 2–5 concrete options | Lock an architectural decision (ADR) | Full system architecture (FAD) | Write code, debug, explain syntax |
| **Typical Prompts** | *"Pino vs Winston"*, *"Zod vs Valibot"* | *"gRPC vs REST for internal services"* | *"Architect real-time streaming pipeline"* | *"Fix this type error"*, *"Write a function"* |
| **Sessions & Time**| 1 session (~5 minutes) | 1–3 sessions (~30 minutes) | 4–30 sessions (Hours–Days) | 0 sessions (Direct execution) |
| **DAG Overhead** | **Zero** (no DAG created or required) | **Zero** (no project DAG or topology) | **Full DAG** (topological dependency resolution) | None |
| **Gate Overhead**| **None** | Optional Track B check | **Mandatory** Track A / Track B Phase 0 Gate | None |
| **Templates Used**| 1 (`COMPARISON-SESSION.template.md`) | 2 (`DECISIONS`, `CONFLICT-RESOLUTION`) | 6 (All templates) | None |
| **Primary Output**| `research/C-NNN-comparison.md` (WEP) | `research/D-NNN.md` (Locked ADR) | `research/FAD.md` (Founding Architecture) | Code files / direct response |

---

### 2.1 Comparison Scope (1 Session → WEP Matrix)
Use when asked to compare, choose, or benchmark 2–5 named technologies, libraries, algorithms, or datastores (e.g., *"Compare Pino vs Winston"*, *"Zod vs Valibot"*).

**MCP Flow:**
1. Call `vivechak_prepare_generator(scope="comparison", context="...")`.
2. Run the returned prompt in a research session with web search enabled.
3. Call `vivechak_save_plan(content=<output>, scope="comparison")`. Done! Output is saved to `research/C-NNN-comparison.md`.

**Manual Flow:**
1. Open [GENERATOR-COMPARISON.md](GENERATOR-COMPARISON.md) and extract the prompt.
2. Fill the `OPTIONS:`, `CRITERIA:`, and `CONTEXT:` slots.
3. Execute in an AI session with web search.
4. Save the Weighted Evaluation Protocol (WEP) matrix output into `research/COMPARISON.md`.

---

### 2.2 Decision Scope (1–3 Sessions → ADR)
Use when resolving an architectural fork, migration, or dilemma in an existing project (e.g., *"Should we migrate our event storage from PostgreSQL to ClickHouse?"*, *"Auth0 vs Clerk vs custom auth"*).

**MCP Flow:**
1. Call `vivechak_prepare_generator(scope="decision", context="...")`.
2. Save the decision plan via `vivechak_save_plan(scope="decision", decision_id="D-NNN", slug="...", content=...)`.
3. Execute the 1–3 sessions (`L` Landscape, `C` Comparison, `F` Falsification) via `vivechak_next_session` and `vivechak_save_session`.
4. (Optional) Stress-test with `vivechak_challenge(decision_id="D-NNN", mode="red_team")`.
5. Lock the decision via `vivechak_record_decision(decision_id="D-NNN")`. Output is saved as canonical ADR `research/D-NNN.md` and compiled into `research/DECISIONS.md`.

**Manual Flow:**
1. Open [GENERATOR-DECISION.md](GENERATOR-DECISION.md) and copy prompt with context.
2. Copy `DECISIONS.template.md` and `CONFLICT-RESOLUTION.template.md` to `research/templates/`.
3. Execute the 1–3 sessions and record the ADR in `research/D-NNN.md`.

---

### 2.3 Project Scope (Full Pipeline → FAD)
Use when architecting a brand new project, platform, or massive subsystem rewrite from scratch.

**MCP Flow:**
1. Call `vivechak_init(scope="project")`.
2. Call `vivechak_prepare_generator(scope="project", context="<project vision>")`.
3. Execute generator prompt and call `vivechak_save_plan(content=...)` to save `RESEARCH-PIPELINE.md`.
4. Loop: `vivechak_next_session` → research → `vivechak_save_session` → `vivechak_record_decision`.
5. Run grand synthesis session `SYN-01` to compile `research/FAD.md`.
6. Run `vivechak_run_gate` to verify Track A and Track B exit criteria before starting implementation.

**Manual Flow:**
1. Open [GENERATOR.md](GENERATOR.md) and extract prompt inside ```` ````markdown ```` fences.
2. Replace `[PASTE YOUR PROJECT DESCRIPTION HERE]` with vision and execute with web search.
3. Save `RESEARCH-PIPELINE.md` and `DECISIONS.md` to `research/`. Copy all 6 templates from `templates/` to `research/templates/`:
   - `DECISIONS.template.md` — for recording architectural decisions
   - `CONFLICT-RESOLUTION.template.md` — for resolving conflicting findings
   - `COMPARISON-SESSION.template.md` — for structuring option comparisons
   - `FOUNDING-ARCHITECTURE.template.md` — for compiling the final FAD
   - `PHASE-0-GATE.template.md` — for the pre-coding exit gate
   - `SESSION.template.md` — for standardized research session outputs
4. Execute session DAG, record ADRs, synthesize `FAD.md`, and pass `PHASE-0-GATE.template.md`. Once the gate passes — start coding.

> **You do NOT need to read FRAMEWORK.md to use Vivechak.** The generator prompts are self-contained — all methodology is operationalized inline.

---

### 2.4 When NOT to Use Vivechak (Anti-Pattern Guardrail)
Vivechak must **never** be invoked for:
1. **Direct Code Authoring:** Writing functions, components, API endpoints, or scripts.
2. **Syntax & Language Inquiries:** Language grammar, type signatures, API parameters.
3. **Local Bug Fixing & Troubleshooting:** Inspecting stack traces, resolving compiler errors, fixing tests.
4. **Micro-Refactoring & Code Cleanup:** Renaming variables, extracting helpers, reformatting.
5. **Non-Architectural Conventions:** Choices with zero reversal risk and zero lasting impact.

> **Rule for Agents:** If the user's intent is to write code or debug an error, write the code or diagnose the issue directly. Do NOT invoke Vivechak tools.

---

## 3. Developing Vivechak Itself

This section applies only when improving the meta-framework — editing FRAMEWORK.md, GENERATOR.md, or templates/.

1. **Read [FRAMEWORK.md](FRAMEWORK.md)** — the complete methodology specification (8 principles, evidence grading, pipeline topology, prompt anatomy).
2. **Read [CONTRIBUTING.md](CONTRIBUTING.md)** — all changes must be traceable to evidence.
3. **⚠️ Run the Change Propagation Map** (CONTRIBUTING.md § Change Propagation Map) before committing. FRAMEWORK.md is the specification; GENERATOR.md is the implementation. Changing the spec without updating the implementation means improvements exist in documentation but never affect generated pipelines. Every change to FRAMEWORK.md, GENERATOR.md, or templates/ MUST be checked against the propagation map.
4. All changes must follow the 8 Core Principles (summarized below).
5. The sealed `docs/meta-research/` directory contains the empirical evidence base — 15 research artifacts validating every design decision.

### The 8 Core Principles

| # | Principle | Summary |
|---|---|---|
| P1 | Context Architecture Law | Decompose by attention budget and coupling, not arbitrary counts. Mandate synthesis after decomposition. |
| P2 | Reversibility-Calibrated Rigor | Deep research for One-Way Doors, fast spikes for Two-Way Doors. |
| P3 | Evidentiary Grounding | Every claim carries evidence grade (A–E), modifiers, and verification method. Recalled = Grade D cap. |
| P4 | Prescriptive Scope, Dynamic Method | Prescriptive on WHAT/WHY/BOUNDARIES. Directional on HOW. No hardcoded queries or personas. |
| P5 | Commodity-Maximized Composition | Compose 100% commodity infra. Build 100% custom only for proprietary domain logic. |
| P6 | Dual-Audience Artifacts | Hybrid Markdown + YAML frontmatter for human reading and machine synthesis. |
| P7 | Staged Triangulation | Single-model default → critique probe → full triangulation only for contested One-Way Doors. |
| P8 | Structured Falsification | Prioritize disconfirming evidence, document rejected alternatives, schedule review triggers, premortems. |

---

## 4. Repository Structure

```
vivechak/
├── README.md                       ← Overview + install + quickstart
├── GENERATOR.md                    ← Project-scope generator (full pipeline → FAD)
├── GENERATOR-DECISION.md           ← Decision-scope generator (1–3 sessions → ADR)
├── GENERATOR-COMPARISON.md         ← Comparison-scope generator (1 session → WEP matrix)
├── FRAMEWORK.md                    ← Complete methodology spec (for development only)
├── AGENTS.md                       ← This file
├── ROADMAP.md                      ← Project history + future plans
├── CHANGELOG.md                    ← Release history
├── CONTRIBUTING.md · SECURITY.md   ← Contribution & security policies
├── CODE_OF_CONDUCT.md · LICENSE    ← Contributor Covenant v2.1 · MIT License
├── go.mod · go.sum                 ← Go 1.25+ module
├── Makefile                        ← Standard build, test, lint & embed-sync targets
├── .goreleaser.yml                 ← 7-platform cross-compilation + Linux nfpms
├── Dockerfile · flake.nix          ← Multi-arch container & Nix flake
├── install.sh · install.ps1        ← Shell installers
├── npm/ · pypi/                    ← Zero-script npx & uvx runner packages
│
├── cmd/vivechak/                   ← MCP server + CLI entry point (vck alias)
│   ├── main.go                     ← Entry point (serve, setup, mcp-config, doctor, gate, version)
│   ├── config.go                   ← setup & mcp-config (auto-detection + 21 presets)
│   ├── doctor.go                   ← Workspace integrity checker
│   └── gate.go                     ← Headless Phase 0 exit gate evaluator (CI-ready)
│
├── internal/                       ← Server implementation (Go)
│   ├── core/                       ← Pure logic (workspace, DAG, validation, injection)
│   ├── mcp/                        ← 13 MCP tool handlers + envelope
│   ├── store/                      ← Atomic I/O, locking, os.Root confinement
│   └── embed/                      ← go:embed generators + templates
│
├── templates/                      ← 6 operational contracts (copy to projects)
│   ├── DECISIONS.template.md       ← YAML frontmatter ADR format
│   ├── CONFLICT-RESOLUTION.template.md ← ACH-style conflict resolution
│   ├── COMPARISON-SESSION.template.md  ← WEP comparison output format
│   ├── FOUNDING-ARCHITECTURE.template.md ← Map-Reduce synthesis to FAD
│   ├── PHASE-0-GATE.template.md    ← Two-track pre-codebase exit gate
│   └── SESSION.template.md         ← Standard research session output format
│
├── examples/                       ← Concrete adoption walkthroughs
│   ├── SAMPLE-COMPARISON.md        ← Comparison scope walkthrough (Pino vs Winston WEP matrix)
│   ├── SAMPLE-DECISION.md          ← Decision scope walkthrough (ClickHouse migration ADR)
│   └── SAMPLE-PROJECT.md           ← Project scope walkthrough (Katha full pipeline)
│
├── docs/                           ← Documentation
│   ├── QUICKSTART.md               ← Vivechak in 5 Minutes
│   ├── MANUAL-WORKFLOW.md          ← Complete manual copy-paste guide
│   ├── HOST-SETUP.md               ← Universal MCP setup & desktop shortcuts
│   ├── TROUBLESHOOTING.md          ← Diagnostic guides & error code catalog
│   ├── MCP-TOOLS.md                ← 13-tool reference with examples
│   ├── ARCHITECTURE.md             ← Server internals for contributors
│   ├── BRAND.md                    ← Brand identity, geometry & design tokens
│   ├── HARNESS-COMPATIBILITY.md    ← Compatibility matrix across 8 harness categories
│   ├── PARALLEL-EXECUTION.md       ← Concurrent subagents & locking guide
│   ├── HYBRID-EXECUTION.md         ← Native hybrid (MCP + manual) operational guide
│   ├── assets/logo.svg             ← Minimalist Devanagari V-sieve logo mark
│   └── meta-research/              ← Empirical evidence base
│       ├── README.md               ← Provenance index
│       ├── DECISIONS.md            ← All 14 locked ADRs (D-001–D-014)
│       ├── RESEARCH-PIPELINE-v1.md ← Initial methodology meta-research DAG
│       ├── v2/                     ← MCP server & multi-scope research evidence
│       │   ├── FINAL-PLAN.md       ← 5-phase server implementation plan
│       │   └── RESEARCH-PIPELINE-v2.md ← Multi-scope & MCP server research pipeline
│       └── v3/                     ← Production dogfooding, CLI ergonomics & UX evidence
│           ├── D-UX-01-cli-design.md
│           ├── D-UX-02-validation-depth.md
│           └── D-UX-03-hybrid-workflow.md
│
├── tests/harness/                  ← Autonomous multi-harness stress-testing subsystem
│   ├── README.md                   ← Architecture, commands & lifecycle policy
│   ├── SCENARIOS.md                ← 21 human-realistic stress scenarios
│   ├── TEST-PLAN.md                ← Autonomous test harness strategy & execution model
│   ├── runner/                     ← Watchdog, dynamic triage & pluggable adapters
│   ├── reports/                    ← Permanent milestone audit reports (audit-v0.1.0-...)
│   ├── defects/DEFECTS.md          ← Living defect and polish ledger
│   └── runs/                       ← Ephemeral local test runs (gitignored)
│
├── scripts/run_harness.mjs         ← Multi-harness test runner CLI (--auto, --probe, etc.)
│
├── Formula/ · scoop/ · winget/     ← Package manager manifests
│
└── .github/
    ├── workflows/ci.yml · release.yml
    ├── ISSUE_TEMPLATE/
    └── PULL_REQUEST_TEMPLATE.md
```

---

## 5. Working on the MCP Server

When modifying Go code in `cmd/` or `internal/`:

1. **Package dependency direction:** `core/` has NO dependency on `mcp/`. `mcp/` depends on `core/`. `store/` is independent.
2. **Embed sync:** `internal/embed/generators/` and `internal/embed/templates/` must stay in sync with root-level generators and templates. Verified by `TestEmbeddedFilesMatchRoot` and enforced by CI (`.github/workflows/ci.yml`).
3. **13 MCP tools:** `vivechak_init`, `vivechak_prepare_generator`, `vivechak_save_plan`, `vivechak_status`, `vivechak_next_session`, `vivechak_save_session`, `vivechak_record_decision`, `vivechak_amend_session`, `vivechak_validate`, `vivechak_run_gate`, `vivechak_challenge`, `vivechak_replan`, `vivechak_visualize`. Each tool is one file in `internal/mcp/tool_*.go`. Always provide 'scope' ('project', 'decision', or 'comparison') explicitly in `vivechak_init` and `vivechak_prepare_generator`; `tool_save_plan` auto-initializes the workspace on the fly if `vivechak_init` was skipped, auto-recovers decision IDs, and saves comparison prompts/matrices without DAG constraints; `tool_challenge` resiliently accepts `decision_id` (alias for `session_id`) and `challenge_type` (alias for `mode`), supporting two-tier lookup into `research/` for standalone ADRs (`D-*.md`) and comparison matrices (`C-*.md`); `tool_amend_session` accepts `content` and `reason` aliases with two-tier lookup for standalone ADRs and comparison matrices; `tool_validate` in workspace mode auto-discovers and validates decision plans (`D-*-plan.md`), standalone ADRs, and comparison matrices (`C-*-comparison.md`); `tool_save_session` surfaces dedicated `vivechak_run_gate` next_step upon FAD synthesis save, and provides `auto_draft_from` one-liners for ADR recording.
4. **Guided Worker pattern:** Every tool response includes `next_step`. Hard refusals ONLY for impossible operations, never for "wrong order."
5. **Quality coaching:** `core.ObserveSessionQuality` runs inside `vivechak_save_session`, surfacing real-time non-blocking `Q-*` observations (grade inflation, missing discovered concerns, Delta confirmation bias, low findings count) directly in `next_step`.
6. **Single-source ADRs:** `vivechak_record_decision` treats individual `D-*.md` files as canonical and compiles the consolidated `research/DECISIONS.md` registry.
7. **Automated Phase 0 Gate:** `vivechak_run_gate` accepts optional `track` filter (`track_a`, `track_b`, or `all`) and mechanically evaluates decoupled Track A (reversibility confirmation for Two-Way Doors) and Track B (rigorous checks B1–B11 for One-Way Doors, including B3 threshold, B4 verification, B5 rejected alternatives, B6 reversal triggers, B10 substantive content, and B11 adversarial red-team challenge), returning `track_a_result`, `track_b_result`, and `decisions` array while auto-persisting `research/PHASE-0-GATE.md`.
8. **Concurrency, Status Sync & Atomic I/O:** `store.WriteFileAtomic` employs monotonic `tmpFileCounter` to prevent Windows nanosecond collision races; mutating operations acquire advisory locks; `tool_save_session` uses `core.SetFrontmatterField` to enforce `status: draft` on validation defects and promote valid sessions to `status: complete`; `core.ScanCompletedSessions` respects frontmatter `status: draft` and receives DAG `knownIDs...` to resolve slugged filenames without explicit frontmatter IDs.
9. **Special research files:** Use `core.IsSpecialResearchFile` when scanning `research/` to prevent non-ADR markdown files from corrupting decision registries.
10. **Testing:** `go test ./...` runs all tests. `internal/mcp/server_test.go` and `internal/mcp/massive_real_lifecycle_test.go` have wire-level JSON-RPC and end-to-end multi-scope lifecycle tests; `cmd/vivechak/*_test.go` contains live subprocess integration, concurrency, and workspace resolution tests.

---

## 6. Key Operational Rules

- **No legacy dogma:** Do not use 8-section XML prompts, expert personas, hardcoded search queries, minimum search counts, or rigid output skeletons. These are empirically refuted.
- **Front-load briefs:** Never drip-feed instructions across turns (39% performance drop documented).
- **Grade everything:** Every factual claim needs an inline evidence grade with modifiers and verification method.
- **Qualified provenance:** Grade A claims require canonical URLs, domain anchors, RFCs, or specific doc titles. Computing memory (RAM/caching) is never penalized as model recall.
- **Two-Way Doors move fast:** Don't over-research reversible decisions. Spike or decide by convention.
- **One-Way Doors move carefully:** Require corroborated Grade A/B evidence with qualified provenance, substantive Discovered Concerns (≥50 chars), red-team challenge before commitment, locked ADR with rejected alternatives, and premortem.

---

## 7. Autonomous Multi-Harness Testing Subsystem

Vivechak includes an empirical multi-harness autonomous testing engine in `tests/harness/` to stress-test Vivechak against diverse agent hosts, models, and headless MCP clients.

### Key Operational Workflows:
1. **Probe Active Harnesses & Quota:**
   `make harness-probe` or `node scripts/run_harness.mjs --probe`
   *Checks all registered harnesses in parallel for binary presence, auth status, and active token quota.*
2. **Auto-Run Across All Funded/Available Harnesses:**
   `make harness-test` (core batch) or `node scripts/run_harness.mjs --auto --batch all`
   *Runs tests only on harnesses with active tokens/credits.*
3. **Run Zero-Token Headless MCP Conformance Test (CI / Offline):**
   `node scripts/run_harness.mjs --harness direct_mcp --scenario 01`
   *Tests all 13 tools deterministically over JSON-RPC stdio with zero external token expenditure.*
4. **Ingest Dynamic Triage & Extract Polish:**
   `make harness-triage` or `node scripts/run_harness.mjs --triage`
   *Dynamically parses runs, evaluates pass/fail status, and clusters Turn 4 feedback.*
5. **Clean Workspace & Temp Sandboxes for Fresh Iteration:**
   `make harness-clean` (wipes ephemeral runs in `tests/harness/runs/`)
   `make harness-clean-sandboxes` (wipes isolated test sandboxes in OS temp directory).
6. **Archive Release Milestone Audit:**
   `node scripts/run_harness.mjs --archive <version>`
   *Compiles active runs into an immutable audit report in `tests/harness/reports/audit-<version>-<date>.md` and updates index.*
7. **Living Defect Ledger:**
   All empirical findings, defects, and upstream engine fixes are permanently tracked in `tests/harness/defects/DEFECTS.md`.
8. **Universal Model Overrides & Automatic Failover Cascades:**
   Pass `--model <id>` to target specific models. If a model encounters quota exhaustion, rate limits, or timeouts, adapters automatically fail over to the next healthy model (e.g. `beam-free` → `big-pickle` → `nemotron-3.5-lightning-free` → `glm-5.3-flash-free` for OpenCode, or `gemini-3.8-flash-high` → `gemini-3.8-flash` → `claude-sonnet-4-6` for Antigravity).
9. **Dead-Session Short-Circuiting & 21-Scenario Semantic Triage:**
   Turn 1 failures short-circuit Turns 2–4 to prevent wasting 270s on dead processes. Triage accurately evaluates semantic invariants across all 21 scenarios without false-negative file traps.

> **Rule for Agents:** Never commit raw run files in `tests/harness/runs/` to git. All empirical findings must be distilled into `tests/harness/defects/DEFECTS.md` or a release audit report.

