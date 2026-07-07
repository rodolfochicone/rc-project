# Changelog

## [Unreleased]

### Added

- **`rc-doctor`** — nova skill de health-check da própria instalação do rc:
  hooks presentes/válidos/wired, manifestos JSON com versões consistentes,
  frontmatter de skills/commands/agents bem-formado, paridade dos guard hooks
  Claude↔OpenCode e ferramentas exigidas (jq, gh) no PATH. Read-only: prescreve
  o fix, não aplica.
- **CI de validação** (`.github/workflows/validate.yml`): `bash -n` em todos os
  hook scripts, `jq empty` + consistência de versão nos manifestos, checagem de
  frontmatter (incluindo a regra de angle brackets do marketplace) e paridade
  dos guard hooks no `rc-hooks.ts` — a cada push/PR.
- **Mapa de uso** no README (pipeline mermaid, tabela "I want to…" por intenção
  Claude/OpenCode, dia típico) e versão HTML interativa em
  `docs/usage-map.html`.

### Removed

- **~450 arquivos não-produto destrackeados** e adicionados ao `.gitignore`:
  `.agents/` (skills de terceiros + cópia pré-1.0.0 obsoleta das skills rc e as
  antigas `cy-*`), os symlinks `.junie/`/`.kilocode/`/`.pi/`/`.qwen/` (lixo do
  rename cy→rc) e `.claude/ship/` + `.claude/workflows/` (artefatos de outros
  projetos). O marketplace clona o repo inteiro, então isso tudo era
  distribuído a cada instalação do plugin. Os arquivos permanecem no disco,
  apenas fora do versionamento. `.rc/tasks/` foi mantido versionado de
  propósito — dogfooding coerente com o design do rc (artefatos versionados
  junto ao código).

### Fixed

- **Descriptions com `<slug>`** em `rc-code-review` e `rc-simplify-review`
  trocadas para `{slug}` — angle brackets em `description` quebram a validação
  do marketplace de plugins do Claude Code.
- **README** dizia "10 agents" na tabela de layout; são 12 (10 de fase + 2
  read-only de apoio).

### Changed

- **`rc-explorer` e `rc-librarian` (OpenCode)** agora desligam `write`/`edit`
  no frontmatter, espelhando a restrição de tools que o lado Claude já impunha
  aos agents read-only.

## [1.0.1] - 2026-07-07

### Fixed

- **Migração da memória legada.** O único fato preso no `.rc/memory.db` (SQLite
  descontinuado) foi recuperado para o store file-based em `.rc/memory/`.
- **Sidecars do SQLite não são mais versionados.** O `.gitignore` passou a cobrir
  `.rc/memory.db*` (antes só `.rc/memory.db`), evitando commitar `-wal`/`-shm`.

### Changed

- **`rc-deepwork`** ganhou delimitação explícita contra o pipeline PRD formal.
- **`rc-reflect`** e **`rc-analyze`**: ajustes de texto para remover menções
  ambíguas ao antigo comando/daemon.
- **`CLAUDE.md`** documenta que os conjuntos de commands Claude/OpenCode diferem
  por design (não devem ser forçados a 1:1).

## [1.0.0] - 2026-07-07

### Changed

- **BREAKING: rc agora é só um plugin de skills, commands, hooks e agents** para
  Claude Code e OpenCode — markdown e shell puro, sem binário e sem build. A
  instalação passa a ser via marketplace de plugin (Claude Code) ou cópia do
  bundle `opencode/` (OpenCode); não há mais `rc setup`.
- **Skills e commands convertidos para prompt-only.** As skills que dirigiam o
  pipeline pelo binário (`rc tasks run`, `rc reviews fix`, `rc exec`, `rc tasks
  validate`, etc.) agora orientam o agente a executar cada fase diretamente.
  Nenhuma invocação do binário `rc` permanece.
- **Memória de projeto agora é file-based.** O subsistema `rc memory` (SQLite
  `.rc/memory.db` + embeddings) foi substituído por arquivos markdown em
  `.rc/memory/` (`<scope>__<key>.md`, um fato por arquivo), consultados por Grep
  e versionados no git. Todas as skills consumidoras e os hooks de recall/
  precompact foram atualizados.
- **Docs reescritos** (README, CLAUDE.md, AGENTS.md, CONTRIBUTING.md, runbook do
  plugin) para um repo markdown+shell sem binário.

### Added

- **Agents do plugin Claude Code.** Os dez agentes de fase do OpenCode foram
  portados para `agents/` (formato de agente de plugin): `rc` (orquestrador),
  `rc-prd`, `rc-techspec`, `rc-tasks`, `rc-exec`, `rc-exec-bulk`, `rc-review`,
  `rc-fix`, `rc-gan`, `rc-git` — cada um fixando um modelo à fase e delegando à
  skill correspondente.
