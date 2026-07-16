# RC — Guia de Comandos

Referência rápida das skills e comandos do RC. RC é um **plugin de agente** (skills, commands,
agents, hooks) — não há binário, daemon nem CLI. Tudo roda dentro do seu agente (Claude Code,
OpenCode e outras ferramentas); os artefatos vivem em `.rc/` dentro do projeto.

> **Pré-requisito:** instale o plugin pelo mecanismo de plugin/marketplace do seu host.
> No Claude Code: `/plugin marketplace add rodolfochicone/rc-project` e `/plugin install rc@rc-project`.

---

## 1. Pipeline ideia → código (skills, dentro do agente)

| Comando | O que faz |
| --- | --- |
| `/rc-create-prd <nome>` | Brainstorm + pesquisa → **PRD** de negócio |
| `/rc-create-techspec <nome>` | PRD → **TechSpec** (arquitetura, APIs, modelos de dados) |
| `/rc-create-tasks <nome>` | PRD + Spec → **tasks** executáveis |
| `/rc-tasks-workflow <nome>` | Executa as tasks via o Workflow tool (Claude Code) |
| `/rc-review-round <nome>` | Revisão de código multi-lente → gera issues |
| `/rc-review-workflow <nome>` | Loop automatizado review → fix → re-review |
| `/rc-fix-reviews <nome>` | Corrige os issues de review |
| `/rc-analyze <pergunta>` | Análise profunda, baseada em evidência, de código existente |
| `/rc-final-verify` | Exige evidência de verificação antes de declarar "concluído" |
| `/rc-git [ticket]` | Cria branch, push e abre PR com confirmação em cada passo |
| `/rc-board` | Cria, lê, comenta e move issues em qualquer board (Linear, Jira) via MCP oficial |

Memória e aprendizado: `/rc-project-memory` (fatos duráveis do projeto), `rc-workflow-memory`
(contexto entre tasks de um workflow), `/rc-instincts` (padrões recorrentes).

---

## 2. Executar tasks

- **Claude Code:** `/rc-tasks-workflow <nome>` — roda as tasks em ordem de dependência via o
  `Workflow` tool, uma por vez.
- **Qualquer host:** execute cada `task_NN.md` pela skill `rc-execute-task`, em ordem de dependência.
- **Validar antes de rodar:** `node "$CLAUDE_PLUGIN_ROOT/scripts/validate-tasks.mjs" --slug <nome>`.

## 3. Reviews

- `/rc-review-round <nome>` — revisão multi-lente; gera issues em `reviews-NNN/`.
- `/rc-review-workflow <nome>` — loop automatizado review → fix → re-review.
- `/rc-fix-reviews <nome>` — triagem e correção dos issues do round.

## 4. Loop autônomo (opcional)

O loop é para **migração e build-out grande atrás de um harness verde** — feature normal continua
em `/rc-pipe`. Autonomia se conquista: o `/rc-loop` só roda depois das quatro perguntas de
prontidão (harness forte? feedback rápido? condição de parada confiável? backlog grande o
bastante?). Qualquer "não" → fique no fluxo humano-gated e invista no harness.

| Comando | O que faz |
| --- | --- |
| `/rc-roadmap [create\|next\|status]` | Autora/lê o `.rc/ROADMAP.md` — as fases-épico. É o passo de **intenção humana**: o loop executa, não inventa. |
| `/rc-loop` | Anda o roadmap fase a fase (plan → execute → verify → aprende → fecha) até esgotar ou uma fase não ficar verde. |

Por trás, a skill `rc-lessons` carrega as lições confirmadas no planejamento e registra as novas no
verify — é o que impede o loop de repetir os próprios bugs a cada fase.

O loop deixa a working tree verde e commitada por fase. **PR, push e escrita no Linear/Jira nunca
são autônomos** — continuam sendo passo humano confirmado (`/rc-git`).

---

## Fluxo típico de ponta a ponta

```text
/rc-create-prd minha-feature
  → /rc-create-techspec minha-feature
  → /rc-create-tasks minha-feature
  → /rc-tasks-workflow minha-feature      (ou rc-execute-task por task, em qualquer host)
  → /rc-review-round minha-feature
  → /rc-fix-reviews minha-feature
  → /rc-git
```

---

## Onde as coisas vivem

- **Skills / commands** rodam dentro do seu agente de IA.
- **Artefatos** vivem em `.rc/tasks/<nome>/` (markdown versionável): PRD, TechSpec, tasks, ADRs,
  reviews. Memória e aprendizado (instincts) curados em `.rc/memory/`, análises em `.rc/analysis/`.
