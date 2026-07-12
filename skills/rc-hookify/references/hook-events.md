# Claude Code hook events — I/O contract

The load-bearing, easy-to-get-wrong details: when each event fires, whether the model sees stdout,
and what the payload carries. Get this wrong and a hook is silently useless (stdout ignored) or
breaks sessions (blocks when it shouldn't). Source: Claude Code hooks docs.

## Stdout & blocking, per event

| Event | Fires | stdout → model? | exit 2 effect | Use for |
|---|---|---|---|---|
| `PreToolUse` | before a tool runs | no (transcript only) | **blocks the tool**, stderr → agent | guards (block/warn) |
| `PostToolUse` | after a tool finished | no (via `additionalContext` JSON only) | tool already ran; stderr → agent as feedback | format, observe, nudge |
| `UserPromptSubmit` | user sends a message | **yes — added to context** | blocks the prompt | inject context on submit |
| `SessionStart` | session start (`startup`/`resume`/`clear`/`compact`) | **yes — added before first prompt** | n/a | warm-start / surface state |
| `PreCompact` | before compaction | **no** (debug log only) | stderr → user only | persist state (side-effect) |
| `SessionEnd` | session terminates | **no** (debug log only) | stderr → user only | cleanup, summaries |
| `Stop` | agent finished responding | no | can force continue (intrusive) | notify |
| `Notification` | permission/idle prompt | no | n/a | notify |

Key consequences:

- **Only `SessionStart` and `UserPromptSubmit` inject stdout into the model's context.** A "remind
  the agent" hook on `PreCompact`/`SessionEnd` is invisible to the model — don't put nudges there.
- **Only `PreToolUse` blocks the action.** `PostToolUse` exit 2 sends feedback but the tool already
  ran. Everything else: exit non-zero only affects the user/transcript, never the flow.
- Prefer plain stdout for `SessionStart` (documented: "text printed to stdout is added as context").
  JSON `{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"..."}}` also works.

## Payload on stdin (common fields)

All events receive JSON on stdin with at least: `session_id`, `transcript_path`, `cwd`,
`hook_event_name`. Event-specific:

- `PreToolUse` / `PostToolUse`: `tool_name`, `tool_input` (e.g. `tool_input.command`,
  `tool_input.file_path`), and on Post also `tool_response`.
- `SessionStart`: `source` (`startup`/`resume`/`clear`/`compact`), optional `model`, `agent_type`.
- `PreCompact` / `SessionEnd`: common fields only (no event-specific fields documented).

Read fields with `jq -r '.tool_name // empty'` etc., and **fail open if `jq` is missing**
(`command -v jq >/dev/null 2>&1 || exit 0`) — see `hooks/scripts/observe.sh` for the pattern.

## Matchers

`PreToolUse`/`PostToolUse` match on tool name (regex-ish alternation): `"Bash"`,
`"Edit|Write|MultiEdit"`, `"Task"`. `SessionStart` matches on `source`; omit the matcher to run on
all sources. Lifecycle events (`Stop`, `Notification`, `PreCompact`, `SessionEnd`) take no matcher —
omit it.
