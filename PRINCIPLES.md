# Universal Research Pipeline — Builder's Constitution
### 8 Evidence-Grounded Principles · URP v3.0
*Empirically validated via 14 meta-research sessions (Aug 2026)*

---

## The One Rule

**Every architectural decision must be grounded in traceable, graded evidence — not instinct, convention, or cached training data.**

If a research session produces generic advice applicable to any web app, the session has failed. If a decision cannot point to its supporting evidence records, the decision is ungrounded.

---

## The 8 Principles

### P1: The Context Architecture Law
*Supersedes: v2.0 Aspect-Isolation Law*

Research quality is governed by **attention budget, context purity, and task boundaries** — not by arbitrary session counts. The constraint is real (token/search budget explains ~80% of performance variance — Anthropic BrowseComp), but the remedy is conditional decomposition, not blanket fragmentation.

**Rules:**
- Decompose into dedicated research contexts when sub-tasks have low interdependency, high individual complexity, or divergent search spaces.
- Integrate coupled topics into structured joint sessions when evaluating holistic system tradeoffs.
- Every decomposed investigation **must** conclude with an explicit downstream synthesis pass.
- Over-isolation triggers the split-attention effect, introduces 4–15× token overhead, and misses systemic cross-cutting trade-offs.

> **Evidence:** E-001 (Anthropic BrowseComp R²=0.80), E-002 (MTI +12.4% on joint tasks), E-003 (Chandler & Sweller split-attention), E-004 (Chroma context rot)

### P2: Reversibility-Calibrated Rigor
*New in v3.0*

The depth of research, evidentiary burden, and review overhead allocated to a decision must scale directly with its **reversibility and blast radius**.

| Door Type | Definition | Research Depth | Evidence Bar |
|---|---|---|---|
| **One-Way (Type 1)** | Consequential, costly/impossible to reverse (primary datastore, data model, auth architecture, regulatory compliance, public API contracts) | Deep research, multi-session, corroborated evidence | Grade A/B, human review, premortem |
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
*Supersedes: v2.0 8-section XML prompt anatomy*

Research briefs must be:
- **Prescriptive** on WHAT to investigate, WHY it matters, WHAT boundaries apply, and WHAT coverage is required.
- **Directional** on HOW to execute: no pre-scripted search queries, no artificial search counts, no expert role-playing personas, no rigid output skeletons.

Front-load the complete brief in a single turn. Never drip-feed instructions across multiple turns (39% performance drop documented).

> **Evidence:** E-017 (Laban et al. ICLR 2026, 39% multi-turn drop), E-018 (persona debunking), E-019 (ReAct paradigm), E-020 (format restriction penalty)

### P5: Commodity-Maximized Composition
*Supersedes: v2.0 fixed ~40/60 ratio*

- **Compose 100%** of commodity/utility components from battle-tested providers (Auth, DB, Storage, Queues, UI primitives, CI/CD).
- **Build 100% custom** only for proprietary domain intelligence, core state machines, differentiated business logic, and unique algorithms.

The ratio varies by domain (80/20 for standard SaaS to 15/85 for novel algorithmic engines). Use Wardley evolution mapping, not fixed percentages.

> **Evidence:** E-024 (Wardley mapping), E-025 (Fowler MonolithFirst), E-026 (DORA loosely-coupled architecture)

### P6: Dual-Audience Artifact Architecture

Research outputs must serve both human readers and machine consumers:
- **Hybrid Markdown + YAML frontmatter** — human-readable body with machine-parseable metadata.
- **Standardized 7-section H2 skeleton** — consistent structure enabling automated Map-Reduce synthesis.
- **Strict separation** between `Recommendation` and `Alternatives Considered` — prevents AI code-generation contamination from rejected options.
- **JSON Schema validation** on frontmatter in CI.

> **Evidence:** E-021 (MADR 4.0), E-022 (GraphRAG header chunking), E-023 (Git diff mechanics)

### P7: Staged Triangulation & Diagnostic Disagreement
*Supersedes: v2.0 mandatory 3-model triangulation*

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

> **Evidence:** E-028 (Klein 1989, 30% risk reduction), Heuer's ACH, Annie Duke decision journaling, Superforecasting reference-class forecasting
