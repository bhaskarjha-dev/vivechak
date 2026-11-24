# Changelog

All notable changes to **Vivechak (विवेचक)** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2026-11-24 — Initial Production Release: 9-Tool MCP Server, CLI & Multi-Scope Epistemic Engine

First production release of **Vivechak (`vck` / `vivechak`)**, delivering the settled 9-tool MCP surface (`<= 10 KB` always-loaded wire footprint), strict `yaml vivechak-plan` DAG engine, mechanical G1–G12 Phase 0 exit gate, interactive human sign-off (`vck signoff`), single-source generated documentation, and fail-closed release automation.

### Added
- **Settled 9-Tool MCP Protocol Triumvirate (`internal/mcp/`):**
  - **9 Canonical MCP Tools (`6,115 B` `tools/list`, `9,063 B` total always-loaded payload):** `vivechak_start`, `vivechak_save`, `vivechak_record_decision`, `vivechak_next`, `vivechak_status`, `vivechak_review`, `vivechak_amend`, `vivechak_replan`, and `vivechak_gate` with explicit JSON input schemas (`required`, `enum`, `additionalProperties: true` emitting `W-UNKNOWN-ARG`, zero `outputSchema`) and standardized `Envelope` responses (`{ok, code?, message, fix?, data, warnings, next, truncated?}`).
  - **4 Native MCP Prompts:** `compare` (WEP matrix), `decision` (ADR), `project` (FAD pipeline), and `gate` (Phase 0 gate audit).
  - **8 Read-Only MCP Resources:** All 6 operational templates (`vivechak://templates/{decisions,comparison,fad,conflict-resolution,gate,session}`) and 2 methodology reference cards (`vivechak://methodology/{grades,principles}`).
- **3-Scope Architecture + Out-of-Scope Guardrail:**
  - **Comparison Scope (`ScopeComparison`):** 1-session Weighted Evaluation Protocol (WEP) matrix (`research/C-NNN-comparison.md`, zero DAG, zero gate).
  - **Decision Scope (`ScopeDecision`):** 1–3 session architectural evaluation (`L`/`C`/`F`) producing canonical ADRs (`research/D-NNN-decision.md`) compiled into `research/DECISIONS.md`.
  - **Project Scope (`ScopeProject`):** Topological session DAG (`research/RESEARCH-PIPELINE.md`) with untrusted-data-wrapped upstream context injection (`<upstream-findings>`) synthesizing into `research/FAD.md`.
  - **Out-of-Scope Guardrail:** Explicit directives instructing agents to skip Vivechak for routine coding, debugging, syntax queries, and reversible local conventions.
- **Machine-Readable `yaml vivechak-plan` DAG Engine (`internal/core/dag.go`):**
  - Every project pipeline and decision plan carries one fenced `yaml vivechak-plan` block (`plan`, `sessions`, `decisions`) and one `prompt` fence per session under `## Sessions`; Markdown tables and Mermaid diagrams remain pure prose.
  - Enforces strict `V-PLAN-*` structural diagnostics (`V-PLAN-MISSING-BLOCK`, `V-PLAN-INVALID-BLOCK`, `V-PLAN-NO-SESSIONS`, `V-PLAN-MISSING-PROMPT`, `V-PLAN-DUPLICATE-PROMPT`, `V-PLAN-UNLISTED-SESSION`, `V-PLAN-UNKNOWN-DEPENDENCY`, `V-PLAN-CYCLE`, `V-PLAN-SYNTHESIS`), identifies synthesis solely via `kind: synthesis`, and preserves surrounding prose byte-for-byte on `vivechak_replan`.
- **Two-Track Phase 0 Exit Gate & Interactive Human Sign-Off (`internal/core/gate.go`, `cmd/vivechak/gate.go`, `cmd/vivechak/signoff.go`):**
  - Evaluates rules **G1–G12** across **Track A** (Two-Way Door fast reversibility) and **Track B** (One-Way Door rigor: G6 URL-anchored Grade A/B claims, G7 exact `fetched` / `human-provided` (`human`) verification labels, G8 zero unexpanded placeholders, G9 measurable reversal triggers, G10 current `D-NNN-R<k>.md` red-team challenge record matching `decision_hash`, and G11 interactive human sign-off).
  - Headless `vck gate` CLI (`--track <all|track_a|track_b>`, `--format <text|json>`, `--verbose`) reporting `N/A` on tracks where no rules ran, paired with TTY-enforced `vck signoff <D-id> --reviewer "<name>"` (`E-TTY-REQUIRED` on non-interactive stdin/stdout).
