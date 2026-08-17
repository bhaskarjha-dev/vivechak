# [PROJECT_NAME] — Founding Architecture Document (FAD)
### The Definitive Technical Blueprint & Single Source of Truth
*Derived from Synthesis of All Research Sessions (SYN-01)*

---

## 1. Executive Summary & Product Scope
- **Vision:** [1-paragraph crisp product summary]
- **Architecture Paradigm:** [e.g., Modular Monorepo / Monolith with Clean Architecture]
- **The Compose vs Build Split:**
  - **Composed (~40%):** Auth ([Better Auth/Clerk]), DB ([Neon/PostgreSQL]), Storage ([Cloudflare R2]), Real-time ([Socket.IO]), UI ([shadcn/ui]).
  - **Custom Built (~60%):** [Proprietary Domain Engines, State Machines, Custom Workflows].

---

## 2. Pinned Technology Stack

| Layer | Pinned Technology | Version | Research Justification |
|---|---|---|---|
| **Monorepo** | pnpm + Turborepo | latest | T0-01: Zero-overhead multi-package orchestration |
| **Frontend** | [Next.js / Vite] | [vX.X] | T2-01: App router, SSR, bundle optimization |
| **Backend** | [Fastify / Hono] | [vX.X] | T2-02: Type safety, high throughput, plugin system |
| **Database** | PostgreSQL | [v17] | T2-03: Row-level security, JSONB capabilities |
| **ORM** | Drizzle ORM | [vX.X] | T2-03: TypeScript-first, zero runtime bloat |
| **Authentication** | [Better Auth / Clerk] | [vX.X] | T2-04: Multi-tenant organization support, RBAC |
| **UI & Styling** | shadcn/ui + Tailwind | [v4] | T2-05: Accessible Radix primitives, mobile-first |
| **Storage** | Cloudflare R2 | S3 API | T2-07: Zero egress fees, client-side presigned uploads |
| **Deployment** | [Dokploy / Cloudflare / Docker] | latest | T2-08: Self-hostable, low-cost scalable VM |

---

## 3. Core Data Architecture & Entity Tree
- Entity Relationship Diagram (Mermaid)
- Multi-tenancy RLS isolation policies
- Sensitive data encryption boundaries (AES-256-GCM field encryption)

---

## 4. Proprietary Domain Engines & Pure Functions
- Algorithmic specifications (input → computation → output)
- Test cases and mathematical edge cases

---

## 5. State Machine & Pipeline Workflows
- Full state chart (Mermaid statechart)
- Stage transitions, guards, sub-states, and stale triggers

---

## 6. Implementation Sequence & Phase 1 Roadmap
1. **Phase 1: Project Scaffolding & Core Auth/DB Setup**
2. **Phase 2: Core Domain Data Models & CRUD**
3. **Phase 3: Domain Engine & State Machine Integration**
4. **Phase 4: Frontend Views & Real-Time Sync**
5. **Phase 5: End-to-End Verification & Deployment**
