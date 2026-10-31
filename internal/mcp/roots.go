package mcputil

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// getClientRoots queries the MCP client for advertised root directories if supported.
func getClientRoots(ctx context.Context, req *sdkmcp.CallToolRequest) []string {
	if req == nil || req.Session == nil {
		return nil
	}
	res, err := req.Session.ListRoots(ctx, nil)
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
