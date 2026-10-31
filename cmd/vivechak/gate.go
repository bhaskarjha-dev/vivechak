package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

func runGateWithArgs(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gate", flag.ContinueOnError)
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		fmt.Fprintln(stdout, "Usage: vivechak gate [flags] [workspace_path]  (or: vck gate [flags] [workspace_path])")
		fmt.Fprintln(stdout, "")
		fmt.Fprintln(stdout, "Evaluate Phase 0 Gate criteria mechanically before codebase implementation.")
		fmt.Fprintln(stdout, "")
		fmt.Fprintln(stdout, "Flags:")
		fmt.Fprintln(stdout, "  --track string   Track filter: 'all', 'track_a', or 'track_b' (default \"all\")")
		fmt.Fprintln(stdout, "  --format string  Output format: 'text' or 'json' (default \"text\")")
		fmt.Fprintln(stdout, "  --verbose, -v    Show detailed issue breakdown")
		return 0
	}

	fs.SetOutput(stdout)

	var track string
	var format string
	var jsonFlag bool
	var verbose bool

	fs.StringVar(&track, "track", "all", "Track filter: 'all', 'track_a', or 'track_b'")
	fs.StringVar(&format, "format", "text", "Output format: 'text' or 'json'")
	fs.BoolVar(&jsonFlag, "json", false, "Output format: json")
	fs.BoolVar(&verbose, "verbose", false, "Show detailed issue breakdown")
	fs.BoolVar(&verbose, "v", false, "Show detailed issue breakdown (shorthand)")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 1
	}

	track = strings.ToLower(strings.TrimSpace(track))
	if track != "all" && track != "track_a" && track != "track_b" {
		fmt.Fprintf(stderr, "Error: invalid track filter %q (must be 'all', 'track_a', or 'track_b')\n", track)
		return 1
	}

	var explicit string
	if fs.NArg() > 0 {
		explicit = fs.Arg(0)
	}

	if explicit != "" && core.IsSystemOrAppDir(explicit) {
		fmt.Fprintf(stderr, "Error: %q is an application or system directory, not a research workspace.\n", explicit)
		return 1
	}

	workspace, err := core.ResolveWorkspace(explicit)
	if err != nil {
		fmt.Fprintf(stderr, "Error: workspace not initialized at %s (%v)\n", explicit, err)
		return 1
	}

	if core.IsSystemOrAppDir(workspace) {
		fmt.Fprintf(stderr, "Error: %q is an application or system directory, not a research workspace.\n", workspace)
		return 1
	}

	if !core.WorkspaceExists(workspace) {
		fmt.Fprintf(stderr, "Error: workspace not initialized at %s\n", workspace)
		return 1
	}

	ws, err := store.OpenWorkspace(workspace)
	if err != nil {
		fmt.Fprintf(stderr, "Error opening workspace: %v\n", err)
		return 1
	}
	defer ws.Close()

	info := core.InspectWorkspace(workspace)
	gateResult := core.EvaluateGate(ws, info, workspace)

	// Persist PHASE-0-GATE.md if applicable
	if gateResult.Verdict != "not_applicable" {
		projectName := core.DetectProjectName(ws, workspace)
		gateContent := core.RenderGateReport(ws, workspace, projectName, gateResult)
		if len(gateContent) > 0 {
			var existingGate []byte
			if d, err := ws.ReadFile(core.GateFile); err == nil {
				existingGate = d
			}
			mergedContent := core.MergeGateArtifactContent(existingGate, gateContent)
			_ = ws.MkdirAll(core.ResearchDir, 0o755)
			_ = store.WriteFileAtomic(ws.Root(), core.GateFile, mergedContent, 0o644)
		}
	}

	passed := true
	if track == "track_a" {
		for _, r := range gateResult.Rules {
			if r.ID == "G12" && r.Status == "FAIL" {
				passed = false
			}
		}
	} else if track == "track_b" {
		for _, r := range gateResult.Rules {
			if r.ID != "G12" && r.Status == "FAIL" {
				passed = false
			}
		}
	} else {
		passed = (gateResult.Verdict == "PASS" || gateResult.Verdict == "not_applicable")
	}

	gateStatus := "PASS"
	if !passed {
		gateStatus = "FAIL"
	}

	if jsonFlag || format == "json" {
		outMap := map[string]any{
			"workspace":  workspace,
			"scope":      info.Scope,
			"track":      track,
			"status":     gateStatus,
			"passed":     passed,
			"verdict":    gateResult.Verdict,
			"target":     gateResult.Target,
			"rules":      gateResult.Rules,
			"advisories": gateResult.Advisories,
			"warnings":   gateResult.Warnings,
			"next_step":  gateResult.NextStep,
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(outMap)
	} else {
		fmt.Fprintf(stdout, "Phase 0 Gate: %s\n", strings.ToUpper(gateResult.Verdict))
		trackAStatus := "PASS"
		trackBStatus := "PASS"
		for _, r := range gateResult.Rules {
			if r.ID == "G12" && r.Status == "FAIL" {
				trackAStatus = "FAIL"
			}
			if strings.HasPrefix(r.ID, "G") {
				num, _ := strconv.Atoi(strings.TrimPrefix(r.ID, "G"))
				if num >= 6 && num <= 11 && r.Status == "FAIL" {
					trackBStatus = "FAIL"
				}
			}
		}
		if track == "track_a" || track == "all" {
			fmt.Fprintf(stdout, "  • Track A (Two-Way Doors): %s\n", trackAStatus)
		}
		if track == "track_b" || track == "all" {
			fmt.Fprintf(stdout, "  • Track B (One-Way Doors): %s\n", trackBStatus)
		}
		for _, r := range gateResult.Rules {
			fmt.Fprintf(stdout, "  • %s: %s - %s\n", r.ID, r.Status, r.Reason)
		}

		if len(gateResult.Advisories) > 0 {
			fmt.Fprintln(stdout, "\nAdvisories:")
			for _, adv := range gateResult.Advisories {
				fmt.Fprintf(stdout, "  • %s\n", adv)
			}
		}

		fmt.Fprintf(stdout, "\nNext Step: %s\n", gateResult.NextStep)
	}

	if !passed {
		return 1
	}
	return 0
}
