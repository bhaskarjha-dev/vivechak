package mcputil

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestResourceListing(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	res, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources: %v", err)
	}

	expectedResources := []string{
		"vivechak://templates/decisions",
		"vivechak://templates/comparison",
		"vivechak://templates/fad",
		"vivechak://templates/conflict-resolution",
		"vivechak://templates/gate",
		"vivechak://templates/session",
		"vivechak://methodology/grades",
		"vivechak://methodology/principles",
	}

	if len(res.Resources) != len(expectedResources) {
		t.Fatalf("expected %d resources, got %d", len(expectedResources), len(res.Resources))
	}

	foundMap := make(map[string]*mcp.Resource)
	for _, r := range res.Resources {
		foundMap[r.URI] = r
		if r.MIMEType != "text/markdown" {
			t.Errorf("resource %s: expected mime type 'text/markdown', got %q", r.URI, r.MIMEType)
		}
		if r.Title == "" {
			t.Errorf("resource %s: missing title", r.URI)
		}
		if r.Description == "" {
			t.Errorf("resource %s: missing description", r.URI)
		}
	}

	for _, expectedURI := range expectedResources {
		if _, ok := foundMap[expectedURI]; !ok {
			t.Errorf("expected resource URI not found: %s", expectedURI)
		}
	}
}

func TestResourceRead_AllCanonicalResources(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	testCases := []struct {
		uri          string
		expectedText string
	}{
		{
			uri:          "vivechak://templates/decisions",
			expectedText: "schema_version",
		},
		{
			uri:          "vivechak://templates/comparison",
			expectedText: "WEP",
		},
		{
			uri:          "vivechak://templates/fad",
			expectedText: "Founding Architecture Document",
		},
		{
			uri:          "vivechak://templates/conflict-resolution",
			expectedText: "Analysis of Competing Hypotheses",
		},
		{
			uri:          "vivechak://templates/gate",
			expectedText: "Phase 0 Exit Gate",
		},
		{
			uri:          "vivechak://templates/session",
			expectedText: "Evidence Ledger",
		},
		{
			uri:          "vivechak://methodology/grades",
			expectedText: "Primary Empirical",
		},
		{
			uri:          "vivechak://methodology/principles",
			expectedText: "Context Architecture Law",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.uri, func(t *testing.T) {
			res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{
				URI: tc.uri,
			})
			if err != nil {
				t.Fatalf("ReadResource(%s): %v", tc.uri, err)
			}
			if len(res.Contents) != 1 {
				t.Fatalf("expected 1 content block, got %d", len(res.Contents))
			}
			content := res.Contents[0]
			if content.URI != tc.uri {
				t.Errorf("expected URI %q, got %q", tc.uri, content.URI)
			}
			if content.MIMEType != "text/markdown" {
				t.Errorf("expected MIMEType 'text/markdown', got %q", content.MIMEType)
			}
			if !strings.Contains(content.Text, tc.expectedText) {
				t.Errorf("expected content to contain %q, but got:\n%s", tc.expectedText, content.Text[:min(len(content.Text), 200)])
			}
		})
	}
}

func TestResourceRead_NotFound(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	invalidURIs := []string{
		"vivechak://nonexistent/resource",
		"vivechak://templates/nonexistent",
		"http://vivechak.dev/templates/decisions",
		"file:///templates/decisions",
		"",
	}

	for _, uri := range invalidURIs {
		_, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{
			URI: uri,
		})
		if err == nil {
			t.Errorf("expected error reading invalid URI %q, got nil", uri)
		}
	}
}

func TestResourceRead_PathTraversalAndMalformed(t *testing.T) {
	cs := testServer(t)
	ctx := context.Background()

	traversalURIs := []string{
		"vivechak://templates/../../etc/passwd",
		"vivechak://templates/../DECISIONS.template.md",
		"vivechak://templates/%2e%2e%2f",
		"vivechak://templates/decisions/extra",
		"vivechak://TEMPLATES/decisions",
		"vivechak://templates/decisions?param=1",
		"vivechak://templates/decisions#section",
		"vivechak://invalid_scheme",
		"file:///etc/passwd",
	}

	for _, uri := range traversalURIs {
		_, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{
			URI: uri,
		})
		if err == nil {
			t.Errorf("expected error reading traversal/malformed URI %q, got nil", uri)
		}
	}
}

