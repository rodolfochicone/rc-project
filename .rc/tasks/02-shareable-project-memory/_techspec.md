# TechSpec: Shareable Project Memory

> No `_prd.md` exists for this feature. This spec is grounded in the analysis report
> `.rc/analysis/memoria-projeto-funcionamento-001.md` and the design decisions recorded in
> project memory (`decision/project-memory-sharing`) and ADR-001/ADR-002.

## Executive Summary

The per-project memory store keeps curated facts in a binary SQLite database
(`.rc/memory.db`) that is gitignored and therefore local to one machine. This feature makes
project memory shareable across machines and teammates by adding a committed **text mirror**:
one markdown-per-fact file under `.rc/memory/`, synced explicitly with two new subcommands —
`rc memory export` (DB → files) and `rc memory import` (files → DB). The database stays the
runtime query engine (FTS5/BM25) and remains gitignored; the committed mirror becomes the
durable, shareable source of truth.

The primary technical trade-off: we accept added code (a markdown writer, a YAML-frontmatter
parser, a timestamp-guarded batch upsert, and two subcommands) and a shift of "source of
truth" from the database to the committed files, in exchange for per-fact git-merge
granularity and human-readable diffs — rejecting both the binary-DB commit (unmergeable) and
the single-JSONL file (conflicts concentrated in one file).

## System Architecture

### Component Overview

- **`projectmemory` mirror (new file in `internal/core/projectmemory`)** — pure functions to
  serialize a `Memory` to markdown-with-frontmatter and parse it back, plus a deterministic
  filename rule. Owns the on-disk text format.
- **`projectmemory.Store` (extended)** — gains `Import`, a transactional batch upsert that
  preserves `id`/`created_at`/`updated_at` and applies most-recent-wins. Export reuses the
  existing `Store.List` (unbounded) to read every record.
- **`rc memory export` / `rc memory import` (new subcommands in `internal/cli/memory.go`)** —
  resolve the workspace and `.rc` base via the existing `openProjectMemory`, then write or
  read the `.rc/memory/` directory.

Data flow:
`rc memory export` → `Store.List(all)` → serialize each → write `.rc/memory/<file>.md`.
`rc memory import` → read `.rc/memory/*.md` → parse each → `Store.Import(records)` →
most-recent-wins upsert into `memory.db`.

## Implementation Design

### Core Interfaces

```go
// Import upserts fully-specified records (preserving id, created_at, updated_at) inside one
// transaction, applying most-recent-wins by UpdatedAt. Identity is (Scope, Key) when Key is
// set, otherwise ID. It never deletes rows.
func (s *Store) Import(ctx context.Context, records []Memory) (ImportResult, error)

// ImportResult reports the outcome of a batch import.
type ImportResult struct {
	Added   int
	Updated int
	Skipped int // already present and not strictly newer
}

// MarshalMemory renders a memory as markdown with YAML frontmatter (id, scope, key, title,
// tags, source, created_at, updated_at) followed by the body.
func MarshalMemory(m Memory) ([]byte, error)

// ParseMemory parses a mirror file's bytes back into a Memory. It errors on missing required
// fields (id, scope, title, body, created_at, updated_at).
func ParseMemory(data []byte) (Memory, error)

// MirrorFileName is the deterministic file name for a memory: "<scope>__<sanitized-key>.md"
// when Key is set, otherwise "<id>.md".
func MirrorFileName(m Memory) string
```

### Data Models

No schema change. The existing `Memory` (`internal/core/projectmemory/store.go:38`) is the
unit: `ID, Scope, Key, Title, Body, Tags[], Source, CreatedAt, UpdatedAt`.

Mirror file shape (`.rc/memory/<name>.md`):

```markdown
---
id: mem-335245db2db2e8e6
scope: decision
key: project-memory-sharing
title: Memoria de projeto deve ser compartilhada via mirror em texto
tags: [architecture, memory, sharing]
source: rc-analyze
created_at: 2026-06-30T16:24:25Z
updated_at: 2026-06-30T16:24:25Z
---

<body text>
```

Frontmatter is read/written with `gopkg.in/yaml.v3`, consistent with
`internal/core/tasks/parser.go`. Timestamps use the store's UTC format
(`store.FormatTimestamp`/`ParseTimestamp`).

