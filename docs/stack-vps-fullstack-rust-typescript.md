# Stack fullstack em VPS: Rust, TypeScript e alternativas

Análise de stacks para projetos fullstack e real-time em VPS, com foco em performance, menor custo e menor consumo de CPU/RAM. Contexto do autor: solo, SSR/SEO, dashboards ao vivo, monólito ou front separado conforme o caso.

## Perfil da VPS

| Recurso | Hostinger KVM 4 |
|--------|------------------|
| CPU | 4 vCPU |
| RAM | 16 GB |
| Disco | 200 GB |
| Banda | 16 TB |
| SO | Ubuntu 24.04 LTS |
| Local | Brazil – Campinas |

Com 16 GB a máquina **não é apertada**: roda várias stacks “pesadas” (Next + Postgres + Redis) e ainda sobra. Sendo **solo**, o gargalo costuma ser **velocidade de entrega e manutenção**, não a máquina. Ainda assim, stack enxuta deixa margem para vários projetos na mesma VPS.

## Critérios (ordem de prioridade)

1. **SSR + SEO** (páginas públicas indexáveis)
2. **Real-time** (dashboards ao vivo: WebSocket / SSE / LiveView)
3. **Performance + baixo consumo** (vários apps na mesma VPS)
4. **Velocidade solo** (pouco tempo pelejando com build/tooling)

---

## Rust vs TypeScript

| Critério | **Rust** (Axum/Actix + front) | **TypeScript** (Node/Bun + Next/SvelteKit) |
|----------|-------------------------------|--------------------------------------------|
| Throughput / latência HTTP | Excelente (topo TechEmpower com frameworks Rust/Go) | Bom o suficiente; Node fica atrás de Rust/Go em RPS sintético |
| Memória por processo | Muito baixa (binário nativo, sem GC pesado) | Mais alta; Bun ~25–40% menos que Node |
| Conexões WebSocket | Muito eficiente | OK; mais RAM por conexão que Rust/Go/BEAM |
| SSR/SEO maduro | Existe (Leptos), mas ecossistema menor | **Muito maduro** (Next, SvelteKit, Nuxt) |
| Real-time dashboard | Manual (WS + front) | Ecossistema rico (Socket.IO, tRPC, etc.) |
| DX solo | Compilação lenta, curva alta, menos libs “prontas” | **Muito mais rápido** de entregar |
| Custo da VPS | Cabe dezenas de serviços leves | Cabe vários apps; Next SSR usa mais RAM |
| Risco solo | Travamento em borrow-checker / tooling | Baixo — ecossistema enorme |

**Resumo**

- Máxima eficiência de recurso → Rust (e Go) vencem.
- Fullstack + SSR/SEO + dashboards ao vivo sendo solo → TypeScript (ou Elixir) entregam mais produto por hora.
- Em 16 GB, a diferença de RAM raramente justifica 2–3× o tempo de dev do Rust full-stack.

---

## Combinações com Rust no backend

### 1) Recomendada (híbrida): Axum + SvelteKit (SSR)

| Camada | Stack |
|--------|--------|
| API / real-time | **Rust + Axum** (Tokio) + WebSocket ou SSE |
| Front SSR/SEO | **SvelteKit** (adapter-node na VPS) ou **Next.js** se preferir React |
| DB | PostgreSQL (+ Redis se precisar de pub/sub / cache) |
| Proxy | Caddy ou Nginx |
| Deploy | 1 binário Rust + 1 processo Node/Bun do SvelteKit |

**Por quê**

- SSR/SEO no front maduro (SvelteKit/Next).
- Backend enxuto e rápido para APIs e streams de dashboard.
- Padrão já usado em produção (Axum + SvelteKit).
- SvelteKit costuma ser **mais leve** que Next no VPS.

**Monólito vs separado**

- **Separado (maioria dos casos):** API Rust em `:3000`, SvelteKit em `:3001`, Nginx/Caddy na frente.
- **“Monólito de deploy”:** build do SvelteKit → Axum serve `static/` + API no mesmo binário (um só serviço systemd). Bom para app pequeno.

**Real-time:** WebSocket no Axum; o dashboard Svelte consome e atualiza charts. SSE se for só push unidirecional (mais simples e barato).

### 2) Full Rust: Leptos (SSR) + Axum

| Camada | Stack |
|--------|--------|
| Full-stack | **Leptos** (SSR + hydration) sobre **Axum** |
| Front | WASM + HTML SSR nativo |
| Real-time | server functions / WS no mesmo stack |

