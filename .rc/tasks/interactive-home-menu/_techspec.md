# TechSpec: Interactive Home Menu for `rc`

## Executive Summary

The bare `rc` command currently renders a startup splash (banner + tips + footer) on an interactive terminal and falls back to `cmd.Help()` when output is non-interactive (`internal/cli/root.go:82-91`). This change extends that handler: on an interactive terminal it renders the banner and then enters an interactive menu loop built with `huh.NewSelect`; on a non-interactive stream it renders the banner followed by a static enumerated menu and exits without blocking.

The menu offers five actions — List Skills (read-only), Install RTK, Setup, Help, Quit — and dispatches each by re-invoking existing cobra subcommands (Setup, Help) or by calling small new local helpers that reuse existing primitives (the setup catalog loader for List Skills; `internal/setup/rtk.go` for Install RTK). The primary trade-off: we keep the change surgical and consistent with the existing huh-based CLI layer at the cost of dropping the number quick-select shortcut from the PRD (huh's select has no native number binding), and we reuse setup logic in place rather than refactoring it into standalone functions.

## System Architecture

### Component Overview

All changes live in the `internal/cli` package; no new package or directory.

- **Bare-command handler** (`internal/cli/root.go`, modified): branches on `terminalWidth(cmd.OutOrStdout())`. Interactive → render banner once, then run the menu loop. Non-interactive → render banner + static enumerated menu, return nil.
- **Menu loop + dispatcher** (`internal/cli`, new, e.g. `home_menu.go`): renders the `huh.NewSelect` list, maps the selection to an action, executes it, and re-renders until Quit or user abort.
- **Static menu renderer** (`internal/cli`, new, in the same file): produces the enumerated non-interactive listing, styled consistently with the banner via lipgloss/`charmtheme`.
- **List Skills helper** (`internal/cli`, new): loads the effective catalog via the existing `loadEffectiveSetupCatalog()` and prints `Skill.Name` + `Skill.Description`. Read-only.
- **Install RTK helper** (`internal/cli`, new): composes `setup.DetectRTK`, `setup.ResolveRTKInstall`, and `setup.RunRTKInstall` with a huh confirmation.
- **Setup / Help actions**: resolved from the cobra command graph (`cmd.Root()`) and invoked directly; no new logic.

**Data flow:** `RunE` → (interactive) menu loop → user selects option → dispatcher → action (cobra subcommand re-invoke OR local helper) → control returns to the loop → repeat until Quit.

**External interactions:** Install RTK shells out to the platform install command via the existing `setup.RunRTKInstall` (`exec.CommandContext`). No other external systems.

## Implementation Design

### Core Interfaces

The menu is modeled as a small set of options dispatched by a switch. The primary type other code depends on is the option identity and its dispatcher signature:

```go
// homeMenuOption identifies a single entry in the bare-`rc` home menu.
type homeMenuOption string

const (
	menuListSkills homeMenuOption = "list-skills"
	menuInstallRTK homeMenuOption = "install-rtk"
	menuSetup      homeMenuOption = "setup"
	menuHelp       homeMenuOption = "help"
	menuQuit       homeMenuOption = "quit"
)

// runHomeMenu renders the interactive menu under the banner and loops
// until the user selects Quit or aborts (Esc/Ctrl-C). It dispatches each
// selection by re-invoking cobra subcommands or local helpers.
func runHomeMenu(cmd *cobra.Command, width int) error

// renderStaticHomeMenu returns the banner plus an enumerated, non-blocking
// listing of the menu options for non-interactive streams.
func renderStaticHomeMenu(width int) string
```

Dispatch maps each `homeMenuOption` to an action; `menuQuit` and huh's user-abort error both end the loop.

### Data Models

No new persistent data models. Existing types are reused:

- `setup.Skill{ Name, Description, ... }` (`internal/setup/types.go`) — source for List Skills.
- `setup.RTKStatus{ Installed, Path, Version }` and `setup.RTKInstallCommand{ Display, Name, Args, Runnable, Manual }` (`internal/setup/rtk.go`) — source for Install RTK.
- `setup.EffectiveCatalog{ Skills, ... }` — returned by `loadEffectiveSetupCatalog()`.

### API Endpoints

Not applicable — this is a CLI command surface, not a network API.

## Integration Points

- **RTK installer (external process):** Install RTK uses `setup.RunRTKInstall(ctx, command, w)`, which executes the resolved platform command (`brew`, install script, or `cargo`) with stdout/stderr wired to the command's writer. If `RTKInstallCommand.Runnable` is false, surface `RTKInstallCommand.Manual` guidance instead of executing. RTK detection (`setup.DetectRTK`) reports already-installed status so the helper can short-circuit with a message.

## Impact Analysis

| Component | Impact Type | Description and Risk | Required Action |
|-----------|-------------|---------------------|-----------------|
| `internal/cli/root.go` (bare `RunE`) | modified | Replace the `cmd.Help()` non-TTY fallback and add the interactive menu loop. Low risk; isolated to the no-args path. | Branch on `terminalWidth`; call `runHomeMenu` / `renderStaticHomeMenu`. |
| `internal/cli/home_menu.go` | new | Menu loop, dispatcher, static renderer, List Skills + Install RTK helpers. Medium risk (new interactive code path). | Implement with huh + existing theme; map abort to clean exit. |
| `internal/cli/banner.go` | reused | Banner functions called as-is. No change. | None. |
| `internal/cli/setup.go` (`terminalWidth`, `loadEffectiveSetupCatalog`, themes) | reused | Reused for TTY detection, catalog loading, and huh theming. No change. | None. |
| `internal/setup/rtk.go` | reused | RTK primitives composed by the helper. No change. | None. |
| Cobra command graph (`setup`, help) | reused | Setup/Help re-invoked from the menu. Risk: passing the correct command/context. | Resolve subcommand from `cmd.Root()`; invoke its `RunE`/`Help`. |

## Testing Approach

### Unit Tests

- **Static menu renderer** (`renderStaticHomeMenu`): assert the output contains the banner and all five enumerated options in order; table-driven over representative widths. This is the deterministic, non-interactive path and the primary regression guard for the "never hang" guarantee.
- **Dispatcher mapping**: test that each `homeMenuOption` routes to the correct action and that `menuQuit` ends the loop. Inject the action functions (catalog loader, RTK primitives, subcommand invoker) as function fields so the test asserts routing without running real setup or shelling out — mock at the seam, per the repo's interface-driven testing convention.
- **List Skills helper**: with a stubbed catalog loader returning known `Skill` values, assert the printed output includes each name and description; assert read-only (no install call).
- **Install RTK helper**: with stubbed `DetectRTK` (already-installed → short-circuit message) and stubbed resolve/run (not-runnable → manual guidance; runnable → invokes runner once). Assert no real process is spawned.
- **TTY branch selection**: assert `terminalWidth == 0` selects the static path and `> 0` selects the interactive path. Verify the non-interactive path never calls into huh (no blocking).

Tests must encode intent: the non-interactive test exists specifically to prove CI/agent callers cannot hang, and the read-only test proves List Skills does not mutate state.

### Integration Tests

- **Bare command, non-interactive**: execute the root command with a non-TTY writer and assert it returns promptly with banner + enumerated menu and a nil error (the realistic CI/pipe scenario).
- **Setup/Help re-dispatch**: assert that selecting Setup/Help resolves and invokes the corresponding cobra subcommand object (verified via a spy command registered under a test root), confirming the re-dispatch wiring.

Interactive huh navigation itself is not unit-tested (it requires a PTY); coverage focuses on the dispatcher, helpers, and the static path.

## Development Sequencing

### Build Order

1. **Static menu renderer + option model** (`homeMenuOption`, constants, `renderStaticHomeMenu`) — no dependencies. Deterministic and fully unit-testable.
2. **Wire the non-interactive branch in `root.go`** — depends on step 1. Replace the `cmd.Help()` fallback with `renderStaticHomeMenu` when `terminalWidth == 0`; lock in the "never hang" behavior with the integration test.
3. **List Skills helper** — depends on step 1 (option model) and the existing catalog loader. Independent of RTK and the loop.
4. **Install RTK helper** — depends on step 1; composes existing `internal/setup/rtk.go` primitives. Independent of List Skills.
5. **Menu loop + dispatcher** (`runHomeMenu`) — depends on steps 1, 3, 4 (actions to dispatch) and on resolving Setup/Help from the cobra graph. Maps huh selection to actions; maps abort to clean exit; loops until Quit.
6. **Wire the interactive branch in `root.go`** — depends on steps 2 and 5. Render banner once, then call `runHomeMenu` when `terminalWidth > 0`.
7. **Tests and `make verify`** — depends on all prior steps. Unit + integration coverage; full pipeline must pass.

### Technical Dependencies

- No new third-party dependencies: `huh`, `lipgloss`, `cobra`, `golang.org/x/term`, and the `charmtheme` package are already in the module.
- No infrastructure or external service prerequisites.

## Monitoring and Observability

CLI-local feature; no runtime telemetry. Per project convention, any non-fatal operational messages (e.g., RTK install progress/failure) use `log/slog` and user-facing output goes through the command's writer with lipgloss styling. Errors from actions are wrapped with `fmt.Errorf("context: %w", err)` and surfaced to the user; huh user-abort is treated as a normal Quit, not an error.

## Technical Considerations

### Key Decisions

- **Decision:** Build the menu with `huh.NewSelect`, looping until Quit.
  - **Rationale:** Matches the existing CLI interaction pattern and brand theme; minimal new code.
  - **Trade-offs:** No number quick-select (not native to huh).
  - **Alternatives rejected:** Custom Bubble Tea model — more code, diverges from the CLI layer's convention.

- **Decision:** Non-interactive streams get banner + static enumerated menu, then exit.
  - **Rationale:** Honors "always present the home menu" while guaranteeing no hang for CI/pipes/agents.
  - **Trade-offs:** Non-interactive output is informational only.
  - **Alternatives rejected:** Reverting to `cmd.Help()` (conflicts with the product direction).

- **Decision:** Dispatch actions by re-invoking existing cobra subcommands plus two small helpers.
  - **Rationale:** Surgical; reuses tested setup/help/RTK behavior.
  - **Trade-offs:** Menu depends on the cobra command graph being reachable from the handler (it is — same root).
  - **Alternatives rejected:** Extracting reusable functions from setup/RTK — larger blast radius, out of scope.

### Known Risks

- **huh abort handling (likely):** Esc/Ctrl-C surfaces as huh's user-abort error; if not mapped to "Quit" it would print a spurious error. Mitigation: detect the abort error explicitly and exit the loop cleanly.
- **Subcommand re-invoke fidelity (medium):** Re-invoking Setup/Help must pass the correct command/context so flags, IO streams, and signal handling behave as in direct invocation. Mitigation: resolve and invoke the subcommand object from `cmd.Root()` so it carries its own configuration; cover with the re-dispatch integration test.
- **TTY misclassification (low):** `terminalWidth` returns 0 for writers without an `Fd()`; treated as non-interactive, which is the safe default (no hang). No mitigation required beyond reusing the existing, proven helper.

## Architecture Decision Records

- [ADR-001: Integrated home menu as the default entry point for bare `rc`](adrs/adr-001.md) — Banner plus navigable five-item list with in-session actions, over a minimal launcher or extensible hub.
- [ADR-002: Home menu as the universal default, even outside an interactive terminal](adrs/adr-002.md) — The home menu is the default for bare `rc`; safe non-interactive behavior deferred to this TechSpec.
- [ADR-003: Use huh.NewSelect for the home menu, looping until Quit](adrs/adr-003.md) — huh-based select consistent with the CLI layer; drops the number quick-select shortcut.
- [ADR-004: Non-interactive realization — banner plus static enumerated menu](adrs/adr-004.md) — Concretizes ADR-002: render banner + enumerated options and exit when non-interactive.
- [ADR-005: Dispatch menu actions by re-invoking existing cobra subcommands](adrs/adr-005.md) — Reuse setup/help commands and add small helpers for List Skills and Install RTK, instead of refactoring setup.
