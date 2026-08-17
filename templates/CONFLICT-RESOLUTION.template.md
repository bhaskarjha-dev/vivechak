# Checkpoint CHK-01: Architectural Conflict Resolution
### Resolving Divergences Across Independent Research Sessions

---

## 1. Conflict Audit Table

| Conflict ID | Domain | Session A Claim | Session B Claim | Core Tension |
|---|---|---|---|---|
| **C-01** | Backend Runtime | T2-02 recommends Fastify (Node.js) for tRPC type-sharing | T2-08 benchmarks Go for 10x lower memory / raw latency | Development velocity & type safety vs microsecond execution efficiency |
| **C-02** | Database Hosting | T2-03 recommends Neon (Serverless Postgres) | T2-08 highlights self-hosted Docker Postgres on VM for cost control | Operational zero-maintenance vs predictable flat hosting pricing |

---

## 2. Deep First-Principles Evaluation

### Conflict C-01: Node.js (Fastify) vs Go
- **Evaluation Criteria:**
  1. Full-stack TypeScript type sharing (Frontend ↔ Backend)
  2. Solo-developer maintenance overhead
  3. Real-world traffic requirements (MVP vs 100k DAU)
  4. Ecosystem package availability for domain algorithms
- **Verdict & Resolution:**  
  **Select Node.js (Fastify 5 + tRPC).** Full-stack end-to-end type safety eliminates an entire class of integration bugs and enables 2-3x faster feature iteration for a lean team. CPU latency is not the bottleneck; IO and developer velocity dominate.
- **Updated Decision Record:** Updated `D-003` → `RESOLVED` with Fastify.

---

## 3. Final Stack Consensus Table

| Layer | Final Choice | Resolving Checkpoint | Superseded Options |
|---|---|---|---|
| Backend | Node.js 24 + Fastify + tRPC | CHK-01 | Go, NestJS, Express |
| Database | PostgreSQL on Neon | CHK-01 | Self-hosted VM Postgres |
