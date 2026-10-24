package core

import (
	"path/filepath"
	"testing"

	"github.com/bhaskarjha-dev/vivechak/internal/store"
)

func TestChallengeRecord_PersistenceAndVerification(t *testing.T) {
	tmpDir := t.TempDir()
	ws, err := store.OpenWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("OpenWorkspace: %v", err)
	}
	defer ws.Close()

	targetContent := []byte("# S-01: Research Session\nFindings on database performance.")
	expectedHash := ComputeContentHash(targetContent)

	if HasValidChallengeRecord(ws, "S-01") {
		t.Errorf("expected no valid challenge record before save")
	}

	rec := ChallengeRecord{
		TargetID:   "S-01",
		TargetHash: expectedHash,
		Mode:       "red_team",
		Prompts:    []string{"failure mode 1"},
		Findings:   "Premortem completed",
	}

	relPath, err := SaveChallengeRecord(ws, rec)
	if err != nil {
		t.Fatalf("SaveChallengeRecord failed: %v", err)
	}

	if relPath != filepath.Join(ChallengesDir, "S-01-challenge.json") {
		t.Errorf("expected path %s, got %s", filepath.Join(ChallengesDir, "S-01-challenge.json"), relPath)
	}

	if !HasValidChallengeRecord(ws, "S-01") {
		t.Errorf("expected valid challenge record after save")
	}
}

func TestCheckHumanOnlyField_Rejection(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "human_reviewed true in frontmatter",
			content: "---\nid: D-001\nhuman_reviewed: true\n---\n# Title",
			wantErr: true,
		},
		{
			name:    "human_reviewed quoted true",
			content: "---\nid: D-001\nhuman_reviewed: \"true\"\n---\n# Title",
			wantErr: true,
		},
		{
			name:    "human_reviewed false allowed",
			content: "---\nid: D-001\nhuman_reviewed: false\n---\n# Title",
			wantErr: false,
		},
		{
			name:    "human_reviewed null allowed",
			content: "---\nid: D-001\nhuman_reviewed: null\n---\n# Title",
			wantErr: false,
		},
		{
			name:    "no human_reviewed field",
			content: "---\nid: D-001\n---\n# Title",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckHumanOnlyField([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckHumanOnlyField() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanPlaceholders_Detection(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantLen int
	}{
		{
			name:    "contains bracketed todo",
			content: "# Heading\n[TODO: fill in this section]",
			wantLen: 1,
		},
		{
			name:    "contains condition or date placeholder",
			content: "review_trigger: \"[Condition or date for mandatory re-evaluation]\"",
			wantLen: 1,
		},
		{
			name:    "contains HTML TODO comment",
			content: "# Heading\n<!-- TODO: finish this later -->",
			wantLen: 1,
		},
		{
			name:    "clean research content",
			content: "# Heading\nProduction deployment milestone or scheduled quarterly review.",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := ScanPlaceholders([]byte(tt.content))
			if len(matches) != tt.wantLen {
				t.Errorf("ScanPlaceholders() found %d matches (%v), want %d", len(matches), matches, tt.wantLen)
			}
		})
	}
}
