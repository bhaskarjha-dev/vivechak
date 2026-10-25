# Quickstart — Vivechak in 5 Minutes

Transform technical decisions from gut-feel and cached LLM training data into structured, evidence-grounded research.

Vivechak runs at three scope levels depending on what you are deciding:

| Scope | Entry Generator | Output Artifact | Typical Investment |
|---|---|---|---|
| **Full Project** | [`GENERATOR-PROJECT.md`](../GENERATOR-PROJECT.md) | Founding Architecture Document (FAD) | 4–30 sessions (DAG-ordered) |
| **Single Decision** | [`GENERATOR-DECISION.md`](../GENERATOR-DECISION.md) | Architecture Decision Record (ADR) | 1–3 sessions |
| **Bounded Comparison** | [`GENERATOR-COMPARISON.md`](../GENERATOR-COMPARISON.md) | Weighted Evaluation (WEP) Matrix | 1 session |

Below are the two ways to use Vivechak: **Path A** via automated MCP server (recommended), or **Path B** manually in your browser.

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
npx -y vck serve                              # or: npx -y vivechak serve

# Python / uvx (macOS, Linux, Windows)
uvx vck serve                                 # or: uvx vivechak serve

# Universal Agent Skill (Claude Code, Cursor, Antigravity, OpenCode)
npx skills add bhaskarjha-dev/vivechak

# Docker / Cloud Agents (Devin, OpenHands, CI)
docker run -i --rm -v "${PWD}:/workspace" ghcr.io/bhaskarjha-dev/vck:latest serve
```

**Or Native Install via Package Managers:**
```sh
# Homebrew (macOS & Linuxbrew)
brew install bhaskarjha-dev/tap/vck           # or: brew install bhaskarjha-dev/tap/vivechak

# Windows (WinGet or Scoop)
winget install bhaskarjha-dev.Vivechak
scoop bucket add vivechak https://github.com/bhaskarjha-dev/scoop-bucket && scoop install vck

# Debian / Ubuntu (.deb) & Fedora / RHEL (.rpm)
sudo dpkg -i vivechak_0.1.0_linux_amd64.deb   # Debian / Ubuntu
sudo rpm -i vivechak_0.1.0_linux_amd64.rpm    # Fedora / RHEL

# Nix / NixOS
nix run github:bhaskarjha-dev/vivechak

# Direct Shell Installers
curl -fsSL https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.sh | sh      # macOS/Linux
irm https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.ps1 | iex           # Windows
```

**Or Build from Source (Go 1.25+):**
```sh
# Clone and build locally with zero external C dependencies:
git clone https://github.com/bhaskarjha-dev/vivechak.git
cd vivechak
make build                                    # builds production-optimized vck and vivechak binaries (-s -w stripped)

# Or manually:
# go build -trimpath -ldflags="-s -w -X main.version=0.1.0" -o vck ./cmd/vivechak/

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
*(For all 21 presets, terminal harnesses, or manual configs, see the [Host Setup Guide](HOST-SETUP.md).)*

### 3. Prompt Your AI Agent Based on Scope (or Use Slash Commands)

Open your agent chat in your project repository and choose your scope (or invoke native **MCP Slash Commands** in compatible clients):

- **For Quick Comparisons (~5 mins, 1 session):**
  > `/compare` *(or `/vivechak_compare`)* or *"Vivechak compare: Pino vs Winston for high-throughput Node.js microservices."*
- **For Architectural Decisions (~30 mins, 1–3 sessions):**
  > `/decision` *(or `/vivechak_decision`)* or *"Vivechak decision: Should we migrate our primary analytics datastore from PostgreSQL to ClickHouse?"*
- **For Full Project Pipelines (Hours–Days, greenfield architecture):**
  > `/project` *(or `/vivechak_project`)* or *"Use Vivechak to research our architecture. We need to build a multi-tenant B2B SaaS analytics platform ingesting 20M events/day."*
- **For Pre-Code Phase 0 Gate Evaluation:**
  > `/gate` *(or `/vivechak_gate`)* or *"Evaluate Phase 0 Gate before implementation."*

> [!NOTE]
> **When NOT to use Vivechak:** Do NOT invoke Vivechak for routine coding, writing functions, debugging errors, or syntax explanations. Let your agent write code directly without research tools.

### 4. How the Agent Executes Each Scope

