# Phase 0 Exit Gate Template — URP v3.0
### Two-Track Pre-Codebase Gate Checklist

> **Usage:** Complete this checklist before initializing repository scaffolding.
> Route each decision through Track A or Track B based on door type.

---

## Project: `[Project Name]`
**Gate Date:** `[YYYY-MM-DD]`
**Gate Owner:** `[Name]`

---

## Decision Routing Summary

| D-ID | Decision Title | Door Type | Track | Gate Status |
|---|---|---|---|---|
| D-001 | `[Title]` | `[1-way/2-way]` | `[A/B]` | `[PASS/FAIL/PENDING]` |
| D-002 | `[Title]` | `[1-way/2-way]` | `[A/B]` | `[PASS/FAIL/PENDING]` |
| D-NNN | `[Title]` | `[1-way/2-way]` | `[A/B]` | `[PASS/FAIL/PENDING]` |

---

## Track A: Fast-Track Gate (Two-Way Door Decisions)

For each Two-Way Door decision:

- [ ] Decision logged in ADR with `door_type: two-way`
- [ ] Reversibility confirmed — can be changed without major refactoring
- [ ] At least one corroborated source supports the choice

**Track A Result:** `[PASS / FAIL]`

---

## Track B: Rigorous 9-Step Gate (One-Way Door Decisions)

For each One-Way Door decision:

### B1. DAG Closure
- [ ] All required dependency paths have terminated in `status: final` session artifacts
- [ ] No orphaned sessions remain in the DAG

### B2. Contradiction Resolution
- [ ] All cross-model divergences resolved via ACH matrix or explicit trade-off rationale
- [ ] No unresolved `contested` corroboration flags on critical claims

### B3. Evidentiary Threshold
- [ ] Zero uncorroborated Grade C/D/E claims underpin irreversible architectural pillars
- [ ] All critical claims backed by Grade A or B evidence

### B4. Verification Integrity
- [ ] 100% of critical citations carry `verification_method: fetched` or `cached`
- [ ] Zero `recalled` citations support any Type 1 decision

### B5. Rejected Alternatives Documented
- [ ] Every locked ADR explicitly details evaluated and rejected competing options
- [ ] Rejection rationale is causal, not preferential

### B6. Decay Triggers Assigned
- [ ] Every locked ADR contains an explicit `review_trigger` condition or date
- [ ] Review triggers are specific and measurable (not "review when needed")

### B7. Premortem Protocol (Gary Klein, 1989)
- [ ] 30-minute prospective hindsight exercise completed
- [ ] Prompt: *"It is 12 months from now. The system has suffered a catastrophic architectural failure. What caused it?"*
- [ ] Top 3 failure scenarios documented with mitigations
- [ ] Mitigations incorporated into FAD risk register

### B8. Human Architect Review
- [ ] Named Principal Architect has reviewed all Type 1 ADRs
- [ ] Reviewer: `[Name]`
- [ ] Review date: `[YYYY-MM-DD]`

### B9. Founding Architecture Document Sealed
- [ ] FAD compiled with full traceability matrix
- [ ] FAD committed to repository root
- [ ] Repository scaffolding ready to generate

**Track B Result:** `[PASS / FAIL]`

---

## Gate Verdict

| Track | Result | Blocking Issues |
|---|---|---|
| Track A (Two-Way) | `[PASS/FAIL]` | `[None / List issues]` |
| Track B (One-Way) | `[PASS/FAIL]` | `[None / List issues]` |
| **Overall** | **`[PASS/FAIL]`** | |

**Signed Off By:** `[Name]`
**Date:** `[YYYY-MM-DD]`

> **PASS:** Proceed to repository scaffolding and development.
> **FAIL:** Address all blocking issues before re-gating.
