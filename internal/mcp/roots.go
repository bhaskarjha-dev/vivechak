package mcputil

import (
	"context"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// rootsTimeout bounds the one roots/list round trip per session.
const rootsTimeout = 200 * time.Millisecond

// getClientRoots returns the root URIs of the client that sent req. It asks the client (roots/list) only when
// the client's initialize request advertised the roots capability, and at most once per session: the answer
// (or the lack of one within rootsTimeout) is cached until the client sends notifications/roots/list_changed.
func getClientRoots(ctx context.Context, req *sdkmcp.CallToolRequest) []string {
	ss := sessionOf(req)
	if ss == nil || !clientAdvertisesRoots(ss) {
		return nil
	}
	st := stateForSession(ss)
	st.rootsMu.Lock()
	defer st.rootsMu.Unlock()
	if st.rootsFetched {
		return st.roots
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, rootsTimeout)
	defer cancel()
	//nolint:staticcheck // SA1019: ListRoots is functional during 12-month deprecation window (SEP-2577)
	res, err := ss.ListRoots(timeoutCtx, nil)
	st.rootsFetched = true
	st.roots = nil
	if err != nil || res == nil {
		return nil
	}
	for _, r := range res.Roots {
		if r != nil && r.URI != "" {
			st.roots = append(st.roots, r.URI)
		}
	}
	return st.roots
}

// clientAdvertisesRoots reports whether the client's initialize request declared the roots capability.
func clientAdvertisesRoots(ss *sdkmcp.ServerSession) bool {
	params := ss.InitializeParams()
	//nolint:staticcheck // SA1019: the roots capability is deprecated by SEP-2577 but still sent by clients
	return params != nil && params.Capabilities != nil && params.Capabilities.RootsV2 != nil
}

// handleRootsListChanged forgets the cached roots of the session that sent the notification.
func handleRootsListChanged(_ context.Context, req *sdkmcp.RootsListChangedRequest) {
	if req == nil {
		return
	}
	if st := stateForSession(req.Session); st != nil {
		st.forgetRoots()
	}
}
