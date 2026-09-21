# Vivechak — Decision Registry (ADR Corpus)
### Ground-Truth Architectural Decision Records & Empirical Verdicts (Vivechak v3.0)
**Registry ID:** Vivechak-META-DECISIONS  
**Synthesized:** August 2026 via Grand Synthesis SYN-01  
**Schema Version:** 3.0  
**Status:** Sealed Baseline Specification  

---

## Executive Summary & Registry Metadata

This registry documents the empirical validation and formal architectural decisions for the Vivechak (विवेचक) Meta-Framework. Each entry represents a founding hypothesis from Vivechak v2.0 subjected to rigorous empirical testing across 11 meta-research sessions (T1-01 through T3-01) and resolved into a permanent architectural invariant for Vivechak v3.0.

### Hypothesis Verdict Summary Table

| ID | Topic / Hypothesis | Door Type | Status | Informed By | Confidence |
|---|---|---|---|---|---|
| **D-001** | Aspect-Isolation Law vs Context Architecture | **One-Way** | `REFINED` | T1-02, T1-03, T2-01 | **High** |
| **D-002** | Multi-Model Triangulation Protocol | **One-Way** | `REFINED` | T1-02, T2-02 | **High** |
| **D-003** | Constrained DAG Pipeline Topology | **Two-Way** | `REFINED` | T1-01, T2-03 | **High** |
| **D-004** | Risk-Calibrated 4-Tier Scaling Model | **One-Way** | `REFUTED` | T1-01, T2-04 | **High** |
| **D-005** | GRADE-Aligned Evidence Standard (A–E) | **One-Way** | `VALIDATED (Refined)` | T1-01, T1-02, T2-05 | **High** |
| **D-006** | 5-Block Prompt Anatomy & Dynamic Search | **Two-Way** | `REFUTED` | T1-03, T2-06 | **Very High** |
| **D-007** | Hybrid Output Architecture (MD + YAML) | **Two-Way** | `VALIDATED (Refined)` | T1-01, T2-07 | **High** |
| **D-008** | Commodity-Maximized Composition | **One-Way** | `REFINED` | T1-01, T2-04, SYN-01 | **High** |
| **D-009** | Two-Track Phase 0 Exit Gate & Premortem | **One-Way** | `REFINED` | T1-01, T2-05, SYN-01 | **High** |
| **D-010** | 5-Layer Hybrid Generator Architecture | **One-Way** | `REFUTED` | T1-03, T3-01 | **High** |

---

## Master Evidentiary Index (`E-001` through `E-031`)

The following empirical evidence nodes underpin the decisions in this registry:

