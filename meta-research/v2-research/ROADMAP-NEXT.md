# Vivechak: Next Roadmap

> **Created:** 2026-09-23
> **Basis:** First-principles analysis across competitive landscape, AI trends, methodology critique, and MCP architecture research. Synthesized from `what-next-old.md` (actionable items) and `WHATS-NEXT.md` (deep analysis).

---

## Strategic Direction (One Sentence)

**Vivechak evolves from a pre-development-only prompt into a multi-scope research methodology delivered as an MCP-based Agent Plugin.**

---

## Why This Direction

### The Core Tension

Vivechak has a genuine insight — *"how you frame a research question measurably biases what evidence AI models report"* — validated by peer-reviewed research (Zheng et al. EMNLP 2024, Laban et al. ICLR 2026, Kim et al. ICML 2025) and increasingly relevant as "vibe coding" produces architectural disasters. No competitor tool addresses this with structured multi-session research, evidence grading, and decision tracking.

But Vivechak delivers this insight as **106KB of documentation and a copy-paste workflow.** The framework is right. The delivery is wrong. Fix the delivery.

### Competitive Position (September 2026)

**No direct competitor exists.** The closest are behavioral patterns (asking ChatGPT ad-hoc), partial tools (MADR/adr-tools for recording, Elicit for academic research, Taskade for AI-critiqued design), and emerging threats (deep research tools gaining multi-session capability within 6-12 months).

**The window is closing.** If the methodology isn't made accessible before a competitor packages something simpler, the first-mover advantage is lost.

### Target Audience

The **"post-vibe" developer** — someone who tried vibe coding, hit an architectural wall at scale, and realized they need structured research. These developers don't search for "pre-development research frameworks." They search for "how to choose a database for my SaaS." Vivechak's marketing must meet them where their pain is.



## Phase 5 (Revised): From Engine to MCP Plugin

> The original Phase 5 planned a standalone "Vivechak Engine" — a custom CLI/web service requiring a framework choice (LangGraph vs ADK vs CrewAI). That approach is **replaced**. Rationale: One-Way Door risk (volatile framework landscape), zero adoption data, compound agent failure risk (55% success over 30 steps), and MCP achieving 80% of the value at 10% of the risk.

### Wave 1: Methodology Expansion — Multi-Scope Vivechak

**Goal:** Vivechak works at 3 scope levels, not just project-level.

| Deliverable | What to Build | Files Affected | Door Type |
|---|---|---|---|
| **Decision-level generator prompt** | A shorter generator prompt (~5KB) that takes a "decision context" (not a project vision) and produces 1–3 focused research session prompts + a single ADR template. Input: what decision, what constraints, what stage of the project. Output: lightweight pipeline + proposed ADR. | New: `GENERATOR-DECISION.md` | Two-Way |
| **Comparison-level generator prompt** | A minimal prompt (~2KB) that takes specific options + criteria and produces a single research session prompt using the Weighted Evaluation Protocol. Input: "Option A vs Option B for [context]". Output: single session prompt with WEP structure. | New: `GENERATOR-COMPARISON.md` | Two-Way |
| **Framework updates for multi-scope** | Add §9 to FRAMEWORK.md: "Multi-Scope Research" documenting the 3 levels (Project, Decision, Comparison), when to use each, and how complexity scoring adapts per scope. | Modify: `FRAMEWORK.md` | Two-Way |
| **README/AGENTS updates** | Update quick-start to show all 3 entry points. Add examples for each scope level. | Modify: `README.md`, `AGENTS.md` | Two-Way |
| **Reposition tagline** | Change from "pre-development research" to "proportional architecture research" or "evidence-grounded research for technical decisions" across all files. | Modify: `README.md`, `AGENTS.md`, `GENERATOR.md` BRIEF | Two-Way |
| **Context engineering acknowledgment** | Add note to P1 in FRAMEWORK.md linking it to the "context engineering" paradigm. Address backlog item PE-01. | Modify: `FRAMEWORK.md` | Two-Way |

**Dependency:** None. Pure methodology work. Can start immediately.

**Validation:** Use the decision-level generator to research the MCP architecture decision for Vivechak itself (dogfooding). This simultaneously validates the new scope level AND produces grounded evidence for Wave 2.

---

### Wave 2: MCP Server — "The Guide"

**Goal:** A Python MCP server that manages Vivechak workflow, validates outputs, and tracks state. The host agent (Claude, Antigravity, Cursor, etc.) does the actual thinking and research.

**Architecture:** Approach B — the MCP server is a project manager, not a worker. It doesn't call LLM APIs or do web research. It provides methodology structure, validates compliance, and manages DAG state.

