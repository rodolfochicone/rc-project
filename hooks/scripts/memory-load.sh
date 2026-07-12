#!/usr/bin/env bash
# SessionStart: warm-start the session with this project's durable memory.
#
# Prints a bounded summary of .rc/memory/ to stdout, which Claude Code adds to the
# session context before the first prompt. It surfaces curated facts (INDEX.md) and
# distilled learnings (LEARNINGS.md), and — when raw observations pile up — nudges
# the agent to run the rc-memory skill to distill them. It closes the continuous-
# learning loop that the `observe` hook opens: observe captures, this surfaces.
#
# On by default; RC_DISABLED_HOOKS=memory-load (or a non-RC project) skips it. Never
# blocks and stays silent when there is no memory to load. Output is capped so it
# warms the session without eating the context window.
set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "memory-load" "minimal" || exit 0

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
mem="$rc_dir/memory"
[ -d "$mem" ] || exit 0

FACTS_CAP=20
LEARNINGS_CAP=20
OBS_NUDGE=40

idx="$mem/INDEX.md"
learn="$mem/LEARNINGS.md"
obs="$mem/observations.jsonl"

out=""
if [ -s "$idx" ]; then
    total=$(grep -c '^-' "$idx" 2>/dev/null || echo 0)
    body=$(grep '^-' "$idx" 2>/dev/null | head -"$FACTS_CAP")
    if [ -n "$body" ]; then
        out="${out}Facts (.rc/memory/INDEX.md):
${body}
"
        [ "$total" -gt "$FACTS_CAP" ] && out="${out}[+$((total - FACTS_CAP)) more — read INDEX.md]
"
    fi
fi

if [ -s "$learn" ]; then
    total=$(grep -c '^- \[' "$learn" 2>/dev/null || echo 0)
    body=$(head -"$LEARNINGS_CAP" "$learn" 2>/dev/null)
    if [ -n "$body" ]; then
        out="${out}
Learnings (.rc/memory/LEARNINGS.md, by domain then confidence):
${body}
"
        [ "$total" -gt "$LEARNINGS_CAP" ] && out="${out}[+more learnings — read LEARNINGS.md]
"
    fi
fi

if [ -s "$obs" ]; then
    n=$(wc -l <"$obs" 2>/dev/null | tr -d ' ')
    if [ "${n:-0}" -ge "$OBS_NUDGE" ]; then
        out="${out}
Note: ${n} raw observations pending in observations.jsonl — run the rc-memory skill to distill them into learnings.
"
    fi
fi

[ -z "$out" ] && exit 0
printf '=== RC memory (this project) ===\n%s=== end RC memory ===\n' "$out"
exit 0
