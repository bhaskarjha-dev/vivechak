# Changelog

All notable changes to **Vivechak (विवेचक)** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.1.0] - 2026-09-22 — Framework Polish & Domain Expansion

### Added
- **Bounded Exploration Mandate (P4 enhancement):** Prescriptive scope now explicitly defines a minimum coverage floor, not a ceiling. Research agents may expand scope when evidence reveals critical concerns beyond the stated checklist — justified with evidence. Applied across all 3 pipeline layers: the generator prompt itself, the generated research pipeline, and each individual research session prompt. Addresses the documented 'Prompting Inversion' effect where overly rigid constraints on frontier models stifle native reasoning capabilities.
- **Weighted Evaluation Protocol (Framework §6.4):** Comparison sessions (2+ competing options) now require a quantitative weighted scoring matrix alongside qualitative analysis. Protocol: derive 5-8 criteria from project context, assign justified weights, score 1-5 with evidence references, compute weighted totals, run sensitivity check. Counters narrative volume bias, verbosity bias, and vendor marketing contamination.
- **Discovered Concerns section:** Standard Prompt Template DELIVERABLE now includes a dedicated section for material concerns discovered during research beyond the original scope.
- **Pipeline Failure Modes (Framework §3.1):** Recovery protocols for garbage sessions, synthesis deadlocks, and Phase 0 Gate gaps — users previously had no guidance when things went wrong.
- **Session Unit Normalization (Framework §3.2):** Defines what a "session" means in practice — calibrated for deep research modes vs standard chat, with guidance that tier budgets are relative proportions.
- **Calibration Closing Loop (P8 + DECISIONS template):** Added `review_date` and `prediction` fields to ADR schema. New "Calibration Record" section in DECISIONS template for post-review feedback. Without a closing loop, decision calibration can never improve.
- **Domain Adaptation (Framework §8):** Guidance for using Vivechak on non-software domains — proven by real usage. Includes archetype maintenance model and revision cadence.
- **Rubric Calibration (Framework §3.2):** Guidance for tuning complexity scoring boundaries against observed outcomes.
- **Extract-2-files guidance (README):** Instructions for extracting generated files from web chat interfaces.

