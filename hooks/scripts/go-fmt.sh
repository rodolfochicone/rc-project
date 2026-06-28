#!/usr/bin/env bash
# PostToolUse(Edit|Write|MultiEdit): gofmt the edited Go file.
# PostToolUse cannot block; this only normalizes formatting deterministically.
set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "go-fmt" "standard" || exit 0

input=$(cat)
command -v jq >/dev/null 2>&1 || exit 0

path=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')
[ -z "$path" ] && exit 0

case "$path" in
*.go) ;;
*) exit 0 ;;
esac

[ -f "$path" ] || exit 0
command -v gofmt >/dev/null 2>&1 || exit 0

gofmt -w "$path" >/dev/null 2>&1 || true
exit 0
