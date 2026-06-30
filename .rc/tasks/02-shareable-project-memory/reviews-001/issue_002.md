---
provider: manual
pr:
round: 1
round_created_at: 2026-06-30T17:20:00Z
status: resolved
file: internal/core/projectmemory/import.go
line: 90
severity: low
author: claude-code
provider_ref:
---

# Issue 002: Import accepts zero timestamps, writing an unreadable row

## Review Comment

`importOne` validates that `id`, `scope`, `title`, and `body` are non-empty
(`internal/core/projectmemory/import.go:87-91`) but does not validate the timestamps. It writes
`store.FormatTimestamp(record.CreatedAt)` / `...UpdatedAt` directly. `FormatTimestamp` returns an
**empty string** for a zero `time.Time` (`internal/store/values.go:13-18`). A row inserted with
`created_at=""` / `updated_at=""` then fails to read back: `scanMemoryRow` → `hydrateMemory` calls
`store.ParseTimestamp("")`, which errors, so a later `Get`/`List`/`Search` over that row returns
an error. The import reports success while leaving the database in a state where reads fail.

This is not reachable through the `rc memory import` command today, because `ParseMemory` rejects
missing/unparseable timestamps before records reach `Import` (`mirror.go:75-83`). But `Import` is
an **exported** method on `Store`, and the most-recent-wins comparison also relies on a real
`UpdatedAt` (a zero time would always lose). The validation gap is a latent trap for any future
caller and is silent when it triggers.

Suggested fix: extend the existing required-field guard in `importOne` to also reject a zero
`CreatedAt` or `UpdatedAt` with `ErrInvalidInput`, symmetric to the id/scope/title/body check.
Add a unit test asserting `Store.Import` returns `ErrInvalidInput` for a record with a zero
timestamp.

## Triage

- Decision: `VALID`
- Notes: Confirmed real for the exported `Store.Import` API. A zero `time.Time` formats to `""`
  and produces a row that fails `ParseTimestamp` on read. Not reachable via the CLI (ParseMemory
  validates first), but a latent silent trap for any caller. Fix applied in
  `internal/core/projectmemory/import.go`: `importOne` now rejects a zero `CreatedAt`/`UpdatedAt`
  with `ErrInvalidInput`, symmetric to the existing id/scope/title/body guard. Regression test
  `TestImport_RejectsZeroTimestamps` added. `make verify` passes.
