# The Vivechak Meta-Framework
### Complete Specification for Evidence-Grounded Pre-Development Research
*Version 1.1 — Empirically Validated via 11 Meta-Research Sessions*

---

## 1. Executive Philosophy

> **The Research-First Law:**
> *Before beginning development, investigate every irreversible architectural decision with graded, verifiable evidence. Develop nothing based on cached assumptions, outdated training data, or hallucinated conclusions.*

Technical failures and architectural rewrites are rarely caused by poor implementation — they are caused by **premature decisions made on unverified assumptions.** The Vivechak is a structured pre-development engine that discovers, evaluates, resolves, and synthesizes the technical foundation of any project before committing to irreversible implementation decisions.

### How the Framework Operates

```
PROJECT VISION
    │
    ▼
┌─────────────────────────────────────────────────────┐
│  COMPLEXITY SCORING (8 dimensions, 0–24 pts)        │
│  → Determines research depth tier (0–3)             │
│  → Routes each decision by reversibility            │
└───────────────────────┬─────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────┐
│  CONSTRAINED DAG EXECUTION                          │
│  Sessions run when dependencies are satisfied       │
│  Default: unblocked. Hard deps block; soft don't.   │
│  Adaptive checkpoints expand/contract scope         │
└───────────────────────┬─────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────┐
│  STAGED TRIANGULATION & EVIDENCE GRADING            │
│  Single-model default → critique probe →            │
│  full triangulation (One-Way Doors only)            │
│  A–E grades + modifiers + verification tracking     │
└───────────────────────┬─────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────┐
│  TWO-TRACK PHASE 0 GATE                             │
│  Track A: Fast-track for Two-Way Doors              │
│  Track B: 9-step gate + premortem for One-Way Doors │
└───────────────────────┬─────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────┐
│  FOUNDING ARCHITECTURE DOCUMENT (FAD)               │
│  Map-Reduce synthesis → sealed specification        │
│  → Repository scaffolding begins                    │
└─────────────────────────────────────────────────────┘
```

---

## 2. Core Principles

Vivechak v1.1 is governed by 8 evidence-grounded principles. Each was empirically validated through the meta-research pipeline and cites its supporting evidence.

### P1: The Context Architecture Law
*Supersedes: pre-validation Aspect-Isolation Law*

Research quality is governed by **attention budget, context purity, and task boundaries** — not by arbitrary session counts. The constraint is real (token/search budget is the dominant predictor of performance on web research benchmarks — OpenAI BrowseComp scaling analysis), but the remedy is conditional decomposition, not blanket fragmentation.

**Rules:**
- Decompose into dedicated research contexts when sub-tasks have low interdependency, high individual complexity, or divergent search spaces.
- Integrate coupled topics into structured joint sessions when evaluating holistic system tradeoffs.
- Every decomposed investigation **must** conclude with an explicit downstream synthesis pass.
- Over-isolation triggers the split-attention effect, introduces 4–15× token overhead, and misses systemic cross-cutting trade-offs.

> **Evidence:** E-001 (OpenAI BrowseComp — accuracy scales with search budget/compute), E-002 (MTI +12.4% on joint tasks), E-003 (Chandler & Sweller split-attention), E-004 (Chroma context rot)

### P2: Reversibility-Calibrated Rigor

The depth of research, evidentiary burden, and review overhead allocated to a decision must scale directly with its **reversibility and blast radius**.

| Door Type | Definition | Research Depth | Evidence Bar |
|---|---|---|---|
| **One-Way (Type 1)** | Consequential, costly/impossible to reverse (primary datastore, data model, auth architecture, core language/runtime, regulatory compliance, public API contracts, hosting infrastructure, deployment architecture) | Deep research, multi-session, corroborated evidence | Grade A/B, human review, premortem |
| **Two-Way (Type 2)** | Cheap, fast to reverse (UI framework, styling, CI tooling, non-core utility libraries) | Fast spike or convention decision on ~70% information | Grade B/C acceptable, no gate required |

> **Evidence:** E-010 (Cynefin), E-011 (DORA defect concentration), E-012 (Amazon Type 1/2 doors)

### P3: Evidentiary Grounding & Verification Provenance

No technical assertion or architectural decision may be accepted without:
- An explicit **evidence grade** (A through E)
- **Provenance tracking** (`verification_method`: fetched, cached, recalled, secondhand, human-provided)
- **Contextual modifiers** (corroboration, recency, directness)

**Hard rule:** Unverified AI parametric recall is capped at Grade D. Zero recalled citations may support One-Way Door decisions.

> **Evidence:** E-013 (GRADE framework), E-014 (Admiralty Code collapse), E-015 (AI citation confabulation)

### P4: Prescriptive Scope, Dynamic Method
*Supersedes: pre-validation 8-section XML prompt anatomy*

