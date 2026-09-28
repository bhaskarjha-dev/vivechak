# Vivechak (विवेचक) — Agent Operating Manual
### Vivechak v1.1 · Single Source of Truth for AI Agent Operations

> **Read this file COMPLETELY before doing anything in this repository.**

---

## 1. What This Repository Is

**Vivechak** is a meta-framework that generates evidence-grounded research for technical decisions. It works at three scope levels: full project pipelines (→ FAD), single decisions (→ ADR), and bounded comparisons (→ WEP matrix).

**The methodology tools are [GENERATOR.md](GENERATOR.md), [GENERATOR-DECISION.md](GENERATOR-DECISION.md), and [GENERATOR-COMPARISON.md](GENERATOR-COMPARISON.md).** The MCP server (`cmd/vivechak/`) automates orchestration — workspace setup, DAG resolution, context injection, validation, and exit gates. Everything else supports them.

---

## 2. Using Vivechak for a New Project

This is the common case — an agent is asked to generate and/or execute a research pipeline for a project.

### Generate the Pipeline

1. Open [GENERATOR.md](GENERATOR.md) and extract the prompt from inside the ```` ````markdown ```` code fences.
2. Replace `[PASTE YOUR PROJECT DESCRIPTION HERE]` with the project vision.
3. Execute the prompt in a fresh AI session with web search enabled.
4. You will receive **two files**: `RESEARCH-PIPELINE.md` (session DAG + prompts) and `DECISIONS.md` (initial decision registry).

### Set Up the Project Workspace

5. Save both generated files to the project's `research/` directory.
6. Copy the **5 templates** from `templates/` into `research/templates/`:
   - `DECISIONS.template.md` — for recording architectural decisions
   - `CONFLICT-RESOLUTION.template.md` — for resolving conflicting findings
   - `COMPARISON-SESSION.template.md` — for structuring option comparisons
   - `FOUNDING-ARCHITECTURE.template.md` — for compiling the final FAD
   - `PHASE-0-GATE.template.md` — for the pre-coding exit gate

```
project/
└── research/
    ├── RESEARCH-PIPELINE.md         ← Generated
    ├── DECISIONS.md                 ← Generated
    ├── sessions/                    ← Empty folder for session outputs
    └── templates/                   ← Copied from Vivechak
        ├── DECISIONS.template.md
        ├── CONFLICT-RESOLUTION.template.md
        ├── COMPARISON-SESSION.template.md
        ├── FOUNDING-ARCHITECTURE.template.md
        └── PHASE-0-GATE.template.md
```

**The project workspace is now 100% self-contained.** No further reference to this repository is needed.

### Execute the Pipeline

7. Take each prompt from the generated `RESEARCH-PIPELINE.md` and run it in an independent AI research session. Save outputs to `research/sessions/`.
8. After key sessions, record decisions in `DECISIONS.md` using the `DECISIONS.template.md` format. Classify each as one-way or two-way door.
9. If sessions produce conflicting recommendations, resolve using `CONFLICT-RESOLUTION.template.md`.
10. Synthesize all findings into `FAD.md` using `FOUNDING-ARCHITECTURE.template.md`.
11. Verify exit criteria with `PHASE-0-GATE.template.md`. Once the gate passes — start coding.

> **You do NOT need to read FRAMEWORK.md to use Vivechak.** GENERATOR.md is self-contained — all methodology is operationalized inline in the generator prompt and baked into every generated session prompt.

---

## 3. Developing Vivechak Itself

This section applies only when improving the meta-framework — editing FRAMEWORK.md, GENERATOR.md, or templates/.

