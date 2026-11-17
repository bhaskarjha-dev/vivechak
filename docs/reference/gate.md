# Phase 0 Exit Gate Reference (Rules G1–G12)

> Generated from `internal/core/gate_rules_spec.go` (`go run ./tools/docgen`). Do not edit by hand.

The Phase 0 Gate is Vivechak's mechanical pre-codebase exit audit (`vivechak_gate` / `vck gate`). It checks that architectural decisions are grounded, reversible or signed off, and traceable before implementation begins.

> **Architectural Disclaimer:** The gate is necessary, not sufficient: it checks structure and the agent's own labels, not the truth of the sources. The sign-off is a typed attestation by the named reviewer; it raises the bar but is not a security boundary.

## 1. Dual-Track Architecture

| Track | Target Decisions | Governing Rules | Philosophy |
|---|---|---|---|
| **Track A** | **Two-Way Doors** (Reversible) | Rule G12 | **Velocity-Calibrated:** Fast spike or convention decision (~70% information) with reversibility path and review trigger. |
| **Track B** | **One-Way Doors** (Hard to Reverse) | Rules G1–G11 | **Rigor-Calibrated:** Grade A/B claims with URLs, `fetched`/`human` provenance, substantive rejected alternatives, red-team challenge record, and interactive human sign-off (`vck signoff`). |

---

## 2. Mechanical Gate Rules Ledger

| Rule | Name | Track | Verification Criteria |
|---|---|---|---|
| **G1** | Plan & DAG Validity | Track B | Plan file parses via its `yaml vivechak-plan` block with no cycles or unknown dependencies, and project pipelines declare exactly one `kind: synthesis` session. |
| **G2** | Session Completeness & Quality | Track B | Every non-synthesis session has `status: complete` and substantive, non-stub research content without blocking validation issues. |
| **G3** | FAD Synthesis Contract | Track B | Project scope has a completed `research/FAD.md` with no blocking issues, and every accepted ADR appears in the Locked Registry. |
| **G4** | Decision Traceability | Track B | Every decision ID referenced by plan or session `informs_decisions` exists as an ADR in `research/`. |
| **G5** | Locked ADR Status | Track B | Target ADRs are locked (`accepted`, `superseded`, or `rejected`); none remain `proposed`. |
| **G6** | One-Way Door Grade A/B Depth | Track B | Accepted One-Way Door ADRs carry ≥2 Grade A/B claims, each on a line or ledger row with its `https://` URL, in the ADR or the sessions that inform it (`[A · single · fresh · direct · fetched](url)`, `[Grade A: url \| fetched]`, `(Grade A · corroborated · fresh \| fetched)`, or an evidence-ledger row whose Grade cell is `A` or `B`). |
| **G7** | Verification Provenance | Track B | Of the claims G6 counts, ≥2 carry the verification label `fetched` or `human` (`human-provided`) in the tag's last field or the ledger row's Verification cell; `cached`, `recalled`, `verified`, `unverified`, and prose words never count. Labels are as reported by the agent; Vivechak does not fetch or verify sources. |
| **G8** | No Placeholders & Rejected Alternatives | Track B | Accepted One-Way Door ADRs contain no unexpanded template placeholders and document substantive rejected alternatives. |
| **G9** | Measurable Review Trigger | Track B | Accepted One-Way Door ADRs declare a concrete, non-placeholder `review_trigger` threshold or review date. |
| **G10** | Adversarial Red-Team Challenge | Track B | Accepted One-Way Door ADRs have a current challenge record (`research/D-NNN-R<k>.md` matching `decision_hash` from `vivechak_review(mode="red_team", id="D-NNN")`) cited with ≥50 chars in `## Challenge Findings`. |
| **G11** | Human Architectural Sign-Off | Track B | Accepted One-Way Door ADRs carry a valid `vck signoff` attestation (`human_reviewed: true`, `reviewed_by`, `reviewed_at`, and matching `reviewed_hash`). |
| **G12** | Two-Way Door Reversibility | Track A | Accepted Two-Way Door ADRs confirm reversibility rationale, non-placeholder content, and a `review_trigger`. |
