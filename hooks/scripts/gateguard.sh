#!/usr/bin/env bash
# PreToolUse(Edit|Write|MultiEdit): fact-forcing gate before the FIRST edit of a
# file in a session. The first attempt to edit a given file is blocked once with
# a short investigation checklist; a per-session marker is then recorded so the
# retry (and every later edit of that file) proceeds. This interrupts blind edits
# and forces the agent to ground itself in callers/contracts first — the cheapest
# defense against editing without understanding impact.
#
# Off by default. Active only under RC_HOOK_PROFILE=strict. Fails open: any
# environment problem (no jq, unwritable temp dir) lets the edit through.
set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "gateguard" "strict" || exit 0

input=$(cat)
command -v jq >/dev/null 2>&1 || exit 0

path=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')
[ -z "$path" ] && exit 0
session=$(printf '%s' "$input" | jq -r '.session_id // "nosession"')

# Skip noise: vcs internals, vendored deps, lockfiles.
case "$path" in
*/.git/* | */node_modules/* | */vendor/* | *go.sum) exit 0 ;;
esac

key=$(printf '%s' "$path" | cksum | awk '{print $1}')
state_dir="${TMPDIR:-/tmp}/rc-gateguard/${session}"
marker="${state_dir}/${key}"

# Already greeted this file in this session — let the edit through.
[ -f "$marker" ] && exit 0

# Record the marker first so the retry passes even if this block is the last word.
mkdir -p "$state_dir" 2>/dev/null || exit 0
: >"$marker" 2>/dev/null || exit 0

base=$(basename "$path")
rc_block "gateguard" "First edit to ${base} this session. Before editing, state in your reply: (1) who imports/calls this file or the symbol you're changing; (2) which public API/contract this change affects; (3) the shape of the data involved; (4) the user's literal instruction this satisfies. Then make the edit again — this gate won't fire twice for ${base}."
