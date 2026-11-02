package mcputil

import (
	"context"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// getClientRoots queries the MCP client for advertised root directories if supported.
func getClientRoots(ctx context.Context, req *sdkmcp.CallToolRequest) []string {
	if req == nil || req.Session == nil {
		return nil
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	//nolint:staticcheck // SA1019: ListRoots is functional during 12-month deprecation window (SEP-2577)
	res, err := req.Session.ListRoots(timeoutCtx, nil)
	if err != nil || res == nil {
		return nil
	}
	var uris []string
	for _, r := range res.Roots {
		if r != nil && r.URI != "" {
			uris = append(uris, r.URI)
		}
	}
	return uris
}
