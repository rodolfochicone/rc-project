---
status: completed
title: Interactive menu loop and action dispatcher
type: backend
complexity: medium
dependencies:
  - task_01
  - task_03
  - task_04
---

# Task 5: Interactive menu loop and action dispatcher

## Overview
Implement the interactive home menu: a `huh.NewSelect` list that loops until the user selects Quit or aborts, dispatching each selection to its action. Setup and Help re-invoke existing cobra subcommands; List Skills and Install RTK call the task_03 and task_04 helpers. This is the core interactive behavior of the bare `rc` command on a terminal.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST implement `runHomeMenu(cmd *cobra.Command, width int) error` using `huh.NewSelect[...]` themed with the existing CLI theme, reusing the option label table from task_01.
- MUST loop: present the select, dispatch the choice, then re-present until the user selects Quit.
- MUST treat huh's user-abort (Esc/Ctrl-C) as a clean Quit, not a surfaced error.
- MUST dispatch Setup and Help by resolving the corresponding subcommand from `cmd.Root()` and invoking it (no setup refactor); MUST dispatch List Skills and Install RTK to the task_03 / task_04 helpers.
- MUST propagate non-abort action errors wrapped with `fmt.Errorf("…: %w", err)`.
- MUST keep the dispatcher mapping table-driven and injectable so routing is unit-testable without a PTY or real setup run.
</requirements>

## Subtasks
- [x] 5.1 Build the huh select from the shared option labels and run it in a loop (`selectHomeOption` + `dispatchHomeMenu`).
- [x] 5.2 Implement the dispatcher mapping each `homeMenuOption` to its action.
- [x] 5.3 Resolve and invoke Setup (via `runSetupSubcommand`) / Help (via `cmd.Root().Help()`) from the cobra command graph.
- [x] 5.4 Route List Skills and Install RTK to their helpers.
- [x] 5.5 Map user-abort to Quit and exit the loop cleanly; propagate real errors.
- [x] 5.6 Add unit tests for dispatch routing, loop termination, and abort handling.

## Implementation Details
Extend `internal/cli/home_menu.go`. Build the select following the `huh.NewSelect`/`huh.NewForm(...).WithTheme(...)` pattern in `internal/cli/setup.go` (`resolveSkillSelection`, `runPromptField`). Resolve subcommands via `cmd.Root().Commands()` by name (Setup) and `cmd.Root().Help()` (Help). Factor the dispatch as a map/switch from `homeMenuOption` to a `func() error` so tests inject spies. See TechSpec "Core Interfaces", "Technical Considerations", ADR-003 (huh + loop, no number shortcut) and ADR-005 (re-dispatch).

### Relevant Files
- `internal/cli/setup.go` — `huh` select/form + theme patterns and `terminalWidth`.
- `internal/cli/home_menu.go` — option model (task_01), List Skills (task_03), Install RTK (task_04) to dispatch to.
- `internal/cli/theme.go` — `darkHuhTheme()` for the select.
- `internal/cli/root.go` — provides the `*cobra.Command` whose root is used to resolve Setup/Help.

### Dependent Files
- `internal/cli/root.go` — task_06 calls `runHomeMenu` on the interactive branch.

### Related ADRs
- [ADR-003: Use huh.NewSelect for the home menu, looping until Quit](../adrs/adr-003.md) — UI library, loop, abort-as-quit, no number shortcut.
- [ADR-005: Dispatch menu actions by re-invoking existing cobra subcommands](../adrs/adr-005.md) — Setup/Help re-dispatch and helper routing.
- [ADR-001: Integrated home menu as the default entry point for bare `rc`](../adrs/adr-001.md) — In-session action execution.

## Deliverables
- `runHomeMenu(cmd, width)` with loop, dispatcher, and abort handling in `internal/cli/home_menu.go`.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests for Setup/Help re-dispatch via a spy command under a test root **(REQUIRED)**

## Tests
- Unit tests:
  - [x] Dispatcher routes `menuListSkills` to the List Skills helper, `menuInstallRTK` to the Install RTK helper (verified via injected spies).
  - [x] Selecting `menuQuit` ends the loop and returns nil without invoking any action.
  - [x] A simulated user-abort error from the select is mapped to a clean exit (nil error), not surfaced.
  - [x] A non-abort error returned by an action is propagated wrapped.
  - [x] After a non-Quit action completes, the loop requests the select again (loop-continues behavior).
- Integration tests:
  - [x] With a spy subcommand registered under a test root named `setup`, selecting Setup resolves and invokes that subcommand exactly once.
  - [x] Selecting Help triggers the root help output.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- Menu loops through actions and exits cleanly on Quit/abort; Setup/Help re-dispatch and helper routing verified.
- `make verify` passes.
