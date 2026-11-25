# Quickstart — Vivechak in 5 Minutes

Transform technical decisions from gut-feel and cached LLM training data into structured, evidence-grounded research.

Vivechak runs at three scope levels depending on what you are deciding:

| Scope | Entry Generator | Output Artifact | Typical Investment |
|---|---|---|---|
| **Full Project** | [`GENERATOR-PROJECT.md`](../GENERATOR-PROJECT.md) | Founding Architecture Document (FAD) | 4–30 sessions (DAG-ordered) |
| **Single Decision** | [`GENERATOR-DECISION.md`](../GENERATOR-DECISION.md) | Architecture Decision Record (ADR) | 1–3 sessions |
| **Bounded Comparison** | [`GENERATOR-COMPARISON.md`](../GENERATOR-COMPARISON.md) | Weighted Evaluation (WEP) Matrix | 1 session |

Below are the three ways to use Vivechak: **Path A** via the automated MCP server & CLI (recommended), **Path B** via the standalone Agent Skill (`npx skills add`, zero binary required), or **Path C** manually in browser AI chats.

---

## Running Example: Choosing a Database for a Multi-Tenant SaaS

Throughout this quickstart, we use a concrete scenario:
> *Building a multi-tenant B2B analytics platform ingesting 20M events/day. We need to decide whether to use PostgreSQL with TimescaleDB, ClickHouse, or DynamoDB + Athena, taking into account tenant isolation, write throughput, and query latency.*

---

## Path A: MCP Server (Recommended)

Give your AI coding agent (Cursor, VS Code, Claude Desktop, Antigravity, etc.) direct access to Vivechak's 9 consolidated orchestration tools.

### 1. Install & Run

**Instant / Ephemeral Runners (Zero Pre-Installation Required):**
```sh
# Node.js / npx (macOS, Linux, Windows)
npx -y vivechak serve

# Python / uvx (macOS, Linux, Windows)
uvx vck serve                                 # or: uvx vivechak serve

# Universal Agent Skill (Claude Code, Cursor, Antigravity, OpenCode)
npx skills add bhaskarjha-dev/vivechak

# Docker / Cloud Agents (Devin, OpenHands, CI)
docker run -i --rm --user "$(id -u):$(id -g)" -e VIVECHAK_PROJECT_ROOT=/workspace -v "${PWD}:/workspace" ghcr.io/bhaskarjha-dev/vck:latest serve
```

**Or Native Install via Package Managers:**
```sh
# Homebrew (macOS & Linuxbrew)
brew install --cask bhaskarjha-dev/tap/vivechak

# Windows (WinGet or Scoop)
winget install bhaskarjha-dev.Vivechak
scoop bucket add vivechak https://github.com/bhaskarjha-dev/scoop-bucket && scoop install vivechak

# Debian / Ubuntu (.deb) & Fedora / RHEL (.rpm)
sudo dpkg -i vivechak_0.1.0_linux_amd64.deb   # Debian / Ubuntu
sudo rpm -i vivechak_0.1.0_linux_amd64.rpm    # Fedora / RHEL

# Nix / NixOS
nix run github:bhaskarjha-dev/vivechak

# Direct Shell Installers
curl -fsSL https://vivechak.bhaskarjha.dev/install.sh | bash                                  # macOS/Linux
irm https://vivechak.bhaskarjha.dev/install.ps1 | iex                                         # Windows
```

**Or Build from Source (Go 1.26+, toolchain Go 1.26.11):**
```sh
# Clone and build locally with zero external C dependencies:
git clone https://github.com/bhaskarjha-dev/vivechak.git
cd vivechak
make build                                    # builds production-optimized vck and vivechak binaries (-s -w stripped)

# Or manually:
# go build -trimpath -ldflags="-s -w -X github.com/bhaskarjha-dev/vivechak/internal/version.Version=0.1.0" -o vck ./cmd/vivechak/

# Optional: Install to PATH or run directly
./vck version                                 # (or .\vck.exe on Windows)
```
Or install via Go toolchain:
```sh
go install github.com/bhaskarjha-dev/vivechak/cmd/vivechak@latest
```