| Evidence ID | Description / Citation | Source Type | Base Grade | Modifiers |
|---|---|---|---|---|
| **E-001** | Anthropic regression on OpenAI BrowseComp benchmark ($R^2=0.80$ variance explained by token budget) | Lab Empirical Data | Grade B | `corroborated · fresh · direct` |
| **E-002** | Multi-Task Inference (MTI) benchmark (+12.4% accuracy on coupled tasks) | Empirical Benchmark | Grade B | `single · fresh · direct` |
| **E-003** | Chandler & Sweller (1992) Cognitive Load & Split-Attention Effect | Academic Literature | Grade A | `corroborated · aging · indirect` |
| **E-004** | Chroma Research (2025) Context Rot across 18 frontier LLMs | Empirical Benchmark | Grade B | `corroborated · fresh · direct` |
| **E-005** | Kim et al. / Goel et al. (ICML 2025) Correlated Errors (~60% shared wrong answers on HELM) | Peer-Reviewed Study | Grade A | `corroborated · fresh · direct` |
| **E-006** | Gao & Xiao (2026) Nonstandard Errors in AI Agents (Model house styles on subjective tasks) | Peer-Reviewed Study | Grade A | `corroborated · fresh · direct` |
| **E-007** | Lorenz et al. (2011) Wisdom of Crowds breakdown under correlated priors | Academic Literature | Grade A | `corroborated · aging · indirect` |
| **E-008** | PMBOK Mandatory vs Discretionary Dependency Classification Standard | Industry Standard | Grade A | `corroborated · aging · direct` |
| **E-009** | T2-03 Dependency Mapping (67% of Tier 2 sessions have $\le 1$ Tier 1 dependency) | Meta-Research Artifact | Grade B | `corroborated · fresh · direct` |
| **E-010** | Snowden (1999) Cynefin Framework for Decision Contexts | Management Science | Grade A | `corroborated · aging · direct` |
| **E-011** | DORA & Walkinshaw et al. (2018) Defect Concentration & Architectural Failure Rates | Empirical Study | Grade A | `corroborated · fresh · direct` |
| **E-012** | Amazon Type 1 vs Type 2 Door Decision Framework (Bezos 2015) | Industry Standard | Grade B | `corroborated · aging · direct` |
| **E-013** | GRADE Working Group Certainty & Recommendation Standards (Guyatt et al.) | Clinical Standard | Grade A | `corroborated · aging · direct` |
| **E-014** | Critical Review of the Admiralty Code (1968/2019: 87% diagonal collapse) | Intelligence Analysis | Grade A | `corroborated · aging · direct` |
| **E-015** | LLM Reference Hallucination & Citation Confabulation Instances (T1-02) | Empirical Finding | Grade B | `corroborated · fresh · direct` |
| **E-016** | WHO Surgical Safety Checklist Trial (Gawande et al., 47% mortality reduction) | Clinical Trial | Grade A | `corroborated · aging · indirect` |
| **E-017** | Laban et al. (ICLR 2026 Outstanding Paper, 39% multi-turn performance drop) | Peer-Reviewed Study | Grade A | `corroborated · fresh · direct` |
| **E-018** | Zheng et al. (EMNLP 2024) / Basil et al. (2025) Persona Accuracy Degradation | Peer-Reviewed Study | Grade A | `corroborated · fresh · direct` |
| **E-019** | ReAct Search Paradigm (Yao et al. 2023) Interleaved Reasoning & Acting | Peer-Reviewed Study | Grade A | `corroborated · fresh · direct` |
| **E-020** | Tam et al. (EMNLP 2024) Format Restrictions Impair Reasoning Ability | Peer-Reviewed Study | Grade A | `corroborated · fresh · direct` |
| **E-021** | Markdown Any Decision Records (MADR 4.0) & Python PEP Standard | Open Standard | Grade A | `corroborated · fresh · direct` |
| **E-022** | Microsoft Research GraphRAG & Header-Based Chunking Benchmarks | Empirical Study | Grade B | `corroborated · fresh · direct` |
| **E-023** | Git Line-Level Diff Mechanics in JSON vs YAML vs Markdown (Grumpy Gamer) | Engineering Analysis | Grade B | `corroborated · aging · direct` |
| **E-024** | Wardley Mapping: Evolution of Genesis, Custom, Product, Commodity | Strategy Framework | Grade A | `corroborated · aging · direct` |
| **E-025** | Martin Fowler: MonolithFirst & Microservice Premium Architectures | Industry Standard | Grade A | `corroborated · aging · direct` |
| **E-026** | DORA Metrics: Loosely-Coupled Architecture & Delivery Performance | Industry Study | Grade A | `corroborated · fresh · direct` |
| **E-027** | Systematic Reviews of Clinical Decision Support Alert Overrides (49–96% rate) | Clinical Literature | Grade A | `corroborated · aging · indirect` |
| **E-028** | Gary Klein (1989) Prospective Hindsight / Premortem Method (30% risk reduction) | Cognitive Science | Grade A | `corroborated · aging · direct` |
| **E-029** | Kent Beck: Extreme Programming Technical Spike Methodology | Engineering Practice| Grade A | `corroborated · aging · direct` |
| **E-030** | Generator Architectures: Nx, Yeoman, Plop, cookiecutter Ecosystem Survey | Engineering Analysis | Grade A | `corroborated · fresh · direct` |
| **E-031** | Prompt De-composition & Prompt-Chaining vs Monolithic LLM Generation | Empirical Study | Grade B | `corroborated · fresh · direct` |

