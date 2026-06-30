---
status: completed
title: "`rc memory import` subcommand (mirror files → DB) + integration tests"
type: backend
complexity: medium
dependencies:
  - task_01
  - task_02
  - task_03
---

# Task 4: `rc memory import` subcommand (mirror files → DB) + integration tests

## Overview
Add the `rc memory import` subcommand that reads every `.md` file under `.rc/memory/`, parses
it with task_01's parser, and upserts the records via task_02's `Store.Import` (most-recent-wins).
This task also carries the end-to-end integration tests for the whole sync, since import is the
half that closes the round-trip.

<critical>
- ALWAYS READ the TechSpec ("System Architecture", "Testing Approach") and ADR-002 before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST add `newMemoryImportCommand()` and wire it into `newMemoryCommand`'s `AddCommand` list.
- MUST read all `*.md` files under `<base>/memory/`, parse each with `ParseMemory`, and pass them to `Store.Import`.
- MUST report a malformed file and continue importing the rest — a single bad file MUST NOT abort the run (contract `memory-mirror.import-skips-malformed`).
- MUST print `added/updated/skipped` counts from `ImportResult` and follow the existing exit-code convention.
- MUST treat the import as additive/most-recent-wins only — never delete DB rows (relies on task_02's guarantee).
- MUST cover the export→import round-trip and a cross-machine merge end-to-end in tests.
</requirements>

## Subtasks
- [x] 4.1 Add `newMemoryImportCommand()`, mirroring sibling subcommands.
- [x] 4.2 Read `<base>/memory/*.md`, parse each, collecting valid records and per-file parse errors.
- [x] 4.3 Call `Store.Import` with the valid records; print counts and report skipped files.
- [x] 4.4 Wire the command into `newMemoryCommand`.
- [x] 4.5 Add integration tests: round-trip, cross-DB merge, and malformed-file skip.

## Implementation Details
Add the command to `internal/cli/memory.go`, resolving `<base>/memory/` exactly as task_03
does. Use `filepath.Glob` or `os.ReadDir` to list `.md` files. Accumulate parse failures and
log/report them (via `slog`, like `closeProjectMemory`) without failing the batch; return a
non-zero exit only if `Store.Import` itself fails. See TechSpec "Testing Approach" for the
integration scenarios. Reuse the add/list command shape for the state struct and `RunE`.

### Relevant Files
- `internal/cli/memory.go` — sibling subcommands and `openProjectMemory`; `mapMemoryError` (104-111), `closeProjectMemory` (96-100).
- `internal/core/projectmemory/store.go` — `Import` (task_02) and `Search`/`List` for assertions.
- `internal/core/projectmemory/mirror.go` (task_01) — `ParseMemory`.

### Dependent Files
- `.rc/memory/` — the directory read at runtime.

### Patterns to Mirror
```go
// SOURCE: internal/cli/memory.go:312-335
func (s *memorySearchState) run(cmd *cobra.Command, args []string) error {
	ctx, stop := signalCommandContext(cmd)
	defer stop()
	...
	st, retriever, err := openProjectMemory(ctx)
	if err != nil {
		return withExitCode(2, err)
	}
	defer closeProjectMemory(ctx, st)
```

```go
// SOURCE: internal/cli/memory.go:96-100
func closeProjectMemory(ctx context.Context, st *projectmemory.Store) {
	if err := st.Close(ctx); err != nil {
		slog.Default().Warn("close project memory", "error", err)   // report-and-continue style
	}
}
```

### Related ADRs
- [ADR-002: Mirror identity and most-recent-wins conflict resolution](adrs/adr-002.md) — the import semantics this command surfaces.

## Deliverables
- `newMemoryImportCommand` wired into the `memory` command tree.
- Unit/command tests with 80%+ coverage **(REQUIRED)**.
- Integration tests: export→import round-trip, cross-DB merge, malformed-file skip **(REQUIRED)**.

## Tests
- Unit tests:
  - [ ] A directory with 2 valid files and 1 file missing `title` imports the 2 and reports the 1 skipped, exit code reflects partial success per convention.
  - [ ] Empty/absent `.rc/memory/` imports nothing and reports `added=0 updated=0 skipped=0`.
- Integration tests:
  - [ ] Round-trip: `add` 3 memories → `export` → open a fresh empty DB → `import` → `list` matches the original set (ids, scopes, keys, bodies, timestamps).
  - [ ] Cross-machine merge: a keyed fact edited later in DB-B (newer `updated_at`) wins over DB-A's version after both export into one dir and a third DB imports; keyless facts from both coexist.
  - [ ] Re-import of unchanged files yields all skipped (idempotent).
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- Export→import round-trip is lossless
- A malformed file never aborts the import
