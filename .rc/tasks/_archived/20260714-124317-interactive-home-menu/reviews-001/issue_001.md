---
provider: manual
pr:
round: 1
round_created_at: 2026-06-20T20:49:52Z
status: resolved
file: internal/cli/home_menu.go
line: 247
severity: medium
author: claude-code
provider_ref:
---

# Issue 001: Install RTK action uses a non-signal-aware context

## Review Comment

`defaultHomeMenuActions` captures `ctx := cmd.Context()` (which is
`context.Background()` for the bare `rc` command, since `RunE` is reached via a
plain `Execute()`), and the Install RTK action forwards it to `installRTK` →
`setup.RunRTKInstall(ctx, …)`, which shells out via `exec.CommandContext`.

The canonical setup flow does **not** do this: `setupCommandState.run`
(`internal/cli/setup.go`) derives a SIGINT/SIGTERM-cancellable context with
`signalCommandContext(cmd)` before calling `ensureRTK`/`RunRTKInstall`. As a
result, an RTK install launched from the home menu cannot be gracefully
cancelled the way the same install can when launched via `rc setup` — Ctrl-C
during a long network install does not flow through context cancellation, so the
child process is not torn down via the established mechanism. This also conflicts
with the project's concurrency discipline ("use `select` with `ctx.Done()` /
context cancellation for long-running operations") and the PRD goal that menu
actions reuse the canonical flows rather than diverge from them.

Suggested fix: derive a signal-aware context for the in-session actions, mirroring
setup, e.g. in `defaultHomeMenuActions`:

```go
ctx, stop := signalCommandContext(cmd)
// ensure stop() is called when the menu loop exits
```

and thread that context (with proper `stop()` cleanup) into `installRTK` so the
install honors Ctrl-C/SIGTERM exactly like `rc setup`.

## Triage

- Decision: `VALID`
- Notes:
  - Root cause: `defaultHomeMenuActions` captures `cmd.Context()` (Background for
    the bare command) and forwards it to `installRTK` → `setup.RunRTKInstall`,
    which shells out via `exec.CommandContext`. The canonical `rc setup` flow
    derives a SIGINT/SIGTERM-cancellable context via `signalCommandContext(cmd)`
    before running the installer; the menu did not, so an RTK install launched
    from the menu does not honor Ctrl-C/SIGTERM through context cancellation.
  - Fix: in `defaultHomeMenuActions`, derive `ctx, stop := signalCommandContext(cmd)`
    and ensure `stop()` runs when the menu loop exits. Threaded the cancel func
    out so `runHomeMenu` defers `stop()`; the same signal-aware ctx now feeds
    both List Skills and Install RTK, matching setup's interrupt behavior.
  - Verification (whole round): `make verify` — fmt + lint (0 issues) + build
    pass; `internal/cli` passes with `-race` except the PRE-EXISTING, unrelated
    failure `TestValidateTasksCommandPassesCommittedACPFixtures`, caused by 34
    ACP fixture files deleted in the working tree before this session (not
    touched by these fixes; restoring them needs explicit user permission).
