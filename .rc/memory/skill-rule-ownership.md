---
title: Uma regra, um dono — e o dono quase nunca é a seção de recap
scope: convention
key: skill-rule-ownership
tags: [skills, doutrina, duplicacao, recap, hooks]
source: rc-memory (distilled 2026-07-17)
created: 2026-07-17
updated: 2026-07-17
---

Auditadas as 64 skills contra a doutrina da `rc-skill-best-practices` (que só chegou em
2026-07-17 — antes ela tinha 402 palavras de mecânica de spec e nenhum vocabulário para
*single source of truth*, *no-op*, *sediment*), **47 falhavam em "uma regra, um dono" (73%)**.
Causa única: uma seção final (`Critical Rules` / `Guardrails` / `Must not do` / `Constraints`)
reafirmando o que o Workflow já dizia. 36 das 64 carregavam a seção. Não eram 47 descuidos —
era o template da casa.

**A régua, degrau a degrau. O primeiro que casar decide:**

1. **Um hook já garante?** → deleta a **proibição**, mantém a **direção positiva**.
2. **Outra skill é dona?** → uma linha de ponteiro. Nunca reescreve o conteúdo dela.
3. **Um `references/` desta skill é dono?** → ponteiro com "in full" e com o *quando*.
4. **Desta skill e governa todos os steps?** → um enunciado, **acima** dos steps.
5. **Desta skill e governa um step?** → dentro daquele step.
6. **Sem dono, e um modelo de fronteira já faria?** → no-op, deleta.
7. **Sem dono e delta real?** → fica.

**As duas emendas que os arquivos impuseram** (a versão ingênua da regra estava errada):

- **Degrau 1 não é "deleta a prosa".** O `db-guard` bloqueia escrita no banco de verdade, mas
  um hook *impede* a violação — não *ensina* a alternativa. Sem prosa, o agente tenta o
  `UPDATE`, toma bloqueio e só então recomenda: um turno queimado, toda vez. Deleta a
  proibição, mantém a direção. É a própria regra de "Negation" da doutrina.
- **Um `references/` é dono tão legítimo quanto outra skill — e é o que mais escapa**, porque
  não aparece no `SKILL.md`. Metade das deleções do fanout tinha dono num `references/`.

**Seções `## Anti-Pattern:` têm critério próprio — e não é o nome do heading:**
ficam se **nomeiam uma racionalização** que o agente vai fazer ("quando o nome da feature soa
técnico, você VAI ser tentado a discutir HOW"); saem se só reafirmam a regra com outro nome.
No `rc-create-prd`, 2 dos 3 eram duplicação e 1 sobreviveu. Um sweep por nome de heading
destruiria o que é conteúdo.

**O ganho não é linha economizada.** `rc-no-workarounds`, `rc-final-verify` e
`rc-testing-anti-patterns` existiam e **ninguém apontava para elas** — as skills parafraseavam
seu conteúdo. O `rc-fix-analysis` comprimia uma skill de 1957 palavras num bullet; os dois iam
derivar. Depois da regra, elas passam a ser *usadas*.

**A prova de que duplicação deriva** (não é teoria): o bloco `Resolving the .rc base directory`
está em **8 skills e já são 5 variantes**. E a deriva é comportamental — num monorepo, o
`rc-code-review` resolve sozinho e o `rc-create-prd` sempre pergunta. Colapsar isso numa fonte
segue aberto.

**Antes de declarar contradição, leia a referência.** "up to 3 questions" (Checklist) e "at
least one full round" (Workflow) *parecem* conflitar e **compõem** — teto e piso, os dois fiéis
ao `question-protocol.md`. Quase virou achado falso.

**O que o plugin distribui decide quem pode ser dono.** O `~/.claude/CLAUDE.md` do usuário é
privado; o `CLAUDE.md`/`AGENTS.md` da raiz dizem "for agents working in **this repository**".
Quem instala o `rc` não recebe nenhum dos três — então "CLAUDE.md Rule 3 já cobre" **não** é
argumento para deletar de uma skill. Só hook e skill viajam com o plugin.

Ver [[gate-sensor-over-patch]] para os sensores que passaram a vigiar parte disso.