#### Core Tools (Build in this order)

| # | Tool | What It Does | Priority | Complexity |
|---|---|---|---|---|
| 1 | `vivechak_init` | Creates `research/` directory, copies 4 templates, initializes `.vivechak/state.json` | Must-have | Simple |
| 2 | `vivechak_get_generator_prompt` | Reads the appropriate generator prompt (project/decision/comparison), inserts user's vision/context, returns the complete prompt ready for the agent to execute | Must-have | Medium |
| 3 | `vivechak_save_pipeline` | Parses agent-generated pipeline + decisions output, validates structure, saves to correct files, initializes session tracking in state | Must-have | Medium |
| 4 | `vivechak_status` | Scans workspace, reports: sessions done/pending, decisions by status, DAG progress, what's next | Must-have | Medium |
| 5 | `vivechak_next_session` | Checks DAG dependencies, returns the next actionable session(s) with their full prompts | Must-have | Medium |
| 6 | `vivechak_save_session` | Validates a completed session output (evidence grades present? YAML valid? minimum quality rubric?) and saves to `sessions/` | Must-have | Medium |
| 7 | `vivechak_record_decision` | Validates ADR against 18-field schema, saves/updates in `DECISIONS.md` | Must-have | Medium |
| 8 | `vivechak_validate` | Validates any Vivechak artifact (session, ADR, FAD, pipeline) against its schema | Should-have | Medium |
| 9 | `vivechak_synthesize` | Collects all finalized sessions + locked decisions, injects into FAD template, returns the synthesis context for the agent to compile | Should-have | Medium |
| 10 | `vivechak_run_gate` | Executes Phase 0 Gate checks programmatically: evidence coverage, decision lock status, conflict resolution, returns structured pass/fail | Should-have | Complex |

#### Technology Stack

| Component | Choice | Rationale |
|---|---|---|
| Language | Go 1.22+ | Single binary distribution, no runtime dependency, cross-platform. Server does file I/O + validation, not AI — Python's LLM advantage is irrelevant. See `temp/PYTHON-VS-GO.md` for full evaluation. |
| MCP SDK | `github.com/modelcontextprotocol/go-sdk` | Official Tier 1 SDK, maintained by MCP org + Google. Full protocol support. |
| Transport | stdio (primary) | Works with all local agents. 5–20ms Go startup vs 50–150ms Python. |
| State | File-based `.vivechak/state.json` per project | No database. State lives with the project. |
| Validation | Go struct tags + `gopkg.in/yaml.v3` | Compile-time type safety, auto JSON schema from struct tags. |
| Packaging | Agent Plugins 1.0.0 spec | `plugin.json` manifest + `mcp.json` config. Maximum distribution. |
| Distribution | GitHub Releases (binaries) + `go install` + Agent Plugins | Zero-dependency binary. No "do you have Python 3.11?" problem. |

#### Repository Structure

```
vivechak/                          ← Existing repo (methodology stays here)
├── GENERATOR.md                   ← Existing project-level generator
├── GENERATOR-DECISION.md          ← NEW: decision-level generator  
├── GENERATOR-COMPARISON.md        ← NEW: comparison-level generator
├── FRAMEWORK.md                   ← Updated with §9 Multi-Scope
├── templates/                     ← Existing (embedded in Go binary)
├── cmd/                           ← NEW: MCP server entry point
│   └── vivechak/
│       └── main.go                ← Server startup, transport config
├── internal/                      ← NEW: Internal packages
│   ├── tools/                     ← One file per tool
│   │   ├── init.go
│   │   ├── generator.go
│   │   ├── pipeline.go
│   │   ├── session.go
│   │   ├── decision.go
│   │   ├── status.go
│   │   ├── validate.go
│   │   ├── synthesize.go
│   │   └── gate.go
│   ├── models/                    ← Go structs with validation tags
│   │   ├── pipeline.go
│   │   ├── session.go
│   │   ├── decision.go
│   │   └── state.go
│   └── parsers/                   ← Markdown/YAML parsing
│       ├── pipeline.go
│       └── frontmatter.go
├── plugin.json                    ← NEW: Agent Plugins 1.0.0 manifest
├── mcp.json                       ← NEW: MCP server configuration
├── go.mod                         ← NEW: Go module
├── go.sum                         ← NEW: Go dependencies
├── Makefile                       ← NEW: Build targets (cross-compile)
└── tests/                         ← NEW: Test suite
    ├── init_test.go
    ├── generator_test.go
    ├── pipeline_test.go
    ├── validation_test.go
    └── testdata/                   ← Sample pipelines, sessions, ADRs
```

