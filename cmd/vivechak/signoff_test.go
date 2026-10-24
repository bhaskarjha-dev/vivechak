package main

import (
	"strings"
	"testing"
)

func TestSignoff_RequiresInteractiveTTY(t *testing.T) {
	// In standard go test execution, Stdin is not an interactive character device terminal.
	// runSignoff MUST strictly reject non-TTY execution with E-TTY-REQUIRED.
	err := runSignoff([]string{"D-001"})
	if err == nil {
		t.Fatalf("expected error from runSignoff in non-TTY test environment, got nil")
	}
	if !strings.Contains(err.Error(), "E-TTY-REQUIRED") {
		t.Errorf("expected error to contain E-TTY-REQUIRED, got: %v", err)
	}
}
