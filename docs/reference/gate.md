# Phase 0 Exit Gate Reference (Rules G1–G12)

The Phase 0 Gate is Vivechak's mechanical pre-codebase exit audit. It verifies that technical decisions are grounded, reversible, and traceable before any production code is written.

## 1. Dual-Track Architecture

| Track | Target Decisions | Governing Rules | Philosophy |
|---|---|---|---|
| **Track A** | **Two-Way Doors** (Type 2 Decisions) | Checks A1–A4 (Rule G12 + common rules) | **Velocity-Calibrated:** Fast spike or convention decision (~70% information). Reversibility confirmation and review triggers. |
| **Track B** | **One-Way Doors** (Type 1 Decisions) | Checks B1–B11 (Rules G1–G11) | **Rigor-Calibrated:** High evidentiary threshold (Grade A/B), verified provenance, substantive alternatives, human sign-off, and red-team challenge. |

---

## 2. Mechanical Gate Rules Ledger

| Rule | Name | Track | Verification Criteria |
|---|---|---|---|
| **G1** | Workspace Integrity | All | `research/` directory exists with valid `.vivechak.json` metadata marker. |
| **G2** | Complete Session Status | Track B | All required DAG sessions must have frontmatter `status: complete` (draft saves do not pass). |
| **G3** | FAD Map-Reduce Synthesis | Track B | Greenfield projects require compiled `research/FAD.md` synthesizing all domain findings. |
| **G4** | Decision Traceability | All | Every session `informs_decisions` ID must correspond to an active ADR in `research/`. |
| **G5** | Locked ADR Status | All | All identified ADRs must have `status != proposed` (`accepted` or `superseded`). |
| **G6** | Grade A/B Threshold | Track B | One-Way Doors require ≥2 unique Grade A/B citations supporting the primary decision. |
| **G7** | Verified Provenance | Track B | One-Way Door Grade A/B claims must carry `fetched` or `cached` provenance (never `recalled`). |
| **G8** | Substantive Alternatives | Track B | One-Way Doors must document ≥1 rejected alternative with substantive tradeoffs. |
| **G9** | Measurable Review Trigger | All | Every ADR must record a quantitative metric threshold or calendar review trigger. |
| **G10** | Red-Team Challenge | Track B | One-Way Doors must record an adversarial red-team challenge before commitment. |
| **G11** | Human-in-the-Loop Review | Track B | One-Way Doors must have `human_reviewed: true` confirmed before final gate passage. |
| **G12** | Two-Way Door Fast Spike | Track A | Two-Way Doors confirm rationale, reversibility path, and 70% information sufficiency. |
