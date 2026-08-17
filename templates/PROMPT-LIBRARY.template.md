# Prompt Library Template — URP v3.0
### 5-Block Prompt Anatomy

> **Usage:** For each session in the research pipeline, create one prompt
> using the 5-block structure below. Each prompt should be a complete,
> front-loaded brief suitable for a single AI deep research session.

---

## Session: `[T#-##]` — `[Session Title]`

**Decision Informed:** `[D-XXX: Decision Title]`
**Door Type:** `[one-way / two-way]`
**Dependencies:** `[Hard: T#-## | Soft: T#-## | None]`

---

### BRIEF

We are investigating `[core technical/architectural question]`.

This research will directly inform Architectural Decision `[D-XXX: Title]`.
The target audience is a Principal Architect requiring rigorous, production-grade
technical evaluation with concrete tradeoffs, operational failure modes, and
verified benchmarks — not high-level introductory summaries.

`[2–4 sentences of project-specific context: what the product does, why this
decision matters for this specific product, what's at stake if we get it wrong.]`

---

### SCOPE

- **Temporal Anchor:** Today's date is `[YYYY-MM-DD]`. Focus heavily on
  developments within the last 18–24 months. Flag any findings older than
  `[YYYY-MM-DD minus 2 years]` as potentially stale.
- **In Scope:** `[Explicit boundary 1]`, `[Explicit boundary 2]`,
  `[Explicit boundary 3]`.
- **Out of Scope:** `[Explicit exclusion 1]`, `[Explicit exclusion 2]`.
- **Source Priorities:** Prioritize primary sources (official engineering
  documentation, RFCs, source code repositories, peer-reviewed benchmarks)
  and direct engineering postmortems over secondary aggregators, sponsored
  content, and SEO content farms.

---

### APPROACH

- **Exploration Strategy:** Begin with broad landscape queries to map the
  solution space, then dynamically formulate targeted queries to investigate
  specific trade-offs, failure modes, and benchmarks.
- **Effort Calibration:** Scale search effort to the complexity discovered.
  Spend sufficient search calls to trace genuine technical contradictions
  across sources rather than stopping at the first consensus hit.
- **Epistemic Discipline:**
  - State all necessary assumptions explicitly rather than silently resolving ambiguity.
  - Surface technical disagreements between competing sources rather than smoothing them over.
  - Frame all inquiries neutrally; actively search for disconfirming evidence against favored options.

`[Optional: 1–3 project-specific research directions or angles worth exploring]`

---

### DELIVERABLE

Deliver a structured Markdown document covering the following required checklist:

1. **Executive Summary & Recommendation:** Clear, unambiguous architectural guidance.
2. **Options Evaluation Matrix:** Comparative analysis covering `[Criteria 1]`,
   `[Criteria 2]`, `[Criteria 3]`, operational overhead, and failure modes.
3. **Deep Technical Analysis:** Detailed breakdown of top 2–3 viable contenders
   with architecture diagrams, code/config samples, and production-grade considerations.
4. **Evidentiary Grading:** Assign inline evidence grades (A–E with modifiers
   and verification method) for all factual and benchmark claims.
5. **Open Risks & Reversal Triggers:** Explicit failure conditions under which
   this decision must be revisited.

---

### FORMAT

Deliver the entire output as a single, complete Markdown file artifact.

Include YAML frontmatter:
```yaml
---
id: [T#-##]
title: "[Session Title]"
session_date: [YYYY-MM-DD]
status: final
topic: [topic-tag]
tags: [tag1, tag2, tag3]
informs_decisions: [D-XXX]
confidence: [high/medium/low]
open_questions: [N]
schema_version: "3.0"
---
```

Structure the body with these H2 sections:
1. `## Research Question`
2. `## Key Findings` (3–7 atomic bullets)
3. `## Recommendation` (isolated from rejected options)
4. `## Alternatives Considered`
5. `## Detailed Findings`
6. `## Open Questions & Risks`
7. `## Sources & Evidentiary Ledger`