- **Cinco skills** inspiradas em padrões do `oh-my-opencode-slim`, escritas no
  nosso formato markdown: `rc-codemap` (mapa hierárquico por diretório),
  `rc-worktrees` (git worktrees como lanes isoladas), `rc-deepwork` (scheduler
  para sessões pesadas com gates), `rc-loop` (loop generate→verify→retry contra
  um gate de sucesso explícito) e `rc-reflect` (recomenda o menor asset
  reutilizável a criar).
- **Dois agentes de apoio read-only** (Claude e OpenCode): `rc-explorer` (busca
  rápida no codebase) e `rc-librarian` (pesquisa de libs/docs externas via
  context7 + web), acionáveis pelo orquestrador `rc` em qualquer fase.
- **Hook `phase-reminder`** (SessionStart, opt-in via `RC_PHASE_REMINDER=1`):
  infere a fase do pipeline a partir dos artefatos em `.rc/tasks/<slug>` e injeta
  um lembrete de uma linha com a fase atual e o próximo passo.

### Removed

- **Todo o módulo Go e o CLI**: `internal/`, `cmd/`, `pkg/`, `sdk/`,
  `extensions/`, `rc.go`, `go.mod`/`go.sum`, o binário e o daemon.
- **App web e desktop**, monorepo JS (`web/`, `apps/`, `packages/`) e todo o
  tooling de build/release (Makefile, goreleaser, golangci, turbo, bun, vitest,
  CI de build).

## [0.13.0] - 2026-06-22

### Added

- **`rc install --headroom` e listagem de recursos.** O comando `rc install`
  agora instala mais de um recurso: além do `rtk`, suporta `--headroom`
  (instala o pacote Python `headroom-ai[all]` via pipx, pip3 ou pip; imprime
  instruções manuais quando não há instalador Python disponível). Rodar
  `rc install` sem flag lista os recursos instaláveis. A orquestração de
  detecção/instalação foi generalizada e é compartilhada por todos os recursos
  e pelo passo de RTK do `rc setup` (sem duplicação).
- **Tutorial de primeiros passos por recurso.** Após instalar (ou quando o
  recurso já está presente), `rc install` imprime um bloco "Getting started"
  com os comandos principais e o link da documentação oficial. A flag
  `--guide` (ex.: `rc install --rtk --guide`) mostra esse tutorial sob demanda,
  sem detectar nem instalar nada.

## [0.12.0] - 2026-06-22

### Added

- **Comando `rc install --rtk`.** Instala o `rtk` (runtime toolkit) diretamente,
  sem precisar passar pelo fluxo completo de `rc setup`. Detecta o `rtk` no
  `PATH` e reporta a versão quando já presente; quando ausente, roda o instalador
  apropriado para o SO (Homebrew, script oficial ou cargo) ou imprime instruções
  manuais quando nenhum instalador pode rodar. Com `--yes` instala de forma
  desassistida; sem ele, em terminal interativo, confirma antes de instalar. A
  lógica de RTK é compartilhada com `rc setup` (sem duplicação).

## [0.11.1] - 2026-06-21

### Changed

- **Docs e help do `rc init`.** O comando passa a aparecer na lista de
  subcomandos em `rc --help`, e o README ganhou um destaque e a seção
  "Start a new project" documentando `rc init` e a skill `rc-new-project`.

## [0.11.0] - 2026-06-21

### Added

- **Scaffold de projeto novo a partir do template TypeScript da rodolfochicone.** Duas
  frentes para começar um projeto do zero:

  **Comando `rc init [nome]`:** cria um repositório **privado** na organização
  **rodolfochicone** a partir do template `rodolfochicone/typescript-template` e o clona em
  `./<nome>/` no diretório atual (`gh repo create … --template … --private --clone`).
  - Fluxo **híbrido** do nome: com argumento usa direto; sem argumento e em
    terminal interativo, pergunta o nome (com validação); sem TTY e sem nome,
    retorna erro acionável em vez de travar.
  - Pré-valida o GitHub CLI e, em erro de configuração, **orienta como
    configurar**: `gh` não instalado, não autenticado ou sem acesso à org
    rodolfochicone — cada caso com comandos copiáveis, sem stacktrace.

  **Skill `rc-new-project`:** versão agêntica (Claude/Codex) do mesmo fluxo, com
  fases em ordem, confirmação no passo externo, comandos `gh` explícitos,
  orientação de configuração do GitHub e verificação real do clone antes de
  declarar sucesso.

## [0.10.0] - 2026-06-21

### Added

