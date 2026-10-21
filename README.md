<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.svg">
  <img src="docs/assets/logo.svg" alt="Vivechak" width="180" />
</picture>

# Vivechak (विवेचक)

**Evidence-grounded research for architectural, strategic, and technical decisions.**

*vi- (apart) + √vic (to sift/separate) + -aka (the agent who does) = "the discerning analyst"*

</div>

---

### v0.1.0 — Separating evidence from assumption, truth from bias

> **Transform high-stakes decisions from gut-feel and cached training data into structured, evidence-graded research — whether you're architecting a system or business venture, researching an irreversible decision, or comparing competing options, models, or algorithms.**

> **What you're installing:** A single standalone Go binary (`vck`) with **zero runtime dependencies**, **zero cloud phone-home**, and **zero telemetry**. It runs 100% locally on your machine as both a developer CLI (`vck doctor`, `vck setup`, `vck gate`) and an orchestration MCP server (`vck serve`) for AI coding agents.

---

## Install & Run

Vivechak installs both the canonical binary **`vivechak`** and its official shorthand **`vck`** across every platform:

### 1. Instant / Ephemeral Runners (Zero Pre-Installation Required)
```sh
# Node.js / npx (macOS, Linux, Windows)
npx -y vck serve                              # or: npx -y vivechak serve

# Python / uvx (macOS, Linux, Windows)
uvx vck serve                                 # or: uvx vivechak serve

# Docker / Cloud Agents (Devin, OpenHands, CI)
docker run -i --rm -v "${PWD}:/workspace" ghcr.io/bhaskarjha-dev/vck:latest serve
```

### 2. Native Package Managers
```sh
# Homebrew (macOS & Linuxbrew)
brew install bhaskarjha-dev/tap/vck           # or: brew install bhaskarjha-dev/tap/vivechak

# Windows (WinGet or Scoop)
winget install bhaskarjha-dev.Vivechak
scoop bucket add vivechak https://github.com/bhaskarjha-dev/scoop-bucket && scoop install vck

# Debian / Ubuntu (.deb)
sudo dpkg -i vivechak_0.1.0_linux_amd64.deb

# Fedora / RHEL (.rpm)
sudo rpm -i vivechak_0.1.0_linux_amd64.rpm

# Nix / NixOS
nix run github:bhaskarjha-dev/vivechak

# Direct Shell Installers
curl -fsSL https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.sh | sh      # macOS/Linux
irm https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.ps1 | iex           # Windows
```

### 3. Build from Source (Go 1.25+)
If you prefer not to use package managers or pre-built binaries, build directly from source with zero C dependencies:

```sh
# 1. Clone repository
git clone https://github.com/bhaskarjha-dev/vivechak.git
cd vivechak

# 2. Build both vck and vivechak binaries:
make build

# Or manually with production optimization flags:
# (-trimpath strips paths; -s -w strips debug symbols & DWARF info for ~35% smaller binary)
# macOS / Linux:
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=0.1.0" -o vck ./cmd/vivechak/
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=0.1.0" -o vivechak ./cmd/vivechak/

# Windows (PowerShell):
$env:CGO_ENABLED="0"; go build -trimpath -ldflags="-s -w -X main.version=0.1.0" -o vck.exe ./cmd/vivechak/
$env:CGO_ENABLED="0"; go build -trimpath -ldflags="-s -w -X main.version=0.1.0" -o vivechak.exe ./cmd/vivechak/

# 3. (Optional) Verify binary and test suite:
./vck version                                 # displays: vivechak 0.1.0 (or .\vck.exe version on Windows)
make test                                     # runs 100% of unit tests

# 4. (Optional) Install to your system PATH:
# macOS / Linux:
sudo cp vck vivechak /usr/local/bin/
# Windows (PowerShell - copy to user bin or add repo folder to PATH):
Copy-Item vck.exe, vivechak.exe -Destination "$env:USERPROFILE\bin\"
```

Or install directly into your `$GOPATH/bin` via the Go toolchain:
```sh
go install github.com/bhaskarjha-dev/vivechak/cmd/vivechak@latest
```

