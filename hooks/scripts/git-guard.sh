#!/usr/bin/env bash
# PreToolUse(Bash): block destructive/history-rewriting git commands.
# Exit 2 blocks the tool call and feeds stderr back to the agent.
set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "git-guard" "minimal" || exit 0

input=$(cat)
command -v jq >/dev/null 2>&1 || exit 0

cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty')
[ -z "$cmd" ] && exit 0

norm=$(printf '%s' "$cmd" | tr '\n' ' ' | tr -s '[:space:]' ' ')

case "$norm" in
*"git reset --hard"*) rc_block "git-guard" "'git reset --hard' discards committed/working changes. Forbidden without explicit user approval." ;;
*"git restore"*) rc_block "git-guard" "'git restore' discards working-tree changes. Forbidden without explicit user approval." ;;
*"git clean"*) rc_block "git-guard" "'git clean' deletes untracked files. Forbidden without explicit user approval." ;;
*"git checkout -- "* | *"git checkout ."*) rc_block "git-guard" "'git checkout' that discards changes is forbidden. Use a non-destructive alternative or ask the user." ;;
*"git rebase"*) rc_block "git-guard" "'git rebase' rewrites history. Forbidden without explicit user approval." ;;
*"git filter-branch"*) rc_block "git-guard" "'git filter-branch' rewrites history. Forbidden." ;;
esac

case "$norm" in
*"git push"*"--force"* | *"git push"*" -f"* | *"git push"*"--force-with-lease"*) rc_block "git-guard" "force-push is forbidden (rewrites shared history)." ;;
esac

exit 0
