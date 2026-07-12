#!/usr/bin/env bash
# PostToolUse(Edit|Write|MultiEdit|Bash): append a compact observation for the
# learning loop (the `rc-memory` skill distills these into learned patterns).
#
# On by default; set RC_INSTINCTS=0 to opt out. Never blocks and adds negligible
# overhead per tool call. Captures only the tool name and a truncated target (file
# path or command) — never file contents — into <project>/.rc/memory/observations.jsonl.
set -u
[ "${RC_INSTINCTS:-1}" = "0" ] && exit 0
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "observe" "minimal" || exit 0

input=$(cat)
command -v jq >/dev/null 2>&1 || exit 0

tool=$(printf '%s' "$input" | jq -r '.tool_name // .tool // empty')
file=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')
cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty' | cut -c1-200)
target="$file"
[ -z "$target" ] && target="$cmd"
[ -z "$tool$target" ] && exit 0

# Resolve the nearest .rc directory walking up from CWD; default to ./.rc.
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
[ -z "$rc_dir" ] && rc_dir="$PWD/.rc"
mkdir -p "$rc_dir/memory" 2>/dev/null || exit 0

ts="$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo unknown)"
jq -nc --arg ts "$ts" --arg tool "$tool" --arg target "$target" \
    '{ts:$ts, tool:$tool, target:$target}' >>"$rc_dir/memory/observations.jsonl" 2>/dev/null || exit 0
exit 0
