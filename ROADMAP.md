# Vivechak Roadmap & Future Direction
### Autonomous Evidence-Grounded Research for Technical Decisions

---

## 1. Current State: v0.1.0 (October 2026)

**Status:** Shipped, battle-tested, and automated via Model Context Protocol (MCP).

Vivechak v0.1.0 delivers a unified, evidence-grounded research methodology powered by an autonomous, single-binary Go MCP server and multi-scope generators. It operates across three distinct scope levels (Project → FAD, Decision → ADR, Comparison → WEP matrix) while maintaining a strict Guided Worker pattern with zero LLM API dependencies.

### What v0.1.0 Delivers

| Component | State | Description |
|---|---|---|
| **13 Atomic MCP Tools** | ✅ Shipped | Full lifecycle support: `init`, `prepare_generator`, `save_plan`, `status`, `next_session`, `save_session`, `record_decision` (with ADR supersession), `amend_session`, `validate`, `run_gate`, `challenge` (adversarial red-teaming), `replan` (in-flight DAG mutation), and `visualize` (Mermaid graph + status table) |
| **Multi-Scope Generators** | ✅ Shipped | `GENERATOR.md` (Project), `GENERATOR-DECISION.md` (Decision), `GENERATOR-COMPARISON.md` (Comparison) with synchronized `CORE:BEGIN/END` methodology kernels and Research Depth Mandates |
| **Namespaces & Protocols** | ✅ Shipped | `C-NNN` comparison namespace, `T-NNN` project tasks, `D-NNN` ADRs; ADR supersession protocol with transitive stale-reference detection |
| **CLI & Diagnostics** | ✅ Shipped | `vck` shorthand alias, `vck setup` (1-second host configuration + 14 presets), `vck mcp-config` (universal JSON config), and `vck doctor` workspace integrity checker |
| **FRAMEWORK.md** | ✅ Shipped | Complete standalone specification (8 principles, evidence grading A–E, 8-block prompt anatomy, Weighted Evaluation Protocol, Phase 0 Gate) |
| **6 Operational Templates** | ✅ Shipped | `DECISIONS.template.md`, `CONFLICT-RESOLUTION.template.md`, `COMPARISON-SESSION.template.md`, `FOUNDING-ARCHITECTURE.template.md`, `PHASE-0-GATE.template.md`, `SESSION.template.md` (`schema_version: "0.1.0"`) |
| **Distribution & Packages** | ✅ Shipped | Single-binary cross-compilation (6 platforms via GoReleaser), Homebrew, Scoop, WinGet, shell installers (`install.sh`, `install.ps1`) |
| **meta-research/** | ✅ Sealed | 14 locked ADRs (D-001–D-014), 27 research artifacts, empirical evidence base validating every design decision |

---

## 2. Deferred Features & Backlog (Tracked & Prioritized)

Every deferred feature is tracked with its origin, explicit gate condition, and design rationale. We explicitly reject feature creep; features are built only when evidence demonstrates demand.

### 2.1 Near-Term Packaging & Workflows (Awaiting Demand Signal)

| Feature | Description | Origin | Gate Condition | Est. Effort |
|---|---|---|---|---|
| **Incremental Session Building** | `vivechak_add_finding` / `vivechak_save_draft` tool endpoints allowing agents to stream findings incrementally rather than writing complete sessions. | Definitive Plan / `API-03` | 3+ user requests showing context window overflow on single session writes | Medium |
| **`vivechak_export`** | One-command compilation of FAD and ADR registry into self-contained HTML/PDF reports with embedded SVG charts and offline citation snapshots. | Definitive Plan / Missing Tools | User demand for stakeholder presentation beyond Markdown | Low |
| **Thin Agent Plugin Package** | `plugin.json` + `mcp.json` + quick-start skills bundle for IDEs supporting the Agent Plugins standard (Cursor, VS Code extensions). Supplementary to the native binary. | ADR D-014 / Meta-Research R-05 | Demand from users on harnesses with native plugin discovery; or Claude Desktop/Antigravity adding Agent Plugins support | Low |
| **MCP Registry Listing** | Official `server.json` manifest submission to the canonical Model Context Protocol Registry (`modelcontextprotocol/registry`). | FINAL-PLAN.md Phase 3 | MCP registry open submission window reaches public release | Low |
| **Claude Desktop Extension (`.mcpb`)** | Anthropic-native `.mcpb` bundle package for one-click extension install without manually editing `claude_desktop_config.json`. | FINAL-PLAN.md Phase 3 | Anthropic stabilizes and documents `.mcpb` distribution specification | Low |
| **Binary Code Signing** | Apple Developer ID notarization (`gon`) and Windows Authenticode (Azure Trusted Signing) to eliminate SmartScreen/Gatekeeper warnings. | FINAL-PLAN.md Phase 0.4 | Community download volume exceeds 500+ weekly downloads | Medium |
| **Post-Save Git Hooks** | Optional auto-commit hook (`vck-git`) that commits completed sessions and compiled ADR registries into git history automatically. | Definitive Plan / `SL-03` | Demand from solo developers wanting zero-overhead git tracking | Low |
| **TUI Dashboard (`vck tui`)** | Interactive terminal dashboard for human architects to monitor parallel agent research progress, view Mermaid DAGs, and review decisions. | Definitive Plan / `Polish 1` | Demand for terminal-first visual monitoring | Medium |
| **Cross-Session Field Guide** | Comprehensive operational guide for running 20+ session pipelines across distributed agent sessions. | Audit Backlog / `T3-02` | Multi-day project research pipelines in production | Low |
| **Examiner Calibration Questions** | Pre-generated calibration questions for human architects to probe agent-generated FADs during Phase 0 reviews. | Audit Backlog / `T3-03` | Team adoption where human architects review autonomous agent research | Low |
| **Worked Failed Pipeline Example** | A complete negative example in `examples/FAILED-PIPELINE.md` demonstrating failure recovery, contradictory evidence resolution, and gate rejection. | Audit (OP-02) | Real failure logs harvested from live dogfooding | Low |
| **Search Budget Density Tracking** | Density tracking of searches per decision rather than hardcoded quotas (D-UX-02). | Dogfooding Synthesis (F-2) | Live telemetry showing agents continuing to skimp searches despite qualified provenance | Low |
| **Structured Evidence Receipts** | Cryptographic/structured hash receipts for individual evidence citations. | Dogfooding Synthesis (F-1) | Regulated industry adoption requiring verifiable chain-of-custody | Medium |
| **Belief Evolution Ledger Injection** | Dynamic injection of historical belief deltas across DAG layers to highlight confirmation bias. | Dogfooding Synthesis (F-3) | Multi-layer pipelines where Q-BIAS advisories fail to curb confirmation drift | Low |
| **Cumulative Prior Tracker** | Cross-layer tracking of accumulated architectural biases across 5+ session DAGs. | Dogfooding Synthesis (P1-2) | Longitudinal evidence that dynamic routing alone is insufficient | Medium |

### 2.2 Enterprise & Ecosystem Extensions (v0.2.x+)

| Feature | Description | Origin | Gate Condition | Est. Effort |
|---|---|---|---|---|
| **C4 Architecture & Backstage Export** | Automatically export synthesized FAD architectures into C4 model container/component diagrams and Spotify Backstage `catalog-info.yaml` entities. | Audit Backlog / `Polish 3` | Enterprise adoption requiring architectural portal integration | Medium |
| **Composable Research Packs (`vck import`)** | Domain-curated technology profiles and evidence packs (e.g. FinTech compliance, Edge IoT, Web3 security) imported directly into new pipelines. | Audit Backlog / `Polish 2` | Demand for reusable domain seed evidence across multiple client teams | Medium |
| **Modular DAG JSON Architecture** | JSON-based modular representation for massive enterprise pipelines (>30 sessions) replacing single-document markdown DAGs. | Forensic Audit / `TE-01` | Research pipelines scaling beyond 30 DAG-ordered sessions | Medium |
| **Multi-Currency & Regional Cost Modeling** | Currency-agnostic and region-aware infrastructure pricing models in WEP matrices (EUR, GBP, INR, JPY alongside USD). | Forensic Audit / `EC-02` | Multi-national enterprise teams operating multi-cloud infrastructure | Low |
| **Calibration Daemon (`vck audit-calibration`)** | Background CI sentinel comparing ADR cost, latency, and throughput predictions against live telemetry and monitoring metrics. | Audit Backlog / `Polish 4` | Production projects with 6+ months post-deployment telemetry | High |

### 2.3 Methodology Evolution & Epistemic Backlog

| Proposal | Description | Origin | Gate Condition |
|---|---|---|---|
| **PRISMA-ScR Systematic Review Checklist** | Integrating PRISMA-ScR (Preferred Reporting Items for Systematic Reviews and Meta-Analyses) guidelines into evidence synthesis protocols. | Audit Backlog / `NEW-08` | Academic or safety-critical engineering research pipelines |
| **Empirical Re-evaluation of P1 (1M+ Context)** | Re-evaluating Principle P1 (Context Architecture Law) in light of 1M–2M context frontier LLMs to measure information loss during large synthesis runs. | Audit Backlog / `NEW-03` | Frontier LLMs establishing reliable 1M+ effective context recall benchmarks |
| **Empirical Re-evaluation of P7 (Reasoning Models)** | Testing whether test-time compute reasoning models (o1/o3, Flash Thinking) reduce the necessity for multi-model triangulation on One-Way Doors. | Audit Backlog / `NEW-04` | Empirical benchmarks on hallucination rates in frontier reasoning models |
| **Semantic Circularity & Deduplication Engine** | Detecting circular citations where multiple Grade B/C secondary sources re-cite the same underlying vendor press release. | Forensic Audit / `EI-02` | Agent harness support for scraping citation link graphs |
| **Tool-Generated Benchmark Verification Tier** | Introducing an explicit "tool-generated" evidence tier (`ES-03`) distinct from "fetched" vs "recalled" for empirical spike runs. | Audit Backlog / `ES-03` | Adoption of the Empirical Sandbox Spike Runner |
| **Real Options Engineering (Valuation Windows)** | Introducing formal financial real-options framing for `deferred` decisions with explicit valuation expiry dates. | Audit Backlog / `Vision 3` | Teams managing venture-backed technical risk tradeoffs |
| **Multi-Tier Cognitive Projections** | Auto-generating dual views of the FAD: executive high-level summary for CTO/leadership vs deep technical specification for Staff Engineers. | Audit Backlog / `Vision 4` | Enterprise engineering leadership request for dual-audience deliverables |
| **Complexity Score Cliff Smoothing** | Refining scoring rubric transitions between Tier 2 (15 pts) and Tier 3 (16 pts) to eliminate sharp session count cliffs. | Audit Backlog / `SM-02` | Usage data across 10+ completed enterprise project pipelines |

### 2.4 Research Frontiers (Data-Gated Phase 6)

| Feature | Description | Origin | Gate Condition |
|---|---|---|---|
| **Cross-Project Knowledge Graph** | Reusable evidence ledger and technology profiles indexed across multiple independent projects. | Phase 6.2 | 5+ production projects with tracked evidence in consistent schema |
| **Adaptive Prompt Evolution** | Prompts that automatically adapt tone, search strategies, and domain probes based on previous session outcomes within a pipeline. | Phase 6.3 | Empirical evidence showing prompt underperformance patterns across 10+ pipelines |
| **Longitudinal Calibration Tracking** | Tetlock-style Brier scoring tracking whether predictions made in ADRs match production outcomes at 3, 6, and 12 months. | Phase 6.4 | 3+ projects with 6+ months post-FAD production telemetry |
| **DSPy Prompt Optimization** | Compile-time prompt optimization mathematically tuning prompt tokens for frontier models. | Phase 6.1 | Quantifiable automated research quality metric defined and validated |
| **SHA-256 Content Addressing** | Cryptographic hashing of evidence citations and session artifacts to guarantee tamper-proof audit trails for regulated industries. | Phase 6 / `NEW-07` | Enterprise compliance demand signal (SOC2 / ISO / FDA) |
| **Multi-Agent Structured Debate Protocol** | Automated multi-agent dialectic debate between opposing architectural paradigms for contentious One-Way Doors. | Phase 6.5 | Frontier models supporting asynchronous multi-agent coordination |

### 2.5 Explicitly Evaluated as Rejected (Out of Scope)

| Proposal | Why Rejected | Where Handled |
|---|---|---|
| **Tool-use calibration** | Host harness responsibility, not Vivechak's job. | Host Agent Harness (Cursor, Claude, Antigravity) |
| **Code generation / Scaffolding** | Vivechak produces evidence-grounded research; coding tools produce code. Mixing them dilutes epistemic rigor. | Downstream Coding Agents |
| **Standalone Orchestration Engine** | Replaced by standard MCP server. Eliminates duplicate LLM loops, vendor lock-in, and fragile multi-agent state engines. | Sealed ADR D-011 |
| **SQLite State Store** | Workspaces contain <50 files. Markdown IS the database. Principle P6 (Dual-Audience Artifacts) is foundational. | Principle P6 |
| **Structured JSON Tool Inputs** | Breaks Principle P6 (Dual-Audience Artifacts). YAML frontmatter + Markdown body is the standard contract for human and machine consumption. | Principle P6 |
| **Blocking Minimum Finding Counts** | Refuted by empirical gaming analysis (D-UX-02). Handled via non-blocking advisory quality coaching (`Q-DEPTH` warning) instead of hard errors. | Quality Coaching Engine |
| **Mandatory Cloud Sync** | Violates workspace-as-database principle. All state must remain local Markdown/Git. | Principle P6 |
| **HTTP Web Crawling Stack (`vivechak_fetch_source`)** | Host agent harnesses already have built-in web search and retrieval tools. Building an HTTP crawler creates unnecessary bloat. | Principle P4 |
| **Domain-Based Grade Capping & Diversity Targets** | Refuted by D-UX-02: presence metrics incentivize agents to cite weak, unverified spam to hit quotas, and breaks offline harnesses. | Anti-Gaming Rule |
| **Blocking Gates on Human Review** | Hard pauses deadlock fully autonomous agent swarms. Implemented as non-blocking advisory warnings. | Guided Worker Pattern |
| **Consolidation of 30 Files into 3 Files** | Destroys DAG modularity, granular upstream injection, and isolated git diff tracking. | Principle P1 |
| **Hardcoded Domain Archetype Heuristics** | Prescribing fixed search queries violates Principle P4. Frontier models generate domain probes dynamically. | Principle P4 |
| **4D Reversibility Tensor** | Unnecessary mathematical over-engineering; binary door classification with R/N/B/X profiling is robust and intuitive. | Simplicity / Pragmatism |
| **MCP Resources & MCP Prompts (`resources.go`/`prompts.go`)** | Redundant surface. Workspace files are accessible via standard host file tools; prompts are delivered via tool responses. | Lean MCP Surface |
| **Shadow State Files (`.vivechak-state.json`)** | Violates Principle P6 (Dual-Audience Artifacts). Creates hidden shadow state that desynchronizes from Markdown. All state is derived directly from the filesystem and artifact frontmatter. | Principle P6 & D-UX-02 |

---

## 3. Long-Term Architectural Vision

### 3.1 LLM Bridge (BYOK Intelligence Layer)

Vivechak's core guarantee is that **every tool executes deterministically without calling LLM APIs**. The MCP server is a zero-intelligence, high-reliability state and guidance engine.

In future releases, an optional **BYOK (Bring-Your-Own-Key) LLM Bridge** may be introduced:

```
┌────────────────────────────────────────────────────────┐
│ Host Agent (Cursor, Claude Code, Antigravity, etc.)    │
└──────────────────────────┬─────────────────────────────┘
                           │ MCP stdio
┌──────────────────────────▼─────────────────────────────┐
│ Vivechak MCP Server (v0.1.0 Core — Always Active)      │
│  - Workspace resolution & OS confinement (os.Root)     │
│  - DAG dependency resolution & topological sorting     │
│  - Frontmatter parsing & template validation           │
│  - Evidentiary grounding & B1–B9 Phase 0 exit gate     │
└──────────────────────────┬─────────────────────────────┘
                           │ (Optional BYOK Hook)
┌──────────────────────────▼─────────────────────────────┐
│ BYOK LLM Bridge (Optional Enrichment Layer)            │
│  - Automated red-team challenge execution              │
│  - Inline semantic conflict detection                  │
│  - Dynamic domain probe formulation                    │
│  - Synthesizer co-pilot                                │
└────────────────────────────────────────────────────────┘
```

**Key Architectural Invariants of the LLM Bridge:**
1. **Passive by default:** If no API key is configured, Vivechak operates in standard mode (generating prompts for the host agent to execute).
2. **Never gate progress:** Bridge failures or API rate limits fall back gracefully to passive mode.
3. **Local first:** Supports local OpenAI-compatible endpoints (Ollama, vLLM) alongside commercial providers.

### 3.2 Empirical Sandbox & Spike Runner

For Two-Way Door decisions that require empirical benchmarks rather than literature review:
- Ephemeral container runner (Docker / Podman) executing standardized spike scripts.
- Automatically captures CPU, memory, and latency profiles as Grade B empirical evidence.
- Directly links benchmark telemetry into session evidence ledgers.

### 3.3 Living Reversal Sentinels

Architectural Decision Records in Vivechak contain explicit **reversal triggers** and **scheduled review dates**:
- Background sentinels (GitHub Actions / scheduled CI jobs) periodically check upstream dependencies, security CVE databases, and major version releases.
- If a reversal condition is met (e.g., library deprecated, upstream pricing modified >30%), an automated issue is opened with a pre-filled `vivechak_challenge` prompt.

---

## 4. Harness Ecosystem Compatibility

Vivechak is intentionally harness-agnostic. By speaking standard MCP over `stdio`, it integrates seamlessly across all major AI runtime environments. Detailed integration instructions are documented in [docs/HARNESS-COMPATIBILITY.md](docs/HARNESS-COMPATIBILITY.md).

| Category | Environments | Integration Mode | Status |
|---|---|---|---|
| **IDE AI Harnesses** | Cursor, VS Code (Copilot/Cline), Claude Desktop, Google Antigravity, Windsurf | MCP stdio via `vck setup` (14 presets) | ✅ Production |
| **CLI Agent Harnesses** | Claude Code, `agy` CLI, OpenHands, Aider | MCP stdio or headless CLI | ✅ Production |
| **Cloud Agent Platforms** | Devin, Factory Droid, GitHub Workspaces | Pre-installed binary via `install.sh` / `vck mcp-config` | ✅ Production |
| **Parallel Agent Swarms** | Subagent spawning (Antigravity, CrewAI, LangGraph) | Multi-process concurrent access with advisory file locking | ✅ Production |
| **CI/CD Pipelines** | GitHub Actions, GitLab CI | Headless `vck doctor` and `vivechak_run_gate` | ✅ Production |

---

## 5. Architecture Principles

Vivechak is governed by three architectural pillars:

1. **Workspace As Database (P6):** There is no hidden internal database (`.vivechak/state.json`). The `research/` folder, its Markdown session files, and YAML frontmatters ARE the database. It is 100% inspectable, diffable, and version-controllable via Git.
2. **Guided Worker Pattern:** The MCP server never dictates rigid turn-by-turn state machines or throws hard errors for "wrong order." Every tool returns a constructive `next_step` guidance string to orchestrate autonomous agents smoothly.
3. **Zero-Dependency Core:** The Go implementation relies solely on the standard library and official MCP SDK. It compiles to a standalone, zero-dependency binary that launches in 10ms with zero runtime overhead.

---

## 6. Development History & Decision Archive

<details>
<summary><strong>Click to view Repository Evolution & Pre-v1.0 Architectural Decisions (OVH-01 – OVH-07)</strong></summary>

### Repository Evolution Timeline

```
2024–2025          8 independent project pipelines → pre-development patterns discovered
Mid 2026           Methodology Genesis: 11 meta-research sessions tested legacy axioms
Late 2026 (v0.1.0) Initial Unified Release: Go MCP Server (13 tools), Multi-Scope Generators, Package Distribution
Next (v0.1.x)      Host feedback, client auto-discovery, telemetry & cost tracking
Future (v1.0.0)    First LTS release — frozen MCP tool schemas, locked template contracts, knowledge graph
```

### Architectural Overhaul Decisions (OVH)

#### OVH-01: File Consolidation (38 → 27 files)
- **Rationale:** First-principles audit revealed that of 38 files, only 6 were actually touched by users. Merged PRINCIPLES.md and EVIDENCE-GRADING.md into FRAMEWORK.md. Consolidated CONTEXT.md and VISION.md into README.md. Deleted premature JSON schemas and redundant templates.

#### OVH-02: Generator Prompt Redesign
- Replaced expert personas with neutral audience framing (Zheng et al. 2024).
- Eliminated hardcoded search queries in favor of dynamic agentic ReAct loops.
- Replaced rigid section templates with quality-based coverage checklists.
- Banned multi-turn drip-feeding (Laban et al. ICLR 2026 documented 39% performance degradation).

#### OVH-03: Self-Contained Workspace Design
- Designed project research workspaces to be 100% self-contained by copying operational templates into `research/templates/`.

#### OVH-04: Self-Documenting Generator Output
- Generator outputs include execution instructions inline, eliminating external dependencies.

#### OVH-05: SaaS De-Anchoring
- Eliminated web-app SaaS bias from templates, ensuring the framework applies equally to systems programming, infrastructure, and non-software technical decisions.

#### OVH-06: Bounded Exploration Mandate (P4)
- Stated scope serves as a coverage floor, not a ceiling. Explicitly permits agents to discover critical risks and unlisted options.

#### OVH-07: Weighted Evaluation Protocol (WEP)
- Standardized multi-criteria decision analysis for all 2+ option comparisons to counteract narrative volume and verbosity biases.

### Evolution Waves Completed

- **Phase 4 (Framework Polish):** Pipeline failure modes documented, calibration loops added to P8, change propagation map instituted.
- **Phase 4.5 (Multi-Scope Expansion):** Decision-scope generator (`GENERATOR-DECISION.md`) and comparison generator (`GENERATOR-COMPARISON.md`) created; Admiralty-Code evidence grading formalized; `CORE:BEGIN/END` markers aligned across all generators.
- **Phase 5 (MCP Server Automation):** Single-binary Go MCP server with 13 tools, atomic I/O, advisory locking, and automated Phase 0 gate evaluation.

### Source Material Extraction Status

All original research notes from `brain-archive/` and historical project references (`references/project-1` through `project-8`) have been 100% extracted, empirically tested, and consolidated into `FRAMEWORK.md` and the sealed `docs/meta-research/` directory.

</details>