- **Comparison Scope:** Agent calls `vivechak_start(scope="comparison", question="...")` → executes prompt with search → calls `vivechak_save(kind="comparison")` to persist WEP matrix in `research/C-NNN-comparison.md`. Done in 5 minutes! Zero DAG, zero gate.
- **Decision Scope:** Agent calls `vivechak_start(scope="decision", question="...")` → saves plan via `vivechak_save(kind="plan")` → runs 1–3 sessions → calls `vivechak_record_decision` to lock canonical ADR `research/D-NNN.md` and compiled registry `research/DECISIONS.md`. Zero project DAG, zero FAD.
- **Project Scope:** Autonomous orchestration across 9 consolidated tools:
```
[Agent: vivechak_start (scope: 'project', question: '<vision>')]
  └─ Initializes research/ workspace and returns actionable DAG generator brief
[Agent: executes generator prompt]
  └─ Generates RESEARCH-PIPELINE.md (DAG) & DECISIONS.md (registry)
[Agent: vivechak_save (kind: 'plan', id: 'DAG')]
  └─ Validates DAG dependencies & persists research plan
[Agent: Execution Loop]
  ├─ vivechak_next           ──> Injects upstream findings into next ready prompt
  ├─ [Agent executes deep research with web search]
  ├─ vivechak_save           ──> Validates evidence grades (A-E) and saves session (kind='session')
  └─ vivechak_record_decision ──> Locks ADR with One-Way vs Two-Way Door classification
[Agent: FAD Synthesis (SYN-01)]
  └─ Synthesizes all findings into research/FAD.md via vivechak_save (kind: 'fad')
[Agent: vivechak_gate (or CLI: vck gate)]
  └─ Evaluates Phase 0 Exit Gate across Track A and Track B (Rules G1–G12)
```

See the [MCP Tools Reference](MCP-TOOLS.md) for full tool schemas, parameter specs, and error handling.

### Upgrade & Uninstall

**To upgrade Vivechak:**
- **Homebrew (macOS & Linux):** `brew update && brew upgrade bhaskarjha-dev/tap/vivechak`
- **Windows (PowerShell):** Re-run installer: `irm https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.ps1 | iex`
- **Linux / macOS Shell:** Re-run installer: `curl -fsSL https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.sh | sh`
- **WinGet / Scoop:** `winget upgrade bhaskarjha-dev.Vivechak` or `scoop update vck`
- **Go toolchain:** `go install github.com/bhaskarjha-dev/vivechak/cmd/vivechak@latest`
- **Zero-install runners (`npx` / `uvx`):** Always run the latest package automatically.

**To uninstall:**
1. **Remove binary:**
   - Homebrew: `brew uninstall bhaskarjha-dev/tap/vivechak`
   - WinGet / Scoop: `winget uninstall bhaskarjha-dev.Vivechak` or `scoop uninstall vck`
   - Shell / Manual: Delete `vck` (and `vivechak`) from your binary directory (e.g. `~/.local/bin/vck` or `%LOCALAPPDATA%\Programs\vivechak\vck.exe`).
2. **Remove MCP configuration:**
   - Delete the `"vivechak"` block from your host config file (e.g. `.cursor/mcp.json`, `.vscode/mcp.json`, `~/.claude.json`), or restore the automatic `<config>.bak` backup created during `vck setup`.

---

## Path B: Manual Workflow (No Installation Needed)

No binaries or CLI required. Works in ChatGPT, Gemini, or Claude web browsers with web search enabled.

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

- **Evidence Grades (A–E):** Grade A (Official docs, RFCs, source code), Grade B (Empirical benchmarks, reproducible tests), Grade C (Vendor claims), Grade D (Blogs, unverified tutorials; AI recall capped here), Grade E (Unsubstantiated claims).
- **Two-Way vs. One-Way Doors:** Two-Way doors (reversal cost low) need ~70% confidence and run fast. One-Way doors (persistence schemas, data migrations, public API contracts) require corroborated Grade A/B evidence, falsification checks, and locked reversal triggers.
- **Structured Falsification:** Prompts explicitly direct the AI to hunt for failure modes and disconfirming evidence against the leading candidate, preventing confirmation bias.

---

## Next Steps

- Setting up your editor? Check [Host Setup Guide](HOST-SETUP.md).
- Building with AI agents? See [MCP Tools Reference](MCP-TOOLS.md).
- Running manually in web chats? See [Manual Workflow Guide](MANUAL-WORKFLOW.md).
- Deep dive into methodology? Read the complete [Framework Specification](../FRAMEWORK.md).
