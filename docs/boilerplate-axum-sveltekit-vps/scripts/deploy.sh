#!/usr/bin/env bash
# Build local ou na VPS e publica em /opt/app.
# Uso:
#   APP_ROOT=/opt/app bash scripts/deploy.sh
#   # ou, a partir da raiz do boilerplate:
#   bash scripts/deploy.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_ROOT="${APP_ROOT:-/opt/app}"
ENV_FILE="${ENV_FILE:-$APP_ROOT/.env}"

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
elif [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

echo "==> Postgres (docker compose)"
cd "$ROOT"
if command -v docker >/dev/null 2>&1; then
  docker compose --env-file "${ENV_FILE:-$ROOT/.env.example}" up -d postgres
else
  echo "docker não encontrado — pulei o Postgres"
fi

echo "==> Build Axum (release)"
cd "$ROOT/backend"
cargo build --release
mkdir -p "$APP_ROOT/bin"
cp -f target/release/app-api "$APP_ROOT/bin/app-api"
chmod +x "$APP_ROOT/bin/app-api"

echo "==> Build SvelteKit (Bun)"
cd "$ROOT/frontend"
if command -v bun >/dev/null 2>&1; then
  echo "    bun $(bun --version)"
  bun install --frozen-lockfile 2>/dev/null || bun install
  bun run build
  mkdir -p "$APP_ROOT/web"
  rsync -a --delete build/ "$APP_ROOT/web/"
else
  echo "bun não encontrado — pulei o front (instale: curl -fsSL https://bun.sh/install | bash)"
fi

if [[ -f "$ROOT/.env" && ! -f "$APP_ROOT/.env" ]]; then
  cp "$ROOT/.env" "$APP_ROOT/.env"
fi

chown -R deploy:deploy "$APP_ROOT" 2>/dev/null || true

echo "==> Restart services (se existirem)"
if command -v systemctl >/dev/null 2>&1; then
  sudo systemctl restart app-api 2>/dev/null || true
  sudo systemctl restart app-web 2>/dev/null || true
fi

echo
echo "Deploy em $APP_ROOT concluído."
echo "  API bin: $APP_ROOT/bin/app-api"
echo "  Web:     $APP_ROOT/web"
