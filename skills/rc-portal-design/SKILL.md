---
name: rc-portal-design
description: Design system e padrões de frontend do rc-portal. Use ao criar ou alterar QUALQUER UI no rc-portal — telas, componentes, estilos/Tailwind, tokens/tema (light/dark/named themes), formulários, gráficos, Storybook, responsividade ou acessibilidade. Garante reuso dos componentes existentes (shared/ui, primitives), uso exclusivo de tokens semânticos e os gates obrigatórios (stories no mesmo PR, axe zero violações). Não use para lógica de BFF/Route Handlers, autenticação/Clerk ou regras de negócio — para isso siga CLAUDE.md e agents-docs/.
---

# rc-design — UI do rc-portal

Stack real: **Next.js 15 App Router + React 19 + TypeScript**, **Tailwind 3.4** com preset
`@escaletech/delta-tailwind` + tokens locais, **shadcn/ui (style new-york, cssVariables)**,
Radix primitives, **lucide-react** (ícones), react-hook-form + Zod, TanStack Query v5,
Recharts, Storybook, Jest + Testing Library, Playwright + axe-core. Gerenciador: **yarn**
(não pnpm). Dark mode por classe; temas nomeados por escopo de CSS vars.

## Ordem de decisão ao criar UI (reuso primeiro)

1. **Já existe?** Procure nesta ordem: `src/shared/ui/` (43+ componentes prop-driven),
   `src/components/primitives/` (Button, Card, Container), `src/components/ui/` (shadcn:
   card, chart, collapsible, popover, skeleton), `src/components/` (data_table, drawer,
   confirm_modal, header_page, …). Só então considere criar.
2. **Vai criar e tem potencial de reuso?** Nasce **domain-agnostic e prop-driven em
   `src/shared/ui/`** — nunca acoplado à feature. Arquivos em `snake_case.tsx` com
   `*.stories.tsx` ao lado.
3. **É primitivo transversal?** `src/components/primitives/<Nome>/` com a estrutura
   `types.ts` + `constants.ts` + `<Nome>.tsx` + `index.ts` (ver o README da pasta).
4. **Precisa de um primitivo shadcn novo?** Adicione via CLI do shadcn (config em
   `components.json`, aliases `@/components`, `@/lib/utils`) — não copie à mão.

Botão canônico: `Button` de `src/components/Button` — variants `primary` (coral),
`secondary`, `tertiary`, `outline`, `nofill`, `message`, `table`; sizes `sm | md | xl`.
Classes combinadas sempre com `cn()` de `src/lib/utils.ts` (clsx + tailwind-merge).

## Tokens — regra inegociável

**Apenas tokens semânticos**: `bg-background`, `text-foreground`, `bg-card`,
`text-muted-foreground`, `bg-primary`, `text-primary-foreground`, `bg-brand`,
`border-border`, `bg-sidebar-*`, `text-destructive`, `chart-*`, `pastel-*`…
Proibido raw palette em UI de produto (`text-neutral-500`, `bg-purple-600`).

Cadeia de verdade: `src/app/globals.css` (CSS vars light/dark/named themes) →
`tailwind.config.ts` (mapa semântico) → `src/theme/` (acesso via TS). Antes de criar
cor nova, verifique se um token equivalente já existe nos três. Cor que aparece em
2+ lugares **vira token global antes do segundo uso** — e entra em todos os escopos
de tema. Tela de signin só via tokens `signin-*`. Radius: `--radius` (0.625rem) e
`--radius-btn` (1.5rem). Detalhes: `agents-docs/TAILWIND_THEME.md`.

## Formulários (3+ campos)

React Hook Form + Zod via `useZodForm` (`src/shared/forms/use_zod_form.ts`).
Um `*.schema.ts` por formulário/entidade em `src/core/` (preferido); o **mesmo schema**
valida cliente (`zodResolver`) e servidor (`safeParse`). Mensagens i18n via factory
`createXSchema(t)`. Erros de API: `setError('root', …)`. Nada de `useState` por campo.
Referência: `agents-docs/FORM-VALIDATION-RHF-ZOD.md` e `create_service_modal.tsx`.

## Dados na UI

`useQuery`/`useInfiniteQuery` com keys em `src/shared/query_keys.ts` e
`queryFn: ({ signal }) => …`; mutações invalidam listagens. **Nunca** `useEffect` +
`fetch` para listagem. Não duplicar estado que já vive no cache do React Query.
Ver `agents-docs/REACT-QUERY-DATA-FETCHING-AND-PERFORMANCE.md`.

## Responsividade

Breakpoints Tailwind: `sm` 640 / `md` 768 / `lg` 1024 / `xl` 1280. Sidebar e DataTable
tratam `< lg` (1024) como mobile (drawer / cards). O hook `useIsMobile`
(`src/hooks/use_is_mobile.ts`) usa 640 por default — use `lg`/1024 para casar com
sidebar/DataTable, e o hook/`sm` para comportamento phone-only. Cobertura esperada:
`agents-docs/MOBILE-RESPONSIVE-COVERAGE.md`.

## Storybook — mesmo PR (obrigatório)

Componente novo ou mudança visual/de fluxo → criar/atualizar `*.stories.tsx` no mesmo PR,
colado ao componente. Título `Features/<Área>/<Subárea>/<Nome>`. Variantes mínimas:
`Default` + `loading`/`empty`/`error` existentes + modais/abas tocadas pelo diff.
Mocks em `__fixtures__/`; decorators mínimos em `*_story_decorators.tsx`. No máximo
**uma** story `DesignReview` exploratória. Storybook não substitui testes de lógica
(e Jest não substitui Storybook para UI estática). Comandos: `yarn storybook`,
`yarn build-storybook`. Ver `agents-docs/STORYBOOK-UI-COMPONENTS.md`.

## Acessibilidade

WCAG 2.1 AA obrigatório; **zero violações axe** nas rotas ativas. Rode
`yarn playwright:test:a11y` antes de entregar mudança de UI. Landmarks, labels em
formulários e botões/links semânticos: `agents-docs/ACCESSIBILITY.md`.

## Limites que esta skill não afrouxa

- Máx. **300 linhas/arquivo**; sem comentários narrativos em `src/`.
- Não tocar no proxy Clerk (`/__clerk`, `src/middleware/clerk_proxy.ts`) sem aprovação.
- Rota `/` nunca pode 500/502 — cuidado com SSR/fetch no servidor (fail-safe).
- Gate de entrega: `yarn lint` + testes do escopo + `yarn build` quando aplicável, e os
  checklists `FRONTEND-BFF-IMPLEMENTATION-CHECKLIST.md` (antes/depois) e
  `POST-CHANGE-BRANCH-REVIEW.md` (ao finalizar).
