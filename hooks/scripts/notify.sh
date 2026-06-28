#!/usr/bin/env bash
# Stop / Notification: play a short sound when the turn ends or the agent needs
# attention. Opt-in via RC_SOUND=1; does nothing otherwise. Never blocks.
#
# The event is read from the hook payload's `hook_event_name`:
#   Stop / SubagentStop  → "done" sound
#   Notification         → "attention" sound
set -u
[ "${RC_SOUND:-0}" = "1" ] || exit 0
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "notify" "minimal" || exit 0

input=$(cat 2>/dev/null || true)
event=""
if command -v jq >/dev/null 2>&1; then
    event=$(printf '%s' "$input" | jq -r '.hook_event_name // empty' 2>/dev/null || true)
fi

case "$event" in
Notification) mac_sound="/System/Library/Sounds/Funk.aiff" ;;
*) mac_sound="/System/Library/Sounds/Hero.aiff" ;;
esac

# Play in the background so the hook returns immediately. Pick the first player
# available for the platform; do nothing (still exit 0) if none is found.
if command -v afplay >/dev/null 2>&1; then
    [ -f "$mac_sound" ] && (afplay "$mac_sound" >/dev/null 2>&1 &)
elif command -v paplay >/dev/null 2>&1; then
    (paplay /usr/share/sounds/freedesktop/stereo/complete.oga >/dev/null 2>&1 &)
elif command -v aplay >/dev/null 2>&1; then
    (aplay -q /usr/share/sounds/alsa/Front_Center.wav >/dev/null 2>&1 &)
fi
exit 0
