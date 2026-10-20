# Changelog

All notable changes to **Vivechak (विवेचक)** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2026-10-20 — Hardened Initial Release, MCP Server & Autonomous Multi-Harness Engine

Unified release delivering the complete pre-release hardening wave across engine parsers, decoupled exit gates, headless CLI verification, concurrency locks, multi-harness autonomous testing subsystem, and packaging:

- **Autonomous Multi-Harness Testing Engine (`tests/harness/`, `scripts/run_harness.mjs`):** Built-in zero-trust autonomous stress-testing subsystem with 21 human-realistic scenarios across all three research scopes plus anti-pattern boundary testing. Includes a zero-stall process-tree watchdog runner (`execWithWatchdog` with stdin-clamping), parallel pre-flight health and token quota probes (`--probe`), dynamic semantic triage (`--triage`), and pluggable adapters (`direct_mcp`, `agy`, `opencode`, `claude_code`).
- **Phase 1 & Phase 2 Multi-Harness Empirical Audits (64 Runs Ingested):** Ingested 64 autonomous test runs across Google Antigravity CLI, OpenCode, and headless Direct MCP. Achieved 100% deterministic conformance pass across all 21 scenarios, proving zero engine regressions and 100% boundary guardrail adherence (Scenario 21 HMAC direct authoring).
- **Comprehensive Engine Hardening Remediations (D-05–D-13):** Implemented 10 core Go engine remediations discovered through autonomous multi-agent testing:
  - *DAG Phantom Session Exclusion (`D-12-01`, `dag.go`):* Excluded document titles (e.g. `# D-001-PLAN.md`) from DAG node matching.
  - *Parenthetical Qualifier Stripping (`D-12-02`, `dag.go`):* Handled dependency annotations like `| Depends | D-001-S1 (constrains) |`.
  - *Mermaid Underscore Normalization (`dag.go`):* Normalized flowchart node IDs (`T0_01` $\to$ `T0-01`) to eliminate spurious mismatch warnings.
  - *Synthesis Alias Tolerance (`validate.go`):* Accepted both `FAD` and `SYN-01` as valid synthesis targets during declared session cross-checks.
  - *Windows Path Normalization (`tool_save_plan.go`, `tool_save_session.go`):* Enforced `filepath.ToSlash` on relative paths, fixing markdown links on Windows.
  - *Structured Validation Envelope for Comparisons (`D-06-01`, `tool_save_plan.go`):* Returned explicit L1–L2 validation receipts in comparison scope.
  - *Comparison Auto-Slug Recovery (`D-05-02`, `tool_save_plan.go`):* Automatically extracted URL slugs from H1/H2 headings if omitted by models.
  - *Universal Parameter Aliasing (`tool_status.go`, `tool_visualize.go`, `tool_save_plan.go`):* Supported `root` $\leftrightarrow$ `project_root` and `id` $\leftrightarrow$ `comparison_id` / `decision_id`.
  - *Clean Uninitialized Status Semantics (`D-13-01`, `tool_status.go`):* Sanitized status output to return `initialized: false` and `status: "uninitialized"` without ambiguous error fields.
  - *Generator Next-Step Clarity (`tool_prepare_generator.go`):* Clarified that prepared prompts can be executed immediately in-place.