---

## Detailed Architectural Decision Records (`D-001` to `D-010`)

---

### D-001: Aspect-Isolation Law vs Context Architecture Law
```yaml
id: D-001
title: "Superseding Aspect-Isolation Law with Context Architecture Law"
status: accepted
door_type: one-way
date: 2026-08-18
confidence: high
evidence_refs: [E-001, E-002, E-003, E-004]
informed_by_sessions: [T1-02, T1-03, T2-01]
supersedes: null
review_trigger: "Re-evaluate if LLM context architectures achieve 100% flat retrieval curves with zero lost-in-the-middle degradation across 1M+ tokens"
tags: [methodology, context-engineering, aspect-isolation, synthesis]
authored_by: "principal-research-architect"
human_reviewed: true
schema_version: "3.0"
```

#### Context & Problem Statement
Vivechak v2.0 mandated an absolute "Aspect-Isolation Law": every research topic must be isolated into a single-topic session with zero topic bundling. The hypothesis was that bundling multiple topics catastrophically degrades research depth due to search budget dilution.

#### Evaluated Options
1. **Option 1 (v2.0 Dogma):** Absolute session isolation for all topics without exception.
2. **Option 2:** Monolithic joint research prompts covering all aspects in single sessions.
3. **Option 3 (v3.0 Resolution):** Conditional context decomposition based on coupling, paired with mandatory explicit synthesis.

#### Decision Outcome & Status
**Status:** `REFINED`  
**Chosen Option:** Option 3 — The Context Architecture Law.

#### Ground-Truth Empirical Evidence
- Anthropic regression on the OpenAI BrowseComp benchmark proves that finite token/tool-call budget ($R^2=0.80$) is the true physical constraint (E-001). Diluting search queries degrades factual depth.
- However, absolute isolation fails: the Multi-Task Inference (MTI) benchmark demonstrates up to +12.4% accuracy and 1.46× faster inference when coupled, interdependent sub-tasks are evaluated jointly (E-002).
- Over-isolation triggers the split-attention effect (E-003), introduces 4–15× token overhead, and misses systemic cross-cutting architectural trade-offs (E-004).

#### Vivechak v3.0 Resolution & Operational Rule
Decompose into dedicated research contexts when sub-tasks have low interdependency, high individual complexity, or divergent search spaces. Integrate coupled topics into structured joint sessions when evaluating holistic system tradeoffs. Every decomposed investigation **must** conclude with an explicit downstream synthesis pass.

---

### D-002: Multi-Model Triangulation Protocol
```yaml
id: D-002
title: "Risk-Gated Staged Triangulation Protocol"
status: accepted
door_type: one-way
date: 2026-08-18
confidence: high
evidence_refs: [E-005, E-006, E-007]
informed_by_sessions: [T1-02, T2-02]
supersedes: null
review_trigger: "Re-evaluate if frontier AI architectures diverge fundamentally in training corpora and eliminate correlated factual errors"
tags: [triangulation, multi-model, decision-science, verification]
authored_by: "principal-research-architect"
human_reviewed: true
schema_version: "3.0"
```

#### Context & Problem Statement
Vivechak v2.0 mandated running identical research prompts across Claude, ChatGPT, and Gemini for all major architectural decisions to eliminate single-model bias.

#### Evaluated Options
1. **Option 1 (v2.0 Dogma):** Mandatory 3-model triangulation across all research sessions.
2. **Option 2:** Single-model research only, with zero multi-model cross-checks.
3. **Option 3 (v3.0 Resolution):** Staged, risk-triggered triangulation protocol (Single model default $\to$ Cheap critique probe $\to$ Full triangulation).

