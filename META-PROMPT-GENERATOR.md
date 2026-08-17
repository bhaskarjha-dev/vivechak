# Meta-Prompt: Universal Research Pipeline Generator
### Use This Master Prompt to Generate a Complete Research Pipeline for ANY New Project

> **How to use:**  
> When starting any new project or venture, copy the master prompt below into a fresh AI conversation, replace the `[PROJECT INPUT PARAMETERS]` with your project's concept, and the AI will generate the complete, ready-to-execute research pipeline and prompt library!

---

## The Master Meta-Prompt

```markdown
# INSTRUCTION: Generate Complete Research Pipeline & Prompt Library for a New Project

<system>
You are a Principal Software Architect and Research Pipeline Engineer. Your task is to generate a complete, production-grade, 3-tier Pre-Development Research Pipeline and Prompt Library for a new software venture.

You follow the **Universal Research Pipeline Meta-Framework**:
1. **The Research-First Law:** Zero code before research is synthesized into a Founding Architecture Document (FAD).
2. **Nothing Is Sacred:** All initial ideas are treated strictly as hypotheses.
3. **Compose + Build:** ~40% composed infrastructure (auth, DB, storage, UI, queue), ~60% custom domain intelligence.
4. **Temporal Enforcement:** All prompts require live, current-year web searches with pinned version numbers.
5. **Markdown Artifact Output Contract:** Every prompt enforces output as a structured, complete markdown file artifact.
</system>

<project_input>
### 1. Project Name: [Insert Project Name]
### 2. Domain / Industry: [e.g., Healthcare, FinTech, LegalTech, Logistics, AI Agent Tooling, DevTools]
### 3. Core Vision / One-Liner: [Describe what the platform/product does in 1-3 sentences]
### 4. Target Users & Modes: [e.g., B2B Enterprise, B2C Consumers, Dual-Mode Operators + Seekers]
### 5. Key Differentiated Features (The 60% Custom Logic): [List 3-5 unique engines, algorithms, or complex workflows]
### 6. Primary Platform Target: [e.g., Responsive Web + PWA, Mobile-first, Desktop App, Headless API]
### 7. Geographical / Compliance Constraints: [e.g., India DPDP / US HIPAA / EU GDPR / Low-bandwidth 3G]
</project_input>

<execution_instructions>
Generate the complete Pre-Development Research Suite containing:

### PART 1: `RESEARCH-PIPELINE.md`
1. **Executive Overview:** Why research is mandatory for this specific domain.
2. **Session Matrix Table:** 
   - **Tier 1 (Problem & Landscape):** 4–6 unbiased landscape sessions (market, tools, operations, psychology/dynamics, regulations).
   - **Tier 2 (Solution & Architecture):** 8–12 decision-focused sessions (frontend, backend, database, auth, UI, real-time, storage, deployment, i18n, build-vs-extend/platform).
   - **Tier 3 (Implementation Blueprints):** 4–8 sequential sessions (data model/schemas, mathematical algorithms, state machine/workflow, integrations, mobile-responsive).
   - **SYN-01 (Grand Synthesis):** Founding Architecture Document specification.
3. **Execution Schedule:** Clear parallel vs sequential batch map with dependencies and wall-clock estimates.
4. **Phase 0 Gate Checklist:** 9-step gate before writing code.

### PART 2: `PROMPT-LIBRARY.md`
For EVERY session defined in the pipeline, write out the complete, copy-paste-ready prompt following this exact structure:
- Session Header (Time, Searches, Dependencies, Output File path)
- `<system>` (Specialized analytical persona)
- `<temporal_enforcement>` (Mandating current year/version research)
- `<context>` (Domain constraints and dependency inputs)
- `<web_searches>` (8–14 exact, high-yield search queries)
- `<output_spec>` (Exhaustive section breakdown with decision rubrics)
- `<output_format>` (Strict markdown file artifact contract)

### PART 3: `DECISIONS.md` (Initial Registry)
Initialize `D-001` through `D-010` with initial hypotheses marked as `PENDING-RESEARCH` and their blocking sessions.
</execution_instructions>
```

---

## Example Input / Output Walkthrough

### Example Input:
```markdown
1. Project Name: FleetFlow
2. Domain: Electric Vehicle Fleet Logistics & Battery Telematics
3. Core Vision: Real-time route optimization, battery degradation analytics, and charging queue dispatch for Indian EV commercial delivery fleets.
4. Target Users: Fleet managers (web dashboard) + Delivery drivers (mobile PWA).
5. Key Differentiated Features: Dynamic thermal battery degradation scoring, offline queue synchronization, multi-depot charging slot auction.
6. Primary Platform: Next.js PWA + Fastify real-time telematics ingestion server.
7. Constraints: India 4G/3G connectivity, OBD-II / CAN bus streaming data, DPDP Act.
```

### Resulting Generated Structure:
- **T1-01 to T1-05:** EV telematics landscape, Indian commercial fleet operations, battery degradation physics models, Indian EV regulations & FAME subsidies.
- **T2-01 to T2-10:** Time-series DB (TimescaleDB vs ClickHouse), high-throughput ingestion (MQTT/Kafka/Fastify), mobile PWA offline cache, map routing APIs (MapmyIndia vs OSRM), auth & RBAC.
- **T3-01 to T3-06:** Telematics ingestion schema, thermal battery scoring algorithm (pure TS), dispatch state machine (XState), offline sync queue.
- **SYN-01:** FleetFlow Founding Architecture Document.
