package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/store"
	"github.com/tailscale/hujson"
)

// ChallengesDir is the directory where challenge records are stored.
const ChallengesDir = "research/.challenges"

// ChallengeRecord documents an adversarial challenge executed against a session or decision.
type ChallengeRecord struct {
	TargetID   string   `json:"target_id"`
	TargetHash string   `json:"target_hash"` // SHA-256 hash of target content
	Timestamp  string   `json:"timestamp"`
	Mode       string   `json:"mode"`
	Prompts    []string `json:"prompts"`
	Findings   string   `json:"findings,omitempty"`
}

// ComputeContentHash computes the hex-encoded SHA-256 hash of normalized content.
func ComputeContentHash(content []byte) string {
	normalized := store.NormalizeLineEndings(content)
	h := sha256.Sum256(normalized)
	return hex.EncodeToString(h[:])
}

// SaveChallengeRecord persists a tamper-evident challenge record into research/.challenges/
func SaveChallengeRecord(ws *store.Workspace, record ChallengeRecord) (string, error) {
	if record.Timestamp == "" {
		record.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling challenge record: %w", err)
	}

	filename := fmt.Sprintf("%s-challenge.json", record.TargetID)
	relPath := filepath.Join(ChallengesDir, filename)

	if err := ws.WriteArtifact(relPath, data, 0o644, true); err != nil {
		return "", fmt.Errorf("writing challenge record: %w", err)
	}

	return relPath, nil
}

// HasValidChallengeRecord checks if a target has a valid recorded challenge in research/.challenges/
func HasValidChallengeRecord(ws WorkspaceFS, targetID string) bool {
	filename := fmt.Sprintf("%s-challenge.json", targetID)
	relPath := filepath.Join(ChallengesDir, filename)
	if data, err := ws.ReadFile(relPath); err == nil && len(data) > 0 {
		var rec ChallengeRecord
		if std, sErr := hujson.Standardize(data); sErr == nil {
			data = std
		}
		if err := json.Unmarshal(data, &rec); err == nil {
			if strings.EqualFold(rec.TargetID, targetID) && len(rec.Prompts) > 0 {
				return true
			}
		}
	}
	return false
}
