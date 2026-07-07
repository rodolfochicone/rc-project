---
scope: decision
key: project-memory-sharing
title: Memória de projeto deve ser compartilhada via mirror em texto (Opção B)
tags: [architecture, memory, sharing]
source: rc-analyze
updated: 2026-06-30
---

O `.rc/memory.db` (SQLite, gitignored) é local por máquina e não compartilha entre a equipe.
Decisão: versionar um arquivo de texto (markdown por fato) como fonte compartilhada; o `.db`
seria apenas um cache local reconstruído a partir dele. A Opção A (commitar o `.db` binário)
foi rejeitada por conflitos de merge irresolúveis com 2+ pessoas.

Status: implementado no rc v1.0.0 — o store passou a ser `.rc/memory/<scope>__<key>.md`
versionado no git, e o SQLite foi descontinuado junto com o binário. Ver [[rc-project-memory]].