**Prós:** um só idioma, SSR/SEO nativo, performance de ponta, footprint baixo.  
**Contras (solo):** ecossistema UI/charts menor; curva e compile times altos.

**Use se:** já é produtivo em Rust e quer zero Node no servidor.  
**Evite se:** prioridade for shipping dashboards com libs de chart/UI maduras agora.

Para SEO web-first, **Leptos** costuma estar mais maduro em SSR do que Dioxus.

### 3) Rust API + Next.js (App Router)

Mesma ideia do (1), com React. SSR/SEO excelentes; Next **standalone** + PM2 na VPS funciona, mas costuma consumir **mais RAM** que SvelteKit. Bun no front/runtime pode reduzir memória ~25–40% vs Node.

- **Next** se o ecossistema for React/shadcn/TS.
- **SvelteKit** se quiser menos JS e menos RAM no VPS.

---

## TypeScript puro (sem Rust)

| Stack | Perfil na VPS |
|-------|----------------|
| **SvelteKit (Bun/Node) + Postgres** | Fullstack leve, SSR, WebSocket; ótimo solo |
| **Next.js + Bun + Postgres** | Mais libs; um pouco mais pesado; ainda ok em 16 GB |
| **Hono/Elysia (Bun) API + SvelteKit** | API mais rápida/leve que Express clássico |

Para dashboards ao vivo + SEO, TypeScript sozinho resolve o requisito funcional. O “custo” é mais RAM/CPU por conexão e request do que Rust — com 16 GB e poucos milhares de sockets simultâneos, costuma ser irrelevante.

---

## Outras stacks de alta performance / baixo consumo

### Elixir + Phoenix LiveView

| Ponto | Por quê |
|-------|---------|
| Real-time | **Nativo** (LiveView = WebSocket + diff de HTML) |
| SSR/SEO | LiveView + layouts; bom para apps; marketing pages ok |
| Memória | Eficiente por conexão; hibernação de sockets |
| Solo | Alta produtividade para dashboards ao vivo |
| VPS | BEAM + Postgres cabe sobrando em 16 GB |

**Melhor escolha se** o produto for muito real-time e houver disposição de aprender Elixir.  
**Trade-off:** menos “mercado” que TS; UI é HEEx/HTML-first (hooks JS leves para charts).

### Go (Fiber/Chi/Echo) + Templ/HTMX ou + SvelteKit

