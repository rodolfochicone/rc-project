#!/usr/bin/env bash
# PreToolUse(Bash): block destructive/history-rewriting git commands.
# Exit 2 blocks the tool call and feeds stderr back to the agent.
set -u

input=$(cat)
command -v jq >/dev/null 2>&1 || exit 0

cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty')
[ -z "$cmd" ] && exit 0

norm=$(printf '%s' "$cmd" | tr '\n' ' ' | tr -s '[:space:]' ' ')

block() {
    printf 'rc git-guard: %s\n' "$1" >&2
    exit 2
}

case "$norm" in
*"git reset --hard"*) block "'git reset --hard' discards committed/working changes. Forbidden without explicit user approval." ;;
*"git restore"*) block "'git restore' discards working-tree changes. Forbidden without explicit user approval." ;;
*"git clean"*) block "'git clean' deletes untracked files. Forbidden without explicit user approval." ;;
*"git checkout -- "* | *"git checkout ."*) block "'git checkout' that discards changes is forbidden. Use a non-destructive alternative or ask the user." ;;
*"git rebase"*) block "'git rebase' rewrites history. Forbidden without explicit user approval." ;;
*"git filter-branch"*) block "'git filter-branch' rewrites history. Forbidden." ;;
esac

case "$norm" in
*"git push"*"--force"* | *"git push"*" -f"* | *"git push"*"--force-with-lease"*) block "force-push is forbidden (rewrites shared history)." ;;
esac

exit 0
