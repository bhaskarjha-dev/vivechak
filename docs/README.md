# Vivechak Documentation

Welcome to the documentation hub for **Vivechak** (*vivechaka* — "the discerning analyst") — the local-first Model Context Protocol (MCP) server and CLI (`vck` / `vivechak`) for evidence-graded technical decisions.

- 🌐 **Official Web Platform:** [vivechak.bhaskarjha.dev](https://vivechak.bhaskarjha.dev) &mdash; Interactive Playground, 1-Click Host Configurator & Visual Gate Inspector
- 📖 **Live Documentation Hub:** [vivechak.bhaskarjha.dev/docs](https://vivechak.bhaskarjha.dev/docs)
- 💻 **Source Repository:** [github.com/bhaskarjha-dev/vivechak](https://github.com/bhaskarjha-dev/vivechak)

---

## 1. Getting Started

| Guide | What It Covers |
|---|---|
| [**Quickstart (`QUICKSTART.md`)**](QUICKSTART.md) | End-to-end 5-minute walkthrough across MCP (`Path A`) and manual browser chat (`Path B`). |
| [**Host Setup (`HOST-SETUP.md`)**](HOST-SETUP.md) | Configuring Claude Code, Cursor, VS Code, Antigravity, Windsurf, Zed, Roo Code, OpenCode, Goose, and ephemeral `npx` / `uvx` / Docker runners. |
| [**Verified Host Presets (`hosts.md`)**](hosts.md) | Auto-generated table of all 10 verified `vck setup <preset>` targets, config file paths, schema keys, and upstream host docs. |
| [**Troubleshooting (`TROUBLESHOOTING.md`)**](TROUBLESHOOTING.md) | Running `vck doctor`, resolving workspace root errors (`E-ROOT-*`), and closing Phase 0 Exit Gate failures (`G1`–`G12`). |

---

## 2. Workflows & Execution Modes

| Guide | What It Covers |
|---|---|
| [**MCP Tools Overview (`MCP-TOOLS.md`)**](MCP-TOOLS.md) | Canonical response envelope, 7-step fail-closed workspace resolution ladder, and tool sequences across all 3 scopes. |
| [**Manual Workflow (`MANUAL-WORKFLOW.md`)**](MANUAL-WORKFLOW.md) | Running Vivechak copy-paste prompts in browser deep-research tools without installing the MCP server or CLI. |
| [**Hybrid Execution (`HYBRID-EXECUTION.md`)**](HYBRID-EXECUTION.md) | Combining MCP-automated sessions with browser deep-research runs and cross-vendor triangulation (Principle P7). |
| [**Parallel Execution (`PARALLEL-EXECUTION.md`)**](PARALLEL-EXECUTION.md) | Scaling multi-session DAG pipelines across concurrent subagents with `parallelism_hint`, advisory file locks, and atomic writes. |
| [**Harness Compatibility (`HARNESS-COMPATIBILITY.md`)**](HARNESS-COMPATIBILITY.md) | Integration matrix across IDEs, terminal agents, multi-agent frameworks, and headless CI/CD pipelines. |

---

## 3. Auto-Generated Reference Manuals (`docs/reference/`)

Generated deterministically from Go source registries via `go run ./tools/docgen`:

| Reference | Source of Truth | Contents |
|---|---|---|
| [**Tools Schema Reference (`reference/tools.md`)**](reference/tools.md) | `internal/mcp/tool_*.go` | All 9 MCP tools, annotations, required/optional parameters, and allowed enums. |
| [**Phase 0 Gate Rules (`reference/gate.md`)**](reference/gate.md) | `internal/core/gate_rules_spec.go` | Dual-track gate architecture (`Track A` / `Track B`) and mechanical verification criteria for `G1`–`G12`. |
| [**Diagnostic & Error Codes (`reference/codes.md`)**](reference/codes.md) | `internal/core/codes.go` | Every operational error (`E-*`), validation error (`V-*`), warning (`W-*`), and quality coaching code (`Q-*`) with `fix` instructions. |

---

## 4. Architecture, Methodology & Brand

| Document | Audience & Purpose |
|---|---|
| [**Framework Specification (`../FRAMEWORK.md`)**](../FRAMEWORK.md) | Complete epistemic methodology: the 8 Core Principles, Admiralty-derived `A`–`E` evidence grading, WEP scoring, and prompt anatomy. |
| [**Server Architecture (`ARCHITECTURE.md`)**](ARCHITECTURE.md) | Contributor guide to Go packages (`cmd/vivechak`, `internal/{core,workspace,mcp,store,embed,hosts,method}`), `os.Root` confinement, and adding tools. |
| [**Brand & Design System (`BRAND.md`)**](BRAND.md) | Parametric geometry ($S = 36$), optical sizing, color tokens, and dark/light theme contrast rules for [`assets/`](assets/). |
| [**Meta-Research Archive (`meta-research/`)**](meta-research/README.md) | Sealed empirical research artifacts and locked ADRs (`D-001`–`D-014`) backing Vivechak's design decisions. |