### API Endpoints

Not applicable — this is CLI surface, not HTTP. New subcommands:

- `rc memory export [--workspace <dir>]` — writes every DB record to `.rc/memory/`. Prints a
  count of files written. Creates `.rc/memory/` if absent.
- `rc memory import [--workspace <dir>]` — reads `.rc/memory/*.md` and upserts them. Prints
  `added/updated/skipped`. A malformed file is reported and skipped without aborting the run.

## Integration Points

None outside the codebase. Sync is local file I/O against the resolved `.rc` base directory.

## Impact Analysis

| Component | Impact Type | Description and Risk | Required Action |
|-----------|-------------|----------------------|-----------------|
| `internal/core/projectmemory` (new mirror file) | new | Serializer/parser/filename + `Import`. Low risk: additive, pure functions + one transactional method. | Implement with table-driven tests, including round-trip. |
| `internal/core/projectmemory/store.go` | modified | Add `Import` and `ImportResult`. Risk: timestamp-guarded upsert must not re-stamp `now()`. | Add method; do not change `Add`. |
| `internal/cli/memory.go` | modified | Two subcommands wired into `newMemoryCommand`. Low risk. | Add `newMemoryExportCommand`/`newMemoryImportCommand`. |
| `internal/core/projectmemory/store.go:1` (package doc) | modified | "no markdown mirror" tenet is superseded. | Update the doc comment to describe the mirror. |
| `skills/rc-project-memory/SKILL.md` | modified | Document export/import and the shared-mirror workflow. | Add command reference + workflow note. |
| `.rc/memory/` (new tracked dir) | new | Committed mirror. `.rc/memory.db` stays gitignored. | No `.gitignore` change. |

## Testing Approach

### Unit Tests

- **Round-trip**: `MarshalMemory` → `ParseMemory` returns an equal `Memory` (all fields,
  including tags order and timestamps). Table-driven, with `t.Parallel()`.
- **Filename**: keyed → `<scope>__<sanitized-key>.md`; keyless → `<id>.md`; sanitization is
  deterministic and idempotent; odd-character keys are covered.
- **Parse errors**: missing required field (id/scope/title/body/timestamps) returns an error
  naming the field; malformed frontmatter is rejected.
- **`Import` most-recent-wins**: strictly-newer `updated_at` updates; equal or older skips;
  absent inserts; identity by `(scope,key)` vs `id`. Use `WithClock` and `t.TempDir()` DBs.

### Integration Tests

- **CLI round-trip**: `rc memory add` × N → `export` → wipe DB → `import` → `list` matches the
  original set. Run against a `t.TempDir()` workspace.
- **Cross-"machine" merge**: export from DB A and DB B into the same dir, import into a third
  DB; assert keyed facts converge to one record (newest wins) and keyless facts coexist.
- **Malformed file**: a junk `.md` in the dir is skipped, the rest import, exit code reflects
  partial success per existing `mapMemoryError` conventions.

## Development Sequencing

### Build Order

1. **Mirror format** — `MarshalMemory`, `ParseMemory`, `MirrorFileName` + unit tests. No
   dependencies.
2. **`Store.Import` + `ImportResult`** — transactional most-recent-wins upsert + unit tests.
   Depends on step 1 only for the `Memory` shape it consumes (already exists); independent of
   the serializer.
3. **`rc memory export`** — wire `Store.List` + `MarshalMemory` + `MirrorFileName` into a
   subcommand. Depends on steps 1.
4. **`rc memory import`** — read dir + `ParseMemory` + `Store.Import` into a subcommand.
   Depends on steps 1 and 2.
5. **CLI integration tests** — export/import round-trip and merge. Depends on steps 3 and 4.
6. **Docs** — update the package doc comment and `skills/rc-project-memory/SKILL.md`. Depends
   on steps 3 and 4 (final command surface).

### Technical Dependencies

- `gopkg.in/yaml.v3` (already a direct dependency) for frontmatter.
- No infrastructure or external service dependencies.

## Monitoring and Observability

