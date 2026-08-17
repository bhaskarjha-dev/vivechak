# [PROJECT_NAME] — Decisions Registry
### Permanent Architectural & Strategic Decision Log

> **Status Values:**  
> - `ACTIVE` (needs implementation action)  
> - `RESOLVED` (decided, justified, and implemented)  
> - `PENDING-RESEARCH` (blocked on specific research session output)  
> - `DROPPED` (considered but discarded, with recorded reason)

---

## Pre-Research Hypotheses (D-001 to D-010)

### D-001: Core Brand & Architecture Pillar Model
**Date:** [DATE]  
**Decision:** [Description of architectural pillars]  
**Status:** `RESOLVED`  
**Rationale:** [Why this structure serves the core vision]

---

### D-002: Frontend Framework Selection
**Date:** [DATE]  
**Decision:** [Initial hypothesis, e.g., Next.js 16 App Router]  
**Status:** `PENDING-RESEARCH` — Blocked on `T2-01`  
**Rationale:** Team velocity vs performance to be empirically benchmarked.

---

### D-003: Backend Runtime & API Paradigm
**Date:** [DATE]  
**Decision:** [Initial hypothesis, e.g., Fastify + tRPC]  
**Status:** `PENDING-RESEARCH` — Blocked on `T2-02`  
**Rationale:** Benchmarking Node.js vs Go vs Rust for throughput and solo-maintainability.

---

### D-004: Database & Multi-Tenancy Architecture
**Date:** [DATE]  
**Decision:** [Initial hypothesis, e.g., PostgreSQL with Neon / RLS]  
**Status:** `PENDING-RESEARCH` — Blocked on `T2-03`  
**Rationale:** Row-level security vs schema-per-tenant isolation models.

---

### D-005: Build from Scratch vs Compose Primitives vs Extend Platform
**Date:** [DATE]  
**Decision:** Compose ~40% infrastructure + Build ~60% custom domain logic.  
**Status:** `PENDING-RESEARCH` — Blocked on `T2-10`  
**Rationale:** Validating open-core platforms vs primitive composition.
