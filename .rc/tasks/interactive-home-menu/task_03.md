---
status: completed
title: List Skills read-only helper
type: backend
complexity: low
dependencies:
  - task_01
---

# Task 3: List Skills read-only helper

## Overview
Add a read-only helper that loads the effective skill catalog and prints each skill's name and description. This backs the "List Skills" menu action; it must never install or mutate anything, keeping installation the sole responsibility of Setup.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST add a helper in `internal/cli` that loads the catalog via the existing `loadEffectiveSetupCatalog(ctx, resolver)` and prints each `setup.Skill` `Name` and `Description` to a provided writer.
- MUST be strictly read-only — no install, no filesystem writes, no catalog mutation.
- MUST accept its output writer and `context.Context` as parameters so it is testable and cancellable.
- MUST wrap errors with context using `fmt.Errorf("…: %w", err)` and return them rather than printing-and-swallowing.
- MUST be injectable/seam-friendly so task_05's dispatcher can route to it and tests can stub the catalog loader.
</requirements>

## Subtasks
- [x] 3.1 Implement a `listSkills`-style helper that loads the effective catalog and renders names + descriptions.
- [x] 3.2 Format output consistently with the brand theme (lipgloss), grouped/ordered for readability.
- [x] 3.3 Expose the catalog-loading dependency as a function value (or accept it) so it can be stubbed in tests.
- [x] 3.4 Add unit tests with a stubbed catalog covering rendering and the read-only guarantee.

## Implementation Details
Add the helper to `internal/cli/home_menu.go` (or a sibling file in `internal/cli`). Reuse `loadEffectiveSetupCatalog` from `internal/cli/setup_assets.go` and the `setup.Skill{Name, Description}` fields from `internal/setup/types.go`. Resolver options follow the existing setup pattern (`setup.ResolverOptions`). See TechSpec "Implementation Design → Data Models" and ADR-005.

### Relevant Files
- `internal/cli/setup_assets.go` — `loadEffectiveSetupCatalog(ctx, resolver)` returning `setup.EffectiveCatalog`.
- `internal/setup/types.go` — `Skill{ Name, Description, … }` fields rendered by the helper.
- `internal/cli/setup.go` — `resolverOptions()` / `skillNames()` patterns for catalog handling.
- `internal/charmtheme` / `internal/cli/theme.go` — styling for the listing.

### Dependent Files
- `internal/cli/home_menu.go` — dispatcher in task_05 routes the List Skills option here.

### Related ADRs
- [ADR-005: Dispatch menu actions by re-invoking existing cobra subcommands](../adrs/adr-005.md) — Specifies List Skills as a small new helper reusing the catalog loader (no new command).
- [ADR-001: Integrated home menu as the default entry point for bare `rc`](../adrs/adr-001.md) — List Skills is read-only in the MVP.

## Deliverables
- Read-only List Skills helper in `internal/cli`.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests verifying the helper renders a real (bundled) catalog without mutation **(REQUIRED)**

## Tests
- Unit tests:
  - [x] With a stubbed loader returning two known skills, output contains both names and both descriptions.
  - [x] With a stubbed loader returning an empty catalog, output is a clear "no skills" message and returns nil.
  - [x] When the stubbed loader returns an error, the helper returns a wrapped error and writes nothing partial that implies success.
  - [x] The helper invokes no install path (structurally guaranteed — `listSkills` depends only on a read-only loader and a writer; the bundled-catalog test confirms zero filesystem mutation).
- Integration tests:
  - [x] Calling the helper with the real `loadEffectiveSetupCatalog` against bundled skills lists at least one known bundled skill (e.g., `rc-create-prd`) and leaves the filesystem unchanged.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- Helper lists skill names + descriptions and performs zero mutations.
- `make verify` passes.