- **Single-Source Registries & Generated Reference Docs (`internal/method/evidence.json`, `internal/core/codes.go`, `internal/core/gate_rules_spec.go`, `internal/hosts/presets.go`, `tools/docgen/`):**
  - Embedded `evidence.json` drives `vivechak://methodology/grades`, `FRAMEWORK.md` §5.1–§5.3, all 3 generators' `CORE` kernels, `skills/vivechak/SKILL.md`, `templates/rules/{antigravity.md,claude.md,cursor.mdc}`, `README.md`, and `docs/QUICKSTART.md` (with `internal/embed/` twins).
  - `tools/docgen` generates `docs/reference/{tools,codes,gate}.md` and `docs/hosts.md` directly from live Go registries, backed by 1:1 source parity (`TestAllCodes_ExactSourceParity` across all `78` diagnostic codes) and live `stdio` documentation flow replay (`DocsFlows`).
- **Host Setup & Safe Config Merging (`cmd/vivechak/config.go`, `cmd/vivechak/setup.go`):**
  - `vck setup` and `vck mcp-config` support 10 host-doc-verified presets (`cursor`, `vscode`, `claude-desktop` / `claude`, `claude-code`, `antigravity` / `agy`, `windsurf`, `zed`, `roo`, `opencode`, `goose`) plus `vck setup <path>` for any other MCP host.
  - Non-destructive JSONC/JSON/YAML merging (`hujson.Standardize` + `dec.UseNumber()`), `0600` file permissions on created configs, single immutable `<config>.vivechak-bak-<timestamp>` backup, secret-safe focused `--dry-run` diffs, and accurate `PATH` vs. absolute-path launcher detection.
- **Autonomous Multi-Harness Testing Subsystem (`tests/harness/`, `scripts/run_harness.mjs`):**
  - 21 scenarios across Comparison, Decision, Project, CLI, and Out-of-Scope boundaries with process-tree watchdog execution (`execWithWatchdog`) and strict `triage.mjs` evaluation over recorded Turn 1 tool call results and sandbox file inventories (`direct_mcp` pre-flight probe completes a real MCP `initialize` handshake).
- **Multi-Channel Distribution & Fail-Closed Release Pipeline (`.goreleaser.yml`, `.github/workflows/release.yml`, `install.sh`, `install.ps1`):**
  - Cross-compiled `vivechak` and `vck` binaries across macOS (universal), Linux (`amd64`/`arm64` + `.deb`/`.rpm`/`.apk`/Arch nFPM packages), Windows (`amd64`/`arm64`), and FreeBSD (`amd64`), plus `Dockerfile`, `flake.nix`, `npm/vivechak`, `pypi/{vivechak,vck}`, Homebrew cask (`{{ .StagedPath }}` quarantine hook), and Scoop manifest.
  - Tag-only, fail-closed `release.yml` verifying publishing secrets, claimed registry endpoints, and `TAP_GITHUB_TOKEN` push permissions in `preflight` before any release job runs, building and `twine check`ing both PyPI packages prior to upload, and guarding `:latest` / npm `latest` on prerelease tags.

### Changed & Hardened
- **Per-Session Workspace & ID Isolation (`internal/workspace/resolve.go`, `internal/mcp/session_state.go`, `internal/mcp/artifact_ids.go`):**
  - Strict 7-step workspace resolution (`project_root` → client `roots/list` cached once per session when capability is advertised → `VIVECHAK_PROJECT_ROOT` / `VCK_PROJECT_ROOT` → per-session memory → `research/.vivechak.json` marker walk-up → cwd project anchor); refuses unanchored subdirectories with `E-ROOT-UNRESOLVED`.
  - Server-allocated `C-NNN` / `D-NNN` IDs with per-session placeholder resolution (`D-NEW` / `D-NNN` / `D-NNN-S1` / `D-NNN-plan` → `W-ID-ALLOCATED`), plan routing by `id` and plan block shape before inferred workspace scope, plan-seeded draft ADR titles, and zero template copies into `research/templates/`.
