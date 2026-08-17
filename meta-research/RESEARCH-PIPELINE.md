# Universal Research Pipeline — Meta-Research Pipeline v1.0
### Empirically Validating the Framework That Validates Frameworks
*Created: August 2026*

---

## 1. The Recursive Imperative

The URP's own Axiom 2 states: *"Nothing Is Sacred — all initial assumptions are hypotheses until empirical research validates them."*

The current URP was built **inductively** from 6 projects' experiences (Yugm, Triyantra, Vivah, Portfolio, Forge/Rachak, Pramedha/Gemmra). Its core axioms — aspect-isolation, multi-model triangulation, the 3-tier staging model, fixed session counts, prompt anatomy — were never **deductively validated** against empirical evidence about how AI deep research actually works, what prompt engineering science says, or how other fields approach structured decision-making.

**This pipeline applies the URP to itself.** Every core assumption is now a testable hypothesis. The research will either validate, refute, or refine each one — and the resulting evidence will be the foundation for URP v3.0, replacing all prior reference pipelines as the single source of truth.

---

## 2. The 10 Hypotheses Under Test

| ID | Current URP Assumption | Research Session |
|---|---|---|
| **H-01** | Aspect-isolation universally produces better research than integrated multi-topic prompts | T2-01 |
| **H-02** | Multi-model triangulation always adds value over single-model deep research | T2-02 |
| **H-03** | The 3-tier sequence (Landscape → Architecture → Blueprint) is the optimal decomposition | T2-03 |
| **H-04** | Pipelines should be 17–27 sessions for most projects | T2-04 |
| **H-05** | Grade A–E evidence classification improves decision quality | T2-05 |
| **H-06** | The current 8-section prompt anatomy is optimal | T2-06 |
| **H-07** | Markdown artifacts are the right output format for downstream use | T2-07 |
| **H-08** | ~40/60 compose/build is a meaningful, validated ratio | SYN-01 |
| **H-09** | A Phase 0 Gate with strict exit criteria prevents architectural mistakes | SYN-01 |
| **H-10** | The meta-prompt generator should produce fixed templates per project type | T3-01 |

---

## 3. Pipeline Design Principles

This meta-research pipeline is designed with three principles that depart from the current URP's approach:

### Principle 1: Full Session Independence
Every research session (T1-01 through T3-01) is **fully independent**. No session depends on another's output. This prevents cascading confirmation bias — each topic is explored with a fresh, uncontaminated analytical lens. Only SYN-01 ingests all outputs for synthesis.

> **Rationale:** If T1-02 concludes "search budget dilution is real," and T2-01 receives that finding as context, T2-01 is anchored toward confirming isolation rather than independently evaluating it. Independence ensures each session discovers truth on its own terms.

### Principle 2: Directional Freedom in Prompts
Prompts define WHAT to learn and WHY it matters, but give the AI complete freedom in HOW to research. No pre-specified search queries, no minimum search counts, no rigid output sections. Frontier deep research modes are sophisticated reasoners that can formulate better queries, determine adequate search depth, and discover unexpected angles when not constrained by prescriptive instructions.

### Principle 3: Selective Triangulation
Not every session warrants multi-model triangulation. Triangulation is recommended only where structural model biases could meaningfully alter conclusions — primarily for sessions evaluating subjective design decisions where interpretation diversity matters.

---

## 4. Session Matrix

### Tier 1: Empirical Foundations

| Session | Name | Triangulation | Output |
|---|---|---|---|
| T1-01 | Pre-Development Research Methodologies Landscape | 🔬 Single-Model | `meta-research/research/T1-01-methodologies-landscape.md` |
| T1-02 | Frontier AI Deep Research — Mechanisms, Limits & Optimal Use | 🔬 Single-Model | `meta-research/research/T1-02-ai-deep-research-capabilities.md` |
| T1-03 | Prompt, Context & Harness Engineering for Frontier AI | 🔬 Single-Model | `meta-research/research/T1-03-prompt-context-engineering.md` |

### Tier 2: Core Design Decisions

| Session | Name | Triangulation | Output |
|---|---|---|---|
| T2-01 | Aspect-Isolation vs Integrated Research | 🔺 **Recommended** | `meta-research/research/T2-01-aspect-isolation.md` |
| T2-02 | Multi-Model Triangulation — Value vs Overhead | 🔺 **Meta-Required** ¹ | `meta-research/research/T2-02-triangulation-value.md` |
| T2-03 | Pipeline Topology — Sequence, DAG, Iterative, or Adaptive | 🔬 Single-Model | `meta-research/research/T2-03-pipeline-topology.md` |
| T2-04 | Adaptive Scaling — Right-Sizing for Project Complexity | 🔬 Single-Model | `meta-research/research/T2-04-adaptive-scaling.md` |
| T2-05 | Decision Science, Evidence Classification & Traceability | 🔬 Single-Model | `meta-research/research/T2-05-evidence-grading.md` |
| T2-06 | Research Prompt Design — Optimal Structure for Deep Research | 🔺 Optional ² | `meta-research/research/T2-06-prompt-anatomy.md` |
| T2-07 | Research Output Architecture — Format, Structure & Downstream Use | 🔬 Single-Model | `meta-research/research/T2-07-output-architecture.md` |

