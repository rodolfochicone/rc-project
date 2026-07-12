---
title: Cross-links entre artefatos .rc/
scope: convention
key: artifact-cross-links
tags: [pipeline, prd, techspec, tasks, review]
source: rc-memory (distilled 2026-07-12)
created: 2026-07-12
updated: 2026-07-12
---

Os artefatos do pipeline em `.rc/tasks/<slug>/` cross-linkam entre si via **links markdown relativos** (seções `Related Artifacts`), nunca `[[wikilinks]]`: PRD ⇄ TechSpec ⇄ tasks, e cada `task_NN.md` → `adrs/adr-NNN.md` (caminho `adrs/`, **não** `../adrs/` — os `task_NN.md` ficam na raiz do slug junto de `adrs/`).

Funciona porque os nomes core são determinísticos (`_prd.md`, `_techspec.md`, `_tasks.md`), então os links são estáticos e podem "pendurar" sem quebrar.

Nos reviews, cada `issue_NNN.md` linka de volta à feature com `> Feature: [PRD](../_prd.md) · ...` no **corpo**. **Nunca** adicionar campos de backlink no frontmatter das issues — ele é parseado por tooling Go estrito (`prompt.ParseReviewContext()`, `reviews.ExtractIssueNumber()`) e campos novos quebram o parser.

**Why:** dá um grafo navegável dos artefatos no host e em editores estilo wiki, sem app externo nem convenção nova (a ideia veio ao avaliar o `inkeep/open-knowledge` — que foi descartado para integração).
