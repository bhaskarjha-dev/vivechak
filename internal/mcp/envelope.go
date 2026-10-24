// Package mcputil provides the standardized response envelope and helpers
// shared across all Vivechak MCP tool handlers.
//
// Every tool returns a response in the Envelope format per FINAL-PLAN.md:
//
//	{
//	  "success": true,
//	  "message": "Session R-01 saved as draft (2 warnings, 0 errors)",
//	  "data": { ... },
//	  "warnings": ["W-LOW-OPTIONS"],
//	  "next_step": "Record decisions using vivechak_record_decision, then run vivechak_next_session."
//	}
package mcputil

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var validIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.\-]*$`)

const MaxContentBytes = 10 * 1024 * 1024 // 10 MB

func validateContentSize(content string) error {
	if len(content) > MaxContentBytes {
		return fmt.Errorf("content exceeds maximum size of %d bytes (got %d)", MaxContentBytes, len(content))
	}
	return nil
}

func isValidID(id string) bool {
	return len(id) <= 128 && validIDPattern.MatchString(id)
}

// NextAction specifies the structured recommended next step for an agent (AX-04).
type NextAction struct {
	Tool string         `json:"tool"`
	Args map[string]any `json:"args,omitempty"`
	Why  string         `json:"why"`
}

// Envelope is the standardized response every Vivechak tool returns.
// It carries both ok/failure state and Guided Worker process guidance.
type Envelope struct {
	// Ok indicates whether the operation completed without errors (v0.1.0 standard).
	Ok bool `json:"ok"`

	// Success indicates whether the operation completed without errors (backward compat).
	Success bool `json:"success"`

	// Message is a human-readable summary of what happened.
	Message string `json:"message"`

	// Data carries the tool-specific payload (varies by tool).
	Data any `json:"data,omitempty"`

	// Warnings lists non-fatal issues found during the operation.
	// Each warning uses a W-PREFIX code (e.g., "W-LOW-OPTIONS", "W-OUTPUT-SIZE").
	Warnings []string `json:"warnings,omitempty"`

	// Next provides structured next-action guidance for agents (tool, args, why).
	Next *NextAction `json:"next,omitempty"`

	// NextStep tells the agent what to do next (Guided Worker pattern, free text).
	NextStep string `json:"next_step"`

	// Meta carries non-functional metadata about the response.
	Meta *EnvMeta `json:"meta,omitempty"`
}

// EnvMeta carries metadata about the response itself.
type EnvMeta struct {
	// API is the envelope schema version (currently 1).
	API int `json:"api"`

	// Tool is the name of the tool that produced this response.
	Tool string `json:"tool"`

	// Truncated indicates the response was shortened due to size limits.
	// When true, the agent should re-call with verbose=true for full output.
	Truncated bool `json:"truncated,omitempty"`
}

// NewMeta creates an EnvMeta for the given tool name.
func NewMeta(tool string) *EnvMeta {
	return &EnvMeta{API: 1, Tool: tool}
}

// ToResult packages an Envelope into a CallToolResult with BOTH
// structuredContent AND text content (dual-channel delivery).
//
// Why dual-channel: Cursor drops structuredContent-only responses;
// Gemini CLI rejects missing structuredContent. We serve both.
func (env Envelope) ToResult() (*mcp.CallToolResult, Envelope, error) {
	// Synchronize Ok and Success
	if env.Ok && !env.Success {
		env.Success = true
	} else if env.Success && !env.Ok {
		env.Ok = true
	}
	// Synchronize Next and NextStep
	if env.Next != nil && env.NextStep == "" {
		env.NextStep = env.Next.Why
	} else if env.NextStep != "" && env.Next == nil {
		env.Next = &NextAction{Why: env.NextStep}
	}

	jsonBytes, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return nil, env, fmt.Errorf("marshaling envelope: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
		// StructuredContent is auto-populated from the Envelope (Out value)
		// by the SDK's AddTool generic handler.
	}, env, nil
}

// ErrorEnvelope creates a failure Envelope with the given error and guidance.
func ErrorEnvelope(tool string, err error, nextStep string) Envelope {
	return Envelope{
		Ok:       false,
		Success:  false,
		Message:  err.Error(),
		NextStep: nextStep,
		Next: &NextAction{
			Why: nextStep,
		},
		Meta: NewMeta(tool),
	}
}

// ErrorEnvelopeWithNext creates a failure Envelope with structured NextAction.
func ErrorEnvelopeWithNext(tool string, err error, next *NextAction) Envelope {
	nextStep := ""
	if next != nil {
		nextStep = next.Why
	}
	return Envelope{
		Ok:       false,
		Success:  false,
		Message:  err.Error(),
		NextStep: nextStep,
		Next:     next,
		Meta:     NewMeta(tool),
	}
}

// ErrorResult creates a CallToolResult for a tool-level error.
// Per MCP spec, tool errors are returned as Content with IsError=true,
// NOT as JSON-RPC protocol errors.
func ErrorResult(tool string, err error, nextStep string) (*mcp.CallToolResult, Envelope, error) {
	env := ErrorEnvelope(tool, err, nextStep)
	jsonBytes, _ := json.MarshalIndent(env, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
		IsError: true,
	}, env, nil
}

// ErrorResultWithNext creates a CallToolResult for a tool-level error with structured NextAction.
func ErrorResultWithNext(tool string, err error, next *NextAction) (*mcp.CallToolResult, Envelope, error) {
	env := ErrorEnvelopeWithNext(tool, err, next)
	jsonBytes, _ := json.MarshalIndent(env, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
		IsError: true,
	}, env, nil
}

// BoolPtr returns a pointer to a bool value.
// Used for optional ToolAnnotation fields (DestructiveHint, OpenWorldHint).
func BoolPtr(b bool) *bool { return &b }