- **Pre-Release Forensic Hardening Wave (E-54–E-57 / D-28-01–D-28-04):** Four core remediations across DAG topology, planning validation, and local spillage prevention:
  - *Natural-Language Synthesis Dependency Expansion (`E-55` / `D-28-01`, `dag.go`):* Recognized natural language dependencies like `ALL prior sessions (hard)`, `all previous sessions`, and `*` via `isAllPriorSessions()`, dynamically expanding them in `ParsePipeline()`, `ValidateDAG()`, `AddSession()`, and `UpdateDependencies()` to all topologically preceding sessions in DAG order.
  - *Status-Aware Seed Decision Validation (`E-56` / `D-28-02`, `validate.go`):* Suppressed premature `W-NO-EVIDENCE-GRADES` warnings on candidate decision records with `status: proposed` in `research/DECISIONS.md` during initial pipeline planning.
  - *Multi-Decision Registry Validation & Chunking (`E-54` / `D-28-03`, `tool_save_plan.go`):* Replaced naive regex chunking with pure line-based state machine `extractUnanchoredDecisionBlocks` with look-ahead ID checks and backward trimming to validate each ADR stub individually with `DECISION D-NNN:` prefixes, eliminating false `V-MISSING-FIELD` blocking warnings and suppressing `W-SHORT-DECISION` skeleton noise.
  - *Hardened Tool Directives Against Brain Artifact Writes (`E-57` / `D-28-04`, `tool_prepare_generator.go`):* Hardened `next_step` guidance across `project`, `decision`, and `comparison` scopes to explicitly mandate passing generated pipeline content directly to `vivechak_save_plan` and prohibiting local IDE artifact or brain file writes.
