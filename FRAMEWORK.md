# The Universal Research Pipeline Meta-Framework
### Complete Specification for Evidence-Grounded Pre-Development Research
*Version 3.0 — Empirically Validated via 14 Meta-Research Sessions*

---

## 1. Executive Philosophy

> **The Research-First Law:**
> *Before beginning development, investigate every irreversible architectural decision with graded, verifiable evidence. Develop nothing based on cached assumptions, outdated training data, or hallucinated conclusions.*

Software failures and architectural rewrites are rarely caused by poor coding — they are caused by **premature decisions made on unverified assumptions.** The Universal Research Pipeline is a structured pre-development engine that discovers, evaluates, resolves, and synthesizes the technical foundation of any software venture before a single line of application code is written.

**What changed from v2.0:** URP v3.0 was produced by applying the framework to itself — 11 independent research sessions across frontier AI platforms, cross-referenced with established disciplines (medicine, intelligence analysis, decision science, cognitive science), synthesized into 10 hypothesis verdicts with 31 evidence nodes. Every v2.0 axiom was tested empirically; none survived unchanged. See [meta-research/DECISIONS.md](meta-research/DECISIONS.md) for the complete evidentiary record.

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

URP v3.0 is governed by 8 evidence-grounded principles. Each principle was empirically validated and cites its supporting evidence. See [PRINCIPLES.md](PRINCIPLES.md) for the complete specification with evidence references.

| # | Principle | One-Sentence Summary |
|---|---|---|
| P1 | **Context Architecture Law** | Decompose research by attention budget and coupling, not arbitrary session counts; mandate synthesis after every decomposition |
| P2 | **Reversibility-Calibrated Rigor** | Scale research depth with decision reversibility — deep for One-Way Doors, fast spikes for Two-Way Doors |
| P3 | **Evidentiary Grounding** | Every claim must carry an evidence grade (A–E), contextual modifiers, and verification provenance |
| P4 | **Prescriptive Scope, Dynamic Method** | Be prescriptive on WHAT/WHY/BOUNDARIES; be directional on HOW/SEARCH — no hardcoded queries or personas |
| P5 | **Commodity-Maximized Composition** | Compose 100% of commodity infrastructure; build 100% custom only for proprietary domain logic |
| P6 | **Dual-Audience Artifacts** | Hybrid Markdown + YAML frontmatter for human reading and machine synthesis |
| P7 | **Staged Triangulation** | Single-model default; escalate to multi-model only for contested One-Way Doors |
| P8 | **Structured Falsification** | Prioritize disconfirming evidence, document rejected alternatives, schedule review triggers |

---

## 3. Pipeline Architecture

### 3.1 Topology: Constrained DAG with Adaptive Checkpoints

URP v3.0 replaces rigid 3-tier stage gating with a **Directed Acyclic Graph** governed by explicit information dependencies.

#### The Inverted Dependency Default
Every research session defaults to **unblocked** (eligible to execute immediately) unless an explicit hard information dependency is declared. This inverts v2.0's assumption that everything must wait for prior tiers.

#### Dependency Classification

| Type | Definition | Effect |
|---|---|---|
| **Hard (Information)** | Session B literally cannot formulate valid outputs without an artifact from Session A | Blocks execution — DAG edge |
| **Soft (Contextual)** | Session B benefits from terminology or constraints from Session A | No blocking — inject living Shared Context Brief |

**~67% of architectural decisions** depend on ≤1 landscape session, not all 4–6. Rigid stage-gating artificially inflates the critical path.

#### Adaptive Checkpoints
- **Expansion:** If a session reveals unexpected Cynefin-Complex trade-offs, spawn a maximum of 2 child sub-sessions (capped per checkpoint).
- **Contraction:** If a planned session's core question has been authoritatively resolved upstream, close it immediately with a "Resolved upstream" ADR record.
- **Incremental Synthesis:** Partial FAD compilation begins as soon as any structural branch closes, rather than waiting for 100% corpus completion.

#### Example DAG for a Standard Project

```
Layer 0: Landscapes (Fully Parallel)
├── Market & Competitor Landscape
├── Regulatory & Compliance Scan
├── Target User & Workload Profile
└── Technical Stack Landscape

Layer 1: Architectural Decisions (Sparse Dependency-Gated)
├── Frontend Paradigm          ← unblocked (eligible immediately)
├── API Design Pattern         ← unblocked
├── Datastore Selection        ← soft: User Workload Profile
├── Auth & Tenancy Architecture ← hard: Regulatory Scan
└── Cloud & Hosting Infra      ← hard: Regulatory Scan (data residency)

Layer 2: Blueprints (Hard-Gated)
├── Component Architecture     ← hard: Frontend + API
├── Core Data Schema           ← hard: Datastore Selection
├── Security & Compliance Spec ← hard: Auth + Cloud
└── Deployment & CI/CD         ← hard: Cloud Infra

Sink: Grand Synthesis (SYN-01) ← all Layer 2 sessions
```

