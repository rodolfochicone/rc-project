---
name: rc-fullstack-axum-svelte
description: Routes fullstack Axum + SQLx/Postgres + SvelteKit work to the right specialist skills (rc-axum, rc-sqlx, rc-sveltekit), covering API, data layer, SSR front, security, tests, and VPS deploy with Bun. Use when building or reviewing the Rust/SvelteKit stack together, scaffolding fullstack features, or unsure which skill to load. Do not use for React/Next stacks, pure Python, or workflow PRD/task pipeline phases.
user-invocable: true
model: sonnet
effort: medium
---

# Fullstack Axum + SvelteKit (umbrella)

Single entry point for the **Rust API + Postgres + SvelteKit SSR** stack used on VPS deploys. This skill **does not** replace the specialists — it **loads them** and coordinates workflow, runtime/package manager, and verify gates.

## Stack (pinned targets)

| Layer | Tech | Skill |
| ----- | ---- | ----- |
| HTTP API / WS | Axum **0.8+**, Tokio, Tower | `rc-axum` |
| Data | SQLx **0.8+**, PostgreSQL | `rc-sqlx` |
| Front SSR/SEO | SvelteKit **2**, Svelte **5**, adapter-node | `rc-sveltekit` |
| JS runtime + package manager | **Bun ≥ 1.3** (install + scripts + SSR process) | this skill |
| Deploy reference | Caddy + systemd + Postgres | `docs/boilerplate-axum-sveltekit-vps/` |

Verified local target: **Bun 1.3.x** (e.g. 1.3.14). Prefer current stable (`bun --version`; upgrade with `bun upgrade`).

## Runtime & package manager: Bun only

For all frontend work in this stack, use **Bun** — not Node, npm, yarn, or pnpm (unless the user explicitly overrides).

| Action | Command |
| ------ | ------- |
| Install | `bun install` |
| Dev | `bun run dev` |
| Build | `bun run build` |
| Check | `bun run check` (if script exists) |
| Test | `bun test` or `bun run test` |
| Lint | `bun run lint` |
| Add dep | `bun add <pkg>` / `bun add -d <pkg>` |
| Start SSR (prod) | `bun run start` → runs adapter-node output via Bun |

Lockfile: commit **`bun.lock`** (or `bun.lockb` on older Bun). Do not introduce `package-lock.json` / `yarn.lock` for this stack.

Backend remains **Cargo** (`cargo build`, `cargo test`, `cargo clippy`).

SvelteKit still uses **`@sveltejs/adapter-node`** for the server build; **run that output with Bun** (`bun ./build/index.js` or the `start` script). That keeps SSR portable while using Bun as the process runtime.

## Routing — which skill to load

Read the matching skill’s `SKILL.md` and the referenced files under that skill’s `references/` **before** implementing.

| Situation | Load |
| --------- | ---- |
| Handlers, Router, middleware, WS, CORS, clippy | **`rc-axum`** (all of `references/`) |
| Queries, pool, migrations, binds, DB tests | **`rc-sqlx`** |
| Routes, SSR load, forms, hooks, CSP, adapter-node | **`rc-sveltekit`** |
| End-to-end feature (API + DB + page) | **all three** specialists, in order below |
| A11y on UI | + `rc-a11y` |
| SEO meta / technical SEO | + `rc-seo` |
| SQL indexes / EXPLAIN depth | + `rc-sql` |
| Logs/metrics/SLOs | + `rc-observability` |

## End-to-end feature workflow

1. **Contract** — path, method, payload, auth, SSR vs client data, real-time needs.
2. **Schema** — migration via `rc-sqlx` (`references/migrations.md`).
3. **API** — routes + errors + tests via `rc-axum`.
4. **Data access** — repo functions via `rc-sqlx` (binds only).
5. **Front** — `+page.server.ts` / actions / UI via `rc-sveltekit`; install/run with **Bun**.
6. **Security pass** — each skill’s `references/security.md` for touched layers.
7. **Verify** — run gates below; then `rc-final-verify` before claiming done.

## Verify gates (default)

```bash
# API
cargo fmt --all -- --check
cargo clippy --all-targets -- -D warnings
cargo test
cargo build --release

# Front (from frontend/ or monorepo package root)
bun --version   # expect 1.3.x+
bun install
bun run check   # if present
bun run test    # if present
bun run build
```

Deploy path: see `docs/boilerplate-axum-sveltekit-vps/README.md` (scripts and systemd must use Bun).

## Architecture defaults

```
Browser ──► Caddy (TLS)
              ├─ /api /ws /health ──► Axum :3000 (127.0.0.1)
              └─ /* ──────────────► SvelteKit (Bun) :3001
Postgres ◄── Axum (DATABASE_URL)
SSR load ──► API_INTERNAL_URL=http://127.0.0.1:3000
```

- API and web bind **loopback**; only Caddy is public.
- Prefer same-origin `/api` and `/ws` so the browser needs no loose CORS.
- Secrets: env / `.env` never `PUBLIC_*`.

## Must do

- Activate specialist skills by domain; do not invent Axum/SQLx/Kit patterns that contradict their guides.
- Use **Bun** for JS install, scripts, and the production SSR process in this stack.
- Keep handlers thin; SQL in a db/repo module; server-only code out of client bundles.
- Run security checklists when touching auth, cookies, WS, or raw SQL.

## Must not do

- Use Node/npm/yarn/pnpm for this stack’s frontend by default.
- Skip migrations for schema changes.
- Put `DATABASE_URL` or private keys in SvelteKit public env.
- Claim done without cargo + bun verify evidence (`rc-final-verify`).

## Output template (fullstack change)

1. Migration (if any) + API routes/types
2. SQLx access layer
3. SvelteKit routes (SSR/actions) + Bun commands used
4. Security notes (auth, cookies, binds)
5. Commands run and results (cargo + bun)

## Related

- Specialists: `rc-axum`, `rc-sqlx`, `rc-sveltekit`
- Boilerplate: `docs/boilerplate-axum-sveltekit-vps/`
- Stack analysis: `docs/stack-vps-fullstack-rust-typescript.md`
