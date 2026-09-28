# Vivechak Roadmap & Development History
### Evolution, Decisions, and Future Direction

---

## Current State: v0.1.0 (September 2026)

**Status:** Operational, battle-tested, and automated via Model Context Protocol (MCP).

Vivechak v0.1.0 marks the initial unified release combining the evidence-grounded research methodology with an autonomous Go MCP server and multi-scope pipeline generators.

### What v0.1.0 Delivers

| Component | State | Description |
|---|---|---|
| **Multi-Scope Generators** | ✅ Ship-ready | `GENERATOR.md` (Project), `GENERATOR-DECISION.md` (Decision), `GENERATOR-COMPARISON.md` (Comparison) |
| **Go MCP Server** | ✅ Shipped | 9 atomic tools implementing the Guided Worker pattern with 32/32 tests passing |
| **CLI & Diagnostics** | ✅ Shipped | `vivechak mcp-config` (7 AI hosts) and `vivechak doctor` workspace integrity checker |
| **FRAMEWORK.md** | ✅ Consolidated | Complete standalone spec (principles + evidence grading + multi-scope methodology) |
| **5 Templates** | ✅ Agent-ready | Decisions, Conflict Resolution, Comparison Session, FAD, Phase 0 Gate (`schema_version: "0.1.0"`) |
| **Package Distribution** | ✅ Shipped | GoReleaser (6 platforms), Homebrew tap, Scoop bucket, WinGet manifest, shell installers |
| **meta-research/** | ✅ Sealed | 14 locked ADRs (D-001–D-014), 27 research artifacts, empirical evidence base |

### Repository Evolution

```
2024–2025          8 independent project pipelines → pre-development patterns discovered
Mid 2026           Methodology Genesis: 11 meta-research sessions tested legacy axioms
Late 2026 (v0.1.0) Initial Unified Release: Go MCP Server (9 tools), Multi-Scope Generators, Package Distribution
Next (v0.2.0)      Host feedback, client auto-discovery, telemetry & cost tracking
Future (v1.0.0)    First LTS release — frozen MCP tool schemas, locked template contracts, knowledge graph
```

---

## Architecture Overhaul Decisions

These architectural decisions were made during the pre-v1.0 empirical overhaul and are captured here for provenance. They are NOT part of the sealed meta-research — they are structural decisions about the repository and tooling.

### OVH-01: File Consolidation (38 → 27 files)

**Rationale:** First-principles audit revealed that of 38 files, only 6 were actually touched by users. 23 were meta/internal-use, 7 were redundant/premature.

| Merged | Into | Why |
|---|---|---|
| PRINCIPLES.md + EVIDENCE-GRADING.md | FRAMEWORK.md | Eliminated fragmentation; one spec file, one read |
| CONTEXT.md + VISION.md + ROADMAP.md | README.md (+ this file) | About-Vivechak content consolidated; no user ever reads 3 separate "about" docs |
| META-PROMPT-GENERATOR.md | GENERATOR.md | Cleaner name, focused on the prompt (removed architecture spec for non-existent CLI) |

| Deleted | Why |
|---|---|
| templates/RESEARCH-PIPELINE.template.md | Redundant — generator already produces this |
| templates/PROMPT-LIBRARY.template.md | Redundant — generator already produces this |
| templates/COMPLEXITY-SCORING.template.md | Redundant — generator includes scoring |
| templates/EVIDENCE-RECORD.template.md | Premature — standalone evidence records are a v4.0 feature |
| schemas/ (3 JSON schemas) | Premature — no validation tooling exists to consume them |

### OVH-02: Generator Prompt Redesign (7 Problems Fixed)

| Problem | Fix | Evidence |
|---|---|---|
| Used persona ("You are a Research Pipeline Architect") | Task framing, no role assignment | Zheng et al. EMNLP 2024, Basil et al. 2025 |
| 8 mandatory structured input fields | Open-ended vision dump; AI extracts structure | P4: prescriptive on WHAT, not HOW |
| Team Size biased complexity score downward | Renamed to "Coordination Complexity," assessed by AI | Solo + AI agents ≤ small project in 2026 |
| Self-reported risk profile | AI infers regulatory exposure from vision | Users systematically underestimate risk |
| Not self-contained (name-dropped concepts) | All methodology operationalized inline | The receiving AI has never seen Vivechak |
| Told instead of showed | Operational step-by-step instructions | P4: showing > telling |
| No uncertainty handling | Undecided elements become research questions | First principles |

### OVH-03: Self-Contained Workspace Design

**Decision:** New projects copy the 4 templates into their `research/templates/` directory, making the workspace 100% independent of the meta-repo.

**Why:** AI agents (Antigravity, Claude Code, Cursor) need local contracts to autonomously execute Steps 3-5. Without templates in the workspace, agents hallucinate structure.

### OVH-04: Self-Documenting Generator Output

**Decision:** The generator prompt instructs the AI to include a "How to Execute This Pipeline" section at the top of the generated RESEARCH-PIPELINE.md.

**Why:** The generated output itself should tell the user/agent how to run the pipeline, record decisions, synthesize, and gate — without referencing the meta-repo.

### OVH-05: SaaS De-Anchoring (Critique Response)

**Trigger:** External critique identified that FOUNDING-ARCHITECTURE.template.md hardcoded SaaS web-app examples (Auth/Clerk, DB/PostgreSQL, Client?API?Services?DB diagram, package.json). Validated against our own meta-research finding T1-03 (framing bias).

**Decision:** Fix the template and generator narrowly. Reject the proposed 3-tier "Universal Epistemic Engine" rewrite.

| Critique Claim | Verdict | Action |
|---|---|---|
| FAD template induces SaaS anchoring bias | **Correct** | Fixed: neutral placeholders, "Project Structure Specification" |
| Generator scope too narrow ("software project") | **Partially correct** | Fixed: "technical project," broadened examples |
| 3-Tier architecture with Domain Archetype Profiles | **Rejected** | Premature abstraction, zero demand evidence |
| Vocabulary overhaul ("Asset Primitives") | **Rejected** | Enterprise jargon hurts usability |
| "Universal Epistemic Engine for all enterprise" | **Rejected** | Scope creep without empirical validation |

**Rationale for rejection:** The critique asked us to make a One-Way Door architectural rewrite based on Grade E evidence (zero empirical validation, zero demand signal). This is the exact mistake Vivechak exists to prevent. Domain-agnostic expansion added to Phase 6 as a validation-gated frontier instead.

### OVH-06: Bounded Exploration Mandate (P4 Enhancement)

**Problem:** Vivechak prompts were prescriptive on WHAT to cover but never stated the coverage checklist is a floor, not a ceiling. Frontier models exhibited "hyper-literalism" — constraining native reasoning to only listed elements, missing emergent concerns the prompt author couldn't anticipate.

**Evidence:** "Prompting Inversion" effect documented in 2025–2026 research — overly rigid constraints on frontier models stifle native reasoning capabilities that would otherwise discover critical concerns organically.

**Decision:** Add Bounded Exploration Mandate to P4. Applied at all 3 pipeline layers:
1. Generator prompt BRIEF: "The project vision defines the starting point, not the ceiling"
2. Generator Step 3: may add sessions for concerns user didn't mention
3. Generator Step 4: instructs generated prompts to include exploration permission in APPROACH block

**Constraint:** Exploration is bounded — "justify with evidence," not unbounded freedom.

### OVH-07: Weighted Evaluation Protocol (Comparison Bias Correction)

**Problem:** Qualitative-only comparisons are vulnerable to narrative volume bias (popular tech has more positive text), verbosity bias (longer analysis reads as stronger), and vendor marketing contamination (Grade C evidence at scale).

**Evidence:** Multi-Criteria Decision Analysis (MCDA) and Analytic Hierarchy Process (AHP) are established bias-correction techniques. LLM-as-Judge research (2025–2026) documents position bias, verbosity bias, and self-preference as systematic.

**Decision:** Add Weighted Evaluation Protocol (Framework §6.4) for comparison sessions. 6-step protocol: project-derived criteria, justified weights, evidence-referenced 1-5 scores, weighted totals, sensitivity check, qualitative-quantitative synthesis.

**Constraint:** Score complements qualitative analysis, never replaces it. Coarse 1-5 scale prevents false precision.

---

## Completed Evolution Phases

### Phase 4: Framework Polish (DONE — September 2026)

Real-world usage across multiple projects (including non-software domains) validated the core methodology but revealed documentation gaps and refinement opportunities. All items completed and committed.

#### Wave 1: Documentation Gaps ✅

| Item | Where | Status |
|---|---|---|
| Pipeline failure modes | FRAMEWORK.md §3.1 | ✅ 3 recovery protocols (garbage sessions, synthesis deadlocks, gate gaps) |
| Extract-2-files guidance | README.md Step 1 | ✅ Blockquote with extraction instructions |
| Session unit normalization | FRAMEWORK.md §3.2 | ✅ Definition of "session" with deep-research vs standard-chat calibration |

#### Wave 2: Methodology Refinements ✅

| Item | Where | Status |
|---|---|---|
| Calibration closing loop | P8 in FRAMEWORK.md + DECISIONS.template.md | ✅ `review_date`, `prediction` fields, Calibration Record section |
| Scoring rubric refinement | FRAMEWORK.md §3.2 | ✅ Rubric calibration paragraph |
| Domain-agnostic formalization | FRAMEWORK.md §8 | ✅ "Beyond Software" + "Archetype Maintenance" subsections |

#### Wave 3: Archetypes & Templates ✅

| Item | Where | Status |
|---|---|---|
| Archetype expansion guidance | FRAMEWORK.md §8 | ✅ Domain adaptation steps documented |
| Archetype revision cadence | FRAMEWORK.md §8 | ✅ Maintenance model with review triggers |

#### Systemic Fix: Change Propagation Map ✅

Added after discovering that Phase 4 improvements to FRAMEWORK.md hadn't been propagated to GENERATOR.md. Now in CONTRIBUTING.md with mandatory check in AGENTS.md §3.

### Open Backlog (Identified via Audit, Not Yet Addressed)

Items identified during the comprehensive repository audit that are valid but not yet implemented. Ordered by severity:

| ID | Severity | What | Where | Why Not Done Yet |
|---|---|---|---|---|
| SM-01 | Moderate | Domain Novelty / Technical Novelty dimension overlap risk | FRAMEWORK.md §3.2 scoring rubric | Requires real-world data on whether double-counting inflates scores. Track across next 3 projects. |
| SM-02 | Moderate | Sharp tier boundary cliff at score 15→16 | FRAMEWORK.md §3.2 tier mapping | Needs rubric calibration data. Added guidance for boundary cases (F-22) but structural fix needs evidence. |
| AB-03 | ~~Moderate~~ | ~~One-way door examples missing inter-service patterns~~ | GENERATOR.md Step 3 | ✅ DONE — Added 4 patterns (inter-service, event sourcing, monolith/micro, data partitioning) |
| GA-02 | Opportunity | No protocol for human stakeholder disagreement on ADRs | templates/ | ACH handles model disagreement but not team disagreement. Consider adding to CONFLICT-RESOLUTION template. |
| GA-04 | Opportunity | Quality rubric exists but no evaluation process defined | FRAMEWORK.md §4.3 | Who evaluates session output quality? Self-assessed? Template-based? Define when friction emerges. |
| ES-03 | Opportunity | Missing "tool-generated" verification method | FRAMEWORK.md §5 evidence system | Evidence produced by running benchmarks/code is distinct from "fetched." Add when Engine exists. |
| OP-02 | Opportunity | No worked example of a failed/corrected pipeline | examples/ | A failure example would be more instructive than the success example alone. Create when real failure data available. |
| OP-03 | Opportunity | No positioning vs Structured MADR 1.0 (2026) | README.md | MADR 4.0 + Structured MADR 1.0 converge on Vivechak's ADR format. Clarify differentiation. |
| DA-03 | Opportunity | Context injection token budget not quantified | FRAMEWORK.md §3.1 dependency protocol | Protocol says "3-5 sentences" but doesn't specify token budget. Quantify from Engine usage data. |
| PE-01 | ~~Opportunity~~ | ~~\"Context engineering\" terminology evolution not reflected~~ | FRAMEWORK.md | ✅ DONE — Added context engineering note to P1 (Phase 1.7) |

**Explicitly evaluated as OUT OF SCOPE** (not Vivechak's responsibility):
- Tool-use calibration — about agent harness design, not research methodology
- FAD → Fitness Functions — post-Vivechak concern (the coding tool owns fitness functions)
- Archetype fallback monitoring — only relevant when archetypes are programmatic (Engine v1.0+)

### Phase 4.5: Methodology Expansion (DONE — 2026-09-27)

Three-scope model: Vivechak now operates at project, decision, and comparison scope levels. Research corpus: 12 sessions (~770KB), audited implementation plan.

| # | Deliverable | Status |
|---|---|---|
| **1.1** | `GENERATOR-DECISION.md` — decision-scope generator (R/N/B/X profiling, 4-lane routing, L/C/F roles, 3-session cap) | ✅ |
| **1.2** | `GENERATOR-COMPARISON.md` & `templates/COMPARISON-SESSION.template.md` — single-prompt WEP generator with SCOPE CHECK guard | ✅ |
| **1.3** | `FRAMEWORK.md` §9: Multi-Scope Research — Scope × Depth model, 9 invariants (I1-I9), R/N/B/X routing matrix | ✅ |
| **1.4** | Reposition all public-facing text — "pre-development" → "evidence-grounded research for technical decisions" | ✅ |
| **1.5** | Evidence grading lineage — "Cochrane/GRADE lineage" → "Admiralty-Code-derived source grading with GRADE-inspired modifiers" | ✅ |
| **1.6** | Drift prevention — `CORE:BEGIN/END` marker blocks across 3 generators; CONTRIBUTING.md propagation map updated | ✅ |
| **1.7** | Context engineering note on P1 (addresses PE-01) | ✅ |

### Phase 5: MCP Server — Evidence-Grounded Automation

**Status:** ✅ Completed and Shipped in v0.1.0 (September 2026)

**What was built:** 12 research sessions (~770KB of evidence) investigated how to transform Vivechak from a copy-paste workflow into tooling that agents can invoke directly. The research decisively concluded that an MCP (Model Context Protocol) server is the correct delivery vehicle — not a standalone Engine built on LangGraph, ADK, or CrewAI.

**Key decisions (all locked, evidence-grounded):**

| Decision | Type | Key Evidence |
|---|---|---|
| **MCP server replaces standalone Engine** | Locked | MCP 2026-07-28 spec (stateless, universal agent support); compound failure risk of standalone Engine |
| **Go for server language** | Locked | Single binary distribution, 5–20ms startup, `os.Root` security, Tier 1 SDK |
| **Guided Worker pattern** | Locked | Server returns guidance in response data; no server-side FSM; no LLM API calls |
| **9 tools with `vivechak_` prefix** | Locked | Code-level audit dropped `synthesize` (violates Guided Worker), renamed 2 tools |
| **Workspace files as canonical state** | Locked | No `.vivechak/state.json`; research/ directory IS the database; Git-compatible |
| **3 scope levels (Project / Decision / Comparison)** | Locked | Cochrane, ODNI ICD 203, R-03/R-06 cross-validation |
| **Binary-first distribution** | Locked | Agent Plugins not universal; PATH truncation on macOS GUI hosts |

**What the MCP server automates:**

```
Current (Manual):
  Human → Copy GENERATOR prompt → Paste into AI chat → Get pipeline
  → Copy each session prompt → Paste into separate chats → Save outputs
  → Record decisions → Resolve conflicts → Synthesize FAD → Run gate

Automated (MCP Server):
  Agent → vivechak_init → vivechak_prepare_generator → generate pipeline
  → vivechak_save_plan → vivechak_next_session → run sessions
  → vivechak_save_session → vivechak_record_decision
  → vivechak_run_gate → sealed FAD
```

**Distribution:** GitHub Releases (GoReleaser, 6 platforms) → shell installers → `vivechak mcp-config --client <host> --write` → Homebrew/Scoop/winget → thin Agent Plugin → MCP Registry.

**Evidence base:** 12 research sessions (~770KB evidence corpus) archived in `meta-research/v2-research/`. Full implementation plan in `meta-research/v2-research/FINAL-PLAN.md`.

#### Deferred Convenience Items → MCP Tool Targets

These items were evaluated during Phase 4 planning and deferred to the MCP server rather than being implemented as standalone scripts:

| Item | What | Target | Rationale for Deferral |
|---|---|---|---|
| Template init script (`vivechak init`) | Automate the 5-template copy + directory creation | `vivechak_init` tool ✅ | Implemented in v0.1.0 |
| YAML frontmatter validator | Validate ADR schema before synthesis | `vivechak_validate` tool ✅ | Implemented in v0.1.0 |
| Code-based generator CLI | Replace copy-paste prompt with CLI interface | `vivechak_prepare_generator` tool ✅ | Implemented in v0.1.0 |
| Blast-radius tracker | Track which decisions affect which components | Post-Phase 5 | Needs 5+ projects with tracked evidence |
| AI cost tracking | Track token/API costs per session and pipeline | Post-Phase 5 | Only relevant when host agent tracks costs |

---

## Future Roadmap

### Phase 6: Research Frontiers (v0.2.0+ and v1.0.0 Milestones)

These require significant accumulated project data from MCP server usage. Each has explicit gate conditions:

| Frontier | Description | Gate Condition |
|---|---|---|
| **DSPy Prompt Optimization** | Automated prompt refinement via compile-time optimization | Requires a quantifiable "research quality" metric — needs usage data to define |
| **Cross-Project Knowledge Graph** | Reusable evidence records across projects | Requires 5+ projects with tracked evidence in a consistent format |
| **Adaptive Prompt Evolution** | Prompts that improve from session to session within a pipeline | Requires MCP server usage data showing where prompts underperform |
| **Longitudinal Calibration** | Track prediction accuracy over time (Tetlock-style) | Requires 3+ projects with 6+ months post-FAD development data |
| **Multi-Agent Debate** | Structured adversarial debate for contested One-Way Doors | Requires understanding of where single-agent research actually fails — needs usage data |

---

## Source Material Status

### brain-archive/ — FULLY EXTRACTED ✅

| File | Content | Where It Went |
|---|---|---|
| `multi-project-synthesis-history.md` | Lineage of 7 independent project pipelines + common DNA | README.md "Origin & Philosophy" + meta-research (informed pre-validation baseline) |
| `yugm-inception-transcript-summary.md` | Conversation history that seeded the meta-framework concept | meta-research/RESEARCH-PIPELINE.md (informed the hypothesis list for empirical validation) |

**Verdict:** Both files are historical transcripts. Their content has been:
1. Synthesized into the pre-validation baseline (commit `77d1f58`)
2. Tested by the 11 meta-research sessions
3. Superseded by v1.0's evidence-grounded framework

**Safe to delete from disk.** The information lives in the current framework and meta-research provenance.

### references/ — FULLY EXTRACTED ✅

These are the 8 original project research pipelines (70 files, ~65 MB) that independently discovered the patterns Vivechak consolidated:

| Project | Files | Extraction |
|---|---|---|
| project-1/ | 5 pipeline docs | Common patterns → initial principles → tested in meta-research |
| project-2/ | 1 pipeline doc | Data pipeline patterns → initial principles → tested |
| project-3/ | 1 pipeline + 16 decisions | Decision registry pattern → D-NNN schema → validated in T2-05 |
| project-4/ | 2 docs | Module partitioning → initial aspect isolation → refined to P1 Context Architecture |
| project-5/ | 18 HTML/DOCX files | Strategic vs Technical pipeline → initial tiers → refuted/refined in T2-03 |
| project-6/ | 11 pipeline + synthesis docs | Tier-based prompts → initial prompt library → replaced by 5-block anatomy |
| project-7/ | 3 docs | Unbiased cataloging → initial principle → refined in T1-03, T2-06 |
| project-8/ | 5 docs | 3-tier pipeline → initial topology → refined to DAG in T2-03 |

**Extraction chain:** `references/` → pre-validation baseline → meta-research tests every pattern → v1.0 replaces all patterns with evidence-grounded versions.

**Verdict:** The references served as the raw material for the initial framework, which was then empirically validated/refuted into v1.0. Every pattern from the references has been either:
- **Validated** and incorporated into FRAMEWORK.md (e.g., decision registries, evidence grading)
- **Refined** with empirical corrections (e.g., aspect isolation → context architecture)
- **Refuted** and replaced (e.g., fixed session counts → 4-tier adaptive scaling)

**Safe to delete from disk.** The meta-research/ directory contains the complete audit trail, and FRAMEWORK.md contains the resulting methodology.
