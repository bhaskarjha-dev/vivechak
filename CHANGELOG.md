# Changelog

All notable changes to **Vivechak (विवेचक)** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

Improvements addressing forensic audits (`temp/audit/`), implementing active epistemic coaching, programmatic Phase 0 Gate verification, single-source ADR compilation, and Windows concurrency safety.

### Added
- **Quality Coaching & Semantic Observation Engine (`ObserveSessionQuality`):** Real-time advisory feedback during `vivechak_save_session` to eliminate research superficiality and confirmation bias without blocking workflow:
  - **Grade Inflation Scanner:** Emits `Q-EVIDENCE` when Grade A citations exceed 70% of claims, coaching realistic web distributions (~25% A, ~50% B, ~25% C).
  - **Grade A URL Verification (`EI-01`):** Requires `http://` or `https://` URLs for all `Grade A ... fetched` citations.
  - **Discovered Concerns Validator:** Emits `Q-CONCERNS` advisory if `## Discovered Concerns` section is missing.
  - **Confirmation Bias Detector (P8, `SRF-01`):** Inspects the Delta table and emits `Q-BIAS` when 0 prior beliefs are contradicted, refined, or updated.
  - **Key Findings Depth:** Emits `Q-DEPTH` when a session produces fewer than 3 key findings.
  - **Rejected Alternatives Check:** Emits `Q-RIGOR` when Recommendations lack reference to rejected alternatives.
  - **Stale Date Detection:** Emits `Q-FRESHNESS` on outdated year references while ignoring false positives on port numbers, latencies, and memory sizes.
- **Automated Phase 0 Gate Evidentiary Engine (`tool_run_gate.go`):**
  - Programmatic verification of **B3** (Evidentiary Threshold: scans one-way door sessions for low grades C/D/E), **B4** (Verification Integrity: detects ungrounded recalled claims), **B5** (Rejected Alternatives: requires $\ge 30$ chars of documented rejected options), and **B6** (Reversal Triggers: enforces measurable review conditions).
  - Automated persistence of formatted exit gate checklist to `research/PHASE-0-GATE.md` with timestamps and verdict routing.
- **Single-Source ADR Compilation (`tool_record_decision.go`):**
  - Individual `D-*.md` files are the canonical single source of truth; `compileDecisionsRegistry` automatically generates `research/DECISIONS.md` wrapping YAML frontmatter in collapsible `<details>` blocks.
  - Validation skips `DECISIONS.md` to prevent double-counting ADRs.
- **Real-Time Mini-Status Progress Embedding:**
  - `vivechak_save_session` and `vivechak_record_decision` return a `progress` object (`sessions_completed`, `sessions_total`, `decisions_recorded`), eliminating the need for redundant `vivechak_status` polling.
- **Transitive Downstream Stale Cascades (`tool_amend_session.go`):**
  - BFS traversal of DAG dependencies identifies completed downstream sessions and emits `W-STALE-DOWNSTREAM` warnings with `potentially_stale_sessions`.
- **Synthesis Single-Write & Root Copy Mirroring:**
  - Saving synthesis (`SYN-01`) writes directly to canonical `research/FAD.md` and mirrors byte-for-byte to `FOUNDING-ARCHITECTURE.md` at project root.
- **6th Embedded Operational Contract:**
  - Standardized `templates/SESSION.template.md` embedded into binary (`internal/embed/templates/SESSION.template.md`) and scaffolded during `vivechak_init`.
- **Expanded Server Instructions:**
  - Expanded `ServerInstructions` to ~1,500 bytes with canonical 4-step workflow, grading definitions, and token economics guidelines.

### Changed
- **Centralized Research File Filtering:** Replaced fragile scattered blacklist loops with `core.IsSpecialResearchFile`, preventing non-ADR markdown files (`*-plan.md`, `PHASE-0-GATE.md`, `NOTES.md`) from corrupting ADR counts.
- **Guided Worker Next-Step Advice:** Dynamic `next_step` prompts advise whether the next task is a One-Way or Two-Way door with corresponding evidence standards.

### Fixed
- **High-Concurrency Windows File Locking:** Added monotonic atomic counter `tmpFileCounter` in `internal/store/atomic.go` to eliminate Windows temporary filename rename collisions under rapid concurrent operations.
- **Windows File Sharing Violations:** Serialized decision writing and registry compilation under `core.DecisionsFile` lock with exponential backoff retry loops.

