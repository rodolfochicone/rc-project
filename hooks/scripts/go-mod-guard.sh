#!/usr/bin/env bash
# PreToolUse(Edit|Write|MultiEdit): block hand-editing go.mod / go.sum.
# Dependencies must change via `go get` / `go mod tidy`, not manual edits.
set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "go-mod-guard" "standard" || exit 0

input=$(cat)
command -v jq >/dev/null 2>&1 || exit 0

path=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')
[ -z "$path" ] && exit 0

case "$(basename "$path")" in
go.mod | go.sum)
    rc_block "go-mod-guard" "do not edit $(basename "$path") by hand. Use \`go get <pkg>\` for dependencies (and \`go mod tidy\`). If this is a legitimate change (go directive / replace), ask the user first."
    ;;
esac

exit 0