**Then configure your AI host:**
Vivechak installs the official 3-letter shorthand **`vck`** alongside `vivechak`. Configure your AI host in 1 second:
```sh
# Fast host setup (auto-resolves config paths and merges cleanly across 21 presets):
vck setup cursor                             # Cursor IDE
vck setup vscode                             # VS Code (Copilot Agent mode)
vck setup claude-code                        # Claude Code CLI (~/.claude.json or .mcp.json)
vck setup mcp                                # Universal portable standard (.mcp.json)
vck setup opencode                           # OpenCode terminal agent
vck setup continue                           # Continue.dev (.continue/mcpServers/)
vck setup goose                              # Block Goose (config.yaml)
vck setup claude                             # Claude Desktop
vck setup agy                                # Google Antigravity
vck setup                                    # auto-detect host in current workspace

# Output universal MCP JSON configuration:
vck mcp-config

# Verify workspace health and diagnostic readiness:
vck doctor

# Headless Phase 0 exit gate evaluation (CI-ready):
vck gate --track all
```

Done. Your agent now has 13 MCP tools for evidence-grounded research. See [Host Setup Guide](docs/HOST-SETUP.md) for harness instructions, 21 desktop shortcuts, and manual setup details.

---

## What Is This?

You're making a high-stakes decision — whether architecting a full system, designing a new business venture, making an irreversible strategic choice, or selecting between competing AI models, algorithms, or datastores. Some decisions are **irreversible** (one-way doors). If you get them wrong, you're looking at painful rewrites, failed migrations, or stranded capital.

**Vivechak generates evidence-grounded research at three scope levels:**

| Scope | Entry Point | What You Get |
|---|---|---|
| **Project** | [GENERATOR-PROJECT.md](GENERATOR-PROJECT.md) | Full research pipeline → Founding Architecture Document |
| **Decision** | [GENERATOR-DECISION.md](GENERATOR-DECISION.md) | 1–3 focused sessions → proposed ADR |
| **Comparison** | [GENERATOR-COMPARISON.md](GENERATOR-COMPARISON.md) | 1 structured comparison → WEP matrix |

Every scope level applies the same methodology: traceable evidence grades (A–E), structured falsification, bias-corrected weighted evaluation, and explicit reversal triggers on every decision.

**Vivechak is not a deep-research tool** — it is the **orchestration layer above** tools like Gemini Deep Research, ChatGPT Deep Research, and Perplexity. Those tools execute individual research sessions; Vivechak structures which sessions to run, how evidence is graded consistently across sessions, how decisions are tracked with reversal triggers, and how findings are synthesized into a coherent architecture.

---

## Quick Start

Vivechak supports two primary execution workflows: **Path A** with your AI coding agent via MCP (recommended, 1-prompt orchestration), or **Path B** manually in any browser AI chat.

### Path A: AI Agent Flow via MCP (Recommended)

Once you've configured your AI host (`vck setup <host>`):

1. **Prompt your agent in your project chat based on your need:**
   - **For Bounded Comparisons (~5 minutes, 1 session):**  
     > *"Vivechak compare: Pino vs Winston for high-throughput Node.js microservices."*  
     *(Agent runs `vivechak_prepare_generator(scope="comparison")` → executes with web search → saves Weighted Evaluation Matrix via `vivechak_save_plan`. Zero DAG, zero gate!)*
   - **For Architectural Decisions (~30 minutes, 1–3 sessions):**  
     > *"Vivechak decision: Should we migrate our analytics datastore from PostgreSQL to ClickHouse?"*  
     *(Agent runs `vivechak_prepare_generator(scope="decision")` → executes 1–3 focused sessions → locks canonical ADR `D-NNN.md` with reversal triggers via `vivechak_record_decision`. Zero project DAG, zero FAD!)*
   - **For Full Project Pipelines (Hours–Days, greenfield architecture):**  
     > *"Use Vivechak to research our architecture. We need to build a real-time event streaming pipeline in Go handling 50k events/sec with PostgreSQL and Kafka."*  
     *(Agent runs `vivechak_init(scope="project")`, prepares DAG pipeline, loops through sessions, synthesizes `FAD.md`, and verifies `vivechak_run_gate` before writing code.)*

2. **When NOT to use Vivechak (Direct Implementation):**
   - Skip Vivechak for direct code writing, syntax questions, local bug fixing, or trivial refactoring (e.g., *"Write a Go function to parse JWT"*). Your agent writes code directly without calling research tools.

---

### Path B: Manual Flow (Browser / No MCP Installation)

If you are using ChatGPT, Claude, or Gemini in your browser without an MCP agent harness:

