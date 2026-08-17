# URP v3.0 — Evidence Grading & Verification Standard
### GRADE-Aligned Classification with Contextual Modifiers
*Grounded in: E-013 (GRADE framework), E-014 (Admiralty Code analysis), E-015 (AI citation confabulation)*

---

## 1. Base Evidentiary Tiers (A through E)

| Grade | Source Classification | Description | Starting Confidence |
|---|---|---|---|
| **A** | **Primary / Authoritative** | Official source code, published RFCs, formal API specifications, vendor engineering documentation, peer-reviewed studies in top venues | Very High (subject to recency) |
| **B** | **Empirical / Experimental** | Independently reproducible benchmarks, published engineering postmortems, peer-reviewed empirical studies, controlled experiments | High (when corroborated) |
| **C** | **Vendor / Motivated** | Vendor whitepapers, commercial comparison pages, marketing benchmarks, sponsored evaluations, vendor-authored case studies | Moderate (requires corroboration) |
| **D** | **Secondary / Opinion** | Unofficial tutorials, blog posts, community forums, conference talks without peer review, AI parametric recall without live verification | Low-to-Moderate |
| **E** | **Untraceable / Speculative** | Unverifiable claims, model confabulations, unattributed speculation, sources that cannot be independently located or confirmed | Zero (cannot support One-Way Doors) |

### Key Design Decisions

- **Grade E is strictly "Untraceable / Unverifiable"** — not "outdated" or "old." Stale but real sources receive their actual grade with a `stale` recency modifier, not a blanket E. This prevents the v2.0 "wavy lines" flaw where stale Grade A documentation was unfairly conflated with AI hallucinations.
- **5 tiers is the optimal granularity.** Binary (verified/unverified) loses critical nuance between official docs and blog posts. The full 36-cell Admiralty Code matrix collapses to the diagonal 87% of the time (E-014). Five tiers match the GRADE clinical certainty model's proven sweet spot (E-013).

---

## 2. Contextual Modifiers

Every cited claim carries three orthogonal modifiers that adjust the base grade's reliability:

### 2.1 Corroboration

| Value | Definition |
|---|---|
| `single` | Backed by exactly one source |
| `corroborated` | Confirmed by ≥2 independent, non-affiliated sources |
| `contested` | Direct conflict or contradiction identified across authoritative sources |

### 2.2 Recency

| Value | Definition |
|---|---|
| `fresh` | Published/verified within the domain half-life (<6 months for fast-moving libraries; <18 months for mature infrastructure) |
| `aging` | Approaching expected half-life; core claims require active verification |
| `stale` | Exceeds expected half-life or superseded by newer major releases |

### 2.3 Directness

| Value | Definition |
|---|---|
| `direct` | Evaluates the exact workload, language, and deployment target under consideration |
| `indirect` | Analogical evidence (e.g., Postgres benchmarked in C++ used to infer behavior in Node.js) |

---

## 3. Verification Method (Mandatory for AI-Authored Research)

| Value | Definition | Grade Cap |
|---|---|---|
| `fetched` | Retrieved and validated live via tool/web call during this research session | None |
| `cached` | Read from local repository file or previously verified artifact | None |
| `recalled` | Generated from model parametric memory without live tool verification | **Capped at Grade D** |
| `secondhand` | Sourced from an article summarizing a primary source | None (but grade reflects the summarizing source, not the original) |
| `human-provided` | Injected directly by a human engineer | None |

### Hard Rules

1. **Recalled claims are automatically capped at Grade D** regardless of the apparent source quality. An AI claiming "PostgreSQL 16 supports X" from memory without fetching the docs is Grade D at best.
2. **Zero recalled citations may underpin One-Way Door decisions.** All Type 1 architectural decisions require `fetched` or `cached` verification.
3. **Contested claims must be surfaced, not resolved by majority.** When corroboration status is `contested`, the decision record must explicitly document the contradiction and the causal reasoning for resolution.

---

## 4. Composite Citation Format

Every factual claim in a research artifact carries an inline composite signature:

```
Claim text (Grade [A-E] · [corroboration] · [recency] · [directness] | [verification_method])
```

**Examples:**

```
PostgreSQL 16 logical replication supports bi-directional failover under 
active-active topologies.
(Grade B · corroborated · fresh · direct | fetched)

React Server Components reduce client bundle size by ~30% on typical e-commerce 
applications according to Vercel benchmarks.
(Grade C · single · fresh · direct | fetched)

XState v5 supports parallel state machines with inter-machine communication.
(Grade A · corroborated · fresh · direct | fetched)

Based on general architectural patterns, event sourcing adds ~40% development 
overhead compared to standard CRUD.
(Grade D · single · aging · indirect | recalled)
```

---

## 5. Evidence Record Schema (E-NNN)

Standalone evidence records are stored in `evidence/E-NNN-[slug].md` for cross-project reuse and blast-radius tracking:

```yaml
---
id: E-047
title: "PostgreSQL 16 Logical Replication Active-Active Throughput Benchmark"
grade: B
modifiers:
  corroboration: corroborated
  recency: fresh
  directness: direct
verification:
  method: fetched
  fetched_date: 2026-08-18
  fetched_by: research-agent-1
source:
  url: "https://engineering.example.com/postgres-16-benchmarks"
  type: engineering-postmortem
  author: "ExampleCorp Core Platform Team"
tags: [postgres, replication, benchmarks]
used_by: [D-001, D-118]
---
```

The `used_by` field enables Shepard's Citations-style blast-radius tracking: when an evidence record is revised or invalidated, all dependent decisions are automatically flagged for review.

---

## 6. Decision Confidence (Decoupled from Evidence Grade)

Decision confidence is a **separate, independent assessment** from evidence grade:

| Confidence | Definition |
|---|---|
| `high` | Multiple corroborated Grade A/B sources agree; alternatives thoroughly evaluated; low residual uncertainty |
| `medium` | Adequate evidence supports the decision but some alternatives lack thorough evaluation, or key claims rest on single sources |
| `low` | Decision made under uncertainty; evidence is thin, contested, or indirect; explicit review trigger should be near-term |

A decision can have high confidence on Grade B evidence (well-corroborated engineering postmortems from multiple companies) or low confidence on Grade A evidence (official docs describe a feature, but real-world behavior under the specific workload is untested).
