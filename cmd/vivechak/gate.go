package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	var verbose bool

	fs.StringVar(&track, "track", "all", "Track filter: 'all', 'track_a', or 'track_b'")
	fs.StringVar(&format, "format", "text", "Output format: 'text' or 'json'")
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
		fmt.Fprintf(stderr, "Workspace resolution failed: %v\n", err)
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

	gatePassed := gateResult.Passed
	gateStatus := gateResult.Status
	switch track {
	case "track_a":
		gatePassed = gateResult.TrackAChecks.Passed
		if gatePassed {
			gateStatus = "PASS"
		} else {
			gateStatus = "FAIL"
		}
	case "track_b":
		gatePassed = gateResult.TrackBChecks.Passed
		if gatePassed {
			gateStatus = "PASS"
		} else {
			gateStatus = "FAIL"
		}
	}

	if format == "json" {
		outMap := map[string]any{
			"workspace":         workspace,
			"scope":             info.Scope,
			"track":             track,
			"status":            gateStatus,
			"passed":            gatePassed,
			"structural_checks": gateResult.StructuralChecks,
			"track_a_checks":    gateResult.TrackAChecks,
			"track_b_checks":    gateResult.TrackBChecks,
			"advisories":        gateResult.Advisories,
			"warnings":          gateResult.Warnings,
			"next_step":         gateResult.NextStep,
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(outMap)
	} else {
		fmt.Fprintf(stdout, "Phase 0 Gate: %s (Track: %s, Scope: %s)\n", gateStatus, track, info.Scope)
		fmt.Fprintf(stdout, "  • Track A (Two-Way Doors): %s (%s)\n", passStr(gateResult.TrackAChecks.Passed), gateResult.TrackAChecks.Score)
		fmt.Fprintf(stdout, "  • Track B (One-Way Doors): %s (%s)\n", passStr(gateResult.TrackBChecks.Passed), gateResult.TrackBChecks.Score)
		fmt.Fprintf(stdout, "  • Structural Completeness: %s (%s)\n", passStr(gateResult.StructuralChecks.Passed), gateResult.StructuralChecks.Score)

		if verbose || !gatePassed {
			if len(gateResult.TrackAChecks.Issues) > 0 && (track == "all" || track == "track_a") {
				fmt.Fprintln(stdout, "\nTrack A Issues:")
				for _, iss := range gateResult.TrackAChecks.Issues {
					fmt.Fprintf(stdout, "  - %s\n", iss)
				}
			}
			if len(gateResult.TrackBChecks.Issues) > 0 && (track == "all" || track == "track_b") {
				fmt.Fprintln(stdout, "\nTrack B Issues:")
				for _, iss := range gateResult.TrackBChecks.Issues {
					fmt.Fprintf(stdout, "  - %s\n", iss)
				}
			}
			if len(gateResult.StructuralChecks.Issues) > 0 && track == "all" {
				fmt.Fprintln(stdout, "\nStructural Issues:")
				for _, iss := range gateResult.StructuralChecks.Issues {
					fmt.Fprintf(stdout, "  - %s\n", iss)
				}
			}
		}

		if len(gateResult.Advisories) > 0 && verbose {
			fmt.Fprintln(stdout, "\nAdvisories:")
			for _, adv := range gateResult.Advisories {
				fmt.Fprintf(stdout, "  • %s\n", adv)
			}
		}

		fmt.Fprintf(stdout, "\nNext Step: %s\n", gateResult.NextStep)
	}

	if gatePassed {
		return 0
	}
	return 1
}

func passStr(p bool) string {
	if p {
		return "PASS"
	}
	return "FAIL"
}
