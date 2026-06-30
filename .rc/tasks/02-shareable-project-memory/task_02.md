---
status: completed
title: "Store.Import — transactional most-recent-wins batch upsert"
type: backend
complexity: medium
dependencies: []
---

# Task 2: Store.Import — transactional most-recent-wins batch upsert

## Overview
Add a store method that imports fully-specified memory records, preserving their `id`,
`created_at`, and `updated_at` and applying most-recent-wins by `updated_at`. This is distinct
from `Add`, which always stamps `now()` and overwrites unconditionally — import must compare
timestamps and keep the newer version so local and remote edits both survive.

<critical>
- ALWAYS READ the TechSpec ("Core Interfaces", "Data Models") and ADR-002 before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST add `Import(ctx context.Context, records []Memory) (ImportResult, error)` and the `ImportResult{Added, Updated, Skipped int}` type to package `projectmemory`.
- MUST identify each record by `(scope, key)` when key is set, otherwise by `id` (contract `memory-mirror.import-preserves-identity`).
- MUST update an existing row only when the incoming `updated_at` is strictly newer; otherwise count it Skipped (contract `memory-mirror.import-most-recent-wins`).
- MUST store the incoming `id`, `created_at`, and `updated_at` verbatim — never re-stamp with `now()`.
- MUST NOT delete any row, even when records are absent (contract `memory-mirror.import-never-deletes`).
- MUST run the whole batch in a single transaction, rolling back on error, mirroring the migration transaction pattern.
- MUST keep the FTS5 indexes consistent (writes go through the `memories` table so existing triggers maintain `memories_fts`/`memories_trigram`).
</requirements>

## Subtasks
- [x] 2.1 Define `ImportResult` and the `Import` method signature in the store.
- [x] 2.2 Resolve each record's existing row by `(scope,key)` or `id` within the transaction.
- [x] 2.3 Insert when absent; update in place when strictly newer; skip otherwise — accumulating counts.
- [x] 2.4 Preserve incoming id/timestamps on insert and update (no `now()` stamping).
- [x] 2.5 Add table-driven unit tests with a `t.TempDir()` DB and `WithClock`.

## Implementation Details
Add the method to `internal/core/projectmemory/store.go` (or a sibling file in the package).
Follow the transaction-with-rollback pattern from `applyStep` in `schema.go`. Reuse the
existing column list (`memoryColumns`) and `scanMemoryRow`/`GetByKey`/`Get` for lookups.
Insert should write all nine columns directly (including id/created_at/updated_at) rather than
calling `Add` (which stamps `now()`). See TechSpec "Core Interfaces" for the signature.

### Relevant Files
- `internal/core/projectmemory/store.go` — `Add` (lines 144-188) to contrast; `GetByKey`/`Get`, `memoryColumns`, `normalizeTags`, `scanMemoryRow` to reuse.
- `internal/core/projectmemory/schema.go` — `applyStep` (transaction begin/commit/rollback) pattern to mirror.

### Dependent Files
- `internal/cli/memory.go` — task_04's `import` subcommand calls `Store.Import`.

### Patterns to Mirror
```go
// SOURCE: internal/core/projectmemory/schema.go:174-200
tx, err := db.BeginTx(ctx, nil)
if err != nil {
	return fmt.Errorf("projectmemory: begin migration %d: %w", step.version, err)
}
committed := false
defer func() {
	if !committed {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			retErr = errors.Join(retErr, fmt.Errorf("...rollback...: %w", rollbackErr))
		}
	}
}()
```

```go
// SOURCE: internal/core/projectmemory/store.go:144-154
func (s *Store) Add(ctx context.Context, in AddInput) (Memory, error) {
	scope := strings.TrimSpace(in.Scope)
	...
	stamp := store.FormatTimestamp(s.now())   // import must NOT do this — preserve incoming stamps
}
```

### Related ADRs
- [ADR-002: Mirror identity and most-recent-wins conflict resolution](adrs/adr-002.md) — defines identity and the strictly-newer rule.

## Deliverables
- `Import` method and `ImportResult` type in package `projectmemory`.
- Unit tests with 80%+ coverage **(REQUIRED)**.
- Tests proving most-recent-wins, identity resolution, and never-deletes **(REQUIRED)**.

## Tests
- Unit tests:
  - [ ] Absent record is inserted with its id/created_at/updated_at preserved (Added=1).
  - [ ] Keyed record with strictly-newer `updated_at` updates the existing row (Updated=1) and keeps the existing id.
  - [ ] Keyed record with equal `updated_at` is skipped (Skipped=1), content unchanged.
  - [ ] Keyed record with older `updated_at` is skipped, content unchanged.
  - [ ] Keyless record is matched by id; a second import of the same id-newer updates it.
  - [ ] Importing records does not remove any pre-existing row absent from the batch.
  - [ ] A failure mid-batch rolls back the whole transaction (no partial writes).
  - [ ] After import, `Search` finds an imported record (FTS index stayed consistent).
- Integration tests:
  - [ ] Covered end-to-end in task_04 (CLI round-trip).
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- Import is idempotent: re-running with the same files yields all Skipped
- No row is ever deleted by import
