---
title: Fósseis do fork Compozy (era Go/CLI)
scope: context
key: compozy-fork-fossils
tags: [fork, ci, legado, de-fork, go, install, hooks]
source: rc-memory (distilled 2026-07-14)
created: 2026-07-14
updated: 2026-07-17
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

**Update 2026-07-15:** os fósseis Go foram removidos no corte pre-slim (commits 743377b + dd4a344): agents/README.md (agente fantasma rc:README), hooks go-fmt/go-mod-guard (+ canal opencode), rc-fix-coderabbit-review, rc-app-renderer-systems, rc-portal-design e o cluster Go/TUI inteiro. Tag `pre-slim` guarda o estado anterior.

**Update 2026-07-17 — "os fósseis Go foram removidos" era otimista.** Sobreviveu a **duas**
limpezas (de-fork e pre-slim) escondido *dentro da prosa de uma skill*, não em config:
`rc-create-techspec` exigia "at least one **Go interface or struct** definition" na seção Core
Interfaces — numa skill stack-agnostic, num repo Rust/SvelteKit. **Toda TechSpec gerada recebia
essa ordem.** Confirma o padrão desta memória num eixo novo: as duas limpezas caçaram *config e
nomes de arquivo*; ninguém leu o corpo das skills. Corolário: o fóssil seguinte está na prosa,
não no `.github/`.

**Uma segunda fonte de fóssil, que esta memória não previa: a vendorização.** Boa parte do hub
veio de `pedronauck/skills` (17 skills com contraparte 1:1) e de outros upstreams. O
`rc-systematic-debugging` mandava usar `superpowers:test-driven-development` e
`superpowers:verification-before-completion` — família que nunca existiu aqui; os equivalentes
(`rc-tdd`, `rc-final-verify`) estavam a um diretório de distância. O `find-docs` (citado pelo
CLAUDE.md e pelo `rc-video`) era da mesma classe: ponteiro para skill inexistente.

**O tell da skill vendorizada e nunca adaptada** — medido, não intuído: das que batem por nome
com o upstream, **71% falham** o gate contra **15%** das nativas. As 5 vendorizadas limpas são
exatamente as que alguém adaptou. Sintomas: identity prose na `description` ("Comprehensive
guide for…", "Expert guide", "Essential for…"), ausência de anti-trigger, e `references/` que o
`SKILL.md` nunca cita. O gate hoje vê os três (ver [[gate-sensor-over-patch]]).

## O fóssil também mora FORA do repo — e ali nenhum gate alcança

Descoberto em 2026-07-17 ao atualizar o plugin instalado. Existia um install **manual e
pré-plugin** em `~/.claude/rc/hooks/scripts/`, wired por **caminho absoluto** no
`~/.claude/settings.json` — nunca documentado no repo (`grep` em `docs/` e `README.md` volta
vazio). Consequências, todas ativas havia ~1 mês:

- **Rodava em paralelo com o plugin.** O `rc@rc-project` está habilitado e seu `hooks.json` já
  wira os 9 via `${CLAUDE_PLUGIN_ROOT}`; o `settings.json` wirava mais 8 por cima. Cinco hooks
  disparavam **duas vezes**, e a cópia manual era a **velha**.
- **Ressuscitava hook deletado.** `go-fmt.sh` e `go-mod-guard.sh` saíram do repo no corte
  pre-slim (`dd4a344`) e seguiam disparando na máquina, a cada Edit e a cada Bash. **Deletar do
  repo não desinstala.**
- **Tinha conteúdo que nunca foi commitado.** O `observe.sh` instalado não bate com *nenhuma*
  versão do histórico: era da era `rc-instincts` (opt-in, `.rc/instincts/`), enquanto o repo já
  tinha `rc-memory` (opt-out, `.rc/memory/`). Install mantido à mão deriva.

**O tell:** a mensagem de erro do hook nomeia o caminho. `[~/.claude/rc/hooks/scripts/git-guard.sh]`
= install manual; `[${CLAUDE_PLUGIN_ROOT}/hooks/scripts/git-guard.sh]` = plugin. Foi assim que a
duplicação apareceu — e assim que se prova a remoção.

**Resolvido em 2026-07-17:** os 8 hooks do rc saíram do `settings.json` (sobrou só o
`ship_gate.py`, que não é do rc) e `~/.claude/rc/` foi apagado. Backups:
`~/.claude/settings.json.bak-2026-07-17` e `~/.claude/rc-hooks-backup-2026-07-17.tgz`.

**A lição que generaliza:** o gate só vê o repositório. Config de máquina (`~/.claude/`,
dotfiles, `settings.json`) é ponto cego permanente — quando o repo remove um hook, uma skill ou
um script, **verifique também o install**, senão o defeito continua vivo em quem já instalou. O
canal do plugin (marketplace + `hooks.json`) é a única forma que se atualiza sozinha; qualquer
wiring manual vira fóssil no dia seguinte.
