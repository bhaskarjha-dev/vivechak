package core

import (
	"strings"
	"testing"
)

func TestExtractSection_ChildHeadingsPreserved(t *testing.T) {
	doc := `## Recommendation
We recommend adopting LibSQL for edge persistence.
### Performance Rationale
LibSQL benchmarked at sub-5ms read latency.
#### Micro-benchmarks
Detailed micro-benchmarks show 3x throughput improvement over standard SQLite.
## Rejected Alternatives
PostgreSQL was rejected due to memory footprint.`

	rec := ExtractSection(doc, "recommend")
	if !strings.Contains(rec, "We recommend adopting LibSQL") {
		t.Errorf("missing primary recommendation text")
	}
	if !strings.Contains(rec, "### Performance Rationale") || !strings.Contains(rec, "LibSQL benchmarked at sub-5ms") {
		t.Errorf("child heading ### was erroneously dropped or truncated")
	}
	if !strings.Contains(rec, "#### Micro-benchmarks") {
		t.Errorf("child heading #### was erroneously dropped")
	}
	if strings.Contains(rec, "PostgreSQL was rejected") {
		t.Errorf("sibling heading ## Rejected Alternatives was not treated as boundary")
	}
}

func TestExtractSection_SiblingBoundary(t *testing.T) {
	doc := `# Title
## Background
Context information here.
## Alternatives
- Option A: Redis
- Option B: Memcached
## Verdict
Select Redis.`

	alt := ExtractSection(doc, "alternative")
	if !strings.Contains(alt, "Option A: Redis") {
		t.Errorf("expected alternative content")
	}
	if strings.Contains(alt, "Select Redis") {
		t.Errorf("sibling heading ## Verdict should have stopped extraction")
	}
}

func TestExtractSection_CaseInsensitiveKeyword(t *testing.T) {
	doc := `### KEY FINDINGS
- Finding 1: Scalability is confirmed.
- Finding 2: Latency is verified.
### NEXT STEPS
Review findings.`

	findings := ExtractSection(doc, "key finding")
	if !strings.Contains(findings, "Finding 1: Scalability") {
		t.Errorf("expected findings to match case-insensitively")
	}
	if strings.Contains(findings, "Review findings") {
		t.Errorf("next heading should stop extraction")
	}
}

func TestExtractSection_NotFoundAndEmpty(t *testing.T) {
	if got := ExtractSection("", "recommend"); got != "" {
		t.Errorf("expected empty string for empty content, got %q", got)
	}
	if got := ExtractSection("## Hello\nWorld", "nonexistent"); got != "" {
		t.Errorf("expected empty string for missing section, got %q", got)
	}
}

func TestExtractSection_CodeFenceWithComments(t *testing.T) {
	doc := `## Discovered Concerns
Here are the primary risks:
` + "```bash\n# Check system requirements\ncat /etc/os-release\n# Verify limits\nulimit -n\n```" + `
These concerns are substantial and exceed 50 characters in total length.

## Next Section
Other content.
`
	got := ExtractSection(doc, "discovered concerns")
	if !strings.Contains(got, "cat /etc/os-release") {
		t.Fatalf("ExtractSection truncated content before code block, got:\n%s", got)
	}
	if !strings.Contains(got, "These concerns are substantial") {
		t.Fatalf("ExtractSection truncated content at code comment, got:\n%s", got)
	}
	if strings.Contains(got, "Other content") {
		t.Fatalf("ExtractSection bled into Next Section, got:\n%s", got)
	}

	// Also test tilde fences (~~~)
	docTilde := `## Delta
` + "~~~python\n# Invalidate cache\nredis.flushall()\n~~~" + `
Delta continues here.
## Summary
Done.`
	gotTilde := ExtractSection(docTilde, "delta")
	if !strings.Contains(gotTilde, "redis.flushall") || !strings.Contains(gotTilde, "Delta continues here") {
		t.Fatalf("ExtractSection failed on tilde code fence, got:\n%s", gotTilde)
	}
	if strings.Contains(gotTilde, "Done.") {
		t.Fatalf("ExtractSection bled past tilde fence section, got:\n%s", gotTilde)
	}
}