### 2. Configure Your AI Host

Configure your AI host or editor in 1 second using the official shorthand **`vck`** (or `vivechak`):

```sh
# Fast setup for your AI host / IDE:
vck setup cursor                             # Cursor IDE
vck setup vscode                             # VS Code (Copilot Agent mode)
vck setup claude                             # Claude Desktop
vck setup agy                                # Google Antigravity
vck setup                                    # auto-detect host in current workspace

# Or output universal MCP JSON configuration:
vck mcp-config
```
*(For every preset, terminal harnesses, or manual configs, see the [Host Setup Guide](HOST-SETUP.md).)*

### 3. Prompt Your AI Agent Based on Scope (or Use Slash Commands)

Open your agent chat in your project repository and choose your scope (or invoke native **MCP Slash Commands** in compatible clients):

- **For Quick Comparisons (~5 mins, 1 session):**
  > `/compare` or *"Vivechak compare: Pino vs Winston for high-throughput Node.js microservices."*
- **For Architectural Decisions (~30 mins, 1–3 sessions):**
  > `/decision` or *"Vivechak decision: Should we migrate our primary analytics datastore from PostgreSQL to ClickHouse?"*
- **For Full Project Pipelines (Hours–Days, greenfield architecture):**
  > `/project` or *"Use Vivechak to research our architecture. We need to build a multi-tenant B2B SaaS analytics platform ingesting 20M events/day."*
- **For Pre-Code Phase 0 Gate Evaluation:**
  > `/gate` or *"Evaluate Phase 0 Gate before implementation."*

> [!NOTE]
> **When NOT to use Vivechak:** Do NOT invoke Vivechak for routine coding, writing functions, debugging errors, or syntax explanations. Let your agent write code directly without research tools.

### 4. How the Agent Executes Each Scope

- **Comparison Scope:** Agent calls `vivechak_start(scope="comparison", question="...")` → executes prompt with web search → calls `vivechak_save(kind="comparison", content="<matrix>")` to persist a WEP matrix with ±20% weight sensitivity in `research/C-NNN-comparison.md`. Done in 5 minutes! Zero DAG, zero gate.
- **Decision Scope:** Agent calls `vivechak_start(scope="decision", question="...")` → saves plan via `vivechak_save(kind="plan", id="D-NNN", content="<plan>")` → runs 1–3 sessions → calls `vivechak_record_decision(id="D-NNN", content="<adr>")` to record canonical ADR `research/D-NNN-decision.md` and compiled `research/DECISIONS.md` → challenges One-Way Doors with `vivechak_review(mode="red_team", id="D-NNN")` → human runs `vck signoff D-001 --reviewer "<name>"` in an interactive terminal → `vivechak_gate(decision="D-NNN")`. Zero project DAG, zero FAD.
- **Project Scope:** Autonomous orchestration across 9 consolidated tools:
```
[Agent: vivechak_start(scope="project", question="<vision>")]
  └─ Initializes research/ workspace and returns actionable DAG generator brief
[Agent: executes generator prompt]
  └─ Generates RESEARCH-PIPELINE.md (DAG) & DECISIONS.md (registry)
[Agent: vivechak_save(kind="plan", id="DAG", content="<pipeline>")]
  └─ Validates DAG dependencies & persists research plan
[Agent: Execution Loop]
  ├─ vivechak_next()          ──> Injects upstream findings into next ready prompt
  ├─ [Agent executes deep research with web search]
  ├─ vivechak_save(kind="session", id="T1-01", content="<session>") ──> Validates evidence grades (A-E)
  └─ vivechak_record_decision(id="D-001", content="<adr>")          ──> Records ADR
[Agent: FAD Synthesis (SYN-01)]
  └─ Synthesizes all findings into research/FAD.md via vivechak_save(kind="fad", id="SYN-01", content="<fad>")
[Agent: vivechak_gate() (or CLI: vck gate)]
  └─ Evaluates Phase 0 Exit Gate across Track A and Track B (Rules G1–G12)
```

See the [MCP Tools Reference](MCP-TOOLS.md) for full tool schemas, parameter specs, and error handling.

