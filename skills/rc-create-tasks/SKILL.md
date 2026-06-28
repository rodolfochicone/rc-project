---
name: rc-create-tasks
description: Decomposes PRDs and TechSpecs into detailed, independently implementable task files with enrichment from codebase exploration. Use when a PRD or TechSpec exists and needs to be broken down into executable tasks, or when task files need enrichment with implementation context. Do not use for PRD creation, TechSpec generation, or direct task execution.
argument-hint: "[feature-name] [prd-file]"
model: sonnet
effort: high
---

# Create Tasks

Decompose requirements into detailed, actionable task files with codebase-informed enrichment.

## Code navigation (Serena)

If the Serena MCP is available, prefer its symbolic tools over whole-file reads when enriching tasks from the codebase — they are LSP-accurate and token-efficient:

- `get_symbols_overview` to grasp a file's structure before reading it; `find_symbol` (by name path, e.g. `Type/method`) to jump straight to a definition.
- `find_referencing_symbols` to map every caller of a symbol before reasoning about impact.

Fall back to Grep/Glob + Read when Serena is unavailable or for plain-text (non-symbol) searches.

## Required Inputs

- Feature name identifying the `.rc/tasks/<name>/` directory.
- At minimum, `_prd.md` or `_techspec.md` in that directory.

## Resolving the `.rc` base directory

rc supports monorepos, where more than one `.rc` directory can exist. Before reading or writing any `.rc/...` path, resolve which `.rc` directory this run uses; its parent is the base directory. Treat every `.rc/...` path in this skill as relative to that base.

1. Search the project recursively for `.rc` directories, skipping `node_modules`, `.git`, `vendor`, and any `_archived/` directory.
2. Resolve the base from what you find:
   - **None found** — use `.rc/` at the project root, creating it on first write. Ordinary single-folder projects behave exactly as before.
   - **Exactly one found** — use it without asking.
   - **Two or more found** — select the `.rc` whose `tasks/` directory contains the feature's `<NN>-<slug>` directory. If the feature exists under more than one `.rc` (or under none), ask the user which `.rc` to use via the interactive question tool that pauses execution, listing the discovered directories by their path relative to the project root.

## Workflow

1. Load type registry.
   - Resolve the `.rc` base directory as described in "Resolving the `.rc` base directory" above; every `.rc/...` path below is relative to it.
   - Read `.rc/config.toml`.
   - If it contains `[tasks].types`, use that list as the allowed `type` values.
   - Otherwise use the built-in defaults: `frontend`, `backend`, `docs`, `test`, `infra`, `refactor`, `chore`, `bugfix`.

2. Load context.
   - Read `_prd.md` and `_techspec.md` from `.rc/tasks/<name>/`.
   - Read existing ADRs from `.rc/tasks/<name>/adrs/` to understand the decision context behind requirements and design choices.
   - If `_techspec.md` is missing:
     - Warn the user that tasks will be higher-level without TechSpec implementation guidance.
     - Derive tasks from PRD functional requirements and user stories instead of TechSpec implementation sections.
     - During enrichment, rely more heavily on codebase exploration to fill `## Implementation Details`, `### Relevant Files`, and `### Dependent Files`.
     - Mark `<requirements>` with PRD-derived behavioral requirements instead of TechSpec-derived technical requirements.
     - Explicitly call out missing implementation detail gaps in the task body instead of inventing specifics.
   - If both `_prd.md` and `_techspec.md` are missing, stop and ask the user to create at least one first.
   - Spawn an Agent tool call to explore the codebase for files to create or modify, test patterns, and coding conventions.

3. Break down into tasks.
   - **Challenge scope before decomposing (YAGNI).** The TechSpec already applied YAGNI to the architecture; here the guard is task *scope*. Before turning a requirement into a task, ask: does this need to exist in this delivery, or is it speculative work beyond what the PRD asks? Can the codebase exploration from step 2 already cover it (an existing helper, pattern, or endpoint)? Drop speculative tasks and merge tasks that only scaffold "for later". Never silently remove work the PRD explicitly requests — when something looks like YAGNI but sits in PRD scope, keep it and flag it for the user in the step 4 approval instead.
   - Decompose implementation sections from the TechSpec into granular, independently implementable tasks.
   - **Each task must be independently implementable once all of its declared dependencies are met**, because the executor runs tasks in isolation and any undeclared coupling is what makes one fail mid-run. No task may require undeclared work from another task. If two tasks share a tight coupling, either merge them or extract the shared piece into a dependency task.
   - **No circular dependencies**: if task A depends on task B, task B must not depend on task A (directly or transitively), or neither can ever start.
   - Each task must have: title, type, complexity, and dependencies.
   - Assign complexity using these criteria:
     - `low`: Single file change, no new interfaces, no concurrency, straightforward logic.
     - `medium`: 2-4 files, may introduce a new interface or struct, limited integration points.
     - `high`: 5+ files, new subsystem or significant refactor, multiple integration points, concurrency involved.
     - `critical`: Cross-cutting change affecting many packages, high risk of regression, requires coordination with other tasks.
   - When a task directly implements or is constrained by a specific ADR, include the ADR reference in the task's "Related ADRs" section under Implementation Details.
   - Embed test requirements in every task. Never create separate tasks dedicated solely to testing.
   - Follow the structure defined in `references/task-template.md`.
   - Refer to `references/task-context-schema.md` for metadata field definitions.

