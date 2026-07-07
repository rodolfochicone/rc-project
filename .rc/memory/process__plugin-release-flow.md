---
scope: process
key: plugin-release-flow
title: Fluxo de release do plugin — bump de versão é obrigatório para chegar às instalações
tags: [release, plugin, marketplace, versioning]
source: session
updated: 2026-07-07
---

O updater do Claude Code só baixa o plugin de novo quando a versão muda: commit na main
sem bump deixa `claude plugin update` respondendo "already at the latest version" e o
cache desatualizado. Aprendido na prática no release da v1.0.2.

Fluxo validado:

1. `CHANGELOG.md`: mover o conteúdo de `## [Unreleased]` para `## [X.Y.Z] - <data>`,
   mantendo o heading Unreleased vazio no topo.
2. Bumpar `version` em `.claude-plugin/plugin.json` **e** `.claude-plugin/marketplace.json`
   (`plugins[0].version`) — o CI `validate.yml` falha se divergirem, e a `rc-doctor`
   exige que batam com o último release do CHANGELOG.
3. Commit `chore(release): pin plugin manifests and changelog at vX.Y.Z` e push na main.
4. `claude plugin marketplace update rc-project` → `claude plugin update rc@rc-project`
   → reiniciar o Claude Code para a sessão usar a nova versão.
5. Verificar: grep da mudança em `~/.claude/plugins/cache/rc-project/rc/X.Y.Z/` e CI
   verde (`gh run list --workflow=validate.yml`).
6. Higiene: versões antigas ficam no cache (`claude plugin prune` não as remove);
   conferir que `installed_plugins.json` aponta só para a nova e apagar os diretórios
   antigos. O check 7 da [[rc-doctor]] ("Installed copy freshness") detecta ambos.
