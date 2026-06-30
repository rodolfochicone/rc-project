---
provider: manual
pr:
round: 1
round_created_at: 2026-06-30T17:20:00Z
status: resolved
file: internal/core/projectmemory/mirror.go
line: 110
severity: high
author: claude-code
provider_ref:
---

# Issue 001: Lossy key sanitization lets two memories collide to one mirror file

## Review Comment

`MirrorFileName` derives a keyed memory's file name from `sanitizeMirrorSegment(scope) + "__" +
sanitizeMirrorSegment(key) + ".md"`, and `sanitizeMirrorSegment` lowercases and collapses every
non-alphanumeric run to `-` (`internal/core/projectmemory/mirror.go:118-122`). The database
uniqueness constraint is on the **raw** `(scope, mem_key)` pair
(`schema.go` `uq_memories_scope_key`), so two genuinely distinct memories whose keys differ only
by case or punctuation are separate rows but map to the **same** file name. Examples within one
scope: `db-driver` and `DB Driver`; `db-driver` and `db_driver`.

Consequence: `rc memory export` loops over all records and writes each to its file path
(`memory_sync.go` export loop). When two records collide, the second write silently overwrites
the first, so the mirror ends up with **fewer files than records** — violating the behavioral
contract `memory-mirror.export-one-file-per-record`. On a fresh clone that only runs
`rc memory import`, the overwritten memory never arrives: a curated fact is silently lost,
which is exactly the data this feature exists to preserve and share.

Suggested fix: make collisions impossible or loud. Either (a) detect a name collision during
`export` (track written file names; if a second record maps to an already-written name, fail
with a clear error naming both ids/keys instead of overwriting), or (b) append a short
disambiguator derived from the raw key (e.g. a hash suffix) when the sanitized name is not
unique. Option (a) is simplest and preserves the "one file per record" contract by refusing to
silently drop data. Add a test with two same-scope keys that sanitize identically.

## Triage

- Decision: `VALID`
- Notes: Confirmed real. Root cause: `MirrorFileName` derives the file name from a lossy
  sanitization of `(scope, key)` while the DB uniqueness is on the raw pair, so two distinct
  rows can collide to one file and `export` silently overwrites one. Fix applied in
  `internal/cli/memory_sync.go`: the export loop now tracks written file names and returns a
  clear error (exit code 1) naming both record ids when two map to the same file, instead of
  overwriting — preserving the `export-one-file-per-record` contract by refusing to drop data.
  Regression test `TestMemoryExportFailsOnFileNameCollision` added (keys `db-driver` vs
  `DB Driver`). `make verify` passes.
