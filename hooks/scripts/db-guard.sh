#!/usr/bin/env bash
# PreToolUse(Bash): rc db-guard — Regra 9 — acesso a banco é read-only por padrão.
# Bloqueia comandos de cliente de banco (psql/mysql/etc.) contendo SQL de escrita
# ou DDL. Exit 2 bloqueia a chamada e devolve a mensagem ao agente.
# Fail-open: sem jq ou sem comando, deixa passar.
set -u

input=$(cat)
command -v jq >/dev/null 2>&1 || exit 0

cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty')
[ -z "$cmd" ] && exit 0

case "$cmd" in
*psql* | *mysql* | *sqlite3* | *"pg_dump"*) ;;
*) exit 0 ;;
esac

if printf '%s' "$cmd" | grep -qiE '\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|CREATE|GRANT|REVOKE)\b'; then
    printf 'rc db-guard: acesso a banco é read-only por padrão (Regra 9). Comando com escrita/DDL detectado. Recomende o SQL ao usuário e execute apenas com autorização explícita.\n' >&2
    exit 2
fi

exit 0