#### Decision Outcome & Status
**Status:** `REFINED`  
**Chosen Option:** Option 3 — Staged, Risk-Triggered Triangulation Protocol.

#### Ground-Truth Empirical Evidence
- On factual and verifiable questions, frontier models converge 70–90% and share identical correlated error distributions (~60% shared wrong answers on HELM, Kim et al. ICML 2025; E-005). Triangulating factual queries buys false confidence from correlated consensus (E-007).
- Divergence is genuine and valuable on subjective, ambiguous, or predictive trade-offs where model training priors differ (E-006).
- Indiscriminate 3-model execution incurs 3× token spend, 3× latency, and heavy human reconciliation fatigue without proportional error reduction.

#### Vivechak v3.0 Resolution & Operational Rule
Default to single-model deep research. Use cheap cross-model critique probes for medium-stakes decisions. Escalate to full multi-model triangulation exclusively for high-stakes, irreversible One-Way Doors with genuine expert contestation, or when probes detect active divergence.

---

### D-003: Constrained DAG Pipeline Topology
```yaml
id: D-003
title: "Constrained DAG Topology with Adaptive Checkpoints"
status: accepted
door_type: two-way
date: 2026-08-18
confidence: high
evidence_refs: [E-008, E-009]
informed_by_sessions: [T1-01, T2-03]
supersedes: null
review_trigger: "Re-evaluate if human coordination error on DAG execution exceeds 15% of pipeline time"
tags: [topology, dag, execution-graph, dependencies]
authored_by: "principal-research-architect"
human_reviewed: true
schema_version: "3.0"
```

#### Context & Problem Statement
Vivechak v2.0 enforced a rigid 3-tier linear staging model (Tier 1 Landscape $\to$ Tier 2 Architecture $\to$ Tier 3 Blueprints), blocking downstream tiers until an entire upstream stage finished.

#### Evaluated Options
1. **Option 1 (v2.0 Dogma):** Strict 3-tier sequential stage gating.
2. **Option 2:** Fully unconstrained ad-hoc research execution.
3. **Option 3 (v3.0 Resolution):** Constrained DAG with inverted dependency defaults and bounded adaptive checkpoints.

#### Decision Outcome & Status
**Status:** `REFINED`  
**Chosen Option:** Option 3 — Constrained DAG with Adaptive Checkpoints.

#### Ground-Truth Empirical Evidence
- Rigid stage-gating forces unrelated research to wait, artificially inflating the critical path (e.g., Frontend Framework blocked on Competitor Scan; E-009).
- ~67% of Tier 2 architectural decisions depend on $\le 1$ Tier 1 session, not all 5 (E-009).
- PMBOK standards distinguish mandatory (hard information) from discretionary (soft contextual) dependencies (E-008).

#### Vivechak v3.0 Resolution & Operational Rule
Sessions default to unblocked (eligible to run immediately). Hard information dependencies block execution; soft contextual dependencies are supplied via the living Shared Context Brief (`_index.yaml` + locked ADRs) without blocking. Tiers are retained strictly as organizational metadata.

---

### D-004: Risk-Calibrated 4-Tier Scaling Model
```yaml
id: D-004
title: "Risk-Calibrated 4-Tier Scaling Model and Decision Routing Matrix"
status: accepted
door_type: one-way
date: 2026-08-18
confidence: high
evidence_refs: [E-010, E-011, E-012]
informed_by_sessions: [T1-01, T2-04]
supersedes: null
review_trigger: "Re-evaluate if projects scoring <5 points suffer >15% architectural failure rate in production"
tags: [scaling, complexity-scoring, door-type, risk-calibration]
authored_by: "principal-research-architect"
human_reviewed: true
schema_version: "3.0"
```

#### Context & Problem Statement
Vivechak v2.0 prescribed a fixed count of 17–27 research sessions for every software venture, treating all projects as requiring identical diligence.