Tier labels (Landscape, Architecture, Blueprint) are **organizational taxonomy**, not execution gates.

---

### 3.2 Adaptive Scaling Model

#### Step 1: 8-Dimension Complexity Scoring (0–24 Points)

Score each dimension from 0 (minimal) to 3 (extreme):

| Dimension | 0 (Low) | 1 (Moderate) | 2 (Substantial) | 3 (Extreme) |
|---|---|---|---|---|
| **Domain Novelty** | Standard CRUD | Established B2B SaaS | Unconventional workflow | New category/paradigm |
| **Technical Novelty** | Team's standard stack | Familiar language, new lib | New framework/paradigm | Unproven infra |
| **Regulatory** | Zero sensitive data | Internal data | PII, GDPR, SOC 2 | HIPAA, PCI-DSS, KYC/AML |
| **Reversibility** | Throwaway/rewritable | Modularly swappable | Core schema/multi-tenant | Deep platform infra |
| **Investment** | Hackathon/weekend | Lean MVP/internal tool | Funded venture | Enterprise mission-critical |
| **Team Size** | Solo developer | Small (2–4) | Mid-size (5–12) | Multi-team (>12) |
| **Expected Longevity** | Days to weeks | Months (prototype) | 1–3 years (product) | 5+ years (platform) |
| **Integration Complexity** | Standalone | 1–2 REST APIs | Multiple webhooks/APIs | Regulated rails/legacy ERP |

#### Step 2: Tier Mapping

| Score | Tier | Session Budget | Character |
|---|---|---|---|
| **0–4** | Tier 0: Minimal | 1–3 sessions | Fast spike on core unknown; all reversible items decided in the room |
| **5–9** | Tier 1: Light | 4–8 sessions | Standard SaaS/internal tools. Fast spikes on stack; deep check on core schema |
| **10–15** | Tier 2: Standard | 9–16 sessions | Commercial product. Full landscape, core ADRs, key blueprints |
| **16–24** | Tier 3: Deep | 17–30 sessions | Novel, regulated, multi-tenant platform. Full multi-aspect deep research |

