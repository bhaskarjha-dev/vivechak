package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// WorkspaceFS provides an abstract read-only interface for inspecting workspace files and directories.
type WorkspaceFS interface {
	ReadFile(relPath string) ([]byte, error)
	ListDir(relPath string) ([]os.DirEntry, error)
	Stat(relPath string) (os.FileInfo, error)
}

// WorkspaceLayout describes the directory structure of a Vivechak research workspace.
const (
	// ResearchDir is the root research directory within a project.
	ResearchDir = "research"
	// SessionsDir holds individual research session outputs.
	SessionsDir = "research/sessions"
	// TemplatesDir holds the copied template files.
	TemplatesDir = "research/templates"

	// PipelineFile is the research pipeline DAG (project scope).
	PipelineFile = "research/RESEARCH-PIPELINE.md"
	// DecisionsFile is the decision registry.
	DecisionsFile = "research/DECISIONS.md"
	// FADFile is the Founding Architecture Document (terminal synthesis).
	FADFile = "research/FAD.md"
	// GateFile is the Phase 0 Exit Gate record.
	GateFile = "research/PHASE-0-GATE.md"
	// MetadataFile stores workspace metadata (scope, init timestamp).
	MetadataFile = "research/.vivechak.json"
)

// WorkspaceMeta holds metadata written during vivechak_init.
type WorkspaceMeta struct {
	Scope     Scope  `json:"scope"`
	CreatedAt string `json:"created_at,omitempty"`
}

// TemplatesToCopy lists the template files that vivechak_init copies
// into the workspace's research/templates/ directory.
var TemplatesToCopy = []string{
	"DECISIONS.template.md",
	"CONFLICT-RESOLUTION.template.md",
	"COMPARISON-SESSION.template.md",
	"FOUNDING-ARCHITECTURE.template.md",
	"PHASE-0-GATE.template.md",
	"SESSION.template.md",
}

// TemplatesForScope returns the template files appropriate for the given scope.
func TemplatesForScope(scope Scope) []string {
	switch scope {
	case ScopeDecision:
		return []string{
			"DECISIONS.template.md",
			"CONFLICT-RESOLUTION.template.md",
		}
	case ScopeComparison:
		return []string{
			"COMPARISON-SESSION.template.md",
		}
	default: // ScopeProject
		return TemplatesToCopy // all 6
	}
}

// IsSpecialResearchFile returns true if name is a reserved pipeline or non-ADR artifact
// in research/ (e.g. DECISIONS.md, FAD.md, FOUNDING-ARCHITECTURE.md, PHASE-0-GATE.md,
// RESEARCH-PIPELINE.md, *-plan.md, *-comparison.md, *-conflict-resolution.md).
func IsSpecialResearchFile(name string) bool {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, "-plan.md") ||
		strings.HasSuffix(lower, "-comparison.md") ||
		strings.HasSuffix(lower, "-conflict-resolution.md") ||
		strings.EqualFold(name, "DECISIONS.md") ||
		strings.EqualFold(name, "FAD.md") ||
		strings.EqualFold(name, "FOUNDING-ARCHITECTURE.md") ||
		strings.EqualFold(name, "PHASE-0-GATE.md") ||
		strings.EqualFold(name, "RESEARCH-PIPELINE.md") {
		return true
	}
	return false
}

func isSystemOrAppDirNormalized(checkPath string) bool {
	isDirOrChild := func(p, target string) bool {
		return p == target || strings.HasPrefix(p, target+"/")
	}

	// Windows application and system paths
	if strings.Contains(checkPath, "/appdata/local/programs") ||
		isDirOrChild(checkPath, "/program files") ||
		isDirOrChild(checkPath, "/program files (x86)") ||
		isDirOrChild(checkPath, "/programdata") ||
		isDirOrChild(checkPath, "/windows") {
		return true
	}

	// macOS application bundles (*.app or *.app/Contents/...) and system directories
	if strings.Contains(checkPath, ".app/") || strings.HasSuffix(checkPath, ".app") ||
		isDirOrChild(checkPath, "/applications") ||
		isDirOrChild(checkPath, "/system") ||
		isDirOrChild(checkPath, "/library") {
		return true
	}

	// Linux / Unix system binary and config paths (exact binary directories only)
	if isDirOrChild(checkPath, "/usr/bin") ||
		isDirOrChild(checkPath, "/usr/local/bin") ||
		isDirOrChild(checkPath, "/usr/sbin") ||
		isDirOrChild(checkPath, "/usr/lib") ||
		isDirOrChild(checkPath, "/bin") ||
		isDirOrChild(checkPath, "/sbin") ||
		isDirOrChild(checkPath, "/etc") {
		return true
	}

	// Bare /opt root is a system root, but subdirectories like /opt/my-app are valid project targets
	if checkPath == "/opt" || checkPath == "/opt/" {
		return true
	}

	return false
}