1. **Choose Your Scope Generator:**
   - **Comparison:** Open [GENERATOR-COMPARISON.md](GENERATOR-COMPARISON.md), inject options & criteria, run in a web search chat, and save the resulting WEP matrix into `research/COMPARISON.md`.
   - **Decision:** Open [GENERATOR-DECISION.md](GENERATOR-DECISION.md), paste the dilemma, execute 1–3 sessions, and record the ADR in `research/D-NNN.md` using `templates/DECISIONS.template.md`.
   - **Project:** Open [GENERATOR-PROJECT.md](GENERATOR-PROJECT.md), paste your vision, save `RESEARCH-PIPELINE.md` (DAG) and `DECISIONS.md`, copy all 6 templates, execute sessions into `research/sessions/`, synthesize `FAD.md`, and pass `templates/PHASE-0-GATE.template.md` before coding.

---

## Repository Structure

> **For most users:** Install the binary, run `vck setup <host>`, and you're done. The repository structure below is for contributors and those who want to understand the internals.

```
vivechak/
├── cmd/vivechak/                   ← MCP server + CLI entry point
│   ├── main.go                     ← Entry point (serve, setup, mcp-config, doctor, gate, version)
│   ├── config.go                   ← setup & mcp-config (auto-detection + 21 presets)
│   ├── doctor.go                   ← Workspace integrity checker
│   └── gate.go                     ← Headless Phase 0 exit gate evaluator (CI-ready)
│
├── internal/                       ← Server implementation
│   ├── core/                       ← Pure logic (workspace, DAG, validation)
│   ├── mcp/                        ← 13 MCP tool handlers
│   ├── store/                      ← Atomic file I/O with os.Root confinement
│   └── embed/                      ← Embedded generators + templates
│
├── .goreleaser.yml                 ← 7-platform cross-compilation + Linux nfpms
├── Dockerfile · flake.nix          ← Multi-arch container & Nix flake
├── install.sh · install.ps1        ← Shell installers
├── npm/ · pypi/                    ← Zero-script npx & uvx runner packages
├── Formula/ · scoop/ · winget/     ← Package manager manifests
├── .github/workflows/              ← CI + release pipelines
│   ├── ci.yml                      ← Test + vet + build on push/PR
│   └── release.yml                 ← Decoupled release workflow
│
├── GENERATOR-PROJECT.md            ← Project-scope generator (full pipeline → FAD)
├── GENERATOR-DECISION.md           ← Decision-scope generator (1–3 sessions → ADR)
├── GENERATOR-COMPARISON.md         ← Comparison-scope generator (1 session → WEP matrix)
├── FRAMEWORK.md                    ← Deep methodology reference
├── docs/HOST-SETUP.md              ← Universal MCP setup & desktop shortcuts
├── docs/TROUBLESHOOTING.md        ← Diagnostic guides & error code catalog
│
├── templates/                      ← 6 operational contracts (copy to projects)
├── examples/                       ← Concrete walkthroughs across all 3 scopes (Comparison, Decision, Project)
├── docs/meta-research/              ← 14 locked ADRs + evidence provenance
│
├── AGENTS.md                       ← AI agent operating manual
├── ROADMAP.md · CHANGELOG.md       ← Project history
├── CONTRIBUTING.md · SECURITY.md   ← Contribution & security policies
├── LICENSE · Makefile              ← MIT License · Developer shortcuts
└── go.mod · go.sum                 ← Go 1.25+ module
```

---

## Documentation & Guides