Research briefs must be:
- **Prescriptive** on WHAT to investigate, WHY it matters, WHAT boundaries apply, and WHAT coverage is required.
- **Directional** on HOW to execute: no pre-scripted search queries, no artificial search counts, no expert role-playing personas, no rigid output skeletons.

**Bounded Exploration Mandate:** Prescriptive scope defines the **minimum coverage floor**, not a ceiling. If research reveals critical concerns, risks, dependencies, or opportunities beyond the stated scope that are material to the decision being informed, the researching agent should investigate and include them — justified with evidence. This prevents the "hyper-literalism" effect where frontier models constrain their native reasoning to only the explicitly listed elements, missing emergent or adjacent concerns the prompt author could not have anticipated.

Front-load the complete brief in a single turn. Never drip-feed instructions across multiple turns (39% performance drop documented).

> **Evidence:** E-017 (Laban et al. ICLR 2026, 39% multi-turn drop), E-018 (persona debunking), E-019 (ReAct paradigm), E-020 (format restriction penalty)

**Audience framing vs. persona assignment:** Specifying the target audience (e.g., "a Principal Architect needing production-grade tradeoffs") sets the QUALITY BAR for the output. This is distinct from persona assignment (e.g., "You are an expert database engineer"), which attempts to alter the model's reasoning behavior. Audience framing is prescriptive scope (WHAT quality level); persona assignment is method prescription (HOW to think). The former is permitted under P4; the latter is empirically refuted.

### P5: Commodity-Maximized Composition
*Supersedes: pre-validation fixed ~40/60 ratio*

- **Compose 100%** of commodity/utility components from battle-tested providers (Auth, DB, Storage, Queues, UI primitives, CI/CD).
- **Build 100% custom** only for proprietary domain intelligence, core state machines, differentiated business logic, and unique algorithms.

The ratio varies by domain. Use Wardley evolution mapping, not fixed percentages.

> **Evidence:** E-024 (Wardley mapping), E-025 (Fowler MonolithFirst), E-026 (DORA loosely-coupled architecture)

### P6: Dual-Audience Artifact Architecture

Research outputs must serve both human readers and machine consumers:
- **Hybrid Markdown + YAML frontmatter** — human-readable body with machine-parseable metadata.
- **Standardized 7-section H2 skeleton** — consistent structure enabling automated Map-Reduce synthesis.
- **Strict separation** between `Recommendation` and `Alternatives Considered` — prevents AI code-generation contamination from rejected options.

> **Evidence:** E-021 (MADR 4.0), E-022 (GraphRAG header chunking), E-023 (Git diff mechanics)

### P7: Staged Triangulation & Diagnostic Disagreement
*Supersedes: pre-validation mandatory 3-model triangulation*

Multi-model consensus is an **escalation tool, not a mandatory ritual**.

| Stage | When | Method |
|---|---|---|
| **Default** | All sessions | Single-model deep research |
| **Critique Probe** | Medium-stakes decisions | Feed Pass 1 output to a second model for adversarial critique |
| **Full Triangulation** | One-Way Doors + Novel Domain + Genuine Contestation | Run identical prompt across 2–3 models; synthesize divergence via ACH matrix |

Disagreement is treated as a **diagnostic signal** of problem ambiguity, not a vote to average out.

> **Evidence:** E-005 (Kim et al. ICML 2025, 60% correlated errors), E-006 (Gao & Xiao, model house styles), E-007 (Lorenz et al. crowds breakdown)

### P8: Structured Falsification & Diagnostic Review

Architectural analysis must prioritize **falsification over confirmation**.
- Research prompts must explicitly seek disconfirming evidence against favored options.
- Decision records must document rejected alternatives with causal rationale.
- Every locked ADR must carry a `review_trigger` — explicit conditions or dates for mandatory re-evaluation.
- One-Way Doors require a **Gary Klein Premortem** before commitment.
- Every One-Way Door ADR should carry a `review_date` set BEFORE the outcome is known, and an optional `prediction` field recording the expected outcome. Post-project retrospectives compare predictions to actuals — without this closing loop, decision calibration can never improve over time.

> **Evidence:** E-028 (Mitchell, Russo & Pennington 1989 prospective hindsight, 30% risk identification improvement; popularized as the Premortem by Klein 2007), Heuer's ACH, Annie Duke decision journaling

---

## 3. Pipeline Architecture

### 3.1 Topology: Constrained DAG with Adaptive Checkpoints

Vivechak v1.1 replaces rigid stage gating with a **Directed Acyclic Graph** governed by explicit information dependencies.

#### The Inverted Dependency Default
Every research session defaults to **unblocked** (eligible to execute immediately) unless an explicit hard information dependency is declared.

#### Dependency Classification