| Ponto | Por quê |
|-------|---------|
| Performance | Top tier TechEmpower (junto com Rust/C#) |
| RAM | Binário estático, footprint baixo |
| DX | Mais simples que Rust; compile rápido |
| SSR | Templ + HTMX, ou front SvelteKit separado |
| Real-time | WebSocket maduro |

Meio-termo: quase a eficiência do Rust, menos atrito solo. Para admin/dashboard, **Go + Templ + HTMX + WS** é monólito enxuto e SEO-friendly.

### .NET (ASP.NET + Blazor)

Performance de elite nos benchmarks; no ecossistema Linux/VPS solo BR costuma ser menos natural que Go/Rust/Node. Só se já viver em C#.

### Evitar como stack principal por consumo/throughput

Python (Django/FastAPI) e Ruby: ótimos para prototipar; piores se o critério #1 for recurso/performance com real-time pesado.

---

## Ranking para este cenário

Solo + KVM 4 (16 GB) + SSR/SEO + dashboards ao vivo.

| Rank | Stack | Por quê |
|------|--------|---------|
| **1** | **Axum (Rust) + SvelteKit (SSR) + Postgres + Caddy/Nginx** | Backend max desempenho, front SEO maduro, real-time controlado, footprint baixo |
| **1-alt** | **Phoenix LiveView + Postgres** | Se real-time for o centro do produto e quiser menos “cola” front/back |
| **2** | **SvelteKit ou Next (Bun) full-stack + Postgres** | Máxima velocidade solo; 16 GB engole fácil |
| **3** | **Go + (Templ/HTMX ou SvelteKit)** | Quase Rust em eficiência, DX mais simples |
| **4** | **Leptos full Rust** | Só se quiser 100% Rust e aceitar ecossistema UI menor |
| **5** | Next/Node multi-worker sem cuidado | Funciona, mas come mais RAM à toa |

---

## Layout de deploy na VPS

```
Internet
   │
 Caddy/Nginx (TLS, reverse proxy)
   │
   ├─ app1.seudominio  → SvelteKit (SSR) :3001
   ├─ api1.seudominio  → Axum (API + WS)  :3000
   ├─ app2...          → outro projeto
   │
 PostgreSQL (compartilhado ou por app)
 Redis (opcional: pub/sub dashboards, filas)
```

### Orçamento de RAM (ordem de grandeza)

| Serviço | Idle típico |
|---------|-------------|
| Axum API | ~10–50 MB |
| SvelteKit/Bun | ~50–150 MB por app |
| Next standalone | ~100–300+ MB por app |
| Postgres | ~100–300 MB (configurável) |
| Redis | ~10–50 MB |
| SO + overhead | ~500 MB–1 GB |

Com **16 GB** cabem **3–6 projetos** fullstack + DB, se não deixar Next em modo dev e não abrir workers demais.

---

## Decisão rápida: Rust vs TypeScript

| Se você… | Escolha |
|----------|---------|
| Quer aprender Rust e API de alto desempenho, com SEO e UI hoje | **Axum + SvelteKit** |
| Quer só shipping e dashboards rápidos de codar | **SvelteKit/Next + Bun** |
| Real-time é o produto (muitos users ao vivo) | **Phoenix LiveView** |
| Quer eficiência quase Rust com menos dor | **Go + SvelteKit ou Go + Templ** |
| Quer full Rust a qualquer custo | **Leptos + Axum** |

Não escolha Rust full-stack só “porque gasta menos RAM” na KVM 4: a máquina já é grande. Escolha Rust no backend se valorizar modelo mental, segurança e eficiência de longo prazo — e deixe o SSR com SvelteKit/Next.

---

## Recomendação final

**Padrão default**

**Rust (Axum) no backend + SvelteKit (TypeScript) no frontend com SSR + PostgreSQL**, deploy com Caddy/Nginx; real-time via WebSocket/SSE no Axum.

**Plano B** (dashboard multiplayer ao vivo, menos peças): **Phoenix LiveView**.

**Plano C** (zero Rust agora, máxima velocidade solo): **SvelteKit + Bun + Postgres** (migrar hot paths para Rust depois, se precisar).

---

## Fontes

- [TechEmpower Framework Benchmarks](https://www.techempower.com/benchmarks/) — ranking de throughput (Rust, Go, C# no topo; Node atrás em RPS sintético)
- [Leptos vs Dioxus (SSR / web-first)](https://rustify.rs/articles/leptos-vs-dioxus-rust-frontend-2026)
- [Axum + SvelteKit (discussão / uso)](https://www.reddit.com/r/sveltejs/comments/1muqg58/hear_me_out_sveltekit_static_adapter_backend/)
- [CryptoFlow: Axum + SvelteKit](https://dev.to/sirneij/cryptoflow-building-a-secure-and-scalable-system-with-axum-and-sveltekit-part-0-mn5)
- [Bun vs Node (memória ~25–40% menor)](https://strapi.io/blog/bun-vs-nodejs-performance-comparison-guide)
- [Node / Deno / Bun 2025](https://dev.to/dataformathub/nodejs-deno-bun-in-2025-choosing-your-javascript-runtime-41fh)
- [Phoenix LiveView — memória e real-time](https://sheer.tj/three_years_of_phoenix_liveview_and_elixir.html)
- [LogRocket: top Rust web frameworks](https://blog.logrocket.com/top-rust-web-frameworks/)

---

## Boilerplate de deploy

Template pronto (Axum + SvelteKit + Postgres + Caddy + systemd):

→ [`boilerplate-axum-sveltekit-vps/`](./boilerplate-axum-sveltekit-vps/README.md)

## Skills RC (hub)

Guias de implementação, segurança, lints e testes para a stack:

| Skill | Escopo |
| ----- | ------ |
| [`rc-fullstack-axum-svelte`](../skills/rc-fullstack-axum-svelte/SKILL.md) | Guarda-chuva — roteia as três skills; front com **Bun ≥ 1.3** |
| [`rc-axum`](../skills/rc-axum/SKILL.md) | Axum 0.8 — API, WS, middleware, security, tests, clippy |
| [`rc-sqlx`](../skills/rc-sqlx/SKILL.md) | SQLx 0.8 + Postgres — queries, migrations, security, tests |
| [`rc-sveltekit`](../skills/rc-sveltekit/SKILL.md) | SvelteKit 2 + Svelte 5 — SSR, forms, CSP, adapter-node, **Bun** |

---

*Documento gerado a partir da análise de stack para a VPS Hostinger KVM 4 (2026-07-16).*
