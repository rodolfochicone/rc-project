---
status: completed
title: Input coordinator model types and runtime config flags
type: backend
complexity: low
dependencies: []
---

# Task 1: Input coordinator model types and runtime config flags

## Overview
Define the shared, dependency-free types that the whole feature depends on: the
`InputCoordinator` interface and its request/response value types, plus the two
new `RuntimeConfig` fields that gate and carry interactivity. These live in
`internal/core/model` so the agent client, executor, and daemon can all depend on
them without import cycles.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST define `InputCoordinator`, `PendingInput`, `InputOption`, and `UserResponse`
  in `internal/core/model` exactly as specified in the TechSpec "Core Interfaces"
  and "Data Models" sections.
- MUST add `Interactive bool` and `InputCoordinator InputCoordinator` to
  `model.RuntimeConfig`; a nil coordinator and `Interactive=false` MUST preserve
  current behavior (no interactivity).
- MUST keep `internal/core/model` free of new external dependencies (only
  `context`).
- MUST add a compile-time interface assertion pattern note for future
  implementers (the concrete impl in task_03 will satisfy it).
</requirements>

## Subtasks
- [x] 1.1 Add the `InputCoordinator` interface with `Await` and `Submit`.
- [x] 1.2 Add the `PendingInput`, `InputOption`, and `UserResponse` value types.
- [x] 1.3 Extend `RuntimeConfig` with `Interactive` and `InputCoordinator` fields.
- [x] 1.4 Document the zero-value contract (nil coordinator ⇒ non-interactive).

## As-built notes (deviations for downstream tasks)
- **`InputCoordinator` carries `json:"-"`.** `model.RuntimeConfig` is mirrored
  into the public, JSON-serialized `extension.RuntimeConfig` and enforced by
  `sdk/extension/compat_test.go`. A live service handle must not appear in that
  public payload, so the field is tagged `json:"-"` (the test's sanctioned escape
  for runtime-only fields). The `Interactive bool` flag is plain config and was
  also added to the public mirror `extension.RuntimeConfig`. Decision approved by
  the user; the coordinator stays on `RuntimeConfig` (not moved to `RunScope`).
  Downstream tasks (04/05/06) still read `cfg.InputCoordinator` / `cfg.Interactive`.
- **Field is `UserResponse.Canceled` (one `l`).** The repo's `misspell` linter
  (US locale) flags the standalone word "Cancelled"; the TechSpec sample spelled
  it `Cancelled` but the as-built field is `Canceled`.

## Implementation Details
Add the types to the `internal/core/model` package (new file, e.g.
`input.go`), and extend the existing `RuntimeConfig` struct. Follow the field
shapes in the TechSpec "Core Interfaces" / "Data Models" sections — do not invent
extra fields. No behavior is wired here; this task only provides the contracts.

### Relevant Files
- `internal/core/model` — home of `RuntimeConfig` and shared runtime structs; the
  new types belong here to avoid import cycles between agent/daemon/executor.

### Dependent Files
- `internal/core/agent/client.go` — will consume `InputCoordinator`/`UserResponse`
  in task_04.
- `internal/core/run/exec/exec.go` — will read the new `RuntimeConfig` fields in
  task_05.
- `internal/daemon/run_manager.go` — will construct and set the coordinator in
  task_06.

### Related ADRs
- [ADR-002: Block-in-callback permissions + live-session re-prompt](../adrs/adr-002.md) — defines the `Await`/`Submit` broker contract.
- [ADR-004: Interactivity is opt-in via a flag set at run start](../adrs/adr-004.md) — motivates the `Interactive` flag on `RuntimeConfig`.

## Deliverables
- `InputCoordinator`, `PendingInput`, `InputOption`, `UserResponse` types in
  `internal/core/model`.
- `RuntimeConfig.Interactive` and `RuntimeConfig.InputCoordinator` fields.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests for the type contract via downstream compilation **(REQUIRED)**

## Tests
- Unit tests:
  - [ ] `UserResponse` zero value has `Cancelled=false` and empty `OptionID`/`Text`.
  - [ ] A `RuntimeConfig` with `Interactive=false` and nil `InputCoordinator` is a
        valid, usable zero-extended config (struct construction + field read).
  - [ ] `PendingInput` round-trips its fields (ID, Kind, Text, Options) without loss.
- Integration tests:
  - [ ] Package `internal/core/model` builds and its existing tests still pass with
        the new fields present (no regression in `RuntimeConfig` consumers).
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- `go build ./...` succeeds with the new types
- Downstream packages can reference the new types without import cycles
