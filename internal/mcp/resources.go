package mcputil

import (
	"context"

	vembed "github.com/bhaskarjha-dev/vivechak/internal/embed"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

const evidenceGradesDoc = `# Vivechak Evidence Grading Taxonomy (Grades A–E)

Every factual claim in Vivechak research must carry an inline evidence grade with modifiers and verification method.

## 1. Base Evidence Grades
- **Grade A (Primary / Authoritative):** Official vendor documentation, specification RFCs, source code implementations, peer-reviewed benchmarks, or official postmortems. Requires qualified provenance (canonical URL or doc title).
- **Grade B (Empirical / Experimental):** Independent third-party reproducible benchmarks, production load tests, empirical user studies, or detailed postmortems from recognized engineering teams.
- **Grade C (Vendor / Motivated):** Marketing whitepapers, sponsored benchmarks, corporate case studies, or developer advocacy posts where the author has a direct financial interest.
- **Grade D (Secondary / Opinion / Model Recall):** Community blogs, StackOverflow answers, tutorials, forum discussions, or unverified AI model recall.
  *Rule:* Recalled knowledge from model weights is ALWAYS capped at Grade D regardless of apparent authority.
- **Grade E (Speculative / Unverified):** Claims without empirical backing, future promises, roadmap statements, or uncorroborated assertions.

## 2. Modifiers
- **Corroboration:**
  - ` + "`[single]`" + `: Claim attested by only one independent source.
  - ` + "`[corroborated]`" + `: Claim independently attested across two or more unrelated sources.
  - ` + "`[contested]`" + `: Conflicting claims exist across credible sources.
- **Recency:**
  - ` + "`[fresh]`" + `: Evidence published within the last 12 months.
  - ` + "`[aging]`" + `: Evidence published between 12 and 24 months ago.
  - ` + "`[stale]`" + `: Evidence older than 24 months (flagged as potentially obsolete).
- **Directness:**
  - ` + "`[direct]`" + `: Measured directly on the exact technology/version under evaluation.
  - ` + "`[indirect]`" + `: Extrapolated from analogous architectures or related systems.

## 3. Verification Methods
- **` + "`fetched`" + `:** Actively retrieved during the session via search or direct web fetch (URL required).
- **` + "`cached`" + `:** Retrieved from verified local repository assets or prior recorded research.
- **` + "`recalled`" + `:** Generated from LLM training weights without active verification (capped at Grade D).
- **` + "`human`" + `:** Provided explicitly by the user or domain principal.
`

const corePrinciplesDoc = `# Vivechak Core Principles (P1–P8)

The 8 foundational tenets governing epistemic research in Vivechak:

| # | Principle | Summary |
|---|---|---|
| **P1** | **Context Architecture Law** | Decompose decisions by attention budget and coupling, never arbitrary counts. Mandate map-reduce synthesis after decomposition. |
| **P2** | **Reversibility-Calibrated Rigor** | Two-Way Doors move fast (spikes or convention). One-Way Doors move carefully (deep falsification, red-teaming, quantitative reversal triggers). |
| **P3** | **Evidentiary Grounding** | Every claim carries an inline grade (A–E), modifiers, and verification method. Recalled knowledge is capped at Grade D. Grade A requires qualified URLs or RFCs. |
| **P4** | **Prescriptive Scope, Dynamic Method** | Prescriptive on WHAT, WHY, and BOUNDARIES. Directional on HOW. No rigid personas or arbitrary search quotas. |
| **P5** | **Commodity-Maximized Composition** | Compose 100% commodity infrastructure. Build 100% custom only for proprietary, differentiating domain logic. |
| **P6** | **Dual-Audience Artifacts** | Hybrid Markdown with YAML frontmatter designed for both human comprehension and automated machine synthesis. |
| **P7** | **Staged Triangulation** | Single-model default → critique probe → full multi-model triangulation only for contested One-Way Doors. |
| **P8** | **Structured Falsification** | Actively seek disconfirming evidence, document rejected alternatives, perform premortems, and establish quantitative review triggers. |
`

// registerResources registers the 8 canonical MCP resources.
func registerResources(server *sdkmcp.Server) {
	resources := []struct {
		uri         string
		name        string
		title       string
		description string
		readFn      func() (string, error)
	}{
		{
			uri:         "vivechak://templates/decisions",
			name:        "template_decisions",
			title:       "ADR Template (Architecture Decision Record)",
			description: "Canonical YAML frontmatter Architecture Decision Record format for recording technical decisions.",
			readFn: func() (string, error) {
				b, err := vembed.ReadTemplate("DECISIONS.template.md")
				return string(b), err
			},
		},
		{
			uri:         "vivechak://templates/comparison",
			name:        "template_comparison",
			title:       "WEP Comparison Template",
			description: "Weighted Evaluation Protocol matrix template for structured multi-option evaluation.",
			readFn: func() (string, error) {
				b, err := vembed.ReadTemplate("COMPARISON-SESSION.template.md")
				return string(b), err
			},
		},
		{
			uri:         "vivechak://templates/fad",
			name:        "template_fad",
			title:       "Founding Architecture Document Template (FAD)",
			description: "Map-Reduce terminal synthesis architecture template for greenfield systems.",
			readFn: func() (string, error) {
				b, err := vembed.ReadTemplate("FOUNDING-ARCHITECTURE.template.md")
				return string(b), err
			},
		},
		{
			uri:         "vivechak://templates/conflict-resolution",
			name:        "template_conflict_resolution",
			title:       "Conflict Resolution Template",
			description: "Analysis of Competing Hypotheses (ACH) template for resolving contested technical claims.",
			readFn: func() (string, error) {
				b, err := vembed.ReadTemplate("CONFLICT-RESOLUTION.template.md")
				return string(b), err
			},
		},
		{
			uri:         "vivechak://templates/gate",
			name:        "template_gate",
			title:       "Phase 0 Exit Gate Template",
			description: "Two-track (Track A reversibility & Track B checks B1–B11) pre-codebase exit criteria template.",
			readFn: func() (string, error) {
				b, err := vembed.ReadTemplate("PHASE-0-GATE.template.md")
				return string(b), err
			},
		},
		{
			uri:         "vivechak://templates/session",
			name:        "template_session",
			title:       "Research Session Template",
			description: "Standard research session output format with YAML frontmatter, Prior, Delta, and Evidence Ledger.",
			readFn: func() (string, error) {
				b, err := vembed.ReadTemplate("SESSION.template.md")
				return string(b), err
			},
		},
		{
			uri:         "vivechak://methodology/grades",
			name:        "methodology_grades",
			title:       "Evidence Grading Reference Card (Grades A–E)",
			description: "Taxonomy of evidence grades A–E, modifiers (corroboration, recency, directness), and verification methods.",
			readFn: func() (string, error) {
				return evidenceGradesDoc, nil
			},
		},
		{
			uri:         "vivechak://methodology/principles",
			name:        "methodology_principles",
			title:       "Core Principles Reference Card (P1–P8)",
			description: "The 8 epistemic grounding principles: Context Architecture, Reversibility-Calibrated Rigor, Evidentiary Grounding, etc.",
			readFn: func() (string, error) {
				return corePrinciplesDoc, nil
			},
		},
	}

	for _, item := range resources {
		readContent := item.readFn
		targetURI := item.uri
		server.AddResource(
			&sdkmcp.Resource{
				URI:         targetURI,
				Name:        item.name,
				Title:       item.title,
				Description: item.description,
				MIMEType:    "text/markdown",
			},
			func(_ context.Context, req *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
				if req.Params == nil || req.Params.URI != targetURI {
					return nil, sdkmcp.ResourceNotFoundError(targetURI)
				}
				content, err := readContent()
				if err != nil {
					return nil, err
				}
				return &sdkmcp.ReadResourceResult{
					Contents: []*sdkmcp.ResourceContents{
						{
							URI:      targetURI,
							MIMEType: "text/markdown",
							Text:     content,
						},
					},
				}, nil
			},
		)
	}
}
