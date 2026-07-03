#!/usr/bin/env bash
# SessionStart: surface a small POINTER to this project's durable rc knowledge so it is
# recalled without the agent having to remember to look — the highest-confidence instincts
# plus a nudge to search curated memory. It injects a pointer, not a dump: bodies stay in
# the store and are pulled on demand via `rc memory search`, keeping context cheap.
#
# Opt-in: does nothing unless RC_RECALL=1, so it adds zero overhead by default. Never blocks.
# Reads only local .rc files; needs neither the rc binary nor network. Bounded by
# RC_RECALL_MAX_CHARS (default 1500).
set -u
[ "${RC_RECALL:-0}" = "1" ] || exit 0
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "session-recall" "standard" || exit 0

max="${RC_RECALL_MAX_CHARS:-1500}"

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

block=""
instincts="$rc_dir/instincts/INSTINCTS.md"
if [ -f "$instincts" ]; then
    top=$(grep -E '^- \[0\.[89]' "$instincts" 2>/dev/null | head -n 10)
    [ -n "$top" ] && block="High-confidence project instincts:
$top
"
fi
if [ -d "$rc_dir/memory" ] || [ -f "$rc_dir/memory.db" ]; then
    block="${block}Curated rc project memory is available — run \`rc memory search \"<task terms>\"\` before deciding or implementing."
fi
[ -z "$block" ] && exit 0

note="rc recall — durable knowledge for this project:
$(printf '%s' "$block" | cut -c1-"$max")"

if command -v jq >/dev/null 2>&1; then
    jq -nc --arg c "$note" '{hookSpecificOutput:{hookEventName:"SessionStart",additionalContext:$c}}'
else
    printf '%s\n' "$note"
fi
exit 0