### Changed
- **FAD Template de-anchored from SaaS assumptions:** Removed hardcoded Auth/Clerk, DB/PostgreSQL, Storage/R2, Hosting/Vercel examples and Client-API-Services-DB Mermaid diagram. All sections now use neutral placeholders. 'Repository Scaffolding' renamed to 'Project Structure Specification.'
- **Generator scope broadened:** BRIEF changed from 'software project' to 'technical project,' from 'before writing application code' to 'before committing to irreversible implementation decisions.'
- **Split Generation consolidated into single-mode:** v1.0.0 offered Dual-Mode Generator (Single-Session + Split Generation for web chat output limits). v1.1.0 consolidates to single-mode with extract-2-files guidance instead — the AI produces two clearly separated documents, and the README provides extraction instructions for web chat users.
- **README tagline updated:** From "software architecture decisions" to "architecture decisions" — reflecting proven non-software domain usage.
- **Version naming normalized:** All legacy internal naming (URP, v2.0, v3.0, Generation 1/2/3) translated to Vivechak's public v1.x versioning across FRAMEWORK.md, ROADMAP.md, README.md, CHANGELOG.md, and CONTRIBUTING.md. meta-research/ preserved as sealed provenance.
- **ROADMAP rewritten:** Current State updated to v1.1 (battle-tested). Phase 4 → Framework Polish (current). Phase 5 → Vivechak Engine research phase (using Vivechak's own methodology). Phase 6 → Research Frontiers with explicit gate conditions.
- **ADR YAML schema inlined in GENERATOR.md:** Complete 18-field schema now embedded directly in the generator prompt DELIVERABLE §2, replacing the unreachable file reference to `templates/DECISIONS.template.md`. Ensures self-containment — the generating AI no longer needs filesystem access to produce correct decision records.
- **Routing matrix synced to FRAMEWORK.md:** GENERATOR.md routing matrix now matches FRAMEWORK.md exactly: 'CONFIRM & COMMIT — 1 session + ADR' and 'DEEP RESEARCH — 2–5 sessions + ADR + premortem'. Previously the generator silently dropped the premortem requirement.
- **Change Propagation Map added to CONTRIBUTING.md:** Documents all duplication points between FRAMEWORK.md, GENERATOR.md, and templates/ with mandatory check before committing.

### Design Decisions
- **Rejected 3-Tier Universal Epistemic Engine rewrite:** External critique proposed restructuring Vivechak into Tier 0 (domain-agnostic) + Tier 1 (pluggable domain profiles) + Tier 2 (artifact generators). Rejected as premature abstraction with zero empirical validation. The 2 valid observations (FAD template SaaS bias, generator scope) were accepted as surgical fixes.
- **Bounded, not unbounded exploration:** Exploration mandate says 'justify with evidence,' not 'research whatever you want.'
- **Weighted scoring complements, never replaces qualitative analysis:** Coarse 1-5 scale prevents false precision. Sensitivity check catches fragile recommendations.
- **Phase 5 requires its own Vivechak-based research:** The Vivechak Engine (automated pipeline execution) is a One-Way Door architectural decision. Building it without proper research would violate P2. Use GENERATOR.md to generate the Engine's own research pipeline first.

---

## [1.0.0] - 2026-08-19 — Initial Public Release

### Added
- **Initial Public Release:** First release of **Vivechak (विवेचक)** (*vi-* + *√vic* + *-aka* = "the discerning analyst") — an evidence-grounded pre-development research meta-framework.
- **8 Core Principles:** Empirically grounded methodology principles (Context Architecture Law, Reversibility-Calibrated Rigor, GRADE-aligned Evidence Grading, Prescriptive Scope / Dynamic Method, Commodity-Maximized Composition, Dual-Audience Artifacts, Staged Triangulation, Structured Falsification).
- **5-Block Prompt Anatomy:** Complete elimination of artificial personas and hardcoded search quotas, replacing them with structured briefs (BRIEF, SCOPE, APPROACH, DELIVERABLE, FORMAT).
- **Adaptive Complexity Scoring:** 8-dimension scoring model (0–24 points) dynamically routing projects across 4 tiers (Tier 0 to Tier 3: 1–30 sessions).
- **Two-Track Phase 0 Exit Gate:** Track A Fast-Track for reversible Two-Way Doors; Track B 9-Step Gate + Gary Klein Premortem for irreversible One-Way Doors.
- **Dual-Mode Generator:** Single-Session (IDE agents like Antigravity / Cursor) and Split Generation (web chat interfaces) in `GENERATOR.md`.
- **Self-Contained Workspace Templates:**
  - `templates/DECISIONS.template.md` (YAML frontmatter ADR schema)
  - `templates/CONFLICT-RESOLUTION.template.md` (ACH inconsistency matrix & divergence resolution)
  - `templates/FOUNDING-ARCHITECTURE.template.md` (Map-Reduce synthesis to FAD)
  - `templates/PHASE-0-GATE.template.md` (Two-track pre-codebase exit gate)
- **Sample Pipeline Walkthrough:** End-to-end walkthrough of Project "Katha" in `examples/SAMPLE-PIPELINE.md`.
- **Sealed Meta-Research Provenance Base:** 15 primary research artifacts (642 KB), 10 ADR verdicts, and 31 empirical evidence nodes in `meta-research/`.
- **Logo & Identity:** Minimalist geometric prism logo (`docs/assets/logo.jpg`) symbolizing evidence separation.
- **Governance & Health:** Added `LICENSE` (MIT), `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `VERSION`, and GitHub issue/PR templates.

---

## [0.2.0-incubating] - 2026-08-01 — Pre-Validation Baseline

### Added
- Consolidated initial research methodology across 7 independent project pipelines.
- Static 3-tier pipeline topology and XML-based prompt templates.
- Empirical self-validation meta-research sessions that refuted dogmatic axioms and laid the groundwork for v1.0.0.

---

## [0.1.0-incubating] - 2024-11-15 — Early Research

### Added
- Initial pre-development research concepts, aspect isolation experiments, and early decision registry patterns.
