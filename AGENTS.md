# Universal Research Pipeline — Agent Operating Manual

> **Read this file COMPLETELY before modifying anything in this repository.**  
> This is the single source of truth for how any AI agent (or human) works on the Universal Research Pipeline Meta-Framework.

---

## 1. What This Repository Is

**The Universal Research Pipeline (URP)** is a production-grade Meta-Framework and Autonomous Pre-Development Generator.

**Core Purpose:**  
Transform software architecture and technical pre-development from gut-feel, outdated training data, and hallucinated conclusions into a structured, evidence-graded, aspect-isolated, multi-model research engine *before a single line of application code is written*.

**Key Capabilities:**
1. **Aspect-Isolated Prompt Generation:** Generating 15–30 highly specialized research prompts where every single aspect of a venture gets its own dedicated deep-dive session (10–15 targeted web searches per session).
2. **Multi-Model Triangulation Engine:** Running identical prompts across Claude, ChatGPT Deep Research, and Gemini Deep Research to eliminate single-model bias and isolate divergence.
3. **5-Tier Evidence Grading Standard (Grade A–E):** Enforcing strict verification of claims from official RFCs/code (Grade A) down to unverified speculation (Grade E).
4. **Compose (~40%) + Build (~60%) Strategic Formula:** Grounding all architectures in composed world-class primitives and custom domain intelligence.
5. **Phase 0 Gate & Founding Architecture Document (FAD):** Structuring the bridge from research synthesis into clean repository scaffolding.

---

## 2. Repository Structure

```
D:\dev\pro\research-pipeline\
├── AGENTS.md                           ← THIS FILE: The single entry point & operating manual
├── README.md                           ← Repository overview, quick links, directory map
├── CONTEXT.md                          ← Tacit knowledge, historical lineage & design insights
├── VISION.md                           ← WHY: Mission, philosophy & long-term vision
├── PRINCIPLES.md                       ← Builder's Constitution & Non-Negotiable Axioms
├── FRAMEWORK.md                        ← The full Meta-Specification (Philosophy, 3-Tier Model, Triangulation)
├── META-PROMPT-GENERATOR.md            ← Master prompt to generate pipelines for ANY new project
├── ROADMAP.md                          ← Evolution & perfection roadmap for this meta-system
│
├── templates/                          ← Reusable blueprints for instant project bootstrapping
│   ├── RESEARCH-PIPELINE.template.md   ← 3-tier master plan & session matrix template
│   ├── PROMPT-LIBRARY.template.md      ← Structured copy-paste prompt library template
│   ├── DECISIONS.template.md           ← D-001... architectural decision registry template
│   ├── CONFLICT-RESOLUTION.template.md ← Checkpoint (CHK-01) trade-off resolution template
│   ├── FOUNDING-ARCHITECTURE.template.md ← SYN-01 Founding Architecture Document (FAD) template
│   └── PHASE-0-GATE.template.md        ← 9-step pre-codebase exit gate checklist
│
├── references/                         ← Proven empirical reference pipelines (Extraction Corpus)
│   ├── yugm/                           ← Yugm 22-session pipeline, prompt library, decisions & final stack
│   ├── triyantra/                      ← Triangulation methodology, execution plans & synthesis
│   ├── portfolio/                      ← 16-session decision pipeline & DP-USYN universal synthesis
│   ├── vivah-soodh/                    ← 8-session unbiased research pipeline & legacy prompts
│   ├── forge-rachak/                   ← Strategic & technical deep research pipelines
│   ├── rachak-research/                ← Architecture direction, competitive atlas, & vision maps
│   ├── pramedha/                       ← 72-module autonomous AI research plan
│   └── gemmra/                         ← Drug safety AI data/model research pipeline
│
└── brain-archive/                      ← Transcripts & synthesis history from Antigravity IDE brain
    ├── yugm-inception-transcript-summary.md
    └── multi-project-synthesis-history.md
```

---

## 3. The Core Laws & Axioms

### 1. The Aspect-Isolation Law (Anti-Degradation)
**Never bundle multiple architecture decisions or research topics into a single research session.**
- *Why:* If you combine 5 topics into one session, the model spreads 10 web searches across 5 topics (2 searches/topic) and dilutes its reasoning context. The output collapses into superficial, generic summaries.
- *Rule:* Each distinct aspect (Frontend, Database, Auth, Storage, Real-time, State Machines, Threat Modeling) MUST have its own dedicated session with 10–15 targeted searches dedicated solely to that topic.

### 2. The Research-First Law
**Zero application code is written before empirical research completes.** All pre-written documentation, assumptions, and instincts are treated strictly as *hypotheses* until proven or refuted by live empirical research.

### 3. The Multi-Model Triangulation Engine
Run identical prompts across frontier models (Claude, ChatGPT Deep Research, Gemini Deep Research) to find consensus, isolate disputed claims, and grade evidence (Grade A to E).

### 4. The Compose vs Build Split (~40% / ~60%)
Compose standard infrastructure from best-in-class primitives (auth, DB, storage, UI, queue) and invest engineering effort into proprietary domain algorithms and state machines.

### 5. The Markdown Artifact Contract
Every research prompt strictly mandates output as a structured, complete **Markdown File Artifact** that can be saved directly as a `.md` file without human editing.

---

## 4. Operational Workflow for New Sessions in this Repo

When a session starts in `D:\dev\pro\research-pipeline`:
1. Read `CONTEXT.md` and `PRINCIPLES.md` to understand the meta-architecture.
2. If improving the meta-framework: enhance `FRAMEWORK.md`, `META-PROMPT-GENERATOR.md`, and `templates/` by extracting deep patterns from `references/`.
3. When the user requests a research pipeline for a new project: invoke `META-PROMPT-GENERATOR.md` or generate the customized `RESEARCH-PIPELINE.md` and `PROMPT-LIBRARY.md` adhering to the Aspect-Isolation Law.
