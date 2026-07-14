# PRD: Interactive Home Menu for `rc`

## Overview

Running `rc` with no subcommand today shows a welcome banner (wordmark, tips, and a status footer) on an interactive terminal. Users who do not already know the subcommand names have no guided, browsable way to reach the most common actions.

This feature adds an **interactive home menu** to the bare `rc` command. The existing banner stays at the top, and a navigable list of actions appears directly below it in the same screen. From this menu the user can List Skills, Install RTK, run Setup, open Help, or Quit — each action runs within the same session.

The menu serves two audiences equally: newcomers, who discover what RC can do without memorizing commands, and recurring users, who get a fast hub on the bare command. Named subcommands continue to work directly and bypass the menu, so the menu is an additive entry point, not a gate.

## Goals

- Give every user a single, discoverable entry point when they type `rc` with no arguments.
- Let users reach the most common actions (List Skills, Install RTK, Setup, Help) and exit, without knowing subcommand names.
- Preserve the existing RC banner and brand identity.
- Keep expert workflows frictionless: direct subcommands still bypass the menu.
- Make the home menu the consistent default face of the bare command.

## User Stories

**Primary persona — New RC user (just installed):**
- As a new user, I want to see a list of things I can do when I run `rc`, so that I can start without reading docs or memorizing commands.
- As a new user, I want to browse the available skills, so that I understand what RC offers before committing to setup.
- As a new user, I want to run setup from the menu, so that I can get RC configured immediately.

**Primary persona — Recurring RC user:**
- As a recurring user, I want a quick hub on the bare command, so that I can jump to common actions without typing subcommands.
- As a recurring user, I want to install RTK from the menu when it is missing, so that I complete my environment in one place.
- As a recurring user, I want to leave the menu quickly, so that it never gets in my way.

**Secondary persona — Scripts / CI / AI agents:**
- As an automated caller, I want `rc` to never hang waiting for input I cannot provide, so that my pipeline or agent run completes reliably.

## Core Features

Listed in priority order. The MVP is exactly the five menu items below.

1. **Home menu on bare `rc`** — When the user runs `rc` with no subcommand, the splash shows the existing banner on top and a navigable selection list of actions directly below it. This is the default behavior of the bare command.

2. **Navigation and selection** — The list supports arrow keys and j/k to move, Enter to confirm the highlighted item, Esc or q to quit, and number keys as a quick-select shortcut. The currently highlighted item is visually clear.

3. **List Skills (read-only)** — Selecting this shows the catalog of available skills with their names and descriptions. It does not install anything; users install via Setup. After viewing, the user returns to the menu or exits.

4. **Install RTK** — Selecting this runs the RTK installation guidance/flow within the session (RTK is the optional external skills runtime). If RTK is already present, the user is informed.

5. **Setup** — Selecting this runs the existing RC setup flow within the session.

6. **Help** — Selecting this presents the standard RC help/usage content.

7. **Quit** — Selecting this (or pressing Esc/q) exits cleanly with no action.

**Interaction between features:** Items 3–6 execute their action in the same session ("run the action right here") rather than only printing a command to copy. List Skills is intentionally read-only and does not overlap with Setup's install responsibility.

## User Experience

**Entry:** The user types `rc` and presses Enter. The banner renders as today, with the action list immediately below it. The first item is highlighted by default.

**Primary flow (new user, browse then setup):**
1. User runs `rc` and sees banner + menu.
2. User arrows to "List Skills", presses Enter, and reviews the catalog.
3. User returns to the menu, selects "Setup", and completes configuration in-session.

**Primary flow (recurring user, quick action):**
1. User runs `rc`.
2. User presses the number for "Install RTK" (or arrows + Enter).
3. The RTK flow runs; on completion the user is back at a clear state.

**Expert path:** A user who types `rc setup` (or any named subcommand) goes straight to that command and never sees the menu.

**UI/UX considerations:**
- Navigation must accept both arrow keys and j/k; Enter confirms; Esc/q quits.
- The highlighted item and the available keys should be visibly indicated.
- Visual style stays consistent with the current RC banner and brand theme.
- "Quit" is always present as an explicit, obvious exit.