- **Guarded Concurrency & `os.Root` Confinement (`internal/store/`):**
  - Strict `os.Root` confinement keeping all atomic writes inside `research/` (`E-CONFINEMENT`), forward-slash path normalization on Windows, 10 MB session read cap, advisory `flock` locking, and nanosecond-stamped `research/.history/` revision archives naming existing artifact titles on `E-REVISE-REQUIRED`.

---

## [0.1.0-beta.1] - 2026-10-05 — Quality Audit Hardening & MCP Tool Expansion

Comprehensive production release fusing the 18 quality audit findings (`CRIT-01`–`CRIT-04`, `HIGH-01`–`HIGH-06`, `MED-01`–`MED-05`, `LOW-01`–`LOW-03`) with the 12 definitive architecture enhancements (T1–T12).

### Added
- **3 New MCP Tools (Tool Roster 10 → 13 Tools):**
  - `vivechak_challenge` (T1): Passive adversarial red-teaming generating targeted stress-test prompts across three operational modes (`red_team`, `evidence_audit`, `cross_session`).
  - `vivechak_replan` (T2): Safe mid-flight DAG mutation (`add_session`, `remove_session`, `update_deps`, `update_prompt`) with cycle prevention and round-trip markdown serialization.
  - `vivechak_visualize` (T3): Interactive pipeline DAG visualization generating Mermaid TD flowcharts (color-coded by execution state) and ASCII status tables.
- **ADR Supersession Protocol (T4 / `MED-05`):**
  - Added `supersedes` parameter to `vivechak_record_decision`.
  - Automatically updates superseded ADR frontmatter (`status: superseded`, `superseded_by: D-NEW`).
  - Records `supersedes: D-OLD` in the new ADR frontmatter.
  - Scans `research/sessions/` for completed sessions referencing the old decision, emitting `W-STALE-DECISION-REFERENCE` warnings.
- **`C-NNN` Comparison Namespace (T5):**
  - Added `core.SessionPrefix` mapping `ScopeComparison` to `C-` (`C-01`), `ScopeDecision` to `D-`, and `ScopeProject` to `T`.
  - `vivechak_next_session` dynamically formats comparison session IDs while preserving backward compatibility with `CMP-` and `COMP-`.
- **Generator Research Depth Mandate (T6):**
  - Added ALLIED TERMS, MULTIPLE SEARCH STRATEGIES (Direct, Adversarial, Comparative, Community), 8–12 substantive findings quota, and `[UPSTREAM_FINDINGS]` injection slot enforcement in `GENERATOR.md`.
- **Quality Coaching & Semantic Observation Engine (`ObserveSessionQuality`):** Real-time advisory feedback during `vivechak_save_session` to eliminate research superficiality and confirmation bias without blocking workflow:
  - **Canonical Provenance Scanner:** Emits `Q-EVIDENCE` when Grade A claims lack canonical URL, RFC, or DOI anchors.
  - **Citation Diversity Scanner (`Q-DENSITY`):** Emits advisory when substantive research sessions cite fewer than 2 distinct external sources.
  - **Grade A URL Verification (`EI-01`):** Requires `http://` or `https://` URLs for all `Grade A ... fetched` citations.
  - **Discovered Concerns Validator:** Emits `Q-CONCERNS` advisory if `## Discovered Concerns` section is missing.
  - **Confirmation Bias Detector (P8, `SRF-01`):** Emits `Q-BIAS` when 0 prior beliefs are contradicted, refined, or updated.
  - **Key Findings Depth:** Emits `Q-DEPTH` when a session produces fewer than 3 key findings.
  - **Rejected Alternatives Check:** Emits `Q-RIGOR` when Recommendations lack reference to rejected alternatives.
  - **Stale Date Detection:** Emits `Q-FRESHNESS` on outdated year references while ignoring false positives on port numbers, latencies, and memory sizes.
- **Documentation & Harness Ecosystem:**
  - Forward-looking `ROADMAP.md` with Deferred Features registry, Gate Conditions, and BYOK LLM Bridge vision.
  - `docs/HARNESS-COMPATIBILITY.md` detailing integration across 8 AI harness categories.
  - `docs/PARALLEL-EXECUTION.md` detailing concurrent subagent execution and lock mechanics.
  - Complete 13-tool reference in `docs/MCP-TOOLS.md`.

