# Boilerplate: Axum + SvelteKit + Postgres + Caddy + systemd

Deploy fullstack na VPS (Ubuntu 24.04), alinhado à análise em
[`../stack-vps-fullstack-rust-typescript.md`](../stack-vps-fullstack-rust-typescript.md).

| Peça | Função |
|------|--------|
| **Axum** | API REST + WebSocket real-time (`:3000`) |
| **SvelteKit** (adapter-node + **Bun**) | SSR + SEO + dashboard (`:3001`) |
| **Postgres 16** | Docker, bind só em `127.0.0.1:5432` |
| **Caddy** | TLS + reverse proxy (API, WS, front) |
| **systemd** | `app-api` e `app-web` (web via Bun) |

## Layout na VPS

```
Internet
   │
 Caddy :443/:80
   │
   ├─ /ws/* /api/* /health  → 127.0.0.1:3000  (Axum)
   └─ /*                    → 127.0.0.1:3001  (SvelteKit via Bun)

/opt/app/
  .env
  bin/app-api
  web/          ← output do adapter-node (rodado com Bun)
```

## Estrutura deste diretório

```
boilerplate-axum-sveltekit-vps/
├── backend/           # Rust + Axum + SQLx
├── frontend/          # SvelteKit SSR + dashboard WS (Bun)
├── deploy/
│   ├── Caddyfile
│   └── systemd/
│       ├── app-api.service
│       └── app-web.service
├── scripts/
│   ├── install-vps.sh # root: docker, caddy, bun, rust, user deploy
│   ├── deploy.sh      # build release + copia para /opt/app
│   └── dev.sh         # dev local
├── docker-compose.yml # só Postgres
└── .env.example
```

## 1. Desenvolvimento local

Pré-requisitos: Docker, Rust, **Bun ≥ 1.3**.

```bash
bun --version   # ex.: 1.3.14
# se desatualizado: bun upgrade

cd docs/boilerplate-axum-sveltekit-vps
cp .env.example .env
# opcional: bash scripts/dev.sh

docker compose up -d postgres

# terminal 1
cd backend && cargo run

# terminal 2
cd frontend && bun install && bun run dev
```

- Front: http://127.0.0.1:5173 (proxy de `/api` e `/ws` no Vite)
- API: http://127.0.0.1:3000/health
- Dashboard: http://127.0.0.1:5173/dashboard

## 2. Provisionar a VPS (uma vez)

Como root na Hostinger (ou similar):

```bash
# copie este boilerplate para a máquina, ex.:
# scp -r docs/boilerplate-axum-sveltekit-vps root@SEU_IP:/tmp/app-src

cd /tmp/app-src   # ou onde estiver o boilerplate
bash scripts/install-vps.sh
```

O script instala Docker, Caddy, **Bun** (user `deploy`), Rust, abre UFW (22/80/443) e prepara `/opt/app`.

## 3. Configurar e publicar

```bash
# como root ou deploy com sudo
sudo mkdir -p /opt/app/src
sudo rsync -a ./ /opt/app/src/
sudo chown -R deploy:deploy /opt/app

sudo -u deploy cp /opt/app/src/.env.example /opt/app/.env
sudo -u deploy nano /opt/app/.env
```

Edite no mínimo:

- `DOMAIN=seu-dominio.com` (DNS A apontando para o IP da VPS)
- `POSTGRES_PASSWORD=...`
- `DATABASE_URL=postgres://app:SENHA@127.0.0.1:5432/app`

```bash
sudo -u deploy bash /opt/app/src/scripts/deploy.sh

sudo cp /opt/app/src/deploy/systemd/app-api.service /etc/systemd/system/
sudo cp /opt/app/src/deploy/systemd/app-web.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now app-api app-web

# Caddy: injeta DOMAIN a partir do .env
set -a && source /opt/app/.env && set +a
sudo cp /opt/app/src/deploy/Caddyfile /etc/caddy/Caddyfile
# se o Caddyfile usar {$DOMAIN}, exporte DOMAIN no serviço caddy
# ou substitua o host manualmente no arquivo:
#   sed -i "s/\${DOMAIN:localhost}/$DOMAIN/" ...
echo "DOMAIN=$DOMAIN" | sudo tee /etc/caddy/env
sudo mkdir -p /etc/systemd/system/caddy.service.d
printf '[Service]\nEnvironmentFile=/opt/app/.env\n' | sudo tee /etc/systemd/system/caddy.service.d/env.conf
sudo systemctl daemon-reload
sudo systemctl enable --now caddy
sudo systemctl reload caddy
```

Verifique:

```bash
systemctl status app-api app-web caddy
curl -sS http://127.0.0.1:3000/health
curl -sS http://127.0.0.1:3001/
curl -sSI https://seu-dominio.com/
```

## 4. Redeploy (código novo)

```bash
cd /opt/app/src
sudo -u deploy git pull   # se versionado
sudo -u deploy bash scripts/deploy.sh
# deploy.sh já reinicia app-api e app-web
```

## Endpoints da API (Axum)

| Método | Path | Descrição |
|--------|------|-----------|
| GET | `/health` | liveness |
| GET | `/api/hello` | exemplo JSON (usado no SSR) |
| GET | `/api/metrics` | últimas métricas |
| POST | `/api/metrics` | `{ "name": "...", "value": 1.0 }` |
| WS | `/ws/metrics` | samples a cada 2s (dashboard) |

## Variáveis de ambiente

| Var | Quem usa | Notas |
|-----|----------|--------|
| `DATABASE_URL` | API | obrigatória em prod |
| `API_HOST` / `API_PORT` | API | default `127.0.0.1:3000` |
| `API_INTERNAL_URL` | SvelteKit SSR | `http://127.0.0.1:3000` |
| `PUBLIC_API_URL` | browser (opcional) | vazio = same-origin via Caddy |
| `PUBLIC_WS_URL` | browser (opcional) | vazio = `wss://host/ws/metrics` |
| `DOMAIN` | Caddy | hostname para TLS |
| `PORT` / `HOST` | SvelteKit | `3001` / `127.0.0.1` |

## Segurança (mínimo)

- API e web escutam só em **127.0.0.1**; só o Caddy expõe 80/443.
- Postgres bind **127.0.0.1:5432** — não abra 5432 no UFW.
- Troque senhas do `.env`; não commite `.env`.
- Serviços rodam como usuário `deploy` (sem root).
- Opcional: `fail2ban` (instalado no `install-vps.sh`).

## Multi-app na mesma VPS

Duplique units com nomes diferentes (`app2-api.service`), portas (`3002`/`3003`) e um bloco extra no `Caddyfile` por domínio. Postgres pode ser compartilhado com databases distintos.

## Troubleshooting

| Sintoma | Checagem |
|---------|----------|
| 502 no domínio | `systemctl status app-web app-api`; portas 3000/3001 |
| WS não conecta | Caddy `handle /ws/*` antes do catch-all; browser `wss://` |
| SSR sem dados | `API_INTERNAL_URL=http://127.0.0.1:3000` no `.env` do web |
| API não sobe | `journalctl -u app-api -n 50`; Postgres up? `DATABASE_URL`? |
| TLS falha | DNS A correto? porta 80 aberta? `journalctl -u caddy` |

## Licença

Mesma do repositório pai (uso como template livre).