- **Tier 3 Complex Project Execution Trace & Decision Registry Hardening (E-58–E-60 / D-29-01–D-29-03):** Three core engine and validation enhancements derived from forensic audit of a 20-session, 8-ADR, 111 KB real-world Tier 3 pipeline:
  - *Centralized Multi-Decision Registry Validation (`E-58` / `D-29-03`, `validate.go`, `tool_validate.go`, `tool_save_plan.go`):* Implemented `core.ValidateDecisionRegistry` and `core.ExtractUnanchoredDecisionBlocks` in core logic, unifying decision registry validation across `tool_validate` and `tool_save_plan`. Eliminates false `V-MISSING-FRONTMATTER` errors when validating `DECISIONS.md` containing introductory preambles, overview tables, or collapsible `<details>` blocks.
  - *Collapsible `<details>` Frontmatter Parity (`E-59` / `D-29-03`, `frontmatter.go`):* Extended `core.ParseFrontmatter` to support both ````yaml` fences and standard `---` delimited YAML frontmatter blocks inside `<details>` HTML wrappers, and combined inner/outer body content cleanly.
  - *Tier 3 Planning Guidance & Schema Expansion (`E-60` / `D-29-01`, `tool_prepare_generator.go`, `tool_validate.go`):* Clarified `nextStep` directives in `tool_prepare_generator` to explicitly permit agents to assemble large multi-session pipelines in internal scratch before passing to `vivechak_save_plan`, while strictly forbidding workspace writes outside `research/`. Added `artifact_type="decisions"` schema support and multi-decision registry auto-detection to `vivechak_validate`.
- **Calibrated Watchdog Timeout for Reasoning Models (`D-AUDIT-01`):** Increased Turn 1 timeout to 300s/360s in `orchestrator.mjs` to accommodate high-reasoning models executing 6–8 live web searches and writing 200+ line WEP matrices and ADRs.

- **Decoupled Pre-Codebase Exit Gate (`gate.go`, `tool_run_gate.go`):** Decoupled Track A (Two-Way Doors fast reversibility) from Track B (One-Way Doors B1–B11 checks). Track A explicitly passes and authorizes prototyping when all Two-Way Doors satisfy reversibility, without requiring FAD or full project DAG. Comparison scope receives clean zero-gate clearance with advisory.
- **Headless Gate CLI (`vck gate`, `cmd/vivechak/gate.go`):** Added `vck gate` subcommand with `--track <all|track_a|track_b>`, `--format <text|json>`, and `--verbose` for automated CI/CD gating and headless verification.
- **First-Class Comparison Validation (`ValidateComparison`, `validate.go`, `tool_validate.go`):** Added structured WEP matrix validation, checking criterion tables, scores, and evidence grades for comparison artifacts (`C-*.md`, `*-comparison.md`, `*-cmp-*.md`).
- **Scope Auto-Inference Precedence Calibration (`scope.go`):** Calibrated `InferScope` so explicit comparison or decision queries containing product context (e.g. *"compare pino vs winston for our new startup"*) route accurately to `ScopeComparison` or `ScopeDecision` rather than defaulting to `ScopeProject`.
- **Special Research File Isolation (`workspace.go`, `tool_record_decision.go`):** Isolated comparison files (`-cmp-`, `COMPARISON.md`) from ADR scans, preventing comparison matrices from corrupting architectural decision registries.
- **Comparison Scope Visualization (`visualize_comparison.go`, `tool_visualize.go`):** Added Mermaid flowchart and ASCII table rendering for Comparison Scope WEP matrices without requiring DAG pipeline files.
- **Multi-Session Decision ID Preservation (`tool_challenge.go`):** Replaced naive hyphen-splitting with frontmatter inspection to resolve multi-session decision IDs (`D-001-S1`, `D-001-S2`), preventing cross-session overwrite defects during challenges.
- **Content-Aware Decision Header Verification (`workspace.go`):** `InspectWorkspace` now inspects `DECISIONS.md` content for non-header decision rows, preventing empty 93-byte boilerplate tables from triggering split-brain states with gate evaluation.
- **Guided Re-planning & Reason/Impact Metadata (`tool_replan.go`):** Extended `vivechak_replan` with `reason` and `impact` fields and implemented a non-destructive guided overview mode returning uncompleted candidate sessions when `operation` is omitted.
- **Multi-Stage Local & Multi-Arch Docker Build (`Dockerfile`, `.dockerignore`):** Converted container build to a multi-stage Go builder compiling from source for both local `docker build .` and multi-arch CI (`linux/amd64`, `linux/arm64`).
- **Asynchronous Signal-Forwarding npm Runner (`npm/vivechak/bin/runner.js`):** Upgraded runner to `child_process.spawn`, forwarding `SIGINT` and `SIGTERM` signals cleanly to the native binary.
- **Safe PyPI Streaming Archive Download (`pypi/vivechak/src/vivechak/cli.py`):** Upgraded download logic to `urllib.request.urlopen` with HTTP 200 response validation.
- **Nix Flake Proxy Vendoring (`flake.nix`):** Configured `proxyVendor = true;` in `buildGoModule`.
- **Windows CRLF Normalization in Decision Registry (`tool_record_decision.go`):** Normalized `\r\n` to `\n` in `wrapFrontmatterForRegistry`, preventing dangling carriage returns on Windows.
- **4-Scope Mental Model & Out-of-Scope Guardrails:** Formalized the complete 4-scope operational hierarchy across `README.md`, `AGENTS.md`, `FRAMEWORK.md`, `docs/`, and `ServerInstructions`: Comparison Scope (1 session → WEP matrix, 0 DAG, 0 gate), Decision Scope (1–3 sessions → ADR, 0 project DAG), Project Scope (Full pipeline → FAD), and Out-of-Scope (mandatory guardrail directing agents to skip research tools for direct code authoring, bug fixing, or syntax questions).
- **Dynamic Scope-Aware Guided Worker Guidance (`tool_prepare_generator.go`):** `vivechak_prepare_generator` emits tailored, scope-appropriate `next_step` instructions across all 3 tiers (WEP comparison matrix for comparisons, decision plan and ADR skeleton for decisions, research pipeline for full projects).
- **Architectural Migration Scope Inference (`scope.go`):** Enhanced `InferScope` to recognize architectural migration queries (*"should we migrate"*, *"migrate from"*, *"migration from"*, *"architectural fork"* / *"architectural dilemma"*) and route them to `ScopeDecision`.
- **Massive Live Subprocess Multi-Scope Integration Suite (`cmd/vivechak/live_massive_multiscope_real_testing_test.go`):** Comprehensive end-to-end integration test suite compiling native binaries and communicating over stdio JSON-RPC across all 3 scopes (Comparison, Decision, Project), verifying template counts (1, 2, 6), DAG traversal, upstream context injection, FAD root mirroring, Phase 0 Gate pass, and CLI developer ergonomics.
- **AST Code Fence Truncation Defense (`markdown.go`, `inject.go`, `quality.go`, `challenge.go`):** CommonMark section extractor now tracks code block boundaries via `FenceTracker`, preventing `#` comments in code blocks from prematurely truncating discovered concerns or recommendation text.
- **Dual-Source Decision Split-Brain Fix (`workspace.go`, `gate.go`, `tool_validate.go`):** Workspace resolution, gate verification, and validation now scan both standalone `D-*.md` files and `research/DECISIONS.md` concurrently with ID deduplication.
- **Release Packaging Automation (`release.yml`):** Release workflow dynamically injects git tag versions into npm and PyPI package manifests, dependency bounds, and binary runners before publishing.
- **Evidence Ledger Provenance Cross-Referencing (`quality.go`):** Grade A citations are verified against URLs and RFCs documented in the Evidence Ledger table without raising false alarms on finding bullets, and table rows are excluded from claim counts.
- **Scope Auto-Inference & Shorthand (`scope.go`):** Added native recognition for `vck` and `vivechak` command prefixes (`vck plan`, `vck decide`, `vck compare`, `vck vs`), comparative signal precedence, and human slang/typo resilience (`shud`, `bw`, etc.).

