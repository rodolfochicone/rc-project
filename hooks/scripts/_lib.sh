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
# one. git/commit guards are "minimal" (always on); formatting and the fact-gate
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
