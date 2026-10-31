package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	toolCallRegex = regexp.MustCompile(`(?s)\*\*Tool Call:\*\*\s*` + "`" + `(\w+)` + "`" + `\s*` + "```" + `json\s*\n(.*?)\n` + "```" + `\s*\*\*Tool Response:\*\*\s*` + "```" + `json\s*\n(.*?)\n` + "```")
	pathRegex     = regexp.MustCompile(`research/[\w./-]+\.md`)
)

type docToolStep struct {
	toolName string
	args     map[string]any
	rawArgs  string
	rawResp  string
}

// TestDocsFlows_ReplaysExampleWalkthroughs verifies that every tool call documented in
// examples/*.md replays cleanly against the live compiled Vivechak binary over stdio MCP.
func TestDocsFlows_ReplaysExampleWalkthroughs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live docs flows replay test in short mode")
	}

	binPath := resolveOrBuildTestBinary(t)

	// Locate repo root
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}

	exampleFiles, err := filepath.Glob(filepath.Join(repoRoot, "examples", "*.md"))
	if err != nil || len(exampleFiles) == 0 {
		t.Fatalf("no example markdown files found in %s: %v", filepath.Join(repoRoot, "examples"), err)
	}

	for _, exFile := range exampleFiles {
		baseName := filepath.Base(exFile)
		t.Run(baseName, func(t *testing.T) {
			contentBytes, err := os.ReadFile(exFile)
			if err != nil {
				t.Fatalf("reading %s: %v", exFile, err)
			}
			content := string(contentBytes)

			matches := toolCallRegex.FindAllStringSubmatch(content, -1)
			if len(matches) == 0 {
				t.Skipf("no tool call pairs found in %s", baseName)
			}

			var steps []docToolStep
			for _, m := range matches {
				name := m[1]
				argsJSON := m[2]
				respJSON := m[3]

				var args map[string]any
				if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
					t.Fatalf("%s: invalid JSON in tool call arguments for %s: %v", baseName, name, err)
				}
				steps = append(steps, docToolStep{
					toolName: name,
					args:     args,
					rawArgs:  argsJSON,
					rawResp:  respJSON,
				})
			}

			// Run the sequence in a dedicated clean temporary project
			projDir := t.TempDir()

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			session, cleanup := startLiveMCPServer(t, ctx, binPath, projDir)
			defer cleanup()

			for i, step := range steps {
				stepArgs := make(map[string]any)
				for k, v := range step.args {
					stepArgs[k] = v
				}
				stepArgs["project_root"] = projDir

				res, err := session.CallTool(ctx, &mcp.CallToolParams{
					Name:      step.toolName,
					Arguments: stepArgs,
				})
				if err != nil {
					t.Fatalf("%s step %d (%s): CallTool transport error: %v", baseName, i+1, step.toolName, err)
				}
				if res.IsError {
					var errText string
					if len(res.Content) > 0 {
						if tc, ok := res.Content[0].(*mcp.TextContent); ok {
							errText = tc.Text
						}
					}
					t.Fatalf("%s step %d (%s): returned isError: true. Response: %s", baseName, i+1, step.toolName, errText)
				}

				// Verify documented paths appear in live response
				if len(res.Content) > 0 {
					if tc, ok := res.Content[0].(*mcp.TextContent); ok {
						liveText := tc.Text
						documentedPaths := pathRegex.FindAllString(step.rawResp, -1)
						for _, p := range documentedPaths {
							if !strings.Contains(liveText, p) {
								t.Errorf("%s step %d (%s): documented path %q not found in real response:\n%s",
									baseName, i+1, step.toolName, p, liveText)
							}
						}
					}
				}
			}
		})
	}
}

// TestDocsFlows_FailsOnMissingRequiredArg asserts that dropping a required parameter
// (such as kind for vivechak_save) is rejected by the server schema validation.
func TestDocsFlows_FailsOnMissingRequiredArg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live test in short mode")
	}

	binPath := resolveOrBuildTestBinary(t)
	projDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	session, cleanup := startLiveMCPServer(t, ctx, binPath, projDir)
	defer cleanup()

	// vivechak_save without required "kind" parameter must fail
	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "vivechak_save",
		Arguments: map[string]any{
			"project_root": projDir,
			"content":      "# Test",
		},
	})
	if err != nil {
		// Protocol-level schema rejection or tool error
		return
	}
	if !res.IsError {
		t.Fatalf("expected vivechak_save without required kind to fail, but got success")
	}
}
