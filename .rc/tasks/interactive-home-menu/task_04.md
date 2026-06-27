---
status: completed
title: Install RTK helper
type: backend
complexity: medium
dependencies:
  - task_01
---

# Task 4: Install RTK helper

## Overview
Add a helper that backs the "Install RTK" menu action by composing the existing RTK primitives: detect whether RTK is present, resolve the platform-appropriate install command, confirm with the user, and run it. RTK has no standalone command today, so this helper provides the in-session install flow without touching the setup command.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST compose `setup.DetectRTK`, `setup.ResolveRTKInstall`, and `setup.RunRTKInstall` from `internal/setup/rtk.go`; MUST NOT reimplement install logic.
- MUST short-circuit with an informative message when `DetectRTK` reports RTK already installed (include version/path when available).
- MUST confirm before installing using a `huh` confirmation themed with the existing CLI theme.
- MUST surface `RTKInstallCommand.Manual` guidance when `Runnable` is false instead of attempting to execute.
- MUST accept `context.Context` and an output writer, and wrap errors with `fmt.Errorf("…: %w", err)`.
- MUST expose its primitives as injectable seams so tests never spawn a real process.
</requirements>

## Subtasks
- [x] 4.1 Implement an `installRTK`-style helper orchestrating detect → resolve → confirm → run.
- [x] 4.2 Handle the already-installed case with a clear, non-error message.
- [x] 4.3 Handle the non-runnable platform case by printing manual guidance.
- [x] 4.4 Add a huh confirmation before running the install command (reuses `confirmRTKInstall`).
- [x] 4.5 Add unit tests with stubbed detect/resolve/run covering all branches without spawning processes.

## Implementation Details
Add the helper to `internal/cli/home_menu.go` (or sibling file). Reuse the RTK primitives and types in `internal/setup/rtk.go` (`RTKStatus{Installed, Path, Version}`, `RTKInstallCommand{Display, Name, Args, Runnable, Manual}`). Resolve `goos`/`hasBrew`/`hasCargo` following how setup wires `ResolveRTKInstall`. Use the existing huh confirm + theme pattern (`darkHuhTheme()`). See TechSpec "Integration Points" and ADR-005.

### Relevant Files
- `internal/setup/rtk.go` — `DetectRTK`, `ResolveRTKInstall`, `RunRTKInstall`, and RTK types.
- `internal/setup/rtk_test.go` — existing patterns for testing RTK resolution/detection.
- `internal/cli/setup.go` — how RTK detect/install are wired into setup (`runPromptField`, confirm patterns) and theme usage.
- `internal/cli/theme.go` — `darkHuhTheme()` for the confirmation prompt.

### Dependent Files
- `internal/cli/home_menu.go` — dispatcher in task_05 routes the Install RTK option here.

### Related ADRs
- [ADR-005: Dispatch menu actions by re-invoking existing cobra subcommands](../adrs/adr-005.md) — Specifies Install RTK as a small helper composing existing primitives.

## Deliverables
- Install RTK helper in `internal/cli` composing the existing RTK primitives with a confirmation step.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests for the detect-only / already-installed path (no real install) **(REQUIRED)**

## Tests
- Unit tests:
  - [x] Already installed: stubbed `DetectRTK` returns `Installed: true` → helper prints version/path message and does not call run.
  - [x] Runnable install confirmed: stubbed detect (absent) + resolve (`Runnable: true`) + confirm "yes" → run stub invoked exactly once.
  - [x] Confirmation declined: confirm "no" → run stub not invoked, returns nil.
  - [x] Non-runnable platform: resolve returns `Runnable: false` → helper prints `Manual` guidance and does not call run.
  - [x] Run failure: run stub returns an error → helper returns a wrapped error.
- Integration tests:
  - [x] The detect-only path resolves a platform command string via the real seams and stops before execution (no process spawned).
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- Helper installs RTK only after confirmation, short-circuits when present, and never spawns a process in tests.
- `make verify` passes.