> **🚨 Hard Override:** If Regulatory Exposure scores **3** (financial rails, health records, payment cards, children's data), all intersecting decisions automatically receive Tier 3 treatment regardless of total score.

#### Step 3: Per-Decision Routing Matrix

Inside any project tier, each decision is routed individually:

| | **Known Pattern** | **Unknown / Novel** |
|---|---|---|
| **Reversible (Two-Way Door)** | **SKIP** — decide by convention, 0 sessions | **FAST SPIKE** — 1 timeboxed session |
| **Irreversible (One-Way Door)** | **CONFIRM & COMMIT** — 1 session + ADR | **DEEP RESEARCH** — 2–5 sessions + ADR + premortem |

---

### 3.3 Staged Triangulation Protocol

| Stage | Trigger | Method | Cost |
|---|---|---|---|
| **Default** | All sessions | Single-model deep research | 1× |
| **Critique Probe** | Medium-stakes decisions | Feed Pass 1 synthesis to Model B for adversarial critique | 1.3× |
| **Full Triangulation** | One-Way Doors + Novel Domain + Genuine Contestation, OR critique probe detects active divergence | Run identical prompt on 2–3 models; synthesize divergence via ACH matrix | 3× |

**Why not always triangulate?** On factual/verifiable questions, frontier models converge 70–90% and share ~60% of their errors (correlated training). Triangulating factual queries buys false confidence from correlated consensus. Triangulation adds genuine value only on subjective, ambiguous, or predictive trade-offs where model training priors cause interpretive diversity.

---

## 4. Research Prompt Design

### 4.1 The 5-Block Prompt Anatomy

URP v3.0 replaces the v2.0 8-section XML prompt with 5 functional blocks:

| Block | Purpose | Prescriptive or Directional? |
|---|---|---|
| **BRIEF** | Goal, target deliverable, audience, decision being informed, required depth | **Prescriptive** — be specific about WHAT and WHY |
| **SCOPE** | Boundaries, time window, in/out scope, source priorities, date anchor | **Prescriptive** — be specific about BOUNDARIES |
| **APPROACH** | Exploration strategy, effort calibration, epistemic discipline, assumption declaration | **Directional** — guide HOW but don't prescribe exact queries |
| **DELIVERABLE** | Required-coverage checklist, comparison parameters, evidence grading, confidence flags | **Prescriptive** — WHAT to cover, not HOW to structure it |
| **FORMAT** | Markdown structure, frontmatter schema, artifact delivery | **Prescriptive** — output format requirements |

#### What Was Removed and Why

| v2.0 Element | Verdict | Evidence |
|---|---|---|
| `<system>` expert persona | **Removed** — personas don't improve factual accuracy and can impair recall | Zheng et al. EMNLP 2024; Basil et al. 2025 |
| `<web_searches>` hardcoded queries | **Removed** — violates agentic ReAct loop; models formulate better queries dynamically | Yao et al. 2023; Anthropic BrowseComp |
| `<bias_resistance>` negative instructions | **Removed** — triggers ironic "white bear" rebound in transformers | Documented in T2-06 |
| `<output_spec>` rigid sections | **Replaced** with coverage checklist — flexible structure, no Procrustean distortion | Tam et al. EMNLP 2024 |
| Minimum search counts | **Removed** — artificial floors don't match the model's judgment of "enough" | T1-02 |
| Multi-turn drip-feeding | **Banned** — front-load everything in one turn (39% performance drop otherwise) | Laban et al. ICLR 2026 |

### 4.2 Standard v3.0 Prompt Template

```markdown
# RESEARCH BRIEF: [Topic Title]

## BRIEF
We are investigating [core technical/architectural question].
This research will directly inform Architectural Decision [D-XXX: Title].
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

## DELIVERABLE
Deliver a structured Markdown document covering:
1. Executive Summary & Recommendation
2. Options Evaluation Matrix (criteria, operational overhead, failure modes)
3. Deep Technical Analysis of top 2–3 contenders
4. Inline evidence grades (A–E with modifiers and verification method)
5. Open Risks & Reversal Triggers

## FORMAT
Deliver as a single, complete Markdown file artifact with YAML frontmatter
per the URP v3.0 session schema.
```

---

## 5. Evidence & Decision System

### 5.1 Evidence Grading

See [EVIDENCE-GRADING.md](EVIDENCE-GRADING.md) for the complete specification including:
- 5-tier A–E base grades
- 3 contextual modifiers (corroboration, recency, directness)
- Mandatory verification method tracking
- Composite citation format
- E-NNN standalone evidence record schema

### 5.2 Architecture Decision Records (D-NNN)

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
review_trigger: "Re-evaluate if ingestion exceeds 50k ops/sec or storage >2TB"
human_reviewed: true      # Mandatory for one-way doors
schema_version: "3.0"
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
- **YAML frontmatter** — machine-parseable metadata validated via JSON Schema (see [schemas/](schemas/))
- **Markdown body** — human-readable narrative with standardized section structure
- **Strict separation** between `Recommendation` and `Alternatives Considered` — prevents AI code-generation contamination

### 6.2 Standardized 7-Section Body Skeleton

| # | Section | Purpose |
|---|---|---|
| 1 | `## Research Question` | Precise 1–2 sentence statement of what this session investigated |
| 2 | `## Key Findings` | 3–7 atomic bullet points extractable by synthesis agents |
| 3 | `## Recommendation` | Explicit, unambiguous technical guidance (isolated from rejected options) |
| 4 | `## Alternatives Considered` | Evaluated competing options and why they were rejected |
| 5 | `## Detailed Findings` | Flexible analytical body (benchmarks, tables, diagrams, code) |
| 6 | `## Open Questions & Risks` | Unresolved items and downstream risks |
| 7 | `## Sources & Evidentiary Ledger` | Numbered citations with composite grade metadata |

### 6.3 Map-Reduce Synthesis Workflow

When synthesizing 10–25 research sessions into the Founding Architecture Document:

1. **Filter:** Ingest only `status: final` sessions via `_index.yaml`
2. **Group:** Cluster by `topic` tags
3. **Map:** Extract atomic Key Findings and Recommendation sections
4. **Reduce:** Merge topic clusters into unified subsystem chapters
5. **Reconcile:** Resolve cross-session contradictions using confidence + date weighting
6. **Trace:** Annotate all FAD blocks with originating session and ADR IDs
7. **Gate:** Subject to Track B Phase 0 Gate before committing

---

## 7. Phase 0 Exit Gate

### 7.1 The Two-Track Standard

| Track | Applies To | Requirements | Speed |
|---|---|---|---|
| **Track A: Fast-Track** | Two-Way Door decisions | Decision logged, reversibility confirmed, single corroborated source | Minutes |
| **Track B: Rigorous Gate** | One-Way Door decisions | Full 9-step checklist below + Gary Klein Premortem | Hours–Days |

### 7.2 Track B: 9-Step Pre-Codebase Exit Checklist

1. **DAG Closure** — All required dependency paths terminated in `status: final`
2. **Contradiction Resolution** — All cross-model/cross-source divergences explicitly resolved
3. **Evidentiary Threshold** — Zero uncorroborated Grade C/D/E claims underpin irreversible pillars
4. **Verification Integrity** — 100% of critical citations carry `fetched` or `cached` (zero `recalled` for Type 1)
5. **Rejected Alternatives Documented** — Every ADR includes evaluated and rejected options with rationale
6. **Decay Triggers Assigned** — Every ADR contains explicit `review_trigger`
7. **Premortem Protocol** — 30-minute Gary Klein prospective hindsight: *"Assume catastrophic failure in 12 months. What caused it?"*
8. **Human Review** — Named Principal Architect signature on all Type 1 ADRs
9. **FAD Sealed** — Founding Architecture Document compiled, committed, ready for scaffolding

---

## 8. Generator Architecture

### 8.1 The 5-Layer Hybrid

The pipeline generator uses a **deterministic/AI hybrid** architecture. Scaffolding is deterministic (repeatable, testable); prompt prose is AI-synthesized (project-specific, non-generic).

| Layer | Function | Implementation |
|---|---|---|
| **0: Input** | Collect 8 project parameters via structured schema | Deterministic — validated JSON/YAML input |
| **1: Classification** | 6-archetype domain classifier + 8-dimension complexity scoring | Deterministic — rule-based decision tree |
| **2: Skeleton** | Assemble session matrix from archetype templates + tier budget | Deterministic — template composition (Yeoman-style `composeWith`) |
| **3: Prompt Synthesis** | Generate project-specific prompt prose for each session slot | AI — N parallel scoped calls (one per session) |
| **4: Registry Seeding** | Generate project-specific D-NNN hypotheses | AI — scoped calls tied to vision + constraints |
| **5: Validation** | Schema validation, archetype completeness, budget verification | Deterministic — CI-style gates |

### 8.2 The 6 Domain Archetypes

| Archetype | Unique Research Needs |
|---|---|
| **B2B SaaS** | Multi-tenancy, RBAC/SAML, CRM integrations, seat/usage pricing |
| **Developer Tools** | DX & time-to-first-value, CLI/SDK idioms, open-source licensing |
| **FinTech** | KYC/AML, banking rails, immutable ledgers, PCI-DSS, fraud detection |
| **AI/ML Systems** | Model selection, eval benchmarks, inference cost, context architecture, guardrails |
| **Consumer Mobile** | App Store compliance, offline-first sync, push/retention, in-app billing |
| **Real-Time / IoT** | MQTT/WebSockets, edge vs cloud compute, fleet OTA, hardware constraints |

**Blending Rule:** When a project matches multiple domains, adopt the Primary Archetype's full session matrix and append only the non-overlapping differentiator sessions from the Secondary Archetype.

### 8.3 Interim Generator (Prompt-Based)

Until the code-based 5-layer generator is built, use [META-PROMPT-GENERATOR.md](META-PROMPT-GENERATOR.md) — a v3.0-compliant master prompt that approximates the hybrid architecture in a single AI interaction.

---

## 9. Composition Strategy

### Wardley Evolution Framework

Replace fixed compose/build ratios with a domain-calibrated Wardley mapping:

| Evolution Stage | Strategy | Examples |
|---|---|---|
| **Commodity/Utility** | 100% compose from established providers | Auth (Clerk/Auth0), DB (Postgres/Supabase), Storage (S3/R2), Queues (SQS/BullMQ), UI primitives (Radix/shadcn) |
| **Product** | Evaluate compose vs customize based on fit | CMS, analytics, email, payments |
| **Custom** | Extend/fork existing tools with domain adaptations | Specialized dashboards, domain-specific UI patterns |
| **Genesis** | 100% custom build — this is the proprietary moat | Core algorithms, domain state machines, business logic engines, unique data models |

---

## 10. Specification Provenance

This framework specification was produced by the URP v3.0 meta-research pipeline:

- **11 independent research sessions** (T1-01 through T3-01) across frontier AI platforms
- **2 triangulated sessions** (T2-01, T2-02) with cross-model validation
- **31 evidence nodes** (E-001 through E-031) from peer-reviewed studies, industry standards, and empirical benchmarks
- **10 hypothesis verdicts** — 0 fully validated as-is, 3 refuted, 4 refined, 3 validated with enhancements

The complete evidence base is preserved in [meta-research/](meta-research/):
- [DECISIONS.md](meta-research/DECISIONS.md) — Sealed ADR corpus with all 10 verdicts
- [research/SYN-01-urp-v3-synthesis.md](meta-research/research/SYN-01-urp-v3-synthesis.md) — Grand synthesis document
- [research/](meta-research/research/) — All 14 primary research artifacts (639KB)

The meta-research serves as the empirical audit trail for v3.0. The framework is fully self-contained and does not require reading the research to operate — but the research is available for anyone who wants to verify why a specific design decision was made.
