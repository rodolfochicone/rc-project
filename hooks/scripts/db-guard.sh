#!/usr/bin/env bash
# PreToolUse(Bash): rc db-guard — Regra 9 — acesso a banco é read-only por padrão.
# Bloqueia comandos de cliente de banco (psql/mysql/etc.) contendo SQL de escrita
# ou DDL. Exit 2 bloqueia a chamada e devolve a mensagem ao agente.
# Fail-open: sem jq, sem comando, ou sem _lib.sh, deixa passar.
#
# Ativo do perfil "minimal" pra cima (sempre), e desligável via
# RC_DISABLED_HOOKS=db-guard. Passa pelo rc_block do _lib.sh para honrar
# RC_DRY_RUN=1 — sem isso os escapes que o _lib.sh anuncia não alcançam este hook.
set -u
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$SCRIPT_DIR/_lib.sh"
rc_hook_active "db-guard" "minimal" || exit 0

input=$(cat)
command -v jq >/dev/null 2>&1 || exit 0

cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty')
[ -z "$cmd" ] && exit 0

case "$cmd" in
*psql* | *mysql* | *sqlite3* | *"pg_dump"*) ;;
*) exit 0 ;;
esac

if printf '%s' "$cmd" | grep -qiE '\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE)\b'; then
    rc_block "db-guard" "acesso a banco é read-only por padrão (Regra 9). Comando com escrita/DDL detectado. Recomende o SQL ao usuário e execute apenas com autorização explícita."
fi

exit 0
