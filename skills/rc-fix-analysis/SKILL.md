---
name: rc-fix-analysis
description: Implements the change planned by a prior rc-analyze report — reads the report's "Implementation plan", applies the root-cause code changes step by step, adds or adjusts tests that encode why the change matters, and verifies with the project's gate. Use when an analysis from rc-analyze exists and the user wants its plan executed. Do not use to investigate or plan (use rc-analyze), to review a change set for defects (use rc-code-review), to remediate external PR review issues (use rc-fix-reviews), or to execute a PRD task file (use rc-execute-task).
model: sonnet
effort: high
---

# Fix From Analysis

Execute the implementation plan produced by `rc-analyze`. This skill is the write-side counterpart of `rc-analyze`: it does not investigate or re-plan — it reads the plan from an existing analysis report and turns it into working, verified code. It is standalone and stack-agnostic; it detects the stack and works in that stack's idioms.

## Code navigation & editing (Serena)

If the Serena MCP is available, prefer its symbolic tools over whole-file reads and line-based edits — they are LSP-accurate and token-efficient:

- `get_symbols_overview` to grasp a file's structure before reading it; `find_symbol` (by name path, e.g. `Type/method`) to jump straight to a definition.
- `find_referencing_symbols` to map every caller before changing a symbol (impact analysis).
- `replace_symbol_body`, `insert_after_symbol`, `insert_before_symbol` for precise edits that don't depend on line numbers.

Fall back to Grep/Glob + Read/Edit when Serena is unavailable or for plain-text (non-symbol) searches.

## Required Inputs

- The path to the `rc-analyze` report to execute, or enough context to locate the most recent one (a feature slug or topic). The report must contain an **Implementation plan** section.
- Optional: a scope to limit which plan steps to apply (specific steps, files, or directories).

## Resolving the `.rc` base directory

RC supports monorepos, where more than one `.rc` directory can exist. Before reading or writing any `.rc/...` path, resolve which `.rc` directory this run uses; its parent is the base directory. Treat every `.rc/...` path in this skill as relative to that base.

1. Search the project recursively for `.rc` directories, skipping `node_modules`, `.git`, `vendor`, and any `_archived/` directory.
2. Resolve the base from what you find:
   - **None found** — there is no analysis to execute; stop and tell the user to run `rc-analyze` first.
   - **Exactly one found** — use it without asking.
   - **Two or more found** — if the prompt names a feature that maps to one `.rc/tasks/<slug>/`, use that `.rc`. Otherwise ask the user which `.rc` to use via the interactive question tool that pauses execution, listing the discovered directories by their path relative to the project root.

## Locating the analysis report

1. If the user gave an explicit report path, use it.
2. Otherwise, search for analysis reports under the resolved base: `.rc/tasks/<slug>/analysis-NNN.md` (feature-scoped) and `.rc/analysis/<topic-slug>-NNN.md` (topic-scoped). Prefer the highest `NNN` matching the named slug or topic.
3. If no report is found, or the chosen report has **no Implementation plan section**, stop. Do not investigate or invent a plan from scratch — tell the user to run `rc-analyze` first so a plan exists. (Planning is `rc-analyze`'s job; this skill only executes.)

## Workflow

1. Read the report in full. Parse the **Implementation plan** into an ordered list of steps. Read the **Summary**, **How it works**, **Key findings**, and **References** sections too, so you execute with the analysis's intent — not just the bare step list.

2. Re-ground every step against the current code. The report may be stale. For each step, open the cited `file:line` and confirm the code still matches what the plan assumes. If a step no longer applies (already fixed, code moved, assumption wrong), note the divergence, adapt the step to current reality, and surface it to the user rather than forcing a change that no longer fits.

3. Implement step by step, fixing the root cause. For each plan step:
   - Make the smallest change that correctly resolves the underlying problem the analysis identified — not a symptom patch. No workarounds: do not silence type errors with assertions, suppress linter rules, swallow errors, add timing hacks, or special-case the failing input to make a check pass.
   - Match the surrounding code's conventions, naming, and structure. Touch only what the step requires; do not refactor adjacent code.
   - Keep edits traceable to the plan step that motivated them.

4. Cover the change with tests that encode intent. Add or adjust tests so they assert *why* the behavior matters, not merely that the current code runs. A test that cannot fail when the business logic regresses is not adequate. Do not test mocks against mocks, and do not add test-only methods to production code to make something assertable.

5. Verify with the project's gate. Detect and run the stack's real verification: a declared command (e.g. `make verify`, `npm test`, `pytest`, `go test ./... -race`, `cargo test`) or the closest equivalent the project provides. Run it for real and read the full output. If it fails, fix the root cause and re-run until it passes — do not declare success on stale or partial evidence. If no gate exists, run the most authoritative checks available (build + tests) and say which.

6. Write the resolution log and report it.
   - Write to `.rc/tasks/<slug>/resolution-NNN.md` when a feature slug applies, otherwise `.rc/analysis/<topic-slug>-resolution-NNN.md`. `NNN` is zero-padded and increments past any existing matching file. Never overwrite the `analysis-NNN.md` it executed.
   - In the log, for each plan step record: the step, the files/`file:line` changed, the tests touched, any divergence from the plan and how it was handled, and the verification command with its result.
   - Print a concise summary to the user: what changed, which steps were skipped or adapted and why, and the verification outcome with evidence.

7. Close with the bottom-line result in one or two sentences, the verification status, and the resolution log path.

## Project memory

Before applying the plan, consult project memory (the `rc-memory` skill, scanning `.rc/memory/INDEX.md`) for the target terms to recover relevant
decisions and gotchas (see the `rc-memory` skill). When the change establishes a
durable decision or reveals a non-obvious gotcha, record it via the `rc-memory` skill.

## Critical Rules

- Execute the plan; do not re-analyze. If there is no plan, stop and route the user to `rc-analyze`.
- Root cause over symptom. Reject workarounds — type assertions to silence checks, lint suppressions, error swallowing, timing hacks, and input special-casing are not fixes.
- Surgical changes. Implement only what the plan's steps require; do not improve adjacent code or restyle untouched files.
- Tests must encode why the change matters and be able to fail when the logic regresses.
- Never declare done without fresh verification evidence. Run the gate, read the output, report it honestly — if it fails or a step was skipped, say so.
- Reason in the target stack's idioms. If a plan step would introduce a harmful pattern, surface it instead of silently implementing it.

## Error Handling

- **No report / no plan** — stop and instruct the user to run `rc-analyze` to produce a plan; do not fabricate one.
- **Stale plan** — when a step's cited code has changed, adapt the step to the current code and flag the divergence; do not force an obsolete edit.
- **Verification fails** — keep the change in progress (not "done"), fix the root cause, and re-run the gate. Report the failing output if you cannot resolve it.
- **Ambiguous or conflicting steps** — ask one focused clarifying question before changing code, rather than guessing.
- **A file the plan targets cannot be read or written** — note which file and stop on that step, continuing with independent steps where safe.
