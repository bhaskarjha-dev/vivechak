# Quickstart — Vivechak in 5 Minutes

Transform technical decisions from gut-feel and cached LLM training data into structured, evidence-grounded research.

Vivechak runs at three scope levels depending on what you are deciding:

| Scope | Entry Generator | Output Artifact | Typical Investment |
|---|---|---|---|
| **Full Project** | [`GENERATOR.md`](../GENERATOR.md) | Founding Architecture Document (FAD) | 4–30 sessions (DAG-ordered) |
| **Single Decision** | [`GENERATOR-DECISION.md`](../GENERATOR-DECISION.md) | Architecture Decision Record (ADR) | 1–3 sessions |
| **Bounded Comparison** | [`GENERATOR-COMPARISON.md`](../GENERATOR-COMPARISON.md) | Weighted Evaluation (WEP) Matrix | 1 session |

Below are the two ways to use Vivechak: **Path A** via automated MCP server (recommended), or **Path B** manually in your browser.

---

## Running Example: Choosing a Database for a Multi-Tenant SaaS

Throughout this quickstart, we use a concrete scenario:
> *Building a multi-tenant B2B analytics platform ingesting 20M events/day. We need to decide whether to use PostgreSQL with TimescaleDB, ClickHouse, or DynamoDB + Athena, taking into account tenant isolation, write throughput, and query latency.*

---

## Path A: MCP Server (Recommended)

Give your AI coding agent (Cursor, VS Code, Claude Desktop, Antigravity, etc.) direct access to Vivechak's 10 orchestration tools.

### 1. Install Vivechak

**macOS / Linux:**
```sh
curl -fsSL https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.ps1 | iex
```

*Or via package managers:*
```sh
brew install bhaskarjha-dev/tap/vivechak                                            # Homebrew
winget install bhaskarjha-dev.Vivechak                                              # WinGet
scoop bucket add vivechak https://github.com/bhaskarjha-dev/scoop-bucket && scoop install vivechak # Scoop
go install github.com/bhaskarjha-dev/vivechak/cmd/vivechak@latest                   # Go 1.25+
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
*(For all 14 presets, terminal harnesses, or manual configs, see the [Host Setup Guide](HOST-SETUP.md).)*

### 3. Prompt Your AI Agent

Open your agent chat in your project repository and run:

> *"Use Vivechak to research the architecture for our B2B SaaS analytics app. We need to choose our primary analytics datastore for 20M events/day across 500 tenants."*

### 4. The Agent Executes the Loop

Your agent coordinates the 10 tools autonomously:

```
[Agent: vivechak_init]
  └─ Creates research/ workspace, directories, and 6 operational contracts
[Agent: vivechak_prepare_generator]
  └─ Fills your vision into GENERATOR.md
[Agent: executes generator prompt]
  └─ Generates RESEARCH-PIPELINE.md (DAG) & DECISIONS.md (registry)
[Agent: vivechak_save_plan]
  └─ Validates DAG dependencies & persists research plan
[Agent: Execution Loop]
  ├─ vivechak_next_session   ──> Returns next ready prompt with upstream findings injected
  ├─ [Agent executes research using web search & deep retrieval]
  ├─ vivechak_save_session   ──> Validates evidence grades (A-E) and saves session output
  └─ vivechak_record_decision ──> Locks ADR with One-Way vs Two-Way Door classification
[Agent: vivechak_run_gate]
  └─ Evaluates Phase 0 exit gate (Track A fast-track or Track B 9-step check)
```

See the [MCP Tools Reference](MCP-TOOLS.md) for full tool schemas, parameter specs, and error handling.

---

## Path B: Manual Workflow (No Installation Needed)

No binaries or CLI required. Works in ChatGPT, Gemini, or Claude web browsers with web search enabled.

### 1. Copy the Generator Prompt
Open [`GENERATOR.md`](../GENERATOR.md) and copy the generator prompt inside the markdown code block.

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

### 4. Copy the 6 Contract Templates
Copy the templates from `templates/` into your project's `research/templates/`:
- `DECISIONS.template.md`
- `CONFLICT-RESOLUTION.template.md`
- `COMPARISON-SESSION.template.md`
- `FOUNDING-ARCHITECTURE.template.md`
- `PHASE-0-GATE.template.md`
- `SESSION.template.md`

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
