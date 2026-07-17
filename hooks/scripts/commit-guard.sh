#!/usr/bin/env bash
# PreToolUse(Bash): block AI co-author / attribution trailers in commit messages.
# Best-effort: inspects the inline command (matches `-m`/heredoc; not `-F file`).
# Fail-open: no jq, no command, or no _lib.sh lets the call through.
#
# Active from the "minimal" profile up (always), and disablable via
# RC_DISABLED_HOOKS=commit-guard. Routes through _lib.sh's rc_block so RC_DRY_RUN=1
# is honored — without it, the escapes _lib.sh advertises never reach this hook.
set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "commit-guard" "minimal" || exit 0

input=$(cat)
command -v jq >/dev/null 2>&1 || exit 0

cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty')
[ -z "$cmd" ] && exit 0

case "$cmd" in
*"git commit"*) ;;
*) exit 0 ;;
esac

if printf '%s' "$cmd" | grep -qiE 'Co-Authored-By:.*(Claude|Anthropic)|Generated with \[?Claude|Claude Code|🤖'; then
    rc_block "commit-guard" 'remove AI attribution from the commit message (no AI "Co-Authored-By", "Generated with Claude", or 🤖 trailer).'
fi

exit 0
