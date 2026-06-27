---
name: rc-create-prd
description: Creates a Product Requirements Document through interactive brainstorming with parallel codebase and web research. Use when starting a new feature or product, building a PRD, or brainstorming requirements. Do not use for technical specifications, task breakdowns, or code implementation.
argument-hint: "[feature-name-or-idea] [idea-file]"
model: sonnet
effort: high
---

# Create PRD

Create a business-focused Product Requirements Document through structured brainstorming.

<HARD-GATE>
The brainstorming must precede the artifact, because a PRD written before the phases finish reflects your assumptions rather than the user's intent — and unexamined product assumptions are the most expensive kind to unwind later. This holds for every PRD regardless of perceived simplicity:
- Do not write the PRD file until all phases are complete and the user has approved the final draft — an unapproved file is a guess committed to disk.
- Do not skip the research phase: each PRD is enriched with codebase and market context so it builds on what exists instead of reinventing it.
- Do not skip user interactions: the user shapes the PRD at every decision point, because they hold the product intent you cannot derive from the code.
- Do not require section-by-section approval: generate the complete draft, then let the user review it — piecemeal sign-off adds delay without adding rigor.
</HARD-GATE>

## Code navigation (Serena)

If the Serena MCP is available, prefer its symbolic tools over whole-file reads during codebase research — they are LSP-accurate and token-efficient:

- `get_symbols_overview` to grasp a file's structure before reading it; `find_symbol` (by name path, e.g. `Type/method`) to jump straight to a definition.
- `find_referencing_symbols` to map every caller of a symbol before reasoning about impact.

Fall back to Grep/Glob + Read when Serena is unavailable or for plain-text (non-symbol) searches.

## Asking Questions

When this skill instructs you to ask the user a question, you MUST use your runtime's dedicated interactive question tool — the tool or function that presents a question to the user and **pauses execution until the user responds**. Do not output questions as plain assistant text and continue generating; always use the mechanism that blocks until the user has answered.

If your runtime does not provide such a tool, present the question as your complete message and stop generating. Do not answer your own question or proceed without user input.

## Anti-Pattern: "This Feature Is Too Simple For Full Brainstorming"

Every PRD goes through the full brainstorming process. A single button, a minor workflow tweak, a configuration option — all of them. "Simple" features are where unexamined business assumptions cause the most rework. The brainstorming can be brief for genuinely simple features, but you MUST ask clarifying questions and get approval on the product approach before writing the artifact.

## Anti-Pattern: End-Of-Flow Bureaucracy

Once the user has answered the clarifying questions and approved an approach, do not force them through a second approval loop for Overview, Goals, User Stories, or any other final document section. Synthesize the approved direction into the PRD directly. The user can review and request edits in the generated file afterward.

## Anti-Pattern: Technical Drift On Technical-Sounding Features

When the feature name sounds technical (e.g., "webhook notifications", "CSV export", "dark mode", "API rate limiting"), you will be tempted to discuss HOW to implement it. Resist this. Your job is the WHAT and WHY:

- WRONG: "Should we use WebSockets or polling for notifications?" (implementation)
- WRONG: "What CSV library format should we target?" (implementation)
- RIGHT: "Which events should trigger a notification to the user?" (user need)
- RIGHT: "What information do users need in their exported reports?" (user need)

Translate every technical-sounding feature into the user experience question behind it.

## Required Inputs

- Feature name or product idea.
- Optional: existing `_idea.md` file as primary input for context.
- Optional: existing `_prd.md` file for update mode.

## Resolving the `.rc` base directory

rc supports monorepos, where more than one `.rc` directory can exist. Before reading or writing any `.rc/...` path, resolve which `.rc` directory this run uses; its parent is the base directory. Treat every `.rc/...` path in this skill as relative to that base.

1. Search the project recursively for `.rc` directories, skipping `node_modules`, `.git`, `vendor`, and any `_archived/` directory.
2. Resolve the base from what you find:
   - **None found** — use `.rc/` at the project root, creating it on first write. Ordinary single-folder projects behave exactly as before.
   - **Exactly one found** — use it without asking.
   - **Two or more found** — ask the user which `.rc` to use via the interactive question tool that pauses execution, listing the discovered directories by their path relative to the project root.

## Checklist

You MUST create a task for each phase and complete them in order:

1. **Determine project & directory** — derive numbered slug, create `.rc/tasks/<slug>/` and `adrs/`
2. **Discover context** — parallel codebase exploration and web research
3. **Understand the need** — ask up to 3 targeted questions to refine scope and intent
4. **Present product approaches** — offer 2-3 approaches with trade-offs, create ADR for the chosen one
5. **Draft the PRD** — write using the canonical template from `references/prd-template.md`
6. **Review with user** — present the draft, iterate until approved
7. **Save the file** — write to `.rc/tasks/<slug>/_prd.md`

## Workflow

1. Determine the project name and working directory.
   - Resolve the `.rc` base directory as described in "Resolving the `.rc` base directory" above; every `.rc/...` path below is relative to it.
   - Derive a base slug from the feature name provided by the user (e.g., `add-field-tag`).
   - Prepend a zero-padded sequential counter to the base slug so directories stay ordered: the final slug is `<NN>-<base-slug>` (e.g., `01-add-field-tag`).
     - List the immediate child directories of `.rc/tasks/` (do NOT descend into `_archived/` or other underscore-prefixed directories) and find those whose name starts with a numeric prefix matching `^[0-9]+-`.
     - Take the highest numeric prefix found and increment it by 1; if no numbered directory exists yet, start at `1`. Zero-pad to at least two digits (`01`, `02`, … `10`, `11`).
     - **Update mode:** if the user is targeting an existing PRD directory (its `_idea.md` or `_prd.md` already exists), reuse that directory's full name as-is — do NOT allocate a new counter.
   - Use `.rc/tasks/<slug>/` as the target directory.
   - If `_idea.md` exists in the target directory, read it as primary context input.
   - If `_prd.md` already exists in the target directory, read it and operate in update mode.
   - If the directory does not exist, create it.
   - Create `.rc/tasks/<slug>/adrs/` directory if it does not exist.

