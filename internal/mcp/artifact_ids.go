package mcputil

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	// placeholderIDRe matches the id placeholders agents copy from generators, templates and docs:
	// NEW, NNN, XXX or TBD, in any case, with or without the D-/C- prefix.
	placeholderIDRe = regexp.MustCompile(`(?i)^(?:([DC])-)?(?:NEW|NNN|XXX|TBD)$`)
	numberedIDRe    = regexp.MustCompile(`(?i)^([DC])-(\d{1,6})$`)
	// placeholderTokenRe finds placeholder decision ids inside plan or ADR content (D-NEW, d-tbd, D-NNN-S1, ...).
	placeholderTokenRe = regexp.MustCompile(`(?i)\bD-(?:NEW|NNN|XXX|TBD)\b`)
)

// idKinds names the artifact behind each id prefix, for error messages.
var idKinds = map[string]string{"D": "a decision id (D-NNN)", "C": "a comparison id (C-NNN)"}

// resolveArtifactID applies the one id rule shared by vivechak_save and vivechak_record_decision to a decision
// ("D") or comparison ("C") id. A missing id or a placeholder (NEW, NNN, XXX, TBD; any case; with or without the
// D-/C- prefix) gets the id this session's vivechak_start reserved for that scope while it is unused (reservedFree
// reports whether the artifact being saved already exists for it), else the next id no artifact uses; allocated
// reports that case (the caller warns W-ID-ALLOCATED). D-n/C-n are normalized to D-NNN/C-NNN. Anything else is
// E-ID-INVALID, before anything is written.
func resolveArtifactID(ss *sdkmcp.ServerSession, ws *store.Workspace, root, prefix, raw string, reservedFree func(id string) bool) (id string, allocated bool, err error) {
	raw = strings.TrimSpace(raw)
	if err := checkArtifactIDSyntax(prefix, raw); err != nil {
		return "", false, err
	}
	if m := numberedIDRe.FindStringSubmatch(raw); m != nil {
		n, _ := strconv.Atoi(m[2])
		return fmt.Sprintf("%s-%03d", prefix, n), false, nil
	}
	if st := stateForSession(ss); st != nil {
		// The decision this session is working on: the one whose plan it saved last, else the id its
		// vivechak_start reserved; comparisons only have the reservation.
		candidates := []string{st.reservedID(root, prefix)}
		if prefix == "D" {
			candidates = []string{st.planDecisionID(root), st.reservedID(root, "D")}
		}
		for _, id := range candidates {
			if id != "" && reservedFree(id) {
				return id, true, nil
			}
		}
	}
	for i := 0; i < 1000; i++ {
		id, err := store.AllocateID(ws, prefix)
		if err != nil {
			return "", false, fmt.Errorf("allocating %s id: %w", prefix, err)
		}
		if !artifactIDInUse(ws, id) {
			return id, true, nil
		}
	}
	return "", false, fmt.Errorf("allocating %s id: no free id found", prefix)
}

// checkArtifactIDSyntax is resolveArtifactID's id rule without allocating anything: "" (no id), D-n/C-n of the
// prefix's kind, or a placeholder; anything else is E-ID-INVALID.
func checkArtifactIDSyntax(prefix, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if m := numberedIDRe.FindStringSubmatch(raw); m != nil {
		if !strings.EqualFold(m[1], prefix) {
			return fmt.Errorf("E-ID-INVALID: %q is not %s", raw, idKinds[prefix])
		}
		return nil
	}
	m := placeholderIDRe.FindStringSubmatch(raw)
	if m == nil {
		return fmt.Errorf("E-ID-INVALID: %q is not D-NNN or C-NNN", raw)
	}
	if m[1] != "" && !strings.EqualFold(m[1], prefix) {
		return fmt.Errorf("E-ID-INVALID: %q is not %s", raw, idKinds[prefix])
	}
	return nil
}

// artifactIDInUse reports whether any research artifact is named after id (<id>-*, <id>_*, <id>.*).
func artifactIDInUse(ws *store.Workspace, id string) bool {
	upper := strings.ToUpper(id)
	for _, dir := range []string{core.ResearchDir, core.SessionsDir, core.DecisionsDir, core.ComparisonsDir} {
		entries, err := ws.ListDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := strings.ToUpper(e.Name())
			if strings.HasPrefix(name, upper+"-") || strings.HasPrefix(name, upper+"_") || strings.HasPrefix(name, upper+".") {
				return true
			}
		}
	}
	return false
}

