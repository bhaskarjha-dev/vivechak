package mcputil

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxContentBytes = 10 * 1024 * 1024 // 10 MB

func validateContentSize(content string) error {
	if len(content) > maxContentBytes {
		return fmt.Errorf("E-INVALID-ARGUMENT: content exceeds maximum size of 10 MB (%d bytes)", len(content))
	}
	return nil
}

// isValidID rejects path traversal, slashes, leading dots, unexpanded placeholders, or illegal filename characters.
func isValidID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, "/\\") || strings.HasPrefix(id, ".") {
		return false
	}
	upper := strings.ToUpper(id)
	if strings.Contains(upper, "NNN") || strings.Contains(upper, "XXX") || strings.Contains(upper, "TBD") || upper == "NEW" || strings.HasSuffix(upper, "-NEW") {
		return false
	}
	for _, r := range id {
		if !isSafeIDRune(r) {
			return false
		}
	}
	return true
}

func isSafeIDRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.'
}

// Warning is a structured non-fatal diagnostic in the canonical Vivechak envelope.
type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Fix     string `json:"fix"`
}

// NextAction specifies the structured recommended next step for an agent.
type NextAction struct {
	Tool string         `json:"tool"`
	Args map[string]any `json:"args"`
	Why  string         `json:"why"`
}

// Envelope is the canonical response every Vivechak MCP tool returns.
type Envelope struct {
	Ok        bool        `json:"ok"`
	Code      string      `json:"code,omitempty"`
	Message   string      `json:"message"`
	Fix       string      `json:"fix,omitempty"`
	Data      any         `json:"data,omitempty"`
	Warnings  []Warning   `json:"warnings,omitempty"`
	Next      *NextAction `json:"next,omitempty"`
	Truncated bool        `json:"truncated,omitempty"`
}

// NewWarning constructs a Warning backed by the canonical code registry for default message/fix text.
func NewWarning(code, message, fix string) Warning {
	entry, ok := core.LookupCode(code)
	if message == "" && ok {
		message = entry.Message
	}
	if fix == "" && ok {
		fix = entry.Fix
	}
	if fix == "" {
		fix = "Review the warning details and adjust if needed."
	}
	return Warning{
		Code:    code,
		Message: message,
		Fix:     fix,
	}
}

// WarningFromIssue converts a core.ValidationIssue into a structured Warning.
func WarningFromIssue(issue core.ValidationIssue) Warning {
	msg := issue.Message
	if issue.Field != "" && !strings.Contains(msg, issue.Field) {
		msg = fmt.Sprintf("%s: %s", issue.Field, msg)
	}
	if issue.Prefix != "" && !strings.HasPrefix(msg, issue.Prefix) {
		msg = issue.Prefix + msg
	}
	return NewWarning(issue.Code, msg, issue.FixHint)
}

var (
	reFixSuffix   = regexp.MustCompile(`\s*\(fix:\s*([^)]*)\)\s*$`)
	reLevelPrefix = regexp.MustCompile(`^(?:[A-Za-z0-9_-]*\[[^\]]+\]\s*)+`)
)

// ParseWarning converts an internal diagnostic string into a structured Warning with a registered code.
func ParseWarning(raw string) Warning {
	trimmed := strings.TrimSpace(raw)
	fix := ""
	if m := reFixSuffix.FindStringSubmatch(trimmed); len(m) > 1 {
		fix = strings.TrimSpace(m[1])
		trimmed = strings.TrimSpace(trimmed[:len(trimmed)-len(m[0])])
	}
	trimmed = reLevelPrefix.ReplaceAllString(trimmed, "")
	code := core.ExtractCode(trimmed)
	if code == "" {
		switch {
		case strings.HasPrefix(trimmed, "LAYER-TRANSITION:"):
			code = "W-LAYER-TRANSITION"
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "LAYER-TRANSITION:"))
		case strings.HasPrefix(trimmed, "PRE-SYNTHESIS ADVISORY:"):
			code = "W-PRE-SYNTHESIS-GAP"
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "PRE-SYNTHESIS ADVISORY:"))
		default:
			code = "W-GATE-RULE-ISSUE"
		}
	} else if after, found := strings.CutPrefix(trimmed, code+":"); found {
		trimmed = strings.TrimSpace(after)
	}
	return NewWarning(code, trimmed, fix)
}