**Key decision: Monorepo.** The MCP server lives inside the existing vivechak repo. Rationale: the server reads GENERATOR.md and templates/ directly. Separate repos create sync problems. The methodology IS the product; the server is the delivery mechanism.

**Dependency:** Wave 1 should be done first (the MCP server needs the decision/comparison generators to exist). However, tools 1–7 can start in parallel with Wave 1 since they work with the existing project-level generator.

---

### Wave 3: Packaging & Distribution

**Goal:** Make Vivechak installable with one command from any MCP-compatible agent.

| Deliverable | What to Do |
|---|---|
| **Cross-platform binaries** | `Makefile` with `GOOS/GOARCH` targets → GitHub Releases with binaries for Windows/Mac/Linux (amd64 + arm64) |
| **`go install` support** | `go install github.com/vivechak/vivechak@latest` for Go developers |
| **Agent Plugins 1.0.0 compliance** | `plugin.json` + `mcp.json` + `skills/` directory |
| **Installation guides** | "Add Vivechak to Claude Desktop" / "Add to Cursor" / "Add to VS Code" / "Add to Antigravity" — one page each |
| **"Vivechak in 5 Minutes" quick-start** | A single page: install → init → generate → research → decide. Replaces the 106KB documentation barrier. |

**Dependency:** Wave 2 tools 1–7 minimum.

---

### Wave 4: Content & Demand Proof

**Goal:** Validate that people want this. Without adoption data, every decision after this point is blind.

| Deliverable | What to Do | Why It Matters |
|---|---|---|
| **Real case study #1** | Run a full Vivechak pipeline on a real project (even Vivechak's own MCP server architecture). Publish the complete research/ directory (anonymized if needed) as `examples/mcp-server-research/` | The current SAMPLE-PIPELINE.md is synthetic. Real artifacts are 10× more compelling. |
| **"Architecture Research for [X]" blog series** | Publish research pipeline outputs for common project types (SaaS backend, CLI tool, mobile app). Post them where developers actually look (DEV.to, HN, Reddit r/programming). | Solo developers search for "how to choose a database," not "research frameworks." This meets them at their actual search queries. Single highest-impact demand generation strategy. |
| **Positioning vs MADR/SMADR** | Add a section to README: "How Vivechak relates to MADR." Key message: MADR records decisions, Vivechak generates the research that informs them. Addresses backlog OP-03. | Structured MADR 1.0 converges on Vivechak's ADR format. Must differentiate before confusion sets in. |
| **GitHub Discussions enabled** | Turn on Discussions in the repo. Create categories: "Show & Tell", "Q&A", "Feature Requests" | Zero feedback channels = zero adoption signal. Can't improve what you can't measure. |
| **README "Who is this for?" section** | Target the "post-vibe" developer — someone who tried vibe coding, hit an architectural wall, and realized they need structured research. NOT the developer who already knows they need it. | The market is invisible. Meet developers where their pain is, not where the tool lives. |

**Dependency:** Wave 3 (need installable product to drive adoption).

---

## Open Backlog — Carried Forward & Updated

Items from the existing ROADMAP backlog, updated with current analysis:

| ID | What | Status | Target |
|---|---|---|---|
| SM-01 | Domain/Technical Novelty dimension overlap risk | **Open** — needs 3+ project calibration data | Wave 5+ |
| SM-02 | Sharp tier boundary cliff at 15→16 | **Open** — needs calibration data | Wave 5+ |
| GA-02 | No protocol for human stakeholder disagreement | **Open** — low urgency | Wave 5+ |
| GA-04 | Quality rubric evaluation process undefined | **Partially addressed** — MCP server's `vivechak_validate` tool will programmatically check the rubric | Wave 2 |
| ES-03 | Missing "tool-generated" verification method | **Open** — add when MCP server generates evidence via code execution | Wave 5+ |
| OP-02 | No worked example of a failed pipeline | **Open** — create when real failure data available | Wave 4 (if failure occurs) |
| OP-03 | Positioning vs Structured MADR 1.0 | **Scheduled** | Wave 4 |
| DA-03 | Context injection token budget not quantified | **Open** — quantify from MCP server usage data | Wave 5+ |
| PE-01 | "Context engineering" terminology | **Scheduled** | Wave 1 |
| AB-03 | One-way door inter-service patterns | **DONE** ✅ | — |

### New Backlog Items (From This Analysis)

