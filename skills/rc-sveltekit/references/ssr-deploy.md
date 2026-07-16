# SvelteKit — SSR deploy (adapter-node + Bun / VPS)

## Runtime

- **Bun ≥ 1.3** for `install`, `run`, and the production process.
- Build still uses `@sveltejs/adapter-node`; start the output with Bun (not Node).

```bash
bun --version   # expect 1.3.x+
bun upgrade     # when behind latest stable
```

## Adapter

```bash
bun add -d @sveltejs/adapter-node
```

```js
// svelte.config.js
import adapter from '@sveltejs/adapter-node';

export default {
	kit: {
		adapter: adapter({ out: 'build' })
	}
};
```

Build output: `build/` (run with `bun ./build/index.js` or `bun run start` — check generated entry; do not require system Node for prod).

## Environment (production)

| Var | Example | Purpose |
| --- | ------- | ------- |
| `HOST` | `127.0.0.1` | Do not expose Node publicly |
| `PORT` | `3001` | Behind Caddy |
| `ORIGIN` | `https://example.com` | Canonical origin (CSRF/cookies) when required |
| `PROTOCOL_HEADER` | `x-forwarded-proto` | If Kit configured for proxy |
| `HOST_HEADER` | `x-forwarded-host` | If Kit configured for proxy |
| `API_INTERNAL_URL` | `http://127.0.0.1:3000` | SSR → Axum |
| `NODE_ENV` | `production` | |

Set `ORIGIN` (or proxy headers) correctly so CSRF and absolute URLs work behind Caddy.

## Reverse proxy (Caddy sketch)

```
example.com {
  handle /ws/* { reverse_proxy 127.0.0.1:3000 }
  handle /api/* { reverse_proxy 127.0.0.1:3000 }
  handle /health { reverse_proxy 127.0.0.1:3000 }
  handle { reverse_proxy 127.0.0.1:3001 }
}
```

## systemd

- `WorkingDirectory` = deployed web root
- `EnvironmentFile=/opt/app/.env`
- `ExecStart=/usr/local/bin/bun /opt/app/web/index.js` (symlink from install; avoid `$HOME/.bun` when `ProtectHome=true`)
- Restart on failure; run as non-root user
- Install Bun for the deploy user, then expose a system path the unit can execute

## Performance

- One Node process is often enough for solo apps; scale with multiple processes only if CPU-bound SSR.
- Enable proxy compression (Caddy `encode`).
- Cache static assets aggressively at the proxy; SSR HTML carefully.

## SEO

- Real content in SSR HTML for public routes
- `<svelte:head>` titles/descriptions per page
- `noindex` on authenticated dashboards when appropriate

## Build commands

```bash
bun install --frozen-lockfile
bun run build
bun run start
# or: bun ./build/index.js
```

## Pair with boilerplate

See `docs/boilerplate-axum-sveltekit-vps/` for a full Caddy + systemd + Bun layout.