---

## [0.1.0] - 2026-09-28 — Initial Unified Release (Go MCP Server & Multi-Scope Engine)

First public release of the automated **Vivechak (विवेचक)** system — uniting evidence-grounded research methodology with an autonomous Model Context Protocol (MCP) server.

### Added
- **Go MCP Server (`cmd/vivechak/`):** Autonomous orchestration engine implementing the Guided Worker pattern with 10 atomic tools:
  - `vivechak_init`: Initializes self-contained project workspaces with required directories and operational contracts.
  - `vivechak_prepare_generator`: Inlines project briefs into multi-scope generator prompts without host LLM file leaks.
  - `vivechak_save_plan`: Parses and validates execution DAGs with cycle detection and parallel execution tracking.
  - `vivechak_status`: Real-time workspace inspection, blocking dependencies, and Phase 0 Gate readiness.
  - `vivechak_next_session`: Dependency-aware next session retrieval with automated upstream context injection.
  - `vivechak_save_session`: Atomic persistence and validation of session research outputs.
  - `vivechak_record_decision`: Architectural Decision Record (ADR) validation and registry updates.
  - `vivechak_amend_session`: Post-hoc session and synthesis amendments preserving historical audit trail.
  - `vivechak_validate`: Pre-flight YAML frontmatter and evidentiary grading validation ladder.
  - `vivechak_run_gate`: Two-Track Phase 0 Exit Gate evaluation and premortem enforcement.
- **Multi-Scope Research Methodology:**
  - **Project-Scope (`GENERATOR.md`):** Full end-to-end research DAG decomposing technical visions into structured research pipelines synthesizing into a Founding Architecture Document (FAD).
  - **Decision-Scope (`GENERATOR-DECISION.md`):** Targeted 1–3 session routing (Landscape, Comparison, Falsification) producing standalone Architectural Decision Records.
  - **Comparison-Scope (`GENERATOR-COMPARISON.md`):** Single-session Weighted Evaluation Protocol (WEP) comparison matrix.
- **Embedded Operational Contracts (`templates/` and `internal/embed/`):**
  - `DECISIONS.template.md`: YAML frontmatter ADR schema with confidence, evidence references, and review triggers.
  - `CONFLICT-RESOLUTION.template.md`: Transposed ACH-style inconsistency matrix for divergence resolution.
  - `COMPARISON-SESSION.template.md`: Standardized WEP comparison session output structure.
  - `FOUNDING-ARCHITECTURE.template.md`: Map-Reduce synthesis template for technical architectures.
  - `PHASE-0-GATE.template.md`: Two-track pre-codebase exit gate with Klein premortem protocol.
  - `SESSION.template.md`: Standard research session output format with YAML frontmatter, evidence ledger, and delta tracking.
- **CLI Subcommands & Shorthand Alias:**
  - `vck` official 3-letter shorthand CLI binary alias installed alongside `vivechak` across Homebrew, Scoop, WinGet, and shell installers (100% collision-free across all OS and package ecosystems).
  - `vck setup` (aliased with `install`) subcommand enabling single-command host setup (`vck setup cursor`, `vck setup claude`, `vck setup vscode`), workspace auto-detection (`vck setup` with zero arguments), direct custom config file targeting, and `--dry-run` (`-n`) preview mode.
  - `vck mcp-config` universal MCP JSON configuration generator with auto-write presets for 14 desktop hosts and agent harnesses (Cursor, VS Code, Claude Desktop, Antigravity, Windsurf, Zed, AWS Kiro, Trae, OMP, OpenHands, Factory Droid, Cline, Roo Code, and Devin).
  - `vck doctor` for workspace contract validation and dependency diagnostics.
- **Cross-Platform Distribution:** GoReleaser matrix across 6 OS/architecture targets, standalone shell installers (`install.sh`, `install.ps1`), and package manager manifests (Homebrew, Scoop, WinGet).
- **Sealed Meta-Research Evidence Base:** 15 foundational research artifacts, 14 locked ADRs (D-001–D-014), and 31 empirical evidence nodes proving the methodology through empirical self-application.
- **Comprehensive Test Suite:** 47+ unit, wire-level in-memory tests, and live subprocess stdio end-to-end integration tests covering DAG parsing, frontmatter validation, atomic file storage, advisory locking, and exit gates.

