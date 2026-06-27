---
provider: manual
pr:
round: 1
round_created_at: 2026-06-20T20:49:52Z
status: resolved
file: internal/cli/home_menu.go
line: 219
severity: low
author: claude-code
provider_ref:
---

# Issue 002: Doubled "install rtk" prefix in wrapped error message

## Review Comment

The Install RTK dispatch path wraps the action error with the same prefix the
helper already applied, producing a duplicated phrase in the surfaced message.

In `dispatchHomeMenu` (line ~219):

```go
case menuInstallRTK:
    if err := actions.installRTK(); err != nil {
        return fmt.Errorf("install rtk: %w", err)
    }
```

and `installRTK` itself already wraps run failures (line ~166):

```go
return fmt.Errorf("install rtk: %w", err)
```

A run failure therefore reads: `install rtk: install rtk: run "brew install rtk": …`.
(The other branches are fine: `install rtk: detect rtk: …`, `list skills: load
skill catalog: …`.)

Suggested fix: drop the redundant outer prefix for the install case, or give the
dispatcher a distinct context word, e.g. `fmt.Errorf("home menu action: %w", err)`
for the install branch — so the message has a single, non-repeated prefix.

## Triage

- Decision: `VALID`
- Notes:
  - Root cause: `dispatchHomeMenu` frames each action error with the action name
    ("install rtk: %w"), and `installRTK` ALSO wrapped its run failure with the
    same noun ("install rtk: %w"), while the innermost `setup.RunRTKInstall`
    already returns `run "<display>": <cause>`. Result on a run failure:
    `install rtk: install rtk: run "...": ...`.
  - Fix: keep the dispatcher's uniform per-action frame and remove only the
    redundant inner wrap on the run failure (the inner error is already
    contextual). The `detect rtk:` wrap is kept because `DetectRTK`'s error has
    no such context. Final message is now `install rtk: run "...": ...`.
