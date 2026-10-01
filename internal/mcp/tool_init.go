package mcputil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	vembed "github.com/bhaskarjha-dev/vivechak/internal/embed"
	store "github.com/bhaskarjha-dev/vivechak/internal/store"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// InitInput holds the arguments for vivechak_init.
type InitInput struct {
	ProjectRoot string `json:"project_root,omitempty" jsonschema:"workspace root path (optional; uses resolution chain if omitted)"`
	Scope       string `json:"scope,omitempty"         jsonschema:"research scope: project | decision | comparison (default: project)"`
}

func registerInit(server *sdkmcp.Server) {
	sdkmcp.AddTool(server,
		&sdkmcp.Tool{
			Name:  "vivechak_init",
			Title: "Initialize Vivechak Workspace",
			Description: "Create a Vivechak research workspace in the target directory. " +
				"Creates research/ directory structure and copies scope-appropriate templates. " +
				"Accepts scope parameter to adjust layout: 'project' (full pipeline, all 5 templates), " +
				"'decision' (single decision, 2 templates), or 'comparison' (bounded comparison, 1 template). " +
				"Safe to call on an already-initialized workspace — reports existing state without overwriting.",
			Annotations: &sdkmcp.ToolAnnotations{
				ReadOnlyHint:    false,
				IdempotentHint:  true,
				DestructiveHint: BoolPtr(false),
				OpenWorldHint:   BoolPtr(false),
			},
		},
		handleInit,
	)
}

func handleInit(_ context.Context, req *sdkmcp.CallToolRequest, in InitInput) (*sdkmcp.CallToolResult, Envelope, error) {
	const tool = "vivechak_init"

	// Resolve workspace
	root := in.ProjectRoot
	if root == "" {
		cwd, _ := os.Getwd()
		root = cwd
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("invalid path %q: %w", root, err),
			"Provide a valid absolute or relative path as project_root.")
	}

	// Guard against initializing at a filesystem root
	if absRoot == "/" || absRoot == "\\" || filepath.Dir(absRoot) == absRoot {
		return ErrorResult(tool, fmt.Errorf("project_root cannot be a filesystem root: %q", absRoot),
			"Provide a project directory path, not a filesystem root.")
	}

	// Guard against unguided initialization directly in user's root home directory (F-05, MED-06)
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if absHome, err := filepath.Abs(home); err == nil && filepath.Clean(absRoot) == filepath.Clean(absHome) {
			return ErrorResult(tool, fmt.Errorf("refusing to initialize workspace directly in user home directory %q", absRoot),
				"Provide a dedicated project subdirectory path as project_root, not the root home directory.")
		}
	}

	// Ensure the root directory exists — init is the only tool that creates it
	if err := os.MkdirAll(absRoot, 0o755); err != nil {
		return ErrorResult(tool, fmt.Errorf("creating directory %q: %w", absRoot, err),
			"Check filesystem permissions.")
	}

	ws, err := store.OpenWorkspace(absRoot)
	if err != nil {
		return ErrorResult(tool, fmt.Errorf("opening workspace %q: %w", absRoot, err),
			"Provide a valid absolute or relative path as project_root.")
	}
	defer ws.Close()


	// Parse scope
	scope := core.ScopeProject
	if in.Scope != "" {
		s, ok := core.ScopeFromString(in.Scope)
		if !ok {
			return ErrorResult(tool, fmt.Errorf("invalid scope %q — must be 'project', 'decision', or 'comparison'", in.Scope),
				"Call vivechak_init with scope set to 'project', 'decision', or 'comparison'.")
		}
		scope = s
	}

	// Check if already initialized
	if core.WorkspaceExists(absRoot) {
		var restored []string
		for _, tmpl := range core.TemplatesForScope(scope) {
			relPath := filepath.Join(core.TemplatesDir, tmpl)
			if !ws.FileExists(relPath) {
				if data, err := vembed.ReadTemplate(tmpl); err == nil {
					_ = ws.MkdirAll(core.TemplatesDir, 0o755)
					if err := store.WriteFileAtomic(ws.Root(), relPath, data, 0o644); err == nil {
						restored = append(restored, tmpl)
					}
				}
			}
		}
		info := core.InspectWorkspace(absRoot)
		msg := fmt.Sprintf("Workspace already initialized at %s (%d templates, %d sessions)", absRoot, info.TemplateCount, info.SessionCount)
		if len(restored) > 0 {
			msg = fmt.Sprintf("Workspace repaired at %s: restored %d missing templates (%s)", absRoot, len(restored), strings.Join(restored, ", "))
		}
		env := Envelope{
			Success: true,
			Message: msg,
			Data:    info,
			NextStep: fmt.Sprintf("Workspace exists. Run vivechak_prepare_generator with scope '%s' to get the generator prompt, "+
				"or vivechak_status to see current progress.", scope),
			Meta: NewMeta(tool),
		}
		return env.ToResult()
	}

	// Create directory structure
	dirs := []string{
		core.ResearchDir,
		core.SessionsDir,
		core.TemplatesDir,
	}
	for _, dir := range dirs {
		if err := ws.MkdirAll(dir, 0o755); err != nil {
			return ErrorResult(tool, fmt.Errorf("creating %s: %w", dir, err),
				"Check filesystem permissions and try again.")
		}
	}

	// Copy templates from embedded assets
	var copied []string
	var warnings []string
	for _, tmpl := range core.TemplatesForScope(scope) {
		data, err := vembed.ReadTemplate(tmpl)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("W-TEMPLATE-MISSING: %s not found in embedded assets", tmpl))
			continue
		}

		relPath := filepath.Join(core.TemplatesDir, tmpl)
		if err := store.WriteFileAtomic(ws.Root(), relPath, data, 0o644); err != nil {
			return ErrorResult(tool, fmt.Errorf("writing template %s: %w", tmpl, err),
				"Check filesystem permissions and try again.")
		}
		copied = append(copied, tmpl)
	}

	// Write workspace metadata (scope + creation timestamp)
	meta := core.WorkspaceMeta{
		Scope:     scope,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	metaBytes, _ := json.Marshal(meta)
	if err := store.WriteFileAtomic(ws.Root(), core.MetadataFile, metaBytes, 0o644); err != nil {
		warnings = append(warnings, fmt.Sprintf("W-META-WRITE: could not write workspace metadata: %v", err))
	}

	env := Envelope{
		Success: true,
		Message: fmt.Sprintf("Initialized %s workspace at %s (%d templates copied)", scope, absRoot, len(copied)),
		Data: map[string]any{
			"workspace_root":   absRoot,
			"scope":            scope,
			"templates_copied": copied,
			"directories_created": dirs,
		},
		Warnings: warnings,
		NextStep: fmt.Sprintf("Run vivechak_prepare_generator with your %s description to get the generator prompt. "+
			"Execute that prompt to generate your research pipeline.", scope),
		Meta: NewMeta(tool),
	}
	return env.ToResult()
}
