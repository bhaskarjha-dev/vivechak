package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	rootDir := "."
	// If run from inside tools/docgen, navigate up to project root
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		if _, err := os.Stat("../../go.mod"); err == nil {
			rootDir = "../.."
		}
	}

	refDir := filepath.Join(rootDir, "docs", "reference")
	if err := os.MkdirAll(refDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create docs/reference: %v\n", err)
		os.Exit(1)
	}

	if err := generateToolsDoc(filepath.Join(refDir, "tools.md")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate tools.md: %v\n", err)
		os.Exit(1)
	}

	if err := generateCodesDoc(filepath.Join(refDir, "codes.md")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate codes.md: %v\n", err)
		os.Exit(1)
	}

	if err := generateGateDoc(filepath.Join(refDir, "gate.md")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate gate.md: %v\n", err)
		os.Exit(1)
	}

	if err := generateHostsDoc(filepath.Join(rootDir, "docs", "hosts.md")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate hosts.md: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Generated reference documentation in docs/reference/ and docs/hosts.md")
}

func generateToolsDoc(path string) error {
	content := `# MCP Tools Reference — Vivechak v0.1.0

Vivechak exposes 9 cohesive Model Context Protocol (MCP) tools for autonomous and pair-programming research workflows:

| # | Tool | Purpose | Primary Scope | Guided Action Returned |
|---|---|---|---|---|
| 1 | ` + "`vivechak_start`" + ` | Initialize engagement and return direct actionable research brief | All Scopes | Search & populate research deliverable |
| 2 | ` + "`vivechak_save`" + ` | Atomically persist comparison matrix, pipeline DAG, research session, or FAD | All Scopes | ` + "`vivechak_save`" + ` or ` + "`vivechak_next`" + ` |
| 3 | ` + "`vivechak_record_decision`" + ` | Persist canonical Architectural Decision Record (` + "`D-NNN.md`" + `) and update registry | Decision & Project | ` + "`vivechak_next`" + ` or ` + "`vivechak_gate`" + ` |
| 4 | ` + "`vivechak_next`" + ` | Topologically resolve next unblocked research session in dependency DAG | Project | Search & save session |
| 5 | ` + "`vivechak_status`" + ` | Inspect workspace state, session counts, progress percentage, and DAG blockers | All Scopes | Next appropriate action |
| 6 | ` + "`vivechak_review`" + ` | Evaluate session quality and surface real-time non-blocking quality coaching (` + "`Q-*`" + `) | Decision & Project | Address quality observations |
| 7 | ` + "`vivechak_amend`" + ` | Amend completed research session or record revised findings | All Scopes | Re-evaluate dependent ADRs |
| 8 | ` + "`vivechak_replan`" + ` | Dynamically modify dependency DAG (insert/prune sessions) based on discoveries | Project | ` + "`vivechak_next`" + ` |
| 9 | ` + "`vivechak_gate`" + ` | Mechanically evaluate Track A (reversibility) and Track B (rigor G1–G12) exit criteria | All Scopes | Proceed to implementation or remediate |

---

## 1. ` + "`vivechak_start`" + `
Initializes a Vivechak research workspace and generates a directly executable research brief.
- **Parameters:**
  - ` + "`scope`" + ` (string, required): ` + "`\"comparison\"`" + `, ` + "`\"decision\"`" + `, or ` + "`\"project\"`" + `.
  - ` + "`question`" + ` (string, required): Technical question, comparison dilemma, or system architecture description.
  - ` + "`project_root`" + ` (string, optional): Absolute path to project workspace directory.
- **Guided Action:** Returns directly executable research brief with required hypotheses and evidence contracts.

---

## 2. ` + "`vivechak_save`" + `
Polymorphically persists research deliverables with automatic atomic writes and validation.
- **Parameters:**
  - ` + "`kind`" + ` (string, required): ` + "`\"comparison\"`" + `, ` + "`\"plan\"`" + `, ` + "`\"session\"`" + `, or ` + "`\"fad\"`" + `.
  - ` + "`id`" + ` (string, required): Session ID (` + "`C-01`" + `, ` + "`D-001`" + `, ` + "`T1-01`" + `, ` + "`SYN-01`" + `).
  - ` + "`content`" + ` (string, required): Complete Markdown artifact content.
  - ` + "`project_root`" + ` (string, optional): Absolute path to project workspace directory.

---

## 3. ` + "`vivechak_record_decision`" + `
Records and locks an Architectural Decision Record (ADR) as canonical ` + "`research/D-NNN.md`" + ` and compiles consolidated ` + "`research/DECISIONS.md`" + `.
- **Parameters:**
  - ` + "`decision_id`" + ` (string, required): Canonical ADR identifier (` + "`D-001`" + `, ` + "`D-002`" + `).
  - ` + "`auto_draft_from`" + ` (string, optional): Session ID to automatically extract rationale, options, and findings from.
  - ` + "`auto_save`" + ` (boolean, optional): Atomically persist the draft ADR to disk immediately.
  - ` + "`project_root`" + ` (string, optional): Absolute path to project workspace directory.

---

## 4. ` + "`vivechak_next`" + `
Resolves the next unblocked research session in a project dependency DAG using Kahn's topological sort.
- **Parameters:**
  - ` + "`project_root`" + ` (string, optional): Absolute path to project workspace directory.
- **Guided Action:** Returns the complete session brief for the highest-priority unblocked session.

---

## 5. ` + "`vivechak_status`" + `
Provides workspace health, detected scope, completed sessions count, and current DAG topology state.
- **Parameters:**
  - ` + "`project_root`" + ` (string, optional): Absolute path to project workspace directory.

---

## 6. ` + "`vivechak_review`" + `
Executes non-blocking evidentiary quality coaching across session contents, detecting grade inflation, missing concerns, and delta confirmation bias.
- **Parameters:**
  - ` + "`session_id`" + ` (string, required): Session ID under review.
  - ` + "`content`" + ` (string, required): Session Markdown content to evaluate.

---

## 7. ` + "`vivechak_amend`" + `
Amends an existing completed research session while archiving prior versions in ` + "`research/.history/`" + `.
- **Parameters:**
  - ` + "`session_id`" + ` (string, required): Session ID to amend.
  - ` + "`content`" + ` (string, required): Updated Markdown content.
  - ` + "`reason`" + ` (string, optional): Justification for amendment.

---

## 8. ` + "`vivechak_replan`" + `
Dynamically modifies the project dependency DAG, adding or pruning sessions based on discovered architectural forks.
- **Parameters:**
  - ` + "`action`" + ` (string, required): ` + "`\"add_session\"`" + ` or ` + "`\"remove_session\"`" + `.
  - ` + "`session_id`" + ` (string, required): Target session identifier.
  - ` + "`session_def`" + ` (string, optional): Session definition table line when adding.

---

## 9. ` + "`vivechak_gate`" + `
Mechanically evaluates Phase 0 Exit Gate criteria prior to writing code.
- **Parameters:**
  - ` + "`track`" + ` (string, optional): ` + "`\"all\"`" + `, ` + "`\"track_a\"`" + ` (Two-Way Doors), or ` + "`\"track_b\"`" + ` (One-Way Doors).
  - ` + "`project_root`" + ` (string, optional): Absolute path to project workspace directory.
`
	return os.WriteFile(path, []byte(content), 0644)
}