- **Resilient MCP Schema Aliasing & Cross-Harness Interoperability:** Transparent parameter aliasing across diverse AI agent harnesses:
  - `vivechak_challenge` accepts `decision_id` (alias for `session_id`) and `challenge_type` (alias for `mode`), preventing runtime schema rejections from small-model or third-party agent harnesses.
  - `vivechak_run_gate` accepts optional `track` parameter (`track_a`, `track_b`, or `all`), returning structured machine-readable outcome fields `track_a_result` (`"PASS"` / `"FAIL"`), `track_b_result` (`"PASS"` / `"FAIL"`), and `decisions` array.
- **Workspace Scope Diagnostics (`doctor.go`):** `vck doctor` now explicitly detects and displays the active workspace research scope (`✓ Scope: <project|decision|comparison>`), confirming pipeline tier recognition.
- **Decoupled Plan-File Workspace Discovery (`workspace.go`):** Pipeline plan file detection (`RESEARCH-PIPELINE.md`, `DECISION-PLAN.md`, `COMPARISON-PLAN.md`) is fully decoupled from `.vivechak.json` presence, ensuring `info.HasPipeline = true` is consistently reported across all research scopes.
- **Massive In-Memory Real-Lifecycle Test Suite (`massive_real_lifecycle_test.go`):** Added 8-scenario comprehensive integration test suite verifying end-to-end multi-scope execution, stress concurrency, vague human prompts, resilient schema aliases, dynamic quality coaching, and Phase 0 exit gates.
- **Universal Multi-Ecosystem Packaging & Zero-Script Ephemeral Runners:** Expanded distribution architecture with dual-identity execution parity (`vivechak` and `vck`):
  - **Linux Native Packages via nFPM (`.goreleaser.yml`):** Automatically builds `.deb` (Debian/Ubuntu/Mint), `.rpm` (Fedora/RHEL/CentOS), `.apk` (Alpine), and `.pkg.tar.zst` (Arch Linux) bundling both binaries with `Provides: vck`.
  - **Zero-Script npm / npx Runner (`npm/vivechak`, `npm/vck`):** Pure Node.js lazy-cache binary runner (`bin/runner.js`) with zero postinstall scripts (bypassing pnpm/Yarn/npm v12 script blocking and `--ignore-scripts`) allowing instant execution via `npx -y vck serve` with zero local installation.
  - **PyPI / uvx Python Runner (`pypi/vivechak`, `pypi/vck`):** Pure Python standard-library wrapper with in-memory process replacement (`os.execv`) allowing instant execution via `uvx vck serve`.
  - **Docker Multi-Arch Container (`Dockerfile` & GHCR):** Multi-arch container image (`linux/amd64`, `linux/arm64`) on GitHub Container Registry (`ghcr.io/bhaskarjha-dev/vivechak` and `ghcr.io/bhaskarjha-dev/vck`) with non-root security and clean stdio MCP streaming.
  - **FreeBSD Target:** Added `freebsd/amd64` to GoReleaser matrix.
  - **Nix Flake (`flake.nix`):** Declarative package expression enabling `nix run github:bhaskarjha-dev/vivechak`.
  - **Dual Homebrew Tap & Scoop Bucket Manifests:** Added `Formula/vck.rb` (with `conflicts_with "vivechak"`) and `scoop/vck.json`.
  - **Decoupled CI/CD Workflow (`release.yml`):** Upgraded release action to a decoupled 4-stage pipeline isolating core releases from optional token-dependent jobs (`docker`, `npm`, `pypi`).
  - **GoReleaser Tap Resilience (`.goreleaser.yml`):** Implemented conditional `skip_upload` for Homebrew and Scoop targets, preventing release failure when `TAP_GITHUB_TOKEN` is unset.
  - **OCI Image Specification Compliance (`release.yml`):** Enforced lowercase repository owner conversion in Docker Buildx to comply with strict OCI registry naming rules.
  - **Non-Interactive PyPI Release (`release.yml`):** Configured `--non-interactive` flag in Twine uploads to prevent credential prompt hanging in CI.
  - **Comprehensive Package Documentation (`npm/*/README.md`, `pypi/*/README.md`):** Added complete setup guides, host presets, and CLI usage to standalone npm and PyPI packages.
