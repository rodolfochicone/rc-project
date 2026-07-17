#!/usr/bin/env bash
# Shared helpers for the rc Claude Code hook scripts. Sourced by each hook; not a
# hook itself. Centralizes profile/kill-switch gating, dry-run, and consistent
# block semantics so individual hooks stay small.
#
# Env knobs (read at hook invocation time):
#   RC_HOOK_PROFILE    minimal | standard | strict   (default: standard)
#   RC_DISABLED_HOOKS  comma-separated hook names to force-skip (e.g. "db-guard,gateguard")
#   RC_DRY_RUN         1 to log a would-block decision and allow instead of blocking
#
# Profiles are cumulative: minimal ⊂ standard ⊂ strict. A hook declares the
# lowest profile at which it is active; it runs at that profile and every higher
# one. git/commit guards are "minimal" (always on); repair-guidance and the fact-gate
# default off until "standard"/"strict" respectively.

rc_profile_rank() {
    case "$1" in
    minimal) printf '0' ;;
    standard) printf '1' ;;
    strict) printf '2' ;;
    *) printf '1' ;;
    esac
}

# rc_hook_active <hook-name> <min-profile>
# Returns 0 (run) when the hook is enabled for the active profile and not
# explicitly disabled; returns 1 (skip) otherwise.
rc_hook_active() {
    rc__name="$1"
    rc__minp="$2"
    case ",${RC_DISABLED_HOOKS:-}," in
    *",${rc__name},"*) return 1 ;;
    esac
    [ "$(rc_profile_rank "${RC_HOOK_PROFILE:-standard}")" -ge "$(rc_profile_rank "$rc__minp")" ]
}

# rc_block <hook-name> <message>
# Exit 2 to block the tool call and surface <message> to the agent. Under
# RC_DRY_RUN=1 it logs the decision and exits 0 so the call proceeds.
rc_block() {
    if [ "${RC_DRY_RUN:-0}" = "1" ]; then
        printf 'rc %s (dry-run, would block): %s\n' "$1" "$2" >&2
        exit 0
    fi
    printf 'rc %s: %s\n' "$1" "$2" >&2
    exit 2
}

# --- PreToolUse structured decisions (JSON on stdout, exit 0) -----------------
# Claude Code parses hook JSON only on exit 0; on exit 2 it ignores stdout and
# reads stderr instead. The two signalling styles must not be mixed inside one
# hook, so a hook that needs "ask" uses rc_deny/rc_ask for *every* decision it
# makes and never calls rc_block. Deny-only hooks (db-guard, commit-guard) stay
# on rc_block. Source: code.claude.com/docs/en/hooks.md — "PreToolUse decision
# control" + "Exit code output".

# rc__json_escape <string> — minimal escaping for a JSON string body.
rc__json_escape() {
    printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g' | tr -d '\n'
}

# rc__decision <hook-name> <ask|deny> <reason>
rc__decision() {
    rc__reason=$(rc__json_escape "rc $1: $3")
    printf '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"%s","permissionDecisionReason":"%s"}}\n' "$2" "$rc__reason"
    exit 0
}

# rc_deny <hook-name> <message>
# Refuse the tool call outright; <message> reaches the agent. Deterministic in
# every mode, interactive or headless. RC_DRY_RUN=1 downgrades it to a log.
rc_deny() {
    if [ "${RC_DRY_RUN:-0}" = "1" ]; then
        printf 'rc %s (dry-run, would deny): %s\n' "$1" "$2" >&2
        exit 0
    fi
    rc__decision "$1" "deny" "$2"
}

# rc_ask <hook-name> <message>
# Force a permission prompt the user must answer — it fires even when the command
# is allowlisted or the session is in auto mode. Use for operations that are
# legitimate but must not happen silently.
#
# CAVEAT: what "ask" does under `claude -p` (headless, nobody to answer) is not
# documented. Do not rely on it as a hard stop in CI — reach for rc_deny when a
# call must never proceed without a human, and prefer rc_ask only where the
# operation is recoverable if it slips through.
rc_ask() {
    if [ "${RC_DRY_RUN:-0}" = "1" ]; then
        printf 'rc %s (dry-run, would ask): %s\n' "$1" "$2" >&2
        exit 0
    fi
    rc__decision "$1" "ask" "$2"
}