| Category | Guide | Description |
|---|---|---|
| **Getting Started** | [**Quickstart Guide**](docs/QUICKSTART.md) | Vivechak in 5 Minutes across comparison, decision, and project scopes |
| | [**Host Setup Guide**](docs/HOST-SETUP.md) | Universal MCP setup across 21 AI harnesses (Cursor, Claude Desktop, Antigravity, etc.) |
| | [**Manual Workflow**](docs/MANUAL-WORKFLOW.md) | Zero-install guide for running Vivechak in web chat without the MCP server |
| | [**Troubleshooting**](docs/TROUBLESHOOTING.md) | Diagnostic runbooks and comprehensive error code catalog |
| **Architecture & MCP** | [**Architecture Internals**](docs/ARCHITECTURE.md) | Server internals, package dependency diagrams, and concurrency model |
| | [**MCP Tools Reference**](docs/MCP-TOOLS.md) | Complete 13-tool API reference with example JSON-RPC payloads and schemas |
| | [**Harness Compatibility**](docs/HARNESS-COMPATIBILITY.md) | Compatibility matrix across 8 AI agent harness categories |
| **Advanced Workflows** | [**Parallel Execution**](docs/PARALLEL-EXECUTION.md) | Concurrent subagents and advisory file locking (`flock`) architecture |
| | [**Hybrid Execution**](docs/HYBRID-EXECUTION.md) | Native hybrid operations combining MCP automation with manual human browser research |
| | [**Brand Identity**](docs/BRAND.md) | Brand identity, optical sizing, geometry blueprints, and design tokens |
| **Scientific Provenance** | [**Meta-Research Base**](docs/meta-research/) | Empirical evidence base with all 14 locked ADRs (D-001–D-014) and UX decisions |

---

## Using with AI Agents

Vivechak ships as a universal **MCP server** — install the binary, register it with your AI host or agent harness (`vck setup <host>`), and your AI agent gets 13 tools:

| Tool | What It Does |
|---|---|
| `vivechak_init` | Create workspace structure and copy scope-appropriate operational contracts (1 for comparison, 2 for decision, 6 for project) |
| `vivechak_prepare_generator` | Return scope-appropriate generator prompt with context filled in |
| `vivechak_save_plan` | Validate + persist generated research plan and execution DAG (accepts `pipeline_content` alias at project scope) |
| `vivechak_status` | Report workspace progress (DAG completion, blocked sessions) |
| `vivechak_next_session` | Return next session prompt with upstream findings injected and active research calibration rules |
| `vivechak_save_session` | Validate + persist completed session output with real-time quality coaching (`Q-*`), mini-status, and FAD root copy mirroring |
| `vivechak_record_decision` | Save ADR or conflict resolution, handle ADR supersession, support single-turn drafting + auto-save (`auto_save: true`), and compile consolidated `DECISIONS.md` registry |
| `vivechak_amend_session` | Append post-hoc amendments with FAD root mirroring and transitive downstream stale alerts (`W-STALE-DOWNSTREAM`) |
| `vivechak_validate` | Dry-run validation ladder on any research artifact |
| `vivechak_run_gate` | Phase 0 exit gate (decoupled Track A + Track B with automated B1–B11 checks, auto-persisting `PHASE-0-GATE.md`) |
| `vivechak_challenge` | Adversarial stress-testing of completed sessions (red-team, evidence audit, cross-session consistency) |
| `vivechak_replan` | Mid-flight DAG mutation (add/remove sessions, update dependencies/prompts) |
| `vivechak_visualize` | Render pipeline DAG as Mermaid flowchart or ASCII status table |

The agent calls these tools in sequence. The server handles context injection, DAG resolution, evidence grading validation, and exit gate checks — the agent handles the actual research using its LLM capabilities.

**Still works without the MCP server:** The generators and templates in this repo are fully self-contained. You can copy-paste prompts manually — the MCP server just automates the orchestration.

---

## Autonomous Multi-Harness Verification & Stress Testing

Vivechak includes an autonomous multi-harness empirical testing engine in [`tests/harness/`](tests/harness/) that subjects the framework to zero-trust, continuous stress-testing across diverse AI host CLIs (Google Antigravity, OpenCode, Claude Code) and headless MCP environments:

* **21 Realistic Scenarios:** Evaluates messy, adversarial, and real-world developer inputs across all three scopes (comparison, decision, project) and enforces the out-of-scope code authoring boundary.
* **Deterministic Offline Verification (0 Tokens, ~1.5s):** Tests all 13 MCP tools over JSON-RPC stdio with zero network or AI credit expenditure:
  ```bash
  node scripts/run_harness.mjs --harness direct_mcp --batch all
  ```
* **Parallel Pre-Flight Health & Quota Probing:**
  ```bash
  make harness-probe                          # or: node scripts/run_harness.mjs --probe
  ```
* **Auto-Run Across Available & Funded Harnesses:**
  ```bash
  make harness-test                           # runs core test batch
  make harness-triage                         # clusters dynamic Turn 4 feedback
  ```
