---
status: completed
title: "`rc memory export` subcommand (DB → mirror files)"
type: backend
complexity: low
dependencies:
  - task_01
---

# Task 3: `rc memory export` subcommand (DB → mirror files)

## Overview
Add the `rc memory export` subcommand that writes every memory in the database to one markdown
file per fact under `.rc/memory/`, using the serializer and filename rule from task_01. This is
the DB→files half of the shareable-memory sync.

<critical>
- ALWAYS READ the TechSpec ("System Architecture", "API Endpoints") and ADR-001 before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST add `newMemoryExportCommand()` and wire it into `newMemoryCommand`'s `AddCommand` list.
- MUST write exactly one `.md` file per record returned by `Store.List` with an unbounded limit (contract `memory-mirror.export-one-file-per-record`).
- MUST resolve the `.rc` base via the existing `openProjectMemory`/workspace discovery and write under `<base>/memory/`, creating the directory if absent.
- MUST use `MarshalMemory` and `MirrorFileName` from task_01; MUST NOT introduce a second serialization path.
- MUST NOT delete files and MUST NOT modify `.gitignore`; `.rc/memory.db` stays ignored, `.rc/memory/` is tracked (invariant `memory-mirror.db-stays-gitignored`).
- MUST print a count of files written and follow the existing exit-code convention (`withExitCode`/`mapMemoryError`).
</requirements>

## Subtasks
- [x] 3.1 Add `newMemoryExportCommand()`, mirroring sibling subcommands.
- [x] 3.2 List all records, marshal each, and write `<base>/memory/<MirrorFileName>` (0o600, dir 0o755).
- [x] 3.3 Wire the command into `newMemoryCommand`.
- [x] 3.4 Print the written count; map errors to exit codes.
- [x] 3.5 Add a test writing N memories and asserting N files with expected names/content.

## Implementation Details
Add the command to `internal/cli/memory.go`. Resolve the base directory the same way
`openProjectMemory` resolves the DB path — `filepath.Join(model.RcDir(root), "memory")`.
Read all records via `st.List(ctx, projectmemory.ListFilter{})` (Limit 0 = unbounded). See
TechSpec "System Architecture" for the data flow. Match the `state struct` + `RunE` shape used
by `newMemoryAddCommand`/`newMemoryListCommand`.

### Relevant Files
- `internal/cli/memory.go` — `newMemoryCommand` (lines 21-42), `openProjectMemory` (44-74), `newMemoryAddCommand` (237-285) to mirror.
- `internal/core/projectmemory/store.go` — `List` (lines 284-318) for reading all records.
- `internal/core/model` — `RcDir(root)` to resolve the `.rc` base.

### Dependent Files
- `internal/core/projectmemory/mirror.go` (task_01) — provides `MarshalMemory`/`MirrorFileName`.

### Patterns to Mirror
```go
// SOURCE: internal/cli/memory.go:258-271
func (s *memoryAddState) run(cmd *cobra.Command, _ []string) error {
	ctx, stop := signalCommandContext(cmd)
	defer stop()
	...
	st, _, err := openProjectMemory(ctx)
	if err != nil {
		return withExitCode(2, err)
	}
	defer closeProjectMemory(ctx, st)
```

```go
// SOURCE: internal/cli/memory.go:32-40
cmd.AddCommand(
	newMemoryAddCommand(),
	newMemorySearchCommand(),
	...
)
```

### Related ADRs
- [ADR-001: Share project memory through a committed markdown-per-fact mirror](adrs/adr-001.md) — the directory and file format being written.

## Deliverables
- `newMemoryExportCommand` wired into the `memory` command tree.
- Unit/command tests with 80%+ coverage **(REQUIRED)**.
- Test asserting one file per record with correct names **(REQUIRED)**.

## Tests
- Unit tests:
  - [ ] Export of 3 memories (2 keyed, 1 keyless) writes exactly 3 files with expected names under `<tempdir>/.rc/memory/`.
  - [ ] A written file parses back (via `ParseMemory`) to the original record.
  - [ ] Export creates `.rc/memory/` when it does not exist.
  - [ ] Export of an empty database writes zero files and reports count 0 without error.
  - [ ] Re-export is idempotent: identical content, same filenames.
- Integration tests:
  - [ ] Round-trip with `import` is covered in task_04.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- One file per record, deterministically named
- `.gitignore` is untouched
