---
name: rc-create-techspec
description: Creates a Technical Specification by translating PRD business requirements into implementation designs through interactive technical clarification. Use when a PRD exists and needs a technical plan, or when technical architecture decisions need documentation. Do not use for PRD creation, task breakdown, or direct code implementation.
argument-hint: "[feature-name] [prd-file]"
model: opus
effort: high
---

# Create TechSpec

Translate business requirements into a detailed technical specification.

<HARD-GATE>
The design clarification must precede the artifact, because a TechSpec written before the phases finish encodes assumptions about the existing architecture — and those are exactly the assumptions that cause integration failures downstream. This holds for every TechSpec regardless of perceived simplicity:
- Do not write the TechSpec file until all phases are complete and the user has approved the final draft — an unapproved file is a design decision the user never signed off on.
- Do not skip the codebase exploration: each TechSpec is informed by the existing architecture so the design fits what is already there instead of fighting it.
- Do not skip user interactions: the user shapes the TechSpec at every decision point, because architectural intent and constraints often live with them, not in the code.
</HARD-GATE>

## Code navigation (Serena)

If the Serena MCP is available, prefer its symbolic tools over whole-file reads when grounding the technical design in the codebase — they are LSP-accurate and token-efficient:

- `get_symbols_overview` to grasp a file's structure before reading it; `find_symbol` (by name path, e.g. `Type/method`) to jump straight to a definition.
- `find_referencing_symbols` to map every caller of a symbol before reasoning about impact.

Fall back to Grep/Glob + Read when Serena is unavailable or for plain-text (non-symbol) searches.

## Asking Questions

When this skill instructs you to ask the user a question, you MUST use your runtime's dedicated interactive question tool — the tool or function that presents a question to the user and **pauses execution until the user responds**. Do not output questions as plain assistant text and continue generating; always use the mechanism that blocks until the user has answered.

If your runtime does not provide such a tool, present the question as your complete message and stop generating. Do not answer your own question or proceed without user input.

## Required Inputs

- Feature name identifying the `.rc/tasks/<name>/` directory.
- Optional: existing `_prd.md` as primary input.
- Optional: existing `_techspec.md` for update mode.

## Resolving the `.rc` base directory

RC supports monorepos, where more than one `.rc` directory can exist. Before reading or writing any `.rc/...` path, resolve which `.rc` directory this run uses; its parent is the base directory. Treat every `.rc/...` path in this skill as relative to that base.

1. Search the project recursively for `.rc` directories, skipping `node_modules`, `.git`, `vendor`, and any `_archived/` directory.
2. Resolve the base from what you find:
   - **None found** — use `.rc/` at the project root, creating it on first write. Ordinary single-folder projects behave exactly as before.
   - **Exactly one found** — use it without asking.
   - **Two or more found** — select the `.rc` whose `tasks/` directory contains the feature's `<NN>-<slug>` directory. If the feature exists under more than one `.rc` (or under none), ask the user which `.rc` to use via the interactive question tool that pauses execution, listing the discovered directories by their path relative to the project root.

## Checklist

You MUST create a task for each phase and complete them in order:

1. **Gather context** — read PRD, ADRs, and explore codebase architecture
2. **Ask technical questions** — ask up to 3 targeted questions on architecture, data models, APIs, testing
3. **Create ADRs** — record significant technical decisions (architecture pattern, technology choices, data model approach)
4. **Draft the TechSpec** — write using the canonical template from `references/techspec-template.md`
5. **Review with user** — present the draft, iterate until approved
6. **Save the file** — write to `.rc/tasks/<name>/_techspec.md`

## Workflow

1. Gather context.
   - **Resolve the working directory.** First resolve the `.rc` base directory as described in "Resolving the `.rc` base directory" above; the `.rc/tasks/` lookup below is relative to it. PRD directories carry a zero-padded numeric prefix (`<NN>-<base-slug>`, e.g., `01-add-field-tag`) created by `rc-create-prd`. Match the feature name the user provided against the part after the `<NN>-` prefix to locate the existing directory under `.rc/tasks/` (ignore `_archived/` and other underscore-prefixed directories). If exactly one matches, use it. If several match, ask the user which one. If none matches and you must create a new directory (no PRD context exists), allocate the next counter the same way `rc-create-prd` does: highest existing `^[0-9]+-` prefix + 1, starting at `01`, zero-padded to at least two digits.
   - Check for `_prd.md` in `.rc/tasks/<name>/`. If it exists, read it as the primary input.
   - If no PRD exists, ask the user for a description of what needs technical specification.
   - Read existing ADRs from `.rc/tasks/<name>/adrs/` to understand decisions already made during PRD creation.
   - Create `.rc/tasks/<name>/adrs/` directory if it does not exist.
   - Spawn an Agent tool call to explore the codebase for architecture patterns, existing components, dependencies, and technology stack.
   - If `_techspec.md` already exists, read it and operate in update mode.

2. Ask technical clarification questions.
   - Focus on HOW to implement, WHERE components live, and WHICH technologies to use.
   - Cover architecture approach and component boundaries.
   - Cover data models and storage choices.
   - Cover API design and integration points.
   - Cover testing strategy and performance requirements.
   - Ask only one question per message. If a topic needs more exploration, break it into a sequence of individual questions.
   - Prefer multiple-choice questions when the options can be predetermined.
   - Include a fallback option (e.g., "D) Other — describe") for flexibility.

