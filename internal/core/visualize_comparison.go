package core

import (
	"fmt"
	"strings"
)

// RenderComparisonMermaid generates a Mermaid diagram visualizing options in a comparison artifact.
func RenderComparisonMermaid(data []byte) string {
	fm, body, _ := ParseFrontmatter(data)
	title := "Comparison"
	if fm != nil && fm.GetString("title") != "" {
		title = fm.GetString("title")
	}

	content := string(body)
	if len(content) == 0 {
		content = string(data)
	}

	rec := ExtractSection(content, "recommendation", "verdict")
	rec = strings.TrimSpace(rec)
	rec = TruncateRunes(rec, 60)
	rec = strings.ReplaceAll(rec, "\"", "'")
	rec = strings.ReplaceAll(rec, "\n", " ")

	options := extractComparisonOptions(content)

	var sb strings.Builder
	sb.WriteString("graph TD\n")
	fmt.Fprintf(&sb, "    subgraph C[\"%s\"]\n", strings.ReplaceAll(title, "\"", "'"))
	if rec != "" {
		fmt.Fprintf(&sb, "        REC[\"Recommendation: %s\"]\n", rec)
	}

	for i, opt := range options {
		optID := fmt.Sprintf("OPT_%d", i+1)
		cleanOpt := strings.ReplaceAll(opt, "\"", "'")
		fmt.Fprintf(&sb, "        %s[\"%s\"]\n", optID, cleanOpt)
		if rec != "" {
			fmt.Fprintf(&sb, "        %s --> REC\n", optID)
		}
		if i == 0 {
			fmt.Fprintf(&sb, "        style %s fill:#22c55e,stroke:#16a34a,color:#ffffff\n", optID)
		} else {
			fmt.Fprintf(&sb, "        style %s fill:#3b82f6,stroke:#2563eb,color:#ffffff\n", optID)
		}
	}
	if rec != "" {
		sb.WriteString("        style REC fill:#059669,stroke:#047857,color:#ffffff\n")
	}
	sb.WriteString("    end\n")
	return sb.String()
}

// RenderComparisonTable generates a markdown status table from comparison data.
func RenderComparisonTable(data []byte) string {
	content := string(data)
	lines := strings.Split(content, "\n")
	var tableLines []string
	inTable := false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "|") {
			inTable = true
			tableLines = append(tableLines, l)
		} else if inTable {
			break
		}
	}
	if len(tableLines) >= 3 {
		return strings.Join(tableLines, "\n")
	}

	options := extractComparisonOptions(content)
	if len(options) == 0 {
		return "| Option | Status |\n|---|---|\n| - | No options identified |\n"
	}
	var sb strings.Builder
	sb.WriteString("| # | Option | Status |\n|---|---|---|\n")
	for i, opt := range options {
		status := "Alternative"
		if i == 0 {
			status = "Top Contender"
		}
		fmt.Fprintf(&sb, "| %d | %s | %s |\n", i+1, opt, status)
	}
	return sb.String()
}

func extractComparisonOptions(content string) []string {
	var options []string
	lines := strings.Split(content, "\n")
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "|") && (strings.Contains(trimmed, "Criterion") || strings.Contains(trimmed, "Criteria")) {
			parts := strings.Split(trimmed, "|")
			for _, p := range parts {
				cand := strings.TrimSpace(p)
				if cand != "" && !strings.EqualFold(cand, "Criterion") && !strings.EqualFold(cand, "Criteria") && !strings.EqualFold(cand, "Weight") && !strings.EqualFold(cand, "Rationale") {
					options = append(options, cand)
				}
			}
			if len(options) > 0 {
				return options
			}
		}
	}

	inAlts := false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.Contains(strings.ToLower(trimmed), "alternatives considered") {
			inAlts = true
			continue
		}
		if inAlts && strings.HasPrefix(trimmed, "|") && !strings.Contains(trimmed, "---|") && !strings.Contains(trimmed, "Verdict") {
			parts := strings.Split(trimmed, "|")
			if len(parts) >= 2 {
				cand := strings.TrimSpace(parts[1])
				if cand != "" && !strings.HasPrefix(cand, "[") {
					options = append(options, cand)
				}
			}
		} else if inAlts && strings.HasPrefix(trimmed, "## ") {
			break
		}
	}
	return options
}
