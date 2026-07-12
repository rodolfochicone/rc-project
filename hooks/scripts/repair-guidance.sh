#!/usr/bin/env bash
# PostToolUse(Edit|Write|MultiEdit|Task): repair-guidance hook. When a file edit
# or a delegated task comes back with a *repairable* failure, feed the agent one
# concrete piece of corrective guidance — via exit 2 + stderr, the same channel
# rc_block uses — so it fixes the root cause instead of resending the identical
# failing call. This targets the most expensive LLM failure mode: burning turns
# retrying a byte-off old_string or a malformed delegation.
#
# Active from the "standard" profile up. Fails open on any environment problem
# (no jq, empty input). Never touches files. Run with --selftest to validate the
# pattern matching offline.
#
# VERIFY ONCE PER CLAUDE CODE VERSION: this assumes PostToolUse fires with the
# tool's error text reachable at `.tool_response` (string) or, when it is an
# object, in an `.error`/`.message`/`.errorMessage` field. It also reads
# `.tool_result` as a fallback for older builds. We deliberately do NOT scan a
# serialized object as a whole: a *successful* Edit response is an object that
# embeds the edited file's content (`originalFile`/`structuredPatch`), so
# grepping the whole object matched failure phrases that merely appear in the
# file — firing "did not apply" on every successful edit (false positive). Only
# the status/error text is inspected. Quick check: `claude --debug`, force a
# failing Edit (bad old_string), and confirm this hook runs; then edit a file
# that literally contains "no changes"/"not found" and confirm it does NOT.
set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"

EDIT_GUIDANCE="The %s did not apply. The root cause is almost always old_string not matching the file byte-for-byte (whitespace, indentation, quotes, or the text drifted). Re-Read the exact target lines, copy the literal text, make old_string uniquely identify ONE location, and retry. Do NOT resend the same old_string."
TASK_GUIDANCE="The delegated task failed. Before retrying: (1) confirm subagent_type is a valid agent name; (2) make the prompt self-contained — objective, target files/scope, constraints, expected output, whether edits are allowed; (3) supply any required args named in the error above. Fix the specific error; do NOT resend an identical call."

# --- offline self-test --------------------------------------------------------
if [ "${1:-}" = "--selftest" ]; then
    fail=0
    run() { printf '%s' "$2" | RC_HOOK_PROFILE=standard "$0"; }
    check() { # <label> <expected-rc> <json>
        out=$(run "$1" "$3" 2>&1)
        rc=$?
        if [ "$rc" = "$2" ]; then
            printf 'ok   %s (rc=%s)\n' "$1" "$rc"
        else
            printf 'FAIL %s (want rc=%s got rc=%s: %s)\n' "$1" "$2" "$rc" "$out"
            fail=1
        fi
    }
    check "edit-miss-string" 2 '{"tool_name":"Edit","tool_response":"String to replace not found in file"}'
    check "edit-multiple"    2 '{"tool_name":"MultiEdit","tool_response":"Found 3 matches for the provided old_string"}'
    check "task-unknown"     2 '{"tool_name":"Task","tool_response":{"error":"unknown agent: foo"}}'
    check "edit-ok"          0 '{"tool_name":"Edit","tool_response":"The file has been updated."}'
    # regression: successful edit object embeds file content with trigger phrases.
    # The whole-object grep used to fire "did not apply" on every such success.
    check "edit-ok-object"   0 '{"tool_name":"Edit","tool_response":{"filePath":"x","originalFile":"docs: string to replace not found in file; no changes needed; old_string","structuredPatch":[]}}'
    check "task-ok-object"   0 '{"tool_name":"Task","tool_response":{"content":"I handled the error and the failed path gracefully"}}'
    check "other-tool"       0 '{"tool_name":"Read","tool_response":"error: irrelevant"}'
    check "empty"            0 '{}'
    exit $fail
fi
# ------------------------------------------------------------------------------

rc_hook_active "repair-guidance" "standard" || exit 0
command -v jq >/dev/null 2>&1 || exit 0

input=$(cat)
tool=$(printf '%s' "$input" | jq -r '.tool_name // empty' 2>/dev/null)
# Extract ONLY the status/error text. A string response is the status message
# as-is; an object response exposes its failure in .error/.message/.errorMessage
# (a *successful* edit object has none of these, so resp is empty and we exit).
# We never serialize the whole object: it embeds the edited file's content.
resp=$(printf '%s' "$input" | jq -r '
    (.tool_response // .tool_result) as $r
    | if $r == null then ""
      elif ($r | type) == "string" then $r
      else (($r.error // $r.message // $r.errorMessage // "")
            | if type == "string" then . else tojson end)
      end' 2>/dev/null)
[ -z "$tool" ] && exit 0
[ -z "$resp" ] && exit 0

case "$tool" in
Edit | Write | MultiEdit)
    printf '%s' "$resp" | grep -qiE 'string to replace not found|not found in file|does not appear|found [0-9]+ (match|occurrence)|multiple (match|occurrence)|no changes|has been modified since|could not apply|cannot (apply|edit)|old_string' || exit 0
    # shellcheck disable=SC2059
    rc_block "repair-guidance" "$(printf "$EDIT_GUIDANCE" "$tool")"
    ;;
Task)
    printf '%s' "$resp" | grep -qiE 'error|failed|unknown agent|not allowed|must be one of|invalid|no such|missing required' || exit 0
    rc_block "repair-guidance" "$TASK_GUIDANCE"
    ;;
*)
    exit 0
    ;;
esac