- **Suporte a monorepos com múltiplas pastas `.rc`.** O rc agora descobre a pasta
  `.rc` ativa em projetos que têm mais de uma (ex.: `packages/*/.rc`, `apps/*/.rc`),
  tanto nas skills quanto no binário.

  **Binário (`internal/core/workspace`):**
  - A descoberta continua caminhando **para cima** do diretório atual até o `.rc`
    mais próximo (então `cd <subprojeto> && rc ...` já escolhe o `.rc` certo).
  - Quando **nenhum** `.rc` existe acima, agora busca **para baixo** (ignorando
    `node_modules`, `.git`, `vendor`, `_archived`, limitada à profundidade 6):
    - **1 encontrado** → usa automaticamente.
    - **2+ encontrados** → erro claro listando os candidatos, pedindo `cd` no
      subprojeto ou `--workspace <dir>`.
    - **0 encontrados** → mantém o comportamento atual (raiz como workspace).
  - Nova flag global **`--workspace <dir>`** em todos os comandos para apontar o
    `.rc`/subprojeto explicitamente, sem precisar dar `cd`.

  **Skills:** `rc-create-prd` e `rc-idea-factory` perguntam em qual `.rc` salvar
  quando há mais de uma; `rc-create-techspec`, `rc-create-tasks`, `rc-review-round`
  e `rc-code-review` localizam o `.rc` que contém a tarefa (`<NN>-<slug>`) e só
  perguntam em caso de ambiguidade; `rc-readme` varre todos os `.rc` por ADRs.
  Projetos de pasta única (com ou sem `.rc`) seguem idênticos — sem perguntas novas.

#### Como atualizar o rc

```bash
rc upgrade
```

> ⚠️ O repositório é privado, então o `rc upgrade` precisa de um token. Garanta no
> shell (`~/.zshrc`):
>
> ```bash
> export GH_TOKEN="$(gh auth token)"
> ```

Confirme a versão depois:

```bash
rc --version   # deve mostrar v0.10.0
```

## [0.9.0] - 2026-06-21

### Added

- **`rc setup --sync`** — novo modo de sincronização de skills. Rodando dentro de
  um projeto, ele reconcilia as skills bundled do rc com a versão do binário:
  - ✅ **Atualiza** as skills bundled que o projeto já tem (quando mudaram).
  - ➕ **Adiciona** as skills bundled que estão faltando.
  - ⏭️ **Ignora** as que já estão atualizadas (não reescreve à toa).
  - 🔒 **Não toca** em skills de terceiros/customizadas no mesmo diretório.

#### Como atualizar o rc

```bash
rc upgrade
```

> ⚠️ O repositório é privado, então o `rc upgrade` precisa de um token. Garanta no
> shell (`~/.zshrc`):
>
> ```bash
> export GH_TOKEN="$(gh auth token)"
> ```
>
> Sem isso, o `upgrade` silenciosamente não faz nada.

Confirme a versão depois:

```bash
rc --version   # deve mostrar v0.9.0
```

#### Como usar o comando

```bash
# Claude Code (aceita "claude" ou "claude-code")
rc setup --sync --agent claude

# Codex
rc setup --sync --agent codex
```

Flags combináveis:

| Flag              | Efeito                                                                              |
| ----------------- | ---------------------------------------------------------------------------------- |
| `--agent <nome>`  | Agente alvo (repetível). `claude` → `.claude/skills/`, `codex` → `.agents/skills/` |
| `--yes` / `-y`    | Não-interativo (útil em scripts/CI)                                                |
| `--global` / `-g` | Sincroniza no diretório do usuário em vez do projeto                               |
| `--copy`          | Copia arquivos em vez de symlink                                                   |

Exemplos:

```bash
# Projeto, Claude, sem prompts
rc setup --sync --agent claude --yes

# Vários agentes de uma vez
rc setup --sync --agent claude --agent codex --yes

# Escopo global (na máquina, não no projeto)
rc setup --sync --agent claude --global --yes
```

> ❌ `--sync` não combina com `--all` nem `--skill` (eles selecionam catálogos
> explícitos e contrariam a ideia de sincronizar). O comando avisa se for usado junto.

#### Exemplo de saída

```
Sync Claude Code (project scope)

  ✓ Added (3)
    ✓  rc-readme        ./.claude/skills/rc-readme
    ✓  rc-postman       ./.claude/skills/rc-postman
    ✓  rc-openapi       ./.claude/skills/rc-openapi
  ✓ Updated (1)
    ✓  rc-create-prd    ./.claude/skills/rc-create-prd
  Unchanged  11 already current
```

#### Fluxo recomendado para o time

1. `rc upgrade` (atualiza o binário para v0.9.0).
2. Dentro de cada projeto: `rc setup --sync --agent claude --yes` (ou `codex`).
3. Repetir o passo 2 sempre que sair uma nova versão do rc — só o que mudou é atualizado.

## [0.2.4] - 2026-06-13

### Added
- Initial rc release