### Changed
- **8-Block Prompt Anatomy Harmonization (`CRIT-02`):** Synchronized prompt specification across `FRAMEWORK.md` (§4.1, §4.2, §9.2), `README.md`, `CONTRIBUTING.md`, `docs/ARCHITECTURE.md`, `docs/MCP-TOOLS.md`, and `docs/MANUAL-WORKFLOW.md`.
- **Generator Methodology Kernel Synchronization (`CRIT-04`):** Aligned `<!-- CORE:BEGIN -->` to `<!-- CORE:END -->` across `GENERATOR.md`, `GENERATOR-DECISION.md`, and `GENERATOR-COMPARISON.md`, enforced by CI test `TestCoreMarkersMatchAcrossGenerators`.
- **Centralized Research File Filtering:** Replaced fragile scattered blacklist loops with `core.IsSpecialResearchFile`, preventing non-ADR markdown files from corrupting ADR counts.
- **Guided Worker Next-Step Advice:** Dynamic `next_step` prompts advise whether the next task is a One-Way or Two-Way door with corresponding evidence standards.
- **Single-Source ADR Compilation (`tool_record_decision.go`):** Canonical `D-*.md` files are the single source of truth; `compileDecisionsRegistry` generates `research/DECISIONS.md` wrapping YAML frontmatter in collapsible `<details>` blocks.

### Fixed
- **Synthesis Findings Bleed (`CRIT-01`, `MED-04`):** In `internal/core/inject.go`, `extractFindings` preserves `## Key Findings` and inline evidence grades during synthesis mode (`isSynthesis == true`) while omitting token-heavy ledger tables, and `safeTruncateMarkdown` closes open triple backtick code fences cleanly.
- **Recalled Claim Grade C Validation (`MED-03`):** In `internal/core/validate.go`, `recalledHighGradePattern` flags Grade C recalled claims with `W-RECALLED-GRADE-CAP`.
- **Server Instructions Evidence Tiers (`HIGH-02`):** In `internal/mcp/server.go`, instructions align with `FRAMEWORK.md` §5.1 (Grade C = Vendor/Motivated, Grade D = Secondary/Opinion).
- **Comparison Scope Gate Verification (`HIGH-03`):** In `tool_run_gate.go`, comparison scope inspects issues for `W-NO-EVIDENCE-GRADES` rather than checking `WarningCount() == 0`.
- **Recommendation-Scoped Gate Evidentiary Checks (`HIGH-04`):** In `tool_run_gate.go`, B3 and B4 checks scope validation to claims supporting the recommendation, avoiding penalizing rejected alternatives.
- **Premortem Verification Hardening (`HIGH-01`):** In `tool_run_gate.go`, B7 check rejects unedited template placeholders like `[Risk 1]`.
- **Doctor Regex Collisions (`MED-01`):** In `cmd/vivechak/doctor.go`, word-boundary regexes `\b` prevent collisions between `D-01` and `D-010`.
- **Doctor Session Filename Splitting (`LOW-01`):** Splits on both hyphens and underscores.
- **Case-Insensitive ADR Resolution (`MED-02`):** `resolveDecisionFilename` matches case-insensitively and avoids prefix collisions.
- **Workspace Template Filtering (`LOW-03`):** Template counting ignores non-markdown files and validates exact scope-required templates.
- **WinGet Manifest Submission Workflow (`CRIT-03`):** Clarified post-release submission workflow in `winget/bhaskarjha-dev.Vivechak.yaml` and updated tool count.
- **Native `vck.exe` Installation (`LOW-02`):** In `install.ps1`, prefers native `vck.exe` from extracted archive over overwriting with `vivechak.exe`.
- **Windows File Renaming Concurrency:** Added monotonic atomic counter `tmpFileCounter` in `internal/store/atomic.go` to eliminate Windows temporary filename rename collisions under rapid concurrent operations.

---

## [0.1.0-beta.0] - 2026-09-28 — Initial Unified Preview (Go MCP Server & Multi-Scope Engine)

Preview (never tagged) of the automated **Vivechak (विवेचक)** system — uniting evidence-grounded research methodology with an autonomous Model Context Protocol (MCP) server.

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

[0.0.1-incubating]: https://github.com/bhaskarjha-dev/vivechak/tree/main