- **Expanded AI Harness Presets & Portable Standard (`config.go`):** Expanded `vck setup` and `vck mcp-config` from 14 to 21 presets, adding native support for Claude Code (`claude-code`), Universal Portable MCP (`mcp` / `portable` / `universal` via `.mcp.json`), OpenCode (`opencode`), Continue.dev (`continue`), Block Goose (`goose`), Void Editor (`void`), and GitHub Copilot CLI (`copilot`).
- **Zero-Loss State Preservation for Claude Code (`config.go`):** Implemented number-preserving JSON decoding (`dec.UseNumber()`) in configuration merges, ensuring existing user state, tokens, session history, and custom settings in `~/.claude.json` are never clobbered or truncated.
- **Specialized Schema Dispatch (`config.go`):** Added schema generators for OpenCode (array-format `command: ["<exe>", "serve"]`, `"type": "local"`, `"enabled": true` under `"mcp"` key), Continue.dev modular configuration (`.continue/mcpServers/vivechak.json`), and Block Goose stdio extensions (`config.yaml`).
- **Workspace Host Auto-Detection (`config.go`):** Extended workspace auto-detection to recognize project-root markers `.mcp.json`, `opencode.json`, `.void`, `.continue`, `.goose`, and `.copilot`.
- **Bidirectional Frontmatter Status Lifecycle Sync (`frontmatter.go`, `tool_save_session.go`):** Added `SetFrontmatterField` supporting `<details>` blocks, code fences, CRLF line endings, and comments; enforces `status: draft` upon blocking validation issues and automatically promotes valid sessions to `status: complete` on disk, resolving DAG unblocking deadlocks.
- **Auxiliary Tool `knownIDs` DAG Propagation (`workspace.go`, `mcp`):** Propagated `knownIDs` from DAG session definitions to `scanCompletedSessions` across `tool_save_session.go`, `tool_visualize.go`, `tool_replan.go`, and `tool_record_decision.go`, ensuring slugged session filenames without explicit frontmatter IDs are recognized as completed.
- **Dynamic 5-Column Evidence Ledger & Markdown Link Parsing (`validate_evidence.go`):** Implemented dynamic table column header matching (`Source`, `Citation`, `Reference`, `Provenance`) and row-wide URL scanning in `ValidateEvidenceProvenance`, supporting markdown link syntax `[Title](url)` and eliminating false `V-OWD-VAGUE-PROVENANCE` errors on standard ledgers.
- **Adversarial Parameter Traversal Validation (`tool_amend_session.go`, `tool_record_decision.go`):** Added strict `isValidID` format validation to secondary parameters (`amending_session_id`, `auto_draft_from`) to prevent directory traversal and injection attacks.
- **Comprehensive Multi-Phase Live MCP Verification Suite (`live_massive_perfect_verification_test.go`):** Implemented end-to-end child process test suite over stdio JSON-RPC validating multi-branch DAG progression, status synchronization, in-flight replanning, ADR auto-drafting and supersession, root FAD mirroring, Track A & B Phase 0 Gate, security defenses, and CLI diagnostics (`doctor`, `setup`, `mcp-config`, `version`).
- **Guided Worker & DAG Hardening (E-16–E-21):** Synthesis exit-gate routing (`tool_save_session.go`), L2Block draft dependency gating in DAG resolution (`workspace.go`, `tool_next_session.go`), pre-synthesis evidentiary sanity checks (`tool_next_session.go`), actionable ADR auto-draft guidance, upstream context complacency guard, and pinpoint Phase 0 Gate remediation hints with `help:` links (`gate.go`).
- **Unified AST-Aware Section Extractor (`markdown.go`):** CommonMark heading-level extraction (`ExtractSection`) across `validate.go`, `draft_decision.go`, and `tool_run_gate.go`; preserves H3 subsections within H2 sections and eliminates sibling heading truncation.
- **Strict Track B Phase 0 Gate (`tool_run_gate.go`):** Promotes One-Way Door `B3` (evidentiary threshold), `B4` (verification method), and `B5` (rejected alternatives) defects from non-blocking advisories to fatal blocking issues, preventing Potemkin passes.
- **Canonical Provenance Anchor Verification (`validate.go`):** Verifies Grade A claims against canonical URL (`http://`, `https://`), RFC (`rfc `), or DOI (`doi:`) anchors; eliminated arbitrary lexical `>70% Grade A` bell curve cap per ADR `D-UX-02` and systematic review gold standards.
- **Structural ADR Shape Gate (`tool_run_gate.go`):** Validates the structural presence of Context, Rationale, and Reversal Triggers instead of naive 200-character byte counting.
- **Anti-Goodhart Belief Delta Quality Coaching (`validate.go`):** Identifies performative deltas that start exclusively from "Unknown" to evade `Q-BIAS`, prompting validation against Principle P8 (Structured Falsification).
- **Evidence Citation Diversity Quality Coaching (`validate.go`):** Advisory `Q-DENSITY` coaching ensuring substantive research sessions cite $\ge 2$ distinct independent sources.
- **Eager Advisory Mutex Locking in `replan` (`tool_replan.go`):** Acquires `pipeline` lock before reading or parsing `RESEARCH-PIPELINE.md`, eliminating TOCTOU race conditions during concurrent multi-agent replans.
- **Advisory Mutex on Root FAD Mirror Write (`tool_save_session.go`):** Locks `fad_root` before atomic copy of `FOUNDING-ARCHITECTURE.md` to workspace root.
- **Dynamic Guided Worker Routing (`tool_save_session.go`):** Real-time `[Q-*]` quality advisories surfaced directly in `next_step` instructions; one-way door sessions imperatively route to `vivechak_challenge(mode="red_team")` before recording irreversible decisions.
- **Research Calibration Prompt Injection (`tool_next_session.go`):** Automated active rule injection (Memory ≠ Evidence, Qualified Provenance, Falsification Required, Honest Gaps) on all non-synthesis session prompts returned by `vivechak_next_session`.
- **Qualified Provenance Engine (`validate.go`):** Strict verification requirement for Grade A fetched claims against canonical URLs, domain anchors, RFCs, or doc titles (L2Block on one-way doors, L3Warn on two-way doors).
- **Substantive Discovered Concerns Gate (`validate.go`):** Strict length and non-triviality check (≥50 characters) on one-way doors (`V-OWD-NO-CONCERNS`) while permitting fast two-way door spikes.
- **ADR Substantive Content Gate B10 (`tool_run_gate.go`):** Requires ≥200 characters and explicit evaluated/rejected alternatives for One-Way Door decisions before Phase 0 Gate exit.
- **Embedded Contract Parity (`SESSION.template.md`):** Added `## Discovered Concerns` section with 100% embed sync to ensure template users pass validation out of the box.
- **Topology vs Session Parity Guard (`dag.go`, `validate.go`):** Detects declared sessions in pipeline topologies or Mermaid diagrams that lack corresponding prompt blocks, emitting `W-PARSER-SESSION-MISMATCH`.
- **Synthesis Terminal Sink Guard (`dag.go`):** Enforces in `NextSessions` that synthesis sessions (`SYN-*`, `FAD`) are strictly held back until all non-synthesis research sessions are complete.
- **In-Process Workspace Session Memory (`workspace.go`):** Caches the most recently resolved workspace path during an active stdio connection, eliminating tool failure loops when desktop IDEs omit `project_root`.