#### Evaluated Options
1. **Option 1 (v2.0 Dogma):** Fixed 17–27 session prescription.
2. **Option 2:** Unconstrained user self-selection of session count.
3. **Option 3 (v3.0 Resolution):** 8-dimension complexity scoring rubric (0–24) $\to$ 4 Tiers (1–30 sessions) + Per-Decision Door-Type Routing Matrix.

#### Decision Outcome & Status
**Status:** `REFUTED`  
**Chosen Option:** Option 3 — Risk-Calibrated 4-Tier Scaling Model.

#### Ground-Truth Empirical Evidence
- 17–27 sessions severely over-researches simple CRUD projects and hackathons while under-researching complex, novel platforms (E-010).
- Architectural failure concentrates in a small core of irreversible choices ($<20\%$ of components drive $>80\%$ of architectural risk, DORA / Walkinshaw et al.; E-011).
- Rigor must scale with reversibility and knowability (Amazon 1-way/2-way door model, Bezos; E-012).

#### Vivechak v3.0 Resolution & Operational Rule
Score project complexity across 8 dimensions (0–24 pts) to establish budget ceilings: Tier 0 Minimal (1–3), Tier 1 Light (4–8), Tier 2 Standard (9–16), Tier 3 Deep (17–30). Route individual decisions via the 2×2 Reversibility × Familiarity Matrix. Enforce a hard override for high regulatory exposure ($D_3=3$).

---

### D-005: GRADE-Aligned Evidence Standard (A–E)
```yaml
id: D-005
title: "GRADE-Aligned Evidence Classification with Verification Metadata"
status: accepted
door_type: one-way
date: 2026-08-18
confidence: high
evidence_refs: [E-013, E-014, E-015, E-016]
informed_by_sessions: [T1-01, T1-02, T2-05]
supersedes: null
review_trigger: "Re-evaluate if inline grading metadata reduces developer compliance below 60% in production telemetry"
tags: [evidence-grading, grade-framework, verification-method, provenance]
authored_by: "principal-research-architect"
human_reviewed: true
schema_version: "3.0"
```

#### Context & Problem Statement
Vivechak v2.0 used a static 5-tier evidence grade (A: Primary to E: Speculation) based solely on source prestige, without modifiers or AI verification tracking.

#### Evaluated Options
1. **Option 1 (v2.0 Dogma):** Static 5-tier A–E grading by source prestige alone.
2. **Option 2:** Binary "verified / unverified" classification.
3. **Option 3:** Full 36-cell orthogonal Admiralty Code matrix (Source Reliability × Information Credibility).
4. **Option 4 (v3.0 Resolution):** A–E base grades + 3 lightweight modifiers (`corroboration`, `recency`, `directness`) + mandatory `verification_method` metadata.

#### Decision Outcome & Status
**Status:** `VALIDATED (With Structural Refinement)`  
**Chosen Option:** Option 4 — GRADE-Aligned Evidence Classification.

#### Ground-Truth Empirical Evidence
- 5 tiers is the optimal granularity (matches GRADE clinical certainty; avoids binary oversimplification and prevents Admiralty Code diagonal collapse where 87% of ratings collapse to the diagonal; E-013, E-014).
- Fixed source ranking causes the "wavy lines" flaw: stale Grade A docs outrank fresh, corroborated Grade B postmortems (E-013).
- Grade E conflated fabrication (provenance) with staleness (recency).
- AI citation confabulation (E-015) requires strict verification tracking: unverified parametric recall must be structurally capped at Grade D.

#### Vivechak v3.0 Resolution & Operational Rule
Retain A–E base grades. Redefine Grade E strictly as *Untraceable / Unverifiable*. Attach 3 modifiers (`corroboration`, `recency`, `directness`) and mandatory `verification_method` metadata. Cap `recalled` AI claims at Grade D. Decouple decision confidence from evidence grade. Maintain separate, linked `E-NNN` and `D-NNN` registries.

---

