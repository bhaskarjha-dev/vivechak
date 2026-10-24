package version

import "testing"

func TestVersionDefaults(t *testing.T) {
	if Version == "" {
		t.Error("expected non-empty Version")
	}
	if Commit == "" {
		t.Error("expected non-empty Commit")
	}
	if Date == "" {
		t.Error("expected non-empty Date")
	}
}
