package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/workspace"
)

// WorkspaceFS provides an abstract read-only interface for inspecting workspace files and directories.
type WorkspaceFS interface {
	ReadFile(relPath string) ([]byte, error)
	ListDir(relPath string) ([]os.DirEntry, error)
	Stat(relPath string) (os.FileInfo, error)
}

type dirFS struct {
	dir string
}

func (d dirFS) resolve(relPath string) string {
	if d.dir == "" {
		if relPath == "" || relPath == "." {
			return "."
		}
		return relPath
	}
	if relPath == "" || relPath == "." {
		return d.dir
	}
	return filepath.Join(d.dir, relPath)
}

func (d dirFS) ReadFile(relPath string) ([]byte, error) {
	return os.ReadFile(d.resolve(relPath))
}

func (d dirFS) ListDir(relPath string) ([]os.DirEntry, error) {
	return os.ReadDir(d.resolve(relPath))
}

func (d dirFS) Stat(relPath string) (os.FileInfo, error) {
	return os.Stat(d.resolve(relPath))
}

// DirFS returns a simple WorkspaceFS adapter backed by the given directory.
func DirFS(dir string) WorkspaceFS {
	return dirFS{dir: dir}
}

// Workspace re-exports from internal/workspace for core package consumers.

const (
	ResearchDir    = workspace.ResearchDir
	SessionsDir    = workspace.SessionsDir
	ComparisonsDir = workspace.ComparisonsDir
	DecisionsDir   = workspace.DecisionsDir
	TemplatesDir   = workspace.TemplatesDir
	HistoryDir     = workspace.HistoryDir
	LocksDir       = workspace.LocksDir

	PipelineFile  = workspace.PipelineFile
	DecisionsFile = workspace.DecisionsFile
	FADFile       = workspace.FADFile
	GateFile      = workspace.GateFile
	MetadataFile  = workspace.MetadataFile
)

type WorkspaceMeta = workspace.WorkspaceMeta
type WorkspaceInfo = workspace.WorkspaceInfo

var TemplatesToCopy = workspace.TemplatesToCopy

func TemplatesForScope(scope Scope) []string { return workspace.TemplatesForScope(scope) }
func IsSpecialResearchFile(name string) bool { return workspace.IsSpecialResearchFile(name) }
func SafeRoot(dirPath string) error         { return workspace.SafeRoot(dirPath) }
func IsSystemOrAppDir(dirPath string) bool  { return workspace.IsSystemOrAppDir(dirPath) }
func SetLastResolvedWorkspace(ws string)    { workspace.SetLastResolvedWorkspace(ws) }
func GetLastResolvedWorkspace() string      { return workspace.GetLastResolvedWorkspace() }
func ResetLastResolvedWorkspace()           { workspace.ResetLastResolvedWorkspace() }
func ResolveWorkspace(explicit string) (string, error) {
	return workspace.ResolveWorkspace(explicit)
}
func WorkspaceExists(root string) bool           { return workspace.WorkspaceExists(root) }
func InspectWorkspace(root string) WorkspaceInfo { return workspace.InspectWorkspace(root) }

var sessionIDPrefixRe = regexp.MustCompile(`^(?i)(T\d+-\d+|S\d+-\d+|D-[A-Za-z0-9_-]+-S\d+|SYN-\d+|C-\d+|C\d+-\d+|CMP-\d+|COMP-\d+|FAD)(?:[-_].*)?$`)

// ScanCompletedSessions reads the sessions directory and FAD file using workspace confinement,
// returning completed IDs. Authoritative IDs come from YAML frontmatter (session_id or id fields), falling back to filename stem.
// If frontmatter lacks an ID, session ID prefixes from filename stems (e.g. T1-01 from T1-01-database.md) are recognized to prevent DAG deadlocks.
func ScanCompletedSessions(ws WorkspaceFS, knownIDs ...string) (map[string]bool, error) {
	completed := map[string]bool{}
	entries, err := ws.ListDir(SessionsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading sessions directory: %w", err)
		}
		// Sessions dir not yet created — expected on fresh workspace
	} else {
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
				continue
			}

			relPath := filepath.Join(SessionsDir, e.Name())
			var fm Frontmatter
			if data, err := ws.ReadFile(relPath); err == nil {
				if parsed, _, err := ParseFrontmatter(data); err == nil {
					fm = parsed
				}
			}

			// L2Block check: If session is explicitly saved as draft, do NOT mark completed
			if fm != nil && strings.EqualFold(fm.GetString("status"), "draft") {
				continue
			}

			stem := strings.TrimSuffix(e.Name(), ".md")
			id := ExtractSessionID(e.Name(), fm)

			candidates := []string{stem}
			if id != "" {
				completed[id] = true
				completed[strings.ToUpper(id)] = true
				if !strings.EqualFold(id, stem) {
					candidates = append(candidates, id)
				}
			}

			// Match session ID prefixes against known DAG sessions to prevent deadlocks
			// even when files or frontmatter IDs include title slugs (e.g. T1-01-database-selection)
			for _, cand := range candidates {
				lowerCand := strings.ToLower(cand)
				for _, kid := range knownIDs {
					lowerKid := strings.ToLower(kid)
					if lowerCand == lowerKid || strings.HasPrefix(lowerCand, lowerKid+"-") || strings.HasPrefix(lowerCand, lowerKid+"_") {
						completed[kid] = true
						completed[strings.ToUpper(kid)] = true
					}
				}
				if sub := sessionIDPrefixRe.FindStringSubmatch(cand); len(sub) > 1 {
					completed[sub[1]] = true
					completed[strings.ToUpper(sub[1])] = true
				}
			}
		}
	}

	// Also check for FAD at the research root
	if data, err := ws.ReadFile(FADFile); err == nil {
		var fm Frontmatter
		if parsed, _, err := ParseFrontmatter(data); err == nil {
			fm = parsed
		}
		if fm == nil || !strings.EqualFold(fm.GetString("status"), "draft") {
			id := ExtractSessionID(filepath.Base(FADFile), fm)
			if id != "" {
				completed[id] = true
			}
			completed["FAD"] = true
			completed["SYN-01"] = true
			completed["SYN"] = true
			for _, kid := range knownIDs {
				if IsSynthesisSession(kid) {
					completed[kid] = true
				}
			}
		}
	}

	return completed, nil
}

