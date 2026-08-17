# Universal Research Pipeline — Context & Tacit Knowledge
### Design History, Meta-Research Journey & Architectural Insights
*URP v3.0 · August 2026*

---

## 1. Why This Meta-System Exists

In past software projects, creating high-quality pre-development research required manually hunting down context from 5+ previous repositories — each having independently discovered pieces of the same methodology:

- **Vivah / Soodh:** 8 sessions proving unbiased discovery-first cataloging.
- **Pramedha:** 72 modules proving systematic research partitioning.
- **Triyantra:** 3-model triangulation and Grade A–E evidence grading.
- **Portfolio:** 16-session decision framework with universal final synthesis.
- **Forge & Rachak:** Zero-bias prompt independence and sovereignty filtering.
- **Gemmra:** Drug safety AI validation proving domain-specific ADR rigor.
- **Yugm:** 22-session pipeline proving the compose+build split and Markdown artifact contracts.

**The Solution:** URP v1.0/v2.0 centralized these patterns into a single meta-framework.

**The Problem:** v2.0 was built inductively — synthesized from project experience, not empirically validated. Its core axioms were instincts codified as laws.

---

## 2. The v2.0 → v3.0 Meta-Research Journey

In August 2026, URP was subjected to its own methodology: **11 independent deep research sessions** tested every foundational assumption against peer-reviewed studies, empirical benchmarks, and frontier AI architecture.

### What We Found

| v2.0 Assumption | Empirical Reality |
|---|---|
| "One topic per session, always" | Context architecture matters more than session count. Joint evaluation wins when topics are coupled (+12.4% on MTI). Over-isolation adds 4–15× overhead and triggers split-attention. |
| "Always triangulate across 3 models" | Frontier models share ~60% of their errors (ICML 2025). On factual queries, triangulation buys false confidence from correlated consensus. Valuable only for subjective, contested decisions. |
| "17–27 sessions for every project" | Arbitrary mean that over-researches CRUDs and under-researches novel platforms. Risk should calibrate effort, not project count. |
| "Expert personas improve research" | Debunked: personas don't improve factual accuracy and can impair recall (EMNLP 2024, replicated 2025). |
| "Hardcoded search queries ensure coverage" | Anti-agentic: violates the ReAct loop. Models formulate better queries dynamically. Token budget (~80% of performance) matters more than query strings. |
| "Rigid output skeletons ensure quality" | Procrustean distortion: format restrictions impair reasoning capacity (EMNLP 2024). Coverage checklists outperform rigid skeletons. |

### The Inversion That Changed Everything

The most load-bearing finding: **how you frame a research question measurably biases what "evidence" a model reports back.** Stating a hypothesis as a belief increases model agreement regardless of actual evidence (Sharma et al. 2023/2025). This single finding justifies the entire framework — unstructured "just ask the AI" research is systematically vulnerable to confirmation bias.

---

## 3. Key Design Decisions & Their Rationale

### Why 5 Blocks Instead of 8 Sections
The 8-section v2.0 anatomy contained 3 actively harmful elements (personas, hardcoded queries, negative bias instructions) and 2 redundant elements. 5 blocks cover the same functional ground with fewer moving parts. See [meta-research/DECISIONS.md D-006](meta-research/DECISIONS.md) for the complete evidence.

### Why Conditional Decomposition Instead of Absolute Isolation
Isolation is not free. Anthropic's own data shows 4–15× token overhead. The fix is conditional: decompose when topics are independent and complex; integrate when topics are coupled and require joint evaluation. Every decomposition MUST include explicit synthesis. See [meta-research/DECISIONS.md D-001](meta-research/DECISIONS.md).

### Why Single-Model Default Instead of Mandatory Triangulation
On factual queries, models converge 70–90% and share 60% of errors. Triangulating buys 3× cost for correlated consensus — the AI equivalent of asking three people who read the same Wikipedia article. Triangulation adds genuine value only when model training priors cause interpretive diversity on subjective trade-offs. See [meta-research/DECISIONS.md D-002](meta-research/DECISIONS.md).

### Why Wardley Mapping Instead of Fixed Ratios
"40% compose / 60% build" is an ungrounded generalization. Standard SaaS may be 80/20; novel algorithmic engines may be 15/85. Wardley evolution mapping routes each component individually: commodity → compose; genesis → build. See [meta-research/DECISIONS.md D-008](meta-research/DECISIONS.md).

---

## 4. The Self-Referential Validation Principle

URP v3.0's most distinctive property: **it was validated by the methodology it prescribes.** The meta-research pipeline used aspect-isolated sessions, evidence grading, and synthesis — then discovered that several of these axioms needed refinement. This recursive self-correction is built into v3.0's DNA through mandatory review triggers and decay conditions on every locked ADR.

The meta-research artifacts in [meta-research/](meta-research/) are not just historical records — they are the empirical proof that the system works, including proof of where it initially got things wrong.
