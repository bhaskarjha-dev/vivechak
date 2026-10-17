// Package mcputil provides the Vivechak MCP server and tool registration.
package mcputil

import (
	"log/slog"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerInstructions describes Vivechak's mandatory trigger rules, scope workflows, grading, sections, and token guidelines.
const ServerInstructions = "Vivechak (vck) — Evidence-grounded research for technical decisions across all stages.\n\n" +
	"## MANDATORY TRIGGER RULES (Invoke First):\n" +
	"- ALWAYS invoke vivechak_prepare_generator for comparisons ('A vs B'), tech evaluations, decisions, or project planning.\n" +
	"- DO NOT answer comparisons or decisions from ungrounded memory.\n" +
	"- SKIP Vivechak for direct code authoring, syntax, debugging, or refactoring (write code directly).\n\n" +
	"## Workflows by Scope:\n" +
	"- Comparison ('A vs B', 1 session): prepare_generator(scope='comparison') → execute → save_plan (WEP matrix). 5m! No DAG/gate.\n" +
	"- Decision (ADR, 1–3 sessions): prepare_generator(scope='decision') → execute → record_decision (ADR). 30m! No project DAG.\n" +
	"- Project (Full Pipeline): init → prepare_generator(scope='project') → save_plan → session loop → FAD → run_gate.\n\n" +
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
	"- Use auto_draft_from in record_decision to avoid manual formatting\n"

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
