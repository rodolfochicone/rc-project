# Claude Code hooks (rc plugin)

Deterministic guardrails that ship with the rc Claude Code plugin. They turn rules
that are otherwise only described in `CLAUDE.md` and the rc skills
(`rc-final-verify`, `rc-git`, `rc-execute-task`) into enforcement that does not
depend on the model choosing to obey.

These hooks load automatically when the rc plugin is enabled (convention-based
`hooks/hooks.json`). They apply to the **Claude Code plugin channel** — i.e. when
you run `claude` directly in a repo with the plugin installed.

**OpenCode parity:** `rc setup` also installs an OpenCode plugin (`rc-hooks.ts`,
into both `.opencode/plugin/` and `.opencode/plugins/` for version compatibility)
that shells out to these _same_ scripts via OpenCode's `tool.execute.before` /
`tool.execute.after` hooks, so the guards and the instincts capture behave
identically there. The plugin registers its hooks only once per process even if
both copies load. The scripts are the single source of truth; each harness only
adapts how they are invoked.

They do **not** apply to rc's ACP execution pipeline today; deterministic gates
for the pipeline belong in the Go executor.

## Hooks

| Event          | Matcher                        | Script            | Profile    | Effect                                                                                                                                                                                                        |
| -------------- | ------------------------------ | ----------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PreToolUse`   | `Bash`                         | `git-guard.sh`    | `minimal`  | Blocks destructive/history-rewriting git: `reset --hard`, `restore`, `clean`, change-discarding `checkout`, `rebase`, `filter-branch`, force-push.                                                            |
| `PreToolUse`   | `Bash`                         | `commit-guard.sh` | `minimal`  | Blocks AI attribution in commit messages (`Co-Authored-By`, "Generated with Claude", 🤖).                                                                                                                     |
| `PreToolUse`   | `Edit\|Write\|MultiEdit`       | `go-mod-guard.sh` | `standard` | Blocks hand-editing `go.mod` / `go.sum`; directs to `go get`.                                                                                                                                                 |
| `PreToolUse`   | `Edit\|Write\|MultiEdit`       | `gateguard.sh`    | `strict`   | Fact-gate: blocks the **first** edit of each file in a session once, demanding the agent state callers, affected API, data shape, and the literal instruction. Off by default.                                |
| `PostToolUse`  | `Edit\|Write\|MultiEdit`       | `go-fmt.sh`       | `standard` | Runs `gofmt -w` on the edited `.go` file (never blocks).                                                                                                                                                      |
| `PostToolUse`  | `Edit\|Write\|MultiEdit\|Bash` | `observe.sh`      | opt-in     | Instincts capture: appends a compact `{tool, target}` observation to `.rc/instincts/observations.jsonl` for the `rc-instincts` skill. Off unless `RC_INSTINCTS=1`. Never blocks; never records file contents. |
| `Stop`         | —                              | `notify.sh`       | opt-in     | Plays a short "done" sound at end of turn. Off unless `RC_SOUND=1`. Never blocks.                                                                                                                             |
| `Notification` | —                              | `notify.sh`       | opt-in     | Plays an "attention" sound when the agent needs you (e.g. a permission prompt). Off unless `RC_SOUND=1`. Never blocks.                                                                                        |

Blocking hooks exit `2` and return the message on stderr to the agent. Allowed
calls exit `0`.

## Profiles & kill-switches

All hooks source `scripts/_lib.sh`, which gates them by an env-selected profile
and lets you disable any hook without editing config:

| Env var             | Values                              | Effect                                                                                                                                                                                            |
| ------------------- | ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `RC_HOOK_PROFILE`   | `minimal` \| `standard` \| `strict` | Cumulative. Default `standard`. Each hook declares the lowest profile at which it runs (see table). `minimal` keeps only the irreversible-action guards; `strict` adds the `gateguard` fact-gate. |
| `RC_DISABLED_HOOKS` | comma-separated hook names          | Force-skip specific hooks, e.g. `RC_DISABLED_HOOKS="go-fmt,gateguard"`.                                                                                                                           |
| `RC_DRY_RUN`        | `1`                                 | A hook that _would_ block instead logs `(dry-run, would block)` and allows.                                                                                                                       |

Hook names for the kill-switch are `git-guard`, `commit-guard`, `go-mod-guard`,
`gateguard`, `go-fmt`, `observe`, `notify`.

Two hooks have a separate opt-in and add zero overhead by default:

- `observe.sh` — does nothing unless `RC_INSTINCTS=1` (instincts capture).
- `notify.sh` — does nothing unless `RC_SOUND=1` (end-of-turn / attention sound). On
  macOS it uses `afplay` with system sounds (Hero = done, Funk = attention); on
  Linux it tries `paplay`/`aplay`; otherwise it is silent. The same `RC_SOUND=1`
  toggle drives the equivalent OpenCode plugin events (`session.idle`,
  `permission.asked`).

## Requirements

- `jq` on `PATH` (used to parse the hook payload). If `jq` is missing the hooks
  fail open (exit 0) so they never break a session.
- `gofmt` on `PATH` for `go-fmt.sh` (no-op if absent).
- `cksum`/`awk` on `PATH` for `gateguard.sh` (POSIX; fails open if absent).

## Notes & limits

- `commit-guard.sh` inspects the inline command, so it catches `-m`/heredoc
  messages but not `-F <file>`.
- `git-guard.sh` matches command substrings; it is a guardrail, not a sandbox.
  Hard enforcement still belongs to the permission system.
- `gateguard.sh` records a per-session marker under `$TMPDIR/rc-gateguard/<session>/`
  so it interrupts only the first edit of each file; it skips `.git/`,
  `node_modules/`, `vendor/`, and `go.sum`. It is a focus prompt, not a sandbox.
- `_lib.sh` is a sourced helper, not a hook; `rc setup` copies it next to the
  scripts so the copied channel resolves it the same way the plugin channel does.
