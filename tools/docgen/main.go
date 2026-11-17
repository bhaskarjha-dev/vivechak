package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/hosts"
	mcputil "github.com/bhaskarjha-dev/vivechak/internal/mcp"
)

func main() {
	rootDir := "."
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		if _, err := os.Stat("../../go.mod"); err == nil {
			rootDir = "../.."
		}
	}

	refDir := filepath.Join(rootDir, "docs", "reference")
	if err := os.MkdirAll(refDir, 0o755); err != nil {
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

	if err := generateEvidenceBlocks(rootDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate evidence blocks: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Generated reference documentation in docs/reference/, docs/hosts.md, and evidence blocks from single-source registries")
}

func escapeMDCell(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

func generateToolsDoc(path string) error {
	specs := mcputil.RegisteredToolSpecs()
	var b strings.Builder
	b.WriteString("# MCP Tools Reference — Vivechak v0.1.0\n\n")
	b.WriteString("> Generated from `internal/mcp` tool registrations (`go run ./tools/docgen`). Do not edit by hand.\n\n")
	fmt.Fprintf(&b, "Vivechak exposes %d Model Context Protocol (MCP) tools:\n\n", len(specs))
	b.WriteString("| # | Tool | Required Parameters | Description |\n")
	b.WriteString("|---|---|---|---|\n")
	for i, s := range specs {
		var req []string
		for _, p := range s.Params {
			if p.Required {
				req = append(req, "`"+p.Name+"`")
			}
		}
		reqStr := "*(none)*"
		if len(req) > 0 {
			reqStr = strings.Join(req, ", ")
		}
		fmt.Fprintf(&b, "| %d | `%s` | %s | %s |\n", i+1, s.Name, reqStr, escapeMDCell(s.Description))
	}

	for i, s := range specs {
		b.WriteString("\n---\n\n")
		fmt.Fprintf(&b, "## %d. `%s`\n\n", i+1, s.Name)
		fmt.Fprintf(&b, "%s\n\n", s.Description)
		if s.Annotations != nil {
			destructive := false
			if s.Annotations.DestructiveHint != nil {
				destructive = *s.Annotations.DestructiveHint
			}
			openWorld := false
			if s.Annotations.OpenWorldHint != nil {
				openWorld = *s.Annotations.OpenWorldHint
			}
			fmt.Fprintf(&b, "- **Annotations:** `readOnlyHint=%t`, `idempotentHint=%t`, `destructiveHint=%t`, `openWorldHint=%t`\n",
				s.Annotations.ReadOnlyHint, s.Annotations.IdempotentHint, destructive, openWorld)
		}
		b.WriteString("- **Parameters:**\n")
		for _, p := range s.Params {
			typeStr := p.Type
			if p.Type == "array" && p.ItemType != "" {
				typeStr = p.ItemType + "[]"
			}
			reqStr := "optional"
			if p.Required {
				reqStr = "required"
			}
			enumSuffix := ""
			if len(p.Enum) > 0 {
				var quoted []string
				for _, e := range p.Enum {
					quoted = append(quoted, fmt.Sprintf("`%q`", e))
				}
				enumSuffix = " Allowed: " + strings.Join(quoted, ", ") + "."
			}
			fmt.Fprintf(&b, "  - `%s` (%s, %s): %s%s\n", p.Name, typeStr, reqStr, p.DocText(), enumSuffix)
		}
	}

	return writeIfChanged(path, []byte(b.String()))
}

func generateCodesDoc(path string) error {
	codes := core.AllCodes()
	var b strings.Builder
	b.WriteString("# Vivechak Diagnostic & Error Codes Reference\n\n")
	b.WriteString("> Generated from `internal/core/codes.go` (`go run ./tools/docgen`). Do not edit by hand.\n\n")
	fmt.Fprintf(&b, "Vivechak emits %d structured codes across its MCP envelope, structural/plan validator, warning diagnostics, and quality coach:\n\n", len(codes))

	sections := []struct {
		level string
		title string
		intro string
	}{
		{"error", "1. Operational Error Codes (`E-*`)", "Returned on failed operations (`ok: false`, `isError: true`) with remediation in `fix`."},
		{"validation", "2. Structural & Plan Validation Codes (`V-*`)", "Emitted during artifact or plan validation (`vivechak_save`, `vivechak_record_decision`, `vivechak_review`)."},
		{"warning", "3. Operational & Advisory Warning Codes (`W-*`)", "Returned in `warnings` (`[{code, message, fix}]`) without failing the request."},
		{"coaching", "4. Cognitive Quality Coaching Codes (`Q-*`)", "Surfaced non-blockingly on session saves and reviews to highlight evidentiary gaps."},
	}

	for idx, sec := range sections {
		if idx > 0 {
			b.WriteString("\n---\n\n")
		}
		fmt.Fprintf(&b, "## %s\n\n%s\n\n", sec.title, sec.intro)
		b.WriteString("| Code | Message | Remediation (`fix`) |\n")
		b.WriteString("|---|---|---|\n")
		for _, c := range codes {
			if c.Level != sec.level {
				continue
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s |\n", c.Code, escapeMDCell(c.Message), escapeMDCell(c.Fix))
		}
	}

	return writeIfChanged(path, []byte(b.String()))
}

func generateGateDoc(path string) error {
	rules := core.AllGateRules()
	var b strings.Builder
	b.WriteString("# Phase 0 Exit Gate Reference (Rules G1–G12)\n\n")
	b.WriteString("> Generated from `internal/core/gate_rules_spec.go` (`go run ./tools/docgen`). Do not edit by hand.\n\n")
	b.WriteString("The Phase 0 Gate is Vivechak's mechanical pre-codebase exit audit (`vivechak_gate` / `vck gate`). It checks that architectural decisions are grounded, reversible or signed off, and traceable before implementation begins.\n\n")
	b.WriteString("> **Architectural Disclaimer:** The gate is necessary, not sufficient: it checks structure and the agent's own labels, not the truth of the sources. The sign-off is a typed attestation by the named reviewer; it raises the bar but is not a security boundary.\n\n")
	b.WriteString("## 1. Dual-Track Architecture\n\n")
	b.WriteString("| Track | Target Decisions | Governing Rules | Philosophy |\n")
	b.WriteString("|---|---|---|---|\n")
	b.WriteString("| **Track A** | **Two-Way Doors** (Reversible) | Rule G12 | **Velocity-Calibrated:** Fast spike or convention decision (~70% information) with reversibility path and review trigger. |\n")
	b.WriteString("| **Track B** | **One-Way Doors** (Hard to Reverse) | Rules G1–G11 | **Rigor-Calibrated:** Grade A/B claims with URLs, `fetched`/`human` provenance, substantive rejected alternatives, red-team challenge record, and interactive human sign-off (`vck signoff`). |\n\n")
	b.WriteString("---\n\n")
	b.WriteString("## 2. Mechanical Gate Rules Ledger\n\n")
	b.WriteString("| Rule | Name | Track | Verification Criteria |\n")
	b.WriteString("|---|---|---|---|\n")
	for _, r := range rules {
		fmt.Fprintf(&b, "| **%s** | %s | %s | %s |\n", r.ID, escapeMDCell(r.Name), escapeMDCell(r.Track), r.Criteria)
	}

	return writeIfChanged(path, []byte(b.String()))
}

func generateHostsDoc(path string) error {
	presets := hosts.All()
	var b strings.Builder
	b.WriteString("# Supported AI Editors & Agent Harness Presets\n\n")
	b.WriteString("> Generated from `internal/hosts/presets.go` (`go run ./tools/docgen`). Do not edit by hand.\n\n")
	fmt.Fprintf(&b, "`vck setup <preset>` configures %d hosts whose MCP config file and key are verified against each host's own documentation (URLs in `cmd/vivechak/presets.go` and `internal/hosts/presets.go`). Any other host: `vck setup <path-to-its-mcp-config>`.\n\n", len(presets))
	b.WriteString("## 1. Quick Setup Commands\n\n")
	b.WriteString("| Host / IDE | Shortcut | Target Configuration File | Server Key | Setup Command | Host Docs |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, p := range presets {
		sc := p.Shortcut()
		fmt.Fprintf(&b, "| **%s** | `%s` | %s | `%s` | `vck setup %s` | [Docs](%s) |\n",
			escapeMDCell(p.DisplayName), sc, escapeMDCell(p.ConfigSummary), p.ServerKey, sc, p.DocURL)
	}
	b.WriteString("\nHosts without a preset (for example Cline, Kiro, Continue, Trae, GitHub Copilot CLI): `vck setup <path-to-its-mcp-config>`.\n\n")
	b.WriteString("---\n\n")
	b.WriteString("## 2. Universal Auto-Detection\n\n")
	b.WriteString("Run without arguments in any project workspace:\n")
	b.WriteString("```sh\nvck setup\n```\n")
	b.WriteString("Vivechak inspects directory markers (`.cursor`, `.agents`, `.vscode`, `.mcp.json`, `opencode.json`, `.opencode`, `.goose`, `.roo`) and configures the detected host.\n")

	return writeIfChanged(path, []byte(b.String()))
}
