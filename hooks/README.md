# Claude Code hooks (RC plugin)

Deterministic guardrails that ship with the RC Claude Code plugin. They turn rules
that are otherwise only described in `CLAUDE.md` and the RC skills
(`rc-final-verify`, `rc-git`, `rc-execute-task`) into enforcement that does not
depend on the model choosing to obey.

These hooks load automatically when the RC plugin is enabled (convention-based
`hooks/hooks.json`). On OpenCode, the plugin `opencode/plugin/rc-hooks.ts` shells
out to the **same** scripts, so both channels share one source of truth.

## Hooks

| Event          | Matcher                         | Script               | Effect                                                                                                                                   |
| -------------- | ------------------------------- | -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `SessionStart` | —                               | `memory-load.sh`     | Warm-starts the session: injects a bounded summary of `.rc/memory/` (facts, learnings, pending-observation nudge) into context. Never blocks; silent when there is no memory. |
| `PreToolUse`   | `Bash`                          | `git-guard.sh`       | Blocks destructive/history-rewriting git: `reset --hard`, `restore`, `clean`, change-discarding `checkout`, `rebase`, `filter-branch`, force-push. |
| `PreToolUse`   | `Bash`                          | `commit-guard.sh`    | Blocks AI attribution in commit messages (`Co-Authored-By`, "Generated with Claude", 🤖).                                                |
| `PreToolUse`   | `Bash`                          | `db-guard.sh`        | Enforces read-only database access: blocks DB-client commands (`psql`, `mysql`, …) carrying write/DDL SQL.                               |
| `PreToolUse`   | `Edit\|Write\|MultiEdit`        | `gateguard.sh`       | Fact-forcing gate: blocks the **first** edit of each file once, with an investigation checklist; the retry proceeds (per-session marker). |
| `PostToolUse`  | `Edit\|Write\|MultiEdit\|Bash`  | `observe.sh`         | Appends a compact observation (tool + truncated target, never file contents) to `.rc/memory/observations.jsonl` for the `rc-memory` learning loop. Opt out with `RC_INSTINCTS=0`. |
| `PostToolUse`  | `Edit\|Write\|MultiEdit\|Task`  | `repair-guidance.sh` | On a *repairable* edit/delegation failure, feeds one piece of corrective guidance back to the agent instead of letting it retry the identical failing call. |
| `Stop`         | —                               | `notify.sh`          | Plays a "done" sound at end of turn. Opt-in via `RC_SOUND=1`; never blocks.                                                              |
| `Notification` | —                               | `notify.sh`          | Plays an "attention" sound when the agent needs input. Same `RC_SOUND=1` opt-in.                                                         |

`_lib.sh` is not a hook — it is the shared helper library (profile gating,
kill-switch, dry-run, block semantics) sourced by every script.

Blocking hooks exit `2` and return the message on stderr to the agent. Allowed
calls exit `0`.

## Configuration

Environment knobs, read at hook invocation time (see `_lib.sh`):

- `RC_HOOK_PROFILE` — `minimal | standard | strict` (default: `standard`).
  Profiles are cumulative (`minimal ⊂ standard ⊂ strict`); each hook declares the
  lowest profile at which it activates — the git/commit guards are `minimal`
  (always on), the fact-gate waits for `strict`.
- `RC_DISABLED_HOOKS` — comma-separated hook names to force-skip
  (e.g. `"db-guard,gateguard"`).
- `RC_DRY_RUN` — `1` logs a would-block decision and allows the call instead.
- `RC_SOUND` / `RC_INSTINCTS` — opt-in for `notify.sh` / opt-out for `observe.sh`.

## Requirements

- `jq` on `PATH` (used to parse the hook payload). If `jq` is missing the hooks
  fail open (exit 0) so they never break a session.

## Notes & limits

- `commit-guard.sh` inspects the inline command, so it catches `-m`/heredoc
  messages but not `-F <file>`.
- `git-guard.sh` matches command substrings; it is a guardrail, not a sandbox.
  Hard enforcement still belongs to the permission system.
