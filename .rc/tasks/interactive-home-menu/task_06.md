---
status: completed
title: Wire interactive branch in root command
type: backend
complexity: low
dependencies:
  - task_02
  - task_05
---

# Task 6: Wire interactive branch in root command

## Overview
Connect the interactive menu into the bare `rc` handler: on a terminal, render the banner once and then run the interactive menu loop. This completes the feature end-to-end, with the non-interactive path already handled by task_02.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST, on the interactive branch (`terminalWidth(cmd.OutOrStdout()) > 0`), render the banner and then call `runHomeMenu(cmd, width)`.
- MUST return the error from `runHomeMenu` (after task_05 has mapped abort to nil) so exit codes are correct.
- MUST keep the non-interactive branch from task_02 unchanged.
- MUST NOT alter direct subcommand dispatch — only the no-args interactive path changes.
- MUST keep the banner render call consistent with the current `lipgloss.Fprintln(cmd.OutOrStdout(), …)` output convention.
</requirements>

## Subtasks
- [x] 6.1 Update the interactive branch of the bare-command `RunE` to render banner + call `runHomeMenu` (extracted into `runBareCommand`).
- [x] 6.2 Return the menu loop's error to the caller.
- [x] 6.3 Verify the non-interactive branch (task_02) remains intact.
- [x] 6.4 Add integration coverage for the interactive wiring seam that does not require a PTY (injected `homeMenuRunner` seam + `runBareCommand`).

## Implementation Details
Modify the `RunE` at `internal/cli/root.go` (interactive branch around line 86). Replace the banner-only return with: render banner, then `return runHomeMenu(cmd, width)`. The non-interactive branch added in task_02 stays as-is. Because true huh navigation needs a PTY, integration tests target the wiring/branch selection and delegate menu-behavior coverage to task_05's dispatcher tests. See TechSpec "Development Sequencing" step 6.

### Relevant Files
- `internal/cli/root.go` — bare-command `RunE`; interactive branch modified here.
- `internal/cli/home_menu.go` — `runHomeMenu` from task_05.
- `internal/cli/setup.go` — `terminalWidth` branch decision.
- `internal/cli/root_command_execution_test.go` — patterns for root command execution tests.

### Dependent Files
- None — this is the final wiring task.

### Related ADRs
- [ADR-002: Home menu as the universal default, even outside an interactive terminal](../adrs/adr-002.md) — Completes the "menu is the default face of bare `rc`" intent for the interactive case.
- [ADR-001: Integrated home menu as the default entry point for bare `rc`](../adrs/adr-001.md) — Banner + interactive menu on the bare command.

## Deliverables
- Modified `internal/cli/root.go` interactive branch invoking `runHomeMenu`.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests confirming branch selection and that direct subcommands bypass the menu **(REQUIRED)**

## Tests
- Unit tests:
  - [x] On the interactive branch, the handler renders the banner before invoking the menu (verified via an injected `homeMenuRunner` seam capturing call order).
  - [x] The handler returns the error produced by `runHomeMenu` unchanged.
  - [x] The non-interactive branch still produces the static enumerated menu (regression guard for task_02).
- Integration tests:
  - [x] With a non-interactive writer, bare `rc` exits promptly (no menu loop entered) — confirms branch selection (`TestBareCommandNonInteractive*`).
  - [x] A direct subcommand (e.g., `agents list`) runs without entering the home menu (`TestDirectSubcommandBypassesHomeMenu`).
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- Bare `rc` on a terminal shows banner + interactive menu; non-interactive path and direct subcommands unaffected.
- `make verify` passes.