// decisionPlanExists reports whether research/ holds a plan for decision id (<id>-plan.md, <id>-<slug>-plan.md, <id>.plan.md).
func decisionPlanExists(ws *store.Workspace, id string) bool {
	upper := strings.ToUpper(id)
	entries, err := ws.ListDir(core.ResearchDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		name := strings.ToUpper(e.Name())
		if name == upper+".PLAN.MD" || (strings.HasPrefix(name, upper+"-") && strings.HasSuffix(name, "-PLAN.MD")) {
			return true
		}
	}
	return false
}

// decisionADRExists reports whether research/ holds the ADR for decision id.
func decisionADRExists(ws *store.Workspace, id string) bool {
	_, err := ws.Stat(filepath.Join(core.ResearchDir, resolveDecisionFilename(ws, id, "")))
	return err == nil
}

// placeholderRefRe matches a reference whose decision part is a placeholder (NEW, NNN, XXX or TBD, any case, with
// or without the D- prefix): a decision ("D-NNN") or one of its sessions ("D-NNN-S1", "d-tbd-s2-landscape").
var placeholderRefRe = regexp.MustCompile(`(?i)^(?:D-)?(?:NEW|NNN|XXX|TBD)(-S\d+(?:[-_][A-Za-z0-9_-]*)?)?$`)

// placeholderRefFix is the next step when a placeholder reference cannot be resolved.
const placeholderRefFix = "Save the decision plan first (vivechak_save kind=\"plan\"; vivechak_start returns the decision id), " +
	"or use the session id vivechak_next serves (e.g. D-001-S1)."

// sessionDecisionID returns the decision this MCP session is working on in root, if its plan exists: the decision
// whose plan the session saved last, else the id its vivechak_start reserved.
func sessionDecisionID(ss *sdkmcp.ServerSession, ws *store.Workspace, root string) string {
	st := stateForSession(ss)
	if st == nil {
		return ""
	}
	for _, id := range []string{st.planDecisionID(root), st.reservedID(root, "D")} {
		if id != "" && decisionPlanExists(ws, id) {
			return id
		}
	}
	return ""
}

// resolvePlaceholderRef applies the id rule to the decision part of a reference (HIGH-15): a session id such as
// "D-NNN-S1" (and, when allowDecision, a bare decision placeholder such as "D-NNN") resolves to the decision this MCP
// session is working on, so the documented calls work as written; mapped reports the case (callers warn
// W-ID-ALLOCATED naming the real id). Any other reference is returned unchanged. An unresolvable placeholder, or a
// bare decision placeholder where a session id is required, is E-ID-INVALID; placeholderRefFix names the fix.
func resolvePlaceholderRef(ss *sdkmcp.ServerSession, ws *store.Workspace, root, raw string, allowDecision bool) (id string, mapped bool, err error) {
	raw = strings.TrimSpace(raw)
	m := placeholderRefRe.FindStringSubmatch(raw)
	if m == nil {
		return raw, false, nil
	}
	if m[1] == "" && !allowDecision {
		return "", false, fmt.Errorf("E-ID-INVALID: %q is a placeholder decision id, not a session id (sessions are <decision>-S<n>, e.g. D-001-S1)", raw)
	}
	did := sessionDecisionID(ss, ws, root)
	if did == "" {
		return "", false, fmt.Errorf("E-ID-INVALID: %q has a placeholder decision part, and this session has no decision plan to resolve it to", raw)
	}
	if m[1] == "" {
		return did, true, nil
	}
	return did + "-S" + m[1][2:], true, nil
}

// resolvePlaceholderRefList applies resolvePlaceholderRef to each entry of a comma-separated list of session ids.
func resolvePlaceholderRefList(ss *sdkmcp.ServerSession, ws *store.Workspace, root, list string) (string, []string, error) {
	if strings.TrimSpace(list) == "" {
		return list, nil, nil
	}
	parts := strings.Split(list, ",")
	var notes []string
	for i, p := range parts {
		id, mapped, err := resolvePlaceholderRef(ss, ws, root, p, false)
		if err != nil {
			return "", nil, err
		}
		if mapped {
			notes = append(notes, fmt.Sprintf("W-ID-ALLOCATED: session id %s resolved to %s", strings.TrimSpace(p), id))
			parts[i] = id
		}
	}
	return strings.Join(parts, ","), notes, nil
}