func generateCodesDoc(path string) error {
	content := `# Vivechak Error & Diagnostic Codes Reference

Vivechak uses structured, deterministic codes across its MCP server envelope, validation engine, and Phase 0 Gate.

## 1. Operational Error Codes (` + "`E-*`" + `)

| Error Code | Meaning | Remediation |
|---|---|---|
| ` + "`E-ROOT-UNRESOLVED`" + ` | Workspace directory could not be discovered automatically | Pass ` + "`project_root='<absolute_path>'`" + ` or set ` + "`VCK_PROJECT_ROOT`" + ` |
| ` + "`E-INVALID-SCOPE`" + ` | Unknown research scope provided | Specify ` + "`\"comparison\"`" + `, ` + "`\"decision\"`" + `, or ` + "`\"project\"`" + ` |
| ` + "`E-INVALID-ARGUMENT`" + ` | Missing or malformed required argument | Check tool schema requirements |
| ` + "`E-SESSION-NOT-FOUND`" + ` | Target session ID does not exist in DAG or workspace | Verify session ID with ` + "`vivechak_status`" + ` |
| ` + "`E-INTERNAL`" + ` | Internal handler panic caught cleanly without crashing stdio | Inspect stderr logs and report defect |
| ` + "`E-CONFLICT`" + ` | Conflicting session IDs or concurrent edit detected | Review revision history in ` + "`research/.history/`" + ` |
| ` + "`E-LOCKED`" + ` | Target workspace file is locked by a concurrent subagent | Wait for lock release or inspect ` + "`research/.locks/`" + ` |
| ` + "`E-FILE-EXISTS`" + ` | File already exists and overwrite flag was false | Use amend tool or confirm overwrite |

---

## 2. Evidentiary Quality Coaching Codes (` + "`Q-*`" + `)

Surfaced non-blockingly during ` + "`vivechak_save(kind=\"session\")`" + ` and ` + "`vivechak_review`" + `:

| Coaching Code | Issue Detected | Empirical Rationale |
|---|---|---|
| ` + "`Q-GRADE-INFLATION`" + ` | Session contains >60% Grade A claims without URL verification | Prevents rubber-stamping AI recall as authoritative documentation |
| ` + "`Q-CONCERNS-MISSING`" + ` | Discovered Concerns section is empty or <50 characters | Prevents hyper-literalism; ensures adjacent risks are documented |
| ` + "`Q-DELTA-BIAS`" + ` | 100% of prior assumptions were marked as 'Confirmed' | Guards against confirmation bias; real research surfaces surprises |
| ` + "`Q-LOW-FINDINGS`" + ` | Fewer than 2 key findings documented in session | Session lacks sufficient depth to inform architectural commitments |

---

## 3. Structural Validation Codes (` + "`V-*`" + `)

Evaluated during deliverable persistence:

| Code | Validation Check | Required Structure |
|---|---|---|
| ` + "`V-WEP-MISSING`" + ` | Missing Weighted Evaluation Protocol table | Comparison artifacts must include criteria, weights, and scored options |
| ` + "`V-FM-INVALID`" + ` | YAML frontmatter missing or malformed | Deliverables must begin with valid ` + "`---`" + ` YAML frontmatter |
| ` + "`V-REC-UNSUPPORTED`" + ` | Recommendation not grounded in findings | The chosen option must cite findings documented in the session |
| ` + "`V-GRADE-INVALID`" + ` | Evidence grade not in A–E range | All claims must carry recognized grade ` + "`[A-E]`" + ` |
| ` + "`V-PROVENANCE-INVALID`" + ` | Unrecognized provenance tag | Verification must be ` + "`fetched`" + `, ` + "`cached`" + `, ` + "`recalled`" + `, or ` + "`human`" + ` |
`
	return os.WriteFile(path, []byte(content), 0644)
}

