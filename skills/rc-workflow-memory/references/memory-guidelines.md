# Workflow Memory Guidelines

Use these rules to keep RC workflow memory useful across repeated PRD task runs.

## File Roles

### Shared workflow memory: `MEMORY.md`

Use the shared workflow memory for context that should survive across multiple tasks and multiple runs.

Keep:
- current workflow state that affects more than one task
- durable technical or product decisions
- reusable learnings that will matter again
- open risks or handoff notes that change future execution

Avoid:
- step-by-step scratch notes
- large code excerpts
- facts that are already explicit in `_prd.md`, `_techspec.md`, `_tasks.md`, or the repo itself

### Current task memory: `memory/<task filename>`

Use the task memory for context that is specific to the current task.

Keep:
- the current objective snapshot
- important task-local decisions
- local learnings and corrections
- touched files or surfaces worth remembering next run
- ready-for-next-run notes

Avoid:
- cross-task summaries that belong in `MEMORY.md`
- repeated restatements of the task spec
- low-signal command transcripts

## Promotion Rules

Promote an item from task memory into shared workflow memory only when it is:
- durable across runs
- useful to another task
- likely to prevent repeated mistakes or rediscovery

Leave information in task memory when it is:
- operational only for the current task
- temporary
- too detailed for workflow-wide reuse

## Compaction Rules

When compaction is required:
- preserve current state, durable decisions, reusable learnings, open risks, and handoffs
- remove repetition, stale notes, long transcripts, and derivable facts
- rewrite for clarity, not for completeness
- prefer short factual bullets over narrative logs

## Default Section Boundaries

### `MEMORY.md`

- `## Current State`
- `## Shared Decisions`
- `## Shared Learnings`
- `## Open Risks`
- `## Handoffs`

### `memory/<task filename>`

- `## Objective Snapshot`
- `## Important Decisions`
- `## Learnings`
- `## Files / Surfaces`
- `## Errors / Corrections`
- `## Ready for Next Run`

## Decision and handoff entry format

When a loop (`rc-loop`) or a card flow (`rc-card`) records a durable decision or closes a
phase/sub-issue, use these shapes so the record stays scannable and resumable. These live in the
shared `MEMORY.md` sections above — they are a format, not a new file.

### `## Shared Decisions` — one entry per durable, cross-task decision (`AD-NNN`)

```markdown
### AD-007
- **Decision**: <what was decided, one line>
- **Reason**: <why — the forcing constraint>
- **Trade-off**: <what it costs / defers>
- **Scope**: <which phases/tasks it binds>
- **Status**: active | superseded by AD-NNN
```

Number sequentially (`AD-001`, `AD-002`, …); never renumber. Supersede, don't delete — flip the
old entry's status to `superseded by AD-NNN` so the history survives.

### `## Handoffs` — one entry per completed phase / sub-issue, newest first

```markdown
**<Phase or sub-issue> — COMPLETE (verification PASS, <date>).**
<AC/gate result>; <gate evidence: command + pass count>; <what the next phase needs to know>.
```

A handoff exists so a re-invoked loop resumes without re-deriving context. Record the gate
evidence (the command run and its result), any blocker the phase hit and how it was resolved, and
the one or two facts the next phase must start from. Keep it to a few lines — trim it in the next
compaction if it grows.
