#!/usr/bin/env bash
# Prints the directories whose files changed since the last codemap generation, one per line,
# so rc-codemap regenerates only stale maps. With --record, stores the current HEAD as the new
# baseline instead of printing. State: <root>/.rc/codemap/last-commit (a git SHA).
#
# Fails OPEN: on any git problem (not a repo, no git) it prints every tracked directory so the
# caller does a safe full rebuild, and --record becomes a no-op. Never edits tracked files.
#
# Usage:
#   stale.sh [root]            # list stale directories (default root: $PWD)
#   stale.sh --record [root]   # record current HEAD as the baseline
set -u

mode="list"
if [ "${1:-}" = "--record" ]; then
    mode="record"
    shift
fi
root="${1:-$PWD}"

cd "$root" 2>/dev/null || exit 0
command -v git >/dev/null 2>&1 || { [ "$mode" = "list" ] && exit 0; exit 0; }
git rev-parse --is-inside-work-tree >/dev/null 2>&1 || exit 0

state_dir="$root/.rc/codemap"
last_file="$state_dir/last-commit"

if [ "$mode" = "record" ]; then
    head=$(git rev-parse HEAD 2>/dev/null) || exit 0
    mkdir -p "$state_dir" 2>/dev/null || exit 0
    printf '%s\n' "$head" >"$last_file" 2>/dev/null || exit 0
    exit 0
fi

dirs_of() { # read newline-separated paths on stdin, print unique parent dirs
    grep -v '^$' | xargs -r -I{} dirname {} 2>/dev/null | sort -u
}

last=""
[ -f "$last_file" ] && last=$(cat "$last_file" 2>/dev/null)

if [ -z "$last" ] || ! git cat-file -e "$last^{commit}" 2>/dev/null; then
    # No usable baseline: everything is stale.
    git ls-files 2>/dev/null | dirs_of
    exit 0
fi

# committed changes since baseline + staged/unstaged + untracked-but-not-ignored
{
    git diff --name-only "$last" HEAD 2>/dev/null
    git status --porcelain 2>/dev/null | awk '{print $NF}'
    git ls-files --others --exclude-standard 2>/dev/null
} | dirs_of