### D-006: 5-Block Prompt Anatomy & Dynamic Search Heuristics
```yaml
id: D-006
title: "5-Block Goal-Oriented Prompt Anatomy with Dynamic Search Execution"
status: accepted
door_type: two-way
date: 2026-08-18
confidence: very-high
evidence_refs: [E-017, E-018, E-019, E-020]
informed_by_sessions: [T1-03, T2-06]
supersedes: null
review_trigger: "Re-evaluate if frontier LLMs lose internal planning capacity and require procedural prompting"
tags: [prompt-engineering, react-loop, context-engineering, prompt-anatomy]
authored_by: "principal-research-architect"
human_reviewed: true
schema_version: "3.0"
```

#### Context & Problem Statement
Vivechak v2.0 mandated an 8-section XML-tagged prompt anatomy containing expert personas (`<system>`), 10–14 hardcoded search queries (`<web_searches>`), negative bias suppression (`<bias_resistance>`), and rigid output skeletons (`<output_spec>`).

#### Evaluated Options
1. **Option 1 (v2.0 Dogma):** 8-section prescriptive XML prompt template with fixed queries.
2. **Option 2:** Fully unconstrained, open-ended 1-sentence prompts.
3. **Option 3 (v3.0 Resolution):** 5-block goal-oriented anatomy (`BRIEF`, `SCOPE`, `APPROACH`, `DELIVERABLE`, `FORMAT`) with dynamic search heuristics.

#### Decision Outcome & Status
**Status:** `REFUTED`  
**Chosen Option:** Option 3 — 5-Block Minimal Effective Prompt Anatomy.

#### Ground-Truth Empirical Evidence
- Expert personas do not improve factual accuracy and reduce recall on knowledge benchmarks (Zheng et al. EMNLP 2024; Basil et al. 2025; E-018).
- Hardcoding search queries violates the agentic ReAct loop (Yao et al. 2023) and impairs adaptive search (E-019). Search quality tracks token budget, not pre-written strings (E-001).
- Negative bias constraints trigger ironic "white bear" rebound in transformers (E-017).
- Format restrictions impair reasoning capacity (Tam et al. EMNLP 2024; E-020).
- Single-turn front-loaded briefs dramatically outperform conversational drip-feeding (Laban et al. ICLR 2026; E-017).

#### Vivechak v3.0 Resolution & Operational Rule
Replace 8 XML sections with 5 functional blocks (`BRIEF`, `SCOPE`, `APPROACH`, `DELIVERABLE`, `FORMAT`). Be strictly prescriptive on WHAT, WHY, and BOUNDARIES; be strictly directional on HOW and SEARCH PATH. Use a required-coverage checklist instead of a mandatory output skeleton.

---

### D-007: Hybrid Output Architecture (Markdown + YAML Frontmatter)
```yaml
id: D-007
title: "Hybrid Markdown Body with JSON-Schema Validated YAML Frontmatter"
status: accepted
door_type: two-way
date: 2026-08-18
confidence: high
evidence_refs: [E-021, E-022, E-023]
informed_by_sessions: [T1-01, T2-07]
supersedes: null
review_trigger: "Re-evaluate if AI code generators natively adopt graph database retrieval over plain text Markdown"
tags: [output-architecture, markdown, yaml-frontmatter, json-schema, git-diff]
authored_by: "principal-research-architect"
human_reviewed: true
schema_version: "3.0"
```

#### Context & Problem Statement
Vivechak v2.0 required outputs as pure Markdown files without metadata schemas or machine-readable relationship graphs, relying on manual synthesis.

#### Evaluated Options
1. **Option 1 (v2.0 Dogma):** Pure Markdown with unstructured headers.
2. **Option 2:** Pure JSON or YAML files for entire research artifacts.
3. **Option 3:** Full RDF Knowledge Graph / Graph Database.
4. **Option 4 (v3.0 Resolution):** Hybrid Markdown Body + YAML Frontmatter with standardized 7 H2 sections.