// IsSystemOrAppDir returns true if dirPath appears to be an operating system binary,
// program installation, or application directory (e.g. AppData\Local\Programs, Program Files,
// macOS .app bundles, /usr/bin). It does NOT match valid project locations like
// /opt/my-app, /var/www, /testbed, or /workspace.
// Vivechak research workspaces must not be initialized or resolved inside these directories.
func IsSystemOrAppDir(dirPath string) bool {
	if dirPath == "" {
		return false
	}
	// Always normalize backslashes to forward slashes across all platforms
	normalized := strings.ToLower(strings.ReplaceAll(dirPath, "\\", "/"))
	normalized = path.Clean(normalized)

	// Direct check
	if isSystemOrAppDirNormalized(normalized) {
		return true
	}

	// Strip Windows drive letter (e.g. "c:") so root-based paths match uniformly
	checkPath := normalized
	if len(checkPath) >= 2 && checkPath[1] == ':' && checkPath[0] >= 'a' && checkPath[0] <= 'z' {
		checkPath = checkPath[2:]
		if isSystemOrAppDirNormalized(checkPath) {
			return true
		}
	}

	// If a Windows path with drive letter was prepended by CWD on non-Windows platforms
	// (e.g. "/path/to/cwd/c:/program files/someapp"), check the Windows path portion:
	if idx := strings.Index(normalized, ":/"); idx >= 1 {
		sub := normalized[idx+1:]
		if isSystemOrAppDirNormalized(sub) {
			return true
		}
	}

	return false
}

var (
	lastResolvedWorkspaceMu sync.RWMutex
	lastResolvedWorkspace   string
)

// SetLastResolvedWorkspace records the most recently resolved workspace root in memory.
func SetLastResolvedWorkspace(ws string) {
	lastResolvedWorkspaceMu.Lock()
	defer lastResolvedWorkspaceMu.Unlock()
	lastResolvedWorkspace = ws
}

// GetLastResolvedWorkspace returns the in-memory cached workspace root, if any.
func GetLastResolvedWorkspace() string {
	lastResolvedWorkspaceMu.RLock()
	defer lastResolvedWorkspaceMu.RUnlock()
	return lastResolvedWorkspace
}

// ResolveWorkspace implements the 5-step workspace resolution chain:
//
//  1. Explicit project_root argument (from tool call)
//  2. Environment variables: VIVECHAK_PROJECT_ROOT, VIVECHAK_DEFAULT_ROOT, WORKSPACE, PROJECT_ROOT
//  3. Discover research/ directory walking up from CWD (skipping system/app dirs)
//  4. Fall back to in-process session memory (last resolved workspace with existing research/)
//  5. Error
//
// Returns the absolute path to the workspace root.
func ResolveWorkspace(explicit string) (string, error) {
	// Step 1: Explicit argument
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", fmt.Errorf("resolving explicit path %q: %w", explicit, err)
		}
		// If explicit path points directly to research directory, normalize to parent workspace root
		if filepath.Base(abs) == ResearchDir {
			if _, err := os.Stat(filepath.Join(abs, ResearchDir)); os.IsNotExist(err) {
				if fi, err := os.Stat(abs); err == nil && fi.IsDir() {
					abs = filepath.Dir(abs)
				}
			}
		}
		SetLastResolvedWorkspace(abs)
		return abs, nil
	}

	// Step 2: Environment variables (VIVECHAK_PROJECT_ROOT, VIVECHAK_DEFAULT_ROOT, WORKSPACE, PROJECT_ROOT)
	envRoot := os.Getenv("VIVECHAK_PROJECT_ROOT")
	if envRoot == "" {
		envRoot = os.Getenv("VIVECHAK_DEFAULT_ROOT")
	}
	if envRoot == "" {
		envRoot = os.Getenv("WORKSPACE")
	}
	if envRoot == "" {
		envRoot = os.Getenv("PROJECT_ROOT")
	}
	// Protect against unexpanded IDE macro variables like "${workspaceFolder}"
	if envRoot != "" && !strings.HasPrefix(envRoot, "${") {
		abs, err := filepath.Abs(envRoot)
		if err == nil {
			SetLastResolvedWorkspace(abs)
			return abs, nil
		}
	}

	// Step 3: Walk up from CWD looking for research/ directory (skipping system/app dirs)
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	dir := cwd
	for {
		if !IsSystemOrAppDir(dir) {
			researchPath := filepath.Join(dir, ResearchDir)
			if info, err := os.Stat(researchPath); err == nil && info.IsDir() {
				SetLastResolvedWorkspace(dir)
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding research/
			break
		}
		dir = parent
	}

	// Step 4: Fall back to in-process session memory
	if cached := GetLastResolvedWorkspace(); cached != "" {
		researchPath := filepath.Join(cached, ResearchDir)
		if info, err := os.Stat(researchPath); err == nil && info.IsDir() {
			return cached, nil
		}
	}

	// Step 5: Error
	return "", fmt.Errorf(
		"could not resolve workspace: no explicit path, VIVECHAK_PROJECT_ROOT not set, "+
			"research/ not found walking up from %s, and no prior workspace initialized in this session. "+
			"Run vivechak_init first or pass project_root explicitly", cwd)
}

