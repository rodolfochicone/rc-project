---
title: Fósseis do fork Compozy (era Go/CLI)
scope: context
key: compozy-fork-fossils
tags: [fork, ci, legado, de-fork, go]
source: rc-memory (distilled 2026-07-14)
created: 2026-07-14
updated: 2026-07-14
---

Este repo nasceu de um **fork do Compozy**, que era um **CLI em Go** com daemon, TUI e web UI.
O de-fork removeu o Go, mas sobrou fóssil em todo canto que ninguém abre. Em 2026-07-14 ainda
foram achados (e removidos): `ci.yml` montando Go 1.26.1 + Bun + Playwright e rodando
`make verify`; `.github/actions/setup-go|setup-bun|setup-node|setup-git-cliff|setup-release`
(órfãs); `.github/versions.yml` (go/golangci-lint/cosign/syft, lido por ninguém);
`codeql-config.yml` apontando para `cmd/`, `internal/`, `rc.go`; `auto-docs.yml` mandando
`go run github.com/rc/releasepr` (módulo de **outro org**) escrever em `.release-notes/`;
e `skills/rc-smux-rc-pairing/scripts/run-compozy-start.sh`.

**O padrão do de-fork:** ele reescreveu **conteúdo** e esqueceu **nomes e config**. O
`run-compozy-start.sh` já imprimia `Usage: run-rc-start.sh` no próprio banner — só o nome do
arquivo ficou para trás. Corolário prático: quando um nome e o conteúdo divergem, **o conteúdo
decide quem está stale**.

**Onde procurar o próximo:** config que nenhum código lê (`.github/`, `.vscode/`, dotfiles de
tooling), nomes de arquivo/script, e qualquer coisa que cite `make`, `go`, `bun`, `daemon`,
`ACP` ou `web/`. O gate hoje cobre parte disso (ver [[gate-sensor-over-patch]]), mas só o que é
markdown/workflow — dotfiles de tooling seguem sem sensor.
