# Architectural Decision Record Template — URP v3.0
### YAML Frontmatter ADR with Evidence Traceability

> **Usage:** Create one file per decision: `decisions/D-NNN-[slug].md`

---

```yaml
---
id: D-NNN
title: "[Decision Title]"
status: proposed               # proposed | accepted | rejected | deprecated | superseded
door_type: one-way             # one-way | two-way (sets required evidentiary bar)
date: YYYY-MM-DD
confidence: medium             # high | medium | low (decoupled from evidence grade)
evidence_refs: []              # E-NNN IDs supporting this decision
informed_by_sessions: []       # T#-## session IDs
supersedes: null               # D-NNN ID this supersedes, or null
superseded_by: null            # D-NNN ID that supersedes this, or null
amends: null                   # D-NNN ID this partially updates, or null
review_trigger: "[Condition or date for mandatory re-evaluation]"
tags: []
authored_by: "[agent-id or human name]"
human_reviewed: false          # Mandatory true for one-way doors before acceptance
schema_version: "3.0"
---
```

# D-NNN: [Decision Title]

## Context & Problem Statement

[Describe the technical requirements, architectural forces, and constraints
that necessitate this decision. Include quantitative requirements where
available (throughput targets, latency limits, data volumes, user counts).]

## Evaluated Options

1. **Option 1: [Name]** — Evidence: [E-NNN] (Grade [X] · [modifiers] | [verification])
2. **Option 2: [Name]** — Evidence: [E-NNN] (Grade [X] · [modifiers] | [verification])
3. **Option 3: [Name]** — Evidence: [E-NNN] (Grade [X] · [modifiers] | [verification])

## Decision Outcome

**Chosen Option:** Option [N] — [Name].

### Rationale

[Causal rationale mapping evidence directly to decision criteria. Explain
WHY the chosen option outperforms competitors for THIS specific project's
constraints, not generic superiority claims.]

## Rejected Alternatives & Tradeoffs

- **[Option Name]:** Rejected because [specific causal reason with evidence reference].
- **[Option Name]:** Rejected because [specific causal reason with evidence reference].

## Failure Modes & Reversal Triggers

- If [specific measurable condition], trigger immediate review of [migration path].
- Scheduled review: [date or condition from review_trigger].
