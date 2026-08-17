# Complexity Scoring Template — URP v3.0
### 8-Dimension Project Assessment & Tier Mapping

> **Usage:** Complete this assessment before generating a research pipeline.
> The total score determines research depth, session budget, and decision routing.

---

## Project: `[Project Name]`
**Date:** `[YYYY-MM-DD]`
**Assessed By:** `[Name/Agent]`

---

## Scoring Rubric

### D1: Domain Novelty
> How novel is the problem domain relative to established software patterns?

| Score | Description |
|---|---|
| 0 | Standard CRUD / e-commerce — well-trodden patterns |
| 1 | Established B2B SaaS with known workflows |
| 2 | Unconventional workflow or niche vertical |
| 3 | Entirely new category, paradigm, or market |

**Score:** `[0-3]`
**Rationale:** `[Why this score]`

---

### D2: Technical Novelty
> How familiar is the engineering team with the required technology stack?

| Score | Description |
|---|---|
| 0 | Team's standard production stack |
| 1 | Familiar language, new library or service |
| 2 | New framework or programming paradigm |
| 3 | Unproven, cutting-edge, or experimental infrastructure |

**Score:** `[0-3]`
**Rationale:** `[Why this score]`

---

### D3: Regulatory / Compliance Exposure
> What level of regulated or sensitive data does the system handle?

| Score | Description |
|---|---|
| 0 | Zero sensitive data |
| 1 | Internal / business data only |
| 2 | PII, GDPR scope, SOC 2 |
| 3 | HIPAA, PCI-DSS, KYC/AML, FinTech rails, children's data |

**Score:** `[0-3]`
**Rationale:** `[Why this score]`

> ⚠️ **Hard Override:** If D3 = 3, all intersecting decision categories automatically receive Tier 3 Deep Research treatment, regardless of total project score.

---

### D4: Reversibility / Blast Radius
> How costly is it to change the core architectural decisions after launch?

| Score | Description |
|---|---|
| 0 | Throwaway / easily rewritable |
| 1 | Modularly swappable components |
| 2 | Core schema, multi-tenant data model |
| 3 | Deep platform infrastructure, public API contracts |

**Score:** `[0-3]`
**Rationale:** `[Why this score]`

---

### D5: Investment / Downstream Cost
> What is the financial and organizational investment at stake?

| Score | Description |
|---|---|
| 0 | Hackathon / weekend spike |
| 1 | Lean MVP / internal tool |
| 2 | Funded venture / production application |
| 3 | Enterprise mission-critical asset |

**Score:** `[0-3]`
**Rationale:** `[Why this score]`

---

### D6: Team Size & Coordination
> How many people need to coordinate on architectural decisions?

| Score | Description |
|---|---|
| 0 | Solo developer |
| 1 | Small team (2–4 engineers) |
| 2 | Mid-size team (5–12 engineers) |
| 3 | Multi-team / cross-org (>12) |

**Score:** `[0-3]`
**Rationale:** `[Why this score]`

---

### D7: Expected Longevity
> How long is this system expected to be in production?

| Score | Description |
|---|---|
| 0 | Days to weeks (disposable) |
| 1 | Months (evaluative prototype) |
| 2 | 1–3 years (standard product lifecycle) |
| 3 | 5+ years (core platform infrastructure) |

**Score:** `[0-3]`
**Rationale:** `[Why this score]`

---

### D8: Integration Complexity
> How many external systems, APIs, or legacy systems must be integrated?

| Score | Description |
|---|---|
| 0 | Standalone, no external APIs |
| 1 | 1–2 standard REST APIs |
| 2 | Multiple complex webhooks/APIs |
| 3 | Regulated banking rails, legacy ERP, complex federation |

**Score:** `[0-3]`
**Rationale:** `[Why this score]`

---

## Summary

| Dimension | Score |
|---|---|
| D1: Domain Novelty | `[X]` |
| D2: Technical Novelty | `[X]` |
| D3: Regulatory Exposure | `[X]` |
| D4: Reversibility | `[X]` |
| D5: Investment | `[X]` |
| D6: Team Size | `[X]` |
| D7: Longevity | `[X]` |
| D8: Integration | `[X]` |
| **Total** | **`[X]`** / 24 |

## Tier Assignment

| Score Range | Tier | Session Budget |
|---|---|---|
| 0–4 | Tier 0: Minimal | 1–3 sessions |
| 5–9 | Tier 1: Light | 4–8 sessions |
| 10–15 | Tier 2: Standard | 9–16 sessions |
| 16–24 | Tier 3: Deep | 17–30 sessions |

**Assigned Tier:** `[Tier X]`
**Session Budget:** `[N]` sessions
**Regulatory Override Active:** `[Yes/No]`