func generateGateDoc(path string) error {
	content := `# Phase 0 Exit Gate Reference (Rules G1–G12)

The Phase 0 Gate is Vivechak's mechanical pre-codebase exit audit. It verifies that technical decisions are grounded, reversible, and traceable before any production code is written.

## 1. Dual-Track Architecture

| Track | Target Decisions | Governing Rules | Philosophy |
|---|---|---|---|
| **Track A** | **Two-Way Doors** (Type 2 Decisions) | Checks A1–A4 (Rules G11–G12) | **Velocity-Calibrated:** Fast spike or convention decision (~70% information). Reversibility confirmation and review triggers. |
| **Track B** | **One-Way Doors** (Type 1 Decisions) | Checks B1–B11 (Rules G1–G10) | **Rigor-Calibrated:** High evidentiary threshold (Grade A/B), verified provenance, substantive alternatives, and red-team challenge. |

---

## 2. Mechanical Gate Rules Ledger

| Rule | Name | Track | Verification Criteria |
|---|---|---|---|
| **G1** | Workspace Integrity | All | ` + "`research/`" + ` directory exists with valid ` + "`.vivechak.json`" + ` metadata marker. |
| **G2** | Complete Session Status | Track B | All required DAG sessions must have frontmatter ` + "`status: complete`" + ` (draft saves do not pass). |
| **G3** | FAD Map-Reduce Synthesis | Track B | Greenfield projects require compiled ` + "`research/FAD.md`" + ` synthesizing all domain findings. |
| **G4** | Decision Traceability | All | Every session ` + "`informs_decisions`" + ` ID must correspond to an active ADR in ` + "`research/`" + `. |
| **G5** | Locked ADR Status | All | All identified ADRs must have ` + "`status != proposed`" + ` (` + "`accepted`" + ` or ` + "`superseded`" + `). |
| **G6** | Grade A/B Threshold | Track B | One-Way Doors require ≥2 unique Grade A/B citations supporting the primary decision. |
| **G7** | Verified Provenance | Track B | One-Way Door Grade A/B claims must carry ` + "`fetched`" + ` or ` + "`cached`" + ` provenance (never ` + "`recalled`" + `). |
| **G8** | Substantive Alternatives | Track B | One-Way Doors must document ≥1 rejected alternative with substantive tradeoffs. |
| **G9** | Measurable Review Trigger | All | Every ADR must record a quantitative metric threshold or calendar review trigger. |
| **G10** | Red-Team Challenge | Track B | One-Way Doors must record an adversarial red-team challenge before commitment. |
| **G11** | Human-in-the-Loop Review | Track B | One-Way Doors must have ` + "`human_reviewed: true`" + ` confirmed before final gate passage. |
| **G12** | Two-Way Door Fast Spike | Track A | Two-Way Doors confirm rationale, reversibility path, and 70% information sufficiency. |
`
	return os.WriteFile(path, []byte(content), 0644)
}

