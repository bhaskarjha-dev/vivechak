# Research Pipeline Template — URP v3.0
### Constrained DAG with Adaptive Scaling

> **Usage:** Copy this template when creating a new project research pipeline.
> Fill in the `[PLACEHOLDERS]` with project-specific values.

---

## 1. Project Complexity Assessment

### 1.1 Scoring Rubric

| Dimension | Score (0–3) | Rationale |
|---|---|---|
| Domain Novelty | `[0-3]` | `[Why]` |
| Technical Novelty | `[0-3]` | `[Why]` |
| Regulatory Exposure | `[0-3]` | `[Why]` |
| Reversibility / Blast Radius | `[0-3]` | `[Why]` |
| Investment / Downstream Cost | `[0-3]` | `[Why]` |
| Team Size & Coordination | `[0-3]` | `[Why]` |
| Expected Longevity | `[0-3]` | `[Why]` |
| Integration Complexity | `[0-3]` | `[Why]` |
| **Total** | **`[0-24]`** | |

### 1.2 Tier Assignment

- **Score:** `[X]` / 24
- **Tier:** `[0: Minimal (1-3) | 1: Light (4-8) | 2: Standard (9-16) | 3: Deep (17-30)]`
- **Session Budget:** `[N]` sessions
- **Regulatory Override:** `[Yes/No — if D3=3, all intersecting decisions get Tier 3]`

---

## 2. Domain Archetype Classification

- **Primary Archetype:** `[B2B SaaS | DevTools | FinTech | AI/ML | Consumer Mobile | Real-Time/IoT]`
- **Secondary Archetype:** `[if hybrid, else N/A]`
- **Blending Rule Applied:** `[Yes/No — append non-overlapping differentiators from secondary]`

---

## 3. Session Matrix (DAG)

### Layer 0: Landscape & Discovery (Fully Parallel)

| ID | Title | Topic | Door Type | Dependencies | Status |
|---|---|---|---|---|---|
| T1-01 | `[Title]` | `[topic-tag]` | `[one-way/two-way]` | None (unblocked) | `[ ]` |
| T1-02 | `[Title]` | `[topic-tag]` | `[one-way/two-way]` | None (unblocked) | `[ ]` |
| T1-0N | ... | ... | ... | ... | `[ ]` |

### Layer 1: Architectural Decisions (Sparse Dependency-Gated)

| ID | Title | Topic | Door Type | Dependencies | Status |
|---|---|---|---|---|---|
| T2-01 | `[Title]` | `[topic-tag]` | `[one-way/two-way]` | `[Hard: T1-XX / Soft: T1-XX / None]` | `[ ]` |
| T2-0N | ... | ... | ... | ... | `[ ]` |

### Layer 2: Blueprints & Specifications (Hard-Gated)

| ID | Title | Topic | Door Type | Dependencies | Status |
|---|---|---|---|---|---|
| T3-01 | `[Title]` | `[topic-tag]` | `[one-way/two-way]` | `[Hard: T2-XX]` | `[ ]` |
| T3-0N | ... | ... | ... | ... | `[ ]` |

### Sink: Grand Synthesis

| ID | Title | Dependencies | Status |
|---|---|---|---|
| SYN-01 | Grand Synthesis & FAD | All Layer 2 `status: final` | `[ ]` |

---

## 4. Execution Plan

### Parallel Execution Groups

| Group | Sessions | Can Start When |
|---|---|---|
| **Group A** (Parallel) | `[T1-01, T1-02, T1-03, T1-04]` | Immediately |
| **Group B** (Parallel) | `[T2-01, T2-02, T2-03]` | Immediately (unblocked) or after specified hard deps |
| **Group C** (Gated) | `[T2-04, T2-05]` | After `[T1-XX]` completes |
| **Group D** (Gated) | `[T3-01, T3-02]` | After `[T2-XX]` completes |
| **Synthesis** | SYN-01 | After all Layer 2 sessions |

### Adaptive Checkpoint Rules
- **Expansion cap:** Max 2 child sub-sessions per checkpoint
- **Contraction:** If a session's core question is resolved upstream, close with "Resolved upstream" ADR

---

## 5. Phase 0 Exit Gate

### Track A (Two-Way Door Decisions)
- [ ] Decision logged
- [ ] Reversibility confirmed
- [ ] Single corroborated source identified

### Track B (One-Way Door Decisions)
- [ ] DAG closed & all dependencies synthesized
- [ ] No unresolved contradictions across sessions
- [ ] Corroborated Grade A/B sourcing for all critical claims
- [ ] Zero recalled claims unverified
- [ ] Rejected alternatives documented in ADR
- [ ] Reversal triggers defined in ADR
- [ ] Gary Klein Premortem executed (30 min)
- [ ] Human Architect sign-off
- [ ] FAD sealed and committed
