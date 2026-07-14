---
status: completed
title: Wire non-interactive branch in root command
type: backend
complexity: low
dependencies:
  - task_01
---

# Task 2: Wire non-interactive branch in root command

## Overview
Replace the current `cmd.Help()` fallback in the bare `rc` handler so that, on a non-interactive stream, the command renders the banner plus the static enumerated menu and returns immediately. This locks in the "never hang" guarantee for CI, pipes, and agent callers before any interactive code exists.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST branch on `terminalWidth(cmd.OutOrStdout())`: when it returns `0` (non-interactive), print `renderStaticHomeMenu(...)` to `cmd.OutOrStdout()` and return `nil`.
- MUST remove the existing `cmd.Help()` fallback from the bare no-args path (`internal/cli/root.go`).
- MUST NOT block, prompt, or read input on the non-interactive path.
- MUST leave the interactive (`width > 0`) branch rendering the banner as today (the interactive menu is wired in task_06).
- MUST preserve direct subcommand behavior — only the no-args path changes.
</requirements>

## Subtasks
- [x] 2.1 Update the bare-command `RunE` to call `renderStaticHomeMenu` on the non-interactive branch.
- [x] 2.2 Remove the `cmd.Help()` fallback from this path.
- [x] 2.3 Keep the interactive branch unchanged (banner only) pending task_06.
- [x] 2.4 Add integration tests asserting the non-interactive path returns promptly with banner + enumerated menu and nil error.

## Implementation Details
Modify the `RunE` at `internal/cli/root.go:82-91`. The current logic renders the banner when `terminalWidth > 0` and falls back to `cmd.Help()` otherwise; replace the fallback with `lipgloss.Fprintln(cmd.OutOrStdout(), renderStaticHomeMenu(width))` (width is `0` here, so the renderer uses its default-width path). See TechSpec "System Architecture" and "Development Sequencing" step 2.

### Relevant Files
- `internal/cli/root.go` — bare-command `RunE`; the only file modified by this task.
- `internal/cli/home_menu.go` — provides `renderStaticHomeMenu` from task_01.
- `internal/cli/setup.go` — `terminalWidth(w io.Writer)` used for the branch decision.
- `internal/cli/root_test.go`, `internal/cli/root_command_execution_test.go` — patterns for executing the root command with a custom writer.

### Dependent Files
- `internal/cli/root.go` — further edited by task_06 (interactive branch).

### Related ADRs
- [ADR-004: Non-interactive realization — banner plus static enumerated menu](../adrs/adr-004.md) — Directly implements this branch.
- [ADR-002: Home menu as the universal default, even outside an interactive terminal](../adrs/adr-002.md) — Honors "always present the home menu" without hanging.

## Deliverables
- Modified `internal/cli/root.go` non-interactive branch using `renderStaticHomeMenu`.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests for the bare command on a non-interactive writer **(REQUIRED)**

## Tests
- Unit tests:
  - [x] With a non-`Fd` writer (`terminalWidth == 0`), the handler writes output containing all five enumerated options and returns `nil`.
  - [x] The handler does not invoke `cmd.Help()` on the non-interactive path (assert via output not matching help usage text).
- Integration tests:
  - [x] Executing the root command with a `bytes.Buffer` output completes without blocking and emits banner + enumerated menu (realistic CI/pipe scenario).
  - [x] A known direct subcommand (e.g., `agents list`) still executes normally and bypasses the menu.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- Bare `rc` on a non-interactive stream prints banner + enumerated menu and exits with code 0, never hanging.
- `make verify` passes.
