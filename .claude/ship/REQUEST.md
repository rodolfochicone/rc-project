## Objetivo

Criar um app desktop nativo em Electron para macOS que sirva de "painel de
controle rc": executa todas as funcionalidades do rc (workflows/tasks, exec,
reviews, workspaces) e as acompanha em tempo real, embrulhando (wrap) a web UI
React já existente servida pelo daemon — sem reescrever o frontend. O app
gerencia o ciclo de vida do daemon e adiciona edição de configuração via novos
endpoints HTTP no daemon Go.

## Contexto (estado atual do codebase)

- **Daemon HTTP** (Gin) em `internal/api/httpapi/server.go`, rotas em
  `internal/api/core/routes.go` / `internal/api/contract/routes.go`. Bind em
  127.0.0.1 com validação de Host/Origin (localhost) + CSRF. Porta/socket
  publicados em `~/.rc/daemon/daemon.json`.
- **API REST + SSE** já cobre: runs (list/detail/snapshot/events/`stream` SSE/
  cancel), tasks (list/detail/board/items/spec/memory/start-run/validate/
  archive), reviews (fetch/watch/rounds/issues/start-run), workspaces
  (register/update/delete/resolve/sync), exec, sync.
- **Tempo real**: `GET /api/runs/{id}/stream` (SSE) com `Last-Event-ID`,
  heartbeat, overflow e cursor. Eventos em `pkg/rc/events/event.go` (~25 tipos:
  run/job/session/tool_call/usage/task/review). Journal em
  `~/.rc/runs/<run-id>/events.jsonl`.
- **Web UI** React 19 + Vite 8 + TanStack Router/Query em `web/` (~70-75%
  completa): dashboard, workflows, task board, runs com stream ao vivo, reviews,
  seletor de workspace. Embutida no binário via `web/embed.go` (embed.FS); dev
  via flag `--web-dev-proxy`. Tipos gerados de `openapi/rc-daemon.json`.
- **Lacunas atuais**: config (`~/.rc/config.toml` + `.rc/config.toml`) é
  CLI-only (sem HTTP); sem UI de gestão de workspaces, extensões ou reusable
  agents; sem controle de daemon a partir de um app desktop.
- **Stack**: Go (módulo principal), bun + turbo monorepo, `packages/ui`
  (`@rodolfochicone/ui`, React 19 + Tailwind 4 + base-ui). Verificação Go via
  `make verify`; frontend via `bun run frontend:*`.

## Requisitos

### Shell Electron (novo, provavelmente em `apps/desktop/` ou `electron/`)

- Carregar a web UI servida pelo daemon em `http://127.0.0.1:<porta>`, lendo
  porta/socket de `~/.rc/daemon/daemon.json` (não hardcodar).
- Respeitar validação de Origin/Host/CSRF do daemon (configurar a janela para
  que as requisições passem na verificação localhost existente).
- Janela principal + menu nativo macOS, tray/menu-bar com status do daemon,
  deep-link/atalhos, reabrir no dock.
- Empacotamento `.app` para macOS (electron-builder ou equivalente), com
  arquitetura universal (arm64 + x64), code signing e notarização documentados.

### Ciclo de vida do daemon (gerenciado pelo Electron)

- Fazer spawn do binário `rc daemon start` (descobrir/baixar/apontar o binário),
  monitorar `GET /api/daemon/health` e `/status`, reiniciar em queda, parar
  graciosamente (`POST /api/daemon/stop`) no quit do app.
- Lidar com daemon já em execução (anexar em vez de duplicar — respeitar
  `daemon.lock`).

### Configuração via novos endpoints Go (daemon) + UI

- Implementar handlers HTTP no daemon para **ler e gravar** config global
  (`~/.rc/config.toml`) e por workspace (`.rc/config.toml`), seguindo o schema
  de `internal/core/workspace/config_types.go` (defaults, tasks.run,
  fix_reviews, fetch_reviews, watch_reviews, exec, runs, sound).
  Atualizar `openapi/rc-daemon.json` e regenerar tipos TS (`bun run codegen`).
- Adicionar páginas na web UI compartilhada (beneficiando navegador + Electron):
  edição de config (global/workspace), gestão de workspaces (register/
  unregister/rename), e visão de extensões/reusable agents (read-only no
  mínimo).

### Tempo real

- Reusar o SSE existente (`/api/runs/{id}/stream`) para acompanhar runs ao vivo;
  garantir reconexão por cursor no contexto Electron.

## Critérios de aceitação

- [ ] App `.app` abre em macOS, sobe/anexa ao daemon automaticamente e exibe a
      web UI sem o usuário rodar `rc daemon start` manualmente.
- [ ] É possível iniciar um workflow/exec/review pelo app e ver os eventos
      (job/session/tool_call/usage) atualizando em tempo real via SSE.
- [ ] Ao fechar o app, o daemon é parado graciosamente (ou mantido, conforme
      preferência) sem processos órfãos nem locks presos.
- [ ] É possível ler e salvar config global e de workspace pelo app, com o
      arquivo `config.toml` refletindo as mudanças e o daemon recarregando.
- [ ] É possível registrar/remover/renomear workspaces pela UI.
- [ ] Novos endpoints Go passam `make verify` (fmt + lint zero-issues + test
      -race + build); frontend passa `bun run frontend:typecheck/test`.
- [ ] OpenAPI atualizado e tipos TS regenerados (`codegen-check` verde).
- [ ] Build do app documentado (signing/notarização) e reprodutível.

## Restrições

- **Não reescrever** a web UI existente; reaproveitar `web/` e `packages/ui`.
- Endpoints novos do daemon seguem as convenções de `internal/api`
  (timeouts por classe, header `X-rc-Workspace-ID`, formato de erro com
  `code`/`message`/`request_id`, validação localhost/CSRF).
- Go: erros embrulhados com `%w`, `slog`, `context.Context` nas fronteiras,
  sem `panic`/`log.Fatal` em produção, goroutines com ownership/shutdown.
  Ativar `golang-pro` antes de escrever Go; `testing-anti-patterns` nos testes.
- Não adicionar deps Go à mão (`go get`); seguir CLAUDE.md e dispatch de skills.
- Escrita de config deve ser atômica e preservar comentários/estrutura do TOML
  quando possível (avaliar lib de TOML já usada no projeto antes de adotar nova).
- Manter tudo single-binary/local-first: o Electron é um shell; a fonte da
  verdade continua sendo o daemon e os artefatos em `~/.rc` / `.rc`.
