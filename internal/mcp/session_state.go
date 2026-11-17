package mcputil

import (
	"path/filepath"
	"sync"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// sessionState is what the server remembers for one MCP session (one chat): the workspace it last resolved
// (rule "session"), the client's roots once asked for, and the ids vivechak_start reserved, per workspace and id
// prefix. Calls without a session (unit tests calling a handler directly) have no state.
type sessionState struct {
	workspace core.WorkspaceMemory

	rootsMu      sync.Mutex
	rootsFetched bool
	roots        []string

	mu            sync.Mutex
	reserved      map[string]string // workspace root + "\x00" + prefix ("C" or "D") -> id reserved by vivechak_start
	planDecisions map[string]string // workspace root -> decision id of the plan this session saved last
}

// sessionStates maps each live *sdkmcp.ServerSession to its *sessionState; an entry is dropped when its
// session's connection closes.
var sessionStates sync.Map

// sessionOf returns the MCP session that sent req, or nil.
func sessionOf(req *sdkmcp.CallToolRequest) *sdkmcp.ServerSession {
	if req == nil {
		return nil
	}
	return req.Session
}

// stateForSession returns the state kept for ss, creating it on first use; nil when ss is nil.
func stateForSession(ss *sdkmcp.ServerSession) *sessionState {
	if ss == nil {
		return nil
	}
	if v, ok := sessionStates.Load(ss); ok {
		return v.(*sessionState)
	}
	v, loaded := sessionStates.LoadOrStore(ss, &sessionState{})
	if !loaded {
		go func() {
			_ = ss.Wait()
			sessionStates.Delete(ss)
		}()
	}
	return v.(*sessionState)
}

func reservationKey(root, prefix string) string {
	return filepath.Clean(root) + "\x00" + prefix
}

// reserveID records the id vivechak_start allocated for prefix in root; a later start replaces it. A new
// decision reservation also ends the session's current decision plan: the session has moved on to a new decision.
func (s *sessionState) reserveID(root, prefix, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.reserved == nil {
		s.reserved = make(map[string]string)
	}
	s.reserved[reservationKey(root, prefix)] = id
	if prefix == "D" {
		delete(s.planDecisions, filepath.Clean(root))
	}
}

// reservedID returns the id vivechak_start reserved for prefix in root in this session, if any.
func (s *sessionState) reservedID(root, prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reserved[reservationKey(root, prefix)]
}

// notePlanDecision records the decision whose plan this session just saved in root (its current decision).
func (s *sessionState) notePlanDecision(root, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.planDecisions == nil {
		s.planDecisions = make(map[string]string)
	}
	s.planDecisions[filepath.Clean(root)] = id
}

// planDecisionID returns the decision whose plan this session saved last in root, if any.
func (s *sessionState) planDecisionID(root string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.planDecisions[filepath.Clean(root)]
}

// forgetRoots drops the cached client roots, so the next tool call asks the client again
// (on notifications/roots/list_changed).
func (s *sessionState) forgetRoots() {
	s.rootsMu.Lock()
	defer s.rootsMu.Unlock()
	s.rootsFetched = false
	s.roots = nil
}
