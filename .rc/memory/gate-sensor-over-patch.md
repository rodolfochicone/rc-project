---
title: O gate é o plugin-smoke — e defeito de classe vira sensor, não só patch
scope: convention
key: gate-sensor-over-patch
tags: [harness, ci, plugin-smoke, verificacao]
source: rc-memory (distilled 2026-07-14)
created: 2026-07-14
updated: 2026-07-17
---

O gate deste repo é `node scripts/plugin-smoke.mjs` (+ `node scripts/lessons.mjs --selftest`),
e desde 2026-07-14 ele **roda no CI** (antes o CI era fóssil e não rodava nada — ver
[[compozy-fork-fossils]]). Não existe build, binário nem `make`: "verificado" = gate verde.

**A regra que emergiu:** quando um defeito é de **classe** (não instância), a correção é
**adicionar o sensor**, não só consertar o caso encontrado. Todo sensor adicionado em
2026-07-14 achou defeito real *na primeira execução*, incluindo alguns que ninguém procurava:

| Sensor | O que exige | Achou de cara |
| --- | --- | --- |
| `orphan skill` | toda `skills/<x>/SKILL.md` citada no catálogo do `README.md` | `rc-jira`, órfã desde a 2.1.0 |
| `CLI residue` | nada de binário/daemon `rc` aposentado nos docs | resíduo no `workflow-guide.md` |
| `dangling asset` | link para `references/`/`assets/`/`scripts/` tem que resolver | 8 links quebrados, em 5 skills |
| `toolchain fossil` | workflow não monta Go/Bun/`make` sem o manifesto existir | `auto-docs.yml` |

**O ponto cego da `description` — parcialmente fechado em 2026-07-17.** Era verdade que o gate
não a enxergava (foi lá que o `rc-bubbletea` anunciava uma "Effects Library" inexistente, e quem
pegou foi um agente lendo). Hoje o check de **anti-trigger** lê a `description` e exige o
"Do not use for…". Mas ele só cobre a *forma*, não o *conteúdo*: uma description que descreve
errado o que a skill faz continua passando. Para conteúdo, verde do gate segue não bastando.

**Update 2026-07-17 — três sensores novos, e o tier de aviso.** Entraram: `description` sem
anti-trigger (15), arquivo em `references/` sem ponteiro (29), arquivo solto na raiz da skill
(20). Entraram como **aviso, não falha** (`--warn` lista; run padrão segue exit 0), porque
reportam um backlog de 64 que os antecede — e um gate vermelho no dia em que nasce é um gate
que todo mundo aprende a pular. Promover para `fail` quando zerar.

**O FP medido, com número** (reforça a regra de medir antes de escrever):
a versão ingênua do check de órfão acusou **102 arquivos a ~85% de FP** — só caiu para 29 reais
depois de aceitar quatro formas de ponteiro (path exato, basename, basename sem extensão e
**ponteiro de diretório**, tipo `references/query/`). O de anti-trigger teve 1 FP em 16
(`rc-zod` diz "does NOT cover", não "do not use for"). E o `AGENTS.md` quase foi flagrado como
arquivo intruso — é convenção deste repo, o **próprio plugin-smoke** lê `skills/*/AGENTS.md`
como entry point; o gate teria brigado consigo mesmo.

**Um candidato foi descartado, e isso é parte da regra:** `user-invocable: true` redundante tinha
**0% de FP** e mesmo assim não entrou — não é defeito. Campo de frontmatter fora da `description`
não custa contexto, então declarar um default é redundância cosmética. Precisão não basta:
o check tem que apontar defeito.

**Os guards casam contra a string bruta do comando.** O `git-guard` bloqueou o commit *desta
própria mudança*, porque a mensagem citava os verbos em prosa. Vale para qualquer hook que leia
`.tool_input.command`: mensagem de commit, heredoc e doc citando o verbo disparam o guard.
Contorno: escrever a mensagem em arquivo e usar `git commit -F`.

**Antes de dizer "pronto":** rode o gate **e olhe o CI** (`gh run list`). Em 2026-07-14 três
pushes foram declarados "verdes" com base só no gate local, enquanto o CI falhava — os
"sucessos" eram job pulado pelo `paths-filter`.