| Type | Definition | Effect |
|---|---|---|
| **Hard (Information)** | Session B literally cannot formulate valid outputs without an artifact from Session A | Blocks execution — DAG edge |
| **Soft (Contextual)** | Session B benefits from terminology or constraints from Session A | No blocking — inject a Shared Context Brief (a concise summary of relevant terminology, constraints, and preliminary findings from upstream sessions, included in the prompt's BRIEF block) |


#### Dependency Injection Protocol

When injecting upstream session outputs as context for downstream sessions:

1. **Extract only the specific information** declared in the dependency specification — typically 3–5 sentences summarizing the relevant decision, finding, or constraint.
2. **Do NOT attach full research artifacts** as context. A 10,000-word research report should be condensed to the specific findings the downstream session needs.
3. **Label the dependency type** in the downstream prompt:
   - **Constrains:** The upstream decision must be respected (e.g., "Database: PostgreSQL 16 — this is a locked One-Way Door decision")
   - **Informs:** The upstream finding provides context but should not limit independent evaluation (e.g., "Prior research found latency under 50ms at current scale — verify independently")
4. **Minimizing dependencies maximizes parallelism.** Most Layer 0 and many Layer 1 sessions can run concurrently. Reserve hard dependencies for genuine information prerequisites.
#### Adaptive Checkpoints
- **Expansion:** If a session reveals unexpected Cynefin-Complex trade-offs, spawn a maximum of 2 child sub-sessions.
- **Contraction:** If a planned session's core question has been authoritatively resolved upstream, close it immediately.
- **Incremental Synthesis:** Partial FAD compilation begins as soon as any structural branch closes.

#### Pipeline Termination Criteria
A research pipeline may be terminated early when ALL of the following are met:
- Every identified One-Way Door decision has a locked ADR with Grade A/B evidence.
- No remaining unexecuted sessions would inform an unresolved One-Way Door.
- All executed sessions have produced actionable recommendations (not vague summaries).
- No outstanding `contested` corroboration flags remain on critical claims.

Termination does NOT require executing every planned session. Two-Way Door sessions that remain unexecuted at termination are decided by convention and logged as Track A decisions.

#### Pipeline Failure Modes

When a research pipeline encounters problems, apply these recovery protocols:

**Session produces garbage (vague summaries, hallucinated citations, off-topic output):**
1. Check the prompt — is the BRIEF clear on WHAT to learn and WHY it matters?
2. Try a different model (Gemini → Claude → ChatGPT). Model strengths vary by domain.
3. If the topic is too broad, split into 2 focused sub-sessions.
4. If the model lacks web access, the output will be parametric recall at best — cap at Grade D and note `verification_method: recalled`.

**Synthesis deadlocks (sessions produce contradictory recommendations with equal evidence):**
1. This is expected, not a failure. Use `templates/CONFLICT-RESOLUTION.template.md` — the ACH matrix is designed for exactly this.
2. Check whether the contradiction is real (different evidence) or apparent (different scoping assumptions).
3. If evidence is genuinely equal, classify by reversibility: Two-Way Door → pick either and spike; One-Way Door → escalate to triangulation (run the same prompt on a second model).
4. If deadlock persists after triangulation, the decision may be in the Cynefin Complex domain. Consider a technical spike instead of more research.

**Gate reveals gaps (Phase 0 checklist fails on specific items):**
1. Identify which One-Way Doors lack adequate evidence.
2. Spawn targeted follow-up sessions for those specific gaps — do NOT re-run the entire pipeline.
3. If gaps are in Two-Way Door decisions, decide by convention (Track A) and move forward.
4. If gaps indicate the project vision itself is unclear, pause and clarify scope before spawning more sessions.

#### Progressive Elaboration Protocol

For Tier 2–3 projects where research may span weeks, the project vision may evolve mid-pipeline. When new constraints, pivots, or stakeholder feedback emerge:

1. **Update the RESEARCH-PIPELINE.md** with a dated amendment section noting the change.
2. **Assess impact on existing sessions:** If completed sessions are invalidated, mark their ADRs as deprecated with superseded_by linking to new sessions.
3. **Add or modify sessions** in the DAG to address new concerns. New sessions follow the same tier-appropriate rigor.
4. **Do NOT restart the pipeline from scratch** unless the fundamental project vision has changed (e.g., pivot from B2B to B2C). Incremental amendment preserves completed research investment.

This protocol balances research continuity with responsiveness to change — a core tenet of evidence-grounded decision-making.

### 3.2 Adaptive Scaling Model

#### 8-Dimension Complexity Scoring (0–24 Points)

| Dimension | 0 (Low) | 1 (Moderate) | 2 (Substantial) | 3 (Extreme) |
|---|---|---|---|---|
| **Domain Novelty** | Standard pattern (e.g., CRUD app) | Established category (e.g., B2B SaaS, standard lab protocol) | Unconventional workflow | New category/paradigm |
| **Technical Novelty** | Team's standard stack | Familiar platform, new component | New framework/paradigm | Unproven infrastructure |
| **Regulatory Exposure** | Zero sensitive data | Internal data | PII, GDPR, SOC 2 | HIPAA, PCI-DSS, KYC/AML |
| **Reversibility** | Throwaway/rewritable | Modularly swappable | Core schema/multi-tenant | Deep platform infra |
| **Investment Horizon** | Hackathon/weekend | Lean MVP/internal tool | Funded venture | Enterprise mission-critical |
| **Coordination Complexity** | Single decision-maker | Small team (2–4) | Mid-size (5–12) | Multi-team (>12) |
| **Expected Longevity** | Days to weeks | Months (prototype) | 1–3 years (product) | 5+ years (platform) |
| **Integration Complexity** | Standalone | 1–2 external interfaces (APIs, protocols, feeds) | Multiple complex integrations | Regulated rails/legacy systems |

#### Tier Mapping

| Score | Tier | Session Budget | Character |
|---|---|---|---|
| **0–4** | Tier 0: Minimal | 1–3 sessions | Fast spike on core unknown |
| **5–9** | Tier 1: Light | 4–8 sessions | Standard SaaS/internal tools |
| **10–15** | Tier 2: Standard | 9–16 sessions | Commercial product with core ADRs |
| **16–24** | Tier 3: Deep | 17–30 sessions | Novel, regulated, multi-tenant platform |

**What constitutes a "session":** A session is one independent research execution — a single prompt sent to a frontier AI model's deep research mode, producing one complete output artifact. Sessions are calibrated for deep research modes that perform 20–100+ web searches per run. If using standard (non-deep-research) chat, a single "session" may require multiple conversational turns to achieve equivalent depth — adjust session budgets accordingly. The tier budgets are relative proportions between tiers, not absolute figures to copy without calibration for your specific tools and workflow.

> **🚨 Hard Override:** If Regulatory Exposure = 3, all intersecting decisions automatically receive Tier 3 treatment regardless of total score.

**Rubric calibration:** The 0–3 scales and tier boundaries are structured heuristics, not calibrated measurement instruments. After completing 2–3 projects, compare the assigned tier to the actual research effort required. If projects consistently need more depth than their tier suggests, adjust boundaries upward. If sessions are routinely closed early because the question was already answered, adjust downward. The rubric should evolve with your team's experience — treat it as a living instrument, not a fixed table.

#### Per-Decision Routing Matrix

| | **Known Pattern** | **Unknown / Novel** |
|---|---|---|
| **Reversible (Two-Way Door)** | **SKIP** — decide by convention | **FAST SPIKE** — 1 timeboxed session |
| **Irreversible (One-Way Door)** | **CONFIRM & COMMIT** — 1 session + ADR | **DEEP RESEARCH** — 2–5 sessions + ADR + premortem |

### 3.3 Staged Triangulation Protocol

| Stage | Trigger | Method | Cost |
|---|---|---|---|
| **Default** | All sessions | Single-model deep research | 1× |
| **Critique Probe** | Medium-stakes decisions | Feed Pass 1 synthesis to Model B for adversarial critique | 1.3× |
| **Full Triangulation** | One-Way Doors + Novel Domain + Genuine Contestation | Run identical prompt on 2–3 models; synthesize divergence via ACH matrix | 3× |

---

## 4. Research Prompt Design

### 4.1 The 5-Block Prompt Anatomy

| Block | Purpose | Prescriptive or Directional? |
|---|---|---|
| **BRIEF** | Goal, target deliverable, audience, decision being informed, required depth | **Prescriptive** — what and why |
| **SCOPE** | Boundaries, time window, in/out scope, source priorities, date anchor | **Prescriptive** — boundaries |
| **APPROACH** | Exploration strategy, effort calibration, epistemic discipline | **Directional** — guide how, don't prescribe |
| **DELIVERABLE** | Required-coverage checklist, comparison parameters, evidence grading | **Prescriptive** — what to cover |
| **FORMAT** | Markdown structure, frontmatter schema, artifact delivery | **Prescriptive** — output format |

#### What Was Removed and Why

| Pre-validation Element | Verdict | Evidence |
|---|---|---|
| `<system>` expert persona | **Removed** — personas don't improve factual accuracy | Zheng et al. EMNLP 2024; Basil et al. 2025 |
| `<web_searches>` hardcoded queries | **Removed** — violates agentic ReAct loop | Yao et al. 2023; OpenAI BrowseComp scaling analysis |
| `<bias_resistance>` negative instructions | **Removed** — triggers ironic "white bear" rebound | Documented in T2-06 |
| `<output_spec>` rigid sections | **Replaced** with coverage checklist | Tam et al. EMNLP 2024 |
| Minimum search counts | **Removed** — artificial floors don't match model judgment | T1-02 |
| Multi-turn drip-feeding | **Banned** — front-load everything in one turn | Laban et al. ICLR 2026 (39% drop) |

### 4.2 Standard Prompt Template

```markdown
# RESEARCH BRIEF: [Topic Title]

## BRIEF
We are investigating [core technical/architectural question].
This research will directly inform Architectural Decision [D-NNN: Title].
The target audience is a Principal Architect requiring rigorous, production-grade
technical evaluation with concrete tradeoffs, operational failure modes, and
verified benchmarks — not high-level introductory summaries.

## SCOPE
- **Temporal Anchor:** Today's date is [YYYY-MM-DD]. Focus on developments within
  the last 18–24 months. Flag findings older than [date] as potentially stale.
- **In Scope:** [Explicit boundary 1], [boundary 2], [boundary 3].
- **Out of Scope:** [Explicit exclusion 1], [exclusion 2].
- **Source Priorities:** Prioritize primary sources (official docs, RFCs, source
  code, peer-reviewed benchmarks) over secondary aggregators and SEO content.

## APPROACH
- Begin with broad landscape queries, then dynamically formulate targeted queries
  to investigate specific trade-offs, failure modes, and benchmarks.
- Scale search effort to discovered complexity. Trace genuine technical
  contradictions rather than stopping at the first consensus hit.
- State assumptions explicitly. Surface disagreements rather than smoothing them.
  Frame inquiries neutrally; actively search for disconfirming evidence.
- If your research reveals critical concerns, dependencies, risks, or opportunities
  not listed in the coverage checklist, investigate and include them. The stated
  scope defines the minimum — not the maximum — of what this session should cover.
  Justify any scope expansion with evidence.

## DELIVERABLE
Deliver a structured Markdown document covering:
1. Executive Summary & Recommendation
2. Options Evaluation Matrix — if comparing 2+ options, include a weighted
   scoring matrix per the Weighted Evaluation Protocol (see section 6.4) (project-derived
   criteria, 1-5 scores with evidence refs, sensitivity check)
3. Deep Technical Analysis of top 2–3 contenders
4. Inline evidence grades (A–E with modifiers and verification method)
5. Open Risks & Reversal Triggers
6. Discovered Concerns (if research revealed material concerns beyond the
   stated scope, include a dedicated section — omit if nothing emerged)

## FORMAT
Deliver as a single, complete Markdown file artifact with YAML frontmatter
per the Vivechak v1.1 session schema.
**Filename:** `[Session-ID]-[slug].md` (e.g., `T1-01-primary-datastore-selection.md`)
```

### 4.3 Research Output Quality Rubric

A research session output meets Vivechak's quality bar when it satisfies ALL of the following:

| Criterion | Minimum Standard |
|---|---|
| **Evidence Density** | ≥3 distinct claims backed by Grade A or B evidence |
| **Options Evaluated** | ≥2 competing options analyzed for comparison sessions |
| **Failure Modes** | ≥1 concrete failure scenario identified per recommended option |
| **Reversal Triggers** | Explicit conditions for when to re-evaluate the recommendation |
| **Source Diversity** | Not all evidence from a single vendor or source |
| **Disconfirming Evidence** | At least one actively sought disconfirming argument documented |

Sessions failing this rubric should be re-executed with a refined prompt before their findings inform ADRs.

---

## 5. Evidence & Decision System

### 5.1 Evidentiary Tiers (A through E)

| Grade | Source Classification | Starting Confidence |
|---|---|---|
| **A** | **Primary / Authoritative** — Official source code, published RFCs, formal API specs, peer-reviewed studies | Very High |
| **B** | **Empirical / Experimental** — Independently reproducible benchmarks, engineering postmortems, controlled experiments | High |
| **C** | **Vendor / Motivated** — Vendor whitepapers, commercial comparisons, marketing benchmarks | Moderate |
| **D** | **Secondary / Opinion** — Unofficial tutorials, blog posts, forums, AI parametric recall without verification | Low-to-Moderate |
| **E** | **Untraceable / Speculative** — Unverifiable claims, confabulations, unattributed speculation | Zero |

**Grade E is "unverifiable," not "old."** Stale but real sources keep their actual grade with a `stale` recency modifier.

### 5.2 Contextual Modifiers

| Modifier | Values | Purpose |
|---|---|---|
| **Corroboration** | `single` · `corroborated` (≥2 independent sources) · `contested` (direct contradiction) | How many sources agree? |
| **Recency** | `fresh` (within domain half-life) · `aging` (approaching half-life) · `stale` (past half-life) | Is this current? |
| **Directness** | `direct` (exact workload/target) · `indirect` (analogical evidence) | Does this apply to our case? |
**Domain Half-Life Guidance:** The threshold between `fresh` and `aging` varies by domain:
- Cloud pricing and provider features: ~6 months
- Framework/library APIs and benchmarks: ~12 months
- Security advisories and CVEs: ~3 months
- Database internals and query optimizers: ~18 months
- Regulatory requirements (GDPR, HIPAA, PCI): ~24 months
- Foundational CS/architecture principles: ~5+ years

When in doubt, treat evidence as `aging` rather than `fresh`.

### 5.3 Verification Method

| Method | Definition | Grade Cap |
|---|---|---|
| `fetched` | Retrieved and validated live via tool/web call during this session | None |
| `cached` | Read from local repository file or previously verified artifact | None |
| `recalled` | Generated from model parametric memory without live verification | **Capped at Grade D** |
| `secondhand` | Sourced from an article summarizing a primary source | None |
| `human-provided` | Injected directly by a human engineer | None |

### 5.4 Hard Rules

1. **Recalled claims capped at Grade D** regardless of apparent source quality.
2. **Zero recalled citations may underpin One-Way Door decisions.** All One-Way Door decisions require `fetched` or `cached` verification.
3. **Contested claims must be surfaced, not resolved by majority.** Document the contradiction and causal resolution reasoning.

### 5.5 Composite Citation Format

```
Claim text (Grade [A-E] · [corroboration] · [recency] · [directness] | [verification_method])
```

Example:
```
PostgreSQL 16 logical replication supports bi-directional failover.
(Grade B · corroborated · fresh · direct | fetched)
```

### 5.6 Decision Confidence (Decoupled from Evidence Grade)

| Confidence | Definition |
|---|---|
| `high` | Multiple corroborated Grade A/B sources; alternatives thoroughly evaluated |
| `medium` | Adequate evidence but some alternatives lack evaluation, or key claims rest on single sources |
| `low` | Decision under uncertainty; thin, contested, or indirect evidence |

### 5.7 Architecture Decision Records (D-NNN)

Every architectural decision is documented as a YAML-frontmattered Markdown file:

```yaml
---
id: D-001
title: "Primary Datastore Selection"
status: accepted          # proposed | accepted | rejected | deprecated | superseded
door_type: one-way        # Sets evidentiary bar
date: 2026-08-18
confidence: high          # Decoupled from evidence grade
evidence_refs: [E-012, E-047]
informed_by_sessions: [T2-03, T2-05]
supersedes: null
superseded_by: null
amends: null
review_trigger: "Re-evaluate if ingestion exceeds 50k ops/sec or storage >2TB"
review_date: 2027-02-18   # Set BEFORE outcome is known
prediction: "PostgreSQL will handle projected load for 18+ months"
tags: [datastore, infrastructure]
authored_by: "agent-vivechak"
human_reviewed: true      # Mandatory for one-way doors
schema_version: "1.1"
---
```

**Required body sections:**
1. Context & Problem Statement
2. Evaluated Options (with evidence references)
3. Decision Outcome & Rationale
4. Rejected Alternatives & Tradeoffs
5. Failure Modes & Reversal Triggers

---

## 6. Output Architecture

### 6.1 Hybrid Markdown + YAML Frontmatter

Every research artifact combines:
- **YAML frontmatter** — machine-parseable metadata
- **Markdown body** — human-readable narrative with standardized section structure
- **Strict separation** between `Recommendation` and `Alternatives Considered`

### 6.2 Standardized 7-Section Body Skeleton

| # | Section | Purpose |
|---|---|---|
| 1 | `## Research Question` | Precise 1–2 sentence statement |
| 2 | `## Key Findings` | 3–7 atomic bullet points extractable by synthesis agents |
| 3 | `## Recommendation` | Explicit guidance (isolated from rejected options) |
| 4 | `## Alternatives Considered` | Evaluated competing options and rejection rationale |
| 5 | `## Detailed Findings` | Flexible analytical body |
| 6 | `## Open Questions & Risks` | Unresolved items and downstream risks |
| 7 | `## Sources & Evidentiary Ledger` | Numbered citations with composite grade metadata |

### 6.3 Map-Reduce Synthesis Workflow

When synthesizing research sessions into the Founding Architecture Document:

```yaml
---
id: SYN-01
title: "[Project Name] — Founding Architecture Document"
synthesis_date: YYYY-MM-DD
status: draft                  # draft | sealed
research_sessions_ingested: 0   # Count of final sessions ingested
decisions_locked: 0             # Count of accepted ADRs
open_questions: 0               # Count of unresolved questions
schema_version: "1.0"
---
```

1. **Filter:** Ingest only `status: final` sessions
2. **Group:** Cluster by `topic` tags
3. **Map:** Extract atomic Key Findings and Recommendation sections
4. **Reduce:** Merge topic clusters into unified subsystem chapters
5. **Reconcile:** Resolve cross-session contradictions using confidence + date weighting
6. **Trace:** Annotate all FAD blocks with originating session and ADR IDs
7. **Gate:** Subject to Track B Phase 0 Gate before committing

#### 6.3.1 FAD Versioning and Post-Seal Updates

Once sealed, a FAD is immutable. When a `review_trigger` fires or new evidence invalidates a locked ADR:

1. **Open a new research session** targeting the specific decision under review.
2. **Record a new ADR** (D-NNN+1) with `supersedes: D-NNN` linking to the original.
3. **Issue a FAD Amendment** (SYN-01.1, SYN-01.2, etc.) documenting only the changed sections with full traceability.
4. **Re-run the Phase 0 Gate** for the amended decision only.

Do NOT modify the original sealed FAD. Amendments are additive documents that reference the sealed original.

### 6.4 Weighted Evaluation Protocol (Comparison Sessions)

When a research session involves comparing multiple options (e.g., database selection, framework evaluation, vendor comparison), the qualitative analysis must be complemented by a **quantitative weighted scoring matrix** to counteract narrative volume bias — the systematic tendency of AI models to favor options with more training data, blog coverage, and community advocacy regardless of project-specific fit.

#### Why This Matters

Qualitative-only comparison is vulnerable to three documented biases:
- **Narrative volume bias:** Technologies with larger communities produce more positive text, creating an illusion of superiority that reflects popularity, not fitness.
- **Verbosity bias:** Models prefer longer, more detailed responses — well-documented options get richer analysis, which reads as stronger evidence.
- **Vendor marketing contamination:** Well-funded products produce Grade C evidence (vendor claims) at scale, drowning out Grade A/B evidence for alternatives.

Weighted scoring forces the model to evaluate each option against explicit, project-derived criteria rather than relying on narrative coherence.

#### Protocol

For any session that evaluates 2+ competing options:

1. **Derive Criteria from Project Context:** Extract 5–8 evaluation criteria directly from the project vision and the decision being informed. Do not use generic criteria — every criterion must trace to a specific project requirement or constraint.

2. **Assign Weights (must sum to 1.0):** Weight each criterion by its importance to this specific project. Justify each weight with a one-line rationale. Example:

   | Criterion | Weight | Rationale |
   |---|---|---|
   | Write throughput at scale | 0.25 | Core use case: high-volume event ingestion |
   | Operational complexity | 0.20 | Solo developer, no dedicated DevOps |
   | Query flexibility | 0.20 | Ad-hoc analytics required by product spec |
   | Ecosystem maturity | 0.15 | Long-term maintenance, not greenfield research |
   | Cost at projected scale | 0.10 | Bootstrap budget, cost-sensitive first 18 months |
   | Migration path | 0.10 | Two-way door if wrong, but migration still has cost |

3. **Score Each Option (1–5 scale):** Score every option against every criterion using this scale:
   - **5:** Exceptional fit, clear leader on this criterion (with evidence)
   - **4:** Strong fit, minor gaps
   - **3:** Adequate, meets minimum requirements
   - **2:** Weak, notable gaps or concerns
   - **1:** Poor fit, significant risk or missing capability

   Each score must reference the specific evidence (E-NNN or inline citation) that justifies it. Scores without evidence references are invalid.

4. **Compute Weighted Totals:** Multiply each score by its weight and sum across criteria.

5. **Sensitivity Check:** Identify the top 2 criteria by weight. Re-run the calculation with those weights shifted ±20%. If the ranking changes, flag the recommendation as **weight-sensitive** and note which criteria drive the instability.

6. **Synthesize:** The final recommendation must consider BOTH the weighted score AND the qualitative analysis. The score is a **bias-correction lens**, not the sole decision input. If qualitative analysis and scoring disagree, explicitly address the divergence and explain which signal the recommendation follows and why.

#### Important Constraints
- This protocol applies ONLY to comparison sessions. Landscape scans, single-technology deep dives, and blueprint sessions do not require scoring.
- The weighted score is a complement to qualitative analysis, never a replacement. Architectural decisions involve judgment dimensions (team culture, strategic direction, ecosystem trajectory) that resist quantification.
- Criteria selection itself can introduce bias. The model must justify WHY these criteria and not others.

---

## 7. Phase 0 Exit Gate

### 7.1 The Two-Track Standard

| Track | Applies To | Requirements | Speed |
|---|---|---|---|
| **Track A: Fast-Track** | Two-Way Door decisions | Decision logged, reversibility confirmed, single corroborated source | Minutes |
| **Track B: Rigorous Gate** | One-Way Door decisions | Full 9-step checklist below + Gary Klein Premortem | Hours–Days |

### 7.2 Track B: 9-Step Pre-Codebase Exit Checklist

1. **DAG Closure** — All required dependency paths terminated in `status: final`
2. **Contradiction Resolution** — All divergences explicitly resolved
3. **Evidentiary Threshold** — Zero uncorroborated Grade C/D/E claims underpin irreversible pillars
4. **Verification Integrity** — 100% of critical citations carry `fetched` or `cached`
5. **Rejected Alternatives Documented** — Every ADR includes evaluated and rejected options
6. **Decay Triggers Assigned** — Every ADR contains explicit `review_trigger`
7. **Premortem Protocol** — Gary Klein prospective hindsight: *"Assume catastrophic failure in 12 months. What caused it?"*
8. **Human Review** — Named Principal Architect signature on all One-Way Door ADRs
9. **FAD Sealed** — Founding Architecture Document compiled, committed, ready for scaffolding

---

## 8. Generator Architecture

The pipeline generator uses a **deterministic/AI hybrid** architecture that produces two documents: a unified pipeline containing the execution plan and copy-paste-ready research prompts, plus a separate decision registry that evolves during execution. See [GENERATOR.md](GENERATOR.md) for the working generator prompt.

The long-term vision is a 5-layer code-based tool:

| Layer | Function | Implementation |
|---|---|---|
| **0: Input** | Collect project vision via open-ended description | AI — parameter extraction from natural language |
| **1: Classification** | 6-archetype domain classifier + 8-dimension complexity scoring | Deterministic — rule-based decision tree |
| **2: Skeleton** | Assemble session matrix from archetype templates + tier budget | Deterministic — template composition |
| **3: Prompt Synthesis** | Generate project-specific prompt prose for each session slot | AI — N parallel scoped calls |
| **4: Registry Seeding** | Generate project-specific D-NNN hypotheses | AI — scoped calls tied to vision + constraints |
| **5: Validation** | Schema validation, archetype completeness, budget verification | Deterministic — CI-style gates |

### The 6 Domain Archetypes

| Archetype | Unique Research Needs |
|---|---|
| **B2B SaaS** | Multi-tenancy, RBAC/SAML, CRM integrations, seat/usage pricing |
| **Developer Tools** | DX & time-to-first-value, CLI/SDK idioms, open-source licensing |
| **FinTech** | KYC/AML, banking rails, immutable ledgers, PCI-DSS, fraud detection |
| **AI/ML Systems** | Model selection, eval benchmarks, inference cost, context architecture |
| **Consumer Mobile** | App Store compliance, offline-first sync, push/retention, in-app billing |
| **Real-Time / IoT** | MQTT/WebSockets, edge vs cloud compute, fleet OTA, hardware constraints |

#### Beyond Software: Domain Adaptation

Vivechak's epistemic core — evidence grading (Cochrane/GRADE lineage), ACH conflict resolution (CIA/Heuer), premortem protocol (Gary Klein), reversibility routing (Bezos Type 1/2 doors), and map-reduce synthesis — was not invented for software. These primitives govern how intelligence reasons about truth, risk, and irreversibility in *any* domain.

The generator prompt already uses "technical project" rather than "software project." The 4 operational templates (Decisions, Conflict Resolution, FAD, Phase 0 Gate) contain zero software-specific content. Non-software domains have been successfully researched using adapted versions of the generator.

**To adapt for a non-software domain:**
1. The 8 core principles, evidence grading, and all templates work without modification.
2. Add domain-specific archetypes to the generator (equivalent to the 6 software archetypes above). Each archetype defines the unique research needs for that domain.
3. Adapt the complexity scoring rubric labels — replace software-specific terms ("Standard CRUD," "New library") with domain-equivalent complexity indicators.
4. Adjust the APPROACH block's target audience — replace "Principal Architect" with the domain-equivalent decision-maker.

#### Archetype Maintenance

Archetypes are versioned data, not finished code. Domain research needs shift as technology and regulations evolve. Each archetype should be treated as a living reference:
- Review archetypes after every 2–3 projects that use them.
- If a project required significant ad-hoc adaptation beyond its assigned archetype, consider whether a new archetype is warranted.
- If the generator's fallback path (projects that don't fit any archetype) triggers frequently, that's a signal a new archetype is needed — not evidence the fallback is working.

---

## 9. Composition Strategy

Replace fixed compose/build ratios with Wardley evolution mapping:

| Evolution Stage | Strategy | Examples |
|---|---|---|
| **Commodity/Utility** | 100% compose | Auth, DB, Storage, Queues, UI primitives |
| **Product** | Evaluate compose vs customize | CMS, analytics, email, payments |
| **Custom** | Extend/fork existing tools | Specialized dashboards, domain UI patterns |
| **Genesis** | 100% custom build | Core algorithms, domain state machines, business logic |

---

## 10. Specification Provenance

This framework was produced by applying Vivechak to itself:

- **11 independent research sessions** across frontier AI platforms
- **31 evidence nodes** from peer-reviewed studies, industry standards, and empirical benchmarks
- **10 hypothesis verdicts** — 0 fully validated as-is, 3 refuted, 5 refined, 2 validated with enhancements

The complete evidence base is preserved in [meta-research/](meta-research/). The framework is fully self-contained without it.