2. Discover context through parallel research. You MUST perform BOTH tracks before asking any questions.

   **Track A — Codebase exploration** (REQUIRED):
   - Search the codebase for files, patterns, and features related to the user's request.
   - Look for existing implementations, data models, and integration points that are relevant.
   - Summarize what you found in 3-5 bullet points.

   **Track B — Market and user research** (REQUIRED):
   - Perform 3-5 web searches for market trends, competitive products, and user needs related to the feature.
   - Look for how similar products solve this problem and what users expect.
   - Summarize what you found in 3-5 bullet points.

   Run both tracks in parallel (e.g., two Agent tool calls, two search batches, etc.). Present a brief merged summary of findings from BOTH tracks to the user before moving to questions. If web search tools are unavailable, note the limitation explicitly and proceed with codebase findings only.

3. Ask clarifying questions following `references/question-protocol.md`.
   - Focus exclusively on WHAT features users need, WHY it provides business value, and WHO the target users are.
   - Ask about success criteria and constraints.
   - Never ask technical implementation questions about databases, APIs, frameworks, or architecture.
   - **ONE question per message — strictly enforced.** Your message must contain exactly one question mark. After asking the question, STOP. Do not add follow-up questions, "also" questions, or "additionally" prompts. If a topic needs more exploration, ask a follow-up in the NEXT message after the user responds.

     Anti-pattern (FORBIDDEN):
     "What is the primary user persona? Also, what are the key success metrics?"
     This is TWO questions. Split them into two separate messages.

   - Every question MUST be multiple-choice when reasonable options can be predetermined. Format as labeled options (A, B, C, etc.) so the user can respond with a single letter. Only use open-ended questions when the answer space is genuinely unbounded (e.g., "What problem are you trying to solve?").
   - Include a fallback option (e.g., "D) Other — describe") for flexibility.
   - For complex features with many dimensions, decompose into sub-topics and ask about one dimension at a time. Each sub-topic usually has predeterminable options. Example: instead of the open-ended "What should the collaboration feature include?", ask "Which aspect of team collaboration is most important to start with? A) Shared workspaces B) Real-time presence C) Permission controls D) Activity feeds".
   - Complete at least one full clarification round before presenting approaches.

4. Present product approaches.
   - Offer 2-3 product approaches with trade-offs for each.
   - Lead with the recommended approach and explain why it is preferred.
   - Wait for the user to select an approach before continuing.
   - After the user selects an approach, create an ADR for this decision:
     - Read `references/adr-template.md`.
     - Determine the next ADR number by listing existing files in `.rc/tasks/<slug>/adrs/`.
     - Fill the template: the selected approach as "Decision", rejected approaches as "Alternatives Considered" with their trade-offs, and outcomes as "Consequences". Set Status to "Accepted" and Date to today.
     - Write the ADR to `.rc/tasks/<slug>/adrs/adr-NNN.md` (zero-padded 3-digit number, e.g., `adr-001.md`).

5. Draft the PRD.
   - After the user selects an approach, synthesize the final product design. Do not present each section for separate approval.
   - If the user makes a significant scope decision during clarification or approach selection, create an additional ADR following the same process as step 4.
   - Only pause before writing if a blocking ambiguity remains that would force guessing; otherwise proceed directly to document generation.
   - Read `references/prd-template.md` and fill every section with gathered context.
   - Include an "Architecture Decision Records" section listing all ADRs created during this session with their numbers, titles, and one-line summaries as links to the `adrs/` directory.
   - Apply YAGNI ruthlessly: challenge every feature and remove anything the MVP does not need.
   - The PRD must describe user capabilities and business outcomes only.
   - No databases, APIs, code structure, frameworks, testing strategies, or architecture decisions.
   - Mandatory sections (ALWAYS include): Overview, Goals, User Stories, Core Features, User Experience, Non-Goals, Phased Rollout Plan, Success Metrics, Risks and Mitigations, Architecture Decision Records, Open Questions.
   - Optional sections (include when relevant): High-Level Technical Constraints.
   - Prefer active voice, omit needless words, use definite and specific language over vague generalities. Every sentence should earn its place.
   - Language: **English**. Tone: clear, technical, consistent with existing project artifacts.
   - Present the complete draft to the user for review.

6. Review with the user.
   - Present the draft and ask using the interactive question tool:
     - "Here is the PRD draft. Please review and let me know:"
     - A) Approved — save as is
     - B) Adjust specific sections (tell me which ones)
     - C) Rewrite section X (tell me what to change)
     - D) Discard and start over
   - If B or C: make the changes and present again.
   - If D: go back to step 3.

7. Save the PRD file.
   - Write the completed document to `.rc/tasks/<slug>/_prd.md`.
   - Confirm the file path to the user.
   - Remind the user that the next step is to create a TechSpec using `rc-create-techspec` from this PRD.

## Project memory

Before brainstorming, consult the per-project memory: run `rc memory search` with the
feature's key terms to recover prior product decisions, conventions, and constraints that
should inform this PRD (see the `rc-project-memory` skill). It is keyword-ranked, so search
the concrete nouns of the feature.

## Error Handling

- If the user provides insufficient context to complete a section, note it in the Open Questions section rather than guessing.
- If web research tools are unavailable, proceed with codebase exploration only and note the limitation.
- If the target directory cannot be created, stop and report the filesystem error.
- If operating in update mode, preserve sections the user has not asked to change.