### Changed
- **Dependency Parser Hardening:** Fixed prefix matching in `parseDependencies` so parenthetical annotations (e.g. `(none-blocking)`) do not inadvertently drop valid dependency lists.
- **Frontmatter Diagnostics:** `ParseFrontmatter` returns an explicit error when opening `---` has no closing delimiter, enabling precise `V-INVALID-FRONTMATTER` diagnosis.
- **Guided Worker & API Ergonomics:** `vivechak_save_session` response data includes a `"validation_passed"` boolean; `vivechak_prepare_generator` auto-detects workspace scope from metadata when omitted.
- **Context Injection Safeguards:** Introduced `W-INJECTION-SIZE` warning across `vivechak_next_session` when aggregate injected upstream context exceeds 100 KB.
- **CLI & Test Fixtures:** Hoisted regex compilation in `doctor`, added `--help` flag handling to `mcp-config`, and switched test pipeline fixtures to table format.
- **CI & Linter Modernization:** Migrated `.golangci.yml` to the v2 configuration schema, upgraded CI workflow to `golangci-lint-action@v9` with `v2.14.0`, and canonicalized macOS symlinked temporary paths in workspace resolution tests.

---

## [0.1.0-alpha.2] - 2026-09-21 — Methodology Refinements & Exploration Mandate

Pre-release methodology refinement cycle (formerly drafted as v1.1.0):

### Added
- **Bounded Exploration Mandate (P4 enhancement):** Prescriptive scope now explicitly defines a minimum coverage floor, not a ceiling. Research agents may expand scope when evidence reveals critical concerns beyond the stated checklist — justified with evidence.
- **Weighted Evaluation Protocol (Framework §6.4):** Comparison sessions (2+ competing options) require a quantitative weighted scoring matrix alongside qualitative analysis to counter LLM verbosity and marketing bias.
- **Discovered Concerns section:** Standard Prompt Template DELIVERABLE includes a dedicated section for material concerns discovered during research beyond the original scope.
- **Pipeline Failure Modes (Framework §3.1):** Recovery protocols for low-quality sessions, synthesis deadlocks, and Phase 0 Gate gaps.
- **Session Unit Normalization (Framework §3.2):** Calibrated definition of "session" for deep research modes vs standard chat.
- **Calibration Closing Loop (P8 + DECISIONS template):** Added `review_date` and `prediction` fields to ADR schema for post-review feedback.

### Changed
- **FAD Template de-anchored from SaaS assumptions:** Removed hardcoded Auth/Clerk, DB/PostgreSQL, Storage/R2, Hosting/Vercel examples. All sections use neutral placeholders.
- **Generator scope broadened:** BRIEF generalized from software projects to all technical architecture decisions.
- **ADR YAML schema inlined in GENERATOR.md:** Complete 18-field schema embedded directly in prompt DELIVERABLE block.

---

## [0.1.0-alpha.1] - 2026-08-19 — Methodology Genesis & Core Principles

Foundational meta-research and methodology specification (formerly drafted as v1.0.0):

### Added
- **Core Methodology Framework:** Formalization of the 8 Core Principles (Context Architecture Law, Reversibility-Calibrated Rigor, GRADE-aligned Evidence Grading, Prescriptive Scope / Dynamic Method, Commodity-Maximized Composition, Dual-Audience Artifacts, Staged Triangulation, Structured Falsification).
- **5-Block Prompt Anatomy:** Elimination of artificial personas and hardcoded search quotas, replacing them with structured briefs (BRIEF, SCOPE, APPROACH, DELIVERABLE, FORMAT).
- **Adaptive Complexity Scoring:** 8-dimension scoring model (0–24 points) dynamically routing projects across 4 tiers.
- **Two-Track Phase 0 Exit Gate:** Track A Fast-Track for reversible Two-Way Doors; Track B 9-Step Gate + Gary Klein Premortem for irreversible One-Way Doors.
- **Foundational Operational Templates:** Initial versions of DECISIONS, CONFLICT-RESOLUTION, FOUNDING-ARCHITECTURE, and PHASE-0-GATE templates.

---

## [0.0.1-incubating] - 2024-2025 — Early Research Baseline

### Added
- Initial pre-development research patterns and aspect isolation experiments discovered across 8 independent project pipelines.