### Upgrade & Uninstall

**To upgrade Vivechak:**
- **Homebrew (macOS & Linux):** `brew update && brew upgrade --cask bhaskarjha-dev/tap/vivechak`
- **Windows (PowerShell):** Re-run installer: `irm https://vivechak.bhaskarjha.dev/install.ps1 | iex`
- **Linux / macOS Shell:** Re-run installer: `curl -fsSL https://vivechak.bhaskarjha.dev/install.sh | bash`
- **WinGet / Scoop:** `winget upgrade bhaskarjha-dev.Vivechak` or `scoop update vivechak`
- **Go toolchain:** `go install github.com/bhaskarjha-dev/vivechak/cmd/vivechak@latest`
- **Zero-install runners (`npx` / `uvx`):** Always run the latest package automatically.

**To uninstall:**
1. **Remove binary:**
   - Homebrew: `brew uninstall --cask bhaskarjha-dev/tap/vivechak`
   - WinGet / Scoop: `winget uninstall bhaskarjha-dev.Vivechak` or `scoop uninstall vivechak`
   - Shell / Manual: Delete `vck` (and `vivechak`) from your binary directory (e.g. `~/.local/bin/vck` or `%LOCALAPPDATA%\Programs\vivechak\vck.exe`).
2. **Remove MCP configuration:**
   - Delete the `"vivechak"` block from your host config file (e.g. `.cursor/mcp.json`, `.vscode/mcp.json`, `~/.claude.json`), or restore the automatic `<config>.vivechak-bak-<timestamp>` backup created during `vck setup`.

---

## Path B: Standalone Agent Skill (Zero MCP Server Required)

If you prefer not to run an MCP server or install a binary yet, install the dual-mode Vivechak Agent Skill directly into Claude Code, Cursor, Antigravity, or OpenCode:

```sh
npx skills add bhaskarjha-dev/vivechak
```

Your coding agent learns the exact 3-scope workflow, A–E evidence grading, ±20% WEP sensitivity check, and One-Way Door red-team discipline ([`skills/vivechak/SKILL.md`](../skills/vivechak/SKILL.md)) and writes standard Markdown artifacts directly into `research/` using its built-in file and web search tools. If you install `vck` later, every file is already compatible with `vck doctor`, `vck signoff`, and `vck gate`.

---

## Path C: Manual Workflow (Browser Chat, Zero Installation)

