package mcputil

import "strings"

// testPlan builds a plan in the plan format (HIGH-07): one ```yaml vivechak-plan``` block with the given plan
// field ("PROJECT" or a decision id) and session entries, then a "## Sessions" section made of sections (see
// testSection). preamble is prose placed between the title and the block.
func testPlan(title, plan, preamble string, entries []string, sections ...string) string {
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

// testSection is one session's "### <id>: <title>" heading with its prompt fence.
func testSection(id, title, prompt string) string {
	return "### " + id + ": " + title + "\n\n```prompt\n" + prompt + "\n```\n\n"
}

// threeSessionTestPlan is the project plan many tests use: T1-01 and T1-02 in parallel, then the synthesis
// session SYN-01 (kind: synthesis) after both.
func threeSessionTestPlan(t1, p1, t2, p2, tSyn, pSyn string) string {
	return testPlan("Research Pipeline", "PROJECT", "", []string{
		"{id: T1-01, title: " + t1 + ", depends: []}",
		"{id: T1-02, title: " + t2 + ", depends: []}",
		"{id: SYN-01, title: " + tSyn + ", kind: synthesis, depends: [T1-01, T1-02]}",
	}, testSection("T1-01", t1, p1), testSection("T1-02", t2, p2), testSection("SYN-01", tSyn, pSyn))
}

// oneSessionTestPlan is a project plan with one research session, T1-01, and its synthesis session SYN-01.
func oneSessionTestPlan(title, prompt string) string {
	return testPlan("Research Pipeline", "PROJECT", "", []string{
		"{id: T1-01, title: " + title + ", depends: []}",
		"{id: SYN-01, title: Synthesis, kind: synthesis, depends: [ALL]}",
	}, testSection("T1-01", title, prompt), testSection("SYN-01", "Synthesis", "Synthesize."))
}
