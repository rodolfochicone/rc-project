---
name: rc-project-memory
description: Reads and writes the per-project curated memory stored as markdown files under .rc/memory/. Use to search relevant project memories before working and to record durable decisions, conventions, and gotchas afterward. Do not use for workflow/task-scoped memory (use rc-workflow-memory), for PR review remediation, or for global user preferences.
user-invocable: true
model: sonnet
effort: medium
---

# Project Memory

Consult and curate the per-project memory stored as markdown files under `.rc/memory/`.
This memory is a small set of durable, curated facts about *this project* — decisions,
conventions, gotchas, glossary, and context — that survive across workflows and runs. It is
separate from `rc-workflow-memory`, which holds workflow/task-scoped notes under `.rc/tasks/`.

Each fact is one markdown file. The directory is committed to git, so memory is shared across
machines and teammates automatically — no database, no export/import step.

## File format

One file per fact at `.rc/memory/<scope>__<key>.md`:

```markdown
---
scope: convention
key: db-driver
title: Use pgx, not database/sql
tags: [db, postgres]
source: rc-execute-task
updated: 2026-07-07
---

We standardized on jackc/pgx for Postgres access. database/sql is only kept in the legacy
importer. New code must use the pgx pool from internal/db.
```

- `<scope>__<key>` maps a fact to a stable filename, so refreshing a fact edits the same file
  in place (and merges cleanly through git) instead of creating a duplicate.
- For a one-off fact without a natural key, use a short kebab-case slug as the key.

## Suggested scopes

Use one consistent value per fact. Common values: `decision`, `convention`, `gotcha`,
`glossary`, `context`. Reuse existing scope values rather than inventing near-duplicates.

## When to consult

- Before implementing, deciding, or planning, search `.rc/memory/` for the key terms of the
  task (feature name, package, symbol, error code) with Grep, and read the matching files, to
  recover prior decisions, conventions, and known gotchas.
- When you know the kind of fact you need, narrow by filename prefix (for example
  `.rc/memory/convention__*.md`).

## When to record

Record a fact only when all three hold:

1. A future run would need it to avoid a mistake or rediscovery.
2. It is durable for this project, not specific to one task's execution.
3. It is NOT already obvious from the repository, git history, PRD, or techspec.

Good captures: a cross-cutting decision and its rationale, a project convention not stated in
config, a non-obvious gotcha and its workaround, a domain term and its meaning.

To record, write a new `.rc/memory/<scope>__<key>.md` file, or edit the existing file when the
same `(scope, key)` already exists. Set `source` to the skill that produced the memory and
`updated` to today's date.

## Curation rules

- Keep each memory short and factual: a clear title and a few sentences of body.
- Do not paste large code blocks, stack traces, full specs, or raw session logs — that is how a
  memory store degrades into noise and contradictions.
- Do not record secrets, tokens, or credentials.
- When a fact changes, update or supersede the existing file instead of adding a contradicting
  second copy. To remove a shared fact, delete its `.md` file.
- Treat memory as curated, not automatic: write deliberately chosen facts, never a dump of
  everything that happened in a session.

## Error Handling

- If a fact you expected is missing, list `.rc/memory/` to confirm the filename before assuming
  it is absent.
- If a memory conflicts with the repository or task specification, trust the repository and
  correct or delete the stale memory file.
