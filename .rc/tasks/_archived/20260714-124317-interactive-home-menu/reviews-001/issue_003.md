---
provider: manual
pr:
round: 1
round_created_at: 2026-06-20T20:49:52Z
status: resolved
file: internal/cli/root.go
line: 55
severity: low
author: claude-code
provider_ref:
---

# Issue 003: homeMenuRunner is a mutable package var used only as a test seam

## Review Comment

`homeMenuRunner` is a package-level mutable variable that exists solely so tests
can inject a spy and assert the interactive wiring without a PTY:

```go
var homeMenuRunner = runHomeMenu
```

Production code thus carries mutable global state for test purposes, and the
safety argument relies on a convention ("only the interactive branch reads it,
and no parallel test reaches that branch"). This is brittle: a future test that
exercises the interactive branch in parallel, or any future production reader,
would introduce a data race that the current tests would not catch. It is also a
mild form of production-pollution-for-tests.

Suggested fix: thread the runner through construction instead of a package var —
for example capture it in a closure inside `newRootCommandWithDefaults`, add it
to `commandStateDefaults`, or pass it as a parameter to `runBareCommand`
(`runBareCommand(cmd, width, runMenu)`). Tests then inject the spy locally with
no shared mutable state. If the package-var seam is retained intentionally, add a
short comment documenting why threading was rejected so the trade-off is explicit.

## Triage

- Decision: `VALID`
- Notes:
  - Root cause: the interactive runner was exposed as a mutable package var
    (`var homeMenuRunner = runHomeMenu`) purely so tests could inject a spy,
    relying on the convention that no parallel test reaches the interactive
    branch to stay race-free.
  - Fix: removed the package var and threaded the runner as an explicit
    parameter — `runBareCommand(cmd, width, runMenu func(*cobra.Command, int) error)`.
    Production wires `runHomeMenu` from `RunE`; tests pass a local spy. No shared
    mutable state remains, so the wiring tests no longer swap a global.
