---
title: Processo de release do plugin
scope: convention
key: release-process
tags: [release, versioning, changelog, github]
source: rc-memory (distilled 2026-07-12)
created: 2026-07-12
updated: 2026-07-12
---

Como cortar um release do plugin `rc`:

1. Bump da versão em **ambos** `.claude-plugin/plugin.json` e `.claude-plugin/marketplace.json` (juntos). **Não** bumpar `package.json` (`0.2.4`) nem `extensions/rc-idea-factory/extension.toml` — versionam independente do plugin.
2. CHANGELOG (Keep a Changelog): mover `## [Unreleased]` → `## [X.Y.Z] - <data>`, readicionar `[Unreleased]` vazio. Refs de link no rodapé: `[Unreleased]` → `compare/vULTIMA...main`; versão com release → `/releases/tag/vX.Y.Z`. Só criar ref para versão que tem seção **e** release publicado.
3. Commit `chore(release): vX.Y.Z` — só os 2 manifests + CHANGELOG.
4. Tag anotada: `git tag -a vX.Y.Z -m 'rc vX.Y.Z — <resumo>'`.
5. GitHub release: `gh release create vX.Y.Z --notes-file -` com as notas da seção do CHANGELOG (ou a mensagem da tag se não houver seção).

**Why:** a versão fica duplicada nos 2 manifests do marketplace; esquecer um deixa o plugin inconsistente na distribuição.

**Gotcha:** ao criar vários releases de uma vez, o `gh` marca "Latest" o **último criado por data**, não o maior semver — corrigir com `gh release edit vMAIOR --latest`.
