# Changelog

All notable changes to **Vivechak (विवेचक)** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.1.0] - 2026-09-21 � Bias Correction and Exploration Mandate

### Added
- **Bounded Exploration Mandate (P4 enhancement):** Prescriptive scope now explicitly defines a minimum coverage floor, not a ceiling. Research agents may expand scope when evidence reveals critical concerns beyond the stated checklist � justified with evidence. Applied across all 3 pipeline layers: the generator prompt itself, the generated research pipeline, and each individual research session prompt. Addresses the documented 'Prompting Inversion' effect where overly rigid constraints on frontier models stifle native reasoning capabilities.
- **Weighted Evaluation Protocol (Framework section 6.4):** Comparison sessions (2+ competing options) now require a quantitative weighted scoring matrix alongside qualitative analysis. Protocol: derive 5-8 criteria from project context, assign justified weights, score 1-5 with evidence references, compute weighted totals, run sensitivity check. Counters narrative volume bias, verbosity bias, and vendor marketing contamination.
- **Discovered Concerns section:** Standard Prompt Template DELIVERABLE now includes a dedicated section for material concerns discovered during research beyond the original scope.

### Changed
- **FAD Template de-anchored from SaaS assumptions:** Removed hardcoded Auth/Clerk, DB/PostgreSQL, Storage/R2, Hosting/Vercel examples and Client-API-Services-DB Mermaid diagram. All sections now use neutral placeholders. 'Repository Scaffolding' renamed to 'Project Structure Specification.'
- **Generator scope broadened:** BRIEF changed from 'software project' to 'technical project,' from 'before writing application code' to 'before committing to irreversible implementation decisions.'

### Design Decisions
- **Rejected 3-Tier Universal Epistemic Engine rewrite:** External critique proposed restructuring Vivechak into Tier 0 (domain-agnostic) + Tier 1 (pluggable domain profiles) + Tier 2 (artifact generators). Rejected as premature abstraction with zero empirical validation. The 2 valid observations (FAD template SaaS bias, generator scope) were accepted as surgical fixes. Domain-agnostic expansion added to ROADMAP Phase 6 as validation-gated frontier.
- **Bounded, not unbounded exploration:** Exploration mandate says 'justify with evidence,' not 'research whatever you want.'
- **Weighted scoring complements, never replaces qualitative analysis:** Coarse 1-5 scale prevents false precision. Sensitivity check catches fragile recommendations.

---

## [1.0.0] - 2026-08-19 — Initial Public Release (Generation 3 Architecture)

### Added
- **Sovereign Framework Launch:** Initial public release of **Vivechak (विवेचक)** (*vi-* + *√vic* + *-aka* = "the discerning analyst") as the foundation intelligence layer of the Sovereign Tools Suite.
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
- **Sealed Meta-Research Provenance Base:** 14 primary deep-research sessions (639 KB), 10 ADR verdicts, and 31 empirical evidence nodes in `meta-research/`.
- **Logo & Identity:** Minimalist geometric prism logo (`docs/assets/logo.jpg`) symbolizing evidence separation.
- **Governance & Health:** Added `LICENSE` (MIT), `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `VERSION`, and GitHub issue/PR templates.

---

## [0.2.0-incubating] - 2026-08-01 — Generation 2 (URP Meta-Framework)

### Added
- Consolidated initial research methodology across 7 multi-project pipelines (Pramedha, Triyantra, Rachak, Portfolio, Gemmra, Vivah, Yugm).
- Static 3-tier pipeline topology and XML-based prompt templates.
- Empirical self-validation meta-research sessions that refuted dogmatic axioms and laid the groundwork for v1.0.0.

---

## [0.1.0-incubating] - 2024-11-15 — Generation 1 (Pre-Development Foundations)

### Added
- Initial pre-development research concepts, aspect isolation experiments, and early decision registry patterns.