#### Decision Outcome & Status
**Status:** `VALIDATED (With Frontmatter Extension)`  
**Chosen Option:** Option 4 — Hybrid Markdown Body + YAML Frontmatter.

#### Ground-Truth Empirical Evidence
- Markdown is unmatched for human scannability, line-based Git diffs, and LLM comprehension (E-021, E-023).
- Pure JSON/YAML fails for discursive narrative and Git merge workflows (E-023).
- Full graph databases represent premature, unversionable infrastructure for 10–30 documents (E-022).
- Adding a 12-field YAML frontmatter header provides machine-parseable metadata for automated synthesis without degrading Markdown's readability.

#### Vivechak v3.0 Resolution & Operational Rule
Enforce Hybrid Markdown + YAML Frontmatter validated via JSON Schema in CI. Standardize 7 required H2 sections with strict isolation between `Recommendation` and `Alternatives Considered` to prevent AI code-generation contamination. Generate `_index.yaml` build artifacts automatically.

---

### D-008: Commodity-Maximized Composition
```yaml
id: D-008
title: "Domain-Calibrated Wardley Evolution Composition Framework"
status: accepted
door_type: one-way
date: 2026-08-18
confidence: high
evidence_refs: [E-024, E-025, E-026]
informed_by_sessions: [T1-01, T2-04, SYN-01]
supersedes: null
review_trigger: "Re-evaluate if standard commodity primitives introduce unsustainable operational margin compression"
tags: [strategy, wardley-mapping, compose-vs-build, commodity-primitives]
authored_by: "principal-research-architect"
human_reviewed: true
schema_version: "3.0"
```

#### Context & Problem Statement
Vivechak v2.0 stated an arbitrary ~40% Compose / ~60% Build ratio across all software architectures.

#### Evaluated Options
1. **Option 1 (v2.0 Dogma):** Static universal ~40/60 compose/build ratio.
2. **Option 2:** 100% custom build of all infrastructure primitives.
3. **Option 3 (v3.0 Resolution):** Wardley Mapping Evolution Model (100% commodity compose, 100% proprietary custom build).

#### Decision Outcome & Status
**Status:** `REFINED`  
**Chosen Option:** Option 3 — Commodity-Maximized Composition Framework.

#### Ground-Truth Empirical Evidence
- A static 40/60 ratio is an ungrounded generalization; optimal composition varies from 80/20 in standard SaaS to 15/85 in novel algorithmic engines (E-024, E-025).
- Wardley Mapping dictates that commodity/utility components (Auth, DB, Storage, Queues, UI primitives) must be 100% composed from established providers (E-024).
- Custom engineering effort must be concentrated exclusively on proprietary domain intelligence, core state machines, and business logic (E-026).

#### Vivechak v3.0 Resolution & Operational Rule
Replace fixed percentage ratios with a domain-calibrated Wardley Evolution framework: compose all standard primitives; invest 100% of custom engineering bandwidth into proprietary business state machines and algorithms.

---

### D-009: Two-Track Phase 0 Exit Gate & Premortem
```yaml
id: D-009
title: "Risk-Calibrated Two-Track Phase 0 Exit Gate with Gary Klein Premortem"
status: accepted
door_type: one-way
date: 2026-08-18
confidence: high
evidence_refs: [E-027, E-028, E-029]
informed_by_sessions: [T1-01, T2-05, SYN-01]
supersedes: null
review_trigger: "Re-evaluate if Track A two-way door decisions result in >10% costly architectural rollbacks"
tags: [phase-0-gate, premortem, exit-checklist, risk-gating]
authored_by: "principal-research-architect"
human_reviewed: true
schema_version: "3.0"
```

#### Context & Problem Statement
Vivechak v2.0 enforced a rigid 9-step exit checklist before writing any application code, with zero exceptions.

