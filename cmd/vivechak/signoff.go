package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

func runSignoff(args []string) error {
	// Strictly verify interactive TTY (os.ModeCharDevice check)
	fi, err := os.Stdin.Stat()
	if err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		return fmt.Errorf("E-TTY-REQUIRED: vck signoff requires an interactive terminal session (os.ModeCharDevice). Autonomous agent execution is prohibited.")
	}

	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		return fmt.Errorf("usage: vck signoff <decision-id> (e.g. vck signoff D-001)")
	}
	decisionID := strings.TrimSpace(args[0])

	var explicitWs string
	if len(args) > 1 {
		explicitWs = strings.TrimSpace(args[1])
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

	fmt.Println("=======================================================")
	fmt.Printf("Vivechak Interactive Architecture Decision Sign-off\n")
	fmt.Println("=======================================================")
	fmt.Printf("Decision ID:      %s\n", decisionID)
	fmt.Printf("File:             %s\n", targetFile)
	fmt.Printf("Title:            %s\n", title)
	fmt.Printf("Door Type:        %s\n", doorType)
	fmt.Printf("Review Trigger:   %s\n", reviewTrigger)
	fmt.Println("-------------------------------------------------------")

	bodyStr := string(body)
	if rejected := core.ExtractSection(bodyStr, "rejected alternatives", "rejected alternative"); rejected != "" {
		fmt.Printf("Rejected Alternatives:\n%s\n", strings.TrimSpace(rejected))
		fmt.Println("-------------------------------------------------------")
	}

	fmt.Printf("Type 'CONFIRM' to sign off on Decision %s: ", decisionID)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading confirmation: %w", err)
	}
	input = strings.TrimSpace(input)
	if input != "CONFIRM" {
		return fmt.Errorf("signoff cancelled: expected 'CONFIRM', got %q", input)
	}

	// Detect reviewer
	reviewer := getGitUserName()
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

	fmt.Printf("\n✓ Decision %s successfully signed off by %s.\nSaved to %s.\n", decisionID, reviewer, savedPath)
	return nil
}

func getGitUserName() string {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}
