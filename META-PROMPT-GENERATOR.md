# URP v3.0 — Pipeline Generator
### 5-Layer Hybrid Architecture Specification & Interim Generator Prompt
*Supersedes: v2.0 monolithic master meta-prompt*

---

## Architecture Overview

The v2.0 monolithic master prompt asked a single AI generation to simultaneously plan (which sessions), write (prompt prose), and structure (decision registry) — a pattern that decomposition research consistently finds less reliable than focused steps. The v3.0 generator splits this into **deterministic scaffolding** (repeatable, testable) and **scoped AI synthesis** (project-specific, non-generic).

```
Layer 0: Input Schema         ← Deterministic: collect & validate 8 parameters
Layer 1: Classification       ← Deterministic: domain archetype + complexity score
Layer 2: Skeleton Assembly    ← Deterministic: session matrix from templates + tier budget
Layer 3: Prompt Synthesis     ← AI: N parallel scoped calls (1 per session slot)
Layer 4: Registry Seeding     ← AI: D-NNN hypotheses from vision + constraints
Layer 5: Validation Gates     ← Deterministic: schema check, completeness, budget cap
```

---

## Layer 0: Input Contract

Collect exactly 8 parameters. These are the minimum viable inputs to classify a project and determine research scope:

```yaml
# ProjectSpec v3.0 Input Schema
project:
  name: ""                    # Project name
  domain: ""                  # Primary domain (maps to archetype)
  vision: ""                  # 1–5 sentence description of what the product does
  constraints: ""             # Geographic, regulatory, infrastructure, team constraints
  unique_features: []         # 3–7 key differentiators (the proprietary/genesis components)
  platform_targets: []        # Web, Mobile PWA, Native iOS/Android, Desktop, API-only, CLI
  team_size: ""               # Solo | Small (2-4) | Mid (5-12) | Large (>12)
  risk_profile: ""            # Hackathon | Internal Tool | Commercial Product | Regulated Platform
```

---

## Layer 1: Domain Classification & Complexity Scoring

### 6 Domain Archetypes

| Archetype | Mandatory Sessions | Differentiator Sessions |
|---|---|---|
| **B2B SaaS** | Multi-tenancy, RBAC/SAML/SSO, CRM integrations | Seat/usage pricing, buyer vs user journey |
| **Developer Tools** | DX & time-to-first-value, CLI/SDK idioms | Package distribution, OSS licensing, docs-as-product |
| **FinTech** | KYC/AML compliance, banking rails | Immutable ledger schemas, PCI-DSS, fraud detection |
| **AI/ML Systems** | Model selection, eval frameworks, inference cost | Context architecture, guardrails, RAG/fine-tune |
| **Consumer Mobile** | App Store compliance, offline-first sync | Push/retention loops, in-app billing, onboarding |
| **Real-Time / IoT** | Protocol selection (MQTT/WS), edge compute | Fleet OTA, hardware constraints, power budget |

**Blending Rule:** When a project matches 2+ archetypes (e.g., "AI-powered FinTech"), adopt the Primary's full session matrix and append only the non-overlapping differentiator sessions from Secondary archetypes.

### 8-Dimension Complexity Score (0–24)

See [FRAMEWORK.md §3.2](FRAMEWORK.md) for the complete scoring rubric. The score maps to:

| Score | Tier | Budget | Research Character |
|---|---|---|---|
| 0–4 | Minimal | 1–3 | Fast spike on 1 core unknown |
| 5–9 | Light | 4–8 | Stack spikes + core schema deep-dive |
| 10–15 | Standard | 9–16 | Full landscape + core ADRs + blueprints |
| 16–24 | Deep | 17–30 | Full multi-aspect deep research |

---

## Layer 2: Skeleton Assembly

Deterministic assembly of the session matrix from archetype templates:

1. Start with the Primary Archetype's **mandatory session set**
2. Add Secondary Archetype's **non-overlapping differentiator sessions**
3. Apply tier budget ceiling — prune lowest-priority sessions if over budget
4. Declare dependencies (hard/soft) between sessions
5. Emit the skeleton as a structured YAML session matrix

---

## Layer 3: Scoped Prompt Synthesis

For each session slot in the skeleton, make **one focused AI call** to generate project-specific prompt prose following the 5-block anatomy (BRIEF, SCOPE, APPROACH, DELIVERABLE, FORMAT). Each call receives:
- The session's topic and archetype context
- The project's vision, constraints, and unique features
- The 5-block template from [FRAMEWORK.md §4.2](FRAMEWORK.md)

These calls are **independent and parallelizable**.

---

## Layer 4: Decision Registry Seeding

Generate project-specific D-NNN hypotheses as `status: proposed` ADR stubs:
- Map each unique feature to a testable architectural hypothesis
- Assign `door_type` (one-way/two-way) based on reversibility
- Link each hypothesis to the sessions that will inform it

---

## Layer 5: Validation Gates