func generateHostsDoc(path string) error {
	content := `# Supported AI Editors & Agent Harness Presets

Vivechak auto-detects and configures 21 AI coding editors, agent harnesses, and runtime environments in 1 second using ` + "`vck setup`" + `.

## 1. Quick Setup Commands

| Host / IDE | Shortcut | Target Configuration File | Server Key | 1-Second Setup Command |
|---|---|---|---|---|
| **Cursor** | ` + "`cursor`" + ` | ` + "`.cursor/mcp.json`" + ` or ` + "`~/.cursor/mcp.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup cursor`" + ` |
| **VS Code** | ` + "`vscode`" + ` | ` + "`.vscode/mcp.json`" + ` (Copilot Agent mode) | ` + "`servers`" + ` | ` + "`vck setup vscode`" + ` |
| **Google Antigravity** | ` + "`agy`" + ` | ` + "`~/.gemini/config/mcp_config.json`" + ` or ` + "`.agents/mcp_config.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup agy`" + ` |
| **Claude Code CLI** | ` + "`claude-code`" + ` | ` + "`.mcp.json`" + ` or ` + "`~/.claude.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup claude-code`" + ` |
| **Claude Desktop** | ` + "`claude`" + ` | OS-specific ` + "`claude_desktop_config.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup claude`" + ` |
| **OpenCode** | ` + "`opencode`" + ` | ` + "`opencode.json`" + ` or ` + "`~/.config/opencode/opencode.json`" + ` | ` + "`mcp`" + ` | ` + "`vck setup opencode`" + ` |
| **Zed** | ` + "`zed`" + ` | ` + "`~/.config/zed/settings.json`" + ` (or AppData on Windows) | ` + "`context_servers`" + ` | ` + "`vck setup zed`" + ` |
| **AWS Kiro** | ` + "`kiro`" + ` | ` + "`.kiro/settings/mcp.json`" + ` or ` + "`~/.kiro/settings/mcp.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup kiro`" + ` |
| **Windsurf** | ` + "`windsurf`" + ` | ` + "`~/.codeium/windsurf/mcp_config.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup windsurf`" + ` |
| **Block Goose** | ` + "`goose`" + ` | ` + "`~/.config/goose/config.yaml`" + ` (YAML format) | ` + "`extensions`" + ` | ` + "`vck setup goose`" + ` |
| **GitHub Copilot** | ` + "`copilot`" + ` | ` + "`~/.copilot/mcp-config.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup copilot`" + ` |
| **ByteDance Trae** | ` + "`trae`" + ` | ` + "`.trae/mcp.json`" + ` or ` + "`~/.trae/mcp.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup trae`" + ` |
| **Oh My Pi (OMP)** | ` + "`omp`" + ` | ` + "`.omp/mcp.json`" + ` or ` + "`~/.omp/agent/mcp.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup omp`" + ` |
| **Factory Droid** | ` + "`droid`" + ` | ` + "`.factory/mcp.json`" + ` or ` + "`~/.factory/mcp.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup droid`" + ` |
| **Cline** | ` + "`cline`" + ` | VS Code globalStorage ` + "`cline_mcp_settings.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup cline`" + ` |
| **Roo Code** | ` + "`roo`" + ` | VS Code globalStorage ` + "`cline_mcp_settings.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup roo`" + ` |
| **Cognition Devin** | ` + "`devin`" + ` | ` + "`.devin/mcp_config.local.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup devin`" + ` |
| **OpenHands** | ` + "`openhands`" + ` | ` + "`~/.openhands/mcp.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup openhands`" + ` |
| **Void Editor** | ` + "`void`" + ` | ` + "`.void/mcp.json`" + ` or ` + "`~/.void-editor/mcp.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup void`" + ` |
| **Continue.dev** | ` + "`continue`" + ` | ` + "`.continue/mcpServers/vivechak.json`" + ` | ` + "`mcpServers`" + ` | ` + "`vck setup continue`" + ` |
| **Universal Portable** | ` + "`mcp`" + ` | ` + "`.mcp.json`" + ` in workspace root | ` + "`mcpServers`" + ` | ` + "`vck setup mcp`" + ` |

---

## 2. Universal Auto-Detection

Run without arguments in any project workspace:
` + "```sh" + `
vck setup
` + "```" + `
Vivechak inspects directory markers (` + "`.cursor`" + `, ` + "`.agents`" + `, ` + "`.vscode`" + `, ` + "`.trae`" + `, etc.) and configures the active environment automatically.
`
	return os.WriteFile(path, []byte(content), 0644)
}