3. Create ADRs for significant technical decisions.
   - For each significant decision (architecture pattern chosen, technology selected, data model approach, etc.):
     - Read `references/adr-template.md`.
     - Determine the next ADR number by listing existing files in `.rc/tasks/<name>/adrs/`.
     - Fill the template: the chosen design as "Decision", rejected alternatives as "Alternatives Considered", and trade-offs as "Consequences". Set Status to "Accepted" and Date to today.
     - Write each ADR to `.rc/tasks/<name>/adrs/adr-NNN.md` (zero-padded 3-digit sequential number).

4. Draft the TechSpec.
   - Synthesize the approved direction into the document. Do not present each section for separate approval — the user reviews the complete draft in step 5.
   - Read `references/techspec-template.md` and fill every applicable section.
   - **MANDATORY — Architecture Decision Records section:** The generated TechSpec MUST end with an "Architecture Decision Records" section listing every ADR created during this process. Each entry must include the ADR number (e.g., ADR-001), title, and a one-line summary formatted as a link to the `adrs/` directory. Even simple features require at least one ADR documenting the primary technical approach chosen and alternatives rejected. If no ADRs were created in step 3, go back and create at least one before generating the document.
   - Apply YAGNI ruthlessly: remove any component, interface, or abstraction that is not strictly necessary. Do NOT propose new packages or directories when the feature can be implemented by adding a single file to an existing package.
   - **Trade-offs are mandatory:** the Executive Summary MUST state the primary technical trade-off of the chosen approach.
   - Every PRD goal and user story should map to a technical component.
   - Reference PRD sections by name but do not duplicate business context.
   - Include code examples only for core interfaces, limited to 20 lines each. The Core Interfaces section must contain at least one type definition as a code block, written in the project's own language and idiom (a Rust trait or struct, a TypeScript interface or type, a Python protocol or dataclass — whatever the codebase already uses), even for simple features — show the primary type that other components will depend on.
   - The Development Sequencing section MUST include a numbered Build Order where every step after the first explicitly states which previous steps it depends on.
   - **MANDATORY — Behavioral Contract section:** end the technical body (before the ADR section) with a "Behavioral Contract" of machine-parseable assertions, so requirements are verifiable and grep-able rather than buried in prose. Use the format in "Behavioral contract format" below. Derive the assertions from the PRD's acceptance criteria and the design decisions; each gets a stable `id` that later artifacts (tasks, reviews) reference. This is the durable, checkable core of the spec — write it for every TechSpec, even simple ones.
   - Prefer active voice, omit needless words, use definite and specific language over vague generalities. Every sentence should earn its place.
   - Language: **English**. Tone: clear, technical, consistent with existing project artifacts.
   - Present the complete draft to the user for review.

5. Review with the user.
   - Present the draft and ask using the interactive question tool:
     - "Here is the TechSpec draft. Please review and let me know:"
     - A) Approved — save as is
     - B) Adjust specific sections (tell me which ones)
     - C) Rewrite section X (tell me what to change)
     - D) Discard and start over
   - If B or C: make the changes and present again.
   - If D: go back to step 2.

6. Save the TechSpec file.
   - Write the completed document to `.rc/tasks/<name>/_techspec.md`.
   - Confirm the file path to the user.
   - Remind the user that the next step is to create tasks using `rc-create-tasks` from this TechSpec.

## Behavioral contract format

The Behavioral Contract is a flat list of atomic assertions — each one independently checkable — with stable identifiers in HTML comments so they survive edits and can be grepped by reviewers and task files. Two kinds:

- **Requirement** — a behavior that must hold *when triggered* (an action/endpoint/flow does X under conditions Y).
- **Invariant** — a property that must hold *at all times* (a constraint the system never violates).

```
### Requirement: reject expired tokens
When a request presents an expired auth token, the API responds 401 and does not touch the session store.
<!-- id: auth.reject-expired -->
<!-- enforced: TestAuth_RejectsExpired (internal/auth) -->
<!-- depends_on: auth.session-shape -->

### Invariant: session store is append-only
Session records are never mutated in place; updates write a new revision.
<!-- id: auth.session-append-only -->
<!-- enforced: pending -->
```

Rules for the contract:

- `id` is `<area>.<short-kebab>`, **stable** — never renumber on edit; deletions are explicit. It is the anchor other artifacts cite.
- `enforced` names the test/check that proves the assertion, or `pending` if not yet covered (a reviewer can grep `enforced: pending` to find unverified requirements).
- `depends_on` (optional) lists other assertion ids this one builds on.
- Keep each assertion to one sentence, observable and testable. Vague assertions ("should be fast") are not allowed — quantify or drop them.

`rc-create-tasks` references these ids per task; `rc-code-review`/`rc-review-round` can grep `enforced:` to detect requirements that drifted out of coverage.

## Project memory

Before designing, consult project memory (the `rc-memory` skill, scanning `.rc/memory/INDEX.md`) for the feature and component terms to recover
prior architectural decisions, conventions, and known gotchas (see the `rc-memory`
skill). After the techspec is settled, record any durable cross-cutting decision with
the `rc-memory` skill (scope: decision) — only decisions not already captured in the techspec or
the ADRs.

## Error Handling

- If the PRD is missing, proceed with user-provided context and note the absence in the Executive Summary.
- If codebase exploration reveals conflicting architectural patterns, document both and recommend one with rationale.
- If the user rejects the design proposal, incorporate all feedback and present a revised proposal.
- If the target directory does not exist, create it.
- If operating in update mode, preserve sections the user has not asked to change.
