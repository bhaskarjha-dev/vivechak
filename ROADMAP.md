# Vivechak Roadmap & Development History
### Evolution, Decisions, and Future Direction

---

## Current State: v1.1 (September 2026)

**Status:** Operational and battle-tested. Used on multiple real projects including non-software domains.

The framework was released as v1.0 in August 2026, then refined to v1.1 through real-world usage across multiple projects and a comprehensive first-principles audit. Key v1.1 improvements include: the Bounded Exploration Mandate (P4 enhancement), Weighted Evaluation Protocol (comparison bias correction), domain-agnostic generator language, and extensive documentation cleanup.

### What v1.1 Delivers

| Component | State | Description |
|---|---|---|
| **GENERATOR.md** | ✅ Ship-ready | Open-ended vision input, AI-driven classification, self-contained prompt |
| **FRAMEWORK.md** | ✅ Consolidated | Complete standalone spec (principles + evidence grading + methodology) |
| **4 Templates** | ✅ Agent-ready | Decisions, Conflict Resolution, FAD, Phase 0 Gate |
| **README.md** | ✅ Streamlined | Quick start, repo structure, agentic workflow, origin |
| **meta-research/** | ✅ Sealed | 15 artifacts, 31 evidence nodes, 10 hypothesis verdicts |

### Repository Evolution

```
2024-2025         7 independent project pipelines → pre-development patterns discovered
Early 2026        Consolidated into initial meta-framework (rigid: 17-27 sessions, mandatory triangulation)
Mid 2026          11 independent research sessions tested every legacy axiom
v1.0 (Aug 2026)   First public release of Vivechak (8 principles, 4 tiers, 5-block prompts)
v1.1 (Sep 2026)   Real-world usage refinements, first-principles audit, version normalization
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

## What's Next

### Phase 4: Framework Polish (DONE — v1.1.0, 2026-09-22)

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
| PE-01 | Opportunity | "Context engineering" terminology evolution not reflected | FRAMEWORK.md | P1 aligns with the 2026 "context engineering" paradigm but doesn't use the term. Terminology update. |

### Phase 5: Vivechak Engine — Research Phase

**The vision:** Transform Vivechak from a prompt-paste manual workflow into an automated research execution engine that takes a problem statement and produces a complete research corpus + FAD.

**The approach:** Use Vivechak's own GENERATOR.md to generate a research pipeline for "how to build an automated research execution engine." This is the ultimate dogfooding — using the framework to research its own automation, which simultaneously validates the current generator and produces an evidence-grounded architecture for the Engine.

#### What the Engine Would Automate

```
Current (Manual):
  Human → Copy GENERATOR prompt → Paste into AI chat → Get pipeline
  → Copy each session prompt → Paste into separate chats → Save outputs
  → Record decisions → Resolve conflicts → Synthesize FAD → Run gate

Automated (Engine):
  Human → Provide problem statement → Engine executes entire pipeline
  → Research sessions run in parallel (respecting DAG dependencies)
  → Triangulation for marked sessions (multi-model)
  → Synthesis and gate checks → Complete research/ directory output
```

#### Research Questions for Engine Architecture

These must be answered through proper Vivechak-based research before building:

| Question | Why It Matters |
|---|---|
| **Orchestration framework** — LangGraph vs Google ADK vs CrewAI vs custom | One-Way Door: framework lock-in affects all downstream development. Needs deep research on each framework's DAG support, human-in-the-loop, state persistence, multi-model orchestration, and long-term viability |
| **Language choice** — Python vs Go vs Rust vs TypeScript | One-Way Door: affects framework options, ecosystem, and maintenance. Not all frameworks support all languages |
| **LLM API strategy** — Direct SDK vs unified layer (litellm) vs provider-specific | Affects multi-model triangulation, cost tracking, and provider flexibility |
| **Web search strategy** — Native model capabilities vs supplemental search APIs (Tavily, Exa) | Affects research quality and cost |
| **Deployment model** — CLI tool vs IDE plugin vs web service vs hybrid | Affects who can use it and how |
| **Separate repo or monorepo** — Engine as separate project vs inside vivechak/ | Affects dependency management and framework independence |

#### Phased Engine Development (Post-Research)

| Version | What | Value |
|---|---|---|
| **v0.1** | Single-session runner — CLI takes one prompt, calls LLM API, saves formatted output | Eliminates copy-paste-save for individual sessions |
| **v0.2** | Pipeline executor — parses RESEARCH-PIPELINE.md, executes DAG, parallel sessions | Automates the 10-25 session execution cycle |
| **v0.3** | End-to-end — takes problem statement, generates pipeline, executes, synthesizes FAD | Full automation from idea to architecture document |
| **v1.0** | Production — multi-model triangulation, human-in-the-loop gates, resume from checkpoint, cost tracking | Production-grade research automation |

### Phase 6: Research Frontiers (Future)

These require either the Engine to exist or significant accumulated project data. Each has explicit gate conditions:

| Frontier | Description | Gate Condition |
|---|---|---|
| **DSPy Prompt Optimization** | Automated prompt refinement via compile-time optimization | Requires a quantifiable "research quality" metric — which requires Engine usage data to define |
| **Cross-Project Knowledge Graph** | Reusable evidence records across projects | Requires 5+ projects with tracked evidence in a consistent format |
| **Adaptive Prompt Evolution** | Prompts that improve from session to session within a pipeline | Requires Engine v0.2+ (inter-session context injection only possible with programmatic control) |
| **Longitudinal Calibration** | Track prediction accuracy over time (Tetlock-style) | Requires 3+ projects with 6+ months post-FAD development data |
| **Multi-Agent Debate** | Structured adversarial debate for contested One-Way Doors | Requires understanding of where single-agent research actually fails — needs Engine usage data |

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

These are the 7 original project research pipelines (70 files, ~65 MB) that independently discovered the patterns Vivechak consolidated:

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
