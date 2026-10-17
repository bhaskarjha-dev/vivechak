// Package mcputil provides the Vivechak MCP server and tool registration.
package mcputil

import (
	"log/slog"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerInstructions describes Vivechak's mandatory trigger rules, workflow, grading, sections, and token guidelines.
const ServerInstructions = "Vivechak (vck) — Evidence-grounded research for technical decisions.\n\n" +
	"## MANDATORY TRIGGER RULES (Invoke First):\n" +
	"- ALWAYS invoke vivechak_prepare_generator for comparisons ('A vs B', 'which to choose'), tech evaluations, architectural decisions, or project planning.\n" +
	"- DO NOT answer comparisons or decisions from ungrounded memory. Skip only for direct code/debugging/syntax.\n\n" +
	"## Workflow\n" +
	"1. vivechak_status → orient (or vivechak_init with project_root)\n" +
	"2. vivechak_prepare_generator → get prompt → execute → vivechak_save_plan\n" +
	"3. Loop: vivechak_next_session → research → vivechak_save_session → vivechak_challenge (mode='red_team') → vivechak_record_decision\n" +
	"4. vivechak_run_gate → verify → implement\n\n" +
	"## Rules\n" +
	"- Pass project_root to vivechak_init; follow next_step in every response\n" +
	"- ⚡ = parallel sessions; execute concurrently when possible\n" +
	"- Use vivechak_challenge for stress-tests; vivechak_replan for DAG mutations; vivechak_visualize for graph\n" +
	"- Content format: YAML frontmatter Markdown\n\n" +
	"## Evidence Grades (use in session findings)\n" +
	"Grade A: Primary/Authoritative (specs/docs + URL).\n" +
	"Grade B: Empirical/Experimental (benchmarks, postmortems).\n" +
	"Grade C: Vendor/Motivated (whitepapers).\n" +
	"Grade D: Secondary/Opinion (blogs, tutorials, recalled knowledge capped at D).\n" +
	"Grade E: Speculative.\n" +
	"Verification: fetched (URL required) | cached | recalled | human.\n\n" +
	"## Session Sections (required in save_session content)\n" +
	"Prior → Research Question → Key Findings (with evidence grades) → " +
	"Recommendation → Alternatives Considered → Open Questions & Risks → " +
	"Discovered Concerns → Delta (belief changes table) → Evidence Ledger table\n\n" +
	"## Token Efficiency\n" +
	"- next_session injects prompt with context; don't read templates\n" +
	"- save_session validates automatically; vivechak_validate for dry-runs\n" +
	"- Use auto_draft_from in record_decision to avoid manual formatting\n\n" +
	"Three scopes: project (→ FAD), decision (→ ADR), comparison (→ WEP matrix)."

// NewServer creates and configures the Vivechak MCP server with all 13 tools.
func NewServer(version string, logger *slog.Logger) *sdkmcp.Server {
	server := sdkmcp.NewServer(
		&sdkmcp.Implementation{Name: "vivechak", Version: version},
		&sdkmcp.ServerOptions{
			Instructions: ServerInstructions,
			Logger:       logger,
		},
	)

	// Register all 13 tools in build order
	registerInit(server)
	registerPrepareGenerator(server)
	registerSavePlan(server)
	registerStatus(server)
	registerNextSession(server)
	registerSaveSession(server)
	registerRecordDecision(server)
	registerValidate(server)
	registerRunGate(server)
	registerAmendSession(server)
	registerChallenge(server)
	registerReplan(server)
	registerVisualize(server)

	return server
}
