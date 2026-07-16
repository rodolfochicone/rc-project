#!/usr/bin/env bash
# Provisiona dependências na VPS Ubuntu 24.04 (rodar como root uma vez).
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "Rode como root: sudo bash scripts/install-vps.sh"
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive

apt-get update
apt-get install -y \
  ca-certificates curl gnupg lsb-release \
  build-essential pkg-config libssl-dev \
  ufw fail2ban

# Docker (Postgres)
if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sh
fi
systemctl enable --now docker

# Caddy
if ! command -v caddy >/dev/null 2>&1; then
  apt-get install -y debian-keyring debian-archive-keyring apt-transport-https
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
    | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
  curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
    | tee /etc/apt/sources.list.d/caddy-stable.list
  apt-get update
  apt-get install -y caddy
fi

# Bun (SvelteKit install + SSR runtime) — for user deploy
if ! id deploy >/dev/null 2>&1; then
  useradd --create-home --shell /bin/bash deploy
  usermod -aG docker deploy
fi

if [[ ! -x /home/deploy/.bun/bin/bun ]]; then
  sudo -u deploy bash -lc 'curl -fsSL https://bun.sh/install | bash'
fi
# Optional system-wide symlink for scripts that look on PATH
if [[ -x /home/deploy/.bun/bin/bun && ! -x /usr/local/bin/bun ]]; then
  ln -sf /home/deploy/.bun/bin/bun /usr/local/bin/bun
fi
echo "Bun: $(sudo -u deploy bash -lc 'bun --version' 2>/dev/null || bun --version 2>/dev/null || echo missing)"

# Rust (para o usuário deploy)
if [[ ! -x /home/deploy/.cargo/bin/cargo ]]; then
  sudo -u deploy bash -lc 'curl --proto "=https" --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y'
fi

mkdir -p /opt/app/bin /opt/app/web
chown -R deploy:deploy /opt/app

# Firewall
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

echo
echo "OK. Próximos passos:"
echo "  1) Copie o projeto para /opt/app/src (ou clone o repo)"
echo "  2) cp .env.example /opt/app/.env  e edite as senhas/domínio"
echo "  3) bash scripts/deploy.sh"
echo "  4) instale units: sudo cp deploy/systemd/*.service /etc/systemd/system/"
echo "  5) sudo systemctl daemon-reload && sudo systemctl enable --now app-api app-web"
echo "  6) sudo cp deploy/Caddyfile /etc/caddy/Caddyfile && sudo systemctl reload caddy"
