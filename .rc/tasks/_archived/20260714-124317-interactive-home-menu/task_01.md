---
status: completed
title: Option model and static menu renderer
type: backend
complexity: low
dependencies: []
---

# Task 1: Option model and static menu renderer

## Overview
Define the home menu's option identity model and the non-interactive renderer that prints the banner followed by an enumerated, non-blocking list of options. This is the foundation every other task depends on, and the static renderer is the primary guard for the "never hang" guarantee on non-interactive streams.

<critical>
- ALWAYS READ the PRD and TechSpec before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST define a `homeMenuOption` type and constants for the five options: List Skills, Install RTK, Setup, Help, Quit (see TechSpec "Core Interfaces").
- MUST provide `renderStaticHomeMenu(width int) string` that returns the existing banner plus the options as a numbered list (e.g., `1. List Skills` … `5. Quit`).
- MUST reuse `renderStartupBanner` / banner helpers and the `charmtheme` lipgloss styles for visual consistency; MUST NOT alter existing banner functions.
- MUST be a pure, side-effect-free string builder (no I/O, no blocking) so callers control output and it can never hang.
- MUST live in `internal/cli` as a new file; MUST NOT introduce a new package or directory.
</requirements>

## Subtasks
- [x] 1.1 Add `homeMenuOption` type and the five option constants with stable string identities.
- [x] 1.2 Define the ordered, user-facing label for each option in one place reusable by both the static renderer and the future interactive menu.
- [x] 1.3 Implement `renderStaticHomeMenu(width int)` composing banner + enumerated options.
- [x] 1.4 Add unit tests covering option ordering and static rendering at representative widths.

## Implementation Details
Create `internal/cli/home_menu.go`. Define the option model and labels, and `renderStaticHomeMenu`. Reuse `renderStartupBanner(width)` from `internal/cli/banner.go` and lipgloss styling from `internal/charmtheme`. See TechSpec "Core Interfaces" for the type shape and "System Architecture" for the component boundary. Keep the option label table as the single source of truth so task_05's interactive select reuses it.

### Relevant Files
- `internal/cli/banner.go` — `renderStartupBanner(width)` and styling helpers to compose into the static menu.
- `internal/cli/theme.go` / `internal/charmtheme` — brand lipgloss styles for consistent rendering.
- `internal/cli/banner_test.go` — pattern reference for table-driven render tests (`strings.Contains`, width assertions).

### Dependent Files
- `internal/cli/root.go` — will call `renderStaticHomeMenu` in task_02.
- New `internal/cli/home_menu.go` — extended by task_03, task_04, task_05.

### Related ADRs
- [ADR-004: Non-interactive realization — banner plus static enumerated menu](../adrs/adr-004.md) — Defines the static enumerated output this task renders.
- [ADR-001: Integrated home menu as the default entry point for bare `rc`](../adrs/adr-001.md) — Confirms the five-item set and ordering.

## Deliverables
- `homeMenuOption` type, option constants, and shared label table in `internal/cli/home_menu.go`.
- `renderStaticHomeMenu(width int) string`.
- Unit tests with 80%+ coverage **(REQUIRED)**
- Integration tests for the static render path exercised via the root command **(deferred to task_02; this task covers the renderer in isolation)** **(REQUIRED)**

## Tests
- Unit tests:
  - [x] `renderStaticHomeMenu` output contains the banner marker (block-art `█`) and all five labels in order `List Skills`, `Install RTK`, `Setup`, `Help`, `Quit`.
  - [x] Each option line is numbered `1.`–`5.` matching declared order.
  - [x] Rendering at width 80 and width 120 both include every option (no truncation of option lines).
  - [x] Option constants have distinct, stable string values.
- Integration tests:
  - [ ] Covered in task_02 (root command non-interactive path); not applicable to the pure renderer here.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- `renderStaticHomeMenu` produces banner + five enumerated options with zero I/O or blocking.
- No changes to existing banner functions; `make verify` passes.
