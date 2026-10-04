package mcputil

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ChallengeInput holds arguments for vivechak_challenge.
type ChallengeInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"workspace root path (optional)"`
	SessionID   string `json:"session_id"             jsonschema:"required: which session to challenge (e.g. T1-01)"`
	Mode        string `json:"mode,omitempty"         jsonschema:"challenge mode: 'red_team' (default), 'evidence_audit', or 'cross_session'"`
}

func registerChallenge(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_challenge",
			Title: "Challenge Research Session",
			Description: "Generate adversarial challenge prompts for a completed research session. " +
				"Returns structured prompts that YOU execute to stress-test your own findings. " +
				"Modes: 'red_team' (disconfirming search + assumption audit + steel-man alternative + premortem), " +
				"'evidence_audit' (per-citation URL and primary source verification), 'cross_session' (inter-session consistency). " +
				"This tool does not perform the research — it generates structured prompts for you to execute. " +
				"Save challenge findings or updates via vivechak_amend_session.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    true,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleChallenge,
	)
}

func handleChallenge(ctx context.Context, _ *sdkmcp.CallToolRequest, in ChallengeInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_challenge"

	root, err := core.ResolveWorkspace(in.ProjectRoot)
	if err != nil {
		return ErrorResult(tool, err, "Initialize a workspace first with vivechak_init.")
	}

	sessionID := strings.TrimSpace(in.SessionID)
	if sessionID == "" {
		return ErrorResult(tool, fmt.Errorf("session_id is required"),
			"Provide the session ID to challenge (e.g. 'T1-01').")
	}

	mode := strings.ToLower(strings.TrimSpace(in.Mode))
	if mode == "" {
		mode = string(core.ModeRedTeam)
	}

	ws, err := store.OpenWorkspace(root)
	if err != nil {
		return ErrorResult(tool, err, "Ensure the workspace directory exists.")
	}
	defer ws.Close()

	// Locate session file
	sessionPath := filepath.Join(core.SessionsDir, sessionID+".md")
	sessionData, err := ws.ReadFile(sessionPath)
	if err != nil {
		// Attempt prefix matching for slugged session files
		found := false
		if entries, lErr := ws.ListDir(core.SessionsDir); lErr == nil {
			idUpper := strings.ToUpper(sessionID)
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
					continue
				}
				name := strings.TrimSuffix(e.Name(), ".md")
				nameUpper := strings.ToUpper(name)
				if nameUpper == idUpper || strings.HasPrefix(nameUpper, idUpper+"-") || strings.HasPrefix(nameUpper, idUpper+"_") {
					sessionPath = filepath.Join(core.SessionsDir, e.Name())
					sessionData, err = ws.ReadFile(sessionPath)
					if err == nil {
						found = true
						break
					}
				}
			}
		}
		if !found {
			return ErrorResult(tool, fmt.Errorf("session %s not found in %s", sessionID, core.SessionsDir),
				"Ensure the session has been saved with vivechak_save_session before challenging it.")
		}
	}

	// Read all completed sessions if cross_session mode is requested
	allSessions := make(map[string][]byte)
	if mode == string(core.ModeCrossSession) {
		if entries, lErr := ws.ListDir(core.SessionsDir); lErr == nil {
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
					continue
				}
				sid := strings.TrimSuffix(e.Name(), ".md")
				if parts := strings.Split(sid, "-"); len(parts) >= 2 {
					sid = parts[0] + "-" + parts[1]
				}
				data, rErr := ws.ReadFile(filepath.Join(core.SessionsDir, e.Name()))
				if rErr == nil && len(data) > 0 {
					allSessions[sid] = data
				}
			}
		}
	}

	challengeResult, err := core.GenerateChallenges(sessionData, sessionID, core.ChallengeMode(mode), allSessions)
	if err != nil {
		return ErrorResult(tool, err, "Verify the session content is non-empty and formatted correctly.")
	}

	nextStep := "Execute these challenge prompts in your research session. If findings change or new risks emerge, save the updates via vivechak_amend_session."
	if len(challengeResult.Prompts) == 0 {
		nextStep = "Proceed to vivechak_record_decision or vivechak_next_session."
	}

	env := Envelope{
		Success:  true,
		Message:  fmt.Sprintf("Generated %d challenge prompts for session %s (mode: %s)", len(challengeResult.Prompts), sessionID, mode),
		Data:     challengeResult,
		NextStep: nextStep,
		Meta:     NewMeta(tool),
	}

	return env.ToResult()
}
