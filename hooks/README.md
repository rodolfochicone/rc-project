# Claude Code hooks (rc plugin)

Deterministic guardrails that ship with the rc Claude Code plugin. They turn rules
that are otherwise only described in `CLAUDE.md` and the rc skills
(`rc-final-verify`, `rc-git`, `rc-execute-task`) into enforcement that does not
depend on the model choosing to obey.

These hooks load automatically when the rc plugin is enabled (convention-based
`hooks/hooks.json`). They apply to the **Claude Code plugin channel** — i.e. when
you run `claude` directly in a repo with the plugin installed. They do **not**
apply to other agents, nor (today) to rc's ACP execution pipeline; deterministic
gates for the pipeline belong in the Go executor.

## Hooks

| Event         | Matcher                  | Script            | Effect                                                                                                                                             |
| ------------- | ------------------------ | ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PreToolUse`  | `Bash`                   | `git-guard.sh`    | Blocks destructive/history-rewriting git: `reset --hard`, `restore`, `clean`, change-discarding `checkout`, `rebase`, `filter-branch`, force-push. |
| `PreToolUse`  | `Bash`                   | `commit-guard.sh` | Blocks AI attribution in commit messages (`Co-Authored-By`, "Generated with Claude", 🤖).                                                          |
| `PreToolUse`  | `Edit\|Write\|MultiEdit` | `go-mod-guard.sh` | Blocks hand-editing `go.mod` / `go.sum`; directs to `go get`.                                                                                      |
| `PostToolUse` | `Edit\|Write\|MultiEdit` | `go-fmt.sh`       | Runs `gofmt -w` on the edited `.go` file (never blocks).                                                                                           |

Blocking hooks exit `2` and return the message on stderr to the agent. Allowed
calls exit `0`.

## Requirements

- `jq` on `PATH` (used to parse the hook payload). If `jq` is missing the hooks
  fail open (exit 0) so they never break a session.
- `gofmt` on `PATH` for `go-fmt.sh` (no-op if absent).

## Notes & limits

- `commit-guard.sh` inspects the inline command, so it catches `-m`/heredoc
  messages but not `-F <file>`.
- `git-guard.sh` matches command substrings; it is a guardrail, not a sandbox.
  Hard enforcement still belongs to the permission system.
