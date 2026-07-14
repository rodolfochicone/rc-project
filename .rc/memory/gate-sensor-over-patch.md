---
title: O gate é o plugin-smoke — e defeito de classe vira sensor, não só patch
scope: convention
key: gate-sensor-over-patch
tags: [harness, ci, plugin-smoke, verificacao]
source: rc-memory (distilled 2026-07-14)
created: 2026-07-14
updated: 2026-07-14
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

**O ponto cego que fica:** o gate **não enxerga a `description` do frontmatter** — foi lá que o
`rc-bubbletea` seguia anunciando uma "Effects Library" inexistente, e quem pegou foi um agente
lendo, não o sensor. Frontmatter `description` carrega em toda sessão: mudança de conteúdo exige
olhar humano/agente, verde do gate não basta.

**Antes de dizer "pronto":** rode o gate **e olhe o CI** (`gh run list`). Em 2026-07-14 três
pushes foram declarados "verdes" com base só no gate local, enquanto o CI falhava — os
"sucessos" eram job pulado pelo `paths-filter`.