No binaries, CLI, or skills required. Works in ChatGPT, Gemini, Claude, or Perplexity web browsers with web search enabled — or build prompts and lint ADRs immediately in the browser at [vivechak.bhaskarjha.dev/playground](https://vivechak.bhaskarjha.dev/playground).

### 1. Copy the Generator Prompt
Open [`GENERATOR-PROJECT.md`](../GENERATOR-PROJECT.md) and copy the generator prompt inside the markdown code block.

### 2. Paste & Run with Your Project Vision
Paste the prompt into Claude, ChatGPT Plus, or Gemini Advanced (with search enabled). Fill in `[PASTE YOUR PROJECT DESCRIPTION HERE]`:

```text
Project: Multi-tenant SaaS analytics platform.
Workload: 20M event writes/day, 95th percentile query latency < 500ms for dashboards.
Tenants: 500 B2B customers, strict tenant data isolation required.
Candidates: PostgreSQL + TimescaleDB vs ClickHouse vs DynamoDB.
Constraints: Team knows Postgres well; small DevOps team (managed cloud preferred).
```

### 3. Save the Two Generated Files
The AI returns two separate documents. Create a `research/` directory in your project root and save:
1. `research/RESEARCH-PIPELINE.md` (the session DAG and copy-paste prompts)
2. `research/DECISIONS.md` (the decision registry)

### 4. Copy the Scope-Appropriate Templates
For a full project pipeline, copy all 6 templates from `templates/` into your project's `research/templates/`:
- `DECISIONS.template.md`
- `CONFLICT-RESOLUTION.template.md`
- `COMPARISON-SESSION.template.md`
- `FOUNDING-ARCHITECTURE.template.md`
- `PHASE-0-GATE.template.md`
- `SESSION.template.md`

*(Note: For a single decision scope, you only need the first 2 templates. For a bounded comparison, you only need `COMPARISON-SESSION.template.md`.)*

Your project workspace is now completely self-contained:
```
my-saas-project/
└── research/
    ├── RESEARCH-PIPELINE.md
    ├── DECISIONS.md
    ├── sessions/
    └── templates/
```

### 5. Execute, Record, Synthesize & Gate
1. **Execute Sessions:** Copy each prompt from `RESEARCH-PIPELINE.md` into an AI deep research session. Save each result to `research/sessions/<filename>.md`.
2. **Record Decisions:** Fill in `research/DECISIONS.md` as sessions complete. Classify decisions as **Two-Way Doors** (reversible) or **One-Way Doors** (irreversible, requiring explicit reversal triggers).
3. **Synthesize & Gate:** Compile findings into `FAD.md` using `FOUNDING-ARCHITECTURE.template.md`, then run the checklist in `PHASE-0-GATE.template.md` before writing code.

For complete step-by-step instructions, see the [Manual Workflow Guide](MANUAL-WORKFLOW.md).

---

## Smaller Scope: When You Don't Need a Full Pipeline

Not every technical choice requires a full 4–30 session project pipeline.

### Single Decision (1–3 Sessions → ADR)
Use [`GENERATOR-DECISION.md`](../GENERATOR-DECISION.md) when you have one specific dilemma.
- **Example:** *"Should we use PostgreSQL or ClickHouse for our SaaS analytics datastore?"*
- **What it does:** Profiles reversal cost (R), novelty (N), blast radius (B), and regulatory risk (X). Generates 1–3 targeted session prompts (`L` landscape, `C` comparison, `F` falsification) and a pre-populated ADR skeleton (`D-001-datastore.md`).

### Quick Comparison (1 Session → WEP Matrix)
Use [`GENERATOR-COMPARISON.md`](../GENERATOR-COMPARISON.md) when comparing 2–5 defined options on known criteria.
- **Example:** *"Compare Redis vs Dragonfly vs KeyDB for cache clustering."*
- **What it does:** Emits a single research prompt producing a Weighted Evaluation Matrix (WEP), sensitivity testing (±20% weight variations), concrete failure modes, and disconfirming evidence.

---

## Key Methodology in 30 Seconds

- **Evidence Grades (A–E):** every factual claim carries one:
  <!-- docgen:begin evidence-grades (generated from internal/method/evidence.json by tools/docgen; do not edit by hand) -->
  - **Grade A (Primary Empirical):** Official specifications, canonical source code, RFCs, or primary documentation with explicit URL anchor.
  - **Grade B (Secondary Empirical):** Peer-reviewed papers, reproducible third-party benchmarks, published postmortems, or corroborated engineering reports.
  - **Grade C (Vendor & Practitioner):** Vendor marketing, product whitepapers, vendor benchmarks, conference talks, or interested-party documentation.
  - **Grade D (Informal & Opinion):** Blog posts, forum discussions, StackOverflow answers, or model memory recall (capped at Grade D).
  - **Grade E (Speculative & Unsubstantiated):** Unverified hypotheses, uncorroborated parametric recall, or ungrounded assertions.
  <!-- docgen:end evidence-grades -->
- **Two-Way vs. One-Way Doors:** Two-Way doors (reversal cost low) need ~70% confidence and run fast. One-Way doors (persistence schemas, data migrations, public API contracts) require corroborated Grade A/B evidence, falsification checks, and locked reversal triggers.
- **Structured Falsification:** Prompts explicitly direct the AI to hunt for failure modes and disconfirming evidence against the leading candidate, preventing confirmation bias.

---

## Next Steps

- Setting up your editor? Check [Host Setup Guide](HOST-SETUP.md).
- Building with AI agents? See [MCP Tools Reference](MCP-TOOLS.md).
- Running manually in web chats? See [Manual Workflow Guide](MANUAL-WORKFLOW.md).
- Deep dive into methodology? Read the complete [Framework Specification](../FRAMEWORK.md).