// ParseWarnings converts a slice of internal warning strings into structured Warnings.
func ParseWarnings(raw []string) []Warning {
	if len(raw) == 0 {
		return nil
	}
	out := make([]Warning, 0, len(raw))
	for _, r := range raw {
		if strings.TrimSpace(r) == "" {
			continue
		}
		out = append(out, ParseWarning(r))
	}
	return out
}

// ToResult serializes an Envelope as compact JSON in both TextContent and StructuredContent.
func (env Envelope) ToResult() (*mcp.CallToolResult, Envelope, error) {
	if env.Next != nil && env.Next.Args == nil {
		env.Next.Args = map[string]any{}
	}
	jsonBytes, err := json.Marshal(env)
	if err != nil {
		return nil, env, fmt.Errorf("marshaling envelope: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
		StructuredContent: json.RawMessage(jsonBytes),
	}, env, nil
}

func classifyErrorCode(err error) string {
	if err == nil {
		return "E-INTERNAL"
	}
	msg := err.Error()
	if c := core.ExtractCode(msg); c != "" {
		return c
	}
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "workspace not initialized"):
		return "E-NOT-INITIALIZED"
	case os.IsNotExist(err) || strings.Contains(lower, "not found") || strings.Contains(lower, "does not exist"):
		return "E-NOT-FOUND"
	case strings.Contains(lower, "lock"):
		return "E-LOCK-TIMEOUT"
	case strings.Contains(lower, "cannot remove completed") || strings.Contains(lower, "completed session"):
		return "E-SESSION-COMPLETED"
	case strings.Contains(lower, "invalid identifier") || strings.Contains(lower, "invalid session id") || strings.Contains(lower, "invalid supersedes id") || strings.Contains(lower, "invalid decision_id") || strings.Contains(lower, "invalid auto_draft_from"):
		return "E-ID-INVALID"
	case strings.Contains(lower, "opening workspace") || strings.Contains(lower, "creating ") || strings.Contains(lower, "writing ") || strings.Contains(lower, "reading ") || strings.Contains(lower, "allocating ") || strings.Contains(lower, "saving mutated"):
		return "E-IO"
	default:
		return "E-INVALID-ARGUMENT"
	}
}

// ErrorEnvelope constructs a canonical error Envelope with a registered code and actionable fix.
func ErrorEnvelope(_ string, err error, fix string) Envelope {
	code := classifyErrorCode(err)
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	if !strings.Contains(msg, code) {
		if msg == "" {
			msg = code
		} else {
			msg = code + ": " + msg
		}
	}
	if strings.TrimSpace(fix) == "" {
		if entry, ok := core.LookupCode(code); ok {
			fix = entry.Fix
		} else {
			fix = "Check arguments and workspace state, then retry."
		}
	}
	return Envelope{
		Ok:      false,
		Code:    code,
		Message: msg,
		Fix:     fix,
	}
}

// ErrorEnvelopeWithNext constructs a canonical error Envelope carrying an optional structured NextAction.
func ErrorEnvelopeWithNext(tool string, err error, next *NextAction) Envelope {
	fix := ""
	if next != nil {
		fix = next.Why
		if next.Args == nil {
			next.Args = map[string]any{}
		}
	}
	env := ErrorEnvelope(tool, err, fix)
	if next != nil && next.Tool != "" {
		env.Next = next
	}
	return env
}

// ErrorResult packages an error Envelope as an MCP tool error (IsError=true).
func ErrorResult(tool string, err error, fix string) (*mcp.CallToolResult, Envelope, error) {
	env := ErrorEnvelope(tool, err, fix)
	jsonBytes, _ := json.Marshal(env)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
		StructuredContent: json.RawMessage(jsonBytes),
		IsError:           true,
	}, env, nil
}

// ErrorResultWithNext packages an error Envelope with a NextAction as an MCP tool error (IsError=true).
func ErrorResultWithNext(tool string, err error, next *NextAction) (*mcp.CallToolResult, Envelope, error) {
	env := ErrorEnvelopeWithNext(tool, err, next)
	jsonBytes, _ := json.Marshal(env)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
		StructuredContent: json.RawMessage(jsonBytes),
		IsError:           true,
	}, env, nil
}

// BoolPtr returns a pointer to b for optional ToolAnnotation fields.
func BoolPtr(b bool) *bool { return &b }
