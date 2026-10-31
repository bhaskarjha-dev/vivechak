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

	var options []string
	if fm != nil && fm.Has("options") {
		if rawOpts, ok := fm["options"].([]any); ok {
			for _, o := range rawOpts {
				if s, ok := o.(string); ok && strings.TrimSpace(s) != "" {
					options = append(options, strings.TrimSpace(s))
				}
			}
		} else if strList := fm.GetStringSlice("options"); len(strList) > 0 {
			options = strList
		}
	}
	if len(options) == 0 {
		options = extractComparisonOptions(content)
	}

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
	fm, _, _ := ParseFrontmatter(data)
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

	var options []string
	if fm != nil && fm.Has("options") {
		if rawOpts, ok := fm["options"].([]any); ok {
			for _, o := range rawOpts {
				if s, ok := o.(string); ok && strings.TrimSpace(s) != "" {
					options = append(options, strings.TrimSpace(s))
				}
			}
		} else if strList := fm.GetStringSlice("options"); len(strList) > 0 {
			options = strList
		}
	}
	if len(options) == 0 {
		options = extractComparisonOptions(content)
	}
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
	ignoreCols := map[string]bool{
		"criterion": true, "criteria": true, "weight": true,
		"rationale": true, "evidence": true, "notes": true,
		"source": true, "sources": true, "score": true, "scores": true,
	}

	lines := strings.Split(content, "\n")
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "|") && (strings.Contains(trimmed, "Criterion") || strings.Contains(trimmed, "Criteria")) {
			parts := strings.Split(trimmed, "|")
			for _, p := range parts {
				cand := strings.TrimSpace(p)
				lower := strings.ToLower(cand)
				if cand == "" || ignoreCols[lower] ||
					strings.Contains(lower, "evidence") ||
					strings.Contains(lower, "provenance") ||
					strings.Contains(lower, "grade") ||
					strings.Contains(lower, "rationale") ||
					strings.Contains(lower, "notes") ||
					strings.Contains(lower, "source") ||
					strings.Contains(lower, "justification") {
					continue
				}
				optName := cand
				for _, suffix := range []string{" Score (1-5)", " Score (1-10)", " Score", " (1-5)", " (1-10)"} {
					if strings.HasSuffix(strings.ToLower(optName), strings.ToLower(suffix)) {
						optName = strings.TrimSpace(optName[:len(optName)-len(suffix)])
						break
					}
				}
				lowerOpt := strings.ToLower(optName)
				if optName != "" && !ignoreCols[lowerOpt] &&
					!strings.Contains(lowerOpt, "evidence") &&
					!strings.Contains(lowerOpt, "provenance") &&
					!strings.Contains(lowerOpt, "grade") &&
					!strings.Contains(lowerOpt, "rationale") &&
					!strings.Contains(lowerOpt, "notes") &&
					!strings.Contains(lowerOpt, "source") &&
					!strings.Contains(lowerOpt, "justification") {
					options = append(options, optName)
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