* **Empirical Ledger & Audit Reports:** All empirical findings and engine hardenings are permanently tracked in [`tests/harness/DEFECTS.md`](tests/harness/DEFECTS.md) and frozen in versioned milestone audit reports in [`tests/harness/reports/`](tests/harness/reports/).

---

## Key Capabilities

| Capability | How It Works |
|---|---|
| **Extract-2-Files** | Generator produces two clearly separated artifacts; guidance included for extracting from web chat |
| **Adaptive Scaling** | 8-dimension complexity scoring (0–24) maps to 4 tiers (1–30 sessions) |
| **Open-Ended Input** | Accepts natural language vision dumps; AI extracts parameters and classifies |
| **Constrained DAG** | Sessions run when dependencies are met, not rigid stage gates |
| **8-Block Prompts** | DECISION, BRIEF, SCOPE, KNOWN, CALIBRATION, APPROACH, DONE, FORMAT — no personas, no hardcoded queries |
| **Staged Triangulation** | Single-model default → multi-model only for contested one-way doors |
| **GRADE-Aligned Evidence** | A–E grades + modifiers (corroboration, recency, directness) + verification |
| **Two-Track Gate** | Fast-track for reversible decisions; 11-step + premortem for irreversible |

---

## Origin & Philosophy

### Why This Exists

Software venture failures almost never stem from bad code — they stem from **premature architectural decisions made on unverified assumptions.** The cost of wrong decisions compounds: a bad database choice costs 10× more to fix at month 6 than at month 0.

Vivechak was born from 8 real project research pipelines (2024–2026), each independently discovering pieces of the same methodology: unbiased landscape cataloging, evidence grading, aspect isolation, decision registries, and synthesis protocols.

### The Empirical Self-Validation

In August 2026, the framework was subjected to its own methodology. 11 independent deep research sessions (producing 15 artifacts across multiple AI platforms) tested every foundational assumption. Of 10 hypotheses, **0 survived unchanged**:

| Legacy Dogma | Verdict | Resolution |
|---|---|---|
| Absolute aspect-isolation | Refined | Context Architecture Law — conditional decomposition + synthesis |
| Mandatory 3-model triangulation | Refined | Staged, risk-triggered protocol |
| Fixed 17–27 sessions | Refuted | 4-tier adaptive scaling (1–30) |
| 8-section XML prompts | Refuted | 8-block prompt anatomy |
| Expert personas improve research | Refuted | Personas debunked for factual accuracy |
| Fixed ~40/60 compose/build | Refined | Wardley evolution mapping |

The most load-bearing finding: **how you frame a research question measurably biases what "evidence" a model reports back.** This single finding justifies the entire framework — unstructured "just ask the AI" research is systematically vulnerable to confirmation bias.

### The Self-Referential Validation

Vivechak's most distinctive property: it was validated by the methodology it prescribes. The meta-research discovered that several of its own axioms needed refinement. This recursive self-correction is built into Vivechak's DNA through mandatory review triggers and decay conditions on every locked ADR.

### FRAMEWORK.md — The Deep Reference

[FRAMEWORK.md](FRAMEWORK.md) contains the complete methodology specification — 8 principles, evidence grading system, pipeline topology, prompt anatomy, and synthesis workflows. **You do not need to read it to use Vivechak.** GENERATOR-PROJECT.md is self-contained; all methodology is operationalized inline in the generator prompt. FRAMEWORK.md is the definitive reference for understanding *why* the methodology works and for anyone improving the framework itself.

---

## Roadmap

| Phase | Status | Description |
|---|---|---|
| 1: Foundation (Gen 1) | ✅ | 8 project methodologies consolidated into initial patterns |
| 2: Self-Validation (Gen 2) | ✅ | 11 meta-research sessions → 10 verdicts → Gen 3 spec |
| 3: Framework Release (v0.1.0) | ✅ | Framework, generators, templates, bias correction |
| 4: Methodology Expansion | ✅ | 3 scope levels (Project / Decision / Comparison) |
| 5: MCP Server | ✅ | Go MCP server with 13 tools, full test suite passing |
| 6: Distribution | ✅ | GoReleaser, installers, Homebrew/Scoop/winget, CI pipelines |
| 7: Demand Proof | Next | Real case study, free wedge skill, content marketing |
| 8: Research Frontiers | Future | DSPy optimization, multi-agent debate, longitudinal calibration |

See [ROADMAP.md](ROADMAP.md) for detailed plans and evidence corpus.
