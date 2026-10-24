<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.svg">
  <img src="docs/assets/logo.svg" alt="Vivechak Logo" width="180" />
</picture>

# Vivechak (विवेचक)
### Evidence-Grounded Research for Architectural & Technical Decisions

[![Release](https://img.shields.io/badge/release-v0.1.0-blue.svg?style=flat-square)](https://github.com/bhaskarjha-dev/vivechak/releases)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8.svg?style=flat-square&logo=go&logoColor=white)](https://golang.org)
[![MCP Protocol](https://img.shields.io/badge/MCP-2024--11--05-8A2BE2.svg?style=flat-square)](https://modelcontextprotocol.io)
[![Test Suite](https://github.com/bhaskarjha-dev/vivechak/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/bhaskarjha-dev/vivechak/actions)
[![Zero Telemetry](https://img.shields.io/badge/telemetry-zero%20(100%25%20local)-success.svg?style=flat-square)](#when-not-to-use-vivechak)
[![License: MIT](https://img.shields.io/badge/license-MIT-yellow.svg?style=flat-square)](LICENSE)

*vi- (apart) + √vic (to sift/separate) + -aka (the agent who does) = "the discerning analyst"*

</div>

---

> **Evidence-Graded Technical Decisions for AI Coding Agents**  
> Ground high-stakes architectural choices in reproducible empirical evidence before writing code. Single standalone Go binary (`vck`) with zero runtime dependencies and zero telemetry.

---

## Quick Install

```bash
# macOS & Linux (Homebrew)
brew install bhaskarjha-dev/tap/vivechak

# Windows (PowerShell)
irm https://vivechak.dev/install.ps1 | iex

# Any OS with Node or Python (zero install)
npx -y vivechak serve    # or: uvx vivechak serve
```

## Editor Setup

Auto-configure your AI coding assistant in one command:
```bash
vck setup cursor         # Cursor IDE (.cursor/mcp.json)
vck setup vscode         # VS Code Copilot Agent (.vscode/mcp.json)
vck setup claude-code    # Claude Code (~/.claude.json or .mcp.json)
vck setup agy            # Google Antigravity
vck setup                # Auto-detects editor in current workspace
```
*(Note: MCP slash commands such as `/compare` are supported in hosts like Claude Code).*

## The Three Decision Scopes

Vivechak scales research depth to decision reversibility:

| Scope | Deliverable | Typical Prompt | Overhead |
|---|---|---|---|
| **Comparison** | `research/comparisons/C-NNN.md` | *"Compare Pino vs Winston"* | 1 session (~5 min), 0 DAG, 0 gate |
| **Decision** | `research/decisions/D-NNN.md` | *"PostgreSQL vs ClickHouse for events"* | 1–3 sessions (~30 min), 0 project DAG |
| **Project** | `research/FAD.md` | *"Architect real-time streaming pipeline"* | Full DAG (4–30 sessions), Phase 0 Gate |

## 9 Consolidated MCP Tools

| Tool | Action | Primary Scope |
|---|---|---|
| `vivechak_start` | Initialize workspace and return targeted research brief | All |
| `vivechak_save` | Atomically persist comparison, plan, session, or FAD | All |
| `vivechak_record_decision` | Lock canonical ADR (`D-NNN.md`) and compile `DECISIONS.md` | Decision / Project |
| `vivechak_next` | Resolve next unblocked DAG session via topological sort | Project |
| `vivechak_status` | Return workspace health, progress, and blockers | All |
| `vivechak_review` | Surface real-time quality coaching observations (`Q-*`) | Decision / Project |
| `vivechak_amend` | Revise completed session or update findings | All |
| `vivechak_replan` | Dynamically prune or insert sessions based on discoveries | Project |
| `vivechak_gate` | Evaluate 12 Phase 0 exit rules (Track A & Track B) | All |

## Evidence Taxonomy

Every factual claim in Vivechak requires an inline evidentiary grade:
- **Grade A (`E-A`):** Canonical documentation, empirical measurement, RFC, vendor-neutral spec with URL.
- **Grade B (`E-B`):** Third-party production postmortem, peer-reviewed paper, public issue tracker.
- **Grade C (`E-C`):** Community benchmark, vendor comparison, tutorial, blog post.
- **Grade D (`E-D`):** Unverified forum claim, ecosystem reputation, personal anecdote.
- **Grade E (`E-E`):** Model training recall (recalled without live verification).

*(See [Methodology Specification](FRAMEWORK.md) and [Sample Decision Walkthrough](examples/SAMPLE-DECISION.md) for illustrative examples with graded evidence).*

## When NOT to Use Vivechak

Vivechak is an architectural governance framework, not a code authoring tool:
- ❌ Do NOT use for writing functions, UI components, tests, or bug fixes.
- ❌ Do NOT use for syntax lookups or language trivia.
- ✅ DO use for datastores, core frameworks, security boundaries, and One-Way Doors.

## Documentation Index

- [Quickstart Guide](docs/QUICKSTART.md) · [Host Setup & Presets](docs/hosts.md) · [Methodology Specification](FRAMEWORK.md)
- [MCP Tools Reference](docs/reference/tools.md) · [Validation & Error Codes](docs/reference/codes.md) · [Phase 0 Gate Rules](docs/reference/gate.md)
- [CLI Architecture](docs/ARCHITECTURE.md) · [Contributing Guide](CONTRIBUTING.md)

## License

MIT License · Copyright (c) 2026 Bhaskar Jha.
