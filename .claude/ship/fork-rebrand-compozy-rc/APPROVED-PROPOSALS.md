# Propostas aprovadas — planejar APÓS o fork (não fazem parte do build do fork)

> Aprovadas pelo usuário em 2026-06-13. NÃO implementar dentro do build do fork
> (`/ship-build-wf fork-rebrand-rc-rc`). Cada uma vira um plano `/ship-wf` próprio depois.
> Base: pesquisa web (tendências 2026 p/ agentes de código) + análise de gaps reais no código do rc.

## P-A — Guardrails de custo/budget

Tracking de tokens + custo por run, com **teto rígido** (hard cap) e atribuição por usuário/projeto.

- Gap: rc tem cost tracking raso (~3 arquivos), zero budget cap.
- Valor rc: controlar e atribuir gasto de IA internamente.

## P-B — Observabilidade (OpenTelemetry + Prometheus)

Traces e métricas do pipeline de agentes, exportáveis para Datadog / Grafana / Honeycomb.

- Gap: rc tem OTel mínimo (~1 arquivo).
- Valor rc: visibilidade de Ops/SRE sobre o pipeline.

## P-C — Execução em sandbox

Rodar comandos/código gerado por IA em isolamento: allowlist de rede, isolamento de secrets.

- Gap: rc tem 0 sandbox.
- Valor rc: segurança ao executar código de IA não confiável.

## P-D — Migração `~/.rc` → `~/.rc`

Import automático de config/estado de quem já usava rc.

- Já estava como proposta adiada no PRD (Open Question 3).
- Valor rc: transição suave para quem migra do rc.

## P-E — Rich welcome header (T10 USER DECISION — in-scope, within fork build)

Resolves SPEC §D-Q1 / Open Question 1 / Proposal P1. User decided on 2026-06-13:

> Implement the rich header (Claude Code style), NOT just a re-skin. This is a deliberate net-new feature; AC7 "zero functional change" carries a documented exception for the welcome header only.

**What was approved:**

- Rounded orange box (`lipgloss.RoundedBorder()` with rc orange via theme tokens)
- Top label: `rc // SETUP`
- `Welcome back <user>!` greeting (OS user / git config; generic fallback)
- ASCII peak/triangle mascot (solid rc orange)
- Dimmed context line: `rc vX.Y.Z · <cwd>` (fields sourced from real data only; `model`/`plan`/`org` omitted — no real data source exists yet)
- AC7 exception is documented here; the exception covers ONLY this component

**What was NOT approved:** fabricating `model`, `plan`, or `org` fields. These are OMITTED until a real data source exists.

## Ainda parados no PRD (não aprovados ainda)

Renomear prefixo `cy-*`, política de telemetria, pipeline de publicação (brew/npm/AUR).
