# Changelog

## [Unreleased]

## [0.42.0] - 2026-07-10

### Changed

- **`/rc-card` materializa o workspace local `.rc/tasks/<slug>/`.** O comando deixa
  de só orquestrar skills e passa a persistir os artefatos de implementação: um
  `_techspec.md` **fino** (extrato das decisões técnicas do card + ponteiro para
  ele — cópia local confiável, já que o ticket é untrusted), um `_tasks.md`
  (índice master com `jira_key` em ordem de dependência) e, por tarefa, o
  `task_NN.md` que é o **plano aprovado do `/rc-plano`** carimbado com `jira_key`.
  **Não** gera `_prd.md` — a História refinada já é o PRD. Guardrail novo fixa os
  papéis: **Jira = tracking; `.rc/` local = contrato de implementação**, sem fonte
  de verdade duplicada.

## [0.41.0] - 2026-07-10

### Added

- **`/rc-card [story-key]`** — comando que conduz uma História Jira já refinada
  (ex.: via `rc-council`) de ponta a ponta. Descobre as Tarefas-filhas por
  `parent = <STORY>` com fallback para as keys da seção **Decomposição** da
  descrição, e roda um loop por Tarefa em ordem de dependência:
  `/rc-plano` (aprovação do plano) → executa com o loop verify→fix →
  `/rc-review` → `rc-jira` posta evidência de teste e transiciona o ticket para o
  status correto; roll-up na História no fim. Interativo e portável (pausa para as
  aprovações de plano/review/Jira); trata todo texto de ticket como dado não
  confiável e nunca marca uma Tarefa concluída em gate vermelho ou com achados
  high/critical em aberto. Um repo por execução.

## [0.40.0] - 2026-07-10

### Changed

- **`/rc-review` converge por severidade.** O loop-until-dry agora para quando um
  round não traz issues novos de severidade **alta/crítica** — issues medium/low
  ainda são corrigidos naquele round, mas não disparam um round extra (caro). O
  teto de **3 rounds** é mantido. Atinge `commands/rc-review.md` e
  `skills/rc-review-workflow/SKILL.md` (o schema `ROUND` ganhou `newBlocking`).
- **`/rc-exec` executa em loop verify→fix bounded.** `rc-execute-task` deixou de
  "consertar até resolver" (sem limite) e passou a iterar `gather → fix root cause
  → re-verify` em gate vermelho, até **3 fix cycles** por task, escalando o
  diagnóstico ao `rc-oracle` no último cycle. Se estourar o teto ainda vermelho,
  reporta a task como bloqueada com a evidência — nunca marca completa em gate
  vermelho (guarda contra *premature completion* e *over-ambition*).

## [0.39.1] - 2026-07-08

### Fixed

- **Referências quebradas em `rc-brainstorming`.** A skill apontava para skills
  inexistentes (`writing-plans`, `mcp-builder`); os handoffs agora vão para o
  pipeline real (`rc-create-prd` → `rc-create-techspec` → `rc-create-tasks`) e
  `frontend-design` → `rc-frontend-design`.
- **Nomes de skill não-canônicos em prosa.** `no-workarounds` →
  `rc-no-workarounds` (`rc-execute-task`), `tanstack` → `rc-tanstack`
  (`rc-react`), `test-anti-patterns` → `rc-testing-anti-patterns`
  (`rc-no-workarounds`).

### Changed

- **Leaf-workers agora alcançáveis pelas skills de execução.** As callouts de
  delegação de `rc-execute-task` e `rc-fix-reviews` roteiam lookups de docs para
  `rc-librarian` e apontam `rc-fixer` como upgrade path paralelo (worktree-isolado);
  `rc-fix-reviews` ganhou sua primeira callout de delegação.
- **Anti-triggers adicionados** a `rc-adversarial-review` e
  `rc-fix-coderabbit-review` para desambiguar do restante do cluster de review/fix.
- **`/rc-pipe`** ganhou um passo 0 opcional de warm-up (`rc-codemap`) para baratear
  a exploração das fases seguintes.

## [0.39.0] - 2026-07-08

### Added

- **6 skills novas (padrão hub + `references/`, auto-descobertas por diretório):**
  - `rc-seo` — SEO técnico, on-page e programático (auditoria, otimização de
    conteúdo, geração de páginas em escala).
  - `rc-video` — processamento local com `ffmpeg`, criação de conteúdo
    (Reels/Shorts/YouTube) e integração opcional com VideoDB (SaaS pago).
  - `rc-a11y` — acessibilidade WCAG 2.2 AA (HTML semântico, ARIA, navegação por
    teclado, gestão de foco, contraste, leitores de tela).
  - `rc-sql` — otimização de query (EXPLAIN, índices, N+1) e design de schema;
    read-only por padrão (Rule 9).
  - `rc-observability` — logs, métricas, traces e resposta a incidentes
    (instrumentação, SLOs, postmortem).
  - `rc-resilience` — resiliência event-driven (idempotência, retry/backoff,
    DLQ, poison message, timeouts, circuit breaker).

### Fixed

- **Drift de documentação do path de instincts.** `COMMANDS.md` e `README.md`
  apontavam `.rc/instincts/` para as observações do hook `observe`; corrigido
  para `.rc/memory/observations.jsonl`, que é onde o hook de fato grava.

### Changed

- **Extensão `rc-idea-factory` alinhada à versão do plugin (`0.39.0`).**

## [0.38.0] - 2026-07-08

### Added

