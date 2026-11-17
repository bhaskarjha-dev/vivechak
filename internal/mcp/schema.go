package mcputil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ParamSpec describes one canonical input parameter on a Vivechak MCP tool.
type ParamSpec struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "string", "boolean", "integer", "array"
	ItemType    string   `json:"item_type,omitempty"`
	Description string   `json:"description"`
	Doc         string   `json:"doc,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Required    bool     `json:"required"`
}

// DocText returns Doc when set, falling back to Description.
func (p ParamSpec) DocText() string {
	if p.Doc != "" {
		return p.Doc
	}
	return p.Description
}

// ToolSpec describes a canonical Vivechak MCP tool and its input schema.
type ToolSpec struct {
	Name        string
	Description string
	Annotations *sdkmcp.ToolAnnotations
	Params      []ParamSpec
}

// ToInputSchema builds the wire JSON Schema object for tools/list.
func (ts ToolSpec) ToInputSchema() map[string]any {
	props := make(map[string]any, len(ts.Params))
	req := make([]string, 0)
	for _, p := range ts.Params {
		prop := map[string]any{
			"type": p.Type,
		}
		if p.Description != "" {
			prop["description"] = p.Description
		}
		if len(p.Enum) > 0 {
			prop["enum"] = p.Enum
		}
		if p.Type == "array" && p.ItemType != "" {
			prop["items"] = map[string]any{"type": p.ItemType}
		}
		props[p.Name] = prop
		if p.Required {
			req = append(req, p.Name)
		}
	}
	return map[string]any{
		"type":                 "object",
		"required":             req,
		"additionalProperties": true,
		"properties":           props,
	}
}

var (
	registeredSpecsMu sync.RWMutex
	registeredSpecs   []ToolSpec
)

func recordToolSpec(spec ToolSpec) {
	registeredSpecsMu.Lock()
	defer registeredSpecsMu.Unlock()
	for i, existing := range registeredSpecs {
		if existing.Name == spec.Name {
			registeredSpecs[i] = spec
			return
		}
	}
	registeredSpecs = append(registeredSpecs, spec)
}

// RegisteredToolSpecs returns the canonical tool definitions in registration order.
func RegisteredToolSpecs() []ToolSpec {
	registeredSpecsMu.RLock()
	n := len(registeredSpecs)
	registeredSpecsMu.RUnlock()
	if n == 0 {
		_ = NewServer("0.0.0", nil)
	}
	registeredSpecsMu.RLock()
	defer registeredSpecsMu.RUnlock()
	out := make([]ToolSpec, len(registeredSpecs))
	copy(out, registeredSpecs)
	return out
}

func validateAndExtractArgs(spec ToolSpec, raw json.RawMessage) (knownJSON []byte, unknownKeys []string, fix string, valErr error) {
	trimmed := bytes.TrimSpace(raw)
	var m map[string]json.RawMessage
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		m = map[string]json.RawMessage{}
	} else {
		dec := json.NewDecoder(bytes.NewReader(trimmed))
		dec.UseNumber()
		if err := dec.Decode(&m); err != nil || m == nil {
			return nil, nil, "Pass a JSON object matching the tool's inputSchema.",
				fmt.Errorf("E-INVALID-ARGUMENT: arguments for %s must be a JSON object", spec.Name)
		}
	}

	paramByName := make(map[string]ParamSpec, len(spec.Params))
	for _, p := range spec.Params {
		paramByName[p.Name] = p
		val, present := m[p.Name]
		if !present || bytes.Equal(bytes.TrimSpace(val), []byte("null")) {
			if p.Required {
				return nil, nil,
					fmt.Sprintf("Provide required argument %q (%s).", p.Name, p.DocText()),
					fmt.Errorf("E-INVALID-ARGUMENT: missing required argument %q", p.Name)
			}
			continue
		}
		if err := validateParamValue(p, val); err != nil {
			fixMsg := fmt.Sprintf("Pass a valid %s for %q.", p.Type, p.Name)
			if len(p.Enum) > 0 {
				fixMsg = fmt.Sprintf("Use one of [%s] for %q.", strings.Join(p.Enum, ", "), p.Name)
			}
			return nil, nil, fixMsg, err
		}
	}

	knownMap := make(map[string]json.RawMessage, len(m))
	for k, v := range m {
		if _, ok := paramByName[k]; ok {
			knownMap[k] = v
		} else {
			unknownKeys = append(unknownKeys, k)
		}
	}
	sort.Strings(unknownKeys)

	encoded, err := json.Marshal(knownMap)
	if err != nil {
		return nil, nil, "Pass valid JSON arguments.", fmt.Errorf("E-INVALID-ARGUMENT: encoding arguments: %w", err)
	}
	return encoded, unknownKeys, "", nil
}

func validateParamValue(p ParamSpec, raw json.RawMessage) error {
	switch p.Type {
	case "string":
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return fmt.Errorf("E-INVALID-ARGUMENT: argument %q must be a string", p.Name)
		}
		if p.Required && strings.TrimSpace(s) == "" {
			return fmt.Errorf("E-INVALID-ARGUMENT: required argument %q cannot be empty", p.Name)
		}
		if len(p.Enum) > 0 && s != "" && !slices.Contains(p.Enum, s) {
			return fmt.Errorf("E-INVALID-ARGUMENT: invalid %s %q (allowed: %s)", p.Name, s, strings.Join(p.Enum, ", "))
		}
	case "boolean":
		var b bool
		if err := json.Unmarshal(raw, &b); err != nil {
			return fmt.Errorf("E-INVALID-ARGUMENT: argument %q must be a boolean", p.Name)
		}
	case "integer":
		var n int
		if err := json.Unmarshal(raw, &n); err != nil {
			return fmt.Errorf("E-INVALID-ARGUMENT: argument %q must be an integer", p.Name)
		}
	case "array":
		var arr []json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil {
			return fmt.Errorf("E-INVALID-ARGUMENT: argument %q must be an array", p.Name)
		}
		if p.ItemType == "string" {
			for _, item := range arr {
				var s string
				if err := json.Unmarshal(item, &s); err != nil {
					return fmt.Errorf("E-INVALID-ARGUMENT: argument %q items must be strings", p.Name)
				}
			}
		}
	}
	return nil
}

// registerTool registers a typed handler using ToolSpec without emitting an outputSchema.
func registerTool[In any](
	server *sdkmcp.Server,
	spec ToolSpec,
	handler func(context.Context, *sdkmcp.CallToolRequest, In) (*sdkmcp.CallToolResult, Envelope, error),
) {
	recordToolSpec(spec)
	tool := &sdkmcp.Tool{
		Name:        spec.Name,
		Description: spec.Description,
		InputSchema: spec.ToInputSchema(),
		Annotations: spec.Annotations,
	}
	server.AddTool(tool, func(ctx context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
		var rawArgs json.RawMessage
		if req != nil && req.Params != nil {
			rawArgs = req.Params.Arguments
		}
		knownJSON, unknownKeys, fix, valErr := validateAndExtractArgs(spec, rawArgs)
		if valErr != nil {
			res, _, _ := ErrorResult(spec.Name, valErr, fix)
			return res, nil
		}
		var in In
		if len(knownJSON) > 0 {
			if err := json.Unmarshal(knownJSON, &in); err != nil {
				res, _, _ := ErrorResult(spec.Name, fmt.Errorf("E-INVALID-ARGUMENT: %w", err), "Pass valid arguments matching inputSchema.")
				return res, nil
			}
		}
		res, env, err := handler(ctx, req, in)
		if err != nil {
			errRes, _, _ := ErrorResult(spec.Name, err, "")
			return errRes, nil
		}
		if len(unknownKeys) > 0 && res != nil && !res.IsError {
			var knownNames []string
			for _, p := range spec.Params {
				knownNames = append(knownNames, p.Name)
			}
			w := NewWarning(
				"W-UNKNOWN-ARG",
				fmt.Sprintf("ignored unknown argument(s): %s", strings.Join(unknownKeys, ", ")),
				fmt.Sprintf("Use only %s parameters: %s.", spec.Name, strings.Join(knownNames, ", ")),
			)
			env.Warnings = append([]Warning{w}, env.Warnings...)
			res, _, err = env.ToResult()
			if err != nil {
				errRes, _, _ := ErrorResult(spec.Name, err, "")
				return errRes, nil
			}
		}
		return res, nil
	})
}