1. **Schema Validation** — All frontmatter passes JSON Schema ([schemas/](schemas/))
2. **Archetype Completeness** — All mandatory sessions for the primary archetype are present
3. **Budget Verification** — Total sessions ≤ tier budget ceiling
4. **Dependency Integrity** — All hard dependency targets exist in the session matrix
5. **Coverage Check** — Every unique feature maps to ≥1 research session

---

## Interim Generator Prompt (v3.0)

Until the code-based 5-layer generator is built, use this prompt in a fresh AI conversation to generate a complete research pipeline for any project:

```markdown
# GENERATE: URP v3.0 Research Pipeline

## BRIEF
You are a Research Pipeline Architect generating a complete, ready-to-execute
pre-development research pipeline for a new software project. Your output will
be used by engineering teams to make evidence-grounded architectural decisions
before writing any application code.

You follow the URP v3.0 Meta-Framework:
1. Context Architecture Law: Decompose by coupling and budget, not fixed session counts.
2. Reversibility-Calibrated Rigor: Scale depth with decision reversibility (One-Way vs Two-Way Doors).
3. 5-Block Prompt Anatomy: BRIEF, SCOPE, APPROACH, DELIVERABLE, FORMAT. No personas, no hardcoded queries.
4. Staged Triangulation: Single-model default; escalate for contested One-Way Doors.
5. A–E Evidence Grading: With corroboration, recency, directness modifiers and verification tracking.
6. Commodity-Maximized Composition: Compose all commodity; build custom only for proprietary logic.

## PROJECT PARAMETERS
1. Project Name: [INSERT]
2. Domain / Industry: [INSERT — maps to archetype: B2B SaaS, DevTools, FinTech, AI/ML, Consumer Mobile, Real-Time/IoT, or hybrid]
3. Core Vision: [1–5 sentences describing what the product does]
4. Constraints: [Geographic, regulatory, infrastructure, team constraints]
5. Key Differentiators: [3–7 unique engines, algorithms, or complex workflows that are the proprietary moat]
6. Platform Targets: [Web, Mobile PWA, Native, Desktop, API-only, CLI]
7. Team Size: [Solo | Small (2-4) | Mid (5-12) | Large (>12)]
8. Risk Profile: [Hackathon | Internal Tool | Commercial Product | Regulated Platform]

## DELIVERABLE

Generate THREE documents as complete Markdown file artifacts:

### PART 1: RESEARCH-PIPELINE.md
1. **Complexity Assessment:** Score the project across 8 dimensions (0–24), assign the Tier (0–3), and show the per-dimension rationale.
2. **Session Matrix:** A DAG of research sessions organized by layer (Landscape → Architecture → Blueprints → Synthesis). For each session:
   - ID, title, topic tags
   - Door type (one-way/two-way) of the decision it informs
   - Dependencies (hard/soft with targets)
   - Estimated depth (single session / multi-session)
3. **Execution Plan:** Which sessions can run in parallel, which are dependency-gated, and recommended execution order.
4. **Phase 0 Gate:** Two-track exit criteria (Track A for Two-Way, Track B for One-Way decisions).

### PART 2: PROMPT-LIBRARY.md
For every session in the pipeline, write out the complete, copy-paste-ready research prompt using the 5-block anatomy:
- **BRIEF:** Goal, decision being informed, required analytical depth.
- **SCOPE:** Time window, in/out boundaries, source priorities.
- **APPROACH:** Exploration strategy (directional, not prescriptive), epistemic discipline.
- **DELIVERABLE:** Required-coverage checklist, comparison parameters, inline evidence grading.
- **FORMAT:** YAML frontmatter + 7-section body skeleton.

Do NOT include expert personas, hardcoded search queries, minimum search counts, or rigid output skeletons. Each prompt must be front-loaded as a single complete brief.

### PART 3: DECISIONS.md (Initial Registry)
Initialize D-001 through D-NNN with:
- YAML frontmatter (id, title, status: proposed, door_type, evidence_refs: [], review_trigger)
- Context & Problem Statement
- Initial competing hypotheses
- Sessions that will inform the verdict

## SCOPE
- Today's date: [INSERT DATE]
- Generate prompts appropriate for frontier AI deep research modes (Claude, Gemini, ChatGPT with search enabled)
- Scale session count to the Tier determined by complexity scoring
- Use the Wardley evolution framework for compose/build decisions
- Apply the blending rule if the project matches multiple domain archetypes

## FORMAT
Output each part as a complete, standalone Markdown file artifact.
```

---

## Future: Code-Based Generator

The specification above is designed to be implementable as a code-based tool:
- **Layers 0–2, 5:** Deterministic logic (TypeScript/Python) with archetype configs as data files
- **Layers 3–4:** Scoped LLM API calls (one per session slot, parallelizable)
- **Distribution:** CLI tool (`npx urp-generate`) or programmatic API

See [ROADMAP.md](ROADMAP.md) for the build timeline.