### Changed
- **Token-Optimized Generator Instructions (`GENERATOR.md`):** Updated execution guidance to instruct MCP-enabled agents to receive briefs directly from `vivechak_next_session` without redundant manual `view_file` calls to static templates, saving ~60k tokens and 30 tool calls per pipeline run.
- **Embed Parity Synchronization (`internal/embed/generators/`):** 100% byte-for-byte synchronization across embedded and root generators.
- **Decoupled Phase 0 Gate (`tool_run_gate.go`):** Fixed Track A/Track B semantic inversion; Track A strictly evaluates reversibility and blast radius bounds for Two-Way Doors, while Track B executes rigorous B1–B11 checks on One-Way Doors (including B11 adversarial red-team challenge).
- **Active Track Filtering in Phase 0 Gate (`tool_run_gate.go`):** Added active filtering and validation for `track` parameter (`track_a`, `track_b`, `all`), isolating blocking issues and tailoring failure messages to the evaluated track.
- **Comparison Scope Terminal NextStep (`tool_save_plan.go`, `tool_status.go`):** Emits terminal completion next_step upon WEP matrix save and status checks, eliminating infinite DAG session loops and false Phase 0 Gate referrals for 0-gate Comparison Scope.
- **Scope-Aware Comparison Inspection (`workspace.go`, `gate.go`):** `InspectWorkspace` detects and counts standalone comparison files (`research/C-*.md`) when `sessions/` directory is absent; `EvaluateGate` uses direct `EvidenceGradePattern` regex matching across session and research comparison files.
- **Stdio Protocol Stream Isolation (`runner.js`, `cli.py`):** Configured `childStdio = ['ignore', 2, 2]` in npm runner and `stdout=sys.stderr` in PyPI runner during child archive extraction and local binary build, preserving fd 1 exclusively for MCP JSON-RPC protocol transport.
- **Documented 9-Section Session Body Skeleton (`FRAMEWORK.md` §6.2):** Aligned methodology specification with runtime templates and validation engine, formally standardizing `Prior`, `Discovered Concerns`, and `Delta` sections.
- **ADR Door-Type Drafter (`draft_decision.go`):** Contextually infers `two-way` vs `one-way` door classifications from session text rather than defaulting to irreversible.
- **Unambiguous Phase 0 Gate Framing (`tool_run_gate.go`):** Replaced polite completion phrasing with explicit failure stop-signals (`❌ PHASE 0 GATE FAILED`, `⛔ DO NOT PROCEED TO CODE`) when Track B requirements are not met.
- **One-Way Door Challenge Mandate (`GENERATOR.md`, `DECISIONS.template.md`):** Removed all "(optional)" qualifiers from `vivechak_challenge` for irreversible decisions.

