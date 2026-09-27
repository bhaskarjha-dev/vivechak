# Architectural Decision Record Template — Vivechak v1.1
### YAML Frontmatter ADR with Evidence Traceability

> **Usage:** Add one entry per decision to your project's `DECISIONS.md` registry file.
> Each entry uses the YAML frontmatter + Markdown body format below.
> For projects with many decisions, individual files (`decisions/D-NNN-[slug].md`) may be used instead.

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
review_date: null               # YYYY-MM-DD — calendar date for scheduled review (set BEFORE outcome is known)
prediction: null                # Optional: predicted outcome at decision time (for calibration tracking)
tags: []
authored_by: "[agent-id or human name]"
human_reviewed: false          # Mandatory true for one-way doors before acceptance
schema_version: "1.1"
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

## Calibration Record (Post-Review)

> Fill this section ONLY during scheduled review or when the review_trigger fires.
> This closes the feedback loop — without it, decision calibration can never improve.

- **Review date:** [YYYY-MM-DD]
- **Outcome vs prediction:** [Did the predicted outcome materialize? What actually happened?]
- **Decision quality assessment:** [Was this the right call? Would you make the same decision with current information?]
- **Calibration note:** [Was the confidence level appropriate? Over-confident? Under-confident?]
- **Action:** [No change | Amend (issue D-NNN) | Supersede (issue D-NNN)]
