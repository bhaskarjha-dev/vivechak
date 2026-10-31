package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
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

// ExtractDecisionOutcomeAndAlternatives extracts the Decision Outcome and Rejected Alternatives
// sections from an ADR for tamper-evident challenge hashing per P8.
func ExtractDecisionOutcomeAndAlternatives(adrData []byte) []byte {
	text := string(adrData)
	var b strings.Builder
	// Extract Decision Outcome section
	if loc := regexp.MustCompile(`(?mi)^#+\s+Decision(?:\s+Outcome)?\b`).FindStringIndex(text); loc != nil {
		rest := text[loc[0]:]
		nextLoc := regexp.MustCompile(`(?m)\n#+\s+`).FindStringIndex(rest[len("## Decision"):])
		if nextLoc != nil {
			b.WriteString(rest[:len("## Decision")+nextLoc[0]])
		} else {
			b.WriteString(rest)
		}
	}
	b.WriteString("\n")
	// Extract Rejected Alternatives section
	if loc := regexp.MustCompile(`(?mi)^#+\s+(?:Rejected\s+Alternatives|Alternatives(?:\s+Considered)?)\b`).FindStringIndex(text); loc != nil {
		rest := text[loc[0]:]
		nextLoc := regexp.MustCompile(`(?m)\n#+\s+`).FindStringIndex(rest[len("## Alternatives"):])
		if nextLoc != nil {
			b.WriteString(rest[:len("## Alternatives")+nextLoc[0]])
		} else {
			b.WriteString(rest)
		}
	}
	content := strings.TrimSpace(b.String())
	if content == "" {
		// Fallback to entire body if specific sections aren't separated
		return adrData
	}
	return []byte(content)
}

// ComputeDecisionOutcomeHash computes the SHA-256 hex digest of Decision Outcome and Rejected Alternatives.
func ComputeDecisionOutcomeHash(adrData []byte) string {
	extracted := ExtractDecisionOutcomeAndAlternatives(adrData)
	sum := sha256.Sum256(store.NormalizeLineEndings(extracted))
	return hex.EncodeToString(sum[:])
}

// WriteDecisionChallengeRecord writes the next D-###-R<n>.md challenge record into research/
func WriteDecisionChallengeRecord(ws *store.Workspace, decisionID string, decisionContent []byte, prompts []string) (string, string, error) {
	decUpper := strings.ToUpper(decisionID)
	nextNum := 1
	if entries, err := ws.ListDir(ResearchDir); err == nil {
		re := regexp.MustCompile(fmt.Sprintf(`^(?i)%s-R(\d+)\.md$`, regexp.QuoteMeta(decUpper)))
		for _, e := range entries {
			if m := re.FindStringSubmatch(e.Name()); len(m) > 1 {
				if num, nErr := strconv.Atoi(m[1]); nErr == nil && num >= nextNum {
					nextNum = num + 1
				}
			}
		}
	}
	recordID := fmt.Sprintf("%s-R%d", decUpper, nextNum)
	filename := fmt.Sprintf("%s.md", recordID)
	relPath := filepath.Join(ResearchDir, filename)

	hash := ComputeDecisionOutcomeHash(decisionContent)
	date := time.Now().UTC().Format("2006-01-02")

	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", recordID)
	b.WriteString("type: challenge\n")
	fmt.Fprintf(&b, "decision: %s\n", decUpper)
	fmt.Fprintf(&b, "decision_id: %s\n", decUpper)
	fmt.Fprintf(&b, "decision_hash: %s\n", hash)
	fmt.Fprintf(&b, "date: %s\n", date)
	b.WriteString("status: complete\n")
	b.WriteString("vivechak: 1\n")
	b.WriteString("---\n\n")

	fmt.Fprintf(&b, "# Adversarial Red-Team Challenge: %s (%s)\n\n", decUpper, recordID)
	b.WriteString("## Challenge Prompts\n\n")
	for i, p := range prompts {
		fmt.Fprintf(&b, "### Prompt %d\n\n%s\n\n", i+1, p)
	}

	if err := ws.WriteArtifact(relPath, []byte(b.String()), 0o644, true); err != nil {
		return "", "", fmt.Errorf("writing challenge record %s: %w", filename, err)
	}

	return recordID, hash, nil
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
