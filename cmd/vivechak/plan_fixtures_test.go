package main

import "strings"

// livePlan builds a plan in the plan format (HIGH-07): one ```yaml vivechak-plan``` block with the given plan
// field ("PROJECT" or a decision id) and session entries, then a "## Sessions" section made of sections (see
// liveSection). preamble is prose placed between the title and the block.
func livePlan(title, plan, preamble string, entries []string, sections ...string) string {
	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	if preamble != "" {
		b.WriteString(preamble + "\n\n")
	}
	b.WriteString("```yaml vivechak-plan\nplan: " + plan + "\nsessions:\n")
	for _, e := range entries {
		b.WriteString("  - " + e + "\n")
	}
	b.WriteString("```\n\n## Sessions\n\n")
	for _, s := range sections {
		b.WriteString(s)
	}
	return b.String()
}

// liveSection is one session's "### <id>: <title>" heading with its prompt fence.
func liveSection(id, title, prompt string) string {
	return "### " + id + ": " + title + "\n\n```prompt\n" + prompt + "\n```\n\n"
}