### Fixed
- **Auto-Initialization Advisory Lock Guard (`tool_save_plan.go`):** Auto-initialized workspaces created on the fly automatically write `research/.gitignore` containing `*.lock` when `vivechak_init` was skipped.
- **Automatic ADR Amendment Registry Synchronization (`tool_amend_session.go`):** Amending a standalone ADR (`research/D-*.md`) via `vivechak_amend_session` automatically recompiles `research/DECISIONS.md`.
- **Colon-Separated Evidence Grade Notation (`validate_evidence.go`):** Expanded `EvidenceGradePattern` to support colon notation (e.g. `[Grade: A]`, `Grade: B`), eliminating false evidence warnings.
- **Decoupled Track A Two-Way Door Gate Render (`gate_render.go`):** Decoupled Two-Way Door Track A result from structural FAD file presence, ensuring passing Two-Way Doors render `PASS` even before final FAD synthesis.
- **RE2 Stale Date Regex Parsing (`quality.go`):** Replaced Perl-style lookahead `(?!)` with RE2-compliant character class negation `[^\s\S]`, resolving staticcheck `SA1000` lint failures.
- **Leading Markdown Title Frontmatter Crash (`frontmatter.go`):** `skipLeadingWhitespaceAndComments` now skips markdown titles preceding frontmatter delimiters, fixing `V-MISSING-FRONTMATTER` crashes.
- **ADR Supersession Duplicate Frontmatter (`tool_record_decision.go`):** Dynamically isolates existing YAML frontmatter boundaries, preventing duplicate frontmatter blocks on superseded ADRs.
- **ASCII Arrow Tolerance in DAG Complexity Parser (`dag.go`):** Regex accepts `->`, `-->`, and `=>` in addition to Unicode `→`, preventing silent complexity score zeroing.
- **Standalone Uppercase Session ID Omission (`dag.go`):** Supported standalone uppercase identifiers such as `### FAD:` and `### SYNTHESIS:`.
- **Gate Recommendation Fallback Bibliography Bleed (`tool_run_gate.go`):** Strips `## Sources`, `## References`, and `## Evidence Ledger` from recommendation fallback text, preventing secondary bibliography links from triggering false evidentiary rejections.
- **Workspace OS Boundary Guard Collision (`workspace.go`):** Implemented `isDirOrChild` path boundary matching, eliminating false-positive rejections on Windows paths (`D:\windows-tools`, `C:\binary-tree`) and Unix project folders (`/library-books`, `/bin-project`).
- **Replanning Data Loss Prevention (`dag.go`):** Retained raw `Preamble` (parameters, complexity scoring, execution notes) and `ExitGate` sections across round-trip serialization in `vivechak_replan`.
- **Prompt Truncation on Inner Code Fences (`dag.go`):** Enhanced code fence parser to track outer fence length, preventing premature truncation when encountering untagged inner code fences.
- **Computing Memory False-Positive Penalty (`validate.go`):** Narrowed `recalledHighGradePattern` to model/parametric recall, preventing valid computing memory claims (RAM, caching) from being downgraded.
- **Split-Brain FAD Desynchronization (`tool_amend_session.go`):** Post-hoc amendments to synthesis sessions atomically mirror changes to root `FOUNDING-ARCHITECTURE.md`.
- **Scope Auto-Detection on Empty Project Root (`tool_prepare_generator.go`):** Resolves workspace from environment variables or working directory when `project_root` argument is omitted.
- **Standalone ADR Validation Fallback (`tool_validate.go`):** Correctly parses registered decisions in `research/DECISIONS.md` when standalone `D-*.md` files are not created.

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

[0.1.0]: https://github.com/bhaskarjha-dev/vivechak/releases/tag/v0.1.0
[0.1.0-beta.1]: https://github.com/bhaskarjha-dev/vivechak/releases/tag/v0.1.0-beta.1
[0.1.0-beta.0]: https://github.com/bhaskarjha-dev/vivechak/releases/tag/v0.1.0-beta.0
[0.1.0-alpha.2]: https://github.com/bhaskarjha-dev/vivechak/releases/tag/v0.1.0-alpha.2
[0.1.0-alpha.1]: https://github.com/bhaskarjha-dev/vivechak/releases/tag/v0.1.0-alpha.1
[0.0.1-incubating]: https://github.com/bhaskarjha-dev/vivechak/tree/main