- **Skill `rc-python`** — Python 3.12+ idiomático e tipado, com references
  dedicadas: typing/PEP 695, asyncio/`TaskGroup`, packaging com `uv` e testes
  com pytest.
- **Skill `rc-hookify`** — autoria de hooks RC a partir de uma regra em
  linguagem natural: escreve o script fail-open, conecta no `hooks.json`,
  documenta e verifica; inclui referência de eventos de hook.
- **Hook `memory-load` (`SessionStart`)** — warm-start que injeta no contexto um
  resumo limitado de `.rc/memory/` (fatos + learnings) e avisa quando há
  observações a destilar. Nunca bloqueia; silencioso fora de projetos RC.

### Changed

- **Documentação de `model`/`effort`** e contrato de delegação dos agents
  cost-tiered (`skills/rc/references/delegation-contract.md`).
- Ajustes em `rc-memory`, `README.md` e `hooks/README.md` refletindo o hook
  `memory-load`.

## [0.37.2] - 2026-07-08

### Fixed

- **Hook `repair-guidance` disparava falso-positivo em todo edit bem-sucedido.**
  Quando o `tool_response` do PostToolUse vem como objeto (builds atuais do
  Claude Code), o hook fazia `tojson` do objeto inteiro — que num Edit de
  sucesso embute o conteúdo do arquivo (`originalFile`/`structuredPatch`) — e
  rodava o grep de falha nisso. Qualquer arquivo contendo frases como "not
  found", "no changes" ou "old_string" fazia o hook emitir "Edit did not apply"
  mesmo após um edit aplicado. O mesmo afetava o branch `Task` (grep
  "error|failed" contra a saída inteira do subagente). Agora o hook inspeciona
  apenas o texto de status/erro — a string, ou os campos
  `.error`/`.message`/`.errorMessage` do objeto — nunca o objeto serializado.
  Cobertura adicionada ao `--selftest` (`edit-ok-object`, `task-ok-object`).

## [0.37.1] - 2026-07-08

### Changed

- **`model`/`effort` explícitos em 13 skills comportamentais.** Skills que são
  unidade discreta de trabalho passaram a pinar tier (antes herdavam o da
  sessão), alinhadas à convenção das skills de pipeline:
  - **opus/high** — `rc-council`, `rc-adversarial-review`,
    `rc-refactoring-analysis`, `rc-ubs`.
  - **sonnet/high** — `rc-brainstorming`, `rc-graphify`, `rc-qa-execution`,
    `rc-qa-report`, `rc-fix-coderabbit-review`, `rc-autoresearch`.
  - **sonnet/medium** — `rc-enrichment-prompt`, `rc-to-prompt`,
    `rc-lesson-learned`.

  Cobertura de tier sobe de 27 → 40 das 75 skills. As demais 35 (referência de
  biblioteca/design e guidance cross-cutting como `rc-tdd`,
  `rc-systematic-debugging`, `rc-no-workarounds`, `rc-testing-anti-patterns`,
  `rc-skill-best-practices`) seguem sem pin de propósito — rodam no modelo da
  sessão.

## [0.37.0] - 2026-07-07

### Changed

- **Consolidação de skills (82 → 75).** Skills que fatiavam a mesma
  biblioteca/domínio ou tinham o mesmo job foram fundidas na skill primária,
  preservando todo o conteúdo detalhado (os `rules/`/`references/` foram
  movidos para dentro da primária, não descartados):
  - `rc-tanstack` absorveu `rc-tanstack-query-best-practices`,
    `rc-tanstack-router-best-practices` e `rc-tanstack-start-best-practices`
    (agora em `references/{query,router,start}/`).
  - `rc-git` absorveu `rc-git-rebase` (rebase/conflitos; `references/` e
    `scripts/` movidos). O command `rc-commit-msg` permanece intacto.
  - `rc-readme` absorveu `rc-crafting-effective-readmes` (templates/guidance
    para escrever à mão; `references/`, `templates/` e guias movidos).
  - `rc-vercel-react-best-practices` absorveu `rc-vercel-composition-patterns`
    (em `rules/composition/`).
  - `rc-refactoring-analysis` absorveu `rc-architectural-analysis` (auditoria de
    dead code, anti-patterns e type confusion; metodologia em
    `references/architectural-audit.md`).

### Removed

- Skills `rc-tanstack-query-best-practices`, `rc-tanstack-router-best-practices`,
  `rc-tanstack-start-best-practices`, `rc-git-rebase`,
  `rc-crafting-effective-readmes`, `rc-vercel-composition-patterns` e
  `rc-architectural-analysis` como entradas independentes. **Breaking**:
  invocações por esses nomes deixam de resolver — use a skill primária
  correspondente.

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

- **Suporte a monorepos com múltiplas pastas `.rc`.** O RC agora descobre a pasta
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

#### Como atualizar o RC

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
  um projeto, ele reconcilia as skills bundled do RC com a versão do binário:
  - ✅ **Atualiza** as skills bundled que o projeto já tem (quando mudaram).
  - ➕ **Adiciona** as skills bundled que estão faltando.
  - ⏭️ **Ignora** as que já estão atualizadas (não reescreve à toa).
  - 🔒 **Não toca** em skills de terceiros/customizadas no mesmo diretório.

#### Como atualizar o RC

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
3. Repetir o passo 2 sempre que sair uma nova versão do RC — só o que mudou é atualizado.

## [0.2.4] - 2026-06-13

### Added
- Initial RC release
