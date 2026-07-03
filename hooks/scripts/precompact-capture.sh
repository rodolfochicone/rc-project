#!/usr/bin/env bash
# PreCompact: before the conversation is summarized, remind the agent to persist anything
# durable it learned this session, so the learning survives compaction. Curated memory and
# instincts are the store that outlives the context window; a summary is not.
#
# Opt-in: does nothing unless RC_RECALL=1, so it adds zero overhead by default. Never blocks.
# No binary or network needed. Only nudges inside an rc workspace.
set -u
[ "${RC_RECALL:-0}" = "1" ] || exit 0
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "precompact-capture" "standard" || exit 0

# Resolve the nearest .rc directory walking up from CWD (same idiom as observe.sh).
dir="$PWD"
rc_dir=""
while :; do
    if [ -d "$dir/.rc" ]; then
        rc_dir="$dir/.rc"
        break
    fi
    [ "$dir" = "/" ] || [ -z "$dir" ] && break
    dir="$(dirname "$dir")"
done
[ -z "$rc_dir" ] && exit 0

note="rc capture — context is about to compact. Before it is summarized away, persist anything durable you learned this session: record cross-cutting decisions, conventions, and non-obvious gotchas with \`rc memory add\` (see the rc-project-memory skill), and distill repeated corrections into instincts (rc-instincts). Skip secrets, transient state, and anything already obvious from the repository."

if command -v jq >/dev/null 2>&1; then
    jq -nc --arg c "$note" '{hookSpecificOutput:{hookEventName:"PreCompact",additionalContext:$c}}'
else
    printf '%s\n' "$note"
fi
exit 0
