#!/usr/bin/env bash
# SessionStart: if an rc workflow is in progress, inject a one-line reminder of which
# pipeline phase it is at and the next step, so the agent stays on the pipeline instead of
# improvising. It reads only local .rc files and infers the phase from which artifacts exist.
#
# Opt-in: does nothing unless RC_PHASE_REMINDER=1, so it adds zero overhead by default.
# Never blocks. Needs neither a binary nor network.
set -u
[ "${RC_PHASE_REMINDER:-0}" = "1" ] || exit 0
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "phase-reminder" "standard" || exit 0

# Resolve the nearest .rc directory walking up from CWD (same idiom as session-recall.sh).
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

tasks="$rc_dir/tasks"
[ -d "$tasks" ] || exit 0

# Pick the most recently modified workflow directory (excluding the archive).
wf=""
for d in $(ls -1dt "$tasks"/*/ 2>/dev/null); do
    case "$d" in
    *"/_archived/"*) continue ;;
    esac
    wf="${d%/}"
    break
done
[ -z "$wf" ] && exit 0
slug="$(basename "$wf")"

# Infer the phase from which artifacts are present, latest-first.
if ls -d "$wf"/reviews-*/ >/dev/null 2>&1; then
    phase="Review / Remediation"
    next="triage and fix open review issues with the rc-fix-reviews skill, then re-review"
elif [ -f "$wf/_tasks.md" ]; then
    phase="Execution"
    next="implement the pending task files in order with the rc-execute-task skill; verify with rc-final-verify"
elif [ -f "$wf/_techspec.md" ]; then
    phase="Task decomposition"
    next="run /rc-create-tasks to break the TechSpec into task files"
elif [ -f "$wf/_prd.md" ]; then
    phase="Technical design"
    next="run /rc-create-techspec to turn the PRD into a technical spec"
else
    phase="Requirements"
    next="run /rc-create-prd to write the PRD"
fi

note="rc phase reminder — active workflow \`$slug\` is at the **$phase** phase. Next: $next. Follow the pipeline in order; do not skip phases."

if command -v jq >/dev/null 2>&1; then
    jq -nc --arg c "$note" '{hookSpecificOutput:{hookEventName:"SessionStart",additionalContext:$c}}'
else
    printf '%s\n' "$note"
fi
exit 0