// GateDecisionItem represents an architectural decision record evaluated at the gate.
type GateDecisionItem struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	DoorType      string `json:"door_type"`
	Track         string `json:"track"`
	Status        string `json:"status"`
	HumanReviewed bool   `json:"human_reviewed"`
}

// CollectGateDecisions collects all recorded decisions from ADR files and DECISIONS.md.
func CollectGateDecisions(ws WorkspaceFS) []GateDecisionItem {
	type decisionSource struct {
		file string
		data []byte
	}
	var sources []decisionSource

	if entries, err := ws.ListDir(ResearchDir); err == nil {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".md") {
				continue
			}
			if IsSpecialResearchFile(name) {
				continue
			}
			relPath := filepath.Join(ResearchDir, name)
			isCandidate := strings.HasPrefix(strings.ToUpper(name), "D-")
			dData, err := ws.ReadFile(relPath)
			if err != nil || len(dData) == 0 {
				continue
			}
			if !isCandidate {
				if fm, _, err := ParseFrontmatter(dData); err == nil && fm != nil {
					if !fm.Has("door_type") {
						continue
					}
				} else {
					continue
				}
			}
			sources = append(sources, decisionSource{file: relPath, data: dData})
		}
	}

	// Scan consolidated DECISIONS.md alongside standalone files.
	// The processedIDs map below guarantees standalone files take precedence on ID collisions.
	if decData, err := ws.ReadFile(DecisionsFile); err == nil && len(decData) > 0 {
		sources = append(sources, decisionSource{file: DecisionsFile, data: decData})
	}

	var decisions []GateDecisionItem
	processedIDs := make(map[string]bool)

	checkDecisionFM := func(fm Frontmatter) {
		if fm == nil || (!fm.Has("id") && !fm.Has("decision_id")) {
			return
		}
		id := fm.GetString("id")
		if id == "" {
			id = fm.GetString("decision_id")
		}
		if id == "" || strings.HasPrefix(strings.ToUpper(id), "CHK-") || fm.Has("conflicting_sources") {
			return
		}
		upperID := strings.ToUpper(id)
		if processedIDs[upperID] {
			return
		}
		processedIDs[upperID] = true

		title := fm.GetString("title")
		if title == "" {
			title = id
		}
		doorType := fm.GetString("door_type")
		if doorType == "" {
			doorType = "two-way"
		}
		track := "A"
		if strings.EqualFold(doorType, "one-way") {
			track = "B"
		}
		status := strings.ToUpper(strings.TrimSpace(fm.GetString("status")))
		switch status {
		case "ACCEPTED":
			status = "PASS"
		case "":
			status = "PENDING"
		}

		reviewed := strings.EqualFold(strings.TrimSpace(fm.GetString("human_reviewed")), "true")

		decisions = append(decisions, GateDecisionItem{
			ID:            id,
			Title:         title,
			DoorType:      doorType,
			Track:         track,
			Status:        status,
			HumanReviewed: reviewed,
		})
	}

	anchoredRe := regexp.MustCompile(`(?s)<!-- DECISION:\s*([A-Za-z0-9_-]+)\s*-->\s*(.*?)\s*<!-- /DECISION:\s*[A-Za-z0-9_-]+\s*-->`)

	for _, src := range sources {
		chunk := src.data
		parsedAny := false

		for _, m := range anchoredRe.FindAllSubmatch(chunk, -1) {
			if fm, _, err := ParseFrontmatter(m[2]); err == nil && fm != nil {
				if !fm.Has("id") && !fm.Has("decision_id") {
					fm["id"] = string(m[1])
				}
				checkDecisionFM(fm)
				parsedAny = true
			}
		}

		if !parsedAny {
			for _, block := range ExtractUnanchoredDecisionBlocks(chunk) {
				if fm, _, err := ParseFrontmatter(block); err == nil && fm != nil {
					if fm.Has("id") || fm.Has("decision_id") || fm.Has("door_type") {
						checkDecisionFM(fm)
						parsedAny = true
					}
				}
			}
		}

		if !parsedAny {
			if fm, _, err := ParseFrontmatter(chunk); err == nil && fm != nil {
				checkDecisionFM(fm)
			}
		}
	}

	return decisions
}

// hasDecisionTableRow checks if markdown content contains a table row recording a decision ID.
func hasDecisionTableRow(content string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") {
			parts := strings.Split(trimmed, "|")
			if len(parts) >= 3 {
				col1 := strings.TrimSpace(parts[1])
				upper := strings.ToUpper(col1)
				if strings.HasPrefix(upper, "D-") || strings.HasPrefix(upper, "[D-") {
					return true
				}
			}
		}
	}
	return false
}
