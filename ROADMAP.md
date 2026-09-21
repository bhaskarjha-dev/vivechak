# Vivechak Roadmap & Development History
### Evolution, Decisions, and Future Direction

---

## Current State: v1.0 (August 2026)

**Status:** Complete and operational. Ready for real-world use.

The framework, generator, and templates are fully functional, released as Vivechak v1.0.0. The Generation 3 overhaul was the most significant transformation — subjecting the methodology to its own principles, discovering that 0 of 10 legacy axioms survived unchanged, and rewriting every component from empirical evidence.

### What v1.0 Delivers

| Component | State | Description |
|---|---|---|
| **GENERATOR.md** | ✅ Ship-ready | Open-ended vision input, AI-driven classification, self-contained prompt |
| **FRAMEWORK.md** | ✅ Consolidated | Complete standalone spec (principles + evidence grading + methodology) |
| **4 Templates** | ✅ Agent-ready | Decisions, Conflict Resolution, FAD, Phase 0 Gate |
| **README.md** | ✅ Streamlined | Quick start, repo structure, agentic workflow, origin, this roadmap |
| **meta-research/** | ✅ Sealed | 14 artifacts, 31 evidence nodes, 10 hypothesis verdicts |

### Repository Evolution

```
Gen 1 (2024-2025) 7 independent project pipelines → pre-development patterns discovered
Gen 2 (Early 2026) Consolidated into URP meta-framework (rigid: 17-27 sessions, mandatory triangulation)
Meta-Research     11 independent research sessions tested every legacy axiom
v1.0 (Aug 2026)   First Public Release of Vivechak (Generation 3 Architecture: 8 principles, 4 tiers, 5-block prompts)
```

---

## Generation 3 Architecture Overhaul Decisions

These architectural decisions were made during the Generation 3 overhaul and are captured here for provenance. They are NOT part of the sealed meta-research — they are structural decisions about the repository and tooling.

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
| Team Size biased complexity score downward | Renamed to "Coordination Complexity," assessed by AI | Solo + AI agents ≠ small project in 2026 |
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

**Problem:** Vivechak prompts were prescriptive on WHAT to cover but never stated the coverage checklist is a floor, not a ceiling. Frontier models exhibited "hyper-literalism" � constraining native reasoning to only listed elements, missing emergent concerns the prompt author couldn't anticipate.

**Evidence:** "Prompting Inversion" effect documented in 2025�2026 research � overly rigid constraints on frontier models stifle native reasoning capabilities that would otherwise discover critical concerns organically.

**Decision:** Add Bounded Exploration Mandate to P4. Applied at all 3 pipeline layers:
1. Generator prompt BRIEF: "The project vision defines the starting point, not the ceiling"
2. Generator Step 3: may add sessions for concerns user didn't mention
3. Generator Step 4: instructs generated prompts to include exploration permission in APPROACH block

**Constraint:** Exploration is bounded � "justify with evidence," not unbounded freedom.

### OVH-07: Weighted Evaluation Protocol (Comparison Bias Correction)

**Problem:** Qualitative-only comparisons are vulnerable to narrative volume bias (popular tech has more positive text), verbosity bias (longer analysis reads as stronger), and vendor marketing contamination (Grade C evidence at scale).

**Evidence:** Multi-Criteria Decision Analysis (MCDA) and Analytic Hierarchy Process (AHP) are established bias-correction techniques. LLM-as-Judge research (2025�2026) documents position bias, verbosity bias, and self-preference as systematic.

**Decision:** Add Weighted Evaluation Protocol (Framework �6.4) for comparison sessions. 6-step protocol: project-derived criteria, justified weights, evidence-referenced 1-5 scores, weighted totals, sensitivity check, qualitative-quantitative synthesis.

**Constraint:** Score complements qualitative analysis, never replaces it. Coarse 1-5 scale prevents false precision.

---

## What's Next: Phase 4

### The Honest Assessment

The current roadmap listed "Code-Based Generator" as Phase 4. After the Generation 3 overhaul, this deserves scrutiny:

**The case FOR a code-based generator:**
- Deterministic classification (Layers 0-2) ensures consistency across runs
- Validation (Layer 5) catches malformed output
- CLI convenience: `npx vivechak-generate` vs copy-paste

**The case AGAINST building it now:**
- The copy-paste workflow is functional and takes 1 minute
- The AI's analysis and classification is usually good enough
- Engineering a CLI tool is significant effort for marginal usability gain
- We haven't deployed v1.0 on a real project yet — we might build the wrong tool

**Verdict:** Phase 4 should be **Real-World Validation**, not premature tooling. Build the CLI only after v1.0 has been battle-tested on 2-3 real projects.

### Phase 4: Real-World Validation (Next)

| Step | What | Why |
|---|---|---|
| **4a** | Use Vivechak v1.0 on an actual project | The ultimate validation — does the output prevent architectural mistakes? |
| **4b** | Evaluate generator output quality | Is the complexity scoring accurate? Are the prompts well-scoped? |
| **4c** | Refine the generator prompt | Based on real output, iterate on wording and instructions |
| **4d** | After 2-3 projects: assess tooling need | Does the copy-paste workflow cause friction? Is consistency a problem? |

**Success criteria:**
- The generated pipeline correctly identifies one-way vs two-way doors
- Research sessions produce actionable recommendations (not vague summaries)
- The FAD synthesis is concrete enough to start coding from
- No major architectural surprise within the first 3 months of development

### Phase 5: Tooling (If Validated)

Only build after Phase 4 demonstrates the need:

| Tool | Trigger | Description |
|---|---|---|
| Template initializer script | If copy-paste causes friction | Simple script: `vivechak init my-project` → creates directory structure + copies templates |
| YAML frontmatter validator | If malformed metadata causes synthesis problems | CI-compatible linter for session/ADR frontmatter |
| Code-based generator CLI | If AI classification inconsistency causes real problems | 5-layer deterministic/AI hybrid (the original Phase 4 plan) |
| Blast-radius tracker | If evidence decay causes undetected staleness | Cross-reference E-NNN citations across ADRs |

### Phase 6: v2.0 Research Frontiers (Future)

These require significant research investment and should not be started until Phase 4 validates the core:

| Frontier | Description | Prerequisite |
|---|---|---|
| **DSPy Prompt Optimization** | Automated prompt refinement via compile-time optimization | Requires metric: "research quality" quantified |
| **Multi-Agent Research Sessions** | Structured debate between specialized agents | Requires understanding of where single-agent fails |
| **Cross-Project Knowledge Graph** | Reusable evidence records across projects | Requires 5+ projects with tracked evidence |
| **Longitudinal Calibration** | Track prediction accuracy over time (Tetlock-style) | Requires 3+ projects with 6+ months of development data |
| **Adaptive Prompt Evolution** | Prompts that improve from session to session within a pipeline | Requires understanding of inter-session information flow |
| **Domain-Agnostic Expansion** | Extend Vivechak beyond software to hardware, manufacturing, biopharma, or CapEx decisions | Requires empirical validation on ≥1 non-software domain (see below) |

#### On Domain-Agnostic Expansion

Vivechak's epistemic core — evidence grading (Cochrane/GRADE lineage), ACH conflict resolution (CIA/Heuer), premortem protocol (Gary Klein), reversibility routing (Bezos Type 1/2 doors), and map-reduce synthesis — was not invented for software. These primitives govern how intelligence reasons about truth, risk, and irreversibility in *any* domain.

The generator prompt (v1.0) already says "technical project" rather than "software project." The FAD template uses domain-neutral Wardley Compose/Build framing. The 4 operational templates (Decisions, Conflict Resolution, FAD, Phase 0 Gate) contain zero software-specific content.

**What's missing for non-software domains:**
- The 6 domain archetypes in the generator are software-only (B2B SaaS, DevTools, FinTech, AI/ML, Consumer Mobile, Real-Time/IoT). Non-software use would need archetype-equivalent classification for hardware, manufacturing, biopharma, etc.
- The 8-dimension complexity scoring rubric uses software-centric labels ("Standard CRUD," "New library," "PII/GDPR"). Scoring labels would need domain-neutral equivalents.
- The 5-block prompt anatomy references "Principal Architect" as the target audience. Non-software domains would reference domain-equivalent decision-makers.

**Gate condition:** This frontier activates only when someone empirically uses Vivechak on a non-software domain, documents the results, and identifies specific friction points. Speculative refactoring without demand evidence is the exact premature architecture that Vivechak exists to prevent.

**What would NOT change:** The 8 core principles, evidence grading system, ACH matrix, premortem protocol, and two-track gate are already fully domain-agnostic and require zero modification.

---

## Source Material Status

### brain-archive/ — FULLY EXTRACTED ✅

| File | Content | Where It Went |
|---|---|---|
| `multi-project-synthesis-history.md` | Lineage of 7 projects (Vivah, Pramedha, Triyantra, Portfolio, Gemmra, Yugm) + common DNA | README.md "Origin & Philosophy" + meta-research (informed v2.0 baseline) |
| `yugm-inception-transcript-summary.md` | Conversation history that seeded the meta-framework concept | meta-research/RESEARCH-PIPELINE.md (informed the v2.0 → v3.0 hypothesis list) |

**Verdict:** Both files are historical transcripts. Their content has been:
1. Synthesized into the v2.0 baseline (commit `77d1f58`)
2. Tested by the 11 meta-research sessions
3. Superseded by v3.0's evidence-grounded framework

**Safe to delete from disk.** The information lives in the v3.0 framework and meta-research provenance.

### references/ — FULLY EXTRACTED ✅

These are the 7 original project research pipelines (70 files, ~65 MB) that independently discovered the patterns Vivechak consolidated:

| Project | Files | Extraction |
|---|---|---|
| forge-rachak/ | 5 pipeline docs | Common patterns → v2.0 principles → tested in meta-research |
| gemmra/ | 1 pipeline doc | Data pipeline patterns → v2.0 principles → tested |
| portfolio/ | 1 pipeline + 16 decisions | Decision registry pattern → D-NNN schema → validated in T2-05 |
| pramedha/ | 2 docs | Module partitioning → v2.0 aspect isolation → refined to P1 Context Architecture |
| rachak-research/ | 18 HTML/DOCX files | Strategic vs Technical pipeline → v2.0 tiers → refuted/refined in T2-03 |
| triyantra/ | 11 pipeline + synthesis docs | Tier-based prompts → v2.0 prompt library → replaced by 5-block anatomy |
| vivah-soodh/ | 3 docs | Unbiased cataloging → v2.0 principle → refined in T1-03, T2-06 |
| yugm/ | 5 docs | 3-tier pipeline → v2.0 topology → refined to DAG in T2-03 |

**Extraction chain:** `references/` → v2.0 baseline → meta-research tests every pattern → v3.0 replaces all patterns with evidence-grounded versions.

**Verdict:** The references served as the raw material for v1.0/v2.0, which was then empirically validated/refuted into v3.0. Every pattern from the references has been either:
- **Validated** and incorporated into FRAMEWORK.md (e.g., decision registries, evidence grading)
- **Refined** with empirical corrections (e.g., aspect isolation → context architecture)
- **Refuted** and replaced (e.g., fixed session counts → 4-tier adaptive scaling)

**Safe to delete from disk.** The meta-research/ directory contains the complete audit trail, and FRAMEWORK.md contains the resulting methodology.
