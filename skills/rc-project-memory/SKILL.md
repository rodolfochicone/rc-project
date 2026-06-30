---
name: rc-project-memory
description: Reads and writes the per-project curated memory database (.rc/memory.db) through the `rc memory` command. Use to search relevant project memories before working and to record durable decisions, conventions, and gotchas afterward. Do not use for workflow/task-scoped memory (use rc-workflow-memory), for PR review remediation, or for global user preferences.
user-invocable: true
model: sonnet
effort: medium
---

# Project Memory

Consult and curate the per-project memory stored in `.rc/memory.db`. This memory is a
small set of durable, curated facts about *this project* — decisions, conventions,
gotchas, glossary, and context — that survive across workflows and runs. It is separate
from `rc-workflow-memory`, which holds workflow/task-scoped notes under `.rc/tasks/`.

Access is always through the `rc memory` command. The database is the single source of
truth; there is no markdown mirror. Retrieval is keyword-ranked (SQLite FTS5/BM25), so
exact symbol and identifier matches are preserved.

## Command Reference

- `rc memory search "<terms>" [--scope <s>] [--limit N] [--format json]` — ranked lookup.
- `rc memory add --scope <s> --title "<t>" --body "<b>" [--key <k>] [--tags a,b] [--source <origin>]` — insert or upsert.
- `rc memory get <id>` or `rc memory get --scope <s> --key <k>` — fetch one record.
- `rc memory list [--scope <s>] [--tag <t>] [--limit N] [--format json]` — newest first.
- `rc memory update <id> [--title ...] [--body ...] [--tags ...]` — change fields.
- `rc memory delete <id>` — remove one record.
- `rc memory reindex` — regenerate semantic embeddings (only with embeddings enabled).
- `rc memory export` — write every memory to the committed text mirror under `.rc/memory/` (one markdown file per fact).
- `rc memory import` — load the mirror files back into the database (most-recent-wins by `updated_at`).

Use `--format json` whenever another step must parse the result. Hits and records expose
`id`, `scope`, `key`, `title`, `body`, `tags`, `source`, timestamps, and a `score` on search.

## Suggested scopes

Use one consistent value per fact. Common values: `decision`, `convention`, `gotcha`,
`glossary`, `context`. The store does not enforce an enum, so reuse existing values rather
than inventing near-duplicates.

## When to consult

- Before implementing, deciding, or planning, run `rc memory search` with the key terms
  of the task (feature name, package, symbol, error code) to recover prior decisions,
  conventions, and known gotchas.
- Prefer a scoped search when you know the kind of fact you need (for example
  `--scope convention`).

## When to record

Record a fact only when all three hold:

1. A future run would need it to avoid a mistake or rediscovery.
2. It is durable for this project, not specific to one task's execution.
3. It is NOT already obvious from the repository, git history, PRD, or techspec.

Good captures: a cross-cutting decision and its rationale, a project convention not stated
in config, a non-obvious gotcha and its workaround, a domain term and its meaning.

Use a stable `--key` for a fact that will be refreshed over time (for example
`--scope convention --key db-driver`); re-adding with the same scope and key updates the
existing record instead of creating a duplicate. Set `--source` to the skill or command
that produced the memory.

## Curation rules

- Keep each memory short and factual: a clear title and a few sentences of body.
- Do not paste large code blocks, stack traces, full specs, or raw session logs — that is
  how a memory store degrades into noise and contradictions.
- Do not record secrets, tokens, or credentials.
- When a fact changes, update or supersede the existing record (by `id`, or by re-adding
  with the same scope and key) instead of adding a contradicting second copy.
- Treat memory as curated, not automatic: write deliberately chosen facts, never a dump of
  everything that happened in a session.

## Semantic search (optional)

By default, search is purely lexical (BM25 with prefix and substring matching) — fast,
local, and dependency-free. It matches words and identifiers, but not synonyms with no
shared characters (e.g. "telemetry" vs "logging").

To add semantic ranking, run a local Ollama daemon and enable embeddings:

1. Install Ollama and pull a multilingual model: `ollama pull embeddinggemma`.
2. Set environment variables before running `rc`:
   - `RC_MEMORY_EMBEDDINGS=ollama` (enables it)
   - `RC_MEMORY_MODEL=embeddinggemma` (optional; this is the default)
   - `RC_MEMORY_ENDPOINT=http://localhost:11434` (optional; default local Ollama)
3. Run `rc memory reindex` once so existing memories get embeddings.

When enabled, `add` and `update` embed memories automatically (best-effort: if Ollama is
down the write still succeeds, just without a vector), and `search` fuses lexical and
semantic rankings. If embedding the query fails, search falls back to lexical results.
This keeps data on your machine (no API key, no external network) at the cost of running
the Ollama daemon — it is opt-in and off by default.

## Sharing across machines

The SQLite database (`.rc/memory.db`) is gitignored and local to one machine. To share
memory across machines and teammates, use the committed text mirror under `.rc/memory/`:

1. After recording memories, run `rc memory export` to write the mirror files.
2. Commit `.rc/memory/` to git and push.
3. On another machine, after pulling, run `rc memory import` to load them into the local
   database. Import is most-recent-wins by `updated_at`, so it is safe to run repeatedly.

Keyed facts (`--key`) map to a stable file name (`<scope>__<key>.md`), so the same fact edited
on two machines merges in place through git. Deletions do not propagate automatically: to
remove a shared fact, delete its `.md` file (and run `rc memory delete` locally).

## Error Handling

- A missing `id` or `(scope, key)` returns a not-found error; verify the identifier with
  `rc memory list` before retrying.
- If a memory conflicts with the repository or task specification, trust the repository and
  correct or delete the stale memory.
- `add` requires non-empty scope, title, and body; supply all three.