### Tier 3: Implementation Design

| Session | Name | Triangulation | Output |
|---|---|---|---|
| T3-01 | Meta-Prompt Generator & Template Architecture | 🔬 Single-Model | `meta-research/research/T3-01-generator-architecture.md` |

### Synthesis

| Session | Name | Triangulation | Output |
|---|---|---|---|
| SYN-01 | Grand Synthesis → URP v3.0 Specification | 🔬 Single-Model | `meta-research/research/SYN-01-urp-v3-synthesis.md` |

> ¹ **Meta-Required:** Running T2-02 through triangulation is itself a meta-experiment. If models give substantially different answers about triangulation's value, that's evidence FOR triangulation. If they converge, that's evidence single-model research suffices for this class of question.
>
> ² **Optional:** Models may have self-serving biases about what prompt elements matter. Triangulation helps detect this.

**All sessions T1-01 through T3-01 are fully independent.** Run them in any order, at any time. Only SYN-01 requires all prior outputs.

---

## 5. Execution Model

```
┌──────────────────────────────────────────────────────────────────┐
│                   ALL RESEARCH SESSIONS                          │
│                                                                  │
│   T1-01  T1-02  T1-03  T2-01  T2-02  T2-03  T2-04              │
│   T2-05  T2-06  T2-07  T3-01                                    │
│                                                                  │
│   → Run in ANY order, at ANY time                                │
│   → Each session is fully self-contained                         │
│   → No session sees another session's output                     │
│   → Each produces an independent research artifact               │
└────────────────────────────┬─────────────────────────────────────┘
                             │
                             │ All 11 artifacts fed as input
                             ▼
                 ┌───────────────────────┐
                 │       SYN-01          │
                 │   Grand Synthesis     │
                 │   → URP v3.0 Spec     │
                 └───────────────────────┘
```

**Total: 12 sessions + 1 synthesis = 13**

**Why full independence?** Cross-session dependency creates bias contamination. If you read T1-02's conclusions before running T2-01, you're anchored. If you feed T1-03's findings into T2-06, the model is primed rather than discovering independently. Independence maximizes the epistemic value of each session — and SYN-01's job is to find consensus, resolve conflict, and handle integration.

---

## 6. Triangulation Strategy

| Divergence Risk | When It Applies | Strategy |
|---|---|---|
| **Low** | Factual landscapes, well-documented topics, implementation design | 🔬 Single-Model Deep Research |
| **Medium** | Prompt/output design where models have self-serving biases | 🔺 Optional triangulation |
| **High** | Core axiom evaluation, methodology design where models have structural biases | 🔺 Recommended triangulation |

**Key insight:** When frontier models do deep web research, they access the same internet — triangulation's residual value is in **interpretation diversity**, not information diversity. Full triangulation is valuable only for subjective, judgment-heavy questions where model-specific biases could meaningfully alter conclusions.

---

## 7. Output Directory Structure

```
meta-research/
├── RESEARCH-PIPELINE.md         ← THIS FILE
├── PROMPT-LIBRARY.md            ← Copy-paste prompts for every session
├── DECISIONS.md                 ← Current assumptions as testable hypotheses
└── research/                    ← Save all session outputs here
    ├── T1-01-methodologies-landscape.md
    ├── T1-02-ai-deep-research-capabilities.md
    ├── T1-03-prompt-context-engineering.md
    ├── T2-01-aspect-isolation.md
    ├── T2-02-triangulation-value.md
    ├── T2-03-pipeline-topology.md
    ├── T2-04-adaptive-scaling.md
    ├── T2-05-evidence-grading.md
    ├── T2-06-prompt-anatomy.md
    ├── T2-07-output-architecture.md
    ├── T3-01-generator-architecture.md
    └── SYN-01-urp-v3-synthesis.md
```

---

## 8. Post-Research: What Happens Next

1. All research artifacts saved to `meta-research/research/`.
2. `DECISIONS.md` updated: each D-NNN marked `VALIDATED`, `REFUTED`, or `REFINED` with evidence.
3. `SYN-01` produces the **URP v3.0 Specification** — the evidence-grounded successor to the current framework.
4. Current `FRAMEWORK.md`, `templates/`, and `META-PROMPT-GENERATOR.md` are rewritten based on validated findings.
5. The `references/` directory becomes archival — the meta-research becomes the definitive source of truth.
