// Package mcputil provides the Vivechak MCP server and tool registration.
package mcputil

import (
	"log/slog"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerInstructions describes Vivechak's workflow, grading, sections, and token-efficiency guidelines.
const ServerInstructions = "Vivechak: Evidence-grounded research for technical decisions.\n\n" +
	"## Workflow\n" +
	"1. vivechak_status → orient (or vivechak_init for new workspace)\n" +
	"2. vivechak_prepare_generator → get generator prompt → execute → vivechak_save_plan\n" +
	"3. Loop: vivechak_next_session → research → vivechak_save_session → vivechak_record_decision\n" +
	"4. vivechak_run_gate → verify → implement\n\n" +
	"## Rules\n" +
	"- Follow next_step in every response — it guides the workflow\n" +
	"- ⚡ = parallel sessions; execute concurrently when possible\n" +
	"- Pass content as Markdown with YAML frontmatter (---delimited---)\n\n" +
	"## Evidence Grades (use in session findings)\n" +
	"Grade A: Official docs/specs directly accessed (include URL). " +
	"Grade B: Benchmarks, secondary analysis, vendor data. " +
	"Grade C: Community reports, blog posts, recalled knowledge.\n" +
	"Verification: fetched (URL required) | cached | recalled | human.\n" +
	"Recalled knowledge capped at Grade D.\n\n" +
	"## Session Sections (required in save_session content)\n" +
	"Prior → Research Question → Key Findings (with evidence grades) → " +
	"Recommendation → Alternatives Considered → Open Questions & Risks → " +
	"Discovered Concerns → Delta (belief changes table) → Evidence Ledger table\n\n" +
	"## Token Efficiency\n" +
	"- Don't read templates — next_session provides the session prompt with context\n" +
	"- save_session validates automatically — vivechak_validate is only for dry-run checks\n" +
	"- Use auto_draft_from in record_decision to avoid manual formatting\n\n" +
	"Three scopes: project (→ FAD), decision (→ ADR), comparison (→ WEP matrix)."

// NewServer creates and configures the Vivechak MCP server with all 10 tools.
func NewServer(version string, logger *slog.Logger) *sdkmcp.Server {
	server := sdkmcp.NewServer(
		&sdkmcp.Implementation{Name: "vivechak", Version: version},
		&sdkmcp.ServerOptions{
			Instructions: ServerInstructions,
			Logger:       logger,
		},
	)

	// Register all 10 tools in build order
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

	return server
}