#### Evaluated Options
1. **Option 1 (v2.0 Dogma):** Rigid uniform 9-step gate for all decisions without exception.
2. **Option 2:** Zero exit gates (immediate coding after ad-hoc research).
3. **Option 3 (v3.0 Resolution):** Risk-Calibrated Two-Track Gate (Track A Fast-Track for Two-Way Doors; Track B 9-Step Gate + Gary Klein Premortem for One-Way Doors).

#### Decision Outcome & Status
**Status:** `REFINED`  
**Chosen Option:** Option 3 — Two-Track Phase 0 Exit Gate.

#### Ground-Truth Empirical Evidence
- Uncalibrated rigid gates induce checklist fatigue and high override rates (clinical alerts show 49–96% override rates; E-027).
- Exploratory technical spikes (XP, Kent Beck) are vital pre-development research tools to probe Complex unknowns (E-029).
- Prospective hindsight (Gary Klein 1989 Premortem) reduces failure rates by 30% by surfacing hidden vulnerabilities before commitment (E-028).

#### Vivechak v3.0 Resolution & Operational Rule
Implement a Two-Track Phase 0 Gate: Track A fast-tracks reversible decisions on ~70% information; Track B enforces a rigorous 9-step exit gate (corroborated Grade A/B evidence, locked ADRs, Gary Klein Premortem protocol, and human architect sign-off) on irreversible One-Way Doors.

---

### D-010: 5-Layer Hybrid Generator Architecture
```yaml
id: D-010
title: "5-Layer Hybrid Deterministic/AI Generator Architecture"
status: accepted
door_type: one-way
date: 2026-08-18
confidence: high
evidence_refs: [E-030, E-031]
informed_by_sessions: [T1-03, T3-01]
supersedes: null
review_trigger: "Re-evaluate if reasoning LLMs achieve 100% deterministic multi-instruction schema adherence on monolithic 5,000+ word outputs"
tags: [generator, hybrid-architecture, archetypes, yeoman-blend, prompt-synthesis]
authored_by: "principal-research-architect"
human_reviewed: true
schema_version: "3.0"
```

#### Context & Problem Statement
Vivechak v2.0 relied on a monolithic single-turn master prompt (`META-PROMPT-GENERATOR.md`) to generate an entire project pipeline in one shot.

#### Evaluated Options
1. **Option 1 (v2.0 Dogma):** Monolithic master meta-prompt generating entire pipeline at once.
2. **Option 2:** Pure static template library with parameter substitution (no AI).
3. **Option 3 (v3.0 Resolution):** 5-Layer Hybrid Architecture (Deterministic rules for scaffolding + Scoped parallel AI calls for prose + Validation gates).

#### Decision Outcome & Status
**Status:** `REFUTED`  
**Chosen Option:** Option 3 — 5-Layer Hybrid Generator Architecture.

#### Ground-Truth Empirical Evidence
- Monolithic master prompts suffer from non-determinism, inability to guarantee schema conformance, prompt drift, and decomposition failure when simultaneously planning, writing, and structuring (E-031).
- Pure static templates cannot synthesize domain-specific, non-generic prompts from free-text vision (E-030).
- Proven code generator ecosystems (Nx, Yeoman, Plop, cookiecutter) prove that scaffolding must be deterministic while prose synthesis must be scoped (E-030).

#### Vivechak v3.0 Resolution & Operational Rule
Build a 5-Layer Hybrid Generator: Layers 0–2 (Deterministic rules for input schema, 6-archetype classifier, complexity scoring, and skeleton assembly with Yeoman `composeWith` blending); Layers 3–4 (Scoped parallel AI calls for prompt prose and seeded hypotheses); Layer 5 (Structural schema and budget validation gates in CI).

---

## Final Verification & Governance

This Decision Registry represents the immutable historical record of the empirical validation of the Vivechak. Any future amendments or supersessions must follow the Vivechak v3.0 ADR governance protocol by issuing new `D-NNN` records with explicit `amends` or `supersedes` pointers.

*Sealed and Verified as Vivechak v3.0 Master Decision Registry.*
