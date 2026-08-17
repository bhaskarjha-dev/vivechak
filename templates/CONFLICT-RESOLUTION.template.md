# Conflict Resolution Template — URP v3.0
### ACH-Style Falsification Matrix & Structured Divergence Resolution

> **Usage:** Use this template when research sessions or triangulated models
> produce conflicting recommendations on the same architectural decision.

---

## Conflict: `[CHK-NN]` — `[Decision Title]`

**Decision ID:** `[D-NNN]`
**Door Type:** `[one-way / two-way]`
**Conflicting Sessions:** `[T#-## vs T#-##]` or `[Model A vs Model B]`
**Date:** `[YYYY-MM-DD]`

---

## 1. Conflict Summary

| Dimension | Source A | Source B | Source C (if applicable) |
|---|---|---|---|
| **Source** | `[Session/Model ID]` | `[Session/Model ID]` | `[Session/Model ID]` |
| **Recommendation** | `[Option X]` | `[Option Y]` | `[Option Z]` |
| **Core Argument** | `[1-sentence]` | `[1-sentence]` | `[1-sentence]` |
| **Evidence Grade** | `[Grade + modifiers]` | `[Grade + modifiers]` | `[Grade + modifiers]` |

---

## 2. Analysis of Competing Hypotheses (ACH Matrix)

Rate each piece of evidence against each competing hypothesis:
- **CC** = Consistent and Confirmatory
- **C** = Consistent but not diagnostic
- **N** = Neutral / Not applicable
- **I** = Inconsistent (disconfirming)
- **II** = Strongly Inconsistent

| Evidence / Finding | H1: [Option X] | H2: [Option Y] | H3: [Option Z] |
|---|---|---|---|
| `[Evidence 1]` | `[CC/C/N/I/II]` | `[CC/C/N/I/II]` | `[CC/C/N/I/II]` |
| `[Evidence 2]` | `[CC/C/N/I/II]` | `[CC/C/N/I/II]` | `[CC/C/N/I/II]` |
| `[Evidence 3]` | `[CC/C/N/I/II]` | `[CC/C/N/I/II]` | `[CC/C/N/I/II]` |
| `[Evidence N]` | `[CC/C/N/I/II]` | `[CC/C/N/I/II]` | `[CC/C/N/I/II]` |
| **Inconsistency Count** | `[N]` | `[N]` | `[N]` |

**ACH Verdict:** The hypothesis with the **fewest inconsistencies** (not the most confirmations) is preferred. Reject hypotheses with strong inconsistencies first.

---

## 3. Root Cause of Divergence

| Possible Cause | Applies? | Evidence |
|---|---|---|
| **Different assumptions** about workload/scale | `[Yes/No]` | `[Detail]` |
| **Different evaluation criteria** weighting | `[Yes/No]` | `[Detail]` |
| **Stale evidence** in one source | `[Yes/No]` | `[Detail]` |
| **Model training bias** (house style) | `[Yes/No]` | `[Detail]` |
| **Genuinely contested** — experts disagree | `[Yes/No]` | `[Detail]` |

---

## 4. Resolution

**Chosen Option:** `[Option X/Y/Z]`
**Resolution Method:** `[ACH matrix | Evidence grade superiority | Technical spike | Human architect judgment]`

**Rationale:** `[Explain why, mapping to specific evidence and ACH results]`

**Residual Risk:** `[What remains uncertain despite resolution]`

**Reversal Trigger:** `[Condition under which this resolution should be revisited]`

---

## 5. Gary Klein Premortem (One-Way Doors Only)

> *"It is 12 months from now. The system has suffered a catastrophic architectural failure traced directly to this decision. What went wrong?"*

| Failure Scenario | Probability | Mitigation |
|---|---|---|
| `[Scenario 1]` | `[High/Med/Low]` | `[Action]` |
| `[Scenario 2]` | `[High/Med/Low]` | `[Action]` |
| `[Scenario 3]` | `[High/Med/Low]` | `[Action]` |