**Onboarding and discoverability:** The menu turns "what can this do?" into a browsable list, which is the central onboarding benefit. Labels use plain action language (List Skills, Install RTK, Setup, Help, Quit).

## High-Level Technical Constraints

- Must integrate with the existing bare-command splash so the banner and menu render together.
- Must reuse the existing RC brand visual identity.
- The five menu actions map to capabilities RC already exposes (skill catalog, RTK install guidance, setup, help); the menu must not introduce a divergent second way of performing them that could drift from the canonical flows.
- Must not break direct subcommand invocation.
- Must not hang non-interactive callers (CI, pipes, agents) — see Risks and Open Questions.

## Non-Goals (Out of Scope)

- Installing or modifying skills from the "List Skills" view (read-only for MVP; installation stays in Setup).
- Adding menu items beyond the five specified (e.g., tasks, reviews, ext, daemon) — deferred to future phases.
- A configurable or plugin-based menu registry.
- Live search/filter within the menu (low value at five items).
- Changing the behavior or content of the existing banner itself.
- Localization of menu labels.

## Phased Rollout Plan

### MVP (Phase 1)
- Banner + navigable menu on bare `rc`.
- Five items: List Skills (read-only), Install RTK, Setup, Help, Quit.
- Arrow + j/k navigation, Enter to confirm, Esc/q to quit, number quick-select.
- Each action runs in-session; direct subcommands bypass the menu.
- Defined, non-hanging behavior for non-interactive callers.
- **Success criteria to proceed:** New users can reach each action from the menu without prior command knowledge; expert subcommand paths are unaffected; non-interactive runs never hang.

### Phase 2
- Optional additional menu items beyond the initial five (e.g., tasks, reviews), based on observed demand.
- **Success criteria to proceed:** Demand signal for more actions and confirmed menu usability at the larger item count.

### Phase 3
- Optional in-menu enhancements (e.g., filtering, richer skill browsing, install-from-list) if justified by usage.
- **Success criteria:** Sustained menu usage and clear user requests for these capabilities.

## Success Metrics

- New users reach an action (any of the five) from the menu without consulting docs in usability checks.
- Selecting an item reliably performs the intended action in-session.
- Expert subcommand invocations remain unchanged (zero regressions in direct-command behavior).
- Non-interactive invocations (CI, pipe, agent) complete without hanging in 100% of cases.
- Exit via Quit/Esc/q is always available and predictable.

## Risks and Mitigations

- **Adoption risk — experts find the menu intrusive.** Mitigation: direct subcommands always bypass the menu; the menu only affects the zero-argument path.
- **Behavior risk — hang in non-interactive contexts.** Keyboard navigation needs an interactive terminal; a menu awaiting input where no input device exists can block CI/agents indefinitely. Mitigation: define a non-blocking behavior for the no-input-device case (tracked in Open Questions and ADR-002); the no-hang guarantee is a Phase 1 success criterion.
- **Consistency risk — menu actions drift from canonical flows.** Mitigation: menu items invoke the same underlying capabilities as the named subcommands rather than reimplementing them.
- **Scope creep.** Mitigation: MVP fixed at five items; additional actions and in-menu features are explicitly deferred.

## Architecture Decision Records

- [ADR-001: Integrated home menu as the default entry point for bare `rc`](adrs/adr-001.md) — Banner plus navigable five-item list, each action executed in-session, over a minimal launcher or an extensible hub.
- [ADR-002: Home menu as the universal default, even outside an interactive terminal](adrs/adr-002.md) — The home menu is the default for bare `rc`; the exact safe behavior for non-interactive callers is deferred to the TechSpec.

## Open Questions

- **Non-interactive behavior:** What exactly should `rc` do when there is no interactive input device (CI, pipe, agent) while still honoring "present the home menu" and never hanging? Candidate: emit the menu's contents as static output and exit. To be resolved in the TechSpec.
- **Post-action state:** After an in-session action completes (e.g., Setup), should the user return to the menu or should `rc` exit? Default assumption: return to a clear state; confirm during TechSpec.
- **Install RTK when already present:** Exact messaging/flow when RTK is already installed — confirm wording during TechSpec.