1. **Read [FRAMEWORK.md](FRAMEWORK.md)** — the complete methodology specification (8 principles, evidence grading, pipeline topology, prompt anatomy).
2. **Read [CONTRIBUTING.md](CONTRIBUTING.md)** — all changes must be traceable to evidence.
3. **⚠️ Run the Change Propagation Map** (CONTRIBUTING.md § Change Propagation Map) before committing. FRAMEWORK.md is the specification; GENERATOR.md is the implementation. Changing the spec without updating the implementation means improvements exist in documentation but never affect generated pipelines. Every change to FRAMEWORK.md, GENERATOR.md, or templates/ MUST be checked against the propagation map.
4. All changes must follow the 8 Core Principles (summarized below).
5. The sealed `meta-research/` directory contains the empirical evidence base — 15 research artifacts validating every design decision.

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
├── CONTRIBUTING.md                 ← Evidence-grounding contribution rules
├── CODE_OF_CONDUCT.md · LICENSE    ← Contributor Covenant v2.1 · MIT License
├── VERSION                         ← Release version
├── go.mod · go.sum                 ← Go 1.27+ module
├── .goreleaser.yml                 ← 6-platform cross-compilation
├── install.sh · install.ps1        ← Shell installers
│
├── cmd/vivechak/                   ← MCP server + CLI entry point
│   ├── main.go                     ← Entry point (serve, version, mcp-config, doctor)
│   ├── config.go                   ← mcp-config --client <host> --write
│   └── doctor.go                   ← Workspace integrity checker
│
├── internal/                       ← Server implementation (Go)
│   ├── core/                       ← Pure logic (workspace, DAG, validation, injection)
│   ├── mcp/                        ← 9 MCP tool handlers + envelope
│   ├── store/                      ← Atomic I/O, locking, os.Root confinement
│   └── embed/                      ← go:embed generators + templates
│
├── templates/                      ← 5 operational contracts (copy to projects)
│   ├── DECISIONS.template.md       ← YAML frontmatter ADR format
│   ├── CONFLICT-RESOLUTION.template.md ← ACH-style conflict resolution
│   ├── COMPARISON-SESSION.template.md  ← WEP comparison output format
│   ├── FOUNDING-ARCHITECTURE.template.md ← Map-Reduce synthesis to FAD
│   └── PHASE-0-GATE.template.md    ← Two-track pre-codebase exit gate
│
├── examples/                       ← Concrete adoption walkthroughs
│   └── SAMPLE-PIPELINE.md          ← End-to-end MCP workflow sample (Katha project)
│
├── docs/                           ← Documentation
│   ├── QUICKSTART.md               ← Vivechak in 5 Minutes
│   ├── MANUAL-WORKFLOW.md          ← Complete manual copy-paste guide
│   ├── HOST-SETUP.md               ← Per-host MCP config (7 hosts)
│   ├── MCP-TOOLS.md                ← 9-tool reference with examples
│   ├── ARCHITECTURE.md             ← Server internals for contributors
│   ├── assets/logo.jpg             ← Minimalist prism logo
│   └── meta-research/              ← Empirical evidence base
│       ├── README.md               ← Provenance index
│       ├── DECISIONS.md            ← All 14 locked ADRs (D-001–D-014)
│       ├── RESEARCH-PIPELINE-v1.md ← v1.0: meta-research execution DAG
│       └── v2/                     ← v1.1: MCP server evidence
│           ├── FINAL-PLAN.md       ← Definitive 5-phase plan
│           └── RESEARCH-PIPELINE-v2.md ← v1.1: 6-session pipeline
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
2. **Embed sync:** `internal/embed/generators/` and `internal/embed/templates/` must stay in sync with root-level generators and templates. Verify manually before committing (a CI check is planned but not yet implemented).
3. **9 MCP tools:** `vivechak_init`, `vivechak_prepare_generator`, `vivechak_save_plan`, `vivechak_status`, `vivechak_next_session`, `vivechak_save_session`, `vivechak_record_decision`, `vivechak_validate`, `vivechak_run_gate`. Each tool is one file in `internal/mcp/tool_*.go`.
4. **Guided Worker pattern:** Every tool response includes `next_step`. Hard refusals ONLY for impossible operations, never for "wrong order."
5. **Testing:** `go test ./...` runs all tests. `internal/mcp/server_test.go` has wire-level integration tests.

---

## 6. Key Operational Rules

- **No legacy dogma:** Do not use 8-section XML prompts, expert personas, hardcoded search queries, minimum search counts, or rigid output skeletons. These are empirically refuted.
- **Front-load briefs:** Never drip-feed instructions across turns (39% performance drop documented).
- **Grade everything:** Every factual claim needs an inline evidence grade with modifiers and verification method.
- **Two-Way Doors move fast:** Don't over-research reversible decisions. Spike or decide by convention.
- **One-Way Doors move carefully:** Require corroborated Grade A/B evidence, locked ADR, and premortem before commitment.