// WorkspaceExists checks if a workspace has been initialized at the given path.
// Returns true if the research/ directory exists.
func WorkspaceExists(root string) bool {
	info, err := os.Stat(filepath.Join(root, ResearchDir))
	return err == nil && info.IsDir()
}

// WorkspaceInfo describes the current state of a Vivechak workspace.
type WorkspaceInfo struct {
	// Root is the absolute path to the workspace root.
	Root string `json:"root"`

	// Initialized reports whether research/ directory exists.
	Initialized bool `json:"initialized"`

	// HasPipeline reports whether RESEARCH-PIPELINE.md exists (project scope).
	HasPipeline bool `json:"has_pipeline,omitempty"`

	// HasDecisions reports whether DECISIONS.md exists.
	HasDecisions bool `json:"has_decisions,omitempty"`

	// SessionCount is the number of session files in research/sessions/.
	SessionCount int `json:"session_count"`

	// TemplateCount is the number of template files in research/templates/.
	TemplateCount int `json:"template_count"`

	// Scope is the detected scope (empty if not determinable).
	Scope Scope `json:"scope,omitempty"`
}

// InspectWorkspace scans a workspace and returns its current state.
func InspectWorkspace(root string) WorkspaceInfo {
	info := WorkspaceInfo{Root: root}

	researchPath := filepath.Join(root, ResearchDir)
	if fi, err := os.Stat(researchPath); err == nil && fi.IsDir() {
		info.Initialized = true
	} else {
		return info
	}

	// Check for pipeline
	if _, err := os.Stat(filepath.Join(root, PipelineFile)); err == nil {
		info.HasPipeline = true
	}

	// Try to detect scope from metadata file first
	hasMetadataScope := false
	if metaData, err := os.ReadFile(filepath.Join(root, MetadataFile)); err == nil {
		var meta WorkspaceMeta
		if json.Unmarshal(metaData, &meta) == nil && ValidScope(meta.Scope) {
			info.Scope = meta.Scope
			hasMetadataScope = true
		}
	}

	// Scan research directory for plan files to check info.HasPipeline and infer scope if needed
	researchEntries, _ := os.ReadDir(researchPath)
	var hasDecisionPlan, hasComparisonPlan bool
	for _, e := range researchEntries {
		name := e.Name()
		if !e.IsDir() && strings.HasSuffix(name, ".md") {
			if strings.HasSuffix(name, "-plan.md") {
				hasDecisionPlan = true
				info.HasPipeline = true
			} else if strings.HasSuffix(name, "-comparison.md") || strings.Contains(name, "-cmp-") {
				hasComparisonPlan = true
				info.HasPipeline = true
			}
		}
	}

	// If metadata did not provide a scope, infer scope from pipeline presence or manual workspace files
	if !hasMetadataScope {
		if hasDecisionPlan {
			info.Scope = ScopeDecision
		} else if hasComparisonPlan {
			info.Scope = ScopeComparison
		} else if info.HasPipeline {
			info.Scope = ScopeProject
		} else {
			// Check sessions directory for session ID prefixes if still undetermined
			if sEntries, err := os.ReadDir(filepath.Join(root, SessionsDir)); err == nil {
				for _, se := range sEntries {
					sName := se.Name()
					if strings.HasPrefix(sName, "S1-") || strings.HasPrefix(sName, "D-") {
						info.Scope = ScopeDecision
						break
					} else if strings.HasPrefix(sName, "C-") || strings.HasPrefix(sName, "C1-") || strings.HasPrefix(sName, "CMP-") || strings.HasPrefix(sName, "COMP-") {
						info.Scope = ScopeComparison
						break
					}
				}
			}
			if info.Scope == "" {
				for _, e := range researchEntries {
					name := e.Name()
					if !e.IsDir() && strings.HasSuffix(name, "-decision.md") {
						info.Scope = ScopeDecision
						break
					}
				}
			}
		}
	}

	// Check for decisions (non-empty DECISIONS.md or ADR files in research/)
	if fi, err := os.Stat(filepath.Join(root, DecisionsFile)); err == nil && fi.Size() > 0 {
		info.HasDecisions = true
	} else {
		for _, e := range researchEntries {
			name := e.Name()
			if !e.IsDir() && strings.HasSuffix(name, ".md") && !IsSpecialResearchFile(name) && strings.HasPrefix(strings.ToUpper(name), "D-") {
				info.HasDecisions = true
				break
			}
		}
	}

	// Count sessions (research sessions in sessions/ plus synthesis session FAD.md if present and not in sessions/)
	sessionsPath := filepath.Join(root, SessionsDir)
	hasSynthesisInSessions := false
	if entries, err := os.ReadDir(sessionsPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
				info.SessionCount++
				stem := strings.TrimSuffix(e.Name(), ".md")
				if IsSynthesisSession(stem) {
					hasSynthesisInSessions = true
				}
			}
		}
	}
	if !hasSynthesisInSessions {
		if _, err := os.Stat(filepath.Join(root, FADFile)); err == nil {
			info.SessionCount++
		}
	}

	// Count templates (only non-hidden .md files)
	templatesPath := filepath.Join(root, TemplatesDir)
	if entries, err := os.ReadDir(templatesPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && !strings.HasPrefix(e.Name(), ".") {
				info.TemplateCount++
			}
		}
	}

	return info
}

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

			hasExplicitFMID := fm != nil && (fm.GetString("session_id") != "" || fm.GetString("id") != "")
			id := ExtractSessionID(e.Name(), fm)
			if id != "" {
				completed[id] = true
				completed[strings.ToUpper(id)] = true
			}

			// If no explicit frontmatter session ID was present, also match session ID prefix
			// to avoid deadlocking the DAG when files are saved with title slugs (e.g. T1-01-database.md)
			if !hasExplicitFMID {
				stem := strings.TrimSuffix(e.Name(), ".md")
				lowerStem := strings.ToLower(stem)
				for _, kid := range knownIDs {
					lowerKid := strings.ToLower(kid)
					if lowerStem == lowerKid || strings.HasPrefix(lowerStem, lowerKid+"-") || strings.HasPrefix(lowerStem, lowerKid+"_") {
						completed[kid] = true
						completed[strings.ToUpper(kid)] = true
					}
				}
				if sub := sessionIDPrefixRe.FindStringSubmatch(stem); len(sub) > 1 {
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

	if len(sources) == 0 {
		if decData, err := ws.ReadFile(DecisionsFile); err == nil && len(decData) > 0 {
			sources = append(sources, decisionSource{file: DecisionsFile, data: decData})
		}
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
		if processedIDs[id] {
			return
		}
		processedIDs[id] = true

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
	unanchoredRe := regexp.MustCompile(`(?ms)^---\s*\n(.*?)\n---`)

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
			if matches := unanchoredRe.FindAllSubmatch(chunk, -1); len(matches) > 0 {
				for _, m := range matches {
					block := fmt.Sprintf("---\n%s\n---", string(m[1]))
					if fm, _, err := ParseFrontmatter([]byte(block)); err == nil && fm != nil {
						if fm.Has("id") || fm.Has("decision_id") || fm.Has("door_type") {
							checkDecisionFM(fm)
							parsedAny = true
						}
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