| ID | What | Source | Target |
|---|---|---|---|
| NEW-01 | FAD Amendment Protocol — what happens when a sealed FAD is proven wrong by development reality | First-principles critique | Wave 5+ |
| NEW-02 | Narrow the "10× cost at month 6" claim to data/infrastructure decisions specifically — AI has flattened the cost curve for some decisions | Market research | Wave 1 (README update) |
| NEW-03 | Re-evaluate P1 decomposition threshold for 1M+ context windows | Model evolution analysis | Wave 5+ (evidence-gated) |
| NEW-04 | Re-evaluate P7 triangulation value given increased model diversity | Model evolution analysis | Wave 5+ (evidence-gated) |
| NEW-05 | Monitor deep research tools for multi-session pipeline capability | Competitive threat analysis | Ongoing |

---

## What NOT To Do

| Don't | Why |
|---|---|
| **Don't build a standalone Engine** | Replaced by MCP server. The One-Way Door risk (framework choice) is eliminated by the MCP approach. |
| **Don't add more methodology documentation** | 106KB is already too much. Wave 1 adds generators and a framework section, but the "5 Minutes" guide is the priority access path. |
| **Don't actively pursue non-software domains** | Keep "technical projects" language but don't build domain-specific generators for biotech/hardware. Zero demand signal. |
| **Don't chase Phase 6 frontiers yet** | DSPy, knowledge graphs, multi-agent debate — all gate conditions unmet. Require MCP usage data that doesn't exist. |
| **Don't bump the version for Waves 1–3** | These are delivery improvements, not methodology changes. The methodology version stays 1.1 unless the decision/comparison generators reveal a fundamental principle change. |

---

## Execution Timeline

```
Wave 1: Methodology Expansion         ████████░░░░░░░░░░░░  ~2 weeks
  ├─ Decision-level generator prompt
  ├─ Comparison-level generator prompt  
  ├─ FRAMEWORK.md §9 + repositioning
  ├─ README/AGENTS updates + narrow "10× cost" claim
  └─ Dogfood: research MCP decision using decision-level generator

Wave 2: MCP Server                    ░░░░░░████████████░░  ~3-4 weeks
  ├─ Core tools 1-7 (init → record_decision)
  ├─ Pydantic validation models
  ├─ State management
  ├─ Tools 8-10 (validate, synthesize, gate)
  └─ Test suite

Wave 3: Packaging & Distribution      ░░░░░░░░░░░░░░████░░  ~1 week
  ├─ PyPI package
  ├─ Agent Plugins 1.0.0 compliance
  ├─ Installation guides (4 agents)
  └─ "Vivechak in 5 Minutes" guide

Wave 4: Content & Demand Proof        ░░░░░░░░░░░░░░░░████ ~1-2 weeks + ongoing
  ├─ Real case study #1 (Vivechak's own MCP architecture)
  ├─ "Architecture Research for [X]" blog series (ongoing)
  ├─ MADR/SMADR positioning
  ├─ GitHub Discussions
  └─ "Who is this for?" README section

Total: ~7-9 weeks to installable product with real case study
```

---

## Success Criteria

After Waves 1–4 are complete, Vivechak will have:

- [ ] 3 scope levels operational (project / decision / comparison)
- [ ] Installable MCP server (`go install` or binary download)
- [ ] Agent Plugins 1.0.0 compliant packaging
- [ ] Works with ≥2 agents (Antigravity + Claude Desktop minimum)
- [ ] ≥1 real case study published
- [ ] ≥1 "Architecture Research for [X]" blog post published
- [ ] GitHub Discussions active
- [ ] Adoption metrics trackable (install count, stars, discussions)
- [ ] Time from vision to FAD measured (baseline manual vs MCP-assisted)

**The decision to proceed to Phase 6 (DSPy, knowledge graphs, etc.) is gated on adoption data from Wave 4.** If nobody installs the MCP server, the correct response is to understand why, not to build more features.

---

## Decision Log for This Roadmap

| Decision | Type | Rationale |
|---|---|---|
| Replace Engine with MCP Server | One-Way → Two-Way | Eliminates framework lock-in risk. MCP is the universal standard. Approach B ("The Guide") leverages host agent capabilities instead of duplicating them. |
| Monorepo (server inside vivechak/) | Two-Way | Server reads GENERATOR.md and templates/ directly. Separate repos create sync overhead with zero benefit at current scale. Reversible if needed. |
| Go for MCP server | Two-Way | Single binary distribution eliminates runtime dependency friction. Server does file I/O + validation, not AI — Python's LLM ecosystem is irrelevant. Go SDK is Tier 1 official. See `temp/PYTHON-VS-GO.md`. |
| Agent Plugins 1.0.0 packaging | Two-Way | Thin wrapper. Maximum distribution. No lock-in. Complementary to MCP (MCP=runtime protocol, Plugins=packaging format). Note: Anthropic is NOT on the Agent Plugins steering committee — monitor for ecosystem split. |
| Three scope levels | Two-Way | Additive (new generators, not modifications to existing). Can be reverted by deleting 2 files. |
| Don't bump version | Two-Way | Waves 1–3 are delivery changes, not methodology changes. Version bump if/when methodology principles change. |

