package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

var jsonBlockRe = regexp.MustCompile("(?s)```json\\s*(\\{.*?\\})\\s*```")

func extractComparisonJSONContent(walkthrough []byte) string {
	matches := jsonBlockRe.FindAllSubmatch(walkthrough, -1)
	for _, m := range matches {
		var obj map[string]any
		if err := json.Unmarshal(m[1], &obj); err == nil {
			if kind, _ := obj["kind"].(string); kind == "comparison" {
				if val, ok := obj["content"].(string); ok && val != "" {
					return val
				}
			}
		}
	}
	return ""
}

func extractDecisionJSONContent(walkthrough []byte) string {
	matches := jsonBlockRe.FindAllSubmatch(walkthrough, -1)
	for _, m := range matches {
		var obj map[string]any
		if err := json.Unmarshal(m[1], &obj); err == nil {
			if _, hasDecID := obj["decision_id"]; hasDecID {
				if val, ok := obj["content"].(string); ok && val != "" {
					return val
				}
			}
		}
	}
	return ""
}

// TestValidatorParityCorpus_WalkthroughExamples validates that the exact artifact contents
// documented in sample walkthroughs pass their respective validators with zero blocking issues.
func TestValidatorParityCorpus_WalkthroughExamples(t *testing.T) {
	examplesDir := filepath.Join("..", "..", "examples")

	// 1. Comparison Scope: SAMPLE-COMPARISON.md embedded comparison artifact
	compPath := filepath.Join(examplesDir, "SAMPLE-COMPARISON.md")
	compBytes, err := os.ReadFile(compPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", compPath, err)
	}
	compContent := extractComparisonJSONContent(compBytes)
	if compContent == "" {
		t.Fatalf("failed to extract comparison content from %s", compPath)
	}
	compRes := ValidateComparison([]byte(compContent))
	if compRes.HasBlocking() {
		t.Errorf("SAMPLE-COMPARISON.md embedded artifact failed ValidateComparison:")
		for _, issue := range compRes.Issues {
			if issue.Level <= L2Block {
				t.Errorf("  [%s] %s: %s", issue.Level, issue.Code, issue.Message)
			}
		}
	}

	// 2. Decision Scope: SAMPLE-DECISION.md embedded ADR artifact
	decPath := filepath.Join(examplesDir, "SAMPLE-DECISION.md")
	decBytes, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", decPath, err)
	}
	decContent := extractDecisionJSONContent(decBytes)
	if decContent == "" {
		t.Fatalf("failed to extract decision content from %s", decPath)
	}
	decRes := ValidateDecision([]byte(decContent))
	if decRes.HasBlocking() {
		t.Errorf("SAMPLE-DECISION.md embedded artifact failed ValidateDecision:")
		for _, issue := range decRes.Issues {
			if issue.Level <= L2Block {
				t.Errorf("  [%s] %s: %s", issue.Level, issue.Code, issue.Message)
			}
		}
	}
}

// TestValidatorParityCorpus_DecisionsRegistry validates that decisions in the canonical
// DECISIONS.md registry parse and validate without blocking issues.
func TestValidatorParityCorpus_DecisionsRegistry(t *testing.T) {
	metaDecisionsPath := filepath.Join("..", "..", "docs", "meta-research", "DECISIONS.md")
	data, err := os.ReadFile(metaDecisionsPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", metaDecisionsPath, err)
	}

	blocks := ExtractUnanchoredDecisionBlocks(data)
	if len(blocks) == 0 {
		t.Fatalf("expected to extract decision blocks from %s, got 0", metaDecisionsPath)
	}

	for _, block := range blocks {
		fm, _, err := ParseFrontmatter(block)
		if err != nil || fm == nil {
			continue
		}
		id := fm.GetString("id")
		if id == "" {
			continue
		}

		res := ValidateDecision(block)
		if res.HasBlocking() {
			t.Errorf("decision %s in DECISIONS.md failed ValidateDecision:", id)
			for _, issue := range res.Issues {
				if issue.Level <= L2Block {
					t.Errorf("  [%s] %s: %s", issue.Level, issue.Code, issue.Message)
				}
			}
		}
	}
}

// TestValidatorParityCorpus_TemplatePlaceholders verifies that all canonical templates in templates/
// trigger placeholder detection when unpopulated, proving that the agent-proof scan guards all template contracts.
func TestValidatorParityCorpus_TemplatePlaceholders(t *testing.T) {
	templatesDir := filepath.Join("..", "..", "templates")
	templateFiles, err := filepath.Glob(filepath.Join(templatesDir, "*.template.md"))
	if err != nil || len(templateFiles) == 0 {
		t.Fatalf("failed to find templates in %s (found %d): %v", templatesDir, len(templateFiles), err)
	}

	for _, tplPath := range templateFiles {
		tplBytes, err := os.ReadFile(tplPath)
		if err != nil {
			t.Fatalf("failed to read %s: %v", tplPath, err)
		}

		placeholders := ScanPlaceholders(tplBytes)
		if len(placeholders) == 0 {
			t.Errorf("expected template %s to contain detectable placeholder patterns, but found 0", filepath.Base(tplPath))
		}
	}
}
