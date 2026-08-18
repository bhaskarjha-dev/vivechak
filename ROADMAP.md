# URP Roadmap & Development History
### Evolution, Decisions, and Future Direction

---

## Current State: v3.0 (August 2026)

**Status:** Complete and operational. Ready for real-world use.

The framework, generator, and templates are fully functional. The v3.0 overhaul was the most significant transformation — subjecting URP to its own methodology, discovering that 0 of 10 v2.0 axioms survived unchanged, and rewriting every component from evidence.

### What v3.0 Delivers

| Component | State | Description |
|---|---|---|
| **GENERATOR.md** | ✅ Ship-ready | Open-ended vision input, AI-driven classification, self-contained prompt |
| **FRAMEWORK.md** | ✅ Consolidated | Complete standalone spec (principles + evidence grading + methodology) |
| **4 Templates** | ✅ Agent-ready | Decisions, Conflict Resolution, FAD, Phase 0 Gate |
| **README.md** | ✅ Streamlined | Quick start, repo structure, agentic workflow, origin, this roadmap |
| **meta-research/** | ✅ Sealed | 14 artifacts, 31 evidence nodes, 10 hypothesis verdicts |

### Repository Evolution

```
v1.0 (2024-2025)  7 independent project pipelines → common patterns discovered
v2.0 (Early 2026) Consolidated into meta-framework (rigid: 17-27 sessions, mandatory triangulation)
v2.0 Meta         11 independent research sessions tested every v2.0 axiom
v3.0 (Aug 2026)   Empirically validated rewrite (8 principles, 4-tier scaling, 5-block prompts)
v3.0 Overhaul     38 → 27 files, generator redesign, self-contained workspace workflow
```

---

## v3.0 Overhaul Decisions (This Conversation)

These architectural decisions were made during the v3.0 overhaul and are captured here for provenance. They are NOT part of the sealed meta-research — they are structural decisions about the repository and tooling.

### OVH-01: File Consolidation (38 → 27 files)

**Rationale:** First-principles audit revealed that of 38 files, only 6 were actually touched by users. 23 were meta/internal-use, 7 were redundant/premature.

| Merged | Into | Why |
|---|---|---|
| PRINCIPLES.md + EVIDENCE-GRADING.md | FRAMEWORK.md | Eliminated fragmentation; one spec file, one read |
| CONTEXT.md + VISION.md + ROADMAP.md | README.md (+ this file) | About-URP content consolidated; no user ever reads 3 separate "about" docs |
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
| Not self-contained (name-dropped concepts) | All methodology operationalized inline | The receiving AI has never seen URP |
| Told instead of showed | Operational step-by-step instructions | P4: showing > telling |
| No uncertainty handling | Undecided elements become research questions | First principles |

### OVH-03: Self-Contained Workspace Design

**Decision:** New projects copy the 4 templates into their `research/templates/` directory, making the workspace 100% independent of the meta-repo.

**Why:** AI agents (Antigravity, Claude Code, Cursor) need local contracts to autonomously execute Steps 3-5. Without templates in the workspace, agents hallucinate structure.

### OVH-04: Self-Documenting Generator Output

**Decision:** The generator prompt instructs the AI to include a "How to Execute This Pipeline" section at the top of the generated RESEARCH-PIPELINE.md.

**Why:** The generated output itself should tell the user/agent how to run the pipeline, record decisions, synthesize, and gate — without referencing the meta-repo.

---

## What's Next: Phase 4

### The Honest Assessment

The current roadmap listed "Code-Based Generator" as Phase 4. After the v3.0 overhaul, this deserves scrutiny:

**The case FOR a code-based generator:**
- Deterministic classification (Layers 0-2) ensures consistency across runs
- Validation (Layer 5) catches malformed output
- CLI convenience: `npx urp-generate` vs copy-paste

**The case AGAINST building it now:**
- The copy-paste workflow is functional and takes 1 minute
- The AI's analysis and classification is usually good enough
- Engineering a CLI tool is significant effort for marginal usability gain
- We haven't used v3.0 on a real project yet — we might build the wrong tool

**Verdict:** Phase 4 should be **Real-World Validation**, not premature tooling. Build the CLI only after v3.0 has been battle-tested on 2-3 real projects.

### Phase 4: Real-World Validation (Next)

| Step | What | Why |
|---|---|---|
| **4a** | Use URP v3.0 on an actual project | The ultimate validation — does the output prevent architectural mistakes? |
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
| Template initializer script | If copy-paste causes friction | Simple script: `urp init my-project` → creates directory structure + copies templates |
| YAML frontmatter validator | If malformed metadata causes synthesis problems | CI-compatible linter for session/ADR frontmatter |
| Code-based generator CLI | If AI classification inconsistency causes real problems | 5-layer deterministic/AI hybrid (the original Phase 4 plan) |
| Blast-radius tracker | If evidence decay causes undetected staleness | Cross-reference E-NNN citations across ADRs |

### Phase 6: v4.0 Research Frontiers (Future)

These require significant research investment and should not be started until Phase 4 validates the core:

| Frontier | Description | Prerequisite |
|---|---|---|
| **DSPy Prompt Optimization** | Automated prompt refinement via compile-time optimization | Requires metric: "research quality" quantified |
| **Multi-Agent Research Sessions** | Structured debate between specialized agents | Requires understanding of where single-agent fails |
| **Cross-Project Knowledge Graph** | Reusable evidence records across projects | Requires 5+ projects with tracked evidence |
| **Longitudinal Calibration** | Track prediction accuracy over time (Tetlock-style) | Requires 3+ projects with 6+ months of development data |
| **Adaptive Prompt Evolution** | Prompts that improve from session to session within a pipeline | Requires understanding of inter-session information flow |

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

These are the 7 original project research pipelines (70 files, ~65 MB) that independently discovered the patterns URP consolidated:

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
