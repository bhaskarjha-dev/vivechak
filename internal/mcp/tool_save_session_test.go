package mcputil

import (
	"strings"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

func TestBuildNextStep_Synthesis_ExitGate(t *testing.T) {
	nextStep := buildNextStep("SYN-01", 0, nil, nil)
	if !strings.Contains(nextStep, "vivechak_run_gate") {
		t.Errorf("expected synthesis nextStep to mandate vivechak_run_gate, got: %s", nextStep)
	}

	nextStepFAD := buildNextStep("FAD", 0, nil, nil)
	if !strings.Contains(nextStepFAD, "vivechak_run_gate") {
		t.Errorf("expected FAD nextStep to mandate vivechak_run_gate, got: %s", nextStepFAD)
	}
}

func TestBuildNextStep_ADRAutoDraftGuidance(t *testing.T) {
	dag := &core.DAG{
		Sessions: []core.Session{
			{ID: "T1-01", DoorType: "one-way"},
			{ID: "T1-02", DoorType: "two-way"},
		},
	}
	owdStep := buildNextStep("T1-01", 0, nil, dag)
	if !strings.Contains(owdStep, `auto_draft_from="T1-01"`) {
		t.Errorf("expected OWD nextStep to include auto_draft_from guidance, got: %s", owdStep)
	}

	twdStep := buildNextStep("T1-02", 0, nil, dag)
	if !strings.Contains(twdStep, `auto_draft_from="T1-02"`) {
		t.Errorf("expected TWD nextStep to include auto_draft_from guidance, got: %s", twdStep)
	}
}
