---
name: rc-execute-task
description: Executes one PRD task end-to-end using a provided task file, PRD directory, tracking file paths, and auto-commit mode. Use when a prompt includes a task specification that must be implemented, validated, and reflected in task tracking files. Do not use for PR review batches, generic coding tasks without a PRD task file, or standalone verification-only work.
user-invocable: false
model: sonnet
effort: high
---

# Execute PRD Task

Execute one PRD task from exploration through tracking updates.

## Code navigation & editing (Serena)

If the Serena MCP is available, prefer its symbolic tools over whole-file reads and line-based edits — they are LSP-accurate and token-efficient:

- `get_symbols_overview` to grasp a file's structure before reading it; `find_symbol` (by name path, e.g. `Type/method`) to jump straight to a definition.
- `find_referencing_symbols` to map every caller before changing a symbol (impact analysis).
- `replace_symbol_body`, `insert_after_symbol`, `insert_before_symbol` for precise edits that don't depend on line numbers.

Fall back to Grep/Glob + Read/Edit when Serena is unavailable or for plain-text (non-symbol) searches.

## Required Inputs

- Task specification markdown.
- PRD directory path.
- Task file path.
- Master tasks file path.
- Auto-commit mode.
- Optional workflow memory directory path.
- Optional shared workflow memory path.
- Optional current task memory path.

## Workflow

> **Delegation.** Keep this skill's context lean by routing to specialist subagents per the
> delegation contract in the `rc` skill (`references/delegation-contract.md`): hand broad or
> unfamiliar recon to a `rc-explorer` (cheap/fast, read-only) instead of reading widely
> yourself, route version-specific library/API lookups to a `rc-librarian` (cheap, read-only)
> instead of guessing from memory, and route architecture/security review or a stubborn debug to
> `rc-oracle` (strong model). Do the bounded implementation here yourself — that is this skill's
> lane. When a task splits into genuinely independent, well-scoped mechanical sub-changes, the
> upgrade path is worktree-isolated `rc-fixer`s with per-folder ownership (see the contract);
> do not fan out writers over a shared tree.

1. Ground in repository and PRD context.
   - Read the provided task specification.
   - Read the repository guidance files named by the caller.
   - Read the PRD documents under the provided directory, especially `_techspec.md` and `_tasks.md`.
   - Read ADRs from the `adrs/` subdirectory of the PRD directory to understand the architectural decision context for this task.
   - After reading all sources, check for conflicts between the task specification, techspec, and ADRs. If any requirements contradict each other, stop and report the conflict instead of guessing — do not proceed to step 2.
   - If the caller provides workflow memory paths, use the installed `rc-workflow-memory` skill before editing code.
   - Consult the per-project memory before implementing via the `rc-memory` skill (scan `.rc/memory/INDEX.md`) for the task's key terms to recover prior decisions, conventions, and known gotchas.
   - Reconcile the current workspace state before new edits.

2. Build the execution checklist.
   - Extract deliverables, acceptance criteria, and every explicit `Validation`, `Test Plan`, or `Testing` item into a numbered working checklist.
   - Print the full checklist before starting implementation so it is visible and trackable.
   - Capture the concrete pre-change signal that proves the task is not finished yet.
   - Use this checklist as a gate: mark each item done as evidence is produced during implementation, and do not proceed to validation until every checklist item has been addressed.

3. Implement the task.
   - **Climb the laziness ladder before writing.** You have already understood the task in steps 1-2; now write the *least* code that satisfies it. Stop at the first rung that holds:
     1. **Does this need to exist?** If a piece is speculative or beyond the task's contract, record it as a follow-up note (below) instead of building it.
     2. **Already in the codebase?** Reuse the helper, util, type, or pattern that already lives here — use Serena `find_symbol` / `find_referencing_symbols` to find it before writing a new one. Re-implementing what sits a few files over is the most common slop.
     3. **Standard library does it?** Use it before hand-rolling.
     4. **A dependency already in `go.mod` solves it?** Use it. Never reach for a new dependency for what a few lines cover (and never edit `go.mod` by hand — see CLAUDE.md).
     5. **Only then:** write the minimum idiomatic code that works.
   - **Never simplify away** input validation at trust boundaries, error handling that prevents data loss, security, or correct concurrency — and never drop anything the task explicitly requests. A bug fix stays a root-cause fix, not a symptom patch (see the `rc-no-workarounds` skill). The ladder shortens the solution, never the comprehension in steps 1-2.
   - Keep scope tight to the task specification.
   - Follow repository patterns and real dependency APIs.
   - Record meaningful out-of-scope work as follow-up notes instead of silently expanding the task.

4. Validate in a bounded verify→fix loop.
   - Run every test and validation command listed in the task specification — not just the repository-wide verification — plus the installed `rc-final-verify` skill. This gate is the loop's until-condition and is mandatory regardless of auto-commit mode — always verify before claiming completion.
   - If the gate is green on the first pass, perform a self-review, resolve any blocking issue it surfaces, and proceed.
   - If the gate is red, iterate **gather → fix root cause → re-verify**, up to **3 fix cycles** for this task. Diagnose the real failure before each fix — no symptom patches (see the `rc-no-workarounds` skill). On the 3rd (final) cycle, route the diagnosis to `rc-oracle` (strong model, stubborn-debug) per the delegation contract before attempting the fix.
   - **Never mark the task complete on a red gate.** If the gate is still red after the 3rd cycle, stop, leave the task status unchanged, and report it as blocked — with the failing output and what was tried — instead of declaring premature success.

5. Update task tracking.
   - If workflow memory paths were provided, update the memory files first — record decisions, learnings, and touched surfaces before updating tracking status.
   - Record durable project-level facts via the `rc-memory` skill — a cross-cutting decision, a new convention, or a non-obvious gotcha that future runs need — but only facts not already obvious from the repository, PRD, or techspec.
   - Use the caller-provided task file path and master tasks file path.
   - Mark subtasks complete only when the implementation and evidence are actually complete.
   - Change task status to completed only after clean verification and self-review.
   - Read `references/tracking-checklist.md` when applying status, checklist, or commit updates.
   - Sequence: workflow memory update (if applicable) -> record durable project facts via the `rc-memory` skill (if any) -> task file checkboxes -> task status -> master tasks file -> commit (if applicable).

6. Handle commit behavior.
   - If auto-commit is enabled, create one local commit after clean verification, self-review, and tracking updates.
   - If auto-commit is disabled, leave the diff ready for manual review and commit.
   - Never push automatically.

## Error Handling

- If the pre-change signal cannot be reproduced directly, capture the strongest available baseline signal and state the limitation.
- If validation fails, run the bounded verify→fix loop from step 4: keep the task status unchanged until the gate is green, or until the 3-cycle cap is hit — then report the task as blocked with the failing evidence. Never mark it complete on a red gate.
- If tracking files are missing, stop and report the missing path before marking completion.
