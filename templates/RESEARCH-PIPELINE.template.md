# [PROJECT_NAME] — Research Pipeline v1.0
### Complete Pre-Development Research Plan
*Created: [DATE]*

---

## 1. Why Research Before Code

Every architectural and strategic decision for **[PROJECT_NAME]** — tech stack, database, deployment, UI framework, data models, state machines, domain algorithms, security, and distribution — must be grounded in empirical, current-year research rather than unverified assumptions.

> **The Research-First Law:**  
> Research EVERYTHING that exists before committing to any approach. No hallucinated conclusions.

---

## 2. Pipeline Architecture

Three structured tiers + Grand Synthesis:

```
TIER 1: Problem & Landscape (Unbiased Discovery)
  └─ [N] sessions · Web search required · Domain-agnostic cataloging
  └─ Output: "What exists?" — complete catalogues, zero filtering

TIER 2: Solution & Architecture (Decision-Focused)
  └─ [N] sessions · Web search required · Project-specific trade-offs
  └─ Output: "What should we use?" — decision logs with options/rejected/why

TIER 3: Implementation (Build-Ready Blueprints)
  └─ [N] sessions · Web search + code examples · Strictly sequential
  └─ Output: "How do we build it?" — concrete schemas, pure algorithms, state machines

SYNTHESIS: Grand Integration (SYN-01)
  └─ 1 session · Ingests all T1–T3 outputs · Resolves all conflicts
  └─ Output: Founding Architecture Document (FAD) — single source of truth
```

**Total: [N] sessions · ~[N] hours · [N]+ web searches**

---

## 3. Execution Map

### Tier 1: Problem & Landscape (Independent / Parallel)

| Session | Name | Time | Searches | Output File |
|---|---|---|---|---|
| T1-01 | [Domain / Competitor Landscape] | 90 min | 12+ | `research/T1-01-landscape.md` |
| T1-02 | [Operations & Manual Workflows] | 75 min | 10+ | `research/T1-02-operations.md` |
| T1-03 | [User Dynamics & Decision Psychology] | 60 min | 8+ | `research/T1-03-psychology.md` |
| T1-04 | [Regulatory, Privacy & Security Standards] | 75 min | 10+ | `research/T1-04-security-privacy.md` |
| T1-05 | [Market Size, Monetization & Distribution] | 60 min | 8+ | `research/T1-05-market-distribution.md` |

### Tier 2: Solution & Architecture (Decision-Focused)

| Session | Name | Time | Searches | Depends On | Output File |
|---|---|---|---|---|---|
| T2-01 | Frontend Architecture & Framework | 90 min | 12+ | Nothing | `research/T2-01-frontend.md` |
| T2-02 | Backend Runtime & API Protocol | 90 min | 12+ | Nothing | `research/T2-02-backend.md` |
| T2-03 | Database Engine & Schema Paradigm | 90 min | 12+ | Nothing | `research/T2-03-database.md` |
| T2-04 | Authentication & Multi-Tenancy | 75 min | 10+ | Nothing | `research/T2-04-auth.md` |
| T2-05 | UI Component Library & Design System | 75 min | 10+ | T2-01 | `research/T2-05-ui.md` |
| T2-06 | Real-time & Collaboration Engine | 60 min | 8+ | Nothing | `research/T2-06-realtime.md` |
| T2-07 | File Storage, Vault & Encryption | 60 min | 8+ | Nothing | `research/T2-07-storage.md` |
| T2-08 | Deployment, Cloud Infrastructure & CI/CD | 90 min | 14+ | Nothing | `research/T2-08-deployment.md` |
| T2-09 | Internationalization & Localization | 60 min | 8+ | T2-01 | `research/T2-09-i18n.md` |
| T2-10 | Platform: Build vs Extend Platform | 90 min | 12+ | T1-01 | `research/T2-10-build-vs-extend.md` |

### Tier 3: Implementation (Sequential Blueprints)

| Order | Session | Name | Time | Depends On | Output File |
|---|---|---|---|---|---|
| 1 | T3-01 | Data Model & Entity Schema | 120 min | T2-03, T2-04 | `research/T3-01-data-model.md` |
| 2 | T3-02 | Core Domain Algorithm Engine | 90 min | T3-01 | `research/T3-02-algorithms.md` |
| 3 | T3-03 | Workflow & Pipeline State Machine | 75 min | T3-01 | `research/T3-03-state-machine.md` |
| 4 | T3-04 | Specialized Domain Subsystem | 90 min | T3-01 | `research/T3-04-subsystem.md` |
| 5 | T3-05 | Mobile-First Responsive Architecture | 75 min | T2-01, T2-05 | `research/T3-05-mobile-responsive.md` |

### Synthesis

| Session | Name | Time | Depends On | Output File |
|---|---|---|---|---|
| SYN-01 | **Grand Synthesis — Founding Architecture Document** | 120 min | ALL T1–T3 | `research/SYN-01-founding-architecture.md` |

---

## 4. Phase 0 Gate Transition Checklist

Before writing any application code:
- [ ] All research sessions completed and saved as markdown artifacts.
- [ ] `SYN-01` produced the **Founding Architecture Document (FAD)**.
- [ ] `ARCHITECTURE.md` rewritten with finalized stack decisions.
- [ ] `DECISIONS.md` updated with all decision IDs marked `RESOLVED`.
- [ ] Project scaffold initialized and clean build passing.
