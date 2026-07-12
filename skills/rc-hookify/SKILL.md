---
name: rc-hookify
description: Authors a new RC Claude Code hook from a plain-language description — emits a fail-open shell script in the house style under hooks/scripts/, wires it into hooks/hooks.json, documents it, and verifies it. Use when the user wants a deterministic guardrail, formatter, observer, or session warm-start turned into a hook. Do not use for application code, one-off shell scripts that aren't hooks, or editing existing guard logic unrelated to hooks.
user-invocable: true
model: sonnet
effort: medium
---

# Hookify

Turn "I want the agent to always/never do X" into a RC hook. A rule that must hold *every time*
belongs in a hook, not in prose (CLAUDE.md). This skill writes one in the repo's house style —
never blocks the session on its own errors, guards by pattern, and is registered + documented.

Read `hooks/scripts/_lib.sh` and one existing hook (`hooks/scripts/observe.sh` for an observer,
`hooks/scripts/git-guard.sh` for a blocker) before writing — mirror them exactly.

## Step 1 — pick the event

Match the intent to an event; the full I/O contract is in `references/hook-events.md` (read it —
stdout reaches the model only on some events, and only `PreToolUse` can block).

| Intent | Event | Matcher |
|---|---|---|
| Block/warn before a tool runs | `PreToolUse` | tool name (`Bash`, `Edit\|Write\|MultiEdit`, `Task`) |
| Format/observe/nudge after a tool ran | `PostToolUse` | tool name |
| Warm-start context / surface state at session start | `SessionStart` | omit (all sources) |
| Persist state before compaction / at end | `PreCompact` / `SessionEnd` | omit — side-effect only, model never sees stdout |
| Notify on stop / permission prompt | `Stop` / `Notification` | omit |

## Step 2 — write the script (house style)

`hooks/scripts/<verb-noun>.sh`, `chmod +x`. Name it by what it does (`go-fmt`, `commit-guard`,
`memory-load`). Skeleton — every real hook in `hooks/scripts/` follows it:

```bash
#!/usr/bin/env bash
# <Event>(<matcher>): one-line purpose.
#
# Fail-open: any environment problem exits 0 so it never breaks a session. Opt out via
# RC_DISABLED_HOOKS=<name>. <Say what it captures/does and what it never touches.>
set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "<name>" "minimal" || exit 0   # minimal|standard|strict — lowest profile it runs at

input=$(cat)                                    # the hook payload JSON on stdin
command -v jq >/dev/null 2>&1 || exit 0         # need jq? fail open if absent
# ... read fields with jq, do the work ...
exit 0
```

Non-negotiable house rules:

- **Fail open everywhere.** Missing `jq`, unreadable file, no `.rc`, parse error → `exit 0`. An
  environment problem must never block a tool call. Only `_lib.sh`'s `rc_block` exits non-zero.
- **Block only deliberately, only on `PreToolUse`.** To block, call `rc_block "<name>" "<message>"`
  (exits 2, message to the agent; honors `RC_DRY_RUN`). Match command/path substrings — a guardrail,
  not a sandbox.
- **Gate via `rc_hook_active`** (profile + `RC_DISABLED_HOOKS`); don't invent a new env var unless a
  hook needs its own kill-switch beyond the shared gate.
- **Guard by extension/pattern**, and stay silent when it doesn't apply (e.g. a non-`.go` file, a
  project with no `.rc`) — no noise in unrelated sessions.
- **Never capture file contents, secrets, or large blobs** — only names/paths/patterns.

## Step 3 — wire it into `hooks/hooks.json`

Add under the event, with the matcher from Step 1 and the plugin-root path:

```json
{ "type": "command", "command": "${CLAUDE_PLUGIN_ROOT}/hooks/scripts/<name>.sh" }
```

Reuse an existing event/matcher block if one matches; otherwise add a new one. Keep the JSON valid.

## Step 4 — document

Add one row to the table in `hooks/README.md` (event, matcher, script, effect). Keep it a
representative one-liner in the existing column style.

## Step 5 — verify (required)

- `bash -n hooks/scripts/<name>.sh` — syntax.
- `jq -e . hooks/hooks.json >/dev/null` — valid JSON, and confirm the new entry is present.
- **One self-check**: pipe a realistic payload and assert behavior — e.g.
  `echo '{"tool_name":"Bash","tool_input":{"command":"git reset --hard"}}' | bash hooks/scripts/<name>.sh; echo "exit $?"`
  — verify it blocks (exit 2) / stays silent (exit 0) / emits the expected text, plus the opt-out
  (`RC_DISABLED_HOOKS=<name>`) and the not-applicable case both exit 0 silently.

## Critical rules

- Hooks ship to **every plugin consumer** — a hook that wrongly blocks or errors breaks their
  sessions. Fail-open and precise guards are the whole safety model.
- This skill writes only under `hooks/`; it does not touch application source.
- If the rule is advisory (nice-to-have, not every-time), it belongs in `CLAUDE.md` or a skill —
  not a hook. Say so instead of writing one.
