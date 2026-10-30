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
irm https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.ps1 | iex

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

## Upgrade & Uninstall

### Upgrading
- **Homebrew (macOS & Linux):** `brew update && brew upgrade bhaskarjha-dev/tap/vivechak`
- **Windows (PowerShell):** Re-run `irm https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.ps1 | iex`
- **Linux / macOS Shell:** Re-run `curl -fsSL https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.sh | sh`
- **WinGet / Scoop:** `winget upgrade bhaskarjha-dev.Vivechak` or `scoop update vck`
- **Go toolchain:** `go install github.com/bhaskarjha-dev/vivechak/cmd/vivechak@latest`
- **Zero-install runners (`npx` / `uvx`):** Always run the latest package automatically.

### Uninstalling
1. **Remove binary:**
   - Homebrew: `brew uninstall bhaskarjha-dev/tap/vivechak`
   - WinGet / Scoop: `winget uninstall bhaskarjha-dev.Vivechak` or `scoop uninstall vck`
   - Shell / Manual: Delete `vck` (and `vivechak`) from your bin directory (e.g. `~/.local/bin/vck` or `%LOCALAPPDATA%\Programs\vivechak\vck.exe`).
2. **Remove MCP host configuration:**
   - Delete the `"vivechak"` block from your IDE/host config file (e.g. `.cursor/mcp.json`, `.vscode/mcp.json`, `~/.claude.json`), or restore the automatic `<config>.bak` created during `vck setup`.

## The Three Decision Scopes

Vivechak scales research depth to decision reversibility:

| Scope | Deliverable | Typical Prompt | Overhead |
|---|---|---|---|
| **Comparison** | `research/C-NNN-comparison.md` (WEP) | *"Compare Pino vs Winston"* | 1 session (~5 min), 0 DAG, 0 gate |
| **Decision** | `research/D-NNN.md` (ADR) | *"PostgreSQL vs ClickHouse for events"* | 1–3 sessions (~30 min), 0 project DAG |
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
- **Grade D (`E-D`):** Unverified forum claim, ecosystem reputation, blog post, or model memory recall (capped at Grade D per Principle P3).
- **Grade E (`E-E`):** Speculative hypothesis, uncorroborated assertion, or unsubstantiated claim.

*(See [Methodology Specification](FRAMEWORK.md) and [Sample Decision Walkthrough](examples/SAMPLE-DECISION.md) for illustrative examples with graded evidence).*

## Positioning: What Vivechak Is vs. What Vivechak Is Not

| What Vivechak IS | What Vivechak IS NOT |
|---|---|
| **Epistemic Grounding Engine:** Pre-coding research meta-framework providing empirical evidence grades (`E-A` through `E-E`) | **A Code Generator:** It does not write production code, components, endpoints, or scripts |
| **Architectural Gatekeeper:** Enforces pre-coding exit criteria (`Track A` Two-Way vs. `Track B` One-Way Doors) with interactive human sign-off (`vck signoff`) | **A Linter or Static Analyzer:** It does not analyze ASTs, enforce formatting, or typecheck code |
| **Reversibility-Calibrated Sieve:** 1-session comparisons (`C-NNN`), 1–3 session ADRs (`D-NNN`), or multi-session DAG pipelines | **An Autonomous Orchestrator for Everything:** Not an open-ended loop for routine programming tasks |
| **Structured Falsification:** Mandates disconfirming evidence searches, substantive discovered concerns, and premortems | **A Prompt Persona:** No expert persona roleplays, XML prompt bloat, or arbitrary search count quotas |

### When NOT to Use Vivechak

Vivechak is an architectural governance framework, not a code authoring tool:
- ❌ Do NOT use for writing functions, UI components, tests, or bug fixes.
- ❌ Do NOT use for syntax lookups or language trivia.
- ✅ DO use for datastores, core frameworks, security boundaries, and One-Way Doors.

## Documentation Index

- **Guides:** [Quickstart Guide](docs/QUICKSTART.md) · [Host Setup & Presets](docs/hosts.md) · [Manual Web Workflow](docs/MANUAL-WORKFLOW.md) · [Troubleshooting & Diagnostics](docs/TROUBLESHOOTING.md)
- **Deep Dives:** [MCP Tools Manual](docs/MCP-TOOLS.md) · [Architecture & Internals](docs/ARCHITECTURE.md) · [Parallel Subagent Execution](docs/PARALLEL-EXECUTION.md) · [Hybrid MCP + Manual Workflows](docs/HYBRID-EXECUTION.md)
- **Methodology & Contracts:** [Methodology Specification](FRAMEWORK.md) · [Validation & Error Codes](docs/reference/codes.md) · [Phase 0 Gate Rules](docs/reference/gate.md) · [Brand & Identity](docs/BRAND.md) · [Contributing Guide](CONTRIBUTING.md)

## License

MIT License · Copyright (c) 2026 Bhaskar Jha.
