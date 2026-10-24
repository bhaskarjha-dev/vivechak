package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

func runSignoff(args []string, streams ...io.Writer) error {
	var stdout io.Writer = os.Stdout
	if len(streams) > 0 && streams[0] != nil {
		stdout = streams[0]
	}

	var explicitWs string
	var force bool
	var decisionID string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--help" || arg == "-h" || arg == "help" {
			fmt.Fprintln(stdout, "Usage: vivechak signoff [flags] <decision-id>  (or: vck signoff [flags] <decision-id>)")
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "Interactively sign off on an Architectural Decision (One-Way Door).")
			fmt.Fprintln(stdout, "Requires an interactive terminal session (os.ModeCharDevice).")
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "Flags:")
			fmt.Fprintln(stdout, "  --project-root <path>  Path to project workspace root (default: current directory)")
			fmt.Fprintln(stdout, "  --root <path>          Path to project workspace root (shorthand)")
			fmt.Fprintln(stdout, "  -f, --force            Sign off without prompt confirmation")
			return nil
		} else if arg == "--project-root" || arg == "-project-root" || arg == "--root" || arg == "-root" {
			if i+1 < len(args) {
				explicitWs = args[i+1]
				i++
			}
		} else if strings.HasPrefix(arg, "--project-root=") {
			explicitWs = strings.TrimPrefix(arg, "--project-root=")
		} else if strings.HasPrefix(arg, "--root=") {
			explicitWs = strings.TrimPrefix(arg, "--root=")
		} else if arg == "--force" || arg == "-force" || arg == "-f" || arg == "--f" {
			force = true
		} else if !strings.HasPrefix(arg, "-") {
			if decisionID == "" {
				decisionID = arg
			} else if explicitWs == "" {
				explicitWs = arg
			}
		}
	}

	// Strictly verify interactive TTY (os.ModeCharDevice check)
	isTestAllowed := os.Getenv("VCK_TEST_ALLOW_NON_TTY_SIGNOFF") == "1"
	isTestForceNonTTY := os.Getenv("VCK_TEST_FORCE_NON_TTY_SIGNOFF") == "1"
	fi, err := os.Stdin.Stat()
	if isTestForceNonTTY || (!isTestAllowed && (err != nil || (fi.Mode()&os.ModeCharDevice) == 0)) {
		return fmt.Errorf("E-TTY-REQUIRED: vck signoff requires an interactive terminal session (os.ModeCharDevice); autonomous agent execution is prohibited")
	}

	decisionID = strings.TrimSpace(decisionID)
	if decisionID == "" {
		return fmt.Errorf("usage: vck signoff [flags] <decision-id> (e.g. vck signoff D-001)")
	}

	root, err := core.ResolveWorkspace(explicitWs)
	if err != nil {
		return fmt.Errorf("resolving workspace: %w", err)
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return fmt.Errorf("opening workspace at %s: %w", root, err)
	}
	defer ws.Close()

	// Find the ADR
	candidates := []string{
		filepath.Join(core.DecisionsDir, decisionID+".md"),
		filepath.Join(core.ResearchDir, decisionID+".md"),
	}
	var targetFile string
	var adrData []byte
	for _, cand := range candidates {
		if data, err := ws.ReadFile(cand); err == nil && len(data) > 0 {
			targetFile = cand
			adrData = data
			break
		}
	}
	if targetFile == "" {
		searchDirs := []string{core.DecisionsDir, core.ResearchDir}
		for _, sDir := range searchDirs {
			if entries, err := ws.ListDir(sDir); err == nil {
				idUpper := strings.ToUpper(decisionID)
				for _, e := range entries {
					if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
						continue
					}
					name := strings.TrimSuffix(e.Name(), ".md")
					nameUpper := strings.ToUpper(name)
					if nameUpper == idUpper || strings.HasPrefix(nameUpper, idUpper+"-") || strings.HasPrefix(nameUpper, idUpper+"_") {
						cand := filepath.Join(sDir, e.Name())
						if data, rErr := ws.ReadFile(cand); rErr == nil && len(data) > 0 {
							targetFile = cand
							adrData = data
							break
						}
					}
				}
			}
			if targetFile != "" {
				break
			}
		}
	}

	if targetFile == "" {
		return fmt.Errorf("decision %s not found in workspace", decisionID)
	}

	fm, body, err := core.ParseFrontmatter(adrData)
	if err != nil {
		return fmt.Errorf("parsing frontmatter: %w", err)
	}

	title := fm.GetString("title")
	doorType := fm.GetString("door_type")
	reviewTrigger := fm.GetString("review_trigger")
	if reviewTrigger == "" {
		reviewTrigger = "None specified"
	}

	fmt.Fprintln(stdout, "=======================================================")
	fmt.Fprintf(stdout, "Vivechak Interactive Architecture Decision Sign-off\n")
	fmt.Fprintln(stdout, "=======================================================")
	fmt.Fprintf(stdout, "Decision ID:      %s\n", decisionID)
	fmt.Fprintf(stdout, "File:             %s\n", targetFile)
	fmt.Fprintf(stdout, "Title:            %s\n", title)
	fmt.Fprintf(stdout, "Door Type:        %s\n", doorType)
	fmt.Fprintf(stdout, "Review Trigger:   %s\n", reviewTrigger)
	fmt.Fprintln(stdout, "-------------------------------------------------------")

	bodyStr := string(body)
	if rejected := core.ExtractSection(bodyStr, "rejected alternatives", "rejected alternative"); rejected != "" {
		fmt.Fprintf(stdout, "Rejected Alternatives:\n%s\n", strings.TrimSpace(rejected))
		fmt.Fprintln(stdout, "-------------------------------------------------------")
	}

	if !force {
		fmt.Fprintf(stdout, "Type 'CONFIRM' to sign off on Decision %s: ", decisionID)
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading confirmation: %w", err)
		}
		input = strings.TrimSpace(input)
		if input != "CONFIRM" {
			return fmt.Errorf("signoff cancelled: expected 'CONFIRM', got %q", input)
		}
	} else {
		fmt.Fprintf(stdout, "Force flag specified. Proceeding with sign-off on Decision %s...\n", decisionID)
	}

	// Detect reviewer
	reviewer := getGitUserName(root)
	if reviewer == "" {
		reviewer = os.Getenv("USER")
		if reviewer == "" {
			reviewer = os.Getenv("USERNAME")
		}
		if reviewer == "" {
			reviewer = "Human Architect"
		}
	}

	savedPath, err := core.SignoffDecision(ws, decisionID, reviewer)
	if err != nil {
		return fmt.Errorf("signing off decision: %w", err)
	}

	fmt.Fprintf(stdout, "\n✓ Decision %s successfully signed off by %s.\nSaved to %s.\n", decisionID, reviewer, savedPath)
	return nil
}

func getGitUserName(workspace string) string {
	cmd := exec.Command("git", "config", "user.name")
	if workspace != "" {
		cmd.Dir = workspace
	}
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}
