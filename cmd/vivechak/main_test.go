package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintUsage(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	out := buf.String()

	if !strings.Contains(out, "vivechak") {
		t.Errorf("expected usage to contain 'vivechak', got %q", out)
	}
	if !strings.Contains(out, "Commands:") {
		t.Errorf("expected usage to contain 'Commands:', got %q", out)
	}
	if !strings.Contains(out, "serve") || !strings.Contains(out, "mcp-config") || !strings.Contains(out, "doctor") {
		t.Errorf("missing commands in usage: %q", out)
	}
}

func TestVersionDefault(t *testing.T) {
	if !strings.Contains(version, "0.1.0") {
		t.Errorf("expected version to contain 0.1.0, got %q", version)
	}
}
