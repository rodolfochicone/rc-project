#!/usr/bin/env bash
# Sobe Postgres + API + front em modo dev (máquina local ou VPS).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -f .env ]]; then
  cp .env.example .env
  echo "Criado .env a partir de .env.example"
fi

set -a
# shellcheck disable=SC1091
source .env
set +a

docker compose up -d postgres

echo "Aguardando Postgres..."
for i in {1..30}; do
  if docker compose exec -T postgres pg_isready -U "${POSTGRES_USER:-app}" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

echo "Inicie em dois terminais:"
echo "  cd backend && cargo run"
echo "  cd frontend && bun install && bun run dev"
echo
echo "Ou em background:"

(
  cd backend
  cargo run
) &
API_PID=$!

(
  cd frontend
  bun install
  bun run dev
) &
WEB_PID=$!

trap 'kill $API_PID $WEB_PID 2>/dev/null || true' INT TERM
wait
