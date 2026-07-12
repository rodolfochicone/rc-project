---
name: rc-tasks-workflow
description: Executes the tasks of a RC feature slug by driving the Claude Code Workflow tool — one subagent per task, in dependency order, each implementing and verifying its task and returning structured test evidence. Use when the user wants to run a slug's task_NN.md files as a Claude-orchestrated workflow on a Claude Code host. Do not use to author tasks (use rc-create-tasks), to execute a single task in isolation (use rc-execute-task), or on non–Claude Code hosts (run each task via rc-execute-task in dependency order).
argument-hint: "[slug-or-linear-key]"
model: sonnet
effort: medium
---

# Execute RC tasks as a Claude Workflow

Run every `task_NN.md` of a feature slug through the Claude Code **`Workflow`** tool — each task implemented, verified, and tracked, orchestrated by Claude one task at a time. **Invoking this skill is the explicit opt-in to call the `Workflow` tool.**

This is **Claude Code only** — the `Workflow` tool does not exist in other agents (Codex, Droid, Cursor). On those hosts, run each task through the `rc-execute-task` skill in dependency order instead.

## When to stop and defer

- **No `Workflow` tool available** (non–Claude Code host) → stop and tell the user to run each task via the `rc-execute-task` skill in dependency order. Do not emulate the workflow by hand.
- **Tasks not yet authored** (`.rc/tasks/<slug>/` missing, or no `task_NN.md`) → stop; point at `/rc-create-tasks`.

## Phase 1 — Resolve the slug

1. If the argument is a **slug**, use `.rc/tasks/<slug>/` directly.
2. If it is a **Linear key** (the rc-linear gancho), grep `linear_key:` across `.rc/tasks/**/task_*.md`; the matching files share one directory — that directory name is the slug. No local match → stop and say so (local task files are the source of truth; never reconstruct tasks from Linear text).
3. Confirm `.rc/tasks/<slug>/` holds `_tasks.md` and `task_01.md … task_NN.md`.

## Phase 2 — Validate & order

1. Run the bundled task validator: `node "$CLAUDE_PLUGIN_ROOT/scripts/validate-tasks.mjs" --slug <slug>`; if it exits non-zero, stop and report (never run a broken set).
2. Read `_tasks.md` (the `#` / `Dependencies` table) and each task file's `dependencies` frontmatter. Build a **topological order**. Skip any task already `status: completed`.
3. Show the user the ordered plan (task #, title, deps) and confirm before launching — the workflow edits the working tree.

## Phase 3 — Build and launch the Workflow

Call the `Workflow` tool with a script that runs the tasks **sequentially in dependency order** — one `await agent()` at a time, no parallelism, so tasks never fight over the working tree. Each `agent()`:

- Receives the `task_NN.md` content, the PRD directory `.rc/tasks/<slug>/`, the tracking paths (`_tasks.md` + its own task file), and the auto-commit mode.
- Follows the **`rc-execute-task` contract**: explore → implement the task → run `make verify` (via `rc-final-verify`) → update tracking (task checkboxes → task status → `_tasks.md` → commit if auto-commit is on).
- Returns structured evidence via `schema`.

A task that fails verification marks itself `failed`; skip any task whose `deps` include a failed task and record why. Independent tasks still run.

Pass the ordered task list and slug via the Workflow tool's `args` as real JSON (not a stringified list). Reference script shape — adapt the task list to the real slug:

```js
export const meta = {
  name: 'rc-tasks-workflow',
  description: 'Execute RC feature tasks sequentially with verification',
  phases: [{ title: 'Execute' }],
}

const EVIDENCE = {
  type: 'object',
  properties: {
    task: { type: 'string' },
    status: { type: 'string', enum: ['passed', 'failed', 'skipped'] },
    gate: { type: 'string' },
    evidence: { type: 'string' },
    files: { type: 'array', items: { type: 'string' } },
  },
  required: ['task', 'status', 'gate', 'evidence'],
}

// args = { slug, autoCommit, tasks: [{ id, file, deps }, ...] } in topological order
const failed = new Set()
const results = []
for (const t of args.tasks) {
  const blockers = t.deps.filter(d => failed.has(d))
  if (blockers.length) {
    results.push({ task: t.id, status: 'skipped', gate: '—', evidence: `blocked by ${blockers.join(', ')}` })
    failed.add(t.id)
    continue
  }
  const r = await agent(
    `Execute RC task ${t.id} following the rc-execute-task contract.\n` +
    `Task file: ${t.file}\nPRD dir: .rc/tasks/${args.slug}/\n` +
    `Tracking: .rc/tasks/${args.slug}/_tasks.md and the task file.\n` +
    `Auto-commit: ${args.autoCommit}.\n` +
    `Implement the task, run make verify, update tracking. ` +
    `Return status "passed" only if make verify is clean; otherwise "failed" with the failing output as evidence.`,
    { label: `task:${t.id}`, phase: 'Execute', schema: EVIDENCE }
  )
  results.push(r)
  if (r?.status !== 'passed') failed.add(t.id)
}
return { results }
```

## Phase 4 — Report

The `Workflow` tool runs in the background and notifies on completion; read its result and relay what matters. Summarize per task — status, gate result, a trimmed evidence excerpt, files touched — and the overall outcome (N passed / M failed / K skipped).

When this skill was reached from the **rc-linear *Execute an issue*** flow, hand the per-task evidence back to rc-linear so it can post it to Linear (comment per sub-issue + parent summary). This skill itself does **not** write to Linear.

## Guardrails

- **Claude-only, opt-in via this skill.** Calling `Workflow` is justified only because the user invoked this skill. Never call it for unrelated work.
- **Sequential, no parallelism.** Tasks edit the shared working tree; run one at a time in dependency order. Worktree-isolated parallelism is explicitly out of scope.
- **Local task files are the source of truth.** Resolve the slug from local files; never reconstruct tasks from Linear text.
- **Verify gates completion.** A task is `passed` only after a clean `make verify`. Never mark a failing task done.
- **Claude Code only.** On other hosts, run each task through the `rc-execute-task` skill in dependency order; this Workflow path is the Claude-native runner.