4. Present task breakdown for interactive approval.
   - Show all tasks with: titles, descriptions, complexity ratings, and dependency chains.
   - Wait for user feedback before proceeding.
   - If the user requests changes, revise the breakdown and present again.
   - Iterate until the user explicitly approves.

5. Generate task files.
   - Write `_tasks.md` as the master task list using this exact markdown table format:
     ```markdown
     # [Feature Name] — Task List

     ## Tasks

     | # | Title | Status | Complexity | Dependencies |
     |---|-------|--------|------------|--------------|
     | 01 | [Task title] | pending | [low/medium/high/critical] | [task_NN, ... or —] |
     ```
   - Write individual task files as `task_01.md`, `task_02.md`, through `task_N.md`.
   - Task files use the `task_` prefix without a leading underscore.
   - Each file must start with YAML frontmatter containing `status`, `title`, `type`, `complexity`, and `dependencies`. Use `dependencies: []` when there are no dependencies — do not omit the field.
   - Task numbering must be sequential and consistent between `_tasks.md` and individual files.

6. Enrich each task file.
   - For each task file, check whether it already has `## Overview`, `## Deliverables`, and `## Tests` sections. If all three exist, skip enrichment for that file.
   - Map the task to PRD requirements and TechSpec guidance.
   - Spawn an Agent tool call to discover relevant files, dependent files, integration points, and project rules for this specific task.
   - Fill all template sections from `references/task-template.md`. Every task file must contain each of the following sections, because the task executor reads them as its contract and a missing section leaves it guessing at intent:
     - `## Overview`: what the task accomplishes and why, in 2-3 sentences.
     - `<critical>` block: the standard critical reminders block (read PRD/TechSpec, reference TechSpec, focus on WHAT, minimize code, tests required).
     - `<requirements>` block: specific, numbered technical requirements using MUST/SHOULD language.
     - `## Subtasks`: 3-7 checklist items describing WHAT, not HOW.
     - `## Implementation Details`: file paths to create or modify, integration points. Reference TechSpec for patterns.
     - `### Relevant Files`: discovered paths from codebase exploration with brief reasons.
     - `### Dependent Files`: files that will be affected by this task with brief reasons.
     - `### Patterns to Mirror`: 1-3 short snippets of **real code already in this repo** that the implementation should imitate — the existing error-wrap style, the table-test shape, the handler/constructor pattern. Copy the actual lines (≤10 each) and tag each with `// SOURCE: <path>:<start>-<end>` so the executor mirrors what exists instead of inventing a parallel style. Capture this now (you already explored the codebase) — if the executor would have to grep the repo to find "how we do X here", that knowledge belongs in the task. Omit only when the task genuinely introduces a pattern with no precedent.
     - `### Related ADRs`: links to relevant ADRs if any exist, or omit subsection if no ADRs apply.
     - `## Deliverables`: concrete outputs with mandatory test items and at least 80% coverage target.
     - `## Tests`: specific test cases as checklists, split into unit tests and integration tests categories.
     - `## Success Criteria`: measurable outcomes including "All tests passing" and "Test coverage >=80%".
   - Reassess complexity based on exploration findings and update if changed.
   - **No Prior Knowledge Test.** Before finalizing each task file, apply this check: *could a competent developer (or executor agent) with no prior context implement this task correctly using only the task file plus the referenced TechSpec sections, without having to grep the repo to discover how things are done here?* If not, the gap is missing context — add the concrete file paths, the Patterns to Mirror snippet, or the specific TechSpec/ADR reference that closes it. A task that only an author-who-already-knows could implement is under-enriched.
   - Update the task file in place with enriched content.
   - If enrichment fails for one task, continue to the next and report all failures at the end.

7. Run task validation.
   - Run `rc tasks validate --name <feature>`.
   - If it exits non-zero, fix the reported issues and re-run.
   - Do not mark the skill complete until it exits 0.

## Project memory

Before decomposing, run `rc memory search` with the feature and package terms to recover
conventions and gotchas that should shape task boundaries and implementation notes (see the
`rc-project-memory` skill).

## Anti-Patterns

Do NOT produce tasks with these defects:

- **Mega-tasks.** If a task touches more than 7 files or has more than 7 subtasks, it is too broad. Split it into smaller tasks with explicit dependencies between them.
- **TechSpec duplication.** Do NOT copy interface definitions, code snippets, or architectural diagrams from the TechSpec into task files. Reference the TechSpec section by name (e.g., "See TechSpec 'Core Interfaces' section") instead of reproducing its content.
- **Vague test cases.** Do NOT write test descriptions like "test the happy path" or "verify error handling." Each test case must name the specific input, condition, or behavior being verified (e.g., "POST /job/done with unknown job ID returns 404").
- **Speculative tasks.** Do NOT create tasks for work the PRD does not ask for — config nobody sets, an abstraction with a single caller, scaffolding "for later". If the TechSpec deliberately cut something, do not reintroduce it as a task.

## Error Handling

- If both `_prd.md` and `_techspec.md` are missing, stop and ask the user to create at least one first.
- If the user rejects the task breakdown, incorporate all feedback before presenting again.
- If codebase exploration reveals task boundaries that do not match the TechSpec, note the discrepancy and ask the user how to proceed.
- If the target directory does not exist, create it.
- If a task file already exists and is fully enriched, skip it and move to the next.
