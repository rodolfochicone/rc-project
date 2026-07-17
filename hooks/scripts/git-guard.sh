#!/usr/bin/env bash
# PreToolUse(Bash): gate destructive / history-rewriting git commands.
#
# Two tiers, both emitted as JSON on exit 0 (never exit 2 — Claude Code ignores
# stdout on exit 2, and the two signalling styles must not be mixed in one hook):
#
#   deny — data loss with no cheap undo (reset --hard, restore, clean, filter-branch,
#          bare force-push). Deterministic in every mode, headless included.
#   ask  — legitimate, recoverable, but must never happen silently (rebase,
#          --force-with-lease). Fires a permission prompt even when the command is
#          allowlisted or the session is in auto mode.
#
# Why rebase is `ask` and not `deny`: rc-git ships a rebase-and-resolve-conflicts
# workflow, and a blanket deny made every command that skill teaches unrunnable —
# the plugin blocked its own feature. Rebase is recoverable via reflog, so it earns
# a prompt rather than a refusal. Bare `--force` stays denied: `--force-with-lease`
# is the same operation with a safety catch, so the answer is always to use it.
#
# CAVEAT: "ask" under `claude -p` (nobody to answer) is not documented upstream.
# Anything that must never proceed unattended belongs in the deny tier.
#
# Active from the "minimal" profile up (always), disablable via
# RC_DISABLED_HOOKS=git-guard, and downgraded to a log by RC_DRY_RUN=1.
# Fail-open: no jq, no command, or no _lib.sh lets the call through.
#
# Run with --selftest to check the matching offline.
set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"

decide() {
    norm=$(printf '%s' "$1" | tr '\n' ' ' | tr -s '[:space:]' ' ')

    case "$norm" in
    *"git reset --hard"*) printf 'deny\t%s' "'git reset --hard' discards committed/working changes. Forbidden without explicit user approval." ; return ;;
    *"git restore"*) printf 'deny\t%s' "'git restore' discards working-tree changes. Forbidden without explicit user approval." ; return ;;
    *"git clean"*) printf 'deny\t%s' "'git clean' deletes untracked files. Forbidden without explicit user approval." ; return ;;
    *"git checkout -- "* | *"git checkout ."*) printf 'deny\t%s' "'git checkout' that discards changes is forbidden. Use a non-destructive alternative or ask the user." ; return ;;
    *"git filter-branch"*) printf 'deny\t%s' "'git filter-branch' rewrites history irreversibly. Forbidden." ; return ;;
    esac

    # Force-push: --force-with-lease refuses when the remote moved, so it is the
    # form to reach for; bare --force / -f clobbers whatever landed since fetch.
    #
    # O flag tem que ser argumento DO push, não de qualquer coisa depois dele. Casar
    # o comando inteiro dava FP: `git push origin main && [ -f arquivo ]` acusava
    # force-push, porque o teste de arquivo do bash também contém " -f". Então
    # recorta-se o trecho a partir do `git push` até o próximo separador (`;`, `&&`,
    # `||`, `|`) e só ele é inspecionado.
    case "$norm" in
    *"git push"*)
        push_seg=${norm#*git push}
        push_seg=${push_seg%%;*}
        push_seg=${push_seg%%&&*}
        push_seg=${push_seg%%||*}
        push_seg=${push_seg%%|*}
        case "$push_seg" in
        *"--force-with-lease"*) printf 'ask\t%s' "force-push with lease rewrites the remote branch. Confirm this is your own branch and that no one else builds on it." ; return ;;
        *"--force"* | *" -f"*) printf 'deny\t%s' "bare force-push overwrites whatever landed on the remote since your last fetch. Use 'git push --force-with-lease', which refuses when the remote moved." ; return ;;
        esac
        ;;
    esac

    case "$norm" in
    *"git rebase"*) printf 'ask\t%s' "rebase rewrites history on this branch. Confirm the branch is yours and unshared; recover with 'git reflog' if it goes wrong." ; return ;;
    esac
}

if [ "${1:-}" = "--selftest" ]; then
    fail=0
    check() { # <cmd> <expected-verdict>
        got=$(decide "$1" | cut -f1)
        [ -z "$got" ] && got="allow"
        if [ "$got" != "$2" ]; then printf 'FAIL  %-46s expected=%s got=%s\n' "$1" "$2" "$got"; fail=1
        else printf 'ok    %-46s %s\n' "$1" "$got"; fi
    }
    check "git reset --hard HEAD~1" deny
    check "git clean -fd" deny
    check "git filter-branch --tree-filter x" deny
    check "git push --force origin main" deny
    check "git push -f origin main" deny
    check "git push --force-with-lease origin feat" ask
    check "git rebase -i main" ask
    check "git rebase --continue" ask
    check "git status" allow
    check "git push origin feat" allow
    check "git commit -m 'x'" allow
    # regressão: o flag tem que ser argumento DO push. Estes casavam o antigo
    # `*"git push"*" -f"*` e eram negados por engano — o segundo bloqueou o commit
    # que consertava este arquivo.
    check 'git push origin main && find . -type f' allow
    check 'git push origin main && [ -f arquivo ]' allow
    check 'git push origin main; [ -f x ] && echo ok' allow
    check 'git push -f origin main && [ -f x ]' deny
    exit $fail
fi

rc_hook_active "git-guard" "minimal" || exit 0

input=$(cat)
command -v jq >/dev/null 2>&1 || exit 0

cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty')
[ -z "$cmd" ] && exit 0

verdict=$(decide "$cmd")
[ -z "$verdict" ] && exit 0

what=$(printf '%s' "$verdict" | cut -f1)
why=$(printf '%s' "$verdict" | cut -f2-)

case "$what" in
deny) rc_deny "git-guard" "$why" ;;
ask) rc_ask "git-guard" "$why" ;;
esac

exit 0
