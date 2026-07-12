---
name: rc-memory
description: The single place for RC's durable, cross-session memory and learning. Records and recalls curated project facts (decisions, conventions, gotchas, glossary) AND distills recurring corrections/workflows into confidence-scored learnings — all as plain markdown under .rc/memory/. Use to consult memory before working and to record durable facts or distilled learnings afterward. Do not use for workflow/task-scoped notes (use rc-workflow-memory) or for on-demand reflection on a diff (use lesson-learned).
user-invocable: true
model: sonnet
effort: medium
---

# Memory

The one place for what RC should remember about *this project* across turns and sessions. It holds
two kinds of durable knowledge, both as plain markdown under `.rc/memory/` — no database, no binary:

1. **Facts** — curated decisions, conventions, gotchas, glossary, and context.
2. **Learnings** — recurring corrections/workflows distilled into confidence-scored `trigger → action`
   rules (the continuous-learning loop fed by the `observe` hook).

It is separate from `rc-workflow-memory` (ephemeral, per-workflow notes under `.rc/tasks/<slug>/`)
and from `lesson-learned` (on-demand reflection on a diff, not a store).

## Store layout

```
.rc/memory/
  INDEX.md              # one line per fact — the recall entry point
  <slug>.md             # one durable fact per file
  LEARNINGS.md          # distilled trigger→action learnings, confidence-scored
  observations.jsonl    # raw capture appended by the observe hook (transient)
```

Resolve the project's `.rc` base directory (nearest `.rc` walking up; default `./.rc`).

## A — Facts

Each fact is one `<slug>.md` with frontmatter + a short body:

```markdown
---
title: <short title>
scope: decision | convention | gotcha | glossary | context
key: <stable-kebab-slug or omit>
tags: [a, b]
source: <skill/command that produced it>
created: <YYYY-MM-DD>
updated: <YYYY-MM-DD>
---

<the fact in a few sentences. For a decision, add **Why:** and, if useful, **How to apply:**.>
```

`INDEX.md` holds one line per fact, newest last: `- [<title>](<slug>.md) — <scope> — <one-line hook>`.
It is the only file you read to decide relevance; never dump fact bodies into it.

**Consult** (before implementing/deciding/planning): read `INDEX.md`, scan the one-line hooks for the
task's terms, open only the matching `<slug>.md`. For a large store, `grep -ri "<term>" .rc/memory/`.

**Record** a fact only when all three hold: (1) a future run would need it to avoid a mistake or
rediscovery; (2) it is durable for this project, not one task's execution; (3) it is NOT already
obvious from the repo, git history, PRD, or techspec. Reuse the same slug/`key` to update instead of
duplicating. Keep each fact short; never store secrets, large code blocks, or logs.

## B — Learnings (the continuous-learning loop)

When the same correction, error resolution, or workflow recurs, stop re-learning it each session and
make it a **learning**: one `trigger → action` rule with a confidence score and evidence. Capture is
the `observe` hook (on by default; `RC_INSTINCTS=0` to opt out), appending `{ts, tool, target}` lines
to `.rc/memory/observations.jsonl`. This skill is the distill/curate side, run deliberately. The
`memory-load` SessionStart hook surfaces existing facts/learnings at the start of each session and
nudges you to run this skill once observations accumulate (≥40 lines).

`LEARNINGS.md` groups atomic learnings by domain; each is one line:

```
## code-style
- [0.7] when adding a new error path → wrap with the project's error-wrap helper  (evidence: corrected 3×; updated 2026-07-07; seen 4)
```

- **Confidence** `0.3–0.9`; start fresh at `0.4–0.5`. One trigger → one action (split if you need "and").
- **Evidence** is why it exists — no evidence, not a learning.

**Distill:** (1) read `observations.jsonl` (if present) + the current session — corrections, resolved
errors, sequences repeated 3+×, stated preferences. (2) Cluster into atomic `trigger → action`
candidates; discard one-off or already-obvious ones. (3) Merge with `LEARNINGS.md`: new → add at
`0.4–0.5` with `seen 1`; recurred uncontradicted → raise confidence (+0.1–0.2, cap `0.9`), bump `seen`,
refresh date; contradicted this session → lower (−0.2–0.3), drop below `0.3`; stale + low → drop.
(4) Write `LEARNINGS.md` sorted by domain then confidence, then **truncate** `observations.jsonl` so the
next pass starts clean. (5) Report what was added/reinforced/weakened/dropped.

## Critical rules

- Project-scoped only. Never copy memory or learnings to another project or a global location.
- Facts and learnings are **curated, not a log** — deliberately chosen, pruned aggressively.
- Never store secrets, tokens, credentials, large code blocks, stack traces, or raw session logs.
- When a fact or learning is contradicted by the repo or the user, correct or delete it — trust the
  repo over stale memory.
- This skill writes only under `.rc/memory/`; it does not edit source code.