---

## Risk Assessment

| Risk | Severity | Mitigation |
|---|---|---|
| **"Window closing"** — Deep research tools gain multi-session pipeline capability, eating Vivechak's orchestration value | High | Ship MCP server fast. Vivechak's differentiation is evidence grading + structured falsification + decision tracking, not just multi-session orchestration. But this advantage has a shelf life. |
| **"Over-engineered prompt library"** — perception that 106KB of docs for one prompt + four templates is disproportionate | Medium | MCP server + "5 Minutes" guide transforms perception. Methodology stays rigorous; surface area shrinks to tool calls. |
| **"Self-referential echo chamber"** — Vivechak was validated by its own methodology, creating circularity | Medium | Foundational claims (framing bias, persona debunking, drip-feed penalty) are grounded in *external* peer-reviewed research. Circularity exists in meta-layer but not in evidence base. Acknowledge this limitation explicitly. |
| **"Complexity scoring is pseudoscience"** — ordinal scores summed across dimensions with hard tier boundaries = false precision | Medium | Already partially addressed (SM-01, SM-02). Explicitly document scoring as a "sizing heuristic, not a precision instrument." Smooth tier boundaries when calibration data exists. |
| **Zero adoption** — building for nobody | High | Wave 4 is explicitly designed to test demand. If nobody installs the MCP server or engages via Discussions, the correct response is to understand why, not build more. |
| **Agent Plugins ecosystem split** — Anthropic not on steering committee could mean MCP and Agent Plugins diverge | Low | Agent Plugins packaging is a thin wrapper (Two-Way Door). Can adapt quickly. |

### Stale Assumptions to Monitor

| Assumption | Original Evidence | Current Risk | Monitoring Strategy |
|---|---|---|---|
| P1: Context windows require decomposition | BrowseComp R²=0.80 (Aug 2026) | Medium — windows now 1M-4M tokens | Track if single-session coverage replaces Tier 2 pipelines |
| P7: Triangulation has diminishing returns | Correlated errors ~60% (Kim et al. ICML 2025) | Low-Medium — model diversity increasing | Check error correlation studies annually |
| 8-dimension scoring → session count | 8-project calibration | Medium — no external validation | Collect predicted vs actual sessions across next 3 projects |
| "10× cost at month 6" for wrong architecture | Industry rule of thumb | Medium — AI has flattened cost curve for SOME decisions | Narrow claim to data/infrastructure in Wave 1 (NEW-02) |

---

## Appendix: Evidence Sources

| Source | Grade | Used For |
|---|---|---|
| Zheng et al. EMNLP 2024 (persona debunking) | A | P4 stability validation |
| Laban et al. ICLR 2026 (drip-feed penalty) | A | Prompt anatomy stability |
| Kim et al. ICML 2025 (correlated LLM errors) | A | P7 staleness evaluation |
| Tam et al. EMNLP 2024 (format restriction) | A | P4 stability validation |
| MCP official specification (Anthropic/AAIF) | A | MCP strategy recommendation |
| Agent Plugins 1.0.0 specification (Aug 2026) | A | Packaging strategy |
| SWE-bench 2026 leaderboard | A | Agent reliability assessment |
| Agent reliability studies (LangChain, Temporal, Fiddler AI) | B | Compound failure risk (98% per-step → 55% over 30 steps) |
| LangGraph / ADK / CrewAI / OpenAI Agents SDK docs | A | Framework landscape (volatile, no winner) |
| Gemini / OpenAI / Claude deep research docs | A | Research tool evolution |
| MADR 4.0 / Structured MADR 1.0 repos | A | ADR landscape, positioning |
| Andrej Karpathy on vibe coding (2025) | A | Audience identification |
| Building Evolutionary Architectures (O'Reilly) | A | Evolutionary architecture perspective |
| GitHub Octoverse 2024-2025 | A | Market sizing |
| ThoughtWorks Technology Radar 2025-2026 | B | Industry trends |
| Taskade, Workik, Elicit analysis | B-C | Competitor assessment |
