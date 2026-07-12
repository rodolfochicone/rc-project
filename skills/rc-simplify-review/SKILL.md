---
name: rc-simplify-review
description: Reviews the current change set through a single lens — over-engineering and complexity only — and writes a ranked delete-list to .rc/tasks/<slug>/ (what to delete, replace with stdlib, fold into a native feature, or shrink), ending with the net lines/dependencies a cleanup could remove. Use as an opt-in pre-PR pass to catch bloat that a severity-ranked review buries, or to audit legacy code that never went through the RC ladder. Do not use for correctness, security, or performance defects (use rc-code-review), to generate a remediation round (use rc-review-round), to fix existing review issues (use rc-fix-reviews), or to edit source code.
model: opus
effort: high
---

# Simplify Review

Review the change set for over-engineering only, and report what can be cut. This is the complexity-only counterpart to `rc-code-review`: that skill asks "is this correct and safe?"; this one asks "did this need to exist, and what can be deleted?". The two are deliberately separate passes — a severity-ranked review files complexity findings as `low` and buries them under the criticals, so nobody acts on them.

## Code navigation (Serena)

If the Serena MCP is available, prefer its symbolic tools over whole-file reads — they are LSP-accurate and token-efficient, and they are how you *prove* a cut is safe before proposing it:

- `get_symbols_overview` to grasp a file's structure; `find_symbol` (by name path, e.g. `Type/method`) to jump to a definition.
- `find_referencing_symbols` to confirm a symbol is genuinely dead, single-caller, or a one-product factory before flagging it for deletion. Never mark a cut from a name alone — a function may not do what its name says.

Fall back to Grep/Glob + Read when Serena is unavailable or for plain-text searches.

## Required Inputs

- None required. Defaults to the current change set against `main`.
- Optional: specific files/directories to scope the pass, or a base ref to diff against (default `main`). To audit an existing codebase rather than a diff, the user names the directory to scan in full.

## Resolving the `.rc` base directory

RC supports monorepos, where more than one `.rc` directory can exist. Before reading or writing any `.rc/...` path, resolve which `.rc` directory this run uses; its parent is the base directory. Treat every `.rc/...` path in this skill as relative to that base.

1. Search the project recursively for `.rc` directories, skipping `node_modules`, `.git`, `vendor`, and any `_archived/` directory.
2. Resolve the base from what you find:
   - **None found** — use `.rc/` at the project root, creating it on first write. Ordinary single-folder projects behave exactly as before.
   - **Exactly one found** — use it without asking.
   - **Two or more found** — select the `.rc` whose `tasks/` directory contains the feature's `<NN>-<slug>` directory. If the feature exists under more than one `.rc` (or under none), ask the user which `.rc` to use via the interactive question tool that pauses execution, listing the discovered directories by their path relative to the project root.

## Workflow

1. Scope the change.
   - Use the user's explicit paths, or `git diff <base>...HEAD --name-only` (default base `main`). If the user asked to audit the whole repo, scan the tree instead of a diff.
   - Detect the stack from manifest and config files so every replacement is named in the stack's own idioms (stdlib function, native feature, installed dependency).
   - Read each file in scope before forming conclusions.

2. Hunt for over-engineering. Look for:
   - dependencies the standard library or platform already ships;
   - single-implementation interfaces, factories with one product, wrappers that only delegate;
   - files exporting one thing, dead flags and config, abstraction with a single caller;
   - hand-rolled code that duplicates the stdlib or an already-installed dependency;
   - layers and indirection added "for later" that nothing uses yet.

3. Confirm before flagging. For every candidate cut, prove it with `find_referencing_symbols` (truly dead / single caller) or by reading the replacement (the stdlib function exists and is edge-case-correct). A cut you cannot prove is a hypothesis — label it as such, do not assert it. Only assert a cut you are **>80% sure is safe and real** here; proposing zero cuts (`Lean already. Ship.`) is a valid, expected outcome — never invent cuts to look thorough.

4. Write the report and print it.
   - Write to `.rc/tasks/<slug>/simplify-review-NNN.md`, where `NNN` is zero-padded and increments past any existing `simplify-review-*.md`. If no feature slug applies, write to `.rc/analysis/simplify-review-NNN.md`. Print the same content.
   - Even with nothing to cut, write the report recording the clean verdict so the pass is traceable.

5. Close by noting that the user decides what to cut — the cuts can be applied manually or, where a finding implies a concrete change, executed by `rc-fix-analysis`.

## Output format

One line per finding, ranked biggest cut first:

`<tag> <what to cut>. <replacement>. [path:line]`

Tags:

- `delete:` dead code, unused flexibility, speculative feature. Replacement: nothing.
- `stdlib:` hand-rolled thing the standard library ships. Name the function.
- `native:` dependency or code doing what the platform already does (a DB constraint over app code, a built-in over a helper). Name the feature.
- `yagni:` abstraction with one implementation, config nobody sets, layer with one caller.
- `shrink:` same logic, fewer lines. Show the shorter form.

End with `net: -<N> lines, -<M> deps possible.` Nothing to cut: `Lean already. Ship.`

## When NOT to flag a cut

This is the load-bearing guard — a complexity review that cuts a necessary safeguard is worse than no review. Never propose deleting or shrinking away: input validation at trust boundaries, error handling that prevents data loss, security measures, correct concurrency (locks, context cancellation, goroutine shutdown), accessibility basics, or anything the task or PRD explicitly requested. Smaller code is the goal; the flimsier algorithm is not. When two forms are the same size, the edge-case-correct one stays.

## Critical Rules

- Do not modify source code. This skill writes its report under `.rc/...` and reports findings only.
- Complexity only. Correctness, security, and performance defects are out of scope — route them to `rc-code-review` instead of reporting them here.
- Prove every cut before proposing it; never flag from a name alone. Distinguish a confirmed cut from a hypothesis.
- The `net:` figure is counted from the real findings in this report — it is the lines those cuts would remove. Never invent a savings number or a per-change "you saved X" claim: code that was never written has no baseline to subtract from.
- Express every replacement in the target stack's idioms, with a concrete, actionable form.

## Error Handling

- If the diff is empty or unhelpful and the user gave no paths, ask the user to specify the scope.
- If a file in scope cannot be read, report which file and continue with the rest.

## Boundaries

One-shot, pre-PR. Invoked as the opening pass of the `/rc-review` command (and the review phase of `/rc-pipe`), before the quality round loop, so the rounds review already-lean code. Opt-in at the command level — it runs only when the user invokes that review. Pairs with `rc-execute-task` (which already climbs the laziness ladder at write time) as a safety net for drift and for legacy code that never went through that ladder. Lists findings and applies nothing — the caller applies the cuts.
