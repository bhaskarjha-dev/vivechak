package core

// GateRuleSpec describes a canonical Phase 0 Exit Gate rule (G1–G12).
type GateRuleSpec struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Track    string `json:"track"` // "Track B" or "Track A"
	Criteria string `json:"criteria"`
}

var canonicalGateRules = []GateRuleSpec{
	{
		ID:       "G1",
		Name:     "Plan & DAG Validity",
		Track:    "Track B",
		Criteria: "Plan file parses via its `yaml vivechak-plan` block with no cycles or unknown dependencies, and project pipelines declare exactly one `kind: synthesis` session.",
	},
	{
		ID:       "G2",
		Name:     "Session Completeness & Quality",
		Track:    "Track B",
		Criteria: "Every non-synthesis session has `status: complete` and substantive, non-stub research content without blocking validation issues.",
	},
	{
		ID:       "G3",
		Name:     "FAD Synthesis Contract",
		Track:    "Track B",
		Criteria: "Project scope has a completed `research/FAD.md` with no blocking issues, and every accepted ADR appears in the Locked Registry.",
	},
	{
		ID:       "G4",
		Name:     "Decision Traceability",
		Track:    "Track B",
		Criteria: "Every decision ID referenced by plan or session `informs_decisions` exists as an ADR in `research/`.",
	},
	{
		ID:       "G5",
		Name:     "Locked ADR Status",
		Track:    "Track B",
		Criteria: "Target ADRs are locked (`accepted`, `superseded`, or `rejected`); none remain `proposed`.",
	},
	{
		ID:       "G6",
		Name:     "One-Way Door Grade A/B Depth",
		Track:    "Track B",
		Criteria: "Accepted One-Way Door ADRs carry ≥2 Grade A/B claims, each on a line or ledger row with its `https://` URL, in the ADR or the sessions that inform it (`[A · single · fresh · direct · fetched](url)`, `[Grade A: url \\| fetched]`, `(Grade A · corroborated · fresh \\| fetched)`, or an evidence-ledger row whose Grade cell is `A` or `B`).",
	},
	{
		ID:       "G7",
		Name:     "Verification Provenance",
		Track:    "Track B",
		Criteria: "Of the claims G6 counts, ≥2 carry the verification label `fetched` or `human` (`human-provided`) in the tag's last field or the ledger row's Verification cell; `cached`, `recalled`, `verified`, `unverified`, and prose words never count. Labels are as reported by the agent; Vivechak does not fetch or verify sources.",
	},
	{
		ID:       "G8",
		Name:     "No Placeholders & Rejected Alternatives",
		Track:    "Track B",
		Criteria: "Accepted One-Way Door ADRs contain no unexpanded template placeholders and document substantive rejected alternatives.",
	},
	{
		ID:       "G9",
		Name:     "Measurable Review Trigger",
		Track:    "Track B",
		Criteria: "Accepted One-Way Door ADRs declare a concrete, non-placeholder `review_trigger` threshold or review date.",
	},
	{
		ID:       "G10",
		Name:     "Adversarial Red-Team Challenge",
		Track:    "Track B",
		Criteria: "Accepted One-Way Door ADRs have a current challenge record (`research/D-NNN-R<k>.md` matching `decision_hash` from `vivechak_review(mode=\"red_team\", id=\"D-NNN\")`) cited with ≥50 chars in `## Challenge Findings`.",
	},
	{
		ID:       "G11",
		Name:     "Human Architectural Sign-Off",
		Track:    "Track B",
		Criteria: "Accepted One-Way Door ADRs carry a valid `vck signoff` attestation (`human_reviewed: true`, `reviewed_by`, `reviewed_at`, and matching `reviewed_hash`).",
	},
	{
		ID:       "G12",
		Name:     "Two-Way Door Reversibility",
		Track:    "Track A",
		Criteria: "Accepted Two-Way Door ADRs confirm reversibility rationale, non-placeholder content, and a `review_trigger`.",
	},
}

// AllGateRules returns the canonical G1–G12 Phase 0 Exit Gate rule specifications in order.
func AllGateRules() []GateRuleSpec {
	out := make([]GateRuleSpec, len(canonicalGateRules))
	copy(out, canonicalGateRules)
	return out
}
