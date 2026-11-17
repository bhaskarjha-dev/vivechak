package core

import (
	"regexp"
	"strings"
)

// CodeEntry describes a canonical error, validation, warning, or coaching code emitted by Vivechak.
type CodeEntry struct {
	Code    string `json:"code"`
	Level   string `json:"level"` // "error", "validation", "warning", "coaching"
	Message string `json:"message"`
	Fix     string `json:"fix"`
}

var codeTokenRe = regexp.MustCompile(`\b([EVWQ]-[A-Z]{2,}(?:-[A-Z0-9]+)*)\b`)

var canonicalCodes = []CodeEntry{
	// Errors (E-*)
	{
		Code:    "E-CONFINEMENT",
		Level:   "error",
		Message: "Target artifact path escapes research/ or the workspace boundary.",
		Fix:     "Write artifacts only under research/ within the resolved project workspace.",
	},
	{
		Code:    "E-HUMAN-ONLY-FIELD",
		Level:   "error",
		Message: "Agent attempted to set a human-only sign-off field (human_reviewed, reviewed_by, reviewed_at, reviewed_hash).",
		Fix:     "Remove the sign-off fields and ask a human reviewer to run 'vck signoff <D-id> --reviewer \"<name>\"' in an interactive terminal.",
	},
	{
		Code:    "E-ID-CONFLICT",
		Level:   "error",
		Message: "Target file already exists on disk with a different artifact identifier.",
		Fix:     "Use a distinct id or slug for a new artifact, or match the existing file's id with revise=true.",
	},
	{
		Code:    "E-ID-INVALID",
		Level:   "error",
		Message: "Artifact or session identifier is malformed or does not match the artifact kind.",
		Fix:     "Use D-NNN for decisions, C-NNN for comparisons, or a valid session ID such as T1-01 (or D-NEW to allocate).",
	},
	{
		Code:    "E-INTERNAL",
		Level:   "error",
		Message: "Unexpected internal error while handling request.",
		Fix:     "Inspect server stderr diagnostics and retry the operation.",
	},
	{
		Code:    "E-INVALID-ARGUMENT",
		Level:   "error",
		Message: "Tool argument is missing, has the wrong type, or is outside its allowed values.",
		Fix:     "Pass arguments matching the tool's inputSchema in tools/list.",
	},
	{
		Code:    "E-IO",
		Level:   "error",
		Message: "Filesystem read, directory creation, or atomic write failed in the workspace.",
		Fix:     "Verify that the workspace directory exists and is writable, then retry.",
	},
	{
		Code:    "E-LOCK-TIMEOUT",
		Level:   "error",
		Message: "Timed out waiting to acquire the advisory file lock for an artifact.",
		Fix:     "Wait for concurrent writes to finish and retry the call.",
	},
	{
		Code:    "E-NOT-FOUND",
		Level:   "error",
		Message: "Requested plan, session, comparison, or decision file does not exist in research/.",
		Fix:     "Save the prerequisite artifact first or inspect available IDs with vivechak_status.",
	},
	{
		Code:    "E-NOT-INITIALIZED",
		Level:   "error",
		Message: "Workspace does not contain an initialized research/ directory yet.",
		Fix:     "Call vivechak_start with scope and question to initialize the workspace first.",
	},
	{
		Code:    "E-PLACEHOLDER-REMAINS",
		Level:   "error",
		Message: "Unexpanded template placeholders remain in an artifact that requires concrete content.",
		Fix:     "Replace all bracketed template placeholders with grounded findings before saving as complete or accepted.",
	},
	{
		Code:    "E-REVISE-REQUIRED",
		Level:   "error",
		Message: "Target artifact already exists with different content and revise=true was not set.",
		Fix:     "Pass revise=true to update that same artifact (archiving the old copy in research/.history/), or omit id for a new artifact.",
	},
	{
		Code:    "E-ROOT-AMBIGUOUS",
		Level:   "error",
		Message: "Client advertised multiple MCP workspace roots and no explicit project_root was given.",
		Fix:     "Pass project_root explicitly with the target project's absolute directory path.",
	},
	{
		Code:    "E-ROOT-INVALID",
		Level:   "error",
		Message: "Explicit or configured workspace root path does not exist or is not a directory.",
		Fix:     "Pass an existing project directory path in project_root.",
	},
	{
		Code:    "E-ROOT-SYSTEM",
		Level:   "error",
		Message: "Candidate directory is a system, home, or application directory rather than a project.",
		Fix:     "Pass project_root pointing to your project repository.",
	},
	{
		Code:    "E-ROOT-UNRESOLVED",
		Level:   "error",
		Message: "Could not resolve a safe project workspace root from arguments, client roots, environment, or cwd.",
		Fix:     "Pass project_root with the absolute path of the user's project directory.",
	},
	{
		Code:    "E-ROOT-UNSAFE",
		Level:   "error",
		Message: "Workspace path is empty or points to a restricted filesystem root, home, temp, or OS directory.",
		Fix:     "Choose a dedicated project directory inside your workspace.",
	},
	{
		Code:    "E-SESSION-COMPLETED",
		Level:   "error",
		Message: "Attempted to remove or alter the plan definition of an already completed research session.",
		Fix:     "Keep completed sessions intact and append post-hoc updates with vivechak_amend.",
	},
	{
		Code:    "E-TTY-REQUIRED",
		Level:   "error",
		Message: "vck signoff was invoked without an interactive terminal session.",
		Fix:     "Ask a human reviewer to run 'vck signoff <D-id> --reviewer \"<name>\"' directly in an interactive terminal.",
	},

	// Validation (V-*)
	{
		Code:    "V-CMP-NO-MATRIX",
		Level:   "validation",
		Message: "Comparison artifact lacks a Weighted Evaluation Protocol (WEP) Markdown table.",
		Fix:     "Include a criterion-by-option Markdown table with weights and 1-5 scores.",
	},
	{
		Code:    "V-CMP-OPTIONS",
		Level:   "validation",
		Message: "Comparison matrix evaluates fewer than two distinct options.",
		Fix:     "Compare between 2 and 5 named options as columns in the WEP matrix.",
	},
	{
		Code:    "V-CMP-SCORES",
		Level:   "validation",
		Message: "Comparison matrix cells lack valid numeric 1-5 option scores.",
		Fix:     "Score every option on every criterion from 1 to 5 backed by graded citations.",
	},
	{
		Code:    "V-CMP-WEIGHTS",
		Level:   "validation",
		Message: "Comparison criterion weights are missing or do not normalize to 1.0 (or 100%).",
		Fix:     "Assign positive criterion weights that sum to 1.0 (e.g. 0.40, 0.35, 0.25).",
	},
	{
		Code:    "V-EMPTY",
		Level:   "validation",
		Message: "Artifact content is empty.",
		Fix:     "Provide complete Markdown content matching the artifact's template.",
	},
	{
		Code:    "V-EMPTY-BODY",
		Level:   "validation",
		Message: "Artifact contains YAML frontmatter but no Markdown body.",
		Fix:     "Write the research findings, trade-offs, and evidence ledger below the frontmatter fence.",
	},
	{
		Code:    "V-INVALID-FRONTMATTER",
		Level:   "validation",
		Message: "YAML frontmatter block is syntactically malformed.",
		Fix:     "Fix YAML syntax errors between the opening and closing --- fences.",
	},
	{
		Code:    "V-MISSING-FIELD",
		Level:   "validation",
		Message: "Required YAML frontmatter field is missing.",
		Fix:     "Add the missing required key to the YAML frontmatter block.",
	},
	{
		Code:    "V-MISSING-FRONTMATTER",
		Level:   "validation",
		Message: "Artifact is missing its opening YAML frontmatter block.",
		Fix:     "Begin the document with a --- YAML frontmatter block containing required metadata.",
	},
	{
		Code:    "V-MISSING-SECTION",
		Level:   "validation",
		Message: "Required Markdown section heading is missing from the artifact.",
		Fix:     "Add all required Markdown sections for the scope template.",
	},
	{
		Code:    "V-MISSING-STATUS",
		Level:   "validation",
		Message: "Session frontmatter does not declare status.",
		Fix:     "Set status: complete (or draft) in the session's YAML frontmatter.",
	},
	{
		Code:    "V-NO-GRADED-CLAIM",
		Level:   "validation",
		Message: "Artifact contains no inline evidence grades (A-E).",
		Fix:     "Grade factual claims A-E with modifiers, verification labels, and source URLs.",
	},
	{
		Code:    "V-OWD-NO-CONCERNS",
		Level:   "validation",
		Message: "One-Way Door session lacks a substantive Discovered Concerns section.",
		Fix:     "Add a ## Discovered Concerns section (>=50 chars) detailing failure modes and disconfirming evidence.",
	},
	{
		Code:    "V-OWD-VAGUE-PROVENANCE",
		Level:   "validation",
		Message: "One-Way Door Grade A/B claim lacks a concrete URL or primary standard anchor.",
		Fix:     "Attach a canonical https:// URL to every Grade A/B claim informing a One-Way Door.",
	},
	{
		Code:    "V-PLACEHOLDER",
		Level:   "validation",
		Message: "Unexpanded template placeholder remains in the comparison output.",
		Fix:     "Replace all template placeholders with concrete option analysis and scores.",
	},
	{
		Code:    CodePlanCycle,
		Level:   "validation",
		Message: "Plan session dependency graph contains a cycle or depends on the synthesis session.",
		Fix:     "Remove circular depends entries so sessions form a DAG ending at kind: synthesis.",
	},
	{
		Code:    CodePlanDuplicatePrompt,
		Level:   "validation",
		Message: "Plan contains multiple ### <id>: prompt sections for the same session.",
		Fix:     "Keep exactly one ### <id>: <title> section with one prompt fence per session.",
	},
	{
		Code:    CodePlanInvalidBlock,
		Level:   "validation",
		Message: "Plan's yaml vivechak-plan block is invalid, duplicated, or uses unknown keys.",
		Fix:     "Keep one valid yaml vivechak-plan block using plan, sessions, and decisions keys.",
	},
	{
		Code:    CodePlanMissingBlock,
		Level:   "validation",
		Message: "Plan lacks the required fenced yaml vivechak-plan block.",
		Fix:     "Add one ```yaml vivechak-plan block declaring plan and sessions.",
	},
	{
		Code:    CodePlanMissingPrompt,
		Level:   "validation",
		Message: "Session declared in vivechak-plan has no prompt fence under ## Sessions.",
		Fix:     "Add ### <id>: <title> followed by one ```prompt fence under ## Sessions.",
	},
	{
		Code:    CodePlanNoSessions,
		Level:   "validation",
		Message: "Plan declares zero research sessions.",
		Fix:     "Declare at least one session in the vivechak-plan block's sessions list.",
	},
	{
		Code:    CodePlanSynthesis,
		Level:   "validation",
		Message: "Project pipeline does not have exactly one kind: synthesis session.",
		Fix:     "Set kind: synthesis on the single terminal synthesis session in the project plan.",
	},
	{
		Code:    CodePlanUnknownDependency,
		Level:   "validation",
		Message: "Session depends on a session ID not declared in the plan.",
		Fix:     "List only session IDs declared in the same vivechak-plan block (or [ALL] on synthesis).",
	},
	{
		Code:    CodePlanUnlistedSession,
		Level:   "validation",
		Message: "Prompt section under ## Sessions is not declared in the vivechak-plan block.",
		Fix:     "Add the session to the vivechak-plan sessions list or remove the extra section.",
	},
	{
		Code:    "V-TOO-SHORT",
		Level:   "validation",
		Message: "Artifact body is shorter than the minimum substantive length.",
		Fix:     "Expand the evidence-graded analysis and trade-off discussion before saving.",
	},

	// Warnings (W-*)
	{
		Code:    "W-CMP-GATE-NA",
		Level:   "warning",
		Message: "Comparison scope has no Phase 0 architecture gate requirements.",
		Fix:     "Proceed to implementation, or start decision scope if locking an ADR.",
	},
	{
		Code:    "W-DECISIONS-COMPILE",
		Level:   "warning",
		Message: "Individual ADR was saved, but recompiling research/DECISIONS.md failed.",
		Fix:     "Check write permissions on research/DECISIONS.md and re-run vivechak_record_decision.",
	},
	{
		Code:    "W-GATE-RULE-ISSUE",
		Level:   "warning",
		Message: "Phase 0 Gate surfaced a rule failure or structural check issue.",
		Fix:     "Inspect data.rules and resolve each failing check before coding.",
	},
	{
		Code:    "W-GATE-WRITE",
		Level:   "warning",
		Message: "Phase 0 Gate ran, but persisting research/PHASE-0-GATE.md failed.",
		Fix:     "Check write permissions under research/ and re-run vivechak_gate.",
	},
	{
		Code:    "W-ID-ALLOCATED",
		Level:   "warning",
		Message: "Omitted or placeholder identifier was resolved to a server-allocated ID.",
		Fix:     "Use the allocated ID returned in data.id on subsequent tool calls.",
	},
	{
		Code:    "W-INJECT-FAILED",
		Level:   "warning",
		Message: "Upstream session context could not be injected; serving raw session prompt.",
		Fix:     "Verify upstream session files under research/sessions/ are readable.",
	},
	{
		Code:    "W-INJECTION-SIZE",
		Level:   "warning",
		Message: "Injected upstream findings exceed the recommended context budget.",
		Fix:     "Keep upstream Key Findings and Recommendation sections concise.",
	},
	{
		Code:    "W-INVALID-DATE-FORMAT",
		Level:   "warning",
		Message: "Frontmatter date is not in YYYY-MM-DD format.",
		Fix:     "Format frontmatter date as ISO YYYY-MM-DD.",
	},
	{
		Code:    "W-LAYER-TRANSITION",
		Level:   "warning",
		Message: "All sessions in the prior DAG layer are complete; advancing to the next layer.",
		Fix:     "Check whether findings warrant vivechak_replan before running downstream sessions.",
	},
	{
		Code:    "W-MARKER-CONFLICT",
		Level:   "warning",
		Message: "ADR body contains nested HTML decision registry markers.",
		Fix:     "Remove manual <!-- DECISION: ... --> markers from individual ADR files.",
	},
	{
		Code:    "W-MISSING-DATE",
		Level:   "warning",
		Message: "Frontmatter omits the date field.",
		Fix:     "Add date: YYYY-MM-DD to the YAML frontmatter.",
	},
	{
		Code:    "W-MISSING-DOOR-TYPE",
		Level:   "warning",
		Message: "Artifact omits door_type (one-way or two-way).",
		Fix:     "Set door_type: one-way or door_type: two-way in YAML frontmatter.",
	},
	{
		Code:    "W-MISSING-FIELD",
		Level:   "warning",
		Message: "Recommended YAML frontmatter field is missing.",
		Fix:     "Add the recommended field to the YAML frontmatter.",
	},
	{
		Code:    "W-MISSING-FRONTMATTER",
		Level:   "warning",
		Message: "Artifact omits optional YAML frontmatter.",
		Fix:     "Add YAML frontmatter with id, title, date, and status.",
	},
	{
		Code:    "W-NO-CONCERNS",
		Level:   "warning",
		Message: "Two-Way Door session omits a Discovered Concerns section.",
		Fix:     "Add ## Discovered Concerns noting caveats or operational risks.",
	},
	{
		Code:    "W-NO-EVIDENCE-GRADES",
		Level:   "warning",
		Message: "Artifact body contains no inline evidence grade markers.",
		Fix:     "Tag factual claims with grades A-E (see vivechak://methodology/grades).",
	},
	{
		Code:    "W-OUTPUT-SIZE",
		Level:   "warning",
		Message: "Session prompt exceeded 10K estimated tokens and was truncated.",
		Fix:     "Call vivechak_next with verbose=true to fetch the full prompt.",
	},
	{
		Code:    "W-OWD-NO-CHALLENGE",
		Level:   "warning",
		Message: "One-Way Door ADR lacks a current adversarial red-team challenge record.",
		Fix:     "Run vivechak_review(mode=\"red_team\", id=\"<D-id>\") and cite the returned record_id.",
	},
	{
		Code:    "W-PLACEHOLDER-REMAINS",
		Level:   "warning",
		Message: "Proposed draft ADR still contains unexpanded template placeholders.",
		Fix:     "Replace every placeholder before marking status: accepted.",
	},
	{
		Code:    "W-PRE-SYNTHESIS-GAP",
		Level:   "warning",
		Message: "Completed One-Way Door session has evidentiary or challenge gaps before synthesis.",
		Fix:     "Amend weak One-Way Door sessions via vivechak_amend before running vivechak_gate.",
	},
	{
		Code:    "W-RECALLED-GRADE-CAP",
		Level:   "warning",
		Message: "Claim labelled recalled or parametric is graded higher than Grade D.",
		Fix:     "Cap model-recall claims at Grade D or verify with web search and cite the source URL.",
	},
	{
		Code:    "W-SHORT-BODY",
		Level:   "warning",
		Message: "Artifact Markdown body is unusually short.",
		Fix:     "Expand empirical findings, trade-offs, and citations if incomplete.",
	},
	{
		Code:    "W-SHORT-DECISION",
		Level:   "warning",
		Message: "Decision record body is unusually short.",
		Fix:     "Document context, decision outcome, rejected alternatives, and reversal triggers.",
	},
	{
		Code:    "W-STALE-DECISION-REFERENCE",
		Level:   "warning",
		Message: "Saved session references an ADR that has been superseded.",
		Fix:     "Review the session and append an update with vivechak_amend if needed.",
	},
	{
		Code:    "W-STALE-DOWNSTREAM",
		Level:   "warning",
		Message: "Completed downstream sessions depend on an artifact that was amended.",
		Fix:     "Review listed downstream sessions and amend them if conclusions changed.",
	},
	{
		Code:    "W-UNKNOWN-ARG",
		Level:   "warning",
		Message: "Tool call included unknown argument name(s), which were ignored.",
		Fix:     "Pass only parameter names declared in the tool's inputSchema.",
	},
	{
		Code:    "W-VAGUE-PROVENANCE",
		Level:   "warning",
		Message: "Grade A/B claim lacks an explicit URL or standard identifier.",
		Fix:     "Include a canonical https:// URL on the claim line or ledger row.",
	},

	// Quality coaching (Q-*)
	{
		Code:    "Q-BIAS",
		Level:   "coaching",
		Message: "Session Delta lists no contradicted or updated beliefs, suggesting possible confirmation bias.",
		Fix:     "Actively test disconfirming evidence and record how initial assumptions changed in ## Delta.",
	},
	{
		Code:    "Q-CONCERNS",
		Level:   "coaching",
		Message: "Session omits or under-develops the ## Discovered Concerns section.",
		Fix:     "Document concrete failure modes, operational caveats, and limitations in ## Discovered Concerns.",
	},
	{
		Code:    "Q-DEPTH",
		Level:   "coaching",
		Message: "Session contains fewer than the recommended 4-7 graded key findings.",
		Fix:     "Expand ## Key Findings with additional empirically graded claims where warranted.",
	},
	{
		Code:    "Q-EVIDENCE",
		Level:   "coaching",
		Message: "Session evidence count, URL anchoring, or grade mix warrants review.",
		Fix:     "Anchor Grade A claims with primary URLs and cite multiple independent sources.",
	},
	{
		Code:    "Q-FRESHNESS",
		Level:   "coaching",
		Message: "Session cites potentially stale year references (>2 years old).",
		Fix:     "Verify version-sensitive claims against current upstream documentation.",
	},
	{
		Code:    "Q-RIGOR",
		Level:   "coaching",
		Message: "Session recommends an option without documenting evaluated alternatives.",
		Fix:     "Add an ## Alternatives Considered section weighing rejected candidates and trade-offs.",
	},
	{
		Code:    "Q-STRUCTURE",
		Level:   "coaching",
		Message: "Session omits ## Prior or ## Delta belief-tracking sections.",
		Fix:     "Include both ## Prior and ## Delta to show how evidence updated initial assumptions.",
	},
}

var codeIndex = func() map[string]CodeEntry {
	m := make(map[string]CodeEntry, len(canonicalCodes))
	for _, c := range canonicalCodes {
		m[c.Code] = c
	}
	return m
}()

// AllCodes returns a copy of the canonical Vivechak diagnostic code registry in deterministic order.
func AllCodes() []CodeEntry {
	out := make([]CodeEntry, len(canonicalCodes))
	copy(out, canonicalCodes)
	return out
}

// LookupCode looks up a diagnostic code by exact identifier.
func LookupCode(code string) (CodeEntry, bool) {
	e, ok := codeIndex[strings.TrimSpace(code)]
	return e, ok
}

// ExtractCode extracts the first registered diagnostic code token from text, or "" if none is present.
func ExtractCode(text string) string {
	for _, m := range codeTokenRe.FindAllString(text, -1) {
		if _, ok := codeIndex[m]; ok {
			return m
		}
	}
	return ""
}