CLI tool, not a long-running service. `export`/`import` print human-readable counts; errors
follow the existing `mapMemoryError` exit-code convention (1 = user error, 2 = operational).
Use `slog` for any warning (e.g. a skipped malformed file), matching `closeProjectMemory`.

## Technical Considerations

### Key Decisions

- **Decision**: Committed markdown-per-fact mirror with explicit export/import.
  **Rationale**: per-fact git-merge granularity and readable diffs; mirrors the user's global
  memory pattern. **Trade-offs**: more code than a JSONL/binary commit; source-of-truth shifts
  to the mirror. **Alternatives rejected**: commit the binary `.db` (unmergeable); single
  JSONL (concentrated conflicts). See ADR-001.
- **Decision**: Identity by `(scope,key)` else `id`; most-recent-wins by `updated_at`;
  deletion propagation out of scope for v1. **Rationale**: reuses the store's existing
  identity, protects un-exported local edits, avoids tombstone machinery. See ADR-002.

### Known Risks

- **DB/mirror drift** if a user forgets to sync — mitigated by explicit, documented commands
  and most-recent-wins on import.
- **Deletions don't propagate** in v1 — a stale file re-imports a deleted fact; documented,
  and removed by deleting the file. Likelihood moderate; revisit with tombstones if it bites.
- **Clock skew** could let an older edit win on import — same risk git carries; UTC timestamps
  reduce it. Low likelihood.

## Behavioral Contract

### Requirement: export writes one file per memory
When `rc memory export` runs, it writes exactly one `.md` file under `.rc/memory/` for every
record returned by `Store.List` with an unbounded limit.
<!-- id: memory-mirror.export-one-file-per-record -->
<!-- enforced: pending -->

### Requirement: marshal/parse round-trip is lossless
For any valid memory, `ParseMemory(MarshalMemory(m))` equals `m` across all fields including
tag set and UTC timestamps.
<!-- id: memory-mirror.roundtrip-lossless -->
<!-- enforced: pending -->

### Requirement: keyed facts map to a stable filename
A memory with a non-empty `key` serializes to `<scope>__<sanitized-key>.md`, identical on any
machine for the same `(scope, key)`.
<!-- id: memory-mirror.keyed-stable-filename -->
<!-- enforced: pending -->
<!-- depends_on: memory-mirror.roundtrip-lossless -->

### Requirement: keyless facts map to id filename
A memory with an empty `key` serializes to `<id>.md`.
<!-- id: memory-mirror.keyless-id-filename -->
<!-- enforced: pending -->

### Requirement: import is most-recent-wins
On `rc memory import`, an incoming record updates an existing one only when its `updated_at`
is strictly newer; otherwise the existing record is kept.
<!-- id: memory-mirror.import-most-recent-wins -->
<!-- enforced: pending -->

### Requirement: import preserves identity and timestamps
Imported records are matched by `(scope, key)` when `key` is set and by `id` otherwise, and
their `id`, `created_at`, and `updated_at` are stored verbatim (not re-stamped).
<!-- id: memory-mirror.import-preserves-identity -->
<!-- enforced: pending -->
<!-- depends_on: memory-mirror.import-most-recent-wins -->

### Requirement: malformed mirror file is skipped, not fatal
A mirror file missing a required field is reported and skipped while the remaining valid files
still import.
<!-- id: memory-mirror.import-skips-malformed -->
<!-- enforced: pending -->

### Invariant: the SQLite database stays gitignored
`.rc/memory.db` remains listed in `.gitignore`; only `.rc/memory/` is tracked.
<!-- id: memory-mirror.db-stays-gitignored -->
<!-- enforced: pending -->

### Invariant: import never deletes
`rc memory import` adds or updates rows but never deletes a database row, even when a
previously seen file is absent.
<!-- id: memory-mirror.import-never-deletes -->
<!-- enforced: pending -->

## Architecture Decision Records

- [ADR-001: Share project memory through a committed markdown-per-fact mirror](adrs/adr-001.md) — commit a text mirror, not the binary DB or a single JSONL file.
- [ADR-002: Mirror identity and most-recent-wins conflict resolution](adrs/adr-002.md) — identity by `(scope,key)` else `id`; newest `updated_at` wins; no deletion propagation in v1.
